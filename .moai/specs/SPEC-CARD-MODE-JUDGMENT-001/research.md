# research.md — SPEC-CARD-MODE-JUDGMENT-001

Measured on tree `74ad197380a0711cdb9816d4e37c94d29af730ea` (branch `WT-fail-safe-serial`, clean, unpushed), cwd `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894`. Judging CLI build for every `moai` observation: `moai_cp/20261007_195927` (installed; it predates this tree's HEAD, so a CLI observation is a statement about that build, not about this tree — `verification-claim-integrity.md` §2.2). No Go command was run in this phase. §10 and the guard rows added by the correction round were measured on tree `55d4f1486a454ff446be6f0f0645d4cea76bf0cc` (the plan-phase commit); the template and deployed document trees are identical between the two pins (`git diff --stat 74ad197380a0711cdb9816d4e37c94d29af730ea 55d4f1486a454ff446be6f0f0645d4cea76bf0cc -- internal .claude .agents .codex` printed nothing, exit 0), so every line number above holds on both. Correction round 1 (§11, §12 and the §6 card-id row) was measured on `b7b44433c312da73de9f329ab08c8fa6c5aa4de0`, whose document trees are likewise identical to `74ad197380` (`git diff --stat 74ad197380a0711cdb9816d4e37c94d29af730ea b7b44433c312da73de9f329ab08c8fa6c5aa4de0 -- internal .claude .agents .codex` printed nothing). Notation: `TPL` = `internal/template/templates/.claude`, `DEP` = `.claude`; rule files live under `rules/moai/workflow/`.

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

`board-spec.txt` is the saved `moai decision read`. Quoted in `spec.md` §2: 0c51 (operator, standing), a672 (leader under operator instruction, standing), 02f4 (operator, card t1737 only), and — added by the correction round — d314 (operator ruling, card t1894), read with `moai decision read --scope card:t1894 --project-root /Users/goos/MoAI/moai-adk-go` (exit 0, `board=ok unparseable=0 records=141`) on tree `55d4f1486a`. `moai todo next --help` prints a `--spec` flag on `next` ("Spec-Id to attach when picking") and no `--spec` on `add`; that supports, but does not prove, 0c51's "cannot be added after assign" (R-3).

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
| S12 | D `factory-dispatch-detail.md` | 133 (Routing bullet) | "The leader routes a card WHOLE to a free lane — free means the lane's previous card reached `done` and its evidence was read. A lane busy on a card is not addressed; …" | **rewrite** the Routing bullet (REQ-MJ-006); **add** `### Delegated-dispatch promotion` after the last bullet of `## Factory in-lane 3-stage` (before `## The card-review stage`, line 140) — home of REQ-MJ-007; line 35 diagram keeps the term |
| S13 | M `factory-dispatch-mechanics.md` | 22 (`## Factory Mode mechanics`), 26 | "A lane renews its own card lease at the start of each watchdog pass with `moai factory renew <card>` …" | **rewrite** + **add** `### Lease renewal` (home of REQ-MJ-008/-009); line 193 (live-session fault) **keep** |
| S14 | A `auto-semantics.md` | 113-146 (§5.1): 115-118 "armed once at card intake, before the first stage, and kept until the completion report"; 127-129 "Each fire runs the watchdog's renewal first (§0a …), so the renewal period is also the cadence at which the card lease is renewed"; 130-135 Keep; also 392-396 "the standing carrier of §5.1 stays armed" | **scope** (keep the tokens `standing`, `CronCreate`, `recurring`, `renew_interval`, `before the first stage`, `CronList`, `CronDelete` — pinned by `TestLaneRecheckDoctrine`) |
| S15 | W `moai-lane-watchdog/SKILL.md` | 51-59 (`## 0a. Renew the card lease (first step of every pass)`), 85-87 | renews unconditionally as first step | **scope** (heading must keep "Renew"; `moai factory renew` before "Measure progress" — pinned by `TestLaneWatchdogRenewsFirst`) |
| S16 | `internal/hook/session_start_factory_i18n.go` | `laneRecheckRule` (lines 66-75 comment; 153-157 text) | "before a card's first stage, arm ONE recurring `CronCreate` recheck … `CronDelete` it when you send the completion report" in four locales | **out of scope** (Go; Q4/R-1) |
| S17 | `…/project/doc-generation.md` | 348 (`moai gtd add "[PROJECT] <goal, one line, bounded by scope>"`), 415 | standing-source card via a bare add | **out of scope** (Q2) |
| S18 | F `factory-dispatch.md` | 38 (long form) and 40 (compressed form) | the operator-act promotion clause; line 38 carries the sentence the `--auto` doctrine tests pin; line 40 is ledger row ALB-0256's `after_text` | **keep both verbatim**; **add** one pointer line after line 40 (REQ-MJ-017a) — decided in §10 as P2/P3 |
| S19 | F line 44 and K line 20 | "The one reconciliation is named, not excepted: `/moai:todo --auto` is the operator's batch approval: …" (identical paragraph in both; K's copy is ledger row ALB-0257 `after_text`, `always:` core, verbatim) | **rewrite** the opening clause only, identically in both (REQ-MJ-017b); update ALB-0257's `after_text` in the same change; −6 UTF-16 units on K |
| S20 | D `factory-dispatch-detail.md` | new subsection (see S12) | no statement of the delegated-dispatch case | **add** (home of REQ-MJ-007) |
| S21 | G `gtd.md` | 307-310 (`## Picking the next card`, the [HARD] pick rule) | "The pick is the operator's. Do not preselect, …" (pinned by `autoDocGTDProhibitions`) | **keep the sentence**; **add** one pointer paragraph after the [HARD] paragraph (REQ-MJ-017c) |

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
| `TestAutoPickDocDoctrine` | `internal/cli` | C must contain the "report-only for an operator-picked card" sentence; F must contain "Outside an --auto authorization the leader never picks for the operator" and must NOT contain "Promotion is the operator's act, always." or "The leader never picks for the operator" (case-sensitive) (`todo_auto_pick_doc_test.go` lines 43-58, 135-143) | a pinned sentence is removed or reworded, or an `absent` string is written |
| `TestAutoRankDoctrineAmendment` | `internal/cli` | F both trees must keep the sentence "Outside an --auto authorization the leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item.", "An empty queue is a state to report, not a prompt to invent work.", "never from queue emptiness, card readiness, or a peer's request", "queue ADMISSION (production) stays the operator's."; G keeps its prohibitions incl. "The pick is the operator's. Do not preselect, do not reorder by inferred priority, and do not append a \"start the top one\" default." (`todo_auto_doc_test.go` lines 113-154) | any of those sentences is removed or reworded |
| `TestAutoRankMirrorParity` | `internal/cli` | the F passage from "[HARD] **Promotion is the operator's act, in person or in advance.**" to "[HARD] **The self-dispatch lane exception.**", and the G `--auto` section, byte-identical live vs template, neutral (no SPEC id, REQ token, ISO date, 9-40 hex run) (`todo_auto_doc_test.go` lines 178-232) | the two copies differ inside the passage, or a line carries internal content — the pointer line of S18 and the clause of S19 lie inside the F passage |
| `TestAutoRankMarkerDisclosure` | `internal/cli` | G both trees keep four `[보류` clauses (`todo_auto_doc_test.go` lines 240-273) | a clause is lost; the S21 paragraph lies outside the `--auto` section and does not touch them |
| `TestTodoSkillDocumentsClassification` | `internal/cli` | G both sides | `--classification-file`, `` `serial` ``, `` `parallelizable` ``, `` `high` ``, `` `normal` ``, `` `low` ``, `sorted` absent; or the template copy carries a SPEC id, REQ token, `20YY-MM-DD` date, or a 9-40 hex run |
| `TestBindingLedgerIntegrity` | `internal/template` | `testdata/binding_ledger.json` (363 rows) vs the deployed tree | a row's `after_text` is no longer at its location and its `before_text` is gone from the source ("obligation lost"); also AC-ALB-021(2): a ledger-covered line of ≥40 UTF-16 units copied verbatim into any `paths:`-scoped rule or skill file |
| `TestRoleCoreStubGate` | `internal/template` | every `*-core.md` | > 10,000 UTF-16 code units or a role-core marker |
| `TestDeployedAlwaysLoadedCharBudget` | `internal/template` | deployed always-loaded surface | rules-only subtotal > 150,000 or any always-loaded file > 40,000 |
| `TestFactoryDispatchBodyCarriesRenewIntervalLinkage` | `internal/template` | F (key `A lane that cannot proceed records an explicit wait on disk`) and K (key `A stuck lane records a disk wait`), both trees | that line lacks ``every renewal period of `workflow.factory_lease.renew_interval` `` or carries `every 20 minutes` / `20-min` |
| `TestLaneRecheckDoctrine` | `internal/template` | A `## 5.` section tokens; F `### Lane waits are explicit, and stalls are watched` tokens (`CronCreate`, `recurring`, `disk evidence`) | a token is absent |
| `TestLaneWatchdogRenewsFirst`, `TestLaneWatchdogRecordsRefusalAndMeasures` | `internal/template` | W | the first procedural step no longer carries "Renew" + `moai factory renew`, or does not precede "Measure progress" |
| `TestWorkflowRulePathsPinned` | `internal/template` | `paths:` frontmatter of the factory rule files | frontmatter changed |
| `TestTemplateNoInternalContentLeak`, `TestCardIDBaselineHasNoStaleEntries`/`…LeakClass…` | `internal/template` | all template files | card id, board id, SPEC id, date, SHA in template text |
| card-id leak baseline (`card_id_leak_test.go`) | `internal/template` | the changed files carry five baselined card-id literals: `t1339`, `t1393` (A), `t133`, `t224` (D), `t696` (G) | a card-id literal outside the baseline appears in a changed file, or a baselined literal is removed so the entry goes stale (`TestCardIDBaselineHasNoStaleEntries`) — the sentences carrying them (A:126, A:146, D:24, D:135, G:258) stay; AC-MJ-013 excludes exactly those five literals with a portable ERE and fires on any other |
| catalog hash audit (`catalog_tier_audit_test.go`, e.g. `TestCatalogHashCoversSkillSubfiles`) | `internal/template` | `catalog.yaml` hashes | a changed skill directory with a stale stored hash |

No mirror-parity guard was found for `moai-lane-watchdog/SKILL.md` (`grep` over `internal/template/*_test.go` and `internal/cli/*_test.go` found only the two doctrine tests above, which read the embedded template side). Template/deployed equality for that file is therefore held only by AC-MJ-011's `cmp` and the discipline of the run phase.

## 7. Sweep set R and the violating-sentence definition (card completion test)

**R** (the documents that describe or constrain issuing a card, by the §5 measurement), template side: `skills/moai/workflows/gtd.md`, `rules/moai/workflow/factory-dispatch.md`, `rules/moai/workflow/factory-dispatch-cards.md`, `agents/moai/manager-todo.md`, `agents/moai/manager-lead.md`, `skills/moai-factory-foreman/SKILL.md`. `doc-generation.md` is excluded by the default of Q2 and listed so adding it is one path.

**A violating sentence (V)**: a non-table line (`^(?!\|)`) of a document of R that contains an issuance invocation `moai todo add "` or `moai gtd add "` and does not carry `--classification-file` on the same line. Table rows are flag *reference*, not issuance instruction (S3 and the `--force` row are rows). Known weakness: an issuance instruction written in prose without the command text escapes V; AC-MJ-003/-002 close that with positive presence checks, and AC-MJ-016 is the single-home check.

Observed on `74ad197380`: V fires on exactly one line — `gtd.md:52`. Control: the same pattern over the base commit fires on the same line.

## 8. Budget

`python3` UTF-16 measurement (the guard counts UTF-16 code units): `factory-dispatch-core.md` = **9,999** (budget 10,000: headroom **1 unit**); `factory-dispatch-cards.md` 10,597, `factory-dispatch.md` 25,803, `-detail.md` 27,611, `-mechanics.md` 26,063, `auto-semantics.md` 29,627 (the last five are `paths:`-scoped, so they are not on the always-loaded surface; the 40,000 per-file figure still applies to them in the InstructionsLoaded measurement per `coding-standards.md` § File Size Limits). Always-loaded template rules: 13 files, 117,085 units (budget 150,000; the guard renders the template first, so treat this as ±a few hundred). The core stub is therefore the binding constraint: the stub edits must net ≤ +1 unit; the planned pair nets −2.

Correction round, re-measured on tree `55d4f1486a`: `python3 -c "import sys; n=len(open(sys.argv[1],encoding='utf-8').read().encode('utf-16-le'))//2; print('OVER' if n>10000 else 'OK', n)" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md` printed `OK 9999`, exit 0; `cmp` of the template and deployed stub printed nothing, exit 0. The S19 clause edit (`The one reconciliation is named, not excepted:` → `A reconciliation is named, not excepted:`, ASCII, 22 → 16 characters) changes the stub by **−6** units. Correction round 1 tightened S10 to the K:28 tail `is kept while the lane executes the card.` (+4), so the two stub edits net **−2** (9,999 → 9,997, simulated; `OK 9997`) and **no offsetting trim is required**. Both edits land in the M2 commit together with their template and deployed twins. The promotion pointer itself lives in F and D, not in the stub: the stub carries no promotion clause today (F:38 and F:40 sit outside the stub), so adding one there would create a stub-only statement and spend headroom the stub does not have. Core lines not covered by any ledger row (candidates for an offsetting trim, each to be checked with `git grep` for pinning tests before use): 3, 26, 28, 48, 54, 62, 64, 84 (measured by matching each core line against every `binding`/`normative` `after_text`).

## 9. Not found / negative results

- No test pins `free means the lane's previous card reached`, `Card size ceiling`, or the `### Mode judgment` heading (grep over `internal/` and `scripts/`): S12 and S1 are free to change.
- No `.agents/skills/**` copy of `gtd.md`, the rules, or the watchdog skill exists.
- `AGENTS.md`, `AGENTS.local.md`, `docs-site/content/en`: 0 matches for `classification-file`, `fail-safe default`, `mode=serial` (`git grep` over the pinned tree).
- No Go test pins the opening clause "The one reconciliation is named" of S19: `git grep -n -c -E "one reconciliation|reconciliation is named" -- internal ':!internal/template/templates'` (tree `55d4f1486a`) printed `internal/homestate/fr_unavailable_race_test.go:1` (an unrelated comment) and `internal/template/testdata/binding_ledger.json:2` (the `before_text` and `after_text` lines of row ALB-0257), exit 0. The clause is pinned by that ledger row only, which this change updates.

## 10. Sentences that state the operator-only promotion rule (correction round)

Pin for every command below: tree `55d4f1486a454ff446be6f0f0645d4cea76bf0cc`, cwd `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894`, `git grep` (no judging build applies). The template and deployed trees are identical between this pin and `74ad197380` (header), so each template hit has a deployed twin at the same line; the twin is shown where the command was run on it. The question each hit answers: does the sentence say promotion belongs to the operator alone, so that the ruling d314 makes it false or incomplete? Three outcomes — **rewrite** (the sentence's own words change), **keep** (unchanged; a pointer may sit beside it), **keep, no pointer** (the sentence is about something the ruling does not touch).

Why nothing but one clause is rewritten: three Go doctrine-test families pin the old sentences (`research.md` §6 rows `TestAutoPickDocDoctrine`, `TestAutoRankDoctrineAmendment`, `TestAutoRankMirrorParity`) and the binding ledger pins the compressed form (ALB-0256, `rule-body:`, checked by `after_text` presence only). This card changes documents and fixtures, not Go, so the pinned sentences stay and the exception is added beside them by pointer — which is also what keeps "each rule stated once".

**C10-1** — `git grep -n -E "Promotion is the operator|never picks for the operator|silently promot|operator.s act" -- internal/template/templates/.claude/rules/moai/workflow internal/template/templates/.claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/agents/moai` → exit 0:

```text
internal/template/templates/.claude/rules/moai/workflow/contract-autonomy.md:92:- Card selection, which stays the operator's act.
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:38:[HARD] **Promotion is the operator's act, in person or in advance.** After a `/clear`, the leader presents the queued cards through `AskUserQuestion` and the operator picks; only then does the leader route the card whole to a free lane. Outside an --auto authorization the leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item. An empty queue is a state to report, not a prompt to invent work. The one ranking the queue admits is the `--auto` cycle's own, bounded in the next paragraph.
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:40:[HARD] Promotion is the operator's act, live or in advance: after `/clear` present the queue via `AskUserQuestion`, the operator picks, then route the card whole to a free lane; outside `--auto` never pick, reorder, or silently promote; report an empty queue, never invent work.
internal/template/templates/.claude/skills/moai/workflows/gtd.md:127:Acting on a record is the operator's act, performed through `drop`, `edit`, or
```

**C10-2** — the same command over `.claude/rules/moai/workflow .claude/skills/moai/workflows/gtd.md .claude/agents/moai` → exit 0: the same four lines, each under the `.claude/` prefix (the first and last at 92 and 127; F at 38 and 40).

**C10-3** — `git grep -n -E "^## Entry into the queue|operator picks a card|operator-picked|operator pick|operator has already picked|After the operator picks" -- internal/template/templates/.claude/rules/moai/workflow internal/template/templates/.claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/agents/moai internal/template/templates/.claude/skills/moai-factory-foreman` → exit 0 (hits other than F:38 and F:40, which are C10-1's; text shortened to the clause that matched, the file and line are exact):

```text
…/factory-dispatch-cards.md:67:  … The operator has already picked the card; a leader that then withholds the dispatch because it found an open pull request has overridden an operator act …
…/factory-dispatch-cards.md:73:  … is a skip input for a queued candidate the session chose and is report-only for an operator-picked card: the operator already chose the second …
…/factory-dispatch-core.md:18:## Entry into the queue is an operator act
…/factory-dispatch-detail.md:35:[operator picks a card]  →  leader routes WHOLE to a free lane  →  lane: plan → run → sync  →  [leader reads evidence, marks done]
…/factory-dispatch.md:26:## Entry into the queue is an operator act
…/moai-factory-foreman/SKILL.md:5:  the next operator-picked card to an isolated worker, collect completion
…/moai-factory-foreman/SKILL.md:37:the next operator-picked card to an isolated worker, collect completion
…/moai/workflows/gtd.md:18:SPEC, not a plan, not an estimate. After the operator picks it, the factory leader follows
…/moai/workflows/gtd.md:410:queue. For a card the operator picked, that state is reported, never a reason to
```

**C10-4** — `git grep -n -E "The pick is the operator|Do not preselect|start the top one" -- internal/template/templates/.claude .claude` → exit 0: `…/moai/SKILL.md:186` ("The pick is the operator's: never preselect, never reorder by inferred priority (the `--auto` cycle's own candidate ranking is the one auto-scoped ranking exception — selection order only), never auto-populate from TODO comments or issues.") and `…/moai/workflows/gtd.md:307-308` ("[HARD] The pick is the operator's. Do not preselect, do not reorder by inferred" / "priority, and do not append a "start the top one" default. Where the queue is"), each on both trees.

**C10-5** — `git grep -n -E "one reconciliation|reconciliation clause" -- internal/template/templates .claude` → exit 0: `factory-dispatch-core.md:20` (the paragraph "The one reconciliation is named, not excepted: `/moai:todo --auto` is the operator's batch approval: …"), `factory-dispatch.md:44` (the identical paragraph) and `gtd.md:378` (`reconciliation clause).`, the citation of that paragraph), on both trees.

**C10-6** — `git grep -n -E "operator.s act|never picks|never pick|silently promot|operator pick" -- internal/cli/todo_auto_doc_test.go internal/cli/todo_auto_pick_doc_test.go` → exit 0:

```text
internal/cli/todo_auto_doc_test.go:115:			"Outside an --auto authorization the leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item.",
internal/cli/todo_auto_doc_test.go:122:			"Outside an --auto authorization the leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item.",
internal/cli/todo_auto_doc_test.go:188:			start:  "[HARD] **Promotion is the operator's act, in person or in advance.**",
internal/cli/todo_auto_pick_doc_test.go:46:				"Outside an --auto authorization the leader never picks for the operator",
internal/cli/todo_auto_pick_doc_test.go:51:				"Promotion is the operator's act, always.",
internal/cli/todo_auto_pick_doc_test.go:52:				"The leader never picks for the operator",
internal/cli/todo_auto_pick_doc_test.go:56:				"Outside an --auto authorization the leader never picks for the operator",
```

Lines 51-52 sit in the test's `absent` list: the strings `Promotion is the operator's act, always.` and `The leader never picks for the operator` (capital T) must not appear in F — the new text avoids both.

**C10-7** — `git grep -n -E "operator.s act" -- internal/cli/todo.go` → exit 0: `todo.go:6` ("// (`/moai todo` is the operator's act — the leader never picks for the"), `todo.go:238` ("act (add), and picking the next card is the operator's act too (next <n>)."), `todo.go:1576` ("// pick stays the operator's act. With <n>: mark the addressed item picked").

**C10-8** — `git grep -n -E "self-promot|Deputy dispatch is limited" -- internal/template/templates/.claude/agents/moai/manager-todo.md internal/template/templates/.claude/agents/moai/manager-lead.md` → exit 0: `manager-lead.md:235` ("4. **Queue mutations** — any `moai gtd` add / pick / done / edit / drop. Deputy dispatch is limited to ALREADY-PICKED cards; admission and closure are operator and leader acts (`factory-dispatch.md` § Entry into the queue is an operator act).") and `manager-todo.md:54` ("  self-promotes, admits, drops, or edits cards, and the queue itself is").

**C10-9** — `git grep -n -i -E "operator.s act|never picks|silently promot|Promotion is" -- docs-site AGENTS.md .moai/docs; echo "exit=$?"` → stdout empty, `exit=1`.

### Decisions per sentence

| Id | Site (file:line, both trees) | Decision | Why |
|---|---|---|---|
| P1 | F:26 and K:18 heading `## Entry into the queue is an operator act` | **keep, no pointer** | the heading names queue ENTRY (admission), which the ruling does not touch ("queue ADMISSION (production) stays the operator's" is pinned and true) |
| P2 | F:38 long promotion clause | **keep verbatim + pointer line** (S18) | pinned by `TestAutoRankDoctrineAmendment`, `TestAutoPickDocDoctrine`, `TestAutoRankMirrorParity` (start marker); its limits ("Outside an --auto authorization …") are the rule the ruling adds ONE stated case to; the pointer names the case without restating it |
| P3 | F:40 compressed promotion clause | **keep verbatim** (same pointer line serves both) | ledger row ALB-0256 `after_text` (a `rule-body:` row: the check is presence of the text in the body); rewriting it would add a ledger edit for no gain |
| P4 | F:44 and K:20 opening clause "The one reconciliation is named, not excepted:" | **rewrite the clause only**, identically in both (`A reconciliation is named, not excepted:`) | "The one" claims there is exactly one reconciliation of the operator-act rule; the ruling adds a second stated case, so the claim becomes false. The rest of the paragraph is pinned and about `--auto`, so it stays. Ledger row ALB-0257 `after_text` updated in the same change; stub −6 units |
| P5 | K:22 and F:52 "A card the operator chose to start when it was issued is not a silent promotion …" | **keep, no pointer** | about an operator-chosen card; unchanged and still true; ledger row verbatim |
| P6 | G:307-308 "[HARD] The pick is the operator's. Do not preselect, …" | **keep the sentence + pointer paragraph** (S21) | it governs the operator's own pick presented through `AskUserQuestion`; pinned by `autoDocGTDProhibitions`; but "the pick is the operator's" read alone is incomplete once a leader may promote without a pick, so a one-line pointer follows |
| P7 | G:352 and G:377-378 (the `--auto` batch-approval paragraph and its citation of "the reconciliation clause") | **keep, no pointer** | about the `--auto` invocation only; the delegated-dispatch case is not an `--auto` invocation; inside the `TestAutoRankMirrorParity` passage |
| P8 | G:18, G:127, G:410; C:67, C:73; D:35 | **keep, no pointer** | G:127 is about relation records; G:18, G:410, C:67, C:73 and D:35 describe the default flow and the operator-picked card. C:73 already separates "a queued candidate the session chose" (skip input) from "an operator-picked card" (report-only); a card the leader promotes under the ruling is a candidate the session chose, so the existing wording covers it without an edit (a reading, not a new rule) |
| P9 | `contract-autonomy.md:92` ("Card selection, which stays the operator's act."); `auto-semantics.md` §9.3 (the card-pick authorization) | **keep, no pointer** | the first lists what a signed contract does not touch (the ruling is not a contract); §9.3 defines the `--auto` pick and its keep-set, which the exception cites and does not restate. Flagged in `progress.md` as a doctrine home the ruling does not mention |
| P10 | `moai-factory-foreman/SKILL.md:5,37` ("operator-picked card"); `moai-lane-watchdog/SKILL.md:219` ("never picks, drops, edits"); `moai/SKILL.md:186` ("The pick is the operator's: …") | **keep, no pointer** | the foreman iteration works picked cards; the watchdog sentence is its own queue-read-only boundary; the moai skill line is an index line that routes to G ("For detailed orchestration: Read workflows/gtd.md") where P6's pointer sits |
| P11 | `manager-lead.md:235`; `manager-todo.md:54` | **keep, no pointer** | the ruling confirms "manager-todo untouched" (Q1); the deputy dispatches already-picked cards and the `--auto` cycle "never self-promotes", while the ruling gives the promotion to the leader |
| P12-P14 | `internal/cli/todo.go:6`, `:238`, `:1576` | **out of scope** (Go) | each concerns the operator's own `next` pick |
| P15 | docs-site, `AGENTS.md`, `.moai/docs` | none found | C10-9 |
| S20 | D new `### Delegated-dispatch promotion` | **add** — the one home | REQ-MJ-007 |

## 11. Standing-cron statements (REQ-MJ-010): the ten sites

Pin: tree `b7b44433c312da73de9f329ab08c8fa6c5aa4de0` (the document trees are identical at `d31fb64239a688c053e455000cc9fe37f540b790`, the pin of correction round 2). Source of the line numbers: the sweep of `acceptance.md` E-008a (keep-word plus cron-context word, marker phrase absent) and the site map S7, S8, S10, S13-S15 above; both trees carry the same lines. The ten sites, each of which makes a claim about when the cron is armed or kept and is scoped to `while the lane executes the card`:

| # | Site | Location | Today | Fragment pinned by AC-MJ-008 (b) |
|---|---|---|---|---|
| 1 | K, the always-loaded stub | K:28 | "… stays armed to the completion report." | `is kept while the lane executes the card.` |
| 2 | F, compressed lane-wait clause | F:119 | "… armed before the first stage, kept to the completion report; …" | `recheck armed before the first stage, kept while the lane executes the card;` |
| 3 | F, long-form lane-wait paragraph | F:122 | "… keeps it until its completion report …" | `keeps it while the lane executes the card:` |
| 4 | A §5.1, intake sentence | A:115-118 (keep-word line 117, tail 118) | "armed once at card intake, before the first stage, and kept until the completion report" | `and kept while the lane executes the card.` |
| 5 | A §5.1, fire-and-renew sentence | A:127-129 | "Each fire runs the watchdog's renewal first …" | the two whole lines `  first (§0a of the watchdog skill), which renews the lease of the card the lane` and `  is executing, so the renewal period is also the cadence at which that lease` (E-008e, added in correction round 2: the sentence carries no keep-word and `when the lane stops executing the card` pins row 6, not this row) |
| 6 | A §5.1, delete sentence | A:130-135 (Keep paragraph) | the cron is deleted at the completion report | `when the lane stops executing the card` (E-008d) |
| 7 | A §14, explicit-wait carrier sentence | A:392-396 | "the standing carrier of §5.1 stays armed" during a wait on the leader | `is kept only while the lane executes the card` |
| 8 | W, constraint bullet | W:244 | "armed at card intake and kept to the completion report" | `armed at card intake and kept while the lane executes the card` |
| 9 | M, renew paragraph | M:26 | "A lane renews its own card lease at the start of each watchdog pass …" | the opening ``While it executes the card, a lane renews its own card lease at the start of each watchdog pass with `moai factory renew <card>` `` (E-008f, added in correction round 2: AC-MJ-007 scans the `### Lease renewal` subsection only and does not read M:26) |
| 10 | M, `### Lease renewal` cron sentence | new subsection, fourth pinned line | — | AC-MJ-007 |

Kept unedited because they only name the cron and make no claim about when it is armed or kept: A:103 (§5 prompt introduction) and A:201 (§8 runner row). The one-shot wait recheck of A §14 (a different mechanism) stays; residual risk R-9. Rows 5 and 6 carry no keep-word on the line the sweep reads, which is why E-008a lists seven lines per tree and the other sites are held by the E-008b and E-008d counts, by E-008e (row 5) and E-008f (row 9), and by AC-MJ-007 (row 10).

## 12. The leader-duty enumeration (REQ-MJ-017 (d))

`factory-dispatch-mechanics.md` § The leader's remaining role is a table that declares itself the boundary of what the leader keeps. A promotion the leader performs under d314 is a duty the table would otherwise not list, so the table and the new subsection would disagree. The stub's pointer sentence (`factory-dispatch-core.md:12`, "What the leader keeps is enumerated, not inherited: `factory-dispatch-mechanics.md` § The leader's remaining role") stays true after one row is added by pointer to D `### Delegated-dispatch promotion`. No ledger row and no Go test pins the table (design.md §D-3, the row for that table); `acceptance.md` E-018e is the presence check. Decision: add one row by pointer, no restatement of the limits.
