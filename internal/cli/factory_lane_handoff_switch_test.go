package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
)

// --- fake codex app-server over the production JSON-RPC transport seam ---

type handoffRPCRequest struct {
	Method string
	Params map[string]any
}

// fakeHandoffAppServer answers requests on the codexConn seam with the
// official app-server response shapes (codex-cli 0.155.1
// generate-json-schema: ThreadForkResponse / ThreadStartResponse /
// ThreadStartedNotification). Every request line the client writes is parsed
// back and recorded, so assertions read the wire, not a mock call log.
type fakeHandoffAppServer struct {
	newThreadID  string
	forkedFrom   string // lineage reported by the fork result; "" echoes the request threadId
	responseCwd  string // cwd reported by the result; "" echoes the request cwd
	omitStarted  bool   // never emit thread/started
	startedFirst bool   // emit thread/started before the response
	rejectMethod string // answer this method with a JSON-RPC error

	starts   int
	requests []handoffRPCRequest
}

func (f *fakeHandoffAppServer) start(context.Context, string, []string) (codexConn, error) {
	f.starts++
	return &fakeHandoffConn{srv: f}, nil
}

func (f *fakeHandoffAppServer) methods() []string {
	out := make([]string, 0, len(f.requests))
	for _, r := range f.requests {
		out = append(out, r.Method)
	}
	return out
}

type fakeHandoffConn struct {
	srv   *fakeHandoffAppServer
	queue []string
}

func (c *fakeHandoffConn) send(line string) error {
	var req struct {
		ID     int            `json:"id"`
		Method string         `json:"method"`
		Params map[string]any `json:"params"`
	}
	if err := json.Unmarshal([]byte(line), &req); err != nil {
		return err
	}
	c.srv.requests = append(c.srv.requests, handoffRPCRequest{Method: req.Method, Params: req.Params})
	id := strconv.Itoa(req.ID)
	if req.Method == c.srv.rejectMethod {
		c.queue = append(c.queue, `{"id":`+id+`,"error":{"code":-32600,"message":"rejected by fake"}}`)
		return nil
	}
	switch req.Method {
	case codexMethodInitialize:
		c.queue = append(c.queue, `{"id":`+id+`,"result":{"userAgent":"fake/0.155.1","codexHome":"/x","platformFamily":"unix","platformOs":"macos"}}`)
	case codexMethodThreadFork, codexMethodThreadStart:
		cwd, _ := req.Params["cwd"].(string)
		if c.srv.responseCwd != "" {
			cwd = c.srv.responseCwd
		}
		forked := "null"
		if req.Method == codexMethodThreadFork {
			from, _ := req.Params["threadId"].(string)
			if c.srv.forkedFrom != "" {
				from = c.srv.forkedFrom
			}
			forked = strconv.Quote(from)
		}
		thread := fmt.Sprintf(`{"id":%q,"forkedFromId":%s,"cwd":%q,"cliVersion":"0.155.1","createdAt":1,"updatedAt":1,"ephemeral":false,"modelProvider":"openai","preview":"","projectId":null,"sessionId":%q,"source":"appServer","status":{"type":"idle"},"turns":[]}`,
			c.srv.newThreadID, forked, cwd, c.srv.newThreadID)
		resp := fmt.Sprintf(`{"id":%s,"result":{"approvalPolicy":"never","approvalsReviewer":"user","cwd":%q,"model":"gpt-5-codex","modelProvider":"openai","sandbox":{"type":"readOnly"},"thread":%s}}`, id, cwd, thread)
		started := `{"method":"thread/started","params":{"thread":` + thread + `}}`
		switch {
		case c.srv.omitStarted:
			c.queue = append(c.queue, resp)
		case c.srv.startedFirst:
			c.queue = append(c.queue, started, resp)
		default:
			c.queue = append(c.queue, resp, started)
		}
	default:
		c.queue = append(c.queue, `{"id":`+id+`,"error":{"code":-32601,"message":"method not found"}}`)
	}
	return nil
}

