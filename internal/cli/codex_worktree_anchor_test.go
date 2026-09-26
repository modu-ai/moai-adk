package cli

// codex_worktree_anchor_test.go — `moai codex -w` owns its worktree the way a
// Claude session does: a git worktree lock carrying the pid of the process that
// becomes Codex, a creation base that is verified rather than assumed, and a
// refusal to become a second writer in a tree someone else is anchored in.
// `moai cc -w` gets the same refusal through a pre-check that only reads.
//
// Every fixture is a REAL repository under t.TempDir(). The capture harness
// replaces the process launch; the anchor seams run their real bodies
// (withRealCodexWorktreeAnchor), because those bodies are what is measured.

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/session"
)

// anchorGit runs git with -C dir and fails the test on error.
func anchorGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git -C %s %v: %v\n%s", dir, args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// anchorRepo is a real repository with a project marker, its base branch, and
// the helpers that build trees under <root>/.claude/worktrees/.
type anchorRepo struct {
	root string
	base string
}

func newAnchorRepo(t *testing.T) *anchorRepo {
	t.Helper()
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve temp dir: %v", err)
	}
	root := filepath.Join(tmp, "proj")
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	anchorGit(t, root, "init", "-q")
	anchorGit(t, root, "config", "user.email", "anchor-test@example.com")
	anchorGit(t, root, "config", "user.name", "Anchor Test")
	anchorGit(t, root, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	anchorGit(t, root, "add", "README.md")
	anchorGit(t, root, "commit", "-q", "-m", "seed")
	r := &anchorRepo{root: root, base: anchorGit(t, root, "branch", "--show-current")}

	// Keep every registry the anchor decision consults inside the fixture.
	t.Setenv(config.EnvClaudeProjectDir, root)
	withCodexProjectRoot(t, root)

	// The real materializer runs `git worktree add` in the process cwd, which
	// is this package's own checkout; point it at the fixture instead.
	oldBase, oldAdd := codexWorktreeBase, codexWorktreeAdd
	codexWorktreeBase = func(string) (string, error) { return r.base, nil }
	codexWorktreeAdd = func(path, branch, base string) (string, error) {
		out, err := exec.Command("git", "-C", root, "worktree", "add", "-q", "-b", branch, path, base).CombinedOutput()
		if err != nil {
			return "", errors.New(strings.TrimSpace(string(out)))
		}
		return path, nil
	}
	t.Cleanup(func() {
		codexWorktreeBase, codexWorktreeAdd = oldBase, oldAdd
		out, _ := exec.Command("git", "-C", root, "worktree", "list", "--porcelain").Output()
		for path := range session.ParseWorktreeLocks(string(out)) {
			_ = exec.Command("git", "-C", root, "worktree", "unlock", path).Run()
		}
	})
	return r
}

// tree creates an unlocked linked worktree <root>/.claude/worktrees/<name>.
func (r *anchorRepo) tree(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(r.root, ".claude", "worktrees", name)
	anchorGit(t, r.root, "worktree", "add", "-q", "-b", "WT-"+name, path, r.base)
	return path
}

func lockAnchorTree(t *testing.T, tree, reason string) {
	t.Helper()
	anchorGit(t, tree, "worktree", "lock", "--reason", reason, tree)
}

// observedLock reads the tree's lock straight from git's porcelain, with no
// production helper in between.
func observedLock(t *testing.T, tree string) session.LockInfo {
	t.Helper()
	locks := session.ParseWorktreeLocks(anchorGit(t, tree, "worktree", "list", "--porcelain"))
	return locks[tree]
}

// exitedPID returns the pid of a process that has already exited.
func exitedPID(t *testing.T) int {
	t.Helper()
	cmd := exec.Command("git", "--version")
	if err := cmd.Run(); err != nil {
		t.Fatalf("run short-lived process: %v", err)
	}
	return cmd.ProcessState.Pid()
}

// anchorTreeState is what a refused launch must leave untouched.
type anchorTreeState struct {
	lock   session.LockInfo
	branch string
	head   string
	status string
}

func snapshotAnchorTree(t *testing.T, tree string) anchorTreeState {
	t.Helper()
	return anchorTreeState{
		lock:   observedLock(t, tree),
		branch: anchorGit(t, tree, "branch", "--show-current"),
		head:   anchorGit(t, tree, "rev-parse", "HEAD"),
		status: anchorGit(t, tree, "status", "--porcelain", "--untracked-files=all"),
	}
}

