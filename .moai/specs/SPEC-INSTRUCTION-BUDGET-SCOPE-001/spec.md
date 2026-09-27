---
id: SPEC-INSTRUCTION-BUDGET-SCOPE-001
title: "Align the 40,000-char instruction budget doctrine with the hook, and bring the four over-budget workflow rules under it"
version: "0.8.0"
status: draft
created: 2026-09-28
updated: 2026-09-28
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: ".claude/rules/moai/development/coding-standards.md, .claude/rules/moai/workflow/"
lifecycle: spec-anchored
tier: M
tags: "instruction-budget, rules-diet, doctrine-code-parity, template-mirror"
---

## HISTORY

- 2026-09-28 — v0.1.0 — plan-phase artifacts authored from card t1180's measured premise (`.moai/reports/t1180/verdict.md` in the primary checkout). Scope decision `(c) + (a)` was made by the lead before authoring; option `(b)` (narrowing the hook to session-start loads) was rejected and is recorded in §5.
- 2026-09-28 — v0.2.0 — lead correction. v0.1.0 judged `kanban-dispatch-detail.md` structurally foreclosed from splitting, reasoning only from its first `paths:` pattern (`**/kanban-dispatch*.md`, which self-matches). The file carries three patterns and is splittable on the same grounds as `worktree-integration.md`; the self-match is a naming trap, not a structural foreclosure. Only `session-handoff-examples.md` (one pattern) is genuinely foreclosed. Added REQ-IBS-016 (companion-naming constraint) and REQ-IBS-017/018 (per-section trigger-affinity justification — trigger disjointness alone permits a partition that silently strips a reader of guidance), plus AC-IBS-007b and AC-IBS-024/025/026. The §1 `patterns` column and the saving arithmetic were added so the discriminant is visible rather than inferred. (The AC ids cited in this row are pre-renumbering — see v0.3.0.)
- 2026-09-28 — v0.3.0 — Tier M budget repair, found while re-counting after the v0.2.0 additions. `spec-workflow.md` § SPEC Complexity Tier caps Tier M at 16 requirements AND 16 acceptance criteria independently; v0.2.0 stood at 18 REQ and 26 logical AC, over both. Consolidated to 15 REQ and 16 AC **without dropping a check**: three requirements were retired to §4 as constraints (the reduction order, `spec-workflow.md` compression-first, and the `wc -c` prohibition — none of the three asserts anything about the closing state, so each was a constraint miscast as a requirement), and related criteria merged into multi-condition rows. `acceptance.md` was renumbered `AC-IBS-001`…`AC-IBS-016`; ids in earlier HISTORY rows refer to the pre-renumbering scheme. Retired requirement ids (005, 006, 014) are left as gaps and not reused.
- 2026-09-28 — v0.4.0 — lead correction, second round: the **admissibility test itself** was wrong, and the error originated in the spawn prompt rather than in the reading of it. That prompt required a companion glob "strictly narrower than — and genuinely disjoint in trigger from — its parent's", which cannot both hold: a strict subset of the parent's patterns necessarily overlaps the parent. v0.1.0 resolved the contradiction toward "disjoint", which forecloses every companion of every file; v0.2.0/v0.3.0 corrected the `kanban-dispatch-detail.md` verdict but left the contradictory wording standing in REQ-IBS-007/008/009 and in AC-IBS-004 (which read "strict subset … and shares no pattern with it" — self-contradictory on its face). The operative test is now stated once and consistently: **companion patterns ⊊ parent patterns, complement non-empty**, with characters-loaded-per-trigger as the metric. The prohibited form is the degenerate equal-set case. §1 carries the per-trigger arithmetic table; AC-IBS-004 checks both failure directions and records explicitly that an overlap check must NOT be added to it.