func (c *fakeHandoffConn) recv() (string, bool) {
	if len(c.queue) == 0 {
		return "", false
	}
	l := c.queue[0]
	c.queue = c.queue[1:]
	return l, true
}

func (c *fakeHandoffConn) close() error { return nil }

func withFakeHandoffAppServer(t *testing.T, srv *fakeHandoffAppServer) *fakeHandoffAppServer {
	t.Helper()
	prevSess, prevLook := codexSession, codexLookPath
	codexSession = srv
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }
	t.Cleanup(func() { codexSession, codexLookPath = prevSess, prevLook })
	return srv
}

// --- switch fixture helpers ---

func (f *laneHandoffFixture) wtReady(t *testing.T, mode string) factorymsg.Handoff {
	t.Helper()
	h, err := prepareLaneHandoff(context.Background(), f.request(handoffTestCard, handoffTestSlug, mode), f.deps())
	if err != nil || h.State != factorymsg.HandoffWTReady {
		t.Fatalf("prepare %s = %+v err=%v, want WT_READY", mode, h, err)
	}
	return h
}

func (f *laneHandoffFixture) switchRequest(activity laneHandoffActivity, sourceThread string) laneHandoffSwitch {
	return laneHandoffSwitch{ProjectRoot: f.primary, Activity: activity, SourceThreadID: sourceThread}
}

func (f *laneHandoffFixture) storedHandoff(t *testing.T, id string) factorymsg.Handoff {
	t.Helper()
	hs, err := f.store.HandoffsForLane(context.Background(), handoffTestSlot)
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range hs {
		if h.ID == id {
			return h
		}
	}
	t.Fatalf("handoff %s not stored", id)
	return factorymsg.Handoff{}
}

// requireNoEndpointEffects asserts nothing past SWITCH_PENDING happened: no
// BOUND, tombstone, peer change, or dispatch release (M3 owns all four).
func (f *laneHandoffFixture) requireNoEndpointEffects(t *testing.T, before handoffEndpointRow) {
	t.Helper()
	if after := f.endpointRow(t); after != before {
		t.Fatalf("endpoint row mutated: before=%+v after=%+v", before, after)
	}
	for q, want := range map[string]int{
		`SELECT count(*) FROM lane_handoffs WHERE state='BOUND'`:      0,
		`SELECT count(*) FROM lane_handoff_events WHERE to_state='BOUND'`: 0,
		`SELECT count(*) FROM lane_endpoint_tombstones`:               0,
		`SELECT count(*) FROM messages`:                               0,
	} {
		if n := f.count(t, q); n != want {
			t.Fatalf("%s = %d, want %d", q, n, want)
		}
	}
}

// --- interactive adapter ---

func TestLaneHandoffInteractiveSwitchEmitsGuidanceOnly(t *testing.T) {
	f := newLaneHandoffFixture(t, "develop", true)
	srv := withFakeHandoffAppServer(t, &fakeHandoffAppServer{newThreadID: "thr-new"})
	h := f.wtReady(t, factorymsg.HandoffModeInteractive)
	before := f.endpointRow(t)
	var out bytes.Buffer

	got, err := switchLaneHandoffInteractive(context.Background(), h, f.switchRequest(laneActivityIdle, ""), laneHandoffDeps{Out: &out})
	if err != nil {
		t.Fatalf("interactive switch: %v", err)
	}
	if got.State != factorymsg.HandoffSwitchPendingInteractive || f.storedHandoff(t, h.ID).State != factorymsg.HandoffSwitchPendingInteractive {
		t.Fatalf("state = %s, want %s", got.State, factorymsg.HandoffSwitchPendingInteractive)
	}
	guidance := out.String()
	if !strings.Contains(guidance, "/cd "+h.TargetPath+"\n") || !strings.Contains(guidance, h.Nonce) {
		t.Fatalf("guidance lacks the absolute /cd target or the nonce:\n%s", guidance)
	}
	if strings.Count(guidance, "/cd ") != 1 {
		t.Fatalf("guidance carries %d /cd lines, want 1:\n%s", strings.Count(guidance, "/cd "), guidance)
	}
	if srv.starts != 0 || len(srv.requests) != 0 {
		t.Fatalf("interactive switch reached the app-server: starts=%d requests=%v", srv.starts, srv.methods())
	}
	f.requireNoEndpointEffects(t, before)
}