// withRealCodexWorktreeAnchor puts the anchor seams back on their real
// bodies after withCodexLaunchCapture replaced them.
func withRealCodexWorktreeAnchor(t *testing.T) {
	t.Helper()
	prevCheck, prevLock, prevBase, prevResolve := codexWorktreeWriterCheck, codexWorktreeAnchorLock, codexWorktreeBaseCheck, codexResolveBaseCommit
	codexWorktreeWriterCheck = worktreeWriterRefusal
	codexWorktreeAnchorLock = placeCodexAnchorLock
	codexWorktreeBaseCheck = verifyCodexWorktreeBase
	codexResolveBaseCommit = resolveBaseCommitReal
	t.Cleanup(func() {
		codexWorktreeWriterCheck, codexWorktreeAnchorLock, codexWorktreeBaseCheck, codexResolveBaseCommit = prevCheck, prevLock, prevBase, prevResolve
	})
}

// assertLockedByLauncher checks the headline REQ-DHR-008 property: the tree
// carries a lock whose pid is the process that becomes Codex, and the shared
// anchor decision reads it as anchored.
func assertLockedByLauncher(t *testing.T, tree string) {
	t.Helper()
	lock := observedLock(t, tree)
	if !lock.Locked {
		t.Fatalf("%s: no git worktree lock after launch", tree)
	}
	pid, ok := session.LockReasonPID(lock.Reason)
	if !ok || pid != codexDirectAnchorPID() {
		t.Fatalf("%s: lock reason %q carries pid %d (ok=%v), want the launcher pid %d", tree, lock.Reason, pid, ok, codexDirectAnchorPID())
	}
	if !strings.HasPrefix(lock.Reason, "moai codex session ") {
		t.Errorf("%s: lock reason %q does not name the Codex launcher", tree, lock.Reason)
	}
	if v := session.AnchorDecision(tree, lock, time.Now()); !v.Anchored || v.Source != session.AnchorSourceLock {
		t.Errorf("%s: anchor decision = %+v, want anchored by lock", tree, v)
	}
}

// TestCodexWorktreeAnchorLockAndBase is AC-DHR-006's launch half: new and
// existing trees are locked with the launcher's pid, a new tree's HEAD is the
// resolved base, a base mismatch refuses without deleting the tree, a dead
// lock is replaced, and a live lock is not.
func TestCodexWorktreeAnchorLockAndBase(t *testing.T) {
	t.Run("new_tree", func(t *testing.T) {
		cap := withCodexLaunchCapture(t)
		r := newAnchorRepo(t)
		withRealCodexWorktreeAnchor(t)
		if _, stderr, err := runCodexCmd(t, "-w", "new-tree"); err != nil {
			t.Fatalf("launch: %v (stderr %q)", err, stderr)
		}
		codexWantLaunches(t, cap, 1, 1, 0)
		tree := filepath.Join(r.root, ".claude", "worktrees", "new-tree")
		assertLockedByLauncher(t, tree)
		if head, base := anchorGit(t, tree, "rev-parse", "HEAD"), anchorGit(t, r.root, "rev-parse", r.base); head != base {
			t.Errorf("new tree HEAD %s != resolved base %s (%s)", head, base, r.base)
		}
	})

	t.Run("existing_tree", func(t *testing.T) {
		cap := withCodexLaunchCapture(t)
		r := newAnchorRepo(t)
		withRealCodexWorktreeAnchor(t)
		tree := r.tree(t, "existing-tree")
		if _, stderr, err := runCodexCmd(t, "-w", "existing-tree"); err != nil {
			t.Fatalf("launch: %v (stderr %q)", err, stderr)
		}
		codexWantLaunches(t, cap, 1, 1, 0)
		assertLockedByLauncher(t, tree)
	})

	t.Run("base_mismatch_refused_tree_kept", func(t *testing.T) {
		cap := withCodexLaunchCapture(t)
		r := newAnchorRepo(t)
		withRealCodexWorktreeAnchor(t)
		codexResolveBaseCommit = func(string, string) (string, error) {
			return "0000000000000000000000000000000000000001", nil
		}
		_, stderr, err := runCodexCmd(t, "-w", "drifted")
		if err == nil {
			t.Fatal("launch accepted a tree whose HEAD is not the resolved base")
		}
		if code, ok := ResolveExitCode(err); !ok || code == 0 {
			t.Errorf("exit = (%d, %v), want non-zero", code, ok)
		}
		codexWantLaunches(t, cap, 0, 0, 0)
		tree := filepath.Join(r.root, ".claude", "worktrees", "drifted")
		if _, statErr := os.Stat(tree); statErr != nil {
			t.Errorf("mismatched tree was deleted: %v", statErr)
		}
		if !strings.Contains(stderr, "does not match the resolved base") {
			t.Errorf("diagnostic %q does not name the base mismatch", stderr)
		}
	})

	t.Run("dead_lock_replaced", func(t *testing.T) {
		cap := withCodexLaunchCapture(t)
		r := newAnchorRepo(t)
		withRealCodexWorktreeAnchor(t)
		tree := r.tree(t, "stale")
		lockAnchorTree(t, tree, session.CodexAnchorLockReason("stale", exitedPID(t), ""))
		if _, stderr, err := runCodexCmd(t, "-w", "stale"); err != nil {
			t.Fatalf("launch over a dead lock: %v (stderr %q)", err, stderr)
		}
		codexWantLaunches(t, cap, 1, 1, 0)
		assertLockedByLauncher(t, tree)
	})

	t.Run("live_lock_not_replaced", func(t *testing.T) {
		cap := withCodexLaunchCapture(t)
		r := newAnchorRepo(t)
		withRealCodexWorktreeAnchor(t)
		tree := r.tree(t, "held")
		held := session.CodexAnchorLockReason("held", os.Getppid(), "")
		lockAnchorTree(t, tree, held)
		if _, _, err := runCodexCmd(t, "-w", "held"); err == nil {
			t.Fatal("launch replaced or ignored a live lock")
		}
		codexWantLaunches(t, cap, 0, 0, 0)
		if got := observedLock(t, tree); got.Reason != held {
			t.Errorf("live lock rewritten: %q -> %q", held, got.Reason)
		}
	})
}

