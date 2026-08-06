# Research — SPEC-MOAI-MCP-SERVER-001

> Codebase integration verification. Every file:line reference the design report (`.moai/reports/moai-autonomy-workflow-redesign-20260803.html`, dated 2026-08-03) relies on was re-verified against the CURRENT tree at run-start (`b57de3ab1`, branch `feat/spec-moai-mcp-server`, 2026-08-05). Drifts are recorded explicitly so no requirement rests on a stale location. This is the `research.md` sibling required when a SPEC touches existing code (skill § Red Flags).

## §1. Verification method

Each integration point below was verified by an executed `grep`/`Read` against the worktree at `b57de3ab1`. The commands were batched (agent-common-protocol § Parallel Execution) and their verbatim output drove this table. Lines cited as `file:line` reflect the CURRENT tree; where the report's citation differs, the drift is named. Run-phase MUST re-confirm these before forking code (some will drift again).

## §2. Integration-point inventory (verified)

| # | Integration point | Report citation | Verified current location | Status |
|---|---|---|---|---|
| 1 | `spec.Audit` | `spec/audit.go:156` | `internal/spec/audit.go:156` `func Audit(opts AuditOptions) (*AuditResult, error)` | MATCH |
| 2 | `session.QueryActiveWork` | `registry.go:254` | `internal/session/registry.go:254` `func QueryActiveWork(optSpecID string) ([]Entry, error)` | MATCH |
| 3 | `goal.LoadGoal` | `state.go:43` | `internal/goal/state.go:57` `func LoadGoal(projectRoot, sessionID string) (*Goal, error)` | DRIFT +14 lines |
| 3b | goal arming | (implied `arm`) | NO single `Arm()` fn; CLI `arm` RunE at `internal/cli/goal.go:105`/`:110` composes `goal.NewGoal` (`schema.go:115`) + `goal.SaveGoal` (`state.go:76`) | SHAPE — wrap the composition |
| 4 | `verify.Load` + `RecordCheck` | `store.go:38` | `internal/verify/store.go:38` `Load`; `RecordCheck` at `:107` | MATCH |
| 5 | `mcp-server` subcommand | NEW + `root.go` AddCommand | `internal/cli/root.go`: `rootCmd` var `:18`, `AddCommand` calls `:143-:149`; `internal/cli/mcp_server.go` does NOT exist (NEW) | MATCH (NEW confirmed) |
| 6 | `cli/goal.go` pattern | (blocking RunE stdio) | `internal/cli/goal.go` exists; multiple `RunE` (`:86`, `:110`, `:119`, `:130`, `:145`); `armCmd` `:105` | MATCH (pattern reference) |
| 7 | `.mcp.json` atomic-config | `glm_tools.go` `mutateClaudeJSONAtomic`/`resolveConfigPath`/`buildMoaiMCPServerEntry` | `internal/cli/glm_tools.go`: `resolveConfigPath:362`, `buildZAIMCPEntry:391`, `mutateClaudeJSONAtomic:541`. `buildMoaiMCPServerEntry` does NOT exist → NEW, modeled on `buildZAIMCPEntry` | PARTIAL — helper is NEW |
| 8 | model/effort SSOT | `handlers.go:85` `ResolveAgentModelEffort` | **definition** `internal/template/profile_matrix.go:385` `func ResolveAgentModelEffort(cfg config.LLMConfig, agent string) (me config.ModelEffort, mapped bool)`. `internal/web/handlers.go:86` is a CALLER ("G3-1 repoint"), not the definition | DRIFT — def moved handlers.go→profile_matrix.go |
| 9 | env-var constants | `internal/config/envkeys.go` | exists; 47 constant lines | MATCH (target for new constants) |
| 10 | thresholds/defaults | `config/defaults.go` | `internal/config/defaults.go` exists | MATCH (target for new thresholds) |
| 11 | `loadGLMKey` pattern | (auth loader pattern) | `internal/cli/glm.go:905` `func loadGLMKey() string`; also `internal/hook/session_start.go:1042` `loadGLMKeyFromEnvFile` | MATCH (pattern for codex loader) |
| 12 | wizard page-3 | `wizard/questions.go` `Page3Questions` | package at `internal/cli/wizard/` (NOT `internal/wizard/`); `Page3Questions` ref `internal/cli/wizard/questions.go:15`; `WizardResult` `internal/cli/wizard/types.go:12` | DRIFT — package path |
| 13 | wizard→opts | `init.go applyWizardPage3ToOpts` | `internal/cli/init.go:164` `func applyWizardPage3ToOpts(cmd *cobra.Command, result *wizard.WizardResult, opts *project.InitOptions)` | MATCH |
| 14 | `runtime.AuditCache` | `audit_cache.go:43` | interface `internal/runtime/audit_cache.go:50`; `InMemoryCache.ComputeHash:110`/`Lookup:146`/`Store:157` | DRIFT +7 lines (interface) |
| 15 | reversal sentence | `settings-management.md:33` | `.claude/rules/moai/core/settings-management.md:33` "MoAI-ADK no longer ships or provisions MCP servers via `.mcp.json`." | MATCH (verbatim) |
| 16 | goal MaxTurns infinite | (INFINITE-GOAL dep) | `internal/goal/schema.go:36` comment "MaxTurns == 0 is the infinite entry point"; `evaluate.go:320` `> 0` guard; `DefaultMaxTurns=30` `schema.go:93` | SATISFIED — already landed |
| 17 | web board (read-only) | `internal/web/board.go` handleBoard | `internal/web/board.go:70` `handleBoard` serves `GET /specs` (read-only, refresh-required) | MATCH (out of scope — v3.1 live) |
| 18 | `IsGLMBackend` | (audit GLM-sensitivity) | `internal/template/glm_effort_overlay.go:189` `func IsGLMBackend(cfg config.LLMConfig) bool` | MATCH |
| 19 | z.ai endpoint | (GLM audit transport) | `https://api.z.ai/api/anthropic` (Anthropic-compatible), present in llm.yaml testdata + validated by `internal/config/validation_glm_baseurl_test.go` | MATCH |
| 20 | `review-output.schema.json` | (codex_audit output) | NOT present in tree → adopted as the contract shape (sourced from the codex-plugin-cc surface per report §3.4); M0 locks the shape | NEW — adopt as contract |

