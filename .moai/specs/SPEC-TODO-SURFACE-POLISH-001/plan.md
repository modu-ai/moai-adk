# plan.md — SPEC-TODO-SURFACE-POLISH-001 (카드 t1349)

## §A Context

정식 큐는 홈 DB `~/.moai/db/<project-key>/todo/backlog.db` 하나다
(`internal/cli/todo.go:221`, `todo.go:74-76`). 본 SPEC은 카드 t1349의 표면 정비
5종(①show ②add -f ③list 한도 ④유령 저장소 ⑤라벨 어휘)과 렌더 누락 점검 ⑥을
단일 SPEC으로 정비한다. 모든 file:line 인용은 워크트리 t1349 @ `51abf337a`
기준 2026-09-29 본 세션이 읽고 확인한 값이다.

## §B Known Issues

- show 동사 부재 — 등록부 `todo.go:303-309`에 없음(`grep '"show"'` 0, 양성
  대조 `todo.go:652`). 가드 주석은 이미 `show 401`을 예시 인용(`todo.go:481`).
- `--force` 숏핸드 미등록 — `todo.go:617-618`(`BoolVar`), `BoolVarP` 전무.
  pflag interspersed 파싱이 `-f` 시작 본문을 소비 → 생성 실패(카드 ②).
- `todoListDefaultLimit = 20`(`todo.go:727`) — live 55행(queued 44 + picked
  11, 2026-09-29 실측) 매일 잘림. withheld 고지는 `todo.go:810-814`에 이미 있음.
- 유령 아티팩트 실측(2026-09-29): 홈 `backlog.json` 1,884,049B(09-26),
  프로젝트-로컬 `backlog.json` 652,478B(09-10), `backlog.json.migrated`
  155,043B, 세션 JSON 약 397개. 프로젝트-로컬 `backlog.db`는 부재(stat 확인 —
  리드 전달 사실("로컬 652KB backlog.db")의 652KB는 오늘 디스크에서
  `backlog.json`이고, SQLite 유령은 이미 소멸).
- 검출기 사각지대 — `InspectStaleLocalStores`(`todo_stale_store.go:62-94`)는
  SQLite `backlog.db` 2경로만 스캔; json 계열·세션 JSON·홈 쪽 json은 안 보임.
- 라벨 어휘 — `todo_runtime_assignments.owner_label`에 `lead` 453 /
  `worker-67` 13(2026-09-29 실측, 카드 수치와 일치). 정식 어휘는
  `role.go:41`(`leader`)·`bootstrap.go:246`(`lane-<n>`). REQ-RNC-009
  "레코드 비재작성" 원칙과의 긴장은 spec.md §B.6/§F.1에 화해 기록.
- 렌더 누락 — t1338 관측(09-29 worker-63)은 오늘 재현 불가
  (`grep -c t1338` = 4). 구조적 단서: 코드 DDL 4-상태 CHECK
  (`backlog_sqlite.go:128`, v2 `:53`) vs 실측 홈 DB v1 3-상태 CHECK — hold
  쓰기 경로가 v1 DB에서 어떻게 행동하는지 미규정.
- `.moai/docs/todo-queue-storage.md`(템플릿 원본
  `internal/template/templates/.moai/docs/todo-queue-storage.md`)가 구
  레이아웃을 산 큐로 기술 — 홈 DB 정식화와 모순. 이 문서는 본 트리에는
  추적·클린으로 존재하고(템플릿 미러와 동일 blob), primary 체크아웃에만
  크기·mtime이 다른 이본이 있다(5,911바이트/09-09 대 6,774바이트/09-29 —
  체크아웃 상태 발산; plan-auditor 2026-09-30 00:18 측정, 본 세션 stat
  재확인).

## §C Pre-flight (조사 결과 — 구현이 인용할 파일)

