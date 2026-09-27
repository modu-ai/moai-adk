package escalation

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/constitution"
	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/spec"
)

// Not-armed causes named by the resolver's audit lines (REQ-AE-002,
// REQ-AE-023). Each not-armed outcome carries exactly one of them.
const (
	// CauseNoWorktree: no git worktree root contains the working directory,
	// so there is no card id.
	CauseNoWorktree = "no-worktree"
	// CauseNoClaimant: no contract under the worktree root carries a card
	// field equal to the card id.
	CauseNoClaimant = "no-claimant"
	// CauseTerminal: every claimant's SPEC is terminal (completed/archived).
	CauseTerminal = "terminal"
	// CauseMultipleClaimants: two or more non-terminal contracts claim the
	// card; a warning line names every one.
	CauseMultipleClaimants = "multiple-claimants"
	// CauseUnsigned: the sole candidate verifies unsigned.
	CauseUnsigned = "unsigned"
	// CauseSignedInvalid: the sole candidate verifies signed-invalid; a
	// warning line carries the verify reason codes.
	CauseSignedInvalid = "signed-invalid"
	// CauseCardUnreadable names a warning: a contract's card field could not
	// be read, so it could not be considered as a claimant.
	CauseCardUnreadable = "card-unreadable"
)

// Audit line kinds the resolver emits. The card audit log (a later
// milestone) persists them; the resolver only returns them.
const (
	LineNotArmed = "not-armed"
	LineWarning  = "warning"
)

// AuditLine is one resolver audit line.
type AuditLine struct {
	Kind    string   `json:"kind"`
	Card    string   `json:"card"`
	Cause   string   `json:"cause"`
	Detail  string   `json:"detail,omitempty"`
	Specs   []string `json:"specs,omitempty"`
	Reasons []string `json:"reasons,omitempty"`
}

// Resolution is the resolver's answer for one working directory.
type Resolution struct {
	// Card is the card id: the worktree root's base name ("" without one).
	Card string
	// WorktreeRoot is the top-level directory of the containing worktree.
	WorktreeRoot string
	// Armed is true exactly when one non-terminal contract claims the card
	// and it verifies signed-valid.
	Armed bool
	// SpecID and ContractPath name the sole candidate, whether or not it
	// verified signed-valid; both are "" when there is no sole candidate.
	SpecID       string
	ContractPath string
	// Report is the sole candidate's verify report, nil without one.
	Report *contract.Report
	// Lines are the not-armed and warning audit lines, in order.
	Lines []AuditLine
}

// VerifyEnv carries the caller-derived verify inputs: the policy from the
// effective workflow.autonomy settings and the constitution registry values.
type VerifyEnv struct {
	Policy              contract.Policy
	RegistryRuleIDs     []string
	RegistryFrozenFiles []string
}

// PolicyFromSettings maps effective workflow.autonomy settings onto the verify
// policy, the same mapping `moai contract verify` applies.
func PolicyFromSettings(s config.AutonomySettings) contract.Policy {
	return contract.Policy{
		SecondReview: s.SecondReview,
		PushDevelop:  s.PushDevelop,
		Mode:         s.Mode,
		BudgetDefault: contract.Budget{
			Turns:        s.BudgetDefault.Turns,
			Operations:   s.BudgetDefault.Operations,
			AuditRetries: s.BudgetDefault.AuditRetries,
		},
	}
}

// LoadVerifyEnv assembles the verify inputs for the worktree at root in
// process, equivalent to the `moai contract` command's assembly: the policy
// from s, and the constitution registry's rule IDs and distinct Frozen-zone
// file values read from root (or MOAI_CONSTITUTION_REGISTRY when set). An
// absent or unreadable registry yields empty lists, as it does for the CLI.
func LoadVerifyEnv(root string, s config.AutonomySettings) VerifyEnv {
	env := VerifyEnv{Policy: PolicyFromSettings(s), RegistryRuleIDs: []string{}, RegistryFrozenFiles: []string{}}
	path := os.Getenv(constitution.RegistryPathEnv)
	if path == "" {
		path = filepath.Join(root, filepath.FromSlash(constitution.RegistryRelPath))
	}
	reg, err := constitution.LoadRegistry(path, root)
	if err != nil {
		return env
	}
	for _, r := range reg.Entries {
		env.RegistryRuleIDs = append(env.RegistryRuleIDs, r.ID)
	}
	for _, r := range reg.FilterByZone(constitution.ZoneFrozen) {
		if !slices.Contains(env.RegistryFrozenFiles, r.File) {
			env.RegistryFrozenFiles = append(env.RegistryFrozenFiles, r.File)
		}
	}
	slices.Sort(env.RegistryFrozenFiles)
	return env
}

