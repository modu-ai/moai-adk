---
id: SPEC-WORKTREE-BRANCH-GUARD-001
title: Acceptance Criteria — Main-Checkout Branch-State Guard
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

# Acceptance Criteria — SPEC-WORKTREE-BRANCH-GUARD-001

## §A.1 Acceptance Criteria Design Notes (learned hazards, codified)

The ACs below obey the following anti-hazard rules (applied lessons
`feedback_ac_token_presence_not_reachability`,
`feedback_guard_signal_proves_call_not_effect`,
`feedback_ac_command_vacuous_green_traps`):

- **A1.1 — Reachability over token-presence**: every AC that asserts "the guard
  fires" includes a falsification arm — the SAME command with the guard
  removed/disabled MUST produce a different (allow) result. A grep for the
  regex source is NOT a PASS.
- **A1.2 — Vacuous-GREEN traps**: no AC uses a regex that matches the empty
  case (e.g., `grep -E 'pattern|'` always matches). Every grep is anchored and
  uses `-c` with an expected integer, not `| head -1`.
- **A1.3 — Two-sided falsification**: every "deny" AC has a paired "allow" AC
  for the same command in a different context (worktree OR exempt agent). A
  guard that denies unconditionally fails the paired AC.
- **A1.4 — Mechanically verifiable**: every AC cites a shell command whose
  verbatim output decides PASS/FAIL. No "the system should" phrasing.
- **A1.5 — Falsifiable observation**: every deny AC's observation target (the
  deny reason string) is a unique sentinel that, when removed from the source,
  causes the AC to FAIL (proves the guard ran, not just that the regex
  compiled).

## §B. AC Matrix

### AC-WBG-001 — Conditional Deny Fires in Primary Checkout (REQ-WBG-001)

**Given** a synthetic PreToolUse event with `tool_name: "Bash"`, `tool_input:
{"command": "git switch -c feat/test"}`, in the primary checkout
**When** the branch-guard handler processes the event
**Then** the handler returns `permissionDecision: "deny"` AND the reason starts
with `BRANCH_GUARD_VIOLATION:`.

**Verification command** (run-phase):

```bash
go test ./internal/hook/... -run TestBranchGuard_DeniesGitSwitchInPrimary -count=1 -v
```

**Expected**: `--- PASS: TestBranchGuard_DeniesGitSwitchInPrimary` AND the test
log captures a reason matching `^BRANCH_GUARD_VIOLATION:`.

**Falsification arm** (same event with `MOAI_BRANCH_GUARD_EXEMPT=1` env var
set) MUST return `permissionDecision: "allow"` (otherwise the AC is vacuous —
the guard would be denying unconditionally).

### AC-WBG-002 — Worktree Allow (REQ-WBG-002)

**Given** a synthetic PreToolUse event with `tool_name: "Bash"`, `tool_input:
{"command": "git switch -c feat/test"}`, in a worktree context (git-dir ≠
git-common-dir)
**When** the handler processes the event
**Then** the handler returns `permissionDecision: "allow"`.

**Verification command**:

```bash
go test ./internal/hook/... -run TestBranchGuard_AllowsInWorktree -count=1 -v
```

**Expected**: `--- PASS: TestBranchGuard_AllowsInWorktree`. The test MUST
construct the worktree context via `t.TempDir()` + `git init` + `git worktree
add`, NOT via a mock that overrides the discriminant function (mocking the
discriminant is a vacuous pass — the AC must exercise the real `git rev-parse`
invocation).

### AC-WBG-003 — Exempt Agent Allow (REQ-WBG-003, REQ-WBG-011)

**Given** a synthetic PreToolUse event with `tool_name: "Bash"`, `tool_input:
{"command": "git reset --hard origin/main"}`, in the primary checkout
**When** the handler processes the event with EITHER `HookInput.AgentType ==
"manager-git"` OR `os.Getenv("MOAI_BRANCH_GUARD_EXEMPT") == "1"`
**Then** the handler returns `permissionDecision: "allow"`.

**Verification command** (both exemption paths):

```bash
go test ./internal/hook/... -run 'TestBranchGuard_ExemptAllow' -count=1 -v
```

**Expected**: both `TestBranchGuard_ExemptAllow/AgentType_manager-git` AND
`TestBranchGuard_ExemptAllow/EnvVar_MOAI_BRANCH_GUARD_EXEMPT` subtests PASS.

