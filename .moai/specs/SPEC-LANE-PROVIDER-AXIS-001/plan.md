# Implementation Plan — SPEC-LANE-PROVIDER-AXIS-001

Tier M. Three milestones: M1 (lane pool axis), M2 (provider spread display), M3 (regression guard).

At Tier M the artifact set is three files plus `progress.md`, so the design rationale and the observed current-state evidence that would otherwise live in `design.md` and `research.md` are carried here, in §B and §C.

---

## §A Scope surfaces

[HARD] These six files are the **complete** set this SPEC may touch. The declared surface list — not the requirement ceiling — is the boundary: the preceding card's verdict recorded that a SPEC passes its numeric ceiling while silently crossing a scope declaration, because the crossing does not show up as a number. A file not on this list requires a SPEC amendment, not a judgment call at run-phase.

| # | File | Role in this SPEC | Milestone |
|---|---|---|---|
| S-1 | `internal/web/factory_lanes.go` | the pool aggregate, beside the existing `[]LaneVM` builder | M1, M2 |
| S-2 | `internal/web/factory_lanes_test.go` | aggregate tests + the M3 regression guard | M1, M2, M3 |
| S-3 | `internal/web/screens.templ` | the spread render inside the existing `Factory lanes` panel head | M2 |
| S-4 | `internal/web/screens_templ.go` | generated from S-3 by `templ generate`; never hand-edited (C-3) | M2 |
| S-5 | `internal/web/assets/i18n.js` | the spread's labels, in all four locale blocks | M2 |
| S-6 | `internal/web/factory_lane_section_test.go` | render assertion + locale-parity assertion, extending the tests already there | M2 |

**Explicitly off the list**, and named so a reader does not have to infer it: `internal/template/profile_matrix.go` (owned by `SPEC-MODEL-MATRIX-CORE-001`), `internal/web/widgets.templ` and `widgets_templ.go` (the shared `backendBadge`), `internal/kanban/record.go` (the vestigial constant), and every `internal/cli/` file. Full reasoning: `spec.md` §C.

Note that S-1 is the only non-generated production file in the list. The SPEC's weight sits in tests and in a template, which is why the LOC estimate lands below the Tier S band while the file count lands in the Tier M band — the file count is what selects the tier.

---

## §B Review first — the decisions most likely to change

### §B.1 The crossing partner (M1) — the decision the whole SPEC rests on

The lane axis was proposed as `lane × backend-multiplexing`. That partner is gone (`spec.md` §C), so a partner has to be chosen, and the choice is not obvious. The live candidates, measured:

| Candidate partner | Live on the lane row? | Owned elsewhere? | Verdict |
|---|---|---|---|
| provider (`LaneVM.Backend`) | yes — rendered at `screens.templ:243` | no | **chosen** |
| resolution state (`Unresolved` + 4 reasons) | yes | no | already rendered per row; an aggregate over it is a different SPEC's idea |
| stage (`Stage`, `StageEstimated`) | yes | no | an estimate, so an aggregate over it would aggregate estimates |
| profile (`max`/`medium`/`low`) | no | **yes** — `SPEC-MODEL-MATRIX-CORE-001` | excluded by REQ-LPA-002 |
| agent | no | **yes** — same | excluded by REQ-LPA-002 |

**Design preference: provider.** It needs no new data source, no new owner, and no new read: the value is already on the row and already rendered. The two excluded candidates are excluded by ownership, not by merit — that is worth stating plainly, because "profile is the natural partner" was the original reading and it is the one a reviewer is most likely to re-propose.

The stage candidate deserves one more line, because rejecting it on "it is an estimate" is the kind of reason that sounds like a technicality. It is not: `StageEstimated` marks a stage derived from a heartbeat rather than a recorded transition, so a pool-level stage histogram would present derived values with the same weight as recorded ones, and nothing in the aggregate would carry the estimate flag. That is a display defect, not a scheduling preference.

### §B.2 What the spread must and must not say (M2) — a two-value design

The axis has **two reachable values**, not four and not three (`spec.md` §A.2). Three design shapes, and the choice matters more at two values than it would at ten:

| Shape | At two values | At the vestigial third |
|---|---|---|
| a bar or ratio chart | reads as a near-meaningless two-segment bar | unaffected |
| per-provider counts + an unrecorded count | reads cleanly; `3 claude · 1 glm · 2 unrecorded` | shows nothing, correctly |
| a `switch` over `kanban.Backend*` | reads cleanly | **presents `gpt: 0` as a column** |

