package rules

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/jessfraz/bpfd/types"
)

// Parse parses the rules files and returns an array of rules for each program.
func Parse(files ...string) (map[string][]types.Rule, error) {
	rules := map[string][]types.Rule{}

	for _, file := range files {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("reading file %s failed: %v", file, err)
		}

		var rule types.Rule
		if _, err := toml.Decode(string(b), &rule); err != nil {
			return nil, fmt.Errorf("decoding file %s as rule failed: %v", file, err)
		}

		rules[rule.Program] = append(rules[rule.Program], rule)
	}

	return rules, nil
}
