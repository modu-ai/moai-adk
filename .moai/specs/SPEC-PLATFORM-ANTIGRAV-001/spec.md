---
id: SPEC-PLATFORM-ANTIGRAV-001
version: "0.1.0"
status: draft
created_at: 2026-05-04
updated_at: 2026-05-04
author: Claude (latte triage)
priority: Medium
labels: [platform, port, antigravity, multi-host, capability]
issue_number: 782
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-04 | Claude (latte triage) | Initial draft from issue #782 (port MoAI-ADK to Google Antigravity). Compatibility gap analysis based on Antigravity public docs (codelabs, blog) snapshot 2026-05-04. |

> Status: **draft**. This SPEC formalizes the port-to-Antigravity capability investigation requested in issue #782. Phases: P1 Discovery & Decision (this SPEC) → P2 Adapter Implementation (downstream SPECs, gated by P1 decision). DO NOT begin implementation until §Decision Gate is resolved.

---

# SPEC-PLATFORM-ANTIGRAV-001: MoAI-ADK Google Antigravity Port — Compatibility Adapter Layer

## Overview

이 SPEC은 MoAI-ADK 자산(agents, skills, slash commands/workflows, hooks, settings)을 Google Antigravity 에이전틱 IDE에서 재사용할 수 있도록 하는 **호환 어댑터 계층(compatibility adapter)** 도입을 다룬다. 이슈 #782 가 보고한 사용자 인식("Antigravity는 동일한 agents/skills/workflows 추상화를 가진다")은 **부분적으로만** 사실이며, 본 SPEC은 (1) 정확한 호환 매트릭스, (2) 어댑터 설계 원칙, (3) 단계적 포팅 전략, (4) Decision Gate를 EARS 형식으로 정의한다. 본 SPEC이 PASS 판정을 받으면 후속 구현 SPEC 시리즈 (`SPEC-PLATFORM-ANTIGRAV-002…NNN`)가 분기된다.

핵심 발견: Antigravity Skills는 Anthropic Skills와 매우 유사 (`SKILL.md` + YAML frontmatter, 디렉토리 단위 패키지) 하나, **Agents 정의 모델, Workflow 트리거 시맨틱, Hooks 시스템, 일부 도구(특히 `AskUserQuestion`/`Agent()`/`Task*`)는 호환되지 않는다.** 따라서 1:1 디렉토리 복사는 불가능하며, 변환·번역 계층이 필요하다.

## Why

### 사용자 가치 가설

- **이슈 #782**의 사용자(@thecloudservicesgroup)는 Antigravity가 Anthropic의 agents/skills/slash-commands 추상화를 공유한다고 가정한다. 본 SPEC은 그 가정을 검증하고, 검증 결과에 따라 포팅 작업의 **실현 가능성 + 비용 + 가치 트레이드오프**를 결정한다.
- **선택지 확장**: MoAI-ADK 사용자가 Claude Code 외 Gemini 3, GPT-OSS 120B, Claude (via Antigravity) 모델을 활용할 수 있게 된다. 이는 (a) 모델 비용 최적화, (b) 모델별 강점 활용 (Gemini 3의 긴 컨텍스트 vs Opus의 추론), (c) vendor lock-in 회피 효과를 가진다.
- **MoAI-ADK 메타-가치**: "16-언어 중립" 원칙(CLAUDE.local.md §15)을 IDE/플랫폼 차원으로 확장. 도구(MoAI-ADK)가 특정 플랫폼에 종속되지 않음을 입증.

### 비용/리스크 가설

- **호환성 부분 일치**: 아래 §호환 매트릭스의 결과에 따라 **포팅 비용이 포팅 가치를 초과할 수 있다.** 본 SPEC의 1차 목적은 그 임계점을 정량화하는 것이다.
- **이중 유지보수 리스크**: 어댑터 도입 시 `internal/template/templates/.claude/` (현재) 와 `internal/template/templates/.agent/` (신규) 의 동기화 부담 발생. 변경 1건당 2회 검증 필요.
- **기능 기능 손실 가능성**: `AskUserQuestion` (HARD 규칙), Hook 이벤트 (SessionStart, PreCompact, SessionEnd), `Agent()` 서브에이전트 위임은 Antigravity에 직접 대응이 없다. 일부 기능은 다운그레이드 또는 미지원 상태로 출시된다.

### Goals

