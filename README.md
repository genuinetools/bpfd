# bpfd

[![Travis CI](https://img.shields.io/travis/jessfraz/bpfd.svg?style=for-the-badge)](https://travis-ci.org/jessfraz/bpfd)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](https://godoc.org/github.com/jessfraz/bpfd)
[![Github All Releases](https://img.shields.io/github/downloads/jessfraz/bpfd/total.svg?style=for-the-badge)](https://github.com/jessfraz/bpfd/releases)

Framework for running BPF programs on Linux as a daemon. Container native.

* [Installation](README.md#installation)
   * [Binaries](README.md#binaries)
   * [Via Go](README.md#via-go)
* [Usage](README.md#usage)

## Installation

#### Binaries

For installation instructions from binaries please visit the [Releases Page](https://github.com/jessfraz/bpfd/releases).

#### Via Go

```console
$ go get github.com/jessfraz/bpfd
```

## Usage

```console
$ bpfd -h
bpfd -  Framework for running BPF programs on Linux as a daemon.

Usage: bpfd <command>

Flags:

  -d  enable debug logging (default: false)

Commands:

  daemon   Start the daemon.
  ls       List programs.
  rm       Remove one or more programs.
  version  Show the version information.
```
