package runtime

// SPEC-V3R6-PROMPT-CACHE-001 M1 — cache_control injection at the SDK wrapper.
//
// This file implements the pure-Go transformation that attaches Anthropic
// Prompt Caching breakpoints to an outgoing API request payload. It is a
// threshold-agnostic, side-effect-free function over an in-memory payload
// representation, so it is fully unit-testable without an SDK or network call.
//
// REQ-PC-001: inject cache_control {ephemeral, ttl:"1h"} on the LAST system block.
// REQ-PC-002: inject cache_control {ephemeral, ttl:"5m"} after the SPEC body marker.
// REQ-PC-003: when the active backend is GLM, omit cache_control entirely.
// R4 (decoupled): below MinCacheableTokens, omit the session breakpoint
//   (self-contained fallback — no dependency on per-agent model routing).

// LLMMode identifies the active backend for cache_control eligibility.
type LLMMode string

const (
	// LLMModeClaude is the Anthropic Claude backend (cache_control supported).
	LLMModeClaude LLMMode = "claude"
	// LLMModeGLM is the Z.AI GLM backend (cache_control omitted — REQ-PC-003).
	LLMModeGLM LLMMode = "glm"

	// cacheControlTypeEphemeral is the only cache_control type Anthropic supports.
	cacheControlTypeEphemeral = "ephemeral"

	// ttlSessionDefault and ttlSpecDefault mirror the cache.yaml enum defaults.
	ttlOff = "off"

	// approxCharsPerToken is the rough Anthropic heuristic for token estimation
	// of plain text (len(text)/4). It is intentionally conservative.
	approxCharsPerToken = 4
)

// CacheStrategy is the subset of cache configuration consumed by the injector.
// It mirrors the cacheStrategy section of cache.yaml (see internal/config).
type CacheStrategy struct {
	// Enabled toggles cache_control injection on/off.
	Enabled bool
	// SessionTTL is the session-start breakpoint TTL: "1h" | "5m" | "off".
	SessionTTL string
	// SpecTTL is the SPEC-body breakpoint TTL: "5m" | "off".
	SpecTTL string
	// MinCacheableTokens is the threshold below which the session breakpoint is
	// omitted (R4 threshold-agnostic fallback). Default 2048.
	MinCacheableTokens int
}

// CacheControl is the Anthropic ephemeral cache breakpoint marker. A nil pointer
// means "no breakpoint on this element".
type CacheControl struct {
	// Type is always "ephemeral" for the supported caching mode.
	Type string `json:"type"`
	// TTL is the cache lifetime: "1h" or "5m".
	TTL string `json:"ttl"`
}

// SystemBlock is one element of the Anthropic system prompt array.
type SystemBlock struct {
	Type         string        `json:"type"`
	Text         string        `json:"text"`
	CacheControl *CacheControl `json:"cache_control,omitempty"`
}

// Message is one element of the Anthropic messages array.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	// SpecBodyMarker flags the message that bundles the SPEC body (spec.md +
	// acceptance.md + plan.md) injected on `/moai run SPEC-XXX`. The 5m
	// breakpoint attaches to this message (REQ-PC-002).
	SpecBodyMarker bool          `json:"-"`
	CacheControl   *CacheControl `json:"cache_control,omitempty"`
}

// RequestPayload is the in-memory representation of an outgoing Anthropic API
// request relevant to cache_control placement.
type RequestPayload struct {
	System   []SystemBlock `json:"system,omitempty"`
	Messages []Message     `json:"messages"`
}

// InjectCacheControl returns a payload with cache_control breakpoints applied
// per the active strategy and backend. The input payload is mutated in place and
// also returned for call-site convenience. A nil payload returns nil.
//
// Placement rules:
//   - Session breakpoint (SessionTTL): on the LAST system block, only when
//     enabled, backend is not GLM, SessionTTL != "off", AND the cacheable system
//     payload meets MinCacheableTokens (R4 fallback).
//   - SPEC body breakpoint (SpecTTL): on the message whose SpecBodyMarker is set,
//     only when enabled, backend is not GLM, AND SpecTTL != "off".
//
// @MX:ANCHOR: [AUTO] InjectCacheControl — sole cache_control placement entry point for the SDK wrapper
// @MX:REASON: fan_in >= 3 expected — cc.go SDK wrapper, /moai run SPEC bundle path, and AC-PC-003/004 integration tests all route through this single injection point; placement invariants (system LAST + SPEC-body marker) are contractual.
func InjectCacheControl(payload *RequestPayload, strat CacheStrategy, mode LLMMode) *RequestPayload {
	if payload == nil {
		return nil
	}
	// REQ-PC-003: GLM backend → no injection at all.
	if mode == LLMModeGLM {
		return payload
	}
	if !strat.Enabled {
		return payload
	}

	injectSessionBreakpoint(payload, strat)
	injectSpecBodyBreakpoint(payload, strat)
	return payload
}

// injectSessionBreakpoint applies the 1h/5m session breakpoint on the LAST
// system block, subject to the SessionTTL and MinCacheableTokens gates.
func injectSessionBreakpoint(payload *RequestPayload, strat CacheStrategy) {
	if strat.SessionTTL == ttlOff || strat.SessionTTL == "" {
		return
	}
	if len(payload.System) == 0 {
		return
	}
	// R4 threshold-agnostic fallback: below MinCacheableTokens, omit.
	if estimateSystemTokens(payload.System) < strat.MinCacheableTokens {
		return
	}
	last := len(payload.System) - 1
	payload.System[last].CacheControl = &CacheControl{
		Type: cacheControlTypeEphemeral,
		TTL:  strat.SessionTTL,
	}
}

// injectSpecBodyBreakpoint applies the 5m SPEC-body breakpoint on the message
// carrying the SpecBodyMarker, subject to the SpecTTL gate.
func injectSpecBodyBreakpoint(payload *RequestPayload, strat CacheStrategy) {
	if strat.SpecTTL == ttlOff || strat.SpecTTL == "" {
		return
	}
	for i := range payload.Messages {
		if payload.Messages[i].SpecBodyMarker {
			payload.Messages[i].CacheControl = &CacheControl{
				Type: cacheControlTypeEphemeral,
				TTL:  strat.SpecTTL,
			}
		}
	}
}

// estimateSystemTokens returns a rough token estimate for the system prompt
// array using the len(text)/approxCharsPerToken heuristic.
func estimateSystemTokens(blocks []SystemBlock) int {
	total := 0
	for _, b := range blocks {
		total += len(b.Text) / approxCharsPerToken
	}
	return total
}
