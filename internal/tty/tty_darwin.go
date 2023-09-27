//go:build darwin
// +build darwin

package tty

import "golang.org/x/sys/unix"

const (
	getTermiosIoctl = unix.TIOCGETA
	setTermiosIoctl = unix.TIOCSETA
)
