package types

import (
	"github.com/jessfraz/bpfd/proc"
)

// Rule defines a rule to notify on.
type Rule struct {
	Name         string            `toml:"name,omitempty"`
	Program      string            `toml:"program,omitempty"`
	SearchEvents map[string]Search `toml:"searchEvents,omitempty"`
	FilterEvents Filter            `toml:"filterEvents,omitempty"`
}

// Search defines the values to be searched for.
type Search struct {
	Values []string
}

// Filter defines how to filter events for a rule.
type Filter struct {
	ContainerRuntimes []proc.ContainerRuntime
}
