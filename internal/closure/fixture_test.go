package closure

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/closure/gitio"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/sign"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
	"github.com/modu-ai/moai-adk/internal/escalation"
)

// Fixture identity (acceptance.md §A fixture).
const (
	fixCard     = "c1"
	fixSpecID   = "SPEC-FIXTURE-001"
	fixOperator = "Fixture Operator"
	fixEmail    = "fixture@example.com"
)

// acFixture is the throwaway git repository of acceptance.md §A: a primary
// checkout on develop, a bare remote holding origin/develop, a linked
// worktree whose base name is the card c1, and a signed SPEC in the card
// worktree. Everything lives under a fresh t.TempDir().
type acFixture struct {
	t *testing.T

	Parent  string // tempdir parent
	Root    string // primary checkout, branch develop
	Remote  string // bare remote
	CardDir string // linked worktree named c1

	EvidenceDir string // <CardDir>/.moai/reports/c1
	EscDir      string // <EvidenceDir>/escalation
}

func newACFixture(t *testing.T) *acFixture {
	t.Helper()
	parent := t.TempDir()
	f := &acFixture{
		t:      t,
		Parent: parent,
		Root:   filepath.Join(parent, "primary"),
		Remote: filepath.Join(parent, "remote.git"),
		// The card worktree's BASE NAME must equal the card id (§C.7).
		CardDir:     filepath.Join(parent, fixCard),
		EvidenceDir: filepath.Join(parent, fixCard, ".moai", "reports", fixCard),
		EscDir: filepath.Join(parent, fixCard, ".moai", "reports", fixCard,
			"escalation"),
	}

	f.git(f.Root, "init", "-b", "develop") // chdir after init below
	f.git(f.Root, "config", "user.name", fixOperator)
	f.git(f.Root, "config", "user.email", fixEmail)
	f.write(filepath.Join(f.Root, "README.md"), "# fixture\n")
	f.git(f.Root, "add", "README.md")
	f.git(f.Root, "commit", "-m", "init")

	// Bare remote holding origin/develop.
	f.git(f.Remote, "init", "--bare")
	f.git(f.Root, "remote", "add", "origin", f.Remote)
	f.git(f.Root, "push", "-q", "origin", "develop")

	// Linked card worktree (git resolves the path's base name c1).
	f.git(f.Root, "worktree", "add", f.CardDir, "-b", "WT-card")

	// The SPEC inside the card worktree.
	specDir := filepath.Join(f.CardDir, ".moai", "specs", fixSpecID)
	f.write(filepath.Join(specDir, "spec.md"),
		"---\nid: "+fixSpecID+"\nstatus: in-progress\n---\n\n# fixture spec\n")
	f.write(filepath.Join(specDir, "acceptance.md"), fixAcceptance)
	f.write(filepath.Join(specDir, "progress.md"), fixProgress)
	f.write(filepath.Join(specDir, "contract.yaml"),
		draftContract(fixCard, fixSpecID,
			[]string{"frozen-files", "go test ./..."},
			[]string{"src/**", ".moai/specs/" + fixSpecID + "/**"}))

	// Sign through A1's signing seam (human path, stubbed environment).
	scrm, _ := signtest.Seams(true, map[string]string{}, fixSpecID)
	scrm.GitHead = func(root string) (string, error) { return gitio.Head(root) }
	res, err := sign.Sign(f.signOptions(), scrm)
	if err != nil {
		t.Fatalf("sign fixture contract: %v", err)
	}
	if res.Refusal != "" {
		t.Fatalf("sign refused: %s", res.Refusal)
	}

	// Commit the signed SPEC on the card branch so HEAD is meaningful.
	f.git(f.CardDir, "add", ".moai")
	f.git(f.CardDir, "commit", "-m", "spec")
	return f
}

func (f *acFixture) signOptions() sign.Options {
	return sign.Options{
		ProjectRoot:      f.CardDir,
		SpecIDs:          []string{fixSpecID},
		Mode:             "contract",
		SecondReview:     "required",
		PushDevelop:      true,
		Decider:          "human",
		JevEnabled:       true,
		JevMinConfidence: 0.5,
		BudgetDefault:    contract.Budget{Turns: 60, Operations: 40, AuditRetries: 2},
		AgentMarkers:     append([]string(nil), signtest.Markers...),
	}
}

