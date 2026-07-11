---
id: SPEC-HARNESS-MCP-PROVISION-001
title: "Per-project-type MCP server provisioning in /moai project + harness generation"
version: "0.1.2"
status: draft
created: 2026-07-11
updated: 2026-07-11
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/skills/moai/workflows/project, .claude/skills/moai/workflows/harness-builder.md, .moai/config/sections/mcp-matrix.yaml"
lifecycle: spec-anchored
tags: "project, mcp, provisioning, mcp-json, harness, template-mirror, project-harness-pipeline"
era: V3R6
tier: M
depends_on: [SPEC-PROJECT-HARNESS-BRIDGE-001]
---

# SPEC-HARNESS-MCP-PROVISION-001 — Per-project-type MCP server provisioning in /moai project + harness generation

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-07-11 | 0.1.0 | Initial plan-phase draft (Tier M, 11 REQ / 13 AC). SPEC 2 of the 3-SPEC "Project-Harness Pipeline" Epic; `depends_on: [SPEC-PROJECT-HARNESS-BRIDGE-001]` (reuses the `harness-spec.yaml` `external_systems` / `ui_surface` / `verification` fields introduced there). Two changes: (1) insert a new `/moai project` **Phase 3.6** between LSP detection (3.5) and auto dev-mode (3.7) that detects the stack, selects recommended MCP servers from an externalized matrix, obtains orchestrator-held AskUserQuestion approval, and writes project-scope `.mcp.json` entries additively; (2) extend the `harness-builder.md` GENERATE contract with an OPTIONAL artifact (a `.mcp.json` fragment via the existing `builder-harness` `artifact_type=mcp-server` capability), emitted ONLY when the harness PLAN declares MCP needs. Doc/config-only (markdown + yaml); no Go code. All Template-First. 2 clarification markers in plan.md (mcp-matrix config surface + doctor manifest-mcp validate-vs-tolerate). | manager-spec |
| 2026-07-11 | 0.1.1 | Plan-audit fixes applied (11 REQ / 15 AC). (D1) Both clarifications RESOLVED and marker literals struck from plan.md: mcp-matrix = standalone template DATA RESOURCE (`internal/template/templates/.moai/config/sections/mcp-matrix.yaml`, read as prose-context; NOT a Go config section); doctor manifest-mcp = TOLERATE-ONLY (zero Go change — lenient `json.Unmarshal`; regression-guarded by `DisallowUnknownFields == 0`). (D2) Added AC-HMP-014 covering the previously-uncovered REQ-HMP-003 (write-on-approval at project scope) → coverage claim now accurate. (D3) Replaced the vacuous repo-wide `moai harness doctor` smoke in AC-HMP-010 with a deterministic documented-tolerance grep + `DisallowUnknownFields == 0` regression guard. (D4/D5) Epic canonical numbering: MCP fragment renumbered from artifact 6 → **artifact 7 (OPTIONAL)** (verify skill from SPEC-HARNESS-VERIFY-PROMOTE-001 is the mandatory artifact 6); GENERATE contract reframed as "5 base + verify skill (6) + optional MCP fragment (7)"; added AC-HMP-015 asserting the `harness-builder.md` "exactly 5" wording is reconciled. Manifest `mcp` remains the 9th top-level field. | manager-spec |
| 2026-07-11 | 0.1.2 | AC verification hardening — §C rewritten so every AC discriminates (baseline-0 → post-change ≥1); 7 vacuous compound-alternation greps replaced. No REQ / AC-mapping / scope change. | manager-spec |

## §A. Context and Intent

The `/moai project` flow interviews the user, generates project docs, detects the
language toolchain, and generates a project-specific harness. Today there is **NO
MCP (Model Context Protocol) provisioning anywhere in that flow**. The only
tool-provisioning step is Phase 3.5 (LSP server detection / install). The flow
runs: `3.5 LSP → 3.7 auto dev-mode → 4.1a DB detection → 4 completion`. A new
project therefore ships without any of the per-project-type MCP servers that make
the downstream development loop productive (browser automation for a web frontend,
a read-only DB server for a backend, etc.).

Two structural gaps:

