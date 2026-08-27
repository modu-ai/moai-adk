---
id: SPEC-GUARD-STATE-MODEL-001
title: "Guard liveness state model: declare firing expectations, and decide every entry into exactly one classification (card t347)"
version: "0.3.0"
status: draft
created: 2026-08-28
updated: 2026-08-28
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: ".moai/guards, internal/cli, internal/guard"
lifecycle: spec-anchored
tags: "guard, liveness, state-model, classification, manifest, cadence, totality, t347"
tier: M
era: V3R6
related_specs: [SPEC-GUARD-LIVENESS-001]
---

# SPEC-GUARD-STATE-MODEL-001 — Guard Liveness State Model (card t347)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-08-28 | manager-spec | Initial plan-phase authoring as **card t347**, created by the **scope reduction of `SPEC-GUARD-LIVENESS-001` (card t333) after its iter-3 FAIL + STOP** (0.800 → 0.800 → 0.667). This SPEC receives the state model — the family in which every one of that SPEC's hardest defects landed (D2, D5, N2, T2, T4) — plus the three unresolved findings that travel with it as **starting material rather than history**: T2 (a failed forge query has no admissible classification), N2's unresolved half, and T4. The state model is authored **as a state table, not as prose**: prose is what let a five-value vocabulary cover at least six states without anyone noticing. Totality is demonstrated by construction in both directions (§C.2). |
| 0.2.0 | 2026-08-28 | manager-spec | **Surface fold added, and card t326's landed implementation read before fixing the table** (measured on `origin/develop` at `ec15ec2cd`, a tree diverged from this SPEC's baseline — diverged, `merge-base --is-ancestor` false — where reading the baseline for a t326 surface reports a landed feature as absent). The state table gains a `Surface fold` column and REQ-GSM-013 requires every classification to declare its fold into the three-value surface vocabulary; **no classification folds to `fail`**, reusing t326's recorded reason that a `fail` on a routine check promotes a healthy installation to a failing exit status. **Reuse verdict: the fold discipline and the leniency principle transfer and are adopted with citation; the value vocabulary does not fit and the measurement showing why is recorded** — three of this SPEC's six values have no t326 counterpart and one of t326's four has none here, because the two answer structurally different questions. The leniency principle is bounded at the line between *meaningless* and *incomplete*: `UNREADABLE` folds to `ok` as t326's `not-applicable` does, while `UNKNOWN` and `UNRESOLVED` fold to `warn`, because folding the incomplete cases to green would reproduce this card's own subject inside its solution. AC-GSM-016 adopts all three clauses. Counts 12→13 REQ, 15→16 AC. |
| 0.3.0 | 2026-08-28 | manager-spec | **The producing half of the sibling SPEC's D1 repair, landed in the same commit as its consuming half.** The published contract named `OK` as the clean value in prose and emitted nothing, so a consumer could identify it only by hardcoding the literal — which the sibling's own criteria forbid — or by reading the surface fold, which **under-reports**, because `UNREADABLE` also folds to `ok` (REQ-GSM-013) while `OK` alone is the clean classification. REQ-GSM-012 now requires the result to **carry a machine-readable designation** of the clean value, and AC-GSM-015 gains clause (c) testing it the way a consumer would: given only the result and no value names, the entries must partition. The designator is not a re-enumeration — it carries *which*, not *what the set is* — so the seam stays closed under a vocabulary change. Counts unchanged at 13 REQ / 16 AC (the AC budget is at its ceiling, so the clause extends AC-GSM-015 rather than adding a criterion). |

## §A Context and Problem

### A.1 Where this SPEC came from, and why the split ran along this seam

`SPEC-GUARD-LIVENESS-001` asks *when did this guard last fire, and was it supposed to have fired by now?* Three audit iterations converged on the surfacing half of that question and never converged on this half. The auditor's own accounting:

- The **surfacing model** converged — both prior mutants were re-run at iter-3 and neither could be revived.
- The **state model** did not. D2, D5, N2, T2 and T4 are one family, and each repair relocated the defect rather than closing it: narrowing `UNKNOWN`'s meaning left the excluded state with nowhere to go; adding a fifth value to receive it left a sixth state uncovered; declaring the set closed in one requirement while another routed entries into it produced a contradiction between two requirements over the very classification the trigger is defined on.

