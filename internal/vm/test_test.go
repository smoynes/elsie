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
	t.Parallel()
	log.SetOutput(&t)
}

func (t *testHarness) Write(b []byte) (n int, err error) {
	t.Log(string(b))
	return len(b), nil
}
