# SPEC-MOAI-GATEWAY-001 — 내부 설계

현재 기준 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop`, 브랜치 `develop`,
HEAD `4056f69e1c20d942d4f9fc7363d3d79bffde899a`. 아래 package·파일 이름은 이 기준선에서 다시 판독한
현재 구현 경계와 M14의 구현 책임을 함께 나타낸다. `internal/gateway`와 `internal/codexbridge`는 존재한다.
`ed71054d3`와 `.claude/worktrees/moai-proxy-unified`는 0.6.0 역사 기록에만 적용된다.

명칭 대응: 설계 원문과 iter1 감사는 이 구성 요소를 "proxy"라 부른다. 이 문서는 "gateway"라
부른다(`spec.md` §0). 상류 프로젝트 `ccmproxy`는 원래 이름을 유지한다.


> t649 / 0.10.0의 현재 계약: 제공자별 선택·요청 경계, 시작·Default·네 슬롯·fallback·저장·재개 격리는
> 코어가 소유한다. 0.6.0~0.9.0의 교차 제공자 전환과 PICKER 이관·출시 대기 결정 중 이 범위는 대체되었다.
> 아래 과거 결정·프로브 기록은 당시 근거로 보존하며 현재 판정은 spec.md REQ-MG-019와 acceptance.md t649 보강을 따른다.

## 1. package 경계

| 파일 / 경계 | 책임 | 금지할 결합 |
|---|---|---|
| `internal/cli/gpt.go` | `moai gpt` 등록, 닫힌 동사 집합 라우팅, Factory 진입, launch 조립 | HTTP 변환 로직 직접 포함 |
| `internal/cli/gateway_launch.go` | cc/gpt/glm 공통 launch plan, 상속 env 정리, provider별 직접 모델 슬롯, gateway child 기동, 포트 인계, `MOAI_LAUNCH_PROVIDER` 설정 | 기존 worktree·profile·session 옵션 누락, `settings.local.json`에 base URL·credential 기록 |
| `internal/orchestration` | Factory/Todo/Tasks/Dispatch 공통 조정과 초기 provider 기록 | 과거 기록과 migration 근거의 소급 재작성 |
| `internal/gateway/server.go` | loopback listener, 세션 접근 통제, graceful 종료 | `internal/cli`·`internal/config` import |
| `internal/gateway/catalog.go` | 모델·계정 권한·capability snapshot | 접두사 추측 라우팅 |
| `internal/gateway/router.go` | 요청 모델 검증, adapter 선택 | 공유 `currentProvider` 변수 |
| `internal/gateway/anthropic.go` | Claude Messages, OAuth/API 인증 분기 (passthrough는 T09 게이트 뒤) | 다른 provider 인증 전달 |
| `internal/gateway/appserver.go` | 공식 Codex App Server와 Anthropic Messages ingress 사이의 GPT adapter | 직접 Responses·구독 token store·비공개 endpoint 호출 |
| `internal/codexbridge/` | App Server thread/turn/dynamic-tool JSON-RPC와 pending 원장 | Claude approval/hook 실행 또는 native Codex 도구 실행 |
| `internal/gateway/openai.go` | **[HISTORICAL/INACTIVE for GPT]** 과거 direct Responses 구현 | 현재 GPT transport나 구독/API fallback으로 선택 |
| `internal/gateway/glm.go` | Z.AI Messages 경로 | OpenAI endpoint로 임의 대체 |
| `internal/gateway/translate/` | 공통 Anthropic ingress 정규화와 history 오류 | 현재 GPT 경로에서 Responses wire format 생성 |
| `internal/gateway/auth/` | `CredentialRef` 인터페이스 정의, Anthropic·GLM 구체 구현 | Codex CLI 저장소 수정 |
| `internal/gateway/policy.go` | 전환·기능·본문 크기·외부 송신·path 정책 | 실패 시 무단 fallback |

패키지 이름을 `gateway`로 둔 이유: `internal/harness/router`가 이미 `package router`를 쓰고,
`internal/config/envkeys.go:479`가 이 자리를 "an LLM gateway"라 부르며, "proxy"는 설계 보고서
§5가 기각한 CA 설치·투명 가로채기를 연상시킨다.

plugin framework나 별도 daemon은 만들지 않는다. gateway는 세션 수명만큼 사는 child
프로세스 하나다.

## 2. 핵심 타입

```go
// LaunchPlan — 세 launcher가 공통으로 만들어 supervisor에 넘기는 계획.
type LaunchPlan struct {
    InitialModel    string          // launcher별로 다른 유일한 값
    InitialProvider ProviderID      // MOAI_LAUNCH_PROVIDER로 자식 env에 실린다
    Catalog         CatalogSnapshot // 세 launcher가 공유
    ChildEnv        []string        // 부모 env를 변형하지 않고 조립한 자식 env
    ChildSettings   string          // 자식 전용 settings overlay 임시 파일 경로
}

// ModelEntry — registry의 한 행. 표시 ID와 upstream을 잇는다.
type ModelEntry struct {
    RouteID      string       // Claude Code가 요청 본문 model에 실어 보내는 값
    Provider     ProviderID   // anthropic | openai | zai
    UpstreamID   string       // 해당 provider가 아는 정확한 모델 ID
    AuthMethod   AuthMethod   // oauth-passthrough | pkce | api-key | existing-glm
    Capabilities Capabilities // context 한도, 이미지/PDF, tool, streaming
}

// RequestContext — 요청 하나의 수명 동안만 존재한다. 전역 상태를 두지 않는다.
type RequestContext struct {
    RequestID     string
    Entry         ModelEntry
    CredentialRef CredentialRef // 직접 secret route 전용. App Server는 별도 ManagedSessionAuthority 사용
}
```

`RequestContext`가 요청 단위인 것이 tool 이름 역매핑을 요청 경계 안에 가두는 장치다
(`REQ-MG-013`). 전역 map을 두면 병렬 subagent 요청이 서로의 매핑을 덮어쓴다.

### 2.1 직접 credential과 App Server 관리 세션 권한

`CredentialRef`는 직접 secret을 적용하는 API·GLM·Anthropic route에만 사용한다.
App Server route는 `ManagedSessionAuthority`라는 별도 권한 경계로 profile·account·auth mode·generation·allowed model을
검증한다. 이 명칭은 설계 책임이며 가짜 `CredentialRef`나 빈 secret을 만들라는 뜻이 아니다.
App Server API-key 인증도 Codex가 관리하는 세션이면 이 경계를 사용한다. 직접 API route가 남는 경우의 secret seam은 유지한다.
관리 세션의 상태는 검증된 로그인/account 상태에서 갱신하며 logout·계정/프로필 교체 때 세대를 무효화한다.
선택 검증은 캐시된 로컬 권한만 확인하고 생성·refresh RPC를 일으키지 않는다. 실제 turn 직전에도 세대를 재검사한다.

```go
// CredentialRef resolves a provider credential at send time. The gateway
// never holds credential bytes in RequestContext, logs, or argv.
type CredentialRef interface {
    // Provider reports which provider this reference belongs to. Apply must
    // refuse a request addressed to any other provider.
    Provider() ProviderID

    // Generation returns an opaque, monotone version of the underlying
    // credential. A generation that changed between resolution and send
    // (for example after a logout) invalidates the send.
    Generation() (uint64, error)

    // Apply attaches the credential to an outbound request for Provider()
    // only. It returns ErrCredentialAbsent when no credential exists.
    Apply(req *http.Request) error

    // Redacted returns a log-safe description. It never contains the secret.
    Redacted() string
}
```

| 구체 타입 | 소관 | 이 SPEC에서의 상태 |
|---|---|---|
| GLM (기존 `internal/glmcred` 저장소 재사용) | 이 SPEC | 구현 |
| Anthropic API key | 이 SPEC | 구현 |
| Anthropic 구독 OAuth passthrough | 이 SPEC | T09 측정 양성일 때만 구현 (§2.2) |
| GPT App Server managed/API 세션 | t650~t652 | ManagedSessionAuthority 검증. credential bytes/Apply 없음 |
| 직접 API key route | 기존 secret seam | 해당 provider CredentialRef 유지 |

관리 세션 권한이 없거나 profile/account/generation/model이 맞지 않으면 REQ-MG-023의 401/404로 거절한다.
Codex managed 인증의 성공을 MoAI 소유 토큰 존재로 바꾸어 표현하지 않는다.

### 2.2 passthrough 게이트의 표현

`REQ-MG-016`의 게이트는 설정 키나 런타임 플래그로 표현하지 않는다. **코드와 catalog 항목의
부재**로 표현한다.

- T09 측정이 양성으로 기록되기 전에는 구독 OAuth passthrough용 `CredentialRef` 구체 타입을
  만들지 않는다.
- 같은 조건에서 catalog에는 `AuthMethod`가 `oauth-passthrough`인 항목이 없다.

설정 키로 두지 않는 이유: 사용자가 켤 수 있는 스위치는 측정 전 활성화를 막지 못한다. 코드가
없으면 켤 방법도 없다. `AC-MG-021` (c)는 catalog 열거와 upstream mock의 OAuth 헤더 수신 계수로
이 부재를 판정한다.

## 3. 프로세스 모델 (D2)

### 3.1 왜 goroutine이 아닌가

`internal/cli/launch_exec_posix.go:24-27`:

```go
func execOrSpawnClaude(claudeBin string, args, env []string) error {
	return syscall.Exec(claudeBin, args, withSessionPID(env, os.Getpid()))
}
```

같은 파일의 주석이 그 결과를 명시한다 — "the current shell process becomes claude,
so no defer() runs after this call". 이 프로세스의 PID가 곧 세션 PID이므로
`MOAI_SESSION_PID`를 여기서 각인한다. 프로세스가 치환되는 이상 in-process listener는
POSIX에서 살아남지 못한다.

### 3.2 채택한 순서

```
launcher 프로세스
  ├─ 1. profile / worktree / session 옵션 해석 (기존 경로 그대로)
  ├─ 2. settings.local.json의 GLM 정리 키 집합(14키) 정리 (§6.1)
  ├─ 3. catalog 조립(GLM 항목은 llm.glm.models 네 값) + 초기 모델 실행 가능성 점검
  ├─ 4. gateway child 기동:  moai <내부 서브커맨드> --handoff-fd=<w>
  │       └─ child: 127.0.0.1:0 bind → 포트를 handoff 채널에 한 줄 기록 → 서빙 시작
  ├─ 5. 부모가 포트 수신 (기한 초과 시 child 종료 + 명시 오류)
  ├─ 6. 자식 env 조립
  │     ├─ 6a. 상속 env에서 GLM 정리 키 집합(14키) 삭제. cc·gpt는 Z_AI_API_KEY도 삭제 (§6.1)
  │     ├─ 6b. 더하기: ANTHROPIC_BASE_URL=http://127.0.0.1:<port>,
  │     │        MOAI_LAUNCH_PROVIDER=<claude|gpt|glm>, 세션 접근 토큰, effort 키,
  │     │        glm만 Z_AI_API_KEY=<GLM credential 저장소 값>,
  │     │        glm만 ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL=<llm.glm.models high·medium·low·fable>
  │     └─ tmux 세션 안이어도 tmux 세션 env에는 쓰지도 지우지도 않는다 (§6.6)
  ├─ 7. 자식 전용 settings overlay 임시 파일 작성 (비밀 없음)
  └─ 8. syscall.Exec(claudeBin, args, env)   ← 여기서 launcher 프로세스가 사라진다
```

8단계 이후 부모는 존재하지 않는다. 따라서 정리 책임은 전적으로 gateway child에 있다.

**6a가 6b보다 먼저인 이유 (0.7.0, iter5 G5-B1).** 오늘 launcher는 자식 env의 기반으로 `os.Environ()`을 그대로 넘긴다
(`internal/cli/launcher.go:837`, GLM 분기 `:833-834`). `buildEnvForLaunch`(`:1180-1200`)는 effort 값이 있을 때
`CLAUDE_CODE_EFFORT_LEVEL` 한 키만 바꾸고 나머지 상속 항목은 그대로 두며, GLM 분기의 `buildEnvForGLMLaunch`(`:1236`)도
effort·reasoning 키만 바꾼다. 그 기반에는 tmux 세션 env에서 새 pane으로 들어온 GLM 키가 있을 수 있고, 결정 12에 따라
gateway launch는 오늘 `applyCCMode`가 하던 tmux GLM 키 정리(`launcher.go:224`)를 하지 않는다. 정리를 6b 뒤에 두면
launcher가 방금 더한 loopback `ANTHROPIC_BASE_URL`과 세션 접근 토큰까지 지워지므로, 정리는 상속 기반에만 먼저
적용한다(`spec.md` `REQ-MG-021`). `moai glm`에서는 6a가 지운 네 모델 슬롯 키를 6b가 설정 값으로 다시 더하므로, 순서가 바뀌면
슬롯 키도 사라진다(§6.7).

**M0 측정으로 확정한 운반 키 (0.9.0).** `X-MoAI-Session-Token`을 별도 헤더로 사용한다.
상속된 `ANTHROPIC_AUTH_TOKEN`은 6a에서 제거하며 6b에서 다시 넣지 않는다. 로컬 인증을 마친 gateway는
세션 헤더를 upstream에 전달하지 않는다. Claude가 관리하는 OAuth Bearer·허용 beta는 고정 Anthropic endpoint에만
전달한다. 같은 세션 헤더의 중복·외부 값 충돌은 거절하며 임의 병합하지 않는다. 이 선택을 CG·TEAMMATE에 인계한다.

19시 이후 Opus 5·Sonnet 5 정상 응답과 Opus 5의 직접 token POST 200→같은 본문의 새 Bearer 재전송 200은
`research.md` §19의 M0 양성 근거다. 이것은 passthrough 구현의 선행 게이트 해소이며 제품 통합·도구·재개 PASS가 아니다.
기존 INCONCLUSIVE 보고서는 역사적 증거로 보존한다. credential의 갱신·저장 소유권은 Claude에 남는다.

### 3.3 포트 인계 채널

부모→자식 방향은 인수로 충분하지만, 자식→부모 방향(bind된 포트)은 exec 이전에
받아야 하므로 채널이 필요하다. 후보는 상속 파이프 fd(부모가 `os.Pipe()`로 만들고
자식에 `ExtraFiles`로 넘김)와 임시 파일 폴링 두 가지다. 파이프를 기본으로 삼는다 —
폴링 간격이라는 임의 상수가 없고 자식이 죽으면 read가 EOF로 즉시 끝나기 때문이다.
Windows는 `ExtraFiles` 의미가 다르므로 §3.5에서 따로 다룬다.

### 3.4 고아 수거

gateway child는 부모(launcher)가 exec로 `claude`가 된 뒤에도 살아 있어야 하고, 그
`claude`(lead 세션)가 죽으면 따라 죽어야 한다
(`REQ-MG-008`). 감시 대상은 launcher의 PID이며, exec가 PID를
보존하므로 **exec 전에 읽은 PID가 exec 후의 `claude` PID와 같다**. 이 성질이 감시를
단순하게 만든다.

- POSIX: gateway child가 세션 PID를 주기적으로 확인한다. PID가 사라지면 listener를 닫고 임시 settings 파일을 지운 뒤 종료한다.
- 이중 안전망: gateway child는 자기 수명 상한(세션 최대 길이)도 갖는다. 감시가 어떤
  이유로 실패해도 무한히 남지 않는다.

**teammate와 수명 (0.6.0).** gateway launch는 in-process teammate만 허용한다(§6.6). Claude Code 문서는 in-process
teammate의 백그라운드 작업이 lead 프로세스보다 오래 살 수 없다고 적지만(`research.md` §16), in-process teammate가 lead의
gateway 주소를 물려받는지와 lead 종료 뒤 남는 teammate가 없는지는 측정하지 않았다(`plan.md` M7). 그래서 이 설계는 lead
세션 종료를 gateway child의 종료 조건으로 삼고, 그 뒤 옛 주소로 오는 요청이 연결 오류로 드러나게 실패하는 것을 계약으로
둔다. gateway child의 종료 경로는 어떤 upstream에도 요청을 보내지 않는다. 수명 상한은 이 조건과 무관하게 적용된다. tmux
pane teammate가 lead보다 오래 사는 경우의 수명 계약은 형제 `SPEC-MOAI-GATEWAY-TEAMMATE-001`(제안) 소관이다.

**PID 재사용 방어.** PID가 살아 있다는 것은 그 PID가 **우리** 세션이라는 뜻이 아니다. 감시는
PID만 보지 않고 프로세스 신원 지문을 함께 본다. 저장소에는 이미 선례가 있다 —
`internal/homestate`의 `CurrentProcessFingerprint`와 `ProbeProcessIdentity`는 빌드 태그가 없는
파일에 있어 두 플랫폼에서 모두 컴파일되며, Windows launch 경로가 자식 신원 확인에 이미 쓰고
있다. gateway child는 기동 시 부모 지문을 받아 두고, 감시 때 같은 PID의 지문이 달라졌으면 세션이
끝난 것으로 본다.

**빌드 태그.** POSIX 전용 호출(signal 0 전송 등)을 쓰는 감시 구현은 기존
`launch_exec_posix.go`(`//go:build !windows`)와 `launch_exec_windows.go`(`//go:build windows`)처럼
플랫폼별 파일로 나눈다. 플랫폼 공통 부분은 태그 없는 파일에 둔다.

