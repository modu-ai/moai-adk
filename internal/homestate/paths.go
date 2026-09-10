// Package homestate owns project-scoped state paths below ~/.moai.
package homestate

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gitcore "github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// ProjectKey returns a stable, readable key for a project. Linked worktrees
// converge on the primary checkout through the repository common directory.
func ProjectKey(projectRoot string) string {
	root := CanonicalProjectRoot(projectRoot)
	sum := sha256.Sum256([]byte(root))
	base := filepath.Base(root)
	if base == "." || base == string(filepath.Separator) || base == "" {
		base = "project"
	}
	base = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			return r
		default:
			return '-'
		}
	}, base)
	return fmt.Sprintf("%s-%x", base, sum[:4])
}

// CanonicalProjectRoot normalizes a worktree path to the primary checkout.
func CanonicalProjectRoot(projectRoot string) string {
	if projectRoot == "" {
		projectRoot = "."
	}
	abs, err := filepath.Abs(projectRoot)
	if err == nil {
		projectRoot = abs
	}
	if dirs, err := gitcore.ResolveGitDirs(projectRoot); err == nil && dirs.CommonDir != "" {
		projectRoot = filepath.Dir(dirs.CommonDir)
	}
	if resolved, err := filepath.EvalSymlinks(projectRoot); err == nil {
		projectRoot = resolved
	}
	return filepath.Clean(projectRoot)
}

// ProjectDir returns ~/.moai/db/<project-key>.
func ProjectDir(projectRoot string) (string, error) {
	canonical := CanonicalProjectRoot(projectRoot)
	if !explicitMoaiHome() && pathInside(canonical, os.TempDir()) {
		return filepath.Join(canonical, ".moai", "db", ProjectKey(canonical)), nil
	}
	home, err := paths.MoaiHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "db", ProjectKey(projectRoot)), nil
}

func explicitMoaiHome() bool {
	v := os.Getenv(paths.EnvHome)
	return v != "" && filepath.IsAbs(v)
}

func pathInside(path, parent string) bool {
	if resolved, err := filepath.EvalSymlinks(parent); err == nil {
		parent = resolved
	}
	rel, err := filepath.Rel(parent, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func TodoDir(projectRoot string) (string, error) {
	dir, err := ProjectDir(projectRoot)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "todo"), nil
}

func BacklogDBPath(projectRoot string) (string, error) {
	dir, err := TodoDir(projectRoot)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "backlog.db"), nil
}

func FactoryDir(projectRoot string) (string, error) {
	dir, err := ProjectDir(projectRoot)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "factory"), nil
}

func FactoryDBPath(projectRoot string) (string, error) {
	dir, err := FactoryDir(projectRoot)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "factory.db"), nil
}

func SearchDBPath(projectRoot string) (string, error) {
	home, err := paths.MoaiHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "cache", "search", ProjectKey(projectRoot), "sessions.db"), nil
}

func RunProjectDir(projectRoot string) (string, error) {
	home, err := paths.MoaiHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "run", ProjectKey(projectRoot)), nil
}

// EnsureHomeLayout materializes the supported ~/.moai top-level contract.
// Every directory is private because profiles, credentials, prompts, logs, and
// SQLite rows can all contain session or account context.
func EnsureHomeLayout() error {
	home, err := paths.MoaiHome()
	if err != nil {
		return err
	}
	dirs := []string{
		home,
		filepath.Join(home, "claude-profiles"),
		filepath.Join(home, "config"),
		filepath.Join(home, "credentials"),
		filepath.Join(home, "db"),
		filepath.Join(home, "cache", "search"),
		filepath.Join(home, "run"),
		filepath.Join(home, "logs"),
		filepath.Join(home, "integrations"),
		filepath.Join(home, "bin"),
		filepath.Join(home, "releases"),
		filepath.Join(home, "backups"),
		filepath.Join(home, "reports"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
		if err := os.Chmod(dir, 0o700); err != nil {
			return err
		}
	}
	return nil
}

// EnsureProjectLayout creates the durable project directories with private
// permissions and records the canonical root used to derive the key.
func EnsureProjectLayout(projectRoot string) error {
	canonical := CanonicalProjectRoot(projectRoot)
	if explicitMoaiHome() || !pathInside(canonical, os.TempDir()) {
		if err := EnsureHomeLayout(); err != nil {
			return err
		}
	}
	dir, err := ProjectDir(projectRoot)
	if err != nil {
		return err
	}
	for _, p := range []string{dir, filepath.Join(dir, "todo"), filepath.Join(dir, "factory")} {
		if err := os.MkdirAll(p, 0o700); err != nil {
			return err
		}
		if err := os.Chmod(p, 0o700); err != nil {
			return err
		}
	}
	if explicitMoaiHome() || !pathInside(canonical, os.TempDir()) {
		searchPath, err := SearchDBPath(projectRoot)
		if err != nil {
			return err
		}
		runDir, err := RunProjectDir(projectRoot)
		if err != nil {
			return err
		}
		for _, p := range []string{filepath.Dir(searchPath), runDir, filepath.Join(runDir, "locks"), filepath.Join(runDir, "sockets"), filepath.Join(runDir, "guard-liveness")} {
			if err := os.MkdirAll(p, 0o700); err != nil {
				return err
			}
			if err := os.Chmod(p, 0o700); err != nil {
				return err
			}
		}
	}
	manifest := struct {
		SchemaVersion int    `json:"schema_version"`
		ProjectKey    string `json:"project_key"`
		ProjectRoot   string `json:"project_root"`
	}{1, ProjectKey(projectRoot), canonical}
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	path := filepath.Join(dir, "project.json")
	if existing, readErr := os.ReadFile(path); readErr == nil && string(existing) == string(raw) {
		return os.Chmod(path, 0o600)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

// ProjectRootFromDBPath reads the project manifest adjacent to todo/ and
// factory/. It is used only for one-time legacy imports.
func ProjectRootFromDBPath(dbPath string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(filepath.Dir(dbPath)), "project.json"))
	if err != nil {
		return "", err
	}
	var manifest struct {
		ProjectRoot string `json:"project_root"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return "", err
	}
	if manifest.ProjectRoot == "" {
		return "", fmt.Errorf("project manifest has empty project_root")
	}
	return manifest.ProjectRoot, nil
}
