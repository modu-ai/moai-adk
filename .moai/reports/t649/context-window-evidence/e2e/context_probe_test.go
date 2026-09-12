package cli

import (
 "bufio"
 "bytes"
 "context"
 "encoding/json"
 "fmt"
 "io"
 "net/http"
 "os"
 "strconv"
 "strings"
 "testing"
 "time"
 "github.com/modu-ai/moai-adk/internal/gateway/auth"
)
func TestT649ContextLive(t *testing.T) {
 model:=os.Getenv("T649_MODEL"); n,e:=strconv.Atoi(os.Getenv("T649_WORDS")); if e!=nil||n<1||n>1200000 {t.Fatal("invalid count")}
 start:=time.Now(); result:=map[string]any{"model":model,"repetitions":n,"unit":" x","timeout_seconds":150,"body_limit_bytes":16<<20,"route":"subscription","truncation":"server default, three-position recall checked"}
 defer func(){result["duration_seconds"]=time.Since(start).Seconds(); b,_:=json.MarshalIndent(result,"","  "); if e:=os.WriteFile(os.Getenv("T649_RESULT"),b,0600);e!=nil {t.Fatal(e)};fmt.Println(string(b))}()
 store,e:=openGPTAuthStore(); if e!=nil {result["error"]="store open failed";return};defer store.Close()
 broker,e:=installedGPTBrokerForGateway();if e!=nil {result["error"]="broker unavailable";return}
 ctx,cancel:=context.WithTimeout(context.Background(),150*time.Second);defer cancel()
 ref,e:=store.ResolveFresh(ctx,broker,verifyGPTCredentialOwnership);if e!=nil {result["error"]="credential resolve failed";return};gen,e:=ref.Generation();if e!=nil {result["error"]="generation failed";return}
 text:="START_NEEDLE=MAPLE71\n"+strings.Repeat(" x",n/2)+"\nMIDDLE_NEEDLE=OTTER83\n"+strings.Repeat(" x",n-n/2)+"\nEND_NEEDLE=CEDAR29\nReturn only the three needle values in their order, separated by commas."
 payload,_:=json.Marshal(map[string]any{"model":model,"stream":true,"store":false,"instructions":"Read the entire user input. Return only the requested three values. Do not explain.","reasoning":map[string]string{"effort":"medium"},"input":[]any{map[string]any{"role":"user","content":[]any{map[string]string{"type":"input_text","text":text}}}}})
 result["request_bytes"]=len(payload)
 req,_:=http.NewRequestWithContext(ctx,"POST",auth.SubscriptionEndpoint,bytes.NewReader(payload));req.Header.Set("Content-Type","application/json");req.Header.Set("Accept","text/event-stream")
 resp,e:=store.SendAuthorized(ctx,gen,req,&http.Transport{},auth.SendOptions{WriteTimeout:30*time.Second,PollInterval:100*time.Millisecond,MaxBodyBytes:16<<20});if e!=nil {result["error"]="authorized send failed";result["timeout"]=ctx.Err()!=nil;return};defer resp.Body.Close();result["http_status"]=resp.StatusCode
 safeError:=func(raw json.RawMessage){var m map[string]any;if json.Unmarshal(raw,&m)!=nil{return};for _,k:=range []string{"code","type","message","param"}{if v,ok:=m[k].(string);ok&&len(v)<1024 {result["server_error_"+k]=v}}}
 if resp.StatusCode!=200 {b,_:=io.ReadAll(io.LimitReader(resp.Body,8192));var m map[string]json.RawMessage;if json.Unmarshal(b,&m)==nil{safeError(m["error"])};return}
 scan:=bufio.NewScanner(io.LimitReader(resp.Body,2<<20));scan.Buffer(make([]byte,4096),1<<20);output:="";events:=0
 for scan.Scan(){line:=scan.Text();if !strings.HasPrefix(line,"data: "){continue};var ev struct{Type string `json:"type"`;Delta string `json:"delta"`;Response struct{Status string `json:"status"`;Usage json.RawMessage `json:"usage"`;Error json.RawMessage `json:"error"`} `json:"response"`;Error json.RawMessage `json:"error"`};if json.Unmarshal([]byte(strings.TrimPrefix(line,"data: ")),&ev)!=nil{continue};events++;if ev.Type=="response.output_text.delta"&&len(output)<1024 {output+=ev.Delta};if ev.Type=="response.completed"||ev.Type=="response.failed"{result["event"]=ev.Type;result["response_status"]=ev.Response.Status;if len(ev.Response.Usage)>0{result["usage"]=ev.Response.Usage};safeError(ev.Response.Error)};if ev.Type=="error"{safeError(ev.Error)}}
 result["events"]=events;result["timeout"]=ctx.Err()!=nil;result["stream_read_ok"]=scan.Err()==nil;result["output"]=output;result["needle_recall"]=strings.Contains(output,"MAPLE71")&&strings.Contains(output,"OTTER83")&&strings.Contains(output,"CEDAR29")
}
