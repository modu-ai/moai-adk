---
id: SPEC-QUEUE-UPGRADE-PROOF-001
title: "Prove the v3.1.2-to-next queue upgrade path end to end"
version: "0.2.0"
status: draft
created: 2026-09-03
updated: 2026-09-03
author: manager-spec
priority: High
phase: "v3.2.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "queue, upgrade, migration, testing, kanban"
tier: M
---

## HISTORY

### v0.2.0 (2026-09-03)

- Plan-audit iteration 1 remediation (card t470). Frontmatter corrected so the
  SPEC is mechanically lintable (`tags` string form, `lifecycle: spec-anchored`,
  `priority: High`, quoted `title`/`version`). Added `REQ-QUP-010` so the
  anti-vacuity criterion is anchored in the requirement layer. `AC-QUP-002`
  gained a positive precondition and a relocation sentinel; `AC-QUP-008`'s
  gitignored `git status` limb replaced with a digest comparison; `AC-QUP-010`'s
  mutation replaced with one that actually produces RED. Both
  `[NEEDS CLARIFICATION]` markers are deliberately left open — they are the
  dispatcher's to resolve.

### v0.1.0 (2026-09-03)

- Plan-phase authoring for card t470. Scope is PROOF of the existing upgrade
  mechanism, not repair of it.

---

## §A Context

`v3.1.2` — the newest release tag — ships the backlog queue as a JSON document
at `.moai/state/kanban/backlog.json`. Local `develop` ships it as SQLite at
`.moai/state/todo/backlog.db`. No release tag carries the SQLite merge, so the
NEXT release asks every existing user to cross **two** transitions in one step:
a state-directory rename, and a JSON-to-SQLite conversion.

Each transition is separately well covered. Their **composition**, entered
through the `moai todo` command path rather than through the `BacklogStore`
API, has never been exercised.

Measured, in this worktree, at tree `4e4607abe`:

| Fact | Command | Observed |
|---|---|---|
| No release tag carries the SQLite merge | `git tag --contains 3cb258d62` | 0 lines |
| v3.1.2 kanban surface | `git ls-tree -r --name-only v3.1.2 -- internal/kanban/` | 47 files; `backlog_sqlite.go`, `backlog_migrate.go`, `state_dir.go`, `todo_root.go` all absent |
| v3.1.2 queue location | `git show v3.1.2:internal/kanban/backlog_store.go` (`BacklogPathForRoot`) | `.moai/state/kanban/backlog.json` |
| develop kanban surface | `git ls-tree -r --name-only origin/develop -- internal/kanban/` | 87 files; all four present |
| Schema version, both sides | `grep 'backlogVersion = '` at `v3.1.2` and at HEAD | `const backlogVersion = 1` on both |
| `moai update` cannot touch the queue | `grep -rn "internal/kanban" internal/cli/update/` | 0 matches |
| No CLI test seeds the legacy directory | `grep -rn "LegacyStateDirForRoot\|state.*kanban" internal/cli/` | 0 test seeds |

The upgrade mechanism itself is sound and this SPEC asserts nothing to the
contrary: the migration is fail-closed (parity verified before authority
flips), preserves every field, renames the legacy file to a `.migrated`
quarantine rather than deleting it, is idempotent, and is unreachable from the
template-deployment path.

## §B The two layers

**Directory layer** — `internal/kanban/state_dir.go`. `resolveStateDir(root,
adopt)` decides which state directory is served. Three rules: only-legacy →
relocate (a pure `os.Rename`, never copy-then-delete); both-exist → the new
name wins and the legacy directory is left strictly untouched; relocation
refused → serve the legacy layout READ-ONLY and do not error.

**Storage layer** — `internal/kanban/backlog_migrate.go`. A lazy one-time
migration at store open. On success the legacy file is RENAMED to
`backlog.json.migrated` (`backlogMigratedSuffix`), never deleted or truncated.
On any failure the partial database and its siblings are removed and the legacy
file is left authoritative.

Existing per-layer coverage, measured with `grep -c "^func Test"`:

- `internal/kanban/state_dir_test.go` — 5 tests, plus
  `state_dir_lock_test.go:27 TestStateDirRelocationUnderHeldLock`.
- `internal/kanban/backlog_migrate_test.go` — 10 tests.

Every one of them plants its layout by hand and calls the API directly. The
empty cell is the composed path entered through the CLI.

## §C Requirements (GEARS)

### REQ-QUP-001 — composed upgrade preserves the queue

**When** a user whose project carries only the v3.1.2 layout (a legacy state
directory holding a legacy JSON queue document) runs the first `moai todo`
command, the queue surface shall present every card, every card state, and the
`last_seq` high-water mark exactly as the legacy document recorded them.

### REQ-QUP-002 — composed upgrade relocates the directory

