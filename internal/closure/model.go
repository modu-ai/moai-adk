// Package closure implements A4 of the contract-autonomy epic
// (SPEC-AUTONOMY-CLOSURE-001): the closure report, the second-review record,
// the human verdict record, and the push readiness rule. Every state is
// derived from files and git; no agent-supplied flag is an input
// (spec.md §C.5). The package imports internal/contract and
// internal/escalation for their schemas; it never imports internal/hook.
//
// The readiness evaluator (readiness.go) and the section builders (build.go,
// sections_*.go) are pure functions over injected inputs; git subprocesses
// live behind internal/closure/gitio.
package closure

import "encoding/json"

// SchemaVersion is the schema_version the report and both record kinds carry.
const SchemaVersion = 1

// Report is the closure report's JSON model (design.md §A.3). Field order is
// load-bearing: encoding/json marshals struct fields in declaration order, and
// REQ-CLOSURE-002 / AC-CLOSURE-002 pin the top-level key order. Add no field
// between the section fields without moving the acceptance criteria too.
//
// The Markdown renderer takes this struct as its only input (design.md §A.3),
// so the two forms cannot disagree.
type Report struct {
	SchemaVersion      int    `json:"schema_version"`
	Card               string `json:"card"`
	SpecID             string `json:"spec_id"`
	HeadSHA            string `json:"head_sha"`
	GeneratedAt        string `json:"generated_at"`
	Mode               string `json:"mode"`
	SecondReviewPolicy string `json:"second_review_policy"`

	Summary          SummarySection          `json:"summary"`
	Kickoff          KickoffSection          `json:"kickoff"`
	Reconciliation   ReconciliationSection   `json:"reconciliation"`
	Invariants       InvariantsSection       `json:"invariants"`
	Ownership        OwnershipSection        `json:"ownership"`
	NewAPIs          NewAPIsSection          `json:"new_apis"`
	Escalations      EscalationsSection      `json:"escalations"`
	FirstVerdict     FirstVerdictSection     `json:"first_verdict"`
	SecondVerdict    SecondVerdictSection    `json:"second_verdict"`
	PlanAuditBinding PlanAuditBindingSection `json:"plan_audit_binding"`
	NotPerformed     []NotPerformed          `json:"not_performed"`
	ResidualRisk     []ResidualRiskItem      `json:"residual_risk"`
	HumanVerdict     HumanVerdictSection     `json:"human_verdict"`

	// Sources maps each evidence role to the path read ("" when unread)
	// (REQ-CLOSURE-024). encoding/json sorts map keys, so the rendering is
	// deterministic.
	Sources map[string]string `json:"sources"`
}

// NewReport returns a report with every section zero-valued, every slice
// non-nil (so it marshals as [] and never null), and the closed not-performed
// and residual-risk lists empty. `mode` and `secondReviewPolicy` carry the
// effective workflow.autonomy values the caller resolved.
func NewReport(card, specID, headSHA, mode, secondReviewPolicy string) *Report {
	return &Report{
		SchemaVersion:      SchemaVersion,
		Card:               card,
		SpecID:             specID,
		HeadSHA:            headSHA,
		Mode:               mode,
		SecondReviewPolicy: secondReviewPolicy,
		Summary:            SummarySection{VerifyReasons: []string{}},
		Reconciliation:     ReconciliationSection{Rows: []ReconciliationRow{}, VerifyReasons: []string{}},
		Invariants:         InvariantsSection{Rows: []InvariantRow{}},
		Ownership:          OwnershipSection{Records: []EscalationRecordView{}},
		NewAPIs:            NewAPIsSection{Additions: []AdditionView{}},
		Escalations:        EscalationsSection{Records: []EscalationRecordView{}, Unreadable: []string{}},
		SecondVerdict:      SecondVerdictSection{Backends: []SecondReviewBackendView{}},
		NotPerformed:       []NotPerformed{},
		ResidualRisk:       []ResidualRiskItem{},
		Sources:            map[string]string{},
	}
}

// SummarySection carries the A1 verify facts; the card, SPEC, HEAD, mode, and
// second-review policy are top-level fields.
type SummarySection struct {
	VerifyState   string   `json:"verify_state"` // A1 state, "not observed" when it could not run
	VerifyReasons []string `json:"verify_reasons"`
	Terminal      *bool    `json:"terminal"` // spec.md status terminal in the report HEAD, nil when unreadable
}

