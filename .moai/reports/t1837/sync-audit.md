auditor-model: claude-sonnet-5-5

# Card t1837 — sync audit (independent)

verdict: PASS
audited_sha: 2d649e6a2c2a63ea6e0ac826bc05eecd5a62d8aa

Result: PASS, score 0.93.

Dimension notes:
- Functionality (40%): PASS. Card parts (a) and (b) are present, and the leader constraints (examples, not policy; rationale quoted; Template-First) hold.
- Security (25%): PASS. A documentation-only change with no code, secrets or inputs. The new rule tightens a governance gate: a card that changes `.github/workflows/` waits for the operator before landing.
- Craft (20%): PASS. The two mirror tests passed with a non-zero swept count. Optional wording nits are listed below.
- Consistency (15%): PASS. The mirrors are byte-identical and the 2b step precedes Deliver. No other factory-dispatch file contradicts it.

Per-backend verdicts (audit plan: codex gate=required, claude off, glm off):
- codex (adversarial, `mcp__moai__codex_audit`, project_root = worktree): verdict `pass`, findings [], audit_receipt `rcpt-924b04e90907759ba72bb20a`. The required backend returned a real verdict.
- claude: gate off, not run. glm: gate off, not run.
- audit_multi was not called. The single required backend was reached directly through codex_audit.
- Result digest for the orchestrator's check: overall_verdict = pass (codex), gate_unmet = none observed, plan_source = config.gates. The verdict is not final until the orchestrator writes these into `<toplevel>/.moai/state/audit-plan-result.json` and runs `moai verify audit-plan --result-file`.

## Claim
1. Part (a) is satisfied by `factory-dispatch-cards.md` § "Card size ceiling — split before dispatch". Part (b) is satisfied by § "Workflow-file changes need operator review", landing step 2b and the leader table row.
2. Step 2b sits between Card-review (2) and Deliver (3). Nothing in the factory-dispatch family contradicts it, and nothing cites the old label 5b.
3. The mirrors are byte-identical, no always-loaded file changed, and the diff holds no unrelated file.
4. The narrow tests pass and the swept count is non-zero.
5. The required codex backend returned `pass`.

## Evidence
Quoted text (factory-dispatch-cards.md):
- "Rationale as the card states it: failed agent PRs are larger and touch more files (arXiv:2601.15195); DORA small batches; the Copilot coding agent's workflow-run approval default."
- "Whoever issues or classifies a card states its expected size — files touched and lines changed — in the card text. The operative rule is: **when the work exceeds the stated estimate, split the card before dispatch** ... As an illustration only, a card expected to touch more than about 10 files or 400 changed lines is a candidate for splitting; the numbers are examples, not a policy threshold."
- "A card whose agent-authored branch changes any file under `.github/workflows/` is marked **operator review required before landing**. Detect it with `git diff --name-only <base>...HEAD -- .github/workflows` ... the lane's verdict text states the marking, the lane does NOT self-land it (it stops before Deliver ...) and the leader's checklist lists the card until the operator has reviewed it."

Quoted text (factory-dispatch-mechanics.md, lines 137-141 and 182):
- "2b. **Workflow-file check** — runs before Deliver. A branch changing `.github/workflows/**` is "operator review required before landing": the lane stops here (no `moai factory complete`, no PR) and reports to the leader".
- Line 142 is "3. **Deliver** — `moai factory complete`".
- Line 182 is the leader table row "Workflow-file review checklist | cards marked "operator review required before landing" ... wait on the operator; the leader lists them".

Commands run and observed output:
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
  Only the two docs in both mirror locations plus two lane evidence files appear. `factory-dispatch-core.md`, `factory-dispatch.md`, AGENTS.md, output styles and MEMORY.md are not in the list, so no always-loaded file changed.
