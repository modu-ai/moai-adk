# Card t2028 — rule documents progress record (docs lane)

Scope: items 1, 3, 4, 5. Item 2 (sync-auditor regression table) belongs to builder-harness. Every edit is applied identically to the template source (`internal/template/templates/.claude/rules/moai/...`) and the live mirror (`.claude/rules/moai/...`).

## Item 1 — non-converging repair
- Existing-sentence check: grep over `.claude/rules/moai` for `consecutive|converg|twice in a row|repair (loop|round)` found only unrelated hits (goal-directive stagnation guard, runtime-recovery-doctrine same-rung circuit breaker at 3, agent-common-protocol-reference E1 bug-deadlock at 3). The nearest neighbour is `factory-dispatch-detail.md:168` "Ceiling. After a repair the lane re-reviews at most 2 times", which caps re-review count but does not stop a chain of repair-induced defects. No equivalent sentence: ADDED.
- Home: `workflow/factory-dispatch-detail.md`, new bullet "Non-converging repair" directly after the card-review "Ceiling" bullet (paths-scoped file, zero always-loaded cost).
- Threshold: 2 consecutive is a PROPOSED value, stated as such in the text, with rationale (one induced defect is noise, two in a row shows the approach is not converging; evidence card t2013 K6 -> F1 -> F8).
- prediction: in the next 10 lane cards that run a review/audit repair loop, a card whose progress record shows two consecutive repair-induced defects also records a revert/redesign/narrow-scope choice before any third repair commit; a third same-approach repair commit after two consecutive induced defects, with no recorded choice, falsifies this edit. verified: pending.

## Item 3 — never write secret values into logs/reports/evidence
- Existing-sentence check: `grep -i secret` over `core/moai-constitution.md` found only line 137 "Never commit secrets to version control" (commit-time, not log/report/evidence). AGENTS.md and AGENTS.md.tmpl had no `secret` hit at all. No equivalent: ADDED.
- Constitution: one bullet under "Security Boundaries" after the commit bullet.
- AGENTS.md decision: ADDED one sentence to section 7 (Tools and command output), in both `AGENTS.md` and `internal/template/templates/AGENTS.md.tmpl`. Reason: AGENTS.md is the only rule surface a Codex lane loads (the constitution is not), the cross-harness contract currently has no secret clause at all, and Codex lanes write the same verdict/evidence files. Mirroring: the template source is `AGENTS.md.tmpl`; the repo-root `AGENTS.md` is a rendered variant that already differs from the tmpl in a few paragraphs, so both were edited with the identical sentence rather than relying on a render step.
- prediction: no verdict, progress record, or log written by a lane after this lands contains a literal token/key/password value (a sweep `grep -rE` for known secret shapes over new `.moai/reports/**` finds 0 hits); one hit falsifies it. verified: pending.

## Item 4 — verify a new dependency exists before adding
- Existing-sentence check: grep for `dependenc|exist` in `development/coding-standards.md` found only the Duplicate-Prevention "exists in exactly one location" line; repo-wide grep for `slopsquat|hallucinated package|go list -m|npm view|pip index` in `.claude/rules/moai` found none. The Simplicity ladder step 5 ("already-installed dependency") covers reuse, not existence. ADDED.
- Home: new `## New Dependencies` section in `development/coding-standards.md` (before Advisory-Check Discipline), language-neutral (three registry lookups plus "or the language's equivalent"). The manager-develop agent definition is left to builder-harness.
- prediction: a dependency added to go.mod/package.json/pyproject in a card after this lands is accompanied by a cited lookup output in the progress record or commit body; a card adding a dependency whose name does not resolve in the registry reaching review falsifies it. verified: pending.

## Item 5 — shared-DB schema migration as keep-set category 3 example
- Existing-sentence check: `grep -i "migration|schema"` over `workflow/auto-semantics.md` found no hit; section 9 intro (lines 208-212) and the inventory rows name only origin push and branch disposal. ADDED.
- Edit: section 9 intro, category 3 now reads "irreversible operations touching external shared systems (for example an origin push, or a schema migration applied to a shared database)".
- prediction: a lane card that applies a schema migration to a shared database records a KEEP / operator-held disposition (not an autonomous decision record); an autonomous decision record covering a shared-DB schema migration falsifies it. verified: pending.

