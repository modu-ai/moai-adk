# Decision Index — SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001

Revision 0.1.0 (2026-10-19). Each row states one decision. Rows Q13, Q14, and Q17 come from the parent's decision index (SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001, revision 0.3.1) under the numbers the parent gave them. Every observation below was taken on tree `74ad197380a0711cdb9816d4e37c94d29af730ea` unless it says otherwise; evidence ids (C-n) point to acceptance.md §C, where the command and the verbatim output are recorded.

Labels: DECIDED and POLICY-COVERED need a committed anchor in the authority register (committed artifacts only). None of the three rows has one, so each routes to FOUNDER. `.moai/docs/docs-site-i18n-rules.md` is a committed document outside the register; the rows cite it as a fact, not as an authority. The operator verdicts of Q13 and Q17 are board records, cited by id (`d-20261019T015908Z-220a`); they are recorded in the `Operator verdict:` field as the operator stated them and are not repeated as a recommendation.

No row carries a recommendation. No row carries a `Default:` or a `Planner default applied:` line: Q14 is product-level (a user-facing page is removed or changed; fail-closed, no Default), and Q13 and Q17 are decided.

| Row | Label | Class | Verdict state | Blocks | Kickoff can proceed without it |
|-----|-------|-------|---------------|--------|-------------------------------|
| Q13 | FOUNDER | implementation-level | decided by operator, `d-20261019T015908Z-220a` | nothing open | yes |
| Q14 | FOUNDER | product-level | open | M3 (the four pages and the four `_meta.yaml` entries), AC-SWF-005, AC-SWF-006, and the page parts of AC-PRC-001 and AC-SWF-002; REQ-SWF-004, -005, -006 apply by the answer | yes as far as M1, M2, and the `_index.md` part of AC-PRC-001; M3 and M4 wait. The Kickoff in its operator form answers Q14 then |
| Q17 | FOUNDER | implementation-level | decided by operator, `d-20261019T015908Z-220a` | nothing open | yes |

### Q13: Is each stale citation deleted or replaced?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: closed by the operator verdict below. The determination the verdict needs, whether a current equivalent exists for each citation, was measured for this child: every unresolved citation in the twelve files is the design family (27 lines, C-1; C-2 finds no other unregistered verb), `/moai design` is not a registered command (C-8), the workflow it names is `user-invocable: false` (C-9), and the four routing files do not name the workflow as an entry (C-7). No current equivalent that a reader can type exists, so the verdict resolves to deletion of the lines that hold the citations; the determination is re-run at M1 (plan.md §B). The affected lines per option, for the record: the design table rows of `_index.md` (en 1, ja 1, zh 1, ko 2; C-1); the `_meta.yaml` title (4 lines; C-1) and the page lines (5, 5, 4, 4; C-1), which depend on Q14 because a page that is kept needs a title.
Class: implementation-level
Requirement text changed by the answer: REQ-SWF-002 (replace where an equivalent exists, delete where none exists; no prose replacement that names the same command), and the pass conditions of AC-PRC-001, AC-SWF-002, AC-SWF-003.
Operator verdict: replace with the current equivalent where one exists, delete the line where none exists — operator, `d-20261019T015908Z-220a` (board record, decided by the operator through the question channel, 2026-10-19, scope standing, kind ruling; applies to docs pre-rewrite children whose scope depends on Q13 or Q17)

### Q14: Is the `moai-design` page kept, rewritten, or removed?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the page documents something real that is not a slash command. `.claude/skills/moai/workflows/design.md` carries `user-invocable: false` (C-9), `/moai design` is not registered (C-8), `.claude/rules/moai/design/constitution.md` records the older `/moai design` production route as retired, and `.claude/rules/moai/workflow/spec-workflow.md` § Conditional Design Route describes the live route: UI-surfaced SPECs go plan → design → run, entered after plan-audit PASS and Implementation Kickoff Approval, driven by the `manager-design` agent. No routing workflow names the design workflow as an entry (C-7), so a rewrite has no command to point readers at; it can describe the route and nothing typed. The page is 82, 82, 68, and 68 non-blank lines (en, ko, ja, zh; the ja and zh pages lack the usage-example block) and cites the design family on 5, 5, 4, and 4 lines (C-1). It is linked from the four `_index.md` files, registered in the four `_meta.yaml` files, and also referenced by `docs-site/data/menu/main.yaml:208` (C-5), a file outside the twelve. The three options, each stated by what it does to the twelve files and the neighbours:

