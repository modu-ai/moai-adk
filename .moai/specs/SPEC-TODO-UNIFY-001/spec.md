---
id: SPEC-TODO-UNIFY-001
title: "todo unification (D3) and the post-pick spec_id setter — one canonical queue name with `moai gtd` removed immediately and every residual reference migrated (done-path subset moved to card t1878)"
version: "0.7.0"
status: draft
created: 2026-10-18
updated: 2026-10-19
author: manager-spec (card t1641)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/homestate, internal/factory, internal/template, .claude/skills/moai, .claude/commands/moai, docs-site/content, AGENTS.md"
lifecycle: spec-anchored
tags: "todo, gtd, rename, immediate-removal, residual-zero, spec-id, template-first, docs-substitution"
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
| 0.3.0 | 2026-10-18 | Revision for plan-audit iteration 2 (FAIL 0.80, seven blocking findings; receipt `rcpt-717ae7d0bd489fdeced99ba8`). The Q5 shape the leader confirmed (three machine facts, issuer `auto`, revocable, fail-closed) is unchanged; what changed is how facts 2 and 3 are bound. REQ-TU-017: every fact is read from git objects the card's factory row names (the recorded merge commit, the recorded evidence commit), never from the card's working tree; the audited commit must cover what lands (after the evidence commit the card adds nothing to the merge except its own report files, while main absorbed after the audit is tolerated because the re-measure of fact 3 covers the integrated tree); the re-measure record must name the merged tree and the merge's first parent in its own fields; the facts, the receipt, and the closing edge share one factory-record transaction (iteration-2 D16, D17). REQ-TU-003: the warning is printed from the command's run step, so required-flag, flag-group, and lane-guard refusals carry none (D19). REQ-TU-022/023: the setter works from `picked` through `kickoff`, which covers the leased wedge of board record `d-20261010T001732Z-d4e8`, and a repeat that completes a half-written pair succeeds in every non-terminal state (D20, D21). REQ-TU-019: T29 states its admitted deciders (D26). The log row's issuer moved into REQ-TU-020, and AC-TU-014 now asserts it and the `pushed` clause (D22). The verb count is 33 (D18). The REQ-FCR-014 reversal is named (D23). Requirements 25, criteria 25. |
| 0.4.0 | 2026-10-18 | Post-ceiling correction under the leader's standing rule C (plan-audit iteration 3 FAIL 0.86, three blocking findings D29 to D31; receipt `rcpt-9557497ba4bcc4b92912703c`). No identifier changes (25 and 25) and the Q5 shape is unchanged. REQ-TU-019: without a remote T29 no longer needs the push gate caller to name an integration branch; with a configured branch the test is exact, without one the merge commit must be contained in a local branch other than the card's own (D29). REQ-TU-017: the issuer `auto` is a value the leader's manual issuance refuses, so an `auto` receipt is always an automatic one (D30); every fact is read raw (untrimmed output, replacement objects disabled) and no fact is read where the repository carries a graft or shallow file (D31). Out of Scope names the committed-evidence premise as a known debt (`plan.md` §H, `decision-index.md`). |
| 0.5.0 | 2026-10-19 | Shrunk-scope revision per the operator's ruling of 10-19 (AskUserQuestion; board record `d-20261018T214213Z-94a7`, decision-index.md Q10). The done-path cluster — the M1 → M2a → M2b chain: the automatic done receipt and its revocation, the removal of T18 and the T19 reservation with the replacement edge T29, and the lane own-card `done` exception — moves out of this SPEC to the follow-up card t1878, which owns authoring it (Out of Scope carries the moved note). The CLI-alias policy is reversed by the same ruling: no warning alias and no one-release grace — `moai gtd` is removed immediately in this card and answers unknown command, and every residual `moai gtd` / `/moai:gtd` / `moai:gtd` / `gtd.md` reference moves to `moai todo` in this card (the zero-residual completion judgment; the queue workflow document's rename to `workflows/todo.md` is a consequence, superseding Q3's path preservation). Requirements 25 → 15, criteria 25 → 15: REQ-TU-002/003 (warning) and REQ-TU-015 to REQ-TU-021, REQ-TU-025 (done path) are retired; AC-TU-001/002 and AC-TU-014 to AC-TU-020, AC-TU-023 are retired; surviving identifiers keep their numbers (gaps stand; no identifier is reused with a new meaning). One plan delta audit on the shrunk scope follows (the rule C counter restarts by the operator's act). |
| 0.6.0 | 2026-10-19 | Repair revision for the proxy codex leg's blocking findings (leader repair ruling of 10-19; receipt `rcpt-54ca7c42603570ff69d6589e`; decision-index.md Q11). REQ-TU-011 gains one named UUID-namespace preservation exception — the file:literal pair `internal/factory/gtd_engage.go` × `"moai:gtd-card:"`, the hash namespace of every persisted card UUID — excluded from the grep-0 target because REQ-TU-012 freezes the derived UUID bytes; the completion judgment reads zero residual hits outside that one exception (AC-TU-010 carries the excepted-file row), and AC-TU-011 gains the byte-invariance row for the namespace literal and the hash-input construction. The rule sources' `paths:` key updates move from M8a into M6, so the workflow-document rename, its consumer edits, and the rule-path updates land in one milestone (plan.md M6/M7/M8a/§H/§J); M8 closes with script idempotence and the recorded residual RED, the full-residual GREEN being M9's completion condition. No identifier added or removed (15 and 15). |
| 0.7.0 | 2026-10-19 | Final repair revision for the round-2 proxy audit's blocking findings D1 to D3 (leader final repair ruling of 10-19; receipt `rcpt-03225d18a7dc4a018105dc3d`; decision-index.md Q12; round 3 of the counter — the last before a FAIL holds the card). D1: plan.md M6's split seam keeps the workflow-document rename, the `TestWorkflowRulePathsPinned` rewrite, the five rule `paths:` keys (template and deployed), the `factory-dispatch-mechanics.md:199` cross-reference, and the two document-test path rewrites in one unit — the one same-milestone unit acceptance.md §3 requires — and a run-time split moves only the slash-command, router-skill, and published-skills edits to M6b. D2: the namespace test is planned in M5 — file `internal/factory/card_uuid_namespace_test.go`, named outside the `gtd_*.go` glob AC-TU-011's diff pin guards; package `internal/factory`; name-exact selection — and its expected literal is assembled non-contiguously or read from `gtd_engage.go`'s bytes, so no tracked file other than the exception matches the residual grep (AC-TU-011 carries the constraints). D3: `TestNoResidualGTDSpelling` assembles its four-pattern alternation from non-contiguous parts and writes its planted-mutant fixture into a run-time temporary directory (no tracked fixture), so the completion grep over the tracked tree sees neither, and the completion judgment adds the grep's own re-run over the test file as a confirming row (AC-TU-010). No identifier added or removed (15 and 15). |

