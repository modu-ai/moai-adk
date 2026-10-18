# Plan: SPEC-TODO-UNIFY-001

Tree basis: `026f48658f` (branch `WT-todo-unify`; plan artifacts and the audit report sit on top, the rest of the tree is identical, see `research.md` §R.1). Methodology: TDD (`quality.yaml` `development_mode: tdd`). Priorities and ordering only; no time estimates. Revision 2 (audit iteration 1 FAIL); the decisions of `decision-index.md` Q1 to Q9 are written into `spec.md`.

## §A Decision-reversibility ranking and the dependency graph

Sections and milestones below are ordered by how likely the decision they rest on is to change, highest first. A data-model or interface decision comes before a mechanical step, so review attention lands where a reversal would cost most.

| Rank | Decision | Why it may change | Decision row | Milestone |
|------|----------|-------------------|--------------|-----------|
| 1 | Remove T18 and the T19 reservation; declare the evidence-guarded edge T29 | Removes the human-decided no-remote completion; two draft SPECs touch the same rows | Q6 | M1 |
| 2 | Who may close a card without an operator, and the automatic receipt (three facts, issuer `auto`, revocable) | Changes the approval gate's trust model; leader-confirmed 10-18, conditional | Q4, Q5, Q7 | M2a, M2b |
| 3 | Shape and caller of the post-pick spec_id setter | A new verb on a shipped command | Q8, Q9 | M3a, M3b |
| 4 | The six stage verbs registered on `moai todo` | A user-facing registration | Q1 | M4 |
| 5 | Warning alias behavior and the removal statement | User-visible stderr contract | Q2 | M5 |
| 6 | Slash, skill, and workflow-document wording | Edits of shipped surfaces (paths kept) | Q3 | M6 |
| 7 | Substitution classes and the script | Mechanical; reversible by a revert | Q3 | M7 |
| 8 | Applying the script | Mechanical | none | M8 |
| 9 | Always-loaded text (`AGENTS.md`) and rule statements | Edited last (cache directive 3); size-bounded | none | M9 |

**Dependency graph (replaces the iteration-1 independence claim, which `design.md` §D.5 and §D.6 contradicted).** An arrow means "needs the earlier milestone's code or test to exist".

```
M1 ────┐
       ├──► M2b ───────────────────────────► M9
M2a ───┘
M3a ──► M3b                       (no other milestone depends on it, and it needs none)
M4 ──► M5 ──► M6 ──► M7 ──► M8 ──► M9
```

Reasons, one per arrow: M2b requests the factory edge T29, which M1 declares, and mints the receipt M2a defines; M3b registers a verb on the tree M3a's function backs; M5's warning must cover the verbs M4 registers, or the six would run under `gtd` without it; M6 inverts the slash files before M7's script runs, so the script does not rewrite the wrapper's own notice; M7 supplies the script and allowlist M8 runs; M8 converts the corpus before M9 adds the allowlisted `AGENTS.md` row; M9 states the lane own-card done and the T18/T19 removal that M1 and M2b deliver.

Same-file order (not a logical dependency): `internal/cli/todo.go` is edited by M2b (lane guard), M3b (verb registration), and M4 (verb registration), so the run order is M2b, then M3b, then M4. `internal/cli/factory_card.go` is edited only by M7.

**Run order.** M1, M2a, M2b, M3a, M3b, M4, M5, M6, M7, M8, M9. A milestone that discovers its premise false returns a blocker report; the graph says which later milestones it stops.

## §B Decisions

The nine questions iteration 1 recorded are decided by the lane's decision ladder (`auto-semantics.md` §6); each row of `decision-index.md` carries the evidence, the alternate rejected, and a one-line decision record. Summary, with the requirement each answer lands in:

