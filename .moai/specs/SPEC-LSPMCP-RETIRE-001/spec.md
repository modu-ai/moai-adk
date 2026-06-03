---
id: SPEC-LSPMCP-RETIRE-001
title: "Retire dormant moai-lsp MCP bridge"
version: "0.1.0"
status: completed
created: 2026-06-03
updated: 2026-06-03
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/mcp"
lifecycle: spec-anchored
tags: "lsp, mcp, dead-code-removal, retirement, cleanup"
---

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-06-03 | 0.1.0 | Initial plan-phase draft — retire dormant moai-lsp MCP bridge (supersedes SPEC-LSPMCP-001) |

---

# SPEC-LSPMCP-RETIRE-001: Retire dormant moai-lsp MCP bridge

## Overview

Remove the never-completed `moai-lsp` MCP bridge — the `internal/mcp` package, the
`moai mcp lsp` CLI subcommand, and the `moai-lsp` server entry in the deployable
`.mcp.json.tmpl` template — along with the two template tests and the documentation
references that assert its presence. The predecessor that introduced this bridge
(`SPEC-LSPMCP-001`) is transitioned to `superseded` by this SPEC.

This is a focused dead-code removal. The bridge was a Phase-1 stub that never reached
Phase 2 (real language-server connection); its original motivation has been resolved by
a different, fully-implemented system (`internal/lsp` + `gopls.NewBridge`); and its
6-tool design has been superseded by Claude Code's native `LSP` tool.

## Motivation

Four verified facts establish that the moai-lsp MCP bridge is orphaned dead code:

1. **Stub never completed.** `internal/mcp/lsp_stub.go` defines
   `notConnectedMsg = "LSP bridge not yet connected to a language server. Install gopls
   (Go), typescript-language-server (TS), or pyright (Python) and restart."` and ALL 6
   LSP tools (`gotoDefinition`, `findReferences`, `hover`, `documentSymbols`,
   `diagnostics`, `rename`) return `IsError: true` stub responses. `internal/mcp/handler.go`
   carries the comment "In Phase 1 all LSP operations return a 'not connected' stub
   response." Phase 2 (real language-server connection per predecessor REQ-LSPMCP-003) was
   never implemented.

2. **Predecessor SPEC is legacy.** `.moai/specs/SPEC-LSPMCP-001/spec.md` carries
   `status: archived`, `phase: "v2.x - Legacy"`, `tags: "legacy"`.

3. **Superseded by Claude Code native LSP.** Claude Code now ships a native `LSP` tool
   with 9 operations (goToDefinition, findReferences, hover, documentSymbol,
   workspaceSymbol, goToImplementation, prepareCallHierarchy, incomingCalls,
   outgoingCalls) plus passive post-edit diagnostics feedback, enabled via the
   `ENABLE_LSP_TOOL` user setting. It supersedes the 6-tool moai-lsp design.

4. **Original motivation already resolved elsewhere.** The predecessor's stated goal was
   to fill the nil `DiagnosticsProvider` slot in `internal/cli/deps.go`. That slot is now
   filled by `gopls.NewBridge` (deps.go, the GOPLS-BRIDGE-001 wiring block) via the
   `internal/lsp` suite (status: implemented). The moai-lsp MCP bridge is therefore a
   never-activated dead branch with no production consumer.

### Why retire rather than complete

Completing Phase 2 of the bridge would duplicate functionality that already exists in two
places (Claude Code native LSP + the powernap-based `internal/lsp` client). Carrying a
permanently-stubbed package incurs maintenance cost (CI test surface, template noise,
documentation that describes a non-functional feature) with zero offsetting value.

## Requirements (GEARS Format)

### REQ-LSPMCP-RETIRE-001 (Ubiquitous)
The codebase shall not contain the `internal/mcp` package after retirement; the package
directory and all its files (`handler.go`, `tools.go`, `lsp_stub.go`, `protocol.go`,
`server.go`, and their `*_test.go` siblings) shall be removed.

### REQ-LSPMCP-RETIRE-002 (Ubiquitous)
The `moai` CLI shall not expose the `moai mcp` parent command nor the `moai mcp lsp`
subcommand after retirement; the file `internal/cli/mcp.go` shall be removed and its
cobra registration shall not appear in `moai --help`.

### REQ-LSPMCP-RETIRE-003 (Ubiquitous)
The deployable template `internal/template/templates/.mcp.json.tmpl` shall not contain a
`moai-lsp` server entry after retirement; the `context7` server entry and the
`staggeredStartup` block shall remain unchanged.

### REQ-LSPMCP-RETIRE-004 (Event-driven)
When the orchestrator runs `make build` (or `go build ./...`) after the template edit,
the embedded template filesystem shall reflect the removed `moai-lsp` entry, keeping the
embedded copy consistent with the source template.

