---
id: SPEC-SESSION-REGISTRY-READ-ANCHOR-001
title: "implementation plan — session registry READ-path anchoring"
version: "0.1.0"
created: 2026-09-21
updated: 2026-09-21
author: manager-spec
---

# SPEC-SESSION-REGISTRY-READ-ANCHOR-001 — implementation plan

Milestones are ordered by **decision reversibility**: the decisions most likely to
change sit first, the mechanical work last. The largest decision in this SPEC — the
S2 migration question — is therefore M1, ahead of the repair that has no decision in
it at all.

---

## §A Context

Card t1058 (synthesis A5). Worktree `.claude/worktrees/t1058`, branch
`WT-read-anchor`. Evidence base `.moai/reports/t1058/pre-plan-measurement.md` and
its four raw companions.

The card's premise was reversed by measurement for one of the two read paths
(spec.md §A.2). This plan is written to the measured state, not to the card text.

---

## §B Known issues carried into the plan

1. **The card's framing is wrong and must not be reinstated.** "Anchor the read
   paths" as one change hides that R1 loses live sessions while R2 loses nothing.
   Any restatement that merges them re-creates the defect this SPEC exists to
   record.
2. **The disposal guard is the blast radius, not the registry.** R1's four
   consumers (spec.md §A.6) are all disposal decisions. A wrong "free" verdict
   destroys an unpushed branch's only copy — the failure is silent and permanent.
3. **Figures decay.** The orphan population moved within a day (spec.md §A.8). No
   milestone here may consume a stored count.
4. **R2's mechanism is at synthetic-fixture grade.** M2's first act is to decide
   whether to raise that grade, not to assume it (spec.md §A.4).

---

## §C Pre-flight

Before any milestone runs:

- Confirm the worktree is `.claude/worktrees/t1058` and the branch `WT-read-anchor`.
- Confirm `internal/session/registry_path_anchor_test.go` passes unchanged — it is
  the write-path guard this SPEC must not disturb (REQ-RAR-009).
- Re-read `.moai/reports/t1058/pre-plan-measurement.md`. Do not work from this
  plan's summary of it.

---

## §D Constraints

- Verification is scoped to `./internal/session/...` and `./internal/hook/...`.
  The full suite is CI's job, not this lane's.
- No orphan file is deleted, moved, or rewritten by any milestone (REQ-RAR-010).
- No milestone edits `internal/session/registry.go` write-path anchoring
  (REQ-RAR-009).
- Evidence is exported to `.moai/reports/t1058/` before any verdict cites it.

---

## §E Self-verification

Each milestone closes by naming: the command run, its verbatim output, the tree it
ran against, and what it did not observe. A milestone that cannot name all four is
not closed.

---

## §F Milestones

### M1 — the S2 migration decision (highest change likelihood; operator-owned)

**This is a decision milestone, not an implementation milestone.** It produces a
recorded operator decision and nothing else.

The question, stated as measurement left it: *may a still-live registry entry that
exists only in an orphan file be discarded when the read path stops reading that
file, or must it first be migrated into the primary registry?*

Three branches, presented and not selected (REQ-RAR-005):

| Branch | What it commits to | What it costs |
|---|---|---|
| **B1 — migrate first** | S2 lands only after live orphan-only entries are written into the primary registry | a write to the primary registry, which races other live sessions (Residual-risk); a migration path must be built and is not currently in scope (spec.md §E) |
| **B2 — discard, recorded** | S2 lands with the operator having recorded that the loss is accepted | the disposal guard stops seeing those sessions; the loss is real and the record is what makes it a decision rather than an accident |
| **B3 — do not land S2** | only S1 lands; R1 keeps reading both roots | the orphan population keeps growing as residue, and R1 keeps its current two-root read |

The ordering is by **reversibility**, not preference: B3 changes nothing and is
trivially revisited, B1 adds a mechanism, B2 accepts an irreversible loss.

Exit: an operator decision recorded at the Implementation Kickoff Approval gate.
Where the decision is B1 or B2, AC-RAR-005's re-measurement (AC-RAR-006) is owed
before landing.

### M2 — R2's evidence grade (decision, then optionally work)

Decide whether to raise R2's mechanism from synthetic-fixture grade to a
reproduction against the real orphan population (spec.md §A.4, Gaps G3).

- **Raise it**: construct a reproduction against a real orphan file, with a
  positive control that fires. Record both.
- **Leave it**: proceed on the code reading plus the synthetic observation, and say
  so in the run-phase evidence at that grade. This is admissible — the grade is
  stated rather than upgraded.

Either outcome is acceptable; what is not acceptable is proceeding while narrating
the synthetic observation as a reproduction.

### M3 — S1: the R2 repair (implementation)

Change the R2 registry resolution so that it reaches the repository's primary
registry instead of stopping at the first `active-sessions.json` above a candidate
working directory (REQ-RAR-002, REQ-RAR-003).

- The repair reuses the existing anchor seam rather than growing a second private
  walk — the same seam the write path already uses (spec.md §A.5).
- Fail-open is preserved exactly (REQ-RAR-004).
- Lands in its own commit, with its own verification, independent of any S2 work
  (REQ-RAR-001).

Verification: `go test ./internal/hook/... ./internal/session/...`, with a new test
covering the all-candidates-inside-one-worktree case (AC-RAR-002).

### M4 — S2: the R1 anchoring (implementation; blocked on M1)

Runs only where M1 returned B1 or B2 **and** AC-RAR-006's re-measurement has been
performed against the landing-time tree and host.

Anchor `LiveAnchoredSessions` to the single registry, preserving the disposal
guard's contract (REQ-RAR-008) for every session whose entry reaches that registry.

Verification: `go test ./internal/session/... ./internal/cli/worktree/...`, plus the
disposal-guard control case of AC-RAR-008.

### M5 — preserved-behaviour guards (mechanical; lowest change likelihood)

- `internal/session/registry_path_anchor_test.go` unchanged and passing
  (REQ-RAR-009).
- No orphan file touched — verified by a file-state comparison over the orphan list
  before and after (AC-RAR-009).
- Windows references recorded as unmeasured wherever they appear (REQ-RAR-011).

---

## §G Anti-patterns

- **Re-merging R1 and R2 into "anchor the read paths".** The whole SPEC is the
  split.
- **Citing 68 / 84 / 5 as current.** They are dated host-local values; re-measure
  (spec.md §A.8).
- **Reading a process-existence probe as a session liveness verdict.** It identifies
  a process, and two of the probe-positive entries were unrelated system processes
  (spec.md §A.7).
- **Deleting orphan files "while we are in here".** Out of scope, and it is the
  other card's subject.
- **Landing S2 because S1 passed.** They share a SPEC, not a precondition.
- **Narrating the synthetic-fixture observation as a reproduction.** M2 exists to
  make that choice explicit.

---

## §H Cross-references

- `.moai/reports/t1058/pre-plan-measurement.md` — the evidence base, with
  `orphan-file-list.txt`, `orphan-file-entry-counts.txt`,
  `orphan-entry-liveness.txt`, `live-pid-identity.txt`
- `internal/session/anchor.go` — R1
- `internal/hook/cwd_changed_relocate.go` — R2
- `internal/session/registry.go` — the repaired write path (untouched here)
- `.claude/rules/moai/core/verification-claim-integrity.md` — the grading discipline
  §A.7 applies
