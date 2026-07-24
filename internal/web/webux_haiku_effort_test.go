package web

// Web console UX fix (user visual review): when a per-agent model select resolves
// to "haiku", its paired effort select is disabled — reasoning effort is inert for
// Haiku — with a muted inline hint explaining why.
//
// Four assertion axes (goal-to-test, non-SPEC):
//
//	(a) initial render — the effort select carries `disabled` for a haiku-resolved
//	    agent (templ layer, so the no-JS render is correct too), and does NOT for a
//	    non-haiku agent.
//	(b) client interaction — app.js wires a row-scoped model→effort lock, re-applied
//	    after a tier repopulation.
//	(c) the hint key exists in all four locales.
//	(d) the save path is unaffected by the disabled control (a disabled select does
//	    not submit; the effort is backfilled to the resolved value, no corruption).

import (
	"net/http"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// agentEffortSelect extracts the markup of one agent's effort <select> element so
// attribute assertions cannot be satisfied by the model select or an unrelated
// select elsewhere on the page (mirrors agentModelSelect in webux_followup_test.go).
func agentEffortSelect(t *testing.T, body, agent string) string {
	t.Helper()
	open := strings.Index(body, `name="agentfm.`+agent+`.effort"`)
	if open < 0 {
		t.Fatalf("no effort select rendered for agent %q", agent)
	}
	end := strings.Index(body[open:], "</select>")
	if end < 0 {
		t.Fatalf("unterminated effort select for agent %q", agent)
	}
	return body[open : open+end]
}

// agentRowMarkup returns the markup of one agent's whole .agentfm-row container
// (from the row open tag through to the next row / end), so per-row assertions on
// the muted hint are scoped to the correct agent.
func agentRowMarkup(t *testing.T, body, agent string) string {
	t.Helper()
	marker := `name="agentfm.` + agent + `.model"`
	at := strings.Index(body, marker)
	if at < 0 {
		t.Fatalf("no row rendered for agent %q", agent)
	}
	rowOpen := strings.LastIndex(body[:at], `<div class="agentfm-row">`)
	if rowOpen < 0 {
		t.Fatalf("no .agentfm-row container found for agent %q", agent)
	}
	end := len(body)
	if next := strings.Index(body[at:], `<div class="agentfm-row">`); next >= 0 {
		end = at + next
	}
	return body[rowOpen:end]
}

// --- (a) initial render: haiku disables the effort select ---

// TestHaikuEffortSelectDisabledOnRender is the acceptance bar: an agent whose
// profile-matrix-resolved model is haiku renders its effort select with the
// `disabled` attribute on page load (templ layer — correct even without JS).
func TestHaikuEffortSelectDisabledOnRender(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "", "")
	writeLLMProfileYAML(t, root, "medium", map[string][2]string{"manager-spec": {"haiku", "low"}})

	sel := agentEffortSelect(t, renderAgentFMBody(t, root), "manager-spec")
	if !strings.Contains(sel, "disabled") {
		t.Errorf("effort select for a haiku-resolved agent must render disabled; select was:\n%s", sel)
	}
}

// TestNonHaikuEffortSelectNotDisabled is the negative control: a non-haiku agent's
// effort select stays interactive (no disabled attribute).
func TestNonHaikuEffortSelectNotDisabled(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "", "")
	writeLLMProfileYAML(t, root, "medium", map[string][2]string{"manager-spec": {"sonnet", "high"}})

	sel := agentEffortSelect(t, renderAgentFMBody(t, root), "manager-spec")
	if strings.Contains(sel, "disabled") {
		t.Errorf("effort select for a non-haiku (sonnet) agent must NOT be disabled; select was:\n%s", sel)
	}
}

// TestHaikuHintRenderedVisibleForHaiku verifies the muted note renders in the
// haiku row and is NOT hidden (no is-hidden class → shown on load).
func TestHaikuHintRenderedVisibleForHaiku(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "", "")
	writeLLMProfileYAML(t, root, "medium", map[string][2]string{"manager-spec": {"haiku", "low"}})

	row := agentRowMarkup(t, renderAgentFMBody(t, root), "manager-spec")
	if !strings.Contains(row, `data-i18n="hint.effort.haiku_na"`) {
		t.Errorf("haiku row does not render the effort-N/A hint; row was:\n%s", row)
	}
	if !strings.Contains(row, "agentfm-haiku-hint") {
		t.Errorf("haiku row hint missing the .agentfm-haiku-hint class; row was:\n%s", row)
	}
	if strings.Contains(row, "agentfm-haiku-hint is-hidden") {
		t.Errorf("haiku row hint must be VISIBLE (no is-hidden), was hidden; row was:\n%s", row)
	}
}

