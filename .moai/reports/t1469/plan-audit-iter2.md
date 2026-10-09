auditor-model: claude-opus-5-5

verdict: FAIL
audited_sha: 15de1ad217326996931af68659112ec8c9e0c92c

# SPEC Review Report: SPEC-ALWAYS-LOADED-BUDGET-001
Iteration: 2/3 (delta + regression check; CN-4 re-read in full)
Verdict: FAIL
Overall Score: 0.81 (Tier L 문턱 0.85 미달; iter1 0.69 → 상승, STOP 신호 없음)
Plan Artifact Hash: 파일별 sha256 — acceptance 0ce12441… · design 10a33cd0… · plan f132db77… · research b4fbdc79… · spec 2ff2a26a… (합본 해시는 Go `ComputeHash` 미실행 — Gaps)
Auditor Version: plan-auditor (card t1469, iter2)
Intended destination: <tree agent-a6d780d7ba07bc24c>/.moai/reports/t1469/plan-audit-iter2.md — 세션 격리 가드가 그 트리의 git·awk 실행을 거부해 스크래치에 기록. 리드가 복사·`git add -f` 필요.

Reasoning context ignored per M1 Context Isolation. 운영자 판정(Q1 동결 해제·v3.2.0 제외)과 리더 판정(Q5 수용)은 재론하지 않음.

Cross-model: `audit_multi`(project_root=agent-a6d780d7ba07bc24c, baseBranch) — claude FAIL(anchor) · codex FAIL(required) · glm inconclusive(advisory, 응답 본문 없음). overall `fail`, disagreement_flag `false`, audit_receipt 미발급. codex 가 독립적으로 N1·N2·N4 를 같은 줄에서 짚었고, N3·N5 두 건을 추가로 냈다(이 감사가 원문 대조로 확인).

## 1. Claim

iter1 blocking 8건 가운데 D1(줄→블록 단위)·D2·D3(명명 진입점)·D4·D5(절차)·D7 은 실제로 닫혔다. D8 은 일부만 닫혔다 — release-blocking AC 3개(AC-ALB-004·007·023)에 RED-now 의 명령·exit 가 없다(MP-8 위반). 수정이 새로 낳은 결함 4건이 blocking 이다: 원장 스키마에 AC-ALB-019 가 단정하는 진입점 필드가 없음(N2), 토큰 없는 규범 문단이 블록 경계 밖으로 빠짐(N3), AC-ALB-002 기대값이 `b5815ca80` 에 묶였는데 develop 이 이미 그 표면을 바꿈(N4), REQ-ALB-010 상한이 최종 `additionalContext` 전체가 아니라 core 만 잼(N5). → FAIL.