| Goal ID | Type | Target |
|---------|------|--------|
| Goal Primary | Decision | 본 SPEC 종료 시점에 "포팅 진행/보류/거부" 3택 결정이 정량 근거(호환률 %, 추정 인시, 기능 손실 항목 수)와 함께 기록된다 |
| Goal Secondary 1 | Compatibility Matrix | MoAI-ADK 9개 자산 카테고리(skills, agents, commands, hooks, settings, output-styles, rules, MCP, Go binary) 각각에 대해 Antigravity 호환률을 {Full / Partial / None} 으로 분류한다 |
| Goal Secondary 2 | Adapter Design | "Partial" 카테고리 각각에 대해 어댑터 패턴 (변환 스크립트, 듀얼-템플릿, 폴리필) 1개 이상을 명시한다 |
| Goal Secondary 3 | Cost Estimate | 포팅 진행 시 인시(person-hour) 추정치를 카테고리별로 산출한다 (priority 라벨 기반, 시간 단위 사용 금지 — agent-common-protocol §Time Estimation 준수) |
| Goal Anti | Premature Implementation | 본 SPEC 의 §Decision Gate 통과 전에는 어떤 코드/템플릿 변경도 발생하지 않아야 한다 (read-only investigation only) |

Goal Anti 는 invariant. 위반 시 SPEC 자동 회귀.

## Compatibility Matrix (현 시점 분석, 2026-05-04 기준)

| MoAI-ADK 자산 | Claude Code 위치 | Antigravity 대응 | 호환률 | 어댑터 필요? |
|---|---|---|---|---|
| Skills | `.claude/skills/<name>/SKILL.md` | `.agent/skills/<name>/SKILL.md` | **Full** (frontmatter `name`+`description` 동일, 디렉토리 구조 동일, Progressive Disclosure 패러다임 동일) | 디렉토리 리네임만 |
| Skills `allowed-tools` 필드 | `allowed-tools: Read, Write, ...` | 미지원 (도구 제약 frontmatter 부재) | **Partial** | 필드 제거 또는 무시 |
| Agents | `.claude/agents/<group>/<name>.md` (1 파일/에이전트) | `.agents/agents.md` (단일 파일, `@persona` 섹션) | **None** (모델 차이 — Claude는 파일별 분리, Antigravity는 단일 파일 페르소나) | 변환 스크립트 + 합성 |
| Agent `model:`/`effort:`/`permissionMode:` | Anthropic 모델 직지정 | Antigravity는 사용자 IDE 설정에 위임 | **Partial** | 필드 제거 + 외부 매핑 표 |
| Slash Commands | `.claude/commands/<group>/<cmd>.md` (`Skill("moai")` 디스패처) | `.agents/workflows/<name>.md` (사용자 트리거 프롬프트) | **Partial** (포맷 유사, 시맨틱 다름 — Antigravity 워크플로우는 직접 프롬프트, Claude 명령은 Skill 호출) | 워크플로우 본문에 Skill 호출 인라이닝 |
| Hooks (SessionStart/PreCompact/SessionEnd 등) | `.claude/hooks/moai/*.sh` + `settings.json` | 공식 hook 시스템 부재 (2026-05-04 문서 기준) | **None** | 부재; 일부 자동화 손실 |
| Settings (`.claude/settings.json`) | hooks, MCP servers, env, statusline | 위치/포맷 별도 (Antigravity 자체 IDE 설정) | **Partial** | 분리 + MCP 항목만 이식 |
| Output Styles | `.claude/output-styles/moai/moai.md` | 미지원 | **None** | 손실 (Claude 한정 기능) |
| Rules | `.claude/rules/moai/**` (CLAUDE.md 통해 로드) | Antigravity는 `agents.md` 본문 또는 Skills 본문에 인라이닝 | **Partial** | rules → 공유 reference skill 로 흡수 |
| MCP servers | `mcp_servers` in settings.json | MCP 지원 (점차 확대) | **Full** | 설정 파일만 이식 |
| Go binary `moai` (CLI) | platform-agnostic (Hooks/명령에서 호출) | 호출 가능 (터미널 도구) | **Full** (단, Hook 자동 호출 부재 — §Hook 항목 참조) | 무변경 |
| `AskUserQuestion` 도구 | Anthropic deferred tool, Toolsearch preload | 미지원 (Gemini/Antigravity에 대응 도구 없음) | **None** | 다운그레이드: 자유서술 질문 + 구조화 옵션 (HARD 규칙 일시 완화 필요) |
| `Agent()` (서브에이전트 위임) | Anthropic Agent tool | 부분 — Antigravity는 manager surface로 별도 에이전트 spawn 가능, 그러나 `subagent_type` 카탈로그·전달 패턴 다름 | **Partial** | Antigravity manager surface 에 매핑하는 위임 wrapper |
| `Task*` (TaskCreate/Update/List/Get) | Anthropic deferred tools | 미문서화 | **None** | TodoWrite/in-process state 로 폴리필 |

