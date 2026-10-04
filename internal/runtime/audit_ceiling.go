// audit_ceiling.go — SPEC-AUDIT-CEILING-002: the CLI-counted plan-audit
// iteration ceiling. The counter derives a SPEC's plan-audit round count from
// durable iteration evidence on disk (one recorded file one round), the
// ceiling resolves from the typed harness configuration, and the latest
// verdict is selected by the recorded iteration number — never from
// in-process memory or session state.
//
// The C5 exemption holds: no syscall and no build-tag literal exist in this
// file's scope.
//
// @MX:NOTE: [AUTO] the plan-audit ceiling surface — count, resolve, evaluate, record; the exported functions stay reachable only through the single EvaluatePlanAuditCeiling entry and the spec ceiling CLI verb
// @MX:SPEC: SPEC-AUDIT-CEILING-002
package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/auditverdict"
	"github.com/modu-ai/moai-adk/internal/config"
)

// planAuditRoundFile is one recorded plan-audit file: plan-audit.md ranks as
// iteration 0 and each plan-audit-iter<N>.md carries its parsed N.
type planAuditRoundFile struct {
	path string
	n    int
}

// iterFileName matches the plan-audit-iter<N>.md family; the captured suffix
// must parse as a positive integer. The suffix may be empty (an
// iter-with-no-number file still carries the family shape, so it fails the
// count closed rather than skipping silently).
var iterFileName = regexp.MustCompile(`^plan-audit-iter(.*)\.md$`)

// evidenceDirInput is one deduplicated evidence-directory input carrying its
// directory class.
type evidenceDirInput struct {
	path   string
	listed bool
}

// evidenceDirInputs builds the deduplicated directory inputs: listed
// directories register first so their error-on-absent class wins a dedupe
// collision against the SPEC-scoped class (fail-closed direction).
func evidenceDirInputs(specDir string, listedDirs []string) []evidenceDirInput {
	var inputs []evidenceDirInput
	seen := map[string]bool{}
	for _, d := range listedDirs {
		if d == "" {
			continue
		}
		clean := filepath.Clean(d)
		if seen[clean] {
			continue
		}
		seen[clean] = true
		inputs = append(inputs, evidenceDirInput{path: clean, listed: true})
	}
	if specDir != "" {
		clean := filepath.Clean(specDir)
		if !seen[clean] {
			seen[clean] = true
			inputs = append(inputs, evidenceDirInput{path: clean, listed: false})
		}
	}
	return inputs
}

// evidenceDirList names the deduplicated evidence directories an evaluation
// covered — the record's evidence paths (REQ-ACR-003).
func evidenceDirList(specDir string, listedDirs []string) []string {
	inputs := evidenceDirInputs(specDir, listedDirs)
	paths := make([]string, 0, len(inputs))
	for _, in := range inputs {
		paths = append(paths, in.path)
	}
	return paths
}

// scanEvidenceDirs is the shared scanner behind CountPlanAuditRounds and
// SelectLatestVerdict: it walks the SPEC-scoped directory and the explicitly
// listed directories and returns the plan-audit round files each holds.
//
// Directory-class rules (REQ-ACR-001/013): an absent SPEC-scoped directory
// contributes 0 (an unaudited SPEC is not an error); an explicitly LISTED
// directory that does not exist is an error naming the missing path — a typo
// must never silently count 0. Identical listed paths are deduplicated before
// counting (R6).
func scanEvidenceDirs(specDir string, listedDirs []string) ([]planAuditRoundFile, error) {
	var out []planAuditRoundFile
	for _, in := range evidenceDirInputs(specDir, listedDirs) {
		entries, err := os.ReadDir(in.path)
		if err != nil {
			if in.listed || !os.IsNotExist(err) {
				return nil, fmt.Errorf("read evidence directory %q: %w", in.path, err)
			}
			continue // SPEC-scoped and absent → contributes 0
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if name == "plan-audit.md" {
				out = append(out, planAuditRoundFile{path: filepath.Join(in.path, name), n: 0})
				continue
			}
			m := iterFileName.FindStringSubmatch(name)
			if m == nil {
				continue // not in the plan-audit round family
			}
			n, perr := strconv.Atoi(m[1])
			if perr != nil || n <= 0 {
				return nil, fmt.Errorf("round file %q has an iteration suffix that does not parse as a positive integer", filepath.Join(in.path, name))
			}
			out = append(out, planAuditRoundFile{path: filepath.Join(in.path, name), n: n})
		}
	}
	return out, nil
}

// CountPlanAuditRounds computes a SPEC's plan-audit round count from durable
// iteration evidence on disk, one recorded file one round: plan-audit.md
// contributes one round and each plan-audit-iter<N>.md one round. specDir is
// the SPEC-scoped evidence directory (.moai/reports/<SPEC-ID>); listedDirs are
// the evidence directories explicitly listed in the invocation. The count is
// never derived from in-process memory or session state.
func CountPlanAuditRounds(specDir string, listedDirs []string) (int, error) {
	files, err := scanEvidenceDirs(specDir, listedDirs)
	if err != nil {
		return 0, err
	}
	return len(files), nil
}

