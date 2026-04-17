# SPEC-OPUS47-COMPAT-001: Implementation Plan

## Overview

27 files across 3 priority tiers. All template changes go to `internal/template/templates/` first (single source of truth), then `make build` regenerates embedded files. Local copies (`.claude/`, `.moai/`) are synced via `moai update`.

---

## REQ-OC-001: Opus 4.7 Model Catalog and Profile Update

### Objective

Register `claude-opus-4-7` as the current Opus generation. Extend effort levels from 4 to 5 (add `xhigh`). Update all model references from "Opus 4.6" to "Opus 4.7".

### Implementation Details

#### File 1 (P0): `internal/template/templates/.claude/rules/moai/development/model-policy.md`

**Current state** (line 17-20):
```
Current model generation mapping (as of v2.1.69):
- opus = Opus 4.6 (default effort: medium for Max/Team, use "deepthink" keyword for high effort)
- sonnet = Sonnet 4.6
- haiku = Haiku 4.5
```

**Target state**:
```
Current model generation mapping (as of v2.1.110):
- opus = Opus 4.7 (default effort: xhigh, 5 levels: low/medium/high/xhigh/max)
- sonnet = Sonnet 4.6
- haiku = Haiku 4.5
```

**Effort section** (line 46-53): Update from 3 levels to 5:
```
## Effort Levels (Opus 4.7)

Opus 4.7 supports 5 effort levels:
- low: Fastest responses, minimal reasoning
- medium: Moderate reasoning depth
- high: Deep reasoning, activated by "deepthink" keyword
- xhigh: Default for Opus 4.7 sessions (recommended)
- max: Maximum reasoning depth, highest token consumption
```

**Auto-migration** (line 27): Add `opus-4-6` to the auto-migrated list alongside `opus-4-0`, `opus-4-1`.

#### File 2 (P0): `internal/template/templates/.claude/rules/moai/development/agent-authoring.md`

**Line 37**: Update effort field description:
```
| effort | No | inherit | Session effort override: low, medium, high, xhigh, max |
```

**Line 57**: Update effort detail:
```
**effort**: Overrides session effort level for this agent. Valid values: `low`, `medium`, `high`, `xhigh`, `max`. The `xhigh` value is the default for Opus 4.7. The `max` value enables maximum reasoning depth.
```

#### File 3 (P0): `internal/template/templates/.moai/config/sections/llm.yaml`

No structural changes needed. The `claude_models` section uses abstract tier names (opus/sonnet/haiku), not pinned model IDs. This design is already compatible with Opus 4.7.

**Optional**: Add comment noting Opus 4.7 generation:
```yaml
claude_models:
  high: "opus"      # Opus 4.7 — complex reasoning, architecture, security
  medium: "sonnet"  # Sonnet 4.6 — balanced performance for most tasks
  low: "haiku"      # Haiku 4.5 — fast exploration, simple tasks
```

#### File 4 (P1): `internal/template/templates/.claude/output-styles/moai/moai.md`

Update model identity references from "Opus 4.6" to "Opus 4.7" where the model is mentioned in the output style header or version footer.

#### File 5 (P1): `internal/template/templates/.claude/skills/moai/workflows/run.md`

Update Opus version references in workflow context where model-specific behavior is documented.

#### File 6 (P1): `internal/template/templates/.claude/agents/moai/builder-agent.md`

Update any Opus 4.6-specific comments in agent definitions. The `model` field uses abstract names so no frontmatter change is needed.

#### File 7 (P2): `internal/template/templates/.claude/rules/moai/development/skill-authoring.md`

Update Opus 4.6 references in effort documentation if present.

#### File 8 (P2): `internal/cli/launcher.go`

Review model profile selection logic. If any hardcoded "4.6" references exist in profile naming, update to be generation-agnostic.

---

## REQ-OC-002: Opus 4.7 Prompt Philosophy 5 Principles

### Objective

Codify the 5 principles verified by Anthropic for Opus 4.7 prompting:
1. **xhigh default**: Opus 4.7 performs best at xhigh effort level
2. **Single-turn commitment**: Prefer completing work in one turn rather than spreading across many
3. **Adaptive Thinking**: Model self-regulates reasoning depth per task complexity
4. **Explicit fan-out**: When parallelism is available, use it aggressively
5. **Reasoning up, tools down**: Rely more on native reasoning, fewer tool calls for simple lookups

### Implementation Details

#### File 9 (P0): `internal/template/templates/CLAUDE.md`

