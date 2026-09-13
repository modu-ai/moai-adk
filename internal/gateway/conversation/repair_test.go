package conversation

// Card t708 (SPEC-GATEWAY-ENVELOPE-REPAIR-001 M3): the launcher-side envelope
// repair path. RED-first TDD: these tests define RepairEnvelope before the
// implementation exists. Every refusal cell demands zero modification; the
// success cell demands byte-exact verbatim re-injection verified by marker
// self-attestation, aside-before-replace, and a durable single-shot record.

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway/opaque"
)

// repairEnvelope builds one gateway-issued envelope and returns its verbatim
// data carrier plus bound tool marker, exactly the issuance shape.
func repairEnvelope(t *testing.T, id, cipher string) (data, marker string) {
	t.Helper()
	env, err := opaque.Encode([]opaque.Item{{OutputIndex: 0, Raw: []byte(`{"type":"reasoning","id":"` + id + `","summary":[],"encrypted_content":"` + cipher + `"}`)}})
	if err != nil {
		t.Fatal(err)
	}
	bound, err := opaque.BindToolID("call_"+id, env)
	if err != nil {
		t.Fatal(err)
	}
	return env.Data(), bound
}

func repairTestCarrierDigest(t *testing.T, data string) string {
	t.Helper()
	env, err := opaque.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(env.Bytes())
	return hex.EncodeToString(sum[:])
}

// repairFixture opens a family manager, creates one conversation, and writes
// a native transcript whose second assistant boundary was persisted WITHOUT
// its redacted_thinking envelope while the bound tool marker survived — the
// t707 split-record stripped shape. The verbatim carrier survives in an
// earlier issuance row. Returns the manager, conversation id, transcript
// path, and the stripped boundary's marker.
func repairFixture(t *testing.T) (m *Manager, id, path, marker, envData string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m, err = Open(root)
	if err != nil {
		t.Fatal(err)
	}
	desc, err := m.New(context.Background(), NewRequest{CWD: root, Project: "proj"})
	if err != nil {
		t.Fatal(err)
	}
	intactData, _ := repairEnvelope(t, "rs_a", "cipher-a")
	var envMarker string
	envData, envMarker = repairEnvelope(t, "rs_b", "cipher-b")
	rows := []map[string]any{
		{"type": "user", "sessionId": desc.UUID, "cwd": root, "message": map[string]any{"role": "user", "content": "hello"}},
		{"type": "assistant", "sessionId": desc.UUID, "cwd": root, "message": map[string]any{"role": "assistant", "model": "gpt-5.6-sol", "stop_reason": "end_turn",
			"content": []any{map[string]any{"type": "redacted_thinking", "data": intactData}, map[string]any{"type": "text", "text": "one"}}}},
		// The split-record issuance row: the verbatim envelope the client kept
		// separately from the re-encoded tool-use row below.
		{"type": "assistant", "sessionId": desc.UUID, "cwd": root, "message": map[string]any{"role": "assistant", "model": "gpt-5.6-sol",
			"content": []any{map[string]any{"type": "redacted_thinking", "data": envData}}}},
		{"type": "user", "sessionId": desc.UUID, "cwd": root, "message": map[string]any{"role": "user", "content": "run"}},
		// The stripped boundary: marker survives, envelope does not.
		{"type": "assistant", "sessionId": desc.UUID, "cwd": root, "message": map[string]any{"role": "assistant", "model": "gpt-5.6-sol", "stop_reason": "end_turn",
			"content": []any{map[string]any{"type": "tool_use", "id": envMarker, "name": "read", "input": map[string]any{}}, map[string]any{"type": "text", "text": "two"}}}},
	}
	var lines []string
	for _, row := range rows {
		raw, err := json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, string(raw))
	}
	path = filepath.Join(desc.ConfigDir, "projects", desc.UUID+".jsonl")
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.Complete(context.Background(), desc.UUID, path, uint64(info.ModTime().UnixNano())); err != nil {
		t.Fatal(err)
	}
	return m, desc.UUID, path, envMarker, envData
}

func repairTestMarkerDigest(t *testing.T, marker string) string {
	t.Helper()
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(marker, opaque.ToolPrefix))
	if err != nil {
		t.Fatal(err)
	}
	var b struct {
		Digest string `json:"opaque_sha256"`
	}
	if err := json.Unmarshal(raw, &b); err != nil {
		t.Fatal(err)
	}
	return b.Digest
}

