package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/statusline"
)

// fixtureSessionID matches the transcript filename stem used by the fixtures
// below, proving session_id derivation from the transcript path.
const fixtureSessionID = "44400e2f-bb9f-4a89-9831-2686c5b61050"

// writeTranscriptFixture writes the given JSONL lines to a temp transcript
// file named after fixtureSessionID and returns its path.
func writeTranscriptFixture(t *testing.T, lines []string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), fixtureSessionID+".jsonl")
	content := strings.Join(lines, "\n")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write transcript fixture: %v", err)
	}
	return path
}

// assistantLine builds a transcript assistant line with the nested CC schema
// (model + usage live under "message", verified against a real transcript).
func assistantLine(model string, sidechain bool, usage string) string {
	return fmt.Sprintf(`{"type":"assistant","isSidechain":%t,"sessionId":%q,"cwd":"/tmp/proj","uuid":"u-1","timestamp":"2026-08-17T00:00:00Z","message":{"id":"msg_1","type":"message","role":"assistant","model":%q,"usage":%s}}`,
		sidechain, fixtureSessionID, model, usage)
}

// defaultFixtureLines returns the 6-line card fixture: two glm-5.3 assistant
// messages (one sidechain), two claude-opus-5 (one sidechain, cache fields
// missing), one non-assistant line without usage, one assistant line with
// usage carrying only a subset of fields.
func defaultFixtureLines() []string {
	return []string{
		`{"type":"mode","mode":"normal","sessionId":"` + fixtureSessionID + `"}`,
		assistantLine("glm-5.3", false, `{"input_tokens":100,"cache_creation_input_tokens":50,"cache_read_input_tokens":200,"output_tokens":30}`),
		assistantLine("glm-5.3", true, `{"input_tokens":10,"cache_creation_input_tokens":5,"cache_read_input_tokens":20,"output_tokens":3}`),
		assistantLine("claude-opus-5", false, `{"input_tokens":1000,"cache_creation_input_tokens":500,"cache_read_input_tokens":2000,"output_tokens":300}`),
		assistantLine("claude-opus-5", true, `{"input_tokens":100,"output_tokens":7}`),
		`{"type":"user","isSidechain":false,"message":{"role":"user","content":"hello"}}`,
		assistantLine("glm-5.3", false, `{"input_tokens":7}`),
	}
}

// usage returns a tokenUsageTotals literal for assertions.
func usage(in, cc, cr, out int64) tokenUsageTotals {
	return tokenUsageTotals{
		InputTokens:         in,
		CacheCreationTokens: cc,
		CacheReadTokens:     cr,
		OutputTokens:        out,
	}
}

func TestAggregateTranscriptUsageBucketing(t *testing.T) {
	path := writeTranscriptFixture(t, defaultFixtureLines())
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = f.Close() }()

	agg, err := aggregateTranscriptUsage(f)
	if err != nil {
		t.Fatalf("aggregateTranscriptUsage: %v", err)
	}

	// Message counts: 5 assistant lines, 2 of them sidechain.
	if agg.Messages.Assistant != 5 {
		t.Errorf("messages.assistant = %d, want 5", agg.Messages.Assistant)
	}
	if agg.Messages.Sidechain != 2 {
		t.Errorf("messages.sidechain = %d, want 2", agg.Messages.Sidechain)
	}
	if agg.Skipped != 0 {
		t.Errorf("skipped = %d, want 0", agg.Skipped)
	}

	// glm pool: line 1 (main) + line 7 (main, input only) + line 2 (subagent).
	wantGLM := tokensOriginSplit{
		Main:     usage(107, 50, 200, 30),
		Subagent: usage(10, 5, 20, 3),
	}
	if got := agg.Pools["glm"]; got == nil || *got != wantGLM {
		t.Errorf("pools.glm = %+v, want %+v", got, wantGLM)
	}

	wantClaude := tokensOriginSplit{
		Main:     usage(1000, 500, 2000, 300),
		Subagent: usage(100, 0, 0, 7),
	}
	if got := agg.Pools["claude"]; got == nil || *got != wantClaude {
		t.Errorf("pools.claude = %+v, want %+v", got, wantClaude)
	}

	// Models detail keyed by exact model name.
	wantGLMModel := tokensOriginSplit{
		Main:     usage(107, 50, 200, 30),
		Subagent: usage(10, 5, 20, 3),
	}
	if got := agg.Models["glm-5.3"]; got == nil || *got != wantGLMModel {
		t.Errorf("models[glm-5.3] = %+v, want %+v", got, wantGLMModel)
	}
	if got := agg.Models["claude-opus-5"]; got == nil || *got != wantClaude {
		t.Errorf("models[claude-opus-5] = %+v, want %+v", got, wantClaude)
	}
	if len(agg.Models) != 2 {
		t.Errorf("len(models) = %d, want 2", len(agg.Models))
	}
}