### Must-Pass
- [PASS] MP-1: REQ-ALB-001~025 모두 존재, 중복·결번 없음(spec.md:L67–L103). 문서 순서상 023~025 가 011 과 012 사이(L81–L83)에 있음 — 번호 무결성은 성립, 순서는 minor(N8).
- [PASS] MP-2 (요구 층): 25 REQ 전부 GEARS 또는 레거시 `shall not`(008·013·014·019). REQ-ALB-010 은 iter1 D10 대로 `When` 으로 고쳐짐(L79). AC 는 검증 층 Given-When-Then(acceptance.md:L46–L54) — 이 판정에서 제외.
- [PASS] MP-3: spec.md:L2–L14 12필드, `version: "0.3.0"` 인용, 거부 별칭 없음.
- [PASS] MP-4: 다언어 툴링 SPEC 아님; REQ-ALB-019 가 템플릿 중립성 요구.
- [PASS] MP-5: 참조 SPEC 5개 iter1 에서 `completed` 확인, 새 참조 없음(본문 SPEC ID 집합 불변).
- [PASS] MP-6: `syscall` 0회.
- [PASS] MP-7: plan.md·research.md `[NEEDS CLARIFICATION` 0건. Q4·Q8 은 EVIDENCE-NEEDED 라벨로 M0 측정에 걸려 있음(마커 아님).
- [FAIL] MP-8: 재실행 결과는 §2 Evidence. 재현 성공 — AC-001·008·009·014·015·019·024. 그러나 RB 인 AC-ALB-004(L23: 「`research.md` §1 측정」만), AC-ALB-007(L26: 「191,386 > 115,000(리드 실측)」만), AC-ALB-023(L42: 「미측정」)은 명령·stdout·exit 가 없다 → §2.1 상 미채택 셀. AC-ALB-002(L21)의 RED 명령은 재실행 시 `ok … [no tests to run]` exit 0 — 0개 실행은 아무것도 재현하지 않는다(MP-8 판정 규칙). SPEC 이 이를 「테스트 부재」로 정직하게 표기했으나 셀 자체는 RED 를 재현하지 않으므로 AC-001 grep 을 RED 로 인용하도록 고쳐야 한다. → N1, critical.
- [PASS] MP-9: CN-4 awk verb 는 세션 가드에 거부(GAP) → 수동 판독. plan 마일스톤 순서 M0→M1→M2→M3→M4→M5(plan.md:L55–L94), `Exit:` 바인딩 0. 순서 조항: AC-ALB-025(L44) 「하한 > 115,000 이면 M1 이후 커밋 없이 보고」 = plan M0 중단 지점(L63)과 합치; 「재고정 커밋이 흡수 뒤 첫 규칙 편집보다 앞섬」 = plan §C.2 4·5단계(L50–L51)와 합치. AC-ALB-002 「M1 RED 커밋의 템플릿 = `b5815ca80`」 은 plan §C/§C.2 가 흡수 시점을 M1 앞뒤로 제약하지 않아 **필연적 충돌은 아님**(M1 을 첫 흡수 전에 두면 만족) — 그러나 순서가 문서에 없고 develop 이 이미 이동했으므로 N4 로 blocking 처리.

### Scores
| Dim | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | 블록 경계 정의가 spec §B(L56 「빈 줄이나 제목 전까지」)와 design §2(L25 「같거나 높은 수준의 제목」·「토큰 없는 새 산문 문단」)에서 갈림(N6) |
| Completeness | 0.75 | 0.75 | 7 산출물·Out of Scope H3 3개(L123·L128·L134). 원장 행 스키마(spec L57, design L75)에 진입점 필드 부재(N2), REQ-010 상한 대상 불완전(N5) |
| Testability | 0.75 | 0.75 | 초록은 앵커된 `-run` + 이름 붙은 `--- PASS:` 줄로 고정(acceptance L15) — iter1 D9 해소. 감점: AC-002·003·004·007 기대값이 이미 움직인 기준선에 묶임(N4), RB 3개 RED 무증거(N1) |
| Traceability | 1.00 | 1.00 | AC 표 수기 대조: REQ-001~025 전부 ≥1 AC, 존재하지 않는 REQ 인용 0 (AC-019 → 012·015·025, AC-024 → 018·019, AC-025 → 015·021·022 등). awk 추적 verb 는 가드 거부로 미실행 — 수기 대조 |

## 2. Evidence

