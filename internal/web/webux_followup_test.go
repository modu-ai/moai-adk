package web

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
)

// Web console UX follow-up batch (user visual review), three items:
//
//	D1 — haiku is admitted as a per-agent override model (policy relaxation).
//	D2 — the "(default)" caption is removed from the agent rows.
//	D3 — agent descriptions are localized for ko/ja/zh; en keeps reading the
//	     agent .md frontmatter as the single source of truth.

// --- D1: haiku as a valid per-agent override model ---

// TestD1HaikuInModelOptionList verifies the per-agent model dropdown offers
// haiku, ordered between the inherit sentinel and sonnet (cheapest concrete
// model first).
func TestD1HaikuInModelOptionList(t *testing.T) {
	got := agentFMModelValues()
	want := []string{"inherit", "haiku", "sonnet", "opus", "fable"}
	if len(got) != len(want) {
		t.Fatalf("agentFMModelValues() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("agentFMModelValues() = %v, want %v", got, want)
		}
	}
}

// TestD1HaikuOptionRendered verifies the haiku <option> reaches the rendered
// agent row (a value the form can actually submit). The assertion is scoped to
// the agent's OWN model select — other selects on the page (the GLM tier
// mapping) also carry a haiku option, so a page-wide substring check would pass
// vacuously.
func TestD1HaikuOptionRendered(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "", "")
	writeLLMProfileYAML(t, root, "medium", nil)
	body := renderAgentFMBody(t, root)

	sel := agentModelSelect(t, body, "manager-spec")
	if !strings.Contains(sel, `<option value="haiku"`) {
		t.Errorf("the manager-spec model select renders no haiku option; select was:\n%s", sel)
	}
}

// agentModelSelect extracts the markup of one agent's model <select> element so
// option assertions cannot be satisfied by an unrelated select elsewhere on the
// page.
func agentModelSelect(t *testing.T, body, agent string) string {
	t.Helper()
	open := strings.Index(body, `name="agentfm.`+agent+`.model"`)
	if open < 0 {
		t.Fatalf("no model select rendered for agent %q", agent)
	}
	close := strings.Index(body[open:], "</select>")
	if close < 0 {
		t.Fatalf("unterminated model select for agent %q", agent)
	}
	return body[open : open+close]
}

// TestD1HaikuOverrideSurvivesValidatedReload is the D1 acceptance bar: a haiku
// per-agent override submitted through the console save path persists to
// llm.yaml AND survives the VALIDATED config load (ConfigManager.Load runs
// validateAgentOverrides — the gate that used to reject haiku). LoadRaw would
// not prove anything here: it skips validation, so a value the validator
// rejects would still round-trip. The validated path is what `moai` itself
// uses (internal/cli/launcher.go), so a dropdown whose value fails Load is a
// broken feature even though the bytes landed on disk.
func TestD1HaikuOverrideSurvivesValidatedReload(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "", "")
	writeLLMProfileYAML(t, root, "medium", nil)
	// The validated Load checks the WHOLE config, so the fixture must satisfy the
	// unrelated required fields too — otherwise a failure could not be attributed
	// to the override under test.
	seedMinimalValidConfig(t, root)

	a := newPolicyTestApp(root)
	form := baseSaveForm()
	form.Set("performance_tier", "medium")
	form.Set("agentfm.manager-spec.model", "haiku")
	form.Set("agentfm.manager-spec.effort", "low")

	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /save (haiku override) = %d, want 200; body: %.400s", rec.Code, rec.Body.String())
	}

	// (1) The bytes landed.
	raw, err := config.NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	ov, ok := raw.LLM.AgentOverrides["manager-spec"]
	if !ok {
		t.Fatalf("agent_overrides[manager-spec] not written; overrides=%#v", raw.LLM.AgentOverrides)
	}
	if ov.Model != "haiku" || ov.Effort != "low" {
		t.Fatalf("override = %+v, want {haiku low}", ov)
	}

	// (2) The VALIDATED load accepts it — the actual acceptance bar.
	loaded, err := config.NewConfigManager().Load(root)
	if err != nil {
		t.Fatalf("validated Load rejected the haiku override (dropdown would save a config `moai` cannot read): %v", err)
	}
	if got := loaded.LLM.AgentOverrides["manager-spec"]; got.Model != "haiku" {
		t.Errorf("validated Load override model = %q, want haiku", got.Model)
	}
}

// seedMinimalValidConfig writes the non-llm config sections the VALIDATED
// ConfigManager.Load requires, so a Load failure in a test is attributable to
// the llm.agent_overrides entry under test rather than to unrelated required
// fields the temp fixture happens to lack.
func seedMinimalValidConfig(t *testing.T, root string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "user.yaml"), []byte("user:\n    name: tester\n"), 0o644); err != nil {
		t.Fatalf("write user.yaml: %v", err)
	}
}

