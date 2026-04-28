package template

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// @MX:ANCHOR: [AUTO] Deployer is the core interface that extracts templates from embedded filesystem and deploys to project root. Each file is tracked in manifest.
// @MX:REASON: fan_in=8+, entry point for all project initialization, core contract for template deployment
// Deployer extracts and deploys templates from an embedded filesystem
// to a project root directory, tracking each file in the manifest.
type Deployer interface {
	// Deploy extracts all templates and writes them to projectRoot,
	// registering each file with the manifest manager.
	// If tmplCtx is provided and a Renderer is configured, files ending in
	// .tmpl are rendered with the context and saved without the .tmpl suffix.
	Deploy(ctx context.Context, projectRoot string, m manifest.Manager, tmplCtx *TemplateContext) error

	// ValidateAll validates all templates without writing any files.
	// This is used to implement transactional deployment: validate first,
	// then deploy. If any template fails to parse or render, the error is
	// returned immediately and no files are written.
	ValidateAll(ctx context.Context, tmplCtx *TemplateContext) error

	// ExtractTemplate returns the raw content of a single template by name.
	ExtractTemplate(name string) ([]byte, error)

	// ListTemplates returns the relative paths of all embedded templates.
	ListTemplates() []string
}

// deployer is the concrete implementation of Deployer.
type deployer struct {
	fsys        fs.FS
	renderer    Renderer // Optional: if set, .tmpl files are rendered with TemplateContext
	forceUpdate bool     // If true, overwrite existing files without manifest check (used for updates)
}

// NewDeployer creates a Deployer backed by the given filesystem.
// In production the fs.FS comes from go:embed; in tests use testing/fstest.MapFS.
func NewDeployer(fsys fs.FS) Deployer {
	return &deployer{fsys: fsys}
}

// NewDeployerWithRenderer creates a Deployer that renders .tmpl files using the given Renderer.
func NewDeployerWithRenderer(fsys fs.FS, renderer Renderer) Deployer {
	return &deployer{fsys: fsys, renderer: renderer, forceUpdate: false}
}

// NewDeployerWithForceUpdate creates a Deployer that forces overwrite of existing files.
// This is used for template updates where template files should replace existing versions.
func NewDeployerWithForceUpdate(fsys fs.FS, forceUpdate bool) Deployer {
	return &deployer{fsys: fsys, forceUpdate: forceUpdate}
}

// NewDeployerWithRendererAndForceUpdate creates a Deployer that renders .tmpl files
// and forces overwrite of existing files. Used for template updates with rendering.
func NewDeployerWithRendererAndForceUpdate(fsys fs.FS, renderer Renderer, forceUpdate bool) Deployer {
	return &deployer{fsys: fsys, renderer: renderer, forceUpdate: forceUpdate}
}

