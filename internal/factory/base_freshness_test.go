package factory

// base_freshness_test.go — tests for the pre-integration base-freshness gate
// (SPEC-INTEGRATION-STALE-BASE-001, card t1834). The gate asks one question
// before a candidate may take (or queue for) the integration window: is the
// candidate's re-measure record keyed to its CURRENT tree, and is the base that
// record names at most K landings behind the local integration branch? These
// tests drive the pure function against fixture repositories and a git-runner
// seam; the CLI call sites are pinned in internal/cli.
//
// Fixture shape: one repository whose `main` is the integration branch, one
// worktree on a card branch that carries a commit of its own, and a project
// root holding the re-measure store. Landings are added to `main` AFTER the
// card branched, so the card's tree never changes while its distance grows —
// the content-free-absorb shape the stored-base measurement exists for.

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factorylane"
	"github.com/modu-ai/moai-adk/internal/gitenv"
)

const bfKey = "workflow.base_freshness_gate.max_behind_landings"

// bfGit runs git in dir with the repo-scoping variables removed and a fixed
// identity, so the fixture never reads the developer's own configuration.
func bfGit(t *testing.T, dir string, args ...string) string {
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

type bfFixture struct {
	t     *testing.T
	root  string // project root: the re-measure store and window state live under root/.moai
	repo  string // the main checkout, branch `main` — the integration branch
	card  string // the card worktree, branch WT-card
	base  string // the main commit the card branched from
	sides int
}

func newBFFixture(t *testing.T) *bfFixture {
	t.Helper()
	root := t.TempDir()
	f := &bfFixture{t: t, root: root, repo: filepath.Join(root, "repo"), card: filepath.Join(root, "card")}
	if err := os.MkdirAll(f.repo, 0o755); err != nil {
		t.Fatal(err)
	}
	bfGit(t, f.repo, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(f.repo, "base.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bfGit(t, f.repo, "add", "base.txt")
	bfGit(t, f.repo, "commit", "-q", "-m", "base")
	f.base = bfGit(t, f.repo, "rev-parse", "HEAD")
	bfGit(t, f.repo, "worktree", "add", "-q", "-b", "WT-card", f.card, "main")
	if err := os.WriteFile(filepath.Join(f.card, "card.txt"), []byte("card work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bfGit(t, f.card, "add", "card.txt")
	bfGit(t, f.card, "commit", "-q", "-m", "card work")
	return f
}

func (f *bfFixture) tip() string  { return bfGit(f.t, f.repo, "rev-parse", "main") }
func (f *bfFixture) tree() string { return bfGit(f.t, f.card, "rev-parse", "HEAD^{tree}") }

// writeRecord stores a valid re-measure record keyed to the card's CURRENT
// tree, naming base as the absorbed base; mutate may break it.
func (f *bfFixture) writeRecord(base string, mutate func(*RemeasureRecord)) {
	f.t.Helper()
	rec := RemeasureRecord{
		Tree: f.tree(), Base: base, Command: "true", ExitCode: 0,
		BuildIdentity: "moai test",
		RecordedAt:    time.Date(2026, 10, 18, 9, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}
	if mutate != nil {
		mutate(&rec)
	}
	if err := WriteRemeasureRecord(f.root, rec.Tree, rec); err != nil {
		f.t.Fatal(err)
	}
}

// landEmpty adds n direct commits to main: n first-parent landings.
func (f *bfFixture) landEmpty(n int) {
	f.t.Helper()
	for i := 0; i < n; i++ {
		bfGit(f.t, f.repo, "commit", "-q", "--allow-empty", "-m", fmt.Sprintf("landing %d", i))
	}
}

// landMerges adds n integration merges to main, each merging a side branch of
// `commits` commits: n first-parent landings but n*(commits+1) commits in
// total — the shape that separates a first-parent count from a plain one.
func (f *bfFixture) landMerges(n, commits int) {
	f.t.Helper()
	for i := 0; i < n; i++ {
		f.sides++
		side := fmt.Sprintf("side-%d", f.sides)
		bfGit(f.t, f.repo, "checkout", "-q", "-b", side, "main")
		for c := 0; c < commits; c++ {
			bfGit(f.t, f.repo, "commit", "-q", "--allow-empty", "-m", fmt.Sprintf("%s commit %d", side, c))
		}
		bfGit(f.t, f.repo, "checkout", "-q", "main")
		bfGit(f.t, f.repo, "merge", "-q", "--no-ff", "-m", "Merge "+side, side)
	}
}

// orphanCommit makes a commit that exists in the object database but is not an
// ancestor of main.
func (f *bfFixture) orphanCommit() string {
	f.t.Helper()
	emptyTree := bfGit(f.t, f.repo, "hash-object", "-t", "tree", "--stdin")
	return bfGit(f.t, f.repo, "commit-tree", "-m", "orphan", emptyTree)
}

func (f *bfFixture) input(k int, enabled bool) BaseFreshnessInput {
	return BaseFreshnessInput{
		Root: f.root, Dir: f.card, CardID: "t1", IntegrationBranch: "main",
		K: k, Enabled: enabled,
	}
}

func requireOneLine(t *testing.T, label, line string) {
	t.Helper()
	if line == "" || strings.ContainsAny(line, "\r\n") {
		t.Fatalf("%s: want exactly one non-empty line, got %q", label, line)
	}
}

func TestLandingsBehindCountsFirstParent(t *testing.T) {
	// AC-ISB-003: three integration merges of four-commit side branches are
	// three landings, not fifteen; a record base that is not an ancestor of the
	// tip is stale at every K.
	f := newBFFixture(t)
	f.writeRecord(f.base, nil)
	f.landMerges(3, 4)

	total := bfGit(t, f.repo, "rev-list", "--count", f.base+"..main")
	if total != "15" {
		t.Fatalf("fixture sanity: a plain count of the range must be 15, got %s", total)
	}
	res := EvaluateBaseFreshness(f.input(100, true), BaseFreshnessSeams{})
	if !res.Applicable || !res.Admit {
		t.Fatalf("K=100 must admit a candidate 3 landings behind: %+v", res)
	}
	if res.Behind != 3 {
		t.Fatalf("landings behind must be 3 (first-parent), got %d", res.Behind)
	}

	// The same distance is a refusal once K is below it, and the message
	// carries the first-parent number, never the plain commit count.
	res = EvaluateBaseFreshness(f.input(2, true), BaseFreshnessSeams{})
	if res.Admit || !strings.Contains(res.Refusal, "3 landings behind") {
		t.Fatalf("K=2 must refuse at 3 landings behind: %+v", res)
	}
	if strings.Contains(res.Refusal, "15") {
		t.Fatalf("the refusal must not report the plain commit count: %s", res.Refusal)
	}

	// A base the tip does not descend from is stale whatever K is.
	g := newBFFixture(t)
	g.writeRecord(g.orphanCommit(), nil)
	for _, k := range []int{0, 2, 1000} {
		r := EvaluateBaseFreshness(g.input(k, true), BaseFreshnessSeams{})
		if r.Admit || !r.Stale {
			t.Fatalf("a non-ancestor base must refuse at K=%d: %+v", k, r)
		}
		if !strings.HasPrefix(r.Refusal, BaseFreshnessSentinel) {
			t.Fatalf("the stale refusal must start with the sentinel: %q", r.Refusal)
		}
	}
}

func TestBaseFreshnessRemedyNamesExplicitBranch(t *testing.T) {
	// Follow-up (card t1834): an acquire that carried an explicit --branch gets a
	// remedy naming `remeasure --branch <that branch>` on every refusal path;
	// without it the text stays byte-identical.
	const plainAbsorb = "Absorb local main into the card branch, re-run moai integration remeasure, then retry."
	const plainRecord = "Run moai integration remeasure after the last absorb, then retry."
	cases := []struct {
		name  string
		build func(t *testing.T) *bfFixture
		plain string
		named string
	}{
		{"landings behind", func(t *testing.T) *bfFixture {
			f := newBFFixture(t)
			f.writeRecord(f.base, nil)
			f.landEmpty(1)
			return f
		}, plainAbsorb, "Absorb local main into the card branch, re-run moai integration remeasure --branch main, then retry."},
		{"stale base", func(t *testing.T) *bfFixture {
			f := newBFFixture(t)
			f.writeRecord(f.orphanCommit(), nil)
			return f
		}, plainAbsorb, "Absorb local main into the card branch, re-run moai integration remeasure --branch main, then retry."},
		{"no record", func(t *testing.T) *bfFixture { return newBFFixture(t) },
			plainRecord, "Run moai integration remeasure --branch main after the last absorb, then retry."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := tc.build(t)
			plain := EvaluateBaseFreshness(f.input(0, true), BaseFreshnessSeams{})
			if plain.Admit || !strings.HasSuffix(plain.Refusal, tc.plain) || strings.Contains(plain.Refusal, "--branch") {
				t.Fatalf("without an explicit branch the remedy must stay %q: %+v", tc.plain, plain)
			}
			in := f.input(0, true)
			in.ExplicitBranch = true
			named := EvaluateBaseFreshness(in, BaseFreshnessSeams{})
			if named.Admit || !strings.HasSuffix(named.Refusal, tc.named) {
				t.Fatalf("with an explicit branch the remedy must end %q: %+v", tc.named, named)
			}
			requireOneLine(t, "refusal", named.Refusal)
		})
	}
}

func TestBaseFreshnessKBoundary(t *testing.T) {
	// AC-ISB-004: landings behind equal to K admits, K+1 refuses, and the
	// refusal names the numbers, the key, the tip, and the remedy.
	for _, k := range []int{0, 2} {
		t.Run(fmt.Sprintf("K=%d", k), func(t *testing.T) {
			f := newBFFixture(t)
			f.writeRecord(f.base, nil)
			f.landEmpty(k)
			at := EvaluateBaseFreshness(f.input(k, true), BaseFreshnessSeams{})
			if !at.Admit || at.Behind != k {
				t.Fatalf("behind == K (%d) must admit: %+v", k, at)
			}
			f.landEmpty(1)
			over := EvaluateBaseFreshness(f.input(k, true), BaseFreshnessSeams{})
			if over.Admit || over.Behind != k+1 {
				t.Fatalf("behind == K+1 (%d) must refuse: %+v", k+1, over)
			}
			for _, want := range []string{
				BaseFreshnessSentinel,
				fmt.Sprintf("%d landings behind", k+1),
				fmt.Sprintf("%s=%d", bfKey, k),
				ShortSHA(f.tip()),
				"remeasure",
				"retry",
			} {
				if !strings.Contains(over.Refusal, want) {
					t.Errorf("refusal lacks %q: %s", want, over.Refusal)
				}
			}
			requireOneLine(t, "refusal", over.Refusal)
		})
	}
}

func TestBaseFreshnessRequiresValidRecordForCurrentTree(t *testing.T) {
	// AC-ISB-005: the record is required whatever L is — four record states
	// for the current HEAD tree, at K=0 and again at K=2, with L=0 throughout.
	for _, k := range []int{0, 2} {
		cases := []struct {
			name   string
			setup  func(f *bfFixture)
			refuse bool
			cause  string
		}{
			{"absent", func(f *bfFixture) {}, true, "no re-measure record"},
			{"only for another tree", func(f *bfFixture) {
				// a record keyed to the pre-absorb tree (main's tree), not the card's
				other := bfGit(f.t, f.repo, "rev-parse", "main^{tree}")
				rec := RemeasureRecord{Tree: other, Base: f.base, Command: "true", BuildIdentity: "moai test"}
				if err := WriteRemeasureRecord(f.root, other, rec); err != nil {
					f.t.Fatal(err)
				}
			}, true, "no re-measure record"},
			{"exit code 7", func(f *bfFixture) {
				f.writeRecord(f.base, func(r *RemeasureRecord) { r.ExitCode = 7 })
			}, true, "exited 7"},
			{"valid", func(f *bfFixture) { f.writeRecord(f.base, nil) }, false, ""},
		}
		for _, tc := range cases {
			t.Run(fmt.Sprintf("K=%d/%s", k, tc.name), func(t *testing.T) {
				f := newBFFixture(t)
				tc.setup(f)
				res := EvaluateBaseFreshness(f.input(k, true), BaseFreshnessSeams{})
				if tc.refuse {
					if res.Admit {
						t.Fatalf("must refuse: %+v", res)
					}
					if !strings.HasPrefix(res.Refusal, BaseFreshnessSentinel) ||
						!strings.Contains(res.Refusal, "re-measure record") ||
						!strings.Contains(res.Refusal, ShortSHA(f.tree())) ||
						!strings.Contains(res.Refusal, "remeasure") ||
						!strings.Contains(res.Refusal, tc.cause) {
						t.Fatalf("record refusal text is wrong (want cause %q): %s", tc.cause, res.Refusal)
					}
					return
				}
				if !res.Admit || res.Refusal != "" {
					t.Fatalf("a valid record at L=0 must admit: %+v", res)
				}
			})
		}
	}
}

func TestBaseFreshnessFailOpenCauses(t *testing.T) {
	// AC-ISB-009 (pure half): an applicable call that cannot reach a verdict
	// admits with exactly one line naming the cause, and never prints the
	// sentinel; the not-applicable cases admit silently. The tree is stale
	// throughout, so a gate that treated uncertainty as staleness would refuse.
	stale := func(t *testing.T) *bfFixture {
		f := newBFFixture(t)
		f.writeRecord(f.base, nil)
		f.landEmpty(1)
		return f
	}
	t.Run("uncertain", func(t *testing.T) {
		cases := []struct {
			name  string
			build func(f *bfFixture) (BaseFreshnessInput, BaseFreshnessSeams)
			cause string
		}{
			{"unreadable configuration", func(f *bfFixture) (BaseFreshnessInput, BaseFreshnessSeams) {
				in := f.input(0, true)
				in.ConfigErr = errors.New("workflow.yaml: unreadable")
				return in, BaseFreshnessSeams{}
			}, "unreadable"},
			{"negative K", func(f *bfFixture) (BaseFreshnessInput, BaseFreshnessSeams) {
				return f.input(-1, true), BaseFreshnessSeams{}
			}, "negative"},
			{"integration branch unresolvable", func(f *bfFixture) (BaseFreshnessInput, BaseFreshnessSeams) {
				in := f.input(0, true)
				in.IntegrationBranch = "no-such-branch"
				return in, BaseFreshnessSeams{}
			}, "no-such-branch"},
			{"integration branch unset", func(f *bfFixture) (BaseFreshnessInput, BaseFreshnessSeams) {
				in := f.input(0, true)
				in.IntegrationBranch = ""
				return in, BaseFreshnessSeams{}
			}, "integration branch"},
			{"git failure on rev-list", func(f *bfFixture) (BaseFreshnessInput, BaseFreshnessSeams) {
				return f.input(0, true), BaseFreshnessSeams{Git: failingGitIn(f.card, "rev-list")}
			}, "git rev-list could not run"},
			{"git failure on rev-parse", func(f *bfFixture) (BaseFreshnessInput, BaseFreshnessSeams) {
				return f.input(0, true), BaseFreshnessSeams{Git: failingGitIn(f.card, "rev-parse")}
			}, "git rev-parse could not run"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				f := stale(t)
				in, seams := tc.build(f)
				res := EvaluateBaseFreshness(in, seams)
				if !res.Applicable || !res.Admit || res.Refusal != "" {
					t.Fatalf("an uncertain applicable call must admit without refusing: %+v", res)
				}
				requireOneLine(t, "notice", res.Notice)
				if !strings.HasPrefix(res.Notice, "base-freshness: not measured") || !strings.Contains(res.Notice, tc.cause) {
					t.Fatalf("the notice must name the cause %q: %q", tc.cause, res.Notice)
				}
				if strings.Contains(res.Notice, BaseFreshnessSentinel) {
					t.Fatalf("an uncertain notice must not carry the sentinel: %q", res.Notice)
				}
			})
		}
	})
	t.Run("not applicable", func(t *testing.T) {
		cases := []struct {
			name  string
			build func(f *bfFixture) BaseFreshnessInput
		}{
			{"no card id", func(f *bfFixture) BaseFreshnessInput {
				in := f.input(0, true)
				in.CardID = ""
				return in
			}},
			{"cwd on the integration branch", func(f *bfFixture) BaseFreshnessInput {
				in := f.input(0, true)
				in.Dir = f.repo
				return in
			}},
			{"detached HEAD", func(f *bfFixture) BaseFreshnessInput {
				bfGit(f.t, f.card, "checkout", "-q", "--detach")
				return f.input(0, true)
			}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				f := stale(t)
				res := EvaluateBaseFreshness(tc.build(f), BaseFreshnessSeams{})
				if res.Applicable || !res.Admit || res.Refusal != "" || res.Notice != "" {
					t.Fatalf("a not-applicable call must admit silently: %+v", res)
				}
			})
		}
	})
}

// failingGitIn is a git-runner seam bound to dir that fails every call whose
// first argument is verb and forwards the rest to a real git.
func failingGitIn(dir, verb string) func(args ...string) (string, error) {
	real := factorylane.ExecGitRunner{Dir: dir}
	return func(args ...string) (string, error) {
		if len(args) > 0 && args[0] == verb {
			return "", errors.New("injected " + verb + " failure\nsecond line")
		}
		return real.Git(args...)
	}
}

// seamOverriding is a git-runner seam bound to dir that answers the calls
// match selects with (out, err) and forwards every other call to a real git.
func seamOverriding(dir string, match func(args []string) bool, out string, err error) func(args ...string) (string, error) {
	real := factorylane.ExecGitRunner{Dir: dir}
	return func(args ...string) (string, error) {
		if match(args) {
			return out, err
		}
		return real.Git(args...)
	}
}

func TestBaseFreshnessUncertainGitAnswers(t *testing.T) {
	// A git answer the gate cannot read is uncertainty, never staleness: the
	// tree read failing, the ancestry probe failing with a code other than its
	// "no" (exit 1), and a count that is not a number each admit with one line
	// naming the cause. The exit-1 control proves the probe still answers "no".
	verb := func(name string) func([]string) bool {
		return func(args []string) bool { return len(args) > 0 && args[0] == name }
	}
	cases := []struct {
		name  string
		seam  func(f *bfFixture) func(args ...string) (string, error)
		cause string
	}{
		{"tree read fails", func(f *bfFixture) func(args ...string) (string, error) {
			return seamOverriding(f.card, func(args []string) bool {
				return len(args) > 1 && args[len(args)-1] == "HEAD^{tree}"
			}, "", errors.New("injected tree read failure"))
		}, "git rev-parse could not run"},
		{"ancestry probe fails with a code other than 1", func(f *bfFixture) func(args ...string) (string, error) {
			return seamOverriding(f.card, verb("merge-base"), "",
				&factorylane.GitExitError{ExitCode: 128, Stderr: "fatal: Not a valid commit name"})
		}, "git merge-base failed (exit 128)"},
		{"count is not a number", func(f *bfFixture) func(args ...string) (string, error) {
			return seamOverriding(f.card, verb("rev-list"), "not-a-number\n", nil)
		}, "non-numeric count"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newBFFixture(t)
			f.writeRecord(f.base, nil)
			f.landEmpty(1)
			res := EvaluateBaseFreshness(f.input(0, true), BaseFreshnessSeams{Git: tc.seam(f)})
			if !res.Admit || res.Refusal != "" || res.Stale {
				t.Fatalf("an unreadable git answer must admit, not refuse: %+v", res)
			}
			requireOneLine(t, "notice", res.Notice)
			if !strings.HasPrefix(res.Notice, "base-freshness: not measured") || !strings.Contains(res.Notice, tc.cause) {
				t.Fatalf("the notice must name %q: %q", tc.cause, res.Notice)
			}
		})
	}
	t.Run("control: exit 1 is the probe's no and reads as stale", func(t *testing.T) {
		f := newBFFixture(t)
		f.writeRecord(f.base, nil)
		seam := seamOverriding(f.card, verb("merge-base"), "", &factorylane.GitExitError{ExitCode: 1})
		res := EvaluateBaseFreshness(f.input(0, true), BaseFreshnessSeams{Git: seam})
		if res.Admit || !res.Stale || !strings.HasPrefix(res.Refusal, BaseFreshnessSentinel) {
			t.Fatalf("exit 1 from the ancestry probe is staleness: %+v", res)
		}
	})
}

func TestBaseFreshnessGateOffAdvisesOnlyPure(t *testing.T) {
	// REQ-ISB-008 (pure half): with the gate off a would-be refusal becomes one
	// advisory line and an admit; a fresh candidate stays silent.
	f := newBFFixture(t)
	f.writeRecord(f.base, nil)
	fresh := EvaluateBaseFreshness(f.input(0, false), BaseFreshnessSeams{})
	if !fresh.Admit || fresh.Notice != "" || fresh.Refusal != "" {
		t.Fatalf("a fresh candidate with the gate off must be silent: %+v", fresh)
	}
	f.landEmpty(1)
	res := EvaluateBaseFreshness(f.input(0, false), BaseFreshnessSeams{})
	if !res.Admit || res.Refusal != "" {
		t.Fatalf("the gate off must refuse nothing: %+v", res)
	}
	requireOneLine(t, "advisory", res.Notice)
	if !strings.HasPrefix(res.Notice, "base-freshness (advisory, gate off): would refuse") {
		t.Fatalf("advisory line has the wrong shape: %q", res.Notice)
	}
	if strings.Contains(res.Notice, BaseFreshnessSentinel) {
		t.Fatalf("an advisory must not carry the sentinel: %q", res.Notice)
	}

	// The record case is advised the same way.
	g := newBFFixture(t)
	rec := EvaluateBaseFreshness(g.input(0, false), BaseFreshnessSeams{})
	if !rec.Admit || !strings.HasPrefix(rec.Notice, "base-freshness (advisory, gate off): would refuse") ||
		!strings.Contains(rec.Notice, "re-measure record") {
		t.Fatalf("a missing record with the gate off must be advised: %+v", rec)
	}
}

func TestBaseFreshnessK0MatchesMergeCause2(t *testing.T) {
	// AC-ISB-010: at K=0 the gate tests the predicate merge cause 2 tests
	// (record.Base != tip). Case A is a content-free absorb: the tree and the
	// record are unchanged, the tip moved one landing past the stored base. The
	// gate refuses, the merge step exits 2 on the same fixture, and after the
	// record is rewritten with the new base the gate admits and the merge runs.
	absorb := func(t *testing.T) *stepFixture {
		f := newMergeFixture(t)
		bfGit(t, f.integ, "commit", "-q", "--allow-empty", "-m", "another lane landed")
		// the lane absorbs develop by hand: a merge commit whose tree is unchanged
		bfGit(t, f.cardTree, "merge", "-q", "--no-edit", "develop")
		return f
	}
	gateInput := func(f *stepFixture, k int) BaseFreshnessInput {
		return BaseFreshnessInput{Root: f.root, Dir: f.cardTree, CardID: stepCard, IntegrationBranch: "develop", K: k, Enabled: true}
	}
	t.Run("A1 stale record: gate refuses and merge exits 2", func(t *testing.T) {
		f := absorb(t)
		if tree := bfGit(t, f.cardTree, "rev-parse", "HEAD^{tree}"); tree != f.record.Tree {
			t.Fatalf("fixture sanity: the content-free absorb must keep the tree (%s vs %s)", tree, f.record.Tree)
		}
		res := EvaluateBaseFreshness(gateInput(f, 0), BaseFreshnessSeams{})
		if res.Admit || res.Behind != 1 {
			t.Fatalf("K=0 must refuse a candidate one landing past its stored base: %+v", res)
		}
		card := f.withCardTree(readyCardPtr())
		_, err := RunMergeStep(f.input(), f.seams(card))
		requireCode(t, err, MergeExitBaseMoved)
	})
	t.Run("A2 record rewritten: gate admits and merge proceeds", func(t *testing.T) {
		f := absorb(t)
		newTip := bfGit(t, f.integ, "rev-parse", "develop")
		rewritten := *f.record
		rewritten.Base = newTip
		if err := WriteRemeasureRecord(f.root, rewritten.Tree, rewritten); err != nil {
			t.Fatal(err)
		}
		res := EvaluateBaseFreshness(gateInput(f, 0), BaseFreshnessSeams{})
		if !res.Admit || res.Behind != 0 {
			t.Fatalf("a rewritten record at the tip must admit: %+v", res)
		}
		card := f.withCardTree(readyCardPtr())
		if _, err := RunMergeStep(f.input(), f.seams(card)); err != nil {
			t.Fatalf("the merge must proceed on the rewritten record: %v", err)
		}
	})
	t.Run("B K=3 admits a candidate 2 landings behind; the merge step still exits 2", func(t *testing.T) {
		f := newMergeFixture(t)
		for i := 0; i < 2; i++ {
			bfGit(t, f.integ, "commit", "-q", "--allow-empty", "-m", fmt.Sprintf("landing %d", i))
		}
		res := EvaluateBaseFreshness(gateInput(f, 3), BaseFreshnessSeams{})
		if !res.Admit || res.Behind != 2 {
			t.Fatalf("K=3 must admit 2 landings behind: %+v", res)
		}
		card := f.withCardTree(readyCardPtr())
		_, err := RunMergeStep(f.input(), f.seams(card))
		requireCode(t, err, MergeExitBaseMoved)
		requireWindowReleasedAndCPromoted(t, f)
	})
}

func TestBaseFreshnessGateWritesNothing(t *testing.T) {
	// AC-ISB-014 / REQ-ISB-014: a refusing run changes no byte of the re-measure
	// store, the window record, or the factory database file, and creates no
	// file. A gate that "repaired" a record on the way would fail here.
	f := newBFFixture(t)
	f.writeRecord(f.base, nil)
	f.landEmpty(1)
	state := filepath.Join(f.root, ".moai", "state")
	if err := os.MkdirAll(state, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"factory.db": "database bytes", IntegrationLockFileName: `{"session_id":"sess-x"}`} {
		if err := os.WriteFile(filepath.Join(state, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	before := baseFreshnessSnapshotTree(t, filepath.Join(f.root, ".moai"))
	if len(before) < 3 {
		t.Fatalf("fixture sanity: the snapshot must cover the record, the window and the database, got %d files", len(before))
	}
	res := EvaluateBaseFreshness(f.input(0, true), BaseFreshnessSeams{})
	if res.Admit {
		t.Fatalf("the run under test must be a refusal: %+v", res)
	}
	after := baseFreshnessSnapshotTree(t, filepath.Join(f.root, ".moai"))
	if fmt.Sprint(before) != fmt.Sprint(after) {
		t.Fatalf("the gate wrote: before=%v after=%v", before, after)
	}
}

// baseFreshnessSnapshotTree maps every file under dir to its bytes.
func baseFreshnessSnapshotTree(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		b, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		out[path] = string(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestBaseFreshnessExitCodeIsDistinct(t *testing.T) {
	// REQ-ISB-010 / REQ-ISB-011: exit 16 is not one of the merge step's codes
	// (1-15) nor complete's own (20).
	if BaseFreshnessExitRefused != 16 {
		t.Fatalf("the gate's exit code must be 16, got %d", BaseFreshnessExitRefused)
	}
	taken := []int{
		MergeExitRecordInvalid, MergeExitBaseMoved, MergeExitNotDescendant, MergeExitTreeMismatch,
		MergeExitLandingRefused, MergeExitMergeFailed, MergeExitMergeDirty, MergeExitPostMerge,
		MergeExitOther, MergeExitNothingToMerge, MergeExitCardGate, MergeExitWorktreeDirty,
		MergeExitCollision, MergeExitNotHolder, MergeExitExpiredLease, CompleteExitPostMergeTransitionConflict,
	}
	if len(taken) != 16 {
		t.Fatalf("fixture sanity: expected 16 existing codes, listed %d", len(taken))
	}
	for _, code := range taken {
		if code == BaseFreshnessExitRefused {
			t.Fatalf("exit code %d collides with an existing merge/complete code", code)
		}
	}
}

// thresholdLiteralComparisons returns the comparisons in src that set an
// integer literal against something that reads as the landings-behind count or
// the tolerance K — the shape a hardcoded threshold takes (REQ-ISB-007).
func thresholdLiteralComparisons(t *testing.T, src string) []string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", src, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	// name returns the identifier text of an operand, or "" for anything else.
	name := func(e ast.Expr) string {
		switch v := e.(type) {
		case *ast.Ident:
			return strings.ToLower(v.Name)
		case *ast.SelectorExpr:
			return strings.ToLower(v.Sel.Name)
		}
		return ""
	}
	intLit := func(e ast.Expr) (string, bool) {
		lit, ok := e.(*ast.BasicLit)
		if !ok || lit.Kind != token.INT {
			return "", false
		}
		return lit.Value, true
	}
	// A landings-behind operand set against ANY integer literal is a hardcoded
	// threshold. The tolerance K set against a non-zero literal is one too; K
	// compared with zero is the sign check that rejects a negative setting, not
	// a threshold.
	flagged := func(operand ast.Expr, lit ast.Expr) bool {
		value, ok := intLit(lit)
		if !ok {
			return false
		}
		text := name(operand)
		if strings.Contains(text, "behind") {
			return true
		}
		return text == "k" && value != "0"
	}
	var found []string
	ast.Inspect(file, func(n ast.Node) bool {
		bin, ok := n.(*ast.BinaryExpr)
		if !ok {
			return true
		}
		switch bin.Op {
		case token.GTR, token.GEQ, token.LSS, token.LEQ, token.EQL, token.NEQ:
		default:
			return true
		}
		if flagged(bin.Y, bin.X) || flagged(bin.X, bin.Y) {
			found = append(found, fset.Position(bin.Pos()).String())
		}
		return true
	})
	return found
}

func TestBaseFreshnessDecisionCodeHasNoThresholdLiteral(t *testing.T) {
	// AC-ISB-007's last clause: the decision code compares landings behind
	// against the loaded K, never against a literal.
	control := "package x\nfunc f(behind, k int) bool { return behind > 0 || k == 2 || 3 < behind }\n"
	if got := thresholdLiteralComparisons(t, control); len(got) != 3 {
		t.Fatalf("positive control: the scanner must flag all three literal comparisons, got %v", got)
	}
	signCheck := "package x\nfunc f(k int) bool { return k < 0 }\n"
	if got := thresholdLiteralComparisons(t, signCheck); len(got) != 0 {
		t.Fatalf("negative control: a sign check of K against zero is not a threshold, got %v", got)
	}
	src, err := os.ReadFile("base_freshness.go")
	if err != nil {
		t.Fatalf("read the gate source: %v", err)
	}
	if got := thresholdLiteralComparisons(t, string(src)); len(got) != 0 {
		t.Fatalf("the gate compares landings behind or K against an integer literal at %v", got)
	}
}

func TestBaseFreshnessRefusesRecordNamingAnotherTree(t *testing.T) {
	// AC-ISB-005 (audit-fix round): a record FILE named for the current tree
	// whose inner Tree field names a different tree is not a record for the
	// current tree. ReadRemeasureRecord returns the field as written (the merge
	// step's cause 4 depends on that), so the gate itself must compare it with
	// the tree it resolved; WriteRemeasureRecord forces the field, so the forged
	// file is written directly. At L=0 the stale-base arms cannot mask it.
	for _, k := range []int{0, 2} {
		t.Run(fmt.Sprintf("K=%d", k), func(t *testing.T) {
			f := newBFFixture(t)
			current := f.tree()
			other := bfGit(t, f.repo, "rev-parse", "main^{tree}")
			if other == current {
				t.Fatal("fixture: the two trees must differ")
			}
			forged := RemeasureRecord{
				Tree: other, Base: f.base, Command: "true", BuildIdentity: "moai test",
				RecordedAt: time.Date(2026, 10, 18, 9, 0, 0, 0, time.UTC).Format(time.RFC3339),
			}
			data, err := json.Marshal(forged)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(remeasureDir(f.root), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(remeasureDir(f.root), current+".json"), data, 0o644); err != nil {
				t.Fatal(err)
			}

			res := EvaluateBaseFreshness(f.input(k, true), BaseFreshnessSeams{})
			if res.Admit {
				t.Fatalf("a record whose tree field names another tree must refuse: %+v", res)
			}
			requireOneLine(t, "refusal", res.Refusal)
			if !strings.HasPrefix(res.Refusal, BaseFreshnessSentinel) ||
				!strings.Contains(res.Refusal, "re-measure record") ||
				!strings.Contains(res.Refusal, ShortSHA(current)) ||
				!strings.Contains(res.Refusal, ShortSHA(other)) {
				t.Fatalf("the refusal must carry the sentinel and name both trees: %s", res.Refusal)
			}
		})
	}
}