func TestAggregateTranscriptUsageOtherPoolKey(t *testing.T) {
	path := writeTranscriptFixture(t, []string{
		assistantLine("GPT-4o", false, `{"input_tokens":11,"output_tokens":2}`),
	})
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = f.Close() }()

	agg, err := aggregateTranscriptUsage(f)
	if err != nil {
		t.Fatalf("aggregateTranscriptUsage: %v", err)
	}
	got := agg.Pools["gpt-4o"]
	if got == nil {
		t.Fatalf("pools[gpt-4o] missing; pools = %+v", agg.Pools)
	}
	if want := (tokensOriginSplit{Main: usage(11, 0, 0, 2)}); *got != want {
		t.Errorf("pools[gpt-4o] = %+v, want %+v", got, want)
	}
	if _, ok := agg.Pools["glm"]; ok {
		t.Errorf("pools must not contain glm for a non-glm model")
	}
}

func TestAggregateTranscriptUsageTruncatedTail(t *testing.T) {
	path := writeTranscriptFixture(t, []string{
		assistantLine("glm-5.3", false, `{"input_tokens":42,"output_tokens":1}`),
		`{"type":"assistant","isSidechain":false,"message":{"model":"glm-5.3","usage":{"input_to`, // truncated tail
	})
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = f.Close() }()

	agg, err := aggregateTranscriptUsage(f)
	if err != nil {
		t.Fatalf("aggregateTranscriptUsage: %v — a truncated tail must not abort the record", err)
	}
	if agg.Skipped != 1 {
		t.Errorf("skipped = %d, want 1", agg.Skipped)
	}
	if got := agg.Pools["glm"].Main.InputTokens; got != 42 {
		t.Errorf("pools.glm.main.input = %d, want 42", got)
	}
}

func TestRunTokensRecordAppendsLedger(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)
	stateDir := filepath.Join(tmp, ".moai", "state")
	path := writeTranscriptFixture(t, defaultFixtureLines())

	for i := 0; i < 2; i++ {
		rec, err := runTokensRecord(tokensRecordOpts{
			Transcript: path,
			Card:       "t86",
			Role:       "run",
			StateDir:   stateDir,
		})
		if err != nil {
			t.Fatalf("runTokensRecord call %d: %v", i+1, err)
		}
		if rec.SchemaVersion != tokensSchemaVersion {
			t.Errorf("schema_version = %d, want %d", rec.SchemaVersion, tokensSchemaVersion)
		}
	}

	data, err := os.ReadFile(filepath.Join(stateDir, tokensLedgerFilename))
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("ledger has %d lines, want 2 (append semantics)", len(lines))
	}
	for i, line := range lines {
		var rec TokensRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("ledger line %d is not valid JSON: %v", i+1, err)
		}
		if rec.SessionID != fixtureSessionID {
			t.Errorf("line %d session_id = %q, want %q (from filename)", i+1, rec.SessionID, fixtureSessionID)
		}
		if rec.Card != "t86" || rec.Role != "run" {
			t.Errorf("line %d card/role = %q/%q, want t86/run", i+1, rec.Card, rec.Role)
		}
		if rec.RecordedAt == "" {
			t.Errorf("line %d recorded_at empty", i+1)
		}
	}
}

