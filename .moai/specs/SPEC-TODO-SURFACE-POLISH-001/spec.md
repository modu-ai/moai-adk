---
id: SPEC-TODO-SURFACE-POLISH-001
title: "todo 표면 정비 5종 — show 동사 신설, add -f 오파싱 수리, list 한도 완화, 유령 저장소 소멸, 라벨 어휘 통합 (+렌더 누락 점검)"
version: "0.1.0"
status: in-progress
created: 2026-09-29
updated: 2026-09-30
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/cli + internal/kanban"
lifecycle: spec-anchored
tags: "todo-queue, cli-surface, ghost-store, label-vocabulary, render-completeness, t1349"
tier: M
related_specs: [SPEC-TODO-STALE-STORE-001, SPEC-TODO-QUEUE-HOME-CANON-001, SPEC-ROLE-NAMING-CODE-001, SPEC-TODO-ARCHIVE-QUERY-001, SPEC-TODO-CLAIM-LEASE-001, SPEC-BACKLOG-JSON-DISCLOSURE-001]
---

# SPEC-TODO-SURFACE-POLISH-001 — todo 표면 정비 5종 (카드 t1349)

## HISTORY

- 2026-09-29 v0.1.0 — manager-spec 최초 작성 (카드 t1349, plan-phase). 카드 본문 5종(①show ②add -f ③list 한도 ④유령 저장소 ⑤라벨 어휘) + 실측 추가 항목(렌더 누락)을 단일 SPEC으로 묶음. 측정 근거는 본 문서 §A와 plan.md §A/§C에 각각 기록(2026-09-29 본 세션 실측).

## A. 배경과 문제 정의

정식 큐는 홈 DB `~/.moai/db/<project-key>/todo/backlog.db` 하나다
(`internal/cli/todo.go:221` 도움말, `todo.go:74-76` `resolveTodoQueueRoot`). 이 카드는
보고서 P10(`.moai/reports/autonomy-bottleneck-proposal-20260929.html`)과 카드 본문이
지적한 표면 정비 5종 + 렌더 누락 1건을 하나로 정비한다. 모든 항목은 본 세션
(2026-09-29, 워크트리 t1349 @ 51abf337a)에서 재측정한 사실에 근거한다.

- **① show 부재** — `grep '"show"' internal/cli/todo.go` 적중 0(양성 대조
  `grep '"add"'` = `todo.go:652` 적중). 동사 등록부는 `todo.go:303-309`의
  `AddCommand` 22종이며 show가 없다. 단일 카드 조회는 `history <id>`의
  운명(fate) 한 줄(`internal/cli/todo_history.go:223-253`)로만 가능해, 카드
  전문(全文)을 읽는 표면이 없다. 역설로 안내 가드 주석은 이미 `show 401`을
  예시로 인용하고 있다(`todo.go:481`).
- **② add -f 오파싱** — `newTodoAddCmd`의 `--force`는 숏핸드 없이 선언됐다
  (`todo.go:617-618`, `BoolVar` — `todo*.go` 전체에서 `BoolVarP` 적중 0). 그러나
  pflag의 interspersed 파싱은 `-`로 시작하는 토큰을 플래그로 소비하므로, 본문이
  `-f`로 시작하는 카드(`moai todo add "-f ..."`)는 등록되지 않은 숏핸드 오류로
  생성에 실패한다. 카드 본문이 인용한 `todo.go:601`은 구행 번호며 본 트리에서는
  `:617-618`이다.
- **③ list 기본 20행 한도** — `todoListDefaultLimit = 20`
  (`todo.go:727`). 실측(2026-09-29): live 카드 55장(queued 44 + picked 11)으로
  기본 렌더는 `list: 35 rows withheld — showing 20 of 55 (--limit 0 lists all)`
  로 잘린다(리드 실측과 동일 문구). 잘림 자체는 설계(경계 고지
  `todo.go:810-814`)지만, 한도값이 실제 큐 규모를 2배 이상 웃돌지 못해 매일
  재발한다.
