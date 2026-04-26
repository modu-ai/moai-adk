---
spec_id: SPEC-V3R3-DEF-007
title: Implementation Plan — Convention Compliance Sweep
version: "1.0.0"
status: draft
created: 2026-04-25
updated: 2026-04-25
author: manager-spec
related_spec: .moai/specs/SPEC-V3R3-DEF-007/spec.md
---

# Plan — SPEC-V3R3-DEF-007

## 1. Objectives

- 11개 skill SKILL.md frontmatter에 `progressive_disclosure` 블록 추가 (template + local 양쪽)
- `manager-git` agent body에 `Scope Boundaries`, `Delegation Protocol` 섹션 추가
- `make build`로 embedded.go 갱신 검증
- Phase A 후속 SPEC의 baseline 확립

## 2. Technical Approach

### 2.1 Frontmatter Insertion Pattern

각 SKILL.md의 metadata 블록 직후에 다음 블록 삽입:

```yaml
# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000
```

위치 결정 근거: `.claude/skills/moai-workflow-spec/SKILL.md` line 22-26 패턴 준수.

### 2.2 Agent Body Insertion Pattern

`manager-git.md` body의 `## Primary Mission` 또는 `## Core Capabilities` 섹션 직후에 다음 두 섹션 삽입 (manager-spec body 패턴 참조):

```markdown
## Scope Boundaries

IN SCOPE: <git workflow operations 명시>
OUT OF SCOPE: <SPEC writing, code implementation, doc sync>

## Delegation Protocol

- SPEC creation/modification: Delegate to manager-spec
- Code implementation: Delegate to manager-ddd / manager-tdd
- Documentation sync: Delegate to manager-docs
- Quality validation: Delegate to manager-quality
```

### 2.3 Template Synchronization

각 파일 수정은 template과 local에 동시 적용 (Edit tool 2회 호출). 마지막에 `make build` 실행으로 `internal/template/embedded.go` 재생성.

## 3. Wave / Phase 설계

본 SPEC은 단일 Wave로 진행 (총 task 수 적음, 순차 실행 가능).

### Wave A.1 — Skill Frontmatter Sweep

11개 skill 각각에 progressive_disclosure 블록 추가 (template + local pair, 총 22 file edit).

### Wave A.2 — Agent Body Update

manager-git.md에 Scope Boundaries + Delegation Protocol 섹션 추가 (template + local pair, 총 2 file edit).

### Wave A.3 — Build Verification

`make build` 실행, `go test ./internal/template/...` 검증.

## 4. File 영향 요약

### Skill 파일 (11 × 2 = 22개)

| Skill | Local Path | Template Path |
|-------|------------|---------------|
| moai-domain-backend | `.claude/skills/moai-domain-backend/SKILL.md` | `internal/template/templates/.claude/skills/moai-domain-backend/SKILL.md` |
| moai-domain-frontend | `.claude/skills/moai-domain-frontend/SKILL.md` | `internal/template/templates/.claude/skills/moai-domain-frontend/SKILL.md` |
| moai-domain-db-docs | `.claude/skills/moai-domain-db-docs/SKILL.md` | `internal/template/templates/.claude/skills/moai-domain-db-docs/SKILL.md` |
| moai-formats-data | `.claude/skills/moai-formats-data/SKILL.md` | `internal/template/templates/.claude/skills/moai-formats-data/SKILL.md` |
| moai-framework-electron | `.claude/skills/moai-framework-electron/SKILL.md` | `internal/template/templates/.claude/skills/moai-framework-electron/SKILL.md` |
| moai-library-mermaid | `.claude/skills/moai-library-mermaid/SKILL.md` | `internal/template/templates/.claude/skills/moai-library-mermaid/SKILL.md` |
| moai-library-nextra | `.claude/skills/moai-library-nextra/SKILL.md` | `internal/template/templates/.claude/skills/moai-library-nextra/SKILL.md` |
| moai-library-shadcn | `.claude/skills/moai-library-shadcn/SKILL.md` | `internal/template/templates/.claude/skills/moai-library-shadcn/SKILL.md` |
| moai-tool-ast-grep | `.claude/skills/moai-tool-ast-grep/SKILL.md` | `internal/template/templates/.claude/skills/moai-tool-ast-grep/SKILL.md` |
| moai-workflow-ddd | `.claude/skills/moai-workflow-ddd/SKILL.md` | `internal/template/templates/.claude/skills/moai-workflow-ddd/SKILL.md` |
| moai-workflow-loop | `.claude/skills/moai-workflow-loop/SKILL.md` | `internal/template/templates/.claude/skills/moai-workflow-loop/SKILL.md` |

### Agent 파일 (1 × 2 = 2개)

- `.claude/agents/moai/manager-git.md`
- `internal/template/templates/.claude/agents/moai/manager-git.md`

### 빌드 파일 (자동 갱신)

- `internal/template/embedded.go` (`make build` 결과)

## 5. Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| YAML 들여쓰기 실수로 파싱 실패 | High | Edit tool로 정확한 들여쓰기 보존; sample 검증 |
| manager-git body 기존 내용과 신규 섹션 중복 | Low | 사전 grep으로 중복 keyword 확인 후 삽입 |
| make build 실패 | Medium | 변경 직후 즉시 빌드 실행, 오류 시 즉시 수정 |
| Template-local drift | Medium | 각 파일은 동일 turn에 두 위치 모두 수정 |

## 6. Open Questions

- OQ1: progressive_disclosure 블록 삽입 위치 — metadata 직후 vs frontmatter 마지막? → **Decision**: metadata 직후 (moai-workflow-spec 패턴 준수)
- OQ2: manager-git의 기존 워크플로우 설명과 신규 Scope Boundaries 중복 시 → **Decision**: 신규 섹션은 요약 (3-5 lines), 기존 본문은 보존

## 7. Milestones

- M1: 11개 skill frontmatter 일괄 추가 완료
- M2: manager-git body 섹션 추가 완료
- M3: make build + go test 통과
- M4: 변경 사항 git diff 검증 후 commit-ready 상태

## 8. Definition of Done

- [ ] 22개 SKILL.md 파일 모두 progressive_disclosure 블록 포함
- [ ] 2개 manager-git.md 파일 모두 Scope Boundaries + Delegation Protocol 섹션 포함
- [ ] `make build` 정상 종료
- [ ] `go test ./internal/template/...` 통과
- [ ] git diff 결과가 frontmatter/본문 추가만 포함 (다른 변경 없음)
- [ ] AC-DEF007-01 ~ 05 모두 만족
