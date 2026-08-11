---
id: SPEC-MCP-DEFAULT-ON-001
title: "moai MCP server as a first-class default — template single-entry, default-on provisioning, update-merge safety"
version: "0.1.0"
status: in-progress
created: 2026-08-12
updated: 2026-08-12
author: manager-spec
priority: P1
phase: "v3.2 target"
module: "internal/template/templates/.mcp.json, internal/cli (init.go, update_template_sync.go), internal/cli/wizard"
lifecycle: spec-anchored
tier: M
tags: "mcp, template-first, default-on, opt-out, three-way-merge, neutrality, wizard, amendment"
---

# SPEC-MCP-DEFAULT-ON-001 — moai MCP server as a first-class default (Epic SPEC-A, foundation)

## HISTORY

- 2026-08-12 (plan-phase, iter-1) — Initial Tier M authoring. Foundation SPEC of the moai-MCP integration Epic (A → B → C). Inverts the MCP provisioning gate from opt-in-default-off to default-on, reduces the distributed template `.mcp.json` to the single `moai` entry, adds `.mcp.json` to the `moai update` 3-way-merge set, and carries the two in-place amendments to `SPEC-MOAI-MCP-SERVER-001` and `SPEC-TREND-MCP-001` that the inversion requires. Directed by the project owner; the design decisions are settled and are not re-opened by this SPEC.

## §A. User Story

**As a** MoAI-ADK user who installs the harness and expects its own capabilities to be available,

**I want** the self-hosted `moai` MCP server — the 17 typed tools registered by `registerMoaiMCPTools` (`internal/cli/mcp_server.go:105`) — provisioned into my project by default, without having to discover a wizard confirm and answer it correctly,

**so that** an agent in my session can call `spec_audit`, `verify_trend`, `session_list` and the rest directly, and so that a `moai update` never silently overwrites MCP entries I added myself.

## §B. Scope Summary

This SPEC delivers the **foundation** of the Epic: the provisioning contract flip, the template shape change, the update-merge safety fix, and the documentation reconciliation across the surfaces that describe MCP defaults. It also carries the two in-place amendments to completed SPECs whose requirements currently forbid the flip.

It delivers **no agent wiring** and **no console surface** — those are SPEC-B and SPEC-C, which depend on this one.

Two `.mcp.json` files exist in this repository and they **intentionally diverge** after this SPEC. This is stated as a requirement (REQ-A-2) rather than left implicit, because they are byte-identical today and a future reader would otherwise reasonably "fix" the divergence.

### Out of Scope — agent tool wiring

- Adding `mcp__moai__*` entries to the 12 retained agents' `tools:` frontmatter, the template mirror of those agent files, and the embedded-catalog rebuild → **SPEC-MCP-AGENT-WIRING-001** (SPEC-B). This SPEC makes the server reachable; it grants no agent access to it.

### Out of Scope — web console and authentication

- The per-tool settings surface in `moai web`, codex OAuth configuration, and GLM API-key configuration → **SPEC-MCP-CONSOLE-001** (SPEC-C).

### Out of Scope — MCP tool behavior

- No tool handler, no JSON Schema, no `internal/` integration point is changed. The 17 tools behave exactly as they do at base. This SPEC changes only whether and where the server entry is written.

### Out of Scope — third-party entry deletion

- The three third-party entries (`context7`, `chrome-devtools`, `playwright`) are removed from the **distributed template's active map** only. They remain in the repo-root `.mcp.json`, remain documented in the recipe catalogue, and remain activatable via `moai mcp add` (SPEC-TREND-MCP-001 REQ-TMC-005). No entry is deleted from the world and no recipe documentation is retired.

## §C. Requirements (GEARS)

> Domain prefix `REQ-A-N` maps to this SPEC's role as the Epic's foundation. Every citation below names a file:line read at base commit `ed70e4354`.

### M1 — Template + repo-root `.mcp.json` shape

