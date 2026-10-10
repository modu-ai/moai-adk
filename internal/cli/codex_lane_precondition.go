package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// codexLaneLaunchDir is the directory `moai codex -l` was run from, resolved to
// its git toplevel. The lane launch judges the parent-checkout rule and picks
// the factory root from this directory, not from CLAUDE_PROJECT_DIR: that
// variable carries the project anchor of whichever session exported it, so a
// launch from the parent checkout inside a card session would be refused
// against the card tree (t1628; the cc and glm paths stay with t1604).
func codexLaneLaunchDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	out, err := exec.Command("git", "-C", cwd, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return cwd
	}
	return strings.TrimSpace(string(out))
}

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
