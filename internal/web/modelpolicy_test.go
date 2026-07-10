package web

// SPEC-WEB-CONSOLE-013 M3 — READ-ONLY Model Policy view tests
// (AC-WC13-020..027). The view is READ-ONLY: GET-only route, no write path, no
// form control, no FieldDef persist binding, no status transition.

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// getModelPolicy issues GET /model-policy against a fresh app rooted at root.
func getModelPolicy(t *testing.T, root string) *httptest.ResponseRecorder {
	t.Helper()
	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	req := httptest.NewRequest(http.MethodGet, "/model-policy", nil)
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	return rec
}

// writeWorkflowSection writes root/.moai/config/sections/workflow.yaml.
func writeWorkflowSection(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
}

// writeLLMSection writes root/.moai/config/sections/llm.yaml with a
// performance_tier value.
func writeLLMSection(t *testing.T, root, perfTier string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	body := "llm:\n    performance_tier: \"" + perfTier + "\"\n"
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}
}

// modelPolicyProfilesFixture is a workflow.yaml carrying a populated
// model_routing_profiles map (one cell per perfTier for M-run) so the routing
// table renders declared values instead of pure fallback.
const modelPolicyProfilesFixture = `workflow:
    model_routing_profiles:
        max:
            M-run:
                model: opus
                effort: xhigh
        medium:
            M-run:
                model: sonnet
                effort: high
        low:
            M-run:
                model: sonnet
                effort: low
`

// TestModelPolicyView (AC-WC13-020): the view renders the active performance
// tier + the 3 perfTier × 12 cell routing table.
func TestModelPolicyView(t *testing.T) {
	root := t.TempDir()
	writeLLMSection(t, root, "max")
	writeWorkflowSection(t, root, modelPolicyProfilesFixture)

	rec := getModelPolicy(t, root)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /model-policy = %d, want 200", rec.Code)
	}
	body := rec.Body.String()

	// Active performance tier displayed.
	if !strings.Contains(body, "mp-tier-value") {
		t.Error("model policy view missing the active-tier value block")
	}
	if !strings.Contains(body, ">max<") {
		t.Error("active performance tier 'max' not displayed")
	}

	// 3 perfTier profiles rendered.
	for _, pt := range []string{"max", "medium", "low"} {
		if !strings.Contains(body, ">"+pt+"<") {
			t.Errorf("routing table missing performance-tier profile %q", pt)
		}
	}
	// 3 profiles × 12 cells = the routing table must contain every (tier,phase)
	// combination. Count the profile blocks (3) and a representative cell set.
	if n := strings.Count(body, `class="mp-table"`); n != 3 {
		t.Errorf("mp-table count = %d, want 3 (one per performance tier)", n)
	}
	// Every SPEC tier and phase token appears in the cells.
	for _, tok := range []string{">S<", ">M<", ">L<", ">plan<", ">run<", ">sync<", ">mx<"} {
		if !strings.Contains(body, tok) {
			t.Errorf("routing table missing cell token %q", tok)
		}
	}
	// Declared M-run values are surfaced (max profile → opus/xhigh).
	if !strings.Contains(body, ">opus<") || !strings.Contains(body, ">xhigh<") {
		t.Error("declared max/M-run routing entry (opus/xhigh) not rendered")
	}
}

// TestModelPolicyView_ReadOnlyNoForm (AC-WC13-021): the Model Policy view source
// carries NO form / input / select / hx-post — it is structurally read-only.
func TestModelPolicyView_ReadOnlyNoForm(t *testing.T) {
	for _, name := range []string{"modelpolicy.templ", "modelpolicy_templ.go"} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		src := string(data)
		for _, tok := range []string{"<form", "<input", "<select", "hx-post"} {
			if strings.Contains(src, tok) {
				t.Errorf("%s contains a write affordance %q — the Model Policy view must be READ-ONLY (REQ-WC13-021)", name, tok)
			}
		}
	}
	// The rendered page likewise carries no form control (the interface-language
	// <select> lives in the shared board appbar, which is a separate file).
	body := getModelPolicy(t, t.TempDir()).Body.String()
	// Extract the <main> region — the model-policy content — and assert no form
	// controls inside it (the appbar select is outside <main>).
	mainStart := strings.Index(body, "<main")
	mainEnd := strings.Index(body, "</main>")
	if mainStart < 0 || mainEnd < 0 {
		t.Fatal("rendered model-policy page has no <main> region")
	}
	mainRegion := body[mainStart:mainEnd]
	for _, tok := range []string{"<form", "<input", "<select", "hx-post"} {
		if strings.Contains(mainRegion, tok) {
			t.Errorf("model-policy <main> region contains a write affordance %q (REQ-WC13-021)", tok)
		}
	}
}

