auditor-model: claude-opus-5-5

verdict: FAIL
audited_sha: 6ade6fe387d777a721a4b46d9ede062a86ec69cb

# SPEC Review Report: SPEC-ALWAYS-LOADED-BUDGET-001
Iteration: 3/3 (델타 + 회귀 확인, CN-4 전체 재독) — Tier L 상한 도달, 최종 반복
Verdict: FAIL — STOP (iter2 0.81 → iter3 0.80, 점수 하락) + 상한 도달 에스컬레이션
Overall Score: 0.80 (Tier L 문턱 0.85 미달)
Plan Artifact Hash: 미산출 — 이 세션은 Bash·Grep·Glob이 모두 거부·부재해 sha256 을 계산하지 못함 (Gaps)
Auditor Version: plan-auditor (card t1469, iter3)
Intended destination: <tree agent-a6d780d7ba07bc24c>/.moai/reports/t1469/plan-audit-iter3.md — 지시에 따라 스크래치에만 기록, 커밋하지 않음.

Reasoning context ignored per M1 Context Isolation. 운영자 판정(Q1 동결 해제·v3.2.0 제외)과 리더 판정(Q5 수용)은 재론하지 않음.

Cross-model: `audit_multi`(project_root=agent-a6d780d7ba07bc24c, baseBranch) — claude FAIL(required, 실행 도구 없음) · codex FAIL(required, 대상 트리에서 RED-now 전부 재실행, 증거 `/tmp/alb-iter3-evidence-6ade6fe38.md`) · glm inconclusive(advisory, 응답 본문 없음). overall `fail`, disagreement_flag `false`, `audit_receipt` 미발급. 두 required 백엔드가 독립적으로 같은 P1 두 건(N3 잔존, AC-020(c) 근거 부재)을 같은 줄에서 짚었고, 이 감사가 원문 대조로 확인했다.

## 1. Claim

iter2 blocking 6건 가운데 N2·N4·N5·N6 은 실질적으로 닫혔다. N1 은 RED 신호 면에서는 닫혔지만, 표에 적힌 그대로의 명령 두 개(AC-008·AC-015, 약어 `T`)가 실행되지 않고(exit 2), AC-008 의 stdout 이 축자가 아니다 → MP-8 판정 규칙상 FAIL(수리는 기계적). N3 은 구속 절 안에서만 닫혔다 — 토큰이 전혀 없는 절·파일(`goal-directive.md` 전체)의 규범 문단은 여전히 원장 밖이고, plan 이 그 파일을 분할 대상으로 예정한다. 새 blocking 2건: AC-ALB-020(c) 의 「rationale 로 재분류 → 종류 불일치 FAIL」을 낼 근거가 REQ-ALB-015 에 없음, AC-ALB-021 이 문자 그대로 정상 원장에서 실패함. → FAIL.

