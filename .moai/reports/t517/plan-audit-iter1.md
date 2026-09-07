# Plan-Audit Iteration 1 — SPEC-WEB-WRITE-SAFETY-001

- **카드**: t517 · Tier M · Class B · cycle_type tdd
- **감사 대상**: draft 커밋 `3b11b3ed9` (`.moai/specs/SPEC-WEB-WRITE-SAFETY-001/{spec,plan,acceptance}.md` + progress.md)
- **코드 앵커 베이스**: `0b1e27877` — 본 워크트리(`.claude/worktrees/t517`, `WT-web-write-safety`)의 베이스와 동일하며 draft 커밋은 SPEC 파일만 추가하므로 코드는 무변경. 모든 코드 인용은 **본 실행, 본 트리**에서 직접 판독으로 검증했다.
- **M1 Context Isolation**: SPEC 저자의 추론 컨텍스트는 수령하지 않았다. 리드가 전달한 결함 서술·구속 조건은 **감사 기준**으로만 사용했고, 아티팩트 자체의 문장만 판정 근거로 인용한다.
- **Iteration**: 1/2 (Tier M — `harness.plan_audit_tier_ceilings` M=2)

**Verdict: FAIL** — 차단 결함 1건 (D1).
**Overall Score: 0.88** — Tier M PASS threshold(0.80)는 초과하나, 차단 결함이 수리·재판정 전까지 착지를 막는다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — REQ-WWS-001..007이 spec.md:83-103에 연속 배치, 결번·중복 0. AC-WWS-001..008도 연속(acceptance.md:11-69). 전 REQ id가 전체 표기(`REQ-WWS-00X`)이며 `§D` 매트릭스의 `REQ-WWS-001/002`(acceptance.md:77)는 두 전체 id의 병기로 약축이 아니다 — 추적 도구 토큰화에 안전하다.
- **[PASS] MP-2 GEARS 형식 (요구 계층 기준)** — 7건 전부 패턴 적합. REQ-WWS-002/003/006은 When형, REQ-WWS-005/007은 Ubiquitous형. REQ-WWS-001(spec.md:83, "shall not write … in the absence of …")과 REQ-WWS-004(spec.md:91, "No web console code path shall bypass …")는 각각 정준 부정형(Unwanted)에 범위 한정자·주어 측 부정이 결합한 형태로, 의미상 "The <subject> shall not <action>"과 동치라 적합으로 판정했다(M6 과잉기계 판정 회피). **판정 계층: REQ-XXX 요구 계층만** — AC의 Given-When-Then은 검증 계층 정본 형식이므로 본 기준에서 채점하지 않았다.
- **[PASS] MP-3 YAML Frontmatter** — 12개 정본 필드 전부 존재·타입 적합(spec.md:2-15). `created/updated: 2026-09-07` ISO, `version: "0.1.0"` 인용 semver, `status: draft`(열거형 유효), `priority: P1`, `phase: "v3.2.0"`(릴리스 라벨 — 금지 단계명 아님), `tags` CSV 문자열. snake_case 별칭(`created_at` 등) 0건. `tier: M`, `related_specs`는 선택/비규제 필드.
- **[N/A] MP-4 언어 중립성** — 단일 언어 스코프 SPEC(module: internal/web, internal/config)이므로 N/A 자동 통과. Go 파일 인용은 다중 언어 도구 나열이 아니라 모듈 앵커다.
- **[PASS] MP-5 D7 교차-SPEC 정합** — 본 실행 측정: 참조 4건 전부 존재하며 `status: completed`(SPEC-GITSTRATEGY-SAVE-ISOLATION-001, SPEC-WEB-CONSOLE-011, SPEC-WEB-CONSOLE-010, SPEC-FEEDBACK-AUTO-SUBMIT-001 — 각 `grep '^status:'` 출력 `status: completed`, 트리 `3b11b3ed9`). retired/superseded/archived 참조 0 → 정합 조항 불필요, BLOCKING 없음.
- **[PASS] MP-6 D8 크로스플랫폼** — `grep -rn 'syscall' .moai/specs/SPEC-WEB-WRITE-SAFETY-001/` → 매치 0 (rc=1). D8-4 자동 PASS.
- **[PASS] MP-7 clarification gate** — 콜론형 마커 `grep -rn '\[NEEDS CLARIFICATION:'` → 매치 0(rc=1). plan.md:129의 `## [NEEDS CLARIFICATION] — 없음`은 미해결 마커가 아니라 **부정 선언 본문**(유일 분기 = 외부 동시 작성자 → blocker report, 소관 명시)이며, 마커 컨벤션 자체가 콜론형(`[NEEDS CLARIFICATION: <topic>]`, SKILL.md:178)을 정의하므로 미해결 마커 아님. Tier M이라 research.md 부재 — MP-4 선례대로 N/A 처리.

