---
id: SPEC-WORKTREE-BRANCH-GUARD-001
title: Main-Checkout Branch-State Guard via PreToolUse Conditional Deny — Implementation Plan
version: 0.1.2
status: in-progress
created: 2026-07-28
updated: 2026-07-28
author: manager-spec
priority: High
phase: plan
module: hook
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "hook, pretool, branch-guard, worktree, main-checkout, template-mirror, sanitized-pair, census-p1b"
related_specs:
  - SPEC-WORKTREE-001
  - SPEC-WORKTREE-002
  - SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001
  - SPEC-V3R5-LATE-BRANCH-001
---

# Implementation Plan — SPEC-WORKTREE-BRANCH-GUARD-001

## §A. Context

**Verified preconditions** (orchestrator-observed 2026-07-28, all 5 confirmed
during plan-phase research):

| # | Premise | Evidence |
|---|---------|----------|
| 1 | settings layer is no-op | `.claude/settings.json:309` `bypassPermissions`; `internal/template/templates/.claude/settings.json.tmpl:439` `acceptEdits` |
| 2 | PreToolUse infra exists | `.claude/settings.json:39` PreToolUse array; `.claude/hooks/moai/handle-pre-tool.sh` exists (4383 bytes); `internal/hook/pre_tool.go` handler factory `NewPreToolHandler`/`NewPreToolHandlerWithScanner` |
| 3 | static deny already used | `.claude/settings.json:341` `Bash(git branch:*)`; `:350-353` checkout/switch/stash/merge; `:457` `git reset --hard:*`; `:460` `git rebase -i:*` |
| 4 | `.worktreeinclude` diverged | repo-root 18 lines (mentions `teammateMode`); `internal/template/templates/.worktreeinclude` 26 lines (has "Add your own..." guidance, omits `teammateMode`) — diff is 2 prose blocks |
| 5 | `workflow.worktree` config | `.moai/config/sections/workflow.yaml:112-120` — `auto_cleanup: true`, `auto_create: false`, `auto_merge: true`, `session_name_pattern: moai-{ProjectName}-{SPEC-ID}`, `tmux_preferred: true` |

**Research findings (determinative for the design)**:

1. **`checkBashCommand` is the fast regex path** — `internal/hook/pre_tool.go:623-655`
   iterates `policy.DangerousBashPatterns` and returns `DecisionDeny` on first
   match. It does NOT invoke the quality gate. Census P1-B's 10.58s latency was
   measured on the `quality.IsGitCommit` path (`pre_tool.go:388-397`), which
   only fires for `git commit` invocations — structurally inaccessible to
   branch-state commands.

2. **`HookInput.AgentType` docstring** — `internal/hook/types.go:226`: "Custom
   agent name if `--agent` flag used". Main-thread `--agent` invocation only.
   `Agent(manager-git, ...)` sub-agent spawns do NOT reliably populate
   `agent_type` in the subagent's PreToolUse events. **Contradicts a verified
   precondition** — closed by REQ-WBG-011b (sentinel env var).

3. **Mirror test allowlist + sanitized-pair registry** — the source rule
   carries SPEC-ID + REQ tokens for traceability (REQ-WBG-006), so it is NOT
   §25-clean. §25 (`TestTemplateNoInternalContentLeak`) forbids a SPEC-ID in
   the template mirror, so byte-parity between source and mirror is
   impossible. The sanitized-pair pattern is the resolution:
   `internal/template/sanitized_pair_parity_test.go` carries an explicit
   `sanitizedPairPaths` registry (analogous to `workflowOptMirroredPaths`).
   Enrolling `.claude/rules/moai/workflow/main-checkout-branch-guard.md` in
   `sanitizedPairPaths` (NOT in `workflowOptMirroredPaths`) lets
   `TestSanitizedPairParity` enforce doctrine parity between source and
   sanitized mirror while `TestTemplateNoInternalContentLeak` enforces the
   mirror's §25 cleanliness. Matches the `runtime-recovery-doctrine.md` /
   `zone-registry.md` precedent.

4. **git discriminant portability** — `--path-format=absolute` requires git
   2.31+ (March 2021). Local environment runs git 2.50.1 (Apple Git-155) which
   supports it. Verified empirically: primary checkout → both forms return
   `/Users/goos/MoAI/moai-adk-go/.git` (equal); worktree
   `/Users/goos/.moai/worktrees/moai-adk/branch-guard` → git-dir
   `/Users/goos/MoAI/moai-adk-go/.git/worktrees/branch-guard` ≠ git-common-dir
   `/Users/goos/MoAI/moai-adk-go/.git` (different). Fallback
   `git rev-parse --absolute-git-dir` returns the absolute path; bare
   `git rev-parse --git-common-dir` returns the relative path `.git` and needs
   `filepath.Abs(cwd, ...)`.
