package vm

import (
	"log"
	"testing"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

type testHarness struct{ *testing.T }

func (t testHarness) init() {
	// Using the global logger means we shouldn't parallelize.
	log.SetOutput(&t)
}

func (t *testHarness) Write(b []byte) (n int, err error) {
	t.Log(string(b))
	return len(b), nil
}

func (t *testHarness) Log(args ...any) {
	t.T.Log(args...)
}
