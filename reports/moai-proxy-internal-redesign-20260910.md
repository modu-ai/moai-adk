---
title: "moai-proxy 내부 package 상세 재설계"
date: "2026-09-10"
lang: ko
---

## 1. 설계 결정과 적용 범위

사용자의 최신 지시에 따라 **같은 Go module의 `internal/proxy`에 구현하고 기존 `moai` 바이너리에 포함**한다. `moai-proxy`는 기능 이름이며 별도 다운로드 대상 executable이 아니다. 기존 companion 보고서의 분리 저장소·별도 installer·별도 version handshake·proxy child supervision 계획은 이 설계로 대체한다.

기본 실행은 `moai` 프로세스 안에서 loopback HTTP 서버를 goroutine으로 구동하고 Claude Code만 자식 프로세스로 실행하는 구조다. 사용자는 MoAI 설치 후 `moai proxy login codex`로 최초 인증을 하고, `moai cc --provider codex`로 시작한다. 유효한 Codex 파일 인증이 있는 사용자는 읽기 전용 재사용을 명시적으로 선택할 수 있다. 별도 proxy 바이너리 설치는 없다. **이 사용자 경험은 제안이며 아직 구현·설치 실측 결과가 아니다.**

| 구분 | 확정한 설계 |
|---|---|
| 배포 단위 | 기존 `cmd/moai` 단일 바이너리 |
| 코드 경계 | `internal/proxy`; CLI와 config package가 조립하고 proxy는 cli/config를 import하지 않음 |
| 기본 실행 | in-process HTTP 서버 + supervised Claude 자식 |
| 기본 주소 | `127.0.0.1:0`, OS가 실제 포트 배정 |
| GPT backend | Codex subscription Responses adapter; 공개 OpenAI API adapter와 혼동 금지 |
| GLM backend | Anthropic-compatible Z.AI pass-through adapter |
| 혼합 모드 | Codex와 GLM만 명시적으로 활성화; 기본 세션은 단일 provider |
| Claude 원본 passthrough | MVP 제외. 로그인·헤더 전달 계약을 별도 입증한 뒤 추가 |
| 시스템 변경 | hosts, system CA, CONNECT interception, privileged listener 불필요 |
| 이번 산출물 | 내부 설계와 구현·검증 계획. production 코드 변경 없음 |

GPL을 기술 설계의 선행 승인 조건으로 두지 않는다. 다만 코드의 출처나 라이선스 효력이 없어지는 것은 아니므로 상류 고지와 차용 기록을 보존하고, 내부 편입물을 외부 배포할 때의 라이선스 처리는 미해결 사항으로 기록한다. 이 보고서는 법적 배포 허가를 확정하지 않는다.

## 2. 이번 소스 확인과 이전 실측의 구분

**Claim:** 현재 확인한 MoAI 실행기는 POSIX에서 `syscall.Exec`를 호출한다. 그 경로에서 in-process 서버를 시작하면 Claude 실행으로 프로세스 이미지가 교체될 때 서버가 사라진다. 따라서 proxy-enabled 실행은 spawn/wait 경로가 필요하다.

**Evidence — 이번 실행 명령과 출력 발췌:**

```text
$ git rev-parse --short HEAD
2213871af
$ git branch --show-current
main
$ moai session current
01a0896e-4701-7cb0-bd4e-7064dabb0e26
$ rg -n 'outputStyle' .claude/settings*.json
.claude/settings.json:362:  "outputStyle": "MoAI-Easy",
.claude/settings.local.json:40:  "outputStyle": "MoAI-Easy",
$ sed -n '1,180p' internal/cli/launch_exec_posix.go
func execOrSpawnClaude(claudeBin string, args, env []string) error {
    return syscall.Exec(claudeBin, args, withSessionPID(env, os.Getpid()))
}
```

**Baseline-attribution:** MoAI working tree `main@2213871af`를 읽었다. dirty checkout 전체를 해당 commit의 깨끗한 tree와 같다고 주장하지 않는다. 상류 임시 clone은 `112588175eb2b3b693a5bd23d58d8c501e0ae406`; 이번 `git status --short`는 출력이 없었다. 소스 발췌의 들여쓰기는 읽기 쉽게 정리했다.

상류에서 이번에 다시 읽은 로직은 다음과 같다.

