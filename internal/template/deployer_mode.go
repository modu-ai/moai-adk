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

// @MX:NOTE: [AUTO] IsGlobalFile은 Global 모드에서 ~/.claude/에 배치될 파일을 결정합니다. agents/moai, skills/moai*, rules/moai는 글로벌, hooks와 settings는 로컬에 유지됩니다.
// IsGlobalFile determines if a template file should be deployed to global ~/.claude/
// when in global mode. Returns true for agents/moai/, skills/moai*, rules/moai/.
// Returns false for hooks/, settings files, and other project-specific files.
func IsGlobalFile(relPath string) bool {
	// Normalize path separators
	relPath = filepath.ToSlash(relPath)

	// Files that are ALWAYS local (project-specific)
	// - hooks/moai/ - project-specific hooks
	// - settings.json, settings.local.json - project/team settings
	// - .moai/ - project config
	if strings.HasPrefix(relPath, ".claude/hooks/") {
		return false
	}
	if relPath == ".claude/settings.json" || relPath == ".claude/settings.local.json" {
		return false
	}
	if strings.HasPrefix(relPath, ".moai/") {
		return false
	}

	// Files that go to global in global mode
	// - agents/moai/ - MoAI system agents
	// - skills/moai* - MoAI core and related skills
	// - rules/moai/ - MoAI system rules
	if strings.HasPrefix(relPath, ".claude/agents/moai/") {
		return true
	}
	if strings.HasPrefix(relPath, ".claude/skills/moai") {
		return true
	}
	if strings.HasPrefix(relPath, ".claude/rules/moai/") {
		return true
	}

	// All other files are local
	return false
}

// @MX:NOTE: [AUTO] ResolveDeployPath는 설치 모드에 따라 파일의 최종 배치 경로를 결정합니다. Global 모드에서는 agents/moai, skills/moai*, rules/moai를 ~/.claude/로 리다이렉트합니다.
// ResolveDeployPath determines the destination path for a template file based on
// installation mode. In global mode, system files go to ~/.claude/ instead of
// the project directory.
func ResolveDeployPath(relPath, mode, projectRoot, homeDir string) string {
	relPath = filepath.ToSlash(relPath)

	// In local mode, everything goes to project root
	if mode != "global" {
		return filepath.Join(projectRoot, filepath.FromSlash(relPath))
	}

	// In global mode, check if this is a global file
	if IsGlobalFile(relPath) {
		// Remove .claude/ prefix and place directly in ~/.claude/
		subPath := strings.TrimPrefix(relPath, ".claude/")
		return filepath.Join(homeDir, ".claude", filepath.FromSlash(subPath))
	}

	// Local files stay in project root
	return filepath.Join(projectRoot, filepath.FromSlash(relPath))
}

// modeAwareDeployer wraps a deployer with installation mode support.
type modeAwareDeployer struct {
	fsys             fs.FS
	renderer         Renderer
	forceUpdate      bool
	installationMode string
	homeDir          string
}

// NewDeployerWithMode creates a Deployer that supports installation mode-aware deployment.
// In global mode, system files (agents/moai, skills/moai*, rules/moai) are deployed
// to ~/.claude/ instead of the project directory.
func NewDeployerWithMode(fsys fs.FS, mode, homeDir string) Deployer {
	return &modeAwareDeployer{
		fsys:             fsys,
		installationMode: mode,
		homeDir:          homeDir,
	}
}

// NewDeployerWithModeAndRenderer creates a mode-aware deployer with template rendering.
func NewDeployerWithModeAndRenderer(fsys fs.FS, renderer Renderer, mode, homeDir string) Deployer {
	return &modeAwareDeployer{
		fsys:             fsys,
		renderer:         renderer,
		installationMode: mode,
		homeDir:          homeDir,
	}
}

// NewDeployerWithModeAndRendererForceUpdate creates a mode-aware deployer with rendering
// and force update capability. Used for template updates where files should be overwritten.
func NewDeployerWithModeAndRendererForceUpdate(fsys fs.FS, renderer Renderer, mode, homeDir string, forceUpdate bool) Deployer {
	return &modeAwareDeployer{
		fsys:             fsys,
		renderer:         renderer,
		installationMode: mode,
		homeDir:          homeDir,
		forceUpdate:      forceUpdate,
	}
}

// @MX:NOTE: [AUTO] Deploy는 설치 모드에 따라 파일을 적절한 위치에 배치합니다. Global 모드에서는 시스템 파일을 ~/.claude/로, 프로젝트 파일은 로컬로 배포합니다.
// Deploy implements Deployer interface with mode-aware path resolution.
func (d *modeAwareDeployer) Deploy(ctx context.Context, projectRoot string, m manifest.Manager, tmplCtx *TemplateContext) error {
	projectRoot = filepath.Clean(projectRoot)

	// Use context installation mode if provided
	mode := d.installationMode
	homeDir := d.homeDir
	if tmplCtx != nil {
		if tmplCtx.InstallationMode != "" {
			mode = tmplCtx.InstallationMode
		}
		if tmplCtx.HomeDir != "" {
			homeDir = tmplCtx.HomeDir
		}
	}

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

		// Resolve destination path based on mode
		destPath := ResolveDeployPath(destRelPath, mode, projectRoot, homeDir)

		// Check if this is a global file (deployed to ~/.claude/)
		isGlobalDest := mode == "global" && IsGlobalFile(destRelPath)

		// Validate path security (for project files)
		if !isGlobalDest {
			if err := validateDeployPath(projectRoot, destRelPath); err != nil {
				deployErr = err
				return err
			}
		}

		// Skip existing files unless forceUpdate
		if !d.forceUpdate {
			if _, statErr := os.Stat(destPath); statErr == nil {
				// File exists - check manifest for provenance (project files only)
				if !isGlobalDest {
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
				} else {
					// Global file exists - skip to avoid overwriting user's global customizations
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

		// Track in manifest (only for project-local files, not global files)
		// Global files are not tracked per-project since they're shared
		if !isGlobalDest {
			templateHash := manifest.HashBytes(content)
			if err := m.Track(destRelPath, manifest.TemplateManaged, templateHash); err != nil {
				return fmt.Errorf("template deploy track %q: %w", destRelPath, err)
			}
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