**REQ-A-1** (Ubiquitous) The distributed template `internal/template/templates/.mcp.json` shall carry exactly one active `mcpServers` entry — `moai` → `{"command":"moai","args":["mcp-server"]}` — and no other active entry. The `$schema` key and the `staggeredStartup` block are preserved unchanged. The file shall carry no `env` block on the `moai` entry, no `$comment` JSONC form, and no resolved secret of any kind.

**REQ-A-2** (Ubiquitous) The repo-root `.mcp.json` shall carry four active entries: the three existing third-party entries (`context7`, `chrome-devtools`, `playwright`) plus `moai`. The resulting divergence from the template is **intentional and permanent** — the repo root is a maintainer working tree that uses the third-party tooling directly, while the template is a neutral distribution artifact. The SPEC body, and a comment-free note in this SPEC's plan.md, shall state the divergence explicitly so a future reader does not reconcile the two files.

**REQ-A-3** (Capability gate) **Where** the user has not explicitly declined MCP provisioning, `moai init` shall write the single neutral `moai` entry into the project's `.mcp.json`. The wizard question at `internal/cli/wizard/questions.go:440-446` — currently `mcp_tools_opt_in`, `Type: QuestionTypeConfirm`, `Default: "false"` — shall be restated as a **default-`true` confirm** rather than an inverted opt-out question, and the provisioning function `provisionMCPEntryIfOptedIn` (`internal/cli/init.go:159`, called from `init.go:782`) shall be renamed so its name describes its behavior once the gate direction has flipped. A negative-polarity question ("Skip MCP provisioning?") is rejected: double-negative confirms are a known source of misread defaults, and the surrounding Page-3 questions are all positive-polarity.

**REQ-A-4** (Unwanted) `moai update` shall not overwrite a user's own `.mcp.json` entries. `collectMergeableFiles` (`internal/cli/update_template_sync.go:320-328`) shall include `.mcp.json` in its returned set so the file passes through the 3-way merge engine alongside `.claude/settings.json` and `.moai/status_line.sh`. The comment at `update_template_sync.go:322` — "MoAI no longer ships an MCP template (full MCP removal), so a user's .mcp.json is not a merge target" — is **factually false at base** (the template file exists and `internal/template/deployer.go:104` walks the embedded FS with no dotfile skip, so it is deployed) and shall be corrected rather than deleted silently.

**REQ-A-5** (Ubiquitous) Every documentation surface that states the MCP default shall be reconciled to the single-entry default: the three wizard locale strings (`internal/cli/wizard/translations.go:170`, `:312`, `:454`, each currently reading "Opt-in default-off (REQ-MCP-002)" in ko / ja / zh), `.claude/rules/moai/core/settings-management.md:33` and its template mirror, and the template `CLAUDE.md` §12 MCP inventory (`internal/template/templates/CLAUDE.md:69`, `:230-236`, which currently names Context7 and claude-in-chrome as integrated tools a distributed user will no longer receive by default). Template-side text shall carry no SPEC ID, REQ token, commit SHA, internal date, `/Users/` path, or `CLAUDE.local.md` reference.

### M2 — Amendments to completed SPECs

**REQ-A-6** (Ubiquitous) `SPEC-MOAI-MCP-SERVER-001` shall be amended in place so REQ-MCP-002's opt-in precondition reads as default-on and REQ-MCP-015's `mcp_tools_opt_in` flag reads as an opt-out, with AC-MCP-002 and AC-MCP-006 amended to match, a dated HISTORY `### Amendments` entry recording cause and scope, a `version:` bump, an `updated:` refresh, and `status: completed` left unchanged.

