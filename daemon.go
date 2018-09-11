package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/jessfraz/bpfd/action"
	"github.com/jessfraz/bpfd/api"
	types "github.com/jessfraz/bpfd/api/grpc"
	"github.com/jessfraz/bpfd/program"
	"github.com/jessfraz/bpfd/rules"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	// Register the builtin programs.
	_ "github.com/jessfraz/bpfd/program/bashreadline"
	_ "github.com/jessfraz/bpfd/program/exec"
	_ "github.com/jessfraz/bpfd/program/open"

	// Register the builtin actions.
	_ "github.com/jessfraz/bpfd/action/kill"
	_ "github.com/jessfraz/bpfd/action/stdout"
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
	rls, names, err := rules.ParseFiles(files...)
	if err != nil {
		return fmt.Errorf("reading rules files from directory %s failed: %v", cmd.rulesDirectory, err)
	}
	logrus.Infof("Loaded rules: %s", strings.Join(names, ", "))

	// List all the compiled in programs.
	programList := program.List()
	logrus.Infof("Daemon compiled with programs: %s", strings.Join(programList, ", "))
	programs := map[string]program.Program{}
	for _, p := range programList {
		prog, err := program.Get(p)
		if err != nil {
			return err
		}
		programs[p] = prog
	}
	logrus.Debugf("programs: %#v", programs)

	// List all the compiled in actions.
	actionList := action.List()
	logrus.Infof("Daemon compiled with actions: %s", strings.Join(actionList, ", "))
	actions := map[string]action.Action{}
	for _, a := range actionList {
		acn, err := action.Get(a)
		if err != nil {
			return err
		}
		actions[a] = acn
	}
	logrus.Debugf("actions: %#v", actions)

	// Validate the rules against the programs and actions.
	for _, prs := range rls {
		for _, r := range prs {
			if err := rules.ValidateProgramsAndActions(r, programList, actionList); err != nil {
				return err
			}
		}
	}

	// Create the directory if it doesn't exist.
	if err := os.MkdirAll(filepath.Dir(grpcAddress), 0755); err != nil {
		return fmt.Errorf("creating directory %s failed: %v", filepath.Dir(grpcAddress), err)
	}

	// Remove the old socket.
	if err := os.RemoveAll(grpcAddress); err != nil {
		logrus.Warnf("attempt to remove old sock %s failed: %v", grpcAddress, err)
	}

	// Start the grpc api server.
	l, err := net.Listen("unix", grpcAddress)
	if err != nil {
		return fmt.Errorf("starting listener at %s failed: %v", grpcAddress, err)
	}
	s := grpc.NewServer()
	opt := api.Opts{
		Rules:       rls,
		Programs:    programs,
		Actions:     actions,
		ProgramList: programList,
		ActionList:  actionList,
	}
	svr, err := api.NewServer(opt)
	if err != nil {
		return fmt.Errorf("creating new api server failed: %v", err)
	}
	types.RegisterAPIServer(s, svr)

	logrus.Infof("gRPC api server listening on %s", grpcAddress)

	return s.Serve(l)
}
