// Package escalationtest builds throwaway card-worktree fixtures for tests of
// the escalation detector and its hook wiring. It is test support: every
// function takes a testing.TB and fails the test on a setup error.
//
// A Worktree is a directory whose base name is the card id, placed under a
// fresh t.TempDir() parent, carrying a `.git` directory with a HEAD file and
// one branch ref, so worktree-root discovery and HEAD reads work without any
// git subprocess. SPECs are added with AddSpec; a signed SPEC is signed in
// place through internal/contract/sign with every git and terminal seam
// stubbed, so no process is started and no real repository is read.
package escalationtest

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/constitution"
	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/sign"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
)

// Fixture constants.
const (
	// Branch is the fixture's checked-out branch. It deliberately names no
	// card, so a resolver that read the branch would find nothing useful.
	Branch = "WT-other"
	// HeadSHA is the commit the fixture branch ref points at.
	HeadSHA = "0123456789abcdef0123456789abcdef01234567"
	// OperatorName and OperatorEmail are the stubbed signer identity.
	OperatorName  = "Fixture Operator"
	OperatorEmail = "fixture@example.com"
)

// SignedAt is the stubbed signing clock.
var SignedAt = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

// Policy is the verify and sign policy every fixture contract is signed
// under: second review required, push-develop authorized, budget 60/40/2,
// mode contract. The detector must derive the same policy from the fixture's
// workflow.yaml (WriteMode) for a signed fixture to verify signed-valid.
func Policy() contract.Policy {
	return contract.Policy{
		SecondReview:  "required",
		PushDevelop:   true,
		Mode:          "contract",
		BudgetDefault: contract.Budget{Turns: 60, Operations: 40, AuditRetries: 2},
	}
}

// Worktree is one fixture card worktree.
type Worktree struct {
	t testing.TB
	// Root is the worktree root; its base name is Card.
	Root string
	// Card is the card id (the worktree directory base name).
	Card string
}

// NewWorktree creates a worktree directory named card under a fresh temp
// parent, with a `.git` directory whose HEAD names Branch.
func NewWorktree(t testing.TB, card string) *Worktree {
	t.Helper()
	return NewWorktreeIn(t, t.TempDir(), card)
}

// NewWorktreeIn creates a worktree directory named card under parent.
func NewWorktreeIn(t testing.TB, parent, card string) *Worktree {
	t.Helper()
	w := &Worktree{t: t, Root: filepath.Join(parent, card), Card: card}
	w.Write(".git/HEAD", "ref: refs/heads/"+Branch+"\n")
	w.Write(".git/refs/heads/"+Branch, HeadSHA+"\n")
	return w
}

// Path returns the absolute path of a slash-separated worktree-relative path.
func (w *Worktree) Path(rel string) string {
	return filepath.Join(w.Root, filepath.FromSlash(rel))
}

