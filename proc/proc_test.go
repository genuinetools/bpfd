package proc

import (
	"testing"
)

func TestGetContainerIDAndRuntime(t *testing.T) {
	testcases := map[string]struct {
		name            string
		expectedRuntime ContainerRuntime
		expectedID      string
		input           string
	}{
		"empty": {
			expectedRuntime: RuntimeNotFound,
		},
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
			expectedRuntime: RuntimeNotFound,
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
			expectedID:      "74998d19bd3c7423744214c344c6e814d19b7908a92f165f9d58243073a27a47",
			input: `12:perf_event:/kubepods/burstable/pod98051bd2-a5fa-11e8-9bb9-0a58ac1f31f2/74998d19bd3c7423744214c344c6e814d19b7908a92f165f9d58243073a27a47
11:freezer:/kubepods/burstable/pod98051bd2-a5fa-11e8-9bb9-0a58ac1f31f2/74998d19bd3c7423744214c344c6e814d19b7908a92f165f9d58243073a27a47
10:pids:/kubepods/burstable/pod98051bd2-a5fa-11e8-9bb9-0a58ac1f31f2/74998d19bd3c7423744214c344c6e814d19b7908a92f165f9d58243073a27a47
9:net_cls,net_prio:/kubepods/burstable/pod98051bd2-a5fa-11e8-9bb9-0a58ac1f31f2/74998d19bd3c7423744214c344c6e814d19b7908a92f165f9d58243073a27a47
8:memory:/kubepods/burstable/pod98051bd2-a5fa-11e8-9bb9-0a58ac1f31f2/74998d19bd3c7423744214c344c6e814d19b7908a92f165f9d58243073a27a47
7:cpuset:/kubepods/burstable/pod98051bd2-a5fa-11e8-9bb9-0a58ac1f31f2/74998d19bd3c7423744214c344c6e814d19b7908a92f165f9d58243073a27a47
6:devices:/kubepods/burstable/pod98051bd2-a5fa-11e8-9bb9-0a58ac1f31f2/74998d19bd3c7423744214c344c6e814d19b7908a92f165f9d58243073a27a47
5:blkio:/kubepods/burstable/pod98051bd2-a5fa-11e8-9bb9-0a58ac1f31f2/74998d19bd3c7423744214c344c6e814d19b7908a92f165f9d58243073a27a47
4:rdma:/
3:hugetlb:/kubepods/burstable/pod98051bd2-a5fa-11e8-9bb9-0a58ac1f31f2/74998d19bd3c7423744214c344c6e814d19b7908a92f165f9d58243073a27a47
2:cpu,cpuacct:/kubepods/burstable/pod98051bd2-a5fa-11e8-9bb9-0a58ac1f31f2/74998d19bd3c7423744214c344c6e814d19b7908a92f165f9d58243073a27a47
1:name=systemd:/kubepods/burstable/pod98051bd2-a5fa-11e8-9bb9-0a58ac1f31f2/74998d19bd3c7423744214c344c6e814d19b7908a92f165f9d58243073a27a47`,
		},
		"lxc": {
			expectedRuntime: RuntimeLXC, // this is usually in $container for lxc
			expectedID:      "debian2",
			input: `10:cpuset:/lxc/debian2
9:pids:/lxc/debian2
8:devices:/lxc/debian2
7:net_cls,net_prio:/lxc/debian2
6:freezer:/lxc/debian2
5:blkio:/lxc/debian2
4:memory:/lxc/debian2
3:cpu,cpuacct:/lxc/debian2
2:perf_event:/lxc/debian2
1:name=systemd:/lxc/debian2`,
		},
		"nspawn": {
			expectedRuntime: RuntimeNotFound, // since this variable is in $container
			expectedID:      "nspawntest",
			input: `10:cpuset:/
9:pids:/machine.slice/machine-nspawntest.scope
8:devices:/machine.slice/machine-nspawntest.scope
7:net_cls,net_prio:/
6:freezer:/user/root/0
5:blkio:/machine.slice/machine-nspawntest.scope
4:memory:/machine.slice/machine-nspawntest.scope
3:cpu,cpuacct:/machine.slice/machine-nspawntest.scope
2:perf_event:/
1:name=systemd:/machine.slice/machine-nspawntest.scope`,
		},
		"rkt": {
			expectedRuntime: RuntimeRkt,
			expectedID:      "bfb7d57e-80ff-4ef8-b602-9b907b3f3a38",
			input: `10:cpuset:/
9:pids:/machine.slice/machine-rkt\x2dbfb7d57e\x2d80ff\x2d4ef8\x2db602\x2d9b907b3f3a38.scope/system.slice/debian.service
8:devices:/machine.slice/machine-rkt\x2dbfb7d57e\x2d80ff\x2d4ef8\x2db602\x2d9b907b3f3a38.scope/system.slice/debian.service
7:net_cls,net_prio:/
6:freezer:/user/root/0
5:blkio:/machine.slice/machine-rkt\x2dbfb7d57e\x2d80ff\x2d4ef8\x2db602\x2d9b907b3f3a38.scope/system.slice
4:memory:/machine.slice/machine-rkt\x2dbfb7d57e\x2d80ff\x2d4ef8\x2db602\x2d9b907b3f3a38.scope/system.slice
3:cpu,cpuacct:/machine.slice/machine-rkt\x2dbfb7d57e\x2d80ff\x2d4ef8\x2db602\x2d9b907b3f3a38.scope/system.slice
2:perf_event:/
1:name=systemd:/machine.slice/machine-rkt\x2dbfb7d57e\x2d80ff\x2d4ef8\x2db602\x2d9b907b3f3a38.scope/system.slice/debian.service`,
		},
		"rkt host": {
			expectedRuntime: RuntimeRkt,
			expectedID:      "bfb7d57e-80ff-4ef8-b602-9b907b3f3a38",
			input: `10:cpuset:/
9:pids:/machine.slice/machine-rkt\x2dbfb7d57e\x2d80ff\x2d4ef8\x2db602\x2d9b907b3f3a38.scope/system.slice/debian.service
8:devices:/machine.slice/machine-rkt\x2dbfb7d57e\x2d80ff\x2d4ef8\x2db602\x2d9b907b3f3a38.scope/system.slice/debian.service
7:net_cls,net_prio:/
6:freezer:/user/root/0
5:blkio:/machine.slice/machine-rkt\x2dbfb7d57e\x2d80ff\x2d4ef8\x2db602\x2d9b907b3f3a38.scope/system.slice
4:memory:/machine.slice/machine-rkt\x2dbfb7d57e\x2d80ff\x2d4ef8\x2db602\x2d9b907b3f3a38.scope/system.slice
3:cpu,cpuacct:/machine.slice/machine-rkt\x2dbfb7d57e\x2d80ff\x2d4ef8\x2db602\x2d9b907b3f3a38.scope/system.slice
2:perf_event:/
1:name=systemd:/machine.slice/machine-rkt\x2dbfb7d57e\x2d80ff\x2d4ef8\x2db602\x2d9b907b3f3a38.scope/system.slice/debian.service`,
		},
	}

	for key, tc := range testcases {
		runtime := getContainerRuntime(tc.input)
		if runtime != tc.expectedRuntime {
			t.Errorf("[%s]: expected runtime %q, got %q", key, tc.expectedRuntime, runtime)
		}
		id := getContainerID(tc.input)
		if id != tc.expectedID {
			t.Errorf("[%s]: expected id %q, got %q", key, tc.expectedID, id)
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

func TestGetSeccompEnforceMode(t *testing.T) {
	testcases := map[string]struct {
		name         string
		expectedMode SeccompMode
		input        string
	}{
		"empty": {
			expectedMode: SeccompModeStrict, // since it is enabled by prctl
		},
		"none": {
			expectedMode: SeccompModeStrict, // since it is enabled by prctl
			input: `Name:   cat
Threads:        1
SigQ:   0/127546
SigPnd: 0000000000000000
ShdPnd: 0000000000000000
SigBlk: 0000000000000000
SigIgn: 0000000000000000
SigCgt: 0000000000000000
CapInh: 0000000000000000
CapPrm: 0000000000000000
CapEff: 0000000000000000
CapBnd: 0000003fffffffff
CapAmb: 0000000000000000
NoNewPrivs:     0
Speculation_Store_Bypass:       vulnerable
Cpus_allowed:   ff
Cpus_allowed_list:      0-7
Mems_allowed:   00000000,00000001
Mems_allowed_list:      0
voluntary_ctxt_switches:        1
nonvoluntary_ctxt_switches:     1`,
		},
		"zero": {
			expectedMode: SeccompModeDisabled,
			input: `Name:   cat
Threads:        1
SigQ:   0/127546
SigPnd: 0000000000000000
ShdPnd: 0000000000000000
SigBlk: 0000000000000000
SigIgn: 0000000000000000
SigCgt: 0000000000000000
CapInh: 0000000000000000
CapPrm: 0000000000000000
CapEff: 0000000000000000
CapBnd: 0000003fffffffff
CapAmb: 0000000000000000
NoNewPrivs:     0
Seccomp:        0
Speculation_Store_Bypass:       vulnerable
Cpus_allowed:   ff
Cpus_allowed_list:      0-7
Mems_allowed:   00000000,00000001
Mems_allowed_list:      0
voluntary_ctxt_switches:        1
nonvoluntary_ctxt_switches:     1`,
		},
		"one": {
			expectedMode: SeccompModeStrict,
			input: `Name:   cat
Threads:        1
SigQ:   0/127546
SigPnd: 0000000000000000
ShdPnd: 0000000000000000
SigBlk: 0000000000000000
SigIgn: 0000000000000000
SigCgt: 0000000000000000
CapInh: 0000000000000000
CapPrm: 0000000000000000
CapEff: 0000000000000000
CapBnd: 0000003fffffffff
CapAmb: 0000000000000000
NoNewPrivs:     0
Seccomp:        1
Speculation_Store_Bypass:       vulnerable
Cpus_allowed:   ff
Cpus_allowed_list:      0-7
Mems_allowed:   00000000,00000001
Mems_allowed_list:      0
voluntary_ctxt_switches:        1
nonvoluntary_ctxt_switches:     1`,
		},
		"two": {
			expectedMode: SeccompModeFiltering,
			input: `Name:   cat
Threads:        1
SigQ:   0/127546
SigPnd: 0000000000000000
ShdPnd: 0000000000000000
SigBlk: 0000000000000000
SigIgn: 0000000000000000
SigCgt: 0000000000000000
CapInh: 0000000000000000
CapPrm: 0000000000000000
CapEff: 0000000000000000
CapBnd: 0000003fffffffff
CapAmb: 0000000000000000
NoNewPrivs:     0
Seccomp:        2
Speculation_Store_Bypass:       vulnerable
Cpus_allowed:   ff
Cpus_allowed_list:      0-7
Mems_allowed:   00000000,00000001
Mems_allowed_list:      0
voluntary_ctxt_switches:        1
nonvoluntary_ctxt_switches:     1`,
		},
		"invalid": {
			expectedMode: SeccompModeStrict, // since it is enabled by prctl
			input: `Name:   cat
Threads:        1
SigQ:   0/127546
SigPnd: 0000000000000000
ShdPnd: 0000000000000000
SigBlk: 0000000000000000
SigIgn: 0000000000000000
SigCgt: 0000000000000000
CapInh: 0000000000000000
CapPrm: 0000000000000000
CapEff: 0000000000000000
CapBnd: 0000003fffffffff
CapAmb: 0000000000000000
NoNewPrivs:     0
Seccomp:        17
Speculation_Store_Bypass:       vulnerable
Cpus_allowed:   ff
Cpus_allowed_list:      0-7
Mems_allowed:   00000000,00000001
Mems_allowed_list:      0
voluntary_ctxt_switches:        1
nonvoluntary_ctxt_switches:     1`,
		},
	}

	for key, tc := range testcases {
		mode := getSeccompEnforcingMode(tc.input)
		if mode != tc.expectedMode {
			t.Errorf("[%s]: expected mode %q, got %q", key, tc.expectedMode, mode)
		}
	}
}

func TestGetNoNewPrivileges(t *testing.T) {
	testcases := map[string]struct {
		name     string
		expected bool
		input    string
	}{
		"empty": {},
		"none": {
			input: `Name:   cat
Threads:        1
SigQ:   0/127546
SigPnd: 0000000000000000
ShdPnd: 0000000000000000
SigBlk: 0000000000000000
SigIgn: 0000000000000000
SigCgt: 0000000000000000
CapInh: 0000000000000000
CapPrm: 0000000000000000
CapEff: 0000000000000000
CapBnd: 0000003fffffffff
CapAmb: 0000000000000000
Speculation_Store_Bypass:       vulnerable
Cpus_allowed:   ff
Cpus_allowed_list:      0-7
Mems_allowed:   00000000,00000001
Mems_allowed_list:      0
voluntary_ctxt_switches:        1
nonvoluntary_ctxt_switches:     1`,
		},
		"zero": {
			input: `Name:   cat
Threads:        1
SigQ:   0/127546
SigPnd: 0000000000000000
ShdPnd: 0000000000000000
SigBlk: 0000000000000000
SigIgn: 0000000000000000
SigCgt: 0000000000000000
CapInh: 0000000000000000
CapPrm: 0000000000000000
CapEff: 0000000000000000
CapBnd: 0000003fffffffff
CapAmb: 0000000000000000
NoNewPrivs:     0
Seccomp:        0
Speculation_Store_Bypass:       vulnerable
Cpus_allowed:   ff
Cpus_allowed_list:      0-7
Mems_allowed:   00000000,00000001
Mems_allowed_list:      0
voluntary_ctxt_switches:        1
nonvoluntary_ctxt_switches:     1`,
		},
		"one": {
			expected: true,
			input: `Name:   cat
Threads:        1
SigQ:   0/127546
SigPnd: 0000000000000000
ShdPnd: 0000000000000000
SigBlk: 0000000000000000
SigIgn: 0000000000000000
SigCgt: 0000000000000000
CapInh: 0000000000000000
CapPrm: 0000000000000000
CapEff: 0000000000000000
CapBnd: 0000003fffffffff
CapAmb: 0000000000000000
NoNewPrivs:     1
Seccomp:        1
Speculation_Store_Bypass:       vulnerable
Cpus_allowed:   ff
Cpus_allowed_list:      0-7
Mems_allowed:   00000000,00000001
Mems_allowed_list:      0
voluntary_ctxt_switches:        1
nonvoluntary_ctxt_switches:     1`,
		},
		"invalid": {
			input: `Name:   cat
Threads:        1
SigQ:   0/127546
SigPnd: 0000000000000000
ShdPnd: 0000000000000000
SigBlk: 0000000000000000
SigIgn: 0000000000000000
SigCgt: 0000000000000000
CapInh: 0000000000000000
CapPrm: 0000000000000000
CapEff: 0000000000000000
CapBnd: 0000003fffffffff
CapAmb: 0000000000000000
NoNewPrivs:     17
Seccomp:        17
Speculation_Store_Bypass:       vulnerable
Cpus_allowed:   ff
Cpus_allowed_list:      0-7
Mems_allowed:   00000000,00000001
Mems_allowed_list:      0
voluntary_ctxt_switches:        1
nonvoluntary_ctxt_switches:     1`,
		},
	}

	for key, tc := range testcases {
		nnp := getNoNewPrivileges(tc.input)
		if nnp != tc.expected {
			t.Errorf("[%s]: expected mode %t, got %t", key, tc.expected, nnp)
		}
	}
}

func TestGetUIDGID(t *testing.T) {
	testcases := map[string]struct {
		name        string
		expectedUID uint32
		expectedGID uint32
		input       string
	}{
		"empty": {},
		"none": {
			input: `Name:   cat
Threads:        1
SigQ:   0/127546
SigPnd: 0000000000000000
ShdPnd: 0000000000000000
SigBlk: 0000000000000000
SigIgn: 0000000000000000
SigCgt: 0000000000000000
CapInh: 0000000000000000
CapPrm: 0000000000000000
CapEff: 0000000000000000
CapBnd: 0000003fffffffff
CapAmb: 0000000000000000
Speculation_Store_Bypass:       vulnerable
Cpus_allowed:   ff
Cpus_allowed_list:      0-7
Mems_allowed:   00000000,00000001
Mems_allowed_list:      0
voluntary_ctxt_switches:        1
nonvoluntary_ctxt_switches:     1`,
		},
		"one": {
			input: `Name:   cat
Umask:  0022
State:  R (running)
Tgid:   9314
Ngid:   0
Pid:    9314
PPid:   23600
TracerPid:      0
Uid:    1000    1000    1000    1000
Gid:    1000    1000    1000    1000
FDSize: 256
Groups: 24 25 27 29 30 44 46 101 102 106 111 1000 1001
NStgid: 9314
NSpid:  9314
NSpgid: 9314`,
			expectedUID: 1000,
			expectedGID: 1000,
		},
		"zero": {
			input: `Name:   systemd
Umask:  0000
State:  S (sleeping)
Tgid:   1
Ngid:   0
Pid:    1
PPid:   0
TracerPid:      0
Uid:    0       0       0       0
Gid:    0       0       0       0
FDSize: 256
Groups:  `,
		},
		"invalid": {
			input: `Name:   cat
Threads:        1
SigQ:   0/127546
SigPnd: 0000000000000000
ShdPnd: 0000000000000000
SigBlk: 0000000000000000
SigIgn: 0000000000000000
SigCgt: 0000000000000000
CapInh: 0000000000000000
CapPrm: 0000000000000000
CapEff: 0000000000000000
CapBnd: 0000003fffffffff
CapAmb: 0000000000000000
NoNewPrivs:     17
Seccomp:        17
Speculation_Store_Bypass:       vulnerable
Cpus_allowed:   ff
Cpus_allowed_list:      0-7
Mems_allowed:   00000000,00000001
Mems_allowed_list:      0
voluntary_ctxt_switches:        1
nonvoluntary_ctxt_switches:     1`,
		},
	}

	for key, tc := range testcases {
		uid, gid, err := getUIDGID(tc.input)
		if err != nil {
			t.Errorf("[%s]: error %v", key, err)
		}
		if uid != tc.expectedUID {
			t.Errorf("[%s]: expected uid %d, got %d", key, tc.expectedUID, uid)
		}
		if gid != tc.expectedGID {
			t.Errorf("[%s]: expected gid %d, got %d", key, tc.expectedGID, gid)
		}
	}
}