- 2026-09-28 — v0.5.0 — plan-audit iteration 1 returned **FAIL** (0.825 arithmetic / 0.79 harmonic vs the Tier M 0.80 threshold; the MUST-PASS firewall decided it, not the score). Report: `.moai/reports/t1180/plan-audit.md` in this worktree. Seven blocking defects repaired: `tier: M` added (the machine had been reading Tier **L**, so the entire v0.3.0 "Tier M budget repair" was performed toward a tier the frontmatter never declared); the `kanban-dispatch*` census corrected from two members to **three** with the family's per-trigger total measured at 81,159 characters; REQs renumbered continuously `001`-`015` (the v0.3.0 gaps were an asymmetry, since AC had been renumbered continuously in the same pass); `acceptance.md` rebuilt with RED-now cells, a three-class criterion taxonomy, and six rows folded into the budget row; the `ls`-based decider replaced (in this environment `ls` is aliased to long format, so it returned `0`, neither the expected 2 nor the correct 3); the anchor-resolution baselines split by command with the reverse-order `§` form added; and a falsifier added to the affinity condition. Two auditor findings were **corrected rather than applied** — see §6. Note that the REQ ids cited in the v0.2.0 and v0.4.0 rows above are pre-renumbering and no longer resolve against §2: old `016`/`017`/`018` are now `013`/`014`/`015`, and old `007`/`008`/`009` are now `005`/`006`/`007`. Earlier rows are left as written — a HISTORY row records what was done at the time, and rewriting its identifiers would make the record agree with the present at the cost of no longer describing the past.

- 2026-09-28 — v0.6.0 — three lead addenda to the audit repair, each measured rather than accepted. **Unit discipline**: every reference-count figure now states its unit (files), because the audit's comparable figures are *line* counts and mixing the two is the same defect class as mixing commands. Measured consequence: the reverse-order `§` form adds **zero** new files (`comm -23` of reverse against forward = 0 for all four), and widening the bound 40 → 200 leaves the file counts identical — so both audit blind spots live *inside* the forward file set as per-reference gaps, not as missing files. **Per-target zero rule**: `AC-IBS-004a` records `N/A (0 references measured)` rather than PASS wherever its sweep is empty, which is both files this SPEC splits. **Pin discipline**: the document pin is relabelled the *branch base* SHA and never "develop" — develop moved three times during this card (`e9577def` → `088594d6b` → `7e05ef43b` → `37dc766b9`, last measured 14 commits ahead) while the target files stayed unchanged at every tip, so the SHA is a valid measurement and the branch name is an expired one. Also: `AC-IBS-004c` tightened from a filename `grep -c` to a two-condition per-section line check; a named anti-pattern added for reading three green anchor sub-conditions as three confirmations; REQ-IBS-006 explicitly barred from the always-loaded-stub shape where `⊊` is undefined; and a dangling `§B.5` → `§B.4` repaired in the plan that defines the cross-reference check.

- 2026-09-28 — v0.7.0 — a **false factual claim** removed, found by the lead. v0.6.0 asserted that `develop` **equals** this base (`spec.md` §6, `progress.md` header), derived soundly from "the branch carries zero commits, so `HEAD` = base" and then unsoundly from "base = `develop`" — an identity that held only at worktree creation and had expired several times before it was written. Measured: the base is an **ancestor** of `develop`, which is ahead. Two consequences beyond the wording: the claim that the recommended `git merge develop` absorption is a **no-op is withdrawn** — absorption plus re-measurement in the merged tree is owed unconditionally under `CLAUDE.local.md` §4.1, and the empty target-file diff predicts only that it will be uneventful; and two further sentences asserting the branch "carries zero commits" (`acceptance.md` §Document-level measurement pin and `AC-IBS-007`) had been falsified by the plan-artifact commit itself, so both now state why they hold rather than resting on an empty branch. Commit counts are removed from all prose. The hazard is named in `plan.md` §G: **a moving reference in prose is a claim with an expiry date, and the expiry is invisible in the text** — recorded with its provenance, since the offending sentence appeared in the same document that warns a measurement must name its tree.

