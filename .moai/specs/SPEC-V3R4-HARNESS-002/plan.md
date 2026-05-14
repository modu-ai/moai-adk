# Implementation Plan — SPEC-V3R4-HARNESS-002

This document is the Wave-level implementation plan for the Multi-Event Observer Expansion SPEC. All priorities use P0/P1/P2/P3 labels per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation; no time estimates are used. All file references use project-root-relative paths.

---

## 1. Overview

V3R4-002 extends the PostToolUse-only observer baseline (V3R4-HARNESS-001) to cover three additional Claude Code hook events — `Stop`, `SubagentStop`, `UserPromptSubmit` — with a unified, additive JSONL schema. The plan decomposes into three Waves (A through C). Each Wave is independently PR-able; orchestrator may sequence them as separate squash PRs OR bundle them as a single plan-phase deliverable depending on review bandwidth.

| Wave | Title | Priority | Owning Run-Phase Agent | Acceptance Gate |
|------|-------|----------|------------------------|-----------------|
| Wave A | Schema extension + cobra subcommand scaffolding | P0 | `manager-develop` (delegating to `expert-backend`) | EventType enum has 3 new values; Event struct has all optional fields; 3 cobra subcommands registered; isHarnessLearningEnabled gate reused across all 4 handlers. |
| Wave B | Wrapper script templates + settings.json.tmpl registration | P0 | `manager-develop` (delegating to `expert-devops` for template-first compliance) | 3 new `.sh.tmpl` files created; `make build` regenerates `embedded.go`; settings.json.tmpl has 3 additive entries; existing entries untouched. |
| Wave C | Test suite + 5-Layer Safety integration verification | P1 | `manager-develop` (delegating to `expert-backend` for Go test authoring) | 9 new test cases pass (3 events × {NoOp, Records, PreservesExisting}); PII fail-open test passes; gate-uniformity test passes. |

All three Waves are documented here as the run-phase roadmap. This `/moai plan SPEC-V3R4-HARNESS-002` invocation produces the five SPEC artifacts only (research.md, spec.md, plan.md, acceptance.md, tasks.md); run-phase work happens in `/moai run SPEC-V3R4-HARNESS-002` after this plan PR merges.

---

## 2. Architecture Overview

### 2.1 Current State (pre-V3R4-002, V3R4-001 baseline)

```
+-------------------------+         +-------------------------+
| Claude Code runtime     |         | settings.json (rendered)|
| - PostToolUse fires     |-------->| PostToolUse: handle-    |
|   on Write|Edit         |         |   post-tool.sh          |
| - Stop fires per turn   |         | Stop: handle-stop.sh    |
| - SubagentStop fires    |         | SubagentStop: handle-   |
|   per Agent() exit      |         |   subagent-stop.sh      |
| - UserPromptSubmit      |         | UserPromptSubmit:       |
|   fires per prompt      |         |   handle-user-prompt-   |
+-------------------------+         |   submit.sh             |
                                    +-----------+-------------+
                                                |
                                                v
                              +-------------------------------+
                              | moai CLI binary               |
                              | hookCmd subcommands:          |
                              | - post-tool (no harness wire) |
                              | - stop                        |
                              | - subagent-stop               |
                              | - user-prompt-submit          |
                              | - harness-observe (PostToolUse|
                              |   observer, UNINVOKED gap)    |
                              +-----------+-------------------+
                                          |
                                          v
                              +-------------------------------+
                              | .moai/harness/usage-log.jsonl |
                              | (currently EMPTY because      |
                              |  harness-observe wrapper is   |
                              |  not registered in            |
                              |  settings.json.tmpl)          |
                              +-------------------------------+
```

The V3R4-001 baseline shipped `runHarnessObserve` and the wrapper script `handle-harness-observe.sh.tmpl`, but did NOT register the wrapper in settings.json.tmpl. That gap is OUT OF SCOPE for V3R4-002 (research.md §2.5 OQ-3) and tracked as a V3R4-001 follow-up.

