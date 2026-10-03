---
id: SPEC-FACTORY-MANAGED-CARD-CHILD-001
title: "plan.md — 구현 계획"
version: "0.1.0"
created: 2026-10-03
updated: 2026-10-03
author: GOOS (manager-spec)
tier: M
---

# plan.md — 구현 계획

> 상태축 없음. 마일스톤은 우선순위와 선후 관계로만 적는다(시간 추정 없음). 검토 우선순위는 §A, 실행 순서는 §F 의 M0..M4 다.

## §A. Context

- **카드 / 트리**: t1440(lane-17), 워크트리 `.claude/worktrees/t1440`, 브랜치 `WT-codex-card-managed-path`, 기준 로컬 develop `2b9e4a4d0`.
- **부모·선행**: SPEC-FACTORY-MANAGED-SESSION-001(`completed`), SPEC-FACTORY-MANAGED-HARDEN-001(`completed`) — 둘 다 본문·frontmatter 불변. `depends_on` 이 연결의 전부다(둘 다 `completed` 라 depends_on 사전 점검이 통과한다).
- **아티팩트**: `.moai/specs/SPEC-FACTORY-MANAGED-CARD-CHILD-001/{spec,plan,acceptance,design,progress}.md`. REQ 12 / AC 13.
- **Tier 판단**: **Tier M.** 측정 아님·추정: 비테스트 변경은 `codex_launcher.go` 의 루프·카드 자식 함수(현재 `:915-1061`, 약 150줄 구간)의 분기와 조립 약 40–60줄, 새 파일 `managed_operator_input.go` 약 80–120줄. 시험은 새 파일 두 개에 약 400–600줄(최상위 10개, 하위 케이스 약 20개; 가짜 App Server 를 쓰는 시험은 기존 헬퍼 재사용). 합 약 550–800줄로 M 범위(300–1000줄) 안. 파일 수는 비테스트 2·시험 3(새 2 + 기존 `managed_optin_test.go` 수정 1)·문서 2(+CHANGELOG)로 5–15 범위 안. design.md 가 있는 것은 stdin 소유 설계(D-6)와 환경 차이(D-4) 때문이다 — HARDEN-001 이 Tier M 에 design.md 를 둔 선례를 따른다(Artifact Statelessness 규칙은 design.md 에 `status:` 만 금지).
- **수정 대상(기대)**: 아래 §E 의 목록.
- **PRESERVE(수정 금지)**: `.moai/specs/SPEC-FACTORY-MANAGED-SESSION-001/**`, `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/**`, `internal/factorymsg/store.go`, `internal/cli/managed_codex_factory.go`, `internal/cli/managed_factory_session.go`(소유자·드라이버), `factory_launch_pending.go` 의 술어, `codex_direct_posix.go`/`codex_direct_windows.go`, `internal/config/envkeys.go`(새 키 없음).
- **검토 우선순위(되돌리기 어려운 판단 순)**:
  1. design.md D-4 환경 차이(run id 가 카드 자식 환경에 실린다) 와 D-3 락·claim 불실행 — 운영 계약이 걸린다.
  2. design.md D-6 stdin 펌프 — 새 동시성 코드.
  3. design.md D-1 분기 위치, D-5 끝 조건 — 기계적이다.

## §B. Known Issues (관련 범주만)

