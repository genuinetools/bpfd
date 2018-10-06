// Package interrupt provides an action for the filtered process to be interrupted
package interrupt

import (
	"fmt"
	"os"

	"github.com/genuinetools/bpfd/action"
	"github.com/genuinetools/bpfd/api/grpc"
)

const (
	name = "interrupt"
)

type interruptAction struct{}

func init() {
	action.Register(name, Init)
}

// Init returns a new interrupt action.
func Init() (action.Action, error) {
	return &interruptAction{}, nil
}

func (s *interruptAction) String() string {
	return name
}

func (s *interruptAction) Do(event *grpc.Event) error {
	process, err := os.FindProcess(int(event.PID))
	if err != nil {
		return fmt.Errorf("finding process with pid %d failed: %v", event.PID, err)
	}

	if err := process.Signal(os.Interrupt); err != nil {
		return fmt.Errorf("interrupting process with pid %d failed: %v", event.PID, err)
	}

	return nil
}
