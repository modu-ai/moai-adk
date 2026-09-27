// Package signtest builds throwaway signing fixtures for tests of
// internal/contract/sign and its callers (the contract CLI, the configuration
// reader). It is test support: every function takes a testing.TB and fails the
// test on a setup error.
//
// A Project is a t.TempDir() tree carrying `.moai/specs/SPEC-FIXTURE-001/`
// (spec.md, acceptance.md, an unsigned draft contract.yaml), a plan-audit
// report file, and a git repository whose local config sets user.name and
// user.email and which has one commit. Every git subprocess runs with
// GIT_DIR, GIT_WORK_TREE, and GIT_INDEX_FILE removed from its environment and
// with the global and system git configuration disabled, so the fixture never
// reads or writes the enclosing repository or the developer's settings.
package signtest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/sign"
)

// Fixture identity.
const (
	SpecID              = "SPEC-FIXTURE-001"
	Card                = "t1234"
	OperatorName        = "Fixture Operator"
	OperatorEmail       = "fixture@example.com"
	PlanAuditReportPath = ".moai/reports/plan-audit/SPEC-FIXTURE-001-review-1.md"
)

// Markers is the closed agent-environment marker set of design.md
// § Agent-Environment Markers, as a test passes it through Options. The CLI
// passes the same names from its own constants.
var Markers = []string{"CLAUDECODE", "CLAUDE_CODE_SESSION_ID"}

// Acceptance variants.
const (
	// Acceptance carries two live AC IDs.
	Acceptance = "# acceptance.md — SPEC-FIXTURE-001\n\n" +
		"### AC-FIXTURE-001 — first\n\nGiven a fixture, when it runs, then it passes.\n\n" +
		"### AC-FIXTURE-002 — second\n\nGiven a fixture, when it runs, then it passes.\n"
	// AcceptanceThree adds a third live AC to Acceptance.
	AcceptanceThree = Acceptance + "\n### AC-FIXTURE-003 — third\n\nGiven a change, when it lands, then it passes.\n"
	// AcceptanceAmbiguous marks AC-FIXTURE-001 [RETIRED] once and leaves it
	// unmarked once, so the counter reports it ambiguous.
	AcceptanceAmbiguous = Acceptance + "\nAC-FIXTURE-001 [RETIRED] superseded by a later criterion.\n"
	// AcceptanceZero carries no AC ID at all.
	AcceptanceZero = "# acceptance.md — SPEC-FIXTURE-001\n\nNo criteria are written yet.\n"
)

// Draft selects the variant of the unsigned draft contract. The zero value is
// a valid draft: two AC-bound sections absent (sign writes them), no budget,
// verdict PASS, actions commit/worktree/local-merge-develop/push-develop.
type Draft struct {
	SpecID         string   // default SpecID
	Verdict        string   // default PASS
	Actions        []string // default commit, worktree, local-merge-develop, push-develop
	RecordedSHA256 string   // written as acceptance.sha256 when non-empty
	WithBudget     bool     // write budget 60/40/2
}

