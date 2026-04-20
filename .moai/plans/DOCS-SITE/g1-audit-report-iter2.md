# SPEC-DOCS-SITE-001 — G1 Plan Audit Report (Iteration 2)

- **Auditor**: plan-auditor (independent, adversarial mode, M1~M6 protocols active)
- **Audit Date**: 2026-04-20
- **Audit Inputs**: spec.md (265 LOC), plan.md (428 LOC), acceptance.md (849 LOC)
- **Cross-reference**: `.moai/plans/DOCS-SITE/g1-audit-report.md` (Iteration 1 FAIL), `.moai/plans/DOCS-SITE/spec-iteration2-changelog.md`
- **Bias Protocol**: M1 Context Isolation enforced — author's reasoning in changelog and prompt hints NOT consumed. Evidence-only verification.

---

## 1. Executive Summary

**Verdict: FAIL**

**한 줄 근거**: 32개 사전 결함(D-001~D-025 25건 + Gap 1~7 7건)은 100% 해결되었으나, Chain-of-Verification pass 도중 **신규 Critical 결함 D-026** 이 발견됨 — spec.md 가 AC-G1-01 의 자체 자동 검증 (`test "$(wc -l < spec.md)" -ge 400`) 을 통과하지 못함 (실제 265 LOC < 필수 400 LOC). Iteration 1 의 D-001 과 동일한 성격의 "SPEC이 자기 자신의 G1 자동 검증을 통과할 수 없다" 자가 모순이므로 CRITICAL 로 분류.

**판정 기준 적용**:

| 조건 | 결과 |
|------|------|
| Critical 0 + High ≤ 2 → PASS | 해당 없음 |
| Critical 0 + High 3~5 → CONDITIONAL PASS | 해당 없음 |
| Critical ≥ 1 OR High ≥ 6 → FAIL | **트리거됨 (Critical 1)** |

**Iteration 2 해결률**: 32 / 32 (100%) — 사전 결함·Gap 전량 해결됨.
**신규 결함 수**: Critical 1 (D-026), Medium 1 (D-027, advisory).

Iteration 1 대비 평가: Iteration 2 는 제출된 수정안에 한해 매우 높은 품질을 달성했다 (Critical 2→0, High 6→0, Medium 11→0, Low 6→0). 다만 감사 범위 밖에서 새로운 자가 모순 1건이 발견되어 한 번 더의 미세 조정이 필요하다.

---

## 2. 이전 결함 해결 검증 표 (D-001 ~ D-025 + Gap 1~7)

