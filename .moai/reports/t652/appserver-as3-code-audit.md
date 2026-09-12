# AS3 변경분 독립 코드 감사

SPEC: SPEC-MOAI-GATEWAY-001 0.11.0, AS-007~009

Card: t652

Overall Verdict: **FAIL**

Overall Score: **0.575** — Functionality/Security must-pass 실패

Iteration: 1

## Claim

기존 시험은 통과했지만 실제 codexapp 하위 프로세스를 사용하는 독립 반례에서 취소 직후 배정된 turn 중단 누락과 취소 RPC 누적으로 다른 대화까지 실패하는 문제가 재현됐다. AS3는 통과하지 못했다. 이 보고서는 전체 제품이나 실제 GPT 구독 성공 판정이 아니다.

## Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---:|---|---|
| Functionality (40%) | 50/100 | FAIL | `Step err=context canceled transport err=<nil> interrupt observed=false` |
| Security (25%) | 50/100 | FAIL | `conversation c failed after earlier canceled calls: invalid Codex App Server protocol message; transport=invalid Codex App Server protocol message` |
| Craft (20%) | 50/100 | FAIL | `coverage: 77.9% of statements` |
| Consistency (15%) | 100/100 | PASS | vet, gofmt, diff check stdout 빈 문자열, exit 0 |

적용 profile은 t652의 `.moai/config/evaluator-profiles/default.md`다. SPEC frontmatter에 evaluator_profile이 없고 harness.default_profile은 default다. 해당 profile은 85% 미만을 Craft FAIL로 명시한다. 위험 분기 시험을 우선하되 이 기준을 임의로 면제하지 않았다.

## Findings

- **AS3-F1 / High / blocking / confidence: high** — `internal/codexbridge/engine.go:379`, `:387`, `:480`; `internal/codexapp/client.go:369`. turn/start가 서버에 도착해 turn이 배정된 뒤 HTTP context가 취소되면 Client.Call은 늦은 응답의 상관관계를 버리고 context canceled를 반환한다. Engine은 c.turn을 얻기 전에 종료하므로 deferred interrupt가 빈 turn ID에서 생략된다. 실제 프로세스 fixture에서 배정 기록을 확인한 직후 취소하고 늦은 응답 후까지 관찰했지만 interrupt는 0이었다. **Required fix:** 제한된 수명 안에서 시작 응답의 소유 ID를 회수하여 해당 turn을 interrupt하고 정리한다. 공유 transport 전체 종료를 정상 취소의 대안으로 쓰지 않는다. 기존 fakeRPC 시험은 context를 무시하므로 실제 Client 경계 회귀가 필요하다. AS-009 위반이다.
- **AS3-F2 / High / blocking / confidence: high** — `internal/codexbridge/engine.go:485`, `:170`; `internal/codexapp/client.go:218`, `:394`. Cancel은 turn/interrupt만 보내고 Client의 pending server RPC 등록을 소비 또는 폐기하지 않는다. stopped 대화의 늦은 RPC도 등록 해제 없이 버린다. 허용된 QueueSize=2로 A와 B의 도구 호출을 차례로 취소하면 C의 호출 때 등록 상한에 걸려 공유 transport가 protocol error로 종료된다. **Required fix:** 소유 turn의 보류·대기열·늦은 RPC를 한 번만 정리하는 경계를 추가한다. 두 취소 뒤 별도 대화가 계속 성공하는 실제 subprocess 회귀가 필요하다. 상한 확대만으로는 누적 원인이 사라지지 않는다.
- **AS3-F3 / Medium / blocking (profile gate) / confidence: high** — `internal/codexbridge/engine.go:131`, `internal/codexbridge/store.go:65`, `internal/gateway/appserver.go:33`, `internal/codexapp/private.go:10`. 신규 bridge 전체 77.9%, private helper 69.2%, adapter Send 66.7%로 기본 profile의 85% 기준에 미달한다. **Required fix:** F1/F2 실제 경계와 읽기 queue 초과, 권한 오류, 저장 실패 및 성공 terminal 부재를 의미 있는 회귀로 보강하고 변경 소스의 분모·분자를 별도로 공개한다. 기존 gateway 전체의 filtered coverage를 변경 코드의 커버리지로 오인해서는 안 된다.

