# Progress — SPEC-SESSION-WORKTREE-001

> **Tier L** lifecycle progress. §E section skeleton emitted at plan-phase; §E.2-§E.4 populated only by the respective phase-owning agents.

## §E.1 Plan-phase Audit-Ready Signal

_Populated at plan-phase close._

- plan_status: _pending iter-2 audit (v0.2.1 iter-2 audit-fix; iter-1 verdict FAIL 0.69)_
- plan_complete_at: _pending_
- artifacts: spec.md v0.2.1, plan.md v0.2.1, acceptance.md v0.2.1, design.md v0.2.1, research.md v0.2.1, progress.md (this file)
- tier: L (retained at v0.2.1 — REQ/AC budget 24/24, file surface ≥10; v0.2.0 escalation M→L stands)
- version_trace: v0.1.0 (initial Tier M) → v0.2.0 (tier escalated M→L, 3 additions) → v0.2.1 (iter-2 audit-fix: D1 gitconfig source, D2 Q1-Q4 resolved, D3 defaultBranchDetectorFunc documented, D4 on-touch trigger, D5 path-distinguishing notices, D6 REQ-SW-014 Event-driven relabel, D7 REQ-SW-021 SHALL)
- lint: _pending `moai spec lint`_
- self_check: SPEC ID regex PASS observed; frontmatter 12-field check pending; GEARS notation verified; Out-of-Scope rule satisfied (11 `### Out of Scope —` H3 sub-headings at v0.2.1 — 9 carried from v0.2.0 + 2 added: Profile-vs-profile git identity + Hook isolation); Tier L artifact set complete (spec + plan + acceptance + design + research + progress); all v0.2.0 `[NEEDS CLARIFICATION]` markers resolved (no more blocker rounds).

## §E.2 Run-phase Evidence

### M2 — `moai init` auto-entry (COMPLETE)

**Deliverables:**
- `internal/cli/session_worktree.go` (NEW) — `enterSessionWorktree` wrapper + `materializeSessionWorktree` + function-variable seams (`sessionWorktreeGitWorktreeAdd`, `sessionWorktreeInGitWorktree`, `sessionWorktreeResolveSessionShort`, `sessionWorktreeGitCommonDir`, `sessionWorktreeGitConfigSet`) + `loadSessionWorktreeConfig`.
- `internal/cli/session_worktree_test.go` (NEW) — 9 tests covering default-off, env-forced-on, already-in-worktree skip, fail-back (2 cases), defaultBranch application, branch-name shape, random fallback, falsification.
- `internal/cli/init.go` (EDIT) — `runInit` calls `enterSessionWorktree(..., "init", ...)` at the very top (after printer setup, before any mutation); on success `os.Chdir(wtPath)` re-resolves cwd into the worktree (BI-2 / R1).

**BI-1 resolution (worktree.Add extraction vs shell-out):** SHELL-OUT. Evidence: `grep -rn "func Add\|func.*Add(" internal/cli/worktree/ internal/cli/launcher.go internal/workflow/` returns only mock-interface `Add(path, branch)` methods in `_test.go` files (`subcommands_test.go:22`, `done_test.go:128`, `worktree_orchestrator_test.go:25`) — NO exported reusable helper exists. M2 shells out to `git worktree add -b <branch> <dest>` directly (`gitWorktreeAddReal`). Extraction would be premature; M3/M6 will reuse the same wrapper.

**Q2 resolution ([WT] brackets):** BRACKETS REJECTED → `WT-` prefix fallback. Evidence: `git check-ref-format --branch '[WT]-x'` exits non-zero with `fatal: '[WT]-x' is not a valid branch name`. `SessionWorktreeBranchPrefix = "WT-"` (bracket-free). REQ-SW-006's literal `[WT]` intent is preserved via the `WT-` token; the bracket rejection is the documented EC-3 fallback.

**Q1 resolution:** Same as BI-1 (shell-out decision).

**M7 call-site hook:** `materializeSessionWorktree` applies the M2 init-specific config (`init.defaultBranch=main`, REQ-SW-020) directly and carries the `// M7 (REQ-SW-018/019/021): ApplyGitConfig call site` comment marking where safe.directory / global-gitconfig identity / opt-in options will land. No non-existent function is stubbed (build passes).