### Must-Pass
- [PASS] MP-1: REQ-ALB-001~025 전부 정의, 중복·결번 없음(spec.md:L71–L107). 023~025 가 011 과 012 사이(L85–L87)에 있음 — 번호 무결성엔 영향 없음(N8 유지, optional).
- [PASS] MP-2 (요구 층에서 판정): 25 REQ 전부 GEARS 또는 레거시 `shall not`(008·013·014·019). REQ-ALB-010(L83)은 `When … the hook shall …` 복합형. AC 는 검증 층 Given-When-Then(acceptance.md:L48–L57) — 이 판정에서 제외.
- [PASS] MP-3: spec.md:L2–L14 12필드, `version: "0.4.0"` 인용, `status: draft`, `priority: P1`, `lifecycle: spec-anchored`, 거부 별칭 없음.
- [PASS] MP-4: 다언어 툴링 SPEC 아님. REQ-ALB-019(L101)가 16개 언어 중립성을 요구.
- [PASS] MP-5: 본문 참조 SPEC 5개 상태를 직접 읽음 — HEADROOM-001·DIET-002·DIET-001·INSTRUCTION-BUDGET-SCOPE-001·INSTRUCTIONS-BUDGET-001 모두 `status: completed`(각 spec.md:L5). retired/superseded/archived 없음 → D7 BLOCKING 없음. D7 verb 는 Bash 거부로 미실행, 수기 대조.
- [PASS] MP-6: spec.md 전문(L1–L159)을 읽었고 `syscall` 0회.
- [PASS] MP-7: plan.md(L1–L124)·research.md(L1–L82) 전문을 읽었고 `[NEEDS CLARIFICATION` 0건. Q4·Q8 은 EVIDENCE-NEEDED 라벨(마커 아님).
- [FAIL] MP-8: codex 가 HEAD `6ade6fe38`(템플릿 diff `b5815ca80..HEAD` 무출력 → 문서 핀 성립)에서 재실행. 그대로 재현: AC-001/002/003/005/006/007(grep exit 1, 무출력), 004(`43138`, exit 0), 009(exit 1), 014(exit 1), 019(ls exit 1 — 다만 메시지는 stderr), 023(exit 1), 024(`1`, exit 0), 025(`2`, exit 0). **불재현**: AC-ALB-008 과 AC-ALB-015(그리고 이를 승계한 016·017·018)는 표에 적힌 명령이 `grep: T/rules/…: No such file or directory`, exit 2 를 낸다. `T` 를 펼치면 RED 신호는 재현되지만, AC-008 에 적힌 stdout(L29: `…/cross-session-messaging.md:1` `…/kanban-dispatch.md:1`)은 생략부호를 쓰고 출력 순서가 실제(kanban-dispatch 먼저)와 반대라 축자가 아니다. MP-8 실행 규율(「인용 명령이 실행되지 않으면 pass 로 기록하지 않는다」)에 따라 FAIL → D3-1, critical. 테스트 0개 실행 출력은 더 이상 RED 근거로 쓰이지 않음(L17) — iter2 지적은 해소.
- [PASS] MP-9: CN-4 awk verb 는 세션 가드로 GAP(미실행) → 마일스톤 순서를 손으로 읽음. plan 순서 M0(L55)→M1(L68)→M2(L75)→M3(L85)→M4(L91)→M5(L96), `Exit:` 바인딩 0. 순서 조항 대조: AC-002(L23) 「M1 RED 커밋(앵커 뒤 착지)」 ↔ plan L71 「M1 RED 커밋은 M0 의 첫 흡수·앵커 커밋 뒤에 착지」 — 합치. AC-025(L46) 「하한 > 115,000 이면 M1 이후 커밋 없이 리드 보고」 ↔ plan L65 — 합치. 「재고정 커밋이 그 흡수 뒤 첫 규칙 편집보다 앞섬」 ↔ plan §C.2 4·5단계(L50–L51) — 합치. 충돌 없음.

### Scores
| Dim | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | 단위 경계는 spec §B L57 단일 정의(design L25 는 인용만 함) — N6 해소. 감점: 코드 펜스 내부 규칙(펜스 속 `#` 줄·빈 줄·토큰)은 acceptance §D.2 L66 에만 있음. AC-021 문구가 위치 어휘와 충돌(D3-4) |
| Completeness | 0.75 | 0.75 | `entry_points` 스키마 완비(spec L61, design L74). 감점: 토큰 없는 절·파일의 규범 문단이 원장 범위 밖(D3-2). spec §H L158 「결정 9건(해소 6…)」 ↔ decision-index 실제 8건(해소 5) |
| Testability | 0.75 | 0.75 | 초록 조건은 앵커된 `-run` + 이름 붙은 `--- PASS:`(L15). AC-012 (a)(b)(c) 경계 픽스처 명확. 감점: AC-020(c) 후반부는 REQ 상 근거가 없어 구현이 만들 수 없는 FAIL 을 요구(D3-3), AC-021 은 정상 구현에서 실패(D3-4) |
| Traceability | 0.95 | 1.00 band 근접 | AC 표 수기 대조: REQ-001~025 전부 AC ≥1개, 존재하지 않는 REQ 인용 0(AC-019→012·015·025, AC-020→015·012, AC-025→021·022·015). 감점: AC-020(c) 가 REQ-015 에 없는 실패 조건으로 추적됨. 추적성 awk verb 는 GAP — 수기 |

