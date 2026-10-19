# acceptance.md — SPEC-CARD-MODE-JUDGMENT-001

The verification layer. Requirements (GEARS) are in `spec.md` §4; this file holds Given-When-Then criteria, each with an evidence-ledger entry. A GEARS requirement is never restated here and a Given-When-Then scenario is never a requirement.

## 0. Conventions

- **Document-level pin** (binds every ledger entry that carries no pin of its own, `verification-completeness.md` §2.1): tree SHA `74ad197380a0711cdb9816d4e37c94d29af730ea`, branch `WT-fail-safe-serial` (clean, unpushed). Every command ran with cwd `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894`; the judging `moai` build for CLI observations is `moai_cp/20261007_195927`. No Go command ran in this phase. Entries added or re-measured by the correction round (E-006, E-016, E-018a-d, E-019a-c) carry a criterion-level pin of their own, tree `55d4f1486a454ff446be6f0f0645d4cea76bf0cc`, which wins over the document-level pin; the template and deployed document trees are identical between the two pins (`research.md` header).
- **Carrier**: fenced evidence-ledger entries `E-nnn`, cited by the criteria. Each entry carries `command`, `stdout` (verbatim as printed), `exit` (its own field), and, where it differs from the document pin, `tree`.
- **Classes**: `R` release-blocking (has a RED-now cell that is red now for the stated reason, a green-path cell naming the milestone that flips it, and a mutant probe); `G` regression-guard (green by design now; it records a *known failing input* on which the instrument was observed red, so the green is interpretable — `verification-completeness.md` §1.1); `O` operational (undecidable by a command; recorded, never counted as a pass by this SPEC).
- **Both trees**: every `R` criterion reads the template side and the deployed side in one invocation, so an edit to one side only leaves one operand red. `AC-MJ-011` additionally holds byte equality (`cmp`) for the seven fully-mirrored pairs and `AC-MJ-017` holds the one declared fork.
- **`grep -c` over several files** prints one `path:count` line per file; the order of those lines is not part of the expectation (the tool displayed the deployed copy first although the template path was named first), so green outputs below are stated as sets.
- Milestones M1-M5 are those of `plan.md`.

## 1. Criteria

### AC-MJ-001 — The mode-judgment step exists on both sides with its five contract tokens (R; REQ-MJ-001, -002, -003)

**Given** the cards companion `factory-dispatch-cards.md` on the template side and the deployed side,
**When** each file is scanned for a `### Mode judgment` subsection and the tokens `expected file set`, `in-flight card`, `slot-gated surface`, `conflict check`, `--classification-file` are counted *inside that subsection only*,
**Then** the scan prints `5` for the template copy and `5` for the deployed copy.

- RED-now: `E-001` (`0`, `0`). Red because neither copy has a `### Mode judgment` heading; the card-classes section ends at the parallelism paragraph (`research.md` S1).
- Green path: M1 adds the subsection to the template and then the deployed copy → `5` / `5`.
- Mutants tried: heading with no body → `0`; tokens only in a comment or another section → `0`; subsection in the template only → `5` / `0`; subsection present with the five tokens but the predicate semantically wrong or reversed → **passes** (residual, spec R-2 — read by the plan/sync audit, not by this command).

### AC-MJ-002 — No issuance instruction in sweep set R relies on the default decider (R; REQ-MJ-004)

**Given** the six documents of sweep set R (`research.md` §7),
**When** the sweep for a violating sentence V is run, and a positive presence check runs on the todo skill,
**Then** (c1) the sweep over the working tree prints nothing and exits 1; (c2) the same sweep over the base commit still prints the `gtd.md:52` line and exits 0 (the control: the pattern is live); (c3) the swept set is exactly the six paths; (c4) each copy of `gtd.md` has at least one non-table line carrying both `moai todo add "` and `--classification-file`.

V is defined once: a non-table line (`^(?!\|)`) containing `moai todo add "` or `moai gtd add "` that does not contain `--classification-file`.

