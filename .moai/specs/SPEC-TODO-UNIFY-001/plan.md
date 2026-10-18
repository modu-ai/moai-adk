# Plan: SPEC-TODO-UNIFY-001

Tree basis: `026f48658f` (branch `WT-todo-unify`; plan artifacts and the audit report sit on top, the rest of the tree is identical, see `research.md` §R.1). Methodology: TDD (`quality.yaml` `development_mode: tdd`). Priorities and ordering only; no time estimates. Revision 3 (audit iteration 2 FAIL 0.80; revision 2 answered iteration 1 FAIL 0.62); the decisions of `decision-index.md` Q1 to Q9 are written into `spec.md`. Revision 4 (post-ceiling correction under standing rule C; audit iteration 3 FAIL 0.86, D29 to D31) changed M1 and M2a and added the known debt KD-1 to §H. **Revision 5 (shrunk scope, operator ruling of 10-19, board `d-20261018T214213Z-94a7`, decision-index.md Q10):** the M1 → M2a → M2b chain (automatic done receipt, T18/T19 removal with T29, lane own-card done) moves to the follow-up card t1878 and its milestones are withdrawn from this plan; the CLI-alias policy is reversed — no warning alias, no one-release grace — so `moai gtd` is deleted immediately and every residual `moai gtd` / `/moai:gtd` / `moai:gtd` / `gtd.md` reference migrates to `moai todo` in this card (completion judgment: the residual grep prints nothing; the queue workflow document is renamed to `workflows/todo.md` as a consequence); M5's scope becomes the deletion plus caller migration; M9 loses the T18/T19 and lane-own-card wording; KD-1 moves to t1878 with the receipt machinery it describes. The run order is M3a, M3b, M4, M5, M6, M7, M8, M9. One plan delta audit on the shrunk scope follows this commit (the rule C counter restarts by the operator's act). **Revision 6 (repair for the proxy codex leg's blocking findings P1-P2, leader ruling of 10-19, receipt `rcpt-54ca7c42603570ff69d6589e`, decision-index.md Q11):** the rule sources' `paths:` key updates move from M8a into M6, so the workflow-document rename, its consumer edits, and the rule-path updates land in one milestone; M6's two document tests' post-milestone form reads `todo.md`; and M8 closes with script idempotence plus the recorded residual RED — the full-residual GREEN is M9's completion condition (AC-TU-010). One delta audit with codex required follows this commit (the repair round is round 2 of the counter the operator's act `d-20261018T214213Z-94a7` restarted; a third FAIL holds the card).

## §A Decision-reversibility ranking and the dependency graph

Sections and milestones below are ordered by how likely the decision they rest on is to change, highest first. A data-model or interface decision comes before a mechanical step, so review attention lands where a reversal would cost most.