- **B1 크로스플랫폼**: Windows 크로스 빌드는 회귀 가드다(AC-CC-011). 새 파일에 `syscall` 참조·OS 빌드 태그를 더하지 않는다.
- **B3 서브에이전트 경계**: `internal/cli` 에 `AskUserQuestion` 호출 금지.
- **B4 frontmatter**: `created`/`updated`/`tags` 정규 이름(준수). `plan.md`·`acceptance.md`·`design.md` 는 `status:` 를 갖지 않는다.
- **B5 CI 3층**: spec-lint, golangci-lint(v2.1.6), 테스트가 따로 적색일 수 있다. 사전 적색과 신규 결함을 구별한다.
- **B6 헤딩**: spec.md §F 는 `### Out of Scope — <topic>` H3 + `-` 불릿(작성 완료).
- **B8 트리 위생**: 커밋은 명시 pathspec. `.moai/state/` 무수정. 측정용 임시 파일은 커밋 전 삭제.
- **B9 커밋**: 카드 id `t1440` 을 모든 커밋 메시지에 넣는다(병합 커밋 포함). 레인은 push 하지 않고 병합 SHA 를 리더에 보고한다(gitflow-lane-protocol §4).
- **레인 환경 오염**: 레인 세션의 `MOAI_*` 환경은 이 영역 시험을 가짜 적색으로 만든다 — `go test` 는 항상 acceptance.md 헤더의 정리 복합형으로 돌린다(`TestManagedSwitchDoesNotReachCodexLaneLoop` 가 `sdScrubLauncherEnv` 로 스스로 청소하는 이유다).
- **t1408(TUI attach) 인접**: 같은 `codex_launcher.go` 의 분기 근처와 `managed_codex_factory.go` 를 건드릴 수 있다. 이 카드는 소유자 파일을 건드리지 않고, `codex_launcher.go` 에서는 루프·카드 자식 함수와 이음새 기본 구현(`:252-259`)만 바꾼다 — 충돌은 나중에 병합하는 쪽이 푼다.
- **t1459(시그널·Start/Close) 인접**: 소유자 진입 `runManagedFactoryCodex`(`managed_codex_factory.go:787-845`)의 정리 순서를 t1459 가 바꾼다. 이 카드는 그 함수를 호출만 하고 시그니처를 바꾸지 않는다. 문서 "알려진 한계" 두 불릿(헤드리스·시그널)은 두 카드가 같은 문단을 만진다 — sync 단계에서 병합 순서에 맞춰 다시 읽는다.
- **t1410 인접**: 소유자 `Start()` 의 토큰 디렉터리 정리·핸드셰이크 예산은 이 카드 밖이다. 단 카드 자식 루프에서는 `Start` 실패가 카드 단위로 반복될 수 있어 Q3 로 올렸다.

## §C. Pre-flight — 측정 기준선과 run 단계 측정 항목

### C.1 이 plan 실행의 측정(트리 `2b9e4a4d0`)

| 항목 | 명령 | 관측 |
|---|---|---|
| 카드 트리 | `git rev-parse --short HEAD` | `2b9e4a4d0` |
| 관리 시험 수 | acceptance.md §1.1 AC-CC-013 블록 첫 줄(`-list` 뒤 `grep -c '^Test'`) | `79` |
| 갭을 고정하는 현행 시험 | `go test ./internal/cli -run '^TestManagedSwitchDoesNotReachCodexLaneLoop$' -count=1 -v` (정리 접두) | `--- PASS … (10.99s)`, `ok … 12.381s`, exit 0 |
| 새 시험 이름 선택 | acceptance.md §2 의 R-2 블록 | `ok … 1.131s`, 이름 0개 |
| 문서 옛 문장 | acceptance.md §1.2 의 `grep -c` | 각 `1` |
| 템플릿 미러 | `ls internal/template/templates/.moai/docs/ | grep -c managed` | `0` — 이 문서에 템플릿 미러가 없다 |

### C.2 run 단계 M0 가 **먼저** 측정할 항목 (읽기만 한 전제들)

