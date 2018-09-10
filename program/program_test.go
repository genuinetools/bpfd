package program

import (
	"testing"

	"github.com/jessfraz/bpfd/proc"
	"github.com/jessfraz/bpfd/types"
)

func TestMatch(t *testing.T) {
	testcases := map[string]struct {
		rules    []types.Rule
		data     map[string]string
		runtime  proc.ContainerRuntime
		expected bool
	}{
		"no rules": {
			data: map[string]string{
				"key": "value",
			},
			runtime:  proc.RuntimeNotFound,
			expected: true,
		},
		"no runtime rules false": {
			rules: []types.Rule{
				{
					FilterEvents: map[string]types.Filter{
						"key": types.Filter{
							Values: []string{"thing", "blah"},
						},
					},
				},
			},
			data: map[string]string{
				"key": "value",
			},
			runtime:  proc.RuntimeNotFound,
			expected: false,
		},
		"no runtime rules true": {
			rules: []types.Rule{
				{
					FilterEvents: map[string]types.Filter{
						"key": types.Filter{
							Values: []string{"thing", "blah", "value"},
						},
					},
				},
			},
			data: map[string]string{
				"key": "value",
			},
			runtime:  proc.RuntimeNotFound,
			expected: true,
		},
		"runtime rules no search false": {
			rules: []types.Rule{
				{
					ContainerRuntimes: []proc.ContainerRuntime{proc.RuntimeDocker},
				},
			},
			data: map[string]string{
				"key": "value",
			},
			runtime:  proc.RuntimeNotFound,
			expected: false,
		},
		"runtime rules no search true": {
			rules: []types.Rule{
				{
					ContainerRuntimes: []proc.ContainerRuntime{proc.RuntimeDocker},
				},
			},
			data: map[string]string{
				"key": "value",
			},
			runtime:  proc.RuntimeDocker,
			expected: true,
		},
		"runtime rules with search false": {
			rules: []types.Rule{
				{
					FilterEvents: map[string]types.Filter{
						"key": types.Filter{
							Values: []string{"thing", "blah", "value"},
						},
					},
					ContainerRuntimes: []proc.ContainerRuntime{proc.RuntimeDocker},
				},
			},
			data: map[string]string{
				"key": "value",
			},
			runtime:  proc.RuntimeNotFound,
			expected: false,
		},
		"runtime rules with search true": {
			rules: []types.Rule{
				{
					FilterEvents: map[string]types.Filter{
						"key": types.Filter{
							Values: []string{"thing", "blah", "value"},
						},
					},
					ContainerRuntimes: []proc.ContainerRuntime{proc.RuntimeDocker},
				},
			},
			data: map[string]string{
				"key": "value",
			},
			runtime:  proc.RuntimeDocker,
			expected: true,
		},
	}

	for name, tc := range testcases {
		match := Match(tc.rules, tc.data, tc.runtime)
		if match != tc.expected {
			t.Errorf("[%s]: expected match to be %t, got %t", name, tc.expected, match)
		}
	}
}
