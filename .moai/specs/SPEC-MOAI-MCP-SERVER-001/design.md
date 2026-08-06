# Design — SPEC-MOAI-MCP-SERVER-001

> Architecture for the moai self-hosted MCP server. Covers: the same-core-two-surfaces invariant, the full tool table, the auth/gate/fail-open state machine, the dependency-choice rationale (mark3labs/mcp-go), and the Template-First reversal plan. This is a WHAT/WHY design doc — implementation (HOW: function names, struct shapes, exact SDK symbols) is deferred to run-phase per the SPEC scope boundary (spec.md §B). Integration points are verified in `research.md`.

## §1. Architecture — same-core-two-surfaces

The MCP server is NOT a separate core. It is a thin JSON-RPC transport over the SAME `internal/` functions the CLI Cobra handlers call. This is the industry-standard pattern (Anthropic's own context7 / playwright MCP; Codex ships both `codex` CLI and `codex mcp-server`).

```
   internal/spec · internal/session · internal/goal · internal/verify · internal/runtime
              (single core logic — single source of truth)
                              │
            ┌─────────────────┴─────────────────┐
            ▼                                   ▼
   moai CLI (Cobra)                    moai MCP server
   humans · scripts · hooks            "moai mcp-server" stdio JSON-RPC
   "ad-hoc composition" (Bash)         AI orchestrators · external MCP clients
   already exists                      "declarative · repeatable · schema-safe"
                                       thin wrapper (same core reused)
```

**Why this invariant (C1):**
1. **Zero drift** — the CLI and MCP surfaces cannot diverge because they call the same function.
2. **Swappable SDK (R3)** — the transport layer (mark3labs/mcp-go) is the only thing that touches the SDK; swapping it re-touches only the handler-shape layer.
3. **Minimal attack surface** — only the declared tool set is callable; no arbitrary `moai <subcommand> --flag` combinations (the safety net the fully-autonomous tier needs).
4. **Lower repeated-call cost** — a long-running stateful process absorbs the per-call cold-start + config-load tax that CLI Bash invocation pays every call (report §1.4 / §3.6 comparison table).

**Handler responsibility (what a thin wrapper does):** map JSON-RPC params → the `internal/` function's args; call the function; shape the result into the tool's declared JSON Schema. It does NOT re-implement, branch, or cache core logic. (Anti-pattern AP-1.)

## §2. CLI Bash vs MCP — when each wins

(From report §3.6; recorded here so the design does not over-claim MCP as a replacement.)

| Axis | CLI Bash call (today) | moai MCP server (this SPEC) |
|---|---|---|
| Discovery / schema | AI guesses flags from `--help` | `tools/list` declares name + JSON Schema — type-safe |
| Result parsing | stdout text | structured JSON-RPC object |
| Per-call overhead | Go binary cold-start + config load every call | long-running process, one handshake |
| State | stateless | stateful possible (cache) |
| Attack surface | arbitrary flag combos | only declared tools (narrower) |
| Portability | moai + PATH | standard protocol — Cursor/Cline/Zed/other hosts |
| Flexibility / exploration | AI composes ad hoc (pipe/grep/jq) — wins for exploration | only exposed tools |
| Dev cost | mostly exists | thin Go wrapper (core reused) |

**Conclusion (recorded, not re-litigated):** ad-hoc/exploratory calls stay on CLI Bash; repeated/structured/autonomous/portable calls go through MCP. Both sit on the same core.

## §3. Full tool table (M1-M3 surface)

Every tool wraps a verified `internal/` function (research.md §2). "First CLI/MCP surface" marks tools that today have no CLI exposure.

### M1 — core read/status/trend tools

| Tool | Wraps (verified) | Direction | Notes |
|---|---|---|---|
| `session_list` | `session.QueryActiveWork` (`registry.go:254`) | read | active sessions, optional spec filter |
| `goal_status` | `goal.LoadGoal` (`state.go:57`) | read | armed-goal status for a session |
| `goal_arm` | `goal.NewGoal` (`schema.go:115`) + `goal.SaveGoal` (`state.go:76`) (the `cli/goal.go:110` arm RunE composition) | write | honors mode token + deny/ask rules (report §3.6) |
| `spec_progress` | SPEC scanner (the `/moai codemaps` / board `buildBoardView` path, `internal/web/board.go:88`) | read | SPEC lifecycle distribution |
| `verify_snapshot` | `verify.Load` (`store.go:38`) + `verify.RecordCheck` (`:107`) | read/write | **first CLI/MCP surface** for verify |
| `verify_trend` | `verify.Load` (trend view) | read | verify trend |
| `spec_audit` | `spec.Audit` (`audit.go:156`) | read | the `moai spec audit` wrapping |
| `spec_drift` | `spec.Audit` (drift view) | read | drift subset |
| `audit_cache` | `runtime.AuditCache` `ComputeHash`/`Lookup`/`Store` (`audit_cache.go:50`/`:110`/`:146`/`:157`) | read | 24h PASS cache (read-only reuse — speed optimization is out of scope) |

### M2 — codex audit backend (Phase 1)

| Tool | Implementation | Output schema |
|---|---|---|
| `codex_audit` | `codex` binary shellout; `mode` enum (native→`review/start`, adversarial→`turn/start`+adversarial-review prompt) | `review-output.schema.json` (`verdict`/`summary`/`findings[severity,title,body,file,line,confidence,recommendation]`/`next_steps`) |
| `codex_setup` | Go reimpl: `exec.LookPath("codex")` + `codex --version`; ChatGPT/apiKey/provider classification; `enable_review_gate` toggle | provider status object |
| (Stop hook) `moai hook codex-review-gate` | Go reimpl of the `.sh` concept; opt-in `workflow.codex.review_gate.enabled`; ALLOW/BLOCK; self-gate; 900 s timeout | hook decision JSON |

### M3 — GLM audit backend

| Tool | Implementation | Output schema |
|---|---|---|
| `glm_audit` (selected via `audit_model: glm`) | z.ai GLM API direct call (`https://api.z.ai/api/anthropic`, Anthropic-compatible; reuse existing z.ai client plumbing); NOT the z.ai MCP server | `review-output.schema.json` (same shape as codex for uniform orchestrator translation) |

### Out-of-scope tools (declared, NOT built here)

- `codex_task`, `codex_job_status/result/cancel`, `codex_transfer`, native Go JSON-RPC client → Phase-2 follow-up SPEC.
- `audit_model: multi` convergence orchestration → SPEC-AUDIT-MULTI-MODEL.

## §4. Auth / gate / fail-open state machine

### 4.1 Auditor selection (`audit_model` + per-auditor `audit_gate`)

```
audit_model ∈ {claude, codex, glm, multi}
  └─ single-backend selection (claude|codex|glm): exactly that backend runs (per its gate)
  └─ multi: token ACCEPTED, orchestration deferred to SPEC-AUDIT-MULTI-MODEL

per-auditor audit_gate ∈ {off, advisory, required}   (default: required)
  off       → skip that auditor entirely
  advisory  → result surfaced as systemMessage (NO block)
  required  → that auditor MUST PASS before convergence/merge (block)

default profile: claude=required, codex=required, glm=advisory (user-enabled)
```

### 4.2 Auth branching (per auditor)

```
claude → session backend (already authenticated; always available — the fallback)
codex  → OAuth material in ~/.moai/.env.codex  (loadGLMKey pattern; ~/.moai scope, NOT project)
glm    → API key in ~/.moai/.env.glm           (loadGLMKey pattern; ~/.moai scope, NOT project)
```

Secrets NEVER serialized into git-tracked project-scope `.mcp.json` — only `${CODEX_API_KEY}` / `${GLM_API_KEY}` literals, expanded by the Claude Code runtime at load time. (C3, REQ-MCP-011.)

### 4.3 Fail-open state machine (mandatory — C2/REQ-MCP-012)

```
                  ┌── auditor present + authenticated ──→ run audit ──→ verdict (PASS/FAIL/etc.)
auditor invoked ──┤
                  └── auditor MISSING or UNAUTHENTICATED
                        │
                        ▼
                    return VerdictInconclusive  (structured tool RESULT, not an error)
                        │
                        ▼
                    fall back to active auditor (claude)   ◀── always available
                        │
                        ▼
                    NEVER hard-block the workflow on a missing OPTIONAL dependency
```

Because the MCP tool boundary cannot call `AskUserQuestion` (REQ-MCP-014), `VerdictInconclusive` is returned as a structured result; the orchestrator translates it through its own `AskUserQuestion` channel if a user decision is required. (Cross-reference: `verification-claim-integrity.md` — evidence-of-absence is not evidence-of-failure in either direction; the fail-open path is explicit, not silent.)

### 4.4 Review-gate self-gate (M2, REQ-MCP-008)

```
Stop hook fires (workflow.codex.review_gate.enabled = true)
  │
  ├─ previous turn = NO code edit / status report / review-result  ──→ ALLOW immediately (self-gate)
  │                                                                   (prevents false blocks)
  └─ previous turn = code edit
        │
        ▼
     run codex_audit ──→ PASS ──→ ALLOW
                     └─→ FAIL ──→ BLOCK (ALLOW/BLOCK contract) ──→ orchestrator translates
```

## §5. Dependency-choice rationale — `mark3labs/mcp-go` (R3)

The user decision is `mark3labs/mcp-go`. This section records the rationale and the reversal cost so a future swap is cheap.

| Option | Role | Notes |
|---|---|---|
| **`mark3labs/mcp-go`** (CHOSEN) | community Go MCP SDK | provides stdio JSON-RPC server, `tools/list`+`tools/call`, JSON Schema input validation — the surface REQ-MCP-001/004 need |
| `modelcontextprotocol/go-sdk` | official maintained alternative | the principal fallback if the community SDK's API surface or cadence becomes a liability |
| `metoro-systems/mcp-golang` | older community alternative | not preferred |

**Why a SDK (not hand-rolled JSON-RPC):** the report floated a ~200-LOC hand-rolled stdio JSON-RPC option, but a SDK earns its keep on the JSON Schema validation + the initialize/tools-list/tools-call protocol shape + future transport (SSE/streaming) headroom. The thin-wrapper invariant (C1) keeps the SDK's blast radius to the handler layer only.

**Reversal cost:** bounded by C1. Swapping `mark3labs/mcp-go` → `modelcontextprotocol/go-sdk` re-touches only `internal/cli/mcp_server.go` + handler shape; the `internal/` core is untouched. M0 (plan.md §F) confirms the API surface and pins the version before code forks. The exact SDK API symbols (`server.NewStdioServer` naming, tool-registration ergonomics, JSON Schema helpers) are confirmed at M0 — this plan-phase design does not assert specific symbols beyond the SDK's purpose.

## §6. Template-First reversal plan (M4, REQ-MCP-016/018)

### 6.1 The reversal

`.claude/rules/moai/core/settings-management.md:33` currently states:
> "MoAI-ADK no longer ships or provisions MCP servers via `.mcp.json`."

This is REVERSED to provision exactly ONE local server. The replacement prose + the `.mcp.json` entry are **neutral by construction** (no SPEC-ID, no commit SHA, no internal date, no macOS-bias path) → §25 template-neutrality.

### 6.2 The neutral `.mcp.json` entry

```json
{
  "mcpServers": {
    "moai": {
      "command": "moai",
      "args": ["mcp-server"]
    }
  }
}
```

Generic: `command: "moai"` (PATH-resolved, no absolute path), `args: ["mcp-server"]`. No env block serializing secrets (the `${VAR}` literals, if needed, are added only when a backend is enabled, and never the resolved value). Joins the existing `mcpServers` (context7 / chrome-devtools are local-only, unaffected).

### 6.3 Same-change requirements (AC-MCP-019)

1. Template source mirror: `internal/template/templates/.claude/rules/moai/core/settings-management.md` (+ any template `.mcp.json` provisioner).
2. CI guard alignment: `.github/workflows/template-neutrality-check.yaml` + `internal/template/internal_content_leak_test.go` — confirm the new entry passes; update the guard rule set if needed.
3. `make build` regeneration (templates are embedded via `//go:embed all:templates` in `internal/template/embed.go`).
4. Opt-in default-off preserved (REQ-MCP-002/C6) — the reversal provisions the CAPABILITY; actual `.mcp.json` write per-project is opt-in.

### 6.4 Why this is safe for distribution

- Exactly ONE entry (no third-party bundling — that is SPEC-TREND-MCP).
- The `moai` binary is the user's own install; the entry invokes it with a fixed subcommand. No arbitrary-flag attack surface (the §3.6 "narrower attack surface" point).
- Opt-in: distributed templates ship the capability inert; users enable per-project.

## §7. `.mcp.json` provisioning flow (M1)

```
user opts in (moai init wizard / moai web / moai update flag)
  │
  ▼
resolveConfigPath(scope)                      ── glm_tools.go:362
  │
  ▼
mutateClaudeJSONAtomic(configPath, apply)     ── glm_tools.go:541 (flock + RMW; race-safe)
  │
  └─ apply: insert/merge the "moai" entry from buildMoaiMCPServerEntry()  ── NEW (model buildZAIMCPEntry:391)
        │
        └─ entry carries ONLY: command, args. NO secret values. ${VAR} literals only if a backend enabled.
  │
  ▼
atomic write (flock released)
```

Reuses the existing atomic-config infrastructure verbatim — only `buildMoaiMCPServerEntry` is NEW.

## §8. Model/effort resolution (SSOT — C4/REQ-MCP-013)

```
codex_audit / glm_audit need (model, effort)
  │
  ▼
template.ResolveAgentModelEffort(cfg config.LLMConfig, agent string)   ── profile_matrix.go:385
  │   (the SAME interpreter used by the web console handlers.go:86 + the launcher)
  ▼
(model, effort, mapped)
  │
  ▼
passed to the backend (codex shellout / glm API call)
```

MCP handlers MUST NOT read agent frontmatter or `llm.agent_overrides` directly (AP-4) — that forks the interpretation. `ResolveAgentModelEffort` is the single interpreter; the web console and launcher already use it, so MCP inherits identical resolution.

## §9. Open design questions (none blocking — recorded for transparency)

- **DQ-1**: Should `verify_snapshot`/`verify_trend` (first CLI/MCP surface for verify) also gain a CLI `--json` form in the same change, or is MCP-first acceptable? (Lean: MCP-first is acceptable; a CLI mirror is a separate small follow-up.) Not a blocker — the MCP surface is in scope regardless.
- **DQ-2**: The `review-output.schema.json` shape — adopt the codex-plugin-cc surface verbatim, or normalize field names for the GLM backend to share it? (Lean: normalize so orchestrator translation is uniform — design.md §3 records the shared shape.) Confirmed at M0.

## §10. Cross-references

- spec.md — requirements (REQ-MCP-001..018), constraints (C1-C8), risks (R1-R4).
- plan.md — milestones (M0-M4), critical path, gates, anti-patterns (AP-1..AP-8).
- acceptance.md — AC-MCP-001..024 + edge cases + Definition of Done.
- research.md — verified integration-point inventory (§2) + SDK rationale evidence (§4) + reversal impact (§5).
- Design source: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.4 + §3.6.
