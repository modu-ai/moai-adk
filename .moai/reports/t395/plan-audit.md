# SPEC Review Report: SPEC-BACKLOG-JSON-DISCLOSURE-001

카드: t395 · 트리 `.claude/worktrees/t395` · 브랜치 `WT-stale-backlog-json` @ `ea97eec22`
Iteration: 1/2 (Tier M ceiling)
**Verdict: PASS-WITH-DEBT**
**Overall Score: 0.85** (Tier M PASS threshold 0.80)

M1 Context Isolation: 저자 추론 맥락은 전달받지 않았다. 감사 입력은 Tier M 계약대로
`spec.md` + `plan.md` + `acceptance.md`(+ `progress.md`)와, 리드가 지정한 증거 3건
(`premise-verdict.md`, `reader-surfaces.md`, `r1-repro.md`)뿐이다.

트리 상태 확인: 감사 시작 시 `git rev-parse HEAD` = `ea97eec228aa9afef5eeed97971f826e3ac3bc51` —
리드가 못박은 `ea97eec22`와 일치. HEAD 이동 없음.

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — `REQ-BJD-001`..`REQ-BJD-014` 14개, 결번·중복 없음, 3자리 zero-pad 일관.
  측정: `grep -o "REQ-BJD-[0-9]*" spec.md | sort -u` → 14개 연속. `grep -c` 총 등장 14회(정의당 1회).
- **[PASS] MP-2 GEARS 형식 준수** — 판정 레이어는 `spec.md` §B의 **REQ-XXX 요구 레이어**에 대해 내렸다
  (`acceptance.md`의 Given-When-Then AC는 검증 레이어이므로 M3 §Scope에 따라 Group 4에서 채점).
  14개 전부 5개 GEARS 패턴 중 하나에 부합: Ubiquitous 6(001,006,007,011,013 + 013), Event-driven 4(002,008,014 + 008),
  State-driven 1(010), Where 1(003), Unwanted 4(004,005,009,012). 비형식 문장 0건.
  단, REQ-BJD-003의 레이블 정확성은 D6 참조(형식은 부합, 레이블은 부정확).
- **[PASS] MP-3 YAML frontmatter 유효성** — 정본 12필드 전부 존재·타입 정합:
  `id`/`title`/`version`("0.1.0" quoted)/`status`(draft)/`created`(2026-09-02)/`updated`(2026-09-02)/
  `author`/`priority`(P2)/`phase`/`module`/`lifecycle`(spec-anchored)/`tags`(CSV string).
  거부 대상 snake_case alias(`created_at`/`updated_at`/`labels`/`spec_id`) 0건.
  추가 필드 `tier: M`, `related_specs` 는 스키마 위반이 아니다.
- **[N/A] MP-4 §22 언어 중립성** — 단일 언어(Go) 프로젝트 내부 SPEC. 16개 프로그래밍 언어 도구 체인을
  다루지 않으므로 해당 없음 → auto-pass.
- **[PASS] MP-5 D7 cross-SPEC 조정** — 참조된 4개 SPEC 전부 `.moai/specs/` 에 실재하며 상태는
  `SPEC-TODO-SQLITE-001: completed` · `SPEC-TODO-ARCHIVE-QUERY-001: completed` ·
  `SPEC-KANBAN-TODO-CLI-001: in-progress` · `SPEC-TODO-LANDING-STATE-001: completed`.
  `retired`/`superseded`/`archived` 0건 → BLOCKING 없음.
- **[PASS] MP-6 D8 크로스 플랫폼 규율** — `grep -rn syscall .moai/specs/SPEC-BACKLOG-JSON-DISCLOSURE-001/`
  → 0건. D8-4에 따라 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' plan.md` → 0건
  (`research.md` 는 Tier M 이라 부재, 이는 계약대로).

**Must-Pass 7개 중 실패 0건.**