**REQ-A-7** (Ubiquitous) `SPEC-TREND-MCP-001` shall be amended in place so REQ-TMC-003 reads as default-on for `moai` and REQ-TMC-001's active-entry count reads as one, with AC-TMC-001 and AC-TMC-004 amended to match and the same HISTORY / version / status treatment as REQ-A-6. **REQ-TMC-002** (secret hygiene, `${VAR}` literals, §25 neutrality) and **REQ-TMC-004**'s `$comment`-free clause shall be preserved unchanged and shall still be asserted by the amended acceptance criteria.

**REQ-A-8** (Ubiquitous) Each amendment shall record its reversal rationale in prose adjacent to the amended criterion, answering the question an auditor will ask — why a completed MUST acceptance criterion is being inverted — from the artifact itself, not only from a commit message.

### Cross-cutting

**REQ-A-9** (State-driven) **While** the template `.mcp.json` shape changes, the regression guard `internal/template/mcp_template_neutrality_test.go` shall be updated in the same change: `mcpAllowedActiveKeys` (line ~37) inverts from `{context7, chrome-devtools, playwright}` to `{moai}`, and the assertions at lines 83 (count), 87 (no non-allowed key), and 92 (all allowed keys required) follow. The forbidden-token scan (`mcpForbiddenTokenRes`, lines ~44-52) and the `${VAR}`-literal secret check shall be preserved unchanged and shall still pass on the new single-entry file.

**REQ-A-10** (Ubiquitous) Every template-source edit shall be mirrored per the Template-First rule and followed by `make build`, so the embedded catalog and the committed tree stay consistent and the CI parity guard passes.

## §D. Constraints

- **C-A-1** — No MCP tool handler, schema, or `internal/` integration point is modified. Behavior of the 17 tools is byte-identical.
- **C-A-2** — Secret hygiene is absolute: no resolved credential ever enters a git-tracked `.mcp.json`; every env reference, if any is ever added, is a `${VAR}` literal expanded by the Claude Code runtime.
- **C-A-3** — Template neutrality (§25) binds every byte written under `internal/template/templates/`.
- **C-A-4** — Provisioning stays best-effort and non-fatal: a `.mcp.json` write failure warns and is swallowed, never failing `moai init` (the existing behavior at `internal/cli/init.go:164-167`).
- **C-A-5** — The user's explicit decline is honored absolutely. Default-on is a default, not a mandate.

## §E. Risks

- **R-A-1 — Perceived doctrine reversal.** Two completed SPECs' MUST criteria are inverted within a week of the second one closing. Mitigation: both amendments state cause and preserved-invariant scope in the artifact itself (REQ-A-8), and the load-bearing clauses (neutral single entry, secret hygiene, `$comment`-free, §25) are preserved verbatim rather than restated.
- **R-A-2 — Users lose third-party tools silently.** A user upgrading gets a template whose active map no longer carries `context7` / `chrome-devtools` / `playwright`. Mitigation: REQ-A-4's 3-way merge means an existing user's own entries survive the update; the recipe catalogue and `moai mcp add` remain the documented activation path; REQ-A-5 reconciles the docs that would otherwise promise tools the user no longer has.
- **R-A-3 — A default-on server that fails to start.** If `moai` is not on PATH in the host's environment, the runtime will report a failing MCP server at session start where previously it reported nothing. Accepted: the entry is only written by `moai init`, which by construction ran from a `moai` binary the user has.

## §F. Exclusions

### Out of Scope — agent tool grants

- No `.claude/agents/**` file is edited by this SPEC. Granting agents `mcp__moai__*` tools is SPEC-B.

### Out of Scope — web console

- No `internal/web/**` file is edited by this SPEC. The console surface and the auth configuration are SPEC-C.

### Out of Scope — new MCP tools

- No tool is added, removed, or renamed. The tool set stays at the 17 registered by `registerMoaiMCPTools`.

### Out of Scope — `moai mcp add|remove|list` CLI changes

- The generic entry-management CLI delivered by SPEC-TREND-MCP-001 is unchanged. This SPEC changes what ships by default, not how entries are managed at runtime.
