---
id: SPEC-V3R2-EXT-001
title: Typed Memory Taxonomy (4-type enforcement)
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 2 SPEC writer (Layer 6/7/Cleanup)
priority: P1 High
phase: "v3.0.0 — Phase 7 — Extension"
module: ".claude/agent-memory/, .claude/rules/moai/workflow/moai-memory.md, internal/hook/session_start.go"
dependencies: []
related_gap:
  - pattern-library-M-1
  - r3-cc-architecture-memdir
related_theme: "Theme 7 — Extension"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "memory, taxonomy, user, feedback, project, reference, staleness, v3"
---

# SPEC-V3R2-EXT-001: Typed Memory Taxonomy

## HISTORY

| Version | Date       | Author | Description                                                                |
|---------|------------|--------|----------------------------------------------------------------------------|
| 0.1.0   | 2026-04-23 | Wave 2 | Initial SPEC — 4-type memory taxonomy, staleness, MEMORY.md cap            |

---

## 1. Goal (목적)

pattern-library §M-1 (Typed Memory Taxonomy)을 MoAI의 agent-memory 시스템에 공식 적용한다. R3 §1.1 memdir + §4 Adoption Candidate 7에서 검증된 4-type 분류(**user / feedback / project / reference**)를 `.claude/agent-memory/<agent-name>/` 디렉터리 하위 메모리에 강제하며, `MEMORY.md` 인덱스의 line cap(200줄)과 staleness caveat(§M-5)을 규정한다. 현재 MoAI의 agent-memory는 informal schema를 사용 중이며(agent 프롬프트에서 memory type 예시를 나열하지만 검증 없음), 본 SPEC은 (a) type enum 고정, (b) MEMORY.md 인덱스 포맷, (c) 오래된 메모리의 system-reminder wrap을 runtime 규약으로 전환한다.

### 1.1 배경

pattern-library §M-1: "Four typed memory kinds with distinct lifetimes and prompts — user (about the developer), feedback (correction patterns), project (state of current work), reference (external system pointers). v3 disposition: ADOPT — moai's existing MEMORY.md already uses this schema informally. Formalize as rule in `.claude/rules/moai/workflow/moai-memory.md`. Low-cost, high-value normalization." §M-5 (LLM-Selected Relevance with Staleness Caveat): "Stale memories (>1 day) wrapped in `<system-reminder>` with explicit caveat to prevent over-trust. v3 disposition: ADOPT." 현재 agent memory prompt는 각 agent별로 4-type을 예시 수준에서만 다루며, runtime validation이 없어 drift가 가능한 상태다.

### 1.2 비목표 (Non-Goals)

- Memory graduation protocol(1x→3x→5x→10x) 신설 (SPEC-SLQG-001 기존 메커니즘 재사용)
- Agent memory의 embedding/retrieval 알고리즘 변경 (기존 LLM-selected 유지)
- memdir 전체 도입 (M-1 taxonomy만 적용)
- Three-Layer Memory(M-2) 도입 (본 SPEC은 M-1/M-5만)
- Workflow Memory Induction (M-4) 도입
- Skill Library embeddings (M-3) 도입
- Memory 백업/복원 도구 신설

---

## 2. Scope (범위)

### 2.1 In Scope

- **Owns**: `.claude/rules/moai/workflow/moai-memory.md` (확장), `.claude/agent-memory/<agent-name>/MEMORY.md` 스키마, SessionStart 훅의 staleness 경고 주입.
- 4-type enum 고정: `user | feedback | project | reference` (변경 금지, 추가 금지 — v3.0 scope).
- 각 memory 파일의 frontmatter 필수 키: `name`, `description`, `type` (enum).
- `MEMORY.md` 인덱스 cap: 200줄 이내, 그 초과 시 CI warning (후속 정리 유도).
- 1일 이상 업데이트 없는 memory는 SessionStart 훅에서 `<system-reminder>` wrap하여 loader가 agent context에 주입할 때 staleness caveat 동반.
- 타입별 writing guideline 문서화: user(long-lived profile), feedback(correction patterns), project(current work state), reference(external system pointers).

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- 신규 memory type 추가
- Memory LLM retrieval 알고리즘 구현
- Memory DB / embedding 인덱싱
- Memory 삭제/아카이브 자동화
- Memory encryption / access control
- Multi-project memory sharing
- Lessons.md와의 merge (SPEC-SLQG-001 별도)

---

## 3. Environment (환경)