**Falsification arm**: same event with NEITHER set MUST return deny (AC-WBG-001
covers this).

### AC-WBG-004 — Discriminant Correctness (REQ-WBG-004)

**Given** the host git binary is reachable
**When** `isPrimaryCheckout(projectDir)` is invoked with the primary checkout
path
**Then** it returns `(true, nil)`; invoked with a worktree path, returns
`(false, nil)`.

**Verification command**:

```bash
go test ./internal/hook/... -run TestIsPrimaryCheckout -count=1 -v
```

**Expected**: subtests
`TestIsPrimaryCheckout/PrimaryCheckout`,
`TestIsPrimaryCheckout/Worktree`,
`TestIsPrimaryCheckout/NonGitDirectory` ALL PASS.

### AC-WBG-005 — Discriminant Portability Fallback (REQ-WBG-005)

**Given** the `isPrimaryCheckout` dispatcher's primary code path
(`git rev-parse --path-format=absolute --git-dir` and `--git-common-dir`) is
mocked via an injected `exec.Command` runner that simulates an older-git host
rejecting `--path-format=absolute` (the call exits non-zero with an "unknown
flag" error)
**When** `isPrimaryCheckout` is invoked
**Then** the dispatcher falls back to `git rev-parse --absolute-git-dir` and
cwd-normalized `git rev-parse --git-common-dir` and returns the correct
classification.

**Verification command**:

```bash
go test ./internal/hook/... -run TestIsPrimaryCheckout_Fallback -count=1 -v
```

**Expected**: PASS. The test MUST inject a mock `exec.Command` runner that
simulates the older-git error (primary `--path-format=absolute` call exits
non-zero), then verify the dispatcher in `isPrimaryCheckout` falls back to
`--absolute-git-dir` + cwd-normalized `--git-common-dir`. **Direct invocation
of the fallback function is INSUFFICIENT** — it bypasses the dispatcher and is
a vacuous pass. The mock runner is injected via a package-level
`var execCommand = exec.Command` indirection that the test swaps (matches plan
M1's mock-exec.Command approach).

### AC-WBG-006 — Rule v1.1.0 Bump (REQ-WBG-006)

**Verification commands**:

```bash
# Source rule
grep -c "^Version: 1\.1\.0$" .claude/rules/moai/workflow/main-checkout-branch-guard.md
# Expected: 1

# Mechanical-enforcement section present
grep -c "Mechanical Enforcement" .claude/rules/moai/workflow/main-checkout-branch-guard.md
# Expected: >= 1

# SPEC cross-reference present
grep -c "SPEC-WORKTREE-BRANCH-GUARD-001" .claude/rules/moai/workflow/main-checkout-branch-guard.md
# Expected: >= 1
```

**Falsification arm**: `grep -c "^Version: 1\.0\.0$" .claude/rules/moai/workflow/main-checkout-branch-guard.md`
MUST return 0 (the v1.0.0 line is gone).

### AC-WBG-007 — Template Mirror Sanitized-Pair (REQ-WBG-007)

**Verification commands**:

```bash
# sanitizedPairPaths enrollment (REQ-WBG-007 requires enrollment in the
# sanitized-pair registry, NOT in the byte-parity allowlist
# workflowOptMirroredPaths).
grep -c "main-checkout-branch-guard" internal/template/sanitized_pair_parity_test.go
# Expected: >= 1 (enrollment at line 91)

# Doctrine parity between source and sanitized mirror
go test ./internal/template/... -run TestSanitizedPairParity -count=1 -v
# Expected: PASS

# Mirror §25 cleanliness (no internal SPEC-ID / REQ tokens leak to template)
go test ./internal/template/... -run TestTemplateNoInternalContentLeak -count=1 -v
# Expected: PASS

# Intentional exclusion from the byte-parity allowlist (REQ-WBG-007 rationale)
grep -c "main-checkout-branch-guard" internal/template/rule_template_mirror_test.go
# Expected: >= 1 (a NOTE comment documents the intentional exclusion)

# Sanitization signal: the source has the SPEC-ID line, the mirror does not.
# `diff` exits non-zero — this is the EXPECTED parity-sanitization signal,
# NOT drift. Run-phase confirms both halves:
grep -c "SPEC-WORKTREE-BRANCH-GUARD-001" \
  .claude/rules/moai/workflow/main-checkout-branch-guard.md
# Expected: >= 1 (source retains SPEC-ID per REQ-WBG-006 traceability)

grep -c "SPEC-WORKTREE-BRANCH-GUARD-001" \
  internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md
# Expected: 0 (mirror is §25-sanitized)
```

**Falsification arm**: re-introducing the SPEC-ID
(`SPEC-WORKTREE-BRANCH-GUARD-001`) in the mirror MUST make
`TestTemplateNoInternalContentLeak` FAIL (proves the §25 cleanliness guard is
load-bearing, not vestigial). Run-phase verifies this by temporarily inserting
the SPEC-ID into the mirror, running the test (expects FAIL), reverting.

### AC-WBG-008 — `.worktreeinclude` Reconciliation (REQ-WBG-008)

**Verification commands**:

```bash
# Byte-identity
diff .worktreeinclude internal/template/templates/.worktreeinclude
# Expected: exit 0, no output

# teammateMode mention removed from local
grep -c "teammateMode" .worktreeinclude
# Expected: 0

# Template SSOT still has the guidance block
grep -c "Add your own project-local files" internal/template/templates/.worktreeinclude
# Expected: 1
```

**Falsification arm**: re-introducing the `teammateMode` mention in the local
copy MUST make `diff` return non-zero (proves the reconciliation actually
landed).

### AC-WBG-009 — CLI Worktree Advisory (REQ-WBG-009)

**Verification commands** (pre-flight: `moai init` has NO `--dry-run` flag
— confirmed absent from `internal/cli/init.go:69-78`; `moai update` HAS
`--dry-run` at `internal/cli/update.go:79`. The init AC uses a temp-dir
target; the update AC uses `--dry-run`):

```bash
# init advisory (init has no --dry-run; use a temp-dir target + --non-interactive)
tmpinit=$(mktemp -d /tmp/test-init-XXXXXX)
go run ./cmd/moai init --non-interactive "$tmpinit" 2>&1 | \
  grep -cE "worktree.*isolation|use a worktree|moai (cc|cg) -w|claude --worktree"
# Expected: >= 1
rm -rf "$tmpinit"

# update advisory (update has --dry-run)
go run ./cmd/moai update --dry-run 2>&1 | \
  grep -cE "worktree.*isolation|use a worktree|moai (cc|cg) -w|claude --worktree"
# Expected: >= 1

# web advisory (advisory on startup output or --help)
go run ./cmd/moai web --help 2>&1 | \
  grep -cE "worktree.*isolation|use a worktree|moai (cc|cg) -w|claude --worktree"
# Expected: >= 1 (or in the web command's startup output)
```

**Falsification arm**: removing the advisory call from any one of the three
commands MUST make its corresponding `grep -c` return 0.

### AC-WBG-010 — Census P1-B Timeout-Safety (REQ-WBG-010)

**Given** the PreToolUse timeout budget is 5 seconds (Claude Code default)
**When** the branch-guard handler processes 100 consecutive synthetic
`git switch` events
**Then** each invocation completes in ≤ 500ms wall-time (10x headroom under
the 5s budget), AND the quality-gate path (`quality.IsGitCommit`) is NEVER
invoked for any branch-state command.

**Verification commands**:

```bash
# Latency test (Test* function with internal for-loop, NOT a Benchmark)
go test ./internal/hook/... -run TestBranchGuard_Latency -count=1 -v
# Expected: PASS, output line "per-invocation: <X>ms" with X < 500
# (the test runs an internal `for i := 0; i < 100; i++` loop, NOT -benchtime)

# Quality-gate bypass regression test
go test ./internal/hook/... -run TestBranchGuard_QualityGateNotInvoked -count=1 -v
# Expected: PASS — verifies quality.IsGitCommit is never called for git switch/reset/rebase/etc.

# Deny-origin arm — proves the deny comes from checkBranchState, not checkBashCommand
go test ./internal/hook/... -run TestBranchGuard_CheckBranchStateOrigin -count=1 -v
# Expected: PASS — with the branch-state regex set in checkBranchState blanked,
# the same git switch event returns allow (deny disappeared)
```

**Falsification arm 1 (quality-gate widening)**: if a future refactor widens
`IsGitCommit` to match branch-state commands,
`TestBranchGuard_QualityGateNotInvoked` MUST FAIL (asserts the census P1-B
structural guarantee).

**Falsification arm 2 (deny-origin)**: temporarily blanking the branch-state
regex set in `checkBranchState` MUST cause the deny to disappear for the same
event — proves the deny comes from `checkBranchState` (the NEW regex path
documented in REQ-WBG-010), NOT from `checkBashCommand`'s
`DangerousBashPatterns` or the quality gate.

### AC-WBG-011 — manager-git Exemption Identity-Based (REQ-WBG-011)

**Verification commands**:

```bash
# Exemption resolver is identity-based, not phase-based
grep -cE "\bPhase[ _-]?D\b|\bphaseD\b" internal/hook/branch_guard.go
# Expected: 0 (the exemption code MUST NOT reference phases; word boundaries
# avoid false matches like "phaseData")

# Both exemption paths present
grep -c 'AgentType == "manager-git"' internal/hook/branch_guard.go
# Expected: >= 1

grep -c 'MOAI_BRANCH_GUARD_EXEMPT' internal/hook/branch_guard.go
# Expected: >= 1
```

**Rationale encoding**: the exemption resolver reads `HookInput.AgentType`
first, then `os.Getenv("MOAI_BRANCH_GUARD_EXEMPT")`. The `agent_type` field's
main-thread-only limitation (research finding) is closed by the env var path.

### AC-WBG-012 — Fail-Open with Advisory (REQ-WBG-012)

**Given** a PreToolUse event where git context cannot be determined (e.g., the
`projectDir` is not inside a git repo, OR `git rev-parse` exits non-zero)
**When** the handler processes the event
**Then** the handler returns `permissionDecision: "allow"` AND writes an
advisory to `.moai/logs/branch-guard-audit.log`.

**Verification command**:

```bash
go test ./internal/hook/... -run TestBranchGuard_FailOpenOnGitError -count=1 -v
# Expected: PASS; the test captures the advisory via a temp log file
```

**Falsification arm**: same event with the deny path unconditionally enabled
MUST return deny (proves the fail-open is conditional on the git-error path,
not unconditional).

### AC-WBG-013 — Static Deny Retained as Defense-in-Depth (REQ-WBG-013)

**Verification command**:

```bash
# Static deny entries still present (defense-in-depth retained)
grep -c 'Bash(git reset --hard:\*)' .claude/settings.json
# Expected: >= 1

grep -c 'Bash(git switch:\*)' .claude/settings.json
# Expected: >= 1
```

**Rationale**: the static deny is NOT removed by this SPEC. Its
reconciliation with `manager-git.md` Phase D is tracked in census P1-B residual
debt (out of scope per §E of spec.md).

## §C. Coverage Gates

- `internal/hook/branch_guard.go` ≥ 85% coverage
- `internal/hook/pre_tool.go` — coverage MUST NOT decrease from baseline (the
  new branch-guard branch is covered by integration tests)

## §D. Definition of Done

- All 13 ACs above show PASS with verbatim command output captured in §E.2.
- `go test ./internal/hook/... -count=1` → exit 0.
- `go test ./internal/template/... -count=1` → exit 0 (mirror test green).
- `golangci-lint run ./internal/hook/...` → exit 0.
- `diff .worktreeinclude internal/template/templates/.worktreeinclude` → exit 0.
- `go test ./internal/template/... -run TestSanitizedPairParity -count=1` →
  exit 0 (sanitized-pair doctrine parity green).
- `go test ./internal/template/... -run TestTemplateNoInternalContentLeak
  -count=1` → exit 0 (mirror §25 cleanliness green).
- Sync-phase commit closes the SPEC (3-phase close; `sync_commit_sha` populated
  in progress.md §E.4 by manager-docs).

## §E. Edge Cases (forward-looking)

- **E-1**: a Bash command that pipes through `git` (e.g., `echo foo | git
  switch ...`) — the regex must still match.
- **E-2**: a Bash command that uses `git -C <path> switch ...` where `<path>`
  is a worktree — the discriminant MUST honor the `-C` path if determinable;
  otherwise fail-open with advisory.
- **E-3 (false-positive DENY)**: a Bash command that uses
  `cd <worktree> && git switch ...` from a primary-checkout session — the
  discriminant runs against the hook's `projectDir` (resolved from
  `CLAUDE_PROJECT_DIR`, i.e., the primary checkout), NOT the Bash command's
  `cd` target. All three REQ-WBG-001 conditions hold (primary checkout
  confirmed, branch-state regex matched, agent not exempt) → **DENY fires**.
  This is a false-positive deny of legitimate worktree work (the operator
  intended to operate on the worktree but invoked from the primary session).
  Mitigation: run the command from inside the worktree session itself (where
  `CLAUDE_PROJECT_DIR` points at the worktree), NOT via a `cd` prefix from the
  primary checkout. Parsing Bash-command `cd` targets to recover the intended
  cwd is out of scope (the hook does not parse Bash semantics).
- **E-4**: concurrent PreToolUse events from parallel sessions — the handler
  is stateless per invocation; no shared mutable state.
- **E-5**: a `git checkout <file>` (single-file restore, NOT branch-state) —
  the regex MUST NOT match (otherwise the hook denies legitimate file restores).
  Covered by the regex's "non-flag token after subcommand" rule.

## §F. Forward-Looking Checks (not blocking DoD)

- **F-1 (interim Phase-D break — pending follow-up SPEC)**: until a follow-up
  SPEC lands the orchestrator-side spawn contract (setting
  `MOAI_BRANCH_GUARD_EXEMPT=1` when spawning `Agent(manager-git, ...)`),
  manager-git Phase D invocations via sub-agent spawn WILL BE DENIED by this
  hook. Local-dev interim impact: Phase D steps (`git checkout main && git
  reset --hard origin/main`, per `manager-git.md:117-123`) must run via
  main-thread `claude --agent manager-git` (where `HookInput.AgentType ==
  "manager-git"` populates correctly), NOT via sub-agent spawn, until the
  follow-up SPEC. User-project impact: the retained static deny at
  `.claude/settings.json:341, 350-353, 457, 460` (REQ-WBG-013) ALSO blocks
  Phase D and is NOT addressed by this SPEC's exemption — tracked as census
  P1-B residual debt per spec.md §E. This is the same functional consequence
  and same valence as the static-deny break documented in spec.md §A: both
  are an interim break pending their respective fixes, NOT a "safe default".
  A follow-up SPEC (TBD ID) will address the orchestrator-side spawn contract
  for per-spawn env-var injection.
- **F-2**: a future SPEC may migrate `manager-git.md` Phase D to a worktree,
  eliminating the exemption entirely. The hook would then deny unconditionally
  in the primary checkout.

## §F. Out of Scope

### Out of Scope — AC-level exclusions

- ACs do NOT verify the orchestrator-side spawn contract for
  `MOAI_BRANCH_GUARD_EXEMPT=1` env-var injection (deferred to a follow-up SPEC
  per D1). The ACs verify the hook's read of the env var (AC-WBG-003), not the
  orchestrator's write.
- ACs do NOT verify reconciliation of the retained static deny entries at
  `.claude/settings.json:341, 350-353, 457, 460` with `manager-git.md` Phase
  D (census P1-B residual debt). AC-WBG-013 verifies the static deny is
  retained as defense-in-depth but does NOT verify its Phase-D compatibility.
- ACs do NOT verify the `settings.json` `env` block threat model (N2) — a
  future SPEC may add a denylist guard if threat modeling warrants.

## §G. HISTORY

- 2026-07-28 v0.1.0 — initial ACs (manager-spec, plan-phase). 13 ACs, each with
  verification command + falsification arm. §A.1 codifies the 5 anti-hazard
  rules applied to every AC.
- 2026-07-28 v0.1.1 — plan-audit iter-2 fixes: D1 F-1 honest Phase-D break
  framing; D2 AC-WBG-005 mock-exec.Command dispatcher injection; D3 deny-origin
  AC arm added to AC-WBG-010 (checkBranchState provenance); D4 E-3 corrected
  from fail-open to false-positive DENY; D5 AC-WBG-009 init uses temp-dir
  (no --dry-run flag, pre-flight verified); D6 AC-WBG-010 self-consistent
  Test* + internal loop (no -benchtime); N1 AC-WBG-011 word-boundary regex.
- 2026-07-28 v0.1.2 — AC-WBG-007 rewritten from byte-parity (`diff` exit 0 +
  `TestRuleTemplateMirror`) to sanitized-pair (`sanitizedPairPaths` enrollment
  + `TestSanitizedPairParity` + `TestTemplateNoInternalContentLeak` + source
  SPEC-ID present / mirror SPEC-ID absent). Rationale: §25 forbids a SPEC-ID
  in the template mirror, so byte-parity is impossible when the source
  references its origin SPEC (verified via `TestTemplateNoInternalContentLeak`).
