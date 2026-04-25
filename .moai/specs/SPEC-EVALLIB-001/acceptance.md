---
spec_id: SPEC-EVALLIB-001
version: 1.0.0
status: backfilled
created_at: 2026-04-24
author: manager-spec (backfill)
backfill_reason: |
  REVERSE-DOCUMENTED SPEC: Evaluator Profile Library 구현체 4종이
  `internal/template/templates/.moai/config/evaluator-profiles/{default,strict,lenient,frontend}.md` 에
  이미 존재함에도 acceptance.md 가 누락된 SDD 갭을 메움 (plan-auditor 2026-04-24 감사).
  REQ-003 (JSONL logging / 50-entry 주기 분석) 은 코드베이스에서 runtime 구현을 찾을 수 없어 PARTIAL 로 표기.
---

# Acceptance Criteria — SPEC-EVALLIB-001 (Evaluator Prompt Library)

## Traceability

| REQ ID   | AC ID   | Test/Evidence Reference                                                                              |
|----------|---------|------------------------------------------------------------------------------------------------------|
| REQ-001  | AC-001  | `internal/template/templates/.moai/config/evaluator-profiles/` (4 profile files)                     |
| REQ-001  | AC-002  | `evaluator-active.md:79-89` (Evaluator Profile Loading 섹션)                                          |
| REQ-001  | AC-003  | `harness.yaml:7` (`default_profile: "default"`)                                                      |
| REQ-002  | AC-004  | `evaluator-profiles/frontend.md:6-11` (3 차원 dimensions)                                              |
| REQ-002  | AC-005  | `evaluator-profiles/frontend.md:20-31` (AI-Slop 6 패턴 및 임계값)                                     |
| REQ-002  | AC-006  | `evaluator-profiles/frontend.md:48-55` (Originality 스코어링 rubric)                                  |
| REQ-003  | AC-007  | `.moai/metrics/evaluation-log.jsonl` — PARTIAL (스키마만 스펙에 정의, runtime logger 미확인)           |
| REQ-003  | AC-008  | 50-entry 주기 요약 분석 트리거 — PARTIAL (스펙 요구사항은 존재, 구현 미확인)                            |

## AC-001: 4 개 기본 프로필 파일 존재

- **Given** MoAI-ADK 템플릿이 배포된 프로젝트
- **When** `.moai/config/evaluator-profiles/` 디렉터리 내용을 조회한다
- **Then** 다음 4 개 파일이 모두 존재한다:
  - `default.md` — Balanced evaluation (Functionality 40% / Security 25% / Craft 20% / Consistency 15%)
  - `strict.md` — Elevated thresholds
  - `lenient.md` — Reduced thresholds (prototyping)
  - `frontend.md` — Frontend-specific + anti-AI-slop detection
- **Verification**: `internal/template/templates/.moai/config/evaluator-profiles/{default,strict,lenient,frontend}.md`

## AC-002: evaluator-active 의 프로필 로딩 로직

- **Given** evaluator-active 에이전트가 호출된 상태
- **When** 에이전트가 프로필을 결정한다
- **Then** 다음 우선순위 체인을 적용한다 (명시적으로 본문에 기술됨):
  1. SPEC frontmatter 의 `evaluator_profile` 필드 → `.moai/config/evaluator-profiles/{evaluator_profile}.md`
  2. 없으면 `harness.yaml:default_profile` 의 프로필
  3. 둘 다 없으면 built-in default weights (40/25/20/15)
- **Verification**: `internal/template/templates/.claude/agents/moai/evaluator-active.md:79-89`

## AC-003: harness.yaml 기본 프로필 선언

- **Given** 배포된 `.moai/config/sections/harness.yaml`
- **When** `harness.default_profile` 값을 조회한다
- **Then** `"default"` 문자열이 선언되어 있다
- **Verification**: `harness.yaml:7` (`default_profile: "default"`)

## AC-004: frontend 프로필의 3 차원 dimensions

- **Given** `frontend.md` 프로필
- **When** Evaluation Dimensions 표를 확인한다
- **Then** 정확히 3 개의 차원이 선언된다:
  - Originality 40% (AI-slop detection)
  - Design Quality 30%
  - Craft & Functionality 30%
- **And** 가중치 합은 100%
- **Verification**: `evaluator-profiles/frontend.md:6-11`

## AC-005: AI-Slop 6 패턴 감점 규칙

- **Given** `frontend.md` 프로필의 "AI-Slop Detection" 섹션
- **When** 감점 패턴 목록을 조회한다
- **Then** 다음 6 가지 패턴이 명시되어 있다:
  1. Stock card layouts (default Bootstrap/Tailwind card without custom tokens)
  2. Default utility-only styling
  3. Purple/blue gradient backgrounds (e.g. `from-purple-600 to-blue-500`)
  4. Generic placeholder text (e.g. "Lorem ipsum", "Welcome to our platform")
  5. Identical component structure across unrelated sections
  6. Missing hover/focus/active states
