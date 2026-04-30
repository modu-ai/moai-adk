---
id: SPEC-METRICS-001
status: draft
version: "0.1.0"
priority: Medium
labels: [metrics, contribution, sync, github-analytics, observability, wave-4, tier-3]
issue_number: null
scope: [.moai/metrics, .claude/rules/moai/workflow, .claude/skills/moai-workflow-project, internal/cli]
blockedBy: []
dependents: []
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
wave: 4
tier: 3
---

# SPEC-METRICS-001: Contribution Metrics 통합

## HISTORY

- 2026-04-30 v0.1.0: 최초 작성. Wave 4 / Tier 3. Anthropic "Claude Code Contribution Metrics" 권고를 본 프로젝트의 sync 워크플로우에 통합. PR-level conservative measurement.

---

## 1. Goal (목적)

본 프로젝트는 quality metric (LSP, TRUST 5, evaluator-active)은 보유하나 **contribution metric** (PR/commit별 Claude Code 기여도)이 부재하다. 본 SPEC은 `/moai sync` 워크플로우에 PR-level metric 자동 수집 단계를 추가하고, conservative measurement (confidence threshold 0.70)로 false positive를 회피한다.

### 1.1 배경

- Anthropic blog "Claude Code Contribution Metrics": "We calculate this conservatively, and only code where we have high confidence in Claude Code's involvement is counted as assisted."
- 본 프로젝트의 sync 워크플로우는 PR 생성을 표준화 → metric 수집 hook point로 적합
- PR이 commit-level보다 noise가 적은 proxy

### 1.2 비목표 (Non-Goals)

- `moai metrics show` CLI 명령어 (Phase 2 후속 SPEC)
- GitHub Analytics 통합 (Phase 2 후속 SPEC, 옵션 단계)
- Real-time metric dashboard
- author email 영속화 (privacy)
- commit-level metric (PR-level만)
- Confidence calculation 자동화 (현재 휴리스틱)

---

## 2. Scope (범위)

### 2.1 In Scope

- `.moai/metrics/contribution.jsonl` 영속화 schema 정의
- `/moai sync` 워크플로우에 metric 수집 단계 추가
- Confidence threshold 0.70 (conservative)
- Metric entry schema (pr_id, author, claude_assisted_lines, total_lines, confidence, spec_ids, timestamp, merge_strategy, branch_type)
- `.claude/rules/moai/workflow/contribution-metrics.md` 정책 문서
- `moai-workflow-project` SKILL.md cross-ref
- Privacy 정책 (author = GitHub username, email 제외)
- Opt-in 정책 검토 (`.moai/config.yaml metrics.enabled`)
- Template-First 동기화

### 2.2 Exclusions (What NOT to Build)

- `moai metrics show` CLI 명령어 (Phase 2)
- GitHub Analytics push 통합 (Phase 2)
- Real-time dashboard (Grafana, Prometheus 등)
- commit-level metric
- author email 영속화
- 자동 retention cleanup
- Real-time API 호출 비용 측정
- 팀 간 비교 metric (개인 metric만)

---

## 3. Environment (환경)

- 런타임: `/moai sync` 워크플로우 실행 시
- Storage: `.moai/metrics/contribution.jsonl` (append-only JSONL)
- 영향 파일: `.moai/metrics/`, `.claude/rules/moai/workflow/`, `.claude/skills/moai-workflow-project/`
- Privacy 가정: author = GitHub username 형식

---

## 4. Assumptions (가정)

- A1: 사용자가 `/moai sync`를 통해 PR을 일관되게 생성
- A2: PR body에 SPEC-ID 인용 관례 정착
- A3: `🗿 MoAI` co-author signal이 신뢰 가능
- A4: `moai metrics show` CLI는 Phase 2로 분리
- A5: GitHub Analytics 통합은 옵션 (Team/Enterprise 사용자만)
- A6: Opt-in default (사용자 명시 활성화) 또는 opt-out — plan.md에서 결정

---

## 5. Requirements (EARS Format)

