auditor-model: claude-opus-5-5

verdict: FAIL
audited_sha: 5bfff5682

# SPEC Review Report: SPEC-ALWAYS-LOADED-BUDGET-001
Iteration: 1/3
Verdict: FAIL
Overall Score: 0.69 (Tier L 문턱 0.85 미달)
Plan Artifact Hash: 미산출 (Gaps 참조)
Auditor Version: plan-auditor/v1 (card t1469)
Intended destination: <tree agent-a6d780d7ba07bc24c>/.moai/reports/t1469/plan-audit-iter1.md — 세션 격리 가드가 해당 트리 쓰기·git 을 거부해 스크래치에 기록함. 리드가 복사 후 `git add -f` 커밋 필요.

Reasoning context ignored per M1 Context Isolation. 운영자 판정(Q1·Q7)은 재론하지 않고 SPEC 반영 충실도만 판정.

Cross-model: `audit_multi`(project_root=agent-a6d780d7ba07bc24c, baseBranch) — claude FAIL(anchor) · codex FAIL(required) · glm inconclusive(advisory, 응답 본문 없음). overall `fail`, disagreement_flag `false`, audit_receipt 미발급.

## 1. Claim

구조(프론트매터·REQ 번호·추적성 22/22·RED-now 재현)는 건전하나, 운영자 Q1 의 조건 「구속 의무 무손실을 감사 가능한 원장」을 현 설계가 보장하지 못하고(줄 단위 원장), 역할 주입이 리더급 세션 전부(하위 에이전트·무표지 진입점)에 닿는다는 보장이 없다. blocking 8건, 0.69 < 0.85 → FAIL.

### Must-Pass
- [PASS] MP-1: spec.md:L61–L94 REQ-ALB-001~022 연속.
- [PASS] MP-2 (요구 층): 22 REQ 전부 GEARS 또는 레거시 `shall not`(008·013·014·019). REQ-010 `Where` 오용은 minor(D10). AC 는 Given-When-Then(검증 층).
- [PASS] MP-3: spec.md:L2–L14 12필드, version 인용, 거부 별칭 없음.
- [PASS] MP-4: 다언어 툴링 SPEC 아님; REQ-ALB-019 가 템플릿 중립성 요구.
- [PASS] MP-5: 참조 SPEC 5개 모두 `status: completed`.
- [PASS] MP-6: `syscall` 0회.
- [PASS] MP-7: plan.md·research.md `[NEEDS CLARIFICATION` 0건(Q5 공란은 D7).
- [PASS] MP-8: pin b5815ca80; HEAD 5bfff5682 는 .moai/specs 만 다름. AC-001 grep → 무출력 exit=1 재현; AC-009 grep → 무출력 exit=1 재현; AC-008 명령은 파일 인자 없어 실행 불가 → undecidable(회귀 가드 강등), 인자 보정 시 각 `:1`.
- [PASS] MP-9: CN-4 포트 `COLLECTED: 6 milestones (M0 M1 M2 M3 M4 M5), 0 exit bindings, 1 ordering candidates`; 후보 acceptance.md:34 는 마일스톤 간 순서 없음. M0→M1 과 AC-020 일치.

### Scores
| Dim | Score | Evidence |
|---|---|---|
| Clarity | 0.50 | 원장 위치 모델(D4)·런타임 분류 원천(D2)·흡수 후 base tree(D5)·stub 파일명 미정 |
| Completeness | 0.75 | 7 산출물·Out of Scope H3 3개 있음; 하위 에이전트/무표지 진입점 절 부재(D3), Q5 미결(D7) |
| Testability | 0.50 | AC-002 렌더 입력 의존(D6), AC-007 `ok` 의존(D9), AC-008 실행 불가 |
| Traceability | 1.00 | COLLECTED 22, UNCOVERED 0, ORPHAN 0 |

## 2. Evidence
- `git log --oneline b5815ca80..5bfff5682` → f0aa62fc8, 5bfff5682 (SPEC 문서만, 7 files +527).
- python UTF-16 재측정(템플릿 원본): `rules always 13 153561 over40k [('moai/workflow/kanban-dispatch-detail.md', 43138)]`, `CLAUDE 15457 AGENTS.tmpl 20874` — research.md §1 재현.
- `grep -n '^@' templates/CLAUDE.md` → 9:@AGENTS.md, 115:@…/user.yaml, 116:@…/language.yaml, 172:@AGENTS.local.md. user.yaml.tmpl `name: "{{yamlEscape .UserName}}"`, language.yaml.tmpl `{{yamlEscape .ConversationLanguage}}` → 렌더 크기 입력 의존.
- kanban-dispatch.md:86–87 `- **No explanatory prose.**`, `- **Ceiling: the block is at most 10 lines.**` — `[HARD]` 블록 아래 토큰 없는 의무. cross-session-messaging.md:31–37 — `[HARD]` 는 L31 만, 라우팅·보고 의무는 L34–37. codex: L32–37 삭제 변이에서 `binding_rows_before 6 after 6 unchanged True`.
- manager-lead.md L212–L233 deputy 의무가 kanban-dispatch.md 인용; kanban-dispatch 참조 스킬: gtd.md, moai-kanban-foreman/SKILL.md. envkeys: MOAI_KANBAN_LEAD_ADDR/NAME, MOAI_FACTORY_WORKERS/WORKER/ROLE.
- audit_multi: overall fail; build_lag 설치 바이너리 45600e4ee 는 HEAD 조상.

