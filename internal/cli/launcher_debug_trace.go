package cli

// launcher_debug_trace.go — SPEC-CODEX-DEBUG-MODE-001: the three-runner
// launcher debug surface. One flag spelling (-d / --debug) under one pre---
// scoping discipline across the cc, glm, and codex launchers (REQ-012), one
// stderr destination and one line-prefix vocabulary (REQ-013), and a
// keys-only rule for anything the trace says about the environment (REQ-006:
// the lane keys carry leader addresses and identity — values never print).
//
// The debug trace is gated by the debug tokens alone; MOAI_LOG_LEVEL governs
// the slog default handler and neither enables nor suppresses this trace
// (REQ-008 — two separate axes).

import (
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

const (
	// launcherDebugFlagShort / launcherDebugFlagLong are the debug token
	// spellings every launcher accepts before the -- separator (REQ-012).
	launcherDebugFlagShort = "-d"
	launcherDebugFlagLong  = "--debug"

	// launcherDebugTracePrefix is the single line prefix every launcher
	// debug line carries, on all three runners, to stderr (REQ-005/REQ-013).
	launcherDebugTracePrefix = "moai-launcher-debug:"

	// codexDebugRustLogValue is the level the launcher injects for the codex
	// child's Rust logging when the operator set none (REQ-010).
	codexDebugRustLogValue = "debug"
)

// The debug step vocabulary — one line per executed pre-exec phase
// (REQ-005). The lane phases (join gate, active-run resolution, lane claim,
// exec handoff) reuse the t1378 collector names; these are the phases the
// collector did not already name. They are recorded only under debug, so
// the debug-off threshold report keeps exactly its t1378 step set (REQ-015).
const (
	launchStepBinaryResolve = "binary resolution"
	launchStepProjectRoot   = "project-root resolution"
	launchStepInitGate      = "init-offer gate"
	launchStepLocalInstr    = "local-instruction load"
	launchStepWorktree      = "worktree materialization"
	launchStepChildEnv      = "child-env assembly"
	// The cc/glm phases (plan §F M3): the pre-launch parse, the kanban
	// settings preparation, and the handoff to the backend launch seam.
	launchStepEntryParse    = "entry parse"
	launchStepSettingsPrep  = "settings prep"
	launchStepLaunchHandoff = "launch handoff"
)

// launcherDebugRequested reports whether the invocation carries a debug
// token in the pre--- head. Tokens at or after -- belong to the child and
// are never inspected — the same scoping discipline as the --help scan and
// stripSpawnFlag (the cc/glm form of REQ-002: the token is observe-only
// there, REQ-004).
// @MX:SPEC: SPEC-CODEX-DEBUG-MODE-001
func launcherDebugRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--" {
			return false
		}
		if arg == launcherDebugFlagShort || arg == launcherDebugFlagLong {
			return true
		}
	}
	return false
}

// hasEnvKey reports whether the KEY=value entry list carries key.
func hasEnvKey(env []string, key string) bool {
	for _, entry := range env {
		if k, _, _ := strings.Cut(entry, "="); k == key {
			return true
		}
	}
	return false
}

// codexApplyDebugEnv appends RUST_LOG=debug to the assembled child
// environment when the inherited environment carries no RUST_LOG (REQ-010)
// — the last-wins append posture of codexChildEnv. An operator-supplied
// value is never modified (REQ-011): the caller invokes this only under
// debug mode, and the absence check is what keeps an operator value
// authoritative.
// @MX:SPEC: SPEC-CODEX-DEBUG-MODE-001
func codexApplyDebugEnv(env []string) []string {
	if hasEnvKey(env, config.EnvRustLog) {
		return env
	}
	return append(env, config.EnvRustLog+"="+codexDebugRustLogValue)
}

// codexDebugEnvDetail renders the child-env assembly trace detail from the
// parent's environment: KEY names with presence only — values are never
// emitted (REQ-006; the lane keys carry leader addresses and identity). The
// rustLogInjected caller knowledge distinguishes "injected" from
// "preserved" for the RUST_LOG clause.
// @MX:SPEC: SPEC-CODEX-DEBUG-MODE-001
func codexDebugEnvDetail(rustLogInjected bool) string {
	var parts []string
	if _, ok := os.LookupEnv(config.EnvRustLog); ok {
		parts = append(parts, config.EnvRustLog+" preserved (operator-set)")
	} else if rustLogInjected {
		parts = append(parts, config.EnvRustLog+"=debug injected (absent upstream)")
	}
	var present, absent []string
	for _, key := range codexLaneLaunchEnvKeys {
		if _, ok := os.LookupEnv(key); ok {
			present = append(present, key)
		} else {
			absent = append(absent, key)
		}
	}
	if len(present) > 0 {
		parts = append(parts, "lane keys upstream present (dropped from child): "+strings.Join(present, ", "))
	}
	if len(absent) > 0 {
		parts = append(parts, "lane keys absent: "+strings.Join(absent, ", "))
	}
	return strings.Join(parts, "; ")
}

// stripCodexDebugFlag removes every debug token from the verb-position head
// and reports whether one was found (REQ-001: the token never reaches the
// codex child — its CLI rejects `-d` outright). Only the head is scanned:
// tokens after -- are codex's own and are never inspected (REQ-002). Exact
// token shapes only, mirroring stripSpawnFlag, so a `-d` sitting in the
// value position of another flag is never consumed.
// @MX:SPEC: SPEC-CODEX-DEBUG-MODE-001
func stripCodexDebugFlag(head []string) ([]string, bool) {
	rest := make([]string, 0, len(head))
	found := false
	for _, token := range head {
		if token == launcherDebugFlagShort || token == launcherDebugFlagLong {
			found = true
			continue
		}
		rest = append(rest, token)
	}
	return rest, found
}