```text
$ git -C <upstream-clone> rev-parse HEAD
112588175eb2b3b693a5bd23d58d8c501e0ae406
$ sed -n '1,130p' internal/codex/responses.go
req.Header.Set("Authorization", "Bearer "+accessToken)
req.Header.Set("Session_id", sessionID)
$ sed -n '150,292p' internal/codex/store.go
refreshed, err := s.client.Refresh(context.WithoutCancel(ctx), token.RefreshToken)
if err := os.WriteFile(target.path, data, 0o600); err != nil {
$ sed -n '1,100p' internal/translate/claude_to_codex.go
const SignatureMarker = "ccmp1:"
body, _ = sjson.SetBytes(body, "store", false)
body, _ = sjson.SetBytes(body, "stream", true)
```

위 각 source read의 exit code는 0이다. 이는 로직 존재 확인이지 취약점 재현·런타임 호환성의 새 증거는 아니다.

**이전 세션 증거:** GPT-6 응답·picker·Read 왕복, GLM 응답·Read 왕복, HTTP-only 호출, unit/race/vet, 상류 main의 취약점 15건은 [이전 실측 원장](moai-proxy-feasibility-plan-20260910.md)에 기록돼 있다. 이번 설계 턴에서는 유료 호출과 취약점 검사를 반복하지 않았다. 과거 15건을 현재 취약점 DB에 대한 새 측정치로 취급하지 않는다.

## 3. ccmproxy 로직별 차용과 변경

| 상류 source | 현재 역할 | 내부 재설계 | 시험으로 확인할 계약 |
|---|---|---|---|
| `cmd/ccmproxy/main.go` | CLI·CA·resolver·backend·listeners 조립 | `internal/cli/proxy.go`는 조립만; `internal/proxy/server.go`가 수명 소유 | listener 준비 후에만 Claude 시작 |
| `router/router.go` | model 검사 후 Codex 또는 Anthropic 전달 | 명시 registry로 분기; unknown은 400, disabled provider는 400 | unknown 요청의 외부 송신 0건 |
| `router/body.go` | gzip/deflate 해제·body 수정 | 압축 전/후 크기 별도 제한; duplicate model JSON 거절 | 압축 폭탄·중복 routing key 거절 |
| `translate/claude_to_codex.go` | Messages→Responses | 순수 변환 함수; raw payload와 capability profile 입력 | system/text/tool/image/PDF/schema golden |
| `translate/codex_to_claude.go` | SSE를 Claude block 이벤트로 변환 | 응답별 상태기계, 완료/실패를 명확히 분리 | terminal event 누락·중복·순서 위반 감지 |
| `translate/sanitize.go` | provider 불일치 thinking 제거 | opaque reasoning 상태에 provider/버전 표시, 텍스트·tool 보존 | GPT→GLM 전환 시 private reasoning 제거 |
| `codex/responses.go` | OAuth bearer로 Responses POST | 전용 transport·redirect 제한·bounded error body | 다른 호스트에 credential 송신 0건 |
| `codex/store.go` | auth 후보 탐색·refresh·writeback | MoAI 소유 저장소 refresh; Codex auth는 read-only | 외부 auth 파일 SHA 변화 0건 |
| `codex/models.go` | 고정 모델·effort·suffix 해석 | versioned profile + 계정 가용성 결과 + 사용자 alias | unsupported effort 조용한 강등 금지 |
| `tokenize/tokenize.go` | 문자 수 기반 추정 | 추정치임을 명시; compact-safe margin | 한국어·코드·schema corpus 오차 기록 |
| `proxy`, `ca`, `resolve`, `run.sh` | MITM·CA·DNS 우회·hosts | MVP에서 가져오지 않음; OS resolver와 표준 HTTP 사용 | 전후 hosts/trust 동일 |
| `httplog` | URL·status·bytes 기록 | request ID·provider·status·latency만; query/body/token 제외 | 로그 secret fixture 0건 |

출처는 아래 참조 목록의 SHA 고정 상류 tree다. 행별 변경은 **설계 제안**이며 구현 완료를 뜻하지 않는다.

## 4. 패키지 배치와 의존성

provider가 둘뿐인 초기 단계에서는 거대한 공통 이벤트 체계를 만들지 않는다. HTTP Handler와 작은 provider interface를 쓰고, Codex 변환만 별도 하위 package로 분리한다.

