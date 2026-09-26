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
//	ResolveSpecDir(projectRoot, specID string) (string, error)
//	LoadDir(specDir string) (Inputs, error)        // reads contract.yaml (+ acceptance.md, receipt)
//	ReasonCodes() []string / IsReasonCode(string) bool  // closed reason-code set
//
// Constants: the 23 Reason* codes, State* values, Section* names,
// ContractFile / AcceptanceFile / ReceiptFile, SchemaVersion, SpecIDPattern.
// Sentinel errors: ErrSchemaInvalid, ErrContractMissing,
// ErrPathEscapesSpecDir, ErrInvalidSpecID. LoadDir errors are I/O errors
// (the CLI maps them to exit 2), never reasons.
//
// # Verify semantics implemented so far (milestone M1)
//
//   - A document that fails strict decoding (unknown field at any depth,
//     duplicate key, wrong scalar type, multiple documents, empty or
//     non-mapping document) yields exactly reasons == ["schema_invalid"] and
//     stops evaluation. Its State is "signed-invalid" when a top-level
//     `signature` key is visible to a lenient parse, otherwise "unsigned".
//   - schema_version != 1 → schema_invalid.
//   - spec_id different from Inputs.SpecID, or failing SpecIDPattern →
//     spec_id_mismatch.
//   - A missing required section (RequiredSections; a key whose value is
//     YAML null counts as missing) → schema_invalid.
//     Design choice for `budget`: an unsigned draft may omit it, because
//     `sign` fills it from workflow.autonomy.escalation.budget_default before
//     signing; a signed contract without `budget` is schema_invalid.
//   - No signature → reason "unsigned", State "unsigned".
//   - signature.contract_sha256 different from the recomputed digest →
//     contract_digest_mismatch.
//
// Reasons are always sorted, de-duplicated, and non-nil (JSON `[]`). Valid
// is true only for a signed contract with no reasons. Field rules, derived
// fields, acceptance binding, the signature seal, and receipt validation are
// added by later milestones on top of the same Verify entry point.
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
