package closure

import (
	"strings"
	"testing"
)

// fixtureReceipt returns a v0.5.1 kickoff receipt (AC-CLOSURE-010 fixture a).
func fixtureReceipt() string {
	return `{"receipt_version":1,"spec_id":"SPEC-EXAMPLE-001",` +
		`"requested_decider":"llm","effective_decider":"llm",` +
		`"fallback":{"applied":false,"reason":null},"issued_at":"2026-09-26T08:00:00Z",` +
		`"inputs":{"contract_sha256":"c01","acceptance_sha256":"a01","plan_audit_report":null},` +
		`"llm_answer":{"answer":"approve","confidence":0.82,"reason":"","reason_refs":[]},` +
		`"jev_answer":null,"outcome":"approve"}`
}

// TestDecodeKickoffReceiptLLM pins AC-CLOSURE-010 (a): both deciders, the
// llm answer with confidence, the fallback flag, and the outcome display.
func TestDecodeKickoffReceiptLLM(t *testing.T) {
	v, unknown, err := DecodeKickoffReceipt([]byte(fixtureReceipt()))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(unknown) != 0 {
		t.Fatalf("unknown fields = %v, want none", unknown)
	}
	if v.RequestedDecider != "llm" || v.EffectiveDecider != "llm" {
		t.Fatalf("deciders = %q/%q", v.RequestedDecider, v.EffectiveDecider)
	}
	if v.FallbackApplied {
		t.Fatalf("fallback.applied = true, want false")
	}
	if v.LLMAnswer != "approve" || v.LLMConfidence == nil || *v.LLMConfidence != 0.82 {
		t.Fatalf("llm answer = %q %v", v.LLMAnswer, v.LLMConfidence)
	}
	if v.JevAnswer != "" || v.JevConfidence != nil {
		t.Fatalf("jev answer = %q %v, want empty", v.JevAnswer, v.JevConfidence)
	}
	if v.Outcome != "approve" {
		t.Fatalf("outcome = %q", v.Outcome)
	}
}

// TestDecodeKickoffReceiptLLMJev pins AC-CLOSURE-010 (b): both answers shown
// when the effective decider is llm+jev.
func TestDecodeKickoffReceiptLLMJev(t *testing.T) {
	raw := strings.Replace(fixtureReceipt(), `"requested_decider":"llm"`, `"requested_decider":"llm+jev"`, 1)
	raw = strings.Replace(raw, `"effective_decider":"llm"`, `"effective_decider":"llm+jev"`, 1)
	raw = strings.Replace(raw, `"jev_answer":null`,
		`"jev_answer":{"answer":"escalate","confidence":0.41,"reason":"","reason_refs":[],"request_sha256":"r1","raw_response":"x"}`, 1)
	raw = strings.Replace(raw, `"outcome":"approve"`, `"outcome":"human"`, 1)
	v, unknown, err := DecodeKickoffReceipt([]byte(raw))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(unknown) != 0 {
		t.Fatalf("unknown fields = %v", unknown)
	}
	if v.EffectiveDecider != "llm+jev" || v.JevAnswer != "escalate" ||
		v.JevConfidence == nil || *v.JevConfidence != 0.41 || v.Outcome != "human" {
		t.Fatalf("view = %+v", v)
	}
}

// TestDecodeKickoffReceiptFallback pins AC-CLOSURE-010 (c): fallback applied
// with its reason, no jev_answer.
func TestDecodeKickoffReceiptFallback(t *testing.T) {
	raw := strings.Replace(fixtureReceipt(), `"requested_decider":"llm"`, `"requested_decider":"llm+jev"`, 1)
	raw = strings.Replace(raw, `"fallback":{"applied":false,"reason":null}`,
		`"fallback":{"applied":true,"reason":"jev_low_confidence"}`, 1)
	v, unknown, err := DecodeKickoffReceipt([]byte(raw))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(unknown) != 0 {
		t.Fatalf("unknown fields = %v", unknown)
	}
	if !v.FallbackApplied || v.FallbackReason != "jev_low_confidence" {
		t.Fatalf("fallback = %t/%q", v.FallbackApplied, v.FallbackReason)
	}
}

// TestDecodeKickoffReceiptUnknownField pins AC-CLOSURE-010 (d): an unknown
// top-level field is listed, not dropped.
func TestDecodeKickoffReceiptUnknownField(t *testing.T) {
	raw := strings.Replace(fixtureReceipt(), `{"receipt_version":1`,
		`{"extra_answer":{"answer":"approve"},"receipt_version":1`, 1)
	v, unknown, err := DecodeKickoffReceipt([]byte(raw))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(unknown) != 1 || unknown[0] != "extra_answer" {
		t.Fatalf("unknown = %v, want [extra_answer]", unknown)
	}
	// The known fields still render.
	if v.Outcome != "approve" || v.LLMAnswer != "approve" {
		t.Fatalf("known fields lost: %+v", v)
	}
}

// TestDecodeKickoffReceiptMalformed pins a malformed receipt errors.
func TestDecodeKickoffReceiptMalformed(t *testing.T) {
	if _, _, err := DecodeKickoffReceipt([]byte("{broken")); err == nil {
		t.Fatalf("expected error for malformed receipt, got nil")
	}
}
