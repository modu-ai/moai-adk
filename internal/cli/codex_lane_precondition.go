package cli

import (
	"fmt"
	"path/filepath"
)

// codexLaneParentCheckout is the REQ-SD-010 parent-checkout rule for the lane
// launch, worded for the entry the operator ran. factoryAssertParentCheckout
// names the factory next verb, which a codex -l launch never runs (t1628).
func codexLaneParentCheckout(dir string) error {
	if real, err := filepath.EvalSymlinks(dir); err == nil {
		dir = real
	}
	primary, _, err := identifyPrimaryCheckout(dir)
	if err != nil {
		return fmt.Errorf("moai codex -l: cannot identify the parent checkout of %s: %w", dir, err)
	}
	if !sameDirPath(primary, dir) {
		return fmt.Errorf("moai codex -l: refused — this verb runs from the parent checkout %s, not from %s", primary, dir)
	}
	return nil
}
