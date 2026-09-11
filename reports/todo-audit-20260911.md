# Todo 전수 조사 및 개선 보고

## 결과 요약

**확인한 개선 항목 22건을 구현하고 관련 회귀 검증을 통과했다.** 우선순위는 High 7건, Medium 12건, Low 3건이다. 보고 후 개선 카드 `t647`를 발행하여 저장 계층·CLI·웹·문서 변경을 추적했다. 최종 통합 선택 실행은 kanban, CLI, web 세 패키지 모두 통과했고, 저장·웹 경쟁 상태 검사와 실행 파일 빌드도 성공했다.

이 결과는 **로컬 구현·검증**에 대한 판정이다. 원격 develop 반영, 설치된 바이너리 교체, 실사용 홈 데이터 이관을 실행했다는 뜻이 아니다. 자세한 명령과 검증 경계는 [최종 판정](../.moai/reports/t647/todo-audit-20260911/verdict.md)에 보존한다.

구현 커밋은 **`3e03b1dc68f7a5b23bfc83f315aff24c2d2ebc7a`**이며, 카드 `t647`는 `done t647 landing=unknown` 이후 `archived`로 다시 확인했다. 여기서 `unknown`은 원격 착지 검사를 실행하지 않았다는 뜻이다. [커밋·카드 완료 기록](../.moai/reports/t647/todo-audit-20260911/completion.md).

## Claim

기준: origin/develop ee99507fbe3b4a22c6a0a74815723d222dfdc04d. CLI, SQLite 저장 계층, 상태 전환, 보관·복원, 이관, 웹·SSE, 상태줄·훅·설정, 관련 스킬·규칙·배포 문서를 조사했다. 다음 표는 발견 당시의 문제 목록이며, 수정 후 검증과 구별한다. High 7건, Medium 12건, Low 3건으로 총 22건이다.

