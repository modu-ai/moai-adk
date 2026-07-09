// Package cli — classifier eligibility regression tests.
// SPEC-HARNESS-RATCHET-REWIRE-001 REQ-HRR-003 / REQ-HRR-004 (AC-HRR-003 /
// AC-HRR-004): the tier classifier must NOT promote degenerate lifecycle-noise
// pattern keys (empty context hash AND empty/unknown subject). This file seeds
// a usage-log with both degenerate and control patterns and asserts that
// classifyHarnessPatterns excludes the degenerate ones while still promoting
// the eligible control.
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedEventLines writes a usage-log.jsonl under dir/.moai/harness/ containing
// one JSONL line per supplied event. Used to construct controlled aggregation
// fixtures for the eligibility regression.
func seedEventLines(t *testing.T, dir string, events []map[string]string) {
	t.Helper()
	logDir := filepath.Join(dir, ".moai", "harness")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("mkdir usage-log dir: %v", err)
	}
	var b strings.Builder
	for _, e := range events {
		// Build a minimal valid usage-log line. defaultConfidence is 1.0 and
		// the aggregator derives Count from line multiplicity.
		line := map[string]any{
			"timestamp":      "2026-07-09T10:00:00Z",
			"event_type":     e["event_type"],
			"subject":        e["subject"],
			"context_hash":   e["context_hash"],
			"tier_increment": 0,
			"schema_version": "v2.1",
		}
		raw, err := json.Marshal(line)
		if err != nil {
			t.Fatalf("marshal seed line: %v", err)
		}
		b.Write(raw)
		b.WriteByte('\n')
	}
	logPath := filepath.Join(logDir, "usage-log.jsonl")
	if err := os.WriteFile(logPath, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write usage-log: %v", err)
	}
}

// readPromotionKeys reads tier-promotions.jsonl and returns the set of
// pattern_key values that were promoted.
func readPromotionKeys(t *testing.T, dir string) map[string]bool {
	t.Helper()
	promoPath := filepath.Join(dir, ".moai", "harness", "learning-history", "tier-promotions.jsonl")
	data, err := os.ReadFile(promoPath)
	if err != nil {
		t.Fatalf("read tier-promotions.jsonl: %v", err)
	}
	keys := make(map[string]bool)
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var p map[string]any
		if json.Unmarshal([]byte(line), &p) == nil {
			if k, ok := p["pattern_key"].(string); ok {
				keys[k] = true
			}
		}
	}
	return keys
}

// repeatEvents returns n copies of the given event map, simulating n
// observations of the same pattern.
func repeatEvents(e map[string]string, n int) []map[string]string {
	out := make([]map[string]string, n)
	for i := range out {
		cp := make(map[string]string, len(e))
		for k, v := range e {
			cp[k] = v
		}
		out[i] = cp
	}
	return out
}

// TestClassifyHarnessPatterns_ExcludesDegenerateKeys covers AC-HRR-003: the
// exact regression observed in the live 2026-05-24 promotions —
// subagent_stop:unknown: (empty context_hash, subject "unknown", count 41,
// confidence 1) — is NO LONGER promoted after the eligibility predicate is
// wired into classifyHarnessPatterns (D4: filter at classification time).
// session_stop:: and user_prompt:: are likewise excluded, while the eligible
// control agent_invocation:Bash:<hash> still promotes.
func TestClassifyHarnessPatterns_ExcludesDegenerateKeys(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	t.Chdir(dir)

	// Seed: 41 degenerate subagent_stop:unknown: (the live regression count) +
	// 23 session_stop:: + 23 user_prompt:: + 5 control agent_invocation:Bash:hash.
	events := []map[string]string{}
	events = append(events, repeatEvents(map[string]string{
		"event_type": "subagent_stop", "subject": "unknown", "context_hash": "",
	}, 41)...)
	events = append(events, repeatEvents(map[string]string{
		"event_type": "session_stop", "subject": "", "context_hash": "",
	}, 23)...)
	events = append(events, repeatEvents(map[string]string{
		"event_type": "user_prompt", "subject": "", "context_hash": "",
	}, 23)...)
	events = append(events, repeatEvents(map[string]string{
		"event_type": "agent_invocation", "subject": "Bash", "context_hash": "ctx-hash-1",
	}, 5)...)
	seedEventLines(t, dir, events)

	patternCount, promoCount, err := classifyHarnessPatterns(dir)
	if err != nil {
		t.Fatalf("classifyHarnessPatterns error: %v", err)
	}

	keys := readPromotionKeys(t, dir)

	// AC-HRR-003: subagent_stop:unknown: MUST NOT be promoted.
	if keys["subagent_stop:unknown:"] {
		t.Errorf("REGRESSION: subagent_stop:unknown: was promoted (degenerate key not excluded)")
	}
	// AC-HRR-004: session_stop:: and user_prompt:: MUST NOT be promoted.
	if keys["session_stop::"] {
		t.Errorf("REGRESSION: session_stop:: was promoted (degenerate key not excluded)")
	}
	if keys["user_prompt::"] {
		t.Errorf("REGRESSION: user_prompt:: was promoted (degenerate key not excluded)")
	}
	// Control: agent_invocation:Bash:ctx-hash-1 (non-empty context, 5 obs → rule) MUST promote.
	if !keys["agent_invocation:Bash:ctx-hash-1"] {
		t.Errorf("control agent_invocation:Bash:ctx-hash-1 was NOT promoted (eligible pattern excluded); patternCount=%d promoCount=%d keys=%v",
			patternCount, promoCount, keys)
	}
}
