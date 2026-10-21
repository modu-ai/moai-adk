package cli

// integration_base_freshness_test.go — the CLI half of the pre-integration
// base-freshness gate (SPEC-INTEGRATION-STALE-BASE-001, card t1834): the two
// call sites (`moai integration acquire`, `moai factory complete`), the
// adapters' failure posture, the sentinel and exit code, and the adoption
// exemption. The pure gate is pinned in internal/factory.
//
// Two fixtures. The acquire fixture is one repository whose `main` is the
// integration branch plus a card worktree carrying a commit and a valid
// re-measure record; landings are added to `main` after the card branched so
// the card's tree never changes while its distance grows. The complete tests
// reuse mwq19Fixture, the substantive-completion environment of card t1479.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorylane"
	"github.com/modu-ai/moai-adk/internal/gitenv"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

const bfcSentinel = "BASE_FRESHNESS_VIOLATION:"

func bfcGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(gitenv.Env(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t.local",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t.local",
		"GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v: %s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

type bfcFixture struct {
	t    *testing.T
	root string
	repo string
	card string
	base string
}

func newBFCFixture(t *testing.T) *bfcFixture {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := &bfcFixture{t: t, root: root, repo: filepath.Join(root, "repo"), card: filepath.Join(root, "card")}
	if err := os.MkdirAll(f.repo, 0o755); err != nil {
		t.Fatal(err)
	}
	bfcGit(t, f.repo, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(f.repo, "base.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bfcGit(t, f.repo, "add", "base.txt")
	bfcGit(t, f.repo, "commit", "-q", "-m", "base")
	f.base = bfcGit(t, f.repo, "rev-parse", "HEAD")
	bfcGit(t, f.repo, "worktree", "add", "-q", "-b", "WT-card", f.card, "main")
	if err := os.WriteFile(filepath.Join(f.card, "card.txt"), []byte("card work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bfcGit(t, f.card, "add", "card.txt")
	bfcGit(t, f.card, "commit", "-q", "-m", "card work")
	// The gate measures the configured integration target only (audit iter10
	// H1). The fixture is a github-flow project: its develop_branch is empty by
	// design and the flow-scoped IntegrationTarget resolves to `main`.
	writeGitStrategyFixture(t, root, "github-flow", `""`)
	// A card session's environment names the lane it runs in; the fixture pins
	// only what the verbs under test read.
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	return f
}

func (f *bfcFixture) tree() string { return bfcGit(f.t, f.card, "rev-parse", "HEAD^{tree}") }

// writeRecord stores a valid re-measure record keyed to the card's tree and
// naming base as the absorbed base.
func (f *bfcFixture) writeRecord(base string) {
	f.t.Helper()
	tree := f.tree()
	rec := factory.RemeasureRecord{
		Tree: tree, Base: base, Command: "true", ExitCode: 0, BuildIdentity: "moai test",
		RecordedAt: time.Date(2026, 10, 18, 9, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}
	if err := factory.WriteRemeasureRecord(f.root, tree, rec); err != nil {
		f.t.Fatal(err)
	}
}

func (f *bfcFixture) land(n int) {
	f.t.Helper()
	for i := 0; i < n; i++ {
		bfcGit(f.t, f.repo, "commit", "-q", "--allow-empty", "-m", fmt.Sprintf("landing %d", i))
	}
}

func (f *bfcFixture) acquire(args ...string) (string, string, error) {
	return runIntegrationSplit(f.t, f.card, f.root, append([]string{"acquire"}, args...)...)
}

// bfcWriteWorkflow writes the project's workflow.yaml section.
func bfcWriteWorkflow(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func bfcGateYAML(enabled bool, k int) string {
	return fmt.Sprintf("workflow:\n  base_freshness_gate:\n    enabled: %t\n    max_behind_landings: %d\n", enabled, k)
}

// runIntegrationSplit is runIntegrationIn with the two streams kept apart: the
// gate's one-line notices go to standard error and are counted there.
func runIntegrationSplit(t *testing.T, worktree, root string, args ...string) (string, string, error) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(worktree); err != nil {
		t.Fatalf("chdir %s: %v", worktree, err)
	}
	defer func() { _ = os.Chdir(prev) }()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	cmd := newIntegrationCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	walk(cmd, func(c *cobra.Command) { c.SilenceUsage = true; c.SilenceErrors = true })
	execErr := cmd.Execute()
	return out.String(), errBuf.String(), execErr
}

func freshnessLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(line, "base-freshness") {
			lines = append(lines, line)
		}
	}
	return lines
}

func windowBytes(t *testing.T, root string) string {
	t.Helper()
	data, err := os.ReadFile(lockPathFor(root))
	if errors.Is(err, os.ErrNotExist) {
		return "<absent>"
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// requireFreshnessRefusal asserts the refusal shape of AC-ISB-011: the message
// starts with the sentinel and the process exit code is 16, a code the merge
// step (1-15) and complete (20) never use.
func requireFreshnessRefusal(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected a base-freshness refusal, got success")
	}
	if !strings.HasPrefix(err.Error(), bfcSentinel) {
		t.Fatalf("the refusal must start with %s, got %q", bfcSentinel, err.Error())
	}
	code, ok := ResolveExitCode(err)
	if !ok || code != 16 {
		t.Fatalf("the refusal must exit 16, got %d (resolved=%v): %v", code, ok, err)
	}
	if (code >= 1 && code <= 15) || code == 20 {
		t.Fatalf("exit %d collides with a merge/complete code", code)
	}
}

func TestIntegrationAcquireRefusesStaleBase(t *testing.T) {
	// AC-ISB-001: one landing past the stored base, K=0, gate on. A live foreign
	// holder makes --wait enqueue — unless the gate refuses first. The refusal
	// exits 16 with the sentinel, the window bytes are untouched, and no ticket
	// is queued.
	f := newBFCFixture(t)
	bfcWriteWorkflow(t, f.root, bfcGateYAML(true, 0))
	f.writeRecord(f.base)
	f.land(1)
	sdHoldWindow(t, f.root, "sess-foreign", "lane-9", "main", factory.BranchSourceConfig, f.repo, "t9")
	before := windowBytes(t, f.root)

	_, _, err := f.acquire("--wait=2s", "--card", "t1", "--branch", "main", "--session", "sess-lane1")
	requireFreshnessRefusal(t, err)
	if after := windowBytes(t, f.root); after != before {
		t.Fatalf("a refused acquire must leave the window record byte-identical\nbefore: %s\nafter:  %s", before, after)
	}
	if lock := sdWindow(t, f.root); len(lock.Queue) != 0 {
		t.Fatalf("a refused acquire must enqueue no ticket: %+v", lock.Queue)
	}
}

func TestBaseFreshnessGateOffAdvisesOnly(t *testing.T) {
	// AC-ISB-008: the gate off (the shipped default) — the same stale fixture
	// acquires, records the window, writes one advisory line on standard error,
	// and prints no sentinel.
	f := newBFCFixture(t)
	f.writeRecord(f.base)
	f.land(1)

	stdout, stderr, err := f.acquire("--card", "t1", "--branch", "main", "--session", "sess-lane1")
	if err != nil {
		t.Fatalf("acquire with the gate off must succeed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
	lines := freshnessLines(stderr)
	if len(lines) != 1 || !strings.HasPrefix(lines[0], "base-freshness (advisory, gate off): would refuse") {
		t.Fatalf("want exactly one advisory line on standard error, got %q\nstderr: %s", lines, stderr)
	}
	if strings.Contains(stdout+stderr, bfcSentinel) {
		t.Fatalf("an advisory must not print the sentinel\nstdout: %s\nstderr: %s", stdout, stderr)
	}
	if lock := sdWindow(t, f.root); lock.SessionID != "sess-lane1" {
		t.Fatalf("the window must be recorded for the caller: %+v", lock)
	}
}

func TestBaseFreshnessCLIUncertaintyAdmitsWithCauseLine(t *testing.T) {
	// AC-ISB-009 (CLI half): the tree is stale and the gate on throughout. An
	// uncertain applicable call admits, records the window, and writes exactly
	// one standard-error line naming the cause; a not-applicable call admits and
	// writes no gate line; nothing prints the sentinel. A pure-package test
	// alone would pass a CLI that dropped the line.
	type tc struct {
		name  string
		setup func(f *bfcFixture)
		args  []string
		cwd   func(f *bfcFixture) string
		want  string // cause substring; "" = silent
	}
	card := []string{"--card", "t1", "--branch", "main", "--session", "sess-lane1"}
	cases := []tc{
		{"unreadable configuration", func(f *bfcFixture) {
			bfcWriteWorkflow(f.t, f.root, "workflow:\n  base_freshness_gate:\n    enabled: true\n  agent_tiers:\n    classes:\n      lane: bogus\n")
		}, card, nil, "configuration unreadable"},
		{"negative K", func(f *bfcFixture) { bfcWriteWorkflow(f.t, f.root, bfcGateYAML(true, -1)) }, card, nil, "negative"},
		{"integration branch unresolvable", func(f *bfcFixture) {
			// The configured target names a branch that does not exist; the gate
			// measures the configured target only, so the cause names it.
			writeGitStrategyFixture(f.t, f.root, "git-flow", "no-such-branch")
		}, []string{"--card", "t1", "--session", "sess-lane1"}, nil, "no-such-branch"},
		{"git failure injected", func(f *bfcFixture) {
			orig := baseFreshnessSeams
			baseFreshnessSeams = factory.BaseFreshnessSeams{Git: func(args ...string) (string, error) {
				if len(args) > 0 && args[0] == "rev-list" {
					return "", errors.New("injected rev-list failure")
				}
				return factorylane.ExecGitRunner{Dir: f.card}.Git(args...)
			}}
			f.t.Cleanup(func() { baseFreshnessSeams = orig })
		}, card, nil, "git rev-list could not run"},
		{"no card flag", nil, []string{"--branch", "main", "--session", "sess-lane1"}, nil, ""},
		{"cwd on the integration branch", nil, card, func(f *bfcFixture) string { return f.repo }, ""},
		{"detached HEAD", func(f *bfcFixture) { bfcGit(f.t, f.card, "checkout", "-q", "--detach") }, card, nil, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newBFCFixture(t)
			bfcWriteWorkflow(t, f.root, bfcGateYAML(true, 0))
			f.writeRecord(f.base)
			f.land(1)
			if c.setup != nil {
				c.setup(f)
			}
			dir := f.card
			if c.cwd != nil {
				dir = c.cwd(f)
			}
			stdout, stderr, err := runIntegrationSplit(t, dir, f.root, append([]string{"acquire"}, c.args...)...)
			if err != nil {
				t.Fatalf("must admit, got %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
			}
			if lock := sdWindow(t, f.root); lock.SessionID != "sess-lane1" {
				t.Fatalf("the window must be recorded for the caller: %+v", lock)
			}
			if strings.Contains(stdout+stderr, bfcSentinel) {
				t.Fatalf("no admit may print the sentinel\nstdout: %s\nstderr: %s", stdout, stderr)
			}
			lines := freshnessLines(stderr)
			if c.want == "" {
				if len(lines) != 0 {
					t.Fatalf("a not-applicable call must write no gate line, got %q", lines)
				}
				return
			}
			if len(lines) != 1 || !strings.Contains(lines[0], c.want) {
				t.Fatalf("want exactly one line naming %q, got %q\nstderr: %s", c.want, lines, stderr)
			}
		})
	}
}

func TestBaseFreshnessRefusalSentinelAndExitCode(t *testing.T) {
	// AC-ISB-011: the three refusal kinds — stale past K, past a positive K, and
	// no valid record — each start with the sentinel and exit 16.
	args := []string{"--card", "t1", "--branch", "main", "--session", "sess-lane1"}
	t.Run("landings behind exceeds K=0", func(t *testing.T) {
		f := newBFCFixture(t)
		bfcWriteWorkflow(t, f.root, bfcGateYAML(true, 0))
		f.writeRecord(f.base)
		f.land(1)
		_, _, err := f.acquire(args...)
		requireFreshnessRefusal(t, err)
	})
	t.Run("landings behind exceeds K=2", func(t *testing.T) {
		f := newBFCFixture(t)
		bfcWriteWorkflow(t, f.root, bfcGateYAML(true, 2))
		f.writeRecord(f.base)
		f.land(3)
		_, _, err := f.acquire(args...)
		requireFreshnessRefusal(t, err)
		if !strings.Contains(err.Error(), "3 landings behind") || !strings.Contains(err.Error(), "max_behind_landings=2") {
			t.Fatalf("the refusal must name the numbers: %v", err)
		}
	})
	t.Run("no valid record", func(t *testing.T) {
		f := newBFCFixture(t)
		bfcWriteWorkflow(t, f.root, bfcGateYAML(true, 2))
		_, _, err := f.acquire(args...)
		requireFreshnessRefusal(t, err)
	})
	t.Run("equal to K admits", func(t *testing.T) {
		f := newBFCFixture(t)
		bfcWriteWorkflow(t, f.root, bfcGateYAML(true, 2))
		f.writeRecord(f.base)
		f.land(2)
		if _, _, err := f.acquire(args...); err != nil {
			t.Fatalf("landings behind == K must admit: %v", err)
		}
	})
}

// ---- factory complete ------------------------------------------------------

func TestFactoryCompleteRefusesStaleBaseBeforeAcquire(t *testing.T) {
	// AC-ISB-002: complete is about to acquire the window itself, the base moved
	// one landing, the tip is not an adoption, the gate is on. The refusal
	// precedes the acquisition: the window record, the integration branch and the
	// card state are unchanged.
	root, integ, _ := mwq19Fixture(t)
	bfcWriteWorkflow(t, root, bfcGateYAML(true, 0))
	fcGit(t, integ, "commit", "-q", "--allow-empty", "-m", "another lane landed")
	tipBefore := fcGit(t, integ, "rev-parse", "develop")
	cardBefore := fcCard(t, root, "t1")
	windowBefore := windowBytes(t, root)

	_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	requireFreshnessRefusal(t, err)
	if got := windowBytes(t, root); got != windowBefore {
		t.Fatalf("the window record must be unchanged\nbefore: %s\nafter:  %s", windowBefore, got)
	}
	if got := fcGit(t, integ, "rev-parse", "develop"); got != tipBefore {
		t.Fatalf("the integration branch must be unchanged: %s -> %s", tipBefore, got)
	}
	sdCardUnchanged(t, "refused complete", root, "t1", cardBefore)
}

func TestBaseFreshnessKDoesNotRelaxMergeStep(t *testing.T) {
	// AC-ISB-010 Case B at the CLI: K=3 admits a candidate two landings behind,
	// and the merge step — the authority — still refuses it with its own code 2
	// and releases the window. K buys entry slack, never a relaxed merge.
	root, integ, _ := mwq19Fixture(t)
	bfcWriteWorkflow(t, root, bfcGateYAML(true, 3))
	fcGit(t, integ, "commit", "-q", "--allow-empty", "-m", "landing one")
	fcGit(t, integ, "commit", "-q", "--allow-empty", "-m", "landing two")

	_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	if err == nil {
		t.Fatalf("the merge step must still refuse a moved base")
	}
	if strings.Contains(err.Error(), bfcSentinel) {
		t.Fatalf("K=3 must admit 2 landings behind — the refusal is the merge step's: %v", err)
	}
	code, ok := ResolveExitCode(err)
	if !ok || code != factory.MergeExitBaseMoved {
		t.Fatalf("the merge step's own code %d must stand, got %d (resolved=%v): %v", factory.MergeExitBaseMoved, code, ok, err)
	}
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("the refused merge must release the window: %+v", lock)
	}
}

func TestFactoryCompleteAdoptionNotRefusedByBaseFreshness(t *testing.T) {
	// AC-ISB-016: a card the lane merged through the merge verb earlier is
	// adopted by complete without the gate, whether or not the lane holds the
	// window. Without the exemption the merge commit is a first-parent landing
	// past the candidate record's base and K=0 refuses a legitimate adoption.
	adopted := func(t *testing.T) (root, integ string, cardWT sdCardTree, merge string) {
		root, integ, cardWT = mwq19Fixture(t)
		bfcWriteWorkflow(t, root, bfcGateYAML(true, 0))
		// The merge verb's end state: a --no-ff merge commit whose second parent
		// is the card tip. The card had absorbed develop, so the merged tree IS
		// the candidate's tree and the record keyed to it — still naming the
		// pre-merge tip as its base — is the adoption's evidence.
		fcGit(t, integ, "merge", "-q", "--no-ff", "-m", "M t1", cardWT.branch)
		merge = fcGit(t, integ, "rev-parse", "HEAD")
		rec, err := factory.ReadRemeasureRecord(root, fcGit(t, integ, "rev-parse", "HEAD^{tree}"))
		if err != nil || factory.ValidateRemeasureRecord(rec) != nil {
			t.Fatalf("fixture sanity: a valid record must key the merged tree: %v", err)
		}
		if rec.Base == merge {
			t.Fatalf("fixture sanity: the record's base must be the pre-merge tip, not the merge commit — otherwise the gate admits with no exemption at all")
		}
		return root, integ, cardWT, merge
	}
	t.Run("adoption completes and prints no sentinel", func(t *testing.T) {
		root, _, _, merge := adopted(t)
		stdout, stderr, err := runFactory(t, "complete", "t1", "--run", fcRun)
		if err != nil {
			t.Fatalf("an adoption must complete exactly as without the gate: %v", err)
		}
		if strings.Contains(stdout+stderr, bfcSentinel) {
			t.Fatalf("no sentinel on an adoption\nstdout: %s\nstderr: %s", stdout, stderr)
		}
		card := fcCard(t, root, "t1")
		if card.State != homestate.CardMergedLocal || card.MergeSHA != merge {
			t.Fatalf("the card must land merged-local on the pre-made merge %s, got %s %s", merge, card.State, card.MergeSHA)
		}
	})
	t.Run("a commit after the merge is not an adoption; the gate evaluates", func(t *testing.T) {
		root, _, cardWT, _ := adopted(t)
		if err := os.WriteFile(filepath.Join(cardWT.wt, "later.txt"), []byte("later\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		fcGit(t, cardWT.wt, "add", "later.txt")
		fcGit(t, cardWT.wt, "commit", "-q", "-m", "work after the merge")
		before := fcCard(t, root, "t1")
		_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
		requireFreshnessRefusal(t, err)
		sdCardUnchanged(t, "non-adoption refusal", root, "t1", before)
	})
}

func TestFactoryCompleteAdoptionRequiresRecordTreeIdentity(t *testing.T) {
	// iter12 F1: the adoption exemption asked only ReadRemeasureRecord +
	// ValidateRemeasureRecord, so a record file KEYED to the tip tree whose
	// `tree` field names another valid SHA exempted the gate and complete
	// finished merged-local without the gate ever judging the candidate. The
	// exemption must require the record to NAME the tree it is stored for —
	// the same identity comparison the gate's own verdict makes.
	setup := func(t *testing.T) (root, integ string, cardWT sdCardTree, merge, tipTree string) {
		t.Helper()
		root, integ, cardWT = mwq19Fixture(t)
		bfcWriteWorkflow(t, root, bfcGateYAML(true, 0))
		fcGit(t, integ, "merge", "-q", "--no-ff", "-m", "M t1", cardWT.branch)
		merge = fcGit(t, integ, "rev-parse", "HEAD")
		tipTree = fcGit(t, integ, "rev-parse", "HEAD^{tree}")
		return root, integ, cardWT, merge, tipTree
	}
	forgeTreeField := func(t *testing.T, root, keyTree, treeField string) {
		t.Helper()
		rec, err := factory.ReadRemeasureRecord(root, keyTree)
		if err != nil || factory.ValidateRemeasureRecord(rec) != nil {
			t.Fatalf("fixture sanity: a valid record must key the tree before the forge: %v", err)
		}
		rec.Tree = treeField // the forge: stored under keyTree, naming treeField
		data, err := json.Marshal(rec)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, ".moai", "state", "remeasure", keyTree+".json"), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Run("a record stored for the tip tree but naming another tree is not adopted", func(t *testing.T) {
		root, integ, _, merge, tipTree := setup(t)
		other := fcGit(t, integ, "rev-parse", merge+"^1^{tree}")
		if other == tipTree {
			t.Fatalf("fixture sanity: the first parent's tree must differ from the merged tree, or the forge is no forge")
		}
		forgeTreeField(t, root, tipTree, other)
		before := fcCard(t, root, "t1")
		_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
		requireFreshnessRefusal(t, err)
		sdCardUnchanged(t, "forged-record refusal", root, "t1", before)
	})
	t.Run("a healthy record naming its key tree is still adopted", func(t *testing.T) {
		root, _, _, merge, tipTree := setup(t)
		rec, err := factory.ReadRemeasureRecord(root, tipTree)
		if err != nil || factory.ValidateRemeasureRecord(rec) != nil || rec.Tree != tipTree {
			t.Fatalf("fixture sanity: the merged tree must key a valid record naming it: %v", err)
		}
		stdout, stderr, err := runFactory(t, "complete", "t1", "--run", fcRun)
		if err != nil {
			t.Fatalf("a healthy adoption must complete: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}
		if strings.Contains(stdout+stderr, bfcSentinel) {
			t.Fatalf("no sentinel on a healthy adoption\nstdout: %s\nstderr: %s", stdout, stderr)
		}
		card := fcCard(t, root, "t1")
		if card.State != homestate.CardMergedLocal || card.MergeSHA != merge {
			t.Fatalf("the card must land merged-local on the pre-made merge %s, got %s %s", merge, card.State, card.MergeSHA)
		}
	})
}

// ---- audit-fix round: the refusal must reach the user ----------------------

// runIntegrationThroughRoot runs `moai integration …` through the real root
// execution seam (runFang with moaiErrorHandler), the path a process takes. A
// refusal that only rides the returned error is invisible there: the handler is
// silent for an exit-coded error, so the message has to be written by the verb.
func runIntegrationThroughRoot(t *testing.T, worktree, root string, args ...string) (stdout, stderr string, exit int) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(worktree); err != nil {
		t.Fatalf("chdir %s: %v", worktree, err)
	}
	defer func() { _ = os.Chdir(prev) }()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	rootCmd := &cobra.Command{Use: "moai"}
	rootCmd.AddCommand(newIntegrationCmd())
	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs(append([]string{"integration"}, args...))
	runErr := runFang(context.Background(), rootCmd)
	return out.String(), errBuf.String(), exitCodeForError(runErr)
}

func TestBaseFreshnessRefusalReachesTheUserThroughTheRootErrorPath(t *testing.T) {
	// AC-ISB-001 / AC-ISB-011 (audit-fix round): through the root error handler
	// the stale-base refusal exits 16 AND the sentinel line with its recovery
	// guidance is on standard error, exactly once. The gate-off advisory and the
	// uncertainty notice already printed; the refusal alone was silent.
	f := newBFCFixture(t)
	bfcWriteWorkflow(t, f.root, bfcGateYAML(true, 0))
	f.writeRecord(f.base)
	f.land(1)

	stdout, stderr, exit := runIntegrationThroughRoot(t, f.card, f.root,
		"acquire", "--card", "t1", "--branch", "main", "--session", "sess-lane1")
	if exit != factory.BaseFreshnessExitRefused {
		t.Fatalf("exit = %d, want %d\nstdout: %q\nstderr: %q", exit, factory.BaseFreshnessExitRefused, stdout, stderr)
	}
	var sentinelLines []string
	for _, line := range strings.Split(stderr, "\n") {
		if strings.HasPrefix(line, bfcSentinel) {
			sentinelLines = append(sentinelLines, line)
		}
	}
	if len(sentinelLines) != 1 {
		t.Fatalf("want exactly one sentinel line on stderr, got %d\nstdout: %q\nstderr: %q", len(sentinelLines), stdout, stderr)
	}
	for _, want := range []string{"1 landings behind", "max_behind_landings=0", "moai integration remeasure"} {
		if !strings.Contains(sentinelLines[0], want) {
			t.Fatalf("the visible refusal must carry %q: %s", want, sentinelLines[0])
		}
	}
}

func TestFactoryCompleteRefusalIsVisible(t *testing.T) {
	// AC-ISB-002 / AC-ISB-011 (audit-fix round): `moai factory complete` shares
	// the refusal body; the refusal line must be in the command's own output, not
	// only in the returned error that the root handler renders as nothing.
	root, integ, _ := mwq19Fixture(t)
	bfcWriteWorkflow(t, root, bfcGateYAML(true, 0))
	fcGit(t, integ, "commit", "-q", "--allow-empty", "-m", "another lane landed")

	// Through the root seam: a bare cobra run prints "Error: <msg>" itself and
	// would hide the very silence the audit found.
	rootCmd := &cobra.Command{Use: "moai"}
	rootCmd.AddCommand(newFactoryCommand())
	var outBuf, errBuf bytes.Buffer
	rootCmd.SetOut(&outBuf)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"factory", "complete", "t1", "--run", fcRun})
	runErr := runFang(context.Background(), rootCmd)
	stdout, stderr := outBuf.String(), errBuf.String()
	if got := exitCodeForError(runErr); got != factory.BaseFreshnessExitRefused {
		t.Fatalf("exit = %d, want %d\nstdout: %q\nstderr: %q", got, factory.BaseFreshnessExitRefused, stdout, stderr)
	}
	var n int
	for _, line := range strings.Split(stdout+"\n"+stderr, "\n") {
		if strings.HasPrefix(line, bfcSentinel) {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("want exactly one sentinel line in the command output, got %d\nstdout: %q\nstderr: %q", n, stdout, stderr)
	}
}

// ---- audit-fix round: a config load failure is uncertainty -----------------

func TestBaseFreshnessConfigLoadFailureIsUncertainty(t *testing.T) {
	// AC-ISB-009 (audit-fix round): the shared config loader turns a read or
	// parse failure of workflow.yaml into a warning plus defaults, so the gate
	// would read "gate off" and print the advisory — a verdict taken on defaults
	// the operator did not choose. The gate must route the failure through the
	// uncertainty path with the real cause: admit, exactly one gate line, naming
	// the configuration, and no "advisory, gate off" claim.
	cases := []struct {
		name  string
		setup func(t *testing.T, f *bfcFixture)
		cause string
	}{
		{"malformed YAML", func(t *testing.T, f *bfcFixture) {
			bfcWriteWorkflow(t, f.root, "workflow:\n  base_freshness_gate: [unterminated\n    enabled: true\n")
		}, "parse"},
		{"wrongly typed key", func(t *testing.T, f *bfcFixture) {
			bfcWriteWorkflow(t, f.root, "workflow:\n  base_freshness_gate:\n    enabled: \"not-a-bool\"\n    max_behind_landings: [1]\n")
		}, "parse"},
		{"unreadable file", func(t *testing.T, f *bfcFixture) {
			// A directory where the file belongs: ReadFile fails with a
			// non-NotExist error on every platform, without depending on the
			// test process's privileges.
			dir := filepath.Join(f.root, ".moai", "config", "sections", "workflow.yaml")
			if err := os.MkdirAll(dir, 0o750); err != nil {
				t.Fatal(err)
			}
		}, "read"},
	}
	args := []string{"--card", "t1", "--branch", "main", "--session", "sess-lane1"}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newBFCFixture(t)
			f.writeRecord(f.base)
			f.land(1)
			c.setup(t, f)
			stdout, stderr, err := f.acquire(args...)
			if err != nil {
				t.Fatalf("an unreadable configuration must admit, got %v\nstderr: %s", err, stderr)
			}
			if lock := sdWindow(t, f.root); lock.SessionID != "sess-lane1" {
				t.Fatalf("the window must be recorded for the caller: %+v", lock)
			}
			if strings.Contains(stdout+stderr, "advisory, gate off") || strings.Contains(stdout+stderr, bfcSentinel) {
				t.Fatalf("a config failure must not read as gate-off or as a refusal\nstderr: %s", stderr)
			}
			lines := freshnessLines(stderr)
			if len(lines) != 1 {
				t.Fatalf("want exactly one gate line, got %q\nstderr: %s", lines, stderr)
			}
			if !strings.Contains(lines[0], "configuration unreadable") || !strings.Contains(lines[0], c.cause) ||
				!strings.Contains(lines[0], "workflow.yaml") {
				t.Fatalf("the line must name the configuration, the file and the cause %q: %s", c.cause, lines[0])
			}
		})
	}
}

func TestBaseFreshnessConfigFailureNoticeDoesNotEchoConfigContent(t *testing.T) {
	// Delta-audit finding: the cause line must be fixed and input-free. A parser
	// or validator error can quote the offending value, and a notice is printed
	// to a shared terminal and captured in logs.
	const secret = "sk_SYNTHETIC_PRIVATE_VALUE_72b9"
	cases := []struct {
		name string
		body string // "" = workflow.yaml is a directory (read failure)
	}{
		{"malformed YAML", "workflow:\n  base_freshness_gate:\n    enabled: *" + secret + "\n"},
		{"wrong-type value", "workflow:\n  base_freshness_gate:\n    enabled: " + secret + "\n    max_behind_landings: " + secret + "\n"},
		{"rejected tier token", "workflow:\n  agent_tiers:\n    classes:\n      lane: " + secret + "\n"},
		{"read failure", ""},
	}
	args := []string{"--card", "t1", "--branch", "main", "--session", "sess-lane1"}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newBFCFixture(t)
			f.writeRecord(f.base)
			f.land(1)
			if c.body == "" {
				if err := os.MkdirAll(filepath.Join(f.root, ".moai", "config", "sections", "workflow.yaml"), 0o750); err != nil {
					t.Fatal(err)
				}
			} else {
				bfcWriteWorkflow(t, f.root, c.body)
			}
			stdout, stderr, err := f.acquire(args...)
			if err != nil {
				t.Fatalf("must admit: %v", err)
			}
			all := stdout + stderr + fmt.Sprint(err)
			for _, leak := range []string{secret, "72b9", "SYNTHETIC"} {
				if strings.Contains(all, leak) {
					t.Fatalf("the output echoes config content %q:\n%s", leak, all)
				}
			}
			if lines := freshnessLines(stderr); len(lines) != 1 || !strings.Contains(lines[0], "configuration unreadable") {
				t.Fatalf("want one configuration-unreadable line, got %q", lines)
			}
		})
	}
}

func TestBaseFreshnessOutputNeverEchoesUntrustedInput(t *testing.T) {
	// Class test (third audit): every untrusted input source — record fields,
	// workflow.yaml content, a rejected tier token, git's own error text — gets a
	// synthetic secret, and the whole gate output (stdout, stderr, the returned
	// error text) must hold no fragment of it, gate off and gate on. Causes come
	// from a fixed phrase set; paths we construct and validated SHAs are fine.
	const secret = "sk_SYNTHETIC_PRIVATE_VALUE_72b9"
	fragments := []string{secret, "SYNTHETIC", "72b9"}
	writeRecord := func(t *testing.T, f *bfcFixture, mutate func(*factory.RemeasureRecord)) {
		tree := f.tree()
		rec := factory.RemeasureRecord{Tree: tree, Base: f.base, Command: "true", BuildIdentity: "moai test",
			RecordedAt: time.Date(2026, 10, 18, 9, 0, 0, 0, time.UTC).Format(time.RFC3339)}
		mutate(&rec)
		if err := factory.WriteRemeasureRecord(f.root, tree, rec); err != nil {
			t.Fatal(err)
		}
	}
	sources := []struct {
		name  string
		setup func(t *testing.T, f *bfcFixture)
	}{
		{"record command with a structured-output demand", func(t *testing.T, f *bfcFixture) {
			writeRecord(t, f, func(r *factory.RemeasureRecord) {
				r.Command = "API_KEY=" + secret + " go test ./..."
				r.StructuredRequired, r.HasStructured = true, false
			})
		}},
		{"record string fields with a failing exit", func(t *testing.T, f *bfcFixture) {
			writeRecord(t, f, func(r *factory.RemeasureRecord) {
				r.Command, r.BuildIdentity, r.ExitCode = secret, secret, 3
				r.Failing = []string{secret}
			})
		}},
		{"record with a secret base and tree text", func(t *testing.T, f *bfcFixture) {
			writeRecord(t, f, func(r *factory.RemeasureRecord) { r.Base = secret })
		}},
		{"malformed workflow.yaml", func(t *testing.T, f *bfcFixture) {
			bfcWriteWorkflow(t, f.root, "workflow:\n  base_freshness_gate:\n    enabled: *"+secret+"\n")
		}},
		{"wrong-type workflow.yaml value", func(t *testing.T, f *bfcFixture) {
			bfcWriteWorkflow(t, f.root, "workflow:\n  base_freshness_gate:\n    enabled: "+secret+"\n")
		}},
		{"rejected tier token", func(t *testing.T, f *bfcFixture) {
			bfcWriteWorkflow(t, f.root, "workflow:\n  agent_tiers:\n    classes:\n      lane: "+secret+"\n")
		}},
		{"git failure quoting a config value", func(t *testing.T, f *bfcFixture) {
			cfg := filepath.Join(f.repo, ".git", "config")
			old, err := os.ReadFile(cfg)
			if err != nil {
				t.Fatal(err)
			}
			add := "[core]\n\trepositoryformatversion = 1\n[extensions]\n\tobjectFormat = " + secret + "\n"
			if err := os.WriteFile(cfg, append(old, []byte(add)...), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
		{"unreadable workflow.yaml path", func(t *testing.T, f *bfcFixture) {
			path := filepath.Join(f.root, ".moai", "config", "sections", "workflow.yaml")
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(path, 0o750); err != nil {
				t.Fatal(err)
			}
		}},
	}
	args := []string{"--card", "t1", "--branch", "main", "--session", "sess-lane1"}
	for _, src := range sources {
		for _, enabled := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/enabled=%t", src.name, enabled), func(t *testing.T) {
				f := newBFCFixture(t)
				bfcWriteWorkflow(t, f.root, bfcGateYAML(enabled, 0))
				f.writeRecord(f.base)
				f.land(1)
				src.setup(t, f)
				stdout, stderr, err := f.acquire(args...)
				// The gate's own lines only: the settings-drift assertion that
				// acquire also prints is a neighbouring feature outside this card.
				var all string
				for _, line := range strings.Split(stdout+"\n"+stderr, "\n") {
					if strings.HasPrefix(line, "base-freshness") || strings.HasPrefix(line, bfcSentinel) {
						all += line + "\n"
					}
				}
				if err != nil {
					all += "\n" + err.Error()
				}
				for _, frag := range fragments {
					if strings.Contains(all, frag) {
						t.Fatalf("the gate output echoes untrusted input %q:\n%s", frag, all)
					}
				}
			})
		}
	}
}

// ---- audit iter8 F2: remeasure and acquire read the same configuration root ----

func TestRemeasureAndAcquireAgreeOnIntegrationTargetAcrossLinkedRoot(t *testing.T) {
	// Audit iter8 F2: since card t1680 acquire, complete and the gate resolve the
	// project through integrationLockRoot (the repository common root), while the
	// remeasure verb named its absorbed base through resolveProjectDir
	// (CLAUDE_PROJECT_DIR as given). When CLAUDE_PROJECT_DIR names a linked
	// worktree whose git-strategy.yaml differs from the common root's, the two
	// verbs measured against different integration branches: a candidate that had
	// absorbed the common root's target and remeasured successfully was refused.
	//
	// The common root (the primary checkout, `main` checked out) names `main` as
	// its develop branch; the linked card worktree names `develop`, a branch that
	// stayed at the pre-landing base. The card branched from `main` after the
	// landing, so against the common root's target it is fresh (0 behind).
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repo, card := filepath.Join(root, "repo"), filepath.Join(root, "card")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	bfcGit(t, repo, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(repo, "base.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bfcGit(t, repo, "add", "base.txt")
	bfcGit(t, repo, "commit", "-q", "-m", "base")
	bfcGit(t, repo, "branch", "develop")
	bfcGit(t, repo, "commit", "-q", "--allow-empty", "-m", "landing")
	mainTip := bfcGit(t, repo, "rev-parse", "main")
	bfcGit(t, repo, "worktree", "add", "-q", "-b", "WT-card", card, "main")
	if err := os.WriteFile(filepath.Join(card, "card.txt"), []byte("card work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bfcGit(t, card, "add", "card.txt")
	bfcGit(t, card, "commit", "-q", "-m", "card work")

	// Keep the per-tree configuration out of git status: the remeasure verb
	// refuses a dirty worktree. info/exclude lives in the common git dir, so it
	// covers the primary checkout and the linked worktree alike.
	exclude := filepath.Join(repo, ".git", "info", "exclude")
	if err := os.MkdirAll(filepath.Dir(exclude), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exclude, []byte(".moai/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeGitStrategyFixture(t, repo, "git-flow", "main")
	writeGitStrategyFixture(t, card, "git-flow", "develop")
	bfcWriteWorkflow(t, repo, bfcGateYAML(true, 0))

	// CLAUDE_PROJECT_DIR names the linked worktree for both verbs
	// (runIntegrationSplit sets it from its root argument).
	if _, stderr, err := runIntegrationSplit(t, card, card, "remeasure", "--", "true"); err != nil {
		t.Fatalf("remeasure must succeed: %v\nstderr: %s", err, stderr)
	}
	common := factory.IntegrationCommonRoot(card)
	rec, err := factory.ReadRemeasureRecord(common, bfcGit(t, card, "rev-parse", "HEAD^{tree}"))
	if err != nil {
		t.Fatalf("remeasure must store its record at the common root %s: %v", common, err)
	}
	if rec.Base != mainTip {
		t.Errorf("remeasure must name the common root's integration target (main = %s) as its base, got %s", mainTip, rec.Base)
	}

	stdout, stderr, err := runIntegrationSplit(t, card, card, "acquire", "--card", "t1", "--session", "sess-lane1")
	if err != nil {
		t.Fatalf("acquire after a successful remeasure of an absorbed candidate must admit, got: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
	if lock := sdWindow(t, common); lock.SessionID != "sess-lane1" {
		t.Fatalf("the window must be recorded for the caller at the common root: %+v", lock)
	}
}

// ---- audit iter10 H1/H2: the gate measures the configured integration target ----

func TestAcquireWithoutBranchMeasuresTheConfiguredTargetInGitHubFlow(t *testing.T) {
	// Audit iter10 H1: in a github-flow project develop_branch is empty by design,
	// so without --branch acquire resolved its target to the caller's own branch
	// and the gate then read the call as made from the integration branch and
	// returned without measuring. A candidate one landing behind `main`, K=0, gate
	// on, acquire --card without --branch must be refused.
	f := newBFCFixture(t)
	bfcWriteWorkflow(t, f.root, bfcGateYAML(true, 0))
	f.writeRecord(f.base)
	f.land(1)

	_, _, err := f.acquire("--card", "t1", "--session", "sess-lane1")
	requireFreshnessRefusal(t, err)
	if !strings.Contains(err.Error(), "local main") {
		t.Fatalf("the refusal must name the configured target main: %v", err)
	}
	if got := windowBytes(t, f.root); got != "<absent>" {
		t.Fatalf("a refused acquire must record no window, got %s", got)
	}
}

func TestAcquireRefusesAnExplicitBranchOtherThanTheConfiguredTarget(t *testing.T) {
	// Audit iter10 H1 and the leader's ruling: the gate measures the configured
	// integration target only. An explicit --branch that differs is refused with a
	// reason naming both branches, never admitted and never measured against the
	// other branch. The candidate is FRESH against both (alternate sits on main's
	// tip), so only the mismatch can refuse it.
	f := newBFCFixture(t)
	bfcWriteWorkflow(t, f.root, bfcGateYAML(true, 0))
	f.writeRecord(f.base)
	bfcGit(t, f.repo, "branch", "alternate", "main")

	_, _, err := f.acquire("--card", "t1", "--branch", "alternate", "--session", "sess-lane1")
	requireFreshnessRefusal(t, err)
	for _, want := range []string{"alternate", "main", "configured integration target"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("the refusal must carry %q: %v", want, err)
		}
	}
	if got := windowBytes(t, f.root); got != "<absent>" {
		t.Fatalf("a refused acquire must record no window, got %s", got)
	}
}

func TestAcquireRefusalRemedyInGitHubFlowNamesTheTargetAndFollowingItRecordsIt(t *testing.T) {
	// Audit iter11 F3: in github-flow develop_branch is empty, so `acquire` without
	// --branch records the caller's own branch. The remedy must therefore name the
	// configured target, and following it must record that target as the window.
	f := newBFCFixture(t)
	bfcWriteWorkflow(t, f.root, bfcGateYAML(true, 0))
	f.writeRecord(f.base)
	bfcGit(t, f.repo, "branch", "alternate", "main")

	_, _, err := f.acquire("--card", "t1", "--branch", "alternate", "--session", "sess-lane1")
	requireFreshnessRefusal(t, err)
	if !strings.Contains(err.Error(), "Acquire with --branch main") || strings.Contains(err.Error(), "without --branch") {
		t.Fatalf("the github-flow remedy must name --branch main and not advise omitting it: %v", err)
	}
	if got := windowBytes(t, f.root); got != "<absent>" {
		t.Fatalf("a refused acquire must record no window, got %s", got)
	}

	stdout, stderr, err := f.acquire("--card", "t1", "--branch", "main", "--session", "sess-lane1")
	if err != nil {
		t.Fatalf("following the remedy must admit: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
	if lock := sdWindow(t, f.root); lock.SessionID != "sess-lane1" || lock.Branch != "main" {
		t.Fatalf("following the remedy must record the configured target as the window branch: %+v", lock)
	}
}

func TestAcquireRefusalRemedyInGitFlowKeepsOmitBranch(t *testing.T) {
	// Where a develop branch is configured, omitting --branch takes the configured
	// target, so the remedy keeps advising it.
	f := newBFCFixture(t)
	writeGitStrategyFixture(t, f.root, "git-flow", "main")
	bfcWriteWorkflow(t, f.root, bfcGateYAML(true, 0))
	f.writeRecord(f.base)
	bfcGit(t, f.repo, "branch", "alternate", "main")

	_, _, err := f.acquire("--card", "t1", "--branch", "alternate", "--session", "sess-lane1")
	requireFreshnessRefusal(t, err)
	if !strings.Contains(err.Error(), "Acquire without --branch (the configured target main)") {
		t.Fatalf("the git-flow remedy must keep advising to omit --branch: %v", err)
	}
}

func TestAcquireExplicitBranchOtherThanTheConfiguredTargetIsAdvisoryWhenGateOff(t *testing.T) {
	// The same mismatch with the gate off (the shipped default): acquire still
	// records the window for the explicit branch, prints no sentinel, and the one
	// advisory line names both branches.
	f := newBFCFixture(t)
	f.writeRecord(f.base)
	bfcGit(t, f.repo, "branch", "alternate", "main")

	stdout, stderr, err := f.acquire("--card", "t1", "--branch", "alternate", "--session", "sess-lane1")
	if err != nil {
		t.Fatalf("acquire with the gate off must succeed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
	if strings.Contains(stdout+stderr, bfcSentinel) {
		t.Fatalf("an advisory must not print the sentinel\nstdout: %s\nstderr: %s", stdout, stderr)
	}
	lines := freshnessLines(stderr)
	if len(lines) != 1 || !strings.HasPrefix(lines[0], "base-freshness (advisory, gate off): would refuse") ||
		!strings.Contains(lines[0], "alternate") || !strings.Contains(lines[0], "main") {
		t.Fatalf("want exactly one advisory line naming both branches, got %q\nstderr: %s", lines, stderr)
	}
	if lock := sdWindow(t, f.root); lock.SessionID != "sess-lane1" || lock.Branch != "alternate" {
		t.Fatalf("the window must be recorded for the caller on the explicit branch: %+v", lock)
	}
}

func TestRemeasureThenAcquireAdmitsWhenATagSharesTheIntegrationBranchName(t *testing.T) {
	// Audit iter10 H2: remeasure named its absorbed base with a bare-name
	// merge-base, which git resolves to a tag before a branch, while the gate reads
	// refs/heads/<branch>. With an older tag named like the integration branch the
	// record's base followed the tag and the refusal repeated after the prescribed
	// remedy. The record's base must be the BRANCH tip.
	f := newBFCFixture(t)
	bfcWriteWorkflow(t, f.root, bfcGateYAML(true, 0))
	bfcGit(t, f.repo, "tag", "main", f.base)
	f.land(1)
	bfcGit(t, f.card, "merge", "-q", "--no-ff", "-m", "absorb main", "refs/heads/main")
	tip := bfcGit(t, f.repo, "rev-parse", "refs/heads/main")

	if _, stderr, err := runIntegrationSplit(t, f.card, f.root, "remeasure", "--", "true"); err != nil {
		t.Fatalf("remeasure must succeed: %v\nstderr: %s", err, stderr)
	}
	rec, err := factory.ReadRemeasureRecord(f.root, f.tree())
	if err != nil {
		t.Fatalf("remeasure must store its record: %v", err)
	}
	if rec.Base != tip {
		t.Fatalf("the record's base must be the branch tip %s, got %s (the older tag is %s)", tip, rec.Base, f.base)
	}
	if stdout, stderr, err := f.acquire("--card", "t1", "--session", "sess-lane1"); err != nil {
		t.Fatalf("acquire after a remeasure of an absorbed candidate must admit: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
}
