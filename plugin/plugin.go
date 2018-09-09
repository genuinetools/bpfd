package plugin

import (
	"fmt"
)

var (
	// All registered plugins.
	plugins map[string]InitFunc
)

// InitFunc initializes the plugin.
type InitFunc func() (Plugin, error)

// Plugin defines the basic capabilities of a plugin.
type Plugin interface {
	// String returns a string representation of this plugin.
	String() string
	// Load creates the bpf module and starts collecting the data for the plugin.
	Load() error
	// Unload closes the bpf module and all the probes that all attached to it.
	Unload() error
	// WatchEvents starts the go routine to watch the events for the plugin.
	WatchEvents() error
}

// Init initialized the plugin map.
func init() {
	plugins = make(map[string]InitFunc)
}

// Register registers an InitFunc for the plugin.
func Register(name string, initFunc InitFunc) error {
	if _, exists := plugins[name]; exists {
		return fmt.Errorf("Name already registered %s", name)
	}
	plugins[name] = initFunc

	return nil
}

// Get initializes and returns the registered plugin.
func Get(name string) (Plugin, error) {
	if initFunc, exists := plugins[name]; exists {
		return initFunc()
	}

	return nil, fmt.Errorf("plugin %q does not exist as a supported plugin", name)
}

// List all the registered plugins.
func List() []string {
	keys := []string{}
	for k := range plugins {
		keys = append(keys, k)
	}
	return keys
}