- 2026-09-28 — v0.8.0 — the **sentinel demonstration executed**, closing the only declared gap that bore on an obligation this SPEC *imposes* rather than one it measures. A `C1-macos-bias-path` sentinel (allow-list empty, so any hit is a binary FAIL) injected into `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` was reported by the anchored audit as `TEMPLATE_NEUTRALITY_VIOLATION: class=C1-macos-bias-path file=.claude/rules/moai/workflow/spec-workflow.md`, exit `1`; pre-check exit `0` on a clean target, exit `0` again after revert, tree clean. So `AC-IBS-002-G`'s "two registries, not three" and REQ-IBS-011's registration requirement rest on an executed measurement rather than a structural read of the walk. Full table in `plan.md` §B.4. Also in this version: the AC count's **unit** is stated wherever a count appears, after the repo's `ac-baseline-guard` hook reported `COUNT 10` against the artifacts' `8` — both correct (logical criteria vs live identifiers including a letter suffix), with `8` binding the Tier ceiling and `10` what the sync-phase CHANGELOG convention will count. Fourth unit mismatch in this card.

## 1. Context

A hook handler enforces a 40,000-character budget on every loaded instruction file. Verified in this tree at base `088594d6b`:

- `internal/hook/instructions_loaded.go:103` — `const charBudget = 40000`
- `internal/hook/instructions_loaded.go:67-73` — the budget is checked against `instructionPath` (= `input.FilePath`), i.e. ANY loaded instruction file, and an overage returns `&HookOutput{SystemMessage: err.Error()}`, which reaches the session as a user-visible message.
- `internal/hook/instructions_loaded.go:104-106` — message text: `"%s exceeds 40,000 char budget at %d; split content per coding-standards.md"`
- `internal/hook/instructions_loaded.go:62` — `input.LoadReason` is recorded into the audit row but is never used as a gate. Consequence: `paths:`-scoped rule files are measured on every glob-triggered load.
- `internal/hook/instructions_loaded.go:78-83` — a separate CWD `CLAUDE.md` fallback check logs via `slog.Warn` only and is non-blocking; it reaches no reader. The user-visible path is `:70-72`.

The doctrine states something narrower. `.claude/rules/moai/development/coding-standards.md:42` (§ File Size Limits) reads: "Any project-local instruction file that **also loads in full at every session launch** follows the same size discipline." The code honours no such limitation. That doc/code inconsistency is the first defect this SPEC closes.

The second defect is that four `paths:`-scoped workflow rule files are over budget and therefore emit that message on every matching load. Char counts measured in this tree at `088594d6b` with `python3 -c 'print(len(open(p).read()))'` — **characters, not bytes**; these files are Korean-and-English mixed, so `wc -c` overstates by 200-800:

| File (under `.claude/rules/moai/workflow/`) | chars | over | observed firings | `load_reason` | `paths:` glob | patterns |
|---|---:|---:|---:|---|---|---:|
| `spec-workflow.md` | 40,797 | +797 | 216 | `path_glob_match` | `**/.moai/specs/**,**/.moai/config/sections/quality.yaml` | 2 |
| `worktree-integration.md` | 61,435 | +21,435 | 7 | `path_glob_match` | `**/.claude/agents/**,**/.claude/worktrees/**,**/.claude/teams/**` | 3 |
| `session-handoff-examples.md` | 41,615 | +1,615 | 1 | `path_glob_match` | `**/session-handoff.md` | 1 |
| `kanban-dispatch-detail.md` | 41,034 | +1,034 | 1 | `path_glob_match` | `**/kanban-dispatch*.md,**/.claude/agents/moai/manager-lead.md,**/.claude/skills/moai/workflows/gtd.md` | 3 |

The `patterns` column is load-bearing for §2, because the admissibility test is a **proper-subset** test over pattern sets and the metric is **characters loaded per trigger** — not glob disjointness. A companion scoped to a proper subset of its parent's patterns necessarily overlaps the parent on the patterns it does carry; that overlap is the mechanism, not a defect. For a parent with patterns {A,B,C} at size S, moving X characters into a companion scoped to {A} alone:

| trigger | loaded before | loaded after | effect |
|---|---:|---:|---|
| A | S | (S−X) + X = S | unchanged; both files now individually < 40,000 |
| B | S | S−X | saves X |
| C | S | S−X | saves X |

