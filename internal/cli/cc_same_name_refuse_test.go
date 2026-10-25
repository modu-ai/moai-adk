package cli

// cc_same_name_refuse_test.go — card t1791 (GH #1813) RED. A `-w <name>`
// short name whose MoAI tree (.moai/worktrees/<name>) exists while the
// native .claude/worktrees/<name> tree does not must be REFUSED before the
// backend runs: claude would silently create a duplicate native branch tree
// beside the existing MoAI tree (the 2026-10-08 lane-12 incident). The
// refusal changes no argv — the existing tree is entered by its absolute
// path, `moai cc -w <absolute path>`. When the native tree exists too, the
// existing behavior stands: the backend receives the selection untouched.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/paths"
)

// refuseFixture builds a scratch git repository with an existing MoAI tree
// named name under .moai/worktrees, an isolated MOAI_HOME, and no native
// tree. It returns the MoAI tree's absolute path.
func refuseFixture(t *testing.T, name string) string {
	t.Helper()
	root := m2Repo(t)
	t.Setenv(paths.EnvHome, t.TempDir())
	netScrubLaneEnv(t)
	moai := filepath.Join(root, ".moai", "worktrees", name)
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", moai, err)
	}
	return moai
}

func TestCCRefusesShortNameShadowedByMoaiTree(t *testing.T) {
	moai := refuseFixture(t, "feat")
	launch := netDriveLaunch(t, filepath.Dir(filepath.Dir(filepath.Dir(moai))), netCC, []string{"-w", "feat"})
	if launch.launched {
		t.Errorf("a refused entry launched the backend; argv = %v", launch.args)
	}
	if launch.err == nil {
		t.Fatal("the shadowed short name was accepted, want the refusal")
	}
	if !strings.Contains(launch.err.Error(), moai) {
		t.Errorf("the refusal names no tree path: %q, want %q", launch.err.Error(), moai)
	}
	// Audit r1 P2-1 (rcpt-90420fa3, measured on claude 2.1.309): an absolute
	// path as a --worktree value exits 1 with "Invalid worktree name", so the
	// guidance must be the directory-entry form, never `-w <path>`.
	if !strings.Contains(launch.err.Error(), "cd "+shellSingleQuoted(moai)+" && moai cc") {
		t.Errorf("the refusal carries no directory-entry guidance: %q", launch.err.Error())
	}
	if strings.Contains(launch.err.Error(), "moai cc -w") {
		t.Errorf("the guidance names the measured-broken -w <path> form: %q", launch.err.Error())
	}
}

// TestCCGuidanceQuotesAHostileTreeName (audit r2 P1, rcpt-c6749fd0): the
// guided `cd` command single-quotes the path — a tree name carrying a
// single quote and command substitution must survive the copy-paste as a
// literal, never execute.
func TestCCGuidanceQuotesAHostileTreeName(t *testing.T) {
	root := m2Repo(t)
	t.Setenv(paths.EnvHome, t.TempDir())
	netScrubLaneEnv(t)
	hostile := "fe'at$(touch MOAI_REVIEW_PROBE)"
	moai := filepath.Join(root, ".moai", "worktrees", hostile)
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", moai, err)
	}
	launch := netDriveLaunch(t, root, netCC, []string{"-w", hostile})
	if launch.err == nil || launch.launched {
		t.Fatalf("want the refusal, got err=%v launched=%v", launch.err, launch.launched)
	}
	if !strings.Contains(launch.err.Error(), "cd "+shellSingleQuoted(moai)) {
		t.Errorf("the guidance does not single-quote the hostile path: %q", launch.err.Error())
	}
}

