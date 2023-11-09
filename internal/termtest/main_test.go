package main_test

import (
	"testing"
)

func TestSkip(t *testing.T) {
	t.Skip("termtest is not run by go test. Use go run ./internal/termtest")
}
