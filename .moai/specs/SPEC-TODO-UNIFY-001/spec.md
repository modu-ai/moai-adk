---
id: SPEC-TODO-UNIFY-001
title: "todo unification (D3) and the done path (D4 subset) — one canonical queue name with a warning alias, a lane's own-card done, removal of the git-flow T18/T19 edges, and a post-pick spec_id setter"
version: "0.1.0"
status: draft
created: 2026-10-18
updated: 2026-10-18
author: manager-spec (card t1641)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/homestate, internal/factory, internal/template, .claude/skills/moai, .claude/commands/moai, docs-site/content, AGENTS.md"
lifecycle: spec-anchored
tags: "todo, gtd, rename, deprecation-alias, lane-done, transition-table, spec-id, template-first, docs-substitution"
tier: L
related_specs:
  - SPEC-TODO-LAND-AUTO-DONE-001
  - SPEC-GTD-CANON-BODY-001
  - SPEC-KANBAN-TODO-CLI-001
  - SPEC-FACTORY-RECORD-001
  - SPEC-LOCAL-MAIN-FLOW-001
  - SPEC-GITHUB-FLOW-DEFAULT-001
  - SPEC-FACTORY-COMPLETION-RECOVERY-001
  - SPEC-FACTORY-RECORD-REPAIR-001
---

