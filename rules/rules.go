package rules

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/jessfraz/bpfd/types"
)

// Parse parses the rules files and returns an array of rules for each program.
func Parse(files ...string) (map[string][]types.Rule, []string, error) {
	rules := map[string][]types.Rule{}
	names := []string{}

	for _, file := range files {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, nil, fmt.Errorf("reading file %s failed: %v", file, err)
		}

		var rule types.Rule
		if _, err := toml.Decode(string(b), &rule); err != nil {
			return nil, nil, fmt.Errorf("decoding file %s as rule failed: %v", file, err)
		}

		if len(rule.Name) < 1 {
			rule.Name = strings.TrimSuffix(filepath.Base(file), ".toml")
		}
		names = append(names, rule.Name)

		rules[rule.Program] = append(rules[rule.Program], rule)
	}

	return rules, names, nil
}
