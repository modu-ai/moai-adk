# SPEC-AUDIT-MODEL-CONVERGE-001 — research

> Tier L research artifact (0.1.2). Rows R-1..R-24 were measured in the
> plan-phase run on this tree (`.moai/worktrees/t1423`, branch
> `WT-audit-model-convergence`) at HEAD `c50da9c2f`; rows R-25..R-30 and §R.5
> were re-measured for plan-audit iteration 1 at HEAD `53a42f013` (the same code
> plus the SPEC directory, R-28), unless the row says otherwise. Commands run from the tree root.
> A row quoting "exit" read it from a `sh -c '…; echo exit=$?'` wrapper; the
> command column itself is the plain command.

## §R.1 Measured facts

| # | Claim | Command | Observed (verbatim or trimmed) |
|---|---|---|---|
| R-1 | No non-test code reads `Audit.Model` | `grep -rn --exclude="*_test.go" "Audit\.Model" internal cmd` | (no output), exit 1 |
| R-2 | `activeAuditBackend` has a definition and no caller | `grep -rn --exclude="*_test.go" "activeAuditBackend" internal cmd` | `internal/settings/schema_sections.go:367` (comment) · `internal/config/closed_sets.go:87` (comment) · `internal/cli/mcp_audit.go:6` (comment) · `:46` (comment) · `:52` `func activeAuditBackend(model string) (string, error) {`, exit 0 — five hits, one definition, zero calls |
| R-3 | The stale sentinel and deferral text | `grep -rn "multiConvergenceImplemented" --include="*.go" internal` | `internal/cli/mcp_audit_test.go:40,41` · `internal/cli/mcp_audit.go:27,31` (`const multiConvergenceImplemented = false`),`:49`, exit 0. And `grep -rn "deferred to SPEC-AUDIT-MULTI-MODEL" --include="*.go" internal` → `internal/cli/mcp_audit.go:51`, exit 0. `internal/config/audit_models.go` carries the same deferral in prose at its header and on `AuditModelMulti` (read, not grepped by that exact string). |
| R-4 | `audit_multi` takes gates from the call only; `runMultiAudit` has one caller | read `internal/cli/mcp_audit_multi.go:119-134`; `grep -rn "runMultiAudit\|MultiAuditConfig{" --include="*.go" internal` minus tests | `readGatesArgument` fills claude=required, codex=required, glm=advisory from absent keys; the only production caller of `runMultiAudit` is `handleAuditMulti` (`mcp_audit_multi.go:92`) |
| R-5 | The RAW `audit.gates` block already has three readers (the intake counted one) | read `internal/cli/mcp_convergence.go:783,888-912,945-954`; `internal/cli/mcp_worktree_root.go:101-113`; `internal/cli/mcp_codex.go:1974-1992`; `internal/auditreceipt/store.go:204-257`; `internal/hook/audit_receipt_guard.go:84` | `enforceRequiredGateUnmet` (via `workflowAuditGates` -> `resolveAuditGates`) fails the overall verdict for an explicit `required` backend that returned `inconclusive`; `applyGateUnmet` does the same on `codex_audit`; `rawCodexGate` decides receipts and the SubagentStop guard. All read the raw block and ignore the engine default on purpose. |
| R-6 | Leg budgets; where 900 s lives | `grep -n "claudeAuditTimeout\|glmAuditHTTPTimeout" internal/cli/mcp_claude.go internal/cli/mcp_glm.go`; read `internal/cli/mcp_codex.go:430-436`; `grep -n "DefaultCodexReviewGateTimeout\|DefaultMultiReviewGateTimeout" internal/config/defaults.go`; read `.claude/settings.json:186-193` | Claude leg `5 * time.Minute` (`mcp_claude.go:27`); GLM leg `120 * time.Second` (`mcp_glm.go:80`); codex leg: "a bounded context deadline is the caller's responsibility" (no deadline in the code). 900 s = `DefaultCodexReviewGateTimeout` / `DefaultMultiReviewGateTimeout` (`defaults.go:557,570`) and the `"timeout": 900` of the `handle-codex-review-gate.sh` Stop hook — not an `audit_multi` bound. No `MCP_TOOL_TIMEOUT` is set in the repository's configs. |
| R-7 | Committed yaml: claude pin at `medium`, header comment says medium, no `model` key | `grep -n -A1 "model: claude-opus-5-5" .moai/config/sections/workflow.yaml`; `grep -n "claude-opus-5-5/medium" …`; `grep -nE "^        model: multi$" …` | `24:            model: claude-opus-5-5` / `25-            effort: medium`, exit 0; `20:    # z.ai states low|high|max. Claude defaults to claude-opus-5-5/medium and`, exit 0; (no output), exit 1. `git ls-files` shows the file is tracked. |
| R-8 | The distributed template needs no pin change and has no `model`/`gates` keys | `grep -n "effort: high" internal/template/templates/.moai/config/sections/workflow.yaml`; `grep -c "model: multi" …` | `136:            effort: high` · `139:            effort: high`, exit 0 (the claude pin at 136 is already `high`); `0`, exit 1 |
| R-9 | The auditors are told in prose to branch on `audit_model` | read `.claude/agents/moai/plan-auditor.md:207-232,745`; `.claude/agents/moai/sync-auditor.md:170-200,230`; `.claude/skills/moai-ref-cross-model-audit/SKILL.md:11-37`; `.claude/skills/moai/workflows/review.md:211,213,263` | "Single-backend audit mode (per the project's `audit_model`):" (one hit in each agent, local and template, `grep -cF`); the skill's "When to use" table keyed on `audit_model: …`. Both auditors carry `Bash` and the four audit MCP tools in `tools:`. |
| R-10 | The sync happy path suppresses the cold auditor; the machine predicate has no production caller | read `.claude/skills/moai/workflows/sync.md:77-81`; `grep -rn "IsBinding\|FourDimVerdict" --include="*.go" internal cmd` minus its own file | `Binding promotion … SHALL NOT spawn the cold sync-auditor subagent`; `IsBinding` is referenced only by `internal/runtime/sync_4dim_binding_test.go` (lines 39, 55, 76) — the production consumer is orchestrator prose. |
| R-11 | Existing guards that bind the new code | read `internal/cli/mcp_convergence_test.go:532-547`; `internal/web/mcp_audit_surface_test.go:45-71`; `internal/cli/mcp_audit_multi_test.go:352-373`; `internal/template/claude_audit_surface_test.go:11-68` | `TestConvergence_NoDirectFrontmatterRead` (no `frontmatter`/`agent_overrides` in `mcp_convergence.go`); `TestConvergence_NoNewAuditModelEnum` (no `AuditModel* =` there); `TestWebConsole_AuditNoForkedInterpreter` (sentinel string `"activeAuditBackend"` at line 58, scanned in non-test files of `internal/web`); `TestAuditMulti_NoHardErrorPath_AC_AMM_024`; `TestClaudeAuditTemplateSurfacesAndCatalogHash` (the `## MCP Audit Tools (cross-model second opinion)` section must be identical in the local and template auditors; the skill's catalogue hash must match). |
| R-12 | Tests resolve the project root from the environment, which lane sessions set | read `internal/cli/session.go:262-280` | `resolveProjectDirWithSource`: `CLAUDE_PROJECT_DIR` first, then the process cwd. `go test` in `internal/cli` has no yaml in its cwd; a session that exports `CLAUDE_PROJECT_DIR` points the same tests at the real repository root — where the committed yaml will say `model: multi`. |
| R-13 | Tool provenance | `moai version`; the session's MCP server banner | installed binary `v3.2.0-rc.24 … gc50da9c2f built 2026-10-02T05:34:44Z` (= this tree's HEAD, so `moai spec lint` below was judged by a build from this tree); the MCP server of this session reports `v3.2.0-rc.23 (commit: d194083fb)` and states it keeps the build it started with |
| R-14 | No import cycle for `internal/auditreceipt -> internal/config` | `go list -deps ./internal/config` filtered for `auditreceipt`; `go list -deps ./internal/auditreceipt` filtered for `moai-adk/internal` | `0` matches; only `internal/auditreceipt` itself (both with a pipe — informational, not a ledger row) |
| R-15 | Local and template copies are not byte-identical today | `diff -q` per pair | DIFFER: `.claude/agents/moai/plan-auditor.md`, `.claude/agents/moai/sync-auditor.md`, `.claude/skills/moai/workflows/sync.md`, `.claude/workflows/sync-audit-4dim.js`. IDENTICAL: `.claude/skills/moai-ref-cross-model-audit/SKILL.md`, `.claude/skills/moai/workflows/review.md`, `.claude/rules/moai/core/moai-mcp-tools.md`. The pre-existing differences are unrelated to this SPEC and must not be "fixed" by it. |
| R-16 | `.codex` role TOMLs are emitted from the agent `.md` | `ls .codex/agents/moai/` (root); read `Makefile` targets `agents-emit`, `agents-emit-check`, `build` | the project root has no `.codex/agents/moai/`; `internal/template/templates/.codex/agents/` exists; `make build` runs `agents-emit-check` first (read-only; regeneration is the explicit `make agents-emit`) |
| R-17 | A new MCP tool fans out widely | read `internal/mcp/catalog.go:1-60` and the tool-count prose | 45 catalogue entries; `moai-mcp-tools.md` and its catalogue companion both say "45 tools"; `Makefile` `tool-policy-drift-check` guards the permission sets (see design.md §D.4) |
| R-18 | Methodology | `grep -n development_mode .moai/config/sections/quality.yaml` | `2:    development_mode: tdd` |
| R-19 | `CLAUDE.local.md` | `ls CLAUDE.local.md` in the worktree and in the primary checkout | `No such file or directory` in both; the template-neutrality principle was read from the comment in `internal/config/defaults.go` instead |
| R-20 | The Go default `Audit.Model` is pinned | read `internal/config/mcp_audit_config_test.go:42-48` | `TestAuditConfig_DefaultProfile` asserts `a.Model == AuditModelClaude` |
| R-21 | A deadline on the codex leg stops the process, and the session returns partial text on a context end | read `internal/cli/mcp_codex.go:518-540` (`realCodexSessionRunner.start`: `exec.CommandContext(ctx, binaryPath, args...)`), `:1255-1275` (`awaitCodexTurnReview`: on `ctx.Err()` it returns `bestCodexReviewText(reviewText, agentText), nil`), `:1020-1031` | context expiry kills the codex subprocess and closes its stdout, which ends the blocked `recv()`; but the turn reader returns whatever review text it had with a nil error, so a leg whose context ended — by its own deadline or by the caller's cancellation, the code returns the text either way — must be forced to `inconclusive` and its text discarded (design.md §D.10 item 6) |
| R-22 | The template leak test and the script's allowlist | read `internal/template/internal_content_leak_test.go:870-905`; `TestTemplateNoInternalContentLeak` at line 1541 | template files may not carry internal SPEC identifiers; the only allowlisted identifier in `sync-audit-4dim.js` is the illustrative `SPEC-FOO-001` launch example. Text added to the template copies of `sync.md` and the script must be identifier-free. |
| R-23 | The sync-audit script's agents are read-only and its verdict is pure JS; `audit_multi` is write-capable | read `.claude/workflows/sync-audit-4dim.js` header (`Read-only: every agent … agentType 'Explore'`; `No meta-judge agent`; `No LLM arithmetic`) and the `Verdict` phase; `internal/mcp/catalog.go:36-40`; `grep -c "mcp__moai__"` on both script copies | `0` MCP-tool references in either copy; the script has no write capability and no call between judge collection and the returned verdict; `audit_multi` files receipts and state. Hence the added call sits with the orchestrator in `sync.md` (design.md §D.8). Whether an `Explore` workflow agent could carry MCP tools was not observed. |
| R-24 | No recorded location for the orchestrator's binding statement | `grep -n "verdict\|§E.4\|audit-ready" .claude/skills/moai/workflows/sync.md`; `grep -n "sync-audit-4dim\|FO-SYNC-1\|BINDING\|binding" .claude/skills/moai/workflows/sync/*.md` | `sync.md` names no file that stores the statement of which verdict is binding; `sync/*.md` has no hit for the binding rule. Left as plan.md §B OQ-9. |
| R-25 | An unknown verb under `moai verify` prints the group help and exits 0; the group's root flag and registration pattern | `moai verify audit-plan --help` (stdout beginning ` Shared diagnostic snapshot contract.`, exit 0); `moai verify check --help`; read `internal/cli/verify.go:34-91`, `codex_review_receipt.go:177`, `verify_receipts.go:16` | the group has no argument validation, so the absence of a verb cannot be detected by exit code; `--project-root` is a PERSISTENT flag on the group defaulting to `$CLAUDE_PROJECT_DIR`, then the working directory (`verifyResolveRoot`) — in a worktree-isolated session that environment variable names the primary checkout; verbs register through `verifyExtraCommands` (`sync-gate`, `codex-review`) beside `newVerifyRecordCmd` / `newVerifyCheckCmd`. `moai verify --help` lists record, check, sync-gate, codex-review. |
| R-26 | Import weight of `internal/auditreceipt -> internal/config` | `go list -deps ./internal/hook` filtered for `moai-adk/internal/config$` (count); `go list -deps ./internal/config` filtered for a `.` in the path (count); `go list -deps ./internal/auditreceipt` filtered for `moai-adk` (count) | `1` (the hook package already imports `internal/config`); `29` module-path dependencies of `internal/config`; `1` (`internal/auditreceipt` imports no other moai package today). All three use a pipe — informational, not ledger rows. |
| R-27 | What bounds the `audit_multi` call, and the codex siblings | read `internal/config/defaults.go:572-593` (`DefaultCodexTaskTimeout = 600 * time.Second` at 587; `DefaultCodexAuditTimeout = 20 * time.Minute` at 593, "An audit reads a SPEC and its tree and can take minutes"); `internal/cli/codex_audit_launch.go:640-655` (`codexAuditExec` bounds one audit process with it); `internal/cli/codex_task.go:172`; `grep -n "runMultiAudit\|performCodexAudit" internal/cli/codex_review_gate.go internal/cli/multi_review_gate.go`; read `internal/cli/multi_review_gate.go:13,79,120` | both codex bounds are `var`s ("Not a compile-time const … so a test can shorten it"); the two review-gate hook files reference neither `runMultiAudit` nor `performCodexAudit` (no output), and the multi-review gate only loads a result already persisted by `persistConvergenceResult` — so neither the 900 s Stop-hook budget nor any hook bounds an `audit_multi` MCP call; no host tool timeout is configured in the repository (R-6) |
| R-28 | The code is unchanged between the authoring base and the audited HEAD | `git diff --stat c50da9c2f HEAD -- internal .claude .moai/config` | (no output) — `53a42f013` adds only the SPEC directory |
| R-29 | Which tests read `workflow.yaml` or resolve the root from `CLAUDE_PROJECT_DIR` | `grep -rln 'CLAUDE_PROJECT_DIR\|EnvClaudeProjectDir' --include='*_test.go' internal` by package; `grep -rln 'workflow.yaml' --include='*_test.go' internal` by package; `grep -rln 'CodexGateRequired\|workflowAuditPins\|resolveAuditGates\|handleAuditMulti\|runMultiAudit' --include='*_test.go' internal`; the top-level test-name prefixes of the audit-reading files | files naming `CLAUDE_PROJECT_DIR`: `internal/cli` 128, `internal/hook` 43, `internal/template` 11, `internal/codexwiring` 3, `internal/kanban` 2, `internal/session` 2, `internal/constitution` 1, `internal/navigator` 1. Files naming `workflow.yaml`: `internal/cli` 44, `internal/config` 22, `internal/hook` 12, `internal/web` 10, `internal/template` 8, `internal/settings` 6, `internal/auditreceipt` 3, `internal/harness` 3, others 1-2. Test files that reach the gate readers: 15 in `internal/cli` (among them `mcp_project_root_codex_test.go`, `required_gate_block_test.go`, `codex_verdict_divergence_test.go`, `codex_audit_required_block_test.go`, `mcp_build_identity_test.go`, `wsr_state_root_test.go`) plus `internal/auditreceipt/store_test.go`. `internal/cli` audit-reading name families (by count): `TestCodexAudit`, `TestConverge`, `TestAuditMulti`, `TestReviewGate`, `TestMultiReviewGate`, `TestGLMAudit`, `TestClaudeAudit`, `TestAC`, `TestRunMultiAudit`, `TestRunCodexReviewGate`, `TestRunMultiReviewGate`, `TestPersistConvergenceResult`, `TestConfigOrphanedWorktree`, `TestWSR0…`, `TestPerformCodexAudit`, `TestPerformGLMAudit`, `TestHandleCodexReviewGate`, … The review-gate tests live in `internal/cli` (the gates are `internal/cli/*_review_gate.go`), not in `internal/hook`; the hook-side receipt-guard tests are `TestAuditReceiptGuard…`, `TestWSR007/008/009`, `TestSubagentStop…` (`internal/hook/audit_receipt_guard_test.go`, `wsr_audit_receipt_tree_test.go`). `internal/hook`'s full suite is a heavy run and is not used (AC-ACV-018 names the families). |
| R-30 | Facts the plan-audit cited, re-measured | read `internal/cli/audit_pin.go:34-64` (`loadWorkflowAuditSection` returns `config.AuditConfig{}, nil` for an absent file and an error for a read or parse failure; `workflowAuditPins` turns that error into a zero value); `internal/auditreceipt/store.go:236-257` (`rawCodexGate`: a private struct, `return ""` on a read or parse error); `internal/cli/mcp_worktree_root.go:99-113` (`resolveAuditGates`); `internal/cli/mcp_codex.go:1263-1268` (`awaitCodexTurnReview` returns `bestCodexReviewText(reviewText, agentText), nil` on `ctx.Err()` and on a closed channel); `grep -n "moai:closure-second-review" …sync-auditor.md`; `grep -n "audit_model" .claude/skills/moai/workflows/review.md` | all three raw readers confirmed; the partial-text return confirmed on any context end; local `sync-auditor.md` carries the closure markers at lines 178 and 194 (template 162 and 178); `review.md` interprets `audit_model` at lines 211, 213, 263, 531, 542. |

