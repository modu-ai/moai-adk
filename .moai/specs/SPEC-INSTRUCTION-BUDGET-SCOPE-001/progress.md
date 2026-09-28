# Progress — SPEC-INSTRUCTION-BUDGET-SCOPE-001

Card: t1180 · Branch: `WT-rules-40k-split` · **Branch base: `088594d6b`**

The pin is the SHA, never a branch name. `088594d6b` was `develop`'s tip when this worktree was created; it is an **ancestor** of `develop` now, which is ahead of it. `develop` moved repeatedly during this card, and the target files were measured unchanged at each tip that was read — `git diff --stat 088594d6b develop -- <the four files, both coding-standards copies, the hook, the template rules mirror>` returned empty every time, re-measured rather than carried forward. So the SHA is a live measurement of those files while a branch name for them expires without notice.

Absorption of local `develop` plus **re-measurement in the merged tree** is owed at the integration window per `CLAUDE.local.md` §4.1 — unconditionally, not because anything diverged. The empty diff predicts the absorption will be uneventful; it does not make it unnecessary.

No commit count appears in this header on purpose (plan.md §G).

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M**, declared in `spec.md` frontmatter (`tier: M`) as of v0.5.0. Before that the field was absent and the machine read Tier **L** — the correction is the first item below.
- Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (the Tier M set).
- Requirements: **15**, `REQ-IBS-001` … `REQ-IBS-015`, continuous with no gaps, GEARS notation.
- Acceptance criteria: **8 logical criteria** (`AC-IBS-001` … `AC-IBS-008`) against the Tier M ceiling of 16. The count fell from 16 by folding six rows plus one condition into `AC-IBS-002`; `acceptance.md` § Folding record carries the per-row justification and no check was dropped.
  - **Unit matters, and a second count exists.** The repo's `ac-baseline-guard` pre-commit hook counts **distinct live identifiers including a letter suffix** and reports **10** for this file (`AC-IBS-002a` and `AC-IBS-002d` are prefixed identifiers in their own right). `8` binds the Tier ceiling; `10` is what the sync-phase CHANGELOG convention will count (`manager-develop-prompt-template.md` § B12). Both are recorded in `acceptance.md` with their commands, so sync does not have to adjudicate a discrepancy nobody flagged.
  - Surfaced by the hook's own report line on the v0.7.0 commit — `no count moved; unrecorded (report only): … COUNT 10` — and reconciled rather than dismissed. This is the **fourth** unit mismatch in this card, after line-vs-file reference counts (twice) and `grep -c` lines vs occurrences during the `develop` enumeration.
- Criterion classes: 3 release-blocking (each with a RED-now cell observed on this tree), 4 regression-guard (green at arrival — not recorded as a pass), 1 process check.
- `moai spec lint` on this SPEC: `✓ No findings`, real exit `0` (measured without a pipe; a piped `; echo $?` reports `tail`'s status, not lint's).
- Status: `draft`. Run phase not entered.

### Plan-audit iteration 1 — FAIL, and what it changed

Verdict **FAIL** at 0.825 arithmetic / 0.79 harmonic against the Tier M threshold of 0.80. The score was not what decided it: the MUST-PASS firewall did, on REQ-number gaps plus three MUST-PASS rows whose deciding commands could not judge their own conditions. Report: `.moai/reports/t1180/plan-audit.md` (this worktree; the primary checkout carries only `verdict.md`). Dimension scores: Clarity 0.85, Completeness 0.90, **Testability 0.55**, Traceability 1.00.

Repairs landed in v0.5.0:

| # | Defect | Repair |
|---|---|---|
| D7 | `tier:` absent → machine read Tier L | `tier: M` added. This one came first because it changes what the other numbers mean: at Tier L the ceiling reads 25/25 and the PASS threshold 0.85, so the entire v0.3.0 "Tier M budget repair" had been performed toward a tier the frontmatter never declared. |
| D2 | `kanban-dispatch*` census said 2 members | Corrected to **3** at the time, with a per-trigger total of 81,159 characters. **Both figures were superseded in iteration 3**: the census had been taken by filename prefix, which misses files that HOLD the pattern without sharing the prefix. By `grep -rl` the family has **4** co-loaders totalling **100,165** characters, of which 65,264 are reducible (spec.md §1). |
| D3 | `kanban-dispatch-mechanics.md` called a REQ violation | Half applied, half corrected — see § Corrections below. |
| D1 | REQ ids skipped 005/006/014 | Renumbered continuously `001`-`015`. The gap was an asymmetry rather than a cosmetic issue: AC had been renumbered continuously in the same v0.3.0 pass, so the cost of continuity was already paid on one axis and withheld on the other. |
| D4 | `AC-IBS-006` ⚠️ **pre-renumbering id — see the note below** — its `git diff --stat` could not attribute a new file to a parent | Replaced by the **companion roster** (`AC-IBS-002-R`): M2 records the authorised companion filenames before content moves, and the check is a set difference against the new-file set. |
| D5 | `AC-IBS-011b` (pre-renumbering; no current id) cited a baseline its own command did not produce | Baselines split per command (`acceptance.md` § Reference-count scope note), reverse-order `§` form added, bound widened 40 → 200 characters. |
| D6 | affinity conditions only checked the author's own claim | `AC-IBS-002-A-falsifier` added — the one condition that can falsify an affinity claim rather than check it for internal consistency. |
| F3 | the `ls`-based decider | Replaced. In this environment `ls` is aliased to long format, so `ls … \| grep -c '^kanban-dispatch'` returned `0` — neither the expected 2 nor the correct 3. |
| F7 | `REQ-IBS-009` carried a declarative middle clause | Trimmed; the arithmetic moved to §1 where it was already established. |

**Pre-renumbering ids in the table above — and one collision that matters more than the dangling ones.** Every `AC-IBS-*` id in this dated table is from the pre-v0.3.0 scheme. Most simply no longer resolve (`AC-IBS-011b`), which is visible to any reader who looks them up.

`AC-IBS-006` is the dangerous case: that id **exists today** and denotes something else entirely — the template-and-hook-packages row. So the D4 row above reads as a live cross-reference and silently points at the wrong criterion. A dangling id announces itself; a collided id does not, and it is the one a reader is likely to act on.

The ids are **marked rather than rewritten**, on the principle already stated in `spec.md`'s v0.5.0 HISTORY row: a dated record describes what was done at the time, and renumbering its identifiers would make it agree with the present at the cost of no longer describing the past. The current criterion set is `AC-IBS-001` … `AC-IBS-008` in `acceptance.md`, which is the document to resolve any id against.

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
- ~~The **neutrality sentinel demonstration** is not executed.~~ **CLOSED at plan phase.** A `C1-macos-bias-path` sentinel injected into the mirror of `spec-workflow.md` was reported by the anchored audit — `TEMPLATE_NEUTRALITY_VIOLATION: class=C1-macos-bias-path file=.claude/rules/moai/workflow/spec-workflow.md`, exit `1`; pre-check exit `0` on a clean target, exit `0` again after revert, tree clean. So `AC-IBS-002-G`'s "two registries, not three" is measured rather than read, and REQ-IBS-011 rests on an executed premise. Full table with every command and exit code: `plan.md` §B.4.
  - Closed at plan phase rather than deferred because it is the only gap bearing on an obligation the SPEC **imposes** (which registries a companion must join) rather than one it merely measures. Had the walk carried an unfound filter, REQ-IBS-011 would have required the wrong registrations and an audit would have scored a false premise.
  - Run without a resource-slot lease, deliberately: the anchored single-test selector is not the package suite that `gitflow-lane-protocol.md` §8's lease discipline targets. The unanchored suite remains un-run and stays a Gap below.