- 런타임: Claude Code agent memory loader, `internal/hook/session_start.go`의 memory eval 구간
- 영향 디렉터리:
  - 수정: `.claude/rules/moai/workflow/moai-memory.md`
  - 참조: `.claude/agent-memory/<agent-name>/MEMORY.md` (28+ agent 디렉터리)
  - 수정: `internal/hook/session_start.go` (staleness caveat 주입)
- 외부 레퍼런스: pattern-library §M-1, §M-5; R3 §1.1 memdir

---

## 4. Assumptions (가정)

- Agent memory는 이미 `.claude/agent-memory/<agent-name>/` 아래 존재하며 MEMORY.md 파일이 있다.
- Memory 파일은 text markdown이며 frontmatter를 가질 수 있다.
- SessionStart 훅은 Go handler로 실행되고 file mtime을 읽을 수 있다.
- Agent memory content는 민감정보가 아니므로 system-reminder wrap은 안전하다.
- MEMORY.md 200줄 cap은 Claude Code의 컨텍스트 로딩 한계(200줄 후 truncate)와 일치한다.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-EXT001-001**
Every memory file under `.claude/agent-memory/<agent-name>/` (except `MEMORY.md` index) **shall** declare frontmatter key `type` with value from the enum `{user, feedback, project, reference}`.

**REQ-EXT001-002**
Every memory file **shall** declare frontmatter keys `name` (string) and `description` (string).

**REQ-EXT001-003**
`MEMORY.md` **shall** contain no more than 200 lines; content beyond line 200 **shall** be considered truncated by Claude Code's context loader.

**REQ-EXT001-004**
The `type` enum **shall** be fixed to exactly 4 values in v3.0.0 and addition requires a new SPEC.

**REQ-EXT001-005**
The rule file `.claude/rules/moai/workflow/moai-memory.md` **shall** enumerate per-type writing guidelines: when to save, how to use, body structure.

### 5.2 Event-Driven Requirements

**REQ-EXT001-006**
**When** SessionStart hook loads agent memory, it **shall** detect files with `mtime > 24 hours` and wrap their content in `<system-reminder>` with caveat "This memory may be stale; verify against current state before acting on it."

**REQ-EXT001-007**
**When** a new memory file is written without `type` frontmatter, PostToolUse hook **shall** emit warning `MEMORY_MISSING_TYPE` (non-blocking).

**REQ-EXT001-008**
**When** `MEMORY.md` exceeds 200 lines, PostToolUse hook **shall** emit warning `MEMORY_INDEX_OVERFLOW` suggesting archival.

### 5.3 State-Driven Requirements

**REQ-EXT001-009**
**While** a memory file's `type` is `user`, its content **shall** describe the developer's role/preferences/responsibilities — never ephemeral task details.

**REQ-EXT001-010**
**While** a memory file's `type` is `feedback`, its body **shall** lead with the rule, then `**Why:**` and `**How to apply:**` sub-sections.

**REQ-EXT001-011**
**While** a memory file's `type` is `project`, its body **shall** lead with the fact/decision, then `**Why:**` and `**How to apply:**` sub-sections.

**REQ-EXT001-012**
**While** a memory file's `type` is `reference`, its content **shall** point to external systems (Linear project, Grafana dashboard, Slack channel) rather than memorize details.

### 5.4 Optional Requirements

**REQ-EXT001-013**
**Where** SPEC-V3R2-EXT-002 implements a Go loader for memory, the loader **shall** parse the 4-type enum and reject files with invalid `type` values.

**REQ-EXT001-014**
**Where** memory older than 30 days remains unread in 5 consecutive sessions, the system **may** recommend archival to `lessons-archive.md` (no auto-delete).

### 5.5 Complex Requirements (Unwanted Behavior / Composite)

**REQ-EXT001-015 (Unwanted Behavior)**
**If** memory content matches excluded categories (code patterns, git history, debugging recipes, CLAUDE.md content, ephemeral task state), **then** the content audit **shall** warn `MEMORY_EXCLUDED_CATEGORY` with the category name.

**REQ-EXT001-016 (Unwanted Behavior)**
**If** a duplicate memory is detected (same `description` across 2+ files), **then** the audit **shall** recommend merge with `MEMORY_DUPLICATE`.

