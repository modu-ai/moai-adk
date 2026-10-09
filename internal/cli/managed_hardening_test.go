package cli

// managed_hardening_test.go — SPEC-FACTORY-MANAGED-HARDEN-001 M1 (card t1409):
// the RED baseline for F3 (server-originated Codex requests) and F4 (turn-level
// failure isolation). Every test here compiles against the BASE API — no new
// production symbol — and is expected to fail on the base tree for the reason
// acceptance.md §2.1 states; M2/M3 flip them green.
//
// Fake server shape: the test goroutine IS the App Server. newHardenPair joins
// a real managedCodexAppClient to an in-process loopback WebSocket whose server
// end the test drives frame by frame, which keeps every ordering deterministic.
// Time bounds follow the acceptance.md header convention: a blocking call runs
// in a goroutine and the test fails it at an absolute 5 s watchdog; cleanup
// closes the connection so a blocked call is released and nothing leaks.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const (
	// hardenWatchdog is the absolute bound a blocking in-process call gets.
	hardenWatchdog = 5 * time.Second
	// hardenChildWatchdog bounds a driver run over a real fake-app-server child
	// (process start and loopback round trips).
	hardenChildWatchdog = 15 * time.Second
	// hardenReplyWait is how long the policy test waits for the owner's answers
	// before it records that none came.
	hardenReplyWait = 2 * time.Second
	// hardenQuietWindow is the "exactly once" window: one polling tick of the
	// repo itself (DefaultManagedSessionPollInterval = 500ms).
	hardenQuietWindow = 500 * time.Millisecond
	// hardenLogPoll and hardenLogWait bound every log-line assertion: a polling
	// wait, never a read that depends on the order of log line and answer.
	hardenLogPoll = 10 * time.Millisecond
	hardenLogWait = 5 * time.Second
)

// hardenLogSink is the writer a test stores in managedLogOutput. Writes and
// snapshot reads share one mutex, so a reader goroutine logging while the test
// polls is race-free; tests never touch the raw buffer.
type hardenLogSink struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *hardenLogSink) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *hardenLogSink) snapshot() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

// hardenLogCapture points the owner's log output at a fresh synchronized sink
// and restores the default on cleanup. A test that also starts a client
// registers this FIRST (cleanups run LIFO), so the client is shut down and its
// read goroutine awaited before the pointer is restored.
func hardenLogCapture(t *testing.T) *hardenLogSink {
	t.Helper()
	sink := &hardenLogSink{}
	var sinkWriter io.Writer = sink
	managedLogOutput.Store(&sinkWriter)
	t.Cleanup(func() { managedLogOutput.Store(nil) })
	return sink
}

// waitFor polls until the sink holds the substring or hardenLogWait elapses.
func (s *hardenLogSink) waitFor(substr string) bool {
	deadline := time.Now().Add(hardenLogWait)
	for {
		if strings.Contains(s.snapshot(), substr) {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(hardenLogPoll)
	}
}

// hardenFrame is one frame the client wrote, as the fake server reads it. ID is
// kept raw so integer and string ids both survive and replies compare byte-wise.
type hardenFrame struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	// Raw is the frame exactly as it crossed the wire.
	Raw json.RawMessage `json:"-"`
}

// isReply reports a client answer to a server request: an id and no method.
func (f hardenFrame) isReply() bool { return f.Method == "" && len(f.ID) > 0 }

// hardenServer is the server end of a hardenPair.
type hardenServer struct {
	t      *testing.T
	conn   *websocket.Conn
	frames chan hardenFrame
	log    *hardenLogSink
}

// newHardenPair joins a started managedCodexAppClient (read goroutine running)
// to the server end the test drives, and points the owner's log output at a
// fresh synchronized sink (srv.log). Cleanup order is LIFO: the client is shut
// down and its read goroutine awaited first, then the server connection and
// listener close, and only then is the log output pointer restored — so no
// goroutine of this pair can write to a later test's sink.
func newHardenPair(t *testing.T) (*managedCodexAppClient, *hardenServer) {
	t.Helper()
	sink := hardenLogCapture(t)
	accepted := make(chan *websocket.Conn, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		accepted <- conn
	}))
	t.Cleanup(ts.Close)

	conn, _, err := managedCodexDial(strings.Replace(ts.URL, "http://", "ws://", 1), nil)
	if err != nil {
		t.Fatalf("dial the in-process fake app server: %v", err)
	}
	var serverConn *websocket.Conn
	select {
	case serverConn = <-accepted:
	case <-time.After(hardenWatchdog):
		_ = conn.Close()
		t.Fatalf("fake app server did not accept within %s", hardenWatchdog)
	}
	t.Cleanup(func() { _ = serverConn.Close() })

	server := &hardenServer{t: t, conn: serverConn, frames: make(chan hardenFrame, 256), log: sink}
	go server.readFrames()

	client := &managedCodexAppClient{conn: conn, events: make(chan managedCodexAppReply, 32), done: make(chan struct{})}
	go client.read()
	t.Cleanup(func() {
		client.shutdown()
		drained := make(chan struct{})
		go func() {
			for range client.events {
			}
			close(drained)
		}()
		select {
		case <-drained:
		case <-time.After(hardenWatchdog):
			t.Error("client read goroutine did not stop after shutdown")
		}
	})
	return client, server
}

func (s *hardenServer) readFrames() {
	defer close(s.frames)
	for {
		_, raw, err := s.conn.ReadMessage()
		if err != nil {
			return
		}
		frame, err := hardenParseFrame(raw)
		if err != nil {
			return
		}
		s.frames <- frame
	}
}

// hardenParseFrame decodes one wire frame, keeping the wire bytes in Raw.
func hardenParseFrame(raw []byte) (hardenFrame, error) {
	var frame hardenFrame
	if err := json.Unmarshal(raw, &frame); err != nil {
		return hardenFrame{}, err
	}
	frame.Raw = append([]byte(nil), raw...)
	return frame, nil
}

func (s *hardenServer) send(msg map[string]any) {
	s.t.Helper()
	if err := s.conn.WriteJSON(msg); err != nil {
		s.t.Errorf("fake server write: %v", err)
	}
}

// awaitMethod returns the next client frame carrying the method, skipping the
// frames before it.
func (s *hardenServer) awaitMethod(method string) hardenFrame {
	s.t.Helper()
	timer := time.After(hardenWatchdog)
	for {
		select {
		case frame, ok := <-s.frames:
			if !ok {
				s.t.Fatalf("connection closed while the fake server waited for %s", method)
			}
			if frame.Method == method {
				return frame
			}
		case <-timer:
			s.t.Fatalf("fake server saw no %s within %s", method, hardenWatchdog)
		}
	}
}

// awaitReply waits for the client's answer to the server request carrying the
// raw id.
func (s *hardenServer) awaitReply(id string, wait time.Duration) (hardenFrame, bool) {
	timer := time.After(wait)
	for {
		select {
		case frame, ok := <-s.frames:
			if !ok {
				return hardenFrame{}, false
			}
			if frame.isReply() && string(frame.ID) == id {
				return frame, true
			}
		case <-timer:
			return hardenFrame{}, false
		}
	}
}

// collectReplies gathers client answers until `want` distinct ids arrived and a
// quiet window passed with no further reply, or until hardenReplyWait elapsed.
// duplicates counts answers repeating an id already seen.
func (s *hardenServer) collectReplies(want int) (replies map[string]hardenFrame, duplicates int) {
	replies = map[string]hardenFrame{}
	overall := time.NewTimer(hardenReplyWait)
	defer overall.Stop()
	var quiet <-chan time.Time
	for {
		select {
		case frame, ok := <-s.frames:
			if !ok {
				return replies, duplicates
			}
			if !frame.isReply() {
				continue
			}
			if _, seen := replies[string(frame.ID)]; seen {
				duplicates++
			} else {
				replies[string(frame.ID)] = frame
			}
			if len(replies) >= want {
				quiet = time.After(hardenQuietWindow)
			}
		case <-quiet:
			return replies, duplicates
		case <-overall.C:
			return replies, duplicates
		}
	}
}

// startTurn answers the client's turn/start request with turn id T and emits
// no lifecycle notification; the caller scripts the rest of the turn.
func (s *hardenServer) startTurn(turnID string) json.RawMessage {
	s.t.Helper()
	frame := s.awaitMethod("turn/start")
	s.send(map[string]any{"id": frame.ID, "result": map[string]any{"turn": map[string]string{"id": turnID}}})
	return frame.ID
}

func (s *hardenServer) notify(method string, params any) {
	s.t.Helper()
	s.send(map[string]any{"method": method, "params": params})
}

func (s *hardenServer) turnStarted(turnID string) {
	s.t.Helper()
	s.notify("turn/started", map[string]any{"turn": map[string]string{"id": turnID}})
}

