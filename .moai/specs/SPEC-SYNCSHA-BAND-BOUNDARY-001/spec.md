---
id: SPEC-SYNCSHA-BAND-BOUNDARY-001
title: "sync_commit_sha length-band boundary fixtures"
version: "0.3.0"
status: completed
created: 2026-08-31
updated: 2026-08-31
author: manager-spec
priority: P2
phase: "v3.2.0"
module: "internal/spec"
lifecycle: spec-anchored
tags: "syncsha, lint, fixtures, mutation-testing"
tier: S
---

# SPEC-SYNCSHA-BAND-BOUNDARY-001: sync_commit_sha length-band boundary fixtures

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-08-31 | Initial draft (card t380). Closes the surviving-mutant debt D1 recorded in SPEC-SYNC-SHA-SLOT-FORMAT-001 §E.4, and corrects that card's diagnosis of its own cause. |
| 0.2.0 | 2026-08-31 | Plan-audit amendment (PASS-WITH-DEBT 0.9216, `.moai/reports/t380/plan-audit.md`). Operator rulings on the two blocking defects: the outside-band cases get their own test function (audit D1, option (a)), and AC-SBB-005 / AC-SBB-008 merge to bring the criterion count to the Tier S ceiling of 8 (audit D2). REQ-SBB-005 rewritten to match. Alphabet residual declared (D3); `priority` corrected (D5); one citation and one timing label corrected (D6, D7). |

---

## §A. Context

SPEC-SYNC-SHA-SLOT-FORMAT-001 (card t299) landed a `sync_commit_sha` value
grammar whose SHA production is `SHA := [0-9a-fA-F]{7,40}`, one shared shape
predicate `isCommitSHAToken` (`internal/spec/syncsha.go:107`, over
`commitSHATokenPattern` at `:63`), and two consumers of it:

| Side | Consumer | Coordinate (this tree, `3f03d9c36`) |
|---|---|---|
| write | `needsSHABackfill` = `!isCommitSHAToken(syncSHAValueToken(value))` | `internal/spec/closer.go:421` |
| read | `SyncSHASlotFormatRule.Check` | `internal/spec/lint_syncsha.go:83`, calling the predicate at `:103` |

t299 recorded a SURVIVING mutant as debt D1: inlining `^[0-9a-fA-F]{8,40}$` in
`lint_syncsha.go` in place of the `isCommitSHAToken` call leaves
`go test ./internal/spec/... -count=1` GREEN (exit 0, 34.439s) while the two
gates then disagree about a 7-character SHA. That figure is quoted from t299's
§E.4; it was not re-measured here.

### §A.1 The originating card's diagnosis was wrong

t299 diagnosed D1 as **"AC-SSF-007 has no CI carrier"**. This SPEC states plainly
that the diagnosis is wrong, rather than correcting it silently: a silent
correction leaves the next reader hunting for a missing CI carrier that was never
the problem, and the true cause — a fixture-corpus gap — would stay open while
the wrong repair (automating a grep) was built.

The measured cause:

```
$ grep -n 'sync_commit_sha' internal/spec/testdata/syncsha/sha-short/progress.md
9:sync_commit_sha: a6bbbf82b
```

The fixture named `sha-short` holds **nine** hex characters, not seven. Measured
across all four SHA-bearing fixtures in this tree:

```
$ awk '/^sync_commit_sha:/ {v=$2; gsub(/"/,"",v); print FILENAME, v, length(v)}' \
    internal/spec/testdata/syncsha/{sha-short,sha-full,sha-quoted,sha-annotated}/progress.md
internal/spec/testdata/syncsha/sha-short/progress.md     a6bbbf82b                                9
internal/spec/testdata/syncsha/sha-full/progress.md      a6bbbf82b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6 40
internal/spec/testdata/syncsha/sha-quoted/progress.md    a6bbbf82b                                9
internal/spec/testdata/syncsha/sha-annotated/progress.md a6bbbf82b                                9
```

