# SPEC-SYNCSHA-BAND-BOUNDARY-001 — Implementation Plan

Tier S. Card t380, worktree `.claude/worktrees/t380`, branch
`WT-shared-predicate-guard`, base `3f03d9c36` (= `origin/develop` at dispatch).

Milestones are ordered by **decision reversibility**: the choices most likely to
change on review come first, the mechanical steps last.

---

## §A. Context

See `spec.md` §A. In one line: the band `[7,40]` has no fixture at either inside
edge, so a one-step mutation of an edge is invisible; t299 misattributed that to a
missing CI carrier for AC-SSF-007.

Nothing here needs a new decision about the grammar. The band, the predicate, and
the two gates all stay exactly as they are.

---

## §B. Known issues carried in

- **AC-SSF-007 has no CI automation.** True, and NOT what this card fixes
  (`spec.md` §C). It stays open after this card closes.
- **Fixtures share the SPEC id `SPEC-SSFA-001`.** Deliberate; the helper lints one
  fixture at a time. New fixtures must keep the same id, or
  `DuplicateSPECIDRule` behavior changes for reasons unrelated to the band.
- **Era demotion is a live trap.** A fixture that classifies as grandfathered or
  terminal makes its findings advisory and quietly changes what the criteria
  measure. `acceptance.md` §A carries the precondition; the flag-expecting
  criteria assert `Advisory == false` rather than trusting it.

---

## §C. Pre-flight

Measured in this tree before authoring; re-confirm at run-phase entry:

```
$ git rev-parse --short HEAD                → 3f03d9c36
$ git status --short                        → "?? .moai/reports/t380/" and
                                              "?? .moai/specs/SPEC-SYNCSHA-BAND-BOUNDARY-001/"
                                              (re-measured AFTER authoring these
                                              artifacts; before authoring it was
                                              the reports entry alone)
$ go test ./internal/spec/ -run 'TestSyncSHASlot|TestIsCommitSHAToken' -count=1
  ok  github.com/modu-ai/moai-adk/internal/spec  2.496s   (exit 0)
$ grep -c '^func Test' internal/spec/lint_syncsha_test.go   → 3
```

The `func Test` count of 3 is the pre-work baseline, **not** a constraint: the
M2 ruling raises it to 4 (see AC-SBB-005).

---

## §D. Constraints

- No production source change (`spec.md` REQ-SBB-007). The four planted mutants
  are transient and reverted.
- Do not touch the concurrent lane's files listed in `spec.md` §C.
- Lane-local verification only: `go test ./internal/spec/... -count=1`. Do NOT run
  the full suite locally.
- Every measurement recorded with its command; no figure carried from t299 without
  saying so.

---

## §E. Milestones

### M1 — Fixture value choice (most reversible decision; review this first)

Decide the four token values. Proposal, with the reasoning a reviewer would want
to disagree with:

| Fixture | Value | Length | Why this value |
|---|---|---|---|
| `sha-min7` | `19b6f76` | 7 | Real prefix of a commit in this repo's history; the probe already used it and observed the mutant red with it. |
| `sha-full` | *(existing, unchanged)* | 40 | Already at the ceiling's inside edge — reused, not duplicated. |
| `sha-below6` | `19b6f7` | 6 | The same prefix one character shorter, so the ONLY difference from `sha-min7` is length. A different alphabet or a different word would leave two variables moving at once. |
| `sha-above41` | `a6bbbf82b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f` | 41 | The existing `sha-full` value plus one hex char — same reasoning, and the identical value already appears as the 41-character row of `TestIsCommitSHAToken_LengthBand` (`internal/spec/syncsha_test.go:62`), so predicate-level and rule-level coverage name the same token. |

Open point, now ANSWERED by the plan audit: `sha-below6`'s value is 6 all-hex
characters, a *shape* rejection rather than an alphabet rejection. The audit's
answer is **keep it** — the card's axis is length, and a value differing from
`sha-min7` only in length is what isolates that axis; a non-hex character would
move two variables and weaken mutant 3 as evidence. The genuinely-uncovered
alphabet case belongs to a different axis and is declared as a residual rather
than fixed here (`spec.md` §C).

### M2 — Where the outside-band cases get exercised — DECIDED (operator ruling)

**This point is closed. It is not open for run-phase re-litigation.**

`TestSyncSHASlot_SilentOnSHA` asserts **0** findings, so `sha-below6` and
`sha-above41` cannot join its case list. The plan-audit found that the apparent
alternative — extending `TestSyncSHASlot_FlagsProse` — is unavailable: that
function's name is quoted verbatim in a live criterion of a `completed` SPEC
(`.moai/specs/SPEC-SYNC-SHA-SLOT-FORMAT-001/acceptance.md:56`), so overloading it
destroys AC-SSF-001's identity and renaming it falsifies AC-SSF-001's command
string.

**Decision**: add a fourth test function, `TestSyncSHASlot_FlagsOutOfBand`,
carrying `sha-below6` and `sha-above41`. `TestSyncSHASlot_FlagsProse` keeps only
the prose case, unchanged in name and behavior.

**Reason**: this option relaxes **this card's own** criterion (the pre-audit
AC-SBB-005 pinned the `func Test` count at 3), while the alternative would have
amended a completed SPEC's judgment criterion. Amending a draft is cheap;
amending a completed SPEC's acceptance criterion is not.

Consequences carried into the artifacts:

