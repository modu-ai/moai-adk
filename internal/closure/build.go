package closure

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/config/atomicfile"
	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/escalation"
)

// BuildInput is everything the closure report builder reads. Every field is
// a file or a measured fact — no agent-supplied flag is an input
// (spec.md §C.5). The section builders live in sections_*.go; this file owns
// the orchestration and the on-disk write.
type BuildInput struct {
	Card   string
	SpecID string
	// Home is the card evidence home (§C.7): A2's NeedsDecision reads the
	// card's record directory under it.
	Home               string
	Mode               string // effective workflow.autonomy.mode
	SecondReviewPolicy string

	GeneratedAt time.Time

	// Verify is A1's verify report for the SPEC's contract at report time;
	// Verify.Contract carries the decoded contract.
	Verify contract.Report

	// Receipt is the raw kickoff-receipt.json (nil when absent).
	Receipt        []byte
	ReceiptPresent bool

	// ProgressMD and AcceptanceMD are the SPEC's lifecycle files (nil absent).
	ProgressMD   []byte
	AcceptanceMD []byte
	// MeasuredAC is contract.CountAC over the normalized acceptance; valid
	// only when ACAvailable.
	MeasuredAC  contract.ACCountResult
	ACAvailable bool

	// SecondReviews and SecondReviewSkipped come from LoadSecondReviews.
	SecondReviews       []SecondReviewRecord
	SecondReviewSkipped []string
	// Verdicts come from LoadVerdictRecords.
	Verdicts []VerdictRecord

	// Records are the parsed A2 escalation records of the card evidence
	// home; UnreadableRecords names the files that failed to parse.
	Records           []escalation.Record
	UnreadableRecords []string
	// DetectorArmed is true when the card audit log shows an arming in force
	// AND the card state file's armed snapshot carries this contract's
	// recorded digest (armed for the contract in force, not disarmed).
	DetectorArmed bool

	// Facts is the injected git access over the card evidence home.
	Facts GitFacts
	// EvalCommit is the report HEAD the currency rule applies to.
	EvalCommit string

	// NewAPICompare runs the read-only class-4 comparison; nil (or an error)
	// renders "not observed" (R2 disposition — A2's comparison is unexported;
	// a follow-up card exports it).
	NewAPICompare NewAPICompareFunc

	// PreviousReportHash is the canonical hash of the closure report on disk
	// at build time ("" when none) — the report the Human Verdict section
	// marks currency against.
	PreviousReportHash string
}

// NewAPICompareFunc runs the class-4 comparison between the card base and
// the report HEAD without writing any escalation record.
type NewAPICompareFunc func() ([]escalation.Addition, error)

// Build assembles the report in the fixed section order (REQ-CLOSURE-002).
// The returned report is deterministic over the inputs apart from
// GeneratedAt.
//
// @MX:ANCHOR: [AUTO] the closure report's single assembly point.
// @MX:REASON: Every section builder mutates the one Report in the fixed
// REQ-CLOSURE-002 order; a second assembly path would let the JSON and the
// Markdown disagree.
func Build(in BuildInput) (*Report, error) {
	r := NewReport(in.Card, in.SpecID, in.EvalCommit, in.Mode, in.SecondReviewPolicy)
	r.GeneratedAt = in.GeneratedAt.UTC().Format(time.RFC3339)

	buildSummary(r, in)
	buildKickoff(r, in)
	buildReconciliation(r, in)
	buildInvariants(r, in)
	buildOwnership(r, in)
	buildNewAPIs(r, in)
	buildEscalations(r, in)
	buildFirstVerdict(r, in)
	buildSecondVerdict(r, in)
	buildPlanAuditBinding(r, in)
	buildHumanVerdict(r, in)
	buildResidualRisk(r, in)
	buildSources(r, in)
	return r, nil
}

// buildSources fills the evidence-role → path map (REQ-CLOSURE-024): the
// path of every evidence file the builder read, "" when it was not read.
func buildSources(r *Report, in BuildInput) {
	specDir := ""
	if in.Home != "" && in.SpecID != "" {
		specDir = filepath.Join(in.Home, ".moai", "specs", in.SpecID)
	}
	r.Sources["contract"] = filepath.Join(specDir, "contract.yaml")
	if in.ProgressMD != nil {
		r.Sources["progress"] = filepath.Join(specDir, "progress.md")
	} else {
		r.Sources["progress"] = ""
	}
	if in.AcceptanceMD != nil {
		r.Sources["acceptance"] = filepath.Join(specDir, "acceptance.md")
	} else {
		r.Sources["acceptance"] = ""
	}
	if in.ReceiptPresent {
		r.Sources["receipt"] = filepath.Join(specDir, "kickoff-receipt.json")
	} else {
		r.Sources["receipt"] = ""
	}
	ev := EvidenceFor(in.Home, in.Card)
	if len(in.SecondReviews) > 0 || len(in.SecondReviewSkipped) > 0 {
		r.Sources["second-review"] = ev.SecondReview
	} else {
		r.Sources["second-review"] = ""
	}
	if len(in.Verdicts) > 0 {
		r.Sources["closure-verdict"] = ev.ClosureVerdict
	} else {
		r.Sources["closure-verdict"] = ""
	}
	if len(in.Records) > 0 || len(in.UnreadableRecords) > 0 {
		r.Sources["escalation-records"] = escalation.RecordDir(in.Home, in.Card)
	} else {
		r.Sources["escalation-records"] = ""
	}
}

// WriteReportPair writes closure-report.md and closure-report.json into the
// card evidence directory atomically (each file replaced by rename) and
// returns the Markdown path the command prints (REQ-CLOSURE-001). The JSON
// is written last: it is the form readers and the verdict recorder hash.
func WriteReportPair(ev EvidencePaths, r *Report) (string, error) {
	if err := os.MkdirAll(ev.Dir, 0o755); err != nil {
		return "", err
	}
	md := RenderMarkdown(r)
	if err := atomicfile.Write(ev.ReportMD, []byte(md), 0o644); err != nil {
		return "", err
	}
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	if err := atomicfile.Write(ev.ReportJSON, data, 0o644); err != nil {
		return "", err
	}
	return ev.ReportMD, nil
}