평균 0.80. iter2 0.81 대비 하락 → **STOP 신호**(LEAN 규칙). 하락은 iter2 수리의 퇴행이 아니라 이번 반복에서 새로 찾은 결함 두 건(D3-3·D3-4) 때문이다. N1·N2·N4·N5·N6 수리 자체는 실질이 있다.

## 2. Evidence

이 감사의 직접 관측(Read):
- spec.md:L57 단위 경계 단일 정의, L59 규범 문단 정의, L60 「구속 절 — 구속 토큰을 가진 줄이 하나라도 있는 절」, L61 원장 행 스키마의 `entry_points`(`binding`·`normative` 필수, `delivery` ∈ `always`/`REQ-ALB-007`/`REQ-ALB-024`), L62 `companion:` 은 `rationale` 행만.
- spec.md:L94 REQ-ALB-015 실패 조건은 6개: 행 누락 / `entry_points` 누락·빈값 / role-core 진입점의 delivery 가 007·024 아님 / 위치 어휘 밖 또는 binding·normative 행의 `companion:` / `always:` 변경 후 텍스트 부재 / `role-core:` 변경 후 텍스트 부재. **앵커 시점의 종류와 비교하는 조건이 없다.**
- acceptance.md:L41 AC-020(c): 「그 문단을 `rationale` 로 바꿔 companion 으로 옮긴 변이 → 종류 불일치로 `--- FAIL`」.
- design.md:L29 「구속 절 밖의 문단은 원장 대상이 아니며 분할 대상이 될 수 있다」. plan.md:L22 는 `goal-directive`·`moai-mcp-tools` 를 「역할 경로로 보내면 일반 세션이 의무를 잃으므로」 M3 core+companion 분할 대상으로 둔다.
- 템플릿 `goal-directive.md`(L1–L48) 전문: `[HARD]`·`MUST`·`shall ` 없음(research §1 L28 「구속 줄 0」과 일치). 그런데 L25 「Arming a goal does not authorize autonomous run-phase entry … never authorizes creating a PR or a destructive operation」, L35 「an armed goal does not relax the "confirm before hard-to-reverse / shared-system actions" boundary」는 안전 의무다.
- acceptance.md:L42 AC-021: 「원장 테스트가 위치가 `paths:` 파일·skill 인 행 0 을 단정」 ↔ spec L62 는 `rationale` 행에 `companion:`(companion 은 `paths:` 파일)을 허용하고, REQ-ALB-014(L93)·Q5 로 역할 한정 규칙 2개(`role-core:` 행의 위치)가 최상위 `paths:` 를 갖는다.
- 템플릿 `cross-session-messaging.md`:L39–L47 — AC-020(c) 가 가리키는 토큰 없는 문단이 존재하며 L31 `[HARD]` 블록과 같은 「Rules」 절 안에 있음 → spec §B 정의로 원장 단위가 됨(N3 의 원래 사례는 닫힘).
- decision-index.md: Q1~Q8 8건(RESOLVED Q1·Q2·Q5·Q6·Q7 = 5, OUT OF SCOPE Q3, EVIDENCE-NEEDED Q4·Q8).

