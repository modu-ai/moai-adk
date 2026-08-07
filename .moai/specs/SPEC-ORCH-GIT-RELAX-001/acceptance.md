# acceptance.md — SPEC-ORCH-GIT-RELAX-001

> Verification layer. Given-When-Then scenarios for each AC-OGR token. Binary-testable. The GEARS obligation lives in `spec.md` (REQ-OGR-* tokens); this file does NOT restate GEARS requirements.

---

## §A. AC Matrix

| AC | Verifies REQ | Milestone | Severity | Indirect verification |
|----|--------------|-----------|----------|----------------------|
| AC-OGR-001 | REQ-OGR-003, 004, 005 | M2, M4 | MUST | grep + make build |
| AC-OGR-002 | REQ-OGR-001, 002, 003, 007, 009 | M1, M2, M4 | MUST | doctrine cross-ref audit (M4.2) |
| AC-OGR-003 | REQ-OGR-010 | M4 | MUST | grep audit (M4.2) |
| AC-OGR-004 | REQ-OGR-011, 012 | M3 | MUST | Go test suite |
| AC-OGR-005 | REQ-OGR-014 | M4 | MUST | full gate (M4.4) |
| AC-OGR-006 | REQ-OGR-013 | M4 | MUST | regression test (M4.1) |
| AC-OGR-007 | REQ-OGR-008, 016 (Late-Branch limb) | M2 | MUST | byte-identity diff |
| AC-OGR-008 | REQ-OGR-015 | M4 | MUST | no-mirror audit (M4.3) |
| AC-OGR-009 | REQ-OGR-016 (catalog-retention + frontmatter-preservation limbs) | M2, M4 | MUST | catalog + frontmatter diff |
| AC-OGR-010 | REQ-OGR-006 | M4 | MUST | sync-check behavior test |

---

## §B. Severity Definitions

- **MUST (blocking)** — failure blocks sync-phase entry. All 10 ACs are MUST.
- **SHOULD** — failure emits a sync-auditor finding but does not block. None in this SPEC (the relaxation is tight enough that every criterion is binary-blocking).
- **MAY** — optional. None.

---

## §C. Acceptance Criteria (Given-When-Then)

### AC-OGR-001 — orchestrator-direct Tier S/M push+PR without spawning manager-git

**Given** the orchestrator has completed run-phase for a Tier S (< 300 LOC, < 5 files) or Tier M (300–1000 LOC, 5–15 files) change, AND the user has NOT set the `--pr` flag, AND the Pre-Spawn Sync Check returns `0 N` (local ahead, no foreign active session),
**When** the orchestrator runs `MOAI_BRANCH_GUARD_EXEMPT=1 git switch -c feat/SPEC-XXX` → `git push -u origin feat/SPEC-XXX` → `gh pr create --base main ...` directly via Bash (no `Agent(manager-git)` invocation),
**Then** the branch is created on the primary checkout, the push succeeds, the PR URL is returned, AND no `manager-git` SubagentStart event appears in the session transcript for this push+PR sequence.

---

### AC-OGR-002 — manager-git NOT invoked for Tier S/M (retained for Tier L / `--pr`) — inversion + retention principle