## 3. Baseline-attribution
트리 agent-a6d780d7ba07bc24c, 브랜치 WT-always-loaded-budget, HEAD 5bfff5682; 템플릿 바이트 = b5815ca80(= 로컬 develop 팁). 계수는 이 감사의 직접 실행. 리드의 렌더링 합계 191,386 은 재측정 안 함.

## 4. Defects
D1 LEDGER-LINE-UNIT — spec.md:L52–53,L82; design.md §1 — 원장 단위가 토큰 줄이라 후속 줄·목록 의무가 원장 밖, REQ-013 도 줄만 막아 companion 으로 합법 이탈 가능 — critical — blocking — 「구속 블록」(토큰 줄+연속 줄+직속 목록/표) 단위로 재정의, 후속 줄 삭제 변이 AC 추가.
D2 RUNTIME-CLASSIFICATION-SOURCE — design.md:L27 — 훅이 읽을 분류가 SPEC fixture(미배포, 그대로 배포 시 REQ-019 충돌) — major — blocking — 배포 규칙 안 중립 표지를 런타임 원천으로 명시, production deploy 만으로 역할 core 구성 AC 추가.
D3 ROLE-REACH-GAP — spec.md:L70–71,L74 — deputy 하위 에이전트, foreman 루프, `/moai:todo --auto` 등 무표지 리더급 경로가 변경 후 의무를 못 받음 — critical — blocking — 대상 집합 확장(SubagentStart 주입 또는 일반 분류 유지)과 가드 AC, 판정 규칙에 「역할 밖 진입점이 있으면 일반」 명문화.
D4 LEDGER-LOCATION-MODEL — spec.md:L53,L82; acceptance.md:L30–31 — 역할 행 위치가 role core(가상)인지 paths: 파일인지 미정; 후자면 AC-015 필연 실패 — major — blocking — `always:<path>`/`role-core` 어휘 고정, role-core 행은 훅 생성 함수 출력에서 대조.
D5 LEDGER-BASE-DRIFT — spec.md:L78; plan.md:L35,L40,L52 — M0 기준선 뒤 t1399 흡수·컷 후 착지; base tree 재정의 절차 없음 — major — blocking — 흡수마다 기준선 재생성·별도 커밋·증감 부록.
D6 AC002-RENDER-INPUTS — acceptance.md:L18–20 — 191386 동치가 렌더 입력 의존, 입력 미지정; 「기준 트리 b5815ca80 에서」는 테스트 부재로 실행 불가 — major — blocking — 렌더 입력 고정+기대값 재측정, 기준 트리를 M1 RED 커밋(템플릿 동일 diff 증명)으로.
D7 Q5-FOUNDER-OPEN — decision-index.md:L33–38; spec.md:L80 — FOUNDER Q5 판정 공란인데 REQ-014 예외가 답 전제 — major — blocking — 운영자 RESOLVED 또는 REQ-014 예외·M3 단계를 조건부로.
D8 RED-NOW-ELEMENTS — acceptance.md:L13–36 — AC-020 외 분류 미표기, 다수 RED 칸에 명령·stdout·exit 없음, AC-008 명령 실행 불가 — major — blocking — AC 별 release-blocking/regression-guard 표기, 네 요소 보강 또는 강등.
D9 AC007-VACUOUS-OK — acceptance.md:L19,L23 — `ok` 토큰은 0-테스트에도 성립 — major — optional — `-v` + `--- PASS: TestDeployedAlwaysLoadedCharBudget` 를 초록 조건으로.
D10 REQ010 — spec.md:L73 — `Where` 오용; 구속 줄만 16,182자라 상한이 작으면 불가능 분기 — minor — optional — `When` 으로, 초과 시 fail-visible 동작 명시.
D11 MIRROR-TEST-SCOPE — acceptance.md:L34 — (codex 측정) 대칭 목록에 두 규칙 부재 — minor — optional — 변경·신설 규칙 쌍 등록을 초록 조건에.
D12 ROLE-MARKER-SET — plan.md:L35 — 가드가 기존 표지 목록이면 t1399 신규 표지 미탐 — minor — optional — 런처·훅 공유 레지스트리에서 표지 도출.

## 5. Gaps
- 렌더링 합계 191,386 미재측정.
- awk verb 가 세션 가드에 거부 → 파이썬 포트로 실행(원본과 줄 대조 안 함).
- 감사 후반 세션 Bash·대상 트리 Write 전면 거부(세션 격리 앵커가 develop 으로 표시됨). 그 결과 D11 재측정·Plan Artifact Hash·판정서 대상 경로 기록·커밋 미수행. 가드 우회 시도 안 함.
- glm inconclusive; audit_receipt 없음.

## 6. Residual-risk
- 재작성 행 의미 보존은 사람 판정(운영자 수용 위험) — sync 감사의 행 단위 확인 필수.
- research §2 개산상 M0 가 하한 > 115,000 으로 멈출 가능성 높음(Q6 경로).
- design §4 AGENTS.md 포인터는 t1450(d) 편집과 상호 영향.
- t1453 M4 순서는 타 카드 의무(리드 큐 몫).

## Iteration History
- iter1 (2026-10-03, HEAD 5bfff5682): FAIL 0.69, blocking D1–D8, optional D9–D12.
