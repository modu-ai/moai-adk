---
id: SPEC-V3R2-EXT-003
title: Plugin System (Design-Only Scope for v3.0.0)
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 2 SPEC writer (Layer 6/7/Cleanup)
priority: P3 Low
phase: "v3.0.0 — Phase 7 — Extension"
module: ".claude-plugin/, internal/plugin/ (future)"
dependencies:
  - SPEC-V3R2-WF-001
related_gap:
  - r3-cc-architecture-plugins
related_theme: "Theme 7 — Extension"
breaking: false
bc_id: []
lifecycle: design-only
tags: "plugin, manifest, scope, design-only, deferred, v3.1, v3"
---

# SPEC-V3R2-EXT-003: Plugin System (Design-Only)

## HISTORY

| Version | Date       | Author | Description                                                         |
|---------|------------|--------|---------------------------------------------------------------------|
| 0.1.0   | 2026-04-23 | Wave 2 | Initial SPEC — design-only plugin manifest + scope for v3.0.0       |

---

## 1. Goal (목적)

Claude Code v2.x의 **plugin** 메커니즘(`~/.claude/plugins/` 하위에 독립 플러그인 배포)을 MoAI 생태계에도 맞춰 **design-only** 수준으로 v3.0.0에 정착시킨다. 본 SPEC은 **구현이 아니라 설계 계약**을 고정한다: (a) 매니페스트 스키마(`.claude-plugin/plugin.json`), (b) 디렉터리 레이아웃, (c) plugin scope rules (plugin이 건드릴 수 있는/없는 영역), (d) CLI 표면의 예약어(reservation). 전체 구현은 **v3.1+ 로 연기**한다. 본 SPEC은 향후 plugin ecosystem이 MoAI의 FROZEN zone, design constitution, 24-skill 카탈로그와 충돌하지 않도록 하는 방어적 스키마를 먼저 확정한다.

### 1.1 배경

R3 §CC-architecture-reread: plugins는 Claude Code의 공식 확장 메커니즘으로, `.claude-plugin/plugin.json` 매니페스트 + scope 기반 격리를 제공한다. MoAI는 현재 plugin concept을 사용하지 않으나, 향후 3rd-party extension(예: custom skills, agents, output-styles를 단일 번들로 배포)을 고려할 때 스키마를 미리 고정해 두는 것이 lock-in 방지에 유리하다. `builder-plugin` agent가 이미 존재하므로 (agent catalog §Builder), 설계 계약 자체는 코드 없이도 agent-facing 문서로 배포 가능하다.

### 1.2 비목표 (Non-Goals)

- Plugin 런타임 구현 (v3.1+ 연기)
- Plugin CLI 서브커맨드 (`moai plugin install/update/remove` 등) 구현
- Plugin 마켓플레이스 / 레지스트리
- Plugin 간 의존성 그래프
- Plugin sandbox / permission enforcement 구현
- Go 측 loader 추가
- 1st-party plugin 예시 번들 작성
- Plugin 이 MoAI core FROZEN zone을 override하는 메커니즘 (**영구 금지**)

---

## 2. Scope (범위)

> **DESIGN-ONLY SPEC — no implementation in v3.0.0 GA.** Full plugin system execution (CLI, loader, runtime, sandbox) is deferred to v3.1.0. This SPEC defines manifest schema + plugin scope rules only. The REQs below are design specifications, not implementation targets for v3.0.0.

### 2.1 In Scope

- **Owns**: `.claude-plugin/plugin.json` JSON Schema 정의, plugin 디렉터리 레이아웃 명세, plugin scope rules 문서화, CLI 예약어 확정.
- 매니페스트 필수 필드: `name` (semver-compatible unique id), `version` (semver), `description`, `author`, `compatibility` (moai-adk version range), `provides` (skills/agents/commands/output-styles subset), `scope` (permissions).
- 매니페스트 선택 필드: `homepage`, `repository`, `license`, `dependencies` (other plugins).
- Plugin scope rules: plugin은 `.claude/skills/<plugin-name>/`, `.claude/agents/<plugin-name>/`, `.claude/commands/<plugin-name>/`, `.claude/output-styles/<plugin-name>/` 아래에서만 쓰기 가능.
- **FROZEN** override 금지: plugin은 `.claude/rules/moai/design/constitution.md`, `.moai/project/brand/`, agency-absorbed skills(`moai-domain-copywriting`, `moai-domain-brand-design`)을 수정할 수 없다.
- CLI 예약어: `moai plugin`, `moai plugins` (`moai plugin install <url>`, `moai plugin list` 등은 v3.1+에서 구현).

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Plugin 런타임 loader 구현
- Plugin marketplace/index/registry
- Plugin 의존성 resolver
- Plugin 버전 자동 업데이트
- Plugin permission sandbox
- Plugin hot-reload
- Signed/verified plugin 검증 메커니즘
- Plugin 예시 번들 작성
- Plugin 간 conflict resolution 전략 (name collision 등)

