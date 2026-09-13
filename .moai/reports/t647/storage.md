# t647 저장 계층 개선 증거

## Claim

Todo 전수 감사의 T02–T06, T17–T19, T21을 구현하고 저장 계층 회귀 검증을 통과했다. 이 문서는 저장 계층의 구현·검증 증거이며 전체 카드 완료, 커밋, PR, 원격 반영을 주장하지 않는다.

| 항목 | 구현 | 검증 |
|---|---|---|
| T02 | record 전체를 하나의 deferred read transaction으로 읽음 | 300장 live/archive 이동과 1,000회 조회가 일관된 총량 유지; writer 실제 커밋 수 검사 |
| T03 | query-only reader, 구버전 archive/landing 읽기 호환, 생성자 무변경, archive vouch 공통 reader | 테이블 보존, DB 바이트·파일 목록 보존, SQL 쓰기 거부, 읽기 전용 DB, WAL 가시성, 없는 DB 생성 금지 |
| T04 | 기존 DB version을 writable open 전에 검사 | 미래 version 테이블 수 1→1, 명시적 거부 |
| T05 | 아직 archive인 상대 카드에 관계를 보존하고 양쪽 복원 후 live로 반환 | 부분 복원 시 dangling finding 없음; 마지막 복원 후 관계 1건 회복 |
| T06 | 실제 quarantine 성공 후에만 pending 표지 제거 | 장애물 제거 후 후속 Load에서 JSON 격리 성공 |
| T17 | Git metadata 부모 대신 실제 primary worktree 해석 | separate-git-dir 독립 저장소 두 개가 서로 다른 root로 해석 |
| T18 | staging DB에 기록·parity 검증·close 후 canonical 이름으로 게시 | 기존 실패/원본 보존/이관 parity 회귀; 옛 불완전 DB+JSON은 pure·adopting 모두 명시적 오류 |
| T19 | 변경되지 않은 archive 쓰기 생략 | 카드 1장 추가 때 archive 1,000장의 행 쓰기 2,000→0 |
| T21 | source lock 아래 이전 표지를 남겨 대기 중이던 옛 writer 거부 | 신규 게시·이미 게시된 target 모두 늦은 writer 거부; 게시 전 중단 후 재시작 복구 |

T05는 과거에 이미 삭제되어 어디에도 없는 endpoint를 참조하는 기록을 새로 삭제하지 않는다. 기존 복원 계약을 보존한다. 상대 카드가 archive에 남아 있는 경우에만 관계를 그 카드에 보관한다.

## Baseline-attribution

- 원래 반례 측정: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop`, `ee99507fbe3b4a22c6a0a74815723d222dfdc04d`.
- 구현 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/todo-audit`, `WT-todo-audit`, 같은 HEAD 위의 미커밋 변경.
- 최초 반례 fixture: `/tmp/moai-todo-storage-audit.TI3Io7/audit_test.go` 및 `overlay.json`.
- 커밋된 운영 데이터는 사용하지 않았다. 데이터베이스와 Git fixture는 `t.TempDir()` 안에서 만들었고 실행 환경의 MoAI/Git 경로 override를 제거했다.
- 동시 작업자가 수정하는 CLI·web·문서는 이 담당자가 변경하지 않았다.
- 조사 파일 집합: `internal/kanban`의 backlog, todo_root, state_dir, temp_origin 이름을 포함한 파일 35개(새 회귀 테스트 포함). 관련 구현 `internal/homestate/paths.go`, `internal/core/git/checkout.go`도 조사했다. 모든 테스트 파일의 모든 줄을 별도로 검토했다는 뜻은 아니다.

## Evidence — RED

구현 전에 다음 명령으로 추가 회귀 테스트의 실패를 확인했다.

```sh
unset MOAI_HOME MOAI_CONFIG_DIR MOAI_PROJECT_ROOT GIT_DIR GIT_WORK_TREE GIT_COMMON_DIR && go test ./internal/kanban -run '^TestAudit' -v -count=1 -timeout=60s
```

실패 출력에서 각 반례의 관측 줄을 그대로 발췌했다.

```text
LoadPure archive tables: before=false after=true
two independent git repositories resolve to same todo root
reopened live count=0; retained legacy={"version":1,"last_seq":1,"items":[{"id":"t1","text":"one","state":"queued"}]}
one Add with 1000 unchanged archived cards executes 2000 archive row writes
late Add returned id=t2; canonical count=1
unknown-version store mutated before refusal
restored live finding references archived card
read 2 mixed commits: live=300 archived=300 total=600 want=300
after obstruction removed and Load retried: legacy JSON still present=true
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.966s
FAIL
```

후속 독립 검토와 추가 RED:

```text
stale writer err=<nil>; expected relocated refusal after target already existed
pure interrupted read error=<nil>, want explicit refusal
pure reader sidecars changed: before=[backlog.db backlog.lock] after=[backlog.db backlog.db-shm backlog.db-wal backlog.lock]
```

생성자 검증도 수정 전에 실패했다. `TestAuditConstructorPreservesLegacyLock`은 `constructor removed legacy lock: stat .../backlog.json.lock: no such file or directory`를 출력했다. 해당 명령은 `go test ./internal/kanban -run '^TestAuditConstructor' -v -count=1 -timeout=60s`였다.

## Evidence — GREEN

최종 저장 계층 선택 실행:

```sh
unset MOAI_HOME MOAI_CONFIG_DIR MOAI_PROJECT_ROOT GIT_DIR GIT_WORK_TREE GIT_COMMON_DIR && go test ./internal/kanban -run 'Test(Audit|PureBacklogReader|PeerRelocation|InspectArchiveVouch|Backlog|Migration|StateDir|RelocateQueue|AdoptingPath|ResolveTodo|TodoQueue|Queued|ConcurrencyStress|DuplicateID|Reordered)' -coverprofile=/tmp/moai-todo-storage-audit.TI3Io7/storage-focused.cover -count=1 -timeout=120s
```

