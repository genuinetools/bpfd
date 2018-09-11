package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessfraz/bpfd/api/grpc"
	"github.com/sirupsen/logrus"
)

const traceHelp = `Live trace the events returned after filtering.`
const longHelp = traceHelp + "\n\nThis does not include past events. Consider it like a tail."

func (cmd *traceCommand) Name() string      { return "trace" }
func (cmd *traceCommand) Args() string      { return "[OPTIONS]" }
func (cmd *traceCommand) ShortHelp() string { return traceHelp }
func (cmd *traceCommand) LongHelp() string  { return longHelp }
func (cmd *traceCommand) Hidden() bool      { return false }

func (cmd *traceCommand) Register(fs *flag.FlagSet) {
}

type traceCommand struct {
}

func (cmd *traceCommand) Run(ctx context.Context, args []string) error {
	tracing := true

	// On ^C, or SIGTERM handle exit.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for sig := range c {
			tracing = false
			logrus.Infof("Received %s, exiting", sig.String())
			os.Exit(0)
		}
	}()

	// Create the grpc client.
	client, err := getClient(ctx, grpcAddress)
	if err != nil {
		return err
	}

	// Get the events from the stream.
	stream, err := client.LiveTrace(context.Background(), &grpc.LiveTraceRequest{})
	if err != nil {
		return fmt.Errorf("sending LiveTrace request failed: %v", err)
	}

	for tracing {
		event, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("receiving event from stream failed: %v", err)
		}

		if event == nil || event.Data == nil || len(event.Data) < 1 {
			// continue the loop
			continue
		}

		logrus.WithFields(logrus.Fields{
			"program":           event.Program,
			"pid":               fmt.Sprintf("%d", event.PID),
			"tgid":              fmt.Sprintf("%d", event.TGID),
			"container_runtime": string(event.ContainerRuntime),
			"container_id":      event.ContainerID,
		}).Infof("%#v", event.Data)
	}

	return nil
}