// KickoffSection displays the signature block and, for a receipt signature,
// the receipt (REQ-CLOSURE-010). Receipt content is display-only: no readiness
// decision may read any field of this section.
type KickoffSection struct {
	Method           string   `json:"method"` // signature method, "not recorded" when unsigned
	SignerKind       string   `json:"signer_kind"`
	ReceiptPresent   bool     `json:"receipt_present"`
	RequestedDecider string   `json:"requested_decider"`
	EffectiveDecider string   `json:"effective_decider"`
	FallbackApplied  bool     `json:"fallback_applied"`
	FallbackReason   string   `json:"fallback_reason,omitempty"`
	LLMAnswer        string   `json:"llm_answer"` // "not recorded" when absent
	LLMConfidence    *float64 `json:"llm_confidence"`
	JevAnswer        string   `json:"jev_answer"`
	JevConfidence    *float64 `json:"jev_confidence"`
	Outcome          string   `json:"outcome"`
	VerifyState      string   `json:"verify_state"` // the contract's A1 verify state beside the receipt
}

// ReconciliationRow is one live acceptance-criterion ID of acceptance.md
// (REQ-CLOSURE-004). Status carries the progress.md §E.2 status verbatim, or
// "not reported" when no row exists, or "unknown" for an ID absent from
// acceptance.md.
type ReconciliationRow struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
}

// ReconciliationSection lists one row per live AC ID plus the recorded versus
// measured acceptance binding and the A1 verify state.
type ReconciliationSection struct {
	Rows                  []ReconciliationRow `json:"rows"`
	RecordedAcceptanceSHA string              `json:"recorded_acceptance_sha"` // "not recorded" when unsigned/absent
	MeasuredAcceptanceSHA string              `json:"measured_acceptance_sha"`
	RecordedACCount       string              `json:"recorded_ac_count"` // "not recorded" when unsigned/absent
	MeasuredACCount       int                 `json:"measured_ac_count"`
	VerifyState           string              `json:"verify_state"`
	VerifyReasons         []string            `json:"verify_reasons"`
}

// InvariantRow is one contract invariant and its one result (REQ-CLOSURE-005).
// Kind is constitution, frozen-files, or command.
type InvariantRow struct {
	Invariant string `json:"invariant"`
	Kind      string `json:"kind"`
	Result    string `json:"result"`
}

// Invariant result values (REQ-CLOSURE-005).
const (
	InvariantViolationOpen     = "violation recorded (open)"
	InvariantViolationResolved = "violation recorded (resolved)"
	InvariantNoViolation       = "no violation recorded"
	InvariantNotObserved       = "not observed"
)

// Invariant kinds.
const (
	InvariantKindConstitution = "constitution"
	InvariantKindFrozenFiles  = "frozen-files"
	InvariantKindCommand      = "command"
)

// InvariantsSection lists every contract invariant.
type InvariantsSection struct {
	Rows []InvariantRow `json:"rows"`
}

// EscalationRecordView is the display projection of one A2 escalation record
// (REQ-CLOSURE-008).
type EscalationRecordView struct {
	Kind        string `json:"kind"`
	Class       string `json:"class"`
	Status      string `json:"status"`
	Occurrences int    `json:"occurrences"`
	ContractRef string `json:"contract_ref"`
	Decider     string `json:"decider"`
}

// OwnershipSection lists the card's ownership-move records and the
// violation count (REQ-CLOSURE-006). Count renders as a JSON number only
// while the escalation detector was armed for the contract in force and not
// disarmed; otherwise it renders the string "not observed" (CountValue).
type OwnershipSection struct {
	Records []EscalationRecordView `json:"records"`
	Count   CountValue             `json:"count"`
}

// CountValue marshals as a JSON number when observed and as the string
// "not observed" when not, so a section can carry both shapes under one key
// (REQ-CLOSURE-006, AC-CLOSURE-006).
type CountValue struct {
	Observed bool
	N        int
}

// NotObservedCount is the unobserved count value.
func NotObservedCount() CountValue { return CountValue{} }

// MarshalJSON renders the number, or the "not observed" string.
func (c CountValue) MarshalJSON() ([]byte, error) {
	if !c.Observed {
		return []byte(`"not observed"`), nil
	}
	return json.Marshal(c.N)
}

