---
id: SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001
title: "Main-Checkout Branch-State Guard — git branch Mutation-vs-Query Flag-Class Completion"
version: "0.1.0"
status: draft
created: 2026-09-03
updated: 2026-09-03
author: manager-spec
priority: High
phase: "v3.1.5"
module: internal/hook
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "hook, branch-guard, git-branch, flag-class, under-match, bug-fix"
depends_on:
  - SPEC-WORKTREE-BRANCH-GUARD-001
  - SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001
  - SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001
related_specs:
  - SPEC-WORKTREE-BRANCH-GUARD-001
  - SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001
  - SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001
---

# SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001 — git branch Mutation-vs-Query Flag-Class Completion

## §A. Problem Statement

The main-checkout branch-state guard's `git branch` pattern —
`internal/hook/branch_guard.go:120`, pattern
`\bgit\s+branch\s+(-[dDmMcC]\s+)?[^\s-]` — matches branch mutation via a
single-character flag class `[dDmMcC]` followed by the "non-flag token after
subcommand" rule `[^\s-]`. Two consequences:

1. **Under-match (the defect).** Every mutating flag form outside the
   character class passes through the guard un-denied in the primary
   checkout. **Measured** on tree `d592b0551` (plan-phase synthetic probe
   driving `matchBranchStateCommand`; probe deleted after measurement, full
   verbatim output in acceptance.md §D.0):

   | Command form | matched (pre-fix) |
   |---|---|
   | `git branch -f topic abc1234` | false |
   | `git branch --force topic abc1234` | false |
   | `git branch -df old` / `-fm renamed` / `-vD old` (combined short flags) | false |
   | `git branch -u origin/main topic` | false |
   | `git branch --set-upstream-to=origin/main topic` | false |
   | `git branch --unset-upstream topic` | false |
   | `git branch -t topic origin/main` | false |
   | `git branch --edit-description topic` | false |

   `git branch -f <branch> <sha>` is a branch-pointer rewrite against the
   shared primary checkout — exactly the hazard class the guard exists to
   deny — and it is ALLOWED through. Discovered by synthetic measurement
   during card t458 (doc-fix card, landed on develop).

2. **Over-match axis is clean in current source (measured, same probe).**
   `git branch --list develop -v`, `--merged main`, `--no-merged main`,
   `--points-at HEAD`, `--format %(refname)` all return `matched=false`
   (allowed), consistent with the existing pinned read-only test set at
   `internal/hook/branch_guard_test.go:308-320` (`-v`, `-vv`, `-a`, `-r`,
   bare, `--list`, `--show-current`, `--contains HEAD`). The historical
   friction recorded in CLAUDE.local.md §4.1 (read-only `git branch --list` /
   `-vv` denied) does NOT reproduce against this source — the source comment
   at `branch_guard.go:90-106` records a pickaxe finding that no committed
   revision ever carried an undiscriminating `git branch` pattern, and the
   guard additionally ships opt-in-disabled by default
   (`Workflow.BranchGuard.Enabled`, distributed default false). The stale
   friction note is therefore NOT evidence of a live over-match; M1's
   measurement matrix is the authoritative word on both axes.

### Value statement

t458 landed doctrine stub v1.3.2 at
`.claude/rules/moai/workflow/main-checkout-branch-guard.md`: the forbidden
table names only MUTATION forms, and a Query-vs-mutate discrimination bullet
permits read-only `git branch` queries (bare list, `--list`, `-v`/`-vv`,
`--show-current`, `--contains`/`--merged`/`--points-at`). The CODE already
implements the query half of that discrimination but not the full mutation
half. This SPEC makes the code's flag-class discrimination match the
doctrine's — closing the under-match axis without reopening the over-match
axis.

## §B. Scope

### In Scope

1. **The `git branch` branch-state pattern only** (`branch_guard.go:120`):
   widen/distinguish the flag discrimination so every MUTATION form of
   `git branch` denies in the primary checkout and every QUERY form stays
   allowed, mirroring the v1.3.2 doc discrimination.
2. **Full-flag-token classification.** Long flags classified by their
   complete token, never by substring — `--format` ≠ `--force`, and
   `--contains` / `--merged` / `--no-merged` must not be denied via the
   `c`/`m` letters embedded in their names.
3. **Synthetic measurement harness (M1).** A permanent Go table test driving
   the guard's matcher with a pinned expected matrix covering both axes; the
   fix class is decided only after this matrix exists and has been run.
4. **Condition-documenting tests (M3).** Opt-in gate (`Workflow.BranchGuard.Enabled`
   distributed default false; tests enable it explicitly) and both exemption
   axes' unreachability from tool-spawned subagents, encoded as documented
   test conditions so future readers do not misread a guard bypass.
5. **Preserved surfaces**: `d/D/m/M/c/C` + bare-name denial, deny sentinel
   prefix `BRANCH_GUARD_VIOLATION:`, primary-vs-worktree discriminant,
   fail-open semantics (deny only on positive evidence; uncertainty → allow
   + audit-log append).

