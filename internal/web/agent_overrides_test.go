package web

// agent_overrides_test.go — SPEC-WEB-AGENTFM-RESTORE-001 M2: the inverted
// agent-settings tests (plan §D.5 row 1 — the observational-RED baseline).
// agent_settings_removed_test.go asserted the surface's ABSENCE
// (SPEC-AGENT-MODEL-INHERIT-001 REQ-AMI-011 — its two absence-mode tests are
// deleted with that file, unquoted here so the AC-AFR-012 0-hit grep stays
// clean); card t1411 restores the console surface, so these tests assert its
// PRESENCE + EFFECT instead:
//
//   - TestAgentOverridesSubsection  — the llm-panel sub-section renders every
//     parseable .claude/agents/moai/ agent with wired selects (REQ-AFR-001);
//     a parse-failed row downgrades to agentfm.unavailable; harness rows are
//     scanned but never rendered.
//   - TestPerfTierSave              — the profile selector persists
//     llm.profile (never the retired llm.performance_tier alias) and treats
//     custom/empty as preserve (REQ-AFR-003).
//   - TestAgentOverridesSave        — per-agent pin/clear against the profile
//     default, non-matrix skip, effort backfill, and the atomic reject
//     (REQ-AFR-004/006/007).
//   - TestAgentFrontmatterUntouched — a legal submission leaves agent
//     frontmatter byte-identical (REQ-AFR-005, REQ-MPM-040 succession).
//
// These are written BEFORE the M3 save path and M4 render exist: their first
// run is the verbatim RED recorded in progress.md §E.2 (E8), and M3/M4 make
// them pass. The two absence-mode tests this file replaces are deleted with
// their file — AC-AFR-012 pins their identifier 0-hit, so they are
// deliberately not quoted here.

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/settings/agentfm"
	"github.com/modu-ai/moai-adk/internal/template"
	"gopkg.in/yaml.v3"
)

// seedAgentOverridesProject seeds a schema-fixture project with agent
// definitions: two parseable moai agents, one frontmatter-broken moai file,
// and one harness specialist (scanned, never rendered).
func seedAgentOverridesProject(t *testing.T) (*app, string) {
	t.Helper()
	a, root := newSchemaTestApp(t)
	moai := filepath.Join(root, ".claude", "agents", "moai")
	harness := filepath.Join(root, ".claude", "agents", "harness")
	for dir := range map[string]bool{moai: true, harness: true} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		filepath.Join(moai, "manager-develop.md"): "---\nname: manager-develop\ndescription: Implementation specialist for run-phase work\n---\nbody\n",
		filepath.Join(moai, "manager-todo.md"):    "---\nname: manager-todo\ndescription: Todo-queue management agent\n---\nbody\n",
		filepath.Join(moai, "broken-agent.md"):    "no frontmatter delimiter here\n",
		filepath.Join(harness, "hns-probe.md"):    "---\nname: hns-probe\ndescription: harness specialist\n---\nbody\n",
	}
	for p, b := range files {
		if err := os.WriteFile(p, []byte(b), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return a, root
}

// writeLLMYAML replaces the seeded llm fixture with the given body (the
// pin/clear/no-op assertions need exact byte control).
func writeLLMYAML(t *testing.T, root, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// llmYAMLMap parses the project's llm.yaml into a map for parsed-key
// assertions.
func llmYAMLMap(t *testing.T, root string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"))
	if err != nil {
		t.Fatalf("read llm.yaml: %v", err)
	}
	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal llm.yaml: %v", err)
	}
	return doc
}

// agentOverridesForm is the minimal valid save submission that reaches the
// llm persist steps (profile fields only — no llm.yaml side effects).
func agentOverridesForm(extra map[string]string) url.Values {
	form := url.Values{
		"__profile":       {"default"},
		"permission_mode": {"acceptEdits"},
	}
	for k, v := range extra {
		form.Set(k, v)
	}
	return form
}

// TestAgentOverridesSubsection covers AC-AFR-001 (REQ-AFR-001): the llm panel
// renders the agent-overrides sub-section with one row per parseable moai
// agent, wired model/effort selects, and the section title. Parse-failed rows
// downgrade (no selects, no page failure); harness rows never render.
func TestAgentOverridesSubsection(t *testing.T) {
	a, _ := seedAgentOverridesProject(t)

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /settings status = %d", rec.Code)
	}
	body := rec.Body.String()

	for _, marker := range []string{
		`data-section="agent-overrides"`,
		`sec.agentfm.title`,
	} {
		if !strings.Contains(body, marker) {
			t.Errorf("GET /settings lacks %q — the restored sub-section is not rendered yet", marker)
		}
	}
	for _, name := range []string{"manager-develop", "manager-todo"} {
		for _, marker := range []string{
			`data-agent-row="` + name + `"`,
			`name="agentfm.` + name + `.model"`,
			`name="agentfm.` + name + `.effort"`,
		} {
			if !strings.Contains(body, marker) {
				t.Errorf("GET /settings lacks %q", marker)
			}
		}
	}
	// Parse-failed row: downgraded (the canonical shape renders the note and
	// NO editable row marker — a non-editable row is not a data-agent-row).
	if !strings.Contains(body, "agentfm.unavailable") {
		t.Error("the parse-failed agent row must render downgraded as agentfm.unavailable")
	}
	if strings.Contains(body, `name="agentfm.broken-agent.`) {
		t.Error("the parse-failed row must not render model/effort selects")
	}
	// Harness rows are scanned but not rendered.
	if strings.Contains(body, `data-agent-row="hns-probe"`) || strings.Contains(body, `name="agentfm.hns-probe.`) {
		t.Error("a .claude/agents/harness/ row must not render in the sub-section")
	}
}

