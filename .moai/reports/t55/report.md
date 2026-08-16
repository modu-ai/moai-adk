# t55 — 칸반 SessionStart 공지의 에픽 안내를 todo 큐 요약으로 교체

- 카드: t55 (칸반 배치 tjv7iy round 3, cluster C, run 레인)
- 워크트리: 본 워크트리 (branch `WT-t55`, base = `release/v3.1.1` @ `051a2fa94`, clean fast-forward)
- 대상: 칸반 lead SessionStart 공지의 standing-context 라인 — 기존 `moai epic status <prefix>` 에픽 안내를 `moai todo` 큐 요약(N장 대기)으로 교체
- 선행 판단: 리드 확정 **축소(reduction)** — 에픽 추적·`internal/cli/epic.go`·t32 범위는 그대로 유지. 공지만 에픽을 가리키기를 멈춤

## 1. 주장 (Claim)

lead 세션이 실제로 움직이는 단위는 backlog(`.moai/state/kanban/backlog.json`) → 6컬럼 보드이지 에픽 마일스톤이 아니므로, 공지의 standing-context 라인은 에픽 안내 대신 **큐에 대기 중인 카드 수(N)와 `moai todo` 명령**을 안내해야 한다. 4개 로케일(en/ko/ja/zh) 전부 일관 교체되었고, N은 "정직하게 대기"의 정의에 따라 `state == "queued"` 카드만 센다. backlog 파일이 없거나 깨져 있어도 공지는 0으로 정상 렌더링되며 SessionStart는 절대 실패하지 않는다.

### N 산정 규칙 (어떤 상태를 세는가, 왜)

백로그 아이템의 상태는 `internal/kanban/backlog_store.go:44-51`에 따라 정확히 3개: `queued` / `picked` / `dropped`. 이 중 "waiting"에 정직한 것은 **`queued` 단 하나**:

- `picked` — 이미 다른 레인에서 진행 중(선행 카드가 주워 간 상태)
- `dropped` — 운영자가 폐기한 카드
- 완료 카드 — `moai todo done`이 행을 파일에서 **통째로 삭제**하므로(`internal/cli/todo.go` newTodoDoneCmd) 애초에 파일에 존재하지 않음

따라서 N = `Load()` 후 `State == kanban.BacklogStateQueued`인 아이템 수. `moai todo next`의 bare 출력이 같은 필터(`it.State != kanban.BacklogStateQueued` skip)를 쓰는 것과 동일한 정의다.

### 로케일별 라인

- en: `Kanban backlog: %d waiting — run `moai todo` to view the queue.`
- ko: `칸반 백로그: %d장 대기 중 — `moai todo` 를 실행하면 큐를 볼 수 있습니다.`
- ja: `かんばんバックログ: %d件が待機中 — `moai todo` を実行するとキューを確認できます。`
- zh: `看板待办队列：%d 张卡片在等待 — 运行 `moai todo` 可查看队列。`

`moai todo` 명령은 프로토콜 토큰으로 모든 로케일에서 backtick 그대로 유지(운영자가 복사해 붙는 대상). 영어는 복수형 굴절 문제("%d cards"의 N=1 어색함)를 피하려고 타원형 "%d waiting"으로 썼고, ko/ja/zh는 원래 굴절이 없어 수사(장/件/张)가 모든 N에 자연스럽다.

## 2. 증거 (Evidence) — TDD 순서대로

### RED — 새 동작을 주장하는 테스트가 구현 없이 실패

테스트 먼저 전면 교체(신규 3개 + 기존 갱신): `backlogSummary` 필드 존재·`%d` verb 보유, 프로토콜 토큰 `moai epic status <prefix>` 제거·`` `moai todo` `` 추가, 새 시그니처 `kanbanBootstrapNotice(root, lang)` / `kanbanLeadNotice(runID, root, lang)`, queued-only 카운팅, 4-로케일 패리티, fail-open 축소 케이스 4종.

    $ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook/ -run 'SessionStartKanban|Kanban' -count=1
    # github.com/modu-ai/moai-adk/internal/hook [github.com/modu-ai/moai-adk/internal/hook.test]
    internal/hook/session_start_kanban_i18n_test.go:34:24: m.backlogSummary undefined (type kanbanMessages has no field or method backlogSummary)
    internal/hook/session_start_kanban_i18n_test.go:127:37: too many arguments in call to kanbanBootstrapNotice
    	have (string, string)
    	want (string)
    internal/hook/session_start_kanban_test.go:42:38: too many arguments in call to kanbanBootstrapNotice
    	have (string, string)
    	want (string)
    ... (같은 계열 오류 계속)
    FAIL	github.com/modu-ai/moai-adk/internal/hook [build failed]

실측 exit code: **1** (별도 실행 `> /tmp/t55-red.log 2>&1` 후 `$?` 확인 → `1`, `build failed` 1회).

### GREEN — 구현 후 통과