### 2.2 Target State (post-V3R4-002)

```
+-------------------------+         +-------------------------+
| Claude Code runtime     |         | settings.json (rendered)|
| - same 4 events as      |-------->| PostToolUse:            |
|   before                |         |   handle-post-tool.sh   |
+-------------------------+         |                         |
                                    | Stop:                   |
                                    |   handle-stop.sh        |
                                    |   handle-harness-       | <- NEW (additive)
                                    |     observe-stop.sh     |
                                    |                         |
                                    | SubagentStop:           |
                                    |   handle-subagent-stop.sh|
                                    |   handle-harness-       | <- NEW (additive)
                                    |     observe-subagent-   |
                                    |     stop.sh             |
                                    |                         |
                                    | UserPromptSubmit:       |
                                    |   handle-user-prompt-   |
                                    |     submit.sh           |
                                    |   handle-harness-       | <- NEW (additive)
                                    |     observe-user-prompt-|
                                    |     submit.sh           |
                                    +-----------+-------------+
                                                |
                                                v
                              +-------------------------------+
                              | moai CLI binary               |
                              | hookCmd subcommands:          |
                              | - harness-observe (existing,  |
                              |   PostToolUse pattern)        |
                              | - harness-observe-stop        | <- NEW
                              | - harness-observe-subagent-   | <- NEW
                              |     stop                      |
                              | - harness-observe-user-prompt-| <- NEW
                              |     submit                    |
                              | (all four reuse               |
                              |  isHarnessLearningEnabled)    |
                              +-----------+-------------------+
                                          |
                                          v
                              +-------------------------------+
                              | .moai/harness/usage-log.jsonl |
                              | (extended JSONL schema —      |
                              |  optional fields per event;   |
                              |  baseline 4 fields always     |
                              |  present)                     |
                              +-------------------------------+

+-----------------------------------------------------------+
| 5-Layer Safety (UNCHANGED, REQ-HRN-FND-005 preserved)     |
| L1 Frozen Guard | L2 Canary | L3 Contradiction | L4 Rate  |
| Limiter         | L5 Human Oversight (AskUserQuestion)    |
+-----------------------------------------------------------+
```

Key invariants:
- The four cobra subcommands all share `isHarnessLearningEnabled` (REQ-HRN-OBS-008).
- The settings.json.tmpl change is ADDITIVE (REQ-HRN-OBS-017): existing handle-*.sh entries are untouched.
- The Event struct schema is ADDITIVE (REQ-HRN-OBS-009): all new fields are `omitempty`.
- No new AskUserQuestion call sites (REQ-HRN-OBS-011 preserves REQ-HRN-FND-015).

### 2.3 Per-Event Schema Mapping

| Runtime event | EventType enum value | Subject semantics | Extended payload fields |
|---------------|---------------------|--------------------|--------------------------|
| PostToolUse (existing) | `agent_invocation` | `toolName` from stdin | none (baseline 4 fields only) |
| Stop (new) | `session_stop` | empty OR SPEC ID from cwd | `session_id`, `last_assistant_message_hash`, `last_assistant_message_len` |
| SubagentStop (new) | `subagent_stop` | `agentName` from stdin | `agent_name`, `agent_type`, `agent_id`, `parent_session_id` |
| UserPromptSubmit (new) | `user_prompt` | empty OR SPEC ID regex match in prompt | (default Strategy A) `prompt_hash`, `prompt_len`, `prompt_lang`; (opt-in B) `prompt_preview`; (opt-in C) `prompt_content` |

---

## 3. Wave Decomposition (Run-Phase Roadmap)

This section describes the run-phase Wave structure. The plan-phase PR (this manager-spec session) ships only the five SPEC artifacts; the Wave implementations are downstream `/moai run` work.

### 3.1 Wave A — Schema Extension + Cobra Subcommand Scaffolding (P0)

**Goal**: Extend the `Event` struct and `EventType` enum in `internal/harness/types.go`. Add three new cobra subcommands to `internal/cli/hook.go`. Reuse `isHarnessLearningEnabled` across all four observer handlers.