## Files touched (each in both trees)
- `core/moai-constitution.md` (always-loaded)
- `development/coding-standards.md` (paths-scoped via frontmatter)
- `workflow/auto-semantics.md` (paths-scoped via frontmatter)
- `workflow/factory-dispatch-detail.md` (paths-scoped via frontmatter)
- `AGENTS.md` and `internal/template/templates/AGENTS.md.tmpl` (always-loaded)

## Byte delta (wc -c, before -> after)
- moai-constitution.md 15155 -> 15296 (+141, always-loaded)
- coding-standards.md 13367 -> 13771 (+404, paths-scoped)
- auto-semantics.md 30589 -> 30670 (+81, paths-scoped)
- factory-dispatch-detail.md 29435 -> 30108 (+673, paths-scoped)
- AGENTS.md 22526 -> 22662 (+136, always-loaded; limit 40000)
- AGENTS.md.tmpl 22504 -> 22640 (+136)
- Always-loaded growth this card: +277 B across two files; no single edit exceeds the 1,000 B statement threshold of rule-authoring.md.

## Template-side follow-ups found by tests (recorded for traceability)
- The card id was removed from the template copy of the item 1 bullet: `TestTemplateNoInternalContentLeak` (class C9-card-id) and `TestReviewOwnership_CardReviewDoctrine` reject internal card tokens in template files. The evidence (card t2013, K6 -> F1 -> F8) lives only in this record.
- `internal/template/testdata/binding_ledger.json` row ALB-0115 (Security Boundaries list): `after_text` updated to include the new secret-values bullet; `before_text` is the anchor baseline and stays unchanged (changing it fails the row-coverage check). Without this, `TestBindingLedgerIntegrity` fails with "after-text absent for row ALB-0115".
- Catalog hash: `make build` ran `gen-catalog-hashes.go --all`; `catalog.yaml` came out unchanged (24603 bytes, no diff in `git status`), so no hash regeneration was needed for these rule/AGENTS edits.

## Verification (this run, tree = card branch before the commit)
- `go test ./internal/template/... -count=1 -timeout 40m`: exit 0; `ok internal/template 972.553s`, agentemit/commandemit/embedemit ok. Two earlier runs failed for the reasons above (card-id token, ledger row) plus environmental toolchain/default-10m-timeout noise.
- `make build`: exit 0.

## Agent-definition and rule edits (builder-harness pass, items a-f)
Scope: text only. Agent sources live in `internal/template/agentemit/source/`; `go run ./internal/template/agentemit/cmd/agents-emit` regenerated the template `.claude/agents`, `.codex/agents` copies and the live `super-advisor.md`. The live `.claude/agents/moai/{sync-auditor,manager-develop,plan-auditor}.md` are NOT emitter-managed (they already differ from the template for unrelated reasons), so the same text was applied to them by hand. `.moai/policies/**` is a stale pre-cutover snapshot (last touched by commit 82677fd271, previous writer left it alone): not edited.

### (a) Repair stop rule (A1 + card item 1)
- Existing-sentence check: `factory-dispatch-detail.md:169` already held the rule (previous writer); `grep -n "repair\|consecutive"` over sync-auditor.md and manager-develop.md: no hit before this edit. Operator-set audit-round budget: `git grep -i "round budget|audit rounds|budget at card"` over `.claude`, `internal/template/templates`, `.moai/policies`: no hit, so ONE sentence added, in factory-dispatch-detail only.
- Added: sync-auditor "Repair stop rule" paragraph, manager-develop "Repair stop rule" paragraph, plus a definition of "repair-induced" and the budget sentence in factory-dispatch-detail. Threshold 2 stated as a STARTING value in all three.
- prediction: in the next 10 lane cards with an audit repair loop, a card whose audits show repair-born blocking findings twice in a row records a revert/redesign/narrow/split choice before a third repair commit; a third same-approach repair commit with no recorded choice falsifies it. verified: pending.

