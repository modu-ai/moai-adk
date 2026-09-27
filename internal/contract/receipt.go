package contract

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Kickoff receipt (REQ-CONTRACT-023; design.md § Kickoff Receipt). A1 owns the
// fields and a validator that checks structure and internal consistency
// only; the cross-check rules that derive the outcome are A3's and are not
// evaluated here.

// KickoffReceipt is the typed `.moai/specs/<SPEC-ID>/kickoff-receipt.json`.
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

// ReceiptFallback records a Jev-side fallback from llm+jev to llm. An absent
// fallback block reads as `applied: false` with no reason.
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

// ReceiptFileRef names a repo-relative file and the SHA-256 of its raw bytes.
type ReceiptFileRef struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// ReceiptAnswer is one decider's answer. Confidence is a pointer so an
// absent confidence is distinguishable from 0.
type ReceiptAnswer struct {
	Answer     string   `json:"answer"`
	Confidence *float64 `json:"confidence"`
	Reason     string   `json:"reason"`
	ReasonRefs []string `json:"reason_refs"`
}

// ReceiptJevAnswer is the Jev answer: an answer plus the request hash and the
// verbatim response body. The LLM answer is a separate type, so those two
// fields on llm_answer are unknown fields (strict decode rejects them).
type ReceiptJevAnswer struct {
	ReceiptAnswer
	RequestSHA256 string `json:"request_sha256"`
	RawResponse   string `json:"raw_response"`
}

// Receipt vocabulary.
var (
	receiptDeciders = []string{DeciderLLM, DeciderLLMJev}
	// fallbackReasons is the closed fallback-reason set (field rule 5).
	fallbackReasons = []string{"jev_disabled", "jev_low_confidence", "jev_malformed_response", "jev_call_failed", "jev_key_missing"}
	receiptAnswers  = []string{"approve", "reject", "escalate"}
	receiptOutcomes = []string{"approve", "reject", "human"}
	reasonRefRe     = regexp.MustCompile(`^contract\.yaml:([1-9][0-9]*)$`)
)

// ReceiptCheck carries everything ValidateKickoffReceipt needs. The caller
// supplies configuration and measurements; this package reads no file itself.
type ReceiptCheck struct {
	// Raw is the receipt file's bytes.
	Raw []byte
	// Path is the receipt path as given, relative to the project root
	// (the caller relativizes an absolute --receipt value). It must resolve
	// to `.moai/specs/<SpecID>/kickoff-receipt.json`.
	Path string
	// SpecID is the SPEC being signed.
	SpecID string
	// ContractLines is the draft contract.yaml's line count
	// (ContractLineCount); reason_refs must stay within it.
	ContractLines int
	// EffectiveDecider is the effective workflow.autonomy.kickoff.decider.
	EffectiveDecider string
	// Signer is the --signer value.
	Signer string
	// SignableContractSHA256 is Report.SignableContractSHA256 for the draft.
	SignableContractSHA256 string
	// AcceptanceSHA256 is the measured AcceptanceHash of acceptance.md.
	AcceptanceSHA256 string
	// ReadRepoFile reads a repo-relative file (the plan-audit report). A nil
	// callback or a read error is an input mismatch.
	ReadRepoFile func(rel string) ([]byte, error)
}

// DecodeKickoffReceipt strictly decodes a kickoff receipt: exactly one JSON
// object, no unknown field at any depth, no trailing data.
func DecodeKickoffReceipt(raw []byte) (*KickoffReceipt, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return nil, errors.New("contract: kickoff receipt is not a JSON object")
	}
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	dec.DisallowUnknownFields()
	var r KickoffReceipt
	if err := dec.Decode(&r); err != nil {
		return nil, fmt.Errorf("contract: kickoff receipt: %w", err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("contract: kickoff receipt: trailing data after the JSON object")
	}
	return &r, nil
}

