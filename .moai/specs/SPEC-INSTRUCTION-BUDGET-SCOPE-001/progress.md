# Progress — SPEC-INSTRUCTION-BUDGET-SCOPE-001

Card: t1180 · Branch: `WT-rules-40k-split` · **Branch base: `088594d6b`**

The pin is the SHA, never a branch name. `088594d6b` was `develop`'s tip when this worktree was created; it is an **ancestor** of `develop` now, which is ahead of it. `develop` moved repeatedly during this card, and the target files were measured unchanged at each tip that was read — `git diff --stat 088594d6b develop -- <the four files, both coding-standards copies, the hook, the template rules mirror>` returned empty every time, re-measured rather than carried forward. So the SHA is a live measurement of those files while a branch name for them expires without notice.

Absorption of local `develop` plus **re-measurement in the merged tree** is owed at the integration window per `CLAUDE.local.md` §4.1 — unconditionally, not because anything diverged. The empty diff predicts the absorption will be uneventful; it does not make it unnecessary.

No commit count appears in this header on purpose (plan.md §G).

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M**, declared in `spec.md` frontmatter (`tier: M`) as of v0.5.0. Before that the field was absent and the machine read Tier **L** — the correction is the first item below.
- Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (the Tier M set).
- Requirements: **15**, `REQ-IBS-001` … `REQ-IBS-015`, continuous with no gaps, GEARS notation.
- Acceptance criteria: **8** (`AC-IBS-001` … `AC-IBS-008`) against the Tier M ceiling of 16. The count fell from 16 by folding six rows plus one condition into `AC-IBS-002`; `acceptance.md` § Folding record carries the per-row justification and no check was dropped.
- Criterion classes: 3 release-blocking (each with a RED-now cell observed on this tree), 4 regression-guard (green at arrival — not recorded as a pass), 1 process check.
- `moai spec lint` on this SPEC: `✓ No findings`, real exit `0` (measured without a pipe; a piped `; echo $?` reports `tail`'s status, not lint's).
- Status: `draft`. Run phase not entered.

### Plan-audit iteration 1 — FAIL, and what it changed

Verdict **FAIL** at 0.825 arithmetic / 0.79 harmonic against the Tier M threshold of 0.80. The score was not what decided it: the MUST-PASS firewall did, on REQ-number gaps plus three MUST-PASS rows whose deciding commands could not judge their own conditions. Report: `.moai/reports/t1180/plan-audit.md` (this worktree; the primary checkout carries only `verdict.md`). Dimension scores: Clarity 0.85, Completeness 0.90, **Testability 0.55**, Traceability 1.00.

Repairs landed in v0.5.0:

| # | Defect | Repair |
|---|---|---|
| D7 | `tier:` absent → machine read Tier L | `tier: M` added. This one came first because it changes what the other numbers mean: at Tier L the ceiling reads 25/25 and the PASS threshold 0.85, so the entire v0.3.0 "Tier M budget repair" had been performed toward a tier the frontmatter never declared. |
| D2 | `kanban-dispatch*` census said 2 members | Corrected to **3**, with the family's per-trigger total measured at 81,159 characters (spec.md §1). |
| D3 | `kanban-dispatch-mechanics.md` called a REQ violation | Half applied, half corrected — see § Corrections below. |
| D1 | REQ ids skipped 005/006/014 | Renumbered continuously `001`-`015`. The gap was an asymmetry rather than a cosmetic issue: AC had been renumbered continuously in the same v0.3.0 pass, so the cost of continuity was already paid on one axis and withheld on the other. |
| D4 | `AC-IBS-006`'s `git diff --stat` could not attribute a new file to a parent | Replaced by the **companion roster** (`AC-IBS-002-R`): M2 records the authorised companion filenames before content moves, and the check is a set difference against the new-file set. |
| D5 | `AC-IBS-011b` cited a baseline its own command did not produce | Baselines split per command (`acceptance.md` § Reference-count scope note), reverse-order `§` form added, bound widened 40 → 200 characters. |
| D6 | affinity conditions only checked the author's own claim | `AC-IBS-002-A-falsifier` added — the one condition that can falsify an affinity claim rather than check it for internal consistency. |
| F3 | the `ls`-based decider | Replaced. In this environment `ls` is aliased to long format, so `ls … \| grep -c '^kanban-dispatch'` returned `0` — neither the expected 2 nor the correct 3. |
| F7 | `REQ-IBS-009` carried a declarative middle clause | Trimmed; the arithmetic moved to §1 where it was already established. |

### Corrections to the audit — two findings not applied

Both are recorded in `spec.md` §6 with their measurements. In brief:

- **D3's parentage.** `kanban-dispatch-mechanics.md` is a **sibling** of `-detail.md`, not its child; both are companions of `kanban-dispatch.md`, which carries no `paths:` frontmatter at all. The proper-subset test compares a companion against *its parent's* pattern set, so it is inapplicable here and the REQ-IBS-006 violation is not substantiated. D3's other half — that all three files co-load on `**/kanban-dispatch*.md` — stands and is now the measured precedent for the naming constraint.
- **Residual-risk (1) — direction inverted.** The report reads the family as already diverged on `develop`. Measured: this base is an **ancestor** of `develop`, and `develop` fully contains `main`, which is further behind and carries the older two-file state. The primary checkout sits on `main` — the intended steady state here — so reading it produced the older tree and the inference ran backwards. **Absorption is still owed**: the `CLAUDE.local.md` §4.1 lane duty requires absorbing local `develop` and re-measuring in the merged tree regardless of what the diff says. The empty target-file diff predicts an uneventful absorption; it does not make one unnecessary. (A v0.6.0 version of this bullet called the absorption a no-op, which was wrong for the reason recorded in the v0.7.0 HISTORY row.)

