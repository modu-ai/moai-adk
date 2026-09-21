---
id: SPEC-SESSION-REGISTRY-READ-ANCHOR-001
title: "session registry READ paths — the two are not one change, and one of them loses live sessions"
version: "0.1.0"
status: draft
created: 2026-09-21
updated: 2026-09-21
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: internal/session
lifecycle: spec-anchored
tier: M
tags: "session-registry, read-path, worktree-anchor, disposal-guard, orphan-registry, visibility-loss"
---

# SPEC-SESSION-REGISTRY-READ-ANCHOR-001 — anchoring the read paths

## HISTORY

- 2026-09-21 · v0.1.0 · manager-spec · Initial authoring from card t1058. Every
  measured statement is attributed to `.moai/reports/t1058/pre-plan-measurement.md`
  by section (M1-M4, Gaps, Residual-risk) and to its four raw-evidence companions.
  No measurement is re-derived here and no figure appears here that is not in that
  record.

---

## §0 Governing principle [HARD]

> **Anchoring a read path does not neutralise orphan state. It removes that state
> from view — and what leaves the view includes live sessions.**

The card that produced this SPEC carried the opposite hypothesis. The measurement
that preceded this authoring reversed it for one of the two read paths and
confirmed it for the other. This SPEC exists to keep those two apart, because a
single prescription over both hides the reversal.

---

## §A Background

### A.1 Evidence base

The single evidence base is `.moai/reports/t1058/pre-plan-measurement.md`, measured
2026-09-21 on this host against tree `3dfae918a` (branch `WT-read-anchor`), with raw
companions `orphan-file-list.txt`, `orphan-file-entry-counts.txt`,
`orphan-entry-liveness.txt`, and `live-pid-identity.txt` beside it. Section
references below (`M1`, `M3-a`, `Gaps G3`) point into that file.

### A.2 The card's premise, and what measurement did to it

The card hypothesised: *anchoring the read path may make the orphan
`active-sessions.json` files harmless, which would make the A6 deletion card
unnecessary.*

Measurement (M3) establishes that this does not hold for the first read path:

- The orphan entries are **disjoint from the primary registry** — zero of the
  orphan entries measured also exist in the primary file. The orphan set is not a
  copy of primary; it is a separate population.
- A non-zero number of those orphan-only entries belong to **live `claude`
  processes** (M3-a). Anchoring the first read path removes them from the
  disposal guard's input.
- The resulting loss is **the same loss A6's file deletion would cause** (M3-b).

### A.3 The asymmetry — the two read paths have opposite risk profiles

| Path | Coordinate | What it reads | Risk of anchoring |
|---|---|---|---|
| R1 | `internal/session/anchor.go:65` (`LiveAnchoredSessions`) | `DefaultRegistryPath` joined to two roots — `treePath` and `callerProjectRoot()` | **Loses** the live orphan-only entries from the disposal guard's input (M3) |
| R2 | `internal/hook/cwd_changed_relocate.go:100` (`findRegistryUpwardFrom`, driven by `relocateSessionCwd`) | walks up from a candidate working directory, stops at the **first** `active-sessions.json` found | **Pure repair** — stopping at an orphan file is a net loss with nothing on the other side (M4) |

The card treats the two as one change. They are not, and the plan scopes them
apart (§C).

### A.4 R2's failure mechanism, and its evidence grade

`relocateSessionCwd` iterates candidate working directories
(`input.OldCwd`, `input.NewCwd`, `input.CWD`). When the registry found above a
candidate does not contain the session, the loop continues
(`internal/hook/cwd_changed_relocate.go:55-57`) — and that continuation advances to
the **next candidate working directory**, not to the parent directory. Where all
three candidates sit inside the same linked worktree, all three resolve to the same
orphan file, and there is structurally no path to the primary registry.

[HARD] **Evidence grade, carried verbatim from the record and not upgraded.** The
upward-walk stall was observed by agent-14 on a **synthetic fixture with a positive
control**, plus this lane's code reading (M1). It has **NOT** been reproduced
against the real orphan files (Gaps G3). M1 supplied the code basis for the
mechanism; it is not a reproduction against the live population.