1. **No MCP provisioning step exists in `/moai project`.** The interview already
   elicits enough signal to recommend MCP servers — `SPEC-PROJECT-HARNESS-BRIDGE-001`
   made the stack intent machine-readable in `.moai/project/harness-spec.yaml`
   (`external_systems`, `ui_surface`, `verification`). But nothing consumes those
   fields to provision `.mcp.json`. This SPEC inserts **Phase 3.6** to do so.

2. **The harness Builder cannot emit MCP artifacts even though the capability
   already exists.** The `builder-harness` agent ALREADY supports
   `artifact_type=mcp-server` (it scaffolds `.mcp.json` entries with stdio / http /
   sse transports). But the v4 harness Builder GENERATE phase
   (`harness-builder.md`) documents its output as **5 base artifact types** (thin
   command / Runner JS / specialist agents / companion skills / manifest.json) — MCP
   is **not wired in**. Per the Epic canonical artifact order, the GENERATE contract
   grows to **5 base + verify skill (artifact 6, mandatory — owned by
   SPEC-HARNESS-VERIFY-PROMOTE-001) + optional MCP fragment (artifact 7, this SPEC)**.
   This SPEC extends GENERATE with an OPTIONAL **artifact 7** that reuses the existing
   `artifact_type=mcp-server` capability, emitted only when the harness PLAN declares
   MCP needs. It also reconciles the `harness-builder.md` "exactly 5 artifact types"
   prose so the bare count is no longer left uncontextualized against the added
   artifacts (AC-HMP-015).

**Design premise (Anthropic-verified pattern).** MCP project scope is configured
via a checked-in `.mcp.json` at the repo root (per-user approval prompt on first
use), 3-5 servers maximum (token overhead grows beyond that), vendor-maintained
servers preferred (2026 MCP CVE surge), and secrets expressed via `${VAR}`
env-var expansion — never inlined. This SPEC realizes that pattern for both the
`/moai project` provisioning step and the harness generation step.

**Boundary principle.** This is SPEC 2 of a 3-SPEC Epic. It consumes the
`harness-spec.yaml` contract from `SPEC-PROJECT-HARNESS-BRIDGE-001` (the FOUNDATION
SPEC) and provisions MCP servers from it. It does NOT re-open the interview or the
`harness-spec.yaml` schema (owned by BRIDGE-001).

## §B. Scope Summary

**In scope**:
- Insert `/moai project` **Phase 3.6** (between LSP 3.5 and dev-mode 3.7) in
  `project/doc-generation.md`: detect stack (reuse existing language / framework
  detection + `harness-spec.yaml` `external_systems` / `ui_surface`) → select
  recommended MCP servers from the matrix → orchestrator-held AskUserQuestion
  approval → write `.mcp.json` entries at project scope.
- Externalize the MCP recommendation matrix (web / mobile / backend rows) to a
  standalone **data resource** `.moai/config/sections/mcp-matrix.yaml` — authored in
  the template tree at `internal/template/templates/.moai/config/sections/mcp-matrix.yaml`
  and read by `project/doc-generation.md` as prose-context (NOT a new Go config
  section; no typed loader / struct field) — NOT hardcoded in skill prose beyond a
  fallback pointer.
- Enforce hard caps: 3-5 servers recommended maximum; vendor-maintained preferred;
  per-server explicit AskUserQuestion approval for any credentialed server.
- Write `.mcp.json` additively / idempotently (merge, never clobber) at project
  scope; secrets in `${VAR}` env-var form (never inline a literal token).
- Extend `harness-builder.md` GENERATE with an OPTIONAL **artifact 7** (`.mcp.json`
  fragment via `builder-harness` `artifact_type=mcp-server`), emitted ONLY when the
  harness PLAN declares MCP needs (derived from `harness-spec.yaml`
  `external_systems` / `verification`); omitted otherwise (byte-identical to today).
  Reconcile the `harness-builder.md` "exactly 5 artifact types" prose to the canonical
  order (5 base + verify skill artifact 6 + optional MCP fragment artifact 7).
- `moai harness doctor` reference-integrity tolerates the optional manifest `mcp`
  field.