// TestD1HaikuOverrideRendersSelected closes the UI round-trip: a persisted haiku
// override comes back as the SELECTED option on the next render. Without this,
// a dropdown could accept and store haiku yet display something else.
func TestD1HaikuOverrideRendersSelected(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "", "")
	writeLLMProfileYAML(t, root, "medium", map[string][2]string{"manager-spec": {"haiku", "low"}})

	sel := agentModelSelect(t, renderAgentFMBody(t, root), "manager-spec")
	if !strings.Contains(sel, `<option value="haiku" selected`) {
		t.Errorf("a persisted haiku override is not shown as selected; select was:\n%s", sel)
	}
}

// TestD1ValidatedLoadStillRejectsAnUnknownOverrideModel is the negative control
// for the relaxation: admitting haiku must not turn the override model field
// into a free-text field. A garbage value is still rejected by the same
// validator that now accepts haiku.
func TestD1ValidatedLoadStillRejectsAnUnknownOverrideModel(t *testing.T) {
	root := t.TempDir()
	seedMinimalValidConfig(t, root)
	writeLLMProfileYAML(t, root, "medium", map[string][2]string{"manager-spec": {"gpt-9", "low"}})

	if _, err := config.NewConfigManager().Load(root); err == nil {
		t.Fatal("validated Load accepted an unknown override model (gpt-9) — the closed set was widened too far")
	} else if !strings.Contains(err.Error(), "gpt-9") {
		t.Errorf("validation error does not name the offending model: %v", err)
	}
}

// TestD1HaikuRelaxationIsScopedToOverrides pins the boundary of the relaxation:
// haiku is admitted ONLY as an llm.agent_overrides value. The default profile
// matrix and the model_routing closed set stay haiku-free.
func TestD1HaikuRelaxationIsScopedToOverrides(t *testing.T) {
	// (a) The Go-default profile matrix has zero haiku cells.
	for profileName, groups := range template.DefaultProfileMatrix() {
		for group, me := range groups {
			if me.Model == "haiku" {
				t.Errorf("defaultProfileMatrix[%s][%s] = haiku — the relaxation must not seed haiku into the profile columns", profileName, group)
			}
		}
	}

	// (b) model_routing_profiles still rejects haiku (a separate closed set).
	src, err := os.ReadFile(filepath.Join("..", "config", "model_routing.go"))
	if err != nil {
		t.Fatalf("read model_routing.go: %v", err)
	}
	block := regexp.MustCompile(`(?s)validRoutingModels\s*=\s*map\[string\]bool\{.*?\}`).Find(src)
	if block == nil {
		t.Fatal("could not locate the validRoutingModels map literal")
	}
	if strings.Contains(string(block), "haiku") {
		t.Error("validRoutingModels admits haiku — the relaxation must be scoped to llm.agent_overrides only")
	}
}

// --- D2: the "(default)" caption is removed ---

// TestD2DefaultCaptionRemovedFromRows verifies the per-row "(default)" tag no
// longer renders on either the model or the effort control.
func TestD2DefaultCaptionRemovedFromRows(t *testing.T) {
	root := t.TempDir()
	// A no-override agent is exactly the case that used to show "(default)".
	seedAgentFMFile(t, root, "moai", "manager-develop", "", "")
	writeLLMProfileYAML(t, root, "medium", nil)
	body := renderAgentFMBody(t, root)

	// Sanity: the row itself rendered (otherwise the absence assertions are vacuous).
	if !strings.Contains(body, `agentfm.manager-develop.model`) {
		t.Fatal("the manager-develop agent row did not render — absence assertions would be vacuous")
	}
	for _, banned := range []string{`data-i18n="agentfm.default"`, `agentfm-default-tag`} {
		if strings.Contains(body, banned) {
			t.Errorf(`the "(default)" caption markup %q must be removed from the agent rows`, banned)
		}
	}
}

// TestD2DefaultKeyRemovedFromDictionary verifies the now-unused agentfm.default
// key is gone from all four locales, and the orphaned CSS rule is gone too.
func TestD2DefaultKeyRemovedFromDictionary(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	if n := strings.Count(dict, `"agentfm.default":`); n != 0 {
		t.Errorf("i18n.js still defines agentfm.default in %d locale(s) — the key is unused after D2", n)
	}
	css := readEmbeddedAsset(t, "console.css")
	if strings.Contains(css, ".agentfm-default-tag") {
		t.Error("console.css still carries the orphaned .agentfm-default-tag rule")
	}
}

// --- D3: localized agent descriptions (ko/ja/zh; en = the .md SSOT) ---