| ID | Severity | 권장 수정안 요약 (Iter1) | Iter2 실제 반영 (line 증거) | 판정 |
|----|----------|---------------------------|--------------------------------|------|
| D-001 | Critical | AC-G1-02 상한 30→35 + plan.md 15~35 통일 | acceptance.md:L41 `$1 >= 15 && $1 <= 36`, plan.md:L363 "15~36" ; spec.md REQ 실측 36 건 — awk 자기검증 PASS 확인 완료 | **RESOLVED** |
| D-002 | Critical | REQ-DS-32를 32a(재바인딩, G4 이전)/32b(프로모션, G4 이후) 분할 | spec.md:L208 REQ-DS-32a "When Phase 7 enters the cutover preparation step...", spec.md:L210 REQ-DS-32b "When acceptance.md § G4 checklist passes..." — 이벤트 트리거가 Phase 7 step 2(재바인딩, plan.md:L266) 및 step 5(프로모션, plan.md:L281)와 일치 | **RESOLVED** |
| D-003 | High | Phase 2 전용 Gate G1.5 신설 (scaffold 기준 5종) | plan.md:L367 "G1.5 — Scaffold Readiness Gate", 5개 기준 plan.md:L369~373 + acceptance.md AC-G1.5-01 ~ AC-G1.5-05 (acceptance.md:L68, L80, L96, L115, L123) | **RESOLVED** |
| D-004 | High | DNS TTL 조항 제거 + Phase 7 사전 조사 AC 신설 | spec.md HISTORY:L20 "DNS TTL 조항을 REQ-DS-32 본문에서 제거", REQ-DS-32a (spec.md:L208) "DNS is **not** modified during this step", Exclusions:L241 [HARD] DNS 사전 승인 금지, AC-PRE-01 신설 (acceptance.md:L456~466) | **RESOLVED** |
| D-005 | High | AC-G2-06 를 `test "$COUNT" -eq 569` 로 강화 | acceptance.md:L229 `test "$COUNT" -eq 569` (tolerance 완전 제거) | **RESOLVED** |
| D-006 | High | AC-G2-04 를 `test "$TOTAL" -eq 735` 로 강화 | acceptance.md:L200 `test "$TOTAL" -eq 735` (tolerance 완전 제거) | **RESOLVED** |
| D-007 | High | REQ-DS-13 runtime:'edge' 명시 + Phase 5 제약 검토 + 헤더 검증 AC | spec.md:L158 REQ-DS-13 본문에 `export const config = { runtime: 'edge' };` 명시; plan.md:L176 "Platform constraint 사전 검토"; AC-G3-01 (acceptance.md:L325) runtime grep + AC-G4-03 (acceptance.md:L554) 헤더 검증 | **RESOLVED** |
| D-008 | High | plan.md G1 범위를 "15~35"로 통일 | plan.md:L363 "15~36" (acceptance.md:L41 `<= 36`와 일치, D-001과 연계 수치 통일) | **RESOLVED** (범위 값이 35→36으로 조정된 것은 REQ-DS-35 신규 반영, D-001 상한 재산정과 정합) |
| D-009 | Medium | 누락 REQ(19/20/29) 자동 검증 AC 추가 | AC-G2-08 (acceptance.md:L246 REQ-DS-19 배너 partial), AC-G2-09 (L260 REQ-DS-29 금지 패턴 CI), AC-G3-08 (L418 REQ-DS-20 3종 partial) 모두 신설 + Appendix C 매트릭스 (acceptance.md:L810~849) 추가 | **RESOLVED** |
| D-010 | Medium | REQ-DS-17 에 스냅샷 대상 커밋 (previous release tag) 명시 | spec.md:L168 "against the commit tagged as the previous release (e.g., the commit tagged v2.12.X), not HEAD of main" | **RESOLVED** |
| D-011 | Medium | AC-G3-04 테스트 이름 교정 + TestMajorRelease 추가 | acceptance.md:L360~362 TestMinorRelease (v2.12→v2.13), TestPatchRelease (v2.12.1→v2.12.2), TestMajorRelease (v2.12.0→v3.0.0) 3건 정상 배치 | **RESOLVED** |
| D-012 | Medium | Non-goals 에 이모지 예외 조항 추가 | spec.md:L69 "예외 (coding-standards.md 필수 적용): ... 국기 이모지 → 'KO / EN / JA / ZH' 텍스트 라벨로 전환..." + spec.md:L237 Exclusions 디자인 리디자인 금지 항목에 괄호 예외 명시 | **RESOLVED** |
| D-013 | Medium | AC-G4-07 에 "Nextra baseline 대비 Performance -5 이내" 추가 | acceptance.md:L617 표에 "Nextra baseline 대비 (Desktop) baseline - 5 이내" 컬럼, plan.md:L264 Phase 7 step 1 baseline 측정 단계 | **RESOLVED** |
| D-014 | Medium | REQ-DS-35 (Unwanted) + AC-MON-03 신설 | spec.md:L216 REQ-DS-35 "If any AC-MON-01 threshold is violated ..., then the operator shall immediately revert..." — Unwanted EARS 패턴 준수 + acceptance.md:L710 AC-MON-03 4단계 롤백 절차 | **RESOLVED** |
| D-015 | Medium | Hugo module system 잠금 (submodule 금지) | spec.md:L225 제약 "Hextra 통합 방식 (D-015 잠금): ... Hugo module system(go.mod import) 으로만 통합 ... git submodule 방식은 사용하지 않는다"; plan.md:L76 step 3 재확인, plan.md:L407 Technical Approach 재확인 | **RESOLVED** |
| D-016 | Medium | REQ-DS-15 "empty skeleton" 내용 구체화 | spec.md:L162 "containing a single `_index.md` with `draft: true` frontmatter plus a redirect alias to the equivalent ko locale path" (placeholder 내용 명시), plan.md:L114 `aliases: ["/ko/<section>/"]` 구체화 | **RESOLVED** |
| D-017 | Medium | Vercel 빌드 timeout 선제 조사 지침 추가 | plan.md:L80 "빌드 시간 베이스라인 측정 ... Vercel 기본 빌드 timeout (10분)과의 여유 계산" + AC-G1.5-05 (acceptance.md:L129) + AC-G2-11 (L288) | **RESOLVED** |
| D-018 | Medium | §4 용어집 뒤 "Hextra 한계" 소섹션 신설 | spec.md:L102 "### 4.1 Hextra 한계 (Acknowledged Limitations)" — versioning 내장 없음, React/JSX 미지원, 서버사이드 Mermaid 프리렌더 미제공 3건 명시 | **RESOLVED** |
| D-019 | Medium | Phase 4 pre-production SEO 영향 조사 AC 추가 | plan.md:L151 step 8 "SEO 영향 사전 조사"; AC-PRE-02 (acceptance.md:L468) neutral rendering | **RESOLVED** |
| D-020 | Low | plan.md §7 G2 표기 "(Phase 3/4 → Phase 5)" 수정 | plan.md:L375 "G2 — Hugo Build Gate (Phase 3/4 → Phase 5, D-020 표기 수정)" | **RESOLVED** |
| D-021 | Low | REQ-DS-34 주어 명확화 | spec.md:L214 "the system (Phase 8 automation, executed by manager-git and expert-devops) shall archive..." — 수행 주체 및 Phase 8 귀속 명시 | **RESOLVED** |
| D-022 | Low | moai-docs gate/memo/observability yaml 3개 이관 제외 | spec.md:L240 "[HARD] moai-docs 전용 config 이관 제외 (D-022) — ... gate.yaml, memo.yaml, observability.yaml 3개 파일은 docs-site 런타임 운영과 무관하므로 본 SPEC의 이관 대상에서 완전히 제외한다" | **RESOLVED** |
| D-023 | Low | REQ-DS-19 배너 구현 메커니즘 plan.md Phase 5 구체화 | plan.md:L186~190 "배너 구현 메커니즘 (D-023): ... layouts/partials/version-banner.html ... hasPrefix 또는 findRE ... baseof.html 에서 조건부 호출" | **RESOLVED** |
| D-024 | Low | AC-G3-05 정규식 `grep -Eq "^### *(§?17\.[1-6])"` 강화 | acceptance.md:L380~385 6개 소섹션 각각 `grep -Eq "^### *(§?17\.N)"` 패턴 적용 | **RESOLVED** |
| D-025 | Low | AC-G4-03 `$PREVIEW_URL` 초기화 지침 추가 | acceptance.md:L539 "사용 지침: Preview URL 은 Vercel Dashboard → Deployments → 해당 Preview Deployment 의 URL 을 복사하여 환경변수로 export", L548 `test -n "$PREVIEW_URL" \|\| { echo "PREVIEW_URL not set"; exit 1; }` 가드 — AC-G4-03/04/05/06/11 전역 적용 | **RESOLVED** |
| Gap 1 | — | Vercel 무중단 근거 AC-PRE-03 신설 | AC-PRE-03 (acceptance.md:L480~490) "기존 Vercel project의 Git source 변경 시 도메인 바인딩이 유지되어 무중단 전환 가능하다는 근거" + plan.md:L263 Phase 7 사전 준비 | **RESOLVED** |
| Gap 2 | — | subtree 크기 실측 AC-G2-10 | AC-G2-10 (acceptance.md:L276~286) "Phase 2 이전/이후 `.git/` 크기 또는 squash 커밋 크기 실측치" + plan.md:L79 step 6 | **RESOLVED** |
| Gap 3 | — | Hugo 빌드 시간 베이스라인 AC-G2-11 | AC-G2-11 (acceptance.md:L288~298) 빈 콘텐츠/219 페이지 두 시점 시간 기록, Vercel timeout 여유치 계산 | **RESOLVED** |
| Gap 4 | — | FlexSearch 기본 비활성 결정 고정 | spec.md:L192 REQ-DS-27 "Hextra FlexSearch (disabled by default in this SPEC, enabling deferred to post-Phase 8 follow-up)" + plan.md:L411 "기본 비활성(`search: false` 유지)" | **RESOLVED** |
| Gap 5 | — | llms.txt URL 검증 AC-G4-11 | AC-G4-11 (acceptance.md:L660~672) 200 OK + `moai-adk-docs.vercel.app` 잔재 0건 | **RESOLVED** |
| Gap 6 | — | og.png 크기 AC-G4-05 보강 | acceptance.md:L587~589 `OG_SIZE ... <= 512000` (500 KB 한도) 자동 검증 | **RESOLVED** |
| Gap 7 | — | `_meta.ts` JSON Schema CI AC-G3-09 | AC-G3-09 (acceptance.md:L437~447) `go test ./scripts/validate-meta-schema/...` + plan.md:L200 step 5 `scripts/validate-meta-schema.go` | **RESOLVED** |

