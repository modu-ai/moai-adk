package toolpolicy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPathHelpers covers SettingsPath / TemplateSettingsPath / PolicyYAMLPath.
func TestPathHelpers(t *testing.T) {
	t.Parallel()
	if got := SettingsPath("/proj"); got != filepath.Join("/proj", ".claude", "settings.json") {
		t.Errorf("SettingsPath = %q", got)
	}
	if got := TemplateSettingsPath("/repo"); !strings.HasSuffix(got, filepath.Join("template", "templates", ".claude", "settings.json.tmpl")) {
		t.Errorf("TemplateSettingsPath = %q", got)
	}
	if got := PolicyYAMLPath("/proj"); !strings.HasSuffix(got, filepath.Join(".moai", "config", "sections", "tool-policy.yaml")) {
		t.Errorf("PolicyYAMLPath = %q", got)
	}
}

// TestLoadFromProjectDir covers the convenience loader entry-point.
func TestLoadFromProjectDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	yaml := `metadata: {version: "1.0.0"}
entries:
  - tool: Read
    args_pattern: ""
    risk_tier: read
    decision: allow
    owner_agent: orchestrator
    audit: read
`
	if err := os.WriteFile(filepath.Join(sectionsDir, "tool-policy.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	doc, err := LoadFromProjectDir(dir)
	if err != nil {
		t.Fatalf("LoadFromProjectDir: %v", err)
	}
	if len(doc.Entries) != 1 {
		t.Errorf("entries = %d; want 1", len(doc.Entries))
	}
}

// TestLoadFromProjectDir_MissingFile verifies the loader returns a clear
// error when the file does not exist (EC-5).
func TestLoadFromProjectDir_MissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, err := LoadFromProjectDir(dir)
	if err == nil {
		t.Fatal("LoadFromProjectDir returned nil for missing file; want error")
	}
	if !strings.Contains(err.Error(), "tool-policy read") {
		t.Errorf("error = %q; want substring 'tool-policy read'", err.Error())
	}
}

// TestDecisionFor covers the (specifier → decision) lookup used by the
// round-trip equivalence test.
func TestDecisionFor(t *testing.T) {
	t.Parallel()
	doc := &PolicyDocument{
		Entries: []PolicyEntry{
			{Tool: "Bash", ArgsPattern: "git push:*", RiskTier: RiskTierWrite, Decision: DecisionAllow, OwnerAgent: "o", Audit: "a"},
			{Tool: "WebSearch", ArgsPattern: "*", RiskTier: RiskTierRead, Decision: DecisionDeny, OwnerAgent: "o", Audit: "a", EnvGate: &EnvGate{BaseURLContains: "api.z.ai"}},
		},
	}
	if dec, ok := doc.DecisionFor("Bash(git push:*)"); !ok || dec != DecisionAllow {
		t.Errorf("DecisionFor(Bash(git push:*)) = (%q,%v); want (allow,true)", dec, ok)
	}
	// env-gated entry must NOT be returned (it is not emitted as a static specifier).
	if _, ok := doc.DecisionFor("WebSearch"); ok {
		t.Error("DecisionFor(WebSearch) returned ok for env-gated entry; want false")
	}
	if _, ok := doc.DecisionFor("Nonexistent"); ok {
		t.Error("DecisionFor(Nonexistent) returned ok; want false")
	}
}

// TestCountSpecifierByDecision covers the tally helper.
func TestCountSpecifierByDecision(t *testing.T) {
	t.Parallel()
	doc := &PolicyDocument{
		Entries: []PolicyEntry{
			{Tool: "A", RiskTier: RiskTierRead, Decision: DecisionAllow, OwnerAgent: "o", Audit: "a"},
			{Tool: "B", RiskTier: RiskTierRead, Decision: DecisionAllow, OwnerAgent: "o", Audit: "a"},
			{Tool: "C", RiskTier: RiskTierRead, Decision: DecisionDeny, OwnerAgent: "o", Audit: "a"},
			{Tool: "D", RiskTier: RiskTierRead, Decision: DecisionDeny, OwnerAgent: "o", Audit: "a", EnvGate: &EnvGate{BaseURLContains: "x"}},
		},
	}
	got := doc.CountSpecifierByDecision()
	if got[DecisionAllow] != 2 {
		t.Errorf("allow count = %d; want 2", got[DecisionAllow])
	}
	if got[DecisionDeny] != 1 { // env-gated entry is excluded
		t.Errorf("deny count = %d; want 1 (env-gated excluded)", got[DecisionDeny])
	}
	if got[DecisionAsk] != 0 {
		t.Errorf("ask count = %d; want 0", got[DecisionAsk])
	}
}

// TestEnsureNoCR covers the CRLF normalization helper.
func TestEnsureNoCR(t *testing.T) {
	t.Parallel()
	in := []byte("line1\r\nline2\r\n")
	out := ensureNoCR(in)
	if strings.Contains(string(out), "\r") {
		t.Errorf("ensureNoCR left CR: %q", out)
	}
	if string(out) != "line1\nline2\n" {
		t.Errorf("ensureNoCR = %q; want %q", out, "line1\nline2\n")
	}
}

