# Research — SPEC-METRICS-001 (Contribution Metrics 통합)

**SPEC**: SPEC-METRICS-001
**Wave**: 4 / Tier 3 (장기/폴리싱)
**Created**: 2026-04-30
**Author**: manager-spec

---

## 1. 출처 (Anthropic 공식 자료)

**Source**: Anthropic blog "Claude Code Contribution Metrics"
**URL**: https://claude.com/blog/contribution-metrics
**Accessed**: 2026-04-30 (verified via WebFetch)

### 1.1 Verbatim 인용 (§ "How we're shipping at Anthropic" + § "Measure velocity with Claude Code")

> "While pull requests alone are an incomplete measure of developer velocity, we've found them to be a close proxy for what engineering teams care about: shipping features, fixing bugs, and delighting users faster."

— Section: "How we're shipping at Anthropic"

> "67% increase in PRs merged per engineer per day. Across teams, 70–90% of code is now being written with Claude Code assistance."

— Section: "How we're shipping at Anthropic"

> "We calculate this conservatively, and only code where we have high confidence in Claude Code's involvement is counted as assisted."

— Section: "Measure velocity with Claude Code"

### 1.2 Anthropic의 측정 철학

- **Conservative measurement**: 확신이 있을 때만 count → false positive 회피
- **PR as proxy**: PR-level metric이 commit-level보다 noise가 적음
- **Confidence threshold**: 명시적 threshold (예: 0.70) 활용
- **Aggregation**: 개인 metric vs team metric 분리 (개인은 개인 GitHub view)

### 1.3 핵심 metric 종류

- PRs merged per engineer per day (rate)
- Claude-assisted line ratio (%)
- Time-to-merge (latency)
- Reviewers per PR (collaboration metric)

---

## 2. 현재 상태 (As-Is)

### 2.1 moai-adk-go의 Quality vs Contribution metric

기존 quality metric:
- LSP error/warning count (`.moai/config/sections/quality.yaml` thresholds)
- TRUST 5 score (manager-quality)
- evaluator-active 4-dimension score (Functionality / Security / Craft / Consistency)
- Test coverage (per-language toolchain)

부재 항목 (Contribution metric):
- PR-level Claude Code 기여도 추적 ❌
- 시간별 PR 처리 속도 ❌
- 개인/팀 contribution dashboard ❌
- GitHub Analytics 통합 ❌

### 2.2 운영 격차

| 영역 | 현재 (As-Is) | 목표 (To-Be) | 격차 |
|------|--------------|--------------|------|
| PR 기여도 추적 | 부재 | sync 단계에서 자동 수집 | 신규 워크플로우 |
| Claude-assisted line 측정 | 부재 | conservative threshold 0.70 | 신규 정책 |
| 통계 영속화 | 부재 | `.moai/metrics/contribution.jsonl` | 신규 파일 |
| Dashboard | 부재 | `moai metrics show` CLI | 신규 명령어 |
| GitHub Analytics 연동 | 부재 | 옵션 (Team/Enterprise) | 신규 통합 |

### 2.3 metric entry schema (안)

```jsonl
{
  "pr_id": 741,
  "author": "GOOS",
  "claude_assisted_lines": 1234,
  "total_lines": 1500,
  "confidence": 0.85,
  "spec_ids": ["SPEC-XXX-001"],
  "timestamp": "2026-04-30T12:00:00Z",
  "merge_strategy": "merge",
  "branch_type": "feature"
}
```

---

## 3. 코드베이스 분석 (Affected Files)

### 3.1 Primary 신규 대상

| 파일 | 수정 유형 | 변경 사유 |
|------|-----------|-----------|
| `.moai/metrics/contribution.jsonl` | 신규 (런타임 생성) | Contribution metric 영속화 |
| `.claude/rules/moai/workflow/contribution-metrics.md` | 신규 | 측정 정책 문서 |
| `.claude/skills/moai-workflow-project/modules/contribution-metrics.md` | 신규 | dashboard 사용 가이드 |

### 3.2 Secondary 수정

