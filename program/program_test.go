package program

import (
	"testing"

	"github.com/jessfraz/bpfd/api/grpc"
	"github.com/jessfraz/bpfd/proc"
)

func TestMatch(t *testing.T) {
	testcases := map[string]struct {
		rules    []grpc.Rule
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
			rules: []grpc.Rule{
				{
					FilterEvents: map[string]*grpc.Filter{
						"key": {
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
			rules: []grpc.Rule{
				{
					FilterEvents: map[string]*grpc.Filter{
						"key": {
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
			rules: []grpc.Rule{
				{
					ContainerRuntimes: []string{string(proc.RuntimeDocker)},
				},
			},
			data: map[string]string{
				"key": "value",
			},
			runtime:  proc.RuntimeNotFound,
			expected: false,
		},
		"runtime rules no search true": {
			rules: []grpc.Rule{
				{
					ContainerRuntimes: []string{string(proc.RuntimeDocker)},
				},
			},
			data: map[string]string{
				"key": "value",
			},
			runtime:  proc.RuntimeDocker,
			expected: true,
		},
		"runtime rules with search false": {
			rules: []grpc.Rule{
				{
					FilterEvents: map[string]*grpc.Filter{
						"key": {
							Values: []string{"thing", "blah", "value"},
						},
					},
					ContainerRuntimes: []string{string(proc.RuntimeDocker)},
				},
			},
			data: map[string]string{
				"key": "value",
			},
			runtime:  proc.RuntimeNotFound,
			expected: false,
		},
		"runtime rules with search true": {
			rules: []grpc.Rule{
				{
					FilterEvents: map[string]*grpc.Filter{
						"key": {
							Values: []string{"thing", "blah", "value"},
						},
					},
					ContainerRuntimes: []string{string(proc.RuntimeDocker)},
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
