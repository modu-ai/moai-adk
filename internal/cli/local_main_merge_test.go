package cli

// local_main_merge_test.go — SPEC-LOCAL-MAIN-FLOW-001 (card t1616), M1 step 1
// RED tests for the landing merge on the primary checkout's local main
// (REQ-LMF-003, REQ-LMF-004, REQ-LMF-006; plan §B3, §B4, §B5).
//
// Each merge test drives the exported factory.RunMergeStep directly, with the
// primary checkout as IntegrationWorktree (the seam). The landing verb's
// surface gate (integrationMergeWorktree) refuses the primary before the step
// runs, so a verb-level call cannot reach these assertions; the surface is
// covered by local_main_verb_test.go. Apart from the worktree, the seam passes
// the verb's operands (integration_merge.go:100-106) and the same card-gate
// read (integrationReadMergeCardForRun). The workflow gate is set as each
// fixture states it, but the step does not read it: the surface decision
// belongs to the verb, and the seam bypasses it.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

const (
	lmfCard    = "t1"
	lmfBranch  = "main" // the configured integration branch (git_strategy.manual.develop_branch)
	lmfSession = "sess-lane-1"

	// lmfMovedHeadHook moves HEAD to another branch at the merge commit after
	// the merge has written it. SHA equality still holds; the symbolic HEAD no
	// longer names the configured branch.
	lmfMovedHeadHook = "#!/bin/sh\ngit update-ref refs/heads/lmf-moved HEAD && git symbolic-ref HEAD refs/heads/lmf-moved\n"
)

// lmfSpec describes one primary-surface landing fixture.
type lmfSpec struct {
	gitignore string            // committed ignore rules; "" commits no .gitignore
	cardPaths []string          // paths the card adds on its WT- branch
	forceAdd  bool              // git add -f: the card paths match the ignore rules
	local     map[string]string // ignored bytes the primary holds (written after the card branch exists)
	holder    string            // session holding the window; "" is the lane itself
	hook      string            // optional post-merge hook body installed in the primary
}

