# M1 demographics — SPEC-SPECLINT-GATE-SIGNAL-001 (card t525)

Milestone M1 (REQ-SLGS-001): re-derivation of the SPEC lint finding distribution —
per-rule × severity × advisory-marking — on the then-current `origin/develop` tree.

## Measurement provenance

| Field | Value |
|---|---|
| Measured tree | `b6efc874f` (branch `WT-speclint-red`, after absorbing `origin/develop` `19cf21408`; absorb window `a849d99d2..19cf21408` carried 67 commits, **zero** touching `internal/spec/`) |
| Binary | **`go run ./cmd/moai` — tree build from this worktree** (per lead discipline 2026-09-08: installed `moai` binary e79c010b8 is 1069 commits behind origin/develop and is NOT used for any M1 verdict measurement) |
| Date | 2026-09-08 |
| Corpus | 804 directories under `.moai/specs/SPEC-*/` (`ls -d .moai/specs/SPEC-*/ | wc -l` → 804) |

**Binary provenance of earlier baselines (lead directive 2026-09-08)**: the pre-M1 baseline
measurements `4368@dd1439502` and `4379@dc8e10068` (4377 via `go run` at the latter tree) were
taken with a binary **1069 commits behind origin/develop** — superseded for verdict purposes by
the tree-build census below. The CI 5-consecutive-red record in spec.md §1 is CI-native (CI
builds from tree) and unaffected.

## Commands (verbatim) and observed outputs

```
$ go run ./cmd/moai spec lint
→ final line: "0 error(s), 4378 warning(s)" ; exit code 0 (RC_NONSTRICT=0)
$ go run ./cmd/moai spec lint --json > .moai/reports/t525/census-b6efc874f.json
→ exit code 0 (CENSUS_RC=0); jq 'length' → 4378 findings
$ go run ./cmd/moai spec lint --strict
→ see verdict.md (b) row — captured in this milestone
```

Aggregation used `jq` over the census JSON (note: `advisory` is `omitempty` in the Finding
struct — `internal/spec/lint.go:45` — so absent = false; aggregation reads `.advisory // false`).

## Census — per-rule × severity × advisory (tree `b6efc874f`)

| Rule code | Severity | Advisory | Count |
|---|---|---|---|
| CoverageIncomplete | warning | true | 3620 |
| ModalityMalformed | warning | true | 412 |
| MovingRefUnpinned | warning | true | 115 |
| StatusTransitionInvalid | warning | true | 102 |
| LegacyEARSKeyword | warning | true | 48 |
| MissingExclusions | warning | true | 27 |
| StatusGitConsistency | warning | true | 18 |
| FrontmatterInvalid | warning | true | 14 |
| StatusTokenUnrecognized | warning | true | 7 |
| InvalidREQID | warning | true | 6 |
| SyncSHASlotFormat | warning | true | 6 |
| **SpecsDirMissingSpecFile** | warning | **false** | **2** |
| OwnershipTransitionInvalid | warning | true | 1 |
| **Total** | warning 4378, error 0 | advisory=true 4376 / advisory=false **2** | **4378** |

Raw census: `census-b6efc874f.json` (committed alongside this file).

## Material findings against the card's premises

1. **The non-advisory standing stock is 2, not thousands.** The card's historical sample
   (tree `0b1e27877`, spec.md §1.1: CoverageIncomplete 3,588 · ModalityMalformed 412 · …)
   was an advisory-agnostic table — the advisory column did not exist in that output — and
   spec.md §1's "advisory 가 아닌 경고가 수천 건 서 있다" was an inference over it, not a
   measurement of the advisory split. Today's tree-build census measures the split directly:
   **exactly 2 non-advisory warnings** (the `SpecsDirMissingSpecFile` pair on
   SPEC-V3R4-CC2X-ADOPT-001/002), 4,376 advisory.
2. **Why the population is almost entirely advisory.** `applyEraDemotion`
   (`internal/spec/lint.go:345`) marks every warning on a *protected* SPEC advisory:
   protected = grandfather-era (era.go file-content classification — dates + progress.md
   markers, NOT git-dependent) OR terminal frontmatter status. Landed `dd644b5e0`
   (2026-07-21), i.e. before the historical sample tree. `StatusGitConsistency` is
   additionally inherently advisory (environment-dependent git-implied signal).
   The demotion cause is file-content-based — CI and local see the same split.
3. **Today's red is carried by exactly the M4 pair.** `SpecsDirMissingSpecFile` landed
   `85a783c9e` (2026-09-02, t365 — "surface SPEC-*/ directories with no spec.md instead of
   skipping them"); the rule is directory-level (no SPEC doc to attach era demotion to), so
   its 2 findings are non-advisory. The CI failure sequence recorded in spec.md §1 begins
   09-03 (run `25a3212a9`), one day after t365 landed — consistent with the pair being the
   carried red. (CI history itself is the spec's recorded evidence, run ids cited there;
   this row is the consistency reading, not a new CI measurement.)
4. **Corpus moved +10 since plan-time**: total warnings 4378 here vs 4368 at `dd1439502`
   (plan-time, binary-provenance caveat above) vs 4344 at CI `d4162b368` — every count is
   recorded against its own tree; none is frozen as a threshold (REQ-SLGS-003).

## Freeze discipline (REQ-SLGS-003)

No integer on this page is a threshold, constant, or test expectation. Every count is a
re-derivation at tree `b6efc874f` on 2026-09-08 and expires at the next corpus change.