- **④ 유령 저장소** — 홈 DB 컷오버(SPEC-TODO-QUEUE-HOME-CANON-001) 이후에도
  유령 아티팩트가 남는다. 실측(2026-09-29): 홈
  `~/.moai/db/moai-adk-go-1bd3d038/todo/backlog.json` 1,884,049바이트(09-26),
  프로젝트-로컬 `.moai/state/todo/backlog.json` 652,478바이트(09-10),
  `backlog.json.migrated` 155,043바이트, 세션 레코드 JSON 약 397개
  (`.moai/state/todo/<uuid>.json`, 1개 200~240바이트). 프로젝트-로컬
  `backlog.db`는 이제 존재하지 않는다(`stat` 부재 확인 — t1307 시대의 SQLite
  유령은 이미 처분된 것으로 관측). 기존 검출기
  `kanban.InspectStaleLocalStores`(`internal/kanban/todo_stale_store.go:62-94`)는
  SQLite `backlog.db` 두 경로(todo/kanban 디렉터리)만 보므로 위 아티팩트 전체에
  눈이 없고, 홈 쪽 `backlog.json`은 검출 대상 디렉터리(프로젝트-로컬 2곳) 밖에
  있다.
- **⑤ 저장 라벨 어휘 불일치** — 홈 DB `todo_runtime_assignments.owner_label`
  에 `lead` 453행 + `worker-67` 13행(2026-09-29 GROUP BY 실측 — 카드 본문 수치와
  정확히 일치). 정식 어휘는 leader 역할 `leader`(`internal/kanban/role.go:41`,
  `lead`는 `:46`의 감지 전용 레거시 표기)와 레인 라벨 `lane-<n>`
  (`internal/kanban/bootstrap.go:246`; `worker-<n>`/`agent-<n>`은 감지 전용,
  `:252-255`)이다. 표나 레코드를 새로 쓰지 않는다는 기존 원칙
  (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-009)과 카드가 요청한 일괄 이관은 충돌하므로,
  본 SPEC은 운영자의 카드 요청을 따라 이 표 한정 일회 마이그레이션으로 해소하되
  그 경계를 명시한다(§F.1).
- **⑥ 렌더 누락(카드 실측 추가)** — 2026-09-29 worker-63 관측: t1338이 홈 DB
  items에 `queued`로 존재하는데 `moai todo` 목록(--limit 60)에 미표출. 본 세션
  재현 시도: `moai todo list --limit 60 | grep -c t1338` = 4(행 1 + findings) —
  **오늘은 재현되지 않는다**. 렌더 필터(`todo.go:778-793`)는 dropped 제외가
  유일하다. 한 가지 구조적 단서: 코드 DDL은 4-상태 CHECK
  (`internal/kanban/backlog_sqlite.go:128`, schema v2 `:53`)인데 실측 홈 DB는
  schema_version=1 + 3-상태 CHECK(`sqlite_master` 실측)로, v1 DB에서 hold 경로의
  동작이 규정 밖이다. 누락의 재발을 구조적으로 막는 것은 "렌더된 행 집합 = 응답
  스토어의 필터 만족 행 집합" 불변식이다(REQ-TSP-030).

## B. 요구사항 (GEARS)

### B.1 항목 ① — `todo show <id>` 단일 카드 조회 동사

- **REQ-TSP-001**: **When** `moai todo show <id|n>` 이 실행되면, the CLI shall
  `history`의 조회 기계(`todo_history.go:223-253`)와 같은 읽기 스토어에서 그
  카드 한 장의 전체 기록 — id, 현재 상태(`queued`/`picked`/`dropped`/`hold`
  전부), landing 셀, 전이 스탬프, **절단 없는 카드 본문 전문** — 을 한 줄
  탭 구분 형식(본문은 마지막 필드, `todoPRCell` 평탄화)으로 stdout에 내놓는다.
- **REQ-TSP-002**: **When** show가 큐에 없는 id를 받으면, the CLI shall
  `<id>\tabsent` 를 stdout에 내고, 그 id가 이 큐의 발급 상한(last_seq) 이하이면
  stderr에 history와 동일한 "발급되었으나 파괴되었을 수 있다" 보충 한 줄을
  추가한다(REQ-TAQ-004 선례).
- **REQ-TSP-003**: The show 동사 shall 읽기 전용이다 — `LoadPure` 경유,
  Mutate·adopt·lock 없음 — 그리고 기존 읽기 동사와 같은 stderr 고지
  (`discloseQueueLayout`: backlog.json 비권위 고지 + 스테일 로컬 스토어 발산
  고지)를 함께 나른다. show는 `todoLaneReadOnlyVerbs`(`todo.go:331-337`)에
  읽기 허용 동사로 등재되고, 오타 동사 가드의 동사 목록은 등록 트리에서
  파생된다(`todoVerbNames`, `todo.go:563-576` — 별도 수작업 목록 금지).

### B.2 항목 ② — `add` 본문 `-f` 시작 오파싱 수리

