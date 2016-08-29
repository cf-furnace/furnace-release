package routing

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/cf-furnace/controller/routing/iptables"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/cache"
	v1core "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3/typed/core/v1"
	"k8s.io/kubernetes/pkg/controller/framework"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/watch"
)

const (
	PROCESS_GUID_LABEL = "cloudfoundry.org/process-guid"

	NODE_PORTS_ANNOTATION = "cloudfoundry.org/node-ports"
)

//go:generate counterfeiter -o fakes/fake_locker.go . locker
type locker interface {
	sync.Locker
}

//go:generate counterfeiter -o fakes/fake_port_pool.go . portPool
type portPool interface {
	Acquire() (uint32, error)
	Release(port uint32)
	Remove(port uint32) error
}

type NodePort uint32
type ContainerPort uint32

type PortMapping struct {
	ContainerPort ContainerPort `json:"container_port"`
	NodePort      NodePort      `json:"node_port"`
}

type NodeRouteController struct {
	kubeClient v1core.CoreInterface

	store      cache.Store
	controller *framework.Controller
	portPool   portPool
	ipt        iptables.IPTables

	lock      locker
	localPods map[string]*v1.Pod
}

func NewNodeRouteController(
	coreClient v1core.CoreInterface,
	resyncPeriod time.Duration,
	nodeName string,
	portPool portPool,
	ipt iptables.IPTables,
	locker locker,
) *NodeRouteController {
	nrc := &NodeRouteController{
		kubeClient: coreClient,
		portPool:   portPool,
		ipt:        ipt,
		lock:       locker,
		localPods:  map[string]*v1.Pod{},
	}

	labelSelector, err := labels.Parse(PROCESS_GUID_LABEL)
	if err != nil {
		return nil
	}

	nrc.store, nrc.controller = framework.NewInformer(
		&cache.ListWatch{
			ListFunc: func(options api.ListOptions) (runtime.Object, error) {
				options.FieldSelector = fields.Set{"spec.nodeName": nodeName}.AsSelector()
				options.LabelSelector = labelSelector
				return coreClient.Pods(api.NamespaceAll).List(options)
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				options.FieldSelector = fields.Set{"spec.nodeName": nodeName}.AsSelector()
				options.LabelSelector = labelSelector
				return coreClient.Pods(api.NamespaceAll).Watch(options)
			},
		},
		&v1.Pod{},
		resyncPeriod,
		nrc,
	)

	return nrc
}

func (nrc *NodeRouteController) Run(stopCh <-chan struct{}) {
	go nrc.controller.Run(stopCh)
	<-stopCh
}

func (nrc *NodeRouteController) IfritRun(signals <-chan os.Signal, ready chan<- struct{}) error {
	stopChan := make(chan struct{})

	go nrc.Run(stopChan)
	close(ready)

	<-signals
	close(stopChan)

	return nil
}

func (nrc *NodeRouteController) HasSynced() bool {
	return nrc.HasSynced()
}

func (nrc *NodeRouteController) OnAdd(obj interface{}) {
	nrc.OnUpdate(nil, obj)
}

func (nrc *NodeRouteController) OnUpdate(old, new interface{}) {
	newPod := AsPod(new)
	if newPod == nil {
		return
	}

	if newPod.Status.HostIP == "" || newPod.Status.PodIP == "" {
		return
	}

	nrc.lock.Lock()
	defer nrc.lock.Unlock()

	_, processed := nrc.localPods[processGuidFromPod(newPod)]
	if processed {
		return
	}

	if _, ok := newPod.Annotations[NODE_PORTS_ANNOTATION]; !ok {
		mappings, err := nrc.allocatePorts(extractContainerPorts(newPod)...)
		if err != nil {
			panic(err)
		}

		marshalledPods, err := json.Marshal(mappings)
		if err != nil {
			panic(err)
		}

		nrc.setupNATRules(newPod, mappings)

		if len(newPod.Annotations) == 0 {
			newPod.Annotations = map[string]string{}
		}

		for {
			newPod.Annotations[NODE_PORTS_ANNOTATION] = string(marshalledPods)
			newPod, err = nrc.kubeClient.Pods(newPod.Namespace).Update(newPod)
			if err == nil {
				break
			}
			if !errors.IsConflict(err) {
				panic(err)
			}
		}
	}

	nrc.localPods[processGuidFromPod(newPod)] = newPod
}