func (s *hardenServer) turnCompleted(turnID, status string) {
	s.t.Helper()
	s.notify("turn/completed", map[string]any{"turn": map[string]string{"id": turnID, "status": status}})
}

// request sends a server-originated request: an id AND a method.
func (s *hardenServer) request(id any, method string, params any) {
	s.t.Helper()
	s.send(map[string]any{"id": id, "method": method, "params": params})
}

// hardenDeliver runs DeliverTurn in a goroutine so the caller can bound it.
func hardenDeliver(sess *managedCodexSession, prompt string) <-chan error {
	done := make(chan error, 1)
	go func() { done <- sess.DeliverTurn(prompt) }()
	return done
}

func hardenJSONEqual(a, b []byte) bool {
	var x, y any
	if json.Unmarshal(a, &x) != nil || json.Unmarshal(b, &y) != nil {
		return false
	}
	return reflect.DeepEqual(x, y)
}

func hardenWantResult(want string) func(*testing.T, hardenFrame) {
	return func(t *testing.T, f hardenFrame) {
		t.Helper()
		if f.Error != nil {
			t.Fatalf("answered with error %+v, want result %s", *f.Error, want)
		}
		if !hardenJSONEqual(f.Result, []byte(want)) {
			t.Fatalf("result %s, want %s", f.Result, want)
		}
	}
}

// hardenWantLegacyDenied pins the legacy approval refusal shape: codex 0.161.0
// ReviewDecision carries a denial as {"denied":{"rejection":<string>}}, not as
// the bare string "denied". The reply is decoded and its shape asserted, with
// the rejection text compared to the production constant.
func hardenWantLegacyDenied() func(*testing.T, hardenFrame) {
	return func(t *testing.T, f hardenFrame) {
		t.Helper()
		if f.Error != nil {
			t.Fatalf("answered with error %+v, want a denied decision result", *f.Error)
		}
		var result struct {
			Decision struct {
				Denied *struct {
					Rejection *string `json:"rejection"`
				} `json:"denied"`
			} `json:"decision"`
		}
		if err := json.Unmarshal(f.Result, &result); err != nil {
			t.Fatalf("result %s is not the {decision:{denied:{rejection}}} object: %v", f.Result, err)
		}
		if result.Decision.Denied == nil || result.Decision.Denied.Rejection == nil {
			t.Fatalf("result %s, want {\"decision\":{\"denied\":{\"rejection\":<string>}}}", f.Result)
		}
		if got := *result.Decision.Denied.Rejection; got == "" || got != managedLegacyApprovalRejection {
			t.Fatalf("rejection %q, want the non-empty constant %q", got, managedLegacyApprovalRejection)
		}
	}
}

func hardenWantError(code int) func(*testing.T, hardenFrame) {
	return func(t *testing.T, f hardenFrame) {
		t.Helper()
		if f.Error == nil {
			t.Fatalf("answered with result %s, want a JSON-RPC error %d", f.Result, code)
		}
		if f.Error.Code != code {
			t.Fatalf("error code %d (%q), want %d", f.Error.Code, f.Error.Message, code)
		}
	}
}

// hardenPolicyRow is one design.md D-1 response-policy row.
type hardenPolicyRow struct {
	name    string
	method  string
	params  map[string]any
	check   func(*testing.T, hardenFrame)
	outcome string // the log token after "->"
	logTail string // what follows the token on the log line (elicitation only)
}

// hardenPolicyRows is the design.md D-1 response policy as data: ten listed
// methods plus one the table does not name. No row may answer with an accept
// family decision or a non-empty permission grant.
func hardenPolicyRows() []hardenPolicyRow {
	base := func(extra map[string]any) map[string]any {
		params := map[string]any{"threadId": "th-policy", "turnId": "T1", "itemId": "item-1"}
		for k, v := range extra {
			params[k] = v
		}
		return params
	}
	return []hardenPolicyRow{
		{"command_execution_approval", "item/commandExecution/requestApproval", base(nil), hardenWantResult(`{"decision":"decline"}`), "decline", ""},
		{"file_change_approval", "item/fileChange/requestApproval", base(nil), hardenWantResult(`{"decision":"decline"}`), "decline", ""},
		{"permissions_approval", "item/permissions/requestApproval", base(nil), hardenWantResult(`{"permissions":{}}`), "empty", ""},
		{"mcp_elicitation", "mcpServer/elicitation/request", base(map[string]any{"serverName": "other-server", "message": "need input"}), hardenWantResult(`{"action":"decline"}`), "decline", ` serverName="other-server" turn=none broker_declined=0`},
		{"tool_request_user_input", "item/tool/requestUserInput", base(nil), hardenWantError(-32000), "error", ""},
		{"dynamic_tool_call", "item/tool/call", base(map[string]any{"tool": "t", "arguments": map[string]any{}}), func(t *testing.T, f hardenFrame) {
			t.Helper()
			if f.Error != nil {
				t.Fatalf("answered with error %+v, want a failed tool result", *f.Error)
			}
			var result struct {
				Success      *bool `json:"success"`
				ContentItems []struct {
					Type string `json:"type"`
				} `json:"contentItems"`
			}
			if err := json.Unmarshal(f.Result, &result); err != nil {
				t.Fatalf("result %s is not a tool result: %v", f.Result, err)
			}
			if result.Success == nil || *result.Success {
				t.Fatalf("result %s, want success:false", f.Result)
			}
			if len(result.ContentItems) == 0 || result.ContentItems[0].Type != "inputText" {
				t.Fatalf("result %s, want an inputText content item", f.Result)
			}
		}, "failed", ""},
		{"chatgpt_token_refresh", "account/chatgptAuthTokens/refresh", base(nil), hardenWantError(-32000), "error", ""},
		{"attestation_generate", "attestation/generate", base(nil), hardenWantError(-32000), "error", ""},
		{"legacy_apply_patch_approval", "applyPatchApproval", base(nil), hardenWantLegacyDenied(), "denied", ""},
		{"legacy_exec_command_approval", "execCommandApproval", base(nil), hardenWantLegacyDenied(), "denied", ""},
		{"unknown_method", "item/unlisted/needsAnswer", base(nil), hardenWantError(-32601), "error", ""},
	}
}

// TestManagedCodexServerRequestPolicy is AC-MH-001: during one turn the fake
// server sends ten listed server requests and one unlisted method, each with an
// integer id. The owner must answer each exactly once, with the same id, per
// the D-1 policy table, and leave one log line per answered request carrying
// the quoted method (the elicitation line also carries serverName, turn and
// broker_declined). The M1 form of this test had no log seam; M2 added it, and
// on the base owner (no answers) the fake server records the absence after
// hardenReplyWait.
func TestManagedCodexServerRequestPolicy(t *testing.T) {
	client, srv := newHardenPair(t)
	sess := &managedCodexSession{client: client, threadID: "th-policy"}
	rows := hardenPolicyRows()

	done := hardenDeliver(sess, "policy turn")
	srv.startTurn("T1")
	srv.turnStarted("T1")
	ids := make([]string, len(rows))
	for i, row := range rows {
		id := 101 + i
		ids[i] = strconv.Itoa(id)
		srv.request(id, row.method, row.params)
	}
	replies, duplicates := srv.collectReplies(len(rows))
	srv.turnCompleted("T1", "completed")
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("DeliverTurn = %v, want nil (the policy never ends a turn)", err)
		}
	case <-time.After(hardenWatchdog):
		t.Fatalf("DeliverTurn did not return within %s after the turn completed", hardenWatchdog)
	}

	if duplicates != 0 {
		t.Errorf("%d duplicate answers within the %s quiet window, want each request answered exactly once", duplicates, hardenQuietWindow)
	}
	for id, reply := range replies {
		lowered := strings.ToLower(string(reply.Result))
		if strings.Contains(lowered, "accept") || strings.Contains(lowered, "approved") {
			t.Errorf("answer to request %s grants approval: %s", id, reply.Result)
		}
	}
	for i, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			reply, ok := replies[ids[i]]
			if !ok {
				t.Fatalf("no answer to server request %s (%s) within %s: the owner does not answer server-originated requests", ids[i], row.method, hardenReplyWait)
			}
			row.check(t, reply)
			wantLine := fmt.Sprintf("Factory server request answered: %q -> %s%s\n", row.method, row.outcome, row.logTail)
			if !srv.log.waitFor(wantLine) {
				t.Fatalf("log has no line %q within %s; log:\n%s", strings.TrimSuffix(wantLine, "\n"), hardenLogWait, srv.log.snapshot())
			}
			prefix := fmt.Sprintf("Factory server request answered: %q ->", row.method)
			if n := strings.Count(srv.log.snapshot(), prefix); n != 1 {
				t.Fatalf("%d log lines for %s, want exactly one", n, row.method)
			}
		})
	}
}