---

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.80 | 0.75 | 서술이 이례적으로 정밀하고 자기 반증적. 감점은 D3(동사 범위 3중 불일치)와 D6(사이트 4 vs 파일 4). |
| Completeness | 0.90 | 1.0 근접 | HISTORY·§A Context·§B Requirements·§C Exclusions(H3 4개 각각 `-` bullet 보유)·acceptance.md 전부 존재. 감점은 D2가 남긴 체크 집합의 공백. |
| Testability | 0.70 | 0.75 미만 | 4개 AC가 기계 검증 불가이거나 공허 통과 가능(D1/D2/D4/D5). "one AC"(0.75)보다 많고 "대부분 주관적"(0.50)보다는 훨씬 적다. |
| Traceability | 1.00 | 1.0 | 직접 계수: REQ 14개 전부 ≥1 AC 보유, AC 16개 전부 §D.6 표에 등장, 고아 0·미커버 0. |

집계 = (0.80 + 0.90 + 0.70 + 1.00) / 4 = **0.85** ≥ 0.80 (Tier M).

---

## 리드가 지목한 7개 압박 지점 — 판정

### 1. AC-BJD-008 — 관측된 R1 red 를 회귀로 전환하는가 · **성립한다**

배송 스크립트(`.claude/skills/moai-kanban-foreman/SKILL.md:95`)를 직접 읽어 확인했다.
분기하는 값은 `[ -f "$f" ]` 와 `cksum "$f"` 둘뿐이고 `f=.moai/state/todo/backlog.json` 이다.
AC-BJD-008의 fixture 는 "`backlog.db` 가 정본이고 `backlog.json` 이 **없는**" 레이아웃 —
즉 `cur` 이 영구히 `missing` 으로 고정되어 `[ "$cur" != "$last" ]` 가 첫 반복 이후 영원히 거짓이 되는
바로 그 상태다. 수리 전 red 는 `r1-repro.md` §B 에서 **관측**됐고(JSON-WATCH 45초 0건 vs
DB-WATCH 대조군 발화), 수리 후 green 은 같은 대조군이 이미 시연했다.

따라서 **두 감시 대상에 대해 동시에 통과하는 공허한 기준이 아니다.** 명시된 falsifiability
조건("watch target 을 `backlog.json` 으로 되돌리면 이 기준은 fail 해야 한다")은 관측에 근거하며
`plan.md` §G 가 "정적 grep 으로 대체 불가"까지 못박아 두었다. 리드가 가장 의심한 지점은 견고하다.

단 하나의 결함은 재바인딩 문구다 — D1 참조.

### 2. AC-BJD-009 — 추론 출처가 기록됐고 기계 검증 가능한가 · **양쪽 다 성립한다**

출처 기록 검증: `r1-repro.md` Gaps 절에 "Case B(스테일 json 이 존재하는 이 저장소의 현재 상태)는
Monitor 로 재현하지 않았다 … 관측은 Case A 만 했다" 가 실제로 존재한다. AC-BJD-009 의
"Recorded provenance: Case B was **reasoned but not observed**" 는 원문에 정확히 대응한다.
상속이 아니라 의도적 재주장이라는 서술도 사실에 부합한다.

기계 검증 가능성: Given(정적 json 병존) / When(창 안 변경) / Then(이벤트 관측) 전부 이진 판정이다.
수리 전 반증도 성립한다 — 정적 json 의 cksum 은 불변이므로 이벤트 경로가 없다. 공허하지 않다.

### 3. AC-BJD-010 — 실측인가 탈출구인가 · **절반만 막혔다. 이것이 D4다**

먼저 전제부터 검증했다. `internal/kanban/backlog_sqlite.go:229` 가 `_pragma journal_mode(WAL)` 를
설정하고 `backlog_sqlite.go:13` 이 "WAL journal mode so readers never block"을 명시한다.
**REQ-BJD-010 의 전제는 실재한다** — 가상의 위험을 막는 기준이 아니다. 이 점은 SPEC 에 유리하다.

리드가 지목한 안티패턴(관측 창을 넓혀 통과시키기)은 **명시적으로 봉쇄돼 있다**: 기준문 자체가
"not by widening the window until it happens to pass" 라고 적었고 `plan.md` §G 가 이를
안티패턴으로 재차 못박았다. 이 축은 막혔다.

