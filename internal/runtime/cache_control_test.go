package runtime

import (
	"testing"
)

// SPEC-V3R6-PROMPT-CACHE-001 M1 — cache_control injection at SDK wrapper.
//
// These tests define the contract for InjectCacheControl, which transforms an
// outgoing Anthropic API request payload by attaching ephemeral cache_control
// breakpoints (REQ-PC-001 1h session start, REQ-PC-002 5m SPEC body,
// REQ-PC-003 GLM omit, R4 threshold-agnostic fallback).

// helperSystemBlock returns a single system prompt block of approximately the
// requested token count. The token estimate used by InjectCacheControl is
// len(text)/4 (Anthropic rough heuristic), so we size the text accordingly.
func helperSystemBlock(approxTokens int) SystemBlock {
	return SystemBlock{Type: "text", Text: makeText(approxTokens * 4)}
}

func makeText(chars int) string {
	if chars <= 0 {
		return "x"
	}
	b := make([]byte, chars)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

// AC-PC-003: cache_control schema integration test — session start 1h.
//
// TestCacheControl_SessionStart_OneHour verifies that with caching enabled and
// a non-GLM backend, InjectCacheControl attaches cache_control on the LAST
// system block with type "ephemeral" and ttl "1h".
func TestCacheControl_SessionStart_OneHour(t *testing.T) {
	t.Parallel()

	payload := &RequestPayload{
		System: []SystemBlock{
			helperSystemBlock(1000), // CLAUDE.md-ish
			helperSystemBlock(1000), // always-loaded rules
			helperSystemBlock(1000), // output style (LAST item)
		},
		Messages: []Message{
			{Role: "user", Content: "hello"},
		},
	}

	strat := CacheStrategy{
		Enabled:            true,
		SessionTTL:         "1h",
		SpecTTL:            "5m",
		MinCacheableTokens: 2048,
	}

	result := InjectCacheControl(payload, strat, LLMModeClaude)

	last := result.System[len(result.System)-1]
	if last.CacheControl == nil {
		t.Fatalf("expected cache_control on LAST system block, got nil")
	}
	if last.CacheControl.Type != "ephemeral" {
		t.Errorf("LAST system block cache_control.type: want %q, got %q", "ephemeral", last.CacheControl.Type)
	}
	if last.CacheControl.TTL != "1h" {
		t.Errorf("LAST system block cache_control.ttl: want %q, got %q", "1h", last.CacheControl.TTL)
	}

	// Non-last system blocks must NOT carry cache_control (single session breakpoint).
	for i := 0; i < len(result.System)-1; i++ {
		if result.System[i].CacheControl != nil {
			t.Errorf("system block %d unexpectedly carries cache_control (only LAST should)", i)
		}
	}
}

// AC-PC-003: cache_control schema integration test — full payload schema.
//
// TestCacheControl_AnthropicPayloadSchema verifies the complete Anthropic
// payload schema: a 1h breakpoint on the LAST system block AND a 5m breakpoint
// on the message that follows the SPEC body marker.
func TestCacheControl_AnthropicPayloadSchema(t *testing.T) {
	t.Parallel()

	payload := &RequestPayload{
		System: []SystemBlock{
			helperSystemBlock(1500),
			helperSystemBlock(1500),
		},
		Messages: []Message{
			{Role: "user", Content: "context preamble"},
			{Role: "user", Content: "<spec-body>SPEC-XXX bundle: spec + acceptance + plan</spec-body>", SpecBodyMarker: true},
			{Role: "assistant", Content: "ack"},
		},
	}

	strat := CacheStrategy{
		Enabled:            true,
		SessionTTL:         "1h",
		SpecTTL:            "5m",
		MinCacheableTokens: 2048,
	}

	result := InjectCacheControl(payload, strat, LLMModeClaude)

	// System LAST item: 1h ephemeral.
	last := result.System[len(result.System)-1]
	if last.CacheControl == nil || last.CacheControl.Type != "ephemeral" || last.CacheControl.TTL != "1h" {
		t.Fatalf("system LAST item must carry ephemeral/1h cache_control, got %+v", last.CacheControl)
	}

	// The message carrying the SPEC body marker must get the 5m breakpoint.
	var specMsg *Message
	for i := range result.Messages {
		if result.Messages[i].SpecBodyMarker {
			specMsg = &result.Messages[i]
			break
		}
	}
	if specMsg == nil {
		t.Fatal("expected a message with SpecBodyMarker == true")
	}
	if specMsg.CacheControl == nil {
		t.Fatal("SPEC body message must carry cache_control")
	}
	if specMsg.CacheControl.Type != "ephemeral" || specMsg.CacheControl.TTL != "5m" {
		t.Errorf("SPEC body cache_control: want ephemeral/5m, got %s/%s", specMsg.CacheControl.Type, specMsg.CacheControl.TTL)
	}

	// Messages without the SPEC body marker must NOT carry cache_control.
	for i := range result.Messages {
		if !result.Messages[i].SpecBodyMarker && result.Messages[i].CacheControl != nil {
			t.Errorf("non-SPEC-body message %d unexpectedly carries cache_control", i)
		}
	}
}

// AC-PC-004: GLM backend → cache_control omit entirely (REQ-PC-003).
//
// TestCacheControl_GLMMode_NoInjection verifies that when the active backend is
// GLM, InjectCacheControl performs NO injection at all, regardless of strategy.
func TestCacheControl_GLMMode_NoInjection(t *testing.T) {
	t.Parallel()

	payload := &RequestPayload{
		System: []SystemBlock{
			helperSystemBlock(3000),
		},
		Messages: []Message{
			{Role: "user", Content: "<spec-body>x</spec-body>", SpecBodyMarker: true},
		},
	}

	strat := CacheStrategy{
		Enabled:            true,
		SessionTTL:         "1h",
		SpecTTL:            "5m",
		MinCacheableTokens: 2048,
	}

	result := InjectCacheControl(payload, strat, LLMModeGLM)

	for i, blk := range result.System {
		if blk.CacheControl != nil {
			t.Errorf("GLM mode: system block %d must not carry cache_control", i)
		}
	}
	for i, msg := range result.Messages {
		if msg.CacheControl != nil {
			t.Errorf("GLM mode: message %d must not carry cache_control", i)
		}
	}
}

// TestCacheControl_BelowThreshold_NoInjection verifies the R4 threshold-agnostic
// fallback: when the cacheable system payload is below MinCacheableTokens, NO
// session cache_control is injected.
func TestCacheControl_BelowThreshold_NoInjection(t *testing.T) {
	t.Parallel()

	payload := &RequestPayload{
		System: []SystemBlock{
			helperSystemBlock(100), // ~100 tokens, well under 2048
		},
		Messages: []Message{
			{Role: "user", Content: "hi"},
		},
	}

	strat := CacheStrategy{
		Enabled:            true,
		SessionTTL:         "1h",
		SpecTTL:            "5m",
		MinCacheableTokens: 2048,
	}

	result := InjectCacheControl(payload, strat, LLMModeClaude)

	last := result.System[len(result.System)-1]
	if last.CacheControl != nil {
		t.Errorf("below MinCacheableTokens: session cache_control must be omitted, got %+v", last.CacheControl)
	}
}

// TestCacheControl_Disabled_NoInjection verifies that when caching is disabled,
// no injection occurs even for a non-GLM backend above threshold.
func TestCacheControl_Disabled_NoInjection(t *testing.T) {
	t.Parallel()

	payload := &RequestPayload{
		System:   []SystemBlock{helperSystemBlock(5000)},
		Messages: []Message{{Role: "user", Content: "<spec-body>x</spec-body>", SpecBodyMarker: true}},
	}

	strat := CacheStrategy{
		Enabled:            false,
		SessionTTL:         "1h",
		SpecTTL:            "5m",
		MinCacheableTokens: 2048,
	}

	result := InjectCacheControl(payload, strat, LLMModeClaude)

	if result.System[len(result.System)-1].CacheControl != nil {
		t.Error("disabled: session cache_control must be omitted")
	}
	if result.Messages[0].CacheControl != nil {
		t.Error("disabled: SPEC body cache_control must be omitted")
	}
}

// TestCacheControl_SessionTTLOff_OmitsSessionBreakpoint verifies REQ-PC-005:
// when SessionTTL == "off", no session-level breakpoint is injected, but the
// SPEC body breakpoint (spec_ttl) is still applied.
func TestCacheControl_SessionTTLOff_OmitsSessionBreakpoint(t *testing.T) {
	t.Parallel()

	payload := &RequestPayload{
		System: []SystemBlock{helperSystemBlock(5000)},
		Messages: []Message{
			{Role: "user", Content: "<spec-body>x</spec-body>", SpecBodyMarker: true},
		},
	}

	strat := CacheStrategy{
		Enabled:            true,
		SessionTTL:         "off",
		SpecTTL:            "5m",
		MinCacheableTokens: 2048,
	}

	result := InjectCacheControl(payload, strat, LLMModeClaude)

	if result.System[len(result.System)-1].CacheControl != nil {
		t.Error("session_ttl=off: session cache_control must be omitted")
	}
	if result.Messages[0].CacheControl == nil {
		t.Error("session_ttl=off: SPEC body breakpoint must still be applied")
	}
}

// TestCacheControl_SpecTTLOff_OmitsSpecBreakpoint verifies that when
// SpecTTL == "off", the SPEC body breakpoint is omitted but the session
// breakpoint is still applied.
func TestCacheControl_SpecTTLOff_OmitsSpecBreakpoint(t *testing.T) {
	t.Parallel()

	payload := &RequestPayload{
		System: []SystemBlock{helperSystemBlock(5000)},
		Messages: []Message{
			{Role: "user", Content: "<spec-body>x</spec-body>", SpecBodyMarker: true},
		},
	}

	strat := CacheStrategy{
		Enabled:            true,
		SessionTTL:         "1h",
		SpecTTL:            "off",
		MinCacheableTokens: 2048,
	}

	result := InjectCacheControl(payload, strat, LLMModeClaude)

	if result.System[len(result.System)-1].CacheControl == nil {
		t.Error("spec_ttl=off: session breakpoint must still be applied")
	}
	if result.Messages[0].CacheControl != nil {
		t.Error("spec_ttl=off: SPEC body cache_control must be omitted")
	}
}

// TestCacheControl_NilPayload_Safe verifies defensive handling of a nil payload.
func TestCacheControl_NilPayload_Safe(t *testing.T) {
	t.Parallel()

	if got := InjectCacheControl(nil, CacheStrategy{Enabled: true}, LLMModeClaude); got != nil {
		t.Errorf("nil payload must return nil, got %+v", got)
	}
}

// TestCacheControl_EmptySystem_NoPanic verifies that an empty system array does
// not panic and produces no session breakpoint.
func TestCacheControl_EmptySystem_NoPanic(t *testing.T) {
	t.Parallel()

	payload := &RequestPayload{
		System:   nil,
		Messages: []Message{{Role: "user", Content: "<spec-body>x</spec-body>", SpecBodyMarker: true}},
	}
	strat := CacheStrategy{Enabled: true, SessionTTL: "1h", SpecTTL: "5m", MinCacheableTokens: 2048}

	result := InjectCacheControl(payload, strat, LLMModeClaude)
	if result == nil {
		t.Fatal("expected non-nil result for empty system")
	}
	// SPEC body breakpoint still applies independent of system blocks.
	if result.Messages[0].CacheControl == nil {
		t.Error("SPEC body breakpoint should apply even with empty system array")
	}
}