- 5. **Existing static deny contradiction** — `.claude/settings.json:341,
   350-353, 457, 460` denies `git branch`, `git checkout`, `git switch`,
   `git stash`, `git merge`, `git reset --hard`, `git rebase -i`. This
   contradicts `manager-git.md:117-123` Phase D (`git checkout main && git
   reset --hard origin/main`). The SPEC retains these as defense-in-depth
   (REQ-WBG-013) but does NOT remove them — their reconciliation is census P1-B
   residual debt.

## §B. Known Issues

- **[NEEDS CLARIFICATION: orchestrator-side MOAI_BRANCH_GUARD_EXEMPT=1 spawn
  contract]**: the sentinel env var approach (REQ-WBG-011b) requires the
  orchestrator to set `MOAI_BRANCH_GUARD_EXEMPT=1` when spawning
  `Agent(manager-git, ...)` for Phase D work. The orchestrator-side
  spawn-contract change is a one-line `os.Setenv` or `Env: []string{...}` in
  the `Agent()` spawn — but the exact spawn-point code lives in the orchestrator
  (deferred to a follow-up SPEC per the orchestrator's iteration-2 decision).

  **Interim Phase-D break (honest framing)**: until the follow-up SPEC lands
  the per-spawn env-var injection, manager-git Phase D invocations via
  sub-agent spawn WILL BE DENIED by this hook. Local-dev interim impact: Phase
  D steps (`git checkout main && git reset --hard origin/main`, per
  `manager-git.md:117-123`) must run via main-thread
  `claude --agent manager-git` (where `HookInput.AgentType == "manager-git"`
  populates correctly), NOT via sub-agent spawn, until the follow-up SPEC.
  User-project impact: the retained static deny at
  `.claude/settings.json:341, 350-353, 457, 460` (REQ-WBG-013) ALSO blocks
  Phase D and is NOT addressed by this SPEC's exemption — tracked as census
  P1-B residual debt per spec.md §E. This is the **same functional consequence
  and same valence** as the static-deny break documented in spec.md §A (lines
  46-50): both are an interim break pending their respective fixes, NOT a
  "safe default". A follow-up SPEC (TBD ID) will address the orchestrator-side
  spawn contract for per-spawn env-var injection.

- **B-2**: the `.worktreeinclude` template SSOT contains an "Add your own
  project-local files below..." guidance block that the local copy omits. The
  reconciliation direction is template → local (Template-First Rule). The local
  copy's `teammateMode` mention is redundant prose (§22.3 documents teammateMode
  as a `settings.local.json` field; `.worktreeinclude` already copies that file).

- **B-3**: the existing static deny at `.claude/settings.json:341` for
  `Bash(git branch:*)` will STILL fire even after the hook is in place (it is
  defense-in-depth per REQ-WBG-013). The hook + static deny compose: the hook
  runs first (PreToolUse), and if it allows (worktree or exempt agent), the
  static deny in settings.json is the next gate. This means legitimate worktree
  `git branch` use STILL hits the static deny unless the user amends
  settings.json. **This is a known limitation**, NOT a regression introduced by
  this SPEC; the static deny was already present. Reconciliation is tracked in
  census P1-B residual debt (out of scope per §E).

## §C. Pre-flight Checks (run-phase entry)

Before M1:

1. `git -C /Users/goos/MoAI/moai-adk-go rev-parse --path-format=absolute --git-dir`
   → exits 0, prints `.../.git`.
2. `git -C /Users/goos/MoAI/moai-adk-go rev-parse --path-format=absolute --git-common-dir`
   → exits 0, prints same path (primary checkout).
3. `go test ./internal/hook/... -run TestPreTool -count=1` → establishes
   baseline test pass.
4. `make build` → confirms binary rebuilds cleanly.

## §D. Constraints Carried Into Run-Phase

- **Template-First**: edit `internal/template/templates/...` first, `make build`,
  sync to local, run `go test ./internal/template/...`.
- **Sanitized-pair invariant**: source rule retains SPEC-ID + REQ tokens
  (traceability per REQ-WBG-006); template mirror is §25-sanitized (SPEC-ID
  line → generic prose, SPEC-ID cross-reference bullet dropped); the rule is
  enrolled in `sanitizedPairPaths` (NOT `workflowOptMirroredPaths`)
  per REQ-WBG-007.
