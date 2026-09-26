package codexadapter

import "github.com/modu-ai/moai-adk/internal/hook"

// @MX:NOTE: [AUTO] the single place the (Codex, Stop) parse-failure exemption lives — no second event list may encode it
// @MX:SPEC: SPEC-HOOK-STDIN-FAILCLOSED-001
// HostLacksStopBlockCap reports whether host h has no cap on consecutive Stop
// blocks for event ev. It is true only for the Codex host on Stop.
//
// A dispatcher that cannot parse a decision-bearing event's stdin answers it
// fail-closed, except where this predicate is true: a fail-closed Stop on a
// host with no block cap keeps the turn going for as long as the parse failure
// persists, so there the dispatcher answers with the no-opinion object and
// records the exemption instead.
//
// Evidence: .moai/reports/t1152/q2-codex-stop-cap.md (local evidence). With
// codex-cli 0.156.1, `codex exec` accepted a JSON decision:block from a Stop
// hook 191 consecutive times over about ten minutes and never ended the turn
// on its own. Scope of that observation: default configuration, the
// non-interactive exec form, one model — an absence of a cap within that run,
// not a proof that none exists. Claude Code is not covered here: its Stop
// block cap is assumed from repository doctrine and has not been measured.
//
// Revisit this predicate when Codex introduces a Stop block cap: it is the one
// place the exemption is decided.
func HostLacksStopBlockCap(h Harness, ev hook.EventType) bool {
	return h == HarnessCodex && ev == hook.EventStop
}