**Owner**: `manager-develop` (delegating to `expert-backend` for Go struct/enum extension)

**Inputs**:
- REQ-HRN-OBS-001, REQ-HRN-OBS-002, REQ-HRN-OBS-003 (cobra subcommand registration)
- REQ-HRN-OBS-004, REQ-HRN-OBS-005, REQ-HRN-OBS-006 (per-event observation entries)
- REQ-HRN-OBS-009 (schema additivity)
- REQ-HRN-OBS-012, REQ-HRN-OBS-013, REQ-HRN-OBS-014 (PII handling for UserPromptSubmit)
- REQ-HRN-OBS-015 (EventType enum extension)
- Existing `internal/harness/types.go` Event struct definition
- Existing `internal/cli/hook.go` cobra subcommand registration block

**Tasks** (T-A1 through T-A5 in tasks.md):
- T-A1: Extend `EventType` enum with three new constants (`session_stop`, `subagent_stop`, `user_prompt`).
- T-A2: Extend `Event` struct with optional fields, all tagged `json:"...,omitempty"`.
- T-A3: Implement `runHarnessObserveStop` (mirrors `runHarnessObserve` pattern).
- T-A4: Implement `runHarnessObserveSubagentStop`.
- T-A5: Implement `runHarnessObserveUserPromptSubmit` including the PII strategy switch driven by a NEW helper `resolveUserPromptStrategy(projectRoot string) Strategy`.

**Acceptance** (Wave A complete when):
- `EventType` constants present and unit-tested.
- 4 cobra subcommands appear on `hookCmd.Commands()` slice.
- All 4 observer handlers call `isHarnessLearningEnabled(cwd)` before any I/O.
- Wave A tests in `internal/cli/hook_harness_observe_multi_test.go` pass.

**Risk**:
- Enum/struct mutation could break existing consumers. Mitigation: additive-only; all new fields `omitempty`; new enum values cannot break parsers that handle unknown values as default.

### 3.2 Wave B — Wrapper Script Templates + settings.json.tmpl Registration (P0)

**Goal**: Author three new wrapper script templates under `internal/template/templates/.claude/hooks/moai/`. Run `make build` to regenerate `embedded.go`. Add three additive entries to `settings.json.tmpl`.

**Owner**: `manager-develop` (delegating to `expert-devops` for template-first compliance verification)

**Inputs**:
- REQ-HRN-OBS-016 (wrapper script authoring contract)
- REQ-HRN-OBS-017 (settings.json.tmpl additive registration)
- Existing `handle-harness-observe.sh.tmpl` (the canonical 33-line wrapper to clone)
- Existing `settings.json.tmpl` Stop / SubagentStop / UserPromptSubmit slots

**Tasks** (T-B1 through T-B4 in tasks.md):
- T-B1: Create `handle-harness-observe-stop.sh.tmpl` cloning the existing pattern; only change subcommand name.
- T-B2: Create `handle-harness-observe-subagent-stop.sh.tmpl`.
- T-B3: Create `handle-harness-observe-user-prompt-submit.sh.tmpl`.
- T-B4: Edit `settings.json.tmpl` to add three additive entries (one per event slot) routing to the new wrappers, with platform-conditional path quoting matching the existing pattern.

**Acceptance** (Wave B complete when):
- 3 new `.sh.tmpl` files exist.
- `make build` succeeds; `embedded.go` is regenerated.
- `settings.json.tmpl` has 3 additive entries; existing entries are byte-identical to before.
- A rendered settings.json (from a test fixture) contains 7 hook entries under the three slots (3 existing + 3 new + the SubagentStart entry; 4 if including SubagentStart-Subagent-stop disambiguation).
- A static check confirms no `handle-stop.sh` / `handle-subagent-stop.sh` / `handle-user-prompt-submit.sh` line was modified by Wave B.