- **5s timeout budget**: the hook handler MUST complete in measured wall-time
  ≤ 500ms per invocation (REQ-WBG-010). The fast regex path meets this trivially;
  the AC encodes a measurement.
- **Sentinel prefix on deny reason**: `BRANCH_GUARD_VIOLATION:` prefix so the
  orchestrator can pattern-match without parsing the full reason.

## §E. Self-Verification (run-phase §E.1 — placeholder)

Run-phase will populate §E.1 audit-ready signal: cite commit SHAs, lint exit 0,
test pass count, mirror-parity test green.

## §F. Milestones (ordered by decision-reversibility — highest-change first)

### M1 — Discriminant + Exemption Mechanism (design-risk-heavy)

**Why first**: the exemption mechanism (`agent_type` vs env var) and the
discriminant code path (`--path-format` vs fallback) are the two decisions most
likely to change based on empirical results. Landing them first gives the
run-phase audit a stable foundation.

**Files**:

- `internal/hook/branch_guard.go` (NEW) — the discriminant function
  `isPrimaryCheckout(projectDir string) (bool, error)` and the exemption
  resolver `isExemptAgent(input *HookInput) bool`.
- `internal/hook/branch_guard_test.go` (NEW) — table-driven tests for both
  functions. The older-git fallback path is tested by injecting a mock
  `exec.Command` runner into the `isPrimaryCheckout` dispatcher: the mock
  simulates the primary `--path-format=absolute` call exiting non-zero
  ("unknown flag"), then the dispatcher MUST fall back to
  `--absolute-git-dir` + cwd-normalized `--git-common-dir`. Direct invocation
  of the fallback function is INSUFFICIENT (vacuous pass — bypasses the
  dispatcher). The mock runner is injected via a package-level
  `var execCommand = exec.Command` indirection that the test swaps.

**Branch-state pattern set** (encoded as `*regexp.Regexp`):

| Pattern | Regex (case-insensitive) | Deny reason suffix |
|---------|--------------------------|--------------------|
| `git switch` | `\bgit\s+switch\b` | `git switch` |
| `git checkout <branch>` | `\bgit\s+checkout\s+(-b\s+)?\S` | `git checkout <branch/-b>` |
| `git branch` (create/delete) | `\bgit\s+branch\s+(-[dDmM]\s+)?\S` | `git branch` |
| `git reset --hard` | `\bgit\s+reset\s+--hard\b` | `git reset --hard` |
| `git stash` (push/pop) | `\bgit\s+stash(\s+(push|pop|apply|drop)\b)?` | `git stash` |
| `git rebase` (onto checked-out) | `\bgit\s+rebase\b` | `git rebase` |
| `git merge` (onto checked-out) | `\bgit\s+merge\b` | `git merge` |

**Permitted** (do NOT match): `git checkout -- <path>` (path-restore, not
branch-state), `git checkout <file>` (single-file checkout), `git branch -v`
(list-only). The regex set is tuned to require a non-flag token after the
subcommand to avoid catching list-only invocations.

**Exemption resolver**:

```go
func isExemptAgent(input *HookInput) bool {
    if input.AgentType == "manager-git" {
        return true
    }
    if os.Getenv("MOAI_BRANCH_GUARD_EXEMPT") == "1" {
        return true
    }
    return false
}
```

**Acceptance**: M1 is complete when the discriminant correctly classifies a
primary checkout, a worktree, and a non-git directory in tests, AND the
exemption resolver returns true for both `AgentType == "manager-git"` and
`MOAI_BRANCH_GUARD_EXEMPT=1`.

### M2 — Hook Handler Integration (wiring into pre_tool.go)

**Files**:

- `internal/hook/pre_tool.go` — extend `preToolHandler.Handle` to invoke the
  branch-guard check AFTER the existing `checkBashCommand` and BEFORE returning
  the default allow. The check runs only for `input.ToolName == "Bash"` (already
  gated above).
- `internal/hook/branch_guard.go` (continued) — `checkBranchState(input *HookInput,
  projectDir string) (decision string, reason string)` returns `DecisionDeny`
  when all three conditions hold.
- `internal/hook/pre_tool_test.go` (EXTEND) — add tests covering REQ-WBG-001
  through REQ-WBG-013.

**Integration point** (insertion in `Handle`):

