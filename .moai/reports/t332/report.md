# Backlog hygiene sweep — consolidated reading (card t332, SPEC-BACKLOG-HYGIENE-001)

**No card was mutated by this sweep, and the disposition list below awaits the operator.**
It is a reading, not a decision. Nothing here drops, edits, closes, reorders, unpicks or picks a
card, and no fold is performed.

| | |
|---|---|
| Worktree / HEAD | `.claude/worktrees/t332`, `WT-backlog-hygiene`, `6165f9f5e` (after absorbing `origin/develop` `ee50984ab`) |
| Snapshot | `queue-snapshot-run.tsv`, 101 rows, captured 2026-08-30T11:16:45Z |
| Pinned `origin/develop` | `ee50984abe4f11ac337382b48a26328f091e200a` (fetched once, 11:16:22Z) |
| Pinned `origin/main` | `48239c7dc7428c8751a04f6321887c2d36123884` |
| Landing method | path A — direct git. Governing `strings ~/go/bin/moai \| grep -c worktree_base_branch` = **0** |
| In scope | **62 cards** (62 `queued`, minus `t332` which is itself `picked`) |
| Out of scope | 17 `picked` (live lanes), 18 `dropped` (already decided) |
| Relations recorded | **0** (see `02-overlaps.md`) |

## §1 What the sweep found, in one table

| Axis | Result |
|---|---|
| premise `holds` | **42** |
| premise `falsified` | **11** |
| premise `unverified` | **9** |
| landing `not-landed` | **55** |
| landing `landed` | **3** — t201, t313, t347 |
| landing `in-flight-unlanded` | **3** — t154, t216, t337 |
| landing `unknown` | **1** — t224 |
| disposition `keep` | **34** |
| disposition `needs-operator-decision` | **21** |
| disposition `already-landed` | **5** |
| disposition `drop` | **2** — t254, t281 |

**Eleven of 62 cards — nearly one in five — carry a premise that measurement contradicts.** That is
the number the card was written to produce, and it is the argument for the sweep existing: each of
those eleven reads as actionable in the queue and is not.

## §2 The three failure shapes of §A, measured

**The premise rots.** 11 falsified. The sharpest three:

| Card | What the card says | What measurement shows |
|---|---|---|
| t243 | a hook was deleted with its siblings, the siblings were restored two commits later and this one was left orphaned | the file has **exactly one commit ever — its creation.** There is no deletion to be orphaned by |
| t247 | PR #1600 is 497 files and CONFLICTING, split it | the PR is **MERGED at 10 files**, and its merge commit is an ancestor of pinned `main` |
| t281 | an operator decision of 2026-08-26 keeps `develop` local-only and disposable | commit `11216d13f` on pinned develop states in its own message that it is the *"Operator-directed reversal of the 2026-08-26 t281 decision"* — git-flow was adopted the next day |

**The work already landed.** 3 cards are `landed` and still queued; **t313** is the clean instance —
merged as `62ff3c2e6` (`SPEC-WORKTREE-BASEREF-001`), `--is-ancestor` exit 0 against pinned develop,
card still sitting in the queue.

**Two cards describe one defect.** **Measured absent.** All 62 were cross-compared and no pair
qualifies for any of the four relation tokens (`contains` / `absorbs` / `replaces` / `conflicts`).
Every candidate turned out to be a shared *origin*, a shared *file*, or a shared *precondition* —
adjacency, not duplication. Recording those as relations would have manufactured findings the
operator would then act on. Detail and the full candidate table: `02-overlaps.md`.

## §3 Findings the sweep produced about its own instruments

Four, and all four are the shape this card exists to catch, reproducing inside the card:

