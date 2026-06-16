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

func TestGetAgentModel(t *testing.T) {
	tests := []struct {
		name      string
		policy    ModelPolicy
		agentName string
		want      string
	}{
		// Retained Manager agents (SPEC-CC2178-MODEL-POLICY-REPAIR-001 M2 catalog)
		{"spec_high", ModelPolicyHigh, "manager-spec", "opus"},
		{"spec_medium", ModelPolicyMedium, "manager-spec", "opus"},
		{"spec_low", ModelPolicyLow, "manager-spec", "sonnet"},
		{"develop_high", ModelPolicyHigh, "manager-develop", "sonnet"},
		{"develop_medium", ModelPolicyMedium, "manager-develop", "sonnet"},
		{"develop_low", ModelPolicyLow, "manager-develop", "haiku"},
		{"docs_high", ModelPolicyHigh, "manager-docs", "sonnet"},
		{"docs_low", ModelPolicyLow, "manager-docs", "haiku"},
		{"git_high", ModelPolicyHigh, "manager-git", "haiku"},

		// Retained Builder agent
		{"builder_harness_high", ModelPolicyHigh, "builder-harness", "sonnet"},
		{"builder_harness_low", ModelPolicyLow, "builder-harness", "haiku"},

		// Archived phantom agents (post-cleanup): return "" (absent from map)
		{"quality_archived", ModelPolicyHigh, "manager-quality", ""},
		{"backend_archived", ModelPolicyHigh, "expert-backend", ""},
		{"security_archived", ModelPolicyHigh, "expert-security", ""},
		{"builder_agent_archived", ModelPolicyHigh, "builder-agent", ""},

		// Unknown agent: returns "" (skip sentinel - preserve current model)
		{"unknown_agent", ModelPolicyHigh, "nonexistent-agent", ""},

		// Invalid policy: returns "sonnet" as safe fallback
		{"invalid_policy", ModelPolicy("invalid"), "manager-spec", "sonnet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAgentModel(tt.policy, tt.agentName)
			if got != tt.want {
				t.Errorf("GetAgentModel(%q, %q) = %q, want %q", tt.policy, tt.agentName, got, tt.want)
			}
		})
	}
}

