---
id: SPEC-FACTORY-MANAGED-HARDEN-001
title: "plan.md — 구현 계획"
version: "0.1.0"
created: 2026-10-03
updated: 2026-10-03
author: GOOS (manager-spec)
tier: M
---

# plan.md — 구현 계획

> 상태축 없음. 마일스톤은 우선순위와 선후 관계로만 적는다(시간 추정 없음). 순서는 **판단이 가장 바뀔 가능성이 큰 것부터** 검토하도록 §A 에 검토 우선순위를 따로 두고, 실행 순서는 §F 의 M1..M5 다.

## §A. Context

- **카드 / 트리**: t1409, 워크트리 `.moai/worktrees/t1409`, 브랜치 `WT-managed-session-hardening`, 기준 develop `7109e0900`(plan 시작 시 `git diff --stat develop...HEAD` 빈 출력 = 카드 브랜치에 코드 변경 없음).
- **부모**: SPEC-FACTORY-MANAGED-SESSION-001(`completed`). 본문·frontmatter 불변. 이 SPEC의 `related_specs` 가 유일한 연결이다.
- **아티팩트**: `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/{spec,plan,acceptance,design,progress}.md`. Tier M(REQ 16 / AC 15, 상한 16/16). design.md는 세 설계 판단을 담기 위해 Tier M 기본 세트에 더한 것이다(스키마상 허용: Artifact Statelessness는 Tier와 무관).
- **수정 대상**: `internal/cli/managed_codex_factory.go`, `internal/cli/managed_factory_session.go`, `internal/config/defaults.go`, 신규 `internal/cli/launch_signals.go`(+ 시험), 기존 시험 파일 증보. 문서: `.moai/docs/factory-managed-session.md`(sync), CHANGELOG(sync).
- **PRESERVE(수정 금지)**: `.moai/specs/SPEC-FACTORY-MANAGED-SESSION-001/**`, `internal/factorymsg/store.go`, `managedSession` 인터페이스 세 메서드, `factoryMoAIMCPApprovalArgs` 와 그 승인 인수, `launch_exec_posix.go`/`launch_exec_windows.go`, `internal/cli/factory.go` 의 어휘 판별, `codex_launcher.go` 의 factory 분기.
- **검토 우선순위(되돌리기 어려운 판단 순)**:
  1. design.md D-1 응답 정책표 — 보안 경계다. 승인 권한이 이 표에서 결정된다.
  2. design.md D-2 오류 분류표와 상한 3 — 실패 시 닫힌 기본값과 UNMEASURED 상수.
  3. design.md D-3 시그널 설계(B안) — 드라이버 시그니처 변경 10곳, `Close` 동시 안전화, `launch_signals.go` 의 `syscall.` 배치.
  4. 기계적인 것: 프레임 분류 리팩터, 로그 한 줄, 문서 문장, 게이트.

## §B. Known Issues (관련 범주만)

- **B1 크로스플랫폼**: Windows 크로스 빌드가 게이트다(AC-MH-011). `syscall.SIGTERM`·`SIGHUP` 은 Windows에서 컴파일됨을 실측했다(design.md D-3). 새 파일에 OS 빌드 태그를 두지 않는다.
- **B3 서브에이전트 경계**: `internal/cli` 에 `AskUserQuestion` 호출 금지.
- **B4 frontmatter**: `created`/`updated`/`tags` 정규 이름(이 SPEC은 준수).
- **B5 CI 3층**: spec-lint, golangci-lint, 테스트가 각각 따로 적색일 수 있다. 사전 적색(guardstate census)과 신규 결함을 구별한다(acceptance.md §3).
- **B6 헤딩**: spec.md §F 는 `### Out of Scope — <topic>` H3 + `-` 불릿(작성 완료).
- **B8 트리 위생**: 커밋은 명시 pathspec. `.moai/state/` 무수정. 측정용 임시 시험 파일은 커밋 전에 삭제한다.
- **B9 커밋**: 카드 id `t1409` 를 모든 커밋 메시지에 넣는다. `Authored-By-Agent: manager-develop`(run) 트레일러 단독 단락, 마지막 단락 `🗿 MoAI`. 레인은 push하지 않고 병합 SHA를 리더에 보고한다(gitflow-lane-protocol §4).
- **B10 PRESERVE**: §A 목록.
- **t1410 인접**: t1410(F8·F9·F13)이 이 카드 뒤에 실행된다(factory 기록 `after=t1409`). `Start()` 는 t1410의 F8(조기 실패 시 토큰 디렉터리) 영역과 같은 함수이므로 이 카드의 `Start()` 변경은 준비 대기 컨텍스트 파생 한 줄로 최소화한다.
- **t1408(TUI attach) 인접**: 같은 `managed_codex_factory.go` 를 건드릴 수 있다. 병합 충돌은 나중에 병합하는 쪽 몫이며 이 카드는 쓰기 지점·`Close`·`read()` 를 건드리는 변경을 한 번에 몰아 커밋해 충돌 면적을 줄인다.

