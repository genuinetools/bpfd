package types

import (
	"github.com/jessfraz/bpfd/proc"
)

// Rule defines a rule to notify on.
type Rule struct {
	Program      string
	SearchEvents map[string]string
	FilterEvents Filter
}

// Filter defines how to filter events for a rule.
type Filter struct {
	ContainerRuntimes []proc.ContainerRuntime
}
