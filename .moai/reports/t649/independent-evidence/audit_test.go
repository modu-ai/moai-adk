package translate
import("encoding/json";"testing")
func TestAuditReasoningPublicItemOrderAndPhase(t *testing.T) {
 h:=&testHistory{}
 _,c,e:=Request("gpt-5.6-sol",[]byte(`{"max_tokens":100,"messages":[{"role":"user","content":"hello"}]}`),Limits{History:h});if e!=nil{t.Fatal(e)}
 first:=textItem("msg_c","commentary");first["phase"]="commentary"
 last:=textItem("msg_f","final");last["phase"]="final_answer"
 rs:=func(id string)map[string]any{return map[string]any{"type":"reasoning","id":id,"summary":[]any{},"encrypted_content":"cipher_"+id}}
 output,e:=c.Response(j(finalResponse("completed",rs("rs_1"),first,rs("rs_2"),last)));if e!=nil{t.Fatal(e)}
 result:=decode(t,output)
 input:=map[string]any{"max_tokens":100,"messages":[]any{map[string]any{"role":"user","content":"hello"},map[string]any{"role":"assistant","content":result["content"]},map[string]any{"role":"user","content":"continue"}}}
 replay,_,e:=Request("gpt-5.6-sol",j(input),Limits{History:h});if e!=nil{t.Fatal(e)}
 var root map[string]any;json.Unmarshal(replay,&root)
 rows:=root["input"].([]any)
 t.Logf("observed upstream replay: %s",replay)
 if len(rows)!=6 {t.Fatalf("public output item boundaries lost: got %d total input items, want 6",len(rows))}
 if rows[2].(map[string]any)["phase"]!="commentary"||rows[4].(map[string]any)["phase"]!="final_answer" {t.Fatal("public message phase lost")}
}
func TestAuditReasoningAfterTwoMessages(t *testing.T) {
 h:=&testHistory{}
 _,c,e:=Request("gpt-5.6-sol",[]byte(`{"max_tokens":100,"messages":[{"role":"user","content":"hello"}]}`),Limits{History:h});if e!=nil{t.Fatal(e)}
 rs:=map[string]any{"type":"reasoning","id":"rs_3","summary":[]any{},"encrypted_content":"cipher"}
 output,e:=c.Response(j(finalResponse("completed",textItem("msg_1","a"),textItem("msg_2","b"),rs,textItem("msg_3","c"))));if e!=nil{t.Fatal(e)}
 result:=decode(t,output)
 input:=map[string]any{"max_tokens":100,"messages":[]any{map[string]any{"role":"user","content":"hello"},map[string]any{"role":"assistant","content":result["content"]},map[string]any{"role":"user","content":"continue"}}}
 _,_,e=Request("gpt-5.6-sol",j(input),Limits{History:h});if e!=nil{t.Fatalf("previous successful output cannot be replayed: %v",e)}
}