## 1. Background and problem

### 1.1 What the card asks for

The card carries two redesign decisions of the 3.2 plan (D3 "todo unification" and the done-path subset of D4) plus one addendum:

- **D3.** `moai todo` becomes the only canonical queue spelling. Operator revision 5 (10-19): `moai gtd` is removed immediately — no warning alias, no grace release — and every residual reference moves to `moai todo` in this card. The six verbs that exist only under `moai gtd` are cleaned up. The slash command, the skill, the documents, and the rules carry the canonical name.
- **D4 (subset).** A lane may close its own card. The git-flow edges T18 and T19 of the card transition table are removed. Completion criterion: done is automatic for the most recent ten landed cards. (Moved out of this SPEC to the follow-up card t1878 by the operator's ruling of 10-19 — see §5 Out of Scope.)
- **Addendum.** After a card is picked, no verb records its spec_id.

### 1.2 Measured state (tree `026f48658f`; commands and verbatim outputs are in `research.md` §R.2)

1. **Two spellings, one tree.** `moai gtd` is built from the same verb tree as `moai todo` and adds six verbs that `moai todo` does not have: capture, clarify, organize, reflect, engage, answer. The guidance (the queue workflow skill) tells the reader to run `moai todo capture|clarify|organize|reflect|engage|answer`; `moai todo reflect` ends with `Unknown command "reflect" for "moai todo"` (installed build, see `research.md` §R.1 for its provenance). This is the guidance-versus-execution mismatch of t1590.
2. **The canonical name is inverted in part of the shipped surfaces.** `AGENTS.md`, the command help, the router skill, and the header of the queue workflow document already call `moai todo` canonical. The slash command files do not: `/moai:gtd` is described as the canonical entry, `/moai:todo` as its compatibility alias that dispatches `gtd`, the published skills copy that, and a golden test pins the arrangement. The queue workflow document is named `workflows/gtd.md`; that path is an identifier wired into five rule files' `paths:` keys, a generated embed line, a recorded fixture, and seven test files (`research.md` §R.10), so it keeps its path.
3. **The spelling is wide.** 371 tracked text files contain "gtd" (385 paths when binary files are counted); 204 of them sit outside the historical SPEC directories; 73 of those carry the command spelling `moai gtd` (255 lines). Some of the 73 are executable consumers (an argument vector in the board mod, a shell script default, an in-tree error message), not prose.
4. **Persisted names must not move.** The gtd tables (`gtd_items`, `gtd_relations`, and four more), the `gtd-<16 hex>` item id, and the `.moai/gtd/answers` directory are persisted formats; 47 files under `internal/` carry GTD identifier-shaped tokens.
5. **A lane cannot close its own card.** The lane guard admits seven read-only verbs; `done` is not among them. The leader-approval receipt gate then binds every close of a factory-linked card, and only the leader mints a receipt, so a landed card whose lane is gone waits for a leader act.
6. **T18 and T19.** T18 is `merged-local → done`, refused while the repository has a remote, human-decided, receipt-gated; T19 is the reserved `pushed → ci-green` edge. The repository's own configuration is still `mode: manual`, `workflow: git-flow`, and `SPEC-GITHUB-FLOW-DEFAULT-001` is `in-progress`, so the "GitHub Flow transition complete" premise of the decision is not met here today. The removal and its replacement edge moved to the follow-up card t1878 (operator revision 5).
7. **No post-pick spec_id path.** The queue spec_id is set only by `moai todo next <n> --spec` at pick; the factory card row's spec_id is set by `moai factory assign --spec` only while the card is `picked`, and a later attempt is refused ("its hints and pointers change only while picked"). A card dispatched without a spec_id wedges kickoff, whose SPEC lookup is `<worktree>/.moai/specs/<spec_id>`. The observed case (board record `d-20261010T001732Z-d4e8`, read with `moai decision read --scope card:t1641`) is a card that was **leased** without a spec_id whose SPEC the lane authored afterwards; the leader stamped the factory row by a direct database edit. Two guards read the identifier: the plan-audit verdict admission (`plan-audit → kickoff`) and the audit kickoff (`kickoff → run`), so the wedge sits in the states between lease and kickoff (`research.md` §R.15).

### 1.3 Why

One canonical name removes a standing documentation contradiction and a standing runtime mismatch (items 1 and 2) and ends the "fully supported alias" posture that keeps every guidance surface bilingual. The done-path changes — which would have removed the human act the 3.2 redesign measured as the dominant wedge in completing cards — moved to t1878 with the rest of that cluster. The spec_id setter removes a known kickoff wedge that has so far been repaired by direct database edits.

## 2. Scope and tier

**Tier L.** Reasons, each measured: more than 15 files are affected (101 tracked files carry a residual `moai gtd` / `/moai:gtd` / `moai:gtd` / `gtd.md` reference outside the excluded classes at HEAD `3c1809e2f1`, plus the Go packages `internal/cli`, `internal/homestate`, `internal/factory`, `internal/template`); the change is estimated at well over 1,000 lines once tests are counted; and it is constitutional in the sense of the tier table: it edits the canonical rule and skill surfaces, `AGENTS.md`, and the CLI command surface whose shape tests pin. Artifact set: spec.md, plan.md, acceptance.md, design.md, research.md (plus decision-index.md and progress.md, not counted). Requirement and criterion counts are within the Tier L ceilings (15 and 15 against 25 each, after the revision-5 shrink).

Two tracks, kept separable by the dependency graph of `plan.md` §A; the former Track B (the done path) moved to the follow-up card t1878:

- **Track A, one canonical name with `moai gtd` removed (REQ-TU-001 to REQ-TU-014).**
- **Track C, the spec_id setter (REQ-TU-022 to REQ-TU-024).**

## 3. Requirements (GEARS)

Questions Q1 to Q9 are decided in `decision-index.md`; each answer is written into the requirement it affects. A requirement names observable behavior; identifiers and file names live in `design.md`.

### 3.1 Track A, one canonical name

**REQ-TU-001 (Ubiquitous).** The `moai todo` command tree shall be the single registration of every queue verb, and the `moai gtd` command tree shall be removed entirely (the bare `gtd` parent and every former `gtd` verb), so that after the change `moai gtd` — with no argument, with any former verb, or with any other argument — exits non-zero with the unknown-command error cobra produces for an unregistered command, and no verb has a second handler under the `gtd` spelling. (Operator revision 5 reverses the warning-alias design of Q2; the removal is immediate, with no alias, no warning, and no grace release.)

**REQ-TU-004 (Ubiquitous).** Each of the six stage verbs (capture, clarify, organize, reflect, engage, answer) shall be registered on the `moai todo` command tree with its present arguments, flags, and behavior, so that `moai todo <verb>` runs the handler the `moai gtd` tree runs today; no stage verb shall be deleted. (Decision Q1.)

**REQ-TU-005 (Ubiquitous).** Every `moai <command> <verb>` spelling that a skill, rule, workflow document, README, docs-site page, help text, or error message instructs the reader to run shall resolve to an executable verb of the command it names. (The t1590 clause: guidance that cannot be executed does not land.)

**REQ-TU-006 (Where).** Where a stage verb is registered on `moai todo`, the lane-session guard shall classify it explicitly: `reflect` without `--rebuild-projection` is read-only and admitted; `reflect --rebuild-projection` (it writes the private projection files), `capture`, `clarify`, `organize`, `engage`, and `answer` are mutating and refused with the standing refusal text; a stage verb without a declared classification shall fail a test rather than inherit a default.

**REQ-TU-007 (Ubiquitous).** The canonical slash entry shall be `/moai:todo` and shall dispatch `todo`; the `/moai:gtd` slash command file shall be deleted from the template source and the deployed tree (no `/moai:gtd` entry remains in any surface), and the `moai-gtd` published skill shall leave the command-to-skill publisher's set and `.agents/skills/`; the published skills shall equal the regenerated publisher output; the router skill shall route `todo` as the only queue spelling; and the queue workflow document shall be renamed to `workflows/todo.md` (template copy, deployed copy, and the deployed-only `.moai/workflows/` copy) and shall name `moai todo` canonical in its header and body. (Operator revision 5; the rename is a consequence of the zero-residual judgment of REQ-TU-011, which supersedes Q3's path preservation for this document — the persisted formats of REQ-TU-012 are untouched.)