---

## 3. Environment (환경)

- 런타임: 본 SPEC은 design-only이므로 런타임 환경 영향 없음.
- 영향 디렉터리:
  - 신설 (문서): `.moai/docs/plugin-schema.md` (JSON Schema 기술서)
  - 예약 (미래): `.claude-plugin/` 디렉터리 루트, `~/.claude/plugins/<plugin-name>/`
- 외부 레퍼런스: R3 §CC-architecture-reread, `.claude/agents/builder-plugin.md`, Claude Code 공식 plugin docs

---

## 4. Assumptions (가정)

- Claude Code v2.x의 plugin 매커니즘은 v3 release window 내 안정적이다.
- MoAI 사용자 중 plugin 작성자는 소수이며 v3.0에서는 demand가 적다.
- 향후 plugin 마켓플레이스는 3rd-party가 호스팅할 수 있다.
- 현재 MoAI의 24-skill 카탈로그(SPEC-V3R2-WF-001)는 1st-party로 간주되며 plugin으로 재배포되지 않는다.
- Plugin scope rules는 Go-level enforcement 없이도 documentation-level contract로 충분하다 (implementation은 v3.1+).

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-EXT003-001**
The system **shall** define a plugin manifest schema at `.claude-plugin/plugin.json` with required fields: `name`, `version`, `description`, `author`, `compatibility`, `provides`, `scope`.

**REQ-EXT003-002**
The `name` field **shall** be globally unique, lowercase, dash-separated, and cannot start with `moai-` (reserved for 1st-party).

**REQ-EXT003-003**
The `version` field **shall** follow semver 2.0.0.

**REQ-EXT003-004**
The `compatibility` field **shall** declare a moai-adk version range (e.g., `">=3.0.0 <4.0.0"`).

**REQ-EXT003-005**
The `provides` field **shall** enumerate at most 4 subsets: `skills[]`, `agents[]`, `commands[]`, `output-styles[]`.

**REQ-EXT003-006**
The `scope` field **shall** declare a subset from `{ read, write, network }`, defaulting to `read` only.

**REQ-EXT003-007**
Plugin-provided files **shall** live under `<plugin-root>/{skills,agents,commands,output-styles}/<plugin-name>/` — never at the top level of `.claude/`.

### 5.2 Event-Driven Requirements

**REQ-EXT003-008**
**When** a plugin manifest lacks a required field, future plugin loader (v3.1+) **shall** reject with `PLUGIN_MANIFEST_INCOMPLETE`.

**REQ-EXT003-009**
**When** a plugin attempts to write to a path outside its scope, the runtime **shall** reject with `PLUGIN_SCOPE_VIOLATION` (v3.1+ enforcement; v3.0 documentation only).

**REQ-EXT003-010**
**When** a plugin's `compatibility` range does not include the current moai-adk version, the loader **shall** reject with `PLUGIN_INCOMPATIBLE`.

### 5.3 State-Driven Requirements

**REQ-EXT003-011**
**While** the plugin is installed, the files under `<plugin-root>/` **shall** be read-only from moai-adk core's perspective (core doesn't auto-update plugin content).

**REQ-EXT003-012**
**While** v3.0.0 is active, no Go runtime plugin loader **shall** be implemented — only the schema and documentation.

### 5.4 Optional Requirements

**REQ-EXT003-013**
**Where** a plugin wants to declare dependencies on other plugins, the manifest **may** include a `dependencies: {<name>: <semver-range>}` map.

**REQ-EXT003-014**
**Where** a plugin wants to self-describe metadata, the manifest **may** include `homepage`, `repository`, `license`.

### 5.5 Complex Requirements (Unwanted Behavior / Composite)

**REQ-EXT003-015 (Unwanted Behavior)**
**If** a plugin manifest attempts to declare it `provides: rules`, **then** validation **shall** reject with `PLUGIN_RULES_FORBIDDEN` (rules are core, not pluggable in v3.0).

**REQ-EXT003-016 (Unwanted Behavior)**
**If** a plugin attempts to modify FROZEN zone files (`.claude/rules/moai/design/constitution.md`, `.moai/project/brand/*`, `moai-domain-copywriting`, `moai-domain-brand-design`), **then** the v3.1+ loader **shall** reject with `PLUGIN_FROZEN_VIOLATION`.