// TestCodexWorktreeAnchorLockReplacementRace is AC-DHR-006's race half: two
// launchers that both saw the same dead lock try to replace it at once;
// exactly one proceeds, the other is refused, and the lock left behind names
// the one that proceeded.
func TestCodexWorktreeAnchorLockReplacementRace(t *testing.T) {
	r := newAnchorRepo(t)
	tree := r.tree(t, "contested")
	pids := [2]int{os.Getpid(), os.Getppid()}

	for round := 0; round < 5; round++ {
		lockAnchorTree(t, tree, session.CodexAnchorLockReason("contested", exitedPID(t), ""))

		// Both launchers pass their initial read before either takes the guard.
		var read sync.WaitGroup
		read.Add(2)
		prevHook := codexAnchorReplaceHook
		codexAnchorReplaceHook = func() { read.Done(); read.Wait() }

		var wg sync.WaitGroup
		errs := [2]error{}
		for i := range pids {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				errs[i] = placeCodexAnchorLock(tree, pids[i], "")
			}(i)
		}
		wg.Wait()
		codexAnchorReplaceHook = prevHook

		winner := -1
		for i, err := range errs {
			if err == nil {
				if winner != -1 {
					t.Fatalf("round %d: both launchers proceeded", round)
				}
				winner = i
			}
		}
		if winner == -1 {
			t.Fatalf("round %d: both launchers refused: %v / %v", round, errs[0], errs[1])
		}
		if loser := errs[1-winner]; !strings.Contains(loser.Error(), "worktree lock") {
			t.Errorf("round %d: refusal %q does not name the worktree lock", round, loser)
		}
		got, ok := session.LockReasonPID(observedLock(t, tree).Reason)
		if !ok || got != pids[winner] {
			t.Fatalf("round %d: final lock pid %d (ok=%v), want the winner's %d", round, got, ok, pids[winner])
		}
		gitDir := anchorGit(t, tree, "rev-parse", "--absolute-git-dir")
		if _, err := os.Stat(filepath.Join(gitDir, codexAnchorReplaceGuardName)); !os.IsNotExist(err) {
			t.Fatalf("round %d: replacement guard left behind: %v", round, err)
		}
		anchorGit(t, tree, "worktree", "unlock", tree)
	}

	// The interleaving the guard exists for: the second launcher read the dead
	// lock, then the first finished replacing it. Without the re-read under
	// the guard, the second would unlock the first's fresh lock and take it.
	for round := 0; round < 3; round++ {
		lockAnchorTree(t, tree, session.CodexAnchorLockReason("contested", exitedPID(t), ""))
		var first sync.Once
		readDone, proceed, lateDone := make(chan struct{}), make(chan struct{}), make(chan struct{})
		prevHook := codexAnchorReplaceHook
		codexAnchorReplaceHook = func() {
			late := false
			first.Do(func() { late = true })
			if late {
				close(readDone)
				<-proceed
			}
		}
		var lateErr error
		go func() {
			defer close(lateDone)
			lateErr = placeCodexAnchorLock(tree, pids[1], "")
		}()
		<-readDone
		earlyErr := placeCodexAnchorLock(tree, pids[0], "")
		close(proceed)
		<-lateDone
		codexAnchorReplaceHook = prevHook

		if earlyErr != nil {
			t.Fatalf("stale round %d: the launcher that finished first was refused: %v", round, earlyErr)
		}
		if lateErr == nil {
			t.Fatalf("stale round %d: the launcher holding a stale read also proceeded", round)
		}
		if got, ok := session.LockReasonPID(observedLock(t, tree).Reason); !ok || got != pids[0] {
			t.Fatalf("stale round %d: final lock pid %d (ok=%v), want %d", round, got, ok, pids[0])
		}
		anchorGit(t, tree, "worktree", "unlock", tree)
	}
}