- **REQ-TSP-010**: **When** `moai todo add` 가 `-f`로 시작하는 단일 본문
  인자(따옴표로 묶인)를 받으면, the CLI shall 그 본문을 **그대로(verbatim)**
  본문으로 하는 카드를 생성한다 — 플래그 파싱이 본문을 소비해 "unknown
  shorthand flag"로 생성이 실패하는 today의 행동은 결함이다.
- **REQ-TSP-011**: The 수리 shall 기존 플래그 의미를 보존한다 — `--pick`,
  `--force`, `--classification-file`(`todo.go:615-620`)와 `--` 구분자 경로,
  그리고 부모 낙하 t69/오타 가드(t203)의 행동은 변경 없어야 하며, MCP 표면이
  같이 쓰는 `runTodoAddAppendRoot`(`todo.go:636`)의 시그니처 소비자는 영향
  받지 않는다.

### B.3 항목 ③ — list 기본 한도 완화

- **REQ-TSP-020**: The list 렌더의 기본 한도(`todoListDefaultLimit`,
  `todo.go:727`) shall 20에서 **100으로** 상향한다(확정값 — 2026-09-30
  Kickoff 운영자 결정). 100은 오늘의 live 큐(55행)를 잘림 없이 렌더하고, 다시
  잘릴 때는 기존 stderr withheld 한 줄(`todo.go:810-814`)이 그 사실을 밝힌다
  — withheld 고지 행동은 유지된다. 잘림 고지 설계(t403)는 유지된다.
- **REQ-TSP-021**: The list 의 기존 계약 shall 보존된다 — `--limit 0` 무제한,
  `--json`은 한도 무시(전체 판독), `--limit <0` 거절(`todo.go:753-755`),
  `--dropped` 전용 보기.

### B.4 항목 ⑥ — 렌더 완전성(누락 행 구조적 차단)

- **REQ-TSP-030**: **When** list(또는 bare todo)가 한 스토어에서 대답하면,
  the 렌더 shall 그 스토어의 행 중 렌더 필터(기본: live 전 상태 — hold 포함;
  `--dropped`: dropped)를 만족하는 행 집합과 **정확히 같은 집합**의 id를
  렌더한다 — 스토어에는 있는데 렌더가 누락하는 행은 상태와 무관하게 결함이다.
  이 불변식은 재현 불가능한 일회 관측(t1338)에 의존하지 않는 속성 검사로
  고정한다.
- **REQ-TSP-031**: The run 단계 shall t1338 관측(2026-09-29 worker-63)의 진단
  기록을 progress §E.2에 남긴다 — 재현 시도와 결과, 스키마 v1(3-상태 CHECK) 대
  v2(4-상태 CHECK) DB에서 hold 쓰기 경로의 행동 확인, 결론. 침묵으로 닫는 것은
  금지다.

### B.5 항목 ④ — 유령 저장소 소멸(확장 감지·정리 점검·첫 발견 1회 안내)

- **REQ-TSP-040**: The 단일 유령 검출기(`InspectStaleLocalStores`,
  `todo_stale_store.go:62-94`) shall SQLite `backlog.db` 발산 팩트에 **더해**
  비(非)SQLite 유령 클래스를 같은 팩트 구조에서 별개 사실로 보고한다 — (a)
  레거시 `backlog.json`(홈 `~/.moai/db/<key>/todo/` 와 프로젝트-로컬
  todo/kanban 디렉터리 양쪽), (b) `backlog.json.migrated`, (c) todo 상태
  디렉터리의 세션 레코드 JSON(`<uuid>.json`). 두 번째 검출기는 만들지 않는다
  (REQ-TSS-004 "no second inspector" 원칙 승계). 검출은 모든 분기에서 읽기
  전용이다 — 유령 파일을 열어 쓰거나, 마커를 새기거나, lock을 잡지 않는다.
- **REQ-TSP-041**: **When** 유령 클래스가 처음 발견되면, the 읽기·쓰기 동사의
  stderr 고지 shall 그 발견을 한 번만 안내하고, 안내 사실을 **살아 있는 상태
  디렉터리**의 마커에 기록해 이후 읽기에서 침묵한다 — 마커는 절대 유령 파일
  안/옆에 새겨지지 않는다. stdout(`--json` 포함)은 고지 유무와 무관하게 바이트
  동일이다(REQ-TSS-002/REQ-BJD-004 승계). `moai doctor`는 마커와 무관하게 계속
  보고한다.
