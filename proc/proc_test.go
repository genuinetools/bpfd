package container

import (
	"testing"
)

func TestGetContainerIDAndRuntime(t *testing.T) {
	testcases := map[string]struct {
		name            string
		expectedRuntime string
		expectedID      string
		input           string
	}{
		"typical docker": {
			expectedRuntime: RuntimeDocker,
			expectedID:      "68fad1f9e0985989408aff30e7b83e7dada1d235ff46a22c5465ca193ddf0fac",
			input: `11:pids:/docker/68fad1f9e0985989408aff30e7b83e7dada1d235ff46a22c5465ca193ddf0fac
10:devices:/docker/68fad1f9e0985989408aff30e7b83e7dada1d235ff46a22c5465ca193ddf0fac
9:freezer:/docker/68fad1f9e0985989408aff30e7b83e7dada1d235ff46a22c5465ca193ddf0fac
8:net_cls,net_prio:/docker/68fad1f9e0985989408aff30e7b83e7dada1d235ff46a22c5465ca193ddf0fac
7:perf_event:/docker/68fad1f9e0985989408aff30e7b83e7dada1d235ff46a22c5465ca193ddf0fac
6:cpuset:/docker/68fad1f9e0985989408aff30e7b83e7dada1d235ff46a22c5465ca193ddf0fac
5:memory:/docker/68fad1f9e0985989408aff30e7b83e7dada1d235ff46a22c5465ca193ddf0fac
4:blkio:/docker/68fad1f9e0985989408aff30e7b83e7dada1d235ff46a22c5465ca193ddf0fac
3:cpu,cpuacct:/docker/68fad1f9e0985989408aff30e7b83e7dada1d235ff46a22c5465ca193ddf0fac
2:hugetlb:/docker/68fad1f9e0985989408aff30e7b83e7dada1d235ff46a22c5465ca193ddf0fac
1:name=systemd:/docker/68fad1f9e0985989408aff30e7b83e7dada1d235ff46a22c5465ca193ddf0fac
0::/system.slice/containerd.service`,
		},
		"uncontainerized process": {
			expectedRuntime: "",
			expectedID:      "",
			input: `11:pids:/system.slice/ssh.service
10:devices:/system.slice/ssh.service
9:freezer:/
8:net_cls,net_prio:/
7:perf_event:/
6:cpuset:/
5:memory:/system.slice/ssh.service
4:blkio:/system.slice/ssh.service
3:cpu,cpuacct:/system.slice/ssh.service
2:hugetlb:/
1:name=systemd:/system.slice/ssh.service
0::/system.slice/ssh.service`,
		},
		"kubernetes": {
			expectedRuntime: RuntimeKubernetes,
			expectedID:      "",
			input:           ``,
		},
		"lxc": {
			expectedRuntime: RuntimeLXC,
			expectedID:      "",
			input:           ``,
		},
		"nspawn": {
			expectedRuntime: RuntimeNspawn,
			expectedID:      "",
			input:           ``,
		},
		"rkt": {
			expectedRuntime: RuntimeRkt,
			expectedID:      "",
			input:           ``,
		},
		"podman": {
			expectedRuntime: RuntimePodman,
			expectedID:      "",
			input:           ``,
		},
	}

	for key, tc := range testcases {
		runtime := getContainerRuntime(tc.input)
		if runtime != tc.expectedRuntime {
			t.Fatalf("[%s]: expected runtime %q, got %q", key, tc.expectedRuntime, runtime)
		}
		id := getContainerID(tc.input)
		if id != tc.expectedID {
			t.Fatalf("[%s]: expected id %q, got %q", key, tc.expectedID, id)
		}
	}
}

func TestGetUserMappings(t *testing.T) {
	f := `         0     100000       1000
      1000       1000          1
      1001     101001      64535`
	expected := []UserMapping{
		{
			ContainerID: 0,
			HostID:      100000,
			Range:       1000,
		},
		{
			ContainerID: 1000,
			HostID:      1000,
			Range:       1,
		},
		{
			ContainerID: 1001,
			HostID:      101001,
			Range:       64535,
		},
	}

	userNs, mappings, err := readUserMappings(f)
	if err != nil {
		t.Fatal(err)
	}

	if !userNs {
		t.Fatal("expected user namespaces to be true")
	}

	if len(expected) != len(mappings) {
		t.Fatalf("expected length %d got %d", len(expected), len(mappings))
	}
}
