package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/state"
)

// readJSONL reads the raw cache-usage.jsonl content under root.
func readJSONL(t *testing.T, root string) string {
	t.Helper()
	path := state.CacheUsageLogPath(root)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cache-usage.jsonl at %s: %v", path, err)
	}
	return string(data)
}

// TestPostToolUseCache_JSONLAppend (AC-PC-005) verifies the cache telemetry
// recorder extracts both cache token fields from a synthetic API-response usage
// block and appends a JSONL entry carrying BOTH
// cache_creation_input_tokens AND cache_read_input_tokens keys.
func TestPostToolUseCache_JSONLAppend(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	rec := NewCacheUsageRecorder()
	usage := CacheTokenUsage{
		SessionID:     "sess-append",
		Turn:          1,
		CacheCreation: 12450,
		CacheRead:     48200,
		Model:         "claude-sonnet-4-6",
		Elapsed:       6 * time.Minute, // not single-turn (long enough)
	}

	res := rec.Record(nil, usage)
	if res.Err != nil {
		t.Fatalf("Record returned error: %v", res.Err)
	}

	content := readJSONL(t, root)
	for _, key := range []string{"cache_creation_input_tokens", "cache_read_input_tokens"} {
		if !strings.Contains(content, key) {
			t.Errorf("JSONL entry missing key %q; got: %s", key, content)
		}
	}
}

// TestCacheUsage_TwoTurnSession_Turn2HitsCache (AC-PC-006) verifies that a
// synthetic 2-turn session records turn 1 with cache_creation>0 / cache_read==0
// and turn 2 with cache_read>0 (cache hit verified).
func TestCacheUsage_TwoTurnSession_Turn2HitsCache(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	rec := NewCacheUsageRecorder()
	// Turn 1: cache write (creation), no read.
	if res := rec.Record(nil, CacheTokenUsage{
		SessionID: "sess-2turn", Turn: 1, CacheCreation: 8000, CacheRead: 0,
		Model: "claude-sonnet-4-6", Elapsed: 30 * time.Second,
	}); res.Err != nil {
		t.Fatalf("turn1 Record: %v", res.Err)
	}
	// Turn 2: cache hit (read > 0).
	if res := rec.Record(nil, CacheTokenUsage{
		SessionID: "sess-2turn", Turn: 2, CacheCreation: 0, CacheRead: 7600,
		Model: "claude-sonnet-4-6", Elapsed: 2 * time.Minute,
	}); res.Err != nil {
		t.Fatalf("turn2 Record: %v", res.Err)
	}

	entries, err := state.ReadCacheUsage(root)
	if err != nil {
		t.Fatalf("ReadCacheUsage: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(entries))
	}

	var turn1, turn2 *state.CacheUsageEntry
	for i := range entries {
		switch entries[i].Turn {
		case 1:
			turn1 = &entries[i]
		case 2:
			turn2 = &entries[i]
		}
	}
	if turn1 == nil || turn2 == nil {
		t.Fatalf("missing turn entries: turn1=%v turn2=%v", turn1, turn2)
	}
	if turn1.CacheCreation <= 0 {
		t.Errorf("turn1 CacheCreation = %d, want > 0", turn1.CacheCreation)
	}
	if turn1.CacheRead != 0 {
		t.Errorf("turn1 CacheRead = %d, want 0", turn1.CacheRead)
	}
	if turn2.CacheRead <= 0 {
		t.Errorf("turn2 CacheRead = %d, want > 0 (cache hit)", turn2.CacheRead)
	}
}

// TestPostToolUseCache_SingleTurnSession_PenaltyWarning (AC-PC-010) verifies the
// single-turn cache-write penalty warning. A session with turn==1 only AND
// elapsed wall-time < 5min must emit a log line containing
// "single-turn cache write penalty risk" AND a recommendation containing
// `session_ttl: "off"`. A 2+ turn fixture must NOT emit the warning.
func TestPostToolUseCache_SingleTurnSession_PenaltyWarning(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	rec := NewCacheUsageRecorder()

	// Positive case: turn==1, elapsed < 5min → warning expected.
	resSingle := rec.Record(nil, CacheTokenUsage{
		SessionID: "sess-single", Turn: 1, CacheCreation: 9000, CacheRead: 0,
		Model: "claude-sonnet-4-6", Elapsed: 90 * time.Second,
	})
	if resSingle.Err != nil {
		t.Fatalf("single-turn Record: %v", resSingle.Err)
	}
	if !resSingle.PenaltyWarning {
		t.Errorf("single-turn session: PenaltyWarning = false, want true")
	}
	if !strings.Contains(resSingle.WarningMessage, "single-turn cache write penalty risk") {
		t.Errorf("warning message missing penalty string; got: %q", resSingle.WarningMessage)
	}
	if !strings.Contains(resSingle.WarningMessage, `session_ttl: "off"`) {
		t.Errorf("warning message missing `session_ttl: \"off\"` recommendation; got: %q", resSingle.WarningMessage)
	}

	// Negative case: 2-turn session → NO warning on turn 2 (false-positive guard).
	root2 := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root2)
	rec2 := NewCacheUsageRecorder()
	if res := rec2.Record(nil, CacheTokenUsage{
		SessionID: "sess-multi", Turn: 1, CacheCreation: 9000, CacheRead: 0,
		Model: "claude-sonnet-4-6", Elapsed: 30 * time.Second,
	}); res.Err != nil {
		t.Fatalf("multi turn1 Record: %v", res.Err)
	}
	resTurn2 := rec2.Record(nil, CacheTokenUsage{
		SessionID: "sess-multi", Turn: 2, CacheCreation: 0, CacheRead: 7600,
		Model: "claude-sonnet-4-6", Elapsed: 3 * time.Minute,
	})
	if resTurn2.Err != nil {
		t.Fatalf("multi turn2 Record: %v", resTurn2.Err)
	}
	if resTurn2.PenaltyWarning {
		t.Errorf("2-turn session turn2: PenaltyWarning = true, want false (false-positive)")
	}
	if strings.Contains(resTurn2.WarningMessage, "single-turn cache write penalty risk") {
		t.Errorf("2-turn session turn2 must NOT emit penalty warning; got: %q", resTurn2.WarningMessage)
	}
}

