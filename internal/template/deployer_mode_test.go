package template

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// TestNewDeployerWithMode verifies the basic constructor creates a working deployer.
func TestNewDeployerWithMode(t *testing.T) {
	fs := fstest.MapFS{
		"README.md": &fstest.MapFile{Data: []byte("hello")},
	}
	d := NewDeployerWithMode(fs, "", "")
	if d == nil {
		t.Fatal("NewDeployerWithMode returned nil")
	}

	// Verify it can list templates
	list := d.ListTemplates()
	if len(list) != 1 || list[0] != "README.md" {
		t.Errorf("ListTemplates() = %v, want [README.md]", list)
	}
}

// TestNewDeployerWithModeAndRenderer verifies construction with renderer.
func TestNewDeployerWithModeAndRenderer(t *testing.T) {
	fs := fstest.MapFS{
		"config.sh.tmpl": &fstest.MapFile{
			Data: []byte("#!/bin/bash\necho {{.ProjectName}}"),
		},
	}
	renderer := NewRenderer(fs)
	d := NewDeployerWithModeAndRenderer(fs, renderer, "", "")
	if d == nil {
		t.Fatal("NewDeployerWithModeAndRenderer returned nil")
	}

	// Verify it can list templates (should strip .tmpl suffix)
	list := d.ListTemplates()
	if len(list) != 1 || list[0] != "config.sh" {
		t.Errorf("ListTemplates() = %v, want [config.sh]", list)
	}
}

// TestNewDeployerWithModeAndRendererForceUpdate verifies construction with
// renderer and forceUpdate flag.
func TestNewDeployerWithModeAndRendererForceUpdate(t *testing.T) {
	tests := []struct {
		name        string
		forceUpdate bool
	}{
		{"force_enabled", true},
		{"force_disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := fstest.MapFS{
				"file.md": &fstest.MapFile{Data: []byte("content")},
			}
			renderer := NewRenderer(fs)
			d := NewDeployerWithModeAndRendererForceUpdate(fs, renderer, "", "", tt.forceUpdate)
			if d == nil {
				t.Fatal("NewDeployerWithModeAndRendererForceUpdate returned nil")
			}
		})
	}
}

