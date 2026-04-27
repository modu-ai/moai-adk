package template

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// TestDeployer_TransactionalValidation verifies that template validation happens
// BEFORE any files are written to disk. This prevents leaving the project in a
// broken state if a template fails to parse (issue #733).
func TestDeployer_TransactionalValidation(t *testing.T) {
	t.Run("validation_happens_before_file_writes", func(t *testing.T) {
		// Create a filesystem with one valid template and one invalid template
		fs := fstest.MapFS{
			"valid.tmpl": &fstest.MapFile{
				Data: []byte("Hello {{.ProjectName}}"),
			},
			"invalid.tmpl": &fstest.MapFile{
				// This template has a syntax error: unescaped }}
				Data: []byte("Broken {{ } } template"),
			},
		}

		r := NewRenderer(fs)
		d := NewDeployerWithRenderer(fs, r)

		// Create a temp directory for deployment
		tempDir := t.TempDir()
		mgr := manifest.NewManager()

		ctx := context.Background()
		tmplCtx := NewTemplateContext(WithProject("Test", tempDir))

		// Deploy should fail during validation
		err := d.Deploy(ctx, tempDir, mgr, tmplCtx)
		if err == nil {
			t.Error("expected error for invalid template, got nil")
		}

		// CRITICAL: No files should have been written to disk
		// because validation failed BEFORE the write phase
		entries, _ := os.ReadDir(tempDir)
		if len(entries) > 0 {
			t.Errorf("expected no files written after validation failure, got %d files", len(entries))
			for _, e := range entries {
				t.Logf("  - %s", e.Name())
			}
		}
	})

	t.Run("valid_templates_write_after_validation", func(t *testing.T) {
		// Create a filesystem with valid templates only
		fs := fstest.MapFS{
			"file1.txt.tmpl": &fstest.MapFile{
				Data: []byte("Content 1: {{.Version}}"),
			},
			"file2.txt.tmpl": &fstest.MapFile{
				Data: []byte("Content 2: {{.Version}}"),
			},
		}

		r := NewRenderer(fs)
		d := NewDeployerWithRenderer(fs, r)

		tempDir := t.TempDir()
		mgr := manifest.NewManager()
		// Initialize manifest in temp directory
		_, _ = mgr.Load(tempDir)

		ctx := context.Background()
		tmplCtx := NewTemplateContext(WithVersion("test-value"))

		// Deploy should succeed
		err := d.Deploy(ctx, tempDir, mgr, tmplCtx)
		if err != nil {
			t.Fatalf("expected no error for valid templates, got: %v", err)
		}

		// Files should have been written
		file1 := filepath.Join(tempDir, "file1.txt")
		file2 := filepath.Join(tempDir, "file2.txt")

		if _, err := os.Stat(file1); os.IsNotExist(err) {
			t.Error("file1.txt should have been written")
		}
		if _, err := os.Stat(file2); os.IsNotExist(err) {
			t.Error("file2.txt should have been written")
		}

		// Verify content
		content1, _ := os.ReadFile(file1)
		if string(content1) != "Content 1: test-value" {
			t.Errorf("file1.txt content = %q, want %q", string(content1), "Content 1: test-value")
		}
	})
}

// TestDeployer_ValidateAllTemplates verifies a new method that validates
// all templates without writing any files.
func TestDeployer_ValidateAllTemplates(t *testing.T) {
	t.Run("all_templates_valid", func(t *testing.T) {
		fs := fstest.MapFS{
			"template1.tmpl": &fstest.MapFile{
				Data: []byte("Value: {{.Version}}"),
			},
			"template2.tmpl": &fstest.MapFile{
				Data: []byte("Name: {{.ProjectName}}"),
			},
		}

		r := NewRenderer(fs)
		d := NewDeployerWithRenderer(fs, r)

		ctx := context.Background()
		tmplCtx := NewTemplateContext(WithVersion("test"), WithProject("Goos", "/tmp/goos"))

		// Validate should succeed
		err := d.ValidateAll(ctx, tmplCtx)
		if err != nil {
			t.Errorf("expected no validation error, got: %v", err)
		}
	})

	t.Run("invalid_template_fails_validation", func(t *testing.T) {
		fs := fstest.MapFS{
			"valid.tmpl": &fstest.MapFile{
				Data: []byte("OK: {{.Version}}"),
			},
			"invalid.tmpl": &fstest.MapFile{
				// Syntax error: missing closing brace
				Data: []byte("Broken: {{.Version"),
			},
		}

		r := NewRenderer(fs)
		d := NewDeployerWithRenderer(fs, r)

		ctx := context.Background()
		tmplCtx := NewTemplateContext(WithVersion("test"))

		// Validate should fail
		err := d.ValidateAll(ctx, tmplCtx)
		if err == nil {
			t.Error("expected validation error for invalid template, got nil")
		}
	})
}