**해결률**: 32 / 32 (100%).

**자기 일관성 검증 (AC-G1-02 awk 자가 테스트)**:
- `grep -c "^\*\*REQ-DS-" spec.md` = **36**
- awk 범위 `15 ≤ N ≤ 36` → 36 은 경계값으로 **PASS**
- plan.md §7 G1 "15~36" 와 일치

---

## 3. 신규 결함 (Iteration 2 Chain-of-Verification 발견)

### D-026 (CRITICAL) — AC-G1-01 자기 일관성 FAIL: spec.md 265 LOC < 400 LOC 필수

- **위치**: acceptance.md:L28 (`test "$(wc -l < .moai/specs/SPEC-DOCS-SITE-001/spec.md)" -ge 400`), plan.md:L362 ("spec.md 400+ LOC"), spec.md 실제 LOC=265
- **기대**: AC-G1-01 이 요구하는 spec.md 최소 400 LOC 충족
- **실제**: `wc -l .moai/specs/SPEC-DOCS-SITE-001/spec.md` = 265 → `test -ge 400` exit 1
- **영향**: Iteration 1 D-001 과 동일한 유형의 자가 모순. SPEC 이 스스로 정의한 G1 자동 검증 스크립트를 통과하지 못해 G1 게이트를 통과할 수 없음.
- **상세 증거**:
  - acceptance.md L21 "각 파일이 최소 LOC 기준 (spec.md ≥ 400, plan.md ≥ 300, acceptance.md ≥ 250) 을 만족해야 한다"
  - acceptance.md L28 자동 검증 bash: `test "$(wc -l < .moai/specs/SPEC-DOCS-SITE-001/spec.md)" -ge 400`
  - plan.md L362 "3-file set 존재 및 최소 길이 준수 (spec.md 400+ / plan.md 300+ / acceptance.md 250+ LOC)"
  - 실측: spec.md 265 LOC, plan.md 428 LOC, acceptance.md 849 LOC
  - plan.md, acceptance.md 는 각각 300/250 하한 충족. spec.md 만 하한 미달.
