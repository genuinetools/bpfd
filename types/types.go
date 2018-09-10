package types

import (
	"github.com/jessfraz/bpfd/proc"
)

// Rule defines a rule to notify on.
type Rule struct {
	Name              string            `toml:"name,omitempty"`
	Program           string            `toml:"program,omitempty"`
	FilterEvents      map[string]Filter `toml:"filterEvents,omitempty"`
	ContainerRuntimes []proc.ContainerRuntime
}

// Filter defines how to filter events for a rule.
type Filter struct {
	Values []string
}
