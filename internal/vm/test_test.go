package vm

import (
	"io"
	"os"
	"testing"

	"github.com/smoynes/elsie/internal/log"
)

func NewTestHarness(t *testing.T) *testHarness {
	t.Parallel()
	th := &testHarness{
		T:      t,
		logger: nil,
	}
	th.logger = makeTestLogger(t, th)

	return th
}

type testHarness struct {
	*testing.T
	logger *log.Logger
}

func (t *testHarness) Make() *LC3 {
	opts := []OptionFn{
		WithLogger(t.logger),
		WithSystemPrivileges(),
	}
	vm := New(opts...)

	return vm
}

func makeTestLogger(t *testing.T, out io.Writer) *log.Logger {
	return log.NewFormattedLogger(os.Stdout)
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
