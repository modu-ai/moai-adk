# plan.md — SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001

> Implementation plan. Milestones ordered by decision-reversibility
> (highest-change-likelihood first). Run-phase manager-develop MAY choose
> between the two implementation seams identified in §C; both satisfy the
> REQ set.

## §A. Context

### §A.1 Problem (one paragraph)

`checkBranchState(input, h.projectRoot())` at `internal/hook/pre_tool.go:487`
passes `$CLAUDE_PROJECT_DIR`-resolved primary checkout path into
`isPrimaryCheckout`. When the agent operates inside a Claude-native worktree
(`.claude/worktrees/<name>/`), the agent's git commands run with cwd = worktree,
but the discriminant queries the primary checkout's git context and therefore
classifies the context as primary, denying legitimate `git rebase` /
`git push --force-with-lease` / `git switch` inside the worktree. The fix:
query the git context at the actual command cwd (`input.CWD`), and keep the
audit-log placement pinned to the primary checkout for central logging.

### §A.2 Scope summary

- Bug locus: `internal/hook/pre_tool.go:487` and `internal/hook/branch_guard.go`
  (`checkBranchState` L231, `isPrimaryCheckout` L167, `appendBranchGuardAdvisory`
  L276).
- Doctrine: `.claude/rules/moai/workflow/main-checkout-branch-guard.md`
  v1.1.0 → v1.2.0.
- Template mirror: `internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md`
  (sanitized-pair parity; the rule is already enrolled in `sanitizedPairPaths`
  per -001 REQ-WBG-007 — no enrollment change, only content parity).
- Tests: `internal/hook/branch_guard_test.go` (+ possibly a new
  `branch_guard_worktree_test.go` if the worktree-fixture setup is bulky; the
  in-place extension is preferred).

### §A.3 Predecessors

- SPEC-WORKTREE-BRANCH-GUARD-001 (completed) — landed the guard.
- SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 (completed) — added the default-off
  config gate and read-only pattern refinement.

### §A.4 PRESERVE targets (must not regress)

- `MOAI_BRANCH_GUARD_EXEMPT=1` semantics and the `manager-git` identity check
  (byte-identical; -OPTIN-001 REQ-6).
- `branchStatePatterns` regex set (the -OPTIN-001 read-only refinement is
  untouched).
- The `Workflow.BranchGuard.Enabled` default-off gate (the fix applies ONLY on
  the enabled path).
- Fail-open contract (-001 REQ-WBG-012): uncertainty at the command cwd MUST
  still emit `permissionDecision: "allow"` + audit advisory.
- Latency ceiling (≤ 500ms measured wall-time per invocation; -001 REQ-WBG-010).
- The B7 priority order for handlers OTHER than the branch guard
  (`resolveProjectRootFromEnv` stays as-is for post_tool_metrics, observability,
  session-start env injection, etc.).

### §A.5 EXTEND targets

- `checkBranchState` signature (the command-cwd argument needs to enter
  somehow — see §C seam decision).
- The doctrine rule's v1.1.0 prose (bump to v1.2.0 documenting the
  discriminant directory correction).
- The sanitized-pair mirror's content (kept in parity with the source rule).

## §B. Known Issues (B1-B12 filter)

- **B1 Cross-platform build tags**: not applicable — no `syscall` use; the fix
  runs `git` via `exec.Command` which is already cross-platform.
- **B2 Cross-SPEC policy conflict pre-scan**: predecessors -001 and -OPTIN-001
  are both `completed`; this SPEC is the explicit follow-up. No superseded SPEC
  conflict.
- **B3 C-HRA-008 subagent boundary**: the hook package already enforces the
  no-`AskUserQuestion` static guard; this fix adds no new user-interaction
  surface. Verify with the existing grep:
  `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go`
  MUST return 0 matches (unchanged from baseline).
- **B4 Frontmatter canonical schema**: this SPEC's own frontmatter uses the 12
  canonical fields; not a constraint on the run-phase.
- **B5 CI 3-tier awareness**: spec-lint + golangci-lint + Go test may each fail
  independently; run all three at §E self-verification.
- **B6 spec-lint heading convention**: this plan.md is itself a SPEC artifact;
  no `## Out of Scope` H2-only heading is introduced (the exclusions live in
  spec.md §B with proper `### Out of Scope — <topic>` H3 sub-headings).
- **B7 observer.go / capture path resolution**: this is the CORE of the bug.
  The branch guard MUST use the actual command cwd (`input.CWD`), NOT
  `$CLAUDE_PROJECT_DIR`. See §C for the seam decision.
- **B8 Working tree hygiene**: do not modify runtime-managed files
  (`.moai/harness/usage-log.jsonl`, `.moai/state/`); the audit log at
  `.moai/logs/branch-guard-audit.log` is the only log this SPEC touches, and
  only by appending fail-open entries during tests (tests use `t.TempDir()`).
- **B9 Git commit + push**: repo-local PR-mandatory policy overrides Route A —
  ALL work lands via PR (Route B). manager-develop commits on a feature branch
  and the orchestrator opens the PR. Conventional Commits format required.
