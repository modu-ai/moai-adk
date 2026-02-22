package template

import (
	"context"
	"os"
	"path/filepath"
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
		// Manager agents
		{"spec_high", ModelPolicyHigh, "manager-spec", "opus"},
		{"spec_medium", ModelPolicyMedium, "manager-spec", "opus"},
		{"spec_low", ModelPolicyLow, "manager-spec", "sonnet"},
		{"docs_high", ModelPolicyHigh, "manager-docs", "sonnet"},
		{"docs_low", ModelPolicyLow, "manager-docs", "haiku"},
		{"quality_high", ModelPolicyHigh, "manager-quality", "haiku"},

		// Expert agents
		{"backend_high", ModelPolicyHigh, "expert-backend", "opus"},
		{"backend_medium", ModelPolicyMedium, "expert-backend", "sonnet"},
		{"backend_low", ModelPolicyLow, "expert-backend", "sonnet"},
		{"security_high", ModelPolicyHigh, "expert-security", "opus"},
		{"security_medium", ModelPolicyMedium, "expert-security", "opus"},

		// Builder agents
		{"builder_agent_high", ModelPolicyHigh, "builder-agent", "opus"},
		{"builder_agent_low", ModelPolicyLow, "builder-agent", "haiku"},

		// Team agents
		{"researcher_high", ModelPolicyHigh, "team-researcher", "haiku"},
		{"tester_high", ModelPolicyHigh, "team-tester", "opus"},
		{"tester_low", ModelPolicyLow, "team-tester", "haiku"},

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
		agentsDir := filepath.Join(root, ".claude", "agents", "moai")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}

		// Create a mock agent file with model: line
		agentContent := `---
name: expert-backend
description: Backend expert agent
model: opus
---
# Expert Backend Agent
`
		if err := os.WriteFile(filepath.Join(agentsDir, "expert-backend.md"), []byte(agentContent), 0o644); err != nil {
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

		// Apply low policy (expert-backend should change from opus to sonnet)
		err := ApplyModelPolicy(root, ModelPolicyLow, mgr)
		if err != nil {
			t.Fatalf("ApplyModelPolicy error: %v", err)
		}

		// Verify the file was updated
		content, err := os.ReadFile(filepath.Join(agentsDir, "expert-backend.md"))
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		if got := string(content); got == agentContent {
			t.Error("file was not modified by ApplyModelPolicy")
		}
		// The model line should now be "model: sonnet"
		want := "model: sonnet"
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
		agentsDir := filepath.Join(root, ".claude", "agents", "moai")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}

		// Create a file where the model is already the target
		agentContent := `---
name: expert-backend
model: opus
---
`
		if err := os.WriteFile(filepath.Join(agentsDir, "expert-backend.md"), []byte(agentContent), 0o644); err != nil {
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

		// Apply high policy: expert-backend already has "opus" for high
		err := ApplyModelPolicy(root, ModelPolicyHigh, mgr)
		if err != nil {
			t.Fatalf("ApplyModelPolicy error: %v", err)
		}

		// Content should remain unchanged
		content, err := os.ReadFile(filepath.Join(agentsDir, "expert-backend.md"))
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		if string(content) != agentContent {
			t.Error("file was modified but model was already correct")
		}
	})
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

// containsString checks if s contains substr.
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
