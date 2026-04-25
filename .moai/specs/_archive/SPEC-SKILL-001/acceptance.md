---
spec_id: SPEC-SKILL-001
version: 1.0.0
status: backfilled
created_at: 2026-04-24
author: manager-spec (backfill)
backfill_reason: |
  UNDERSPECIFIED — recommend archive or full rewrite.
  spec.md 가 892 바이트로 "구현 요약" 수준만 기술되어 있으며, EARS REQ 및 acceptance 시나리오가
  없음. 본 backfill 은 실제 shipped 구현 (commit beba31b7c, PR #547) 에서 관찰된 동작만
  AC-OBSERVED 형태로 기록한다. REQ-derived AC 는 도출 불가.
---

# Acceptance Criteria — SPEC-SKILL-001 (effort frontmatter + worktree sparsePaths)

## Status Note

[UNDERSPECIFIED] 원본 spec.md 는 REQ / Given-When-Then / Exclusions 섹션이 부재한 "구현 요약 메모"
수준이다. 따라서 본 acceptance.md 는 observed-behavior 기반으로만 작성된다.
향후 본 SPEC 을 유지하려면 EARS 형식 REQ 로 전체 재작성을 권장한다.

## Traceability

| Source (spec.md 구현 요약)                        | AC ID        | Observed Artifact                                                                     |
|---------------------------------------------------|--------------|---------------------------------------------------------------------------------------|
| moai-workflow-thinking effort: high              | AC-OBS-001   | `.claude/skills/moai-workflow-thinking/SKILL.md:10`                                   |
| moai-foundation-philosopher effort: high         | AC-OBS-002   | `.claude/skills/moai-foundation-philosopher/SKILL.md:10`                              |
| moai-workflow-loop effort: low                   | AC-OBS-003   | `.claude/skills/moai-workflow-loop/SKILL.md:11`                                       |
| skill-authoring.md effort 문서화                  | AC-OBS-004   | `.claude/rules/moai/development/skill-authoring.md:25, 79`                            |
| workflow.yaml worktree + sparse_paths 섹션       | AC-OBS-005   | `.moai/config/sections/workflow.yaml:110-121`                                         |

## AC-OBS-001: moai-workflow-thinking 스킬에 effort: high

- **Observed**: `internal/template/templates/.claude/skills/moai-workflow-thinking/SKILL.md` 의
  frontmatter 10번째 라인에 `effort: high` 가 선언되어 있음
- **Rationale (commit message)**: Deep analysis 스킬로 분류되어 xhigh/high 추론 예산 부여
- **Verification**: `grep "^effort:" .claude/skills/moai-workflow-thinking/SKILL.md`

## AC-OBS-002: moai-foundation-philosopher 스킬에 effort: high

- **Observed**: `internal/template/templates/.claude/skills/moai-foundation-philosopher/SKILL.md` 의
  frontmatter 에 `effort: high` 가 선언되어 있음
- **Rationale (commit message)**: 전략적 의사결정 프레임워크로서 심층 추론 필요
- **Verification**: `grep "^effort:" .claude/skills/moai-foundation-philosopher/SKILL.md`

## AC-OBS-003: moai-workflow-loop 스킬에 effort: low

- **Observed**: `internal/template/templates/.claude/skills/moai-workflow-loop/SKILL.md` 의
  frontmatter 에 `effort: low` 가 선언되어 있음
- **Rationale (commit message)**: 반복 실행 루프로 속도 중시, 추론 깊이 불필요
- **Verification**: `grep "^effort:" .claude/skills/moai-workflow-loop/SKILL.md`

## AC-OBS-004: skill-authoring.md 에 effort 필드 문서화

- **Observed**: `internal/template/templates/.claude/rules/moai/development/skill-authoring.md` 에
  effort 필드가 문서화되어 있음 (line 25: 유효 값 low/medium/high/xhigh/max;
  line 79: 사용 예시)
- **Verification**: `grep -n "effort" .claude/rules/moai/development/skill-authoring.md`

## AC-OBS-005: workflow.yaml 에 worktree + sparse_paths 섹션 존재

- **Observed**: `internal/template/templates/.moai/config/sections/workflow.yaml:110-121` 에
  `worktree:` 섹션이 존재하며 다음 키가 선언됨:
  - `auto_create: true`
  - `auto_merge: true`
  - `auto_cleanup: true`
  - `tmux_preferred: true`
  - `session_name_pattern: "moai-{ProjectName}-{SPEC-ID}"`
  - `sparse_paths:` 주석 처리된 예시 (사용자가 필요 시 활성화)
- **Observed Rationale**: Claude Code v2.1.76+ sparse checkout 기능을 monorepo 대응용으로
  템플릿에 선택적 샘플로 제공 (default 비활성)
- **Verification**: `sed -n '110,121p' .moai/config/sections/workflow.yaml`

## Exclusions

spec.md 에 exclusions 섹션이 없어 원본 스펙에서 도출 불가. 관찰된 구현 경계:

- sparse_paths 는 기본값에서 주석 처리 (opt-in) — 기존 워크트리 동작을 회귀시키지 않음
- effort 필드는 각 스킬 고유이며 session-level effort 를 오버라이드하는 용도

## Quality Gate Criteria (Observed only)

- [x] 3 개 스킬 frontmatter 에 effort 필드 적용
- [x] skill-authoring rule 문서 업데이트
- [x] workflow.yaml 에 worktree.sparse_paths 샘플 배포
- [x] 기존 프로젝트 회귀 없음 (sparse_paths opt-in, effort 필드는 Claude Code 2.1.80+ 만 인식)

## Definition of Done (Historical)

- [x] PR #547 merged (commit beba31b7c, 2026-03-20)
- [x] 3 개 SPEC 일괄 구현 (SPEC-HOOK-009 + SPEC-STATUSLINE-002 + SPEC-SKILL-001)
- [x] backfill acceptance.md (본 문서)

## Recommended Next Action

본 SPEC 은 아카이브 또는 EARS 재작성을 강하게 권장한다. 이유:

1. spec.md 가 구현 메모이지 스펙이 아님 — "plan" 이 없고 REQ 가 없음
2. 관련 기능이 각각 Claude Code 버전 의존 (v2.1.76 / v2.1.80) 이므로 독립 SPEC 2 개
   (SPEC-EFFORT-FRONTMATTER-001, SPEC-SPARSE-PATHS-001) 로 분리 재작성 권장
3. 현 상태에서는 변경 영향 분석 및 회귀 테스트 작성 기준이 부재