func TestApplyModelPolicy(t *testing.T) {
	t.Run("applies_policy_to_agent_files", func(t *testing.T) {
		root := t.TempDir()
		// SPEC-CC2178-MODEL-POLICY-REPAIR-001 M2: manager-develop lives in core/ and is in the retained map.
		agentsDir := filepath.Join(root, ".claude", "agents", "moai")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}

		// Create a mock agent file with model: line
		agentContent := `---
name: manager-develop
description: Run-phase implementation agent
model: opus
---
# Manager Develop Agent
`
		if err := os.WriteFile(filepath.Join(agentsDir, "manager-develop.md"), []byte(agentContent), 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}

		// Set up manifest
		mgr := manifest.NewManager()
		moaiDir := filepath.Join(root, ".moai")
		if err := os.MkdirAll(moaiDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("manifest Load error: %v", err)
		}

		// Apply low policy (manager-develop low tuple is haiku → change from opus to haiku)
		err := ApplyModelPolicy(root, ModelPolicyLow, mgr)
		if err != nil {
			t.Fatalf("ApplyModelPolicy error: %v", err)
		}

		// Verify the file was updated
		content, err := os.ReadFile(filepath.Join(agentsDir, "manager-develop.md"))
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		if got := string(content); got == agentContent {
			t.Error("file was not modified by ApplyModelPolicy")
		}
		// The model line should now be "model: haiku" (manager-develop low tuple)
		want := "model: haiku"
		if got := string(content); !containsString(got, want) {
			t.Errorf("content does not contain %q:\n%s", want, got)
		}
	})

	t.Run("no_agents_directory", func(t *testing.T) {
		root := t.TempDir()
		mgr := manifest.NewManager()
		moaiDir := filepath.Join(root, ".moai")
		if err := os.MkdirAll(moaiDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("manifest Load error: %v", err)
		}

		// Should not error when agents directory does not exist
		err := ApplyModelPolicy(root, ModelPolicyHigh, mgr)
		if err != nil {
			t.Fatalf("ApplyModelPolicy error: %v", err)
		}
	})

	t.Run("skips_non_md_files", func(t *testing.T) {
		root := t.TempDir()
		// Post SPEC-V3R6-AGENT-FOLDER-SPLIT-001: any domain folder works for non-.md exclusion test.
		agentsDir := filepath.Join(root, ".claude", "agents", "moai")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}

		// Create a non-.md file
		if err := os.WriteFile(filepath.Join(agentsDir, "readme.txt"), []byte("not an agent"), 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}

		mgr := manifest.NewManager()
		moaiDir := filepath.Join(root, ".moai")
		if err := os.MkdirAll(moaiDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("manifest Load error: %v", err)
		}

		err := ApplyModelPolicy(root, ModelPolicyHigh, mgr)
		if err != nil {
			t.Fatalf("ApplyModelPolicy error: %v", err)
		}
	})

	t.Run("skips_directories", func(t *testing.T) {
		root := t.TempDir()
		// Post SPEC-V3R6-AGENT-FOLDER-SPLIT-001: any domain folder works for directory-skip test.
		agentsDir := filepath.Join(root, ".claude", "agents", "moai")
		subDir := filepath.Join(agentsDir, "subdir.md")
		if err := os.MkdirAll(subDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}

		mgr := manifest.NewManager()
		moaiDir := filepath.Join(root, ".moai")
		if err := os.MkdirAll(moaiDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("manifest Load error: %v", err)
		}

		err := ApplyModelPolicy(root, ModelPolicyHigh, mgr)
		if err != nil {
			t.Fatalf("ApplyModelPolicy error: %v", err)
		}
	})

	t.Run("skips_unknown_agents", func(t *testing.T) {
		root := t.TempDir()
		// Post SPEC-V3R6-AGENT-FOLDER-SPLIT-001: unknown agent placed in core/ for walker discovery.
		agentsDir := filepath.Join(root, ".claude", "agents", "moai")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}

		// Create an agent file for an unknown agent name
		agentContent := `---
name: unknown-agent
model: opus
---
`
		if err := os.WriteFile(filepath.Join(agentsDir, "unknown-agent.md"), []byte(agentContent), 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}

		mgr := manifest.NewManager()
		moaiDir := filepath.Join(root, ".moai")
		if err := os.MkdirAll(moaiDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("manifest Load error: %v", err)
		}

		err := ApplyModelPolicy(root, ModelPolicyHigh, mgr)
		if err != nil {
			t.Fatalf("ApplyModelPolicy error: %v", err)
		}

		// File should be unchanged since unknown-agent returns "" (skip sentinel)
		content, err := os.ReadFile(filepath.Join(agentsDir, "unknown-agent.md"))
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		if string(content) != agentContent {
			t.Error("unknown agent file was modified, should have been skipped")
		}
	})

	t.Run("skips_unchanged_content", func(t *testing.T) {
		root := t.TempDir()
		// SPEC-CC2178-MODEL-POLICY-REPAIR-001 M2: manager-spec lives in core/ and is in the retained map.
		agentsDir := filepath.Join(root, ".claude", "agents", "moai")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}

		// Create a file where the model is already the target (manager-spec high tuple = opus)
		agentContent := `---
name: manager-spec
model: opus
---
`
		if err := os.WriteFile(filepath.Join(agentsDir, "manager-spec.md"), []byte(agentContent), 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}

		mgr := manifest.NewManager()
		moaiDir := filepath.Join(root, ".moai")
		if err := os.MkdirAll(moaiDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("manifest Load error: %v", err)
		}

		// Apply high policy: manager-spec high tuple is opus — file already at opus, no change
		err := ApplyModelPolicy(root, ModelPolicyHigh, mgr)
		if err != nil {
			t.Fatalf("ApplyModelPolicy error: %v", err)
		}

		// Content should remain unchanged
		content, err := os.ReadFile(filepath.Join(agentsDir, "manager-spec.md"))
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		if string(content) != agentContent {
			t.Error("file was modified but model was already correct")
		}
	})
}