1. **AC-BH-006 was measuring a dead file.** The queue migrated to SQLite at `3cb258d62`
   (`SPEC-TODO-SQLITE-001`, card t306); `backlog.json` has not been written since. The criterion's
   digest returned its own plan-phase figure unchanged *after* the queue had moved seven cards and
   gained one — **a criterion that could not red.** Repaired to the live store with four controls
   re-measured there. (`00-tooling-baseline.md` § Finding B1'.)
2. **AC-BH-004 could not green.** Its non-emptiness conjunct demanded `wc -l` ≥ the number of cards
   read — which is precisely the per-card live invocation REQ-BH-001 **prohibits**. The mirror image
   of (1): one criterion could never fail, the other could never pass, and both were written in the
   same plan phase that was watching for exactly this.
3. **A recorded relation cannot be annotated with later evidence.** `t313 contains t295` was
   recorded 2026-08-27. t313 has since landed; measurement shows it did **not** resolve t295 —
   `gitWorktreeAddArgs` still hardcodes `-b <branch>`, so no existing-branch checkout path exists.
   The attempt to append that measurement was refused: *"t313 contains t295 is already recorded."*
   The relation stays frozen at its original claim. (`02-overlaps.md` §3.)
4. **plan.md's own gitignore premise is false.** plan.md B5 and spec.md §E both state that evidence
   written under `.moai/reports/t332/` inside this worktree "is gitignored and is lost on
   disposal", making an out-of-band recovery copy to the primary checkout an explicit M5 step.
   Measured at M5, and re-measured in this sync phase:

   ```
   $ git check-ignore -v .moai/reports/t332/report.md
   (no output)
   $ echo $?
   1
   ```

   **Not ignored.** `git status --short` lists the tree as untracked-but-trackable, so the evidence
   is committed on the card branch and travels with it. No `cp` + `cmp` recovery was performed and
   none was needed. This is the **fourth** instance of the same shape in this card's own lifecycle.
   The lead flagged the second: the plan-audit iter-1/iter-2 reports asserted their own output was
   gitignored, which the same one-line command refutes. That correction closed the instance in
   front of it and left the identical sentence standing one file over, in plan.md B5 and spec.md
   §E, where nothing re-ran the deciding command. The false sentences in plan.md B5 and spec.md §E
   were **not** repaired by this sweep — repairing SPEC body content is out of this card's scope
   (spec.md §C "Out of Scope — repairing what the sweep finds"), so the correction is a disposition
   item for the operator, not a performed act.

Two smaller observations, reported and not acted on: the `relate` error message names `backlog.json`
— the dead file — rather than the live `backlog.db`; and three fan-out workers wrote compound
premise verdicts (`holds` for one half, `unverified` for the other), which AC-BH-011 does not admit
and which the orchestrator normalized to `unverified` in place, with the normalization stated in
each entry.

## §4 The three no-mutation observables

| Observable | Result | Where |
|---|---|---|
| **Invocation log** (AC-BH-004) — every `moai todo` the sweep issued | 2 invocations: one snapshot read, one `relate` that was **refused** and wrote nothing. Zero card-mutating verbs. | `invocations.log` |
| **Count comparison** (AC-BH-005) | `queued` 62 → 62, `dropped` 18 → 18 unchanged; `picked` 17 → 16 | `01-scope.md` §8.1 |
| **Card-row digest** (AC-BH-006) | opening `86a3fb05…dc20` → closing `053e9991…7b33` — **differs**, and every differing card is attributed | `01-scope.md` §8 |

The digest moved because the queue is shared and live. Exactly one card differs — **t346**, closed
by the lead, its sync-close commit `282daef19` already on pinned develop — and it was on this
sweep's picked exclusion list from before the first card was read. Projecting both captures to the
62 in-scope rows and diffing gives **no output**: every in-scope card is byte-identical in id, state
and text. The three observables fail independently and none of them fired.

## §5 Disposition proposal — 62 rows, awaiting the operator

Each row names the card, its premise verdict, its landing verdict, the proposed disposition, and the
single piece of evidence the proposal rests on. **These are proposals.** Nothing was performed.
No row proposes an un-drop, and no row proposes a fold whose counterpart is a `dropped` card.

| Card | Premise | Landing | Proposed disposition | Evidence it rests on |
|---|---|---|---|---|
| t90 | `holds` | `not-landed` | **`needs-operator-decision`** | `needs-operator-decision` — the coded precondition (M1-M5) looks satisfied, so the remaining gate is purely the operator's demand judgment; only the operator can say whether that's been cleared. |
| t125 | `unverified` | `not-landed` | **`needs-operator-decision`** | `needs-operator-decision` — verifying the external LICENSE state requires a live check the operator (or a WebFetch-capable follow-up) should perform; this sweep cannot settle it. |
| t154 | `holds` | `in-flight-unlanded` | **`needs-operator-decision`** | `needs-operator-decision` — the operator declined this change on 2026-08-20 (recorded in the card's own text), so no implementation is pending; but the card is not `already-landed` — `git merge-base --is-ancestor dbb87f14f <pinned |
| t191 | `unverified` | `not-landed` | **`keep`** | `keep` — precondition #1601 confirmed merged; card is legitimately actionable next, pending the t170① confirmation the card itself flags. |
| t196 | `falsified` | `not-landed` | **`needs-operator-decision`** | `needs-operator-decision` — the card's precondition (t88/M4) is now satisfied and the agent-TOML problem is confirmed live, but the skills-mirror figures need re-measurement before the card's scope can be trusted as stated; the op |
| t201 | `falsified` | `landed` | **`already-landed`** | `already-landed` — drop from the backlog. |
| t204 | `holds` | `not-landed` | **`needs-operator-decision`** | `needs-operator-decision` — only the operator can confirm whether their own testing-completion declaration has occurred; this sweep can only confirm the tag hasn't been cut. |
| t216 | `holds` | `in-flight-unlanded` | **`needs-operator-decision`** | `needs-operator-decision` — this card explicitly frames its central open question as a design decision ("moai update가 추가된 훅 엔트리를 기존 프로젝트에 반영해야 하는가?"), which only the operator can settle; the mechanical D-1 finding itself is re-con |
| t223 | `unverified` | `not-landed` | **`needs-operator-decision`** | `needs-operator-decision` — the precondition-check requires reading a specific SPEC's milestone status, which is outside this card's bounded investigation depth; flagging for a follow-up check rather than guessing. |
| t224 | `falsified` | `unknown` | **`needs-operator-decision`** | `needs-operator-decision` — the doctrine-text half looks satisfied but cannot be confirmed as a direct response to this card, and the card's core runtime-symptom complaint is unverifiable from static file inspection alone; the ope |
| t231 | `holds` | `not-landed` | **`keep`** | `keep` — the premise is current and the requested exit-2 discrimination + REQ-WR-016 amendment genuinely has not happened; rests on the `clean_lock_unreadable_test.go` non-blocking assertion above. |
| t233 | `falsified` | `not-landed` | **`needs-operator-decision`** | `needs-operator-decision` — the card as currently worded (centered on silent/무통지 pass) is falsified, but the narrower remaining gap (no biome/oxlint lint step) may still warrant a rewritten, narrower card. Rests on the `markSkippe |
| t236 | `unverified` | `not-landed` | **`needs-operator-decision`** | `needs-operator-decision` — the checkable half holds; the runtime half cannot be settled from this tree and would need a live reproduction (enter/exit a worktree, inspect whether `CwdChanged` fired) rather than static reading. |
| t237 | `holds` | `not-landed` | **`keep`** | `keep` — premise holds, not landed, patch reportedly exists but unmerged. |
| t239 | `unverified` | `not-landed` | **`keep`** | `keep` — structural mechanism confirmed; recommend the operator re-verify which file currently carries the audit-pin values before scoping a fix. |
| t240 | `falsified` | `not-landed` | **`needs-operator-decision`** | `needs-operator-decision` — the measurement/marker half is resolved; whether a distinct doctrine file still needs a matching edit needs the operator (or a follow-up read of `.moai/reports/t225/sync-audit-review-2.md`) to confirm. |
| t242 | `holds` | `not-landed` | **`keep`** | `keep` — premise holds strongly; this is a `needs-operator-decision`-shaped card by its own design (asks for a judgment, not a fix), so `keep` as-is is the right disposition for the sweep to propose, letting the operator make the  |
| t243 | `falsified` | `not-landed` | **`already-landed`** | `already-landed` — the specific claimed defect (file missing, sibling hooks restored, this one orphaned) does not describe the current tree; the file is present. Rests on the single-commit `git log` result above. The still-open "r |
| t244 | `holds` | `not-landed` | **`keep`** | `keep` — premise holds; this is a `needs-operator-decision`-shaped card, same pattern as t242/t243. |
| t247 | `falsified` | `not-landed` | **`already-landed`** | `already-landed` — rests on the `gh pr view` state above (MERGED, 10 files, ancestor of pinned main). |
| t248 | `holds` | `not-landed` | **`needs-operator-decision`** | `needs-operator-decision` — narrow single-file check supports `keep`, but the gaps above (2 of 3 named tools + the persistence layer unchecked) mean a fuller read is needed before committing to scope. |
| t252 | `holds` | `not-landed` | **`keep`** | `keep` — premise holds, card is well-formed and still actionable; rests on the spec.md/progress.md read above. |
| t253 | `holds` | `not-landed` | **`keep`** | `keep` — premise holds on direct code read; the card correctly identifies an unbounded-write path. |
| t254 | `falsified` | `not-landed` | **`drop`** | `drop` — the specific defect claimed does not exist at either cited location; rests on the line-55/line-32 reads above plus the project's own prior memory finding on code-span vs table-cell escape behavior. |
| t255 | `holds` | `not-landed` | **`keep`** | `keep` — premise holds; well-formed, correctly gated follow-up. |
| t260 | `unverified` | `not-landed` | **`needs-operator-decision`** | `needs-operator-decision` — the card explicitly asks for a decision on channel design/scope, not a mechanical fix; the structural premise is plausible but the specific numbers are unverified here. |
| t262 | `holds` | `not-landed` | **`keep`** | `keep` — premise holds on direct source inspection; genuine gap. |
| t263 | `falsified` | `not-landed` | **`needs-operator-decision`** | `needs-operator-decision` — the premise as stated is false today (t216 hasn't landed), but the underlying direction may still be correct once t216 does land; the operator should decide whether to defer this card behind t216 or dro |
| t264 | `holds` | `not-landed` | **`keep`** | `keep` — premise holds and has strengthened; rests on the branch-count and worktree-occupancy comparison above. |
| t280 | `holds` | `not-landed` | **`keep`** | `keep` — premise holds; well-evidenced structural gap between what ships and what doesn't. |
| t281 | `falsified` | `not-landed` | **`drop`** | `drop` — commit `11216d13f` on pinned `origin/develop` states in its own message it is the "Operator-directed reversal of the 2026-08-26 t281 decision", recorded in the same file the card would have edited. |
| t284 | `holds` | `not-landed` | **`keep`** | `keep` — premise holds on direct code read. |
| t286 | `unverified` | `not-landed` | **`needs-operator-decision`** | `needs-operator-decision` — rests on: the found candidate guard already documents the false-positive fix and explicitly accepts the flag-order gap as correct-by-design; the operator should confirm whether issue #1658 targets this  |
| t287 | `unverified` | `not-landed` | **`needs-operator-decision`** | `needs-operator-decision` — rests on: this session's own two directly-observed refusals of non-git compound commands, matching the "Claude Code worktree isolation guard" that `branch_guard.go` itself says is external to this repo' |
| t288 | `holds` | `not-landed` | **`keep`** | `keep` — rests on: `parseCondition`'s single-keyword classifier is unchanged and is shared verbatim between the CLI and the MCP wrapper, so the misclassification path the card describes is real and reachable via the MCP entry poin |
| t295 | `holds` | `not-landed` | **`keep`** | `keep` — rests on: `session_worktree.go`'s own doc comment hardcodes `-b <branch>` (always new branch) and no call site was found passing an existing branch, while the lower-level git manager already supports it — a real gap betwe |
| t296 | `holds` | `not-landed` | **`keep`** | `keep` — rests on: direct read of the cited section (no "16" or language list) plus a direct grep confirming multiple dangling citations still pointing at it. |
| t297 | `holds` | `not-landed` | **`keep`** | `keep` — rests on: no prune/reap function found anywhere in the two most relevant packages, confirming scope item (2) — the reaping half of the card — is genuinely unaddressed, even though the dedup half (scope item (1)) is partia |
| t300 | `holds` | `not-landed` | **`keep`** | `keep` — rests on: the originating commit itself both confirms the defect and explicitly names t300 as the not-yet-done recurrence-prevention follow-up. |
| t302 | `holds` | `not-landed` | **`keep`** | `keep` — rests on: direct verbatim read of both contradicting passages in the same live file, plus independent corroboration that the "NOT binding" phrasing is externally visible (this session's own skill listing). |
| t304 | `holds` | `not-landed` | **`keep`** | `keep` — rests on: direct confirmation that all six named packages are absent from the tree and at least one codemaps file still cites them. |
| t305 | `holds` | `not-landed` | **`keep`** | `keep` — rests on: the profiling report's own numbers substantiate every figure the card cites, and no commit matching this card exists on either pinned ref. |
| t313 | `holds` | `landed` | **`already-landed`** | `already-landed` — rests on the `--is-ancestor` exit 0 evidence above. |
| t315 | `holds` | `not-landed` | **`keep`** | `keep` — two concrete carry-forward obligations, each with a clear originating SPEC and defect id, not yet actioned. |
| t319 | `holds` | `not-landed` | **`needs-operator-decision`** | `needs-operator-decision` — the card itself frames this as needing a decision (retire the file vs. add a pointer), consistent with the file's actual orphan status observed here. |
| t320 | `falsified` | `not-landed` | **`needs-operator-decision`** | `needs-operator-decision` — the falsified cause suggests re-scoping rather than dropping outright: either narrow the card to "confirm against the exact binary version used" or re-open it as "audit every ReleaseIntegrationLock erro |
| t323 | `holds` | `not-landed` | **`keep`** | `keep` — mechanical claim reproduced exactly as stated; the card frames the choice (a)/(b)/(c) as needing an operator decision, which I did not adjudicate. |
| t324 | `holds` | `not-landed` | **`keep`** | `keep` — operator-flagged decision card (`[운영자 판정 2026-08-27]` prefix in the card text itself), premise reproduces exactly, design questions remain genuinely open. |
| t325 | `holds` | `not-landed` | **`keep`** | `keep` — hypothesis remains open and structurally plausible; card correctly declines to overstate to a confirmed finding. |
| t327 | `holds` | `not-landed` | **`keep`** | `keep`, with a note attached before dispatch: correct the location citation to `internal/mx/provenance.go:223-227`. |
| t329 | `holds` | `not-landed` | **`keep`** | `keep` — reproduces exactly, scope is already well-bounded by the card itself. |
| t337 | `holds` | `in-flight-unlanded` | **`needs-operator-decision`** | `already-landed` is NOT proposed (verified not an ancestor of develop, exit 1); given a live worktree already exists and the premise independently holds, `needs-operator-decision` is proposed only to confirm whether the in-flight  |
| t339 | `holds` | `not-landed` | **`keep`** | `keep` — all three named defects independently reconfirmed present at HEAD; the fix is a small, scoped documentation edit to a closed SPEC's plan/spec artifacts. Rests on the three verbatim greps above. |
| t344 | `holds` | `not-landed` | **`needs-operator-decision`** | `needs-operator-decision` — rests on the fact that the card's own two branch options (a: generalize, b: treat as closed) are both still open per the evidence above; a sweep worker choosing between them would be making the operator |
| t345 | `holds` | `not-landed` | **`needs-operator-decision`** | `needs-operator-decision` — the card itself frames three sub-decisions (what counts as observation, whether it discriminates audit-absorption from authoring-absorption, whether it's cheaper than manual reading) that are genuinely  |
| t347 | `holds` | `landed` | **`already-landed`** | `already-landed` — rests on commit `37263c222`'s subject line and body explicitly naming "card t347" and "state-table delivery column" as delivered, verified as an ancestor of pinned develop. |
| t348 | `holds` | `not-landed` | **`keep`** | `keep` — rests on the completed SPEC's own plan.md/design.md explicitly naming this exact follow-up candidate as out-of-scope and unbuilt. |
| t353 | `holds` | `not-landed` | **`keep`** | `keep` — rests on the verdict.md text confirming the deferral is real, deliberate, and explicitly awaiting a resume trigger not yet checked by this sweep. |
| t359 | `holds` | `not-landed` | **`keep`** | `keep` — rests on t331-A's landing (verified ancestor of pinned develop) having satisfied the precondition, and D1/D2's underlying requirement text matching the card's citations exactly. |
| t360 | `holds` | `not-landed` | **`keep`** | `keep` — rests on the one directly-verified call site plus the UI lock function matching the card's description exactly, and t350's own commit message naming this as a real, unaddressed follow-up. |
| t361 | `unverified` | `not-landed` | **`keep`** | `keep` — rests on the structural evidence above (comment text + switch scope) matching the card's attribution chain steps 3-4 exactly; step 1's reproduction remains undone by both the card's author and this sweep. |
| t363 | `holds` | `not-landed` | **`keep`** | `keep` — rests on the verbatim `concurrency.group` expression and the `push`/`pull_request` trigger blocks in `ci.yml`, both directly re-read at HEAD. |

### §5.1 The rows that actually need the operator, grouped

**Already resolved — the queue has not caught up (5).** `t201 t243 t247 t313 t347`
Each rests on a cited commit or PR state, not on a resemblance. `t313` and `t347` merged with
`--is-ancestor` exit 0 against pinned develop; `t201`'s commit is an ancestor of both pinned refs;
`t243` and `t247` are resolved through a mechanism other than a card-id-citing commit (a git history
that refutes the premise, and a MERGED PR respectively) and say so in their entries.

**Premise contradicted by measurement (11).**
`t196 t201 t224 t233 t240 t243 t247 t254 t263 t281 t320`
Falsified is not the same as droppable — only `t254` and `t281` are proposed for `drop`, because
only there does the whole card fall with its premise. The rest are stale in wording while a real
concern survives underneath, and rewriting versus closing is the operator's call.

**Needs a decision that is genuinely not the sweep's (21).**
`t90 t125 t154 t196 t204 t216 t223 t224 t233 t236 t240 t248 t260 t263 t286 t287 t319 t320 t337 t344 t345`
Three recurring reasons: the card asks for a scope judgment by design (t344, t345); the deciding
evidence is outside a static read — a live runtime trace, a GitHub issue body, a corpus scan
(t236, t286, t287, t260); or work is already in flight on an unmerged branch and the queue entry may
be redundant with it (t154, t216, t337).

**Keep, unchanged (34).** The remaining cards' premises hold, nothing has landed against them, and
no overlap displaces them.

**A recorded relation cannot be annotated with later evidence — a disposition proposal for the
relation itself, not for any card.** `t313 contains t295` was recorded 2026-08-27. This sweep
measured that t313 has since landed and did **not** resolve t295. The attempt to append that
measurement was refused (`invocations.log`, 2026-08-30T11:41:00Z, cwd
`/Users/goos/MoAI/moai-adk-go`):

```
$ moai todo relate t313 t295 --relation contains --note "<M4 measurement note, full text in 02-overlaps.md §3>"
mutation refused: todo relate: t313 contains t295 is already recorded.
```

No write occurred; zero relations were recorded by this sweep. The guard keys on the ordered triple
(subject, related, relation) and refuses regardless of the note, so **a relation is frozen at its
original claim from the moment it is recorded** — later evidence cannot be attached to it. There is
no verb in this card's own deliverables that can update it: the sweep is read-only by construction
(spec.md §C), and no non-mutating update path exists. Whether to delete the relation and re-record
it with the measured substance is therefore an operator decision, stated here as a proposal and
nothing more — it does **not** add a 63rd row to §5's table, which stays at exactly 62 (AC-BH-016).
Smaller observation kept alongside it: the refusal message names `backlog.json`, the file the queue
stopped writing at the SQLite migration `3cb258d62`, rather than the live `backlog.db`.

## §6 Evidence sections

**Claim.** All 62 in-scope cards were read against one queue snapshot and one pinned pair of refs;
each carries a three-valued premise verdict, a landing verdict citing its deciding commands, the
five evidence sections, and a disposition proposal. No card was mutated; the three independent
no-mutation observables agree, and the single queue difference is attributed to another actor with a
cited commit. Zero relations were recorded, because no pair qualified for the four-token vocabulary.

**Evidence.** `00-tooling-baseline.md` (fetch, pinned SHAs, `strings` = 0, t342 controls, worktree
list, the SQLite-migration probes); `01-scope.md` (snapshot counts, the 68→62 delta reconciliation,
opening and closing digests, the id-level attribution and the byte-identical in-scope diff);
`cards/batch-1..6.md` (62 entries); `02-overlaps.md` (candidate table, the falsified `t313 contains
t295` substance, the verbatim `relate` refusal); `invocations.log` (both invocations verbatim).
Mechanical checks run over the assembled set: entry count 62, zero duplicate ids across batches,
extracted id set `diff`-identical to the in-scope list, intersection with picked ∪ dropped empty,
and all eight required blocks present 62 times each.

**Baseline-attribution.** Every git query against `origin/develop` `ee50984abe4f11ac337382b48a26328f091e200a`
and `origin/main` `48239c7dc7428c8751a04f6321887c2d36123884`, both fetched once at
2026-08-30T11:16:22Z and used as literals throughout — no landing verdict names a branch. Source
reads at worktree HEAD `6165f9f5e`. Card text from `queue-snapshot-run.tsv` captured
2026-08-30T11:16:45Z. Store digests from `/Users/goos/MoAI/moai-adk-go/.moai/state/todo/backlog.db`,
resolved both times via `git rev-parse --path-format=absolute --git-common-dir` →
`/Users/goos/MoAI/moai-adk-go/.git`, parent recorded at each capture.

**Gaps.**
- **The cross-comparison was not all-pairs over full card text.** It ran over the 62 premise
  restatements plus targeted reads of the cards named as candidates — not 1,891 full-text pairs. A
  duplicate whose two restatements do not resemble each other would be missed.
- **Nine cards are `unverified` and stay that way.** Each names why: a live runtime trace, a GitHub
  issue body, or a corpus scan outside a static read. None was promoted to `holds`.
- **Two `in-flight-unlanded` verdicts were completed by the orchestrator, not the worker.** t216's
  `--is-ancestor` was left as a self-declared Gap by its worker and filled in post-check; t154 was
  reclassified from `unknown` after the orchestrator measured its tip. Both corrections are stated
  in place in the entries.
- **The `findings` negative control was not exercised live**, because zero relations were recorded.
  It remains established only by the scratch-copy probe.
- **Landing was decided by commit-message id citation.** Work that landed without naming its card id
  reads `not-landed` here. Three entries caught this shape by reading matched commits in full and
  ruling them mentions rather than deliveries; a delivery that cites no id at all is invisible to the
  method, and `t243`/`t247` are the two cases where a different mechanism was needed to see it.
- **The pinned refs bound the sweep in time.** Anything landing on develop after 2026-08-30T11:16:22Z
  is outside it by construction.

**Residual-risk.**
- **The disposition proposals are readings of a moving queue.** The lead dispatched seven cards
  between plan and run and closed one during the run; a proposal here can be overtaken before the
  operator reads it. Each row cites its evidence so a stale row is falsifiable rather than merely
  doubtful.
- **`keep` is the weakest verdict in this report.** It means nothing displaced the card within a
  bounded read — not that the card is well-formed, correctly scoped, or worth doing. The sweep
  measured premises, not value.
- **Adjacency without a relation token is unrecorded scheduling risk.** t295/t297 (worktree
  launcher) and t324/t325 (the develop→main boundary) are close enough that dispatching them in
  parallel could collide, and none of the four relation tokens says "schedule these apart", so
  nothing in the queue will warn about it.
- **This report is authored by the actor whose restraint it certifies** for two of the three
  observables. Only the card-row digest is independent of it, and that one is now a SQLite read
  whose correctness rests on the four controls in `00-tooling-baseline.md` rather than on this
  report's own assertion.

## §7 Post-merge re-measurement against a moved pin (integration window, 2026-08-30)

The landing verdicts of §5 were decided against pinned `origin/develop`
`ee50984abe4f11ac337382b48a26328f091e200a`. By the time this card reached its integration window
that ref had advanced to `52c3fe590e3ea11b37389d4248162055f22f1c59` — **33 commits**. The lead
directed a re-run inside the merge tree, with the result recorded **beside** the original verdict
rather than over it: which card flipped in which interval is itself an output of this sweep.

**Nothing in §5 is rewritten by this section.** §5 remains the reading as of the original pin, and
this section is the reading as of the new one.

### §7.1 Monotonicity, and why only one direction needed re-running

```
$ git merge-base --is-ancestor ee50984abe4f11ac337382b48a26328f091e200a 52c3fe590e3ea11b37389d4248162055f22f1c59
$ echo $?
0
```

The old pin is an ancestor of the new one, so a commit that was reachable from the old pin is still
reachable from the new one: **no `landed` verdict can flip to `not-landed`.** Only the
`not-landed` → `landed` direction was re-measured. Deductive claims are not left un-probed, so one
control was re-run against the new pin: t313's landing commit `62ff3c2e6` → exit **0**, still landed.

### §7.2 Method — the same method A, dumped once instead of looped

The deciding query is unchanged (`git log <pinned-sha> --grep` per card id, plus
`merge-base --is-ancestor` for in-flight tips). One mechanical deviation, stated because it changes
how the evidence is reproduced and not what it decides: this session runs inside a worktree whose
guard refuses a `git` call placed inside a shell loop, so the delta's commit messages were dumped
once —

```
$ git log --format='COMMIT %h%n%s%n%b%n---END---' \
    ee50984abe4f11ac337382b48a26328f091e200a..52c3fe590e3ea11b37389d4248162055f22f1c59 \
    > .moai/reports/t332/03-delta-commits.txt
$ grep -c '^COMMIT ' .moai/reports/t332/03-delta-commits.txt
33
```

— and each of the 62 in-scope ids searched in that file with a word-boundary regex (`\bt<NNN>\b`).
The corpus is the same set of commit messages `--grep` reads; only the number of `git` invocations
differs. `03-delta-commits.txt` is committed alongside this report, so the search is re-runnable.

### §7.3 Result — 0 flips of 62

| Card | §5 landing verdict (pin `ee50984ab`) | §7 re-measurement (pin `52c3fe590`) | Deciding evidence |
|---|---|---|---|
| t201, t313, t347 | `landed` | **`landed`** (unchanged) | monotonic under §7.1; control re-run for t313 → `--is-ancestor 62ff3c2e6 52c3fe590` exit `0` |
| t154 | `in-flight-unlanded` | **unchanged** | `--is-ancestor dbb87f14f 52c3fe590` → exit `1`; `git worktree list` tip still `dbb87f14f` (`WT-lint-heading`) |
| t216 | `in-flight-unlanded` | **unchanged** | `--is-ancestor 8aa96bfb1 52c3fe590` → exit `1`; tip still `8aa96bfb1` (`WT-hook-wiring-drift`) |
| t337 | `in-flight-unlanded` | **unchanged** | `--is-ancestor c72a517c3 52c3fe590` → exit `1`; tip still `c72a517c3` (`WT-windows-stamp-liveness`) |
| t224 | `unknown` | **unchanged** | no commit in the 33-commit delta references `t224`; the reason recorded in §5 (the requested doctrine text may predate the card) is untouched by the delta |
| t363 | `not-landed` | **`not-landed`** — one grep hit, ruled a mention | see §7.4 |
| the other 55 | `not-landed` | **unchanged** | zero grep hits across the 33-commit delta |

**Flips: 0 of 62.** The batch's 33 commits closed `picked` cards (t294, t343, t357, t358 and
siblings), and the in-scope set is the `queued` cards — a flip would require a queued card's work to
land incidentally, which none did.

### §7.4 The one grep hit, read in full and ruled a mention

`t363` matched two commits. Reading them rather than counting them:

```
$ git log -1 --format='%h %s%n%b' 66a44c278
66a44c278 docs(SPEC-CI-PR-TRIGGER-FILTER-001): name t363 as the axis-B inheritor in CHANGELOG (t294)
The deferral was recorded without the receiving card id, so a reader
could not follow the duplicate-run work to where it actually lives.
```

The commit names t363 as the card that will **receive** deferred work — it records where the work
lives, and does not do it. The second hit, `d7010f86a`, is the merge commit carrying the first. Same
discipline as §5's three mention-only hits: a card-id match is a mention until the commit is read.
`t363` stays `not-landed`, disposition `keep`.

### §7.5 Queue movement during the window, attributed

The closing capture recorded 96 store rows; at re-measurement the store holds **100**.

```
$ sqlite3 -json /Users/goos/MoAI/moai-adk-go/.moai/state/todo/backlog.db \
    "select id,state,text from items order by id;" | shasum -a 256
5892574384191c85d58917b224955e5bda8b4e15dfee24f6e4d5f1aacb131e94  -
```

| Difference | Rows | Attribution |
|---|---|---|
| admitted | `t364` `t365` `t366` `t367` | the lead, in its integration-window message: five cards `t363`–`t367` admitted (`t363` was already present at the closing capture) |
| state change | `t347` `queued` → `picked` | the lead, same message: it moved `t347` and `t366` to `picked` |
| removed | none | — |

**This is queue movement by another actor, not a mutation by this sweep.** The sweep issued no
`moai todo` invocation in the integration window; `invocations.log` still carries exactly two lines,
and the no-mutation observables of §4 are unaffected by rows this sweep never touched.

One consequence for the operator reading §5: **`t347`'s row proposes `already-landed` while the card
was `queued`, and it is now `picked`.** The proposal is overtaken by the lead's own act — no action
is required on it.
