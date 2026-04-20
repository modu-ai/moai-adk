# SPEC Review Report: SPEC-DOCS-SITE-001 (Iteration 3 / Final)

Auditor: plan-auditor (독립 감사 프로토콜)
Date: 2026-04-17 (Iteration 3)
Input: `.moai/specs/SPEC-DOCS-SITE-001/{spec.md, plan.md, acceptance.md}`
Note: Reasoning context ignored per M1 Context Isolation — 본 감사는 SPEC 파일 3개만을 1차 증거로 채택.

---

## Executive Summary

**최종 판정: FAIL** (Iteration 3)

**Verdict rationale**:
- Iteration 3의 목표는 Iteration 2에서 발견된 **D-026 (자가 검증 LOC 임계값 모순, Critical)** 을 "3개 파일의 LOC 임계값 400 → 250 통일"로 해결하는 것이었다.
- 실증 결과, 지정된 3개 위치(`acceptance.md:L21`, `acceptance.md:L28`, `plan.md:L362`)는 모두 정상 수정되었다.
- 그러나 **plan.md 내 또 다른 G1 기준 서술 지점(`plan.md:L55`)이 수정에서 누락**되어 있다. 이 위치는 여전히 "spec.md 400+ LOC"라는 구 임계값을 명시하고 있어, 동일 문서(plan.md) 내에서 G1 Gate 기준이 상호 모순되는 상태가 발생했다.
- 결과: 현재 `wc -l spec.md = 265 LOC` 기준으로
  - `acceptance.md:L21/L28` 자동 검증: **PASS** (265 ≥ 250)
  - `plan.md:L362` Gate 서술: **PASS** (265 ≥ 250)
  - `plan.md:L55` Phase 1 Gate G1 기준: **FAIL** (265 < 400)
- 즉, 동일 SPEC 내 두 G1 기준이 동일 입력에 대해 상반된 PASS/FAIL 판정을 산출한다. 이는 D-026이 해결된 것이 아니라 **부분 해결 + 새로운 내부 모순으로 변형**된 상태이다.

**최종 판정 기준 적용**:
- Critical 1건 (D-028 신규, 이는 본질적으로 D-026의 불완전 해결)
- High 0건
- Critical ≥ 1 이므로 판정 기준에 따라 FAIL.

---

## D-026 해결 검증 (Iteration 2에서 지정된 3개 위치)

### 위치 1: `acceptance.md:L21` — PASS (수정 완료)

Evidence:
```
21:**Then** 각 파일이 최소 LOC 기준 (spec.md ≥ 250, plan.md ≥ 300, acceptance.md ≥ 250) 을 만족해야 한다.
```

Verdict: 기대한 대로 `spec.md ≥ 250`으로 수정됨.

### 위치 2: `acceptance.md:L28` — PASS (수정 완료)

Evidence:
```
28:test "$(wc -l < .moai/specs/SPEC-DOCS-SITE-001/spec.md)" -ge 250
```

Verdict: bash 테스트의 임계값이 `-ge 250`으로 정확히 교정됨.

### 위치 3: `plan.md:L362` — PASS (수정 완료)

Evidence:
```
362:- 3-file set 존재 및 최소 길이 준수 (spec.md 250+ / plan.md 300+ / acceptance.md 250+ LOC) — D-026 Iteration 3 통일 조정
```

Verdict: `spec.md 250+`로 수정됨. "D-026 Iteration 3 통일 조정" 주석까지 명시됨.

### 누락된 위치: `plan.md:L55` — FAIL (수정 누락)

Evidence (actual current content):
```
55:- 3-file set 존재 및 최소 길이 (spec.md 400+ LOC, plan.md 300+ LOC, acceptance.md 250+ LOC)
```

Verdict: 여전히 **구 임계값 `spec.md 400+ LOC`** 를 명시하고 있음. Iteration 3의 "3개 위치 수정" 지시에서 누락되었다. 동일 문서(plan.md)에 G1 Gate를 설명하는 두 절(§4 Phase 1 / §7 G1 세부기준)이 존재하는데, §7만 갱신되고 §4는 방치되었다.

### 실증 Grep Sweep

전수 확인:
```
grep -n "400\+\|spec\.md.*400" .moai/specs/SPEC-DOCS-SITE-001/*.md

→ .moai/specs/SPEC-DOCS-SITE-001/plan.md:55:- 3-file set 존재 및 최소 길이 (spec.md 400+ LOC, plan.md 300+ LOC, acceptance.md 250+ LOC)
```

구 임계값 `400` 잔재 **정확히 1건** 확인. Iteration 3 수정이 불완전했음을 입증하는 직접 증거.

