// SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 — M1 CG mode detector.
//
// Implements REQ-WTL-006 (CG mode definition) and REQ-WTL-009 (drift warning
// when teammateMode=tmux but GLM env is absent). Used by worktree --team
// dispatch (M2+) to choose between P1 (moai glm) and P2 (moai cc) launch.

package tmux

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// settingsLocalMin is a deliberately minimal view of .claude/settings.local.json.
// Avoids importing internal/cli/* (R4 import-cycle mitigation).
type settingsLocalMin struct {
	TeammateMode string `json:"teammateMode"`
}

// IsCGMode reports whether the current session is in CG (Claude + GLM) mode.
//
// CG mode requires ALL of:
//  1. Process is inside a tmux session (TMUX env var set, per InTmuxSession).
//  2. settings.local.json TeammateMode field equals "tmux".
//  3. GLM environment evidence is present: either ANTHROPIC_AUTH_TOKEN is
//     non-empty, or ANTHROPIC_BASE_URL contains "z.ai".
//
// When condition 2 holds but condition 3 fails (the drift case from
// REQ-WTL-009), a warning line is written to stderrSink (if non-nil) and
// (false, nil) is returned so callers fall back to P2 (moai cc).
//
// Returns (false, nil) when settings.local.json does not exist (graceful — a
// user without the file is simply not in CG mode). Returns (false, error) when
// the file exists but cannot be parsed.
//
// stderrSink may be nil; warnings are silently dropped in that case.
func IsCGMode(settingsPath string, stderrSink io.Writer) (bool, error) {
	if !NewDetector().InTmuxSession() {
		return false, nil
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("read settings.local.json: %w", err)
	}

	var s settingsLocalMin
	if err := json.Unmarshal(data, &s); err != nil {
		return false, fmt.Errorf("parse settings.local.json: %w", err)
	}

	if s.TeammateMode != "tmux" {
		return false, nil
	}

	if !hasGLMEnv() {
		if stderrSink != nil {
			fmt.Fprintln(stderrSink,
				"warning: teammateMode=tmux but GLM env vars are absent; "+
					"falling back to Claude (P2)")
		}
		return false, nil
	}

	return true, nil
}

// hasGLMEnv returns true when GLM-routed credentials are observable in the
// process environment. Either token presence or a z.ai base URL suffices.
func hasGLMEnv() bool {
	if os.Getenv("ANTHROPIC_AUTH_TOKEN") != "" {
		return true
	}
	if strings.Contains(os.Getenv("ANTHROPIC_BASE_URL"), "z.ai") {
		return true
	}
	return false
}
