package contract

import "errors"

// KickoffReceipt is the typed kickoff-receipt.json (design.md § Kickoff Receipt).
type KickoffReceipt struct {
	ReceiptVersion   int               `json:"receipt_version"`
	SpecID           string            `json:"spec_id"`
	RequestedDecider string            `json:"requested_decider"`
	EffectiveDecider string            `json:"effective_decider"`
	Fallback         *ReceiptFallback  `json:"fallback"`
	IssuedAt         string            `json:"issued_at"`
	Inputs           *ReceiptInputs    `json:"inputs"`
	LLMAnswer        *ReceiptAnswer    `json:"llm_answer"`
	JevAnswer        *ReceiptJevAnswer `json:"jev_answer"`
	Outcome          string            `json:"outcome"`
}

// ReceiptFallback records a Jev-side fallback from llm+jev to llm.
type ReceiptFallback struct {
	Applied bool    `json:"applied"`
	Reason  *string `json:"reason"`
}

// ReceiptInputs are the hashes of the inputs the decider saw.
type ReceiptInputs struct {
	ContractSHA256   string          `json:"contract_sha256"`
	AcceptanceSHA256 string          `json:"acceptance_sha256"`
	PlanAuditReport  *ReceiptFileRef `json:"plan_audit_report"`
}

// ReceiptFileRef names a repo-relative file and its raw SHA-256.
type ReceiptFileRef struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// ReceiptAnswer is one decider's answer.
type ReceiptAnswer struct {
	Answer     string   `json:"answer"`
	Confidence *float64 `json:"confidence"`
	Reason     string   `json:"reason"`
	ReasonRefs []string `json:"reason_refs"`
}

// ReceiptJevAnswer is the Jev answer.
type ReceiptJevAnswer struct {
	ReceiptAnswer
	RequestSHA256 string `json:"request_sha256"`
	RawResponse   string `json:"raw_response"`
}

// ReceiptCheck carries everything ValidateKickoffReceipt needs.
type ReceiptCheck struct {
	Raw                    []byte
	Path                   string
	SpecID                 string
	ContractLines          int
	EffectiveDecider       string
	Signer                 string
	SignableContractSHA256 string
	AcceptanceSHA256       string
	ReadRepoFile           func(rel string) ([]byte, error)
}

// DecodeKickoffReceipt strictly decodes a kickoff receipt.
func DecodeKickoffReceipt(raw []byte) (*KickoffReceipt, error) {
	return nil, errors.New("M3 RED stub")
}

// ValidateKickoffReceipt applies the A1 receipt field rules.
func ValidateKickoffReceipt(chk ReceiptCheck) (*KickoffReceipt, string) {
	return nil, "" // M3 RED stub
}

// ReceiptOutcome applies the post-validation signer steps.
func ReceiptOutcome(r *KickoffReceipt) (string, bool) {
	return "", false // M3 RED stub
}

// ContractLineCount returns the number of lines in raw.
func ContractLineCount(raw []byte) int {
	return 0 // M3 RED stub
}
