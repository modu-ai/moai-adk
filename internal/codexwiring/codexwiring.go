// Package codexwiring generates and refreshes the Codex-side wiring files of a
// MoAI project: <project>/.codex/hooks.json and <project>/.codex/config.toml.
//
// It is the first production consumer of internal/codexadapter: the hook table
// is DERIVED from the adapter's EventTable (never re-enumerated here), and
// every hooks.json write passes through the adapter's measured-whitelist
// ValidateConfig gate before a single byte reaches disk — Codex silently
// disables an entire hooks file over one stray key (t83 Finding D), so only
// validated bytes may land (SPEC-CODEX-WIRING-001 REQ-CW-003).
//
// Determinism follows the agentemit rules: fixed key order, no timestamps, no
// absolute paths, no environment-derived values — regenerating over unchanged
// input must be byte-identical (REQ-CW-006).
//
// The package never prompts: diagnostics go to a warn writer and refusals are
// returned as errors (C-HRA-008 subagent boundary).
package codexwiring

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Relative artifact paths inside a project root.
const (
	// HooksRelPath is the project-layer Codex hooks file.
	HooksRelPath = ".codex/hooks.json"
	// ConfigRelPath is the project-layer Codex config file.
	ConfigRelPath = ".codex/config.toml"
	// SidecarPath is the trust sidecar recording the sha256 of the last
	// generated wiring content, relative to the project root.
	SidecarPath = ".moai/state/codex-wiring.json"
)

// Handler-level constants (plan D1 data contract).
const (
	// moaiHandlerPrefix identifies MoAI-managed handlers in a merged document:
	// any handler whose command starts with this prefix is refreshed by the
	// generator; everything else is user-owned and preserved.
	moaiHandlerPrefix = "moai hook "

	// harnessCodexSuffix is appended to every emitted handler command so the
	// runtime dispatcher wraps its output through the codex adapter.
	harnessCodexSuffix = " --harness codex"

	// sessionEndTimeoutCeiling is Codex's documented SessionEnd timeout cap
	// (default 1, ceiling 3 — spec §A.5).
	sessionEndTimeoutCeiling = 3

	// defaultHandlerTimeout is the table constant for every non-SessionEnd
	// handler (D1; matches the t83 observed-behavior case).
	defaultHandlerTimeout = 10
)

// Guidance constants (REQ-CW-008). The tokens "codex /hooks" and
// "/hooks to re-trust" are acceptance-criterion anchors (AC-CW-008) — do not
// reword them without updating acceptance.md.
const (
	// FirstTrustGuidance is printed when hooks.json is created for the first
	// time.
	FirstTrustGuidance = "Codex wiring: .codex/hooks.json created. Codex loads project hooks only from a trusted .codex/ layer — approve it when prompted, then run codex /hooks to review and trust the MoAI hooks."

	// ReTrustGuidance is printed when a regeneration CHANGED hooks.json
	// content: Codex records trust against the hook-definition hash, so
	// changed hooks stop running until re-approved (spec §A.5).
	ReTrustGuidance = "Codex wiring: .codex/hooks.json changed. Codex stops changed hooks until they are re-approved — run codex /hooks to re-trust the MoAI hooks."
)

// Result reports what a wiring pass did, so callers (init, update, doctor
// tests) can branch on it without re-reading files.
type Result struct {
	// HooksWritten reports hooks.json bytes were written this run.
	HooksWritten bool
	// HooksChanged reports the rendered content differed from the on-disk
	// file (fresh creations count as changed).
	HooksChanged bool
	// HooksExisted reports hooks.json existed before the pass.
	HooksExisted bool
	// HooksSkipped reports the on-disk hooks.json was unparseable and was
	// left untouched with a diagnostic (§B edge case — warn and continue).
	HooksSkipped bool
	// ConfigWritten reports config.toml bytes were written this run.
	ConfigWritten bool
}

// wiringFilesExist reports whether either wiring artifact exists — the opt-in
// marker REQ-CW-009 gates the update path on.
func wiringFilesExist(projectRoot string) bool {
	for _, rel := range []string{HooksRelPath, ConfigRelPath} {
		if _, err := os.Stat(filepath.Join(projectRoot, rel)); err == nil {
			return true
		}
	}
	return false
}

// warnf writes a diagnostic to the warn writer when one was supplied.
func warnf(warn io.Writer, format string, args ...any) {
	if warn == nil {
		return
	}
	_, _ = fmt.Fprintf(warn, "warning: "+format+"\n", args...)
}
