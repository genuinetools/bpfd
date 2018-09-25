package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/genuinetools/bpfd/api/grpc"
)

const listHelp = `List rules.`

func (cmd *listCommand) Name() string      { return "ls" }
func (cmd *listCommand) Args() string      { return "[OPTIONS]" }
func (cmd *listCommand) ShortHelp() string { return listHelp }
func (cmd *listCommand) LongHelp() string  { return listHelp }
func (cmd *listCommand) Hidden() bool      { return false }

func (cmd *listCommand) Register(fs *flag.FlagSet) {
}

type listCommand struct {
}

func (cmd *listCommand) Run(ctx context.Context, args []string) error {
	// Create the grpc client.
	c, err := getClient(ctx, grpcAddress)
	if err != nil {
		return err
	}

	// List the rules.
	resp, err := c.ListRules(context.Background(), &grpc.ListRulesRequest{})
	if err != nil {
		return fmt.Errorf("sending ListRules request failed: %v", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	fmt.Fprint(w, "NAME\tTRACER\n")

	for _, rule := range resp.Rules {
		fmt.Fprintf(w, "%s\t%s\n", rule.Name, rule.Tracer)
	}

	w.Flush()
	return nil
}
