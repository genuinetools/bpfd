package program

import (
	"fmt"
	"strings"

	"github.com/jessfraz/bpfd/proc"
	"github.com/jessfraz/bpfd/types"
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
	// String returns a string representation of this program.
	String() string
	// Load creates the bpf module and starts collecting the data for the program.
	Load() error
	// Unload closes the bpf module and all the probes that all attached to it.
	Unload()
	// WatchEvent defines the function to watch the events for the program.
	WatchEvent(rules []types.Rule) (*Event, error)
	// Start starts the map for the program.
	Start()
}

// Event defines the data struct for holding event data.
type Event struct {
	PID              uint32
	TGID             uint32
	Data             map[string]string
	ContainerRuntime proc.ContainerRuntime
}

// Init initialized the program map.
func init() {
	programs = make(map[string]InitFunc)
}

// Register registers an InitFunc for the program.
func Register(name string, initFunc InitFunc) error {
	if _, exists := programs[name]; exists {
		return fmt.Errorf("Name already registered %s", name)
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

// Match checks the rules search and filter properties against the data from
// the event.
// TODO: combine so we are not iterating over the rules twice.
func Match(rules []types.Rule, data map[string]string, pidRuntime proc.ContainerRuntime) bool {
	hasSearch := false
	hasFilter := false
	correctFilter := false
	foundSearch := false

	for _, rule := range rules {
		for _, runtime := range rule.FilterEvents.ContainerRuntimes {
			hasFilter = true
			if pidRuntime == runtime {
				correctFilter = true
				if foundSearch {
					// return early
					return true
				}
			}
		}

		for key, ogValue := range data {
			s, _ := rule.SearchEvents[key]
			for _, find := range s.Values {
				hasSearch = true
				if strings.Contains(ogValue, find) {
					foundSearch = true
					if correctFilter {
						// return early
						return true
					}
				}
			}
		}

	}

	if !hasSearch && !hasFilter {
		// In the case that we do not have any for searches or filters then we can
		// return true to return all events.
		return true
	}

	if hasSearch && hasFilter && correctFilter && foundSearch {
		// This is the case where everything matched.
		return true
	}

	if hasFilter && !hasSearch && correctFilter {
		return true
	}

	if hasSearch && !hasFilter && foundSearch {
		return true
	}

	return false
}
