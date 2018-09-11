package rules

import (
	"testing"

	"github.com/jessfraz/bpfd/api/grpc"
	"github.com/jessfraz/bpfd/proc"
)

func TestMatch(t *testing.T) {
	testcases := map[string]struct {
		rule     grpc.Rule
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
			rule: grpc.Rule{
				FilterEvents: map[string]*grpc.Filter{
					"key": {
						Values: []string{"thing", "blah"},
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
			rule: grpc.Rule{
				FilterEvents: map[string]*grpc.Filter{
					"key": {
						Values: []string{"thing", "blah", "value"},
					},
				},
			},
			data: map[string]string{
				"key": "value",
			},
			runtime:  proc.RuntimeNotFound,
			expected: true,
		},
		"runtime rules no filter false": {
			rule: grpc.Rule{
				ContainerRuntimes: []string{string(proc.RuntimeDocker)},
			},
			data: map[string]string{
				"key": "value",
			},
			runtime:  proc.RuntimeNotFound,
			expected: false,
		},
		"runtime rules no filter true": {
			rule: grpc.Rule{
				ContainerRuntimes: []string{string(proc.RuntimeDocker)},
			},
			data: map[string]string{
				"key": "value",
			},
			runtime:  proc.RuntimeDocker,
			expected: true,
		},
		"runtime rules with filter false": {
			rule: grpc.Rule{
				FilterEvents: map[string]*grpc.Filter{
					"key": {
						Values: []string{"thing", "blah", "value"},
					},
				},
				ContainerRuntimes: []string{string(proc.RuntimeDocker)},
			},
			data: map[string]string{
				"key": "value",
			},
			runtime:  proc.RuntimeNotFound,
			expected: false,
		},
		"runtime rules with filter true": {
			rule: grpc.Rule{
				FilterEvents: map[string]*grpc.Filter{
					"key": {
						Values: []string{"thing", "blah", "value"},
					},
				},
				ContainerRuntimes: []string{string(proc.RuntimeDocker)},
			},
			data: map[string]string{
				"key": "value",
			},
			runtime:  proc.RuntimeDocker,
			expected: true,
		},
	}

	for name, tc := range testcases {
		match := Match(tc.rule, tc.data, string(tc.runtime))
		if match != tc.expected {
			t.Errorf("[%s]: expected match to be %t, got %t", name, tc.expected, match)
		}
	}
}