**Section 12 (MCP Servers & Deep Analysis Modes)**: Add Opus 4.7 Adaptive Thinking documentation:
```
- **Opus 4.7 Adaptive Thinking**: Native reasoning depth self-regulation. 
  Opus 4.7 automatically adjusts reasoning effort per task complexity. 
  External deepthink MCP is still available but less frequently needed.
  Default effort level is xhigh.
```

Update the model identity line in the existing content if CLAUDE.md references the model version.

**CRITICAL**: CLAUDE.md is currently ~27KB. The 40KB character limit (from coding-standards.md) leaves room for ~13KB additional content. Any new content must be concise and reference external rule files rather than inlining details.

#### File 10 (P0): `internal/template/templates/.claude/rules/moai/core/moai-constitution.md`

**Agent Core Behaviors section**: Integrate principles where they naturally fit:

- **Behavior #4 (Enforce Simplicity)**: Add note that Opus 4.7 natively self-mitigates over-engineering. The rule remains but the framing shifts from "compensate for model tendency" to "reinforce model behavior."

- **Parallel Execution section**: Strengthen fan-out language to align with Principle #4 (explicit fan-out).

#### File 11 (P0): `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`

**Background Agent Execution**: No change needed (already well-defined).

**New section**: Add "Opus 4.7 Prompt Efficiency" section:
```
## Opus 4.7 Prompt Efficiency

When running on Opus 4.7:
- Prefer native reasoning over tool calls for simple lookups
- Trust Adaptive Thinking: do not add effort-level instructions per task
- Single-turn preference: batch related work instead of spreading across turns
- Default effort is xhigh; override only when explicitly needed
```

#### File 12 (P0): `internal/template/templates/.claude/skills/moai-workflow-thinking/SKILL.md`

Update the thinking workflow skill to document the relationship between:
- `--deepthink` (Sequential Thinking MCP) -- external, step-by-step
- `ultrathink` (native extended reasoning) -- internal, high effort
- Opus 4.7 Adaptive Thinking -- automatic, xhigh default

Clarify when each is appropriate with Opus 4.7.

#### File 13 (P1): `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`

Update any references to model-specific behavior in the SPEC workflow documentation.

#### File 14 (P1): `internal/template/templates/.claude/skills/moai-foundation-core/SKILL.md`

Update core skill to reference Opus 4.7 capabilities where relevant.

#### File 15 (P1): `internal/template/templates/.claude/skills/moai-foundation-cc/SKILL.md`

Update Claude Code reference skill with v2.1.110 features and Opus 4.7 model info.

#### File 16 (P1): `internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-cli-reference-official.md`

Update CLI reference with new model options and effort levels.

---

## REQ-OC-003: v2.1.110 Runtime Adaptation

### Objective

Document and adapt to 4 runtime changes in Claude Code v2.1.110/111:
1. `/doctor` MCP scope duplicate detection
2. `PermissionRequest` updatedInput re-validation
3. `bypassPermissions` policy clarification
4. Bash timeout ceiling (600,000ms = 10 minutes)

### Implementation Details

#### File 17 (P0): `internal/template/templates/.claude/rules/moai/core/hooks-system.md`

Add documentation for new runtime behaviors:

```markdown
## v2.1.110 Runtime Changes

### /doctor MCP Scope Deduplication
Claude Code v2.1.110 `/doctor` now detects duplicate MCP tool scopes.
If two MCP servers expose the same tool name, /doctor flags it.
MoAI settings must ensure no overlapping MCP tool scopes in .mcp.json.

### PermissionRequest updatedInput
v2.1.110 adds `updatedInput` field to PermissionRequest hook events.
When Claude modifies its tool call after a PreToolUse hook returns,
the PermissionRequest event now includes the updated input for re-validation.
Hook handlers should validate updatedInput, not just the original input.

### Bash Timeout Ceiling
Maximum Bash timeout is 600,000ms (10 minutes). Values exceeding this
are clamped by the runtime. MoAI hook timeouts should stay well below
this ceiling (recommended: 5-60 seconds for hooks).
```

#### File 18 (P0): `internal/template/templates/.claude/rules/moai/core/settings-management.md`

**Permission Management section**: Add `bypassPermissions` policy documentation:

```markdown
### bypassPermissions Policy (v2.1.110+)

The `bypassPermissions` permission mode skips ALL permission checks including:
- File write/edit operations
- Bash command execution  
- Network access

[HARD] `bypassPermissions` is prohibited for:
- Agents with `background: true` (background agents auto-deny prompts anyway)
- Agents touching files outside project directory
- Agents executing destructive git operations

Allowed only for:
- Fully trusted local automation (CI/CD pipelines)
- Agents with explicit user pre-authorization
```

