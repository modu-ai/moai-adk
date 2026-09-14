# SPEC-MOAI-GATEWAY-001 — 구현 계획


> t649 / 0.10.0의 현재 계약: 제공자별 선택·요청 경계, 시작·Default·네 슬롯·fallback·저장·재개 격리는
> 코어가 소유한다. 0.6.0~0.9.0의 교차 제공자 전환과 PICKER 이관·출시 대기 결정 중 이 범위는 대체되었다.
> 아래 과거 결정·프로브 기록은 당시 근거로 보존하며 현재 판정은 spec.md REQ-MG-019와 acceptance.md t649 보강을 따른다.

## 0.11.0 현재 실행 계획

운영자의 App Server 전면 재설계 지시와 HTML 보고 뒤 발급된 t650~t654를 순서대로 실행한다.
기존 M0~M9 중 direct GPT backend·자체 구독 token/opaque 처리 부분은 아래 현재 계획으로 대체한다.
Claude/GLM 및 제공자별 picker·인증·보안·Windows 회귀는 유지한다. 과거 확정 결정은 당시 이력이다.

| 카드 | 현재 범위 | 선행 | 완료 증거 |
|---|---|---|---|
| t650 | 기존 RPC 재사용, 설치 capability, managed browser/device 및 API auth | t649 보고 | 프로토콜·인증·native 실행 차단 음성 |
| t651 | 지연 schema 등록 전략 비교, Claude dynamic tool 왕복, ToolSearch | t650 | 실제 도구·tool_reference·schema 대조 |
| t652 | HTTP 간 turn/pending RPC, 결과 단일 소비, crash·취소 | t651 | 중복·foreign·강제 종료 음성/복구 양성 |
| t653 | 소유 thread resume, model 변경, fork, compaction | t652 | 실제 Claude 경로의 정합성 |
| t654 | 세 picker·auth/context 표시·실제 구독/API·Windows 회귀 | t653 | PTY와 provider 왕복·GitHub Windows 실행 |

기본 ToolSearch의 최초 전체 schema 전제는 Claude2.1.269 픽스처로 반증되었다. 운영자가 hybrid를 승인했으므로
초기 정의는 native schema, 검색 후 발견된 정의는 dispatcher로 구현한다. ToolSearch는 유지한다.
이름·schema·대화 귀속·인수 검증과 등록 시 thread 수명 불변을 t651의 여섯 음성 시험으로 고정한다.
Codex native 도구 실행 0은 설정 파일 존재가 아니라 부정 실행 시험으로 판정한다.
thread/resume.history·raw reasoning 주입은 정상 경로로 채택하지 않는다. HTTP가 끝나도 pending RPC는 살아 있어야 하며
프로세스 재시작 후 옛 RPC ID 재생이나 불명확한 도구 재실행은 금지한다.

t653은 계정·family·agent·thread와 완료 public prefix 원장을 고정하고 resume 및 idle 모델 변경을 연결한다.
Claude 압축은 정상 요약 turn으로 처리하고 반환 summary와 인증된 PostCompact를 대조한 뒤 public history를
재설정한다. PreCompact 다음 HTTP 추측이나 추가 명시 compact RPC는 사용하지 않는다.
일반 Agent의 독립 child context와 launcher의 --fork-session을 구별한다. 명시 세션 분기는 원본 family 및
exact prefix의 completedTurnID로 공식 lastTurnId 분기를 수행한다. native Agent(fork)/subtask는 별도 연결 실증
대상이며 현재 부모 귀속 추정으로 지원하지 않는다. 실제 회상·재개·병렬 자식·명시 분기 양성 조건을 유지한다.
t653에서 먼저 설치 Claude의 native Agent(fork)/subtask 가용성을 실제 호출 표면으로 확인하고 결과를 기록한다.
기능을 찾지 못하면 전제 실패 NOT-RUN으로 남기되 기존 native fork 양성 의무를 삭제하지 않는다.
가용할 때 부모 이력 상속·정확한 분기 위치·병렬/중첩 격리·재개 양성과 foreign/변조 귀속 음성을 검증한다.
이 증거 전에는 AS4 및 전체 지원 완료를 보류하며 일반 자식·--fork-session 성공으로 대신하지 않는다.
순서는 가용성 확인과 독립적으로 진행 가능한 resume·모델·압축·일반 자식·명시 세션 분기 구현을 병행하고,
마지막 완료 판정에서 native fork 증거를 포함해 AS-013 전체를 확인하는 것이다.
후속 운영자 승인 “API도 App Server 출력 정책 사용”에 따라 구독/API 모두 서버 출력 정책을 채택한다.
MoAI 바이트·취소 제한을 Claude 생성 토큰 상한과 동일하게 표시하지 않고 선택 인증 방식과 API 별도 과금을 유지한다.

현재 미커밋 gateway 의존을 새 원격 기준 worktree에 있다고 가정하지 않는다. 오케스트레이터가 재현 가능한 의존
snapshot·소유권을 확인한 격리 트리를 정하고 카드별 근거를 남긴다. 추적 하위 카드를 독립 통합 완료로 표시하지 않는다.

## 0.13.0 AS-5 실행 계획 (t654)

카드 t654의 범위: 세 launcher 통합, context 실증, 로컬 배포 판정. 선행 t653 완료. 바인딩 조건 —
API 과금 자동전환 금지, 구독은 공식 App Server managed auth 사용(직접 구독 backend 호출·토큰 읽기의
실행 경로 대체), 구독/API 모두 공식 App Server 출력 정책(운영자 추가 승인, MoAI 바이트·취소 제한 유지,
Claude max_tokens 동일 생성 상한 보장 표기 금지), 완료 주장은 카드별 실제 명령·출력·기준 SHA·Gaps로 증명,
push·PR·병합·워크트리 제거는 별도 지시. CHANGELOG는 t653 판정(`.moai/reports/t653/verdict.md` § CHANGELOG
결정)을 계승해 이 카드에서 사용자 가시 표면 기준으로 발행 검토한다. 공식 기준 문서:
https://learn.chatgpt.com/docs/app-server . 증거는 전부 `.moai/reports/t654/` 아래 `as5-` 접두사로
남긴다 — 그 디렉터리에 옛 t654 가드 보고서(`verdict.md`, `probe-outputs.txt`)가 이미 있어 충돌을 피한다.

기준선(착수 시 재측정): worktree `.claude/worktrees/t654`, branch `WT-gateway-launchers`, base
`d416f8162`(로컬 develop head). 착지 시점 코드 상태는 `research.md` §20 — `moai gpt` 대기 오류 경로,
`NewRebaseLedger` 비-테스트 호출자 0건, `gateway_product_binding.go` 공통 capability 선언, ci.yml의
windows Go-test 레그 부재.

### 마일스톤 (통합 → 검증 → 배포 게이트 순)

| 마일스톤 | 우선순위 | 범위 | 대응 AC | 증거 |
|---|---|---|---|---|
| A5-M1 | High | launcher 생산 통합 | AC-MG-026 (a) | as5-launcher-integration.md |
| A5-M2 | High | compaction epoch 생산 복원 + fork prefix 원장 대조 | AC-MG-026 (b)(c) | as5-epoch-fork-wiring.md |
| A5-M3 | High | 경로별 context 재판정·표시 | AS-020 (AC-MG-010) | as5-context-paths.md |
| A5-M4 | High | 구독/API 이중 경로 검증 | AS-021 (AC-MG-020), AS-014 인증 표시 | as5-auth-modes.md |
| A5-M5 | High | 실제 PTY 실증 스윕 | AS-014·AS-017·AS-018·AS-019 | as5-pty-*.log |
| A5-M6 | Medium | Windows GitHub CI 실행 증거 | AS-022 (AC-MG-006) | as5-windows-ci-verdict.md |
| A5-M7 | High (종결) | rc 로컬 배포 게이트 + CHANGELOG 발행 검토 | AC-MG-026 (d) | as5-deploy-verdict.md |

**A5-M1 — launcher 생산 통합 (우선순위: High).** 세 launcher가 provider 전용 catalog·picker
구성(`REQ-MG-019`)·인증 방식 표시·App Server transport를 하나의 launch 조립으로 결합한다.
예상 변경: `internal/cli/gpt.go`(대기 오류 경로의 게이트화 — 검증 AC 통과 전 유지, 통과 후 제거),
`internal/cli/launcher.go`(같은 대기 오류 리터럴의 두 번째 위치 — 공통 launch 경로의
`mode == "gpt" && binding == nil` 분기, `launcher.go:142`. `AC-MG-026` (a)의 리터럴 0건 판정은
`internal/cli` 전체를 보므로 이 위치의 게이트화가 빠지면 첫 GREEN이 낙오된다),
`internal/cli/gateway_launcher.go`·`gateway_prepare.go`·`gateway_session.go`(조립 결합),
`internal/cli/gateway_product_binding.go`(provider별 auth 표시 전달). RED→GREEN: 대기 오류 리터럴의
비-테스트 존재 판정 시험을 먼저 적색으로 세우고(게이트 통과 트리 기준), 제거로 녹색 만든다. 게이트
통과 전 트리에서는 같은 시험이 대조군(대기 오류 유지)으로 남는다.