// ccEntryLaunch runs the `moai cc` entry with a recording launch function.
func ccEntryLaunch(t *testing.T, args ...string) (launched int, err error) {
	t.Helper()
	c := &cobra.Command{Use: "cc"}
	var errB bytes.Buffer
	c.SetErr(&errB)
	c.SetOut(&errB)
	err = runClaudeEntry(c, args, "cc", "claude", kanban.BackendClaude, func(string, string, []string) error {
		launched++
		return nil
	})
	return launched, err
}

// TestWorktreeLaunchRejectsConcurrentWriter is AC-DHR-007: a tree anchored by a
// live lock, an unreadable lock, or a live registry entry refuses both
// launchers, names the source and holder, and is left exactly as it was; a
// tree whose only lock is dead is still launchable, and the `moai cc -w`
// pre-check never writes a lock.
func TestWorktreeLaunchRejectsConcurrentWriter(t *testing.T) {
	live := os.Getppid()
	type fixture struct {
		name   string
		setup  func(t *testing.T, tree string)
		source string
		holder string
	}
	fixtures := []fixture{
		{
			name: "live_lock",
			setup: func(t *testing.T, tree string) {
				lockAnchorTree(t, tree, "claude session other (pid "+strconv.Itoa(live)+")")
			},
			source: "source: lock", holder: "pid " + strconv.Itoa(live),
		},
		{
			name:   "unreadable_lock",
			setup:  func(t *testing.T, tree string) { lockAnchorTree(t, tree, "held by an operator") },
			source: "source: lock", holder: "held by an operator",
		},
		{
			name: "registry_only",
			setup: func(t *testing.T, tree string) {
				host, _ := os.Hostname()
				data, err := json.Marshal([]session.Entry{{
					SessionID: "aaaaaaaa-1111-2222-3333-444444444444", PID: live, Host: host, CWD: tree,
					StartedAt: time.Now().UTC(), LastHeartbeat: time.Now().UTC(),
				}})
				if err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(tree, session.DefaultRegistryPath)
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, data, 0o644); err != nil {
					t.Fatal(err)
				}
			},
			source: "source: registry", holder: "pid " + strconv.Itoa(live),
		},
	}

	for _, fx := range fixtures {
		for _, launcher := range []string{"codex", "cc"} {
			t.Run(fx.name+"/"+launcher, func(t *testing.T) {
				cap := withCodexLaunchCapture(t)
				r := newAnchorRepo(t)
				withRealCodexWorktreeAnchor(t)
				tree := r.tree(t, fx.name)
				fx.setup(t, tree)
				before := snapshotAnchorTree(t, tree)

				var diag string
				var err error
				switch launcher {
				case "codex":
					_, diag, err = runCodexCmd(t, "-w", fx.name)
					codexWantLaunches(t, cap, 0, 0, 0)
					if err != nil {
						if code, ok := ResolveExitCode(err); !ok || code == 0 {
							t.Errorf("exit = (%d, %v), want non-zero", code, ok)
						}
					}
				case "cc":
					var launched int
					launched, err = ccEntryLaunch(t, "-w", fx.name)
					if launched != 0 {
						t.Errorf("cc launched %d time(s) into an anchored tree", launched)
					}
					if err != nil {
						diag = err.Error()
					}
				}
				if err == nil {
					t.Fatalf("%s -w %s: launch into an anchored tree was allowed", launcher, fx.name)
				}
				for _, want := range []string{worktreeWriterAnchoredSentinel, fx.source, fx.holder} {
					if !strings.Contains(diag, want) {
						t.Errorf("diagnostic lacks %q:\n%s", want, diag)
					}
				}
				if after := snapshotAnchorTree(t, tree); after != before {
					t.Errorf("tree changed by a refused launch:\n before %+v\n after  %+v", before, after)
				}
			})
		}
	}

	t.Run("dead_lock_allowed/codex", func(t *testing.T) {
		cap := withCodexLaunchCapture(t)
		r := newAnchorRepo(t)
		withRealCodexWorktreeAnchor(t)
		tree := r.tree(t, "dead")
		lockAnchorTree(t, tree, "claude session dead (pid "+strconv.Itoa(exitedPID(t))+")")
		if _, stderr, err := runCodexCmd(t, "-w", "dead"); err != nil {
			t.Fatalf("dead lock over-refused: %v (%s)", err, stderr)
		}
		codexWantLaunches(t, cap, 1, 1, 0)
	})

	t.Run("dead_lock_allowed/cc", func(t *testing.T) {
		withCodexLaunchCapture(t)
		r := newAnchorRepo(t)
		withRealCodexWorktreeAnchor(t)
		tree := r.tree(t, "dead")
		reason := "claude session dead (pid " + strconv.Itoa(exitedPID(t)) + ")"
		lockAnchorTree(t, tree, reason)
		launched, err := ccEntryLaunch(t, "-w", "dead")
		if err != nil || launched != 1 {
			t.Fatalf("dead lock over-refused: launched=%d err=%v", launched, err)
		}
		// The pre-check reads; the lock is Claude Code's to replace.
		if got := observedLock(t, tree); got.Reason != reason {
			t.Errorf("cc pre-check rewrote the lock: %q -> %q", reason, got.Reason)
		}
	})

	t.Run("cc_precheck_never_locks", func(t *testing.T) {
		withCodexLaunchCapture(t)
		r := newAnchorRepo(t)
		withRealCodexWorktreeAnchor(t)
		tree := r.tree(t, "free")
		launched, err := ccEntryLaunch(t, "-w", "free")
		if err != nil || launched != 1 {
			t.Fatalf("free tree refused: launched=%d err=%v", launched, err)
		}
		if got := observedLock(t, tree); got.Locked {
			t.Errorf("cc pre-check wrote a lock: %+v", got)
		}
	})
}