### (b) Verdict-function baseline table (card item 2 + A2)
- Existing-sentence check: `git grep -i "allow/deny|differing row|verdict function"` over `.claude/rules` and `.claude/agents`: only an unrelated hooks-system row. ADDED.
- Added: item E9 in `manager-develop-prompt-template.md` (both trees); sync-auditor "Verdict-function cards" paragraph (auditor may add corpus rows; not for ordinary cards). No helper code.
- prediction: a later card that changes a guard/classifier/policy function carries an E9 table naming both SHAs, and an auditor-found wrong decision on such a card (the t2013 F1/F8 shape) appears as a differing row in that table before audit; a verdict-function card reaching audit without the table, or a wrong decision absent from the table, falsifies it. verified: pending.

### (c) Finding-location classification (A4)
- Existing-sentence check: sync-auditor "Finding-consumption discipline" has blocking/optional only; `git grep "in-diff|touched-file"` over rules/agents: no hit. ADDED to sync-auditor (right after that subsection) and factory-dispatch-detail (card-review list, bullet "Finding location").
- prediction: in the next 10 sync audits of lane cards, the number of rounds that fail only on `outside` findings is 0, and every High security finding outside the diff appears in the report as an operator escalation; one FAIL driven solely by an `outside` finding without reproduction falsifies it. verified: pending.

### (d) Delegation preamble (A7)
- Existing-sentence check: `grep -n "preamble|env -u|unset "` over manager-develop source and the prompt template: no hit. ADDED "Delegation Preamble (write-capable spawns)" (6 numbered lines, under 10) after E9 in the prompt template, and a one-line reference in manager-develop (both trees). The preamble names `factory-dispatch-mechanics.md` and `worktree-integration.md` as the source of truth; facts taken from mechanics § Isolation (guard refuses `-C .`, relative, computed, other-tree paths; plain git and `-C <own absolute path>` pass) and § Verification load (`unset VARS && command`).
- prediction: lane write-capable spawn prompts that include the preamble show fewer worktree-guard refusals (`tool_failure:Bash:*` rows in `.moai/lessons-inbox.jsonl` for the spawn) than spawns that omit it; equal or higher refusal counts across 10 spawns falsify it. verified: pending.

### (e) Dependency existence (card item 4)
- Existing-sentence check: coding-standards.md § New Dependencies exists (previous writer); manager-develop had no pointer (`grep -n "dependenc"`: only ANALYZE line and a security-route row). ADDED one pointer line to manager-develop, no restatement.
- prediction: a card adding a dependency cites the registry lookup output in its progress record or commit body; a dependency name that does not resolve reaching review falsifies it. verified: pending.

### (f) Catalog numbers and descriptions (A10)
- `delegation.yaml` comment "13 retained agents" replaced by "the retained agents (the files under .claude/agents/moai/ are the list; this map is the routing SSOT)" (live and template). `plan-auditor.md` line 692 "up to 3 times per SPEC (max_iterations: 3 per harness.yaml)" replaced by the tier-ceiling wording (`plan_audit_tier_ceilings` in harness.yaml: S 1, M 2, L 3; checked in `.moai/config/sections/harness.yaml:73-77`). `super-advisor` description and its "[1m]-safe Wiring" body paragraph said the model is injected at spawn and not a frontmatter pin, while the frontmatter declares `model: opus` / `effort: high`; both rewritten.
- Not touched on purpose: "13 retained agents" (12 custom + built-in Explore, enforced by `internal/harness/rosterguard`) in `agent-authoring.md`, `agents-reference.md`, `manager-design.md`, `docs-truth.md` codemap: outside the card's named list; recorded as a follow-up gap.
- Tests pinning old wording: `git grep` of the old strings in `*.go`: only rosterguard fixtures about "CLAUDE.md §4 (the 13 retained agents" (comments in internal/harness, not edited) and codex rollout testdata under `internal/cli/testdata` (frozen fixtures, not edited).
- prediction: no new file states the retained-agent count as a number outside the roster guard; a doc-vs-disk count mismatch (disk 12 files) appearing in `delegation.yaml` or a plan-auditor retry statement disagreeing with `plan_audit_tier_ceilings` falsifies it. verified: pending.

