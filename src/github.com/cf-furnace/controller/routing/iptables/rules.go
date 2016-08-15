package iptables

import (
	"fmt"
)

type DNATRule struct {
	HostAddress      string
	HostPort         int32
	ContainerAddress string
	ContainerPort    int32
}

func (r *DNATRule) Flags() []string {
	return []string{
		"--protocol", "tcp",
		"--destination", r.HostAddress,
		"--destination-port", fmt.Sprintf("%d", r.HostPort),
		"--jump", "DNAT",
		"--to-destination", fmt.Sprintf("%s:%d", r.ContainerAddress, r.ContainerPort),
	}
}

type JumpRule struct {
	TargetChain string
}

func (r *JumpRule) Flags() []string {
	return []string{"--jump", r.TargetChain}
}