// TestPRMergeCleanupRefusesAnchoredCodexTree is AC-DHR-008's PR-merge half: a
// merged WT-* tree that is dirty, lock-anchored, or holds unpushed commits
// stays, with the reason named, and a Codex tree comes out exactly like a
// Claude tree in the same state.
func TestPRMergeCleanupRefusesAnchoredCodexTree(t *testing.T) {
	base := t.TempDir()
	t.Setenv(config.EnvClaudeProjectDir, base)
	live := os.Getpid()
	states := []string{"dirty", "locked", "unpushed", "clean"}
	wantCause := map[string]string{"dirty": "cause=dirty", "locked": "cause=anchored-by-lock", "unpushed": "cause=unpushed-commits"}

	type tree struct{ creator, state, path, branch string }
	var trees []tree
	var porcelain strings.Builder
	porcelain.WriteString(wtListPorcelainPrimary() + "\n")
	for _, creator := range []string{"codex", "claude"} {
		for _, state := range states {
			name := creator + "-" + state
			tr := tree{creator: creator, state: state, path: filepath.Join(base, ".claude", "worktrees", name), branch: "WT-" + name}
			if err := os.MkdirAll(tr.path, 0o755); err != nil {
				t.Fatal(err)
			}
			porcelain.WriteString(wtListEntry(tr.path, tr.branch))
			if state == "locked" {
				reason := "claude session " + name + " (pid " + strconv.Itoa(live) + ")"
				if creator == "codex" {
					reason = session.CodexAnchorLockReason(name, live, "")
				}
				porcelain.WriteString("locked " + reason + "\n")
			}
			porcelain.WriteString("\n")
			trees = append(trees, tr)
		}
	}
	byPath := map[string]tree{}
	var branches []string
	for _, tr := range trees {
		byPath[tr.path] = tr
		branches = append(branches, tr.branch)
	}
	swapPRMergeSeams(t, prMergeSeams{
		wtList:       func() (string, error) { return porcelain.String(), nil },
		ghLookPath:   func() bool { return false },
		branchMerged: func() ([]string, error) { return branches, nil },
	})
	removed := map[string]bool{}
	swapSessionWorktreeSeams(t, swSeams{
		remove: func(p string) error { removed[p] = true; return nil },
		statusPorc: func(p string) (string, error) {
			if byPath[p].state == "dirty" {
				return "?? draft.txt\n", nil
			}
			return "", nil
		},
		ignoredPorc: func(string) (string, error) { return "", nil },
		hasUnpushed: func(p string) (bool, error) { return byPath[p].state == "unpushed", nil },
	})

	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)

	got := map[string]string{}
	for _, tr := range trees {
		class := "removed"
		if !removed[tr.path] {
			class = "kept:no-reason"
			for _, line := range strings.Split(out.String(), "\n") {
				if strings.Contains(line, "worktree "+tr.path+" preserved") {
					class = "kept:" + line[strings.Index(line, "cause="):strings.Index(line, ";")]
				}
			}
		}
		if want, keep := wantCause[tr.state]; keep {
			if class != "kept:"+want {
				t.Errorf("%s: outcome %q, want kept with %s\n%s", tr.path, class, want, out.String())
			}
		} else if class != "removed" {
			t.Errorf("%s: clean merged tree outcome %q, want removed", tr.path, class)
		}
		got[tr.creator+"-"+tr.state] = class
	}
	for _, state := range states {
		if got["codex-"+state] != got["claude-"+state] {
			t.Errorf("state %s: codex %q differs from claude %q", state, got["codex-"+state], got["claude-"+state])
		}
	}
}