The band is `[7,40]`. The fixtures sit at 9 / 40 / 9 / 9. **No fixture sits at
the inside edge of the floor**, so a mutant that moves the floor by ONE
(`7 → 8`) is invisible to every fixture. The existing test's own comment
(`lint_syncsha_test.go:123-124`) anticipates a narrowing to `{40}` — which nine
characters DOES catch — so the anticipated mutation is caught and the one-step
mutation is not.

### §A.2 The band boundary is covered at the predicate level, not at the rule level

Measured in this tree: `TestIsCommitSHAToken_LengthBand`
(`internal/spec/syncsha_test.go:52`) already carries all four boundary rows —
`"a6bbbf8"` (7, true), the 40-character value (true), `"a6bbbf"` (6, false), and
the 41-character value (false).

That table does not save the rule. The surviving mutant does not change
`commitSHATokenPattern`; it **bypasses** the predicate by inlining a second copy
of the hex test at the call site, so the predicate's own table stays green by
construction. Coverage of the band therefore exists at
`isCommitSHAToken` and is absent at `SyncSHASlotFormatRule.Check` — and the rule
is where the mutation lands.

The plan audit reproduced this mechanism independently: with the full fixture set
installed, its mutant 1 left `TestIsCommitSHAToken_LengthBand` green while turning
the rule-level criterion red (`.moai/reports/t380/plan-audit.md`, § Central claim).

### §A.3 Probe (performed and fully reverted at plan phase)

Evidence: `.moai/reports/t380/probe-01..05`; `probe-05-mutant-red.txt` carries the
verbatim FAIL. Adding one 7-character fixture (value `19b6f76`) and registering it
in the EXISTING `TestSyncSHASlot_SilentOnSHA` case list turns the surviving mutant
RED:

```
--- FAIL: TestSyncSHASlot_SilentOnSHA (1.35s)
    lint_syncsha_test.go:128: sha-min7: expected 0 findings on a well-formed SHA, got 1: [...]
```

The mutant was then reverted and the suite observed GREEN again.

**Timing label**: at the moment that probe was taken — before these SPEC artifacts
were authored — `git status --short` in this tree reported exactly one entry,
`?? .moai/reports/t380/`. Re-measured after authoring, at `3f03d9c36`, it reports
two, the second being `?? .moai/specs/SPEC-SYNCSHA-BAND-BOUNDARY-001/`. Both
figures describe the same fact: the tree carries no source change.

### §A.4 Independent reproduction (plan audit)

The plan audit planted all four band mutants under its own hands with the full
fixture set installed, and every direction turned red — mutant 1 killed by
`sha-min7`, mutant 2 by `sha-full`, mutant 3 by `sha-below6`, mutant 4 by
`sha-above41`. Baseline `ok … 3.201s` (exit 0); everything reverted, `git diff
--stat` empty. Attributed to `.moai/reports/t380/plan-audit.md`, not measured here.

---

## §B. Requirements (GEARS)

- **REQ-SBB-001** — The fixture corpus under `internal/spec/testdata/syncsha/`
  shall carry a `sync_commit_sha` value at each inside edge of the SHA length
  band: a 7-character token and a 40-character token.

- **REQ-SBB-002** — The fixture corpus shall carry a `sync_commit_sha` value one
  character outside each edge of the SHA length band: a 6-character token and a
  41-character token.

- **When** `SyncSHASlotFormatRule` evaluates a fixture whose token sits at an
  inside edge of the band, the rule shall produce zero findings
  (**REQ-SBB-003**).

- **When** `SyncSHASlotFormatRule` evaluates a fixture whose token sits one
  character outside an edge of the band, the rule shall produce exactly one
  finding (**REQ-SBB-004**).

- **REQ-SBB-005** — The boundary fixtures shall be exercised through the existing
  shared helper `syncSHAFindings` and through an existing criterion wherever that
  criterion's expected outcome matches the fixture's; **where** an existing
  function's name is a load-bearing token in another SPEC's live acceptance
  criterion, the delivery shall introduce a separate test function rather than
  overload it, and that function shall not duplicate an existing function's
  assertion body.