**A5-M2 — epoch 생산 복원 + fork prefix 원장 대조 (우선순위: High).** t653 잔여 위험 두 가지를
닫는다. (b) rebase 성공 지점(`Store.Rebase`)에서 마지막 적용 epoch를 대화 scope 영구 상태에
기록하고, 세션 재시작 시 그 값을 판독해 `NewRebaseLedger`에 주입한다 — 기록 없음은 0, 판독 불가·훼손은
명시 오류(높은 값의 stale 오판 방향을 피하기 위한 정확값 원칙). (c) `--fork-session` 자식 배리어
수용 전 gateway 계층이 `Manifest.ChainTo(경계)` 완료 체인과 자식 prefix를 대조하고, 불일치·변조·미지
원본은 자식 상태 생성 전 명시 거절한다. 예상 변경: `internal/gateway/`(어댑터 배선),
`internal/codexbridge/`(호출 경계), `internal/gateway/receipt/`(판독 보조). RED→GREEN:
`TestRebaseLedgerRestoration*`(고정값 주입 현행 구현이 적색), `TestForkPrefixCrossCheck*`
(caller-asserted 수용이 변조 변형에서 적색).

**A5-M3 — 경로별 context 재판정·표시 (우선순위: High).** `gateway_product_binding.go`의 공통
`Capabilities{ContextTokens: 1000000, Images: true}`를 provider별 양성·음성으로 재판정한다 — GLM은
text-only(이미지 입력 명시 거절), Claude는 이미지 수용. UI 표시는 모델 명목 창·현재 경로 유효 한도·
누적 사용량의 세 값을 구분하고, 미검증 수치(1M·921k·872k)를 수용 보장으로 표시하지 않는다
(`plan.md` "Native 정책 구현 인계" 절의 AS5 조항 집행). RED→GREEN: provider별 capability 시험.

**A5-M4 — 구독/API 이중 경로 검증 (우선순위: High).** Given 구독 managed 계정과 API 키 프로필일 때,
두 모드를 각각 실제 선택·실행하고 표시된 인증 방식이 실제 선택과 일치함을 확인한다. 음성: 구독 실패·
만료 유도 시 API 과금 경로 요청 계수 0(자동전환 금지), MoAI의 토큰 파일 접근 0. 구독 인증의 근거는
공식 App Server managed auth다.

**A5-M5 — 실제 PTY 실증 스윕 (우선순위: High).** 실제 launcher PTY에서 도구검색(hybrid ToolSearch
제품 실행), 서브에이전트(네 슬롯 별칭의 제공자 경계 해석), 재개·모델전환(Claude/GLM launcher 측 —
AS-010·011·012의 GPT thread 실세션 양성은 T21대로 t844 소관이며 이 카드가 흡수하지 않는다),
picker 제공자 경계(AS-014 재판정)를 각각 수행한다. 실제 계정 권한이 필요한 항목은 이 세션 환경에서
불가할 수 있다 — 그 경우 Gap으로 기록하고 운영자 판정(라이브 계측 창)을 요청한다. 임시 Claude 직접
실험으로 제품 판정을 대체하지 않는다.

**A5-M6 — Windows GitHub CI 실행 증거 (우선순위: Medium).** 기본 경로: `release-pr-multi-os.yml`의
`workflow_dispatch` 실행 → windows-latest 레그가 `-tags=integration` 없이 `./...`를 실행 →
`test-stream-release-verify-windows-latest` 아티팩트 판독(`AC-MG-006`의 판독 절차 준용 — 이름을 정한
시험별 `"Action":"pass"`, skip·부재는 PASS 아님). ci.yml에 상시 windows 레그를 추가하는 안은 카드·
develop CI 시간 비용이 매 변경에 붙으므로 기본 채택하지 않고 운영자 결정 사항으로 남긴다. cross-compile
exit 0만으로 이 마일스톤을 PASS로 세지 않는다.

**A5-M7 — rc 로컬 배포 게이트 + CHANGELOG 발행 검토 (우선순위: High, 종결).** 전제: A5-M1~M6와
AS-014~AS-022의 판정이 PASS(또는 근거 갖춘 Gap — 단 배포 게이트 자체의 전제는 검증 PASS)다. 절차는
`AC-MG-026` (d)의 명령 형태를 따른다: `make build VERSION=v<다음 미사용 rc>` → `rm -f ~/go/bin/moai
&& cp bin/moai ~/go/bin/moai`(clean 재설치 — 생략 시 exit 137 전례) → `~/go/bin/moai version` exit 0 →
`strings ~/go/bin/moai | grep <기준 SHA>` binary lag 검증. rc 번호는 `.moai/docs/version-management.md`
Local RC Numbering의 다음 미사용 번호다 — 카드 문구의 "rc.8"은 2026-09-12 발행 시점 표기이며 발행 시점에
이미 소비됐으면 다음 번호를 쓴다(이 차이는 조용히 흡수하지 않고 배포 판정 보고서에 명시한다). 같은 보고서에
CHANGELOG 발행 검토 결과(사용자 가시 표면 기준, B12 사전-발행 grep `grep -c 'SPEC-MOAI-GATEWAY-001'
CHANGELOG.md` 포함)를 남긴다. push·PR·병합·워크트리 제거는 없다.

## A. Context

`spec.md` §A가 배경이고 `research.md`가 근거다. 이 문서는 그 위에서 무엇을 어떤
순서로 만들지를 정한다. 시간 추정은 쓰지 않는다 — 우선순위(High/Medium/Low)와 단계
순서로만 표기한다.

기준선: 작업 트리 `.claude/worktrees/moai-proxy-unified`(세션 고정 경로라 디렉터리 이름은
그대로다), 브랜치 `WT-unified-gateway`, HEAD `ed71054d3`(0.6.0 재기준. 0.5.0까지는 `d060e0d13`). 착수 시점에는 §C에 따라 다시 잰다.

명칭 대응: 설계 원문·핸드오프·iter1 감사는 "proxy" / `REQ-MP` / `AC-MP`, 이 문서는
"gateway" / `REQ-MG` / `AC-MG`를 쓴다(`spec.md` §0).

**Epic 구성.** 이 SPEC은 3분할 Epic의 첫 조각이며, 0.6.0에서 코어의 두 표면을 형제 SPEC 제안 두 건으로 더 떼어 냈다(결정 11).

| SPEC | 범위 | 관계 |
|---|---|---|
| `SPEC-MOAI-GATEWAY-001` (이 문서) | gateway 코어 — CLI 계약(kanban/factory 포함), supervisor, ingress, adapter, registry, launch provider signal, `settings.local.json` 계약, `CredentialRef` 인터페이스, 실패·보안 | 나머지 둘의 기반 |
| `SPEC-MOAI-GPT-AUTH-001` (제안) | GPT PKCE login/logout, MoAI 소유 credential 저장, GPT `CredentialRef` 구체 타입 | 이 SPEC이 고정한 인터페이스 뒤에 구현. 독립 착수 가능 |
| `SPEC-MOAI-CG-RETIRE-001` (제안) | `moai cg` 철거 스윕, `team_mode: cg` 마이그레이션, 문서 4-locale, `internal/tmux`의 `sessionEnvHasGLM`·`hasGLMEnv` 처리 | 이 SPEC과 병행 가능하나 `REQ-MG-021`(launch provider signal)이 선행되면 쉬워진다 |
| `SPEC-MOAI-GATEWAY-PICKER-001` (제안, 0.6.0) | picker·모델 선택 표면 — picker overlay 구성과 "Default" 행 불변식, 초기 모델 전달 방식(명시 `--model`, 빈 기본값, 재개 경로), `/model <id>` 저장 기본값의 안내 | t649에서 이 범위를 코어로 회수. 별도 PICKER 착지를 요구하지 않음 |
| `SPEC-MOAI-GATEWAY-TEAMMATE-001` (제안, 0.6.0) | tmux pane teammate 표면 — tmux 세션 env 소유권, pane teammate의 연결·수명·모델 ID | 이 SPEC의 in-process 한정(결정 12)을 넓히는 쪽이다 |

## B. 알려진 문제와 결정 가능성 순서

아래 마일스톤은 **되돌리기 어려운 결정을 앞에 둔다.** 데이터 모델·타입 인터페이스·
사용자에게 보이는 흐름이 먼저 오고, 기계적 작업은 뒤로 간다. 검토자가 가장 바뀌기 쉬운
결정부터 보게 하려는 배치다.

단, 되돌리기 비용보다 앞서는 것이 하나 있다 — **출시 여부 자체를 가르는 측정**이다. T09 측정은
그 결과에 따라 Anthropic adapter의 존재가 결정되므로 모든 구축 마일스톤보다 먼저 온다(M0).

## C. Pre-flight

착수 전에 확인할 것.

- [x] t649 구현·검증 및 카드 발행은 2026-09-12 운영자의 “모두 검증해서 처리해줘”, “카드 발행해서 진행해줘”로 승인되었다.
- [ ] `git rev-parse --short HEAD`, `git branch --show-current`, `git status --short`를
      착수 시점에 **다시** 측정. 이 문서의 기준선은 작성 시점 값이며 보증이 아니다.
- [ ] 형제 SPEC 네 개(`SPEC-MOAI-GPT-AUTH-001`, `SPEC-MOAI-CG-RETIRE-001`, `SPEC-MOAI-GATEWAY-PICKER-001`,
      `SPEC-MOAI-GATEWAY-TEAMMATE-001`)의 ID가 실제로 발급되었는지 확인(제안 ID는 아직 예약이 아니다).
- [ ] **`internal/tmux/cg_detect.go` 인계 확인.** 이 SPEC은 그 파일을 수정하지 않는다. 오케스트레이터가
      `SPEC-MOAI-CG-RETIRE-001`을 발급할 때 두 가지를 그 SPEC의 입력으로 넘긴다: `sessionEnvHasGLM`·
      `hasGLMEnv`의 처리 책임, 그리고 M0에서 정할 세션 접근 토큰 운반 키(그 키가
      `ANTHROPIC_AUTH_TOKEN`이면 두 판정은 모든 gateway 세션에서 참이 된다).