The lesson the auditor drew, and this SPEC adopts as its authoring method: **prose is what let a five-value vocabulary cover at least six states without anyone noticing.** A sentence defining a classification reads complete in isolation. Only a table, with one row per condition the mechanism can encounter, makes an uncovered condition visible as an empty cell.

### A.2 The defect this whole family exists to catch

*Absent execution is not suppressed failure.* Nothing runs, so nothing can go red, and a mechanism reports success accurately about a set that had silently become the wrong set. An exit code of zero means *everything selected passed*; it does not mean *everything that should have passed, passed*. **Green is accurate and uninformative at the same time.**

The empirical instances grounding this — a deployment lag that made a landed fix never run, two guards whose `pull_request` trigger stopped matching after the git-flow transition, a default run listing that hides low-frequency guards, a guard that fires and does nothing, and a `-run` selector naming two tests that did not exist — are recorded in `SPEC-GUARD-LIVENESS-001` §A.2 through §A.7 and in `.moai/reports/t333/trigger-axis-observation.md`. They are not restated here; this SPEC's job is the mechanism, and the mechanism's correctness condition is that **it must not reproduce the defect at its own layer**.

### A.3 The three findings inherited as starting material

These arrive unresolved from the parent SPEC's iter-3 audit. They are the work, not the background.

- **T2 — a failed forge query has no admissible classification.** Two requirements disagreed about the extent of the "could not interpret" value: one routed *every* non-retention failure into it, the other defined it as *kind has no reader* only. A workflow entry whose query **errors** — the single most likely degraded run in production — fell in the gap between those two sentences, and the acceptance criteria assumed a classification the requirements forbade. Resolved here by the state table: the error case is its own row with its own value (§C.2, row 2).
- **N2's unresolved half.** The fifth value repaired the no-reader hole and left the more common one open. The state table is the structural fix — a hole is now an empty cell rather than a sentence nobody wrote.
- **T4 — the plan's own vocabulary went stale under a repair.** The five-value closed set landed in the requirement layer and never reached the plan's milestone body, and a criterion was scheduled at a milestone that could not deliver it. Addressed here by making the table the single artifact both layers cite (§C.2) and by clause-splitting criteria across milestones where they genuinely split (`plan.md` §F).

### A.4 The correctness condition this SPEC must meet at its own layer

An evaluator that reads outcomes only for entries its manifest already lists **is** the defect: accurate about what it examined, silent about what it never examined. Two requirements exist for that reason and are load-bearing rather than housekeeping — REQ-GSM-008 compares the manifest against the disk enumeration, and REQ-GSM-010 refuses an all-clear when either input to that comparison came back empty. The audit found the query side guarded and the disk side not (T1); both are guarded here.

## §B The published contract

`SPEC-GUARD-LIVENESS-001` consumes exactly three clauses from this SPEC, and nothing else:

> every entry in an evaluation result carries exactly one classification; exactly one value of that vocabulary means *nothing to report* (the **clean** value); the advisory fires on any other.

REQ-GSM-012 is this SPEC's obligation to hold up that contract. The consuming SPEC does not name the vocabulary, does not know how many values it has, and does not decide any entry — so **this SPEC may add, rename, or re-decide classifications freely, provided exactly one clean value continues to exist.** That is the whole coupling, and it is why the two SPECs are independently implementable rather than one blocking the other.

The asymmetry is worth stating plainly: the consuming SPEC can verify that its own trigger uses the clean/non-clean partition, but it cannot verify that exactly one clean value exists. That half of the contract is verified here, by AC-GSM-015.

## §C Requirements (GEARS)

Budget: Tier M ≤ 16 requirements. **Count: 13.**

### C.1 The manifest — where firing expectations are written