| id | 전제 | 측정 방법 | 결과가 바꾸는 것 |
|---|---|---|---|
| P-1 | **H-1**: 직접 exec 문 기본 구현이 POSIX 에서 `syscall.Exec` 로 프로세스를 교체해(`codex_direct_posix.go:53`) 현행 레인 루프가 첫 카드 자식에서 끝난다 | 임시 시험(커밋하지 않음)으로 실제 기본 이음새를 `exit 0` 하는 가짜 `codex` 스크립트에 연결하고 카드 둘을 큐에 둔 `moai codex -f lane` 을 서브프로세스로 돌려 두 번째 카드의 lease 가 일어나는지 본다. 양성 대조: 같은 시험에서 이음새를 스텁으로 바꾸면 둘 다 lease 된다 | 참이면 Q4 로 리더에 보고(이 SPEC 범위 밖 결함 후보). 거짓이면 design.md D-2 의 "함의" 문단을 정정한다. 어느 쪽이든 REQ·AC 는 바뀌지 않는다 |
| P-2 | 루프의 임대 루트와 소유자의 브로커 루트가 같다: 루프는 `factoryCardRoot()`=`resolveTodoQueueRoot()`(`factory_card.go:39`, `todo.go:74-76`)를, 소유자는 `launchProjectRoot()`=`resolveProjectDir()`(`managed_codex_factory.go:788`, `launcher_blockcap_infinite.go:138-140`)를 쓴다 | `kanban.ResolveTodoQueueRoot(base)`(`todo_root.go`)가 `base` 와 달라지는 경우(임시 원점 가드·홈 큐)를 읽고, 루프 시험 픽스처 둘(일반 루트, 임시 디렉터리 루트)에서 두 값을 찍어 비교한다 | 달라질 수 있으면 소유자 이음새에 루트를 싣는 방법을 design.md 에 추가(소유자 시그니처는 건드리지 않고 `CLAUDE_PROJECT_DIR` 정렬 같은 우회 먼저). 같으면 AC-CC-007 이 그대로 증거 |
| P-3 | 같은 런처 PID·같은 레인 라벨로 두 번째 launch-pending 등록이 첫 카드의 바인딩된 행을 교체한다(읽은 근거: `store.go:485-491` — 같은 소유 PID·process_start 면 live-owner 거부가 걸리지 않는다) | AC-CC-007 의 `sequential_cards_rebind` 를 먼저 쓰고 돌린다 | 실패하면 `store.go` 를 고치지 않고 blocker 보고(REQ-MS-008 의 선례) |
| P-4 | 소유자가 App Server 명령행 `-c developer_instructions=…` 를 받아도 스레드에 적용되는지는 관측되지 않았다 | 가짜 서버로는 관측 불가 — 실제 codex 가 필요하다. 이 plan 은 적용 여부를 AC 로 두지 않고 문서의 미관측 한계로만 남긴다 | 문서 문장 한 줄(§1.2 (6)) |

## §D. 열린 질문 (사용자 확인 대상 — run 진입 전 오케스트레이터가 해소)

다음 세 표지는 Implementation Kickoff 전에 해소돼야 한다. 각각 이 SPEC 의 **기본값**이 적혀 있어, 답이 없으면 기본값으로 진행할 수 있다.

