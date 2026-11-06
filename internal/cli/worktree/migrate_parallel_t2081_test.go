package worktree

// Card t2081 — is it safe to run `git worktree move` for several trees of ONE
// repository at the same time?
//
// The migration moves trees one after another. t2070 measured eight moves at
// 300ms together against 724ms one by one with no failure, in two samples. Two
// samples do not prove git's registration handling, so this file is the proof
// the card asks for before any concurrent move is built: repeated rounds of
// concurrent moves with other git processes reading the same repository and the
// same trees at the same time, checked afterwards against git's own view
// (`worktree list`, `worktree prune --dry-run`, the gitdir back-pointers).
// A positive control shows the check can see a race: moving ONE tree twice at
// once must fail for exactly one of the two.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// stressWorld is a repository with n registered trees under <repo>/.moai/worktrees
// and a second folder to move them to and back.
type stressWorld struct {
	repo  string
	trees []string // current location of tree i
	home  string   // the folder of the other spelling
}

func newStressWorld(t *testing.T, n int) *stressWorld {
	t.Helper()
	root := m6Real(t, t.TempDir())
	repo := filepath.Join(root, "stress-repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	m6Git(t, repo, "init", "-q", "--template=", "-b", "dev")
	m6Git(t, repo, "config", "user.email", "t2081@example.com")
	m6Git(t, repo, "config", "user.name", "T2081")
	m6Git(t, repo, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m6Git(t, repo, "add", "-A")
	m6Git(t, repo, "commit", "-q", "-m", "seed")
	w := &stressWorld{repo: repo, home: filepath.Join(root, "home-root")}
	if err := os.MkdirAll(w.home, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		p := filepath.Join(repo, ".moai", "worktrees", fmt.Sprintf("s%02d", i))
		m6Git(t, repo, "worktree", "add", "-q", "-b", fmt.Sprintf("WT-s%02d", i), "--", p)
		w.trees = append(w.trees, p)
	}
	return w
}

// destinationOf alternates a tree between its two spellings.
func (w *stressWorld) other(i int, from string) string {
	if strings.HasPrefix(from, w.home) {
		return filepath.Join(w.repo, ".moai", "worktrees", fmt.Sprintf("s%02d", i))
	}
	return filepath.Join(w.home, fmt.Sprintf("s%02d", i))
}

// verifyAgainstGit compares the repository's registrations with where the
// trees really are, using git's own commands.
func (w *stressWorld) verifyAgainstGit(t *testing.T, round int) {
	t.Helper()
	if err := w.checkAgainstGit(round); err != nil {
		t.Fatal(err)
	}
}

// checkAgainstGit is verifyAgainstGit as an error, so a control can feed it a
// damaged repository and demand that it notices.
func (w *stressWorld) checkAgainstGit(round int) error {
	listing, err := exec.Command("git", "-C", w.repo, "worktree", "list", "--porcelain").Output()
	if err != nil {
		return fmt.Errorf("round %d: worktree list: %v", round, err)
	}
	var listed []string
	for _, line := range strings.Split(string(listing), "\n") {
		if p, ok := strings.CutPrefix(line, "worktree "); ok && canonicalTreePath(p) != canonicalTreePath(w.repo) {
			listed = append(listed, canonicalTreePath(p))
		}
	}
	var want []string
	for _, p := range w.trees {
		want = append(want, canonicalTreePath(p))
	}
	sort.Strings(listed)
	sort.Strings(want)
	if strings.Join(listed, "\n") != strings.Join(want, "\n") {
		return fmt.Errorf("round %d: git lists different trees than where they are\n listed: %v\n actual: %v", round, listed, want)
	}
	// Nothing git would prune: a half-updated registration shows up here.
	prune, perr := exec.Command("git", "-C", w.repo, "worktree", "prune", "--dry-run", "-v").CombinedOutput()
	if perr != nil || strings.TrimSpace(string(prune)) != "" {
		return fmt.Errorf("round %d: git would prune registrations (err %v): %s", round, perr, prune)
	}
	// Every tree answers for itself from where it stands, and its back-pointer
	// in the repository names that place.
	for i, p := range w.trees {
		top, terr := exec.Command("git", "-C", p, "rev-parse", "--show-toplevel").Output()
		if terr != nil || canonicalTreePath(strings.TrimSpace(string(top))) != canonicalTreePath(p) {
			return fmt.Errorf("round %d: tree %d at %s answers toplevel %q (err %v)", round, i, p, top, terr)
		}
		gitdir, gerr := exec.Command("git", "-C", p, "rev-parse", "--git-dir").Output()
		if gerr != nil {
			return fmt.Errorf("round %d: tree %d git-dir: %v", round, i, gerr)
		}
		back, rerr := os.ReadFile(filepath.Join(strings.TrimSpace(string(gitdir)), "gitdir"))
		if rerr != nil {
			return fmt.Errorf("round %d: tree %d back-pointer: %v", round, i, rerr)
		}
		if canonicalTreePath(filepath.Dir(strings.TrimSpace(string(back)))) != canonicalTreePath(p) {
			return fmt.Errorf("round %d: tree %d back-pointer %q does not name %s", round, i, back, p)
		}
	}
	return nil
}

// TestGitWorktreeMove_TheConsistencyCheckSeesADamagedRegistration is the other
// control: the check passing above means something only if it fails on a
// repository whose registration is wrong. The back-pointer of one tree is
// pointed at a place the tree is not.
func TestGitWorktreeMove_TheConsistencyCheckSeesADamagedRegistration(t *testing.T) {
	w := newStressWorld(t, 2)
	if err := w.checkAgainstGit(0); err != nil {
		t.Fatalf("a healthy repository must pass: %v", err)
	}
	gitdir, err := exec.Command("git", "-C", w.trees[0], "rev-parse", "--git-dir").Output()
	if err != nil {
		t.Fatal(err)
	}
	pointer := filepath.Join(strings.TrimSpace(string(gitdir)), "gitdir")
	if werr := os.WriteFile(pointer, []byte(filepath.Join(w.home, "nowhere", ".git")+"\n"), 0o644); werr != nil {
		t.Fatal(werr)
	}
	if cerr := w.checkAgainstGit(1); cerr == nil {
		t.Fatalf("a registration whose back-pointer names a place the tree is not was reported consistent")
	}
}

// TestGitWorktreeMove_ConcurrentMovesOfDifferentTreesAreConsistent runs rounds
// of all trees moving at once (to the other spelling and back) while other git
// processes read the repository and the trees.
func TestGitWorktreeMove_ConcurrentMovesOfDifferentTreesAreConsistent(t *testing.T) {
	trees, rounds := 12, 8
	if v, _ := strconv.Atoi(os.Getenv("MOAI_T2081_STRESS_TREES")); v > 0 {
		trees = v
	}
	if v, _ := strconv.Atoi(os.Getenv("MOAI_T2081_STRESS_ROUNDS")); v > 0 {
		rounds = v
	}
	w := newStressWorld(t, trees)
	w.verifyAgainstGit(t, 0)

	for round := 1; round <= rounds; round++ {
		stop := make(chan struct{})
		var readers sync.WaitGroup
		var readerFailures int32
		spawnReader := func(args ...string) {
			readers.Add(1)
			go func() {
				defer readers.Done()
				for {
					select {
					case <-stop:
						return
					default:
					}
					// A reader may see a registration mid-update and report a
					// transient error; what it must never do is wedge or damage
					// the repository, which the checks after the round decide.
					if out, err := exec.Command("git", args...).CombinedOutput(); err != nil && strings.Contains(string(out), "index.lock") {
						atomic.AddInt32(&readerFailures, 1)
					}
				}
			}()
		}
		spawnReader("-C", w.repo, "worktree", "list", "--porcelain")
		spawnReader("-C", w.repo, "worktree", "list", "--porcelain")
		for i := 0; i < len(w.trees) && i < 4; i++ {
			spawnReader("-C", w.trees[i], "status", "--porcelain")
		}

		next := make([]string, len(w.trees))
		errs := make([]error, len(w.trees))
		var movers sync.WaitGroup
		for i := range w.trees {
			next[i] = w.other(i, w.trees[i])
			movers.Add(1)
			go func(i int) {
				defer movers.Done()
				errs[i] = forwardGitMoveReal(w.repo, w.trees[i], next[i])
			}(i)
		}
		movers.Wait()
		close(stop)
		readers.Wait()

		for i, err := range errs {
			if err != nil {
				t.Fatalf("round %d: concurrent move of tree %d (%s -> %s) failed: %v", round, i, w.trees[i], next[i], err)
			}
		}
		w.trees = next
		w.verifyAgainstGit(t, round)
	}
	t.Logf("%d trees x %d rounds of concurrent moves with concurrent readers: git's view stayed consistent every round", trees, rounds)
}

// forwardGitMoveReal is the plain `git worktree move`, immune to a test that
// replaced the forwardGitMove seam.
func forwardGitMoveReal(root, from, to string) error {
	_, err := gitOutputWithin(gitMoveTimeout, root, "worktree", "move", from, to)
	return err
}

// TestGitWorktreeMove_TheSameTreeMovedTwiceAtOnceFailsForOne is the positive
// control: the consistency check above is only evidence if a real race makes
// something fail. Two moves of ONE tree at the same moment must not both
// succeed, and the repository must still agree with the disk afterwards.
func TestGitWorktreeMove_TheSameTreeMovedTwiceAtOnceFailsForOne(t *testing.T) {
	w := newStressWorld(t, 1)
	src := w.trees[0]
	dstA := filepath.Join(w.home, "a")
	dstB := filepath.Join(w.home, "b")
	var wg sync.WaitGroup
	errs := make([]error, 2)
	start := make(chan struct{})
	for i, dst := range []string{dstA, dstB} {
		wg.Add(1)
		go func(i int, dst string) {
			defer wg.Done()
			<-start
			errs[i] = forwardGitMoveReal(w.repo, src, dst)
		}(i, dst)
	}
	close(start)
	wg.Wait()
	ok := 0
	for _, err := range errs {
		if err == nil {
			ok++
		}
	}
	if ok != 1 {
		t.Fatalf("two simultaneous moves of one tree: %d succeeded (errors %v); exactly one must", ok, errs)
	}
	// Whichever won, git and the disk agree.
	if _, err := os.Stat(dstA); err == nil {
		w.trees[0] = dstA
	} else {
		w.trees[0] = dstB
	}
	w.verifyAgainstGit(t, 1)
}
