---
title: "Todo DB 통일 · 착수 및 구현 인계 보고"
lang: ko
---

# Claim — 현재까지 완료한 작업

**t648을 발행·착수 상태로 기록하고, 새 작업 공간에 선행 Todo 감사 변경을 통합했다. DB 통일·완료 자동 확정·Graph·시각화의 신규 구현은 아직 하지 않았다.**

운영자가 확정한 범위는 카드 및 실행의 정확성에 필요한 정보를 프로젝트 Todo DB로 통일하는 것이다. heartbeat와 상세 로그는 별도 저장할 수 있다. 네 단계는 독립적인 상태 저장소가 아니라 하나의 공통 모델을 순차 확장하므로, 이번에는 t648 한 장 아래 단계별로 추적한다.

| 순서 | 우선순위 | 구현 범위 | 현재 상태 |
|---|---|---|---|
| 1 | High | 카드 identity, 실행 소유권, 완료 기록, 중단·중복·재개 복구를 Todo DB에서 처리 | 조사, 미구현 |
| 2 | High | 칸반·팩토리·sync의 공통 완료 호출과 단일 상태 조회 | 조사, 미구현 |
| 3 | Medium | 카드·실행·SPEC·커밋·검증·완료 기록의 의미 있는 관계 | 조사, 미구현 |
| 4 | Low | 기존 Todo 화면에서 의존성·완료 근거·복구 대기 사유 표시 | 연결 위치 확인, 미구현 |

## 코드 조사로 확인한 저장 경계

아래는 소스 구조 관측이며, 신규 설계가 동작한다는 검증 결과가 아니다.

| 영역 | 관측한 소스 | 설계에 반영할 사항 |
|---|---|---|
| Factory 실행 시작 | `internal/kanban/factory_runtime.go` → `homestate.RecordRun` | 실행의 정확성에 필요한 기록은 Todo 공통 서비스로 이동 |
| Factory 카드 상태 | `internal/cli/todo.go` → `recordFactoryCardState` → `homestate.RecordCard` | 별도 Factory 카드 상태를 완료 판단의 기준으로 유지하지 않음 |
| 별도 스키마 | `internal/homestate/factory.go`와 `internal/kanban/backlog_sqlite.go` | 두 DB의 `meta.schema_version` 계약이 다르므로 경로만 바꾸지 않음 |
| Todo 변경 잠금 | `BacklogStore.Mutate`와 `writeRecordArchive` | 모든 writer가 공통 잠금을 사용하고 완료 관련 변경을 같은 SQL transaction에 포함 |
| 실행 소유권 | `internal/kanban/factory_slots.go`의 슬롯 선점 | heartbeat와 구별하여 슬롯 소유권도 정확성 데이터로 취급 |
| 재개 권리 | `internal/homestate/handoff.go`의 claim token·만료·소비 처리 | 오래된 실행이 재개 권리를 행사하지 않도록 공통 실행 세대와 연결 |
| sync 체크포인트 | `internal/session/checkpoint.go`의 `SyncCheckpoint` | SPEC/PR/DocsSynced만으로 카드 완료를 확정하지 않음 |
| 웹 읽기 | `internal/web/todo_queue_read.go`의 `LoadPure` | 조회가 migration·recovery·완료 처리를 실행하지 않도록 유지 |
| 기존 Graph | `internal/graph/graph.go`, `reader.go` | JSONL 코드 그래프와 Todo의 정식 관계를 구분하고 기존 탐색 함수를 필요한 곳에 재사용 |

Factory worker 정보에는 슬롯 선점과 heartbeat가 함께 존재한다. 따라서 workers 테이블 전체를 단순 관측 로그로 남긴다는 해석은 승인된 범위에 맞지 않는다. 슬롯 소유권과 실행 세대는 Todo의 공통 상태에 연결하고, heartbeat는 실행 생존 여부의 참고 자료로 분리하는 방향으로 설계해야 한다.

## 최소 변경 설계안 — 아직 구현되지 않음

1. 카드의 불변 UUID와 표시용 `tNN`, 실행 세대를 구분한다. 기존 live/archive 기록도 식별자를 유지한다.
2. 실행 선점은 owner token과 generation으로 보호한다. 완료 취소 또는 실행 인수 후 이전 token으로 온 완료 요청을 거절한다.
3. 증거 수집·Git 조회는 쓰기 잠금 밖에서 수행한다. 확정 시 현재 실행과 증거 대상을 다시 대조한다.
4. 하나의 Todo 쓰기 transaction에서 완료 기록·카드 archive·실행 완료를 확정한다. 같은 요청의 재실행은 기존 완료 기록을 반환한다.
5. 외부 알림 등 실제 DB 밖으로 전달할 일이 있을 때만 재시도 기록을 둔다. Factory 카드 상태를 복제하기 위한 별도 동기화는 만들지 않는다.
6. Graph 관계에는 종류·출처·대상 식별자·근거 유효성을 둔다. 추론한 관계만으로 pick/drop/완료를 수행하지 않는다.
7. 기존 `/todo` 화면의 읽기 전용 원칙, 텍스트 escaping, 접근성을 유지한다. 완료된 카드의 과거 상태 `picked`와 현재 완료 여부를 구분하여 표시한다.

