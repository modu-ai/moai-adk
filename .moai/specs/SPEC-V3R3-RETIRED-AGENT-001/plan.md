# SPEC-V3R3-RETIRED-AGENT-001 Implementation Plan (Phase 1B)

> Implementation plan for retired-stub compatibility fix + manager-cycle template alignment.
> Companion to `spec.md` v0.1.0 and `research.md` v0.1.0.
> Authored against branch `feature/SPEC-V3R3-RETIRED-AGENT-001` at `/Users/goos/MoAI/moai-adk-go` (solo mode, no worktree).

## HISTORY

| Version | Date       | Author                        | Description                                                                                                       |
|---------|------------|-------------------------------|-------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow (Phase 1B) | 최초 작성 — P0 (manager-cycle add + retired stub standardize + SubagentStart guard) + P1 (worktreePath validation + text/template) + P2 (lessons #11) 5-milestone plan + REQ↔AC matrix |

---

## 1. Plan Overview

### 1.1 Goal restatement

본 plan은 spec.md REQ-RA-001..015를 실행 가능한 5-milestone 작업 분해로 변환한다. 핵심 deliverable:

- **신규 active agent file**: `internal/template/templates/.claude/agents/moai/manager-cycle.md` — mo.ai.kr 10245-byte version reference + moai-adk-go quality 검증 후 import.
- **신규 hook handler**: `internal/hook/agent_start.go` — SubagentStart event dispatch + retired-rejection guard + factory integration.
- **신규 audit/regression test**: `internal/template/agent_frontmatter_audit_test.go`, `internal/template/manager_cycle_present_test.go`, `internal/hook/agent_start_test.go`.
- **표준화**: `manager-tdd.md` retired stub frontmatter (`retired: true`, replacement, hint, empty arrays).
- **Hook wrapper update**: `handle-subagent-start.sh.tmpl` exit code propagation.
- **Wrapper validation guard**: `internal/cli/launcher.go` (또는 신규 `agent_wrapper.go`) `worktreePath` validation + sentinel error.
- **Path interpolation refactor**: 3-5 callsites string concat → `text/template`.
- **Documentation substitution**: 7 references across 6 files.
- **Lessons**: lessons.md #11 entry (5-layer defect chain anti-pattern).

### 1.2 Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd` (assumed; verify at run time). Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` Phase Run.

- **RED (M1)**: 4 신규 audit/unit tests를 먼저 작성. 모두 실패 상태 확인 (`agent_frontmatter_audit_test.go`, `manager_cycle_present_test.go`, `agent_start_test.go`, regression test for empty-worktreePath validation).
- **GREEN part 1 (M2)**: `manager-cycle.md` template 추가 (mo.ai.kr 10245-byte reference 검증 후 import) + `manager-tdd.md` retired stub frontmatter standardization. agent_frontmatter_audit_test + manager_cycle_present_test → GREEN.
- **GREEN part 2 (M3)**: `internal/hook/agent_start.go` 신규 handler + `internal/hook/agents/factory.go` dispatch 확장 + `handle-subagent-start.sh.tmpl` exit-code propagation. agent_start_test → GREEN.
- **GREEN part 3 (M4)**: `internal/cli/launcher.go` (또는 wrapper) `validateWorktreeReturn` 함수 + 3-5 callsite refactor (string concat → text/template). worktreePath validation regression test → GREEN.
- **REFACTOR (M5)**: 7 documentation reference substitutions + lessons.md #11 entry + final `make build` + full `go test ./...` + `golangci-lint run`.

### 1.3 Deliverables

| Deliverable | Path | REQ Coverage |
|---|---|---|
| Unified DDD/TDD agent definition | `internal/template/templates/.claude/agents/moai/manager-cycle.md` (NEW) | REQ-RA-001, REQ-RA-013 |
| Retired stub standardization | `internal/template/templates/.claude/agents/moai/manager-tdd.md` (MODIFY) | REQ-RA-002, REQ-RA-011 |
| SubagentStart hook handler | `internal/hook/agent_start.go` (NEW) | REQ-RA-004, REQ-RA-007, REQ-RA-008 |
| Hook factory dispatch extension | `internal/hook/agents/factory.go` (MODIFY) | REQ-RA-009 |
| Hook wrapper exit-code propagation | `internal/template/templates/.claude/hooks/moai/handle-subagent-start.sh.tmpl` (MODIFY) | REQ-RA-007 |
| Worktree validation wrapper | `internal/cli/launcher.go` (MODIFY) or `internal/cli/agent_wrapper.go` (NEW) | REQ-RA-005, REQ-RA-010 |
| Path interpolation refactor (≤5 callsites) | `internal/cli/launcher.go` + related callsite files | REQ-RA-006 |
| Frontmatter audit test | `internal/template/agent_frontmatter_audit_test.go` (NEW) | REQ-RA-002, REQ-RA-013, REQ-RA-016 |
| manager-cycle presence test | `internal/template/manager_cycle_present_test.go` (NEW) | REQ-RA-003 |
| SubagentStart hook unit test | `internal/hook/agent_start_test.go` (NEW) | REQ-RA-004, REQ-RA-007, REQ-RA-008, REQ-RA-009, REQ-RA-012 |
| Worktree validation regression test | `internal/cli/launcher_worktree_validation_test.go` (NEW) | REQ-RA-005, REQ-RA-010 |
| CLAUDE.md §4 + §5 substitution | `internal/template/templates/CLAUDE.md` (MODIFY) | REQ-RA-013 |
| agent-authoring.md substitution | `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` (MODIFY) | REQ-RA-013 |
| agent-hooks.md substitution | `internal/template/templates/.claude/rules/moai/core/agent-hooks.md` (MODIFY) | REQ-RA-013 |
| spec-workflow.md substitution | `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` (MODIFY) | REQ-RA-013 |
| manager-strategy.md substitution | `internal/template/templates/.claude/agents/moai/manager-strategy.md` (MODIFY) | REQ-RA-013 |
| manager-ddd.md inline reference substitution | `internal/template/templates/.claude/agents/moai/manager-ddd.md` (MODIFY, 2 lines) | REQ-RA-013 |
| `moai agents list --retired` (optional) | `internal/cli/agents.go` (NEW or MODIFY) | REQ-RA-014 |
| lessons.md entry | `~/.claude/projects/{hash}/memory/lessons.md` (auto-memory APPEND) | REQ-RA-016 (Trackable) |

[HARD] Embedded-template parity: 모든 `.claude/...` 변경은 `internal/template/templates/.claude/...` mirror + `make build` 필수 (CLAUDE.local.md §2 Template-First HARD).

### 1.4 Traceability Matrix (REQ → AC mapping)

Plan-auditor PASS criterion #2: every REQ maps to at least one AC.

| REQ ID | Category | Mapped AC(s) |
|---|---|---|
| REQ-RA-001 | Ubiquitous | AC-RA-01, AC-RA-16 |
| REQ-RA-002 | Ubiquitous | AC-RA-02, AC-RA-17 |
| REQ-RA-003 | Ubiquitous | AC-RA-03 |
| REQ-RA-004 | Ubiquitous | AC-RA-04, AC-RA-07 |
| REQ-RA-005 | Ubiquitous | AC-RA-05, AC-RA-10, AC-RA-18 |
| REQ-RA-006 | Ubiquitous | AC-RA-06, AC-RA-18 |
| REQ-RA-007 | Event-Driven | AC-RA-04, AC-RA-07, AC-RA-17 |
| REQ-RA-008 | Event-Driven | AC-RA-08 |
| REQ-RA-009 | Event-Driven | AC-RA-09 |
| REQ-RA-010 | Event-Driven | AC-RA-10, AC-RA-18 |
| REQ-RA-011 | State-Driven | AC-RA-11 |
| REQ-RA-012 | State-Driven | AC-RA-12 |
| REQ-RA-013 | State-Driven | AC-RA-13, AC-RA-16 |
| REQ-RA-014 | Optional | AC-RA-14 |
| REQ-RA-015 | Unwanted | AC-RA-10, AC-RA-18 |
| REQ-RA-016 | Unwanted (Composite) | AC-RA-15 |

Coverage: **16/16 REQs mapped, 18/18 ACs validated** (some ACs cover multiple REQs; see acceptance.md §1 for full Given/When/Then).

---

## 2. Milestone Breakdown (M1-M5)

각 milestone은 **priority-ordered** (no time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation HARD rule).

### M1: Test scaffolding (RED phase) — Priority P0

Reference: `internal/template/lang_boundary_audit_test.go` (SPEC-V3R2-WF-005 M1 패턴), `internal/hook/agents/base_handler_test.go` (existing hook test pattern), `internal/cli/glm_test.go` (existing CLI test pattern).

Owner role: `expert-backend` (Go test) or direct `manager-cycle` execution (with `cycle_type=tdd`).

Scope:

1. **`internal/template/agent_frontmatter_audit_test.go`** (REQ-RA-002, REQ-RA-013, REQ-RA-016):
   - `TestAgentFrontmatterAudit`: walk `internal/template/templates/.claude/agents/moai/*.md` in embedded FS; for each file with `retired: true`, assert all required retirement fields present (`retired_replacement`, `retired_param_hint`, `tools: []`, `skills: []`); for non-retired agents, assert no `retired:` field.
   - `TestRetirementCompletenessAssertion`: for each `retired: true` agent, verify the corresponding active replacement file exists (REQ-RA-016's RETIREMENT_INCOMPLETE_<agent> CI rule). E.g., manager-tdd retired → manager-cycle.md must exist.

2. **`internal/template/manager_cycle_present_test.go`** (REQ-RA-003):
   - `TestManagerCyclePresentInEmbeddedFS`: open `internal/template/embedded.go` derived FS, assert `_, err := fs.Stat(".claude/agents/moai/manager-cycle.md"); err == nil`. Body length ≥ 5000 bytes (sanity check).

3. **`internal/hook/agent_start_test.go`** (REQ-RA-004, REQ-RA-007, REQ-RA-008, REQ-RA-009, REQ-RA-012):
   - `TestAgentStartHandlerRoutesViaFactory`: invoke `factory.NewHandler("agent-start")` (or equivalent) — assert returns non-nil handler of type `*AgentStartHandler`.
   - `TestAgentStartBlocksRetiredAgent`: stub HookInput with `agentName: "manager-tdd"` + temp test agent file with `retired: true` frontmatter; invoke handler.Handle(); assert output contains decision=block + reason mentions retired_replacement.
   - `TestAgentStartAllowsActiveAgent`: stub HookInput with `agentName: "manager-cycle"` + frontmatter without `retired`; invoke handler.Handle(); assert output is allow (no decision field or decision=allow).
   - `TestAgentStartAllowsUnknownAgent`: stub HookInput with `agentName: "non-existent"`; invoke handler.Handle(); assert output is allow (REQ-RA-008: unknown agent bypass).
   - `TestAgentStartHandlerPerformance`: invoke handler 100x with retired-stub frontmatter; assert avg latency < 500ms (REQ-RA-012). Use `t.Skip` if CI runner is too slow.

4. **`internal/cli/launcher_worktree_validation_test.go`** (REQ-RA-005, REQ-RA-010):
   - `TestValidateWorktreeReturnRejectsEmptyObject`: stub Agent.Return value with `worktreePath: ""` (or empty struct); invoke validation; assert returns error with sentinel `WORKTREE_PATH_INVALID`.
   - `TestValidateWorktreeReturnRejectsNullPath`: stub with nil pointer; assert error with WORKTREE_PATH_INVALID.
   - `TestValidateWorktreeReturnAcceptsValidPath`: stub with `worktreePath: "/tmp/abc123"`; assert no error.
   - `TestValidateWorktreeReturnSkipsWhenIsolationNotWorktree`: stub call without `isolation: "worktree"`; assert no error even with empty worktreePath (validation only applies when worktree was requested).

Exit criteria for M1:
- All 4 test files compile (`go build ./internal/template/... ./internal/hook/... ./internal/cli/...`)
- All listed test functions exist and are runnable
- All tests fail with predictable error messages (no implementation yet)
- `go test ./internal/template/ ./internal/hook/ ./internal/cli/ -run "TestAgentFrontmatterAudit|TestManagerCyclePresent|TestAgentStart|TestValidateWorktree" 2>&1 | grep -c FAIL` ≥ 4

### M2: manager-cycle template + retired stub standardization (GREEN part 1) — Priority P0

Reference: mo.ai.kr `manager-cycle.md` (10245 bytes) + research.md §6.1 quality checklist.

Owner role: `expert-backend` (template authoring) or direct `manager-cycle` execution.

Scope:

1. **Create `internal/template/templates/.claude/agents/moai/manager-cycle.md`** (REQ-RA-001):
   - Reference: mo.ai.kr 10245-byte version (read with Read tool from `/Users/goos/MoAI/mo.ai.kr/.claude/agents/moai/manager-cycle.md`)
   - Quality check per research.md §6.1:
     - 16-language neutrality
     - anti-bias (no language preference)
     - frontmatter parity: `model: sonnet`, `permissionMode: bypassPermissions`, `memory: project`
     - skills: `moai-foundation-core`, `moai-workflow-ddd`, `moai-workflow-tdd`, `moai-workflow-testing`
     - hooks: PreToolUse cycle-pre-implementation, PostToolUse cycle-post-implementation, SubagentStop cycle-completion
   - Adjust hook action names from `tdd-*` (legacy if present) to `cycle-*` (unified)
   - Verify `document: cycle_type` field is supported (or remove if causes warnings)

2. **Create corresponding hook handler stubs in `internal/hook/agents/cycle_handler.go`** (NEW; if cycle-* actions don't yet have handlers):
   - `cycle-pre-implementation`: PreToolUse stub (no logic; pass-through)
   - `cycle-post-implementation`: PostToolUse stub
   - `cycle-completion`: SubagentStop stub
   - Update `internal/hook/agents/factory.go` switch to dispatch these actions.
   - **Note**: This is incidentally part of M2 because the manager-cycle.md frontmatter references hook actions that must dispatch correctly. If hook handlers are missing, hook chain fails silently (current behavior of unknown actions is `default_handler.go` fallthrough — acceptable but non-functional).

3. **Modify `internal/template/templates/.claude/agents/moai/manager-tdd.md`** to retired stub (REQ-RA-002, REQ-RA-011):
   - Replace existing 6407-byte content with retired stub per research.md §6.2 frontmatter
   - Body content: retain the migration notes from existing mo.ai.kr stub (Old → New invocation, What Changed, Documentation references)
   - Total file size: target 1000-2000 bytes (smaller than current 6407, larger than 976 due to expanded migration notes)

4. **Run tests**:
   - `go test ./internal/template/ -run "TestAgentFrontmatterAudit|TestManagerCyclePresent" -v` → expect both PASS

Exit criteria for M2:
- Both M1 audit/presence tests now PASS
- `make build` succeeds (regenerates `internal/template/embedded.go`)
- `go vet ./internal/template/...` clean
- `golangci-lint run ./internal/template/...` clean

### M3: SubagentStart hook handler (GREEN part 2) — Priority P0

Reference: `internal/hook/agents/factory.go` existing dispatch, `internal/hook/agents/base_handler.go`.

Owner role: `expert-backend` (Go) or direct `manager-cycle` (with `cycle_type=tdd`) execution.

Scope:

1. **Create `internal/hook/agent_start.go`** (REQ-RA-004, REQ-RA-007, REQ-RA-008, REQ-RA-012):
   - struct `AgentStartHandler` extending `baseHandler` (action: "agent-start", event: hook.EventSubagentStart)
   - `NewAgentStartHandler() hook.Handler`
   - `Handle(ctx, input)` logic:
     a. Extract `agentName` from input.Data
     b. Locate agent file: try `.claude/agents/moai/<name>.md`, then `.claude/agents/<name>.md`
     c. If not found, return `hook.NewAllowOutput()` (REQ-RA-008)
     d. Read file, parse YAML frontmatter (use existing YAML parser from internal/config or yaml.v3)
     e. If frontmatter has `retired: true`:
        - Construct reason: `"agent <name> retired (SPEC-V3R2-ORC-001), use <retired_replacement> with <retired_param_hint>"`
        - Return `hook.NewBlockOutput(reason)` with exit code 2
     f. Else: return `hook.NewAllowOutput()`
   - Performance budget: ≤500ms (single YAML parse, no network)

2. **Modify `internal/hook/agents/factory.go`** (REQ-RA-009):
   - Add new case branch: `case "agent-start": return NewAgentStartHandler(), nil`
   - Note: Place new case logically (e.g., before default) — verify with existing factory dispatch ordering convention

3. **Modify `internal/template/templates/.claude/hooks/moai/handle-subagent-start.sh.tmpl`** (REQ-RA-007):
   - Change `exec moai hook subagent-start` to non-exec form so exit code can be captured and propagated:
     ```bash
     moai hook subagent-start < "$temp_file" 2>/dev/null
     exit $?
     ```
   - Maintain timeout: 5s
   - Test on macOS + Linux (POSIX bash)

4. **Run tests**:
   - `go test ./internal/hook/ -run "TestAgentStart" -v` → expect 5 tests PASS
   - `go vet ./internal/hook/...` clean
   - `golangci-lint run ./internal/hook/...` clean

Exit criteria for M3:
- All M1 SubagentStart tests PASS
- handle-subagent-start.sh.tmpl exit code propagation verified manually (echo test)
- `make build` regenerates embedded FS

### M4: worktreePath validation + path interpolation refactor (GREEN part 3) — Priority P1

Reference: `internal/cli/launcher.go` existing dispatch, Go `text/template` package.

Owner role: `expert-backend` (Go) or direct `manager-cycle` execution.

Scope:

1. **Identify Agent() callsites consuming worktreePath / worktreeBranch**:
   - Run `grep -rn 'worktreePath\|worktreeBranch' internal/ 2>/dev/null` to enumerate
   - Estimated 3-5 callsites; if >5, escalate to user via blocker report (per research.md §5)

2. **Add `validateWorktreeReturn(result, isolationMode) error` helper**:
   - Location: `internal/cli/launcher.go` (or new `internal/cli/agent_wrapper.go`)
   - Logic:
     a. If `isolationMode != "worktree"`, return nil (skip validation)
     b. If `result.WorktreePath` is empty string / nil / map without "path" key: return `errors.New("WORKTREE_PATH_INVALID: agent return value missing valid worktreePath")`
     c. If `result.WorktreeBranch` is empty / nil: include in error message but don't fail solely on branch (worktreeBranch may be optional)
     d. Else: return nil

3. **Refactor path interpolation callsites** (REQ-RA-006):
   - For each callsite found in Step 1, replace `fmt.Sprintf("%s/%s/%s", root, branch, path)` patterns with `text/template` parsed once
   - Type-safe data struct: `struct { Root, Branch, Path string }` ensures non-string values produce compile-time or runtime error
   - Example pattern:
     ```go
     pathTemplate, err := template.New("worktreePath").Parse("{{.Root}}/{{.Branch}}/{{.Path}}")
     // ... execute with typed struct
     ```

4. **Run tests**:
   - `go test ./internal/cli/ -run "TestValidateWorktree" -v` → 4 tests PASS
   - `go test ./internal/cli/ -run "TestUnifiedLaunch" -v` → existing tests still PASS (no regression)
   - `go vet ./internal/cli/...` clean
   - `golangci-lint run ./internal/cli/...` clean

Exit criteria for M4:
- All M1 worktree validation tests PASS
- Path interpolation callsites use text/template (or equivalent type-safe approach)
- No new errors / warnings in `go test ./...` full suite
- `make build` clean

### M5: Documentation substitution + lessons + final validation (REFACTOR phase) — Priority P2

Reference: research.md §3.6 substitution scope (7 references across 6 files).

Owner role: `manager-docs` for documentation substitutions, `expert-backend` for lessons.md entry.

Scope:

1. **Documentation substitutions** (REQ-RA-013, AC-RA-13):
   - `internal/template/templates/CLAUDE.md`:
     - §4 Manager Agents: list update (`spec, ddd, tdd, ...` → `spec, cycle, docs, ...` if active list; or keep "Manager Agents (8)" as effective active count + list update)
     - §5 Agent Chain for SPEC Execution: Phase 3 (`expert-backend`) sequence — verify no manager-tdd/ddd reference
     - "MoAI Command Flow": `manager-ddd or manager-tdd subagent (per quality.yaml development_mode)` → `manager-cycle subagent (with cycle_type per quality.yaml development_mode)`
   - `internal/template/templates/.claude/rules/moai/development/agent-authoring.md`:
     - § Manager Agents (8): list — replace `manager-tdd: TDD implementation cycle` and `manager-ddd: DDD implementation cycle` with `manager-cycle: Unified DDD/TDD implementation cycle`
     - Add `retired:` field documentation to "Supported Frontmatter Fields" table (Yes/No, default false, description "Marks agent as retired; SubagentStart hook blocks spawn")
   - `internal/template/templates/.claude/rules/moai/core/agent-hooks.md`:
     - Agent Hook Actions table: replace `manager-tdd | tdd-pre-implementation | tdd-post-implementation | tdd-completion` and `manager-ddd | ddd-pre-transformation | ddd-post-transformation | ddd-completion` rows with `manager-cycle | cycle-pre-implementation | cycle-post-implementation | cycle-completion`
   - `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`:
     - Phase Overview table row `Run | /moai run | manager-ddd/tdd (per quality.yaml) | 180K | DDD/TDD implementation` → `Run | /moai run | manager-cycle (per quality.yaml development_mode) | 180K | DDD/TDD implementation`
   - `internal/template/templates/.claude/agents/moai/manager-strategy.md`:
     - Line containing `manager-ddd or manager-tdd` → `manager-cycle`
   - `internal/template/templates/.claude/agents/moai/manager-ddd.md`:
     - 2 inline references to `manager-tdd` → `manager-cycle with cycle_type=tdd`

2. **Optional: `moai agents list --retired` subcommand** (REQ-RA-014):
   - Decision point: use AskUserQuestion to confirm in-scope vs deferred
   - If in-scope: implement in `internal/cli/agents.go` (NEW) — list all `.claude/agents/**/*.md` files with `retired: true` frontmatter
   - If deferred: document in spec-compact.md §Open Items + create follow-up SPEC issue in GitHub

3. **lessons.md #11 entry** (P2, Trackable):
   - Append entry to `~/.claude/projects/{hash}/memory/lessons.md`
   - Category: workflow + architecture
   - Pattern: "Retired agent stub frontmatter without standardization → 5-layer defect chain"
   - Correct approach: REQ-RA-002 standardized frontmatter + REQ-RA-007 SubagentStart guard + REQ-RA-005 worktreePath validation
   - Date: 2026-05-04
   - Tags: agent-runtime, retired-stub, hook-guard, worktree, defect-chain, mo.ai.kr-incident

4. **Final validation**:
   - `make build` — embedded FS regenerated
   - `go test ./...` — full suite PASS (including all M1-M4 new tests + existing tests)
   - `go vet ./...` — clean
   - `golangci-lint run ./...` — clean
   - `make install` — local binary updated
   - Manual smoke test: `moai update` in `/tmp/test-project` and verify `manager-cycle.md` appears + `manager-tdd.md` is retired stub

Exit criteria for M5:
- All 7 documentation substitutions verified via grep
- All M1-M4 tests PASS in final test run
- lessons.md #11 entry added (or proposed via AskUserQuestion to user)
- No regressions in existing tests
- `make build && make install` clean

---

## 3. File-Level Changes Summary

| File | Action | LOC delta (estimated) | Phase |
|---|---|---|---|
| `internal/template/templates/.claude/agents/moai/manager-cycle.md` | NEW | +250 | M2 |
| `internal/template/templates/.claude/agents/moai/manager-tdd.md` | REWRITE (6407→1500 bytes) | -120 | M2 |
| `internal/hook/agent_start.go` | NEW | +60 | M3 |
| `internal/hook/agents/cycle_handler.go` | NEW (incidental, M2) | +50 | M2 |
| `internal/hook/agents/factory.go` | MODIFY (add cases) | +10 | M2/M3 |
| `internal/template/templates/.claude/hooks/moai/handle-subagent-start.sh.tmpl` | MODIFY (exit propagation) | +2 | M3 |
| `internal/cli/launcher.go` | MODIFY (validateWorktreeReturn) | +30 | M4 |
| Path interpolation callsites refactor | MODIFY 3-5 callsites | -10 / +20 | M4 |
| `internal/template/agent_frontmatter_audit_test.go` | NEW | +80 | M1 |
| `internal/template/manager_cycle_present_test.go` | NEW | +30 | M1 |
| `internal/hook/agent_start_test.go` | NEW | +120 | M1 |
| `internal/cli/launcher_worktree_validation_test.go` | NEW | +80 | M1 |
| `internal/template/templates/CLAUDE.md` | MODIFY (§4, §5, MoAI Command Flow) | ±15 | M5 |
| `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` | MODIFY (Manager Agents list + retired field doc) | +15 | M5 |
| `internal/template/templates/.claude/rules/moai/core/agent-hooks.md` | MODIFY (Agent Hook Actions table) | +5 | M5 |
| `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` | MODIFY (Phase Overview row) | ±2 | M5 |
| `internal/template/templates/.claude/agents/moai/manager-strategy.md` | MODIFY (single line) | ±1 | M5 |
| `internal/template/templates/.claude/agents/moai/manager-ddd.md` | MODIFY (2 inline references) | ±2 | M5 |
| `internal/cli/agents.go` (--retired flag) | NEW (optional) | +40 | M5 (optional) |
| lessons.md #11 entry | APPEND auto-memory | +20 | M5 |

**Total LOC delta estimate**: +500 / -150 = +350 net.

**Files created**: 7 (5 Go test+source + 1 agent definition + 1 hook handler stub).
**Files modified**: 9 (1 retired stub rewrite + 8 reference substitutions + path callsites).
**Files deleted**: 0.

---

## 4. mx_plan (10 MX tags / 7 files)

Per `.claude/rules/moai/workflow/mx-tag-protocol.md` and SPEC-V3R3-HYBRID-001 mx_plan pattern.

| MX type | File | Anchor function/section | Reason |
|---|---|---|---|
| @MX:ANCHOR | `internal/hook/agent_start.go` | `Handle` method (REQ-RA-007) | Critical retired-rejection guard; high fan_in (every SubagentStart event) |
| @MX:ANCHOR | `internal/cli/launcher.go` | `validateWorktreeReturn` (REQ-RA-005) | Critical worktreePath validation; high fan_in (every Agent isolation:worktree return) |
| @MX:ANCHOR | `internal/hook/agents/factory.go` | switch case "agent-start" (REQ-RA-009) | Single point of dispatch for new event type |
| @MX:NOTE | `internal/template/templates/.claude/agents/moai/manager-tdd.md` | Frontmatter section | Documents retirement decision + replacement (REQ-RA-002, REQ-RA-011) |
| @MX:NOTE | `internal/template/templates/.claude/agents/moai/manager-cycle.md` | Header / cycle_type parameter section | Documents unified DDD/TDD agent (REQ-RA-001) |
| @MX:NOTE | `internal/cli/launcher.go` | Path template definition | Documents text/template adoption rationale (REQ-RA-006) |
| @MX:WARN | `internal/hook/agent_start.go` | YAML parse logic | Untrusted file content read; parse errors must not panic |
| @MX:WARN | `internal/template/templates/.claude/hooks/moai/handle-subagent-start.sh.tmpl` | exit code propagation | Shell exit code semantics differ between bash/dash; propagation is critical for spawn block |
| @MX:TODO | `internal/cli/agents.go` (if optional REQ-RA-014 deferred) | (entire file) | Resolve in follow-up SPEC; currently placeholder |
| @MX:LEGACY | `internal/template/templates/.claude/agents/moai/manager-ddd.md` | Existing agent body (no changes scope) | Out-of-scope retirement consideration; flagged for SPEC-V3R3-RETIRED-DDD-001 (가칭) |

Total: 10 MX tags / 7 files (3 ANCHOR + 3 NOTE + 2 WARN + 1 TODO + 1 LEGACY).

---

## 5. Risk Table (file-anchored mitigations)

| 리스크 | 영향 | 확률 | File anchor | 완화 |
|---|---|---|---|---|
| SubagentStart hook does not actually block spawn on exit 2 | H | M | `internal/hook/agent_start.go` | M3 empirical test in `/tmp/test-project`; fallback to PreToolUse hook on Agent tool with matcher |
| `retired: true` custom field rejected by Claude Code YAML parser | H | L | `internal/template/templates/.claude/agents/moai/manager-tdd.md` | M2 verify with single test agent spawn; fallback to `description:` field encoding (research.md §4.3) |
| Path interpolation refactor scope >5 callsites → drive-by violation | M | M | `internal/cli/launcher.go` and grep results | M4 measure first; if >5, escalate to user, scope-cut to validation-only |
| mo.ai.kr `manager-cycle.md` 10245-byte version contains language bias / non-neutral content | M | M | `internal/template/templates/.claude/agents/moai/manager-cycle.md` | M2 quality checklist (research.md §6.1); manual review before commit |
| `manager-ddd` retired stub case discovered during audit test → out-of-scope drift | L | H | `internal/template/agent_frontmatter_audit_test.go` | Test scope manager-tdd only; manager-ddd separate SPEC `SPEC-V3R3-RETIRED-DDD-001` |
| Embedded FS regeneration fails after large template change | M | L | `internal/template/embedded.go` (auto-generated) | M2 + M3 + M5 each run `make build` checkpoint; fail fast |
| Hook wrapper `exec` → non-exec change breaks Linux+macOS shell semantics | M | L | `internal/template/templates/.claude/hooks/moai/handle-subagent-start.sh.tmpl` | M3 manual test on both platforms; use POSIX bash form `exit $?` |
| `moai update` overwrites user-modified agent files in mo.ai.kr | M | M | (user-side) | CHANGELOG warns; user should commit local changes first; `.claude/` is template-managed per CLAUDE.local.md §2 |
| `text/template` migration adds noticeable LOC overhead vs `fmt.Sprintf` | L | H | path interpolation callsites | M4 LOC delta budget +20; if exceeded, evaluate simpler validation-only fix |
| Performance regression on agent spawn due to YAML parse overhead | L | L | `internal/hook/agent_start.go` | REQ-RA-012 ≤500ms budget; YAML parse <50ms typical; well within budget |
| factory.go dispatch case naming collision (`agent-start` vs `subagent-start`) | M | L | `internal/hook/agents/factory.go` | research.md §4.4: stdin uses `agentStart` event; verify exact action string at M3 |
| AC-RA-15 CI assertion (RETIREMENT_INCOMPLETE) creates flaky tests if frontmatter parse fails | L | L | `internal/template/agent_frontmatter_audit_test.go` | YAML parse error → test SKIP with explicit reason, not FAIL; document |
| Manager Agents count documentation inconsistency (8 vs 7 vs 9) | L | M | `CLAUDE.md`, `agent-authoring.md` | M5 substitution: "8" stays (active manager-cycle replaces retired manager-tdd) — net 8 effective |
| lessons.md auto-memory write requires special path discovery | L | L | `~/.claude/projects/{hash}/memory/lessons.md` | M5 use Skill("moai-foundation-core") guidance for path resolution; fallback: present to user via AskUserQuestion |

---

## 6. Solo Mode Path Discipline (4 HARD rules)

Per CLAUDE.local.md §15 + worktree-integration.md HARD rules:

1. [HARD] All write-target paths in agent prompts use project-root-relative form (e.g., `internal/hook/agent_start.go`), not absolute paths to `/Users/goos/MoAI/moai-adk-go/...`
2. [HARD] No `cd /absolute/path` in Bash commands within plan-driven agent invocations
3. [HARD] Reference files (skills via `${CLAUDE_SKILL_DIR}`) may use absolute paths; write targets cannot
4. [HARD] Solo mode (no worktree per user directive) — all changes happen in `feature/SPEC-V3R3-RETIRED-AGENT-001` branch in working tree at `/Users/goos/MoAI/moai-adk-go`

---

## 7. No Implementation Code in Plan Documents

Per CLAUDE.local.md §16 (자가 점검) + spec.md §1.3:

- This plan describes WHAT each milestone delivers, WHY it matters, and at what file location
- Concrete Go function bodies, test fixture data, agent body text, etc. are NOT included in plan.md
- Implementation details are deferred to `/moai run SPEC-V3R3-RETIRED-AGENT-001` execution phase

Plan-auditor verification: search this plan.md for code blocks with full Go function bodies — only stubs/skeletons should appear (e.g., `Handle(ctx, input) (output, error)` signature, not full implementation).

---

## 8. Plan-Audit-Ready Checklist

All 18 criteria PASS per acceptance.md + spec.md cross-reference:

- C1: Frontmatter v0.2.0 (9 required fields) ✅
- C2: HISTORY v0.1.0 entry ✅
- C3: 16 EARS REQs across 5 categories (Ubiquitous 6, Event-Driven 4, State-Driven 3, Optional 1, Unwanted 2) ✅
- C4: 18 ACs with 100% REQ mapping (16/16 REQ → AC traceability matrix in §1.4) ✅
- C5: BC scope clarity (`breaking: false`, `bc_id: []`) — backward-compatible fix ✅
- C6: File:line anchors ≥10 (research.md: 30+, plan.md: 25+) ✅
- C7: Exclusions section present (spec.md §1.3 Non-Goals + §2.2 Out of Scope) ✅
- C8: TDD methodology declared ✅
- C9: mx_plan section (10 tags / 7 files; 3 ANCHOR + 3 NOTE + 2 WARN + 1 TODO + 1 LEGACY) ✅
- C10: Risk table with file-anchored mitigations (spec.md §8: 13 risks; plan.md §5: 14 risks) ✅
- C11: Solo mode path discipline (4 HARD rules) ✅
- C12: No implementation code in plan documents ✅
- C13: Acceptance.md G/W/T format with edge cases (18 ACs covered) ✅
- C14: tasks.md owner roles aligned with TDD ✅ (deferred to optional tasks.md if produced)
- C15: Cross-SPEC consistency (SPEC-V3R2-ORC-001 dependency declared, SPEC-V3R3-HYBRID-001 related) ✅
- C16: BC migration completeness (spec.md §10: no BC, backward-compatible fix) ✅
- C17: 5-layer defect chain documented (research.md §2 + spec.md §1.1) ✅
- C18: External evidence verified (mo.ai.kr file size diff via `ls -la`, manager-cycle absence via `ls`) ✅

---

End of plan.md.