### 3.5 Windows 계약 (`REQ-MG-009`)

`execOrSpawnClaude`의 Windows 구현은 이미 spawn-and-wait이다. `child.Start()` 뒤
`homestate.ProbeProcessIdentity`로 자식 신원을 확인하고 profile lease를 넘긴 다음
`child.Wait()`로 종료 코드를 전파한다. 부모가 살아 있으므로 Windows에서는 gateway를
**부모가 직접 감독**할 수 있다. 두 플랫폼이 같은 코드일 필요는 없고, 같은 계약
(세션 종료 시 gateway 종료, 포트 해제, 임시 파일 삭제)만 지키면 된다.

검증은 로컬에서 하지 않는다. 판정 지점은 **release PR 게이트**다(운영자 결정, 0.3.1).

- **시험의 자리.** supervisor Windows 시험은 launcher/supervisor 코드를 소유한 패키지(`internal/gateway`
  또는 `internal/cli`)의 평범한 Go 시험으로 둔다. `test/integration/harness/` 아래에 두지 않고
  `integration` 빌드 태그도 달지 않는다. `.github/workflows/release-pr-multi-os.yml:203`은
  `go test -json -race -timeout 25m ./...`를 `-tags=integration` 없이 실행하므로, 태그를 단 시험은 그
  레그에서 조용히 빠진다.
- **판정 지점.** 같은 워크플로의 windows-latest 레그다. 매트릭스는 `:91`, 트리거는 `main` 대상
  PR(`:14`) 가운데 head가 `release/*`인 것(`:35`)이며, 각 OS 레그가 release 게이트를 막는다(`:93`
  주석). 이 배치는 저장소의 의도된 정책이다 — `ci.yml:110-112`는 cross-platform 런타임 커버리지를
  release 시점으로 옮겼다고 적는다.
- **평소 CI가 하지 않는 일.** 카드·develop CI(`ci.yml`)는 이 시험을 Windows에서 실행하지 않는다.
  `test` 잡 매트릭스는 `[ubuntu-latest]`(`:125`)이고, 3-OS인 `test-integration`은
  `./test/integration/harness/...`만 실행한다(`:402`). 그래서 Windows 회귀는 release PR 시점에야
  드러난다. 잔여 위험이다.
- **공허 방지.** 시험은 Windows에서 `t.Skip` 하지 않는다. 판정은 업로드된 이벤트 스트림 아티팩트
  `test-stream-release-verify-${{ matrix.os }}`(업로드 단계 `:213-220`, 파일은 `test-stream.json.gz`)를
  읽어, windows-latest에서 이름을 정한 supervisor 시험 각각이 `"Action":"pass"`로 끝났는지 단언한다.
  업로드 단계가 `if-no-files-found: warn`(`:220`)이므로 아티팩트가 없어도 레그는 실패하지 않는다 —
  아티팩트 부재나 이름을 정한 시험의 부재는 PASS가 아니다.
- **남는 간극.** 대화형 TTY와 job control 동작은 CI runner에서 재현되지 않는다. Windows에 대해
  로컬에서 잰 것은 없다.
- **판정 주체·시점·기록.** 카드 run/sync가 닫히는 시점에는 이 판정이 아직 없으므로, 카드 완료 보고는
  `AC-MG-006`의 Windows 절반을 PASS가 아니라 Gap(판정 대기)으로 기록한다. release 배치의 리드(release PR을
  여는 세션)가 windows-latest 레그가 끝난 뒤 아티팩트 보존 기간(`retention-days: 7`, `:219`) 안에
  아티팩트를 읽고, 이름을 정한 시험별 종료 동작과 실행 URL을
  `.moai/reports/SPEC-MOAI-GATEWAY-001/windows-release-verdict.md`에 기록한다. 기간 안에 읽지 못했으면
  워크플로의 `workflow_dispatch` 트리거(`:16`)로 다시 실행해 읽는다. 만료로 사라진 아티팩트는 PASS의
  근거가 되지 않는다.

## 4. ingress 처리 흐름

```
POST /v1/messages
  ├─ 본문 크기·JSON 형태·model 키 중복 검사        → 위반 시 4xx (외부 송신 없음)
  ├─ 선택 시점 검증 요청인가 (A-VAL-2.1.268, §4.1)  → 예: 로컬 응답으로 끝 (404/401/200, upstream 송신 없음)
  ├─ registry exact match (접두사 추측 없음)        → 미등록 시 4xx
  ├─ route별 인증 권한 확인 (§2.1, generation)     → 불일치 시 4xx
  ├─ history 재구성 (다른 provider 요청 사전 거절,
  │   미완료 tool pair 있으면 거절)
  ├─ adapter 호출
  │    ├─ anthropic: 헤더 허용 목록만 전달 (passthrough는 T09 게이트)
  │    ├─ openai:    Messages → App Server thread/turn
  │    └─ zai:       Messages 그대로, beta 헤더 제거
  └─ 응답
       ├─ 비스트리밍: 형식 역변환 후 반환
       └─ SSE: message_start → blocks → message_delta → message_stop
            중간 실패 시 성공 terminal event 합성 금지, 송신 후 재시도 금지
```

**판정 순서 (0.6.0, iter4 G4-A6).** 검증 요청 판정은 registry·credential 판정보다 **먼저** 한다. §4.1의 응답 표는
검증 요청에 HTTP 404 `not_found_error`와 HTTP 401 `authentication_error`라는 형식을 요구하는데, 일반 4xx 거절이 먼저 오면 그
형식을 보장할 수 없다. 검증 요청 판정 자체는 본문 형태만 보므로 registry·credential 조회에 의존하지 않으며, 인식된 요청은
그 안에서 §4.1 응답 표의 순서(catalog → credential)대로 답한다. 본문 크기·JSON 형태·`model` 키 중복 검사는 검증 요청에도
먼저 적용된다.

경로별 정책(`REQ-MG-014`): `/v1/messages`, `/v1/messages/count_tokens`, `/v1/models`은
명시 처리한다. 나머지는 기본 거절이며, 개별 path를 열 때만 정책 표에 추가한다.
모든 unknown path를 Anthropic으로 흘려보내는 catch-all은 두지 않는다 — 그것이
"모르는 요청을 외부로 보내는" 경로다.

### 4.1 선택 시점 검증 요청 (`REQ-MG-023`, `plan.md` 결정 9)

**관측.** Claude Code 2.1.267은 `/model <id>`로 모델을 고를 때 turn과 별개로 검증 요청 하나를 보낸다
(`research.md` §15). 경로는 `POST /v1/messages?beta=true`이고, mock 기록으로는 `stream`이 거짓, `max_tokens: 1`, 도구 0개,
user 메시지 `"Hi"` 하나다. Claude 모델로 되돌아갈 때도 보낸다. 401 `authentication_error`나 404 `not_found_error`를
받으면 오류 줄만 표시하고 전환·저장·기록 추가를 하지 않는다. 성공하면 확인 대화상자를 띄우며, 확인 뒤에는 추가
요청이 없다. picker `s` 경로는 이 요청을 보내지 않는다. 같은 두 프로브에서 turn 요청과 제목 생성 요청은 모두
`stream`이 참, `max_tokens` 32000이었고, 첫 turn과 제목 생성 요청은 메시지가 하나였다.

**관측의 한계 (0.6.0, iter4 G4-B2).** 위 관측은 요청 형태를 고정할 만큼 강하지 않다. 다음 셋을 Gap으로 기록한다
(`research.md` §15.6).

- **억제 플래그.** 두 프로브는 `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`, `DISABLE_TELEMETRY=1`,
  `DISABLE_ERROR_REPORTING=1`, `DISABLE_AUTOUPDATER=1`을 켠 채 실행했다(프로브 1 README "What was run", 프로브 2 README는
  "Same isolation as probe 1"). 이 SPEC의 launch 정리는 `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`를 지우므로 제품 세션은
  측정하지 않은 조건에서 돈다. 그 조건에서만 나타나는 다른 1토큰 비스트리밍 요청은 목록에 없다.
- **`stream`은 JSON 값으로 관측되지 않았다.** 프로브 1 mock은 `stream`을 `bool(req.get("stream"))`로, 도구를
  `len(req.get("tools") or [])`로 기록했다(`.moai/state/gwprobe/mock.py:89`, `:93`). `stream` 키가 없는 요청과 `false`인
  요청이 같은 기록을 남기고, 도구 키가 없는 요청과 빈 배열인 요청도 같은 기록을 남긴다.
- **원본 본문이 없다.** `requests.jsonl`에는 mock이 파생한 요약만 있고 요청 본문과 질의 문자열의 원문은 없다.

그래서 인식 기준의 **근거**는 이 관측이 아니라 `plan.md` M1 진입 게이트의 원본 캡처다. 캡처는 실제 Claude Code TUI를
loopback 캡처 mock에 붙여, 네 억제 플래그 없이 `/model <id>` 검증 요청, 첫 turn, 제목 생성, 이후 turn의 요청 본문을 가공
없이 기록한다. 픽스처는 그 원본에서만 고정한다. 캡처한 검증 요청의 형태가 아래 기준과 다르면 구현을 멈추고, 기준을 이
SPEC에서 먼저 고친 뒤 코드를 쓴다.

**가정 A-VAL-2.1.268 (0.8.0에서 캡처로 보정한 가정).** 다음 네 조건을 **모두** 만족하는 `/v1/messages` 요청만 선택 시점 검증 요청으로 본다.
하나라도 어긋나면 검증 요청이 아니며 통상 경로를 탄다.

| 조건 | 값 | 어긋나는 예 (모두 인식하지 않음) |
|---|---|---|
| `stream` | 키가 없거나 값이 JSON `false` | `true`, 문자열 `"false"`, `null` |
| `max_tokens` | 정수 `1` | `2` 이상, 키 없음 |
| `messages` | 길이 1이고 그 항목의 `role`이 `"user"` | 길이 0 또는 2 이상, `role`이 `"user"`가 아님 |
| `tools` | 키가 없거나 빈 배열 | 도구 정의가 하나 이상인 배열 |

첫 turn과 제목 생성 요청도 메시지가 하나이므로 메시지 수만으로는 가를 수 없어, 네 조건을 함께 요구한다. 사용자 text 값
(`"Hi"`)은 기준에 넣지 않았다. 경로 정책은 `?beta=true` 질의 문자열이 붙은 형태를 `/v1/messages`로 받아야 한다.

**`stream` 키 부재를 허용하는 근거 (0.8.0).** 억제 플래그 없이 캡처한 Claude Code 2.1.268의 선택 검증 요청
`request-003.json`에는 이 키가 없고, 나머지 세 조건은 모두 맞았다(`research.md` §17). 따라서 키 부재 또는 JSON
불리언 `false`를 허용한다. 두 경우는 별도 양성 대조군으로 시험하며, 참거짓 변환으로 `null`·문자열·`true`를 받아서는 안 된다.
관측한 첫 turn·후속 turn·제목 요청 네 건은 `stream: true`, `max_tokens: 32000`으로 모두 제외된다.
이는 모든 정상 작업 요청과 구분된다는 증명이 아니다. 같은 네 조건을 갖는 다른 작업 요청이 관측되면 기존 M1 중단 조건을
유지한다. 명시 `false`는 이번 원문에서 관측한 값이 아니라 기존 계약을 보존하는 허용값이다.

**`tools`를 조건에 넣는 이유.** 도구 정의를 실은 요청은 모델이 도구를 호출할 수 있는 작업 요청이므로 인식에서 뺀다. 이
판단은 요청의 의미에 근거한 것이며, 관측된 turn 요청이 도구를 실었는지에 근거한 것이 아니다. 도구 키가 없는 경우와 빈
배열인 경우는 mock 기록으로 가를 수 없었으므로 둘 다 받아들이고, 원본 캡처가 실제 형태를 기록한다.

**응답.** 위에서부터 순서대로 판정한다. 어느 행도 upstream에 요청을 보내지 않는다.

| 조건 | 응답 |
|---|---|
| 모델 ID가 세션 catalog에 없음 | HTTP 404, 오류 `type` `not_found_error` |
| 직접 route credential 부재 또는 App Server ManagedSessionAuthority 검증 실패 | HTTP 401, 오류 `type` `authentication_error` |
| 그 밖 | HTTP 200, Anthropic Messages 형식의 최소 비스트리밍 응답(형식 세부는 run 단계가 픽스처로 정한다) |

**App Server 검증.** profile·account·auth mode·generation·allowed model의 현재 로컬 snapshot을 검증한다.
토큰 파일을 읽거나 Apply를 호출하지 않는다. 불일치는 401, 세션 catalog 밖 ID는 404이며 생성·refresh·유료 호출은 0이다.
이는 서버 entitlement 확인이 아니므로 실제 turn 실패는 그대로 전달한다.