### Byte delta (wc -c, HEAD -> working tree; none of these is always-loaded)
- manager-develop-prompt-template.md 16855 -> 18222 (+1367, paths-scoped)
- factory-dispatch-detail.md 30156 -> 30905 (+749, paths-scoped)
- .claude/agents/moai/sync-auditor.md 24406 -> 25819 (+1413), manager-develop.md 21185 -> 21977 (+792), plan-auditor.md 69968 -> 69998 (+30), super-advisor.md 8733 -> 8760 (+27): agent bodies load only on spawn
- delegation.yaml 4570 -> 4626 (+56, config)
- Always-loaded growth this pass: 0 bytes (no AGENTS.md, constitution, output style, or unscoped rule touched).

### Corrections forced by the first test run (this pass)
- Run 1 (`go test ./internal/template/... ./internal/harness/rosterguard/... -count=1`) failed three ways: (1) `TestDeployedRoleRuleEntryPointMarkers`: my manager-develop repair-rule line cited `factory-dispatch-detail.md` (a role-gated rule), which requires the `moai:role-rules-required` marker; the citation was removed from all four manager-develop copies. (2) `TestManifestHashFormat` CATALOG_HASH_UNSTABLE for manager-develop, plan-auditor, sync-auditor, super-advisor: fixed by `make build` (gen-catalog-hashes --all rewrote catalog.yaml). (3) rosterguard `TestRegisteredSitesMatchTheirDeclaredAxis` / `TestNumeralBreadthSetEqualsTheDeclaredUnion` / `TestNumeralResidualArithmeticCloses`: the registry pins the `delegation.yaml` citation text "the 13 retained agents (...)" as a declared site (the 13 = 12 custom + built-in Explore). Rewording it needs a Go registry edit, which is out of scope, so the `delegation.yaml` edit (both copies) was REVERTED. A10 delegation.yaml part is NOT done; follow-up needs a card allowed to edit `internal/harness/rosterguard/registry.go`.
- binding_ledger.json: no row touched.

## Verification (this run, tree = card branch before the commit, base fe4634ff43)
- `GOCACHE=<private> go test ./internal/template/... ./internal/harness/rosterguard/... -count=1 -timeout 40m`: exit 0; ok internal/template 531.128s, agentemit 1.267s, commandemit 0.495s, embedemit 0.536s, rosterguard 47.688s.
- `make build`: exit 0 (agents-emit-check specs=12 copies=4 diff=0; catalog.yaml rewritten by gen-catalog-hashes, 24603 bytes).
- Gaps: `internal/cli` tests not run (out of scope); live `.claude/agents/moai/{manager-develop,sync-auditor,plan-auditor}.md` are hand-synced, no test pins them to the template.

## Jev step shadow record (decision ladder step 4)
- Existing-sentence check: `grep -n verbatim` over `auto-semantics.md` and `moai-lane-watchdog/SKILL.md`: only unrelated hits (SKILL.md "verbatim error text" in shell-error remedy). §12 and step 4 had no request/response record rule. ADDED. Basis: rc40 Jev analysis §C.1 stage 0 only (no stages 1/2).
- Files touched (identical text in live and template): `workflow/auto-semantics.md` (§6 row 4 cell; new "Shadow record" paragraph in §12), `skills/moai-lane-watchdog/SKILL.md` (step 4, 3 lines). `.agents/skills` has no watchdog mirror (ls: none). `internal/template/catalog.yaml` regenerated by `make build` (skill whole-tree hash). binding_ledger.json: no row touched.
- Not changed: Jev-alone decisions, operator gates, keep-set, other steps; no Go files.
- Byte delta: auto-semantics.md 30670 -> 30986 (+316, paths-scoped); SKILL.md 13964 -> 14155 (+191, loads on invocation). Always-loaded growth: 0.
- Verification: `GOCACHE=/tmp/gc2028b go test ./internal/template/... ./internal/harness/rosterguard/... -count=1 -timeout 40m` exit 0 (template 459.840s, rosterguard 30.244s). The run before `make build` failed TestManifestHashFormat and TestCatalogHashCoversSkillSubfiles (stale watchdog hash); fixed by make build. `make build` exit 0.
- prediction: in the next 10 lane cards that call `jev_ask`, each decision record names a file or section holding the verbatim request and response with `model`; a Jev-consulted lane decision with no such record falsifies it. verified: pending.