| Rank | Decision | Why it may change | Decision row | Milestone |
|------|----------|-------------------|--------------|-----------|
| 1 | Immediate removal of the `moai gtd` command tree; zero residual references (the operator's reversal of the Q2 grace design) | A user-visible command removal; the completion judgment is the operator's zero-residual grep | Q10 (supersedes Q2) | M5 |
| 2 | Shape and caller of the post-pick spec_id setter | A new verb on a shipped command | Q8, Q9 | M3a, M3b |
| 3 | The six stage verbs registered on `moai todo` | A user-facing registration | Q1 | M4 |
| 4 | Slash, skill, and workflow-document wording (the document renamed to `workflows/todo.md`) | Edits of shipped surfaces; the rename is the zero-residual consequence | Q3 (narrowed by Q10) | M6 |
| 5 | Substitution classes and the script | Mechanical; reversible by a revert | Q3 | M7 |
| 6 | Applying the script | Mechanical | none | M8 |
| 7 | Always-loaded text (`AGENTS.md`) and rule statements | Edited last (cache directive 3); size-bounded | none | M9 |
| — | The done-path cluster (automatic receipt, T18/T19 removal with T29, lane own-card done) | Moved whole to the follow-up card t1878 by the operator's ruling; not planned here | Q4, Q5, Q6, Q7 (moved) | (t1878) |

**Dependency graph.** An arrow means "needs the earlier milestone's code or test to exist".

```
M3a ──► M3b                       (no other milestone depends on it, and it needs none)
M4 ──► M5 ──► M6 ──► M7 ──► M8 ──► M9
```

Reasons, one per arrow: M3b registers a verb on the tree M3a's function backs; M5's deletion must come after M4 has registered the six stage verbs on the `todo` tree, or the verbs would be deleted with the `gtd` tree that still holds them; M6 inverts the slash files before M7's script runs, so the script does not rewrite a wrapper notice (there is none after revision 5, and M6 keeps it that way while deleting the `gtd` command file); M7 supplies the script and class table M8 runs; M8 converts the corpus before M9 makes the final always-loaded `AGENTS.md` edit that turns the residual grep green.

Same-file order (not a logical dependency): `internal/cli/todo.go` is edited by M3b (verb registration) and M4 (verb registration), so the run order is M3b, then M4. `internal/cli/gtd.go` and its tests are deleted in M5. `internal/cli/factory_card.go` is edited only by M7.

**Run order.** M3a, M3b, M4, M5, M6, M7, M8, M9. A milestone that discovers its premise false returns a blocker report; the graph says which later milestones it stops.

## §B Decisions

The nine questions iteration 1 recorded are decided by the lane's decision ladder (`auto-semantics.md` §6); each row of `decision-index.md` carries the evidence, the alternate rejected, and a one-line decision record. Summary, with the requirement each answer lands in:

| Q | Decision | Lands in |
|---|----------|----------|
| Q1 | Register all six stage verbs on `moai todo`; delete none | REQ-TU-004, REQ-TU-006 |
| Q2 | ~~The warning states removal "one release after the release that first carries the warning"~~ **Superseded (revision 5, Q10):** the operator reversed the grace design — `moai gtd` is removed immediately, no warning exists | (the warning requirements REQ-TU-002/003 are retired) |
| Q3 | Rename the user-facing spelling only; internals and persisted formats stay. **Narrowed by Q10 (revision 5):** the queue workflow document is renamed to `workflows/todo.md` — its path is no longer preserved; every other Q3 preservation stands | REQ-TU-007, REQ-TU-012 |
| Q4 | **Moved to t1878 (revision 5):** the lane own-card path, the receipt inside the two close paths, and edge T29 leave this SPEC; the no-observer clause stands for both cards | (t1878) |
| Q5 | **Moved to t1878 (revision 5):** the three machine facts, issuer `auto`, revocation, fail-closed — shape unchanged, implemented by the follow-up card | (t1878) |
| Q6 | **Moved to t1878 (revision 5):** the T18/T19 removal and the T29 declaration leave this SPEC | (t1878) |
| Q7 | **Moved to t1878 (revision 5):** the both-records done close leaves this SPEC | (t1878) |
| Q8 | A new verb `moai todo spec <n> <SPEC-ID>` writing both carriers, set-once per carrier; first fill from `picked` through `kickoff`, a completing repeat in any non-terminal state (revision 3) | REQ-TU-022, REQ-TU-023 |
| Q9 | Leader only | REQ-TU-023 |
| Q10 | **Operator ruling 10-19** (AskUserQuestion, board `d-20261018T214213Z-94a7`): accept the recorded shrink — this card keeps the rename only (M3a/M3b, M4-M8, a trimmed M9); the done-path cluster moves whole to t1878; `moai gtd` is removed immediately with no warning alias and no one-release grace; every residual reference (`moai gtd`, `/moai:gtd`, `moai:gtd`, `gtd.md`) migrates to `moai todo` in this card, completion judged by the residual grep reaching 0; one plan delta audit on the shrunk scope, the rule C counter restarting by this operator act | REQ-TU-001, REQ-TU-007, REQ-TU-011, REQ-TU-014; the §5 moved note |

## §C Pre-flight

Run before M3a (read-only; commands are plain, one per call):

1. `git rev-parse --short HEAD` and `git branch --show-current` (expect a descendant of `026f48658f` and `WT-todo-unify`).
2. Re-measure the substitution baseline twice, with both commands: the path-level baseline `git grep -I -l 'moai gtd' -- . ':!.moai/specs' ':!.moai/reports' ':!CHANGELOG.md'` (73 paths at `026f48658f`; the list is in `research.md` §R.4) and the completion-judgment baseline `git grep -E 'moai gtd|/moai:gtd|moai:gtd|gtd\.md' -- . ':!.moai/specs' ':!.moai/reports' ':!CHANGELOG.md' ':!*.jsonl'` (397 lines in 101 files at `3c1809e2f1`; the operator measured 387 on 10-19 — the run phase records its own measurement and never re-cites either figure without re-measuring). Re-measure after absorbing `main`.
3. Confirm the recorded fixtures' consumers are known before M8 migrates their text: the tests reading `internal/template/testdata/binding_ledger.json`, the tests reading the `internal/cli/testdata/codex-rollouts-t1171/roles*/manager-lead.toml` role files, and `TestWorkflowRulePathsPinned` (the `workflows/gtd.md` path pin that M6 rewrites to the new path).
4. Build the tool under measurement from this tree and invoke it by path; record the judging build's commit next to the tree HEAD (verification-claim-integrity §2.2). The installed `moai` at plan time is `ee981a932-dirty`, a strict ancestor 516 commits behind.
5. Windows build check: `GOOS=windows GOARCH=amd64 go build ./...`.
6. Verification scoping: changed packages and direct importers (§F). No full-suite run; no background load; take the heavy-run slot only for the `internal/cli` name-exact selection, per the lane protocol.
7. Classify the residual-grep file set (the 101 files of item 2's second command, minus the JSONL recordings the exclusion already drops, minus `AGENTS.md` and its template, which M9 edits by hand, and minus `internal/factory/gtd_engage.go` — REQ-TU-011's named exception, a preserved Go source no milestone edits and the script never touches) into template-sourced, template-mirror, deployed-only, Go test prose, and recorded-fixture classes with `git ls-files`; the classification is the M8 input.

## §D Constraints

- PRESERVE: `internal/factory/gtd_*.go` and `backlog_gtd_schema.go` content, the `gtd_*` table names, `gtdIDPattern` (wherever M5 leaves its defining file), the `.moai/gtd/answers` path, and the lane guard's REQ-SD-015 refusal text. The recorded `.jsonl` rollout under `internal/cli/testdata/codex-rollouts-t1171/real/` stays byte-identical (history, excluded by the completion grep's `*.jsonl` clause); the role `.toml` files beside it migrate with the corpus and the tests that read them are updated in the same milestone.
- No test skips (`t.Skip*`) added; no timeout value raised, in code or in any test invocation.
- Templates are edited in `internal/template/templates/`, never first in `.claude/`.
- The always-loaded files in the conversion set are `AGENTS.md` and `internal/template/templates/AGENTS.md.tmpl` only (the rule files that carry the spelling have `paths:` keys, measured by `git grep -m1 '^paths:'`, `research.md` §R.10). Their edit waits for M9 (cache-aware directive 3) and carries the byte-size and non-invoking-cost statement in the commit body when it changes the file by more than 1,000 bytes; the residual grep turns green only after M9's edit.
- Commits stage by explicit pathspec; every commit message carries `card t1641`.

