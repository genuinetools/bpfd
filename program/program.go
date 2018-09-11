package program

import (
	"fmt"
	"strings"

	"github.com/jessfraz/bpfd/api/grpc"
	"github.com/sirupsen/logrus"
)

var (
	// All registered programs.
	programs map[string]InitFunc
)

// InitFunc initializes the program.
type InitFunc func() (Program, error)

// Program defines the basic capabilities of a program.
type Program interface {
	// Load creates the bpf module and starts collecting the data for the program.
	Load() error
	// Unload closes the bpf module and all the probes that all attached to it.
	Unload()
	// WatchEvent defines the function to watch the events for the program.
	WatchEvent() (*grpc.Event, error)
	// Start starts the map for the program.
	Start()
	// String returns a string representation of this program.
	String() string
}

// Init initialized the program map.
func init() {
	programs = make(map[string]InitFunc)
}

// Register registers an InitFunc for the program.
func Register(name string, initFunc InitFunc) error {
	if _, exists := programs[name]; exists {
		return fmt.Errorf("program name already registered %s", name)
	}
	programs[name] = initFunc

	return nil
}

// Get initializes and returns the registered program.
func Get(name string) (Program, error) {
	if initFunc, exists := programs[name]; exists {
		return initFunc()
	}

	return nil, fmt.Errorf("program %q does not exist as a supported program", name)
}

// List all the registered programs.
func List() []string {
	keys := []string{}
	for k := range programs {
		keys = append(keys, k)
	}
	return keys
}

// UnloadAll unloads all the registered programs.
func UnloadAll() {
	for p := range programs {
		prog, _ := Get(p)
		prog.Unload()
		logrus.Infof("Successfully unloaded program: %s", p)
	}
}

// Match checks the filter properties for a rule against the data from
// the event. It returns a boolean and the actions for the rule.
// TODO: make better
func Match(rule grpc.Rule, data map[string]string, pidRuntime string) (bool, []string) {
	hasFilters := false
	hasRuntimeFilter := false
	correctRuntime := false
	passedFilters := false

	for _, runtime := range rule.ContainerRuntimes {
		hasRuntimeFilter = true
		if pidRuntime == runtime {
			correctRuntime = true
			if passedFilters {
				// return early
				return true, rule.Actions
			}
		}
	}

	for key, ogValue := range data {
		s, ok := rule.FilterEvents[key]
		if !ok {
			continue
		}
		for _, find := range s.Values {
			hasFilters = true
			if strings.Contains(ogValue, find) {
				passedFilters = true
				if correctRuntime {
					// return early
					return true, rule.Actions
				}
			}
		}
	}

	if !hasFilters && !hasRuntimeFilter {
		// In the case that we do not have any for searches or filters then we can
		// return true to return all events.
		return true, rule.Actions
	}

	if hasFilters && hasRuntimeFilter && correctRuntime && passedFilters {
		// This is the case where everything matched.
		return true, rule.Actions
	}

	if hasRuntimeFilter && !hasFilters && correctRuntime {
		return true, rule.Actions
	}

	if hasFilters && !hasRuntimeFilter && passedFilters {
		return true, rule.Actions
	}

	return false, nil
}