#### File 19 (P0): `internal/template/templates/.claude/settings.json.tmpl`

Review hook timeout values. Ensure no hook exceeds 60 seconds (well under the 600s ceiling). Current values appear correct (5-60s range).

Add comment documenting the 600s ceiling:
```json
// NOTE: Claude Code v2.1.110 enforces a 600,000ms (10 min) Bash timeout ceiling.
// Hook timeouts should remain at 5-60 seconds.
```

#### File 20 (P1): `internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-settings-official.md`

Update settings reference with v2.1.110 changes.

#### File 21 (P1): `internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-hooks-official.md`

Update hooks reference with PermissionRequest.updatedInput field documentation.

#### File 22 (P1): `internal/template/templates/.mcp.json.tmpl`

Verify no duplicate MCP tool scopes across configured servers (context7, sequential-thinking, pencil, claude-in-chrome). Current config appears clean.

---

## REQ-OC-004: Windows CLAUDE_ENV_FILE Normalization

### Objective

Ensure Windows users get CLAUDE_ENV_FILE injection during SessionStart, equivalent to macOS/Linux behavior.

### Implementation Details

#### File 23 (P0): `internal/hook/session_start.go`

**Current state**: The `Handle` method performs GLM credential injection, teammate mode detection, tmux env injection, telemetry pruning, and skill symlink creation. None of these branches have Windows-specific CLAUDE_ENV_FILE handling.

**Target**: Add a new function `ensureCLAUDEEnvFile()` that:
1. Checks if `CLAUDE_ENV_FILE` environment variable is set
2. If not set and running on Windows (`runtime.GOOS == "windows"`):
   - Looks for `.claude/.env` in the project directory
   - If found, sets `CLAUDE_ENV_FILE` to that path
3. Returns a status message for the session data

**Integration point** (after line 96, before telemetry pruning):
```go
// Ensure CLAUDE_ENV_FILE is set on Windows (REQ-OC-004)
if input.ProjectDir != "" && runtime.GOOS == "windows" {
    if msg := ensureCLAUDEEnvFile(input.ProjectDir); msg != "" {
        data["claude_env_file"] = msg
        slog.Info("CLAUDE_ENV_FILE injected for Windows", "message", msg)
    }
}
```

#### File 24 (P0): `internal/hook/session_start_test.go`

Add test cases for the new `ensureCLAUDEEnvFile` function:
- Test: Windows with missing env file variable
- Test: Non-Windows platforms skip injection
- Test: Already-set variable is not overwritten

---

## REQ-OC-005: Remove Legacy Defensive Scaffolding

### Objective

Remove redundant defensive instructions that Opus 4.7 handles natively. Opus 4.7's Adaptive Thinking and improved instruction following make certain explicit guards unnecessary.

### Candidates for Removal/Simplification

#### File 25 (P0): `internal/template/templates/.claude/output-styles/moai/moai.md`

**Dark Flow Warning section**: Opus 4.7 natively detects and self-corrects "productive feeling, broken output" patterns. Simplify from a detailed warning to a brief note:

Before: Full paragraph about dark-flow detection and verification intensity escalation.
After: Single line referencing Opus 4.7 native verification behavior.

**Over-engineering Guard**: Opus 4.7 is less prone to bloat than 4.6. Simplify the explicit guard from "Opus 4.6 tends toward bloat; push back explicitly" to a model-neutral "reject unrequested complexity."

#### File 26 (P1): `internal/template/templates/.claude/rules/moai/core/moai-constitution.md`

**Agent Core Behavior #3 (Push Back)**: The anti-sycophancy rule stays but can be simplified. Opus 4.7 has improved sycophancy resistance. Remove the anti-pattern example detail, keep the rule itself.

**Agent Core Behavior #4 (Enforce Simplicity)**: Keep the rule. Remove the "natural tendency of code generation toward over-engineering" framing since Opus 4.7 is better calibrated.

#### File 27 (P2): `internal/template/templates/.claude/rules/moai/workflow/file-reading-optimization.md`

Review whether the aggressive token budget warnings are still necessary with Opus 4.7's improved context efficiency. Simplify if appropriate while keeping the practical tiered loading rules.

---

## Priority Summary

### P0 (Critical Path) -- 15 files

