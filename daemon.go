package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jessfraz/bpfd/program"
	"github.com/sirupsen/logrus"

	// register the bashreadline program
	_ "github.com/jessfraz/bpfd/program/bashreadline"
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
	// On ^C, or SIGTERM handle exit.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for sig := range c {
			logrus.Infof("Received %s, exiting", sig.String())
			logrus.Info("Gracefully shutting down and unloading all programs")
			program.UnloadAll()
			os.Exit(0)
		}
	}()

	daemon := make(chan bool)

	// List all the compiled in programs.
	programs := program.List()
	logrus.Infof("Daemon compiled with programs: %s", strings.Join(programs, ", "))

	// Load all the compiled in programs.
	for _, p := range programs {
		// We can ignore the error below since we are using the list from our code
		// so the program has to exist in the map.
		prog, _ := program.Get(p)
		if err := prog.Load(); err != nil {
			return fmt.Errorf("loading program %s failed: %v", p, err)
		}

		// Watch the events for the program.
		if err := prog.WatchEvents(); err != nil {
			return fmt.Errorf("starting watch events for program %s failed: %v", p, err)
		}
		logrus.Infof("Watching events for plugin %s", p)
	}

	<-daemon // Block forever
	return nil
}
