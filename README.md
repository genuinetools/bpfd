# bpfd

[![Travis CI](https://img.shields.io/travis/jessfraz/bpfd.svg?style=for-the-badge)](https://travis-ci.org/jessfraz/bpfd)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](https://godoc.org/github.com/jessfraz/bpfd)
[![Github All Releases](https://img.shields.io/github/downloads/jessfraz/bpfd/total.svg?style=for-the-badge)](https://github.com/jessfraz/bpfd/releases)

Framework for running BPF programs with rules on Linux as a daemon. Container native.

### NOTE: WIP If you want to contribute see "How it Works" below and consider adding more example rules or programs. Thanks!!

##### How it Works

Currently the programs are in the [programs/ folder](programs). The idea is
that you can add any tracers you would like and then create [rules](examples) for
them.

**Programs**

The sames that exist today are based off a few bcc-tools programs. Writing
these requires knowledge of BPF but you can use the tools provided here to
create your own programs and add them in a fork, if you so wish for say an
enterprise who doesn't want others to know what they are tracking.

**Rules**

These are toml files that hold some logic for what you would like to trace. You
can search for anything returned by a `Program` in it's `map[string]string`
data struct.

You can also filter based off the container runtime you would like to alert on.

**Notifications**

COMING SOON

There will also be an interface for notifications. That way you can send alerts 
on the rules you set up to Slack, email, or even run arbitrary code so you can
kill a container, pause a container, or checkpoint a container to restore it
elsewhere without even having to login to a computer.

* [Installation](README.md#installation)
   * [Binaries](README.md#binaries)
   * [Via Go](README.md#via-go)
* [Usage](README.md#usage)

## Installation

To build, you need to have `libbcc` installed [SEE INSTRUCTIONS HERE](https://github.com/iovisor/bcc/blob/master/INSTALL.md)


#### Binaries

For installation instructions from binaries please visit the [Releases Page](https://github.com/jessfraz/bpfd/releases).

#### Via Go

```console
$ go get github.com/jessfraz/bpfd
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
