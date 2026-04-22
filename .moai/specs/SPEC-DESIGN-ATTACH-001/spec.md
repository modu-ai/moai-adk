---
id: SPEC-DESIGN-ATTACH-001
version: 0.2.0
status: draft
created_at: 2026-04-20
updated_at: 2026-04-21
author: moai-adk-go
priority: Medium
labels: [design, context, auto-load, skill, workflow-extension, token-budget]
issue_number: null
depends_on: [SPEC-DESIGN-CONST-AMEND-001, SPEC-DESIGN-DOCS-001]
related_specs: [SPEC-DESIGN-PENCIL-001]
---

# SPEC-DESIGN-ATTACH-001: moai-workflow-design-context Skill + /moai design Phase B2.5 Auto-Loading

## HISTORY

- 2026-04-21 v0.2.0: plan-auditor iteration 1 FAIL 후속 수정. Frontmatter MoAI 표준 (version, labels, created_at, priority=Medium) 적용. REQ↔AC traceability 완전화(14 REQ × ≥1 AC, 총 17 AC). 각 REQ에 `**Rationale**:` 한 줄 추가. D8 token-budget 알고리즘(ceiling(char/4) + 10% buffer) REQ-5에 명시. D5 `.md` suffix 불일치 해소(bare token 통일). D13 REQ-1↔REQ-9 모순 해소(REQ-1에 auto_load=true 조건 게이팅 명시). D6 REQ-13 수동태/주체 명시로 재작성. 기타 D9/D10/D12 해소.
- 2026-04-20 v0.1.0: SPEC 최초 작성.

---

## Background

### 사용자 요구사항

사용자는 `/moai design` 커맨드가 실행되거나 디자인 작업이 요청될 때, `.moai/design/` 폴더에 위치한 사람이 작성한 디자인 브리프 문서(`research.md`, `system.md`, `spec.md`, `pencil-plan.md`)의 내용이 오케스트레이터 프롬프트에 자동으로 첨부되어 `expert-frontend`, `moai-domain-brand-design`, Pencil 등 하위 에이전트가 소비할 수 있도록 요구함. 이는 moai-studio가 이 폴더를 모든 리디자인 이터레이션의 입력으로 사용하는 패턴을 따름.

### 기존 인프라

- **워크플로우 스킬**: `.claude/skills/moai/workflows/design.md`는 Phase 0 (preflight) → Phase 1 (route) → Phase A (Claude Design) / Phase B (code-based) → Phase C (GAN loop) 구조를 가짐. 본 SPEC은 Phase B3 (BRIEF generation) 이전에 Phase B2.5 (Design Context Loading)을 삽입함.
- **FROZEN 규정**: `.claude/rules/moai/design/constitution.md` Section 3 (SPEC-DESIGN-CONST-AMEND-001 적용 후 v3.3.0)은 이미 자동 로딩의 HARD 규칙과 우선순위를 선언함. 본 SPEC은 해당 규칙을 구현함.
- **디자인 설정**: `.moai/config/sections/design.yaml`은 `brand_context.dir`, `claude_design` 키를 포함함. 본 SPEC은 `design_docs` 섹션을 추가 확장함.
- **Template-First 원칙**: 신규 스킬은 `internal/template/templates/.claude/skills/moai-workflow-design-context/SKILL.md`에 원본을 두고 `make build`로 embedded.go에 반영함.

### 신규 스킬 설계 개요

`moai-workflow-design-context`:
- **유형**: workflow skill
- **범위**: project-level
- **트리거**: `/moai design`의 Phase B가 실행될 때 로드되며, 디자인 컨텍스트 점검 용도로 standalone 호출도 가능함
- **입력**: `.moai/design/` 디렉터리 경로 (기본값) 또는 명시적 `dir` 인자
- **출력**: 오케스트레이터 프롬프트 주입용으로 포맷된 통합 컨텍스트 블록
- **동작**: 우선순위 순으로 `.md` 파일을 읽고, `_TBD_` 전용 파일은 건너뛰며, 토큰 예산을 준수하고, 구조화된 출력을 반환함

### 용어 정의 (Terminology)