- **권장 수정 (둘 중 택1)**:
  1. **하한 조정 (간단)**: AC-G1-01 (acceptance.md:L21, L28) 및 plan.md:L362 의 spec.md 기준을 `>= 400` → `>= 250` 또는 `>= 260` 으로 완화. D-001 해결 시 acceptance.md와 plan.md 범위를 동시 통일한 패턴과 동일한 접근.
  2. **spec.md 증량 (보수)**: 본문에 Phase 0 Discovery 핵심 요약 (~135 LOC) 을 §5 직전에 추가하거나, EARS 요구사항에 간단한 근거 주석을 부기하여 400+ LOC 달성.
  추천: (1). 사유: spec.md 의 본문 밀도는 이미 높으며, 하한 400 은 초기 초안 당시 임의로 설정된 값으로 근거가 약하다. 반면 본문 증량은 낭비.

### D-027 (Medium, advisory) — YAML Frontmatter `created_at` / `labels` 필드 미검증

- **위치**: spec.md:L2~L10 프론트매터
- **기대**: Iteration 감사 MP-3 기준 상 `created_at` (ISO date string), `labels` (array or string) 필드 존재
- **실제**: `created: 2026-04-17` (not `created_at`), `labels` 필드 부재. `issue_number: null` 존재.
- **비고**: 본 결함은 Iteration 1 에서도 동일하게 존재했으나 당시 리포트에 언급되지 않음 — 프로젝트 관례 혹은 선행 판정의 묵시적 수용으로 해석 가능. Adversarial pass (M2) 에 따라 기록만 남기고 CRITICAL 승격은 하지 않음.
- **영향**: 낮음. spec 주 내용 해석에는 영향 없음. 메타데이터 도구가 `created_at` 혹은 `labels` 에 의존할 경우 후속 파이프라인에서 경고 가능.
- **권장 수정 (선택적)**: `created` → `created_at` 필드명 재정비 또는 프로젝트 내부 YAML 스키마 공식화. 본 감사에서는 Iteration 1 판정 일관성을 위해 **선택 권고**로 둠.