func (nrc *NodeRouteController) OnDelete(obj interface{}) {
	pod := AsPod(obj)
	if pod != nil {
		nrc.lock.Lock()
		defer nrc.lock.Unlock()
		delete(nrc.localPods, processGuidFromPod(pod))

		if nodePortsAnnotation, ok := pod.Annotations[NODE_PORTS_ANNOTATION]; ok {
			portMappings := []PortMapping{}
			err := json.Unmarshal([]byte(nodePortsAnnotation), &portMappings)
			if err != nil {
				return
			}

			err = nrc.teardownNATRules(pod)
			if err != nil {
				panic(err)
			}

			for _, portMapping := range portMappings {
				nrc.portPool.Release(uint32(portMapping.NodePort))
			}
		}
	}
}

func (nrc *NodeRouteController) Get(processGuid string) (*v1.Pod, bool) {
	nrc.lock.Lock()
	defer nrc.lock.Unlock()

	pod, ok := nrc.localPods[processGuid]
	return pod, ok
}

func (nrc *NodeRouteController) allocatePorts(ports ...ContainerPort) ([]PortMapping, error) {
	portMappings := []PortMapping{}

	for _, containerPort := range ports {
		nodePort, err := nrc.portPool.Acquire()
		if err != nil {
			return nil, err
		}
		portMappings = append(portMappings, PortMapping{
			ContainerPort: containerPort,
			NodePort:      NodePort(nodePort),
		})
	}

	return portMappings, nil
}

func processGuidFromPod(pod *v1.Pod) string {
	return pod.Labels[PROCESS_GUID_LABEL]
}

func extractContainerPorts(pod *v1.Pod) []ContainerPort {
	containerPorts := []ContainerPort{}

	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			if port.Protocol == v1.ProtocolTCP {
				containerPorts = append(containerPorts, ContainerPort(port.ContainerPort))
			}
		}
	}

	return containerPorts
}

func (nrc *NodeRouteController) setupNATRules(pod *v1.Pod, mappings []PortMapping) error {
	chainName, err := iptables.InstanceChainName(iptables.InstanceChainPrefix, pod)
	if err != nil {
		return err
	}

	err = nrc.ipt.CreateChain(iptables.NAT, chainName)
	if err != nil {
		return err
	}

	err = nrc.ipt.AppendRule(iptables.NAT, iptables.FurnacePreroutingChain, &iptables.JumpRule{TargetChain: chainName})
	if err != nil {
		return err
	}

	for _, portMapping := range mappings {
		err := nrc.ipt.AppendRule(iptables.NAT, chainName, &iptables.DNATRule{
			HostAddress:      pod.Status.HostIP,
			HostPort:         int32(portMapping.NodePort),
			ContainerAddress: pod.Status.PodIP,
			ContainerPort:    int32(portMapping.ContainerPort),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (nrc *NodeRouteController) teardownNATRules(pod *v1.Pod) error {
	chainName, err := iptables.InstanceChainName(iptables.InstanceChainPrefix, pod)
	if err != nil {
		return err
	}

	err = nrc.ipt.DeleteChainReferences(iptables.NAT, iptables.FurnacePreroutingChain, chainName)
	if err != nil {
		return err
	}

	err = nrc.ipt.DeleteChain(iptables.NAT, chainName)
	if err != nil {
		return err
	}

	return nil
}

func AsPod(obj interface{}) *v1.Pod {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return nil
	}

	return pod
}