## §E Milestones

Each milestone: RED first (the new test observed failing for the stated reason, classified `EXPECTED_RED` per the TDD result contract, verbatim output recorded in `progress.md` §E.2), then GREEN (new test plus the regression set), then refactor. Size figures are illustration-level (files / changed lines); a milestone that would exceed roughly 10 files or 400 lines at run time splits at the seam it names before it starts. Every `-run` selector below names tests that exist or that the milestone's RED step creates; the selected count is read before the verdict.

### M1, M2a, M2b — withdrawn from this card (moved to t1878)

The three done-path milestones of revisions 1 to 4 (M1 remove T18/T19 and declare T29; M2a the automatic receipt and its revocation; M2b lane own-card done and the both-records close) are no longer planned here: the operator's ruling of 10-19 (board `d-20261018T214213Z-94a7`, decision-index.md Q10) moves the whole M1 → M2a → M2b chain to the follow-up card t1878, which authors its own requirements and criteria (the retired identifiers REQ-TU-015 to REQ-TU-021 and REQ-TU-025 are not reused). Their plan text, the audit findings it closed (D29 to D31), the iteration-4 hardening findings D38 to D40, and the known debt KD-1 travel to that card's SPEC. This card touches no transition-table, receipt-store, or close-path code.

### M3a — The spec_id setter function (Priority High)

