---
id: SPEC-V3-CLI-001
title: "CLI Subcommand Restructure — plugin / auth / setup / doctor families"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P2 Medium
phase: "v3.0.0 — Phase 6b Tier 2 Polish"
module: "internal/cli/"
dependencies:
  - SPEC-V3-PLG-001
related_gap:
  - gm#127
  - gm#128
  - gm#133
  - gm#135
  - gm#136
related_theme: "Theme — Bootstrap / CLI / Plugin (master-v3 §8.6)"
breaking: false
lifecycle: spec-anchored
tags: "cli, subcommand, plugin, auth, setup, doctor, v3"
---

# SPEC-V3-CLI-001: CLI Subcommand Restructure

## HISTORY

| Version | Date       | Author | Description                                    |
|---------|------------|--------|------------------------------------------------|
| 0.1.0   | 2026-04-22 | GOOS   | Initial v3 draft (Wave 4, Bootstrap/CLI bundle) |

---

## 1. Goal (목적)

Introduce Claude Code-aligned subcommand families — `moai plugin`, `moai auth`, `moai setup`, `moai doctor` — so that v3.0 CLI taxonomy matches CC's 52-subcommand convention (findings-wave1-bootstrap-cli.md §2.1, `main.tsx:3875-4492`). Existing moai top-level subcommands (`init`, `update`, `version`, `doctor`, `hook`, `glm`, `cg`, `cc`, `worktree`, `migrate`, `cron`) keep their current behavior; the restructure is purely additive. The `moai plugin` family is the primary driver (SPEC-V3-PLG-001 ships the installer/marketplace logic; this SPEC wires the CLI surface); `moai auth` is a forward-looking namespace for claude.ai token management; `moai setup` consolidates one-shot bootstrap tasks; `moai doctor` gains subcommands that pair with the validation layers shipped by SPEC-V3-SCH-001 / SPEC-V3-HOOKS-001 / SPEC-V3-MIG-001.

### 1.1 배경

findings-wave1-bootstrap-cli.md §2.1 lists 52 CC subcommands, of which `plugin` (11 subcommands), `auth` (3), `mcp` (8), and `plugin marketplace` (4) are `CommanderCommand` parent-child groups. moai-adk-go today registers 11 flat top-level subcommands via cobra (W1.6 §1) with no parent-child nesting except `moai worktree {new,list,remove}` and `moai migrate {agency,v2-to-v3}`.

