package template

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk-go/internal/manifest"
)

// Deployer extracts and deploys templates from an embedded filesystem
// to a project root directory, tracking each file in the manifest.
type Deployer interface {
	// Deploy extracts all templates and writes them to projectRoot,
	// registering each file with the manifest manager.
	Deploy(ctx context.Context, projectRoot string, m manifest.Manager) error

	// ExtractTemplate returns the raw content of a single template by name.
	ExtractTemplate(name string) ([]byte, error)

	// ListTemplates returns the relative paths of all embedded templates.
	ListTemplates() []string
}

// deployer is the concrete implementation of Deployer.
type deployer struct {
	fsys fs.FS
}

// NewDeployer creates a Deployer backed by the given filesystem.
// In production the fs.FS comes from go:embed; in tests use testing/fstest.MapFS.
func NewDeployer(fsys fs.FS) Deployer {
	return &deployer{fsys: fsys}
}

// Deploy walks the embedded filesystem and writes every file to projectRoot.
func (d *deployer) Deploy(ctx context.Context, projectRoot string, m manifest.Manager) error {
	projectRoot = filepath.Clean(projectRoot)

	var deployErr error
	walkErr := fs.WalkDir(d.fsys, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check context cancellation before each file
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip the root directory itself
		if path == "." {
			return nil
		}

		// Skip directories (they are created on demand)
		if entry.IsDir() {
			return nil
		}

		// Validate path security
		if err := validateDeployPath(projectRoot, path); err != nil {
			deployErr = err
			return err
		}

		// Read template content
		content, err := fs.ReadFile(d.fsys, path)
		if err != nil {
			return fmt.Errorf("template deploy read %q: %w", path, err)
		}

		// Compute destination path
		destPath := filepath.Join(projectRoot, filepath.FromSlash(path))

		// Create parent directories
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return fmt.Errorf("template deploy mkdir %q: %w", destDir, err)
		}

		// Write file
		if err := os.WriteFile(destPath, content, 0o644); err != nil {
			return fmt.Errorf("template deploy write %q: %w", destPath, err)
		}

		// Track in manifest
		templateHash := manifest.HashBytes(content)
		if err := m.Track(path, manifest.TemplateManaged, templateHash); err != nil {
			return fmt.Errorf("template deploy track %q: %w", path, err)
		}

		return nil
	})

	if walkErr != nil {
		return walkErr
	}
	return deployErr
}

// ExtractTemplate returns the content of a single named template.
func (d *deployer) ExtractTemplate(name string) ([]byte, error) {
	data, err := fs.ReadFile(d.fsys, name)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTemplateNotFound, name)
	}
	return data, nil
}

// ListTemplates returns sorted relative paths of all files in the embedded FS.
func (d *deployer) ListTemplates() []string {
	var list []string

	_ = fs.WalkDir(d.fsys, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors during listing
		}
		if path == "." || entry.IsDir() {
			return nil
		}
		list = append(list, path)
		return nil
	})

	return list
}

// validateDeployPath ensures a template path does not escape projectRoot.
func validateDeployPath(projectRoot, relPath string) error {
	// Clean and normalize
	cleaned := filepath.Clean(filepath.FromSlash(relPath))

	// Reject absolute paths
	if filepath.IsAbs(cleaned) {
		return fmt.Errorf("%w: absolute path %q", ErrPathTraversal, relPath)
	}

	// Reject path traversal components
	if strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, string(filepath.Separator)+"..") {
		return fmt.Errorf("%w: parent reference in %q", ErrPathTraversal, relPath)
	}

	// Verify containment: the resolved path must be under projectRoot
	absPath := filepath.Join(projectRoot, cleaned)
	if !strings.HasPrefix(absPath, projectRoot+string(filepath.Separator)) && absPath != projectRoot {
		return fmt.Errorf("%w: %q escapes project root", ErrPathTraversal, relPath)
	}

	return nil
}
