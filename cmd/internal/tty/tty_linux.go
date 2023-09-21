//go:build linux
// +build linux

package tty

import (
	"golang.org/x/sys/unix"
)

const (
	getTermiosIoctl = unix.TCGETS
	setTermiosIoctl = unix.TCSETS
)