Real reduction on every trigger the companion does not claim, no regression on the one it does. The prohibited same-glob shard is the **degenerate case** where the companion's pattern set equals the parent's: then every trigger co-loads, the saving is zero on all of them, and the only achievement is that each shard individually slips under the threshold — the metric-gaming form. The test is therefore: companion patterns ⊊ parent patterns, with a non-empty complement.

A parent with exactly one pattern cannot satisfy it — the only proper subset of a one-element set is empty — so `session-handoff-examples.md`, the sole single-pattern file of the four, is foreclosed from splitting by arithmetic rather than by judgement.

### The `kanban-dispatch*` family has three members, and the per-trigger total is not one file

The arithmetic above models a parent as a single file. For the `**/kanban-dispatch*.md` trigger that model is wrong at base, and the correction is measured rather than argued:

```
$ git ls-tree --name-only 088594d6b:.claude/rules/moai/workflow/ | grep '^kanban-dispatch'
kanban-dispatch-detail.md
kanban-dispatch-mechanics.md
kanban-dispatch.md
```

| file | chars | `paths:` |
|---|---:|---|
| `kanban-dispatch.md` | 34,901 | **none — no frontmatter at all**; the file declares itself "Intentionally always-loaded" |
| `kanban-dispatch-detail.md` | 41,034 | `**/kanban-dispatch*.md,**/.claude/agents/moai/manager-lead.md,**/.claude/skills/moai/workflows/gtd.md` |
| `kanban-dispatch-mechanics.md` | 5,224 | `**/kanban-dispatch*.md,**/.claude/agents/moai/manager-lead.md,**/.moai/state/integration/**` |

All three match `**/kanban-dispatch*.md`, so touching anything named `kanban-dispatch*` loads **81,159 characters** across three files, not 41,034 across one. The proper-subset test's *direction* is unaffected — the per-trigger table above still holds for each file individually — but its absolute figures describe one file at a time, and this family's trigger total is the sum. M2's census therefore enumerates the whole family rather than a single parent.

**`kanban-dispatch-mechanics.md` is a sibling of `-detail.md`, not its child.** Its own header names `kanban-dispatch.md` as the SSOT and `-detail.md` as a "sibling companion", and both are companions of the stub. This matters for what REQ-IBS-006 does and does not say: the proper-subset test compares a companion's pattern set against **its parent's**, and this family's parent (`kanban-dispatch.md`) has no pattern set at all. The test is therefore inapplicable to this family as written, and mechanics does not violate REQ-IBS-006 — recording a violation that cannot be substantiated would be worse than recording none.

What mechanics does realise is the REQ-IBS-013 naming trap, measurably: its filename matches the `kanban-dispatch*` pattern that `-detail.md` carries, which is why all three co-load. That is the concrete precedent for the naming constraint rather than a hypothetical one, and it is the reason a new companion of `-detail.md` must be named outside that pattern.

**Retroactive repair of `kanban-dispatch-mechanics.md` is out of scope** — see §5.

**A shape this SPEC does not model: always-loaded stub → several `paths:`-scoped companions.**

[HARD] **REQ-IBS-006 does NOT govern this shape, and MUST NOT be extended to cover it.** The test is `companion patterns ⊊ parent patterns`. Where the parent carries no `paths:` field at all — as `kanban-dispatch.md` does not — there is no parent set, so `⊊` is **undefined**, not merely hard to evaluate. Stretching the requirement over this shape would re-introduce exactly the incoherence v0.4.0 removed: a test whose two halves cannot both hold. The shape needs a different analysis, and this SPEC does not supply one.

Two consequences a later reader needs, because both are easy to get backwards:

- **`kanban-dispatch-mechanics.md` does not violate REQ-IBS-006.** Its parent has no pattern set to be a subset of. What it violates is nothing in this SPEC; what it *demonstrates* is the REQ-IBS-013 naming hazard.
- **The stub's cost is per-session, not per-trigger.** `kanban-dispatch.md` is always-loaded by declaration, so its 34,901 characters are paid on every session regardless of what is touched. The per-trigger arithmetic above measures what a companion split saves **against a trigger**; no companion split reduces a fixed per-session cost. A saving calculation for this shape that folds the stub into the trigger total would overstate the reduction by 34,901 characters — the whole stub.