---

## 4. 회귀 검증 결과

### 4.1 Iteration 1 강점 보존 검증

| 강점 항목 (iter1 §5) | iter2 상태 | 증거 |
|-----------------------|------------|------|
| Phase 0 R1~R10 Risk Register 추적성 | 보존 + 2건 추가 (R11, R12) | plan.md:L343~356, 신규 R11 (48h 모니터링 중 의사결정 지연), R12 (Hextra 빌드 timeout) |
| Dead code 명시적 Exclusions | 보존 | spec.md:L231 [HARD] Dead code 이식 금지 — 기존 7개 항목 유지 |
| SPEC-I18N-001 흡수·아카이브 프로세스 | 보존 | spec.md:L196 REQ-DS-28 동일 문구 유지, AC-G3-07 (acceptance.md:L405) 연결 |

결론: 3/3 강점 모두 보존, 일부는 오히려 강화됨 (Risk Register 에 2건 추가).

### 4.2 REQ 번호 연속성 검증

- 01~31 연속 ✓
- 32 자체는 부재, 32a/32b 로 분할 (HISTORY L17~18, plan.md §7 G1, acceptance.md 부록 C 에서 3개 파일에 모두 일관 명시)
- 33, 34, 35 연속 ✓
- 총 36 개 식별자, MP-1 엄격 해석 상 32 부재는 "gap" 으로 볼 수 있으나, 분할 이유와 매핑이 문서 3개에 동시 기록되어 있어 bookkeeping error 로 분류되지 않음. **합격** (단, 후속 감사에서 관용을 제한하려면 명시적 재넘버링 권고 가능).

### 4.3 D1~D4 이중 명시 검증

- spec.md §5 L119~122 D1/D2/D3/D4 명시 ✓
- plan.md §2 L19~22 D1/D2/D3/D4 명시 ✓
- 양쪽 문서 문구가 실질적으로 일치 (v2.12.0 스냅샷 시점, archive 태그명, Edge Function 경로 등 핵심 값 동일)

### 4.4 EARS 패턴 준수 (신규 REQ)

| 신규 REQ | 선언 패턴 | 실제 문구 시작 | 판정 |
|----------|----------|----------------|------|
| REQ-DS-32a | Event-driven | "When Phase 7 enters the cutover preparation step..." | PASS |
| REQ-DS-32b | Event-driven | "When acceptance.md § G4 checklist passes..." | PASS |
| REQ-DS-35 | Unwanted | "If any AC-MON-01 threshold is violated..., then the operator shall..." | PASS |

MP-2 EARS Format Compliance: **PASS**.

### 4.5 이모지 / 한국어 준수

