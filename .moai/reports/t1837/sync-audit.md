auditor-model: claude-sonnet-5-5

# Card t1837 — sync audit (independent)

verdict: PASS
audited_sha: 1519e9ce1a8c0ce69e0a83bb5005e1f9b6179365

Result: PASS, score 0.93 (round 1 at 2d649e6a2; round 2 delta audit at 1519e9ce1 confirms it unchanged).

Dimension notes:
- Functionality (40%): PASS. Card parts (a) and (b) are present, and the leader constraints (examples, not policy; rationale quoted; Template-First) hold.
- Security (25%): PASS. A documentation-only change with no code, secrets or inputs. The new rule tightens a governance gate: a card that changes `.github/workflows/` waits for the operator before landing.
- Craft (20%): PASS. The two mirror tests passed with a non-zero swept count, in both rounds. Optional wording nits are listed below.
- Consistency (15%): PASS. The mirrors are byte-identical and the 2b step precedes Deliver. No other factory-dispatch file contradicts it, and no rule file changed on main in between.

Per-backend verdicts (audit plan: codex gate=required, claude off, glm off):
- codex (adversarial, `mcp__moai__codex_audit`, project_root = worktree): `pass`, findings [], in every run. The required backend returned a real verdict each time.
- Receipts: round 1 `rcpt-924b04e90907759ba72bb20a` (HEAD 2d649e6a2); re-runs for the stop-hook corroboration `rcpt-90a0b2f81374473706b304aa` and `rcpt-6b8d3a195669ac910b8f0f10` (HEAD dc47a3b85, same card content); round 2 `rcpt-31695f1bc1fb9aa3fabea86e` (HEAD 1519e9ce1).
- claude: gate off, not run. glm: gate off, not run.
- audit_multi was not called in either round. The single required backend was reached directly through codex_audit, so the earlier audit_multi convergence check was not run.
- Result digest for the orchestrator's check: overall_verdict = pass (codex), gate_unmet = none observed, plan_source = config.gates. The verdict is not final until the orchestrator writes these into `<toplevel>/.moai/state/audit-plan-result.json` and runs `moai verify audit-plan --result-file`.

## Iteration history
- Round 1 — audited 2d649e6a2c2a63ea6e0ac826bc05eecd5a62d8aa (base fb749d372): PASS 0.93. Full text and evidence are below, retained unchanged except where marked round 1.
- Round 2 — delta audit at 1519e9ce1a8c0ce69e0a83bb5005e1f9b6179365. Why: main advanced 82 commits (cards t1852, t1599 and others) after round 1, and the coordinator absorbed it with a no-ff merge. The factory gate binds the recorded evidence to the verdict's `audited_sha`, so the audit was re-read at the new head. What I verified, read-only:
  - HEAD is 1519e9ce1a8c0ce69e0a83bb5005e1f9b6179365 and `git status --porcelain` was empty at the start.
  - `git diff --name-only main...HEAD` lists exactly seven files: the two docs in both mirror locations plus `.moai/reports/t1837/{card-review,sync-audit,verdict}.md`.
  - `git diff 2d649e6a2 HEAD -- <the four doc files>` printed nothing, so the four docs are byte-identical to what round 1 audited. The two mirror pairs are still `cmp`-identical.
  - `git grep -n '5b\|2b\.'` over both rule trees: only the `2b.` step itself (mechanics.md:137 in each tree). No live citation of 5b.
  - `git diff 2d649e6a2 HEAD --stat -- .claude/rules internal/template/templates/.claude/rules` printed nothing. Main's 82 commits changed no rule file in either tree, so none touches the landing sequence or the card classes.
  - Narrow mirror tests re-run once: PASS, 50 subtests, both touched files among them.
  - Required codex backend re-run: `pass`, receipt `rcpt-31695f1bc1fb9aa3fabea86e`.
  - Outcome: verdict and findings are unchanged.

## Claim
1. Part (a) is satisfied by `factory-dispatch-cards.md` § "Card size ceiling — split before dispatch". Part (b) is satisfied by § "Workflow-file changes need operator review", landing step 2b and the leader table row.
2. Step 2b sits between Card-review (2) and Deliver (3). Nothing in the factory-dispatch family contradicts it, and nothing cites the old label 5b.
3. The mirrors are byte-identical, no always-loaded file changed, and the diff holds no unrelated file (still true at 1519e9ce1).
4. The narrow tests pass and the swept count is non-zero.
5. The required codex backend returned `pass`.
6. (Round 2) The four doc files at 1519e9ce1 equal the round-1 audited content, and main's absorbed commits changed no rule file.

## Evidence
Quoted text (factory-dispatch-cards.md):
- "Rationale as the card states it: failed agent PRs are larger and touch more files (arXiv:2601.15195); DORA small batches; the Copilot coding agent's workflow-run approval default."
- "Whoever issues or classifies a card states its expected size — files touched and lines changed — in the card text. The operative rule is: **when the work exceeds the stated estimate, split the card before dispatch** ... As an illustration only, a card expected to touch more than about 10 files or 400 changed lines is a candidate for splitting; the numbers are examples, not a policy threshold."
- "A card whose agent-authored branch changes any file under `.github/workflows/` is marked **operator review required before landing**. Detect it with `git diff --name-only <base>...HEAD -- .github/workflows` ... the lane's verdict text states the marking, the lane does NOT self-land it (it stops before Deliver ...) and the leader's checklist lists the card until the operator has reviewed it."