// DraftContract renders an unsigned contract.yaml. The text deliberately
// carries comments and blank lines, so a test can check that signing keeps
// the author's layout.
func DraftContract(d Draft) string {
	if d.SpecID == "" {
		d.SpecID = SpecID
	}
	if d.Verdict == "" {
		d.Verdict = "PASS"
	}
	if d.Actions == nil {
		d.Actions = []string{"commit", "worktree", "local-merge-develop", "push-develop"}
	}
	var b strings.Builder
	w := func(format string, a ...any) { fmt.Fprintf(&b, format+"\n", a...) }
	w("# Contract for the fixture SPEC.")
	w("schema_version: 1")
	w("spec_id: %s", d.SpecID)
	w("card: %s", Card)
	w("")
	w("acceptance:")
	w("  file: acceptance.md   # the SPEC's acceptance criteria")
	if d.RecordedSHA256 != "" {
		w("  sha256: %q", d.RecordedSHA256)
	}
	w("")
	w("# Invariants the run must keep.")
	w("invariants:")
	w("  - frozen-files")
	w("  - \"go test ./...\"")
	w("")
	w("ownership:")
	w("  write:")
	w("    - \"internal/fixture/**\"")
	w("    - \".moai/specs/%s/**\"", d.SpecID)
	w("  never:")
	w("    - \"internal/x/**\"")
	w("  scratch:")
	w("    - \".moai/state/verify/**\"")
	w("")
	w("approach: \"Implement the fixture feature in one package.\"")
	w("")
	if len(d.Actions) == 0 {
		w("actions: []")
	} else {
		w("actions:")
		for _, a := range d.Actions {
			w("  - %s", a)
		}
	}
	w("")
	w("reobserve:")
	w("  - contract.yaml")
	w("  - acceptance.md")
	w("")
	w("review:")
	w("  second_model: codex")
	w("  human: closure-report")
	if d.WithBudget {
		w("")
		w("budget:")
		w("  turns: 60")
		w("  operations: 40")
		w("  audit_retries: 2")
	}
	w("")
	w("escalate_on:")
	for _, t := range []string{"acceptance-change", "invariant-violation", "ownership-move",
		"new-architecture-or-api", "contradictory-evidence", "irreversible-action"} {
		w("  - %s", t)
	}
	w("")
	w("plan_audit:")
	w("  verdict: %s", d.Verdict)
	return b.String()
}

// Project is one fixture project.
type Project struct {
	t    testing.TB
	Root string
	// Head is the fixture repository's HEAD commit.
	Head string
}

// New creates a fixture project with SPEC SpecID (a valid unsigned draft)
// and a committed git repository.
func New(t testing.TB) *Project {
	t.Helper()
	p := &Project{t: t, Root: t.TempDir()}
	p.AddSpec(SpecID, Draft{}, Acceptance)
	p.WriteFile(PlanAuditReportPath, "# plan-audit report — SPEC-FIXTURE-001\n\nVerdict: PASS\n")
	p.Git("init", "-q")
	p.Git("config", "user.name", OperatorName)
	p.Git("config", "user.email", OperatorEmail)
	p.Git("config", "commit.gpgsign", "false")
	// A commit otherwise starts a detached `git maintenance run --auto` that
	// keeps writing .git after the commit returns, racing Snapshot's walk.
	p.Git("config", "maintenance.auto", "false")
	p.Git("add", "-A")
	p.Git("commit", "-q", "-m", "fixture")
	p.Head = strings.TrimSpace(p.Git("rev-parse", "HEAD"))
	return p
}

// AddSpec writes spec.md, acceptance.md, and contract.yaml for id.
func (p *Project) AddSpec(id string, d Draft, acceptance string) {
	p.t.Helper()
	d.SpecID = id
	p.WriteFile(SpecRel(id, "spec.md"), "---\nid: "+id+"\nstatus: draft\n---\n\n# "+id+"\n")
	p.WriteFile(SpecRel(id, contract.AcceptanceFile), acceptance)
	p.WriteFile(SpecRel(id, contract.ContractFile), DraftContract(d))
}

// SpecRel returns the slash path of a file inside a SPEC directory.
func SpecRel(id, name string) string { return ".moai/specs/" + id + "/" + name }

// ReceiptRel is the fixed receipt path of SpecID.
func ReceiptRel() string { return SpecRel(SpecID, contract.ReceiptFile) }

// Path returns the absolute path of a slash-separated project-relative path.
func (p *Project) Path(rel string) string {
	return filepath.Join(p.Root, filepath.FromSlash(rel))
}

