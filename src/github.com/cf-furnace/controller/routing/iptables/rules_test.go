package iptables_test

import (
	"github.com/cf-furnace/controller/routing/iptables"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rules", func() {
	Describe("DNATRule", func() {
		It("returns appropriate DNAT flags", func() {
			rule := &iptables.DNATRule{
				HostAddress:      "10.11.12.13",
				HostPort:         60001,
				ContainerAddress: "1.2.3.4",
				ContainerPort:    8080,
			}

			Expect(rule.Flags()).To(ConsistOf(
				"--protocol", "tcp",
				"--destination", "10.11.12.13",
				"--destination-port", "60001",
				"--jump", "DNAT",
				"--to-destination", "1.2.3.4:8080",
			))
		})
	})

	Describe("JumpRule", func() {
		It("returns jump flags", func() {
			rule := &iptables.JumpRule{
				TargetChain: "target",
			}

			Expect(rule.Flags()).To(ConsistOf("--jump", "target"))
		})
	})
})