RED-now 재실행 (트리 agent-a6d780d7ba07bc24c 루트, HEAD `15de1ad217326996931af68659112ec8c9e0c92c` — gitdir `HEAD`·ref 파일 직독):
- AC-001 `grep -rl TestDeployedAlwaysLoadedCharBudget internal/template` → 무출력, exit 1. 재현.
- AC-002 `go test ./internal/template/ -run TestDeployedAlwaysLoadedCharBudget -count=1 -v` → `testing: warning: no tests to run` / `PASS` / `ok  github.com/modu-ai/moai-adk/internal/template 0.424s [no tests to run]`, exit 0. 0개 실행 — RED 재현 아님.
- AC-008 `grep -c 'Intentionally always-loaded' …/kanban-dispatch.md …/cross-session-messaging.md` → `…cross-session-messaging.md:1` `…kanban-dispatch.md:1`, exit 0. 재현.
- AC-009 `grep -l -e role-core -e kanban-dispatch internal/hook/session_start_kanban.go … subagent_start.go` → 무출력, exit 1. 재현.
- AC-014 `grep -rl 'moai:role-core-start' internal/template/templates` → 무출력, exit 1. 재현.
- AC-015 `grep -c moai:role-rules-required …manager-lead.md …moai-kanban-foreman/SKILL.md …gtd.md` → 세 파일 각 `:0`, exit 1. 재현.
- AC-019 `ls internal/template/testdata/binding_ledger.json` → `No such file or directory`, exit 1. 재현.
- AC-022(RG) 13개 기준 규칙에 `grep -l '^paths:'` → 무출력, exit 1. 재현.
- AC-024 `grep -c -e kanban-dispatch -e cross-session-messaging internal/template/rule_template_mirror_test.go` → `1`, exit 0. 재현.

D3 도달 확인:
- `grep -rlE 'kanban-dispatch|cross-session-messaging' internal/template/templates` → agent·skill·workflow 인용 파일: `agents/moai/manager-lead.md`, `skills/moai-kanban-foreman/SKILL.md`, `skills/moai-lane-watchdog/SKILL.md`, `skills/moai/workflows/gtd.md` (+ `.codex/agents/moai/manager-lead.toml` 방출본). watchdog 은 AC 목록에 없으나 AC-ALB-018 전수 가드가 잡는다(레인 표지로도 덮임).
- manager-lead.md:L202 「deputy — a manager-lead instance obtained through `subagent-spawn`… same skill set」 → 정의 본문 표지가 deputy 에 결정적으로 도달. gtd.md:L327·L339 `--auto` 절 존재. moai/SKILL.md:L82 `todo` 는 gtd 로 라우팅 → `/moai:todo --auto` 도 gtd.md 로 도달.

D2: design.md:L39 표지는 `<!-- moai:evolvable-start -->` 계열 — 그 표지는 배포 템플릿 다수(예 `rules/moai/core/moai-constitution-detail.md`, `skills/moai-foundation-core/SKILL.md`)에 이미 있음(grep 실측). REQ-ALB-023(L81)·AC-ALB-014(L33) 원장 없는 `t.TempDir()` 생성 검사.

N3 원문: `cross-session-messaging.md:L39–L47` (템플릿) — L38 빈 줄 뒤 토큰 없는 문단 「A `STOPPED_TEAMMATE_VIOLATION` deny is never a bug to route around: … route coordination through the owning orchestrator, or respawn the name deliberately.」 design §2 L25 규칙(「토큰 없는 새 산문 문단에서 끝난다」)으로는 원장 밖.

N4 실측(이 세션 트리 develop `7142018fb`): `git diff --stat b5815ca80 HEAD -- internal/template/templates/.claude/rules internal/template/templates/CLAUDE.md internal/template/templates/AGENTS.md.tmpl` → `core/native-idiom-and-register.md | 2 +-`, `design/constitution.md | 6 +-`, `development/agent-authoring.md | 4 +-`, `AGENTS.md.tmpl | 91 +++…`, `CLAUDE.md | 4 +-` (5 files, +54 −53); `git merge-base --is-ancestor b5815ca80 HEAD` exit 0. 상시 규칙 하나(native-idiom)와 `CLAUDE.md` 가 바뀜 → 렌더 독립 소계와 원장 기준선 모두 이동. codex 실측: `b5815ca80=169018`, develop `7142018fb=169291`(이 감사가 재측정하지 않음 — Gaps).

N5 원문: `internal/hook/session_start.go:L438–L446` 이 같은 `AdditionalContext` 에 세션 귀속 3줄을 이미 넣는다. REQ-ALB-010(spec L79)은 「the role core exceeds…」로 core 크기만 비교.