codex 재실행 증거(`/tmp/alb-iter3-evidence-6ade6fe38.md` — 이 감사가 그 파일을 읽어 인용; 명령 자체는 이 세션에서 재실행하지 못함):
- 측정 트리 HEAD `6ade6fe387d777a721a4b46d9ede062a86ec69cb`, `git status --short` 공란, `git diff --stat b5815ca80 HEAD -- internal/template/templates` 공란.
- `grep -c 'Intentionally always-loaded' T/rules/moai/workflow/kanban-dispatch.md T/rules/moai/workflow/cross-session-messaging.md` → stderr `grep: T/rules/moai/workflow/kanban-dispatch.md: No such file or directory` ×2, exit 2. 펼친 형태 → `…/kanban-dispatch.md:1` / `…/cross-session-messaging.md:1`, exit 0.
- `grep -c moai:role-rules-required T/agents/moai/manager-lead.md …` → No such file ×3, exit 2. 펼친 형태 → 세 파일 `:0`, exit 1.
- `wc -m …/kanban-dispatch-detail.md` → `   43138 …`, exit 0.
- `ls internal/template/testdata/binding_ledger.json` → stdout 없음, stderr `ls: …: No such file or directory`, exit 1.
- `grep -c 'pending run-phase' …/progress.md` → `2`, exit 0(이 감사도 progress.md L14·L18 두 줄을 읽어 확인).
- 이름 지정 `go test … -run '^TestDeployedAlwaysLoadedCharBudget$' -count=1 -v` → `testing: warning: no tests to run` / `ok … [no tests to run]`, exit 0 (0개 실행 — 표가 이를 RED 로 쓰지 않음을 확인).
- 필터 없는 `go test ./internal/template/` → `ok … 385.871s`, exit 0 — acceptance.md:L17 의 「`-run` 앵커 없이 돌린 … `[no tests to run]`」 참고 기술은 재현되지 않음(D3-6).

## 3. Baseline-attribution

- SPEC 트리: agent-a6d780d7ba07bc24c, 브랜치 `WT-always-loaded-budget`, HEAD `6ade6fe38` (codex 가 같은 트리에서 `git rev-parse` 로 관측). 이 감사의 Read 는 같은 트리 작업본을 읽음.
- 템플릿 바이트 = `b5815ca80`: codex 의 `git diff --stat b5815ca80 HEAD -- internal/template/templates` 공란 출력으로 확인 — acceptance L14 의 문서 핀이 성립.
- moai MCP 서버 빌드 `45600e4ee` — HEAD 의 조상(build_lag 경고). audit_multi 의 판정 대상은 diff 와 트리 파일이라, 서버 빌드 지연이 판정 내용에 영향을 주지 않음.

## 4. Defects (structured)

D3-1. RED-NOW-LITERAL (iter2 N1 잔존 면) — acceptance.md:L14,L29,L36 (+L37–L39 승계) — 표에 적힌 명령이 약어 `T` 때문에 그대로는 실행되지 않음(exit 2). AC-008 stdout 은 생략부호를 쓰고 순서가 실제와 반대라 축자 아님. AC-019 의 인용 「stdout」은 실제로 stderr — Severity: critical (MP-8) — Class: blocking — Required fix: AC-008·015 명령에서 `T` 를 전체 경로로 펼치고, 실제 출력을 그대로 붙임(AC-008 은 kanban-dispatch 줄 먼저, 생략부호 없음). AC-019 는 「stderr → … (exit 1), stdout 없음」으로 표기. 수리는 기계적이며, RED 신호 자체는 재현된다.