**Risk**:
- Concurrent SPEC editing the same template lines. Mitigation: Wave B additive-only changes minimize merge conflicts; if a conflict arises, the rebase resolution is trivial (append to event slot's hook array).
- `make build` fails on a developer machine without Go bin in PATH. Mitigation: CI runs `make build` as part of pre-merge gate.

### 3.3 Wave C — Test Suite + 5-Layer Safety Integration Verification (P1)

**Goal**: Add table-driven tests for all three new event handlers. Each handler gets a NoOp test (learning disabled), a Records test (learning enabled), and a PreservesExisting test (existing log entries unmodified). Add specific tests for the PII fail-open path (REQ-HRN-OBS-014) and gate uniformity (REQ-HRN-OBS-008).

**Owner**: `manager-develop` (delegating to `expert-backend` for Go test authoring)

**Inputs**:
- REQ-HRN-OBS-004, REQ-HRN-OBS-005, REQ-HRN-OBS-006 (per-event observation entries to test)
- REQ-HRN-OBS-007 (5-Layer Safety preservation — architecture-level assertion)
- REQ-HRN-OBS-008 (gate uniformity across all 4 handlers)
- REQ-HRN-OBS-009 (schema additivity)
- REQ-HRN-OBS-014 (PII fail-open invariant)
- REQ-HRN-OBS-018 (latency budget)
- Existing `internal/cli/hook_harness_observe_test.go` (pattern source)

**Tasks** (T-C1 through T-C5 in tasks.md):
- T-C1: Author `internal/cli/hook_harness_observe_stop_test.go` with 3 test functions following the V3R4-001 pattern.
- T-C2: Author `internal/cli/hook_harness_observe_subagent_stop_test.go` with 3 test functions.
- T-C3: Author `internal/cli/hook_harness_observe_user_prompt_submit_test.go` with 3 test functions PLUS a PII strategy switch table-driven test covering all four `learning.user_prompt_content` values (hash / preview / full / none) PLUS a fail-open invalid-value test.
- T-C4: Author a shared gate-uniformity test `internal/cli/hook_harness_gate_uniformity_test.go` asserting that all 4 observer handlers no-op when `learning.enabled: false`.
- T-C5: Author a 5-Layer Safety preservation architectural test `internal/harness/safety_preservation_test.go` (read-only assertion that the constitution §5 layer count is 5 and the design.yaml strict_mode default is unchanged).

**Acceptance** (Wave C complete when):
- 9 new test functions (Stop:3 + SubagentStop:3 + UserPromptSubmit:3) pass.
- PII fail-open test passes.
- Gate-uniformity test passes (4-handler iteration confirms no JSONL write when disabled).
- Architectural safety preservation test passes.
- Code coverage for `internal/cli/hook.go` does not regress; ideally improves by 5+ percentage points due to new handlers.

**Risk**:
- Test fixture path collisions if t.TempDir() is reused across parallel tests. Mitigation: each test creates its own `t.TempDir()` and calls `t.Chdir(dir)` (per CLAUDE.local.md §6 Testing Guidelines).
- macOS/Linux path-separator differences. Mitigation: use `filepath.Join` everywhere; never hardcode `/`.

---

## 4. File-Level Changes (Run-Phase Reference)

This section summarizes file-level changes the run-phase Waves will produce. The plan-phase PR (this manager-spec session) only produces the five SPEC artifacts under `.moai/specs/SPEC-V3R4-HARNESS-002/`.

| File | Operation | Wave | Notes |
|------|-----------|------|-------|
| `.moai/specs/SPEC-V3R4-HARNESS-002/research.md` | Created | Plan | Research artifact (this plan-phase deliverable). |
| `.moai/specs/SPEC-V3R4-HARNESS-002/spec.md` | Modified (seed expansion) | Plan | Expanded from 9KB seed to full EARS-format enumeration. |
| `.moai/specs/SPEC-V3R4-HARNESS-002/plan.md` | Created | Plan | This file. |
| `.moai/specs/SPEC-V3R4-HARNESS-002/acceptance.md` | Created | Plan | AC definitions. |
| `.moai/specs/SPEC-V3R4-HARNESS-002/tasks.md` | Created | Plan | T-NNN task breakdown. |
| `internal/harness/types.go` | Modified (extend Event struct + EventType enum) | Wave A | New constants; new optional fields with `omitempty`. |
| `internal/cli/hook.go` | Modified (3 new cobra subcommand registrations + 3 new handler functions + 1 PII strategy resolver) | Wave A | Adds approximately 200 LOC; existing `runHarnessObserve` unchanged. |
| `internal/template/templates/.claude/hooks/moai/handle-harness-observe-stop.sh.tmpl` | Created | Wave B | Clone of existing harness-observe wrapper. |
| `internal/template/templates/.claude/hooks/moai/handle-harness-observe-subagent-stop.sh.tmpl` | Created | Wave B | Clone. |
| `internal/template/templates/.claude/hooks/moai/handle-harness-observe-user-prompt-submit.sh.tmpl` | Created | Wave B | Clone. |
| `internal/template/templates/.claude/settings.json.tmpl` | Modified (3 additive entries) | Wave B | Stop / SubagentStop / UserPromptSubmit slots get one new entry each. |
| `internal/template/embedded.go` | Regenerated | Wave B | Output of `make build`; tracked in git. |
| `internal/cli/hook_harness_observe_stop_test.go` | Created | Wave C | 3 test functions (NoOp / Records / PreservesExisting). |
| `internal/cli/hook_harness_observe_subagent_stop_test.go` | Created | Wave C | 3 test functions. |
| `internal/cli/hook_harness_observe_user_prompt_submit_test.go` | Created | Wave C | 3 + PII strategy + fail-open test functions. |
| `internal/cli/hook_harness_gate_uniformity_test.go` | Created | Wave C | Shared gate-uniformity assertion. |
| `internal/harness/safety_preservation_test.go` | Created | Wave C | Architectural assertion test. |

**Files explicitly NOT modified in this SPEC**:
- `.claude/rules/moai/design/constitution.md` (FROZEN)
- `internal/cli/hook_harness_observe_test.go` (V3R4-001 baseline; left untouched)
- `.claude/hooks/moai/handle-stop.sh`, `handle-subagent-stop.sh`, `handle-user-prompt-submit.sh` (existing wrappers — additive strategy means we don't modify them)
- `internal/harness/observer.go` (the `Observer.RecordEvent` and `Event` struct serialization remain unchanged; we only extend the struct fields)

---

## 5. Test Strategy

### 5.1 Unit Tests (Wave C)

- 9 new table-driven test functions covering NoOp / Records / PreservesExisting for each of the three new events.
- 1 PII strategy switch test covering all four `learning.user_prompt_content` values plus the fail-open invalid-value case.
- 1 gate-uniformity test asserting all 4 handlers no-op when learning disabled.
- 1 architectural test asserting design constitution §5 layer count is preserved.

### 5.2 Integration Tests

- Manual: in a Claude Code session, set `learning.enabled: true` and observe that a turn produces:
  - 1 entry from PostToolUse (per write/edit) with `event_type: agent_invocation`.
  - 1 entry from Stop at turn end with `event_type: session_stop`.
  - 1 entry from each SubagentStop (per Agent() invocation) with `event_type: subagent_stop`.
  - 1 entry from UserPromptSubmit at prompt input with `event_type: user_prompt` and the Strategy A fields.
- Schema additivity verification: a pre-existing `.moai/harness/usage-log.jsonl` from V3R4-001 (containing only baseline 4-field entries) MUST remain parseable; the JSONL reader does not break on missing optional fields.

### 5.3 Static Verification

- `grep -nE 'AskUserQuestion' internal/cli/hook.go` shows zero matches in the new handler functions.
- `grep -nE 'prompt_content|prompt_preview' internal/harness/types.go` shows the new optional fields are tagged `omitempty`.
- `git diff main..HEAD -- internal/template/templates/.claude/settings.json.tmpl | grep -E '^[-]' | grep -v '^[-]{3}'` shows zero LINES REMOVED from the template (additive-only invariant for Wave B).

### 5.4 Plan-Auditor Run

- `Agent(subagent_type: "plan-auditor")` invoked at Phase 2.5 of this `/moai plan` session with the five SPEC artifacts as input.
- Pass threshold: 0.80 minimum. Below that, findings are addressed in iterative drafts (max 3 iterations).
- If iteration 3 still fails, manager-spec escalates findings to the orchestrator as a blocker report.

---

## 6. MX Tag Plan (run-phase guidance)

The plan-phase PR (this session) does NOT add MX tags to source code. The following MX tag plan is FORWARD-LOOKING guidance for the run-phase Waves:

| File | Function | MX tag type | Rationale |
|------|----------|-------------|-----------|
| `internal/cli/hook.go` | `runHarnessObserveStop` | `@MX:NOTE` | Cite REQ-HRN-OBS-004 + REQ-HRN-OBS-008. Document that the gate is fail-open per V3R4-001 semantics. |
| `internal/cli/hook.go` | `runHarnessObserveSubagentStop` | `@MX:NOTE` | Cite REQ-HRN-OBS-005 + REQ-HRN-OBS-008. |
| `internal/cli/hook.go` | `runHarnessObserveUserPromptSubmit` | `@MX:WARN` + `@MX:REASON` | High-risk PII handling; the WARN requires explicit @MX:REASON documenting Strategy A default and the fail-open to Strategy A on invalid config. |
| `internal/cli/hook.go` | `resolveUserPromptStrategy` | `@MX:ANCHOR` + `@MX:REASON` | If `fan_in` reaches 3+ (called from at least 3 sites: the UserPromptSubmit handler, the PII test fixture, the gate-uniformity test). REASON cites the fail-open invariant. |
| `internal/harness/types.go` | `EventType` constants block | `@MX:NOTE` | Cite REQ-HRN-OBS-015 enum extension. Note that downstream classifiers (SPEC-V3R4-HARNESS-003) consume these values. |
| `internal/harness/types.go` | `Event` struct (new fields) | `@MX:NOTE` | Cite REQ-HRN-OBS-009 schema additivity. Note that all new fields are `omitempty`. |
| `internal/cli/hook.go` | New cobra subcommand init block | `@MX:NOTE` | Cite REQ-HRN-OBS-001 / 002 / 003 wiring. |

The run-phase agent (`manager-develop`) will execute these tag additions in the GREEN phase. The plan-phase here only DECLARES the targets.

---

## 7. Risks (Top 5)

| Risk | Likelihood | Severity | Mitigation |
|------|------------|----------|------------|
| PII leakage through misconfigured `learning.user_prompt_content` | Low | High | REQ-HRN-OBS-014 mandates fail-open to Strategy A. Test T-C3 explicitly verifies this with the `learning.user_prompt_content: garbage_value` case. |
| Performance regression from extra JSONL writes per turn | Medium | Low | REQ-HRN-OBS-018 enforces 100ms latency budget. Manual measurement in the run-phase Wave C delivers the evidence. |
| Settings.json.tmpl merge conflict if a concurrent SPEC adds an entry to the same event slot | Medium | Medium | Additive strategy (WIRE-A) keeps the resolution trivial: append to the JSON array. The diff is local to each event slot. |
| EventType enum extension breaks unknown downstream consumer | Low | Medium | The enum is a Go string type; adding values cannot break parsers. Downstream classifier code is NOT modified by this SPEC. Validated by REQ-HRN-OBS-010 preservation. |
| Tests over-rely on `t.TempDir()` + `t.Chdir()` and become flaky in parallel | Low | Low | Each test creates its own TempDir; t.Chdir within t.Run subtest is process-local on Go 1.23+. CLAUDE.local.md §6 documents the pattern. |

---

## 8. Dependencies

### 8.1 Inbound Dependencies (this SPEC depends on)

- `SPEC-V3R4-HARNESS-001` (foundation): REQ-HRN-FND-005 (5-Layer Safety), REQ-HRN-FND-009 (gate), REQ-HRN-FND-010 (baseline schema), REQ-HRN-FND-011 (4-tier ladder), REQ-HRN-FND-015 (subagent prohibition).
- `.claude/rules/moai/core/hooks-system.md` § Hook Event stdin/stdout Reference: Claude Code event payload contracts.
- `.claude/rules/moai/design/constitution.md` §2, §5: FROZEN zone and 5-Layer Safety.
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline + § Hierarchical Acceptance Criteria Schema.

### 8.2 Outbound Dependencies (SPECs that depend on this)

- `SPEC-V3R4-HARNESS-003` (embedding-cluster classifier): consumes the multi-event observation stream produced by this SPEC. The new EventType enum values are the input vocabulary.
- `SPEC-V3R4-HARNESS-004` through `SPEC-V3R4-HARNESS-008`: indirectly blocked; they build on the architecture established here.

### 8.3 Co-temporal Dependencies

None. V3R4-002 is independent of other in-flight Sprint 13 SPECs (HRN-001 re-evaluation, MIG-001).

---

## 9. Out of Scope (Explicit List)

The following items are explicitly NOT in any Wave of this SPEC:

1. Embedding-cluster classifier → `SPEC-V3R4-HARNESS-003`.
2. Reflexion self-critique → `SPEC-V3R4-HARNESS-004`.
3. Principle-based scoring → `SPEC-V3R4-HARNESS-005`.
4. Multi-objective scoring → `SPEC-V3R4-HARNESS-006`.
5. Skill library auto-organization → `SPEC-V3R4-HARNESS-007`.
6. Cross-project federation → `SPEC-V3R4-HARNESS-008`.
7. Retroactive PostToolUse wrapper registration in `settings.json.tmpl` (V3R4-001 follow-up gap).
8. Migration of pre-V3R4-002 `usage-log.jsonl` entries to a new schema.
9. GUI / dashboard for observation log inspection.
10. Networking, telemetry upload, or external API integration.

---

## 10. Execution Order (Plan-Phase)

This `/moai plan SPEC-V3R4-HARNESS-002` session executes the following phases:

| Phase | Activity | Status |
|-------|----------|--------|
| Phase 0 | Deep research (read V3R4-001 artifacts, internal/cli/hook.go, internal/harness/types.go, settings.json.tmpl, existing wrapper scripts) | Complete |
| Phase 1 | Branch setup: `feature/SPEC-V3R4-HARNESS-002` in worktree at `~/.moai/worktrees/MoAI-ADK/SPEC-V3R4-HARNESS-002` | Complete (worktree already exists) |
| Phase 2 | Draft `research.md` (Phase 0.5 deliverable) | Complete |
| Phase 2.5 | Draft `spec.md` with frontmatter canonical-9-field reconciliation + 18 EARS-format REQs | Complete |
| Phase 2.6 | Draft `plan.md` (this file) with Wave A/B/C decomposition | Complete |
| Phase 2.7 | Draft `acceptance.md` with 13 ACs (hierarchical AC schema allowed per spec-workflow.md) | Complete |
| Phase 2.8 | Draft `tasks.md` with T-A1..T-A5, T-B1..T-B4, T-C1..T-C5 | Complete |
| Phase 2.9 | Self-audit checklist (9-field frontmatter, REQ↔AC pairing, Non-Goals verbatim, V3R4-001 contract preservation, PII REQ+AC present) | Complete |
| Phase 3 | Return structured summary to orchestrator | Pending |
| Phase 4 | Orchestrator invokes `plan-auditor` (iter 1) | Pending |
| Phase 5 | Orchestrator commits + delegates PR creation to `manager-git` | Pending |

The run-phase Waves A-C happen in `/moai run SPEC-V3R4-HARNESS-002` after the plan PR merges into main.

---

End of plan.md.
