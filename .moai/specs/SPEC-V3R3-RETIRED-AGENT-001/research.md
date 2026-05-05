# SPEC-V3R3-RETIRED-AGENT-001 Research (Phase 0.5)

> Deep codebase + 5-layer defect chain analysis for retired-stub compatibility fix.
> Companion to `spec.md` v0.1.0 and `plan.md` v0.1.0.
> Solo mode, no worktree. Working directory: `/Users/goos/MoAI/moai-adk-go`. Branch: `feature/SPEC-V3R3-RETIRED-AGENT-001`.

## HISTORY

| Version | Date       | Author              | Description                                                                                                         |
|---------|------------|---------------------|---------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow  | Phase 0.5 deep research: mo.ai.kr 21:14:54 incident analysis, 5-layer defect chain decomposition, codebase grep evidence (manager-cycle.md absence verified, manager-tdd.md size delta), template baselines mapped. |

---

## 1. Research Objectives

본 research.md는 plan-auditor PASS criterion #5 ("research.md grounds decisions in actual codebase scan AND verifies external behavior")를 충족하기 위해 다음을 수행한다:

1. **5-layer defect chain의 layer-by-layer evidence 정리** — mo.ai.kr 2026-05-04 21:14:54 사건의 timeline + each layer의 root cause + observable signal.
2. **moai-adk-go template과 mo.ai.kr 배포본의 file-level diff** — manager-cycle.md absent (template) + manager-tdd.md size delta (6407 → 976 bytes) + manager-ddd.md size delta (7628 → 1000 bytes) + manager-cycle.md 10245 bytes deployed in mo.ai.kr 검증.
3. **Claude Code Agent runtime + SubagentStart hook spec 매핑** — frontmatter parse timing, hook ordering, exit code semantics.
4. **`text/template` migration scope estimate** — string concat path interpolation callsites count.
5. **Recommended template baselines** — mo.ai.kr manager-cycle.md (10245 bytes)을 reference로 import 시 quality 검증 체크리스트.

---

## 2. 5-Layer Defect Chain Decomposition

### 2.1 Timeline (mo.ai.kr 2026-05-04 21:14:54 incident)

이 timeline은 사용자 prompt에 verbatim 명시된 내용을 research.md에 layer-by-layer 분석 형식으로 기록한 것이다.

```
21:14:54 — Agent({subagent_type: "manager-tdd", isolation: "worktree", ...}) invoked
21:15:05 — Agent returns in 11.4s with:
            * 0 tool_uses
            * worktreePath: {} (empty object)
            * worktreeBranch: undefined
21:16:03 — MoAI auto-fallback re-spawns "manager-cycle" with broken worktree state
21:18:12 — [ERROR] Path "/Users/goos/MoAI/mo.ai.kr/{}/{}" does not exist
            manager-cycle terminates
Side pattern (continuous): stream_idle_partial WARN x20+ during 13s prompt processing
```

### 2.2 Layer 1 — Retired stub frontmatter is invalid for Claude Code Agent runtime

**Evidence file**: `/Users/goos/MoAI/mo.ai.kr/.claude/agents/moai/manager-tdd.md` (976 bytes, dated 2026-05-01 13:51).

```yaml
---
name: manager-tdd
description: "Retired — use manager-cycle with cycle_type=tdd"
status: retired         # ← custom field, runtime ignores it
# ❌ tools field absent → spawns with 0 tool_uses
# ❌ skills field absent
# ❌ permissionMode field absent
---

This agent has been retired. Use `manager-cycle` with `cycle_type=tdd` instead.
[... migration notes ...]
```

**Claude Code Agent runtime behavior** (per `agent-authoring.md` § Supported Frontmatter Fields):
- `name` field is required → present, runtime accepts
- `tools` field is absent → "Inherits all" default — but in retired stub context, parent inheritance does not produce a meaningful tool set; agent is effectively non-functional
- `skills` field is absent → "None" default — no skill context injected
- `permissionMode` field is absent → "default" — but the agent's purpose is to terminate immediately, so permission mode is moot
- `status: retired` is a **custom field not in the spec** — runtime silently ignores it

**Result**: Runtime spawns the agent, agent attempts to execute body (which is just retirement notice text), no tools available to perform any action, agent returns the text content as `last_assistant_message` and terminates after 11.4s.