## §C. Pre-flight — 측정 기준선 (기준 트리 `7109e0900`, 이 plan 실행)

| 항목 | 명령 | 관측 |
|---|---|---|
| 카드 트리 | `git rev-parse --show-toplevel` / `git branch --show-current` / `git rev-parse --short HEAD` | `…/.moai/worktrees/t1409` / `WT-managed-session-hardening` / `7109e0900` |
| 스코프 회귀(정리 접두) | `unset MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli -run '^.*(Managed\|managed).*$' -count=1` | `ok  github.com/modu-ai/moai-adk/internal/cli  27.048s` (같은 선택을 리터럴 부분 문자열 패턴으로 돌린 첫 측정은 `ok … 30.959s`; 두 패턴 모두 `-v` 최상위 66개) |
| 같은 선택, 접두 없이 | 리터럴 부분 문자열 패턴의 스코프 명령 | `--- FAIL: TestManagedSwitchDoesNotReachCodexLaneLoop (0.48s)` … `lane boundary: a lane session cannot mutate the queue` — 레인 환경 때문의 가짜 적색 |
| 브로커 lease 재배달 | `go test ./internal/factorymsg -run '^TestDispatchResultExactlyOnce$' -count=1 -v` | 하위 7개(`lost_receipt_redelivery` 포함) PASS, `ok … 1.671s` |
| Windows 빌드 | `GOOS=windows GOARCH=amd64 go build ./...` | 출력 없음(성공) |
| config | `go test ./internal/config -count=1` | `ok  …/internal/config  31.134s` |
| guardstate | `go test ./internal/guardstate -count=1` | **사전 적색**: `TestCensus_SetDifferenceEmptyBothDirections` — `.github/workflows/workflow-parse-guard.yaml exists on disk with no manifest entry`, `declared 19 entries against 20 workflow files` |
| template | `go test ./internal/template -count=1` | `ok  …/internal/template  414.986s` — **분 단위 스위트**, 임대 필수 |
| lint 버전 | `golangci-lint --version` | `v2.1.6` |
| 사전 구조 사실 | `grep -rn 'syscall\.' internal/cli/managed_*.go` / `grep -rn 'signal.Notify\|NotifyContext' internal/cli/managed_*.go` / `grep -n DefaultManagedSessionMaxConsecutiveTurnFailures internal/config/defaults.go` | 셋 다 출력 없음, exit 1 |
| 문서 미러 | `ls internal/template/templates/.moai/docs/factory-managed-session.md` | `No such file or directory` — 이 문서에는 템플릿 미러가 없다 |
| 서버 요청 실측(임시 시험, 삭제함) | 기준 트리에 `zz_t1409_scratch_test.go`(시험 3개)를 두고 `go test ./internal/cli -count=1 -v` 에 그 세 시험만 고르는 앵커 패턴을 줘 실행한 뒤 파일 삭제, `git status --short` 빈 출력 확인 | `drop: no client frame within 1s: … i/o timeout` · `collision: call result="" err=<nil>` · `string-id: err=managed codex app server connection closed` · `ok … 2.902s` |
| 시그널 기본 동작 실측(작은 프로그램) | 임시 프로그램에서 SIGTERM/SIGHUP 후 상태 관측 | `SIGTERM launcher_rc=143 child=alive tokendir=present defer=not-run` · `SIGHUP launcher_rc=129 child=alive tokendir=present defer=not-run` (SIGINT 미관측: 비대화형 셸의 `&` 잡은 SIGINT를 무시) |
| codex 스키마 | `codex app-server generate-json-schema --out <dir>` (codex-cli 0.160.0) | `ServerRequest.json` 의 method 10종과 `RequestId = anyOf[string, int64]` 확인 |

**Gaps(미관측)**: 실제 codex 세션에서 거부·오류 응답에 대한 모델 행동(라이브 게이트 `MOAI_FACTORY_LIVE_ROOT` 미설정). 실제 런처에 대한 kill 실험. SIGINT의 기본 동작. 세션 종료 후 orphan claim 행의 운명(소스 판독 추론). 부모 sync 기록의 managed 커버리지 89.3% 재측정.