---

## AC-G1-01 / AC-G1-02 자가 검증 실행 결과

### AC-G1-01 — 3-file 존재 및 최소 LOC (acceptance.md:L25~L30 bash 블록)

실측치:
- `spec.md`: 265 LOC (≥ 250) → PASS
- `plan.md`: 428 LOC (≥ 300) → PASS
- `acceptance.md`: 849 LOC (≥ 250) → PASS

기대 exit code: `0`

AC-G1-01 자동 검증 결과: **PASS** (acceptance.md의 검증 로직만 단독 실행 시).

단, `plan.md:L55`의 서술 기준(`spec.md 400+`)을 같은 판정자가 함께 적용하는 경우 판정이 뒤집힌다. 본 AC는 bash 블록에 bind되어 있으므로 자동 검증은 통과하지만, 문서적 일관성에서는 결함이 있다.

### AC-G1-02 — EARS 패턴 + Exclusions (acceptance.md:L40~L46 bash 블록)

실측치:
- `grep -c "^\*\*REQ-DS-" spec.md` = **36건** (15 ≤ 36 ≤ 36) → PASS
- `grep -q "## 8. 제외사항" spec.md` → line 228에서 매치 → PASS
- `grep -c "^- \[HARD\]" spec.md` = **12건** (≥ 10) → PASS

기대 exit code: `0`

AC-G1-02 자동 검증 결과: **PASS**.

정의된 REQ 번호 (중복 제외):
```
REQ-DS-01 ~ REQ-DS-31 (31건, 연속)
REQ-DS-32a, REQ-DS-32b (32 대체 분할)
REQ-DS-33, REQ-DS-34, REQ-DS-35 (3건)
총 36건
```

MP-1 REQ 번호 일관성: PASS. `REQ-DS-32`는 텍스트 내 cross-reference("REQ-DS-32 → 32a/32b 분할")로만 등장하며 독립 정의는 없음 (HISTORY L17, Exclusions/본문 언급뿐). 이는 명명 규약상 허용.

---

## Must-Pass Results (MP)

- **[PASS] MP-1 REQ 번호 일관성**: 36건, 15~36 범위 내. `spec.md:L128~L216` 연속 배치 확인. 중복 0건.
- **[PASS] MP-2 EARS 포맷 준수**: 36건 전량이 Ubiquitous/Event-driven/State-driven/Optional/Unwanted 중 하나로 명시적 레이블링. 대표 증거 `spec.md:L132` "REQ-DS-03 (Unwanted) — If ...", `spec.md:L150` "REQ-DS-10 (State-driven) — While ...".
- **[PASS] MP-3 YAML frontmatter**: `spec.md:L1-L10` 에 `id`, `version`, `status`, `created`, `author`, `priority`, `issue_number` 필드 존재. (note: `created_at` 대신 `created`, `labels` 부재 — advisory, MP 체계상 허용.)
- **[N/A] MP-4 언어 중립성**: 본 SPEC은 docs-site 단일 도메인(Hugo/Hextra 전환) 범위로 multi-language tooling 이슈 없음.

---

## Category Scores (Rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 band (minor ambiguity in one area) | D-028 때문에 plan.md:L55와 L362 두 위치 간 G1 기준 서술이 모순되어, 감사자가 어느 쪽을 "정식 기준"으로 삼을지 단일 해석이 불가. `plan.md:L55 vs plan.md:L362`. |
| Completeness | 1.00 | 1.0 band | 필수 섹션(HISTORY/WHY/WHAT/REQ/AC/Exclusions) 전원 존재. `spec.md:L14,45,53,76,124,228`. |
| Testability | 0.95 | 0.75~1.0 band | 모든 AC가 자동 검증 bash 블록 또는 명시적 수동 체크리스트를 가짐. 단 AC-G4-09 "사용자 승인" 등은 binary-testable but requires human. |
| Traceability | 1.00 | 1.0 band | `acceptance.md:L811~L849` REQ ↔ AC 매핑 매트릭스가 36건 REQ 전량을 명시적으로 트래킹함. 고아 AC/고아 REQ 0건. |

Overall Score: **0.92** (단, MP 섹션 외 Critical 결함 D-028 존재로 최종 판정 FAIL).

---

## 회귀 검증 (Iteration 2 해결 32건 + Gap 7건 유지 확인)

Iteration 2에서 해결된 주요 defect들의 현재 상태:

| Defect ID | Iter 2 해결 기준 | Iter 3 현재 상태 | Verdict |
|-----------|------------------|------------------|---------|
| D-001 (REQ 상한 36 상향) | AC-G1-02, plan.md §7 G1 모두 "15~36" | acceptance.md:L37 "15~36", plan.md:L56 "15~36개", plan.md:L363 "15~36개" | PASS 유지 |
| D-002 (REQ-DS-32 분할) | 32a/32b 분할 | spec.md:L208, L210 | PASS 유지 |
| D-003 (Gate G1.5 신설) | Phase 2 전용 게이트 | acceptance.md:L66~L135, plan.md:L89~L95 | PASS 유지 |
| D-004 (DNS 조항 이관) | REQ-DS-32 본문에서 제거 → AC-PRE-01 | spec.md:L208 "DNS is **not** modified", acceptance.md:L456~L466 AC-PRE-01 | PASS 유지 |
| D-005 (Mermaid tolerance 제거) | `-eq 569` | acceptance.md:L229 `test "$COUNT" -eq 569` | PASS 유지 |
| D-006 (Callout tolerance 제거) | `-eq 735` | acceptance.md:L200 `test "$TOTAL" -eq 735` | PASS 유지 |
| D-007 (Edge runtime 명시) | `runtime: 'edge'` + platform check | spec.md:L158 REQ-DS-13, plan.md:L176-L177 Phase 5 | PASS 유지 |
| D-008 (G1 범위 통일) | "15~36" 양쪽 일치 | plan.md:L56, L363 acceptance.md:L37 동일 | PASS 유지 |
| D-009 (누락 REQ AC 추가) | AC-G2-08, AC-G3-08, AC-G2-09 | acceptance.md:L246, L418, L260 | PASS 유지 |
| D-010 (릴리스 태그 기준) | previous release tag commit 명시 | spec.md:L168 REQ-DS-17 | PASS 유지 |
| D-011 (테스트 이름 교정) | TestMinorRelease / TestPatchRelease / TestMajorRelease | acceptance.md:L360~L363 | PASS 유지 |
| D-012 (이모지 예외 조항) | Non-goals 예외 허용 | spec.md:L69 | PASS 유지 |
| D-013 (Nextra baseline) | -5 이내 조항 | acceptance.md:L614 | PASS 유지 |
| D-014 (48h 롤백) | REQ-DS-35 + AC-MON-03 | spec.md:L216, acceptance.md:L710~L728 | PASS 유지 |
| D-015 (Hugo module 잠금) | git submodule 금지 | spec.md:L225, plan.md:L76 | PASS 유지 |
| D-016 (placeholder 구체화) | _index.md + draft: true + redirect | spec.md:L162 REQ-DS-15 | PASS 유지 |
| D-017 (Vercel build timeout 조사) | Phase 2 완료 단계 | plan.md:L80 | PASS 유지 |
| D-018 (Hextra 한계 소섹션) | §4.1 신설 | spec.md:L102~L108 | PASS 유지 |
| D-019 (SEO 영향 조사 AC) | AC-PRE-02 | acceptance.md:L468~L478 | PASS 유지 |
| D-020 (G2 표기 수정) | "Phase 3/4 → Phase 5" | plan.md:L375 | PASS 유지 |
| D-021 (REQ-DS-34 주어) | Phase 8 automation 명확화 | spec.md:L214 | PASS 유지 |
| D-022 (moai-docs config 이관 제외) | Exclusions 명시 | spec.md:L240 | PASS 유지 |
| D-023 (배너 구현 구체화) | plan.md Phase 5 구체화 | plan.md:L186~L190 | PASS 유지 |
| D-024 (정규식 강화) | `grep -Eq "^### *(§?17\.[1-6])"` | acceptance.md:L380~L385 | PASS 유지 |
| D-025 ($PREVIEW_URL 지침) | AC-G4-03 지침 추가 | acceptance.md:L539~L544 | PASS 유지 |
| Gap 1~7 (Vercel 무중단 근거 등) | AC-PRE-03, AC-G2-10/11, AC-G4-11, AC-G4-05 og.png, AC-G3-09 | acceptance.md:L480, L276, L288, L660, L571, L437 | PASS 유지 |
| D-026 (자가 검증 모순, Iter 2 Critical) | 3개 위치 250 통일 | **부분 해결. plan.md:L55 누락** | **FAIL (D-028로 변형)** |

**Iter 2 32건 중 31건은 회귀 없음을 확인**. D-026만 불완전 해결.

---

## 신규 결함 (Iteration 3)

### D-028 — plan.md:L55 Phase 1 Gate G1 기준이 구 LOC 임계값(`spec.md 400+`)을 유지 — Severity: **Critical**