## Evidence

작업 디렉터리는 아래 Baseline과 같다. 독립 점검은 병렬 batch로 실행했다.

```text
$ go test -race ./internal/codexbridge ./internal/gateway -run 'Test(ToolSegment|TwoConversation|RestartPending|PendingBarrier|EOF|OutputLimit|CancelBefore|UntrustedTurn|ContextCancellation|ResultBarrier|UnsafeState|ScopeDelimiter|Managed|AppServer)' -count=1 -timeout=90s
ok  	github.com/modu-ai/moai-adk/internal/codexbridge	3.449s
ok  	github.com/modu-ai/moai-adk/internal/gateway	2.995s
exit 0

$ go test ./internal/codexbridge ./internal/codexapp ./internal/gateway -run 'Test(ToolSegment|TwoConversation|RestartPending|PendingBarrier|EOF|OutputLimit|CancelBefore|UntrustedTurn|ContextCancellation|ResultBarrier|UnsafeState|ScopeDelimiter|Managed|AppServer|ValidatePrivate)' -coverprofile=<audit-temp>/coverage.out -count=1 -timeout=90s
ok  	github.com/modu-ai/moai-adk/internal/codexbridge	2.576s	coverage: 77.9% of statements
ok  	github.com/modu-ai/moai-adk/internal/codexapp	0.926s	coverage: 55.5% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway	2.277s	coverage: 16.3% of statements
exit 0

$ go vet ./internal/codexbridge ./internal/codexapp ./internal/gateway
$ gofmt -l internal/codexbridge internal/codexapp/private.go internal/codexapp/private_test.go internal/gateway/appserver*.go
$ git diff --check
stdout: ""
exit 0
```

보안 점검은 OWASP 객체 귀속·인증·자원 상한·외부 프로토콜 검증에 대해 코드 판독과 runtime 음성군을 결합했다. 신규 production bridge/appserver 파일의 credential 형상, shell 실행, auth.json/token 필드, HTTP client 호출을 Python 정규식으로 검색하고 manifest diff도 확인했다. 검색의 부재만으로 안전을 판정하지 않았다.

```text
credential_literal=[]
shell_exec=[]
auth_file=[]
external_network=[]
manifest_changes=
exit 0
```

독립 반례는 Go overlay로 추가했다. 저장소 구현 파일은 수정하지 않았다. fixture는 별도 private 임시 디렉터리와 제한된 context 및 defer client.Close를 사용하며 실제 GPT를 호출하지 않는다.

```text
$ go test -overlay=<audit-temp>/overlay.json ./internal/codexbridge -run TestAuditRealTransportEOFWakesBridge -v -count=1 -timeout=10s
=== RUN   TestAuditRealTransportEOFWakesBridge
    audit_external_test.go:24: Step err=Codex App Server EOF transport err=Codex App Server EOF elapsed=308.094375ms done=false
--- PASS: TestAuditRealTransportEOFWakesBridge (1.48s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/codexbridge	2.786s

$ go test -overlay=<audit-temp>/overlay.json ./internal/codexbridge -run TestAuditCancelBeforeRealStartReplyInterruptsTurn -v -count=1 -timeout=10s
=== RUN   TestAuditCancelBeforeRealStartReplyInterruptsTurn
    audit_external_test.go:58: Step err=context canceled transport err=<nil> interrupt observed=false
    audit_external_test.go:59: allocated turn was not interrupted after late turn/start reply
--- FAIL: TestAuditCancelBeforeRealStartReplyInterruptsTurn (1.38s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/codexbridge	1.784s
FAIL
exit 1

$ go test -overlay=<audit-temp>/overlay.json ./internal/codexbridge -run TestAuditCanceledRPCDoesNotExhaustSharedTransport -v -count=1 -timeout=10s
=== RUN   TestAuditCanceledRPCDoesNotExhaustSharedTransport
    audit_external_test.go:91: conversation c failed after earlier canceled calls: invalid Codex App Server protocol message; transport=invalid Codex App Server protocol message
--- FAIL: TestAuditCanceledRPCDoesNotExhaustSharedTransport (0.73s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/codexbridge	1.160s
FAIL
exit 1
```

