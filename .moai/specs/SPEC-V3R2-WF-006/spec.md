---
id: SPEC-V3R2-WF-006
title: Output Styles Alignment (MoAI, Einstein)
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 2 SPEC writer (Layer 6/7/Cleanup)
priority: P2 Medium
phase: "v3.0.0 — Phase 6 — Multi-Mode Workflow"
module: ".claude/output-styles/moai/, internal/template/templates/.claude/output-styles/moai/"
dependencies: []
related_gap:
  - r6-output-styles
related_theme: "Theme 7 — Extension"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "output-styles, moai, einstein, claude-code-schema, precedence, v3"
---

# SPEC-V3R2-WF-006: Output Styles Alignment

## HISTORY

| Version | Date       | Author | Description                                                  |
|---------|------------|--------|--------------------------------------------------------------|
| 0.1.0   | 2026-04-23 | Wave 2 | Initial SPEC — align MoAI + Einstein styles, schema, loading |

---

## 1. Goal (목적)

현재 MoAI-ADK는 두 개의 output style(`MoAI`, `Einstein`)을 제공하며 R6 §3은 둘 다 "KEEP"로 판정했다. 본 SPEC은 (a) 두 style의 frontmatter가 Claude Code 공식 output-style 스키마를 엄격히 준수하도록 검증 규칙을 codify하고, (b) loading precedence(`project > user`)를 명문화하며, (c) template/local byte-identity를 CI로 강제한다. v3.0.0에서 output-styles의 안정적 base를 확립하고 차기 SPEC-V3R2-EXT-002의 Go loader가 의존할 스키마를 고정한다.

### 1.1 배경

R6 §3: "Only 2 styles (MoAI, Einstein), both active, both template-synced. Well-bounded. No action needed." §3.4: "Both files are identical between `.claude/output-styles/moai/` and `internal/template/templates/.claude/output-styles/moai/`. Clean sync." R6 §3.1-§3.2 frontmatter 샘플:
- MoAI: `name: MoAI`, `description`, `keep-coding-instructions: true`
- Einstein: `name: Einstein`, `description`, `keep-coding-instructions: false`

현재 loading은 `settings.json`의 `outputStyle` 키로 결정되며 user-level config(`~/.claude/settings.json`)가 project-level(`.claude/settings.json`)을 override할 수 있다. 공식 precedence 규칙은 문서화되지 않았다. EXT-002가 Go loader를 구현하기 전 정책적으로 고정할 필요가 있다.

### 1.2 비목표 (Non-Goals)

- 신규 output style 추가 (v3.0에서 2개 고정)
- 기존 style의 body rewrite
- `outputStyle` 설정 key 이름 변경
- Einstein style의 Context7 grounding / mermaid / Notion integration rewrite
- user-level config에서 project를 override하도록 precedence 변경 (본 SPEC은 project > user로 고정)
- Output style 수준의 i18n 도입

---

## 2. Scope (범위)

### 2.1 In Scope

- **Owns**: `.claude/output-styles/moai/moai.md`, `.claude/output-styles/moai/einstein.md`, 동일 template 경로.
- Output style frontmatter 스키마 검증 rule: `name`, `description`, `keep-coding-instructions` (bool).
- Loading precedence: `project settings.json outputStyle` > `user settings.json outputStyle` > hardcoded default "MoAI".
- Template/local byte-identity CI (기존 commands pattern 확장).
- SPEC-V3R2-EXT-002의 Go loader가 활용할 manifest 경로(`.claude/output-styles/moai/{name}.md`).

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- 3번째 output style 추가
- Output style 내부 body rewrite
- `/config` UI 개선
- Output style 간 런타임 switching API
- Output style에 다국어 지원 추가
- EXT-002의 Go loader 구현 (본 SPEC은 schema 확립만)

---

## 3. Environment (환경)

