# SPEC-AUTONOMY-BATCH-GATE-001 — Compact (run-phase load)

> `spec.md` 전문 대신 로딩하는 압축본이다(0.2.0에서 spec.md로부터 다시 만들었다). 불일치 시 `spec.md`가 정본이다. 열린 결정은 `decision-index.md`(Q1–Q11)이며, Q4·Q5는 운영자가 범위 안으로 판정했다(2026-10-02). 열린 행은 Q1–Q3, Q6–Q10이고 Q8(Go 강제)은 범위 밖이다.

card t1344 · version 0.2.0 · `tier: M`(선언 그대로 — 파일 수 재측정은 `spec.md` §A.5: 미러를 따로 세면 18개, 쌍을 한 파일로 세면 13개, 판정은 오케스트레이터) · status draft · 제품 Go 변경은 리더 공지 문자열(4파일)에 한정, 테스트 Go 3파일 · 문서(규칙·스킬) 변경 + 가드 테스트 + 리더 공지 문장.

## Requirements (GEARS)

C.1 Batch unit and membership
- REQ-BGS-001 (Ubiquitous) — The session that holds the operator dialogue (the leader session in Kanban or Factory Mode, otherwise the orchestrating session) shall consolidate pending operator-form decisions into one batch gate summary only when the rows share the same gate row of the auto-semantics gate inventory; a lane session shall present only its own card's gate and shall not form a cross-card batch.
- REQ-BGS-002 (State-driven) — While a card's operator-form gate is ready for decision, the session shall not hold that card back to wait for further rows; a row that becomes ready after a summary has been presented shall join the next summary or be asked individually.
- REQ-BGS-003 (Event-driven) — When one judgment is the common decision subject of two or more pending rows that are not reserved under REQ-BGS-010, the session shall ask that judgment once, naming every affected card in the question and in the report that precedes it.

C.2 Summary format and the single approval
- REQ-BGS-004 (Ubiquitous) — The batch gate summary shall be rendered as a findings report in the response body before the decision question, with one row per card carrying the card id, the gate row, the SPEC id, the audit verdict reference with its score and its margin to the tier's PASS threshold, the plan-artifact-hash-unchanged check, the reference to the card's own decision record, and the counter-evidence field of REQ-BGS-007; rows carrying counter-evidence shall be listed before rows without it.
- REQ-BGS-005 (Ubiquitous) — The report shall be followed by exactly one decision question whose approval covers the listed approvable rows only, and the report shall state that rows added later and reserved or blocked rows are not covered by it.
- REQ-BGS-006 (Ubiquitous) — The decision question shall always offer a way to pull any single row out of the approval for individual handling, and pulling a row out shall leave the approval of the remaining rows intact.

C.3 Counter-evidence and per-row records
- REQ-BGS-007 (Ubiquitous) — Each row shall carry a `counter_refs=` field naming the strongest evidence against proceeding, drawn from audit warnings or recorded debt, the margin to the PASS threshold, unresolved decision-index rows, divergent audit-cross opinions, open blockers or wait records, and path overlap with another row of the same summary.
- REQ-BGS-008 (Event-detected) — When no counter-evidence is found for a row, the row shall read `counter_refs=none` followed by `searched=` naming the search that found none, and a bare `none` shall not count as a counter-evidence statement.
- REQ-BGS-009 (Event-driven) — When a row is approved through the summary, the session shall write that card's own decision record in the auto-semantics decision-record form, additively carrying `counter_refs=` and naming the batch in its ladder path, and a row without its own record shall not be treated as approved.

