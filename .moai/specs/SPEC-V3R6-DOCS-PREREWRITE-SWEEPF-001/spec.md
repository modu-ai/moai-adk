---
id: SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001
title: "Docs Pre-Rewrite Cleanup — workflow-commands index sweep: stale design-family slash citations in 3 pages x 4 locales, the sidebar map, and the count statements (card t1727)"
version: "0.1.3"
status: completed
created: 2026-10-19
updated: 2026-10-19
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "docs-site/content/{en,ko,ja,zh}/workflow-commands, docs-site/data/menu"
lifecycle: spec-anchored
tags: "docs, docs-site, i18n, cleanup, slash-citations, workflow-commands, sidebar, v3.2"
tier: M
related_specs: [SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001, SPEC-V3R6-DOCS-PREREWRITE-GATE-001, SPEC-DOCS-SITE-001]
---

# SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001 — workflow-commands index sweep (child 11 of 11)

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-10-19 | 0.1.0 | manager-spec | Initial plan-phase authoring for factory card t1727 (Class C, docs only), the SWEEPF child of the umbrella SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (split ruling `d-20261014T060405Z-0fcd`; routing `d-20261014T114834Z-e039`). It owns the page part of REQ-PRC-001 and AC-PRC-001 for the twelve files named in § Scope, under the parent numbers, and adds REQ-SWF-002 to -007 and AC-SWF-002 to -006 for what the Q13, Q14, and Q17 answers need. Every measurement was taken on tree `74ad197380a0711cdb9816d4e37c94d29af730ea` (branch `WT-workflow-cmd-index`, equal to local `main`); nothing is carried over from the parent ledger. **Operator verdicts applied** (board record `d-20261019T015908Z-220a`, decided by the operator through the question channel on 2026-10-19, scope standing): Q13 = replace with the current equivalent where one exists, delete where none exists; Q17 = all four locales change in the same change. Both rows leave the open set. **Q14 stays open** (the `moai-design` page), and its marker option carries a conditional amendment of REQ-PRC-001 (REQ-SWF-006). **Found while measuring, not in the card text:** (1) `docs-site/data/menu/main.yaml` lines 204 to 208 carries the same `/moai design` name map in four locales plus a `ref:` to the page, and `docs-site/.locale-parity-baseline` line 51 lists the page; both were outside the twelve files (open items O-2, O-3); (2) the Korean `_index.md` carries two design rows and a `/moai goal` summary row that the other locales lack (23 `/moai` table rows against 21), so the locales are not line-for-line twins; (3) no slash command or CLI verb enters the design phase (premise of Q13's "current equivalent", measured in acceptance.md C-7). Status stays `draft`. |
| 2026-10-19 | 0.1.1 | manager-spec | Plan-phase revision applying the leader ruling `d-20261019T020706Z-1c58` (scope `card:t1727`, kind ruling; decided by the leader, session `d1aa9eb6`, under operator delegation; ladder path 6, leader judgment; evidence: the lane-23 plan commit `5d16c25af` and the agent definition of `hns-oss-docs-structure-curator-specialist`). This is a **leader** ruling, not an operator verdict, and it is recorded apart from the operator record `d-20261019T015908Z-220a` (Q13, Q17), which is unchanged. The ruling closes the four open items: **O-1** (a): the lane spawns the structure curator for `_meta.yaml` and `docs-site/data/menu/main.yaml` and the content author / locale translator for page bodies, and manager-develop does not edit curator-owned config (plan.md §A names who edits what); **O-2**: `main.yaml` lines 204 to 208 join the change set (new REQ-SWF-009, AC-SWF-008); **O-3**: `.locale-parity-baseline` line 51 is updated (the ruling's condition, Q14 removes the page, is met by the Q14 verdict below) in the same change (new REQ-SWF-010, AC-SWF-009); **O-4**: the "All subcommands (16)" statements are corrected to the actual row count (new REQ-SWF-008, AC-SWF-007). The change set grows from twelve files to fourteen paths (REQ-SWF-007 and AC-SWF-004 restated for the larger set). **Q14 is then decided by the operator** (board record `d-20261019T021508Z-258e`, decided by the operator through the question channel, 2026-10-19, scope `card:t1727`, kind ruling; cited as relayed by the dispatching coordinator, acceptance.md G-10): delete the `moai-design` page in all four locales, remove the `main.yaml` entry, remove line 51 of `.locale-parity-baseline`, and delete the `_index.md` design rows (the 1c58 items O-2 and O-3 unchanged). The rewrite and marker branches are retired: REQ-SWF-005, REQ-SWF-006, and AC-SWF-006 are kept as numbered stubs marked superseded (no renumbering), REQ-SWF-004, -009, -010 and AC-SWF-005, -008, -009 become unconditional, and REQ-PRC-001 needs no amendment because deletion satisfies its page part. Live requirements 8 (10 numbered), live acceptance criteria 8 (9 numbered). Re-measured on tree `74ad197380` (every RED-now figure of 0.1.0 reproduced unchanged). **Found while measuring:** (a) `docs-site/layouts/_default/baseof.html` lines 58 to 86 builds the previous/next page chain from `main.yaml` with no page-exists guard, so a stale entry for a removed page would link its neighbours to a missing page (read from the template, not rendered); this corrects the 0.1.0 statement that the menu template hides the entry; (b) the only reader of `.locale-parity-baseline` is the oss-docs harness verify recipe, not the CI step "Validate 4-locale parity" (acceptance.md C-15, C-16); (c) two new open items, O-5 (`/moai design` in comments of `.moai/config/sections/design.yaml`) and O-6 (`/moai gtd` and `/moai todo` have no row under "All subcommands"), neither added to this change. Status stays `draft`. |
| 2026-10-19 | 0.1.2 | manager-spec | Repair after plan-audit iteration 1 (FAIL, MP-9; verdict file `.moai/reports/t1727/plan-audit.md` at `14171149a8`). Fixed D1 (the M2 check of plan.md no longer expects the whole-set C-5 to be empty before M3 deletes the five `_index.md` rows; the M2 check is C-5 narrowed to the `_meta.yaml`, page, and `docs-site/data` paths), D2 (the AC-SWF-004 guard now includes `.moai/config` in its no-change pathspec, so an edit to `.moai/config/sections/design.yaml` fails it), D3 (AC-SWF-005 Then 2 names `new_items` as the exempt non-page key), D4 (C-12 flips after M3, not M2). No requirement or criterion added; counts unchanged. Status stays `draft`. |
| 2026-10-19 | 0.1.3 | manager-spec | Repair after delta audit iteration 2 (FAIL 0.90; D7, ac-wording). The narrowed C-5 of the M2 check named the pre-change commit `74ad197380`, which prints five lines however correct M2 is. The M2-closing check now targets `<M2-closing-sha>` and the `74ad197380` form is labelled the pre-deletion control (plan.md §B M2 step 4); a closing-run convention (plan.md §B, acceptance.md §A) states that every ledger cell is run with the closing commit substituted, which fixes the sibling cells C-13 and C-14 and the other closing checks in the same pass; AC-SWF-005's green-path milestone reads M2, M3. No requirement or criterion added. Status stays `draft`. |

**Tier note.** Tier M: ten files edited and four deleted (fourteen paths, the parity baseline included), inside the 5 to 15 band; 10 numbered requirements (8 live) and 9 numbered acceptance criteria (8 live), ceilings 16 and 16. Artifact set: spec.md, plan.md, acceptance.md, plus `decision-index.md` (the decision gate is on in this repository) and `progress.md`. No `design.md` or `research.md`.

## Overview (WHY)

The three `workflow-commands` index pages carry citations of a slash command that does not exist. Measured on tree `74ad197380` (evidence ids refer to acceptance.md §C):

- Twelve content files exist: `_index.md`, `_meta.yaml`, and `moai-design.md` under `docs-site/content/{en,ko,ja,zh}/workflow-commands/` (C-3).
- 27 lines in the twelve files match `/moai design` or `/moai:design` (C-1). The same files hold no other unresolved `/moai <verb>` form: a negative-list search over the 17 registered commands (C-8) returns the same 27 lines and nothing else (C-2). Every other slash citation in the twelve files resolves.
- `/moai design` is not a registered slash command (C-8), and the workflow it names carries `user-invocable: false` (C-9). No file under `.claude/skills/moai/` that routes plan, run, or sync names the design workflow as an entry (C-7), so no current equivalent that a reader can type exists.
- The locales are not twins: `_index.md` has 1 design line in en, ja, and zh and 2 in ko; the page has 5 lines in en and ko and 4 in ja and zh (C-1); the Korean `_index.md` has 23 `/moai` table rows against 21 elsewhere (C-4).
- The same citation sits outside the twelve files in the sidebar data: `docs-site/data/menu/main.yaml` lines 204 to 207 carry `/moai design` in four locales and line 208 the `ref` to the page (C-13). A bounded sweep of `docs-site/` outside `content/` and of the repo-level config finds no other docs surface that cites it; the only other hits are comments in `.moai/config/sections/design.yaml` (O-5) and stylesheet names (C-18).
- The page is linked from the four `_index.md` files, from the four `_meta.yaml` files, and from `main.yaml` (C-5).
- `docs-site/.locale-parity-baseline` line 51 lists the page (C-14). Its reader is the oss-docs harness verify recipe, not the CI step (C-15, C-16).
- The heading "All subcommands (16)" and the sentence below it state 16, which equals the 16 subcommand table rows today (C-11, C-12); deleting the design row makes the number false by one.

The two operator verdicts of record `d-20261019T015908Z-220a` fix what happens to the citations: where a current equivalent exists, the citation is replaced; where none exists, it is deleted; all four locales change together. The leader ruling `d-20261019T020706Z-1c58` fixes who edits what, and brings the sidebar entry, the parity baseline line, and the count statements into this change. The operator verdict `d-20261019T021508Z-258e` fixes the page itself (Q14): it is deleted in all four locales, with its sidebar entry and its baseline line.

## Scope (WHAT)

The files this child changes. Paths are repository-relative; `<L>` is each of `en`, `ko`, `ja`, `zh`. The **change set** is rows 1 to 14: ten files edited, four deleted, fourteen paths in all (Q14 verdict `d-20261019T021508Z-258e`).

| # | Path | Change |
|---|------|--------|
| 1-4 | `docs-site/content/<L>/workflow-commands/_index.md` | delete the design table rows (en, ja, zh: 1 line each; ko: 2 lines) per the Q13 verdict, and correct the "All subcommands" heading and the sentence below it to the actual row count (REQ-SWF-008) |
| 5-8 | `docs-site/content/<L>/workflow-commands/_meta.yaml` | remove the `moai-design` entry (key plus `title: "/moai design"`) |
| 9-12 | `docs-site/content/<L>/workflow-commands/moai-design.md` | delete |
| 13 | `docs-site/data/menu/main.yaml` (lines 203 to 208) | delete the sidebar entry as one block (REQ-SWF-009) |
| 14 | `docs-site/.locale-parity-baseline` (line 51) | delete the line (REQ-SWF-010) |

No other file is changed.

## Definitions

- **Twelve files.** Rows 1 to 12 of § Scope, the content files. Requirements and criteria that say "the twelve files" mean these.
- **Change set.** Rows 1 to 14 of § Scope.
- **Design family.** The text `/moai design` or `/moai:design` followed by a character that is not alphanumeric or a hyphen (the pattern of acceptance.md C-1).
- **Current equivalent.** A slash command registered under `.claude/commands/moai/` or a `moai` CLI verb that does what the citation says it does. For the design family none exists (C-7, C-8, C-9), so the Q13 verdict resolves to deletion for these citations. The determination is measured, not assumed, and is re-run at M1.
- **Subcommand row.** A table row whose first cell links a `/moai` command, counted in the "All subcommands" section of one `_index.md` (the counting method is fixed in AC-SWF-007).
- **Pin.** Unless a cell states another commit, every tree-attributed observation is on `74ad197380a0711cdb9816d4e37c94d29af730ea`.

## Requirements (GEARS)

REQ-PRC-001 keeps the parent's number (page part). The other requirements are this child's own.

### Citations resolve (parent REQ-PRC-001, page part)

- **REQ-PRC-001** (Ubiquitous): Each of the twelve files **shall** cite only `/moai` slash commands whose verb resolves to a file under `.claude/commands/moai/`. Deleting the `moai-design` pages satisfies the page part; no amendment applies.

### Verdict-driven handling (Q13, Q17 — operator verdicts of record `d-20261019T015908Z-220a`)

- **REQ-SWF-002** (Ubiquitous): **Where** a stale citation in the twelve files has a current equivalent, the citation **shall** be replaced by that equivalent; **where** none exists, the line holding the citation **shall** be deleted. For the design family no equivalent exists (C-7), so each table row that cites it **shall** be deleted and **shall not** be replaced by a prose sentence that names the same command.
- **REQ-SWF-003** (Ubiquitous): The edits of REQ-PRC-001 and REQ-SWF-002 **shall** be made in all four locales (ko, en, ja, zh) in the same change, and after the change the four locales **shall** carry the same set of deletions: the per-locale count of `/moai` table rows in `_index.md` **shall** fall by the number of design rows that locale held.

### Page handling (Q14 — operator verdict of record `d-20261019T021508Z-258e`: remove)

- **REQ-SWF-004** (Ubiquitous): The four `moai-design.md` page files **shall** be deleted, the `moai-design` entry **shall** be removed from each `_meta.yaml`, and no file under `docs-site/content/` or `docs-site/data/` **shall** keep a link to the removed page.
- **REQ-SWF-005** [SUPERSEDED by the Q14 verdict `d-20261019T021508Z-258e` — the rewrite branch is not taken; no obligation; number kept so cross-references do not move].
- **REQ-SWF-006** [SUPERSEDED by the Q14 verdict `d-20261019T021508Z-258e` — the marker branch is not taken; no amendment of REQ-PRC-001; number kept].

### Leader-ruling items (leader ruling `d-20261019T020706Z-1c58`, O-2, O-3, O-4)

- **REQ-SWF-008** (Ubiquitous; ruling O-4, a factual correction under the Q13 verdict): In each of the four `_index.md` files, the number stated in the "All subcommands" heading and the number stated in the sentence below it **shall** equal the count of subcommand rows in that section, counted after the design row is deleted by the method of AC-SWF-007, and the four locales **shall** state the same number.
- **REQ-SWF-009** (Ubiquitous; ruling O-2, Q14 verdict): `docs-site/data/menu/main.yaml` **shall** carry no design-family form and no reference to the removed page after the change: the sidebar entry (lines 203 to 208: the four-locale name map and its `ref`) **shall** be deleted as one block, leaving neither a dangling name map nor a dangling `ref`.
- **REQ-SWF-010** (Ubiquitous; ruling O-3, Q14 verdict): Line 51 of `docs-site/.locale-parity-baseline` (`workflow-commands/moai-design.md`) **shall** be deleted in the same change and the change **shall** add no line to that file.

### Preservation

- **REQ-SWF-007** (Unwanted): The change **shall not** modify any file other than the change set of § Scope, and in particular **shall not** modify the other pages of `workflow-commands/` (`moai-goal`, `moai-harness`, `moai-plan`, `moai-project`, `moai-run`, `moai-sync`) in any locale, `docs-site/vercel.json` or `docs-site/layouts/`, and **shall** modify `docs-site/.locale-parity-baseline` only as REQ-SWF-010 states.

## Acceptance

The Given-When-Then criteria AC-PRC-001 and AC-SWF-002 to -009, their RED-now and green-path cells, and the evidence ledger are in `acceptance.md`.

## Out of Scope

Each item is named so that a later change cannot claim it silently. Scope of the other ten children lives in their own SPECs.

### Out of Scope — files outside the change set

- The other six pages of `workflow-commands/` in any locale, and every other page of `docs-site/content/`.
- `docs-site/layouts/` (the sidebar, previous/next chain, and landing templates that read `main.yaml`), `docs-site/static/` (the `moai-design.css` stylesheet name is not the page, C-18), and `docs-site/vercel.json` (its `/design` redirects belong to the REDIR child, Q15).
- `.moai/config/sections/design.yaml`, whose comment lines 2 and 36 name `/moai design`. It is a repo-level config file outside `docs-site/`; whether to change it is open item O-5.
- Adding `/moai gtd` and `/moai todo` rows to the "All subcommands" tables or rewording "All" (open item O-6). This change states the row count that exists; it does not add rows.

### Out of Scope — other stale citations and flags

- Stale citations outside the change set (SWEEPA to SWEEPE, README), and flag citations (`--factory-run`, `--clear-policy`; FLAG child, Q18).

### Out of Scope — checker, gates, and code

- The cited-verbs checker and the Hugo and link-check jobs (GATE child, SPEC-V3R6-DOCS-PREREWRITE-GATE-001, card t1717). This child does not depend on them and does not build them; they will measure these files later.
- Any change under `internal/`, `cmd/`, or `.claude/` (including the design workflow skill and its `user-invocable` flag, and the oss-docs harness verify recipe that reads the parity baseline).
- Spawning the harness specialists: the SPEC names who edits what; the lane does the spawning (plan.md §A).
- Pushing, opening a pull request, or any remote action.

## Dependencies and Parallel Work

- No `depends_on`. This child is implementable without the GATE checker, without the parent's M6, and without any sibling. The GATE checker (card t1717) and the parent's closing run M6 will later measure these files; they read this child's committed criteria by reference (parent plan §B.3) and cannot block it.
- **Non-overlapping siblings.** README (t1720), SWEEPA (t1721), SWEEPB (t1722), SWEEPC (t1723), SWEEPD (t1724), SWEEPE (t1725), FLAG (t1726), RESID (t1719), REDIR (t1718), and GATE (t1717) list no file under `docs-site/content/*/workflow-commands/` in the parent's child table. This was read from the parent table; the siblings' own file lists are not in this tree and were not measured. **`docs-site/data/menu/main.yaml` and `docs-site/.locale-parity-baseline` were not checked against the siblings' file lists** (acceptance.md G-8): a sibling that also edits either file would need a serial order, which the leader sets.
- **Q24 marker.** Not used by this child: the marker branch is retired by the Q14 verdict, so no exemption token is needed from the GATE checker for these files.
- **Single writer per file.** The four `_meta.yaml` files and `main.yaml` belong to the structure curator; page bodies to the content author (ko) and the locale translator (en, ja, zh); see plan.md §A.
