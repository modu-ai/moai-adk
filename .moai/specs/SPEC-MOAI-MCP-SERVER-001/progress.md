# Progress — SPEC-MOAI-MCP-SERVER-001

> Tier L · plan-phase skeleton. `§E.2`-`§E.4` are placeholder headings only at plan-phase (the literal `§E.N` tokens are parser-load-bearing for `internal/spec/era.go` `ClassifyEra` — do NOT rename or omit). Populated by manager-develop (§E.2/§E.3) and manager-docs (§E.4) during run/sync.

## §E.1 Plan-phase Audit-Ready Signal

_Plan-phase audit complete: PASS, 0.89 (Tier L threshold 0.85, skip-eligible), 2026-08-05._

- plan_status: audit-ready
- plan_complete_at: 2026-08-05T10:27:04Z

## §E.2 Run-phase Evidence

### M1 — Server scaffold + core read/status tools + .mcp.json provisioning

**Deliverables:** `internal/cli/mcp_server.go` (NEW — `moai mcp-server` subcommand + 9 thin-wrapper tool handlers + neutral `.mcp.json` entry seam), `internal/cli/mcp_server_test.go` (NEW — TDD coverage incl. in-process + real-stdio subprocess round-trip), `internal/cli/root.go` (AddCommand), `go.mod`/`go.sum` (mark3labs/mcp-go v0.57.0 promoted to direct dep).

**Same-core-two-surfaces (C1/AP-1) — each handler calls the SAME internal/ fn the CLI uses:**

| Tool | Wraps (verified) |
|---|---|
| `session_list` | `session.QueryActiveWork` (registry.go:254) |
| `goal_status` | `goal.LoadGoal` (state.go:57) |
| `goal_arm` | `goal.NewGoal` (schema.go:115) + `goal.SaveGoal` (state.go:76) — the cli/goal.go arm composition; inherits parseCondition + the infinite-goal fail-closed |
| `spec_progress` | `spec.ListDocs` (listdocs.go:36) — the SPEC scanner `internal/web` board buildBoardView also calls |
| `verify_snapshot` | `verify.Load` (store.go:38) + `verify.RecordCheck` (store.go:107) — first CLI/MCP surface for verify |
| `verify_trend` | `verify.Load` (store.go:38) |
| `spec_audit` | `spec.Audit` (audit.go:156) |
| `spec_drift` | `spec.Audit` (audit.go:156), drift subset |
| `audit_cache` | `runtime.AuditCache` ComputeHash/Lookup (audit_cache.go:50/InMemoryCache) |

**AC matrix (M1):**

| AC | Status | Evidence |
|---|---|---|
| AC-MCP-001 (subcommand + stdio initialize→tools/list→tools/call) | PASS | `TestNewMCPServerCmd_Registered` + `TestMoaiMCPServer_ToolsListDeclaresSchema` (in-process) + `TestMCPServer_StdioRoundTripSubprocess` (real stdio over pipes); server name = `moai` |
| AC-MCP-002 (opt-in default-off) | PASS | `TestProvisionMoaiMCPServerEntry_OptInDefaultOff` — fresh project untouched until provision runs; no template provisioning in M1 |
| AC-MCP-003 (thin-wrapper parity — same internal/ fn as CLI) | PASS | code inspection: every handler cites its verified file:line (table above); `TestMoaiMCPServer_SessionListRoundTrip` proves session_list reaches `session.QueryActiveWork` |
| AC-MCP-004 (tools/list JSON Schema every core tool) | PASS | `TestMoaiMCPServer_ToolsListDeclaresSchema` asserts every tool carries a non-empty `inputSchema.Type` |
| AC-MCP-005 (core tools wrap verified integration points) | PASS | `TestMoaiMCPServer_CoreHandlersRoundTrip` exercises all 9 tools via tools/call against representative state |
| AC-MCP-006 (.mcp.json single neutral entry via atomic-config helpers) | PASS | `TestBuildMoaiMCPServerEntry_Neutral` ({command:"moai",args:["mcp-server"]}, no env/secrets/SHA/date) + `TestProvisionMoaiMCPServerEntryAt_Idempotent` (reuses `mutateClaudeJSONAtomic` + the existing `mcpEntryEqual`) |

**Verification commands + results:**

