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
	"context"
	"encoding/json"
	"errors"
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
	"github.com/modu-ai/moai-adk/internal/factorymsg"
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
)

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
}

// isReply reports a client answer to a server request: an id and no method.
func (f hardenFrame) isReply() bool { return f.Method == "" && len(f.ID) > 0 }

// hardenServer is the server end of a hardenPair.
type hardenServer struct {
	t      *testing.T
	conn   *websocket.Conn
	frames chan hardenFrame
}

// newHardenPair joins a started managedCodexAppClient (read goroutine running)
// to the server end the test drives. Cleanup order is LIFO: the client is shut
// down and its read goroutine awaited first, then the server connection and
// listener close.
func newHardenPair(t *testing.T) (*managedCodexAppClient, *hardenServer) {
	t.Helper()
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

	server := &hardenServer{t: t, conn: serverConn, frames: make(chan hardenFrame, 256)}
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
		var frame hardenFrame
		if err := s.conn.ReadJSON(&frame); err != nil {
			return
		}
		s.frames <- frame
	}
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
	name   string
	method string
	params map[string]any
	check  func(*testing.T, hardenFrame)
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
		{"command_execution_approval", "item/commandExecution/requestApproval", base(nil), hardenWantResult(`{"decision":"decline"}`)},
		{"file_change_approval", "item/fileChange/requestApproval", base(nil), hardenWantResult(`{"decision":"decline"}`)},
		{"permissions_approval", "item/permissions/requestApproval", base(nil), hardenWantResult(`{"permissions":{}}`)},
		{"mcp_elicitation", "mcpServer/elicitation/request", base(map[string]any{"serverName": "other-server", "message": "need input"}), hardenWantResult(`{"action":"decline"}`)},
		{"tool_request_user_input", "item/tool/requestUserInput", base(nil), hardenWantError(-32000)},
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
		}},
		{"chatgpt_token_refresh", "account/chatgptAuthTokens/refresh", base(nil), hardenWantError(-32000)},
		{"attestation_generate", "attestation/generate", base(nil), hardenWantError(-32000)},
		{"legacy_apply_patch_approval", "applyPatchApproval", base(nil), hardenWantResult(`{"decision":"denied"}`)},
		{"legacy_exec_command_approval", "execCommandApproval", base(nil), hardenWantResult(`{"decision":"denied"}`)},
		{"unknown_method", "item/unlisted/needsAnswer", base(nil), hardenWantError(-32601)},
	}
}

// TestManagedCodexServerRequestPolicy is AC-MH-001's M1 form: during one turn
// the fake server sends ten listed server requests and one unlisted method,
// each with an integer id. The owner must answer each exactly once, with the
// same id, per the D-1 policy table. The base owner never answers, so the fake
// server records the absence after hardenReplyWait.
//
// The per-request log-line assertions (design.md D-1 log seam) arrive with the
// seam in M2; the seam does not exist on the base API.
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

// TestManagedCodexDeclinedBrokerElicitationFailsTurn is the compilable M1 form
// of AC-MH-006 #13/#14: a MoAI-broker elicitation that arrives during a turn
// the App Server then marks `completed` must still surface as a failed
// delivery, or the unwritten receipt turns into a silent redelivery loop. The
// base owner returns nil. The assertion is only "a non-nil error" — the marker
// symbol does not exist on the base API (M3 adds it and tightens this).
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
// The `Factory turn failed (1/` log-line assertion arrives with the log seam
// in M3.
func TestManagedDriverIsolatesTurnFailure(t *testing.T) {
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
