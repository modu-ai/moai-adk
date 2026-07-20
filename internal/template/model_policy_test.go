package template

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

func TestValidModelPolicies(t *testing.T) {
	policies := ValidModelPolicies()
	if len(policies) == 0 {
		t.Fatal("ValidModelPolicies() returned empty list")
	}
	if len(policies) != 3 {
		t.Errorf("ValidModelPolicies() returned %d items, want 3", len(policies))
	}

	expected := map[string]bool{"high": true, "medium": true, "low": true}
	for _, p := range policies {
		if !expected[p] {
			t.Errorf("unexpected policy: %q", p)
		}
	}
}

func TestIsValidModelPolicy(t *testing.T) {
	tests := []struct {
		policy string
		valid  bool
	}{
		{"high", true},
		{"medium", true},
		{"low", true},
		{"", false},
		{"ultra", false},
		{"HIGH", false},
		{"Medium", false},
		{"none", false},
	}

	for _, tt := range tests {
		t.Run(tt.policy, func(t *testing.T) {
			got := IsValidModelPolicy(tt.policy)
			if got != tt.valid {
				t.Errorf("IsValidModelPolicy(%q) = %v, want %v", tt.policy, got, tt.valid)
			}
		})
	}
}

// TestModelClaudeOpus48Constant verifies the claude-opus-4-8 model ID constant.
func TestModelClaudeOpus48Constant(t *testing.T) {
	if ModelIDOpus48 == "" {
		t.Error("ModelIDOpus48 constant is empty, want non-empty model ID")
	}
	want := "claude-opus-4-8"
	if ModelIDOpus48 != want {
		t.Errorf("ModelIDOpus48 = %q, want %q", ModelIDOpus48, want)
	}
}

// TestEffortLevelConstants verifies xhigh and max constants exist.
func TestEffortLevelConstants(t *testing.T) {
	if EffortLevelXHigh == "" {
		t.Error("EffortLevelXHigh constant is empty")
	}
	if EffortLevelMax == "" {
		t.Error("EffortLevelMax constant is empty")
	}
	if EffortLevelXHigh != "xhigh" {
		t.Errorf("EffortLevelXHigh = %q, want %q", EffortLevelXHigh, "xhigh")
	}
	if EffortLevelMax != "max" {
		t.Errorf("EffortLevelMax = %q, want %q", EffortLevelMax, "max")
	}
}

func TestNewDeployerWithRenderer(t *testing.T) {
	fsys := testFS()
	r := NewRenderer(fsys)
	d := NewDeployerWithRenderer(fsys, r)
	if d == nil {
		t.Fatal("NewDeployerWithRenderer returned nil")
	}
	// Verify it functions by listing templates
	list := d.ListTemplates()
	if len(list) == 0 {
		t.Error("ListTemplates() returned empty list from DeployerWithRenderer")
	}
}

func TestNewDeployerWithForceUpdate(t *testing.T) {
	fsys := testFS()
	d := NewDeployerWithForceUpdate(fsys, true)
	if d == nil {
		t.Fatal("NewDeployerWithForceUpdate returned nil")
	}
	list := d.ListTemplates()
	if len(list) == 0 {
		t.Error("ListTemplates() returned empty list from DeployerWithForceUpdate")
	}
}

func TestNewDeployerWithRendererAndForceUpdate(t *testing.T) {
	fsys := testFS()
	r := NewRenderer(fsys)
	d := NewDeployerWithRendererAndForceUpdate(fsys, r, true)
	if d == nil {
		t.Fatal("NewDeployerWithRendererAndForceUpdate returned nil")
	}
	list := d.ListTemplates()
	if len(list) == 0 {
		t.Error("ListTemplates() returned empty list from DeployerWithRendererAndForceUpdate")
	}
}

func TestDeployWithForceUpdate(t *testing.T) {
	root, mgr := setupDeployProject(t)
	fsys := testFS()

	// First deploy normally
	d := NewDeployer(fsys)
	if err := d.Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("initial Deploy error: %v", err)
	}

	// Modify a deployed file to simulate user changes
	claudeMDPath := filepath.Join(root, "CLAUDE.md")
	if err := os.WriteFile(claudeMDPath, []byte("user modified content"), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	// Deploy with forceUpdate=true should overwrite
	fd := NewDeployerWithForceUpdate(fsys, true)
	if err := fd.Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("force Deploy error: %v", err)
	}

	content, err := os.ReadFile(claudeMDPath)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if string(content) == "user modified content" {
		t.Error("forceUpdate did not overwrite user-modified file")
	}
}

func TestDeployWithTemplateRendering(t *testing.T) {
	tmplFS := fstest.MapFS{
		"config.yaml.tmpl": &fstest.MapFile{
			Data: []byte("project: {{.ProjectName}}\nversion: {{.Version}}\n"),
		},
	}

	root, mgr := setupDeployProject(t)
	r := NewRenderer(tmplFS)
	d := NewDeployerWithRenderer(tmplFS, r)

	ctx := NewTemplateContext(
		WithProject("test-project", root),
		WithVersion("1.0.0"),
	)

	if err := d.Deploy(context.Background(), root, mgr, ctx); err != nil {
		t.Fatalf("Deploy error: %v", err)
	}

	// Verify the rendered file (without .tmpl suffix)
	content, err := os.ReadFile(filepath.Join(root, "config.yaml"))
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if !containsString(string(content), "project: test-project") {
		t.Errorf("rendered content missing project name: %s", content)
	}
	if !containsString(string(content), "version: 1.0.0") {
		t.Errorf("rendered content missing version: %s", content)
	}
}

func TestDeployShellScriptPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix file permissions not supported on Windows")
	}

	fsys := fstest.MapFS{
		"scripts/run.sh": &fstest.MapFile{
			Data: []byte("#!/bin/bash\necho hello"),
		},
		"docs/readme.md": &fstest.MapFile{
			Data: []byte("# Readme"),
		},
	}

	root, mgr := setupDeployProject(t)
	d := NewDeployer(fsys)

	if err := d.Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("Deploy error: %v", err)
	}

	// Shell scripts should have executable permissions
	info, err := os.Stat(filepath.Join(root, "scripts", "run.sh"))
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}
	perm := info.Mode().Perm()
	if perm&0o100 == 0 {
		t.Errorf("shell script should be executable, got permissions: %o", perm)
	}

	// Non-shell files should NOT be executable
	info2, err := os.Stat(filepath.Join(root, "docs", "readme.md"))
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}
	perm2 := info2.Mode().Perm()
	if perm2&0o100 != 0 {
		t.Errorf("non-shell file should not be executable, got permissions: %o", perm2)
	}
}

func TestDeployExistingUserFile(t *testing.T) {
	fsys := testFS()
	root, mgr := setupDeployProject(t)

	// Pre-create a file that is NOT tracked in manifest
	claudeDir := filepath.Join(root, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}
	settingsPath := filepath.Join(root, ".claude", "settings.json")
	if err := os.WriteFile(settingsPath, []byte(`{"user": true}`), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	d := NewDeployer(fsys)
	if err := d.Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("Deploy error: %v", err)
	}

	// The pre-existing file should be preserved (not overwritten)
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if string(content) != `{"user": true}` {
		t.Errorf("existing user file was overwritten: %s", content)
	}

	// It should be tracked as user_created in manifest
	entry, found := mgr.GetEntry(".claude/settings.json")
	if !found {
		t.Error("expected manifest entry for user file")
	} else if entry.Provenance != manifest.UserCreated {
		t.Errorf("provenance = %v, want UserCreated", entry.Provenance)
	}
}

// TestMapModelPolicyToTier (REQ-MPM-002 alias) asserts the legacy ModelPolicy→tier
// mapping is tier-only (high→max, medium→medium, low→low) — the read-time alias
// source for legacy performance_tier: high → profile: max.
func TestMapModelPolicyToTier(t *testing.T) {
	tests := []struct {
		policy ModelPolicy
		want   string
	}{
		{ModelPolicyHigh, "max"},
		{ModelPolicyMedium, "medium"},
		{ModelPolicyLow, "low"},
		{ModelPolicy(""), "medium"},      // empty → default-when-absent
		{ModelPolicy("bogus"), "medium"}, // unknown → default-when-absent
	}
	for _, tt := range tests {
		if got := MapModelPolicyToTier(tt.policy); got != tt.want {
			t.Errorf("MapModelPolicyToTier(%q) = %q, want %q", tt.policy, got, tt.want)
		}
	}
}

// TestMapModelPolicyToEffort asserts the runtime-LAUNCH effort projection of the
// legacy {high,medium,low} ModelPolicy vocabulary: high→high, medium→medium,
// low→low on the EFFORT axis. Empty/unknown → "" (no override).
func TestMapModelPolicyToEffort(t *testing.T) {
	tests := []struct {
		policy ModelPolicy
		want   string
	}{
		{ModelPolicyHigh, "high"},
		{ModelPolicyMedium, "medium"},
		{ModelPolicyLow, "low"},
		{ModelPolicy(""), ""},    // empty → no override (byte-identical to today)
		{ModelPolicy("xyz"), ""}, // unknown → no override
	}
	for _, tt := range tests {
		if got := MapModelPolicyToEffort(tt.policy); got != tt.want {
			t.Errorf("MapModelPolicyToEffort(%q) = %q, want %q", tt.policy, got, tt.want)
		}
	}
}

// TestNormalizeToTier asserts the call-site resolver accepts BOTH the canonical
// performance-tier vocabulary ({max, medium, low}) and the legacy ModelPolicy
// vocabulary ({high, medium, low}), defaulting to medium.
func TestNormalizeToTier(t *testing.T) {
	cases := map[string]string{
		"max": "max", "medium": "medium", "low": "low",
		"high": "max", "": "medium", "bogus": "medium",
	}
	for in, want := range cases {
		if got := NormalizeToTier(in); got != want {
			t.Errorf("NormalizeToTier(%q) = %q, want %q", in, got, want)
		}
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// writeLLMSection writes a minimal llm.yaml under root's sections dir with the
// given body (already indented under `llm:`).
func writeLLMSection(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile llm.yaml: %v", err)
	}
}

// TestResolveProjectPerformanceTier — the legacy alias axis reads the persisted
// performance_tier from llm.yaml; absent/empty → medium (default).
func TestResolveProjectPerformanceTier(t *testing.T) {
	root := t.TempDir()
	if got := ResolveProjectPerformanceTier(root); got != "medium" {
		t.Errorf("absent llm.yaml: got %q, want medium", got)
	}

	writeLLMSection(t, root, "llm:\n  performance_tier: \"max\"\n")
	if got := ResolveProjectPerformanceTier(root); got != "max" {
		t.Errorf("explicit max: got %q, want max", got)
	}

	writeLLMSection(t, root, "llm:\n  performance_tier: \"\"\n")
	if got := ResolveProjectPerformanceTier(root); got != "medium" {
		t.Errorf("empty performance_tier: got %q, want medium", got)
	}

	writeLLMSection(t, root, "llm:\n  performance_tier: low\n")
	if got := ResolveProjectPerformanceTier(root); got != "low" {
		t.Errorf("unquoted low: got %q, want low", got)
	}
}