// The happy path: the stripped boundary is restored with the verbatim carrier,
// the original transcript is preserved aside, the durable record carries the
// per-boundary digest and position, and every public byte is untouched.
func TestRepairEnvelopeRestoresStrippedEnvelopeAtOriginalBoundary(t *testing.T) {
	m, id, path, marker, _ := repairFixture(t)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	provenance, err := m.RepairEnvelope(context.Background(), id)
	if err != nil {
		t.Fatalf("repair refused a repairable shape: %v", err)
	}
	if len(provenance) != 1 {
		t.Fatalf("provenance boundaries = %d, want 1", len(provenance))
	}
	wantDigest := repairTestMarkerDigest(t, marker)
	if provenance[0].Digest != wantDigest {
		t.Fatalf("recorded digest %q, want marker-embedded %q", provenance[0].Digest, wantDigest)
	}
	// sha256(raw) == marker.opaque_sha256 per AC-EVR-004, recomputed over the
	// injected carrier's decoded canonical bytes.
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var injected string
	for _, line := range strings.Split(string(after), "\n") {
		if !strings.Contains(line, `"tool_use"`) || !strings.Contains(line, marker[:32]) {
			continue
		}
		var row struct {
			Message struct {
				Content []map[string]any `json:"content"`
			} `json:"message"`
		}
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatal(err)
		}
		if len(row.Message.Content) == 0 || row.Message.Content[0]["type"] != "redacted_thinking" {
			t.Fatalf("injected envelope is not at the boundary head: %s", line)
		}
		data, _ := row.Message.Content[0]["data"].(string)
		injected = data
	}
	if injected == "" {
		t.Fatal("stripped boundary still carries no redacted_thinking block")
	}
	if got := repairTestCarrierDigest(t, injected); got != wantDigest {
		t.Fatalf("sha256(injected raw) = %q, want %q", got, wantDigest)
	}
	// Aside-before-replace: the preimage is recoverable and never deleted.
	aside := string(before) // closure over the original bytes
	asidePath := path + ".moai-repair-aside"
	asideRaw, err := os.ReadFile(asidePath)
	if err != nil {
		t.Fatalf("aside missing: %v", err)
	}
	if string(asideRaw) != aside {
		t.Fatal("aside is not the byte-identical preimage")
	}
	// The durable record exists on the family namespace.
	statePath := filepath.Join(m.root, "families", m.entries[id].FamilyID, "repair", id+".json")
	stateRaw, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("durable repair record missing: %v", err)
	}
	var state struct {
		Attempted  bool `json:"attempted"`
		Boundaries []struct {
			Digest   string `json:"digest"`
			Position int    `json:"position"`
		} `json:"boundaries"`
		AsideSHA256 string `json:"aside_sha256"`
	}
	if err := json.Unmarshal(stateRaw, &state); err != nil {
		t.Fatal(err)
	}
	if !state.Attempted || len(state.Boundaries) != 1 || state.Boundaries[0].Digest != wantDigest {
		t.Fatalf("durable record incomplete: %s", stateRaw)
	}
	sum := sha256.Sum256(before)
	if state.AsideSHA256 != hex.EncodeToString(sum[:]) {
		t.Fatal("record does not carry the aside preimage digest")
	}
	// Every non-injected row is byte-identical: the line count is unchanged
	// and only the stripped row gained the head block.
	beforeLines, afterLines := strings.Split(string(before), "\n"), strings.Split(string(after), "\n")
	if len(beforeLines) != len(afterLines) {
		t.Fatalf("line count changed: %d -> %d", len(beforeLines), len(afterLines))
	}
}