# SPEC: todo unification (D3) and the done path (D4 subset)

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-10-18 | Initial plan-phase authoring for card t1641 (run tn1p4e, Class C, operator-approved), grounded in tree `026f48658f` (branch `WT-todo-unify`). Absorbs cards t1602 (finish the gtd to todo rename) and t1590 (the six gtd verbs' guidance-versus-execution mismatch), both dropped into t1641 on 2026-10-10. Nine Open Questions are recorded in `plan.md` §B and `decision-index.md`; every requirement below holds under every answer to them, or names the question it is conditional on. |

## 1. Background and problem

### 1.1 What the card asks for

The card carries two redesign decisions of the 3.2 plan (D3 "todo unification" and the done-path subset of D4) plus one addendum:

- **D3.** `moai todo` becomes the only canonical queue spelling. `moai gtd` survives for one release as a warning alias that delegates to `moai todo`, then is removed. The six verbs that exist only under `moai gtd` are cleaned up. The slash command, the skill, the documents, and the rules carry the canonical name.
- **D4 (subset).** A lane may close its own card. The git-flow edges T18 and T19 of the card transition table are removed. Completion criterion: done is automatic for the most recent ten landed cards.
- **Addendum.** After a card is picked, no verb records its spec_id.

### 1.2 Measured state (tree `026f48658f`; commands and verbatim outputs are in `research.md` §R.2)

1. **Two spellings, one tree.** `moai gtd` is built from the same verb tree as `moai todo` and adds six verbs that `moai todo` does not have: capture, clarify, organize, reflect, engage, answer. The guidance (the queue workflow skill) tells the reader to run `moai todo capture|clarify|organize|reflect|engage|answer`; `moai todo reflect` ends with `Unknown command "reflect" for "moai todo"` (installed build, see `research.md` §R.1 for its provenance). This is the guidance-versus-execution mismatch of t1590.
2. **The canonical name is inverted in part of the shipped surfaces.** `AGENTS.md`, the command help, the router skill, and the header of the queue workflow document already call `moai todo` canonical. The slash command files do not: `/moai:gtd` is described as the canonical entry, `/moai:todo` as its compatibility alias that dispatches `gtd`, the published skills copy that, and a golden test pins the arrangement. The queue workflow document is still named `workflows/gtd.md`.
3. **The spelling is wide.** 371 tracked text files contain "gtd" (385 paths when binary files are counted); 204 of them sit outside the historical SPEC directories; 73 of those carry the command spelling `moai gtd` (255 lines). Some of the 73 are executable consumers (an argument vector in the board mod, a shell script default, an in-tree error message), not prose.
4. **Persisted names must not move.** The gtd tables (`gtd_items`, `gtd_relations`, and four more), the `gtd-<16 hex>` item id, and the `.moai/gtd/answers` directory are persisted formats; 47 files under `internal/` carry GTD identifier-shaped tokens.
5. **A lane cannot close its own card.** The lane guard admits seven read-only verbs; `done` is not among them. The leader-approval receipt gate then binds every close of a factory-linked card.
6. **T18 and T19.** T18 is `merged-local → done`, refused while the repository has a remote; T19 is the reserved `pushed → ci-green` edge. The repository's own configuration is still `mode: manual`, `workflow: git-flow`, and `SPEC-GITHUB-FLOW-DEFAULT-001` is `in-progress`, so the "GitHub Flow transition complete" premise of the decision is not met in this repository today.
7. **No post-pick spec_id path.** The queue spec_id is set only by `moai todo next <n> --spec` at pick; the factory card row's spec_id is set by `moai factory assign --spec` only while the card is `picked`, and a later attempt is refused ("its hints and pointers change only while picked"). A card dispatched without a spec_id wedges kickoff, whose SPEC lookup is `<worktree>/.moai/specs/<spec_id>`.

### 1.3 Why

One canonical name removes a standing documentation contradiction and a standing runtime mismatch (items 1 and 2) and ends the "fully supported alias" posture that keeps every guidance surface bilingual. The done-path changes remove the human act that the 3.2 redesign measured as the dominant wedge in completing cards. The spec_id setter removes a known kickoff wedge that has so far been repaired by direct database edits.

## 2. Scope and tier

**Tier L.** Reasons, each measured: more than 15 files are affected (73 files carry the command spelling, plus the Go packages `internal/cli`, `internal/homestate`, `internal/factory`, `internal/template`); the change is estimated at well over 1,000 lines once tests are counted (the milestone sizes in `plan.md` §E sum to roughly 2,500 illustration-level lines, 255 of them the mechanical substitution); and it is constitutional in the sense of the tier table: it edits the canonical rule and skill surfaces, `AGENTS.md`, the transition table whose size a test pins, and the approval gate that binds every close. Artifact set: spec.md, plan.md, acceptance.md, design.md, research.md (plus progress.md, not counted). Requirement and criterion counts are within the Tier L ceilings (24 and 24 against 25 each).

Three tracks, kept separable so that the answer to any Open Question invalidates at most one milestone:

- **Track A, one canonical name (REQ-TU-001 to REQ-TU-014).**
- **Track B, the done path (REQ-TU-015 to REQ-TU-021).**
- **Track C, the spec_id setter (REQ-TU-022 to REQ-TU-024).**

## 3. Requirements (GEARS)

Open Questions referenced below (OQ-1 to OQ-9) are defined in `plan.md` §B and mirrored one-to-one as rows Q1 to Q9 of `decision-index.md`.

### 3.1 Track A, one canonical name

**REQ-TU-001 (Ubiquitous).** The `moai todo` command tree shall be the single registration of every queue verb, and `moai gtd` shall be built from that tree, so that no verb has a second handler under the `gtd` spelling.

**REQ-TU-002 (Event-driven).** When any `moai gtd` invocation executes (any verb, or the bare parent), the command shall write exactly one warning to stderr that names the `moai todo` spelling of what was run and the release that removes `moai gtd`, and shall then run the `moai todo` handler, leaving stdout and the exit status identical to the `moai todo` form. (Completion criterion "gtd warns, then delegates".)

**REQ-TU-003 (Unwanted).** The `moai gtd` warning shall not appear on stdout and shall not appear for a help request, so that machine-readable output and help text remain byte-identical to the `moai todo` form.

**REQ-TU-004 (Ubiquitous).** Each of the six stage verbs (capture, clarify, organize, reflect, engage, answer) shall carry exactly one recorded disposition, absorbed into `moai todo` or deleted, backed by a usage inventory measured at the start of the run phase, and no verb shall remain registered under `moai gtd` only. (Depends on OQ-1; the requirement holds under every answer.)

**REQ-TU-005 (Ubiquitous).** Every `moai <command> <verb>` spelling that a skill, rule, workflow document, README, docs-site page, help text, or error message instructs the reader to run shall resolve to an executable verb of the command it names. (The t1590 clause: guidance that cannot be executed does not land.)

**REQ-TU-006 (Where).** Where a stage verb is absorbed into `moai todo`, the lane-session guard shall classify it explicitly as read-only (admitted) or mutating (refused), and no absorbed verb shall depend on the guard's default for its classification.

**REQ-TU-007 (Ubiquitous).** The canonical slash entry shall be `/moai:todo`; `/moai:gtd` shall remain for the grace release as a thin wrapper that prints a one-line rename notice and dispatches to the same workflow; the published skills `moai-todo` and `moai-gtd` shall equal the command-to-skill publisher's output; the router skill shall route `todo` as canonical and `gtd` as alias; and the queue workflow document shall live at `workflows/todo.md` in the template source and the deployed tree, with `workflows/gtd.md` absent from both.

**REQ-TU-008 (Ubiquitous).** Every executable consumer of the `gtd` spelling (argument vectors, shell scripts and their defaults, in-tree error and help messages) shall be changed to `todo` by reviewed edits, not by the bulk text substitution.

**REQ-TU-009 (Ubiquitous).** The bulk replacement of the `moai gtd`, `/moai gtd`, and `/moai:gtd` spellings in documentation and template files shall be performed by one repeatable script that touches only the classes the substitution table allows and leaves every excluded class (historical SPEC directories, reports, the changelog, binary files, recorded fixtures, Go identifiers and persisted names) byte-identical.

**REQ-TU-010 (Event-driven).** When the substitution script runs on a tree it has already converted, it shall change no file and report zero replacements.

**REQ-TU-011 (Ubiquitous).** After the change, a `moai gtd` or `/moai gtd` token shall remain in a tracked file outside the excluded classes only on a line the residual ledger allows (the alias and deprecation statements), and a repository test shall enforce that.

**REQ-TU-012 (Ubiquitous).** Internal identifiers and persisted formats shall keep their present names: the Go identifiers and file names, the `gtd_*` tables, the `gtd-<16 hex>` item id, and the `.moai/gtd/answers` directory. (Depends on OQ-3; the requirement is the default reading.)

**REQ-TU-013 (Ubiquitous).** Every changed file that has a template source shall be edited in `internal/template/templates/`, the embedded catalog shall be regenerated by the project build, and the deployed mirror shall equal the regenerated output wherever a mirror guard exists.

**REQ-TU-014 (Ubiquitous).** `AGENTS.md` shall name `moai todo` canonical and `moai gtd` a warning alias, and shall stay within the Codex contract byte ceiling.

### 3.2 Track B, the done path

**REQ-TU-015 (Event-driven).** When a lane session runs `done` for the card its own lane currently holds and that card has landing evidence of the closed form the auto-done scan already accepts, the lane guard shall admit the call.

**REQ-TU-016 (Unwanted).** The lane guard shall not admit `done` for a card the lane does not hold, nor for a held card without landing evidence, nor any other verb outside the read-only allowlist; the own-card `done` path is the only change to what a lane may do on the queue.

**REQ-TU-017 (Event-driven).** When an own-card `done` is admitted, the close shall pass through the same approval-gate path as every other close, under the receipt policy that OQ-5 selects, and shall write a log row whose source names the lane own-card path.

**REQ-TU-018 (Ubiquitous).** The card transition table shall hold no `merged-local → done` row and no reserved `pushed → ci-green` entry, both requests shall be refused as undeclared transitions, and the accepted and refused pair counts that the table test pins shall not change. (Depends on OQ-6.)

**REQ-TU-019 (Where).** Where the `merged-local → done` row is removed, a merged-local card in a repository without a remote shall have the completion path that OQ-6 selects, and the push gate shall request only edges the table declares.

**REQ-TU-020 (Event-driven).** When a card with landing evidence reaches its closing act (a lane's own-card `done` or the leader's `auto-done` scan), every record that OQ-7 names (the queue row, the factory card row, or both) shall end in `done` without any operator act, and the log shall record which path closed it. (Depends on OQ-4, OQ-5, OQ-7.)

**REQ-TU-021 (Ubiquitous).** The criterion "done is automatic for the most recent ten landed cards" shall be evaluated by a repeatable fixture test and, at sync, by a live read-only measurement whose command and output the sync record carries.

### 3.3 Track C, the spec_id setter

**REQ-TU-022 (Event-driven).** When the spec_id setter is invoked on a `picked` or `assigned` card whose spec_id is empty, with a well-formed SPEC identifier, the setter shall write the identifier to the queue row and to the factory card row as one operation and record the change.

**REQ-TU-023 (Unwanted).** If the card already carries a different spec_id, or is in any state other than `picked` or `assigned`, or the identifier is malformed, the setter shall refuse and change nothing; an identical value shall succeed as a no-op.

**REQ-TU-024 (Ubiquitous).** `next --spec`, `factory assign --spec`, and the unpick clear shall keep their present contracts, and a card whose spec_id the setter wrote shall pass kickoff's SPEC lookup exactly as one whose spec_id was set at pick.

## 4. Constraints

- **Methodology.** `quality.yaml` sets `development_mode: tdd`: every milestone starts with a RED step whose failure is observed for the stated reason before the implementation lands.
- **No test skips and no timeout increases** anywhere in the change.
- **Verification scope.** The landing re-measure runs the changed packages and their direct importers; inside `internal/cli` it runs a name-exact `-run` selection of the tests that name the changed files or identifiers; the remainder of `internal/cli` is carried to batch CI as an explicit Gap.
- **Template-First.** Sources live under `internal/template/templates/`; deployed `.claude/` and `.moai/` copies are produced by the build, never edited first.
- **No time estimates.** Priorities and ordering only.

## 5. Out of Scope

### Out of Scope — internal renames and persisted formats

- Renaming Go identifiers, `gtd_*.go` file names, or the `internal/factory` types that carry GTD in their names (default reading of OQ-3).
- Renaming the `gtd_*` tables, the `gtd-<16 hex>` item id, or the `.moai/gtd/answers` directory, and any data migration they would need.

### Out of Scope — other 3.2 stage 2 work

- The landing-to-done machine transition driven by three observers and the automatic issuing of approval receipts as a general mechanism: card t1639 (factory v2, currently on hold) owns it, subject to OQ-4. This SPEC delivers the lane own-card path and the receipt policy seam only.
- The `pushed → done` decider path, the lease and expiry repairs, and the transition-table redesign (cards t1621 and t1639).
- Web pages, worktree relocation, and the CLI split (cards t1609, t1640, and stage 3).

### Out of Scope — history and release mechanics

- Editing the historical SPEC directories, `.moai/reports/**`, or `CHANGELOG.md`.
- Choosing the release that removes `moai gtd` (OQ-2) and performing that removal; this SPEC delivers the warning alias and the single constant that names the removal release.
- Pushing, opening pull requests, or batch landing; the leader owns them.

## 6. References

- Card t1641 text and absorbed cards t1602, t1590 (queue records; quoted in `research.md` §R.3).
- `.moai/reports/moai-32-master-upgrade-plan-20261010.md` (stage 2, item 3.2-2-4) and `.moai/reports/moai-adk-32-cross-harness-redesign-20261010.html` (decisions D3 and D4). Both are untracked report material: they are evidence of the operator's decision, not authority citations.
- `AGENTS.md` §11 (the `moai todo` verb row).
- `.claude/rules/moai/workflow/factory-dispatch.md` (lane spawn authority and the leader's remaining role).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` (frontmatter and artifact statelessness).
