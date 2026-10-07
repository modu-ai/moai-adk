package cli

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

// factoryGHUnexpectedCalls counts gh calls no test asked for. The default gh
// seam of the whole test binary is a failing double: no test in this package
// may reach the real gh CLI or the network (SPEC-GITHUB-FLOW-DEFAULT-001 M2-B).
// A test that exercises the delivery edge installs its own double.
var factoryGHUnexpectedCalls atomic.Int64

func init() {
	factoryGH = func(context.Context, string, ...string) ([]byte, error) {
		factoryGHUnexpectedCalls.Add(1)
		return nil, errors.New("gh must not run in tests without a double")
	}
}

// The default seam is a failing double: a test that forgets to install its own
// gets an error, and the call is counted — it never reaches the real gh.
func TestFactoryGHDefaultDoubleIsFailing(t *testing.T) {
	before := factoryGHUnexpectedCalls.Load()
	out, err := factoryGH(context.Background(), t.TempDir(), "pr", "view", "x")
	if err == nil || out != nil {
		t.Fatalf("default gh seam answered (%q, %v); it must fail", out, err)
	}
	if got := factoryGHUnexpectedCalls.Load(); got != before+1 {
		t.Fatalf("unexpected-call counter = %d, want %d", got, before+1)
	}
	// A read with no double installed fails closed instead of reaching gh.
	if _, _, err := factoryReadPR(t.TempDir(), "WT-x"); err == nil {
		t.Fatal("factoryReadPR answered through the default seam")
	}
}
