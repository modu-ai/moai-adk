# Plan-Audit Iteration 2 — SPEC-WEB-WRITE-SAFETY-001 (D1 델타 재감사)

- **카드**: t517 · Tier M · Class B · cycle_type tdd · **Iteration: 2/2 (Tier M 상한)**
- **감사 대상**: 정정 커밋 `4e94f9607` ("docs(t517): SPEC-WEB-WRITE-SAFETY-001 0.1.1 — resolve plan-audit D1-D4")
- **델타 스코프**: Retry Loop Contract에 따라 iter1 결함(D1-D4) 정정 + 회귀 검사로 한정. 전면 재감사 아님.
- **델타 실측**(`git diff --stat 3b11b3ed9 4e94f9607`, 본 실행): 5파일 전부 `.moai/` 하위 — **코드 파일 무변경**이므로 iter1의 코드 앵커 판독(manager.go·sectionroute.go·schemaform.go·projectconfig.go·handlers.go·server.go·app.go)은 본 트리에서 여전히 유효.

**Verdict: PASS**
**Overall Score: 1.00** (Tier M PASS threshold 0.80)

---

## D1 Resolution — **RESOLVED** (차단 결함 해소)

| 확인 항목 | 증거 |
|-----------|------|
| REQ-WWS-008 신설 (§2.6, shall 어근) | spec.md:108 — "When a repair … is designed or implemented, the repair **shall cite** the two first-measurement conclusions — M-a … and M-b …" + 정지 조항 "**no repair code shall be authored before both measurements are complete**" |
| GEARS 적합 (MP-2) | When형 패턴 + 복합 shall 응답(인용 의무 + 작성 금지). M3의 RED 유닛 테스트 저작(plan.md:94)은 "repair code"가 아니므로 정지 조항과 TDD RED-first 무충돌 확인 |
| AC-WWS-008 재앵커 | acceptance.md:65 제목 "측정 귀속과 측정-선결 [귀속/게이트]", Given이 REQ-WWS-008을 명시 참조, Then에 위반 조항("두 측정 완료 전에 작성된 수리 코드는 본 AC 위반") — 커밋 순서 검증으로 이진 판정 가능 |
| §D 매트릭스 8행 | acceptance.md:84 — "REQ-WWS-008"으로 교체, **순환 참조 "§4 측정 귀속" 제거 확인** |
| §D.3 매핑 행 | acceptance.md:119 — "REQ-WWS-008 (측정-선결) \| AC-WWS-008" 추가 → 8 REQ ↔ 8 AC 양방향 완전 매핑, 고아 0 |
| §5 #2 앵커 | spec.md:159 — "…수리 설계에 선행한다(REQ-WWS-008)" |

## D2-D4 채택 확인 (optional-class, 전부 반영됨)

- **D2** — plan.md M2에 "6종 판별 증거" 불릿 신설(O1이 시사하는 2/6 관측 + "전체 Save() 통과 vs 부분 경로" 교차 판별점, 과소 특정 경고 포함) · M4 exemplar를 5섹션 전부로 정정(manager.go:187/192/197/202/224 앵커, git-strategy는 gated 6번째) · spec.md C3를 5종 일반화로 정정 — 본 감사의 iter1 코드 판독과 정확히 일치.
- **D3** — AC-WWS-002 열거에 `/static/` + "라우트 테이블 app.go의 기타 전 GET 표면" 포괄절 · §D.2에 `/static/`·glmkey reveal edge 추가.
- **D4** — REQ-WWS-001 허용 예외를 닫힌 4라우트 목록(/save, profile 3종)으로 교체 + `/__shutdown__`·glmkey reveal 명시 제외 + **사용자 개시 판별 기준**(콘솔 자신의 렌더·폴링·htmx 자동 트리거는 사용자 개시로 인정하지 않는다). 이 판별 기준은 §1.4 가설 1을 "허용 동작"으로 축소해석하는 경로를 차단해 조기 원인 확정 편향 저항까지 강화했다.

## 회귀 검사 (must-pass 7 + 카드 구속조건 6)

