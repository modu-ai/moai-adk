package cli

// factory_card_pr.go — the github-flow card delivery edge of
// `moai factory complete` (SPEC-GITHUB-FLOW-DEFAULT-001 M2-B, REQ-GFD-004/005/006).
//
// RED skeleton: the `gh` execution seam only. The delivery edge lands in the
// implementation commit.

import (
	"context"
	"os"
	"os/exec"
	"time"
)

// factoryGHTimeout bounds one `gh` call of the delivery edge (design D-5: 10 s,
// no retry; the delivery edge reuses the bound the landing predicate uses).
var factoryGHTimeout = 10 * time.Second

// factoryGH is the `gh` execution seam of the delivery edge: it runs `gh` in
// dir under ctx and returns stdout. Tests replace it with a double; nothing in
// the test binary may reach the real gh CLI or the network.
var factoryGH = func(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GH_PROMPT_DISABLED=1")
	// A gh that ignores the kill must not hold the call past its bound.
	cmd.WaitDelay = 2 * time.Second
	return cmd.Output()
}
