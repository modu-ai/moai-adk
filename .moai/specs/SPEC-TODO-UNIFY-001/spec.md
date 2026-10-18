---
id: SPEC-TODO-UNIFY-001
title: "todo unification (D3) and the done path (D4 subset) — one canonical queue name with a warning alias, a lane's own-card done under an evidence-minted receipt, removal of the git-flow T18/T19 edges, and a post-pick spec_id setter"
version: "0.2.0"
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
| 0.1.0 | 2026-10-18 | Initial plan-phase authoring for card t1641 (run tn1p4e, Class C, operator-approved), grounded in tree `026f48658f` (branch `WT-todo-unify`). Absorbs cards t1602 (finish the gtd to todo rename) and t1590 (the six gtd verbs' guidance-versus-execution mismatch), both dropped into t1641 on 2026-10-10. Nine Open Questions were recorded in `plan.md` §B and `decision-index.md`. |
| 0.2.0 | 2026-10-18 | Revision for plan-audit iteration 1 (FAIL 0.62, MP-7, ten blocking findings; receipt `rcpt-90e345b1d5488c65cc03140e`). All nine questions are now decided by the lane's decision ladder and written into the requirements as concrete conditions (`decision-index.md` Q1 to Q9; Q5 was shaped by the leader's confirmation of 10-18, conditional). Changes: the six stage verbs are absorbed (REQ-TU-004); the warning covers executed invocations only, not argument-refused ones (REQ-TU-002/003); the queue workflow document keeps its path (REQ-TU-007, REQ-TU-012); a lane's own-card done closes both records under a receipt issued automatically only when three machine facts hold, issuer `auto`, revocable (REQ-TU-017, REQ-TU-025); T18 and the T19 reservation are removed and one evidence-guarded edge T29 replaces T18's role (REQ-TU-018/019); both records end `done` with a stated order and recovery (REQ-TU-020); the spec_id setter states its write order and recovery (REQ-TU-022); the milestone dependency graph replaces the independence claim (`plan.md` §A). Requirements 25, criteria 25 (Tier L ceilings 25 / 25). |

## 1. Background and problem

### 1.1 What the card asks for

The card carries two redesign decisions of the 3.2 plan (D3 "todo unification" and the done-path subset of D4) plus one addendum:

- **D3.** `moai todo` becomes the only canonical queue spelling. `moai gtd` survives for one release as a warning alias that delegates to `moai todo`, then is removed. The six verbs that exist only under `moai gtd` are cleaned up. The slash command, the skill, the documents, and the rules carry the canonical name.
- **D4 (subset).** A lane may close its own card. The git-flow edges T18 and T19 of the card transition table are removed. Completion criterion: done is automatic for the most recent ten landed cards.
- **Addendum.** After a card is picked, no verb records its spec_id.

### 1.2 Measured state (tree `026f48658f`; commands and verbatim outputs are in `research.md` §R.2)

