# SPEC-DOCS-SITE-001 — Iteration 2 Changelog

- **Author**: manager-spec
- **Date**: 2026-04-20
- **Scope**: g1-audit-report.md의 25개 결함 (Critical 2 + High 6 + Medium 11 + Low 6) + 누락 항목 7건 대응
- **입력**: `.moai/plans/DOCS-SITE/g1-audit-report.md` (Iteration 1 FAIL 리포트)

---

## 1. 파일 변화 요약 (LOC)

| 파일 | Iteration 1 (before) | Iteration 2 (after) | 증감 |
|------|---------------------|---------------------|------|
| spec.md | 222 | 265 | +43 |
| plan.md | 370 | 428 | +58 |
| acceptance.md | 541 | 849 | +308 |
| **합계** | **1,133** | **1,542** | **+409** |

## 2. REQ 변화

- **Iteration 1**: REQ-DS-01 ~ REQ-DS-34 총 34건
- **Iteration 2**: REQ-DS-01 ~ REQ-DS-35 (단, REQ-DS-32는 32a/32b로 분할) 총 **36건**
- **상한 조정**: AC-G1-02 `15 ≤ N ≤ 30` → `15 ≤ N ≤ 36` (D-001)
- **plan.md §7 G1**: `15~25` → `15~36` (D-008 통일)

## 3. Critical / High 해결 상황 (필수 8건)

| ID | 심각도 | 제목 | 해결 여부 | 반영 위치 |
|----|--------|------|-----------|-----------|
| D-001 | Critical | REQ 개수 상한 모순 | O | AC-G1-02 상한 36, plan.md G1 `15~36` |
| D-002 | Critical | REQ-DS-32 시퀀싱 모순 | O | spec.md REQ-DS-32a (cutover prep, G4 이전) / REQ-DS-32b (promotion, G4 이후) 분할 |
| D-003 | High | Phase 2 전용 게이트 누락 | O | plan.md §7 G1.5 (Scaffold Readiness) 신설, acceptance.md AC-G1.5-01 ~ AC-G1.5-05 5건 추가 |
| D-004 | High | DNS TTL 조항 근거 부족 | O | REQ-DS-32에서 (e) DNS TTL 제거, acceptance.md AC-PRE-01 신설 (Vercel 지원팀 근거 기반 조사 기록) |
| D-005 | High | REQ-DS-09 vs AC-G2-06 tolerance 불일치 | O | AC-G2-06 `test "$COUNT" -eq 569` 절대 보존으로 강화 |
| D-006 | High | REQ-DS-05 vs AC-G2-04 tolerance 불일치 | O | AC-G2-04 `test "$TOTAL" -eq 735` 절대 보존으로 강화 |
| D-007 | High | Edge Function runtime 명시 누락 | O | REQ-DS-13 본문에 `runtime: 'edge'` 명시, plan.md Phase 5 platform constraint 사전 검토 단계 추가, AC-G3-01에 runtime 선언 grep 검증 + AC-G4-03에 헤더 검증 추가 |
| D-008 | High | G1 EARS 범위 문서 간 불일치 | O | plan.md §7 G1 `15~36`으로 통일 (acceptance.md와 일치) |

**요약**: Critical 2/2 해결, High 6/6 해결 — **필수 8건 100% 해결**.

## 4. Medium 해결 상황 (11건)

