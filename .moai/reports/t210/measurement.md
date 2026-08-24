# t210 — plan-phase measurement record

Base: origin/main `cd0cee1b8` (dispatch named `f7d4b7824`; verified ancestor — the base is newer, not divergent).
All commands below were run read-only from `.claude/worktrees/t210`.

## M1 — the premise reproduces

`CLAUDE_PROJECT_DIR=<primary> moai todo list` against the live queue, cross-read
against `gh pr list --state open`:

| card | queue state | open PR | verdict |
|---|---|---|---|
| t88  | picked | #1619 | in flight, queue agrees |
| t200 | queued | #1612 | **divergent** |
| t201 | queued | #1611 | **divergent** |
| t202 | queued | #1613 | **divergent** |
| t203 | queued | #1614 | **divergent** |
| t205 | queued | #1617 | **divergent** |

5 of the 6 named cards reproduce the lead's finding. t88 had already been picked
by the time this lane measured, so the live divergence count is 5, not 6.

## M2 — every open PR carries a card token, but no carrier is both complete and precise

`gh pr list --state open` (11 PRs), scanning title / body / commit messages for `\bt[0-9]{1,4}\b`:

| PR | title token | body tokens | commit-message tokens | delivering card |
|---|---|---|---|---|
| 1621 | t211 | t211 | t211 | t211 |
| 1619 | — | t88 | t184,t187,t204,t83,t88 | t88 |
| 1617 | t205 | t205 | t205 | t205 |
| 1614 | t203 | t1,t151,t203,t69,t9 | t151,t203,t69 | t203 |
| 1613 | t202 | t202 | t202 | t202 |
| 1612 | t200 | t200,t201 | t200 | t200 |
| 1611 | — | t201 | t201 | t201 |
| 1606 | t187 | t187,t192 | t184,t187 | t187 |
| 1605 | t194 | t181,t194 | t181,t194 | t194 |
| 1601 | — | t188 | t188 | t188 |
| 1600 | — | t184 | t127,t131,t142,t158,t159,t164,t165,t170,t171,t173,t81,t82,t83,t89,t91 | t184 |

Carrier scorecard (n=11):

- **PR title** — precision 7/7 (every token present is exactly the delivering
  card), recall 7/11 (64%). Precise, incomplete.
- **PR body** — recall 11/11 (100%), precision poor: 5 PRs carry extra tokens,
  #1614 carries 5 tokens for 1 card. Complete, noisy.
- **commit messages** — recall 10/11, and **wrong on #1600**, whose 15 tokens do
  not include its delivering card t184. A branch that merges the release branch
  inherits every other card's commits, so the noise scales with integration
  rather than with the card. The worst carrier of the three, despite being the
  one `kanban-dispatch.md` § Isolation already makes [HARD].

Consequence for the design: a reverse index over the PR **body** finds every card
but mislabels several; over the **title** it never mislabels but misses four.
Neither is usable alone. This is a measurement, not a preference.

## M3 — the queue already has a "record, never mutate" precedent

`.claude/skills/moai/workflows/todo.md` § What the analyser may do carries a
[HARD] clause: analysis "never folds one card into another, never reorders the
queue, never drops a card, and never edits one... Acting on a record is the
operator's act." The `findings[]` array in `backlog.json` holds observations
ABOUT cards without touching them.

That precedent settles the boundary question the dispatch raised: an observation
surface already exists in this subsystem and is already fenced off from mutation.

## M4 — no existing card↔PR machinery

`grep -rn 'gh pr' internal/ --include=*.go` finds `gh` used for PR *state* checks
(`session_worktree_prmerge.go`, `branch_protection.go`, `internal/github/gh.go`,
`internal/statusline/forge.go`) but nothing mapping a card id to a PR. The index
would be new.

## Gaps (explicitly not measured)

- Closed/merged PRs were not scanned — only `--state open`. A card whose PR has
  already merged (the t199 case the dispatch names) is a different query, and its
  carrier statistics are unmeasured.
- `gh` latency was not measured. A per-render network call may be too slow for
  `moai todo list`, which is documented as lock-free and cheap.
- Whether a non-GitHub forge (the `glab` path `internal/statusline/forge.go`
  contemplates) needs the same index is unexamined.

## M5 — `gh` latency closes one of the gaps above

    time gh pr list --state open --limit 40 --json number,title,body
    → 0.878s total (0.14s user, 0.06s sys — the rest is network)

Roughly nine-tenths of a second per query, all of it round-trip. `moai todo list`
is documented as lock-free and cheap; making it pay ~0.9s of network on every
render would change that character. This argues for a dedicated read verb or an
opt-in flag rather than a default-on column, and it is a measurement rather than
a preference.

## M6 — the title convention is already emergent practice

`gh pr list --state merged --limit 15`: 8 of 15 merged PR titles carry a card
token. Of the 7 that do not, most deliver no card at all — `release: v3.1.3`,
the v3.1.3 batch PR, and three `chore(release-update)` entries. Among merged PRs
that do deliver a single card, the title token is close to universal.

REQ-3(b) is therefore codification of a convention the repository already mostly
follows, not a new burden imposed on contributors.

## Remaining gaps after M5/M6

- Merged-PR carrier statistics were sampled (M6) but not scored for precision the
  way M2 scores the open set. The merged-side query stays out of scope.
- Non-GitHub forge support remains unexamined.