1. **Two spellings, one tree.** `moai gtd` is built from the same verb tree as `moai todo` and adds six verbs that `moai todo` does not have: capture, clarify, organize, reflect, engage, answer. The guidance (the queue workflow skill) tells the reader to run `moai todo capture|clarify|organize|reflect|engage|answer`; `moai todo reflect` ends with `Unknown command "reflect" for "moai todo"` (installed build, see `research.md` §R.1 for its provenance). This is the guidance-versus-execution mismatch of t1590.
2. **The canonical name is inverted in part of the shipped surfaces.** `AGENTS.md`, the command help, the router skill, and the header of the queue workflow document already call `moai todo` canonical. The slash command files do not: `/moai:gtd` is described as the canonical entry, `/moai:todo` as its compatibility alias that dispatches `gtd`, the published skills copy that, and a golden test pins the arrangement. The queue workflow document is named `workflows/gtd.md`; that path is an identifier wired into five rule files' `paths:` keys, a generated embed line, a recorded fixture, and seven test files (`research.md` §R.10), so it keeps its path.
3. **The spelling is wide.** 371 tracked text files contain "gtd" (385 paths when binary files are counted); 204 of them sit outside the historical SPEC directories; 73 of those carry the command spelling `moai gtd` (255 lines). Some of the 73 are executable consumers (an argument vector in the board mod, a shell script default, an in-tree error message), not prose.
4. **Persisted names must not move.** The gtd tables (`gtd_items`, `gtd_relations`, and four more), the `gtd-<16 hex>` item id, and the `.moai/gtd/answers` directory are persisted formats; 47 files under `internal/` carry GTD identifier-shaped tokens.
5. **A lane cannot close its own card.** The lane guard admits seven read-only verbs; `done` is not among them. The leader-approval receipt gate then binds every close of a factory-linked card, and only the leader mints a receipt, so a landed card whose lane is gone waits for a leader act.
6. **T18 and T19.** T18 is `merged-local → done`, refused while the repository has a remote, human-decided, receipt-gated; T19 is the reserved `pushed → ci-green` edge. The repository's own configuration is still `mode: manual`, `workflow: git-flow`, and `SPEC-GITHUB-FLOW-DEFAULT-001` is `in-progress`, so the "GitHub Flow transition complete" premise of the decision is not met here today. The operator's instruction of 10-17 releases the item for now, so this SPEC removes the edges and declares the replacement.
7. **No post-pick spec_id path.** The queue spec_id is set only by `moai todo next <n> --spec` at pick; the factory card row's spec_id is set by `moai factory assign --spec` only while the card is `picked`, and a later attempt is refused ("its hints and pointers change only while picked"). A card dispatched without a spec_id wedges kickoff, whose SPEC lookup is `<worktree>/.moai/specs/<spec_id>`.

### 1.3 Why

One canonical name removes a standing documentation contradiction and a standing runtime mismatch (items 1 and 2) and ends the "fully supported alias" posture that keeps every guidance surface bilingual. The done-path changes remove the human act that the 3.2 redesign measured as the dominant wedge in completing cards. The spec_id setter removes a known kickoff wedge that has so far been repaired by direct database edits.

## 2. Scope and tier

**Tier L.** Reasons, each measured: more than 15 files are affected (73 files carry the command spelling, plus the Go packages `internal/cli`, `internal/homestate`, `internal/factory`, `internal/template`); the change is estimated at well over 1,000 lines once tests are counted; and it is constitutional in the sense of the tier table: it edits the canonical rule and skill surfaces, `AGENTS.md`, the transition table whose size a test pins, and the approval gate that binds every close. Artifact set: spec.md, plan.md, acceptance.md, design.md, research.md (plus decision-index.md and progress.md, not counted). Requirement and criterion counts are within the Tier L ceilings (25 and 25 against 25 each).

Three tracks, kept separable by the dependency graph of `plan.md` §A:

- **Track A, one canonical name (REQ-TU-001 to REQ-TU-014).**
- **Track B, the done path (REQ-TU-015 to REQ-TU-021, REQ-TU-025).**
- **Track C, the spec_id setter (REQ-TU-022 to REQ-TU-024).**

## 3. Requirements (GEARS)

Questions Q1 to Q9 are decided in `decision-index.md`; each answer is written into the requirement it affects. A requirement names observable behavior; identifiers and file names live in `design.md`.

### 3.1 Track A, one canonical name

**REQ-TU-001 (Ubiquitous).** The `moai todo` command tree shall be the single registration of every queue verb, and `moai gtd` shall be built from that tree, so that no verb has a second handler under the `gtd` spelling.

**REQ-TU-002 (Event-driven).** When a `moai gtd` invocation passes argument and flag validation and begins to execute (any verb, or the bare parent), the command shall write exactly one warning to stderr whose first line begins `Warning: moai gtd has been renamed to moai todo`, that names the `moai todo` spelling of what was run and states that `moai gtd` will be removed one release after the release that first carries the warning, and shall then run the `moai todo` handler, leaving stdout and the exit status identical to the `moai todo` form. (Completion criterion "gtd warns, then delegates".)