| ID | 제목 | 해결 여부 | 반영 위치 |
|----|------|-----------|-----------|
| D-009 | 다수 REQ의 AC 커버리지 누락 | O | AC-G2-08 (REQ-DS-19 배너 partial), AC-G2-09 (REQ-DS-29 금지 패턴 CI), AC-G3-08 (REQ-DS-20 3종 partial + 이모지 0건) 신설. 부록 C에 REQ↔AC 매핑 매트릭스 추가 |
| D-010 | v2.12.0 스냅샷 타이밍 모호성 | O | REQ-DS-17에 "the commit tagged as the previous release (v2.12.X), not HEAD of main" 명시 |
| D-011 | AC-G3-04 테스트 케이스 이름 오류 | O | TestMinorRelease (v2.12.0→v2.13.0), TestPatchRelease (v2.12.1→v2.12.2), TestMajorRelease (v2.12.0→v3.0.0) 3종 교정 + 신규 |
| D-012 | 플래그 이모지 제거와 디자인 리디자인 금지 충돌 | O | spec.md §2 Non-goals에 "이모지 금지 정책에 따른 국기 이모지 → 텍스트 라벨 전환 예외 허용" 조항 추가, §7 제약에 상호 참조, §8 Exclusions의 디자인 리디자인 금지 항목에 괄호 예외 명시 |
| D-013 | Lighthouse Nextra baseline 상대 기준 누락 | O | AC-G4-07 표에 "Nextra baseline 대비 Performance -5 이내" 컬럼 추가, plan.md Phase 7 step 1에 baseline 측정 단계 추가 |
| D-014 | 48h 모니터링 실패 롤백 누락 | O | spec.md REQ-DS-35 (Unwanted) 신설, acceptance.md AC-MON-03 대응 AC 추가 (롤백 절차 4단계 + 기록 경로) |
| D-015 | Hugo module vs submodule 결정 | O | spec.md §7 제약에 "Hextra는 Hugo module system(go.mod import)으로만 통합, git submodule 사용 금지" 잠금. plan.md Phase 2 step 3 및 §8 Technical Approach에서 재확인 |
| D-016 | empty skeleton placeholder folders 구체화 | O | REQ-DS-15 본문에 "single `_index.md` with `draft: true` + redirect alias to ko locale" 명시. plan.md Phase 3 step 2에 `aliases: ["/ko/<section>/"]` 구체화 |
| D-017 | Vercel 빌드 timeout 검증 | O | plan.md Phase 2 step 7 "빌드 시간 베이스라인 측정" 추가, AC-G1.5-05 + AC-G2-11 자동 검증 |
| D-018 | Hextra 한계 문서화 | O | spec.md §4.1 "Hextra 한계 (Acknowledged Limitations)" 소섹션 신설 — versioning 내장 없음, React/JSX 미지원, 서버 사이드 Mermaid 프리렌더 미제공 3건 명시 |
| D-019 | aggregateRating SEO 영향 평가 | O | plan.md Phase 4 step 8 "SEO 영향 사전 조사" 추가, acceptance.md AC-PRE-02 신설 (Google Search Console 조사 + phase-4-seo-impact.md 기록) |

**요약**: 11/11 해결 — **100%**.

## 5. Low 해결 상황 (6건)

| ID | 제목 | 해결 여부 | 반영 위치 |
|----|------|-----------|-----------|
| D-020 | plan.md §7 G2 표기 | O | "(Phase 3 → Phase 5)" → "(Phase 3/4 → Phase 5)" 수정 |
| D-021 | REQ-DS-34 Phase 7/8 불일치 | O | REQ-DS-34 주어를 "the system (Phase 8 automation, executed by manager-git and expert-devops)"으로 명확화 |
| D-022 | moai-docs 전용 config 이관 제외 | O | spec.md §8 Exclusions에 "moai-docs의 gate.yaml / memo.yaml / observability.yaml 3개 이관 제외" 명시 |
| D-023 | REQ-DS-19 배너 구현 메커니즘 | O | plan.md Phase 5 step 2 배너 구현 세부 (partial 파일, hasPrefix/findRE 판정, latest URL 산출) 추가. AC-G2-08 자동 검증 |
| D-024 | AC-G3-05 정규식 취약 | O | `grep -Eq "^### *(§?17\.[1-6])"`로 6개 소섹션 각각 강화 |
| D-025 | AC-G4-03 PREVIEW_URL 사용 지침 | O | AC-G4-03에 "Vercel Dashboard → Deployments에서 URL 복사 후 export" 지침 명시, 모든 Preview URL 사용 AC에 `test -n "$PREVIEW_URL"` 가드 추가 |

**요약**: 6/6 해결 — **100%**.

## 6. Gap (누락 항목) 해결 상황 (7건)

| Gap | 제목 | 해결 여부 | 반영 위치 |
|-----|------|-----------|-----------|
| Gap 1 | Vercel rebinding 무중단 근거 | O | acceptance.md AC-PRE-03 신설 (Vercel 공식 문서/지원팀 근거 → phase-7-zero-downtime-evidence.md) |
| Gap 2 | git subtree 압축 후 크기 실측 | O | plan.md Phase 2 step 6 + AC-G1.5-05 + AC-G2-10 자동 검증 |
| Gap 3 | Hugo 빌드 시간 베이스라인 | O | plan.md Phase 2 step 7 + AC-G1.5-05 + AC-G2-11 자동 검증 |
| Gap 4 | FlexSearch 기본 활성/비활성 결정 | O | spec.md REQ-DS-27 본문에 "disabled by default in this SPEC, enabling deferred to post-Phase 8 follow-up" 명시. plan.md §8 Technical Approach에 "기본 비활성 (`search: false` 유지)" 고정 |
| Gap 5 | llms.txt 새 도메인 URL 검증 | O | AC-G4-11 신설 (200 OK + moai-adk-docs.vercel.app 잔재 0건 검증) |
| Gap 6 | og.png ≤ 500 KB AC | O | AC-G4-05에 Content-Length ≤ 512000 bytes 자동 검증 추가 |
| Gap 7 | `_meta.ts` JSON Schema CI AC | O | plan.md Phase 5 step 5 `scripts/validate-meta-schema.go` 추가, AC-G3-09 자동 검증 |