- **REQ-GSM-001** — The system shall carry a guard-liveness manifest declaring one expectation entry per workflow file under `.github/workflows/`. The manifest shall live outside `.moai/config/`, because `moai update` deletes that root wholesale and a manifest lost on update is a liveness record that itself silently stops. The manifest shall hold its watched set as **data** — each entry carrying its kind, its locator, and its expected cadence — and shall not be shaped so that only a forge workflow can occupy an entry.
- **REQ-GSM-002** — Each manifest entry shall carry the subject's locator, the triggering event or events under which firing is expected, an expectation window, and exactly one measured quantity.
- **REQ-GSM-003** — The measured-quantity vocabulary shall be exactly `fired-at-all`, `fired-with-effect`, and `verdict-rendered`, where `fired-with-effect` excludes runs whose conclusion is `skipped` or `cancelled`, and `verdict-rendered` additionally requires a terminal `success` or `failure`. Each entry shall name exactly one, and **no entry shall be written whose single number is asked to measure both whether a guard fired and whether a firing caught anything** — these are different axes, and a guard that runs faithfully and catches nothing scores full marks on the first.
- **REQ-GSM-004** — Where a subject is legitimately expected to be quiet outside a release cycle, its entry shall declare that condition explicitly rather than being omitted, so "correctly quiet" is a recorded expectation rather than an absence a reader must infer.

### C.2 The state table — the normative classification artifact

- **REQ-GSM-005** — When the evaluator runs, it shall query run history **per subject**, and shall not derive any subject's last-fired time from a repository-global run listing. The global listing is measurably incapable of answering the question for a low-frequency subject, and its incapacity is invisible from inside it.

- **REQ-GSM-006** — The evaluator shall decide every entry it encounters by the **state table below, which is normative**. The table is the artifact; the prose around it describes it and does not extend it.

  | # | Entry condition | Classification | Surface fold | Implied action | Flipped by (and what it needs first) |
  |---|---|---|---|---|---|
  | 1 | Entry's declared kind has no reader in this deliverable | `UNREADABLE` | `ok` | None until a reader for that kind exists | M2 — **needs M1's `kind` field**; without it this row is undeliverable regardless of M2 |
  | 2 | Kind has a reader; the query did not return a result (error, auth failure, rate limit, timeout) | `UNRESOLVED` | `warn` | Retry; check credentials and forge availability | M2 — needs the query path's error channel to be distinguishable from an empty result |
  | 3 | Query returned; a qualifying run exists inside the expectation window | `OK` | `ok` | None | M2 — **needs M1's window field** and M1's measured-quantity value (which runs qualify) |
  | 4 | Query returned; a qualifying run exists, most recent one older than the window | `STALE` | `warn` | Investigate why the subject stopped firing | M2 — same M1 dependencies as row 3 |
  | 5 | Query returned zero qualifying runs; the entry's declared condition says firing is not currently expected | `OK` | `ok` | None — this is the observation the entry predicted | M2 — **needs M1's release-cycle-conditional field** (REQ-GSM-004); absent it, this row collapses into row 6 and the value becomes wrong rather than missing |
  | 6 | Query returned zero qualifying runs; the entry's declared condition does not account for the absence | `UNKNOWN` | `warn` | Look again with a longer window; retention may have consumed the history | M2 |
  | 7 | A workflow file exists on disk with no manifest entry | `UNDECLARED` | `warn` | Add an entry, or record why the subject is unwatched | M2 — needs the disk enumeration, whose degradation REQ-GSM-010 guards |

  **The last column is not decoration, and it is the reason this table has five columns rather than four.** The predecessor SPEC's milestone map passed a union count — every criterion appeared in some milestone's flip list — while one criterion was assigned to a milestone that could not deliver it: it required a classification the milestone's own vocabulary bullet did not yet contain. A union count answers *"is every criterion listed?"* and is **structurally incapable** of answering *"can the listed milestone deliver it?"*. Recording the delivering milestone per cell, together with the M1 field each cell depends on, is what makes the second question checkable — and without it this table would prove totality of *classification* while saying nothing about totality of *delivery*, reproducing the defect inside the artifact built to prevent it.

  Three rows (1, 3/4, 5) depend on an M1 schema field. Rows 3 and 5 are where a missing dependency is most dangerous: it does not leave the cell empty, it makes the cell produce the **wrong value** — row 5 without its conditional field silently becomes row 6, turning a correctly-quiet release-only subject into a reported anomaly on every sweep.