// TestModeAwareDeployer_Deploy tests the Deploy method of modeAwareDeployer.
func TestModeAwareDeployer_Deploy(t *testing.T) {
	t.Run("successful_deployment", func(t *testing.T) {
		root, mgr := setupDeployProject(t)
		fs := fstest.MapFS{
			".claude/agents/moai/backend.md": &fstest.MapFile{
				Data: []byte("# Backend Agent"),
			},
			"CLAUDE.md": &fstest.MapFile{
				Data: []byte("# MoAI Execution Directive"),
			},
		}
		d := NewDeployerWithMode(fs, "", "")

		err := d.Deploy(context.Background(), root, mgr, nil)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		// Verify files exist on disk
		expectedFiles := []string{
			".claude/agents/moai/backend.md",
			"CLAUDE.md",
		}
		for _, f := range expectedFiles {
			absPath := filepath.Join(root, f)
			if _, err := os.Stat(absPath); err != nil {
				t.Errorf("expected file %q to exist: %v", f, err)
			}
		}

		// Verify manifest tracking
		for _, f := range expectedFiles {
			entry, ok := mgr.GetEntry(f)
			if !ok {
				t.Errorf("expected manifest entry for %q", f)
				continue
			}
			if entry.Provenance != manifest.TemplateManaged {
				t.Errorf("entry %q provenance = %v, want %v", f, entry.Provenance, manifest.TemplateManaged)
			}
			if entry.TemplateHash == "" {
				t.Errorf("entry %q has empty TemplateHash", f)
			}
		}
	})

	t.Run("creates_intermediate_directories", func(t *testing.T) {
		root, mgr := setupDeployProject(t)
		fs := fstest.MapFS{
			"deep/nested/dir/file.md": &fstest.MapFile{
				Data: []byte("deeply nested"),
			},
		}
		d := NewDeployerWithMode(fs, "", "")

		err := d.Deploy(context.Background(), root, mgr, nil)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		absPath := filepath.Join(root, "deep", "nested", "dir", "file.md")
		if _, err := os.Stat(absPath); err != nil {
			t.Errorf("nested file should exist: %v", err)
		}
	})

	t.Run("template_rendering", func(t *testing.T) {
		root, mgr := setupDeployProject(t)
		fs := fstest.MapFS{
			"config.txt.tmpl": &fstest.MapFile{
				Data: []byte("project={{.ProjectName}}"),
			},
		}
		renderer := NewRenderer(fs)
		d := NewDeployerWithModeAndRenderer(fs, renderer, "", "")

		tmplCtx := NewTemplateContext(
			WithProject("myproject", root),
		)

		err := d.Deploy(context.Background(), root, mgr, tmplCtx)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		// Verify rendered file exists without .tmpl suffix
		absPath := filepath.Join(root, "config.txt")
		data, err := os.ReadFile(absPath)
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		expected := "project=myproject"
		if string(data) != expected {
			t.Errorf("content = %q, want %q", string(data), expected)
		}

		// Verify .tmpl file does NOT exist
		tmplPath := filepath.Join(root, "config.txt.tmpl")
		if _, err := os.Stat(tmplPath); err == nil {
			t.Error("expected .tmpl file to NOT exist at destination")
		}

		// Verify manifest tracks without .tmpl suffix
		entry, ok := mgr.GetEntry("config.txt")
		if !ok {
			t.Error("expected manifest entry for config.txt")
		} else if entry.Provenance != manifest.TemplateManaged {
			t.Errorf("provenance = %v, want %v", entry.Provenance, manifest.TemplateManaged)
		}
	})

	t.Run("template_without_renderer_copies_raw", func(t *testing.T) {
		root, mgr := setupDeployProject(t)
		rawContent := []byte("project={{.ProjectName}}")
		fs := fstest.MapFS{
			"config.txt.tmpl": &fstest.MapFile{
				Data: rawContent,
			},
		}
		// No renderer - .tmpl files should be copied as-is
		d := NewDeployerWithMode(fs, "", "")

		err := d.Deploy(context.Background(), root, mgr, nil)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		// Without renderer, .tmpl file is kept with original path
		absPath := filepath.Join(root, "config.txt.tmpl")
		data, err := os.ReadFile(absPath)
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		if string(data) != string(rawContent) {
			t.Errorf("content = %q, want raw template content", string(data))
		}
	})

	t.Run("context_cancellation", func(t *testing.T) {
		root, mgr := setupDeployProject(t)
		largeFS := make(fstest.MapFS)
		for i := range 100 {
			name := "files/" + string(rune('a'+i%26)) + string(rune('0'+i/26)) + ".md"
			largeFS[name] = &fstest.MapFile{Data: []byte("content")}
		}
		d := NewDeployerWithMode(largeFS, "", "")

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := d.Deploy(ctx, root, mgr, nil)
		if err == nil {
			t.Fatal("expected error from cancelled context")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got: %v", err)
		}
	})

	t.Run("shell_script_permissions", func(t *testing.T) {
		root, mgr := setupDeployProject(t)
		fs := fstest.MapFS{
			"hooks/run.sh": &fstest.MapFile{
				Data: []byte("#!/bin/bash\necho hello"),
			},
			"docs/readme.md": &fstest.MapFile{
				Data: []byte("# README"),
			},
		}
		d := NewDeployerWithMode(fs, "", "")

		err := d.Deploy(context.Background(), root, mgr, nil)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		// Check shell script has executable permission
		shInfo, err := os.Stat(filepath.Join(root, "hooks", "run.sh"))
		if err != nil {
			t.Fatalf("Stat shell script: %v", err)
		}
		shPerm := shInfo.Mode().Perm()
		if shPerm&0o100 == 0 {
			t.Errorf("shell script permission %o missing owner execute bit", shPerm)
		}

		// Check non-shell file has standard permissions
		mdInfo, err := os.Stat(filepath.Join(root, "docs", "readme.md"))
		if err != nil {
			t.Fatalf("Stat markdown: %v", err)
		}
		mdPerm := mdInfo.Mode().Perm()
		if mdPerm&0o100 != 0 {
			t.Errorf("markdown file permission %o should not have execute bit", mdPerm)
		}
	})

	t.Run("force_update_overwrites_existing", func(t *testing.T) {
		root, mgr := setupDeployProject(t)

		// Pre-create file with old content
		writeFile(t, root, "CLAUDE.md", []byte("old content"))

		fs := fstest.MapFS{
			"CLAUDE.md": &fstest.MapFile{
				Data: []byte("new content"),
			},
		}
		d := NewDeployerWithModeAndRendererForceUpdate(fs, nil, "", "", true)

		err := d.Deploy(context.Background(), root, mgr, nil)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		if string(data) != "new content" {
			t.Errorf("content = %q, want %q (force update should overwrite)", string(data), "new content")
		}
	})

	t.Run("non_force_skips_user_modified_files", func(t *testing.T) {
		root, mgr := setupDeployProject(t)

		// Track file as user_modified in manifest
		writeFile(t, root, "CLAUDE.md", []byte("user modified content"))
		templateHash := manifest.HashBytes([]byte("original"))
		if err := mgr.Track("CLAUDE.md", manifest.UserModified, templateHash); err != nil {
			t.Fatalf("Track error: %v", err)
		}

		fs := fstest.MapFS{
			"CLAUDE.md": &fstest.MapFile{
				Data: []byte("template content"),
			},
		}
		d := NewDeployerWithMode(fs, "", "")

		err := d.Deploy(context.Background(), root, mgr, nil)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		// File should NOT be overwritten
		data, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		if string(data) != "user modified content" {
			t.Errorf("content = %q, want user modified content (should be preserved)", string(data))
		}
	})

	t.Run("non_force_skips_user_created_files", func(t *testing.T) {
		root, mgr := setupDeployProject(t)

		// Track file as user_created in manifest
		writeFile(t, root, "custom.md", []byte("user created"))
		templateHash := manifest.HashBytes([]byte("user created"))
		if err := mgr.Track("custom.md", manifest.UserCreated, templateHash); err != nil {
			t.Fatalf("Track error: %v", err)
		}

		fs := fstest.MapFS{
			"custom.md": &fstest.MapFile{
				Data: []byte("template version"),
			},
		}
		d := NewDeployerWithMode(fs, "", "")

		err := d.Deploy(context.Background(), root, mgr, nil)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		// File should NOT be overwritten
		data, err := os.ReadFile(filepath.Join(root, "custom.md"))
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		if string(data) != "user created" {
			t.Errorf("content = %q, want user created (should be preserved)", string(data))
		}
	})

	t.Run("non_force_skips_untracked_existing_files", func(t *testing.T) {
		root, mgr := setupDeployProject(t)

		// Pre-create file NOT tracked in manifest
		writeFile(t, root, "existing.md", []byte("pre-existing content"))

		fs := fstest.MapFS{
			"existing.md": &fstest.MapFile{
				Data: []byte("template content"),
			},
		}
		d := NewDeployerWithMode(fs, "", "")

		err := d.Deploy(context.Background(), root, mgr, nil)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		// File should NOT be overwritten
		data, err := os.ReadFile(filepath.Join(root, "existing.md"))
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		if string(data) != "pre-existing content" {
			t.Errorf("content = %q, want pre-existing content (untracked should be preserved)", string(data))
		}

		// Verify the file is now tracked as user_created
		entry, ok := mgr.GetEntry("existing.md")
		if !ok {
			t.Error("expected manifest entry for untracked existing file")
		} else if entry.Provenance != manifest.UserCreated {
			t.Errorf("provenance = %v, want %v", entry.Provenance, manifest.UserCreated)
		}
	})

	t.Run("non_force_overwrites_template_managed_files", func(t *testing.T) {
		root, mgr := setupDeployProject(t)

		// Track file as template_managed in manifest
		writeFile(t, root, "managed.md", []byte("old template version"))
		templateHash := manifest.HashBytes([]byte("old template version"))
		if err := mgr.Track("managed.md", manifest.TemplateManaged, templateHash); err != nil {
			t.Fatalf("Track error: %v", err)
		}

		fs := fstest.MapFS{
			"managed.md": &fstest.MapFile{
				Data: []byte("updated template version"),
			},
		}
		d := NewDeployerWithMode(fs, "", "")

		err := d.Deploy(context.Background(), root, mgr, nil)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		// File SHOULD be overwritten (template_managed)
		data, err := os.ReadFile(filepath.Join(root, "managed.md"))
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		if string(data) != "updated template version" {
			t.Errorf("content = %q, want updated template version", string(data))
		}
	})

	t.Run("file_content_matches", func(t *testing.T) {
		root, mgr := setupDeployProject(t)
		expectedContent := []byte("# MoAI Execution Directive v2")
		fs := fstest.MapFS{
			"CLAUDE.md": &fstest.MapFile{Data: expectedContent},
		}
		d := NewDeployerWithMode(fs, "", "")

		if err := d.Deploy(context.Background(), root, mgr, nil); err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		if string(data) != string(expectedContent) {
			t.Errorf("content = %q, want %q", string(data), string(expectedContent))
		}
	})

	t.Run("deploys_to_project_root", func(t *testing.T) {
		root, mgr := setupDeployProject(t)
		fs := fstest.MapFS{
			".claude/agents/moai/expert-backend.md": &fstest.MapFile{
				Data: []byte("# Expert Backend Agent"),
			},
			".claude/skills/moai/SKILL.md": &fstest.MapFile{
				Data: []byte("# MoAI Core Skill"),
			},
		}
		tmplCtx := NewTemplateContext(
			WithProject("testproject", root),
		)
		d := NewDeployerWithMode(fs, "", "")

		err := d.Deploy(context.Background(), root, mgr, tmplCtx)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		// Verify files are in project root subdirectories
		expectedFiles := []string{
			".claude/agents/moai/expert-backend.md",
			".claude/skills/moai/SKILL.md",
		}
		for _, f := range expectedFiles {
			absPath := filepath.Join(root, f)
			if _, err := os.Stat(absPath); err != nil {
				t.Errorf("expected file %q in project: %v", f, err)
			}
		}
	})

	t.Run("rendered_template_with_shell_script", func(t *testing.T) {
		root, mgr := setupDeployProject(t)
		fs := fstest.MapFS{
			"hooks/setup.sh.tmpl": &fstest.MapFile{
				Data: []byte("#!/bin/bash\necho {{.ProjectName}}"),
			},
		}
		renderer := NewRenderer(fs)
		d := NewDeployerWithModeAndRenderer(fs, renderer, "", "")

		tmplCtx := NewTemplateContext(
			WithProject("testproj", root),
		)

		err := d.Deploy(context.Background(), root, mgr, tmplCtx)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		// Rendered file should be .sh (without .tmpl)
		absPath := filepath.Join(root, "hooks", "setup.sh")
		data, err := os.ReadFile(absPath)
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		if string(data) != "#!/bin/bash\necho testproj" {
			t.Errorf("content = %q, want rendered shell script", string(data))
		}

		// Shell scripts should have executable permission
		info, err := os.Stat(absPath)
		if err != nil {
			t.Fatalf("Stat error: %v", err)
		}
		if info.Mode().Perm()&0o100 == 0 {
			t.Errorf("rendered .sh file should have executable permission, got %o", info.Mode().Perm())
		}
	})
}