- The **`.moai/docs/**` hook-scope premise** (M6 ladder rung 2) is not measured, and the ladder states that measuring it is part of the rung rather than a precondition someone might skip.
- **Corpus-wide `moai spec lint`**: errors are closed by the auditor's run (0 errors across 963 SPECs, from lint's own stdout); the 8,204 warnings are pre-existing corpus state attributable to nothing in this card. My own earlier background attempt established nothing — it was piped to `grep`, so the exit code reported was the pipe's and the per-SPEC lines were discarded.
- The `[HARD]` token sum `4 + 21 + 8 + 12 = 45` is **carried from M0, not re-measured this iteration**. AC-IBS-002-H uses it only as corroboration and forbids citing it alone, so nothing decides on it.

### v0.6.0 — three lead addenda, each measured

- **Unit discipline.** Every reference-count figure now states its unit. The audit's comparable figures (74 vs 70, and 11) are **line** counts; mine are **file** counts, and mixing them is the same defect class as mixing commands. Measured consequence: `comm -23` of the reverse-order file list against the forward list gives **0 for all four files** — every file carrying a reverse-order `§` reference also carries a forward-order one — and widening the bound 40 → 200 leaves the file counts identical at 44/22/8/7. So **both** audit blind spots live inside the forward file set as per-reference gaps, not as missing files. Both are still worth enumerating; neither changes the baseline.
- **Per-target zero rule.** `AC-IBS-004a` records `N/A (0 references measured)` rather than PASS where its sweep is empty. The sweep is `16 / 0 / 0 / 0`, and the zeros land on `worktree-integration.md` (M4) and `kanban-dispatch-detail.md` (M5) — the only two files this SPEC splits. So the row is inert on every file where a broken anchor could be introduced and live only on the one that gets no split. `plan.md` §G carries the composition as a named anti-pattern, because three green sub-conditions would otherwise report the split as proven.
- **Pin discipline.** The document pin is the branch-base SHA, never the branch name (see the header). `AC-IBS-004c` was also tightened: a filename `grep -c` satisfies a pointer that names the companion without naming the relocated section, leaving the inbound `§` reference dangling — so the check is now two conditions read off the same matched line.

Also this iteration: REQ-IBS-006 explicitly barred from the always-loaded-stub shape, where `⊊` is undefined because the parent has no pattern set at all; the stub's 34,901 characters recorded as a **per-session** cost that no companion split reduces; and a dangling `§B.5` → `§B.4` repaired in the plan that defines the cross-reference check — a self-application failure in the document specifying that every cross-reference resolves.

### Plan-audit iteration 2 — FAIL, and the ten repairs of iteration 3

Iteration 2 returned FAIL. Iteration 3 was granted past the Tier M ceiling of 2 on the lead's judgement: the score drop was **audit-coverage expansion, not quality decay** — Traceability held at 1.00, every must-pass item passed, and each defect was a wrong sentence or a mis-scoped criterion rather than a structural fault. Standing condition: **if iteration 3 scores under 0.80, stop** — no further iteration, report, and the decision goes to the operator.

