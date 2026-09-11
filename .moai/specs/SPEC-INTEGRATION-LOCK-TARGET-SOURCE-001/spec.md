---
id: SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001
title: "Integration lock: surface the branch's provenance and the card in the window record (card t637)"
version: "0.3.0"
status: draft
created: 2026-09-11
updated: 2026-09-11
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/cli, internal/kanban, internal/config"
lifecycle: spec-anchored
tags: "integration-lock, kanban, factory, git-flow, observability, t637"
tier: M
---

# SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001 — Integration Lock Target Provenance and Card Visibility

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-09-11 | manager-spec | Initial draft (card t637). Scope = operator decision "option C, all four items", recorded in `.moai/reports/t637/verdict.md` §8. |
| 0.2.0 | 2026-09-11 | manager-spec | plan-audit iter1 repair (`.moai/reports/t637/plan-audit-SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001-iter1.md`, FAIL 0.74). REQ-ILT-006 fixes when the warning is emitted (after the record is written; never on a refused acquire) — D10. REQ-ILT-010 narrowed to the sites enumerated in §B premise 7, with runtime guidance messages excluded in §E — D6. §D adds `internal/template` to the verification scope — D7. acceptance.md rewritten for D1-D5, D11, D12, D15. |
| 0.3.0 | 2026-09-11 | manager-spec | plan-audit iter2 repair (`.moai/reports/t637/plan-audit-SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001-iter2.md`, FAIL 0.88), acceptance.md and plan.md only — no requirement, design, or scope change. N1: the fixture binary is written literally, the `BIN` variable is removed. N2: mutation row b names only the github-flow cell; new row i (absent config treated as git-flow) names the no-config cell. N3: fixture `cd` runs in a subshell and every relative-path command carries its own `cd`. N4: the seam test is named concretely with an anchored selector. N5: the release line is checked to carry no `--card <`. N6: the neutrality positive control is no longer SPEC-ID-shaped. iter1 D13 (ownership trailer) stays declined — reason recorded in `progress.md` §E.1. |

## §A Context and Problem

`moai integration acquire` records who holds the release-integration window and on which
`branch` / `worktree`. Card t449 already made the branch resolution three-tiered — an explicit
`--branch` value, then the project's configured git-flow develop branch, then the caller's own
tree — and gave the `branch` field the meaning "the integration TARGET (the branch the merge
lands on)". That meaning is correct and stays.

The defect this SPEC addresses is **visibility**, not resolution. The evidence base is
`.moai/reports/t637/verdict.md` (read on tree `4c99d973e`; the SPEC target files are
byte-unchanged between that tree and `origin/develop` `ee99507fb`). It establishes two layers:

1. **Configuration layer — the fallback is silent.** When a project's git strategy is git-flow
   but its `develop_branch` is empty, resolution falls to the caller's tree without any signal.
   A lane sitting in one card worktree then records that card's branch as the "integration
   target" while it may be merging a different card. The verdict's fixture reproduction shows
   both shapes: cell C1 (same tree, own branch — the record happens to match) and cell C2
   (caller in tree A, merging branch B — the record names A). Nothing in the record or in the
   `status` output lets a reader tell a configured target from a caller-tree fallback.
2. **Surface layer — the card is invisible in text.** The field that identifies *which card*
   is being integrated is `card` (from `--card`). The record carries it and `status --json`
   prints it, but the `status` text output does not. In addition, no documented `acquire`
   invocation passes `--card` at all, and the `--branch` help text ("Branch being integrated")
   reads as the SOURCE card branch while the implementation treats it as the TARGET — a lane
   following the help text reproduces the verdict's cell C4, where the window is recorded
   against the card tree.

The production occurrence is currently masked: the primary checkout's `git-strategy.yaml`
manual block has since been filled (`develop_branch: develop`). The SPEC still targets making
the fallback visible and the card visible, because the masking rests on a local, hand-restored
configuration value that `moai update` is documented to reset (CLAUDE.local.md §2.3).

## §B Premises (measured, line-pinned to tree `1ad0fdc09`)

Line numbers are pinned to this worktree's HEAD `1ad0fdc09`; re-verify before citing them from
any other tree.

