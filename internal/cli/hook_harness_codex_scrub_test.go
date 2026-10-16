package cli

// hook_harness_codex_scrub_test.go — card t1776 audit round-1 P2: the
// --harness codex lane-env scrub must bypass the EXPLICIT Codex factory
// launches. The launcher (moai codex -l) stamps backend=gpt; a bypass that
// only recognizes backend=codex scrubbed the launcher's own lane env before
// the SessionStart factory checks could read it — the fail-fast and the
// owned-card rule were both unreachable in a real Codex lane. A CLAUDE
// lane's factory identity (backend=claude, or any claude-shaped partial
// stamp) stays scrubbed — REQ-CFR-022's boundary is unchanged.

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

func TestCodexHarnessScrubsLaneEnv(t *testing.T) {
	for _, tc := range []struct {
		backend string
		want    bool
	}{
		{"", true},            // plain codex launch: scrub
		{"claude", true},      // a Claude lane's identity: scrub (REQ-CFR-022)
		{BackendCodex, false}, // explicit managed codex factory launch: bypass
		// The moai codex -l launcher stamp: bypass (card t1776 audit round-1).
		{factory.BackendGPT, false},
	} {
		if got := codexHarnessScrubsLaneEnv(tc.backend); got != tc.want {
			t.Errorf("backend %q: scrub=%v, want %v", tc.backend, got, tc.want)
		}
	}
}