Quoted text (factory-dispatch-mechanics.md, lines 137-141 and 182):
- "2b. **Workflow-file check** — runs before Deliver. A branch changing `.github/workflows/**` is "operator review required before landing": the lane stops here (no `moai factory complete`, no PR) and reports to the leader".
- Line 142 is "3. **Deliver** — `moai factory complete`".
- Line 182 is the leader table row "Workflow-file review checklist | cards marked "operator review required before landing" ... wait on the operator; the leader lists them".

Round 1 commands and observed output (tree 2d649e6a2):
- `git rev-parse --show-toplevel` -> `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1837`
- `git rev-parse HEAD` -> `2d649e6a2c2a63ea6e0ac826bc05eecd5a62d8aa`
- `git diff --name-only fb749d372e61a6b10c53091362c76b4412827427...HEAD` ->
  ```
  .claude/rules/moai/workflow/factory-dispatch-cards.md
  .claude/rules/moai/workflow/factory-dispatch-mechanics.md
  .moai/reports/t1837/card-review.md
  .moai/reports/t1837/verdict.md
  internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md
  internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md
  ```
  `factory-dispatch-core.md`, `factory-dispatch.md`, AGENTS.md, output styles and MEMORY.md are not in the list, so no always-loaded file changed.
- `cmp` of both mirror pairs printed `IDENT1` and `IDENT2`. The template diff is +10 lines in cards.md and +6 lines in mechanics.md, all additions.
- Grep over `factory-dispatch.md`, `-detail.md`, `-core.md`, `-cards.md` and `-mechanics.md` for `5b|2b|workflow-file|operator review required|Deliver`: hits only in cards.md:44 and mechanics.md:137,138,142,182. The three other files have zero hits, so they neither contradict 2b nor mention it. The detail file's card-review stage list (run-exit / card-review / integration / report) does not describe Deliver.
- `grep -rnE "5b\b|step 5b|5b\."` over both rule trees and `.moai/reports/t1837` -> one hit, in `card-review.md:13`, which is history.
- `git diff --name-only fb749d372...HEAD -- .github` -> empty, so the lane's "not marked" claim holds.
- Narrow tests (env-scrubbed compound command, TMPDIR/GOTMPDIR under `/private/tmp`):
  ```
  --- PASS: TestRuleTemplateMirrorDrift/factory-dispatch-mechanics.md (0.00s)
  --- PASS: TestRuleTemplateMirrorDrift/factory-dispatch-cards.md (0.00s)
  PASS
  ok  	github.com/modu-ai/moai-adk/internal/template	0.265s
  ```
  Counts: `--- PASS: TestDeclaredRuleMirrorForks (0.00s)`, `--- PASS: TestRuleTemplateMirrorDrift (0.00s)`, and 50 `TestRuleTemplateMirrorDrift/` subtests.
- `moai verify audit-plan --project-root <toplevel>` -> `config_status: ok`, codex gate `required`, claude `off`, glm `off`, `enforced_required: ["codex"]`.
- `mcp__moai__codex_audit` -> `{"verdict":"pass","summary":"Verdict: pass","findings":[], ... "audit_receipt":"rcpt-924b04e90907759ba72bb20a","build_commit":"ee981a932"}`.

Round 2 commands and observed output (tree 1519e9ce1):
- `git rev-parse HEAD` -> `1519e9ce1a8c0ce69e0a83bb5005e1f9b6179365`; `git status --porcelain` -> empty.
- `git diff --name-only main...HEAD` ->
  ```
  .claude/rules/moai/workflow/factory-dispatch-cards.md
  .claude/rules/moai/workflow/factory-dispatch-mechanics.md
  .moai/reports/t1837/card-review.md
  .moai/reports/t1837/sync-audit.md
  .moai/reports/t1837/verdict.md
  internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md
  internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md
  ```
- `git diff 2d649e6a2 HEAD -- <the four doc files>` -> no output.
- `cmp` of both mirror pairs -> `IDENT1`, `IDENT2`.
- `git grep -n '5b\|2b\.' -- .claude/rules/moai/workflow internal/template/templates/.claude/rules/moai/workflow` ->
  ```
  .claude/rules/moai/workflow/factory-dispatch-mechanics.md:137:2b. **Workflow-file check** — runs before Deliver. A branch changing
  internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md:137:2b. **Workflow-file check** — runs before Deliver. A branch changing
  ```