1. Resolution order lives in `internal/cli/integration.go:117-131`; the caller fallback reads the
   caller's branch (`currentBranch()`, line 97) and the process working directory (line 129).
2. The configured develop branch comes from `internal/config/loader_integration_branch.go:34-48`,
   which returns a non-empty value only when `mode == manual` AND the active profile's
   `workflow == git-flow` AND `develop_branch` is non-blank. It returns the empty string for
   "not git-flow", "git-flow with an empty develop_branch", and "file absent or unreadable"
   alike — the three cases are indistinguishable to its caller today.
3. The lock record type is `kanban.IntegrationLock` (`internal/kanban/integration_lock.go:109-133`);
   its optional fields already use `omitempty` (`session_name`, `pid_source`, `card`, the two
   settings-drift fields).
4. The `status` text printer (`internal/cli/integration.go:213-214`) prints holder, branch,
   worktree and since — no card.
5. The `--branch` flag help (`internal/cli/integration.go:292`) reads "Branch being integrated
   (default: the configured git-flow develop branch, else the current branch)".
6. The PreToolUse guard (`internal/hook/integration_lock_guard.go:100-101`) uses `lock.Branch`
   only inside its refusal message; it is not a decision input.
7. Documented `acquire` invocations (tracked files in this worktree): kanban-dispatch rule line
   230 (local copy and template mirror, currently byte-identical), the local git-flow lane
   protocol lines 47 and 51, the release harness specialist line 345, and CLAUDE.local.md lines
   370, 389 and 403. None passes `--card`.

## §C Requirements (GEARS)

### Card visibility in text status

- **REQ-ILT-001** — While the recorded window carries a non-empty card, the `status` text output shall print a `card:` line naming that card.
- **REQ-ILT-002** — While the recorded window carries no card, the `status` text output shall
  print no `card:` line, so a card-less record keeps today's text shape.

### Branch provenance

- **REQ-ILT-003** — When `acquire` records a window, the lock record shall carry a
  `branch_source` value naming how the recorded branch was resolved: `flag` when a non-blank
  `--branch` value decided it, `config` when the configured git-flow develop branch decided it,
  and `caller` when the caller's own tree decided it.
- **REQ-ILT-004** — While a record carries `branch_source`, the `status` text output shall show
  that provenance on the branch line, and the `status --json` output shall carry it as the
  `branch_source` key of the lock object.
- **REQ-ILT-005** — When `status` reads a record written before `branch_source` existed, the command shall succeed, print the branch line in today's exact shape without any provenance annotation, omit the `branch_source` key from its JSON lock object, and leave the record file unmodified.
- **REQ-ILT-006** — Where the project's git strategy is git-flow (manual mode with an active profile whose workflow is git-flow), when `acquire` resolves the branch through the caller fallback because the configured develop branch is empty or blank and no non-blank `--branch` was given, the `acquire` command shall write exactly one warning line to standard error that names the recorded caller branch and both remedies (setting the develop branch in the git strategy configuration, or passing `--branch`), and shall write it only after the window record has been written — an `acquire` refused because the window is held emits no warning.
- **REQ-ILT-007** — Where the project's git strategy is not git-flow, or the git strategy file is absent or unreadable, the caller fallback shall emit no warning; a `flag` or `config` resolution shall emit no warning in any project.
- **REQ-ILT-008** — The `acquire` command shall not refuse the window, change its exit status,
  change its standard-output lines, or alter the resolved branch and worktree because of the
  REQ-ILT-006 warning condition.

### Flag help

- **REQ-ILT-009** — The `--branch` flag help of `acquire` shall describe the value as the
  integration TARGET branch (the branch the merge lands on), shall not describe it as the branch
  being integrated, and shall keep naming the default resolution (configured git-flow develop
  branch, else the current branch).

### Documented invocations

- **REQ-ILT-010** — Every lane-level `acquire` invocation among the documentation sites enumerated in §B premise 7 shall carry `--card <card-id>`; the release back-merge invocation among them shall carry an explicit note that a release integration delivers no card, instead of a `--card` value.
- **REQ-ILT-011** — The distributed template mirror of the kanban-dispatch rule shall remain
  byte-identical to its local copy after the edit and shall gain no card id, SPEC ID, date, or
  commit hash; the card placeholder shall be the literal `<card-id>`.

