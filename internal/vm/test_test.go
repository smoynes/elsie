package vm

import (
	"io"
	"log"
	"strings"
	"testing"
)

func NewTestHarness(t *testing.T) *testHarness {
	t.Parallel()
	th := &testHarness{
		T:   t,
		log: nil,
	}
	th.log = makeTestLogger(t, th)

	return th
}

type testHarness struct {
	*testing.T
	log logger
}

func (t *testHarness) Make() *LC3 {
	opts := []OptionFn{
		WithLogger(t.log),
		WithSystemPrivileges(),
	}
	vm := New(opts...)

	return vm
}

func makeTestLogger(t *testing.T, out io.Writer) logger {
	flag := log.Lshortfile | log.Lmsgprefix
	s := strings.Split(t.Name(), "/")
	logPrefix := s[len(s)-1] + ": "

	return log.New(out, logPrefix, flag)
}

func (t *testHarness) Write(b []byte) (n int, err error) {
	if b[len(b)-1] == '\n' {
		t.Log(string(b[:len(b)-1]))
		return len(b), nil
	} else {
		t.Log(string(b))
		return len(b), nil
	}
}

func (t *testHarness) Log(args ...any) {
	t.T.Helper()
	t.T.Log(args...)
}
