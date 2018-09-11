package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/jessfraz/bpfd/api/grpc"
	"github.com/jessfraz/bpfd/proc"
	"github.com/jessfraz/bpfd/program"
	"github.com/jessfraz/bpfd/rules"
	"github.com/sirupsen/logrus"

	// Register the builtin programs.
	_ "github.com/jessfraz/bpfd/program/bashreadline"
	_ "github.com/jessfraz/bpfd/program/exec"
	_ "github.com/jessfraz/bpfd/program/open"
)

const daemonHelp = `Start the daemon.`

func (cmd *daemonCommand) Name() string      { return "daemon" }
func (cmd *daemonCommand) Args() string      { return "[OPTIONS]" }
func (cmd *daemonCommand) ShortHelp() string { return daemonHelp }
func (cmd *daemonCommand) LongHelp() string  { return daemonHelp }
func (cmd *daemonCommand) Hidden() bool      { return false }

func (cmd *daemonCommand) Register(fs *flag.FlagSet) {
	fs.StringVar(&cmd.rulesDirectory, "rules-dir", "/etc/bpfd/rules", "Directory that stores the rules files")
}

type daemonCommand struct {
	rulesDirectory string
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

	// Get all the rules from the rule directory.
	fi, err := ioutil.ReadDir(cmd.rulesDirectory)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("listing files in rules directory %s failed: %v", cmd.rulesDirectory, err)
		}
	}
	files := []string{}
	for _, file := range fi {
		files = append(files, filepath.Join(cmd.rulesDirectory, file.Name()))
	}
	rules, names, err := rules.Parse(files...)
	if err != nil {
		return fmt.Errorf("reading rules files from directory %s failed: %v", cmd.rulesDirectory, err)
	}
	logrus.Infof("Loaded rules: %s", strings.Join(names, ", "))

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

		progRules, _ := rules[p]

		go func(p string, prog program.Program, progRules []grpc.Rule) {
			for {
				// Watch the events for the program.
				event, err := prog.WatchEvent()
				if err != nil {
					logrus.Warnf("watch event for program %s failed: %v", p, err)
				}

				if event == nil {
					continue
				}

				runtime := proc.GetContainerRuntime(int(event.TGID), int(event.PID))

				// Verify the event matches for the rules.
				if !program.Match(progRules, event.Data, runtime) {
					// We didn't find what we were searching for so continue.
					continue
				}

				logrus.WithFields(logrus.Fields{
					"program":           p,
					"pid":               fmt.Sprintf("%d", event.PID),
					"tgid":              fmt.Sprintf("%d", event.TGID),
					"container_runtime": string(runtime),
					"container_id":      proc.GetContainerID(int(event.TGID), int(event.PID)),
				}).Infof("%#v", event.Data)
			}
		}(p, prog, progRules)

		// Start the program.
		prog.Start()
		logrus.Infof("Watching events for plugin %s", p)
	}

	<-daemon // Block forever
	return nil
}