**REQ-EXT001-017 (Complex: State + Event)**
**While** SessionStart loads memory, **when** 10+ files are wrapped as stale, the system **shall** emit a single aggregated warning rather than 10 separate ones.

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-EXT001-01**: Given a memory file missing `type` frontmatter When PostToolUse hook runs Then `MEMORY_MISSING_TYPE` warning (maps REQ-EXT001-001, REQ-EXT001-007).
- **AC-EXT001-02**: Given a memory file with `type: unknown` When loader runs Then file is rejected (maps REQ-EXT001-001, REQ-EXT001-013).
- **AC-EXT001-03**: Given `MEMORY.md` with 250 lines When PostToolUse hook runs Then `MEMORY_INDEX_OVERFLOW` warning (maps REQ-EXT001-008).
- **AC-EXT001-04**: Given a feedback memory with proper structure When inspected Then body has rule + `**Why:**` + `**How to apply:**` sections (maps REQ-EXT001-010).
- **AC-EXT001-05**: Given a user-type memory file When inspected Then content describes developer role/preferences, not task state (maps REQ-EXT001-009).
- **AC-EXT001-06**: Given `.claude/rules/moai/workflow/moai-memory.md` When inspected Then per-type writing guidelines are present (maps REQ-EXT001-005).
- **AC-EXT001-07**: Given a memory file with mtime > 24h When SessionStart loads Then content is wrapped in `<system-reminder>` with staleness caveat (maps REQ-EXT001-006).
- **AC-EXT001-08**: Given 10+ stale memory files When SessionStart runs Then a single aggregated warning is emitted (maps REQ-EXT001-017).
- **AC-EXT001-09**: Given a PR adding a 5th `type` value When CI runs Then rejection points to REQ-EXT001-004 (maps REQ-EXT001-004).
- **AC-EXT001-10**: Given a memory file containing CLAUDE.md-sourced content When content audit runs Then `MEMORY_EXCLUDED_CATEGORY` warning (maps REQ-EXT001-015).
- **AC-EXT001-11**: Given 2 memory files with identical `description` When audit runs Then `MEMORY_DUPLICATE` recommendation (maps REQ-EXT001-016).
- **AC-EXT001-12**: Given a reference-type memory When inspected Then it points to external system (URL/project name), not reproduce contents (maps REQ-EXT001-012).
- **AC-EXT001-13**: Given a project-type memory When inspected Then body has fact/decision + `**Why:**` + `**How to apply:**` (maps REQ-EXT001-011).

---

## 7. Constraints (제약)

- 4-type enum은 v3.0에서 고정 (추가/변경은 새 SPEC 필요).
- MEMORY.md 200-line cap은 Claude Code 컨텍스트 로더의 truncation 한계와 일치.
- Staleness threshold는 24시간 기본 (configurable via `.moai/config/sections/workflow.yaml`, 본 SPEC 기본값만 정의).
- 9-direct-dep 정책 준수.
- Memory content는 민감정보 미포함 가정.

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 완화 |
|---|---|---|
| 기존 informal memory가 type 미선언 | 대량 warning | 초기 wave에서는 warning only (non-blocking), subsequent wave에서 점진 강제 |
| MEMORY.md 200줄 cap이 엄격함 | 컨텐츠 압축 부담 | archive 경로 제공 + 주기적 정리 회고 |
| Staleness wrap이 agent 혼란 야기 | 컨텍스트 저하 | wrap 메시지 포맷 고정, agent prompt에 해석 가이드 포함 |
| Duplicate memory 오탐 | 잘못된 merge 제안 | REQ-EXT001-016는 recommend only (auto-merge 없음) |
| 타입별 body 규격이 사용자별 다름 | 일관성 저하 | 예시 2-3개를 `moai-memory.md`에 고정 제공 |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- 없음.

### 9.2 Blocks

- SPEC-V3R2-EXT-002: Go loader가 4-type enum을 활용.

### 9.3 Related

- SPEC-SLQG-001 lessons protocol (auto-capture 메커니즘 재사용).

---

## 10. Traceability (추적성)

- REQ 총 17개: Ubiquitous 5, Event-Driven 3, State-Driven 4, Optional 2, Complex 3.
- AC 총 13개, 모든 REQ에 최소 1개 AC 매핑 (100% 커버리지).
- Wave 2 소스 앵커: pattern-library §M-1, §M-5; R3 §1.1 memdir; 기존 `.claude/rules/moai/workflow/moai-memory.md`.
- BC 영향: 없음 (warnings are non-blocking in v3.0).
- 구현 경로 예상:
  - `.claude/rules/moai/workflow/moai-memory.md` (4-type 가이드 확장)
  - `internal/hook/session_start.go` (staleness caveat 주입)
  - `internal/hook/post_tool.go` (MEMORY_MISSING_TYPE / MEMORY_INDEX_OVERFLOW 검증)

---

End of SPEC.
