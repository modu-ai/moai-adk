package closure

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/modu-ai/moai-adk/internal/contract"
)

// receiptKnownFields is the v0.5.1 receipt field set the display decoder
// knows (REQ-CLOSURE-010). Any other top-level field is listed under Not
// Performed (receipt-field-unrecognized), never dropped and never an error.
var receiptKnownFields = map[string]bool{
	"receipt_version":   true,
	"spec_id":           true,
	"requested_decider": true,
	"effective_decider": true,
	"fallback":          true,
	"issued_at":         true,
	"inputs":            true,
	"llm_answer":        true,
	"jev_answer":        true,
	"outcome":           true,
}

// ReceiptView is the display projection of a kickoff receipt: the fields
// REQ-CLOSURE-010 shows. Decider tokens are displayed verbatim; the display
// hard-codes no set. Receipt content is display-only — no readiness decision
// may read it.
type ReceiptView struct {
	RequestedDecider string
	EffectiveDecider string
	FallbackApplied  bool
	FallbackReason   string
	LLMAnswer        string
	LLMConfidence    *float64
	JevAnswer        string
	JevConfidence    *float64
	Outcome          string
}

// DecodeKickoffReceipt decodes a kickoff receipt for display. It returns the
// view, the sorted list of unknown top-level field names, and an error only
// when the file is not a JSON object (a malformed receipt).
//
// The typed fields come from A1's contract.KickoffReceipt — the schema is
// A1's, and this package re-implements none of it (plan.md §G).
func DecodeKickoffReceipt(raw []byte) (ReceiptView, []string, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return ReceiptView{}, nil, fmt.Errorf("closure: decode kickoff receipt: %w", err)
	}
	var unknown []string
	for name := range fields {
		if !receiptKnownFields[name] {
			unknown = append(unknown, name)
		}
	}
	sort.Strings(unknown)

	var kr contract.KickoffReceipt
	if err := json.Unmarshal(raw, &kr); err != nil {
		return ReceiptView{}, unknown, fmt.Errorf("closure: decode kickoff receipt (A1 shape): %w", err)
	}
	v := ReceiptView{
		RequestedDecider: kr.RequestedDecider,
		EffectiveDecider: kr.EffectiveDecider,
		Outcome:          kr.Outcome,
	}
	if kr.Fallback != nil {
		v.FallbackApplied = kr.Fallback.Applied
		if kr.Fallback.Reason != nil {
			v.FallbackReason = *kr.Fallback.Reason
		}
	}
	if kr.LLMAnswer != nil {
		v.LLMAnswer = kr.LLMAnswer.Answer
		v.LLMConfidence = kr.LLMAnswer.Confidence
	}
	if kr.JevAnswer != nil {
		v.JevAnswer = kr.JevAnswer.Answer
		v.JevConfidence = kr.JevAnswer.Confidence
	}
	return v, unknown, nil
}
