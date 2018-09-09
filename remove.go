package main

import (
	"context"
	"errors"
	"flag"
)

const removeHelp = `Remove one or more programs.`

func (cmd *removeCommand) Name() string      { return "rm" }
func (cmd *removeCommand) Args() string      { return "[OPTIONS] PROGRAM [PROGRAM...]" }
func (cmd *removeCommand) ShortHelp() string { return removeHelp }
func (cmd *removeCommand) LongHelp() string  { return removeHelp }
func (cmd *removeCommand) Hidden() bool      { return false }

func (cmd *removeCommand) Register(fs *flag.FlagSet) {
}

type removeCommand struct {
}

func (cmd *removeCommand) Run(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return errors.New("must pass at least one program")
	}

	return nil
}