// @MX:NOTE: [AUTO] Checks context cancellation for per-file abort capability. Files with .tmpl suffix are rendered via Renderer and saved without the suffix.
// Deploy walks the embedded filesystem and writes every file to projectRoot.
// Files ending in .tmpl are rendered using the Renderer (if configured) and
// saved without the .tmpl suffix.
func (d *deployer) Deploy(ctx context.Context, projectRoot string, m manifest.Manager, tmplCtx *TemplateContext) error {
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

		// Determine if this is a template file that needs rendering
		isTemplate := strings.HasSuffix(path, ".tmpl")
		var content []byte
		var destRelPath string

		if isTemplate && d.renderer != nil && tmplCtx != nil {
			// Render the template
			rendered, renderErr := d.renderer.Render(path, tmplCtx)
			if renderErr != nil {
				return fmt.Errorf("template render %q: %w", path, renderErr)
			}
			content = rendered
			// Remove .tmpl suffix for destination path
			destRelPath = strings.TrimSuffix(path, ".tmpl")
		} else {
			// Read raw content
			rawContent, readErr := fs.ReadFile(d.fsys, path)
			if readErr != nil {
				return fmt.Errorf("template deploy read %q: %w", path, readErr)
			}
			content = rawContent
			destRelPath = path
		}

		// Compute destination path
		destPath := filepath.Join(projectRoot, filepath.FromSlash(destRelPath))

		// Existing file protection: skip files that already exist at the
		// destination. This prevents overwriting user-created or
		// programmatically-generated files (e.g., config YAMLs from Step 2
		// of init, or pre-existing CLAUDE.md).
		// Skip this check in forceUpdate mode (used for template updates).
		if !d.forceUpdate {
			if _, statErr := os.Stat(destPath); statErr == nil {
				// File exists — check manifest for provenance
				if entry, found := m.GetEntry(destRelPath); found {
					if entry.Provenance == manifest.UserModified || entry.Provenance == manifest.UserCreated {
						// Respect user files
						return nil
					}
					// template_managed files are safe to overwrite (re-init / update)
				} else {
					// Existing file not tracked in manifest — record as user_created and skip
					templateHash := manifest.HashBytes(content)
					_ = m.Track(destRelPath, manifest.UserCreated, templateHash)
					return nil
				}
			}
		}

		// Create parent directories
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return fmt.Errorf("template deploy mkdir %q: %w", destDir, err)
		}

		// Determine file permissions based on extension
		// Shell scripts and other executable files need executable bit
		perm := fs.FileMode(0o644) // Default: read/write for owner, read for others
		if strings.HasSuffix(destRelPath, ".sh") {
			perm = 0o755 // Executable: read/write/execute for owner, read/execute for others
		}

		// Write file
		if err := os.WriteFile(destPath, content, perm); err != nil {
			return fmt.Errorf("template deploy write %q: %w", destPath, err)
		}

		// Track in manifest (use destRelPath, not original path with .tmpl)
		templateHash := manifest.HashBytes(content)
		if err := m.Track(destRelPath, manifest.TemplateManaged, templateHash); err != nil {
			return fmt.Errorf("template deploy track %q: %w", destRelPath, err)
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
		// Strip .tmpl suffix to return deployment target paths
		targetPath := path
		if before, ok := strings.CutSuffix(path, ".tmpl"); ok {
			targetPath = before
		}
		list = append(list, targetPath)
		return nil
	})

	return list
}

// @MX:NOTE: [AUTO] Implements transactional validation - renders all templates in memory without writing to disk
// ValidateAll validates all templates without writing any files.
// This implements transactional deployment: validate first, then deploy.
// If any template fails to parse or render, the error is returned immediately.
func (d *deployer) ValidateAll(ctx context.Context, tmplCtx *TemplateContext) error {
	// Only validate if we have a renderer configured
	if d.renderer == nil {
		return nil
	}

	var validationErrors []error
	walkErr := fs.WalkDir(d.fsys, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check context cancellation during validation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip directories and non-templates
		if path == "." || entry.IsDir() || !strings.HasSuffix(path, ".tmpl") {
			return nil
		}

		// Try to render the template (this will catch parse errors)
		_, renderErr := d.renderer.Render(path, tmplCtx)
		if renderErr != nil {
			validationErrors = append(validationErrors,
				fmt.Errorf("template %q: %w", path, renderErr))
		}

		return nil
	})

	if walkErr != nil {
		return walkErr
	}

	if len(validationErrors) > 0 {
		// Return the first validation error
		return validationErrors[0]
	}

	return nil
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

	// Convert projectRoot to absolute path for reliable comparison
	absProjectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return fmt.Errorf("resolve project root: %w", err)
	}

	// Verify containment: the resolved path must be under projectRoot
	absPath := filepath.Join(absProjectRoot, cleaned)
	if !strings.HasPrefix(absPath, absProjectRoot+string(filepath.Separator)) && absPath != absProjectRoot {
		return fmt.Errorf("%w: %q escapes project root", ErrPathTraversal, relPath)
	}

	return nil
}