- Mirror all edits into `internal/template/templates/...` (Template-First),
  `make build`, keep the neutrality CI guard green, 16-language neutral matrix.

**Preserve**:
- The `/moai project` NO-SPEC scope guard (project flow never writes to
  `.moai/specs/**`).
- The pre-MCP harness GENERATE output when no MCP need is declared (artifact 7 is
  additive; when omitted, the output is byte-identical to the without-artifact-7
  baseline).
- The `builder-harness` `artifact_type=mcp-server` internals (reused, unchanged).
- The `harness-spec.yaml` schema + the adaptive interview (owned by
  `SPEC-PROJECT-HARNESS-BRIDGE-001`; consumed here, not modified).

**Out of scope** — see §E.

## §C. Requirements (GEARS notation)

### C.1 Phase 3.6 insertion + stack detection + matrix selection

- **REQ-HMP-001** (Event-driven): When the `/moai project` flow completes Phase 3.5
  (LSP detection) and before Phase 3.7 (auto dev-mode), it shall execute a new
  **Phase 3.6** that detects the project stack — reusing the existing
  language / framework detection PLUS the `harness-spec.yaml` `external_systems`
  and `ui_surface` fields — and selects recommended MCP servers from the MCP matrix
  (§D).
- **REQ-HMP-002** (Unwanted behavior): Phase 3.6 shall obtain approval for the
  selected MCP servers through the **orchestrator's** AskUserQuestion channel; a
  subagent-executed step (e.g. a delegated `builder-harness` scaffold) shall not
  prompt the user directly — it returns the recommendation / a blocker report for
  the orchestrator to surface.
- **REQ-HMP-003** (Event-driven): When the user approves the MCP-server selection,
  Phase 3.6 shall write the selected servers as `.mcp.json` entries at **project
  scope** (the repo-root `.mcp.json`).

### C.2 MCP recommendation matrix externalized to config

- **REQ-HMP-004** (Ubiquitous): The MCP recommendation matrix — carrying at least
  the web / mobile / backend rows — shall be externalized to a standalone **data
  resource** at `.moai/config/sections/mcp-matrix.yaml` (authored in the template tree
  at `internal/template/templates/.moai/config/sections/mcp-matrix.yaml` and mirrored
  to the local copy). The resource is read by `project/doc-generation.md` as
  prose-context — it is NOT a new Go config section and adds no typed loader or config
  struct field. The skill prose shall carry at most a fallback pointer to that
  resource, NOT a hardcoded duplicate of the matrix rows.

### C.3 Hard caps + credential-approval gate

- **REQ-HMP-005** (Ubiquitous): The Phase 3.6 recommendation shall cap the
  recommended server count at **3-5 servers maximum** and shall prefer
  vendor-maintained servers over community-maintained equivalents.
- **REQ-HMP-006** (Unwanted behavior): Where a recommended MCP server requires
  credentials or tokens, Phase 3.6 shall require an EXPLICIT per-server
  AskUserQuestion approval before writing that server, and shall not auto-write a
  credentialed server without that explicit approval.

### C.4 .mcp.json write discipline (additive, project scope, env-var secrets)

- **REQ-HMP-007** (Event-driven): When Phase 3.6 writes `.mcp.json`, the write shall
  be additive / idempotent — merging the selected servers into any existing
  `.mcp.json` rather than clobbering it — at project scope; and any secret in a
  written server entry shall be expressed in `${VAR}` env-var expansion form, never
  as an inlined literal credential / token value.

### C.5 harness generation consumes MCP (optional artifact 7)

- **REQ-HMP-008** (Ubiquitous): The `harness-builder.md` GENERATE contract shall be
  extended with an OPTIONAL **artifact 7** — a `.mcp.json` fragment — produced via
  the existing `builder-harness` `artifact_type=mcp-server` capability. Per the Epic
  canonical artifact order the contract reads **5 base + verify skill (artifact 6,
  mandatory, owned by SPEC-HARNESS-VERIFY-PROMOTE-001) + optional MCP fragment
  (artifact 7, this SPEC)**; the `harness-builder.md` "exactly 5 artifact types" prose
  shall be reconciled so the bare count no longer stands uncontextualized against the
  added artifacts.
