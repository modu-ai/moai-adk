package gateway

import (
	"context"
	"encoding/json"
	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/codextools"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// This is a subprocess protocol fixture, not a live provider/model acceptance.
func TestAppServerSubprocessHTTPToolContinuation(t *testing.T) {
	for _, kind := range []string{"chatgpt", "apiKey"} {
		t.Run(kind, func(t *testing.T) {
			python, err := exec.LookPath("python3")
			if err != nil {
				t.Skip("python3 subprocess fixture unavailable")
			}
			dir, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			if err = os.Chmod(dir, 0700); err != nil {
				t.Fatal(err)
			}
			script := filepath.Join(dir, "fake-codex")
			source := `import sys,json
turns=0
responses=0
def emit(x):
 print(json.dumps(x),flush=True)
for line in sys.stdin:
 m=json.loads(line)
 method=m.get('method')
 if method=='initialize': emit({'id':m['id'],'result':{'userAgent':'fake-app-server'}})
 elif method=='initialized': pass
 elif method=='account/read': emit({'id':m['id'],'result':{'account':{'type':'chatgpt','planType':'test'}}})
 elif method=='thread/start':
  assert 'env' not in m['params'] and m['params']['environments']==[]
  assert all(t['type']=='function' for t in m['params']['dynamicTools'])
  emit({'id':m['id'],'result':{'thread':{'id':'thread-http'}}})
 elif method=='turn/start':
  assert m['params']['environments']==[]
  turns+=1
  assert turns==1
  emit({'id':m['id'],'result':{'turn':{'id':'turn-http'}}})
  emit({'id':'rpc-http','method':'item/tool/call','params':{'threadId':'thread-http','turnId':'turn-http','callId':'call-http','tool':'echo','arguments':{'text':'hello'}}})
 elif method=='turn/interrupt': emit({'id':m['id'],'result':{}})
 elif method is None and m.get('id')=='rpc-http':
  responses+=1
  assert responses==1 and m['result']['success'] and m['result']['contentItems'][0]['text']=='done'
  emit({'method':'item/agentMessage/delta','params':{'threadId':'thread-http','turnId':'turn-http','itemId':'text','delta':'verified answer'}})
  emit({'method':'turn/completed','params':{'threadId':'thread-http','turn':{'id':'turn-http','status':'completed'}}})
 else: raise RuntimeError('unexpected protocol')
`
			source = strings.ReplaceAll(source, "'chatgpt'", "'"+kind+"'")
			// A pyenv-shim python3 is itself a script, and macOS rejects the
			// nested shebang with ENOEXEC (Go's fork/exec has no userspace
			// fallback); exec python through a binary interpreter instead.
			bash, err := exec.LookPath("bash")
			if err != nil {
				t.Skip("bash unavailable")
			}
			impl := filepath.Join(dir, "fake-codex-impl.py")
			if err = os.WriteFile(impl, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(script, []byte("#!"+bash+"\nexec "+strconv.Quote(python)+" "+strconv.Quote(impl)+" \"$@\"\n"), 0700); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			client, err := codexapp.Start(ctx, codexapp.Config{Binary: script, Home: dir})
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if cerr := client.Close(); cerr != nil {
					t.Error(cerr)
				}
			}()
			if _, err = client.Initialize(ctx, "fixture", "1"); err != nil {
				t.Fatal(err)
			}
			state := filepath.Join(dir, "bridge")
			if err = os.Mkdir(state, 0700); err != nil {
				t.Fatal(err)
			}
			store, err := codexbridge.OpenStore(state)
			if err != nil {
				t.Fatal(err)
			}
			engine, err := codexbridge.New(ctx, client, codexbridge.Config{Store: store})
			if err != nil {
				t.Fatal(err)
			}
			defer engine.Close()
			authority, err := NewAppServerAuthority(client, kind, func(context.Context) (string, error) { return "fixture-generation", nil })
			if err != nil {
				t.Fatal(err)
			}
			tool := codextools.Definition{Name: "echo", InputSchema: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}`)}
			adapter, err := NewAppServerAdapter(AppServerAdapterConfig{Engine: engine, Authority: authority, Prepare: func(_ context.Context, r RoutedRequest) (codexbridge.Request, error) {
				var body struct {
					Result string `json:"test_result"`
				}
				if err := json.Unmarshal(r.Body, &body); err != nil {
					return codexbridge.Request{}, err
				}
				q := codexbridge.Request{Owner: codextools.Binding{ConversationID: "http-conversation", AccountScope: r.Managed.Scope()}, Model: r.Entry.UpstreamID, CWD: dir, Tools: []codextools.Definition{tool}, PrefixDigest: "one"}
				if body.Result == "" {
					q.Input = []any{map[string]string{"type": "text", "text": "hello"}}
				} else {
					q.ExpectedPrefix = "one"
					q.PrefixDigest = "two"
					q.Results = []codexbridge.ToolResult{{ID: body.Result, Success: true, Content: []codexbridge.Content{{Type: "inputText", Text: "done"}}}}
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
			send := func(result string, stream bool) *httptest.ResponseRecorder {
				body, _ := json.Marshal(map[string]any{"model": "gpt-test", "max_tokens": 100, "stream": stream, "messages": []any{map[string]string{"role": "user", "content": "hello"}}, "test_result": result})
				r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body))).WithContext(ctx)
				r.Header.Set("X-Test", "fixture-session")
				w := httptest.NewRecorder()
				server.ServeHTTP(w, r)
				return w
			}
			first := send("", false)
			if first.Code != 200 {
				t.Fatal(first.Code, first.Body.String())
			}
			var message struct {
				Content    []struct{ Type, ID, Name string }
				StopReason string `json:"stop_reason"`
			}
			if err = json.Unmarshal(first.Body.Bytes(), &message); err != nil {
				t.Fatal(err)
			}
			if message.StopReason != "tool_use" || len(message.Content) != 1 || message.Content[0].Name != "echo" {
				t.Fatal(first.Body.String())
			}
			second := send(message.Content[0].ID, true)
			if second.Code != 200 || !strings.Contains(second.Body.String(), "verified answer") || !strings.Contains(second.Body.String(), "event: message_stop") {
				t.Fatal(second.Code, second.Body.String())
			}
			replay := send(message.Content[0].ID, true)
			if replay.Code == 200 || strings.Contains(replay.Body.String(), "message_stop") {
				t.Fatal(replay.Code, replay.Body.String())
			}
		})
	}
}