func TestLaneHandoffInteractiveSwitchRejectsNonIdle(t *testing.T) {
	for _, tc := range []struct {
		activity laneHandoffActivity
		want     string
	}{
		{laneActivityActiveTurn, factorymsg.NackLaneActiveTurn},
		{laneActivityPermissionWait, factorymsg.NackLanePermissionWait},
		{laneActivityInterrupting, factorymsg.NackLaneInterrupting},
	} {
		t.Run(string(tc.activity), func(t *testing.T) {
			f := newLaneHandoffFixture(t, "develop", true)
			h := f.wtReady(t, factorymsg.HandoffModeInteractive)
			before := f.endpointRow(t)
			var out bytes.Buffer
			_, err := switchLaneHandoffInteractive(context.Background(), h, f.switchRequest(tc.activity, ""), laneHandoffDeps{Out: &out})
			requireHandoffNack(t, err, tc.want)
			if st := f.storedHandoff(t, h.ID); st.State != factorymsg.HandoffNack || st.Reason != tc.want {
				t.Fatalf("stored = %s/%s, want NACK/%s", st.State, st.Reason, tc.want)
			}
			if strings.Contains(out.String(), "/cd ") {
				t.Fatalf("guidance emitted for a non-idle lane: %q", out.String())
			}
			f.requireNoEndpointEffects(t, before)
		})
	}
}

// --- headless adapter ---

func TestLaneHandoffHeadlessSwitchForksStoredHistory(t *testing.T) {
	for _, startedFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("started_first=%v", startedFirst), func(t *testing.T) {
			f := newLaneHandoffFixture(t, "develop", true)
			srv := withFakeHandoffAppServer(t, &fakeHandoffAppServer{newThreadID: "thr-forked", startedFirst: startedFirst})
			h := f.wtReady(t, factorymsg.HandoffModeHeadless)
			before := f.endpointRow(t)
			var out bytes.Buffer

			got, err := switchLaneHandoffHeadless(context.Background(), h, f.switchRequest(laneActivityIdle, "thr-source"), laneHandoffDeps{Out: &out})
			if err != nil {
				t.Fatalf("headless switch: %v", err)
			}
			if got.State != factorymsg.HandoffSwitchPendingHeadless {
				t.Fatalf("state = %s", got.State)
			}
			if want := []string{codexMethodInitialize, codexMethodThreadFork}; !reflect.DeepEqual(srv.methods(), want) {
				t.Fatalf("requests = %v, want %v", srv.methods(), want)
			}
			if want := map[string]any{"threadId": "thr-source", "cwd": h.TargetPath}; !reflect.DeepEqual(srv.requests[1].Params, want) {
				t.Fatalf("thread/fork params = %v, want %v", srv.requests[1].Params, want)
			}
			rel, ok, err := f.store.HeadlessRelocationFor(context.Background(), h.ID)
			if err != nil || !ok {
				t.Fatalf("relocation evidence ok=%v err=%v", ok, err)
			}
			want := factorymsg.HeadlessRelocation{
				HandoffID: h.ID, Nonce: h.Nonce, Method: factorymsg.RelocationMethodThreadFork,
				SourceThreadID: "thr-source", ThreadID: "thr-forked", ForkedFromID: "thr-source", ThreadStarted: true,
				RequestCwd: h.TargetPath, ResponseCwd: h.TargetPath,
				ReadbackCwd: h.TargetPath, ReadbackBranch: handoffTestBranch, ReadbackHead: f.developPin,
				RecordedAt: rel.RecordedAt,
			}
			if rel != want {
				t.Fatalf("relocation = %+v, want %+v", rel, want)
			}
			if strings.Contains(out.String(), "/cd ") {
				t.Fatalf("headless switch emitted interactive guidance: %q", out.String())
			}
			f.requireNoEndpointEffects(t, before)
		})
	}
}