- **RED.** (1) The state table (`TestCardSpecSetStateSet`): fixtures in each of `picked`, `assigned`, `leased`, `plan`, `plan-audit`, `kickoff` with empty carriers: the function fills both carriers, one `card.fields` event, version plus one; fixtures in `run`, `sync`, `sync-audit`, `merge-ready`, `merging`, `merged-local`, `pushed`, and the terminal states: refused, neither carrier changes; (2) refusal matrix: a different id in either carrier, malformed id: neither carrier changes; (3) identical values in both carriers: no-op success in the six states and in `run`, refused in `done`; (4) a fault injected after the queue write and before the factory commit leaves the queue carrier set and the factory carrier empty, and a repeat with the same id completes the pair; a fault in the queue write leaves both carriers unchanged; (5) the completing repeat after the card moved on (`TestTodoSpecSetterRepeatAfterLease` at the verb level, the function-level rows here): to `leased` and `plan` (first-fill states), to `run` (outside the first-fill set, completing repeat admitted, different id refused), and to a terminal state (refused).
- **GREEN.** A set-once-per-carrier function in `internal/homestate` whose factory write transaction runs the queue write in its pre-commit step (the seam transitions already use for the dispatch record, `design.md` §D.7); its state rule is the first-fill set (`picked` through `kickoff`) for a first fill and any non-terminal state for a completing repeat.
- **Files (about 4 / 260 lines).** `internal/homestate/card_picked.go` (or a new `card_spec_set.go`), a small helper in `internal/factory`, tests in both packages.
- **Packages.** `internal/homestate`, `internal/factory`.
- **Name-exact selection.** `TestCardSpecSetOnce|TestCardSpecSetStateSet|TestCardSpecSetFailureBetweenCarriers` (created at RED).

### M3b — The setter verb (Priority High; needs M3a)

- **RED.** The verb writes through M3a's function for the leader; refuses lane sessions with the standing refusal text; malformed ids are refused before any write; `next --spec`, `factory assign --spec`, and the unpick clear keep their present tests green (names in `acceptance.md` AC-TU-022); kickoff's SPEC lookup passes after the setter, including the wedge fixture (a card leased with empty carriers, SPEC authored afterwards, setter while `leased`, then `plan-audit` admission and the audit kickoff resolve the directory; the control without the setter is refused by both); the lease-between-attempts rows at the verb level (`TestTodoSpecSetterRepeatAfterLease`).
- **GREEN.** Register `spec` on the `todo` tree (the 27th verb: the todo tree reads 27 until M4 registers the six stage verbs, then 33); the lane guard keeps it refused by not listing it.
- **Files (about 3 / 200 lines).** a new `internal/cli/todo_spec.go`, **`internal/cli/todo.go` (the single registration of the verb tree)**, tests `todo_spec_test.go`.
- **Packages.** `internal/cli`.
- **Name-exact selection.** `TestTodoSpecSetter|TestTodoSpecSetterLaneRefused|TestTodoSpecSetterPartialRepeat|TestTodoSpecSetterRepeatAfterLease|TestKickoffAfterSpecSetter` (created at RED) plus the six existing tests named in AC-TU-022.

### M4 — The six stage verbs on `moai todo` (Priority Medium)

- **Step 0 (already measured at plan time).** `reflect --rebuild-projection` writes private projection files (`research.md` §R.12), so `reflect` is admitted for lanes only without that flag; the other five write and stay refused.
- **RED.** (1) `TestGuidanceVerbsResolve`: every `moai todo <verb>` token in the guidance corpus resolves in the `todo` command tree (observed red: six lines of the workflow skill name verbs `moai todo` lacks; the token and file floors are in AC-TU-005); (2) the verb-surface parity test (`gtd_compat_test.go`, "the `gtd` verb set equals the `todo` verb set, six stage verbs included") stays green until M5 deletes it with the `gtd` tree, and `TestStageVerbsOnTodoTree` pins the literal 33 names on the `todo` tree (the 26 present, `spec` from M3b, the six; the baseline 26 and 32 was measured at HEAD `3dc1052f6f`, `design.md` §D.4); (3) the five-stage path of the existing `TestGTDFiveStageCLIUsesSameSQLite` run again through the `todo` spelling with an equal store; (4) the lane classification test of REQ-TU-006.
- **GREEN.** Register the six constructors on the `todo` tree, delete the `gtd`-only registration line, classify each for the lane guard; the mistyped-verb guard reads the tree, so no separate list changes.
- **Files (about 6 / 260 lines).** `internal/cli/todo.go`, `gtd.go`, `gtd_answer.go`, tests `gtd_compat_test.go`, a new `todo_guidance_resolves_test.go`, `todo_surface_test.go`.
- **Packages.** `internal/cli`.
- **Name-exact selection.** `TestGTDAllTodoVerbsParity|TestGTDFiveStageCLIUsesSameSQLite|TestGTDCLIInputRefusalsDoNotFallThrough|TestStageVerbsOnTodoTree|TestGuidanceVerbsResolve|TestLaneClassifiesStageVerbs` (the first three exist at `gtd_compat_test.go:49`, `:124`, `:164`; the rest are created at RED).