---

## Category Scores

| Dimension | Score | 근거 |
|-----------|-------|------|
| Clarity | 0.75 | REQ-WWS-001(spec.md:83)의 허용 예외가 "POST /save 제출, /profile/* 제출 **등의** 사용자 개시 쓰기"로 열린 목록 — 1개 요구의 경미한 모호성(0.75 밴드). 나머지 6 REQ는 단일 해석. |
| Completeness | 1.00 | HISTORY(spec.md:18-22)·WHY(§1)·WHAT(§2-3)·REQ(§2)·AC(§5→acceptance.md)·Out of Scope 4개 H3+불릿(spec.md:122-136) 전부 존재. Edge Cases(acceptance.md:99-105)와 이진 최신성·스냅샷·귀속 제약(spec.md:140-145)까지 충실히 명시. |
| Testability | 1.00 | AC 8건 전부 Given-When-Then + 판정 방법 명시(acceptance.md:11-69). weasel word 0. RED-now 셀 4요소 구조(§D.1, 88-97)와 "올바른 이유의 RED" 문구(97)까지 규정. |
| Traceability | 0.75 | REQ→AC 방향은 §D.3(107-119)에서 전건 커버. 그러나 **AC-WWS-008이 REQ 없음**(§D 매트릭스 84행: 대응 REQ 칸이 "§4 측정 귀속" — 제약 절을 가리킬 뿐 요구 계층 앵커 부재) → 고아 AC 1건(8건 중). |

**Overall: 0.88** (4차원 평균). REQ 7/16 · AC 8/16 — Tier M 예산 내.

---

## Defects Found