func TestLaneHandoffHeadlessSwitchStartsWithoutHistory(t *testing.T) {
	f := newLaneHandoffFixture(t, "develop", true)
	srv := withFakeHandoffAppServer(t, &fakeHandoffAppServer{newThreadID: "thr-fresh"})
	h := f.wtReady(t, factorymsg.HandoffModeHeadless)
	got, err := switchLaneHandoffHeadless(context.Background(), h, f.switchRequest(laneActivityIdle, ""), laneHandoffDeps{})
	if err != nil || got.State != factorymsg.HandoffSwitchPendingHeadless {
		t.Fatalf("switch = %+v err=%v", got, err)
	}
	if want := []string{codexMethodInitialize, codexMethodThreadStart}; !reflect.DeepEqual(srv.methods(), want) {
		t.Fatalf("requests = %v, want %v", srv.methods(), want)
	}
	if want := map[string]any{"cwd": h.TargetPath}; !reflect.DeepEqual(srv.requests[1].Params, want) {
		t.Fatalf("thread/start params = %v, want %v", srv.requests[1].Params, want)
	}
	rel, ok, err := f.store.HeadlessRelocationFor(context.Background(), h.ID)
	if err != nil || !ok || rel.Method != factorymsg.RelocationMethodThreadStart || rel.ThreadID != "thr-fresh" || rel.ForkedFromID != "" || !rel.ThreadStarted {
		t.Fatalf("relocation = %+v ok=%v err=%v", rel, ok, err)
	}
}

func TestLaneHandoffHeadlessSwitchRejectsNonIdleBeforeRPC(t *testing.T) {
	for _, tc := range []struct {
		activity laneHandoffActivity
		want     string
	}{
		{laneActivityActiveTurn, factorymsg.NackLaneActiveTurn},
		{laneActivityPermissionWait, factorymsg.NackLanePermissionWait},
		{laneActivityInterrupting, factorymsg.NackLaneInterrupting},
	} {
		t.Run(string(tc.activity), func(t *testing.T) {
			f := newLaneHandoffFixture(t, "develop", true)
			srv := withFakeHandoffAppServer(t, &fakeHandoffAppServer{newThreadID: "thr-x"})
			h := f.wtReady(t, factorymsg.HandoffModeHeadless)
			before := f.endpointRow(t)
			_, err := switchLaneHandoffHeadless(context.Background(), h, f.switchRequest(tc.activity, "thr-source"), laneHandoffDeps{})
			requireHandoffNack(t, err, tc.want)
			if srv.starts != 0 {
				t.Fatalf("app-server started %d times for a non-idle lane", srv.starts)
			}
			if st := f.storedHandoff(t, h.ID); st.State != factorymsg.HandoffNack || st.Reason != tc.want {
				t.Fatalf("stored = %s/%s, want NACK/%s", st.State, st.Reason, tc.want)
			}
			if n := f.count(t, `SELECT count(*) FROM lane_handoff_events WHERE to_state='SWITCH_PENDING_HEADLESS'`); n != 0 {
				t.Fatalf("SWITCH_PENDING_HEADLESS entered %d times", n)
			}
			f.requireNoEndpointEffects(t, before)
		})
	}
}

