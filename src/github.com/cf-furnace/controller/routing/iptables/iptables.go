package iptables

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"github.com/cloudfoundry/gunk/command_runner"
)

//go:generate counterfeiter -o fakes/fake_iptables.go . IPTables
type IPTables interface {
	CreateChain(table Table, name string) error
	DeleteChain(table Table, name string) error
	DeleteChainReferences(table Table, chain, referenced string) error
	FlushChain(table Table, name string) error
	ListChains(table Table) ([]string, error)

	PrependRule(table Table, chain string, rule Rule) error
	AppendRule(table Table, chain string, rule Rule) error
}

//go:generate counterfeiter -o fakes/fake_command_runner.go . commandRunner
type commandRunner interface {
	command_runner.CommandRunner
}

type Table string

const (
	NAT Table = "nat"
)

func (t Table) String() string {
	return string(t)
}

type Rule interface {
	Flags() []string
}

type iptables struct {
	commandPath string
	runner      command_runner.CommandRunner
}

func New(commandPath string, runner command_runner.CommandRunner) IPTables {
	return &iptables{
		commandPath: commandPath,
		runner:      runner,
	}
}

func (i *iptables) run(action string, args ...string) error {
	cmd := exec.Command(i.commandPath, args...)
	return i.runCommand(action, cmd)
}

func (i *iptables) runCommand(action string, command *exec.Cmd) error {
	if err := i.runner.Run(command); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			returnCode := -65535
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				returnCode = status.ExitStatus()
			}
			return fmt.Errorf("%s: (%d) %s", action, returnCode, exiterr.Stderr)
		} else {
			return fmt.Errorf("%s: %s", action, err.Error())
		}
	}

	return nil
}

func (i *iptables) CreateChain(table Table, chain string) error {
	return i.run("create-chain", "--wait", "--table", table.String(), "-N", chain)
}

func (i *iptables) DeleteChain(table Table, chain string) error {
	return i.run("delete-chain", "--wait", "--table", table.String(), "-X", chain)
}

func (i *iptables) DeleteChainReferences(table Table, chain, referenced string) error {
	command := exec.Command(
		"/bin/sh", "-ce",
		strings.Join([]string{
			i.commandPath, "--wait", "--table", table.String(), "-S", chain, "|",
			"grep", "'" + referenced + "'", "|",
			"sed", "-e", "'s/-A/-D/'", "|",
			"xargs", "--no-run-if-empty", "--max-lines=1", i.commandPath, "--wait", "--table", table.String(),
		}, " "),
	)

	return i.runCommand("delete-chain-references", command)
}

func (i *iptables) FlushChain(table Table, chain string) error {
	return i.run("flush-chain", "--wait", "--table", table.String(), "-F", chain)
}

func (i *iptables) ListChains(table Table) ([]string, error) {
	command := exec.Command(
		"/bin/sh", "-ce",
		strings.Join([]string{
			i.commandPath, "--wait", "--table", table.String(), "-S", "|",
			"grep", "'^-A'", "|",
			"cut", "-f2", "-d' '",
		}, " "),
	)

	stdout := &bytes.Buffer{}
	command.Stdout = stdout

	err := i.runCommand("list-chains", command)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(stdout.String(), "\n")

	return lines[:len(lines)-1], nil
}

func (i *iptables) PrependRule(table Table, chain string, rule Rule) error {
	return i.run("insert-rule", append([]string{"--wait", "--table", table.String(), "-I", chain, "1"}, rule.Flags()...)...)
}

func (i *iptables) AppendRule(table Table, chain string, rule Rule) error {
	return i.run("append-rule", append([]string{"--wait", "--table", table.String(), "-A", chain}, rule.Flags()...)...)
}
