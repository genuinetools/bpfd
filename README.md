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
   * [Run the daemon](README.md#run-the-daemon)
   * [Create rules dynamically](README.md#create-rules-dynamically)
   * [Remove rules dynamically](README.md#remove-rules-dynamically)
   * [List active rules](README.md#list-active-rules)
   * [Live tracing events](README.md#live-tracing-events)

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

The current compiled in programs are:

- [bashreadline](program/bashreadline): trace commands being entered into
    the bash command line
- [exec](program/exec): trace calls to exec binaries
- [open](program/open): trace calls to open files

These must implement the `Program` interface:

```go
// Program defines the basic capabilities of a program.
type Program interface {
    // Load creates the bpf module and starts collecting the data for the program.
    Load() error
    // Unload closes the bpf module and all the probes that all attached to it.
    Unload()
    // WatchEvent defines the function to watch the events for the program.
    WatchEvent() (*grpc.Event, error)
    // Start starts the map for the program.
    Start()
    // String returns a string representation of this program.
    String() string
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
    PID              uint32            // Process ID.
    TGID             uint32            // Task group ID.
    Data             map[string]string
    ContainerRuntime string            // Filled in after the program is run so you don't need to.
    ContainerID      string            // Filled in after the program is run so you don't need to.
    Program          string            // Filled in after the program is run so you don't need to.
}
```

### Rules

These are toml files that hold some logic for what you would like to trace. 
You can search for anything returned by a `Program` in its `map[string]string`
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

actions = ["stdout"]

[filterEvents]
  [filterEvents.command]
  values = ["sshd", "dbus-daemon-lau", "ping", "ping6", "critical-stack-", "pmmcli", "filemng", "PassengerAgent", "bwrap", "osdetect", "nginxmng", "sw-engine-fpm", "start-stop-daem"]

containerRuntimes = ["docker","kube"]
```

If you are wondering where the `command` key comes from, it's defined in the
`exec` program [here](https://github.com/jessfraz/bpfd/blob/master/program/exec/exec.go#L200).

### Actions

Actions do "something" on an event. This way you can send filtered events to 
Slack, email, or even run arbitrary code. You could
kill a container, pause a container, or checkpoint a container to restore it
elsewhere without even having to login to a computer.

The current compiled in actions are:

- [stdout](action/stdout): print to stdout
- [kill](action/kill): kill the process

Actions implement the `Actions` interface:

```go
// Action performs an action on an event.
type Action interface {
    // Do runs the action on an event.
    Do(event *grpc.Event) error
    // String returns a string representation of this program.
    String() string
}
```

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

  -d, --debug  enable debug logging (default: false)
  --grpc-addr  Address for gRPC api communication (default: /run/bpfd/bpfd.sock)

Commands:

  create   Create one or more rules.
  daemon   Start the daemon.
  ls       List rules.
  rm       Remove one or more rules.
  trace    Live trace the events returned after filtering.
  version  Show the version information.
```

### Run the daemon

You can preload rules by passing `--rules-dir` to the command or placing
rules in the default directory: `/etc/bpfd/rules`.

```console
$ bpfd daemon -h
Usage: bpfd daemon [OPTIONS]

Start the daemon.

Flags:

  -d, --debug  enable debug logging (default: false)
  --grpc-addr  Address for gRPC api communication (default: /run/bpfd/bpfd.sock)
  --rules-dir  Directory that stores the rules files (default: /etc/bpfd/rules)
```

### Create rules dynamically

You can create rules on the fly with the `create` command. You can pass more
than one file at a time.

```console
Usage: bpfd create [OPTIONS] RULE_FILE [RULE_FILE...]

Create one or more rules.

Flags:

  -d, --debug  enable debug logging (default: false)
  --grpc-addr  Address for gRPC api communication (default: /run/bpfd/bpfd.sock)
```

### Remove rules dynamically

You can delete rules with the `rm` command. You can pass more than one
rule name at a time.

```console
$ bpfd rm -h
Usage: bpfd rm [OPTIONS] RULE_NAME [RULE_NAME...]

Remove one or more rules.

Flags:

  -d, --debug  enable debug logging (default: false)
  --grpc-addr  Address for gRPC api communication (default: /run/bpfd/bpfd.sock)
```

### List active rules

You can list the rules that the daemon is filtering with by using the `ls`
command.

```console
$ bpfd ls
NAME                PROGRAM
bashreadline        bashreadline
password_files      open
setuid_binaries     exec
```

### Live tracing events

You can live trace the events returned after filtering with the `trace`
command.

This does not include past events. Consider it like a tail.

```console
$ bpfd trace
INFO[0000] map[string]string{"filename":"/etc/shadow", "command":"sudo", "returnval":"4"}  container_id= container_runtime=not-found pid=12893 program=open tgid=0
INFO[0000] map[string]string{"command":"sudo", "returnval":"4", "filename":"/etc/sudoers.d/README"}  container_id= container_runtime=not-found pid=12893 program=open tgid=0
INFO[0000] map[string]string{"command":"sudo", "returnval":"4", "filename":"/etc/sudoers.d"}  container_id= container_runtime=not-found pid=12893 program=open tgid=0
INFO[0000] map[string]string{"filename":"/etc/sudoers", "command":"sudo", "returnval":"3"}  container_id= container_runtime=not-found pid=12893 program=open tgid=0
INFO[0000] map[string]string{"command":"sudo bpfd trace"}  container_id= container_runtime=not-found pid=23751 program=bashreadline tgid=0
INFO[0000] map[string]string{"command":"vim README.md"}  container_id= container_runtime=not-found pid=23751 program=bashreadline tgid=0
INFO[0000] map[string]string{"filename":"/etc/shadow", "command":"sudo", "returnval":"4"}  container_id= container_runtime=not-found pid=12786 program=open tgid=0
INFO[0000] map[string]string{"command":"sudo", "returnval":"4", "filename":"/etc/sudoers.d/README"}  container_id= container_runtime=not-found pid=12786 program=open tgid=0
INFO[0000] map[string]string{"filename":"/etc/sudoers.d", "command":"sudo", "returnval":"4"}  container_id= container_runtime=not-found pid=12786 program=open tgid=0
INFO[0000] map[string]string{"filename":"/etc/sudoers", "command":"sudo", "returnval":"3"}  container_id= container_runtime=not-found pid=12786 program=open tgid=0
```
