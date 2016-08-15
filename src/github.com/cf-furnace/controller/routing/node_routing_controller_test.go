package routing_test

import (
	"os"
	"time"

	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/types"
	"k8s.io/kubernetes/pkg/watch"

	"github.com/cf-furnace/controller/routing"
	"github.com/cf-furnace/controller/routing/fakes"
	"github.com/cf-furnace/controller/routing/iptables"
	iptables_fakes "github.com/cf-furnace/controller/routing/iptables/fakes"
	"github.com/cf-furnace/pkg/cloudfoundry"
	"github.com/cf-furnace/pkg/kube_fakes"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RouteEmitter", func() {
	var (
		coreClient *kube_fakes.FakeCoreClient
		podClient  *kube_fakes.FakePodClient

		resyncPeriod time.Duration
		nodeName     string
		portPool     *fakes.FakePortPool
		ipt          *iptables_fakes.FakeIPTables
		locker       *fakes.FakeLocker

		controller *routing.NodeRouteController

		pg                cloudfoundry.ProcessGuid
		pod               *v1.Pod
		expectedChainName string
	)

	BeforeEach(func() {
		coreClient = &kube_fakes.FakeCoreClient{}
		podClient = &kube_fakes.FakePodClient{}
		coreClient.PodsReturns(podClient)

		resyncPeriod = time.Second
		nodeName = "test-node-0"
		portPool = &fakes.FakePortPool{}
		ipt = &iptables_fakes.FakeIPTables{}
		locker = &fakes.FakeLocker{}

		guid, err := uuid.NewV4()
		Expect(err).NotTo(HaveOccurred())

		pg = cloudfoundry.ProcessGuid{AppGuid: guid, AppVersion: guid}
		pod = &v1.Pod{
			ObjectMeta: v1.ObjectMeta{
				UID: types.UID(guid.String()),
				Labels: map[string]string{
					routing.PROCESS_GUID_LABEL: pg.ShortenedGuid(),
				},
			},
		}

		expectedChainName, err = iptables.InstanceChainName("k-", pod)
		Expect(err).NotTo(HaveOccurred())

		controller = routing.NewNodeRouteController(coreClient, resyncPeriod, nodeName, portPool, ipt, locker)
	})

	Describe("Run", func() {
		var (
			eventCh chan watch.Event
			stopCh  chan struct{}
		)

		BeforeEach(func() {
			podList := &v1.PodList{Items: []v1.Pod{*pod}}
			podClient.ListReturns(podList, nil)

			eventCh = make(chan watch.Event, 1)
			watch := &kube_fakes.FakeWatch{}
			watch.ResultChanReturns(eventCh)
			podClient.WatchReturns(watch, nil)

			stopCh = make(chan struct{})
		})

		AfterEach(func() {
			select {
			case _, ok := <-stopCh:
				if ok {
					close(stopCh)
				}
			default:
			}
		})

		It("runs until stopped", func() {
			doneCh := make(chan struct{})
			go func() { controller.Run(stopCh); close(doneCh) }()

			Consistently(doneCh).ShouldNot(BeClosed())
			close(stopCh)
			Eventually(doneCh).Should(BeClosed())
		})

		It("lists all pods on the specified node", func() {
			go controller.Run(stopCh)

			Eventually(podClient.ListCallCount).Should(Equal(1))
			Eventually(podClient.WatchCallCount()).Should(Equal(1))
		})
	})

	Describe("IfritRun", func() {
		var (
			eventCh chan watch.Event
			ready   chan struct{}
			signals chan os.Signal
		)

		BeforeEach(func() {
			podList := &v1.PodList{Items: []v1.Pod{*pod}}
			podClient.ListReturns(podList, nil)

			eventCh = make(chan watch.Event, 1)
			watch := &kube_fakes.FakeWatch{}
			watch.ResultChanReturns(eventCh)
			podClient.WatchReturns(watch, nil)

			ready = make(chan struct{})
			signals = make(chan os.Signal)
		})

		It("closes the ready channel", func() {
			go controller.IfritRun(signals, ready)
			Eventually(ready).Should(BeClosed())
			close(signals)
		})

		It("terminates when signalled", func() {
			doneCh := make(chan struct{})
			go func() { controller.IfritRun(signals, ready); close(doneCh) }()

			Consistently(doneCh).ShouldNot(BeClosed())
			Eventually(signals).Should(BeSent(os.Kill))
			Eventually(doneCh).Should(BeClosed())
		})
	})

	Describe("Get", func() {
		It("returns nil and false when the pod is not found", func() {
			p, ok := controller.Get("missing-key")

			Expect(p).To(BeNil())
			Expect(ok).To(BeFalse())
		})

		It("returns the value and true for present when the pod is found", func() {
			controller.OnAdd(pod)

			p, ok := controller.Get(pg.ShortenedGuid())
			Expect(p).To(Equal(pod))
			Expect(ok).To(BeTrue())
		})
	})

	Describe("OnAdd", func() {
		It("protects the data with a mutex", func() {
			controller.OnAdd(pod)

			Expect(locker.LockCallCount()).To(Equal(1))
			Expect(locker.UnlockCallCount()).To(Equal(1))
		})

		It("tracks the pod in the controller", func() {
			controller.OnAdd(pod)

			p, ok := controller.Get(pg.ShortenedGuid())
			Expect(p).To(Equal(pod))
			Expect(ok).To(BeTrue())
		})
	})

	Describe("OnUpdate", func() {
		var updated *v1.Pod

		BeforeEach(func() {
			newPod := *pod
			newPod.ObjectMeta.ResourceVersion = "new-resource-version"
			updated = &newPod
		})

		It("protects the data with a mutex", func() {
			controller.OnUpdate(pod, updated)

			Expect(locker.LockCallCount()).To(Equal(1))
			Expect(locker.UnlockCallCount()).To(Equal(1))
		})

		It("replaces the old pod in the controller", func() {
			controller.OnAdd(pod)
			controller.OnUpdate(pod, updated)

			p, ok := controller.Get(pg.ShortenedGuid())
			Expect(p).To(Equal(updated))
			Expect(ok).To(BeTrue())
		})

		Context("when containers with ports are present", func() {
			BeforeEach(func() {
				pod.Status.HostIP = "host-address"
				pod.Status.PodIP = "pod-address"
				pod.Spec.Containers = []v1.Container{{
					Name: "application",
					Ports: []v1.ContainerPort{
						{Name: "web", Protocol: v1.ProtocolTCP, ContainerPort: 8080},
						{Name: "web2", Protocol: v1.ProtocolTCP, ContainerPort: 9090},
					},
				}}

				portPool.AcquireStub = func() (uint32, error) {
					return uint32(portPool.AcquireCallCount()), nil
				}
			})

			It("reserves a port for each route", func() {
				controller.OnAdd(pod)
				Expect(portPool.AcquireCallCount()).To(Equal(2))
			})

			It("annotates the pod with the mapping", func() {
				controller.OnAdd(pod)

				Expect(podClient.UpdateCallCount()).To(Equal(1))
				updated := podClient.UpdateArgsForCall(0)
				Expect(updated.Annotations[routing.NODE_PORTS_ANNOTATION]).To(MatchJSON(`{"8080":1, "9090":2}`))
			})

			It("creates a NAT instance chain", func() {
				controller.OnAdd(pod)

				Expect(ipt.CreateChainCallCount()).To(Equal(1))
				table, name := ipt.CreateChainArgsForCall(0)
				Expect(table).To(Equal(iptables.NAT))
				Expect(name).To(Equal(expectedChainName))
			})

			It("adds the instance chain to the furnace prerouting rule chain", func() {
				controller.OnAdd(pod)

				Expect(ipt.AppendRuleCallCount()).NotTo(Equal(0))
				table, chain, rule := ipt.AppendRuleArgsForCall(0)
				Expect(table).To(Equal(iptables.NAT))
				Expect(chain).To(Equal(iptables.FurnacePreroutingChain))
				Expect(rule).To(Equal(&iptables.JumpRule{TargetChain: expectedChainName}))
			})

			It("adds rules to the NAT instance chain for each container port", func() {
				controller.OnAdd(pod)

				rules := []iptables.Rule{}
				for i := 1; i < ipt.AppendRuleCallCount(); i++ {
					table, chainName, rule := ipt.AppendRuleArgsForCall(i)
					Expect(table).To(Equal(iptables.NAT))
					Expect(chainName).To(Equal(expectedChainName))

					rules = append(rules, rule)
				}

				Expect(rules).To(ConsistOf(
					&iptables.DNATRule{
						HostAddress:      "host-address",
						HostPort:         1,
						ContainerAddress: "pod-address",
						ContainerPort:    8080,
					},
					&iptables.DNATRule{
						HostAddress:      "host-address",
						HostPort:         2,
						ContainerAddress: "pod-address",
						ContainerPort:    9090,
					},
				))
			})
		})

		Context("when a node port annotation is already present", func() {
			BeforeEach(func() {
				pod.Annotations = map[string]string{routing.NODE_PORTS_ANNOTATION: `{"8080": 1}`}
			})

			It("ignores the update", func() {
				controller.OnAdd(pod)

				Expect(podClient.UpdateCallCount()).To(Equal(0))
				Expect(ipt.CreateChainCallCount()).To(Equal(0))
			})
		})
	})

	Describe("OnDelete", func() {
		BeforeEach(func() {
			controller.OnAdd(pod)
		})

		It("protects the data with a mutex", func() {
			controller.OnDelete(pod)

			Expect(locker.LockCallCount()).To(Equal(2))
			Expect(locker.UnlockCallCount()).To(Equal(2))
		})

		It("removes the pod from the controller", func() {
			controller.OnDelete(pod)

			p, ok := controller.Get(pg.ShortenedGuid())
			Expect(p).To(BeNil())
			Expect(ok).To(BeFalse())
		})

		Context("when a node ports annotation is present", func() {
			BeforeEach(func() {
				pod.Annotations = map[string]string{routing.NODE_PORTS_ANNOTATION: `{"8080": 1, "9090": 2}`}
			})

			It("releases the ports", func() {
				controller.OnDelete(pod)

				ports := []uint32{}
				for i := 0; i < portPool.ReleaseCallCount(); i++ {
					ports = append(ports, portPool.ReleaseArgsForCall(i))
				}
				Expect(ports).To(ConsistOf(uint32(1), uint32(2)))
			})

			It("deletes the NAT instance chain", func() {
				controller.OnDelete(pod)

				Expect(ipt.DeleteChainReferencesCallCount()).To(Equal(1))
				table, parent, target := ipt.DeleteChainReferencesArgsForCall(0)
				Expect(table).To(Equal(iptables.NAT))
				Expect(parent).To(Equal(iptables.FurnacePreroutingChain))
				Expect(target).To(Equal(expectedChainName))

				Expect(ipt.DeleteChainCallCount()).To(Equal(1))
				table, target = ipt.DeleteChainArgsForCall(0)
				Expect(table).To(Equal(iptables.NAT))
				Expect(target).To(Equal(expectedChainName))
			})
		})
	})

	Describe("AsPod", func() {
		It("returns nil when the object is nil", func() {
			Expect(routing.AsPod(nil)).To(BeNil())
		})

		It("returns nil when object is not a pod", func() {
			Expect(routing.AsPod("string")).To(BeNil())
		})

		It("returns a pod when the object is a pod", func() {
			pod := &v1.Pod{
				ObjectMeta: v1.ObjectMeta{Name: "my-test-pod"},
			}

			Expect(routing.AsPod(pod)).To(Equal(pod))
		})
	})
})