// FindWorktreeRoot returns the nearest ancestor of dir (dir included) that
// holds a `.git` entry — a directory for a primary checkout, a file for a
// linked worktree. It reads only directory entries, never runs git.
func FindWorktreeRoot(dir string) (string, bool) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", false
	}
	for d := abs; ; {
		if _, err := os.Lstat(filepath.Join(d, ".git")); err == nil {
			return d, true
		}
		parent := filepath.Dir(d)
		if parent == d {
			return "", false
		}
		d = parent
	}
}

// claimant is one contract whose card field names the card.
type claimant struct {
	specID string
	dir    string
	status string
}

// Resolve selects the contract that arms the card of the worktree containing
// cwd (design.md §C.8). It takes the card id from the worktree directory
// name, reads every `.moai/specs/*/contract.yaml` under the worktree root,
// keeps those whose card field equals the card id and whose SPEC is not
// terminal, and verifies the sole survivor in process. Neither the branch
// name nor the queue is an input.
//
// Every not-armed outcome carries exactly one not-armed line naming its
// cause; two or more claimants add one warning naming every SPEC ID, and a
// signed-invalid candidate adds one warning carrying the verify reason codes.
// A returned error is an unreadable input — a detector fault (REQ-AE-004).
//
// @MX:ANCHOR: [AUTO] the single contract resolver — card id from the worktree name, SPEC from the contract's card field
// @MX:REASON: lead ruling 09-26 (3) #1 keeps resolution in one function so F1's worktree↔card record can replace step 1 later without touching any class detector; arming (REQ-AE-023), the disarm check (REQ-AE-017), and every class detector consume its answer
func Resolve(cwd string, env VerifyEnv) (Resolution, error) {
	root, ok := FindWorktreeRoot(cwd)
	if !ok {
		return Resolution{Lines: []AuditLine{{Kind: LineNotArmed, Cause: CauseNoWorktree,
			Detail: "no git worktree root contains " + cwd}}}, nil
	}
	res := Resolution{Card: filepath.Base(root), WorktreeRoot: root}

	claimants, warnings, err := findClaimants(root, res.Card)
	if err != nil {
		return Resolution{}, err
	}
	res.Lines = append(res.Lines, warnings...)

	var live []claimant
	for _, c := range claimants {
		if !isTerminal(c.status) {
			live = append(live, c)
		}
	}

	switch {
	case len(live) == 0 && len(claimants) > 0:
		res.notArmed(CauseTerminal, "every contract claiming the card belongs to a terminal SPEC", specIDs(claimants))
		return res, nil
	case len(live) == 0:
		res.notArmed(CauseNoClaimant, "no contract carries card: "+res.Card, nil)
		return res, nil
	case len(live) > 1:
		ids := specIDs(live)
		res.notArmed(CauseMultipleClaimants, fmt.Sprintf("%d contracts claim the card", len(live)), ids)
		res.Lines = append(res.Lines, AuditLine{Kind: LineWarning, Card: res.Card, Cause: CauseMultipleClaimants,
			Detail: "contracts claiming card " + res.Card + ": " + strings.Join(ids, ", "), Specs: ids})
		return res, nil
	}

	c := live[0]
	rep, err := verifySpec(c.dir, c.status, env)
	if err != nil {
		return Resolution{}, err
	}

	res.SpecID = c.specID
	res.ContractPath = filepath.Join(c.dir, contract.ContractFile)
	res.Report = &rep
	switch rep.State {
	case contract.StateSignedValid:
		res.Armed = true
	case contract.StateUnsigned:
		res.notArmed(CauseUnsigned, "the sole candidate contract is unsigned", []string{c.specID})
	default:
		res.notArmed(CauseSignedInvalid, "the sole candidate contract is signed-invalid", []string{c.specID})
		res.Lines = append(res.Lines, AuditLine{Kind: LineWarning, Card: res.Card, Cause: CauseSignedInvalid,
			Detail: c.specID + " verify reasons: " + strings.Join(rep.Reasons, ", "),
			Specs:  []string{c.specID}, Reasons: slices.Clone(rep.Reasons)})
	}
	return res, nil
}