```text
internal/proxy/
  server.go                 listener·ready·shutdown·local auth
  handler.go                endpoint·body 검증·모델 해석·오류 envelope
  registry.go               alias/slot→provider/upstream model 해석
  types.go                  Options, Route, Capabilities, Provider
  policy.go                 URL·redirect·body·concurrency 제한
  catalog.go                지원 profile·계정 probe 결과
  codex.go                  Responses 호출·credential 획득
  glm.go                    Z.AI Messages passthrough
  credentials.go            전용 auth·read-only Codex adapter
  credential_lock_posix.go  MoAI 소유 store process lock
  credential_lock_windows.go
  translate/
    request.go              Messages→Responses 순수 변환
    stream.go               Responses→Messages SSE 상태기계
    aggregate.go            non-stream collector
    reasoning.go            provider 간 opaque 상태 처리
  testdata/                 credential-free protocol fixtures

internal/cli/proxy.go        serve/login/logout/status 명령
internal/cli/proxy_launch.go proxy Start→Claude spawn/wait→Close
internal/cli/proxy_launch_posix.go   terminal·signal·child 관리
internal/cli/proxy_launch_windows.go Windows job/console 종료 관리
internal/cli/doctor_proxy.go 정적 진단 + 명시 --live probe
```

기존 `internal/config` 타입을 `internal/proxy`에서 직접 읽지 않는다. CLI가 `config.ProxySettings`를 `proxy.Options`로 변환한다. 흐름은 `cli → config`와 `cli → proxy → translate`이며 역방향 import는 금지한다. 실제 파일 수는 구현 시 응집도에 따라 합칠 수 있다.

```go
// 제안 API: 아직 존재하는 코드가 아님.
type Provider interface {
    ServeMessages(context.Context, http.ResponseWriter, Request, Route) error
}
type Options struct {
    ListenAddress string
    SessionToken  string
    Providers     map[string]Provider
    Registry      *Registry
    Limits        Limits
}
type Runtime interface {
    BaseURL() string
    Done() <-chan error
    Close(context.Context) error
}
func Start(context.Context, Options) (Runtime, error)
```

`Start`는 `net.Listen`이 성공하고 handler가 준비된 뒤 반환한다. 임시 포트를 검색했다가 다시 bind하지 않는다. `Runtime.Close`는 여러 번 호출돼도 안전해야 하며 listener 닫기·in-flight 취소·idle connections 해제를 끝낸다. readiness는 account entitlement를 의미하지 않는다. 계정 상태는 `doctor --live` 또는 첫 요청에서 별도로 판정한다.

## 5. 실행 수명과 CLI 사용자 흐름

| 명령 제안 | 동작 | 설정 지속성 |
|---|---|---|
| `moai proxy login codex` | PKCE 로그인, MoAI 전용 token 저장 | 사용자 auth만 변경 |
| `moai proxy logout codex` | MoAI가 소유한 auth만 삭제 | Codex CLI auth는 유지 |
| `moai cc --provider codex` | 내장 서버 시작 후 Claude 실행 | 부모 shell·전역 settings 불변 |
| `moai cc --provider glm` | 동일 런처를 통한 GLM 단일 route | 기존 `moai glm` 유지 |
| `moai cc --provider mixed` | 활성화한 GPT/GLM registry의 picker | 명시 opt-in |
| `moai proxy serve --listen 127.0.0.1:0` | 같은 binary의 진단용 foreground 서버 | 기본 daemon/service 없음 |
| `moai proxy status --json` | config/auth 존재·catalog profile 표시 | 기본 네트워크 요청 없음 |
| `moai doctor --proxy --live` | 명시적 backend probe | synthetic prompt 여부·비용을 명시 |

실행 순서:

1. 기존 프로젝트·worktree·permission/model profile 해석을 재사용한다.
2. proxy-specific provider/model을 검증한다. `team_mode=glm`을 GPT 선택으로 덮어쓰지 않는다.
3. `crypto/rand`로 세션 token을 만들고 loopback listener를 시작한다.
4. user settings의 미관련 키를 보존하는 session overlay를 만든다. generated model aliases와 slot env만 주입한다.
5. Claude 자식에 `ANTHROPIC_BASE_URL`, local token, picker와 privacy env를 적용한다. 기존 Anthropic/ZAI bearer가 local token과 충돌하지 않도록 자식 env에서 해당 인증 키를 명시적으로 정리한다.
6. terminal input/output을 자식과 연결하고 기다린다. 기존 POSIX exec 경로는 proxy-disabled 실행에 그대로 둔다.
7. Claude 종료 시 runtime shutdown 후 그 exit code를 반환한다. 서버가 먼저 치명적으로 종료되면 소유한 Claude 자식을 종료시키고 supervisor 오류로 반환한다.