**Design preference: the middle shape** — a small count set, derived from the values actually present on the rows plus one named bucket for the unrecorded case. The third shape is the trap: it is the most natural Go code to write, it passes every test that checks the two live values, and its defect (a column no writer can fill) is invisible until someone asks what `gpt` means. REQ-LPA-005 and AC-LPA-005 exist for that one failure.

The unrecorded bucket is not a rounding detail. An unresolved lane carries an empty `Backend` by construction — `unresolvedLane` sets no provider precisely so that no plausible substitute is asserted — so any aggregate that does not name the empty case will silently fold those lanes into whichever provider its zero value implies. The lane section already refuses to do that per row; the aggregate must refuse it too (C-4, D-3).

### §B.3 How the guard is made able to fail (M3)

This is the milestone most likely to ship as something that cannot fail. The invariant already holds structurally, so a guard written carelessly asserts a property the compiler and the type system already enforce, passes forever, and is mistaken for coverage.

Two mutation directions, and they catch different things:

1. **Coherent widening** — `Backend string` becomes `Backend []string` **and** the assignment at `factory_lanes.go:150` becomes a slice construction, in the same edit. This is the mutation that matters. The *incoherent* half-edit (widen the field, leave the assignment) is a compile error, so it proves nothing about the guard: a build failure is not the guard firing.
2. **Second assignment site** — a second `row.Backend = …` appears in the file, for instance as a fallback on an unresolved row. Nothing about the type changes, so no compiler signal exists at all, and the semantic invariant (one lane, one provider, one source) is what breaks.

