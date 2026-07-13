package web

// SPEC-WEB-CONSOLE-013 M3 — READ-ONLY Model Policy view tests
// (AC-WC13-020..027). The view is READ-ONLY: GET-only route, no write path, no
// form control, no FieldDef persist binding, no status transition.

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
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

// TestModelPolicyView_ScopedWriteInvariant (AC-MTP-021, D2 sanctioned exception):
// the Model Policy view carries EXACTLY ONE write affordance — the plan_type
// selector form posting to /model-policy/plan-type. No OTHER field is writable,
// there is no hx-post, and the <main> region's only <form> targets the scoped
// persist endpoint. This supersedes the former READ-ONLY-no-form assertion, which
// D2 (persisting plan_type selector) intentionally relaxes for exactly one field
// (SPEC-MODEL-TIER-PLANTYPE-001 §B.4 sanctioned exception to REQ-WC13-021).
func TestModelPolicyView_ScopedWriteInvariant(t *testing.T) {
	body := getModelPolicy(t, t.TempDir()).Body.String()
	mainStart := strings.Index(body, "<main")
	mainEnd := strings.Index(body, "</main>")
	if mainStart < 0 || mainEnd < 0 {
		t.Fatal("rendered model-policy page has no <main> region")
	}
	mainRegion := body[mainStart:mainEnd]

	// Exactly one <form> in the model-policy content region — the plan_type selector.
	if n := strings.Count(mainRegion, "<form"); n != 1 {
		t.Errorf("<main> region <form> count = %d, want 1 (only the scoped plan_type selector)", n)
	}
	// That form MUST target the scoped persist endpoint, never the page route.
	if !strings.Contains(mainRegion, `action="/model-policy/plan-type"`) {
		t.Error("the plan_type selector form must POST to /model-policy/plan-type (scoped write path)")
	}
	// The single writable control is the plan_type field — no other field name is
	// enrolled in a write path.
	if !strings.Contains(mainRegion, `name="plan_type"`) {
		t.Error("the scoped selector must expose the plan_type field")
	}
	// No htmx post affordance anywhere in the content region (D2 uses a plain form).
	if strings.Contains(mainRegion, "hx-post") {
		t.Error("model-policy <main> region must not carry an hx-post affordance (D2 scope is a plain form)")
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

// --- SPEC-MODEL-TIER-PLANTYPE-001 M4: plan_type selector + persist endpoint ---

// writeLLMPlanType writes root/.moai/config/sections/llm.yaml carrying an explicit
// plan_type line (the persist regex needs a line to replace) plus a performance_tier
// line so the file has more than one line for byte-identity assertions.
func writeLLMPlanType(t *testing.T, root, planType string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	body := "llm:\n    performance_tier: \"max\"\n    plan_type: " + planType + "\n"
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}
}

// readLLM returns the raw llm.yaml bytes for byte-identity comparisons.
func readLLM(t *testing.T, root string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"))
	if err != nil {
		t.Fatalf("read llm.yaml: %v", err)
	}
	return b
}

// postPlanType issues a loopback-authenticated POST through the real app mux
// (host-check + Sec-Fetch-Site same-origin satisfied so the CSRF/DNS-rebind gate
// passes) to the given target path with the plan_type form value.
func postPlanType(t *testing.T, root, target, value string) *httptest.ResponseRecorder {
	t.Helper()
	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	form := url.Values{"plan_type": {value}}
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Host = "127.0.0.1"
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	return rec
}

// TestModelPolicyActivePlan (AC-MTP-019): the active plan type is displayed from
// llm.plan_type; an absent field renders the default-subscription label.
func TestModelPolicyActivePlan(t *testing.T) {
	rootAPI := t.TempDir()
	writeLLMPlanType(t, rootAPI, "api")
	rec := getModelPolicy(t, rootAPI)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /model-policy (api) = %d, want 200", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, `value="api" selected`) {
		t.Error("plan_type: api did not pre-select the api option in the selector")
	}

	rootAbsent := t.TempDir()
	writeLLMSection(t, rootAbsent, "max") // performance_tier only — no plan_type key
	recAbsent := getModelPolicy(t, rootAbsent)
	if recAbsent.Code != http.StatusOK {
		t.Fatalf("GET /model-policy (absent plan_type) = %d, want 200", recAbsent.Code)
	}
	body := recAbsent.Body.String()
	if !strings.Contains(body, "mp.plan.default") {
		t.Error("absent plan_type did not render the default-subscription label (mp.plan.default)")
	}
	// The effective plan (subscription) is pre-selected even when the field is absent.
	if !strings.Contains(body, `value="subscription" selected`) {
		t.Error("absent plan_type did not pre-select the subscription option (effective default)")
	}
}