- **REQ-TSP-042**: The doctor 점검 shall 유령 아티팩트 목록(경로·클래스·바이트
  크기)을 3상태(부재 OK / 존재 WARN / 판독 불가·모순 FAIL)로 판정하고,
  **상수 등록 + `namesAddedAfterBaseline` allowlist 항목 + 골든 스냅샷
  재생성**을 한 묶음으로 착지한다(REQ-TSS-011/012 규율 승계 — binary_lag 쌍
  없는 doctor 점검은 결함).
- **REQ-TSP-043**: The 저장소 문서
  (`internal/template/templates/.moai/docs/todo-queue-storage.md` — 현재 구
  레이아웃 `.moai/state/todo/backlog.db`를 산 큐로 기술) shall 홈 DB 정식
  레이아웃과 유령 아티팩트 표로 본문을 정정해 **템플릿 원본에서** 고치고, 그
  정정을 본 브랜치의 추적 파일에 반영한다. 이 문서는 본 트리에는 추적·클린
  상태로 존재하고(템플릿 미러와 동일 blob `007aa8c5`), primary 체크아웃에만
  크기·mtime이 다른 이본이 있다(5,911바이트/09-09 대 6,774바이트/09-29 —
  체크아웃 상태 발산; plan-auditor 2026-09-30 00:18 측정, 본 세션 stat
  재확인). 정정으로 두 표면의 모순이 해소됨을 plan/progress에 기록한다(침묵
  방치 금지).

### B.6 항목 ⑤ — 저장 라벨 어휘 통합

- **REQ-TSP-050**: The 시스템 shall 홈 DB `todo_runtime_assignments.owner_label`
  (`internal/kanban/todo_runtime.go:54-56`)의 레거시 표기 행을 정식 어휘로
  **한 번의 lock된 마이그레이션**으로 이관한다 — `lead` → leader 정식 표기,
  `worker-<n>`/`agent-<n>` → `lane-<n>` — 이관 전후 행수(453/13 실측 대비)를
  progress에 기록한다. 이관 대상은 이 열 하나뿐이다.
- **REQ-TSP-051**: The doctor 라벨 드리프트 점검 shall owner_label에 남은 레거시
  표기(`lead*`, `worker-<n>`, `agent-<n>` — `role.go:46`,
  `bootstrap.go:340-346`의 감지기 재사용)를 세어 보고한다 — 0행이면 OK. 이
  점검도 REQ-TSP-042와 같은 binary_lag 쌍 + 골든 묶음을 따른다.
- **REQ-TSP-052**: **While** 마이그레이션이 착지한 뒤, the owner_label 쓰기
  경로 shall 레거시 표기를 절대 새로 쓰지 않는다(정식 어휘만 기록).

## C. 성공 기준

1. `moai todo show t<n>` 이 한 줄로 카드 전문을 돌려주고(본문 절단 없음),
   없는 id는 `absent` + stderr 보충으로 대답한다.
2. 본문이 `-f`로 시작하는 카드가 `moai todo add "<text>"` 로 생성된다(RED →
   GREEN, run 단계에서 today의 실패 출력을 먼저 포착).
3. `moai todo` 기본 렌더가 55행 live 큐를 잘림 없이 보여준다.
4. 렌더 완전성 불변식이 속성 검사로 고정되고, t1338 관측의 진단 기록이
   progress에 존재한다.
5. 유령 아티팩트(json 계열 + 세션 JSON)가 단일 검출기 팩트에 잡히고, 읽기
   표면은 첫 발견 1회만 안내하며, `moai doctor`가 목록을 판정한다.
6. 저장 문서가 홈 DB 정식 레이아웃으로 정정돼 브랜치에 추적된다.
7. `todo_runtime_assignments.owner_label`의 레거시 표기가 0행이 되고, doctor
   라벨 점검이 그 사실을 재현 가능하게 판정한다.

## D. 제약

- 개발 언어/주석/godoc: 영어(`code_comments: en`). 사용자 대상 stderr 문자열:
  영어(기존 고지 문안 어조 — "NOT the queue" 계열 — 와 일치).
- 테스트는 `t.TempDir()` + 기존 시임(`kanban.HomeDirFn`, `paths.EnvHome`)으로
  홈 DB 격리. 실제 `~/.moai/db/...` 와 primary 체크아웃 `.moai/` 를 테스트가
  건드리지 않는다.
- 유령 아티팩트·라벨 행의 **삭제/파괴는 없다** — 판정과 안내까지가 코드 소관이고,
  실제 처분은 REQ-TSS-021(운영자+리드 확인 게이트)의 정신을 승계해 운영 행위로
  남긴다.