// Text renders the value for the Markdown form.
func (c CountValue) Text() string {
	if !c.Observed {
		return "not observed"
	}
	b, _ := json.Marshal(c.N)
	return string(b)
}

// UnmarshalJSON is the inverse of MarshalJSON, so a report file round-trips
// through the struct (the canonical hash re-marshals a parsed file).
func (c *CountValue) UnmarshalJSON(data []byte) error {
	if string(data) == `"not observed"` {
		*c = CountValue{}
		return nil
	}
	var n int
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*c = CountValue{Observed: true, N: n}
	return nil
}

// AdditionView is one new architecture or API addition (REQ-CLOSURE-007).
type AdditionView struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// NewAPIsSection lists every observed addition. When the comparison could not
// run, Observed is false and the section states "not observed" rather than an
// empty list (REQ-CLOSURE-007).
type NewAPIsSection struct {
	Additions []AdditionView `json:"additions"`
	Observed  bool           `json:"observed"`
}

// EscalationsSection lists every escalation record of the card, the
// needs-decision state, and the record files that could not be parsed
// (REQ-CLOSURE-008).
type EscalationsSection struct {
	Records       []EscalationRecordView `json:"records"`
	NeedsDecision string                 `json:"needs_decision"` // yes | no | not observed
	Unreadable    []string               `json:"unreadable"`
}

// FirstVerdictSection shows progress.md §E.3 verbatim (REQ-CLOSURE-009).
// Absent fields render "not recorded". The counts are strings so the absent
// form and the recorded form share one field; the recorded text is the
// §E.3 value verbatim.
type FirstVerdictSection struct {
	RunStatus     string `json:"run_status"`
	ACPassCount   string `json:"ac_pass_count"`
	ACFailCount   string `json:"ac_fail_count"`
	RunCommitSHA  string `json:"run_commit_sha"`
	CountMismatch bool   `json:"count_mismatch"`
}

// First-verdict sentinel for absent fields.
const NotRecorded = "not recorded"

// SecondReviewBackendView is one counted backend verdict of the selected
// second-review record (REQ-CLOSURE-014).
type SecondReviewBackendView struct {
	Backend string `json:"backend"`
	Verdict string `json:"verdict"`
}

// Second-review states (REQ-CLOSURE-013 / REQ-CLOSURE-014).
const (
	SecondReviewPerformed    = "performed"
	SecondReviewNotPerformed = "not-performed"
	SecondReviewStale        = "stale"
	SecondReviewNotRequired  = "not-required"
)

// Second-review not-performed causes (REQ-CLOSURE-013).
const (
	SecondReviewCauseNoRecord        = "no-record"
	SecondReviewCauseUnbound         = "unbound"
	SecondReviewCauseContractChanged = "contract-changed"
	SecondReviewCauseScopeNotCovered = "scope-not-covered"
	SecondReviewCauseNoSecondModel   = "no-second-model"
	SecondReviewCauseNotInHistory    = "not-in-history"
)

// SecondVerdictSection renders the second-review state (REQ-CLOSURE-014).
// The Markdown renderer derives the literal NOT PERFORMED / STALE / FAILED
// text from State, Cause, Verdict, and SupersedingCommit.
type SecondVerdictSection struct {
	State               string                    `json:"state"`
	Cause               string                    `json:"cause,omitempty"`
	Verdict             string                    `json:"verdict,omitempty"` // pass | fail
	Backends            []SecondReviewBackendView `json:"backends"`
	ContractSecondModel string                    `json:"contract_second_model"` // "not recorded" when the contract omits it
	SubstituteBackend   bool                      `json:"substitute_backend"`
	Warning             string                    `json:"warning,omitempty"` // advisory-policy warning line
	SupersedingCommit   string                    `json:"superseding_commit,omitempty"`
}

// PlanAuditBindingSection renders the plan-audit binding (REQ-CLOSURE-011).
type PlanAuditBindingSection struct {
	State    string `json:"state"` // bound | mismatch | self-reported
	File     string `json:"file,omitempty"`
	Verdict  string `json:"verdict,omitempty"` // the bound file's verdict line
	Expected string `json:"expected,omitempty"`
}

// Plan-audit binding states.
const (
	PlanAuditBound        = "bound"
	PlanAuditMismatch     = "mismatch"
	PlanAuditSelfReported = "self-reported"
)