`MOAI_SESSION_PID`는 중요하다. 현재 exec 경로는 `os.Getpid()`를 넣지만 spawn 경로의 부모 PID는 Claude PID가 아니다. 새 경로는 고정된 자신 소유 wrapper entry를 self-reexec한 뒤 wrapper에서 실제 PID를 env에 설정하고 Claude로 exec하는 POSIX 방안, Windows의 생성 PID 전달 방안을 검토·시험해야 한다. 임의 `sh -c` wrapper를 만들지 않는다. launcher 구현 전에 **session PID contract test를 별도 작업으로 완료**한다.

POSIX interactive job control도 별도 계약이다. foreground process group과 SIGINT/SIGTSTP/SIGCONT 전달을 PTY 시험으로 검증한다. 단순 `Cmd.Run` 성공만으로 interactive 지원을 선언하지 않는다. 일반 종료는 supervisor가 정리하고, 부모 SIGKILL 때 in-process HTTP 서버는 OS에 의해 사라지지만 Claude orphan 방지는 추가 장치가 필요하다. MVP는 치명적 부모 종료 뒤 잔존 Claude 문제를 시험·기록하고 플랫폼별 parent-death/Job Object 또는 wrapper 감시를 적용한다.

## 6. 라우팅과 요청 경계

| endpoint | 기본 동작 | 외부 송신 |
|---|---|---|
| `HEAD /api/hello` | 로컬 200 | 없음 |
| `GET /healthz` | readiness만 반환; 계정/토큰 없음 | 없음 |
| `GET /v1/models` | 활성 catalog 반환; picker discovery와 구분 | 없음 |
| `POST /v1/messages` | 인증·body·route 검사 뒤 provider 실행 | 선택 provider 한 곳 |
| `POST /v1/messages/count_tokens` | provider별 estimate/지원 기능 | 명시 policy |
| 기타 `/api/*` | 지원 profile에서 요구하는 최소 로컬 응답 또는 404 | Anthropic 기본 전달 없음 |
| unknown method/path | 404/405 | 없음 |

`/healthz` 외 endpoint는 세션 local token을 요구한다. loopback이라고 인증을 생략하지 않는다. Host는 실제 bind된 loopback 주소만 허용하고 Origin 있는 브라우저 요청은 기본 거절한다. CORS `*`를 제공하지 않는다. 이는 같은 OS 사용자에게서 token을 숨기는 완전한 격리가 아니라 다른 로컬 client의 오접속과 browser-driven 요청을 줄이는 경계다.

routing은 prefix만 보지 않고 registry의 정확한 ID를 조회한다. `gpt-6-astra[1m]` suffix는 client 표시 속성으로 분리하며 upstream에는 base ID를 보낸다. `[1m]`이 계정의 실제 문맥 사용 권한을 만들지 않는다.

Claude main model 이외의 opus/sonnet/haiku 및 실제 client가 쓰는 추가 slot을 활성 모델로 매핑한다. session-title/compact/서브에이전트 요청에서 Claude model ID가 나온다고 Anthropic으로 넘기지 않는다. 알려진 slot alias는 session provider로 해석하고 미등록 canonical ID는 명시적 오류를 낸다. current client별 slot 목록과 이 호출들은 실사용 canary로 확인해야 한다.

**제안 초기 자원 한도**는 측정된 최적값이 아니다. 압축 body 16 MiB, 해제 후 32 MiB, SSE event 전체 4 MiB, 동시 upstream 요청 4개를 시작값으로 두고 corpus와 병렬 agent 시험 후 조정한다. chunked body에도 같은 해제 한도를 적용한다. semaphore를 body 읽기 전 획득해 동시 buffering을 제한한다. 4×32 MiB는 raw body 예산일 뿐 전체 RSS 상한은 아니며 JSON·이미지·stream 복사량을 별도 측정한다.

## 7. Messages→Responses 상세 변환