## 3. Baseline-attribution

- SPEC 트리: agent-a6d780d7ba07bc24c, 브랜치 `WT-always-loaded-budget`, HEAD `15de1ad21` (ref 파일 직독; 그 트리 git 은 가드 거부). 템플릿 바이트는 SPEC 표기상 `b5815ca80`; RED-now grep 9건은 이 실행에서 그 트리 작업본을 직접 읽음.
- 상류 이동 측정: 이 세션 트리(develop `7142018fb`)의 git.
- moai MCP 서버 빌드 `45600e4ee` — HEAD 의 조상(build_lag 경고). audit_multi 의 검토 대상은 git diff 이므로 빌드 지연이 판정 내용에 영향 주지 않음.

## 4. Defects (structured)

D-N1. RED-NOW-INCOMPLETE (iter1 D8 잔존) — acceptance.md:L21,L23,L26,L42 — RB AC-ALB-004·007·023 에 명령·stdout·exit 없음; AC-ALB-002(·003·005 가 승계)의 RED 명령은 0개 실행으로 재현 불가 — Severity: critical (MP-8) — Class: blocking — Required fix: AC-004 는 `grep -c` 등 단일 명령으로 43,138 근거(또는 AC-001 grep)를 인용, AC-007 은 AC-001 grep 을 RED 로 인용, AC-023 은 해석 테스트 부재를 보이는 grep 을 넣거나 RG 로 강등; AC-002/003/005 의 RED-now 칸을 AC-001 grep 인용으로 바꾸고 go test 출력은 참고로만.
D-N2. LEDGER-ENTRYPOINT-FIELD — spec.md:L57, design.md:L75 vs acceptance.md:L38 — AC-ALB-019 는 「`role-core:` 행마다 진입점 목록이 REQ-007·024 덮는 집합 안」을 단정하지만 원장 행 스키마에 진입점 필드가 없다; 목록 누락·빈 목록 처리도 미정. REQ-ALB-025 분류의 기계 검사가 공허해짐 — Severity: major — Class: blocking — Required fix: 행 스키마에 필수 `entry_points` 필드(진입점별 전달 근거 REQ-007/REQ-024)를 추가하고, 누락·빈 목록·덮이지 않는 진입점에서 `--- FAIL` 하는 하위 테스트를 AC-019 초록 조건에 명시.
D-N3. TOKENLESS-NORMATIVE-PARAGRAPH (iter1 D1 의 남은 면) — design.md:L25, spec.md:L56; 원문 cross-session-messaging.md:L39–L47 — 구속 토큰 없는 별도 규범 문단(「deny is never a bug to route around … route coordination through the owning orchestrator」)이 블록 경계 밖이라 원장에 행이 없고, REQ-ALB-017 의 「companion 은 근거·사례·사건·상세표만」은 기계 검사되지 않아 무손실 조건(운영자 Q1)이 이 형태에서 깨질 수 있다 — Severity: major — Class: blocking — Required fix: (a) 블록이 「the prohibition above」류 역참조 문단·의무형 문장(“never… / route… / do not…”)을 포함하도록 경계를 넓히거나, (b) M0 원장에 「토큰 없는 규범 문단」 행 종류를 두고 처리 방식을 기록, 그리고 이 실제 문단을 지우는 변이에서 원장 테스트가 FAIL 하는 AC 를 추가.
D-N4. BASELINE-PINNED-VS-ABSORPTION (iter1 D5·D6 결합 잔존) — acceptance.md:L21–L23,L26; plan.md:L35,L45–L51,L69 — AC-002·003·004·007 과 M1 은 `b5815ca80` 템플릿·169,018·17 구성원에 묶였는데, §C.2 는 흡수 시 앵커만 바꾸고 이 기대값은 갱신하지 않는다. 측정상 develop 은 이미 `CLAUDE.md`·`native-idiom-and-register.md` 를 바꿨다 — Severity: major — Class: blocking — Required fix: 「M1 RED 커밋은 첫 develop 흡수보다 먼저」를 plan 에 명시하거나, §C.2 재고정 단계에 AC-002 기대 소계·구성원 목록·근거 SHA 재측정과 그 값의 별도 커밋을 넣는다. AC-025 「흡수마다 재고정 커밋」도 §C.2 1단계의 무변경 예외와 맞춘다.
D-N5. SIZE-CAP-SCOPE — spec.md:L79, acceptance.md:L31; 원문 internal/hook/session_start.go:L438–L446 — REQ-ALB-010 은 core·블록 크기만 Q4 한도와 비교하지만 같은 `additionalContext` 에 세션 귀속문·레인 공지·Read 지시가 합쳐진다. core 가 한도 이하라도 합본이 넘어 잘릴 수 있다 — Severity: major — Class: blocking — Required fix: 상한을 「최종 `additionalContext` 전체(기존 맥락 + 역할 core + 지시)」에 적용하도록 REQ-010 을 고치고, 합본 경계값 픽스처를 AC-012 에 추가.
D-N6. BLOCK-BOUNDARY-WORDING — spec.md:L56 vs design.md:L25 — 종료 조건이 「제목 전까지」(모든 제목) 대 「같거나 높은 수준의 제목」으로 다름; 추출기 구현이 갈린다 — Severity: minor — Class: blocking (내부 일관성) — Required fix: 한 정의로 통일(N3 수정과 함께), AC-020(a) 에 하위 제목 픽스처 추가.
D-N7. PLAN-AC-XREF — plan.md:L70 — 원장 변이 검사를 「AC-ALB-021」로 인용; 실제는 AC-ALB-020 — Severity: minor — Class: optional — Required fix: 번호 정정.
D-N8. REQ-ORDER — spec.md:L81–L83 — REQ-023~025 가 011 과 012 사이 — Severity: minor — Class: optional — Required fix: 절 내 번호 순 재배열 또는 그대로 두되 무해.
D-N9. AC022-COMMAND-PROSE — acceptance.md:L41 — RG 라 MP-8 밖이나 명령이 서술형(13개 경로 미기재) — Severity: minor — Class: optional — Required fix: 13개 경로를 적은 단일 명령으로.
D-N10. AC025-CLASS — acceptance.md:L44 — 새 의무(재고정 순서·M0 중단)를 RG(「이미 성립하는 성질」)로 분류 — Severity: minor — Class: optional — Required fix: RB 로 올리고 RED-now(`progress.md` §E.2 pending grep) 기재, 또는 분류 정의 문구 조정.