func TestLaneHandoffHeadlessSwitchNacksBrokenRelocation(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		srv        fakeHandoffAppServer
	}{
		{"rpc_rejected", factorymsg.NackRelocationRPCFailed, fakeHandoffAppServer{newThreadID: "thr-x", rejectMethod: codexMethodThreadFork}},
		{"thread_started_missing", factorymsg.NackRelocationRPCFailed, fakeHandoffAppServer{newThreadID: "thr-x", omitStarted: true}},
		{"empty_thread_id", factorymsg.NackRelocationRPCFailed, fakeHandoffAppServer{newThreadID: ""}},
		{"lineage_foreign", factorymsg.NackRelocationEvidenceInvalid, fakeHandoffAppServer{newThreadID: "thr-x", forkedFrom: "thr-other"}},
		{"response_cwd_foreign", factorymsg.NackRelocationEvidenceInvalid, fakeHandoffAppServer{newThreadID: "thr-x", responseCwd: "/elsewhere"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newLaneHandoffFixture(t, "develop", true)
			srv := tc.srv
			withFakeHandoffAppServer(t, &srv)
			h := f.wtReady(t, factorymsg.HandoffModeHeadless)
			before := f.endpointRow(t)
			_, err := switchLaneHandoffHeadless(context.Background(), h, f.switchRequest(laneActivityIdle, "thr-source"), laneHandoffDeps{})
			requireHandoffNack(t, err, tc.want)
			if st := f.storedHandoff(t, h.ID); st.State != factorymsg.HandoffNack || st.Reason != tc.want {
				t.Fatalf("stored = %s/%s, want NACK/%s", st.State, st.Reason, tc.want)
			}
			if _, ok, _ := f.store.HeadlessRelocationFor(context.Background(), h.ID); ok {
				t.Fatal("broken relocation was recorded as evidence")
			}
			for _, m := range srv.methods() {
				if m == codexMethodTurnStart || m == "turn/steer" {
					t.Fatalf("relocation manufactured evidence with %s", m)
				}
			}
			f.requireNoEndpointEffects(t, before)
		})
	}
}

func TestLaneHandoffHeadlessSwitchNacksReadbackMismatchBeforeRPC(t *testing.T) {
	f := newLaneHandoffFixture(t, "develop", true)
	srv := withFakeHandoffAppServer(t, &fakeHandoffAppServer{newThreadID: "thr-x"})
	h := f.wtReady(t, factorymsg.HandoffModeHeadless)
	handoffGit(t, h.TargetPath, "commit", "-q", "--allow-empty", "-m", "moved off pin")
	_, err := switchLaneHandoffHeadless(context.Background(), h, f.switchRequest(laneActivityIdle, "thr-source"), laneHandoffDeps{})
	requireHandoffNack(t, err, factorymsg.NackTargetReadbackMismatch)
	if srv.starts != 0 {
		t.Fatalf("app-server started %d times for a drifted target", srv.starts)
	}
}

func TestLaneHandoffSwitchModesDoNotCross(t *testing.T) {
	f := newLaneHandoffFixture(t, "develop", true)
	srv := withFakeHandoffAppServer(t, &fakeHandoffAppServer{newThreadID: "thr-x"})
	h := f.wtReady(t, factorymsg.HandoffModeInteractive)
	if _, err := switchLaneHandoffHeadless(context.Background(), h, f.switchRequest(laneActivityIdle, "thr-source"), laneHandoffDeps{}); err == nil {
		t.Fatal("interactive handoff took the headless switch")
	}
	if srv.starts != 0 {
		t.Fatalf("app-server started %d times for an interactive handoff", srv.starts)
	}
	if st := f.storedHandoff(t, h.ID); st.State != factorymsg.HandoffWTReady {
		t.Fatalf("stored state = %s, want WT_READY untouched", st.State)
	}

	g := newLaneHandoffFixture(t, "develop", true)
	hh := g.wtReady(t, factorymsg.HandoffModeHeadless)
	var out bytes.Buffer
	if _, err := switchLaneHandoffInteractive(context.Background(), hh, g.switchRequest(laneActivityIdle, ""), laneHandoffDeps{Out: &out}); err == nil {
		t.Fatal("headless handoff took the interactive switch")
	}
	if out.Len() != 0 {
		t.Fatalf("guidance emitted for a headless handoff: %q", out.String())
	}
	if st := g.storedHandoff(t, hh.ID); st.State != factorymsg.HandoffWTReady {
		t.Fatalf("stored state = %s, want WT_READY untouched", st.State)
	}
}