**When** the first `moai todo` command completes against a project carrying
only the legacy state directory, the directory layer shall have relocated that
directory to the current name, leaving no queue-bearing directory under the
legacy name.

### REQ-QUP-003 — the legacy document is quarantined, not destroyed

**When** the composed upgrade completes successfully, the storage layer shall
leave the legacy queue document present under its `.migrated` quarantine name,
and shall not delete or truncate it.

### REQ-QUP-004 — the database artifact exists afterwards

**When** the composed upgrade completes successfully, the storage layer shall
have produced the SQLite queue artifact beside the quarantined document.

### REQ-QUP-005 — the proof is a regression guard

The test suite shall carry the composed-path proof as an automated test, so a
later change to either layer that breaks the composition fails a check rather
than reaching a release.

### REQ-QUP-006 — the fixture matches a reachable state

The composed-path proof shall build its fixture from the record shape a
`v3.1.2` binary actually writes — `version`, `last_seq`, `items` and nothing
else — because those are the only three fields that release's `BacklogRecord`
carries.

### REQ-QUP-007 — forward-compatible fields survive the composition

**Where** a queue document in the legacy directory carries the additive
`findings` and `archived` fields, the composed upgrade shall preserve them
through to the migrated store. This state is reachable only for someone who ran
a development build, so this requirement is OPTIONAL.

### REQ-QUP-008 — isolation

The composed-path proof shall operate entirely inside a per-test temporary
directory and shall not read from or write to this repository's live queue.

### REQ-QUP-009 — no production change

This SPEC shall not change the behavior of the directory layer, the storage
layer, the migration, or the queue-root resolver. Its deliverable is tests, and
documentation of what was measured.

### REQ-QUP-010 — the proof is not accepted on an unfalsified GREEN

The composed-path proof shall not be accepted on a GREEN it has never been
shown capable of failing. A mutation that severs the composed path shall be
applied once, and the resulting failure shall be recorded verbatim, before the
passing result is claimed.

## §D Constraints

- **C-1** Every test uses `t.TempDir()` isolation. Nothing may touch
  `.moai/state/todo/backlog.db` in this repository.
- **C-2** Verification scope is `go test ./internal/kanban/... ./internal/cli/`.
  A full local suite is prohibited.
- **C-3** `./internal/cli/` alone runs past 600s on the development machine; any
  timing criterion must account for that.
- **C-4** No test may spawn background load. Anything requiring contention must
  be cleanup-guaranteed.
- **C-5** Run-phase evidence lands at `.moai/reports/t470/verdict.md`.

## §E Exclusions

This SPEC is a proof, not a repair. Everything below is deliberately out of
scope, recorded with its reason so a later reader does not re-open it as an
oversight.

### Out of Scope — production behavior change

- No change to `resolveStateDir`, `relocateStateDir`, `migrateLegacyBacklog`,
  `assertBacklogParity`, or the queue-root resolver. The mechanism is sound;
  changing it under a proof card would destroy the thing being proved.
- No redesign of the store, the migration, or the directory resolver.

### Out of Scope — cross-process concurrency (G3)

- Two OS processes contending on the same queue is NOT covered here.
- The common framing of this gap is wrong and the correction matters:
  `internal/kanban/backlog_concurrency_test.go:135 TestConcurrencyStress`
  already exists, but it drives in-process goroutines against a single
  `NewBacklogStore(...)` object. The real gap is cross-process, not
  "no concurrency test at all".
- Excluded because the surface is large (it requires spawning a second
  process), it is orthogonal to the upgrade path this card proves, and C-4
  forbids spawning background load. Candidate for a separate card.

### Out of Scope — a `moai doctor` check for the stale layout (G5)

- Measured: `grep -rln "kanban\|backlog" internal/cli/doctor*.go` returns 0, so
  `moai doctor` reports nothing about a project still on the old layout.
- Adding such a check is a feature, not a proof, and this card's scope is proof.
- The one situation where the silence bites, recorded so it is not lost: when
  relocation is REFUSED (a cross-device mount, an unexpected permission),
  `resolveStateDir` fails open and the user keeps running READ-ONLY on the
  legacy layout with no diagnostic on any surface.

### Out of Scope — the downgrade-comment tension

- `state_dir.go` carries a comment asserting the queue document keeps the name
  `backlog.json` so that an older binary can read it. The tension between that
  assertion and the observed rename + directory move is recorded in `plan.md`
  as a clarification marker. This SPEC neither resolves it nor proposes any
  change on its account.

### Out of Scope — schema versioning work

- `backlogVersion` is the constant `1` on BOTH `v3.1.2` and develop (the schema
  is additive), so there is no version-bump hazard to design around. Stated
  explicitly so a future reader does not re-open the question.