## D. 제약

- POSIX `syscall.Exec` 보존 (`REQ-MG-005`). `MOAI_SESSION_PID` 각인이 여기 올라타 있다.
- 새 환경변수 이름은 `internal/config/envkeys.go`를 거친다. 맨몸 `ANTHROPIC_*` 리터럴은
  AST 기반 가드 시험 `TestNoBareAnthropicEnvVarLiteralsInProduction`이 잡는다 — 이 가드는
  **빌드가 아니라 변경 패키지 테스트에서 실패**하므로 `go build` 통과를 근거로 삼지 않는다.
- `settings.local.json`의 `env` 값과 프로세스 env의 우선순위는 측정하지 않았다. 구현은 어느 쪽도
  전제하지 않는다(`design.md` §6.4).
- 전체 스위트를 로컬에서 돌리지 않는다. 변경 패키지만 돌리고 전 패키지 판정은 CI에
  맡긴다(`CLAUDE.local.md` §4.1).
- `.moai/specs/SPEC-MOAI-GATEWAY-001/` 밖의 파일은 plan 단계에서 건드리지 않는다.

## E. 자기 검증

각 마일스톤 종료 시 5구획 보고(Claim / Evidence / Baseline-attribution / Gaps /
Residual-risk)를 남긴다. 실행하지 않은 명령의 출력을 근거로 인용하지 않는다. 부재
주장에는 탐색 범위를 명시하고 가능한 경우 동일 범위 양성 대조군을 붙이며, 이 SPEC 자신의
산출물과 `.moai/reports/`가 탐색 범위에 들어가면 **둘 다** 제외하고 그 사실을 적는다.

## F. 마일스톤

### M0 — T09 선행 측정: Claude 구독 OAuth × 로컬 gateway 인증 공존 (우선순위: High)

**run 단계의 첫 항목이다.** Anthropic adapter를 만드는 어떤 마일스톤보다 먼저 수행한다.

- **하네스.** 이 측정은 버리는 spike용 최소 loopback forwarder로 수행하며, M1 이후의 gateway
  코드를 전제하지 않는다. forwarder는 요청을 원래 Anthropic endpoint로 넘기고 헤더와 결과만
  기록한다. 측정이 끝나면 버린다.
- 기존 Claude 구독 로그인 상태에서 로컬 gateway 접근 통제를 함께 두었을 때, OAuth 전달·refresh·
  필수 beta 헤더가 실제 클라이언트에서 유지되는지 측정한다.
- 같은 측정에서 **세션 접근 토큰의 운반 키**(`ANTHROPIC_AUTH_TOKEN` 또는 별도 헤더)를 정한다.
  그 결정은 `design.md` §3.2와 §C Pre-flight의 형제 SPEC 인계 입력이 된다.
- 결과를 5구획 보고로 기록하고 `REQ-MG-016`의 게이트 값으로 삼는다. 게이트는 설정 키가 아니라
  코드와 catalog 항목의 부재로 표현한다(`design.md` §2.2).
- **음성이면 Anthropic passthrough adapter는 출시하지 않는다.** 설계 보고서 §7은 local token
  충돌이 해결되지 않은 구독 모드를 출시하지 않는다고 못 박는다. 이 경우 M6의 Anthropic 경로는
  API gateway 모드만 남고, "기존 로그인 그대로 3사 전환"이라는 문구는 쓸 수 없다.
- 대응 REQ: MG-016 / AC: AC-MG-021 (c)

핸드오프 §7은 T03 mock 재현을 첫 항목으로 두었다. 운영자 결정으로 이 측정이 그 앞에 온다.
두 항목은 서로의 결과에 의존하지 않으며, T03 재현은 첫 **구축** 마일스톤(M1)으로 유지된다.

**0.8.0 후속 시험 조건 (2026-09-11 운영자 지시).** Claude 실서비스 시험은 2026-09-11 19:00 Asia/Seoul 이후에만
재개하며, Opus 5와 Sonnet 5를 사용한다. 정확한 모델 ID와 실제 계정 접근 가능 여부를 실행 전에 확인하고 각각 기록한다.
과거 `claude-sonnet-4-5` 캡처는 당시의 요청 형태 증거로 보존하며 새 모델 시험을 대체하지 않는다. M0의 429는 INCONCLUSIVE로
유지하고 정상 응답·OAuth refresh 확인 전 T09를 PASS로 바꾸거나 AUTH 게이트를 열지 않는다.

**0.9.0 측정 인계.** `research.md` §19의 두 정상 응답 모델과 직접 refresh 상관관계로 M0 선행 게이트의
양성 근거를 확보했다. 운반 키는 design §3.2의 별도 헤더로 확정했다. 이전 429·changed-Bearer-only 보고서를
덮어쓰지 않으며 제품 통합은 별도 검증한다.

### M1 — T03 same-session `/model` 전환을 loopback mock으로 재현 (우선순위: High)

**구축 마일스톤 중 가장 먼저 해야 하는 이유**: 이 SPEC 전체가 "같은 세션에서 `/model`로 고른
모델 ID가 다음 요청의 `model` 필드에 실려 온다"는 전제 위에 서 있다. 설계 보고서 §3의 실측은
비대화형 단발 요청 두 건까지였다. 0.5.0의 클라이언트 프로브는 실제 Claude Code 2.1.267 TUI를 Python loopback mock에
붙여 **한 세션 안의 연속 전환을 관측했다** — 전환 뒤 turn 요청이 새 모델 ID와 이전 대화 기록을 함께 실었다
(`research.md` §15). 그러나 그 mock은 제품이 아니다. Go gateway를 거친 전환은 아직 관측되지 않았고, 이 전제가
gateway에서 틀리면 뒤의 모든 작업이 헛돈다.

- **진입 게이트 — 원본 요청 캡처 (결정 9, iter4 G4-B2).** M1의 첫 작업이며, 끝나기 전에는 검증 요청 인식기와 그 픽스처를
  만들지 않는다.
  - 실제 Claude Code TUI를 loopback 캡처 mock에 붙인다. mock은 요청 본문과 질의 문자열을 **가공 없이** 기록한다. 파생
    필드(`bool` 변환, 개수)만 남기지 않는다.
  - `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`, `DISABLE_TELEMETRY`, `DISABLE_ERROR_REPORTING`, `DISABLE_AUTOUPDATER` 네 억제
    플래그를 **켜지 않고** 실행한다. 0.5.0 프로브는 넷을 모두 켰고, 이 SPEC의 launch 정리는 첫째를 지운다.
  - 네 요청을 캡처한다: `/model <id>` 검증 요청, 첫 turn, 제목 생성, 이후 turn. 같은 조건에서 나타난 그 밖의 비스트리밍
    요청과 `max_tokens`가 작은 요청도 목록으로 남긴다.
  - 캡처는 추적되는 시험 픽스처 `internal/gateway/testdata/validation-capture/claude-code-<버전>/`로 옮긴다. 요청마다 원본
    본문 하나를 `.json` 파일로 두고, 질의 문자열·Claude Code 버전·캡처 조건은 같은 디렉터리의 `README.md`에 적는다. 원본
    요청 본문을 패키지 `testdata/`에 두는 저장소 선례는 `internal/hook/testdata/agent_pretool_payload.json`이다. 옮기기 전에
    본문에 비밀 값이나 기계 로컬 절대 경로가 없는지 확인한다. 비식별 처리가 필요하면 원본은 보존하고, 변경한 필드·처리 이유·
    원본 및 픽스처 SHA-256을 README에 연결한다. 인식 조건의 네 필드는 바꾸지 않는다. 기계 로컬 원본만으로는
    재현 가능한 추적 픽스처 게이트를 충족하지 않는다.
  - 게이트의 대조 결과 — 캡처한 요청별 네 조건 판정, blocker 여부, 실행한 Claude Code 버전 — 는
    `.moai/reports/<card-id>/m1-capture-gate.md`에 기록한다. 카드 없이 수행하면 `.moai/reports/SPEC-MOAI-GATEWAY-001/m1-capture-gate.md`다
    (`.moai/docs/audit-artifact-convention.md`의 경로 규칙). `internal/gateway/testdata/` 아래와
    `.moai/reports/SPEC-MOAI-GATEWAY-001/` 아래 경로는 `git check-ignore`에 걸리지 않는다(2026-09-11, HEAD `81c1d58f9`에서 잼).
  - 결과를 가정 A-VAL-2.1.268의 네 조건(`design.md` §4.1)과 대조한다. 캡처한 검증 요청이 네 조건과 다르거나(예: `stream: null`) 같은 조건에서 네 조건을 모두 만족하는 다른 요청이 보이면, 구현을 멈추고 오케스트레이터에 blocker로 보고한다.
    인식 기준은 코드보다 먼저 이 SPEC에서 고친다.
  - 대상 Claude Code 버전이 2.1.268과 다르면 이 게이트를 그 버전에서 수행하고 버전을 기록한다.
- loopback mock upstream을 세워 `POST /v1/messages`의 `model` 값을 기록한다.
- 각 제공자 세션 안에서 허용 모델을 전환하고 turn 요청의 `model`이 선택과 일치하는지 본다. `/model <id>` 전환은 선택 시점
  검증 요청을 먼저 보내므로 turn 요청과 검증 요청을 구분해 기록한다. picker `s` 전환은 시험 하네스가 사용자 지정 picker
  항목 하나를 두어 재현한다(0.5.0 프로브 2와 같은 방식). t649 제품 판정에서는 실제 launcher가 만든 picker 구성을 사용한다.