Wave 2 roadmap `T2-CLI-01` and `T3-CLI-02..05` group together the CLI refresh items: cobra `PersistentPreRunE` wiring for migration auto-run (gm#133), shell completion subcommand (gm#135, gm#136), `--bare` minimal mode (gm#128 — deferred to v3.1 per §9 question #2), and subcommand family restructure (gm#127). This SPEC covers **only the subcommand family additions**; other T3-CLI items ship under sibling SPECs in Phase 6b.

### 1.2 Non-Goals

- `--bare` / `--print` minimal-mode flags (deferred to v3.1; out of v3.0 scope per master-v3 §10 rejection list).
- `moai mcp` subcommand family (deferred — moai's tools are not MCP-exposed; master-v3 §10 T4-MCP-01).
- `moai auth login` full OAuth flow implementation (this SPEC stubs the subcommand tree; actual claude.ai auth shipped in v3.1).
- `moai completion <shell>` wiring (sibling SPEC via T3-CLI-05; out of scope here).
- Changes to behavior of existing subcommands (`init`, `update`, `version`, `doctor`, `hook`, `glm`, `cg`, `cc`, `worktree`, `migrate`, `cron`).
- Root flag restructure (`-p`, `--print`, `--debug` — out of scope; moai has no REPL).
- Removal or renaming of any existing subcommand.

---

## 2. Scope (범위)

### 2.1 In Scope

- Introduce `moai plugin` cobra parent command with children: `install`, `uninstall`, `list`, `enable`, `disable`, `update`, `validate`, `marketplace`.
- Introduce `moai plugin marketplace` sub-parent with children: `add`, `list`, `remove`, `update`.
- Introduce `moai auth` cobra parent command with children: `login` (stub), `logout` (stub), `status`.
- Introduce `moai setup` cobra parent command with children: `token` (long-lived token scaffold, stub), `shell` (POSIX shell profile integration advisory, stub).
- Extend `moai doctor` with new subcommands: `config` (SPEC-V3-SCH-001 validator runner), `migration` (SPEC-V3-MIG-001 dry-run), `hook` (SPEC-V3-HOOKS-001 wrapper validator), `agent` (SPEC-V3-AGT-001 frontmatter validator). Default `moai doctor` behavior (today: auto-update health) is preserved.
- Alias `moai plugins` → `moai plugin` (CC parity, W1.5 §4.1).
- Help text: `createSortedHelpConfig` equivalent — each parent command's children render alphabetically; root help lists families in fixed order (core / subsystem / diagnostic).
- cobra `PersistentPreRunE` hook on the new families to wire SPEC-V3-MIG-001 runner and SPEC-V3-SCH-001 strict-mode env check. Conservative trigger list per master-v3 §9 question #10: migration runner fires on `moai init/update/doctor/migrate/plugin` only.
- Graceful degradation: when SPEC-V3-PLG-001 installer or schema validator returns "feature disabled", subcommands print human-readable guidance instead of stack traces.

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- `moai mcp` subcommand family (master-v3 §10 T4-MCP-01 reject).
- Full OAuth login flow for `moai auth login` (v3.1 scope; stub only in v3.0).
- `--bare`, `--print`, `--dangerously-skip-permissions` root flags (master-v3 §10 reject).
- `moai ps`, `moai logs`, `moai attach`, `moai kill` (CC background session primitives — moai has no REPL, reject).
- `moai completion <shell>` (sibling SPEC under T3-CLI-05).
- `moai daemon`, `moai server`, `moai ssh`, `moai open` (CC networking subcommands; master-v3 §10 reject).
- `moai task {create,list,get,update,dir}` (CC ant-only; moai uses TaskCreate/TaskUpdate APIs exposed by Claude Code, not a CLI).
- Changes to flag parsing libraries (continue using cobra; do not introduce `commander-js`-style subcommand inheritance primitives).
- Moving existing subcommands under new parents (no breaking renames).
- `moai auth` token storage implementation (stub returns "not implemented in v3.0").
- `moai doctor --json` output (deferred; `moai doctor config --json` covered by SPEC-V3-OUT-001).

---

## 3. Environment (환경)

- 런타임: Go 1.23+, moai-adk-go v3.0.0+.
- Cobra (`github.com/spf13/cobra`) — already in go.mod (W1.6 §14.1, 9-direct-dep budget).
- Claude Code 2.1.111+ (version-agnostic for CLI; CC interaction is indirect via hooks).
- Platforms: macOS / Linux / Windows (all three exercised in CI per CLAUDE.local.md §15).
- 대상 디렉터리: `internal/cli/` (new files: `plugin.go`, `plugin_marketplace.go`, `auth.go`, `setup.go`, `doctor_config.go`, `doctor_migration.go`, `doctor_hook.go`, `doctor_agent.go`).
- 영향 테스트: `internal/cli/cli_test.go`, `internal/template/commands_audit_test.go` (thin-command pattern unchanged).
- 의존 SPEC: SPEC-V3-PLG-001 (plugin installer backend); coordinates with SPEC-V3-SCH-001, SPEC-V3-MIG-001, SPEC-V3-HOOKS-001, SPEC-V3-AGT-001 for doctor subcommands.

---

## 4. Assumptions (가정)

- Cobra `Parent`/`AddCommand` nesting is sufficient; no need to introduce a different CLI framework.
- Users run `moai --help` and expect family groupings similar to `git --help` and `claude --help`.
- Existing moai subcommands remain at the root level indefinitely; family restructure is purely additive.
- SPEC-V3-PLG-001 ships `internal/plugin/` package before this SPEC's `moai plugin` CLI wiring merges.
- `moai auth` stubs do not call any external service in v3.0; they print guidance and exit with code 0.
- `moai setup token` and `moai setup shell` are documented but deferred for full implementation; v3.0 ships informational stubs.
- Shell completion scripts (bash/zsh/fish) are generated by cobra built-in machinery when needed (deferred to sibling SPEC but the new family structure must not break `cobra.GenBashCompletion` output).

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-CLI-001-001 (Ubiquitous) — plugin family**
The `moai` CLI **shall** expose a `plugin` parent subcommand with the following children registered via cobra: `install`, `uninstall`, `list`, `enable`, `disable`, `update`, `validate`, `marketplace`.

**REQ-CLI-001-002 (Ubiquitous) — plugin marketplace sub-family**
The `moai plugin marketplace` subcommand **shall** itself be a cobra parent with children: `add`, `list`, `remove`, `update`.

**REQ-CLI-001-003 (Ubiquitous) — auth family**
The `moai` CLI **shall** expose an `auth` parent subcommand with children: `login`, `logout`, `status`. In v3.0 `login`/`logout` are stubs that print guidance and exit with code 0.

**REQ-CLI-001-004 (Ubiquitous) — setup family**
The `moai` CLI **shall** expose a `setup` parent subcommand with children: `token`, `shell`. In v3.0 both are documented stubs; behavior is informational only.

**REQ-CLI-001-005 (Ubiquitous) — doctor subcommand extensions**
The `moai doctor` subcommand **shall** accept additional child subcommands `config`, `migration`, `hook`, `agent`. Each child maps to a validator defined by a sibling SPEC (SCH-001 / MIG-001 / HOOKS-001 / AGT-001).

**REQ-CLI-001-006 (Ubiquitous) — existing subcommand preservation**
Existing root subcommands (`init`, `update`, `version`, `doctor` default, `hook`, `glm`, `cg`, `cc`, `worktree`, `migrate`, `cron`) **shall** retain their v2.12 behavior byte-for-byte.

**REQ-CLI-001-007 (Ubiquitous) — plugins alias**
The `moai plugins` invocation **shall** behave as an alias of `moai plugin` for CC parity (findings-wave1-bootstrap-cli.md §2.1, `main.tsx:4148`).

**REQ-CLI-001-008 (Ubiquitous) — help ordering**
For any parent command introduced by this SPEC, the generated `--help` output **shall** list child subcommands alphabetically; the root `moai --help` **shall** group subcommands by family order: core (init/update/version) → workflow (worktree/migrate/cron) → subsystem (plugin/auth/setup) → runtime (hook/glm/cg/cc) → diagnostic (doctor).

### 5.2 Event-Driven Requirements

**REQ-CLI-001-010 (Event-Driven) — plugin dispatch to SPEC-V3-PLG-001**
**When** a user invokes `moai plugin <install|uninstall|list|enable|disable|update|validate>`, the CLI **shall** delegate execution to the corresponding `internal/plugin/` function exposed by SPEC-V3-PLG-001. The CLI layer performs only argument parsing and output formatting.

**REQ-CLI-001-011 (Event-Driven) — plugin marketplace dispatch**
**When** a user invokes `moai plugin marketplace <add|list|remove|update>`, the CLI **shall** delegate execution to `internal/plugin/marketplace.go` functions.

**REQ-CLI-001-012 (Event-Driven) — doctor config**
**When** a user invokes `moai doctor config`, the CLI **shall** call the SPEC-V3-SCH-001 validator registry and emit structured output (SPEC-V3-OUT-001 format) listing each section's validation result.

**REQ-CLI-001-013 (Event-Driven) — doctor migration**
**When** a user invokes `moai doctor migration`, the CLI **shall** call the SPEC-V3-MIG-001 runner in dry-run mode and display pending migrations with version, ID, description, and diff summary.

**REQ-CLI-001-014 (Event-Driven) — doctor hook**
**When** a user invokes `moai doctor hook`, the CLI **shall** call the SPEC-V3-HOOKS-001 wrapper validator and emit a per-wrapper status report (legacy / v2 / invalid).

**REQ-CLI-001-015 (Event-Driven) — doctor agent**
**When** a user invokes `moai doctor agent`, the CLI **shall** validate all agent frontmatter files under `.claude/agents/` against the SPEC-V3-AGT-001 schema and report any violations.

**REQ-CLI-001-016 (Event-Driven) — auth stub**
**When** a user invokes `moai auth login` or `moai auth logout` in v3.0, the CLI **shall** print a notice ("claude.ai auth flow is scheduled for v3.1; use your existing CLAUDE_CODE credentials") and exit code 0.

**REQ-CLI-001-017 (Event-Driven) — auth status**
**When** a user invokes `moai auth status`, the CLI **shall** print the current authentication source (CLAUDE_CODE env, GLM env, or "anonymous") and exit code 0.

### 5.3 State-Driven Requirements

**REQ-CLI-001-020 (State-Driven) — migration PersistentPreRun gate**
**While** a user invokes any of `moai init`, `moai update`, `moai doctor`, `moai migrate`, `moai plugin`, the cobra `PersistentPreRunE` hook **shall** run the SPEC-V3-MIG-001 runner (in dry-run mode by default, apply mode when `--yes` is passed). Other subcommands **shall NOT** trigger migration auto-run.

**REQ-CLI-001-021 (State-Driven) — plugin-disabled env**
**While** the environment variable `MOAI_PLUGINS_DISABLED=1` is set, the CLI **shall** accept `moai plugin` commands but reject execution with error `CLI_PLUGINS_DISABLED` and exit code 1.

### 5.4 Optional Features

**REQ-CLI-001-030 (Optional) — verbose flag**
**Where** the user passes `--verbose` to any of the new subcommand families, the CLI **shall** emit extended diagnostic output (per SPEC-V3-OUT-001 format).

**REQ-CLI-001-031 (Optional) — JSON output**
**Where** the user passes `--json` to `moai doctor config`, `moai doctor migration`, `moai doctor hook`, or `moai plugin list`, the CLI **shall** emit machine-readable JSON instead of the human-readable table.

### 5.5 Unwanted Behavior

**REQ-CLI-001-040 (Unwanted) — no silent rename**
If a user invokes an unknown subcommand (e.g., `moai plugin foo`), then the CLI **shall** print the parent `--help` with a "did you mean …" suggestion and exit code 1 — it **shall NOT** silently dispatch to an existing command or to the plugin binary.

**REQ-CLI-001-041 (Unwanted) — no shell execution through auth stub**
If a user invokes `moai auth login` in v3.0, then the CLI **shall NOT** exec any external binary, open a browser, or modify `~/.claude/credentials`. The stub is informational only.

**REQ-CLI-001-042 (Unwanted) — no auto-install**
If a user invokes `moai plugin <cmd>` when `internal/plugin/` is not yet wired (SPEC-V3-PLG-001 not merged), then the CLI **shall** return exit code 1 with error `CLI_PLUGIN_BACKEND_UNAVAILABLE` — it **shall NOT** fall back to executing arbitrary git clone or network operations.

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-CLI-001-01**: `moai plugin --help` lists children alphabetically: `disable, enable, install, list, marketplace, uninstall, update, validate` with descriptions. Verified by golden file in `internal/cli/plugin_test.go`.
- **AC-CLI-001-02**: `moai plugin marketplace --help` lists children alphabetically: `add, list, remove, update`. Exit code 0.
- **AC-CLI-001-03**: `moai plugins` (alias) exits with same output as `moai plugin`. Byte-identical help text. (maps REQ-CLI-001-007)
- **AC-CLI-001-04**: `moai auth --help` lists `login, logout, status`. `moai auth login` prints stub notice and exits 0. `moai auth status` prints current auth source. (maps REQ-CLI-001-003, -016, -017)
- **AC-CLI-001-05**: `moai setup --help` lists `shell, token` (alphabetical). Both stubs print informational messages and exit 0. (maps REQ-CLI-001-004)
- **AC-CLI-001-06**: `moai doctor config` invokes SCH-001 validator and prints structured output; `moai doctor migration` invokes MIG-001 dry-run; `moai doctor hook` invokes HOOKS-001 validator; `moai doctor agent` invokes AGT-001 validator. Each exits 0 on clean repo. (maps REQ-CLI-001-012 through -015)
- **AC-CLI-001-07**: Running `moai version` after the SPEC lands prints identical output to v2.12.0 (except for the version string itself). `moai init`, `moai update`, `moai doctor` (no subcommand), `moai hook <event>`, `moai glm`, `moai cg`, `moai cc`, `moai worktree`, `moai migrate`, `moai cron` all retain v2.12 behavior. (maps REQ-CLI-001-006)
- **AC-CLI-001-08**: `moai foo` (unknown) prints `moai --help` with "did you mean 'migrate'?" suggestion and exits code 1. (maps REQ-CLI-001-040)
- **AC-CLI-001-09**: With `MOAI_PLUGINS_DISABLED=1`, `moai plugin list` exits 1 with `CLI_PLUGINS_DISABLED` error. (maps REQ-CLI-001-021)
- **AC-CLI-001-10**: `moai plugin list --json` emits JSON array to stdout; no ANSI colors; parseable by `jq '.'`. (maps REQ-CLI-001-031)
- **AC-CLI-001-11**: Running `moai init`, `moai update`, or `moai doctor` triggers the SPEC-V3-MIG-001 runner in dry-run; running `moai version` does NOT. (maps REQ-CLI-001-020)
- **AC-CLI-001-12**: `go test ./internal/cli/...` passes with ≥ 85% coverage for new files; `internal/template/commands_audit_test.go` still passes (thin-command rule preserved).

---

## 7. Constraints (제약)

- **[HARD] Additive only**: No existing subcommand may be removed, renamed, or moved under a new parent.
- **[HARD] Cobra-only**: Do not introduce another CLI library; continue using `github.com/spf13/cobra`.
- **[HARD] 9-direct-dep budget** (master-v3 §1.2, CLAUDE.local.md §14): No new top-level Go module dependencies introduced by this SPEC.
- **[HARD] Thin-command pattern preserved**: `.claude/commands/**/*.md` files are template assets for Claude Code slash commands; this SPEC does NOT touch them. The CLI restructure is in `internal/cli/` only.
- **[HARD] Migration auto-run gated**: `PersistentPreRunE` migration firing is limited to init/update/doctor/migrate/plugin per master-v3 §9 question #10 (conservative default).
- **[HARD] Help text language**: Cobra command descriptions remain English per `.claude/rules/moai/development/coding-standards.md` (Language Policy). User-facing messages follow `language.yaml` conversation_language for error text only.
- **[HARD] Windows parity**: All new subcommands must function identically on Windows; path handling via `filepath` stdlib, not string concat.
- Binary size delta ≤ 400 KB for CLI restructure alone (excluding SPEC-V3-PLG-001 payload).
- Alphabetical child sort enforced by `createSortedHelpConfig`-equivalent helper; tested via golden file.

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 확률 | 영향 | 완화 |
|--------|------|------|------|
| Users interpret new `moai auth` as functional and log in via claude.ai expecting it to work | Medium | Medium | v3.0 stub prints explicit "scheduled for v3.1" notice; docs-site migration guide (`docs-site/.../migration/v3.md`) lists stubs section |
| `PersistentPreRunE` migration trigger fires unexpectedly on `moai version` or `moai hook` | Low | Medium | REQ-CLI-001-020 enumerates the exact trigger set; regression test `TestPersistentPreRunOnlyOnAllowedCommands` verifies negative cases |
| Help text ordering changes break user scripts parsing `moai --help` | Low | Low | Subcommand names unchanged; machine-parseable output remains via `--json` (REQ-CLI-001-031) |
| Plugin backend (SPEC-V3-PLG-001) ships late; `moai plugin` CLI exists without backend | Medium | Medium | REQ-CLI-001-042 gates CLI with `CLI_PLUGIN_BACKEND_UNAVAILABLE`; Phase 6a ordering ensures PLG-001 merges before this SPEC's plugin subcommand wiring lands |
| Windows cobra completion output drifts after subcommand additions | Low | Low | CI matrix includes Windows; `TestCobraBashCompletion`, `TestCobraZshCompletion` golden files updated alongside family additions |
| Alias `moai plugins` confuses shell completions (two names for same command) | Low | Low | Cobra built-in `Aliases` field; golden file tests verify both paths yield identical help |
| `moai setup token` stub becomes load-bearing when users expect real token persistence | Medium | Low | Stub exits 0 with clear message; no credentials written; docs note v3.1 schedule |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3-PLG-001** — provides `internal/plugin/` package that `moai plugin` subcommands delegate to. This SPEC's plugin subcommand wiring cannot merge before PLG-001.

### 9.2 Blocks

- **SPEC-V3-MIGRATE-001** — `moai migrate v2-to-v3` tool CLI wiring reuses the doctor/migration patterns established here.
- **SPEC-V3-OUT-001** — `--json` and verbose output rely on the renderer contract this SPEC references.

### 9.3 Related

- **SPEC-V3-SCH-001** — `moai doctor config` consumes the validator registry.
- **SPEC-V3-MIG-001** — `moai doctor migration` consumes the migration runner; `PersistentPreRunE` invokes it.
- **SPEC-V3-HOOKS-001** — `moai doctor hook` consumes the hook wrapper validator.
- **SPEC-V3-AGT-001** — `moai doctor agent` consumes the agent frontmatter validator.

---

## 10. Traceability (추적성)

- Theme: master-v3 §8.6 (Bootstrap / CLI / Plugin / Migration / Schema grouping); T2-CLI-01 in priority roadmap.
- Gap rows: gm#127 (subcommand family restructure), gm#128 (`--bare` — out of scope per §1.2), gm#133 (PersistentPreRunE migration wiring), gm#135 / gm#136 (completion — out of scope per §1.2).
- Wave 1 sources:
  - findings-wave1-bootstrap-cli.md §2.1 (52 CC subcommands, `main.tsx:3875-4492`)
  - findings-wave1-bootstrap-cli.md §2.2 (root flags — out of scope)
  - findings-wave1-bootstrap-cli.md §1.7 (profileCheckpoint markers — deferred to T3-CLI items)
  - findings-wave1-moai-current.md §1 (current moai CLI subcommand inventory)
  - findings-wave1-moai-current.md §8 (command catalog)
- BC-ID: none (additive).
- REQ 총 개수: 17 (Ubiquitous 8, Event-Driven 8, State-Driven 2, Optional 2, Unwanted 3, minus shared IDs — discrete 17).
- 예상 AC 개수: 12.
- 코드 구현 예상 경로:
  - `internal/cli/plugin.go` (REQ-CLI-001-001, -007, -010, -042)
  - `internal/cli/plugin_marketplace.go` (REQ-CLI-001-002, -011)
  - `internal/cli/auth.go` (REQ-CLI-001-003, -016, -017, -041)
  - `internal/cli/setup.go` (REQ-CLI-001-004)
  - `internal/cli/doctor_config.go`, `doctor_migration.go`, `doctor_hook.go`, `doctor_agent.go` (REQ-CLI-001-005, -012, -013, -014, -015)
  - `internal/cli/root.go` (REQ-CLI-001-006, -008, -020, -040 — persistentPreRunE, help ordering)
  - `internal/cli/cli_test.go` golden files for `--help` outputs.

---

End of SPEC.
