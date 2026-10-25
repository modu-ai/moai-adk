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
	if !strings.Contains(launch.err.Error(), "moai cc -w "+moai) {
		t.Errorf("the refusal carries no absolute-path entry guidance: %q", launch.err.Error())
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
	if !strings.Contains(launch.err.Error(), "moai cc -w "+moai) {
		t.Errorf("the refusal carries no absolute-path entry guidance: %q", launch.err.Error())
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