// WriteFile writes a project-relative file, creating parent directories.
func (p *Project) WriteFile(rel, content string) {
	p.t.Helper()
	path := p.Path(rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		p.t.Fatalf("signtest: mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		p.t.Fatalf("signtest: write %s: %v", rel, err)
	}
}

// ReadFile reads a project-relative file.
func (p *Project) ReadFile(rel string) []byte {
	p.t.Helper()
	data, err := os.ReadFile(p.Path(rel))
	if err != nil {
		p.t.Fatalf("signtest: read %s: %v", rel, err)
	}
	return data
}

// Git runs git in the project root with a scrubbed environment and returns
// its standard output.
func (p *Project) Git(args ...string) string {
	p.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = p.Root
	cmd.Env = append(ScrubbedEnv(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		p.t.Fatalf("signtest: git %v: %v: %s", args, err, errb.String())
	}
	return out.String()
}

// ScrubbedEnv returns os.Environ() without GIT_DIR, GIT_WORK_TREE, and
// GIT_INDEX_FILE.
func ScrubbedEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		switch name {
		case "GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE":
			continue
		}
		env = append(env, kv)
	}
	return env
}

// Snapshot records the bytes of every regular file under the project root.
func (p *Project) Snapshot() map[string]string {
	p.t.Helper()
	snap := map[string]string{}
	err := filepath.WalkDir(p.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(p.Root, path)
		snap[filepath.ToSlash(rel)] = string(data)
		return nil
	})
	if err != nil {
		p.t.Fatalf("signtest: snapshot: %v", err)
	}
	return snap
}

// AssertUnchanged fails t when any file differs from snap, was added, or was
// removed.
func (p *Project) AssertUnchanged(t testing.TB, snap map[string]string) {
	t.Helper()
	now := p.Snapshot()
	var diff []string
	for rel, before := range snap {
		after, ok := now[rel]
		switch {
		case !ok:
			diff = append(diff, "removed "+rel)
		case after != before:
			diff = append(diff, "changed "+rel)
		}
	}
	for rel := range now {
		if _, ok := snap[rel]; !ok {
			diff = append(diff, "added "+rel)
		}
	}
	slices.Sort(diff)
	if len(diff) > 0 {
		t.Errorf("fixture files not byte-identical: %v", diff)
	}
}

// Options returns human-path signing options for the given SPEC IDs with the
// fixture's policy: mode guided, second_review required, push_develop true,
// budget 60/40/2, the fixture markers.
func (p *Project) Options(ids ...string) sign.Options {
	if len(ids) == 0 {
		ids = []string{SpecID}
	}
	return sign.Options{
		ProjectRoot:      p.Root,
		SpecIDs:          ids,
		Mode:             "guided",
		SecondReview:     "required",
		PushDevelop:      true,
		Decider:          "human",
		JevEnabled:       true,
		JevMinConfidence: 0.5,
		BudgetDefault:    contract.Budget{Turns: 60, Operations: 40, AuditRetries: 2},
		AgentMarkers:     slices.Clone(Markers),
	}
}

// ReceiptOptions returns receipt-path options: mode contract, the given
// effective decider as configuration, signer as --signer, and the fixed
// receipt path.
func (p *Project) ReceiptOptions(decider, signer string) sign.Options {
	o := p.Options()
	o.Mode = "contract"
	o.Decider = decider
	o.Signer = signer
	o.ReceiptPath = ReceiptRel()
	return o
}

// Recorder observes the interactive seams of one Sign call.
type Recorder struct {
	Out       bytes.Buffer
	ReadLines int
	TTYCalls  int
	lines     []string
	env       map[string]string
}

// Seams returns seams that report tty for the terminal check, answer
// ReadLine with lines in order (io.EOF once exhausted), read the environment
// from env only (nothing else is set), and write output to the recorder.
// The git and time seams are left nil, so Sign uses its defaults.
func Seams(tty bool, env map[string]string, lines ...string) (sign.Seams, *Recorder) {
	r := &Recorder{lines: lines, env: env}
	return sign.Seams{
		IsTTY: func() bool {
			r.TTYCalls++
			return tty
		},
		Getenv: func(k string) string { return r.env[k] },
		ReadLine: func() (string, error) {
			r.ReadLines++
			if len(r.lines) == 0 {
				return "", io.EOF
			}
			l := r.lines[0]
			r.lines = r.lines[1:]
			return l, nil
		},
		Out: &r.Out,
	}, r
}