- **검증 요청 인식 픽스처.** 진입 게이트의 원본 캡처에서 검증 요청 픽스처를 고정하고, 인식 기준이 그 픽스처는 받아들이고
  캡처한 turn·제목 생성 요청과 `AC-MG-003` (c)의 아홉 변형은 받아들이지 않는지 시험한다(`design.md` §4.1).
- **중단 조건**: `/model` 변경이 gateway 요청에 반영되지 않으면 추측으로 진행하지 않고
  보고한다(핸드오프 §7).
- **t649 회수.** 실제 MoAI launcher가 만든 picker에서 제공자별 행·Default·현재 행과 `s`/Enter/직접 입력을
  검증한다. 임시 Claude 직접 실행은 참고 실험이며 제품 판정을 대신하지 않는다. 공유 설정 해시와 다른
  제공자·일반 Claude의 시작 모델도 전후 비교한다.

- 대응 AC: AC-MG-003, AC-MG-001(최초 turn 요청의 모델 기록)

### M2 — 타입 계약, model registry, `CredentialRef` 인터페이스 (우선순위: High)

`ModelEntry` / `CatalogSnapshot` / `LaunchPlan` / `RequestContext` / `CredentialRef`의 형태를
테스트 우선으로 고정한다. 되돌리기 비용이 가장 큰 결정이고 형제 GPT-AUTH SPEC이 기다리는
seam이므로 M1 직후에 둔다.

- exact match 라우팅, 접두사 추측 금지.
- 네 GPT ID(`gpt-5.6-sol`, `gpt-5.6-terra`, `gpt-5.6-luna`, `gpt-6-astra`)만 등록.
  `gpt-5.3`·`gpt-6-astro` 부재 가드.
- GLM 항목은 `llm.glm.models`의 네 tier 값을 route ID로 Z.AI route에 등록한다(같은 ID는 한 항목). gateway `moai glm`의 tier
  매핑이 이 등록에 기댄다(`REQ-MG-019`, `REQ-MG-018`).
- launcher별 초기 모델·초기 provider 하나만 다름.
- 직접 secret route의 `CredentialRef`와 App Server `ManagedSessionAuthority`를 구분한다(`design.md` §2.1). profile/account/auth mode/generation/model 불일치는 실행 전 거절한다.
- 대응 REQ: MG-011(부분), MG-017, MG-019, MG-023 / AC: AC-MG-001, AC-MG-021

### M3 — CLI 계약: `moai gpt` 등록, 공통 launch plan, kanban/factory 진입 (우선순위: High)

사용자에게 보이는 표면이므로 앞쪽에 둔다.

- `internal/cli/gpt.go` — 닫힌 동사 집합 라우팅(`codexVerbRouting` 형태를 택한다).
  **선택 근거**: `glm`의 이중 등록(cobra `AddCommand` + `runGLM` 안의 수동 `args[0]`
  switch)은 같은 정보를 두 곳에 두어 어긋날 여지를 만든다. `codex`의 닫힌 집합은 한
  곳에서 정의되고 미등록 동사에 대한 진단 출력이 결정적이다. `login`/`logout`은 구현이
  형제 SPEC에 있으므로 **동사 집합에 등록하되, 기능이 아직 제공되지 않는다는 명시 오류**로
  시작한다. 그 오류는 사용자 용어로 말하고 내부 SPEC 식별자를 출력하지 않는다.
- launcher 이름을 **최소 세 곳**에 반영한다: `helpGroupFrequency`의 `launch` 항목,
  `rootHelpGroups()`의 Launchers 행, 그리고 `moai gpt` 실행 함수 안의 `spawnLaunch` 리터럴.
- `internal/cli/gateway_launch.go` — 세 launcher 공통 plan 조립.
- **초기 모델·선택·재개.** 제공자별 catalog와 명시 시작 모델을 구성한다. REQ-MG-019의 우선순위를
  적용하고 네 슬롯·fallback·Default·보조 요청을 같은 제공자로 묶는다. 자식 전용 modelPicker 설정과
  제공자·대화별 저장 격리를 구현한다. 타 제공자 ID는 UI와 무관하게 송신 전에 거절한다.


- **`BackendGPT`와 kanban/factory 진입** (`design.md` §7):
  - `internal/kanban`에 `BackendGPT = "gpt"` 추가. 감사 backend 집합에는 추가하지 않는다.
  - `moai gpt`에 `moai cc`·`moai glm`과 같은 네 진입 분기(factory lead, factory worker, kanban lead,
    kanban companion)를 신설한다. 네 분기 모두 `exportKanbanLaunchFacts`에 `kanban.BackendGPT`를
    넘기고, factory lead 분기는 `RecordFactoryRunStart`에도 넘긴다.
  - backend 상수 선언 주석, `Record.Backend` 주석, `EnvMoaiKanbanBackend` 주석을 "launcher의
    초기 provider" 의미로 다시 쓴다.
  - 웹 콘솔 `backendBadge`가 `gpt`에 대해 과금 방식을 단정해 표시하지 않게 한다.
- **출시 판단 입력.** 이 SPEC만 착지하면 `moai gpt`는 GPT credential 구체 타입이 없어 초기 모델
  요청이 모두 명시 오류로 끝난다. 도움말 노출과 출시는 `SPEC-MOAI-GPT-AUTH-001`(제안)의 착지와
  함께 판단하도록 오케스트레이터에 넘긴다(`design.md` §7.4). 이 SPEC은 출시 시점을 정하지 않는다.
  - 코어 노출·출시는 t649의 제공자별 실제 제품 picker·저장 격리 검증 완료를 기다린다. PICKER 형제 착지 대기는 대체되었다.
    그 SPEC이 착지하기 전에는 설정된 Claude 기본 모델이 빈 `moai cc`가 Claude Code가 저장한 `/model` 선택으로 시작할 수
    있다(`research.md` §15 프로브 1). 그러면 gateway 모델 ID가 시작 모델이 되는데도 초기 provider는 `claude`로 기록되고
    (`REQ-MG-026`), gateway를 거치지 않는 `claude` 실행이 저장된 그 ID를 Anthropic으로 보낼 수 있다. 기본 설정 디렉터리
    (`~/.claude`)에서의 저장은 측정하지 않았다(`research.md` §15.5).
- 대응 REQ: MG-001, MG-002, MG-003, MG-004, MG-019(초기 모델), MG-026 / AC: AC-MG-001, AC-MG-014, AC-MG-015, AC-MG-023

### M4 — supervisor와 프로세스 수명 (우선순위: High)

- gateway child 기동 + 파이프 포트 인계 + exec 순서(`design.md` §3.2).
- 고아 수거: 세션 PID 감시 + 수명 상한 이중 안전망. lead 세션이 끝나면 gateway child도 끝나고, 그 뒤 옛 주소로
  오는 요청은 연결 오류로 실패한다(`REQ-MG-008`, `design.md` §3.4). tmux pane teammate의 수명 계약은 형제
  `SPEC-MOAI-GATEWAY-TEAMMATE-001`(제안) 소관이다.
- 종료 코드·signal·PTY·job control 보존 확인.
- Windows는 spawn-and-wait 경로로 별도 계약. **판정 지점은 release PR 게이트다**(운영자 결정,
  `design.md` §3.5). supervisor Windows 시험은 launcher/supervisor를 소유한 패키지(`internal/gateway`
  또는 `internal/cli`)의 평범한 Go 시험으로 둔다. `test/integration/harness/` 아래에 두지 않고
  `integration` 빌드 태그를 달지 않는다 — `release-pr-multi-os.yml`의 windows-latest 레그가
  `-tags=integration` 없이 `./...`를 실행하기 때문이다. 시험은 Windows에서 `t.Skip` 하지 않는다.
- 카드·develop CI는 이 시험을 Windows에서 돌리지 않는다. Windows 회귀는 release PR 시점에야
  드러나며, 이것은 잔여 위험으로 남는다. 대화형 TTY·job control은 CI runner에서 재현되지 않는다.
- 카드 종료 시 Windows 절반은 Gap(판정 대기)으로 기록한다. release 배치의 리드가 아티팩트 보존 7일 안에
  읽어 `.moai/reports/SPEC-MOAI-GATEWAY-001/windows-release-verdict.md`에 기록하고, 놓치면
  `workflow_dispatch`로 다시 실행한다(`design.md` §3.5).
- 대응 REQ: MG-005, MG-006, MG-008, MG-009, MG-010 / AC: AC-MG-006, AC-MG-016, AC-MG-017

### M5 — ingress·Claude 도구 계약 (t651·t652로 대체)

Anthropic 형식 입력 검증과 제공자 경계를 유지한다. GPT 요청은 공개 새 입력·typed tool 결과를 App Server turn으로
연결한다. 직접 Responses 변환·opaque 복원 경로는 GPT 구독에서 사용하지 않는다. t651·t652 AC를 실행한다.

### M6 — adapter (t650·t653·t654로 대체)

Claude와 GLM adapter는 기존 회귀를 유지한다. GPT는 App Server managed/API 선택을 연결하고 model/list 지원값을
교차 검증한다. 인증·tool·thread·fork·compact 실제 증거가 나오기 전 전체 지원으로 표시하지 않는다.

### M7 — launch provider signal, `settings.local.json`, teammate 표시 계약 (우선순위: High)

**이 마일스톤의 회귀는 조용하다.** 놓치면 `moai cc`·`moai gpt` 세션의 요청이 gateway를 거치지 않고
Z.AI로 곧장 가거나, 세션 종료마다 사용자 설정이 변형될 수 있다. gateway 로그에는 흔적이 남지 않는다.
그래서 우선순위를 High로 올리고, 키 투영 비교 시험(`AC-MG-018`)을 구현보다 먼저 세운다.