- **REQ-SBB-006** — The four boundary fixtures shall together make BOTH
  directions of BOTH edges of the length band observable, so that a band-widening
  mutant and a band-narrowing mutant each turn at least one criterion red.

- **REQ-SBB-007** — The delivery shall not modify `isCommitSHAToken`,
  `syncSHAValueToken`, `isSyncSHAPlaceholder`, `commitSHATokenPattern`,
  `syncSHAPlaceholderPattern`, or `needsSHABackfill`.

- **REQ-SBB-008** — The delivery shall not weaken, restate, or re-scope AC-SSF-007
  of SPEC-SYNC-SHA-SLOT-FORMAT-001, nor edit
  `.moai/specs/SPEC-SYNC-SHA-SLOT-FORMAT-001/acceptance.md`.

### §B.1 Why four fixtures and not one — per-fixture mutation mapping

Each fixture catches a mutation the others do not. The 7-character and
6-character fixtures catch **opposite directions** of the same edge: with only
one of them present, a mutant simply moves the other way and stays invisible.
This is the whole reason the deliverable is four fixtures rather than one.

| Fixture | Token length | Position | Expected findings | Mutation it catches |
|---|---|---|---|---|
| `sha-min7` (new) | 7 | inside the floor | 0 | floor **narrowed**: `{7,40}` → `{8,40}` |
| `sha-full` (existing, reused) | 40 | inside the ceiling | 0 | ceiling **narrowed**: `{7,40}` → `{7,39}` |
| `sha-below6` (new) | 6 | one below the floor | 1 | floor **widened**: `{7,40}` → `{6,40}` |
| `sha-above41` (new) | 41 | one above the ceiling | 1 | ceiling **widened**: `{7,40}` → `{7,41}` |

`sha-full` already holds a 40-character value (measured in §A.1) and is already
registered in `TestSyncSHASlot_SilentOnSHA` (`lint_syncsha_test.go:126`). It is
**reused, not duplicated**: adding a second 40-character fixture would add a file
and catch nothing the existing one does not.

Read as a pair per edge: a narrowing mutant is caught by the INSIDE fixture going
from 0 findings to 1; a widening mutant is caught by the OUTSIDE fixture going
from 1 finding to 0. Neither direction is observable from one side alone.

### §B.2 Why the outside-band cases get their own test function

`TestSyncSHASlot_SilentOnSHA` asserts zero findings, so the outside-band fixtures
cannot join its case list. The apparently obvious home —
`TestSyncSHASlot_FlagsProse` — is unavailable for a reason found by the plan
audit: that function's name is a load-bearing token in a live criterion of a
`completed` SPEC. `.moai/specs/SPEC-SYNC-SHA-SLOT-FORMAT-001/acceptance.md:56`
reads ``**when** `go test ./internal/spec/ -run TestSyncSHASlot_FlagsProse` runs``,
so overloading the function means a failure in it no longer identifies AC-SSF-001,
and renaming it would falsify that command string.

Operator ruling: the outside-band cases get their own function,
`TestSyncSHASlot_FlagsOutOfBand`; `TestSyncSHASlot_FlagsProse` keeps only the
prose case, name and behavior untouched. The constraint that is relaxed to make
this possible is **this card's own** delivery-shape criterion — which previously
pinned the `func Test` count at 3 — not another SPEC's criterion. That asymmetry
is the whole reason this option is cheaper than renaming: amending a completed
SPEC's judgment criterion costs a great deal more than amending a draft's.

---

## §C. Exclusions

### Out of Scope — the shared predicate and the write-side gate

- `isCommitSHAToken`, `syncSHAValueToken`, `isSyncSHAPlaceholder`
  (`internal/spec/syncsha.go`) are NOT modified. No production source change is
  expected at all: the deliverable is test fixtures plus test-side case lists.
- `needsSHABackfill` (`internal/spec/closer.go:421`) is NOT modified.
- `commitSHATokenPattern` and `syncSHAPlaceholderPattern` are NOT modified. In
  particular the band `{7,40}` is not changed — this card observes the band, it
  does not redefine it.

### Out of Scope — the alphabet clause of the §D.1 grammar

