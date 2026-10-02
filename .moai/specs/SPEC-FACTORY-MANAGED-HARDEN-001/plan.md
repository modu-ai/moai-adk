---
id: SPEC-FACTORY-MANAGED-HARDEN-001
title: "plan.md — 구현 계획 (F3·F4)"
version: "0.5.0"
created: 2026-10-03
updated: 2026-10-03
author: GOOS (manager-spec)
tier: M
---

# plan.md — 구현 계획

> 상태축 없음. 마일스톤은 우선순위와 선후 관계로만 적는다(시간 추정 없음). 순서는 **판단이 가장 바뀔 가능성이 큰 것부터** 검토하도록 §A 에 검토 우선순위를 따로 두고, 실행 순서는 §F 의 M1..M4 다.
>
> **개정 이력**: 0.4.0 은 운영자의 범위 분할 결정(F5 → 카드 t1459)에 따른 개정이다 — F5 마일스톤(구 M4)·수명 순서·시그널 관련 내용을 모두 뺐다. 이전 판은 git 이력(`95dfd85c8`, `820eff47f`, `951f2bfb6`)에만 있다. 처분표와 감사 이력은 progress.md §E.1.

## §A. Context

- **카드 / 트리**: t1409, 워크트리 `.moai/worktrees/t1409`, 브랜치 `WT-managed-session-hardening`, 기준 develop `7109e0900`(계획 커밋 이후도 소스는 동일 — 카드 브랜치에 코드 변경 없음).
- **부모**: SPEC-FACTORY-MANAGED-SESSION-001(`completed`). 본문·frontmatter 불변. 이 SPEC의 `related_specs` 가 유일한 연결이다.
- **아티팩트**: `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/{spec,plan,acceptance,design,progress}.md` + run 단계가 M1에서 만드는 증거 파일 `red-baseline.md`(같은 디렉터리, 추적됨, 감사 캐시 해시 대상 목록 밖). REQ 12 / AC 13.
- **Tier 판단 (축소 범위 재평가)**: **Tier M 유지.** 방법을 구분해 적는다. 측정: 기준 트리의 대상 비테스트 파일 크기 `wc -l` — `managed_codex_factory.go` 568, `managed_factory_session.go` 483(합 1,051줄); 기존 managed 시험 파일 4개 합 1,895줄(`managed_codex_factory_test.go` 673, `managed_factory_session_test.go` 747, `managed_failure_paths_test.go` 137, `managed_loopback_test.go` 338). **가정(측정 아님)**: 비테스트 변경 약 300줄(프레임 분류·정책표·쓰기 뮤텍스·턴 창 상태 약 230, 표식 오류·드라이버 상한·로그·상수 약 70; 근거는 같은 파일의 비슷한 함수 `call()`·`waitTurn()` 이 각각 약 25–30줄이라는 관측), 시험 약 700–900줄(최상위 10개·하위 케이스 약 40개, 하위 케이스당 약 15–25줄 가정). 합 약 1,000–1,200줄로 `spec-workflow.md` 의 M 범위(300–1,000줄)의 위쪽 경계에 걸친다 — LOC 는 지침이지 강제가 아니다. 파일 수는 비테스트 3(`managed_codex_factory.go`, `managed_factory_session.go`, `defaults.go`)·시험 3–5·문서 2(+CHANGELOG)·SPEC 증거 1 로 약 10–12개(M 범위 5–15 안). REQ 12 / AC 13 은 M 상한 16/16 안에 여유가 있다. 헌장적 변경·5산출물이 필요한 규모가 아니다. 분할 전의 추정(약 2,200–2,800줄)과 달리 경계에 걸리는 정도이므로 plan-audit 이 L 로 올리라 하면 따른다(판단은 사용자 몫).
- **수정 대상**: `internal/cli/managed_codex_factory.go`, `internal/cli/managed_factory_session.go`, `internal/config/defaults.go`, 기존 시험 파일 증보. 문서: `.moai/docs/factory-managed-session.md`(sync), CHANGELOG(sync). **수정하지 않는 곳**: `Start()` 의 자원 생성부(`MkdirTemp`–`cmd.Start`), `Close`, 소유자 진입의 시그널·정리 순서(전부 t1459).
- **PRESERVE(수정 금지)**: `.moai/specs/SPEC-FACTORY-MANAGED-SESSION-001/**`, `internal/factorymsg/store.go`, `managedSession` 인터페이스 세 메서드, `driveManagedFactorySession` 시그니처, `factoryMoAIMCPApprovalArgs` 와 그 승인 인수, `launch_exec_posix.go`/`launch_exec_windows.go`, `internal/cli/factory.go` 의 어휘 판별, `codex_launcher.go` 의 factory 분기.
- **검토 우선순위(되돌리기 어려운 판단 순)**:
  1. design.md D-1 응답 정책표와 broker elicitation 거부의 귀속 규칙 — 보안 경계다. 승인 권한과 조용한 재배달 루프의 탐지가 여기서 결정된다.
  2. design.md D-2 오류 분류표와 상한 3 — 실패 시 닫힌 기본값과 UNMEASURED 상수.
  3. 기계적인 것: 프레임 분류 리팩터, 쓰기 뮤텍스, 로그 줄과 로그 이음새, 문서 문장, 게이트.