// TestPerfTierSave covers AC-AFR-002 (REQ-AFR-003): a non-custom tier
// submission persists llm.profile verbatim (a max submission records
// profile: "max" — the selector's wire value) and never creates the retired
// llm.performance_tier key; custom and empty submissions preserve the file
// byte-identically.
func TestPerfTierSave(t *testing.T) {
	t.Run("max persists to llm.profile", func(t *testing.T) {
		a, root := seedAgentOverridesProject(t)
		writeLLMYAML(t, root, "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n")

		rec := postSave(t, a, agentOverridesForm(map[string]string{"performance_tier": "max"}))
		if rec.Code != http.StatusOK {
			t.Fatalf("save status = %d; body:\n%s", rec.Code, rec.Body.String())
		}
		doc := llmYAMLMap(t, root)
		llm, _ := doc["llm"].(map[string]any)
		if llm["profile"] != "max" {
			t.Errorf("llm.profile = %v, want \"max\" (the selector wire value persists verbatim)", llm["profile"])
		}
		if _, exists := llm["performance_tier"]; exists {
			t.Error("llm.performance_tier must not be created — the alias key stays retired")
		}
	})

	t.Run("custom preserves byte-identical", func(t *testing.T) {
		a, root := seedAgentOverridesProject(t)
		body := "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n"
		writeLLMYAML(t, root, body)

		rec := postSave(t, a, agentOverridesForm(map[string]string{"performance_tier": "custom"}))
		if rec.Code != http.StatusOK {
			t.Fatalf("save status = %d", rec.Code)
		}
		if got := readSectionFile(t, root, "llm"); got != body {
			t.Errorf("custom submission changed llm.yaml:\n%s", got)
		}
	})

	t.Run("empty submission preserves byte-identical", func(t *testing.T) {
		a, root := seedAgentOverridesProject(t)
		body := "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n"
		writeLLMYAML(t, root, body)

		rec := postSave(t, a, agentOverridesForm(nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("save status = %d", rec.Code)
		}
		if got := readSectionFile(t, root, "llm"); got != body {
			t.Errorf("tierless submission changed llm.yaml:\n%s", got)
		}
	})
}

// TestAgentOverridesSave covers AC-AFR-003 + AC-AFR-004 (REQ-AFR-004/002/006):
// pins, clears against the profile default, non-matrix skip, effort backfill,
// and the atomic out-of-set reject.
func TestAgentOverridesSave(t *testing.T) {
	// The medium-column default cell for manager-develop (the no-profile
	// baseline the pin/clear comparisons resolve against).
	const devDefaultModel = "opus"
	const devDefaultEffort = "medium"

	t.Run("override pin", func(t *testing.T) {
		a, root := seedAgentOverridesProject(t)
		writeLLMYAML(t, root, "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n")

		rec := postSave(t, a, agentOverridesForm(map[string]string{
			"agentfm.manager-develop.model":  "opus",
			"agentfm.manager-develop.effort": "xhigh",
		}))
		if rec.Code != http.StatusOK {
			t.Fatalf("save status = %d; body:\n%s", rec.Code, rec.Body.String())
		}
		llm, _ := llmYAMLMap(t, root)["llm"].(map[string]any)
		ov, _ := llm["agent_overrides"].(map[string]any)
		dev, _ := ov["manager-develop"].(map[string]any)
		if dev == nil || dev["model"] != "opus" || dev["effort"] != "xhigh" {
			t.Errorf("llm.agent_overrides.manager-develop = %v, want {model: opus, effort: xhigh}", dev)
		}
	})

	t.Run("clear on profile-default submission, sibling preserved", func(t *testing.T) {
		a, root := seedAgentOverridesProject(t)
		writeLLMYAML(t, root, "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n"+
			"  agent_overrides:\n    manager-develop:\n      model: opus\n      effort: xhigh\n"+
			"    manager-todo:\n      model: haiku\n      effort: low\n")

		// Submitting the profile-default pair for manager-develop clears its
		// override; manager-todo is unsubmitted, so its override must survive.
		rec := postSave(t, a, agentOverridesForm(map[string]string{
			"agentfm.manager-develop.model":  devDefaultModel,
			"agentfm.manager-develop.effort": devDefaultEffort,
		}))
		if rec.Code != http.StatusOK {
			t.Fatalf("save status = %d; body:\n%s", rec.Code, rec.Body.String())
		}
		llm, _ := llmYAMLMap(t, root)["llm"].(map[string]any)
		ov, _ := llm["agent_overrides"].(map[string]any)
		if _, exists := ov["manager-develop"]; exists {
			t.Errorf("override equal to the profile default must be cleared: %v", ov)
		}
		todo, _ := ov["manager-todo"].(map[string]any)
		if todo == nil || todo["model"] != "haiku" {
			t.Errorf("unsubmitted sibling override must survive: %v", ov)
		}
	})

	t.Run("no-op submission keeps llm.yaml byte-identical", func(t *testing.T) {
		a, root := seedAgentOverridesProject(t)
		body := "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n"
		writeLLMYAML(t, root, body)

		rec := postSave(t, a, agentOverridesForm(nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("save status = %d", rec.Code)
		}
		if got := readSectionFile(t, root, "llm"); got != body {
			t.Errorf("a no-op submission rewrote llm.yaml:\n%s", got)
		}
	})

	t.Run("non-matrix agent creates no override", func(t *testing.T) {
		a, root := seedAgentOverridesProject(t)
		writeLLMYAML(t, root, "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n")

		rec := postSave(t, a, agentOverridesForm(map[string]string{
			"agentfm.user-added-agent.model":  "opus",
			"agentfm.user-added-agent.effort": "high",
		}))
		if rec.Code != http.StatusOK {
			t.Fatalf("save status = %d", rec.Code)
		}
		llm, _ := llmYAMLMap(t, root)["llm"].(map[string]any)
		if ov, exists := llm["agent_overrides"]; exists && len(asMap(t, ov)) > 0 {
			t.Errorf("a non-matrix submission must not create overrides: %v", ov)
		}
	})

	t.Run("unsubmitted effort backfills the resolved value", func(t *testing.T) {
		a, root := seedAgentOverridesProject(t)
		writeLLMYAML(t, root, "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n")

		// A disabled haiku-style effort select submits model only; the save
		// backfills effort with the resolved value so the pin is complete.
		rec := postSave(t, a, agentOverridesForm(map[string]string{
			"agentfm.manager-develop.model": "sonnet",
		}))
		if rec.Code != http.StatusOK {
			t.Fatalf("save status = %d", rec.Code)
		}
		llm, _ := llmYAMLMap(t, root)["llm"].(map[string]any)
		ov, _ := llm["agent_overrides"].(map[string]any)
		dev, _ := ov["manager-develop"].(map[string]any)
		if dev == nil || dev["model"] != "sonnet" || dev["effort"] != devDefaultEffort {
			t.Errorf("backfilled override = %v, want {model: sonnet, effort: %s}", dev, devDefaultEffort)
		}
	})

	t.Run("out-of-set effort is an atomic reject", func(t *testing.T) {
		a, root := seedAgentOverridesProject(t)
		llmBody := "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n"
		writeLLMYAML(t, root, llmBody)
		agentPath := filepath.Join(root, ".claude", "agents", "moai", "manager-develop.md")
		agentBefore, err := os.ReadFile(agentPath)
		if err != nil {
			t.Fatal(err)
		}

		rec := postSave(t, a, agentOverridesForm(map[string]string{
			"agentfm.manager-develop.effort": "absurd",
		}))
		if rec.Code != http.StatusOK {
			t.Fatalf("save status = %d, want the 200 re-render", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "agentfm.manager-develop.effort") {
			t.Error("the re-render must carry a per-field error for the rejected effort")
		}
		if got := readSectionFile(t, root, "llm"); got != llmBody {
			t.Errorf("an atomic reject rewrote llm.yaml:\n%s", got)
		}
		agentAfter, err := os.ReadFile(agentPath)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(agentBefore, agentAfter) {
			t.Error("an atomic reject touched agent frontmatter")
		}
	})

	t.Run("out-of-set tier is an atomic reject", func(t *testing.T) {
		a, root := seedAgentOverridesProject(t)
		llmBody := "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n"
		writeLLMYAML(t, root, llmBody)

		rec := postSave(t, a, agentOverridesForm(map[string]string{"performance_tier": "bogus"}))
		if rec.Code != http.StatusOK {
			t.Fatalf("save status = %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "performance_tier") {
			t.Error("the re-render must carry a per-field error for the rejected tier")
		}
		if got := readSectionFile(t, root, "llm"); got != llmBody {
			t.Errorf("an atomic reject rewrote llm.yaml:\n%s", got)
		}
	})
}

// asMap narrows an any carrying a YAML map for assertions.
func asMap(t *testing.T, v any) map[string]any {
	t.Helper()
	m, ok := v.(map[string]any)
	if !ok {
		t.Fatalf("expected a map, got %T (%v)", v, v)
	}
	return m
}

// TestAgentFrontmatterUntouched covers AC-AFR-005 (REQ-AFR-005): a fully
// legal submission — a tier change plus an override pin — leaves every agent
// frontmatter file byte-identical (the REQ-MPM-040 contract succession; the
// write layer of the old agentfm package stays dead).
func TestAgentFrontmatterUntouched(t *testing.T) {
	a, root := seedAgentOverridesProject(t)
	moai := filepath.Join(root, ".claude", "agents", "moai")
	before := map[string][]byte{}
	entries, err := os.ReadDir(moai)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		b, err := os.ReadFile(filepath.Join(moai, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		before[e.Name()] = b
	}

	rec := postSave(t, a, agentOverridesForm(map[string]string{
		"performance_tier":               "max",
		"agentfm.manager-develop.model":  "opus",
		"agentfm.manager-develop.effort": "xhigh",
	}))
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d; body:\n%s", rec.Code, rec.Body.String())
	}
	for name := range before {
		got, err := os.ReadFile(filepath.Join(moai, name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before[name], got) {
			t.Errorf("agent frontmatter %s changed — the console must never write frontmatter", name)
		}
	}
}

// TestHaikuEffortLock covers AC-AFR-009 (REQ-AFR-009): a row whose resolved
// model is haiku renders its effort select disabled with the data-haiku-hint
// note; the client wiring re-applies the lock on model change and on tier
// repopulation (the app.js half — asserted here at the wire-function level by
// the re-port's presence and the initial server render).
func TestHaikuEffortLock(t *testing.T) {
	a, root := seedAgentOverridesProject(t)
	// manager-develop pinned to haiku: the row must render the lock state.
	writeLLMYAML(t, root, "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n"+
		"  agent_overrides:\n    manager-develop:\n      model: haiku\n      effort: low\n")

	body := renderSettingsBody(t, a)

	row := sliceAgentRow(t, body, "manager-develop")
	if !strings.Contains(row, `name="agentfm.manager-develop.effort" disabled`) &&
		!strings.Contains(row, `disabled?`) && !strings.Contains(row, "disabled") {
		t.Errorf("the haiku row's effort select must render disabled:\n%s", row)
	}
	if !strings.Contains(row, "data-haiku-hint") {
		t.Errorf("the haiku row must carry the data-haiku-hint note:\n%s", row)
	}
	if !strings.Contains(body, `hint.effort.haiku_na`) {
		t.Error("the haiku hint i18n key must be referenced")
	}
	// A non-haiku row stays enabled: its effort select tag must not carry the
	// disabled attribute (checked on the select tag itself — the row slice can
	// run to the section end where other panels' markup appears).
	other := sliceAgentRow(t, body, "manager-todo")
	selectStart := strings.Index(other, `name="agentfm.manager-todo.effort"`)
	if selectStart < 0 {
		t.Fatal("manager-todo effort select not rendered")
	}
	selectTag := other[selectStart:]
	if cut := strings.Index(selectTag, ">"); cut >= 0 {
		selectTag = selectTag[:cut]
	}
	if strings.Contains(selectTag, "disabled") {
		t.Errorf("a non-haiku row must not disable its effort select: <select %s>", selectTag)
	}
	// The client lock wiring ships (re-applied on model change and tier
	// repopulation — REQ-AFR-009's re-application clause).
	js, err := os.ReadFile(filepath.Join("assets", "app.js"))
	if err != nil {
		t.Fatal(err)
	}
	for _, fn := range []string{"function applyHaikuEffortLock(", "function wireHaikuEffortLock(", "function reapplyHaikuLocks(", "reapplyHaikuLocks();"} {
		if !strings.Contains(string(js), fn) {
			t.Errorf("app.js lacks the haiku-lock wiring %q (the client re-application contract)", fn)
		}
	}
}

// TestGLMReasoningColumn covers AC-AFR-010 (REQ-AFR-010): under a GLM backend
// each row carries data-glm-reasoning whose value equals
// template.ResolveGLMReasoningForModel's reading of the row's resolved effort
// (the single-derivation assertion); under a Claude backend no row carries it.
func TestGLMReasoningColumn(t *testing.T) {
	t.Run("glm backend renders the column", func(t *testing.T) {
		a, root := seedAgentOverridesProject(t)
		writeLLMYAML(t, root, "llm:\n  mode: \"\"\n  team_mode: glm\n  glm_env_var: GLM_API_KEY\n"+
			"  glm:\n    base_url: https://api.z.ai/api/anthropic\n"+
			"    models:\n      high: glm-5.3\n      medium: glm-5.3\n      low: glm-5.3\n      fable: glm-5.3\n")

		body := renderSettingsBody(t, a)
		// Single-derivation: the rendered value must equal the resolver's
		// reading for the same row (manager-develop carries the coding-max
		// override → max even at low effort).
		want := template.ResolveGLMReasoningForModel("glm-5.3", "manager-develop", "medium").Name
		if want == "" {
			t.Fatal("resolver returned an empty state — the single-derivation comparison is vacuous")
		}
		row := sliceAgentRow(t, body, "manager-develop")
		if !strings.Contains(row, `data-glm-reasoning="`+want+`"`) {
			t.Errorf("the GLM row must carry data-glm-reasoning=%q (resolver agreement):\n%s", want, row)
		}
	})
	t.Run("claude backend renders nothing extra", func(t *testing.T) {
		a, _ := seedAgentOverridesProject(t)
		body := renderSettingsBody(t, a)
		if strings.Contains(body, "data-glm-reasoning") {
			t.Error("a Claude backend must not render the GLM reasoning column")
		}
	})
}

// renderSettingsBody renders GET /settings and returns the body.
func renderSettingsBody(t *testing.T, a *app) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /settings status = %d", rec.Code)
	}
	return rec.Body.String()
}

// sliceAgentRow extracts one agent row's markup (from its data-agent-row
// marker to the next row marker or the section end) for row-scoped assertions.
func sliceAgentRow(t *testing.T, body, name string) string {
	t.Helper()
	start := strings.Index(body, `data-agent-row="`+name+`"`)
	if start < 0 {
		t.Fatalf("row %s not rendered", name)
	}
	end := len(body)
	if next := strings.Index(body[start+1:], `data-agent-row="`); next >= 0 {
		end = start + 1 + next
	}
	return body[start:end]
}

// TestAgentOverridesSeams covers AC-AFR-011 (REQ-AFR-012): the save path's
// list/parse/persist seams are injectable — substituting them observes ONLY
// the substitution (the substituted persist owns the write, so llm.yaml is
// untouched by the real seam), and the default wiring stays in place on a
// fresh app (agentfm.List / applyAgentOverrides assigned by newApp).
func TestAgentOverridesSeams(t *testing.T) {
	a, root := seedAgentOverridesProject(t)
	writeLLMYAML(t, root, "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n")

	var observed []string
	defaultList := a.listAgentFMs
	a.listAgentFMs = func(dir string) ([]agentfm.AgentInfo, error) {
		observed = append(observed, "listAgentFMs")
		return defaultList(dir) // chain to the default scan: the parse path needs the roster
	}
	a.patchAgentFM = func(string, map[string]config.ModelEffort, []string) error {
		observed = append(observed, "patchAgentFM")
		return nil // the substituted persist writes nothing
	}

	rec := postSave(t, a, agentOverridesForm(map[string]string{
		"agentfm.manager-develop.model":  "opus",
		"agentfm.manager-develop.effort": "xhigh",
	}))
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d; body:\n%s", rec.Code, rec.Body.String())
	}
	seen := strings.Join(observed, ",")
	if !strings.Contains(seen, "listAgentFMs") || !strings.Contains(seen, "patchAgentFM") {
		t.Errorf("the substituted seams were not both observed: %s", seen)
	}
	if got := readSectionFile(t, root, "llm"); strings.Contains(got, "opus") {
		t.Errorf("the real persist ran although the seam was substituted:\n%s", got)
	}

	// Default wiring identity on a fresh app: both seams are non-nil (the
	// canonical agentfm.List / applyAgentOverrides assignments), so the
	// substitution above is the only actor that differed.
	fresh := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	if fresh.listAgentFMs == nil || fresh.patchAgentFM == nil || fresh.applyPerfTierEdits == nil {
		t.Fatal("default agent-overrides wiring is nil on a fresh app — the seams were not wired")
	}
}

// slicePerfTierRadio extracts the perf-tier radio group's markup: from the
// first performance_tier radio to the matrix island that follows the group.
// The whole-page `value="max" checked` assertion is NOT usable here — the GLM
// effort selects legitimately render `value="max" checked` from their own
// fixture defaults (measured: 4 such occurrences), so the assertion must be
// scoped to the radio group.
func slicePerfTierRadio(t *testing.T, body string) string {
	t.Helper()
	start := strings.Index(body, `name="performance_tier"`)
	if start < 0 {
		t.Fatal("perf-tier radio group not rendered")
	}
	end := strings.Index(body[start:], `id="moai-profile-matrix"`)
	if end < 0 {
		t.Fatal("matrix island not rendered after the radio group")
	}
	return body[start : start+end]
}

// TestPerfTierSeedRoundTrip covers the F2 fix (sync-audit, card t1411): the
// perf-tier radio re-selects what was saved. The selector's wire set is
// {max, medium, low}; a stored "max" folded by EffectiveProfile to "high"
// matched no option — the render boundary restores the top wire value so
// both spellings of the column (the selector's "max", the config-canonical
// "high") round-trip to a checked max radio.
func TestPerfTierSeedRoundTrip(t *testing.T) {
	t.Run("stored max renders a checked max radio", func(t *testing.T) {
		a, root := seedAgentOverridesProject(t)
		writeLLMYAML(t, root, "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n  profile: \"max\"\n")

		radio := slicePerfTierRadio(t, renderSettingsBody(t, a))
		if !strings.Contains(radio, `value="max" checked`) {
			t.Errorf("a stored max must render the max radio checked (the fold must not disconnect the round-trip):\n%s", radio)
		}
	})
	t.Run("stored high renders the same top column", func(t *testing.T) {
		a, root := seedAgentOverridesProject(t)
		writeLLMYAML(t, root, "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n  profile: \"high\"\n")

		radio := slicePerfTierRadio(t, renderSettingsBody(t, a))
		if !strings.Contains(radio, `value="max" checked`) {
			t.Errorf("a stored canonical high is the same column as the max wire value — the radio must show it checked:\n%s", radio)
		}
	})
	t.Run("save max then re-render keeps the radio and the value", func(t *testing.T) {
		a, root := seedAgentOverridesProject(t)
		writeLLMYAML(t, root, "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n")

		rec := postSave(t, a, agentOverridesForm(map[string]string{"performance_tier": "max"}))
		if rec.Code != http.StatusOK {
			t.Fatalf("save status = %d", rec.Code)
		}
		// AC-AFR-002: the wire value persists verbatim. Parsed-key
		// granularity — the upsert splice may spell the scalar without
		// quotes, which is the same YAML value.
		llm := llmYAMLMap(t, root)["llm"].(map[string]any)
		if llm["profile"] != "max" {
			t.Fatalf("llm.profile = %v, want \"max\" persisted verbatim (AC-AFR-002)", llm["profile"])
		}
		radio := slicePerfTierRadio(t, renderSettingsBody(t, a))
		if !strings.Contains(radio, `value="max" checked`) {
			t.Errorf("the save→re-render round trip lost the max selection:\n%s", radio)
		}
	})
}

// TestAgentOverridesPersistFailureRollsBack is the F3 repair probe
// (sync-audit, card t1411): the two llm.yaml writes of steps 7→8
// (applyPerfTierEdits → patchAgentFM) are one logical persistence unit —
// REQ-AFR-007 covers persistence errors, so a step-8 failure must not leave
// step 7's profile write on disk. The probe injects the failure at
// stepPatchAgentFM and asserts the byte-identical rollback.
func TestAgentOverridesPersistFailureRollsBack(t *testing.T) {
	a, root := seedAgentOverridesProject(t)
	llmSeed := "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n"
	writeLLMYAML(t, root, llmSeed)

	a.patchAgentFM = func(string, map[string]config.ModelEffort, []string) error {
		return assertErr("patchAgentFM injected failure")
	}

	rec := postSave(t, a, agentOverridesForm(map[string]string{
		"performance_tier":               "max",
		"agentfm.manager-develop.model":  "opus",
		"agentfm.manager-develop.effort": "xhigh",
	}))
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d, want the 200 failure re-render", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "agent override write failed") {
		t.Error("the re-render must name the failing persistence step")
	}
	if got := readSectionFile(t, root, "llm"); got != llmSeed {
		t.Errorf("step-8 failure left step-7's write on disk — REQ-AFR-007 persistence atomicity:\nseed:\n%s\ndisk:\n%s", llmSeed, got)
	}
}
