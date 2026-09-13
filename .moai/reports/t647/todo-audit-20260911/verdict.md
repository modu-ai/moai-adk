# Todo 전수 감사 개선 — 최종 구현 판정

## Claim

T01–T22의 로컬 구현과 관련 회귀 검증을 완료했다. 최종 통합 선택 실행은 저장 계층, CLI, 웹 세 패키지에서 통과했다. 이 판정은 전체 저장소 CI, 원격 통합, 배포 또는 실사용 홈 이관의 완료를 뜻하지 않는다.

| 범위 | card | 결과 | 상세 근거 |
|---|---|---|---|
| 저장·경로·이관 9건 | t647 | 구현 및 범위별 회귀 통과 | ../storage.md |
| CLI·분석 8건 | t647 | 구현 및 범위별 회귀 통과 | ../cli.md |
| 웹·문서·규칙 5건 | t647 | 구현·실제 감시/HTTP 시험·문서 대조 통과 | ../surfaces.md |

## Evidence

최종 통합 명령:

```sh
unset MOAI_HOME MOAI_CONFIG_DIR MOAI_PROJECT_ROOT CLAUDE_PROJECT_DIR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_COMMON_DIR && go test ./internal/kanban ./internal/cli ./internal/web -run '^(TestAudit|TestPureBacklogReader|TestPeer|TestTodoAudit|TestTodoHistoryLeavesStorageByteIdentical|TestTodoHistoryDegradesWithoutArchiveTables|TestTodoPR_QueueDirUnchanged|TestTodoPR_ProjectRootUnchangedWithEvidence|TestTodoUnreadable|TestTodoWatcher|TestResolvedWatchPaths|TestTodoReadsDoNotRefreshThemselves|TestTodoCommitsStillRefresh)' -count=1 -timeout=120s
```

```text
ok  	github.com/modu-ai/moai-adk/internal/kanban	1.089s
ok  	github.com/modu-ai/moai-adk/internal/cli	26.787s
ok  	github.com/modu-ai/moai-adk/internal/web	9.830s
```

최종 웹 경쟁 상태 검사:

```sh
unset MOAI_HOME MOAI_CONFIG_DIR MOAI_PROJECT_ROOT CLAUDE_PROJECT_DIR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_COMMON_DIR && go test -race ./internal/web -run 'TestTodoUnreadableDoesNotClaimEmpty|TestTodoWatcherRegistersLateDirectories|TestTodoReadsDoNotRefreshThemselves|TestTodoCommitsStillRefresh|TestResolvedWatchPathsIncludeHomeTodoAndFactory' -count=1 -timeout=90s
```

```text
ok  	github.com/modu-ai/moai-adk/internal/web	11.491s
```

저장 담당의 최종 race 명령과 출력:

```sh
unset MOAI_HOME MOAI_CONFIG_DIR MOAI_PROJECT_ROOT GIT_DIR GIT_WORK_TREE GIT_COMMON_DIR && go test -race ./internal/kanban -run 'Test(AuditReadSnapshot|AuditResolvedWriterAfterRelocation|AuditRetirementBeforePublishResumes|PeerRelocation|PureBacklogReader|ConcurrencyStress)' -count=1 -timeout=120s
```

```text
ok  	github.com/modu-ai/moai-adk/internal/kanban	11.809s
```

관련 홈 경로 패키지 전체 실행은 `ok github.com/modu-ai/moai-adk/internal/homestate 8.963s`였다. 저장 담당의 원래 명령은 ../storage.md에 있다. 변경한 저장 구현 6개 파일의 선택 실행 커버리지는 795/938 statements, 84.8%이며, 전체 kanban 패키지 커버리지와 구분한다.

실행 파일 빌드:

```sh
unset MOAI_HOME MOAI_CONFIG_DIR MOAI_PROJECT_ROOT CLAUDE_PROJECT_DIR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_COMMON_DIR && go build -o /tmp/moai-t647-build.PcNVlq/moai ./cmd/moai
```

표준 출력 없음, exit 0. 설치 바이너리는 교체하지 않았다.

새 바이너리의 격리된 CLI 흐름 검사: `MOAI_HOME=/tmp/moai-t647-smoke.yBWlaa/home`, `CLAUDE_PROJECT_DIR=/tmp/moai-t647-smoke.yBWlaa`를 지정하고 그 임시 Git 저장소에서 실행했다. `todo add`, `drop`, 거부되어야 하는 `next`, `pr t999`, `list --limit -1`을 먼저 확인한 뒤 다음 명령을 순서대로 실행했다.

```text
moai todo undrop t1
moai todo next t1
moai todo done t1
moai todo history t1
moai todo undone t1
moai todo history t1
```

명령별 관측 stdout:

```text
undropped t1 t647 isolated smoke card
picked t1 t647 isolated smoke card
done t1 landing=unknown
t1	archived	picked	t647 isolated smoke card
undone t1 t647 isolated smoke card
t1	live	picked	t647 isolated smoke card
```