- **REQ-HMP-009** (Event-driven): When the harness PLAN phase declares MCP needs
  (derived from `harness-spec.yaml` `external_systems` / `verification`), GENERATE
  shall emit artifact 7; when the PLAN declares no MCP need, artifact 7 shall be
  omitted and the GENERATE output shall remain byte-identical to the without-artifact-7
  baseline.

### C.6 harness doctor manifest tolerance

- **REQ-HMP-010** (State-driven): While `moai harness doctor` runs its
  reference-integrity smoke gate, an OPTIONAL `mcp` block present in a
  `manifest.json` shall not produce a doctor ERROR finding. The decision is
  **TOLERATE-ONLY** with zero Go change: the manifest decoders use plain
  `json.Unmarshal` with no `DisallowUnknownFields` (`internal/cli/harness/doctor.go`,
  `internal/harness/applier.go`) and `internal/harness/v4manifest/validate.go` checks
  only the required fields, so an unknown `mcp` block is silently tolerated. Active
  schema validation of the `mcp` block (which would add `MCP *MCPBlock` to
  `v4manifest/types.go` + a `Validate()` branch) is explicitly OUT OF SCOPE here,
  deferred to a follow-up Go SPEC.

### C.7 Invariants (NO-SPEC guard + Template-First + neutrality)

- **REQ-HMP-011** (Unwanted behavior): Phase 3.6 and the harness artifact-7 change
  shall not write to `.moai/specs/**`; `.mcp.json` lives at the repo root and
  `mcp-matrix.yaml` under `.moai/config/sections/`; every edit shall be made in
  `internal/template/templates/...` first, mirrored byte-identically to the local
  `.claude/` / `.moai/` copy, compiled via `make build`, and shall preserve
  template neutrality (no internal SPEC IDs / dates / SHAs) and 16-language
  neutrality (the matrix shall not privilege one language / stack over the others).

## §D. Schemas (SSOT)

### D.1 `.moai/config/sections/mcp-matrix.yaml` — recommendation matrix

Externalized per-project-type MCP recommendation matrix. Read by
`project/doc-generation.md` Phase 3.6 (and referenced by `harness-builder.md` PLAN).

```yaml
# .moai/config/sections/mcp-matrix.yaml — per-project-type MCP recommendation matrix
# Each row: project-type -> ordered list of recommended servers.
# Per-server keys: name, transport (stdio|http|sse), install, vendor_maintained (bool),
#                  requires_credentials (bool).  3-5 servers max per row (REQ-HMP-005).
matrix:
  web-frontend:
    - { name: playwright,      transport: stdio, install: "npx @playwright/mcp",      vendor_maintained: true,  requires_credentials: false }
    - { name: chrome-devtools, transport: stdio, install: "npx chrome-devtools-mcp",  vendor_maintained: true,  requires_credentials: false }
    - { name: figma-dev-mode,  transport: sse,   install: "figma dev-mode mcp",       vendor_maintained: true,  requires_credentials: true  }  # design->code, optional
  mobile:
    - { name: maestro,         transport: stdio, install: "maestro mcp",              vendor_maintained: true,  requires_credentials: false }
    # appium is a secondary option once Maestro coverage is insufficient
  backend-db:
    - { name: postgres,        transport: stdio, install: "postgres mcp (read-only)", vendor_maintained: true,  requires_credentials: true  }
    - { name: context7,        transport: http,  install: "context7 mcp",             vendor_maintained: true,  requires_credentials: false }
universal_starter:            # fallback when the stack is ambiguous
  - { name: github,           transport: http,  install: "github mcp",                vendor_maintained: true,  requires_credentials: true  }
  - { name: context7,         transport: http,  install: "context7 mcp",              vendor_maintained: true,  requires_credentials: false }
  - { name: playwright,       transport: stdio, install: "npx @playwright/mcp",       vendor_maintained: true,  requires_credentials: false }
```

### D.2 project-scope `.mcp.json` entry shape (repo root)