**REQ-TU-008 (Ubiquitous).** Every executable consumer of the `gtd` spelling (argument vectors, shell scripts and their defaults, in-tree error and help messages) shall be changed to `todo` by reviewed edits, not by the bulk text substitution.

**REQ-TU-009 (Ubiquitous).** The bulk replacement of the `moai gtd`, `/moai gtd`, `/moai:gtd`, `moai:gtd`, and `gtd.md` (as a path string) spellings in documentation, template, fixture-text, and rule files shall be performed by one repeatable script that touches only the classes the substitution table allows and leaves every excluded class (historical SPEC directories, reports, the changelog, binary files, JSONL recordings, Go identifiers and persisted names, and the always-loaded files `AGENTS.md` and its template, which M9 edits by hand) byte-identical.

**REQ-TU-010 (Event-driven).** When the substitution script runs on a tree it has already converted, it shall change no file and report zero replacements.

**REQ-TU-011 (Ubiquitous).** After the change, the residual-reference grep `moai gtd|/moai:gtd|moai:gtd|gtd\.md` over tracked files — excluding `.moai/specs`, `.moai/reports`, `CHANGELOG.md`, and `*.jsonl` (history) — shall print nothing outside one named preservation exception, and a repository test shall enforce that. No general allowlist and no grace lines replace the exclusions: the operator's completion judgment (revision 5, decision Q10) is zero residual references, not a ledger of permitted ones. The one exception, ordered by the leader's repair ruling (revision 6, decision Q11, receipt `rcpt-54ca7c42603570ff69d6589e`), is the file:literal pair `internal/factory/gtd_engage.go` × `"moai:gtd-card:"`: `gtdCardUUID` derives every persisted card UUID from `sha256("moai:gtd-card:" + itemID)`, the literal matches the `moai:gtd` alternative, and REQ-TU-012 freezes the derived UUID bytes, so the two clauses are satisfiable together only by excepting these bytes. The exception is exactly this file and this literal — any other matching line in the file, and any other file, stays residual — and AC-TU-011 pins the namespace input and the hash construction byte-invariant. Grep-0 means zero residual hits outside the named exception. Measured baseline before the change: 397 matching lines in 101 tracked files at HEAD `3c1809e2f1` with the grep above (this revision's measurement; the operator measured 387 on 10-19; the figure includes the exception's one line today).

