package contract

// Contract is the typed form of `.moai/specs/<SPEC-ID>/contract.yaml`
// (schema_version 1). Field order is load-bearing: the canonical digest
// marshals this struct to JSON in declaration order, so reordering fields
// changes every digest ever recorded. Add new fields only through a schema
// amendment (design.md § Contract Schema).
//
// Fields a draft may omit are pointers (Acceptance.SHA256, Acceptance.ACCount,
// Budget, Signature) or are tracked through the decode-time presence set
// (Ownership.Scratch), so "absent" is distinguishable from a zero value.
type Contract struct {
	SchemaVersion int         `yaml:"schema_version" json:"schema_version"`
	SpecID        string      `yaml:"spec_id" json:"spec_id"`
	Acceptance    *Acceptance `yaml:"acceptance" json:"acceptance"`
	Invariants    []string    `yaml:"invariants" json:"invariants"`
	Ownership     *Ownership  `yaml:"ownership" json:"ownership"`
	Approach      string      `yaml:"approach" json:"approach"`
	Actions       []string    `yaml:"actions" json:"actions"`
	Reobserve     []string    `yaml:"reobserve" json:"reobserve"`
	Review        *Review     `yaml:"review" json:"review"`
	Budget        *Budget     `yaml:"budget,omitempty" json:"budget"`
	EscalateOn    []string    `yaml:"escalate_on" json:"escalate_on"`
	PlanAudit     *PlanAudit  `yaml:"plan_audit" json:"plan_audit"`
	// Signature is written only by `moai contract sign` and is excluded from
	// the digest (Digest drops it before marshaling).
	Signature *Signature `yaml:"signature,omitempty" json:"signature,omitempty"`

	// present records which sections the decoded document carried with a
	// non-null value. Nil for a Contract built in Go rather than decoded;
	// HasSection falls back to a zero-value check in that case.
	present map[string]bool
}

// Acceptance binds the contract to the SPEC's acceptance.md.
type Acceptance struct {
	// File is fixed to "acceptance.md" in schema_version 1.
	File string `yaml:"file" json:"file"`
	// SHA256 is written by sign: SHA-256 of the LF-normalized, BOM-stripped
	// acceptance.md. Absent (nil) in an unsigned draft.
	SHA256 *string `yaml:"sha256,omitempty" json:"sha256"`
	// ACCount is written by sign: the published AC counter's live count.
	// Absent (nil) in an unsigned draft.
	ACCount *int `yaml:"ac_count,omitempty" json:"ac_count"`
}

// Ownership declares where the run may write.
type Ownership struct {
	Write []string `yaml:"write" json:"write"`
	Never []string `yaml:"never" json:"never"`
	// Scratch is optional; absent and empty are equivalent (both omitted
	// from the canonical form).
	Scratch []string `yaml:"scratch,omitempty" json:"scratch,omitempty"`
}

// Review names the second-model reviewer and the human review form.
type Review struct {
	SecondModel string `yaml:"second_model" json:"second_model"`
	Human       string `yaml:"human" json:"human"`
}

// Budget bounds the autonomous run. AuditRetries is the shared retry cap for
// plan-audit and sync-audit.
type Budget struct {
	Turns        int `yaml:"turns" json:"turns"`
	Operations   int `yaml:"operations" json:"operations"`
	AuditRetries int `yaml:"audit_retries" json:"audit_retries"`
}

// PlanAudit carries the self-reported plan-audit verdict.
type PlanAudit struct {
	Verdict string `yaml:"verdict" json:"verdict"`
}

// Signature is the block written by `moai contract sign`. Optional fields
// carry omitempty so the seal's canonical JSON omits them when absent
// (design.md § Signature Seal).
type Signature struct {
	SignerKind       string   `yaml:"signer_kind" json:"signer_kind"`
	Operator         Operator `yaml:"operator" json:"operator"`
	SignedAt         string   `yaml:"signed_at" json:"signed_at"`
	HeadSHA          string   `yaml:"head_sha" json:"head_sha"`
	ContractSHA256   string   `yaml:"contract_sha256" json:"contract_sha256"`
	AcceptanceSHA256 string   `yaml:"acceptance_sha256" json:"acceptance_sha256"`
	Method           string   `yaml:"method" json:"method"`
	Receipt          *Receipt `yaml:"receipt,omitempty" json:"receipt,omitempty"`
	BatchID          string   `yaml:"batch_id,omitempty" json:"batch_id,omitempty"`
	Supersedes       string   `yaml:"supersedes,omitempty" json:"supersedes,omitempty"`
	Seal             string   `yaml:"seal,omitempty" json:"seal,omitempty"`
}

// Operator is the signer's git identity.
type Operator struct {
	Name  string `yaml:"name" json:"name"`
	Email string `yaml:"email" json:"email"`
}

// Receipt references the kickoff receipt on the receipt signing path.
type Receipt struct {
	Path       string `yaml:"path" json:"path"`
	SHA256     string `yaml:"sha256" json:"sha256"`
	Provenance string `yaml:"provenance" json:"provenance"`
}

// Section names as they appear at the top level of contract.yaml.
const (
	SectionAcceptance = "acceptance"
	SectionInvariants = "invariants"
	SectionOwnership  = "ownership"
	SectionApproach   = "approach"
	SectionActions    = "actions"
	SectionReobserve  = "reobserve"
	SectionReview     = "review"
	SectionBudget     = "budget"
	SectionEscalateOn = "escalate_on"
	SectionPlanAudit  = "plan_audit"
	SectionSignature  = "signature"

	// Nested ownership keys, addressed as "ownership.<key>" by HasSection.
	SectionOwnershipWrite   = "ownership.write"
	SectionOwnershipNever   = "ownership.never"
	SectionOwnershipScratch = "ownership.scratch"
)

// RequiredSections lists, in schema order, the sections every contract must
// carry (REQ-CONTRACT-002). SectionBudget is required only once the contract
// is signed; an unsigned draft may omit it because sign fills it from
// workflow.autonomy.escalation.budget_default.
var RequiredSections = []string{
	SectionAcceptance,
	SectionInvariants,
	SectionOwnership,
	SectionOwnershipWrite,
	SectionOwnershipNever,
	SectionApproach,
	SectionActions,
	SectionReobserve,
	SectionReview,
	SectionBudget,
	SectionEscalateOn,
	SectionPlanAudit,
}

// File names inside a SPEC directory.
const (
	ContractFile   = "contract.yaml"
	AcceptanceFile = "acceptance.md"
	ReceiptFile    = "kickoff-receipt.json"
)

// SchemaVersion is the only schema_version this package accepts.
const SchemaVersion = 1

// SpecIDPattern is the canonical SPEC ID pattern. It is byte-identical to
// the spec-lint frontmatter pattern (internal/spec/lint.go specIDPattern);
// a test in internal/spec pins the two together so they cannot drift.
const SpecIDPattern = `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`
