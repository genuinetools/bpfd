package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/genuinetools/bpfd/api/grpc"
	"github.com/sirupsen/logrus"
)

const removeHelp = `Remove one or more rules.`

func (cmd *removeCommand) Name() string      { return "rm" }
func (cmd *removeCommand) Args() string      { return "[OPTIONS] RULE_NAME [RULE_NAME...]" }
func (cmd *removeCommand) ShortHelp() string { return removeHelp }
func (cmd *removeCommand) LongHelp() string  { return removeHelp }
func (cmd *removeCommand) Hidden() bool      { return false }

func (cmd *removeCommand) Register(fs *flag.FlagSet) {
}

type removeCommand struct {
}

func (cmd *removeCommand) Run(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return errors.New("must pass at least one rule")
	}

	// Create the grpc client.
	c, err := getClient(ctx, grpcAddress)
	if err != nil {
		return err
	}
	logrus.Debugf("Removing rules: %s", strings.Join(args, ", "))

	// Remove the rules.
	for _, name := range args {
		_, err := c.RemoveRule(ctx, &grpc.RemoveRuleRequest{
			Name: name,
		})
		if err != nil {
			return fmt.Errorf("sending RemoveRule request for name %s failed: %v", name, err)
		}
		fmt.Println(name)
	}

	return nil
}