```go
// After checkBashCommand, before falling through to NewSafeDefaultOutput:
if input.ToolName == "Bash" {
    if decision, reason := checkBranchState(input, h.projectDir); decision == DecisionDeny {
        slog.Warn("branch guard denied",
            "tool_name", input.ToolName,
            "session_id", input.SessionID,
            "reason", reason,
        )
        return NewDenyOutput(reason), nil
    }
}
```

**Sentinel reason format**: `"BRANCH_GUARD_VIOLATION: <command-pattern> in primary checkout (use a worktree or invoke via manager-git)"`.

**Acceptance**: M2 is complete when a synthetic PreToolUse event with a
`git switch` command in the primary checkout returns deny, and the same event
with `MOAI_BRANCH_GUARD_EXEMPT=1` set returns allow.

### M3 — Rule v1.1.0 + Template Mirror (sanitized-pair)

**Files**:

- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` — bump `Version:
  1.0.0` → `1.1.0`; add a new section "## Mechanical Enforcement (v1.1.0)"
  documenting the PreToolUse hook as the enforcer. Cross-reference
  SPEC-WORKTREE-BRANCH-GUARD-001 (source retains SPEC-ID + REQ tokens for
  traceability per REQ-WBG-006).
- `internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md`
  — §25-sanitized copy: the SPEC-ID line becomes generic prose
  (`Origin: the run-phase SPEC that landed the v1.1.0 mechanical enforcer.`)
  and the SPEC-ID cross-reference bullet is dropped. NO internal SPEC IDs, NO
  REQ tokens, NO commit SHAs (CLAUDE.local.md §25 neutrality).
- `internal/template/sanitized_pair_parity_test.go` — enroll
  `.claude/rules/moai/workflow/main-checkout-branch-guard.md` in
  `sanitizedPairPaths` (line 91). The rule is intentionally NOT added to
  `workflowOptMirroredPaths` in `rule_template_mirror_test.go` — a NOTE
  comment in that file records the intentional exclusion.

**Wording guidance**: the v1.1.0 section names the hook handler
(`internal/hook/pre_tool.go` `checkBranchState`), the sentinel deny prefix
(`BRANCH_GUARD_VIOLATION`), the exemption mechanism (`HookInput.AgentType ==
"manager-git"` OR `MOAI_BRANCH_GUARD_EXEMPT=1`), and the discriminant
(`--path-format=absolute` git 2.31+, with fallback).

**Acceptance**: M3 is complete when
`go test ./internal/template/... -run TestSanitizedPairParity` is green,
`go test ./internal/template/... -run TestTemplateNoInternalContentLeak` is
green, and `grep -c "Version: 1.1.0"` returns 1 for both files. The source
MUST carry `SPEC-WORKTREE-BRANCH-GUARD-001` (>= 1 occurrence); the mirror
MUST carry 0 occurrences.

### M4 — CLI Worktree Advisory (user-facing UX)

**Files**:

- `internal/cli/init.go` — emit advisory on first-run init: "Tip: this checkout
  is shared across concurrent sessions. For branch-changing work (switch,
  reset, rebase), use a worktree: `moai cc -w` / `moai cg -w`. See
  `.claude/rules/moai/workflow/main-checkout-branch-guard.md`."
- `internal/cli/update.go` — emit the same advisory (one-line form) when
  `update` completes.
- `internal/cli/web.go` — emit the advisory when `web` starts (web work should
  happen in a worktree).
- The advisory consumes `workflow.worktree.auto_create` (`workflow.yaml:117`):
  when `auto_create` is `true`, the advisory becomes "auto-creating a worktree
  for you..."; when `false` (current default), it stays a recommendation.

**Acceptance**: M4 is complete when each CLI command emits the advisory on
stdout (or stderr for `web`) at least once during a smoke run, and the advisory
text references the worktree flag (`-w` / `--worktree`).

### M5 — `.worktreeinclude` Reconciliation (mechanical, lowest design-risk)

**Files**:

- `.worktreeinclude` (repo-root) — replace contents with the template SSOT,
  removing the local `teammateMode` mention.
- `internal/template/templates/.worktreeinclude` — unchanged (SSOT).
- Commit both in the same sync-phase commit.

**Acceptance**: M5 is complete when `diff .worktreeinclude
internal/template/templates/.worktreeinclude` exits 0 (empty diff) and both
files are staged in the same commit.

### M6 — Tests + Characterization + Sync

**Files**:

- `internal/hook/branch_guard_test.go` (continued) — characterization tests
  capturing the regex set's true-positive and true-negative coverage.
- `internal/hook/pre_tool_branch_guard_integration_test.go` (NEW) — end-to-end
  PreToolUse event → deny decision, including the worktree-allow and
  exempt-allow paths.
- `internal/hook/pre_tool_test.go` (EXTEND) — regression tests proving the
  quality-gate path is NOT invoked for branch-state commands (census P1-B
  structural guarantee).
- A latency test asserting the deny decision completes in ≤ 500ms (REQ-WBG-010).
- Sync-phase commit closes the SPEC (3-phase close).

**Acceptance**: M6 is complete when the full test suite passes
(`go test ./internal/hook/... -count=1`), lint is clean
(`golangci-lint run ./internal/hook/...`), and coverage on
`internal/hook/branch_guard.go` ≥ 85%.

## §G. Anti-Patterns (run-phase avoid)

- **AP-1**: encoding the exemption as a settings.json `allow` entry (would
  bypass the hook entirely and re-introduce the static-deny lockout problem in
  reverse). The exemption MUST be a hook-internal check.
- **AP-2**: parsing the agent identity from the Bash command itself (e.g.,
  looking for a sentinel comment). The agent identity comes from
  `HookInput.AgentType` or the env var — never from the command string.
- **AP-3**: invoking the quality gate for branch-state commands. The quality
  gate is structurally inaccessible (`IsGitCommit` only matches `git commit`),
  but a future refactor might widen its scope — the M6 regression test guards
  against this.
- **AP-4**: editing the local `.worktreeinclude` first and then "syncing up" to
  the template. Template-First (CLAUDE.local.md §2) requires the template to be
  the SSOT edit target.
- **AP-5**: removing the existing static deny entries at `.claude/settings.json`
  during this SPEC's run-phase. They are defense-in-depth per REQ-WBG-013; their
  reconciliation is separate debt.

## §H. Cross-References

- spec.md (this SPEC) — REQ-WBG-001 through REQ-WBG-013
- acceptance.md — AC-WBG-001 through AC-WBG-013
- `.claude/rules/moai/development/coding-standards.md` § Bash Risk-Amplifier
  Doctrine (FAIL-OPEN norm referenced by REQ-WBG-012)
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` v1.0.0 (to be
  bumped to v1.1.0 by M3)