- `internal/config/envkeys.go`에 `EnvMoaiLaunchProvider = "MOAI_LAUNCH_PROVIDER"` 등록.
- 공통 launch plan이 초기 provider(`claude` | `gpt` | `glm`)를 자식 env에 싣는다.
- **launch 단계 정리** — 세 launcher 모두 exec 전에 `settings.local.json`의 GLM 정리 키 집합(14키 —
  층 (a) 라우팅 키, 층 (b) 동작 영향 키, 비회귀 키)을 정리한다(`design.md` §6.1). 기존
  `removeGLMEnv`는 이 가운데 `CLAUDE_CODE_MAX_CONTEXT_TOKENS`를 지우지 않으므로, 그 함수를 그대로
  호출하는 것만으로는 계약을 채우지 못한다.
- **Claude child env의 상속 키 정리 (iter5 G5-B1)** — 세 launcher는 자식 env를 조립하기 전에 상속 env에서 GLM 정리 키
  집합 14키를 지우고, `moai cc`·`moai gpt`는 `Z_AI_API_KEY`도 지운다. gateway `moai glm`은 GLM credential 저장소에서 읽은
  `Z_AI_API_KEY`를 싣고, 네 GLM tier 모델 슬롯 키 `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL`을 `llm.glm.models`의
  `high`·`medium`·`low`·`fable` 값으로 더한다(2026-09-11 운영자 결정, `design.md` §6.7). `moai cc`·`moai gpt`도 네 슬롯을 해당 제공자의 허용 ID로 더한다. 정리는 launcher가 더하는 키보다 먼저 한다(`design.md` §3.2 6a, §6.1 프로세스 env 투영). 오늘
  조립기는 `os.Environ()`을 그대로 넘기므로 그대로 쓰면 계약을 채우지 못한다. 판정은 `AC-MG-018` (a)의 상속 키 판정이고,
  슬롯 키가 가리키는 ID의 route 해석은 `AC-MG-025`가 본다.
- **기록 금지** — 어떤 gateway launch도 `injectGLMEnv`를 부르지 않고, base URL이나 GLM credential을
  그 파일에 쓰지 않는다. `injectGLMEnv`는 이미 프로덕션 호출자가 없는 옛 쓰기 함수이며, 이 금지는 그것을
  되살리지 말라는 뜻이다(`design.md` §6.2).
- **판정 이관** — `hookProcessEnvHasGLM`은 signal이 있으면 `MOAI_LAUNCH_PROVIDER == glm`으로,
  `cleanupGLMSettingsLocal`은 signal이 있으면 정리를 건너뛴다. SessionStart의 `ensureGLMCredentials`도
  signal이 있으면 **분기 이전 함수 입구에서** 건너뛴다. 억제는 토큰 없음 분기의 credential 주입과 토큰
  있음 분기의 context 창 키 쓰기를 모두 덮는다(`design.md` §6.3). signal이 없는 세션은 세 곳 모두 기존
  동작을 유지한다(`design.md` §5.4).
- `internal/tmux`의 `sessionEnvHasGLM`·`hasGLMEnv`는 건드리지 않는다(§C Pre-flight 인계).
- **teammate 표시와 tmux 세션 env (결정 12)** — gateway launch로 시작한 세션은 SessionStart 체인이 끝난 시점에
  `settings.local.json` 최상위 `teammateMode`가 `"in-process"`다. tmux 세션 안이든 밖이든 같다. gateway launch는 tmux 세션
  env에 어떤 키도 쓰거나 지우지 않으며, `applyGLMMode`의 tmux 주입과 `applyCCMode`의 CLI tmux 정리는 gateway launch에서
  돌지 않는다(`design.md` §6.6). signal이 없는 세션의 `ensureTeammateMode`는 오늘 동작을 유지한다.
- **훅 tmux 억제** — signal이 있으면 `ensureTmuxGLMEnv`는 tmux 세션 env에 쓰지 않고, SessionEnd의 tmux 세션
  env 정리는 돌지 않는다(`design.md` §5.4). 두 훅 동작에 시험 대역을 끼울 경계를 만든다 — 오늘은 함수 안에서
  세션 관리자를 만들거나 `tmux` 명령을 직접 실행한다.
- **측정 과제 — in-process teammate의 gateway 주소 상속.** gateway launch 세션에서 in-process teammate를 띄우고, teammate의
  요청이 lead의 loopback gateway에 도착하는지(gateway 요청 기록), 그 요청이 어떤 모델 ID를 싣는지, teammate의 훅 env에
  `MOAI_LAUNCH_PROVIDER`가 보이는지 실측한다. 표시 방식이 실제로 in-process였는지(split pane이 생기지 않았는지)도 함께
  기록한다. 결과는 5구획 보고로 남긴다. **음성이면** — teammate 요청이 gateway에 오지 않거나, `"in-process"` 설정에도 pane이
  생기면 — 구현을 이어 가지 않고 오케스트레이터를 거쳐 운영자에게 blocker로 보고한다. 다른 경로로 주소를 흘려 넣어 우회하지
  않는다.
- **측정 과제 — lead 종료 뒤 teammate.** lead 세션이 끝난 뒤 in-process teammate가 남는지 실측한다. 결과는 `AC-MG-006`의
  lead 종료 뒤 요청 판정의 입력이 된다.
- 대응 REQ: MG-008, MG-018, MG-021, MG-022 / AC: AC-MG-006(lead 종료 뒤 요청), AC-MG-018

### M8 — 실패 의미론과 보안 마감 (우선순위: Medium)

- 무단 fallback 부재 가드(다중 mock upstream 계수), 미인증·미등록 모델 거절, redirect에 인증
  미전달.
- 로그 마스킹, credential argv 배제, 새 env 이름 `envkeys.go` 등록.
- Codex credential 불변식 가드(신규 구현이 아니라 고정).
- 대응 REQ: MG-022, MG-023, MG-024, MG-025 / AC: AC-MG-013, AC-MG-019, AC-MG-020, AC-MG-022

### M9 — 회귀와 통합 (우선순위: Medium)

- 기존 `moai cc` / `moai glm` 회귀, worktree·profile·spawn 경로 회귀.
- hooks·MCP·`/compact` 회귀는 대화형 세션이 필요하므로 Gap 처리 가능(AC-MG-012).
- 변경 패키지 테스트 + `go vet` + `golangci-lint run`, 나머지는 CI 판정.

## G. 안티패턴

- **모델 ID 접두사로 provider 추측.** exact match만 쓴다.
- **unknown path를 Anthropic으로 흘려보내는 catch-all.** 모르는 요청을 외부로 보내는
  경로다.
- **중간 EOF에 `message_stop`을 합성.** 실패를 성공으로 위장한다.
- **전역 tool ID map.** 병렬 subagent 요청이 서로의 매핑을 덮어쓴다.
- **`syscall.Exec`를 spawn-and-wait으로 "단순화".** `MOAI_SESSION_PID`, job control,
  종료 코드 전파 세 가지를 동시에 다시 증명해야 한다.
- **`internal/web`의 CSRF 미들웨어를 그대로 붙이기.** Claude Code는 `Sec-Fetch-Site`를
  보내지 않으므로 모든 요청이 403이 된다.
- **provider를 base URL에서 계속 추론하기.** gateway 아래에서 모든 base URL이 같다.
- **signal 부재를 "비-GLM"으로 읽기.** 부재는 gateway launch가 아니라는 뜻이다. gateway 이전 방식
  세션의 정리를 끊는다.
- **stale 키 정리를 SessionEnd 훅에만 맡기기.** 세션이 비정상 종료하면 훅이 돌지 않는다. 다음 launch가
  exec 전에 치운다.
- **settings 불변을 파일 전체 해시로 판정하기.** `ensureTeammateMode`와 `removeGLMEnv`가 `teammateMode`를
  정당하게 다시 써서, 올바른 구현도 적색이 된다. GLM 정리 키 집합 투영으로 판정한다.
- **settings `env`와 프로세스 env의 우선순위를 전제하기.** 측정하지 않았다.
- **in-process teammate가 lead의 gateway 주소를 물려받는다고 전제하기.** 측정하지 않았다. M7 측정 과제 전에는 그
  전제에 기대는 판정을 PASS로 적지 않는다.
- **gateway launch에서 tmux 세션 env를 연결 수단으로 되살리기.** tmux 세션 하나에는 env가 하나뿐이라 두 번째 gateway가
  덮어쓰고, 마지막 gateway 뒤에는 signal이 남는다(iter4 G4-B1). tmux pane teammate는 형제 SPEC 소관이다.
- **`"auto"`를 in-process로 읽기.** Claude Code 문서는 `"auto"`가 tmux 세션 안에서 split pane을 연다고 적는다. gateway
  launch는 `"in-process"`를 남긴다.
- **검증 요청 픽스처를 mock 요약에서 옮기기.** 0.5.0 프로브 기록은 `stream`을 `bool` 변환값으로 남겼다. 픽스처는 M1 진입
  게이트의 원본 캡처에서만 고정한다.
- **`BackendGPT`를 이름이 같은 감사 backend 집합에 넣기.** 다른 개념이다.
- **AC가 있으니 PASS해야 한다는 압력.** `AC-MG-021`의 계정 권한 부분은 미리 선언한 Gap이다.
- **텍스트 검색 결과만으로 결함·완료를 단정.** 부재 주장에는 범위와 양성 대조군을
  붙이고, 자기 산출물과 감사 보고서를 범위에서 뺀다.
- **범위 확대.** `cg` 철거와 PKCE 구현은 형제 SPEC이다. "이 파일을 여는 김에"가
  그 경계를 무너뜨린다.

