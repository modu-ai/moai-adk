---
id: SPEC-SESSION-REGISTRY-READ-ANCHOR-001
title: "acceptance criteria — session registry READ-path anchoring"
version: "0.1.0"
created: 2026-09-21
updated: 2026-09-21
author: manager-spec
---

# SPEC-SESSION-REGISTRY-READ-ANCHOR-001 — acceptance criteria

Every criterion below is binary and mechanically verifiable. **None of them
consumes a stored population count** — where a count is involved, the criterion
requires a measurement taken at execution time (spec.md §A.8).

---

## §A Scope-separation criteria

### AC-RAR-001 — S1 lands independently of S2

Given the R2 repair is complete and the S2 migration decision is still open,
When the S1 change is prepared for landing,
Then the S1 change is a commit whose diff touches no R1 coordinate
(`internal/session/anchor.go`) and whose verification passes on its own,
And the landing is not blocked by the absence of an S2 decision.

Mechanical check: the S1 commit's changed-file list, read from the commit itself,
contains no R1 coordinate; the S1 verification command exits 0 on that commit.

Maps REQ-RAR-001.

---

## §B R2 criteria — the pure repair

### AC-RAR-002 — a relocation whose candidates all sit inside one worktree reaches the primary registry

Given a repository whose linked worktree carries its own `active-sessions.json`,
and a session whose entry lives only in the repository's primary registry,
When a working-directory-change relocation runs with every candidate working
directory inside that worktree,
Then the session's entry in the primary registry carries the new working directory,
And the worktree-local file is unchanged byte-for-byte.

Mechanical check: a test constructing that fixture, asserting the primary entry's
working directory changed and the worktree-local file's content did not. A positive
control asserts the test fails when the repair is reverted — a passing test over an
unexercised path is not evidence.

Maps REQ-RAR-002, REQ-RAR-003.

### AC-RAR-003 — fail-open is preserved

Given the registry resolved for a relocation is absent, unreadable, or contains no
matching session,
When the relocation runs,
Then every registry file involved is unchanged byte-for-byte,
And the hook exits without error.

Mechanical check: three test cases (absent file, unreadable file, entry-less file),
each asserting file content equality before and after and a nil error. Each case
carries a mutation control: with the fail-open branch removed, the case fails.

Maps REQ-RAR-004.

### AC-RAR-004 — the existing relocation behaviour is unchanged where it already worked

Given a session whose entry lives in the registry the current implementation
already finds,
When the relocation runs after the S1 change,
Then the resulting registry content is identical to the content the
pre-change implementation produces for the same input.

Mechanical check: the pre-existing tests in `internal/hook/` covering relocation
pass unmodified. A test file edited to accommodate the change is a failure of this
criterion unless the edit is itself justified in the run-phase evidence.

Maps REQ-RAR-002.

---

## §C R1 criteria — gated

### AC-RAR-005 — [HARD] no S2 landing without a recorded decision

Given the S2 change (R1 anchoring) is prepared,
When it is proposed for landing,
Then a recorded operator decision exists naming branch B1 (migrate first), B2
(discard, recorded), or B3 (do not land S2), taken at the Implementation Kickoff
Approval gate,
And where the decision is B3, no S2 change lands at all.

[HARD] This criterion **cannot** be satisfied by the 2026-09-21 measurement record,
by reading `internal/session/anchor.go`, or by any figure in spec.md. It is
satisfied only by the recorded decision itself.

Maps REQ-RAR-005.

### AC-RAR-006 — [HARD] the population is re-measured at landing time

Given an S2 change is proposed for landing,
When the proposal is made,
Then the orphan population and its live-entry count have been measured against the
tree and host at that time, and the measurement names its command, its verbatim
output, and the tree it ran against,
And no figure from the 2026-09-21 record is presented as the current state.

Mechanical check: the run-phase evidence carries a measurement whose recorded tree
identifier equals the tree the S2 change is landing on. A measurement attributed to
any other tree fails this criterion.

Maps REQ-RAR-006.

### AC-RAR-007 — a liveness claim names its process

Given a claim that a registry entry is live is made in any run-phase or sync-phase
artifact,
When that claim is read,
Then it names the owning process (its command line or equivalent identity), not
only the result of a process-existence probe.

Mechanical check: each liveness claim in the evidence has an adjacent
process-identity record. A claim resting on a probe result alone fails.

Maps REQ-RAR-007.

---

## §D Preserved-behaviour criteria — apply to every scope

### AC-RAR-008 — [HARD] the disposal guard does not free a live-anchored worktree