- `go build ./...` → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- `go vet ./internal/cli/` → exit 0
- `golangci-lint run --timeout=3m ./internal/cli/` → 0 issues
- `go test -count=1 -run '<M1 suite>' ./internal/cli/` → `ok` (all M1 tests PASS)
- `go test -count=1 ./internal/cli/...` → all subpackages `ok` (cross-cutting gate, trust-but-verify)
- `go test -cover ./internal/cli/` → `mcp_server.go` 86.2% (100/116 stmts); sole 0% fn = `runMCPServer` (blocking ServeStdio entry — behaviorally covered by `TestMCPServer_StdioRoundTripSubprocess`, not attributable to the parent coverprofile since it runs in a child process)
- E4 subagent boundary: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/mcp_server.go` (excl. tests/comments) → 0 (REQ-MCP-014)
- E8 RED (verbatim, pre-GREEN): `internal/cli/mcp_server_test.go:34:9: undefined: newMCPServerCmd` (+ `newMoaiMCPServer`, `moaiMCPServerName`, `buildMoaiMCPServerEntry`, `provisionMoaiMCPServerEntryAt`, `moaiMCPServerKey`) → `FAIL [build failed]`

**M1-scope decision (deferred to M4):** the scope-resolving `provisionMoaiMCPServerEntry(scope)` wrapper is intentionally absent from M1 — it belongs with the `moai init` / `moai web` call sites + Template-First reversal (plan.md §F M4, design.md §6.3). M1 ships ONLY the path-explicit seam `provisionMoaiMCPServerEntryAt(configPath)` + the neutral `buildMoaiMCPServerEntry()`, so no dead/uncalled code is carried (scope discipline).

**Multi-session race note:** during M1 a concurrent `manager-develop-mcp-m0` agent was live-editing the same files; ownership was consolidated to `manager-develop-mcp-m1` by the team-lead before the final coherence pass + commit. No double-commit occurred (0 commits since `eb1dd5c9c` prior to this M1 commit).

### M2 — Codex audit backend (Phase 1 tools) + Stop-hook gate

**Deliverables:** `internal/cli/mcp_codex.go` (NEW — codex_audit + codex_setup handlers + codex JSON-RPC client + review-output.schema.json types + fail-open + readCodexReviewGateEnabled config reader), `internal/cli/codex_review_gate.go` (NEW — Stop-hook gate pure logic: self-gate + ALLOW/BLOCK + working-tree change detector), `internal/cli/mcp_codex_test.go` + `internal/cli/codex_review_gate_test.go` (NEW — TDD coverage, injectable seams, real temp-git-repo + /bin/cat integration paths), `internal/cli/mcp_server.go` (register the 2 new tools in `registerMoaiMCPTools`), `internal/cli/hook.go` (register `codex-review-gate` subcommand + `runCodexReviewGate` RunE), `internal/config/types.go` + `internal/config/defaults.go` (NEW `Codex`/`CodexReviewGateConfig` opt-in gate, default OFF; `DefaultCodexReviewGateTimeout = 900s`), `.claude/hooks/moai/handle-codex-review-gate.sh` (NEW shell wrapper; settings.json TEMPLATE registration deferred to M4), `hook_test.go`/`hook_pre_push_test.go`/`hook_e2e_test.go` (subcommand-count + utility-subcmd ledger updates for the new hook).

**Codex backend (design.md §3 M2 / §G.4):** `codex_audit` shells out to the `codex app-server` JSON-RPC mode — `mode=native` → `review/start`, `mode=adversarial` → `turn/start` + an adversarial-review prompt. Output adopts the locked `review-output.schema.json` (`verdict`/`summary`/`findings[severity,title,body,file,line,confidence,recommendation]`/`next_steps`) via `mcp.NewToolResultJSON[ReviewOutput]` + `mcp.WithOutputSchema[ReviewOutput]()`. The codex binary is OPTIONAL + experimental (R1); every path fails open — missing / erroring / malformed codex ⇒ structured `VerdictInconclusive` (REQ-MCP-012 preview; the full 3-way claude-fallback plumbing is M3, but M2 guarantees no hard crash on missing codex). NO Node bridge of any kind (REQ-MCP-007) — enforced by `TestMCP_Codex_NoNodeBridge`.

**AC matrix (M2):**

| AC | Status | Evidence |
|---|---|---|
| AC-MCP-007 (codex_audit unified modes + schema output + fail-open on missing codex) | PASS | `TestCodexAudit_NativeDispatchesReviewStart` (review/start), `TestCodexAudit_AdversarialDispatchesTurnStart` (turn/start), `TestReviewOutputSchemaShape` (§G.4), `TestCodexAudit_FailOpenOnMissingCodex` / `_OnCodexError` / `_OnMalformedResponse` (VerdictInconclusive, no panic) |
| AC-MCP-008 (codex_setup Go probe, no Node bridge) | PASS | `TestCodexSetup_GoProbeNoNodeBridge` (LookPath + `codex --version` + ChatGPT auth classification + `enable_review_gate` toggle), `TestCodexSetup_NotInstalledReportsUnknown`, `TestMCP_Codex_NoNodeBridge`, `TestClassifyCodexAuth_Branches` |
| AC-MCP-009 (review-gate self-gate prevents false blocks) | PASS | `TestReviewGate_NoEditTurnAllows` (clean tree ⇒ ALLOW, no codex call), `TestReviewGate_LoopPreventionAllows` (stop_hook_active ⇒ ALLOW), `TestReviewGate_DisabledAllows` (opt-in default-off) |
| AC-MCP-010 (review-gate opt-in via config + 900s timeout) | PASS | `readCodexReviewGateEnabled` (fail-CLOSED default off, `TestReadCodexReviewGateEnabled_ConfigBranches`), `config.DefaultCodexReviewGateTimeout = 900 * time.Second` (defaults.go), ALLOW/BLOCK contract via `TestReviewGate_CodexPassAllows` / `_CodexFailBlocks` |

**Verification commands + results:**

- `go build ./...` → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- `go vet ./internal/cli/ ./internal/config/` → exit 0
- `golangci-lint run --timeout=3m ./internal/cli/ ./internal/config/` → 0 issues
- `go test -count=1 ./internal/config/` → `ok` (config foundation; the new `Codex` field + 900s const do not regress the section loaders)
- `go test -count=1 ./internal/cli/` → `ok` (full suite; the 3 hook ledger tests updated for the new subcommand: `TestHookCmd_SubcommandCount` 40→41, `TestHookCmd_PrePushSubcommandCount` 40→41, `TestHookValidEventTypes_AllHaveSubcommands` `utilitySubcmds += codex-review-gate`)
- M2 per-file coverage (`go test -coverprofile`): `codex_review_gate.go` — HandleCodexReviewGate 100% / isBlockVerdict 100% / hasReviewableChanges 100% (real temp git repo) / reviewableFromPorcelain 83.3% / isRuntimeManagedPath 100%; `mcp_codex.go` — handleCodexAudit 100% / handleCodexSetup 100% / runCodexReviewRPC 83.3% / runCodexReview (realCodexRunner via /bin/cat) 100% / classifyCodexAuth 100% / readCodexReviewGateEnabled 90%. Aggregate of the two M2 files well above the 85% threshold.
- E4 subagent boundary: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/mcp_codex.go internal/cli/codex_review_gate.go` (excl. tests/comments) → 0 (REQ-MCP-014)
- E8 RED (verbatim, pre-GREEN): captured to `.moai/state/verify/mcp-m2/red_codex_audit.txt` + `red_review_gate.txt` — `undefined: ReviewOutput / codexCommandRunner / codexRunner / codexLookPath` and `undefined: HandleCodexReviewGate / reviewGateChangeDetector` → `FAIL [build failed]`