- 국기 이모지 spec.md:L69 에 1건 존재 — 컨텍스트: "기존 LanguageSelector의 국기 이모지(🇰🇷/🇺🇸/🇯🇵/🇨🇳)는 ... 텍스트 라벨로 전환한다" 라는 제거 대상 설명용 인용. REQ-DS-29 CI 검증 대상은 `docs-site/` 본문이며, SPEC 파일은 대상 경로가 아니므로 위반 아님.
- 한국어 서술 일관성: 전반적으로 한국어, 기술 용어는 영어 그대로 사용 (예: "Phase 7", "Edge Function") — 정상.

### 4.6 AC 모순 / 중복 검사

- AC-G2-04 (`-eq 735`) vs REQ-DS-05 ("735 total, absolute preservation") → 일치
- AC-G2-06 (`-eq 569`) vs REQ-DS-09 ("all 569 ... absolute count, no tolerance") → 일치
- AC-G3-01 runtime grep vs AC-G4-03 Preview 헤더 검증 → 역할 분리 일관 (로컬 정적 선언 검증 + 실제 배포 후 동적 헤더 검증)
- AC-PRE-01 DNS 조사 vs REQ-DS-32a "DNS is **not** modified during this step" → 조사 결과에 따라 수행 분기, 모순 없음

결론: 회귀/모순 0건.

---

## 5. Chain-of-Verification Pass (M6)

### 2차 점검에서 추가 발견

1. **D-026 (CRITICAL)**: 1차 pass 는 AC-G1-02 (REQ 개수) 만 자동 실행하고 AC-G1-01 (LOC) 은 자연스럽게 건너뛰었다. 2차 pass 에서 "AC-G1-01 의 bash 블록도 동일한 자가 검증이므로 실제 실행해야 한다" 고 판단하고 `wc -l` 을 실제 돌린 결과 265 LOC < 400 LOC 검출. Iteration 1 D-001 과 완전히 동일한 자가 모순 패턴. 만약 이를 건너뛰면 Iteration 3 Run 진입 시 CI 가 동일한 이유로 다시 FAIL 될 위험.
2. **D-027 (Medium, advisory)**: YAML frontmatter 의 `created` vs `created_at` 차이, `labels` 부재는 MP-3 엄격 해석 대상. 단, Iteration 1 에서 동일 상태로 수용된 점을 존중하여 어드바이저리 처리.
3. **재점검한 영역** — 지시된 항목 모두 end-to-end 확인:
   - REQ 번호 01~35 (32 분할 포함) 전수 확인 via sorted grep 출력
   - Exclusions 항목 12개 (≥10 기준) 확인
   - 부록 C REQ↔AC 매트릭스 36행 전수 확인 (`grep -c` 36)
   - 4개 locale 관련 REQ (04/12/13/15) 재점검
   - D1~D4 이중 명시 재점검
   - 신규 게이트 G1.5 + 신규 AC 블록 (AC-G1.5-01~05, AC-G2-08~11, AC-G3-08/09, AC-G4-11, AC-MON-03, AC-PRE-01~03) 전수 열람

---

## 6. Regression Check (Iteration 1 결함 해결 집계)

| 이전 결함 분류 | 총 건수 | 해결 | 미해결 |
|-----------------|--------|------|--------|
| Critical | 2 | 2 | 0 |
| High | 6 | 6 | 0 |
| Medium | 11 | 11 | 0 |
| Low | 6 | 6 | 0 |
| Gap (누락) | 7 | 7 | 0 |
| **합계** | **32** | **32** | **0** |

모든 이전 결함 해결 확인.

---

## 7. 최종 판정

### 판정: **FAIL** (Critical 1, High 0, Medium 1 advisory)

### 판정 근거

- 판정 기준 "Critical ≥ 1 → FAIL" 이 D-026 에 의해 엄격히 트리거됨.
- 다만 D-026 은 단일 값 조정(400 → 250 또는 260) 으로 해결되는 비구조적 결함이며, Iteration 2 의 나머지 산출물은 사전 결함 32 건 전량 해결로 매우 높은 품질을 보여줌.
- Iteration 1 대비 정성적 품질은 크게 개선되었으나, plan-auditor 는 criteria 를 임의로 완화하지 않는다 (M5 Must-Pass Firewall).