The reusable point, and the reason it is recorded rather than just acted on: in this repository the primary checkout and the branch base are far apart by design, so a measurement's tree must be **named** rather than assumed. No divergence risk is recorded in the SPEC, because there is no divergence — one branch is behind the other, which is the normal state here; `kanban-dispatch-mechanics.md` and the `gtd.md` pattern exist on `develop` and reach `main` through the release PR.

**This lesson has now fired three times in one card, twice against me.** Once when I read the lead's stale measurement as a broken read path; once by the auditor reading the primary checkout as the branch base; and once when I wrote that `develop` **equals** this base — in the same document that carries the warning. The generalisation is in `plan.md` §G rather than only here, because a hazard recorded as a HISTORY note is read once and a hazard recorded as an anti-pattern is read every time someone edits the section it guards: **a moving reference in prose is a claim with an expiry date, and the expiry is invisible in the text.**

### Measurement conventions adopted this iteration

- **Character counts** use the single-invocation form `python3 -c "import io;print(len(io.open('<path>',encoding='utf-8').read()))"`. The prior `CHARS()` macro took `<path>` on **stdin redirect**, which `verification-completeness.md` §2.1 places outside the RED-now command form; the macro's parameter was what pushed the shape toward a redirect, so the macro is gone and each row names its own file.
- **No `go run` behind an exit-code predicate.** `go run` collapses every nonzero exit to 1 — measured: `moai spec lint` on a nonexistent SPEC exits `3`, and `go run` reported `1`. No deciding command in the artifacts uses `go run` (`grep -n "go run"` across all three files → no output).
- **Non-conforming command forms counted properly**: 4 of 30 command-shaped spans in the prior `acceptance.md` (one redirect, three pipes), not "roughly half". The first pass over-flagged 12; two of my own exclusions were wrong and were corrected. Recording the sequence rather than only the figure, because a count that lands right on the first try usually means the test was too loose — the excluded case that mattered was a `|` inside a single-quoted regex alternation, which required reading the quoting rather than matching the character.

### Gaps at plan close

- `go test ./internal/template/...` and `./internal/hook/...` (AC-IBS-006) are **not measured**. Both are heavy package suites and `.claude/rules/local/gitflow-lane-protocol.md` §8 requires a resource-slot lease before a heavy run. Measured in run phase under a lease.
- The **neutrality sentinel demonstration** is not executed. AC-IBS-002-G's "two registries, not three" rests on a structural read of `findNeutralityRoot` (`:200-222`), the `.md`-inclusive extension map (`:87-94`), and the absence of a directory skip in the walk bodies. M2 converts it to a measurement with an anchored selector.
- The **`.moai/docs/**` hook-scope premise** (M6 ladder rung 2) is not measured, and the ladder states that measuring it is part of the rung rather than a precondition someone might skip.
- **Corpus-wide `moai spec lint`**: errors are closed by the auditor's run (0 errors across 963 SPECs, from lint's own stdout); the 8,204 warnings are pre-existing corpus state attributable to nothing in this card. My own earlier background attempt established nothing — it was piped to `grep`, so the exit code reported was the pipe's and the per-SPEC lines were discarded.
- The `[HARD]` token sum `4 + 21 + 8 + 12 = 45` is **carried from M0, not re-measured this iteration**. AC-IBS-002-H uses it only as corroboration and forbids citing it alone, so nothing decides on it.

### v0.6.0 — three lead addenda, each measured

- **Unit discipline.** Every reference-count figure now states its unit. The audit's comparable figures (74 vs 70, and 11) are **line** counts; mine are **file** counts, and mixing them is the same defect class as mixing commands. Measured consequence: `comm -23` of the reverse-order file list against the forward list gives **0 for all four files** — every file carrying a reverse-order `§` reference also carries a forward-order one — and widening the bound 40 → 200 leaves the file counts identical at 44/22/8/7. So **both** audit blind spots live inside the forward file set as per-reference gaps, not as missing files. Both are still worth enumerating; neither changes the baseline.
- **Per-target zero rule.** `AC-IBS-004a` records `N/A (0 references measured)` rather than PASS where its sweep is empty. The sweep is `16 / 0 / 0 / 0`, and the zeros land on `worktree-integration.md` (M4) and `kanban-dispatch-detail.md` (M5) — the only two files this SPEC splits. So the row is inert on every file where a broken anchor could be introduced and live only on the one that gets no split. `plan.md` §G carries the composition as a named anti-pattern, because three green sub-conditions would otherwise report the split as proven.
- **Pin discipline.** The document pin is the branch-base SHA, never the branch name (see the header). `AC-IBS-004c` was also tightened: a filename `grep -c` satisfies a pointer that names the companion without naming the relocated section, leaving the inbound `§` reference dangling — so the check is now two conditions read off the same matched line.

Also this iteration: REQ-IBS-006 explicitly barred from the always-loaded-stub shape, where `⊊` is undefined because the parent has no pattern set at all; the stub's 34,901 characters recorded as a **per-session** cost that no companion split reduces; and a dangling `§B.5` → `§B.4` repaired in the plan that defines the cross-reference check — a self-application failure in the document specifying that every cross-reference resolves.

### Artifact state

Measured this turn. All four files untracked (`?? .moai/specs/SPEC-INSTRUCTION-BUDGET-SCOPE-001/`), worktree HEAD `088594d6b`.

| file | version |
|---|---|
| `spec.md` | 0.7.0 |
| `plan.md` | 0.7.0 |
| `acceptance.md` | 0.7.0 |
| `progress.md` | (this write) |

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
