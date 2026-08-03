package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-AUTONOMY-TIERS-001 M8 — moai web console autonomy-tier toggle (AC-002).
//
// The toggle LOGIC (config.TierToggleOptions) is delivered in M5; this file
// owns the HTTP handler + the console UI fragment that surfaces the 3-tier
// toggle in the browser. It is a pure consumer of the M5 gating core — it does
// NOT re-derive the sandbox-proof / kill-switch gating. The handler reads the
// two gates from their env seams via config.SandboxProofKind + config.IsBypassDisabled,
// calls config.TierToggleOptions, and renders one control per tier with the
// `disabled` attribute set when the gating core reports Enabled=false.

// handleAutonomyTiers renders the 3-tier autonomy toggle fragment (AC-002).
// GET-only (read-only); mutating requests are 405'd. The fragment lists the
// 3 tiers with semi-auto first; fully-autonomous carries the `disabled`
// attribute when no sandbox proof is present OR the kill-switch is engaged.
func (a *app) handleAutonomyTiers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_, proofOK := config.SandboxProofKind()
	options := config.TierToggleOptions(proofOK, config.IsBypassDisabled())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, renderAutonomyToggle(options))
}

// renderAutonomyToggle renders the toggle fragment for the given gated options.
// One control per tier; a control whose Enabled is false carries the `disabled`
// attribute so the browser greys it out and will not submit it. semi-auto is
// always the first option and carries a "checked" marker (the pre-selected
// default, REQ-006 — no fully-autonomous default ships).
func renderAutonomyToggle(options []config.TierToggleOption) string {
	var b strings.Builder
	b.WriteString(`<section class="autonomy-toggle" aria-label="Autonomy tier">`)
	b.WriteString(`<h2>Autonomy tier</h2>`)
	for _, opt := range options {
		disabled := ""
		if !opt.Enabled {
			disabled = " disabled"
		}
		checked := ""
		if opt.Tier == config.AutonomyTierSemiAuto {
			checked = " checked"
		}
		fmt.Fprintf(&b,
			`<label><input type="radio" name="autonomy_tier" value="%s"%s%s> %s</label>`,
			opt.Tier, checked, disabled, opt.Tier)
		b.WriteString("\n")
	}
	b.WriteString(`</section>`)
	return b.String()
}