- **파일 우선순위 토큰(priority token)**: 디자인 브리프 파일을 가리키는 bare token. 정식 집합은 `[spec, system, research, pencil-plan]`이며, 실제 파일 시스템상 파일명은 `<token>.md` 규칙으로 매핑됨 (예: `spec` → `.moai/design/spec.md`). `.md` 접미사는 UI/로그/내러티브 서술에서만 선택적으로 사용하며, 본 SPEC의 REQ·AC·yaml 모두 **bare token** 형식으로 통일함.
- **token_budget (문서 토큰 예산)**: Claude 컨텍스트 주입 전에 스킬이 준수해야 하는 누적 토큰 상한. `design.yaml`의 `design_docs.token_budget`로 설정되며 기본값 20000.
- **tokenization 방법**: 본 SPEC에서 "토큰 카운트"는 `ceiling(character_count / 4)` 근사법을 사용함(REQ-5). Claude 실제 토크나이저와 일치하지 않을 수 있으므로 누적 합계에 **10% overhead buffer**를 가산하여 비교함. 즉, `estimated_tokens = ceiling(char_count / 4) * 1.10`을 `token_budget`과 비교.

---

## Requirements (EARS)

다음 요건은 EARS(Easy Approach to Requirements Syntax) 패턴으로 기술됨. 총 14개 요건. 각 요건은 한 줄 Rationale을 포함함.

- **REQ-1 (Event-Driven)**: `WHEN /moai design workflow reaches Phase B2 (before BRIEF generation) AND design_docs.auto_load_on_design_command is true, the system SHALL invoke moai-workflow-design-context skill to load .moai/design/ bare-token files.`
  - **Rationale**: Phase B2.5 자동 로딩은 사용자의 핵심 요구이며, 명시적 opt-in(auto_load=true) 조건을 REQ-1 트리거에 게이팅해 REQ-9와의 모순을 제거함.

- **REQ-2 (Event-Driven)**: `WHEN moai-workflow-design-context loads files, it SHALL follow priority order spec > system > research > pencil-plan (bare tokens, mapped to <token>.md on disk) as codified in constitution Section 3.2.`
  - **Rationale**: FROZEN constitution Section 3.2가 정의한 우선순위를 실행 레벨에서 강제하여 팀 간 해석 차이를 원천 차단함.

- **REQ-3 (Event-Driven)**: `WHEN a bare-token file contains ONLY _TBD_ content (no user-authored sections), the skill SHALL skip that file and log the skip reason.`
  - **Rationale**: 초기화 직후 `_TBD_`만 포함된 스켈레톤이 오케스트레이터 프롬프트를 오염시키는 것을 방지함.

- **REQ-4 (Event-Driven)**: `WHEN loading files, the skill SHALL respect the token budget declared in design.yaml design_docs.token_budget (default 20000).`
  - **Rationale**: Plan Phase 30K 토큰 할당 안에서 디자인 컨텍스트가 다른 오케스트레이터 자원을 고갈시키지 않도록 상한을 강제함.

- **REQ-5 (Unwanted Behavior / Conditional)**: `IF cumulative estimated_tokens (computed as ceiling(character_count / 4) * 1.10 overhead buffer) exceeds design_docs.token_budget, THEN the skill SHALL truncate in REVERSE priority order (drop pencil-plan first, then research, then system, preserving spec as last). IF a single file alone exceeds the remaining budget after higher-priority files are included, the skill SHALL truncate that file at the nearest Markdown section boundary (##/###) BEFORE exceeding the budget AND SHALL append a trailing marker line "> truncated: <filename> at char_offset=N" to signal partial inclusion.`
  - **Rationale**: 토큰 카운팅 알고리즘과 단일 파일 오버플로 동작을 요건으로 명문화하여 구현자 재량을 제거하고 결정적(deterministic) 동작을 보장함 (audit D8 해소).

- **REQ-6 (Event-Driven)**: `WHEN the skill outputs consolidated context, it SHALL include file-source citation comments: "> source: .moai/design/<filename>" before each section.`
  - **Rationale**: 하위 에이전트가 어떤 파일에서 유래한 지시인지 추적할 수 있어야 R-1 truncation 리스크가 관측 가능해짐.

- **REQ-7 (Event-Driven)**: `WHEN .moai/design/ directory does not exist, the skill SHALL return gracefully with an empty context block AND log "design docs not initialized — run /moai init or SPEC-DESIGN-DOCS-001 to create".`
  - **Rationale**: 디렉터리 부재는 정상 상태(초기 프로젝트)이므로 실패가 아닌 빈 컨텍스트 + 안내 로그로 우아하게 처리.

- **REQ-8 (Event-Driven)**: `WHEN design.yaml is rendered, it SHALL include a design_docs section with keys: dir, auto_load_on_design_command, token_budget, priority.`
  - **Rationale**: 설정 스키마를 고정하여 템플릿-소비자(스킬) 간 계약을 명확히 함.