종합 호환률 (가중치 미적용 단순 카운트): **Full 3 / Partial 6 / None 4** = 13개 카테고리 중 약 23% Full, 46% Partial, 31% None.

> 위 매트릭스는 **공개 문서 스냅샷(2026-05-04)** 기반 분석. Antigravity 가 hooks/AskUserQuestion 등을 추후 도입하면 본 SPEC 의 호환률이 상승할 수 있다. P1 Discovery 단계에서 분기마다 갱신해야 한다.

## EARS Requirements

### REQ-ANTIGRAV-001 (Ubiquitous — Compatibility Matrix Authority)

The MoAI-ADK Antigravity port investigation **shall** maintain a single canonical Compatibility Matrix document at `.moai/specs/SPEC-PLATFORM-ANTIGRAV-001/compatibility-matrix.md` that enumerates every MoAI-ADK asset category (minimum: skills, agents, commands, hooks, settings, output-styles, rules, MCP servers, Go binary, `AskUserQuestion`, `Agent()`, `Task*` tools) with a mapped Antigravity counterpart and a compatibility classification of exactly one of {Full, Partial, None}.

### REQ-ANTIGRAV-002 (Event-Driven — Decision Gate Trigger)

**When** the Compatibility Matrix is finalized AND the cost estimate (in priority labels per `agent-common-protocol §Time Estimation`) is recorded, the SPEC owner **shall** convene a Decision Gate review that produces exactly one verdict from the set {Proceed, Defer, Reject} with a written rationale referencing specific matrix entries.

### REQ-ANTIGRAV-003 (State-Driven — Read-Only Investigation Phase)

**While** SPEC-PLATFORM-ANTIGRAV-001 status is `draft` or `in-progress` and the Decision Gate has not yet returned `Proceed`, the implementation **shall not** create, modify, or delete any file under `internal/template/templates/`, `.claude/`, or any other production asset directory. Investigation artifacts under `.moai/specs/SPEC-PLATFORM-ANTIGRAV-001/` are exempt from this freeze.

### REQ-ANTIGRAV-004 (Unwanted Behavior — Speculative Antigravity Asset Generation)

**If** any agent, contributor, or automated workflow attempts to create files under `internal/template/antigravity-templates/`, `.agent/`, or any path that would constitute an Antigravity-specific asset before §Decision Gate returns `Proceed`, **then** the operation **shall** be rejected at code-review time AND the proposing party **shall** justify the action via an amendment to this SPEC's HISTORY block.

### REQ-ANTIGRAV-005 (Complex — Adapter Pattern Selection per Partial Category)

**While** the Compatibility Matrix contains at least one row classified as `Partial`, **when** the SPEC owner produces the §Adapter Design output, the document **shall** specify for each Partial-classified row: (a) the adapter pattern selected from {dual-template, transformer-script, polyfill-skill, runtime-shim, manual-port}, (b) the estimated effort priority label (High / Medium / Low — no time units), (c) the feature loss list (if any) that survives the adapter, and (d) the reversibility classification (reversible / one-way) of the adapter.

### REQ-ANTIGRAV-006 (Event-Driven — Antigravity Documentation Refresh)

**When** Antigravity publishes a documentation update affecting any Compatibility Matrix row (e.g., adding hooks support, AskUserQuestion equivalent, or Task tools), the SPEC owner **shall** refresh the matrix entry within the next plan-auditor review cycle and bump the SPEC version per HISTORY conventions.

### REQ-ANTIGRAV-007 (Ubiquitous — `AskUserQuestion` Downgrade Policy)

If the §Decision Gate returns `Proceed`, the Antigravity port **shall** ship with a documented `AskUserQuestion` downgrade policy that (a) explicitly suspends the `[HARD]` AskUserQuestion-Only Interaction rule (`CLAUDE.md §1`) for Antigravity-hosted runs, (b) provides a fallback interaction pattern (e.g., structured Markdown options + user free-text reply), and (c) notes the bias-prevention regression cost (per `askuser-protocol.md §Option Description Standards`).

### REQ-ANTIGRAV-008 (Event-Driven — Hook System Substitute Decision)