// A switched handoff is no longer WT_READY: a second switch changes nothing
// and emits no second guidance.
func TestLaneHandoffSwitchRefusesSecondSwitch(t *testing.T) {
	f := newLaneHandoffFixture(t, "develop", true)
	h := f.wtReady(t, factorymsg.HandoffModeInteractive)
	if _, err := switchLaneHandoffInteractive(context.Background(), h, f.switchRequest(laneActivityIdle, ""), laneHandoffDeps{}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if _, err := switchLaneHandoffInteractive(context.Background(), h, f.switchRequest(laneActivityIdle, ""), laneHandoffDeps{Out: &out}); err == nil {
		t.Fatal("second interactive switch accepted")
	}
	if out.Len() != 0 {
		t.Fatalf("second switch emitted guidance: %q", out.String())
	}
	if n := f.count(t, `SELECT count(*) FROM lane_handoff_events WHERE to_state='SWITCH_PENDING_INTERACTIVE'`); n != 1 {
		t.Fatalf("SWITCH_PENDING_INTERACTIVE events = %d, want 1", n)
	}
}

// Without a codex binary the headless relocation cannot run: the handoff is
// NACKed RELOCATION_RPC_FAILED and no evidence is recorded.
func TestLaneHandoffHeadlessSwitchNacksMissingCodexBinary(t *testing.T) {
	f := newLaneHandoffFixture(t, "develop", true)
	srv := withFakeHandoffAppServer(t, &fakeHandoffAppServer{newThreadID: "thr-x"})
	codexLookPath = func(string) (string, error) { return "", exec.ErrNotFound }
	h := f.wtReady(t, factorymsg.HandoffModeHeadless)
	_, err := switchLaneHandoffHeadless(context.Background(), h, f.switchRequest(laneActivityIdle, "thr-source"), laneHandoffDeps{})
	requireHandoffNack(t, err, factorymsg.NackRelocationRPCFailed)
	if srv.starts != 0 {
		t.Fatalf("app-server started %d times without a binary", srv.starts)
	}
	if _, ok, _ := f.store.HeadlessRelocationFor(context.Background(), h.ID); ok {
		t.Fatal("evidence recorded without a relocation")
	}
}

// --- AC-FLH-014 ---

// handoffControlFiles are the production files that drive a lane handoff.
var handoffControlFiles = []string{"factory_lane_handoff.go", "factory_lane_handoff_switch.go"}

// scanHandoffControl lists every private or overstated control path the
// handoff sources reach for: forbidden imports (sockets, hook, MCP), forbidden
// identifiers (tmux spawn, peer registration, message send, dialing), and
// forbidden string literals (tmux keys, slash automation, model turns,
// Desktop Handoff).
func scanHandoffControl(t *testing.T, files []string) []string {
	t.Helper()
	forbiddenImports := []string{"net", "net/http", "net/rpc", "github.com/modu-ai/moai-adk/internal/hook", "github.com/modu-ai/moai-adk/internal/mcp", "github.com/mark3labs/mcp-go"}
	forbiddenIdents := map[string]bool{
		"tmuxSpawnFn": true, "defaultTmuxSpawn": true, "RegisterPeer": true, "BindLaunchPending": true,
		"RegisterLaunchPending": true, "Send": true, "Dial": true, "DialContext": true,
		"codexMethodTurnStart": true, "runTurn": true,
	}
	forbiddenLiterals := []string{"tmux", "send-keys", "chatgpt", "desktop", "turn/start", "turn/steer", "sessionstart", "unix:"}
	var hits []string
	fset := token.NewFileSet()
	for _, name := range files {
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, imp := range file.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			for _, bad := range forbiddenImports {
				if path == bad || strings.HasPrefix(path, bad+"/") {
					hits = append(hits, name+": import "+path)
				}
			}
		}
		ast.Inspect(file, func(n ast.Node) bool {
			switch v := n.(type) {
			case *ast.Ident:
				if forbiddenIdents[v.Name] {
					hits = append(hits, fmt.Sprintf("%s: identifier %s", fset.Position(v.Pos()), v.Name))
				}
			case *ast.BasicLit:
				if v.Kind == token.STRING {
					lit := strings.ToLower(v.Value)
					for _, bad := range forbiddenLiterals {
						if strings.Contains(lit, bad) {
							hits = append(hits, fmt.Sprintf("%s: literal %s", fset.Position(v.Pos()), v.Value))
						}
					}
				}
			}
			return true
		})
	}
	return hits
}

