package lifecycle_test

import (
	"fmt"
	"strings"
	"tests/helpers/assets"
	"time"

	"k8s.io/kubernetes/pkg/api"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3"
	v1core "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3/typed/core/v1"
	"k8s.io/kubernetes/pkg/labels"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

const DEFAULT_TIMEOUT = 2 * time.Minute

var _ = Describe("Lifecycle", func() {
	var appName string
	var client v1core.CoreInterface

	BeforeEach(func() {
		appName = generator.PrefixedRandomName("KUBE-APP-")

		var err error
		clientSet, err := clientset.NewForConfig(kubeConfig.ClientConfig())
		Expect(err).NotTo(HaveOccurred())

		client = clientSet.Core()
	})

	It("creates, scales, and deletes ruby buildpack based application pods", func() {
		By("pushing an application")
		Expect(cf.Cf(
			"push", appName,
			"-b", config.RubyBuildpackName,
			"-m", "512M",
			"-p", assets.NewAssets().Dora,
			"-d", config.AppsDomain,
		).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		appGuid := strings.TrimSpace(string(cf.Cf("app", appName, "--guid").Wait(DEFAULT_TIMEOUT).Out.Contents()))

		Eventually(replicationControllers(client, appGuid), DEFAULT_TIMEOUT).Should(HaveLen(1))
		Eventually(pods(client, appGuid), DEFAULT_TIMEOUT).Should(HaveLen(1))

		By("scaling application")
		Expect(cf.Cf("scale", appName, "-i", "3").Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		Eventually(pods(client, appGuid), DEFAULT_TIMEOUT).Should(HaveLen(3))

		By("stopping application")
		Expect(cf.Cf("stop", appName).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		Eventually(replicationControllers(client, appGuid), DEFAULT_TIMEOUT).Should(HaveLen(0))
		Eventually(pods(client, appGuid), DEFAULT_TIMEOUT).Should(HaveLen(0))
	})

	It("creates, scales, and deletes docker based application pods", func() {
		By("pushing an application")
		Expect(cf.Cf(
			"push", appName,
			"-m", "512M",
			"-o", "cloudfoundry/lattice-app",
			"-d", config.AppsDomain,
		).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		appGuid := strings.TrimSpace(string(cf.Cf("app", appName, "--guid").Wait(DEFAULT_TIMEOUT).Out.Contents()))

		Eventually(replicationControllers(client, appGuid), DEFAULT_TIMEOUT).Should(HaveLen(1))
		Eventually(pods(client, appGuid), DEFAULT_TIMEOUT).Should(HaveLen(1))

		By("scaling application")
		Expect(cf.Cf("scale", appName, "-i", "3").Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		Eventually(pods(client, appGuid), DEFAULT_TIMEOUT).Should(HaveLen(3))

		By("stopping application")
		Expect(cf.Cf("stop", appName).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		Eventually(replicationControllers(client, appGuid), DEFAULT_TIMEOUT).Should(HaveLen(0))
		Eventually(pods(client, appGuid), DEFAULT_TIMEOUT).Should(HaveLen(0))
	})
})

func replicationControllers(client v1core.CoreInterface, appGuid string) func() []string {
	return func() []string {
		rcList, err := client.ReplicationControllers(api.NamespaceAll).List(api.ListOptions{
			LabelSelector: labels.Set{"cloudfoundry.org/app-guid": appGuid}.AsSelector(),
		})
		if err != nil {
			fmt.Fprintf(GinkgoWriter, "List replication controller failed: %s\n", err.Error())
			return nil
		}

		result := []string{}
		for _, rc := range rcList.Items {
			result = append(result, rc.Name)
		}

		return result
	}
}

func pods(client v1core.CoreInterface, appGuid string) func() []string {
	return func() []string {
		podList, err := client.Pods(api.NamespaceAll).List(api.ListOptions{
			LabelSelector: labels.Set{"cloudfoundry.org/app-guid": appGuid}.AsSelector(),
		})
		if err != nil {
			fmt.Fprintf(GinkgoWriter, "List pods failed: %s\n", err.Error())
			return nil
		}

		result := []string{}
		for _, pod := range podList.Items {
			result = append(result, pod.Name)
		}

		return result
	}
}
