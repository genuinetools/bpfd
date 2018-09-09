// Package proc provides tools for inspecting proc.
package container

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/syndtr/gocapability/capability"
	"golang.org/x/sys/unix"
)

const (
	// RuntimeDocker is the string for the docker runtime.
	RuntimeDocker = "docker"
	// RuntimeRkt is the string for the rkt runtime.
	RuntimeRkt = "rkt"
	// RuntimeNspawn is the string for the systemd-nspawn runtime.
	RuntimeNspawn = "systemd-nspawn"
	// RuntimeLXC is the string for the lxc runtime.
	RuntimeLXC = "lxc"
	// RuntimeLXCLibvirt is the string for the lxc-libvirt runtime.
	RuntimeLXCLibvirt = "lxc-libvirt"
	// RuntimeOpenVZ is the string for the openvz runtime.
	RuntimeOpenVZ = "openvz"
	// RuntimeKubernetes is the string for the kubernetes runtime.
	RuntimeKubernetes = "kube"
	// RuntimeGarden is the string for the garden runtime.
	RuntimeGarden = "garden"
	// RuntimePodman is the string for the podman runtime.
	RuntimePodman = "podman"

	uint32Max = 4294967295

	cgroupContainerID = ":(/docker/|/kube.*/.*/|/kube.*/.*/.*/.*/|/system.slice/docker-|/machine.slice/machine-rkt-|/machine.slice/machine-|/lxc/|/lxc-libvirt/|/garden/|/podman/)([[:alnum:]\\-]{1,64})(.scope|$)"
)

var (
	runtimes = []string{RuntimeDocker, RuntimeRkt, RuntimeNspawn, RuntimeLXC, RuntimeLXCLibvirt, RuntimeOpenVZ, RuntimeKubernetes, RuntimeGarden, RuntimePodman}

	cgroupContainerIDRegex = regexp.MustCompile(cgroupContainerID)
)

// GetContainerRuntime returns the container runtime the process is running in.
// If pid is less than one, it returns the runtime for "self".
func GetContainerRuntime(pid int) string {
	file := "/proc/self/cgroup"
	if pid > 0 {
		file = fmt.Sprintf("/proc/%d/cgroup", pid)
	}

	// read the cgroups file
	a := readFile(file)
	runtime := getContainerRuntime(a)
	if len(runtime) > 0 {
		return runtime
	}

	// /proc/vz exists in container and outside of the container, /proc/bc only outside of the container.
	if fileExists("/proc/vz") && !fileExists("/proc/bc") {
		return RuntimeOpenVZ
	}

	a = os.Getenv("container")
	runtime = getContainerRuntime(a)
	if len(runtime) > 0 {
		return runtime
	}

	// PID 1 might have dropped this information into a file in /run.
	// Read from /run/systemd/container since it is better than accessing /proc/1/environ,
	// which needs CAP_SYS_PTRACE
	a = readFile("/run/systemd/container")
	runtime = getContainerRuntime(a)
	if len(runtime) > 0 {
		return runtime
	}

	return ""
}

func getContainerRuntime(input string) string {
	if len(strings.TrimSpace(input)) < 1 {
		return ""
	}

	for _, runtime := range runtimes {
		if strings.Contains(input, runtime) {
			return runtime
		}
	}

	return ""
}

// GetContainerID returns the container ID for a process if it's running in a container.
// If pid is less than one, it returns the container ID for "self".
func GetContainerID(pid int) string {
	file := "/proc/self/cgroup"
	if pid > 0 {
		file = fmt.Sprintf("/proc/%d/cgroup", pid)
	}

	return getContainerID(readFile(file))
}

func getContainerID(input string) string {
	if len(strings.TrimSpace(input)) < 1 {
		return ""
	}

	// rkt encodes the dashes as ascii, replace them.
	input = strings.Replace(input, `\x2d`, "-", -1)

	lines := strings.Split(input, "\n")
	for _, line := range lines {
		matches := cgroupContainerIDRegex.FindStringSubmatch(line)
		if len(matches) > 2 {
			return matches[2]
		}
	}

	return ""
}

// AppArmorProfile determines the apparmor profile for a process.
// If pid is less than one, it returns the apparmor profile for "self".
func AppArmorProfile(pid int) string {
	file := "/proc/self/attr/current"
	if pid > 0 {
		file = fmt.Sprintf("/proc/%d/attr/current", pid)
	}

	f := readFile(file)
	if f == "" {
		return "none"
	}
	return f
}

// UserMapping holds the values for a {uid,gid}_map.
type UserMapping struct {
	ContainerID int64
	HostID      int64
	Range       int64
}