| ID | 우선순위 | 확인한 문제 | 개선 범위 |
|---|---|---|---|
| T01 | High | 프로젝트 A 큐를 읽고 현재 디렉터리 B의 커밋으로 landed 판정 | PR 및 착지 조회를 큐 프로젝트에 고정 |
| T02 | High | 300개 카드의 동시 보관·복원 중 live 300 + archive 300으로 읽힘 | 한 읽기 트랜잭션으로 일관된 스냅샷 확보 |
| T03 | High | LoadPure가 없던 archive 테이블 생성 | 읽기 전용 연결과 구버전 읽기 호환 |
| T04 | High | 지원하지 않는 schema_version=999를 거부하기 전에 테이블 1개를 5개로 늘림 | 버전 검사 선행, 거부 전 스키마 변경 방지 |
| T05 | Medium | 두 관련 카드 보관 후 한 장만 복원하면 보관 중인 카드를 참조하는 live 관계 생성 | 관계 보존과 양 끝 카드 복원 시점 처리 |
| T06 | Medium | JSON 격리 실패 후 pending 표지가 지워져 재시도 불가 | 성공 전까지 재시도 표지 보존 |
| T07 | Medium | dropped 카드에 next 실행 시 폐기 표지를 남긴 채 picked로 전환 | undrop을 통한 명시적 복원 요구 |
| T08 | Medium | GH 조회 실패가 JSON에서 no-link로 표시 | 조회 불확실성의 기계 판독 가능 표현 |
| T09 | Medium | 줄바꿈·탭을 포함한 카드가 PR 표에서 가짜 추가 행으로 출력 | 표 렌더 경계에서 제어문자 처리, 원문 보존 |
| T10 | Low | 존재하지 않는 PR 조회 ID가 queue is empty로 성공하고 GH 호출 | 대상 ID 검증과 정확한 오류 |
| T11 | Low | 빈 큐에서 list --limit -1이 성공 | 데이터와 무관한 인자 검증 |
| T12 | Medium | landed 명령 생성 시 참조를 즉시 조회 | 명령 실행·도움말 시점의 지연 해석 |
| T13 | Medium | analyze가 매 카드 쌍마다 정규화·토큰 집합을 재생성 | 비교 의미를 유지하며 카드별 준비 결과 재사용 |
| T14 | Medium | 손상된 큐를 웹에서 정상 빈 큐로 표시 | 읽기 실패 상태를 화면에 표시 |
| T15 | Medium | 웹 시작 뒤 생성한 todo 디렉터리가 감시에서 누락되고 SSE ready는 폴링 중단 | 늦게 생성된 디렉터리도 갱신 감시 |
| T16 | Medium | 배포 스킬·문서에 옛 저장 경로와 temp+rename 설명 잔존 | 현재 홈 DB 및 SQLite 계약으로 관련 문서 동기화 |
| T17 | High | 서로 다른 separate-git-dir 저장소가 같은 부모를 큐 루트로 해석 | 작업 트리와 공용 Git 메타데이터 관계를 구별 |
| T18 | High | JSON 이관 도중 DB 생성 직후 중단하면 빈 DB가 원본 JSON을 가림 | 검증 완료 DB만 게시하고, 기존의 불완전한 이관은 원본을 보존하며 명시적 오류로 중단 |
| T19 | Medium | 카드 1개 추가 시 보관 카드 1,000개를 DELETE·INSERT하여 2,000행 쓰기 | 변경되지 않은 보관 이력 재기록 방지 |
| T20 | Low | 모든 picked 카드가 plan·SPEC으로 간다는 지침과 Class A/B 규칙 충돌 | Class별 단계와 선택적 SPEC 연결 명시 |
| T21 | High | 홈 DB 이관 전에 옛 경로를 해석한 writer가 이관 뒤 쓰면 성공을 반환하지만 새 큐에 카드 없음 | 잠금 획득 후 이관된 원본의 쓰기 차단 또는 재해석 |
| T22 | Medium | 카드 쓰기 없이 HTTP 조회가 SQLite 파일 이벤트를 만들어 SSE 재조회를 계속 유발 | 읽기의 보조 파일 이벤트와 실제 큐 변경을 구별하여 자기 갱신 반복 차단 |

## Evidence

저장 계층 재현: `go test -overlay /tmp/moai-todo-storage-audit.TI3Io7/overlay.json ./internal/kanban -run '^TestAudit' -count=1 -v -timeout=90s`

```text
LoadPure archive tables: before=false after=true
unsupported schema_version "999" (want "1"): kanban backlog store corrupt; tables before=1 after=5
restored live finding references archived card
read 2 mixed commits: live=300 archived=300 total=600 want=300
after obstruction removed and Load retried: legacy JSON still present=true
FAIL github.com/modu-ai/moai-adk/internal/kanban 0.264s
```

CLI 관측 테스트: `go test -overlay /tmp/todo-cli-audit.4G91GB/overlay.json ./internal/cli -run '^TestTodoAudit' -count=1 -v -timeout=90s`. 이 테스트의 PASS는 잘못된 기존 동작을 재현했다는 뜻이다.

```text
stdout=[{"card_id":"t1","outcome":"no-link"}]
stderr="note: open pull requests unavailable (offline); link column left empty, landed check still ran\n"
stdout="queue is empty\n" err=<nil> subprocesses=1
stdout="t1\tno-link\t\t\tqueued\t\tcard text\nforged\trow\n" err=<nil> records=2
stdout="queue is empty\n" err=<nil>
stdout="picked t1 [DROPPED — obsolete] original card\n" err=<nil> state=picked
constructor populated Long=true refUsage="Ask about this ref instead of origin/main"
stdout=[{"card_id":"t1","outcome":"landed"}]
PASS
ok github.com/modu-ai/moai-adk/internal/cli 3.979s
```

기존 저장 계층 회귀: `go test ./internal/kanban -run 'Test(NormalizeCardText|TokenSetJaccard|ClassifyCardText|Backlog)' -count=1 -timeout=120s`