NUL 구분자 충돌 우려는 현재 validOwner와 TestScopeDelimiterCannotAliasAnotherConversation 실행에서 배제했다. OpenStore는 입력 symlink를 정규화한 뒤 실제 private target을 고정한다. 지금 계약에서 그 자체를 결함으로 입증하지 못했다. Client.Events가 EOF 때 닫히지 않는다는 초기 의심은 client.read의 defer close와 실제 EOF PASS로 철회했다.

## Baseline-attribution

- WT: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t652`
- Branch: `WT-gateway-appserver-turns`
- HEAD: `342efdcfd44ada42dad495682455c3e9019510c6`
- source_session_id: `01a08e7b-6aa0-7361-ab7e-ea8da1f02228` — moai session current 재조회
- Canonical contract: legacy WT의 SPEC-MOAI-GATEWAY-001 AS-007~009 및 AS3 계획 재감사. API도 App Server 출력 정책 사용이라는 후속 승인을 포함한다.
- audit-temp: `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/as3-audit-q276wdzt`
- 검사 종료 해시 대조: `BASELINE_CHANGED []`

## Gaps

실제 구독/API 모델 응답, 정식 Claude history Prepare, 부모·자식 product wiring, CLI 배포는 미검증이다. Windows runtime과 프로세스 강제 종료 후 복구는 GitHub CI 소관으로 남는다. AS-008의 각 지점 강제 kill은 현재 기록·재생 차단 단위시험만 있으므로 전체 실증 완료로 보지 않는다. manifest 변화는 없지만 취약점 데이터베이스 전수 검사는 실행하지 않았다. 전체 OWASP 제품 검증도 아니다.

## Residual-risk

trusted Prepare와 generation 소유자는 AS4/AS5에서 실제 세션·계정 귀속을 제공해야 한다. synthetic 값으로 성공한 HTTP 시험은 실제 인증 귀속을 증명하지 않는다. 계정 교체와 turn 시작 사이의 동시성은 소유자의 직렬화 계약까지 이어 검증해야 한다. 버퍼링은 성공 terminal 안전성을 위한 것이며 실시간 token streaming이나 정확한 usage를 보장하지 않는다.

## Iteration history

1. FAIL 0.575: AS3-F1/F2 실제 subprocess 반례와 AS3-F3 coverage gate. 재감사는 세 결함과 직접 회귀로 제한한다. 기존 verdict.md는 수정하지 않았다.

## Baseline hashes

```json
{
  "internal/codexbridge/store.go": "3d373326b67b98c94ff1e2d95f2e5d2a2e05da19fb0186381225916bfabe530d",
  "internal/codexbridge/engine_test.go": "9129cb35ff1853006e2e32aaeab9ce93483d3de193b239cd32f4082f8f32eee4",
  "internal/codexbridge/engine.go": "e31fc9c5d25c629e1c1efcc1c484e908264d84fa421999cc36a836224bdf6fc8",
  "internal/gateway/appserver_test.go": "4ce3a349fa7bfe81af17c90af31332740dbfa9af76eab69a11da509245210f54",
  "internal/gateway/appserver.go": "cb609c9c2a678ece6dcba548c886270c51031718d5a07563e6274dade1c1cedf",
  "internal/gateway/appserver_integration_test.go": "51adba7b745f7195587aba6b302804b4934abafc3f7504a266109803ffc9b607",
  "internal/gateway/appserver_authority.go": "754842b2d33180d30c9c6d2d66591b053e1087e35fa9e2efece9ca379a971056",
  "internal/gateway/catalog.go": "f8b6b86601d830bb0f85d85a423e2dd4f53f83f57acc55c8b30ba73c608dc727",
  "internal/gateway/router.go": "bef1e965722b34a5f082f3beaedcbebe8fe72bfa4bad8e4f5d02a7f0c05dfcf1",
  "internal/gateway/server.go": "54d4b44fcd706f8fde933bedeefe047318c46726279e2975ff74d6ecd0c1f32e",
  "internal/codexapp/private.go": "130fdbf34c8bb441372d6d5c355c853eb9a842f10700c43371f36fad571c67a7",
  "internal/codexapp/private_test.go": "5b652ef02eba2cfdad47cffdb426f110d895cc0b88d87614d4415cbe5ed7dad1"
}
```

## Reproduction source

```go
package codexbridge
import("context";"testing";"time";"os";"os/exec";"path/filepath";"errors";"github.com/modu-ai/moai-adk/internal/codexapp")
func TestAuditRealTransportEOFWakesBridge(t *testing.T){
 py,err:=exec.LookPath("python3");if err!=nil{t.Skip(err)}
 dir,_:=filepath.EvalSymlinks(t.TempDir());os.Chmod(dir,0700)
 script:=filepath.Join(dir,"fake-codex")
 body:=`import sys,json,time
for line in sys.stdin:
 m=json.loads(line)
 if m.get('method')=='initialize': print(json.dumps({'id':m['id'],'result':{'userAgent':'audit'}}),flush=True)
 elif m.get('method')=='thread/start': print(json.dumps({'id':m['id'],'result':{'thread':{'id':'eof-thread'}}}),flush=True)
 elif m.get('method')=='turn/start':
  print(json.dumps({'id':m['id'],'result':{'turn':{'id':'eof-turn'}}}),flush=True)
  time.sleep(0.15)
  sys.exit(0)
`
 os.WriteFile(script,[]byte("#!"+py+"\n"+body),0700)
 life,cancel:=context.WithTimeout(context.Background(),5*time.Second);defer cancel()
 client,err:=codexapp.Start(life,codexapp.Config{Binary:script,Home:dir});if err!=nil{t.Fatal(err)};defer client.Close()
 if _,err=client.Initialize(life,"audit","1");err!=nil{t.Fatal(err)}
 state:=filepath.Join(dir,"state");os.Mkdir(state,0700);store,err:=OpenStore(state);if err!=nil{t.Fatal(err)}
 e,err:=New(life,client,Config{Store:store});if err!=nil{t.Fatal(err)};defer e.Close()
 ctx,stop:=context.WithTimeout(context.Background(),750*time.Millisecond);defer stop();start:=time.Now()
 seg,err:=e.Step(ctx,request("eof-real"));t.Logf("Step err=%v transport err=%v elapsed=%v done=%v",err,client.Err(),time.Since(start),seg.Done)
 if errors.Is(err,context.DeadlineExceeded){t.Fatal("transport EOF did not wake bridge; waited for HTTP deadline")}
 if err==nil||seg.Done{t.Fatal("EOF synthesized success")}
}