- This card closes the **length-band** slice of the rule-level coverage gap
  diagnosed in §A.2, and no other slice of it. The **alphabet** clause remains
  uncovered at the rule level, and that residual is declared here rather than left
  to be discovered later. Evidence, attributed to the plan audit
  (`.moai/reports/t380/plan-audit.md` D3) and not measured here: with the full
  four-fixture set installed, an alphabet-narrowing rule-level mutant
  `^[0-9a-f]{7,40}$` — dropping uppercase hex — leaves
  `go test ./internal/spec/ -run 'TestSyncSHASlot' -count=1` at `ok … 3.082s`,
  exit 0, because every fixture value is lowercase. The predicate-level table does
  carry an uppercase row (`internal/spec/syncsha_test.go:58`,
  `{"A6BBBF82B", true}`) — which is precisely the layer such a mutant bypasses.
  Closing the alphabet slice is a different axis and would widen this card's
  scope; it is not attempted here.
- Observed bonus, also attributed to the plan audit and **not** claimed as a
  designed property of this fixture set: an anchor-dropping mutant
  (`[0-9a-fA-F]{7,40}`, no `^`/`$`) is killed incidentally by `sha-above41`. The
  fixture set is therefore stronger than REQ-SBB-006 claims for it. No requirement
  or criterion here rests on that behavior.

### Out of Scope — AC-SSF-007 of SPEC-SYNC-SHA-SLOT-FORMAT-001

- AC-SSF-007's grep criterion **stays as it is**. It is not vacuous: it does turn
  red against the mutant by the means it declares. What it lacks is CI
  automation, which is a **different gap** from the one this card closes. This
  SPEC preserves that distinction rather than blurring it — closing the
  fixture-corpus gap neither satisfies nor supersedes the automation gap.
- `.moai/specs/SPEC-SYNC-SHA-SLOT-FORMAT-001/acceptance.md` is not edited, which
  is what keeps AC-SSF-001's command string true (§B.2).

### Out of Scope — a source-scan guard test

- No test that greps `internal/spec/*.go` for a second copy of the hex pattern.
  Rejected: it moves a grep into a test, reads text rather than behavior, and
  leaks the moment a symbol is renamed.

### Out of Scope — structural unification of the two gates

- The two gates are NOT merged behind a single decision function. Rejected on two
  measured grounds:
  - In Go, inlining a regex is expressible under any design, so "the mutant
    becomes inexpressible" does not hold.
  - The two gates must **disagree on placeholders by design**: REQ-SSF-003 has the
    write side treat a placeholder as owing a backfill, while REQ-SSF-005 has the
    read side stay silent on it (`lint_syncsha.go:106-113`,
    `syncsha.go:111-119`). Merging their discriminators would reopen the blind
    spot t299 closed. The shape classifier they share is already single.

### Out of Scope — normalizer and era classification

- `cleanFieldValue` is not touched, and era classification is not touched
  (unchanged from REQ-SSF-008 / AC-SSF-009: the era heuristics H-3/H-4 depend on
  the normalizer's current behavior).

### Out of Scope — files owned by a concurrent lane

- `internal/template/evidence_citation_guard_test.go`,
  `.claude/rules/moai/core/agent-common-protocol.md`,
  `.claude/rules/moai/core/agent-common-protocol-reference.md`,
  `.claude/agents/moai/manager-lead.md`, and their template mirrors / `.codex`
  `.toml` emissions are being edited by another lane and are not touched here.

---

## §D. Traceability

| Requirement | Acceptance criterion |
|---|---|
| REQ-SBB-001 | AC-SBB-001, AC-SBB-002 |
| REQ-SBB-002 | AC-SBB-003, AC-SBB-004 |
| REQ-SBB-003 | AC-SBB-001, AC-SBB-002 |
| REQ-SBB-004 | AC-SBB-003, AC-SBB-004 |
| REQ-SBB-005 | AC-SBB-005 |
| REQ-SBB-006 | AC-SBB-006, AC-SBB-007 |
| REQ-SBB-007 | AC-SBB-005 |
| REQ-SBB-008 | AC-SBB-008 |
