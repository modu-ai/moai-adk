# SPEC-AUTONOMY-BATCH-GATE-001 — Compact (run-phase load)

> `spec.md` 전문 대신 로딩하는 압축본이다(0.3.0에서 spec.md로부터 다시 만들었다). 불일치 시 `spec.md`가 정본이다. 열린 결정은 `decision-index.md`(Q1–Q12)이며, Q4·Q5는 운영자가 범위 안으로 판정했다(2026-10-02). 열린 행은 Q1–Q3, Q6–Q10, Q12이고 Q8(Go 강제)은 범위 밖이다. Q11은 오케스트레이터 판정(Tier L).

card t1344 · version 0.3.0 · `tier: L`(오케스트레이터 판정 R1, 계수 규칙은 `spec.md` §A.5: 경로 단위, 라이브·미러 각각, `.moai/specs/`·`.moai/reports/` 제외 → 17개) · status draft · 제품 Go 변경은 리더 공지 문자열(4파일)에 한정, 테스트 Go 3파일 · 문서(규칙·스킬) 변경 + 가드 테스트 + 리더 공지 문장. 산출물: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`.

## Requirements (GEARS)

C.1 Batch unit and membership
- REQ-BGS-001 (Ubiquitous) — The session that holds the operator dialogue (the leader session in Kanban or Factory Mode, otherwise the orchestrating session) shall consolidate pending operator-form decisions into one batch gate summary only when the rows share the same gate row of the auto-semantics gate inventory; a lane session shall present only its own card's gate and shall not form a cross-card batch.
- REQ-BGS-002 (State-driven) — While a card's operator-form gate is ready for decision, the session shall not hold that card back to wait for further rows; a row that becomes ready after a summary has been presented shall join the next summary or be asked individually.
- REQ-BGS-003 (Event-driven) — When one judgment is the common decision subject of two or more pending rows that are not reserved under REQ-BGS-010, the session shall ask that judgment once, naming every affected card in the question and in the report that precedes it.

C.2 Summary format and the single approval
- REQ-BGS-004 (Ubiquitous) — The batch gate summary shall be rendered as a findings report in the response body before the decision question, with one row per card carrying the card id, the gate row, the SPEC id, the audit verdict reference (the final-iteration verdict of the card's current plan artifacts, with its iteration identifier, its score, and its margin to the tier's PASS threshold), the plan-artifact-hash-unchanged check, the reference to the card's own decision record, the result of the keep-set and leader-held-power check together with the basis on which the row was classified, and the counter-evidence field of REQ-BGS-007; rows carrying counter-evidence shall be listed before rows without it.
- REQ-BGS-005 (Ubiquitous) — The report shall be followed by exactly one decision question whose approval covers the listed approvable rows only, and the report shall state that rows added later, reserved rows, and blocked rows are not covered by it.
- REQ-BGS-006 (Ubiquitous) — The decision question shall offer, for any number of listed rows, a way to pull any single row out of the approval for individual handling, through the question channel's automatic free-text entry in which the operator names the card ids to pull out, while the explicit options stay within the channel's per-question option limit; pulling a row out shall leave the approval of the remaining rows intact.

C.3 Counter-evidence and per-row records
- REQ-BGS-007 (Ubiquitous) — Each row shall carry a `counter_refs=` field naming the strongest evidence against proceeding, drawn from audit warnings or recorded debt, the margin to the PASS threshold, unresolved decision-index rows, divergent audit-cross opinions, open blockers or wait records, and path overlap with another row of the same summary.
- REQ-BGS-008 (Event-detected) — When no counter-evidence is found for a row, the row shall read `counter_refs=none` followed by `searched=` naming the search that found none, and a bare `none` shall not count as a counter-evidence statement.
- REQ-BGS-009 (Event-driven) — When a row is approved through the summary, the session shall write that card's own decision record in the auto-semantics decision-record form, additively carrying `counter_refs=` and, in its `ladder_path`, the gate-row slug followed by `;batch=<id>` where `<id>` is the UTC time of the decision question in the form `YYYYMMDDTHHMMSSZ`, identical on every record of one summary and distinct between summaries, and a row without its own record shall not be treated as approved.

C.4 Reserved and blocked rows
- REQ-BGS-010 (Event-driven) — When a pending row falls in a keep-set category of the auto-semantics gate inventory (environment-impossible, operator-held, or an irreversible operation on an external shared system) or in a power that the kanban-dispatch rule keeps with the leader session (final PASS/FAIL verdicts, final merge approval, operator gates, card issuance and `done` through queue mutations, CodeRabbit slot-wait adjudication, cross-session dispute coordination), the session shall handle that row individually, outside the single approval.
- REQ-BGS-011 (Ubiquitous) — The session shall list a row as approvable only when its most recent audit verdict (the final-iteration verdict of the card's current plan artifacts) is PASS, the plan phase records audit-ready status, the plan-artifact hashes are unchanged since that verdict, and no blocker is open; the session shall show every other row as blocked and exclude it from the single approval, treating PASS-WITH-DEBT, BYPASSED, FAIL, INCONCLUSIVE, and an absent verdict as blocked states, so that the summary never weakens the authority-gate invariant of the decision ladder.
- REQ-BGS-012 (State-driven) — While `workflow.autonomy.mode` is `contract`, the summary shall not stand in for the signing gate: the contract signature verified by `moai contract kickoff-check` remains the plan→run gate and forms no summary row.

C.5 Home, consistency, and baseline
- REQ-BGS-013 (Ubiquitous) — The batch gate summary doctrine shall have one home, a new subsection of the gate-inventory section of the auto-semantics rule, mirrored byte-identically in the template tree; every other surface that mentions it shall carry a pointer to that home and shall not restate the doctrine; and no always-loaded file shall grow beyond the per-edit threshold of the always-loaded cost discipline.
- REQ-BGS-014 (Ubiquitous) — The Kickoff wording of the plan-assembly skill and of the workflow router shall describe the plan→run Kickoff in its default-autonomous form, citing the default-autonomous transition and the batch gate summary by path and section, shall not state that a plan-auditor PASS or a skip-eligible score never substitutes for the operator question on a card outside the keep-set, and shall keep the strings the Kickoff preservation guards pin; this requirement covers those two files only.
- REQ-BGS-015 (Ubiquitous) — The run phase shall land a measured baseline of question-tool gate rounds, recorded in a file at a tracked path inside this SPEC's directory, in its own commit that is an ancestor of the commit that adds the doctrine, naming the measuring command, the observed output, the classification method, and their limits.

C.6 Launcher leader notice
- REQ-BGS-016 (Ubiquitous) — The SessionStart notice of the factory leader and of the kanban leader shall each carry one sentence that names the batch gate summary and points at its doctrine home in the auto-semantics rule by path and section, shall not restate the doctrine, shall not name the question tool, and shall carry no card id, SPEC id, or date.
- REQ-BGS-017 (Ubiquitous) — That sentence shall be present in the English agent-facing copy of each leader notice and in the operator-facing copy of each leader notice in each of the en, ko, ja, and zh locales, and shall be absent from every lane-session and companion-session notice.
- REQ-BGS-018 (Ubiquitous) — Adding that sentence shall leave every previously pinned notice string, block layout, and role-term constraint of the leader, lane, and companion notices intact.

C.7 Approval-to-record consistency
- REQ-BGS-019 (Event-driven) — When the operator's answer to the decision question is received, the session shall re-read, for each row the answer approves, the row's most recent audit verdict and plan-artifact hashes before writing that row's decision record, and shall refuse — not record as approved — a row that no longer satisfies REQ-BGS-011, reporting the refusal to the operator.

열린 행에 대한 초안의 읽기: REQ-001 Q1·Q10, REQ-003 Q1·Q2, REQ-004 Q10, REQ-005 Q3, REQ-007 Q7, REQ-009·019 Q8, REQ-010 Q6·Q2, REQ-014 Q10·Q12, REQ-016 Q10, REQ-015 Q9 (`spec.md` §B.8 표).

## Acceptance criteria (Given/When/Then 요지; 전문·RED-now 칸은 `acceptance.md`)

- AC-001 (rb) 정본 절 제목 `### 9.2 The batch gate summary`가 라이브·미러에 각 1회.
- AC-002 (rb) 서식·keep-set 점검 필드·질문 하나·승인 범위·빼내기(자동 자유 입력에 카드 id)가 정본 절에 있다 — 가드 테스트(A05–A13).
- AC-003 (rb) `counter_refs=`·`none searched=`·행별 §10 기록·`;batch=<id>`·기록 직전 재확인(A14–A22).
- AC-004 (rb) keep-set 세 범주·리더 보유 권한 여섯 항목·contract 모드 제외, "user-facing behavior" 부재(A23–A27).
- AC-005 (rb) 최종 반복 판정 결속, PASS 양성 규칙, PASS-WITH-DEBT·BYPASSED·FAIL·INCONCLUSIVE·부재 각각 차단, audit-ready 미기록·해시 변경·열린 차단(A28–A37).
- AC-006 (rb) 같은 게이트 행·붙잡지 않기·동일 판정 1회(A01–A04).
- AC-007 (rb) 앵커 표 38행마다 변이 본문이 해당 앵커 위반으로 거부(하위 테스트 `--- PASS` 39개 이상).
- AC-008 (rb) `.moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/baseline-gate-rounds.md`가 존재·무시되지 않음·정본 절 커밋 D의 조상인 별도 커밋 B이며 B는 그 파일 하나만 담는다.
- AC-009 (rb) `git diff --name-only develop...HEAD -- '*.go'` → 허용 목록 정확히 일곱 줄(제품 Go 넷 `internal/hook/session_start_{kanban,factory}{,_i18n}.go` + 테스트 Go 셋). 제품 Go 변경이 알림 문장에 한정되는지는 `--numstat` 공시를 감사가 읽는다.
- AC-010 (rg) 예산·미러·LOC·보존 구간·Kickoff 보존·중립성 가드 GREEN, 리더 공지 고정 테스트 63개 GREEN, `run.md` ≤199줄.
- AC-011~013 (rg, 감사 읽기) 혼합 배치(PASS-WITH-DEBT 포함) / 빼내기 / 붙잡지 않기·동일 판정 1회 시나리오.
- AC-014 (rb, 무조건) 낡은 구절(`stays MANDATORY`, `does NOT substitute for the gate`, `never bypasses it`, `score-independent`)이 두 사본에서 0, `§9.1`·`§9.2` 인용 1 이상, 줄 수(597/284/282)·`moai.md` 사전 차이 헌크 넷 불변, `TestSpecAssembly_*` GREEN. 두 파일만 덮는다.
- AC-015 (rb) 두 리더 공지 × 네 로케일에 포인터 토큰 `.claude/rules/moai/workflow/auto-semantics.md` §9.2가 있고, 금지 토큰이 없고, 레인·동반 공지에는 없고, 알림 소스 4파일의 `AskUserQuestion` 토큰 수가 0이며, 변이 본문이 거부된다(`TestLeaderNoticeBatchGatePointer`, `--- PASS` 12개 이상).
- AC-016 (rb) 정본 전용 어휘 `counter_refs=`·`searched=`가 정본 두 사본에만 있음, 포인터 네 곳에 `§9.2`, `kanban-dispatch.md` 라이브·미러 바이트 순증가 <1,000(기준 SHA `c50da9c2f`), `run.md` 추가=삭제, `kanban-dispatch.md` 추가 줄 ≤3.
- AC-017 (rg, 감사 읽기) 응답과 기록 사이에 표류한 행은 기록 직전 재확인으로 거부된다.

