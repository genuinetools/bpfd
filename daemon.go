package main

import (
	"context"
	"flag"
	"strings"

	"github.com/jessfraz/bpfd/plugin"
	"github.com/sirupsen/logrus"
)

const daemonHelp = `Start the daemon.`

func (cmd *daemonCommand) Name() string      { return "daemon" }
func (cmd *daemonCommand) Args() string      { return "[OPTIONS]" }
func (cmd *daemonCommand) ShortHelp() string { return daemonHelp }
func (cmd *daemonCommand) LongHelp() string  { return daemonHelp }
func (cmd *daemonCommand) Hidden() bool      { return false }

func (cmd *daemonCommand) Register(fs *flag.FlagSet) {
}

type daemonCommand struct {
}

func (cmd *daemonCommand) Run(ctx context.Context, args []string) error {
	// List all the compiled in programs.
	programs := plugin.List()
	logrus.Infof("Daemon compiled with programs: %s", strings.Join(programs, ", "))

	return nil
}