// git runs one git command; only "init" runs before a directory exists via
// its parent, so every call targets a directory inside the fixture.
func (f *acFixture) git(dir string, args ...string) {
	f.t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		f.t.Fatalf("mkdir %s: %v", dir, err)
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	// Scrub the enclosing repository's environment (signtest.ScrubbedEnv
	// semantics) so the fixture never reads it.
	cmd.Env = scrubbedEnv()
	if err := cmd.Run(); err != nil {
		f.t.Fatalf("git %s in %s: %v: %s", strings.Join(args, " "), dir, err, strings.TrimSpace(errb.String()))
	}
}

func scrubbedEnv() []string {
	var out []string
	for _, e := range os.Environ() {
		switch {
		case strings.HasPrefix(e, "GIT_DIR="), strings.HasPrefix(e, "GIT_WORK_TREE="),
			strings.HasPrefix(e, "GIT_INDEX_FILE="), strings.HasPrefix(e, "GIT_CONFIG_GLOBAL="):
			continue
		}
		out = append(out, e)
	}
	return append(out, "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
}

func (f *acFixture) write(path, content string) {
	f.t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		f.t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		f.t.Fatalf("write %s: %v", path, err)
	}
}

// facts returns GitFacts over the card worktree via gitio.
func (f *acFixture) facts() GitFacts {
	return GitFacts{
		IsAncestor:      func(a, b string) (bool, error) { return gitio.IsAncestor(f.CardDir, a, b) },
		NonMergeCommits: func(from, to string) ([]CommitPaths, error) { return gitio.NonMergeCommits(f.CardDir, from, to) },
	}
}

// head returns the card worktree's HEAD.
func (f *acFixture) head() string {
	f.t.Helper()
	sha, err := gitio.Head(f.CardDir)
	if err != nil {
		f.t.Fatalf("head: %v", err)
	}
	return sha
}

// verify is A1's verify report over the card worktree's SPEC.
func (f *acFixture) verify() contract.Report {
	f.t.Helper()
	rep, _, _, _, _, err := LoadBuildSideFiles(f.CardDir, fixSpecID,
		config.AutonomySettings{Mode: "contract", SecondReview: "required", PushDevelop: true}, nil, nil)
	if err != nil {
		f.t.Fatalf("load+verify: %v", err)
	}
	return rep
}

// input assembles a BuildInput over the fixture's current files.
func (f *acFixture) input() BuildInput {
	f.t.Helper()
	rep, progress, acceptance, ac, acOK, err := LoadBuildSideFiles(f.CardDir, fixSpecID,
		config.AutonomySettings{Mode: "contract", SecondReview: "required", PushDevelop: true}, nil, nil)
	if err != nil {
		f.t.Fatalf("load+verify: %v", err)
	}
	receipt, receiptOK, _ := readFileOrEmpty(
		filepath.Join(f.CardDir, ".moai", "specs", fixSpecID, contract.ReceiptFile))
	ev := EvidenceFor(f.CardDir, fixCard)
	in := BuildInput{
		Card: fixCard, SpecID: fixSpecID, Home: f.CardDir,
		Mode: "contract", SecondReviewPolicy: "required",
		GeneratedAt:    time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC),
		Verify:         rep,
		Receipt:        receipt,
		ReceiptPresent: receiptOK,
		ProgressMD:     progress,
		AcceptanceMD:   acceptance,
		MeasuredAC:     ac,
		ACAvailable:    acOK,
		Facts:          f.facts(),
		EvalCommit:     f.head(),
	}
	in.PreviousReportHash = PreviousReportHashOf(ev.ReportJSON)
	// A2 records from the card evidence home.
	in.Records, in.UnreadableRecords = mustLoadA2Records(f.t, f.CardDir, fixCard)
	return in
}

// mustLoadA2Records loads the card's escalation records through the
// production loader.
func mustLoadA2Records(t *testing.T, home, card string) ([]escalation.Record, []string) {
	t.Helper()
	recs, unreadable, err := LoadA2Records(home, card)
	if err != nil {
		t.Fatalf("load A2 records: %v", err)
	}
	return recs, unreadable
}