D3-2. TOKENLESS-SECTION-OBLIGATIONS (iter2 N3 잔존 면) — spec.md:L59–L61, design.md:L29, plan.md:L22; 원문 goal-directive.md:L21,L25,L35 — 원장 범위가 「구속 토큰 줄이 있는 절」로 한정됨. 토큰이 전혀 없는 절·파일(goal-directive.md 전체, 토큰 줄 하나뿐인 moai-mcp-tools.md 의 나머지 절, 그리고 토큰 있는 H2 뒤 하위 제목으로 끊긴 토큰 없는 H3)의 안전 의무는 행이 없음. 그래서 REQ-012 의 무손실 보장도 REQ-017 의 「companion 엔 근거만」도 기계적으로 걸리지 않는다. plan 이 바로 그 파일을 의무가 있다며 분할 대상으로 예정함 → 운영자 Q1 조건(「구속 의무는 하나도 떨어뜨리지 않는다」)이 이 형태에서 깨질 수 있음 — Severity: major — Class: blocking — Required fix: 원장 범위를 「이 SPEC 이 편집하는 상시 표면 파일의 모든 절」로 넓혀 토큰 없는 단위도 `normative`/`rationale` 로 분류하거나, 최소한 M3 분할 대상 파일(goal-directive·moai-mcp-tools·CLAUDE.md 포함)의 전 절을 원장 대상으로 명시. 새 AC 번호 대신 AC-020 에 goal-directive L25 문단 삭제·companion 이동 변이 → `--- FAIL` 하위 테스트를 추가.

D3-3. KIND-RECLASSIFICATION-UNGUARDED (신규) — acceptance.md:L41 vs spec.md:L94 — AC-020(c) 는 「normative → rationale 재분류 + companion 이동 → 종류 불일치 FAIL」을 요구하지만 REQ-ALB-015 엔 앵커 시점 종류와의 대조가 없고 원장의 `kind` 는 수정 가능한 단일 열. 재분류된 행은 6개 조건을 모두 통과하므로, AC 가 요구하는 FAIL 은 명세된 구현에서 나오지 않고 막으려는 우회로는 열린 채 남음. §D.1 시나리오(L51)는 삭제만 언급해 AC 본문과도 다름 — Severity: major — Class: blocking — Required fix: REQ-ALB-015 에 실패 조건 추가: 「행의 종류가 앵커에서 고정된 종류와 다를 때」 — M0 기준선 커밋에 단위별 종류를 고정(원장 머리 또는 별도 고정물)하고 §C.2 재고정 커밋에서만 바꾸게 함. 아울러 `companion:` 행의 변경 후 텍스트가 해당 companion 에 있는지 대조하는 조건도 추가. 아니면 AC-020(c) 후반부를 삭제.

D3-4. AC021-LOCATION-CONTRADICTION (신규 발견, iter2 에서 놓침) — acceptance.md:L42 vs spec.md:L62, L93 — 「위치가 `paths:` 파일·skill 인 행 0」은 정상인 `rationale`→`companion:` 행과, Q5 로 최상위 `paths:` 를 갖게 되는 역할 한정 규칙 2개를 가리키는 `role-core:` 행에서 실패한다. 후반부 「companion 안에서 원장 블록 텍스트 조각 발견 0」은 조각 단위가 정의되지 않아, 앵커 시점에 이미 companion 에 있던 짧은 줄로도 실패할 수 있다 — Severity: major — Class: blocking — Required fix: 「`binding`·`normative` 행 가운데 위치가 `companion:` 이거나 skill 인 행 0」으로 고치고, 조각 규칙을 정의(예: `binding`·`normative` 변경 후 텍스트의 N 코드 단위 이상 전체 줄, 앵커 시점 그 companion 에 이미 있던 텍스트 제외).

D3-5. ENTRY-POINT-SELF-DECLARED — spec.md:L94 — `delivery: REQ-ALB-024` 라고 적힌 진입점이 실제로 `moai:role-rules-required` 표지를 단 배포 파일인지, `REQ-ALB-007` 진입점이 레지스트리 표지인지를 검사하지 않음. `entry_point` 식별자 형식도 미정 — Severity: minor — Class: optional — Required fix: 식별자 형식(레지스트리 표지명 또는 배포 경로)을 정하고, 원장 테스트가 024 진입점을 표지 있는 배포 파일로, 007 진입점을 레지스트리 표지로 해석하게 함.