// NotPerformed is one entry of the report's not-performed list
// (design.md §B): the closed token plus a detail string.
type NotPerformed struct {
	Item   string `json:"item"`
	Detail string `json:"detail"`
}

// ResidualRiskItem is one residual-risk entry the report derives from an
// observed state.
type ResidualRiskItem struct {
	Item   string `json:"item"`
	Detail string `json:"detail"`
}

// HumanVerdictSection shows the latest recorded verdict (REQ-CLOSURE-022).
type HumanVerdictSection struct {
	State        string `json:"state"` // current | stale | none recorded
	Verdict      string `json:"verdict,omitempty"`
	Operator     string `json:"operator,omitempty"` // "Name <email>"
	RecordedAt   string `json:"recorded_at,omitempty"`
	ReportSHA256 string `json:"report_sha256,omitempty"`
}

// Human-verdict states.
const (
	HumanVerdictCurrent      = "current"
	HumanVerdictStale        = "stale"
	HumanVerdictNoneRecorded = "none recorded"
)

// Not-performed tokens (design.md §B). The set is closed; IsNotPerformedToken
// rejects anything else so a builder bug cannot render an unnamed token.
const (
	NotPerformedProgressMissing             = "progress-missing"
	NotPerformedACNotReported               = "ac-not-reported"
	NotPerformedFirstVerdictMissing         = "first-verdict-missing"
	NotPerformedSecondReviewNotPerformed    = "second-review-not-performed"
	NotPerformedSecondReviewStale           = "second-review-stale"
	NotPerformedReceiptMissing              = "receipt-missing"
	NotPerformedReceiptFieldUnrecognized    = "receipt-field-unrecognized"
	NotPerformedPlanAuditSelfReported       = "plan-audit-self-reported"
	NotPerformedPlanAuditMismatch           = "plan-audit-mismatch"
	NotPerformedEscalationUnreadable        = "escalation-unreadable"
	NotPerformedEscalationDetectorNotArmed  = "escalation-detector-not-armed"
	NotPerformedInvariantNotObserved        = "invariant-not-observed"
	NotPerformedNewAPIComparisonUnavailable = "new-api-comparison-unavailable"
	NotPerformedHumanVerdictNone            = "human-verdict-none"
	NotPerformedHumanVerdictStale           = "human-verdict-stale"
)

// notPerformedTokens is the closed catalogue of design.md §B.
var notPerformedTokens = []string{
	NotPerformedProgressMissing,
	NotPerformedACNotReported,
	NotPerformedFirstVerdictMissing,
	NotPerformedSecondReviewNotPerformed,
	NotPerformedSecondReviewStale,
	NotPerformedReceiptMissing,
	NotPerformedReceiptFieldUnrecognized,
	NotPerformedPlanAuditSelfReported,
	NotPerformedPlanAuditMismatch,
	NotPerformedEscalationUnreadable,
	NotPerformedEscalationDetectorNotArmed,
	NotPerformedInvariantNotObserved,
	NotPerformedNewAPIComparisonUnavailable,
	NotPerformedHumanVerdictNone,
	NotPerformedHumanVerdictStale,
}

// NotPerformedTokens returns the closed not-performed token catalogue.
func NotPerformedTokens() []string {
	out := make([]string, len(notPerformedTokens))
	copy(out, notPerformedTokens)
	return out
}

// IsNotPerformedToken reports whether tok is in the closed catalogue.
func IsNotPerformedToken(tok string) bool {
	for _, t := range notPerformedTokens {
		if t == tok {
			return true
		}
	}
	return false
}

// AddNotPerformed appends one {item, detail} entry. Calling it with a token
// outside the catalogue panics: a builder bug must fail at build time, never
// render an unnamed token (design.md §B).
func (r *Report) AddNotPerformed(item, detail string) {
	if !IsNotPerformedToken(item) {
		panic("closure: unknown not-performed token " + item)
	}
	r.NotPerformed = append(r.NotPerformed, NotPerformed{Item: item, Detail: detail})
}

// AddResidualRisk appends one residual-risk entry.
func (r *Report) AddResidualRisk(item, detail string) {
	r.ResidualRisk = append(r.ResidualRisk, ResidualRiskItem{Item: item, Detail: detail})
}