### Iteration 3 권고 액션

**단 1개의 수정만으로 PASS 도달 가능**:

1. **D-026 해결 (필수)**:
   - 옵션 A (권장): `acceptance.md:L21` 의 "spec.md ≥ 400" → "spec.md ≥ 250" 수정, `acceptance.md:L28` 의 `-ge 400` → `-ge 250` 수정, `plan.md:L362` 의 "spec.md 400+ LOC" → "spec.md 250+ LOC" 수정. 세 곳 동시에 맞추면 D-001 해결 시 사용한 "세 위치 통일" 패턴과 일관.
   - 옵션 B (대안): spec.md §6 각 REQ 에 간결한 "rationale" 한 줄 주석을 부가하거나 §5 References 섹션을 보강하여 400 LOC 달성.

2. **D-027 (선택 advisory)**:
   - `spec.md:L5` 의 `created: 2026-04-17` → `created_at: "2026-04-17"` 로 ISO 날짜 + 필드명 정합.
   - `labels:` 필드 추가 (예: `labels: [migration, docs, hextra]`).

### Iteration 3 기대 결과

- D-026 단독 해결 → Critical 0, High 0 → **PASS** 획득 가능
- D-027 동시 해결 시 MP-3 YAML Frontmatter Validity 까지 완벽 충족

### 감사 종료 메시지

Iteration 2 는 Iteration 1 의 심각한 자가 모순 및 구조적 결함을 모두 해결했다. 하지만 감사 프로토콜 M6 (Chain-of-Verification) 를 통해 이전 감사가 간과했던 자가 모순 1건이 새로 드러났다. 이 단일 결함은 한 번의 최소한의 개정으로 해소 가능하며, Iteration 3 심사에서는 PASS 가 매우 유력하다.

---

## 8. 증거 인용 요약

| 검증 항목 | 1차 증거 | 2차 증거 |
|-----------|----------|----------|
| REQ 개수 36 | spec.md:L128~L216 `grep -c "^\*\*REQ-DS-"` 결과 36 | acceptance.md:L41 awk `$1 <= 36` PASS |
| D-026 LOC 미달 | spec.md `wc -l` = 265 | acceptance.md:L28 `-ge 400`, plan.md:L362 "400+ LOC" |
| D-002 시퀀싱 분리 | spec.md:L208 REQ-DS-32a, spec.md:L210 REQ-DS-32b | plan.md:L266 step 2 재바인딩 (G4 이전), plan.md:L281 step 5 Production 프로모션 (G4 이후) |
| D-003 G1.5 신설 | plan.md:L367~373 G1.5 5개 기준 | acceptance.md:L68, L80, L96, L115, L123 AC-G1.5-01~05 |
| D-004 DNS TTL 제거 | spec.md:L208 REQ-DS-32a "DNS is not modified" | AC-PRE-01 (acceptance.md:L456~466) 근거 기반 조사 |
| D-005/D-006 절대 보존 | acceptance.md:L200 `-eq 735`, L229 `-eq 569` | REQ-DS-05 / REQ-DS-09 "absolute preservation" 문구 일치 |
| D-007 Edge runtime | spec.md:L158 REQ-DS-13 `runtime: 'edge'` | AC-G3-01 (acceptance.md:L325) + AC-G4-03 (L554) 이중 검증 |
| D-014 롤백 REQ-DS-35 | spec.md:L216 Unwanted EARS | acceptance.md:L710~725 AC-MON-03 4단계 |
| REQ↔AC 커버리지 | acceptance.md:L810~849 부록 C | `grep -c "^\| REQ-DS-"` = 36 |

---

**감사 종료** — Iteration 3 재심사 요청 시 D-026 해결 증거(spec.md 또는 acceptance.md/plan.md 라인 인용)를 changelog 에 포함해주시기 바랍니다.

- **출력 경로**: `/Users/goos/MoAI/moai-adk-go/.moai/plans/DOCS-SITE/g1-audit-report-iter2.md`