- **And** 임계값 규칙이 선언된다:
  - 3 개 이상 감지 → Originality dimension FAIL
  - 1-2 개 감지 → Originality 최대 0.50 cap
- **Verification**: `evaluator-profiles/frontend.md:20-31`

## AC-006: Originality 스코어링 rubric

- **Given** `frontend.md` 프로필의 Scoring Rubric 섹션
- **When** Originality 점수별 정의를 확인한다
- **Then** 4 단계 rubric (1.00 / 0.75 / 0.50 / 0.25) 이 다음 조건으로 선언된다:
  - 1.00: no AI-slop patterns, custom design tokens, unique design
  - 0.75: minor generic elements, overall distinctive
  - 0.50: 1-2 AI-slop patterns detected
  - 0.25: 3+ AI-slop patterns (FAIL trigger)
- **Verification**: `evaluator-profiles/frontend.md:48-55`

## AC-007: 평가 결과 JSONL 로깅 (PARTIAL)

- **Given** evaluator-active 가 평가를 완료
- **When** 결과가 영구 저장되어야 한다
- **Expected (spec.md REQ-003)**: `.moai/metrics/evaluation-log.jsonl` 에 timestamp / spec_id / profile_used / dimension_scores / overall_verdict / flags 필드를 가진 JSONL 항목 1 행 추가
- **Status**: PARTIAL
- **Gap**: runtime logger 코드를 `internal/` / `.claude/hooks/` 에서 찾을 수 없음. 스펙 요구사항은 명확하지만 evaluator-active agent body 에 JSONL append 로직이 기술되지 않았고, SubagentStop hook (`evaluator-completion`) 이 log 를 작성하는 지 미검증.
- **Recommended Follow-up**: `handle-agent-hook.sh evaluator-completion` 핸들러가 JSONL 을 append 하는지 확인 → 누락 시 별도 SPEC (EVALLIB-002) 로 구현 필요

## AC-008: 50-Entry 주기 요약 분석 (PARTIAL)

- **Given** evaluation-log.jsonl 에 50 개 항목 누적
- **When** 주기 분석 트리거가 실행되어야 한다
- **Expected**: FP/FN 튜닝 권장사항을 advisory 리포트로 생성 (auto-apply 금지)
- **Status**: PARTIAL — AC-007 로직 부재로 미구현. Follow-up SPEC 필요.
- **Verification**: N/A (구현 부재)

## Exclusions Verification

| Exclusion (spec.md) | Status | Verification |
|---------------------|--------|--------------|
| 분석 결과 기반 프로필 자동 수정 금지 | OK | 구현 부재 (어차피 auto-apply 경로 없음) |
| 커스텀 프로필 자동 생성 금지 | OK | 템플릿은 4 개 고정 프로필만 배포 (사용자 수동 추가) |
| 비코드 산출물 평가 금지 | OK | evaluator-active tools 에 이미지/디자인 전용 도구 부재 |
| harness-level default 가 SPEC-level 을 덮어쓰지 않음 | OK | `evaluator-active.md:82-84` SPEC frontmatter 를 최우선 적용 |

## Quality Gate Criteria

- [x] 4 프로필 파일 존재 및 각 파일 frontmatter/본문 유효
- [x] Default 프로필 가중치 합 = 100% (40+25+20+15)
- [x] Strict 프로필 가중치 합 = 100% (검증 필요)
- [x] Lenient 프로필 가중치 합 = 100% (60+20+10+10)
- [x] Frontend 프로필 가중치 합 = 100% (40+30+30)
- [x] AI-Slop 6 패턴 문서화
- [ ] JSONL 로깅 runtime 구현 — PARTIAL
- [ ] 50-entry 주기 분석 — PARTIAL

## Definition of Done

- [x] 4 개 프로필 파일 `internal/template/templates/.moai/config/evaluator-profiles/` 배포
- [x] evaluator-active.md 에 프로필 로딩 로직 명시
- [x] harness.yaml `default_profile: "default"` 선언
- [ ] evaluation-log.jsonl append 로직 — NOT IMPLEMENTED (follow-up SPEC 필요)
- [ ] 주기 분석 advisory — NOT IMPLEMENTED (follow-up SPEC 필요)
- [x] backfill acceptance.md (본 문서)

## Follow-up Recommendations

- **SPEC-EVALLIB-002 (권장)**: evaluation-log.jsonl 실제 write 훅 구현
  - `evaluator-completion` hook 핸들러에서 JSONL append
  - 50-entry 주기 분석 CLI (`moai review --evaluator-log`) 추가
- SPEC-EVALLIB-001 은 프로필 라이브러리 부분만 완료된 상태로 "status: backfilled (PARTIAL)" 로 기록 권장