- **B10 Untouched paths PRESERVE**: do NOT touch other hook handlers' path
  resolution (post_tool_metrics, observability, session-start env injection).
  Only the branch-guard call site and `checkBranchState` are in scope.
- **B11 AskUserQuestion prohibited**: on finding a blocker, return a structured
  blocker report; never call `AskUserQuestion` from manager-develop.
- **B12 Sync-phase CHANGELOG emission**: not applicable at run-phase; deferred
  to manager-docs at sync.

## §C. Pre-flight (before any code change)

```bash
# 1. Branch + baseline
git branch --show-current
git rev-parse HEAD

# 2. Build feasibility (both OS)
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. Existing lint baseline (distinguish NEW vs pre-existing)
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. Existing test baseline for branch_guard
go test ./internal/hook/... -run 'BranchGuard|IsPrimaryCheckout|CheckBranchState' -count=1 -v 2>&1 | tail -20

# 5. Confirm the call site
grep -n 'checkBranchState\|h.projectRoot' internal/hook/pre_tool.go | head

# 6. Confirm the sanitized-pair enrollment is intact (no enrollment change needed)
grep -n 'main-checkout-branch-guard' internal/template/sanitized_pair_parity_test.go
```

### Seam decision (manager-develop to resolve at M1)

Two implementation seams both satisfy the REQ set. Pick ONE; do NOT mix.