**Given** the doctrine at all 13 edit-surface locations (§D Table 1 rows #1–#13) reflects the relaxation (post-M1+M2),
**When** a code path classification runs (Tier S/M vs Tier L vs `--pr`),
**Then** `manager-git` is invoked ONLY for: (a) Tier L release PRs, (b) multi-step merges (team-mode `--auto-merge`), (c) Late-Branch 4-Phase closure, OR (d) any tier when the user explicitly sets `--pr`. For Tier S/M without `--pr`, the orchestrator handles git-op directly and `manager-git` is NOT spawned.

**REQ coverage** (iter-2 extension) — this AC verifies:
- REQ-OGR-001 (inversion principle) — state-sensitive Tier S/M routing to orchestrator-direct IS the inversion in action.
- REQ-OGR-002 (subject = orchestrator) — the orchestrator IS the actor holding live HEAD + worktree registry for Tier S/M.
- REQ-OGR-003 (Tier S/M default orchestrator-direct while no `--pr`).
- REQ-OGR-007 (Tier L / `--pr` retention) — `manager-git` REMAINS the canonical owner for the complex-flow class.
- REQ-OGR-009 (Tier S/M anti-pattern) — spawning `manager-git` in that context is the retired anti-pattern.

**Indirect verification**: `grep -rn "manager-git" <doctrine files>` returns manager-git mentions ONLY in the Tier L / `--pr` / Late-Branch contexts.

---

### AC-OGR-003 — all 12 doctrine locations updated consistently

**Given** M1 and M2 have edited all 13 change-surface locations (§D Table 1 rows #1–#13) per spec.md §D Table 1,
**When** the doctrine cross-reference audit runs (M4.2 grep):
```bash
grep -rn "all tiers\|모든 tier.*PR.*manager-git\|push + PR is always delegated\|manager-git.*creates a feature branch and opens a PR per phase" \
  CLAUDE.md .claude/agents/moai/manager-git.md \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/workflow/main-checkout-branch-guard.md \
  .claude/rules/moai/workflow/spec-workflow.md \
  .moai/config/sections/delegation.yaml \
  .moai/docs/git-local-workflow-doctrine.md \
  internal/template/templates/CLAUDE.md \
  internal/template/templates/.claude/agents/moai/manager-git.md \
  internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md \
  | grep -v "Tier L\|--pr\|carve-out\|RELAX-001\|RETIRED"
```
**Then** the command returns 0 matches (no residual "all tiers → manager-git" statements outside the legitimate Tier L / `--pr` carve-out).

---

### AC-OGR-004 — branch-guard exemption path admits orchestrator-direct Tier S/M (Go unchanged or minimal)

**Given** `internal/hook/branch_guard.go` `isExemptAgent` reads `MOAI_BRANCH_GUARD_EXEMPT=1` before the `AgentType == "manager-git"` identity check (lines 144–155),
**When** the test `go test ./internal/hook/... -run TestIsExemptAgent -count=1` runs with `MOAI_BRANCH_GUARD_EXEMPT=1` set and a `HookInput{AgentType: ""}` (simulating orchestrator-direct, no agent identity),
**Then** `isExemptAgent` returns `true` via the env branch, AND the test passes, AND the git diff for `internal/hook/branch_guard.go` is EMPTY (no Go change needed — REQ-OGR-012's minimal-change clause permits a minimal change only if verification fails; the happy path requires zero Go diff).

---

### AC-OGR-005 — full gate passes

**Given** all M1–M4 milestones are complete,
**When** the full gate runs:
```bash
go test ./... -count=1
go test ./internal/hook/... -count=1
go test ./internal/template/... -count=1
moai agent lint
moai workflow lint
make build
golangci-lint run
```
**Then** every command exits 0, AND `internal/template/internal_content_leak_test.go` reports 0 SPEC IDs / REQ tokens / commit SHAs leaked, AND `internal/template/split_namespace_test.go` reports 0 split-harness leaks.

---

### AC-OGR-006 — regression test proves orchestrator-direct resolves the PR-#1338 incident class

**Given** a regression test in `internal/hook/branch_guard_test.go` (or sibling) that constructs the PR-#1338 scenario shape: a multi-worktree state where (a) an isolated snapshot-holder (no live HEAD, no worktree registry — simulating the old manager-git context) would cross-swap when asked to restore primary→main, AND (b) the orchestrator-direct path (reading live HEAD + worktree registry via `git rev-parse` + `git worktree list`) restores primary→main correctly,
**When** `go test ./internal/hook/... -run TestPR1338Regression -count=1` runs,
**Then** the test passes, AND the orchestrator-direct restoration is demonstrated correct, AND the isolated-snapshot path is demonstrated incorrect (the cross-swap failure mode is mechanically reproduced and then resolved by the full-context path).

---

### AC-OGR-007 — Late-Branch 4-Phase body remains byte-identical

**Given** M2 edits `.claude/agents/moai/manager-git.md` frontmatter description (invocation gate) but MUST NOT touch the `### Late-Branch Invocation Pattern` body (Phases A–D),
**When** the byte-identity check runs:
```bash
diff <(sed -n '/### Late-Branch Invocation Pattern/,/^## Mode-Specific/p' .claude/agents/moai/manager-git.md) \
     <(sed -n '/### Late-Branch Invocation Pattern/,/^## Mode-Specific/p' internal/template/templates/.claude/agents/moai/manager-git.md)
```
**Then** the diff is EMPTY (the Late-Branch body is byte-identical between the local file and its template mirror, AND both match the pre-M2 baseline — verified by `git diff HEAD~ -- .claude/agents/moai/manager-git.md` showing changes ONLY in the frontmatter `description:` block, not in the Late-Branch body).

---

### AC-OGR-008 — LOCAL-ONLY files carry no template mirror

**Given** the relaxation edits CLAUDE.local.md, `.moai/docs/git-local-workflow-doctrine.md`, and `.moai/docs/repo-local-pr-policy.md` (the 3 LOCAL-ONLY doctrine files),
**When** the no-mirror audit runs:
```bash
for f in CLAUDE.local.md .moai/docs/git-local-workflow-doctrine.md .moai/docs/repo-local-pr-policy.md; do
  test ! -e "internal/template/templates/$f" && echo "NO-MIRROR OK: $f" || echo "NO-MIRROR VIOLATED: $f"
done
```
**Then** the command prints `NO-MIRROR OK:` for all 3 files (no template counterpart exists or is created by this SPEC).

---

### AC-OGR-009 — manager-git catalog retention + frontmatter preservation (REQ-OGR-016 limbs)

**Given** the relaxation is fully implemented (post-M1+M2+M4), and `manager-git` is retargeted to Tier L / `--pr` only,
**When** the catalog-retention + frontmatter-preservation check runs:
```bash
# (a) manager-git remains in the retained agent catalog (CLAUDE.md §4 table)
grep -c "manager-git" CLAUDE.md   # MUST be ≥ 1 (still listed)
# (b) frontmatter fields unchanged except the description (invocation-gate edit from M2)
diff <(sed -n '/^---$/,/^---$/p' .claude/agents/moai/manager-git.md) \
     <(sed -n '/^---$/,/^---$/p' internal/template/templates/.claude/agents/moai/manager-git.md)
# (c) Late-Branch body byte-identity already covered by AC-OGR-007
```
**Then** (a) `grep -c "manager-git" CLAUDE.md` returns ≥ 1 (`manager-git` is NOT removed from the 11-agent catalog — REQ-OGR-016 "no removal from catalog" limb); AND (b) the frontmatter diff shows changes ONLY in the `description:` field (the M2 invocation-gate edit), with `model:`, `effort:`, `permissionMode:` (or equivalent frontmatter fields) UNCHANGED (REQ-OGR-016 "no frontmatter change" limb). The Late-Branch body limb of REQ-OGR-016 is covered by AC-OGR-007 byte-identity.

---

### AC-OGR-010 — foreign-session auto-isolation on direct push+PR (REQ-OGR-006)

**Given** the orchestrator is about to perform a direct Tier S/M push+PR, AND the Pre-Spawn/Pre-Edit Sync Check `moai session list --json` query returns ≥ 1 foreign active session on the same checkout + SPEC scope,
**When** the orchestrator evaluates the sync-check result,
**Then** the orchestrator HALTS the direct push+PR on the shared tree, AND either (a) auto-isolates into a worktree per `.claude/rules/moai/workflow/worktree-integration.md` § Parallel-Session Branch Conflict Auto-Isolation, OR (b) surfaces an `AskUserQuestion` round (isolate / wait / abort) — the direct push+PR does NOT proceed on the shared tree under contention. The behavior is verified via the regression test (AC-OGR-006 sibling) OR a doctrine grep confirming the halt+route wiring is present in `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn/Pre-Edit Sync Check.

---

## §D. Traceability

- **Full coverage** — every REQ-OGR-001 through REQ-OGR-016 maps to ≥1 AC (iter-2 closed the prior 5-gap deficit; the §A matrix alone mapped only 11/16):

| REQ | Mapped to AC(s) |
|-----|------------------|
| REQ-OGR-001 (inversion principle) | AC-OGR-002 |
| REQ-OGR-002 (subject = orchestrator) | AC-OGR-001, AC-OGR-002 |
| REQ-OGR-003 (Tier S/M orchestrator-direct) | AC-OGR-001, AC-OGR-002 |
| REQ-OGR-004 (env exemption set) | AC-OGR-001 |
| REQ-OGR-005 (Pre-Spawn/Pre-Edit Sync Check) | AC-OGR-001, AC-OGR-010 |
| REQ-OGR-006 (foreign-session auto-isolate) | AC-OGR-010 |
| REQ-OGR-007 (Tier L / `--pr` retention) | AC-OGR-002, AC-OGR-007 |
| REQ-OGR-008 (Late-Branch canonical owner) | AC-OGR-007 |
| REQ-OGR-009 (Tier S/M anti-pattern retired) | AC-OGR-002 |
| REQ-OGR-010 (13-location consistency) | AC-OGR-003 |
| REQ-OGR-011 (env path admits orchestrator-direct) | AC-OGR-004 |
| REQ-OGR-012 (minimal-change clause / no widening) | AC-OGR-004 |
| REQ-OGR-013 (regression coverage) | AC-OGR-006 |
| REQ-OGR-014 (Template-First + neutrality) | AC-OGR-005 |
| REQ-OGR-015 (LOCAL-ONLY no-mirror) | AC-OGR-008 |
| REQ-OGR-016 (no abolition / no frontmatter change / no Late-Branch rewrite) | AC-OGR-007 (Late-Branch limb), AC-OGR-009 (catalog-retention + frontmatter-preservation limbs) |

- **Indirect verification** — REQ-OGR-010 (13-location consistency) is verified indirectly via the M4.2 grep audit (AC-OGR-003). No single test covers all 13 locations directly; the grep is the mechanical proof.
- **Forward-looking** — AC-OGR-006 (regression test) is forward-looking: it constructs a test that did not exist before this SPEC. The test's existence + green status is the AC.

---

## §E. Closure Gates (Definition of Done)

A SPEC is sync-ready when ALL of the following hold:

1. **All 10 ACs are PASS** (verifiable per the Given-When-Then above, with observed command output cited — not summarized).
2. **`moai spec lint`** on this SPEC directory returns 0 errors, 0 warnings.
3. **Full gate (AC-OGR-005) green** — every command's verbatim exit code 0 is observed in the run-phase evidence.
4. **Cross-reference audit (AC-OGR-003) clean** — the grep returns 0 matches.
5. **No `[NEEDS CLARIFICATION]` markers remaining** in plan.md or research.md (resolved at Implementation Kickoff Approval per `[NEEDS CLARIFICATION]` convention).
6. **Sync-phase artifacts** — progress.md §E.2/§E.3/§E.4 populated by manager-develop/manager-docs; era classification confirms V3R6 (3-phase close).

---

## §F. Edge Cases

- **Tier-boundary SPEC (exactly 300 LOC or exactly 5 files)**: the orchestrator's classification defaults to the heavier tier (Tier M at 300 LOC / 5 files; Tier L at 1000 LOC / 15 files). The relaxation applies only to the strictly-below tier; a boundary SPEC still routes through manager-git if it's classified Tier L.
- **`--pr` flag set on a Tier S change**: manager-git is invoked regardless of tier (the user explicitly requested heavy ceremony). The relaxation does not override explicit user intent.
- **Foreign active session detected during direct push+PR**: the orchestrator halts and either auto-isolates to a worktree or surfaces an `AskUserQuestion` round (REQ-OGR-006). The direct push+PR does NOT proceed on the shared tree under contention.
- **Branch guard disabled for the project** (`Workflow.BranchGuard.Enabled = false`): the env-var exemption is dormant (no runtime effect). The orchestrator-direct Tier S/M push+PR proceeds via plain `git` (no `MOAI_BRANCH_GUARD_EXEMPT=1` needed). The exemption-path documentation in main-checkout-branch-guard.md remains correct (defense-in-depth for any downstream maintainer checkout that enables the guard).