## H. 확정된 결정

iter1 감사가 지목한 열린 질문 6건과 iter2 감사가 드러낸 설계 공백 1건은 모두 확정되었다. 아래는
결정과 그 반영 위치다. 결정 8은 iter3 감사에서, 결정 9·10은 0.5.0의 클라이언트 실측에서, 결정 11·12는 iter4 감사(0.6.0)에서
나왔다. 0.6.0에서 결정 6·10은 형제 SPEC으로 옮겼고, 결정 8은 결정 12로 대체되었으며, 결정 9는 M1 진입 게이트를 더해 코어에 남았다. 이 결정들은 Implementation Kickoff Approval을 대신하지 않는다.

**결정 1 — `moai gpt`의 `-k` / `-f`: `BackendGPT`를 추가해 완전 지원한다.**
운영자는 요구사항 예산 비용을 알고 이 선택을 했다. `REQ-MG-026`을 신설했고, 자리는
`REQ-MP-007`을 `REQ-MG-005`에 통합해 만들었다(두 요구사항은 같은 판정 표면 `AC-MG-006`을
공유한다). 수용 기준 예산도 25/25이므로 새 AC를 만들지 않고 launcher CLI 표면 AC인
`AC-MG-014`의 본문 (d)에 흡수했다. 배정 시험 13개는 어느 것도 축약되지 않았다. 의미 재정의와
두 함정(이름이 같은 감사 backend, 배지의 정액제 표시)은 `design.md` §7에 있다.

**결정 2 — launch provider signal: 새 `MOAI_*` 환경변수를 `internal/config/envkeys.go`에 등록한다.**
이름은 `MOAI_LAUNCH_PROVIDER`(상수 `EnvMoaiLaunchProvider`), 값은 `claude` | `gpt` | `glm`.
gateway launcher가 launch의 초기 provider로 설정한다. env는 exec에서 고정되므로 이 signal은 launch
시점 provider만 담는다 — 요청별 provider는 gateway 내부에서만 알 수 있다. 훅의 GLM 정리와 재주입을
억제하는 판정 기준은 이 signal의 **존재**, 즉 "gateway launch인가"다. gateway launch는
`settings.local.json`에 GLM credential을 쓰지 않으므로 훅이 치울 대상이 없기 때문이다(0.3.0에서
결정 7에 맞춰 판정 전제를 고쳤다). 반영: `REQ-MG-021`, `AC-MG-018`, `design.md` §5, M7.

**결정 3 — T09: 선행 측정으로 수행한다.**
run 단계의 첫 항목(M0)이며 Anthropic adapter를 만드는 어떤 마일스톤보다 앞선다. `REQ-MG-016`은
그 결과에 게이트된다. 측정이 음성이면 Anthropic passthrough adapter는 출시하지 않는다 — 설계
보고서 §7은 local token 충돌이 해결되지 않은 구독 모드를 출시하지 않는다고 적는다.

**결정 4 — Windows: release PR 게이트에서 판정한다 (0.3.1, 운영자 결정).**
0.2.0의 "Gap으로 선언하고 CI에 위임한다"는 이 결정으로 대체되었다. iter2 감사(G2-A4)가
`release-pr-multi-os.yml`의 release 시점 3-OS 전 패키지 실행을 지적했고, 운영자는 그 windows-latest
레그를 판정 지점으로 택했다.
- supervisor Windows 시험은 소유 패키지(`internal/gateway` 또는 `internal/cli`)의 평범한 Go 시험이며
  `integration` 빌드 태그를 달지 않는다. 그 레그는 `-tags=integration` 없이 `./...`를 실행한다.
- 판정은 release PR의 windows-latest 레그가 올린 이벤트 스트림 아티팩트에서, 이름을 정한 시험이
  `"Action":"pass"`로 끝났는지로 읽는다. 시험은 Windows에서 `t.Skip` 하지 않고, 아티팩트 부재는 PASS가
  아니다.
- 카드·develop CI는 이 시험을 Windows에서 실행하지 않는다. Windows 회귀가 release PR 시점에야 드러난다는
  것은 잔여 위험이다.
- 대화형 TTY·job control은 CI runner에서 재현되지 않으며, Windows에 대해 로컬에서 잰 것은 없다.
- 카드 종료 시 Windows 절반은 Gap(판정 대기)으로 기록하고, release 배치의 리드가 아티팩트 보존 7일 안에 읽어
  `.moai/reports/SPEC-MOAI-GATEWAY-001/windows-release-verdict.md`에 기록한다. 놓치면 `workflow_dispatch`로 다시
  실행한다(0.4.0, iter3 G3-A2). 검증 장소 문구는 0.4.0에서 `REQ-MG-009` 본문에서 `spec.md` §E로 옮겼다.
반영: `REQ-MG-009`, `spec.md` §E, `AC-MG-006`, `design.md` §3.5, `research.md` §10, M4.

**결정 5 — `CredentialRef`: 이 SPEC이 인터페이스를 고정한다.**
provider 중립 인터페이스로 `design.md` §2.1에 seam 계약을 정의했다. 형제
`SPEC-MOAI-GPT-AUTH-001`(제안)은 그 뒤에 GPT 구체 타입 하나를 구현한다. 반영: `REQ-MG-023`,
M2.

**결정 6 — picker의 `s` 경로: iter1 B4의 정정으로 해소한다.**
**[0.6.0 이관]** 이 결정이 다룬 `REQ-MG-020`은 폐기 묘비가 되었고 picker 표면은 형제 `SPEC-MOAI-GATEWAY-PICKER-001`(제안)로
옮겼다. 코어에 남은 것은 `~/.claude/settings.json` 쓰기 배제(`REQ-MG-006`으로 이동)뿐이다. 아래는 이관 전 기록이다.
`REQ-MG-020`을 측정 게이트로 바꾸고, 전역 settings 불변은 `s`에 기대지 않고 `~/.claude/settings.json`을
쓰기 대상에서 배제하는 것으로 보장한다. 측정은 0.5.0에서 클라이언트 수준으로 이루어졌다(아래).
근거 귀속을 바로잡는다. 0.1.0은 설계 보고서의 "확인되지 않았다"를 `s` 경로에 붙였는데, 보고서
§3의 그 문장은 **native picker의 disabled 행 지원**에 대한 것이었다. `s`가 미측정인 근거는 따로
있다 — 보고서는 `s` 경로라는 축을 **아예 측정하지 않았고**, 공식 문서의 서술을 인용했을 뿐이다.
**0.5.0 측정 결과.** Claude Code 2.1.267에서 picker `s`는 검증 요청 없이 확인 대화상자 뒤 세션 전용 전환을 하고
설정 파일을 만들지 않았다(프로브 2). `/model <id>`는 검증 요청을 보내고, 성공하면 선택을 기본값으로 저장했다
(프로브 1). 그래서 `REQ-MG-020`의 게이트는 측정된 동작에 근거한 안내로 바뀌었고, 측정하지 않은 picker Enter 경로는
안내에서 단정하지 않는다. gateway를 거친 재확인은 M1에 남는다. 바로 위 문단은 0.5.0 이전 상태의 귀속 기록이다.

**결정 7 — `settings.local.json` 계약: 정리한다, 거부하지 않는다 (0.3.0, 오케스트레이터 결정).**
iter2 감사는 gateway launch가 그 파일에 무엇을 쓰고 지우는지가 정해지지 않았다고 지적하고, 정리와
거부를 두 선택지로 제시했다. 오케스트레이터는 **정리**를 택했다. `REQ-MG-022`와 일관되고, `moai cc`가
오늘 이미 같은 정리를 수행한다는 선례와 맞기 때문이다.
- 세 launcher 모두 exec 전에 GLM 정리 키 집합을 정리한다. 0.3.1에서 그 집합을 `removeGLMEnv`의 목록에
  고정하지 않고 두 층 계약(라우팅 키, 동작 영향 키)과 비회귀 규칙으로 정했으며,
  `CLAUDE_CODE_MAX_CONTEXT_TOKENS`를 포함한다(14키, `design.md` §6.1).
- 어떤 gateway launch도 `ANTHROPIC_BASE_URL`이나 GLM credential을 그 파일에 쓰지 않는다.
- SessionStart의 `ensureGLMCredentials`는 `MOAI_LAUNCH_PROVIDER`로 억제하며, 억제는 두 분기의 파일 쓰기를
  모두 덮는다(`design.md` §6.3).
- settings `env`와 프로세스 env의 우선순위는 미측정이며, 정리는 바로 그 이유로 요구된다.
반영: `REQ-MG-021`, `REQ-MG-022`, `AC-MG-018`, `design.md` §5·§6, M7.

**결정 8 — tmux 세션 env: gateway 주소만 싣는다 (0.4.0, 운영자 결정).**
**[0.6.0 대체 — 결정 12]** tmux 세션 env 계약과 teammate pane 수명은 형제 `SPEC-MOAI-GATEWAY-TEAMMATE-001`(제안)로 옮겼고,
코어의 tmux 동작은 결정 12가 정한다. 아래는 대체 전 기록이다.
iter3 감사(G3-B1)가 `moai glm`이 tmux 안에서 GLM 라우팅을 tmux 세션 env에도 싣는데 계약은
`settings.local.json` 한 표면만 덮는다고 지적했다. 운영자는 tmux 세션 env에 gateway 주소만 싣기로 했다.
- tmux 세션 안의 gateway launch는 tmux GLM 정리 키 집합을 지운 뒤 loopback `ANTHROPIC_BASE_URL`과
  `MOAI_LAUNCH_PROVIDER`만 주입한다. 운반 키가 tmux로 전달되면 그 키만 더한다. GLM credential과 Z.AI 주소는
  쓰지 않는다.