**D1. 고아 AC — 측정-선결 의무의 요구 계층 앵커 부재** — acceptance.md:65-69, :84 (§D 매트릭스 8행), spec.md §2 — Severity: **major** — Class: **blocking**
AC-WWS-008(M-a/M-b 측정 귀속)의 "대응 REQ" 칸이 `§4 측정 귀속`(제약 절)을 가리킬 뿐, 유효한 REQ-WWS-XXX를 참조하지 않는다. 이 AC가 검증하는 "첫 측정이 수리에 선행하고 수리 설계가 그 결론을 인용한다"는 것은 Class B 카드의 **핵심 게이트 의무**인데, spec.md §2 요구 계층에는 그 요구가 없다. spec.md:154(§5 #2)가 산문으로 진술하지만 §5는 "acceptance.md의 AC 매트릭스가 유일한 판정 기준"(spec.md:151)이라며 다시 acceptance.md로 되돌리고, acceptance.md의 그 행은 다시 §4를 가리킨다 — **순환 참조로 요구 계층에 닿지 않는다**. plan.md M2/M3의 [HARD] 문구(plan.md:88, :95)는 plan 계층이라 run-phase가 acceptance.md 중심으로 작업할 때 이 의무가 요구 계층에 묶이지 않는다.
**필요한 수정**: spec.md §2에 REQ-WWS-008 추가(예: "When the repair (M4) is designed, the web console repair shall cite the M-a/M-b measurement conclusions, each attributable to command + observed output + tree SHA.") + acceptance.md §D 8행과 §D.3에 매핑 갱신.

**D2. M4 범위 exemplar의 과소 열거 + M-a 판별 증거 미채택** — plan.md:100, spec.md:53 — Severity: minor — Class: optional
M4가 "값 불변 섹션 재기록 제거(REQ-WWS-003). C3(llm.yaml 무조건) 포함"으로 llm.yaml만 exemplar로 명명하지만, 본 실행 판독에서 `Save()`는 **5개 섹션을 무조건 재기록**한다: user.yaml(manager.go:187), language.yaml(:192), quality.yaml(:197), git-convention.yaml(:202), llm.yaml(:224). REQ/AC 계층은 일반 원칙이므로 오답은 아니나, exemplar 편향은 run-phase 드리프트 위험이다. 아울러 관측(O1)에서 "6종 중 git-strategy·feedback 2종만 내용 변경"이라는 사실 자체가 강한 판별 증거다 — Save() 전체가 통과했다면 무조건 5종도 기록됐어야 하므로(내용 동일 round-trip 가능성은 M-a가 배제/확인해야 함), M2가 이 교차 증거를 기록·판정 항목으로 삼아야 귀속이 정밀해진다.
**권고 수정**: plan.md M4 불릿에 5개 무조건 섹션 열거 1줄 추가 + M2에 "6종 중 무엇이 내용 변경됐는지" 판별 기록 항목 추가. (본 감사는 round-trip 동일성을 검증하지 않았다 — run-phase 측정 사항.)

**D3. "GET 라우트 전부" 열거 누락** — acceptance.md:22 (AC-WWS-002), acceptance.md §D.2 — Severity: minor — Class: optional
AC-WWS-002가 "GET 라우트 전부"라 쓰고 7개를 열거했으나, 라우트 테이블(app.go:157-196, 본 실행 판독)에는 GET 응답 표면 `/static/`(:196)이 하나 더 있다. §D.2 edge 목록에도 glmKeyReveal(app.go:187)이 없다. 두 표면 모두 config 쓰기와 무관해 실질 영향은 0에 수렴하나, "전부" 서술과 열거의 불일치는 스윕 완결성 판독을 흐린다.
**권고 수정**: 열거에 추가하거나 "및 기타 전 GET 표면" 절을 붙인다.

**D4. REQ-WWS-001 허용 예외의 열린 목록** — spec.md:83 — Severity: minor — Class: optional
"등의 사용자 개시 쓰기"가 허용 집합을 열어두어, 구현자가 임의 쓰기를 "사용자 개시"로 분류할 여지를 남긴다. AC-WWS-001/002가 무저장 사례를 붙잡으므로 즉각 위험은 없다.
**권고 수정**: 허용 집합을 C8의 POST 라우트 중 config 기록 라우트로 닫힌 목록화.

---

## Regression Check (Iteration 2+ 전용)

N/A — 본 보고는 iteration 1이다.

---

## Recommendation (manager-spec 수정 지시)

1. **[필수 — D1]** spec.md §2에 REQ-WWS-008을 추가하고 acceptance.md §D 8행·§D.3 매핑을 갱신한다. 이것 하나가 본 FAIL의 유일한 원인이다.
2. [권장 — D2] plan.md M4에 무조건 재기록 5섹션 열거 + M2에 6종 판별 증거 기록 항목을 추가한다.
3. [권장 — D3] AC-WWS-002 열거 보완. 4. [권장 — D4] REQ-WWS-001 허용 집합 닫기.
5. 수정 후 재감사는 **D1 델타 스코프**로 수행한다(Retry Loop Contract — 전면 재감사 아님). Tier M 상한 2회 내에 resolution 가능한 소규모 수정이다.

---

## 감사자가 관측한 것 / 관측하지 않은 것

**관측(본 실행, 트리 `3b11b3ed9`/`0b1e27877`)**: 4개 아티팩트 전문 판독 · C1-C3/C5-C8 코드 앵커 직접 판독(manager.go:171-233, sectionroute.go:80-109, schemaform.go:310-359, projectconfig.go:145-175, server.go:90-144, app.go:152-201) · C4 쓰기 호출 9개 시퀀스 grep 확인(handlers.go:465,478,482,494,503,513,523,533,546) · internal/cli/web.go 쓰기 호출 무매치(C7) · D7 4건 status · D8/MP-7 grep.

**미관측(Gaps)**: `moai web` 실물 재현 수행 안 함(M1은 run-phase 소관) · 리드 전달 관측 O1/O2(mtime, primary llm.yaml) 미재측정 — SPEC 자체가 이를 전달 근거로 정확히 표기했는지까지만 감사 · handlers.go:483-489 에러 배너(롤백 없음 문서화) 본문 미판독 — C4는 호출 시퀀스만 검증 · 무조건 5섹션의 byte round-trip 동일성 미측정 · `moai spec lint` 기계 실행 안 함(frontmatter는 SSOT 스키마 문서 대조 육안 판정).

**잔여 위험**: 코드 앵커는 트리 `0b1e27877` 고정값이며 run-phase 착수 시 content-token 재검증이 필요하다(SPEC이 이미 §C #1로 의무화). D2의 미해결 판별(무조건 섹션 round-trip 동일 여부)이 M-a 귀속의 정밀도를 좌우한다 — 채택되지 않으면 M2 결론이 과소 특정될 수 있다.
