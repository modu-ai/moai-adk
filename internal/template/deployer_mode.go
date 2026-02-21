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

// modeAwareDeployer wraps a deployer with optional force update support.
type modeAwareDeployer struct {
	fsys        fs.FS
	renderer    Renderer
	forceUpdate bool
}

// NewDeployerWithMode creates a Deployer for local project deployment.
func NewDeployerWithMode(fsys fs.FS, _, _ string) Deployer {
	return &modeAwareDeployer{
		fsys: fsys,
	}
}

// NewDeployerWithModeAndRenderer creates a deployer with template rendering.
func NewDeployerWithModeAndRenderer(fsys fs.FS, renderer Renderer, _, _ string) Deployer {
	return &modeAwareDeployer{
		fsys:     fsys,
		renderer: renderer,
	}
}

// NewDeployerWithModeAndRendererForceUpdate creates a deployer with rendering
// and force update capability. Used for template updates where files should be overwritten.
func NewDeployerWithModeAndRendererForceUpdate(fsys fs.FS, renderer Renderer, _, _ string, forceUpdate bool) Deployer {
	return &modeAwareDeployer{
		fsys:        fsys,
		renderer:    renderer,
		forceUpdate: forceUpdate,
	}
}

// Deploy implements Deployer interface - all files go to project root.
func (d *modeAwareDeployer) Deploy(ctx context.Context, projectRoot string, m manifest.Manager, tmplCtx *TemplateContext) error {
	projectRoot = filepath.Clean(projectRoot)

	var deployErr error
	walkErr := fs.WalkDir(d.fsys, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip root and directories
		if path == "." || entry.IsDir() {
			return nil
		}

		// Determine if template file needs rendering
		isTemplate := strings.HasSuffix(path, ".tmpl")
		var content []byte
		var destRelPath string

		if isTemplate && d.renderer != nil && tmplCtx != nil {
			rendered, renderErr := d.renderer.Render(path, tmplCtx)
			if renderErr != nil {
				return fmt.Errorf("template render %q: %w", path, renderErr)
			}
			content = rendered
			destRelPath = strings.TrimSuffix(path, ".tmpl")
		} else {
			rawContent, readErr := fs.ReadFile(d.fsys, path)
			if readErr != nil {
				return fmt.Errorf("template deploy read %q: %w", path, readErr)
			}
			content = rawContent
			destRelPath = path
		}

		// All files go to project root
		destPath := filepath.Join(projectRoot, filepath.FromSlash(destRelPath))

		// Validate path security
		if err := validateDeployPath(projectRoot, destRelPath); err != nil {
			deployErr = err
			return err
		}

		// Skip existing files unless forceUpdate
		if !d.forceUpdate {
			if _, statErr := os.Stat(destPath); statErr == nil {
				// File exists - check manifest for provenance
				if entry, found := m.GetEntry(destRelPath); found {
					if entry.Provenance == manifest.UserModified || entry.Provenance == manifest.UserCreated {
						return nil // Respect user files
					}
				} else {
					// Existing file not tracked - record as user_created and skip
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

		// Determine file permissions
		perm := fs.FileMode(0o644)
		if strings.HasSuffix(destRelPath, ".sh") {
			perm = 0o755
		}

		// Write file
		if err := os.WriteFile(destPath, content, perm); err != nil {
			return fmt.Errorf("template deploy write %q: %w", destPath, err)
		}

		// Track in manifest
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
func (d *modeAwareDeployer) ExtractTemplate(name string) ([]byte, error) {
	data, err := fs.ReadFile(d.fsys, name)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTemplateNotFound, name)
	}
	return data, nil
}

// ListTemplates returns sorted relative paths of all files in the embedded FS.
func (d *modeAwareDeployer) ListTemplates() []string {
	var list []string

	_ = fs.WalkDir(d.fsys, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == "." || entry.IsDir() {
			return nil
		}
		targetPath := path
		if before, ok := strings.CutSuffix(path, ".tmpl"); ok {
			targetPath = before
		}
		list = append(list, targetPath)
		return nil
	})

	return list
}
