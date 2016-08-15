package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cf-furnace/controller/routing/cmd/endpoint-manager/testrunner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Endpoint Manager", func() {
	var (
		args testrunner.Args

		runner  *ginkgomon.Runner
		process ifrit.Process
	)

	startEndpointManager := func(args testrunner.Args, extraArgs ...string) (*ginkgomon.Runner, ifrit.Process) {
		runner := createEndpointManagerRunner(args, extraArgs...)
		return runner, ginkgomon.Invoke(runner)
	}

	BeforeEach(func() {
		args = testrunner.Args{
			KubeCluster:  "http://127.0.0.1:8080",
			KubeNodeName: "test-node",
		}
		process = nil
	})

	AfterEach(func() {
		ginkgomon.Interrupt(process, 5*time.Second)
	})

	Context("when -debugAddr is specified", func() {
		BeforeEach(func() {
			args.DebugAddress = fmt.Sprintf("127.0.0.1:%d", portOffset(DEBUG))
			runner, process = startEndpointManager(args)
		})

		It("exposes a debug endpoint", func() {
			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/debug/pprof/cmdline", portOffset(DEBUG)))
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("endpoint-manager"))
			Expect(body).To(ContainSubstring("-debugAddr"))
		})
	})

	Context("when the kubeNodeName flag is missing", func() {
		BeforeEach(func() {
			args.KubeNodeName = ""
			runner = createEndpointManagerRunner(args)
		})

		It("terminates with an error", func() {
			process = ifrit.Background(runner)
			Eventually(runner).Should(gexec.Exit(1))
			Eventually(runner).Should(gbytes.Say("kubeNodeName is a required flag"))
		})
	})

	Context("when the kubeCluster flag is missing", func() {
		BeforeEach(func() {
			args.KubeCluster = ""
			runner = createEndpointManagerRunner(args)
		})

		It("terminates with an error", func() {
			process = ifrit.Background(runner)
			Eventually(runner).Should(gexec.Exit(1))
			Eventually(runner).Should(gbytes.Say("kubeCluster is a required flag"))
		})
	})
})