Written additively by Phase 3.6. Secrets use `${VAR}` env-var expansion.

```jsonc
// .mcp.json (repo root, project scope, checked in) — merged additively
{
  "mcpServers": {
    "playwright": { "command": "npx", "args": ["@playwright/mcp"] },
    "postgres":   { "command": "postgres-mcp", "args": ["--read-only"],
                    "env": { "DATABASE_URL": "${DATABASE_URL}" } }   // ${VAR}, never a literal token
  }
}
```

### D.3 optional manifest `mcp` block (harness artifact 7)

Emitted into `manifest.json` only when the harness PLAN declares MCP needs.
Tolerated by `moai harness doctor` (lenient decoder). The full machine-verifiable
AC matrix (AC-HMP-001 … AC-HMP-015) lives in `acceptance.md` (SSOT).

```jsonc
// manifest.json — optional 9th field (omitted when no MCP need is declared)
"mcp": {
  "servers": [ { "name": "playwright", "transport": "stdio" } ]
}
```

## §E. Exclusions

The following are explicitly out of scope for this SPEC.

### Out of Scope — harness-spec.yaml schema + adaptive interview

- The `.moai/project/harness-spec.yaml` schema, the adaptive clarity-scored
  interview, and the extended interview axes are owned by
  `SPEC-PROJECT-HARNESS-BRIDGE-001` (the FOUNDATION SPEC this one `depends_on`).
  This SPEC CONSUMES `external_systems` / `ui_surface` / `verification`; it does NOT
  add, rename, or re-elicit those fields.

### Out of Scope — builder-harness mcp-server internals

- The `builder-harness` agent's `artifact_type=mcp-server` scaffolding logic (how
  it emits `.mcp.json` entries with stdio / http / sse transports) is unchanged.
  This SPEC WIRES that existing capability into GENERATE artifact 7; it does NOT
  reimplement or alter the scaffolder.

### Out of Scope — Go code changes

- This SPEC is doc / config-only (markdown + yaml under `.claude/skills/...` and
  `.moai/config/sections/mcp-matrix.yaml` and their template mirrors). No
  `internal/` / `pkg/` / `cmd/` Go source is modified. In particular, no MCP-block
  Go parser or `v4manifest` struct field is added here — the `mcp` block is
  tolerated by the current lenient `json.Unmarshal` (the TOLERATE-ONLY decision;
  see REQ-HMP-010 + plan.md §A "Resolved clarifications").

### Out of Scope — MCP server installation / runtime health

- This SPEC provisions `.mcp.json` entries (config-time). Actually installing the
  server binaries, verifying they start, or health-checking a running MCP server is
  a runtime concern outside `/moai project`'s config-generation scope.

### Out of Scope — CHANGELOG / README / docs-site

- CHANGELOG.md is owned by manager-docs (sync-phase); README and docs-site 4-locale
  updates for the new MCP-provisioning behavior are a follow-up sync / docs concern.

## §F. Cross-References

- `.claude/skills/moai/workflows/project/doc-generation.md` — Phase 3.6 insertion
  site (between Phase 3.5 LSP and Phase 3.7 dev-mode).
- `.claude/skills/moai/workflows/harness-builder.md` — GENERATE contract (optional
  artifact 7 extension; "exactly 5" prose reconciliation).
- `.claude/agents/moai/builder-harness.md` — the existing `artifact_type=mcp-server`
  capability reused by artifact 7.
- `.moai/project/harness-spec.yaml` — the machine-readable stack signal
  (`external_systems` / `ui_surface` / `verification`) consumed by Phase 3.6 and by
  the harness PLAN MCP-need derivation. Owned by `SPEC-PROJECT-HARNESS-BRIDGE-001`.
- `.moai/config/sections/mcp-matrix.yaml` — the externalized recommendation matrix
  (§D.1) created by this SPEC.
- `internal/harness/v4manifest/types.go` / `validate.go` +
  `internal/cli/harness/doctor.go` — the manifest schema + doctor smoke gate that
  tolerate the optional `mcp` block (lenient `json.Unmarshal`).
- `plan.md` / `acceptance.md` — implementation plan + AC matrix (SSOT).