```text
ok github.com/modu-ai/moai-adk/internal/kanban 2.366s
```

## Baseline-attribution

초기 참조 검색 후보 217개와 구현 후 동일 검색의 후보 221개 파일은 [조사 목록](todo-audit-inventory-20260911.md)에 기록했다. CLI, 저장 계층, 웹·상태줄·훅·설정, 지침·배포 문서에 걸친 조사이다. 전수 검색 목록과 실제 판독·실행 검증을 구별한다. 파일마다 모든 줄을 정독하거나 모든 분기를 실행했다는 주장은 하지 않는다.

| 영역 | 주요 구현 전 위치 | 확인 방법 |
|---|---|---|
| T01·T08–T10 | internal/cli/todo_pr.go:174,181,212,245 | 두 Git 프로젝트 대조군, GH 실패 주입, 다중행 출력 |
| T07·T11 | internal/cli/todo.go:555,845 | 빈 큐와 dropped 상태의 CLI 실행 |
| T12·T13 | internal/cli/todo_landed.go:53; todo_analysis.go:151 | 생성자 검사와 할당량 벤치마크 |
| T02·T18·T19 | internal/kanban/backlog_migrate.go:49,212,339,493 | 동시 SQL writer, 중단 상태 fixture, 쓰기 trigger 계수 |
| T03·T04 | internal/kanban/backlog_sqlite.go:271,302; backlog_store.go:574 | 호출 전후 schema inventory 비교 |
| T05·T06 | internal/kanban/backlog_store.go:280; backlog_migrate.go:529,591 | 보관·부분 복원, 격리 장애 제거 후 재시도 |
| T17·T21 | internal/kanban/todo_root.go:180; state_dir.go:162,191,203; internal/homestate/paths.go:50 | separate-git-dir 두 저장소, 이관 뒤 기존 writer 재개 |
| T14·T15 | internal/web/todo_queue_read.go:38; events.go:129; assets/app.js:769 | 손상 큐 HTTP 테스트, 실제 watcher와 JS 이벤트 처리 |
| T16·T20 | .claude/skills/moai/workflows/todo.md:17,194; docs-site/content/en/utility-commands/moai-todo.md:91,118 | 내장 템플릿·Class A/B 규칙과 비교 |

위 줄 번호는 구현 전 기준이다. 구현 후 줄 번호는 변경될 수 있다.

추가 격리 재현:

```text
two independent git repositories resolve to same todo root
reopened live count=0; retained legacy={"version":1,"last_seq":1,"items":[{"id":"t1","text":"one","state":"queued"}]}
late Add returned id=t2; canonical count=1
FAIL github.com/modu-ai/moai-adk/internal/kanban 0.836s
```

성능 원본 관측(감사 담당, Apple M4 Max, 각 1회):

```text
BenchmarkTodoAuditAnalyze/100-16  1  5168417 ns/op    5945600 B/op    79208 allocs/op
BenchmarkTodoAuditAnalyze/500-16  1  128476166 ns/op  149705664 B/op  1996012 allocs/op
one Add with 1000 unchanged archived cards executes 2000 archive row writes
```

벤치마크 명령: `go test -overlay /tmp/todo-cli-audit.4G91GB/overlay.json ./internal/cli -run '^$' -bench '^BenchmarkTodoAuditAnalyze$' -benchtime=1x -benchmem -count=1`. 절대 시간은 해당 머신의 단일 측정이다. 데이터셋과 비교 의미를 유지한 수정 후 값을 별도로 기록한다.

조사는 최신 develop 및 동일 HEAD의 새 런처 워크트리 `todo-audit` / `WT-todo-audit`에서 수행했다. 재현은 임시 Git 저장소·임시 DB·Go overlay로 격리했다. 기존 t549·t555 등은 `moai todo pr --json`에서 landed 후보로 확인되었으며, 기존 카드의 실제 종결 여부는 별도 판단이다.