// TestModelPolicyDualPlanPreview (AC-MTP-020 + AC-MTP-023 single source): both plan
// previews render all 11 retained agents × 3 tiers, and an asserted cell value is
// READ FROM the Go tier-profile structure — proving the web layer derives the matrix
// rather than duplicating it as a literal.
func TestModelPolicyDualPlanPreview(t *testing.T) {
	root := t.TempDir()
	writeLLMPlanType(t, root, "api")
	body := getModelPolicy(t, root).Body.String()

	if n := strings.Count(body, `class="mp-plan-preview"`); n != 2 {
		t.Errorf("mp-plan-preview block count = %d, want 2 (api + subscription)", n)
	}
	if !strings.Contains(body, `name="plan_type"`) {
		t.Error("plan_type selector markup missing")
	}
	for _, agent := range template.TierProfileAgents() {
		if !strings.Contains(body, ">"+agent+"<") {
			t.Errorf("preview missing agent row %q", agent)
		}
	}
	// Each plan preview must render 11 agents × 3 tier cells → 33 tier <td> per plan.
	// Assert the aggregate row count via the agent-row marker (11 agents × 2 plans).
	if n := strings.Count(body, `class="mp-plan-row"`); n != len(template.TierProfileAgents())*2 {
		t.Errorf("mp-plan-row count = %d, want %d (11 agents × 2 plans)", n, len(template.TierProfileAgents())*2)
	}

	// Structure-derived cross-check (single-source guarantee): api/max/super-advisor
	// is read from the Go profile, never a hardcoded web literal.
	entry, ok := template.GetTierProfileEntry(config.PlanTypeAPI, "super-advisor", template.PerformanceTierMax)
	if !ok {
		t.Fatal("tier profile lookup failed for api/super-advisor/max")
	}
	if entry.Effort != "xhigh" {
		t.Fatalf("test premise broken: api/super-advisor/max effort = %q, want xhigh", entry.Effort)
	}
	if !strings.Contains(body, ">"+entry.Effort+"<") {
		t.Errorf("structure-derived cell effort %q not rendered in the preview", entry.Effort)
	}
	if !strings.Contains(body, ">"+entry.Model+"<") {
		t.Errorf("structure-derived cell model %q not rendered in the preview", entry.Model)
	}
}

// TestModelPolicyPlanTypePersistRoundTrip (AC-MTP-027b): a valid plan-type change
// POSTed to /model-policy/plan-type through the real mux persists to llm.yaml and the
// subsequent GET renders the new active plan.
func TestModelPolicyPlanTypePersistRoundTrip(t *testing.T) {
	root := t.TempDir()
	writeLLMPlanType(t, root, "subscription")

	rec := postPlanType(t, root, "/model-policy/plan-type", "api")
	if rec.Code < 200 || rec.Code >= 400 {
		t.Fatalf("POST /model-policy/plan-type (valid) = %d, want 2xx/3xx", rec.Code)
	}
	if got := string(readLLM(t, root)); !strings.Contains(got, "plan_type: api") {
		t.Errorf("llm.yaml did not persist plan_type: api after POST; got:\n%s", got)
	}
	if body := getModelPolicy(t, root).Body.String(); !strings.Contains(body, `value="api" selected`) {
		t.Error("GET /model-policy after persist did not render api as the active plan")
	}
}