## §B. Known Issues (관련 범주만)

- **B1 크로스플랫폼**: Windows 크로스 빌드는 회귀 가드다(AC-MH-009). 새 파일·OS 빌드 태그·`syscall` 참조를 더하지 않는다.
- **B3 서브에이전트 경계**: `internal/cli` 에 `AskUserQuestion` 호출 금지.
- **B4 frontmatter**: `created`/`updated`/`tags` 정규 이름(이 SPEC은 준수).
- **B5 CI 3층**: spec-lint, golangci-lint, 테스트가 각각 따로 적색일 수 있다. 사전 적색(guardstate census)과 신규 결함을 구별한다(acceptance.md §3).
- **B6 헤딩**: spec.md §F 는 `### Out of Scope — <topic>` H3 + `-` 불릿(작성 완료).
- **B8 트리 위생**: 커밋은 명시 pathspec. `.moai/state/` 무수정. 측정용 임시 시험 파일은 커밋 전에 삭제한다. `.moai/reports/` 는 gitignore 대상이라 증거를 거기에만 두면 커밋 그래프 밖이다 — RED 원문은 SPEC 디렉터리 안 추적 파일에 둔다.
- **B9 커밋**: 카드 id `t1409` 를 모든 커밋 메시지에 넣는다. `Authored-By-Agent: manager-develop`(run) 트레일러 단독 단락, 마지막 단락 `🗿 MoAI`. 레인은 push하지 않고 병합 SHA를 리더에 보고한다(gitflow-lane-protocol §4).
- **B10 PRESERVE**: §A 목록.
- **t1410 인접 (이 카드가 t1410 에 영향을 주는 곳과 주지 않는 곳만)** — 기준 트리 `7109e0900` 소스를 읽은 결과이며 줄 번호는 `grep`·열람 측정, 의존·충돌 결론은 추론이다. t1410 을 설계하지 않는다.
  - **F8(토큰 디렉터리 실패 경로 정리)**: `Start()` 의 `MkdirTemp`(`managed_codex_factory.go:395`)–`WriteFile`(`:401`)–`cmd.Start`(`:413`) 구간이 대상. 이 카드는 그 구간을 **건드리지 않는다**(측정: 이 카드가 `Start()` 에서 고치는 줄은 클라이언트 생성 `:431` 과 `initialized` 쓰기 `:438` 뿐). 따라서 F8 은 F3·F4 에 **의존하지 않는다**. F8 이 의존하는 것은 `Start`/`Close` 의 구조를 바꾸는 t1459 의 수명주기 설계다(추론). 이전(분할 전) 계획은 이 카드가 그 구간을 바꾸는 것을 전제했으나 분할로 사라졌다.
  - **F9(핸드셰이크 예산 완화)**: `DefaultManagedCodexReadyTimeout`(`defaults.go:117`)의 값이며 사용처는 `managed_codex_factory.go:135`(다이얼 `HandshakeTimeout`)와 `:422`(`readyCtx`). F3·F4 코드와 **독립**이다. F4 가 건드리는 턴 타임아웃 `DefaultManagedCodexTurnTimeout`(`:121`)은 다른 상수이고 세션 치명으로 남는다. 이 카드가 `defaults.go` 에 더하는 상수는 `:121` 뒤에 놓아 텍스트 충돌 면적을 피한다(추론).
  - **F13(`/readyz` 리디렉션 루프백 재검사)**: `managedCodexAppReady`(`:100-126`)의 `http.Client{Timeout}`(`:104`)에 `CheckRedirect` 없음, `client.Do`(`:113`). 이 카드가 그 함수를 건드리지 않으므로 **독립**.
  - **공유 시험 파일**: 이 카드는 가짜 App Server `serveFakeRPC` 에 서버 요청 주입 모드를 더하고 t1410 의 F8·F9 시험도 같은 가짜를 쓸 가능성이 높다(추론). `managed_codex_factory_test.go` 에서 텍스트 충돌이 날 수 있으며 t1410 이 이 카드 뒤에 실행되므로(factory 기록 `after=t1409`) 이 카드 위에 얹으면 된다.