- 런타임: Claude Code output-style loader, `.claude/settings.json`
- 영향 디렉터리:
  - 검증: `.claude/output-styles/moai/*.md`, `internal/template/templates/.claude/output-styles/moai/*.md`
  - 수정: `internal/template/output_styles_audit_test.go` (신규 혹은 확장)
- 외부 레퍼런스: R6 §3 output styles audit

---

## 4. Assumptions (가정)

- Claude Code output-style 스키마(`name`/`description`/`keep-coding-instructions`)는 v3 release window 내 안정적이다.
- User가 `.claude/settings.json`에 `outputStyle: "..."`를 명시할 수 있다.
- Project > user precedence는 현재 Claude Code의 기본 resolution과 호환된다 (검증 필요).
- Template/local drift가 0인 현재 상태는 유지 가능한 invariant이다.
- 두 style 모두 UTF-8 인코딩, CRLF 미사용이다.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-WF006-001**
Each output style file **shall** declare frontmatter keys: `name` (string), `description` (string), `keep-coding-instructions` (boolean).

**REQ-WF006-002**
The v3 output-styles tree **shall** contain exactly two styles: `MoAI` and `Einstein`.

**REQ-WF006-003**
Template (`internal/template/templates/.claude/output-styles/moai/`) and local (`.claude/output-styles/moai/`) **shall** be byte-identical.

**REQ-WF006-004**
The system **shall** document loading precedence: project `settings.json` `outputStyle` overrides user `settings.json` `outputStyle`, which overrides the hardcoded default "MoAI".

**REQ-WF006-005**
`MoAI` style **shall** set `keep-coding-instructions: true`; `Einstein` style **shall** set `keep-coding-instructions: false`.

### 5.2 Event-Driven Requirements

**REQ-WF006-006**
**When** a project sets `outputStyle: "Einstein"` in `.claude/settings.json`, the Claude Code session **shall** load `Einstein` regardless of user-level setting.

**REQ-WF006-007**
**When** a style file violates the frontmatter schema (missing key, wrong type), CI **shall** reject with `OUTPUT_STYLE_SCHEMA_ERROR`.

**REQ-WF006-008**
**When** an unknown style name is referenced in `settings.json` `outputStyle`, the Claude Code session **shall** fall back to "MoAI" default and emit warning log.

### 5.3 State-Driven Requirements

**REQ-WF006-009**
**While** no explicit `outputStyle` is set in any settings file, the default style **shall** be "MoAI".

**REQ-WF006-010**
**While** template and local output-style trees diverge, `make build` **shall** fail with `OUTPUT_STYLE_DRIFT`.

### 5.4 Optional Requirements

**REQ-WF006-011**
**Where** future v3.x versions add a third output style, it **shall** go through this SPEC's schema validation before being registered.

**REQ-WF006-012**
**Where** SPEC-V3R2-EXT-002 implements a Go loader, the loader **shall** consume the frontmatter schema defined here.

### 5.5 Complex Requirements (Unwanted Behavior / Composite)

**REQ-WF006-013 (Unwanted Behavior)**
**If** `keep-coding-instructions` is missing or non-boolean, **then** the style **shall not** load and `OUTPUT_STYLE_SCHEMA_ERROR` with offending key is emitted.

**REQ-WF006-014 (Unwanted Behavior)**
**If** a third output style is added without passing this SPEC's schema validation, **then** the commit **shall** be blocked by `OUTPUT_STYLE_UNVERIFIED`.