| Claude 입력 | Codex 출력 | 처리 계약 |
|---|---|---|
| system text/blocks | developer message 또는 검증된 instructions 위치 | 둘에 중복 삽입하지 않음; 순서 보존 |
| user/assistant text | input message content | role·content 순서 보존 |
| image | input_image | data URL/허용 URL 변환; proxy 자체 arbitrary URL fetch 금지 |
| PDF document | input_file | backend 지원 profile 확인; 미지원은 400 |
| tool definition | function name/description/parameters | schema enum/required/additionalProperties 의미 보존 |
| tool_use | function_call | call ID와 name 역매핑 저장 |
| tool_result | function_call_output | original call ID 복원; 실패 결과도 보존 |
| tool_choice | required/auto/none/function | parallel disable도 함께 변환 |
| thinking / effort | reasoning.effort·summary | model별 capability 밖 값 거절 |
| output_config.format | text.format JSON schema | schema 의미 손실 시 거절 |
| cache_control | provider가 지원하는 경로만 적용 | 지원 안 되면 pricing/cache hit를 날조하지 않음 |
| stop_sequences/max_tokens/temperature | backend capability별 처리 | 전달·정규화·미지원 오류 중 하나를 profile에 명시 |

backend request의 `stream:true`, `store:false`, encrypted reasoning include는 상류 구조를 참고한다. client가 `stream:false`여도 backend는 stream을 읽고 완료된 결과를 collector로 합친다. 완성되지 않은 tool argument JSON을 정상 응답으로 반환하지 않는다.

긴 tool 이름은 해시 suffix로 제한 길이 안에 매핑하고 같은 request에서 충돌 검사를 한다. 서로 다른 긴 이름이 같아지면 변환 실패로 끝낸다. 이름 역매핑과 call-ID 매핑은 응답 단위 상태에 두며 전역 map으로 공유하지 않는다.

## 8. SSE 상태기계와 오류

서버 전송 이벤트(SSE)는 텍스트뿐 아니라 도구 인자 조각과 종료 이유를 순서대로 전달한다. 단순 문자열 치환으로 구현하지 않는다.

| 상태/입력 | 내보낼 Claude event | 불변 조건 |
|---|---|---|
| INIT + response.created | message_start | 정확히 1회 |
| 텍스트 delta | content_block_start, text_delta | block index 단조 증가 |
| reasoning summary delta | thinking block/delta | 암호화 상태는 완료 시에만 signature로 추가 |
| function call added/delta/done | tool_use + input_json_delta + block_stop | 여러 call interleaving을 call ID별 큐로 직렬화 |
| response.completed | 열린 block 닫기, usage, message_delta, message_stop | 정상 terminal 1회 |
| response.incomplete | 이유별 max_tokens 또는 error | 완료와 구분 |
| response.failed/error | stream 전이면 HTTP error, 이후 SSE error | error 뒤 성공 message_stop 금지 |
| terminal 없는 EOF | upstream_truncated error | EOF를 성공으로 해석 금지 |
| client disconnect | upstream context 취소·body close | downstream write 종료 |

line 한도와 event 누적 한도를 함께 둔다. 여러 짧은 `data:` 줄이 합쳐진 event가 한도를 우회하지 않아야 한다. heartbeat는 무출력 연결 유지용으로만 내보내고 upstream idle timeout은 실제 upstream 데이터가 와야 갱신한다. heartbeat 자체가 멈춘 backend를 영구히 살려 두지 않게 한다.

HTTP status 확정 전 401/403/429/5xx는 Anthropic error envelope에 status·안전한 메시지를 보존한다. `Retry-After`는 유효한 값만 전달한다. 이미 stream을 내보냈거나 backend가 요청을 수신했을 수 있는 상황에서 자동 POST 재시도를 하지 않는다. 클라이언트 재시도와 중복 생성·과금 가능성을 겹치지 않게 한다.

실제 1회 기본 예: Claude가 Read 도구를 요구한 응답의 `tool_use.id=call_A`를 받으면 Claude가 실행하고 다음 요청의 `tool_result.tool_use_id=call_A`를 보낸다. translator는 이를 같은 `call_A`의 `function_call_output`으로 보내며, 다음 text 완료를 `end_turn`으로 반환한다. 이 예는 규약 설명이고 이번 턴의 새 실호출 로그가 아니다.

## 9. 인증·세션·모델 상태