**증거** (plan.md 현재 내용, 직접 인용):
```
55:- 3-file set 존재 및 최소 길이 (spec.md 400+ LOC, plan.md 300+ LOC, acceptance.md 250+ LOC)
```

**대비 증거** (동일 plan.md의 §7 G1 서술):
```
362:- 3-file set 존재 및 최소 길이 준수 (spec.md 250+ / plan.md 300+ / acceptance.md 250+ LOC) — D-026 Iteration 3 통일 조정
```

**영향 분석**:
- 현재 `spec.md` 실측 LOC = **265**.
- `plan.md:L55` 기준으로 판정하면 `265 < 400` → G1 Gate **실패**.
- `plan.md:L362` 기준으로 판정하면 `265 ≥ 250` → G1 Gate **성공**.
- 즉, 동일 문서 내 동일 Gate에 대해 상반된 판정을 산출하는 **내부 자가 모순** 상태.
- Iter 2 D-026은 "acceptance.md가 자기 자신에 대해 FAIL을 선언"하는 Critical 결함이었음. Iter 3 수정은 acceptance.md 모순은 제거했으나, **plan.md 내부 모순으로 이관**시킨 결과가 되었다.

**근본 원인**:
- Iteration 3 수정 지시가 "3개 위치(acceptance.md:L21, L28, plan.md:L362)"로 한정되어 전수 grep sweep이 누락됨.
- plan.md는 Phase 1 상세 설명 (§4) 과 G1 Gate 세부 기준 (§7) 두 곳에서 같은 기준을 서술하는 구조인데, 한 곳만 갱신됨.

**수정 권고**:
```
plan.md:L55
- 3-file set 존재 및 최소 길이 (spec.md 400+ LOC, plan.md 300+ LOC, acceptance.md 250+ LOC)
+ 3-file set 존재 및 최소 길이 (spec.md 250+ LOC, plan.md 300+ LOC, acceptance.md 250+ LOC)
```

이 한 줄 교체로 해결 가능.

---

## Chain-of-Verification Pass (두 번째 검증 사이클)

자기 비판 질문에 대한 답변:

1. **"REQ-DS 전수를 읽었는가, 첫 몇 건만 보고 넘어갔는가?"**
   - 36건 모두 grep output으로 확인, 각 EARS 레이블 패턴 (Ubiquitous/Event-driven/State-driven/Optional/Unwanted) 일일 확인. PASS.

2. **"REQ 번호 연속성을 샘플링이 아닌 전수로 확인했는가?"**
   - `grep -oE "REQ-DS-[0-9]+[a-z]?" spec.md | sort -u` 로 전수 열거. 번호 gap 0건, duplicate 0건 (bare REQ-DS-32는 cross-reference로만 등장, 허용). PASS.

3. **"D-026 수정 지시의 3개 위치 외에 추가로 점검했는가?"**
   - 초기 점검은 지시된 3개 위치(acceptance.md:L21/L28, plan.md:L362)만 확인. 이후 전수 grep (`grep -n "400\+\|spec\.md.*400"`) 을 실행하여 **plan.md:L55의 누락을 발견**. 이는 초기 점검에서 놓친 결함.
   - **이 Chain-of-Verification pass가 없었다면 D-028을 보고하지 못했을 것**. 감사 프로토콜 M6의 핵심 가치가 실증된 사례.

4. **"Exclusions 섹션의 항목 수와 구체성을 확인했는가?"**
   - 12건 `[HARD]` 항목 모두 구체적 패턴 명시. D-022 포함. PASS.

5. **"문서 간 cross-reference 정합성 확인했는가?"**
   - D1~D4는 spec.md:L119~L122와 plan.md:L19~L22에서 중복 명시. AC-G1-03 수동 검증 대상으로 등록. PASS.
   - REQ ↔ AC 매핑 매트릭스 (acceptance.md:L810~L849) 전수 확인, 고아 0건. PASS.

6. **"요구사항 간 모순 없는지 확인했는가?"**
   - REQ-DS-32a(재바인딩, G4 이전) vs REQ-DS-32b(프로모션, G4 이후) 명확히 시간 분할됨. REQ-DS-34(Archive, 48h 후) vs REQ-DS-35(롤백 조건) 상호 배타적. REQ 내부 논리 모순 0건. PASS.
   - 단, **D-028 (plan.md:L55 vs L362)은 REQ 간 모순이 아니라 Gate 기준 서술의 모순**으로 분류.

추가 검증된 교차 정합성:
- "15~36" REQ 범위: acceptance.md:L37 + plan.md:L56 + plan.md:L363 3곳 모두 일치.
- spec.md HISTORY L24 "15~35"는 과거 수정 이력 기술이므로 현재 규정이 아니며 허용.
- Callout 735 / Mermaid 569 절대값: acceptance.md (L200, L229), plan.md (L379, L380), spec.md REQ-DS-05 / REQ-DS-09 전체 일치.

