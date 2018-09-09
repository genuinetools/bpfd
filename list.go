package main

import (
	"context"
	"flag"
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
	return nil
}
