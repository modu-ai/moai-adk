package web

// TDD (goal-to-test, non-SPEC) — wire template.ApplyTierProfile into the web
// /model-policy surface so that changing plan_type OR tier re-applies the model
// + effort tier profile to the shipped agent definition files, and make the
// /model-policy page reachable from the console appbar. These tests define the
// desired behavior BEFORE the wiring exists (RED); the GREEN pass adds the
// handler logic, the tier selector, and the nav link.

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
)

// writeAgentMD writes a shipped-agent frontmatter fixture at
// {root}/.claude/agents/moai/<name>.md carrying a deliberately "wrong"
// {model, effort} pair so a subsequent ApplyTierProfile pass has something to
// rewrite. The pair is intentionally off-profile for every {plan_type, tier}
// combination so the apply is observable regardless of resolution.
func writeAgentMD(t *testing.T, root, name, model, effort string) {
	t.Helper()
	dir := filepath.Join(root, ".claude", "agents", "moai")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir agents/moai: %v", err)
	}
	body := "---\n" +
		"name: " + name + "\n" +
		"model: " + model + "\n" +
		"effort: " + effort + "\n" +
		"tools: Read\n" +
		"---\n" +
		"agent body\n"
	if err := os.WriteFile(filepath.Join(dir, name+".md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write agent %s: %v", name, err)
	}
}

// agentFrontmatterField reads {root}/.claude/agents/moai/<name>.md and extracts
// the value of a single frontmatter field (model or effort) via a line-prefix
// scan. Returns "" when the field is absent.
func agentFrontmatterField(t *testing.T, root, name, field string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".claude", "agents", "moai", name+".md"))
	if err != nil {
		t.Fatalf("read agent %s: %v", name, err)
	}
	prefix := field + ": "
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

// ensureMoaiDir creates {root}/.moai so manifest.Manager.Load has a clean parent
// (Load tolerates an absent manifest.json but a present .moai dir mirrors a real
// initialized project).
func ensureMoaiDir(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
}

// postModelPolicy issues a POST to the given /model-policy/* path against a fresh
// app rooted at root, returning the response recorder.
func postModelPolicy(t *testing.T, root, path string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "same-origin") // satisfy CSRF/DNS-rebind gate
	req.Host = "127.0.0.1"                          // satisfy hostCheckMiddleware loopback-Host gate
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	return rec
}

// TestModelPolicyPlanType_AppliesAgentProfile (RED): posting a valid plan_type to
// /model-policy/plan-type MUST re-apply the tier profile to the shipped agent
// definition files — the agent .md frontmatter must change to the profile entry
// for the (new plan_type × current tier) combination, not merely persist
// llm.plan_type.
func TestModelPolicyPlanType_AppliesAgentProfile(t *testing.T) {
	root := t.TempDir()
	ensureMoaiDir(t, root)
	// Current tier on disk = low; we switch plan_type api -> subscription.
	writeLLMSectionFull(t, root, "api", "low")
	writeAgentMD(t, root, "manager-spec", "haiku", "low")

	// Expected post-apply values: subscription × low for manager-spec.
	want, ok := template.GetTierProfileEntry("subscription", "manager-spec", "low")
	if !ok {
		t.Fatalf("manager-spec has no subscription/low profile entry — fixture is wrong")
	}
	if want.Model == "haiku" && want.Effort == "low" {
		t.Fatalf("subscription/low/manager-spec profile == fixture pair (haiku/low); pick a distinguishing fixture")
	}

	rec := postModelPolicy(t, root, "/model-policy/plan-type", url.Values{
		"plan_type": {"subscription"},
	})
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST /model-policy/plan-type = %d, want 303; body: %s", rec.Code, rec.Body.String())
	}

	gotModel := agentFrontmatterField(t, root, "manager-spec", "model")
	gotEffort := agentFrontmatterField(t, root, "manager-spec", "effort")
	if gotModel != want.Model {
		t.Errorf("agent model after plan_type switch = %q, want %q (ApplyTierProfile did not re-apply)", gotModel, want.Model)
	}
	if gotEffort != want.Effort {
		t.Errorf("agent effort after plan_type switch = %q, want %q (ApplyTierProfile did not re-apply)", gotEffort, want.Effort)
	}
}

