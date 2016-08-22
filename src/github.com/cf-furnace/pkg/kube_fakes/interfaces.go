package kube_fakes

import (
	"k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3/typed/core/v1"
	k8swatch "k8s.io/kubernetes/pkg/watch"
)

//go:generate counterfeiter -o fake_core_client.go --fake-name FakeCoreClient . coreClient
type coreClient interface {
	v1.CoreInterface
}

//go:generate counterfeiter -o fake_pod_client.go --fake-name FakePodClient . podClient
type podClient interface {
	v1.PodInterface
}

//go:generate counterfeiter -o fake_watch.go --fake-name FakeWatch . watchInterface
type watchInterface interface {
	k8swatch.Interface
}

//go:generate counterfeiter -o fake_rc_client.go --fake-name FakeReplicationControllerClient . replicationControllerInterface
type replicationControllerInterface interface {
	v1.ReplicationControllerInterface
}