### M5 — Delete the `moai gtd` command tree and migrate its callers (Priority High; needs M4; operator revision 5, Q10)

- **RED (in this order).** (1) `TestGTDCmdUnknown`: built from the tree, `moai gtd` (bare), every verb the `gtd` tree carries today, and a mistyped verb each exit non-zero with the unknown-command error for `moai gtd` (AC-TU-003) — observed red first, because the tree still exists; (2) the removal greps: `git grep -n 'NewGTDCommand' -- internal/cli` and `git grep -n -E "'gtd'|moai gtd" -- mods/moai-board/hooks/data.ts scripts/cutover-precheck.sh internal/cli/factory_card.go` return the callers that must not survive (AC-TU-008's pattern), observed non-empty before the edits; (3) the surviving-tree tests that pinned the alias (the `gtd_compat_test.go` family: tree parity, five-stage CLI, input refusals, and the `gtd_answer_test.go` items that move with their constructors) are deleted in the same milestone as the tree they test, so no test dangles on a removed symbol; (4) the help surfaces: no `moai gtd` token remains in any `Long` help text of the `todo` tree (AC-TU-005's corpus).
- **GREEN.** Delete the `gtd` command constructor and its registration so cobra answers `moai gtd` with the unknown-command error (exit non-zero); move the six stage-verb constructors' registration fully onto the `todo` tree (M4 did the registration — M5 deletes the `gtd`-side lines and the `gtd.go` command file, relocating what the persisted-format rule still needs — `gtdIDPattern` and any helper the `todo` tree or `goal.go` calls — to the file that owns them now); migrate the executable callers by reviewed edits: the board mod's argument vectors and label, `scripts/cutover-precheck.sh`'s default command, the push-gate refusal text in `internal/cli/factory_card.go`, and every in-tree help or error message still naming `moai gtd`; delete the `gtd`-tree test files with their tree. The answer library (`gtd_answer.go`) and its `.moai/gtd/answers` directory are preserved behavior (REQ-TU-012) — their verbs are reachable as `moai todo answer` from M4, and their file names and identifiers stay.
- **Files (about 10 / 400 lines; if the caller migration exceeds the budget at run time, the caller edits move to an M5b that precedes M6).** `internal/cli/gtd.go` (deleted, contents relocated or removed), `gtd_answer.go` (kept, registration moved), `todo.go` (constructor wiring), `factory_card.go` (refusal text), `mods/moai-board/hooks/data.ts`, `mods/moai-board/tests/support.ts`, `scripts/cutover-precheck.sh`, the deleted `gtd_compat_test.go`/`gtd_alias_warn_test.go`/`gtd_answer_test.go`-as-moved, a new `gtd_cmd_unknown_test.go`.
- **Packages.** `internal/cli`.
- **Name-exact selection.** `TestGTDCmdUnknown|TestStageVerbsOnTodoTree|TestGTDFiveStageCLIUsesSameSQLite` (the first created at RED; the second survives M5 and re-runs green; the third is deleted with its file only if its fixture moves — the five-stage path re-homes on the `todo` spelling with the same store assertions, its new name recorded at RED time).

### M6 — Slash command, skills, and the workflow document rename, wording, and rule paths (Priority Medium; needs M5)