// SelectLatestVerdict returns the recorded file with the highest parsed
// iteration number — plan-audit.md ranking as iteration 0 — the verdict the
// disposition branches read. The same highest number appearing in two evidence
// directories is a selection error rather than a silent pick; no evidence at
// all selects nothing (empty path, no error).
func SelectLatestVerdict(specDir string, listedDirs []string) (string, error) {
	files, err := scanEvidenceDirs(specDir, listedDirs)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", nil
	}
	sort.Slice(files, func(i, j int) bool { return files[i].n < files[j].n })
	best := files[len(files)-1]
	var tied []string
	for _, f := range files {
		if f.n == best.n {
			tied = append(tied, f.path)
		}
	}
	if len(tied) > 1 {
		return "", fmt.Errorf("latest plan-audit verdict is ambiguous: iteration %d recorded in %d evidence files (%q and %q); list one evidence directory to disambiguate",
			best.n, len(tied), tied[0], tied[1])
	}
	return best.path, nil
}

// ResolvePlanAuditCeiling resolves the SPEC's ceiling from the configured
// plan_audit_tier_ceilings map. The tier resolves under the shared predicate's
// rule (auditverdict.SpecTier): an absent or unknown tier resolves to L. A
// resolved ceiling that is missing from the map or non-positive is a
// configuration error rather than a ceiling of zero (REQ-ACR-002).
func ResolvePlanAuditCeiling(tier string, ceilings map[string]int) (int, error) {
	resolved := tier
	if resolved != "S" && resolved != "M" && resolved != "L" {
		resolved = "L" // auditverdict.SpecTier's rule: absent or unknown → L
	}
	ceiling, ok := ceilings[resolved]
	if !ok || ceiling <= 0 {
		return 0, fmt.Errorf("plan_audit_tier_ceilings carries no positive ceiling for tier %q (resolved to %q); a mis-configured ceilings map must never resolve to a ceiling of zero", tier, resolved)
	}
	return ceiling, nil
}

// Ceiling disposition values — exactly the three operator-named tokens of
// REQ-ACR-003/D2.
const (
	CeilingDispositionDebtProceed = "debt-proceed"
	CeilingDispositionSplit       = "split"
	CeilingDispositionHold        = "hold"
)

// AuditCeilingStateDir is the machine-local directory the outcome records
// live under, relative to the project root (gitignored parent — the record is
// machine-local state, not a cross-machine durable carrier; SPEC
// §F F2).
const AuditCeilingStateDir = ".moai/state/audit-ceiling"

// CeilingOutcome is one ceiling policy outcome record. Every record carries
// the round count, the ceiling, the latest verdict label, and the evidence
// paths (REQ-ACR-003).
type CeilingOutcome struct {
	SpecID           string   `json:"spec_id"`
	Disposition      string   `json:"disposition"`
	Count            int      `json:"count"`
	Ceiling          int      `json:"ceiling"`
	VerdictLabel     string   `json:"verdict_label"`
	VerdictAdmitted  bool     `json:"verdict_admitted"`
	DebtIDs          []string `json:"debt_ids,omitempty"`
	EvidencePaths    []string `json:"evidence_paths"`
	SplitProposalRef string   `json:"split_proposal_ref,omitempty"`
}

// CeilingInput carries one ceiling evaluation's inputs.
type CeilingInput struct {
	// SpecID is the SPEC whose ceiling is evaluated (names the record file).
	SpecID string
	// SpecEvidenceDir is the SPEC-scoped evidence directory
	// (.moai/reports/<SPEC-ID>); absent contributes 0.
	SpecEvidenceDir string
	// ListedDirs are the evidence directories explicitly listed in the
	// invocation; an absent listing is an error.
	ListedDirs []string
	// Tier is the SPEC's tier; an absent or unknown value resolves to L.
	Tier string
	// Threshold is the plan PASS threshold the admission check uses (the
	// caller resolves it from the SPEC's tier).
	Threshold float64
	// Ceilings is the configured plan_audit_tier_ceilings map.
	Ceilings map[string]int
	// OnFinalHit is the configured plan_audit_ceiling_policy.on_final_hit
	// value; any other or unreadable value fails closed to hold.
	OnFinalHit string
}