// verifySpec runs A1 verify on one SPEC directory in process, assembling the
// inputs exactly as `moai contract verify` does (contract.LoadDir, then the
// policy, the registry values, and the raw spec.md status). No subprocess and
// no network access (A1 § Non-Functional Constraints), so a PreToolUse hook
// may call it.
func verifySpec(dir, status string, env VerifyEnv) (contract.Report, error) {
	in, err := contract.LoadDir(dir)
	if err != nil {
		return contract.Report{}, fmt.Errorf("escalation: load contract %s: %w", filepath.Base(dir), err)
	}
	in.Policy = env.Policy
	in.RegistryRuleIDs = env.RegistryRuleIDs
	in.RegistryFrozenFiles = env.RegistryFrozenFiles
	in.SpecStatus = status
	return contract.Verify(in), nil
}

// notArmed appends the one not-armed line of a resolution.
func (r *Resolution) notArmed(cause, detail string, specs []string) {
	r.Lines = append(r.Lines, AuditLine{Kind: LineNotArmed, Card: r.Card, Cause: cause, Detail: detail, Specs: specs})
}

// findClaimants reads the card field of every contract under root and
// returns those equal to card, with each SPEC's raw frontmatter status. A
// contract whose card field cannot be read yields a warning, not a claimant.
func findClaimants(root, card string) ([]claimant, []AuditLine, error) {
	paths, err := filepath.Glob(filepath.Join(root, ".moai", "specs", "*", contract.ContractFile))
	if err != nil {
		return nil, nil, fmt.Errorf("escalation: list contracts: %w", err)
	}
	var out []claimant
	var warnings []AuditLine
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, nil, fmt.Errorf("escalation: read %s: %w", p, err)
		}
		got, ok := cardField(data)
		if !ok {
			warnings = append(warnings, AuditLine{Kind: LineWarning, Card: card, Cause: CauseCardUnreadable,
				Detail: "card field unreadable in " + p, Specs: []string{filepath.Base(filepath.Dir(p))}})
			continue
		}
		if got != card {
			continue
		}
		dir := filepath.Dir(p)
		// Same reading as `moai contract verify`: a spec.md without a
		// parseable status reads as "" (not terminal). A spec.md that exists
		// but cannot be read is a fault, never a silent non-terminal.
		status, err := spec.ParseStatus(dir)
		if err != nil {
			if _, rerr := os.ReadFile(filepath.Join(dir, "spec.md")); rerr != nil && !errors.Is(rerr, fs.ErrNotExist) {
				return nil, nil, fmt.Errorf("escalation: read status of %s: %w", filepath.Base(dir), rerr)
			}
			status = ""
		}
		out = append(out, claimant{specID: filepath.Base(dir), dir: dir, status: status})
	}
	return out, warnings, nil
}

// cardField reads only the top-level card field of a contract. The strict
// schema decode belongs to verify; the resolver must read the field even from
// a contract verify would reject, to know which contract to verify.
func cardField(data []byte) (string, bool) {
	var doc struct {
		Card string `yaml:"card"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return "", false
	}
	return doc.Card, true
}

// isTerminal applies A1's terminal derivation to a raw spec.md status. Verify
// derives it from the status alone, before and independently of decoding the
// contract, so an empty contract yields exactly the derived value.
func isTerminal(status string) bool {
	return contract.Verify(contract.Inputs{SpecStatus: status}).Terminal
}

// specIDs lists the SPEC IDs of the claimants, in order.
func specIDs(cs []claimant) []string {
	ids := make([]string, 0, len(cs))
	for _, c := range cs {
		ids = append(ids, c.specID)
	}
	return ids
}
