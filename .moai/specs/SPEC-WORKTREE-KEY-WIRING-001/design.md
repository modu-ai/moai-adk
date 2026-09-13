# SPEC-WORKTREE-KEY-WIRING-001 — Design

> Card t655, Tier M+. This file carries the three settled design decisions
> and the measured surface analysis. All file:line anchors were re-verified
> in this tree (lane-4 worktree, branch WT-worktree-keys-wiring, base
> `1d150a27d`) on 2026-09-12.

## §1 Decision 1 — the integration window: Option A (auto-merge acquires)

**Settled: the auto-merge path ACQUIRES the window itself — acquire → merge →
release — via the same lock API the `moai integration acquire` verb uses,
with no `--force` anywhere on the path. Option B (permit auto-merge only
inside an already-held window) is rejected.**

### §1.1 Why A

1. **The automation surfaces have no pre-existing window holder.** The
   trigger is session-exit disposal (§4). At that moment no lane holds the
   window — the automation IS the actor. Option B makes the feature a
   guaranteed no-op (refuse + report) on the only surfaces it has, which is
   the exact "unusable when automation is the point" failure the card
   warns against.
2. **The record is the serialization, not the ceremony.** The PreToolUse
   guard reads the lock record under the PRIMARY checkout's `.moai/state`
   (`integrationLockRoot()`, internal/cli/integration.go:46). A hold written
   programmatically through `kanban.AcquireIntegrationLock` is visible to the
   guard and to `moai integration status` identically to a CLI-written hold —
   serialization is preserved by the record existing, whoever wrote it.
3. **The Go API is already importable.** `kanban.AcquireIntegrationLock` /
   `ReleaseIntegrationLock` / `ReadIntegrationLock` are plain functions
   consumed by the CLI verbs today (integration.go:284, :361). No new lock
   mechanism is needed.
4. **REQ-SW-008 default-manual philosophy is honored by the toggle, not by
   a manual pre-step.** `auto_cleanup`'s precedent: the feature is
   default-OFF and does the full manual choreography when enabled. The auto
   path performing acquire → merge → release is STRONGER discipline than
   option B, which would skip acquire/release entirely.

### §1.2 Never-force is the safety load-bearing half

The manual protocol allows `--force` as a deliberate human decision
("Take the window over from a live holder (recorded, never silent)",
integration.go:334). The auto path takes it never:

- window held LIVE → skip + notice (the holder wins; concurrent manual
  integration is exactly the race the window exists to serialize);
- window held STALE → skip + notice as well. Stale holds are
  "reclaimable" for a human (`moai integration status` says so); an
  automated reclaim would turn a dead session's unreleased hold into silent
  machine displacement. The notice tells the operator what happened.

Failure direction: a busy window skips a merge the operator can do by hand
— recoverable, reported. A forced auto-merge through a live lane's window is
not recoverable.

### §1.3 Deliberate deviation from the manual ceremony: no absorb, no re-measure

The lane's manual chain is: acquire → `git merge origin/develop` (absorb,
in the card tree) → re-measure in the merge tree → merge `--no-ff` into
develop → release. The auto path performs ONLY the local `--no-ff` merge:

- absorb requires a fetch (network) and the card explicitly scopes the
  feature to "LOCAL --no-ff merge ONLY. NEVER push";
- re-measurement is lane/CI judgment, not mechanical — the auto path cannot
  do it and must not fake it.

Consequence (accepted): the auto merge may build on a local develop that is
behind origin. Mitigations: the merge notice names the resulting merge
commit so the lead's batch-push flow reads what landed; the lead batch-push
doctrine (CLAUDE.local.md §4.1, 2026-09-02) is the push authority either way.

## §2 Decision 2 — integration target: reuse `develop_branch`, no new key

**Settled: auto-merge merges into `git_strategy.<mode>.develop_branch`
resolved through `config.LoadGitFlowIntegrationConfig`
(internal/config/loader_integration_branch.go:64). The card's suggested NEW
key (`git_strategy.<mode>.integration_branch`) is NOT created.**

Why: the card's stated properties for the new key — default EMPTY = inert;
this repo's local config the only setter (`develop`); template-neutral —
are ALL already true of `develop_branch`:

| Card property | develop_branch today |
|---|---|
| default EMPTY = inert | template `git-strategy.yaml.tmpl` carries no develop_branch key (measured: grep 0 hits); every loader failure path yields "" (loader_integration_branch.go:14-16) |
| only this repo sets it (develop) | CLAUDE.local.md §2.3 restore discipline; template-neutral per §15 |
| has a reader | `LoadGitFlowIntegrationConfig` + `LoadGitFlowDevelopBranch` (card t449/t637) |
| names the integration target | it IS the branch `moai integration acquire` records (`resolveIntegrationTarget`, integration.go:121) |

The decisive argument: the window record carries `branch`. If auto-merge
targeted a second key, the window could be held for `develop` while the
auto-merge merges into `integration_branch` — the serialization record and
the serialized act would name different branches. One fact, one key. The
git-flow/manual gating comes free and is coherent: a github-flow project has
no develop branch, so auto-merge is inert there (REQ-WKW-003 notice).