Given a worktree with a live session anchored inside it, whose registry entry is
reachable by the read path after the change,
When the disposal guard runs at each of its four consumer coordinates
(`internal/cli/worktree/remove.go`, `internal/cli/worktree/done.go` twice,
`internal/session/anchor_lock.go`),
Then each refuses disposal.

Mechanical check: a test fixture per coordinate asserting refusal, plus the
converse control — a worktree with no live session anchored is reported free, so
the test is not passing by refusing everything.

Maps REQ-RAR-008.

### AC-RAR-009 — no orphan file is modified or deleted

Given the set of orphan `active-sessions.json` files present before the change,
When the change lands and its verification runs,
Then every file in that set still exists with identical content.

Mechanical check: a content-hash comparison over the orphan file list taken
immediately before and immediately after, run at execution time — not against the
stored 2026-09-21 list. **Content hash, not modification time or size**: a rewrite
with identical content is not a state change, and mtime alone reports one that did
not happen.

Maps REQ-RAR-010.

### AC-RAR-010 — the write-path guard is untouched

Given `internal/session/registry_path_anchor_test.go`,
When the change lands,
Then that file is unchanged and its test passes.

Mechanical check: the commit's changed-file list does not contain it; the test
command exits 0.

Maps REQ-RAR-009.

### AC-RAR-011 — unmeasured axes are recorded as unmeasured

Given any run-phase or sync-phase artifact referencing Windows behaviour of the
anchor resolution, R2's reproduction against the real orphan population, or the
working directory a doubly-recorded process occupies,
When that reference is read,
Then it is marked as unmeasured,
And it is not stated as absent, as working, or as established.

Maps REQ-RAR-011.

### AC-RAR-012 — the A6 card is not decided here

Given the SPEC's artifacts and its run-phase and sync-phase evidence,
When they are read,
Then none of them closes, drops, reopens, or reprioritises the A6 orphan-file
deletion card,
And any statement about A6 is a finding presented to the operator, never a queue
mutation.

Mechanical check: no `moai gtd` mutation against the A6 card appears in this SPEC's
command record.

Maps REQ-RAR-012.

---

## §D.1 Traceability (REQ → AC)

| REQ | AC |
|---|---|
| REQ-RAR-001 | AC-RAR-001 |
| REQ-RAR-002 | AC-RAR-002, AC-RAR-004 |
| REQ-RAR-003 | AC-RAR-002 |
| REQ-RAR-004 | AC-RAR-003 |
| REQ-RAR-005 | AC-RAR-005 |
| REQ-RAR-006 | AC-RAR-006 |
| REQ-RAR-007 | AC-RAR-007 |
| REQ-RAR-008 | AC-RAR-008 |
| REQ-RAR-009 | AC-RAR-010 |
| REQ-RAR-010 | AC-RAR-009 |
| REQ-RAR-011 | AC-RAR-011 |
| REQ-RAR-012 | AC-RAR-012 |

---

## §E Edge cases

- **A repository with no linked worktree.** R2's resolution must behave exactly as
  before; there is no orphan file to stop at. Covered by AC-RAR-004.
- **A directory outside any repository.** The existing home-directory boundary must
  continue to stop the walk — a session working outside a checkout must not have its
  entry relocated into global state. Covered by AC-RAR-003's absent-registry case.
- **An orphan file that is a copy of primary rather than disjoint.** Measurement
  found the two populations disjoint, but that is a property of one host at one
  hour, not an invariant. A criterion that assumes disjointness would rest on a
  dated figure; none above does.
- **Two entries for one process in two worktrees.** At most one is correct. No
  criterion above treats such an entry as authoritative about location.

---

## §F Quality gates

- `go test ./internal/hook/... ./internal/session/...` exits 0 for S1.
- `go test ./internal/session/... ./internal/cli/worktree/...` exits 0 for S2.
- `go vet` over the changed packages exits 0.
- Every new test carries a mutation control: reverting the change under test makes
  it fail. A test that passes both with and without the change proves nothing.

---

## §G Definition of Done

- S1 landed, its test and mutation control recorded, its verification output
  exported to `.moai/reports/t1058/`.
- The S2 decision recorded, and S2 either landed under a satisfied AC-RAR-005 and
  AC-RAR-006, or explicitly not landed under branch B3.
- Every preserved-behaviour criterion in §D verified, each with its command and
  verbatim output.
- Every unmeasured axis named as unmeasured in the closing evidence, at the grade
  spec.md §A.7 carries — not upgraded.
