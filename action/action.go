package action

import (
	"fmt"

	"github.com/jessfraz/bpfd/api/grpc"
)

var (
	// All registered actions.
	actions map[string]InitFunc
)

// InitFunc initializes the action.
type InitFunc func() (Action, error)

// Action performs an action on an event.
type Action interface {
	// Do runs the action on an event.
	Do(event *grpc.Event) error
	// String returns a string representation of this action.
	String() string
}

// Init initialized the action map.
func init() {
	actions = make(map[string]InitFunc)
}

// Register registers an InitFunc for the action.
func Register(name string, initFunc InitFunc) error {
	if _, exists := actions[name]; exists {
		return fmt.Errorf("action name already registered %s", name)
	}
	actions[name] = initFunc

	return nil
}

// Get initializes and returns the registered action.
func Get(name string) (Action, error) {
	if initFunc, exists := actions[name]; exists {
		return initFunc()
	}

	return nil, fmt.Errorf("action %q does not exist as a supported action", name)
}

// List all the registered actions.
func List() []string {
	keys := []string{}
	for k := range actions {
		keys = append(keys, k)
	}
	return keys
}