- `git diff 2d649e6a2 HEAD --stat -- .claude/rules internal/template/templates/.claude/rules` -> no output. The whole-tree stat from 2d649e6a2 shows 58 files changed (4213 insertions, 84 deletions), none under either rules tree. The changes are docs, reports, specs, `internal/cli`, `internal/homestate`, `internal/hook`, scripts and similar; the t1837 reports rows are the coordinator's commit of round 1's sync-audit.md and the verdict.md note.
- Mirror tests (env scrubbed: `unset CLAUDE_PROJECT_DIR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_CONFIG_GLOBAL MOAI_SESSION_ID && TMPDIR=/private/tmp/sa-run2/tmp GOTMPDIR=/private/tmp/sa-run2/gotmp go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift|TestDeclaredRuleMirrorForks' -count=1 -v`):
  ```
  --- PASS: TestDeclaredRuleMirrorForks (0.00s)
  --- PASS: TestRuleTemplateMirrorDrift (0.00s)
  PASS
  ok  	github.com/modu-ai/moai-adk/internal/template	0.177s
      --- PASS: TestRuleTemplateMirrorDrift/factory-dispatch-mechanics.md (0.00s)
      --- PASS: TestRuleTemplateMirrorDrift/factory-dispatch-cards.md (0.00s)
  ```
  Swept count: 50 `TestRuleTemplateMirrorDrift/` subtests. The go-test-heavy slot was not taken and the full package was not run.
- `mcp__moai__codex_audit` (adversarial, baseBranch, project_root = worktree) -> `{"verdict":"pass","summary":"Verdict: pass","findings":[], ... "audit_receipt":"rcpt-31695f1bc1fb9aa3fabea86e","build_commit":"ee981a932"}`.

## Baseline-attribution
- Round 1 tree: HEAD `2d649e6a2c2a63ea6e0ac826bc05eecd5a62d8aa`, branch WT-card-size-workflow-review, base `fb749d372e61a6b10c53091362c76b4412827427`.
- Round 2 tree: HEAD `1519e9ce1a8c0ce69e0a83bb5005e1f9b6179365` (no-ff merge of main into the card branch), measured in this run with a clean working tree at the start.
- Judging build: `moai version` -> `v3.2.0-rc.36-t1822c  moai_cp/20261007_195927-1329-gee981a932-dirty  built 2026-10-15T21:56:00Z`, commit `ee981a932`. codex_audit reported `build_lag` in each run: ee981a932 is an ancestor of the tree HEAD, so the installed binary is older than the tree. The Go tests were built from the tree by `go test`. Binary-lag exposure applies only to the CLI and MCP results, and neither depends on the new text.
- The lane's measurements (`make build`, the 360 s full-package run, and the 2-test pass) are claims. I re-observed only the narrow tests and the byte comparisons myself.

## Gaps
- No CI or CodeRabbit observed (no PR exists yet).
- The landing re-measure is run by the lane right after this audit; I did not observe it.
- The full `./internal/template/` package was not run in either round (instructed not to; another lane holds the slot). The lane's 360 s pass predates the 1969709ce fix and is unobserved by me. After absorbing main, the package has not been re-run by me; Go and script changes from main (58 files, none in the rules trees) were not independently reviewed here beyond confirming that no rule file changed.
- audit_multi was not called in either round; only the required codex leg was run, and the claude and glm legs are gate off, so there is no cross-model disagreement signal. The earlier audit_multi convergence check was not run.
- The codex review is static and ran no tests.
- The detection command was observed only as an empty result on this branch. A positive case (a branch that does change a workflow file) was not exercised.
- The lane's slot-violation gap (an unslotted 360 s run while lane-5 held go-test-heavy) is recorded in verdict.md. I did not re-adjudicate it.
- The installed moai binary is older than the tree HEAD (see Baseline-attribution).

## Residual-risk
- The "about 10 files / 400 lines" figures could still be read as a threshold by a lane. The text labels them "illustration only ... not a policy threshold", so the risk is wording-level.
- Nothing mechanical enforces either rule. They are doctrine, and the leader's checklist is a human or agent practice. The cards text hands the lane a stop-and-report duty with no hook behind it.
- `2b` is a non-integer insertion in a numbered list. It is cosmetic and clear in context.
- A later main advance would again move the head; the delta check above covers 1519e9ce1 only.

## Findings
- F1 [low] [optional] internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md:182 — The leader table row says marked cards "wait on the operator" without naming that the lane stops before Deliver. It is consistent with step 2b and the cards section, but not self-contained (the lane's own residual-risk note lists the same point). Required fix: none necessary; optionally add "(stopped before Deliver)" to the row.
- F2 [low] [optional] internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md:137 — The label `2b` is a non-integer step in a numbered list. Required fix: none necessary; a renumber would also change the Close and Sweep references, so it is not worth the churn.
- F3 [low] [optional] .claude/rules/moai/workflow/factory-dispatch-cards.md:38 — The size-estimate duty is stated without a hook or evidence field that shows whether the estimate was recorded. This is a continued-firing gap under verification-completeness §1.3, but the card asked only for guidance wording. Required fix: none for this card; a follow-up may add a completion-read check if desired.

No blocking findings.

AUDIT-VERDICT: PASS spec=card-t1837 receipts=rcpt-924b04e90907759ba72bb20a,rcpt-90a0b2f81374473706b304aa,rcpt-6b8d3a195669ac910b8f0f10,rcpt-31695f1bc1fb9aa3fabea86e