No work under this SPEC creates that shape. It is recorded because splitting an always-loaded file is what would create it, and because the shape is present in this very family, one file away from what the SPEC does touch.

A separate hazard rides on the same pattern set and is NOT settled by the proper-subset test: whether the content moved actually belongs to the trigger the companion claims. That is the trigger-affinity condition (REQ-IBS-017/018).

Local copy and template mirror are byte-identical for all four (verified by `diff -q` in this tree).

Firing counts are attributed to `.moai/logs/rule-load-audit.jsonl` **in the primary checkout**, not in this worktree: 8,910 rows, 2,155 distinct files, 272 over-budget firing events, `load_reason` distribution `path_glob_match` 5,120 / `session_start` 2,512 / `nested_traversal` 598 / `include` 552 / `compact` 128. This SPEC restates those figures as carried evidence from `.moai/reports/t1180/verdict.md`; it does not claim to have re-measured the log.

Two findings shape the design and are the reason a char-count-ordered priority would fix the wrong file first:

1. The file that fires most is barely over budget — `spec-workflow.md`, 2% over, 79% of all firings, because its glob `**/.moai/specs/**` matches on every SPEC-touching turn.
2. The file that is massively over budget barely fires — `worktree-integration.md`, 53% over, 7 firings.

Noise and bloat live in different files. Priority therefore follows firing order, not char order.

## 2. Requirements (GEARS)

- REQ-IBS-001 — The `coding-standards.md` § File Size Limits section shall state that the 40,000-character budget applies to every instruction file the InstructionsLoaded hook measures, whether always-loaded or `paths:`-scoped.
- REQ-IBS-002 — The amended § File Size Limits section shall not contradict itself: where it prescribes moving detailed content to `paths:`-scoped rules as a remedy for overage, it shall carry a qualifier stating that a `paths:`-scoped destination is subject to the same budget.
- REQ-IBS-003 — Where a doctrine amendment lands in `.claude/rules/`, the system shall carry the byte-identical amendment in `internal/template/templates/.claude/rules/`.
- REQ-IBS-004 — Each of the four files named in §1 shall measure below 40,000 characters.
- REQ-IBS-005 — The system shall not reduce a file by relocating content into a new companion whose `paths:` pattern set equals its parent's.
- REQ-IBS-006 — Where a new companion file is created, its `paths:` pattern set shall be a proper subset of its parent's — every companion pattern appears in the parent, and at least one parent pattern is absent from the companion — and that relation shall be demonstrated by an enumerated pattern-by-pattern comparison naming the non-empty complement, rather than asserted in prose.
- REQ-IBS-007 — Where a parent's `paths:` pattern set has exactly one element, the only admissible reduction shall be removal or compression. (The arithmetic — that a one-element set has no proper subset with a non-empty complement — is established in §1 and is not restated as a requirement. Of the four files this binds `session-handoff-examples.md` alone.)
- REQ-IBS-008 — Every `[HARD]` clause present in a modified file before the change shall be present afterwards, at a stated destination, evidenced by a per-clause transfer table.
- REQ-IBS-009 — Every cross-reference that points into a modified file shall still resolve after the change.
- REQ-IBS-010 — The `paths:` frontmatter of the four existing files shall be preserved, unless this SPEC is amended to argue a specific change.
- REQ-IBS-011 — Where a new companion file is created, it shall be registered in the sanitized-pair parity registry and the rule-template mirror registry.
- REQ-IBS-012 — The system shall not modify `CLAUDE.local.md` under this SPEC.
- REQ-IBS-013 — Where a companion is created for a parent whose `paths:` contains a filename-shaped pattern, the companion's filename shall not match that pattern.
- REQ-IBS-014 — Where a section is relocated into a companion, the system shall state that section's trigger affinity — which of the parent's patterns the section is needed on, which it is not needed on, and why.
- REQ-IBS-015 — Where a section's trigger affinity spans more than one of the parent's patterns, that section shall remain in the parent and be reduced by compression or not at all.

