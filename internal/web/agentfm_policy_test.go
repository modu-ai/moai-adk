package web

// Tests for the plan_type / performance_tier selectors migrated into the TOP
// of the Sub-agent Frontmatter (agentfm) panel from the retired standalone
// Model Policy page (goal-to-test, non-SPEC). The standalone page + its two
// dedicated POST routes are gone; the two write affordances now live inside
// the single /save form, alongside the per-agent frontmatter edits.

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/template"
)

// writeLLMYAML writes root/.moai/config/sections/llm.yaml with the given
// plan_type + performance_tier values.
func writeLLMYAML(t *testing.T, root, planType, perfTier string) {
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

// readLLMYAML returns the raw llm.yaml bytes for byte-identity comparisons.
func readLLMYAML(t *testing.T, root string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"))
	if err != nil {
		t.Fatalf("read llm.yaml: %v", err)
	}
	return b
}

// agentFrontmatterValue reads {root}/.claude/agents/moai/<name>.md and returns
// the value of the given frontmatter field ("model" or "effort"); "" if absent.
func agentFrontmatterValue(t *testing.T, root, name, field string) string {
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

// newPolicyTestApp builds an app rooted at root with the profile read/write
// seams faked to no-ops, so the profile store itself is never touched — these
// tests exercise only the plan_type/performance_tier + agent-frontmatter write
// paths.
func newPolicyTestApp(root string) *app {
	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	a.readPreferences = func(string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{}, nil
	}
	a.writePreferences = func(string, profile.ProfilePreferences) error { return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	return a
}

// baseSaveForm returns the minimal valid /save form payload — permission_mode
// empty is a valid "(project default)" value. Tests add plan_type /
// performance_tier on top.
func baseSaveForm() url.Values {
	return url.Values{"__profile": {"default"}}
}

// TestAgentFMPolicy_PlanTypePersistsAndReapplies (b): a valid plan_type change
// submitted through /save persists llm.plan_type and re-applies the tier
// profile to the shipped agent frontmatter.
func TestAgentFMPolicy_PlanTypePersistsAndReapplies(t *testing.T) {
	root := t.TempDir()
	writeLLMYAML(t, root, "api", "low")
	seedAgentFMFile(t, root, "moai", "manager-spec", "haiku", "low")

	want, ok := template.GetTierProfileEntry("subscription", "manager-spec", "low")
	if !ok {
		t.Fatal("manager-spec has no subscription/low profile entry — fixture is wrong")
	}
	if want.Model == "haiku" && want.Effort == "low" {
		t.Fatal("subscription/low/manager-spec profile == fixture pair (haiku/low); pick a distinguishing fixture")
	}

	a := newPolicyTestApp(root)
	form := baseSaveForm()
	form.Set("plan_type", "subscription")
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /save (plan_type change) = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	if got := string(readLLMYAML(t, root)); !strings.Contains(got, "plan_type: subscription") {
		t.Errorf("llm.yaml did not persist plan_type: subscription; got:\n%s", got)
	}
	if gotModel := agentFrontmatterValue(t, root, "manager-spec", "model"); gotModel != want.Model {
		t.Errorf("agent model after plan_type switch = %q, want %q (tier profile not re-applied)", gotModel, want.Model)
	}
	if gotEffort := agentFrontmatterValue(t, root, "manager-spec", "effort"); gotEffort != want.Effort {
		t.Errorf("agent effort after plan_type switch = %q, want %q (tier profile not re-applied)", gotEffort, want.Effort)
	}
}

// TestAgentFMPolicy_PerfTierPersistsAndReapplies (c): a valid performance_tier
// change submitted through /save persists llm.performance_tier and re-applies
// the tier profile to the shipped agent frontmatter.
func TestAgentFMPolicy_PerfTierPersistsAndReapplies(t *testing.T) {
	root := t.TempDir()
	writeLLMYAML(t, root, "subscription", "low")
	seedAgentFMFile(t, root, "moai", "manager-spec", "haiku", "low")

	want, ok := template.GetTierProfileEntry("subscription", "manager-spec", "max")
	if !ok {
		t.Fatal("manager-spec has no subscription/max profile entry — fixture is wrong")
	}

	a := newPolicyTestApp(root)
	form := baseSaveForm()
	form.Set("performance_tier", "max")
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /save (performance_tier change) = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	if got := string(readLLMYAML(t, root)); !strings.Contains(got, "performance_tier: max") {
		t.Errorf("llm.yaml did not persist performance_tier: max; got:\n%s", got)
	}
	if gotModel := agentFrontmatterValue(t, root, "manager-spec", "model"); gotModel != want.Model {
		t.Errorf("agent model after tier switch = %q, want %q (tier profile not re-applied)", gotModel, want.Model)
	}
	if gotEffort := agentFrontmatterValue(t, root, "manager-spec", "effort"); gotEffort != want.Effort {
		t.Errorf("agent effort after tier switch = %q, want %q (tier profile not re-applied)", gotEffort, want.Effort)
	}
}

// TestAgentFMPolicy_InvalidValuesRejected (d): an out-of-set plan_type or
// performance_tier (including the legacy wizard "high" — NOT the canonical
// {max, medium, low}) is rejected 4xx and mutates neither llm.yaml nor the
// agent frontmatter.
func TestAgentFMPolicy_InvalidValuesRejected(t *testing.T) {
	root := t.TempDir()
	writeLLMYAML(t, root, "subscription", "low")
	before := readLLMYAML(t, root)
	seedAgentFMFile(t, root, "moai", "manager-spec", "haiku", "low")

	a := newPolicyTestApp(root)

	cases := []struct {
		name  string
		field string
		value string
	}{
		{"plan_type out-of-set", "plan_type", "enterprise"},
		{"performance_tier out-of-set (legacy wizard vocabulary)", "performance_tier", "high"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			form := baseSaveForm()
			form.Set(tc.field, tc.value)
			rec := servePost(t, a.routes(), "/save", form)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("POST /save (%s=%s) = %d, want 400", tc.field, tc.value, rec.Code)
			}
			if after := readLLMYAML(t, root); string(after) != string(before) {
				t.Errorf("llm.yaml mutated by a rejected %s post", tc.field)
			}
			if gotModel := agentFrontmatterValue(t, root, "manager-spec", "model"); gotModel != "haiku" {
				t.Errorf("agent model mutated by rejected post: %q (want haiku unchanged)", gotModel)
			}
		})
	}
}

// TestAgentFMPolicy_EmptySubmissionPreservesCurrent (e): submitting /save
// without plan_type/performance_tier at all preserves both the persisted
// llm.yaml values and the shipped agent frontmatter untouched.
func TestAgentFMPolicy_EmptySubmissionPreservesCurrent(t *testing.T) {
	root := t.TempDir()
	writeLLMYAML(t, root, "api", "max")
	before := readLLMYAML(t, root)
	seedAgentFMFile(t, root, "moai", "manager-spec", "opus", "xhigh")
	beforeModel := agentFrontmatterValue(t, root, "manager-spec", "model")
	beforeEffort := agentFrontmatterValue(t, root, "manager-spec", "effort")

	a := newPolicyTestApp(root)
	rec := servePost(t, a.routes(), "/save", baseSaveForm())
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /save (no plan_type/performance_tier) = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	if after := readLLMYAML(t, root); string(after) != string(before) {
		t.Errorf("llm.yaml changed on an empty plan_type/performance_tier submission; before=%q after=%q", before, after)
	}
	if gotModel := agentFrontmatterValue(t, root, "manager-spec", "model"); gotModel != beforeModel {
		t.Errorf("agent model changed on empty submission: %q -> %q", beforeModel, gotModel)
	}
	if gotEffort := agentFrontmatterValue(t, root, "manager-spec", "effort"); gotEffort != beforeEffort {
		t.Errorf("agent effort changed on empty submission: %q -> %q", beforeEffort, gotEffort)
	}
}

// TestAgentFMPolicy_SelectorsRenderAtTopOfPanel verifies the plan_type /
// performance_tier selectors render inside the agentfm section, BEFORE the
// sub-tabs, and pre-select the active values.
func TestAgentFMPolicy_SelectorsRenderAtTopOfPanel(t *testing.T) {
	root := t.TempDir()
	writeLLMYAML(t, root, "api", "max")
	body := renderAgentFMBody(t, root)

	if !strings.Contains(body, `name="plan_type"`) {
		t.Error("agentfm panel missing the plan_type selector")
	}
	if !strings.Contains(body, `name="performance_tier"`) {
		t.Error("agentfm panel missing the performance_tier selector")
	}
	if !strings.Contains(body, `value="api" selected`) {
		t.Error("plan_type selector did not pre-select the active plan (api)")
	}
	if !strings.Contains(body, `value="max" selected`) {
		t.Error("performance_tier selector did not pre-select the active tier (max)")
	}

	secIdx := strings.Index(body, `data-i18n="sec.agentfm.title"`)
	planIdx := strings.Index(body, `name="plan_type"`)
	subtabIdx := strings.Index(body, `data-agentfm-tab="subagents"`)
	if secIdx < 0 || planIdx < 0 || subtabIdx < 0 {
		t.Fatal("could not locate agentfm section markers")
	}
	if secIdx >= planIdx || planIdx >= subtabIdx {
		t.Error("plan_type selector is not positioned at the TOP of the agentfm panel (before the sub-tabs)")
	}
}

// TestAgentFMPolicy_EmptyValueHints verifies an absent llm.yaml renders the
// plan_type "(default: subscription)" hint and pre-selects the effective
// default. plan_type has no config-layer default (config.NewConfigManager's
// Defaults() only seeds performance_tier=medium — see
// internal/config/defaults.go), so an absent llm.yaml genuinely leaves
// plan_type empty; performance_tier is covered separately below because an
// absent file already resolves to the non-empty "medium" default upstream.
func TestAgentFMPolicy_EmptyValueHints(t *testing.T) {
	root := t.TempDir() // no llm.yaml at all → plan_type empty
	body := renderAgentFMBody(t, root)
	if !strings.Contains(body, "agentfm.plantype.default") {
		t.Error("empty plan_type did not render the default hint (agentfm.plantype.default)")
	}
	if !strings.Contains(body, `value="subscription" selected`) {
		t.Error("absent plan_type did not pre-select the effective default (subscription)")
	}
	if !strings.Contains(body, `value="medium" selected`) {
		t.Error("absent performance_tier did not pre-select the medium default")
	}
}

// TestAgentFMPolicy_ExplicitEmptyTierHint verifies an EXPLICIT
// `performance_tier: ""` in llm.yaml (distinct from an absent file, which
// resolves to the non-empty "medium" config default) renders the
// "(runtime default: medium)" hint.
func TestAgentFMPolicy_ExplicitEmptyTierHint(t *testing.T) {
	root := t.TempDir()
	writeLLMYAML(t, root, "api", "") // performance_tier: "" explicitly
	body := renderAgentFMBody(t, root)
	if !strings.Contains(body, "agentfm.tier.default") {
		t.Error("explicit empty performance_tier did not render the default hint (agentfm.tier.default)")
	}
	if !strings.Contains(body, `value="medium" selected`) {
		t.Error("explicit empty performance_tier did not pre-select the medium default")
	}
}

// TestAgentFMPolicy_I18nKeysInAllLocales verifies the new plan_type /
// performance_tier i18n keys exist in all 4 locales.
func TestAgentFMPolicy_I18nKeysInAllLocales(t *testing.T) {
	for _, key := range []string{
		"agentfm.plantype.title", "agentfm.plantype.desc", "agentfm.plantype.default",
		"agentfm.tier.title", "agentfm.tier.desc", "agentfm.tier.default",
	} {
		if !i18nKeyInAllLocales(t, key) {
			t.Errorf("i18n.js missing agentfm policy key %q in all 4 locales", key)
		}
	}
}