| 인증 source | 읽기 | refresh | 기록 위치 | 실패 동작 |
|---|---|---|---|---|
| MoAI 전용 Codex OAuth | 가능 | MoAI가 소유 | 사용자 data dir 아래 auth/codex.json | bounded retry 후 login 안내 |
| Codex CLI auth.json | 명시 선택 시 | 금지 | 없음 | 파일 재읽기; 만료면 Codex login 또는 MoAI login 안내 |
| Codex keyring/ephemeral | adapter 구현 전 지원 안 함 | 없음 | 없음 | 지원 안 되는 저장 방식 명시 |
| Z.AI | 기존 glmcred helper로 획득 | 해당 없음 | 기존 사용자 credential store | 키 누락 진단 |
| local gateway token | session 생성 | 불필요 | 기본 메모리·child env | mismatch 401 |

**중요한 교정:** MoAI가 Codex 파일 옆에 lock을 만들어도 Codex CLI가 같은 lock 규약을 사용하지 않으면 양쪽 갱신은 직렬화되지 않는다. 따라서 기존 보고서의 “공유 auth에 lock+rename이면 충분”한 해석을 폐기한다. read-only borrow와 소유권 있는 독립 refresh를 분리한다. CLI refresh token을 별도 파일로 단순 복제해도 같은 회전 체계를 공유하므로 독립 로그인과 같지 않다.

MoAI 소유 auth는 directory 0700/file 0600, process lock → 디스크 재읽기 → expiry 재판정 → bounded refresh → 임시 파일 fsync → atomic replace 순서를 적용한다. Windows는 ACL과 ReplaceFile semantics를 별도 시험한다. refresh 결과의 expiry/token은 로그로 내보내지 않으며, 실패한 응답으로 유효한 기존 파일을 덮어쓰지 않는다.

세션 ID, thread ID, request ID를 구분한다. launch마다 session ID, resume할 때 thread ID, HTTP request마다 request ID를 둔다. Codex header 이름·scope·originator는 `CompatibilityProfile`의 versioned fixture로 관리한다. 이전 조사에서 관측한 `Session_id`와 `session-id` 차이는 조정 후보이지, 옛 헤더가 지금 모든 요청에서 실패한다는 뜻이 아니다.

모델 metadata는 context 지원값, Claude 표시값, 실제 계정 probe 결과를 분리한다. 보수적인 검증값으로 기본 compaction을 설정하고 장문맥은 명시 opt-in으로만 열어 둔다. 이전 보고서의 `[1m]` 기본 사용 제안은 이 정책으로 대체한다. `MOAI_STATUSLINE_CONTEXT_SIZE`는 표시 보정일 뿐 Claude의 실제 compaction 제어를 증명하지 않는다.

reasoning signature는 `moaip1:` 버전 envelope 안에 provider/model family와 opaque 데이터를 구분해 담는다. 다른 provider로 전환하면 해당 opaque 상태는 제거하되 사용자·assistant text와 tool history는 보존한다. 같은 provider라도 모델 간 reasoning 재사용은 profile에서 허용한 조합만 통과시킨다. 손상된 envelope는 upstream에 보내지 않고 진단한다.

## 10. GLM과 혼합 세션

GLM은 Messages를 직접 받으므로 Codex 변환기를 거치지 않는다. body의 model만 registry에 따라 바꾸고, upstream credential을 Z.AI 키로 새로 설정한다. Claude local token·Codex bearer·account ID·Codex session header는 GLM으로 전달하지 않는다.

GLM stream은 가능한 한 bytes를 보존한다. unsupported beta/body field는 실제 거부 fixture가 있는 경우에만 provider profile에서 처리한다. 미리 모든 모르는 필드를 삭제하지 않는다. GLM effort와 Claude effort의 대응은 모델별 fixture로 따로 검증하며 문서 이름만으로 같은 의미라고 판단하지 않는다.

mixed 모드는 picker에 등록된 모델을 request마다 해석한다. global provider를 변경하는 공유 변수는 두지 않는다. GPT request와 GLM subagent request가 동시에 들어와도 각 request의 Route/credential/translator가 독립적이어야 한다. tool continuation의 provider 변경과 compact/title/agent slot 선택까지 함께 시험한다. 지원되지 않는 조합은 picker에서 숨기고 이유를 doctor에 표시한다.

## 11. 설정·기존 코드 통합

독립 `proxy.yaml` loader를 새로 만드는 대신 기존 `llm.yaml` 안에 `proxy` 항목을 추가한다. 배포 원본은 `internal/template/templates/.moai/config/sections/llm.yaml`이다. CLI flag가 세션에서 최우선이고, 다음은 프로젝트 proxy 설정, 마지막은 안전한 기본값이다. credential은 YAML에 넣지 않는다.

