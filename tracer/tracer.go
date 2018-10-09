package tracer

import (
	"context"
	"fmt"

	"github.com/genuinetools/bpfd/api/grpc"
	"github.com/sirupsen/logrus"
)

var (
	// All registered tracers.
	tracers map[string]InitFunc
)

// InitFunc initializes the tracer.
type InitFunc func() (Tracer, error)

// Tracer defines the basic capabilities of a tracer.
type Tracer interface {
	// Load creates the bpf module and starts collecting the data for the tracer.
	Load() error
	// Unload closes the bpf module and all the probes that all attached to it.
	Unload()
	// WatchEvent defines the function to watch the events for the tracer.
	WatchEvent(context.Context) (*grpc.Event, error)
	// Start starts the map for the tracer.
	Start()
	// String returns a string representation of this tracer.
	String() string
}

// Init initialized the tracer map.
func init() {
	tracers = make(map[string]InitFunc)
}

// Register registers an InitFunc for the tracer.
func Register(name string, initFunc InitFunc) error {
	if _, exists := tracers[name]; exists {
		return fmt.Errorf("tracer name already registered %s", name)
	}
	tracers[name] = initFunc

	return nil
}

// Get initializes and returns the registered tracer.
func Get(name string) (Tracer, error) {
	if initFunc, exists := tracers[name]; exists {
		return initFunc()
	}

	return nil, fmt.Errorf("tracer %q does not exist as a supported tracer", name)
}

// List all the registered tracers.
func List() []string {
	keys := []string{}
	for k := range tracers {
		keys = append(keys, k)
	}
	return keys
}

// UnloadAll unloads all the registered tracers.
func UnloadAll() {
	for p := range tracers {
		prog, _ := Get(p)
		prog.Unload()
		logrus.Infof("Successfully unloaded tracer: %s", p)
	}
}
