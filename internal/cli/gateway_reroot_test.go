package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway/conversation"
	"github.com/modu-ai/moai-adk/internal/gateway/opaque"
	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

// Card t700 (SPEC-GATEWAY-WEDGE-REROOT-001): the launcher-side wedge recovery
// path. Every test names the requirement it locks: trailing-boundary-bounded
// removal (REQ-WRR-003-1), the durable single-shot bound (REQ-WRR-003-2,
// REQ-WRR-006), non-destructive aside-before-replace (REQ-WRR-003-3), the
// no-gateway-write property (REQ-WRR-003-4), explicit-user invocation
// (REQ-WRR-005), and the never-recoverable forged/stripped tail (REQ-WRR-007).

// rerootWriteTranscript writes one JSON object per line, the native transcript
// layout, and returns the raw bytes written.
func rerootWriteTranscript(t *testing.T, path string, rows []any) []byte {
	t.Helper()
	var b strings.Builder
	enc := json.NewEncoder(&b)
	for _, row := range rows {
		if err := enc.Encode(row); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(b.String()), 0600); err != nil {
		t.Fatal(err)
	}
	return []byte(b.String())
}

func rerootUserRow(cwd, content string) map[string]any {
	return map[string]any{"type": "user", "cwd": cwd, "message": map[string]any{"role": "user", "content": content}}
}

func rerootAssistantRow(cwd string, content []any) map[string]any {
	return map[string]any{"type": "assistant", "cwd": cwd, "message": map[string]any{"role": "assistant", "model": "gpt-5.6-sol", "stop_reason": "end_turn", "content": content}}
}

func rerootToolResultRow(cwd, toolUseID string) map[string]any {
	return map[string]any{"type": "user", "cwd": cwd, "message": map[string]any{"role": "user", "content": []any{
		map[string]any{"type": "tool_result", "tool_use_id": toolUseID, "content": "done"},
	}}}
}

// rerootAPIErrorRow is the terminal display row Claude Code records for a
// failed request; the recovery must preserve it untouched (it is neither the
// unpublished boundary nor a dependent tool result).
func rerootAPIErrorRow(cwd string) map[string]any {
	return map[string]any{"type": "assistant", "cwd": cwd, "isApiErrorMessage": true, "error": "api error: 400",
		"message": map[string]any{"role": "assistant", "model": "<synthetic>", "stop_reason": "stop_sequence",
			"content": []any{map[string]any{"type": "text", "text": "API Error: request failed"}}}}
}

// TestGatewayRerootRemovesOnlyTrailingUnpublishedBoundary locks REQ-WRR-003-1
// (AC-WRR-009): the recovery removes the trailing never-published assistant
// boundary and its dependent tool results, keeps every user-authored turn and
// the whole published prefix, and preserves the removed content aside where it
// stays retrievable.
func TestGatewayRerootRemovesOnlyTrailingUnpublishedBoundary(t *testing.T) {
	transcript := filepath.Join(t.TempDir(), "projects", "p", "u.jsonl")
	cwd := filepath.Join(filepath.Dir(filepath.Dir(transcript)), "proj")
	rows := []any{
		rerootUserRow(cwd, "hello"),
		rerootAssistantRow(cwd, []any{map[string]any{"type": "text", "text": "one"}}),
		rerootUserRow(cwd, "fix it"),
		rerootAssistantRow(cwd, []any{map[string]any{"type": "tool_use", "id": "toolu_phantom", "name": "write", "input": map[string]any{}}}),
		rerootToolResultRow(cwd, "toolu_phantom"),
		rerootUserRow(cwd, "continue"),
	}
	rerootWriteTranscript(t, transcript, rows)

	out, err := rerootGatewayTranscript(transcript)
	if err != nil {
		t.Fatal(err)
	}
	if out.Removed != 2 || out.Aside == "" || out.Guidance != "" {
		t.Fatalf("outcome=%+v", out)
	}
	after, err := os.ReadFile(transcript)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"toolu_phantom", `"write"`} {
		if strings.Contains(string(after), forbidden) {
			t.Fatalf("unpublished content survived in transcript: %q", forbidden)
		}
	}
	for _, required := range []string{`"hello"`, `"one"`, `"fix it"`, `"continue"`} {
		if !strings.Contains(string(after), required) {
			t.Fatalf("user/published content lost: %q", required)
		}
	}
	aside, err := os.ReadFile(out.Aside)
	if err != nil {
		t.Fatal(err)
	}
	for _, kept := range []string{"toolu_phantom", `"write"`, `"tool_result"`} {
		if !strings.Contains(string(aside), kept) {
			t.Fatalf("aside lost removed content: %q", kept)
		}
	}
}

