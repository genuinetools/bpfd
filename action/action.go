package action

import (
	"github.com/jessfraz/bpfd/api/grpc"
)

// Action performs an action on an event.
type Action interface {
	// Do runs the action on an event.
	Do(event *grpc.Event) error
	// String returns a string representation of this program.
	String() string
}