// hardenResultSchemas maps every server-request kind the owner answers with a
// result to the vendored codex 0.161.0 response schema that result must satisfy.
// A kind absent here is answered with a JSON-RPC error, validated against
// JSONRPCError.json instead.
var hardenResultSchemas = map[string]string{
	"item/commandExecution/requestApproval": "CommandExecutionRequestApprovalResponse.json",
	"item/fileChange/requestApproval":       "FileChangeRequestApprovalResponse.json",
	"item/permissions/requestApproval":      "PermissionsRequestApprovalResponse.json",
	managedElicitationMethod:                "McpServerElicitationRequestResponse.json",
	"item/tool/call":                        "DynamicToolCallResponse.json",
	"applyPatchApproval":                    "ApplyPatchApprovalResponse.json",
	"execCommandApproval":                   "ExecCommandApprovalResponse.json",
}

// hardenCompileSchema compiles one vendored draft-07 schema file. The files
// carry only internal "#/definitions" references, so no loader is needed.
func hardenCompileSchema(t *testing.T, file string) *jsonschema.Schema {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "codex-0.161.0", file))
	if err != nil {
		t.Fatalf("read vendored schema %s: %v", file, err)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("parse vendored schema %s: %v", file, err)
	}
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft7)
	url := "https://moai.invalid/codex-0.161.0/" + file
	if err := compiler.AddResource(url, doc); err != nil {
		t.Fatalf("add vendored schema %s: %v", file, err)
	}
	schema, err := compiler.Compile(url)
	if err != nil {
		t.Fatalf("compile vendored schema %s: %v", file, err)
	}
	return schema
}

// hardenValidate validates raw JSON against a compiled schema.
func hardenValidate(t *testing.T, schema *jsonschema.Schema, raw []byte) error {
	t.Helper()
	value, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("answer %s is not JSON: %v", raw, err)
	}
	return schema.Validate(value)
}

// hardenErrorFrameBytes returns the bytes the schema guard validates for an
// error-answered server request: the frame exactly as it crossed the wire.
// Re-assembling it from the decoded struct would supply a missing field (an
// absent "message" decodes to "" and re-encodes as present), hiding the very
// defect the guard exists for.
func hardenErrorFrameBytes(t *testing.T, reply hardenFrame) []byte {
	t.Helper()
	if len(reply.Raw) == 0 {
		t.Fatalf("reply %s carries no wire bytes", reply.ID)
	}
	return reply.Raw
}

// TestManagedSchemaGuardSeesMessagelessErrorFrame pins that the schema guard
// validates the error frame as it crossed the wire: a frame whose error object
// carries no "message" must be rejected by JSONRPCError.json, not repaired by
// re-assembly into one that has an empty message.
func TestManagedSchemaGuardSeesMessagelessErrorFrame(t *testing.T) {
	rpcError := hardenCompileSchema(t, "JSONRPCError.json")
	for _, wire := range []string{
		`{"id":7,"error":{"code":-32601}}`,
		`{"id":"7","jsonrpc":"2.0","error":{"code":-32601}}`,
	} {
		frame, err := hardenParseFrame([]byte(wire))
		if err != nil {
			t.Fatalf("parse %s: %v", wire, err)
		}
		if !frame.isReply() || frame.Error == nil {
			t.Fatalf("%s did not parse as an error reply", wire)
		}
		if hardenValidate(t, rpcError, hardenErrorFrameBytes(t, frame)) == nil {
			t.Errorf("wire frame %s lacks error.message yet the guard accepts it", wire)
		}
	}
	// Positive control: a well-formed wire frame still passes.
	ok, err := hardenParseFrame([]byte(`{"id":7,"error":{"code":-32601,"message":"m"}}`))
	if err != nil {
		t.Fatalf("parse control: %v", err)
	}
	if err := hardenValidate(t, rpcError, hardenErrorFrameBytes(t, ok)); err != nil {
		t.Errorf("well-formed error frame rejected: %v", err)
	}
}

// TestManagedServerRequestPolicyMatchesCodexSchema is the schema-conformance
// guard for the D-1 response policy: the owner's real wire answer to every kind
// of server request must validate against the codex 0.161.0 response schema for
// that kind (or, for the error-answered kinds and the unknown-method fallback,
// against JSONRPCError). It complements, and does not replace, the exact
// answers pinned by TestManagedCodexServerRequestPolicy: a schema-valid answer
// can still be the wrong policy (abort is valid and interrupts the turn), and a
// policy-correct-by-name answer can be schema-invalid (the legacy refusal sent
// as the bare string "denied"). It needs neither the codex binary nor a network.
func TestManagedServerRequestPolicyMatchesCodexSchema(t *testing.T) {
	rows := hardenPolicyRows()
	rowMethods := map[string]bool{}
	for _, row := range rows {
		rowMethods[row.method] = true
	}
	for method, policy := range managedServerRequestPolicies {
		if !rowMethods[method] {
			t.Errorf("policy table kind %q has no row in hardenPolicyRows: the schema guard would skip it", method)
		}
		if _, hasSchema := hardenResultSchemas[method]; hasSchema != (policy.result != nil) {
			t.Errorf("policy table kind %q: result-answered=%v but vendored result schema mapped=%v", method, policy.result != nil, hasSchema)
		}
	}
	for method := range hardenResultSchemas {
		if _, ok := managedServerRequestPolicies[method]; !ok {
			t.Errorf("hardenResultSchemas names %q, which the policy table does not", method)
		}
	}
	rpcError := hardenCompileSchema(t, "JSONRPCError.json")

	t.Run("validator_catches_the_bare_denied_string", func(t *testing.T) {
		for _, file := range []string{"ExecCommandApprovalResponse.json", "ApplyPatchApprovalResponse.json"} {
			schema := hardenCompileSchema(t, file)
			if hardenValidate(t, schema, []byte(`{"decision":"denied"}`)) == nil {
				t.Errorf("%s accepts the bare string \"denied\": the guard cannot see the defect it exists for", file)
			}
			if err := hardenValidate(t, schema, []byte(`{"decision":{"denied":{"rejection":"x"}}}`)); err != nil {
				t.Errorf("%s rejects the denied object: %v", file, err)
			}
		}
	})

	client, srv := newHardenPair(t)
	sess := &managedCodexSession{client: client, threadID: "th-policy"}
	done := hardenDeliver(sess, "schema turn")
	srv.startTurn("T1")
	srv.turnStarted("T1")
	ids := make([]string, len(rows))
	for i, row := range rows {
		id := 101 + i
		ids[i] = strconv.Itoa(id)
		srv.request(id, row.method, row.params)
	}
	replies, _ := srv.collectReplies(len(rows))
	srv.turnCompleted("T1", "completed")
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("DeliverTurn = %v, want nil", err)
		}
	case <-time.After(hardenWatchdog):
		t.Fatalf("DeliverTurn did not return within %s after the turn completed", hardenWatchdog)
	}

	for i, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			reply, ok := replies[ids[i]]
			if !ok {
				t.Fatalf("no answer to server request %s (%s) within %s", ids[i], row.method, hardenReplyWait)
			}
			file, resultKind := hardenResultSchemas[row.method]
			if !resultKind {
				if reply.Error == nil {
					t.Fatalf("%s answered with result %s, want a JSON-RPC error", row.method, reply.Result)
				}
				frame := hardenErrorFrameBytes(t, reply)
				if err := hardenValidate(t, rpcError, frame); err != nil {
					t.Fatalf("error frame %s violates JSONRPCError.json: %v", frame, err)
				}
				return
			}
			if reply.Error != nil {
				t.Fatalf("%s answered with error %+v, want a result", row.method, *reply.Error)
			}
			if err := hardenValidate(t, hardenCompileSchema(t, file), reply.Result); err != nil {
				t.Fatalf("answer %s to %s violates %s: %v", reply.Result, row.method, file, err)
			}
		})
	}
}