- **RED.** `TestTodoCanonicalSurfaceGolden` (inverts `TestGTDCanonicalSurfaceGolden`): `/moai:todo` dispatches `todo`; no `/moai:gtd` entry exists anywhere (the command file is gone, M5); the published skills equal the regenerated publisher output and the set no longer carries `moai-gtd`; the router skill routes `todo` as the only queue spelling; the queue workflow document exists at its new path `workflows/todo.md` (template copy, deployed copy, and the deployed-only `.moai/workflows/` copy) and names `moai todo` canonical in header and body; `TestWorkflowRulePathsPinned` is rewritten first to demand the new path, observed red while the rule sources still carry the old one (AC-TU-007); and the two document tests of the name-exact selection below are rewritten first to read `workflows/todo.md`, observed red while the document stands at its old path (revision 6: both build their fixture paths from `workflows/gtd.md` today — `todo_skill_doc_test.go:22-23`, `todo_skill_doc_parity_test.go:24-25`).
- **GREEN.** Rename the workflow document files (`git mv` in the template source and the deployed copies, the deployed-only copy included); edit the template sources (the todo command file, the router skill rows, the workflow document header and body wording); update the five rule sources' `paths:` keys that reference the renamed document — measured at HEAD, the `paths:` key at line 3 of `cross-session-messaging.md`, `factory-dispatch.md`, `factory-dispatch-cards.md`, `factory-dispatch-detail.md`, and `factory-dispatch-gates.md`, each in its template source and its deployed mirror, edited as a glob-shape-preserving `workflows/gtd.md` → `workflows/todo.md` segment swap (revision 6: these move here from M8a, because the rewritten pin test reads the embedded tree and cannot pass without the rule edit in the same milestone) — and the `factory-dispatch-mechanics.md:199` prose cross-reference naming the old path; the two document tests' post-milestone form reads `todo.md`, their assertions unchanged; update the path-pinning test and the embed-manifest generation (regenerated by the project build); regenerate the published skills.
- **Files (about 24 / 190 lines; above the 10-file rule, so the seam is named: if the milestone splits at run time, the seam is M6b — the rule `paths:` keys, the mechanics cross-reference, and the two document-test path rewrites move to an M6b that lands immediately after the rename edits and before M7, and neither M7 nor M8 starts before both land).** template command file (todo) and its deployed copy, template and deployed router skill, the workflow document renamed in three locations (template, deployed skill tree, deployed-only `.moai/workflows/`), the five template rule sources of the `paths:` update and their deployed mirrors, the `factory-dispatch-mechanics.md` pair (template and deployed), `internal/cli/todo_skill_doc_test.go` and `internal/cli/todo_skill_doc_parity_test.go` (fixture paths rewritten to `todo.md`), `internal/template/embed_manifest_gen.go` (regenerated output), `internal/template/workflow_rule_paths_pinned_test.go`, `internal/template/gtd_canonical_surface_test.go` (inverted), `internal/template/published_skills.go` (the published set loses `moai-gtd`), `.agents/skills/` (regenerated).
- **Packages.** `internal/template`, `internal/cli`.
- **Name-exact selection.** `internal/template`: `TestTodoCanonicalSurfaceGolden|TestWorkflowRulePathsPinned`; `internal/cli`: `TestTodoSkillDocumentsHistoryVerb|TestTodoSkillDocumentsJevSource` (they exist at `todo_skill_doc_test.go:17` and `todo_skill_doc_parity_test.go:19`).

### M7 — The substitution script and its class table (Priority Low; mechanical; needs M6)