## 3. Acceptance Criteria

Enumerated in `acceptance.md` (Tier M).

## 4. Constraints

- **Template-First.** `.claude/**` changes carry a mirror obligation into `internal/template/templates/.claude/**`. The template mirror content is guard-tested in `./internal/template/...`, so that package is in the verification scope in full.
- **Measurement unit.** Character counts use `python3 -c 'import sys;print(len(sys.stdin.read()))'` fed from a file read or `git show <ref>:<path>`. `wc -c` is bytes and is not admissible for a budget claim — a character-count claim is never derived from a byte count.
- **Reduction order.** The four files are addressed in firing order: `spec-workflow.md`, then `worktree-integration.md`, then `kanban-dispatch-detail.md`, then `session-handoff-examples.md`. This is milestone sequencing (plan.md §F), so it is a constraint rather than a requirement — the closing state is the same whichever order the work takes, and only the order in which value lands differs.
- **`spec-workflow.md` is compression-first, by lead direction.** The file carries two `paths:` patterns and is nominally splittable, but a 797-character overage is a dedup target and a split there would be disproportionate. Stated as direction, not as a structural finding — unlike `session-handoff-examples.md`, whose single pattern forecloses splitting structurally (REQ-IBS-009).
- **Evidence path convention.** `.moai/reports/t1180/verdict.md` (primary checkout) already exists and carries the premise measurement plus a correction section. It is referenced, never overwritten.
- **Card id.** Every commit on branch `WT-rules-40k-split` names `t1180`.
- **Non-goal: hook behaviour.** No change to `internal/hook/instructions_loaded.go`. The `slog.Warn`-only CLAUDE.md fallback at `:78-83` is noted as an observation, not a work item.

## 5. Exclusions

### Out of Scope — hook narrowing (option b, rejected)

- Gating the budget check on `input.LoadReason == "session_start"`. Rejected by the lead: it changes user-visible warning behaviour, and it would silence the 20 `CLAUDE.local.md` firings arriving via `nested_traversal` (14) and `compact` (6), which are not `session_start`.
- Any change to `charBudget`'s value, the message text, or the non-blocking CLAUDE.md fallback.

### Out of Scope — CLAUDE.local.md

- `CLAUDE.local.md` copies fire 47 over-budget events at 43,986-45,810 chars (`session_start` 27, `nested_traversal` 14, `compact` 6). The file is local-only and non-templated, so it is outside this card. Recorded here as a follow-up card candidate with those figures. The file is not touched.

### Out of Scope — retroactive repair of `kanban-dispatch-mechanics.md`

- The file realises the REQ-IBS-013 naming trap at base: 5,224 characters, `paths:` matching `**/kanban-dispatch*.md`, co-loading with `kanban-dispatch.md` (34,901) and `kanban-dispatch-detail.md` (41,034) for a per-trigger total of 81,159 characters. It is under budget individually, so the hook never warns about it.
- Repairing it is a **follow-up card candidate**, not this card's work: renaming it or rescoping its `paths:` would change what loads on the `manager-lead.md` and `.moai/state/integration/**` triggers, which is a behaviour change to a family this SPEC only measures.
- Recorded so the census is not mistaken for a repair list.

### Out of Scope — the other 2,151 measured files

- The audit log names 2,155 distinct instruction files. Only the four in §1 plus `CLAUDE.local.md` were observed over budget. No preventive reduction of under-budget files is in scope.

### Out of Scope — same-glob sharding as a reduction method

- Splitting a file into two shards that load on the same trigger. Prohibited by REQ-IBS-007; recorded here so it is not reintroduced as an implementation convenience. Total characters loaded per trigger would be unchanged or higher (an extra header and an extra pointer block), with the only effect being that each shard individually slips under the threshold.

## 6. Corrections to the audit

Two findings in `.moai/reports/t1180/plan-audit.md` are recorded here as corrected rather than applied, each with the measurement that corrects it. Both were verified in this worktree at `088594d6b`.