**When** the §Adapter Design phase reaches the Hooks row of the Compatibility Matrix, the SPEC owner **shall** decide between (a) accepting the loss of SessionStart/PreCompact/SessionEnd automation, (b) implementing a polling/manual replacement (e.g., user-triggered workflow that performs equivalent setup), or (c) blocking the port until Antigravity adds hook support. The decision **shall** be recorded with rationale.

## Files to Create (Investigation Phase Only)

The following files belong to this SPEC's investigation phase and are SPEC-internal — they do **not** create Antigravity port code:

- `.moai/specs/SPEC-PLATFORM-ANTIGRAV-001/spec.md` — this document
- `.moai/specs/SPEC-PLATFORM-ANTIGRAV-001/compatibility-matrix.md` — REQ-ANTIGRAV-001 deliverable (full table per category, evidence links to Antigravity docs)
- `.moai/specs/SPEC-PLATFORM-ANTIGRAV-001/adapter-design.md` — REQ-ANTIGRAV-005 deliverable (Partial-row adapter pattern catalog)
- `.moai/specs/SPEC-PLATFORM-ANTIGRAV-001/decision-gate.md` — REQ-ANTIGRAV-002 deliverable (verdict + rationale + downstream SPEC IDs if `Proceed`)
- `.moai/specs/SPEC-PLATFORM-ANTIGRAV-001/plan.md` — phase plan generated during `/moai run` if Decision Gate returns `Proceed`

## Files to Modify

None during investigation phase. (REQ-ANTIGRAV-003, REQ-ANTIGRAV-004 freeze.)

## Files to Create (Conditional — Implementation Phase, gated by Decision Gate `Proceed`)

The following are **forecasts only**, not commitments. Final scope is set by downstream SPECs spawned post-Decision Gate.

- `internal/template/antigravity-templates/.agent/skills/` — mirror of `.claude/skills/moai/**` with `allowed-tools` field stripped
- `internal/template/antigravity-templates/.agents/agents.md` — synthesized from 24 agent files in `.claude/agents/moai/` via transformer script (REQ-ANTIGRAV-005 transformer-script pattern)
- `internal/template/antigravity-templates/.agents/workflows/*.md` — mirror of `.claude/commands/moai/*.md` with `Skill("moai")` calls inlined
- `internal/template/transformer/antigravity.go` — Go transformer translating Claude → Antigravity assets at `make build` time
- `pkg/version/version.go` — extension to expose `BuildPlatform` ({"claude-code", "antigravity"})
- `cmd/moai/init.go` — extension: `moai init --platform antigravity` flag (or sibling subcommand `moai init-antigravity`)
- Documentation: `docs-site/content/{ko,en,ja,zh}/guides/antigravity-port.md` (4-locale per `CLAUDE.local.md §17.3`)

## What NOT to Build (Investigation Phase Hard Out-of-Scope)

- Speculative `.agent/` directory under repo root (REQ-ANTIGRAV-004 blocker)
- Speculative `internal/template/antigravity-templates/` directory (gated)
- Modification of any existing `.claude/` asset to "make it Antigravity-friendly" preemptively
- Antigravity-specific auto-detection in `moai init` command (no work until Decision Gate)
- Promotional documentation ("MoAI-ADK now supports Antigravity!") prior to Decision Gate

## What NOT to Build (Implementation Phase, even if Decision Gate `Proceed`)

These items are **explicitly excluded** from any Antigravity port implementation work:

- **Bidirectional sync** between `.claude/` and `.agent/` at runtime — adapter is build-time only (`make build` regenerates Antigravity templates from Claude templates as the canonical source). Reverse direction is rejected to prevent the dual-source-of-truth anti-pattern.
- **Antigravity-first features** — `.claude/` (Anthropic Claude Code) remains the canonical authoring surface. Antigravity templates are derived, not authored independently.
- **Hook polyfill via polling** without explicit user opt-in — any replacement of SessionStart/PreCompact via background processes requires user consent (privacy + resource cost).
- **Custom MCP server** specifically for Antigravity — existing MCP servers (Context7, sequential-thinking, etc.) are reused as-is.
- **`AskUserQuestion` polyfill via chat-side interaction** that pretends to be structured — bias-prevention rules (`askuser-protocol.md §Option Description Standards`) become unenforceable; the downgrade is acknowledged honestly per REQ-ANTIGRAV-007.
- **Antigravity-only SPECs** — every functional SPEC produced by `/moai plan` continues to target `.claude/` first; Antigravity equivalents are generated, not authored.

## Decision Gate (Investigation Output)

The SPEC reaches **PASS** state when all of the following are true:

1. `compatibility-matrix.md` is complete with one of {Full, Partial, None} for every category in the canonical list (REQ-ANTIGRAV-001).
2. `adapter-design.md` covers every `Partial` row with adapter pattern + effort priority + feature loss + reversibility (REQ-ANTIGRAV-005).
3. `decision-gate.md` records exactly one verdict in {Proceed, Defer, Reject} with rationale (REQ-ANTIGRAV-002).
4. If verdict = `Proceed`: downstream SPEC IDs (`SPEC-PLATFORM-ANTIGRAV-002`+) are pre-allocated and listed in `decision-gate.md`.
5. If verdict ∈ {Defer, Reject}: a sunset condition is documented (e.g., "revisit when Antigravity adds AskUserQuestion equivalent").
6. plan-auditor review (`agent-common-protocol §plan-auditor`) returns PASS verdict on the four investigation artifacts.

## Acceptance Criteria

- AC-1 (must-pass): All four SPEC-internal artifacts (matrix, adapter, decision-gate, this spec.md) exist and are non-empty before SPEC status is set to `complete`.
- AC-2 (must-pass): No file under `internal/template/templates/` is modified in any commit referencing `SPEC-PLATFORM-ANTIGRAV-001` while status is `draft` or `in-progress` (REQ-ANTIGRAV-003 enforcement).
- AC-3 (must-pass): Decision Gate verdict is recorded by an explicit user/maintainer action (not auto-derived from heuristics) — bias prevention.
- AC-4 (should): Compatibility matrix entries cite Antigravity public documentation URLs with retrieval date (per CLAUDE.md §10 Web Search Protocol).
- AC-5 (should): plan-auditor identifies zero `[HIGH]` defects on the four artifacts.

## Open Questions (resolved during `/moai plan` → `/moai run` transition)

These questions block implementation but **not** investigation. They are recorded here so the Decision Gate review can address them:

- OQ-1 (Antigravity Agent model): Antigravity의 단일-파일 `agents.md` 페르소나 모델 vs MoAI-ADK 의 24개 에이전트 카탈로그를 어떻게 매핑할 것인가? 1:1 페르소나 변환 vs 멀티-페르소나 통합 vs 카탈로그를 SKILL 로 변환?
- OQ-2 (`AskUserQuestion` HARD 규칙): Antigravity 호스트에서 HARD 규칙을 일시 완화할 때, MoAI-ADK Constitution v3.4.0 §1 의 "AskUserQuestion-Only Interaction" 절을 어떻게 표현하는가? Constitution 자체를 platform-conditional 로 분기할 것인가?
- OQ-3 (CLAUDE.md vs AGENTS.md): MoAI-ADK 는 CLAUDE.md 를 project instructions 로 사용한다. Antigravity 는 `.agents/agents.md` 를 사용한다. 두 파일을 동기화할 것인가, 단일 source 로부터 build-time 생성할 것인가, 별도 유지할 것인가?
- OQ-4 (Hook 손실 비용): SessionStart hook 이 수행하는 GLM credentials 로드, teammate mode 활성화, env injection 의 사용자-가시 가치는 얼마인가? 이를 잃어도 Antigravity 사용자에게 동등한 가치를 제공할 수 있는가?
- OQ-5 (Decision Gate 결정자): 본 SPEC 의 Decision Gate 는 누가 내리는가? 메인테이너 1인 (@GoosLab) 단독? 사용자 커뮤니티 투표? 이슈 #782 발의자(@thecloudservicesgroup) 의견 가중치는?

OQ-1~OQ-5 는 `/moai plan` 으로 P2 implementation SPEC 분기 시 `plan.md §Open Questions` 에 다시 등장한다.

## References

- Issue #782 — original feature request (https://github.com/modu-ai/moai-adk/issues/782)
- Antigravity Skills 공식 문서 — https://antigravity.google/docs/skills
- Antigravity Skills Codelab — https://codelabs.developers.google.com/getting-started-with-antigravity-skills
- Autonomous AI Developer Pipelines (agents.md / skills.md / workflows) — https://codelabs.developers.google.com/autonomous-ai-developer-pipelines-antigravity
- MoAI-ADK Constitution v3.4.0 — `.claude/rules/moai/core/moai-constitution.md`
- AskUserQuestion Protocol — `.claude/rules/moai/core/askuser-protocol.md`
- Agent Common Protocol — `.claude/rules/moai/core/agent-common-protocol.md`
- 16-Language Neutrality (precedent for platform-neutral design) — `CLAUDE.local.md §15`