**직접 secret route의 credential 존재 확인 수단.** §2.1 인터페이스에는 존재만 묻는 메서드가 없다. 송신하지 않는 요청에 `Apply`를
적용해 `ErrCredentialAbsent`를 보는지, `Generation()`의 오류로 보는지는 run 단계가 고정된 인터페이스 안에서
정한다. 어느 쪽이든 그 요청은 외부로 나가지 않는다. Anthropic passthrough(T09 게이트)가 출시되는 경우의
"credential 존재"는 그 구체 타입이 해석하는 값을 따르며, 게이트가 닫힌 동안에는 해당하지 않는다.

**경계.**

- 전환 시점에 확인되는 것은 credential **존재**다. 만료·폐기는 첫 실제 turn의 upstream 응답에서 드러나며, 그 오류는
  `REQ-MG-022`에 따라 fallback 없이 클라이언트에 전달된다.
- picker `s` 경로는 검증 요청을 보내지 않는다. 그 경로로 credential 없는 모델로 전환하면 첫 turn이 `REQ-MG-023`의
  외부 송신 전 거절로 끝나야 하며, 이것이 그 경로의 유일한 경보다.
- 검증 요청이 모든 요청에 같은 클라이언트 bearer를 싣는다는 관측(F9)은 세션 접근 토큰 판정과 무관하게 이 응답을
  정하지 않는다. 운반 키 판정은 M0 결정을 따른다.

**잔여 위험.**

- **형태 변경.** 클라이언트가 검증 요청의 형태를 바꾸면 인식이 빗나가 그 요청은 통상 경로로 대상 provider에
  전달된다. 사용자가 고른 provider로 가므로 우회는 아니지만 최소한의 유료 호출이 생기고, credential 부재 시 오류
  형식이 위 표와 달라질 수 있다. Claude Code 버전을 올릴 때 M1 진입 게이트의 원본 캡처를 다시 한다.
- **오인식.** 다른 클라이언트나 미래 버전이 같은 형태로 실제 요청을 보내면 로컬 성공 응답을 받는다. 네 조건을 함께
  요구해 폭을 줄였지만 없앨 수는 없다.
- **통상 turn의 오류 형식.** 위 표의 오류 형식은 검증 요청에 대한 의무다. 통상 turn의 미등록 모델·credential 부재
  오류가 같은 형식을 쓰는지는 이 판에서 정하지 않았다.

### 4.2 `messages` 안의 `role: "system"` 항목 (`REQ-MG-015`)

프로브 2의 GPT turn 요청은 메시지 8개 가운데 user 3개, system 3개, assistant 2개였고, system 항목에는 클라이언트가
주입한 알림(예: `<total_tokens>…`)이 실려 있었다(`research.md` §15). provider를 바꾸는 재구성은 이 항목을 입력으로
받는다.

M5 문서 결정(2026-09-11, 0.8.0 유지): 최상위 `system`의 text 블록은 순서·각 문자열을 보존하여 `\n\n`으로
연결한 Responses `instructions`로 옮긴다. `messages` 내부의 `role: "system"` text 항목은 대화 내 같은 위치의 Responses
system 메시지로 보존하며 최상위 instructions로 끌어올리지 않는다. 지원하지 않는 content 형태는 조용히 버리지 않고 명시
오류로 거절한다. 이는 결정적 변환 규칙의 확정이며 실제 provider가 이 입력을 수용했다는 판정은 아니다. 주입 항목 수의
변동이 프로브 메시지 수 감소 원인인지는 여전히 미측정이다.

### 4.3 App Server GPT adapter — 0.11.0 채택

이 절은 0.9.0~0.10.0 직접 Responses·opaque/receipt 복원 설계를 현재 GPT 구독 경로에서 대체한다.
이전 인증·carrier 실험의 관측을 취소하지 않으며, 그 경로의 성공을 App Server 성공으로 승계하지 않는다.

## 최소 구성과 재사용