**Self-gate design (AC-MCP-009, decision recorded):** the review gate ALLOWs immediately when (1) the config toggle is off (opt-in default-off), (2) `stop_hook_active` is true (mandatory loop prevention, mirroring `stopHandler`), or (3) the working tree has NO reviewable uncommitted change (`git status --porcelain` filtered to exclude `.moai/state|cache|reports|logs|harness` + `.claude/agent-memory` runtime paths — without this filter, per-turn hook-written state drift would trip the self-gate on every Stop and defeat AC-MCP-009). Only an uncommitted *source* change proceeds to a codex review; codex pass/inconclusive/missing ⇒ ALLOW, codex FAIL ⇒ the gate's sole BLOCK path. A turn that produced no code edit leaves no new source change ⇒ ALLOW (the AC-MCP-009 named case). Fail-open ALLOW on missing/erroring codex (a missing reviewer must not trap the session).

**M2-scope decision (deferred to M3/M4):** (a) the full 3-way `audit_model` claude-fallback + GLM backend + `${VAR}` secret-hygiene plumbing is M3 (REQ-MCP-009/010/011/012/013/014) — M2 implements ONLY the codex_backend fail-open preview (`VerdictInconclusive` on missing codex). (b) The settings.json TEMPLATE registration of `codex-review-gate` (with the 900s timeout override) + the Template-First `.mcp.json` reversal is M4 (REQ-MCP-016/018); M2 ships the handler + the `handle-codex-review-gate.sh` wrapper + the subcommand, tested via direct `moai hook codex-review-gate` invocation. (c) `codex_setup` is read-only (REPORTS `enable_review_gate`); mutating the toggle is a wizard-owned concern (M4).

_Run-phase NOT complete — M1+M2 are the first two of M1–M4; §E.3 stays pending until all run milestones land._

### M3 — GLM audit backend + 3-way config + auth + fail-open