// TestBuildPermissions_DuplicateSpecifiers verifies that duplicate specifiers
// within the same decision list are deduplicated (AC-TPS-013 idempotency
// sub-property — the output must not depend on how many times an entry
// appears in the YAML).
func TestBuildPermissions_DuplicateSpecifiers(t *testing.T) {
	t.Parallel()
	doc := &PolicyDocument{
		Entries: []PolicyEntry{
			{Tool: "Read", RiskTier: RiskTierRead, Decision: DecisionAllow, OwnerAgent: "o", Audit: "a"},
			{Tool: "Read", ArgsPattern: "*", RiskTier: RiskTierRead, Decision: DecisionAllow, OwnerAgent: "o", Audit: "a"}, // same specifier ("" and "*" both render as "Read")
			{Tool: "Read", RiskTier: RiskTierRead, Decision: DecisionAllow, OwnerAgent: "o", Audit: "a"},
		},
	}
	block, res, err := BuildPermissions(doc, "")
	if err != nil {
		t.Fatalf("BuildPermissions: %v", err)
	}
	if len(block.Allow) != 1 {
		t.Errorf("dedup failed: allow = %v; want exactly [Read]", block.Allow)
	}
	if res.AllowEmitted != 1 {
		t.Errorf("AllowEmitted = %d; want 1 (deduped)", res.AllowEmitted)
	}
}

// TestBuildPermissions_UnknownDecision covers the default branch error path.
func TestBuildPermissions_UnknownDecision(t *testing.T) {
	t.Parallel()
	doc := &PolicyDocument{
		Entries: []PolicyEntry{
			{Tool: "X", RiskTier: RiskTierRead, Decision: Decision("bogus"), OwnerAgent: "o", Audit: "a"},
		},
	}
	// Bypass Validate (which would reject the bogus decision) by calling
	// BuildPermissions directly with an already-constructed (invalid) document.
	_, _, err := BuildPermissions(doc, "")
	if err == nil {
		t.Fatal("BuildPermissions returned nil for unknown decision; want error")
	}
	if !strings.Contains(err.Error(), "unknown decision") {
		t.Errorf("error = %q; want substring 'unknown decision'", err.Error())
	}
}

// TestBuildInto_DefaultModeOverride covers the --default-mode flag path.
func TestBuildInto_DefaultModeOverride(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(`{"permissions":{"defaultMode":"default","allow":["Read"]}}`), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	doc := &PolicyDocument{
		Entries: []PolicyEntry{
			{Tool: "Read", RiskTier: RiskTierRead, Decision: DecisionAllow, OwnerAgent: "o", Audit: "a"},
		},
	}
	if _, err := BuildInto(path, doc, TargetJSON, "bypassPermissions"); err != nil {
		t.Fatalf("BuildInto: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(body), `"defaultMode": "bypassPermissions"`) {
		t.Errorf("defaultMode override not applied; body = %s", body)
	}
}

// TestBuildInto_UnknownTargetKind covers the default branch error path.
func TestBuildInto_UnknownTargetKind(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(`{"permissions":{"defaultMode":"default","allow":["Read"]}}`), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	doc := &PolicyDocument{
		Entries: []PolicyEntry{
			{Tool: "Read", RiskTier: RiskTierRead, Decision: DecisionAllow, OwnerAgent: "o", Audit: "a"},
		},
	}
	_, err := BuildInto(path, doc, TargetKind("bogus"), "")
	if err == nil {
		t.Fatal("BuildInto returned nil for unknown target kind; want error")
	}
	if !strings.Contains(err.Error(), "unknown target kind") {
		t.Errorf("error = %q; want substring 'unknown target kind'", err.Error())
	}
}

// TestBuildInto_ReadError covers the missing-file error path.
func TestBuildInto_ReadError(t *testing.T) {
	t.Parallel()
	doc := &PolicyDocument{Entries: []PolicyEntry{}}
	_, err := BuildInto("/nonexistent/path/settings.json", doc, TargetJSON, "")
	if err == nil {
		t.Fatal("BuildInto returned nil for missing file; want error")
	}
}

// TestRenderSettingsJSON_NoPermissions covers the error path when the body
// has no permissions block.
func TestRenderSettingsJSON_NoPermissions(t *testing.T) {
	t.Parallel()
	body := []byte(`{"env": {}}`)
	_, err := RenderSettingsJSON(body, &PermissionsBlock{Raw: map[string]json.RawMessage{}})
	if err == nil {
		t.Fatal("RenderSettingsJSON returned nil for missing permissions block; want error")
	}
}

// TestAssertPermissionBlockNoTemplate_Failure covers the sentinel assertion
// failure path (template directive leaked into the block).
func TestAssertPermissionBlockNoTemplate_Failure(t *testing.T) {
	t.Parallel()
	bad := []byte(`{"allow": ["{{evil}}"]}`)
	err := AssertPermissionBlockNoTemplate(bad)
	if err == nil {
		t.Fatal("AssertPermissionBlockNoTemplate returned nil for {{ leakage; want error")
	}
}

// TestSortedBySpecifier_IdempotentOrder verifies the sort is stable and
// produces a deterministic order regardless of input order.
func TestSortedBySpecifier_IdempotentOrder(t *testing.T) {
	t.Parallel()
	doc := &PolicyDocument{
		Entries: []PolicyEntry{
			{Tool: "Zed", RiskTier: RiskTierRead, Decision: DecisionAllow, OwnerAgent: "o", Audit: "a"},
			{Tool: "Alpha", RiskTier: RiskTierRead, Decision: DecisionAllow, OwnerAgent: "o", Audit: "a"},
			{Tool: "Bash", ArgsPattern: "git:*", RiskTier: RiskTierRead, Decision: DecisionAllow, OwnerAgent: "o", Audit: "a"},
		},
	}
	got := doc.SortedBySpecifier()
	want := []string{"Alpha", "Bash(git:*)", "Zed"}
	for i, w := range want {
		if i >= len(got) {
			t.Errorf("sorted output shorter than expected: %d entries", len(got))
			break
		}
		if got[i].SettingsSpecifier() != w {
			t.Errorf("sorted[%d] = %q; want %q", i, got[i].SettingsSpecifier(), w)
		}
	}
}
