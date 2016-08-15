package testrunner

import (
	"os/exec"
	"strconv"
	"time"

	"github.com/tedsuo/ifrit/ginkgomon"
)

type Args struct {
	DebugAddress   string
	DropsondePort  int
	LogLevel       string
	KubeCACert     string
	KubeClientCert string
	KubeClientKey  string
	KubeCluster    string
	KubeNodeName   string
	ResyncInterval time.Duration
}

func (a Args) ArgSlice() []string {
	args := []string{
		"-dropsondePort", strconv.Itoa(a.DropsondePort),
		"-kubeCACert", a.KubeCACert,
		"-kubeClientCert", a.KubeClientCert,
		"-kubeClientKey", a.KubeClientKey,
		"-kubeCluster", a.KubeCluster,
		"-kubeNodeName", a.KubeNodeName,
		"-resyncInterval", a.ResyncInterval.String(),
	}

	if a.DebugAddress != "" {
		args = append(args, "-debugAddr", a.DebugAddress)
	}

	if a.LogLevel != "" {
		args = append(args, "-logLevel", a.LogLevel)
	}

	return args
}

func New(binpath string, args Args, extraArgs ...string) *ginkgomon.Runner {
	return ginkgomon.New(ginkgomon.Config{
		Command:       exec.Command(binpath, append(args.ArgSlice(), extraArgs...)...),
		StartCheck:    "endpoint-manager.started",
		AnsiColorCode: "97m",
	})
}
