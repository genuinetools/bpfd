package action

import (
	"github.com/jessfraz/bpfd/api/grpc"
)

// Action performs an action on an event.
type Action interface {
	Do(event *grpc.Event) error
}