## Gaps

Windows 실제 실행, 전원 차단 내구성, 모든 외부 파일시스템의 잠금 특성은 측정하지 않았다. 결함 후보를 새 버그로 오인하지 않도록 최신 코드의 재현을 우선했다. 전수 조사는 관련 표면의 조사 범위를 뜻하며 모든 가능한 입력의 결함 부재를 보증하지 않는다.

전체 kanban 패키지 실행은 foreman 관측 테스트 도중 120초 제한에 도달했다. CLI의 넓은 선택 실행에서 나온 파일 무변경 실패는 수정 후 해당 실패 위치를 모두 재실행하여 통과했지만, 같은 넓은 선택 전체를 마지막에 다시 실행하지는 않았다. 저장·CLI·웹의 최종 범위별 회귀와 전체 저장소 CI는 다른 검증이다.

## Residual-risk

실사용 홈 상태 이관과 기존 카드의 정리·폐기는 이번 구현에서 별도 실행하지 않는다. 기존 live 카드 수에는 이미 착지한 작업이 포함될 수 있다. 이전 59건 보고는 큐 잔류 수이며 실제 미구현 수로 사용할 수 없다.

query-only 연결은 카드·스키마에 대한 SQL 쓰기를 거부하지만, SQLite의 임시 WAL/SHM 조정 파일 사용 자체를 금지하지는 않는다. 정지 상태의 DB는 조회 종료 후 바이트·파일 목록 보존을, 활성 WAL은 커밋된 카드 조회를 검증했다. 구버전 바이너리는 새로운 이관 차단 표지를 이해하지 못하므로 신구 바이너리의 동시 쓰기는 보호 범위 밖이다. 분석은 여전히 카드 쌍을 모두 비교하며, 보관 이력 읽기·동등성 비교도 이력 크기에 비례한다.

## 성능 개선 관측

| 항목 | 수정 전 | 수정 후 | 해석 |
|---|---:|---:|---|
| 100개 카드 analyze | 5,168,417 ns/op | 410,000 ns/op | 동일 합성 입력, 각 1회 |
| 500개 카드 analyze | 128,476,166 ns/op | 8,709,042 ns/op | 약 14.75배 빠른 단회 관측이며 운영 보장이 아님 |
| 500개 카드 분석 할당량 | 149,705,664 B/op | 228,312 B/op | 카드별 준비 결과 재사용 |
| 500개 카드 분석 할당 횟수 | 1,996,012 allocs/op | 2,503 allocs/op | 비교 판정은 144개 대조 조합으로 보존 확인 |
| 보관 카드 1,000개가 있는 큐에 카드 1개 추가 | 보관 테이블 2,000행 쓰기 | 보관 테이블 0행 쓰기 | 변경되지 않은 보관 이력에 대한 trigger 계수 |
| 큐 쓰기 없는 HTTP 조회 후 자기 갱신 | 4/4회 연속 발생 | 0/4회 발생 | 실제 DB·WAL 커밋 알림은 별도 양성 대조군으로 유지 확인 |

수정 후 벤치마크 명령은 `go test ./internal/cli -run '^$' -bench '^BenchmarkTodoAuditAnalyze$' -benchtime=1x -benchmem -count=1`이다. 시간값과 할당량의 전체 원문은 CLI 근거에 있다.

## 배포 템플릿 무결성

Todo 워크플로 문서 수정으로 배포 `moai` 스킬 트리의 내용 해시가 달라졌다. 변경 직후 `TestManifestHashFormat`에서 다음 오류가 발생했다.

```text
CATALOG_HASH_UNSTABLE: moai stored hash=b2e47f5f0131333ca791dc3a039637c10e60729f8f3ec4fbcbe67a1a7ddf3efa, computed hash=fab8717372fcb55360a2ceb0e3a643544851e4105c083bf2747b0fe7a4424bd7 (source=.claude/skills/moai/ (whole tree))
FAIL github.com/modu-ai/moai-adk/internal/template 0.474s
```