// TestGetAgentEffort verifies the agentEffortMap and GetAgentEffort function.
//
// SPEC-CC2178-MODEL-POLICY-REPAIR-001 M2: the map was reconciled against the
// retained catalog — 3 archived phantom keys removed (manager-strategy,
// expert-security, expert-refactoring), plan-auditor/sync-auditor synced
// high→xhigh, manager-develop/builder-harness added.
func TestGetAgentEffort(t *testing.T) {
	tests := []struct {
		name      string
		agentName string
		want      string
	}{
		// Retained reasoning agents with explicit effort (post-reconciliation)
		{"manager-spec xhigh", "manager-spec", "xhigh"},
		{"plan-auditor xhigh", "plan-auditor", "xhigh"}, // REQ-MPR-011b synced high→xhigh
		{"sync-auditor xhigh", "sync-auditor", "xhigh"}, // REQ-MPR-011b synced high→xhigh
		{"manager-develop xhigh", "manager-develop", "xhigh"}, // REQ-MPR-011a added
		{"builder-harness high", "builder-harness", "high"},   // REQ-MPR-011a added
		// Archived phantoms (post-cleanup): return "" (absent from map)
		{"manager-strategy archived", "manager-strategy", ""},
		{"expert-security archived", "expert-security", ""},
		{"expert-refactoring archived", "expert-refactoring", ""},
		// Agents not in effort map: return "" (runtime default applies)
		{"manager-docs unset", "manager-docs", ""},
		{"manager-git unset", "manager-git", ""},
		{"unknown-agent unset", "some-nonexistent-agent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAgentEffort(tt.agentName)
			if got != tt.want {
				t.Errorf("GetAgentEffort(%q) = %q, want %q", tt.agentName, got, tt.want)
			}
		})
	}
}