SQLite는 같은 DB에서 여러 reader와 하나의 writer를 허용한다. WAL을 사용하는 별도 DB들을 단순히 연결한다고 전체 원자성이 보장되는 것은 아니다. 이는 정확성 데이터를 한 DB에 모으는 설계의 근거이며, 이번 작업의 성능 실측은 아니다. [SQLite transaction 공식 문서](https://www.sqlite.org/lang_transaction.html), [SQLite WAL 공식 문서](https://www.sqlite.org/wal.html)

## 새 구현에서 필요한 검증 — 실행 전 계획

- 완료 기록 insert 실패 시 카드·실행 변경도 함께 rollback되는지 확인한다.
- commit 직후 응답 전 종료 후 재시도해도 완료 기록이 한 개만 남는지 확인한다.
- 동일 카드의 동시 선점과 동시 완료를 두 모드에서 경쟁시킨다.
- undone 후 이전 실행 token으로 완료를 요청해도 live 카드가 유지되는지 확인한다.
- 같은 표시 ID가 다른 프로젝트에 있을 때 서로 영향을 주지 않는지 확인한다.
- schema backfill 중단 후 구 schema 또는 완전한 새 schema만 관찰되는지 확인한다.
- active WAL을 읽을 때 카드·완료 기록·실행 정보가 같은 스냅숏인지 확인한다.
- history·Graph·웹 조회가 저장된 카드나 schema를 변경하지 않는지 확인한다.
- Graph의 순환·깊이 제한·archive 이력·근거 누락을 검증한다.
- 실제 화면에서 키보드 탐색·작은 화면·오류와 빈 결과 구분을 확인한다.

# Evidence — 실행 명령과 실제 출력

## 카드

등록 명령은 `moai todo add '<승인 범위와 완료 조건 전문>' --pick`이었다. 실제 출력:

```text
picked t648 [TODO-UNIFIED-20260911 · 운영자 범위 확정 및 진행 ...
```

완료 처리하지 않고 아래 명령으로 다시 읽었다.

```bash
moai todo list --json | jq '{items:[.items[] | select(.id=="t648") | {id,state,text}]}'
```

출력 중 상태 필드의 원문:

```json
"id": "t648",
"state": "picked"
```

## 선행 변경 통합

전문 Git 담당이 새 작업 공간에서 실행한 명령:

```bash
git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/todo-unified merge --no-ff --no-commit 2bae4f1cc1892af61632b0c672056bdff7198932
```

실제 출력:

```text
Auto-merging internal/cli/mcp_build_identity_test.go
CONFLICT (content): Merge conflict in internal/cli/mcp_build_identity_test.go
Auto-merging internal/template/catalog.yaml
Automatic merge failed; fix conflicts and then commit the result.
```

충돌은 실제 소스 위치에 맞춰 `todo_landed.go:216`에서 `:217`로 고친 테스트 기대값과 주석이었다. home-state 위치 `245/253`은 유지했다. 이후 담당이 실행한 회귀 검증:

```bash
unset MOAI_HOME MOAI_CONFIG_DIR MOAI_PROJECT_ROOT CLAUDE_PROJECT_DIR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_COMMON_DIR &&
go test ./internal/cli -run '^TestAuditLagUsesBinlagSeam$' -count=1 -timeout=120s
```

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	2.566s
```

통합 커밋 출력:

```text
[WT-todo-unified 8ab93ee20] chore(todo): integrate audit baseline (card t648)
```

루트 담당이 통합 후 다시 실행한 `git rev-parse HEAD` 출력:

```text
8ab93ee20d3adb4750552067f29fcf96977fb724
```

통합 직후 `git status --short`는 출력이 없었다. 이 보고서를 추가하기 전 상태이며, 보고서 작성 후에도 clean이었다는 주장이 아니다.

## 실행 환경의 차단

전문 설계 담당 생성 두 차례의 출력:

```text
collab spawn failed: agent thread limit reached
```

이전 연결 담당을 재개하려 한 한 차례의 출력:

```text
collab tool failed: agent thread limit reached
```

이는 MoAI 애플리케이션의 오류가 아니라 현재 대화의 에이전트 도구 제한이다. 구현 담당·독립 감사 담당을 추가로 실행하지 못했으며, 제한 해제 방법은 관측하지 못했다. 에이전트가 완료 상태인 것을 확인한 뒤에도 새 생성은 거부되었다.

# Baseline-attribution — 이번 관측의 기준

- 일자: 2026-09-11.
- 새 작업 공간은 원격 기본 브랜치 `origin/develop`의 `b155c95f9942206d820ad3b7b4fc1cb132f41bb4`에서 생성했다.
- 선행 감사 브랜치 `2bae4f1cc1892af61632b0c672056bdff7198932`를 통합한 기준은 `8ab93ee20d3adb4750552067f29fcf96977fb724`다.
- 저장소 연구 담당의 일부 소스 읽기는 해당 병합이 진행 중인 트리에서 수행했다. 테스트 기준과 혼동하지 않는다.
- primary checkout의 기존 dirty 변경은 보존했다.
- 과거 t647 성능 수치를 새 설계의 측정값으로 사용하지 않았다.

# Gaps — 아직 하지 않은 일

- 신규 SPEC 작성과 독립 plan audit, DB 통일 구현, 자동 완료 연결, Graph와 화면 구현.
- 위 신규 검증 계획의 실행, 전체 테스트와 원격 CI, 브라우저 E2E.
- 운영 홈 DB 이전, 설치 바이너리 교체, 원격 push/PR/병합/배포.
- 실제 SQLite `synchronous` 설정 및 전원 손실 내구성 시험. 프로세스 종료 복구와 전원 장애 내구성은 다르다.
- 과거 완료 누락 사례 전체의 원인 확정. 소스 관측만으로 모든 운영 사례의 원인을 단정하지 않는다.

# Residual-risk — 후속 작업의 주의점

한 DB로 합쳐도 writer 경쟁은 남는다. 기존 전체 레코드 저장과 새 SQL writer의 잠금 규약이 다르면 갱신을 잃을 수 있으므로 반드시 공통 변경 경계를 사용해야 한다. 기존 schema 불변성 테스트는 새 버전을 지원하도록 바꾸되 조회 전후 불변성 검증 자체는 제거하지 않는다. 새 schema와 옛 바이너리의 혼용, 활성 세션이 있는 운영 데이터 전환은 별도의 안전 조건이 필요하다.

**재개 기록:** `.moai/reports/t648/dispatch.md`. 현재 카드의 승인은 유지되며, t648을 중복 등록하거나 선행 통합을 다시 실행하지 않는다. 작업 공간은 미전달 변경을 보유하므로 삭제하지 않는다.

## 후속 구현용 생산 경로 목록

조사 담당이 최종 `8ab93ee20`에서 재확인한 연결이다. 줄 번호는 후속 편집 전에 다시 확인한다.

| 소비자 | 현재 호출 | 통일 시 놓치지 않을 부분 |
|---|---|---|
| `internal/cli/cc.go:172`, `glm.go:228` | `RecordFactoryRunStart` | 실행 시작 실패 전파 |
| `internal/cli/todo.go:693,710,871,917,924` | archive/pick/unpick 후 별도 Factory 기록 | 한 번의 공통 상태 변경 |
| `internal/cli/factory.go:341` | `ClaimFactoryWorkerName` | 슬롯 소유권 |
| `internal/hook/session_start_factory.go:150` | `FactoryFreeSlots` | 실제 배차 입력 |
| `internal/web/factory_lanes.go:63` | worker PID와 세션 기록 결합 | 공통 claim을 표시 기준으로 사용 |
| `internal/homestate/runtime_census.go:99` | workers PID 활성 조사 | 운영 이전 안전 검사도 새 저장소 반영 |
| `internal/hook/handoff/pending.go:86,138,164,227,278,288` | save/clear/read/claim/expire/finish | 재개 claim과 ack의 정확성 |
| `internal/hook/handoff_inject.go:63,114,136` | read→claim→finish | 주입 성공과 소비 확정 연결 |
| `internal/cli/factory_handoff_recover.go:25` | legacy resume 복구 | 이전 token·process identity·활성 조사 보존 |
| `internal/session/store.go:84,186` | checkpoint와 중단 전이 탐지 | 완료 기록과 연결하되 JSON 존재만으로 확정 금지 |

현재 import 방향은 `cli/hook/web → kanban → homestate`이다. `homestate → kanban`를 추가하지 않는다. `BacklogPathForRoot`와 `BacklogStore.EnginePath`가 실제 큐 위치를 정하므로 home의 예정 경로만 직접 열어 별도 빈 큐를 만드는 방식을 피한다.

기존 resume 기록에는 card UUID/run revision이 없으므로 임의의 카드로 연결하지 않는다. 이전 dry-run에서 충돌·고아·연결 불가 기록을 별도로 보고하고 원본을 보존한다. Factory의 completed 행만으로 기존 Todo 카드를 일괄 완료 처리하지 않는다. memory handoff는 카드 완료의 근거로 사용하지 않는 한 별도 유지할 수 있다.
