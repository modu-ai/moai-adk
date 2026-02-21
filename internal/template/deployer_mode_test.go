package template

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// TestDeployerWithInstallationMode tests the deployer with installation mode support.
// TAG-002: Deployer Interface Enhancement

func TestDeployerWithInstallationMode(t *testing.T) {
	t.Run("local_mode_deploys_to_project_root", func(t *testing.T) {
		// RED: Test that local mode deploys files to project .claude/
		root, mgr := setupDeployProject(t)

		fs := fstest.MapFS{
			".claude/agents/moai/expert-backend.md": &fstest.MapFile{
				Data: []byte("# Expert Backend Agent"),
			},
			".claude/skills/moai/SKILL.md": &fstest.MapFile{
				Data: []byte("# MoAI Core Skill"),
			},
		}

		// Create deployer with local mode context
		tmplCtx := NewTemplateContext(
			WithProject("testproject", root),
			WithInstallationMode("local"),
		)
		d := NewDeployer(fs)

		err := d.Deploy(context.Background(), root, mgr, tmplCtx)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		// In local mode, files should be in project .claude/
		expectedFiles := []string{
			".claude/agents/moai/expert-backend.md",
			".claude/skills/moai/SKILL.md",
		}
		for _, f := range expectedFiles {
			absPath := filepath.Join(root, f)
			if _, err := os.Stat(absPath); err != nil {
				t.Errorf("local mode: expected file %q in project: %v", f, err)
			}
		}
	})

	t.Run("global_mode_deploys_agents_to_home_claude", func(t *testing.T) {
		// RED: Test that global mode deploys agents/skills/rules to ~/.claude/
		root, mgr := setupDeployProject(t)
		homeDir := t.TempDir() // Simulated home directory

		fs := fstest.MapFS{
			".claude/agents/moai/expert-backend.md": &fstest.MapFile{
				Data: []byte("# Expert Backend Agent"),
			},
			".claude/skills/moai-foundation-core/SKILL.md": &fstest.MapFile{
				Data: []byte("# Foundation Core Skill"),
			},
			".claude/rules/moai/core/constitution.md": &fstest.MapFile{
				Data: []byte("# MoAI Constitution"),
			},
			".claude/hooks/moai/handle-session-start.sh": &fstest.MapFile{
				Data: []byte("#!/bin/bash\necho 'hook'"),
			},
			".claude/settings.json": &fstest.MapFile{
				Data: []byte(`{"hooks":{}}`),
			},
		}

		// Create deployer with global mode context
		tmplCtx := NewTemplateContext(
			WithProject("testproject", root),
			WithInstallationMode("global"),
			WithHomeDir(homeDir),
		)

		// Need a deployer that supports mode-aware deployment
		d := NewDeployerWithMode(fs, "global", homeDir)

		err := d.Deploy(context.Background(), root, mgr, tmplCtx)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		// In global mode:
		// - agents/moai/ -> ~/.claude/agents/moai/
		// - skills/moai*/ -> ~/.claude/skills/moai*/
		// - rules/moai/ -> ~/.claude/rules/moai/
		// - hooks/moai/ -> project/.claude/hooks/moai/ (always local)
		// - settings.json -> project/.claude/settings.json (always local)

		globalClaude := filepath.Join(homeDir, ".claude")

		// Global files should be in ~/.claude/
		globalFiles := []string{
			filepath.Join(globalClaude, "agents/moai/expert-backend.md"),
			filepath.Join(globalClaude, "skills/moai-foundation-core/SKILL.md"),
			filepath.Join(globalClaude, "rules/moai/core/constitution.md"),
		}
		for _, f := range globalFiles {
			if _, err := os.Stat(f); err != nil {
				t.Errorf("global mode: expected file %q in home: %v", f, err)
			}
		}

		// Local files should still be in project root
		localFiles := []string{
			filepath.Join(root, ".claude/hooks/moai/handle-session-start.sh"),
			filepath.Join(root, ".claude/settings.json"),
		}
		for _, f := range localFiles {
			if _, err := os.Stat(f); err != nil {
				t.Errorf("global mode: expected file %q in project: %v", f, err)
			}
		}
	})

	t.Run("global_mode_does_not_overwrite_existing_global_files", func(t *testing.T) {
		// RED: Test that global mode respects existing files in ~/.claude/
		root, mgr := setupDeployProject(t)
		homeDir := t.TempDir()

		// Pre-create global file with different content
		globalClaude := filepath.Join(homeDir, ".claude")
		agentsDir := filepath.Join(globalClaude, "agents", "moai")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatalf("mkdir error: %v", err)
		}
		existingContent := []byte("# Existing Agent - DO NOT OVERWRITE")
		if err := os.WriteFile(filepath.Join(agentsDir, "expert-backend.md"), existingContent, 0o644); err != nil {
			t.Fatalf("write error: %v", err)
		}

		fs := fstest.MapFS{
			".claude/agents/moai/expert-backend.md": &fstest.MapFile{
				Data: []byte("# New Agent Content"),
			},
		}

		tmplCtx := NewTemplateContext(
			WithProject("testproject", root),
			WithInstallationMode("global"),
			WithHomeDir(homeDir),
		)
		d := NewDeployerWithMode(fs, "global", homeDir)

		err := d.Deploy(context.Background(), root, mgr, tmplCtx)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		// Existing file should NOT be overwritten
		data, err := os.ReadFile(filepath.Join(agentsDir, "expert-backend.md"))
		if err != nil {
			t.Fatalf("read error: %v", err)
		}
		if string(data) != string(existingContent) {
			t.Errorf("global mode overwrote existing file: got %q, want %q", string(data), string(existingContent))
		}
	})
}