| 대상 | 위치 |
|---|---|
| 동사 등록부 | `internal/cli/todo.go:303-309` `cmd.AddCommand(... 22종)` — show 추가 지점 |
| 레인 읽기 허용 목록 | `internal/cli/todo.go:331-337` `todoLaneReadOnlyVerbs` — `"show": true` 추가 |
| 오타 동사 가드 | `todo.go:428-446` `todoMistypedVerbGuard`; 동사 목록 파생 `todo.go:563-576` `todoVerbNames` — 트리 파생이라 무수정 |
| id 정규화 | `todo.go:1256` `normalizeTodoRef` — show가 재사용 |
| history 조회 기계(show 모델) | `internal/cli/todo_history.go:56-98`(동사), `:223-253`(`renderTodoHistoryLookup`), `:51`(기본 한도 20 선례) |
| 셀 평탄화 | `internal/cli/todo_pr.go:361` `todoPRCell` — 탭/개행 치환, 절단 없음 |
| add 플래그 | `todo.go:585-622` `newTodoAddCmd`(`:615-620` 플래그 선언), `:628-680` `runTodoAddAppend`/`Root`(MCP 공유 — `:636`) |
| list 한도·필터 | `todo.go:727`(상수), `:745-816` `runTodoList/Root`(`:778-793` 필터+한도, `:810-814` withheld) |
| 유령 검출기 | `internal/kanban/todo_stale_store.go:62-94` `InspectStaleLocalStores`, `:102-130` `readStaleStoreLastSeq`(읽기전용 판독 패턴) |
| 고지 표면 | `internal/cli/todo_disclosure.go:45-68` `discloseStaleLocalStores`, `:93-100` `discloseQueueLayout`(읽기 동사 공용 진입) |
| doctor 등록 | `internal/cli/doctor.go:226`(`todoStoreDivergenceCheckName` 등록 행), `:272`(`factoryRunCheckName` 라벨 점검 선례), 등록부는 `runGroupedChecksObserved` 계열 슬라이스 |
| binary_lag 쌍 | `internal/cli/binary_lag_test.go:198-226` `namesAddedAfterBaseline`(상수 등록은 bare 키), `:250` `TestBinaryLag_AllowlistKeysAreLiveNames`, `:298` `TestBinaryLag_DoctorCheckNameSetIsUnchanged` |
| doctor 골든 | `internal/cli/doctor_golden_test.go` + `testdata/*.golden` — `UPDATE_GOLDEN=1` 재생성 |
| 라벨 표 DDL | `internal/kanban/todo_runtime.go:54-56`(`todo_runtime_assignments`), `:32`(`OwnerLabel` 필드) |
| 어휘 상수·감지기 | `internal/kanban/role.go:41,46,52-54`; `internal/kanban/bootstrap.go:246,252-255,320-330,340-346` |
| hold 상태 | `internal/kanban/backlog_store.go:71`; 쓰기 `internal/cli/todo_hold.go:87`; DDL `internal/kanban/backlog_sqlite.go:128`(4-상태), v2 `:53` |
| 저장 문서 | `internal/template/templates/.moai/docs/todo-queue-storage.md`(템플릿 원본) |
| 큐 루트 해석 | `todo.go:74-76`(adopting), `todo.go:113-116`(읽기 순수) |

## §D Constraints

- 개발 모드: quality.yaml 구성 따름(현 ddd/tdd). ②·⑥은 RED 먼저(구현 전
  고장 출력 포착 — verification-completeness §2 두 칸 채택).
- 테스트 격리: `t.TempDir()` + 홈 시임 주입. 실제 홈 DB·primary `.moai/` 금지.
- 유령 아티팩트 삭제 금지 — 감지·안내·판정까지가 코드 소관.
- doctor 신규 점검 = 상수 등록 + allowlist bare 키 + 골든 재생성 한 커밋(HARD).
- stdout 바이트 동일성 — 고지는 stderr 전용(REQ-TSS-002 승계).
- 사용자 대상 stderr 문안 영어, 기존 어조 일치.
- [RESOLVED 2026-09-30 — Kickoff 운영자 결정] list 기본 한도 = **100**
  (stderr withheld 고지 행동 유지). 후보 50/0은 기각 — 50은 현 live 55행을
  여전히 자르고, 0(무제한 기본)은 t403 경계 고지 철학과 긴장.
