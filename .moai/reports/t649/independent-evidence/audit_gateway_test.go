package gateway
import("context";"encoding/json";"io";"net/http";"strings";"testing";"github.com/modu-ai/moai-adk/internal/gateway/opaque")
type auditHistory struct{}
func(auditHistory)Check(context.Context,string,string,[]byte)error{return nil}
func(auditHistory)Publish(context.Context,string,string,[]byte)error{return nil}
func TestAuditNativeEgressPreservesOpaqueItemBytes(t *testing.T){
 raw:=`{ "type":"reasoning", "id":"rs_audit", "summary":[], "encrypted_content":"synthetic_cipher" }`
 env,e:=opaque.Encode([]opaque.Item{{OutputIndex:0,Raw:[]byte(raw)}});if e!=nil{t.Fatal(e)}
 input:=map[string]any{"model":"gpt-5.6-sol","max_tokens":10,"messages":[]any{map[string]any{"role":"user","content":"hello"},map[string]any{"role":"assistant","content":[]any{map[string]any{"type":"redacted_thinking","data":env.Data()},map[string]any{"type":"text","text":"answer"}}},map[string]any{"role":"user","content":"continue"}}}
 body,_:=json.Marshal(input)
 var sent string
 tr,calls:=oaiTLS(t,func(w http.ResponseWriter,r *http.Request){b,_:=io.ReadAll(r.Body);sent=string(b);w.Header().Set("Content-Type","text/event-stream");io.WriteString(w,oaiSSE())})
 store,ref,gen:=oaiStore(t);cfg:=oaiConfig(tr);cfg.Subscription=store;cfg.Limits.History=auditHistory{}
 adapter,e:=NewOpenAIAdapter(cfg);if e!=nil{t.Fatal(e)}
 resp,e:=adapter.Send(context.Background(),RoutedRequest{Entry:oaiEntry(AuthPKCE),Body:body,Credential:ref,Generation:gen});oaiRead(t,resp,e)
 if calls.Load()!=1||resp.StatusCode!=200{t.Fatalf("probe failed status=%d calls=%d",resp.StatusCode,calls.Load())}
 t.Logf("observed subscription request: %s",sent)
 if !strings.Contains(sent,raw){t.Fatal("opaque item original bytes lost at subscription egress")}
}