**REQ-WF006-015 (Complex: State + Event)**
**While** user-level `settings.json` specifies `outputStyle: "Einstein"` AND project-level specifies `outputStyle: "MoAI"`, **when** Claude Code session starts, the project-level **shall** win and "MoAI" loads.

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-WF006-01**: Given `.claude/output-styles/moai/moai.md` When frontmatter is parsed Then `name="MoAI"`, `keep-coding-instructions: true` (maps REQ-WF006-001, REQ-WF006-005).
- **AC-WF006-02**: Given `.claude/output-styles/moai/einstein.md` When frontmatter is parsed Then `name="Einstein"`, `keep-coding-instructions: false` (maps REQ-WF006-001, REQ-WF006-005).
- **AC-WF006-03**: Given v3 output-styles tree When listed Then exactly 2 files: moai.md, einstein.md (maps REQ-WF006-002).
- **AC-WF006-04**: Given `diff -rq .claude/output-styles/moai internal/template/templates/.claude/output-styles/moai` When run Then output is empty (maps REQ-WF006-003).
- **AC-WF006-05**: Given project `settings.json outputStyle: "Einstein"` AND user `settings.json outputStyle: "MoAI"` When session starts Then Einstein loads (maps REQ-WF006-004, REQ-WF006-006).
- **AC-WF006-06**: Given user-level `outputStyle: "Einstein"` AND project-level `outputStyle: "MoAI"` When session starts Then MoAI loads (maps REQ-WF006-015).
- **AC-WF006-07**: Given no `outputStyle` set in any settings When session starts Then "MoAI" default loads (maps REQ-WF006-009).
- **AC-WF006-08**: Given `outputStyle: "NonExistent"` When session starts Then "MoAI" fallback with warning log (maps REQ-WF006-008).
- **AC-WF006-09**: Given a style file missing `keep-coding-instructions` When CI runs Then `OUTPUT_STYLE_SCHEMA_ERROR` rejection (maps REQ-WF006-007, REQ-WF006-013).
- **AC-WF006-10**: Given template and local output-styles diverge When `make build` runs Then `OUTPUT_STYLE_DRIFT` failure (maps REQ-WF006-010).
- **AC-WF006-11**: Given a PR adding a 3rd style without schema pass When CI runs Then `OUTPUT_STYLE_UNVERIFIED` rejection (maps REQ-WF006-014).
- **AC-WF006-12**: Given SPEC-V3R2-EXT-002 Go loader when implemented Then it parses frontmatter per REQ-WF006-001 (maps REQ-WF006-012).

---

## 7. Constraints (제약)

- Output style 개수는 v3.0.0에서 2개 고정 (MoAI, Einstein).
- Frontmatter 스키마 변경은 breaking change로 분류 (본 SPEC scope 외).
- 9-direct-dep 정책 준수.
- UTF-8, LF-only 인코딩.
- Project > user precedence (override 없음).

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 완화 |
|---|---|---|
| Claude Code output-style 스키마 변경 | v3 style 호환 실패 | Claude Code release note 모니터링, 변경 시 본 SPEC 개정 |
| Template/local drift 발생 | 사용자 혼란 | REQ-WF006-010 CI 강제 |
| 사용자가 user-level override를 기대 | UX 혼란 | REQ-WF006-004 precedence 명문화 + 문서 |
| 3번째 style 추가 요구 | scope creep | 본 SPEC은 2개 고정; 3번째는 별도 SPEC 필요 |
| Einstein style의 MCP 의존성(Context7) 변경 | 런타임 실패 | Einstein body의 MCP 의존은 soft-fallback 유지 (본 SPEC body 수정 안함) |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- 없음.

### 9.2 Blocks

- SPEC-V3R2-EXT-002: Go loader가 본 SPEC의 schema를 consume.

### 9.3 Related

- R6 §3 output styles audit.

---

## 10. Traceability (추적성)

- REQ 총 15개: Ubiquitous 5, Event-Driven 3, State-Driven 2, Optional 2, Complex 3.
- AC 총 12개, 모든 REQ에 최소 1개 AC 매핑 (100% 커버리지).
- Wave 2 소스 앵커: R6 §3.1/§3.2/§3.4.
- BC 영향: 없음 (기존 style behavior 보존).
- 구현 경로 예상:
  - `.claude/rules/moai/core/settings-management.md` (outputStyle precedence 추가)
  - `internal/template/output_styles_audit_test.go` (신규 또는 확장)

---

End of SPEC.
