package rules

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/jessfraz/bpfd/api/grpc"
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

	if len(rule.Name) < 1 {
		return grpc.Rule{}, errors.New("rule name cannot be empty")
	}

	if len(rule.Program) < 1 {
		return grpc.Rule{}, errors.New("rule program cannot be empty")
	}

	return rule, nil
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
