package iptables_test

import (
	"errors"
	"os/exec"

	"code.cloudfoundry.org/lager/lagertest"

	"github.com/cf-furnace/controller/routing/iptables"
	"github.com/cf-furnace/controller/routing/iptables/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IPTables", func() {
	var (
		runner *fakes.FakeCommandRunner
		logger *lagertest.TestLogger
		ipt    iptables.IPTables
	)

	BeforeEach(func() {
		runner = &fakes.FakeCommandRunner{}
		logger = lagertest.NewTestLogger("test")
		ipt = iptables.New(logger, "/bin/echo", runner)
	})

	Describe("CreateChain", func() {
		It("runs the command with appropriate arguments", func() {
			err := ipt.CreateChain(iptables.NAT, "chain-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(runner.RunCallCount()).To(Equal(1))
			cmd := runner.RunArgsForCall(0)
			Expect(cmd.Path).To(Equal("/bin/echo"))
			Expect(cmd.Args).To(Equal([]string{"/bin/echo", "--wait", "--table", "nat", "-N", "chain-name"}))
		})

		Context("when run fails with an ExitError", func() {
			BeforeEach(func() {
				err := exec.Command("false").Run()
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitErr.Stderr = []byte("standard error message")
				}
				runner.RunReturns(err)
			})

			It("returns a meaningful error", func() {
				err := ipt.CreateChain(iptables.NAT, "chain-name")
				Expect(err).To(MatchError("create-chain: (1) standard error message"))
			})
		})

		Context("when run fails", func() {
			BeforeEach(func() {
				runner.RunReturns(errors.New("woops"))
			})

			It("returns a context specific error", func() {
				err := ipt.CreateChain(iptables.NAT, "chain-name")
				Expect(err).To(MatchError("create-chain: woops"))
			})
		})
	})

	Describe("DeleteChain", func() {
		It("runs the command with appropriate arguments", func() {
			err := ipt.DeleteChain(iptables.NAT, "chain-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(runner.RunCallCount()).To(Equal(1))
			cmd := runner.RunArgsForCall(0)
			Expect(cmd.Path).To(Equal("/bin/echo"))
			Expect(cmd.Args).To(Equal([]string{"/bin/echo", "--wait", "--table", "nat", "-X", "chain-name"}))
		})

		Context("when run fails with an ExitError", func() {
			BeforeEach(func() {
				err := exec.Command("false").Run()
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitErr.Stderr = []byte("standard error message")
				}
				runner.RunReturns(err)
			})

			It("returns a meaningful error", func() {
				err := ipt.DeleteChain(iptables.NAT, "chain-name")
				Expect(err).To(MatchError("delete-chain: (1) standard error message"))
			})
		})

		Context("when run fails", func() {
			BeforeEach(func() {
				runner.RunReturns(errors.New("woops"))
			})

			It("returns a context specific error", func() {
				err := ipt.DeleteChain(iptables.NAT, "chain-name")
				Expect(err).To(MatchError("delete-chain: woops"))
			})
		})
	})

	Describe("DeleteChainReferencesReferences", func() {
		It("runs the command with appropriate arguments", func() {
			err := ipt.DeleteChainReferences(iptables.NAT, "chain-name", "referenced-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(runner.RunCallCount()).To(Equal(1))
			cmd := runner.RunArgsForCall(0)
			Expect(cmd.Path).To(Equal("/bin/sh"))
			Expect(cmd.Args).To(Equal([]string{
				"/bin/sh",
				"-ce",
				"/bin/echo --wait --table nat -S chain-name | grep 'referenced-name' | sed -e 's/-A/-D/' | xargs --no-run-if-empty --max-lines=1 /bin/echo --wait --table nat"}))
		})

		Context("when run fails with an ExitError", func() {
			BeforeEach(func() {
				err := exec.Command("false").Run()
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitErr.Stderr = []byte("standard error message")
				}
				runner.RunReturns(err)
			})

			It("returns a meaningful error", func() {
				err := ipt.DeleteChainReferences(iptables.NAT, "chain-name", "referenced-name")
				Expect(err).To(MatchError("delete-chain-references: (1) standard error message"))
			})
		})

		Context("when run fails", func() {
			BeforeEach(func() {
				runner.RunReturns(errors.New("woops"))
			})

			It("returns a context specific error", func() {
				err := ipt.DeleteChainReferences(iptables.NAT, "chain-name", "referenced-name")
				Expect(err).To(MatchError("delete-chain-references: woops"))
			})
		})
	})

	Describe("FlushChain", func() {
		It("runs the command with appropriate arguments", func() {
			err := ipt.FlushChain(iptables.NAT, "chain-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(runner.RunCallCount()).To(Equal(1))
			cmd := runner.RunArgsForCall(0)
			Expect(cmd.Path).To(Equal("/bin/echo"))
			Expect(cmd.Args).To(Equal([]string{"/bin/echo", "--wait", "--table", "nat", "-F", "chain-name"}))
		})

		Context("when run fails with an ExitError", func() {
			BeforeEach(func() {
				err := exec.Command("false").Run()
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitErr.Stderr = []byte("standard error message")
				}
				runner.RunReturns(err)
			})

			It("returns a meaningful error", func() {
				err := ipt.FlushChain(iptables.NAT, "chain-name")
				Expect(err).To(MatchError("flush-chain: (1) standard error message"))
			})
		})

		Context("when run fails", func() {
			BeforeEach(func() {
				runner.RunReturns(errors.New("woops"))
			})

			It("returns a context specific error", func() {
				err := ipt.FlushChain(iptables.NAT, "chain-name")
				Expect(err).To(MatchError("flush-chain: woops"))
			})
		})
	})

	Describe("ListChains", func() {
		It("runs the command with appropriate arguments", func() {
			_, err := ipt.ListChains(iptables.NAT)
			Expect(err).NotTo(HaveOccurred())

			Expect(runner.RunCallCount()).To(Equal(1))

			cmd := runner.RunArgsForCall(0)
			Expect(cmd.Path).To(Equal("/bin/sh"))
			Expect(cmd.Args).To(Equal([]string{
				"/bin/sh",
				"-ce",
				"/bin/echo --wait --table nat -S | grep '^-A' | cut -f2 -d' '"}))
		})

		It("returns the chains in as a slice", func() {
			runner.RunStub = func(cmd *exec.Cmd) error {
				cmd.Stdout.Write([]byte("chain-one\nchain-two\nchain-three\n"))
				return nil
			}

			result, err := ipt.ListChains(iptables.NAT)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(ConsistOf("chain-one", "chain-two", "chain-three"))
		})

		Context("when run fails with an ExitError", func() {
			BeforeEach(func() {
				err := exec.Command("false").Run()
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitErr.Stderr = []byte("standard error message")
				}
				runner.RunReturns(err)
			})

			It("returns a meaningful error", func() {
				_, err := ipt.ListChains(iptables.NAT)
				Expect(err).To(MatchError("list-chains: (1) standard error message"))
			})
		})

		Context("when run fails", func() {
			BeforeEach(func() {
				runner.RunReturns(errors.New("woops"))
			})

			It("returns a context specific error", func() {
				_, err := ipt.ListChains(iptables.NAT)
				Expect(err).To(MatchError("list-chains: woops"))
			})
		})
	})

	Describe("PrependRule", func() {
		var rule *iptables.JumpRule

		BeforeEach(func() {
			rule = &iptables.JumpRule{
				TargetChain: "target-chain",
			}
		})

		It("runs the command with appropriate arguments", func() {
			err := ipt.PrependRule(iptables.NAT, "chain-name", rule)
			Expect(err).NotTo(HaveOccurred())

			Expect(runner.RunCallCount()).To(Equal(1))
			cmd := runner.RunArgsForCall(0)
			Expect(cmd.Path).To(Equal("/bin/echo"))
			Expect(cmd.Args).To(Equal([]string{"/bin/echo", "--wait", "--table", "nat", "-I", "chain-name", "1", "--jump", "target-chain"}))
		})

		Context("when run fails with an ExitError", func() {
			BeforeEach(func() {
				err := exec.Command("false").Run()
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitErr.Stderr = []byte("standard error message")
				}
				runner.RunReturns(err)
			})

			It("returns a meaningful error", func() {
				err := ipt.PrependRule(iptables.NAT, "chain-name", rule)
				Expect(err).To(MatchError("insert-rule: (1) standard error message"))
			})
		})

		Context("when run fails", func() {
			BeforeEach(func() {
				runner.RunReturns(errors.New("woops"))
			})

			It("returns a context specific error", func() {
				err := ipt.PrependRule(iptables.NAT, "chain-name", rule)
				Expect(err).To(MatchError("insert-rule: woops"))
			})
		})
	})

	Describe("AppendRule", func() {
		var rule *iptables.JumpRule

		BeforeEach(func() {
			rule = &iptables.JumpRule{
				TargetChain: "target-chain",
			}
		})

		It("runs the command with appropriate arguments", func() {
			err := ipt.AppendRule(iptables.NAT, "chain-name", rule)
			Expect(err).NotTo(HaveOccurred())

			Expect(runner.RunCallCount()).To(Equal(1))
			cmd := runner.RunArgsForCall(0)
			Expect(cmd.Path).To(Equal("/bin/echo"))
			Expect(cmd.Args).To(Equal([]string{"/bin/echo", "--wait", "--table", "nat", "-A", "chain-name", "--jump", "target-chain"}))
		})

		Context("when run fails with an ExitError", func() {
			BeforeEach(func() {
				err := exec.Command("false").Run()
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitErr.Stderr = []byte("standard error message")
				}
				runner.RunReturns(err)
			})

			It("returns a meaningful error", func() {
				err := ipt.AppendRule(iptables.NAT, "chain-name", rule)
				Expect(err).To(MatchError("append-rule: (1) standard error message"))
			})
		})

		Context("when run fails", func() {
			BeforeEach(func() {
				runner.RunReturns(errors.New("woops"))
			})

			It("returns a context specific error", func() {
				err := ipt.AppendRule(iptables.NAT, "chain-name", rule)
				Expect(err).To(MatchError("append-rule: woops"))
			})
		})
	})
})