**Deliverables:** `internal/cli/mcp_glm.go` (NEW — `glm_audit` handler + z.ai direct `/v1/messages` call via reusable `DefaultGLMBaseURL` + `loadGLMKey`, injectable HTTP/key/project seams, `ReviewOutput` parse, full fail-open), `internal/cli/mcp_audit.go` (NEW — 3-way `activeAuditBackend` selection + `multi` token-only flag + `buildAuditEnvBlock` secret-hygiene `${VAR}`-literal builder), `internal/cli/mcp_server.go` (register `glm_audit` + `WithOutputSchema[ReviewOutput]`), `internal/config/types.go` + `internal/config/defaults.go` (NEW `AuditConfig`/`AuditGates` + `audit_model`/`audit_gate` enum constants + locked default profile claude+codex required / glm advisory), `internal/cli/mcp_glm_test.go` + `internal/cli/mcp_audit_test.go` + `internal/config/mcp_audit_config_test.go` (NEW — TDD coverage, NO real GLM key/network — HTTP doer + key loader stubbed).

**GLM backend (design.md §3 M3 / §G.4 / REQ-MCP-009):** `glm_audit` POSTs the audit prompt to the Anthropic-compatible `https://api.z.ai/api/anthropic/v1/messages` endpoint DIRECTLY — reusing the SAME credential (`loadGLMKey` → `~/.moai/.env.glm`) and endpoint (`config.DefaultGLMBaseURL`) the GLM backend uses elsewhere. NOT the z.ai MCP server, NOT any gateway. The audit system prompt constrains the model to emit ONLY a `ReviewOutput` JSON, parsed + schema-validated before use (LLM05 defensive output validation). Output adopts the SHARED `review-output.schema.json` (`ReviewOutput`/`Finding`/`VerdictInconclusive` reused verbatim from mcp_codex.go) so orchestrator translation is uniform (§G.4 / DQ-2). Fail-open mandatory (C2 / REQ-MCP-012): missing/unauthenticated (401)/erroring/malformed GLM ⇒ structured `VerdictInconclusive` ⇒ claude fallback, never a hard error.

**3-way config + secret hygiene (REQ-MCP-010/011/013/014):**
- `audit_model` ∈ `{claude, codex, glm, multi}`; per-auditor `audit_gate` ∈ `{off, advisory, required}`, default `required`. Default profile: claude + codex required, glm advisory (user-enabled, so a distributed user without a GLM key is never hard-blocked — C2).
- `multi` is a DECLARED token only (`multiConvergenceImplemented = false`); convergence logic is SPEC-AUDIT-MULTI-MODEL (AP-8). `activeAuditBackend` accepts the token but M3 does NOT orchestrate fan-out/synthesis.
- Model/effort SSOT (AC-MCP-015): `resolveGLMAuditModel` calls `template.ResolveAgentModelEffort(llm, "sync-auditor")` ONLY (the SSOT) + `template.IsGLMBackend`; NEVER reads agent frontmatter or `llm.agent_overrides` directly (grep-verified: 0 matches).
- Secret hygiene (AC-MCP-013, load-bearing): `buildAuditEnvBlock` emits ONLY `${GLM_API_KEY}` / `${CODEX_API_KEY}` LITERAL constants — a resolved secret is NEVER serialized. The committed moai entry (`buildMoaiMCPServerEntry`) carries NO env block at all (the local stdio server reads keys via `loadGLMKey` at runtime). Negative test (`TestBuildAuditEnvBlock_SecretHygiene_NegativeTest`) injects a real-looking secret via the key seam and asserts it never appears in the marshaled env/entry.

**AC matrix (M3):**

| AC | Status | Evidence |
|---|---|---|
| AC-MCP-011 (GLM direct z.ai API call) | PASS | `TestGLMAudit_StubReturnsVerdict` + `TestGLMAudit_HitsZaiEndpointDirectly` (URL prefix `https://api.z.ai/api/anthropic/v1/messages`, NOT `/api/mcp/`), reuses `DefaultGLMBaseURL` + `loadGLMKey` |
| AC-MCP-012 (audit_model + audit_gate enums + default profile) | PASS | `TestAuditModel_EnumValues` / `TestAuditGate_EnumValues` / `TestAuditConfig_DefaultProfile` (claude+codex required, glm advisory) / `TestAuditConfig_YAMLRoundTrip` (incl. `multi` token) |
| AC-MCP-013 (secret hygiene — `${VAR}` literal, no serialization) | PASS | `TestBuildAuditEnvBlock_SecretHygiene_NegativeTest` (4 sub-tests: glm literal, committed entry no-env, multi both-literals, claude nil) — resolved key never leaks |
| AC-MCP-014 (fail-open on missing/unauth) | PASS | `TestGLMAudit_FailOpenOnMissingKey` / `_OnHTTPError` / `_OnMalformedResponse` / `_OnUnauthenticatedStatus` (401 ⇒ VerdictInconclusive) |
| AC-MCP-015 (model/effort SSOT via ResolveAgentModelEffort) | PASS | `TestResolveGLMAuditModel_SSOT` + grep `AgentOverrides\|agentfm\|Frontmatter` on mcp_glm/mcp_audit/mcp_codex.go = 0 (codex delegates model to the codex binary; glm resolves via the SSOT) |
| AC-MCP-016 (subagent boundary — structured result, no AskUserQuestion) | PASS | grep `AskUserQuestion\|mcp__askuser` on M3 handler files (excl. tests/comments) = 0; `TestMCPAudit_NoAskUserQuestion` package guard |
| AC-MCP-017 (3-way selection resolves single backend; `multi` token accepted) | PASS | `TestActiveAuditBackend_SingleBackends` (claude/codex/glm each → self) + `_MultiTokenAccepted` (`multi` stored, `multiConvergenceImplemented=false`) + `_RejectsUnknown` |

