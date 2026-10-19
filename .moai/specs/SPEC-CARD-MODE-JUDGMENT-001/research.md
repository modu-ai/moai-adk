# research.md — SPEC-CARD-MODE-JUDGMENT-001

Measured on tree `74ad197380a0711cdb9816d4e37c94d29af730ea` (branch `WT-fail-safe-serial`, clean, unpushed), cwd `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894`. Judging CLI build for every `moai` observation: `moai_cp/20261007_195927` (installed; it predates this tree's HEAD, so a CLI observation is a statement about that build, not about this tree — `verification-claim-integrity.md` §2.2). No Go command was run in this phase. Notation: `TPL` = `internal/template/templates/.claude`, `DEP` = `.claude`; rule files live under `rules/moai/workflow/`.

## 1. Premise record (card text re-checked; premise-check.md P1–P12 plus what this plan added)

| # | Premise | Observed | Reading |
|---|---|---|---|
| P1 | Template and deployed `factory-dispatch-cards.md` agree | `cmp` exit 0 (below §3) | holds. The card's "diff 0" is green before any edit, so it is only meaningful as a paired presence-on-both-sides check (AC-MJ-001..010 use `grep -c` / `awk` over BOTH trees) plus the regression guard AC-MJ-011 |
| P1a | **(new)** the same holds for `factory-dispatch.md` | `diff` exit 1; `git diff --no-index --shortstat` prints ` 1 file changed, 1 insertion(+), 1 deletion(-)` | **false, and deliberately so**: the deployed copy carries one live-only sentence (the worktree-sweep disposal sentence at line 222). It is a *declared fork*: `internal/template/rule_template_mirror_test.go` leaves `factory-dispatch.md` out of byte-parity "(the t1525 CI defect)" and `TestAutoPickMirrorParity` (`internal/cli/todo_auto_pick_doc_test.go`) requires exactly that one-line difference. A "cmp = 0" criterion on `factory-dispatch.md` would be red forever for a reason this work cannot fix |
| P2 | No mode-judgment step exists | headings of C: `## Card classes` line 22 … no mode heading | holds |
| P3/P4 | Issuance procedure documents | see §5 | only two documents describe issuing: `gtd.md:52` (instruction) and `factory-dispatch.md:30` (leader clause); `manager-todo.md` and `manager-lead.md` describe none |
| P5 | `gtd.md` line 56 states the code default accurately | row begins `\| \`moai todo add "<text>"\``, says "Without it the deterministic default decider records `priority=normal, blocked=false, mode=serial, decider=default`" | holds; REQ-MJ-015 forbids rewriting it |
| P6 | Code cite `todo_classify.go:41` | line 41 is `todoClassificationFallbackNotice`; default decider is `factory.DefaultCardDecider{}` at line 36; `ClassModeSerial`/`ClassModeParallelizable` at `internal/factory/classification.go:38-39` | holds (pointer, not a rule location) |
| P7 | Renewal wording conflicts with a672 | see §4 | holds, and **widens**: five document sites, not three, plus one Go site |
| P8 | Free-lane wording | `factory-dispatch-detail.md:133` "free means the lane's previous card reached `done` and its evidence was read"; `…-mechanics.md:193` "A lane with no live session is a fault, not a wait" | two different notions; kept separate (REQ-MJ-006) |
| P9 | Authorities | `moai decision read > /private/tmp/lxq/board-spec.txt` → exit 0, `board=ok unparseable=0 records=821`; 0c51 at lines 6-8, a672 at 16-18, 02f4 at 12-13 of that file | quoted verbatim in `spec.md` §2 |
| P10 | Code is t1737's | `moai factory status` lists `t1737 … owner=lane-18 mode=serial`; `SPEC-FACTORY-DEFAULT-MODE-001` is not in this tree's `.moai/specs` | holds; not referenced in `related_specs` because it does not exist here |
| P11 | Template-First build chain | `Makefile` `build:` = `agents-emit-check commands-emit-check tool-policy-drift-check embed-manifest-check templ-generate` then `go run ./internal/template/scripts/gen-catalog-hashes.go --all` | see §6 |
| P12 | `spec_id` on the card | `SPEC-CARD-MODE-JUDGMENT-001` | holds |

## 2. Authorities

`board-spec.txt` is the saved `moai decision read`. Quoted in `spec.md` §2: 0c51 (operator, standing), a672 (leader under operator instruction, standing), 02f4 (operator, card t1737 only). `moai todo next --help` prints a `--spec` flag on `next` ("Spec-Id to attach when picking") and no `--spec` on `add`; that supports, but does not prove, 0c51's "cannot be added after assign" (R-3).

## 3. Mirror relation per target (cmp exit codes, tree `74ad197380`)

| Target (under `…/`) | `cmp TPL DEP` | Other copies |
|---|---|---|
| `rules/moai/workflow/factory-dispatch-cards.md` | 0 | `.moai/policies/workflow/…` differs (stale projection: lacks the card-size-ceiling and workflow-file sections) |
| `rules/moai/workflow/factory-dispatch-core.md` | 0 | none (`.moai/policies/workflow/factory-dispatch-core.md` is not tracked) |
| `rules/moai/workflow/factory-dispatch.md` | **1** (declared one-line fork, P1a) | `.moai/policies/…` differs at char 1 |
| `rules/moai/workflow/factory-dispatch-detail.md` | 0 | `.moai/policies/…` differs |
| `rules/moai/workflow/factory-dispatch-mechanics.md` | 0 | `.moai/policies/…` differs |
| `rules/moai/workflow/auto-semantics.md` | 0 | `.moai/policies/…` differs |
| `skills/moai/workflows/gtd.md` | 0 | `.moai/workflows/gtd.md` differs; **no** `.agents/skills` copy carries its body (`.agents/skills/moai-gtd/SKILL.md` is a 256-byte dispatcher stub generated from `.claude/commands/moai/*` by `internal/template/commandemit`) |
| `skills/moai-lane-watchdog/SKILL.md` | 0 | none tracked |
| `agents/moai/manager-todo.md`, `agents/moai/manager-lead.md`, `skills/moai-factory-foreman/SKILL.md` | 0 | `.codex/agents/moai/manager-{todo,lead}.toml` are machine-emitted from the agent `.md` (`make agents-emit`); not edited by default (Q1) |

`.moai/policies/**` and `.moai/workflows/gtd.md` are the deploy-time projection (`internal/template/harness_fs.go` line 138 maps `.moai/policies` ← `.claude/rules/moai`). They are tracked here, already stale (`diff` of the template C against `.moai/policies/workflow/factory-dispatch-cards.md` printed the `## Card size ceiling` and `## Workflow-file changes need operator review` sections as present only on the template side) and not a guard target.

## 4. Site map — sentences to add or rewrite (identical line numbers on both trees unless noted)

Command: `grep -n` over each file, plus `Read` of the ranges. Each line below was read on the pinned tree.

| Id | File | Lines | Today | Change class |
|---|---|---|---|---|
| S1 | C `factory-dispatch-cards.md` | 22-34 (`## Card classes`; table 26-30; Class-A rule 32; parallelism 34) | no mode axis | **add** `### Mode judgment` after line 34 (home of REQ-002/-003/-005, partly -004) |
| S2 | G `gtd.md` | 52 | "run `moai todo add "<description>"`; a bare `/moai todo` runs `moai todo list`." | **rewrite** the issuance instruction (flag + cite C); the only line of sweep set R that fires V (§7) |
| S3 | G `gtd.md` | 56 (table row) | factual default-decider sentence | **keep** (REQ-MJ-015) |
| S4 | G `gtd.md` | 267-270 | "`spec_id` — a record-only, optional annotation attached only when `next --spec` is explicitly given. An empty `spec_id` on a picked card is the normal state; …" | **qualify** for Class C (REQ-MJ-005) |
| S5 | G `gtd.md` | 330-333 | "Record it with `moai todo next <n> [--spec <SPEC-ID>]` … Attach the SPEC only when one exists and its identifier is known. … A pick alone neither creates a SPEC nor requires one." | **qualify** for Class C |
| S6 | F `factory-dispatch.md` | 30 | "[HARD] The leader is the sole queue producer … `moai gtd add` turns an operator request into a card …" — pinned by ledger row ALB-0254 | **keep verbatim**; **add** one new [HARD] line directly after it (REQ-MJ-004) |
| S7 | F `factory-dispatch.md` | 119 | compressed lane-wait line "… a recurring `CronCreate` recheck armed before the first stage, kept to the completion report; …" — ledger row ALB-0276 (`role-core:` location, `processing: rewrite`) | **rewrite** (scope) + update ledger row |
| S8 | F `factory-dispatch.md` | 122 | long paragraph; "… keeps it until its completion report …"; key line of `TestFactoryDispatchBodyCarriesRenewIntervalLinkage` | **rewrite** (scope), keep ``every renewal period of `workflow.factory_lease.renew_interval` `` on the same line |
| S9 | F `factory-dispatch.md` | 222 | worktree-sweep sentence (declared fork) | **do not touch** |
| S10 | K `factory-dispatch-core.md` | 28 | "[HARD] A stuck lane records a disk wait … a recurring recheck cron (`CronCreate`, every renewal period of `workflow.factory_lease.renew_interval`) stays armed to the completion report." | **rewrite** (scope) within the budget (§8) |
| S11 | K | 32, 36, 68 | Card-classes paragraph (ALB-0265), worktree-sharing "serially" (ALB-0268), "routes each card WHOLE to a free lane" (ALB-0304) | **keep verbatim** (ledger-covered) |
| S12 | D `factory-dispatch-detail.md` | 133 (Routing bullet) | "The leader routes a card WHOLE to a free lane — free means the lane's previous card reached `done` and its evidence was read. A lane busy on a card is not addressed; …" | **rewrite** (home of REQ-MJ-006/-007); line 35 diagram keeps the term |
| S13 | M `factory-dispatch-mechanics.md` | 22 (`## Factory Mode mechanics`), 26 | "A lane renews its own card lease at the start of each watchdog pass with `moai factory renew <card>` …" | **rewrite** + **add** `### Lease renewal` (home of REQ-MJ-008/-009); line 193 (live-session fault) **keep** |
| S14 | A `auto-semantics.md` | 113-146 (§5.1): 115-118 "armed once at card intake, before the first stage, and kept until the completion report"; 127-129 "Each fire runs the watchdog's renewal first (§0a …), so the renewal period is also the cadence at which the card lease is renewed"; 130-135 Keep; also 392-396 "the standing carrier of §5.1 stays armed" | **scope** (keep the tokens `standing`, `CronCreate`, `recurring`, `renew_interval`, `before the first stage`, `CronList`, `CronDelete` — pinned by `TestLaneRecheckDoctrine`) |
| S15 | W `moai-lane-watchdog/SKILL.md` | 51-59 (`## 0a. Renew the card lease (first step of every pass)`), 85-87 | renews unconditionally as first step | **scope** (heading must keep "Renew"; `moai factory renew` before "Measure progress" — pinned by `TestLaneWatchdogRenewsFirst`) |
| S16 | `internal/hook/session_start_factory_i18n.go` | `laneRecheckRule` (lines 66-75 comment; 153-157 text) | "before a card's first stage, arm ONE recurring `CronCreate` recheck … `CronDelete` it when you send the completion report" in four locales | **out of scope** (Go; Q4/R-1) |
| S17 | `…/project/doc-generation.md` | 348 (`moai gtd add "[PROJECT] <goal, one line, bounded by scope>"`), 415 | standing-source card via a bare add | **out of scope** (Q2) |

## 5. Issuance documents measured (P3/P4 resolved)

`git grep -n "todo add\|gtd add"` over `internal/template/templates` (md/tmpl/yaml/sh):

- `rules/moai/core/moai-mcp-tools-catalogue.md:187` — table row naming `mcp__moai__todo_add` (tool listing, not a procedure).
- `rules/moai/workflow/factory-dispatch.md:30` — leader clause (S6).
- `rules/moai/workflow/main-checkout-branch-guard-detail.md:128-129` — a quoted example of a command that carries text as data (unrelated).
- `skills/moai/workflows/gtd.md:52,56,69` — S2, S3, and the `--force` row.
- `skills/moai/workflows/project/doc-generation.md:336,348,415` — S17.
- `skills/moai-factory-foreman/SKILL.md:66` — "Never run `moai gtd add`" (a prohibition).
- `agents/moai/manager-todo.md`: 0 matches for `todo add` / `gtd add` / `classif`; it says the `--auto` cycle "never self-promotes, admits, drops, or edits cards". `agents/moai/manager-lead.md:235`: lists `moai gtd` add among queue mutations the deputy may not perform.

The card's scope item (2) names the leader, manager-todo and the todo skill. Measured: the leader's issuance clause is `factory-dispatch.md:30`; the todo skill's is `gtd.md:52`; `manager-todo.md` has no issuance step to make explicit. Recorded as Q1.

`moai todo add --help` run from this lane session prints a refusal ("lane boundary: a lane session cannot mutate the queue"), consistent with the catalogue row and the lane rules; it is not a defect.

## 6. Sync mechanism and what fails when sides differ

- **Template-First** (`AGENTS.local.md` §2 "Template-First Rule"): edit `internal/template/templates/<path>` first → `make build` → sync to the deployed copy (`moai update` or manual copy). `make build` runs `agents-emit-check`, `commands-emit-check`, `tool-policy-drift-check`, `embed-manifest-check`, `templ-generate`, then `gen-catalog-hashes.go --all` and `go build`. The `*-check` targets are read-only drift detectors, not syncs.
- **Embed manifest**: `embed-manifest-check` compares the committed `embed_manifest_gen.go` with the tracked template *file set*; editing existing files adds none, so it is unaffected (verified by reading the target text, not run).
- **Catalog hashes**: `internal/template/catalog.yaml` stores a sha256 per skill/agent entry; skill entries hash the whole directory (`gen-catalog-hashes.go` header). Entries touched: `moai` (`templates/.claude/skills/moai/`, contains `workflows/gtd.md`), `moai-lane-watchdog`. `gen-catalog-hashes.go --entry <name>` rewrites one. Rule files are not catalog entries.
- **Guards that fail when the two sides differ or a pinned token moves** (all Go, run-phase only):

| Guard | Package | Reads | Fails when |
|---|---|---|---|
| `TestRuleTemplateMirrorDrift` (+ `TestDeclaredRuleMirrorForks`) | `internal/template` | enrolled rule pairs incl. `factory-dispatch-cards`, `-gates`, `-core`, `-mechanics` (`rule_template_mirror_test.go` lines 239-240, 332-333) | the deployed and template copy differ by a byte |
| `TestReviewOwnership_CardReviewDoctrine` | `internal/template` | `factory-dispatch-detail.md` both trees (`codex_review_ownership_m4_test.go` line 540) | the two copies differ at all, or the `[card-review]` ordered stage list / evidence path / ceiling is missing from D |
| `TestAutoPickMirrorParity` | `internal/cli` | seven pairs: F, C, A, G, `manager-todo.md`, foreman SKILL, `moai-mcp-tools-catalogue.md` | any pair other than F differs; F must differ by **exactly** the line containing `moai worktree sweep` |
| `TestAutoPickDocDoctrine` | `internal/cli` | C must contain the "report-only for an operator-picked card" sentence | that sentence is removed or reworded |
| `TestTodoSkillDocumentsClassification` | `internal/cli` | G both sides | `--classification-file`, `` `serial` ``, `` `parallelizable` ``, `` `high` ``, `` `normal` ``, `` `low` ``, `sorted` absent; or the template copy carries a SPEC id, REQ token, `20YY-MM-DD` date, or a 9-40 hex run |
| `TestBindingLedgerIntegrity` | `internal/template` | `testdata/binding_ledger.json` (363 rows) vs the deployed tree | a row's `after_text` is no longer at its location and its `before_text` is gone from the source ("obligation lost"); also AC-ALB-021(2): a ledger-covered line of ≥40 UTF-16 units copied verbatim into any `paths:`-scoped rule or skill file |
| `TestRoleCoreStubGate` | `internal/template` | every `*-core.md` | > 10,000 UTF-16 code units or a role-core marker |
| `TestDeployedAlwaysLoadedCharBudget` | `internal/template` | deployed always-loaded surface | rules-only subtotal > 150,000 or any always-loaded file > 40,000 |
| `TestFactoryDispatchBodyCarriesRenewIntervalLinkage` | `internal/template` | F (key `A lane that cannot proceed records an explicit wait on disk`) and K (key `A stuck lane records a disk wait`), both trees | that line lacks ``every renewal period of `workflow.factory_lease.renew_interval` `` or carries `every 20 minutes` / `20-min` |
| `TestLaneRecheckDoctrine` | `internal/template` | A `## 5.` section tokens; F `### Lane waits are explicit, and stalls are watched` tokens (`CronCreate`, `recurring`, `disk evidence`) | a token is absent |
| `TestLaneWatchdogRenewsFirst`, `TestLaneWatchdogRecordsRefusalAndMeasures` | `internal/template` | W | the first procedural step no longer carries "Renew" + `moai factory renew`, or does not precede "Measure progress" |
| `TestWorkflowRulePathsPinned` | `internal/template` | `paths:` frontmatter of the factory rule files | frontmatter changed |
| `TestTemplateNoInternalContentLeak`, `TestCardIDBaselineHasNoStaleEntries`/`…LeakClass…` | `internal/template` | all template files | card id, board id, SPEC id, date, SHA in template text |
| catalog hash audit (`catalog_tier_audit_test.go`, e.g. `TestCatalogHashCoversSkillSubfiles`) | `internal/template` | `catalog.yaml` hashes | a changed skill directory with a stale stored hash |

No mirror-parity guard was found for `moai-lane-watchdog/SKILL.md` (`grep` over `internal/template/*_test.go` and `internal/cli/*_test.go` found only the two doctrine tests above, which read the embedded template side). Template/deployed equality for that file is therefore held only by AC-MJ-011's `cmp` and the discipline of the run phase.

## 7. Sweep set R and the violating-sentence definition (card completion test)

**R** (the documents that describe or constrain issuing a card, by the §5 measurement), template side: `skills/moai/workflows/gtd.md`, `rules/moai/workflow/factory-dispatch.md`, `rules/moai/workflow/factory-dispatch-cards.md`, `agents/moai/manager-todo.md`, `agents/moai/manager-lead.md`, `skills/moai-factory-foreman/SKILL.md`. `doc-generation.md` is excluded by the default of Q2 and listed so adding it is one path.

**A violating sentence (V)**: a non-table line (`^(?!\|)`) of a document of R that contains an issuance invocation `moai todo add "` or `moai gtd add "` and does not carry `--classification-file` on the same line. Table rows are flag *reference*, not issuance instruction (S3 and the `--force` row are rows). Known weakness: an issuance instruction written in prose without the command text escapes V; AC-MJ-003/-002 close that with positive presence checks, and AC-MJ-016 is the single-home check.

Observed on `74ad197380`: V fires on exactly one line — `gtd.md:52`. Control: the same pattern over the base commit fires on the same line.

## 8. Budget

`python3` UTF-16 measurement (the guard counts UTF-16 code units): `factory-dispatch-core.md` = **9,999** (budget 10,000: headroom **1 unit**); `factory-dispatch-cards.md` 10,597, `factory-dispatch.md` 25,803, `-detail.md` 27,611, `-mechanics.md` 26,063, `auto-semantics.md` 29,627 (the last five are `paths:`-scoped, so they are not on the always-loaded surface; the 40,000 per-file figure still applies to them in the InstructionsLoaded measurement per `coding-standards.md` § File Size Limits). Always-loaded template rules: 13 files, 117,085 units (budget 150,000; the guard renders the template first, so treat this as ±a few hundred). The core stub is therefore the binding constraint: S10 must be net ≤ 0 units. Core lines not covered by any ledger row (candidates for an offsetting trim, each to be checked with `git grep` for pinning tests before use): 3, 26, 28, 48, 54, 62, 64, 84 (measured by matching each core line against every `binding`/`normative` `after_text`).

## 9. Not found / negative results

- No test pins `free means the lane's previous card reached`, `Card size ceiling`, or the `### Mode judgment` heading (grep over `internal/` and `scripts/`): S12 and S1 are free to change.
- No `.agents/skills/**` copy of `gtd.md`, the rules, or the watchdog skill exists.
- `AGENTS.md`, `AGENTS.local.md`, `docs-site/content/en`: 0 matches for `classification-file`, `fail-safe default`, `mode=serial` (`git grep` over the pinned tree).