// TestFactoryLaneHandoffNoPrivateControl is AC-FLH-014 (REQ-FLH-006/007/012).
func TestFactoryLaneHandoffNoPrivateControl(t *testing.T) {
	// Static half, read before the fixture moves the process cwd.
	if hits := scanHandoffControl(t, handoffControlFiles); len(hits) != 0 {
		t.Fatalf("private or overstated control path in handoff sources:\n%s", strings.Join(hits, "\n"))
	}

	// Runtime half: spies on every channel a handoff could reach.
	tmuxCalls := 0
	prevTmux := tmuxSpawnFn
	tmuxSpawnFn = func(string, string) (string, error) { tmuxCalls++; return "", nil }
	t.Cleanup(func() { tmuxSpawnFn = prevTmux })
	var commands [][]string
	prevCmd := handoffCommand
	handoffCommand = func(name string, args ...string) *exec.Cmd {
		commands = append(commands, append([]string{name}, args...))
		return prevCmd(name, args...)
	}
	t.Cleanup(func() { handoffCommand = prevCmd })

	runMode := func(t *testing.T, mode string) (string, *fakeHandoffAppServer, factorymsg.Handoff) {
		t.Helper()
		f := newLaneHandoffFixture(t, "develop", true)
		srv := withFakeHandoffAppServer(t, &fakeHandoffAppServer{newThreadID: "thr-new"})
		h := f.wtReady(t, mode)
		before := f.endpointRow(t)
		var out bytes.Buffer
		deps := laneHandoffDeps{Out: &out}
		var err error
		if mode == factorymsg.HandoffModeInteractive {
			_, err = switchLaneHandoffInteractive(context.Background(), h, f.switchRequest(laneActivityIdle, ""), deps)
		} else {
			_, err = switchLaneHandoffHeadless(context.Background(), h, f.switchRequest(laneActivityIdle, "thr-source"), deps)
		}
		if err != nil {
			t.Fatalf("%s switch: %v", mode, err)
		}
		// Hook/model channels: no peer registration, no message to the model.
		f.requireNoEndpointEffects(t, before)
		return out.String(), srv, h
	}

	guidance, srv, h := runMode(t, factorymsg.HandoffModeInteractive)
	if srv.starts != 0 {
		t.Fatalf("interactive app-server sessions = %d, want 0 (no model turn)", srv.starts)
	}
	if strings.Count(guidance, "/cd "+h.TargetPath+"\n") != 1 || !strings.Contains(guidance, h.Nonce) {
		t.Fatalf("interactive guidance = %q, want exactly one /cd line with the nonce", guidance)
	}

	guidance, srv, _ = runMode(t, factorymsg.HandoffModeHeadless)
	if guidance != "" {
		t.Fatalf("headless emitted operator guidance %q", guidance)
	}
	if want := []string{codexMethodInitialize, codexMethodThreadFork}; srv.starts != 1 || !reflect.DeepEqual(srv.methods(), want) {
		t.Fatalf("headless app-server = %d sessions %v, want 1 session %v", srv.starts, srv.methods(), want)
	}

	if tmuxCalls != 0 {
		t.Fatalf("tmux spawns = %d, want 0", tmuxCalls)
	}
	if len(commands) == 0 {
		t.Fatal("subprocess spy observed nothing; the spy is not wired")
	}
	for _, c := range commands {
		if c[0] != "git" {
			t.Fatalf("handoff ran %v; only git readback is allowed", c)
		}
		for _, a := range c {
			if a == "send-keys" {
				t.Fatalf("handoff ran %v", c)
			}
		}
	}
}