기존 생성기 `go run ./internal/template/scripts/gen-catalog-hashes.go --entry moai`로 해당 항목만 재생성했다. 생성기는 `catalog.yaml updated successfully (12899 bytes)`를 출력했고, diff는 해당 해시 한 줄이다. 다음 명령으로 다시 확인했다.

```text
go test ./internal/template -run 'TestManifestHashFormat|Test.*Todo|Test.*Kanban' -count=1 -timeout=90s
ok  github.com/modu-ai/moai-adk/internal/template 0.488s
```

## Card Cross-Check

`t647` 발행 및 picked 확인. T01–T22를 저장 계층, CLI, 웹·문서 세 구현 묶음으로 추적한다. 이관 writer와 웹 자기 갱신 반복 재현 결과는 같은 카드에 추가 반영했다. 기존 t549·t555·t575·t591·t592 등은 별도 기록을 유지한다.

| 구현 묶음 | 항목 | card | 근거 |
|---|---|---|---|
| 저장·경로·이관 | T02–T06, T17–T19, T21 | t647 | storage.md |
| CLI·분석 | T01, T07–T13 | t647 | cli.md |
| 웹·스킬·규칙·배포 문서 | T14–T16, T20, T22 | t647 | surfaces.md |

세 구현 묶음을 하나의 감사 개선 카드로 발행·추적했으며, 기존 카드에 대한 추가 종결·폐기 작업은 포함하지 않는다.

주의: 기준 트리에 과거의 다른 작업에 속한 `.moai/reports/t647/verdict.md`가 이미 존재한다(`691f489eb`). 현재 큐의 `t647` 원문은 이번 감사 작업으로 확인했다. 과거 보고서는 보존하고, 이번 최종 판정은 `.moai/reports/t647/todo-audit-20260911/verdict.md`로 구분한다. 번호 재사용의 원인은 이번 실행에서 확인하지 않았으며, 카드 번호만으로 과거 작업의 완료와 이번 구현의 완료를 동일시하지 않는다.

## 구현의 주요 판단

- **읽기와 이관을 분리했다.** PR·이력·설명 조회는 순수 읽기 경로를 사용한다. 구버전 스키마에 없던 필드는 읽기 시 호환 처리하며, 조회를 이유로 테이블을 추가하지 않는다.
- **새 저장소가 준비되기 전에는 권위를 넘기지 않는다.** JSON 이관은 같은 디렉터리의 임시 DB에 기록하고 재조회로 내용이 일치하는지 확인한 뒤 정식 이름으로 게시한다. 옛 버전이 남긴 불완전한 DB와 JSON의 조합은 자동으로 어느 쪽을 삭제하거나 덮어쓰지 않고 오류를 반환한다.
- **늦은 쓰기는 성공으로 위장하지 않는다.** 이관 원본에는 쓰기 차단 표지를 남긴다. 옛 경로를 잡은 명령은 명확한 오류를 받고, 프로젝트 경로를 다시 해석한 명령으로 재시도해야 한다. 롤백용 원본은 보존한다.
- **관계의 양쪽 카드가 복원될 때까지 관계도 보관한다.** 한 장만 복원했을 때 살아 있는 관계가 아직 보관 중인 카드를 가리키지 않도록 한다.
- **조회 실패와 빈 결과를 구별한다.** PR 조회 실패·상한 도달은 JSON에도 표시하고, 웹의 DB 읽기 실패는 정상적인 0건과 다른 화면으로 표시한다.
- **비교 결과는 유지하고 반복 준비만 줄인다.** 분석의 정규화·토큰 집합을 카드별로 한 번 준비한다. 보관 내용이 바뀌지 않은 일반 카드 추가에서는 보관 테이블을 다시 쓰지 않는다.

### 교차 검토에서 추가로 막은 경로