**AC binary PASS/FAIL matrix (M2-relevant):**

| AC | Status | Verification | Actual Output |
|---|---|---|---|
| AC-SW-002 (default-off init byte-identical) | PASS | `TestEnterSessionWorktree_DefaultOffReturnsEmpty` + `TestEnterSessionWorktree_DefaultOffWithConfigFalse` | `enterSessionWorktree(nil,...) → ""`; no git invocation; no notice (out.Len()==0) |
| AC-SW-005 (non-git fail-back + notice) | PASS | `TestEnterSessionWorktree_MaterializeFailFallsBack` + `TestEnterSessionWorktree_FailBackFalsification` | `→ ""` + notice contains "materialization failed" + the failure reason |
| AC-SW-007 (branch name shape) | PASS | `TestSessionWorktreeBranchName_Shape` | `WT-abcdef12-init` (Q2 bracket fallback applied) |
| AC-SW-012 (already-in-worktree skip) | PASS | `TestEnterSessionWorktree_AlreadyInWorktreeSkips` | `→ ""`; notice contains "already inside a git worktree"; git add NOT invoked |
| AC-SW-020 (init.defaultBranch=main) | PASS | `TestEnterSessionWorktree_SuccessAppliesDefaultBranch` | `git config init.defaultBranch main` called on the worktree path |

**E8 RED → GREEN proof (falsification round-trip):** the fail-back notice string was mutated from "materialization failed" to "REMOVED_FAILBACK" (simulating removal of the fail-back). Both fail-back tests went RED:
```
--- FAIL: TestEnterSessionWorktree_MaterializeFailFallsBack
    expected fail-back notice, got "...REMOVED_FAILBACK..."
--- FAIL: TestEnterSessionWorktree_FailBackFalsification
    expected fail-back notice for non-git dir, got "...REMOVED_FAILBACK..."
```
After restoring the file, both tests returned GREEN (`ok ... 1.224s`). This proves the fail-back code path is load-bearing.

**Coverage (new init-wrapper decision code):**
- `enterSessionWorktree`: 100.0%
- `sessionWorktreeBranchName`: 100.0%
- `materializeSessionWorktree`: 90.9%
- `resolveSessionShortReal`: 45.5% (random-fallback branch hit; session-id-available branch is exercised in production via the side-channel file)
- `*Real` exec shells (gitWorktreeAddReal, inGitWorktreeReal, gitCommonDirReal, gitConfigSetReal, loadSessionWorktreeConfig): 0% in unit tests — these are thin `exec.Command` wrappers exercised in production; the load-bearing decision logic is covered via the seams. Exceeds the 85% target on the wrapper decision code.

**Cross-platform build:**
- `go build ./...` → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0

**Lint:** `golangci-lint run --timeout=3m ./internal/cli/` → 0 issues.

**Vet:** `go vet ./internal/cli/... ./internal/config/...` → clean.

**Subagent boundary (C-HRA-008):** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/session_worktree.go internal/cli/session_worktree_test.go` → 0 matches (no new violations).

**Scope discipline (B10):** touched only init entry (`internal/cli/init.go` runInit top), the new wrapper (`internal/cli/session_worktree.go`), and tests. Did NOT modify web.go (M3), profile.go (M6), or the M7 git-config helper body.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- tier: L
- scope (file count): 11
- domain count: 3 (config loader, CLI entrypoints init/profile/web, worktree subsystem)
- file language mix: Go (100%)
- concurrency benefit: LOW (coding-heavy, per Anthropic coding-task parallelism caveat)
- Decision: sub-agent (Mode 5) — sequential per-milestone delegation
- Justification: Tier L coding-heavy work; Anthropic caveat "most coding tasks involve fewer truly parallelizable tasks than research". Mode 5 sequential sub-agent is the safe default. M1→M8 ordered by decision-reversibility (config flag first, PR-merge cleanup last).
- Implementation Kickoff Approval: PASSED (user explicit, this session).
- plan-audit: iter-2 PASS (0.92).