## Files to modify

- `.claude/rules/moai/workflow/auto-semantics.md` + `internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md` — §9.2 신설, §9.1 끝·§10 한 줄
- `.claude/rules/moai/workflow/kanban-dispatch.md` + 미러 — Boundaries 포인터(보존 구간 `:29-33` 밖, `:177` 사전 차이 유지; 실행은 마지막)
- `.claude/skills/moai/workflows/run.md` + 미러 — `:137` 줄 수 불변 포인터
- `.claude/skills/moai/workflows/plan/spec-assembly.md:202-208` + 미러 — 낡은 표현 정합(597줄 불변)
- `.claude/skills/moai/workflows/moai.md:144,240` + 미러 — 낡은 표현 정합(284/282줄 불변, 사전 차이 줄 `:215,245,253,283-284` 미접촉, 사본마다 따로 편집)
- `internal/hook/session_start_kanban_i18n.go`, `session_start_kanban.go`, `session_start_factory_i18n.go`, `session_start_factory.go` — 리더 공지 문장 필드와 네 로케일 값, 기존 블록에 한 줄 합치기
- `internal/hook/session_start_kanban_i18n_test.go` — 필드 빈값 표에 새 필드 덧붙임
- `internal/hook/session_start_leader_gate_notice_test.go` — 신규(문장 규칙·변이·소스 스캔)
- `internal/template/batch_gate_summary_doctrine_test.go` — 신규(정본 절 가드, 테스트 전용)
- (저장소 파일 17개 — `spec.md` §A.5의 계수. 아래는 계수 밖) `.moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/baseline-gate-rounds.md` — 신규(M0 단독 커밋, 추적되는 SPEC 디렉터리 안)