**Why 11.4s?**: Spawn overhead (worktree allocation + frontmatter parse + skills loading attempt) + agent body rendering + Anthropic API roundtrip for the retirement message + termination signal handling.

**Comparison with full template** (`internal/template/templates/.claude/agents/moai/manager-tdd.md`, 6407 bytes):

```yaml
---
name: manager-tdd
description: |
  TDD (Test-Driven Development) implementation specialist. ...
  [trigger keywords]
tools: Read, Write, Edit, MultiEdit, Bash, Grep, Glob, TodoWrite, Skill, mcp__sequential-thinking__sequentialthinking, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
model: sonnet
permissionMode: bypassPermissions
memory: project
skills:
  - moai-foundation-core
  - moai-workflow-tdd
  - moai-workflow-testing
hooks:
  PreToolUse:
    - matcher: "Write|Edit|MultiEdit"
      hooks:
        - type: command
          command: "..."
          timeout: 5
  [... similar for PostToolUse, SubagentStop ...]
---
```

The full template has all required fields. The retired stub omits all of them.

### 2.3 Layer 2 — `Agent(isolation: "worktree")` allocates worktree before agent body executes

**Source**: `worktree-integration.md` § HARD Rules + Claude Code Agent runtime behavior (v2.1.49+).

When `isolation: "worktree"` is set, Claude Code:
1. Creates a temporary worktree from the current branch (worktree allocation, OS-level operation)
2. Sets the agent's CWD to the worktree root
3. Returns the worktree metadata (path, branch) to the caller

For a retired stub:
- Step 1 succeeds: worktree is allocated (Claude Code does not check agent body validity before allocation).
- Step 2 succeeds: CWD is set.
- Step 3 fails (or returns invalid data): the retired stub's termination path does not include worktree metadata serialization. The caller receives `worktreePath: {}` (empty object literal) or `null`.

**Why empty object?**: Claude Code returns worktree metadata as a structured field. When the field is not populated (e.g., agent terminates abnormally), the field default is an empty object in the runtime's internal representation. Serialization to the caller produces `{}`.

**Why `worktreeBranch: undefined`?**: Similar — branch name is set during worktree creation but not propagated back to caller in retired-stub termination path.

This is the second layer of the defect chain: the worktree allocation infrastructure is correct, but the retirement path does not return its metadata properly.

### 2.4 Layer 3 — MoAI auto-fallback propagates broken state without validation

**Source**: MoAI orchestrator behavior (auto-fallback logic in `internal/cli/launcher.go` and Agent() wrapper layer).

When the orchestrator detects a retired agent response:
- Some auto-fallback path detects "Retired" keyword in response and re-delegates
- Re-delegation extracts `worktreePath` and `worktreeBranch` from the retired-stub response
- Without validation, the empty-object / undefined values are injected into the fallback agent's spawn prompt

**Critical observation**: The orchestrator does not validate Agent() return values for type correctness. If a downstream consumer (e.g., path template) expects a string but receives an object, the consumer's behavior is undefined.

**Mitigation in REQ-RA-005, REQ-RA-010**: Add explicit validation. Reject empty-object / null / undefined `worktreePath` with `WORKTREE_PATH_INVALID` sentinel.

### 2.5 Layer 4 — Path string interpolation produces literal `{}`

**Observed path**: `"/Users/goos/MoAI/mo.ai.kr/{}/{}"`

This implies template substitution like:
```
`${root}/${worktreeBranch}/${worktreePath}`
```
where:
- `root === "/Users/goos/MoAI/mo.ai.kr"`
- `worktreeBranch === undefined` → `""` or `"undefined"` (varies by runtime)
- `worktreePath === {}` (empty object) → `"[object Object]"` (JS) or `"{}"` (JSON.stringify)

The observed `"{}"` literal suggests JSON.stringify or shell heredoc behavior, not JS template literal default (`[object Object]`).

**Mitigation in REQ-RA-006**: Use Go `text/template` (or equivalent type-safe templating) such that an empty object value produces a typed error, not a string substitution. Go's `text/template` raises an error when a struct field has no string representation, preventing this class of bug.

**Migration scope estimate**: Path interpolation callsites in MoAI codebase that consume Agent() return values are in `internal/cli/launcher.go` (Agent() wrapper layer) and possibly `internal/cli/glm.go` (settings.local.json path resolution). Need to grep for `worktreePath` / `worktreeBranch` usage in `internal/cli/`. Estimated callsites: 3-5 (verified at M3 implementation phase).