### REQ-LSPMCP-RETIRE-005 (Ubiquitous)
The two template tests that assert the presence of a `moai-lsp` server entry —
`TestMCPTemplateAlwaysLoadAbsentOnMoaiLSP` and the `moai-lsp` assertion block within
`TestMCPTemplateExistingFieldsPreserved` (both in `internal/template/settings_test.go`) —
shall be removed or adjusted so the test suite passes with the `moai-lsp` entry absent,
while the `context7` and `staggeredStartup` assertions in the same file remain intact.

### REQ-LSPMCP-RETIRE-006 (Ubiquitous)
The documentation that describes `moai-lsp` `alwaysLoad` behavior — both lines in
`.claude/rules/moai/core/settings-management.md` and the byte-identical template mirror
`internal/template/templates/.claude/rules/moai/core/settings-management.md` — shall be
updated to remove the `moai-lsp` references, and both copies shall remain byte-identical
per the mirror-parity invariant.

### REQ-LSPMCP-RETIRE-007 (State-driven)
While the retirement is applied, the predecessor SPEC `SPEC-LSPMCP-001` frontmatter
`status` shall transition from `archived` to `superseded`, with a `superseded_by:
SPEC-LSPMCP-RETIRE-001` field recorded, documenting the supersession link.

### REQ-LSPMCP-RETIRE-008 (Event-detected — replaces legacy unwanted/IF-THEN form)
When `go build ./...` is run for both the host platform and `GOOS=windows GOARCH=amd64`
after removal, the build shall succeed (exit 0) with no dangling import of
`github.com/modu-ai/moai-adk/internal/mcp` and no unresolved reference to the `moai mcp`
command surface.

### REQ-LSPMCP-RETIRE-009 (State-driven)
While the moai-lsp block is removed from `.mcp.json.tmpl`, the rendered template shall
remain valid JSON across all platform branches of the Go template (`{{- if eq .Platform
"windows"}}` / `{{- else}}`), with no trailing-comma or dangling-brace defects.

### REQ-LSPMCP-RETIRE-010 (Ubiquitous — PRESERVE invariant)
The `internal/lsp` package shall remain byte-unchanged by this retirement. `internal/lsp`
is a distinct system (the powernap-based multi-language LSP client) with 10 verified
consumers; it is NOT part of the moai-lsp MCP bridge and MUST NOT be touched.

## Affected Files

### Removed (entire files)

- `internal/mcp/handler.go`
- `internal/mcp/handler_test.go`
- `internal/mcp/tools.go`
- `internal/mcp/tools_test.go`
- `internal/mcp/lsp_stub.go`
- `internal/mcp/protocol.go`
- `internal/mcp/protocol_test.go`
- `internal/mcp/server.go`
- `internal/mcp/server_test.go`
- `internal/cli/mcp.go`

Verified scope boundary: `internal/cli/mcp.go` is the ONLY non-test importer of
`internal/mcp` (`grep "moai-adk/internal/mcp" --include="*.go"` returns a single match).
The package is an isolated dead branch.

### Modified

- `internal/template/templates/.mcp.json.tmpl` — remove the `moai-lsp` server entry
  (the `"moai-lsp": { ... }` block, ~6 lines); keep `context7` + `staggeredStartup`.
- `internal/template/settings_test.go` — remove `TestMCPTemplateAlwaysLoadAbsentOnMoaiLSP`
  and the `moai-lsp` assertion block inside `TestMCPTemplateExistingFieldsPreserved`.
- `.claude/rules/moai/core/settings-management.md` — remove the two `moai-lsp` references.
- `internal/template/templates/.claude/rules/moai/core/settings-management.md` — mirror the
  same removal (byte-parity invariant).
- `internal/template/embedded.go` — regenerated by `make build` (auto-generated; not
  hand-edited).
- `.moai/specs/SPEC-LSPMCP-001/spec.md` — frontmatter `status: archived → superseded` +
  `superseded_by: SPEC-LSPMCP-RETIRE-001`.

### Review-only (verified NOT to require change — documented to forestall scope creep)

- `internal/lsp/subprocess/launcher_main_test.go` — the string `moai-lsp-subprocess-test-*`
  is a `MkdirTemp` prefix, coincidental, unrelated to the MCP bridge. No change.
- `internal/cli/glm_tools_test.go` — constructs its OWN synthetic `moai-lsp` fixture entry
  to assert GLM tooling preserves unrelated MCP entries; it does NOT read the real
  `.mcp.json.tmpl`. The fixture is self-contained; the test is independent of template
  content. No change required (run-phase MAY rename the fixture key for clarity, but this
  is optional and out of the must-pass scope).
- `internal/cli/update_clean_install_test.go` — `moai-lsp` appears only in a comment
  illustrating PATH-resolved MCP servers. Comment-only; optional touch-up.

## HARD PRESERVE — DO NOT remove or touch

### `internal/lsp/` package (ENTIRELY)

`internal/lsp` is a DIFFERENT system from `internal/mcp`: the powernap-based
multi-language LSP client. Verified consumers (10 importers, all `status: implemented`
upstream):