- **REQ-9 (State-Driven)**: `WHILE design_docs.auto_load_on_design_command is false, the skill SHALL NOT auto-invoke during /moai design AND SHALL respond only to explicit user request (e.g., standalone invocation with explicit dir).`
  - **Rationale**: 컨텍스트 예산이 부족하거나 수동 검토 워크플로가 필요한 사용자를 위해 opt-out 경로를 보장 (audit D12 state-driven 재작성).

- **REQ-10 (Event-Driven)**: `WHEN the skill is invoked standalone (not from /moai design), it SHALL accept an optional dir argument to point at a different design folder (e.g., a submodule's .moai/design/).`
  - **Rationale**: 모노레포·서브모듈에서 여러 디자인 폴더를 교차 점검할 수 있는 유연성을 제공.

- **REQ-11 (State-Driven)**: `WHILE loading files, the skill SHALL read files in parallel using Read tool to minimize latency.`
  - **Rationale**: 4개 파일 최대를 병렬로 읽어 Phase B2.5 추가 지연이 사용자 체감 속도에 악영향을 주지 않도록 함.

- **REQ-12 (Event-Driven)**: `WHEN consolidated context is returned to the orchestrator, it SHALL be formatted as a single Markdown block starting with the exact line "## Design Context (from .moai/design/)" as its first non-empty line.`
  - **Rationale**: 고정 헤더를 계약화하여 하위 에이전트가 해당 섹션을 결정적으로 감지·파싱할 수 있음.

- **REQ-13 (Event-Driven, Passive Responder)**: `WHEN Phase B2.5 is added to workflows/design.md by the implementer, the modification SHALL place the new section between existing Phase B2 (brand context load) and Phase B3 (BRIEF generation), AND SHALL NOT renumber downstream phase labels (B3, B4, ...).`
  - **Rationale**: 기존 Phase 번호에 의존하는 문서·테스트를 깨지 않도록 삽입 위치와 재매기기 금지를 명시 (audit D6 grammar/responder 해소).

- **REQ-14 (Unwanted Behavior / Conditional)**: `IF the skill fails to read any file (permission error, corruption), THEN it SHALL report the failure in a warnings array but SHALL continue with remaining files (partial success semantics).`
  - **Rationale**: 단일 파일 장애로 전체 디자인 컨텍스트 로딩이 실패하면 실제 현장 사용성이 크게 떨어지므로 부분 성공을 허용.

### Additional Edge-Case Requirements (D9/D10 해소)

다음 두 요건은 REQ-1~14 외의 edge case 보완이며, 상위 REQ와 동일 레벨에서 검증됨:

- **REQ-15 (Conditional / Unwanted Behavior)**: `IF .moai/design/ directory exists but every discoverable bare-token file is _TBD_-only (all candidates skipped per REQ-3), THEN the skill SHALL emit the empty context block, log "design docs present but all are _TBD_ — no content loaded", AND continue without error.`
  - **Rationale**: 모든 파일이 스켈레톤 상태인 과도기 상황을 결정적으로 처리해 하위 에이전트가 빈 `## Design Context` 블록과 빈 컨텍스트를 혼동하지 않게 함.

- **REQ-16 (Conditional / Fallback)**: `IF design_docs section is absent from design.yaml at runtime, THEN the skill SHALL use compiled-in defaults (dir=".moai/design", auto_load_on_design_command=true, token_budget=20000, priority=[spec, system, research, pencil-plan]) AND log "design_docs not configured — using defaults".`
  - **Rationale**: 기존 프로젝트 설정 파일이 업그레이드되지 않은 경우에도 하위 호환을 보장.

---

## Acceptance Criteria

총 17개 수용 기준. 모든 REQ(1-16)가 최소 1개의 AC에 대응되도록 traceability 완전화 (audit D4 해소).

### 구조 및 스키마 AC

- **AC-1**: `internal/template/templates/.claude/skills/moai-workflow-design-context/SKILL.md` exists with valid YAML frontmatter (name, description, triggers, allowed-tools CSV).
- **AC-2** (→ REQ-1, REQ-13): `.claude/skills/moai/workflows/design.md` contains a Phase B2.5 section between existing Phase B2 and Phase B3 with the exact heading `### Phase B2.5: Load .moai/design/ Context`, AND downstream phase labels `### Phase B3`, `### Phase B4`, ... are present with identical text (only line offsets shifted by the B2.5 insertion).
- **AC-3** (→ REQ-4, REQ-8): `internal/template/templates/.moai/config/sections/design.yaml` contains a `design_docs:` subkey under `design:` with 4 subkeys (`dir`, `auto_load_on_design_command`, `token_budget`, `priority`) and default `token_budget: 20000`.
- **AC-4** (→ REQ-2): Priority array in `design_docs.priority` in design.yaml equals `[spec, system, research, pencil-plan]` verbatim (bare tokens, no `.md` suffix).