**요약**: 7/7 해결 — **100%**.

## 7. 주요 구조 변경

### 7.1 REQ-DS-32 분할 (D-002 해결)

- **이전 (Iteration 1)**:
  - REQ-DS-32 (Event-driven): "When G4 is granted, ... reconfigure Vercel project ... DNS TTL..." — G4 이후 재바인딩 + DNS (논리적 모순: Preview 검증 불가).
- **이후 (Iteration 2)**:
  - **REQ-DS-32a (Event-driven)**: Phase 7 cutover preparation 진입 시 재바인딩 (G4 **이전**). DNS 조작 제거.
  - **REQ-DS-32b (Event-driven)**: G4 승인 후 Production 프로모션 (G4 **이후**).
  - DNS 관련 조항은 AC-PRE-01로 이관 (사전 조사 후 필요 시 별도 계획).

### 7.2 Gate 추가 (D-003)

- **새 Gate G1.5 — Scaffold Readiness (Phase 2 → Phase 3)**: plan.md §7에 신설.
- 대응 AC: AC-G1.5-01 ~ AC-G1.5-05 (squash 단일성, hugo.yaml 필드, Nextra 잔재 0건, hugo server 기동, 베이스라인 문서).

### 7.3 신규 REQ-DS-35 (D-014)

- Unwanted 패턴. 48h 위반 시 Vercel Git source 즉시 원복 + `.moai/plans/DOCS-SITE/phase-7-rollback.md` 기록 의무.
- 대응 AC-MON-03에 4단계 롤백 절차 명시.

### 7.4 Preflight AC 신설

- AC-PRE-01 (DNS 필요성, D-004)
- AC-PRE-02 (SEO aggregateRating 영향, D-019)
- AC-PRE-03 (Vercel 무중단 근거, Gap 1)

### 7.5 Hextra 한계 소섹션 (D-018)

- spec.md §4.1 신설. 향후 "왜 React 컴포넌트 못 씀?" 유형 재논의 방지.

### 7.6 부록 C: REQ ↔ AC 매핑 매트릭스 (D-009)

- acceptance.md 끝에 35개 REQ 각각에 대응 AC를 표로 매핑. Definition of Done의 "모든 REQ에 대응 AC" 조항을 자동 검증 가능한 형태로 확보.

## 8. HISTORY 엔트리

spec.md `## HISTORY` 섹션에 `v0.2.0 (Iteration 2, 2026-04-20)` 엔트리 신설. 25개 결함 ID(D-001 ~ D-025) + Gap 1~7 각각 one-liner 요약 기록.

## 9. 검증 체크리스트

Iteration 2 완료 시점에서 다음 검증이 선행되어야 한다:

- [x] spec.md REQ 개수 grep 결과: 36 (15~36 범위 내)
- [x] AC-G1-02 awk 검증식 (`$1 >= 15 && $1 <= 36`)과 plan.md §7 G1 "15~36" 일치
- [x] REQ-DS-32 분할 후 spec.md / plan.md / acceptance.md 3개 파일 모두 32a/32b 참조 교차 갱신
- [x] G1.5 신설 후 plan.md §3 테이블, §7 Gate 세부 기준, acceptance.md AC-G1.5 5건 모두 존재
- [x] Edge runtime 검증이 AC-G3-01 (로컬) + AC-G4-03 (Preview) 두 곳에서 일관
- [x] `aggregateRating` 제거 검증이 REQ-DS-21 + AC-G4-06 일관
- [x] Callout 735 / Mermaid 569 절대 보존이 REQ-DS-05 / REQ-DS-09 / AC-G2-04 / AC-G2-06 일관
- [x] HISTORY 엔트리에 25개 D-번호 + 7개 Gap 모두 기재
- [x] 이모지 0건 (본 changelog 포함)
- [x] 한국어 작성 준수

## 10. Plan Auditor 재심사 요청

본 Iteration 2 결과물을 Plan Auditor에게 독립 재심사 요청할 준비 완료. 기대 결과:

- Critical 0 + High ≤ 2 → **PASS** 달성 목표
- 본 changelog가 resolution 증거 (라인 인용 포함) 역할 수행

**산출물 3종 경로**:
- `/Users/goos/MoAI/moai-adk-go/.moai/specs/SPEC-DOCS-SITE-001/spec.md` (265 LOC)
- `/Users/goos/MoAI/moai-adk-go/.moai/specs/SPEC-DOCS-SITE-001/plan.md` (428 LOC)
- `/Users/goos/MoAI/moai-adk-go/.moai/specs/SPEC-DOCS-SITE-001/acceptance.md` (849 LOC)
- **본 changelog**: `/Users/goos/MoAI/moai-adk-go/.moai/plans/DOCS-SITE/spec-iteration2-changelog.md`

---

**Iteration 2 종료**. Plan Auditor 재심사 대기.