**D3's parentage claim.** The report calls `kanban-dispatch-mechanics.md` a companion of `kanban-dispatch-detail.md` and concludes it violates REQ-IBS-006 (then numbered 008) by carrying a pattern its parent lacks. Mechanics is a **sibling** of `-detail.md`: its header names `kanban-dispatch.md` as the SSOT and `-detail.md` as a sibling companion, and `kanban-dispatch.md` carries no `paths:` frontmatter at all (`sed -n '1,5p'` shows the file opening at its H1). The proper-subset test compares a companion against *its parent's* pattern set, and this family's parent has none — so the test is inapplicable here and the violation is not substantiated. D3's other half stands and is recorded in §1: all three files co-load on `**/kanban-dispatch*.md`, which is the measured realisation of the REQ-IBS-013 naming trap.

**Residual-risk (1) — the divergence direction is inverted.** The report states that `kanban-dispatch-detail.md`'s `paths:` has already diverged in the primary checkout (`gtd.md` → `todo.md`), that `kanban-dispatch-mechanics.md` is absent there, and that `develop` may have reorganised the family — recommending a `git merge develop` absorption before M2. Measured:

```
$ git ls-tree --name-only develop:.claude/rules/moai/workflow/ | grep '^kanban-dispatch'
kanban-dispatch-detail.md
kanban-dispatch-mechanics.md
kanban-dispatch.md
$ git merge-base --is-ancestor main develop && echo "main IS ancestor of develop"
main IS ancestor of develop
$ git rev-list --count --left-right main...develop
0	6546
```

This base is an **ancestor** of `develop`, which is currently ahead of it; and `develop` fully contains `main`, which is further behind and still carries the older two-file, `todo.md` state. The primary checkout is on `main` — the intended steady state for this repository, where local `main` is a synchronisation reference nobody branches from — so reading it produced the older state and the inference ran backwards: `main` is behind, `develop` did not diverge.

**Absorption is still owed, and the finding's failure does not excuse it.** The `CLAUDE.local.md` §4.1 lane duty requires absorbing local `develop` and **re-measuring in the merged tree** before the merge, whatever the diff says — that obligation is unconditional and does not depend on this family having diverged. What the measurement establishes is only the expected blast radius: `git diff --stat 088594d6b develop --` over the four target files, both `coding-standards.md` copies, the hook, and the template rules mirror is **empty** at the `develop` tip read in this run. Uneventful is not unnecessary, and the re-measurement is the part that cannot be skipped.

No commit count and no `develop` tip SHA appear in this prose, deliberately. Either would be wrong by the time it is read, which is the same failure one level down from the one this section corrects — `plan.md` §G carries the general rule and its dated-record carve-out.

**No divergence risk is recorded in this SPEC**, because there is no divergence: one branch is behind the other, which is the normal state here. `kanban-dispatch-mechanics.md` and the `gtd.md` pattern exist on `develop` and have not reached `main` yet; they arrive through the release PR. The family census is correct as measured on this base. What survives from the auditor's finding is the ordinary obligation already stated in plan.md §A — absorb develop at the integration window and re-measure there.

The reusable point, since this is the second time in this card that a reading of the primary checkout was mistaken for a reading of the branch base: in this repository those two trees are far apart by design, so a measurement's tree must be **named**, not assumed. The hazard generalises past tree identity to any moving reference in prose — `plan.md` §G carries it as a named anti-pattern, because the version of this section that stated `develop` *equals* this base was itself an instance of it, written in the same document that warns against it.

## 7. Cross-references

- `internal/hook/instructions_loaded.go` — the enforcing handler (read-only for this SPEC)
- `.claude/rules/moai/development/coding-standards.md` § File Size Limits — the doctrine amended by REQ-IBS-001/002
- `internal/template/sanitized_pair_parity_test.go` — `sanitizedPairPaths` registry (REQ-IBS-013)
- `internal/template/rule_template_mirror_test.go` — `workflowOptMirroredPaths` registry (REQ-IBS-013)
- `.moai/reports/t1180/verdict.md` (primary checkout) — premise measurement