- **Seam A — internal resolution** (preferred for minimal call-site churn):
  change `checkBranchState`'s signature to `checkBranchState(input *HookInput,
  auditLogProjectDir string)`. Inside `checkBranchState`, resolve the git-context
  cwd via `input.CWD` (falling back through the existing chain). The caller at
  `pre_tool.go:487` continues to pass `h.projectRoot()` for the audit-log
  directory only.
- **Seam B — caller-side split**: the caller resolves the git-context cwd and
  passes both `(gitContextCwd, auditLogProjectDir)` into a signature-changed
  `checkBranchState(input, gitContextCwd, auditLogProjectDir)`.

Seam A is recommended (the resolution helper `resolveProjectRootFromInputOrEnv`
at `path_resolve.go:88` already exists and prefers `input.CWD`); Seam B is
acceptable if the implementer prefers explicit data flow at the call site.

**Anti-pattern**: do NOT change `isPrimaryCheckout` to read `$CLAUDE_PROJECT_DIR`
directly — it must remain a pure function of its argument. The fix is in which
directory the caller asks it to query, not in `isPrimaryCheckout` itself.

## §D. Constraints

- DO NOT remove or rewrite `resolveProjectRootFromEnv` — it is the correct
  helper for handlers that want the project root (B7 priority order preserved).
- DO NOT change `isPrimaryCheckout`'s contract — it takes a directory and
  returns `(bool, error)`. Only the caller's choice of which directory to pass
  changes.
- DO NOT introduce a new git subprocess. The single
  `git -C <cwd> rev-parse --path-format=absolute --git-dir --git-common-dir`
  (plus the existing `--absolute-git-dir` / `--git-common-dir` fallback when the
  primary path errors) is the budget; REQ-WBG-D-008 forbids additional spawns.
- DO NOT rotate the audit log to the worktree. REQ-WBG-D-004 binds this.
- DO NOT use `--no-verify` on commits.
- Conventional Commits format required; end the commit body with the `🗿 MoAI`
  trailer per repo convention.

## §E. Self-Verification (run-phase deliverable — E1..E7)

Follow the 5-section Evidence-Bearing Report format (Claim / Evidence /
Baseline-attribution / Gaps / Residual-risk) per
`.claude/rules/moai/core/verification-claim-integrity.md` §3, AND the
attribution triple (a/b/c) per `manager-develop-prompt-template.md` §E.

**E1. AC binary PASS/FAIL matrix** — every AC in `acceptance.md` MUST be
exercised by a named test command and its verbatim output recorded.

**E2. Cross-platform build**: `go build ./...` AND
`GOOS=windows GOARCH=amd64 go build ./...` MUST both exit 0.

**E3. Coverage**: `go test -cover ./internal/hook/...` for the affected files
MUST stay ≥ 85% (the package already exceeds this; the fix must not regress).

**E4. Subagent boundary grep**:
`grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go`
MUST return 0 matches (unchanged baseline).

**E5. Lint status**: `golangci-lint run --timeout=2m` MUST be clean for the
changed files (pre-existing findings in unrelated files are out of scope).

**E6. Branch HEAD + push state**: list new commit SHAs and the PR URL.

**E7. Blocker report**: if the seam decision (§C) surfaces a third option, OR a
test reveals that `input.CWD` is NOT reliably populated by Claude Code for
PreToolUse Bash events on a worktree cwd, STOP and return a blocker report.
Do NOT silently fall back to `$CLAUDE_PROJECT_DIR` — that re-introduces the
bug. The orchestrator will re-delegate.

## §F. Milestones

### M1 — Discriminant directory correction + regression tests

Decision-reversibility: HIGH (the seam choice affects signature shape). Lead
with this milestone so human review can redirect the seam before tests and
doctrine are written against the chosen shape.

- Decide seam A vs. seam B (§C) and implement.
- The git-context query MUST use `input.CWD` as the authoritative source.
- The audit-log project dir MUST continue to resolve via `$CLAUDE_PROJECT_DIR`
  → `os.Getwd()` fallback.
- Add the worktree fixture unit test (AC-WBG-D-001). The test MUST create a
  real git worktree via `t.TempDir()` + `git init` + `git worktree add`, then
  assert `isPrimaryCheckout(worktreeCwd)` returns `(false, nil)`. Avoid mocking
  where real git is available; mock only on OSes where `git` is absent (the
  existing test suite already handles this).
- Add the regression unit test (AC-WBG-D-003): `isPrimaryCheckout(primaryCwd)`
  still returns `(true, nil)`.
- Add the end-to-end PreToolUse test (AC-WBG-D-002): `input.CWD = <worktree>`
  and command `git rebase` → `permissionDecision: "allow"`.
- Add the audit-log placement test (AC-WBG-D-004): when command cwd is a
  worktree but rev-parse fails (simulated), the advisory MUST land at
  `<primary>/.moai/logs/branch-guard-audit.log`, not at
  `<worktree>/.moai/logs/`.
- Add the fail-open test (AC-WBG-D-005): `input.CWD = t.TempDir()` (non-git)
  → `permissionDecision: "allow"` + audit entry on primary.
- Confirm exemption test (AC-WBG-D-006) still passes unchanged.
- Run the latency test (AC-WBG-D-007): measured wall-time ≤ 500ms per
  invocation; record the number in §E.

**Exit gate**: all M1 tests green; seam decision recorded in the M1 commit
message body.

### M2 — Doctrine bump v1.1.0 → v1.2.0 + sanitized-pair mirror parity

Decision-reversibility: LOW (documentation only). Defers to the end so the
doctrine prose can be finalized against the implemented seam.

- Bump `.claude/rules/moai/workflow/main-checkout-branch-guard.md` to
  `Version: 1.2.0`.
- Add a short subsection documenting the discriminant directory correction:
  the discriminant queries `input.CWD` (the command cwd), NOT
  `$CLAUDE_PROJECT_DIR`. Cite REQ-WBG-D-001 and this SPEC ID.
- Mirror the prose to
  `internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md`
  (the sanitized-pair — SPEC-ID stripped per -001 REQ-WBG-007).
- Verify `TestSanitizedPairParity` passes (the rule path is already enrolled;
  no enrollment change).
- Verify `TestTemplateNoInternalContentLeak` passes for the mirror (the mirror
  MUST remain §25-clean — no SPEC-ID, no REQ tokens).

**Exit gate**: parity test green; doctrine version bump observable via
`grep '^Version: 1.2.0'` on both source and mirror.

## §G. Anti-Patterns (named)

- **AP-D-001 — Mutate `isPrimaryCheckout` to read env vars**. The function is a
  pure directory query; mixing env resolution into it destroys testability.
  Resolve cwd at the caller, pass a plain string into `isPrimaryCheckout`.
- **AP-D-002 — Rotate the audit log to the worktree**. REQ-WBG-D-004 binds
  central logging on the primary checkout. A worktree cwd that points at a
  disposed worktree would lose the audit trail.
- **AP-D-003 — Silent `$CLAUDE_PROJECT_DIR` fallback in the discriminant**.
  If `input.CWD` is absent, the fallback is permitted for the git-context query,
  but the choice MUST be observable in the audit advisory (the fail-open path
  already logs `command` + `projectDir`; ensure the resolved cwd is also
  recorded). Silent fallback re-introduces the bug when Claude Code omits cwd.
- **AP-D-004 — Touch other handlers' path resolution**. The B7 priority order
  is correct for post_tool_metrics, observability, and session-start env
  injection. Only the branch-guard call site changes.
- **AP-D-005 — Add a second `git rev-parse` subprocess**. The latency ceiling
  (REQ-WBG-D-008) forbids it; the single combined `--git-dir --git-common-dir`
  call remains.
- **AP-D-006 — Skip the worktree-fixture test and rely on mocked
  `execCommand`**. The existing `execCommand` indirection is fine for the
  fallback path (AC-WBG-005 of -001), but the headline worktree classification
  (AC-WBG-D-001) MUST be exercised against a real git worktree to avoid the
  class of "test passes against mock, fails in production" regression.

## §H. Cross-References

- spec.md (this SPEC) §A..§G — WHAT and WHY.
- acceptance.md (this SPEC) — Given-When-Then AC matrix.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` §E —
  self-verification attribution triple.
- `.claude/rules/moai/workflow/repo-local-pr-policy.md` — repo-local
  PR-mandatory override (Route B for all tiers).
- Predecessors: SPEC-WORKTREE-BRANCH-GUARD-001, SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001.
