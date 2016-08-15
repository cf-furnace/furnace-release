package routing_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"

	"testing"
)

func TestRouting(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Routing Suite")
}

func LoadClientConfig() *restclient.Config {
	home := os.Getenv("HOME")
	config, err := clientcmd.LoadFromFile(filepath.Join(home, ".kube", "config"))
	Expect(err).NotTo(HaveOccurred())

	context := config.Contexts[config.CurrentContext]

	return &restclient.Config{
		Host:     config.Clusters[context.Cluster].Server,
		Username: config.AuthInfos[context.AuthInfo].Username,
		Password: config.AuthInfos[context.AuthInfo].Password,
		Insecure: config.Clusters[context.Cluster].InsecureSkipTLSVerify,
		TLSClientConfig: restclient.TLSClientConfig{
			CertFile: config.AuthInfos[context.AuthInfo].ClientCertificate,
			KeyFile:  config.AuthInfos[context.AuthInfo].ClientKey,
			CAFile:   config.Clusters[context.Cluster].CertificateAuthority,
		},
	}
}
