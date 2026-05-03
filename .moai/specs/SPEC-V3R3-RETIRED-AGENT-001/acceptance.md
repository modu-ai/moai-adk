# SPEC-V3R3-RETIRED-AGENT-001 Acceptance Criteria (Phase 1B)

> Given/When/Then scenarios for retired-stub compatibility fix acceptance.
> Companion to `spec.md` v0.1.0 and `plan.md` v0.1.0.
> All 18 ACs cover the 15 REQs (1-to-1 minimum mapping confirmed in plan.md §1.4).

## HISTORY

| Version | Date       | Author              | Description                                                                                |
|---------|------------|---------------------|--------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow  | 18 G/W/T scenarios + edge cases for SPEC-V3R3-RETIRED-AGENT-001                            |

---

## 1. Acceptance Criteria

### AC-RA-01: manager-cycle.md exists in template + has full frontmatter (REQ-RA-001, REQ-RA-016)

**Given** a clean checkout of the moai-adk-go repository at branch `feature/SPEC-V3R3-RETIRED-AGENT-001` after M2 implementation

**When** the user runs `ls -la internal/template/templates/.claude/agents/moai/manager-cycle.md`

**Then** the file exists
**And** file size is ≥ 5000 bytes (full agent definition, not stub)
**And** the YAML frontmatter contains all required fields per `agent-authoring.md` § Supported Frontmatter Fields:
- `name: manager-cycle`
- `description: <multi-line, includes both DDD and TDD trigger keywords>`
- `tools: <CSV string with at least Read, Write, Edit, MultiEdit, Bash, Grep, Glob, TodoWrite, Skill>`
- `model: sonnet`
- `permissionMode: bypassPermissions`
- `memory: project`
- `skills: <YAML array including moai-workflow-ddd and moai-workflow-tdd>`
- `hooks: <map with PreToolUse, PostToolUse, SubagentStop entries>`
**And** the body content describes both DDD (ANALYZE-PRESERVE-IMPROVE) and TDD (RED-GREEN-REFACTOR) cycles
**And** the body content explains the `cycle_type` parameter handling

**Edge cases**:
- File size ~10KB (matches mo.ai.kr 10245-byte reference): expected
- File contains language-specific examples (Go-only, Python-only): rejected by 16-language neutrality check (M2 manual review)
- Hook action names use `tdd-*` legacy naming instead of `cycle-*`: rejected by M2 quality check

**Test Anchor**: `internal/template/manager_cycle_present_test.go TestManagerCyclePresentInEmbeddedFS`

---

### AC-RA-02: manager-tdd.md retired stub has all 5 standardized fields (REQ-RA-002)

**Given** a clean checkout of the moai-adk-go repository at branch `feature/SPEC-V3R3-RETIRED-AGENT-001` after M2 implementation

**When** the user reads `internal/template/templates/.claude/agents/moai/manager-tdd.md`

**Then** the YAML frontmatter contains all 5 retirement standardization fields:
- `retired: true` (boolean, not string)
- `retired_replacement: manager-cycle` (string, exact match to active replacement file basename)
- `retired_param_hint: "cycle_type=tdd"` (string, parameter invocation hint)
- `tools: []` (explicit empty YAML array)
- `skills: []` (explicit empty YAML array)
**And** `name: manager-tdd` is preserved
**And** the legacy custom field `status: retired` is REMOVED
**And** the body content describes (a) retirement reason, (b) replacement agent name, (c) old → new invocation pattern

**Edge cases**:
- `retired: "true"` (string): rejected — must be boolean per `agent_frontmatter_audit_test.go` audit
- `tools:` (no value, just key): rejected — must be explicit `[]`
- `retired_replacement: cycle-manager` (typo): rejected — must match real file basename
- Body content is just retirement notice without migration table: SHOULD include migration table (mo.ai.kr 976-byte version pattern)

**Test Anchor**: `internal/template/agent_frontmatter_audit_test.go TestAgentFrontmatterAudit`

---

### AC-RA-03: `make build` regenerates embedded FS with both files (REQ-RA-003)