### Out of Scope — Non-branch patterns

- The `git switch`, `git checkout`, `git reset --hard`, `git stash`,
  `git rebase`, `git merge` patterns in the same `branchStatePatterns` set
  are NOT touched. Their behavior stays byte-identical.
- `git checkout <file>` single-file-restore lexical collision (documented
  residual at `branch_guard.go:72-76`) is NOT re-litigated.

### Out of Scope — Guard redesign

- No new exemption surface, no discriminant change, no audit-log path change,
  no latency-budget change. The fix is the `git branch` pattern (or an
  equivalent token-level discrimination replacing that single pattern entry)
  plus tests.
- Shell-wrapped / obfuscated git invocations (`bash -c "git branch -f …"`)
  remain under-matched — the documented correct direction for a fail-open
  guard.

### Out of Scope — Doctrine bump

- `main-checkout-branch-guard.md` is already at v1.3.2 with the intended
  discrimination wording; no doctrine edit is required by this SPEC. If M1's
  measurement contradicts the v1.3.2 permitted list, a blocker report — not
  a doctrine edit — is the escalation path.

### Out of Scope — CLAUDE.local.md friction-note hygiene

- Updating or annotating the CLAUDE.local.md §4.1 historical friction note
  is not part of this card; the M1 matrix output is the artifact a future
  card would cite.

## §C. Requirements (GEARS notation)

### REQ-WBG-F-001 — Doctrine-mirroring flag discrimination (Ubiquitous)

The `git branch` branch-state guard SHALL classify command forms by a
mutation-vs-query flag discrimination that mirrors the v1.3.2 doctrine
stub: a form is denied in the primary checkout if and only if it presents a
mutating flag or a bare branch-name creation operand, and a form presenting
only query flags is allowed.

### REQ-WBG-F-002 — Mutation forms denied (Event-detected)

**When** a `git branch` command presents any mutating flag — at minimum
`-f`, `--force`, `-d`, `-D`, `-m`, `-M`, `-c`, `-C`, `-u`, `--set-upstream`,
`--set-upstream-to`, `--unset-upstream`, `-t`, `--track`, `--no-track`,
`--edit-description`, or any combined short-flag cluster containing one of
`d/D/m/M/c/C/f` — the guard SHALL deny it in the primary checkout with the
`BRANCH_GUARD_VIOLATION:` sentinel prefix. A bare branch-name creation
operand (`git branch <name>`, `git branch <name> <start-point>`) SHALL
continue to deny.

### REQ-WBG-F-003 — Query forms allowed (Event-detected)

**When** a `git branch` command presents only query flags — bare `git branch`,
`--list`, `-v`, `-vv`, `-a`, `-r`, `--show-current`, `--contains`,
`--merged`, `--no-merged`, `--points-at`, `--format`, `--sort`, `-q`,
`-i` — the guard SHALL allow it in the primary checkout.

### REQ-WBG-F-004 — Whole-token long-flag classification (Ubiquitous)

The guard SHALL classify each long flag by its complete token, never by a
substring or an embedded-letter character class, such that `--format` is
never classified as `--force`, and `--contains` / `--merged` / `--no-merged`
are never denied via the `c` / `m` letters embedded in their names.

### REQ-WBG-F-005 — Preserved surfaces (State-driven)