- **MP-1 PASS** — REQ-WWS-001..008 연속(spec.md:84-108, grep 실측), AC-WWS-001..008 연속(acceptance.md:11-65). 결번·중복 0.
- **MP-2 PASS** — REQ-WWS-008 When+shall 적합. REQ-WWS-001 정정 후에도 정준 부정형 유지(부가 문장은 범위 정의이지 비정형 요구가 아님). 판정 계층: 요구 계층.
- **MP-3 PASS** — frontmatter: `version: "0.1.1"`(인용 semver), 12필드 유지, HISTORY 0.1.1 행 추가(spec.md:22), snake_case 별칭 0.
- **MP-4 N/A** — 단일 언어 스코프 unchanged.
- **MP-5 PASS** — related_specs unchanged(diff 문맥행 실측), iter1 측정대로 참조 4건 `status: completed`, 신규 SPEC 참조 없음.
- **MP-6 PASS** — `grep -rn 'syscall'` 재실행 → 0매치.
- **MP-7 PASS** — `grep -rn '\[NEEDS CLARIFICATION:'` 재실행 → 0매치.
- **구속조건**: ① 게이트 무결성 — REQ-WWS-008로 **강화**(요구 계층 정지 조항 + AC 위반 조항), 외부 작성자→blocker 분기 unchanged ② 재현 절차(§C/§D) unchanged ③ t509/t510 경계 침범 0 ④ 부재-가드 규율(AC-001..007 매트릭스) unchanged ⑤ 전체 REQ id ⑥ tdd — 전부 유지.

## Category Scores

| Dimension | Score | 비고 |
|-----------|-------|------|
| Clarity | 1.00 | D4의 닫힌 목록+판별 기준으로 REQ-WWS-001 단일 해석 확보 |
| Completeness | 1.00 | C3/M4/M2 정밀도 향상 |
| Testability | 1.00 | AC-WWS-008 위반 조항까지 이진 판정 가능 |
| Traceability | 1.00 | 8 REQ ↔ 8 AC 양방향, 고아 0, 전체 id |

## Defects Found (잔여)

D5. Out of Scope의 C3 이명 잔존 — spec.md:141 — "C3(llm.yaml 매 `Save()` 재기록)" 괄호 해설이 §1.3에서 5종으로 일반화된 C3의 구(舊) 협의 표기를 유지 — Severity: minor — Class: **optional** — 흡수 논리 자체는 유효하고 §1.3 정의가 정본이므로 행동 영향 0. 재감사 불요; sync-phase 통과 정정 또는 현행 유지 모두 허용.

그 외 결함 없음.

## Regression Check (iteration 2)

iter1 결함 4건 전부:
- D1 — **RESOLVED** (위 표)
- D2 — **ADOPTED** (M2 불릿 + M4 exemplar + C3 일반화)
- D3 — **ADOPTED** (AC-WWS-002 + §D.2)
- D4 — **ADOPTED** (닫힌 목록 + 판별 기준)
미해결 0건.

## Recommendation

PASS. run-phase 진행 가능(Implementation Kickoff Approval 게이트는 본 판정으로 면제되지 않음). run-phase 착수 시 plan.md §C #1의 content-token 앵커 재검증과 M2의 6종 판별 기록을 실행 계획에 반영할 것. D5는 재량.

## Gaps / Residual-risk

**Gaps**: 델타 스코프 밖은 재검증 안 함(계약상) — 단, 델타가 코드 무변경임을 `git diff --stat`로 실측해 iter1 코드 판독의 유효성을 담보 · `moai spec lint` 기계 실행은 iter1과 동일하게 미수행(스키마 SSOT 대조 판정).
**Residual-risk**: 코드 앵커는 트리 `0b1e27877` 기준이며 본 트리 HEAD는 `4e94f9607`(SPEC 파일만 변경) — run-phase에서 코드가 움직이면 재검증 의무가 활성화된다. REQ-WWS-001의 "glmkey reveal은 config를 기록하지 않는다" 서술은 §D.2의 확인 항목으로 run-phase가 재확인한다(단일 판독 미수).