C.4 Reserved and blocked rows
- REQ-BGS-010 (Unwanted) — The single approval shall not cover a row in a keep-set category (environment-impossible, operator-held, or an irreversible operation on an external shared system) or a row whose decision the authority gates reserve to the operator or the leader (final verdicts, merge approval, queue mutations, operator gates, user-facing behavior changes, CodeRabbit slot-wait judgment); such rows shall keep their individual approval.
- REQ-BGS-011 (Event-driven) — When a row's audit verdict is FAIL, INCONCLUSIVE, or absent, its plan-artifact hashes changed since the verdict, or a blocker is open, the session shall show the row as blocked and exclude it from the approval, so that the summary never weakens the authority-gate invariant of the decision ladder.
- REQ-BGS-012 (State-driven) — While `workflow.autonomy.mode` is `contract`, the summary shall not stand in for the signing gate: the contract signature verified by `moai contract kickoff-check` remains the plan→run gate and forms no summary row.

C.5 Home, consistency, and baseline
- REQ-BGS-013 (Ubiquitous) — The doctrine shall have one home, a new subsection of the gate-inventory section of `auto-semantics.md`, byte-identical in the template tree; every other surface shall carry a pointer only, an always-loaded file shall not grow by more than 1,000 bytes, and the run workflow skill shall gain no line.
- REQ-BGS-014 (Ubiquitous) — The Kickoff wording of the plan-assembly skill and of the workflow router shall describe the plan→run Kickoff in its default-autonomous form, citing the default-autonomous transition and the batch gate summary by path and section, shall not state that a plan-auditor PASS or a skip-eligible score never substitutes for the operator question on a card outside the keep-set, and shall keep the strings the Kickoff preservation guards pin.
- REQ-BGS-015 (Ubiquitous) — The run phase shall land a measured baseline of question-tool gate rounds in its own commit, ahead of the doctrine commit, naming the measuring command, the observed output, the classification method, and their limits.

C.6 Launcher leader notice
- REQ-BGS-016 (Ubiquitous) — The SessionStart notice of a leader session — the factory leader and the kanban leader, each in its agent-facing and its operator-facing copy and in each of the en, ko, ja, and zh locales — shall carry one sentence that names the batch gate summary and points at its doctrine home in `auto-semantics.md` by path and section, shall not restate the doctrine, shall not name the question tool, shall carry no card id, SPEC id, or date, shall not appear in a lane or companion notice, and shall leave every previously pinned notice string intact.

## Acceptance criteria (Given/When/Then 요지; 전문·RED-now 칸은 `acceptance.md`)

- AC-001 (rb) 정본 절 제목 `### 9.2 The batch gate summary`가 라이브·미러에 각 1회.
- AC-002 (rb) 서식·질문 하나·승인 범위·pull-out이 정본 절에 있다 — 가드 테스트.
- AC-003 (rb) `counter_refs=`·`none searched=`·행별 §10 기록 — 가드 테스트.
- AC-004 (rb) keep-set·권한 예약·contract 모드 제외 — 가드 테스트.
- AC-005 (rb) 차단 조건 다섯(FAIL·INCONCLUSIVE·부재·해시 변경·열린 차단) — 가드 테스트.
- AC-006 (rb) 같은 게이트 행·붙잡지 않기·동일 판정 1회 — 가드 테스트.
- AC-007 (rb) 앵커를 뺀 변이 본문마다 해당 불변식 위반(하위 테스트 PASS 10개 이상).
- AC-008 (rb) `.moai/reports/t1344/baseline-gate-rounds.md`가 정본 절 커밋의 조상인 별도 커밋(`git merge-base --is-ancestor` exit 0).
- AC-009 (rb) `git diff --name-only develop...HEAD -- '*.go'` → 허용 목록 정확히 일곱 줄(제품 Go 넷 `internal/hook/session_start_{kanban,factory}{,_i18n}.go` + 테스트 Go 셋). 제품 Go 변경이 알림 문장에 한정되는지는 `--numstat` 공시를 감사가 읽는다.
- AC-010 (rg) 예산·미러·LOC·보존 구간·Kickoff 보존·중립성 가드 GREEN, 리더 공지 고정 테스트 63개 GREEN, `kanban-dispatch.md` 순증가 <1,000 B, `run.md` ≤199줄.
- AC-011~013 (rg, 감사 읽기) 혼합 배치 / pull-out / 붙잡지 않기·동일 판정 1회 시나리오.
- AC-014 (rb, 무조건) 낡은 구절(`stays MANDATORY`, `does NOT substitute for the gate`, `never bypasses it`, `score-independent`)이 두 사본에서 0, `§9.1`·`§9.2` 인용 1 이상, 줄 수(597/284/282)·`moai.md` 사전 차이 헌크 넷 불변, `TestSpecAssembly_*` GREEN.
- AC-015 (rb) 두 리더 공지 × 네 로케일에 포인터 토큰 `.claude/rules/moai/workflow/auto-semantics.md` §9.2가 있고, 금지 토큰(질문 도구 이름·`SPEC-`·카드 id·날짜·`moai todo`·`lead`·`리드`·`epic`)이 없고, 레인·동반 공지에는 없고, 알림 소스 4파일의 `AskUserQuestion` 토큰 수가 0이며, 변이 본문이 거부된다(`TestLeaderNoticeBatchGatePointer`, `--- PASS` 12개 이상).

