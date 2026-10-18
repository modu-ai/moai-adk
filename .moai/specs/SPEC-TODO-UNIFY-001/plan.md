# Plan: SPEC-TODO-UNIFY-001

Tree basis: `026f48658f` (branch `WT-todo-unify`, clean). Methodology: TDD (`quality.yaml` `development_mode: tdd`). Priorities and ordering only; no time estimates.

## §A Decision-reversibility ranking

Sections and milestones below are ordered by how likely the decision they rest on is to change, highest first. A data-model or interface decision comes before a mechanical step, so review attention lands where a reversal would cost most.

| Rank | Decision | Why it may change | Gated by | Milestone |
|------|----------|-------------------|----------|-----------|
| 1 | Remove the git-flow T18/T19 edges from the card transition table | Removes the no-remote completion path; premise "GitHub Flow transition complete" is unmet in this repository; two draft SPECs touch the same rows | OQ-6 | M1 |
| 2 | Who may close a card without an operator, and under which receipt policy | Changes who can complete a factory-linked card; overlaps card t1639 | OQ-4, OQ-5, OQ-7 | M2 |
| 3 | Shape and caller of the post-pick spec_id setter | A new verb on a shipped command; widens or keeps the lane guard | OQ-8, OQ-9 (implementation-level, default applied) | M3 |
| 4 | Disposition of the six stage verbs | Three incompatible readings in the card and its absorbed cards | OQ-1 | M4 |
| 5 | Warning alias behavior and the removal-release constant | User-visible stderr contract | OQ-2 (implementation-level, default applied) | M5 |
| 6 | Slash, skill, and workflow-document names | Renames of shipped surfaces | none beyond OQ-1 | M6 |
| 7 | Substitution classes and the script | Mechanical; reversible by a revert | OQ-3 fixes the class table | M7 |
| 8 | Applying the script (docs-site, README, templates, deployed mirrors) | Mechanical | none | M8 |
| 9 | Always-loaded rule text and AGENTS.md | Edited last (cache directive 3); size-bounded | none | M9 |

**Independence guarantee.** M1, M2, and M3 touch disjoint code and depend on neither Track A nor each other, so an OQ-4/5/6 answer that cancels M1 or M2 leaves every other milestone valid. Track A runs M4, then M5, then M6, then M7, then M8. M9 is last.

## §B Open Questions

Every question below is a real fork in the card text or in the measured tree. None is guessed. Each is mirrored as a row of `decision-index.md` with the same number. A `Default` names the option the published default rule ranks first; it is a policy application and not a recommendation.

[NEEDS CLARIFICATION: OQ-1 — what "clean up the 6 gtd subverbs" means]