## §3. Environment facts

- **Module**: `github.com/modu-ai/moai-adk`; **Go**: `1.26.4` (`go.mod`).
- **MCP SDK in go.mod**: NONE present → adding `github.com/mark3labs/mcp-go` is the approved single new dependency (user decision; REQ-MCP-001).
- **gorilla/websocket in go.mod**: absent → confirms web live board is NOT in scope (v3.1).
- **Existing `.mcp.json`**: template source `internal/template/templates/` has NONE (consistent with the settings-management.md:33 statement to be reversed). The local worktree `.mcp.json` carries `context7` + `chrome-devtools` under a `mcpServers`/`staggeredStartup` schema — the NEW `moai` entry joins `mcpServers` and must preserve the existing `$schema` reference and `staggeredStartup` block.
- **Prior MCP art**: `SPEC-CC2122-MCP-001` (status: implemented) covered the `alwaysLoad` field — useful precedent for `.mcp.json` shape, but its frontmatter uses the REJECTED snake_case aliases (`created_at`/`updated_at`/`labels`); this SPEC uses the canonical 12 fields per `spec-frontmatter-schema.md`.

## §4. SDK choice — `mark3labs/mcp-go` (user decision, not re-litigated)

The user decision records `mark3labs/mcp-go` as the Go MCP SDK. The rationale recorded here is for `design.md` and for a future reversal-cost estimate (R3); the decision itself is not re-opened.

- **`mark3labs/mcp-go`** (chosen) — the widely-adopted community Go MCP SDK; provides the stdio JSON-RPC server, `tools/list` + `tools/call`, and JSON Schema input validation that REQ-MCP-001/004 need.
- **`modelcontextprotocol/go-sdk`** (official alternative) — the MCP org's maintained Go SDK; the principal alternative if the community SDK's API surface or maintenance cadence becomes a liability.
- **`metoro-systems/mcp-golang`** (legacy alternative) — older; not preferred.

**Reversal cost is bounded by design (C1)**: because the MCP layer is a thin wrapper over `internal/`, swapping the SDK re-touches only the handler-shape layer — the `internal/` core is untouched. M0 (plan.md §F) confirms the API surface and pins the version before any code forks. The exact API surface (`server.NewStdioServer` naming, tool-registration ergonomics, JSON Schema helper) is a run-phase detail confirmed at M0; this plan-phase research does not assert specific API symbols beyond the SDK's purpose.

> No open clarifications — the SDK decision, scope depth, and scope boundary are all settled by the team-lead's delegation prompt. The codename→canonical-ID normalization (`SPEC-MOAI-MCP-SERVER` → `SPEC-MOAI-MCP-SERVER-001`) is mechanical (regex-enforced) and is surfaced, not a blocker.

## §5. Template-First reversal — impact analysis

- The reversal touches `.claude/rules/moai/core/settings-management.md` (a distributed-rule file loaded into every session) AND the template source mirror AND the §25 CI guard. Per CLAUDE.local.md §2 / §25, all three move together with a `make build`.
- The new `.mcp.json` entry is **neutral by construction**: `{"command":"moai","args":["mcp-server"]}` contains no SPEC-ID, no commit SHA, no internal date, no macOS-bias path — it passes §25 template-neutrality. The CI guard (`template-neutrality-check.yaml` trigger on path change + `internal_content_leak_test.go`) is the safety net; M4 authors the reversal to pass it on the first try (AC-MCP-019).
- The local-only `.mcp.json` (context7 + chrome-devtools) is unaffected; the `moai` entry is additive under `mcpServers`.

## §6. Fail-open + secret-hygiene design verification

- **Fail-open (C2/REQ-MCP-012)**: the MCP tool boundary CANNOT call `AskUserQuestion` (subagent boundary, REQ-MCP-014), so a missing/unauthenticated auditor returns `VerdictInconclusive` as a structured tool RESULT; the orchestrator translates that result through its own `AskUserQuestion` channel if a user decision is needed. The fallback to claude is in-process (claude is always available — it is the session backend).
- **Secret hygiene (C3/REQ-MCP-011)**: the git-tracked `.mcp.json` is a project-scope file; serializing `CODEX_API_KEY`/`GLM_API_KEY` values would leak secrets. The `${VAR}` literal form is expanded by the Claude Code runtime at load time, so the committed file carries only the literal. `mutateClaudeJSONAtomic` writes the literal; the resolved value never enters project scope. The `loadGLMKey` pattern (`glm.go:905`) reads the resolved value from `~/.moai/.env.*` at runtime, user-scope.

## §7. Cross-references

- Design report: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.4 + §3.6 (integration map at §3.6 "내장 통합맵 (C1 실측)").
- spec.md §C (REQ-MCP-001..018), §E (constraints C1-C8), §G (risks R1-R4).
- plan.md §B (drift catalogue — same data, plan-facing), §F (milestones), §G (anti-patterns).
- design.md §3 (full tool table), §4 (auth/gate/fail-open state machine), §5 (dependency-choice rationale), §6 (Template-First reversal plan).
- Schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md` (frontmatter), `internal/spec/lint.go:715` (`specIDPattern`).