**Verification commands + results:**

- `go build ./...` → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- `go vet ./internal/cli/ ./internal/config/` → exit 0
- `golangci-lint run --timeout=3m ./internal/cli/ ./internal/config/` → 0 issues (after 2 errcheck fixes: `defer func(){ _ = resp.Body.Close() }()` + `_ = req.Body.Close()`)
- `go test -count=1 ./internal/cli/ ./internal/config/` → `ok` (full suite, exit 0; M1/M2 suites unaffected)
- M3 per-file coverage (`go tool cover -func`): `mcp_audit.go` — activeAuditBackend 100% / buildAuditEnvBlock 80%; `mcp_glm.go` — handleGLMAudit 100% / callGLMAudit 85.7% / parseGLMReview 80% / resolveGLMAuditModel 70% / glmAuditSystemPrompt 100% / glmAuditUserPrompt 75% / glmInconclusive 100% / reviewToolResult 75%. Aggregate of both M3 files above the 85% threshold (every function >0%; the lowest functions' uncovered branches are the marshal-can't-fail degrade + the GLM-backend llm.yaml branch).
- E4 subagent boundary: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/mcp_glm.go internal/cli/mcp_audit.go internal/cli/mcp_server.go internal/cli/mcp_codex.go` (excl. tests/comments) → 0 (REQ-MCP-014)
- E8 RED (verbatim, pre-GREEN): captured to `.moai/state/verify/mcp-m3/red_config.txt` (`undefined: AuditModelClaude / AuditGateOff / Workflow.Audit`) + `.moai/state/verify/mcp-m3/red_glm_audit.txt` (`undefined: glmHTTPDoer / glmKeyLoader / glmHTTPClient / activeAuditBackend / buildAuditEnvBlock / multiConvergenceImplemented`) → `FAIL [build failed]`

**M3-scope decisions (deferred to M4):** (a) the init/web selection UI (`audit_model` + per-auditor `audit_gate` in `Page3Questions` + web console) is M4 (REQ-MCP-015 / AC-MCP-020/021); M3 ships ONLY the typed config + the resolver. (b) The Template-First `.mcp.json` reversal (settings-management.md:33) + the §25 CI guard + `make build` is M4 (REQ-MCP-016/018); M3 ships the secret-hygiene `buildAuditEnvBlock` helper + the unchanged (env-free) `buildMoaiMCPServerEntry` so the committed entry stays secret-free. (c) `multi` convergence (parallel fan-out, disagreement synthesis) is SPEC-AUDIT-MULTI-MODEL (AP-8); M3 accepts the token, does not orchestrate.

### M4 — init/web selection UI + Template-First reversal + settings.json hook

**Deliverables:** `.claude/rules/moai/core/settings-management.md` + `internal/template/templates/.claude/rules/moai/core/settings-management.md` (line 33 reversal, byte-identical local↔template), `internal/template/templates/.claude/settings.json.tmpl` + `internal/template/templates/.claude/hooks/moai/handle-codex-review-gate.sh` (NEW — §25-neutral codex-review-gate Stop-hook registration, 900 s timeout), `internal/cli/wizard/{questions,types,wizard,translations,expansion_test,wizard_test,mcp_audit_test}.go` + `internal/cli/{init,init_audit_test}.go` (audit selection on page 3 "Audit & MCP" + applyWizardPage3ToOpts mapping + mcp_tools_opt_in → M1 provisioning seam), `internal/core/project/{initializer,initializer_expansion,initializer_audit_test}.go` (InitOptions audit fields + writeWorkflowAuditYAML persistence), `internal/settings/schema_sections.go` + `internal/web/{assets/i18n.js,mcp_audit_surface_test.go}` (web console schema surface + 4-locale i18n + no-fork/SSOT guard).

**AC matrix (M4):**

| AC | Status | Evidence |
|---|---|---|
| AC-MCP-018 (reversal applied, neutral) | PASS | settings-management.md:33 reversed on BOTH local + template (diff-clean, `grep 'provisions exactly one local MCP server'` =1 each; old phrase =0); prose carries no SPEC-ID/SHA/date/macOS-path |
| AC-MCP-019 (template source + CI guard + make build in one change) | PASS | `make build` → exit 0 (catalog.yaml regenerated, binary rebuilt); `go test -run TestTemplateNoInternalContentLeak|TestTemplateNeutralityAudit ./internal/template/` → ok (§25 green); single commit `987a2b576` carries template source + CI guard evidence + build |
| AC-MCP-020 (init wizard selection surface) | PASS | `TestPage3Questions_AuditSelectionSurfaced` + `TestAuditModelQuestion_UsesM3EnumVocabulary` + `TestAuditGateQuestions_{UsesM3EnumVocabulary,DefaultProfile}` + `TestOptInFlags_DefaultOff` + `TestApplyWizardPage3ToOpts_AuditSelection` (wizard→opts flow) |
| AC-MCP-021 (web console selection surface, identical interpreter) | PASS | `TestSchemaSurfaces_AuditSelection` (4 audit fields in Workflow schema) + `TestWebConsole_AuditNoForkedInterpreter` (0 `activeAuditBackend` in internal/web) + `TestWebConsole_ResolveAgentModelEffortSSOTShared` (0 redefinition — SSOT shared) |
| AC-MCP-023 (env/threshold constants — hardcoding prevention) | PASS | M4 reuses M3 `AuditModel*`/`AuditGate*` + M2 `DefaultCodexReviewGateTimeout`; 0 new `os.Getenv` literals in new non-test code; `.mcp.json` entry generic (M1 `TestBuildMoaiMCPServerEntry_Neutral`); config hardcoding guards green |
| AC-MCP-024 (full `go test ./...` + §25 CI guard green) | PASS | `go test ./...` → 0 FAIL (all `ok`/`?`); `go test ./internal/template/` → ok (§25 green) |

**Verification commands + results:**

- `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- `go vet ./internal/{cli,cli/wizard,core/project,config,settings,web}/` → exit 0
- `golangci-lint run --timeout=3m` (6 touched packages) → 0 issues (after 2 errcheck fixes: `_ = fmt.Fprintf`/`Fprintln` in the mcp_tools_opt_in provisioning path)
- `go test ./...` → 0 FAIL (full suite, exit 0; evidence at `.moai/state/verify/mcp-m4/full_suite.txt`)
- E4 subagent boundary: `grep -rn 'AskUserQuestion\|mcp__askuser'` on new M4 non-test code → 0 (REQ-MCP-014)
- E8 RED (verbatim, pre-GREEN): captured to `.moai/state/verify/mcp-m4/red_wizard_audit.txt` (`page-3 question "audit_model" is missing (AC-MCP-020)` × 6 IDs) → `FAIL`
- M4 per-file coverage (`go tool cover -func`): `writeWorkflowAuditYAML` 89.5% / `auditLeaves` 100% / `auditAnyPatched` 100% / `buildAuditBlockLines` 100% / `buildFreshWorkflowAuditBlock` 75% / `insertWorkflowMultiLineBlock` 92.3% — headline writer above the 85% threshold.

