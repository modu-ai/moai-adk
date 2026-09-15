package cli

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
)

// This exercises the production assembly with a local App Server protocol peer.
// It never contacts a provider or attaches to a user's gateway/native process.
func TestProductionGatewayClearStartsDistinctAttestedThread(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("local shared peer requires Unix socket")
	}
	home, err := filepath.EvalSymlinks(sharedGPTFixtureHome(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOAI_HOME", home)
	binDir, _ := installProductionWiringFakeCodex(t)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	const family = "320262b9-f176-4b4d-b72c-509adb15a34a"
	const next = "420262b9-f176-4b4d-b72c-509adb15a34a"
	const unknown = "520262b9-f176-4b4d-b72c-509adb15a34a"
	const token = "synthetic-clear-token"
	base := filepath.Join(home, "families", family)
	receiptDir := filepath.Join(base, "receipt")
	if err := os.MkdirAll(receiptDir, 0700); err != nil {
		t.Fatal(err)
	}
	ledger, err := receipt.OpenStore(context.Background(), receiptDir, family, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := ledger.Close(); err != nil {
		t.Fatal(err)
	}
	native := filepath.Join(base, "native", "projects", "project")
	if err := os.MkdirAll(native, 0700); err != nil {
		t.Fatal(err)
	}
	rows := []map[string]any{
		{"type": "user", "uuid": "clear-command", "sessionId": next, "cwd": home, "isSidechain": false, "message": map[string]string{"role": "user", "content": "<command-name>/clear</command-name> <command-message>clear</command-message> <command-args></command-args>"}},
		{"type": "system", "subtype": "local_command", "parentUuid": "clear-command", "sessionId": next, "cwd": home, "isSidechain": false, "content": "<local-command-stdout></local-command-stdout>"},
	}
	var transcript []byte
	for _, row := range rows {
		raw, e := json.Marshal(row)
		if e != nil {
			t.Fatal(e)
		}
		transcript = append(transcript, raw...)
		transcript = append(transcript, '\n')
	}
	if err := os.WriteFile(filepath.Join(native, next+".jsonl"), transcript, 0600); err != nil {
		t.Fatal(err)
	}
	peer := startClaudeGatewayPeer(t, home)
	payload, err := json.Marshal(gatewayPrivatePayload{Version: 1, SessionToken: token, ModelIDs: []string{"gpt-5.6-sol"}, Conversation: &gatewayPrivateConversation{FamilyID: family, SessionID: family, ReceiptDir: receiptDir, CWD: home}})
	if err != nil {
		t.Fatal(err)
	}
	handler, err := productionGatewayHandlerFactory(payload)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := handler.(io.Closer).Close(); err != nil {
			t.Error(err)
		}
	})
	send := func(session, header string) *httptest.ResponseRecorder {
		identity, e := json.Marshal(map[string]string{"session_id": session, "device_id": "offline-device", "account_uuid": ""})
		if e != nil {
			t.Fatal(e)
		}
		body, e := json.Marshal(map[string]any{"model": "gpt-5.6-sol", "max_tokens": 100, "stream": true, "system": "MOAI_CHILD_PROBE: reply CHILD_GATEWAY_OK", "metadata": map[string]string{"user_id": string(identity)}, "messages": []any{map[string]string{"role": "user", "content": "reply"}}})
		if e != nil {
			t.Fatal(e)
		}
		request := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body)))
		request.Header.Set("Authorization", "Bearer "+token)
		request.Header.Set("X-Claude-Code-Session-Id", header)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response
	}
	for _, session := range []string{family, next} {
		response := send(session, session)
		if response.Code != 200 || !strings.Contains(response.Body.String(), "CHILD_GATEWAY_OK") {
			t.Fatalf("session %s: status=%d body=%s", session, response.Code, response.Body.String())
		}
	}
	peer.mu.Lock()
	threads, turns := peer.threads, peer.children
	peer.mu.Unlock()
	if threads != 2 || turns != 2 {
		t.Fatalf("clear did not create two independent model threads: threads=%d turns=%d", threads, turns)
	}
	for _, ids := range [][2]string{{unknown, unknown}, {next, family}} {
		response := send(ids[0], ids[1])
		if response.Code != 400 {
			t.Fatalf("unattested or conflicting session: status=%d body=%s", response.Code, response.Body.String())
		}
	}
	peer.mu.Lock()
	afterThreads, afterTurns := peer.threads, peer.children
	peer.mu.Unlock()
	if afterThreads != threads || afterTurns != turns {
		t.Fatalf("rejected identity reached model execution: threads=%d turns=%d", afterThreads, afterTurns)
	}
	t.Logf("old+clear HTTP200; independent threads=%d turns=%d; unattested+conflicting HTTP400 with zero additional model execution", threads, turns)
}
