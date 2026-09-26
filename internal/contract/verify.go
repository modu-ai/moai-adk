package contract

import "regexp"

// Verify states.
const (
	StateUnsigned      = "unsigned"
	StateSignedValid   = "signed-valid"
	StateSignedInvalid = "signed-invalid"
)

// Policy carries the workflow.autonomy values verify depends on. The caller
// (CLI) reads configuration; this package never imports internal/config.
type Policy struct {
	SecondReview string // required | advisory | off (anything else reads as required)
	PushDevelop  bool   // workflow.autonomy.contract.push_develop
	Mode         string // guided | contract (reported only)
	// BudgetDefault is workflow.autonomy.escalation.budget_default; it fills
	// an absent budget in the signable digest.
	BudgetDefault Budget
}

// Inputs is everything Verify needs, supplied by the caller. LoadDir fills
// SpecID, Contract, Acceptance(+Present), and Receipt(+Present); the caller
// adds Policy, the constitution registry values, and the spec.md status.
type Inputs struct {
	SpecID              string // containing SPEC directory name
	Contract            []byte // raw contract.yaml
	Acceptance          []byte // raw acceptance.md
	AcceptancePresent   bool
	Receipt             []byte // raw kickoff-receipt.json
	ReceiptPresent      bool
	Policy              Policy
	RegistryRuleIDs     []string // constitution registry rule IDs
	RegistryFrozenFiles []string // registry Frozen-zone file paths (literal, repo-relative)
	SpecStatus          string   // spec.md frontmatter status, raw (quotes allowed)
}

// Report is the verify result. JSON field names are the stable names of
// design.md § `show --json` object; slices marshal as [] and never null.
type Report struct {
	SpecID        string   `json:"spec_id"`
	Card          string   `json:"card"`
	SchemaVersion int      `json:"schema_version"`
	State         string   `json:"state"`
	Valid         bool     `json:"valid"`
	Reasons       []string `json:"reasons"`
	// ContractSHA256 is the digest recomputed from the file (empty when the
	// document failed to decode).
	ContractSHA256 string `json:"contract_sha256"`
	// RecordedContractSHA256 is signature.contract_sha256 (empty when unsigned).
	RecordedContractSHA256 string `json:"recorded_contract_sha256"`
	// SignableContractSHA256 is the digest `sign` would record now
	// (acceptance binding measured, budget filled). Empty when no signable
	// binding exists: acceptance section or file absent, or the count unusable.
	SignableContractSHA256 string `json:"signable_contract_sha256"`

	Acceptance        ReportAcceptance `json:"acceptance"`
	Actions           []string         `json:"actions"`
	PushRequiresLease bool             `json:"push_requires_lease"`
	Terminal          bool             `json:"terminal"`
	EffectiveNever    []string         `json:"effective_never"`
	Scratch           []string         `json:"scratch"`
	FrozenFiles       []string         `json:"frozen_files"`
	SecondReview      string           `json:"second_review"`
	Mode              string           `json:"mode"`
	Budget            *Budget          `json:"budget"`
	Signature         *SignatureView   `json:"signature"`

	// Contract is the decoded contract (nil when decoding failed).
	Contract *Contract `json:"-"`
}

// ReportAcceptance is the recorded and measured acceptance binding. The
// recorded values are null when the contract omits them; the measured count
// is 0 when acceptance.md is absent or its count is unusable.
type ReportAcceptance struct {
	SHA256          *string `json:"sha256"`
	MeasuredSHA256  string  `json:"measured_sha256"`
	ACCount         *int    `json:"ac_count"`
	MeasuredACCount int     `json:"measured_ac_count"`
}

