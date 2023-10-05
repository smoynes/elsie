//go:build tools
// +build tools

// Package tools declares Go tool dependencies.
package tools

import (
	_ "golang.org/x/tools/cmd/stringer" //nolint
)