// TestManagedCodexTurnSurvivesServerRequest is AC-MH-002: a turn that waits on
// the owner's answer to a server request must complete, and a request that
// arrives while no turn is running must still be answered. The fake server
// completes the turn only after it sees the answer, so on the base tree
// DeliverTurn is still blocked when the test's own watchdog fires.
func TestManagedCodexTurnSurvivesServerRequest(t *testing.T) {
	t.Run("during_turn", func(t *testing.T) {
		client, srv := newHardenPair(t)
		sess := &managedCodexSession{client: client, threadID: "th-during"}
		done := hardenDeliver(sess, "turn with a pending approval")
		srv.startTurn("T1")
		srv.turnStarted("T1")
		deadline := time.After(hardenWatchdog)
		srv.request(201, "item/commandExecution/requestApproval", map[string]any{"threadId": "th-during", "turnId": "T1", "itemId": "i"})
		replied := make(chan struct{})
		go func() {
			// The reader goroutine of hardenServer owns srv.frames; only this
			// goroutine consumes it until it closes `replied`.
			if _, ok := srv.awaitReply("201", hardenWatchdog); ok {
				close(replied)
			}
		}()
		select {
		case <-replied:
			srv.turnCompleted("T1", "completed")
		case <-deadline:
			t.Fatalf("DeliverTurn still blocked after %s: the server request was never answered, so the server never completed the turn", hardenWatchdog)
		}
		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("DeliverTurn = %v, want nil", err)
			}
		case <-time.After(hardenWatchdog):
			t.Fatalf("DeliverTurn did not return within %s after the answer arrived", hardenWatchdog)
		}
	})
	t.Run("between_turns", func(t *testing.T) {
		client, srv := newHardenPair(t)
		sess := &managedCodexSession{client: client, threadID: "th-between"}
		done := hardenDeliver(sess, "ordinary turn")
		srv.startTurn("T1")
		srv.turnStarted("T1")
		srv.turnCompleted("T1", "completed")
		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("first turn = %v, want nil", err)
			}
		case <-time.After(hardenWatchdog):
			t.Fatalf("first turn did not return within %s", hardenWatchdog)
		}
		// No turn is running now: the request must be answered anyway.
		srv.request(301, "item/commandExecution/requestApproval", map[string]any{"threadId": "th-between", "turnId": "T1", "itemId": "i"})
		reply, ok := srv.awaitReply("301", hardenWatchdog)
		if !ok {
			t.Fatalf("a server request arriving between turns got no answer within %s", hardenWatchdog)
		}
		if reply.Error != nil {
			t.Fatalf("answered with error %+v, want a result", *reply.Error)
		}
	})
}

// TestManagedCodexServerRequestIDCollision is AC-MH-003: a server request whose
// id equals the id of the client's pending call, or is a string, must not be
// taken for the response. On the base tree the integer collision swallows the
// genuine response (call returns an empty result with no error) and the string
// id kills the read loop (connection closed).
func TestManagedCodexServerRequestIDCollision(t *testing.T) {
	type callResult struct {
		raw json.RawMessage
		err error
	}
	runCall := func(client *managedCodexAppClient, method string) <-chan callResult {
		out := make(chan callResult, 1)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*hardenWatchdog)
			defer cancel()
			raw, err := client.call(ctx, method, map[string]any{"threadId": "th-collision"})
			out <- callResult{raw: raw, err: err}
		}()
		return out
	}
	await := func(t *testing.T, out <-chan callResult, what string) callResult {
		t.Helper()
		select {
		case res := <-out:
			return res
		case <-time.After(hardenWatchdog):
			t.Fatalf("%s did not return within %s", what, hardenWatchdog)
			return callResult{}
		}
	}
	const genuine = `{"genuine":true}`

	t.Run("integer_collision", func(t *testing.T) {
		client, srv := newHardenPair(t)
		out := runCall(client, "turn/start")
		pending := srv.awaitMethod("turn/start")
		// A server request carrying the SAME id as the pending client call,
		// then the genuine response.
		srv.request(pending.ID, "item/commandExecution/requestApproval", map[string]any{"threadId": "th-collision", "turnId": "T1", "itemId": "i"})
		srv.send(map[string]any{"id": pending.ID, "result": map[string]any{"genuine": true}})
		res := await(t, out, "call")
		if res.err != nil {
			t.Fatalf("call error = %v, want the genuine response", res.err)
		}
		if !hardenJSONEqual(res.raw, []byte(genuine)) {
			t.Fatalf("call result %q, want the genuine response %s (the server request was taken for the response)", string(res.raw), genuine)
		}
		if _, ok := srv.awaitReply(string(pending.ID), hardenWatchdog); !ok {
			t.Fatalf("the colliding server request got no answer within %s", hardenWatchdog)
		}
	})
	t.Run("string_id", func(t *testing.T) {
		client, srv := newHardenPair(t)
		out := runCall(client, "turn/start")
		pending := srv.awaitMethod("turn/start")
		srv.request("srv-1", "item/commandExecution/requestApproval", map[string]any{"threadId": "th-collision", "turnId": "T1", "itemId": "i"})
		srv.send(map[string]any{"id": pending.ID, "result": map[string]any{"genuine": true}})
		res := await(t, out, "call")
		if res.err != nil {
			t.Fatalf("call error = %v, want the genuine response (a string-id request must not end the read loop)", res.err)
		}
		if !hardenJSONEqual(res.raw, []byte(genuine)) {
			t.Fatalf("call result %q, want %s", string(res.raw), genuine)
		}
		reply, ok := srv.awaitReply(`"srv-1"`, hardenWatchdog)
		if !ok {
			t.Fatalf("the string-id server request got no answer within %s", hardenWatchdog)
		}
		if string(reply.ID) != `"srv-1"` {
			t.Fatalf("answer id %s, want the string id echoed back", reply.ID)
		}
		// The read loop is still alive: a later call succeeds.
		next := runCall(client, "thread/name/set")
		follow := srv.awaitMethod("thread/name/set")
		srv.send(map[string]any{"id": follow.ID, "result": map[string]any{}})
		if res := await(t, next, "second call"); res.err != nil {
			t.Fatalf("second call = %v, want success after the string-id request", res.err)
		}
	})
}

// TestManagedCodexConcurrentWrites is AC-MH-004: the read goroutine writes the
// answers to server requests while another goroutine's call() writes its own
// frame. A connection write that is not serialized shows up as a data race
// report under -race or as gorilla's concurrent-write panic. Each round fires
// five server requests together with one client call so the writes overlap,
// and ends only when all five answers and the call's response have arrived. The
// round count is the repeat that turns a schedule-dependent race into a likely
// one (acceptance.md AC-MH-004).
func TestManagedCodexConcurrentWrites(t *testing.T) {
	const (
		rounds           = 200
		requestsPerRound = 5
	)
	client, srv := newHardenPair(t)
	for round := 0; round < rounds; round++ {
		callDone := make(chan error, 1)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), hardenWatchdog)
			defer cancel()
			_, err := client.call(ctx, "thread/name/set", map[string]any{"round": round})
			callDone <- err
		}()
		for k := 0; k < requestsPerRound; k++ {
			srv.request(10000+round*requestsPerRound+k, "item/commandExecution/requestApproval",
				map[string]any{"threadId": "th-race", "turnId": "T1", "itemId": "i"})
		}
		answered, responded := 0, false
		timer := time.After(hardenWatchdog)
		for answered < requestsPerRound || !responded {
			select {
			case frame, ok := <-srv.frames:
				if !ok {
					t.Fatalf("round %d: connection closed with %d of %d answers", round, answered, requestsPerRound)
				}
				switch {
				case frame.isReply():
					answered++
				case frame.Method == "thread/name/set":
					srv.send(map[string]any{"id": frame.ID, "result": map[string]any{}})
					responded = true
				}
			case <-timer:
				t.Fatalf("round %d: %d of %d answers and call frame seen=%v within %s", round, answered, requestsPerRound, responded, hardenWatchdog)
			}
		}
		select {
		case err := <-callDone:
			if err != nil {
				t.Fatalf("round %d: call = %v, want the response to arrive", round, err)
			}
		case <-time.After(hardenWatchdog):
			t.Fatalf("round %d: call did not return within %s", round, hardenWatchdog)
		}
	}
}