**영향 가능 가드(run 단계가 재측정)**: 이 변경은 `.go` 파일과 `.moai/docs` 문서(미러 없음)뿐이다. `internal/guardstate` 의 census 는 워크플로 파일 대 매니페스트를 보는 것이라 직접 영향이 없고(기준 트리에서 이미 적색), `internal/template` 의 `.moai/docs` 접근 시험은 다른 문서(`jev-local-operations.md`)를 읽는다(grep 확인: `factory-managed-session` 문자열을 가진 template/guardstate 시험 0건). `internal/config` 는 `defaults.go` 상수 추가의 영향을 받는다. 이 셋을 §E 명령으로 재측정해 사전 적색과 신규 실패를 구별한다.

## §D. Constraints (위반 금지)

- 부모 SPEC 디렉터리, `store.go`, `managedSession` 인터페이스, 승인 인수는 건드리지 않는다.
- 금지 명령: `git add -A`, `git add .`, `git commit -a`(명시 pathspec만), `--no-verify`, `--amend`, force push, `git stash`.
- 승인 권한은 넓히지 않는다. 응답 정책표에 accept 계열을 넣지 않는다.
- 연속 실패 상한은 `defaults.go` 상수이고 구현에 숫자를 쓰지 않는다.
- `managed_*.go` 에 `syscall.` 참조를 두지 않는다. 신호 상수는 `launch_signals.go` 에만 둔다.
- 정리 호출(launch-pending 롤백, 토큰 디렉터리 삭제)에 시그널 컨텍스트를 쓰지 않는다.
- 로컬에서 `go test ./internal/cli` 전체를 돌리지 않는다. `internal/template` 스위트는 `moai slot acquire --resource internal-template-suite --max-duration 15m` → 실행 → `moai slot release --resource internal-template-suite` 안에서만 돈다(측정 414.986s).
- 문서 한국어 / 코드 주석·커밋·로그 영어.

## §E. Self-Verification (run 단계 완료 보고용 — 명령은 단일 호출형, `go test` 는 §acceptance 머리말의 정리 접두 사용)

- AC 매트릭스: acceptance.md §1 의 명령 그대로. 원문 출력과 트리 HEAD SHA를 progress.md §E.2 에 인용한다(명령·관측·귀속 3요소).
- 스코프 회귀: `go test ./internal/cli -run '^.*(Managed|managed).*$' -count=1` (부분 문자열 `Managed`·`managed` 선택의 앵커형 동치 — 최상위 66개를 리터럴 패턴과 같게 고름을 측정)
- 인접 패키지: `go test ./internal/config -count=1`, `go test ./internal/guardstate -count=1`(사전 적색 1건 고려), `go test ./internal/template -count=1`(임대 안에서)
- Windows: `GOOS=windows GOARCH=amd64 go build ./...`, `GOOS=windows GOARCH=amd64 go vet ./internal/cli/`
- 정적: `gofmt -l internal/cli internal/config`, `go vet ./internal/cli ./internal/config`, `golangci-lint run ./internal/cli/... ./internal/config/...`
- 경계 grep: acceptance.md §5.
- 변이 확인 1건: 상수 값을 일시적으로 바꿔 AC-MH-007 시험이 따라오는지(되돌린 뒤 `git diff --stat` 빈 출력 확인).

## §F. Milestones (실행 순서)

### M1 — RED 기준선 (독립 커밋, 가장 먼저)

- 재현 시험을 **기준 트리 API에 대해 컴파일되는 형태로** 쓴다(새 심볼 없이): `TestManagedCodexTurnSurvivesServerRequest`(가짜 App Server가 응답을 받지 못하면 유한 대기 뒤 `no reply` 를 기록하고 시험이 그 기록으로 실패), `TestManagedCodexServerRequestIDCollision`(프로세스 내 클라이언트 + 임시 서버), `TestManagedDriverIsolatesTurnFailure`(실제 `pumpManagedStreamTurn` 을 쓰는 스크립트 세션), `TestManagedLauncherSignalTeardown`(시험 바이너리 재실행 + 시그널).
- 각 시험을 기준 트리에서 돌려 **옳은 이유로 붉은지** 확인하고 원문을 `.moai/reports/t1409/red-baseline.md` 에 적는다: 명령, 출력 원문(50줄/2KB 넘으면 파일 리다이렉트 + 꼬리), exit 코드, 측정 트리 SHA, 붉은 이유 한 줄. 컴파일 오류나 타임아웃으로 붉은 것은 옳은 이유가 아니다(wrong-reason red) — 시험을 고친다.
- 시험 파일과 `red-baseline.md` 를 **한 커밋**에 담는다. 이 커밋이 수리 커밋들의 조상이어야 한다(REQ-MH-015, AC-MH-013). 같은 커밋에 구현 변경을 섞지 않는다(커밋 그래프가 선후의 유일한 증인이다: verification-claim-integrity §2.3). 커밋 게이트가 실패하는 시험을 담은 커밋을 거부하면 `--no-verify` 를 쓰지 말고 blocker 보고로 돌려준다.
- 커밋 제목 예: `test(SPEC-FACTORY-MANAGED-HARDEN-001): M1 RED baseline for F3 F4 F5 (card t1409)`.