**OQ-1. The six stage verbs (capture, clarify, organize, reflect, engage, answer).** Three readings exist.
- (a) **Absorb all six into `moai todo`** (card t1590, direction A: register the six factories on the canonical tree; `moai gtd` then delegates for every verb). Consequence: nothing is deleted; the guidance already tells readers to run `moai todo <stage verb>`, so it becomes true.
- (b) **Absorb the verbs that measured use justifies and delete the rest** (the redesign's D3 sentence). Consequence: a user-facing feature is removed; the deletion set is decided by the M4 usage inventory.
- (c) **Leave the six under `moai gtd` only** (card t1602 item 1: "GTD-stage verbs stay under `moai gtd`"). Consequence: contradicts "gtd warns, then delegates" and "removed after one release" for those six; the alias could not be removed without deleting them.

Measured inputs: the CLI verbs have no in-tree caller. `goal.go` calls the factory library functions directly (`ReflectGTDStore`, `EngageGTDItem`), not the CLI verbs; `gtd answer` writes `.moai/gtd/answers/<t-id>.md` and no tracked code reads it; the verbs are documented in the queue workflow skill, four READMEs, and four docs-site locales. Design robustness: REQ-TU-001 to REQ-TU-005 and AC-TU-001 to AC-TU-005 hold under (a) and (b); under (c) REQ-TU-001/002 narrow to the verbs that exist on both spellings and the removal release cannot be named for the six. Default (rule: preserves current behavior): (a). Alternate: (b).

[NEEDS CLARIFICATION: OQ-2 — release naming for the grace period]

**OQ-2. Which release first carries the warning and which removes `moai gtd`.** Options: (a) the warning text names a literal release held in one constant; (b) the text says "the next release" with no number. Consequence: (a) needs the release plan's number before sync; (b) cannot be tested for an exact removal release. Design: the removal release is one constant and AC-TU-001 and AC-TU-002 assert the warning contains it and that it is non-empty; both options satisfy that. Default (rule: smaller user-visible surface; neither option changes current behavior and each undoes by one revert): (b), the constant holding the phrase "the next release". Alternate: (a), the literal taken from the release plan.

[NEEDS CLARIFICATION: OQ-3 — depth of "rename gtd to todo everywhere"]

**OQ-3. Rename depth.** (a) **User-facing spelling only**: command, slash entry, skills, documents, rules, guidance; Go identifiers, `gtd_*.go` names, the `gtd_*` tables, the `gtd-<16 hex>` item id, and `.moai/gtd/answers` stay (this SPEC's default reading). (b) **Everything**: also the internals. Consequence of (b): persisted formats change, so existing queue databases need a migration and a dual-read of the old item id; about 47 files change identifiers. Design: REQ-TU-012 and AC-TU-011 pin (a); (b) would be a separate SPEC with its own migration, and nothing in M1 to M9 blocks it. Default (rule: preserves current behavior): (a). Alternate: (b).

[NEEDS CLARIFICATION: OQ-4 — scope split with card t1639 (factory v2)]

**OQ-4. Is the landing-to-done machine transition (three observers) and automatic receipt issuing in this card's scope or t1639's?** The master plan assigns "landing to done machine transition (three observers), receipts automatic" to t1639 (on hold, stage 2-2) and "lane own-card done, T18/T19 removal" to t1641 (stage 2-4), while t1641's completion criterion reads "done is automatic for the most recent 10 landed cards". Options: (a) this SPEC delivers only the lane own-card path and the receipt policy seam (the card's listed scope); the criterion is met through the lane path plus the existing leader `auto-done` scan. (b) this SPEC also delivers the SessionEnd and `factory complete` observers. Consequence of (b): overlap with t1639 and a hook that must satisfy the advisory-check discipline (bounded, cached, never blocking). Design: M2 isolates the close admission behind one seam, so (b) adds observers without changing M2. Default (rule: smaller user-visible surface): (a). Alternate: (b).

[NEEDS CLARIFICATION: OQ-5 — what satisfies the leader-approval receipt gate for a lane's own-card done]

**OQ-5. Receipt policy for the own-card `done`.** The gate binds every close of a factory-linked card, and only the leader can mint a receipt. Options: (a) landing evidence of the closed auto-done form substitutes for the receipt for that card only; (b) the receipt is issued automatically from the verdict file, the landing evidence, and the remeasure hash; (c) the lane's `done` is admitted by the guard but still needs a receipt the leader issued. Consequence: under (c) the "automatic for the most recent ten landed cards" criterion cannot hold without a leader act, so the criterion would be reported unmet. Design: M2 puts the receipt condition behind one function the AC tests through; the admission matrix of AC-TU-014 to AC-TU-016 is identical under (a) and (b). Default (rule: preserves current behavior; (c) is the only option that leaves the current approval gate unchanged): (c). Alternate: (a). The row is product-level, so the default is recorded but not applied; it stays open for the operator (`decision-index.md` Q5).

[NEEDS CLARIFICATION: OQ-6 — precondition for removing T18 and T19]

**OQ-6. Removing T18/T19 before the GitHub Flow transition completes.** Measured: this repository's own config is `mode: manual`, `workflow: git-flow`; `SPEC-GITHUB-FLOW-DEFAULT-001` is `in-progress`; `PushGateTarget` returns `done` for a card whose repository has no remote, so T18 is today's only no-remote completion edge; draft `SPEC-FACTORY-COMPLETION-RECOVERY-001` REQ-FCR-010 would implement T19, and draft `SPEC-FACTORY-RECORD-REPAIR-001` moved its `pushed → done` item to t1639. Options: (a) remove T18 and the T19 reservation now and name the no-remote completion path; (b) remove only the T19 reservation and keep T18 guarded for no-remote repositories; (c) defer the removal until `SPEC-GITHUB-FLOW-DEFAULT-001` closes. Consequence: (a) leaves no-remote projects without a factory completion path unless a replacement is named; (b) contradicts the decision text for T18; (c) leaves the card's removal item undelivered. Design: M1 is independent; REQ-TU-018/019 are conditional on the answer, and the pinned counts (71 accepted, 370 refused) do not move under (a) or (b) because both pairs are already refused in the pinned fixture. Default (rule: preserves current behavior): (c). Alternate: (b).

[NEEDS CLARIFICATION: OQ-7 — what "done" and "landed" mean in the ten-card criterion]

**OQ-7.** Two records carry done: the backlog queue (archive) and the factory card row (state `done`). "Landed" differs by flow: in this repository's local-main flow the merge is on local `main` while `origin/main` advances by batch. Options for the record: (a) both; (b) queue only; (c) factory row only. Options for landed: (i) merge SHA reachable from the resolved landed ref (the auto-done evidence form); (ii) merge SHA on local `main`. Consequence: a card done in one record and open in the other is the observed wedge, so (b) and (c) leave it possible; (a) changes what a lane's `done` does to the factory row. Design: REQ-TU-020 and AC-TU-019 name "the records OQ-7 names", so a narrower answer narrows the criterion without invalidating M2. Default (rule: preserves current behavior; today the close paths change the queue and the factory row's done is the leader's gate): record (b), landed (i). Alternate: record (a). The row is product-level, so the default is recorded but not applied.

[NEEDS CLARIFICATION: OQ-8 — shape of the spec_id setter]

**OQ-8.** (a) A new verb, `moai todo spec <n> <SPEC-ID>`, writing both carriers set-once. (b) Widen `moai factory assign --spec` to set-if-empty after assignment (factory row only; the queue spec_id, which `contract_report` reads, stays empty). (c) Let `next <n> --spec` re-enter a picked card (contradicts the positive pick gate). Consequence: (b) fixes kickoff but not the queue-side reader; (c) removes a safety gate. Default (rule: preserves current behavior; every existing verb is unchanged): (a). Alternate: (b).

[NEEDS CLARIFICATION: OQ-9 — who may call the spec_id setter]

**OQ-9.** (a) Leader only (the lane guard refuses it like every other queue mutation). (b) A lane for its own card only, when the spec_id is empty (the lane learns the SPEC id at plan phase). Consequence: (a) leaves the plan-phase lane unable to unwedge itself, so the leader must be messaged; (b) widens the lane allowlist a second time. Default (rule: preserves current behavior): (a).

## §C Pre-flight

Run before M1 (read-only; commands are plain, one per call):

1. `git rev-parse --short HEAD` and `git branch --show-current` (expect `026f48658f` or its absorbed successor, and `WT-todo-unify`).
2. Re-measure the substitution baseline: `git grep -I -l 'moai gtd' -- . ':!.moai/specs' ':!.moai/reports' ':!CHANGELOG.md'` (expect 73 paths at `026f48658f`; the list is in `research.md` §R.4). Re-measure after absorbing `main`; never re-cite 73 without re-measuring.
3. Confirm the four open sibling facts have not moved: `SPEC-GITHUB-FLOW-DEFAULT-001` status, `git_strategy.manual.workflow`, `SPEC-FACTORY-COMPLETION-RECOVERY-001` status, `SPEC-FACTORY-RECORD-REPAIR-001` status.
4. Build the tool under measurement from this tree and invoke it by path; record the judging build's commit next to the tree HEAD (verification-claim-integrity §2.2). The installed `moai` at plan time is `ee981a932-dirty`, a strict ancestor 516 commits behind.
5. Windows build check: `GOOS=windows GOARCH=amd64 go build ./...` (the change touches `internal/cli` launch-adjacent files only through verbs; confirm no build-tag regressions).
6. Verification scoping: changed packages and direct importers (§F). No full-suite run; no background load; take the heavy-run slot only for the `internal/cli` name-exact selection, per the lane protocol.

## §D Constraints

- PRESERVE: `internal/factory/gtd_*.go` and `backlog_gtd_schema.go` content (OQ-3 default), the `gtd_*` table names, `gtdIDPattern`, the `.moai/gtd/answers` path, the 71/370 pair counts in `fr_transition_test.go`, the auto-done scan's closed evidence set and three misfire guards, the lane guard's REQ-SD-015 refusal text, the leader-only issuing of receipts (unless OQ-5 answers otherwise).
- No test skips (`t.Skip*`) added; no timeout value raised, in code or in any test invocation.
- Templates are edited in `internal/template/templates/`, never first in `.claude/`.
- Editing always-loaded files (rules without `paths:`, `AGENTS.md`) waits for M9 (cache-aware directive 3); each edit that grows such a file by more than 1,000 bytes carries the byte-size and non-invoking-cost statement in the commit body (rule-authoring duty).
- Commits stage by explicit pathspec; every commit message carries `card t1641`.

## §E Milestones

Each milestone: RED first (the new test observed failing for the stated reason, classified `EXPECTED_RED` per the TDD result contract, verbatim output recorded in `progress.md` §E.2), then GREEN (new test plus the regression set), then refactor. Size figures are illustration-level (files / changed lines) and exist to bound review; a milestone that would exceed roughly 10 files or 400 lines at run time splits at its natural seam before it starts.

### M1 — Remove T18 and the T19 reservation (Priority High; gated by OQ-6)

- **RED.** (1) A table test asserting the transition table holds no `merged-local → done` row and no reserved-edge refusal text; (2) the existing pair-count pin stays green; (3) a refusal-class test: both requests return the undeclared-transition error, not the reserved-edge error. Observed red on the current tree: row `T18` at `card_transition.go:157` and the reserved-edge branch at `:333`.
- **GREEN.** Remove the row, the reserved-edge predicate and its error value, the `guardNoRemote` branch, and the push gate's `done` target for no-remote cards; implement the completion path OQ-6 selects.
- **Files (about 9 / 330 lines).** `internal/homestate/card_transition.go`, `card_evidence.go`, `card_pr_states.go` (comment), `internal/cli/factory_card.go` (push gate), tests `fr_transition_test.go`, `leader_approval_test.go`, `f7_gate_owner_admission_test.go`, `fr_evidence_test.go`, `card_pr_states_test.go`.
- **Packages.** `internal/homestate`, `internal/cli`; direct importers listed in §F.
- **Name-exact selection.** `internal/homestate`: `-run 'TestFR_AC005|TestPushGateTarget|TestLeaderApproval|TestCardPRStates|TestFR_Evidence'` (confirm exact names by `git grep -n '^func Test'` in the five test files at RED time); `internal/cli`: tests naming `factory_card.go` push-gate identifiers.

### M2 — Lane own-card done and the receipt seam (Priority High; gated by OQ-4, OQ-5, OQ-7)

- **RED.** (1) Lane env, card held by that lane, landing evidence present: `done` admitted and every record OQ-7 names ends `done`; (2) the refusal matrix: another lane's card, no landing evidence, every other non-allowlisted verb, each leaves the queue byte-identical; (3) the close passes the approval gate and writes a log row whose source names the own-card path; (4) mutant: a guard that admits `done` for any card fails matrix row 1 of (2).
- **GREEN.** One admission function in the lane guard, evidence via the existing closed evidence form, close via the existing gate path, one seam for the receipt condition.
- **Files (about 5 / 380 lines).** `internal/cli/todo.go` (guard), a new `internal/cli/todo_lane_done.go`, `internal/cli/leader_approval_gate.go`, tests `todo_lane_done_test.go` and an addition to the existing lane-guard test.
- **Packages.** `internal/cli`; `internal/factory` only if the evidence helper moves.
- **Name-exact selection.** `TestTodoLane|TestLeaderApprovalGate|TestAutoDone` plus the new test names.

### M3 — The spec_id setter (Priority High; OQ-8 and OQ-9 default-applied)

- **RED.** (1) Picked and assigned fixtures with empty spec_id: setter writes both carriers, one event row; (2) refusal matrix: a different id, `leased` and later states, terminal states, malformed id, none of which changes either store; (3) an identical value is a no-op success; (4) kickoff's SPEC lookup passes after the setter (no empty-path refusal); (5) `next --spec`, `factory assign --spec`, and the unpick clear keep their present tests green.
- **GREEN.** A set-once rule in the homestate field-update path, a queue-side write under the queue lock, one verb.
- **Files (about 6 / 420 lines).** new `internal/cli/todo_spec.go`, `internal/homestate/card_picked.go`, a small helper in `internal/factory`, tests in the three packages.
- **Packages.** `internal/cli`, `internal/homestate`, `internal/factory`.

### M4 — Stage-verb disposition and registration (Priority Medium; gated by OQ-1)

- **Step 0 (inventory).** Measure the usage of each verb (code callers, docs, hooks, scripts, tests) and write the disposition table to `.moai/reports/t1641/verb-inventory.md`; the table is the input to this milestone.
- **RED.** (1) `TestGuidanceVerbsResolve`: every `moai todo <verb>` and `moai gtd <verb>` token in the queue guidance corpus resolves in its command tree (observed red: six lines of the workflow skill name verbs `moai todo` lacks); (2) the verb-surface parity test rewritten to the disposition table; (3) the lane classification test of REQ-TU-006.
- **GREEN.** Register the absorbed verbs on the `todo` tree, remove them from the `gtd`-only constructor, classify each for the lane guard, delete the deleted verbs, and update the mistyped-verb guard's known-verb list through the tree (no separate list).
- **Files (about 6 / 300 lines).** `internal/cli/todo.go`, `gtd.go`, `gtd_answer.go`, tests `gtd_compat_test.go`, new `todo_guidance_resolves_test.go`, `todo_surface_test.go`.
- **Packages.** `internal/cli`.
- **Name-exact selection.** `TestGTD|TestTodoSurface|TestTodoVerbLeak|TestGuidanceVerbsResolve|TestTodoSkillDoc`.

### M5 — The warning alias (Priority Medium)

- **RED.** `TestGTDAliasWarnsThenDelegates` over every verb (warning once on stderr, names the `moai todo` form and the removal-release constant; stdout and exit equal to the `todo` form; `--json` stdout parses; help requests carry no warning); the old parity assertion `gtdErr != todoErr` is rewritten to "stderr equals the warning plus the `todo` stderr".
- **GREEN.** `gtd` builds from the todo tree and chains one pre-run hook that writes the warning once, reusing the existing deprecation-warning convention (stderr, `Warning:` prefix, `warnDeprecatedAgentModelFlags` as the precedent).
- **Files (about 4 / 220 lines).** `internal/cli/gtd.go`, a small warning helper, tests `gtd_compat_test.go`, new `gtd_alias_warn_test.go`.
- **Packages.** `internal/cli`.

### M6 — Slash command, skills, and the workflow document (Priority Medium)

- **RED.** `TestTodoCanonicalSurfaceGolden` (inverts `TestGTDCanonicalSurfaceGolden`): `/moai:todo` dispatches `todo`; `/moai:gtd` dispatches `todo` and carries the rename notice; published skills equal the publisher output (count stays 17 if both skills remain); the router skill lists `todo` canonical; `workflows/gtd.md` is absent and `workflows/todo.md` present in the template source and the deployed tree; the live-versus-template parity tests read the renamed path.
- **GREEN.** Edit template sources, rename the workflow document (template and deployed), regenerate the published skills with the project build, update the router skill rows.
- **Files (about 10 / 120 lines, mostly renames).** template commands (2), template workflow document, template and deployed router skill, `.agents/skills/moai-todo` and `moai-gtd` (generated), `internal/template/gtd_canonical_surface_test.go`, `internal/cli/todo_skill_doc_parity_test.go`, `internal/cli/todo_skill_doc_test.go`.
- **Packages.** `internal/template`, `internal/cli`.

### M7 — The substitution script and the executable consumers (Priority Low; mechanical)

- **RED.** (1) A fixture test of the script against a temp tree holding one file per substitution class and one per excluded class: expected output exact, excluded files byte-identical; (2) a second run changes nothing; (3) `TestNoResidualGTDSpelling` with a planted-mutant fixture proving the check can go red; (4) the consumer edits: argument vectors in the board mod, the cutover script default, the push-gate help line in `factory_card.go`.
- **GREEN.** Land `scripts/gtd-to-todo-substitute.sh` (design.md §D.3 holds the class table and the dry-run contract), its test, and the reviewed consumer edits.
- **Files (about 6 / 380 lines).** the script, its test, `mods/moai-board/hooks/data.ts`, `mods/moai-board/tests/support.ts`, `scripts/cutover-precheck.sh`, `internal/cli/factory_card.go`.
- **Gap carried.** The board mod's JavaScript tests are not run by any Go selection; the run phase records the runner used, or lists them as an unverified Gap.

### M8 — Apply the script (Priority Low; mechanical, three batches)

- **Batches.** M8a: template sources under `internal/template/templates/` and `internal/cli` and `internal/template` test strings and goldens that are prose; M8b: deployed mirrors in `.claude/`, `.codex/`, `.agents/`, `.moai/docs`, `.moai/policies`, `.moai/project`, `.moai/workflows` (produced by the build, not hand edits); M8c: `docs-site/content` in four locales, the four READMEs, the docs-site menu and `_meta.yaml` rows, and the redirect rows for the retired `moai-gtd` page.
- **Mechanism.** Run the script in dry-run mode first; the printed path set must equal the expected set derived from the pre-flight measurement minus the excluded classes; then apply; then re-run, expecting zero replacements and a green `TestNoResidualGTDSpelling`.
- **Docs-site obligations.** Four locales change in the same commit batch; the locale-parity baseline is not extended; a redirect row for each retired page; the docs-site verification recipe (warning-free build, parity, URL scan) runs before the batch is accepted.

### M9 — AGENTS.md, always-loaded rules, and the closing guards (Priority Low; last)

- **Content.** `AGENTS.md` §11 row (canonical `moai todo`, `moai gtd` a warning alias); the hand-authored statements in the factory-dispatch rule family (the lane own-card `done` exception, the removal of T18/T19, the leader's remaining role) in the template source and the deployed copy and the `.moai/policies` projection.
- **Guards.** The Codex contract ceiling guard (`TestCodexContractByteCeiling`, which measures the live file and the template mirror) stays green (AGENTS.md is 23,780 bytes against 24,576); the always-loaded growth statement is written for any edit above 1,000 bytes; the rule-loading budget measurement (count and bytes of always-loaded files before and after) is recorded.
- **RED.** The residual-ledger test lists the AGENTS.md row as the one allowed alias statement; before the edit it fails on the old wording.

## §F Affected packages and landing re-measure scope

Changed packages: `internal/cli`, `internal/homestate`, `internal/factory`, `internal/template`; `internal/config` through the AGENTS.md ceiling guard. Non-Go surfaces: `scripts/`, `mods/moai-board`, `docs-site`, templates and deployed mirrors.

Direct importers to include in the re-measure (non-test importers found by `git grep` of the import path at `026f48658f`; a `go list` reverse-dependency query was refused by the worktree guard, see `research.md` §R.1):

- of `internal/homestate`: `internal/cli`, `internal/codexwiring`, `internal/decision`, `internal/discovery`, `internal/escalation`, `internal/factory`, `internal/factorymsg`, `internal/hook` and `internal/hook/handoff`, `internal/web`.
- of `internal/factory`: `internal/cli`, `internal/discovery`, `internal/factorymsg`, `internal/graph`, `internal/hook`, `internal/statusline`, `internal/web`, and `cmd/t657-merge`.
- of `internal/template`: `internal/cli`, `internal/core/project`, `internal/hook`, `internal/mirrornotice`, `internal/settings`, `internal/userassets`, `internal/web`.

`internal/cli` is run by name-exact `-run` selections of the tests that name the changed files or identifiers (each milestone above lists its selection); the rest of `internal/cli` is carried to batch CI as an explicit Gap. The landing note states that Gap and the selections run.

## §G Template-First obligations

1. Source of truth: `internal/template/templates/**`. The deployed `.claude/`, `.codex/`, `.agents/`, `.moai/` copies are regenerated by the project build and the update path; none is edited first.
2. Commands render differently in the deployed tree than in the template (the template body carries the dual-harness dispatch block); mirror checks therefore compare rendered output, not raw bytes. The workflow document and the router skill are byte-identical mirrors and are guarded by the existing parity tests.
3. `.agents/skills/moai-todo` and `moai-gtd` carry "Generated by the MoAI command-to-skill publisher; regenerate, do not edit": regenerate, then confirm the drift check is green.
4. After the build, `git status --short` shows only the intended change set; a regenerated embedded catalog is part of that set and is staged by explicit pathspec.
5. Instruction files stay English; docs-site and README keep four locales in step.

## §H Risks

| Risk | Effect | Mitigation |
|------|--------|-----------|
| Substitution touches a persisted name or a recorded fixture | Data or test breakage | Class table with explicit exclusions; excluded files byte-identical in the script test; `TestNoResidualGTDSpelling` and the OQ-3 pin |
| Argument-vector consumers use `'moai', 'gtd'`, which a `moai gtd` text pattern does not match | Alias warning reaches a parser, or the consumer breaks at removal | M7 reviewed edits plus a grep over the argument-vector form |
| M1 removes the no-remote completion path | No-remote projects cannot complete a factory card | OQ-6 gate; M1 independent; completion path named before removal |
| Overlap with t1621 and t1639 on the transition-table files (t1663's overlap is not measured) | Merge conflicts or contradictory edges | Pre-flight item 3; M1 rebases on whichever lands first |
| Lane own-card done widens the lane's authority | A lane completes a card without audit | Evidence is the closed landing form only; refusal matrix; OQ-5 keeps the approval gate in the path |
| AGENTS.md and always-loaded rules grow by one byte per substituted token | Ceiling or budget breach | M9 measurement before and after; byte statement |
| The installed `moai` lags the tree | RED/GREEN observed on the wrong build | Build from the tree and invoke by path; record both commits |

## §I Self-verification deliverables (run phase)

Each item carries the command, the verbatim output, and the tree HEAD of this run: the AC matrix with RED-now and GREEN cells; the change-scope measurement from the absorbed-ref merge-base; the cross-platform build; changed-package coverage; lint; the substitution script's dry-run path set and second-run zero; the always-loaded budget measurement; and the Gaps list (the unrun `internal/cli` remainder, the JavaScript tests, the docs-site build if unavailable).

## §J Anti-patterns for this change

- Hand-editing the 73 files instead of running the script.
- Substituting `gtd` as a bare word (it appears in identifiers, table names, hex-prefixed ids, recorded fixtures, and the GTD methodology name).
- Writing the warning to stdout or printing it for help.
- Adding a second handler under the `gtd` spelling.
- Editing the deployed `.claude/` copy before the template source.
- Treating a green run with an empty selection as a pass (a name-exact selector that matches no test).

## §K Cross-references

`spec.md` (requirements), `acceptance.md` (criteria), `design.md` (mechanisms), `research.md` (measurements and the evidence ledger), `decision-index.md` (Q1 to Q9). Related completed SPECs: `SPEC-TODO-LAND-AUTO-DONE-001`, `SPEC-GTD-CANON-BODY-001`, `SPEC-FACTORY-RECORD-001`, `SPEC-LOCAL-MAIN-FLOW-001`.