// TestModelPolicyTier_AppliesAgentProfile (RED): posting a valid tier to
// /model-policy/tier MUST persist performance_tier AND re-apply the tier profile
// to the shipped agent definition files.
func TestModelPolicyTier_AppliesAgentProfile(t *testing.T) {
	root := t.TempDir()
	ensureMoaiDir(t, root)
	writeLLMSectionFull(t, root, "subscription", "low")
	writeAgentMD(t, root, "manager-spec", "haiku", "low")

	want, ok := template.GetTierProfileEntry("subscription", "manager-spec", "max")
	if !ok {
		t.Fatalf("manager-spec has no subscription/max profile entry — fixture is wrong")
	}

	rec := postModelPolicy(t, root, "/model-policy/tier", url.Values{
		"performance_tier": {"max"},
	})
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST /model-policy/tier = %d, want 303; body: %s", rec.Code, rec.Body.String())
	}

	gotModel := agentFrontmatterField(t, root, "manager-spec", "model")
	gotEffort := agentFrontmatterField(t, root, "manager-spec", "effort")
	if gotModel != want.Model {
		t.Errorf("agent model after tier switch = %q, want %q (ApplyTierProfile did not re-apply)", gotModel, want.Model)
	}
	if gotEffort != want.Effort {
		t.Errorf("agent effort after tier switch = %q, want %q (ApplyTierProfile did not re-apply)", gotEffort, want.Effort)
	}
}

// TestModelPolicyTier_RejectsInvalid (RED): an out-of-set tier (e.g. "high" —
// the legacy wizard vocabulary, NOT the canonical {max, medium, low}) MUST be
// rejected with 4xx and MUST NOT mutate the agent files.
func TestModelPolicyTier_RejectsInvalid(t *testing.T) {
	root := t.TempDir()
	ensureMoaiDir(t, root)
	writeLLMSectionFull(t, root, "subscription", "low")
	writeAgentMD(t, root, "manager-spec", "haiku", "low")

	rec := postModelPolicy(t, root, "/model-policy/tier", url.Values{
		"performance_tier": {"high"},
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /model-policy/tier high = %d, want 400; body: %s", rec.Code, rec.Body.String())
	}
	if got := agentFrontmatterField(t, root, "manager-spec", "model"); got != "haiku" {
		t.Errorf("agent model mutated by rejected tier post: %q (want haiku unchanged)", got)
	}
}

// (The /model-policy/tier route registration is covered by
// TestModelPolicyTier_AppliesAgentProfile — a missing route surfaces there as a
// 404 once the GREEN middleware headers are satisfied.)

// TestSettingsPage_HasModelPolicyNav (RED): the main settings console (GET /)
// MUST surface a navigation link to /model-policy so the Model Policy page is
// reachable from the console (it was previously an orphaned route).
func TestSettingsPage_HasModelPolicyNav(t *testing.T) {
	root := t.TempDir()
	ensureMoaiDir(t, root)
	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `href="/model-policy"`) {
		t.Error("settings page missing /model-policy nav link")
	}
}

// writeLLMSectionFull writes {root}/.moai/config/sections/llm.yaml with BOTH a
// plan_type and a performance_tier value (the persisted resolution inputs for
// ResolveProjectPlanType / ResolveProjectPerformanceTier).
func writeLLMSectionFull(t *testing.T, root, planType, perfTier string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	body := "llm:\n" +
		"    plan_type: \"" + planType + "\"\n" +
		"    performance_tier: \"" + perfTier + "\"\n"
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}
}
