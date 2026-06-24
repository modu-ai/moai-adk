// SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 — M1 CG mode detector.
// SPEC-V3R6-CG-MODE-HARDENING-001 — M3 detector SSOT realignment (REQ-CGH-006).
//
// Implements REQ-WTL-006 (CG mode definition) and REQ-WTL-009 (drift warning
// when teammateMode=tmux but GLM env is absent). Used by worktree --team
// dispatch (M2+) to choose between P1 (moai glm) and P2 (moai cc) launch.
//
// REQ-CGH-006 reworks IsCGMode into a layered OR so CG mode is detectable from a
// CLEAN leader process env (the design-clean signal), not only from the leader
// PROCESS env. The new path keys off the deterministic llm.yaml team_mode value
// (written by persistTeamMode at `moai cg` time) corroborated by the tmux SESSION
// env. The existing teammateMode=tmux + process-env path is PRESERVED verbatim as
// a sufficient fallback so the sibling SPEC's IsCGMode tests stay green.

package tmux

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// settingsLocalMin is a deliberately minimal view of .claude/settings.local.json.
// Avoids importing internal/cli/* (R4 import-cycle mitigation).
type settingsLocalMin struct {
	TeammateMode string `json:"teammateMode"`
}

// sessionEnvReaderFn reads the tmux SESSION environment (via `tmux show-environment`)
// and returns it as a key→value map. It is a package-level seam (REQ-CGH-006
// testability) so tests can inject a recording fake without a live tmux session.
// The default implementation shells out to tmux; on any error it returns the error
// and IsCGMode degrades gracefully (EC-4).
var sessionEnvReaderFn = defaultSessionEnvReader

// defaultSessionEnvReader runs `tmux show-environment` and parses its KEY=value
// output into a map. Lines prefixed with `-` denote unset vars in tmux and are
// skipped.
func defaultSessionEnvReader() (map[string]string, error) {
	out, err := defaultRun(context.Background(), "tmux", "show-environment")
	if err != nil {
		return nil, fmt.Errorf("tmux show-environment: %w", err)
	}
	env := make(map[string]string)
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "-") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		env[key] = val
	}
	return env, nil
}

// IsCGMode reports whether the current session is in CG (Claude + GLM) mode.
//
// CG mode requires the process to be inside a tmux session (TMUX env var set).
// Within tmux, CG mode is detected via a layered OR:
//
//  1. NEW (REQ-CGH-006, design-clean) path — the llm.yaml `team_mode` field equals
//     "cg" (written deterministically by `moai cg`) AND the tmux SESSION env carries
//     GLM markers (ANTHROPIC_AUTH_TOKEN present, or ANTHROPIC_BASE_URL containing
//     "z.ai"). This path detects CG mode even when the leader PROCESS env is clean,
//     which is the whole point of CG mode (the leader runs Claude on a clean env).
//     When team_mode=="cg" but the session env lacks GLM markers, the REQ-WTL-009
//     drift warning fires (reconciled, relocated to the session-env layer) and
//     (false, nil) is returned.
//
//  2. PRESERVED (existing) path — settings.local.json `teammateMode` equals "tmux"
//     AND the leader PROCESS env carries GLM markers (hasGLMEnv). When teammateMode
//     is "tmux" but the process env lacks GLM markers (the original REQ-WTL-009
//     drift case), the warning fires and (false, nil) is returned.
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

	// NEW additive path: detect CG from a clean leader process env via the
	// deterministic llm.yaml team_mode + the tmux SESSION env.
	if readLLMTeamMode(settingsPath) == "cg" {
		if sessionEnvHasGLM() {
			return true, nil
		}
		// team_mode=cg but the session env lacks GLM markers — drift.
		emitGLMDriftWarning(stderrSink)
		return false, nil
	}

	// PRESERVED existing path (sufficient fallback — keeps sibling tests green).
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
		emitGLMDriftWarning(stderrSink)
		return false, nil
	}

	return true, nil
}

// emitGLMDriftWarning writes the REQ-WTL-009 drift warning to stderrSink (if
// non-nil). The literal "GLM env vars are absent" substring is load-bearing — the
// sibling SPEC's tests grep for it (AC-WTL-005 / AC-CGH-006 Scenario 6b).
func emitGLMDriftWarning(stderrSink io.Writer) {
	if stderrSink == nil {
		return
	}
	// Best-effort warning emit; stderr write errors are not actionable here
	// (caller already has the (false, nil) signal it needs).
	_, _ = fmt.Fprintln(stderrSink,
		"warning: teammateMode=tmux but GLM env vars are absent; "+
			"falling back to Claude (P2)")
}

// llmSectionMin is a minimal view of .moai/config/sections/llm.yaml for reading
// only the team_mode field. Avoids importing internal/config (keeps the detector
// dependency-light, R4 import-cycle mitigation).
type llmSectionMin struct {
	LLM struct {
		TeamMode string `yaml:"team_mode"`
	} `yaml:"llm"`
}

// readLLMTeamMode reads the llm.yaml `team_mode` value for the project that owns
// settingsPath. settingsPath is `<root>/.claude/settings.local.json`, so the
// project root is two directories up and llm.yaml lives at
// `<root>/.moai/config/sections/llm.yaml`. Returns "" on any error (absent file,
// parse failure) so the caller falls back to the preserved process-env path.
func readLLMTeamMode(settingsPath string) string {
	claudeDir := filepath.Dir(settingsPath) // <root>/.claude
	root := filepath.Dir(claudeDir)         // <root>
	llmPath := filepath.Join(root, ".moai", "config", "sections", "llm.yaml")

	data, err := os.ReadFile(llmPath)
	if err != nil {
		return ""
	}
	var cfg llmSectionMin
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return ""
	}
	return cfg.LLM.TeamMode
}

// sessionEnvHasGLM reports whether the tmux SESSION env (read via sessionEnvReaderFn)
// carries GLM markers: a non-empty ANTHROPIC_AUTH_TOKEN, or an ANTHROPIC_BASE_URL
// containing "z.ai". On reader error it returns false (graceful degradation, EC-4).
func sessionEnvHasGLM() bool {
	env, err := sessionEnvReaderFn()
	if err != nil {
		return false
	}
	if env["ANTHROPIC_AUTH_TOKEN"] != "" {
		return true
	}
	if strings.Contains(env["ANTHROPIC_BASE_URL"], "z.ai") {
		return true
	}
	return false
}

// hasGLMEnv returns true when GLM-routed credentials are observable in the
// process environment. Either token presence or a z.ai base URL suffices.
//
// PRESERVED from the original SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 detector: this
// process-env read is the sufficient-fallback disjunct of the layered OR
// (REQ-CGH-006 design Option D). It is INTENTIONALLY retained — removing it would
// break the sibling SPEC's IsCGMode tests that set ANTHROPIC_AUTH_TOKEN in the
// process env (C-7 + AC-CGH-006 Scenario 6b). The NEW design-clean path above is
// additive, not a replacement.
func hasGLMEnv() bool {
	if os.Getenv("ANTHROPIC_AUTH_TOKEN") != "" {
		return true
	}
	if strings.Contains(os.Getenv("ANTHROPIC_BASE_URL"), "z.ai") {
		return true
	}
	return false
}
