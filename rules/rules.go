package rules

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/jessfraz/bpfd/api/grpc"
	"github.com/jessfraz/bpfd/proc"
)

// ParseFiles parses the rules files and returns an array of rules for each program.
func ParseFiles(files ...string) (map[string]map[string]grpc.Rule, []string, error) {
	rules := map[string]map[string]grpc.Rule{}
	names := []string{}

	for _, file := range files {
		rule, err := Parse(file)
		if err != nil {
			return nil, nil, fmt.Errorf("reading file %s failed: %v", file, err)
		}

		names = append(names, rule.Name)

		// Add the rule to our existing rules for the program.
		// TODO: decide to error or not on overwrite
		_, ok := rules[rule.Program]
		if !ok {
			rules[rule.Program] = map[string]grpc.Rule{rule.Name: rule}
			continue
		}

		rules[rule.Program][rule.Name] = rule
	}

	return rules, names, nil
}

// Parse parses a rules file and returns the rule.
func Parse(file string) (grpc.Rule, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return grpc.Rule{}, fmt.Errorf("reading file %s failed: %v", file, err)
	}

	var rule grpc.Rule
	if _, err := toml.Decode(string(b), &rule); err != nil {
		return grpc.Rule{}, fmt.Errorf("decoding file %s as rule failed: %v", file, err)
	}

	if len(rule.Name) < 1 {
		rule.Name = strings.TrimSuffix(filepath.Base(file), ".toml")
	}

	// Validate the rule.
	if err := Validate(rule); err != nil {
		return grpc.Rule{}, err
	}

	return rule, nil
}

// Validate checks that the rule is valid.
func Validate(rule grpc.Rule) error {
	// Check the rule name.
	if len(rule.Name) < 1 {
		return errors.New("rule name cannot be empty")
	}

	// Check the program name.
	if len(rule.Program) < 1 {
		return errors.New("rule program cannot be empty")
	}

	// Check the container runtimes against the valid container runtimes.
	for _, runtime := range rule.ContainerRuntimes {
		if !proc.IsValidContainerRuntime(runtime) {
			return fmt.Errorf("[%s]: %s is not a valid container runtime", rule.Name, runtime)
		}
	}

	return nil
}

// ValidateProgramsAndActions checks that the program and action parts of a rule
// are valid against the given options.
func ValidateProgramsAndActions(rule grpc.Rule, programs, actions []string) error {
	// Check that the program is valid.
	if !in(programs, rule.Program) {
		return fmt.Errorf("[%s]: %s is not a valid program", rule.Name, rule.Program)
	}

	// Check that the actions are valid.
	for _, a := range rule.Actions {
		if !in(actions, a) {
			return fmt.Errorf("[%s]: %s is not a valid action", rule.Name, a)
		}
	}

	return nil
}

// Match checks the filter properties for a rule against the data from
// the event.
func Match(rule grpc.Rule, data map[string]string, pidRuntime string) bool {
	// Return early if we have nothing to filter on.
	if len(rule.ContainerRuntimes) < 1 && len(rule.FilterEvents) < 1 {
		return true
	}

	matchedRuntime := false
	for _, runtime := range rule.ContainerRuntimes {
		if pidRuntime == runtime {
			// Return early if we know we have nothing else to filter on.
			if len(rule.FilterEvents) < 1 {
				return true
			}

			// Continue to the next check.
			matchedRuntime = true
			break
		}
	}

	// Return early here if we never matched a runtime.
	if len(rule.ContainerRuntimes) > 0 && !matchedRuntime {
		return false
	}

	// Return early here if we have nothing else to filter on.
	if len(rule.FilterEvents) < 1 {
		return true
	}

	for key, ogValue := range data {
		s, ok := rule.FilterEvents[key]
		if !ok {
			continue
		}
		for _, find := range s.Values {
			if strings.Contains(ogValue, find) {
				// Return early since we have nothing else to filter on.
				return true
			}
		}
	}

	// We did not match any filters.
	return false
}

func in(a []string, s string) bool {
	for _, b := range a {
		if b == s {
			return true
		}
	}
	return false
}
