package factory

// integration_merge_collision_test.go — the RED fixtures the run phase
// writes FIRST (card t1479, acceptance.md AC-MWQ-018 rows 13a-13d and the
// codex-P1 gitlink conversion). Every case builds its OWN fresh repository
// under t.TempDir() — codex-P2: a case that inherits the previous case's
// leftovers does not reproduce, so each case is self-contained (new
// directory, new git init). The repo shape mirrors the acceptance.md
// re-executable command sequence: one scratch repository, branch `main`
// standing in for the integration branch, branch `cand` the candidate the
// pinned SHA reads. No network; the submodule is added from a local source
// repository.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factorylane"
)

// collisionRepo is one fresh repository per case: branch main (checked out
// — its tree IS the integration worktree the collision check probes) and
// branch cand (the card branch the pinned SHA reads).
type collisionRepo struct {
	t    *testing.T
	root string
	dir  string // the single worktree
}

func newCollisionRepo(t *testing.T) *collisionRepo {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "repo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	r := &collisionRepo{t: t, root: root, dir: dir}
	r.git("init", "-q", "-b", "main")
	r.git("config", "user.email", "t@t.local")
	r.git("config", "user.name", "t")
	return r
}

// git runs one git command in the repo and fails the test on error.
func (r *collisionRepo) git(args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// on switches branches.
func (r *collisionRepo) on(branch string) { r.git("checkout", "-q", branch) }

// fromMain creates branch from main and switches to it.
func (r *collisionRepo) fromMain(branch string) { r.git("checkout", "-q", "-b", branch, "main") }

// commit writes rel=content files and commits them on the current branch.
// The add is forced, like the acceptance.md command sequence's
// `git add -f`: a candidate that starts tracking an ignored path is
// exactly the shape the fixtures exist to test, and a plain add would
// refuse it before the check ever runs.
func (r *collisionRepo) commit(message string, files map[string]string) {
	r.t.Helper()
	for rel, content := range files {
		full := filepath.Join(r.dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			r.t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			r.t.Fatal(err)
		}
		r.git("add", "-f", rel)
	}
	r.git("commit", "-q", "-m", message)
}

// addIgnored writes an untracked (ignored or plain untracked) file under
// the worktree without committing it.
func (r *collisionRepo) addIgnored(rel, content string) string {
	full := filepath.Join(r.dir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		r.t.Fatal(err)
	}
	return full
}

// tipSHA reads the branch's tip.
func (r *collisionRepo) tipSHA(branch string) string { return r.git("rev-parse", branch) }

// checksum returns the sha1 of a file, or "" when it is gone.
func (r *collisionRepo) checksum(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	cmd := exec.Command("shasum")
	cmd.Stdin = strings.NewReader(string(data))
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.Fields(string(out))[0]
}

// assertCause13 runs the collision check and asserts the refused outcome:
// at least one collision is reported, every wanted path is among them, and
// the check itself never writes anything (it reads trees and lstats the
// worktree only). The worktree is first checked out back on main — the
// integration worktree the merge verb runs in carries the integration
// branch, and a `ls-files` probe answered from a cand checkout would see
// the added paths as tracked.
func (r *collisionRepo) assertCause13(tipSHA, pinnedSHA string, wantCollisions []string) []string {
	r.t.Helper()
	r.on("main")
	got, err := FindAddedPathCollisions(r.dir, tipSHA, pinnedSHA)
	if err != nil {
		r.t.Fatalf("collision check errored: %v", err)
	}
	if len(got) == 0 {
		r.t.Fatalf("RED: collision check returned no collisions — the merge would destroy ignored or untracked bytes and the check does not yet see this shape")
	}
	for _, want := range wantCollisions {
		if !containsPath(got, want) {
			r.t.Fatalf("collision %q not reported; got %v", want, got)
		}
	}
	return got
}

func containsPath(paths []string, want string) bool {
	for _, p := range paths {
		if p == want {
			return true
		}
	}
	return false
}

func TestCollisionAddedPathItself13a(t *testing.T) {
	// AC-MWQ-018 row 13a: the candidate adds `runtime.local`; the
	// integration worktree holds an ignored file `runtime.local` with known
	// bytes. Fresh repository, own case (codex-P2). The ignored bytes are
	// written AFTER returning to main — the acceptance.md sequence's order
	// — so the candidate's commit never sees them.
	r := newCollisionRepo(t)
	r.commit("base", map[string]string{".gitignore": "runtime.local\n", "base.txt": "base"})
	r.fromMain("cand")
	r.commit("cand adds runtime.local", map[string]string{"runtime.local": "candidate-bytes"})
	r.on("main")
	ignoredPath := r.addIgnored("runtime.local", "secretbytes")
	ignoredBefore := r.checksum(ignoredPath)
	got := r.assertCause13(r.tipSHA("main"), r.tipSHA("cand"), []string{"runtime.local"})
	if got[0] != "runtime.local" {
		t.Fatalf("the added path itself must be the collision: %v", got)
	}
	if after := r.checksum(ignoredPath); after != ignoredBefore {
		t.Fatalf("the ignored bytes must be untouched: before=%s after=%s", ignoredBefore, after)
	}
}

func TestCollisionAncestorOfAddedPath13b(t *testing.T) {
	// AC-MWQ-018 row 13b (the auditor's RED fixture): the candidate adds
	// `runtime.local/payload`; the integration worktree holds an ignored
	// regular FILE `runtime.local` with known bytes. Fresh repository.
	r := newCollisionRepo(t)
	r.commit("base", map[string]string{".gitignore": "runtime.local\n", "base.txt": "base"})
	r.fromMain("cand")
	r.commit("cand adds runtime.local/payload", map[string]string{"runtime.local/payload": "p"})
	// Back on main, the ignored regular file appears (the acceptance.md
	// sequence writes it after returning to main) — the checkout above ran
	// while the path did not exist, so cand's tracking of
	// runtime.local/payload is unaffected by it.
	r.on("main")
	ignoredPath := r.addIgnored("runtime.local", "secretbytes")
	ignoredBefore := r.checksum(ignoredPath)
	r.assertCause13(r.tipSHA("main"), r.tipSHA("cand"), []string{"runtime.local"})
	if after := r.checksum(ignoredPath); after != ignoredBefore {
		t.Fatalf("runtime.local must stay a regular file with the same bytes: before=%s after=%s", ignoredBefore, after)
	}
}

func TestCollisionBeneathAddedPath13c(t *testing.T) {
	// AC-MWQ-018 row 13c: the candidate adds the FILE `runtime.local`; the
	// integration worktree holds an ignored DIRECTORY `runtime.local/`
	// containing a file with known bytes. Fresh repository. The ignored
	// directory is written after returning to main, so the candidate's
	// commit of the FILE at the same path never collides with it on disk.
	r := newCollisionRepo(t)
	r.commit("base", map[string]string{".gitignore": "runtime.local/secret\n", "base.txt": "base"})
	r.fromMain("cand")
	r.commit("cand adds runtime.local as a file", map[string]string{"runtime.local": "candidate-bytes"})
	r.on("main")
	secretPath := r.addIgnored("runtime.local/secret", "secretbytes")
	secretBefore := r.checksum(secretPath)
	r.assertCause13(r.tipSHA("main"), r.tipSHA("cand"), []string{"runtime.local"})
	if after := r.checksum(secretPath); after != secretBefore {
		t.Fatalf("the ignored directory's file must be untouched: before=%s after=%s", secretBefore, after)
	}
}

// p1FixtureRepo builds the codex-P1 fixture: main carries an INITIALISED
// submodule `node` (added from a local source repository, no network) with
// an ignored file `node/secret` in its working tree; cand flips the gitlink
// at the same path. The flip direction (file or directory) is the caller's.
func p1FixtureRepo(t *testing.T, flipToDirectory bool) (*collisionRepo, string, string) {
	t.Helper()
	r := newCollisionRepo(t)
	src := filepath.Join(r.root, "nodesrc")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	sr := &collisionRepo{t: t, root: r.root, dir: src}
	sr.git("init", "-q", "-b", "main")
	sr.git("config", "user.email", "t@t.local")
	sr.git("config", "user.name", "t")
	sr.commit("inner", map[string]string{"inner.txt": "inner"})

	r.commit("base", map[string]string{"base.txt": "base"})
	r.git("-c", "protocol.file.allow=always", "submodule", "add", "-q", src, "node")
	r.commit("add node submodule", map[string]string{".gitignore": "node/secret\n"})

	r.fromMain("cand")
	r.git("rm", "-q", "--cached", "node")
	if err := os.RemoveAll(filepath.Join(r.dir, "node")); err != nil {
		t.Fatal(err)
	}
	if flipToDirectory {
		r.commit("flip node from gitlink to directory", map[string]string{"node/payload": "p"})
	} else {
		r.commit("flip node from gitlink to file", map[string]string{"node": "candidate-bytes"})
	}

	// Back on main: the conversion branch's fixture work removed the
	// submodule working directory, and a fresh checkout of main brings the
	// gitlink back WITHOUT checking the submodule out. A real integration
	// worktree enters the window with the submodule checked out and its
	// local files in place — that is the state the merge verb runs in — so
	// the fixture recreates the ignored local file here.
	r.on("main")
	secretPath := r.addIgnored("node/secret", "secretbytes")
	secretBefore := r.checksum(secretPath)
	if secretBefore == "" {
		t.Fatal("fixture: the ignored submodule file must exist")
	}
	return r, secretPath, secretBefore
}

// TestP1GitlinkToFileConversionMustRefuse is THE codex-P1 regression test
// (spec.md:188, data-loss class). When the same path changes from a gitlink
// to a file, the path is a LEAF on both sides — so the added-path set comes
// back EMPTY and a check that reads only added paths refuses nothing. The
// merge that follows destroys the submodule working tree's ignored file.
// The check must therefore ALSO examine the type-change target and its
// local files beneath it. Observed on the scratch fixture (2026-10-05, git
// 2.54.0): merge exit 0, `node/secret` DELETED, status clean throughout.
func TestP1GitlinkToFileConversionMustRefuse(t *testing.T) {
	r, secretPath, secretBefore := p1FixtureRepo(t, false)

	// The leaf sets the check reads: `node` is a leaf on BOTH sides, and
	// the added-path set is empty — this is the shape that defeats an
	// added-path-only check.
	tipLeaves, err := readLeafSet(factorylane.ExecGitRunner{Dir: r.dir}, r.tipSHA("main"))
	if err != nil {
		t.Fatal(err)
	}
	pinnedLeaves, err := readLeafSet(factorylane.ExecGitRunner{Dir: r.dir}, r.tipSHA("cand"))
	if err != nil {
		t.Fatal(err)
	}
	if tipLeaves["node"] != "gitlink" || pinnedLeaves["node"] != "leaf" {
		t.Fatalf("fixture: node must be a gitlink leaf in main and a file leaf in cand: %v %v", tipLeaves, pinnedLeaves)
	}

	// THE assertion: the type-change target and its local files must be
	// examined even though the added-path set is empty.
	r.assertCause13(r.tipSHA("main"), r.tipSHA("cand"), []string{"node"})
	if after := r.checksum(secretPath); after != secretBefore {
		t.Fatalf("node/secret must survive: before=%s after=%s", secretBefore, after)
	}
}

func TestP1GitlinkToDirectoryConversionMustRefuse(t *testing.T) {
	// The dispatch's P1 wording: "when a merge-window entry flips a
	// submodule/gitlink to a regular directory, the conversion must examine
	// the type-change target AND its local files". Fresh repository.
	r, secretPath, secretBefore := p1FixtureRepo(t, true)
	r.assertCause13(r.tipSHA("main"), r.tipSHA("cand"), []string{"node"})
	if after := r.checksum(secretPath); after != secretBefore {
		t.Fatalf("node/secret must survive: before=%s after=%s", secretBefore, after)
	}
}

func TestCollisionSiblingStringPrefixMustNotRefuseD8(t *testing.T) {
	// D8: ancestors are PATH COMPONENTS. The candidate adds
	// `runtime.local/payload`; the worktree holds an ignored regular file
	// `runtime` — a sibling sharing a string prefix, not an ancestor — and
	// it must NOT refuse.
	r := newCollisionRepo(t)
	r.commit("base", map[string]string{".gitignore": "runtime\n", "base.txt": "base"})
	r.fromMain("cand")
	r.commit("cand adds runtime.local/payload", map[string]string{"runtime.local/payload": "p"})
	r.on("main")
	r.addIgnored("runtime", "sibling-bytes")
	got, err := FindAddedPathCollisions(r.dir, r.tipSHA("main"), r.tipSHA("cand"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("a string-prefix sibling must not refuse (D8): %v", got)
	}
}

func TestCollisionReverseLeafToDirectoryMustNotRefuseD9(t *testing.T) {
	// D9's reverse shape: the tip tracks a FILE runtime.local; the candidate
	// turns it into a DIRECTORY (leaf runtime.local/tracked). Nothing
	// untracked or ignored stands in the way, so no false refusal — the
	// merge may proceed. Fresh repository.
	r := newCollisionRepo(t)
	r.commit("base", map[string]string{"runtime.local": "tracked-file", "base.txt": "base"})
	r.fromMain("cand")
	r.git("rm", "-q", "runtime.local")
	r.commit("cand turns runtime.local into a directory", map[string]string{"runtime.local/tracked": "t"})
	r.on("main")
	got, err := FindAddedPathCollisions(r.dir, r.tipSHA("main"), r.tipSHA("cand"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("the reverse leaf-to-directory shape must not refuse (D9): %v", got)
	}
}

func TestCollisionGitlinkPointerChangeMustNotRefuseD9(t *testing.T) {
	// D9's gitlink shape: a submodule pointer MOVING (gitlink -> gitlink at
	// a different commit) is a mode-class-stable leaf on both sides — the
	// added-path set is empty and no type change happens, so a checked-out
	// submodule with local files must not refuse. Fresh repository.
	r := newCollisionRepo(t)
	src := filepath.Join(r.root, "nodesrc")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	sr := &collisionRepo{t: t, root: r.root, dir: src}
	sr.git("init", "-q", "-b", "main")
	sr.git("config", "user.email", "t@t.local")
	sr.git("config", "user.name", "t")
	sr.commit("inner", map[string]string{"inner.txt": "inner"})

	r.commit("base", map[string]string{"base.txt": "base"})
	r.git("-c", "protocol.file.allow=always", "submodule", "add", "-q", src, "node")
	r.commit("add node submodule", map[string]string{})
	secretPath := r.addIgnored("node/secret", "secretbytes")
	secretBefore := r.checksum(secretPath)

	// Advance the submodule one commit; move the pointer on cand.
	sr.commit("inner2", map[string]string{"inner2.txt": "inner2"})
	newPointer := sr.tipSHA("main")
	r.fromMain("cand")
	r.git("update-index", "--cacheinfo", fmt.Sprintf("160000,%s,node", newPointer))
	r.git("commit", "-q", "-m", "move the node pointer")

	r.on("main")
	got, err := FindAddedPathCollisions(r.dir, r.tipSHA("main"), r.tipSHA("cand"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("a gitlink pointer move must not refuse (D9): %v", got)
	}
	if after := r.checksum(secretPath); after != secretBefore {
		t.Fatalf("the submodule's local bytes must survive a pointer move: before=%s after=%s", secretBefore, after)
	}
}

func TestCollisionSymlinkAtAddedPathRefusesO5(t *testing.T) {
	// O5: a SYMLINK at an added path is bytes the merge would overwrite —
	// the existence probe is lstat-based, so the symlink refuses.
	r := newCollisionRepo(t)
	r.commit("base", map[string]string{"base.txt": "base"})
	r.fromMain("cand")
	r.commit("cand adds runtime.local", map[string]string{"runtime.local": "candidate-bytes"})
	r.on("main")
	if err := os.Symlink("base.txt", filepath.Join(r.dir, "runtime.local")); err != nil {
		t.Fatal(err)
	}
	got, err := FindAddedPathCollisions(r.dir, r.tipSHA("main"), r.tipSHA("cand"))
	if err != nil {
		t.Fatal(err)
	}
	if !containsPath(got, "runtime.local") {
		t.Fatalf("an ignored symlink at the added path must refuse (O5): %v", got)
	}
}

func TestCollisionDirectoryReplacedByFile13d(t *testing.T) {
	// AC-MWQ-018 row 13d (the second auditor RED fixture): the tip tracks
	// the DIRECTORY runtime.local/ with a tracked child; the candidate
	// replaces it with the FILE runtime.local (the leaf is in the pinned
	// set, not the tip's — a directory-to-leaf change COUNTS as an added
	// path, v0.9.0 D5); the worktree holds an ignored file
	// runtime.local/secret. Fresh repository.
	r := newCollisionRepo(t)
	r.commit("base", map[string]string{".gitignore": "runtime.local/secret\n", "runtime.local/tracked": "t"})
	r.fromMain("cand")
	r.git("rm", "-q", "-r", "runtime.local")
	r.commit("cand replaces the directory with a file", map[string]string{"runtime.local": "candidate-bytes"})
	// Back on main, the ignored file appears inside the (still tracked on
	// main) directory — the acceptance.md sequence's order.
	r.on("main")
	secretPath := r.addIgnored("runtime.local/secret", "secretbytes")
	secretBefore := r.checksum(secretPath)
	r.assertCause13(r.tipSHA("main"), r.tipSHA("cand"), []string{"runtime.local"})
	if after := r.checksum(secretPath); after != secretBefore {
		t.Fatalf("runtime.local/secret must survive: before=%s after=%s", secretBefore, after)
	}
}
