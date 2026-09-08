package web

// codex_panel_test.go — SPEC-WEB-CODEX-PANEL-001 guards.
//
// The codex tab is a READ-ONLY MIRROR: every mirrored field stays declared,
// rendered and editable on its owning tab, and the mirror renders no form
// element carrying a name attribute — the hidden bool companion
// (<name>__present) included, because parseSchemaForm reads a submitted
// companion with an empty control as an explicit false (spec.md §B.1).
//
// Every panel-scoped assertion below slices with panelHTML(t, html, "codex"),
// never a whole-body render. panelHTML falls back to end-of-document for the
// LAST panel, so codex being not-last is load-bearing for the scoping and is
// asserted in TestCodexPanel_NoNamedFormElements.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/settings"
)

// codexMirrorOwningPanel maps each mirrored field to the panel that owns its
// editing surface. It is pinned here rather than derived from the production
// helper: a test that asks the implementation where a field lives cannot
// detect the implementation putting it in the wrong place.
var codexMirrorOwningPanel = map[string]string{
	"workflow.audit.gates.codex":         "audit",
	"workflow.audit.codex.model":         "audit",
	"workflow.audit.codex.effort":        "audit",
	"workflow.audit.model":               "audit", // the one declared exception
	"workflow.codex.review_gate.enabled": "mcp",
	"workflow.codex.task.allow_write":    "mcp",
	"mcp.tools.codex_audit.enabled":      "mcp",
	"mcp.tools.codex_setup.enabled":      "mcp",
	"mcp.tools.codex_task.enabled":       "mcp",
	"mcp.tools.codex_job_status.enabled": "mcp",
	"mcp.tools.codex_job_result.enabled": "mcp",
	"mcp.tools.codex_job_cancel.enabled": "mcp",
}

// wantCodexTokenFields is the INDEPENDENT ORACLE for AC-WCP-006: the 11
// registry fields carrying a codex token. workflow.audit.model is deliberately
// absent — it carries no codex token, so no predicate reaches it, and it is
// asserted separately by TestCodexMirrorDeclaredException (spec.md §C.1).
var wantCodexTokenFields = []string{
	"workflow.audit.gates.codex",
	"workflow.audit.codex.model",
	"workflow.audit.codex.effort",
	"workflow.codex.review_gate.enabled",
	"workflow.codex.task.allow_write",
	"mcp.tools.codex_audit.enabled",
	"mcp.tools.codex_setup.enabled",
	"mcp.tools.codex_task.enabled",
	"mcp.tools.codex_job_status.enabled",
	"mcp.tools.codex_job_result.enabled",
	"mcp.tools.codex_job_cancel.enabled",
}