- `internal/ralph/engine.go` (/moai loop diagnostics)
- `internal/mx/resolver.go` + `internal/mx/fanin_lsp.go` (@MX fan-in analysis)
- `internal/hook/post_tool.go` + `internal/hook/teammate_idle.go`
- `internal/hook/quality/lint_instruction.go`
- `internal/loop/state.go` + `internal/loop/go_feedback.go`
- `internal/cli/deps.go` (the real DiagnosticsProvider wiring)
- `internal/cli/lsp_doctor.go` (`moai lsp doctor` / `lsp_doctor` command)

Removing `internal/lsp` would break /moai loop, /moai fix, and @MX. This SPEC's acceptance
criteria include a preserve guard (AC-LSPMCP-RETIRE-010) asserting `internal/lsp` is
byte-unchanged.

### `internal/cli/deps.go` `gopls.NewBridge`

The GOPLS-BRIDGE-001 wiring block in deps.go is the REAL `DiagnosticsProvider`; it is part
of the `internal/lsp` suite, NOT moai-lsp. DO NOT touch.

### The cobra `tools` GroupID

`internal/cli/mcp.go` sets `GroupID: "tools"`. This group is shared by ~10 other commands
(`agent_lint`, `clean`, `constitution`, `harness`, `github`, `hook`, `lsp_doctor`, `loop`,
`profile`, `mx`). Removing `mcp.go` does NOT remove the group — there is no group-registration
dependency. No collateral CLI breakage.

## Exclusions (What NOT to Build)

This section enumerates what is deliberately out of scope for the retirement. The
binding "out of scope" boundary is the H3 subsection below; the H4 subsection lists
non-goals that complement it.

### Out of Scope — Retirement boundary (preserve / forbidden)

- **Do NOT remove or modify `internal/lsp`** (any subpackage). It is the active LSP client
  with 10 consumers, entirely distinct from the dead `internal/mcp` bridge.
- **Do NOT touch `internal/cli/deps.go` `gopls.NewBridge` / GOPLS-BRIDGE-001 wiring.**
- **Do NOT remove or modify the `moai lsp doctor` command** (`internal/cli/lsp_doctor.go`)
  — it serves the `internal/lsp` suite, not moai-lsp.
- **Do NOT implement Phase 2** of the bridge (real language-server connection). The
  decision is retirement, not completion.
- **Do NOT add a replacement MCP server** or any new MCP tool. Native Claude Code LSP +
  `internal/lsp` already cover the capability.
- **Do NOT alter the `context7` server entry or the `staggeredStartup` block** in
  `.mcp.json.tmpl`. Only the `moai-lsp` entry is removed.
- **Do NOT modify the local `.mcp.json`** at the repo root. It already lacks a `moai-lsp`
  entry (pre-existing drift) and is a runtime-managed, machine-specific file (per
  CLAUDE.local.md §2). The template edit makes template and local consistent.
- **Do NOT introduce moai-adk internal SPEC IDs, REQ tokens, dates, or commit SHAs into
  the template files** (`internal/template/templates/**`). Template content stays generic
  per CLAUDE.local.md §15 (16-language neutrality) and §25 (internal-content isolation).
  The settings-management.md mirror edit must remove `moai-lsp` lines without introducing
  internal-tracking tokens.
- **Do NOT change the predecessor SPEC body content** beyond the frontmatter
  `status`/`superseded_by` fields (status transition only).

#### Out of Scope — Deferred / not-this-SPEC

- Migration of users from `moai mcp lsp` to native Claude Code LSP (the bridge never
  connected, so there is no functional migration — only the dead surface is removed).
- Any change to the native Claude Code LSP tool configuration (`ENABLE_LSP_TOOL`).
- Refactoring or cleanup of `internal/lsp` adjacent code "while we are here."

## Dependencies

- Supersedes `SPEC-LSPMCP-001` (status: archived → superseded).
- Adjacent (non-blocking, must remain intact): `SPEC-CC2122-MCP-001` (module "mcp",
  status: implemented — "Claude Code v2.1.119 MCP alwaysLoad 통합"). The `alwaysLoad: true`
  logic on the `context7` entry is owned by CC2122-MCP-001 and MUST survive the moai-lsp
  removal; AC-LSPMCP-RETIRE-003 and AC-LSPMCP-RETIRE-009 guard this. Note: REQ-005 deletes
  `TestMCPTemplateAlwaysLoadAbsentOnMoaiLSP`, which is the named acceptance test for
  CC2122-MCP-001 REQ-002 (the moai-lsp `alwaysLoad`-absent guard); this retirement
  intentionally retires that CC2122-MCP-001 acceptance test, which becomes vacuous once
  the `moai-lsp` entry it guarded no longer exists — no orphaned-AC regression results
  because the guarded entry is gone.
- Adjacent (must remain intact): `SPEC-LSP-CORE-002` + `SPEC-LSP-FLAKY-002` (the
  `internal/lsp` suite, status: implemented) — the PRESERVE boundary.

## Non-Goals

The deferred / not-this-SPEC items are enumerated under the `#### Out of Scope — Deferred`
subsection above (consolidated there so the spec-lint `MissingExclusions` H3 boundary and
the non-goals share a single source).