| # | Defect | Repair |
|---|---|---|
| D1 | AC-IBS-005 demanded byte-identity of a pair whose divergence is permanent by design | Row split. 005a keeps byte-identity for the four workflow files; 005b asserts the `coding-standards.md` pair's **diff is unchanged from its recorded baseline** (measured: 2 lines, 1 content line, exit `1`), with a positive control. The line is local-only because it carries a SPEC ID + REQ tokens, which `C1-spec-id` excludes from the template — so byte-identity there was **impossible-red**, satisfiable by no correct work. |
| D11 | literal base-SHA pins in range predicates | `CARD_BASE=$(git merge-base develop HEAD)` resolved at read time, plus two parts the audit did not carry: **pre-merge-only** declared (after the merge `merge-base` is the card tip, the range empties, the predicate passes vacuously — measured precedent in `gitflow-lane-protocol.md` §8) and a **non-empty-range control** (measured: 4 files). Applied at **three** sites; the audit named two, and AC-IBS-008(b) is the same class — extension flagged, not silent. |
| D2 | family census counted by filename prefix | Recounted by pattern **holding**. `grep -rl 'kanban-dispatch\*'` surfaces `cross-session-messaging-detail.md`, which shares no prefix; `session-handoff*` has **two** such holders. Totals: 4 files / 100,165 and 4 files / 74,732. Rung 3 reopened for the session-handoff family. |
| D3 | one total conflated reducible and always-loaded characters | Two totals stated separately and never summed: co-loading (100,165 / 74,732) vs reducible (65,264 / 58,623). §4 forbids folding the stub, so its characters are a per-session cost no rung can move. |
| D6 | "no commit count appears in this prose" — refuted six lines above | The `0\t6546` transcript stays (attributed to its command and run, which the carve-out permits); the false claim is replaced by the actual distinction — a transcript carries provenance, a bare sentence does not. |
| D7 | RED cell named `grep -rc` but quoted `-rl` output and the wrong exit code | Measured: `-rc` prints `<file>:0` per scanned file and exits `0`; `-rl` prints nothing and exits `1`. Cell now names `-rl` with its real output, and the mismatch is recorded rather than quietly swapped. |
| D8 | § Artifact state false in the present tense | Every row dated, with its command; the HEAD and count rows qualified, since a figure measured before the commit recording it cannot describe the tree after it. |
| D5 | four pre-renumbering AC ids in live prose | **[corrected 2026-09-28, round 4]** The v0.9.0 disposition on this row — "Marked, not rewritten" — answered a different defect than the one named in the defect column: the marking applied to pre-renumbering ids in HISTORY rows and in this table, while the four **live-prose** sites (spec.md ×3, plan.md ×1) were left untouched. One is a **collision**: the cited id resolves today to a different row, so it read as a live cross-reference pointing at the wrong criterion — more dangerous than the dangling ids, because it does not announce itself. The four live-prose sites are repaired in v0.10.0. |
| D4 | Folding record arithmetic: `16 - 7 = 9`, not 8 | Recorded as **unrecoverable**. v0.4.0's criterion list was replaced in place, so no copy of the 16-row set survives to diff against; whether the eighth disposition was a fold or a drop cannot be established. Recorded rather than reconstructed, because a plausible reconstruction is indistinguishable from a measured one once written. |
| D10 | proportionality claim carried a numerator with no denominator | Denominator declared a **Gap** until M4 measures it with its command. `1 of 2` and `1 of 40` share the numerator and support opposite conclusions. |

**Two earlier defects came back resolved and were not re-touched**: the companion roster closed iteration 1's D4 attribution problem, and the affinity falsifier can genuinely falsify. **One claim was checked and left standing**: `plan.md` §A's "no commit count appears in this section" — measured against the section after D6 taught that self-referential claims need measuring like any other; it holds.

### Plan-audit iteration 3 — FAIL (0.7875 / 0.7619), and the round-4 repair

Verdict **FAIL** at 0.7875 arithmetic / 0.7619 harmonic against the Tier M threshold 0.80 — and, independently of the score, FAIL under the Retry Loop Contract because iteration 2's D5 was unresolved (iteration-3 report F0). Report: `.moai/reports/t1180/plan-audit-iter3.md` (in this worktree). The score improved on both means against iteration 2 and no must-pass item failed; what kept the verdict FAIL was D5 unrepaired while three surfaces claimed it repaired.

Repairs landed in v0.10.0:

| # | Defect (iteration-3 report) | Repair |
|---|---|---|
| F0 | four mis-resolving ids in live prose, unchanged since iteration 2, with three false repair claims | All four sites repaired: `spec.md` trigger-affinity → REQ-IBS-014/015, structural foreclosure → REQ-IBS-007, same-glob shard → REQ-IBS-005; `plan.md` same-glob shard → REQ-IBS-005 + AC-IBS-002-P. Each mapping re-verified against the REQ definitions in `spec.md` §2 in this round, not inherited from the report. Full sweep re-run: every `REQ-IBS-`/`AC-IBS-` occurrence outside HISTORY resolves against the live sets (REQ 001-015; AC 001-008 with letter sub-conditions) or is an explicitly marked pre-renumbering mention in a dated record. The three false claims corrected: the v0.9.0 HISTORY D5 sentence (amended in place with a dated correction), the D5 row above (amended), and the commit message below (footnote). |
| F1 | 005b's byte-equality-against-baseline predicate admitted a mutant in M1's window | Predicate restated as the invariant: exactly one hunk, its single content line byte-equal to the recorded content line, no hunk added or removed; only the hunk header's line numbers may move. The M1 re-run note now states that re-recording the baseline REQUIRES demonstrating the content line unchanged — closing the "record whatever the diff says" mutant. |
| F2 | `acceptance.md`'s headline ("not by dropping checks") contradicted its own Folding record | Restated as what is established: seven dispositions recorded and none of those dropped; one unrecorded and unrecoverable, so the unqualified claim is withdrawn. |
| F3 | the D2 repair (rung 3 reopened) landed in `plan.md` but not in 002d's green path | 002d's green path now names compression, non-rule relocation, and dedup within the co-loading family (the reopened rung 3), keeping "never a split" with REQ-IBS-007 binding — the cross-layer revision sweep `verification-completeness.md` §3 prescribes. |
| F4 | `spec.md` §6 cited a transcript permission `plan.md` §G did not carry | The clause added to `plan.md` §G as the carve-out's second limb (a quoted command transcript MAY carry a figure, attributed to its command; it cannot launder a conclusion), so the citation now points at a rule that exists. |

**Correction footnote — the `171089a77` commit message.** That commit's subject says "ten defects, two controls executed" and its body claims D5 among the repairs. The tree it committed did not contain the D5 repair (iteration-3 report F0, measured). The commit is in history and is not rewritten; this footnote is the correction of record: of the ten, **D5 was not repaired at `171089a77`** — the repair is the v0.10.0 commit.

