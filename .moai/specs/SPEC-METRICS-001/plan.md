---
id: SPEC-METRICS-001
plan_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Implementation Plan — SPEC-METRICS-001

## 1. Overview

`/moai sync` 워크플로우에 PR-level contribution metric 자동 수집 단계 추가. Anthropic conservative measurement (confidence threshold 0.70) 적용. `moai metrics show` CLI는 Phase 2 후속 SPEC.

## 2. Approach Summary

**전략**: Conservative-First, JSONL-Local, Phase-Separated.

1. `.claude/rules/moai/workflow/contribution-metrics.md` 정책 문서 신설
2. `.moai/metrics/contribution.jsonl` 영속화 schema (9 필드)
3. `/moai sync` 워크플로우에 metric 수집 단계 추가
4. Confidence calculation 휴리스틱 정의
5. Privacy 정책 (author = GitHub username, email 제외)
6. Opt-in/out 정책 결정 + `.moai/config.yaml` 통합

## 3. Milestones (Priority-based, no time estimates)

### M0 — Pre-flight (Priority: Critical)

- [ ] 본 프로젝트의 최근 30 PR 분석 — SPEC-ID 인용 비율, `🗿 MoAI` co-author 비율
- [ ] Anthropic blog "Claude Code Contribution Metrics" verbatim 재확인
- [ ] confidence threshold 0.70의 base rate 검증 (false positive rate 추정)
- [ ] Privacy 정책 결정: author = GitHub username only (email 제외)
- [ ] Opt-in vs opt-out 정책 결정 (M3에서 final)

**Exit Criteria**: baseline + privacy 결정 확보

### M1 — Policy Document 작성 (Priority: High)

- [ ] `.claude/rules/moai/workflow/contribution-metrics.md` 신규 작성
  - §1 Overview + scope
  - §2 Schema (9 필드)
  - §3 Confidence Calculation Heuristic
  - §4 Privacy Policy (email 제외)
  - §5 Opt-in/out (M3에서 결정)
  - §6 Storage (`.moai/metrics/contribution.jsonl`)
  - §7 Retention (사용자 책임)
  - §8 Phase Separation (Phase 1 = local JSONL, Phase 2 = `moai metrics show` + GitHub Analytics)
  - §9 Anti-Patterns (email leak, real-time push 등)
- [ ] 정책 문서 2-3KB

**Exit Criteria**: 9 절 모두 작성

### M2 — Schema 정의 (Priority: High)

- [ ] Metric entry schema (9 필드):
  ```jsonl
  {
    "pr_id": 741,
    "author": "GOOS",
    "claude_assisted_lines": 1234,
    "total_lines": 1500,
    "confidence": 0.85,
    "spec_ids": ["SPEC-CRON-PATTERN-001"],
    "timestamp": "2026-04-30T12:00:00Z",
    "merge_strategy": "merge",
    "branch_type": "feature"
  }
  ```
- [ ] `claude_assisted_lines` 계산:
  - SPEC commit으로부터 직접 파생된 line
  - SPEC-ID 인용 + `🗿 MoAI` co-author 모두 충족 시 신뢰
- [ ] `confidence` 계산 (heuristic):
  - PR-SPEC 인용 명시 (+0.30)
  - All commits have `🗿 MoAI` co-author (+0.30)
  - Commit messages reference SPEC-ID (+0.20)
  - PR title contains type (feat/fix/docs/etc.) (+0.10)
  - PR has reviewer approval (+0.10)
  - Total cap: 1.0; floor: 0.0

**Exit Criteria**: schema + confidence 휴리스틱 명시

### M3 — Opt-in/out 정책 + Privacy (Priority: High)

- [ ] Opt-in vs opt-out 결정:
  - **Opt-in default** 권고: `.moai/config.yaml metrics.enabled: false` (사용자 명시 활성화)
  - 이유: privacy first, 명시 동의
  - 활성화 시: `metrics.enabled: true`
- [ ] Privacy rules:
  - `author` = GitHub username (예: "GOOS", "goosadk")
  - email 영속화 금지
  - PR body / commit message 원문 영속화 금지 (요약만)
  - PII 검출 시 entry skip
- [ ] `.moai/config/sections/metrics.yaml` 신설 검토 또는 `quality.yaml` 통합
  - 결정: 신규 `metrics.yaml` (단일 책임)

**Exit Criteria**: opt-in 결정 + privacy 5 rules 명시

### M4 — Workflow Integration (Priority: High)

- [ ] `/moai sync` 워크플로우에 metric 수집 단계 추가:
  - Step 1: PR 생성 후 metric collection 호출
  - Step 2: Confidence calculation (heuristic)
  - Step 3: Author resolution (GitHub username only)
  - Step 4: SPEC-ID 추출 (PR body에서 SPEC-XXX-NNN 패턴 grep)
  - Step 5: JSONL append to `.moai/metrics/contribution.jsonl`
- [ ] Best-effort: metric 수집 실패 시 sync 차단 X
- [ ] `.claude/skills/moai-workflow-project/modules/contribution-metrics.md` 신설 (사용자 가이드)
- [ ] sync.md skill에 cross-ref 추가

**Exit Criteria**: workflow 통합 정의 명시

### M5 — Conservative Measurement Validation (Priority: Medium)

