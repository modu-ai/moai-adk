# Context — MoAI-ADK Catalog Slimming Proposal

## Background

MoAI-ADK는 Claude Code 위에서 SPEC-First DDD/TDD 워크플로우를 제공하는 오픈소스 메타 시스템입니다. 현재 v2.x 운영 중이며 다음 마이너 릴리스 (v2.20 또는 v3.0) 에서 카탈로그 구조 변경을 계획하고 있습니다.

### 현재 상태 (2026-05-12 기준)

- **Skills**: 36개 (foundation 4 / workflow 13 / domain 11 / platform 3 / framework 1 / meta 1 / ref 4)
- **Agents**: 29개 (manager-* 8 / expert-* 9 / builder-* 4 / evaluator-active / plan-auditor / researcher / claude-code-guide / manager-brain)
- **Harness 인프라**: 이미 존재 (`moai-meta-harness`, `moai-harness-learner`, `builder-harness`)
- **Anthropic plugin marketplace**: 2026 표준 인프라 (이미 공식 도입됨)
- **Context budget**: Opus 4.7 1M context의 1% = 10K tokens가 skill description 예산

### 문제 정의

1. **모든 36 skills + 29 agents가 `moai init` 시 deploy** — 사용 안 하는 자산도 모두 (예: Chrome-extension 프로젝트가 아닌데 `moai-platform-chrome-extension` 로드)
2. **Context budget overflow 위험** — 사용자가 marketplace plugin 추가 시 즉시 1% 한도 초과 → MoAI 자체 skill description 일부 drop → trigger 정확도 저하
3. **Harness 인프라 미활용** — `moai-meta-harness`가 프로젝트별 동적 skill/agent 생성 능력을 이미 갖췄지만, 정적 deploy가 이를 대체
4. **`moai update` 위험 증가** — 카탈로그가 클수록 update 시 충돌과 사용자 변경 손실 가능성 증가

## User Answers (Phase 1 Discovery)

`/moai brain` 워크플로우 Phase 1 Discovery 라운드에서 사용자가 4가지 핵심 결정에 답변:

| 질문 | 답변 |
|------|------|
| 코어 분류 기준 | Workflow + Foundation 필수만 (~18개) |
| harness 위상 | `/moai project`에서 자동 제안, opt-out |
| 정리 범위 | skills + agents 둘 다 정리 |
| 마이그레이션 | Non-breaking 신규 설치 + **`moai update` 시 기존 프로젝트도 안전하게 catalog 동기화 (재설계 필요)** |

사용자의 추가 요구: "문제가 없도록 다시 설계해서 보고" — `moai update` 안전성이 최우선 제약.

## Proposed Solution Summary

**"3-tier Catalog + cruft-style Safe Sync"**

```
Tier 1 — Core (always deployed)         : 18 skills + 15 agents
Tier 2 — Optional Packs (opt-in)        : moai pack add <name>
Tier 3 — Harness-Generated (per-project): /moai project 인터뷰 opt-out
```

Safety: `moai update --catalog-sync` 가 snapshot → drift 감지 → 3-way merge → 사용자 확인 → rollback 인프라로 구성.

## Constitutional Constraints (HARD)

검토자가 인지해야 할 MoAI 헌법 (변경 불가):

- **AskUserQuestion 독점**: 모든 user-facing 질문은 AskUserQuestion 경유 (산문 질문 금지)
- **16개 언어 중립성**: `internal/template/templates/` 는 16개 프로그래밍 언어 동등 취급
- **Template-First**: 모든 신규 자산은 먼저 `internal/template/templates/` 에 추가
- **No tech-stack assumptions in brain**: proposal은 능력 (capability) 만 기술, 구현은 후속 SPEC
- **PR-based migration**: 모든 변경은 Conventional Commits + PR 경유, force push 금지

## SPEC Decomposition (proposal.md 참조)

7개 SPEC으로 분해:
- SPEC-V3R4-CATALOG-001: manifest 스키마
- SPEC-V3R4-CATALOG-002: 디렉토리 재배치
- SPEC-V3R4-CATALOG-003: moai pack 명령
- SPEC-V3R4-CATALOG-004: moai update 안전 동기화 ⚠️ **highest risk**
- SPEC-V3R4-CATALOG-005: /moai project 인터뷰 확장
- SPEC-V3R4-CATALOG-006: moai doctor catalog
- SPEC-V3R4-CATALOG-007: 마이그레이션 docs 4개국어

권장 실행 순서: Wave 1 (001+002) → Wave 2 (003+005) → Wave 3 (004) → Wave 4 (006+007)

## Why This Matters

이 결정은 **MoAI-ADK가 Claude Code 생태계 표준에 정렬되어 장기적으로 확장 가능한 distribution architecture로 전환**하는 분기점입니다. 잘못된 결정 시:

- 사용자 손실 (moai update 충돌)
- 카탈로그 분할로 인한 유지보수 부담 증가 (마이그레이션 미흡)
- 코어가 너무 좁아 신규 사용자 첫 사용 마찰 증가

이를 방지하기 위한 독립 review를 요청합니다.
