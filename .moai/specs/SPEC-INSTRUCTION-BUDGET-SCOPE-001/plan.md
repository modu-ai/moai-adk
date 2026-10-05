---
id: SPEC-INSTRUCTION-BUDGET-SCOPE-001
title: "Implementation plan — instruction-budget scope alignment and four-file reduction"
version: "0.10.0"
created: 2026-09-28
---

# Plan — SPEC-INSTRUCTION-BUDGET-SCOPE-001

Milestones are ordered by decision-reversibility: the doctrine wording and the proper-subset adjudication come first because they are the decisions most likely to change under review; the mechanical reductions come last.

## A. Context

**Branch base: `088594d6b`** — worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1180`, branch `WT-rules-40k-split`, card `t1180`.

[HARD] **Every measurement in these artifacts is pinned to the SHA, never to the branch name.** `088594d6b` was `develop`'s tip when this worktree was created; it is an **ancestor** of `develop` now, not equal to it. A sentence of the form "measured against develop at `088594d6b`" reads as false to anyone who resolves `develop` afterwards, so the noun is the SHA and the branch name appears only where a moving ref is genuinely what is meant. §G states the hazard in general form; this paragraph is its application.

The pins survive `develop` advancing, and the reason is measured rather than assumed: `git diff --stat 088594d6b develop --` over the four target files, both `coding-standards.md` copies, `internal/hook/instructions_loaded.go`, and `internal/template/templates/.claude/rules/` came back **empty at every `develop` tip read during this card**, each re-measured rather than carried forward. So `088594d6b` remains a valid measurement *of these files* while being stale as a *name for develop*.

**Integration-time obligation.** An empty diff read now is a fact about now, not a guarantee. At the integration window the lane absorbs local `develop` and **re-measures in the merged tree** — all four char counts, the four `paths:` globs, the family census, and the three reference-count baselines. `AC-IBS-003`'s recorded glob values and `AC-IBS-002`'s RED-now figures are the specific things that would go stale, and neither announces it. The absorption is owed under `CLAUDE.local.md` §4.1 whatever the diff says; the empty diff predicts only that it will be uneventful.

No `develop` tip SHA and no commit count appear in this section. Either would be a moving reference wearing a fixed-looking value — see §G, which this card earned the hard way.

Scope is `(c) + (a)` — the doctrine fix plus the four-file reduction. Option `(b)` was rejected before planning (spec.md §5).

## B. Known issues going in

1. **The admissibility test is proper-subset over `paths:` pattern sets, not glob disjointness.** A companion scoped to a proper subset necessarily overlaps its parent on the patterns it carries; that overlap is the saving mechanism, not a violation (spec.md §1 carries the per-trigger arithmetic). The prohibited form is the degenerate case where the companion's set EQUALS the parent's — every trigger co-loads, the saving is zero everywhere, and each shard merely slips under the threshold individually. Test: companion patterns ⊊ parent patterns, complement non-empty. The discriminant is the pattern count, not the char count and not the filename:
   - `worktree-integration.md` — 3 patterns (`**/.claude/agents/**`, `**/.claude/worktrees/**`, `**/.claude/teams/**`). Splittable. A 21,435-char reduction is not plausibly reachable by dedup alone, so this is where the largest split belongs.
   - `kanban-dispatch-detail.md` — 3 patterns (`**/kanban-dispatch*.md`, `**/.claude/agents/moai/manager-lead.md`, `**/.claude/skills/moai/workflows/gtd.md`). **Splittable** — the complement is non-empty. Subject to the naming constraint in §B.2; the companion carries exactly one of the two non-self-matching patterns.
   - `session-handoff-examples.md` — 1 pattern (`**/session-handoff.md`). Genuinely foreclosed: the only proper subset of a one-element set is empty, so any companion co-loads on the only trigger there is and saves nothing. REQ-IBS-007 binds it — removal or compression only, worked through the M6 ladder.
   - `spec-workflow.md` — 2 patterns, so nominally splittable, but compression-first **by lead direction** (spec.md §4 constraint, not a requirement) because the overage is 797 characters. Do not open a split there.
2. **The `kanban-dispatch-detail.md` naming trap is already realised — by a third family member neither the SPEC nor the lead saw until the audit.** The family has **three** members at base, not two (spec.md §1 carries the census and the measurement): `kanban-dispatch.md` (34,901 chars, no `paths:` at all — always-loaded), `kanban-dispatch-detail.md` (41,034), and `kanban-dispatch-mechanics.md` (5,224). All three match `**/kanban-dispatch*.md`, and so does a fourth the earlier prefix census missed — `cross-session-messaging-detail.md` (19,006), which HOLDS the pattern without sharing the prefix. The trigger therefore loads **100,165 characters** across four files, of which **65,264** are reducible (the always-loaded stub's 34,901 are a per-session cost no split or dedup touches — spec.md §1 carries both totals and why they are never summed). Mechanics is a **sibling** of `-detail.md`, not its child — both are companions of the stub — so REQ-IBS-006's proper-subset test does not apply to it and no violation is recorded (spec.md §6 carries the correction to the audit's D3). What it does show, measurably, is the naming trap REQ-IBS-013 forbids: a companion named inside the parent's own filename pattern co-loads with it forever. A new companion named outside that pattern (for example `card-class-detail.md`) and scoped to one of the other two patterns engages no prohibition.
3. **Proper-subset is necessary but not sufficient — trigger affinity is the second condition, and it rides on the same pattern set.** A split that passes the proper-subset test is still wrong if the content moved does not belong to the trigger the companion claims. Move content that a `manager-lead.md` editor needs into a companion scoped to `**/kanban-dispatch*.md`, and that session silently loses guidance it used to have: the characters improve, the reader is worse off, and nothing signals it — the complement triggers are exactly the ones that lose. REQ-IBS-014 and REQ-IBS-015 bind the per-section affinity claim, and AC-IBS-002-A-falsifier is the only condition that can **falsify** one rather than check it for internal consistency. The hazard is largest in M4, where 21,435 characters must be partitioned across three patterns and a size-driven cut is the likeliest place for this defect to land.
4. **The neutrality guard needs no registration — measured by sentinel, not read.** Two registries need an entry, not three. The structural read said so; the sentinel demonstration now shows it.

   **Sentinel demonstration** (plan phase, branch base `088594d6b`). A `/Users/` string — class `C1-macos-bias-path`, whose allow-list is **empty**, so any hit is a binary FAIL — was appended to `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`, the mirror of a target file, then reverted.

   | step | command | exit | observed |
   |---|---|---:|---|
   | pre-check | `grep -c "/Users/" <mirror>` | — | `0` — target clean, so the green below is not pre-existing |
   | pre-check | `go test -run '^TestTemplateNeutralityAudit$' ./internal/template/...` | `0` | `ok github.com/modu-ai/moai-adk/internal/template 0.571s` — **no `[no tests to run]`**, so the swept set was non-empty |
   | sentinel | same command, sentinel present | **`1`** | `TEMPLATE_NEUTRALITY_VIOLATION: class=C1-macos-bias-path file=.claude/rules/moai/workflow/spec-workflow.md` |
   | revert | `git checkout -- <mirror>`; `git status --short` | — | `0` hits; empty status |
   | post-revert | same command | `0` | `ok … 0.446s` |

   The violation names the file **by its path under `.claude/rules/moai/workflow/`**, which is the premise: the walk reaches that directory in the template tree, so a new mirrored `.md` there is covered without enrolment. `AC-IBS-002-G`'s "two registries, not three" is therefore an executed measurement, and REQ-IBS-011 requires the right two entries rather than resting on a read.

   Two notes on method. The selector is **anchored**, so the pass and the failure are both attributable to the one test named. And the sibling packages print `[no tests to run]` — the empty-sweep token — which is why the `internal/template` line is read specifically rather than the aggregate exit code: an aggregate `0` with every package empty would look identical to a real pass.

   **Lease note.** This is the anchored single-test form, not the package suite. `.claude/rules/local/gitflow-lane-protocol.md` §8's lease discipline targets full-suite runs competing across lanes; one test is not that. The unanchored package suite stays un-run and remains a declared Gap (`progress.md` §E.1).

## C. Pre-flight

- `git rev-parse --short HEAD` and `git branch --show-current` re-read immediately before any commit.
- Confirm local/template byte-identity for the four files before editing, so a post-change mirror diff is attributable.
- Capture the pre-change `[HARD]` inventory per file (the M0 baseline below) before the first edit; it is the left column of every transfer table.

## D. Constraints

- Character counts only via the single-invocation form — `python3 -c "import io;print(len(io.open('<path>',encoding='utf-8').read()))"`. No pipe, no redirect: `.claude/rules/moai/development/verification-completeness.md` §2.1 places both outside the RED-now command form, and the previous `... < <path>` shape was a redirect. `wc -c` is bytes and inadmissible.
- Verification scope includes `./internal/template/...` in full.
- Commit subjects name `t1180`.
- No edit to `internal/hook/instructions_loaded.go`, and none to `CLAUDE.local.md`.

## E. Self-verification

Each milestone closes only when its `acceptance.md` AC rows are green and the evidence is written into `progress.md` §E.2.

## F. Milestones

### M0 — baseline capture (no content change)

Record, per file: char count, `[HARD]` clause inventory with line numbers, `paths:` glob, and the inbound-reference count. Baseline counts measured in this tree at `088594d6b`:

| File | chars | `[HARD]` occurrences | inbound refs (files naming it, across `.claude/`, `.moai/docs/`, template mirror) |
|---|---:|---:|---:|
| `spec-workflow.md` | 40,797 | 4 | 78 |
| `worktree-integration.md` | 61,435 | 21 | 38 |
| `session-handoff-examples.md` | 41,615 | 12 | 8 |
| `kanban-dispatch-detail.md` | 41,034 | 8 | 7 |

The `[HARD]` occurrence counts are `grep -c` hits, which is a token count and NOT a clause count — a line may carry the token twice, and a clause may span lines. M0 must resolve each hit into a named clause before it can serve as the transfer-table left column. Treating the `grep -c` number as the clause count is the count-only claim REQ-IBS-008 forbids.

M0 also records both family censuses — enumerated by `grep -rl '<pattern>' .claude/rules/` rather than by filename prefix, with the co-loading and reducible totals stated separately (spec.md §1) — because a baseline taken on `kanban-dispatch-detail.md` alone describes 41% of what that trigger loads, and a prefix-based census misses a co-loader in each family.

### M1 — doctrine amendment (REQ-IBS-001, 002, 003)

The highest-reversibility decision: the exact wording of `coding-standards.md:42`, and whether `:45` ("Move detailed content to path-scoped rules …") needs a qualifier. After M1 that sentence is no longer a full escape from the budget, so leaving it unqualified would make the section self-contradictory (REQ-IBS-002).

Both copies, byte-identical:
- `.claude/rules/moai/development/coding-standards.md`
- `internal/template/templates/.claude/rules/moai/development/coding-standards.md`

Wording change only — the section is not restructured.

### M2 — proper-subset adjudication (REQ-IBS-005, 006, 007, 013)

Before any content moves, decide per file whether a proper subset with a non-empty complement exists, and record the enumerated pattern-by-pattern comparison naming that complement. Expected verdicts from §B, to be demonstrated rather than assumed so a reviewer can disagree before content moves:

| File | expected verdict | basis |
|---|---|---|
| `worktree-integration.md` | splittable | 3 patterns |
| `kanban-dispatch-detail.md` | splittable, with the REQ-IBS-013 naming constraint — the companion's filename MUST NOT match `kanban-dispatch*` | 3 patterns; first pattern is filename-shaped |
| `spec-workflow.md` | compression-first by lead direction | 2 patterns, 797-char overage |
| `session-handoff-examples.md` | foreclosed — compression only | 1 pattern |

M2 also records, per splittable file, the candidate companion filename and the single parent pattern it will carry, so REQ-IBS-013 is checked before the file exists rather than after. The record is the **companion roster** AC-IBS-002-R decides against; it is written to `progress.md` §E.2 before any content moves, which is what makes a new file attributable to a parent at all.

For the `kanban-dispatch*` family the census covers **all four co-loaders** (§B.2), not `-detail.md` alone — the trigger loads 100,165 characters across four files, and a subset judgement made against one of them describes the wrong thing. The census is taken by pattern holding (`grep -rl`), which is what surfaced the fourth.

M2 also lands the two mechanical artifacts the criteria depend on:

- **`TestWorkflowRulePathsPinned`** (AC-IBS-003) — asserts the four `paths:` globs against values recorded in the test source, reusing `parseFrontmatterAndBody` and following the precedent of `skill_authoring_paths_glob_test.go`. Recorded values, not `git show 088594d6b:<path>`: REQ-IBS-010 preserves the globs *unless this SPEC is amended*, a this-card constraint, so a tree-pinned guard would outlive its requirement and become an unattributable stale guard after close.
- **The neutrality sentinel demonstration** — inject a sentinel into a template workflow `.md`, run `go test -run '^TestTemplateNeutralityAudit$' ./internal/template/...`, observe it reported, revert, confirm clean with `git status --short`. This converts AC-IBS-002-G's "two registries, not three" from a structural read of `findNeutralityRoot` into an executed measurement. The selector is **anchored**, and the anchoring is load-bearing rather than cosmetic: without the `^`/`$` pair the pattern also selects longer test names, which makes the pass unattributable to the test actually named — the same empty-or-wrong-swept-set defect as a zero-hit grep read as a pass. This repo's `VacuousTestAssertion` lint enforces it.

> Writing the unanchored form here as a counter-example is itself flagged: `VacuousTestAssertion` scans for the pattern and cannot tell a prescription from an illustration of what not to do. Observed on this file at v0.5.0 — one warning, on prose that prescribed the anchored form correctly. Hence the hazard is described rather than quoted. The general shape is worth carrying: a text-scanning guard cannot read intent, so doctrine that quotes the bad form trips its own guard, and the repair is to name the defect instead of exhibiting it.

### M3 — `spec-workflow.md` → under 40,000 (REQ-IBS-004, 008, 009, 010)

First in firing order (216 firings, 79% of the total). 797 chars to remove by dedup/compression. Highest inbound-reference count of the four (78 files), so M3 carries the largest anchor-resolution burden.

### M4 — `worktree-integration.md` → under 40,000 (REQ-IBS-004, 006, 008, 009, 010, 011, 014, 015)

21,435 chars, partitioned across three patterns. The largest split in the SPEC and the highest-risk milestone for the affinity defect in §B.3: every moved section carries a stated affinity claim, and a section whose affinity spans patterns stays in the parent. If a companion is created it is registered in both registries:
- `internal/template/sanitized_pair_parity_test.go:65` — `sanitizedPairPaths`
- `internal/template/rule_template_mirror_test.go:42` — `workflowOptMirroredPaths`

(`template_neutrality_audit_test.go` needs no entry — walk-based, per §B.4.)

### M5 — `kanban-dispatch-detail.md` → under 40,000 (REQ-IBS-004, 006, 008, 009, 010, 011, 013, 014, 015)

1,034 chars. Splittable per §B.1, under the §B.2 naming constraint. Apply the reduction ladder below in order; a split is the last rung, not the first, and at 1,034 characters an earlier rung will probably suffice. Whichever rung lands, a split path registers the companion in both registries (see M4) and carries the affinity claims.

**Rung 3 is unusually promising here**: `kanban-dispatch-mechanics.md` (5,224 chars) co-loads with this file on every `kanban-dispatch*` trigger, so duplication between the two costs characters twice on the same trigger and removing it is a pure win — no relocation, no new file, no affinity question. Measure the overlap before reaching for a split.

### M6 — `session-handoff-examples.md` → under 40,000 (REQ-IBS-004, 007, 008, 009, 010)

1,615 chars. Foreclosed from *splitting* (§B.1) — but "foreclosed from splitting" is not "compression or blocker", and treating it that way is what would manufacture a false blocker. The file is already a relocation destination, so compression alone may not reach 1,615 characters without touching a `[HARD]` clause. Work the ladder below; a blocker is correct only after rung 2 is measured and found unavailable.

**Rung 2 precondition — measure whether the hook reaches `.moai/docs/**` at all.** The rung rests on a premise nobody has verified: that a destination outside `.claude/rules/` is not in the InstructionsLoaded hook's measured set. Before relocating anything, establish it — the hook measures `input.FilePath` for whatever the runtime declares as a loaded instruction file, and the audit log in the primary checkout records 2,155 distinct measured files, so the check is whether any `.moai/docs/` path appears among them. If it does, rung 2 is closed for this file and the ladder moves to rung 4. **Recording the measurement is part of the rung**: an unmeasured premise presented as a reduction path is the same defect class as an unmeasured blocker.

### The reduction ladder (M5 and M6)

Applies in order. A blocker is real only at the bottom.

| Rung | Instrument | Notes |
|---|---|---|
| 1 | **Compression / duplicate removal in place** | Cheapest, no structural change, no affinity question. Always attempted first. |
| 2 | **Relocation outside `.claude/rules/`** — e.g. `.moai/docs/<topic>.md` with a prose pointer, the pattern `CLAUDE.local.md` § References already uses | Not a companion, so REQ-IBS-005/006's subset test does not apply and no `[HARD]` clause is cut. **Gated on the precondition above.** |
| 3 | **Dedup within the co-loading family** | Wherever a family co-loads on one trigger — **both** the `kanban-dispatch*` case (M5) **and** the `session-handoff*` case (M6). Removes characters from the trigger total without moving anything. Family membership is decided by `grep -rl '<pattern>' .claude/rules/`, never by filename prefix (spec.md §1). |
| 4 | **Split into a companion** | Requires a non-empty complement (so: M4 and M5 only, never M6) plus the naming constraint and the affinity claims. |
| 5 | **Blocker report to the lead** | Correct only once rungs 1-4 are each attempted or measured unavailable, with the measurement recorded. A blocker raised before that is a report of unexplored options, not a blocker. |

**Rung 3 was previously closed to "kanban only", and that was wrong in the most costly direction.** It shut the cheapest instrument on `session-handoff-examples.md` — the file this SPEC itself names as the top false-blocker risk, and the one where foreclosure from splitting (single-pattern `paths:`) leaves compression as the only other route. Measured, the family has three reducible co-loaders besides the stub: `session-handoff-examples.md` 41,615 + `session-handoff-format.md` 5,266 + `context-window-management-detail.md` 11,742 = **58,623** characters on one trigger, of which two files were invisible to the earlier prefix-based census. Duplication between co-loaders costs characters twice on the same trigger, so removing it is a pure win — no relocation, no new file, no affinity question, and no `[HARD]` clause at risk.

**The foreclosure verdict itself is untouched.** It rests on single-element-set arithmetic — the only proper subset of a one-pattern set is empty — and does not depend on how many siblings the family has. Reopening rung 3 gives M6 a second instrument; it does not reopen the split question, and REQ-IBS-007 still binds.

`[HARD]` clause trimming appears nowhere on the ladder and is not a rung (§G). AC-IBS-002-H failing together with the arm's char count is the signal that a blocker is genuine.

### M7 — closure sweep (REQ-IBS-003, 009, 010)

Re-measure all four (and any companion) in one batch; run the anchor resolution check across the whole reference surface; confirm local/template byte-identity; run `go test ./internal/template/... ./internal/hook/...`.

## G. Anti-patterns

- **Same-glob shard (the degenerate case).** The prohibited condition is precise: the companion's `paths:` pattern set **equals** the parent's, so the complement is empty, every trigger co-loads both files, and the saving is zero on every trigger while each shard individually slips under 40,000. Covered by REQ-IBS-005 + AC-IBS-002-P. The tempting form is real: both `kanban-dispatch-detail.md` and `session-handoff-examples.md` are *already* companions, so "add one more companion" reads as the established pattern. It is legitimate exactly when the complement is non-empty — which holds for `kanban-dispatch-detail.md` (3 patterns) and cannot hold for `session-handoff-examples.md` (1 pattern).
- **Naming-trap shard.** A companion named `kanban-dispatch-<anything>.md` re-enters the parent's own first pattern and is a same-glob shard wearing a narrower-looking `paths:` line. REQ-IBS-013 forbids it.
- **Size-driven partition.** Cutting a file at whatever boundary yields the needed characters, then writing whichever glob makes the cut legal. This inverts the correct order: affinity decides the boundary, and the glob follows the affinity (REQ-IBS-014/015). Its signature is a companion whose sections have no common trigger.
- **Affinity asserted rather than stated per section.** "This companion holds the worktree-specific material" is a summary, not an affinity claim. The claim is per moved section and names both the pattern it is needed on and the patterns it is not.
- **Naming a moving reference in prose.** [HARD] **A moving reference written into prose is a claim with an expiry date, and the expiry is invisible in the text.** `develop`, `main`, `HEAD`, `origin/<branch>`, "current", "latest" — each resolves differently tomorrow, and nothing in the sentence signals that it has stopped being true. A fixed object (a commit SHA) cannot rot; a branch name asserts a relationship that the next push dissolves silently.
  - The rule: in prose, **name the SHA and state the relation** (`this base is an ancestor of develop`), never the identity (`develop equals this base`). Where a moving ref is genuinely the subject — a provenance claim about the mainline, an obligation to absorb whatever `develop` holds at integration time — the moving ref is correct and the SHA would weaken it. The discriminator is whether a flip would be a true signal about the subject or noise from upstream (`verification-completeness.md` §4).
  - **Also do not state a commit count or a moving ref's current SHA.** A count of commits ahead is a measurement, and a measurement in prose is a moving reference wearing a fixed-looking value; naming the tip SHA a branch "currently" points at has the same defect one step removed.
  - **Carve-out: a dated record.** A `HISTORY` row or a dated progress entry MAY state a measurement, because the date is in the row — the expiry is visible, which is exactly the property the hazard is about. Live prose carries no date, so it carries no expiry signal.
  - **The carve-out's boundary: the date licenses a measurement, never a conclusion.** `Measured 14 commits ahead on 2026-09-28` belongs in a dated row. `develop is 14 commits ahead` does not — not even in a dated row, because a reader scanning the document reads it as present tense and the date does not travel with the sentence. The observation is what the date makes safe; the assertion built on it is not, and this is where the carve-out will be misused if the distinction is left implicit.
  - **The carve-out's second limb: a quoted command transcript.** A figure inside a verbatim command transcript MAY state a measurement, because the transcript carries its own provenance — the command and its output are both visible in the quote, so the reader can re-run the command rather than trust the sentence. The measurement is attributed to the command and to the run that produced it, never asserted in prose; the same boundary as the dated record applies at one remove. What the transcript may not do is launder a conclusion: a figure quoted with its command is a measurement, and any inference built on it still needs its own dated row.
  - Provenance, because it is the strongest argument available: this hazard is named here because **a v0.6.0 sentence violated it** — `spec.md` asserted that `develop` *equals* the branch base, derived soundly from "the branch has zero commits, so HEAD = base" and then unsoundly from "base = develop", which had been true only at worktree creation and had expired three times by the time it was written. It was written in the same document that warned a measurement must name its tree. Applying a rule one level down while breaking it one level up is the recurring shape; this is its third instance in this card.
- **Reading three green anchor sub-conditions as three independent confirmations.** Named because the composition is counter-intuitive and would otherwise be invisible to a reader of the result: **anchor verification is thinnest exactly where the work is heaviest.**
  - `worktree-integration.md` carries the largest relocation (21,435 chars) and `kanban-dispatch-detail.md` the naming-constrained one, so both have the highest chance of introducing a broken reference.
  - `AC-IBS-004a` is **inert on both** — 0 `#anchor` references measured, nothing there to break. Its green is an empty sweep.
  - So resolution for the two split targets rests on `004b` and `004c` alone; and `004b` is the sub-condition patched mid-plan for two blind spots (reverse order, bound width), which leaves `004c` as the least-patched leg under the heaviest load.
  - `004a` is live only on `spec-workflow.md` (16 files), which is compression-first and gets no split.
  - Consequence for M4 and M5: three green sub-conditions do NOT mean three confirmations. Record `004a` as `N/A (0 references measured)` on those targets so the shortfall is visible in the evidence rather than inferable only by someone who re-reads the baselines.
- **`wc -c` as a budget measurement.** These files are Korean-and-English mixed; bytes exceed chars by 200-800, which is larger than three of the four overages. A `wc -c`-based claim could report a file as over budget when it is under, or the reverse.
- **Counting `[HARD]` instead of transferring clauses.** A matching count before and after is compatible with one clause lost and another duplicated.
- **Trimming a `[HARD]` clause to buy characters.** The budget is satisfied by removing duplication and compressing prose, never by dropping a binding clause. A clause that genuinely belongs in a companion moves; it does not shrink.
- **Editing the local copy and deferring the mirror.** The mirror obligation is per-commit, not per-milestone.

## H. Cross-references

- `spec.md` §1 — the measured premise
- `acceptance.md` — the AC matrix these milestones close
- `.moai/reports/t1180/verdict.md` (primary checkout) — premise evidence, not to be overwritten