- `REQ-SBB-005` is rewritten to permit a separate function precisely where an
  existing function's name is load-bearing elsewhere, and to forbid duplicating
  an existing assertion body.
- The merged `AC-SBB-005` replaces the count constraint with the
  no-duplicated-assertion-body constraint. The `func Test` count becomes **4**
  and that is expected.
- The new function shares the helper `syncSHAFindings` (`lint_syncsha_test.go:50`)
  and asserts the same five properties `TestSyncSHASlot_FlagsProse` does (count,
  severity, sibling `progress.md`, `Advisory == false`, flagged-line content) —
  over its own case list, parameterized by expected line prefix rather than
  hardcoding `"sync_commit_sha: TBD"` as `lint_syncsha_test.go:109` does.

### M3 — Fixture files

Three new directories under `internal/spec/testdata/syncsha/`: `sha-min7`,
`sha-below6`, `sha-above41`. Each gets `spec.md` + `progress.md` copied from the
`sha-short` shape measured in this tree, changing only the title text and the
`sync_commit_sha` value. Keep `id: SPEC-SSFA-001` and `status: in-progress`.

### M4 — Test registration

Add `sha-min7` to `TestSyncSHASlot_SilentOnSHA`'s case list
(`lint_syncsha_test.go:126`, resolved at base `3f03d9c36`; this card's M4 commit
moved that line to `:138`). Add the new `TestSyncSHASlot_FlagsOutOfBand`
carrying `sha-below6` and `sha-above41` per the M2 ruling; leave
`TestSyncSHASlot_FlagsProse` untouched. Update the doc comments so each names the
mutation it catches, matching the per-fixture table of `spec.md` §B.1 — the file's
existing convention.

### M5 — Bidirectional mutation observation (the evidence that decides the card)

Four mutants, planted one at a time at `internal/spec/lint_syncsha.go:103`,
replacing the `isCommitSHAToken(token)` call with an inlined match:

| # | Mutant | Direction | Expected red |
|---|---|---|---|
| 1 | `^[0-9a-fA-F]{8,40}$` | floor narrowed | `sha-min7` |
| 2 | `^[0-9a-fA-F]{7,39}$` | ceiling narrowed | `sha-full` |
| 3 | `^[0-9a-fA-F]{6,40}$` | floor widened | `sha-below6` (finding count → 0) |
| 4 | `^[0-9a-fA-F]{7,41}$` | ceiling widened | `sha-above41` (finding count → 0) |

For each: plant → run `go test ./internal/spec/ -run TestSyncSHASlot -count=1` →
capture verbatim output **and the exit code as its own field** to
`.moai/reports/t380/mutant-0N-*.txt` → revert → confirm GREEN. A mutant that does
NOT turn red is a finding about the fixture set, not a reason to move on.

The plan audit already planted all four under its own hands and observed every one
red (`spec.md` §A.4). That is a prior observation, not a substitute: this milestone
re-measures each, because two of the audit's four entries (R-3, R-4 in
`acceptance.md` §B) carry no verbatim stdout and none carries an exit code.

### M6 — Close-out (mechanical)

`go test ./internal/spec/... -count=1` and `go vet ./internal/spec/...` on the
clean tree; record exit codes. Fill `progress.md` §E.2 with the per-criterion
four-element cells (command / verbatim stdout / exit code / tree SHA) completing
ledger entries R-1..R-4 — a summary matrix is not sufficient — and §E.3 with the
run-phase audit-ready signal.

---

## §F. Risks

- **A mutant fails to turn red.** Most likely cause: the fixture was demoted to
  advisory (era precondition) and the criterion is measuring nothing. Diagnose by
  asserting `Advisory == false` before chasing the regex.
- **The lint corpus prediction is wrong.** See §G.
- **Scope creep toward the two rejected designs.** Both rejections are recorded in
  `spec.md` §C with their grounds; re-opening either is an operator decision, not
  an implementer's.
- **Scope creep toward the alphabet residual.** The plan audit measured that an
  alphabet-narrowing mutant survives even with all four fixtures installed
  (`spec.md` §C, "Out of Scope — the alphabet clause"). It is DECLARED, not fixed.
  Adding an uppercase fixture would widen the card onto a second axis.
- **Re-litigating the M2 home.** Closed by operator ruling; run phase implements
  `TestSyncSHASlot_FlagsOutOfBand` as decided.

---

## §G. Prediction to be checked at run phase (not an observation)

`moai spec lint` is predicted NOT to move on this card: the new fixtures live under
`internal/spec/testdata/` and not under `.moai/specs/`, so they should be outside
the linted corpus. This was **not measured** at plan phase. Run phase records the
before/after finding counts and reports the prediction as confirmed or falsified.

---

## §H. Cross-references

- `SPEC-SYNC-SHA-SLOT-FORMAT-001` — the originating card (t299): the grammar
  (§D.1), the shared predicate, AC-SSF-007, and the §E.4 debt D1 this card closes.
- `internal/spec/syncsha.go` — the grammar and the shared predicate.
- `internal/spec/lint_syncsha.go` — the read-side gate; line 103 is the mutation
  site.
- `internal/spec/closer.go:421` — the write-side gate, untouched.
- `internal/spec/syncsha_test.go:52` — predicate-level band coverage that the
  mutant bypasses (`spec.md` §A.2).
- `.moai/reports/t380/probe-01..05` — the plan-phase probe, fully reverted.