### 동작 AC

- **AC-5** (→ REQ-3): Given `.moai/design/research.md` contains only `_TBD_` lines, when the skill is invoked, it logs a skip entry for `research` and the output block does NOT include any content derived from research.md.
- **AC-6** (→ REQ-6): When the skill outputs context, each included section starts with a `> source: .moai/design/<filename>` comment line as its first line.
- **AC-7**: `go test ./internal/template/...` passes after template changes (embedded.go regeneration).
- **AC-8**: The skill frontmatter `description` field is ≤ 250 characters (single-line collapsed length of the folded YAML `>` block) per skill-authoring coding standard.
- **AC-9** (→ REQ-5): Given a design folder whose four files yield `ceiling(char/4) * 1.10 = 22000` estimated tokens (budget=20000), the skill's output omits content from `pencil-plan` first. IF still over budget, it omits `research`; always preserves `spec`. IF a single preserved file alone still exceeds remaining budget, the output contains that file truncated at the nearest `##`/`###` boundary AND ends with a line matching the regex `^> truncated: .+ at char_offset=\d+$`.
- **AC-10** (→ REQ-7): When `.moai/design/` does not exist, the skill returns a context block that is either empty (zero non-comment characters) or contains only the standard header from REQ-12, AND logs a line containing the substring `design docs not initialized`.
- **AC-11** (→ REQ-9, REQ-1): When `design_docs.auto_load_on_design_command=false`, a simulated `/moai design` Phase B2 trace does NOT invoke `moai-workflow-design-context` (verified by the absence of the `## Design Context (from .moai/design/)` header in the orchestrator's outgoing subagent prompt).
- **AC-12** (→ REQ-10): Invoking the skill standalone with `dir=/tmp/alt-design/` causes its Read tool calls to target paths under `/tmp/alt-design/` and zero reads to target `.moai/design/` in the project root.
- **AC-13** (→ REQ-11): The skill's Read tool calls for the 2-to-4 candidate bare-token files are issued as a single batched parallel tool-call set (verifiable in the agent trace: all Read calls share the same orchestration turn).
- **AC-14** (→ REQ-12): The consolidated output block's first non-empty line equals exactly `## Design Context (from .moai/design/)` (byte-for-byte, no leading/trailing whitespace).
- **AC-15** (→ REQ-14): Given `system.md` is unreadable (permission error simulated by chmod 000 or an injected Read error), the skill returns a warnings array whose elements contain the substring `system` AND the output block still contains `> source:` lines for each of the readable files (`spec`, `research`, `pencil-plan` as applicable).

### Edge-Case AC

- **AC-16** (→ REQ-15): Given `.moai/design/` exists and every candidate bare-token file contains only `_TBD_` content, the skill returns the standard header from REQ-12 with no following `> source:` lines AND logs a line containing the substring `all are _TBD_`.
- **AC-17** (→ REQ-16): Given `design.yaml` has no `design_docs` key, the skill still executes using defaults (`dir=.moai/design`, `auto_load=true`, `token_budget=20000`, `priority=[spec, system, research, pencil-plan]`) AND logs a line containing the substring `using defaults`.

### REQ↔AC Traceability Matrix

| REQ | Covered By | Coverage Type |
|-----|------------|---------------|
| REQ-1 | AC-2, AC-11 | Direct (phase insertion) + indirect (auto_load gate) |
| REQ-2 | AC-4, AC-9 | Direct (priority array) + behavioral (truncation order) |
| REQ-3 | AC-5 | Direct |
| REQ-4 | AC-3, AC-9 | Direct (schema) + behavioral (enforcement) |
| REQ-5 | AC-9 | Direct (algorithm + single-file overflow) |
| REQ-6 | AC-6 | Direct |
| REQ-7 | AC-10 | Direct |
| REQ-8 | AC-3 | Direct |
| REQ-9 | AC-11 | Direct |
| REQ-10 | AC-12 | Direct |
| REQ-11 | AC-13 | Direct |
| REQ-12 | AC-14 | Direct |
| REQ-13 | AC-2 | Direct (downstream phase labels preserved) |
| REQ-14 | AC-15 | Direct |
| REQ-15 | AC-16 | Direct |
| REQ-16 | AC-17 | Direct |

모든 REQ가 ≥1 AC로 커버됨. AC-1/AC-7/AC-8은 일반 품질 게이트(구조/빌드/스타일)로 REQ에 1:1 대응하지 않음.

---

## Skill Frontmatter Template (참조용)

신규 스킬 `moai-workflow-design-context`의 YAML frontmatter 템플릿. 구현 시 그대로 사용함.

```yaml
---
name: moai-workflow-design-context
description: >
  Loads human-authored design briefs from .moai/design/ (research, system, spec,
  pencil-plan) and injects them into /moai design workflow context with priority
  truncation when token budget is exceeded.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-04-20"
  tags: "design, context, attach, auto-load, brand, design-brief"

progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

triggers:
  keywords: ["design context", "attach design", "design brief", ".moai/design"]
  agents: ["expert-frontend"]
  phases: ["design"]
---
```

---

## design.yaml Extension Snippet

기존 `design:` 최상위 키 아래에 `design_docs` 서브섹션을 추가함. 다른 키(brand_context, claude_design 등)는 변경하지 않음.

```yaml
design:
  # ...existing keys...
  design_docs:
    dir: ".moai/design"
    auto_load_on_design_command: true
    token_budget: 20000
    priority: [spec, system, research, pencil-plan]
```

---

## Phase B2.5 Workflow Section Skeleton

`.claude/skills/moai/workflows/design.md`의 Phase B2와 Phase B3 사이에 삽입할 신규 섹션 스켈레톤. 하위 Phase 번호(B3, B4 등)는 재매기지 않음.

```
### Phase B2.5: Load .moai/design/ Context

1. Check .moai/design/ exists. If absent: skip, log "design docs not initialized".
2. Check design_docs.auto_load_on_design_command. If false: skip (user may invoke standalone).
3. Read README.md for attach rules (if present).
4. Invoke moai-workflow-design-context skill with dir=".moai/design".
5. Receive consolidated context block (Markdown, token-capped per REQ-5 algorithm).
6. Prepend context block to the orchestrator's next subagent prompt (expert-frontend or moai-domain-brand-design).
7. Proceed to Phase B3 (BRIEF generation).
```

---

## Scope

### IN SCOPE

- 신규 스킬 `moai-workflow-design-context` 정의 및 template-first 경로에 SKILL.md 생성 (※ 본 SPEC에서는 스킬 파일 자체를 **생성하지 않음** — 생성은 후속 run 단계. 본 SPEC은 요건·수용기준·프론트매터 템플릿·yaml 확장 스펙만 확정.)
- `.moai/config/sections/design.yaml`의 `design_docs` 섹션 확장 스펙
- `.claude/skills/moai/workflows/design.md`에 Phase B2.5 삽입 스펙
- 우선순위 기반 로딩, 토큰 예산 준수, 역순 절단(truncation), 단일 파일 오버플로 경계 절단 동작
- `_TBD_` 전용 파일 스킵 로직 + "전체 _TBD_" edge case 처리
- 부분 성공(partial success) 시맨틱 및 warning 리포트
- standalone 호출 지원(`dir` 인자)
- `design_docs` 미설정 시 컴파일-인 기본값 fallback

### OUT OF SCOPE

- Pencil MCP 통합 (SPEC-DESIGN-PENCIL-001에서 처리)
- `.moai/design/` 실제 템플릿 파일 내용 작성 (SPEC-DESIGN-DOCS-001에서 처리)
- Claude Design bundle import (기존 `moai-workflow-design-import` 스킬이 이미 처리)
- GAN loop (Phase C) 수정
- constitution.md Section 3 개정 (SPEC-DESIGN-CONST-AMEND-001에서 처리)
- 사용자가 `design_docs.priority` 기본값을 변경했을 때의 실제 강제 검증(본 SPEC은 warning 로깅만 권장, 강제 검증은 향후 SPEC)

---

## Exclusions (What NOT to Build)

[HARD] 본 SPEC이 명시적으로 배제하는 항목:

- **E-1**: Pencil MCP 파일 포맷 파싱 또는 `.pen` 파일 처리는 본 SPEC에서 구현하지 않음 (SPEC-DESIGN-PENCIL-001 범위)
- **E-2**: `.moai/design/` 초기 템플릿 콘텐츠 작성(research.md, system.md 등의 초기 스켈레톤)은 본 SPEC에서 생성하지 않음 (SPEC-DESIGN-DOCS-001 범위)
- **E-3**: Claude Design handoff bundle import 경로 수정은 수행하지 않음 (기존 `moai-workflow-design-import` 유지)
- **E-4**: GAN loop 이터레이션 동작, sprint contract 프로토콜, evaluator-active 점수 로직 변경은 포함하지 않음
- **E-5**: constitution.md의 FROZEN zone 텍스트 수정은 포함하지 않음 (SPEC-DESIGN-CONST-AMEND-001 전제)
- **E-6**: `.moai/design/` 외부 경로(예: 다른 프로젝트 또는 원격 저장소)에서의 원격 로딩은 미지원 (standalone `dir` 인자는 로컬 경로만 허용)
- **E-7**: 파일 내용 자동 번역/로컬라이제이션은 수행하지 않음 (사용자 작성 원문 그대로 전달)
- **E-8**: 로딩된 컨텍스트의 영속 캐싱은 구현하지 않음 (매 호출마다 재로딩)
- **E-9**: `design_docs.priority` 사용자 커스터마이즈 값에 대한 강제 검증/차단은 본 SPEC에서 구현하지 않음 (warning 로깅은 R-5에서 권장, 강제화는 향후 SPEC)
- **E-10**: Claude 공식 토크나이저 정확 토큰 카운트 호출은 구현하지 않음 (REQ-5의 `ceiling(char/4) * 1.10` 근사 알고리즘으로 확정)

---

## Risks

- **R-1**: 우선순위 기반 절단으로 인해 컨텍스트가 불완전해질 수 있음 → REQ-6/AC-6의 `> source:` 인용 주석과 REQ-5/AC-9의 `> truncated:` 마커로 오케스트레이터가 어떤 파일이 포함/부분포함/제외되었는지 인지 가능.
- **R-2**: 스킬의 토큰 예산 계산과 실제 Claude API 토큰 카운트 사이 불일치 가능 → REQ-5에서 10% overhead buffer를 알고리즘의 일부로 정규화함(Risk → Requirement 승격, audit D8 해소).
- **R-3**: Phase B2.5 추가가 기존 `design.md` 통합 테스트를 깨뜨릴 가능성 → `/moai run` 단계에서 통합 검증 수행, AC-2로 downstream phase 라벨 보존 확인, AC-7로 `go test ./internal/template/...` 통과 확인.
- **R-4**: 대형 디자인 브리프(한 파일당 50KB+)에서 병렬 Read의 메모리 피크 → REQ-11에서 병렬 Read를 강제하지만, 개별 파일 크기가 토큰 예산을 초과하는 경우 REQ-5의 단일 파일 경계 절단 로직으로 완화됨.
- **R-5**: `design_docs.priority` 키를 사용자가 임의로 재정렬할 경우 constitution Section 3.2 FROZEN 우선순위와 불일치 발생 가능 → 구현 시 기본값과 일치하지 않으면 warning 로깅 권장. 강제화는 E-9에 따라 본 SPEC 범위 외.

---

## Verification Checklist

- [ ] Directory format: `.moai/specs/SPEC-DESIGN-ATTACH-001/`
- [ ] SPEC ID uniqueness 검증됨
- [ ] spec.md 작성 완료 (본 파일)
- [ ] Frontmatter: id, version, status, created_at, updated_at, author, priority (Medium), labels, issue_number, depends_on, related_specs 전부 존재
- [ ] EARS format 14 base requirements + 2 edge-case requirements (REQ-15, REQ-16) 포함
- [ ] Acceptance Criteria 17개 포함 및 REQ↔AC traceability matrix 포함
- [ ] 모든 REQ에 `**Rationale**:` 한 줄 존재
- [ ] Exclusions 섹션 존재 (E-1 ~ E-10)
- [ ] Scope IN/OUT SCOPE 명시
- [ ] Risks 섹션 존재 (R-1 ~ R-5)
- [ ] Template frontmatter, design.yaml extension, Phase B2.5 스켈레톤 포함
- [ ] spec.md에 구현 세부사항(함수명, 클래스 구조) 미포함 (what/why 중심)
- [ ] `.md` suffix 일관성: REQ·AC·yaml 모두 bare token 사용 (spec, system, research, pencil-plan)
- [ ] REQ-1 ↔ REQ-9 모순 해소: REQ-1 트리거에 `auto_load_on_design_command=true` 조건 명시됨
