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

func testFS() fstest.MapFS {
	return fstest.MapFS{
		".claude/settings.json": &fstest.MapFile{
			Data: []byte(`{"hooks":{}}`),
		},
		".claude/agents/moai/expert-backend.md": &fstest.MapFile{
			Data: []byte("# Expert Backend Agent"),
		},
		"CLAUDE.md": &fstest.MapFile{
			Data: []byte("# MoAI Execution Directive"),
		},
		".gitignore": &fstest.MapFile{
			Data: []byte("node_modules/\n.env\n"),
		},
	}
}

func setupDeployProject(t *testing.T) (string, manifest.Manager) {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("manifest Load error: %v", err)
	}
	return root, mgr
}

func TestDeployerDeploy(t *testing.T) {
	t.Run("successful_deployment", func(t *testing.T) {
		root, mgr := setupDeployProject(t)
		d := NewDeployer(testFS())

		err := d.Deploy(context.Background(), root, mgr, nil)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		// Verify all files exist on disk
		expectedFiles := []string{
			".claude/settings.json",
			".claude/agents/moai/expert-backend.md",
			"CLAUDE.md",
			".gitignore",
		}
		for _, f := range expectedFiles {
			absPath := filepath.Join(root, f)
			if _, err := os.Stat(absPath); err != nil {
				t.Errorf("expected file %q to exist: %v", f, err)
			}
		}

		// Verify files tracked in manifest
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
				Data: []byte("nested content"),
			},
		}
		d := NewDeployer(fs)

		err := d.Deploy(context.Background(), root, mgr, nil)
		if err != nil {
			t.Fatalf("Deploy error: %v", err)
		}

		absPath := filepath.Join(root, "deep", "nested", "dir", "file.md")
		if _, err := os.Stat(absPath); err != nil {
			t.Errorf("nested file should exist: %v", err)
		}
	})

	t.Run("context_cancellation", func(t *testing.T) {
		root, mgr := setupDeployProject(t)

		// Create a large FS to ensure we hit the cancellation
		largeFS := make(fstest.MapFS)
		for i := range 100 {
			name := filepath.Join("files", filepath.Base(filepath.Join("dir", string(rune('a'+i%26))+".md")))
			largeFS[name] = &fstest.MapFile{Data: []byte("content")}
		}

		d := NewDeployer(largeFS)
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

	t.Run("file_content_matches", func(t *testing.T) {
		root, mgr := setupDeployProject(t)
		expectedContent := []byte("# MoAI Execution Directive")
		fs := fstest.MapFS{
			"CLAUDE.md": &fstest.MapFile{Data: expectedContent},
		}
		d := NewDeployer(fs)

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
}

func TestDeployerExtractTemplate(t *testing.T) {
	t.Run("existing_template", func(t *testing.T) {
		d := NewDeployer(testFS())

		data, err := d.ExtractTemplate("CLAUDE.md")
		if err != nil {
			t.Fatalf("ExtractTemplate error: %v", err)
		}
		if len(data) == 0 {
			t.Error("expected non-empty content")
		}
		if string(data) != "# MoAI Execution Directive" {
			t.Errorf("content = %q, want %q", string(data), "# MoAI Execution Directive")
		}
	})

	t.Run("nonexistent_template", func(t *testing.T) {
		d := NewDeployer(testFS())

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

func TestDeployerListTemplates(t *testing.T) {
	t.Run("returns_all_files", func(t *testing.T) {
		d := NewDeployer(testFS())
		list := d.ListTemplates()

		if len(list) != 4 {
			t.Fatalf("ListTemplates() returned %d items, want 4", len(list))
		}

		expected := map[string]bool{
			".claude/settings.json":                 true,
			".claude/agents/moai/expert-backend.md": true,
			"CLAUDE.md":                             true,
			".gitignore":                            true,
		}
		for _, item := range list {
			if !expected[item] {
				t.Errorf("unexpected template: %q", item)
			}
		}
	})

	t.Run("empty_fs", func(t *testing.T) {
		d := NewDeployer(fstest.MapFS{})
		list := d.ListTemplates()
		if len(list) != 0 {
			t.Errorf("expected 0 templates from empty FS, got %d", len(list))
		}
	})
}

func TestValidateDeployPath(t *testing.T) {
	// Use t.TempDir() to get a real directory path on the current platform
	root := t.TempDir()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid_relative", ".claude/settings.json", false},
		{"valid_nested", ".claude/agents/moai/file.md", false},
		{"valid_simple", "CLAUDE.md", false},
		{"traversal_dotdot", "../etc/passwd", true},
		{"traversal_nested", "foo/../../etc/passwd", true},
		{"traversal_complex", ".claude/./../../secret", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDeployPath(root, tt.path)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for path %q", tt.path)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for path %q: %v", tt.path, err)
			}
			if tt.wantErr && err != nil && !errors.Is(err, ErrPathTraversal) {
				t.Errorf("expected ErrPathTraversal, got: %v", err)
			}
		})
	}

	// Test absolute paths separately (platform-dependent)
	t.Run("absolute_path", func(t *testing.T) {
		absPath := filepath.Join(root, "absolute")
		err := validateDeployPath(root, absPath)
		if err == nil {
			t.Errorf("expected error for absolute path %q", absPath)
		}
		if err != nil && !errors.Is(err, ErrPathTraversal) {
			t.Errorf("expected ErrPathTraversal, got: %v", err)
		}
	})
}