| # | File | REQ | Change Type |
|---|------|-----|-------------|
| 1 | `templates/.claude/rules/moai/development/model-policy.md` | OC-001 | Model generation, effort levels |
| 2 | `templates/.claude/rules/moai/development/agent-authoring.md` | OC-001 | effort field: add xhigh |
| 3 | `templates/.moai/config/sections/llm.yaml` | OC-001 | Comment update |
| 9 | `templates/CLAUDE.md` | OC-002 | Section 12 Adaptive Thinking |
| 10 | `templates/.claude/rules/moai/core/moai-constitution.md` | OC-002 | Philosophy integration |
| 11 | `templates/.claude/rules/moai/core/agent-common-protocol.md` | OC-002 | Prompt efficiency section |
| 12 | `templates/.claude/skills/moai-workflow-thinking/SKILL.md` | OC-002 | Thinking mode clarification |
| 17 | `templates/.claude/rules/moai/core/hooks-system.md` | OC-003 | v2.1.110 runtime docs |
| 18 | `templates/.claude/rules/moai/core/settings-management.md` | OC-003 | bypassPermissions policy |
| 19 | `templates/.claude/settings.json.tmpl` | OC-003 | Timeout ceiling comment |
| 23 | `internal/hook/session_start.go` | OC-004 | Windows env file injection |
| 24 | `internal/hook/session_start_test.go` | OC-004 | Test cases |
| 25 | `templates/.claude/output-styles/moai/moai.md` | OC-005 | Dark flow, over-eng removal |
| -- | `make build` | ALL | Regenerate embedded templates |
| -- | `go test ./...` | ALL | Verify all tests pass |

### P1 (Important) -- 8 files

| # | File | REQ | Change Type |
|---|------|-----|-------------|
| 4 | `templates/.claude/output-styles/moai/moai.md` | OC-001 | Model identity update |
| 5 | `templates/.claude/skills/moai/workflows/run.md` | OC-001 | Opus version refs |
| 6 | `templates/.claude/agents/moai/builder-agent.md` | OC-001 | Agent Opus refs |
| 13 | `templates/.claude/rules/moai/workflow/spec-workflow.md` | OC-002 | Model behavior refs |
| 14 | `templates/.claude/skills/moai-foundation-core/SKILL.md` | OC-002 | Core skill update |
| 15 | `templates/.claude/skills/moai-foundation-cc/SKILL.md` | OC-002 | CC skill update |
| 16 | `templates/.claude/skills/moai-foundation-cc/reference/claude-code-cli-reference-official.md` | OC-003 | CLI ref update |
| 26 | `templates/.claude/rules/moai/core/moai-constitution.md` | OC-005 | Simplify behaviors |

### P2 (Nice to Have) -- 4 files

| # | File | REQ | Change Type |
|---|------|-----|-------------|
| 7 | `templates/.claude/rules/moai/development/skill-authoring.md` | OC-001 | Effort docs |
| 8 | `internal/cli/launcher.go` | OC-001 | Profile review |
| 22 | `templates/.mcp.json.tmpl` | OC-003 | Scope dedup verify |
| 27 | `templates/.claude/rules/moai/workflow/file-reading-optimization.md` | OC-005 | Token warning review |

---

## Execution Order

```
Phase 1: REQ-OC-001 (Model catalog) -- P0 files [1, 2, 3]
Phase 2: REQ-OC-002 (Philosophy)    -- P0 files [9, 10, 11, 12]
Phase 3: REQ-OC-003 (Runtime)       -- P0 files [17, 18, 19]
Phase 4: REQ-OC-004 (Windows)       -- P0 files [23, 24]
Phase 5: REQ-OC-005 (Scaffolding)   -- P0 file  [25]
Phase 6: P1 batch                   -- files [4-6, 13-16, 20-21, 26]
Phase 7: P2 batch                   -- files [7-8, 22, 27]
Phase 8: make build + go test ./... -- validation gate
```

---

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| CLAUDE.md exceeds 40KB limit | Build break | Measure before/after; move detail to rule files |
| Template/local divergence | User confusion | Document in CHANGELOG; `moai update` handles sync |
| Opus 4.6 regression | Existing users | Keep backward-compatible effort values; test EC-1 |
| Windows env file edge cases | Platform break | Test with `t.TempDir()`; handle missing .env gracefully |
| GLM effort incompatibility | GLM users | Skip xhigh for GLM models; test EC-2 |

---

## Dependencies

- Claude Code v2.1.110+ release notes (verified)
- Opus 4.7 prompt philosophy documentation (Anthropic verified)
- Existing test suite green (`go test ./...`)
- `make build` succeeds with current templates

---

Version: 1.0.0
Created: 2026-04-17
Plan-auditor: PASS (iteration 2, 7/7 must-pass)