func TestResolveDeployPath(t *testing.T) {
	// Test helper function for path resolution based on installation mode

	t.Run("local_mode_returns_project_path", func(t *testing.T) {
		root := "/project"
		relPath := ".claude/agents/moai/expert-backend.md"
		homeDir := "/home/user"

		result := ResolveDeployPath(relPath, "local", root, homeDir)
		expected := filepath.Join(root, relPath)

		if result != expected {
			t.Errorf("ResolveDeployPath(local) = %q, want %q", result, expected)
		}
	})

	t.Run("global_mode_agents_to_home", func(t *testing.T) {
		root := "/project"
		homeDir := "/home/user"

		tests := []struct {
			relPath  string
			expected string
		}{
			{
				".claude/agents/moai/expert-backend.md",
				filepath.Join(homeDir, ".claude/agents/moai/expert-backend.md"),
			},
			{
				".claude/skills/moai-foundation-core/SKILL.md",
				filepath.Join(homeDir, ".claude/skills/moai-foundation-core/SKILL.md"),
			},
			{
				".claude/rules/moai/core/constitution.md",
				filepath.Join(homeDir, ".claude/rules/moai/core/constitution.md"),
			},
		}

		for _, tt := range tests {
			result := ResolveDeployPath(tt.relPath, "global", root, homeDir)
			if result != tt.expected {
				t.Errorf("ResolveDeployPath(global, %q) = %q, want %q", tt.relPath, result, tt.expected)
			}
		}
	})

	t.Run("global_mode_keeps_hooks_local", func(t *testing.T) {
		root := "/project"
		homeDir := "/home/user"
		relPath := ".claude/hooks/moai/handle-session-start.sh"

		result := ResolveDeployPath(relPath, "global", root, homeDir)
		expected := filepath.Join(root, relPath)

		if result != expected {
			t.Errorf("ResolveDeployPath(global, hooks) = %q, want %q", result, expected)
		}
	})

	t.Run("global_mode_keeps_settings_local", func(t *testing.T) {
		root := "/project"
		homeDir := "/home/user"

		tests := []string{
			".claude/settings.json",
			".claude/settings.local.json",
		}

		for _, relPath := range tests {
			result := ResolveDeployPath(relPath, "global", root, homeDir)
			expected := filepath.Join(root, relPath)

			if result != expected {
				t.Errorf("ResolveDeployPath(global, %q) = %q, want %q", relPath, result, expected)
			}
		}
	})
}