// TestGatewayRerootTakesBoundaryOnlyWithoutDependents covers the broken
// mid-tool-call edge: an unpublished turn with a tool_use and no dependent
// tool results yet removes the boundary only.
func TestGatewayRerootTakesBoundaryOnlyWithoutDependents(t *testing.T) {
	transcript := filepath.Join(t.TempDir(), "u.jsonl")
	rows := []any{
		rerootUserRow("/p", "hello"),
		rerootAssistantRow("/p", []any{map[string]any{"type": "tool_use", "id": "toolu_bare", "name": "read", "input": map[string]any{}}}),
	}
	rerootWriteTranscript(t, transcript, rows)
	out, err := rerootGatewayTranscript(transcript)
	if err != nil {
		t.Fatal(err)
	}
	if out.Removed != 1 {
		t.Fatalf("removed=%d, want 1", out.Removed)
	}
	after, _ := os.ReadFile(transcript)
	if strings.Contains(string(after), "toolu_bare") {
		t.Fatal("boundary survived")
	}
}

// TestGatewayRerootIsSingleShotAcrossProcesses locks REQ-WRR-003-2 and
// REQ-WRR-006 (AC-WRR-008): the attempt marker is durable, so a further
// invocation — the same executor or a fresh one reading only the record —
// removes nothing and surfaces the guidance unchanged.
func TestGatewayRerootIsSingleShotAcrossProcesses(t *testing.T) {
	transcript := filepath.Join(t.TempDir(), "u.jsonl")
	rows := []any{
		rerootUserRow("/p", "hello"),
		rerootAssistantRow("/p", []any{map[string]any{"type": "text", "text": "one"}}),
		rerootAssistantRow("/p", []any{map[string]any{"type": "text", "text": "never published"}}),
	}
	rerootWriteTranscript(t, transcript, rows)
	first, err := rerootGatewayTranscript(transcript)
	if err != nil {
		t.Fatal(err)
	}
	if first.Removed == 0 {
		t.Fatal("first attempt removed nothing")
	}
	recovered, err := os.ReadFile(transcript)
	if err != nil {
		t.Fatal(err)
	}
	// Same-process re-invocation: the incident is exhausted by the marker.
	again, err := rerootGatewayTranscript(transcript)
	if err != nil {
		t.Fatal(err)
	}
	if again.Removed != 0 || again.Guidance == "" {
		t.Fatalf("second outcome=%+v", again)
	}
	// Fresh-process re-invocation: the executor holds no state of its own —
	// the marker on disk is the only record — so this call is exactly what a
	// new process performs when it reads the conversation record.
	fresh, err := rerootGatewayTranscript(transcript)
	if err != nil {
		t.Fatal(err)
	}
	if fresh.Guidance != again.Guidance {
		t.Fatalf("guidance not stable: %q vs %q", fresh.Guidance, again.Guidance)
	}
	now, _ := os.ReadFile(transcript)
	if string(now) != string(recovered) {
		t.Fatal("exhausted incident mutated the transcript")
	}
}

// TestGatewayRerootRefusesWithoutAssistantBoundary covers the clean-refusal
// edge: nothing to re-root, no file touched.
func TestGatewayRerootRefusesWithoutAssistantBoundary(t *testing.T) {
	transcript := filepath.Join(t.TempDir(), "u.jsonl")
	before := rerootWriteTranscript(t, transcript, []any{rerootUserRow("/p", "hello")})
	out, err := rerootGatewayTranscript(transcript)
	if err != nil {
		t.Fatal(err)
	}
	if out.Removed != 0 || out.Guidance == "" || out.Aside != "" {
		t.Fatalf("outcome=%+v", out)
	}
	now, _ := os.ReadFile(transcript)
	if string(now) != string(before) {
		t.Fatal("refusal mutated the transcript")
	}
}

// TestGatewayRerootPathHoldsNoReceiptStore is the structural half of
// AC-WRR-010: the recovery implementation takes no receipt-store argument and
// opens none — no receipt handle can exist in the recovery path's files.
func TestGatewayRerootPathHoldsNoReceiptStore(t *testing.T) {
	for _, name := range []string{"gateway_reroot.go", filepath.Join("..", "gateway", "conversation", "reroot.go")} {
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(src), "\n") {
			code := line
			if i := strings.Index(code, "//"); i >= 0 {
				code = code[:i]
			}
			for _, token := range []string{"receipt.", "OpenStore", "NewReceiptHistory", "receipt.NewGPTSubscriptionReceiptHistory"} {
				if strings.Contains(code, token) {
					t.Fatalf("%s references %s: %s", name, token, line)
				}
			}
		}
	}
}