// ValidateKickoffReceipt applies the A1 receipt field rules 1-9 of
// design.md § Kickoff Receipt in table order and returns the refusal code of
// the first failing rule (RefuseReceiptInvalid, RefuseReceiptSignerMismatch,
// or RefuseReceiptInputMismatch), or "" when the receipt is accepted. The
// decoded receipt is returned whenever decoding succeeded. Acceptance is not
// yet a decision to sign: the caller applies ReceiptOutcome next.
//
// @MX:ANCHOR: [AUTO] Structural and consistency gate for kickoff receipts.
// @MX:REASON: The receipt signing path (sign), the downstream receipt
// issuance (A3), and the gate rewiring all depend on its rule order and its
// refusal codes; a reordered rule changes which code an operator sees.
func ValidateKickoffReceipt(chk ReceiptCheck) (*KickoffReceipt, string) {
	// Rule 1 — strict decode, version, SPEC, fixed path (plus the
	// structural shape of the plan-audit report path).
	r, err := DecodeKickoffReceipt(chk.Raw)
	if err != nil {
		return nil, RefuseReceiptInvalid
	}
	if r.ReceiptVersion != 1 || !ValidSpecID(chk.SpecID) || r.SpecID != chk.SpecID ||
		!isFixedReceiptPath(chk.Path, chk.SpecID) {
		return r, RefuseReceiptInvalid
	}
	if in := r.Inputs; in != nil && in.PlanAuditReport != nil && !relativeClean(in.PlanAuditReport.Path) {
		return r, RefuseReceiptInvalid
	}

	// Rule 2 — decider vocabulary.
	req, eff := r.RequestedDecider, r.EffectiveDecider
	if !slices.Contains(receiptDeciders, req) || !slices.Contains(receiptDeciders, eff) {
		return r, RefuseReceiptInvalid
	}

	// Rule 3 — configuration and --signer.
	if req != chk.EffectiveDecider || eff != chk.Signer {
		return r, RefuseReceiptSignerMismatch
	}

	// Rule 4 — the fallback flag is set exactly when the deciders differ,
	// and the only permitted difference is llm+jev → llm.
	applied := r.Fallback != nil && r.Fallback.Applied
	differ := req != eff
	if applied != differ || (differ && (req != DeciderLLMJev || eff != DeciderLLM)) {
		return r, RefuseReceiptSignerMismatch
	}

	// Rule 5 — the reason is present exactly when applied, from the closed set.
	hasReason := r.Fallback != nil && r.Fallback.Reason != nil
	if hasReason != applied || (hasReason && !slices.Contains(fallbackReasons, *r.Fallback.Reason)) {
		return r, RefuseReceiptInvalid
	}

	// Rule 6 — llm_answer always; jev_answer exactly when Jev is effective.
	if r.LLMAnswer == nil || (r.JevAnswer != nil) != (eff == DeciderLLMJev) {
		return r, RefuseReceiptInvalid
	}

	// Rule 7 — each answer's fields.
	if !validAnswer(*r.LLMAnswer, chk.ContractLines) {
		return r, RefuseReceiptInvalid
	}
	if j := r.JevAnswer; j != nil && (!validAnswer(j.ReceiptAnswer, chk.ContractLines) ||
		strings.TrimSpace(j.RawResponse) == "" || !isHex64(j.RequestSHA256)) {
		return r, RefuseReceiptInvalid
	}

	// Rule 8 — input hashes equal the current values.
	if !inputsMatch(r.Inputs, chk) {
		return r, RefuseReceiptInputMismatch
	}

	// Rule 9 — outcome vocabulary; approve needs an LLM approve.
	if !slices.Contains(receiptOutcomes, r.Outcome) ||
		(r.Outcome == "approve" && r.LLMAnswer.Answer != "approve") {
		return r, RefuseReceiptInvalid
	}
	return r, ""
}

// ReceiptOutcome applies the signer steps that follow an accepted receipt
// (design.md § Kickoff Receipt): (1) the interim A1 rule — while A3's
// amendment has not landed, an effective decider of llm+jev (Jev actually
// answered) is refused with RefuseReceiptRequiresHuman even when both answers
// approve; (2) the recorded outcome — approve signs (ok), reject is
// RefuseReceiptRejected, human is RefuseReceiptRequiresHuman. A nil receipt
// or an outcome outside the set (the validator rejects both) is
// RefuseReceiptInvalid.
func ReceiptOutcome(r *KickoffReceipt) (refusal string, ok bool) {
	if r == nil {
		return RefuseReceiptInvalid, false
	}
	if r.EffectiveDecider == DeciderLLMJev {
		return RefuseReceiptRequiresHuman, false
	}
	switch r.Outcome {
	case "approve":
		return "", true
	case "reject":
		return RefuseReceiptRejected, false
	case "human":
		return RefuseReceiptRequiresHuman, false
	}
	return RefuseReceiptInvalid, false
}

// ContractLineCount returns the number of lines in raw: newline count, plus
// one for a final line without a trailing newline. It is the bound for a
// receipt's `contract.yaml:<line>` references.
func ContractLineCount(raw []byte) int {
	if len(raw) == 0 {
		return 0
	}
	n := bytes.Count(raw, []byte("\n"))
	if raw[len(raw)-1] != '\n' {
		n++
	}
	return n
}

// isFixedReceiptPath reports whether p resolves to the fixed receipt path.
func isFixedReceiptPath(p, specID string) bool {
	if p == "" {
		return false
	}
	return path.Clean(filepath.ToSlash(p)) == ".moai/specs/"+specID+"/"+ReceiptFile
}

func validAnswer(a ReceiptAnswer, lines int) bool {
	if !slices.Contains(receiptAnswers, a.Answer) || a.Confidence == nil ||
		*a.Confidence < 0 || *a.Confidence > 1 || strings.TrimSpace(a.Reason) == "" ||
		len(a.ReasonRefs) == 0 {
		return false
	}
	for _, ref := range a.ReasonRefs {
		m := reasonRefRe.FindStringSubmatch(ref)
		if m == nil {
			return false
		}
		line, err := strconv.Atoi(m[1])
		if err != nil || line > lines {
			return false
		}
	}
	return true
}

func inputsMatch(in *ReceiptInputs, chk ReceiptCheck) bool {
	if in == nil || in.PlanAuditReport == nil || chk.ReadRepoFile == nil {
		return false
	}
	if !isHex64(chk.SignableContractSHA256) || in.ContractSHA256 != chk.SignableContractSHA256 {
		return false
	}
	if !isHex64(chk.AcceptanceSHA256) || in.AcceptanceSHA256 != chk.AcceptanceSHA256 {
		return false
	}
	data, err := chk.ReadRepoFile(in.PlanAuditReport.Path)
	return err == nil && sumHex(data) == in.PlanAuditReport.SHA256
}