// TestModeAwareDeployer_ExtractTemplate tests the ExtractTemplate method.
func TestModeAwareDeployer_ExtractTemplate(t *testing.T) {
	t.Run("existing_template", func(t *testing.T) {
		fs := fstest.MapFS{
			"CLAUDE.md": &fstest.MapFile{
				Data: []byte("# MoAI Directive"),
			},
		}
		d := NewDeployerWithMode(fs, "", "")

		data, err := d.ExtractTemplate("CLAUDE.md")
		if err != nil {
			t.Fatalf("ExtractTemplate error: %v", err)
		}
		if string(data) != "# MoAI Directive" {
			t.Errorf("content = %q, want %q", string(data), "# MoAI Directive")
		}
	})

	t.Run("nested_template", func(t *testing.T) {
		fs := fstest.MapFS{
			".claude/agents/moai/backend.md": &fstest.MapFile{
				Data: []byte("# Backend"),
			},
		}
		d := NewDeployerWithMode(fs, "", "")

		data, err := d.ExtractTemplate(".claude/agents/moai/backend.md")
		if err != nil {
			t.Fatalf("ExtractTemplate error: %v", err)
		}
		if len(data) == 0 {
			t.Error("expected non-empty content for nested template")
		}
	})

	t.Run("nonexistent_template", func(t *testing.T) {
		fs := fstest.MapFS{}
		d := NewDeployerWithMode(fs, "", "")

		data, err := d.ExtractTemplate("nonexistent.txt")
		if err == nil {
			t.Fatal("expected error for nonexistent template")
		}
		if !errors.Is(err, ErrTemplateNotFound) {
			t.Errorf("expected ErrTemplateNotFound, got: %v", err)
		}
		if data != nil {
			t.Errorf("expected nil data, got %d bytes", len(data))
		}
	})
}