// TestCodexSpawnAnchorsToPanePID — on the new-window path the process that
// becomes Codex is the pane's, so the lock takes the pane pid, and a lock that
// cannot be placed closes the pane rather than leaving an unanchored writer.
func TestCodexSpawnAnchorsToPanePID(t *testing.T) {
	checkCodexSpawnAnchorsToPanePID(t)
}

// TestCodexSpawnAnchorsToPanePIDUnderLaneEnv exports a lane's launch
// variables before the same check; the spawn must not take the factory
// launch-pending path the test never set up (card t1222).
func TestCodexSpawnAnchorsToPanePIDUnderLaneEnv(t *testing.T) {
	t.Setenv(config.EnvHome, t.TempDir())
	t.Setenv(config.EnvMoaiKanbanID, "run-t1222-probe")
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-7")
	t.Setenv(config.EnvMoaiFactoryWorkers, "3")
	checkCodexSpawnAnchorsToPanePID(t)
}

func checkCodexSpawnAnchorsToPanePID(t *testing.T) {
	t.Helper()
	// defaultCodexSpawnLaunch reads os.Environ(); without this a lane's run id
	// routes the spawn into factory launch-pending registration.
	clearFactoryTestEnv(t)
	oldSpawn, oldIdentity, oldCleanup, oldAnchor := tmuxSpawnFn, codexSpawnPaneIdentityFn, codexSpawnCleanupPaneFn, codexSpawnAnchorFn
	t.Cleanup(func() {
		tmuxSpawnFn, codexSpawnPaneIdentityFn, codexSpawnCleanupPaneFn, codexSpawnAnchorFn = oldSpawn, oldIdentity, oldCleanup, oldAnchor
	})
	tmuxSpawnFn = func(string, string) (string, error) { return "%7", nil }
	codexSpawnPaneIdentityFn = func(string) (int, string, error) { return 4242, "start-x", nil }
	cleanups := 0
	codexSpawnCleanupPaneFn = func(string) error { cleanups++; return nil }

	var gotPID int
	var gotStart string
	codexSpawnAnchorFn = func(pid int, start string) error { gotPID, gotStart = pid, start; return nil }
	if err := defaultCodexSpawnLaunch(t.TempDir(), "/test/codex", nil); err != nil {
		t.Fatalf("spawn: %v", err)
	}
	if gotPID != 4242 || gotStart != "start-x" || cleanups != 0 {
		t.Fatalf("anchor got (%d, %q), cleanups %d; want the pane identity and no cleanup", gotPID, gotStart, cleanups)
	}

	codexSpawnAnchorFn = func(int, string) error { return errors.New("lock refused") }
	if err := defaultCodexSpawnLaunch(t.TempDir(), "/test/codex", nil); err == nil {
		t.Fatal("spawn succeeded although the anchor lock was refused")
	}
	if cleanups != 1 {
		t.Fatalf("pane cleanups = %d, want 1 after a refused anchor", cleanups)
	}
}