## §R.2 Corrections to the card intake

1. **"The only config-gates consumer is `auditreceipt/store.go:256`."** Incomplete.
   That is the only reader that names `Audit.Gates` directly; two more read the
   same raw block through `workflowAuditPins(root).Gates` — `resolveAuditGates`
   (`mcp_worktree_root.go:101`), called by `enforceRequiredGateUnmet`
   (`mcp_convergence.go:783`) and `applyGateUnmet` (`mcp_codex.go:1974`). So the
   fail-closed rule for an explicitly configured `required` already holds on both
   audit surfaces; what is missing is a way for a *token* to set it (R-5).
2. **"Verify whether the template yaml needs the effort alignment."** It does
   not: the template claude pin is already `high` (R-8). Only the committed
   dev-repo yaml differs (R-7).
3. **"900 s hook call budget."** The figure is the Stop-hook wrapper budget, not
   a bound on `audit_multi` (R-6). The legs' own bounds are 5 min / 120 s /
   request context.
4. **Fact 2 stands but is narrower than it reads.** `audit_multi` ignores config
   for *which backends run*; it does not ignore config for *whether an unmet gate
   fails* (R-5).

## §R.3 Hazards the measurements expose

- **Sync-audit never reaches `audit_multi` on the clean path** (R-10). Wiring
  the auditors alone would leave the sync half of the card a no-op whenever the
  4-dimension verdict is clean.
