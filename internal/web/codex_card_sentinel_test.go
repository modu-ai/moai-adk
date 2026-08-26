package web

// codex_card_sentinel_test.go — SPEC-CODEX-LAUNCHER-001 M3 (t197) cross-
// surface sentinel bridge (c) of AC-CL-007: the moai web console's Codex
// card consumes the SAME probe values the launcher readout (a) and the
// codex_setup MCP tool (b) consume. One sentinel injected at the console's
// probe-injection seam must surface in the RENDERED card — a missing
// sentinel would mean the card derives its value from some other path than
// the shared probe, which is exactly the second-classifier hazard the AC
// polices. The console's CodexStateView mirrors the probe's typed output
// (SPEC-MCP-CONSOLE-001), and the CLI layer wires the real probe at server
// construction (internal/cli/web.go) — no classification logic lives here.

import (
	"context"
	"strings"
	"testing"
)

// sentinelCodexCard* mirror the M3 sentinel values (internal/cli test
// constants): distinct from any real value the probe could produce.
const (
	sentinelCodexCardBinary = "/sentinel/path/codex"
	sentinelCodexCardVer    = "SENTINEL-VER-9x9"
	sentinelCodexCardAuth   = "sentinel-provider"
)

// TestCodexCard_SentinelPropagationCrossSurface — surface (c) of the AC-CL-
// 007 cross-bridge: a probe sentinel reaching the card render verbatim.
// codexAuthProviderLabel's default branch renders an unknown token verbatim,
// so a sentinel auth provider MUST appear as itself — any mapping to a known
// spelling would prove the card classified the token itself.
func TestCodexCard_SentinelPropagationCrossSurface(t *testing.T) {
	state := CodexStateView{
		Installed:    true,
		Binary:       sentinelCodexCardBinary,
		Version:      sentinelCodexCardVer,
		AuthProvider: sentinelCodexCardAuth,
	}
	body := renderAppBody(t, codexTestApp(t, state))

	for _, want := range []string{
		sentinelCodexCardBinary,
		sentinelCodexCardVer,
		sentinelCodexCardAuth, // rendered VERBATIM, not mapped to a known spelling
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered codex card missing sentinel %q — the card is not consuming the shared probe values", want)
		}
	}
}

// TestCodexCard_ProbeSeamIsTheSoleSource pins the injection discipline: the
// default (unwired) probe fail-opens to unknown, and a wired sentinel
// replaces it wholesale — the render has no second value source.
func TestCodexCard_ProbeSeamIsTheSoleSource(t *testing.T) {
	// Unwired: the fail-open default reports unknown, not a guess.
	if got := defaultCodexStateProbe(context.Background()); got.Installed || got.AuthProvider != codexAuthUnknown {
		t.Errorf("default probe = %+v, want fail-open unknown", got)
	}
	// Wired sentinel on the same seam: the card shows exactly what it got.
	// Installed=true because the card renders the auth block only for an
	// installed codex — the not-installed branch is a different (correct)
	// shape, not a seam bypass.
	body := renderAppBody(t, codexTestApp(t, CodexStateView{Installed: true, AuthProvider: sentinelCodexCardAuth}))
	if !strings.Contains(body, sentinelCodexCardAuth) {
		t.Errorf("wired sentinel not rendered: card is bypassing the seam")
	}
}