// TestModelClaudeOpus47Constant verifies the claude-opus-4-7 model ID constant.
func TestModelClaudeOpus47Constant(t *testing.T) {
	if ModelIDOpus47 == "" {
		t.Error("ModelIDOpus47 constant is empty, want non-empty model ID")
	}
	want := "claude-opus-4-7"
	if ModelIDOpus47 != want {
		t.Errorf("ModelIDOpus47 = %q, want %q", ModelIDOpus47, want)
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

// TestAgentModelMapSignatureUnchanged verifies the existing agentModelMap type
// and all existing public API functions remain unchanged (NFR-1 no-break).
func TestAgentModelMapSignatureUnchanged(t *testing.T) {
	// ValidModelPolicies still returns exactly 3 items.
	policies := ValidModelPolicies()
	if len(policies) != 3 {
		t.Errorf("ValidModelPolicies() len = %d, want 3 (signature must not change)", len(policies))
	}

	// IsValidModelPolicy still accepts only high/medium/low.
	for _, p := range []string{"high", "medium", "low"} {
		if !IsValidModelPolicy(p) {
			t.Errorf("IsValidModelPolicy(%q) = false, want true", p)
		}
	}
	// "xhigh" is NOT a ModelPolicy (it's an effort level, separate concern).
	if IsValidModelPolicy("xhigh") {
		t.Error("IsValidModelPolicy(xhigh) = true, want false (xhigh is effort, not policy)")
	}

	// GetAgentModel still returns "" for unknown agents.
	if got := GetAgentModel(ModelPolicyHigh, "nonexistent"); got != "" {
		t.Errorf("GetAgentModel(high, nonexistent) = %q, want empty string", got)
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

// TestApplyEffortPolicy verifies ApplyEffortPolicy behaviour across multiple scenarios.
func TestApplyEffortPolicy(t *testing.T) {
	t.Run("injects_effort_for_reasoning_agent", func(t *testing.T) {
		root := t.TempDir()
		// Post SPEC-V3R6-AGENT-FOLDER-SPLIT-001: manager-* agents live in core/.
		agentsDir := filepath.Join(root, ".claude", "agents", "moai")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}

		// manager-spec has no effort: field yet — should be injected as xhigh
		agentContent := `---
name: manager-spec
model: opus
permissionMode: bypassPermissions
---
# Manager Spec Agent
`
		if err := os.WriteFile(filepath.Join(agentsDir, "manager-spec.md"), []byte(agentContent), 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}

		mgr := manifest.NewManager()
		moaiDir := filepath.Join(root, ".moai")
		if err := os.MkdirAll(moaiDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("manifest Load error: %v", err)
		}

		if err := ApplyEffortPolicy(root, mgr); err != nil {
			t.Fatalf("ApplyEffortPolicy error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(agentsDir, "manager-spec.md"))
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		if !containsString(string(content), "effort: xhigh") {
			t.Errorf("expected effort: xhigh injected, got:\n%s", content)
		}
		// Existing fields must be preserved
		if !containsString(string(content), "model: opus") {
			t.Errorf("model field was lost:\n%s", content)
		}
	})

	t.Run("preserves_existing_effort_value", func(t *testing.T) {
		root := t.TempDir()
		// SPEC-CC2178-MODEL-POLICY-REPAIR-001 M2: plan-auditor lives in meta/ and is in the retained effort map (xhigh).
		agentsDir := filepath.Join(root, ".claude", "agents", "moai")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}

		// plan-auditor already has effort: max (user override) — must not be changed to xhigh
		agentContent := `---
name: plan-auditor
model: inherit
effort: max
---
# Plan Auditor
`
		if err := os.WriteFile(filepath.Join(agentsDir, "plan-auditor.md"), []byte(agentContent), 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}

		mgr := manifest.NewManager()
		moaiDir := filepath.Join(root, ".moai")
		if err := os.MkdirAll(moaiDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("manifest Load error: %v", err)
		}

		if err := ApplyEffortPolicy(root, mgr); err != nil {
			t.Fatalf("ApplyEffortPolicy error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(agentsDir, "plan-auditor.md"))
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		if string(content) != agentContent {
			t.Errorf("existing effort was modified; want preserved:\ngot:\n%s\nwant:\n%s", content, agentContent)
		}
	})

	t.Run("no_op_for_agent_not_in_effort_map", func(t *testing.T) {
		root := t.TempDir()
		// SPEC-CC2178-MODEL-POLICY-REPAIR-001 M2: manager-docs lives in core/ and is NOT in the effort map.
		agentsDir := filepath.Join(root, ".claude", "agents", "moai")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}

		// manager-docs is NOT in agentEffortMap — nothing should be injected
		agentContent := `---
name: manager-docs
model: haiku
---
# Manager Docs
`
		if err := os.WriteFile(filepath.Join(agentsDir, "manager-docs.md"), []byte(agentContent), 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}

		mgr := manifest.NewManager()
		moaiDir := filepath.Join(root, ".moai")
		if err := os.MkdirAll(moaiDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("manifest Load error: %v", err)
		}

		if err := ApplyEffortPolicy(root, mgr); err != nil {
			t.Fatalf("ApplyEffortPolicy error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(agentsDir, "manager-docs.md"))
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		if string(content) != agentContent {
			t.Errorf("agent not in effort map was modified; want no-op:\ngot:\n%s\nwant:\n%s", content, agentContent)
		}
	})

	t.Run("no_agents_directory", func(t *testing.T) {
		root := t.TempDir()
		mgr := manifest.NewManager()
		moaiDir := filepath.Join(root, ".moai")
		if err := os.MkdirAll(moaiDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("manifest Load error: %v", err)
		}

		// Should not error when directory is absent
		if err := ApplyEffortPolicy(root, mgr); err != nil {
			t.Fatalf("ApplyEffortPolicy error: %v", err)
		}
	})

	t.Run("no_frontmatter_not_modified", func(t *testing.T) {
		root := t.TempDir()
		// Post SPEC-V3R6-AGENT-FOLDER-SPLIT-001: manager-* agents live in core/.
		agentsDir := filepath.Join(root, ".claude", "agents", "moai")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}

		// File without YAML frontmatter — skip silently
		agentContent := "# manager-spec\nNo frontmatter here.\n"
		if err := os.WriteFile(filepath.Join(agentsDir, "manager-spec.md"), []byte(agentContent), 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}

		mgr := manifest.NewManager()
		moaiDir := filepath.Join(root, ".moai")
		if err := os.MkdirAll(moaiDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("manifest Load error: %v", err)
		}

		if err := ApplyEffortPolicy(root, mgr); err != nil {
			t.Fatalf("ApplyEffortPolicy error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(agentsDir, "manager-spec.md"))
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		if string(content) != agentContent {
			t.Errorf("file without frontmatter was modified:\ngot:\n%s\nwant:\n%s", content, agentContent)
		}
	})

	t.Run("manifest_tracked_after_injection", func(t *testing.T) {
		root := t.TempDir()
		// Post SPEC-V3R6-AGENT-FOLDER-SPLIT-001: plan-auditor lives in meta/.
		agentsDir := filepath.Join(root, ".claude", "agents", "moai")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}

		agentContent := `---
name: plan-auditor
model: opus
---
# Plan Auditor
`
		if err := os.WriteFile(filepath.Join(agentsDir, "plan-auditor.md"), []byte(agentContent), 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}

		mgr := manifest.NewManager()
		moaiDir := filepath.Join(root, ".moai")
		if err := os.MkdirAll(moaiDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("manifest Load error: %v", err)
		}

		if err := ApplyEffortPolicy(root, mgr); err != nil {
			t.Fatalf("ApplyEffortPolicy error: %v", err)
		}

		// Manifest entry must be updated for the injected file
		relPath := filepath.Join(".claude", "agents", "moai", "plan-auditor.md")
		entry, found := mgr.GetEntry(relPath)
		if !found {
			t.Errorf("manifest entry not found for %q after injection", relPath)
		} else if entry.TemplateHash == "" {
			t.Errorf("manifest entry TemplateHash is empty for %q", relPath)
		}
	})
}

// === SPEC-CC2178-MODEL-POLICY-REPAIR-001 M2 (RED characterization tests) ===
//
// These tests assert the POST-CLEANUP state of agentModelMap and agentEffortMap.
// They are written BEFORE the GREEN map edits (TDD RED phase) and are expected
// to FAIL against the current (stale) maps. After the M2 GREEN edits remove
// the 16 phantom model keys + 3 phantom effort keys + add manager-develop/
// builder-harness + reconcile effort values, these tests PASS.

// canonicalPhantomModelKeys is the authoritative 16-key enumeration from
// spec.md §C.3. All 16 MUST be absent from agentModelMap post-cleanup.
var canonicalPhantomModelKeys = []string{
	"manager-ddd", "manager-tdd", // legacy aliases of manager-develop
	"manager-quality", "manager-project", "manager-strategy", // archived core managers
	"expert-backend", "expert-frontend", "expert-security", "expert-devops",
	"expert-performance", "expert-debug", "expert-testing", "expert-refactoring", // archived expert-*
	"builder-agent", "builder-skill", "builder-plugin", // archived/legacy builder variants
}

// TestAgentModelMap_NoPhantomKeys (AC-MPR-007, REQ-MPR-008) verifies that all
// 16 canonical phantom keys are ABSENT from agentModelMap.
func TestAgentModelMap_NoPhantomKeys(t *testing.T) {
	for _, key := range canonicalPhantomModelKeys {
		t.Run("phantom_absent_"+key, func(t *testing.T) {
			if got := GetAgentModel(ModelPolicyHigh, key); got != "" {
				t.Errorf("GetAgentModel(high, %q) = %q; phantom key MUST be absent (return \"\"), want \"\"", key, got)
			}
		})
	}
}

// TestAgentModelMap_RetainedAgents (AC-MPR-008, REQ-MPR-009 iter-2) verifies
// the 5 retained agents are present with correct tuples. manager-develop and
// builder-harness use the iter-2 tuple {sonnet, sonnet, haiku} (NOT the
// iter-1 {opus, sonnet, sonnet} derived from retired aliases — D6 rationale).
func TestAgentModelMap_RetainedAgents(t *testing.T) {
	tests := []struct {
		name      string
		policy    ModelPolicy
		agentName string
		want      string
	}{
		// manager-spec (retained-correct): {opus, opus, sonnet}
		{"manager-spec_high", ModelPolicyHigh, "manager-spec", "opus"},
		{"manager-spec_medium", ModelPolicyMedium, "manager-spec", "opus"},
		{"manager-spec_low", ModelPolicyLow, "manager-spec", "sonnet"},
		// manager-docs (retained-correct): {sonnet, haiku, haiku}
		{"manager-docs_high", ModelPolicyHigh, "manager-docs", "sonnet"},
		{"manager-docs_medium", ModelPolicyMedium, "manager-docs", "haiku"},
		{"manager-docs_low", ModelPolicyLow, "manager-docs", "haiku"},
		// manager-git (retained-correct): {haiku, haiku, haiku}
		{"manager-git_high", ModelPolicyHigh, "manager-git", "haiku"},
		// manager-develop (iter-2 tuple): {sonnet, sonnet, haiku}
		{"manager-develop_high", ModelPolicyHigh, "manager-develop", "sonnet"},
		{"manager-develop_medium", ModelPolicyMedium, "manager-develop", "sonnet"},
		{"manager-develop_low", ModelPolicyLow, "manager-develop", "haiku"},
		// builder-harness (iter-2 tuple): {sonnet, sonnet, haiku}
		{"builder-harness_high", ModelPolicyHigh, "builder-harness", "sonnet"},
		{"builder-harness_medium", ModelPolicyMedium, "builder-harness", "sonnet"},
		{"builder-harness_low", ModelPolicyLow, "builder-harness", "haiku"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAgentModel(tt.policy, tt.agentName); got != tt.want {
				t.Errorf("GetAgentModel(%q, %q) = %q, want %q", tt.policy, tt.agentName, got, tt.want)
			}
		})
	}
}

// TestAgentModelMap_EntryCount verifies exactly 5 entries remain post-cleanup
// (manager-spec, manager-develop, manager-docs, manager-git, builder-harness).
// This is a structural guard against accidental re-addition of phantoms.
func TestAgentModelMap_EntryCount(t *testing.T) {
	if got := len(agentModelMap); got != 5 {
		t.Errorf("agentModelMap has %d entries, want exactly 5 (manager-spec, manager-develop, manager-docs, manager-git, builder-harness)", got)
	}
}

// TestAgentEffortMap_NoPhantomKeys (AC-MPR-009, REQ-MPR-010/011a) verifies the
// 3 archived phantom effort keys are ABSENT.
func TestAgentEffortMap_NoPhantomKeys(t *testing.T) {
	phantoms := []string{"manager-strategy", "expert-security", "expert-refactoring"}
	for _, key := range phantoms {
		t.Run("effort_phantom_absent_"+key, func(t *testing.T) {
			if got := GetAgentEffort(key); got != "" {
				t.Errorf("GetAgentEffort(%q) = %q; phantom effort key MUST be absent, want \"\"", key, got)
			}
		})
	}
}

// TestAgentEffortMap_ReconciledValues (AC-MPR-010, REQ-MPR-011a/011b) verifies
// the map↔file reconciliation: plan-auditor and sync-auditor synced from high
// to xhigh; manager-develop and builder-harness added.
func TestAgentEffortMap_ReconciledValues(t *testing.T) {
	tests := []struct {
		name      string
		agentName string
		want      string
	}{
		// Retained + reconciled (map←file sync)
		{"manager-spec_xhigh", "manager-spec", "xhigh"},
		{"plan-auditor_xhigh", "plan-auditor", "xhigh"}, // was "high", synced to "xhigh"
		{"sync-auditor_xhigh", "sync-auditor", "xhigh"}, // was "high", synced to "xhigh"
		// Added retained agents
		{"manager-develop_xhigh", "manager-develop", "xhigh"}, // missing, added as xhigh per file
		{"builder-harness_high", "builder-harness", "high"},   // missing, added as high per file
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAgentEffort(tt.agentName); got != tt.want {
				t.Errorf("GetAgentEffort(%q) = %q, want %q", tt.agentName, got, tt.want)
			}
		})
	}
}

// containsString checks if s contains substr.
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