부정 입력은 각각 nonzero를 반환했고, stdout/stderr에서 `backlog item t1 is dropped — use moai todo undrop t1 before picking`, `No backlog item t999.`, `Todo list: --limit must be >= 0 (got -1).`를 확인했다. 임시 프로젝트에 config sections가 없어 defaults 사용 경고가 함께 출력되었다. 이 카드는 실제 사용자 큐의 t647가 아닌 임시 큐의 t1이다.

배포 메타데이터 검증은 `go test ./internal/template -run 'TestManifestHashFormat|Test.*Todo|Test.*Kanban' -count=1 -timeout=90s`로 실행하여 `ok github.com/modu-ai/moai-adk/internal/template 0.488s`를 확인했다. 스킬 트리 변경으로 발생한 해시 불일치는 기존 생성기의 `--entry moai`로 해당 항목만 갱신했다.

`git diff --check && node --check internal/web/assets/i18n.js && cmp .claude/skills/moai/workflows/todo.md internal/template/templates/.claude/skills/moai/workflows/todo.md`는 출력 없이 exit 0이었다.

## Baseline-attribution

- 기준: origin/develop `ee99507fbe3b4a22c6a0a74815723d222dfdc04d`.
- 구현 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/todo-audit`.
- 브랜치: `WT-todo-audit`.
- source_session_id: `01a08e64-c9d3-76f3-a589-5d5d893e1b62` (`moai session current`로 다시 확인).
- 계측 플랫폼: macOS arm64, Apple M4 Max. 시점: 2026-09-11.
- 각 구현 전 실패와 수정 후 성공은 별도 기록했다. 위 최종 통합 실행은 세 담당의 코드 변경과 좌표 fixture 보정 이후의 같은 작업 트리에서 실행했다.

## Gaps

- 전체 저장소 테스트·CI를 실행하지 않았다. 전체 kanban 시도는 foreman 관측 테스트 도중 120초 제한에 도달했다. 범위별 최종 성공을 전체 suite 성공으로 해석해서는 안 된다.
- Windows/Linux 파일 감시 조합, 네트워크 파일시스템, 실제 전원 차단·프로세스 강제 종료, 실서비스 GitHub 네트워크 E2E, 브라우저 전체 사용자 여정은 검증하지 않았다.
- 원격 push/PR/merge, 설치 바이너리 교체, 실사용 홈 이관은 실행하지 않았다. 현재 Git 전략은 manual/local이며 auto_push/auto_pr가 false다.
- 기존 Git 설정의 `core.hooksPath`는 `/dev/null`이었다. 이 설정을 바꾸거나 이번 작업에서 hook을 우회 설정하지 않았으며, Git hook 실행을 검증 근거로 삼지 않는다. 위 명시적 검사를 별도로 실행했다.
- 최초 통합 선택 실행에서 `TestAuditLagUsesBinlagSeam`이 실패했다. 원래 기준 소스로 재구성한 검사에서도 home-state의 2개 좌표가 어긋남을 확인했고, Todo의 1개 좌표는 이번 변경으로 이동했다. 실제 ancestry 동작이나 예외 집합은 바꾸지 않고 수동 fixture 좌표와 주석만 동기화했다. 같은 통합 명령의 마지막 실행은 위와 같이 통과했다.

## Residual-risk

- SQLite 조회가 일시적인 WAL/SHM 조정 파일을 만들 수 있다. DB 데이터·스키마·조회 종료 뒤의 안정된 파일 목록 보존과 활성 WAL 읽기를 확인한 것이며, 모든 순간의 파일 이벤트 부재를 주장하지 않는다. 웹은 읽기만의 SHM/WAL 생명주기 이벤트를 제외하고 실제 WAL 쓰기 알림은 유지한다.
- 구버전 바이너리는 새 이관 차단 표지를 이해하지 못한다. 신구 바이너리 동시 쓰기 방지까지 검증한 것은 아니다.
- 옛 불완전 DB+JSON은 원본을 보존하며 오류로 멈춘다. 자동 데이터 선택·덮어쓰기 복구는 하지 않는다.
- 성능은 동일 합성 입력 단회 관측이다. analyze의 비교 횟수는 여전히 N(N-1)/2, 보관 이력 읽기·비교는 이력 크기에 비례한다.
- 과거 다른 작업에 속한 `t647/verdict.md`가 이미 있었으므로 보존했다. 현재 큐 카드 원문과 날짜별 경로로 이번 작업을 구분한다. 과거 번호 재사용의 원인은 단정하지 않는다.

## Card Cross-Check

22개 발견 사항을 세 구현 묶음으로 나누어 한 카드 t647에 반영했다. 현재 큐의 원문 접두사 `[TODO 전수 감사 2026-09-11`을 확인하고 수정했으며, 과거 동명 카드의 착지 기록을 이번 구현의 완료 근거로 사용하지 않았다. 구현 커밋 `3e03b1dc68f7a5b23bfc83f315aff24c2d2ebc7a` 이후 카드의 archived 상태를 다시 확인했다. 정확한 명령과 출력은 [완료 기록](completion.md)에 남겼다.
