// Package closuretest builds the throwaway repository fixture of
// SPEC-AUTONOMY-CLOSURE-001 acceptance.md §A: a primary checkout on
// `develop`, a bare remote holding `origin/develop`, a linked worktree whose
// base name is the card id, and a SPEC-FIXTURE-001 with a contract signed
// through A1's signing seam (card c1, push-develop, ownership
// src/** + .moai/specs/SPEC-FIXTURE-001/**). It is test support for the
// closure, cli, and hook packages: every function takes a testing.TB and
// fails the test on a setup error. No test touches the real .moai/.
package closuretest

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/closure/gitio"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/sign"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
)

// Fixture identity.
const (
	Card       = "c1"
	SpecID     = "SPEC-FIXTURE-001"
	Branch     = "WT-card"
	Operator   = "Fixture Operator"
	Email      = "fixture@example.com"
	RepoBranch = "develop"
)

// Acceptance is the fixture acceptance.md: three live AC IDs.
const Acceptance = "# acceptance.md — SPEC-FIXTURE-001\n\n" +
	"### AC-FIXTURE-001 — first\n\nGiven a fixture, when it runs, then it passes.\n\n" +
	"### AC-FIXTURE-002 — second\n\nGiven a fixture, when it runs, then it passes.\n\n" +
	"### AC-FIXTURE-003 — third\n\nGiven a change, when it lands, then it passes.\n"

// Progress is the fixture progress.md with §E.2 rows and a §E.3 block.
const Progress = "# Progress\n\n" +
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

// Fixture is the assembled throwaway repository.
type Fixture struct {
	t *testing.T

	Parent  string // tempdir parent
	Root    string // primary checkout, branch develop
	Remote  string // bare remote (origin)
	CardDir string // linked worktree named after Card

	EvidenceDir string // <CardDir>/.moai/reports/<Card>
	EscDir      string // <EvidenceDir>/escalation
	SpecDir     string // <CardDir>/.moai/specs/<SpecID>
}

// New builds the fixture: primary checkout, bare remote with origin/develop,
// the card worktree, the SPEC files, and the signed contract.
func New(t *testing.T) *Fixture {
	t.Helper()
	parent := t.TempDir()
	f := &Fixture{
		t:           t,
		Parent:      parent,
		Root:        filepath.Join(parent, "primary"),
		Remote:      filepath.Join(parent, "remote.git"),
		CardDir:     filepath.Join(parent, Card),
		EvidenceDir: filepath.Join(parent, Card, ".moai", "reports", Card),
		EscDir:      filepath.Join(parent, Card, ".moai", "reports", Card, "escalation"),
		SpecDir:     filepath.Join(parent, Card, ".moai", "specs", SpecID),
	}

	f.Git(f.Root, "init", "-b", RepoBranch)
	f.Git(f.Root, "config", "user.name", Operator)
	f.Git(f.Root, "config", "user.email", Email)
	f.Write(filepath.Join(f.Root, "README.md"), "# fixture\n")
	f.Git(f.Root, "add", "README.md")
	f.Git(f.Root, "commit", "-m", "init")

	f.Git(f.Remote, "init", "--bare")
	// The bare remote's default head must resolve, so the review
	// base-resolution chain (origin/HEAD → main) finds it exactly as it
	// would in a real clone.
	f.Git(f.Remote, "symbolic-ref", "HEAD", "refs/heads/"+RepoBranch)
	f.Git(f.Root, "remote", "add", "origin", f.Remote)
	f.Git(f.Root, "push", "-q", "-u", "origin", RepoBranch)
	f.Git(f.Root, "remote", "set-head", "origin", "-a")

	f.Git(f.Root, "worktree", "add", f.CardDir, "-b", Branch)

	f.Write(filepath.Join(f.SpecDir, "spec.md"),
		"---\nid: "+SpecID+"\nstatus: in-progress\n---\n\n# fixture spec\n")
	f.Write(filepath.Join(f.SpecDir, "acceptance.md"), Acceptance)
	f.Write(filepath.Join(f.SpecDir, "progress.md"), Progress)
	f.Write(filepath.Join(f.SpecDir, "contract.yaml"),
		DraftContract(Card, SpecID,
			[]string{"frozen-files", "go test ./..."},
			[]string{"src/**", ".moai/specs/" + SpecID + "/**"}))

	if res, err := sign.Sign(f.SignOptions(), f.SignSeams()); err != nil {
		t.Fatalf("closuretest: sign: %v", err)
	} else if res.Refusal != "" {
		t.Fatalf("closuretest: sign refused: %s", res.Refusal)
	}

	f.Git(f.CardDir, "add", ".moai")
	f.Git(f.CardDir, "commit", "-m", "spec")
	return f
}

// SignOptions is the signing policy the fixture contract is signed under
// (contract mode, second review required, push-develop).
func (f *Fixture) SignOptions() sign.Options {
	return sign.Options{
		ProjectRoot:      f.CardDir,
		SpecIDs:          []string{SpecID},
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

// SignSeams returns human-path seams over the fixture's own git (the
// environment is scrubbed so the host session's markers never leak in).
func (f *Fixture) SignSeams() sign.Seams {
	seams, _ := signtest.Seams(true, map[string]string{}, SpecID)
	seams.GitHead = func(root string) (string, error) { return gitio.Head(root) }
	return seams
}

// Autonomy is the effective autonomy configuration matching the signing.
func Autonomy() config.AutonomySettings {
	return config.AutonomySettings{
		Mode: "contract", SecondReview: "required", PushDevelop: true,
	}
}

// Git runs one git command in dir with the fixture-scrubbed environment.
func (f *Fixture) Git(dir string, args ...string) {
	f.t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		f.t.Fatalf("closuretest: mkdir %s: %v", dir, err)
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	cmd.Env = ScrubbedEnv()
	if err := cmd.Run(); err != nil {
		f.t.Fatalf("closuretest: git %s in %s: %v: %s",
			strings.Join(args, " "), dir, err, strings.TrimSpace(errb.String()))
	}
}

// GitOut runs one git command and returns its trimmed stdout.
func (f *Fixture) GitOut(dir string, args ...string) string {
	f.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	cmd.Env = ScrubbedEnv()
	if err := cmd.Run(); err != nil {
		f.t.Fatalf("closuretest: git %s in %s: %v: %s",
			strings.Join(args, " "), dir, err, strings.TrimSpace(errb.String()))
	}
	return strings.TrimSpace(out.String())
}

// ScrubbedEnv is the environment minus the git variables that would leak the
// enclosing repository into fixture commands.
func ScrubbedEnv() []string {
	var out []string
	for _, e := range os.Environ() {
		switch {
		case strings.HasPrefix(e, "GIT_DIR="),
			strings.HasPrefix(e, "GIT_WORK_TREE="),
			strings.HasPrefix(e, "GIT_INDEX_FILE="):
			continue
		}
		out = append(out, e)
	}
	return append(out, "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
}

// Write writes a fixture file (directories created).
func (f *Fixture) Write(path, content string) {
	f.t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		f.t.Fatalf("closuretest: mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		f.t.Fatalf("closuretest: write %s: %v", path, err)
	}
}

// Head is the card worktree's HEAD.
func (f *Fixture) Head() string {
	sha, err := gitio.Head(f.CardDir)
	if err != nil {
		f.t.Fatalf("closuretest: head: %v", err)
	}
	return sha
}

// DraftContract renders an unsigned draft contract.yaml (structure mirrors
// signtest.DraftContract; the caller's values).
func DraftContract(card, specID string, invariants, write []string) string {
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