**Given** the developer has completed M2 (manager-cycle.md NEW + manager-tdd.md MODIFIED)

**When** the developer runs `make build` from `/Users/goos/MoAI/moai-adk-go`

**Then** the command succeeds with exit code 0
**And** `internal/template/embedded.go` is regenerated with new content (`git diff internal/template/embedded.go` shows non-empty changes)
**And** the regenerated file includes both `manager-cycle.md` and the standardized `manager-tdd.md`

**When** the developer runs `go test ./internal/template/ -run "TestManagerCyclePresent|TestAgentFrontmatterAudit" -v`

**Then** both tests PASS
**And** test output shows:
- `manager-cycle.md` body length ≥ 5000 bytes
- `manager-tdd.md` frontmatter has all 5 retirement fields

**Edge cases**:
- `make build` fails with embedded FS regeneration error: rollback M2 + investigate
- Test passes locally but fails in CI: investigate `go:embed` directive ordering
- `embedded.go` regenerated but tests still see old content: clear Go build cache (`go clean -cache`)

**Test Anchor**:
- `internal/template/manager_cycle_present_test.go TestManagerCyclePresentInEmbeddedFS`
- `internal/template/agent_frontmatter_audit_test.go TestAgentFrontmatterAudit`

---

### AC-RA-04: SubagentStart hook returns block decision for retired agent (REQ-RA-004, REQ-RA-007)

**Given** the M3 implementation deployed (`internal/hook/agent_start.go` + factory.go dispatch + handle-subagent-start.sh.tmpl exit propagation)
**And** an agent definition file at `.claude/agents/moai/manager-tdd.md` with `retired: true` frontmatter

**When** Claude Code spawns a subagent of type `manager-tdd` and triggers SubagentStart hook
**And** the hook stdin JSON contains `{"agentType": "manager-tdd", "agentName": "manager-tdd", "agent_id": "<uuid>"}`

**Then** the hook returns exit code 2
**And** stdout contains JSON `{"decision": "block", "reason": "agent manager-tdd retired (SPEC-V3R2-ORC-001), use manager-cycle with cycle_type=tdd"}`
**And** the spawn is blocked at the runtime layer (worktree allocation does NOT proceed)
**And** Claude Code surfaces the block reason to the orchestrator

**Edge cases**:
- Hook fires after worktree allocation (Claude Code spec ambiguity): SubagentStart timing verified at M2/M3 implementation; if late, fall back to PreToolUse hook on Agent tool with matcher
- Multiple `retired: true` agents in same session: each hook invocation independent
- Stdin malformed JSON: hook returns exit 0 (allow, fail-safe); does not crash

**Test Anchor**:
- `internal/hook/agent_start_test.go TestAgentStartBlocksRetiredAgent`
- Manual integration test in `/tmp/test-project` after M3

---

### AC-RA-05: launcher.go rejects empty-object worktreePath (REQ-RA-005, REQ-RA-010)

**Given** the M4 implementation deployed (`validateWorktreeReturn` helper in `internal/cli/launcher.go`)

**When** an Agent() invocation with `isolation: "worktree"` returns a value where:
- `worktreePath` is empty struct `{}` or nil pointer or empty string `""`
- `worktreeBranch` is undefined / empty string

**And** the orchestrator's wrapper layer calls `validateWorktreeReturn(result, "worktree")`

**Then** the function returns a non-nil error
**And** the error message contains the sentinel string `WORKTREE_PATH_INVALID`
**And** the error message includes the agent name + invocation context (which Agent() call returned the broken value)
**And** the broken value is NOT propagated to fallback re-delegation logic

**Edge cases**:
- `worktreePath: "/tmp/abc123"` (valid): returns nil (no error) — happy path
- `worktreePath: "/{}/{}"` (literal `{}`): returns error — string substitution byproduct from old code; confirms the bug class is detected
- `isolation: "none"` (no worktree requested): validation skipped (returns nil); `worktreePath` may be empty
- `isolation: "worktree"` + valid worktreePath but undefined branch: branch is included in error message but doesn't fail solely on branch (worktreeBranch may be optional)