// TestManagedCodexCompletionEventCarriesBrokerVerdict checks the read
// goroutine's turn window at client level (design.md D-1 decision 3): the
// verdict that a declined MoAI broker elicitation belongs to the turn rides on
// that turn's completion event, a request from any other server does not count,
// and a request that arrives after the window closed is logged without a turn
// and does not leak into the next one. Turning the verdict into a marked
// DeliverTurn error is M3.
func TestManagedCodexCompletionEventCarriesBrokerVerdict(t *testing.T) {
	elicitation := func(serverName string) map[string]any {
		return map[string]any{"threadId": "th-verdict", "turnId": nil, "serverName": serverName, "message": "approve?"}
	}
	completion := func(t *testing.T, client *managedCodexAppClient) managedCodexAppReply {
		t.Helper()
		timer := time.After(hardenWatchdog)
		for {
			select {
			case event, ok := <-client.events:
				if !ok {
					t.Fatal("events closed before a turn/completed event arrived")
				}
				if event.Method == "turn/completed" {
					return event
				}
			case <-timer:
				t.Fatalf("no turn/completed event within %s", hardenWatchdog)
			}
		}
	}
	t.Run("broker_request_in_window", func(t *testing.T) {
		client, srv := newHardenPair(t)
		client.armTurn()
		srv.turnStarted("T1")
		srv.request(501, "mcpServer/elicitation/request", elicitation(moaiMCPServerKey))
		if !srv.log.waitFor(`serverName="moai" turn=T1 broker_declined=1`) {
			t.Fatalf("the broker request was not attributed to T1; log:\n%s", srv.log.snapshot())
		}
		srv.turnCompleted("T1", "completed")
		if event := completion(t, client); !event.brokerDeclined {
			t.Fatal("completion event of the turn with a declined broker elicitation carries no verdict")
		}
	})
	t.Run("other_server_request", func(t *testing.T) {
		client, srv := newHardenPair(t)
		client.armTurn()
		srv.turnStarted("T1")
		srv.request(502, "mcpServer/elicitation/request", elicitation("other-server"))
		if !srv.log.waitFor(`serverName="other-server" turn=none broker_declined=0`) {
			t.Fatalf("log has no uncounted line for the other server; log:\n%s", srv.log.snapshot())
		}
		srv.turnCompleted("T1", "completed")
		if event := completion(t, client); event.brokerDeclined {
			t.Fatal("a request from a server other than the MoAI broker made the turn fail")
		}
	})
	t.Run("request_after_window_closed", func(t *testing.T) {
		client, srv := newHardenPair(t)
		client.armTurn()
		srv.turnStarted("T1")
		srv.turnCompleted("T1", "completed")
		if event := completion(t, client); event.brokerDeclined {
			t.Fatal("a turn without a broker request carries a verdict")
		}
		srv.request(503, "mcpServer/elicitation/request", elicitation(moaiMCPServerKey))
		if !srv.log.waitFor(`serverName="moai" turn=none broker_declined=0`) {
			t.Fatalf("the trailing request was counted or unlogged; log:\n%s", srv.log.snapshot())
		}
		client.armTurn()
		srv.turnStarted("T2")
		srv.turnCompleted("T2", "completed")
		if event := completion(t, client); event.brokerDeclined {
			t.Fatal("a request that arrived between turns made the next turn fail")
		}
	})
}

// TestManagedCodexDeclinedBrokerElicitationFailsTurn is the compilable M1 form
// of AC-MH-006 #13/#14: a MoAI-broker elicitation that arrives during a turn
// the App Server then marks `completed` must still surface as a failed
// delivery, or the unwritten receipt turns into a silent redelivery loop. The
// base owner returns nil. M1 asserted only "a non-nil error" because the marker
// symbol did not exist yet; M3 tightens it to the turn-scoped marker.
func TestManagedCodexDeclinedBrokerElicitationFailsTurn(t *testing.T) {
	elicitation := func(turnID any) map[string]any {
		return map[string]any{"threadId": "th-broker", "turnId": turnID, "serverName": moaiMCPServerKey, "message": "approve factory_msg_receipt?"}
	}
	cases := []struct {
		name   string
		script func(srv *hardenServer)
	}{
		{"single_request_null_turn_id", func(srv *hardenServer) {
			srv.startTurn("T1")
			srv.turnStarted("T1")
			srv.request(401, "mcpServer/elicitation/request", elicitation(nil))
			srv.turnCompleted("T1", "completed")
		}},
		{"two_requests_in_one_turn", func(srv *hardenServer) {
			srv.startTurn("T1")
			srv.turnStarted("T1")
			srv.request(402, "mcpServer/elicitation/request", elicitation(nil))
			srv.request(403, "mcpServer/elicitation/request", elicitation("T1"))
			srv.turnCompleted("T1", "completed")
		}},
		{"request_before_turn_started", func(srv *hardenServer) {
			srv.startTurn("T1")
			srv.request(404, "mcpServer/elicitation/request", elicitation(nil))
			srv.turnStarted("T1")
			srv.turnCompleted("T1", "completed")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, srv := newHardenPair(t)
			sess := &managedCodexSession{client: client, threadID: "th-broker"}
			done := hardenDeliver(sess, "turn that needs a broker tool")
			tc.script(srv)
			select {
			case err := <-done:
				if err == nil {
					t.Fatal("DeliverTurn = nil for a turn whose MoAI broker elicitation was declined, want a non-nil error")
				}
				if !errors.Is(err, errManagedTurnFailed) {
					t.Fatalf("DeliverTurn = %v, want a turn-scoped error (errors.Is errManagedTurnFailed) so the driver isolates it", err)
				}
			case <-time.After(hardenWatchdog):
				t.Fatalf("DeliverTurn did not return within %s", hardenWatchdog)
			}
		})
	}
}

// hardenStreamSession is a managedSession whose turns run through the REAL
// pumpManagedStreamTurn over scripted stream-json output, one script entry per
// turn (priming turn first).
type hardenStreamSession struct {
	mu      sync.Mutex
	outputs []string
	prompts []string
}

func (s *hardenStreamSession) Start() error { return nil }
func (s *hardenStreamSession) Close() error { return nil }

func (s *hardenStreamSession) DeliverTurn(prompt string) error {
	s.mu.Lock()
	index := len(s.prompts)
	s.prompts = append(s.prompts, prompt)
	s.mu.Unlock()
	if index >= len(s.outputs) {
		return errors.New("hardening script exhausted")
	}
	return pumpManagedStreamTurn(strings.NewReader(s.outputs[index]), io.Discard, io.Discard, prompt)
}

func (s *hardenStreamSession) delivered() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.prompts...)
}

// TestManagedDriverIsolatesTurnFailure is AC-MH-005's M1 form: after a
// successful priming turn, a second turn ends with an `is_error` result and a
// third succeeds. The driver must deliver the third turn instead of returning
// at the second; the base driver returns the second turn's error.
//
// M3 adds the operator-visible record: one `Factory turn failed (1/N` line on
// the owner's log output (never stdout) for the failed turn.
func TestManagedDriverIsolatesTurnFailure(t *testing.T) {
	sink := hardenLogCapture(t)
	sess := &hardenStreamSession{outputs: []string{
		`{"type":"result","is_error":false}` + "\n",
		`{"type":"result","result":"overloaded","is_error":true}` + "\n",
		`{"type":"result","is_error":false}` + "\n",
	}}
	noClaim := func() ([]factorymsg.Claim, error) { return nil, nil }
	toPrompt := func([]factorymsg.Claim) string { return "inbox" }
	errCh := make(chan error, 1)
	go func() {
		errCh <- driveManagedFactorySession(sess, strings.NewReader("op-fails\nop-succeeds\n/exit\n"), make(chan time.Time), noClaim, toPrompt)
	}()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("driver returned %v after one failed turn, want it to keep delivering and end nil on /exit", err)
		}
	case <-time.After(hardenWatchdog):
		t.Fatalf("driver did not return within %s", hardenWatchdog)
	}
	want := []string{managedPrimingPrompt, "op-fails", "op-succeeds"}
	if got := sess.delivered(); !reflect.DeepEqual(got, want) {
		t.Errorf("turns delivered = %q, want %q", got, want)
	}
	wantLine := fmt.Sprintf("Factory turn failed (1/%d consecutive): ", config.DefaultManagedSessionMaxConsecutiveTurnFailures)
	if !sink.waitFor(wantLine) {
		t.Errorf("log has no line starting %q within %s; log:\n%s", wantLine, hardenLogWait, sink.snapshot())
	}
}

// TestManagedCodexNonCompletedTurnIsolated is AC-MH-006 #2's M1 form over a
// real Codex owner: the fake App Server ends the turns after the priming turn
// as `failed`, `interrupted` and `completed`. The driver must deliver the
// third operator turn; the base driver returns at the first non-completed
// status.
func TestManagedCodexNonCompletedTurnIsolated(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "isolated.log")
	backend := fakeAppServerScript(t)
	env := []string{
		fakeAppServerRoleEnv + "=appserver",
		fakeAppServerLogEnv + "=" + logPath,
		fakeAppServerTurnStatusesEnv + "=completed,failed,interrupted,completed",
	}
	session, err := newManagedCodexSession(backend, []string{backend}, env)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Start(); err != nil {
		t.Fatalf("handshake: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	noClaim := func() ([]factorymsg.Claim, error) { return nil, nil }
	toPrompt := func([]factorymsg.Claim) string { return "inbox" }
	errCh := make(chan error, 1)
	go func() {
		errCh <- driveManagedFactorySession(session, strings.NewReader("op-failed\nop-interrupted\nop-normal\n/exit\n"), make(chan time.Time), noClaim, toPrompt)
	}()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("driver returned %v after a non-completed turn, want it to keep delivering and end nil on /exit", err)
		}
	case <-time.After(hardenChildWatchdog):
		t.Fatalf("driver did not return within %s", hardenChildWatchdog)
	}
	log, readErr := os.ReadFile(logPath)
	if readErr != nil {
		t.Fatalf("fake app server log: %v", readErr)
	}
	for _, prompt := range []string{"op-failed", "op-interrupted", "op-normal"} {
		if !strings.Contains(string(log), "turn-prompt "+prompt+"\n") {
			t.Errorf("turn %q was never delivered to the app server:\n%s", prompt, log)
		}
	}
}