- tmux 세션 env의 `ANTHROPIC_AUTH_TOKEN`은 지우거나 세션 접근 토큰으로 덮어쓴다. 두 정리 함수의 불일치에서
  훅 쪽 근거를 택했다(`design.md` §6.6).
- signal이 있으면 `ensureTmuxGLMEnv`의 tmux 쓰기와 SessionEnd tmux 정리를 억제한다.
- gateway child는 lead 세션이 끝나고 주소를 받은 teammate pane이 모두 끝날 때까지 남는다. 생존을 판정할 수
  없으면 드러나는 실패를 계약으로 삼는다.
- teammate의 tmux 세션 env 상속과 teammate pane 생존 판정은 미측정이며 M7의 측정 과제다.
반영: `REQ-MG-008`, `REQ-MG-018`, `REQ-MG-021`, `REQ-MG-022`, `AC-MG-006`, `AC-MG-018`, `design.md`
§3.4·§5.4·§6.6, M4·M7.

**결정 9 — 선택 시점 검증 요청: gateway가 로컬로 답한다 (0.5.0, 운영자 결정).**
프로브 1은 `/model <id>`가 선택 시점에 검증 요청을 보낸다는 것을, 프로브 2는 그 요청에 대한 401·404 응답이 전환을
막는다는 것을 관측했다(`research.md` §15).
- gateway는 그 요청을 upstream에 넘기지 않고 로컬 상태만으로 답한다. catalog에 없는 모델은 404 `not_found_error`,
  route credential 부재는 401 `authentication_error`, 그 밖에는 최소 성공 응답이다. fallback도 유료 호출도 없다.
- 인식 기준은 2.1.267의 실측 형태에 묶인 명명된 가정이며 픽스처로 고정한다(M1). 실제 turn이 로컬로 답해지지 않을
  만큼 좁게 둔다. 클라이언트가 형태를 바꾸면 인식이 빗나가 그 요청이 통상 경로로 upstream에 가는 잔여 위험이 있다.
- 전환 시점에 확인되는 것은 credential 존재뿐이다. 만료·폐기는 첫 실제 turn에서 드러나며, 검증 요청을 보내지 않는
  `s` 경로도 첫 turn에서 fallback 없이 드러나는 오류에 기댄다.
- **코어에 남기는 이유 (0.6.0, 운영자 결정).** 이 결정 없이 코어만 출시하면 `/model <id>`의 `max_tokens: 1` 검증 요청이
  upstream으로 전달되어 전환마다 최소한의 유료 호출이 생긴다(`design.md` §4.1 잔여 위험 "형태 변경"이 이미 적은 경로). 이는
  사용자가 예상하지 못한 유료 경로를 두지 않는다는 확정 방침(`REQ-MG-022`)과 충돌한다.
- **M1 진입 게이트 (0.6.0, iter4 G4-B2).** 인식 기준의 근거는 0.5.0 프로브가 아니라 M1 첫 작업의 원본 캡처다. 프로브는 억제
  플래그 넷을 켰고, `stream`을 `bool` 변환값으로, 요청을 파생 요약으로만 기록했다. 기준은 `stream` 키가 있고 값이 JSON
  `false`, `max_tokens`가 정수 `1`, 메시지 하나이고 `role`이 `user`, `tools`가 없거나 빈 배열인 네 조건으로 좁혔다. 캡처한
  형태가 다르면 구현을 멈추고 SPEC을 먼저 고친다. 인식 판정은 registry·credential 판정보다 먼저 한다(`design.md` §4).
반영: `REQ-MG-023`, `AC-MG-003`, `design.md` §4·§4.1, `research.md` §15.6, M1·M5.

**결정 10 — 초기 모델은 언제나 명시 `--model`로 넘긴다 (0.5.0, 운영자 결정).**
**[0.6.0 이관]** 이 결정은 형제 `SPEC-MOAI-GATEWAY-PICKER-001`(제안)로 옮겼다(결정 11). 설정된 기본 모델이 빈 경우와 재개
경로에서 "언제나 `--model`"이 정의되지 않았고(iter4 G4-B3), 그 답은 picker 표면의 결정과 함께 내려야 하기 때문이다. 코어의
`REQ-MG-019`는 초기 모델만 정하고 전달 방식은 정하지 않는다. 아래는 이관 전 기록이다.
프로브 1은 `/model <id>`의 선택이 새 세션의 기본값으로 저장된다는 것을 관측했다.
- 세 launcher는 초기 모델을 언제나 `--model`로 넘겨, 저장된 기본값이 launcher의 시작 모델을 바꾸지 못하게 한다.
  사용자가 준 `--model`은 그대로 이긴다.
- 시스템은 여전히 `~/.claude/settings.json`을 쓰지 않는다. `/model <id>`에 따른 Claude Code 자신의 쓰기는 막을 수
  없으므로 문서화하고, 세션 전용 전환에는 picker `s`를 안내한다.
- 명시 `--model`이 저장된 `model` 설정보다 우선하는지는 미측정이다. M1의 게이트 측정이며, 우선하지 않으면 이 결정은
  불충분하므로 대안을 미리 정하지 않고 운영자에게 되돌린다.
- 오늘 launcher는 해석된 모델이 비면 `--model`을 생략해 사용자 범위 마지막 선택에 맡긴다. 이 결정은 그 동작을
  바꾼다(`research.md` §15.4).
반영: `REQ-MG-019`, `REQ-MG-002`, `REQ-MG-020`, `AC-MG-001`, M1·M3.

**결정 11 — 두 표면을 형제 SPEC 제안으로 떼어 낸다 (0.6.0, 운영자 결정).**
iter4 감사(FAIL 0.80, STOP 신호)의 차단 넷 가운데 셋(G4-B1, G4-B3, G4-B4)이 picker 표면과 tmux teammate 표면에서 나왔다.
운영자는 두 표면을 코어에서 떼어 형제 SPEC 제안으로 넘겼다.
- `SPEC-MOAI-GATEWAY-PICKER-001`(제안) — `REQ-MG-020` 전체(폐기 묘비), `REQ-MG-019`의 "언제나 명시 `--model`" 조항(결정 10),
  `design.md` §6.7의 picker 구성·"Default" 행 위험·`moai cc`·`moai gpt` 슬롯 키 부분(gateway `moai glm`의
  tier 슬롯 키는 0.7.0에서 코어로 정함 — `REQ-MG-021`), `AC-MG-002`(폐기 묘비), `AC-MG-001`의 저장된 기본값 판정, `AC-MG-005`의 picker `s` 판정, M1의 overlay·`s`
  재확인·`--model` 우선순위 게이트, iter4 G4-B3·G4-B4와 권고 G4-A2·A3·A4·A10. 0.7.0에서 코어의 노출과 출시를 이 SPEC의
  착지에 묶었다(2026-09-11 운영자 결정, iter5 G5-B2, M3).
- `SPEC-MOAI-GATEWAY-TEAMMATE-001`(제안) — `REQ-MG-008`의 teammate pane 수명, `REQ-MG-021`의 tmux 세션 env 주입 계약(15키
  정리 집합 포함), `REQ-MG-018`의 tmux pane teammate GLM tier 질문, M7의 teammate pane 측정, iter4 G4-B1과 권고 G4-A5·A8·A9.
- 코어에 남긴 안전 조항: `~/.claude/settings.json` 쓰기 배제(`REQ-MG-006`으로 이동), gateway launch가 tmux 세션 env에 GLM
  credential·Z.AI 주소를 싣지 않는다는 금지와 훅 tmux 억제(`REQ-MG-021`, 결정 12로 재서술).
반영: `spec.md` §G·§H·`REQ-MG-006`·`008`·`018`·`019`·`020`·`021`·`022`, `acceptance.md` `AC-MG-001`·`002`·`005`·`006`·`018`·
`022`·§E, `design.md` §3.4·§6.6·§6.7, M1·M3·M4·M7.

**결정 12 — gateway launch는 in-process teammate만 허용한다 (0.6.0, 운영자 결정).**
- gateway launch로 시작한 세션에서 SessionStart 체인이 끝난 시점의 `settings.local.json` 최상위 `teammateMode`는
  `"in-process"`다. tmux 안이든 밖이든 같다. 오늘 `ensureTeammateMode`는 tmux 안에서 `"tmux"`, 밖에서 `"auto"`를 쓴다
  (`internal/hook/session_start.go:1021-1023`). Claude Code 문서는 `"auto"`도 tmux 안에서 split pane을 연다고 적으므로
  (`research.md` §16), 운영자 지시의 "`tmux`가 아닐 것"을 `"in-process"`로 구체화했다.
- gateway launch는 tmux 세션 env에 어떤 키도 쓰거나 지우지 않는다. G4-B1(소유권 없는 공유 tmux env)은 코어에서 사라진다.
  그 결과 gateway `moai cc`는 오늘 `moai cc`가 launch 때 하던 tmux GLM 키 정리도 하지 않는다(`design.md` §6.5).
- (0.7.0, iter5 G5-B1·G5-B3) 그 대신 세 launcher는 Claude child env를 조립하기 전에 상속 env에서 GLM 정리 키 집합을 지운다
  (`REQ-MG-021`). gateway `moai glm`은 그 정리 뒤에 GLM tier 슬롯 키 넷을 `llm.glm.models` 값으로 더한다(2026-09-11 운영자
  결정). in-process 한정은 tmux pane에서 Z.AI로 곧장 가는 경로를 좁히지만 닫지 못하며, 닫는 일은
  `SPEC-MOAI-GATEWAY-TEAMMATE-001`(제안) 소관이다(`REQ-MG-022`).