// codexPanelTestApp builds an app over a seeded temp project root, optionally
// overwriting workflow.yaml / mcp.yaml so a test can put a sentinel value on
// disk and prove the panel read it through the shared view model.
func codexPanelTestApp(t *testing.T, workflowYAML, mcpYAML string) *app {
	t.Helper()
	root := t.TempDir()
	seedWebSections(t, root, allSectionFixtures...)
	dir := filepath.Join(root, ".moai", "config", "sections")
	if workflowYAML != "" {
		if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(workflowYAML), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if mcpYAML != "" {
		if err := os.WriteFile(filepath.Join(dir, "mcp.yaml"), []byte(mcpYAML), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	a.readPreferences = func(string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{}, nil
	}
	a.writePreferences = func(string, profile.ProfilePreferences) error { return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	a.recordLastProfile = func(string) error { return nil }
	return a
}

// ─── AC-WCP-003 / AC-WCP-004 / AC-WCP-002 (not-last) ─────────────────────

// TestCodexPanel_NoNamedFormElements is the REQ-WCP-002 regression guard:
// the codex panel emits no named form element and no form control at all.
func TestCodexPanel_NoNamedFormElements(t *testing.T) {
	// AC-WCP-002: codex must not be the last tab, or panelHTML's
	// end-of-document fallback silently widens every panel-scoped assertion
	// below into a whole-page one — vacuous, and with no failure.
	tabs := consoleTabs()
	if len(tabs) == 0 || tabs[len(tabs)-1].ID == "codex" {
		t.Fatalf("codex must not be the last tab (panelHTML would slice to end of document)")
	}

	html := renderConsolePage(t)
	body := panelHTML(t, html, "codex")

	// Non-vacuity: a slicing bug producing an empty region must not pass.
	if n := strings.Count(body, `class="key"`); n == 0 {
		t.Fatalf("codex panel region contains no mirror rows (region length %d) — the assertions below would be vacuous", len(body))
	}

	// AC-WCP-003 arm 1: the symptom.
	if n := strings.Count(body, `name="`); n != 0 {
		t.Errorf(`codex panel region contains %d occurrences of name="; want 0 (REQ-WCP-002)`, n)
	}
	// AC-WCP-003 arm 2: the condition — a form control in a read-only panel.
	for _, tag := range []string{"<input", "<select", "<textarea"} {
		if n := strings.Count(body, tag); n != 0 {
			t.Errorf("codex panel region contains %d %q elements; want 0", n, tag)
		}
	}
	// AC-WCP-004: the hidden bool companion, asserted separately. A mirror
	// emitting only the companion submits no value for the field, which
	// parseSchemaForm records as an explicit false — six MCP tools would be
	// turned off by opening a page.
	if n := strings.Count(body, "__present"); n != 0 {
		t.Errorf("codex panel region contains %d __present companions; want 0 (a lone companion is read as an explicit false)", n)
	}
}

// ─── AC-WCP-005 ──────────────────────────────────────────────────────────

// TestCodexMirrorFieldsStayOnOwningPanel asserts the per-field, per-panel
// invariant that actually forecloses the silent edit loss: a mirrored field
// renders as a control in exactly one panel — its owning one.
func TestCodexMirrorFieldsStayOnOwningPanel(t *testing.T) {
	html := renderConsolePage(t)
	for name, panel := range codexMirrorOwningPanel {
		body := panelHTML(t, html, panel)
		marker := `name="` + name + `"`
		inPanel := strings.Count(body, marker)
		if inPanel == 0 {
			t.Errorf("panel %q missing control %q (the mirror must not remove a field from its owning tab)", panel, name)
			continue
		}
		if total := strings.Count(html, marker); total != inPanel {
			t.Errorf("control %q renders %d times page-wide but %d times inside panel %q (a field must live on exactly one tab)", name, total, inPanel, panel)
		}
	}
}

// ─── AC-WCP-006 ──────────────────────────────────────────────────────────

// TestCodexMirrorCoverage compares the mirror predicate against an
// independently pinned list in BOTH directions, then sweeps the registry for
// codex-token fields the predicate never saw (REQ-WCP-009).
func TestCodexMirrorCoverage(t *testing.T) {
	want := map[string]bool{}
	for _, n := range wantCodexTokenFields {
		want[n] = true
	}

	got := map[string]bool{}
	for _, n := range codexMirrorFieldNames() {
		got[n] = true
	}
	for n := range want {
		if !got[n] {
			t.Errorf("mirror predicate misses codex field %q", n)
		}
	}
	for n := range got {
		if !want[n] {
			t.Errorf("mirror predicate returns %q, which is not in the pinned list", n)
		}
	}

	// The registry-sweep arm — the only non-circular carrier of REQ-WCP-009.
	// A narrowed predicate stops expecting what it stopped matching, so the
	// comparison above cannot detect a registry field the predicate never saw.
	sweep := map[string]bool{}
	for _, f := range settings.AllFields() {
		if strings.Contains(f.Name, "codex") {
			sweep[f.Name] = true
		}
	}
	for n := range sweep {
		if !want[n] {
			t.Errorf("registry field %q contains a codex token but is absent from the pinned mirror list", n)
		}
	}
	for n := range want {
		if !sweep[n] {
			t.Errorf("pinned mirror field %q is absent from settings.AllFields()", n)
		}
	}
}

// ─── AC-WCP-007 ──────────────────────────────────────────────────────────

// TestCodexMirrorRowLinksToOwningTab asserts each mirror row carries the
// dot-path identifier, the current disk value, and a link to the owning tab.
// The values are sentinels seeded on disk, so a value read from anywhere other
// than the shared view model would not carry them.
func TestCodexMirrorRowLinksToOwningTab(t *testing.T) {
	const workflowYAML = `workflow:
  audit:
    model: codex
    codex:
      model: sentinel-codex-model-x7
      effort: high
  codex:
    review_gate:
      enabled: true
    task:
      allow_write: true
`
	const mcpYAML = `mcp:
  tools:
    codex_audit:
      enabled: false
`
	a := codexPanelTestApp(t, workflowYAML, mcpYAML)
	html := renderAppBody(t, a)
	body := panelHTML(t, html, "codex")

	// Every mirrored field's dot-path identifier appears in the region.
	for name := range codexMirrorOwningPanel {
		if !strings.Contains(body, ">"+name+"<") {
			t.Errorf("codex panel region missing dot-path identifier for %q", name)
		}
	}

	// The audit-side sentinel: a free-form text value read from workflow.yaml.
	if !strings.Contains(body, "sentinel-codex-model-x7") {
		t.Error("codex panel region does not carry the seeded workflow.audit.codex.model value")
	}
	// The MCP-side sentinel: an explicit false on one tool, distinguishable
	// from the five that carry no explicit value.
	if got := codexRowValue(t, body, "mcp.tools.codex_audit.enabled"); got != "false" {
		t.Errorf("mirror row for mcp.tools.codex_audit.enabled shows value %q, want %q", got, "false")
	}
	if got := codexRowValue(t, body, "workflow.codex.task.allow_write"); got != "true" {
		t.Errorf("mirror row for workflow.codex.task.allow_write shows value %q, want %q", got, "true")
	}

	// The owning-tab links.
	for _, want := range []string{`href="/settings?tab=audit"`, `href="/settings?tab=mcp"`} {
		if !strings.Contains(body, want) {
			t.Errorf("codex panel region missing owning-tab link %s", want)
		}
	}
	// Each row's link points at ITS owner, not merely at some owner.
	for name, panel := range codexMirrorOwningPanel {
		row := codexRowHTML(t, body, name)
		want := `href="/settings?tab=` + panel + `"`
		if !strings.Contains(row, want) {
			t.Errorf("mirror row %q links to the wrong owning tab (want %s) — row: %s", name, want, row)
		}
	}
}

// codexRowHTML slices the markup of one mirror row: from the row's dot-path
// identifier to the next row boundary (or the end of the region).
func codexRowHTML(t *testing.T, region, name string) string {
	t.Helper()
	marker := ">" + name + "<"
	i := strings.Index(region, marker)
	if i < 0 {
		t.Fatalf("mirror row for %q not found in the codex panel region", name)
	}
	rest := region[i+len(marker):]
	if next := strings.Index(rest, `class="key"`); next >= 0 {
		rest = rest[:next]
	}
	return rest
}

// codexRowValue reads the displayed value out of one mirror row.
func codexRowValue(t *testing.T, region, name string) string {
	t.Helper()
	row := codexRowHTML(t, region, name)
	const open = `class="in in--ro"`
	i := strings.Index(row, open)
	if i < 0 {
		t.Fatalf("mirror row for %q carries no value span", name)
	}
	rest := row[i+len(open):]
	j := strings.Index(rest, ">")
	if j < 0 {
		t.Fatalf("mirror row for %q has a malformed value span", name)
	}
	rest = rest[j+1:]
	k := strings.Index(rest, "<")
	if k < 0 {
		t.Fatalf("mirror row for %q has an unterminated value span", name)
	}
	return strings.TrimSpace(rest[:k])
}

// ─── AC-WCP-008 ──────────────────────────────────────────────────────────

// TestCodexPanel_ProbeSentinel asserts the codex panel CONSUMES the injected
// probe rather than reclassifying it. The three sentinels already render today
// on the MCP panel via codexAuthBlock, so the whole-body form of this
// assertion is green before the codex panel exists — which is why the first
// two arms are scoped to the codex region and the whole-body arm survives only
// to prove the MCP surface was not disturbed.
func TestCodexPanel_ProbeSentinel(t *testing.T) {
	state := CodexStateView{
		Installed:    true,
		Binary:       "/sentinel/bin/codex-x7",
		Version:      "9.8.7-sentinel",
		AuthProvider: "sentinel-provider-x7",
	}
	a := codexPanelTestApp(t, "", "")
	a.codexStateProbe = func(_ context.Context) CodexStateView { return state }
	html := renderAppBody(t, a)

	body := panelHTML(t, html, "codex")
	for _, want := range []string{state.Binary, state.Version, state.AuthProvider} {
		if !strings.Contains(body, want) {
			t.Errorf("codex panel region missing probe sentinel %q (the panel must display the probe verbatim, not reclassify it)", want)
		}
	}

	// The MCP surface still carries the same three sentinels, unchanged.
	mcp := panelHTML(t, html, "mcp")
	for _, want := range []string{state.Binary, state.Version, state.AuthProvider} {
		if !strings.Contains(mcp, want) {
			t.Errorf("MCP panel region lost probe sentinel %q — the mirror must disturb nothing", want)
		}
	}

	// The not-installed branch renders in the codex region.
	a2 := codexPanelTestApp(t, "", "")
	a2.codexStateProbe = func(_ context.Context) CodexStateView { return CodexStateView{Installed: false} }
	body2 := panelHTML(t, renderAppBody(t, a2), "codex")
	if !strings.Contains(body2, "f.mcp.codex.not_installed") {
		t.Error("codex panel region does not render the not-installed state when the probe reports the binary absent")
	}
}

// ─── AC-WCP-013 ──────────────────────────────────────────────────────────

// TestCodexTabRailCount asserts the rail count and the panel header agree at
// zero: the panel owns no fields, so the rail shows 0 and the header states no
// count at all. Announcing a field count on a page where nothing is editable
// would be a claim about editability the panel cannot honour.
func TestCodexTabRailCount(t *testing.T) {
	if names := settingsTabFieldNames("codex"); len(names) != 0 {
		t.Errorf("settingsTabFieldNames(\"codex\") = %v, want empty (the panel owns no fields)", names)
	}
	if n := settingsTabFieldCount("codex", pageView{}, settingsTabFieldNames("codex")); n != 0 {
		t.Errorf("settingsTabFieldCount(\"codex\") = %d, want 0", n)
	}
	body := panelHTML(t, renderConsolePage(t), "codex")
	if strings.Contains(body, `class="panel__meta"`) {
		t.Error("codex panel header states a field count; it must state none (the panel owns no fields)")
	}
}

// ─── AC-WCP-014 ──────────────────────────────────────────────────────────

// TestCodexMirrorDeclaredException asserts the one declared exception —
// workflow.audit.model, the shared audit backend selector — is present in the
// panel, labelled as shared, and linked to Audit; and that no predicate
// reaches it. Asserted apart from AC-WCP-006 so the derivation invariant and
// the judgement row each stay falsifiable.
func TestCodexMirrorDeclaredException(t *testing.T) {
	body := panelHTML(t, renderConsolePage(t), "codex")
	if !strings.Contains(body, ">workflow.audit.model<") {
		t.Fatal("codex panel region missing the declared exception row workflow.audit.model")
	}
	row := codexRowHTML(t, body, "workflow.audit.model")
	if !strings.Contains(row, `href="/settings?tab=audit"`) {
		t.Error("the declared-exception row does not link to the audit tab")
	}
	if !strings.Contains(row, codexSharedBackendI18nKey) {
		t.Errorf("the declared-exception row is not labelled as the shared audit backend selector (want data-i18n %q)", codexSharedBackendI18nKey)
	}
	for _, n := range codexMirrorFieldNames() {
		if n == "workflow.audit.model" {
			t.Error("the mirror predicate returns workflow.audit.model; it carries no codex token and must arrive as the declared exception instead")
		}
	}
}

// ─── AC-WCP-010 ──────────────────────────────────────────────────────────

// TestCodexPanelI18nKeysInAllLocales pins the new dictionary keys in all four
// locale blocks.
func TestCodexPanelI18nKeysInAllLocales(t *testing.T) {
	for _, k := range codexPanelI18nKeys {
		if !i18nKeyInAllLocales(t, k) {
			t.Errorf("i18n.js missing codex panel key %q in all 4 locales", k)
		}
	}
}