// TestExtractCacheTokenUsage_FromAPIResponse verifies token extraction from a
// raw Anthropic API response usage block (the field-extraction half of
// REQ-PC-004).
func TestExtractCacheTokenUsage_FromAPIResponse(t *testing.T) {
	raw := []byte(`{
		"model": "claude-sonnet-4-6",
		"usage": {
			"input_tokens": 120,
			"cache_creation_input_tokens": 12450,
			"cache_read_input_tokens": 48200,
			"output_tokens": 300
		}
	}`)
	usage, ok := ExtractCacheTokenUsage(raw)
	if !ok {
		t.Fatalf("ExtractCacheTokenUsage returned ok=false for valid usage block")
	}
	if usage.CacheCreation != 12450 {
		t.Errorf("CacheCreation = %d, want 12450", usage.CacheCreation)
	}
	if usage.CacheRead != 48200 {
		t.Errorf("CacheRead = %d, want 48200", usage.CacheRead)
	}
	if usage.Model != "claude-sonnet-4-6" {
		t.Errorf("Model = %q, want claude-sonnet-4-6", usage.Model)
	}
}

// TestExtractCacheTokenUsage_NoUsageBlock verifies graceful handling when the
// response has no usage block (ok=false, no panic).
func TestExtractCacheTokenUsage_NoUsageBlock(t *testing.T) {
	if _, ok := ExtractCacheTokenUsage([]byte(`{"foo":"bar"}`)); ok {
		t.Errorf("ExtractCacheTokenUsage ok=true for response without usage block, want false")
	}
	if _, ok := ExtractCacheTokenUsage([]byte(`not json`)); ok {
		t.Errorf("ExtractCacheTokenUsage ok=true for malformed JSON, want false")
	}
}

