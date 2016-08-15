package iptables_test

import (
	"encoding/base32"
	"errors"
	"fmt"
	"strings"

	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/types"

	"github.com/cf-furnace/controller/routing/iptables"
	"github.com/cf-furnace/controller/routing/iptables/fakes"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GlobalChains", func() {
	var (
		ipt *fakes.FakeIPTables
	)

	BeforeEach(func() {
		ipt = &fakes.FakeIPTables{}
	})

	Describe("SetupNAT", func() {
		It("tears down the infrastructure first", func() {
			ipt.DeleteChainStub = func(table iptables.Table, chain string) error {
				Expect(ipt.CreateChainCallCount()).To(Equal(0))
				return nil
			}

			err := iptables.SetupNAT(ipt)
			Expect(err).NotTo(HaveOccurred())
			Expect(ipt.DeleteChainCallCount()).To(Equal(1))
			Expect(ipt.CreateChainCallCount()).To(Equal(1))
		})

		It("creates the furnace prerouting chain", func() {
			err := iptables.SetupNAT(ipt)
			Expect(err).NotTo(HaveOccurred())

			Expect(ipt.CreateChainCallCount()).To(Equal(1))
			table, chain := ipt.CreateChainArgsForCall(0)
			Expect(table).To(Equal(iptables.NAT))
			Expect(chain).To(Equal(iptables.FurnacePreroutingChain))
		})

		It("inserts the furnace prerouting chain into the system prerouting chain", func() {
			err := iptables.SetupNAT(ipt)
			Expect(err).NotTo(HaveOccurred())

			Expect(ipt.PrependRuleCallCount()).To(Equal(1))
			table, systemChain, rule := ipt.PrependRuleArgsForCall(0)
			Expect(table).To(Equal(iptables.NAT))
			Expect(systemChain).To(Equal(iptables.SystemPreroutingChain))
			Expect(rule).To(Equal(&iptables.JumpRule{TargetChain: iptables.FurnacePreroutingChain}))
		})

		Context("when tearing down the infrastructure fails", func() {
			BeforeEach(func() {
				ipt.DeleteChainReferencesReturns(errors.New("welp"))
			})

			It("propagates the error", func() {
				err := iptables.SetupNAT(ipt)
				Expect(err).To(MatchError("welp"))
			})
		})

		Context("when creating the furnace prerouting chain fails", func() {
			BeforeEach(func() {
				ipt.CreateChainReturns(errors.New("oh noes"))
			})

			It("propagates the error", func() {
				err := iptables.SetupNAT(ipt)
				Expect(err).To(MatchError("oh noes"))
			})
		})

		Context("when prepending to the prerouting chain fails", func() {
			BeforeEach(func() {
				ipt.PrependRuleReturns(errors.New("mango"))
			})

			It("propagates the error", func() {
				err := iptables.SetupNAT(ipt)
				Expect(err).To(MatchError("mango"))
			})
		})
	})

	Describe("InstanceChainName", func() {
		var (
			guid string
			pod  *v1.Pod
		)

		BeforeEach(func() {
			guid = "7daf005b-53fc-11e6-8ae5-5a1d4b8013d9"

			pod = &v1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-pod-name",
					UID:  types.UID(guid),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Name: "my-app-container",
						Ports: []v1.ContainerPort{{
							Name:          "irrelevant",
							ContainerPort: 8080,
							Protocol:      v1.ProtocolTCP,
						}},
					}},
				},
			}
		})

		It("generates a chain name from the pod UID", func() {
			name, err := iptables.InstanceChainName("w-", pod)
			Expect(err).NotTo(HaveOccurred())

			guid, err := uuid.ParseHex(string(pod.ObjectMeta.UID))
			Expect(err).NotTo(HaveOccurred())

			shortGuid := strings.TrimRight(base32.StdEncoding.EncodeToString(guid[:]), "=")
			Expect(name).To(Equal("w-" + shortGuid))
		})

		Context("when the prefix is too long", func() {
			It("returns an error", func() {
				_, err := iptables.InstanceChainName("too-long-", pod)
				Expect(err).To(MatchError("chain prefix too long: too-long-"))
			})
		})
	})

	Describe("TeardownNAT", func() {
		BeforeEach(func() {
			chainList := []string{
				"k-chain-1",
				"k-chain-2",
				"unrleated-chain",
				"k-chain-3",
			}
			ipt.ListChainsReturns(chainList, nil)
		})

		It("deletes references to FurnacePreroutingChain from PREROUTING", func() {
			err := iptables.TeardownNAT(ipt)
			Expect(err).NotTo(HaveOccurred())

			Expect(ipt.DeleteChainReferencesCallCount()).To(Equal(1))
			table, systemChain, furnaceChain := ipt.DeleteChainReferencesArgsForCall(0)
			Expect(table).To(Equal(iptables.NAT))
			Expect(systemChain).To(Equal(iptables.SystemPreroutingChain))
			Expect(furnaceChain).To(Equal(iptables.FurnacePreroutingChain))
		})

		It("lists all chains in the NAT table and removes those starting with the instance prefix", func() {
			err := iptables.TeardownNAT(ipt)
			Expect(err).NotTo(HaveOccurred())

			Expect(ipt.ListChainsCallCount()).To(Equal(1))
			Expect(ipt.ListChainsArgsForCall(0)).To(Equal(iptables.NAT))

			Expect(ipt.FlushChainCallCount()).To(Equal(4))
			Expect(ipt.DeleteChainCallCount()).To(Equal(4))

			for i := 0; i < 3; i++ {
				flushTable, flushChain := (ipt.FlushChainArgsForCall(i))
				Expect(flushTable).To(Equal(iptables.NAT))
				Expect(flushChain).To(Equal(fmt.Sprintf("k-chain-%d", i+1)))

				deleteTable, deleteChain := ipt.DeleteChainArgsForCall(i)
				Expect(deleteTable).To(Equal(iptables.NAT))
				Expect(deleteChain).To(Equal(fmt.Sprintf("k-chain-%d", i+1)))
			}

			flushTable, flushChain := ipt.FlushChainArgsForCall(3)
			Expect(flushTable).To(Equal(iptables.NAT))
			Expect(flushChain).To(Equal(iptables.FurnacePreroutingChain))

			deleteTable, deleteChain := ipt.DeleteChainArgsForCall(3)
			Expect(deleteTable).To(Equal(iptables.NAT))
			Expect(deleteChain).To(Equal(iptables.FurnacePreroutingChain))
		})

		Context("when deleting chain references fails", func() {
			BeforeEach(func() {
				ipt.DeleteChainReferencesReturns(errors.New("woops"))
			})

			It("propagages the error", func() {
				err := iptables.TeardownNAT(ipt)
				Expect(err).To(MatchError("woops"))
			})
		})

		Context("when flushing an instance chain fails", func() {
			BeforeEach(func() {
				ipt.FlushChainStub = func(table iptables.Table, chain string) error {
					if strings.HasPrefix(chain, "k-") {
						return errors.New("fail the flush: " + chain)
					}
					return nil
				}
			})

			It("propagates the error", func() {
				err := iptables.TeardownNAT(ipt)
				Expect(err).To(MatchError("fail the flush: k-chain-1"))
			})
		})

		Context("when deleting an instance chain fails", func() {
			BeforeEach(func() {
				ipt.DeleteChainStub = func(table iptables.Table, chain string) error {
					if strings.HasPrefix(chain, "k-") {
						return errors.New("fail the delete: " + chain)
					}
					return nil
				}
			})

			It("propagates the error", func() {
				err := iptables.TeardownNAT(ipt)
				Expect(err).To(MatchError("fail the delete: k-chain-1"))
			})
		})

		Context("when flushing the furnace prerouting chain fails", func() {
			BeforeEach(func() {
				ipt.FlushChainStub = func(table iptables.Table, chain string) error {
					if chain == iptables.FurnacePreroutingChain {
						return errors.New("fail the flush: " + chain)
					}
					return nil
				}
			})

			It("keeps calm and carries on", func() {
				err := iptables.TeardownNAT(ipt)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when deleting the furnace prerouting chain fails", func() {
			BeforeEach(func() {
				ipt.DeleteChainStub = func(table iptables.Table, chain string) error {
					if chain == iptables.FurnacePreroutingChain {
						return errors.New("fail the delete: " + chain)
					}
					return nil
				}
			})

			It("keeps calm and carries on", func() {
				err := iptables.TeardownNAT(ipt)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
