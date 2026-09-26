// Package contract is the pure verification core for SPEC autonomy contracts
// (`.moai/specs/<SPEC-ID>/contract.yaml`, schema_version 1). The normative
// schema lives in the SPEC's design.md § Contract Schema.
//
// The package imports only the standard library and gopkg.in/yaml.v3. It
// never imports os/exec or net (directly or transitively), and it never
// imports internal/config, internal/constitution, internal/spec,
// internal/hook, or internal/mission: policy values, constitution registry
// rule IDs, and registry Frozen files are passed in by the caller through
// Inputs. That keeps Verify safe to call from hooks.
//
// # Public API
//
// Types:
//
//	Contract, Acceptance, Ownership, Review, Budget, PlanAudit  // contract body
//	Signature, Operator, Receipt                               // signature block
//	Policy, Inputs, Report                                     // verify I/O
//	ReportAcceptance, SignatureView                            // show --json sub-objects
//	ACCountResult                                              // AC counter outcome
//	KickoffReceipt, ReceiptFallback, ReceiptInputs,            // kickoff-receipt.json
//	  ReceiptFileRef, ReceiptAnswer, ReceiptJevAnswer
//	ReceiptCheck                                               // receipt validator inputs
//
// Functions:
//
//	Decode(raw []byte) (*Contract, error)          // strict YAML decode; errors wrap ErrSchemaInvalid
//	(*Contract).HasSection(name string) bool       // decode-time presence of a section
//	(*Contract).MissingSections() []string         // required sections absent
//	Canonical(c Contract) Contract                 // deep copy with set-valued lists sorted
//	Digest(c *Contract) (string, error)            // canonical body digest (signature excluded)
//	DigestBytes(raw []byte) (string, error)        // Decode + Digest
//	Verify(in Inputs) Report                       // pure evaluation; never errors
//	ValidSpecID(id string) bool                    // SpecIDPattern match
//	ValidCard(card string) bool                    // CardPattern match
//	ResolveSpecDir(projectRoot, specID string) (string, error)
//	LoadDir(specDir string) (Inputs, error)        // reads contract.yaml (+ acceptance.md, receipt)
//	ReasonCodes() []string / IsReasonCode(string) bool  // closed verify reason-code set
//	SignRefusalCodes() []string / IsSignRefusalCode(string) bool  // closed sign refusal-code set
//	NormalizeAcceptance(raw []byte) []byte         // strip one leading BOM, CRLF → LF
//	AcceptanceHash(raw []byte) string              // lowercase-hex SHA-256 of normalized bytes
//	CountAC(normalized []byte) (ACCountResult, error)  // Go port of the MOAI-AC-COUNTER awk program
//	MatchGlob(pattern, name string) bool           // doublestar ownership glob (design.md § Glob Semantics)
//	ComputeSeal(sig Signature) (string, error)     // signature seal (design.md § Signature Seal)
//	DecodeKickoffReceipt(raw []byte) (*KickoffReceipt, error)  // strict JSON decode
//	ValidateKickoffReceipt(chk ReceiptCheck) (*KickoffReceipt, string)  // field rules 1-9; "" = accepted
//	ReceiptOutcome(r *KickoffReceipt) (refusal string, ok bool)  // signer steps after validation
//	ContractLineCount(raw []byte) int              // bound for contract.yaml:<line> references
//
// Variables: FrozenInstructionFiles (the hook's frozen instruction basenames,
// pinned by a test in internal/hook).
//
// Constants: the 24 Reason* codes, the 20 Refuse* sign refusal codes, State*
// values, Section* names, ContractFile / AcceptanceFile / ReceiptFile,
// SchemaVersion, SpecIDPattern, CardPattern, Method* / ProvenanceFile,
// SignerHuman / DeciderLLM / DeciderLLMJev, ActionPushDevelop,
// SecondModelNone, ReviewClosureReport, InvariantFrozenFiles,
// InvariantConstitutionPrefix.
// Sentinel errors: ErrSchemaInvalid, ErrContractMissing,
// ErrPathEscapesSpecDir, ErrInvalidSpecID, ErrACPrefixInvalid. LoadDir errors are I/O errors
// (the CLI maps them to exit 2), never reasons.
//
// Policy fields: SecondReview (required | advisory | off; any other value
// reads as required), PushDevelop, Mode (reported only), BudgetDefault (fills
// an absent budget in the signable digest).
//
// # Verify semantics
//
//   - A document that fails strict decoding (unknown field at any depth,
//     duplicate key, wrong scalar type, multiple documents, empty or
//     non-mapping document) yields exactly reasons == ["schema_invalid"] and
//     stops evaluation. Its State is "signed-invalid" when a top-level
//     `signature` key is visible to a lenient parse, otherwise "unsigned".
//     Every other rule is collected; none stops evaluation.
//   - schema_version != 1 → schema_invalid.
//   - spec_id different from Inputs.SpecID, or failing SpecIDPattern →
//     spec_id_mismatch.
//   - card missing, empty, or not matching CardPattern → card_invalid (a
//     collected reason, not schema_invalid). Report.Card carries the value.
//   - A missing required section (RequiredSections; a key whose value is
//     YAML null counts as missing) → schema_invalid. The section's own field
//     rule still runs on the absent value, so e.g. a missing `actions` also
//     reports actions_empty.
//     Design choice for `budget`: an unsigned draft may omit it, because
//     `sign` fills it from workflow.autonomy.escalation.budget_default before
//     signing; a signed contract without `budget` is schema_invalid.
//   - Field rules (design.md § Field rules): actions (actions_empty,
//     forbidden_action, unknown_action, push_develop_disabled), escalate_on
//     (escalate_on_incomplete unless exactly the six triggers), ownership
//     (ownership_invalid: empty write, an absolute / drive / UNC /
//     backslash-rooted / `..` / malformed glob, an identical string in
//     never∩write or never∩scratch, or no write glob matching
//     `.moai/specs/<Inputs.SpecID>/contract.yaml`), invariants
//     (invariant_unresolved: empty, or a `constitution:<glob>` matching no
//     Inputs.RegistryRuleIDs entry), reobserve (reobserve_incomplete),
//     review (second_model outside codex|glm|none → schema_invalid; none or
//     empty under a required second review → second_review_missing; human
//     other than closure-report → schema_invalid), budget (budget_invalid),
//     plan_audit (plan_audit_not_passing), blank approach, acceptance.file
//     other than acceptance.md, or a recorded ac_count below 1 →
//     schema_invalid.
//   - Acceptance binding: acceptance.md absent → acceptance_missing (and no
//     other binding reason); recorded sha256 ≠ measured →
//     acceptance_hash_mismatch; recorded ac_count ≠ measured →
//     ac_count_mismatch; a measured count that is ambiguous, or a counter
//     error (uncompilable prefix), → ac_count_ambiguous. A recorded value
//     absent from a signed contract is a mismatch; absent from an unsigned
//     draft it reports nothing.
//   - No signature → reason "unsigned", State "unsigned".
//   - Signed: signature.contract_sha256 ≠ recomputed digest →
//     contract_digest_mismatch; seal ≠ ComputeSeal → signature_seal_mismatch;
//     the consistency table of design.md § Signature Seal →
//     signature_inconsistent; signature.acceptance_sha256 ≠ measured hash →
//     signature_acceptance_mismatch (skipped when acceptance.md is absent);
//     method receipt with no receipt block, receipt.path other than
//     kickoff-receipt.json, the file absent, or its raw SHA-256 ≠
//     receipt.sha256 → receipt_mismatch.
//
// Reasons are always sorted, de-duplicated, and non-nil (JSON `[]`). Valid
// is true only for a signed contract with no reasons.
//
// # Derived fields (Report)
//
// effective_never = ownership.never, plus the SPEC's contract.yaml and
// acceptance.md when signed; scratch = ownership.scratch; frozen_files = the
// sorted union of Inputs.RegistryFrozenFiles, `**/<basename>` for each
// FrozenInstructionFiles entry, and ownership.never, only when the invariants
// contain `frozen-files`; push_requires_lease = actions contain push-develop;
// terminal = Inputs.SpecStatus (trimmed, one pair of surrounding quotes
// removed) is completed or archived; signable_contract_sha256 = the digest
// with the measured acceptance binding written and an absent budget filled
// from Policy.BudgetDefault (empty when no usable measurement exists). Every
// slice is sorted, de-duplicated, and marshals as `[]`, never null.
//
// # Kickoff receipt
//
// ValidateKickoffReceipt checks structure and internal consistency only
// (design.md § Kickoff Receipt, field rules 1-9) and returns the refusal code
// of the first failing rule in table order. ReceiptOutcome then applies the
// interim A1 rule (effective decider llm+jev → receipt_requires_human) and
// the recorded outcome. The A3-owned cross-check rules are not evaluated.
//
// # Digest
//
// Digest drops the signature block, sorts every set-valued list
// (invariants, ownership.write/never/scratch, actions, escalate_on), keeps
// reobserve in author order, marshals the typed struct to JSON in its fixed
// field order, and returns the lowercase-hex SHA-256. YAML comments,
// indentation, key order, and set order therefore do not change it; any
// field value does. An absent ownership.scratch and an empty one digest the
// same; an absent optional value otherwise (budget, acceptance.sha256,
// acceptance.ac_count) is a different value from a present one.
package contract