func TestRunTokensRecordJSONFlagNoLedger(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)
	stateDir := filepath.Join(tmp, ".moai", "state")
	path := writeTranscriptFixture(t, defaultFixtureLines())

	var buf bytes.Buffer
	rec, err := runTokensRecord(tokensRecordOpts{
		Transcript: path,
		AsJSON:     true,
		Stdout:     &buf,
		StateDir:   stateDir,
	})
	if err != nil {
		t.Fatalf("runTokensRecord: %v", err)
	}
	var printed TokensRecord
	if err := json.Unmarshal(buf.Bytes(), &printed); err != nil {
		t.Fatalf("stdout is not valid record JSON: %v\nstdout: %s", err, buf.String())
	}
	if !reflect.DeepEqual(printed, *rec) {
		t.Errorf("printed record differs from returned record")
	}

	if _, err := os.Stat(filepath.Join(stateDir, tokensLedgerFilename)); !os.IsNotExist(err) {
		t.Errorf("--json must not create the ledger file; stat err = %v", err)
	}
}

func TestRunTokensRecordSessionResolver(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)
	path := writeTranscriptFixture(t, defaultFixtureLines())

	fake := func(sessionID string) (string, error) {
		if sessionID != "aaaa-bbbb-cccc" {
			return "", fmt.Errorf("unexpected session id: %s", sessionID)
		}
		return path, nil
	}

	rec, err := runTokensRecord(tokensRecordOpts{
		Session:  "aaaa-bbbb-cccc",
		Card:     "t86",
		Role:     "lead",
		Resolver: fake,
	})
	if err != nil {
		t.Fatalf("runTokensRecord: %v", err)
	}
	if rec.SessionID != "aaaa-bbbb-cccc" {
		t.Errorf("session_id = %q, want the --session value", rec.SessionID)
	}
	if rec.Transcript != path {
		t.Errorf("transcript = %q, want resolver-returned %q", rec.Transcript, path)
	}
}

func TestRunTokensRecordContextSnapshot(t *testing.T) {
	path := writeTranscriptFixture(t, defaultFixtureLines())

	// Present: fields appear under record.context.
	tmp := t.TempDir()
	t.Chdir(tmp)
	stateDir := filepath.Join(tmp, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}
	// The record is addressed per session (SPEC-SESSION-TELEMETRY-001
	// REQ-ST-001): it lives under state/context-usage/ named for the session
	// being accounted for, which here is the transcript fixture's own id.
	// The fixture is marshalled from the statusline record type rather than
	// hand-authored, so this package declares no second copy of that schema
	// (REQ-ST-005) — the duplicate declaration this test's fixture used to
	// carry is exactly what the consolidation removed.
	snap, err := json.Marshal(&statusline.SessionTelemetryRecord{
		SchemaVersion:     1,
		SessionID:         fixtureSessionID,
		WriterPID:         42,
		CapturedAt:        "2026-08-17T00:00:00Z",
		ContextWindowSize: 1000000,
		TokensUsed:        500000,
		RawPct:            50.0,
		Stage:             "soft",
		Band:              "large",
	})
	if err != nil {
		t.Fatalf("marshal context fixture: %v", err)
	}
	snapPath := statusline.SessionTelemetryPath(stateDir, fixtureSessionID)
	if err := os.MkdirAll(filepath.Dir(snapPath), 0o755); err != nil {
		t.Fatalf("mkdir context-usage: %v", err)
	}
	if err := os.WriteFile(snapPath, snap, 0o644); err != nil {
		t.Fatalf("write context fixture: %v", err)
	}

	rec, err := runTokensRecord(tokensRecordOpts{Transcript: path, StateDir: stateDir})
	if err != nil {
		t.Fatalf("runTokensRecord: %v", err)
	}
	if rec.Context == nil {
		t.Fatalf("record.context is nil, want embedded snapshot")
	}
	if rec.Context.SessionID != fixtureSessionID ||
		rec.Context.ContextWindowSize != 1000000 ||
		rec.Context.TokensUsed != 500000 ||
		rec.Context.Stage != "soft" ||
		rec.Context.Band != "large" {
		t.Errorf("record.context = %+v, want embedded snapshot fields", rec.Context)
	}

	// Absent: context is nil and recording still succeeds.
	tmp2 := t.TempDir()
	t.Chdir(tmp2)
	rec2, err := runTokensRecord(tokensRecordOpts{
		Transcript: path,
		StateDir:   filepath.Join(tmp2, ".moai", "state"),
	})
	if err != nil {
		t.Fatalf("runTokensRecord without context snapshot: %v", err)
	}
	if rec2.Context != nil {
		t.Errorf("record.context = %+v, want nil when snapshot absent", rec2.Context)
	}
}

