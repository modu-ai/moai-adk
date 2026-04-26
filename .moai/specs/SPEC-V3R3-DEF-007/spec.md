---
id: SPEC-V3R3-DEF-007
title: Convention Compliance Sweep — manager-git body + 11 skills progressive_disclosure
version: "1.0.0"
status: implemented
created: 2026-04-25
updated: 2026-04-26
author: manager-spec
priority: P0 Critical
phase: "v3.0.0 R3 — Phase A — Convention Compliance Baseline"
module: ".claude/skills/, .claude/agents/moai/, internal/template/templates/.claude/"
dependencies: []
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "convention, compliance, baseline, skill-frontmatter, agent-body, template-sync, v3r3, phase-a"
related_theme: "Phase A — Defect Cleanup"
released_in: v2.15.0
---

# SPEC-V3R3-DEF-007: Convention Compliance Sweep

## HISTORY

| Version | Date       | Author       | Description                                                                                                  |
|---------|------------|--------------|--------------------------------------------------------------------------------------------------------------|
| 1.0.0   | 2026-04-25 | manager-spec | Initial draft. Phase A baseline P0 — establishes convention compliance for 11 skills + manager-git agent body. |

---

## 1. Goal (목적)

MoAI-ADK v3.0.0 R3 Phase A의 모든 후속 SPEC이 의존하는 **convention compliance baseline**을 확립한다. 두 가지 누락을 동시에 해소한다:

1. 11개 skill의 SKILL.md frontmatter에 `progressive_disclosure` 블록 누락 → 다른 skill들과 일관된 token-budget 메타데이터 제공
2. `manager-git` agent body에 `Scope Boundaries`, `Delegation Protocol` 섹션 누락 → 다른 22개 agent와 동일한 구조적 일관성 회복

본 SPEC은 additive(기능 추가 없음, 메타데이터/문서만 보강)이며 후속 SPEC(DEF-001, ARCH-003, ARCH-007, COV-001)의 baseline으로 작동한다.

### 1.1 배경

- Progressive Disclosure 시스템은 SKILL.md frontmatter에 `progressive_disclosure: { enabled, level1_tokens, level2_tokens }` 블록을 통해 token budget을 명시한다 (`.claude/rules/moai/workflow/spec-workflow.md` §Progressive Disclosure 참조).
- v3.0.0 시점에 다수 skill에 해당 블록이 추가되었으나 다음 11개는 누락 상태로 남아 있다:
  - moai-domain-backend, moai-domain-frontend, moai-domain-db-docs, moai-formats-data,
  - moai-framework-electron, moai-library-mermaid, moai-library-nextra, moai-library-shadcn,
  - moai-tool-ast-grep, moai-workflow-ddd, moai-workflow-loop
- 22개 agent 중 21개는 `Scope Boundaries` (IN SCOPE / OUT OF SCOPE) 및 `Delegation Protocol` 섹션을 본문에 포함한다. `manager-git`만 누락되어 위임 경계가 불명확한 상태.

### 1.2 비목표 (Non-Goals)

- Skill 본문(Quick Reference / Implementation Guide / Advanced 섹션) 내용 수정 금지
- Skill의 trigger keyword, allowed-tools, related-skills 변경 금지
- 다른 agent의 Scope Boundaries / Delegation Protocol 섹션 수정 금지
- Token budget 수치(level1_tokens / level2_tokens) 항목별 재산정 금지 (모두 100 / 5000 fixed)
- Template 외 다른 파일(테스트, 문서) 수정 금지

---

## 2. Scope (범위)

### 2.1 In Scope

- **Owns**: 11개 skill의 SKILL.md frontmatter 한 블록(`progressive_disclosure:`) 추가
- **Owns**: `manager-git` agent body 두 섹션(`Scope Boundaries`, `Delegation Protocol`) 추가
- **Owns**: Template + local 동시 적용 (`.claude/` + `internal/template/templates/.claude/`)
- 변경된 파일에 대한 `make build` 실행으로 embedded.go 재생성 확인

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- 다른 skill의 progressive_disclosure 수치 조정
- 다른 agent의 본문 구조 수정
- Skill의 모듈/예제 파일(`modules/*.md`, `examples.md`) 변경
- 신규 skill / agent 추가
- CI/CD workflow 수정

---

## 3. Environment

- Go 1.24+ (template 빌드용)
- File system writable: `.claude/skills/`, `.claude/agents/moai/`, `internal/template/templates/.claude/`
- `make build` 실행 환경

## 4. Assumptions

- 11개 skill 모두 현재 valid frontmatter를 보유 (YAML parse 가능)
- `manager-git` agent는 표준 frontmatter를 보유 (다른 22 agent와 동일한 시작 패턴)
- Template과 local의 파일 내용이 동기화 상태 (DEF-007 시작 시점 기준)

## 5. Requirements (EARS)

### REQ-DEF007-001 (Ubiquitous)

The MoAI-ADK skill collection **shall** ensure that every SKILL.md file under `.claude/skills/` and `internal/template/templates/.claude/skills/` contains a `progressive_disclosure` frontmatter block with `enabled: true`, `level1_tokens: 100`, and `level2_tokens: 5000` for the 11 target skills.