// Write writes a worktree-relative file, creating parent directories.
func (w *Worktree) Write(rel, content string) {
	w.t.Helper()
	p := w.Path(rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		w.t.Fatalf("escalationtest: mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		w.t.Fatalf("escalationtest: write %s: %v", rel, err)
	}
}

// Read reads a worktree-relative file.
func (w *Worktree) Read(rel string) []byte {
	w.t.Helper()
	data, err := os.ReadFile(w.Path(rel))
	if err != nil {
		w.t.Fatalf("escalationtest: read %s: %v", rel, err)
	}
	return data
}

// SetBranch points the fixture HEAD at a branch of the given name (the ref
// file carries HeadSHA).
func (w *Worktree) SetBranch(branch string) {
	w.t.Helper()
	w.Write(".git/HEAD", "ref: refs/heads/"+branch+"\n")
	w.Write(".git/refs/heads/"+branch, HeadSHA+"\n")
}

// WriteMode writes .moai/config/sections/workflow.yaml. An empty mode writes
// an autonomy block without a mode key; modeAbsent writes no autonomy block
// at all. push_develop is always true, matching Policy.
func (w *Worktree) WriteMode(mode string, modeAbsent bool) {
	w.t.Helper()
	var b strings.Builder
	b.WriteString("workflow:\n")
	if !modeAbsent {
		b.WriteString("  autonomy:\n")
		if mode != "" {
			b.WriteString("    mode: " + mode + "\n")
		} else {
			b.WriteString("    mode: \"\"\n")
		}
		b.WriteString("    contract:\n      push_develop: true\n")
	}
	w.Write(".moai/config/sections/workflow.yaml", b.String())
}

// WriteContractMode writes a workflow.yaml selecting mode contract with
// push_develop true (matching Policy), plus any extra lines under
// workflow.autonomy (each indented by the caller relative to autonomy, e.g.
// "escalation:\n  budget_default:\n    operations: 1").
func (w *Worktree) WriteContractMode(autonomyExtra string) {
	w.t.Helper()
	var b strings.Builder
	b.WriteString("workflow:\n  autonomy:\n    mode: contract\n    contract:\n      push_develop: true\n")
	for _, l := range strings.Split(strings.TrimRight(autonomyExtra, "\n"), "\n") {
		if l != "" {
			b.WriteString("    " + l + "\n")
		}
	}
	w.Write(".moai/config/sections/workflow.yaml", b.String())
}

// Replace rewrites the first occurrence of old with new in a
// worktree-relative file and fails the test when old is absent.
func (w *Worktree) Replace(rel, old, new string) {
	w.t.Helper()
	data := string(w.Read(rel))
	if !strings.Contains(data, old) {
		w.t.Fatalf("escalationtest: %s does not contain %q", rel, old)
	}
	w.Write(rel, strings.Replace(data, old, new, 1))
}

// SpecOptions selects the variant of an added SPEC.
type SpecOptions struct {
	// Card is written as the contract's card field; default the worktree's.
	Card string
	// Status is the spec.md frontmatter status line value, written verbatim
	// (quotes included); default in-progress.
	Status string
	// Unsigned leaves the contract an unsigned draft.
	Unsigned bool
	// Verdict is the draft's plan_audit.verdict; default PASS.
	Verdict string
	// Edit, when set, rewrites the draft contract text before signing.
	Edit func(draft string) string
}

// RegistryRelPath is where WriteRegistry puts the zone registry, the default
// path the contract CLI and the detector read.
const RegistryRelPath = ".claude/rules/moai/core/zone-registry.md"

// RegistryEntry is one zone-registry rule.
type RegistryEntry struct {
	ID, Zone, File string
}

// WriteRegistry writes a zone registry holding entries; signing and the
// detector both read it from the worktree.
func (w *Worktree) WriteRegistry(entries ...RegistryEntry) {
	w.t.Helper()
	var b strings.Builder
	b.WriteString("# registry\n\n```yaml\n")
	for i, e := range entries {
		fmt.Fprintf(&b, "- id: %s\n  zone: %s\n  file: %s\n  anchor: \"#a%d\"\n  clause: clause %d\n  canary_gate: true\n",
			e.ID, e.Zone, e.File, i, i)
	}
	b.WriteString("```\n")
	w.Write(RegistryRelPath, b.String())
}

// registry returns the rule IDs and distinct Frozen files of the worktree's
// registry, or empty lists without one.
func (w *Worktree) registry() (ids, frozen []string) {
	reg, err := constitution.LoadRegistry(w.Path(RegistryRelPath), w.Root)
	if err != nil {
		return nil, nil
	}
	for _, r := range reg.Entries {
		ids = append(ids, r.ID)
	}
	for _, r := range reg.FilterByZone(constitution.ZoneFrozen) {
		if !slices.Contains(frozen, r.File) {
			frozen = append(frozen, r.File)
		}
	}
	slices.Sort(frozen)
	return ids, frozen
}

// AddSpec writes spec.md, acceptance.md, and contract.yaml for id and, unless
// o.Unsigned, signs the contract in place on the human path with every seam
// stubbed. A fixture that verify would refuse to sign fails the test.
func (w *Worktree) AddSpec(id string, o SpecOptions) {
	w.t.Helper()
	if o.Card == "" {
		o.Card = w.Card
	}
	if o.Status == "" {
		o.Status = "in-progress"
	}
	dir := ".moai/specs/" + id + "/"
	draft := signtest.DraftContract(signtest.Draft{SpecID: id, Verdict: o.Verdict})
	draft = strings.Replace(draft, "card: "+signtest.Card+"\n", "card: "+o.Card+"\n", 1)
	if o.Edit != nil {
		draft = o.Edit(draft)
	}
	w.Write(dir+"acceptance.md", signtest.Acceptance)
	w.Write(dir+contract.ContractFile, draft)
	if !o.Unsigned {
		w.sign(id)
	}
	// spec.md is written last so a status that verify treats as terminal
	// does not affect signing.
	w.Write(dir+"spec.md", "---\nid: "+id+"\nstatus: "+o.Status+"\n---\n\n# "+id+"\n")
}

// sign signs one SPEC's contract on the human path.
func (w *Worktree) sign(id string) {
	w.t.Helper()
	p := Policy()
	ids, frozen := w.registry()
	opts := sign.Options{
		RegistryRuleIDs:     ids,
		RegistryFrozenFiles: frozen,
		ProjectRoot:         w.Root,
		SpecIDs:             []string{id},
		Mode:                p.Mode,
		SecondReview:        p.SecondReview,
		PushDevelop:         p.PushDevelop,
		Decider:             "human",
		JevEnabled:          true,
		JevMinConfidence:    0.5,
		BudgetDefault:       p.BudgetDefault,
		AgentMarkers:        append([]string(nil), signtest.Markers...),
	}
	answered := false
	seams := sign.Seams{
		IsTTY:  func() bool { return true },
		Getenv: func(string) string { return "" },
		ReadLine: func() (string, error) {
			if answered {
				return "", io.EOF
			}
			answered = true
			return id, nil
		},
		Out:         io.Discard,
		Now:         func() time.Time { return SignedAt },
		GitIdentity: func(string) (string, string, error) { return OperatorName, OperatorEmail, nil },
		GitHead:     func(string) (string, error) { return HeadSHA, nil },
	}
	res, err := sign.Sign(opts, seams)
	if err != nil {
		w.t.Fatalf("escalationtest: sign %s: %v", id, err)
	}
	if res.Refusal != "" {
		w.t.Fatalf("escalationtest: sign %s refused %s (%s)", id, res.Refusal, res.Cause)
	}
}
