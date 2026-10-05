package constitution

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// RegistryPathEnv names the environment variable whose non-empty value
// overrides the registry path.
const RegistryPathEnv = "MOAI_CONSTITUTION_REGISTRY"

// RegistryRelPath is the project-relative path of the default zone registry.
const RegistryRelPath = ".claude/rules/moai/core/zone-registry.md"

// ResolveRegistryPath returns the zone-registry path by precedence: a
// non-empty MOAI_CONSTITUTION_REGISTRY, then
// <CLAUDE_PROJECT_DIR>/.claude/rules/moai/core/zone-registry.md when
// CLAUDE_PROJECT_DIR is non-empty, then the same path under projectDir.
// SPEC-CON-AMEND-APPLY-001 REQ-CAA-019.
//
// @MX:ANCHOR: [AUTO] the one registry path resolver shared by the CLI and Pipeline.Execute
// @MX:REASON: REQ-CAA-019 — the file the CLI validates must be the file Execute writes; a second join reintroduces the split
func ResolveRegistryPath(projectDir string) string {
	if v := os.Getenv(RegistryPathEnv); v != "" {
		return v
	}
	if d := os.Getenv(config.EnvClaudeProjectDir); d != "" {
		return filepath.Join(d, filepath.FromSlash(RegistryRelPath))
	}
	return filepath.Join(projectDir, filepath.FromSlash(RegistryRelPath))
}

// LoadAmendRegistry is the amend path's registry load (REQ-CAA-020,
// REQ-CAA-021): the containment check on the registry path runs before the
// file is read, then the unchanged LoadRegistry, then the containment check
// on every entry's file:. Commands that only read the registry keep calling
// LoadRegistry, which performs no containment check (operator decision D4).
//
// @MX:ANCHOR: [AUTO] amend-path registry admission — used by Pipeline.Execute and runConstitutionAmend
// @MX:REASON: REQ-CAA-021 one check for both amend callers; moving it into LoadRegistry reaches the five read-only callers D4 excludes
func LoadAmendRegistry(registryPath, projectDir string) (*Registry, error) {
	if err := checkContained(projectDir, registryPath); err != nil {
		return nil, fmt.Errorf("registry path: %w", err)
	}
	reg, err := LoadRegistry(registryPath, projectDir)
	if err != nil {
		return nil, err
	}
	for _, e := range reg.Entries {
		if err := checkContained(projectDir, ruleFilePath(projectDir, e.File)); err != nil {
			return nil, fmt.Errorf("rule %s file: %w", e.ID, err)
		}
	}
	return reg, nil
}

// ruleFilePath returns an entry's file: joined with projectDir when relative,
// and as given when absolute.
//
// @MX:ANCHOR: [AUTO] the one join of an entry's file: with projectDir, shared by the containment check, the rule-file alias check, and the apply read
// @MX:REASON: fan_in 3 (LoadAmendRegistry, Pipeline.Execute, prepareApply); a second join could make the path that is checked differ from the path that is written
func ruleFilePath(projectDir, file string) string {
	if filepath.IsAbs(file) {
		return file
	}
	return filepath.Join(projectDir, file)
}

// checkContained is the one containment check of REQ-CAA-021: the path is
// cleaned, made absolute against the working directory, and has its symbolic
// links resolved; it is inside only when the result equals the likewise
// resolved projectDir or continues from it past a path separator.
func checkContained(projectDir, path string) error {
	root, err := resolveExisting(projectDir)
	if err != nil {
		return fmt.Errorf("resolve project dir %s: %w", projectDir, err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve %s: %w", path, err)
	}
	resolved, err := resolveExisting(abs)
	if err != nil {
		return fmt.Errorf("resolve %s: %w", abs, err)
	}
	if !isWithin(root, resolved) {
		return fmt.Errorf("path %s (resolves to %s) is outside project dir %s", abs, resolved, projectDir)
	}
	return nil
}

// isWithin reports whether p equals root or continues from it past a path
// separator, so a sibling such as <root>-evil is outside <root>.
func isWithin(root, p string) bool {
	rel, err := filepath.Rel(root, p)
	if err != nil || filepath.IsAbs(rel) {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// resolveExisting makes p absolute and resolves its symbolic links. A path
// that does not exist yet is judged by resolving its nearest existing
// ancestor and appending the remaining components.
func resolveExisting(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	var tail []string
	for cur := abs; ; {
		r, err := filepath.EvalSymlinks(cur)
		if err == nil {
			for i := len(tail) - 1; i >= 0; i-- {
				r = filepath.Join(r, tail[i])
			}
			return r, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return abs, nil
		}
		tail = append(tail, filepath.Base(cur))
		cur = parent
	}
}