// TestCacheUsageRecorder_AppendErrorIsObservationOnly verifies that a JSONL
// append failure is surfaced in the result (res.Err) but never panics — the
// recorder is observation-only. Here .moai is a regular FILE, so MkdirAll of
// .moai/state/ fails.
func TestCacheUsageRecorder_AppendErrorIsObservationOnly(t *testing.T) {
	root := t.TempDir()
	// Create .moai as a file so the state-dir creation under it fails.
	if err := os.WriteFile(filepath.Join(root, ".moai"), []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("seed .moai-as-file: %v", err)
	}
	input := &HookInput{CWD: root, SessionID: "s"}
	res := NewCacheUsageRecorder().Record(input, CacheTokenUsage{
		SessionID: "s", Turn: 1, CacheCreation: 100, Elapsed: 6 * time.Minute,
	})
	if res.Err == nil {
		t.Errorf("expected append error when .moai is a file, got nil")
	}
	// The penalty heuristic should not fire here (elapsed >= 5min).
	if res.PenaltyWarning {
		t.Errorf("PenaltyWarning should be false (elapsed >= 5min)")
	}
}

// TestLogCacheUsage_AppendErrorTolerated verifies the live path tolerates an
// append failure (no panic, no propagation).
func TestLogCacheUsage_AppendErrorTolerated(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".moai"), []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("seed .moai-as-file: %v", err)
	}
	// Must not panic even though the append will fail.
	logCacheUsage(&HookInput{
		CWD:       root,
		SessionID: "s",
		ToolResponse: []byte(`{"model":"m","turn":1,"usage":{"cache_read_input_tokens":10,"cache_creation_input_tokens":0}}`),
	})
}

// TestLogCacheUsage_LiveHandlePath verifies the live PostToolUse integration:
// a tool response carrying a usage block appends a JSONL entry; a response
// without a usage block is a no-op.
func TestLogCacheUsage_LiveHandlePath(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	withUsage := &HookInput{
		SessionID: "live-sess",
		CWD:       root,
		ToolResponse: []byte(`{
			"model": "claude-sonnet-4-6",
			"turn": 2,
			"usage": {"cache_creation_input_tokens": 0, "cache_read_input_tokens": 7600}
		}`),
	}
	logCacheUsage(withUsage)

	entries, err := state.ReadCacheUsage(root)
	if err != nil {
		t.Fatalf("ReadCacheUsage: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("want 1 entry from live path, got %d", len(entries))
	}
	if entries[0].CacheRead != 7600 || entries[0].Turn != 2 {
		t.Errorf("entry = %+v, want CacheRead=7600 Turn=2", entries[0])
	}

	// No usage block → no append.
	root2 := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root2)
	logCacheUsage(&HookInput{SessionID: "x", CWD: root2, ToolResponse: []byte(`{"foo":"bar"}`)})
	entries2, _ := state.ReadCacheUsage(root2)
	if len(entries2) != 0 {
		t.Errorf("response without usage block must not append; got %d entries", len(entries2))
	}

	// Nil input / empty response → no panic, no append.
	logCacheUsage(nil)
	logCacheUsage(&HookInput{SessionID: "y"})
}

// TestCacheUsageRecorder_ProjectRootResolution verifies the recorder resolves
// the project root via input.CWD when present (B7 path resolution), writing the
// JSONL under that root rather than os.Getwd().
func TestCacheUsageRecorder_ProjectRootResolution(t *testing.T) {
	cwdRoot := t.TempDir()
	envRoot := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", envRoot)

	rec := NewCacheUsageRecorder()
	input := &HookInput{CWD: cwdRoot, SessionID: "s", ToolName: "Agent"}
	res := rec.Record(input, CacheTokenUsage{
		SessionID: "s", Turn: 1, CacheCreation: 100, CacheRead: 0, Elapsed: 6 * time.Minute,
	})
	if res.Err != nil {
		t.Fatalf("Record: %v", res.Err)
	}

	// input.CWD takes precedence over CLAUDE_PROJECT_DIR.
	if _, err := os.Stat(filepath.Join(cwdRoot, ".moai", "state", "cache-usage.jsonl")); err != nil {
		t.Errorf("expected JSONL under input.CWD root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(envRoot, ".moai", "state", "cache-usage.jsonl")); err == nil {
		t.Errorf("JSONL must NOT be written under CLAUDE_PROJECT_DIR when input.CWD is set")
	}
}