```yaml
llm:
  proxy:
    enabled: false
    provider: codex
    listen: 127.0.0.1:0
    credential_source: moai
    allow_mixed: false
    long_context: false
    models:
      primary: gpt-6-astra
      compact: gpt-6-astra
      background: gpt-6-astra
```

이 YAML의 필드명은 제안이다. 실제 wrapper 형태는 현재 loader/Save roundtrip과 일치하도록 구현 시 확정한다. enabled/provider 조합, loopback 주소, model availability를 validation한다. arbitrary backend URL은 프로젝트가 token 탈취 endpoint를 지정할 수 있으므로 기본 사용자-facing 설정에서 제외한다. custom endpoint는 사용자 범위의 명시 신뢰 설정에서만 허용하고 redirect와 DNS 주소 정책을 함께 적용한다.

| 기존 파일 | 필요한 수정 | 보존할 계약 |
|---|---|---|
| `internal/cli/cc.go` | provider flag 등록 | 일반 cc 동작 |
| `internal/cli/launcher.go` | 공통 LaunchSpec 생성 후 executor 선택 | worktree·profile·continue·permission 처리 |
| `internal/cli/launch_exec_posix.go` | 기본 경로 유지 | non-proxy exec semantics |
| 신규 `proxy_launch*.go` | spawn/wait·server 수명 | session PID·PTY·signal |
| `internal/config/types.go`, `defaults.go`, `validation.go` | ProxySettings 추가 | 기존 GLM URL 검증 유지 |
| `internal/config/envkeys.go` | proxy child-env 키 등록 | 인증 충돌 정리 |
| template `llm.yaml` | 선택 기능 기본 disabled | 기존 프로젝트 update 보존 |
| `internal/cli/doctor.go` | proxy 진단 합류 | 일반 doctor에 유료 호출 없음 |
| `internal/statusline` | profile 기반 표시 보정 | statusline에서 live network 금지 |

별도 proxy binary downloader, checksum registry, updater, CA installer는 만들지 않는다. Go toolchain 갱신과 vulnerability gate는 MoAI 전체 배포물 기준으로 다시 측정해야 한다. 상류의 15건 결과를 그대로 MoAI binary의 결과라고 표기하지 않는다.

## 12. 구현 단계와 이진 합격 기준

| 단계 | 범위 | 완료 산출물 |
|---|---|---|
| A | upstream 변환 corpus·LaunchSpec 계약 고정 | fixture manifest, 지원 기능 표 |
| B | internal server/router + fake providers | local auth·route·limits·shutdown 시험 |
| C | Codex request/stream + credentials | translation golden, refresh 동시성·읽기 전용 인증 시험 |
| D | cc executor·설정·doctor 통합 | PTY·session PID·종료·resume 회귀 |
| E | GLM passthrough·mixed route | secret 격리·provider 전환 E2E |
| F | 신규 설치·배포 검증 | 단일 binary 설치·runtime matrix·승인형 live canary |

| ID | 구체적 입력/동작 | 합격 기준 |
|---|---|---|
| ROUTE-01 | 3개 활성 route + unknown + duplicate model + 잘못된 suffix | 선택 provider 한 곳 또는 400; 잘못된 provider 요청 0 |
| AUTH-01 | inbound local/Anthropic/Codex/ZAI secret sentinel 조합 | 잘못된 outbound header·로그 sentinel 0 |
| AUTH-02 | Codex borrowed file 읽기·만료·외부 교체 | 파일 전후 SHA 동일; expired token 송신 금지 |
| AUTH-03 | 2 MoAI process 동시 만료 refresh, 중간 write 실패 | 단일 refresh 또는 회전 후 재읽기; 유효 JSON·권한 유지 |
| BODY-01 | chunked/gzip/deflate·limit+1·중복 key | 한도 초과 413, unsupported encoding 415, routing ambiguity 400 |
| STREAM-01 | text/tool interleaving·중복 terminal·terminal 없는 EOF | frame golden 일치; 정상 종료 event 최대 1; truncated는 오류 |
| STREAM-02 | multiline event 한도 초과·idle·disconnect | 자원 상한 준수; upstream cancel; goroutine 누수 0 |
| TOOL-01 | 긴 이름 충돌·call ID·tool error 왕복 | 이름/ID/실패 의미 보존 |
| LAUNCH-01 | fake Claude가 받은 env·PID·stdin을 기록 | 실제 Claude PID와 session PID 일치; 부모 env 불변 |
| LAUNCH-02 | PTY Ctrl-C/Ctrl-Z/fg, Claude exit 0/7, proxy fail | terminal 복원, exit 보존, listener·child 잔존 0 |
| LAUNCH-03 | 부모 SIGKILL·resume·같은 프로젝트 두 세션 | orphan 정책 충족; 서로 다른 token/port; 타 세션 종료 0 |
| SETTINGS-01 | 기존 settings/GLM 프로젝트에 활성화·해제 | 미관련 키와 사용자 auth SHA 불변 |
| GLM-01 | GPT↔GLM tool/thinking 전환·동시 requests | opaque 상태·credential 교차 송신 0 |
| INSTALL-01 | 깨끗한 fixture user dir에 MoAI 설치→login→cc | proxy 추가 binary 다운로드 0; built-in server 실행 |
| LIVE-01 | exact Claude version의 단문·Read·compact·title·subagent·MCP | run manifest에 model/route/status/exit 기록; 실패 기능 미지원 표시 |
| RELEASE-01 | Linux/macOS/Windows runtime와 Go vulnerability 검사 | 지원 target 모두 pass; 실행하지 못한 target은 미검증 표시 |