## Regression Check (iter1 → iter2)
- D1 LEDGER-LINE-UNIT — RESOLVED(줄 단위 문제): spec L56 구속 블록 정의, design §2 L25–L30, REQ-013 L88 「token line, continuation line, or subordinate item」, AC-020 L39 변이. 잔여 면은 D-N3 로 신규 등재.
- D2 RUNTIME-CLASSIFICATION-SOURCE — RESOLVED: REQ-023 L81, spec L53, design L39·L74, AC-014 L33.
- D3 ROLE-REACH-GAP — RESOLVED(명명 진입점): REQ-024 L82, REQ-025 L83, design §4 L44–L52, AC-015~018 L34–L37; deputy 도달은 manager-lead.md:L202 로 확인. 분류 검사의 공허함은 D-N2.
- D4 LEDGER-LOCATION-MODEL — RESOLVED: spec L58 어휘, REQ-015 L90 생성 함수 출력 대조, design L76.
- D5 LEDGER-BASE-DRIFT — RESOLVED(절차): plan §C.2 L45–L51. 기대값 미갱신은 D-N4.
- D6 AC002-RENDER-INPUTS — PARTIAL: 고정 렌더 입력 spec L48·렌더 독립 소계 design L87 은 해소; 기준 트리 고정은 D-N4 로 잔존.
- D7 Q5-FOUNDER-OPEN — RESOLVED: decision-index L33–L38(리더 결정 ACCEPTED), REQ-014 L89.
- D8 RED-NOW-ELEMENTS — UNRESOLVED(부분): 분류 열·대부분 셀 보강은 됨(L13–L44), AC-004·007·023 미보강 → D-N1.
- D9 (optional) — RESOLVED: acceptance L15.
- D10 (optional) — RESOLVED: spec L79.
- D11 (optional) — RESOLVED: REQ-018 L96, AC-024 L43.
- D12 (optional) — RESOLVED: REQ-011 L80, design L68–L70.