// Load returns the verify inputs of a SPEC with the given options' policy.
func (p *Project) Load(id string, o sign.Options) contract.Inputs {
	p.t.Helper()
	dir, err := contract.ResolveSpecDir(p.Root, id)
	if err != nil {
		p.t.Fatalf("signtest: resolve %s: %v", id, err)
	}
	in, err := contract.LoadDir(dir)
	if err != nil {
		p.t.Fatalf("signtest: load %s: %v", id, err)
	}
	in.Policy = contract.Policy{
		SecondReview:  o.SecondReview,
		PushDevelop:   o.PushDevelop,
		Mode:          o.Mode,
		BudgetDefault: o.BudgetDefault,
	}
	in.RegistryRuleIDs = o.RegistryRuleIDs
	in.RegistryFrozenFiles = o.RegistryFrozenFiles
	return in
}

// Verify runs contract.Verify on a SPEC's current files.
func (p *Project) Verify(id string, o sign.Options) contract.Report {
	return contract.Verify(p.Load(id, o))
}

// Receipt returns a kickoff receipt for SpecID that validates under
// configuration decider llm and --signer llm: inputs bound to the current
// signable digest, acceptance hash, and plan-audit report; an llm_answer
// approve; outcome approve. mutate, when non-nil, edits it before encoding.
func (p *Project) Receipt(o sign.Options, mutate func(*contract.KickoffReceipt)) []byte {
	p.t.Helper()
	rep := p.Verify(SpecID, o)
	conf := 0.82
	r := contract.KickoffReceipt{
		ReceiptVersion:   1,
		SpecID:           SpecID,
		RequestedDecider: "llm",
		EffectiveDecider: "llm",
		Fallback:         &contract.ReceiptFallback{Applied: false},
		IssuedAt:         "2026-09-26T09:00:00Z",
		Inputs: &contract.ReceiptInputs{
			ContractSHA256:   rep.SignableContractSHA256,
			AcceptanceSHA256: rep.Acceptance.MeasuredSHA256,
			PlanAuditReport: &contract.ReceiptFileRef{
				Path:   PlanAuditReportPath,
				SHA256: SHA256Hex(p.ReadFile(PlanAuditReportPath)),
			},
		},
		LLMAnswer: &contract.ReceiptAnswer{
			Answer: "approve", Confidence: &conf,
			Reason: "Scope and actions match the plan.", ReasonRefs: []string{"contract.yaml:2"},
		},
		Outcome: "approve",
	}
	if mutate != nil {
		mutate(&r)
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		p.t.Fatalf("signtest: encode receipt: %v", err)
	}
	return append(data, '\n')
}

// JevAnswer returns a Jev answer carrying every required field.
func JevAnswer(answer string, confidence float64) *contract.ReceiptJevAnswer {
	return &contract.ReceiptJevAnswer{
		ReceiptAnswer: contract.ReceiptAnswer{
			Answer: answer, Confidence: &confidence,
			Reason: "Actions match the contract.", ReasonRefs: []string{"contract.yaml:3"},
		},
		RequestSHA256: strings.Repeat("c", 64),
		RawResponse:   `{"answer":"` + answer + `"}`,
	}
}

// SHA256Hex returns the lowercase-hex SHA-256 of raw bytes.
func SHA256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// WriteReceipt writes data at the fixed receipt path.
func (p *Project) WriteReceipt(data []byte) {
	p.t.Helper()
	p.WriteFile(ReceiptRel(), string(data))
}
