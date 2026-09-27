// Package homestate owns project-scoped state paths below ~/.moai.
package homestate

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gitcore "github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/gitenv"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// ProjectKey returns a stable, readable key for a project. Linked worktrees
// share one key through the repository common directory — the primary
// checkout's key in an ordinary repository; see CanonicalProjectRoot for the
// layouts where that shared root is a git directory instead.
func ProjectKey(projectRoot string) string {
	return projectKeyFromCanonicalRoot(CanonicalProjectRoot(projectRoot))
}

// ProjectKeyForCanonicalRoot returns the project key of a root that is
// already canonical (the root CanonicalProjectRoot would return). It exists
// for callers on a per-tool-call hook path that derive the canonical root
// from the repository's files without starting git, and must still land on
// exactly the key ProjectKey computes.
func ProjectKeyForCanonicalRoot(canonicalRoot string) string {
	return projectKeyFromCanonicalRoot(canonicalRoot)
}

func projectKeyFromCanonicalRoot(root string) string {
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

// CanonicalProjectRoot normalizes a worktree path to the root every worktree of
// the repository shares: the primary checkout in an ordinary repository. In a
// --separate-git-dir, bare, or submodule repository git records no checkout
// for the main worktree, so a linked worktree resolves to the git directory
// itself (the metadata dir, the bare repo, .git/modules/<name>). That root is a
// key, not a place to write: ProjectDir never lays state out inside it (t1221).
func CanonicalProjectRoot(projectRoot string) string {
	if projectRoot == "" {
		projectRoot = "."
	}
	abs, err := filepath.Abs(projectRoot)
	if err == nil {
		projectRoot = abs
	}
	if dirs, err := gitcore.ResolveGitDirs(projectRoot); err == nil && dirs.CommonDir != "" {
		if dirs.GitDir == dirs.CommonDir {
			// Metadata may live outside the checkout (--separate-git-dir).
			if out, err := scrubbedGit(projectRoot, "rev-parse", "--show-toplevel").Output(); err == nil {
				projectRoot = strings.TrimSpace(string(out))
			}
		} else if root, ok := primaryCheckoutRootFromCommonDir(dirs.CommonDir); ok {
			projectRoot = root
		} else if out, err := scrubbedGit(projectRoot, "worktree", "list", "--porcelain").Output(); err == nil {
			// Git lists the main worktree first. With --separate-git-dir, bare
			// or submodule layouts that entry is the git directory itself.
			first, _, _ := strings.Cut(string(out), "\n")
			if root, ok := strings.CutPrefix(first, "worktree "); ok {
				projectRoot = root
			}
		}
	}
	if resolved, err := filepath.EvalSymlinks(projectRoot); err == nil {
		projectRoot = resolved
	}
	return filepath.Clean(projectRoot)
}

// scrubbedGit builds `git -C dir args...` without the caller's repository-
// locating variables: an inherited GIT_DIR / GIT_WORK_TREE (git exports them
// into hooks) would otherwise answer about the caller's checkout (t1208).
func scrubbedGit(dir string, args ...string) *exec.Cmd {
	cmd := gitcore.ExecCommand("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = gitenv.Env()
	return cmd
}

// primaryCheckoutRootFromCommonDir handles Git's ordinary linked-worktree
// layout without enumerating every registered worktree. Repositories with a
// separate or bare common directory fall back to `git worktree list` above.
func primaryCheckoutRootFromCommonDir(commonDir string) (string, bool) {
	commonDir = filepath.Clean(commonDir)
	if filepath.Base(commonDir) != ".git" {
		return "", false
	}
	return filepath.Dir(commonDir), true
}

// tempRoots mirrors the production anchor set of kanban's TempOriginReason
// (internal/kanban/temp_origin.go defaultTempRoots, REQ-THG-002): os.TempDir()
// plus the /tmp and /var/folders spellings, whose containment os.TempDir()
// alone misses on machines where TMPDIR points at the per-user directory
// (/var/folders/... on macOS) while a project sits under /tmp. /var/tmp stays
// excluded for the same reboot-survival reason the kanban set records. The
// set is duplicated here — homestate cannot import kanban (kanban imports
// homestate) — and TestTempDiscriminantParity asserts the two discriminants
// agree so the sibling resolvers cannot split again.
func tempRoots() []string {
	return []string{os.TempDir(), "/tmp", "/var/folders"}
}

// insideTempRoots reports whether path lies within any temp anchor.
// pathInside resolves the anchor side through EvalSymlinks, so the
// macOS /tmp -> /private/tmp spelling is covered without a second anchor
// form; path is the canonical root, already resolved by
// CanonicalProjectRoot.
func insideTempRoots(path string) bool {
	for _, root := range tempRoots() {
		rootAbs, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		if pathInside(path, rootAbs) {
			return true
		}
	}
	return false
}

// ProjectDir returns ~/.moai/db/<project-key>.
//
// The two branches spell the key argument differently — the temp branch passes
// the already-canonical root, the home branch the caller's raw projectRoot —
// and the spellings are equivalent because ProjectKey canonicalizes its own
// argument before hashing it. Reading the asymmetry as a defect that splits one
// project across two keys has already been filed once and measured false
// (TestProjectKeyArgumentEquivalence pins the equivalence across the six root
// shapes callers actually pass, with a positive control). The guard is what
// would break if CanonicalProjectRoot ever stopped being idempotent; until it
// does, neither spelling is wrong.
func ProjectDir(projectRoot string) (string, error) {
	canonical := CanonicalProjectRoot(projectRoot)
	key := projectKeyFromCanonicalRoot(canonical)
	if rootLayout(canonical) {
		return filepath.Join(canonical, ".moai", "db", key), nil
	}
	home, err := paths.MoaiHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "db", key), nil
}

// rootLayout reports whether project state lives under <canonical>/.moai
// rather than the home layout: a temp-rooted project with no explicit
// MOAI_HOME — unless the canonical root is a git directory, which is a key and
// never a place to write (t1221). The three callers share this one predicate so
// the directories they create cannot disagree about where state lives.
func rootLayout(canonical string) bool {
	return !explicitMoaiHome() && insideTempRoots(canonical) && !isGitDir(canonical)
}

// isGitDir reports whether dir is a git directory (a --separate-git-dir
// metadata dir, a bare repository, .git/modules/<name>) or lies inside one,
// rather than in a work tree. Two answers are combined because each misses a
// case the other catches: `--absolute-git-dir` equal to dir recognises a git
// directory that carries core.worktree, where `--is-inside-git-dir` says false
// (a submodule's always does; t1221 re-audit F1), while `--is-inside-git-dir`
// recognises a path below a git directory, such as .git/refs (delta audit N1).
func isGitDir(dir string) bool {
	out, err := scrubbedGit(dir, "rev-parse", "--absolute-git-dir", "--is-inside-git-dir").Output()
	if err != nil {
		return false
	}
	gitDir, inside, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	return sameDir(gitDir, dir) || strings.TrimSpace(inside) == "true"
}

// sameDir reports whether a and b name the same directory once cleaned and
// symlink-resolved (git may spell a path through /private on macOS).
func sameDir(a, b string) bool {
	resolve := func(p string) string {
		if r, err := filepath.EvalSymlinks(p); err == nil {
			p = r
		}
		return filepath.Clean(p)
	}
	return resolve(a) == resolve(b)
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
	if !rootLayout(canonical) {
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
	if !rootLayout(canonical) {
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