**REQ-TU-012 (Ubiquitous).** Internal identifiers and persisted formats shall keep their present names and values: the Go identifiers and their names (wherever the deletion of the `gtd` command tree moves their defining file), the `gtd_*` tables, the `gtd-<16 hex>` item id pattern, and the `.moai/gtd/answers` directory. (Decision Q3, as narrowed by revision 5: the queue workflow document's path is no longer preserved — REQ-TU-007 renames it; everything Q3 preserved besides that document stands.)

**REQ-TU-013 (Ubiquitous).** Every changed file that has a template source shall be edited in `internal/template/templates/`, the embedded catalog shall be regenerated by the project build, and the deployed mirror shall equal the regenerated output wherever a mirror guard exists.

**REQ-TU-014 (Ubiquitous).** The `moai todo` row of `AGENTS.md` and of its template mirror shall name `moai todo` canonical and shall carry no `moai gtd` alias statement at all — no "fully supported compatibility alias", no warning-alias sentence, no `moai gtd` spelling anywhere in the file — and the file shall stay within the Codex contract byte ceiling.

### 3.2 Track C, the spec_id setter

**REQ-TU-022 (Event-driven).** When the spec_id setter is invoked by the leader with a well-formed SPEC identifier on a card whose queue row is `picked` and whose factory row is in `picked`, `assigned`, `leased`, `plan`, `plan-audit`, or `kickoff` (the states from the pick through the last gate that reads the identifier), or that has no factory row, the setter shall write the identifier to each carrier whose spec_id is empty (the factory card row, with a field-change event, and the queue row) under the queue lock, and shall never leave a carrier holding a value other than the requested identifier; if a failure leaves one carrier written and the other empty, the setter shall exit non-zero naming both carriers, and a repeat invocation naming the same identifier shall fill the empty carrier and succeed in every card state that is not terminal, because a repeat that completes a pair swaps no held value. (Decisions Q8 and Q9.)

**REQ-TU-023 (Event-driven).** When the caller is a lane session, the identifier is malformed, the card already carries a different spec_id in either carrier, the factory row is in a terminal state, or the factory row is in `run` or a later non-terminal state and the invocation does not complete a half-written pair (REQ-TU-022), the setter shall refuse and change neither carrier; where both carriers already hold the identical value and the caller is the leader, the setter shall succeed as a no-op in every non-terminal state.

**REQ-TU-024 (Ubiquitous).** `next --spec`, `factory assign --spec`, and the unpick clear shall keep their present contracts, and a card whose spec_id the setter wrote shall pass kickoff's SPEC lookup exactly as one whose spec_id was set at pick, including a card the setter filled while it was `leased`.

## 4. Constraints

- **Methodology.** `quality.yaml` sets `development_mode: tdd`: every milestone starts with a RED step whose failure is observed for the stated reason before the implementation lands.
- **No test skips and no timeout increases** anywhere in the change.
- **Verification scope.** The landing re-measure runs the changed packages and their direct importers; inside `internal/cli` it runs a name-exact `-run` selection of the tests that name the changed files or identifiers; the remainder of `internal/cli` is carried to batch CI as an explicit Gap.
- **Template-First.** Sources live under `internal/template/templates/`; deployed `.claude/` and `.moai/` copies that have a template source follow it and never lead it.
- **No time estimates.** Priorities and ordering only.

## 5. Out of Scope

### Out of Scope — moved to the follow-up card t1878 (operator revision 5)

[MOVED to t1878: the done-path cluster — milestones M1/M2a/M2b — comprising the automatic done receipt (in-transaction issuance behind the three machine facts, the reserved issuer `auto`, revocation, the closing-act logging of both records, `factory.VerifyRemeasureBinding`), the removal of the T18/T19 transition-table edges with the replacement edge T29 and the push-gate no-configuration guard, and the lane own-card `done` exception with the both-records close.] The new card owns authoring these requirements and criteria (the retired identifiers REQ-TU-015 to REQ-TU-021, REQ-TU-025 and AC-TU-014 to AC-TU-020, AC-TU-023 are not reused here), and it takes over with them the known debt KD-1 (the committed-evidence premise, below) and the plan-audit iteration-4 hardening findings D38 to D40 (a replacement-ref-safe landing read, the no-configuration containment tightening, the submodule-diff blind spot, `plan-audit-iter4.md`).

### Out of Scope — internal renames and persisted formats

- Renaming Go identifiers, `gtd_*.go` file names, or the `internal/factory` types that carry GTD in their names (decision Q3).
- Renaming the `gtd_*` tables, the `gtd-<16 hex>` item id, or the `.moai/gtd/answers` directory, and any data migration they would need.
- Retiring or renaming the `moai-gtd` docs-site pages as pages (their content migrates to `moai todo`; the pages themselves, their paths, and the menu entries are content-migration surfaces of M8, not deletions).

### Out of Scope — other 3.2 stage 2 work

- The landing-to-done machine transition driven by three observers (a SessionEnd observer, a `factory complete` observer) and the general mechanism that issues approval receipts at observation time: card t1639 (factory v2, currently on hold) owns it. This SPEC delivers no observer, no receipt, and no close-path change; the receipt work this card once carried moved to t1878.
- The `pushed → done` edge and its decider path (a card whose factory row the leader already moved to `pushed`), the lease and expiry repairs, and the transition-table redesign (cards t1621 and t1639); the T18/T19 removal this card once carried is t1878's.
- A lane-side surface for recording a spec_id (no sibling SPEC carries it: `research.md` §R.8).
- A fourth receipt fact that a lane cannot author, such as a verdict that cites an audit receipt id the audit-receipt store corroborates: the verdict file carries no machine-readable receipt line today and receipts exist only where a required backend is configured (`research.md` §R.14). The lane-authored-verdict residual belongs to t1878 with the receipt machinery it would compensate.
- Tightening the merge edge's own re-measure callback (T16) to compare the record's `tree` and `base` fields: the adjacent gap in the older callback is recorded in `research.md` §R.14 and left to its owner; t1878's receipt re-checks the fields on its own path.
- Editing the draft `SPEC-FACTORY-COMPLETION-RECOVERY-001`: the narrowing of that draft's REQ-FCR-014 for a lane-triggered mint moves to t1878 with the mint (`research.md` §R.8); the draft itself is not edited here or there.
- Web pages, worktree relocation, and the CLI split (cards t1609, t1640, and stage 3).
- Requiring lanes to commit their `.moai/reports/<card-id>/` evidence. `.gitignore:235` ignores `.moai/reports/*` by the operator's policy; the committed-evidence premise it breaks was fact 2 of the moved receipt design, so the premise, its debt statement (KD-1), and any fix move to t1878 together (`plan.md` §H carries the pointer).

### Out of Scope — history and release mechanics

- Editing the historical SPEC directories, `.moai/reports/**`, or `CHANGELOG.md`.
- Pushing, opening pull requests, or batch landing; the leader owns them.

## 6. References

- Card t1641 text and absorbed cards t1602, t1590 (queue records; quoted in `research.md` §R.3).
- `.moai/reports/t1641/plan-audit.md` and `.moai/reports/t1641/plan-audit-iter2.md` (the two plan audits whose findings revisions 2 and 3 close; the per-defect closure map is in the SPEC's `progress.md` §E.1).
- `.moai/reports/moai-32-master-upgrade-plan-20261010.md` (stage 2, item 3.2-2-4) and `.moai/reports/moai-adk-32-cross-harness-redesign-20261010.html` (decisions D3 and D4). Both are untracked report material: they are evidence of the operator's decision, not authority citations.
- `AGENTS.md` §11 (the `moai todo` verb row).
- `.claude/rules/moai/workflow/factory-dispatch.md` (lane spawn authority and the leader's remaining role).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` (frontmatter and artifact statelessness).
