---
id: SPEC-TODO-LAND-AUTO-DONE-001
title: "Auto-done on remote land — an evidence-gated landing scan that closes landed cards at the moment the lead confirms origin/develop, with three misfire guards and a reversible audit log"
version: "0.4.0"
status: completed
created: 2026-09-13
updated: 2026-09-13
author: manager-spec (card t684)
priority: P1
phase: "v3.2.1 target"
module: "internal/kanban, internal/cli, .claude/skills/moai/workflows/todo.md, internal/template/templates/.claude/skills/moai/workflows/todo.md"
lifecycle: spec-anchored
tags: "kanban, backlog-queue, landing-evidence, auto-done, remote-landing, misfire-guard, audit-log, cli"
tier: M
related_specs:
  - SPEC-TODO-LANDING-EVIDENCE-001
  - SPEC-TODO-LANDING-ATTRIBUTION-001
  - SPEC-TODO-LANDING-STATE-001
  - SPEC-TODO-DESTRUCTIVE-GUARD-001
  - SPEC-KANBAN-QUEUE-PR-SYNC-001
---

# SPEC: Auto-done on remote land

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-09-13 | Initial plan-phase authoring (card t684), grounded in the tree at `WT-auto-done-on-land` HEAD `74d872aaf`. Motivating measurement supplied by the lead (2026-09-13): 45 cards already landed on remote `origin/develop` were left in `queued`/`picked` state because nobody ran `done`; the lead batch-completed them by hand. Existing surfaces verified in-source: `newTodoDoneCmd` + `--require-landed` + `todoRequireLanded` (`internal/cli/todo.go`), `GitLandedQuerier.Landed` + `subjectAttribution` (`internal/kanban/prlink_landed.go`), `LandingEvidence` + closed `operator` provenance (`internal/kanban/landing_evidence.go`), `moai todo landed` (`internal/cli/todo_landed.go`), `moai todo undone` (`internal/cli/todo_undone.go`), the integration window (`internal/cli/integration.go`, `internal/kanban/integration_lock.go`). Trigger axis decided in plan.md §C (option (c): a scan command invoked at the remote-landing-confirmation moment). |
| 0.2.0 | 2026-09-13 | Plan-audit delta fix (FAIL 0.75 → re-audit scoped to D1-D3). D1: `--fetch` (REQ-AD-003) gained AC-AD-015 (fetch-count assertion with a counting script-git fixture). D2: REQ-AD-002's close-surface exclusivity gained AC-AD-016 (scope test that only `done` and `auto-done` reach the archive transition, plus a `todo landed` control run proving no archive occurs). D3: the skip `reason` vocabulary is enumerated as a closed four-token set in REQ-AD-010 and the DoD count corrected. Recommended items folded in: AC-AD-008's outcome pinned to the always-emitted `skip <id> reason=query-inconclusive` line with the exit-code policy made explicit; REQ-AD-014's doc-step promoted to AC-AD-017 (template + mirror content check); the `LandedRefForWithLevel` citation corrected to its actual definition site (`internal/kanban/prlink_landedref.go:63`, call site `internal/cli/todo.go:134`). Four lane-level design decisions recorded in progress.md §E.1. |
| 0.3.0 | 2026-09-13 | Re-audit PASS (0.875) final delta: E1 — AC-AD-016's `todo landed` control-run clause re-worded to the observable contract (State/position/text/spec-id unchanged, no archive entry; only the `landing` field may change — the previous byte-identity wording contradicted `runTodoLanded`'s `Landing`-column write, `internal/cli/todo_landed.go:127-140`). E2 — the doc-step AC folded back into DoD §D.3. New AC-AD-017 (fixture slot reused) added from the lead's field observations (2026-09-13): two real FALSE-NEGATIVE shapes (t603 comma-form subject `d8b7836aa`; t681 recorded-SHA-only `ae980ef2d`/`4fb28a5c6`) that must attribute/close, stated in contrast to the still-binding reissued-id collision gates (t654/t656/t657). §D Out of Scope bullet amended to carve the comma-form attribution into scope. AC count 14 → 17. |
| 0.4.0 | 2026-09-13 | E3 one-character fix (entry verdict PASS 0.9375): AC-AD-017 Shape B's repair-commit SHA corrected `4fb28a4c6` → `4fb28a5c6` (verified in-tree: resolves to `fix(worktree): make clean --json a pure read in every flag combination`, second parent of merge `ae980ef2d`); the same erroneous SHA corrected in the 0.3.0 HISTORY row and progress.md §E.1 where it had propagated. No other content touched — the run-gate re-check scope is E3 only. |

## §A Requirements

### §A.0 Background and problem