**REQ-TU-003 (Unwanted).** The `moai gtd` warning shall not appear on stdout, shall not appear for a help request, and shall not appear for an invocation refused by argument or flag validation (an unknown verb, a missing argument, an unknown flag), so that machine-readable output, help text, and those refusals stay byte-identical to the `moai todo` form's apart from the command name. (Those refusals return before the command's pre-run step: `research.md` §R.11.)

**REQ-TU-004 (Ubiquitous).** Each of the six stage verbs (capture, clarify, organize, reflect, engage, answer) shall be registered on the `moai todo` command tree with its present arguments, flags, and behavior, so that `moai todo <verb>` and `moai gtd <verb>` run the same handler; no stage verb shall remain registered under `moai gtd` only, and none shall be deleted. (Decision Q1.)

**REQ-TU-005 (Ubiquitous).** Every `moai <command> <verb>` spelling that a skill, rule, workflow document, README, docs-site page, help text, or error message instructs the reader to run shall resolve to an executable verb of the command it names. (The t1590 clause: guidance that cannot be executed does not land.)

**REQ-TU-006 (Where).** Where a stage verb is registered on `moai todo`, the lane-session guard shall classify it explicitly: `reflect` without `--rebuild-projection` is read-only and admitted; `reflect --rebuild-projection` (it writes the private projection files), `capture`, `clarify`, `organize`, `engage`, and `answer` are mutating and refused with the standing refusal text; a stage verb without a declared classification shall fail a test rather than inherit a default.

**REQ-TU-007 (Ubiquitous).** The canonical slash entry shall be `/moai:todo` and shall dispatch `todo`; `/moai:gtd` shall remain for the grace release as a thin wrapper that prints a one-line rename notice and dispatches `todo`; the published skills `moai-todo` and `moai-gtd` shall equal the command-to-skill publisher's output; the router skill shall route `todo` as canonical and `gtd` as alias; and the queue workflow document shall name `moai todo` canonical in its header and body while keeping its path.

**REQ-TU-008 (Ubiquitous).** Every executable consumer of the `gtd` spelling (argument vectors, shell scripts and their defaults, in-tree error and help messages) shall be changed to `todo` by reviewed edits, not by the bulk text substitution.

**REQ-TU-009 (Ubiquitous).** The bulk replacement of the `moai gtd`, `/moai gtd`, and `/moai:gtd` spellings in documentation and template files shall be performed by one repeatable script that touches only the classes the substitution table allows and leaves every excluded class (historical SPEC directories, reports, the changelog, binary files, recorded fixtures, Go identifiers and persisted names, and the always-loaded files `AGENTS.md` and its template) byte-identical.

**REQ-TU-010 (Event-driven).** When the substitution script runs on a tree it has already converted, it shall change no file and report zero replacements.

**REQ-TU-011 (Ubiquitous).** After the change, a `moai gtd` or `/moai gtd` token shall remain in a tracked file outside the excluded classes only on a line the residual ledger allows (the alias and deprecation statements), and a repository test shall enforce that.

**REQ-TU-012 (Ubiquitous).** Internal identifiers and persisted formats shall keep their present names and values: the Go identifiers and file names, the `gtd_*` tables, the `gtd-<16 hex>` item id pattern, the `.moai/gtd/answers` directory, and the queue workflow document's path together with every path string that names it. (Decision Q3.)

**REQ-TU-013 (Ubiquitous).** Every changed file that has a template source shall be edited in `internal/template/templates/`, the embedded catalog shall be regenerated by the project build, and the deployed mirror shall equal the regenerated output wherever a mirror guard exists.

**REQ-TU-014 (Ubiquitous).** The `moai todo` row of `AGENTS.md` and of its template mirror shall name `moai todo` canonical and `moai gtd` a warning alias, shall not retain the words "fully supported", and the file shall stay within the Codex contract byte ceiling.

### 3.2 Track B, the done path