### REQ-DEF007-002 (Ubiquitous)

The 11 target skills **shall** be: moai-domain-backend, moai-domain-frontend, moai-domain-db-docs, moai-formats-data, moai-framework-electron, moai-library-mermaid, moai-library-nextra, moai-library-shadcn, moai-tool-ast-grep, moai-workflow-ddd, moai-workflow-loop.

### REQ-DEF007-003 (Ubiquitous)

The `manager-git` agent body **shall** contain a `## Scope Boundaries` section with explicit `IN SCOPE` and `OUT OF SCOPE` enumerations, mirroring the pattern used by other 22 agents (e.g., manager-spec).

### REQ-DEF007-004 (Ubiquitous)

The `manager-git` agent body **shall** contain a `## Delegation Protocol` section enumerating delegation rules (which subagents to call when crossing scope boundaries).

### REQ-DEF007-005 (Event-Driven)

**When** a developer modifies any of the 11 skill SKILL.md files or the manager-git agent file, the change **shall** be applied to both `.claude/` and `internal/template/templates/.claude/` paths in the same commit.

### REQ-DEF007-006 (Event-Driven)

**When** the convention sweep is applied, the system **shall** preserve all existing frontmatter fields (no field deletion, no field reordering beyond standard YAML practice).

### REQ-DEF007-007 (State-Driven)

**While** template and local files are out of sync, `make build` execution **shall** regenerate `internal/template/embedded.go` to reflect the template changes.

### REQ-DEF007-008 (Unwanted)

The convention sweep **shall not** modify the body content (sections after frontmatter) of any of the 11 target skills.

### REQ-DEF007-009 (Unwanted)

The convention sweep **shall not** alter the existing description, allowed-tools, related-skills, or metadata blocks of the 11 target skills.

---

## 6. Acceptance Criteria (요약)

전체 acceptance.md 참조. 핵심:

- AC-DEF007-01: 11개 skill SKILL.md 모두 `progressive_disclosure:` 블록 포함 (local + template)
- AC-DEF007-02: `manager-git.md` body에 `## Scope Boundaries`, `## Delegation Protocol` 섹션 존재
- AC-DEF007-03: 변경 전후 `git diff` 결과가 frontmatter 추가만 포함, 본문 무수정
- AC-DEF007-04: `make build` 정상 완료, embedded.go 갱신
- AC-DEF007-05: `go test ./internal/template/...` 통과

---

## 7. Constraints

- **C1**: Template-First Rule (CLAUDE.local.md §2) — 모든 변경은 template에 먼저 반영 후 local 동기화
- **C2**: 16-language neutrality — 본 SPEC은 언어 중립적 메타데이터만 다룸
- **C3**: progressive_disclosure 블록 위치 — 다른 skill의 패턴(metadata 블록 직후 또는 별도 분리 위치)을 따름; 본 SPEC은 metadata 블록 직후를 표준으로 채택
- **C4**: manager-git agent의 신규 섹션 위치 — agent body의 "Primary Mission" / "Core Capabilities" 섹션 직후, 기존 워크플로우 설명 직전 (다른 agent 패턴과 일치)

---

## 8. Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| Frontmatter 추가 시 YAML parse error | High | YAML linter (`yq`) 또는 `go test ./internal/template/...`로 사전 검증 |
| Template / local drift 재발 | Medium | 단일 commit에 양쪽 동시 변경 + `make build` 실행 |
| manager-git 기존 워크플로우 설명과 신규 섹션 중복 | Low | 추가 시 본문 중복 제거 검토; 기존 내용은 보존 |

---

## 9. Dependencies

없음. Phase A의 baseline SPEC.

후속 SPEC 의존성:
- SPEC-V3R3-DEF-001 (선택적): DEF-007 완료 후 진행 권장
- SPEC-V3R3-ARCH-003: 독립
- SPEC-V3R3-ARCH-007: 독립
- SPEC-V3R3-COV-001: ARCH-003 의존, DEF-007 무관

---

## 10. Traceability

| REQ ID | Acceptance Criteria | Source |
|--------|---------------------|--------|
| REQ-DEF007-001 | AC-DEF007-01 | spec-workflow.md §Progressive Disclosure |
| REQ-DEF007-002 | AC-DEF007-01 | 본 SPEC §1.1 |
| REQ-DEF007-003 | AC-DEF007-02 | 다른 22 agent 패턴 (manager-spec body 참조) |
| REQ-DEF007-004 | AC-DEF007-02 | 다른 22 agent 패턴 |
| REQ-DEF007-005 | AC-DEF007-03, AC-DEF007-04 | CLAUDE.local.md §2 Template-First Rule |
| REQ-DEF007-006 | AC-DEF007-03 | 본 SPEC §1.2 비목표 |
| REQ-DEF007-007 | AC-DEF007-04, AC-DEF007-05 | CLAUDE.local.md §2 |
| REQ-DEF007-008 | AC-DEF007-03 | §1.2 |
| REQ-DEF007-009 | AC-DEF007-03 | §1.2 |