**Test Anchor**: `internal/cli/launcher_worktree_validation_test.go TestValidateWorktreeReturnRejectsEmptyObject`

---

### AC-RA-06: path interpolation uses text/template, no string concat (REQ-RA-006)

**Given** the M4 implementation deployed (path interpolation refactor across 3-5 callsites)

**When** the developer runs `grep -rn 'fmt.Sprintf(.*"/.*{.*}.*/' internal/cli/ 2>/dev/null` (looking for legacy string concat patterns)

**Then** the result is empty (no callsites match the legacy pattern)

**When** the developer runs `grep -rn 'text/template\|template.New' internal/cli/ 2>/dev/null`

**Then** the result includes new uses introduced by M4 (path template definitions)

**When** an Agent() return value has `worktreePath` as a non-string type (e.g., `map[string]interface{}{}`) and the path template is executed with this value

**Then** the template execution returns a typed error (e.g., `template: cannot execute template with non-string value`) instead of the string `"[object Object]"` or `"{}"`

**Edge cases**:
- `worktreePath` is `*string` pointer with nil value: template returns "<nil>" — must be caught by validateWorktreeReturn before template execution
- `worktreePath` is properly typed `string`: template executes normally
- Multiple callsites use different template definitions: each compiled once at package init

**Test Anchor**: `internal/cli/launcher_worktree_validation_test.go TestPathTemplateRejectsNonStringValue`

---

### AC-RA-07: retired-rejection guard returns proper JSON + exit 2 (REQ-RA-004, REQ-RA-007)

**Given** the M3 implementation deployed
**And** stdin JSON `{"agentType": "manager-tdd", "agentName": "manager-tdd", "agent_id": "test-uuid-123"}`

**When** the developer runs:
```bash
echo '{"agentType":"manager-tdd","agentName":"manager-tdd","agent_id":"test-uuid-123"}' | bash .claude/hooks/moai/handle-subagent-start.sh
```

**Then** the bash exit code is 2
**And** stdout is valid JSON parseable by `jq -e '.decision == "block"'`
**And** the JSON `reason` field contains:
- The string `manager-tdd` (the retired agent name)
- The string `manager-cycle` (the replacement)
- The string `cycle_type=tdd` (the param hint)
- Optionally a SPEC reference like `SPEC-V3R2-ORC-001`

**Edge cases**:
- Wrapper script `exec` form retained instead of `exit $?` non-exec form: exit code may not propagate; M3 must use non-exec form
- Hook timeout exceeded (5s default): hook returns timeout error to Claude Code; spawn proceeds (fail-safe)
- moai binary not found: wrapper falls back to `~/go/bin/moai` or `~/.local/bin/moai` per existing logic

**Test Anchor**:
- `internal/hook/agent_start_test.go TestAgentStartBlocksRetiredAgent`
- Manual shell integration test post-M3

---

### AC-RA-08: unknown agent name bypasses guard (exit 0) (REQ-RA-008)

**Given** the M3 implementation deployed
**And** the stdin JSON contains `{"agentType": "unknown-agent-xyz", "agentName": "unknown-agent-xyz", "agent_id": "test-uuid-456"}`

**When** the SubagentStart hook handler is invoked