| 파일 | 수정 유형 | 변경 사유 |
|------|-----------|-----------|
| `.claude/skills/moai-workflow-project/SKILL.md` | cross-ref + sync 단계 보강 | metric 자동 수집 |
| `internal/cli/metrics.go` | 신규 (conditional) | `moai metrics show` 명령어 (Phase 2) |
| `internal/template/templates/.claude/rules/moai/workflow/contribution-metrics.md` | 신규 | Template-First |

### 3.3 Conservative measurement 정책

Confidence 0.70 threshold criteria (Anthropic 권고 적용):
- PR이 SPEC-ID와 명시적으로 연결됨 (commit message 또는 PR body 인용)
- 모든 commit이 `🗿 MoAI` co-author 포함
- AI-assisted lines = SPEC commit으로부터 직접 파생된 line
- Confidence < 0.70 → assisted lines = 0 (conservative)

---

## 4. 위험 및 가정

### 4.1 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|-----------|
| Metric 측정이 침해성 (사용자 추적) | Medium | High | opt-in 정책, `.moai/config/sections/metrics.yaml metrics.enabled: false` 기본값 검토 |
| Confidence threshold 부정확 | High | Medium | conservative bias (0.70+), false positive 회피 |
| `.moai/metrics/contribution.jsonl` 누적 | Medium | Low | retention 정책 (예: 365일) |
| GitHub Analytics 통합 복잡 | High | Medium | Phase 2로 분리, Phase 1은 local JSONL만 |
| 개인 정보 (author email) 영속화 | Medium | High | author는 GitHub username만, email 제외 |

### 4.2 Assumptions

- A1: 사용자가 `/moai sync`를 통해 PR을 일관되게 생성
- A2: PR body에 SPEC-ID 인용 관례 정착
- A3: `🗿 MoAI` co-author가 신뢰할 수 있는 신호
- A4: `moai metrics show`는 v2.x.0 후속 SPEC (현재 SPEC은 데이터 수집만)
- A5: GitHub Analytics 통합은 옵션 (모든 사용자가 Team/Enterprise는 아님)

---

## 5. 측정 계획 (Baseline + Validation)

| Metric | Baseline 측정 방법 | 목표 |
|--------|-------------------|------|
| metric entry 생성 | sample 5 sync 호출 | 100% 생성 |
| schema 정확성 | JSONL parse 테스트 | 0 error |
| Confidence threshold | sample 5 측정 | 모든 < 0.70 → assisted = 0 |
| author 익명화 | email 노출 검증 | 0 leak |
| `.moai/metrics/contribution.jsonl` 누적 | 5 entries 후 검증 | 5 line |
| 정책 문서 | 문서 검토 | 모든 절 작성 |

---

## 6. 대안 검토 (Alternatives Considered)

| 대안 | 채택 여부 | 이유 |
|------|-----------|------|
| commit-level metric | ❌ | noise가 큼, Anthropic도 PR-level 권고 |
| Claude API 직접 호출 (real-time) | ❌ | offline 작업 불가 |
| Prometheus/Grafana 통합 | ❌ | overkill, JSONL이 충분 |
| `~/.claude/metrics/` 위치 | ❌ | project-scope (SPEC-ID 인용)이므로 `.moai/metrics/` |
| commit message에 metadata 임베드 | ❌ | commit message 오염, separate log가 더 깨끗 |
| author email 영속화 | ❌ | privacy risk |

---

## 7. 참고 SPEC

- SPEC-OBSERVE-001 (Wave 3): observability 정책 — 본 SPEC의 metric 일종
- SPEC-CRON-PATTERN-001 (이번 wave sibling): Pattern P4 (CI failure)에서 metric 활용 가능
- SPEC-V3R2-WF-001: 워크플로우 표준 — sync 단계에 metric 수집 통합

---

## 8. Open Questions (Plan 단계 해결 대상)

- OQ1: opt-in vs opt-out 정책 (privacy 우려) → plan.md에서 결정
- OQ2: GitHub Analytics 통합 우선순위? → Phase 2로 분리
- OQ3: confidence calculation 알고리즘 정확한 정의? → plan.md에서 fully spec
- OQ4: `moai metrics show` CLI는 본 SPEC scope인가? → 본 SPEC은 데이터 수집만, CLI는 후속
- OQ5: retention 정책 (365일?) → plan.md에서 결정

---

End of research.md (SPEC-METRICS-001).