// TestNonHaikuHintHiddenOnRender verifies the hint still exists in the DOM for a
// non-haiku agent (so JS can toggle it live) but carries the is-hidden class.
func TestNonHaikuHintHiddenOnRender(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "", "")
	writeLLMProfileYAML(t, root, "medium", map[string][2]string{"manager-spec": {"sonnet", "high"}})

	row := agentRowMarkup(t, renderAgentFMBody(t, root), "manager-spec")
	if !strings.Contains(row, `data-i18n="hint.effort.haiku_na"`) {
		t.Errorf("non-haiku row should still render the hint (client-toggled); row was:\n%s", row)
	}
	if !strings.Contains(row, "agentfm-haiku-hint is-hidden") {
		t.Errorf("non-haiku row hint must be hidden (is-hidden class); row was:\n%s", row)
	}
}

// --- (b) client interaction: app.js wiring ---

// TestAppJSHaikuEffortLockWired verifies app.js carries a row-scoped model→effort
// lock, wired on load and re-applied after a tier repopulation.
func TestAppJSHaikuEffortLockWired(t *testing.T) {
	js := readEmbeddedAsset(t, "app.js")
	for _, marker := range []string{
		"function applyHaikuEffortLock", // the per-select lock routine
		`closest(".agentfm-row")`,       // row-scoped pairing (not a global name match)
		`select[name$=".effort"]`,       // finds the paired effort select
		"function wireHaikuEffortLock",  // the initial wiring
		"wireHaikuEffortLock();",        // called from initConsole
	} {
		if !strings.Contains(js, marker) {
			t.Errorf("app.js is missing the haiku-effort-lock marker %q", marker)
		}
	}

	// The lock must re-apply after wireProfileMatrix repopulates cells on a tier
	// change (a repopulation may flip a model cell to/from haiku).
	pmStart := strings.Index(js, "function wireProfileMatrix")
	if pmStart < 0 {
		t.Fatal("app.js no longer defines wireProfileMatrix")
	}
	pmEnd := strings.Index(js[pmStart+1:], "\n  function ")
	if pmEnd < 0 {
		pmEnd = len(js) - pmStart - 1
	}
	if !strings.Contains(js[pmStart:pmStart+1+pmEnd], "reapplyHaikuLocks") {
		t.Error("wireProfileMatrix does not re-apply the haiku lock (reapplyHaikuLocks) after a tier repopulation")
	}
}

// --- (c) i18n: the hint key in all four locales ---

// TestHaikuHintKeyInFourLocales verifies hint.effort.haiku_na exists in en/ko/ja/zh.
func TestHaikuHintKeyInFourLocales(t *testing.T) {
	blocks := localeBlocks(t, readEmbeddedAsset(t, "i18n.js"))
	for _, loc := range []string{"en", "ko", "ja", "zh"} {
		if !strings.Contains(blocks[loc], `"hint.effort.haiku_na":`) {
			t.Errorf("i18n.js locale %q is missing the hint.effort.haiku_na key", loc)
		}
	}
}

// --- (d) save path unaffected by the disabled (non-submitting) control ---

// TestHaikuSaveWithoutEffortFieldSucceeds is the requirement-4 guard: a disabled
// effort select does not submit its value, so /save receives model=haiku with NO
// effort field. The save path must still succeed, backfill the effort to the
// resolved value (never blank/corrupt), and pass the VALIDATED reload.
func TestHaikuSaveWithoutEffortFieldSucceeds(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "", "")
	writeLLMProfileYAML(t, root, "medium", nil)
	seedMinimalValidConfig(t, root)

	a := newPolicyTestApp(root)
	form := baseSaveForm()
	form.Set("performance_tier", "medium")
	form.Set("agentfm.manager-spec.model", "haiku")
	// Deliberately NO agentfm.manager-spec.effort — mimics the disabled select.

	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /save (haiku, effort field omitted) = %d, want 200; body: %.400s", rec.Code, rec.Body.String())
	}

	raw, err := config.NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	ov, ok := raw.LLM.AgentOverrides["manager-spec"]
	if !ok {
		t.Fatalf("a haiku override should be written even without an effort field; overrides=%#v", raw.LLM.AgentOverrides)
	}
	if ov.Model != "haiku" {
		t.Errorf("override model = %q, want haiku", ov.Model)
	}
	if ov.Effort == "" {
		t.Error("override effort is blank — the save path must backfill the resolved effort for an unsubmitted (disabled) effort select, not corrupt/drop it")
	}

	// The VALIDATED load (what `moai` itself uses) must accept the result.
	if _, err := config.NewConfigManager().Load(root); err != nil {
		t.Fatalf("validated Load rejected the haiku-without-effort override: %v", err)
	}
}
