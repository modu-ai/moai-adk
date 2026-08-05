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

_Run-phase NOT complete — M1 is the first of M1–M4; §E.3 stays pending until all run milestones land._

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

- sync_commit_sha: _pending-backfill_

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
