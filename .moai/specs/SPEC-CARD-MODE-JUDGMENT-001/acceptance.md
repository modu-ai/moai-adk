# acceptance.md — SPEC-CARD-MODE-JUDGMENT-001

The verification layer. Requirements (GEARS) are in `spec.md` §4; this file holds Given-When-Then criteria, each with an evidence-ledger entry. A GEARS requirement is never restated here and a Given-When-Then scenario is never a requirement.

## 0. Conventions

- **Document-level pin** (binds every ledger entry that carries no pin of its own, `verification-completeness.md` §2.1): tree SHA `74ad197380a0711cdb9816d4e37c94d29af730ea`, branch `WT-fail-safe-serial` (clean, unpushed). Every command ran with cwd `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894`; the judging `moai` build for CLI observations is `moai_cp/20261007_195927`. No Go command ran in this phase.
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

### AC-MJ-006 — The assignment input is defined, subject to verdict Q3 (R; REQ-MJ-007; blocked on decision-index row Q3)

**Given** D on both sides, **When** the literal `assignment input` is counted, **Then** each copy reports `1`.

- RED-now: `E-006` (`:0`, `:0`, exit 1). Red because no document defines the input to assignment.
- Green path: M3, *after* the operator verdict for Q3 — the sentence is not written under a planner default. If the verdict is the Alternate (the leader may promote queued parallelizable cards), this criterion and REQ-MJ-007 are rewritten by amendment before M3 runs; they are not silently satisfied.
- Mutants tried: the term in a comment (passes; read by the audit); the sentence stating that the leader promotes queued cards without the verdict (passes the command, violates the open item — caught by the Q3 row staying empty).

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

- Evidence now: `E-012a` prints `OK 9999` (headroom 1). This is not vacuous: the work must edit this file (site S10) and the candidate wording adds 9 units (`design.md` §D-4).
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

`internal/template`: `TestBindingLedgerIntegrity`, `TestRoleCoreStubGate`, `TestDeployedAlwaysLoadedCharBudget`, `TestRuleTemplateMirrorDrift`, `TestDeclaredRuleMirrorForks`, `TestReviewOwnership_CardReviewDoctrine`, `TestFactoryDispatchBodyCarriesRenewIntervalLinkage`, `TestLaneRecheckDoctrine`, `TestLaneWatchdogRenewsFirst`, `TestLaneWatchdogRecordsRefusalAndMeasures`, `TestWorkflowRulePathsPinned`, `TestTemplateNoInternalContentLeak`, the catalog hash audit (`TestCatalogHashCoversSkillSubfiles` and its neighbours in `catalog_tier_audit_test.go`). `internal/cli`: `TestAutoPickMirrorParity`, `TestAutoPickDocDoctrine`, `TestTodoSkillDocumentsClassification`.

- Disposition: not re-executable now (a Go run is not a single read-only invocation and none is allowed in this phase). It takes the undecidable disposition: it is a regression-guard, **not recorded as a pass** by this plan. Its RED is observable by the mutants above (for example the pointer-form K edit turns `TestRoleCoreStubGate` red).

### AC-MJ-015 — The lanes acknowledge receipt of the new guidance (O; operational)

**Given** the landed change, **When** the leader sends the lane notice (the changed rules, with the SessionStart notice caveat of spec R-1), **Then** each live lane replies or records that it read the rules before its next card.

- Not re-executable by a command (a lane's reading is not observable on disk). The leader records the acknowledgements in the card's progress record; they are not counted in this SPEC's PASS and the criterion is never "satisfied" by this SPEC's own commands.

### AC-MJ-016 — Each rule has exactly one home (R; REQ-MJ-012)

**Given** the eight changed template files,
**When** the files are listed that contain any of the five home tokens `slot-gated surface`, `no card is actually executing`, `only the card the lane is actually executing`, `expiry is harmless`, `cannot be added after assign`,
**Then** exactly `factory-dispatch-cards.md`, `factory-dispatch-detail.md` and `factory-dispatch-mechanics.md` are listed.

- RED-now: `E-016` (no file listed, exit 1). Red because none of the home sentences exists yet (the set is empty, not the right three).
- Green path: M1-M3 → the three homes.
- Mutants tried: restate a home token in F or K (the extra file is listed → red); paraphrase a rule into a second file (evades — residual, read by the audit); put all rules in one file (only one listed → red).

### AC-MJ-017 — The declared fork of `factory-dispatch.md` stays exactly one differing line (G; REQ-MJ-013)

**Given** the two copies of `factory-dispatch.md`,
**When** `git diff --no-index --shortstat` compares them,
**Then** it prints ` 1 file changed, 1 insertion(+), 1 deletion(-)` (exit 1).

- Evidence now: `E-017`. The `1 insertion, 1 deletion` figure is a baseline measured on the pinned tree, not an absolute law: the criterion is "unchanged from the baseline measured at the start of the run phase". If another landing resolves the fork before this SPEC runs (the Go guard `TestAutoPickMirrorParity` would change with it), re-measure with the `E-017` command and re-pin before M1 (`verification-completeness.md` §4); never re-quote this number. Known failing input: any edit to one copy only changes the counts (observed by construction: the pair already differs by the worktree-sweep line).
- Mutants tried: apply the same edit to both → still `1/1` (correct); edit one side only → `2/2` or more (red).

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
E-006
command: grep -c -F "assignment input" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md .claude/rules/moai/workflow/factory-dispatch-detail.md
stdout:
.claude/rules/moai/workflow/factory-dispatch-detail.md:0
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md:0
exit: 1
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
command: grep -l -F -e "slot-gated surface" -e "no card is actually executing" -e "only the card the lane is actually executing" -e "expiry is harmless" -e "cannot be added after assign" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md internal/template/templates/.claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md
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

## 3. Edge cases

- **A ledger row is stale after the rewrite.** If M2 rewrites F:119 and leaves ALB-0276 untouched, `TestBindingLedgerIntegrity` reports "after-text absent … obligation lost"; the row's `after_text` is updated in the same commit (REQ-MJ-016).
- **The core stub exceeds 10,000 units by a few characters.** AC-MJ-012 prints `OVER`; the marker-form wording plus an 8-unit offset resolves it (`design.md` §D-4).
- **The operator's Q3 verdict is the Alternate.** AC-MJ-006 and REQ-MJ-007 are rewritten by amendment before M3 runs.
- **`factory-dispatch.md` edited identically in both trees but the declared fork line also moves.** AC-MJ-017 shows a count other than `1/1`.
- **The per-line trap.** A scoped sentence whose physical tail line still reads "kept until the completion report" keeps AC-MJ-008(a) red; reflow the paragraph.

## 4. Quality gates and Definition of Done

- All `R` criteria (AC-MJ-001..009, -016) have observed RED cells (above) and pass on the run-phase tree with the command output pasted into the verdict file `.moai/reports/t1894/verdict.md` (decision-bearing lines only, per the evidence-export obligation).
- All `G` criteria (AC-MJ-010..014, -017) pass; AC-MJ-014's Go logs are cited from the verdict file.
- AC-MJ-015 is recorded by the leader and is not a SPEC PASS.
- Decision-index rows Q1, Q2, Q4, Q5 are stamped `DEFAULT-APPLIED`; Q3 carries the operator verdict before M3; Q6 is POLICY-COVERED.
- Template-First order held: every template edit precedes the deployed edit (visible in the commit order of M1-M3).
- `moai spec lint --strict` on `spec.md` exits 0 (recorded in `progress.md` §E.1).
