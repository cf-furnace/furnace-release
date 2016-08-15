package iptables

import (
	"encoding/base32"
	"fmt"
	"strings"

	uuid "github.com/nu7hatch/gouuid"
	"k8s.io/kubernetes/pkg/api/v1"
)

const (
	FurnacePreroutingChain = "FURNACE_PREROUTING"
	SystemPreroutingChain  = "PREROUTING"

	InstanceChainPrefix = "k-"

	MaxChainNameLength = 28
)

func SetupNAT(ipt IPTables) error {
	if err := TeardownNAT(ipt); err != nil {
		return err
	}

	err := ipt.CreateChain(NAT, FurnacePreroutingChain)
	if err != nil {
		return err
	}

	err = ipt.PrependRule(NAT, SystemPreroutingChain, &JumpRule{TargetChain: FurnacePreroutingChain})
	if err != nil {
		return err
	}

	return nil
}

func TeardownNAT(ipt IPTables) error {
	err := ipt.DeleteChainReferences(NAT, SystemPreroutingChain, FurnacePreroutingChain)
	if err != nil {
		return err
	}

	chains, err := ipt.ListChains(NAT)
	if err != nil {
		return err
	}

	for _, chain := range chains {
		if strings.HasPrefix(chain, InstanceChainPrefix) {
			if err := ipt.FlushChain(NAT, chain); err != nil {
				return err
			}
			if err := ipt.DeleteChain(NAT, chain); err != nil {
				return err
			}
		}
	}

	// Ignore errors since the chain may not exist
	_ = ipt.FlushChain(NAT, FurnacePreroutingChain)
	_ = ipt.DeleteChain(NAT, FurnacePreroutingChain)

	return nil
}

func InstanceChainName(prefix string, pod *v1.Pod) (string, error) {
	guid, err := uuid.ParseHex(string(pod.ObjectMeta.UID))
	if err != nil {
		return "", err
	}

	shortGuid := strings.TrimRight(base32.StdEncoding.EncodeToString(guid[:]), "=")
	chainName := prefix + shortGuid

	if len(chainName) > MaxChainNameLength {
		return "", fmt.Errorf("chain prefix too long: %s", prefix)
	}

	return chainName, nil
}