// hardenFuncSession is a managedSession whose Nth delivery runs the Nth script
// function (priming turn first). A delivery past the end of the script fails
// with an UNMARKED error, so a driver that runs past what the test planned ends
// the session instead of hanging.
type hardenFuncSession struct {
	mu      sync.Mutex
	turns   []func(prompt string) error
	prompts []string
}

func (s *hardenFuncSession) Start() error { return nil }
func (s *hardenFuncSession) Close() error { return nil }

func (s *hardenFuncSession) DeliverTurn(prompt string) error {
	s.mu.Lock()
	index := len(s.prompts)
	s.prompts = append(s.prompts, prompt)
	s.mu.Unlock()
	if index >= len(s.turns) {
		return errors.New("hardening script exhausted")
	}
	return s.turns[index](prompt)
}

func (s *hardenFuncSession) delivered() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.prompts...)
}

func hardenTurnOK(string) error { return nil }

// hardenTurnFails ends the turn with a turn-scoped (marked) failure.
func hardenTurnFails(cause string) func(string) error {
	return func(string) error { return fmt.Errorf("%w: %s", errManagedTurnFailed, cause) }
}

// hardenTurnErrors ends the turn with the given error as is.
func hardenTurnErrors(err error) func(string) error {
	return func(string) error { return err }
}

// hardenRunDriver runs the real delivery driver over sess with stdin as the
// operator input and no broker claims, bounded by the 5 s watchdog. stdin must
// end with /exit so a driver that continues where the test expects a return
// ends with nil instead of hanging.
func hardenRunDriver(t *testing.T, sess managedSession, stdin string) error {
	t.Helper()
	noClaim := func() ([]factorymsg.Claim, error) { return nil, nil }
	toPrompt := func([]factorymsg.Claim) string { return "inbox" }
	errCh := make(chan error, 1)
	go func() {
		errCh <- driveManagedFactorySession(sess, strings.NewReader(stdin), make(chan time.Time), noClaim, toPrompt)
	}()
	select {
	case err := <-errCh:
		return err
	case <-time.After(hardenWatchdog):
		t.Fatalf("driver did not return within %s", hardenWatchdog)
		return nil
	}
}

// hardenWaitTurn runs waitTurn in a goroutine under the 5 s watchdog.
func hardenWaitTurn(t *testing.T, client *managedCodexAppClient, turnID string) error {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*hardenWatchdog)
		defer cancel()
		done <- client.waitTurn(ctx, turnID)
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(hardenWatchdog):
		t.Fatalf("waitTurn(%s) did not return within %s", turnID, hardenWatchdog)
		return nil
	}
}

func hardenBrokerElicitation(turnID any) map[string]any {
	return map[string]any{"threadId": "th-class", "turnId": turnID, "serverName": moaiMCPServerKey, "message": "approve factory_msg_receipt?"}
}

const hardenElicitationMethod = "mcpServer/elicitation/request"