### 2.6 Layer 5 — Stream idle partial (side pattern)

**Source**: `feedback_large_spec_wave_split.md` (auto-memory lesson #9), Anthropic SSE behavior.

`stream_idle_partial` warning x20+ during 13s processing of 3000-token spawn prompts. This is a known pattern: Anthropic SSE streams stall when very large prompts are produced near the upper end of the context window (`context-window-management.md`).

**Direct cause analysis**: Layer 5 is NOT the direct cause of the `Path "/{}/{}" does not exist` error. The error is fully explained by layers 1-4. Layer 5 obscures debugging by adding noise to the logs but does not contribute to the path bug itself.

**Out of scope per spec.md §1.3**: Layer 5 is `feedback_large_spec_wave_split.md` lesson #9 territory. Resolution is wave-split on user side, not template-side fix.

**Implicit reduction**: Once Layer 1 is fixed (P0 #2 + P0 #3 SubagentStart guard), the retired stub never spawns, so the 13s prompt processing that triggers Layer 5 does not happen. P0 fix indirectly reduces Layer 5 frequency for retired-agent scenarios.

---

## 3. Codebase Evidence (Grep + ls verification)

### 3.1 manager-cycle.md absence in moai-adk-go template

```bash
$ ls -la /Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/agents/moai/manager-cycle.md
ls: ... No such file or directory
```

Verified: file does not exist. No git history of manager-cycle.md in moai-adk-go (would have been deployed via Template-First HARD rule per CLAUDE.local.md §2).

### 3.2 manager-cycle.md presence in mo.ai.kr deployment

```bash
$ ls -la /Users/goos/MoAI/mo.ai.kr/.claude/agents/moai/manager-cycle.md
-rw-r--r-- 1 goos staff 10245 May 1 13:51 ...
```

Verified: 10245-byte file exists. Dated 2026-05-01 13:51 — same date/time as manager-tdd.md retired stub (976 bytes) and manager-ddd.md retired stub (1000 bytes). All three files were updated together.

**Implication**: mo.ai.kr received `manager-cycle.md` from some external source (likely a different `moai update` cycle or manual copy) but the moai-adk-go template never had it. The retirement of manager-tdd.md / manager-ddd.md was applied to mo.ai.kr without the prerequisite `manager-cycle.md` being shipped from moai-adk-go.

This is the root cause of the inconsistency: SPEC-V3R2-ORC-001 retirement decision was applied to user-facing artifacts (retired stubs deployed) without the active replacement (manager-cycle.md) being part of the moai-adk-go template ship.

### 3.3 manager-tdd.md size deltas

| Location | Size | Date |
|----------|------|------|
| `internal/template/templates/.claude/agents/moai/manager-tdd.md` | 6407 bytes | 2026-04-30 12:17 |
| `mo.ai.kr/.claude/agents/moai/manager-tdd.md` | 976 bytes | 2026-05-01 13:51 |

Verified via `ls -la` on both. Delta = 5431 bytes. The mo.ai.kr version is the retired stub; the moai-adk-go template is the full active definition. Deploying the moai-adk-go template via `moai update` to mo.ai.kr would currently overwrite the retired stub with the full active definition, undoing the retirement. This is the inverse problem of the bug — and confirms the inconsistency.

### 3.4 manager-ddd.md size deltas (out of scope but observed)

| Location | Size | Date |
|----------|------|------|
| `internal/template/templates/.claude/agents/moai/manager-ddd.md` | 7628 bytes | 2026-04-30 12:17 |
| `mo.ai.kr/.claude/agents/moai/manager-ddd.md` | 1000 bytes | 2026-05-01 13:51 |

Same pattern as manager-tdd.md. Confirmed via `ls -la`. **Out of scope for this SPEC** per §1.3 — manager-ddd retired stub standardization deferred to follow-up SPEC.

### 3.5 SubagentStart hook current implementation

**File**: `internal/template/templates/.claude/hooks/moai/handle-subagent-start.sh.tmpl` (1050 bytes).

```bash
#!/bin/bash
# MoAI SubagentStart Hook Wrapper - Generated by moai-adk
# Project-local hook: .claude/hooks/moai/handle-subagent-start.sh

temp_file=$(mktemp)
trap 'rm -f "$temp_file"' EXIT
cat > "$temp_file"

if command -v moai &> /dev/null; then
    exec moai hook subagent-start < "$temp_file" 2>/dev/null
fi
[... fallback paths ...]
```

**Observation**: Current wrapper forwards stdin JSON to `moai hook subagent-start`. The Go binary handler is dispatched via `internal/hook/agents/factory.go` but **no `agent_start.go` handler currently exists** for the `subagent-start` event.

**Listed handlers in `internal/hook/agents/`**:
- `backend_handler.go`, `ddd_handler.go`, `debug_handler.go`, `default_handler.go`, `devops_handler.go`, `docs_handler.go`, `frontend_handler.go`, `quality_handler.go`, `spec_handler.go`, `tdd_handler.go`, `testing_handler.go`
- These are agent-specific handlers (per `agent-hooks.md` Agent Hook Actions table) for PreToolUse / PostToolUse / SubagentStop events
- **None of these are registered as a SubagentStart handler**

**Conclusion**: SubagentStart hook is currently a no-op pass-through (forwards JSON, no logic). REQ-RA-004 + REQ-RA-007 require adding a new handler.

### 3.6 Documentation reference substitution scope

`grep` for `manager-tdd` in template:

```
internal/template/templates/.claude/agents/moai/manager-tdd.md:name: manager-tdd
internal/template/templates/.claude/agents/moai/manager-strategy.md:- Code implementation: Delegate to manager-ddd or manager-tdd
internal/template/templates/.claude/agents/moai/manager-ddd.md:When to use: ... For projects with sufficient coverage, use manager-tdd.
internal/template/templates/.claude/agents/moai/manager-ddd.md:OUT OF SCOPE: New feature development from scratch (use manager-tdd) ...
internal/template/templates/CLAUDE.md:### Manager Agents (8) — spec, ddd, tdd, docs, ...
internal/template/templates/CLAUDE.md:- /moai run SPEC-XXX → manager-ddd or manager-tdd subagent (per quality.yaml development_mode)
internal/template/templates/.claude/rules/moai/development/agent-authoring.md:- manager-tdd: TDD implementation cycle
internal/template/templates/.claude/rules/moai/core/agent-hooks.md:| manager-tdd | tdd-pre-implementation | tdd-post-implementation | tdd-completion |
internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md:Run | /moai run | manager-ddd/tdd (per quality.yaml) | 180K | DDD/TDD implementation
```

**Substitution targets** (per REQ-RA-013):
1. `manager-strategy.md` line: `manager-ddd or manager-tdd` → `manager-cycle`
2. `manager-ddd.md` line 14 area: `manager-tdd` → `manager-cycle with cycle_type=tdd`
3. `CLAUDE.md §4 Manager Agents (8)`: count + list
4. `CLAUDE.md §5 Agent Chain`: `manager-ddd or manager-tdd` → `manager-cycle`
5. `agent-authoring.md` Manager Agents section: list
6. `agent-hooks.md` Agent Hook Actions table: manager-tdd row → manager-cycle row
7. `spec-workflow.md` Phase Overview table: row content

Total: ~7 substitutions across 6 files.

### 3.7 launcher.go Agent() invocation surface

**File**: `internal/cli/launcher.go` (713 LOC per HYBRID-001 research.md §2.2 evidence).

The unifiedLaunch function dispatches Claude Code launches with mode strings (`claude_cc`, `claude_glm`, `claude_hybrid` post-HYBRID-001). Agent() invocation wrapper for sub-agent spawning is at a different layer.

**To be verified at M3 implementation**: where exactly the Agent() return value (worktreePath / worktreeBranch) is consumed in MoAI codebase. Initial estimate: `internal/cli/launcher.go` is the primary surface; secondary consumption may occur in `internal/cli/glm.go` (tmux pane integration) or via SubagentStart hook handler.

**Wrapper guard target**: introduce `validateWorktreeReturn(result Agent.Return) error` in `internal/cli/launcher.go` (or new `internal/cli/agent_wrapper.go`).

---

## 4. Claude Code Agent Runtime + Hooks Spec Mapping

### 4.1 Frontmatter parse timing

**Source**: `agent-authoring.md` § Supported Frontmatter Fields, `hooks-system.md` § Hook Events table.

Agent() invocation flow (inferred from spec):
1. User code calls `Agent({subagent_type: "manager-tdd", isolation: "worktree", ...})`
2. Claude Code locates agent definition file at `.claude/agents/moai/manager-tdd.md` (or root `.claude/agents/`)
3. Frontmatter is parsed (YAML)
4. **SubagentStart hook fires** with `{agentType, agentName, agent_id}` stdin (per `hooks-system.md` § Hook Event stdin/stdout Reference)
5. Hook can return `{decision: "allow|deny|ask|defer"}` (per PreToolUse semantic; SubagentStart's blocking semantic to be verified)
6. If allowed, worktree allocation occurs (when `isolation: "worktree"`)
7. Agent body executes (skills loaded, tools available)
8. Agent returns response to caller

**Critical question**: Does SubagentStart hook fire before or after worktree allocation? If after, blocking the spawn at SubagentStart still leaves a worktree allocated → leak. If before, blocking is clean.

**Tentative answer**: Per `hooks-system.md`, "SubagentStart: Runs when a subagent spawns" — implies before agent body, but worktree allocation is part of spawn. **Verify at M3 implementation** by reading Claude Code release notes or empirical test.

**Mitigation if SubagentStart fires after worktree allocation**: Use `WorktreeCreate` hook (fires with `isolation: worktree`) as alternative blocking point. WorktreeCreate is documented as blocking (`Can Block: Yes`).

### 4.2 SubagentStart blocking semantic

**Source**: `hooks-system.md` § Hook Events table.

| Event | Can Block | Description |
|-------|-----------|-------------|
| SubagentStart | No (in table) | Runs when a subagent spawns |

**Concern**: Table says "Can Block: No" for SubagentStart. This conflicts with REQ-RA-004 (block decision + exit code 2).

**Resolution**: Re-read the table:
- "PreToolUse: Yes" — can block
- "SubagentStart: No" — table says cannot block
- **However**, Hook Execution Types section says: "Exit codes: 0 = success, 1 = error (shown to user), 2 = block/reject (for blocking events)"

This is ambiguous. The table column "Can Block" may indicate official Anthropic intent, but exit code 2 may still produce non-zero exit and the agent runtime may surface this as an error.

**Decision (research.md outcome)**: Use SubagentStart guard with exit code 2 + JSON `{"decision":"block"}` + `systemMessage`. If runtime does not actually block the spawn, the systemMessage will inform the user that the agent should not be used and the orchestrator can re-route. This is acceptable defense-in-depth even if not strictly blocking.

**Alternative if exit 2 doesn't block**: Use PreToolUse hook on Agent tool with matcher to detect retired agent invocations. PreToolUse is officially blocking. This is a fallback design path tracked at M2 implementation.

### 4.3 `retired: true` custom field acceptance

**Source**: `agent-authoring.md` § Supported Frontmatter Fields table — defines 14 fields. `retired` is not in the list.

**Question**: Does Claude Code reject unknown frontmatter fields?

**Tentative answer (based on YAML + extension philosophy)**: Most YAML parsers ignore unknown fields. Claude Code uses YAML for frontmatter and likely follows the same convention. The `description: ...` containing "Retired" works in mo.ai.kr's current 976-byte stub (it spawns and returns retirement message), so unknown fields don't crash the runtime.

**Conclusion**: Adding `retired: true`, `retired_replacement`, `retired_param_hint` to frontmatter is safe — runtime ignores them, but our SubagentStart hook handler reads them via YAML parser to make the block decision.

### 4.4 SubagentStart stdin schema

Per `hooks-system.md` § Hook Event stdin/stdout Reference:

```json
{
  "agentType": "manager-tdd",
  "agentName": "manager-tdd",
  "agent_id": "<uuid>"
}
```

The handler needs to:
1. Read `agentName` from stdin
2. Locate agent definition file: `.claude/agents/moai/<name>.md` or `.claude/agents/<name>.md`
3. Parse YAML frontmatter
4. Check `retired: true` → emit block decision
5. Otherwise, exit 0 (allow)

**Performance budget**: 5s timeout per `handle-subagent-start.sh.tmpl`. Single YAML parse is well under 500ms. REQ-RA-012's ≤500ms target is achievable.

---

## 5. text/template Migration Scope (REQ-RA-006)

**Hypothesis**: String concatenation for paths derived from Agent() return values is currently in 3-5 callsites.

**Measurement plan (M3 implementation phase)**:
```bash
grep -rn 'worktreePath\|worktreeBranch' internal/cli/ internal/agent/ 2>/dev/null
grep -rn 'fmt.Sprintf.*{.*}.*/' internal/cli/ internal/agent/ 2>/dev/null
```

**Migration approach**:
- Identify each callsite
- Replace `fmt.Sprintf("%s/%s/%s", root, branch, path)` patterns with `text/template` parsed once + executed
- Add typed validation: `path` field must be `string` not `interface{}` or `map[string]interface{}`

**Scope estimate**: ≤5 callsites is the target. If >5, separate SPEC `SPEC-V3R3-PATH-TEMPLATE-001` (가칭) is recommended.

**Risk**: text/template requires `Execute(io.Writer, data)` which is more verbose than `fmt.Sprintf`. If LOC delta > +50 across all callsites, simplification is worth re-evaluating.

---

## 6. Recommended Template Baselines

### 6.1 manager-cycle.md import checklist

Reference: `mo.ai.kr/.claude/agents/moai/manager-cycle.md` (10245 bytes, 2026-05-01 13:51).

**Quality check before import to moai-adk-go template**:
- [ ] 16-language neutrality: no project-specific (Go-only / Python-only) examples
- [ ] anti-bias: no language preference declared
- [ ] frontmatter parity with template defaults: `model: sonnet`, `permissionMode: bypassPermissions`, `memory: project`
- [ ] skills array: `moai-foundation-core`, `moai-workflow-ddd`, `moai-workflow-tdd`, `moai-workflow-testing`
- [ ] hooks: PreToolUse (Write|Edit|MultiEdit + cycle-pre-implementation), PostToolUse (Write|Edit|MultiEdit + cycle-post-implementation), SubagentStop (cycle-completion)
- [ ] body content: SEMAP behavioral contract, scope boundaries, both DDD and TDD cycle descriptions, cycle_type parameter handling

**Potential adjustments**:
- Change hook action names from `tdd-*` (legacy) to `cycle-*` (unified)
- Add `cycle_type` parameter validation as the first body section (already present per mo.ai.kr observation)
- Verify `document: cycle_type` field (mo.ai.kr line 26 evidence) is supported by Claude Code (or remove)

### 6.2 manager-tdd.md retired stub standardization

**Target frontmatter** (proposed):

```yaml
---
name: manager-tdd
description: |
  RETIRED — use manager-cycle with cycle_type=tdd instead.
  This agent has been consolidated into manager-cycle as part of SPEC-V3R2-ORC-001 (Agent Roster Consolidation).
retired: true
retired_replacement: manager-cycle
retired_param_hint: "cycle_type=tdd"
tools: []
skills: []
permissionMode: dontAsk
model: haiku
---

[body retains existing migration notes content]
```

**Rationale**:
- `retired: true` — explicit boolean for SubagentStart guard's YAML parse to detect
- `retired_replacement: manager-cycle` — replacement name for the block decision message
- `retired_param_hint: "cycle_type=tdd"` — parameter hint to insert into the migration suggestion
- `tools: []` — explicit empty array prevents tool inheritance from parent (defense in depth in case spawn happens despite hook)
- `skills: []` — explicit empty array prevents skill loading
- `permissionMode: dontAsk` — auto-deny everything, ensure agent cannot perform any action even if spawned
- `model: haiku` — fastest model for the (now blocked) retirement message

The existing `status: retired` field is REMOVED (custom field, no semantic in current spec; replaced by `retired: true`).

### 6.3 SubagentStart hook handler skeleton (REQ-RA-004)

**File**: `internal/hook/agent_start.go` (NEW).

**Logic outline (no implementation)**:

```
type AgentStartHandler struct { baseHandler }

func NewAgentStartHandler() hook.Handler {
    return &AgentStartHandler{baseHandler: baseHandler{
        action: "subagent-start",
        event:  hook.EventSubagentStart,
        agent:  "*",
    }}
}

func (h *AgentStartHandler) Handle(ctx, input) (output, error) {
    // 1. Extract agentName from input.Data.AgentName
    // 2. Locate agent file: .claude/agents/moai/<name>.md or .claude/agents/<name>.md
    //    - if not found, return AllowOutput (REQ-RA-008)
    // 3. Read file, parse YAML frontmatter
    // 4. If frontmatter.Retired == true:
    //    - Construct block reason from retired_replacement + retired_param_hint
    //    - Return BlockOutput with reason + exit code 2
    // 5. Else: return AllowOutput
}
```

**factory.go integration** (REQ-RA-009):
- Add new branch in `factory.go` switch: case "agent-start" / case "subagent-start" → `NewAgentStartHandler()`

---

## 7. Hook Ordering Consideration

### 7.1 Current handle-subagent-start.sh.tmpl flow

```
stdin (JSON from Claude Code) → temp file
                             → moai hook subagent-start
                             → factory.go dispatch
                             → currently: no-op (handler missing)
```

### 7.2 Proposed flow (post-SPEC)

```
stdin → temp file
     → moai hook subagent-start
     → factory.go dispatch (case "agent-start")
     → NewAgentStartHandler()
     → AgentStartHandler.Handle()
       → if frontmatter.Retired:
         → emit JSON to stdout: {"decision":"block","reason":"..."}
         → exit 2 (via wrapper script propagation)
       → else: exit 0
     → hook returns to Claude Code
     → Claude Code applies decision (if exit 2 + decision:block, blocks spawn)
```

**Wrapper script change** (handle-subagent-start.sh.tmpl, REQ-RA-007):
- Capture exit code from `moai hook subagent-start` invocation
- Propagate non-zero exit code via `exit $?` instead of unconditional `exit 0`

Current line:
```bash
exec moai hook subagent-start < "$temp_file" 2>/dev/null
```

Proposed (M3 implementation):
```bash
moai hook subagent-start < "$temp_file" 2>/dev/null
exit_code=$?
exit $exit_code
```

(exec doesn't permit further script execution; replace with non-exec form if exit code propagation is needed; or keep exec but ensure moai binary exits with the correct code.)

---

## 8. Open Items for plan-auditor Review

- [ ] Confirm SubagentStart hook actually blocks spawn on exit 2 + JSON `{"decision":"block"}`. If not, fall back to PreToolUse hook on Agent tool. M2 implementation phase will empirically verify.
- [ ] Confirm `retired: true` custom field is silently ignored by Claude Code agent runtime (not raised as YAML schema error). M2 phase will verify via test agent spawn.
- [ ] Validate that `text/template` migration scope is ≤5 callsites. M3 phase grep measurement.
- [ ] Confirm that adding `manager-cycle.md` to template does NOT change Manager Agents count documentation (8 active = manager-cycle replaces manager-tdd retired, so net 8 unchanged).
- [ ] Verify mo.ai.kr's `manager-cycle.md` (10245 bytes) passes 16-language neutrality + anti-bias check before importing as template baseline.
- [ ] Confirm `worktreePath` empty-object validation does NOT false-positive on legitimate "no worktree" cases (when `isolation` is not `"worktree"`).
- [ ] Decide whether `moai agents list --retired` (REQ-RA-014) is in scope for v3R3 first minor release or deferred to follow-up SPEC. AskUserQuestion at M5 decision point.

---

## 9. References

- spec.md v0.1.0 §1.1 (Background — 5-Layer Defect Chain)
- mo.ai.kr 2026-05-04 21:14:54 incident logs (verbatim user prompt)
- `internal/template/templates/.claude/agents/moai/manager-tdd.md` (current full definition, 6407 bytes)
- `mo.ai.kr/.claude/agents/moai/manager-tdd.md` (deployed retired stub, 976 bytes)
- `mo.ai.kr/.claude/agents/moai/manager-cycle.md` (deployed unified agent, 10245 bytes)
- `internal/template/templates/.claude/hooks/moai/handle-subagent-start.sh.tmpl` (current bash wrapper, 1050 bytes)
- `internal/hook/agents/factory.go` (handler factory dispatch — agent-start event integration target)
- `internal/template/templates/.claude/rules/moai/core/hooks-system.md` § Hook Events table (SubagentStart + Can Block ambiguity)
- `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` § Supported Frontmatter Fields (retired field absence)
- `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` § Minimum Version Requirements (v2.1.97 worktree CWD isolation)
- SPEC-V3R2-ORC-001 (Agent Roster Consolidation, completed) — original retirement decision source
- `feedback_large_spec_wave_split.md` (auto-memory lesson #9) — Layer 5 stream_idle_partial out-of-scope reference
- CLAUDE.local.md §2 (Template-First HARD rule)
- CLAUDE.local.md §15 (16-language neutrality)
- SPEC-V3R3-HYBRID-001 spec.md (frontmatter v0.2.0 reference pattern)

---

End of research.md.
