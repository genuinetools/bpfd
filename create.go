package main

import (
	"context"
	"errors"
	"flag"
)

const createHelp = `Create one or more rules.`

func (cmd *createCommand) Name() string      { return "rm" }
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

	return nil
}