- **A token-chosen `required` is fail-open unless gate readers see it** (R-5).
  `enforceRequiredGateUnmet` keys on the raw `audit.gates` block, so
  `model: multi` alone would change which backends run and nothing about what
  happens when codex does not answer.
- **Env leak into existing tests after the yaml change** (R-12). The existing
  audit tests that do not pass `project_root` resolve their root from
  `CLAUDE_PROJECT_DIR`; in a lane session that is the real repository.
- **The web guard goes vacuous when `activeAuditBackend` is deleted** (R-11).
- **The installed MCP server cannot show a new tool** (R-13), which is the
  measured reason the read-only surface is a CLI verb.
- **A deadline alone is not enough** (R-21, R-30): the codex turn reader returns
  partial text on ANY context end — deadline or caller cancellation — so the leg
  must discard it in both arms.
- **An older server looks like a pass** (R-5, E22): a `model: multi` yaml with
  codex `inconclusive` returns `overall_verdict: pass` with an empty `gate_unmet`
  from a server that ignores the token; keying fail-closed behaviour on
  `gate_unmet` alone misses it, so the result is checked against the plan.
- **An unknown verb is not an error under `moai verify`** (R-25): an older
  binary is detected by the verb's output contract, not its exit code.
- **The sync call cannot live in the script** (R-23): read-only agents and a
  pure-JS verdict; it lives with the orchestrator, and the template copies must
  stay identifier-free (R-22).