// TestModelPolicyPlanTypeOutOfSet (AC-MTP-021b): an out-of-set POST is rejected 4xx
// and leaves llm.yaml byte-identical.
func TestModelPolicyPlanTypeOutOfSet(t *testing.T) {
	root := t.TempDir()
	writeLLMPlanType(t, root, "subscription")
	before := readLLM(t, root)

	rec := postPlanType(t, root, "/model-policy/plan-type", "enterprise")
	if rec.Code < 400 || rec.Code >= 500 {
		t.Fatalf("POST out-of-set plan_type = %d, want 4xx", rec.Code)
	}
	if after := readLLM(t, root); string(after) != string(before) {
		t.Errorf("out-of-set POST mutated llm.yaml; before=%q after=%q", before, after)
	}
}

// TestModelPolicyPlanTypeOnlyPlanLineChanges (AC-MTP-021c): a successful persist
// changes ONLY the plan_type line — every other line is byte-identical.
func TestModelPolicyPlanTypeOnlyPlanLineChanges(t *testing.T) {
	root := t.TempDir()
	writeLLMPlanType(t, root, "subscription")
	before := strings.Split(string(readLLM(t, root)), "\n")

	rec := postPlanType(t, root, "/model-policy/plan-type", "api")
	if rec.Code < 200 || rec.Code >= 400 {
		t.Fatalf("POST valid plan_type = %d, want 2xx/3xx", rec.Code)
	}
	after := strings.Split(string(readLLM(t, root)), "\n")
	if len(before) != len(after) {
		t.Fatalf("line count changed: before=%d after=%d", len(before), len(after))
	}
	for i := range before {
		if strings.Contains(before[i], "plan_type:") {
			if !strings.Contains(after[i], "plan_type: api") {
				t.Errorf("plan_type line not updated: %q", after[i])
			}
			continue
		}
		if before[i] != after[i] {
			t.Errorf("non-plan_type line %d changed: %q -> %q", i, before[i], after[i])
		}
	}
}

// TestModelPolicyNoOtherWritePath (AC-MTP-021a/d): the page route stays GET-only
// (405 on non-GET) and no other sub-path under /model-policy accepts a write.
func TestModelPolicyNoOtherWritePath(t *testing.T) {
	root := t.TempDir()
	writeLLMPlanType(t, root, "subscription")
	before := readLLM(t, root)

	// (a) POST to the page route itself → 405 (page route is GET-only).
	if rec := postPlanType(t, root, "/model-policy", "api"); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /model-policy = %d, want 405 (page route is GET-only)", rec.Code)
	}
	// (d) POST to an unregistered sub-path → not a 2xx/3xx, llm.yaml unchanged.
	rec := postPlanType(t, root, "/model-policy/bogus", "api")
	if rec.Code >= 200 && rec.Code < 400 {
		t.Errorf("POST /model-policy/bogus = %d, want >= 400 (no write path)", rec.Code)
	}
	if after := readLLM(t, root); string(after) != string(before) {
		t.Error("a non-plan-type POST mutated llm.yaml")
	}
}

// TestModelPolicyPlanI18nKeys (AC-MTP-022): the new plan_type i18n keys exist in all
// 4 locales. Spot-anchor: mp.plan.title appears exactly 4 times (one per locale).
func TestModelPolicyPlanI18nKeys(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	if n := strings.Count(dict, `"mp.plan.title":`); n != 4 {
		t.Errorf(`"mp.plan.title" locale count = %d, want 4`, n)
	}
	body := getModelPolicy(t, t.TempDir()).Body.String()
	keyRe := regexp.MustCompile(`data-i18n="(mp\.plan[^"]*)"`)
	seen := map[string]bool{}
	for _, m := range keyRe.FindAllStringSubmatch(body, -1) {
		if seen[m[1]] {
			continue
		}
		seen[m[1]] = true
		if !i18nKeyInAllLocales(t, m[1]) {
			t.Errorf("i18n.js missing plan key %q in all 4 locales", m[1])
		}
	}
	if len(seen) == 0 {
		t.Fatal("no mp.plan.* data-i18n keys found in the rendered page")
	}
}