D3-6. AC-HEADER-FALSE-REFERENCE — acceptance.md:L17 — 「`-run` 앵커 없이 돌린 `go test ./internal/template/` 는 … `[no tests to run]`」은 재실행으로 재현되지 않음(실제 `ok … 385.871s`). 참고 기술이지 RED 셀은 아니지만, 관측되지 않은 출력을 적은 것 — Severity: minor — Class: optional — Required fix: 실제 쓴 `-run '^TestDeployedAlwaysLoadedCharBudget$'` 명령과 그 관측 출력으로 교체.

D3-7. FENCE-RULE-OUTSIDE-SSOT (N6 잔여) — spec.md:L57 vs acceptance.md:L66 — 펜스 내부의 불투명성(펜스 속 `#` 줄·빈 줄·토큰이 단위·절을 끊지 않음)은 acceptance 에만 있음 — Severity: minor — Class: optional — Required fix: §B 「단위 경계」·「구속 절」에 펜스 내부 불투명 규칙을 옮기고, AC-020(a) 에 펜스 속 `#` 줄 픽스처 추가.

D3-8. STALE-REFS — spec.md:L158 「결정 9건(해소 6…)」 ↔ 실제 8건(해소 5). AC-002 초록 조건(L23)의 「17개」 하드코딩은 앵커 도출 원칙(L18)과 긴장. research.md 머리(L11)는 「모든 수치는 `b5815ca80`」인데 §1.1 은 `d7112d005`. progress.md:L3 「base `b5815ca80`」 — Severity: minor — Class: optional — Required fix: 수치·참조 정정, 「17개」는 「원장 머리의 구성원 목록」으로.

D3-9. AC004-RED-SHAPE — acceptance.md:L25 — `wc -m` exit 0 은 L17 이 스스로 정한 「비어 있지 않은 실패 신호만」 규칙에 맞지 않고(검사 부재는 AC-001 grep 이 보임), `wc -m` 은 로캘에 따라 바이트를 셈 — Severity: minor — Class: optional — Required fix: RED 근거를 AC-001 grep 으로 두고, 43,138 은 research §1 의 UTF-16 계수기로 참고 표기.

D-N8 (iter2, optional) — REQ 순서 — 유지(무해).

## Regression Check (iter2 → iter3)
- D-N1 RED-NOW-INCOMPLETE — PARTIAL: AC-004(L25)·007(L28)·023(L44) 모두 명령·stdout·exit 기재, 0개 실행 출력은 RED 근거에서 제외(L17). 표에 적힌 그대로의 명령 형태 결함 → D3-1.
- D-N2 LEDGER-ENTRYPOINT-FIELD — RESOLVED: spec L61 필수 `entry_points`, REQ-015 L94 누락/빈값/미덮음 실패 조건, AC-019 L40 세 변이. 자기 신고 잔여 → D3-5(optional).
- D-N3 TOKENLESS-NORMATIVE-PARAGRAPH — PARTIAL: 원래 사례(cross-session L39–L47)는 spec L59·design L28·AC-020(c) 전반부로 닫힘. 토큰 없는 절·파일 → D3-2. 재분류 우회 → D3-3.
- D-N4 BASELINE-PINNED-VS-ABSORPTION — RESOLVED: acceptance L18 앵커 기대값, plan L57 M0 첫 행동 = 흡수·앵커, L71 M1 RED 는 앵커 뒤 + 템플릿 diff 무출력, §C.2 4단계(L50) 기대값 재측정 별도 커밋.
- D-N5 SIZE-CAP-SCOPE — RESOLVED: REQ-010 L83 최종 `additionalContext` 합본, design L65, AC-012 L33 (a)(b)(c) 합본 경계 픽스처. (c) 경우의 최종 길이 미한정은 의도된 동작(경고 + 지시)으로 판단.
- D-N6 BLOCK-BOUNDARY-WORDING — RESOLVED: spec L57 유일 정의, design L25 인용만. 펜스 규칙 잔여 → D3-7(optional).
- D-N7 — RESOLVED: plan L72 AC-ALB-020.
- D-N9 — RESOLVED: AC-022 L43 13개 경로 명시(codex 재실행 exit 1 무출력).
- D-N10 — RESOLVED: AC-025 L46 RB + RED-now `2`.
- D-N8 — 유지(optional, 무해).