기존 loopback 인증·제공자별 picker·MoAI launcher·대화 소유권 경계를 유지하고 GPT 구독 adapter 뒤만 App Server로 바꾼다.
직접 구독 backend 요청과 MoAI의 구독 토큰 읽기·refresh는 이 경로에서 제거한다. Codex managed auth가 소유한다.
로그인 화면의 세 선택은 제안이다: ChatGPT 브라우저 managed / ChatGPT device-code managed / OpenAI API 키.
앞의 둘은 같은 구독 과금 경로의 로그인 방식 차이이며 세 제공자를 뜻하지 않는다. 기존 제품의 관측된 메뉴라고 표시하지 않는다.
API 키 모드는 별도 명시 선택으로 지원하고 Codex의 apiKey 모드를 우선 검토한다. 구독 실패 시 API 과금으로 자동 전환하지 않는다.
[공식 인증 모드](https://learn.chatgpt.com/docs/app-server#authentication-modes)

`internal/cli/mcp_codex.go:425` 이후에 NDJSON stdio 연결, spawn/close, handshake와 재사용 session handle이 이미 있다.
`codexConn`의 send/recv/close, 직렬 송신 lock, bounded scanner, PID 소유권 및 cancel ID 관리가 재사용 후보다.
`codex_session_test.go`, `codex_live_protocol_probe_test.go`도 존재한다. 보고된 `codex_session.go`는 이 트리에서 없었다.
review 전용 완료 판독을 GPT adapter에 가져오지 않는다. 중립 RPC 계층을 추출할 때 기존 review 소비자의 회귀를 실행한다.
현재 단일 recv 흐름에 서버 요청·알림·응답의 ID별 분배와 backpressure·취소를 더해야 하며 별도 중복 프로세스 프레임워크는 만들지 않는다.

## 도구 목록과 ToolSearch

**기존 A 전제의 반증 (2026-09-12).** Claude 2.1.269 local mock 기록
`probe/claude-tool-surface-result.json`을 재판독했다. 요청1은 deferred placeholder만 포함하고 fixture echo schema는
없었다. ToolSearch 뒤 요청2에 `mcp__fixture__echo` 전체 schema와 같은 이름의 tool_reference가 함께 추가되었다.
따라서 기본 ToolSearch 상태의 첫 요청 전체 schema 선등록은 이 픽스처에서 성립하지 않는다.

**승인된 hybrid.** 운영자는 “ToolSearch 유지·발견한 도구만 공통 통로 사용”을 선택했다.
최초 완전 정의 도구는 native schema로 thread 시작 때 등록하고 고정 dispatcher 하나도 함께 등록한다.
ToolSearch 뒤 처음 발견한 도구는 dispatcher의 `{name, arguments}`를 원래 Claude tool_use로 역매핑한다.
초기 native 도구는 dispatcher로 우회할 수 없고 dispatcher 자체를 다시 호출하는 재귀도 차단한다.

도구별 원래 이름·route·schema digest·발견 epoch를 소유 대화에 결박하고 pending RPC에는 그 snapshot을 고정한다.
한 요청 안의 중복 참조·중복 정의는 거절한다. 이후 요청에서 이미 알려진 도구가 동일 schema로 다시 발견되는
정상 ToolSearch는 멱등 처리한다. 발견 epoch·schema digest·기존 route를 바꾸지 않고 native 도구를 dispatcher로 옮기지 않는다.
tool_reference나 모델에게 보이는 설명만으로 실행 권한을 주지 않는다. 인증된 Claude 요청의 typed tool 정의,
정상 발견 결과, 현재 허용 상태를 함께 검증한다. 후발 정의가 기존 이름을 덮어쓰거나 pending schema를 바꾸면
명시 거절한다. canonical data는 중복 key를 거절하고 object key 순서만 정규화하며 배열 순서와 schema 의미를 보존한다.
validator의 dialect·$ref 지원 범위를 선언하고 미지원 구조는 실행 전에 거절한다. required-only 검사는 충분하지 않다.
등록을 위해 thread/start·fork·resume를 다시 호출하거나 원 history를 재생하지 않는다.

ToolSearch 비활성 선등록은 채택하지 않았다. 비교 기록의 초기 tools 구조 UTF-8은 10998→17295 bytes(+6297),
input_schema는 10534→16476 bytes(+5942)였다. 이는 토큰 수가 아니다.
Sol 단일 dispatcher 프로브의 ToolSearch→fixture_echo→MOAI_DISPATCHER_OK 성공은 구조 가능성 근거이며
실제 MoAI hybrid·다중 도구·잘못된 인수·native 실행 차단의 전체 성공을 뜻하지 않는다.
후발 도구의 모델 측 표현이 도구별 native schema와 같은 품질이라고 주장하지 않는다.

루트가 별도 실행 담당에게 받은 dynamic tool 왕복 PASS는 이 문서 작성자가 직접 실행한 증거가 아니며,
그 실행에 hooks/MCP가 상속되었다는 보고도 있다. 따라서 단일 도구 왕복과 실행권 격리 통과를 분리해야 한다.
`environments: []`, hooks 비활성, MCP 비활성은 다음 격리 프로브의 후보이며 지원 확인 전 확정 설정으로 쓰지 않는다.
어느 안이든 Codex shell/file/MCP/agent 자체 실행을 차단하고 Claude 실행만 허용하는 native 도구 inventory·실행 0 증거가 필요하다.
readonly sandbox나 approvalPolicy never는 그 증거를 대신하지 않는다. thread/shellCommand도 호출하지 않는다.

## thread/turn/item 대응

| Claude/MoAI 사건 | 대응 제안 | 실패 경계 |
|---|---|---|
| 첫 사용자 turn | 소유권을 확인한 대화에 thread/start 후 turn/start | 모델·auth·tool 목록 검증 전 생성 요청 금지 |
| 텍스트 응답 | App Server agentMessage 이벤트를 Messages SSE로 변환 | turn ack를 응답 완료로 표시 금지 |
| tool 호출 | item/tool/call의 ID를 보관하고 Claude tool_use로 응답 | HTTP 종료와 App Server turn 종료를 혼동하지 않음 |
| tool_result 다음 HTTP 요청 | 소유 대화·call·결과를 검증해 보관된 RPC에 응답 | 중복·foreign·변조 결과를 다른 turn에 주입 금지 |
| tool 이후 최종 응답 | 이어지는 item/turn 완료를 같은 사용자 turn에 연결 | tool 실행 재시도·중복 비용을 조용히 발생시키지 않음 |
| 정상 새 사용자 메시지 | 이미 반영한 prefix 이후의 새 입력만 turn/start | 전체 Claude history를 매번 다시 넣지 않음 |
| 프로세스 재시작 뒤 재개 | 저장된 소유 thread ID로 thread/resume | 불안정 history 주입이나 raw reasoning 재생 금지 |
| 미완료 tool 대기 중 crash | pending 실행 여부를 판정해 복구 불가 시 명시 실패 | 사라진 RPC ID를 새 프로세스에 답하거나 도구 재실행 금지 |
| `/model` GPT 변경 | idle 경계에서 turn model override, 지원 목록 확인 | active tool 도중 변경 및 미검증 family 호환을 숨기지 않음 |
| compaction | 정상 요약 turn 후 인증된 PostCompact와 반환 요약 대조로 public history rebase | 다음 HTTP 추측 금지; 추가 명시 compact 호출 0회 |
| Claude Agent / 명시 session fork | 일반 자식은 독립 thread, launcher의 명시 분기는 완료 prefix 원장과 lastTurnId | 일반 자식과 conversation fork 혼동 및 부모 추측 금지 |

thread/turn/call ID, 반영 prefix digest, 결과 수신·소비 상태를 대화별로 원자 저장한다. 이는 RPC 재생 권한을 주는 토큰이 아니다.
reasoning·암호문은 Codex가 관리하고 MoAI는 raw 내부 history 조회·복원을 정상 경로로 요구하지 않는다.
### AS4 압축: 정상 요약 turn과 완료 후 이력 재설정

Claude의 압축 HTTP를 사전에 분류할 필요가 없다. 일반 요청과 같은 공개 이력 비교를 거쳐 새 요약 지시만
App Server의 정상 turn으로 보낸다. PreCompact는 다음 HTTP의 종류를 결정하지 않는다. 실제 반환한 모델 요약을
그대로 Claude에 전달하며 MoAI가 `thread/compact/start`를 추가 호출하지 않는다. Codex 자체 자동 압축은
공식 runtime 소유로 둔다. 이는 Claude UI·App Server·사실 보존 목표를 유지하면서 이전 예약 방식의 판별 의존을 없앤다.

Claude 2.1.269 합성 probe `../../reports/t649/appserver-redesign/probe/compact-history-shape-verdict.md`에서
이전 user·assistant의 공개 text 유지, system text의 배열/문자열 표현 변경, 마지막 요약 지시 추가를 관측했다.
반환 요약과 PostCompact compact_summary의 SHA-256이 일치했고 후속 history에는 요약 wrapper와 별도 새 질의
블록이 있었다. 이것은 설치본 구조 근거이며 실제 모델 요약 품질이나 모든 버전 지원의 증거가 아니다.

1. 각 정상 응답의 공개 text digest와 요청 digest, 계정 세대·family·agent·thread·epoch·완료 turn ID를 영속 저장한다.
   동일 요청의 재시도는 저장된 응답을 반환한다. 실행 완료 여부가 불명확하면 재생성하지 않는다.
   공개 이력 비교에서 cache_control만 제외하고 같은 text의 문자열/배열 표현만 정규화하며 역할·도구 ID·순서는 유지한다.
   이미 반영한 prefix는 다시 보내지 않고 새 요약 지시를 한 번 보낸다. prefix 변경을 추정하여 무시하지 않는다.
2. 세션 전용 IPC 권한으로 인증한 PostCompact가 compact_summary를 전달한다. 검증자는 scope/epoch 및 완료 응답의
   exact text digest를 대조한다. prompt_id는 hook 중복 식별 보조 값일 뿐 HTTP 결속 증명이나 인증 수단이 아니다.
   후보 응답이 없거나 둘 이상이면 rebase를 거절한다. PreCompact/SessionStart만으로 완료 처리하지 않는다.
3. 다음 HTTP는 설치본별 검증된 continuation wrapper 구조와 저장된 exact summary를 대조해 새 공개 이력 기준을
   확정한다. 단순 substring 일치로 일반 입력을 버리지 않는다. wrapper·native command/hook 메타데이터와 새 사용자
   블록을 구조적으로 분리하고 분리가 모호하면 송신 전에 거절한다. 요약 wrapper는 Codex에 재주입하지 않고 새 입력과
   검증된 현재 system 변경만 보낸다. reasoning은 Codex 소유 이력에 남긴다.
4. PostCompact 중복은 멱등 처리한다. 통지가 늦어 변경된 history 요청이 먼저 오면 대기 또는 재시도 가능한 명시 오류로
   처리하고 외부 생성을 하지 않는다. 통지가 유실되면 조용히 전체 이력을 재전송하지 않는다. 완료 응답·통지·rebase
   상태를 원자 저장하여 새 프로세스에서도 같은 결과를 사용한다. 오래된 epoch·foreign 통지·요약 변조를 거절한다.

수동 print/resume, 대화형 UI, 자동·자식 압축을 각각 검증한다. 실제 요약 표시, 압축 뒤 사실·도구 결과 회상,
새 프로세스 resume 및 후발 ToolSearch 사용이 양성 조건이다. 모호한 입력 거절만으로 완료 판정하지 않는다.

### AS4 재개·모델 변경·일반 자식과 명시 분기

정상 resume는 인증된 계정 세대·family·agent 범위에 저장된 thread ID와 이력 epoch를 확인한다.
`thread/resume.history`나 raw reasoning 재구성을 쓰지 않는다. 모델 변경은 idle 경계와 현재 계정 허용 목록을
확인하며 실패·경고를 보존한다. 새 thread나 다른 모델로 몰래 대체하지 않는다.

[공식 subagents 문서](https://code.claude.com/docs/en/sub-agents#what-loads-at-startup)는 non-fork 자식이
독립 맥락과 위임 메시지로 시작하고 conversation fork는 예외임을 명시한다. 일반 Agent는 인증된 family와
실제 agent-id에 독립 thread를 연결하고 Claude가 보낸 child context를 첫 입력으로 사용한다. 부모의 전체 이력을
추가 주입하지 않는다. 같은 자식의 후속 요청은 같은 thread에, 병렬·중첩 일반 자식은 각자의 thread에 연결한다.
`agent-identity-parallel-verdict.md`는 자식별 ID의 안정성을 입증하지만 부모 도구 호출 매핑은 입증하지 않는다.

명시 `--fork-session`은 launcher가 원본 family와 새 family를 알고 있는 제어 경로에서 수행한다. 원본의 인증된
계정·thread와 Claude가 분기할 exact public prefix를 저장된 `prefix digest → completedTurnID` 원장에 대조한다.
원장의 완료 경계에 정확히 대응할 때만 `thread/fork(lastTurnId=completedTurnID)`로 포함 범위를 고정한다.
부모가 이후 turn을 진행해도 최신 위치가 아니라 지정된 완료 경계로 분기하며, 반환 thread ID는 새 family에 저장한다.
미완료 도구·알 수 없는 prefix·다른 계정·변조 원장은 거절한다. 원장은 completed turn 결과와 함께 원자 저장한다.
설치 ThreadForkParams schema는 lastTurnId가 inclusive이며 in-progress turn을 사용할 수 없다고 명시한다.

Agent의 `subagent_type=fork` 및 `/subtask`는 일반 non-fork와 다르다. 공식 문서가 부모 이력 상속을 명시하므로
--fork-session 성공을 이 경로의 성공으로 확대하지 않는다. 인증된 부모/분기 경계를 native 표면과 연결하는
검증이 추가로 필요하며, 그 전에는 명시 미지원으로 처리한다. 부모를 프롬프트·시간·모델명으로 추측하지 않는다.
이 경계는 가용성 및 연결 실증 Gap이며 기존 native fork 양성 의무를 없애지 않는다. 설치본에서 기능을 찾지 못하면
NOT-RUN 전제 실패로 기록하고 AS4 및 전체 지원 완료를 보류한다. 일반 자식·명시 세션 분기 성공으로 대체하지 않는다.
설치 가용성 확인 뒤 native fork의 부모 이력·분기 위치·병렬/중첩 격리·재개와 잘못된 귀속 거절을 검증한다.
이 완료 게이트는 독립적인 다른 AS4 구현·검증을 중단하는 게이트가 아니다.

## context와 기능 표시

운영자가 API에도 App Server 출력 정책 사용을 승인했다. 구독·API 키 모두 공식 App Server 경로를 유지하며
Claude max_tokens의 동일 생성 상한 전달을 보장하지 않는다. MoAI 바이트·취소 제한은 별도로 적용한다.
API 별도 과금과 선택한 인증 방식을 표시하고 구독 실패를 API 경로로 자동 전환하지 않는다.

이전 direct endpoint에서 관측한 최대 921k 입력은 App Server 수용 한도로 승계하지 않는다.
오케스트레이터가 보고한 모델 metadata 872k는 출처가 다른 값이며, 설치 App Server model/list와 실제 turn 결과로 다시 판정한다.
UI에는 모델 명목 창, 현재 경로의 유효 한도, 누적 사용량을 구분한다. 1M 표시를 근거 없이 붙이지 않는다.
구독 출력은 서버 정책을 따르며 Claude max_tokens와 같은 생성량 보장을 주장하지 않는다.


### 4.4 문서·설치 버전·실증 기준

공식 기준은 https://learn.chatgpt.com/docs/app-server (접근 2026-09-12)다.
설치 Codex 0.154.0의 schema에서는 dynamicTools가 thread/start에만 있고 turn/start·resume에는 없다.
공식 문서에는 resume에서 새 dynamic tools를 주지 않으면 복원한다는 문구가 있어 버전 차이가 있다.
실제 실행 capability는 설치 schema와 검증 결과로 결정하고 문서 문구만으로 필드를 보내지 않는다.
thread/inject_items의 존재는 raw reasoning/history를 정상 경로에서 재구성할 이유가 아니다.

별도 실증 기록 `.moai/reports/t649/appserver-redesign/probe/isolation-result.json`을 읽어 확인했다.
그 기록은 Codex 0.154.0, account_type chatgpt, moai_echo 호출 1건과 최종 MOAI_APPSERVER_ECHO_OK,
errors=[]를 담는다. hooks/native 호출은 이 단일 실행에서 관측되지 않았지만 MCP startup status 이벤트가
1건 남아 있어 완전 격리 통과로 세지 않는다. 실제 Claude 도구 왕복·부정 실행 유도·재개·압축은 별도 게이트다.


## 5. launch provider signal (`REQ-MG-021`)

### 5.1 선택: `MOAI_LAUNCH_PROVIDER`

gateway launcher는 launch 시점의 초기 provider를 새 환경변수 `MOAI_LAUNCH_PROVIDER`로 자식
env에 싣는다. `internal/config/envkeys.go`에 `EnvMoaiLaunchProvider`로 등록한다. 값 어휘는
`claude` | `gpt` | `glm`이며, 활성 Dispatch/Orchestration이 provider 사실로 읽는다.

이름은 기존 규약을 따른다 — launcher가 운반하는 launch 사실은 `EnvMoaiSessionPID`처럼
`EnvMoai*` 상수에 `MOAI_*` 값을 쓴다. 이 트리에서
`MOAI_LAUNCH_PROVIDER` / `EnvMoaiLaunchProvider`는 사용처 0이다(`research.md` §1.4).

역사적 `EnvMoaiKanbanBackend` 주석의 “추론하지 말고 운반한다” 원칙만 provenance로 차용한다.
그 식별자와 명칭은 활성 계약이나 새 write가 아니다.

역사적 `MOAI_KANBAN_BACKEND`는 migration reader 외에 재사용하지 않는다. 일반 런타임은
`MOAI_LAUNCH_PROVIDER`와 `MOAI_DISPATCH_*`만 쓰며 폐기 입력은 env write 전에 거절한다.

### 5.2 정직한 경계와 판정 전제

env는 exec 시점에 고정된다. 따라서 이 signal은 **launch 시점의 provider만** 담는다. 세션 도중
`/model`로 바뀐 요청별 provider는 gateway 프로세스 안에서만 알 수 있으며, 이 SPEC은 그 값을
훅에 노출하지 않는다.

훅의 GLM 정리가 무엇을 위한 것인지는 gateway 도입으로 바뀐다.

- **gateway 이전(현재 코드).** 현재 프로덕션 launch 경로 중 GLM 키를 `settings.local.json`에 쓰는 것은
  없다. `applyGLMMode`는 env를 `setGLMEnv`로 프로세스에만 싣고 파일 쓰기를 의도적으로 뺐으며, `moai glm
  setup`은 `~/.moai/.env.glm`만 쓴다. 살아 있는 쓰기 주체는 SessionStart 훅 `ensureGLMCredentials`
  하나다 — 파일에 GLM 모델 슬롯이 남아 있으면 credential·base URL·context 창 키를 다시 써넣는다(§6.3).
  `injectGLMEnv`는 그 파일에 GLM credential과 base URL을 쓰던 옛 동작의 기록이지만 프로덕션 호출자가
  없다(`research.md` §1.5). 따라서 정리 대상은 현재 코드가 새로 만드는 키가 아니라, 옛 바이너리가 쓰거나
  사람이 고쳐 **남은 상태**다. SessionEnd 정리는 그 남은 키를 치우는 역할을 해 왔다.
- **gateway 이후.** 어떤 gateway launch도 그 파일에 base URL이나 GLM credential을 쓰지 않는다
  (§6.2). 이전 방식이 남긴 stale 키는 launch 단계 정리가 exec 전에 치운다(§6.1). 따라서 gateway
  세션의 훅에는 정리할 대상이 없다. gateway launch는 tmux 세션 env에 쓰지도 지우지도 않으므로, 그 세션의
  훅도 tmux 세션 env를 건드리지 않는다(§6.6).

그러므로 판정 기준은 "이 세션이 GLM인가"가 아니라 **"이 세션이 gateway launch인가"**다.
`MOAI_LAUNCH_PROVIDER`의 존재가 그 사실을 운반한다. 값(`claude`/`gpt`/`glm`)은 판정 기준이
아니라 guardrail 리마인더처럼 초기 provider가 필요한 곳에서 쓴다.

### 5.3 GLM 활성 판정 — 이관하는 둘, 넘기는 둘

이 SPEC이 이관하는 두 지점:

| 기구 | 지점 | 읽는 곳과 판정 | 이 SPEC의 처리 |
|---|---|---|---|
| 단일 부분 문자열 | `hookProcessEnvHasGLM` (`internal/hook`) | 훅 프로세스 env의 `ANTHROPIC_BASE_URL`에 `"z.ai"`가 들어 있는지만 본다. 토큰 판정은 없다 | signal이 있으면 `MOAI_LAUNCH_PROVIDER == glm`으로 판정. 없으면 기존 판정 유지 |
| 존재 | `cleanupGLMSettingsLocal` (`internal/hook`) | `settings.local.json`의 `env` 블록에 `ANTHROPIC_BASE_URL` **키가 있는지**만 본다 | signal이 있으면 정리를 수행하지 않음. 없으면 기존 동작 유지 |

`hookProcessEnvHasGLM`은 gateway 아래에서 base URL 부분 문자열로 GLM 세션을 알아볼 수 없다.
REQ-MG-006이 자식 env의 base URL을 loopback으로 두기 때문이다. 이것은 코드 판독에 근거한 추론이며
관측한 실패는 아니다.

**존재 술어의 쓰기 부작용.** SessionEnd는 프로젝트 디렉터리가 있으면 조건 없이
`cleanupGLMSettingsLocal`을 부르고, 이 함수는 키가 있으면 GLM 세션으로 판정해
`MOAI_BACKUP_AUTH_TOKEN` 복원 또는 `ANTHROPIC_AUTH_TOKEN` 삭제, base URL과 tier 모델 키 삭제를
수행한 뒤 **`settings.local.json`을 다시 쓴다**. 이 술어는 프로세스 env가 아니라 파일을 읽으므로,
결과는 파일에 무엇이 남는지에 달려 있다. gateway 아래에서는 §6의 계약으로 파일에 이 키가 남지
않고, signal이 있으면 정리 자체를 건너뛴다. 두 장치가 겹쳐 있어 어느 한쪽이 빠져도 사용자
설정이 조용히 다시 쓰이지 않는다.

이 SPEC이 **이관하지 않고 형제 SPEC에 넘기는** 두 지점:

| 지점 | 판정 | 넘기는 이유 |
|---|---|---|
| `sessionEnvHasGLM` (`internal/tmux`) | tmux 세션 env의 `ANTHROPIC_AUTH_TOKEN`이 비어 있지 않으면 먼저 참을 반환하고, 그렇지 않을 때만 `ANTHROPIC_BASE_URL`에 `"z.ai"`가 있는지 본다 | 비테스트 호출자가 `IsCGMode` 하나이고 `IsCGMode`에는 비테스트 호출이 없어 프로덕션에서 도달하지 않는다 |
| `hasGLMEnv` (`internal/tmux`) | 프로세스 env에 대해 같은 두 판정 | 위와 같다. 토큰 판정 쪽은 `SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001`의 시험을 살리려고 의도적으로 보존된 동작이다 |

두 판정이 gateway 아래에서 어떤 값을 내는지는 **단정하지 않는다.** 세션 접근 토큰을 어느 키로
싣느냐(§3.2)에 따라 결과가 뒤집히기 때문이다. 이 사실과 운반 키 결정은
`SPEC-MOAI-CG-RETIRE-001`(제안)에 넘기는 입력이다.

### 5.4 판정 규칙

| `MOAI_LAUNCH_PROVIDER` | SessionEnd `settings.local.json` GLM teardown (`cleanupGLMSettingsLocal`) | SessionEnd tmux 세션 env 정리 (`internal/hook`의 `clearTmuxSessionEnv`) | SessionStart `ensureGLMCredentials` (두 분기 모두, §6.3) | SessionStart `ensureTmuxGLMEnv`의 tmux 세션 env 쓰기 | guardrail 리마인더 |
|---|---|---|---|---|---|
| `glm` | 실행하지 않음 | 실행하지 않음 | 실행하지 않음 | 실행하지 않음 | 주입 |
| `claude` 또는 `gpt` | 실행하지 않음 | 실행하지 않음 | 실행하지 않음 | 실행하지 않음 | 주입하지 않음 |
| 없음 (gateway 이전 방식 launch, 또는 `claude` 직접 실행) | 기존 존재 술어 동작 유지 | 기존 동작 유지 (`TMUX`만 확인하고 지운다) | 기존 동작 유지 | 기존 동작 유지 | 기존 부분 문자열 판정 유지 |

마지막 행은 하위 호환이다. signal이 없다는 것은 "비-GLM"이 아니라 "gateway launch가 아님"이다.
gateway 이전 방식으로 연 세션의 정리를 끊지 않는다. 그 행에 남는 기존 위험은 이 SPEC이 새로
만든 것이 아니다.

`glm` 행에서도 teardown과 재주입을 실행하지 않는 이유: 재주입은 base URL을 `DefaultGLMBaseURL`로
파일에 다시 써서 gateway를 우회시킨다. teardown은 치울 것이 없다 — launch가 exec 전에 정리했고,
gateway launch는 파일에 쓰지 않으며, 재주입도 막혀 있기 때문이다.

**teammate 표시.** 위 표 밖의 SessionStart 쓰기 주체 `ensureTeammateMode`도 signal에 따라 달라진다. signal이 있으면
SessionStart 체인이 끝난 시점의 최상위 `teammateMode`는 `"in-process"`여야 하고, 이 함수는 `"tmux"`나 `"auto"`를 남겨서는
안 된다. signal이 없으면 오늘 동작(tmux 안 `"tmux"`, 밖 `"auto"`)을 유지한다(§6.6).

**tmux 두 동작을 억제하는 이유 (0.6.0 재서술).**

- **SessionEnd tmux 정리.** 훅의 `clearTmuxSessionEnv`는 SessionEnd 핸들러에서 `TMUX` 확인 말고는 조건 없이 불리며
  (`internal/hook/session_end.go:94`, 정의 `:651`), `ANTHROPIC_AUTH_TOKEN`·`ANTHROPIC_BASE_URL`·세 tier 모델 슬롯을 tmux
  세션 env에서 지운다(목록 `glmEnvVarsToClean` `:639-645`). gateway launch는 tmux 세션 env에 아무것도 쓰지 않으므로 그
  세션이 끝날 때 치울 것이 없다. 남아 있는 키가 있다면 같은 tmux 세션의 다른 launch(gateway 이전 방식의 `moai glm`, 형제
  SPEC이 철거하기 전의 `moai cg`)가 쓴 것이고, gateway 세션이 그것을 지우면 그 launch의 teammate pane이 credential을 잃는다.
  gateway 세션은 자기가 쓰지 않은 tmux 키를 지우지 않는다.
- **`ensureTmuxGLMEnv`.** 발동 조건은 `TMUX` 존재, `teammateMode == "tmux"`, `settings.local.json` `env`의 비어 있지 않은
  `ANTHROPIC_AUTH_TOKEN` 세 가지다(`internal/hook/glm_tmux.go:80`, `:101`, `:116-119`). 체인에서 이 함수(`session_start.go:638`)
  보다 먼저 도는 `ensureTeammateMode`(`:631`)가 signal 아래에서 `"in-process"`를 남기면 둘째 조건이 거짓이 되어 대개 입구에서
  돌아간다. **그래도 억제를 둔다.** `ensureTeammateMode`는 파일 읽기·해석·쓰기가 실패하면 쓰지 않고 빈 문자열을 돌려주며
  (`:1000-1001`, `:1006-1007`, `:1060-1064`), 같은 프로젝트의 다른 세션이 파일을 `"tmux"`로 다시 쓸 수도 있다. 그런 경우
  launch 정리가 `MOAI_BACKUP_AUTH_TOKEN`에서 복원한 사용자 토큰이 `buildGLMTmuxEnvVars`(`glm_tmux.go:39-58`)를 거쳐 민감 값
  채널로 tmux 세션 env에 들어갈 수 있다. M0에서 별도 헤더를 선택했으므로 세션 토큰 자체를 이 AUTH_TOKEN 키에
  넣는 분기는 없다. 복원된 파일 토큰의 전달 가능성은 코드 판독에 근거한 추론이며 이 변형을 실행해 보지 않았다.

## 6. `settings.local.json`과 tmux 세션 env 계약 (`REQ-MG-021`, `REQ-MG-022`)

### 6.1 launch 단계 정리 — 두 층 계약과 비회귀 규칙

세 gateway launcher는 `moai glm`을 포함해 모두 exec 전에 `settings.local.json`의 `env` 블록에서 아래
키를 정리한다. 이 집합은 기존 함수 하나의 목록을 옮긴 것이 아니다. 저장소에 있는 GLM 삭제 목록 네
개는 서로 같은 키 집합이 아니며(`research.md` §6.5), 이 SPEC은 그중 어느 것도 정답으로 고르지 않고
쓰기 주체에서 집합을 유도한다.

**층 (a) — 라우팅 키 (필수).** 남으면 요청 경로나 인증을 바꾸는 키다.

| 동작 | 키 |
|---|---|
| 조건부 복원 | `MOAI_BACKUP_AUTH_TOKEN`이 비어 있지 않으면 그 값을 `ANTHROPIC_AUTH_TOKEN`에 넣는다. 백업이 없거나 비어 있으면 `ANTHROPIC_AUTH_TOKEN`을 지운다. 어느 경우든 `MOAI_BACKUP_AUTH_TOKEN` 키는 값이 비어 있어도 지운다 |
| 삭제 | `ANTHROPIC_BASE_URL` |
| 삭제 | `ANTHROPIC_DEFAULT_OPUS_MODEL`, `ANTHROPIC_DEFAULT_SONNET_MODEL`, `ANTHROPIC_DEFAULT_HAIKU_MODEL`, `ANTHROPIC_DEFAULT_FABLE_MODEL` |

이 행은 파일 계약이다. gateway `moai glm`은 같은 네 슬롯 키를 Claude child env에만 설정 값으로 더하고 파일에는 쓰지 않는다
(아래 프로세스 env 투영, §6.7). 세 launcher의 슬롯·picker 구성은 t649 코어 계약이다(§6.7).

**층 (b) — 동작 영향 키 (쓰기 주체에서 유도).** 요청 경로는 바꾸지 않지만 세션 동작을 바꾸는 키다.
`injectGLMEnv`가 쓰는 키(`internal/cli/glm.go:1000-1030`)와 `ensureGLMCredentials`가 쓰는 키(§6.3)의
합집합에서 층 (a)를 뺀 것이다.

| 키 | 쓰는 주체 |
|---|---|
| `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS` | `injectGLMEnv`, `ensureGLMCredentials` |
| `API_TIMEOUT_MS` | `injectGLMEnv` |
| `CLAUDE_CODE_AUTO_COMPACT_WINDOW` | `injectGLMEnv`, `ensureGLMCredentials` |
| `CLAUDE_CODE_MAX_CONTEXT_TOKENS` | `injectGLMEnv`, `ensureGLMCredentials` |

**`CLAUDE_CODE_MAX_CONTEXT_TOKENS`는 포함한다.** 0.3.0은 이 키를 넣을지 열어 두었다. 훅은 오늘도 이 키를
파일에 쓰는데 세 파일 정리 함수 어느 것도 이 키를 지우지 않는다(§6.5). 넣지 않으면 gateway 세션이 이전
GLM 모델의 context 창 선언을 그대로 물려받을 수 있다.

**출처의 한계.** 층 (b) 가운데 `injectGLMEnv` 쪽 근거는 지금은 호출되지 않는 쓰기 함수의 기록이다. 옛
바이너리 각각이 정확히 무엇을 썼는지를 증명하지는 않는다. 이 층은 "알려진 쓰기 주체가 쓰는 키"를 덮는다는
뜻이지, 과거 모든 버전이 쓴 키를 망라한다는 뜻이 아니다.

**비회귀 규칙.** 어떤 gateway launch도 오늘 `moai cc`가 지우는 키보다 적게 지우지 않는다. 그래서
`removeGLMEnv`(`internal/cli/launcher.go:380-435`)가 지우는 키 가운데 위 두 층에 없는 세 키를 더한다.

| 키 | 근거 |
|---|---|
| `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC` | `removeGLMEnv`가 지운다 |
| `CLAUDE_CODE_TEAMMATE_DISPLAY` | `removeGLMEnv`가 지운다 |
| `MOAI_STATUSLINE_CONTEXT_SIZE` | `removeGLMEnv`가 문자열 리터럴로 지운다. 상수 `config.EnvStatuslineContextSize`(`internal/config/envkeys.go:64`)와 같은 키다 |

위 정리 뒤 `env` 블록이 비면 `env` 키 자체를 지운다.

**최종 집합.** 층 (a) 7키(`MOAI_BACKUP_AUTH_TOKEN` 포함) + 층 (b) 4키 + 비회귀 3키 = **14키**. 이것이
**GLM 정리 키 집합**이다(`spec.md` §C). 결과적으로 `removeGLMEnv`의 `env` 키 13개에
`CLAUDE_CODE_MAX_CONTEXT_TOKENS`를 더한 것과 같다. `AC-MG-018`은 이 14키의 투영으로 판정한다.

`removeGLMEnv`는 최상위 `teammateMode`도 지운다. 이 키는 `env` 블록 밖이며 이 계약에 들지 않는다. gateway
launch에서 이 키가 어떤 값으로 끝나야 하는지는 teammate 표시 계약(§6.6)이 정한다.

**구현에 대한 함의.** 기존 `removeGLMEnv`만 호출하면 `CLAUDE_CODE_MAX_CONTEXT_TOKENS`가 남아 이 계약을
채우지 못한다. 또 `removeGLMEnv`는 백업 값이 비어 있으면 `MOAI_BACKUP_AUTH_TOKEN` 키를 지우지 않고 남긴다
(`internal/cli/launcher.go:406-411`). 그대로 쓰면 빈 백업 키가 남아 조건부 복원 행을 채우지 못한다. 어떻게 채울지는 run 단계가 정한다.

**프로세스 env 투영 (0.7.0, iter5 G5-B1).** 같은 14키 집합을 Claude child env의 상속 기반에도 적용한다(§3.2 6a). 파일
계약과 달리 프로세스 env에는 복원할 백업이 없으므로 14키를 모두 지운다 — `ANTHROPIC_AUTH_TOKEN`과 `MOAI_BACKUP_AUTH_TOKEN`도
지운다. 집합을 새로 만들지 않고 이 정의 하나를 쓴다. 두 tmux GLM 쓰기 주체가 tmux 세션 env에 넣는 키가 모두 이 집합 안에
있기 때문이다(HEAD `81c1d58f9` 판독).

| 쓰기 주체 | 넣는 키 | 14키 집합 밖의 키 |
|---|---|---|
| `buildTmuxInjectVars` (`internal/cli/glm.go:507-539`) | `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL`, 네 `ANTHROPIC_DEFAULT_*_MODEL` 슬롯, `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`, `API_TIMEOUT_MS`, `MOAI_STATUSLINE_CONTEXT_SIZE`, `CLAUDE_CODE_AUTO_COMPACT_WINDOW`, `CLAUDE_CODE_MAX_CONTEXT_TOKENS` | 없음 |
| `glmTmuxKeys` (`internal/hook/glm_tmux.go:21-31`) | `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL`, OPUS·SONNET·HAIKU 슬롯, `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`, `API_TIMEOUT_MS`, `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`, `MOAI_STATUSLINE_CONTEXT_SIZE` | 없음 |

집합 밖의 키 하나는 launcher마다 다르게 다룬다. 오늘 `setGLMEnv`는 GLM 키를 `Z_AI_API_KEY`로 launcher 프로세스 env에
넣는다(`internal/cli/glm.go:386`, 주석 `:385`). `moai cc`와 `moai gpt`는 상속된 `Z_AI_API_KEY`를 지우고, gateway `moai glm`은
상속 값을 지운 뒤 GLM credential 저장소에서 읽은 값을 싣는다(이유는 §6.5).

**`moai glm`이 다시 더하는 네 키 (0.7.0, 2026-09-11 운영자 결정).** 지운 14키 가운데 네 모델 슬롯 키
`ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL`은 gateway `moai glm`에서만 launcher가 정리 뒤에 다시 더한다. 값은 상속
값이 아니라 `llm.glm.models`의 `high`·`medium`·`low`·`fable`이다(§3.2 6b, `spec.md` `REQ-MG-021`). `moai cc`와 `moai gpt`는
해당 제공자의 네 슬롯을 더한다(§6.7, t649).

### 6.2 기록 금지

어떤 gateway launch도, `moai glm`까지 포함해, `ANTHROPIC_BASE_URL`이나 GLM credential을
`settings.local.json`에 쓰지 않는다. `injectGLMEnv`는 이미 프로덕션 호출자가 없는 옛 쓰기 함수이며
(`research.md` §1.5), gateway launch도 그것을 되살려 호출하지 않는다.

- GLM upstream 인증에 쓰는 credential은 GLM `CredentialRef`가 기존 `internal/glmcred` 저장소에서 읽어 gateway child
  안에서 쓴다. GLM credential과 Z.AI base URL은 `settings.local.json`과 tmux 세션 env(§6.6)에 싣지 않는다. Claude child
  env에도 싣지 않되, gateway `moai glm`만 launcher가 GLM credential 저장소에서 읽은 `Z_AI_API_KEY`를 MCP 도구 인증용으로
  싣는다(§6.1 프로세스 env 투영, §6.5).
- Claude child의 base URL은 언제나 자식 env의 loopback gateway 주소다.
- Z.AI beta 헤더 제거는 gateway 안쪽 책임이고, tier 매핑은 launcher가 Claude child env에 더하는 GLM tier 슬롯 키와 registry
  exact match가 맡는다(`REQ-MG-018`, §6.7). 어느 쪽도 파일에 쓸 이유가 없다.

### 6.3 SessionStart 억제 — `ensureGLMCredentials`의 두 분기 모두

SessionStart의 `ensureGLMCredentials`는 `settings.local.json`의 `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU}_MODEL`
값 중 하나에 `"glm"`이 들어 있을 때만 동작한다. FABLE 슬롯은 보지 않고, 훅 로컬 `isCGMode`가 참이면
건너뛴다. 동작하면 두 분기 중 하나로 파일에 쓴다(`research.md` §6.5).

| 분기 | 파일에 쓰는 키 |
|---|---|
| 토큰 있음 | `CLAUDE_CODE_AUTO_COMPACT_WINDOW`, `CLAUDE_CODE_MAX_CONTEXT_TOKENS` — 두 값 중 어느 하나라도 바뀌면 다시 쓴다 |
| 토큰 없음 | `ANTHROPIC_AUTH_TOKEN`(`~/.moai/.env.glm`에서 읽음), `ANTHROPIC_BASE_URL`(비어 있으면 `DefaultGLMBaseURL`), `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS="1"`, `CLAUDE_CODE_AUTO_COMPACT_WINDOW`, `CLAUDE_CODE_MAX_CONTEXT_TOKENS` |

토큰 없음 분기는 gateway 세션의 base URL을 Z.AI로 되돌릴 수 있다. 토큰 있음 분기는 base URL을 건드리지
않지만 context 창 키 두 개를 파일에 쓴다.

`MOAI_LAUNCH_PROVIDER`가 설정된 세션에서 억제는 **이 함수가 파일에 하는 모든 쓰기**를 덮어야 한다.
credential 주입만 막으면 gateway GLM launch가 토큰 있음 분기를 통해 여전히
`CLAUDE_CODE_MAX_CONTEXT_TOKENS`와 `CLAUDE_CODE_AUTO_COMPACT_WINDOW`를 파일에 쓴다. 그래서 억제 판정은
분기 이전, 함수 입구에서 한 번 한다(§5.4). context 창 보조 함수 둘(`maybeSet1MAutoCompactWindow`,
`maybeDeclareGLMContextWindow`)은 이 함수 안에서만 호출되므로 함께 억제된다.

launch 단계 정리(§6.1)가 모델 슬롯을 지우고 나면 이 함수는 입구의 슬롯 판정에서 이미 돌아간다. 그래서 억제를
빠뜨린 구현도 평범한 launch → SessionStart 흐름에서는 두 분기에 닿지 않아 초록이 된다. `AC-MG-018` (c)가
모델 슬롯을 남긴 픽스처로 이 함수를 직접 부르고, signal 없는 대조군으로 두 분기에 실제로 닿는지 확인하는
이유다.

SessionStart 쓰기 체인의 나머지 쓰기 주체는 GLM 정리 키 집합을 쓰지 않는다.
`ensureTeammateMode`는 최상위 `teammateMode`를 다시 쓸 때 `env`의 legacy `CLAUDE_CODE_TEAMMATE_DISPLAY`
(GLM 정리 키 집합의 구성원)도 지우지만, 값이 이미 맞으면 쓰기 전에 돌아가 그 삭제도 하지 않는다. 그 키를
쓰지는 않으며, launch 정리가 이미 지운 키이므로 투영을 바꾸지 않는다. Windows 전용 `injectCLAUDEEnvFile`은
`CLAUDE_ENV_FILE`만 쓴다(`research.md` §6.2, §6.3). `ensureTmuxGLMEnv`는 파일이 아니라 tmux 세션 env에
쓴다(§5.4, §6.6).

### 6.4 우선순위는 미측정이다

settings 파일의 `env` 값이 Claude Code 프로세스 env보다 우선하는지는 **측정하지 않은 Claude Code
동작**이다. 이 SPEC은 어느 쪽도 전제로 삼지 않는다.

- settings가 우선한다면: 정리하지 않은 stale `ANTHROPIC_BASE_URL`이 자식 env의 loopback 주소를
  덮어써 요청이 gateway를 거치지 않고 Z.AI로 곧장 간다. gateway 로그에는 아무것도 남지 않는다.
  `REQ-MG-022`가 금지한 경로다.
- 프로세스 env가 우선한다면: stale 키는 요청 경로를 바꾸지 않지만, 훅이 파일을 읽어 틀린 판정을
  내리는 원인으로 남는다.

두 경우 모두 launch 단계 정리가 필요하다. **정리는 우선순위를 측정하지 않았기 때문에 요구된다.**

### 6.5 잔여 위험

- **복원된 백업 토큰.** 파일 정리는 `MOAI_BACKUP_AUTH_TOKEN`을 settings `env`의 `ANTHROPIC_AUTH_TOKEN`으로 되돌린다.
  프로세스 상속 값은 §3.2 6a에서 제거되고 M0가 선택한 세션 인증은 별도 헤더다. settings `env`가 적용되는 경우 복원된
  사용자 토큰이 loopback 요청에 실릴 가능성은 남는다. 이것은 그 토큰을 OAuth passthrough credential로 채택할 근거가 아니다.
  허용된 Claude OAuth 자격과 임의 복원 토큰의 출처 구별·설정 우선순위는 실제 통합 음성군으로 확인해야 하며,
  M0의 정상 로그인 관측만으로 해당 오염 변형의 안전성을 선언하지 않는다.
- **상속된 인증 env를 쓰던 구성.** 사용자가 셸에서 `ANTHROPIC_AUTH_TOKEN`을 내보내 두었다면 그 값은 gateway launch의
  Claude child에 닿지 않는다(§3.2 6a). gateway의 upstream 인증은 `CredentialRef`로만 해석하므로(`REQ-MG-023`) 이 삭제와
  무관하지만, Claude Code 자신이 그 값을 쓰던 구성에 미치는 영향은 M0 측정과 운반 키 결정이 함께 확인한다.
- **`moai glm` 세션의 Claude 프로세스 env에 있는 Z.AI 키.** gateway `moai glm`은 GLM credential 저장소에서 읽은
  `Z_AI_API_KEY`를 Claude child env에 싣는다(`REQ-MG-021`의 좁은 예외). 오늘 `setGLMEnv`가 같은 키를 launcher 프로세스
  env에 넣는 이유와 같다(`internal/cli/glm.go:385` 주석 'Z.AI MCP server (zai-mcp-server) reads this env for
  authentication'). `moai glm tools`는 Z.AI MCP 서버를 기본값인 사용자 범위 `~/.claude.json`에 등록하고
  (`internal/cli/glm_tools.go:188`, 경로 해석 `:371-376`), HTTP 항목의 인증 헤더는 리터럴 `Bearer ${Z_AI_API_KEY}`다
  (`:60`, 항목 구성 `:405-413`). Claude Code가 이 리터럴을 프로세스 env에서 채운다는 것은 코드 주석(`:59`, `:404`) 판독이며
  실행으로 확인하지 않았다. 그래서 이 키는 `moai glm` 세션의 Claude 프로세스와 그 자식 프로세스 env에 있다. 같은 이유로
  `moai cc`·`moai gpt`는 상속된 키를 지운다 — 남으면 사용자 범위 MCP 항목이 GLM을 고르지 않은 세션에서 유료 Z.AI를
  부른다(`REQ-MG-022`).
- **남은 상태는 gateway launch 밖에서도 생긴다.** stale 키의 출처는 현재 코드가 아니라 옛 바이너리와
  수동 편집이다. launch 단계 정리(§6.1)는 gateway launch 때만 돈다. gateway 이전 방식으로 연 세션이나
  `claude` 직접 실행 세션에는 기존 정리 동작만 남는다(§5.4 마지막 행).
- **기존 파일 정리 함수의 간극 — 이 SPEC이 만든 것이 아니다.** `CLAUDE_CODE_MAX_CONTEXT_TOKENS`는 훅이
  파일에 쓰고(과거에는 `injectGLMEnv`도 썼다) `removeGLMEnv`·`stripGLMCredsAndSetTeammateMode`·
  `cleanupGLMSettingsLocal` 어느 것도 지우지 않는다. 훅이 이 키를 쓰는 것은 파일에 GLM 모델 슬롯이 남아
  있을 때뿐이므로, 그런 파일을 거친 뒤 `moai cc`로 정리해도 이 키가 남는 경우가 생긴다. 코드 판독에
  근거한 추론이며 런타임에서 재현하지 않았다. gateway launch는 §6.1로 이 키를 지우지만, gateway 밖
  경로의 간극은 그대로 둔다(`research.md` §6.5).
- **판정 입력이 전부 정적 판독이다.** 세션 구간에 GLM 정리 키 집합을 쓰는 다른 경로가 없다는
  근거는 non-test Go 검색이다. 런타임에서 관측하지 않았다.
- **lead 밖에서 gateway 주소를 가진 프로세스.** in-process teammate가 lead의 gateway 주소를 물려받는지, lead 종료 뒤
  남는지는 미측정이다(`plan.md` M7). 남는다면 그 작업은 loopback 연결 오류로 중단된다(`REQ-MG-008`). 드러나는
  실패이지만 작업은 끊긴다.
- **`teammateMode`는 프로젝트 공유 파일의 키다.** 같은 프로젝트에서 gateway 세션과 gateway가 아닌 세션이 함께 돌면,
  나중에 시작한 세션의 SessionStart가 `teammateMode`를 자기 값으로 다시 쓴다. 이미 떠 있는 세션이 이 값을 언제 다시
  읽는지는 측정하지 않았다. gateway 세션이 뒤늦게 split pane으로 teammate를 띄우면 그 pane은 gateway 주소를 tmux 세션
  env로 받지 못한다(이 SPEC은 tmux 세션 env에 쓰지 않는다). base URL이 없을 때의 목적지는 Claude Code 기본 동작이며 이
  SPEC이 잰 것이 아니다. tmux 세션 env에 다른 launch가 남긴 GLM 키가 있으면 그 pane은 그 키로 Z.AI에 곧장 갈 수 있다(iter5
  G5-B3). 그래서 `REQ-MG-022`는 in-process 한정이 이 경로를 좁힐 뿐 닫지 못한다고 적고, 닫는 일을 형제
  `SPEC-MOAI-GATEWAY-TEAMMATE-001`(제안)에 넘긴다.
- **stale tmux GLM 키를 gateway launch가 치우지 않는다.** 오늘 `moai cc`는 launch 때 CLI `clearTmuxSessionEnv`로 tmux
  세션 env의 GLM 키를 지운다(`internal/cli/launcher.go:224`). gateway launch는 tmux 세션 env를 쓰지도 지우지도 않으므로 그
  정리를 하지 않는다. 남은 GLM 키는 새 pane의 프로세스 env로 들어올 수 있지만 gateway 세션의 Claude child에는 닿지 않는다 —
  launcher가 자식 env를 조립하기 전에 상속 env에서 GLM 정리 키 집합과, `moai cc`·`moai gpt`에서는 `Z_AI_API_KEY`까지 지우기
  때문이다(`REQ-MG-021`, §3.2 6a). `moai glm`이 정리 뒤 다시 더하는 네 모델 슬롯 키도 tmux에 남은 값이 아니라
  `llm.glm.models` 값이다(§6.7). 다만 같은 tmux 세션에서 사용자가 새 pane으로 직접 띄운 `claude`는 그 키로 Z.AI에 갈 수
  있고, 뒤늦게 split pane으로 뜬 teammate도 마찬가지다(위 항목). 오늘 `moai glm` 뒤 새 pane에서 벌어지는 일과 같다. 이 정리를 누가
  맡을지는 형제 `SPEC-MOAI-GATEWAY-TEAMMATE-001`(제안)의 tmux 세션 env 소유권 설계와 함께 정한다.
- **teammate UX 변화.** `moai glm` gateway launch의 teammate는 tmux pane으로 나타나지 않고 lead 터미널 안에서 돈다
  (`REQ-MG-018`). 형제 SPEC이 착지할 때까지의 의도한 차이다.

### 6.6 tmux 세션 env와 teammate 표시 — gateway launch는 in-process teammate만 쓴다 (0.6.0)

**결정.** gateway launch는 teammate 표시로 in-process만 허용하고, tmux 세션 env에는 어떤 키도 쓰거나 지우지 않는다
(`plan.md` 결정 12). 0.4.0~0.5.0의 계약 — tmux 세션 안에서 15키를 지운 뒤 loopback 주소와 signal을 tmux 세션 env에 싣는
방식 — 은 iter4 감사가 소유권 없는 공유 상태로 판정했다(G4-B1). tmux 세션 하나에는 env가 하나뿐인데 `--spawn`은 같은
세션에 새 창을 연다(`internal/cli/spawn.go:78-79` 주석 "in the caller's current session", 호출 `:88`). 두 번째 gateway
launch가 같은 키를 덮어쓰고, 마지막 gateway가 끝난 뒤에는 `MOAI_LAUNCH_PROVIDER`가 남아 뒤이어 뜬 gateway 아닌 세션의
SessionEnd GLM 정리를 억제한다. 그 계약은 통째로 형제 `SPEC-MOAI-GATEWAY-TEAMMATE-001`(제안)로 옮겼고, 이 SPEC에는 아래만
남는다.

**오늘 코드가 하는 일 (판독, HEAD `ed71054d3`).**

- `ensureTeammateMode`(`internal/hook/session_start.go:995`)는 매 SessionStart에 `TMUX` 유무(`:996`)로 원하는 값을 정한다.
  기본값은 `"auto"`(`:1021`)이고 tmux 안이면 `"tmux"`(`:1022-1023`)다. 현재 값과 같으면 쓰지 않는다. 체인의 호출은 `:631`이다.
- 호출부 주석(`:628-630`)은 tmux 밖의 `"auto"`를 "in-process display"라 부르고, 함수 주석(`:989`)은 tmux 밖이면 "removes
  override"라고 적는다. 코드는 키를 지우지 않고 `"auto"`를 쓴다. 주석과 코드가 다르며, 이 SPEC은 코드를 따른다.
- Claude Code 문서는 `teammateMode` 기본값을 `"in-process"`로, `"auto"`를 "tmux 세션 안이거나 `it2` CLI가 있는 iTerm2이면
  split pane, 그 밖에는 in-process"로 설명한다(`research.md` §16). 그 설명대로라면 `"auto"`는 tmux 세션 안에서 split pane을
  연다. 문서 판독이며 2.1.267에서 측정하지 않았다.
- `teammateMode`를 `"tmux"`로 쓰는 CLI 경로는 CG 모드의 `stripGLMCredsAndSetTeammateMode`(`internal/cli/settings.go:206`, 호출
  `internal/cli/launcher.go:345`)이고, `moai cc`의 `removeGLMEnv`는 이 키를 지운다(`launcher.go:402`).

**gateway 계약.**

1. **표시 방식.** gateway launch로 시작한 세션에서는 tmux 세션 안이든 밖이든, SessionStart 체인이 끝난 시점의
   `settings.local.json` 최상위 `teammateMode`가 `"in-process"`다. `"tmux"`와 `"auto"`는 둘 다 split pane을 고를 수 있으므로
   남기지 않는다. 그 값을 누가 쓰는지(launcher의 exec 전 쓰기, signal을 읽는 `ensureTeammateMode`, 또는 둘 다)는 run 단계가
   정한다. signal이 없는 세션은 오늘 동작을 유지한다.
2. **tmux 세션 env 무기록.** gateway launch는 tmux 세션 env에 어떤 키도 쓰거나 지우지 않는다. 그래서 GLM credential, Z.AI
   base URL, loopback 주소, `MOAI_LAUNCH_PROVIDER` 어느 것도 gateway launch를 통해 그 env에 실리지 않는다. `applyGLMMode`의
   tmux 주입(`launcher.go:278-284`)과 `applyCCMode`의 CLI tmux 정리(`launcher.go:224`)는 gateway launch에서 돌지 않는다.
3. **훅 억제.** signal이 있으면 SessionEnd tmux 정리와 `ensureTmuxGLMEnv`의 tmux 쓰기를 억제한다(§5.4 표와 이유).

**미측정 — `plan.md` M7의 측정 과제.** in-process teammate가 lead 프로세스의 env, 특히 loopback `ANTHROPIC_BASE_URL`과
`MOAI_LAUNCH_PROVIDER`를 요청 경로와 훅 env로 물려받는지. 문서는 teammate가 lead의 대화 기록을 물려받지 않는다고만 적고 env
상속은 말하지 않는다(`research.md` §16). 음성이면 in-process teammate의 요청이 gateway를 거치지 않을 수 있으므로, 그 결과는
구현을 이어 가기 전에 운영자에게 blocker로 돌아간다. 우회해서 해결하지 않는다.

**형제 SPEC으로 옮긴 것.** tmux pane teammate의 연결 방식(tmux 세션 env 주입, 15키 tmux GLM 정리 키 집합, 두 tmux 정리
함수의 `ANTHROPIC_AUTH_TOKEN` 불일치 해소, 운반 키의 tmux 전달), 같은 tmux 세션의 여러 gateway에 대한 소유권 모델, teammate
pane 수명과 생존 판정, teammate pane 요청의 모델 ID와 GLM tier 매핑, `CLAUDE_CONFIG_DIR`을 tmux에서 지우는 영향(iter4
G4-A8). 0.5.0까지의 표와 판독 근거는 git 이력의 이 절과 `research.md` §6.6에 남는다.

### 6.7 제공자별 picker·모델 슬롯·저장 격리 (t649)

공통 catalog에서 launcher 제공자만 남긴 세션 snapshot을 만든다. 명시 ID도 이 snapshot으로 검증하며
요청 ingress는 UI와 독립적으로 같은 경계를 강제한다. gpt-6-astra를 정확한 route ID로 쓴다.
자식 전용 `--settings`의 `modelPicker.replaceBuiltInOptions`와 유효 options로 목록을 구성한다.
시작 `--model`, 네 `ANTHROPIC_DEFAULT_*_MODEL` 슬롯, fallback은 모두 해당 snapshot 안에 있어야 한다.
Default·현재 행과 `[1m]` 등 클라이언트 접미사는 실제 요청 캡처로 판정하며 임의 model ID 별칭을 추가하지 않는다.

상속 모델 키는 먼저 제거하고 제공자별 새 슬롯을 더한다. GLM tier 매핑은 기존 high·medium·low·fable을
따르고 cc·gpt도 네 슬롯을 같은 제공자 ID로 설정한다. 이 값을 공유 settings나 tmux env에 쓰지 않는다.
모델 선택·캐시·resume 상태는 제공자와 대화 소유 경계를 가진 사설 설정에 저장한다. 기존 인증·프로젝트
설정을 보존하고 상위 managed 정책을 우회하지 않는다. 충돌·빈 목록은 명시 진단하며 gateway 제한은 유지한다.

실제 MoAI launcher PTY에서 세 목록·직접 입력·Enter·s·새 세션·재개를 각각 측정한다.
직접 Claude 임시 실험은 조사 보고서에 보존하며 제품 성공 근거로 승격하지 않는다.


## 7. [SUPERSEDED BY 0.15.0 §12.3] 옛 kanban backend 설계 (역사 기록)

이 절의 현재형 문장은 0.15.0 이전 결정의 원문을 보존한 것이며 구현 지시가 아니다. 활성 설계는 §12.3의
Factory/Todo/Tasks/Dispatch/Orchestration 계약과 `-k` 무부작용 폐기다.

### 7.1 [SUPERSEDED BY 0.15.0 §12.3] `BackendGPT` 옛 추가안

`internal/kanban`의 backend 상수 집합에 `BackendGPT = "gpt"`를 더한다.

이 추가는 기존 두 값의 뜻도 바꾼다. 현재 상수 선언 위 주석은 "A mixed-backend session is not a
kanban session, so no third value exists"라고 적는다. gateway 아래에서는 이 문장의 두 부분이
모두 거짓이 된다 — 세 번째 값이 생기고, 모든 세션이 `/model`로 provider를 섞을 수 있다.

따라서 kanban 기록의 `Backend` 필드는 **세션의 provider**가 아니라 **launcher의 초기 provider**를
뜻하도록 재정의한다. 이 재정의는 `gpt`뿐 아니라 `claude`·`glm`에도 똑같이 적용된다. `moai cc -k`로
시작해 `/model`로 GLM을 쓴 lane의 기록은 여전히 `claude`다.

문구를 고쳐야 하는 곳:

- `internal/kanban`의 backend 상수 선언 위 주석 — "no third value exists" 불변식을 새 의미로 다시 쓴다.
- `Record.Backend` 필드 주석 — "BackendClaude or BackendGLM"을 세 값과 새 의미로.
- `EnvMoaiKanbanBackend` 상수 주석 — 같은 열거를 세 값으로.

### 7.2 [SUPERSEDED BY 0.15.0 §12.3] 이름이 같은 다른 개념

`BackendGPT`는 **`internal/kanban`에만** 추가한다.

- `internal/cli/mcp_convergence.go`에도 이름이 같은 `BackendClaude`·`BackendGLM` 상수가 있지만,
  그것은 **감사(audit) backend** 집합(`claude`·`codex`·`glm`)이다. cross-model 감사의 수렴 대상이며
  launcher provider와 무관하다. 여기에 GPT 값을 넣으면 감사 수렴이 존재하지 않는 backend를 기다린다.
- `internal/cli/model.go`의 모델 프로필 보고서 `Backend` 필드(`"claude"`/`"glm"`)도 LLM 설정에서
  유도한 별개 개념이다. 이 SPEC의 범위가 아니다.

### 7.3 [SUPERSEDED BY 0.15.0 §12.3] 옛 소비자 표

`moai cc`와 `moai glm`의 진입 경로는 네 분기로 나뉜다 — factory lead, factory worker, kanban lead,
kanban companion. 네 분기 모두 `exportKanbanLaunchFacts`에 backend 값을 넘기고, factory lead 분기만
추가로 `RecordFactoryRunStart`를 부른다. `moai gpt`는 같은 네 분기를 같은 형태로 갖는다.

| 소비자 | 현재 | 필요한 변화 |
|---|---|---|
| `moai cc`의 네 진입 분기 | `kanban.BackendClaude` 전달 | 없음 (의미만 재정의) |
| `moai glm`의 네 진입 분기 | `kanban.BackendGLM` 전달 | 없음 (의미만 재정의) |
| `moai gpt`의 네 진입 분기 | 없음 | 같은 형태로 신설해 `kanban.BackendGPT` 전달 |
| `exportKanbanLaunchFacts` | 받은 값을 `MOAI_KANBAN_BACKEND`로 export | 없음 (값 어휘만 확장) |
| SessionStart 기록 작성 | `MOAI_KANBAN_BACKEND`를 읽어 `kanban.NewRecord`에 전달 | 없음 |
| 웹 콘솔 factory lane·운영 뷰 | 기록의 `Backend`를 그대로 표시 | 없음 |
| 웹 콘솔 `backendBadge` | `glm`이면 "metered", **그 밖의 모든 값은 "flat rate"** | `gpt`에 대해 과금 방식을 단정하지 않도록 분기 추가 |

마지막 행은 사소하지 않다. 현재 배지는 `glm`이 아닌 값을 모두 정액제로 표시하므로, 손대지 않으면
`gpt` lane이 과금 방식에 대한 검증되지 않은 주장을 화면에 띄운다. GPT 구독 경로와 API key 경로는
과금 방식이 다르고, 그 판정은 형제 SPEC의 T08 소관이다. 이 SPEC에서 배지는 `gpt`에 대해 과금
방식을 단정하지 않는 표시로 둔다(`spec.md` `REQ-MG-026`).

또한 `moai cg`의 kanban 거부 메시지는 "one-session / one-backend / one-chain premise"를 근거로
든다. gateway 아래에서는 "one-backend" 전제가 `cc`·`glm`에서도 성립하지 않지만, 그 메시지는
`moai cg` 철거와 함께 사라지므로 형제 SPEC `SPEC-MOAI-CG-RETIRE-001`(제안)에서 처리한다.

### 7.4 [SUPERSEDED BY 0.15.0 §12.3] 당시 출시 판단 입력

t650~t654의 managed/API 인증, 도구 실행권, pending 복구, resume/fork/compact, 실제 세 picker와 Windows 실행을
각각 증거로 판정한다. AUTH/PICKER 형제의 문서 착지만으로 지원을 선언하지 않는다. 사용자 인증 오류에는 내부 SPEC명을
표시하지 않는다. 일반 Claude 설정·Codex 타 계정 프로필 격리를 제품 실행 전후로 비교한다.

## 8. 재사용할 선례와 재사용하지 않을 것

| 선례 | 위치 | 판단 |
|---|---|---|
| loopback bind 상수·`Port: 0`·`NewServer`/`Handler()` 분리·mutex listener·SIGINT/SIGTERM + 5s drain | `internal/web/server.go` | 재사용 |
| 포트 점유 탐지·회수 (bind probe, `EADDRINUSE`/WSAEADDRINUSE 분기, moai 프로세스일 때만 회수) | `internal/cli/web_port.go` | 재사용 |
| 프로세스 신원 지문 (`CurrentProcessFingerprint`, `ProbeProcessIdentity`) | `internal/homestate` | 재사용 — 고아 수거의 PID 재사용 방어 |
| GLM 키 파일 정리 (`removeGLMEnv`) | `internal/cli/launcher.go` | 부분 재사용 — 세 gateway launcher의 launch 단계 정리는 이 함수가 지우는 키를 모두 포함하되 `CLAUDE_CODE_MAX_CONTEXT_TOKENS`를 더한다(§6.1) |
| Host 헤더 / `Sec-Fetch-Site` CSRF 게이트 | `internal/web/app.go` | **재사용하지 않음** — Claude Code는 그 헤더를 보내지 않고 `/v1/messages`는 POST 전용 |
| SSE Hub | `internal/web/events.go` | **부적합** — 계약상 "본문에 데이터를 싣지 않는" 신호 전용. upstream 토큰 스트림 중계 선례는 저장소에 없다 |
| GLM HTTP 클라이언트 (`glmHTTPDoer` 시험 seam, `x-api-key` + `anthropic-version: 2023-06-01`) | `internal/cli/mcp_glm.go` | 형태 차용 |
| Anthropic HTTP 클라이언트 (`Authorization: Bearer <oauth>`) | `internal/statusline/usage.go` | 형태 차용 |
| GLM credential 저장 (단일 writer, mode 0600) | `internal/glmcred` | GLM `CredentialRef` 구체 타입이 읽기로 재사용 |
| launch 사실 운반 env (`EnvMoaiKanbanBackend`) | `internal/config/envkeys.go` | 원칙 차용 — `MOAI_LAUNCH_PROVIDER` |
| OpenAI HTTP 클라이언트 | — | **없음.** 이 다리는 선례 0이다 |
| TLS/CA/hosts 가로채기 | 상류 `ccmproxy` | 차용하지 않음. 자식 전용 base URL로 연결한다 |

## 9. Codex 인증 소유권 — 0.11.0

구독 경로는 account/read와 managed login API의 상태만 소비한다. MoAI가 Codex auth.json의 토큰을
읽거나 refresh하지 않으며 별도 MoAI OAuth client를 만들지 않는다. browser/device-code는 같은 구독 모드의
두 로그인 방식이다. API 키 모드는 별도 명시 프로필로 관리하여 구독 상태나 과금을 덮어쓰지 않는다.
기존 audit RPC의 credential 점검 경로는 이 기능의 인증 구현으로 재사용하지 않는다.


## 10. 채택하지 않은 대안

| 대안 | 기각 사유 |
|---|---|
| launcher 프로세스 안 goroutine listener | POSIX `syscall.Exec`가 프로세스를 치환한다. 애초에 불가능 |
| `syscall.Exec`를 spawn-and-wait으로 바꿔 부모를 살림 | `MOAI_SESSION_PID` 각인, job control, 종료 코드 전파가 전부 exec 위에 올라타 있다. 바꾸면 세 가지를 동시에 다시 증명해야 한다 |
| 상시 daemon | 세션 수명을 넘는 상태를 만들고 다중 세션 격리 문제를 새로 만든다. 필요 없다 |
| `/v1/models` 자동 검색에 picker 구성을 맡김 | 자동 검색이 picker에 반영된다는 근거가 없다. 프로브 1은 기동 시 `GET /v1/models`를 관측했고 프로브 2는 관측하지 않았으며, 인수 없는 `/model` picker를 연 프로브 2의 항목에는 슬롯·사용자 지정 항목과 Claude 기본 목록만 있었다(`research.md` §15 F6·F7). 반영 여부가 미측정이므로 picker 구성을 여기에 기댈 수 없다. t649는 modelPicker 자식 설정을 채택하며 자동 검색에 제공자 목록을 맡기지 않는다 |
| 모델 ID 접두사로 provider 추측 | 설계 보고서가 명시 금지. exact match만 쓴다 |
| provider signal을 base URL에서 계속 추론 | gateway 아래에서 모든 base URL이 같다. 기존 `EnvMoaiKanbanBackend` 주석이 이미 이 추론을 "a guess dressed as a measurement"라 부른다 |
| `MOAI_KANBAN_BACKEND` 재사용 | kanban·factory 진입 경로에서만 설정된다. 모든 세션의 훅이 읽을 근거가 되지 못한다 |
| 요청별 provider를 세션 상태 파일로 훅에 노출 | 훅에 필요한 것은 launch 시점 사실이다. 파일 쓰기 경합과 수명 관리라는 비용만 늘린다 |
| stale GLM 키가 있으면 launch를 거부 | 사용자가 할 수 있는 조치가 "파일을 손으로 고쳐라"뿐이고, `moai cc`가 오늘 이미 비슷한 정리를 조용히 수행한다. 그 정리를 세 launcher로 넓히고 빠진 키를 더하는 편이 기존 동작과 일관된다 |
| stale 키 정리를 SessionEnd 훅에만 맡김 | 세션이 비정상 종료하면 훅이 돌지 않는다. 다음 launch가 exec 전에 치우는 것이 확실한 지점이다 |
| passthrough 게이트를 설정 키로 표현 | 사용자가 켤 수 있는 스위치는 측정 전 활성화를 막지 못한다 |
| `internal/tmux`의 두 GLM 판정을 이 SPEC에서 이관 | 프로덕션에서 도달하지 않고, 토큰 판정 쪽은 다른 SPEC의 시험이 고정한 동작이며, 같은 파일을 형제 SPEC이 철거 대상으로 다룬다 |
| `BackendGPT`를 감사 backend 집합에도 추가 | 감사 backend는 다른 개념이다. 추가하면 수렴 엔진이 존재하지 않는 backend를 기다린다 |
| GPT PKCE 구현을 이 SPEC에 흡수 | 3분할 Epic 경계를 무너뜨린다. 인터페이스만 고정하고 구체 타입은 형제 SPEC에 둔다 |

## 11. AS-5 launcher 생산 배선 (0.13.0, t654)

착수 시점 코드 상태의 측정 근거는 `research.md` §20(base `d416f8162`)이다.

### 11.1 launch 조립의 생산 경로

세 launcher의 진입 함수는 공유되지 않는다 — `glm`은 `unifiedLaunch(profileName, "glm", args)`
직행(`internal/cli/glm.go:314`)이고 `cc`·`gpt`만 `runClaudeEntry`를 경유한다
(`internal/cli/cc.go:126`, `internal/cli/gpt.go:56`). 공유되는 것은 그 뒤의 launch 조립
(assembly — provider 전용 snapshot·picker 구성·슬롯·`MOAI_LAUNCH_PROVIDER`, §3.2·§6.1)이지 진입
함수가 아니다. 남은 것은 결합이다: provider 전용 세션 snapshot과 picker 구성(§6.7), 인증 방식(구독/API) 표시,
App Server transport가 하나의 launch 조립에서 함께 연결되어야 한다. `internal/cli/gpt.go`의
launch 경로와 공통 launch 경로의 `mode == "gpt" && binding == nil` 분기
(`internal/cli/launcher.go:142`)가 "GPT gateway launch is awaiting transport verification"
대기 오류를 돌려주는 게이트로 살아 있다. 이 게이트는 코드 상수나 설정 플래그가 아니라 **AC 증거로
판정한다** — AS-014~AS-022 전수가 PASS한 트리에서 대기 오류가 제거되고, 통과 전 트리는 대기 오류를
유지해 `AC-MG-026` (a)의 대조군이 된다. 어느 하나의 결합이 빠진 채 다른 하나만 열리는 부분 개방은
`REQ-MG-027` 위반이다.

### 11.2 compaction `appliedEpoch`의 생산 판독

현행 `receipt.NewRebaseLedger(scope, appliedEpoch)`의 비-테스트 호출자는 0건이다(`research.md` §20)
— 재시작 복원이 호출자 추측에 맡겨진 t653 잔여 위험 그대로다. 생산 판독 위치는 이렇게 정한다:

- **쓰기** — rebase가 성공해 원장이 전진하는 지점(`Store.Rebase`)에서 마지막으로 수용한
  `CompactBase.Epoch`를 대화 scope의 영구 receipt 상태에 같은 트랜잭션으로 기록한다. 원장 전진과
  기록이 어긋나면 복원값이 틀리므로 원자성이 계약이다.
- **읽기** — 어댑터가 세션을 다시 열 때 그 값을 읽어 `NewRebaseLedger`에 주입한다. 기록 없음은 새
  scope로 0이다. 판독 불가·훼손은 명시 오류다 — `AC-MG-009`의 실패 의미론(합성 성공 금지)과 같은
  방향이다.
- **정확값 원칙** — 복원값을 후하게(높게) 읽으면 이후 `base.Epoch`가 모두 stale로 거절되고,
  낮게 읽으면 같은 요약을 두 번 rebase한다. 어느 쪽도 조용한 방향이 아니므로 정확값 원칙이며,
  판독 실패 시 근사값으로 때우는 폴백은 두지 않는다.

### 11.3 fork 자식 inherited prefix의 원장 대조

engine은 caller가 주장한 inherited prefix를 그대로 믿는다(t653 잔여 위험). gateway 계층이
`--fork-session` 자식 배리어를 수용하기 전에 `Manifest.ChainTo(경계 Digest)`가 돌려주는 완료
체인과 자식이 주장하는 prefix를 대조하고, 일치할 때만 수용한다. 불일치·변조·미지 원본의 거절 지점은
자식 상태 생성 **전**이어야 한다 — 생성 뒤 거절은 이미 격리 상태로 존재하는 자식을 남기므로
`REQ-MG-015`의 실행 전 거절을 벗어난다. 이 대조는 `AC-MG-009`/AS-013의 명시 분기 계약이 gateway
계층에서 맺어지는 지점이다.

### 11.4 context 경로 표시와 capability 재판정

`internal/cli/gateway_product_binding.go`는 전 provider에 공통
`Capabilities{ContextTokens: 1000000, Images: true, Tools: true, Streaming: true}`를 선언하고
Anthropic 항목에 한해 `RouteID = id + "[1m]"` 변형을 추가한다. 이는 소스 선언이지 수용 보장이
아니다. t654는 provider별 양성·음성으로 재판정한다 — GLM은 text-only(이미지 입력 명시 거절),
Claude는 이미지 수용. UI 표시는 모델 명목 창·현재 경로 유효 한도·누적 사용량의 세 값을 구분하며
(§4 "context와 기능 표시"의 집행), 미검증 수치(1M·921k·872k)를 수용 보장으로 표시하지 않는다.

### 11.5 Windows GitHub CI 실행 증거

기본 경로는 `release-pr-multi-os.yml`의 `workflow_dispatch`다 — windows-latest 레그가
`-tags=integration` 없이 `./...`를 실행하고 `test-stream-release-verify-windows-latest`
아티팩트를 올린다. 판독은 결정 4와 `AC-MG-006`의 절차를 준용한다: 이름을 정한 시험별
`"Action":"pass"`, skip·부재·아티팩트 부재는 PASS가 아니다. 카드·develop CI(`ci.yml`)의 Go test
job은 ubuntu 전용이며 windows Go-test 레그는 release 시점으로 옮겨져 있다(`research.md` §20).
ci.yml에 상시 windows 레그를 추가하는 안은 그 비용이 모든 변경에 붙으므로 기본 채택하지 않고
운영자 결정 사항으로 남긴다. cross-compile exit 0는 이 마일스톤의 PASS 근거가 아니다.

### 11.6 배포 게이트와 CHANGELOG

배포 게이트(`AC-MG-026` (d))의 절차 근거: git-flow 레인 프로토콜 §9 rc 런북(clean 재설치
`rm -f` + `cp`, 맨손 `go install ./cmd/moai` 금지 — LDFLAGS 누락)과 `CLAUDE.local.md` §11
(exit 137 전례, `strings | grep <SHA>` binary lag 검증). rc 번호는
`.moai/docs/version-management.md` Local RC Numbering의 다음 미사용 번호다 — 카드 문구의 "rc.8"은
2026-09-12 발행 시점 표기이며, 발행 시점에 이미 소비됐으면 다음 번호를 쓰고 그 사실을 배포 판정
보고서에 명시한다. CHANGELOG는 t653 판정(`.moai/reports/t653/verdict.md` § CHANGELOG 결정)을
계승해 이 카드에서 사용자 가시 표면 기준으로 발행 검토하며, 발행 여부와 근거(사전-발행 grep 포함)를
같은 보고서에 남긴다. push·PR·병합·워크트리 제거는 게이트 밖이다(§G).

공식 기준 문서(구독 인증·출력 정책): https://learn.chatgpt.com/docs/app-server (§4.4와 같은 출처,
접근 2026-09-12).

## 12. 0.16.0 최종 설계 — 직접 모델 계약과 Factory 단일화

이 절은 §7의 옛 Kanban/Factory 병행 설계와 §6.7의 GPT 슬롯 일반화를 대체한다. §7은 결정 당시의
감사 이력으로만 읽는다.

### 12.1 모델과 effort의 두 독립 축

| Claude 역할 별칭 | GPT 모델 ID | 의미 |
|---|---|---|
| `fable` | `gpt-6-astra` | 모델 직접 선택 |
| `opus` | `gpt-5.6-sol` | 모델 직접 선택이자 무명시 기본값 |
| `sonnet` | `gpt-5.6-terra` | 모델 직접 선택 |
| `haiku` | `gpt-5.6-luna` | 모델 직접 선택 |

MoAI는 이 표 위에 별도 모델 tier를 만들지 않는다. `max`·`high`·`medium`·`low`·`ultra`는
App Server에 전달할 effort 축이며 모델 ID와 직교한다. launch 조립은 먼저 상속 슬롯을 제거하고 표의
네 값을 명시적으로 넣는다. 그 뒤 명시 effort가 있으면 원값을 별도 필드로 운반하고, 없으면 누락으로
둔다. 지원되지 않는 조합은 모델이나 effort를 바꾸지 않고 명시 오류로 끝낸다.

### 12.2 Messages ingress와 App Server 소유권

Claude Code는 일반 Anthropic Messages `/v1/messages` 입구와 `tool_use`/`tool_result`, 승인, hook,
실제 도구 실행을 소유한다. GPT route는 공식 Codex App Server를 `initialize`한 뒤 thread/turn을 소유하게
한다. production entry `productionGatewayHandlerFactory`는 `internal/cli/gateway_product_binding.go`에,
composition helper `newGatewayHandlerFactory`는 `internal/cli/gateway_factory.go`에 둔다. 조립은 private
`MOAI_HOME`의 유효 auth profile과 명시 Codex binary를 입력으로
`codexapp.Client` initialize → `AppServerAuthority` → durable `codexbridge.Engine` →
`AppServerAdapter` 순서로 만든다. 부분 실패 시 역순 정리하고 handler `Close`가 engine/client를 정확히
한 번 닫는다. GPT managed route에 `OpenAIAdapter`, 구독 token 직접 읽기/refresh, private endpoint 또는
API 과금 fallback은 남지 않는다.

adapter는 App Server의 dynamic tool JSON-RPC 요청마다 raw RPC ID와 thread/turn/call ID를 대화별 ordered
pending batch에 기록하고, 한 turn에서 barrier 전에 도착한 모든 call을 첫 HTTP의 복수 `tool_use`로 반환한다.
후속 HTTP의 `tool_result` 순서는 call 발생 순서와 달라도 된다. 예를 들어 Claude가 `call-b`, `call-a`
순으로 결과를 보내면 원장은 각각 저장한 `rpc-b`, `rpc-a`에 정확히 한 번 응답한다. 일부 결과는 남은
pending을 보존하고, 중복·foreign·cross-thread·이미 완료된 ID·schema 변경은 다른 call을 소비하거나
App Server에 응답하기 전에 거절한다. 모든 pending이 해결된 뒤에만 같은 turn의 최종 text를 이어서
Claude 화면에 반환한다. HTTP 수명과 App Server turn/RPC 수명은 분리한다.

Codex native shell·file·MCP·agent·hook은 이 경로에서 비활성화한다. tool 실행을 App Server와 Claude
Code 양쪽에 허용하면 같은 외부 효과가 두 번 발생하고 Claude approval/hook 경계를 우회할 수 있기 때문이다.

이 결합은 Anthropic이 비-Claude 모델용으로 공식 지원하는 Claude Code 경로가 아니다. 따라서 공식 지원,
약관상 무위험, 기능 동등성을 제품 문구로 약속하지 않는다. OpenAI App Server의 문서화된 인증·프로토콜을
사용하는 것과 전체 조합의 지원 여부는 별개다.

### 12.3 Factory와 이름 사전

활성 실행 모드는 Factory 하나다. `-f`가 Factory lead/worker를 선택하며 `-k`/`--kanban`은 부작용 없이
거절하고 `-f`를 안내한다. 활성 이름은 다음 하나씩만 사용한다.

| 목적 | 활성 이름 |
|---|---|
| 실행 모드 | Factory |
| 작업 큐 | Todo |
| UI 상태 | Tasks |
| 배차·실행 기록 | Dispatch |
| 공통 내부 조정 | Orchestration |

따라서 새 패키지/import는 `internal/orchestration`, GPT 상수는 `orchestration.BackendGPT`이며 저장 값
`"gpt"`는 호환을 위해 유지한다. 과거 SPEC·commit·report ID와 명시적 migration/legacy reader만 옛
명칭을 가질 수 있다. 저장 필드의 물리적 migration이 이 구현에서 안전하게 끝나지 않으면 legacy reader를
유지하고 별도 SPEC으로 분리하되, 활성 CLI와 일반 런타임 식별자의 이관은 미루지 않는다.

구 환경 이름을 읽는 경계는 `research.md` §22.5의 두 legacy reader와 시험으로만 제한한다. reader는 구
이름을 새 Dispatch 구조로 변환할 뿐 구 이름을 다시 쓰지 않는다. 일반 runtime은 `MOAI_DISPATCH_*` 이름을
사용하고 사용자 출력은 Todo/Tasks/Dispatch/Orchestration만 사용한다. allowlist 판정은 경로만 보지 않고
허용 심볼과 read-only 사용까지 AST 또는 소스 가드로 대조한다.

### 12.4 history 400 귀속 상태기계

대화별 상태는 `last completed public prefix`, receipt chain digest, manifest generation, App Server thread와
마지막 완료 turn을 함께 고정한다. transport retry는 동일 요청 idempotency key와 동일 digest일 때 기존
진행/결과에 attach하고 새 `turn/start`를 만들지 않는다. 새 사용자 입력은 완료 prefix 바로 뒤에 정확히
한 번만 추가한다.

다음은 서로 다른 거절 원인이다.

- 공개 history bytes가 바뀜: `history_changed`
- gateway가 발행·인증하지 않은 `agent_summary`: `agent_summary_untrusted`
- request prefix와 receipt/manifest chain 또는 generation 불일치: `receipt_manifest_mismatch`
- resume barrier 뒤 이미 반영한 입력 재제출: `resume_duplicate_input`

네 거절은 App Server 호출 전에 끝나며 production `gateway.Server` error encoder가 HTTP 400,
Anthropic envelope `type=invalid_request_error`, 위 exact cause를 함께 쓴다. 오류 타입이 구현하는
`CauseCode()`도 같은 SSOT token을 반환한다. `ServerConfig.RejectionLogger`에는
`RecordGatewayRejection(fields map[string]string)` recorder를 주입한다. server가 소유하는 recorder는 정확히
`cause`, `route`, 길이 64의 비식별 `digest` 세 필드만 받으며 adapter의 자유형 오류 문자열을 기록하지 않는다.
prompt, summary, tool result, reasoning, token, credential, canary는 key나 value로 싣지 않는다. 단위 fixture는
실제 `gateway.Server`와 error encoder를 지나 envelope, 생성/RPC/upstream HTTP 0, recorder의 safe/forbidden
필드를 검증한다. recorder seam이 없는 현재 상태는 제품 RED이며 M14-R5는 실제 sink record를 따로 보존한다.
`acceptance.md` AC-MG-009의 양성 행이 실제 계수를 움직이는 대조군이고, 음성 행 0회만으로 성공을 주장할 수 없다.

### 12.5 test-first seam·platform·manifest 구현 경계 (M14-R0.2)

`plan.md` M14-R0은 소스 literal grep이 아니라 실제 fake Codex App Server subprocess와 HTTP 두
message, production catalog wiring, temp Todo/Tasks digest, BacklogStore Dispatch/Factory runtime,
launch/prompt, durable engine과 receipt/history HTTP, 실제 tracked Git source set을 계수한다. 명명
시험과 case 수는 `acceptance.md` §D가 소유한다. RED-only 단계에서는 제품 파일을 수정하지 않고
이미 충족된 회귀 가드와 의도한 RED assertion을 구분한다. iteration-5 App Server component는 정상
initialize/thread/turn과 duplicate/foreign/cross-thread 5개가 회귀 가드이고, 두 call 첫 batch·역순
continuation·exact `rpc-a/rpc-b` 3개가 실제 RED다. portable process 6개와 lifecycle 정상 5개는 회귀
가드다. alias env 15, effort RPC 20, production wiring 6, Factory unit boundary 7,
history cause/logger 5, naming schema/current 2는 실제 RED다. `claude-tool-result-continuation`은 Claude가 실행한 tool result를 동일 call로 이어 주는
계약이며 approval/hook 자체를 이 component 시험이 실증했다고 표현하지 않는다. Factory 단위 시험도
`-p`·model·effort·env 전달과 폐기 flag 무부작용만 판정하며 실제 Agent/tool/Tasks/Dispatch는
`acceptance.md`의 AS-017·018 live gate가 소유한다.

M14-R0.2는 production wiring PATH fake `codex`를 protocol-speaking subprocess로 유지하고 독립 self-probe
뒤 production log를 초기화한다. clean close는 EOF/finally marker가 아니라 성공한 첫
Start→Initialize→Account→Close 뒤 같은 private profile/lease에서 두 번째 Start→Initialize→Close가 성공하는
것으로 판정한다. 현재 production selector는 네 `AuthPKCE`, concrete `*gateway.OpenAIAdapter`, absent/poisoned
legacy auth store를 건드리는 조립까지 6 RED다. product GREEN은 managed App Server 조립에서 legacy store
open/read/refresh 0을 함께 증명해야 한다.

프로세스 구현은 플랫폼 파일로 분리한다. POSIX 파일은 `//go:build !windows`와
`syscall.Exec`를 갖고, Windows companion은 `//go:build windows`와 독립 spawn-and-wait를 갖는다.
`TestGatewayProcessContractPortable`은 OS API를 source token으로 검사하지 않고 fake child의 readiness,
wait, cancellation, exit code, flush/atomic replace, permission 결과를 공통 interface 뒤에서 검증한다.
현재 시험은 실제 `gateway.StartChild`/`RunChildWithControl` 경로에서 readiness와 HTTP/overlay,
0600 permission, cancellation, wait, exit 23을 6/6 PASS했다. 이전의 분리된 지역 변수 기반 wait/exit
RED는 제품 경로를 측정하지 않은 무효 증거라 폐기한다.
release CI Windows live 행은 이 portable 시험을 대체하지 않는 별도 완료 gate다.

naming checker는 `research.md` §22.6의 정확한 JSON 경로와 schema를 읽고, migration 이후 tracked Go
source를 네 category로 재분류한 뒤 manifest의 post-migration canonical current set과 양방향 차집합을
구한다. baseline의 옛 `internal/kanban/**` 경로는 provenance/mapping 입력일 뿐 current 배열에 남기지
않는다. `missing_from_tree`,
`extra_in_tree`, `schema_error`는 서로 다른 실패 축이며 어느 하나도 자동으로 manifest를
수정하지 않는다. manifest 생성·갱신은 M14-R3 구현 owner의 소유다.