// rerootDigest hashes every regular file under root (path + content) so a
// byte change anywhere in the receipt store flips the digest.
func rerootDigest(t *testing.T, root string) string {
	t.Helper()
	h := sha256.New()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		h.Write([]byte(rel))
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		h.Write(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// TestGatewayRerootLeavesGatewayStoreByteIdentical is the complement half of
// AC-WRR-010: with a real receipt store present, recovery leaves every store
// file byte-identical.
func TestGatewayRerootLeavesGatewayStoreByteIdentical(t *testing.T) {
	_, d, transcript, _ := rerootFixtureConversation(t, true)
	before := rerootDigest(t, d.ReceiptDir)
	out, err := rerootGatewayTranscript(transcript)
	if err != nil {
		t.Fatal(err)
	}
	if out.Removed == 0 {
		t.Fatal("recovery removed nothing")
	}
	if after := rerootDigest(t, d.ReceiptDir); after != before {
		t.Fatalf("receipt store changed: %s -> %s", before[:12], after[:12])
	}
}

// rerootFixtureConversation builds a real conversation record whose transcript
// carries the launcher-reachable wedge shape: a published end_turn boundary, a
// user turn, a phantom end_turn assistant boundary the gateway never
// published, a plain user turn, and the terminal API-error display row (the
// shape that keeps the native transcript completable). With preregister=true
// the record's transcript pointer is registered through Manager.Complete.
func rerootFixtureConversation(t *testing.T, preregister bool) (families *conversation.Manager, d conversation.Descriptor, transcript string, cwd string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	families, err = conversation.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	cwd = filepath.Join(root, "proj")
	d, err = families.New(context.Background(), conversation.NewRequest{CWD: cwd, Project: "proj"})
	if err != nil {
		t.Fatal(err)
	}
	transcript = filepath.Join(d.ConfigDir, "projects", "proj", d.UUID+".jsonl")
	// The second user turn of the full AC-WRR-009 shape is omitted here: the
	// native completion gate refuses two consecutive user rows before the
	// terminal API-error row, and this fixture is the launcher-reachable
	// wedge that must resume both before and after recovery. The full
	// user-turn-preserved shape is locked by the core recovery tests above.
	rows := []any{
		rerootUserRow(cwd, "hello"),
		rerootAssistantRow(cwd, []any{map[string]any{"type": "text", "text": "one"}}),
		rerootUserRow(cwd, "fix it"),
		rerootAssistantRow(cwd, []any{map[string]any{"type": "text", "text": "half answer"}}),
		rerootAPIErrorRow(cwd),
	}
	// Native rows carry the owning session id; transcriptModel rejects rows
	// without it, and the record's completion gate runs in the wiring tests.
	for _, row := range rows {
		row.(map[string]any)["sessionId"] = d.UUID
	}
	rerootWriteTranscript(t, transcript, rows)
	if preregister {
		if err = families.Complete(context.Background(), d.UUID, transcript, 1); err != nil {
			t.Fatal(err)
		}
	}
	return families, d, transcript, cwd
}

// TestGatewayRerootRequiresExplicitInvocation locks REQ-WRR-005 (AC-WRR-011):
// a chain-classified rejection with no user action modifies no transcript —
// only the explicit --reroot request runs the recovery, and the marker is
// stripped from the child args.
func TestGatewayRerootRequiresExplicitInvocation(t *testing.T) {
	families, d, transcript, _ := rerootFixtureConversation(t, true)
	pre, err := os.ReadFile(transcript)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = prepareGatewayConversation(gatewayLaunchRequest{
		CWD: d.CWD, Project: "proj", Args: []string{"--resume", d.UUID},
	}, families); err != nil {
		t.Fatal(err)
	}
	post, _ := os.ReadFile(transcript)
	if string(post) != string(pre) {
		t.Fatal("launch without --reroot mutated the transcript")
	}
	// The explicit invocation re-roots, then the ordinary resume proceeds.
	if _, err = prepareGatewayConversation(gatewayLaunchRequest{
		CWD: d.CWD, Project: "proj", Args: []string{"--resume", d.UUID, "--reroot"},
	}, families); err != nil {
		t.Fatal(err)
	}
	rerooted, err := os.ReadFile(transcript)
	if err != nil {
		t.Fatal(err)
	}
	if string(rerooted) == string(pre) || strings.Contains(string(rerooted), "half answer") {
		t.Fatal("--reroot did not re-root the transcript")
	}
	got := gatewayConversationPassthrough([]string{"--resume", "old", "--reroot", "--fork-session", "--session-id", "old", "--verbose"})
	if !reflect.DeepEqual(got, []string{"--verbose"}) {
		t.Fatalf("passthrough=%v", got)
	}
}

// TestGatewayRerootRefusesForkCombination keeps the explicit surface
// unambiguous: --reroot (path a, client-side) never combines with
// --fork-session (path b, gateway state).
func TestGatewayRerootRefusesForkCombination(t *testing.T) {
	families, d, _, _ := rerootFixtureConversation(t, true)
	if _, err := prepareGatewayConversation(gatewayLaunchRequest{
		CWD: d.CWD, Project: "proj", Args: []string{"--resume", d.UUID, "--reroot", "--fork-session"},
	}, families); err == nil {
		t.Fatal("--reroot combined with --fork-session accepted")
	}
}

// TestGatewayRerootForgedTailNeverRecoverable locks REQ-WRR-007 (AC-WRR-016):
// a trailing boundary carrying a forged envelope is client-side
// indistinguishable from the wedge; recovery removes it under the same
// single-shot bound, the removed content is preserved aside, no replay ever
// accepts it, and the remaining genuine published prefix is accepted by the
// unchanged check.
func TestGatewayRerootForgedTailNeverRecoverable(t *testing.T) {
	const id = "54545454-5454-8454-8454-545454545454"
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	store, err := receipt.OpenStore(context.Background(), dir, id, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	authority := translate.NewGPTSubscriptionReceiptHistory(store, id, id)

	forged, err := opaque.Encode([]opaque.Item{{OutputIndex: 0, Raw: []byte(`{"type":"reasoning","id":"rs_forged","summary":[],"encrypted_content":"forged-never-published"}`)}})
	if err != nil {
		t.Fatal(err)
	}
	marker, err := opaque.BindToolID("call_forged", forged)
	if err != nil {
		t.Fatal(err)
	}
	carrier := map[string]any{"type": "redacted_thinking", "data": forged.Data()}
	turn1 := []any{
		map[string]any{"role": "user", "content": "hello"},
		map[string]any{"role": "assistant", "content": []any{
			map[string]any{"type": "redacted_thinking", "data": mustEnvelope(t, "rs_real", "real-1").Data()},
			map[string]any{"type": "text", "text": "one"},
		}},
	}
	phantom := map[string]any{"role": "assistant", "content": []any{carrier, map[string]any{"type": "tool_use", "id": marker, "name": "write", "input": map[string]any{}}}}
	dependent := map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": marker, "content": "done"}}}
	trailingUser := map[string]any{"role": "user", "content": "continue"}

	ctx := context.Background()
	if err = authority.Publish(ctx, "gpt-5.6-sol", "owner", mustJSON(t, turn1)); err != nil {
		t.Fatal(err)
	}

	// Native transcript mirroring the replay: published turn1 + wedge tail
	// whose boundary carries the forged envelope.
	cwd := "/forged-proj"
	rows := []any{
		rerootUserRow(cwd, "hello"),
		rerootAssistantRow(cwd, turn1[1].(map[string]any)["content"].([]any)),
		rerootUserRow(cwd, "fix it"),
		rerootAssistantRow(cwd, phantom["content"].([]any)),
		rerootToolResultRow(cwd, marker),
		rerootUserRow(cwd, "continue"),
	}
	transcript := filepath.Join(t.TempDir(), "u.jsonl")
	rerootWriteTranscript(t, transcript, rows)

	out, err := rerootGatewayTranscript(transcript)
	if err != nil {
		t.Fatal(err)
	}
	if out.Removed != 2 || out.Aside == "" {
		t.Fatalf("outcome=%+v", out)
	}

	// The re-rooted remainder is a genuine published-chain prefix: accepted.
	remainder := append(append([]any{}, turn1...), map[string]any{"role": "user", "content": "fix it"}, trailingUser)
	if err = authority.Check(ctx, "gpt-5.6-sol", "owner", mustJSON(t, remainder)); err != nil {
		t.Fatalf("genuine prefix rejected after recovery: %v", err)
	}
	// Reintroducing the removed forged boundary into any replay is rejected.
	reintroduced := append(append([]any{}, remainder...), phantom, dependent)
	if err = authority.Check(ctx, "gpt-5.6-sol", "owner", mustJSON(t, reintroduced)); err == nil {
		t.Fatal("removed forged boundary accepted on replay")
	}
}

func mustEnvelope(t *testing.T, id, cipher string) *opaque.Envelope {
	t.Helper()
	envelope, err := opaque.Encode([]opaque.Item{{OutputIndex: 0, Raw: []byte(`{"type":"reasoning","id":"` + id + `","summary":[],"encrypted_content":"` + cipher + `"}`)}})
	if err != nil {
		t.Fatal(err)
	}
	return envelope
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