// TestManagedTurnFailureClassification is AC-MH-005 / AC-MH-006: which failures
// are turn-scoped (the driver isolates them) and which stay session-fatal, and
// how a declined MoAI broker elicitation is attributed to a turn. Rows #1-#10
// go through the owner or the driver; rows #11-#17 drive the client's turn window
// directly with the consumer held back so the read goroutine runs ahead of it.
// Row names follow acceptance.md §1.3.
func TestManagedTurnFailureClassification(t *testing.T) {
	ceiling := config.DefaultManagedSessionMaxConsecutiveTurnFailures
	failedLine := func(k int) string {
		return fmt.Sprintf("Factory turn failed (%d/%d consecutive): ", k, ceiling)
	}
	// sessionFatalRow runs a driver whose second turn fails with `turn`'s error:
	// the driver must return that error, deliver nothing after it, and log no
	// turn-failure line.
	sessionFatalRow := func(t *testing.T, turn func(string) error, wantText string) {
		t.Helper()
		sink := hardenLogCapture(t)
		sess := &hardenFuncSession{turns: []func(string) error{hardenTurnOK, turn, hardenTurnOK}}
		err := hardenRunDriver(t, sess, "op-fails\nop-later\n/exit\n")
		if err == nil || !strings.Contains(err.Error(), wantText) {
			t.Errorf("driver returned %v, want the session-fatal error containing %q", err, wantText)
		}
		if err != nil && errors.Is(err, errManagedTurnFailed) {
			t.Errorf("session-fatal error %v carries the turn-scoped marker", err)
		}
		if got, want := sess.delivered(), []string{managedPrimingPrompt, "op-fails"}; !reflect.DeepEqual(got, want) {
			t.Errorf("turns delivered = %q, want %q (nothing after a session-fatal failure)", got, want)
		}
		if log := sink.snapshot(); strings.Contains(log, "Factory turn failed") {
			t.Errorf("a session-fatal failure logged a turn-failure line:\n%s", log)
		}
	}
	// cleanTurn scripts a turn the App Server both starts and completes under
	// the given status, for the client-level rows.
	cleanTurn := func(srv *hardenServer, turnID, status string) {
		srv.turnStarted(turnID)
		srv.turnCompleted(turnID, status)
	}

	t.Run("stream_is_error_after_priming", func(t *testing.T) {
		pumpErr := pumpManagedStreamTurn(strings.NewReader(`{"type":"result","result":"overloaded","is_error":true}`+"\n"), io.Discard, io.Discard, "p")
		if !errors.Is(pumpErr, errManagedTurnFailed) {
			t.Errorf("an is_error stream result = %v, want a turn-scoped error (errors.Is errManagedTurnFailed)", pumpErr)
		}
		sink := hardenLogCapture(t)
		sess := &hardenStreamSession{outputs: []string{
			`{"type":"result","is_error":false}` + "\n",
			`{"type":"result","result":"overloaded","is_error":true}` + "\n",
			`{"type":"result","is_error":false}` + "\n",
		}}
		if err := hardenRunDriver(t, sess, "op-fails\nop-succeeds\n/exit\n"); err != nil {
			t.Errorf("driver returned %v, want it to continue past the failed turn", err)
		}
		if got, want := sess.delivered(), []string{managedPrimingPrompt, "op-fails", "op-succeeds"}; !reflect.DeepEqual(got, want) {
			t.Errorf("turns delivered = %q, want %q", got, want)
		}
		if !sink.waitFor(failedLine(1)) {
			t.Errorf("no %q line; log:\n%s", failedLine(1), sink.snapshot())
		}
	})

	t.Run("codex_failed_or_interrupted", func(t *testing.T) {
		client, srv := newHardenPair(t)
		for i, status := range []string{"failed", "interrupted"} {
			turnID := fmt.Sprintf("T%d", i+1)
			client.armTurn()
			cleanTurn(srv, turnID, status)
			err := hardenWaitTurn(t, client, turnID)
			if err == nil {
				t.Errorf("a turn that ended %s returned nil", status)
				continue
			}
			if !errors.Is(err, errManagedTurnFailed) {
				t.Errorf("a turn that ended %s = %v, want a turn-scoped error", status, err)
			}
			if !strings.Contains(err.Error(), "ended as "+status) {
				t.Errorf("error %q does not name the %s status", err, status)
			}
		}
	})

	t.Run("codex_moai_elicitation_turn_completed", func(t *testing.T) {
		client, srv := newHardenPair(t)
		sess := &managedCodexSession{client: client, threadID: "th-class"}
		done := hardenDeliver(sess, "turn needing a broker tool")
		srv.startTurn("T1")
		srv.turnStarted("T1")
		srv.request(601, hardenElicitationMethod, hardenBrokerElicitation(nil))
		srv.turnCompleted("T1", "completed")
		select {
		case err := <-done:
			if !errors.Is(err, errManagedTurnFailed) {
				t.Errorf("DeliverTurn = %v, want a turn-scoped error for a completed turn with a declined broker elicitation", err)
			}
		case <-time.After(hardenWatchdog):
			t.Fatalf("DeliverTurn did not return within %s", hardenWatchdog)
		}
		want := `serverName="moai" turn=T1 broker_declined=1`
		if !srv.log.waitFor(want) {
			t.Errorf("log has no %q; log:\n%s", want, srv.log.snapshot())
		}
	})

	t.Run("priming_is_error", func(t *testing.T) {
		// Invariant guard G2 (acceptance.md §2.5): the priming turn returns before
		// the loop whatever its error says, so no driver mutation reaches this row.
		sess := &hardenStreamSession{outputs: []string{
			`{"type":"result","result":"priming overloaded","is_error":true}` + "\n",
			`{"type":"result","is_error":false}` + "\n",
		}}
		err := hardenRunDriver(t, sess, "op-1\n/exit\n")
		if err == nil || !strings.Contains(err.Error(), "priming overloaded") {
			t.Errorf("driver returned %v, want the priming failure (session-fatal)", err)
		}
		if got, want := sess.delivered(), []string{managedPrimingPrompt}; !reflect.DeepEqual(got, want) {
			t.Errorf("turns delivered = %q, want only the priming turn", got)
		}
	})

	t.Run("stream_closed", func(t *testing.T) {
		sessionFatalRow(t, hardenTurnErrors(errManagedStreamClosed), errManagedStreamClosed.Error())
	})

	t.Run("codex_connection_closed", func(t *testing.T) {
		sessionFatalRow(t, hardenTurnErrors(errors.New("managed codex app server connection closed")), "connection closed")
	})

	t.Run("write_failure", func(t *testing.T) {
		// A real pump over a child stdin that refuses writes.
		writeFails := func(prompt string) error {
			return pumpManagedStreamTurn(strings.NewReader(`{"type":"result","is_error":false}`+"\n"), io.Discard, managedClosedPipe{}, prompt)
		}
		sessionFatalRow(t, writeFails, "pipe closed")
	})

	t.Run("codex_timeout", func(t *testing.T) {
		// startTurn already takes its context: a short deadline gives the
		// per-turn timeout without a seam. The server answers turn/start and
		// never completes the turn, so waitTurn is what the deadline ends.
		client, srv := newHardenPair(t)
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()
		errCh := make(chan error, 1)
		go func() { errCh <- client.startTurn(ctx, "th-timeout", "never completes") }()
		srv.startTurn("T1")
		srv.turnStarted("T1")
		select {
		case err := <-errCh:
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("startTurn = %v, want context.DeadlineExceeded", err)
			}
			if errors.Is(err, errManagedTurnFailed) {
				t.Errorf("a per-turn timeout %v carries the turn-scoped marker, want session-fatal", err)
			}
		case <-time.After(hardenWatchdog):
			t.Fatalf("startTurn did not return within %s of its 150ms deadline", hardenWatchdog)
		}
	})

	t.Run("unclassified_error", func(t *testing.T) {
		sessionFatalRow(t, hardenTurnErrors(errors.New("unclassified owner failure")), "unclassified owner failure")
	})

	t.Run("other_server_elicitation", func(t *testing.T) {
		client, srv := newHardenPair(t)
		sess := &managedCodexSession{client: client, threadID: "th-class"}
		done := hardenDeliver(sess, "turn with an unrelated elicitation")
		srv.startTurn("T1")
		srv.turnStarted("T1")
		other := hardenBrokerElicitation(nil)
		other["serverName"] = "other-server"
		srv.request(602, hardenElicitationMethod, other)
		want := `serverName="other-server" turn=none broker_declined=0`
		if !srv.log.waitFor(want) {
			t.Fatalf("log has no %q; log:\n%s", want, srv.log.snapshot())
		}
		srv.turnCompleted("T1", "completed")
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("DeliverTurn = %v, want nil: only the MoAI broker's elicitation fails a turn", err)
			}
		case <-time.After(hardenWatchdog):
			t.Fatalf("DeliverTurn did not return within %s", hardenWatchdog)
		}
	})

	t.Run("between_turns_reader_ahead", func(t *testing.T) {
		client, srv := newHardenPair(t)
		client.armTurn()
		cleanTurn(srv, "T1", "completed")
		srv.request(701, hardenElicitationMethod, hardenBrokerElicitation(nil))
		// Hold the consumer back until the read goroutine has processed the
		// trailing request, so it is provably ahead of waitTurn.
		const head = `serverName="moai" turn=none broker_declined=`
		if !srv.log.waitFor(head) {
			t.Fatalf("the trailing request was not logged with turn=none within %s; log:\n%s", hardenLogWait, srv.log.snapshot())
		}
		if !strings.Contains(srv.log.snapshot(), head+"0\n") {
			t.Errorf("the trailing request was counted: want %s0; log:\n%s", head, srv.log.snapshot())
		}
		if err := hardenWaitTurn(t, client, "T1"); err != nil {
			t.Errorf("waitTurn = %v, want nil: a request after the completion frame fails no turn", err)
		}
	})

	t.Run("normal_turn_after_declined_turn", func(t *testing.T) {
		client, srv := newHardenPair(t)
		client.armTurn()
		srv.turnStarted("T1")
		srv.request(711, hardenElicitationMethod, hardenBrokerElicitation(nil))
		srv.turnCompleted("T1", "completed")
		// T1's own failure is pinned by rows #3, #13 and #14; this row only needs
		// it consumed, so it observes the reset and nothing else.
		_ = hardenWaitTurn(t, client, "T1")
		client.armTurn()
		cleanTurn(srv, "T2", "completed")
		if err := hardenWaitTurn(t, client, "T2"); err != nil {
			t.Errorf("the next, normal turn T2 = %v, want nil (armTurn resets the count)", err)
		}
	})

	t.Run("elicitation_after_turn_start_response", func(t *testing.T) {
		client, srv := newHardenPair(t)
		client.armTurn()
		srv.request(721, hardenElicitationMethod, hardenBrokerElicitation(nil))
		cleanTurn(srv, "T2", "completed")
		if err := hardenWaitTurn(t, client, "T2"); !errors.Is(err, errManagedTurnFailed) {
			t.Errorf("a request before turn/started = %v, want it to fail T2 with a turn-scoped error", err)
		}
	})

	t.Run("two_requests_in_one_turn", func(t *testing.T) {
		client, srv := newHardenPair(t)
		client.armTurn()
		srv.turnStarted("T1")
		srv.request(731, hardenElicitationMethod, hardenBrokerElicitation(nil))
		srv.request(732, hardenElicitationMethod, hardenBrokerElicitation("T1"))
		// The second line shows the window count, which tells "failed once" from
		// "counted per request".
		const second = `serverName="moai" turn=T1 broker_declined=2`
		if !srv.log.waitFor(second) {
			t.Fatalf("log has no %q within %s; log:\n%s", second, hardenLogWait, srv.log.snapshot())
		}
		srv.turnCompleted("T1", "completed")
		err := hardenWaitTurn(t, client, "T1")
		if !errors.Is(err, errManagedTurnFailed) {
			t.Fatalf("two requests in one turn = %v, want exactly one turn-scoped failure", err)
		}
	})

	t.Run("late_previous_turn_id_before_turn_started", func(t *testing.T) {
		client, srv := newHardenPair(t)
		client.armTurn()
		cleanTurn(srv, "T1", "completed")
		if err := hardenWaitTurn(t, client, "T1"); err != nil {
			t.Fatalf("T1 = %v, want nil", err)
		}
		client.armTurn()
		srv.request(741, hardenElicitationMethod, hardenBrokerElicitation("T1"))
		const head = `serverName="moai" turn=none broker_declined=`
		if !srv.log.waitFor(head) {
			t.Fatalf("a late request of the previous turn was not logged uncounted within %s; log:\n%s", hardenLogWait, srv.log.snapshot())
		}
		if !strings.Contains(srv.log.snapshot(), head+"0\n") {
			t.Errorf("the previous turn's late request was counted; log:\n%s", srv.log.snapshot())
		}
		cleanTurn(srv, "T2", "completed")
		if err := hardenWaitTurn(t, client, "T2"); err != nil {
			t.Errorf("T2 = %v, want nil: a request carrying the previous turn's id fails no turn", err)
		}
	})

	t.Run("turn_id_mismatch_in_open_window", func(t *testing.T) {
		client, srv := newHardenPair(t)
		client.armTurn()
		srv.turnStarted("T2")
		srv.request(751, hardenElicitationMethod, hardenBrokerElicitation("T9"))
		const head = `serverName="moai" turn=none broker_declined=`
		if !srv.log.waitFor(head) {
			t.Fatalf("a request for another turn was not logged uncounted within %s; log:\n%s", hardenLogWait, srv.log.snapshot())
		}
		srv.turnCompleted("T2", "completed")
		if err := hardenWaitTurn(t, client, "T2"); err != nil {
			t.Errorf("T2 = %v, want nil: a request carrying another turn's id fails no turn", err)
		}
	})

	t.Run("string_jsonrpc_id_request_counts", func(t *testing.T) {
		client, srv := newHardenPair(t)
		client.armTurn()
		srv.turnStarted("T2")
		srv.request("srv-9", hardenElicitationMethod, hardenBrokerElicitation("T2"))
		reply, ok := srv.awaitReply(`"srv-9"`, hardenWatchdog)
		if !ok {
			t.Fatalf("the string-id elicitation got no answer within %s", hardenWatchdog)
		}
		if string(reply.ID) != `"srv-9"` {
			t.Errorf("answer id %s, want the string id echoed back", reply.ID)
		}
		srv.turnCompleted("T2", "completed")
		if err := hardenWaitTurn(t, client, "T2"); !errors.Is(err, errManagedTurnFailed) {
			t.Errorf("T2 = %v, want a turn-scoped failure: the JSON-RPC id form does not matter to attribution", err)
		}
	})
}

