package template

import (
	"bytes"
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

// --- M1: Atomic Write + Idempotency Tests (SPEC-V3R3-UPDATE-CLEANUP-001) ---

// TestDeployer_Idempotent verifies that two consecutive Deploy calls to the same
// destination produce byte-identical files with no auxiliary artefacts.
func TestDeployer_Idempotent(t *testing.T) {
	t.Parallel()
	root, mgr := setupDeployProject(t)
	d := NewDeployerWithForceUpdate(testFS(), true)

	ctx := context.Background()

	// First deploy
	if err := d.Deploy(ctx, root, mgr, nil); err != nil {
		t.Fatalf("first Deploy error: %v", err)
	}

	// Snapshot file contents after first deploy
	contents1 := snapshotFileContents(t, root)

	// Second deploy (idempotent)
	if err := d.Deploy(ctx, root, mgr, nil); err != nil {
		t.Fatalf("second Deploy error: %v", err)
	}

	// Snapshot after second deploy
	contents2 := snapshotFileContents(t, root)

	// Verify byte-identical
	for path, data1 := range contents1 {
		data2, ok := contents2[path]
		if !ok {
			t.Errorf("file %q disappeared after second Deploy", path)
			continue
		}
		if !bytes.Equal(data1, data2) {
			t.Errorf("file %q content changed after second Deploy", path)
		}
	}

	// Verify no .moai-tmp residue
	assertNoTmpResidue(t, root)

	// Verify no " 2" suffix files
	assertNoSuffix2Files(t, root)
}

// TestDeployer_NoTmpResidue verifies that no .moai-tmp files are left after
// a successful Deploy.
func TestDeployer_NoTmpResidue(t *testing.T) {
	t.Parallel()
	root, mgr := setupDeployProject(t)
	d := NewDeployerWithForceUpdate(testFS(), true)

	if err := d.Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("Deploy error: %v", err)
	}

	assertNoTmpResidue(t, root)
}

// TestDeployer_TmpCleanupOnFailure verifies that if the atomic rename fails,
// the .moai-tmp file is cleaned up and an error is returned.
// This test uses atomicWriteFile directly to simulate a rename failure.
func TestDeployer_TmpCleanupOnFailure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	destPath := filepath.Join(root, "target.txt")
	content := []byte("hello world")

	// Create a directory at the tmp path location to force rename failure.
	// atomicWriteFile writes to destPath+".moai-tmp"; make that path a directory.
	tmpPath := destPath + ".moai-tmp"
	if err := os.MkdirAll(tmpPath, 0o755); err != nil {
		t.Fatalf("setup: MkdirAll %q: %v", tmpPath, err)
	}

	// atomicWriteFile should fail and clean up
	err := atomicWriteFile(destPath, content, 0o644)
	if err == nil {
		t.Fatal("expected error when rename target is a directory, got nil")
	}

	// The tmp path should be cleaned up (removed) on failure.
	// Since tmpPath was a directory we created, check that atomicWriteFile
	// attempted cleanup (it may or may not succeed on a dir, but should not
	// leave a regular tmp file).
	// Key invariant: no .moai-tmp *file* residue (directories we set up, not atomicWriteFile).
	entries, _ := os.ReadDir(root)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".moai-tmp" && !e.IsDir() {
			t.Errorf("tmp file residue found: %s", e.Name())
		}
	}
}

// snapshotFileContents walks root and returns a map of relpath → content for
// all regular files.
func snapshotFileContents(t *testing.T, root string) map[string][]byte {
	t.Helper()
	result := make(map[string][]byte)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		result[rel] = data
		return nil
	})
	if err != nil {
		t.Fatalf("snapshotFileContents walk error: %v", err)
	}
	return result
}

// assertNoTmpResidue fails if any .moai-tmp file exists under root.
func assertNoTmpResidue(t *testing.T, root string) {
	t.Helper()
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".moai-tmp" {
			t.Errorf("unexpected .moai-tmp residue: %s", path)
		}
		return nil
	})
}

// assertNoSuffix2Files fails if any file with a " 2" suffix exists under root.
func assertNoSuffix2Files(t *testing.T, root string) {
	t.Helper()
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := filepath.Base(path)
		// Check for patterns like "file 2.ext" or "file 2"
		for _, ext := range []string{"", ".md", ".yaml", ".json", ".sh", ".txt"} {
			if len(name) > 2+len(ext) {
				suffix := " 2" + ext
				if len(name) >= len(suffix) && name[len(name)-len(suffix):] == suffix {
					t.Errorf("unexpected \" 2\" suffix file: %s", path)
				}
			}
		}
		return nil
	})
}