func TestRunTokensRecordSiblingSubagentFiles(t *testing.T) {
	// CC 2.1.23x stores subagent sidechains as separate sibling files under
	// <dir>/<session-stem>/subagents/*.jsonl (measured: main file carries 0
	// isSidechain:true lines). Recording the main transcript must aggregate
	// those siblings into the subagent origin.
	tmp := t.TempDir()
	t.Chdir(tmp)

	sessionDir := t.TempDir()
	mainPath := filepath.Join(sessionDir, fixtureSessionID+".jsonl")
	mainLine := assistantLine("glm-5.3", false, `{"input_tokens":100,"output_tokens":10}`)
	if err := os.WriteFile(mainPath, []byte(mainLine), 0o644); err != nil {
		t.Fatalf("write main transcript: %v", err)
	}
	subDir := filepath.Join(sessionDir, fixtureSessionID, "subagents")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir subagents: %v", err)
	}
	subLine := assistantLine("glm-5.3", true, `{"input_tokens":55,"output_tokens":6}`)
	for _, name := range []string{"agent-1.jsonl", "agent-2.jsonl"} {
		if err := os.WriteFile(filepath.Join(subDir, name), []byte(subLine), 0o644); err != nil {
			t.Fatalf("write subagent transcript %s: %v", name, err)
		}
	}

	rec, err := runTokensRecord(tokensRecordOpts{Transcript: mainPath, AsJSON: true})
	if err != nil {
		t.Fatalf("runTokensRecord: %v", err)
	}
	got := rec.Pools["glm"]
	if got == nil {
		t.Fatalf("pools[glm] missing")
	}
	if got.Main.InputTokens != 100 || got.Main.OutputTokens != 10 {
		t.Errorf("glm.main = %+v, want input 100 / output 10", got.Main)
	}
	if got.Subagent.InputTokens != 110 || got.Subagent.OutputTokens != 12 {
		t.Errorf("glm.subagent = %+v, want input 110 / output 12 (2 sibling files)", got.Subagent)
	}
	if rec.Messages.Sidechain != 2 {
		t.Errorf("messages.sidechain = %d, want 2", rec.Messages.Sidechain)
	}
}

func TestRunTokensRecordFlagValidation(t *testing.T) {
	cases := []struct {
		name string
		opts tokensRecordOpts
	}{
		{"neither", tokensRecordOpts{}},
		{"both", tokensRecordOpts{Transcript: "/tmp/a.jsonl", Session: "aaaa-bbbb-cccc"}},
	}
	for _, tc := range cases {
		if _, err := runTokensRecord(tc.opts); err == nil {
			t.Errorf("%s: expected error, got nil", tc.name)
		} else if !strings.Contains(err.Error(), "exactly one of") {
			t.Errorf("%s: error %q does not mention the exactly-one-of requirement", tc.name, err)
		}
	}
}