## Lane verification and close-out (lane-4, own runs)

- Commits on the card branch: merge 1650a20b57, rules fe4634ff43, agents 1804e4fa29, Jev shadow d7c11ab7d3.
- Own checks at d7c11ab7d3: `git status --short` empty; `git diff --stat fe4634ff43 -- '*.go'` empty; added lines with `cc -a|glm -a|[-f | -a]|--agent`: 0; `git diff --name-only release/main-batch HEAD | grep -c -- '-core\.md'`: 0.
- Core stub gate (leader R-note, budget 10,000 UTF-16): command `go test ./internal/template/ -run 'TestRoleCoreStubGate$' -count=1 -v` -> PASS (ok 0.485s); factory-dispatch-core.md 9994 and cross-session-messaging-core.md 8072, identical in the deployed and template trees. No *-core.md file is touched by this card (net 0), so nothing needed moving to a detail file.
- Template/rosterguard test runs and `make build` are the writers' reported results (not re-run by the lane): rules writer `ok internal/template 972.553s`; agents writer `ok internal/template 531.128s` + rosterguard `ok 47.688s`; Jev writer `ok internal/template 459.840s` + rosterguard `ok 30.244s`.
- binding_ledger.json: only row ALB-0115 `after_text` touched (required by TestBindingLedgerIntegrity after the constitution bullet).

## Deferrals / Gaps
- delegation.yaml "13 retained agents" NOT reworded: rosterguard pins the citation text as a registered site, so a reword needs a Go registry edit (out of scope). The 13 = 12 custom agents + built-in Explore. Same phrase left in agent-authoring.md, agents-reference.md, manager-design.md and the docs-truth codemap (not on the card's list). plan-auditor.md:692 and the super-advisor description were fixed.
- `.moai/policies/**` copies untouched (stale pre-cutover snapshot).
- Template copy of the watchdog SKILL.md already differed from the live copy by one card-id phrase (line 202) before this card; left alone.
- All `prediction:` entries are `verified: pending`; the 2-consecutive threshold is a PROPOSED starting value.
- card-review: verdict fail, two outside-the-diff P1 findings on Go batch code (see card-review.md); not re-run.

## AGENTS.md headroom repair (leader instruction, candidate 4855b23ee8 failure)

`TestContractDocumentsKeepHeadroom`: limit 22528 B; before AGENTS.md 22662 / AGENTS.md.tmpl 22640; after 22505 / 22483. Wording-only; the secret-value duty (fe4634ff43) is kept.

| Location | Before | After |
|---|---|---|
| §7 secret sentence | "...into logs, reports, or evidence files; record the variable name or a redaction." | "...into logs or reports; record the name or a redaction." |
| §7 session length | "Prefer one warm session for the same work; a new or cold session re-pays..." | "Prefer one warm session; a new or cold one re-pays..." |
| §7 MCP | "...for library documentation and established patterns, then continue — analysis quality must not depend on MCP availability." | "...for library documentation, then continue; quality must not depend on MCP." |
| §7 tool usage | "Use absolute paths and verify a path exists rather than constructing it from an assumption." | "Use absolute paths; verify a path exists rather than assuming it." |
| §7 output | "A runtime output limit is a backstop, not the target." / "— not forms emitting spinners, ..." | "A runtime limit is only a backstop." / "— not spinners, ..." |

Observed (this run, this tree): `go test -count=1 -run TestContractDocumentsKeepHeadroom ./internal/config/` ok; `./internal/config/` ok 32.673s; `internal/template` `TestRoleCoreStubGate$|TestBindingLedgerIntegrity` ok; Catalog|Manifest|AgentsMd|AGENTS subset ok. Gap: full `internal/template` not re-run; tmpl/live byte parity checked only by identical substitutions.