### A.5 The root, and why neither A5 nor A6 is it

The root was the **unanchored WRITE path**, and it is already repaired:
`RegistryPathFor` (`internal/session/registry.go:62`) resolves through
`stateanchor.FromDirectory` to the repository's one anchor root, with the regression
guard at `internal/session/registry_path_anchor_test.go` (M4).

The orphan files are historical residue of the pre-repair regime. A5 (anchoring the
read paths) does not make that residue harmless — it makes it invisible. A6
(deleting the files) causes the same loss by deletion. **Neither is the root of the
other**; both rest on the same unmeasured premise — whether a still-live orphan-only
entry may be discarded rather than migrated (M4).

### A.6 Consumers of R1 — what changes if it is anchored

`LiveAnchoredSessions` is consumed by the worktree disposal guard at four
coordinates, verified in this tree:

- `internal/cli/worktree/remove.go:51`
- `internal/cli/worktree/done.go:77`
- `internal/cli/worktree/done.go:176`
- `internal/session/anchor_lock.go:111`

Anchoring R1 flips worktrees currently judged *"a live session is anchored here →
refuse disposal"* to *"free"* (M3-b). That is a regression of the disposal guard,
and the guard is what stands between a live lane and the destruction of an unpushed
branch's only copy.

### A.7 What is unmeasured

These grades are carried from the record verbatim. None is upgraded here, and none
appears anywhere in this SPEC as a claim.

- **Windows `stateanchor` behaviour is UNMEASURED, not absent.**
  `internal/stateanchor/stateanchor.go` was read in full and carries **no
  `runtime.GOOS` branching**, so Windows behaviour rides entirely on
  `git.ResolveGitDirs` path formats. That is a **code reading, not an execution
  measurement** (Gaps G2).
- **R2's stall is unreproduced against the real orphan files** (Gaps G3, §A.4).
- **A process-existence probe identifies a PROCESS, not a session** (Gaps G4). Two
  of the probe-positive entries were pid-collision false positives against unrelated
  system processes, removed by process-identity inspection (M3-a). One pid is
  recorded in two different worktrees under two session ids, so **at most one of
  those recorded working directories is where it actually sits** — which of the two
  is not measured.
- **One verification command was REFUSED by the worktree-isolation guard** during
  measurement and was re-run in a split literal-path form; the measurement itself
  was performed and no substitute inference was used (Gaps G1).

### A.8 [HARD] Every population figure is dated and host-local

The orphan-file count, the orphan-entry count, and the live-entry count in the
record are values for **this host, at that hour**. The population moved between the
card's authoring and the measurement a day later. A criterion in this SPEC that
depends on a count therefore **re-measures at execution time** and never cites the
record's figures as the current state (Residual-risk; REQ-RAR-006, AC-RAR-005).

---

## §B Requirements (GEARS)

### Scope separation

- **REQ-RAR-001** — Where the R2 read path is anchored, the change shall land
  independently of any R1 change, in its own commit with its own verification, so
  that the pure repair is not held behind R1's migration precondition.

### R2 — the pure repair

- **REQ-RAR-002** — When a working-directory-change relocation resolves a session's registry, the resolution shall reach the repository's primary registry
  rather than stopping at the first `active-sessions.json` encountered above a
  candidate working directory.
- **REQ-RAR-003** — When every candidate working directory lies inside one linked worktree carrying its own registry file, the relocation shall still reach the primary registry
  rather than stopping inside that worktree.
- **REQ-RAR-004** — Where the R2 resolution is changed, its fail-open behaviour shall be preserved:
  an absent, unreadable, or entry-less registry shall leave every record untouched
  and shall not fail the hook.

### R1 — gated on migration

- **REQ-RAR-005** — Where the R1 read path is anchored to a single registry, the change shall not land
  until either the still-live orphan-only entries have been migrated into that
  registry, or the operator has recorded a decision to discard them.
- **REQ-RAR-006** — When an R1 change is proposed for landing, the proposer shall
  re-measure the orphan population and its live-entry count against the tree and
  host at that time, and shall not cite the 2026-09-21 figures as the current
  state.