**Not touched, deliberately**: F5 (control clause binds "range" vs "range or tree" — one word, optional-class, nothing actually ungated per the report) and F6 (the 16's unit at the Folding record — optional-class, moot per the report since the gap is unrecoverable either way). Neither was ordered for round 4; both are left as the report filed them.

### Artifact state

[HARD] **Every row here is a dated measurement, not a standing description.** The previous version of this block read "all four files untracked, worktree HEAD `088594d6b`", in the present tense, and each clause was false within minutes of being written — the files were committed and HEAD advanced six times. A present-tense sentence about mutable state is a claim with an invisible expiry (plan.md §G); in this block the fix is to date every reading and name the command, so a stale row reads as history rather than as an assertion.

**Measured 2026-09-28, plan-audit iteration 3 repair round:**

| reading | command | value |
|---|---|---|
| tracked state | `git status --short` | clean at the iteration-3 repair commit; all four artifacts tracked, no longer untracked |
| worktree HEAD | `git rev-parse --short HEAD` | `a5e983d54` before this round's commit; the commit closing this round supersedes it |
| card base | `git merge-base develop HEAD` | `088594d6b874933f060bcbadbb833737fd25fb4d` — resolved, not pinned (acceptance.md § Card-base resolution) |
| commit count on branch | `git rev-list --count "$CARD_BASE"..HEAD` | 6 before this round's commit |

| file | version (this round) |
|---|---|
| `spec.md` | 0.9.0 |
| `plan.md` | 0.9.0 |
| `acceptance.md` | 0.9.0 |
| `progress.md` | (this write) |

**Measured 2026-09-28, round-4 repair (v0.10.0):**

| reading | command | value |
|---|---|---|
| tracked state | `git status --short` | the four artifacts modified, uncommitted, at the time of this write; committed by the round-4 commit |
| worktree HEAD before this round's commit | `git rev-parse --short HEAD` | `171089a77` (the iteration-3 repair commit this round corrects) |
| card base | `git merge-base develop HEAD` | `088594d6b874933f060bcbadbb833737fd25fb4d` — resolved, not pinned |
| commit count on branch before this round's commit | `git rev-list --count "$CARD_BASE"..HEAD` | 7 |

| file | version (round 4) |
|---|---|
| `spec.md` | 0.10.0 |
| `plan.md` | 0.10.0 |
| `acceptance.md` | 0.10.0 |
| `progress.md` | (this write) |

The HEAD and count rows are deliberately qualified rather than restated as bare values: a figure measured before the commit that records it cannot describe the tree after it, and pretending otherwise is the self-referential hazard the §E.4 SHA-backfill convention exists for.

## §E.2 Run-phase Evidence

### M0 — baseline capture (measured at run start, tree `474667250`; no content change)

Char counts, single-invocation form `python3 -c "import io; [print(f, len(io.open('.claude/rules/moai/workflow/'+f+'.md',encoding='utf-8').read())) for f in [...]]"`:

| File | measured | plan baseline (`088594d6b`) | match |
|---|---:|---:|---|
| `spec-workflow.md` | 40,797 | 40,797 | ✓ |
| `worktree-integration.md` | 61,435 | 61,435 | ✓ |
| `session-handoff-examples.md` | 41,615 | 41,615 | ✓ |
| `kanban-dispatch-detail.md` | 41,034 | 41,034 | ✓ |

Mirror byte-identity (precondition of AC-IBS-005a), `cmp -s <local> <template mirror>` per pair: **all four identical** (exit 0 each). `paths:` globs, verbatim: spec-workflow `**/.moai/specs/**,**/.moai/config/sections/quality.yaml` · worktree-integration `**/.claude/agents/**,**/.claude/worktrees/**,**/.claude/teams/**` · session-handoff-examples `**/session-handoff.md` · kanban-dispatch-detail `**/kanban-dispatch*.md,**/.claude/agents/moai/manager-lead.md,**/.claude/skills/moai/workflows/gtd.md`.

Inbound references (`grep -rl "<name>" .claude/ .moai/docs/ internal/template/templates/`, SPEC dir excluded): spec-workflow **78** · worktree-integration **38** · session-handoff-examples **8** · kanban-dispatch-detail **7** — all four match the plan baseline exactly.

Family censuses by pattern holding (`grep -rl '<pattern>' .claude/rules/`): `kanban-dispatch*` holders = kanban-dispatch-detail, kanban-dispatch-mechanics, cross-session-messaging-detail; `session-handoff*` holders = session-handoff-format, context-window-management-detail. Totals with the stubs: kanban family 34,901 + 41,034 + 5,224 + 19,006 = **100,165** co-loading / **65,264** reducible; session family 16,109 + 41,615 + 5,266 + 11,742 = **74,732** co-loading / **58,623** reducible. Both match spec.md §1 exactly.

`[HARD]` token → named-clause resolution (the transfer-table left column; token count ≠ clause count, REQ-IBS-008):

- **spec-workflow.md — 4 tokens → 4 clauses**: :23 three-phase lifecycle & route triggers (Frozen) · :49 step ordering rules (Frozen) · :134 Tier S/M/L pre-artifact classification (Evolvable) · :164 plan-phase in main checkout, no worktree (Frozen).
- **worktree-integration.md — 18 tokens → 17 clauses + 1 section marker**: :52 L1-exception (`moai worktree new` sole; `done` refuses L1) · :54 unpushed branch = only instance · :56 WT- slug prefix · :58 slug tokens/card-id exclusion · :229 **section marker** over the Selection Rules block — :255 parallel leaf workers MUST isolate · :256 read-only roles MUST NOT · :257 one-shot ≥3-path writers MUST · :260 GitHub fixer agents MUST · :264 auto-isolation on registry divergence · :376 no absolute paths in prompts · :377 no `cd /abs &&` · :378 root-relative write targets · :379 `$CLAUDE_PROJECT_DIR` allowed · :428 CLI launch no-AskUserQuestion · :432 `--spawn` refuses rather than degrades · :672 per-step applicability deference to spec-workflow § Phase Discipline · :683 disposal contract (both PRs merged). **Plan-table discrepancy recorded**: the M0 table in `plan.md` says 21; measured in this tree (file unchanged since base — char count and audit diff both confirm) it is **18**. The enumeration above is the authority; the 21 stands unexplained and unmatchable against the file.
- **session-handoff-examples.md — 12 tokens → 12 clauses**: :160 Block 0 anchoring · :206 Block 0 launchers verbatim · :230 multi-terminal recommendation · :236 V0 lsof+cwd cross-validation · :284 Block 1 line order · :285 purpose-conditional `mode:` · :298 seed-not-permission · :303 fan-out steering phrase · :304 ultracode variants · :306 UUID fallback · :310 arm-only goal · :327 diet constraints.
- **kanban-dispatch-detail.md — 8 tokens → 2 normative clauses + 6 prose mentions**: normative — :114 dispatch language · :130 manager-lead spawned unnamed. Prose mentions (describe the STUB's clauses; not clauses of this file) — :8, :54, :160, :265, :277, :287.

### M1 — doctrine amendment, measured after edit (tree pre-commit; both copies edited identically)

`§ File Size Limits` third sentence replaced: the budget now names **every instruction file the InstructionsLoaded hook measures** (always-loaded and `paths:`-scoped alike); the "Move detailed content to path-scoped rules" bullet carries the same-budget qualifier. AC-IBS-001 predicates, re-run after the edit (verbatim):

```
$ grep -c "also loads in full at every session launch" .claude/rules/moai/development/coding-standards.md
0
$ grep -c "InstructionsLoaded" .claude/rules/moai/development/coding-standards.md
1
$ grep -c "also loads in full at every session launch" internal/template/templates/.claude/rules/moai/development/coding-standards.md
0
```

AC-IBS-001 is **green** (limiting clause 0 in both copies; InstructionsLoaded ≥ 1). 005b invariant, re-measured — the diff output is **byte-identical to the recorded baseline including the hunk header** (the amendment replaced sentences in place without changing line counts, so `141d140` did not move):

```
$ diff .claude/rules/moai/development/coding-standards.md internal/template/templates/.claude/rules/moai/development/coding-standards.md ; echo $?
141d140
< - `git commit --no-verify` — bypasses the relocated pre-commit quality gate (the harness safety net at the commit tier; enforced mechanically by the PreToolUse guard at `internal/hook/pre_tool.go` per SPEC-PRETOOL-GATE-MOVE-001 REQ-PGM-006 / F5)
1
```

One hunk, its single content line byte-equal to the baseline content line, no hunk added or removed — the 005b predicate holds at its strongest form (header unmoved). REQ-IBS-003 satisfied: both copies carry the identical amendment.

### M2 — proper-subset adjudication + companion roster + paths-pinned guard (recorded BEFORE any content moves)

**Adjudication, pattern-by-pattern with the complement named** (globs verbatim from the M0 measurement):

| File | Pattern set P | Companion carries C | Complement P∖C | Verdict |
|---|---|---|---|---|
| `worktree-integration.md` | `**/.claude/agents/**`, `**/.claude/worktrees/**`, `**/.claude/teams/**` | {`**/.claude/worktrees/**`} | {`**/.claude/agents/**`, `**/.claude/teams/**`} — non-empty | **splittable** |
| `kanban-dispatch-detail.md` | `**/kanban-dispatch*.md`, `**/.claude/agents/moai/manager-lead.md`, `**/.claude/skills/moai/workflows/gtd.md` | {`**/kanban-dispatch*.md`} if rung 4 fires | {`manager-lead.md`, `gtd.md`} — non-empty | splittable, **naming-constrained** (first pattern is filename-shaped; companion name MUST NOT match `kanban-dispatch*` — REQ-IBS-013) |
| `spec-workflow.md` | `**/.moai/specs/**`, `**/.moai/config/sections/quality.yaml` | — | — | **compression-first by lead direction** (797-char overage; a split is disproportionate) |
| `session-handoff-examples.md` | `**/session-handoff.md` | — | only non-empty subset is P itself = same-glob shard | **foreclosed** (single-element arithmetic; ladder rungs 1-3 only) |

**Companion roster** (AC-IBS-002-R; the REQ-IBS-013 name check is done here, before the file exists):

1. `worktree-integration.md` → companion **`worktree-integration-ops.md`**, `paths: "**/.claude/worktrees/**"`. Name check: the parent's patterns are directory globs, none filename-shaped — no naming trap; the name matches no parent pattern. Planned move (sizes from the M0/measured spans, verified at M4): § Disposing a Worktree the Automatic Sweep Does Not Reach (2,897 chars, **0 [HARD] clauses**) + § Refused Commands in a Worktree-Isolated Session (19,194 chars, **0 [HARD] clauses**) = 22,091 out; parent projected 39,344. Affinity claim per moved section (REQ-IBS-014/015): both are operational knowledge for a session already working *inside* a worktree — needed on the `worktrees/**` trigger; not needed when only agent definitions (`agents/**`) or team state (`teams/**`) are touched. One cross-block reference rewires on move: the disposal section's "the unpushed-branch rule above" points at the [HARD] clause staying in the parent's Terminology Glossary → becomes a named cross-file pointer. Inbound references measured: **zero** files outside `worktree-integration.md` name either moved section.
2. `kanban-dispatch-detail.md` → **no companion pre-committed**; the ladder runs first (rung 3, dedup against co-loading `kanban-dispatch-mechanics.md`, is the promising instrument at a 1,034-char overage). If rung 4 fires later, the companion is named outside `kanban-dispatch*` — the constraint is recorded here so it binds before any name is chosen.
3. `spec-workflow.md`, `session-handoff-examples.md` → no companions (compression-first / foreclosed).

**`TestWorkflowRulePathsPinned` landed** (AC-IBS-003 green path; internal/template/workflow_rule_paths_pinned_test.go) — asserts the four `paths:` globs against values recorded in the test source (not tree-derived), reusing `parseFrontmatterAndBody`, reading the embedded template tree. RED was the guard's absence (`grep -rl "TestWorkflowRulePathsPinned" internal/template/` → empty, recorded in acceptance.md's RED-now cell); executed now with the anchored selector:

```
$ go test -v -run '^TestWorkflowRulePathsPinned$' ./internal/template/
--- PASS: TestWorkflowRulePathsPinned (0.00s)
ok  github.com/modu-ai/moai-adk/internal/template 0.333s
```

**Neutrality sentinel demonstration** — executed at v0.8.0 with its full table in `plan.md` §B.4 (pre-check exit 0 → injected sentinel reported `TEMPLATE_NEUTRALITY_VIOLATION: class=C1-macos-bias-path` exit 1 → reverted, exit 0, tree clean). Confirmed as already executed; not re-run this round.

### M3 — spec-workflow.md compression (measured after edit)

Eight in-body compression edits, all dedup/tightening — no heading touched, no `[HARD]` line touched, frontmatter untouched: the Route A/B trigger vocabulary stated twice (§ SPEC Phase Discipline intro + § Phase Transitions intro) stated once; the per-transition Route A/B trigger pairs merged into single (A)/(B) bullets (Plan→Run, Run→Sync, Sync-close); the Agent Teams genealogy paragraph compressed with a pointer to `orchestration-mode-selection.md` §C.1; the skip-policy record sentence and the concurrent-pipeline tail tightened; the Gate Entry Condition's two mutually-redundant "every invocation" bullets merged. Measured after edit:

```
$ python3 -c "import io; print(len(io.open('.claude/rules/moai/workflow/spec-workflow.md',encoding='utf-8').read()))"
39984          # was 40,797 — saved 813, under the 40,000 bound
$ grep -c '\[HARD\]' .claude/rules/moai/workflow/spec-workflow.md
4              # unchanged from the M0 baseline — no clause loss
$ cmp -s <local> <template mirror> && echo identical
mirror-identical
```

Anchor check: the externally-referenced headings all survive verbatim — `## SPEC Phase Discipline` (:19), `## Subcommand Classification (Pipeline vs Multi-Agent)` (:75), `## SPEC Complexity Tier (S/M/L)` (:130), `### DDD Mode — ANALYZE-PRESERVE-IMPROVE` (:201), `### TDD Mode — RED-GREEN-REFACTOR (default)` (:209), `## Phase Transitions` (:311). `go test -run '^TestWorkflowRulePathsPinned$' ./internal/template/` → ok (glob unchanged). AC-IBS-002a arm: **green** (40,797 → 39,984 by compression, no split).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