## §R.4 Not observed in this run

- Live behaviour of `audit_multi` against real codex and GLM backends; no Go
  test was executed in the plan phase (this phase edits no Go).
- Whether the settings wizard writes the default `claude` token to YAML.
- Whether a Codex-hosted read-only auditor role can execute a shell verb.
- `moai update` preservation of a user-set `audit.model` (a code comment only).
- The `internal/cli` import graph for the new verb beyond its registration
  through `verifyExtraCommands` (R-25).
- Real codex adversarial-review durations (the deadline value is derived from
  `DefaultCodexAuditTimeout`, R-27, not measured).
- Whether the settings wizard persists an untouched `claude` radio (decision D7' makes either answer harmless).
- Where the sync binding statement is persisted (R-24), and whether a workflow
  `Explore` agent could carry MCP tools (R-23).

## §R.5 Plan-audit iteration 1 — which cited claims held on this tree

Each file:line the audit cited was re-read or re-run at HEAD `53a42f013`; none was
carried over. Result: **every cited claim held.**

| Audit item | Cited claim | Re-measured here | Held |
|---|---|---|---|
| PA1-D1 | an older MCP server turns a required codex that never answers into a pass with empty `gate_unmet` | reproduced with the throwaway `go test -overlay` test of acceptance.md E22: `overall_verdict=pass gate_unmet="" fail_open=[codex]` | yes |
| PA1-D2 | `workflowAuditPins` swallows the parse error (`audit_pin.go:58-64`); the verb must call `loadWorkflowAuditSection` | read, R-30 | yes |
| PA1-D3 | `awaitCodexTurnReview` returns partial text with a nil error on any context end (`mcp_codex.go:1264-1266`) | read, R-30 (lines 1263-1268) | yes |
| PA1-D4 | `DefaultCodexTaskTimeout` 600 s and `DefaultCodexAuditTimeout` 20 min exist in `defaults.go:587-593`, as `var`s | read, R-27 (589-593 for the audit bound) | yes |
| PA1-D4 (added by the brief) | the 900 s Stop-hook wrapper does not bind the `audit_multi` call because the hook only reads a persisted result | read `multi_review_gate.go:13,79,120`; the two review-gate files reference neither `runMultiAudit` nor `performCodexAudit`, R-27 | yes |
| PA1-D5 | the raw loaders keep an absent `audit.model` empty, while the Go default carries `claude` | read `audit_pin.go:34-53`, `store.go:236-257`, `mcp_worktree_root.go:101-110`, `defaults.go:1272-1304`; E24 (no raw reader constructs the default config) | yes |
| PA1-D6 | `claude` is paired with the default gates in code | `defaults.go:1272-1304` | yes (superseded by decision D7') |
| PA1-D7 | the installed binary has no such verb | E3: the group help, exit 0 (the audit read exit 1 for the old spelling `moai audit plan`; under the new spelling the exit is 0 — R-25) | yes (spelling changed) |
| PA1-D8 | the literal-presence ACs are weak; the leak sweep is a name regex over `internal/cli` | read the 0.1.1 ACs; R-29 gives the wider set | yes |
| PA1-D9 | `IsBinding` has no production caller | `grep -rn "IsBinding" --include="*.go" internal cmd` minus its own file and test finds no production caller (measured at 0.1.0, R-10) | yes |
| PA1-D10 | `review.md` interprets `audit_model`; the local `sync-auditor.md` carries closure markers | R-30 | yes |
| PA1-D11 | importing `internal/config` into `internal/auditreceipt` adds config's dependency set; no cycle | R-26 | yes (and `internal/hook` already imports it) |

Not re-run: the audit's own overlay probes (`TestPlanReviewFailOpenProbes`,
`TestPlanReviewCallerCancellation`); the conclusions were re-derived from the
cited source lines and, for PA1-D1, from E22.