- [RESOLVED 2026-09-30 — Kickoff 운영자 결정] owner_label 이관 목표 =
  **leader/lane-<n>**(role.go:41·bootstrap.go:246 정식값). 카드 본문 표기는
  "lead/lane-N"이라 본 SPEC은 여기서 **의도적으로 발산**한다 — 발산 근거(lead는
  REQ-RNC-009상 감지 전용 레거시, 재도입이 모순)와 중재 창구는 spec.md §F.1
  관계 노트가 담당한다.
- [RESOLVED 2026-09-30 — Kickoff 운영자 결정] show 출력 = **history live 행
  확장형**(탭 구분, 절단 없는 전문 본문이 마지막 필드). 멀티라인 블록 대안은
  기각 — 기계 파싱 우선.
- [RESOLVED 2026-09-30 — Kickoff 운영자 결정] ② 수리 기제 = **add 경로
  DisableFlagParsing + 알려진 플래그 수동 스캔**(--pick/--force/
  --classification-file/-- 존중). -f 숏핸드 등록은 기각 — 단독으로는 본문을
  보존하지 못한다(add "-f x"가 플래그+잉여 토큰으로 갈라져 본문 훼손).

## §E Self-Verification

M1..M3 각 종료 시 변경 패키지 테스트 재측정 + acceptance.md AC 전수 재판정.
전체 스위트는 CI 몫(레인 규율 — 로컬 `go test ./...` 금지). §E 항목은 VCI §3
5-섹션 형식(Claim/Evidence/Baseline-attribution/Gaps/Residual-risk)으로 보고.

## §F Milestones (결정 변동 가능성 순 — 설계 결정이 위, 기계 작업이 아래)

### M1 (High) — CLI 표면: show + add -f 수리 + list 한도 + 렌더 완전성 (①②③⑥)

TDD 우선 순서:

1. **RED(②)**: `moai todo add "-f hello world"`가 오늘 실패함을 고정 테스트로
   포착(시임 주입 홈). 실패 출력 그대로 progress §E.2에 전사.
2. **RED(⑥)**: 렌더 완전성 속성 테스트 — 무작위/고정 fixture 스토어(queued,
   picked, dropped, hold 행 포함)에서 렌더 id 집합 ≠ 필터 만족 집합인 변이를
   하나 심어 검사가 잡는지 먼저 확인.
3. **GREEN(②)**: §D의 기제 결정(기본: add DisableFlagParsing + 수동 플래그
   스캔)으로 본문 verbatim 생성. `--pick`/`--force`/`--classification-file`/
   `--`/t69 낙하/t203 가드 회귀 유지.
4. **GREEN(①)**: `internal/cli/todo_show.go` 신설 — history 조회 기계 재사용,
   전문 본문(마지막 필드), absent + stderr 보충, `discloseQueueLayout` 고지,
   `AddCommand` 등록 + `todoLaneReadOnlyVerbs` `"show": true`.
5. **GREEN(③)**: `todoListDefaultLimit` 상향(§D 확정값) + withheld 고지 유지.
6. **GREEN(⑥)**: 불변식 검사 통과 + 기존 list 회귀 전수.

AC: AC-TSP-001(a-c), AC-TSP-002, AC-TSP-010, AC-TSP-011, AC-TSP-020,
AC-TSP-021, AC-TSP-030.

### M2 (High) — 유령 저장소: 감지 확장 + 1회 안내 + doctor 정리 점검 + 문서 정정 (④)

1. `internal/kanban/todo_stale_store.go` — 팩트 구조에 비(非)SQLite 클래스
   필드 추가(backlog.json 홈+로컬, .migrated, 세션 JSON), SQLite 발산 판정
   로직 무변경, 전 분기 읽기 전용 유지(REQ-TSP-040).