Stagnation: 세 반복 모두에 같은 모습으로 남은 결함은 없음. N1·N3 계열은 반복마다 면이 줄어듦 — 진전 있음.

## Recommendation (manager-spec / 리더 에스컬레이션)

iter3 이 상한이므로 리더가 운영자 채널로 세 선택지를 올려야 한다: (1) PASS-with-debt, (2) 범위 축소, (3) iter4 명시적 연장. 감사자 관점에서는 blocking 4건 모두 문서 몇 줄로 끝나는 수리이고 새 AC 번호가 필요 없다(25/25 상한 유지):
1. D3-1: AC-008·015 명령의 `T` 를 펼치고 실제 출력을 축자로 붙임. AC-019 는 stderr 로 표기.
2. D3-2: 원장 범위를 분할 대상 파일의 전 절로 넓힘(spec §B L60–L61, design L29). AC-020 에 goal-directive L25 변이 하위 테스트.
3. D3-3: REQ-015 에 「앵커 고정 종류와 불일치」·「`companion:` 변경 후 텍스트 부재」 실패 조건 추가. 아니면 AC-020(c) 후반 삭제.
4. D3-4: AC-021 을 `binding`·`normative` 행 한정으로 다시 쓰고 조각 규칙을 정의.
D3-5~D3-9 는 optional.

## 5. Gaps
- 이 세션의 Bash 는 모든 호출이 워크트리 격리 가드에 거부됨(매번 다른 격리 트리를 지목). Grep·Glob 도구는 이 세션에 없음. 따라서 이 감사는 RED-now 명령을 직접 재실행하지 못했다. 재실행 증거는 audit_multi 의 codex 백엔드가 대상 트리에서 쓴 `/tmp/alb-iter3-evidence-6ade6fe38.md` 를 읽어 인용함(측정자: codex). 대체 경로는 Read 직독.
- CN-4 순서 verb·추적성 awk verb·D7/D8 verb 미실행 → 수기 대조(MP-9·추적성·MP-5·MP-6).
- Plan Artifact Hash 미산출.
- HEAD `6ade6fe38` 는 이 세션이 아니라 codex 가 관측함. iter2 대비 문서 diff(`git diff 15de1ad21 6ade6fe38`)는 보지 못했고, 델타는 v0.4.0 HISTORY(spec L24)와 본문 직독으로 대조함 → D3-4 가 iter2 이전부터 있던 문구인지는 미확인.
- glm inconclusive. `audit_receipt` 미발급.

## 6. Residual-risk
- `rewrite` 행의 의미 보존과 `normative`/`rationale` 분류는 사람 판정으로 남음(D3-3 가 고쳐져도 앵커 시점 분류 자체의 정확성은 감사가 행마다 확인해야 함).
- research §2 개산상 M0 가 하한 > 115,000 으로 멈출 가능성이 여전히 높음(Q6 경로).
- D3-2 를 넓히면 원장 행 수와 M0 작업량이 커짐 — 범위 축소 선택지와 맞물림.
- 진입점 열거의 완전성(레인이 띄운 specialist 등)은 D3-5 를 고쳐도 사람 판정이 남음.

## Iteration History
- iter1 (2026-10-03, HEAD 5bfff5682): FAIL 0.69, blocking D1–D8.
- iter2 (2026-10-03, HEAD 15de1ad21): FAIL 0.81, blocking D-N1–D-N6.
- iter3 (2026-10-03, HEAD 6ade6fe38): FAIL 0.80 (STOP — 하락, 상한 도달), N2·N4·N5·N6 해소, N1·N3 부분 해소. blocking D3-1(critical, MP-8)·D3-2·D3-3·D3-4. optional D3-5–D3-9.
