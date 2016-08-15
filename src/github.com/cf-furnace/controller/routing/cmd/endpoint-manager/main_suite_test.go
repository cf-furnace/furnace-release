package main_test

import (
	"encoding/json"
	"testing"
	"time"

	"code.cloudfoundry.org/lager/lagertest"
	"github.com/cf-furnace/controller/routing/cmd/endpoint-manager/testrunner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit/ginkgomon"
)

const (
	heartbeatInterval         = 1 * time.Second
	endpointManagerPortOffset = 15001

	DEBUG = iota
)

var (
	endpointManagerPath string

	logger *lagertest.TestLogger
)

func TestEndpointManager(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(2 * time.Second)
	RunSpecs(t, "Route Emitter Suite")
}

func portOffset(portType int) int {
	switch portType {
	case DEBUG:
		return endpointManagerPortOffset + DEBUG
	default:
		return 0
	}
}

func createEndpointManagerRunner(args testrunner.Args, extraArgs ...string) *ginkgomon.Runner {
	return testrunner.New(endpointManagerPath, args, extraArgs...)
}

var _ = SynchronizedBeforeSuite(func() []byte {
	endpointManager, err := gexec.Build("github.com/cf-furnace/controller/routing/cmd/endpoint-manager", "-race")
	Expect(err).NotTo(HaveOccurred())

	payload, err := json.Marshal(map[string]string{
		"endpoint-manager": endpointManager,
	})

	Expect(err).NotTo(HaveOccurred())

	return payload
}, func(payload []byte) {
	binaries := map[string]string{}

	err := json.Unmarshal(payload, &binaries)
	Expect(err).NotTo(HaveOccurred())

	endpointManagerPath = string(binaries["endpoint-manager"])

	logger = lagertest.NewTestLogger("test")
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})
