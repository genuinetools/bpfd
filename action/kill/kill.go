package kill

import (
	"fmt"
	"os"

	"github.com/genuinetools/bpfd/action"
	"github.com/genuinetools/bpfd/api/grpc"
)

const (
	name = "kill"
)

type killAction struct{}

func init() {
	action.Register(name, Init)
}

// Init returns a new kill action.
func Init() (action.Action, error) {
	return &killAction{}, nil
}

func (s *killAction) String() string {
	return name
}

func (s *killAction) Do(event *grpc.Event) error {
	process, err := os.FindProcess(int(event.PID))
	if err != nil {
		return fmt.Errorf("finding process with pid %d failed: %v", event.PID, err)
	}

	if err := process.Kill(); err != nil {
		return fmt.Errorf("killing process with pid %d failed: %v", event.PID, err)
	}

	return nil
}