- **REQ-GSM-013** — Every classification shall declare its **surface fold** — the value it becomes in the three-value surface vocabulary `ok` / `warn` / `fail` — and the fold shall be the `Surface fold` column of the table above. No classification shall fold to `fail`: this SPEC's consumer renders a **non-blocking** advisory, and a `fail` would promote a routine sweep to a failing exit status. An internal vocabulary that never states its fold is describing values the surface cannot express.

#### C.2.1 The fold, and what was reused from card t326 rather than re-derived

Three SPECs now answer one question — *what do you call an entry you cannot judge?* — and three independent answers would make three vocabularies. Card t326 has a landed implementation, so it was read before this table was fixed. **Measured on `origin/develop` at `ec15ec2cd`**, a tree diverged from this SPEC's baseline (diverged, `merge-base --is-ancestor` false; the surfaces are absent from the baseline tree and reading it reports a landed feature as missing).

The surface vocabulary is three values and has no skipped state:

```
$ git show origin/develop:internal/cli/uikit/types.go
CheckOK   CheckStatus = "ok"
CheckWarn CheckStatus = "warn"
CheckFail CheckStatus = "fail"
```

And t326's own fold, read from `internal/cli/doctor.go` `checkBinaryFreshness`, is **four internal values onto two surface values** — `fail` is never emitted, deliberately:

| t326 internal | folds to |
|---|---|
| `StatusBehind` | `CheckWarn` |
| `StatusFresh` | `CheckOK` |
| `StatusDivergent` | `CheckOK` |
| `StatusNotApplicable` (default arm) | `CheckOK` |

> Reported OK (never Fail — a Fail here would promote every downstream `moai doctor` run to exit 1)

**What was reused.** Two things, both adopted with citation rather than re-derived:

- **The never-emit-`fail` discipline** (REQ-GSM-013). t326's reason transfers exactly: a routine check that can fail the process turns a healthy installation into a reported fault. This SPEC's consumer is explicitly non-blocking, so the same conclusion holds for the same reason.
- **The leniency principle.** t326's `gitCompare` comment records that *five* paths report not-applicable or otherwise-fine rather than a problem, and that "that leniency is deliberate and must not be narrowed: each narrowing turns a perfectly healthy downstream installation into a reported fault." Row 5 of this table is that principle applied here — a declared-quiet subject between releases is *fine*, not unknown — and row 1 folds to `ok` on the same reasoning.

**What did not fit, and this is the more useful finding.** The *vocabulary* does not transfer. Measured against t326's four values, three of this SPEC's six have no counterpart (`UNKNOWN`, `UNRESOLVED`, `UNDECLARED`) and one of t326's four has none here (`Divergent`, an ancestry relation with no meaning for firing history). The two SPECs answer structurally different questions: t326 decides an **ancestry relation on one binary**, this SPEC decides **set membership and history across N subjects**. Adopting t326's four values would have forced the three set-and-history states into `not-applicable`, which is the collapse that failed the predecessor three times.

**And the leniency principle has a boundary this SPEC must draw, because the two subjects diverge exactly here.** t326 folds uncertainty toward `ok`; this SPEC exists because *uncertainty folded to green is the defect* (§A.2). The boundary is the difference between **meaningless** and **incomplete**:

- `UNREADABLE` (row 1) is **meaningless** — no reader exists, so the comparison was never applicable. It folds to `ok`, exactly as t326's `not-applicable` does.
- `UNKNOWN` and `UNRESOLVED` (rows 2 and 6) are **incomplete** — the comparison was applicable and could not be completed. They fold to `warn`, diverging from t326.

Folding the incomplete cases to `ok` would reproduce this card's own subject inside its solution: a mechanism reporting green about a set it never learned. The leniency principle is adopted for the meaningless case and explicitly not extended to the incomplete one.

