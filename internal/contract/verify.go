package contract

// Verify states.
const (
	StateUnsigned      = "unsigned"
	StateSignedValid   = "signed-valid"
	StateSignedInvalid = "signed-invalid"
)

// Policy carries the workflow.autonomy values verify depends on. The caller
// (CLI) reads configuration; this package never imports internal/config.
type Policy struct {
	SecondReview string
	PushDevelop  bool
	Mode         string
}

// Inputs is everything Verify needs, supplied by the caller.
type Inputs struct {
	SpecID              string
	Contract            []byte
	Acceptance          []byte
	AcceptancePresent   bool
	Receipt             []byte
	ReceiptPresent      bool
	Policy              Policy
	RegistryRuleIDs     []string
	RegistryFrozenFiles []string
	SpecStatus          string
}

// Report is the verify result.
type Report struct {
	SpecID                 string    `json:"spec_id"`
	SchemaVersion          int       `json:"schema_version"`
	State                  string    `json:"state"`
	Valid                  bool      `json:"valid"`
	Reasons                []string  `json:"reasons"`
	ContractSHA256         string    `json:"contract_sha256"`
	RecordedContractSHA256 string    `json:"recorded_contract_sha256"`
	SignableContractSHA256 string    `json:"signable_contract_sha256"`
	Contract               *Contract `json:"-"`
}

// ValidSpecID reports whether id matches SpecIDPattern.
func ValidSpecID(id string) bool {
	return false
}

// Verify evaluates a contract against its inputs.
func Verify(in Inputs) Report {
	return Report{}
}