The backlog queue is the delegation channel for Kanban/Factory mode. A card leaves the queue
only through `moai todo done <n>` (`newTodoDoneCmd`, `internal/cli/todo.go`) — a manual act.
Under the git-flow integration chain (CLAUDE.local.md §4.1), a card's work lands on remote
`develop` when the LEAD batch-pushes develop; lanes never push. Between "work merged into
local develop" and "lead pushes" the card's worktree branch is the ONLY copy of the work, so
a local merge is explicitly NOT a landing. On 2026-09-13 the lead measured 45 cards whose
work had landed on `origin/develop` but which still sat in `queued`/`picked` state — stale
queue state that distorts dispatch decisions (a picked card holds a slot; a queued card gets
re-admitted to analysis).

The missing piece is an AUTOMATED, evidence-gated close at the remote-landing-confirmation
moment, with guards against the three misfire classes already observed in this repository,
and with reversibility.

### §A.1 Trigger contract

**REQ-AD-001 (Event-driven)**
**When** the operator runs `moai todo auto-done`, the scan shall resolve the landed ref
through the existing chain (`LandedRefForWithLevel`, defined at
`internal/kanban/prlink_landedref.go:63`, called from `todoLandedRefResolved` at
`internal/cli/todo.go:134`; default `origin/develop`) and evaluate ONLY live backlog
items whose `State` is `BacklogStateQueued` or `BacklogStatePicked`.

**REQ-AD-002 (Ubiquitous)**
The scan command shall be the ONLY surface that closes a card without an explicit
per-card `done` invocation; no other verb, hook, or daemon shall transition a card to
archived as a side effect of its own operation.

**REQ-AD-003 (Event-driven)**
**When** the scan runs with `--fetch`, it shall run `git fetch` for the landed ref's remote
before evaluating; absent the flag it shall read the local remote-tracking refs only (the
offline posture `GitLandedQuerier` already holds — REQ-2.3 of SPEC-KANBAN-QUEUE-PR-SYNC-001).

### §A.2 Evidence forms — what counts as "landed"

**REQ-AD-004 (Ubiquitous)**
The scan shall close a card only on ONE of exactly two evidence forms, both evaluated
against the resolved landed ref:

1. **Recorded delivering SHA** — the card's `Landing.SHA` (operator-asserted via
   `moai todo landed --sha`, provenance `operator`, `internal/kanban/landing_evidence.go`)
   resolves to a commit reachable from the landed ref at scan time
   (`git merge-base --is-ancestor`).
2. **Attributed subject** — the landed ref's subject stream carries a subject that
   `subjectAttribution` (`internal/kanban/prlink_landed.go`) attributes to the card id.

**REQ-AD-005 (Ubiquitous)**
The scan shall never close a card on an inconclusive evaluation: an unanswerable git (no
git binary, ref resolves to nothing, query error) yields a SKIP for that card, never a
close — the three-valued-answer asymmetry (`LandingUnknown` is not evidence of landed)
holds at the scan layer too.

### §A.3 Misfire guards