### Invariants

- **REQ-ILT-012** — The `branch` field shall keep the meaning "integration target", the
  resolution order shall remain flag → configured develop branch → caller tree, and the recorded
  `branch` / `worktree` values shall remain unchanged for every scenario the pre-existing
  resolution tests cover.
- **REQ-ILT-013** — The integration-lock guard shall not read `branch_source` or `card` as a
  decision input; its allow/deny outcome shall be unchanged for every record shape.

## §D Constraints

- Verification is scoped to the touched packages (`internal/cli`, `internal/kanban`,
  `internal/config`, `internal/hook` for the REQ-ILT-013 regression, and `internal/template` for
  the embedded template mirror edit). `go test ./...` is not
  run locally; CI is the full-suite judge. `internal/cli` compilation is a lead-gated slot in
  run-phase.
- No acceptance criterion depends on the primary checkout's configuration. Fixture evidence uses
  `/tmp/t637-fx` with `CLAUDE_PROJECT_DIR=/tmp/t637-fx` set in the same invocation, and Go tests
  use `t.TempDir()` scratch repositories.
- No `.moai/config` file of this repository is edited by this SPEC.
- Template content neutrality (CLAUDE.local.md §2.1) binds the template mirror edit.

## §E Exclusions (What NOT to Build)

### Out of Scope — resolution and configuration changes

- Changing the resolution order, adding a resolution tier, or renaming the `branch` field (the
  verdict's option B is rejected — it would make the field false on the configured path).
- Refusing `acquire` on the git-flow empty-develop-branch fallback, or turning the warning into a
  configurable gate.
- Repairing the primary checkout's `git-strategy.yaml`, or changing how `moai update` treats the
  git-flow keys (lead and operator territory, tracked separately).

### Out of Scope — record and guard behavior

- Making the guard (`integration_lock_guard.go`) decide on `branch`, `branch_source`, or `card`.
- Rewriting or migrating existing lock records to add `branch_source`.
- Adding a warning or provenance field to the `acquire --json` output object; provenance is read
  through `status --json`, and the warning travels on standard error in both output modes.
- Adding `--card` to the `acquire` guidance the CLI and the guard print at runtime (for example
  the guard's "reclaim with `moai integration acquire`" advisory and its `--force` hint); those
  are messages, not documentation sites, and the guard stays unedited (REQ-ILT-013).
- Inferring the source card branch the caller intends to merge — the tool cannot know it; this
  SPEC surfaces the mismatch rather than resolving it.

### Out of Scope — documentation beyond the enumerated invocations

- Regenerating or editing `.moai/project/codemaps/*.md` (generated artifacts; their mentions are
  descriptive, not argument-bearing invocations).
- CHANGELOG.md (owned by manager-docs in sync-phase).
- Rewording the surrounding prose of the edited documents beyond the invocation itself.

## §F Traceability

| Topic | Requirements | Acceptance |
|---|---|---|
| Card line | REQ-ILT-001, 002 | AC-ILT-001, AC-ILT-002, AC-ILT-006 |
| Provenance recorded and shown | REQ-ILT-003, 004 | AC-ILT-003, AC-ILT-004, AC-ILT-005, AC-ILT-006 |
| Old-record compatibility | REQ-ILT-005 | AC-ILT-009 |
| Git-flow fallback warning | REQ-ILT-006 | AC-ILT-005, AC-ILT-006 |
| No warning where legitimate | REQ-ILT-006, 007 | AC-ILT-003, AC-ILT-004, AC-ILT-007 |
| Warn-only, never refuse; no warning on a refused acquire | REQ-ILT-006, 008 | AC-ILT-008 |
| Flag help | REQ-ILT-009 | AC-ILT-010 |
| Documented invocations | REQ-ILT-010 | AC-ILT-011 |
| Template mirror | REQ-ILT-011 | AC-ILT-012 |
| Invariants | REQ-ILT-012, 013 | AC-ILT-013 |
| Mutation guard (rows a-i) | REQ-ILT-001, 003, 005, 006, 007, 008 | AC-ILT-014 |
