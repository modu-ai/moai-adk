# SPEC-OPUS47-COMPAT-001: Claude Code v2.1.110/111 + Opus 4.7 Prompt Philosophy Compatibility

## Metadata

| Field | Value |
|-------|-------|
| SPEC ID | SPEC-OPUS47-COMPAT-001 |
| Status | draft |
| Priority | critical |
| Created | 2026-04-17 |
| GitHub Issue | #671 |

---

## 1. Background

Claude Code v2.1.110/111 introduces significant runtime changes alongside the Opus 4.7 model release. MoAI-ADK templates, configuration, and prompt philosophy must be updated to:

1. Register Opus 4.7 in the model catalog with 5-level effort support
2. Codify Opus 4.7's prompt philosophy (5 principles verified by Anthropic)
3. Adapt to v2.1.110 runtime changes (MCP scope dedup, PermissionRequest updatedInput, bypassPermissions policy, Bash timeout ceiling)
4. Fix Windows CLAUDE_ENV_FILE injection gap
5. Remove defensive scaffolding that Opus 4.7 self-mitigates

---

## 2. Requirements

### REQ-OC-001: Opus 4.7 Model Catalog and Profile Update (Ubiquitous + MODIFY)

**Trigger**: System starts or agent is spawned.
**Action**: `claude-opus-4-7` is recognized, effort 5-level (low/medium/high/xhigh/max) is supported.
**Outcome**: Model routing, agent frontmatter, and llm.yaml reflect the new generation.

### REQ-OC-002: Opus 4.7 Prompt Philosophy 5 Principles (Event-Driven + MODIFY)

**Trigger**: Any prompt is constructed for Opus 4.7.
**Action**: Five principles are applied: xhigh default effort, single-turn commitment, Adaptive Thinking, explicit fan-out, reasoning-up/tools-down.
**Outcome**: Constitution, CLAUDE.md, and agent-common-protocol reflect the philosophy shift.

### REQ-OC-003: v2.1.110 Runtime Adaptation (State-Driven + MODIFY)

**Trigger**: MoAI runs on Claude Code >= v2.1.110.
**Action**: /doctor MCP scope dedup detection documented, PermissionRequest updatedInput re-validation documented, bypassPermissions policy clarified, Bash timeout ceiling (600s) documented.
**Outcome**: Hooks, settings, and runtime docs reflect v2.1.110 behavior.

### REQ-OC-004: Windows CLAUDE_ENV_FILE Normalization (Optional + MODIFY)

**Trigger**: SessionStart on Windows platform.
**Action**: `session_start.go` Windows branch injects CLAUDE_ENV_FILE if missing.
**Outcome**: Windows users get env file support equivalent to macOS/Linux.

### REQ-OC-005: Remove Legacy Defensive Scaffolding (Unwanted + MODIFY)

**Trigger**: Opus 4.7 is the active model.
**Action**: Remove redundant instructions that Opus 4.7 handles natively (over-engineering guards, dark-flow warnings, sycophancy prevention boilerplate).
**Outcome**: Leaner prompts, reduced token consumption, no behavioral regression.

---

## 3. Acceptance Criteria

### Primary Scenarios (GWT-1 ~ GWT-8)

**GWT-1**: Given model-policy.md is loaded, When an agent specifies `model: opus`, Then it resolves to `claude-opus-4-7` with `xhigh` default effort.

**GWT-2**: Given agent-authoring.md is loaded, When an agent sets `effort: xhigh`, Then the value is accepted without error.

**GWT-3**: Given CLAUDE.md Section 12 is loaded, When a user reads the MCP section, Then Opus 4.7 Adaptive Thinking and UltraThink are correctly differentiated.

**GWT-4**: Given v2.1.110 runtime, When /doctor runs with duplicate MCP scopes, Then the behavior is documented in hooks-system.md.

**GWT-5**: Given v2.1.110 runtime, When PermissionRequest fires with updatedInput, Then re-validation behavior is documented.

**GWT-6**: Given a Windows session, When SessionStart fires, Then CLAUDE_ENV_FILE is injected if missing.

**GWT-7**: Given moai-constitution.md, When Opus 4.7 is active, Then "Enforce Simplicity" behavior references Opus 4.7 native self-mitigation.

**GWT-8**: Given the output style (moai.md), When the model info line is rendered, Then it reflects "Opus 4.7" identity.

### Edge Cases (EC-1 ~ EC-5)

**EC-1**: Opus 4.6 fallback: If user pins to Opus 4.6, existing 4-level effort (low/medium/high/max) still works.
**EC-2**: GLM compatibility: GLM modes skip Opus 4.7-specific effort levels.
**EC-3**: `effort: max` vs `effort: xhigh`: Both are valid for Opus 4.7; `max` remains the ceiling.
**EC-4**: Template vs local divergence: Template updates do not overwrite local customizations.
**EC-5**: Bash timeout: Commands specifying >600000ms are documented as clamped.

### Quality Gates (QG-1 ~ QG-5)

**QG-1**: `go test ./...` passes after all changes.
**QG-2**: `go vet ./...` reports no issues.
**QG-3**: No hardcoded model version IDs (e.g., `claude-opus-4-7-20260401`) in templates.
**QG-4**: CLAUDE.md stays under 40,000 characters.
**QG-5**: `make build` succeeds (embedded templates regenerated).

---

## 4. Out of Scope

- Vertex AI / Bedrock provider integration
- /tui, /focus, Remote Control UX changes
- plugin_errors handling
- TRACEPARENT SDK integration
- /ultrareview relationship restructuring

---

## 5. Traceability Matrix

| REQ | GWT | EC | Files (count) |
|-----|-----|----|---------------|
| REQ-OC-001 | GWT-1, GWT-2 | EC-1, EC-3 | 8 |
| REQ-OC-002 | GWT-3, GWT-7, GWT-8 | EC-2 | 9 |
| REQ-OC-003 | GWT-4, GWT-5 | EC-5 | 6 |
| REQ-OC-004 | GWT-6 | EC-4 | 2 |
| REQ-OC-005 | GWT-7, GWT-8 | - | 5 |

---

Version: 1.0.0
Created: 2026-04-17