- **t1408(TUI attach) 인접**: 같은 `managed_codex_factory.go` 를 건드릴 수 있다. 병합 충돌은 나중에 병합하는 쪽 몫이며 이 카드는 `read()`·`call()`·`startTurn`/`waitTurn` 을 건드리는 변경을 마일스톤별로 몰아 커밋해 충돌 면적을 줄인다.
- **t1459 인접**: F5 는 이 카드와 독립이다. 이 카드는 `Close`·`Start()` 자원 생성·소유자 진입 정리 순서를 건드리지 않으므로 t1459 가 같은 줄 영역을 마음대로 설계할 수 있다. 단 이 카드가 `Start()` 의 `:431`·`:438` 두 줄(클라이언트 생성과 `initialized` 쓰기)을 바꾸므로 t1459 가 그 두 줄을 읽을 때 이 카드 병합 뒤 기준으로 본다(추론).

## §C. Pre-flight — 측정 기준선 (기준 트리 `7109e0900`, 이 plan 실행)

`|` 를 담은 명령은 표 밖 fenced 블록(원문 바이트)에 적고, 표에서는 그 블록을 가리킨다.

| 항목 | 명령 | 관측 |
|---|---|---|
| 카드 트리 | `git rev-parse --show-toplevel` / `git branch --show-current` / `git rev-parse --short HEAD` | `…/.moai/worktrees/t1409` / `WT-managed-session-hardening` / 분할 개정 시작 시 `951f2bfb6` |
| 스코프 회귀(정리 접두) | 아래 블록 C-1 | `ok  github.com/modu-ai/moai-adk/internal/cli  27.048s` (같은 선택을 리터럴 부분 문자열 패턴으로 돌린 첫 측정은 `ok … 30.959s`; 두 패턴 모두 `-v` 최상위 66개) |
| 선택 수(`-list`) | 아래 블록 C-1 의 첫 줄 | `66`(이 개정에서 재측정). 표 이스케이프 형태는 `0` — plan-audit 측정과 일치 |
| 같은 선택, 접두 없이 | 리터럴 부분 문자열 패턴의 스코프 명령 | `--- FAIL: TestManagedSwitchDoesNotReachCodexLaneLoop (0.48s)` … `lane boundary: a lane session cannot mutate the queue` — 레인 환경 때문의 가짜 적색 |
| 브로커 lease 재배달 | `go test ./internal/factorymsg -run '^TestDispatchResultExactlyOnce$' -count=1 -v` | 하위 7개(`lost_receipt_redelivery` 포함) PASS, `ok … 1.671s` |
| Windows 빌드 | `GOOS=windows GOARCH=amd64 go build ./...` | 출력 없음(성공) |
| config | `go test ./internal/config -count=1` | `ok  …/internal/config  31.134s` |
| guardstate | `go test ./internal/guardstate -count=1` | **사전 적색**: `TestCensus_SetDifferenceEmptyBothDirections` — `.github/workflows/workflow-parse-guard.yaml exists on disk with no manifest entry`, `declared 19 entries against 20 workflow files` |
| template | `go test ./internal/template -count=1` | `ok  …/internal/template  414.986s` — **분 단위 스위트**, 임대 필수 |
| lint 버전 | `golangci-lint --version` | `v2.1.6` |
| 사전 구조 사실 | `grep -n DefaultManagedSessionMaxConsecutiveTurnFailures internal/config/defaults.go` | 출력 없음, exit 1 |
| 부모 zero-syscall | 아래 블록 C-2 | 출력 없음, exit 1 |
| RED 산출물 경로 | `git check-ignore -v .moai/reports/t1409/red-baseline.md` / 같은 명령을 `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` 에 | 앞은 `.gitignore:235:.moai/reports/*` exit 0(착지 불가), 뒤는 출력 없음 exit 1(착지 가능) |
| 해시 대상 | `internal/runtime/audit_cache.go` `planArtifactNames` | `acceptance.md`·`design.md`·`plan.md`·`research.md`·`spec.md`·`tasks.md` — `red-baseline.md` 는 목록 밖 |
| 문서 미러 | `ls internal/template/templates/.moai/docs/factory-managed-session.md` | `No such file or directory` — 이 문서에는 템플릿 미러가 없다 |
| 서버 요청 실측(임시 시험, 삭제함) | 기준 트리에 `zz_t1409_scratch_test.go`(시험 3개)를 두고 그 세 시험만 고르는 앵커 패턴으로 실행한 뒤 파일 삭제, `git status --short` 빈 출력 확인 | `drop: no client frame within 1s: … i/o timeout` · `collision: call result="" err=<nil>` · `string-id: err=managed codex app server connection closed` · `ok … 2.902s` |
| codex 스키마 | `codex app-server generate-json-schema --out <dir>` (codex-cli 0.160.0) | `ServerRequest.json` 의 method 10종, `RequestId = anyOf[string, int64]`, `McpServerElicitationRequestParams` 필수 필드 `serverName`·`threadId`, `turnId` 는 `string`·`null` 확인 |
| 파일 크기(Tier 판단) | `wc -l internal/cli/managed_codex_factory.go internal/cli/managed_factory_session.go` 및 시험 4개 | 568 · 483 / 673 · 747 · 137 · 338 |