// lmfGate writes workflow.local_main_integration.enabled under root.
func lmfGate(t *testing.T, root string, enabled bool) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := fmt.Sprintf("workflow:\n    local_main_integration:\n        enabled: %t\n", enabled)
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// lmfExclude keeps the fixture's runtime state (.moai) out of git status, as a
// real primary checkout keeps its own runtime directory out of status.
func lmfExclude(t *testing.T, root string) {
	t.Helper()
	info := filepath.Join(root, ".git", "info")
	if err := os.MkdirAll(info, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(filepath.Join(info, "exclude"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString("/.moai/\n"); err != nil {
		t.Fatal(err)
	}
}

// lmfRepo builds a primary checkout whose HEAD names main. The git-flow
// configuration names main as the integration branch, and the gate is set as
// given. The base commit tracks base.txt and, when given, .gitignore.
func lmfRepo(t *testing.T, gateOn bool, gitignore string) (string, *factory.BacklogStore) {
	t.Helper()
	root, store := fcFixture(t)
	sdClearLaneEnv(t)
	writeGitStrategyFixture(t, root, "git-flow", lmfBranch)
	lmfGate(t, root, gateOn)
	lmfExclude(t, root)
	fcQueue(t, store, factory.BacklogStatePicked)
	fcGit(t, root, "symbolic-ref", "HEAD", "refs/heads/"+lmfBranch)
	if err := os.WriteFile(filepath.Join(root, "base.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	files := []string{"base.txt"}
	if gitignore != "" {
		if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte(gitignore+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		files = append(files, ".gitignore")
	}
	fcGit(t, root, append([]string{"add"}, files...)...)
	fcGit(t, root, "commit", "-q", "-m", "base")
	return root, store
}

// lmfMergeFixture adds the landing shape to lmfRepo: a card worktree WT-alpha
// cut from main outside the primary, the card's commit, the primary's ignored
// bytes, a valid re-measure record for the card tree, the card merge-ready
// under a live lease, and the integration window held by the lane (or by
// spec.holder). It returns the primary checkout and the card worktree.
func lmfMergeFixture(t *testing.T, gateOn bool, spec lmfSpec) (string, string) {
	t.Helper()
	root, _ := lmfRepo(t, gateOn, spec.gitignore)
	cardWT := filepath.Join(t.TempDir(), "t1")
	fcGit(t, root, "worktree", "add", "-q", "-b", "WT-alpha", cardWT)
	for _, p := range spec.cardPaths {
		full := filepath.Join(cardWT, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("card:"+p+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	add := []string{"add"}
	if spec.forceAdd {
		add = append(add, "-f")
	}
	fcGit(t, cardWT, append(add, spec.cardPaths...)...)
	fcGit(t, cardWT, "commit", "-q", "-m", "card change")
	for p, content := range spec.local {
		full := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if spec.hook != "" {
		hook := filepath.Join(root, ".git", "hooks", "post-merge")
		if err := os.MkdirAll(filepath.Dir(hook), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(hook, []byte(spec.hook), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(hook, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := factory.RunRemeasure(root, cardWT, lmfBranch, "true"); err != nil {
		t.Fatalf("fixture: remeasure the card tree: %v", err)
	}
	now := time.Now().UTC()
	fcPlace(t, root, homestate.Card{
		CardID: lmfCard, RunID: fcRun,
		State: homestate.CardMergeReady, Stage: homestate.CardMergeReady,
		OwnerLabel: "lane-1", LeaseHolder: "lane-1",
		LeaseExpiresAt: now.Add(time.Hour).Format(time.RFC3339),
		HeartbeatAt:    now.Format(time.RFC3339),
		WorktreePath:   cardWT, Version: 1,
	})
	holder := spec.holder
	if holder == "" {
		holder = lmfSession
	}
	sdHoldWindow(t, root, holder, "lane-1", lmfBranch, factory.BranchSourceConfig, root, lmfCard)
	t.Setenv(config.EnvClaudeCodeSessionID, lmfSession)
	sdLaneEnv(t, "lane-1", "")
	if status := lmfStatus(t, root); status != "" {
		t.Fatalf("fixture: the primary checkout must be clean before the landing, got:\n%s", status)
	}
	return root, cardWT
}

// lmfVerb runs the landing verb for the fixture's card as the lane session.
func lmfVerb(t *testing.T) error {
	t.Helper()
	_, _, err := runIntegrationMerge(t, "merge", "--card", lmfCard, "--run", fcRun, "--session", lmfSession)
	return err
}

// lmfStep runs the merge step directly for the fixture's card, as the landing
// verb runs it after its surface gate (integration_merge.go:100-106): the
// operands the verb passes, the primary checkout as the integration worktree,
// and the card-gate read the verb's ReadCard seam performs. The surface gate is
// bypassed by construction, so a failure here belongs to the step.
func lmfStep(t *testing.T, root string) (string, error) {
	t.Helper()
	lane, err := factoryLaneLabelFromEnv("merge")
	if err != nil {
		t.Fatalf("fixture: the lane label must resolve: %v", err)
	}
	ctx := context.Background()
	return factory.RunMergeStep(factory.MergeStepInput{
		Root:                root,
		IntegrationWorktree: root,
		IntegrationBranch:   lmfBranch,
		CardID:              lmfCard,
		CallerSessionID:     lmfSession,
	}, factory.MergeStepSeams{
		ReadCard: func(cardID string) (factory.MergeCardState, error) {
			return integrationReadMergeCardForRun(ctx, root, fcRun, cardID, lane)
		},
	})
}

// lmfStatus reads the primary's status with untracked files listed one by one.
func lmfStatus(t *testing.T, root string) string {
	t.Helper()
	return fcGit(t, root, "status", "--porcelain=v1", "--untracked-files=all")
}

// lmfHead reads the primary's HEAD commit.
func lmfHead(t *testing.T, root string) string {
	t.Helper()
	return fcGit(t, root, "rev-parse", "HEAD")
}

// lmfAssertFile asserts that path holds exactly want.
func lmfAssertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("the ignored bytes must survive the landing: %v", err)
	}
	if string(got) != want {
		t.Fatalf("%s holds %q, want %q", path, got, want)
	}
}

// lmfCaseInsensitive reports whether the volume holding the test temp dir
// resolves names case-insensitively.
func lmfCaseInsensitive(t *testing.T) bool {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "LmfCase.probe"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := os.Stat(filepath.Join(dir, "lmfcase.probe"))
	return err == nil
}

// lmfExpectCollision asserts that the merge step refused with
// MergeExitCollision, named the colliding path, performed no merge, and
// released the window.
func lmfExpectCollision(t *testing.T, root, path string) {
	t.Helper()
	before := lmfHead(t, root)
	_, err := lmfStep(t, root)
	if err == nil {
		t.Fatalf("a candidate that would overwrite ignored bytes at %q must be refused", path)
	}
	if code, ok := factory.MergeExitCode(err); !ok || code != factory.MergeExitCollision {
		t.Fatalf("the refusal must carry MergeExitCollision (%d), got code %d (ok=%v): %v", factory.MergeExitCollision, code, ok, err)
	}
	if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(path)) {
		t.Fatalf("the refusal must name the colliding path %q: %v", path, err)
	}
	if after := lmfHead(t, root); after != before {
		t.Fatalf("a collision refusal must perform no merge: HEAD %s -> %s", before, after)
	}
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("a collision refusal must release the window: %+v", lock)
	}
}

// --- Primary merge: the landing verb on the primary checkout (REQ-LMF-003) ---

func TestLocalMainMergeMergesIntoPrimary(t *testing.T) {
	root, cardWT := lmfMergeFixture(t, true, lmfSpec{cardPaths: []string{"alpha.txt"}})
	before := lmfHead(t, root)
	cardTip := fcGit(t, cardWT, "rev-parse", "HEAD")
	if _, err := lmfStep(t, root); err != nil {
		t.Fatalf("the step must merge the card into the primary's local main: %v", err)
	}
	if first := fcGit(t, root, "rev-parse", "HEAD^1"); first != before {
		t.Fatalf("the first parent of the new HEAD must be the pre-merge HEAD %s, got %s", before, first)
	}
	if second := fcGit(t, root, "rev-parse", "HEAD^2"); second != cardTip {
		t.Fatalf("the second parent must be the card tip %s, got %s", cardTip, second)
	}
	// B3 step 6: the symbolic HEAD still names the configured branch, because
	// SHA equality alone does not prove the branch did not change.
	if sym := fcGit(t, root, "symbolic-ref", "HEAD"); sym != "refs/heads/"+lmfBranch {
		t.Fatalf("HEAD must still name %s after the merge (the tool never switches branches), got %q", lmfBranch, sym)
	}
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("the window must be released after a successful landing: %+v", lock)
	}
}

func TestLocalMainMergeRefusesDirtyPrimary(t *testing.T) {
	root, _ := lmfMergeFixture(t, true, lmfSpec{cardPaths: []string{"alpha.txt"}})
	if err := os.WriteFile(filepath.Join(root, "base.txt"), []byte("base\nuncommitted\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := lmfHead(t, root)
	_, err := lmfStep(t, root)
	if err == nil {
		t.Fatalf("a primary checkout with an uncommitted change must refuse the landing")
	}
	if code, ok := factory.MergeExitCode(err); !ok || code != factory.MergeExitWorktreeDirty {
		t.Fatalf("the refusal must carry MergeExitWorktreeDirty (%d), got code %d (ok=%v): %v", factory.MergeExitWorktreeDirty, code, ok, err)
	}
	// REQ-LMF-004 guidance: the remedies are named and stashing is forbidden.
	if !strings.Contains(err.Error(), "Do not stash") {
		t.Fatalf("the guidance must forbid stashing (plan §B4): %v", err)
	}
	if after := lmfHead(t, root); after != before {
		t.Fatalf("a dirty-primary refusal must perform no merge: HEAD %s -> %s", before, after)
	}
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("a dirty-primary refusal must release the window: %+v", lock)
	}
}

func TestLocalMainMergeRefusesMovedHead(t *testing.T) {
	// The post-merge hook moves HEAD to another branch at the merge commit.
	// B3 step 6 must hold the landing: the merge commit exists, but the
	// symbolic HEAD no longer names the configured branch.
	root, cardWT := lmfMergeFixture(t, true, lmfSpec{cardPaths: []string{"alpha.txt"}, hook: lmfMovedHeadHook})
	cardTip := fcGit(t, cardWT, "rev-parse", "HEAD")
	_, err := lmfStep(t, root)
	if err == nil {
		t.Fatalf("a HEAD that moved to another branch during the merge must hold the landing (plan §B3 step 6)")
	}
	if code, ok := factory.MergeExitCode(err); !ok || code != factory.MergeExitPostMerge {
		t.Fatalf("the hold must carry MergeExitPostMerge (%d), got code %d (ok=%v): %v", factory.MergeExitPostMerge, code, ok, err)
	}
	if second := fcGit(t, root, "rev-parse", "HEAD^2"); second != cardTip {
		t.Fatalf("the merge commit must exist before the hold: second parent %s, want %s", second, cardTip)
	}
	if sym := fcGit(t, root, "symbolic-ref", "HEAD"); sym != "refs/heads/lmf-moved" {
		t.Fatalf("the tool must not switch HEAD back to the configured branch: got %q", sym)
	}
}

func TestLocalMainMergeStatusSetUnchanged(t *testing.T) {
	// REQ-LMF-006 on the primary surface: the status set after the merge equals
	// the set before it (empty, by the fully-clean rule of plan §B4).
	root, _ := lmfMergeFixture(t, true, lmfSpec{cardPaths: []string{"alpha.txt"}})
	before := lmfStatus(t, root)
	if _, err := lmfStep(t, root); err != nil {
		t.Fatalf("the landing step must complete: %v", err)
	}
	if after := lmfStatus(t, root); after != before {
		t.Fatalf("the status set must be unchanged by the landing (REQ-LMF-006): before %q, after %q", before, after)
	}
}

func TestLocalMainMergePreservesIgnoredFile(t *testing.T) {
	// An ignored file off the merge path must neither block the landing nor be
	// touched by it.
	root, _ := lmfMergeFixture(t, true, lmfSpec{
		gitignore: "build/",
		cardPaths: []string{"alpha.txt"},
		local:     map[string]string{"build/cache.bin": "cache-keep"},
	})
	if _, err := lmfStep(t, root); err != nil {
		t.Fatalf("an ignored file off the merge path must not block the landing: %v", err)
	}
	lmfAssertFile(t, filepath.Join(root, "build", "cache.bin"), "cache-keep")
}

func TestLocalMainMergeRefusesIgnoredAncestor(t *testing.T) {
	// Shape A (probe-shapes.sh, E-79): the primary's ignored file `runtime` is
	// an ancestor of the card's added path `runtime/x`. The planned command
	// overwrites it; the collision check must refuse.
	root, _ := lmfMergeFixture(t, true, lmfSpec{
		gitignore: "runtime",
		cardPaths: []string{"runtime/x"},
		forceAdd:  true,
		local:     map[string]string{"runtime": "local-A"},
	})
	lmfExpectCollision(t, root, "runtime")
	lmfAssertFile(t, filepath.Join(root, "runtime"), "local-A")
}

func TestLocalMainMergeRefusesFileOverIgnoredDirectory(t *testing.T) {
	// Shape B (probe-shapes.sh, E-79): the card adds the file `runtime` over the
	// primary's ignored directory `runtime/`, which holds `runtime/secret`.
	root, _ := lmfMergeFixture(t, true, lmfSpec{
		gitignore: "runtime/",
		cardPaths: []string{"runtime"},
		local:     map[string]string{"runtime/secret": "local-B"},
	})
	lmfExpectCollision(t, root, "runtime")
	lmfAssertFile(t, filepath.Join(root, "runtime", "secret"), "local-B")
}

func TestLocalMainMergeCaseAliasOutcome(t *testing.T) {
	// Shape C (probe-shapes.sh, E-79): the primary's ignored `foo.txt` and the
	// card's `Foo.txt` are one file on a case-insensitive volume. The outcome
	// depends on the volume (plan §B3 step 4a, REQ-LMF-004).
	insensitive := lmfCaseInsensitive(t)
	root, _ := lmfMergeFixture(t, true, lmfSpec{
		gitignore: "foo.txt",
		cardPaths: []string{"Foo.txt"},
		forceAdd:  true,
		local:     map[string]string{"foo.txt": "local-C"},
	})
	if insensitive {
		lmfExpectCollision(t, root, "foo.txt")
		lmfAssertFile(t, filepath.Join(root, "foo.txt"), "local-C")
		return
	}
	if _, err := lmfStep(t, root); err != nil {
		t.Fatalf("on a case-sensitive volume the landing completes and both files keep their own bytes: %v", err)
	}
	lmfAssertFile(t, filepath.Join(root, "foo.txt"), "local-C")
	lmfAssertFile(t, filepath.Join(root, "Foo.txt"), "card:Foo.txt\n")
}