func TestAuditCancelBeforeRealStartReplyInterruptsTurn(t *testing.T){
 py,err:=exec.LookPath("python3");if err!=nil{t.Skip(err)}
 dir,_:=filepath.EvalSymlinks(t.TempDir());os.Chmod(dir,0700)
 script:=filepath.Join(dir,"fake-codex")
 body:=`import sys,json,time,os
for line in sys.stdin:
 m=json.loads(line)
 if m.get('method')=='initialize': print(json.dumps({'id':m['id'],'result':{'userAgent':'audit'}}),flush=True)
 elif m.get('method')=='thread/start': print(json.dumps({'id':m['id'],'result':{'thread':{'id':'eof-thread'}}}),flush=True)
 elif m.get('method')=='turn/start':
  open(os.path.join(os.environ['CODEX_HOME'],'allocated'),'w').write('yes')
  time.sleep(0.25)
  print(json.dumps({'id':m['id'],'result':{'turn':{'id':'eof-turn'}}}),flush=True)
 elif m.get('method')=='turn/interrupt':
  open(os.path.join(os.environ['CODEX_HOME'],'interrupted'),'w').write('yes')
  print(json.dumps({'id':m['id'],'result':{}}),flush=True)
`
 os.WriteFile(script,[]byte("#!"+py+"\n"+body),0700)
 life,cancel:=context.WithTimeout(context.Background(),5*time.Second);defer cancel()
 client,err:=codexapp.Start(life,codexapp.Config{Binary:script,Home:dir});if err!=nil{t.Fatal(err)};defer client.Close()
 if _,err=client.Initialize(life,"audit","1");err!=nil{t.Fatal(err)}
 state:=filepath.Join(dir,"state");os.Mkdir(state,0700);store,err:=OpenStore(state);if err!=nil{t.Fatal(err)}
 e,err:=New(life,client,Config{Store:store});if err!=nil{t.Fatal(err)};defer e.Close()
 ctx,stop:=context.WithCancel(context.Background());defer stop()
 done:=make(chan error,1);go func(){_,err:=e.Step(ctx,request("cancel-real"));done<-err}()
 deadline:=time.NewTimer(2*time.Second);defer deadline.Stop()
 for {if _,err:=os.Stat(filepath.Join(dir,"allocated"));err==nil{break}; select {case <-deadline.C:t.Fatal("allocation never started");case <-time.After(10*time.Millisecond):}}
 stop();err=<-done;if !errors.Is(err,context.Canceled){t.Fatal(err)}
 time.Sleep(450*time.Millisecond)
 _,interrupted:=os.Stat(filepath.Join(dir,"interrupted"));t.Logf("Step err=%v transport err=%v interrupt observed=%v",err,client.Err(),interrupted==nil)
 if interrupted!=nil{t.Fatal("allocated turn was not interrupted after late turn/start reply")}

}