```text
ok  	github.com/modu-ai/moai-adk/internal/kanban	9.352s	coverage: 42.6% of statements
```

전체 kanban 패키지의 42.6%는 검증 범위에 들지 않은 board/factory 기능을 포함한 분모다. 변경한 저장 구현 6개 파일의 statement 수를 같은 profile에서 계산한 결과는 다음과 같다.

```sh
awk '$1 ~ /internal\/kanban\/(backlog_store|backlog_sqlite|backlog_migrate|backlog_archive_vouch|todo_root|state_dir)\.go:/ {total += $2; if ($3 > 0) covered += $2} END {printf "changed storage files: %d/%d statements = %.1f%%\n", covered, total, 100*covered/total}' /tmp/moai-todo-storage-audit.TI3Io7/storage-focused.cover
```

```text
changed storage files: 795/938 statements = 84.8%
```

경쟁 상태 검증:

```sh
unset MOAI_HOME MOAI_CONFIG_DIR MOAI_PROJECT_ROOT GIT_DIR GIT_WORK_TREE GIT_COMMON_DIR && go test -race ./internal/kanban -run 'Test(AuditReadSnapshot|AuditResolvedWriterAfterRelocation|AuditRetirementBeforePublishResumes|PeerRelocation|PureBacklogReader|ConcurrencyStress)' -count=1 -timeout=120s
```

```text
ok  	github.com/modu-ai/moai-adk/internal/kanban	11.809s
```

홈 경로 소비 패키지 검증:

```sh
unset MOAI_HOME MOAI_CONFIG_DIR MOAI_PROJECT_ROOT GIT_DIR GIT_WORK_TREE GIT_COMMON_DIR && go test ./internal/homestate -count=1 -timeout=120s
```

```text
ok  	github.com/modu-ai/moai-adk/internal/homestate	8.963s
```

순수 조회 계약의 직접 검증:

```sh
unset MOAI_HOME MOAI_CONFIG_DIR MOAI_PROJECT_ROOT GIT_DIR GIT_WORK_TREE GIT_COMMON_DIR && go test ./internal/kanban -run '^TestPureBacklogReader' -v -count=1 -timeout=30s
```

```text
=== RUN   TestPureBacklogReaderPreservesSidecarInventory
--- PASS: TestPureBacklogReaderPreservesSidecarInventory (0.01s)
=== RUN   TestPureBacklogReaderWorksWithReadOnlyDatabase
--- PASS: TestPureBacklogReaderWorksWithReadOnlyDatabase (0.01s)
=== RUN   TestPureBacklogReaderRejectsSQLWrites
--- PASS: TestPureBacklogReaderRejectsSQLWrites (0.00s)
=== RUN   TestPureBacklogReaderSeesCommittedWAL
--- PASS: TestPureBacklogReaderSeesCommittedWAL (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.485s
```

마지막 선택 실행에는 추가한 `TestPureBacklogReaderNeverCreatesMissingDB`도 포함되어 있다. `git diff --check -- internal/kanban internal/homestate/paths.go`는 출력 없이 exit 0이었다.

## Gaps

- 전체 kanban 패키지 실행은 120초 제한에 도달했다. 출력은 `panic: test timed out after 2m0s`, 당시 실행 중 테스트는 `TestForemanQueueWatch_SeesWALDeferredCommit`이었다. 여러 16초 관측 창을 사용하는 foreman 검사까지 포함한 전체 패키지 PASS를 주장하지 않는다. 저장 관련 선택 실행과 별도 race 실행은 위와 같이 통과했다.
- Windows 실제 실행, 네트워크 파일시스템, 강제 종료 프로세스 주입은 수행하지 않았다. 중단 복구는 함수가 도달할 수 있는 파일 상태 fixture로 검증했다.
- 저장 계층 변경 중 일어난 CLI·web 검증 및 최종 통합 검증은 각 담당자의 별도 증거를 따른다.
- 홈 경로 패키지 전체 coverage 측정에서 63.5%가 나왔지만, 저장소 전체 테스트가 끝난 결과는 아니므로 통합 PASS 근거로 사용하지 않는다.

## Residual-risk

- query-only는 운영자 카드·스키마의 변경을 금지한다. SQLite가 조회 중 일시적인 WAL/SHM 조정 파일을 생성할 수 있다. checkpointed DB에서는 종료 후 파일 목록과 DB 바이트가 보존되는 것을 확인했고, active WAL에서는 커밋된 내용을 읽는 것을 확인했다. 조회 중 파일 생성 이벤트 자체가 전혀 없다는 뜻은 아니다.
- 과거 바이너리가 만든 불완전 DB와 JSON이 함께 있으면 원본을 보존하고 명시적 오류로 멈춘다. 어느 파일이 운영자의 최신 의도인지 추측하여 자동 덮어쓰지 않는다.
- 이관 전 경로를 잡은 writer는 `ErrBacklogRelocated`를 받는다. 프로젝트 루트를 다시 해석하는 새 명령으로 재시도할 수 있다. 구버전 바이너리는 새 retirement 표지를 이해하지 못하므로 구버전과 신버전의 동시 쓰기는 보호 범위 밖이다.
- archive 행 쓰기는 0으로 줄었지만 레코드 읽기와 archive 동등성 확인은 여전히 이력 크기에 비례한다. 일정 비용 또는 실행시간 감소율을 주장하지 않는다.
