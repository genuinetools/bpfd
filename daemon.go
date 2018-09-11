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
	rules, names, err := rules.Parse(files...)
	if err != nil {
		return fmt.Errorf("reading rules files from directory %s failed: %v", cmd.rulesDirectory, err)
	}
	logrus.Infof("Loaded rules: %s", strings.Join(names, ", "))

	// Start the grpc api server.
	l, err := net.Listen("unix", grpcAddress)
	if err != nil {
		return fmt.Errorf("starting listener at %s failed: %v", grpcAddress, err)
	}
	s := grpc.NewServer()
	svr, err := api.NewServer(rules)
	if err != nil {
		return fmt.Errorf("creating new api server failed: %v", err)
	}
	types.RegisterAPIServer(s, svr)

	logrus.Infof("gRPC api server listening on %s", grpcAddress)

	return s.Serve(l)
}