// writeA2Record writes one escalation record file into the card's record
// directory (the records' storage shape is A2's: Markdown + frontmatter).
func (f *acFixture) writeA2Record(rec escalation.Record) {
	f.t.Helper()
	data, err := rec.Marshal()
	if err != nil {
		f.t.Fatalf("marshal record: %v", err)
	}
	if err := os.MkdirAll(f.EscDir, 0o755); err != nil {
		f.t.Fatalf("mkdir escalation dir: %v", err)
	}
	name := rec.Class + "-" + rec.Fingerprint
	if rec.Occurrences > 1 {
		name += fmt.Sprintf("-%d", rec.Occurrences)
	}
	path := filepath.Join(f.EscDir, name+".md")
	f.write(path, string(data))
}

// writeA2Raw writes an arbitrary (possibly malformed) record file.
func (f *acFixture) writeA2Raw(name, content string) {
	f.t.Helper()
	if err := os.MkdirAll(f.EscDir, 0o755); err != nil {
		f.t.Fatalf("mkdir escalation dir: %v", err)
	}
	f.write(filepath.Join(f.EscDir, name), content)
}

// fixAcceptance is the fixture acceptance.md: three live AC IDs.
const fixAcceptance = "# acceptance.md — SPEC-FIXTURE-001\n\n" +
	"### AC-FIXTURE-001 — first\n\nGiven a fixture, when it runs, then it passes.\n\n" +
	"### AC-FIXTURE-002 — second\n\nGiven a fixture, when it runs, then it passes.\n\n" +
	"### AC-FIXTURE-003 — third\n\nGiven a change, when it lands, then it passes.\n"

// fixProgress is the fixture progress.md with §E.2 rows and a §E.3 block.
var fixProgress = "# Progress\n\n" +
	"## §E.1 Plan-phase Audit-Ready Signal\n\n```yaml\nplan_status: audit-ready\n```\n\n" +
	"## §E.2 Run-phase Evidence\n\n" +
	"| AC | Status | Evidence |\n|---|---|---|\n" +
	"| AC-FIXTURE-001 | PASS | `go test -run TestAC ok` |\n" +
	"| AC-FIXTURE-002 | PASS | `go test -run TestAC ok` |\n" +
	"| AC-FIXTURE-003 | FAIL | `go test -run TestAC missing` |\n\n" +
	"## §E.3 Run-phase Audit-Ready Signal\n\n```yaml\n" +
	"run_status: complete\nac_pass_count: 2\nac_fail_count: 1\n" +
	"run_commit_sha: abc1234\n```\n\n" +
	"## §E.4 Sync-phase Audit-Ready Signal\n\n_<pending sync-phase>_\n"

// draftContract renders an unsigned draft contract.yaml for the fixture
// (structure mirrors signtest.DraftContract; the fixture's own values).
func draftContract(card, specID string, invariants, write []string) string {
	var b strings.Builder
	w := func(format string, a ...any) { fmt.Fprintf(&b, format+"\n", a...) }
	w("# Contract for the fixture SPEC.")
	w("schema_version: 1")
	w("spec_id: %s", specID)
	w("card: %s", card)
	w("")
	w("acceptance:")
	w("  file: acceptance.md")
	w("")
	w("invariants:")
	for _, inv := range invariants {
		w("  - %q", inv)
	}
	w("")
	w("ownership:")
	w("  write:")
	for _, g := range write {
		w("    - %q", g)
	}
	w("  never:")
	w("    - \"internal/x/**\"")
	w("")
	w("approach: \"Implement the fixture feature in one package.\"")
	w("")
	w("actions:")
	w("  - commit")
	w("  - worktree")
	w("  - local-merge-develop")
	w("  - push-develop")
	w("")
	w("reobserve:")
	w("  - contract.yaml")
	w("  - acceptance.md")
	w("")
	w("review:")
	w("  second_model: codex")
	w("  human: closure-report")
	w("")
	w("escalate_on:")
	for _, t := range []string{"acceptance-change", "invariant-violation", "ownership-move",
		"new-architecture-or-api", "contradictory-evidence", "irreversible-action"} {
		w("  - %s", t)
	}
	w("")
	w("plan_audit:")
	w("  verdict: PASS")
	return b.String()
}

// writePlanAudit reports helper for the binding tests.
func (f *acFixture) writeCardPlanAudit(name, lastLine string) {
	f.t.Helper()
	f.write(filepath.Join(f.EvidenceDir, name), "# plan audit\n\n"+lastLine+"\n")
}

func (f *acFixture) writeGlobalPlanAudit(name, lastLine string) {
	f.t.Helper()
	f.write(filepath.Join(f.CardDir, ".moai", "reports", "plan-audit", name),
		"# plan audit\n\n"+lastLine+"\n")
}