구현: i18n 테이블 `epicPointer` → `backlogSummary`(4 로케일), 빌더에 root 스레딩 + `queuedBacklogCount(root)` 헬퍼(`kanban.NewBacklogStore(...).Load()` 재사용 — `moai todo` CLI와 같은 파서, 두 번째 파서 없음), `session_start.go`에서 `input.ProjectDir` → `input.CWD` 폴백으로 root 산출(= `chainLineageBanner`와 같은 우선순위).

    $ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook/ -run 'SessionStartKanban|Kanban|Factory' -count=1
    ok  	github.com/modu-ai/moai-adk/internal/hook	2.408s

신규 테스트 3개 개별 확인(`-run 'BacklogSummary' -v`):

    --- PASS: TestKanbanLeadNoticeBacklogSummaryCountsQueuedOnly (0.00s)
    --- PASS: TestKanbanLeadNoticeBacklogSummaryEveryLocale (0.00s)
    --- PASS: TestKanbanLeadNoticeBacklogSummaryFailsOpen (0.00s)

필터 실행 상세: `--- PASS` 37건 / `--- FAIL` 0건(verbatim 카운트 실측).

### 패키지 전체 (session_start.go도 건드렸으므로)

    $ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook/ -count=1
    ok  	github.com/modu-ai/moai-adk/internal/hook	29.605s

### 정적 검증

`gofmt -l` 대상 6파일(구현 3 + 테스트 3) 전부 출력 없음 — 포맷 준수. `go vet ./internal/hook/` 통과.

### Graceful degradation 설계 (증거로 뒷받침되는 동작)

`queuedBacklogCount`의 축소 경로 — `FailsOpen` 테스트 4서브케이스가 전부 "Kanban backlog: 0 waiting" 렌더를 실측:

1. root 불명(빈 문자열) → 0 (`if root == ""` 조기 반환; session_start.go의 ProjectDir→CWD 폴백 후에도 둘 다 비면 도달)
2. 파일 부재 → `BacklogStore.Load()` 자체 계약(REQ-TODO-012)이 빈 레코드 반환, 에러 아님 → 0
3. items 빈 배열 → 0
4. 파일이 깨진 JSON → `Load()`가 parse 에러 반환 → 헬퍼가 0으로 축소 (SessionStart 중단 없음)

### 비용 프로파일

SessionStart 매 부팅마다 도는 경로: 파일 읽기 1회(`atomicfile.ReadFile`, lock-free — `Load()`는 lock을 잡지 않음), 서브프로세스 0회, 포맷은 카운트 1회 — O(items) 순회가 전부.

## 3. Baseline 귀속

- 본 워크트리(HEAD `051a2fa94` = merge 후 release/v3.1.1 선두)에서 실측한 그대로. 모든 go test 출력은 해당 실행의 verbatim이며 환경은 칸반 런처 변수 5종을 사전 unset한 동일 조건.
- RED는 구현 전, GREEN은 구현 후 동일 트리에서 측정 — 출력 간 트리 drift 없음.

## 4. 미검증 (Gaps)

- `-race` 미실시 (CI `test-race` 잡 소관 — 로컬 전체 스위트/레이스 회피 규율).
- internal/hook 외 패키지 미실시 — `internal/kanban`은 기존 API(`NewBacklogStore`/`Load`/`BacklogStateQueued`)를 소비하기만 하므로 파급 없음. 전체 판정은 배치 PR의 CI 몫.
- 실제 칸반 런처 환경(`moai cc -k`)에서의 라이브 부팅 화면 목격은 안 함 — 핸들러 수준(`Handle` → SystemMessage)까지가 단위 테스트 커버.
- 에픽 잔여 참조 스윕: `internal/cli/epic.go`, `internal/epic/*`, `internal/web/screens_templ.go`, `internal/cli/epic_test.go`에 `moai epic status` 참조가 남아 있으나 이는 선행 판단(축소)상 **의도된 잔존** — 커맨드는 계속 존재한다.

## 5. 잔여 위험 (Residual-risk)

- 운영자가 에픽 마일스톤 진행을 칸반 세션 시작 화면에서 더는 볼 수 없음 — 대신 `moai epic status <prefix>`를 직접 실행하면 됨(커맨드 유지). 안내 노출이 사라진 것뿐이다.
- N은 SessionStart 직전 스냅샷 — 직후 운영자가 카드를 pick하면 화면의 N은 곧 stale해진다. 시작 시점 안내라는 용도에는 충분.
- ko 라인의 backtick 뒤 띄어쓰기("를 실행하면")는 기존 ko epicPointer가 쓰던 규약을 따랐다 — 운영자 복사 시 명령 자체는 backtick 안이라 무결.
- 기존 테스트의 호출부가 root 인자를 받도록 바뀌었다(12+ 호출 지점, 본 커밋에서 전부 갱신) — 향후 이 시그니처를 다시 건드리는 카드는 호출부 전체를 함께 봐야 한다.