---

## Defects Found (Iteration 3 기준)

| ID | 위치 | 설명 | 심각도 |
|----|------|------|--------|
| D-028 | plan.md:L55 | Phase 1 Gate G1 기준 서술이 구 LOC 임계값 `spec.md 400+ LOC`을 유지. plan.md:L362(`250+`) 와 내부 모순. D-026의 불완전 해결이 이관된 결과. | Critical |

그 외 신규 결함: 없음.

Iteration 2 advisory D-027 (YAML `created_at` / `labels` 부재): 본 감사에서 MP-3 기준으로 재평가, `created`/`author`/`priority`/`version`/`status`/`id`/`issue_number` 7개 필드 존재하므로 MP-3 PASS. advisory 수준 유지, blocking 아님.

---

## Recommendation

**Next action**: Phase 2 Scaffold 착수 권고 **X (No-Go)**.

**사유**:
1. Critical 결함 1건 (D-028) 이 확인되었으며, 이는 Iteration 3의 목적이었던 D-026 해결이 불완전함을 의미한다.
2. plan.md 내부 G1 Gate 기준이 자기 자신과 모순되는 상태에서 Phase 2에 진입하면, G1 Gate 통과 여부 자체가 모호해진다.
3. Iteration 4로 진입하여 1줄 수정만으로 해결 가능 (다음 문단 "Fix Instructions" 참조).

**Fix Instructions for Iteration 4** (manager-spec 위임 지침):

1. **단일 편집**: `.moai/specs/SPEC-DOCS-SITE-001/plan.md` 의 L55를 다음과 같이 교체.
   ```
   변경 전 (L55):
   - 3-file set 존재 및 최소 길이 (spec.md 400+ LOC, plan.md 300+ LOC, acceptance.md 250+ LOC)

   변경 후:
   - 3-file set 존재 및 최소 길이 (spec.md 250+ LOC, plan.md 300+ LOC, acceptance.md 250+ LOC) — D-026 통일 조정 (Iteration 3/4)
   ```

2. **검증**: 수정 후 `grep -n "400\+\|spec\.md.*400" .moai/specs/SPEC-DOCS-SITE-001/*.md` 가 0건이어야 한다.

3. **재감사 제외 권고**: 이 변경은 1줄에 한정되며 다른 섹션에 영향이 없다. 단, manager-spec이 변경 후 `plan.md:L55`와 `plan.md:L362`가 문장 구조·임계값 3종 동일함을 eyeball 비교하여 확인할 것.

**Iteration 4 재감사 예상 소요**: plan-auditor 단일 grep 검증으로 PASS 판정 가능 (Chain-of-Verification 재실행 포함 전체 실행 시간 ~1분 수준).

**Phase 2 진입 조건**:
- Iteration 4 수정 완료
- `grep -n "400\+\|spec\.md.*400" .moai/specs/SPEC-DOCS-SITE-001/*.md` 결과 0건
- 본 감사 기준 D-028 RESOLVED 확인 후 Phase 2 착수 가능.

---

## Regression Check Summary

- 전체 감사 항목 32건 중 **31건 유지**, **1건(D-026) 불완전 해결 → D-028로 변형**.
- Iter 1~3 스태그네이션 검출: D-026/D-028은 "LOC 임계값 불일치" 본질이 Iter 2 → Iter 3에 걸쳐 잔존. 그러나 Iter 2 형태(acceptance.md 자가 모순)와 Iter 3 형태(plan.md 내부 모순)가 서로 다른 위치이므로 "전혀 다른 결함"은 아니지만 "동일 defect의 이관"에 해당. **manager-spec의 수정 sweep 범위 설정이 반복적으로 불충분함을 시사**.

권고: Iteration 4에서는 manager-spec에게 "임계값 상수 변경 시 해당 상수가 쓰인 모든 위치를 grep으로 전수 탐색 후 일괄 교체" 가이드를 명시 전달할 것.

---

## 최종 판정

**FAIL** — Iteration 3 종료. Iteration 4 진입 필수.

- Critical: 1 (D-028)
- High: 0
- Medium: 0
- Phase 2 Scaffold 착수: **불가 (No-Go)**

---

Report produced by: plan-auditor (독립 감사, M1~M6 편향 방지 프로토콜 적용)
Output path: `/Users/goos/MoAI/moai-adk-go/.moai/plans/DOCS-SITE/g1-audit-report-iter3.md`
