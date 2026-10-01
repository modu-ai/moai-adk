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

const (
	// launcherDebugFlagShort / launcherDebugFlagLong are the debug token
	// spellings every launcher accepts before the -- separator (REQ-012).
	launcherDebugFlagShort = "-d"
	launcherDebugFlagLong  = "--debug"
)

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