| Q | Decision | Lands in |
|---|----------|----------|
| Q1 | Register all six stage verbs on `moai todo`; delete none | REQ-TU-004, REQ-TU-006 |
| Q2 | The warning states removal "one release after the release that first carries the warning"; no literal release number | REQ-TU-002 |
| Q3 | Rename the user-facing spelling only; internals, persisted formats, and the workflow document path stay | REQ-TU-007, REQ-TU-012 |
| Q4 | This SPEC delivers the lane own-card path, the automatic receipt inside the two close paths, and edge T29; no observers | REQ-TU-017, REQ-TU-020, Out of Scope |
| Q5 | The receipt is issued automatically only when three machine facts hold, issuer `auto`, revocable; otherwise the leader's manual approval stays (fail-closed) | REQ-TU-017, REQ-TU-025 |
| Q6 | Remove T18 and the T19 reservation now; declare T29 (`merged-local → done`, evidence and receipt guarded) | REQ-TU-018, REQ-TU-019 |
| Q7 | Both records end `done`; "landed" is reachability from the landing ref; factory edge first, queue archive second, with a stated recovery | REQ-TU-020 |
| Q8 | A new verb `moai todo spec <n> <SPEC-ID>` writing both carriers, set-once per carrier | REQ-TU-022, REQ-TU-023 |
| Q9 | Leader only | REQ-TU-023 |

## §C Pre-flight

Run before M1 (read-only; commands are plain, one per call):

1. `git rev-parse --short HEAD` and `git branch --show-current` (expect a descendant of `026f48658f` and `WT-todo-unify`).
2. Re-measure the substitution baseline: `git grep -I -l 'moai gtd' -- . ':!.moai/specs' ':!.moai/reports' ':!CHANGELOG.md'` (expect 73 paths at `026f48658f`; the list is in `research.md` §R.4). Re-measure after absorbing `main`; never re-cite 73 without re-measuring.
3. Confirm the open sibling facts have not moved: `SPEC-GITHUB-FLOW-DEFAULT-001` status, `git_strategy.manual.workflow`, `SPEC-FACTORY-COMPLETION-RECOVERY-001` status (its REQ-FCR-010 would implement T19), `SPEC-FACTORY-RECORD-REPAIR-001` status, and whether card t1639 has started (its `pushed → done` edge and observers).
4. Build the tool under measurement from this tree and invoke it by path; record the judging build's commit next to the tree HEAD (verification-claim-integrity §2.2). The installed `moai` at plan time is `ee981a932-dirty`, a strict ancestor 516 commits behind.
5. Windows build check: `GOOS=windows GOARCH=amd64 go build ./...`.
6. Verification scoping: changed packages and direct importers (§F). No full-suite run; no background load; take the heavy-run slot only for the `internal/cli` name-exact selection, per the lane protocol.
7. Classify the 67 paths M8 will convert (the 73 of E-01 minus the 4 recorded fixtures minus `AGENTS.md` and its template) into template-sourced, template-mirror, and deployed-only with `git ls-files`; the classification is the M8 input (D12).

## §D Constraints

