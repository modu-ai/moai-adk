package gateway

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/codextools"
	"net/http/httptest"
)

func TestGPTProductionAppServerToolRoundTrip(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Fatalf("fake App Server interpreter: %v", err)
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Fatalf("fake App Server launcher: %v", err)
	}
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(dir, "rpc.jsonl")
	source := `import sys,json
log_path=` + strconv.Quote(logPath) + `
turns=0
responses=[]
def emit(x): print(json.dumps(x),flush=True)
for line in sys.stdin:
 with open(log_path,'a') as f: f.write(line)
 m=json.loads(line); method=m.get('method')
 if method=='initialize': emit({'id':m['id'],'result':{'userAgent':'fake-app-server'}})
 elif method=='initialized': pass
 elif method=='account/read': emit({'id':m['id'],'result':{'account':{'type':'chatgpt','planType':'test'}}})
 elif method=='thread/start': emit({'id':m['id'],'result':{'thread':{'id':'thread-http'}}})
 elif method=='turn/start':
  turns+=1
  emit({'id':m['id'],'result':{'turn':{'id':'turn-http'}}})
  emit({'id':'rpc-a','method':'item/tool/call','params':{'threadId':'thread-http','turnId':'turn-http','callId':'call-a','tool':'echo_a','arguments':{'text':'hello-a'}}})
  emit({'id':'rpc-b','method':'item/tool/call','params':{'threadId':'thread-http','turnId':'turn-http','callId':'call-b','tool':'echo_b','arguments':{'text':'hello-b'}}})
 elif method=='turn/interrupt': emit({'id':m['id'],'result':{}})
 elif method is None and m.get('id') in ['rpc-a','rpc-b']:
  assert m['id'] not in responses and m['result']['success'] and m['result']['contentItems'][0]['text']=='done'
  responses.append(m['id'])
  if len(responses)==2:
   emit({'method':'item/agentMessage/delta','params':{'threadId':'thread-http','turnId':'turn-http','itemId':'text','delta':'verified answer'}})
   emit({'method':'turn/completed','params':{'threadId':'thread-http','turn':{'id':'turn-http','status':'completed'}}})
 else: raise RuntimeError('unexpected protocol')
`
	impl := filepath.Join(dir, "fake-codex.py")
	script := filepath.Join(dir, "codex")
	if err = os.WriteFile(impl, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(script, []byte("#!"+bash+"\nexec "+strconv.Quote(python)+" "+strconv.Quote(impl)+" \"$@\"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := codexapp.Start(ctx, codexapp.Config{Binary: script, Home: dir})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if _, err = client.Initialize(ctx, "fixture", "1"); err != nil {
		t.Fatal(err)
	}
	bridgeDir := filepath.Join(dir, "bridge")
	if err = os.Mkdir(bridgeDir, 0o700); err != nil {
		t.Fatal(err)
	}
	store, err := codexbridge.OpenStore(bridgeDir)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := codexbridge.New(ctx, client, codexbridge.Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	authority, err := NewAppServerAuthority(client, "chatgpt", func(context.Context) (string, error) { return "fixture-generation", nil })
	if err != nil {
		t.Fatal(err)
	}
	tools := []codextools.Definition{{Name: "echo_a", InputSchema: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}}}`)}, {Name: "echo_b", InputSchema: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}}}`)}}
	adapter, err := NewAppServerAdapter(AppServerAdapterConfig{Engine: engine, Authority: authority, Prepare: func(_ context.Context, r RoutedRequest) (codexbridge.Request, error) {
		var body struct {
			Results      []string `json:"test_results"`
			Conversation string   `json:"test_conversation"`
		}
		if err := json.Unmarshal(r.Body, &body); err != nil {
			return codexbridge.Request{}, err
		}
		q := codexbridge.Request{Owner: codextools.Binding{ConversationID: body.Conversation, AccountScope: r.Managed.Scope()}, Model: r.Entry.UpstreamID, CWD: dir, Tools: tools, PrefixDigest: "one"}
		if len(body.Results) == 0 {
			q.Input = []any{map[string]string{"type": "text", "text": "hello"}}
		} else {
			q.ExpectedPrefix, q.PrefixDigest = "one", "two"
			for _, id := range body.Results {
				q.Results = append(q.Results, codexbridge.ToolResult{ID: id, Success: true, Content: []codexbridge.Content{{Type: "inputText", Text: "done"}}})
			}
		}
		return q, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewCatalog([]ModelEntry{{RouteID: "gpt-test", UpstreamID: "gpt-test", Provider: ProviderOpenAI, AuthMethod: AuthAppServer}})
	if err != nil {
		t.Fatal(err)
	}
	server, err := NewServer(ServerConfig{SessionHeader: "X-Test", SessionToken: "fixture-session", MaxBodyBytes: 4096, Catalog: catalog, ManagedAuthority: authority, Adapters: map[ProviderID]Adapter{ProviderOpenAI: adapter}})
	if err != nil {
		t.Fatal(err)
	}
	send := func(conversation string, results []string) *httptest.ResponseRecorder {
		body, _ := json.Marshal(map[string]any{"model": "gpt-test", "max_tokens": 100, "messages": []any{map[string]string{"role": "user", "content": "hello"}}, "test_results": results, "test_conversation": conversation})
		req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body))).WithContext(ctx)
		req.Header.Set("X-Test", "fixture-session")
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)
		return res
	}
	first := send("conversation-a", nil)
	if first.Code != 200 {
		t.Fatalf("first tool request status=%d body=%s", first.Code, first.Body.String())
	}
	var message struct {
		Content []struct{ ID, Name string }
	}
	if err = json.Unmarshal(first.Body.Bytes(), &message); err != nil || len(message.Content) == 0 {
		t.Fatalf("tool request decode err=%v body=%s", err, first.Body.String())
	}
	firstBatchCount := len(message.Content)
	var continuation *httptest.ResponseRecorder
	if firstBatchCount == 2 {
		continuation = send("conversation-a", []string{message.Content[1].ID, message.Content[0].ID})
	} else {
		next := send("conversation-a", []string{message.Content[0].ID})
		var nextMessage struct {
			Content []struct{ ID, Name string }
		}
		if json.Unmarshal(next.Body.Bytes(), &nextMessage) == nil && len(nextMessage.Content) != 0 {
			message.Content = append(message.Content, nextMessage.Content...)
			continuation = send("conversation-a", []string{nextMessage.Content[0].ID})
		} else {
			continuation = next
		}
	}
	duplicate := send("conversation-a", []string{message.Content[0].ID})
	foreign := send("conversation-a", []string{"foreign-tool-id"})
	crossID := message.Content[0].ID
	if len(message.Content) > 1 {
		crossID = message.Content[1].ID
	}
	crossThread := send("conversation-b", []string{crossID})
	lines := readAppServerContractLog(t, logPath)
	counts := map[string]int{}
	responseIDs := map[string]int{}
	for _, line := range lines {
		var rpc codexapp.Message
		if json.Unmarshal([]byte(line), &rpc) != nil {
			t.Fatalf("invalid captured JSON-RPC line: %s", line)
		}
		counts[rpc.Method]++
		if rpc.Method == "" {
			responseIDs[string(rpc.ID)]++
		}
	}
	cases := []struct {
		name string
		ok   bool
		got  any
		want any
	}{
		{"initialize", counts["initialize"] == 1 && counts["initialized"] == 1 && counts["account/read"] >= 1, counts, "initialize/initialized once and account-read present"},
		{"thread-start", counts["thread/start"] == 1, counts["thread/start"], 1},
		{"turn-start", counts["turn/start"] == 1, counts["turn/start"], 1},
		{"tool-call", firstBatchCount == 2 && len(message.Content) == 2 && message.Content[0].Name != message.Content[1].Name && message.Content[0].ID != message.Content[1].ID, map[string]any{"first_batch": firstBatchCount, "calls": message.Content}, "two distinct dynamic calls in one pending batch"},
		{"claude-tool-result-continuation", continuation.Code == 200 && strings.Contains(continuation.Body.String(), "verified answer"), continuation.Code, 200},
		{"exact-response-rpc-id", responseIDs[`"rpc-a"`] == 1 && responseIDs[`"rpc-b"`] == 1, responseIDs, "rpc-a=1 rpc-b=1"},
		{"duplicate-reject", duplicate.Code != 200 && foreign.Code != 200 && responseIDs[`"foreign-tool-id"`] == 0, []int{duplicate.Code, foreign.Code}, "duplicate/foreign non-200 and zero RPC"},
		{"cross-thread-reject", crossThread.Code != 200 && responseIDs[`"foreign-tool-id"`] == 0, crossThread.Code, "non-200 and no foreign RPC response"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.ok {
				t.Errorf("App Server %s got=%v want=%v", tc.name, tc.got, tc.want)
			}
		})
	}
}

func readAppServerContractLog(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var lines []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
	return lines
}