- [ ] sample 5 sync 호출 측정:
  - PR-SPEC 인용 명시 + co-author 충족 → confidence >= 0.70
  - 인용 누락 → confidence < 0.70 → assisted = 0
- [ ] False positive rate 추정 (manual review 5 entries)
- [ ] Conservative bias 검증: confidence < 0.70 → claude_assisted_lines = 0 (cap)
- [ ] PII detection regex (email, API key) 적용 확인

**Exit Criteria**: 5 sample validation PASS, false positive 0%

### M6 — Template-First Sync + Documentation (Priority: Medium)

- [ ] `internal/template/templates/.claude/rules/moai/workflow/contribution-metrics.md` 동기화
- [ ] `internal/template/templates/.claude/skills/moai-workflow-project/SKILL.md` 동기화
- [ ] `internal/template/templates/.moai/config/sections/metrics.yaml` 신설 (default: enabled: false)
- [ ] `make build` 실행 → embedded.go 재생성
- [ ] CHANGELOG entry under Unreleased

**Exit Criteria**: Template-First sync clean

### M7 — Validation + Acceptance Sign-off (Priority: High)

- [ ] acceptance.md 시나리오 모두 PASS
- [ ] 9 필드 schema 명시 검증
- [ ] confidence threshold 0.70 검증
- [ ] privacy 5 rules 검증
- [ ] cross-ref 2+ 검증 (grep)
- [ ] plan-auditor PASS

**Exit Criteria**: 모든 acceptance + plan-auditor PASS

## 4. Technical Approach

### 4.1 Confidence calculation 의사 코드

```
confidence = 0.0
if pr.has_spec_id_reference:    confidence += 0.30
if pr.all_commits_have_moai_coauthor: confidence += 0.30
if pr.commit_messages_reference_spec: confidence += 0.20
if pr.title_has_conventional_commit_type: confidence += 0.10
if pr.has_reviewer_approval: confidence += 0.10

confidence = min(confidence, 1.0)

if confidence < 0.70:
    claude_assisted_lines = 0  # conservative bias
else:
    claude_assisted_lines = pr.lines_changed (단, SPEC commit 기반 계산)
```

### 4.2 Schema 예시 (3 entries)

```jsonl
{"pr_id":741,"author":"GOOS","claude_assisted_lines":1234,"total_lines":1500,"confidence":0.85,"spec_ids":["SPEC-CRON-PATTERN-001"],"timestamp":"2026-04-30T12:00:00Z","merge_strategy":"merge","branch_type":"feature"}
{"pr_id":742,"author":"GOOS","claude_assisted_lines":0,"total_lines":50,"confidence":0.40,"spec_ids":[],"timestamp":"2026-04-30T13:00:00Z","merge_strategy":"squash","branch_type":"fix"}
{"pr_id":743,"author":"unknown","claude_assisted_lines":120,"total_lines":150,"confidence":0.80,"spec_ids":["SPEC-METRICS-001"],"timestamp":"2026-04-30T14:00:00Z","merge_strategy":"squash","branch_type":"feature"}
```

### 4.3 Privacy Detection 정규식 (예시)

- email: `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`
- API key: `(sk|api|key|token)[_-]?[A-Za-z0-9]{20,}`
- 검출 시 entry 전체 skip

## 5. Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| Metric 측정의 침해성 우려 | Medium | High | opt-in default (사용자 명시 활성화) |
| Confidence threshold 부정확 | High | Medium | conservative bias (< 0.70 → 0), false positive 회피 |
| `.moai/metrics/contribution.jsonl` 누적 | Medium | Low | 사용자 책임 명시 (cleanup 자동화 X) |
| GitHub Analytics 통합 복잡 | High | Medium | Phase 2로 분리 |
| 개인 정보 (author email) 영속화 | Medium | High | privacy 5 rules + regex detection |
| `metrics.yaml` schema migration 필요 | Low | Low | 단일 신규 파일, 기존 config 변경 X |

## 6. Dependencies

- 선행 SPEC: 없음 (standalone)
- 의존 입력: `/moai sync` 워크플로우, GitHub PR API
- sibling SPEC: SPEC-CRON-PATTERN-001 (Pattern P4 = CI Failure 입력)
- 도구: `make build`, plan-auditor, `gh pr view` (manual)

## 7. Open Questions Resolution

- **OQ1** (opt-in vs opt-out): opt-in default (M3에서 결정 — privacy first)
- **OQ2** (GitHub Analytics 통합 우선순위): Phase 2 후속 SPEC (현재 SPEC scope 외)
- **OQ3** (confidence calculation 정확한 정의): M2에서 5-factor 휴리스틱 명시
- **OQ4** (`moai metrics show` CLI 본 SPEC scope): 본 SPEC은 데이터 수집만, CLI는 Phase 2
- **OQ5** (retention 정책): 사용자 책임 (자동 cleanup X)

## 8. Rollout Plan

1. M1-M6 구현 후 본 프로젝트의 다음 sync에 dogfooding
2. 5 sync 호출 측정 → confidence 분포 확인
3. False positive 0% 확인 후 v2.x.0 minor release
4. Phase 2 SPEC: `moai metrics show` CLI + GitHub Analytics 통합 (별도 SPEC)

End of plan.md (SPEC-METRICS-001).