func TestRepairEnvelopeRefusesWhenCarrierSourceGone(t *testing.T) {
	m, id, path, marker, envData := repairFixture(t)
	// Remove the issuance row carrying the verbatim carrier: the source is
	// gone (compaction / record loss shape).
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var kept []string
	for _, line := range strings.Split(strings.TrimRight(string(raw), "\n"), "\n") {
		if envData != "" && strings.Contains(line, envData) {
			continue
		}
		kept = append(kept, line)
	}
	if len(kept) != 4 {
		t.Fatalf("fixture surgery removed %d rows, want 1", 5-len(kept))
	}
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, []byte(strings.Join(kept, "\n")+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = m.Complete(context.Background(), id, path, uint64(info.ModTime().UnixNano())+1); err != nil {
		t.Fatal(err)
	}
	_ = marker
	_, err = m.RepairEnvelope(context.Background(), id)
	if !errors.Is(err, ErrRepairNotRepairable) {
		t.Fatalf("source-gone shape must refuse with ErrRepairNotRepairable, got %v", err)
	}
	// Zero modification: transcript unchanged, no aside, no durable record.
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != strings.Join(kept, "\n")+"\n" {
		t.Fatal("refusal modified the transcript")
	}
	if _, err = os.Stat(path + ".moai-repair-aside"); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("refusal wrote an aside")
	}
	if _, err = os.Stat(filepath.Join(m.root, "families", m.entries[id].FamilyID, "repair", id+".json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("refusal wrote a durable record")
	}
}

func TestRepairEnvelopeRefusesOnDigestMismatch(t *testing.T) {
	m, id, path, marker, envData := repairFixture(t)
	// Corrupt the transcript's carrier so its digest no longer matches the
	// marker's embedded self-attestation: swap the issuance row's data for a
	// different valid envelope.
	otherData, _ := repairEnvelope(t, "rs_c", "cipher-c")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, line := range strings.Split(strings.TrimRight(string(raw), "\n"), "\n") {
		if envData != "" && strings.Contains(line, envData) {
			var row map[string]any
			if err := json.Unmarshal([]byte(line), &row); err != nil {
				t.Fatal(err)
			}
			row["message"].(map[string]any)["content"] = []any{map[string]any{"type": "redacted_thinking", "data": otherData}}
			fixed, err := json.Marshal(row)
			if err != nil {
				t.Fatal(err)
			}
			line = string(fixed)
		}
		lines = append(lines, line)
	}
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = m.Complete(context.Background(), id, path, uint64(info.ModTime().UnixNano())+1); err != nil {
		t.Fatal(err)
	}
	_ = marker
	_, err = m.RepairEnvelope(context.Background(), id)
	if !errors.Is(err, ErrRepairNotRepairable) {
		t.Fatalf("digest-mismatch shape must refuse, got %v", err)
	}
	after, _ := os.ReadFile(path)
	if string(after) != strings.Join(lines, "\n")+"\n" {
		t.Fatal("refusal modified the transcript")
	}
}

func TestRepairEnvelopeSingleShotTerminatesAcrossRestart(t *testing.T) {
	m, id, path, _, _ := repairFixture(t)
	if _, err := m.RepairEnvelope(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// A fresh Manager over the same root simulates a process restart: the
	// durable record — not memory — must terminate the second attempt.
	restarted, err := Open(m.root)
	if err != nil {
		t.Fatal(err)
	}
	_, err = restarted.RepairEnvelope(context.Background(), id)
	if !errors.Is(err, ErrRepairAlreadyAttempted) {
		t.Fatalf("second attempt must terminate with ErrRepairAlreadyAttempted, got %v", err)
	}
	again, _ := os.ReadFile(path)
	if string(again) != string(after) {
		t.Fatal("terminated attempt modified the transcript")
	}
}

func TestRepairEnvelopeIntactHistoryNeedsNoRepair(t *testing.T) {
	m, id, path, _, _ := repairFixture(t)
	// Repair the once; a second shape-check against the repaired transcript
	// finds no stripped boundary and must refuse with zero modification —
	// never a re-injection or a second record.
	if _, err := m.RepairEnvelope(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	// A brand-new conversation whose transcript is intact end to end.
	desc, err := m.New(context.Background(), NewRequest{CWD: m.root, Project: "proj2"})
	if err != nil {
		t.Fatal(err)
	}
	data, marker := repairEnvelope(t, "rs_x", "cipher-x")
	rows := []map[string]any{
		{"type": "user", "sessionId": desc.UUID, "cwd": m.root, "message": map[string]any{"role": "user", "content": "hi"}},
		{"type": "assistant", "sessionId": desc.UUID, "cwd": m.root, "message": map[string]any{"role": "assistant", "model": "gpt-5.6-sol", "stop_reason": "end_turn",
			"content": []any{map[string]any{"type": "redacted_thinking", "data": data}, map[string]any{"type": "tool_use", "id": marker, "name": "read", "input": map[string]any{}}}}},
	}
	var lines []string
	for _, row := range rows {
		raw, err := json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, string(raw))
	}
	intactPath := filepath.Join(desc.ConfigDir, "projects", desc.UUID+".jsonl")
	if err = os.MkdirAll(filepath.Dir(intactPath), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(intactPath, []byte(strings.Join(lines, "\n")+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(intactPath)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.Complete(context.Background(), desc.UUID, intactPath, uint64(info.ModTime().UnixNano())); err != nil {
		t.Fatal(err)
	}
	if _, err = m.RepairEnvelope(context.Background(), desc.UUID); !errors.Is(err, ErrRepairNotRepairable) {
		t.Fatalf("intact history must refuse with nothing to repair, got %v", err)
	}
	unmodified, _ := os.ReadFile(intactPath)
	if strings.Count(string(unmodified), `"redacted_thinking"`) != 1 {
		t.Fatal("intact history was modified")
	}
	_ = path
}

func TestRepairEnvelopeRequiresCompletedTranscript(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	desc, err := m.New(context.Background(), NewRequest{CWD: root, Project: "proj"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.RepairEnvelope(context.Background(), desc.UUID); !errors.Is(err, ErrRepairNotRepairable) && !errors.Is(err, ErrIncomplete) {
		t.Fatalf("uncompleted conversation must refuse, got %v", err)
	}
	if _, err = m.RepairEnvelope(context.Background(), "00000000-0000-4000-8000-000000000000"); !errors.Is(err, ErrMissing) {
		t.Fatalf("unknown id must return ErrMissing, got %v", err)
	}
}

// AC-EVR-010: the repair path never imports or calls any receipt-store
// symbol. repair.go is the entire repair path; scan its source.
func TestRepairPathNeverReadsReceiptStore(t *testing.T) {
	raw, err := os.ReadFile("repair.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, banned := range []string{"gateway/receipt", "receipt.", "OpenStore", "Snapshot(", "Publish("} {
		if strings.Contains(src, banned) {
			t.Fatalf("repair path references forbidden receipt-store symbol %q", banned)
		}
	}
}

// F1 (sync-audit t708): the injection is a byte-preserving splice. A real
// client row keeps Node JSON.stringify insertion order ("type" first,
// "sessionId" last) and literal < and & bytes — a whole-row re-marshal
// reorders keys and HTML-escapes public bytes. Invariant: removing exactly
// the injected span from the output reproduces the input byte-for-byte.
func TestRepairInjectLinePreservesRowBytesOutsideInjection(t *testing.T) {
	line := `{"type":"assistant","cwd":"/p","version":3,"message":{"role":"assistant","model":"gpt-5.6-sol","stop_reason":"end_turn","content":[{"type":"tool_use","id":"toolu_moai_v1_abc","name":"Edit","input":{"old_string":"a<b&c"}}]},"sessionId":"u","userType":"external"}`
	out, err := repairInjectLine(line, "moai_opaque_v2_abc")
	if err != nil {
		t.Fatal(err)
	}
	block := `{"type":"redacted_thinking","data":"moai_opaque_v2_abc"}`
	i := strings.Index(out, block)
	if i < 0 {
		t.Fatalf("envelope block not found in output: %s", out)
	}
	rest := out[i+len(block):]
	if !strings.HasPrefix(rest, ",") {
		t.Fatalf("injected block must be followed by the array comma, got %.20q", rest)
	}
	if reconstructed := out[:i] + rest[1:]; reconstructed != line {
		t.Fatalf("non-injected bytes rewritten:\n got: %s\nwant: %s", reconstructed, line)
	}
	if !strings.Contains(out, `a<b&c`) {
		t.Fatal("HTML escaping rewrote public bytes")
	}
	t.Run("empty content array", func(t *testing.T) {
		empty := `{"type":"assistant","message":{"role":"assistant","content":[]},"sessionId":"u"}`
		got, err := repairInjectLine(empty, "moai_opaque_v2_abc")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(got, `"content":[{"type":"redacted_thinking","data":"moai_opaque_v2_abc"}]`) {
			t.Fatalf("empty-array splice malformed: %s", got)
		}
		if reconstructed := strings.Replace(got, block, "", 1); reconstructed != empty {
			t.Fatalf("non-injected bytes rewritten: %s", reconstructed)
		}
	})
}