// TestManagedBrokerNameMatchesApprovalArgs pins that the server name the owner
// compares elicitation requests against (moaiMCPServerKey) is the same name the
// owned App Server's approval overrides address (mcp_servers.<key>.…). The
// approval arguments are not edited by this SPEC; if the two names ever
// diverge, the broker-elicitation verdict would silently stop matching.
func TestManagedBrokerNameMatchesApprovalArgs(t *testing.T) {
	if moaiMCPServerName != moaiMCPServerKey {
		t.Errorf("moaiMCPServerName %q and moaiMCPServerKey %q diverged", moaiMCPServerName, moaiMCPServerKey)
	}
	args := factoryMoAIMCPApprovalArgs()
	prefix := "mcp_servers." + moaiMCPServerKey + "."
	overrides := 0
	for i, arg := range args {
		if arg == "-c" {
			continue
		}
		overrides++
		if !strings.HasPrefix(arg, prefix) {
			t.Errorf("approval override %d %q does not address %q", i, arg, prefix)
		}
	}
	if overrides == 0 {
		t.Fatal("factoryMoAIMCPApprovalArgs carries no override: the name pin swept nothing")
	}
}

// TestManagedDriverConsecutiveFailureCeiling is AC-MH-007: N consecutive
// turn-scoped failures end the session with the last error and its count, N-1
// do not, and a success resets the count. N is read from the config constant so
// the driver cannot get away with a literal.
func TestManagedDriverConsecutiveFailureCeiling(t *testing.T) {
	n := config.DefaultManagedSessionMaxConsecutiveTurnFailures
	if n < 2 {
		t.Fatalf("ceiling %d: the success_resets scenario needs at least 2", n)
	}
	failedLine := func(k int) string {
		return fmt.Sprintf("Factory turn failed (%d/%d consecutive): ", k, n)
	}
	opLines := func(k int) string {
		var b strings.Builder
		for i := 1; i <= k; i++ {
			fmt.Fprintf(&b, "op-%d\n", i)
		}
		return b.String()
	}
	script := func(steps ...func(string) error) []func(string) error {
		return append([]func(string) error{hardenTurnOK}, steps...)
	}
	fails := func(k int) []func(string) error {
		out := make([]func(string) error, k)
		for i := range out {
			out[i] = hardenTurnFails(fmt.Sprintf("overloaded #%d", i+1))
		}
		return out
	}

	t.Run("at_ceiling_returns", func(t *testing.T) {
		sink := hardenLogCapture(t)
		sess := &hardenFuncSession{turns: script(fails(n)...)}
		err := hardenRunDriver(t, sess, opLines(n)+"/exit\n")
		if err == nil {
			t.Fatalf("driver returned nil after %d consecutive turn-scoped failures, want the last failure", n)
		}
		if !errors.Is(err, errManagedTurnFailed) {
			t.Errorf("driver returned %v, want the last turn-scoped failure", err)
		}
		if want := fmt.Sprintf("%d consecutive", n); !strings.Contains(err.Error(), want) {
			t.Errorf("driver error %q does not state the count (%q)", err, want)
		}
		if want := fmt.Sprintf("overloaded #%d", n); !strings.Contains(err.Error(), want) {
			t.Errorf("driver error %q is not the LAST failure (%q)", err, want)
		}
		if got := len(sess.delivered()); got != 1+n {
			t.Errorf("%d turns delivered, want the priming turn plus %d", got, n)
		}
		if !sink.waitFor(failedLine(n)) {
			t.Errorf("no %q line for the failure that reached the ceiling; log:\n%s", failedLine(n), sink.snapshot())
		}
	})

	t.Run("below_ceiling_continues", func(t *testing.T) {
		sink := hardenLogCapture(t)
		sess := &hardenFuncSession{turns: script(append(fails(n-1), hardenTurnOK)...)}
		if err := hardenRunDriver(t, sess, opLines(n)+"/exit\n"); err != nil {
			t.Errorf("driver returned %v after %d failures (below the ceiling), want it to deliver the next turn and end nil on /exit", err, n-1)
		}
		if got := len(sess.delivered()); got != 1+(n-1)+1 {
			t.Errorf("%d turns delivered, want the priming turn, %d failures and one success", got, n-1)
		}
		if !sink.waitFor(failedLine(n - 1)) {
			t.Errorf("no %q line; log:\n%s", failedLine(n-1), sink.snapshot())
		}
	})

	t.Run("success_resets", func(t *testing.T) {
		sink := hardenLogCapture(t)
		steps := append(fails(n-1), hardenTurnOK, hardenTurnFails("one more"))
		sess := &hardenFuncSession{turns: script(steps...)}
		if err := hardenRunDriver(t, sess, opLines(n+1)+"/exit\n"); err != nil {
			t.Errorf("driver returned %v, want a success to reset the count so the last failure is 1/%d", err, n)
		}
		log := sink.snapshot()
		if got := strings.Count(log, failedLine(1)); got != 2 {
			t.Errorf("%d lines of %q, want 2 (the first failure and the one after the reset); log:\n%s", got, failedLine(1), log)
		}
		if strings.Contains(log, failedLine(n)) {
			t.Errorf("the count reached %d despite the intervening success; log:\n%s", n, log)
		}
	})
}

// TestManagedFailedTurnLeavesClaimUntouched is AC-MH-008, an INVARIANT GUARD: it
// is green on the base tree as well, because the managed layer has no handle on
// the broker row. A turn that carried a claimed message fails; the row must stay
// claimed with the same token, never acknowledged, released or re-addressed by
// this layer (redelivery after the lease is the broker's policy, covered by
// internal/factorymsg TestDispatchResultExactlyOnce).
func TestManagedFailedTurnLeavesClaimUntouched(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root, run := t.TempDir(), "managed-claim-untouched"
	activateManagedRun(t, root, run)
	store, lane, env := seedManagedInbox(t, root, run, "poison body")
	defer closeOnCleanup(t, "factory message broker", store)
	hardenLogCapture(t)
	claim, toPrompt := managedFactoryInboxWiring(store, os.Getpid(), homestate.CurrentProcessFingerprint(), run)

	sess := &hardenFuncSession{turns: []func(string) error{hardenTurnOK, hardenTurnFails("model overloaded")}}
	pr, pw := io.Pipe()
	t.Cleanup(func() { _ = pw.Close() })
	idle := make(chan time.Time, 8)
	errCh := make(chan error, 1)
	go func() { errCh <- driveManagedFactorySession(sess, pr, idle, claim, toPrompt) }()

	idle <- time.Time{}
	deadline := time.After(hardenWatchdog)
	for len(sess.delivered()) < 2 {
		select {
		case <-deadline:
			t.Fatalf("the claimed inbox turn was not delivered within %s", hardenWatchdog)
		case <-time.After(hardenLogPoll):
		}
	}
	token := ""
	for _, field := range strings.Fields(sess.delivered()[1]) {
		if strings.HasPrefix(field, "claim_token=") {
			token = strings.TrimPrefix(field, "claim_token=")
		}
	}
	if token == "" {
		t.Fatalf("the inbox prompt carries no claim token: %q", sess.delivered()[1])
	}
	go func() { _, _ = pw.Write([]byte("/exit\n")) }()
	select {
	case err := <-errCh:
		// Before the driver isolates turn failures it returns the failed turn's
		// error; after, it ends nil on /exit. Either way nothing touches the row.
		if err != nil && !errors.Is(err, errManagedTurnFailed) {
			t.Errorf("driver returned %v, want nil or the failed turn's own error", err)
		}
	case <-time.After(hardenWatchdog):
		t.Fatalf("driver did not return within %s", hardenWatchdog)
	}

	status, err := store.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.Claimed != 1 || status.Pending != 0 || status.Acknowledged != 0 || status.DeadLetter != 0 {
		t.Errorf("broker status after the failed turn = claimed %d pending %d acknowledged %d dead %d, want the one message still claimed",
			status.Claimed, status.Pending, status.Acknowledged, status.DeadLetter)
	}
	body, err := store.ReadBody(context.Background(), lane, env.ID, token)
	if err != nil || string(body) != "poison body" {
		t.Errorf("ReadBody with the original claim token = %q, %v; want the body (the claim token is unchanged)", body, err)
	}
}
