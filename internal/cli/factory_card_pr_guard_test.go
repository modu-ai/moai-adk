package cli

import (
	"context"
	"errors"
	"sync/atomic"
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