**REQ-AD-006 (Event-driven — guard M1, card-id reissue collisions)**
**When** the id under evaluation has carried MORE THAN ONE DISTINCT card text across the
live queue and the archive (the reissue shape observed on t654/t656/t657), the scan shall
close that card ONLY on evidence form 1 (a recorded delivering SHA — the one form that
names THIS card's actual commit); on form 2 alone it shall skip the card with reason
`ambiguous-id`, because a subject mentioning the token `t654` cannot distinguish the
predecessor card's landed work from the reissued card's pending work.

**REQ-AD-007 (Event-driven — guard M2, sync not yet done)**
**When** a card carries a non-empty `SpecID` and the SPEC frontmatter status read at scan
time (`ReadPrimarySpecStatus`) is anything other than `completed`, the scan shall skip the
card with reason `spec-not-completed` — a run commit landing does not license the close
while the card's sync phase is unfinished.

**REQ-AD-008 (Event-driven — guard M3, non-landing declarations)**
**When** a commit subject carries an explicit non-landing declaration (a negation marker:
case-insensitive `not merged` or `not landed` appearing in the same subject as the card
token — the (t581) precedent of a commit message recording that the work did NOT land),
that subject shall attribute NOTHING: the subject is excluded from form-2 evaluation and
the scan shall not count it as landing evidence. A negated subject's card token must not
leak into any other subject's evaluation.

**REQ-AD-009 (Ubiquitous)**
The scan shall not write to the `Landing` evidence column: the machine-derived close
evidence lives in the scan's own execution log (REQ-AD-011), and the
`LandingSHASourceOperator` closed value set (`internal/kanban/landing_evidence.go`
`Validate()`) shall remain unchanged — a stored delivering SHA stays operator-asserted
or absent (REQ-TLE-012).

### §A.4 Transition, log, reversibility

**REQ-AD-010 (Event-driven)**
**When** the scan closes a card, it shall archive it through the same locked
`ArchiveCard` path `moai todo done` uses (byte-identity-on-refusal contract inherited),
and shall print one line per closed card carrying the canonical `done <id> landing=landed`
prefix every existing reader keys off, plus a `source=auto-land` marker and the answering
ref; it shall print one `skip <id> reason=<reason>` line per skipped card and one summary
line last. The skip `reason` vocabulary is CLOSED at exactly four tokens, named here as
the canonical set every reader and the `--help` body enumerate:
`ambiguous-id` (guard M1), `spec-not-completed` (guard M2), `not-landed` (no landing
evidence, including a negated subject), and `query-inconclusive` (REQ-AD-005).

**REQ-AD-011 (Ubiquitous)**
The scan shall append one JSONL row per closed card to an append-only execution log under
the queue's runtime state directory (`RuntimeStateDirForRoot`), each row carrying at
minimum: the card id, the transition instant (RFC 3339 UTC), the evidence form
(`sha-recorded` | `subject-attribution`), the attributed subject line and its commit SHA
when form 2 applies, the recorded delivering SHA when form 1 applies, the landed ref and
its head at scan time, and the skip reason for skip rows.

**REQ-AD-012 (Event-driven — reversibility)**
**When** the operator runs `moai todo undone <n>` on a card the scan closed, the restore
shall behave exactly as it does for a manually closed card (card + findings restored at
the held position, refused on reissued id), and the log shall gain a reversal row naming
the original closure row, so every auto-close is reversible with a traceable inverse.

**REQ-AD-013 (Capability gate)**
**Where** the operator wants to inspect the scan's decisions without mutating the queue,
the scan shall accept `--dry-run`, which evaluates every card, prints every close and
skip line it would produce, and leaves the queue record byte-identical.

### §A.5 Integration with the lead's push procedure

**REQ-AD-014 (Ubiquitous)**
The workflow documentation (`.claude/skills/moai/workflows/todo.md` and its template
mirror) shall name the scan as the step the LEAD runs immediately after its post-push
remote-landing confirmation (`git fetch origin develop && git rev-parse origin/develop`),
so the automation's trigger point is the documented remote-landing-confirmation moment —
never the local merge moment, and never a lane-owned step (lanes never push).

## §B Non-functional constraints

- **NFR-1 (Testability)**: every misfire guard is verified against a fixture git
  repository built with the existing `internal/kanban/temp_origin.go` helper; no test
  touches a real remote.
- **NFR-2 (Scope)**: the scan mutates at most the queue store and its own log file; it
  performs no network I/O unless `--fetch` is given.
- **NFR-3 (Subagent boundary)**: nothing in the scan prompts (C-HRA-008 / REQ-TODO-014);
  all output is structured stdout lines with human-readable errors on stderr.
- **NFR-4 (Idempotence)**: running the scan twice in a row with no intervening change
  closes nothing the second time (already-archived ids are not re-evaluated).

## §C Success criteria

- All three misfire fixtures fail-closed: reissued id (M1), run-landed-sync-pending (M2),
  negated subject (M3) — each leaves the card unarchived and emits the named skip reason.
- A landed card with completed SPEC closes on one scan invocation and its closure is
  reversed by `undone` with a log reversal row.
- `--dry-run` leaves the queue byte-identical while printing the same decisions.

## §D Out of Scope

### Out of Scope — trigger mechanisms not chosen

- No PostToolUse or SessionStart hook is added to fire the scan after a `git push`
  (trigger option (a)) — rejected in plan.md §C.
- No coupling to `moai integration release` (trigger option (b)) — it fires before push
  by design and is a local-merge moment, not a remote-landing moment.
- No daemon, watcher, cron installer, or background loop is shipped; the scan runs when
  invoked (the lead's procedure or an operator-run loop may invoke it).

### Out of Scope — landing-evidence model changes

- No new `LandingSHASource` provenance value is introduced; the `Landing` column's
  closed `operator` value set is untouched (REQ-AD-009).
- The scan does not improve `subjectAttribution`'s six positional shapes beyond adding
  the negation exclusion (REQ-AD-008) and the comma-form trailing-parenthetical
  attribution the observed false-negative fixture requires (acceptance.md AC-AD-017
  Shape A, t603 / commit `d8b7836aa`); broader shape reclassification is
  SPEC-TODO-LANDING-ATTRIBUTION-001's domain.
- No persisted per-card run-vs-sync landing-state field is added — that axis remains
  with the SPEC-TODO-LANDING-STATE-001 family; the scan uses the SPEC frontmatter
  status read instead (REQ-AD-007).

### Out of Scope — queue semantics

- The scan does not reorder, re-prioritize, fold, relate, or admit cards; it only
  archives (closes) and logs.
- `dropped` cards are never resurrected or closed by the scan.
- Batch size, rate, and scheduling of scan invocations are operator policy, not
  CLI behavior.