## Exclusions (What NOT to Build)

- 리더 공지 문장(REQ-BGS-016·017·018)과 그 테스트 기대값을 뺀 제품 Go 변경 전부, `moai factory decide` 확장·결정 기록 자동 기록·MCP `factory_decide` 다중 카드화(Q8, 열림), 게이트 라운드 생산용 측정 도구.
- 레인·동반·스테일 런 공지, 레인 규칙 필드, `laneSpawnAuthority`, SessionStart 방출 경로와 `startup` 원천 게이트, `internal/hook/CLAUDE.md:13` 문구 정정.
- 큐 진입·카드 선택, `/moai:todo --auto`의 "batch approval", `kanban-dispatch.md`의 "Promotion is the operator's act, always." → "The self-dispatch lane exception." 보존 구간.
- keep-set 게이트의 기계 동작(push·병합·abandon·hold 해제)과 단일 승인 편입.
- Jev 능력·3등급 독트린 정의(로컬 문서), `workflow.jev.enabled`, `jev_ask`.
- Q4가 이름 붙인 두 파일 밖의 Kickoff 문구(영향 4파일·표지만 5파일, Q12).
- "게이트 왕복 176→10~20회" 결과 수치의 단언.
- `orchestration-mode-selection.md`(`:18` Frozen) 편집, `SKILL.md`, 사전 존재 미러 차이 줄(`kanban-dispatch.md:177`, `moai.md:215,245,253,283-284`) 수정.