- (a) Remove the page. Immediate result: the four page files are deleted and the `moai-design` entry is removed from the four `_meta.yaml` files (8 of 12 files are edited, 4 deleted). Side effects: the public URLs `/<locale>/workflow-commands/moai-design` stop serving the page (no redirect is in this change; the redirect rules are REDIR's, Q15 relates to `/design`, not to this path); `main.yaml:208` and `.locale-parity-baseline:51` name a page that no longer exists (O-2, O-3); the Hugo menu template hides an entry whose page is missing (read from `layouts/partials/menu.html`, G-4). Requirement text applied: REQ-SWF-004; AC-SWF-005 applies.
- (b) Rewrite the page. Immediate result: the four pages describe the conditional design route with no slash form and with a non-slash title; the four `_meta.yaml` titles are retitled (12 of 12 files are edited). Side effects: the page stays at its URL; four locales of new prose are written (translation fidelity is a review item, G-5); the ja and zh pages are brought to the same structure as en and ko. Requirement text applied: REQ-SWF-005; AC-SWF-006 (rewrite part) applies.
- (c) Mark the page internal-only. Immediate result: the page keeps its text, and each line that still cites `/moai design` carries a marker, with an opening statement that the workflow cannot be typed (12 of 12 files are edited). Side effects: REQ-PRC-001 is amended by REQ-SWF-006 (a documented exception to "cites only resolving commands"); the GATE checker must exempt the marker (G-6), through the Q24 example marker or a named exemption, which is a cross-child interface. Requirement text applied: REQ-SWF-006 (the amendment of REQ-PRC-001); AC-SWF-006 (marker part, unadopted until a token exists).

Interaction with the Q13 verdict: under every option the `_index.md` design rows are deleted (Q13 verdict, no equivalent), so under (b) and (c) the kept page is no longer linked from `_index.md`; a non-slash pointer row would be a new line, part of the Q14 answer if wanted.
Class: product-level (the page is a user-facing document; option (a) removes it and option (c) amends a requirement)
Operator verdict:

### Q17: Do the same edits land in en, ja, and zh in the same change?
Label: FOUNDER
Authority anchor: none in the authority register. The four-locale same-PR obligation is stated in `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65 ("동일 PR 내에서 en/zh/ja 모두 반영"; "ko/en/zh/ja 4개 locale 해당 파일 모두 수정"), a committed document outside the register, cited as a fact.
Why unresolved: closed by the operator verdict below. For this child it means the edits to all twelve files (or the remaining eight and four deletions) land in one change. The locales are not line-for-line twins (C-1, C-4): the Korean `_index.md` holds two design rows and a `/moai goal` summary row that the other locales lack; the en and ko pages hold five design lines, the ja and zh pages four. The verdict brings the asymmetries level by the same deletions in all four locales, and AC-SWF-003 fixes the closing counts (20, 20, 21, 20), with Korean one above the others only because of the untouched `/moai goal` row.
Class: implementation-level
Requirement text changed by the answer: REQ-SWF-003; AC-SWF-003 Then 3 (one commit changes the four `_index.md` files).
Operator verdict: all four locales (ko·en·ja·zh) change in the same change/card — operator, `d-20261019T015908Z-220a` (board record, decided by the operator through the question channel, 2026-10-19, scope standing, kind ruling; applies to docs pre-rewrite children whose scope depends on Q13 or Q17)

## Open items for the leader (not decision rows)

- O-1 — Who edits the `_meta.yaml` files. See plan.md §A; both readings are written there.
- O-2 — `docs-site/data/menu/main.yaml` lines 204 to 208 carries `/moai design` in four locales and `ref: /workflow-commands/moai-design`. It is a design-family citation outside the card's twelve files and the curator's single-writer file. Reading 1: add it to this change (the Q13 verdict would delete or retitle the entry, and Q14 decides the `ref`). Reading 2: leave it to another card, accepting that a stale citation stays in the sidebar data after this child closes.
- O-3 — `docs-site/.locale-parity-baseline` line 51 lists `workflow-commands/moai-design.md`. Under Q14 = remove it names a page that no longer exists; the file's header says to regenerate only to prune. The consumer of the file was not found in the paths searched (acceptance.md G-7).
- O-4 — The count sentences on `_index.md` ("All subcommands (16)" and the sentence naming 16 subcommands, four locales, C-10) are not slash citations. Deleting the design row leaves 15 table rows beside the number 16; the registered count is 17 (C-8). Reading 1: out of scope (observation). Reading 2: part of this change as a consequence of the Q13 verdict.