- Census P1-B handoff memory `project_census_priority1_handoff`
- Motivating analysis `project_worktree_branch_guard_handoff`

## §H. Out of Scope

### Out of Scope — Plan-level exclusions

- The orchestrator-side spawn contract that sets
  `MOAI_BRANCH_GUARD_EXEMPT=1` when spawning `Agent(manager-git, ...)` —
  deferred to a follow-up SPEC per the orchestrator's iteration-2 decision
  (B-1 / D1). This plan covers the hook side only.
- Reconciliation of the retained static deny entries at
  `.claude/settings.json:341, 350-353, 457, 460` with `manager-git.md` Phase D
  — tracked as census P1-B residual debt per spec.md §E.
- Migration of `manager-git.md` Phase D to operate in a worktree rather than
  the primary checkout — a future SPEC may eliminate the exemption entirely.
- Sub-agent runtime metadata plumbing to populate `agent_type` in PreToolUse
  events for `Agent(manager-git, ...)` spawns — closed by the env var
  workaround (REQ-WBG-011b), not by runtime changes.

## §I. HISTORY

- 2026-07-28 v0.1.0 — initial plan (manager-spec, plan-phase). 6 milestones
  ordered by decision-reversibility. 2 NEEDS-CLARIFICATION items surfaced (B-1
  sentinel env var orchestrator-side spawn contract — non-blocking; B-3 static
  deny reconciliation — out of scope).
- 2026-07-28 v0.1.1 — plan-audit iter-2 fixes: D1 B-1 rewritten with honest
  Phase-D break framing (no longer "non-blocking"/"safe-default"; aligned with
  spec.md §A valence); D7 marker convention `[NEEDS CLARIFICATION: ...]`; D2
  M1 mock-exec.Command approach made explicit (matches AC-WBG-005).
- 2026-07-28 v0.1.2 — M3 rewritten from byte-identical copy +
  `workflowOptMirroredPaths` enrollment to §25-sanitized mirror +
  `sanitizedPairPaths` enrollment. §A research finding #3 updated with the
  sanitized-pair rationale. §D byte-parity invariant rewritten as
  sanitized-pair invariant. Implementation is already CI-green; this is a
  wording/AC correction, not a design change.