The guard therefore needs two independent assertions, because direction 1 is a shape claim (reflect on the field's kind) and direction 2 is a count claim (assignment sites in the file). One assertion covering both would have to be a source scan, and a source scan alone cannot distinguish a dangerous widening from a harmless rename.

`[NEEDS CLARIFICATION: guard mechanism for direction 2]` — whether the assignment-site count is measured by `go/ast` over `factory_lanes.go` or by a narrower reflect-plus-invariant test is left to run-phase. The acceptance criterion (AC-LPA-008) constrains the *outcome* — the guard fails when a second assignment site is added — not the mechanism. An `ast` walk is the honest reading of "assignment site"; a regexp over source is not, because it cannot tell a comment from code.

### §B.4 Where the spread renders (M2)

The panel head already carries `len(k.Lanes)` as a `panel__meta` span (`screens.templ:203`). The spread belongs beside it, in the head, rather than as a new section: it is a property of the pool the panel is already introducing, and a separate section would read as a second subject.

When `len(k.Lanes) == 0` the panel renders a `No factory lane is registered.` banner instead of rows. The spread must not render an empty bucket set in that branch — an empty pool has no spread, and printing `0 claude · 0 glm` says less than the banner already does.

---

## §C Observed current state

Measured in this tree at `159dd30df`, with a positive control beside each absence claim. Full measurement record: `.moai/reports/t1047/premeasure.md`.

| Observation | Evidence |
|---|---|
| No `SPEC-MODEL-MATRIX-*` successor owns these files | `factory_lanes` / `kanban` / `launcher` → 0 hits in all four successor directories; positive control `agentfm` → 21 / 10 / 5 / 18 |
| `LaneVM.Backend` appears on exactly two lines repo-wide | `internal/web/factory_lanes.go:43` (declaration), `:150` (sole assignment); positive control on sibling `CardID` fires twice |
| Providers reachable by a live writer: two | `"claude"` ×4, `"glm"` ×2 across the three candidate files; positive control (lowercase string literals, same three files) → 60 |
| A third provider constant exists and is dead | `BackendGPT = "gpt"` at `internal/kanban/record.go:24`; zero live Go references outside its declaration and two doc comments; positive controls `BackendGLM` / `BackendClaude` fire across many live files |
| The provider render surface is the web console only | `backendBadge` is called from `screens.templ` at three sites (role row, lane row, monitor row) and defined at `widgets.templ:69`; the CLI's lane code is parse/launch logic with no per-lane provider print, against a positive control showing the CLI prints elsewhere (15 print statements across the two files) |
| An i18n-parity test pattern already exists for lane keys | `TestSPECIntroducedKeysResolveInEveryLocale` in `internal/web/factory_lane_section_test.go:209` |
| The lane join's read-only property is already asserted | `TestFactoryLanesJoinWritesNothing` in `internal/web/factory_lanes_test.go:233` |

One Gap from the pre-authoring measurement is now closed: the render surface (the last row above) was unmeasured when this SPEC was commissioned, and D-5 rests on the measurement rather than on an assumption.

---

## §D Open clarifications

Both must be settled before Implementation Kickoff Approval.

1. `[NEEDS CLARIFICATION: guard mechanism for direction 2]` — see §B.3. The outcome is fixed by AC-LPA-008; the mechanism is not.
2. `[NEEDS CLARIFICATION: spread label wording]` — the four-locale label set for the spread, including what the unrecorded bucket is called. "unrecorded" is the term used throughout these documents because it describes the record, not the lane; "unknown" would describe the reader's state instead. The wording is copy and should be reviewed as copy, in all four locales, not inferred from the English.

---

## §E Self-verification

Run at the close of each milestone, scoped to the changed package (a full-suite run belongs to CI):

```
go test -count=1 ./internal/web/...
golangci-lint run ./internal/web/...
gofmt -l internal/web
go vet ./internal/web/...
```

Plus, once per milestone that touches S-3: `templ generate` followed by a clean `git status` on S-4, confirming the generated file matches its source rather than a hand edit (C-3).

The two M3 mutation runs (§B.3) are evidence for AC-LPA-007 and AC-LPA-008 and are reverted immediately after being observed. A mutation left in the tree is a defect, not a test.

---

## §F Milestones

Ordered by decision-reversibility: M1 fixes the data shape every later milestone reads, M2 is user-facing, M3 is mechanical.

### M1 — The lane pool axis

Priority: High.

Add the pool aggregate to `internal/web/factory_lanes.go` (S-1), beside `loadFactoryLanes` rather than inside it — the builder's read-only join is already covered by an existing test and should not gain a second responsibility. The aggregate takes the built `[]LaneVM` and returns per-provider counts plus the unrecorded count.

Satisfies REQ-LPA-001, REQ-LPA-002. Tests in S-2.

### M2 — The provider spread display

Priority: High. Depends on M1.

Render the aggregate in the `Factory lanes` panel head (S-3 → S-4 via `templ generate`), add its labels to all four locale blocks (S-5), and assert both the render and the locale parity in S-6, extending the tests already there.

Satisfies REQ-LPA-003 … REQ-LPA-006.

### M3 — The regression guard

Priority: Medium. Independent of M1 and M2 — it asserts a property of the field as it stands today and could land first. It is ordered last because it is the mechanical milestone, not because it is blocked.

Add the two-direction guard to S-2 per §B.3, then demonstrate both mutations fail it.

Satisfies REQ-LPA-007.

---

## §G Anti-patterns

- **Implementing the one-provider-per-lane invariant.** It already holds (`spec.md` §A.3). Writing enforcement code produces a second thing to keep true and a check that can drift from the type it duplicates. The deliverable is a guard.
- **Deriving the bucket set from `kanban.Backend*`.** It is the obvious Go idiom and it presents a column no writer can fill (§B.2).
- **Widening `LaneVM.Backend` "while we are here"** to prepare for a provider the withdrawn gateway would have added. The gateway is gone; the widening would break the invariant this SPEC guards, in the same change that adds the guard.
- **Hand-editing `screens_templ.go`.** It is generated (C-3). A hand edit survives until the next `templ generate` and then vanishes silently.
- **Folding unresolved lanes into a provider bucket** because the empty-string case is "not interesting". It is the case that makes the buckets stop partitioning the pool (REQ-LPA-004).
- **Reading `profile_matrix.go` to "check what the profile axis looks like".** A read is how the scope crossing starts; REQ-LPA-002 forbids the reference, and AC-LPA-002 checks for it.

---

## §H Cross-References

- `spec.md` §C — the full exclusion set, including the preset discriminant and the three dead t843 items.
- `.moai/reports/t1047/premeasure.md` — the measurement record §C summarizes.
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — the tier bands and the 16/16 Tier M ceilings.
- `internal/web/factory_lanes.go` header comment — the non-unique join and the "present with a marker rather than dropped" rule that C-4 and D-3 follow.