- RED-now: `E-002a` (c1 prints the `gtd.md:52` line, exit 0) and `E-002d` (c4 prints `:0`, `:0`, exit 1); `E-002b` (control) and `E-002c` (swept set) hold their green values already. Red because `gtd.md:52` is the sole instruction to issue a card and it names no classification.
- Green path: M1 rewrites that instruction → c1 empty/exit 1, c4 `:1`-or-more on both copies; c2 and c3 unchanged.
- Mutants tried: delete the instruction (c1 passes, c4 fails); add the flag to a table row only (c1 unchanged — the row was never swept; c4 still `0`); put the flag on a later physical line of a wrapped instruction (c1 still fires — a deliberate false positive, strict); write the instruction as prose with no command text (c1 and c4 both blind — AC-MJ-003 and spec REQ-MJ-004's cite of C carry it; residual).

### AC-MJ-003 — The leader's queue-entry clause requires the explicit classification and cites its rule (R; REQ-MJ-004)

**Given** `factory-dispatch.md` on both sides,
**When** lines are counted that start with `[HARD]` and contain both `--classification-file` and `factory-dispatch-cards.md`,
**Then** each copy reports `1`.

- RED-now: `E-003` (`:0`, `:0`, exit 1). Red because the leader clause at line 30 names only the verb `moai gtd add`.
- Green path: M1 adds one new [HARD] line directly after line 30 on both sides (line 30 itself is ledger-pinned and stays verbatim, `design.md` §D-3) → `:1` / `:1`.
- Mutants tried: new line in one tree only → `:1` / `:0`; line without the cite → `:0`; line placed in C instead of F → `:0`.

### AC-MJ-004 — Class C: `--spec` is given at pick, and an empty `spec_id` is not called normal (R; REQ-MJ-005)

**Given** C and `gtd.md` on both sides,
**When** (a) C is searched for the literal `cannot be added after assign` and (b) `gtd.md` is searched for the unqualified fragment ``An empty `spec_id` on a picked card is``,
**Then** (a) reports `1` for each copy of C and (b) reports `0` for each copy of `gtd.md`.

- RED-now: `E-004a` (`:0`, `:0`, exit 1) and `E-004b` (`:1`, `:1`, exit 0). Red because C says nothing about Class C picks and `gtd.md:267-270` still calls an empty `spec_id` the normal state.
- Green path: M1 adds the sentence to `### Mode judgment` and qualifies `gtd.md` S4/S5 → (a) `:1` / `:1`, (b) `:0` / `:0`.
- Mutants tried: sentence in `gtd.md` instead of C (a fails, and AC-MJ-016 lists a second file); delete the `spec_id` paragraph (b passes — the requirement is a prohibition, deletion satisfies it); reword to "An empty `spec_id` on a picked card is rare" (b passes; the sentence now carries no normality claim — acceptable).

### AC-MJ-005 — A free lane is defined by "no card actually executing", and the old definition is gone (R; REQ-MJ-006)

**Given** `factory-dispatch-detail.md` on both sides,
**When** (a) the literal `no card is actually executing` is counted, (b) the literal `free means the lane's previous card reached` is counted, and (c) the literal `not a free lane` (the separation from the live-session condition) is counted,
**Then** (a) reports `1` per copy, (b) reports `0` per copy, and (c) reports at least `1` per copy.

- RED-now: `E-005a` (`:0`, `:0`, exit 1), `E-005b` (`:1`, `:1`, exit 0) and `E-005c` (`:0`, `:0`, exit 1). Red because D:133 still defines free as "the lane's previous card reached `done`" and says nothing about the live-session condition.
- Green path: M3 rewrites the Routing bullet on both sides.
- Mutants tried: add the new sentence and keep the old (b stays `1`); delete the term "free" altogether (a fails); define free in M instead (a fails; AC-MJ-016 lists M).

### AC-MJ-006 — The delegated-dispatch promotion is stated once, with every limit of the operator ruling and no other (R; REQ-MJ-007)

**Given** the detail companion D on both sides,
**When** the subsection `### Delegated-dispatch promotion` is scanned and each non-empty line of it is compared, as a whole line, with the nine lines pinned in `E-006` — the command prints `<distinct subsection lines equal to a pinned line> <non-empty subsection lines>` once per file —
**Then** each copy reports `9 9`.

- RED-now: `E-006` (`0 0` for the template copy, `0 0` for the deployed copy; exit 0; tree `55d4f1486a`). Red because D has no such subsection — its `##` headings run from `## Design intent` to `## Cross-references` and none is this one (`research.md` S12, S20). This work can flip it: M3 adds the subsection to the template copy and then the deployed copy.
- Green path: M3 step 2 → `9 9` and `9 9`. The nine lines are the ruling's limits in a fixed wording: the delegation condition (line 1), where (line 2), which cards (line 3), what is never promoted (line 4), serial cards (line 5), order (line 6), the leader's act (line 7), what is unchanged (line 8), and the assignment input (line 9). Positive control: pasting those nine lines under the heading into any markdown file and running the same command on it prints `9 9`.
- Why whole lines, not tokens: every limit of the ruling is a restriction or a negation, and a token check cannot see a limit reversed inside a sentence that still carries the token. Whole-line equality plus the line count makes a widening visible however it is written — a reworded limit changes a line, an added limit adds a line, a dropped limit removes one.
- Mutants tried — each observed once against a scratch copy of the pinned subsection (the scratch files are not cited; see Residual-risk in `progress.md`): remove "a held card" from the never-promoted line → `8 9`; remove the delegation condition from the first line → `8 9`; replace the serial line with "serial cards are promoted too" → `8 9`; append ", and held cards too" to the first line → `8 9`; replace the leader's-act line by a copy of the order line → `8 9`; add a bullet "Held cards may also be promoted." → `9 10`. Renaming the heading, or moving the lines into another section → `0 0`. Each is red.
- Evades, stated: the same widening written in a second file under a paraphrase that carries none of the six home tokens (AC-MJ-016 lists any file carrying one; a pure paraphrase evades — read by the audit). A wording improvement to a pinned line fails the command by design; the criterion is amended with the wording, the same strictness as AC-MJ-002's deliberate false positive.

### AC-MJ-007 — The renewal rule has one home and states its five facts (R; REQ-MJ-008, -009)

**Given** M on both sides,
**When** the subsection `### Lease renewal` is scanned for the tokens `only the card the lane is actually executing`, `merge-ready`, `held`, `serial-waiting`, `expiry is harmless`, `notice`, `factory next --card`,
**Then** the scan prints `7` and `7`.

- RED-now: `E-007` (`0`, `0`). Red because M:26 renews "at the start of each watchdog pass" with no executing-card scope and there is no `### Lease renewal` subsection.
- Green path: M2 adds the subsection (and rewrites M:26 to point at it) on both sides → `7` / `7`.
- Mutants tried: heading only → `0`; subsection in the template only → `6` / `0`; the six tokens in a paragraph that says the opposite (passes; residual, audit).

### AC-MJ-008 — Every statement that keeps the standing cron to the completion report is scoped, and none is deleted (R; REQ-MJ-010)

**Given** K, F and A (the three files with such a statement) and M on both sides,
**When** (a) lines are counted that contain `kept`, `keeps` or `stays armed` followed by `completion report` and do **not** contain `while the lane executes the card`; (b) the scope marker `while the lane executes the card` is counted in K, F, A and M; (c) the wording ``every renewal period of `workflow.factory_lease.renew_interval` `` is counted in K and F,
**Then** (a) reports `0` for all six files; (b) reports at least `1` for every one of the eight files; (c) reports `1` for each of K and F on both sides.

- RED-now: `E-008a` (K `1`/`1`, F `2`/`2`, A `1`/`1`, exit 0) and `E-008b` (all eight `:0`, exit 1). Red because K:28, F:119, F:122 and A:117-118 say "stays armed / kept / keeps … completion report" with no scope. Control for (a): `E-008d` prints the same counts over the base commit, so the pattern is live. `E-008c` (the wording count) is a regression-guard cell: it is `1`/`1`/`1`/`1` today and must stay so (a deletion mutant of (a) would otherwise pass).
- Green path: M2 → (a) `0` ×6, (b) ≥ 1 ×8, (c) unchanged.
- Mutants tried: delete the cron sentences (a passes; (b) and (c) fail); scope K and F but leave A's wrapped tail line "kept until the completion report." unscoped (a fires — this is the per-physical-line trap of `design.md` §D-5); put the marker in a code span (passes; harmless); scope while changing the renewal wording to `every 20 minutes` (c fails, and `TestFactoryDispatchBodyCarriesRenewIntervalLinkage` fails).

### AC-MJ-009 — The watchdog's first-step renewal is scoped to the executing card (R; REQ-MJ-011)

**Given** `moai-lane-watchdog/SKILL.md` on both sides,
**When** the section `## 0a` is scanned for the tokens `executing` and `merge-ready`,
**Then** the scan prints `2` and `2`.

- RED-now: `E-009` (`0`, `0`). Red because §0a renews "before anything else" with no state condition.
- Green path: M2 adds a scope clause that keeps the heading word `Renew`, the `moai factory renew` call and the closed set of `renew_result` values (`design.md` §D-3).
- Mutants tried: heading only; clause appended in §1 (the awk is section-scoped → `0`); reordering Renew after Measure (the Go guard `TestLaneWatchdogRenewsFirst` fails, run-phase).

### AC-MJ-010 — No reclassification verb is described as available; the factual default sentence stands (G; REQ-MJ-015)

**Given** the template and deployed trees,
**When** (a) tracked files under `internal/template/templates/.claude` and `.claude` are searched for `moai (todo|gtd|factory) reclassif…` or `reclassify verb`, and (b) the `gtd.md` default-decider sentence is counted,
**Then** (a) prints nothing and exits 1; (b) reports `1` per copy.

- Known failing input: `E-010c` — the same pattern over `.moai/reports/t1894/premise-check.md`, which states the verb, prints `1` (exit 0): the instrument goes red on text that describes the verb.
- Evidence now: `E-010a` (empty, exit 1), `E-010b` (`:1`, `:1`, exit 0).
- Mutants tried: a sentence "use `moai todo reclassify <id>`" (a fires); rewriting the default sentence of the `moai todo add` row (b fires); a sentence "reclassification verb" without the word `reclassify` ("re-classify") evades (a) (residual).

### AC-MJ-011 — Template and deployed copies stay byte-identical for the seven fully mirrored files (G; REQ-MJ-013)

**Given** the pairs C, K, D, M, A, G, W,
**When** `cmp <template> <deployed>` runs for each,
**Then** each exits `0` with no output.

- Evidence now: `E-011` (seven exit-0 results). This is green before any edit (premise P1) and is **only** meaningful together with AC-MJ-001..009, which read both sides and are red now; it cannot be satisfied by "edit neither" because those criteria then stay red.
- Known failing input: `E-011f` — `cmp` on the `factory-dispatch.md` pair exits `1` (`differ: char 23127, line 222`); the instrument reports a difference when one exists. That pair is the declared fork and is excluded here and covered by AC-MJ-017.
- Mutants tried: edit the template only → one cmp exits 1 and the matching `R` criterion shows `:0` on the deployed operand; edit both sides to different wording → cmp exits 1.

### AC-MJ-012 — The always-loaded core stub stays within 10,000 UTF-16 code units (G; REQ-MJ-014)

**Given** `factory-dispatch-core.md` (template side; the deployed copy is `cmp`-equal),
**When** its UTF-16 code-unit count is compared with 10,000,
**Then** the command prints `OK <n>` with `n ≤ 10000`.

- Evidence now: `E-012a` prints `OK 9999` (headroom 1; re-measured on tree `55d4f1486a`, exit 0, `cmp` of the two stub copies silent, exit 0). This is not vacuous: the work must edit this file twice — site S10 (marker form, +9 units) and site S19 (the opening clause, −6 units) — so the net before any offsetting trim is +3 and the criterion is red unless a trim of at least 2 units from an uncovered core line (or a tighter S10 wording) lands with them (`design.md` §D-4).
- Known failing input: `E-012b` — the same command on `factory-dispatch.md` prints `OVER 25803`.
- Mutants tried: pointer-form wording (+115) → `OVER 10114`; a trim that removes a ledger-covered line (this criterion passes; `TestBindingLedgerIntegrity` fails, run-phase).
- Limit: the measurement is a UTF-16 count of the file as stored; the Go gate renders the template first. The two agree for this file (no template directives); the Go gate is authoritative.

### AC-MJ-013 — The changed template text carries no card id, board id, SPEC id, date, or long hex run (G; REQ-MJ-014)

**Given** the eight changed template files,
**When** `git diff --stat -G<leak pattern>` is run against the pinned commit,
**Then** it prints nothing.

- Evidence now: `E-013a` (empty, exit 0 — the diff against the pinned commit is empty, so this is green by emptiness and is **not** a pass for the future work; it becomes a measurement only after M1-M3).
- Known failing input: `E-013b` — `git diff --no-index --stat -G<same pattern>` between a template file and `premise-check.md` prints a stat (exit 1): the pattern fires on text that carries `t1894`, board ids and SHAs.
- Run-phase obligation (M4): re-run `E-013a` after M1-M3; a non-empty stat names the leaking file. The pattern avoids `{n}` quantifiers because the worktree guard refused them (Gap G-4).
- Mutants tried: a card id spelled `t 1894` (evades); a board id in a code span (fires).

### AC-MJ-014 — The pinned Go guards stay green after the change (G; REQ-MJ-013, -014, -016; run-phase only)

**Given** the run-phase tree after M1-M4,
**When** the scoped Go checks of `plan.md` §Verification run (one invocation per package; no `go test ./...`),
**Then** every test named below passes and each log is cited from the verdict file.

`internal/template`: `TestBindingLedgerIntegrity`, `TestRoleCoreStubGate`, `TestDeployedAlwaysLoadedCharBudget`, `TestRuleTemplateMirrorDrift`, `TestDeclaredRuleMirrorForks`, `TestReviewOwnership_CardReviewDoctrine`, `TestFactoryDispatchBodyCarriesRenewIntervalLinkage`, `TestLaneRecheckDoctrine`, `TestLaneWatchdogRenewsFirst`, `TestLaneWatchdogRecordsRefusalAndMeasures`, `TestWorkflowRulePathsPinned`, `TestTemplateNoInternalContentLeak`, the catalog hash audit (`TestCatalogHashCoversSkillSubfiles` and its neighbours in `catalog_tier_audit_test.go`). `internal/cli`: `TestAutoPickMirrorParity`, `TestAutoPickDocDoctrine`, `TestTodoSkillDocumentsClassification`, `TestAutoRankDoctrineAmendment`, `TestAutoRankMirrorParity`, `TestAutoRankMarkerDisclosure` (the last three are the guards that pin the old operator-only sentences and the passage the new pointer line sits in — `research.md` §6, §10).

- Disposition: not re-executable now (a Go run is not a single read-only invocation and none is allowed in this phase). It takes the undecidable disposition: it is a regression-guard, **not recorded as a pass** by this plan. Its RED is observable by the mutants above (for example the pointer-form K edit turns `TestRoleCoreStubGate` red).

### AC-MJ-015 — The lanes acknowledge receipt of the new guidance (O; operational)

**Given** the landed change, **When** the leader sends the lane notice (the changed rules, with the SessionStart notice caveat of spec R-1), **Then** each live lane replies or records that it read the rules before its next card.

- Not re-executable by a command (a lane's reading is not observable on disk). The leader records the acknowledgements in the card's progress record; they are not counted in this SPEC's PASS and the criterion is never "satisfied" by this SPEC's own commands.

### AC-MJ-016 — Each rule has exactly one home (R; REQ-MJ-012)

**Given** the eight changed template files,
**When** the files are listed that contain any of the six home tokens `slot-gated surface`, `no card is actually executing`, `only the card the lane is actually executing`, `expiry is harmless`, `cannot be added after assign`, `while the operator has delegated dispatch`,
**Then** exactly `factory-dispatch-cards.md`, `factory-dispatch-detail.md` and `factory-dispatch-mechanics.md` are listed.

- RED-now: `E-016` (no file listed, exit 1; tree `55d4f1486a`). Red because none of the home sentences exists yet (the set is empty, not the right three).
- Green path: M1-M3 → the three homes.
- Mutants tried: restate a home token in F, K or G — including the delegation condition in the new pointer lines (the extra file is listed → red); paraphrase a rule into a second file (evades — residual, read by the audit); put all rules in one file (only one listed → red).

### AC-MJ-017 — The declared fork of `factory-dispatch.md` stays exactly one differing line (G; REQ-MJ-013)

**Given** the two copies of `factory-dispatch.md`,
**When** `git diff --no-index --shortstat` compares them,
**Then** it prints ` 1 file changed, 1 insertion(+), 1 deletion(-)` (exit 1).

- Evidence now: `E-017`. The `1 insertion, 1 deletion` figure is a baseline measured on the pinned tree, not an absolute law: the criterion is "unchanged from the baseline measured at the start of the run phase". If another landing resolves the fork before this SPEC runs (the Go guard `TestAutoPickMirrorParity` would change with it), re-measure with the `E-017` command and re-pin before M1 (`verification-completeness.md` §4); never re-quote this number. Known failing input: any edit to one copy only changes the counts (observed by construction: the pair already differs by the worktree-sweep line).
- Mutants tried: apply the same edit to both → still `1/1` (correct); edit one side only → `2/2` or more (red).

### AC-MJ-018 — Each sentence that called promotion the operator's alone now carries the exception by pointer, and none restates it (R; REQ-MJ-017)

**Given** the leader's body F and the core stub K on both sides, and the todo skill G on both sides,
**When** (a) the exact pointer line of F is counted as a whole line, (b) the opening clause `The one reconciliation is named` is counted in F and K, (c) the replacement clause `A reconciliation is named, not excepted:` is counted in F and K, and (d) the exact pointer line of G is counted as a whole line,
**Then** (a) reports `1` per copy of F, (b) reports `0` for each of the four copies, (c) reports `1` for each of the four copies, and (d) reports `1` per copy of G.

- RED-now: `E-018a` (`:0`, `:0`, exit 1), `E-018b` (`:1` ×4, exit 0), `E-018c` (`:0` ×4, exit 1), `E-018d` (`:0`, `:0`, exit 1); tree `55d4f1486a`. Red because neither pointer line exists and the paragraph at F:44 / K:20 still opens by claiming a single reconciliation (`research.md` §10, P2-P6).
- Green path: M3 step 3 edits F (the pointer line after F:40; the clause at F:44), K (the clause at K:20; the size is AC-MJ-012's) and G (the pointer paragraph after the [HARD] pick rule), each template first → (a) `1`/`1`, (b) `0` ×4, (c) `1` ×4, (d) `1`/`1`.
- Mutants tried: pointer line in the template only → (a) `1`/`0`; pointer added but the clause left → (b) `1`; the clause edited in F and not in K → (b) `0`,`0`,`1`,`1` and (c) `1`,`1`,`0`,`0` (red; also `TestAutoRankMirrorParity`); the paragraph deleted instead of reworded → (c) `0` (red); a pointer line that restates the limits → (a) `0`, because the line differs, and `E-016` lists F when the restatement carries a home token; the pointer moved into K → (a) `0`, and AC-MJ-012 goes red on size.
- Does not establish: that no other sentence in the tree says the same thing in other words — `research.md` §10 is the measured list, and the sync audit reads for a paraphrase.

### AC-MJ-019 — The sentences the doctrine guards pin are still present (G; REQ-MJ-016, REQ-MJ-017)

**Given** F on both sides,
**When** the pinned prohibition sentence ("Outside an --auto authorization the leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item.") and the pinned admission sentence ("queue ADMISSION (production) stays the operator's.") are counted,
**Then** each reports `1` per copy.

- Evidence now: `E-019a` (`:1`, `:1`, exit 0) and `E-019c` (`:1`, `:1`, exit 0) on tree `55d4f1486a`. Green by design before any edit; it guards the run, because the natural wrong edit is to fold the exception into these sentences.
- Known failing input: `E-019b` — the prohibition pattern over the core stub, which does not carry that sentence, prints `0` (exit 1): the instrument goes red when the sentence is missing.
- Mutants tried: reword the sentence to carry the exception ("Outside an --auto authorization or while dispatch is delegated …") → `0` (red) and the Go doctrine tests of AC-MJ-014 go red too; delete it → `0`; move it to the detail companion → `0`.
- Does not establish: the two `absent` strings of `TestAutoPickDocDoctrine` ("Promotion is the operator's act, always.", "The leader never picks for the operator") stay absent. No command here carries a live control for an absence check, so that guard is left to AC-MJ-014's Go run (undecidable now).

## 2. Evidence ledger

```text
E-001
command: awk 'FNR==1{ if(NR>1) print a+b+c+d+e+0; a=b=c=d=e=0; f=0 } /^### Mode judgment/{f=1;next} /^##/{f=0} f&&/expected file set/{a=1} f&&/in-flight card/{b=1} f&&/slot-gated surface/{c=1} f&&/conflict check/{d=1} f&&/classification-file/{e=1} END{print a+b+c+d+e+0}' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md .claude/rules/moai/workflow/factory-dispatch-cards.md
stdout:
0
0
exit: 0
```

```text
E-002a   (sweep, working tree)
command: git grep -nP '^(?!\|).*moai (todo|gtd) add "(?!.*--classification-file)' -- internal/template/templates/.claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md internal/template/templates/.claude/agents/moai/manager-todo.md internal/template/templates/.claude/agents/moai/manager-lead.md internal/template/templates/.claude/skills/moai-factory-foreman/SKILL.md
stdout:
internal/template/templates/.claude/skills/moai/workflows/gtd.md:52:`moai todo add "<description>"`; a bare `/moai todo` runs `moai todo list`.
exit: 0
```

```text
E-002b   (control: same pattern, base commit)
command: git grep -nP '^(?!\|).*moai (todo|gtd) add "(?!.*--classification-file)' 74ad197380a0711cdb9816d4e37c94d29af730ea -- internal/template/templates/.claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md internal/template/templates/.claude/agents/moai/manager-todo.md internal/template/templates/.claude/agents/moai/manager-lead.md internal/template/templates/.claude/skills/moai-factory-foreman/SKILL.md
stdout:
74ad197380a0711cdb9816d4e37c94d29af730ea:internal/template/templates/.claude/skills/moai/workflows/gtd.md:52:`moai todo add "<description>"`; a bare `/moai todo` runs `moai todo list`.
exit: 0
```

```text
E-002c   (swept set)
command: git ls-files -- internal/template/templates/.claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md internal/template/templates/.claude/agents/moai/manager-todo.md internal/template/templates/.claude/agents/moai/manager-lead.md internal/template/templates/.claude/skills/moai-factory-foreman/SKILL.md
stdout:
internal/template/templates/.claude/agents/moai/manager-lead.md
internal/template/templates/.claude/agents/moai/manager-todo.md
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md
internal/template/templates/.claude/skills/moai-factory-foreman/SKILL.md
internal/template/templates/.claude/skills/moai/workflows/gtd.md
exit: 0
```

```text
E-002d   (positive presence on the todo skill)
command: grep -c -E '^[^|].*moai todo add ".*--classification-file' internal/template/templates/.claude/skills/moai/workflows/gtd.md .claude/skills/moai/workflows/gtd.md
stdout:
internal/template/templates/.claude/skills/moai/workflows/gtd.md:0
.claude/skills/moai/workflows/gtd.md:0
exit: 1
```

```text
E-003
command: grep -c -E '^\[HARD\].*--classification-file.*factory-dispatch-cards\.md' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md
stdout:
.claude/rules/moai/workflow/factory-dispatch.md:0
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:0
exit: 1
```

```text
E-004a
command: grep -c -F "cannot be added after assign" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md .claude/rules/moai/workflow/factory-dispatch-cards.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md:0
.claude/rules/moai/workflow/factory-dispatch-cards.md:0
exit: 1
```

```text
E-004b
command: grep -c -F 'An empty `spec_id` on a picked card is' internal/template/templates/.claude/skills/moai/workflows/gtd.md .claude/skills/moai/workflows/gtd.md
stdout:
internal/template/templates/.claude/skills/moai/workflows/gtd.md:1
.claude/skills/moai/workflows/gtd.md:1
exit: 0
```

```text
E-005a
command: grep -c -F "no card is actually executing" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md .claude/rules/moai/workflow/factory-dispatch-detail.md
stdout:
.claude/rules/moai/workflow/factory-dispatch-detail.md:0
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md:0
exit: 1
```

```text
E-005b
command: grep -c -F "free means the lane's previous card reached" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md .claude/rules/moai/workflow/factory-dispatch-detail.md
stdout:
.claude/rules/moai/workflow/factory-dispatch-detail.md:1
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md:1
exit: 0
```

```text
E-006   (the nine pinned lines are the g[1]..g[9] strings of the command; `\047` is the apostrophe of "leader's")
tree: 55d4f1486a454ff446be6f0f0645d4cea76bf0cc
command: awk 'BEGIN{ g[1]="Promotion remains an act of the operator. One further case is stated here, besides the `--auto` batch approval and a lane taking a lease itself: while the operator has delegated dispatch (operator away, or a standing delegation), the leader itself promotes queued cards to `picked` and assigns them, with no per-card operator pick. This section states every limit of that case, once."; g[2]="- Where: only to a free live lane, as defined under Routing above."; g[3]="- Which cards: only queued cards explicitly classified `parallelizable` by the mode judgment recorded at issuance (`factory-dispatch-cards.md` § Mode judgment)."; g[4]="- Never promoted: a keep-set card, a held card, or a [보류-marked card (the keep-set is defined in `auto-semantics.md` §9.3)."; g[5]="- Serial cards: outside this case; a serial card follows the slot rules (`factory-dispatch-mechanics.md` § The leader\047s remaining role)."; g[6]="- Order: queue priority and readiness first."; g[7]="- The leader acts by promotion to `picked` and assignment, nothing else."; g[8]="- Unchanged: lane self-dispatch through `moai factory next`."; g[9]="The assignment input of a free live lane is therefore the picked cards plus, only while dispatch is delegated, the queued `parallelizable` cards."; for(i=1;i<=9;i++) want[g[i]]=1 } FNR==1{ if(NR>1) print m" "n; m=n=0; f=0; split("",seen) } /^### Delegated-dispatch promotion$/{f=1;next} /^##/{f=0} f&&NF{ n++; if(($0 in want)&&!seen[$0]++) m++ } END{ print m" "n }' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md .claude/rules/moai/workflow/factory-dispatch-detail.md
stdout:
0 0
0 0
exit: 0
```

```text
E-007
command: awk 'FNR==1{ if(NR>1) print a+b+c+d+e+g+h+0; a=b=c=d=e=g=h=0; f=0 } /^### Lease renewal/{f=1;next} /^##/{f=0} f&&/only the card the lane is actually executing/{a=1} f&&/merge-ready/{b=1} f&&/held/{c=1} f&&/serial-waiting/{d=1} f&&/expiry is harmless/{e=1} f&&/notice/{g=1} f&&/factory next --card/{h=1} END{print a+b+c+d+e+g+h+0}' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md .claude/rules/moai/workflow/factory-dispatch-mechanics.md
stdout:
0
0
exit: 0
```

```text
E-005c
command: grep -c -F "not a free lane" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md .claude/rules/moai/workflow/factory-dispatch-detail.md
stdout:
.claude/rules/moai/workflow/factory-dispatch-detail.md:0
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md:0
exit: 1
```

```text
E-008a   (unscoped keep-to-completion lines; K, F, A, both trees)
command: grep -c -P '^(?!.*while the lane executes the card).*(kept|keeps|stays armed).*completion report' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md .claude/rules/moai/workflow/factory-dispatch-core.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md .claude/rules/moai/workflow/auto-semantics.md
stdout:
.claude/rules/moai/workflow/factory-dispatch.md:2
.claude/rules/moai/workflow/factory-dispatch-core.md:1
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md:1
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:2
.claude/rules/moai/workflow/auto-semantics.md:1
internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md:1
exit: 0
```

```text
E-008b   (scope marker present; K, F, A, M, both trees)
command: grep -c -F "while the lane executes the card" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md .claude/rules/moai/workflow/factory-dispatch-core.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md .claude/rules/moai/workflow/auto-semantics.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md .claude/rules/moai/workflow/factory-dispatch-mechanics.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md:0
.claude/rules/moai/workflow/factory-dispatch-core.md:0
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:0
internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md:0
.claude/rules/moai/workflow/auto-semantics.md:0
.claude/rules/moai/workflow/factory-dispatch.md:0
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md:0
.claude/rules/moai/workflow/factory-dispatch-mechanics.md:0
exit: 1
```

```text
E-008c   (renewal-period wording retained; regression-guard cell)
command: grep -c -F 'every renewal period of `workflow.factory_lease.renew_interval`' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md .claude/rules/moai/workflow/factory-dispatch-core.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md
stdout:
.claude/rules/moai/workflow/factory-dispatch-core.md:1
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md:1
.claude/rules/moai/workflow/factory-dispatch.md:1
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:1
exit: 0
```

```text
E-008d   (control for E-008a: same pattern, base commit)
command: git grep -c -P '^(?!.*while the lane executes the card).*(kept|keeps|stays armed).*completion report' 74ad197380a0711cdb9816d4e37c94d29af730ea -- internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md
stdout:
74ad197380a0711cdb9816d4e37c94d29af730ea:internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md:1
74ad197380a0711cdb9816d4e37c94d29af730ea:internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md:1
74ad197380a0711cdb9816d4e37c94d29af730ea:internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:2
exit: 0
```

```text
E-009
command: awk 'FNR==1{ if(NR>1) print a+b+0; a=b=0; f=0 } /^## 0a/{f=1;next} /^## /{f=0} f&&/executing/{a=1} f&&/merge-ready/{b=1} END{print a+b+0}' internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md .claude/skills/moai-lane-watchdog/SKILL.md
stdout:
0
0
exit: 0
```

```text
E-010a
command: git grep -c -E 'moai (todo|gtd|factory) reclassif|reclassify verb' -- internal/template/templates/.claude .claude
stdout:
exit: 1
```

```text
E-010b
command: grep -c -F 'Without it the deterministic default decider records `priority=normal, blocked=false, mode=serial, decider=default`' internal/template/templates/.claude/skills/moai/workflows/gtd.md .claude/skills/moai/workflows/gtd.md
stdout:
internal/template/templates/.claude/skills/moai/workflows/gtd.md:1
.claude/skills/moai/workflows/gtd.md:1
exit: 0
```

```text
E-010c   (known failing input for E-010a's pattern)
command: grep -c -E 'moai (todo|gtd|factory) reclassif|reclassify verb' .moai/reports/t1894/premise-check.md
stdout:
1
exit: 0
```

```text
E-011   (seven pairs; one cmp invocation each, run in one batch)
commands: cmp internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md .claude/rules/moai/workflow/factory-dispatch-cards.md
          cmp internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md .claude/rules/moai/workflow/factory-dispatch-core.md
          cmp internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md .claude/rules/moai/workflow/factory-dispatch-detail.md
          cmp internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md .claude/rules/moai/workflow/factory-dispatch-mechanics.md
          cmp internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md .claude/rules/moai/workflow/auto-semantics.md
          cmp internal/template/templates/.claude/skills/moai/workflows/gtd.md .claude/skills/moai/workflows/gtd.md
          cmp internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md .claude/skills/moai-lane-watchdog/SKILL.md
stdout: (empty for each)
exit: 0, 0, 0, 0, 0, 0, 0   (in the order above)
```

```text
E-011f   (known failing input: the declared fork)
command: cmp internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md differ: char 23127, line 222
exit: 1
```

```text
E-012a
command: python3 -c "import sys; n=len(open(sys.argv[1],encoding='utf-8').read().encode('utf-16-le'))//2; print('OVER' if n>10000 else 'OK', n)" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md
stdout:
OK 9999
exit: 0
```

```text
E-012b   (known failing input)
command: python3 -c "import sys; n=len(open(sys.argv[1],encoding='utf-8').read().encode('utf-16-le'))//2; print('OVER' if n>10000 else 'OK', n)" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md
stdout:
OVER 25803
exit: 0
```

```text
E-013a
command: git diff --stat -G'(\bt[0-9][0-9][0-9][0-9]?\b|SPEC-[A-Z][A-Z0-9-]*-[0-9][0-9][0-9]|\bd-2026[0-9]|20[0-9][0-9]-[0-9][0-9]-[0-9][0-9]|\b[0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]+\b|\b(a672|0c51|02f4)\b)' 74ad197380a0711cdb9816d4e37c94d29af730ea -- internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md internal/template/templates/.claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md
stdout: (empty)
exit: 0
```

```text
E-013b   (known failing input for E-013a's pattern)
command: git diff --no-index --stat -G'(\bt[0-9][0-9][0-9][0-9]?\b|SPEC-[A-Z][A-Z0-9-]*-[0-9][0-9][0-9]|\bd-2026[0-9]|20[0-9][0-9]-[0-9][0-9]-[0-9][0-9]|\b[0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]+\b|\b(a672|0c51|02f4)\b)' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md .moai/reports/t1894/premise-check.md
stdout:
 .../reports/t1894/premise-check.md                 | 289 ++-------------------
 1 file changed, 28 insertions(+), 261 deletions(-)
exit: 1
```

```text
E-016
tree: 55d4f1486a454ff446be6f0f0645d4cea76bf0cc
command: grep -l -F -e "slot-gated surface" -e "no card is actually executing" -e "only the card the lane is actually executing" -e "expiry is harmless" -e "cannot be added after assign" -e "while the operator has delegated dispatch" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md internal/template/templates/.claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md
stdout: (empty)
exit: 1
```

```text
E-017
command: git diff --no-index --shortstat internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md
stdout:
 1 file changed, 1 insertion(+), 1 deletion(-)
exit: 1
```

```text
E-018a   (the F pointer line, whole-line match; the line is the one F:42 will carry)
tree: 55d4f1486a454ff446be6f0f0645d4cea76bf0cc
command: grep -c -x -F '[HARD] **The delegated-dispatch exception.** Besides the `--auto` batch approval, the leader may promote queued cards in exactly one further case, stated once in `factory-dispatch-detail.md` § Delegated-dispatch promotion with the limits given there and no others; nothing in this section widens it.' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md
stdout:
.claude/rules/moai/workflow/factory-dispatch.md:0
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:0
exit: 1
```

```text
E-018b   (the old opening clause, F and K, both trees)
tree: 55d4f1486a454ff446be6f0f0645d4cea76bf0cc
command: grep -c -F 'The one reconciliation is named' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md .claude/rules/moai/workflow/factory-dispatch-core.md
stdout:
.claude/rules/moai/workflow/factory-dispatch.md:1
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:1
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md:1
.claude/rules/moai/workflow/factory-dispatch-core.md:1
exit: 0
```

```text
E-018c   (the replacement clause, F and K, both trees)
tree: 55d4f1486a454ff446be6f0f0645d4cea76bf0cc
command: grep -c -F 'A reconciliation is named, not excepted:' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md .claude/rules/moai/workflow/factory-dispatch-core.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:0
.claude/rules/moai/workflow/factory-dispatch-core.md:0
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md:0
.claude/rules/moai/workflow/factory-dispatch.md:0
exit: 1
```

```text
E-018d   (the G pointer line, whole-line match)
tree: 55d4f1486a454ff446be6f0f0645d4cea76bf0cc
command: grep -c -x -F 'Promotion by the leader without a pick is limited to the one case stated in `factory-dispatch-detail.md` § Delegated-dispatch promotion; every pick the operator makes follows the rule above.' internal/template/templates/.claude/skills/moai/workflows/gtd.md .claude/skills/moai/workflows/gtd.md
stdout:
internal/template/templates/.claude/skills/moai/workflows/gtd.md:0
.claude/skills/moai/workflows/gtd.md:0
exit: 1
```

```text
E-019a   (pinned prohibition sentence in F, both trees)
tree: 55d4f1486a454ff446be6f0f0645d4cea76bf0cc
command: grep -c -F 'Outside an --auto authorization the leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item.' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:1
.claude/rules/moai/workflow/factory-dispatch.md:1
exit: 0
```

```text
E-019b   (known failing input: the same sentence over the core stub, which does not carry it)
tree: 55d4f1486a454ff446be6f0f0645d4cea76bf0cc
command: grep -c -F 'Outside an --auto authorization the leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item.' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md
stdout:
0
exit: 1
```

```text
E-019c   (pinned admission sentence in F, both trees)
tree: 55d4f1486a454ff446be6f0f0645d4cea76bf0cc
command: grep -c -F "queue ADMISSION (production) stays the operator's." internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:1
.claude/rules/moai/workflow/factory-dispatch.md:1
exit: 0
```

## 3. Edge cases

- **A ledger row is stale after the rewrite.** If M2 rewrites F:119 and leaves ALB-0276 untouched, `TestBindingLedgerIntegrity` reports "after-text absent … obligation lost"; the row's `after_text` is updated in the same commit (REQ-MJ-016).
- **The core stub exceeds 10,000 units by a few characters.** AC-MJ-012 prints `OVER`; the marker-form wording, the −6 clause edit and a trim of at least 2 units resolve it (`design.md` §D-4).
- **The exception is folded into a pinned sentence.** Rewording F:38, F:40 or the K:20 / F:44 paragraph's pinned clauses to carry the exception turns AC-MJ-019 red and the Go doctrine tests of AC-MJ-014 red; the exception is added beside them, by pointer.
- **A lane or leader treats an unstated delegation as given.** The ruling names no mechanism that establishes "the operator has delegated dispatch" (`spec.md` R-5); AC-MJ-006 pins the text and cannot see the practice.
- **`factory-dispatch.md` edited identically in both trees but the declared fork line also moves.** AC-MJ-017 shows a count other than `1/1`.
- **The per-line trap.** A scoped sentence whose physical tail line still reads "kept until the completion report" keeps AC-MJ-008(a) red; reflow the paragraph.

## 4. Quality gates and Definition of Done

- All `R` criteria (AC-MJ-001..009, -016, -018) have observed RED cells (above) and pass on the run-phase tree with the command output pasted into the verdict file `.moai/reports/t1894/verdict.md` (decision-bearing lines only, per the evidence-export obligation).
- All `G` criteria (AC-MJ-010..014, -017, -019) pass; AC-MJ-014's Go logs are cited from the verdict file.
- AC-MJ-015 is recorded by the leader and is not a SPEC PASS.
- Decision-index rows Q1, Q2, Q4, Q5 are stamped `DEFAULT-APPLIED`; Q3 is met by the operator ruling d314 (its verdict is recorded in the row); Q6 is POLICY-COVERED. No requirement or criterion is gated on a decision.
- Template-First order held: every template edit precedes the deployed edit (visible in the commit order of M1-M3).
- `moai spec lint --strict` on `spec.md` exits 0 (recorded in `progress.md` §E.1).