**REQ-EXT003-017 (Complex: State + Event)**
**While** a plugin declares `scope: network`, **when** the loader starts it, **then** the loader **shall** emit a clear warning to the user asking confirmation (v3.1+ implementation; v3.0 documentation only).

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-EXT003-01**: Given a sample valid manifest When schema validation runs Then all required fields pass (maps REQ-EXT003-001..007).
- **AC-EXT003-02**: Given a manifest with `name: moai-foo` When schema validation runs Then rejection per REQ-EXT003-002 (maps REQ-EXT003-002).
- **AC-EXT003-03**: Given a manifest with `version: 1.0` (non-semver) When schema validation runs Then rejection per REQ-EXT003-003 (maps REQ-EXT003-003).
- **AC-EXT003-04**: Given a manifest with `compatibility: ">=3.0.0"` When loader (v3.1+) is implemented Then rangely compatibility check passes for moai-adk 3.0.5 (maps REQ-EXT003-004, REQ-EXT003-010).
- **AC-EXT003-05**: Given a manifest declaring `provides: rules` When schema validation runs Then `PLUGIN_RULES_FORBIDDEN` rejection (maps REQ-EXT003-015).
- **AC-EXT003-06**: Given the plugin scope documentation When inspected Then forbidden target list includes design-constitution.md, brand/*, copywriting, brand-design (maps REQ-EXT003-016).
- **AC-EXT003-07**: Given a sample plugin directory layout When inspected Then skills/agents/commands/output-styles each scope to `<plugin-name>/` (maps REQ-EXT003-007).
- **AC-EXT003-08**: Given the design-only scope When v3.0.0 releases Then no Go plugin loader exists in `internal/plugin/` (maps REQ-EXT003-012).
- **AC-EXT003-09**: Given the CLI surface When `moai plugin` is invoked in v3.0.0 Then error "plugin support is scheduled for v3.1; current v3.0 is design-only" (maps REQ-EXT003-012).
- **AC-EXT003-10**: Given a manifest declaring `dependencies: {"plugin-foo": "^1.0.0"}` When schema validation runs Then the optional field is accepted (maps REQ-EXT003-013).
- **AC-EXT003-11**: Given a plugin with `scope: [network]` When future loader runs Then a confirmation warning is emitted (maps REQ-EXT003-017).
- **AC-EXT003-12**: Given `.moai/docs/plugin-schema.md` When inspected Then it contains the JSON Schema with all required/optional fields documented (maps REQ-EXT003-001).

---

## 7. Constraints (제약)

- 본 SPEC은 **구현을 포함하지 않음** (design-only). v3.0.0 릴리스에 Go 측 plugin 코드 신설 금지.
- `moai-` prefix는 1st-party reserved.
- FROZEN zone (agency, constitution) override는 영구 금지.
- Plugin은 skills/agents/commands/output-styles 4개 subset에만 contribute 가능 (v3.0 scope).
- 9-direct-dep 정책 준수 (신규 외부 의존성 없음, 본 SPEC은 문서만).

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 완화 |
|---|---|---|
| 스키마 확정 후 v3.1 구현에서 필드 부족 발견 | v3.0 스키마 호환 breakage | `dependencies`, `homepage` 등 optional 확장 필드 여유 |
| 1st-party plugin과 3rd-party 이름 충돌 | install 실패 | `moai-` prefix reservation (REQ-EXT003-002) |
| Plugin이 FROZEN zone을 우회 시도 | 무결성 파괴 | REQ-EXT003-016 + v3.1 loader에서 강제 |
| Plugin이 rules를 override 시도 | core 일관성 파괴 | REQ-EXT003-015 hard rejection |
| 사용자가 v3.0에서 plugin 기능 기대 | 혼란 | CLI stub `moai plugin`이 명확한 not-implemented 메시지 제공 |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3R2-WF-001: 24-skill 카탈로그가 1st-party baseline을 확립해야 plugin scope와 구분 가능.

### 9.2 Blocks

- v3.1+ plugin 런타임 구현 (본 SPEC이 schema baseline 제공).
- No v3.0.0 SPECs are blocked by EXT-003 since it is design-only (lifecycle: design-only). EXT-003 does not gate any Phase 1–9 release.

### 9.3 Related

- `.claude/agents/builder-plugin.md`: plugin 생성 도우미 agent가 본 schema를 consume하게 됨.

---

## 10. Traceability (추적성)

- REQ 총 17개: Ubiquitous 7, Event-Driven 3, State-Driven 2, Optional 2, Complex 3.
- AC 총 12개, 모든 REQ에 최소 1개 AC 매핑 (100% 커버리지).
- Wave 2 소스 앵커: R3 §CC-architecture-reread plugins; `.claude/agents/builder-plugin.md`.
- BC 영향: 없음 (design-only).
- 구현 경로 예상 (문서만):
  - `.moai/docs/plugin-schema.md` (JSON Schema 기술서)
  - `.claude/rules/moai/workflow/plugin-scope.md` (scope rules; 문서 level)

---

End of SPEC.