// TestCCRefusalSurvivesASeparateGitDir (audit r2 P2, rcpt-c6749fd0): the
// primary is the main worktree record, not a git-common-dir parent — a
// --separate-git-dir layout (metadata outside the checkout) must not bypass
// the same-name check.
func TestCCRefusalSurvivesASeparateGitDir(t *testing.T) {
	base := m2Real(t, t.TempDir())
	primary := filepath.Join(base, "repo")
	metadata := filepath.Join(base, "metadata")
	m2Git(t, base, "init", "-q", "--template=", "--separate-git-dir", metadata, "-b", "main", primary)
	m2Git(t, primary, "config", "user.email", "m2-test@example.com")
	m2Git(t, primary, "config", "user.name", "M2 Test")
	if err := os.WriteFile(filepath.Join(primary, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	m2Git(t, primary, "add", "README.md")
	m2Git(t, primary, "commit", "-q", "-m", "seed")
	t.Setenv(paths.EnvHome, t.TempDir())
	netScrubLaneEnv(t)
	moai := filepath.Join(primary, ".moai", "worktrees", "feat")
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", moai, err)
	}
	launch := netDriveLaunch(t, primary, netCC, []string{"-w", "feat"})
	if launch.err == nil || launch.launched {
		t.Fatalf("a separate-git-dir layout must not bypass the refusal: err=%v launched=%v", launch.err, launch.launched)
	}
	if !strings.Contains(launch.err.Error(), moai) {
		t.Errorf("the refusal names no tree path: %q, want %q", launch.err.Error(), moai)
	}
}

// TestCCRefusesWhenThePrimaryCannotBeResolved (audit r2 P2, fail-closed
// arm): a linked worktree whose repository carries an external git dir has
// no git-native address for its main checkout — the launch is refused, not
// guessed through.
func TestCCRefusesWhenThePrimaryCannotBeResolved(t *testing.T) {
	base := m2Real(t, t.TempDir())
	primary := filepath.Join(base, "repo")
	metadata := filepath.Join(base, "metadata")
	m2Git(t, base, "init", "-q", "--template=", "--separate-git-dir", metadata, "-b", "main", primary)
	m2Git(t, primary, "config", "user.email", "m2-test@example.com")
	m2Git(t, primary, "config", "user.name", "M2 Test")
	if err := os.WriteFile(filepath.Join(primary, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	m2Git(t, primary, "add", "README.md")
	m2Git(t, primary, "commit", "-q", "-m", "seed")
	moai := filepath.Join(primary, ".moai", "worktrees", "feat")
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", moai, err)
	}
	linked := filepath.Join(base, "wt-linked")
	m2Git(t, primary, "worktree", "add", "-q", "-b", "wt-linked", linked)
	t.Setenv(paths.EnvHome, t.TempDir())
	netScrubLaneEnv(t)
	launch := netDriveLaunch(t, linked, netCC, []string{"-w", "feat"})
	if launch.err == nil || launch.launched {
		t.Fatalf("an unresolvable layout must refuse, not launch: err=%v launched=%v", launch.err, launch.launched)
	}
	if !strings.Contains(launch.err.Error(), "cannot be resolved") {
		t.Errorf("the refusal does not name the unresolvable layout: %q", launch.err.Error())
	}
}

func TestCCPassesThroughWhenTheNativeTreeExists(t *testing.T) {
	root := m2Repo(t)
	t.Setenv(paths.EnvHome, t.TempDir())
	netScrubLaneEnv(t)
	for _, dir := range []string{
		filepath.Join(root, ".moai", "worktrees", "feat"),
		filepath.Join(root, ".claude", "worktrees", "feat"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	launch := netDriveLaunch(t, root, netCC, []string{"-w", "feat"})
	if launch.err != nil {
		t.Fatalf("the native tree exists; the launch must proceed: %v", launch.err)
	}
	if !launch.launched {
		t.Fatal("the launch never ran")
	}
	if v, ok := worktreeFlagValue(launch.args); !ok || v != "feat" {
		t.Errorf("the backend selection was altered: ok=%v value=%q; the argv passes through untouched", ok, v)
	}
}

func TestGLMRefusesShortNameShadowedByMoaiTree(t *testing.T) {
	moai := refuseFixture(t, "feat")
	launch := netDriveLaunch(t, filepath.Dir(filepath.Dir(filepath.Dir(moai))), netGLM, []string{"-w", "feat"})
	if launch.launched {
		t.Errorf("a refused entry launched the backend; argv = %v", launch.args)
	}
	if launch.err == nil {
		t.Fatal("the shadowed short name was accepted, want the refusal")
	}
	if !strings.Contains(launch.err.Error(), moai) {
		t.Errorf("the refusal names no tree path: %q, want %q", launch.err.Error(), moai)
	}
	if !strings.Contains(launch.err.Error(), "cd "+shellSingleQuoted(moai)+" && moai glm") {
		t.Errorf("the refusal carries no directory-entry guidance: %q", launch.err.Error())
	}
	if strings.Contains(launch.err.Error(), "moai glm -w") {
		t.Errorf("the guidance names the measured-broken -w <path> form: %q", launch.err.Error())
	}
}

// TestCCRefusalSurvivesABrokenMoaiHome (audit r1 P2-2, rcpt-90420fa3): the
// refusal's two-path check resolves the primary through the git common
// directory, independent of the MoAI home metadata — a broken MOAI_HOME
// must not bypass it into the duplicate-creation path.
func TestCCRefusalSurvivesABrokenMoaiHome(t *testing.T) {
	root := m2Repo(t)
	home := filepath.Join(t.TempDir(), "moai-home")
	if err := os.WriteFile(home, []byte("not a directory\n"), 0o644); err != nil {
		t.Fatalf("write home file: %v", err)
	}
	t.Setenv(paths.EnvHome, home)
	netScrubLaneEnv(t)
	moai := filepath.Join(root, ".moai", "worktrees", "feat")
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", moai, err)
	}
	launch := netDriveLaunch(t, root, netCC, []string{"-w", "feat"})
	if launch.err == nil || launch.launched {
		t.Fatalf("a broken MoAI home must not bypass the refusal: err=%v launched=%v", launch.err, launch.launched)
	}
	if !strings.Contains(launch.err.Error(), moai) {
		t.Errorf("the refusal names no tree path: %q, want %q", launch.err.Error(), moai)
	}
}

// TestCCRefusalJudgesTheBackendSelectionNotTheFirstFlag (card t1791 review
// P2-1): the backend resolves the LAST -w selection, so the refusal judges
// the backend's selection — a shadowed name behind a later selection is dead
// text and must not refuse; the shadowed name that IS the selection must.
func TestCCRefusalJudgesTheBackendSelectionNotTheFirstFlag(t *testing.T) {
	root := m2Repo(t)
	t.Setenv(paths.EnvHome, t.TempDir())
	netScrubLaneEnv(t)
	moai := filepath.Join(root, ".moai", "worktrees", "feat")
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", moai, err)
	}
	launch := netDriveLaunch(t, root, netCC, []string{"-w", "feat", "-w", "other"})
	if launch.err != nil {
		t.Fatalf("the backend selects other; the feat shadow is dead text, no refusal: %v", launch.err)
	}
	launch = netDriveLaunch(t, root, netCC, []string{"-w", "other", "-w", "feat"})
	if launch.err == nil || launch.launched {
		t.Fatalf("the backend selection feat is shadowed; want the refusal, got err=%v launched=%v", launch.err, launch.launched)
	}
	if !strings.Contains(launch.err.Error(), moai) {
		t.Errorf("the refusal names no tree path: %q, want %q", launch.err.Error(), moai)
	}
}

// TestCCRefusalNamesTheResolvedTreeNotTheCallingRoot (card t1791 review
// P2-2): the launcher can run with its project root naming a linked tree
// while the candidate search resolves the primary checkout — the guidance
// must name the tree the search actually found, never a path spelled from
// the calling root.
func TestCCRefusalNamesTheResolvedTreeNotTheCallingRoot(t *testing.T) {
	primary := m2Repo(t)
	t.Setenv(paths.EnvHome, t.TempDir())
	netScrubLaneEnv(t)
	moai := filepath.Join(primary, ".moai", "worktrees", "feat")
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", moai, err)
	}
	linked := filepath.Join(filepath.Dir(primary), "wt-linked")
	m2Git(t, primary, "worktree", "add", "-q", "-b", "wt-linked", linked)
	launch := netDriveLaunch(t, linked, netCC, []string{"-w", "feat"})
	if launch.err == nil || launch.launched {
		t.Fatalf("want the refusal from the linked tree, got err=%v launched=%v", launch.err, launch.launched)
	}
	if !strings.Contains(launch.err.Error(), moai) {
		t.Errorf("the guidance names a path other than the resolved tree: %q, want %q", launch.err.Error(), moai)
	}
}

// TestCCNativeAliasOfTheMoaiTreeIsNotRefused (card t1791 review round 2
// P2): a native tree that aliases the MoAI tree (a symlink) is an existing
// native path — claude re-enters it natively and the launch is not refused.
// The native side is judged by its own existence, never through a candidate
// list that dedups same-directory spellings.
func TestCCNativeAliasOfTheMoaiTreeIsNotRefused(t *testing.T) {
	root := m2Repo(t)
	t.Setenv(paths.EnvHome, t.TempDir())
	netScrubLaneEnv(t)
	moai := filepath.Join(root, ".moai", "worktrees", "feat")
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", moai, err)
	}
	nativeBase := filepath.Join(root, ".claude", "worktrees")
	if err := os.MkdirAll(nativeBase, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", nativeBase, err)
	}
	if err := os.Symlink(moai, filepath.Join(nativeBase, "feat")); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	launch := netDriveLaunch(t, root, netCC, []string{"-w", "feat"})
	if launch.err != nil {
		t.Fatalf("the native path exists (aliased); the launch must proceed: %v", launch.err)
	}
	if !launch.launched {
		t.Fatal("the launch never ran")
	}
}

func TestGLMPassesThroughWhenTheNativeTreeExists(t *testing.T) {
	root := m2Repo(t)
	t.Setenv(paths.EnvHome, t.TempDir())
	netScrubLaneEnv(t)
	for _, dir := range []string{
		filepath.Join(root, ".moai", "worktrees", "feat"),
		filepath.Join(root, ".claude", "worktrees", "feat"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	launch := netDriveLaunch(t, root, netGLM, []string{"-w", "feat"})
	if launch.err != nil {
		t.Fatalf("the native tree exists; the launch must proceed: %v", launch.err)
	}
	if !launch.launched {
		t.Fatal("the launch never ran")
	}
	if v, ok := worktreeFlagValue(launch.args); !ok || v != "feat" {
		t.Errorf("the backend selection was altered: ok=%v value=%q; the argv passes through untouched", ok, v)
	}
}