**C-1 (스코프 회귀와 선택 수 — 원문)**

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test ./internal/cli -list '^.*(Managed|managed).*$' | grep -c '^Test'
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test ./internal/cli -run '^.*(Managed|managed).*$' -count=1
```

기대: 첫 줄 66 이상(기준 66), 둘째 줄 `ok`. `[no tests to run]` 는 실패다.

**C-2 (부모 zero-syscall — 원문)**

```
grep -rn 'syscall\.' internal/cli/managed_*.go
```

**Gaps(미관측)** — 이름 붙은 항목:

- **라이브 Codex 관측(실제 codex 세션에서의 거부·오류 응답에 대한 모델 행동, 특히 MoAI 브로커 elicitation, 그리고 요청의 `serverName` 에 config 키 `moai` 가 실제로 오는지)**: 미관측이며 **run 진입 조건이 아니다**. 라이브 게이트 환경변수는 `MOAI_FACTORY_LIVE_ROOT` 와 `MOAI_FACTORY_LIVE_RUN`(부모의 `TestManagedCodexFactoryBrokerLive`, 지금은 SKIP). 돌린다면 볼 것은 design.md D-1 "미관측 전제와 Gap" 의 세 가지다.
- 세션 종료 후 orphan claim 행의 운명(소스 판독 추론), 부모 sync 기록의 managed 커버리지 89.3% 재측정.
- elicitation 귀속의 읽기 고루틴 대 소비자 경합은 plan 작성자가 재현하지 않았다(plan-audit 이 소스 판독과 교차 모델 보고로 지적) — 재현은 M3 의 시험(AC-MH-006 #11)이 한다.
- 명시한 한계(닫지 않음): 막힌 쓰기의 상한 없음(연결 사망·세션 종료까지; `call()` 의 `WriteJSON` `:212` 가 `select` `:215` 앞), `id: null` 프레임(design.md D-1 "공시한 한계").

**영향 가능 가드(run 단계가 재측정)**: 이 변경은 `.go` 파일과 `.moai/docs` 문서(미러 없음)뿐이다. `internal/guardstate` 의 census 는 워크플로 파일 대 매니페스트를 보는 것이라 직접 영향이 없고(기준 트리에서 이미 적색), `internal/template` 의 `.moai/docs` 접근 시험은 다른 문서(`jev-local-operations.md`)를 읽는다(grep 확인: `factory-managed-session` 문자열을 가진 template/guardstate 시험 0건). `internal/config` 는 `defaults.go` 상수 추가의 영향을 받는다. 이 셋을 §E 명령으로 재측정해 사전 적색과 신규 실패를 구별한다.

## §D. Constraints (위반 금지)

- 부모 SPEC 디렉터리, `store.go`, `managedSession` 인터페이스, `driveManagedFactorySession` 시그니처, 승인 인수는 건드리지 않는다.
- 금지 명령: `git add -A`, `git add .`, `git commit -a`(명시 pathspec만), `--no-verify`, `--amend`, force push, `git stash`.
- 승인 권한은 넓히지 않는다. 응답 정책표에 accept 계열을 넣지 않는다.
- 연속 실패 상한은 `defaults.go` 상수이고 구현에 숫자를 쓰지 않는다.
- 시험 이음새는 로그 출력 대상 원자 포인터 하나뿐이다(design.md D-1). 다른 시험 훅을 만들지 않는다.
- 로컬에서 `go test ./internal/cli` 전체를 돌리지 않는다. `internal/template` 스위트는 `moai slot acquire --resource internal-template-suite --max-duration 15m` → 실행 → `moai slot release --resource internal-template-suite` 안에서만 돈다(측정 414.986s).
- 문서 한국어 / 코드 주석·커밋·로그 영어.

## §E. Self-Verification (run 단계 완료 보고용 — 명령은 단일 호출형, `go test` 는 acceptance.md 머리말의 정리 접두 사용)

- AC 매트릭스: acceptance.md §1 의 명령과 §1.1·§1.2 블록 그대로. 원문 출력과 트리 HEAD SHA를 progress.md §E.2 에 인용한다(명령·관측·귀속 3요소).
- 스코프 회귀: 위 C-1 블록의 두 줄(선택 수 66 이상 확인 뒤 `ok`; `[no tests to run]` 는 실패).
- 인접 패키지: `go test ./internal/config -count=1`, `go test ./internal/guardstate -count=1`(사전 적색 1건 고려), `go test ./internal/template -count=1`(임대 안에서)
- Windows: `GOOS=windows GOARCH=amd64 go build ./...`, `GOOS=windows GOARCH=amd64 go vet ./internal/cli/`
- 정적: `gofmt -l internal/cli internal/config`, `go vet ./internal/cli ./internal/config`, `golangci-lint run ./internal/cli/... ./internal/config/...`
- 경계 grep: acceptance.md §5.
- 변이 확인(acceptance.md §2.4 의 mu1–mu18)과 섭동 확인 P1(acceptance.md §6 DoD 3·4): 변이를 **하나씩** 임시 적용해 지목한 케이스가 붉어지는지 보이고, 각각 되돌린 뒤 `git diff --stat` 빈 출력을 확인하고 원문을 progress.md §E.2 에 인용한다.

## §F. Milestones (실행 순서)

### M1 — RED 기준선 (독립 커밋, 가장 먼저; 시험 기반 AC의 채택 사건)

- **선결**: RED 원문을 담을 경로가 추적되어야 한다. 이 SPEC은 `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` 를 쓴다(측정: gitignore 대상 아님, 감사 캐시 해시 목록 밖). `.moai/reports/t1409/` 에 두는 사본은 로컬 편의일 뿐 근거로 인용하지 않는다.
- 재현 시험 **6개**(F3 4·F4 2)를 기준 트리 API에 대해 컴파일되는 형태로 쓴다(새 심볼 없이): `TestManagedCodexServerRequestPolicy`, `TestManagedCodexTurnSurvivesServerRequest`, `TestManagedCodexServerRequestIDCollision`, `TestManagedCodexDeclinedBrokerElicitationFailsTurn`, `TestManagedDriverIsolatesTurnFailure`, `TestManagedCodexNonCompletedTurnIsolated`. 각 시험이 어떤 이유로 붉어야 하는지의 표는 acceptance.md §2.1 이 정본이다(여기서 되풀이하지 않는다). RED-first 면제와 대체 채택(변이)도 같은 곳에 있다.
- 각 시험을 기준 트리에서 돌려 **옳은 이유로 붉은지** 확인하고 원문을 `red-baseline.md` 에 적는다: 명령, 출력 원문(50줄/2KB 넘으면 파일 리다이렉트 + 꼬리), exit 코드, 측정 트리 SHA, 붉은 이유 한 줄, 선택 수(`-list`). 컴파일 오류, 패키지·하네스 타임아웃, 시험 인프라 실패로 붉은 것은 옳은 이유가 아니다(wrong-reason red) — 시험을 고친다. **시험 자신이 정한 시간 상한 단언**이 깨져서 붉은 것은 옳은 이유다(acceptance.md §2.1).
- 시험 파일과 `red-baseline.md` 를 **한 커밋**에 담는다. 이 커밋이 수리 커밋들의 조상이어야 한다(REQ-MH-011, AC-MH-011). 같은 커밋에 구현 변경을 섞지 않는다(커밋 그래프가 선후의 유일한 증인이다: verification-claim-integrity §2.3). 커밋 게이트가 실패하는 시험을 담은 커밋을 거부하면 `--no-verify` 를 쓰지 말고 blocker 보고로 돌려준다(측정: `git config --get core.hooksPath` = `/dev/null` 이라 거부 위험은 낮다).
- 커밋 제목 예: `test(SPEC-FACTORY-MANAGED-HARDEN-001): M1 RED baseline for F3 F4 (card t1409)`.

### M2 — F3 서버 요청 응답 (되돌리기 가장 어려운 판단: 정책표)

- 프레임 분류(`id` 원문 보존), 서버 요청 읽기 고루틴 응답, 응답 정책표(한 곳의 표 데이터)와 그 곳에서 쓰는 로그 한 줄, 로그 출력 대상 원자 포인터(기본 `os.Stderr`), 연결 쓰기 뮤텍스(`call()`, `initialized` 알림, 답장 모두 경유), 읽기 고루틴의 턴 창 상태(`armTurn` 은 `turn/start` 쓰기 전, `prevTurnID` 유지, 판정은 `turn/completed` 프레임을 본 읽기 고루틴이 완료 이벤트에 실음 — design.md D-1 결정 3)와 `moaiMCPServerKey` 상수 비교. 표식 오류 심볼은 M3 가 도입하므로 M2 는 판정을 이벤트에 싣는 데까지, `DeliverTurn` 의 표식 오류 반환 연결은 M3.
- GREEN: AC-MH-001, 002, 003, 004. 변이 확인: mu1–mu5. M1의 F3 재현 시험 중 앞 세 개가 뒤집힌다.
- 커밋: `fix(SPEC-FACTORY-MANAGED-HARDEN-001): M2 answer server-originated codex requests (card t1409)`.

### M3 — F4 턴 단위 실패 격리와 elicitation 판정 연결

- 턴 단위 표식(`errors.Is` 판별)을 스트림 `result.is_error`, Codex `completed` 아닌 종료, 거부된 `moai` broker elicitation이 귀속된 턴(완료 이벤트가 실어 온 판정), 이 세 곳에서만 붙인다. 드라이버: 표식 있으면 로그+계속, 연속 횟수 상한, 성공 시 0으로, 우선 턴은 세지 않음. `defaults.go` 에 `DefaultManagedSessionMaxConsecutiveTurnFailures = 3`(`DefaultManagedCodexTurnTimeout` 뒤, UNMEASURED 표기, 근거 주석).
- GREEN: AC-MH-005, 006(하위 17개 + 서버 이름 고정 시험), 007, 008. M1의 F4 재현 시험 둘(`TestManagedDriverIsolatesTurnFailure`, `TestManagedCodexNonCompletedTurnIsolated`)과 `TestManagedCodexDeclinedBrokerElicitationFailsTurn` 이 뒤집힌다. 변이 확인 mu6–mu18(acceptance.md §2.4).
- 커밋: `fix(SPEC-FACTORY-MANAGED-HARDEN-001): M3 isolate turn failures with a ceiling (card t1409)`.

### M4 — 게이트와 증거 (기계적)

- §E 의 정적·스코프·Windows·인접 패키지 명령을 돌리고 progress.md §E.2/§E.3 에 인용. 변이 확인 전부(mu1–mu18)와 섭동 확인 P1, 커밋 구성 점검 C1. 경계 grep. `./internal/template` 스위트는 슬롯 임대(`moai slot acquire --resource internal-template-suite --max-duration 15m`) 안에서 한 번 돌리는 스모크로만 둔다(AC 근거 아님). 변이 미채택 불변 가드 G1–G8 은 acceptance.md §2.5 가 정본이다.
- 문서·CHANGELOG(AC-MH-010)는 **sync 단계**(manager-docs)의 몫이다: `.moai/docs/factory-managed-session.md` 의 "알려진 한계"에서 F3·F4 문장을 고치거나 지우고(acceptance.md §1.2 의 고정 앵커대로) 남은 한계를 적으며, **F5 "시그널 처리 공백" 줄은 해결 주장이 아니라 카드 t1459 를 가리키는 정확한 안내로 남기고** "후속 카드 t1409 대상" 단락을 고친다. 이 SPEC의 CHANGELOG 엔트리를 쓰고 부모 엔트리의 "후속 카드 t1409 대상" 문구를 정정한다. 부모 SPEC 파일은 건드리지 않는다.

## §G. Anti-Patterns

- RED와 수리를 한 커밋에 합치기(선후 증인 상실).
- RED 원문을 gitignore 대상 경로에만 두기(커밋 그래프 밖).
- 컴파일 실패·하네스 타임아웃으로 붉은 시험을 RED 기준선으로 인용하기(옳은 이유가 아님).
- 표 안에 이스케이프 `\|` 를 둔 `-run`/`-list` 명령을 그대로 복사하기(테스트 0개 선택, 공허 `ok`).
- 구현에 `3` 이나 `5초` 를 코드 상수로 박기.
- 승인류 응답을 "편의상" accept 로 바꾸기, 또는 `cancel`/`abort` 로 턴을 끊기.
- elicitation 판정을 소비자 쪽 계수기로 하기(읽기 고루틴이 앞서 달리면 정상 턴을 실패로 센다).
- 이 카드에서 `Close`·`Start()` 자원 생성부·소유자 진입 순서를 건드리기(t1459 몫이며 t1410 과 충돌한다).
- 시험용 훅을 더 만들기(이 SPEC의 이음새는 로그 대상 원자 포인터 하나).
- 라이브 미관측 사항을 해결로 공시하기(REQ-MH-010), 또는 F5 를 해결로 공시하기, 또는 미관측 전제를 관측으로 귀속하기.

## §H. Cross-References

- `spec.md` §C REQ-MH-001..012, `acceptance.md` AC-MH-001..013, `design.md` D-1, D-2.
- 부모: `.moai/specs/SPEC-FACTORY-MANAGED-SESSION-001/{spec,design,acceptance}.md` — D-1(배달 전용), AC-MS-010/014/017, research.md `:87`(라이브 재확인 대상).
- `.moai/reports/t1409/intake.md`, `.moai/reports/t1409/t1375-f3-f5-verbatim.md`, `.moai/reports/t1409/plan-audit-iter{1,2,3}.md`(로컬 사본, gitignore 대상).
- `.claude/rules/moai/core/verification-claim-integrity.md` §2.3(순서 증인), `.claude/rules/moai/development/verification-completeness.md` §2(RED-now/green path), `.claude/rules/local/gitflow-lane-protocol.md` §4·§8(push·부하).

## §I. 미해결 사항

운영자에게 되물어야 할 미해결 항목은 없다(해결 표식 0건). 남은 불확실성은 질문이 아니라 관측 공백(Gaps)이며 §C 와 design.md D-1 "미관측 전제와 Gap" 에 있다.
