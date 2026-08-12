package web

import "context"

// SPEC-MCP-CONSOLE-001 M3 (REQ-C-4 / REQ-C-5 / REQ-C-6) — codex authentication
// surface view model and probe injection seam.
//
// The console REPORTS codex state by consuming the existing codex_setup probe;
// it does NOT reimplement the auth classification (AC-C-006 — no second auth
// classifier in internal/web). The probe is injected via the app struct's
// codexStateProbe function field (dependency injection — internal/web cannot
// import internal/cli, so the CLI layer that owns the probe wires it at server
// construction time).

// CodexStateView is the view model for the codex authentication surface
// rendered in the MCP console section. It mirrors the fields the codex_setup
// probe reports (internal/cli.ProbeCodexSetup) plus the auth-provider token
// constants the probe emits.
type CodexStateView struct {
	Installed        bool
	Binary           string
	Version          string
	AuthProvider     string // chatgpt | apiKey | provider | unknown
	EnableReviewGate bool
	AllowWrite       bool
}

// codexAuthProvider constants mirror the internal/cli classification tokens.
// They are NOT a second classification — they are the display-layer knowledge
// of which token the probe already produced. The probe itself (the auth
// classifier in internal/cli) remains the sole classifier (AC-C-006).
const (
	codexAuthChatGPT  = "chatgpt"
	codexAuthAPIKey   = "apiKey"
	codexAuthProvider = "provider"
	codexAuthUnknown  = "unknown"
)

// defaultCodexStateProbe is the fallback when no probe is wired (bare app in a
// unit test, or a console launched without CLI injection). It returns a zero
// view — installed: false, auth_provider: unknown — matching the probe's own
// fail-open behavior for the codex-absent case (AC-C-008). The page renders the
// not-installed state without erroring.
func defaultCodexStateProbe(_ context.Context) CodexStateView {
	return CodexStateView{AuthProvider: codexAuthUnknown}
}