2. `internal/cli/todo_disclosure.go` — 첫 발견 1회 안내: 마커는 살아 있는 상태
   디렉터리에(유령 옆 금지), stderr 전용, stdout 바이트 동일(REQ-TSP-041).
   `discloseQueueLayout` 진입 유지 — 읽기 동사별 회귀는 AC-TSS-001 선례의
   교훈대로 동사별 판정.
3. `internal/cli/doctor_todo_store.go`(또는 형제 파일) — 유령 목록 점검:
   경로·클래스·바이트 크기, 3상태 판정. 상수 등록(`doctor.go` 등록부) +
   `namesAddedAfterBaseline` bare 키 추가 + 골든 재생성 — **같은 커밋**
   (REQ-TSP-042).
4. `internal/template/templates/.moai/docs/todo-queue-storage.md` 본문 정정
   (홈 DB 정식 레이아웃 + 유령 아티팩트 표) 후 본 브랜치의 추적 파일에 반영;
   primary 체크아웃 이본(체크아웃 상태 발산)과의 모순 해소 기록(REQ-TSP-043).
5. 진단 기록: t1338 관측 재현 시도 + v1/v2 CHECK-제약에서 hold 쓰기 행동 확인 →
   progress §E.2(REQ-TSP-031).

AC: AC-TSP-031, AC-TSP-040, AC-TSP-041, AC-TSP-042, AC-TSP-043.

### M3 (Medium) — 라벨 어휘: owner_label 일회 이관 + doctor 드리프트 점검 + 쓰기 정규화 (⑤)

1. 마이그레이션: `todo_runtime_assignments.owner_label` 값 일회 이관(§D 확정
   목표 표기) — lock된 단일 쓰기, 전후 행수 기록, 다른 표/열 불변(바이트
   비교 검증). `todo_runtime.go` 소속으로 구현(REQ-TSP-050).
2. doctor 라벨 드리프트 점검: 레거시 표기 카운트(`role.go:52-54`,
   `bootstrap.go:340-346` 감지기 재사용), 0행 OK. 상수 등록 + allowlist +
   골든 한 커밋(REQ-TSP-051).
3. 쓰기 경로 정규화: owner_label 기록부가 정식 어휘만 쓰도록 상수 경유 강제
   (REQ-TSP-052).
4. `internal/homestate` cards 표 등 타 저장소는 건드리지 않음(spec §E Out of
   Scope).

AC: AC-TSP-050, AC-TSP-051, AC-TSP-052.

## §G Anti-Patterns

- 두 번째 유령 검출기를 만드는 것 — REQ-TSS-004 "no second inspector" 위반.
  확장은 `InspectStaleLocalStores` 안에서.
- 안내 마커를 유령 파일 안/옆에 새기는 것 — 유령은 롤백 증거다(REQ-TSS-003).
- doctor 점검을 문자열 리터럴로 등록해 allowlist 키 모양을 갈라뜨리는 것(t1251).
- `-run` 패턴 비공허화 — binary_lag/골든 명령은 anchored 열거형, 실행 시점에
  최소 1 적중(t1307 plan-audit D3/D4 선례).
- stdout에 고지 섞기 — foreman 기계 표면 오염(REQ-BJD-004).
- `lead`를 정식값으로 재도입하는 것 — REQ-RNC-009 모순(§D NEEDS CLARIFICATION
  참조).
- 테스트가 실제 홈 DB를 보는 것 — 시임 주입 필수.

## §H Cross-References

- spec.md §B 요구사항 전수, acceptance.md AC 전수
- `.moai/specs/SPEC-TODO-STALE-STORE-001/` — 확장 기반이 되는 검출기·doctor·
  처분 게이트의 원본 계약 (t1307 경계 조정 노트: spec.md §F.1)
- `.claude/rules/moai/development/verification-completeness.md` §2 — RED-now
  두 칸 채택 근거
- `.claude/rules/moai/core/verification-claim-integrity.md` — 진단·이관 기록의
  증거 규율