- **RED.** (1) A fixture test of the script against a temp tree holding one file per substitution class and one per excluded class: expected output exact, excluded files byte-identical; (2) a second run changes nothing (AC-TU-009 covers both; `TestNoResidualGTDSpelling` is **not** created here: it would stay red over the unconverted corpus until M8 and AC-TU-024 forbids a skip, so it belongs to M8's RED, audit iteration 2 D25). The executable-consumer edits of the old M7 (board mod argument vectors, cutover script default, push-gate refusal text) moved to M5 with the tree deletion.
- **GREEN.** Land `scripts/gtd-to-todo-substitute.sh` (`design.md` §D.3 holds the class table and the dry-run contract; the revision-5 class table adds the `workflows/gtd.md` → `workflows/todo.md` path-string class — whose rule-source members were removed by revision 6, M6 having updated the rule `paths:` keys — the migrated fixture-text classes `binding_ledger.json` and the `roles*/manager-lead.toml` role files, and the rule/policy/manifest/codemap prose classes) and its test.
- **Files (about 3 / 380 lines).** the script, its test, and the class-table fixture files it creates under the test's temp tree.
- **Gap carried.** The board mod's JavaScript tests are not run by any Go selection; the run phase records the runner used, or lists them as an unverified Gap.

### M8 — Apply the script (Priority Low; mechanical; needs M7)

- **Unit and size.** The unit is changed lines, not authored code: about 380 matching lines across roughly 95 files (the 397 hits in 101 files of pre-flight item 2's second command, minus the JSONL recordings the exclusion drops, minus the two always-loaded `AGENTS.md` files M9 edits by hand, and minus `internal/factory/gtd_engage.go` — REQ-TU-011's named exception, one line, no milestone edits it), re-measured at pre-flight after absorbing `main`. The 10-file split rule does not apply; the review is dry-run path-set equality.
- **Batches by origin class (from pre-flight item 7).** M8a: template sources under `internal/template/templates/` (rule prose — the rule `paths:` keys are absent from this batch by revision 6, M6 having updated them to the new path already — skills, policies sources, docs, `AGENTS.md.tmpl` excepted) and the Go test prose strings in `internal/cli` and `internal/template`; the migrated fixture texts (`internal/template/testdata/binding_ledger.json`, the `internal/cli/testdata/codex-rollouts-t1171/roles*/manager-lead.toml` role files) with the tests that consume them checked in the same batch; M8b: deployed copies that have a template source follow the regenerated output, and the deployed-only copies (`.moai/workflows/`, `.moai/docs/`, `.moai/policies/`, `.moai/manifest.json`, `.moai/project/codemaps/`) are converted by the script directly and named in the batch record; M8c: `docs-site/content` in four locales, `docs-site/data/menu/main.yaml`, and the four READMEs (the `moai-gtd` pages keep their paths; their content migrates to `moai todo`; no redirect or page deletion).
- **RED.** `TestNoResidualGTDSpelling` is created here (AC-TU-010), asserting the operator's four-pattern zero-residual grep: its planted-mutant fixture proves the check can go red, its corpus half is observed red against the unconverted tree (397 hits at `3c1809e2f1` are residue until the apply), and its swept-count floor is asserted before the verdict.
- **Mechanism.** Dry-run first; the printed path set must equal the expected set (pre-flight list minus excluded classes); then apply; then re-run, expecting zero replacements and no hit outside the always-loaded pair. M8 closes with the script idempotence observed and `TestNoResidualGTDSpelling` recorded RED — the two `AGENTS.md` files keep it red until M9; the full-residual GREEN is M9's completion condition (AC-TU-010; revision 6 wording).
- **Docs-site obligations.** Four locales change in the same commit batch; the locale-parity baseline is not extended; the docs-site verification recipe (warning-free build, parity, URL scan) runs before the batch is accepted.

### M9 — AGENTS.md, rule statements, and the closing guards (Priority Low; last; needs M8)

- **Content.** The `AGENTS.md` §11 row and its template mirror (`moai todo` canonical; no `moai gtd` spelling, no alias sentence, no "fully supported"); the hand-authored statements in the factory-dispatch rule family (the lane's remaining role, the queue name discipline) and their `.moai/policies` projection, all naming `moai todo` only.
- **Guards.** `TestAgentsMDTodoRowWording` (live file and template mirror) and `TestCodexContractByteCeiling` stay green (AGENTS.md is 23,780 bytes against 24,576); the always-loaded growth statement is written for any edit above 1,000 bytes; the rule-loading budget measurement (count and bytes of always-loaded files before and after) is recorded; after this milestone the residual grep of REQ-TU-011 prints nothing and `TestNoResidualGTDSpelling` is green on the full tree.
- **RED.** The wording test fails on the current row (E-16); the zero-residual test fails while the `AGENTS.md` row still names the alias — the last expected red before the tree is clean.

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
| Argument-vector consumers use `'moai', 'gtd'`, which a `moai gtd` text pattern does not match | A caller keeps invoking a removed command and breaks at run time | M5 reviewed edits plus a grep over the argument-vector form |
| Renaming the queue workflow document breaks a `paths:` loader or the embed manifest | A rule or skill loses its workflow document at load time | The rename, its consumer edits, and the rule sources' `paths:` keys land together in M6 (revision 6); the residual path strings in test prose, policies, and codemaps convert in M8's sweep; `TestWorkflowRulePathsPinned` is rewritten first (red before the rename, green after); the embed line is regenerated by the build; the dry-run path-set equality names every touched file |
| Migrating the recorded fixture texts (`binding_ledger.json`, the `roles*/manager-lead.toml` role files) breaks the tests that read them | A test fails on migrated content | The consuming tests are enumerated at pre-flight item 3 and checked in the same M8a batch; the `.jsonl` recordings stay byte-identical (excluded history) |
| A completing setter repeat in a late card state bumps the factory version | Downstream state churn on an old version | Rare (a half-written pair overtaken past `kickoff`); the stale-version retry contract absorbs it, recorded in `design.md` §D.7 |
| The script test needs `sh` on every CI runner (not measured) | A Windows runner fails the test, or tempts a skip | The test fails loudly if `sh` is absent (no skip, AC-TU-024); M7 RED decides whether the script becomes a Go helper |
| AGENTS.md grows by a few bytes | Ceiling or budget breach | M9 measurement before and after; byte statement |
| The installed `moai` lags the tree | RED/GREEN observed on the wrong build | Build from the tree and invoke by path; record both commits |
| Deleting the `gtd` tree removes a symbol something else still calls (a direct Go caller of a `gtd.go` helper) | Build breakage outside the CLI surface | M5's RED grep enumerates `NewGTDCommand` callers and every `gtd.go` symbol's callers before the deletion; `gtd_answer.go`'s library functions used by `goal.go` stay with their file |

### Known debt

**KD-1 moved to t1878 (revision 5).** The committed-evidence premise of the automatic receipt (`.moai/reports/*` gitignored by operator policy, `.gitignore:235`; the automatic path issues only where a lane force-adds its evidence) belonged to the moved receipt design; the debt statement, its full text, and the trigger to revisit travel to the follow-up card's SPEC with the machinery they describe. This card carries no known-debt item.

## §I Self-verification deliverables (run phase)

Each item carries the command, the verbatim output, and the tree HEAD of this run: the AC matrix with RED-now and GREEN cells; the change-scope measurement from the absorbed-ref merge-base; the cross-platform build; changed-package coverage; lint; the substitution script's dry-run path set and second-run zero; the always-loaded budget measurement; and the Gaps list (the unrun `internal/cli` remainder, the JavaScript tests, the docs-site build if unavailable).

## §J Anti-patterns for this change

- Hand-editing the substitution corpus instead of running the script.
- Substituting `gtd` as a bare word (it appears in identifiers, table names, hex-prefixed ids, recorded fixtures, and the GTD methodology name).
- Restoring any form of warning alias, grace release, or deprecation notice for `moai gtd` — the operator's revision-5 ruling removed the Q2 design; the command is deleted, not deprecated.
- Adding a second handler under the `gtd` spelling, or re-registering a `gtd` verb after M5's deletion.
- Leaving any `moai gtd`, `/moai:gtd`, `moai:gtd`, or `gtd.md` reference outside the excluded classes (`.moai/specs`, `.moai/reports`, `CHANGELOG.md`, `*.jsonl`) — including in `AGENTS.md` (M9), the rule and policy files, the manifest, the codemaps, and the recorded role fixtures. (REQ-TU-011's named exception — the `"moai:gtd-card:"` namespace literal in `internal/factory/gtd_engage.go`, frozen by REQ-TU-012 and pinned by AC-TU-011 — is not a violation; altering those bytes is.)
- Re-introducing the moved done-path scope here (receipt, T29, lane own-card done) instead of leaving it to t1878.
- Editing the deployed `.claude/` copy before the template source.
- Treating a green run with an empty selection as a pass (a name-exact selector that matches no test).

## §K Cross-references

`spec.md` (requirements), `acceptance.md` (criteria), `design.md` (mechanisms), `research.md` (measurements and the evidence ledger), `decision-index.md` (Q1 to Q9, plus Q10, the operator's revision-5 ruling). Related completed SPECs: `SPEC-TODO-LAND-AUTO-DONE-001`, `SPEC-GTD-CANON-BODY-001`, `SPEC-FACTORY-RECORD-001`, `SPEC-LOCAL-MAIN-FLOW-001`.
