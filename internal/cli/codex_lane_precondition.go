package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
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
	out, err := runScrubbedGit(cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return cwd
	}
	return strings.TrimSpace(out)
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

// codexLaneJoinError reads a failed factory join as a sentence. The discovery
// sentinel NO_ACTIVE_FACTORY reaches the error renderer, which shows it as
// "No_active_factory.", so the operator sees a code and no remedy (t1628).
func codexLaneJoinError(err error) error {
	if err == nil || !strings.Contains(err.Error(), "NO_ACTIVE_FACTORY") {
		return err
	}
	return errors.New("no live factory leader to join: start the factory leader first (moai cc -f or moai glm -f), then run moai codex -l again")
}

// codexLaneChildEnv gives the codex child the project anchor this launch
// selected. The child inherits CLAUDE_PROJECT_DIR, and its own start prompt runs
// moai todo --auto, whose precondition reads that anchor before the child's
// working directory; a stale card-tree anchor would refuse it (t1628, card-review P2).
func codexLaneChildEnv(env []string, root string) []string {
	out := make([]string, 0, len(env)+1)
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		if key != config.EnvClaudeProjectDir {
			out = append(out, entry)
		}
	}
	return append(out, config.EnvClaudeProjectDir+"="+root)
}