- **[NEEDS CLARIFICATION: Q1 카드 시작 프롬프트]** 관리 소유자의 드라이버는 우선 턴(`managedPrimingPrompt`, "준비 완료라고 한 줄로 답해. 아직 작업은 시작하지 마.", `managed_factory_session.go:55`) 뒤 유휴로 들어가 브로커 메시지나 운영자 입력을 기다린다. 따라서 관리 카드 자식은 **스스로 카드 작업을 시작하지 않는다**(헤드리스라 운영자도 화면을 보지 못한다). 현행 대화형 카드 자식도 운영자 입력 전에는 시작하지 않는다는 점에서 같지만, 관리 경로에서는 그 입력 채널이 눈에 안 보인다. **기본값**: 카드 시작 프롬프트를 주입하지 않는다(범위 밖 — spec.md §F). 필요하면 별도 카드.
- **[NEEDS CLARIFICATION: Q2 비대화형 stdin]** 드라이버는 stdin EOF 에서 끝나지 않는다(`managed_factory_session.go:315-319`). stdin 이 닫힌/없는 환경(CI, `</dev/null`)에서 관리 카드 세션은 치명 오류가 날 때까지 루프를 붙잡는다. **기본값**: 드라이버 끝 조건을 바꾸지 않고 문서에 한계로 적는다. 대안은 드라이버가 EOF 를 세션 종료로 읽도록 바꾸는 것인데, 그러면 플레인 관리 경로의 현행 동작(EOF 이후에도 브로커 배달 계속)이 바뀌므로 이 SPEC 의 소유자 불변 제약에 어긋난다 — 필요하면 별도 카드.
- **[NEEDS CLARIFICATION: Q3 연속 시작 실패]** 현행 직접 문과 같이 루프는 자식 시작 실패 뒤 다음 카드를 lease 한다. 관리 경로는 시작 실패 요인이 더 많다(`app-server` 하위 명령 부재, 10초 핸드셰이크 시간 초과). 큐가 비기까지 카드가 하나씩 lease 되어 전부 lease 만료까지 묶일 수 있다. **기본값**: 현행과 같다(상한 없음, spec.md §F). 연속 실패 상한은 별도 카드.
- **Q4 (질문이 아닌 보고 항목)**: P-1 이 H-1 을 확인하면 POSIX 현행 레인 루프가 첫 카드 뒤 이어지지 않는다는 뜻이다. 이 SPEC 의 관리 경로는 그 문제를 가진 경로가 아니다(런처가 부모로 남는다). 직접 문 쪽은 이 SPEC 이 건드리지 않는다.

## §E. 수정 파일 (run 단계가 닿을 것으로 기대하는 목록)

| 파일 | 변경 | 이유 |
|---|---|---|
| `internal/cli/codex_launcher.go` | 수정 — `runCodexFactoryLane`/`launchCodexCardSession` 의 분기·조립(`:915-1061`), 이음새 기본 구현 `managedFactoryCodexLaunchFunc`(`:252-259`)의 stdin, `:1180-1181` 주석 정정 | D-1, D-4, D-6 |
| `internal/cli/managed_operator_input.go` | **새 파일** — 프로세스 수명 입력 펌프와 닫을 수 있는 어댑터 | D-6 |
| `internal/cli/managed_operator_input_test.go` | **새 파일** — `TestManagedOperatorInputPumpDetachesEndedSession` | AC-CC-010 단위 |
| `internal/cli/managed_card_child_test.go` | **새 파일** — `TestManagedCardChild*` 9개 | AC-CC-001..010 |
| `internal/cli/managed_optin_test.go` | 수정 — `TestManagedSwitchDoesNotReachCodexLaneLoop`(`:190-224`)를 AC-CC-001 시험으로 개명·반전하고 파일 머리 주석(`:9-11`)·`:187-189` 주석을 정정. 스위치가 꺼진 변형은 AC-CC-002 시험이 운반 | 갭 고정 시험 대체 |
| `.moai/docs/factory-managed-session.md` | 수정(sync) — "켜도 닿지 않는 범위" 26-27행·30-31행, 알려진 한계의 새 항목 | AC-CC-012 |
| `CHANGELOG.md` | 수정(sync) | AC-CC-012 |
| `.moai/specs/SPEC-FACTORY-MANAGED-CARD-CHILD-001/progress.md` | run/sync 증거 | |

**수정하지 않는 곳**: 위 §A PRESERVE 목록. 특히 `managed_codex_factory_test.go` 의 가짜 App Server 는 고치지 않고 기존 헬퍼(`fakeAppServerScript`, `fakeAppServerRoleEnv`/`fakeAppServerLogEnv`, `activateManagedRun`, `sdRecordLeaderRun`, `withManagedCodexWiring`, `runCodexCmd`, `factoryLaneEnv`, `managedOptIn`)만 재사용한다 — t1408·t1459 가 같은 가짜를 건드릴 수 있어 충돌 면적을 줄인다.

