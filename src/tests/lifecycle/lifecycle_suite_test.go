package lifecycle_test

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/kubernetes/pkg/client/restclient"

	"testing"
)

var (
	context     helpers.SuiteContext
	config      helpers.Config
	kubeConfig  *KubeConfig
	environment *helpers.Environment
)

func TestLifecycle(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Lifecycle Suite")
}

var _ = BeforeSuite(func() {
	config = helpers.LoadConfig()
	context = helpers.NewContext(config)

	var err error
	kubeConfig, err = loadKubeConfig()
	Expect(err).NotTo(HaveOccurred())

	environment = helpers.NewEnvironment(context)
	environment.Setup()
})

var _ = AfterSuite(func() {
	environment.Teardown()
})

type KubeConfig struct {
	APIServer string `json:"api_server"`
	CertFile  string `json:"cert_file"`
	KeyFile   string `json:"key_file"`
	CAFile    string `json:"ca_file"`
}

func loadKubeConfig() (*KubeConfig, error) {
	path := os.Getenv("CONFIG")
	if path == "" {
		return nil, errors.New("$CONFIG must point to a test configuration file")
	}

	configFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	config := &KubeConfig{}
	if err := json.NewDecoder(configFile).Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *KubeConfig) ClientConfig() *restclient.Config {
	return &restclient.Config{
		Host: c.APIServer,
		TLSClientConfig: restclient.TLSClientConfig{
			CertFile: c.CertFile,
			KeyFile:  c.KeyFile,
			CAFile:   c.CAFile,
		},
	}
}