- PRESERVE: `internal/factory/gtd_*.go` and `backlog_gtd_schema.go` content, the `gtd_*` table names, `gtdIDPattern`, the `.moai/gtd/answers` path, the `workflows/gtd.md` path and the strings that name it (five rule files' `paths:` keys, the generated embed line, the recorded fixture), the 71/370 pair counts in `fr_transition_test.go`, the auto-done scan's closed evidence set and three misfire guards, the lane guard's REQ-SD-015 refusal text, `moai factory approve` as the leader's manual issuance path (lane-refused), and the receipt gate's verification (`VerifyBinding`) unchanged.
- No test skips (`t.Skip*`) added; no timeout value raised, in code or in any test invocation.
- Templates are edited in `internal/template/templates/`, never first in `.claude/`.
- The always-loaded files in the conversion set are `AGENTS.md` and `internal/template/templates/AGENTS.md.tmpl` only (the four rule files that carry the spelling have `paths:` keys, measured by `git grep -m1 '^paths:'`, `research.md` §R.10). Their edit waits for M9 (cache-aware directive 3) and carries the byte-size and non-invoking-cost statement in the commit body when it grows the file by more than 1,000 bytes.
- Commits stage by explicit pathspec; every commit message carries `card t1641`.

## §E Milestones

Each milestone: RED first (the new test observed failing for the stated reason, classified `EXPECTED_RED` per the TDD result contract, verbatim output recorded in `progress.md` §E.2), then GREEN (new test plus the regression set), then refactor. Size figures are illustration-level (files / changed lines); a milestone that would exceed roughly 10 files or 400 lines at run time splits at the seam it names before it starts. Every `-run` selector below names tests that exist or that the milestone's RED step creates; the selected count is read before the verdict.

### M1 — Remove T18 and the T19 reservation, declare T29 (Priority High)

- **RED (in this order).** (1) The T29 matrix tests, written first and observed red: in a no-remote fixture a merged-local card with a reachable merge commit and a leader-issued receipt reaches `done`; with the branch reset so the merge is unreachable, or with no receipt, it is refused and the row is byte-identical; in a remote fixture T29 admits only when the merge commit is reachable from the remote-tracking landed ref. (2) The removal tests: no edge guarded by the no-remote completion and no reserved-edge refusal remain; `pushed → ci-green` fails with the undeclared-transition error. (3) The pair-count pin stays green (in the fixture the T29 request lacks a receipt, so the pair stays refused).
- **GREEN.** Remove row T18, the `guardNoRemote` branch, the reserved-edge predicate with its error value and call; add T29 for the pair `merged-local → done` with the evidence-and-receipt guard of `design.md` §D.6; the push gate keeps requesting `done` for a no-remote repository, which T29 now declares.
- **Files (about 6 / 300 lines).** `internal/homestate/card_transition.go`, `card_pr_states.go` (comment), tests `fr_transition_test.go`, a new `fr_landed_edge_test.go`, `leader_approval_test.go`, `f7_gate_owner_admission_test.go`.
- **Packages.** `internal/homestate`; `internal/cli` compiles against it (the push gate caller is unchanged).
- **Name-exact selection.** `internal/homestate`: `-run '^(TestFR_AC005_TransitionTableEdgeCount|TestPushGateTargetByRemote|TestLandedEdgeT29.*|TestReservedEdgeRemoved)$'` (the last two are created at RED; the first two exist). Also the leader-approval and evidence tests the removal touches, listed from `git grep -n '^func Test'` in the files named above at RED time.

### M2a — The automatic receipt and its revocation (Priority High)

- **RED.** (1) The three-fact matrix: of the eight combinations of (merged-local state, committed PASS-family verdict with an ancestor `audited_sha`, valid re-measure record), exactly one issues a receipt, and each of the three two-of-three combinations issues nothing; (2) the issued receipt carries issuer `auto` and the leader role marker, and the unchanged gate verifies it; (3) a receipt for another card or run, or a stale version, is refused by the unchanged gate; (4) revocation: after the leader revokes, the close is refused as receipt-absent and no automatic issuance recurs for that card and run; a lane session is refused the revocation; a leader-issued receipt is unaffected.
- **GREEN.** One issuing function in `internal/homestate` that re-reads the three facts at the moment of the call and writes nothing when any is absent; a revocation record in the receipt store (the schema addition is this milestone's, additive); the leader-only revocation surface on `moai factory approve`.
- **Files (about 6 / 450 lines; splits at the issuance / revocation seam if it exceeds 400 at run time).** `internal/homestate/leader_approval.go` (store addition), a new `leader_approval_auto.go`, `internal/cli/leader_approval_gate.go` (revocation flag), tests in both packages.
- **Packages.** `internal/homestate`, `internal/cli`.
- **Name-exact selection.** `TestAutoReceiptThreeFacts|TestAutoReceiptIssuer|TestAutoReceiptRevocation` (created at RED) plus the existing `TestLeaderApproval.*` family named in `leader_approval_test.go` and `leader_approval_gate_test.go` at RED time.

### M2b — Lane own-card done and the both-records close (Priority High; needs M1 and M2a)

- **RED.** (1) Lane env, card held by that lane, landing evidence of the closed form, the three facts present: `done` is admitted, the receipt is issued with issuer `auto`, the factory row ends `done` (edge T29) and the queue row is archived, in that order; (2) the refusal matrix: another lane's card, no landing evidence, every other non-allowlisted verb, each leaves the queue file byte-identical; (3) one fact absent: nothing is issued, the close is refused with the gate's receipt-absent text, both records are unchanged; (4) a fault injected between the factory edge and the archive leaves the factory row `done` and the queue row open, exits non-zero naming both, and a repeat completes the archive without a second factory edge; (5) the leader's auto-done scan closes a landed `merged-local` card the same way; a `pushed` row with no leader-issued receipt is not closed (nothing auto-issued), and with a leader-issued receipt its queue row closes while the factory row stays `pushed`, named in the log; (6) the twelve-card fixture (ten with landing evidence, two without): exactly ten end `done` in both records; (7) mutant: a guard that admits `done` for any card fails row (2).
- **GREEN.** One admission function in the lane guard; the verb layer asks the existing landing predicate and refuses without it; the close helper shared by the lane path and the scan runs issue, factory edge, archive in that order; the log row carries the new source label and the issuer; `todo undone` pairs its reversal with that label.
- **Files (about 7 / 520 lines; splits at the lane-path / scan-integration seam).** `internal/cli/todo.go` (guard), a new `todo_lane_done.go`, `todo_autodone.go`, `leader_approval_gate.go`, tests `todo_lane_done_test.go` and additions to the existing lane-guard and auto-done tests.
- **Packages.** `internal/cli`; `internal/factory` only if the evidence helper moves.
- **Name-exact selection.** `TestTodoLane.*|TestTodoCloseBothRecords|TestTenLandedCardsClose` (created at RED) plus the existing `TestAutoDone.*` and `TestLeaderApprovalGate.*` families (names read by `git grep -n '^func Test'` in `todo_autodone_test.go` and `leader_approval_gate_test.go` at RED time).

### M3a — The spec_id setter function (Priority High)

- **RED.** (1) Picked and assigned fixtures with empty spec_id: the function fills both carriers, one `card.fields` event; (2) refusal matrix: a different id in either carrier, `leased` and later states, terminal states, malformed id: neither carrier changes; (3) identical values in both carriers: no-op success; (4) a fault injected after the queue write and before the factory commit leaves the queue carrier set and the factory carrier empty, and a repeat with the same id completes the pair; a fault in the queue write leaves both carriers unchanged.
- **GREEN.** A set-once-per-carrier function in `internal/homestate` whose factory write transaction runs the queue write in its pre-commit step (the seam transitions already use for the dispatch record, `design.md` §D.7).
- **Files (about 4 / 230 lines).** `internal/homestate/card_picked.go` (or a new `card_spec_set.go`), a small helper in `internal/factory`, tests in both packages.
- **Packages.** `internal/homestate`, `internal/factory`.
- **Name-exact selection.** `TestCardSpecSetOnce|TestCardSpecSetFailureBetweenCarriers` (created at RED).

### M3b — The setter verb (Priority High; needs M3a)

- **RED.** The verb writes through M3a's function for the leader; refuses lane sessions with the standing refusal text; malformed ids are refused before any write; `next --spec`, `factory assign --spec`, and the unpick clear keep their present tests green (names in `acceptance.md` AC-TU-022); kickoff's SPEC lookup passes after the setter.
- **GREEN.** Register `spec` on the `todo` tree; the lane guard keeps it refused by not listing it.
- **Files (about 3 / 200 lines).** a new `internal/cli/todo_spec.go`, **`internal/cli/todo.go` (the single registration of the verb tree)**, tests `todo_spec_test.go`.
- **Packages.** `internal/cli`.
- **Name-exact selection.** `TestTodoSpecSetter|TestTodoSpecSetterLaneRefused|TestKickoffAfterSpecSetter` (created at RED) plus the six existing tests named in AC-TU-022.

### M4 — The six stage verbs on `moai todo` (Priority Medium)

- **Step 0 (already measured at plan time).** `reflect --rebuild-projection` writes private projection files (`research.md` §R.12), so `reflect` is admitted for lanes only without that flag; the other five write and stay refused.
- **RED.** (1) `TestGuidanceVerbsResolve`: every `moai todo <verb>` and `moai gtd <verb>` token in the guidance corpus resolves in its command tree (observed red: six lines of the workflow skill name verbs `moai todo` lacks; the token and file floors are in AC-TU-005); (2) the verb-surface parity test rewritten to "the `gtd` verb set equals the `todo` verb set, six stage verbs included"; (3) the five-stage path of the existing `TestGTDFiveStageCLIUsesSameSQLite` run again through the `todo` spelling with an equal store; (4) the lane classification test of REQ-TU-006.
- **GREEN.** Register the six constructors on the `todo` tree, delete the `gtd`-only registration line, classify each for the lane guard; the mistyped-verb guard reads the tree, so no separate list changes.
- **Files (about 6 / 260 lines).** `internal/cli/todo.go`, `gtd.go`, `gtd_answer.go`, tests `gtd_compat_test.go`, a new `todo_guidance_resolves_test.go`, `todo_surface_test.go`.
- **Packages.** `internal/cli`.
- **Name-exact selection.** `TestGTDAllTodoVerbsParity|TestGTDFiveStageCLIUsesSameSQLite|TestGTDCLIInputRefusalsDoNotFallThrough|TestStageVerbsOnTodoTree|TestGuidanceVerbsResolve|TestLaneClassifiesStageVerbs` (the first three exist at `gtd_compat_test.go:49`, `:124`, `:164`; the rest are created at RED).

### M5 — The warning alias (Priority Medium; needs M4)

- **RED.** `TestGTDAliasWarnsThenDelegates` over every verb (warning once on stderr, first line `Warning: moai gtd has been renamed to moai todo`, naming the `moai todo` form; stdout and exit equal to the `todo` form); `TestGTDAliasWarningStreams` (`--json` stdout parses, help requests carry no warning, and argument- or flag-refused invocations print the `todo` form's error with no warning); the old parity assertion `gtdErr != todoErr` (`gtd_compat_test.go:83`) is rewritten to "stderr equals the warning plus the `todo` stderr".
- **GREEN.** `gtd` builds from the todo tree and chains one pre-run step in front of the lane guard that writes the warning once, reusing the existing deprecation convention (stderr, `Warning:` prefix; `warnDeprecatedAgentModelFlags` is the precedent).
- **Files (about 4 / 200 lines).** `internal/cli/gtd.go`, a small warning helper, tests `gtd_compat_test.go`, a new `gtd_alias_warn_test.go`.
- **Packages.** `internal/cli`.

### M6 — Slash command, skills, and the workflow document wording (Priority Medium; needs M5)

- **RED.** `TestTodoCanonicalSurfaceGolden` (inverts `TestGTDCanonicalSurfaceGolden`): `/moai:todo` dispatches `todo`; `/moai:gtd` dispatches `todo` and carries the rename notice; published skills equal the publisher output (count stays 17); the router skill routes `todo` canonical and `gtd` alias; the queue workflow document's header names `moai todo` canonical and its path is unchanged; the existing doc tests keep passing against the unchanged path.
- **GREEN.** Edit template sources (the two command files, the router skill rows, the workflow document header and body wording), regenerate the published skills with the project build.
- **Files (about 9 / 110 lines).** template commands (2) and their deployed copies, template and deployed router skill, template and deployed workflow document, `.agents/skills/moai-todo` and `moai-gtd` (generated), `internal/template/gtd_canonical_surface_test.go`.
- **Packages.** `internal/template`, `internal/cli`.
- **Name-exact selection.** `internal/template`: `TestTodoCanonicalSurfaceGolden`; `internal/cli`: `TestTodoSkillDocumentsHistoryVerb|TestTodoSkillDocumentsJevSource` (they exist at `todo_skill_doc_test.go:17` and `todo_skill_doc_parity_test.go:19`).

### M7 — The substitution script and the executable consumers (Priority Low; mechanical; needs M6)

- **RED.** (1) A fixture test of the script against a temp tree holding one file per substitution class and one per excluded class: expected output exact, excluded files byte-identical; (2) a second run changes nothing; (3) `TestNoResidualGTDSpelling` with a planted-mutant fixture proving the check can go red; (4) the consumer edits: argument vectors in the board mod, the cutover script default, the push-gate help line in `factory_card.go`, the `moai todo` help text.
- **GREEN.** Land `scripts/gtd-to-todo-substitute.sh` (`design.md` §D.3 holds the class table and the dry-run contract), its test, and the reviewed consumer edits.
- **Files (about 6 / 380 lines).** the script, its test, `mods/moai-board/hooks/data.ts`, `mods/moai-board/tests/support.ts`, `scripts/cutover-precheck.sh`, `internal/cli/factory_card.go`.
- **Gap carried.** The board mod's JavaScript tests are not run by any Go selection; the run phase records the runner used, or lists them as an unverified Gap.

### M8 — Apply the script (Priority Low; mechanical; needs M7)

- **Unit and size.** The unit is changed lines, not authored code: about 244 lines across 67 files (the 73 of E-01, minus 4 recorded fixtures, minus the two always-loaded `AGENTS.md` files), measured against E-01 and re-measured at pre-flight. The 10-file split rule does not apply; the review is dry-run path-set equality.
- **Batches by origin class (from pre-flight item 7).** M8a: template sources under `internal/template/templates/` and the `internal/cli` and `internal/template` test strings that are prose; M8b: deployed copies that have a template source follow the regenerated output, and the deployed-only copies (`.moai/workflows/gtd.md` has no template source, `git ls-files` lists only `README.md` on the template side) are converted by the script directly and named in the batch record; M8c: `docs-site/content` in four locales and the four READMEs (the `moai-gtd` pages stay; no redirect or menu change).
- **Mechanism.** Dry-run first; the printed path set must equal the expected set (pre-flight list minus excluded classes); then apply; then re-run, expecting zero replacements and a green `TestNoResidualGTDSpelling`.
- **Docs-site obligations.** Four locales change in the same commit batch; the locale-parity baseline is not extended; the docs-site verification recipe (warning-free build, parity, URL scan) runs before the batch is accepted.

### M9 — AGENTS.md, rule statements, and the closing guards (Priority Low; last; needs M8, M2b, M1)

- **Content.** The `AGENTS.md` §11 row and its template mirror (`moai todo` canonical, `moai gtd` a warning alias, no "fully supported"); the hand-authored statements in the factory-dispatch rule family (the lane own-card `done` exception, the automatic receipt with issuer `auto`, the removal of T18 and the T19 reservation, edge T29, the leader's remaining role) in the template source and the deployed copy and the `.moai/policies` projection.
- **Guards.** `TestAgentsMDTodoRowWording` (live file and template mirror) and `TestCodexContractByteCeiling` stay green (AGENTS.md is 23,780 bytes against 24,576); the always-loaded growth statement is written for any edit above 1,000 bytes; the rule-loading budget measurement (count and bytes of always-loaded files before and after) is recorded.
- **RED.** The wording test fails on the current row (E-16); the residual-ledger test lists the `AGENTS.md` row as an allowed alias statement and fails before the edit.

## §F Affected packages and landing re-measure scope

Changed packages: `internal/cli`, `internal/homestate`, `internal/factory`, `internal/template`; `internal/config` through the AGENTS.md ceiling guard. Non-Go surfaces: `scripts/`, `mods/moai-board`, `docs-site`, templates and deployed mirrors.

Direct importers to include in the re-measure (non-test importers found by `git grep` of the import path at `026f48658f`; a `go list` reverse-dependency query was refused by the worktree guard, see `research.md` §R.1):

- of `internal/homestate`: `internal/cli`, `internal/codexwiring`, `internal/decision`, `internal/discovery`, `internal/escalation`, `internal/factory`, `internal/factorymsg`, `internal/hook` and `internal/hook/handoff`, `internal/web`.
- of `internal/factory`: `internal/cli`, `internal/discovery`, `internal/factorymsg`, `internal/graph`, `internal/hook`, `internal/statusline`, `internal/web`, and `cmd/t657-merge`.
- of `internal/template`: `internal/cli`, `internal/core/project`, `internal/hook`, `internal/mirrornotice`, `internal/settings`, `internal/userassets`, `internal/web`.

`internal/cli` is run by name-exact `-run` selections of the tests that name the changed files or identifiers (each milestone above lists its selection); the rest of `internal/cli` is carried to batch CI as an explicit Gap. The landing note states that Gap and the selections run.

## §G Template-First obligations

1. Source of truth: `internal/template/templates/**`. A deployed `.claude/`, `.codex/`, `.agents/`, or `.moai/` copy that has a template source follows it; none is edited first. A deployed-only file (`.moai/workflows/gtd.md`) has no template source and is edited directly by the script, named as such in the M8b batch record.
2. Commands render differently in the deployed tree than in the template (the template body carries the dual-harness dispatch block); mirror checks therefore compare rendered output, not raw bytes. The router skill is a byte-identical mirror guarded by the existing parity tests; the template and deployed `.claude/skills/moai/workflows/gtd.md` were byte-identical at `026f48658f` (`diff -q` printed nothing, `research.md` §R.10).
3. `.agents/skills/moai-todo` and `moai-gtd` carry "Generated by the MoAI command-to-skill publisher; regenerate, do not edit": regenerate, then confirm the drift check is green.
4. After the build, `git status --short` shows only the intended change set; a regenerated embedded catalog is part of that set and is staged by explicit pathspec.
5. Instruction files stay English; docs-site and README keep four locales in step.

## §H Risks

| Risk | Effect | Mitigation |
|------|--------|-----------|
| Substitution touches a persisted name or a recorded fixture | Data or test breakage | Class table with explicit exclusions; excluded files byte-identical in the script test; `TestNoResidualGTDSpelling` and the REQ-TU-012 pin |
| Argument-vector consumers use `'moai', 'gtd'`, which a `moai gtd` text pattern does not match | Alias warning reaches a parser, or the consumer breaks at removal | M7 reviewed edits plus a grep over the argument-vector form |
| The automatic receipt moves the issuance of a leader-role receipt into a path a lane can trigger | A card completes without the leader's read of the evidence | Three machine facts, all required, re-read at issuance; fail-closed; issuer `auto`; leader revocation; the gate's verification unchanged; leader-confirmed 10-18, conditional |
| T29 replaces T18's pair with a different guard | A reviewer reads it as a renamed row | AC-TU-018 asserts the behavior that distinguishes them (admits with a remote when the merge is on the landed ref; refuses on unreachable merge; no human-decider requirement) |
| Overlap with t1621 and t1639 on the transition-table files and the `pushed → done` edge (t1663's overlap is not measured) | Merge conflicts or contradictory edges | Pre-flight item 3; M1 rebases on whichever lands first; this SPEC declares no `pushed → done` edge |
| The receipt store gains a revocation record | A schema addition in a store other SPECs read | Additive and read-compatible; `TestAutoReceiptRevocation` plus the existing receipt tests |
| The script test needs `sh` on every CI runner (not measured) | A Windows runner fails the test, or tempts a skip | The test fails loudly if `sh` is absent (no skip, AC-TU-024); M7 RED decides whether the script becomes a Go helper |
| AGENTS.md grows by a few bytes | Ceiling or budget breach | M9 measurement before and after; byte statement |
| The installed `moai` lags the tree | RED/GREEN observed on the wrong build | Build from the tree and invoke by path; record both commits |

## §I Self-verification deliverables (run phase)

Each item carries the command, the verbatim output, and the tree HEAD of this run: the AC matrix with RED-now and GREEN cells; the change-scope measurement from the absorbed-ref merge-base; the cross-platform build; changed-package coverage; lint; the substitution script's dry-run path set and second-run zero; the always-loaded budget measurement; and the Gaps list (the unrun `internal/cli` remainder, the JavaScript tests, the docs-site build if unavailable).

## §J Anti-patterns for this change

- Hand-editing the 67 files instead of running the script.
- Substituting `gtd` as a bare word (it appears in identifiers, table names, hex-prefixed ids, recorded fixtures, and the GTD methodology name).
- Writing the warning to stdout or printing it for help or for an argument-refused invocation.
- Adding a second handler under the `gtd` spelling.
- Renaming `workflows/gtd.md` (five rule files' `paths:` keys, a generated embed line, and a recorded fixture name it).
- Issuing the automatic receipt on two of the three facts, or when the verdict file is uncommitted.
- Editing the deployed `.claude/` copy before the template source.
- Treating a green run with an empty selection as a pass (a name-exact selector that matches no test).

## §K Cross-references

`spec.md` (requirements), `acceptance.md` (criteria), `design.md` (mechanisms), `research.md` (measurements and the evidence ledger), `decision-index.md` (Q1 to Q9). Related completed SPECs: `SPEC-TODO-LAND-AUTO-DONE-001`, `SPEC-GTD-CANON-BODY-001`, `SPEC-FACTORY-RECORD-001`, `SPEC-LOCAL-MAIN-FLOW-001`.