### M2 — F3 서버 요청 응답 (되돌리기 가장 어려운 판단: 정책표)

- 프레임 분류(`id` 원문 보존), 서버 요청 읽기 고루틴 응답, 응답 정책표(한 곳의 표 데이터), 연결 쓰기 뮤텍스(`call()`, `initialized` 알림, 답장 모두 경유).
- GREEN: AC-MH-001, 002, 003, 004(쓰기 경합 부분). M1의 F3 재현 시험이 뒤집힌다.
- 커밋: `fix(SPEC-FACTORY-MANAGED-HARDEN-001): M2 answer server-originated codex requests (card t1409)`.

### M3 — F4 턴 단위 실패 격리

- 턴 단위 표식(`errors.Is` 판별)을 스트림 `result.is_error` 와 Codex `completed` 아닌 종료 두 곳에서만 붙인다. 드라이버: 표식 있으면 로그+계속, 연속 횟수 상한, 성공 시 0으로, 우선 턴은 세지 않음. `defaults.go` 에 `DefaultManagedSessionMaxConsecutiveTurnFailures = 3`(UNMEASURED 표기, 근거 주석).
- GREEN: AC-MH-005, 006, 007, 008. M1의 F4 재현 시험이 뒤집힌다.
- 커밋: `fix(SPEC-FACTORY-MANAGED-HARDEN-001): M3 isolate turn failures with a ceiling (card t1409)`.

### M4 — F5 시그널 정리

- 두 소유자 `Close` 를 한 번만 도는 정리로(동시 호출자는 같은 결과). `driveManagedFactorySession` 에 `ctx` 첫 인자(호출부 10곳) + 한가한/대기 select의 `ctx.Done()` + 중단 오류 우선 확인. `launch_signals.go` 도우미, 소유자 진입의 워처, Codex `Start()` 준비 대기 컨텍스트 파생.
- GREEN: AC-MH-009, 010, 011, 004(Close 경합 부분). M1의 F5 재현 시험이 뒤집힌다.
- 커밋: `fix(SPEC-FACTORY-MANAGED-HARDEN-001): M4 route signals to the managed teardown (card t1409)`.

### M5 — 게이트와 증거 (기계적)

- §E 의 정적·스코프·Windows·인접 패키지 명령을 돌리고 progress.md §E.2/§E.3 에 인용. 변이 확인 1건. 경계 grep.
- 문서·CHANGELOG(AC-MH-012)는 **sync 단계**(manager-docs)의 몫이다: `.moai/docs/factory-managed-session.md` 의 "알려진 한계"에서 세 문장을 지우고 남은 한계를 적으며, 이 SPEC의 CHANGELOG 엔트리를 쓰고 부모 엔트리의 "후속 카드 t1409 대상" 문구를 정정한다. 부모 SPEC 파일은 건드리지 않는다.

## §G. Anti-Patterns

- RED와 수리를 한 커밋에 합치기(선후 증인 상실).
- 컴파일 실패·타임아웃으로 붉은 시험을 RED 기준선으로 인용하기(옳은 이유가 아님).
- 구현에 `3` 이나 `5초` 를 코드 상수로 박기.
- 승인류 응답을 "편의상" accept 로 바꾸기, 또는 `cancel`/`abort` 로 턴을 끊기.
- 시그널 컨텍스트로 정리 호출을 하기(롤백이 즉시 실패).
- 정책표를 분기마다 흩어 쓰기(표 한 곳에서 읽고 시험하는 구조를 깬다).
- 라이브 미관측 사항을 해결로 공시하기(REQ-MH-014).

## §H. Cross-References

- `spec.md` §C REQ-MH-001..016, `acceptance.md` AC-MH-001..015, `design.md` D-1..D-3.
- 부모: `.moai/specs/SPEC-FACTORY-MANAGED-SESSION-001/{spec,design,acceptance}.md` — D-1(배달 전용), D-4(자식 소유), AC-MS-010/014/017.
- `.moai/reports/t1409/intake.md`, `.moai/reports/t1409/t1375-f3-f5-verbatim.md`.
- `.claude/rules/moai/core/verification-claim-integrity.md` §2.3(순서 증인), `.claude/rules/moai/development/verification-completeness.md` §2(RED-now/green path), `.claude/rules/local/gitflow-lane-protocol.md` §4·§8(push·부하).

## §I. 미해결 사항

운영자에게 되물어야 할 미해결 항목은 없다(해결 표식 0건). 남은 불확실성은 질문이 아니라 관측 공백(Gaps)이며 §C 와 design.md D-1 "미관측"에 있다.