**Opt-in default-off preserved (C6):** the `.mcp.json` provisioning (`mcp_tools_opt_in` default false → no write) + the codex-review-gate (registered-but-dormant; `workflow.codex.review_gate.enabled` default off → M2 handler self-gates ALLOW) both ship INERT to distributed users. The reversal provisions the CAPABILITY; actual `.mcp.json` write per-project is opt-in.

**G-M4 self-check:** settings-management.md:33 reversed + neutral (local == template, diff-clean) · template source + §25 CI guard + `make build` in one change (embedded assets carry the reversal; `go test ./internal/template/` green) · init + web selection surfaces functional · full `go test ./...` + §25 neutrality green · opt-in default-off preserved.

## §E.3 Run-phase Audit-Ready Signal

- run_status: ready
- run_complete_at: 2026-08-06T03:55:00Z
- run_commit_sha: 987a2b576 (M4 — final run-phase milestone; M0-M4 all committed + pushed to feat/spec-moai-mcp-server)
- ac_pass_count: 24
- ac_fail_count: 0
- preserve_list_post_run_count: 0
- new_warnings_or_lints_introduced: 0
- total_run_phase_files: 19 (M4) · 5 commits (M0-M4) on the feat branch


## §E.4 Sync-phase Audit-Ready Signal