구현 후 예정 명령은 `go test ./internal/proxy/... -count=1`, 해당 package race, launcher의 지정 회귀 테스트, `go vet` 대상 package, CI의 MoAI 전체 binary vulnerability 검사다. 아직 실행 결과가 없으므로 위 표는 **예정 합격 기준**이다. 테스트 파일과 함수 이름은 구현 시 각 ID와 일대일로 연결한다. 무작정 전체 suite를 로컬에서 반복하지 않는다.

## 13. Gaps와 Residual-risk

**Gaps:** 내부 package와 새로운 executor는 아직 만들지 않았다. 이번 턴은 소스 재확인과 상세 설계다. 실제 OAuth login/refresh, 과금·quota, 272k 초과 장문맥, mixed 세션, PTY job control, parent death cleanup, Windows ACL/Job Object, image/PDF/structured-output, MoAI hooks/subagents/MCP 통합은 새 구현에서 검증해야 한다. 이전 safe-mode 실측은 hooks/skills 보존을 입증하지 않는다.

**Residual-risk:** ChatGPT subscription backend 및 Claude Code client 계약 변경, API 의미론 차이, 계정별 entitlement 차이, in-process translator 결함이 전체 MoAI 프로세스에 미치는 영향, 모델 명칭과 실제 비용 표시 불일치가 남는다. GPL은 이번 기술 진행 조건에서 제외했지만 직접 차용한 코드를 포함한 단일 바이너리의 외부 배포 라이선스는 여전히 별도 해결 대상이다.

## 14. 검토자가 확인할 결정과 출처

이 설계의 핵심 결정은 내부 package·단일 binary·in-process HTTP·Claude spawn/wait·Codex auth 읽기 전용 borrow·GLM native passthrough다. 다음 구현에서 먼저 검증할 부분은 translator보다 **LaunchSpec/session PID/terminal 계약**이며, 그 위에 provider와 stream을 결합한다.

- [ccmproxy 기준 tree](https://github.com/jclab-joseph/claude-code-model-proxy/tree/112588175eb2b3b693a5bd23d58d8c501e0ae406)
- [라우팅](https://github.com/jclab-joseph/claude-code-model-proxy/blob/112588175eb2b3b693a5bd23d58d8c501e0ae406/internal/router/router.go)
- [요청 변환](https://github.com/jclab-joseph/claude-code-model-proxy/blob/112588175eb2b3b693a5bd23d58d8c501e0ae406/internal/translate/claude_to_codex.go)
- [응답 변환](https://github.com/jclab-joseph/claude-code-model-proxy/blob/112588175eb2b3b693a5bd23d58d8c501e0ae406/internal/translate/codex_to_claude.go)
- [인증 저장](https://github.com/jclab-joseph/claude-code-model-proxy/blob/112588175eb2b3b693a5bd23d58d8c501e0ae406/internal/codex/store.go)
- [이전 실측 보고서](moai-proxy-feasibility-plan-20260910.md): 이전 결과와 공식 문서 조회의 원장. 이번 새 런타임 검증과 구분한다.