### 5.1 Ubiquitous Requirements

- **REQ-METRICS-001**: THE FILE `.moai/metrics/contribution.jsonl` SHALL be the canonical storage for PR-level contribution metrics.
- **REQ-METRICS-002**: EACH METRIC ENTRY SHALL include: pr_id, author, claude_assisted_lines, total_lines, confidence, spec_ids, timestamp, merge_strategy, branch_type.
- **REQ-METRICS-003**: THE CONFIDENCE THRESHOLD for counting lines as assisted SHALL be 0.70 (conservative measurement per Anthropic).
- **REQ-METRICS-004**: THE FILE `.claude/rules/moai/workflow/contribution-metrics.md` SHALL document the measurement policy, schema, and confidence calculation heuristic.

### 5.2 Event-Driven Requirements

- **REQ-METRICS-005**: WHEN `/moai sync` completes a PR creation, THE WORKFLOW SHALL append a contribution metric entry to `.moai/metrics/contribution.jsonl`.
- **REQ-METRICS-006**: WHEN computing confidence, THE WORKFLOW SHALL evaluate: (a) PR-SPEC linkage explicit, (b) `🗿 MoAI` co-author present, (c) commit messages reference SPEC-ID.
- **REQ-METRICS-007**: WHEN confidence is below 0.70, THE WORKFLOW SHALL set `claude_assisted_lines = 0` (conservative bias).

### 5.3 State-Driven Requirements

- **REQ-METRICS-008**: WHILE `.moai/config.yaml metrics.enabled` is `false`, THE WORKFLOW SHALL skip metric collection entirely.

### 5.4 Conditional (WHERE / IF) Requirements

- **REQ-METRICS-009**: WHERE `.moai/metrics/` directory does not exist, THE FIRST METRIC WRITE SHALL create the directory automatically.
- **REQ-METRICS-010**: WHERE author identification is not available from PR metadata, THE AUTHOR FIELD SHALL be set to `"unknown"` (no email leak).
- **REQ-METRICS-011**: IF the PR has no SPEC-ID linkage, THE CONFIDENCE SHALL be capped at 0.50 (below threshold, no assisted attribution).
- **REQ-METRICS-012**: WHERE GitHub Analytics integration is enabled (Phase 2), THE WORKFLOW SHALL push aggregated metrics — until then, local JSONL only.

### 5.5 Unwanted (Negative) Requirements

- **REQ-METRICS-013**: THE METRIC ENTRY SHALL NOT include author email addresses.
- **REQ-METRICS-014**: THE WORKFLOW SHALL NOT count PRs without `🗿 MoAI` co-author as assisted (conservative).
- **REQ-METRICS-015**: THE WORKFLOW SHALL NOT block sync execution on metric collection failure (best-effort).
- **REQ-METRICS-016**: THE METRIC FILE SHALL NOT be auto-cleaned (retention is user responsibility).

---

## 6. Success Criteria (성공 기준)

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| 정책 문서 존재 | file existence | EXISTS |
| Schema 정의 | 9 필드 명시 | 100% |
| Confidence threshold | 0.70 명시 | EXISTS |
| Privacy 정책 | email 제외 명시 | EXISTS |
| Opt-in/out 정책 | `.moai/config.yaml` reference | EXISTS |
| Sample 5 sync 호출 | metric 생성 검증 | 100% |
| Confidence < 0.70 → assisted = 0 | sample 검증 | 100% |
| Cross-ref | grep | >= 2 |
| Template-First sync | `make build` diff | clean |

---

## 7. Acceptance References

See `acceptance.md` for Given-When-Then scenarios and Definition of Done.

---

## 8. Constraints

- C1: 본 SPEC은 데이터 수집 + 정책만, dashboard CLI는 Phase 2
- C2: `moai metrics show`는 후속 SPEC
- C3: GitHub Analytics 통합은 Phase 2 옵션
- C4: author email 영속화 금지 (GDPR / privacy)
- C5: Template-First Rule 준수

End of spec.md (SPEC-METRICS-001 v0.1.0).