그러나 **더 개연성 높은 공허 통과 경로가 열려 있다.** Given("commit 됐으나 아직 checkpoint 되지
않은 상태")을 **어떻게 성립시키고 어떻게 성립했음을 확인할지** 기준이 말하지 않는다.
`r1-repro.md` 의 대조군은 변경 직후 `backlog.db` cksum 이 실제로 움직였다고 기록했다
(1125950968 → 1188524958) — 즉 그 실행에서는 지연이 **일어나지 않았다.** run-phase 가 그냥
`moai todo add` 후 감시를 돌리면 checkpoint 가 이미 끝난 상태를 재게 되고, 기준은
**Given 이 한 번도 성립하지 않은 채 통과**한다. 창을 넓히지 않고도 공허해질 수 있다.

acceptance.md 서문이 스스로 세운 원칙("shipped tree 에 대해서도 통과하는 기준은 아무것도 주장하지
않는다")이 이 기준에서 지켜지지 않는다. 탈출구는 맞다 — 다만 SPEC 이 경계한 그 탈출구가 아니다.

### 4. REQ-BJD-006 / AC-BJD-007 재사용 제약 — 검증 가능한가 · **실체는 맞으나 체크는 희망사항이다**

실체 검증부터: `InspectBacklogArchiveVouch`(`backlog_archive_vouch.go:49`)가 실제로 저장소 정체를
이름으로 반환하고(`:53`,`:55`,`:57`), `inspectBacklogLayout`(`backlog_migrate.go:392,395`)이
`jsonExists` 를 이미 계산하며, SQLite 분기(`:57`)가 그 값을 **실제로 버린다**. §A.4 의
"이미 측정되고 나서 버려진다"는 주장은 소스로 확인된다. 경합 후보도 확인했다 —
`BacklogStore.openEngine`(`backlog_store.go:592`)이 State D 를 해소하긴 하지만 저장소 정체를
**보고하지 않고** 엔진을 반환할 뿐이다. 따라서 "유일한 보고자" 주장은 실질적으로 참이다.

문제는 체크 문구다. AC-BJD-007 은 "When `internal/kanban/` **is searched for** store-identity
reporting" 이라고만 쓴다 — 검색 패턴도, 명령도 없다. 이것은 전칭 부정("유일하다")을 사람의 판단에
맡기는 형태이며, acceptance.md 서문의 "Every criterion below is binary and machine-checkable"을
스스로 위반한다. 같은 문서의 AC-BJD-015 가 정확한 `grep -rn` 을 제시하는 것과 대비된다. → D2.

### 5. Traceability — 직접 계수 · **완전하다**

REQ 14개: 001→AC-001 · 002→AC-002 · 003→AC-004 · 004→AC-003 · 005→AC-005,006 · 006→AC-007 ·
007→AC-008,011 · 008→AC-008,009 · 009→AC-008(falsifiability),009 · 010→AC-010 · 011→AC-013 ·
012→AC-012 · 013→AC-014 · 014→AC-011,015,016. **미커버 REQ 0건.**
§D.6 표가 참조하는 AC 를 중복 제거하면 AC-BJD-001..016 전부 — **고아 AC 0건.**

### 6. 범위 규율 · **어긋난 요구 없음**

금지 축 4개를 REQ 14개 전부에 대해 대조했다.
- 삭제/청소: REQ-BJD-005 가 정면으로 **금지**한다("shall not delete, truncate, rename, move, or rewrite").
  §C 첫 Out of Scope 가 근거(미특정 작성자)까지 서술한다. 드리프트 없음.
- 저장 엔진/마이그레이션/downgrade 경로: REQ-BJD-005 가 "shall not run a migration, DDL, or take
  the queue lock"으로 봉쇄. §C 두 번째 Out of Scope 가 `export-json` 의 canonical 경로 쓰기를
  **의도적**이라고 보존한다. 드리프트 없음.
- 큐 스키마: §C 가 "items-table landed-evidence column axis belongs to card t359"로 명시 위임.
  스키마를 건드리는 REQ 없음. 드리프트 없음.
- `quarantineLegacyBacklog` / State D 해소: §A.1 이 "Neither is a defect and neither is touched"로
  선언하고 §C 가 재확인. 소스 확인 결과 두 경로 모두 설계대로 동작(`backlog_migrate.go:558`,
  `backlog_store.go:592-604`). 드리프트 없음.

범위 규율은 이 SPEC 의 가장 강한 면이다. 배제가 근거와 함께 서술돼 있고, 카드 원 문면이 지시한
"청소"를 반증 근거로 **거부**한 것이 특히 그렇다.

### 7. Tier M — 양방향 도전 · **M 이 맞다**

**아래로(S 주장) 기각.** Tier S 는 단일 통과·좁은 범위를 전제한다. 이 SPEC 은 M1~M5 5개
마일스톤이 Go 2개 패키지(`internal/kanban`, `internal/cli`) + 쉘 감시 블록 + 문서 3개 문장 +
템플릿 미러 + `make build` 로 갈라진다. AC 16개는 Tier S 상한을 명백히 넘는다.

**위로(L 주장) 기각.** CLAUDE.md §4 의 휴리스틱(마일스톤 ≥3 **그리고** 파일 ≥10)에 근접하는 것은
사실이다 — 접촉 파일은 `backlog_archive_vouch.go`, cli 명령 파일, foreman SKILL.md,
moai/SKILL.md, workflows/todo.md + 템플릿 미러 3개 ≈ 9~10개. 그러나 그중 6개는 문장 교정과
바이트 미러이고, 되돌리기 비용이 있는 설계 결정은 **struct 필드 1개와 watch target 1개**뿐이다.
`design.md`/`research.md` 를 요구할 아키텍처 표면이 없다. Tier M 의 3파일 산출물 집합이 정확하다.

`progress.md` §E.1 의 자기 선언(요구 14/상한 16, AC 16/상한 16)도 Tier M 예산과 정합한다.

---

## 자기 선언 3개 미명세 지점 — 진짜 처리됐는가, 라벨을 쓴 미해결 위험인가

| 지점 | 판정 | 근거 |
|---|---|---|
| **미특정 작성자** | **진짜 처리됨** | 미해결 사실을 결함으로 두지 않고 **설계 근거로 전환**했다. §A.3 이 "일회성 청소는 미특정 작성자의 재발을 막지 못한다 → 읽는 쪽 공시가 작성자와 무관하게 유효한 유일한 방어"라고 논증하고, §C 가 그 결론으로 청소를 배제한다. Gap 이 SPEC 의 방향을 **결정**했다. 라벨이 아니다. |
| **WAL 지연** | **부분 처리** | 전제는 실재(WAL 확인). 창 넓히기 탈출구는 봉쇄. 그러나 Given 성립 기법이 없어 **다른 공허 경로가 열려 있다** → D4. 라벨이 절반은 위험을 덮고 있다. |
| **동사 범위** | **라벨을 쓴 미해결 위험** | `plan.md` M3 이 "Which verbs carry it is a scoping decision to state explicitly in the run-phase report"로 **결정을 run-phase 로 미룬다**. 그런데 REQ-BJD-002 는 이미 규범적으로 "a `moai todo` invocation"(전 동사)이라고 못박았고, AC-BJD-002 는 "a `moai todo` read command"(좁음)만 검증한다. 규범과 검증이 어긋난 채 결정이 미뤄졌다 → D3. |

---

## Defects Found

**D1** — `acceptance.md` AC-BJD-008 재바인딩 문구 — **Severity: minor · Class: blocking**
"only its **queue path** rebound to the fixture"는 스크립트가 `f=` 변수 **하나**를 갖는다고 전제한다.
그런데 AC-BJD-010 이 명시적으로 허용하는 수리안("the criterion is met by a watch target that
covers the deferral")은 감시 대상이 `backlog.db` + `backlog.db-wal` **둘**이 되는 경우다.
그 수리안이 채택되면 AC-BJD-008 의 재바인딩 지시가 무엇을 가리키는지 미정이 되고, 두 기준이
서로를 무력화한다.
**Required fix**: 재바인딩 대상을 "the queue **directory** the script's watch target(s) resolve
against"로 바꾸어, 단일/복수 target 양쪽에 대해 지시가 성립하게 한다.

**D2** — `acceptance.md` AC-BJD-007 이 기계 검증 불가 — **Severity: major · Class: blocking**
"When `internal/kanban/` is searched for store-identity reporting"에 검색 패턴도 명령도 없다.
전칭 부정("유일한 보고 함수")을 사람 판단에 맡기는 형태이며, 같은 문서 서문의
"Every criterion below is binary and machine-checkable" 및 AC-BJD-015 가 세운 선례
(정확한 `grep -rn`)와 어긋난다. REQ-BJD-006 은 리드가 지목한 재사용 제약의 **유일한** 체크이므로,
이 상태로는 제약이 희망사항으로 남는다.
**Required fix**: 실행 가능한 명령으로 대체한다. 예:
`grep -rn 'BacklogStore\(SQLite\|LegacyJSON\|None\)' internal/kanban/ | grep -v _test.go` 의
정의 위치가 `backlog_archive_vouch.go` 한 파일에 국한될 것, 그리고
`grep -c 'func Inspect.*Backlog.*\(Store\|Vouch\)' internal/kanban/*.go` 합이 1일 것.

**D3** — REQ-BJD-002 / AC-BJD-002 / `plan.md` M3 의 3중 범위 불일치 — **Severity: major · Class: blocking**
같은 요구에 대해 세 산출물이 서로 다른 범위를 말한다: REQ-BJD-002 "a `moai todo` **invocation**"(전
동사) · AC-BJD-002 "a `moai todo` **read command**"(읽기만) · `plan.md` M3 "**at minimum** the read
surface"(하한만, 결정은 run-phase). BLOCKING 기준인 AC-BJD-002 가 규범 요구를 **과소 검증**하며,
어느 범위가 정본인지 어떤 산출물도 답하지 않는다.
**Required fix**: REQ-BJD-002 의 범위를 하나로 확정한다. 읽기 표면으로 좁히는 편을 권한다 —
쓰기 동사는 이미 락을 잡고 stdout 계약이 다르며, `plan.md` M3 의 "at minimum"과도 정합한다.
확정 후 AC-BJD-002 의 문구를 REQ 와 일치시킨다.

**D4** — AC-BJD-010 의 Given 성립 기법 부재 — **Severity: major · Class: blocking**
"commit 됐으나 아직 checkpoint 되지 않은" 상태를 **어떻게 만들고 어떻게 성립을 확인할지**
기준이 말하지 않는다. `r1-repro.md` 대조군은 그 실행에서 지연이 **일어나지 않았음**을 기록했으므로
(db cksum 이 즉시 이동), 소박한 run-phase 실행은 Given 이 한 번도 성립하지 않은 채 통과한다.
SPEC 이 봉쇄한 탈출구(창 넓히기)와는 **다른** 공허 경로다.
**Required fix**: Given 에 성립 기법과 그 확인을 넣는다. 예: `PRAGMA wal_autocheckpoint=0` 으로
자동 checkpoint 를 끈 fixture 에서 변경하고, 감시 전에 `backlog.db-wal` 의 크기가 0이 아니며
`backlog.db` 의 cksum 이 변경 전과 **동일함**을 먼저 확인할 것 — 그 두 값이 Given 이 성립했다는
증거다. 이 확인이 실패하면 기준은 통과가 아니라 Gap 이다.

**D5** — AC-BJD-015 의 grep 이 4번째 사이트의 미러를 구조적으로 못 본다 — **Severity: major · Class: blocking**
AC-BJD-015 는 BLOCKING 이며 템플릿 미러의 유일한 완전성 체크인데, 패턴이
`state/todo/backlog\.json` 이다. 그런데 §A.2 가 세는 4번째 사이트는 홈 폴백
`~/.moai/todo/<project-key>/backlog.json` 이고, **이 문자열은 그 패턴에 매칭되지 않는다.**
직접 측정으로 확인했다:
```
$ grep -rn 'state/todo/backlog\.json' internal/template/templates/
internal/template/templates/.moai/docs/todo-queue-storage.md:55        ← 정상(export)
internal/template/templates/.claude/skills/moai/SKILL.md:170
internal/template/templates/.claude/skills/moai/workflows/todo.md:17
internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md:95
$ grep -rn 'moai/todo/<project-key>' internal/template/templates/
internal/template/templates/.claude/skills/moai/workflows/todo.md:21   ← 위 목록에 없음
```
즉 템플릿의 `todo.md:21` 이 **교정되지 않은 채로도 AC-BJD-015 는 통과**한다. REQ-BJD-012 가 홈
폴백 형태를 명시적으로 금지하고 REQ-BJD-014 가 미러를 요구하는데, 그 교집합을 지키는 체크가 없다.
허용목록이 빠뜨린 것에서 조용히 깨지는 전형이다.
**Required fix**: AC-BJD-015 에 두 번째 grep 을 추가한다 —
`grep -rn 'moai/todo/<project-key>/backlog\.json' internal/template/templates/` 가 0건일 것.

**D6** — "four sites" vs "four mirrored files" 단위 혼동 — **Severity: minor · Class: optional**
사이트는 4개지만 **파일은 3개**다(`moai/SKILL.md`, `workflows/todo.md` ×2 사이트,
`moai-kanban-foreman/SKILL.md`). AC-BJD-016 은 "the **four mirrored files**"를 조회하라고 하는데
그런 파일 집합은 존재하지 않는다. 부수적으로 REQ-BJD-003 의 레이블도 부정확하다 — GEARS `Where`
는 capability gate / feature flag / static config 인데 "큐 레이아웃에 json 이 없음"은 런타임
디스크 상태다(문장 **형식**은 유효하므로 MP-2 에는 영향 없음; State-driven `While` 이 더 정확).
**Required fix**: AC-BJD-016 을 "the three mirrored files"로 정정. REQ-BJD-003 레이블은
`(State-driven)` 으로 바꾸거나 현 형식을 유지하되 레이블만 정정.

**D7** — §A.2 가 인용한 R5 근거가 주장을 실제로 뒷받침하지 않는다 — **Severity: minor · Class: optional**
§A.2 는 "All four sites are mirrored under `internal/template/templates/`
(`reader-surfaces.md` R5)"라고 쓰지만, R5 가 제시한 grep 4줄 중 하나는 **정상인**
`todo-queue-storage.md:55`(export 줄)이고 4번째 리더 사이트(`todo.md:21`)는 그 목록에 없다.
R5 는 "네 곳 모두 템플릿에도 있다"고 결론하면서 서로 다른 4줄을 셌다.
**주장 자체는 참이다** — 이 감사에서 별도 grep 으로 `todo.md:21` 의 템플릿 사본을 직접 확인했다.
결함은 결론이 아니라 귀속이다: SPEC 이 근거로 세우지 못하는 인용을 물려받았다.
**Required fix**: §A.2 의 인용을 이 보고서의 두 번째 grep(D5 참조)으로 교체하거나, R5 의 계수를
"3 files / 4 sites"로 정정.

**D8** — AC-BJD-002 의 "exactly one disclosure line"이 기존 stderr 줄과 경합할 수 있다 — **Severity: minor · Class: optional**
`internal/cli/todo_history.go:99-107` 은 REQ-TAQ-013 에 따라 이미 stderr 로 저장소 정체를 밝힌다
(LegacyJSON 1줄, SQLite 이면서 archive 테이블 부재 시 1줄). 이 SPEC 의 공시가 `moai todo history`
에 실리고 fixture db 에 archive 테이블이 없으면 stderr 에 **2줄**이 나오고, AC-BJD-002 의
"exactly one"이 무엇을 세는지 미정이 된다. archive 테이블이 있는 정상 fixture 에서는 발생하지
않으므로 잠재적 모호성이다.
**Required fix**: AC-BJD-002 에 "the disclosure line introduced by this SPEC" 로 계수 대상을
한정하거나, fixture 가 archive 테이블을 갖춘 레이아웃임을 Given 에 명시.

---

## 부수 확인 — 인용 좌표 정합성 (결함 아님, 기록용)

소스 인용을 트리 `ea97eec22` 에 대해 전수 대조했다. 1건 드리프트:

| SPEC 인용 | 실측 | 판정 |
|---|---|---|
| `backlog_archive_vouch.go:46` `InspectBacklogArchiveVouch` | 함수 정의는 **:49** (:46 은 doc 주석 시작) | 경미한 드리프트 |
| `backlog_archive_vouch.go:41` `BacklogArchiveVouch` struct | :41 | 정합 |
| `plan.md` "`layout.jsonExists` value it already has at `:57`" | `layout :=` 는 :50, SQLite 반환은 **:57** | 정합(반환 지점 기준) |
| `backlog_migrate.go:392` `inspectBacklogLayout` | :392 | 정합 |
| `backlog_migrate.go:556-566` `quarantineLegacyBacklog` | 함수 :558, 범위 안 | 정합 |
| `backlog_store.go:594-604` State D | :594 `case layout.dbExists && layout.jsonExists` | 정합 |
| `internal/cli/todo_export.go:74` `target := store.Path()` | :74 | 정합 |
| `internal/cli/todo_history.go:84,99` | :84 호출, :99 Fprintf | 정합 |
| `.claude/skills/moai/SKILL.md:170` · `workflows/todo.md:17,21` · foreman `:95` | 전부 정합 | 정합 |

8/9 정합. `:46`→`:49` 는 doc 주석 시작 줄을 인용한 것으로 보이며 오도하지 않는다.

---

## Regression Check

Iteration 1 — 해당 없음.

---

## Recommendation

**PASS-WITH-DEBT.** Must-Pass 7개 전부 통과, 집계 0.85 로 Tier M 임계 0.80 상회.
FAIL 요건(must-pass 실패 또는 임계 미달) 어느 쪽도 성립하지 않는다.

이 SPEC 의 강점은 진짜다. 카드의 원 인과를 반증 근거로 **거부**하고 측정된 피해를 따라간 점,
배제를 근거와 함께 서술한 점, AC-BJD-008 에 falsifiability 조건을 명시적으로 붙여 공허한 초록을
스스로 금지한 점, 그리고 미특정 작성자라는 미해결 Gap 을 결함이 아니라 **설계 근거로 전환**한 점은
이 저장소가 반복해 겪은 실패 형태에 대한 정확한 대응이다. Traceability 는 계수로 완전하다.

부채는 5건이 blocking 급이며 **전부 문구 수정으로 닫힌다 — 재설계는 필요 없다.** 킥오프 전
처리를 권하는 순서:

1. **D5** 먼저. BLOCKING 기준이 지켜야 할 4개 사이트 중 하나를 구조적으로 못 본다. grep 한 줄 추가.
2. **D2**. 리드가 지목한 재사용 제약의 유일한 체크가 실행 명령을 갖게 한다.
3. **D4**. Given 성립 기법과 그 확인을 넣어, 창 넓히기 말고 **Given 미성립** 쪽 공허 경로를 막는다.
4. **D3**. REQ-BJD-002 의 범위를 확정(읽기 표면 권장)하고 AC-BJD-002 를 일치시킨다.
5. **D1**. 재바인딩 문구를 디렉터리 기준으로 바꿔 AC-BJD-010 의 수리안과 양립시킨다.

D6/D7/D8 은 optional — 오케스트레이터 재량이며, 이것들 때문에 반복을 돌릴 필요는 없다.

Tier M ceiling 2 이므로 iteration 2 가 한 번 남아 있다. 위 5건이 문구 수정으로 닫히면
재감사는 열거된 defect delta 로 범위가 좁혀진다.
