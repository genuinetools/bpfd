# bpfd

[![Travis CI](https://img.shields.io/travis/jessfraz/bpfd.svg?style=for-the-badge)](https://travis-ci.org/jessfraz/bpfd)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](https://godoc.org/github.com/jessfraz/bpfd)
[![Github All Releases](https://img.shields.io/github/downloads/jessfraz/bpfd/total.svg?style=for-the-badge)](https://github.com/jessfraz/bpfd/releases)

Framework for running BPF programs with rules on Linux as a daemon. Container aware.

### NOTE: WIP If you want to contribute see "How it Works" below and consider adding more example rules or programs. Thanks!!

**Table of Contents**

* [How it Works](README.md#how-it-works)
   * [Programs](README.md#programs)
   * [Rules](README.md#rules)
   * [Actions](README.md#actions)
 * [Installation](README.md#installation)
      * [Binaries](README.md#binaries)
      * [Via Go](README.md#via-go)
      * [Via Docker](README.md#via-docker)
 * [Usage](README.md#usage)


## How it Works

[**Programs**](#programs) retrieve the data... 
[**Rules**](#rules) filter the data... 
[**Actions**](#actions) perform actions on the data.

The programs are in the [program/ folder](program). 
The idea is that you can add any tracers you would like 
and then create [rules](examples) for the data retrieved from the programs.
Any events with data that passes the filters will be passed on to the specified
action.

### Programs

The programs that exist today are based off a few
[bcc-tools](https://github.com/iovisor/bcc) programs. 

You could always add your own programs in a fork if you worry people will
reverse engineer the data you are collecting and alerting on.

These must implement the `Program` interface:

```go
// Program defines the basic capabilities of a program.
type Program interface {
	// String returns a string representation of this program.
	String() string
	// Load creates the bpf module and starts collecting the data for the program.
	Load() error
	// Unload closes the bpf module and all the probes that all attached to it.
	Unload()
	// WatchEvent defines the function to watch the events for the program.
	WatchEvent() (*Event, error)
	// Start starts the map for the program.
	Start()
}
```

As you can see from above you could _technically_ implement this interface with
something other than BPF ;)

The `Event` type defines the data returned from the program. As you can see
below, the `Data` is of type `map[string]string` meaning any key value pair can
be returned for the data. The rules then filter using those key value pairs.

```go
// Event defines the data struct for holding event data.
type Event struct {
    PID              uint32
    TGID             uint32
    Data             map[string]string
    ContainerRuntime proc.ContainerRuntime // Filled in after the program is run so you don't need to.
    ContainerID      string                // Filled in after the program is run so you don't need to.
}
```

### Rules

These are toml files that hold some logic for what you would like to trace. 
You can search for anything returned by a `Program` in it's `map[string]string`
data struct.

You can also filter based off the container runtime you would like to alert on.
The container runtime must be one of the strings defined 
[here](https://github.com/jessfraz/bpfd/blob/master/proc/proc.go#L24).

If you provide no rules for a program, then _all_ the events will be passed to
actions.

The example below describes a rule file to filter the data returned from the
`exec` program. Events from `exec` will only be returned if the `command` matches
one of those values AND the container runtime is `docker` or `kube`.

```toml
program = "exec"

[filterEvents]
  [filterEvents.command]
  values = ["sshd", "dbus-daemon-lau", "ping", "ping6", "critical-stack-", "pmmcli", "filemng", "PassengerAgent", "bwrap", "osdetect", "nginxmng", "sw-engine-fpm", "start-stop-daem"]

containerRuntimes = ["docker","kube"]
```

If you are wondering where the `command` key comes from it's defined in the
`exec` program [here](https://github.com/jessfraz/bpfd/blob/master/program/exec/exec.go#L204).

### Actions

COMING SOON

There will also be an interface for actions. That way you can send alerts 
on the rules you set up to Slack, email, or even run arbitrary code so you can
kill a container, pause a container, or checkpoint a container to restore it
elsewhere without even having to login to a computer.

## Installation

To build, you need to have `libbcc` installed [SEE INSTRUCTIONS HERE](https://github.com/iovisor/bcc/blob/master/INSTALL.md)


#### Binaries

For installation instructions from binaries please visit the [Releases Page](https://github.com/jessfraz/bpfd/releases).

#### Via Go

```console
$ go get github.com/jessfraz/bpfd
```

#### Via Docker

```console
$ docker run --rm -it \
    --name bpfd \
    -v /lib/modules:/lib/modules:ro \
    -v /usr/src:/usr/src:ro \
    --privileged \
    r.j3ss.co/bpfd daemon
```

## Usage

```console
$ bpfd -h
bpfd -  Framework for running BPF programs with rules on Linux as a daemon.

Usage: bpfd <command>

Flags:

  -d  enable debug logging (default: false)

Commands:

  create   Create one or more rules.
  daemon   Start the daemon.
  ls       List rules.
  rm       Remove one or more rules.
  version  Show the version information.
```