// TestModeAwareDeployer_ListTemplates tests the ListTemplates method.
func TestModeAwareDeployer_ListTemplates(t *testing.T) {
	t.Run("returns_all_files", func(t *testing.T) {
		fs := fstest.MapFS{
			".claude/settings.json": &fstest.MapFile{
				Data: []byte(`{"hooks":{}}`),
			},
			".claude/agents/moai/backend.md": &fstest.MapFile{
				Data: []byte("# Backend"),
			},
			"CLAUDE.md": &fstest.MapFile{
				Data: []byte("# MoAI"),
			},
		}
		d := NewDeployerWithMode(fs, "", "")
		list := d.ListTemplates()

		if len(list) != 3 {
			t.Fatalf("ListTemplates() returned %d items, want 3", len(list))
		}

		expected := map[string]bool{
			".claude/settings.json":              true,
			".claude/agents/moai/backend.md":     true,
			"CLAUDE.md":                          true,
		}
		for _, item := range list {
			if !expected[item] {
				t.Errorf("unexpected template: %q", item)
			}
		}
	})

	t.Run("strips_tmpl_suffix", func(t *testing.T) {
		fs := fstest.MapFS{
			"config.yaml.tmpl": &fstest.MapFile{
				Data: []byte("key: {{.Value}}"),
			},
			"hooks/run.sh.tmpl": &fstest.MapFile{
				Data: []byte("#!/bin/bash"),
			},
			"README.md": &fstest.MapFile{
				Data: []byte("# README"),
			},
		}
		d := NewDeployerWithMode(fs, "", "")
		list := d.ListTemplates()

		if len(list) != 3 {
			t.Fatalf("ListTemplates() returned %d items, want 3", len(list))
		}

		expected := map[string]bool{
			"config.yaml":  true,
			"hooks/run.sh": true,
			"README.md":    true,
		}
		for _, item := range list {
			if !expected[item] {
				t.Errorf("unexpected template path: %q (should strip .tmpl suffix)", item)
			}
		}
	})

	t.Run("empty_fs", func(t *testing.T) {
		d := NewDeployerWithMode(fstest.MapFS{}, "", "")
		list := d.ListTemplates()
		if len(list) != 0 {
			t.Errorf("expected 0 templates from empty FS, got %d", len(list))
		}
	})
}