## Files to modify

- `.claude/rules/moai/workflow/auto-semantics.md` + `internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md` — §9.2 신설, §9.1 끝·§10 한 줄
- `.claude/rules/moai/workflow/kanban-dispatch.md` + 미러 — Boundaries 포인터(보존 구간 `:29-33` 밖, `:177` 사전 차이 유지)
- `.claude/skills/moai/workflows/run.md` + 미러 — `:137` 줄 수 불변 포인터
- `.claude/skills/moai/workflows/plan/spec-assembly.md:202-208` + 미러 — 낡은 표현 정합(597줄 불변)
- `.claude/skills/moai/workflows/moai.md:144,240` + 미러 — 낡은 표현 정합(284/282줄 불변, 사전 차이 줄 `:215,245,253,283-284` 미접촉, 사본마다 따로 편집)
- `internal/hook/session_start_kanban_i18n.go`, `session_start_kanban.go`, `session_start_factory_i18n.go`, `session_start_factory.go` — 리더 공지 문장 필드와 네 로케일 값, 기존 블록에 한 줄 합치기
- `internal/hook/session_start_kanban_i18n_test.go` — 필드 빈값 표에 새 필드 덧붙임
- `internal/hook/session_start_leader_gate_notice_test.go` — 신규(문장 규칙·변이·소스 스캔)
- `internal/template/batch_gate_summary_doctrine_test.go` — 신규(정본 절 가드, 테스트 전용)
- `.moai/reports/t1344/baseline-gate-rounds.md` — 신규(M0 단독 커밋)

## Exclusions (What NOT to Build)

- 리더 공지 문장(REQ-BGS-016)과 그 테스트 기대값을 뺀 제품 Go 변경 전부, `moai factory decide` 확장·결정 기록 자동 기록·MCP `factory_decide` 다중 카드화(Q8, 열림), 게이트 라운드 생산용 측정 도구.
- 레인·동반·스테일 런 공지, 레인 규칙 필드, `laneSpawnAuthority`, SessionStart 방출 경로와 `startup` 원천 게이트, `internal/hook/CLAUDE.md:13` 문구 정정.
- 큐 진입·카드 선택, `/moai:todo --auto`의 "batch approval", `kanban-dispatch.md`의 "Promotion is the operator's act, always." → "The self-dispatch lane exception." 보존 구간.
- keep-set 게이트의 기계 동작(push·병합·abandon·hold 해제)과 단일 승인 편입.
- Jev 능력·3등급 독트린 정의(로컬 문서), `workflow.jev.enabled`, `jev_ask`.
- "게이트 왕복 176→10~20회" 결과 수치의 단언.
- `orchestration-mode-selection.md`(`:18` Frozen) 편집, `SKILL.md`, 사전 존재 미러 차이 줄(`kanban-dispatch.md:177`, `moai.md:215,245,253,283-284`) 수정.