## §F. 마일스톤 (되돌리기 어려운 판단 순 — 위쪽이 바뀔 가능성이 크다)

### M0 — 읽기 전제 측정 (우선순위 High)

§C.2 의 P-1..P-3 을 측정하고 결과를 progress.md §E.2 에 원문으로 기록한다. P-2 가 "달라질 수 있다" 로 나오면 M2 전에 design.md 를 보강하도록 blocker 보고(소유자 시그니처 변경 금지).

### M1 — RED (우선순위 High)

1. 새 시험 파일 둘을 쓴다: AC-CC-001..010 의 최상위 10개. RED 원문을 `.moai/specs/SPEC-FACTORY-MANAGED-CARD-CHILD-001/red-baseline.md`(추적됨)에 둔다 — `.moai/reports/` 는 gitignore 대상이다.
2. 한 RED 커밋에 시험 파일과 `red-baseline.md` 를 함께 넣고, 이 커밋이 모든 수리 커밋의 조상이 되게 한다(acceptance.md §2 RED-now 쌍의 순서 증거 — 커밋 그래프만이 선후를 증명한다).
3. AC-CC-002·003 은 기준 트리에서 초록이어야 한다 — 양쪽 팔(전·후)을 측정한다.

### M2 — 분기와 런치 형태 (우선순위 High, 사용자 흐름)

D-1·D-4: 카드 자식 호출 지점에서 `factoryManagedRequested && factoryLaunchEnabled` 로 분기, 관리 분기는 `-C` 없이 `[bin, localArgs...]`, `dir=wt`, env = `codexCardLaunchEnv` + `MOAI_KANBAN_ID`. 앵커 락·claim 스탬프 호출을 더하지 않는다(D-3). 오류는 현행 줄 형태 그대로 stderr(D-5). 직접 exec 문 쪽 줄은 바꾸지 않는다. 대상 AC: 001, 002, 004, 005, 006, 009.

### M3 — 운영자 입력 펌프 (우선순위 High, 새 동시성 코드)

D-6: 새 파일에 펌프·어댑터를 두고 기본 소유자 이음새가 이를 쓰게 한다. `-race` 로 단위 시험과 루프 수준 시험을 돌린다. 대상 AC: 007, 008, 010.

### M4 — 게이트와 문서 (우선순위 Medium, 기계적)

AC-CC-011·013 게이트. 문서·CHANGELOG 는 sync 단계의 manager-docs 가 쓴다(AC-CC-012; 이 plan 은 문장 앵커만 정한다). 변이 mu1..mu8 중 적용 가능한 것을 실제로 적용해 붉어짐을 관측한다.

## §G. 위험

| 위험 | 완화 |
|---|---|
| 같은 런처 PID 의 연속 등록이 막힘(P-3) | AC-CC-007 이 실제 브로커 저장소로 직접 증거. 막히면 store 를 고치지 않고 blocker |
| stdin 펌프가 `-race`·종료 시 새 고루틴 누수 | 어댑터 `Close` 가 막힌 `Read` 를 푼다는 AC-CC-010 단위 케이스. 펌프 고루틴은 프로세스 수명이라 누수가 아니다(문서화) |
| 환경 차이(run id)가 운영 계약을 바꾼다고 읽힘 | design.md D-4 에 기록, 문서(AC-CC-012 (3))에 명시 |
| t1408/t1459 와의 `codex_launcher.go`·문서 충돌 | §B 인접 항목. 마일스톤별로 변경을 몰아 커밋 |
| 가짜 App Server 시험의 POSIX 한정 | 크로스 빌드가 Windows 증거(부모 선례) |

## §H. 교차 참조

spec.md §C·§D·§F, acceptance.md §1·§2, design.md D-1..D-6, `.claude/rules/moai/development/verification-completeness.md`(RED-now 쌍·변이 탐침), `.claude/rules/moai/workflow/kanban-dispatch.md`(레인·추적성 운반체).
