package main

import (
	"context"
	"errors"
	"flag"

	"github.com/genuinetools/pkg/cli"
	"github.com/jessfraz/ship/version"
	"github.com/sirupsen/logrus"
)

var (
	grpcAddress string

	debug bool
)

func main() {
	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "bpfd"
	p.Description = "Framework for running BPF programs with rules on Linux as a daemon"
	// Set the GitCommit and Version.
	p.GitCommit = version.GITCOMMIT
	p.Version = version.VERSION

	// Build the list of available commands.
	p.Commands = []cli.Command{
		&createCommand{},
		&daemonCommand{},
		&listCommand{},
		&removeCommand{},
	}

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("bpfd", flag.ExitOnError)
	p.FlagSet.BoolVar(&debug, "debug", false, "enable debug logging")
	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")
	p.FlagSet.StringVar(&grpcAddress, "grpc-addr", "/run/bpfd/bpfd.sock", "Address for gRPC api communication")

	// Set the before function.
	p.Before = func(ctx context.Context) error {
		// Set the log level.
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		if len(grpcAddress) < 1 {
			return errors.New("gRPC address cannot be empty")
		}

		return nil
	}

	// Run our program.
	p.Run()
}