- `cmp internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md .claude/rules/moai/workflow/factory-dispatch-cards.md && echo IDENT1` -> `IDENT1`. The same command for `factory-dispatch-mechanics.md` printed `IDENT2`.
- The diff of the template tree is +10 lines in cards.md and +6 lines in mechanics.md. All lines are additions.
- Ordering and contradiction grep over `factory-dispatch.md`, `-detail.md`, `-core.md`, `-cards.md` and `-mechanics.md` for `5b|2b|workflow-file|operator review required|Deliver`: hits only in cards.md:44 and mechanics.md:137,138,142,182. `factory-dispatch.md`, `-detail.md` and `-core.md` have zero hits, so they neither contradict 2b nor mention it. The detail file's card-review stage list (run-exit / card-review / integration / report) does not describe Deliver, so there is no conflict.
- `grep -rnE "5b\b|step 5b|5b\."` over both rule trees and `.moai/reports/t1837` -> a single hit, in `card-review.md:13`, which is history ("Step `5b` removed and reinserted as `2b`"). No live citation of 5b.
- `git diff --name-only fb749d372...HEAD -- .github` -> empty. The lane's "not marked" claim holds, and this is the same detection command form that the new text prescribes.
- Narrow tests (first run compound, env-scrubbed: `unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_CONFIG_GLOBAL CLAUDE_PROJECT_DIR && TMPDIR=/private/tmp/sa-run1/tmp GOTMPDIR=/private/tmp/sa-run1/gotmp go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift|TestDeclaredRuleMirrorForks' -count=1 -v`):
  ```
  --- PASS: TestRuleTemplateMirrorDrift/factory-dispatch-mechanics.md (0.00s)
  --- PASS: TestRuleTemplateMirrorDrift/factory-dispatch-cards.md (0.00s)
  PASS
  ok  	github.com/modu-ai/moai-adk/internal/template	0.265s
  ```
  Re-run for counts:
  ```
  === RUN   TestDeclaredRuleMirrorForks
  === RUN   TestRuleTemplateMirrorDrift
  --- PASS: TestDeclaredRuleMirrorForks (0.00s)
  --- PASS: TestRuleTemplateMirrorDrift (0.00s)
  PASS
  ok  	github.com/modu-ai/moai-adk/internal/template	0.266s
  ```
  `grep -c -- "--- PASS: TestRuleTemplateMirrorDrift/"` -> `50` (non-zero swept subtests, including both touched files). The temp dirs are under `/private/tmp`. The go-test-heavy slot was not touched, and the full package was not run.
- `moai verify audit-plan --project-root <toplevel>` -> `config_status: ok`, codex gate `required`, claude `off`, glm `off`, `enforced_required: ["codex"]`.
- `mcp__moai__codex_audit` (mode adversarial, target baseBranch, project_root = worktree) -> `{"verdict":"pass","summary":"Verdict: pass","findings":[],"next_steps":[],"audit_receipt":"rcpt-924b04e90907759ba72bb20a","build_commit":"ee981a932", ...}`.

## Baseline-attribution
- Tree measured in this run: HEAD `2d649e6a2c2a63ea6e0ac826bc05eecd5a62d8aa`, branch WT-card-size-workflow-review, base `fb749d372e61a6b10c53091362c76b4412827427`.
- Judging build: `moai version` -> `v3.2.0-rc.36-t1822c  moai_cp/20261007_195927-1329-gee981a932-dirty  built 2026-10-15T21:56:00Z`, commit `ee981a932`. codex_audit reported `build_lag`: ee981a932 is an ancestor of HEAD 2d649e6a2, so the installed binary is older than the tree. The finished change is a rule-markdown edit and the Go tests were built from the tree by `go test`. Binary-lag exposure applies only to the `moai verify audit-plan` and MCP results, and neither depends on the new text.
- The lane's measurements (`make build`, the 360 s full-package run, and the 2-test pass) are claims. I re-observed only the narrow tests and the byte comparisons myself.

## Gaps
- The full `./internal/template/` package was not run (instructed not to; another lane holds the slot). The lane's 360 s pass predates the 1969709ce fix and is unobserved by me.
- audit_multi was not called. Only the required codex leg was run, and the claude and glm legs are gate off, so there is no cross-model disagreement signal.
- The codex review is static and ran no tests.
- CI and CodeRabbit are not observed (no PR exists).
- The detection command was observed only as an empty result on this branch. A positive case (a branch that does change a workflow file) was not exercised.
- The lane's slot-violation gap (an unslotted 360 s run while lane-5 held go-test-heavy) is recorded in verdict.md. I did not re-adjudicate it; it affects lane-5's timings, not this card's correctness.
- The installed moai binary is older than the tree HEAD (see Baseline-attribution).

## Residual-risk
- The "about 10 files / 400 lines" figures could still be read as a threshold by a lane. The text labels them "illustration only ... not a policy threshold", so the risk is wording-level.
- Nothing mechanical enforces either rule. They are doctrine, and the leader's checklist is a human or agent practice. The cards text hands the lane a stop-and-report duty with no hook behind it.
- `2b` is a non-integer insertion in a numbered list. It is cosmetic and clear in context.

## Findings
- F1 [low] [optional] internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md:182 — The leader table row says marked cards "wait on the operator" without naming that the lane stops before Deliver. It is consistent with step 2b and the cards section, but not self-contained (the lane's own residual-risk note lists the same point). Required fix: none necessary; optionally add "(stopped before Deliver)" to the row.
- F2 [low] [optional] internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md:137 — The label `2b` is a non-integer step in a numbered list. Required fix: none necessary; a renumber would also change the Close and Sweep references, so it is not worth the churn.
- F3 [low] [optional] .claude/rules/moai/workflow/factory-dispatch-cards.md:38 — The size-estimate duty is stated without a hook or evidence field that shows whether the estimate was recorded. This is a continued-firing gap under verification-completeness §1.3, but the card asked only for guidance wording. Required fix: none for this card; a follow-up may add a completion-read check if desired.

No blocking findings.

AUDIT-VERDICT: PASS spec=card-t1837 receipts=rcpt-924b04e90907759ba72bb20a
