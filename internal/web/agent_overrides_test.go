package web

// agent_overrides_test.go — SPEC-WEB-AGENTFM-RESTORE-001 M2: the inverted
// agent-settings tests (plan §D.5 row 1 — the observational-RED baseline).
// agent_settings_removed_test.go asserted the surface's ABSENCE
// (SPEC-AGENT-MODEL-INHERIT-001 REQ-AMI-011); card t1411 restores the console
// surface, so these tests assert its PRESENCE + EFFECT instead:
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
// them pass. The two absence-mode tests this file replaces
// (TestAgentSettingsTab_IsNotRendered, TestAgentSettingsFields_ArePostedWithoutEffect)
// are deleted with their file — AC-AFR-012 pins their 0-hit.

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	// Parse-failed row: downgraded, and NO selects for it.
	if !strings.Contains(body, `data-agent-row="broken-agent"`) || !strings.Contains(body, "agentfm.unavailable") {
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
