# plan.md — SPEC-TREND-MCP-001

> Implementation plan, milestones, technical approach. Tier M (3-artifact set: spec.md + plan.md + acceptance.md). Ordered by decision-reversibility — highest-change-likelihood decisions first.

## §A. Context

### §A.1 Grounding evidence (verified at worktree HEAD `1010e9c43`, 2026-08-07)

- **Repo-root `.mcp.json`** (`.mcp.json`): 2 entries (`context7`, `chrome-devtools`), both `/bin/bash -l -c exec npx -y <pkg>@latest` form. NOT template-managed today (the template tree `internal/template/templates/` carries NO `.mcp.json` — verified via `find -name "*mcp*"` returning empty for the template root).
- **`moai mcp-server` wrapper**: live at `internal/cli/mcp_server.go`, registered at `internal/cli/root.go:214` via `newMCPServerCmd()`. Entry builder `buildMoaiMCPServerEntry` (referenced from `internal/cli/mcp_audit.go:13`, defined in `mcp_server.go`).
- **Atomic-RMW seam**: `mutateClaudeJSONAtomic` at `internal/cli/glm_tools.go:541` — flock + compare-retry + backup-before-publish + idempotent-skip. fan_in = 2 (`runEnableMCPServerForTool`, `enableMCPServerIdempotentForTool`).
- **Existing generic helpers in `glm_tools.go`** (reusable unchanged): `readClaudeJSON`, `readClaudeJSONRaw`, `parseClaudeJSON`, `writeClaudeJSONBytes`, `backupClaudeJSON`, `withClaudeJSONLock`, `getMCPServers`, `mcpEntryEqual`, `normalizeForCompare`. The `extractTokenFromEntry` helper is Z.AI-specific; the generic CLI does NOT reuse it.
- **`mcp-matrix.yaml`**: exists ONLY at `.moai/config/sections/mcp-matrix.yaml` (local-only, acknowledged config orphan per `settings-management.md`); NOT in `internal/template/templates/.moai/config/sections/`. SPEC-HARNESS-MCP-PROVISION-001 claims to have shipped it to template, but the template tree does NOT carry it — this is a discrepancy this SPEC does NOT own (out of scope; flagged for a future HARNESS-MCP-PROVISION audit). The local `mcp-matrix.yaml` is treated as a dogfood reference only.
- **§25 CI guard**: `.github/workflows/template-neutrality-check.yaml` runs `TestTemplateNeutralityAudit` (C1 macOS-bias, C2 V3R prefix, C4 feedback_/memory.md, C5 CLAUDE.local.md, C6 PR #N, C8 GOOS env). `internal/template/internal_content_leak_test.go` covers SPEC-ID (C1 in its taxonomy), commit-SHA (C7), date (C3). The new template `.mcp.json` must pass BOTH.
- **Settings doctrine**: `.claude/rules/moai/core/settings-management.md` § MCP Configuration currently states "MoAI-ADK provisions exactly one local MCP server ... via an opt-in `.mcp.json` entry ... no third-party entries" (per MOAI-MCP-SERVER-001 REQ-MCP-016 reversal). This SPEC re-scopes the "no third-party entries" clause to "no third-party entries that carry secrets / require credentials / fail §25 neutrality" and describes the multi-entry provisioning contract.

### §A.2 Decision-reversibility ordering (this plan leads with the highest-change-likelihood decisions)

1. **Tier / scope decision** (NEEDS CLARIFICATION — see §B.B1): whether to ship the full Tier M scope (template `.mcp.json` + generic CLI + doctrine update + catalogue) OR collapse to a Tier S recipe-only scope (catalogue doc + skip rationale, no template `.mcp.json`, no CLI). This decision changes the file count, the doctrine reversal, and the §25 CI guard interaction. The SPEC is authored Tier M; the orchestrator's user-question channel resolves this BEFORE Implementation Kickoff Approval.
2. **Doctrine reversal scope**: how to reconcile with MOAI-MCP-SERVER-001 REQ-MCP-002 — narrow re-scoping (this SPEC's choice) vs. in-place amendment of the completed SPEC (heavier, requires `amendment_of:` flow).
3. **ast-grep default-disabled rendering**: commented-out JSON (unusual but catalogue-friendly) vs. JSON5/`$comment`-anchored disable instructions vs. a separate `mcp.disabledServers` field (cleaner but a schema change).
4. **Generic CLI surface shape**: `moai mcp add|remove|list` (this SPEC's choice) vs. extending `moai glm tools enable|disable` (rejected — GLM CLI is AUTH-bearing for Z.AI, mixing concerns).
5. **Catalogue surface**: `.moai/docs/mcp-recipes.md` (local-only, like other `.moai/docs/`) vs. `docs-site/content/en/mcp-recipes.md` (distributed public doc) vs. both. docs-site is sync-phase (manager-docs).
6. **Refactor depth of `glm_tools.go`**: reuse-unchanged (this SPEC's choice, REQ-TMC-008) vs. extract a narrower `mcp_entry_ops.go` helper. The reuse-unchanged path is lower-risk; the refactor is additive if extracted.

### §A.3 Affected file inventory (Tier M scope — full pathspec)

**NEW files**:
- `internal/template/templates/.mcp.json` (the template-managed MCP provisioning surface — NEW)
- `internal/cli/mcp.go` (generic `moai mcp add|remove|list` subcommand registration; reuses `mutateClaudeJSONAtomic`)
- `.moai/docs/mcp-recipes.md` (recipe catalogue + skip rationale — local-only doc, matches the `.moai/docs/` convention)
- `.moai/specs/SPEC-TREND-MCP-001/{spec,plan,acceptance,progress}.md` (this SPEC's artifacts)

**MODIFIED files**:
- `.mcp.json` (repo-root — aligned with the new template source after `make build`; PRE-EXISTING local-only file)
- `.claude/rules/moai/core/settings-management.md` (§ MCP Configuration doctrine update)
- `internal/template/templates/.claude/rules/moai/core/settings-management.md` (mirror — Template-First)
- `internal/cli/root.go` (register `newMCPCmd()` — single-line `AddCommand` insertion)
- `internal/config/envkeys.go` (NEW env-var constants only if introduced — TBD M2)
- `internal/config/defaults.go` (NEW thresholds only if introduced — TBD M2)

**TEST files** (Tier M TDD):
- `internal/cli/mcp_test.go` (NEW — generic CLI: idempotent add, partial-remove safety, secret-rejection, atomic-RMW under concurrent writer, `--scope user|project`, `--json` output)
- `internal/template/mcp_template_neutrality_test.go` (NEW or extended — asserts the template `.mcp.json` passes `TestTemplateNeutralityAudit` + `internal_content_leak_test.go`)

### §A.4 PRESERVE list (scope discipline — out-of-scope, must NOT touch)

- `internal/cli/glm_tools.go` body (reused UNCHANGED per REQ-TMC-008; only ADDED calls from `mcp.go`, no edits to existing functions)
- `internal/cli/mcp_server.go` (the `moai mcp-server` wrapper — owned by SPEC-MOAI-MCP-SERVER-001)
- `internal/cli/mcp_audit.go`, `mcp_codex.go`, `mcp_glm.go` (audit backends — owned by SPEC-MOAI-MCP-SERVER-001)
- `internal/cli/mcp_doctor_coverage_test.go`, `mcp_audit_test.go`, `mcp_codex_test.go`, `mcp_glm_test.go`, `mcp_server_test.go` (existing MCP tests — not touched)
- `.moai/config/sections/mcp-matrix.yaml` (local-only dogfood reference; SPEC-HARNESS-MCP-PROVISION-001 scope)
- `internal/spec/`, `internal/runtime/`, `internal/hook/` (no SPEC lifecycle / runtime / hook changes)

## §B. Known Issues + NEEDS CLARIFICATION

### §B.B1 — [NEEDS CLARIFICATION: Tier / scope — full Tier M vs collapsed Tier S]

The autonomy report §3.7 language "번들 추가" (bundle add) for Playwright and "번들(opt-in 토글)" for ast-grep suggests the template ships these entries. MOAI-MCP-SERVER-001 REQ-MCP-002 (completed AFTER the report) prohibits "third-party entries" in the provisioned `.mcp.json`. The two are in tension.

**Option A (full Tier M — this SPEC's authored scope)**: create the template `.mcp.json` with 5 entries (context7, chrome-devtools, playwright default-on, ast-grep default-disabled, moai opt-in), reverse MOAI-MCP-SERVER-001 REQ-MCP-002 narrowly, ship the generic CLI, update doctrine + catalogue. ~10-12 files.

**Option B (collapsed Tier S — recipe-only)**: NO template `.mcp.json` creation. Deliver ONLY the recipe catalogue (`.moai/docs/mcp-recipes.md`) + skip rationale + the generic `moai mcp add|remove|list` CLI (which a user uses to populate their own `.mcp.json`). MOAI-MCP-SERVER-001 REQ-MCP-002 stays literally intact. ~4-5 files.

The orchestrator's user-question channel resolves this BEFORE Implementation Kickoff Approval. If Option B, this SPEC is re-tiered to S and the doctrine-update REQ-TMC-013 + the template-`.mcp.json` REQ-TMC-001/002/003/004 are struck (the generic CLI REQ-TMC-005..010 and the catalogue REQ-TMC-011/012/014 stay).

### §B.B2 — Concurrent writer behavior

`mutateClaudeJSONAtomic` already guards against a live Claude Code process that does not respect the `.lock` file (compare-retry, max 3 retries, fail-open after ceiling — `internal/cli/glm_tools.go:541-625`). The generic CLI inherits this behavior unchanged. No new guard required.

### §B.B3 — §25 neutrality false-positive hazard

The new template `.mcp.json` is the FIRST template file whose primary content is JSON entries referencing external package names (`@playwright/mcp`, `@ast-grep/mcp`, `chrome-devtools-mcp`, `@upstash/context7-mcp`). The §25 guard patterns target SPEC IDs, macOS-bias paths, commit SHAs, dates — none of which appear in MCP package names. Risk is low; the new `mcp_template_neutrality_test.go` is a regression guard.

### §B.B4 — ast-grep default-disabled rendering

Standard `.mcp.json` (per the Claude Code schema) does not define a `disabled` field. The default-disabled rendering must use either:
- A commented-out JSON form (non-standard — JSON has no comments; requires JSONC interpretation), OR
- A `$comment`-anchored disable marker that the Claude Code runtime ignores (verified — the existing `.mcp.json` entries already carry `$comment` fields, e.g. the `context7` entry has `$comment: "Up-to-date documentation..."`), OR
- The entry omitted from the active `mcpServers` map and documented in the recipe catalogue only.

**Plan-phase choice**: the third form (omitted from active map, documented in catalogue) is the cleanest — it avoids JSONC interpretation entirely and the catalogue's `moai mcp add ast-grep ...` one-liner is the enable path. REQ-TMC-004 is re-interpreted accordingly: the template ships a documented placeholder (in the catalogue) and the user runs one CLI command to activate it. The NEEDS CLARIFICATION marker on this rendering choice is removed at M1 entry.

### §B.B5 — Subagent boundary static guard

`internal/cli/web_test.go` `TestWeb_NoAskUserQuestion` is the canonical static guard. The new `internal/cli/mcp.go` MUST be covered by an equivalent `TestMCP_NoAskUserQuestion` (grep `AskUserQuestion|mcp__askuser` in `mcp.go`, expect zero matches excluding comments). This is a hard C-HRA-008 contract.

### §B.B6 — Generic CLI flag design

The `moai mcp add` flag set mirrors `moai glm tools enable --scope user|project` for path resolution. The `--env KEY=VAL` flag rejects positional secrets (REQ-TMC-009). The `--type stdio|http` flag selects the entry shape (stdio = `{command, args, env}`, http = `{type:"http", url, headers}`). The `--headers` flag is JSON-encoded for HTTP entries.

### §B.B7 — Recipe catalogue location

`.moai/docs/mcp-recipes.md` is local-only (matches the `.moai/docs/` convention — `dev-only-commands-isolation.md`, `template-internal-isolation-doctrine.md`, etc. are all local-only). A docs-site page for end-user-facing recipe browsing is sync-phase scope (manager-docs); the plan identifies the surface but does not author the prose.

### §B.B8 — Pre-existing test failures in `internal/template`

The `internal/template` package has pre-existing test failures unrelated to this SPEC (per the `template-neutrality-check.yaml` comment). The CI guard runs `TestTemplateNeutralityAudit` in isolation; the new `mcp_template_neutrality_test.go` is likewise isolated. Run `go test -run TestMCPNeutrality ./internal/template/...` as the gate, NOT `go test ./internal/template/...`.

## §C. Pre-flight (read-only verification BEFORE M1)

```bash
# 1. Worktree + branch state
git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/trend-mcp rev-parse --abbrev-ref HEAD  # → feat/spec-trend-mcp-001
git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/trend-mcp rev-parse --short HEAD         # → 1010e9c43 (or later)

# 2. Confirm no template .mcp.json exists (this SPEC creates the first one)
ls /Users/goos/MoAI/moai-adk-go/.claude/worktrees/trend-mcp/internal/template/templates/.mcp.json 2>&1 | head -1  # → "No such file"

# 3. Confirm the atomic-RMW seam is unchanged (grep the signature)
grep -n "func mutateClaudeJSONAtomic" /Users/goos/MoAI/moai-adk-go/.claude/worktrees/trend-mcp/internal/cli/glm_tools.go  # → ":541"

# 4. Confirm the moai mcp-server wrapper is live (dependency check)
grep -n "newMCPServerCmd" /Users/goos/MoAI/moai-adk-go/.claude/worktrees/trend-mcp/internal/cli/root.go  # → ":214"

# 5. Confirm mcp-matrix is local-only (out-of-scope)
ls /Users/goos/MoAI/moai-adk-go/.claude/worktrees/trend-mcp/internal/template/templates/.moai/config/sections/mcp-matrix.yaml 2>&1 | head -1  # → "No such file"

# 6. Establish lint baseline
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/trend-mcp && golangci-lint run --timeout=2m ./internal/cli/... 2>&1 | tail -5

# 7. Cross-platform build feasibility
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/trend-mcp && go build ./... && GOOS=windows GOARCH=amd64 go build ./...
```

## §D. Constraints (DO NOT VIOLATE)

- DO NOT hand-edit `.mcp.json` or `~/.claude.json` — every write goes through `mutateClaudeJSONAtomic` (REQ-TMC-005).
- DO NOT fork the atomic-RMW seam — reuse `mutateClaudeJSONAtomic`, `writeClaudeJSONBytes`, `backupClaudeJSON`, `readClaudeJSONRaw`, `withClaudeJSONLock` UNCHANGED (REQ-TMC-008).
- DO NOT serialize resolved secrets into `.mcp.json` — only `${VAR}` literals (REQ-TMC-002, REQ-TMC-009, C3).
- DO NOT call `AskUserQuestion` from the new CLI subcommands (C-HRA-008, REQ-TMC-010).
- DO NOT present Semgrep/Codecov/audit tools as gate replacements (REQ-TMC-014, C7).
- DO NOT touch `glm_tools.go` body, `mcp_server.go`, `mcp_audit.go`, `mcp_codex.go`, `mcp_glm.go`, or any existing MCP test file (PRESERVE list §A.4).
- DO NOT touch `mcp-matrix.yaml` or its template mirror (SPEC-HARNESS-MCP-PROVISION-001 scope).
- Conventional Commits format required (`feat(SPEC-TREND-MCP-001): M{N} <subject>`); `🗿 MoAI` trailer; never `--no-verify`.
- `make build` after every template edit; mirror to local repo-root `.mcp.json` after build.

## §E. Self-Verification (Milestone completion — manager-develop §E)

The §E.1 plan-phase audit-ready signal is populated at plan-phase close. The §E.2/§E.3/§E.4 sections are populated during run/sync phases per the canonical progress.md Section Map (`.claude/rules/moai/development/spec-frontmatter-schema.md`).

## §F. Milestones

### M1 — Template `.mcp.json` + §25 neutrality (Priority High)

1. Author `internal/template/templates/.mcp.json` with 4 active entries (`context7`, `chrome-devtools`, `playwright`, `moai` opt-in placeholder via `$comment`) + 1 documented-only catalogue entry (`ast-grep`).
2. Run `make build`; verify the embedded catalog regenerates and the repo-root `.mcp.json` aligns.
3. Run `go test -run TestTemplateNeutrality ./internal/template/...` — must PASS (no SPEC IDs / SHAs / macOS paths / PR refs in the new `.mcp.json`).
4. Run `go test -run TestInternalContentLeak ./internal/template/...` — must PASS.
5. Add `internal/template/mcp_template_neutrality_test.go` regression guard (asserts the template `.mcp.json` parses + every entry is secret-free + no forbidden tokens).

### M2 — Generic `moai mcp add|remove|list` CLI (Priority High)

1. Author `internal/cli/mcp.go` with three subcommands; reuse `mutateClaudeJSONAtomic` for `add`/`remove` and `readClaudeJSON` for `list`.
2. Register `newMCPCmd()` at `internal/cli/root.go` (single-line `rootCmd.AddCommand(newMCPCmd())`).
3. Implement `--scope user|project` flag (reuse `resolveConfigPath` from `glm_tools.go:362`).
4. Implement `--env KEY=VAL` secret-rejection (REQ-TMC-009): VALUE must match `^\$\{[A-Z_][A-Z0-9_]*\}$` or be rejected.
5. Implement `--type stdio|http` flag (selects entry shape).
6. Author `internal/cli/mcp_test.go` with table-driven coverage: idempotent add, partial-remove safety (unrelated entries preserved), secret-rejection, concurrent-writer (inject `claudeJSONGuardPreLockHook`), `--json` output, `--scope` resolution.
7. Author `internal/cli/mcp_boundary_test.go` `TestMCP_NoAskUserQuestion` (C-HRA-008 static guard).

### M3 — Doctrine update + recipe catalogue + skip rationale (Priority Medium)

1. Update `.claude/rules/moai/core/settings-management.md` § MCP Configuration per REQ-TMC-013 (multi-entry provisioning contract, reconciled "no third-party entries THAT CARRY SECRETS..." wording).
2. Mirror to `internal/template/templates/.claude/rules/moai/core/settings-management.md` (Template-First).
3. Run `make build`; verify mirror.
4. Author `.moai/docs/mcp-recipes.md` covering all 10 tools per REQ-TMC-011 (Playwright, ast-grep, Semgrep, GitHub MCP, Postgres/neon, Sentry, Codecov, Sequential Thinking skip, Filesystem skip, Git skip, Memory/KG skip, Brave/Exa skip).
5. Each opt-in recipe carries both a copy-pasteable `.mcp.json` snippet AND a one-line `moai mcp add ...` equivalent (REQ-TMC-012).
6. Each recipe carries the "supply, do not redefine gates" note (REQ-TMC-014).

### M4 — Cross-cutting hardcoding + envkeys + final lint (Priority Medium)

1. Audit `mcp.go` for any env-var name, threshold, or default; lift to `internal/config/envkeys.go` / `defaults.go` constants per REQ-TMC-015.
2. Run `go test ./internal/cli/...` — full cli package green (catches cascading failures from the new subcommand).
3. Run `golangci-lint run --timeout=2m ./internal/cli/...` — zero NEW findings (baseline established at pre-flight).
4. Run `GOOS=windows GOARCH=amd64 go build ./...` — cross-platform build clean (no syscall package use; no platform-specific code expected).
5. Run `go test -race ./internal/cli/...` — race-detector clean (the new CLI touches the same flock + compare-retry path the GLM CLI does).

## §G. Anti-Patterns

- **AP-TMC-001**: hand-editing `.mcp.json` or `~/.claude.json` (bypasses the atomic-RMW guard; mid-session corruption hazard per SPEC-CLIFIX-CONCURRENCY-001).
- **AP-TMC-002**: forking `mutateClaudeJSONAtomic` into a second lock convention (REQ-TMC-008 violation; produces two sources of truth for the RMW discipline).
- **AP-TMC-003**: serializing a resolved secret into the template `.mcp.json` (REQ-TMC-002 / C3 violation; git-tracked credential leak).
- **AP-TMC-004**: calling `AskUserQuestion` from the new CLI subcommands (C-HRA-008 violation; subagent boundary break).
- **AP-TMC-005**: presenting Semgrep / Codecov / audit backends as gate replacements (REQ-TMC-014 violation; gate-conflation hazard).
- **AP-TMC-006**: editing the `glm_tools.go` body to add `mcp.go`-specific logic (PRESERVE list violation; scope-discipline break).
- **AP-TMC-007**: shipping the template `.mcp.json` without running `make build` (catalog staleness → committed-tree staleness → CI parity failure).
- **AP-TMC-008**: re-tiering to Tier S without striking the doctrine-update + template-`.mcp.json` REQs (scope/tier drift; plan-auditor regression).

## §H. Cross-References

- `.moai/specs/SPEC-TREND-MCP-001/spec.md` (this SPEC's requirements).
- `.moai/specs/SPEC-TREND-MCP-001/acceptance.md` (Given-When-Then for every AC-TMC-NNN).
- `.moai/specs/SPEC-MOAI-MCP-SERVER-001/spec.md` (dependency; REQ-MCP-002 reconciliation target).
- `.moai/specs/SPEC-CLIFIX-CONCURRENCY-001/` (atomic-RMW seam owner — `mutateClaudeJSONAtomic`).
- `.moai/specs/SPEC-HARNESS-MCP-PROVISION-001/` (per-project-type `/moai project` MCP matrix owner — distinct surface).
- `.moai/specs/SPEC-GLM-MCP-001/` (AUTH-bearing Z.AI registrar — `${Z_AI_API_KEY}`-literal pattern source).
- `.claude/rules/moai/core/settings-management.md` § MCP Configuration (doctrine surface to update).
- `.github/workflows/template-neutrality-check.yaml` (§25 CI guard — C1/C2/C4/C5/C6/C8).
- `internal/template/internal_content_leak_test.go` (§25 CI guard — SPEC-ID/SHA/date).
- `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.7 (design source).