// EvaluatePlanAuditCeiling is the single evaluation entry composing select →
// count → ceiling → disposition. The ceiling resolution error propagates
// BEFORE any count/ceiling comparison, and nothing is recorded here — the
// caller records the returned outcome through RecordCeilingOutcome, the one
// recording path (REQ-ACR-003). The returned bool reports whether the ceiling
// applies: a clean admitted PASS at the ceiling is not a ceiling outcome at
// all (REQ-ACR-004 — the ceiling caps repetition, not a healthy result), and
// rounds below the ceiling apply nothing.
func EvaluatePlanAuditCeiling(in CeilingInput) (CeilingOutcome, bool, error) {
	latest, err := SelectLatestVerdict(in.SpecEvidenceDir, in.ListedDirs)
	if err != nil {
		return CeilingOutcome{}, false, err
	}
	count, err := CountPlanAuditRounds(in.SpecEvidenceDir, in.ListedDirs)
	if err != nil {
		return CeilingOutcome{}, false, err
	}
	// The configuration error propagates before any comparison is made (R2's
	// Evaluate-level arm): a mis-configured ceilings map never fires the path.
	ceiling, err := ResolvePlanAuditCeiling(in.Tier, in.Ceilings)
	if err != nil {
		return CeilingOutcome{}, false, err
	}
	if count < ceiling {
		return CeilingOutcome{}, false, nil
	}

	// The hashOK argument is true here by design: the plan-artifact hash
	// binding belongs to the run-entry seams (the kickoff evaluator and the
	// card-transition guard) that run the shared predicate before any ceiling
	// evaluation; this path judges repetition against the verdict's recorded
	// fields, not the tree's current hash state.
	fields := auditverdict.Fields{}
	if latest != "" {
		raw, readErr := os.ReadFile(latest)
		if readErr != nil {
			return CeilingOutcome{}, false, fmt.Errorf("read latest verdict %q: %w", latest, readErr)
		}
		fields = auditverdict.Parse(raw)
	}
	admitted, _ := auditverdict.Admit(fields, auditverdict.PhasePlan, in.Threshold, true)

	if admitted && fields.Label == auditverdict.LabelPass {
		return CeilingOutcome{}, false, nil
	}

	out := CeilingOutcome{
		SpecID:          in.SpecID,
		Count:           count,
		Ceiling:         ceiling,
		VerdictLabel:    fields.Label,
		VerdictAdmitted: admitted,
		EvidencePaths:   evidenceDirList(in.SpecEvidenceDir, in.ListedDirs),
	}
	switch {
	case admitted:
		// The only admitted label left is PASS-WITH-DEBT: record debt-proceed
		// referencing the verdict's debt ids (REQ-ACR-003 arm 1).
		out.Disposition = CeilingDispositionDebtProceed
		for _, d := range fields.Debts {
			out.DebtIDs = append(out.DebtIDs, d.ID)
		}
	case in.OnFinalHit == config.PlanAuditCeilingOnFinalSplit:
		out.Disposition = CeilingDispositionSplit
	case in.OnFinalHit == config.PlanAuditCeilingOnFinalHoldAndSplit:
		// One record carrying both halves of the shipped value: disposition
		// hold plus the split-proposal reference (REQ-ACR-003 arm 2 / D2).
		out.Disposition = CeilingDispositionHold
		out.SplitProposalRef = splitProposalReference(in.SpecID)
	default:
		// Any other or unreadable policy value records hold (fail-closed;
		// REQ-ACR-003 arm 4).
		out.Disposition = CeilingDispositionHold
	}
	return out, true, nil
}

// splitProposalReference is the deterministic split-proposal reference a
// hold-and-split record carries: the SPEC must be split before further plan
// audits, and the reference names what the split applies to.
func splitProposalReference(specID string) string {
	return "split required by plan_audit_ceiling_policy.on_final_hit=hold-and-split: propose splitting " + specID
}

// RecordCeilingOutcome writes the outcome record to
// .moai/state/audit-ceiling/<SPEC-ID>.json (machine-local state). It is the
// ONE recording path: no other code writes a ceiling outcome record.
//
// @MX:ANCHOR: [AUTO] the ONE ceiling-outcome recording path — every recorded plan-audit ceiling outcome is written by this function, never elsewhere
// @MX:REASON: a second writer would fork the record format and let two audit-ceiling truth sources drift silently
// @MX:SPEC: SPEC-AUDIT-CEILING-002
func RecordCeilingOutcome(specID string, outcome CeilingOutcome) error {
	if specID == "" || specID == "." || specID == ".." || strings.ContainsAny(specID, `/\`) {
		return fmt.Errorf("RecordCeilingOutcome: invalid SPEC id %q", specID)
	}
	if err := os.MkdirAll(AuditCeilingStateDir, 0o755); err != nil {
		return fmt.Errorf("RecordCeilingOutcome: create %s: %w", AuditCeilingStateDir, err)
	}
	path := filepath.Join(AuditCeilingStateDir, specID+".json")
	data, err := json.MarshalIndent(outcome, "", "  ")
	if err != nil {
		return fmt.Errorf("RecordCeilingOutcome: marshal: %w", err)
	}
	if err := atomicWrite(path, append(data, '\n')); err != nil {
		return fmt.Errorf("RecordCeilingOutcome: write %s: %w", path, err)
	}
	return nil
}