**While** the guard evaluates any command, the existing covered mutation
forms (`git branch <name>`, `-d`, `-D`, `-m`, `-M`, `-c`, `-C`), the deny
sentinel prefix `BRANCH_GUARD_VIOLATION:`, the primary-vs-worktree
discriminant (git-dir vs git-common-dir at the command's actual cwd), the
fail-open contract (deny only on positive evidence; any uncertainty → allow
+ audit-log append to `.moai/logs/branch-guard-audit.log`), and every
non-`branch` pattern in `branchStatePatterns` SHALL remain byte-identical in
behavior.

### REQ-WBG-F-006 — Opt-in gate unchanged (Where capability gate)

**Where** `Workflow.BranchGuard.Enabled` is false (the distributed default),
the guard SHALL perform no pattern evaluation and no deny — the opt-in gate
and its disabled-path cost profile are unchanged from
SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001.

### REQ-WBG-F-007 — Measurement-first sequencing (Ubiquitous)

The fix class SHALL be decided only after the M1 synthetic measurement
matrix (a Go test fixture driving the guard's matcher with representative
command strings and pinned expected classifications) exists and has been
executed on the pre-fix tree; no flag form's classification may rest on
estimation.

### REQ-WBG-F-008 — Exemption-axes conditions documented (Ubiquitous)

The test suite SHALL encode as documented conditions (not code changes)
that both exemption axes are unreachable from a tool-spawned subagent:
`AgentType` is populated only for a main-thread `claude --agent <name>`
launch, and `MOAI_BRANCH_GUARD_EXEMPT=1` is read from the hook process's
own environment (spawned before the guarded command runs, so exporting it
inside the command is a no-op).

## §D. Constraints

- **Tier M artifact set** (4 plan-phase files: spec / plan / acceptance /
  progress) — the dispatch labeled the card Tier S but mandated
  acceptance.md with the full expected measurement matrix; the artifact-set
  tier is therefore M with Tier-S scope discipline (matcher + tests only).
- **Class B defect card (kanban)**: cause established by synthetic
  measurement during t458; the run phase lands the fix, not a redesign.
- **verification-claim-integrity + verification-completeness**: every AC
  carries a RED-now cell pinned to tree `d592b0551` (the plan-phase probe
  output in acceptance.md §D.0) and a green path naming the milestone that
  flips it; no classification claim rests on estimation.
- **Fail-open doctrine** (Bash Risk-Amplifier WARN-ONLY, FAIL-OPEN): no new
  blocking on uncertainty paths.
- **Go code/comments in English**; table-driven tests; package
  `internal/hook` conventions (t.TempDir isolation, no OTEL env in parallel
  tests).
- **Verification scope**: affected packages only (`go test ./internal/hook/...`)
  + `golangci-lint`; no local full suite (lane-load discipline).

## §E. Acceptance Criteria (summary — full Given-When-Then + matrix in acceptance.md)

| AC | Subject | Verifiable by |
|----|---------|---------------|
| AC-WBG-F-001 | Full expected matrix: every mutation form → deny, every query form → allow (§D.1 table) | M1/M3 Go table test over the matcher + handler |
| AC-WBG-F-002 | Combined short-flag clusters containing d/D/m/M/c/C/f → deny (`-df`, `-fm`, `-vD`) | Matrix cells with RED-now on `d592b0551` |
| AC-WBG-F-003 | Query allowlist regression: `--list develop -v`, `-vv`, `--merged`, `--no-merged`, `--points-at`, `--format`, `--contains` → allow | Matrix cells (green-now; must stay green post-fix) |
| AC-WBG-F-004 | Whole-token classification: `--format` allow vs `--force` deny; `--contains`/`--merged`/`--no-merged` allow despite embedded letters | Matrix cells |
| AC-WBG-F-005 | Preserved surfaces: existing deny tests (`branch_guard_test.go:234-245`) + sentinel + discriminant + fail-open tests all stay green | Existing + M3 test run |
| AC-WBG-F-006 | Opt-in gate: disabled default → no pattern evaluation, no deny | Existing `pre_tool_branch_guard_optin_test.go` stays green |
| AC-WBG-F-007 | Fail-open preserved: non-git cwd / missing git / rev-parse failure → allow + audit entry | Existing fail-open tests stay green |
| AC-WBG-F-008 | Exemption-axes conditions encoded as documented test conditions | M3 condition tests / test-doc comments |

## §F. Cross-References

- `internal/hook/branch_guard.go:120` — the pattern under fix;
  `:59-111` the discrimination rationale comment block (t42 `-c`/`-C`
  completion + accepted combined-short-flag residual this SPEC closes);
  `:196` `matchBranchStateCommand`; `:210` exemption env read;
  `:238-246` fail-open contract; `branchGuardViolationPrefix` at `:50`.
- `internal/hook/branch_guard_test.go:234-245` (deny axis pins),
  `:308-320` (query allowlist pins), `:610-660` (sentinel + origin test).
- `internal/hook/pre_tool_branch_guard_optin_test.go:140-175` (exemption
  axes tests — AC-REQ-6a/6b of -OPTIN-001).
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` v1.3.2 —
  doctrine SSOT (§ Rules forbidden table row 2; § Permitted bullet 1;
  § Mechanical enforcement Query-vs-mutate discrimination bullet).
- `.claude/rules/moai/workflow/main-checkout-branch-guard-detail.md` —
  enforcer implementation detail (exemption reachability caveat).
- Predecessors: SPEC-WORKTREE-BRANCH-GUARD-001 (completed),
  SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 (completed),
  SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001 (completed).
- Card t458 (doctrine v1.3.2 landing card) and card t467 (this card).
- CLAUDE.local.md §4.1 — the historical over-match friction note this SPEC
  treats as possibly-stale, resolved by M1 measurement, not by trust.

## §G. HISTORY

- 2026-09-03 v0.1.0 — initial draft (manager-spec, plan-phase, card t467
  Class B). Root cause confirmed by plan-phase synthetic probe on tree
  `d592b0551`: 10 mutation forms outside `[dDmMcC]` pass the matcher
  (under-match axis); all probed query forms already allowed (over-match
  axis clean in source; CLAUDE.local.md §4.1 friction note treated as
  stale-pending-M1). Fix scoped to the single `git branch` pattern + tests;
  all other patterns and guard surfaces preserved.
