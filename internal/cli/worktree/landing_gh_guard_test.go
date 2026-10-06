package worktree

import (
	"context"
	"errors"
	"sync/atomic"
)

// landingGHUnexpectedCalls counts gh calls no test asked for. The default gh
// seam of the whole test binary is a failing double: no test in this package
// may reach the real gh CLI or the network (SPEC-GITHUB-FLOW-DEFAULT-001 M2-A).
// A test that exercises the PR layer installs its own double (installGH).
var landingGHUnexpectedCalls atomic.Int64

func init() {
	landingGH = func(context.Context, string, ...string) ([]byte, error) {
		landingGHUnexpectedCalls.Add(1)
		return nil, errors.New("gh must not run in tests without a double")
	}
}
