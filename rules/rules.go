package rules

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/jessfraz/bpfd/api/grpc"
)

// Parse parses the rules files and returns an array of rules for each program.
func Parse(files ...string) (map[string][]grpc.Rule, []string, error) {
	rules := map[string][]grpc.Rule{}
	names := []string{}

	for _, file := range files {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, nil, fmt.Errorf("reading file %s failed: %v", file, err)
		}

		var rule grpc.Rule
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