// TestModelPolicyView_NoPersistBinding (AC-WC13-022): schema_sections.go and
// sectionapply.go carry NO model_routing_profiles / performance_tier persist
// binding (comments-only matches allowed).
func TestModelPolicyView_NoPersistBinding(t *testing.T) {
	for _, name := range []string{"../settings/schema_sections.go", "../settings/sectionapply.go"} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, ln := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(ln)
			if strings.HasPrefix(trimmed, "//") {
				continue // comments may reference the keys descriptively
			}
			for _, tok := range []string{"model_routing_profiles", "performance_tier"} {
				if strings.Contains(ln, tok) {
					t.Errorf("%s has a non-comment reference to %q — no persist binding may exist (REQ-WC13-021/022)", name, tok)
				}
			}
		}
	}
}

// TestModelPolicyView_LegacyFlatHidden (AC-WC13-023): the rendered view never
// surfaces the legacy flat workflow.model_routing block nor workflow_agents.
func TestModelPolicyView_LegacyFlatHidden(t *testing.T) {
	root := t.TempDir()
	// A workflow.yaml carrying BOTH the legacy flat block and workflow_agents.
	writeWorkflowSection(t, root, `workflow:
    model_routing:
        M-run:
            model: legacy-flat-model
            effort: high
    workflow_agents:
        researcher:
            model: hidden-agent-model
    model_routing_profiles:
        medium:
            M-run:
                model: sonnet
                effort: high
`)
	body := getModelPolicy(t, root).Body.String()
	if strings.Contains(body, "legacy-flat-model") {
		t.Error("legacy flat workflow.model_routing block leaked into the render (REQ-WC13-022)")
	}
	if strings.Contains(body, "hidden-agent-model") || strings.Contains(body, "workflow_agents") {
		t.Error("workflow_agents leaked into the Model Policy render (REQ-WC13-023)")
	}
}

// TestModelPolicyView_EmptyTier (AC-WC13-025): an empty llm.performance_tier
// renders the "(runtime default: medium)" empty-value label.
func TestModelPolicyView_EmptyTier(t *testing.T) {
	root := t.TempDir()
	writeLLMSection(t, root, "") // performance_tier: ""
	rec := getModelPolicy(t, root)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /model-policy (empty tier) = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "(runtime default") {
		t.Error("empty performance_tier did not render the '(runtime default...' label (REQ-WC13-024)")
	}
}

// TestModelPolicyView_AbsentBlock (AC-WC13-026): a project with NO
// model_routing_profiles renders the documented fallback state, not an error.
func TestModelPolicyView_AbsentBlock(t *testing.T) {
	root := t.TempDir() // no config at all → block absent + empty tier
	rec := getModelPolicy(t, root)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /model-policy (absent block) = %d, want 200 (absence is not an error)", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "mp.table.absent") {
		t.Error("absent model_routing_profiles did not render the fallback-state notice (REQ-WC13-026)")
	}
	// The 3 profiles still render (every cell falls back to inherit/medium).
	if n := strings.Count(body, `class="mp-table"`); n != 3 {
		t.Errorf("mp-table count = %d, want 3 even with an absent block (all-fallback)", n)
	}
	if !strings.Contains(body, ">inherit<") || !strings.Contains(body, "mp-fallback") {
		t.Error("all-fallback render missing inherit model or the (fallback) marker")
	}
}

// TestModelPolicyView_GETOnly (AC-WC13-021 method gate): the route responds to
// GET only; mutating methods are rejected 405 (no write path).
func TestModelPolicyView_GETOnly(t *testing.T) {
	a := newApp(Config{ProjectRoot: t.TempDir(), ProfileName: "default"})
	h := a.routes()
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		req := httptest.NewRequest(method, "/model-policy", nil)
		req.Header.Set("Sec-Fetch-Site", "same-origin")
		req.Host = "127.0.0.1"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s /model-policy = %d, want 405 (GET-only)", method, rec.Code)
		}
	}
}

// TestModelPolicyView_I18nParity (AC-WC13-027): every mp.* data-i18n key in the
// rendered view exists in all 4 locale blocks of the dictionary, and the
// disambiguation note distinguishing model_policy vs performance_tier is present.
func TestModelPolicyView_I18nParity(t *testing.T) {
	body := getModelPolicy(t, t.TempDir()).Body.String()
	keyRe := regexp.MustCompile(`data-i18n="(mp\.[^"]+)"`)
	matches := keyRe.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		t.Fatal("no mp.* data-i18n keys found in the rendered model-policy page")
	}
	seen := map[string]bool{}
	for _, m := range matches {
		key := m[1]
		if seen[key] {
			continue
		}
		seen[key] = true
		if !i18nKeyInAllLocales(t, key) {
			t.Errorf("i18n.js missing model-policy key %q in all 4 locales", key)
		}
	}
	// The disambiguation note (REQ-WC13-025): mp.tier.desc mentions BOTH
	// model_policy and performance_tier so the two are not confused.
	dict := readEmbeddedAsset(t, "i18n.js")
	discRe := regexp.MustCompile(`"mp\.tier\.desc":\s*"[^"]*model_policy[^"]*"`)
	if !discRe.MatchString(dict) {
		t.Error("mp.tier.desc must carry the model_policy vs performance_tier disambiguation note (REQ-WC13-025)")
	}
}