- doctor 신규 점검은 상수 등록 + allowlist + 골든이 한 커밋 묶음(HARD).
- TDD 우선: ②(오파싱)와 ⑥(완전성)은 RED 먼저 — today의 고장 출력을 구현 전에
  포착해 증거로 남긴다.

## E. Out of Scope

### Out of Scope — 유령 아티팩트의 자동 삭제·재동기화

- 유령 파일을 최신 큐로 재동기화하거나 코드가 자동 삭제하는 기능. 본 SPEC의
  코드는 감지·안내·판정까지이며, 삭제는 운영자+리드 확인 게이트 안의 운영
  행위다(SPEC-TODO-STALE-STORE-001 REQ-TSS-021 승계).

### Out of Scope — SQLite 유령 스토어 고지·발산·처분 자체

- 스테일 프로젝트-로컬 `backlog.db`의 stderr 고지, doctor 발산 점검, 잔존
  저장소 처분은 SPEC-TODO-STALE-STORE-001(카드 t1307)이 소유하고 이미 착지했다.
  본 SPEC은 그 검출기를 **확장**할 뿐 해당 요구를 다시 정의하지 않는다(§F.1).

### Out of Scope — owner_label 밖의 라벨 저장소

- `internal/homestate` cards 테이블의 `owner_label`
  (`internal/homestate/card_transition.go:620`), 역할 선언
  (`BoardDir/roles/`), workers 레지스트리 등 다른 저장소의 라벨 값은 이관
  대상이 아니다. 카드 본문 수치(453/13)가 정확히 대응하는
  `todo_runtime_assignments` 한정이다.

### Out of Scope — docs-site moai-todo 문서 페이지

- `docs-site/content/{en,ko,ja,zh}/utility-commands/moai-todo.md` 의 저장소
  서술 정합성은 별도 관심사로 남긴다(템플릿 문서 정정 이후 별도 카드 권장).

## F. 교차 참조

- SPEC-TODO-STALE-STORE-001(카드 t1307) — 스테일 SQLite 스토어 고지·doctor
  발산·처분 게이트. **경계 조정 노트는 §F.1.**
- SPEC-TODO-QUEUE-HOME-CANON-001 — 홈 DB 정식화(근원)
- SPEC-BACKLOG-JSON-DISCLOSURE-001(카드 t395) — backlog.json 비권위 고지,
  "no second inspector", stderr 전용 원칙
- SPEC-TODO-ARCHIVE-QUERY-001 — history 표면과 REQ-TAQ-004/007(show의 모델)
- SPEC-ROLE-NAMING-CODE-001 — 어휘 정식화와 REQ-RNC-009(본 SPEC이 이 카드
  범위에서 한정 예외를 두는 원칙)
- SPEC-TODO-CLAIM-LEASE-001(카드 t1342, draft) — 같은
  `todo_runtime_assignments` 표를 쓰는 후속; 마이그레이션은 그 착지 전이라도
  열 구조를 바꾸지 않는다(값만 이관).
- 카드 t1251 — binary_lag 쌍 규율 선례 / 카드 t1308 — hold 상태 도입

### F.1 카드 t1307 경계 조정 노트 (리드 중재용 — 흡수 아님)

큐에 남아 있는 카드 t1307 본문("스테일-로컬-스토어 고지 확장 + doctor 감지 +
잔존 처분")과 본 SPEC ④는 겹쳐 보인다. 2026-09-29 기준 사실: SPEC-TODO-STALE-STORE-001은
이미 `completed`이고 그 ①(다섯 읽기 동사 고지)·②(doctor 발산 점검)·③(0바이트 쌍 +
마이그레이션 백업의 확인 게이트 처분)은 코드로 착지돼 있다
(`todo_stale_store.go`, `doctor_todo_store.go`, `binary_lag_test.go:222-225`
allowlist 항목). 본 세션 실측에서 프로젝트-로컬 `backlog.db`는 이미 부재한다.
따라서 본 SPEC ④가 새로 소유하는 것은: (a) 비SQLite 유령 클래스(json 계열,
세션 JSON — 홈 쪽 포함)의 감지 확장, (b) 정리 doctor 점검, (c) 첫 발견 1회
안내, (d) 저장 문서 정정. SQLite 고지·발산·처분과 **유령 파일 삭제 행위 자체**
는 t1307 소관으로 남는다. 리드가 카드 t1307을 본 SPEC 흡수로 처리할지
폐기할지는 본 노트를 근거로 중재한다 — 본 SPEC은 어느 쪽이든 전제하지 않는다.