- signal 아래의 훅 tmux 억제(SessionEnd tmux 정리, `ensureTmuxGLMEnv`)는 유지한다. 이유는 `design.md` §5.4에 다시 적었다.
- in-process teammate가 lead의 gateway 주소를 물려받는지는 미측정이며 M7의 측정 과제다. 음성이면 운영자에게 blocker로
  돌아간다.
- 사용자에게 보이는 변화: `moai glm` gateway launch의 teammate는 형제 SPEC이 착지할 때까지 tmux pane으로 나타나지 않는다.
반영: `REQ-MG-008`, `REQ-MG-018`, `REQ-MG-021`, `REQ-MG-022`, `AC-MG-006`, `AC-MG-018`, `AC-MG-022`, `design.md`
§3.2·§3.4·§5.2·§5.4·§6.5·§6.6, `research.md` §16, M4·M7.

## I. 예산 대조

| 축 | 사용 | Tier L 상한 | 판정 |
|---|---|---|---|
| 요구사항 | 25 (그 밖에 폐기 묘비 `REQ-MG-007`·`REQ-MG-020` 두 줄, 계수 제외) | 25 | 상한 도달 — 0.13.0에서 `REQ-MG-027` 신설로 0.6.0의 여유 1을 소비했다. 추가 요구사항은 형제 SPEC 분리부터 다시 정해야 한다 |
| 수용 기준 | 25 (그 밖에 폐기 묘비 `AC-MG-002` 한 줄, 계수 제외) | 25 | 상한 도달 — 0.13.0에서 `AC-MG-026` 신설로 여유 1을 소비했다 |

0.5.0까지는 두 축이 모두 상한에 정확히 닿았다. 0.6.0은 결정 11로 두 표면을 형제 SPEC 제안으로 옮겨 각 축에 여유 하나를
만들었고, 결정 12는 새 번호 없이 `REQ-MG-021`에 접었다. **이 여유는 범위를 다시 넓히라는 뜻이 아니다.** 0.3.0부터 0.5.0까지의 개정은 새 요구사항이나
수용 기준 없이 기존 항목의 본문만 고쳤다.

0.5.0의 결정 9·10은 새 요구사항이 필요했지만, 운영자 결정에 따라 기존 요구사항의 정밀화로 접어 넣었다. 결정 9는
`REQ-MG-023`에, 결정 10은 `REQ-MG-019`(`REQ-MG-002` 교차 참조)에 들어갔고, 판정이 늘어난 수용 기준은 `AC-MG-001`과
`AC-MG-003` 둘이다. 계수는 그대로지만 판정 표면은 그대로가 아니다 — 결정 9는 외부 송신 전 거절이라는 기존 의무에
로컬 성공 응답이라는 새 동작을 더한다.

0.6.0에서 `AC-MG-001`은 저장된 기본값 판정을 내보내 판정이 줄었고, `AC-MG-003`은 인식 기준 변형이 늘었다. 0.7.0에서는
`AC-MG-003` (c)에 변형 셋이, `AC-MG-018` (a)에 Claude child env 상속 키 판정(gateway `moai glm`의 슬롯 키 값 판정 포함)과 tmux 밖 판정이, `AC-MG-025`에
GLM tier ID의 registry 해석 판정이 더해졌고 번호는 늘지 않았다.
`AC-MG-003`·`AC-MG-006`·`AC-MG-014`·`AC-MG-018`·`AC-MG-021`은 여러 판정을 운반한다. 이 압축이 더 심해져야 하는 요구가 생기면, 상한을 완화하거나 AC를 더 넓힐 것이 아니라
무엇을 형제 SPEC으로 밀어낼지부터 정해야 한다.

## J. 교차 참조

- `spec.md` — 요구사항 SSOT, ID 대응표
- `acceptance.md` — 수용 기준과 이관·게이트 시험 표
- `design.md` — 내부 경계, `CredentialRef`, 프로세스 모델, launch provider signal, `settings.local.json` 계약, kanban backend 의미
- `research.md` — 코드베이스 근거와 교차 렌즈 모순 C1~C4
- `reports/moai-proxy-three-provider-redesign-20260910.md` — 설계 원문(읽기 전용, 파일명 유지)
- `reports/moai-proxy-next-session-handoff-20260910.md` — 핸드오프(읽기 전용, 파일명 유지)
- `.moai/reports/SPEC-MOAI-PROXY-001/plan-audit-iter1.md` — iter1 감사(옛 경로·옛 식별자 보존)
- `.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter2.md` — iter2 감사
- `.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter3.md` — iter3 감사
- `.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter4.md` — iter4 감사(0.6.0 입력)
- `.moai/state/verify/25b43a41-c0ac-4110-9dad-f3983da6a527/gwprobe/README.md`, `.moai/state/verify/25b43a41-c0ac-4110-9dad-f3983da6a527/gwprobe2/README.md` — 0.5.0 클라이언트 실측(기계 로컬 경로, 요약 `research.md` §15)

## Receipt 변경분 감사 보정 (0.9.0 유지)

RPA-1은 design §4.3의 lifecycle 표로 new/exact resume/continue/선택 resume/fork의 UUID·retained family profile·receipt
사전 인계를 고정한다. 구현 순서는 private lifecycle+호환성 preflight → 정규화/후보 집합 → terminal 전 publish → 실제
재개/유실/병렬 음성군이다. native secure-storage namespace는 변경 전 값을 보존하고 사용자 credential을 복사하지 않는다.
RPA-2와 선택 RPA-A1은 같은 design 수용 표와 AC-MG-009의 정상 A/B·빈 후보·lookup miss 대조군으로 재감사한다.
4096개·8MiB 및 기존 활성화 게이트는 그대로다. receipt 보정 자체는 출력 상한/Windows 결정을 대신하지 않는다. 이후 받은 두 사용자 승인은 각각 design §4.3과 AUTH plan B/E에 별도 기록했다.

## Native 정책 구현 인계 (0.9.0 유지)

native-policy-plan-review.md의 PASS는 후보를 기존 목표 안에서 본문에 보강할 수 있다는 판정이며 제품 PASS가 아니다.
구현자는 design §4.3의 native policy 표·응답 thinking 문법·GPT title/high/keep-all/metadata·계량 계약을 따른다.
순서는 실제 raw003~008의 정책별 RED → native/Responses 문법과 wire 양성·무송신 음성 → native thinking/signature
응답 검증 → 승인된 receipt codec 호출 연결 → 실제 title+main·네 GPT 도구/회상/재개 통합이다. receipt 저장/정규화/
수명·launcher를 이 작업의 소유 범위로 가져오지 않는다. 기존 출력 상한 승인 분기와 AUTH Windows 본문은 그대로다.

Codex shipped context는 research §19.3에서 출처·기본/override/headroom을 읽고 별도 policy profile로 표현한다.
native cap 연구 근거를 2026-09-12 공식 문서로 확인했다. [Claude 모델 표](https://platform.claude.com/docs/en/models/overview)는
Opus 5·Sonnet 5의 context 1M, Haiku 4.5의 context 200K를 명시한다.
[GLM-5.1](https://docs.z.ai/guides/llm/glm-5.1)과 [GLM-4.7](https://docs.z.ai/guides/llm/glm-4.7)은
context 200K·최대 출력 128K·text 입력/출력을 명시한다. 이는 해당 모델 명목값이며 구독/계정 경로의 실측 수용 보장은 아니다.
[Claude model-config](https://code.claude.com/docs/en/model-config#correct-the-window-for-a-gateway-or-custom-model-id)의
custom window 환경변수 적용은 ID 인식·[1m]·압축 설정에 따라 다르다. UI의 200K 가정은 upstream cap 근거가 아니다.
[Token counting](https://platform.claude.com/docs/en/build-with-claude/token-counting)은 system·tools·images를 포함하며
실제 생성 usage와 차이가 가능한 추정이다. provider/model별 tokenizer와 모델 ID로 계량한다.
따라서 AS5에서 모드·정확 모델별 명목값, 계정 경로의 유효 한도, Claude 표시/자동 압축 기준, 실제 사용량을 구분한다.
현재 gateway_product_binding.go:88의 공통 1000000/Images:true는 소스 관측이며 실행 결함 확정이 아니다.
이를 provider별로 재검증하고 GLM text-only와 Claude 이미지 capability를 구별하는 양성/음성 시험을 둔다.
실계정 대형 입력·이미지·Windows runtime은 운영 검증 Gap으로 남기며 다른 모델 숫자로 대신 채우지 않는다.

## t649 실행 순서와 완료 경계

1. High — 제공자별 catalog·picker·초기/재개 모델·저장 격리 계약의 음성 시험을 먼저 고정한다.
2. High — reasoning 최종 item 보존과 hash-only receipt의 요청 권한·완료 게시를 연결한다.
   전체/부분 유실·다른 대화·credential 변경·미확인 family 전환은 송신 전에 거절한다.
3. High — 실제 MoAI PTY 세 목록, GPT-6 Astra 답변·tool 결과 후속 답변·동일 대화 resume를 검증한다.
4. Medium — 관련 회귀와 독립 감사를 실행하고 Windows는 GitHub CI 실행 결과로 판정한다.

근거는 `.moai/reports/t649/` 아래에 남긴다. 로그인·mock·직접 Claude UI·제품 UI·실제 provider 왕복은
각각 분리하여 기록한다. 아직 관측하지 않은 항목은 완료로 올리지 않는다. push·PR·병합·워크트리 제거는 별도 지시 범위다.