- **REQ-GSM-007** — The classification vocabulary shall be exactly the closed set `OK`, `STALE`, `UNKNOWN`, `UNDECLARED`, `UNREADABLE`, `UNRESOLVED`, and each value shall carry a distinct implied action; no two values shall share one. **Totality binds in both directions**: every row of the table shall map to exactly one value (no entry condition is unclassifiable), and every value shall be reachable from at least one row (no value is an unused option). A value no entry can receive is a defect in the evaluator; **an entry no value can receive is a hole in the state space**, and the second direction is the one that has actually failed.

- **REQ-GSM-008** — Where a workflow file exists on disk with no manifest entry, the evaluator shall classify it `UNDECLARED` and report it. This is the **set comparison**: it asks whether a subject *was in the set that was examined, and should have been*, rather than reading outcomes only for subjects the manifest already knew about. An evaluator without it reproduces the defect at its own layer — accurate about what it looked at, silent about what it never looked at.

### C.3 Result integrity

- **REQ-GSM-009** — When the evaluator emits a result, it shall carry its own measurement timestamp, the count of workflow files its disk enumeration returned, the count of entries declared, the count successfully queried, and a per-value count for each classification. **Every carried count shall have a consumer** — a count that is rendered and consumed by nothing is field-level inertness, and the audit found two such counts in the predecessor.
- **REQ-GSM-010** — The evaluator shall not report an all-clear while its successfully-queried count is zero **or its workflow-file enumeration returned zero files**. Both inputs to the set comparison can degrade, and an enumeration that silently returns empty — wrong working directory, a path constant, a permissions failure — makes `UNDECLARED: 0` indistinguishable between "census complete" and "the evaluator never looked".
- **REQ-GSM-011** — The evaluator shall not write to the repository working tree, commit, push, open an issue, or mutate any forge state. It reads and reports.
- **REQ-GSM-012** — The result shall satisfy the published contract (§B): every entry carries exactly one classification; **exactly one value — `OK` — denotes *nothing to report***; and **the result shall carry a machine-readable designation of which value that is** — a result-level clean-value designator, or an equivalent per-entry cleanliness flag — so a consumer can identify clean entries without knowing any value's name. This SPEC may add, rename, or re-decide classifications provided those clauses continue to hold, and the consuming SPEC cannot verify them, so they are verified here.

  **The designator is what the seam actually rests on, and it is not redundant with the surface fold.** More than one classification folds to the clean *surface* value — `UNREADABLE` folds to `ok` (REQ-GSM-013) while `OK` alone is the clean *classification* — so a consumer reading the fold would under-report. Emitting the designation is therefore the producing half of a symmetric obligation whose consuming half is `SPEC-GUARD-LIVENESS-001` REQ-GDL-001 clause (iii)/(iv); neither half is sufficient alone, and a change to one without the other is exactly the cross-seam drift the split exists to make visible.

## §D Out of Scope

### Out of Scope — the advisory and its surfacing
- When the advisory renders, how it reaches the operator, whether it arrives unprompted, how it states its own measurement age, and how it leads with changes. All of it is `SPEC-GUARD-LIVENESS-001` (REQ-GDL-002..008).
- This SPEC publishes the contract in §B and verifies its own half (REQ-GSM-012); it does not consume it.

### Out of Scope — the binary-lag state comparison
- Whether the installed binary's build commit is a strict ancestor of the tree HEAD. Owned by card t326 (`SPEC-BINARY-LAG-VISIBILITY-001`, in flight; absent from this tree, so its cited requirement range is not verifiable from here).

### Out of Scope — repairing individual guards
- Rewiring any workflow the evaluator classifies `STALE` or reports `UNDECLARED`. The evaluator's job is to make the classification visible; acting on a particular one is a separate card.

### Out of Scope — policy-rule firing
- Extending the watched set to policy-layer rules, which have no run records. REQ-GSM-001 requires the manifest to hold its watched set as data so a second kind could be added later without a schema rewrite, but no second kind ships here.

### Out of Scope — guard correctness
- Whether a subject that fired would have caught a real defect. This SPEC measures firing, and REQ-GSM-003 exists to stop a single number from being read as though it measured both.
