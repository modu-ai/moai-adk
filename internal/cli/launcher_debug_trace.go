package cli

// launcher_debug_trace.go — SPEC-CODEX-DEBUG-MODE-001: the launcher debug
// surface. One flag spelling (-d / --debug) under one pre--- scoping
// discipline across the cc and glm launchers (REQ-012), one stderr
// destination and one line-prefix vocabulary (REQ-013), and a keys-only rule
// for anything the trace says about the environment (REQ-006: the lane keys
// carry leader addresses and identity — values never print). The codex
// runner's own debug helpers were removed with the general codex launch
// surface (card t1909).
//
// The debug trace is gated by the debug tokens alone; MOAI_LOG_LEVEL governs
// the slog default handler and neither enables nor suppresses this trace
// (REQ-008 — two separate axes).

const (
	// launcherDebugFlagShort / launcherDebugFlagLong are the debug token
	// spellings every launcher accepts before the -- separator (REQ-012).
	launcherDebugFlagShort = "-d"
	launcherDebugFlagLong  = "--debug"

	// launcherDebugTracePrefix is the single line prefix every launcher
	// debug line carries, on all runners, to stderr (REQ-005/REQ-013).
	launcherDebugTracePrefix = "moai-launcher-debug:"
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
	// The cc/glm phases (plan §F M3): the pre-launch parse, the factory
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
// @MX:NOTE: [AUTO] the one shared pre--- debug-token scan for all three launchers — the uniformity contract's parse site (REQ-012); cc/glm/codex must consume this, not private scans
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
