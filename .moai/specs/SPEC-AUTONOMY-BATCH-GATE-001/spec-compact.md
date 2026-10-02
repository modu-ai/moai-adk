# SPEC-AUTONOMY-BATCH-GATE-001 — Compact (run-phase load)

> 요구사항마다 한 줄 요지만 싣는 압축본이다(0.4.0에서 spec.md를 베끼지 않고 요지로 다시 썼다). 정확한 GEARS 문장과 REQ-BGS-010의 목록(keep-set 세 범주, 리더 보유 권한 여섯 항목)은 `spec.md` §C가 정본이다. 불일치 시 `spec.md`가 이긴다. 열린 결정은 `decision-index.md`(Q1–Q13)이며 Q4·Q5는 운영자가 범위 안으로 판정했다(2026-10-02). 열린 행은 Q1–Q3, Q6–Q10, Q12, Q13이고 Q8(Go 강제)은 범위 밖이다. Q11은 오케스트레이터 판정(Tier L).

card t1344 · version 0.4.1 · `tier: L`(계수 규칙은 `spec.md` §A.5: 경로 단위, 라이브·미러 각각, `.moai/specs/`·`.moai/reports/` 제외 → 17개) · status draft · 제품 Go 변경은 리더 공지 문자열(4파일)에 한정, 테스트 Go 3파일 · 문서(규칙·스킬) 변경 + 가드 테스트 + 리더 공지 문장. 산출물: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`.

## Requirements (요지, 20개)

C.1 Batch unit and membership
- REQ-BGS-001 (Ubiquitous) — 운영자 대화를 쥔 세션이 plan→run Kickoff 행의 운영자 형태 결정만 배치 게이트 요약 하나로 묶는다. 나머지 게이트 행(factory decide 행들, sync 차단 승인, 카드 선택)은 요약 행이 아니고 운영자 답이 필요하면 개별 질문이다. 레인은 교차 카드 배치를 만들지 않는다.
- REQ-BGS-002 (State-driven) — 준비된 카드를 더 많은 행을 기다리느라 붙잡지 않는다. 늦게 준비된 행은 다음 요약이나 개별 질문이다.
- REQ-BGS-003 (Event-driven) — 한 판정이 예약도 차단도 아닌 Kickoff 행 둘 이상의 공통 결정 주체이면 한 번 묻고 그 카드를 전부 적는다.

C.2 Summary format and the single approval
- REQ-BGS-004 (Ubiquitous) — 보고서가 질문보다 앞서고 머리가 Kickoff 행을 적는다. 나열된 카드마다 한 행: 카드 id, SPEC id, 독립 plan-audit 판정(최종 반복·식별자·점수·여유), 해시 불변 확인, 기록 참조, keep-set·리더 권한 점검과 근거, `counter_refs=`. 반대 증거 행이 먼저.
- REQ-BGS-005 (Ubiquitous) — 질문은 정확히 하나(한 호출에 요약 질문 하나). 승인은 나열된 승인 가능 행에만 미치고, 나중에 온 행·예약 행·차단 행은 덮이지 않는다고 보고서가 진술한다.
- REQ-BGS-006 (Ubiquitous) — 빼내기는 자동 자유 입력: 적힌 카드 id는 빼내고 나열된 나머지는 승인한다. 질문 문구가 이 규칙을 적는다. 읽을 수 없는 입력이나 표에 없는 id는 승인 없이 재질문. 명시적 선택지는 상한 안.

C.3 Counter-evidence and per-row records
- REQ-BGS-007 (Ubiquitous) — 행마다 `counter_refs=`(닫힌 원천 목록 여섯).
- REQ-BGS-008 (Event-detected) — 반대 증거가 없으면 보고서와 기록 양쪽에 `counter_refs=none searched=<공백 없는 토큰>`. `searched=` 없는 `none`은 진술이 아니다.
- REQ-BGS-009 (Event-driven) — 승인된 행마다 한 줄 자기 결정 기록: §10 세 필드, 이어서 `counter_refs=`. `ladder_path` = 게이트 행 슬러그;`batch=<UTC YYYYMMDDTHHMMSSZ>`, 한 요약 안 동일, 결정 보드에 같은 값이 있으면 다음 빈 초. 기록 없는 행은 미승인.

C.4 Reserved and blocked rows
- REQ-BGS-010 (Event-driven) — keep-set 범주 또는 리더 보유 권한(목록은 `spec.md`)의 행은 단일 승인 밖에서 개별 처리한다. 목록의 "operator gates" 항목은 요약 모집단인 운영자 형태 Kickoff 행을 포함하지 않는다: 그 행은 keep-set 범주에 걸릴 때만 예약되고 아니면 요약 행으로 REQ-BGS-011이 분류한다.
- REQ-BGS-011 (Ubiquitous) — 승인 가능은 독립 plan-audit 최종 반복 판정 PASS + audit-ready 기록 + 해시 불변 + 열린 차단 없음 *넷 모두*일 때뿐. 그 밖의 행은 차단으로 보고되고 승인에서 빠진다(PASS-WITH-DEBT·BYPASSED·FAIL·INCONCLUSIVE·부재·plan-auditor가 아닌 판정). 보고 위치(표 안/밖)는 Q3.
- REQ-BGS-012 (State-driven) — `contract` 모드에서는 계약 서명이 plan→run 게이트이고 요약 행은 없다.

C.5 Home, consistency, and baseline
- REQ-BGS-013 (Ubiquitous) — 정본은 `auto-semantics.md`의 새 절 하나(템플릿 미러 바이트 동일). 다른 표면은 포인터만. 상시 로딩 파일은 한도 안.
- REQ-BGS-014 (Ubiquitous) — `spec-assembly.md`와 `workflows/moai.md`의 Kickoff 문구를 기본 자율 형태에 맞추고 §9.1·§9.2를 인용한다. 보존 문자열 유지. 두 파일만.
- REQ-BGS-015 (Ubiquitous) — run-phase가 게이트 라운드 기준선을 SPEC 디렉터리 안 추적 경로에 별도 커밋(정본 절 커밋의 조상)으로 남긴다: 명령·관측 출력·분류 방식·한계.

C.6 Launcher leader notice
- REQ-BGS-016 (Ubiquitous) — 두 리더 공지가 이름 `batch gate summary`를 그대로 담고 정본을 경로·절로 가리키는 문장 하나를 싣는다. 정본을 되풀이하지 않고, 질문 도구 이름·카드 id·SPEC id·날짜가 없다.
- REQ-BGS-017 (Ubiquitous) — 그 문장은 두 리더 공지의 영어 에이전트용 사본과 en·ko·ja·zh 운영자용 사본에 있고 레인·동반 공지에는 없다.
- REQ-BGS-018 (Ubiquitous) — 기존 고정 문자열·블록 구조·역할 용어 제약은 그대로다.

C.7 Approval consistency
- REQ-BGS-019 (Event-driven) — 응답을 받으면 기록 직전에 승인 행마다 REQ-BGS-011 네 조건 전부와 REQ-BGS-010 분류(운영자 hold 포함)를 다시 읽고, 어긋난 행은 기록하지 않고 거부하며 운영자에게 알린다.
- REQ-BGS-020 (Ubiquitous) — 단일 승인은 Kickoff의 다른 조건을 약화하지 않는다: 승인된 카드마다 티어·모드 선호·PR 전략·체인 범위가 run 진입 전에 디스크에 있다(카드·SPEC 계약 또는 그 카드를 위한 운영자 대화).

열린 행에 대한 초안의 읽기: REQ-001 Q10·Q13, REQ-003 Q1·Q2, REQ-004 Q10, REQ-005 Q3, REQ-007 Q7, REQ-009·019 Q8, REQ-010 Q6·Q2, REQ-011 Q3, REQ-014 Q10·Q12, REQ-015 Q9, REQ-016 Q10 (`spec.md` §B.8 표).

## Acceptance criteria (Given/When/Then 요지; 전문·RED-now 칸은 `acceptance.md`)

- AC-001 (rb) 정본 절 제목 `### 9.2 The batch gate summary`가 라이브·미러에 각 1회.
- AC-002 (rb) 서식·keep-set 점검 필드·질문 하나·승인 범위·빼내기와 읽기 규칙·선호 배출 비약화가 정본 절에 있다 — 가드 테스트(A05–A13, A41, A42, A45).
- AC-003 (rb) `counter_refs=`·`none searched=<토큰>`·한 줄 §10 기록·`;batch=<id>`·기록 직전 재확인(A14–A22, A43).
- AC-004 (rb) keep-set 세 범주·리더 보유 권한 여섯 항목("operator gates"는 운영자 형태 Kickoff 행을 포함하지 않는다는 한정과 함께)·contract 모드 제외, "user-facing behavior" 부재(A23–A27).
- AC-005 (rb) 최종 반복 독립 판정 결속, PASS 양성 규칙, 차단 토큰 다섯, 차단 사유 셋, 자기 진술 PASS 거부(A28–A37, A44).
- AC-006 (rb) Kickoff 행 한정·다른 게이트 행은 개별 질문·붙잡지 않기·동일 판정 1회(차단·예약 행은 이름에 올리지 않음)(A01–A04, A39, A40).
- AC-007 (rb) 앵커 표 45행마다 변이 본문이 해당 앵커 위반으로 거부 — `--- PASS` 46 이상, `=== RUN`과 같은 수, `[no tests to run]` 없음, 변이 거부 줄 45개.
- AC-008 (rb) M0 종료 점검 V1–V5(파일 존재·무시되지 않음·네 요소 레이블·B 하나·B는 그 파일만)와 끝 점검 V6–V8(파일을 건드린 커밋은 B뿐·정본 절 커밋 D 하나·B가 D의 조상)을 갈라 판정한다.
- AC-009 (rb) `git diff --name-only develop...HEAD -- '*.go'` → 허용 목록 정확히 일곱 줄. 제품 Go 변경이 알림 문장에 한정되는지는 `--numstat` 공시를 감사가 읽는다.
- AC-010 (rg) 예산·미러·LOC·보존 구간·Kickoff 보존·중립성 가드 GREEN, 리더 공지 고정 테스트 63개 GREEN, `run.md` ≤199줄.
- AC-011~013 (rg, 감사 읽기) 혼합 배치(차단 행은 표 안이든 밖이든 차단으로 보고) / 빼내기와 읽기 규칙 / 붙잡지 않기·동일 판정 1회(차단 카드 제외).
- AC-014 (rb, 무조건) 낡은 구절 4종이 두 사본에서 0, `§9.1`·`§9.2` 인용 1 이상, 줄 수(597/284/282)·`moai.md` 사전 차이 헌크 넷 불변, `TestSpecAssembly_*` GREEN. 두 파일만.
- AC-015 (rb) 두 리더 공지 × 네 로케일에 포인터 토큰과 이름 토큰이 있고, 금지 토큰이 없고, 레인·동반 공지에는 없고, 알림 소스 4파일의 `AskUserQuestion` 토큰 수가 0이며, 변이 본문이 거부된다(`TestLeaderNoticeBatchGatePointer`: `--- PASS` 17 이상, `=== RUN`과 같은 수, 변이 거부 줄).
- AC-016 (rb) 정본 전용 어휘 `counter_refs=`·`searched=`가 정본 두 사본에만 있음, 포인터 네 곳에 `§9.2`, 읽는 시점의 `git merge-base develop HEAD` 기준 `kanban-dispatch.md` 바이트 순증가 <1,000·`run.md` 추가=삭제·`kanban-dispatch.md` 추가 ≤3, 범위 비공허 대조(병합 전에만 유효).
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
- Kickoff 외 게이트 행(sync 차단 승인, 카드 선택, factory decide 행들)의 요약 편입과 그 승인 가능성 술어(Q13).
- 레인·동반·스테일 런 공지, 레인 규칙 필드, `laneSpawnAuthority`, SessionStart 방출 경로와 `startup` 원천 게이트, `internal/hook/CLAUDE.md:13` 문구 정정.
- 큐 진입·카드 선택, `/moai:todo --auto`의 "batch approval", `kanban-dispatch.md`의 "Promotion is the operator's act, always." → "The self-dispatch lane exception." 보존 구간.
- keep-set 게이트의 기계 동작(push·병합·abandon·hold 해제)과 단일 승인 편입.
- Jev 능력·3등급 독트린 정의(로컬 문서), `workflow.jev.enabled`, `jev_ask`.
- Q4가 이름 붙인 두 파일 밖의 Kickoff 문구(영향 4파일·표지만 5파일, Q12).
- "게이트 왕복 176→10~20회" 결과 수치의 단언.
- `orchestration-mode-selection.md`(`:18` Frozen) 편집, `SKILL.md`, 사전 존재 미러 차이 줄(`kanban-dispatch.md:177`, `moai.md:215,245,253,283-284`) 수정.