// SignatureView is the signature as show --json reports it: every field
// present (receipt null when absent, batch_id/supersedes "" when absent),
// the digests and the seal left out.
type SignatureView struct {
	SignerKind string   `json:"signer_kind"`
	Operator   Operator `json:"operator"`
	SignedAt   string   `json:"signed_at"`
	HeadSHA    string   `json:"head_sha"`
	Method     string   `json:"method"`
	Receipt    *Receipt `json:"receipt"`
	BatchID    string   `json:"batch_id"`
	Supersedes string   `json:"supersedes"`
}

var (
	specIDRe = regexp.MustCompile(SpecIDPattern)
	cardRe   = regexp.MustCompile(CardPattern)
)

// ValidCard reports whether card matches CardPattern.
func ValidCard(card string) bool {
	return cardRe.MatchString(card)
}

// ValidSpecID reports whether id matches SpecIDPattern.
func ValidSpecID(id string) bool {
	return specIDRe.MatchString(id)
}

// newReport returns a report carrying the caller-derived values and empty
// (non-nil) slices, so an undecodable contract still marshals `[]`.
func newReport(in Inputs) Report {
	return Report{
		SpecID:         in.SpecID,
		Reasons:        []string{},
		Actions:        []string{},
		EffectiveNever: []string{},
		Scratch:        []string{},
		FrozenFiles:    []string{},
		Terminal:       isTerminal(in.SpecStatus),
		SecondReview:   in.Policy.SecondReview,
		Mode:           in.Policy.Mode,
	}
}

// Verify evaluates a contract against its inputs. It is pure: no file,
// process, or network access. It never returns an error; every problem is a
// reason code. See the package documentation for the rules implemented.
//
// @MX:ANCHOR: [AUTO] Single verification entry point for SPEC contracts.
// @MX:REASON: Called by `moai contract verify|show`, by `sign` before it
// writes, and by the downstream escalation hooks; its reason set and State
// semantics are a published contract (design.md § Verify Reason Codes).
func Verify(in Inputs) Report {
	r := newReport(in)

	c, err := Decode(in.Contract)
	if err != nil {
		// A decode failure stops evaluation with exactly [schema_invalid].
		r.State = StateUnsigned
		if hasTopLevelSignature(in.Contract) {
			r.State = StateSignedInvalid
		}
		r.Reasons = []string{ReasonSchemaInvalid}
		return r
	}
	r.Contract = c
	r.SchemaVersion = c.SchemaVersion
	r.Card = c.Card
	signed := c.Signature != nil

	reasons := reasonSet{}
	if c.SchemaVersion != SchemaVersion {
		reasons.add(ReasonSchemaInvalid)
	}
	if c.SpecID != in.SpecID || !ValidSpecID(c.SpecID) {
		reasons.add(ReasonSpecIDMismatch)
	}
	if !ValidCard(c.Card) {
		// Missing, empty, or not one path segment (design.md § Card Field).
		reasons.add(ReasonCardInvalid)
	}
	for _, s := range c.MissingSections() {
		// An unsigned draft may omit budget; sign fills it before signing.
		if s == SectionBudget && !signed {
			continue
		}
		reasons.add(ReasonSchemaInvalid)
	}
	checkFieldRules(c, in, reasons)

	measured := measureAcceptance(in)
	checkAcceptanceBinding(c, measured, signed, reasons)

	digest, err := Digest(c)
	if err != nil {
		reasons.add(ReasonSchemaInvalid)
	}
	r.ContractSHA256 = digest

	if !signed {
		reasons.add(ReasonUnsigned)
	} else {
		r.RecordedContractSHA256 = c.Signature.ContractSHA256
		if c.Signature.ContractSHA256 != digest {
			reasons.add(ReasonContractDigestMismatch)
		}
		checkSignature(c.Signature, measured, in, reasons)
	}

	fillDerived(&r, c, in, measured)

	r.Reasons = reasons.sorted()
	switch {
	case !signed:
		r.State = StateUnsigned
	case len(r.Reasons) == 0:
		r.State = StateSignedValid
	default:
		r.State = StateSignedInvalid
	}
	r.Valid = signed && len(r.Reasons) == 0
	return r
}