func TestAuditCanceledRPCDoesNotExhaustSharedTransport(t *testing.T){
 py,err:=exec.LookPath("python3");if err!=nil{t.Skip(err)}
 dir,_:=filepath.EvalSymlinks(t.TempDir());os.Chmod(dir,0700)
 script:=filepath.Join(dir,"fake-codex")
 body:=`import sys,json
counter=0
def emit(x): print(json.dumps(x),flush=True)
for line in sys.stdin:
 m=json.loads(line); method=m.get('method');p=m.get('params',{})
 if method=='initialize': emit({'id':m['id'],'result':{'userAgent':'audit'}})
 elif method=='thread/start':
  counter+=1
  emit({'id':m['id'],'result':{'thread':{'id':'thread-'+str(counter)}}})
 elif method=='turn/start':
  th=p['threadId'];tr='turn-'+th
  emit({'id':m['id'],'result':{'turn':{'id':tr}}})
  emit({'id':'rpc-'+th,'method':'item/tool/call','params':{'threadId':th,'turnId':tr,'callId':'call-'+th,'tool':'echo','arguments':{'text':'hello'}}})
 elif method=='turn/interrupt':
  emit({'id':m['id'],'result':{}})
  emit({'method':'turn/completed','params':{'threadId':p['threadId'],'turn':{'id':p['turnId'],'status':'interrupted'}}})
`
 os.WriteFile(script,[]byte("#!"+py+"\n"+body),0700)
 life,cancel:=context.WithTimeout(context.Background(),5*time.Second);defer cancel()
 client,err:=codexapp.Start(life,codexapp.Config{Binary:script,Home:dir,QueueSize:2});if err!=nil{t.Fatal(err)};defer client.Close()
 if _,err=client.Initialize(life,"audit","1");err!=nil{t.Fatal(err)}
 state:=filepath.Join(dir,"state");os.Mkdir(state,0700);store,err:=OpenStore(state);if err!=nil{t.Fatal(err)}
 e,err:=New(life,client,Config{Store:store});if err!=nil{t.Fatal(err)};defer e.Close()
 for _,id:=range []string{"a","b","c"}{
 q:=request(id);seg,err:=e.Step(life,q);if err!=nil{t.Fatalf("conversation %s failed after earlier canceled calls: %v; transport=%v",id,err,client.Err())};if seg.Tool==nil{t.Fatal("no tool")}
 if err=e.Cancel(life,q.Owner);err!=nil{t.Fatal(err)}
 }
}

```