- **REQ-RAR-007** — Where a liveness claim about a registry entry is made, the claim shall identify the owning process
  rather than resting on a process-existence probe alone.

### Unwanted behaviour — preserved properties

- **REQ-RAR-008** — The worktree disposal guard shall not report a worktree as free
  while a live session is anchored inside it, whichever registry that session's
  entry currently lives in.
- **REQ-RAR-009** — The write-path anchoring carried by `RegistryPathFor` shall not be modified by this SPEC,
  and neither shall its regression guard.
- **REQ-RAR-010** — The orphan `active-sessions.json` files shall not be deleted by
  this SPEC.
- **REQ-RAR-011** — Where Windows behaviour of the anchor resolution is referenced, the reference shall be recorded as unmeasured
  and shall not be asserted as either absent or working.
- **REQ-RAR-012** — The disposition of the A6 orphan-file deletion card shall not be
  decided by this SPEC.

---

## §C Scope split

Two scopes, deliberately separable. The ordering below is by **decision
reversibility**, not by preference: S2 carries an irreversible loss and an
undecided migration question, S1 carries neither.

| Scope | Subject | Blocking precondition | Reversibility |
|---|---|---|---|
| **S1** | R2 — reach the primary registry instead of the first file found | none | a behaviour change with a test; revertible |
| **S2** | R1 — anchor `LiveAnchoredSessions` | migration of live orphan-only entries, or a recorded operator decision to discard them (REQ-RAR-005) | landing it un-gated destroys unpushed work through the disposal guard |

S1 may land alone. S2 may not land before its precondition is satisfied, and
whether S2 lands at all is an operator decision at the Implementation Kickoff
Approval gate — this SPEC presents it, it does not select it.

---

## §D Acceptance criteria

Full criteria live in `acceptance.md`. Two are load-bearing enough to restate.

[HARD] **AC-RAR-005 (re-measurement) sits ahead of any S2 landing.** It is
satisfied only by a measurement taken against the tree and host at landing time. It
**cannot** be satisfied by the 2026-09-21 record, by reading
`internal/session/anchor.go`, or by any figure in this document.

[HARD] **AC-RAR-008 (disposal guard preserved)** is the criterion that makes the
S1/S2 split observable rather than rhetorical: a change that satisfies every other
criterion while flipping a live-anchored worktree to free has failed.

---

## §E Exclusions

### Out of Scope — orphan-file disposition

- **Deleting the orphan `active-sessions.json` files is NOT performed by this
  SPEC** (REQ-RAR-010). That is the A6 card's subject.
- **Migrating live orphan-only entries into the primary registry is NOT performed
  by this SPEC** unless the plan is deliberately extended to include it. Migration
  is named here as S2's blocking **precondition** (REQ-RAR-005), not as a
  deliverable this SPEC commits to.
- **The fate of the A6 card is an operator queue decision this SPEC does not make**
  (REQ-RAR-012). Measurement established only that A5 and A6 are not root and
  symptom of each other; it did not establish what should happen to either card.

### Out of Scope — already repaired, deliberately untouched

- The WRITE-path anchoring (`RegistryPathFor`, `internal/session/registry.go:62`)
  and its regression guard. This SPEC consumes both unchanged (REQ-RAR-009).
- The `stateanchor` resolution seam itself. This SPEC calls it; it does not alter
  it.

### Out of Scope — unmeasured, therefore neither claimed nor acted on

- **Windows behaviour of the anchor resolution** (Gaps G2, §A.7). It may appear in
  a report as an explicitly-unmeasured note; it must never appear as a claim.
- **Reproduction of R2's upward-walk stall against the real orphan population**
  (Gaps G3). This SPEC rests on the synthetic-fixture observation plus the code
  reading, at that grade.
- **Which of the two recorded working directories a doubly-recorded process
  actually occupies** (Gaps G4). Not measured, not inferred here.
- **The dated population figures as a standing baseline.** They are a measurement
  of one host at one hour and are re-measured rather than cited (§A.8).