// UserNamespace determines if the process is running in a UserNamespace and returns the mappings if so.
// If pid is less than one, it returns the runtime for "self".
func UserNamespace(pid int) (bool, []UserMapping) {
	file := "/proc/self/uid_map"
	if pid > 0 {
		file = fmt.Sprintf("/proc/%d/uid_map", pid)
	}

	f := readFile(file)
	if len(f) < 0 {
		// user namespace is uninitialized
		return true, nil
	}

	userNs, mappings, err := readUserMappings(f)
	if err != nil {
		return false, nil
	}

	return userNs, mappings
}

func readUserMappings(f string) (iuserNS bool, mappings []UserMapping, err error) {
	parts := strings.Split(f, " ")
	parts = deleteEmpty(parts)
	if len(parts) < 3 {
		return false, nil, nil
	}

	for i := 0; i < len(parts); i += 3 {
		nsu, hu, r := parts[i], parts[i+1], parts[i+2]
		mapping := UserMapping{}

		mapping.ContainerID, err = strconv.ParseInt(nsu, 10, 0)
		if err != nil {
			return false, nil, nil
		}
		mapping.HostID, err = strconv.ParseInt(hu, 10, 0)
		if err != nil {
			return false, nil, nil
		}
		mapping.Range, err = strconv.ParseInt(r, 10, 0)
		if err != nil {
			return false, nil, nil
		}

		if mapping.ContainerID == 0 && mapping.HostID == 0 && mapping.Range == uint32Max {
			return false, nil, nil
		}

		mappings = append(mappings, mapping)
	}

	return true, mappings, nil
}

// Capabilities returns the allowed capabilities for the process.
// If pid is less than one, it returns the runtime for "self".
func Capabilities(pid int) (map[string][]string, error) {
	allCaps := capability.List()

	caps, err := capability.NewPid(pid)
	if err != nil {
		return nil, err
	}

	allowedCaps := map[string][]string{}
	allowedCaps["EFFECTIVE | PERMITTED | INHERITABLE"] = []string{}
	allowedCaps["BOUNDING"] = []string{}
	allowedCaps["AMBIENT"] = []string{}

	for _, cap := range allCaps {
		if caps.Get(capability.CAPS, cap) {
			allowedCaps["EFFECTIVE | PERMITTED | INHERITABLE"] = append(allowedCaps["EFFECTIVE | PERMITTED | INHERITABLE"], cap.String())
		}
		if caps.Get(capability.BOUNDING, cap) {
			allowedCaps["BOUNDING"] = append(allowedCaps["BOUNDING"], cap.String())
		}
		if caps.Get(capability.AMBIENT, cap) {
			allowedCaps["AMBIENT"] = append(allowedCaps["AMBIENT"], cap.String())
		}
	}

	return allowedCaps, nil
}

// SeccompEnforcingMode returns the seccomp enforcing level (disabled, filtering, strict)
// for a process.
// If pid is less than one, it returns the runtime for "self".
// TODO: make this function more efficient and read the file line by line.
func SeccompEnforcingMode(pid int) (string, error) {
	// Read from /proc/self/status Linux 3.8+
	file := "/proc/self/status"
	if pid > 0 {
		file = fmt.Sprintf("/proc/%d/status", pid)
	}

	f := readFile(file)

	// Pre linux 3.8
	if !strings.Contains(f, "Seccomp") {
		// Check if Seccomp is supported, via CONFIG_SECCOMP.
		if err := unix.Prctl(unix.PR_GET_SECCOMP, 0, 0, 0, 0); err != unix.EINVAL {
			// Make sure the kernel has CONFIG_SECCOMP_FILTER.
			if err := unix.Prctl(unix.PR_SET_SECCOMP, unix.SECCOMP_MODE_FILTER, 0, 0, 0); err != unix.EINVAL {
				return "strict", nil
			}
		}
		return "disabled", nil
	}

	// Split status file string by line
	statusMappings := strings.Split(f, "\n")
	statusMappings = deleteEmpty(statusMappings)

	mode := "-1"
	for _, line := range statusMappings {
		if strings.Contains(line, "Seccomp:") {
			mode = string(line[len(line)-1])
		}
	}

	seccompModes := map[string]string{
		"0": "disabled",
		"1": "strict",
		"2": "filtering",
	}

	seccompMode, ok := seccompModes[mode]
	if !ok {
		return "", errors.New("could not retrieve seccomp filtering status")
	}

	return seccompMode, nil
}

func fileExists(file string) bool {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		return true
	}
	return false
}

func readFile(file string) string {
	if !fileExists(file) {
		return ""
	}

	b, err := ioutil.ReadFile(file)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if strings.TrimSpace(str) != "" {
			r = append(r, strings.TrimSpace(str))
		}
	}
	return r
}