**REQ-TU-015 (Event-driven).** When a lane session runs `done` for the card its own lane currently holds and that card has landing evidence of the closed form the auto-done scan already accepts, the lane guard shall admit the call.

**REQ-TU-016 (Unwanted).** The lane guard shall not admit `done` for a card the lane does not hold, nor for a held card without landing evidence, nor any other verb outside the read-only allowlist; the own-card `done` path is the only change to what a lane may do on the queue.

**REQ-TU-017 (Event-driven).** When a close reaches the approval gate for a factory-linked card (the lane's own-card `done` or the leader's auto-done scan), the gate shall verify a receipt exactly as it does for every close (leader role marker, issuer is not the card's performing owner, card identity, run, factory version, evidence hash); and the close path shall issue that receipt automatically, only at that moment and only for that card and run, if and only if all three machine facts hold: (1) the card's factory row is in `merged-local`; (2) the card's `sync-audit.md` verdict file under `.moai/reports/<card-id>/` is committed, carries `verdict: PASS` or `verdict: PASS-WITH-DEBT`, and names an `audited_sha` that is an ancestor of the integration branch (`main` in this repository); (3) the re-measure record keyed to the merge commit's tree satisfies the gate that the merge edge applied. If any one fact is absent, the close path shall issue nothing, the close shall be refused with the gate's receipt-absent refusal, and the card shall wait for the leader's manual approval (fail-closed). The issued receipt shall record the issuer `auto`, distinguishable from a leader-issued receipt, and the close shall write a log row naming the closing path and the receipt issuer. (Decision Q5, leader-confirmed 10-18 conditional.)

**REQ-TU-018 (Ubiquitous).** The card transition table shall hold no edge guarded by the human-decided no-remote completion (formerly row T18) and no reserved-edge refusal for `pushed → ci-green` (formerly the T19 reservation); a request for `pushed → ci-green` shall be refused as an undeclared transition; and the accepted and refused pair counts that the table test pins (71 and 370) shall not change. (Decision Q6.)

**REQ-TU-019 (Event-driven).** When a card in `merged-local` requests `done`, the table shall admit it through exactly one declared edge (T29) if and only if (a) the card's recorded merge commit is reachable from the landing ref of its repository (the resolved landed ref when a remote is configured, the local integration branch when none is) and (b) the approval gate verifies a receipt (REQ-TU-017 or a leader-issued receipt); otherwise it shall refuse the request and change nothing. In a repository without a remote the push gate shall request `done`, and a card with that evidence and receipt shall reach `done` without any push step; the push gate shall request only edges the table declares.

**REQ-TU-020 (Event-driven).** When a card with landing evidence of the closed form reaches its closing act (the lane's own-card `done` or the leader's auto-done scan), the closing act shall end the factory card row and the queue row in `done` without an operator act, writing the factory edge (REQ-TU-019) first and the queue archive second, and the log shall record which path closed it. If the queue archive fails after the factory edge committed, the closing act shall exit non-zero naming both records and their states, and a repeat of the same act shall complete the queue archive without requesting the factory edge again. A factory row in a state other than `merged-local` shall be left untouched and named in the log; its queue row closes only where the gate verifies a leader-issued receipt, because the automatic receipt requires `merged-local`. (Decisions Q4 and Q7.)

**REQ-TU-021 (Ubiquitous).** The criterion "done is automatic for the most recent ten landed cards" shall be evaluated by a repeatable fixture test and, at sync, by a live read-only measurement whose command and output the sync record carries.

**REQ-TU-025 (Event-driven).** When the leader runs the revocation act on a card's automatically issued receipt, the receipt store shall record the revocation, a close that depends on that receipt shall be refused as receipt-absent, and no automatic issuance shall occur for that card and run again while the revocation stands; a lane session shall be refused the revocation act, and a leader-issued receipt shall be unaffected by it.

### 3.3 Track C, the spec_id setter

**REQ-TU-022 (Event-driven).** When the spec_id setter is invoked by the leader on a `picked` or `assigned` card with a well-formed SPEC identifier, the setter shall write the identifier to each carrier whose spec_id is empty (the factory card row, with a field-change event, and the queue row) under the queue lock, and shall never leave a carrier holding a value other than the requested identifier; if a failure leaves one carrier written and the other empty, the setter shall exit non-zero naming both carriers, and a repeat invocation with the same identifier shall fill the empty carrier and succeed. (Decisions Q8 and Q9.)

**REQ-TU-023 (Event-driven).** When the card already carries a different spec_id in either carrier, or is in a state other than `picked` or `assigned`, or the identifier is malformed, or the caller is a lane session, the setter shall refuse and change neither carrier; an identical value already present in both carriers shall succeed as a no-op.

**REQ-TU-024 (Ubiquitous).** `next --spec`, `factory assign --spec`, and the unpick clear shall keep their present contracts, and a card whose spec_id the setter wrote shall pass kickoff's SPEC lookup exactly as one whose spec_id was set at pick.

## 4. Constraints

- **Methodology.** `quality.yaml` sets `development_mode: tdd`: every milestone starts with a RED step whose failure is observed for the stated reason before the implementation lands.
- **No test skips and no timeout increases** anywhere in the change.
- **Verification scope.** The landing re-measure runs the changed packages and their direct importers; inside `internal/cli` it runs a name-exact `-run` selection of the tests that name the changed files or identifiers; the remainder of `internal/cli` is carried to batch CI as an explicit Gap.
- **Template-First.** Sources live under `internal/template/templates/`; deployed `.claude/` and `.moai/` copies that have a template source follow it and never lead it.
- **No time estimates.** Priorities and ordering only.

## 5. Out of Scope

### Out of Scope — internal renames and persisted formats

- Renaming Go identifiers, `gtd_*.go` file names, or the `internal/factory` types that carry GTD in their names (decision Q3).
- Renaming the `gtd_*` tables, the `gtd-<16 hex>` item id, or the `.moai/gtd/answers` directory, and any data migration they would need.
- Renaming the queue workflow document `workflows/gtd.md` or retiring the `moai-gtd` docs-site pages; both are paths other surfaces key on, and a rename follows the removal release as its own change.

### Out of Scope — other 3.2 stage 2 work

- The landing-to-done machine transition driven by three observers (a SessionEnd observer, a `factory complete` observer) and the general mechanism that issues approval receipts at observation time: card t1639 (factory v2, currently on hold) owns it. This SPEC issues a receipt only inside the two close paths of REQ-TU-017 and delivers no observer.
- The `pushed → done` edge and its decider path (a card whose factory row the leader already moved to `pushed`), the lease and expiry repairs, and the transition-table redesign (cards t1621 and t1639).
- A lane-side surface for recording a spec_id (no sibling SPEC carries it: `research.md` §R.8).
- Web pages, worktree relocation, and the CLI split (cards t1609, t1640, and stage 3).

### Out of Scope — history and release mechanics

- Editing the historical SPEC directories, `.moai/reports/**`, or `CHANGELOG.md`.
- Choosing the release that removes `moai gtd` and performing that removal; this SPEC delivers the warning alias and the single constant that states the removal as "one release after the release that first carries the warning".
- Pushing, opening pull requests, or batch landing; the leader owns them.

## 6. References

- Card t1641 text and absorbed cards t1602, t1590 (queue records; quoted in `research.md` §R.3).
- `.moai/reports/moai-32-master-upgrade-plan-20261010.md` (stage 2, item 3.2-2-4) and `.moai/reports/moai-adk-32-cross-harness-redesign-20261010.html` (decisions D3 and D4). Both are untracked report material: they are evidence of the operator's decision, not authority citations.
- `AGENTS.md` §11 (the `moai todo` verb row).
- `.claude/rules/moai/workflow/factory-dispatch.md` (lane spawn authority and the leader's remaining role).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` (frontmatter and artifact statelessness).