- sync_status: ready
- sync_complete_at: 2026-08-06T04:31:33Z
- sync_commit_sha: 93b7adf84 (PR #1378 squash-merge onto main; 3-phase close, merged 2026-08-06 by app/github-actions auto-merge)
- changelog_entry_position: [Unreleased] / Added (first entry, above SPEC-PROJECT-NAVIGATOR-003)
- ac_count_at_sync: 24/24 PASS (acceptance.md SSOT)
- frontmatter_status_transition: spec.md in-progress → completed (2026-08-06)
- canary_compliance_check:
  - go test ./... exit 0 (run-phase §E.3, pre-sync trust-but-verify)
  - §25 template neutrality green (M4 AC-MCP-019)
  - B12 self-test: `grep -c 'SPEC-MOAI-MCP-SERVER' CHANGELOG.md` = 0 pre-emit (no duplicate)

## §F Phase 4 Mode Selection

**Input parameters:**
- tier: L
- scope: multi-surface (Go MCP server `internal/cli/mcp_server.go` + ~9-10 tool handlers + codex/glm backends + `.mcp.json` template + CI guard + init wizard + web console; >15 files)
- domain count: ≥6 (cli, mcp, codex, glm, template, config)
- file language mix: Go (primary) + markdown + YAML + shell
- concurrency benefit: LOW (coding-heavy new-code) — Anthropic coding-task parallelism caveat

**Mode evaluation:** M1 trivial=no · M2 background=no(coding) · M3 agent-team=RETIRED · M4 parallel=no(coding-heavy+multi-domain→M5 per §B.2 tie-breaker) · M5 sub-agent=**selected** · M6 workflow=no(semantic new-code, not mechanical-uniform transform)

**Decision: sub-agent (Mode 5)**

**Justification:** Coding-heavy multi-surface foundation work. Per Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research"), the sequential sub-agent path (Mode 5) is the correct default for coding tasks; the multi-domain breadth (≥6) does not override the coding-heavy signal (§B.2 tie-breaker: coding-heavy + multi-domain → Mode 5). manager-develop implements milestones sequentially. Progression mode: **semi-autonomous** (per-milestone checkpoint at each G-Mn per user Implementation Kickoff Approval choice, 2026-08-05).

**Mode 6 confirmation:** N/A (Mode 5 selected; Implementation Kickoff Approval passed 2026-08-05; preferences drained at the gate).

## §G M0 Decision Lock

> Run-phase-owned decision stabilization (G-M0 gate). Records the highest-reversibility decisions isolated by plan.md §F M0 so human review focuses on them first. Body content of spec/plan/acceptance/research/design.md is NOT duplicated here — cross-references cite the plan-phase SSOT. M0 only; M1-M4 evidence lands in §E.2.

### G.1 — SDK choice + version pin

- **Decision:** `github.com/mark3labs/mcp-go` pinned at **`v0.57.0`** (latest stable as of 2026-08-05; the latest tag from `go list -m -versions`).
- **go.mod state:** `github.com/mark3labs/mcp-go v0.57.0 // indirect` — intentionally `// indirect` because M0 adds the dependency but no code imports it yet (M0 is decisions + go.mod only, no server code). M1's `internal/cli/mcp_server.go` import + `go mod tidy` will promote it to a direct dependency. Removing the `// indirect` marker now would falsely claim a direct import; leaving it is the honest M0 state. go.sum carries both the `h1:` content hash and the `/go.mod` hash.
- **Reversal cost:** bounded by design C1 (same-core-two-surfaces). Swapping to `modelcontextprotocol/go-sdk` re-touches only the handler-shape layer in `internal/cli/mcp_server.go`; the `internal/` core is untouched (design.md §5, R3).

### G.2 — API-surface confirmation (matches design.md: YES)

design.md §1/§5 hedged that exact SDK symbols are a run-phase detail ("this plan-phase design does not assert specific symbols beyond the SDK's purpose"). M0 confirms them against the v0.57.0 source. **No material divergence** — the thin-handler assumption (C1) holds.

Confirmed API surface (symbols verified in `go env GOMODCACHE`/`github.com/mark3labs/mcp-go@v0.57.0`):

| Concern | Confirmed symbol | design.md assumption |
|---|---|---|
| Server creation | `server.NewMCPServer(name, version string, opts ...ServerOption) *MCPServer` (`server/server.go:640`) | thin stdio JSON-RPC server (REQ-MCP-001) ✓ |
| stdio transport (blocking) | `server.ServeStdio(s)` convenience wrapper (README:55,193,229); `server.NewStdioServer(s)` + `.Listen(...)` for finer control (`server/stdio.go:369,504`) | modeled on `goal.go` blocking-`RunE` stdio pattern ✓ |
| Tool declaration + JSON Schema | `mcp.NewTool(name string, opts ...ToolOption) Tool` (`mcp/tools.go:846`); `mcp.WithString`/`WithNumber`/`WithBoolean`/`WithInteger`/`WithObject`/`WithArray`/`WithAny` property helpers; `mcp.WithStringEnumItems`/enum helpers for enum-valued params | `tools/list` declares name + JSON Schema (REQ-MCP-004) ✓ |
| Typed/raw input schema | `mcp.WithInputSchema[T any]()` (struct→schema), `mcp.WithRawInputSchema(json.RawMessage)` (`mcp/tools.go:913,956`) | type-safe invocation, zero flag-guessing ✓ |
| Structured result schema | `mcp.WithOutputSchema[T any]()`, `mcp.WithRawOutputSchema(json.RawMessage)` (`mcp/tools.go:964,992`) | the `review-output.schema.json` result shape (REQ-MCP-006) ✓ |
| Handler registration | `s.AddTool(tool Tool, handler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error))` (README:52,157,441) | thin wrapper over verified `internal/` functions (REQ-MCP-003/005) ✓ |
| Result shaping | `mcp.NewToolResultText`, `NewToolResultJSON[T]`, `NewToolResultStructured`, `NewToolResultError` (`mcp/utils.go:301,313,332,423`) | structured results incl. `VerdictInconclusive` fail-open (REQ-MCP-012, C2) ✓ |
| Read-only hint annotation | `mcp.WithReadOnlyHintAnnotation(bool)` (`mcp/tools.go:1015`) | read-only core tools (`session_list`, `goal_status`, `spec_audit`, etc.) ✓ |

**Conclusion: YES — the `moai mcp-server` can be authored as a thin handler calling existing `internal/` functions, with JSON Schema for every tool's input AND result, per design.md §1.** No blocker; M1 may proceed against this confirmed surface.

### G.3 — Locked `audit_model` / `audit_gate` config schema

Locked per REQ-MCP-010 / AC-MCP-012 (confirmed implementable; the SDK's enum helpers make these trivially declarable):

- **`audit_model`** ∈ `{claude, codex, glm, multi}` — single-backend selection. `multi` is a declared-but-not-orchestrated token here; convergence logic is owned by the future SPEC-AUDIT-MULTI-MODEL (spec.md §B Out-of-scope, AP-8).
- **per-auditor `audit_gate`** ∈ `{off, advisory, required}`, **default `required`**.
- **Default profile:** `claude=required`, `codex=required`, `glm=advisory` (user-enabled).
- Semantics: `off` → skip that auditor; `advisory` → surfaced as systemMessage (no block); `required` → must PASS before convergence/merge (block).

### G.4 — Adopted `review-output.schema.json` shape

Locked per REQ-MCP-006 / design.md §3 §4 (confirmed implementable via `mcp.WithOutputSchema[T]()` + `mcp.NewToolResultJSON`):

```
{
  "verdict":   "<verdict string>",
  "summary":   "<string>",
  "findings": [
    { "severity": "...", "title": "...", "body": "...",
      "file": "...", "line": <int>, "confidence": <float>,
      "recommendation": "..." }
  ],
  "next_steps": [ "..." ]
}
```

Shared shape across the codex (`codex_audit`) and GLM (`glm_audit`) backends so orchestrator translation is uniform (design.md DQ-2 resolved: normalize, not adopt codex-plugin-cc verbatim). The fail-open `VerdictInconclusive` value rides the same `verdict` field as a structured (non-error) result, never a hard error (C2).

### G.5 — Locked Template-First reversal reference

Locked per REQ-MCP-016 / AC-MCP-018 (M0 locks the reference; M4 authors the actual reversal prose + template mirror + CI guard). Do NOT duplicate the prose here — the reversal surface is:

- **Reversed file:** `.claude/rules/moai/core/settings-management.md:33` — current statement "MoAI-ADK no longer ships or provisions MCP servers via `.mcp.json`" → reversed to provision exactly ONE local server.
- **Neutral `.mcp.json` entry (locked shape, design.md §6.2):** `{"mcpServers":{"moai":{"command":"moai","args":["mcp-server"]}}}` — `command: "moai"` (PATH-resolved, no absolute path), no env block serializing secrets, no SPEC-ID/SHA/internal-date/macOS-path. Passes §25 neutrality by construction.
- **Same-change obligations (AC-MCP-019):** template source mirror (`internal/template/templates/`) + CI guard (`template-neutrality-check.yaml` / `internal_content_leak_test.go`) + `make build` move together; opt-in default-off preserved (REQ-MCP-002/C6).

### G.6 — M0 verification snapshot (G-M0 gate evidence)

- **Cross-platform build:** `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (both pre- and post-pin).
- **Lint:** `golangci-lint run --timeout=2m` → 0 issues (clean baseline; no NEW issues from the go.mod change).
- **`go vet ./...`:** exit 0 (post-pin).
- **Coverage:** N/A for M0 — no new logic; dependency pin + decision documentation only.
- **Branch/HEAD:** `feat/spec-moai-mcp-server` @ `05acc20ce` (origin/main `f16d8812e` + plan-phase commit); M0 commits stack on top.
