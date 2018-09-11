package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/jessfraz/bpfd/api/grpc"
	"github.com/jessfraz/bpfd/rules"
	"github.com/sirupsen/logrus"
)

const createHelp = `Create one or more rules.`

func (cmd *createCommand) Name() string      { return "create" }
func (cmd *createCommand) Args() string      { return "[OPTIONS] RULE_FILE [RULE_FILE...]" }
func (cmd *createCommand) ShortHelp() string { return createHelp }
func (cmd *createCommand) LongHelp() string  { return createHelp }
func (cmd *createCommand) Hidden() bool      { return false }

func (cmd *createCommand) Register(fs *flag.FlagSet) {
}

type createCommand struct {
}

func (cmd *createCommand) Run(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return errors.New("must pass at least one rule file")
	}

	// Create the grpc client.
	c, err := getClient(ctx, grpcAddress)
	if err != nil {
		return err
	}

	prs, names, err := rules.Parse(args...)
	if err != nil {
		return err
	}
	logrus.Debugf("Creating rules: %s", strings.Join(names, ", "))

	// Create the rules.
	for _, rules := range prs {
		for _, rule := range rules {
			_, err := c.CreateRule(ctx, &grpc.CreateRuleRequest{
				Rule: &rule,
			})
			if err != nil {
				return fmt.Errorf("sending CreateRule request for name %s failed: %v", rule.Name, err)
			}

			fmt.Println(rule.Name)
		}
	}

	return nil
}
