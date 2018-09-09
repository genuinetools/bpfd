package types

import (
	"github.com/jessfraz/bpfd/proc"
)

// Rule defines a rule to notify on.
type Rule struct {
	Program      string
	SearchEvents Search
	FilterEvents Filter
}

// Search defines the values to be searched for.
type Search struct {
	Values []string
}

// Filter defines how to filter events for a rule.
type Filter struct {
	ContainerRuntimes []proc.ContainerRuntime
}
