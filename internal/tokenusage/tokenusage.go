// Package tokenusage parses a Claude Code session transcript JSONL file and
// sums the per-turn token usage into a session-total figure.
//
// It is the measurement baseline for the Token-Economy Epic (SPEC-TOKEN-ACCOUNTING-001,
// milestone M1). Each line of a transcript is a JSON record; assistant-type
// records carry a message.usage object with the four accounted fields
// (input_tokens, output_tokens, cache_creation_input_tokens,
// cache_read_input_tokens). This package sums those fields across all assistant
// turns of a session and computes a cache-hit ratio.
//
// Distinct from the statusline snapshot: internal/statusline reads the CURRENT
// context-window occupancy from stdin (reset on /clear and compaction), which is
// NOT the cumulative billed token figure. This package reads the transcript
// message.usage and SUMS across turns — a different, cumulative source
// (see SPEC-TOKEN-ACCOUNTING-001 §A.2).
//
// Read-only contract: this package treats transcript files as read-only. It
// opens them O_RDONLY and never creates, modifies, or deletes anything under
// ~/.claude/projects/** (REQ-TA-004). Robustness: a malformed JSON line is
// skipped and an absent file surfaces a not-exist error; neither panics nor
// aborts a surrounding audit run (REQ-TA-013).
package tokenusage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Usage is the session-total token figure summed from a transcript's
// assistant-turn message.usage records (REQ-TA-002). TokensSpent is the
// arithmetic sum of the four component fields; CacheHitRatio is derived per
// CacheHitRatio (REQ-TA-003).
type Usage struct {
	TokensInput         int     `json:"tokens_input"`
	TokensOutput        int     `json:"tokens_output"`
	TokensCacheCreation int     `json:"tokens_cache_creation"`
	TokensCacheRead     int     `json:"tokens_cache_read"`
	TokensSpent         int     `json:"tokens_spent"`
	CacheHitRatio       float64 `json:"cache_hit_ratio"`
}

// transcriptRecord is the minimal projection of a transcript JSONL line needed
// for token accounting. Unknown fields (server_tool_use, service_tier,
// cache_creation, inference_geo, iterations, speed, ...) are ignored by
// encoding/json. Message.Usage is a pointer so a usage-absent record is
// distinguishable (nil) from an all-zero usage block.
type transcriptRecord struct {
	Type    string `json:"type"`
	Message struct {
		Usage *usageFields `json:"usage"`
	} `json:"message"`
}

// usageFields captures the four accounted per-turn token fields (REQ-TA-001).
type usageFields struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

// SumSession reads the transcript JSONL file at path and returns the
// session-total Usage, summing the four message.usage fields across every
// assistant-type record (REQ-TA-001/002) and finalizing TokensSpent and
// CacheHitRatio.
//
// Robustness (REQ-TA-013): a JSON-unparseable line is skipped and summation
// continues; a non-assistant record or an assistant record without a usage
// block contributes zero. The file is opened read-only (REQ-TA-004). An absent
// or unreadable file returns the zero Usage and a wrapped error (detectable via
// errors.Is(err, os.ErrNotExist)); the caller is expected to skip-and-continue
// rather than abort.
//
// @MX:ANCHOR: [AUTO] token-accounting extraction entry point — the read-only transcript-sum contract that the Token-Economy attribution/persistence/audit milestones (M2-M4) build on
// @MX:REASON: [AUTO] public API boundary + integration point: the four-field sum, the read-only invariant, and the cache-hit-ratio derivation are the invariants downstream milestones depend on; breaking any silently corrupts every per-SPEC token figure
// @MX:SPEC: SPEC-TOKEN-ACCOUNTING-001
func SumSession(path string) (Usage, error) {
	f, err := os.Open(path) // O_RDONLY — read-only contract (REQ-TA-004)
	if err != nil {
		return Usage{}, fmt.Errorf("tokenusage: open transcript %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	return sumReader(f)
}

// sumReader performs the line-by-line summation over r. It is separated from
// SumSession so the summation logic is exercisable without a filesystem path.
//
// It uses bufio.Reader.ReadString rather than bufio.Scanner because transcript
// assistant records routinely exceed bufio.Scanner's default 64KB token cap;
// Scanner would silently drop long lines.
func sumReader(r io.Reader) (Usage, error) {
	var u Usage
	reader := bufio.NewReader(r)
	for {
		line, readErr := reader.ReadString('\n')
		if s := strings.TrimSpace(line); s != "" {
			addLine(s, &u)
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			// Genuine read error (rare for local files): finalize what was
			// accumulated and surface the error so the caller can decide.
			u.finalize()
			return u, fmt.Errorf("tokenusage: read transcript: %w", readErr)
		}
	}
	u.finalize()
	return u, nil
}

// addLine parses a single trimmed JSONL line and, if it is an assistant record
// carrying a usage block, adds its four fields to u. A parse failure, a
// non-assistant type, or a missing usage block is a no-op (skip-and-continue,
// REQ-TA-013).
func addLine(line string, u *Usage) {
	var rec transcriptRecord
	if err := json.Unmarshal([]byte(line), &rec); err != nil {
		return // malformed line — skip
	}
	if rec.Type != "assistant" || rec.Message.Usage == nil {
		return // only assistant records carry accounted usage
	}
	usage := rec.Message.Usage
	u.TokensInput += usage.InputTokens
	u.TokensOutput += usage.OutputTokens
	u.TokensCacheCreation += usage.CacheCreationInputTokens
	u.TokensCacheRead += usage.CacheReadInputTokens
}

// finalize computes the derived fields from the summed components:
// TokensSpent (arithmetic sum of the four, REQ-TA-002) and CacheHitRatio
// (REQ-TA-003).
func (u *Usage) finalize() {
	u.TokensSpent = u.TokensInput + u.TokensOutput + u.TokensCacheCreation + u.TokensCacheRead
	u.CacheHitRatio = CacheHitRatio(u.TokensInput, u.TokensCacheCreation, u.TokensCacheRead)
}

// CacheHitRatio computes cache_read / (input + cache_creation + cache_read),
// bounded to [0, 1], returning 0 when the denominator is 0 (REQ-TA-003).
//
// @MX:NOTE: [AUTO] the ratio denominator excludes output_tokens by design (cache read/creation applies to the input side); the 0-denominator guard returns 0 rather than NaN so downstream JSON/audit surfaces never carry a non-finite value
func CacheHitRatio(inputTokens, cacheCreation, cacheRead int) float64 {
	denom := inputTokens + cacheCreation + cacheRead
	if denom <= 0 {
		return 0
	}
	ratio := float64(cacheRead) / float64(denom)
	if ratio < 0 {
		return 0
	}
	if ratio > 1 {
		return 1
	}
	return ratio
}