Card-scope refinement recorded: scope item 2 said "New integration-target
key"; the key already exists as of t449/t637 and reusing it satisfies every
stated property. Surfaced to the lead in the plan report.

## §3 Decision 3 — `auto_create`: explicit scope (the card-sanctioned route)

**Settled: `auto_create` keeps the advisory-wording reader as its DECLARED
contract; no new behavior gate is added. The false auto-creation claim in
the `true`-branch wording is removed (REQ-WKW-011).**

Rejected alternative — make `auto_create: true` gate real creation. The only
automated creation path in the codebase is session-worktree materialization
(`enterSessionWorktree`, gated at internal/cli/session_worktree.go:158 by
`config.SessionWorktreeEnabled`). That gate already has a live key
(`workflow.session_worktree.enabled`) plus an env override
(`internal/config/session_worktree.go:31-44`). Adding `auto_create` to it
means either an OR (two keys gating one behavior — a stale `auto_create:
true` would silently re-enable what the user disabled; un-disable-able) or a
rename (breaking a live key). Both are worse than honesty: `auto_create` is
not literally unread — its defect is that its `true` wording promises
creation nothing performs. The card explicitly sanctions "or be explicitly
scoped — your design decides".

Scope of the wording fix: the `true` branch of `emitWorktreeAdvisory`
(worktree_advisory.go:34-39) must keep the AC-WBG-009 observability regex
match while stating only true things — the auto-create preference selects
this notice form and points at `workflow.session_worktree.enabled` for real
automatic isolation.

Assumption surfaced to the lead: card scope item 1 read literally ("new
production readers for auto_create AND auto_merge") is satisfied for
auto_merge only; auto_create resolves via the sanctioned scope route.

## §4 Trigger surface: session-exit, own branch only

Chosen: the session-exit disposal surface — the `cleanupSessionWorktree`
family (internal/cli/session_worktree.go:624), the M4 path. Ordering:
merge attempt BEFORE disposal; the merge does not need the source tree
(it consumes committed state), but disposal deletes the tree, so merge-first
is the only safe order. Composition: with both toggles ON, a clean exit does
merge → dirty/clean guards → remove.

Rejected: the M8 on-touch surface (`moai session register`/`list`,
session_worktree_prmerge.go:149). That sweep enumerates ALL WT-* worktrees —
the same shape applied to merging would merge OTHER sessions' in-flight
branches. Session-exit merges exactly the exiting session's own branch: one
branch, one owner, an explicit life-cycle moment. The M8 surface stays
cleanup-only (§C exclusions of spec.md).

Inherited guard inventory (each re-used or mirrored from the M4/M8
precedent):

| Guard | Precedent | Auto-merge form |
|---|---|---|
| clean-exit-only | REQ-SW-009 (session_worktree.go:637) | non-zero exit → no merge |
| dirty guard, immediate re-check | REQ-SW-010/024, EC-11 | source dirty → skip (REQ-WKW-007); target dirty → skip (REQ-WKW-008) |
| fail-open notices | REQ-SW-004 | every failure non-blocking (REQ-WKW-009) |
| distinct notice prefix | EC-13 | `AutoMerge*` prefix ≠ SessionExit/PRMerge prefixes |
| never force | t73 anchor guard spirit | never displaces any hold (REQ-WKW-004) |

## §5 Mechanics pinned for run phase

- **Window calls**: `kanban.AcquireIntegrationLock(root, lock, false)` — the
  `force` argument is a literal `false` constant on this path, never a
  parameter — then `kanban.ReleaseIntegrationLock(root, sessionID, false)`
  via `defer` so conflict and guard paths release too. Root is
  `integrationLockRoot()`. Holder identity: `integrationSessionID("")`
  semantics; if the session id cannot be resolved, the path skips with a
  notice (a hold with an invented holder cannot be released — the CLI
  already refuses that case; the auto path refuses the merge instead).
- **Target worktree**: `worktreeForBranch(branch)` (integration.go:155)
  resolves the tree holding the integration branch; empty → REQ-WKW-008
  skip. The merge runs with `cmd.Dir` set to that tree.
- **Ahead-of check**: `git rev-list --count <develop>..<branch>` > 0 gates
  the ceremony (REQ-WKW-010) — no window churn for already-merged branches.
- **Seam pattern**: function-variable injection exactly like
  `swapSessionWorktreeSeams` / `swapPRMergeSeams` so every git invocation on
  the path is observable to tests — this is also the mechanism of the
  AC-WKW-006 zero-push assertion (the seam log enumerates every executed git
  subcommand; none may be push/fetch/remote-mutating).
- **Conflict detection**: non-zero `git merge` exit while MERGE_HEAD exists
  → `git merge --abort` in the same dir, then release, then notice
  (REQ-WKW-006).
- **Template edit** (`workflow.yaml` comment) requires `make build`
  (//go:embed) before verification; no `.sh`/`.sh.tmpl` pair is involved.