func TestIsGlobalFile(t *testing.T) {
	// Test helper function to determine if a file should be global

	tests := []struct {
		path     string
		expected bool
	}{
		{".claude/agents/moai/expert-backend.md", true},
		{".claude/agents/moai/manager-spec.md", true},
		{".claude/skills/moai/SKILL.md", true},
		{".claude/skills/moai-foundation-core/SKILL.md", true},
		{".claude/skills/moai-workflow-project/SKILL.md", true},
		{".claude/rules/moai/core/constitution.md", true},
		{".claude/rules/moai/workflow/spec-workflow.md", true},
		{".claude/hooks/moai/handle-session-start.sh", false},
		{".claude/settings.json", false},
		{".claude/settings.local.json", false},
		{".moai/config/config.yaml", false},
		{"CLAUDE.md", false},
	}

	for _, tt := range tests {
		result := IsGlobalFile(tt.path)
		if result != tt.expected {
			t.Errorf("IsGlobalFile(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

// Test that TemplateContext has InstallationMode field
func TestTemplateContextInstallationMode(t *testing.T) {
	t.Run("default_is_local", func(t *testing.T) {
		ctx := NewTemplateContext()
		if ctx.InstallationMode != "local" {
			t.Errorf("default InstallationMode = %q, want %q", ctx.InstallationMode, "local")
		}
	})

	t.Run("with_installation_mode_option", func(t *testing.T) {
		ctx := NewTemplateContext(WithInstallationMode("global"))
		if ctx.InstallationMode != "global" {
			t.Errorf("InstallationMode = %q, want %q", ctx.InstallationMode, "global")
		}
	})

	t.Run("invalid_mode_falls_back_to_default", func(t *testing.T) {
		ctx := NewTemplateContext(WithInstallationMode("invalid"))
		if ctx.InstallationMode != "local" {
			t.Errorf("InstallationMode = %q, want %q", ctx.InstallationMode, "local")
		}
	})
}

func TestModeAwareDeployerMethods(t *testing.T) {
	t.Run("ExtractTemplate", func(t *testing.T) {
		fs := fstest.MapFS{
			".claude/agents/moai/test.md": &fstest.MapFile{
				Data: []byte("# Test Agent"),
			},
		}
		d := NewDeployerWithMode(fs, "local", "/home/test")

		data, err := d.ExtractTemplate(".claude/agents/moai/test.md")
		if err != nil {
			t.Fatalf("ExtractTemplate error: %v", err)
		}
		if string(data) != "# Test Agent" {
			t.Errorf("ExtractTemplate = %q, want %q", string(data), "# Test Agent")
		}
	})

	t.Run("ExtractTemplate_not_found", func(t *testing.T) {
		fs := fstest.MapFS{}
		d := NewDeployerWithMode(fs, "local", "/home/test")

		_, err := d.ExtractTemplate("nonexistent.md")
		if err == nil {
			t.Fatal("expected error for nonexistent template")
		}
	})

	t.Run("ListTemplates", func(t *testing.T) {
		fs := fstest.MapFS{
			".claude/agents/moai/test.md":         &fstest.MapFile{Data: []byte("agent")},
			".claude/skills/moai-foundation/SKILL.md": &fstest.MapFile{Data: []byte("skill")},
		}
		d := NewDeployerWithMode(fs, "local", "/home/test")

		list := d.ListTemplates()
		if len(list) != 2 {
			t.Errorf("ListTemplates returned %d items, want 2", len(list))
		}
	})

	t.Run("NewDeployerWithModeAndRenderer", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.md.tmpl": &fstest.MapFile{Data: []byte("Hello {{.Name}}")},
		}
		renderer := NewRenderer(fs)
		d := NewDeployerWithModeAndRenderer(fs, renderer, "local", "/home/test")

		// Verify it works
		data, err := d.ExtractTemplate("test.md.tmpl")
		if err != nil {
			t.Fatalf("ExtractTemplate error: %v", err)
		}
		if string(data) != "Hello {{.Name}}" {
			t.Errorf("ExtractTemplate = %q", string(data))
		}
	})
}

func TestGlobalModeForceUpdate(t *testing.T) {
	t.Run("force_update_overwrites_existing_global_files", func(t *testing.T) {
		root, mgr := setupDeployProject(t)
		homeDir := t.TempDir()

		// Pre-create global file with different content
		globalClaude := filepath.Join(homeDir, ".claude")
		agentsDir := filepath.Join(globalClaude, "agents", "moai")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatalf("mkdir error: %v", err)
		}
		oldContent := []byte("# OLD CONTENT")
		if err := os.WriteFile(filepath.Join(agentsDir, "expert-backend.md"), oldContent, 0o644); err != nil {
			t.Fatalf("write error: %v", err)
		}

		fs := fstest.MapFS{
			".claude/agents/moai/expert-backend.md": &fstest.MapFile{
				Data: []byte("# NEW CONTENT"),
			},
		}

		// Create deployer with forceUpdate=true
		d := NewDeployerWithModeAndRendererForceUpdate(fs, nil, "global", homeDir, true)

		err := d.Deploy(context.Background(), root, mgr, nil)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		// File should be overwritten
		data, err := os.ReadFile(filepath.Join(agentsDir, "expert-backend.md"))
		if err != nil {
			t.Fatalf("read error: %v", err)
		}
		if string(data) != "# NEW CONTENT" {
			t.Errorf("forceUpdate did not overwrite: got %q, want %q", string(data), "# NEW CONTENT")
		}
	})
}

// TestFilePlacementStrategy verifies the file placement strategy from SPEC-UPDATE-001
// REQ-013: Global mode deploys agents/moai, skills/moai*, rules/moai to ~/.claude/
// REQ-002: hooks/moai always stays local
// REQ-003: settings.json always stays local
func TestFilePlacementStrategy(t *testing.T) {
	t.Run("global_mode_file_placement", func(t *testing.T) {
		root, mgr := setupDeployProject(t)
		homeDir := t.TempDir()
		projectRoot := root

		fs := fstest.MapFS{
			// Global files (should go to ~/.claude/)
			".claude/agents/moai/expert-backend.md": &fstest.MapFile{
				Data: []byte("# Expert Backend Agent"),
			},
			".claude/skills/moai-foundation-core/SKILL.md": &fstest.MapFile{
				Data: []byte("# Foundation Core"),
			},
			".claude/skills/moai-workflow-project/SKILL.md": &fstest.MapFile{
				Data: []byte("# Workflow Project"),
			},
			".claude/rules/moai/core/constitution.md": &fstest.MapFile{
				Data: []byte("# MoAI Constitution"),
			},
			// Local files (always in project)
			".claude/hooks/moai/handle-session-start.sh": &fstest.MapFile{
				Data: []byte("#!/bin/bash\necho 'hook'"),
			},
			".claude/settings.json": &fstest.MapFile{
				Data: []byte(`{"hooks":{}}`),
			},
		}

		d := NewDeployerWithMode(fs, "global", homeDir)
		err := d.Deploy(context.Background(), projectRoot, mgr, nil)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		globalClaude := filepath.Join(homeDir, ".claude")

		// REQ-013: Global files should be in ~/.claude/
		globalExpectedFiles := []string{
			filepath.Join(globalClaude, "agents/moai/expert-backend.md"),
			filepath.Join(globalClaude, "skills/moai-foundation-core/SKILL.md"),
			filepath.Join(globalClaude, "skills/moai-workflow-project/SKILL.md"),
			filepath.Join(globalClaude, "rules/moai/core/constitution.md"),
		}
		for _, f := range globalExpectedFiles {
			if _, err := os.Stat(f); err != nil {
				t.Errorf("REQ-013: expected global file %q: %v", f, err)
			}
		}

		// REQ-002, REQ-003: Local files should be in project root
		localExpectedFiles := []string{
			filepath.Join(projectRoot, ".claude/hooks/moai/handle-session-start.sh"),
			filepath.Join(projectRoot, ".claude/settings.json"),
		}
		for _, f := range localExpectedFiles {
			if _, err := os.Stat(f); err != nil {
				t.Errorf("REQ-002/REQ-003: expected local file %q: %v", f, err)
			}
		}

		// Global files should NOT be in project root
		globalNotInProject := []string{
			filepath.Join(projectRoot, ".claude/agents/moai/expert-backend.md"),
			filepath.Join(projectRoot, ".claude/skills/moai-foundation-core/SKILL.md"),
		}
		for _, f := range globalNotInProject {
			if _, err := os.Stat(f); err == nil {
				t.Errorf("global file should NOT be in project: %q", f)
			}
		}
	})

	t.Run("local_mode_all_files_in_project", func(t *testing.T) {
		root, mgr := setupDeployProject(t)
		projectRoot := root

		fs := fstest.MapFS{
			".claude/agents/moai/expert-backend.md": &fstest.MapFile{
				Data: []byte("# Expert Backend Agent"),
			},
			".claude/skills/moai-foundation-core/SKILL.md": &fstest.MapFile{
				Data: []byte("# Foundation Core"),
			},
			".claude/rules/moai/core/constitution.md": &fstest.MapFile{
				Data: []byte("# MoAI Constitution"),
			},
			".claude/hooks/moai/handle-session-start.sh": &fstest.MapFile{
				Data: []byte("#!/bin/bash\necho 'hook'"),
			},
			".claude/settings.json": &fstest.MapFile{
				Data: []byte(`{"hooks":{}}`),
			},
		}

		d := NewDeployerWithMode(fs, "local", "/home/test")
		err := d.Deploy(context.Background(), projectRoot, mgr, nil)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		// REQ-014: All files should be in project .claude/
		expectedFiles := []string{
			filepath.Join(projectRoot, ".claude/agents/moai/expert-backend.md"),
			filepath.Join(projectRoot, ".claude/skills/moai-foundation-core/SKILL.md"),
			filepath.Join(projectRoot, ".claude/rules/moai/core/constitution.md"),
			filepath.Join(projectRoot, ".claude/hooks/moai/handle-session-start.sh"),
			filepath.Join(projectRoot, ".claude/settings.json"),
		}
		for _, f := range expectedFiles {
			if _, err := os.Stat(f); err != nil {
				t.Errorf("REQ-014: expected local file %q: %v", f, err)
			}
		}
	})
}