## Recommendation (manager-spec)
1. D-N1: acceptance.md L21·L23·L26·L42 의 RED-now 칸을 단일 명령 + stdout + exit 로 채우거나(AC-001 grep 인용 가능) RG 로 강등.
2. D-N2: 원장 행에 `entry_points` 필드 추가(spec L57, design L75), AC-019 에 누락/빈/미덮음 실패 하위 테스트.
3. D-N3 + D-N6: 블록 경계를 하나로 정의하고 토큰 없는 규범 문단을 원장에 포함시키는 규칙 + cross-session-messaging L39–L47 삭제 변이 AC.
4. D-N4: plan 에 M1 RED 커밋 ≺ 첫 흡수 순서를 적거나 §C.2 에 AC-002 기대값 재측정 단계 추가; AC-025 문구를 §C.2 1단계 예외와 맞춤.
5. D-N5: REQ-010 상한 대상을 최종 `additionalContext` 합본으로, AC-012 합본 경계 픽스처.
REQ/AC 수가 Tier L 상한 25/25 이므로 신규 검사는 기존 AC 의 초록 조건 안에 하위 테스트로 넣을 것(새 AC 번호 추가 시 상한 초과).

## 5. Gaps
- CN-4·추적성 awk verb 는 세션 가드(`awk -f` 거부)로 미실행 → 수기 대조로 대체. 그 트리의 `git` 도 전부 거부 → HEAD 는 gitdir ref 파일 직독, `git log`·diff 는 이 세션 트리(develop)에서만.
- 169,291(develop 렌더 독립 소계)은 codex 측정 — 이 감사 미재측정. 이 감사가 잰 것은 diff stat 의 파일 이동 사실뿐.
- Plan Artifact Hash 합본은 Go `ComputeHash` 미실행 — 파일별 sha256 만.
- glm inconclusive; audit_receipt 미발급.
- 판정서 지정 경로(대상 트리 `.moai/reports/t1469/`) 쓰기 미수행 — 스크래치에 기록.

## 6. Residual-risk
- 압축 재작성 행의 의미 보존은 사람 판정(운영자 수용 위험) — sync 감사의 행 단위 확인 필요.
- research §2 개산상 M0 가 하한 > 115,000 으로 멈출 가능성 여전히 높음(Q6 경로).
- REQ-025 분류는 D-N2 가 고쳐져도 「블록이 묶을 수 있는 진입점 전부」의 열거 완전성은 사람 판정으로 남음 — 레인이 띄운 specialist(manager-develop 등)를 묶는 블록이 role-core 로 잘못 분류될 위험.

## Iteration History
- iter1 (2026-10-03, HEAD 5bfff5682): FAIL 0.69, blocking D1–D8, optional D9–D12.
- iter2 (2026-10-03, HEAD 15de1ad21): FAIL 0.81, iter1 D1–D5·D7·D9–D12 해소, D6 부분·D8 부분 잔존; blocking D-N1(critical, MP-8)·D-N2·D-N3·D-N4·D-N5·D-N6, optional D-N7–D-N10.