저장소 이관 대상이 이미 존재하는 경우의 조기 반환에서도 옛 원본의 쓰기 차단이 필요했다. 독립 검토에서 아래 실패를 재현하여 T21의 회귀 범위를 확장했다.

```text
go test -overlay /tmp/todo-surface-audit.IYzgx9/storage-overlay.json ./internal/kanban -run TestPeerRelocationAlreadyPublishedFencesSource -count=1 -v
stale writer err=<nil>; expected relocated refusal after target already existed
FAIL github.com/modu-ai/moai-adk/internal/kanban 0.416s
```

수정 후 검증 결과는 각 담당 근거와 최종 판정 문서에 기록한다. [CLI 근거](../.moai/reports/t647/cli.md), [저장 계층 근거](../.moai/reports/t647/storage.md), [웹·문서 근거](../.moai/reports/t647/surfaces.md).

### 웹 조회가 스스로 재조회를 유발하는 결함

`TestPeerTodoReadWatcherFeedback`는 큐를 만든 뒤 추가 쓰기를 하지 않고 감시를 시작한다. 아무 조회도 하지 않은 대조 구간에는 이벤트가 없지만, HTTP `/todo`를 읽을 때마다 `kanban` 이벤트를 받아 다시 읽는 과정이 4회 연속 유지되었다. 같은 overlay 검사를 원래 기준 트리와 현재 트리에 각각 실행했으므로 이번 수정이 새로 만든 결함으로 분류하지 않는다.

```text
read-triggered kanban cycles=4/4
HTTP reads sustain todo refresh feedback without queue writes
FAIL github.com/modu-ai/moai-adk/internal/web 3.889s
```

원래 기준 트리에서도 같은 관측으로 실패했다(`4.290s`). T22는 이벤트 알림을 모두 끄는 방식이 아니라, 실제 DB·WAL 커밋 알림을 유지하는 조건으로 수정한다.

수정 후 같은 진단 시험은 `read-triggered kanban cycles=0/4`와 PASS를 출력했다(`7.042s`). 영구 회귀 시험에서는 읽기 후 자기 갱신이 없음을 확인하고, 종료된 writer의 저장·열린 writer의 WAL 커밋·뒤늦게 생성된 두 저장 경로의 알림이 유지되는 것을 함께 확인했다(`8.304s`).

## 최종 통합 검증

명령은 프로젝트·홈·Git 경로 override를 한 번의 호출 안에서 제거한 뒤 실행했다. 상세 선택자는 최종 판정 문서에 그대로 보존했다.

```text
ok  github.com/modu-ai/moai-adk/internal/kanban 1.089s
ok  github.com/modu-ai/moai-adk/internal/cli 26.787s
ok  github.com/modu-ai/moai-adk/internal/web 9.830s
```

웹의 최종 race 선택 실행:

```text
go test -race ./internal/web -run 'TestTodoUnreadableDoesNotClaimEmpty|TestTodoWatcherRegistersLateDirectories|TestTodoReadsDoNotRefreshThemselves|TestTodoCommitsStillRefresh|TestResolvedWatchPathsIncludeHomeTodoAndFactory' -count=1 -timeout=90s
ok  github.com/modu-ai/moai-adk/internal/web 11.491s
```

최종 통합에서 소스의 줄 번호를 검사하는 `TestAuditLagUsesBinlagSeam`의 좌표 불일치도 확인했다. `home_state_coverage.go`의 2개 좌표는 기준 커밋 소스로 재구성한 검사에서도 실패했고, `todo_landed.go`의 1개 좌표만 이번 수정으로 이동했다. 실제 ancestry 검사 코드는 바꾸지 않고 기대 좌표와 대응 주석만 동기화했다. 별도 실행은 PASS(`1.597s`), 위 통합 재실행도 PASS다. 이는 새 기능 수정으로 집계하지 않고 검사 기록의 정합성 보정으로 구분한다.