**Then** the handler attempts to locate the agent file at `.claude/agents/moai/unknown-agent-xyz.md`
**And** the file does not exist (Stat returns ENOENT)
**And** the handler returns exit code 0 (allow)
**And** stdout is empty or contains a permissive output (no `decision: block`)
**And** the spawn proceeds normally (Claude Code's default behavior for unknown agents)

**Edge cases**:
- Agent file exists at `.claude/agents/<name>.md` (root level, not moai/) → handler tries both locations; if found at either, parses frontmatter
- Agent name with special characters (e.g., `agent/with/slash`): rejected at hook layer with appropriate error; spawn fails (Claude Code's responsibility)
- Empty `agentName` in stdin: handler returns exit 0 (allow, fail-safe)

**Test Anchor**: `internal/hook/agent_start_test.go TestAgentStartAllowsUnknownAgent`

---

### AC-RA-09: factory.go dispatch for agent-start event (REQ-RA-009)

**Given** the M3 implementation deployed (`internal/hook/agents/factory.go` extension)

**When** the developer reads `internal/hook/agents/factory.go`

**Then** the file contains a switch case branch dispatching `case "agent-start":` (or equivalent action name) to `NewAgentStartHandler()`
**And** the case is placed before `default` (so it takes precedence)
**And** unknown action strings still fall through to `default_handler.go`

**When** the developer runs `go test ./internal/hook/agents/ -run "TestFactoryDispatch" -v`

**Then** the test PASSes (factory returns the correct handler type for `agent-start` action)
**And** existing factory dispatch tests for other actions (e.g., `tdd-completion`, `backend-validation`) still PASS (no regression)

**Edge cases**:
- Action string casing mismatch (`Agent-Start` vs `agent-start`): factory uses lowercase comparison
- Multiple new action strings (`agent-start`, `subagent-start` synonyms): handler accepts both, single underlying handler
- Plugin-defined agents bypass factory dispatch (per `agent-authoring.md` Plugin Agent Limitations): no impact (plugins can't define hooks)

**Test Anchor**: `internal/hook/agent_start_test.go TestAgentStartHandlerRoutesViaFactory`

---

### AC-RA-10: WORKTREE_PATH_INVALID sentinel emitted with context (REQ-RA-005, REQ-RA-010)

**Given** the M4 implementation deployed
**And** an Agent() return value with `worktreePath: ""`, agent name `manager-cycle`, isolation `worktree`

**When** the orchestrator's wrapper layer calls `validateWorktreeReturn(result, "worktree", agentName)`

**Then** the function returns an error whose `Error()` method includes:
- The sentinel string `WORKTREE_PATH_INVALID`
- The agent name `manager-cycle`
- A description of why the path is invalid (e.g., "empty string", "nil pointer", "non-string type")
**And** the error is NOT silently swallowed by the caller
**And** the error propagates up to the orchestrator level for logging + user visibility

**Edge cases**:
- Validation called with `isolation: "none"` and empty worktreePath: returns nil (validation skipped) — does NOT emit sentinel
- Multiple agents fail validation in same session: each error independent, sentinel always present
- Error wrapped via `fmt.Errorf("...: %w", err)` chain: `errors.Is()` or `errors.As()` finds the sentinel correctly

**Test Anchor**: `internal/cli/launcher_worktree_validation_test.go TestValidateWorktreeReturnSentinel`

---

### AC-RA-11: retired stub body describes reason + replacement + migration (REQ-RA-011)

**Given** the M2 implementation deployed (`manager-tdd.md` retired stub)

**When** the user reads `internal/template/templates/.claude/agents/moai/manager-tdd.md` body content (post-frontmatter)

**Then** the body content includes:
- A retirement reason statement (e.g., "consolidated into manager-cycle as part of SPEC-V3R2-ORC-001")
- The replacement agent name (`manager-cycle`)
- The old → new invocation pattern table or example (e.g., "Old: Use the manager-tdd subagent ... | New: Use the manager-cycle subagent with cycle_type=tdd ...")
- A reference to documentation for the replacement (e.g., link to `.claude/agents/moai/manager-cycle.md`)

**Edge cases**:
- Body too brief (<200 chars): rejected — must include all three components
- Body too verbose (>2000 chars): warning, not rejection — keep stub minimal but informative
- Body contains code examples in language-specific syntax: rejected by 16-language neutrality

**Test Anchor**: Manual review at M2; optionally `internal/template/agent_frontmatter_audit_test.go TestRetiredAgentBodyContains`

---

### AC-RA-12: retired-rejection guard ≤500ms response time (REQ-RA-012)

**Given** the M3 implementation deployed
**And** a benchmark test invoking `AgentStartHandler.Handle()` with retired-stub frontmatter input

**When** the developer runs `go test ./internal/hook/ -run "TestAgentStartHandlerPerformance" -bench=. -benchmem -benchtime=10s`

**Then** the average latency per Handle() invocation is ≤500ms
**And** memory allocation per call is bounded (≤100KB)
**And** no blocking I/O calls (network, large file reads) occur in the handler

**When** comparing to the mo.ai.kr 21:14:54 incident (11.4s observed)

**Then** the new guard is ≥22x faster (11400ms ÷ 500ms)
**And** the time savings are realized at every retired-agent invocation

**Edge cases**:
- YAML parser slow on first invocation (lazy init): warm up in test setup; benchmark only steady-state
- Disk I/O slow on cold cache: include cache warmup in benchmark setup
- CI runner (e.g., GitHub Actions) is consistently slower than local: budget includes 2x slack (target 500ms, soft fail above 1000ms)

**Test Anchor**: `internal/hook/agent_start_test.go TestAgentStartHandlerPerformance`

---

### AC-RA-13: all 6 documentation references substituted (REQ-RA-013)

**Given** the M5 implementation deployed (documentation substitution across 6 files)

**When** the developer runs `grep -rn 'manager-tdd\|manager-ddd' internal/template/templates/.claude/ internal/template/templates/CLAUDE.md 2>/dev/null`

**Then** the result contains:
- `manager-tdd.md` itself (frontmatter `name: manager-tdd` and body migration notes — expected, NOT a substitution target)
- `manager-ddd.md` itself (full agent definition — expected, OUT OF SCOPE per spec.md §1.3)
- ZERO mentions in:
  - `CLAUDE.md` §4 Manager Agents listing (manager-tdd should not appear in active agent list)
  - `CLAUDE.md` §5 Agent Chain (`manager-ddd or manager-tdd subagent` should be `manager-cycle subagent`)
  - `agent-authoring.md` Manager Agents (8) section (manager-tdd entry replaced by manager-cycle entry)
  - `agent-hooks.md` Agent Hook Actions table (manager-tdd row replaced by manager-cycle row)
  - `spec-workflow.md` Phase Overview table (`manager-ddd/tdd` → `manager-cycle`)
  - `manager-strategy.md` Code implementation line (`manager-ddd or manager-tdd` → `manager-cycle`)
  - `manager-ddd.md` body inline references (2 mentions to manager-tdd → `manager-cycle with cycle_type=tdd`)

**When** the developer runs `grep -rn 'manager-cycle' internal/template/templates/.claude/ internal/template/templates/CLAUDE.md 2>/dev/null`

**Then** the result includes new mentions across all 6 files (and the new manager-cycle.md file itself)

**Edge cases**:
- A reference is missed (e.g., a typo or alternate naming): caught by `agent_frontmatter_audit_test.go TestNoOrphanedManagerTDDReference`
- A reference appears in a quoted code block (e.g., example showing legacy invocation): preserve quoting; substitute only outside code fences

**Test Anchor**: Manual grep verification at M5 + `internal/template/agent_frontmatter_audit_test.go` reference audit subtest

---

### AC-RA-14: `moai agents list --retired` subcommand surfaced or deferred (REQ-RA-014)

**Given** M5 decision point (in-scope vs deferred)

**Branch A — In Scope**:
**When** the user runs `moai agents list --retired`
**Then** the CLI outputs a list of all agents with `retired: true` frontmatter
**And** the output format is structured (table or JSON)
**And** for each retired agent, the output includes name + replacement + retirement SPEC reference

**Branch B — Deferred**:
**When** plan-auditor reviews this AC at M5 decision time
**Then** the user explicitly approves deferral via AskUserQuestion
**And** the deferral is documented in `spec-compact.md §Open Items`
**And** a follow-up SPEC `SPEC-V3R3-AGENTS-CLI-001` (가칭) issue is created in GitHub

**Edge cases**:
- User wants the feature in-scope but M5 implementation cost is high: AskUserQuestion at M5 confirms scope decision
- Deferred but a follow-up SPEC is never created: tracked via TODO in lessons.md #11

**Test Anchor**: M5 manual decision + (Branch A) `internal/cli/agents_test.go TestListRetired`

---

### AC-RA-15: CI rejects RETIREMENT_INCOMPLETE_<agent> if any of (a)-(d) missing (REQ-RA-016)

**Given** the M1+M2+M5 implementation deployed (`agent_frontmatter_audit_test.go` + standardized retirement workflow)

**Setup**: A simulated incomplete retirement scenario for testing — temporarily add a 5th provider file to the codebase and remove its corresponding manager-cycle (or similar replacement) before running tests.

**When** a future agent retirement PR fails one of:
(a) Missing standardized frontmatter (`retired: true`, replacement, hint, empty arrays)
(b) Outdated documentation reference (e.g., `agent-hooks.md` still mentions retired agent in active table)
(c) Missing audit assertion (no entry in `TestAgentFrontmatterAudit`)
(d) Missing active replacement file (e.g., manager-tdd retired but manager-cycle.md absent)

**Then** the test `TestRetirementCompletenessAssertion` fails with sentinel `RETIREMENT_INCOMPLETE_<agent-name>` in the error message
**And** CI builds fail (`go test ./...` non-zero exit)
**And** the failure clearly identifies which of (a)-(d) is missing

**When** all four conditions are satisfied

**Then** the test PASSes for that agent

**Edge cases**:
- A retired agent has no documented replacement (intentional): test FAILs with custom message; user must explicitly mark in test as edge case
- Replacement file exists but with retired stub itself (transitive retirement chain): test FAILs; transitive retirement is anti-pattern

**Test Anchor**: `internal/template/agent_frontmatter_audit_test.go TestRetirementCompletenessAssertion`

---

### AC-RA-16: manager-cycle.md spawn via Agent() succeeds with valid worktreePath (REQ-RA-001, REQ-RA-013)

**Given** the M5 implementation deployed (full template + hook chain functional)
**And** a clean test project at `/tmp/test-project-retired-agent-001` initialized with `moai init`

**When** the orchestrator invokes `Agent({subagent_type: "manager-cycle", isolation: "worktree", cycle_type: "tdd", ...})`

**Then** the Agent() call succeeds within reasonable latency (≤30s for spawn + minimal task)
**And** the return value has:
- `worktreePath`: a non-empty string pointing to a valid directory under `.claude/worktrees/`
- `worktreeBranch`: a non-empty string (e.g., `cycle-abc123` or similar auto-generated name)
**And** `validateWorktreeReturn` accepts the return value without raising `WORKTREE_PATH_INVALID`
**And** the agent body has access to all configured tools (Read, Write, Edit, etc.)

**Edge cases**:
- Worktree allocation fails (disk full, permission denied): Agent() returns specific error; not WORKTREE_PATH_INVALID
- `cycle_type` parameter not provided: agent body raises validation error per its SEMAP contract
- Multiple parallel manager-cycle spawns in same session: each gets independent worktree (no conflicts)

**Test Anchor**: Integration test at M5 in `/tmp/test-project-retired-agent-001` (manual or via Go integration test)

---

### AC-RA-17: manager-tdd retired stub spawn via Agent() blocked at SubagentStart layer (REQ-RA-002, REQ-RA-007) — REGRESSION TEST FOR mo.ai.kr 21:14:54

**Given** the full M3+M5 implementation deployed
**And** a test project with the standardized manager-tdd retired stub deployed
**And** SubagentStart hook chain functional

**When** the orchestrator invokes `Agent({subagent_type: "manager-tdd", isolation: "worktree", ...})`

**Then** the SubagentStart hook fires
**And** the hook handler detects `retired: true` in manager-tdd.md frontmatter
**And** the hook returns block decision (per AC-RA-04)
**And** the spawn is blocked
**And** the orchestrator surfaces the block reason: "agent manager-tdd retired (SPEC-V3R2-ORC-001), use manager-cycle with cycle_type=tdd"
**And** worktree allocation does NOT proceed (no orphan worktree directory created)
**And** the response time is ≤500ms (per AC-RA-12)

**Critical regression assertion**:
**And** the orchestrator does NOT receive a response with `worktreePath: {}` or `worktreeBranch: undefined`
**And** the post-block fallback path does NOT propagate broken state to a manager-cycle re-spawn
**And** no error like `[ERROR] Path "/{}/{}" does not exist` is logged

**Edge cases**:
- SubagentStart hook is bypassed (e.g., hook chain disabled): worktreePath validation (REQ-RA-005) catches the empty-object return; defense-in-depth ensures no path-string bug
- User has stale manager-tdd.md from old template (without `retired: true`): full agent spawns normally; no block (this is the pre-SPEC behavior, and our new behavior only triggers when retired field is present)

**Test Anchor**:
- `internal/hook/agent_start_test.go TestAgentStartBlocksRetiredAgent`
- Integration test at M5 reproducing mo.ai.kr scenario in `/tmp/test-project`

---

### AC-RA-18: Agent() returning empty-object worktreePath triggers WORKTREE_PATH_INVALID instead of `[ERROR] Path "/{}/{}" does not exist` (REQ-RA-005, REQ-RA-006)

**Given** the M4 implementation deployed (worktreePath validation + text/template path interpolation)
**And** a synthetic test scenario where Agent() returns `worktreePath: {}` (empty object) — simulating a malfunctioning agent or pre-fix retired stub spawn

**When** the wrapper layer receives the return value

**Then** the validation layer raises `WORKTREE_PATH_INVALID` error before any path interpolation occurs
**And** the error message includes:
- `WORKTREE_PATH_INVALID` sentinel
- The agent name
- Description of the invalid value (e.g., "worktreePath is empty object {}")
**And** NO subsequent log line contains `Path "/.*/{}/.*"` or `Path "/{}/{}/"` patterns
**And** NO subsequent log line contains `[ERROR] Path "<root>/{}/{}" does not exist`
**And** the broken value is NOT propagated to fallback re-delegation
**And** the orchestrator surfaces the WORKTREE_PATH_INVALID error to the user with actionable context (which Agent() call returned the broken value)

**Critical regression assertion**:
**And** if a downstream consumer attempts path interpolation with the broken value, the text/template execution returns a typed error (per AC-RA-06) rather than producing the literal string `"{}"` in the path

**Edge cases**:
- worktreePath is null pointer (vs empty object): same WORKTREE_PATH_INVALID error
- worktreePath is `*string` pointing to empty string: same error (validation checks dereferenced value)
- worktreePath is properly typed but with unusual content (e.g., relative path): not validated by this layer (just type check); path correctness is downstream concern

**Test Anchor**:
- `internal/cli/launcher_worktree_validation_test.go TestValidateWorktreeReturnRejectsEmptyObject`
- `internal/cli/launcher_worktree_validation_test.go TestPathTemplateRejectsNonStringValue`
- Integration regression test at M5 reproducing mo.ai.kr Layer 4 in `/tmp/test-project`

---

## 2. Definition of Done (DoD)

본 SPEC implementation은 다음 모두 충족 시 done:

- [ ] 모든 18 ACs PASS (G/W/T scenarios verified)
- [ ] M1-M5 milestones complete per plan.md
- [ ] `make build` clean (embedded FS regenerated)
- [ ] `go test ./...` 전체 PASS (no regression in existing tests + 4 new test files)
- [ ] `go vet ./...` clean
- [ ] `golangci-lint run ./...` clean
- [ ] `make install` 후 manual smoke test in `/tmp/test-project`:
  - `moai update` syncs new template files
  - `manager-cycle.md` exists in deployed `.claude/agents/moai/`
  - `manager-tdd.md` is retired stub (small file, has `retired: true`)
  - Attempting to spawn `manager-tdd` produces block message + ≤500ms response
- [ ] mo.ai.kr 사이드 프로젝트에 `moai update` 실행 후 `Path "/{}/{}"` 에러 재현 안 됨 (사용자 확인)
- [ ] CHANGELOG entry per spec.md §10 added
- [ ] PR created with `--label type:fix,priority:P0,area:templates,area:hooks`
- [ ] plan-auditor review-1.md ≥ 0.80 score
- [ ] All 4 review-1.md identified defects resolved (or explicitly deferred with justification)

---

End of acceptance.md.