// coreAgentDescKeys is the set of agentdesc.* keys the console must translate —
// the profile-matrix members that actually own an agent .md file. Explore is a
// profile-matrix member but is the Anthropic built-in (no .claude/agents/moai/
// file), so it never renders a description row.
func coreAgentDescKeys() []string {
	var out []string
	for _, a := range template.ProfileMatrixAgents() {
		if a == "Explore" {
			continue
		}
		out = append(out, "agentdesc."+a)
	}
	return out
}

// TestD3DescriptionCarriesI18nKeyWithEnglishBaseline verifies the row emits a
// per-agent data-i18n key AND keeps the English .md frontmatter text as the
// server-side baseline (so en renders the SSOT, never a duplicated copy).
func TestD3DescriptionCarriesI18nKeyWithEnglishBaseline(t *testing.T) {
	root := t.TempDir()
	agentsDir := filepath.Join(root, ".claude", "agents", "moai")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fm := "---\nmodel: inherit\ndescription: SPEC creation specialist for plan-phase artifact authoring.\n---\n# manager-spec\n"
	if err := os.WriteFile(filepath.Join(agentsDir, "manager-spec.md"), []byte(fm), 0o644); err != nil {
		t.Fatal(err)
	}
	writeLLMProfileYAML(t, root, "medium", nil)
	body := renderAgentFMBody(t, root)

	if !strings.Contains(body, `data-i18n="agentdesc.manager-spec"`) {
		t.Error(`the description span does not carry data-i18n="agentdesc.manager-spec"`)
	}
	if !strings.Contains(body, "SPEC creation specialist") {
		t.Error("the English .md baseline text is no longer rendered server-side")
	}
	// The i18n binding must sit on the description span itself.
	spanRe := regexp.MustCompile(`<span class="agentfm-desc"[^>]*data-i18n="agentdesc\.manager-spec"[^>]*>`)
	if !spanRe.MatchString(body) {
		t.Error("data-i18n is not bound to the .agentfm-desc span")
	}
}

// TestD3AgentDescKeysInThreeLocales verifies every core agent carries an
// agentdesc.* entry in ko, ja and zh.
func TestD3AgentDescKeysInThreeLocales(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	blocks := localeBlocks(t, dict)
	for _, key := range coreAgentDescKeys() {
		for _, loc := range []string{"ko", "ja", "zh"} {
			if !strings.Contains(blocks[loc], `"`+key+`":`) {
				t.Errorf("i18n.js locale %q is missing %q", loc, key)
			}
		}
	}
}

// TestD3AgentDescIsEnExempt documents and enforces the deliberate en exemption:
// the agentdesc.* prefix carries NO en entries, because English reads the agent
// .md frontmatter description (the SSOT) as the server-rendered baseline and
// applyI18n leaves a node untouched when the key is absent. Adding en copies
// would duplicate the .md text into a second surface that silently goes stale
// whenever an agent's description is edited.
func TestD3AgentDescIsEnExempt(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	blocks := localeBlocks(t, dict)
	if strings.Contains(blocks["en"], `"agentdesc.`) {
		t.Error("the en locale defines agentdesc.* keys — en must read the agent .md description (SSOT), not a duplicated copy")
	}
}

// TestD3MissingKeyKeepsBaseline verifies the fallback the en exemption relies
// on: applyI18n only overwrites a node when the looked-up value is a non-empty
// string, so an absent key leaves the server-rendered English baseline intact
// (it is never blanked).
func TestD3MissingKeyKeepsBaseline(t *testing.T) {
	js := readEmbeddedAsset(t, "app.js")
	start := strings.Index(js, "function applyI18n")
	if start < 0 {
		t.Fatal("app.js does not define applyI18n")
	}
	end := strings.Index(js[start:], "\n  function ")
	if end < 0 {
		end = len(js) - start
	}
	fn := js[start : start+end]
	if !strings.Contains(fn, `typeof str === "string" && str.length > 0`) {
		t.Error("applyI18n no longer guards the assignment on a non-empty string — a missing agentdesc.* key would blank the English baseline")
	}
}

// localeBlocks splits the i18n dictionary into its four top-level locale blocks
// so per-locale key assertions do not leak across locales.
func localeBlocks(t *testing.T, dict string) map[string]string {
	t.Helper()
	order := []string{"en", "ko", "ja", "zh"}
	idx := make(map[string]int, len(order))
	for _, loc := range order {
		i := strings.Index(dict, "\n  "+loc+": {")
		if i < 0 {
			t.Fatalf("i18n.js has no %q locale block", loc)
		}
		idx[loc] = i
	}
	out := make(map[string]string, len(order))
	for n, loc := range order {
		end := len(dict)
		if n+1 < len(order) {
			end = idx[order[n+1]]
		}
		out[loc] = dict[idx[loc]:end]
	}
	return out
}
