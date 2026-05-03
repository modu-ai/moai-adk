# SPEC-GLM-MCP-001 Compact Reference

> **Purpose**: Run-phase token-efficient extract of REQs + critical ACs only.
> **Token savings**: ~30% vs full spec.md + acceptance.md.
> **Single source of truth**: `spec.md`, `plan.md`, `acceptance.md`. This file is a derived projection.

## Identity

- ID: SPEC-GLM-MCP-001
- Domain: GLM-MCP
- Status: draft
- Priority: Medium
- Recommended option: **Opt-in subcommand** (`moai glm tools enable|disable`)

## EARS Distribution

| Pattern | Count | REQ IDs |
|---------|-------|---------|
| Ubiquitous | 2 | 001, 002 |
| Event-Driven | 4 | 003, 004, 005, 006 |
| State-Driven | 1 | 007 |
| Optional | 1 | 008 |
| Unwanted | 2 | 009, 010 |
| **Total** | **10** | — |

## Requirements (Compact)

- **REQ-GMC-001** (Ubiquitous): `moai glm tools enable|disable [vision|websearch|webreader|all]` subcommand exists and is idempotent.
- **REQ-GMC-002** (Ubiquitous): No conflict with SPEC-GLM-001 (`DISABLE_BETAS=1`, `DISABLE_PROMPT_CACHING=1`). MCP registration is orthogonal to env var policy.
- **REQ-GMC-003** (Event-Driven): WHEN `enable` runs AND Node>=22 AND `GLM_AUTH_TOKEN` set → write `mcpServers.zai-mcp-server` to `~/.claude.json` with `command: "npx"`, `args: ["-y", "@z_ai/mcp-server@latest"]`, `env.Z_AI_API_KEY: <token>`, `env.Z_AI_MODE: "ZAI"`. Print activated tools + Pro plan + Claude Code restart guidance.
- **REQ-GMC-004** (Event-Driven): WHEN `disable` runs → remove `zai-mcp-server` entry; preserve all other mcpServers entries; print removed tools.
- **REQ-GMC-005** (Event-Driven): WHEN `enable`/`disable` runs → backup `~/.claude.json` to `~/.claude.json.bak-<ISO ts>` before modification. Skip backup on idempotent skip (no change).
- **REQ-GMC-006** (Event-Driven): WHEN `enable` runs AND existing `zai-mcp-server` entry present → (a) idempotent skip if token matches, (b) refuse + show `--force` guidance if token differs (R1 mitigation, no overwrite).
- **REQ-GMC-007** (State-Driven): WHILE `GLM_AUTH_TOKEN` absent or empty → reject `enable`; output token-registration guide (`moai github auth glm <token>` or equivalent).
- **REQ-GMC-008** (Optional): WHERE `--scope project` flag present → write to project-root `.mcp.json` instead of `~/.claude.json` (default user scope).
- **REQ-GMC-009** (Unwanted): IF `node` absent OR version < v22.0.0 → graceful fail with detected version + min requirement + install guidance; non-zero exit; `~/.claude.json` unchanged.
- **REQ-GMC-010** (Unwanted): IF user-defined or third-party `mcpServers` entries present → enable/disable MUST NOT modify them. Scope limited to `mcpServers.zai-mcp-server` key only.

## Critical Acceptance Criteria (Compact)

(Full GWT scenarios in `acceptance.md`. Below: 10 must-pass scenarios.)

| ID | Scenario | Expected Result |
|----|----------|-----------------|
| GWT-1 | `enable vision` twice → second run is no-op | First run creates entry; second run idempotent skip |
| GWT-3 | `moai cc` ↔ `moai glm` mode switch after enable | SPEC-GLM-001 env policy preserved; zai-mcp-server entry retained |
| GWT-4 | `enable vision` from clean state | `~/.claude.json` `mcpServers.zai-mcp-server` has 4 fields exactly |
| GWT-7 | `disable all` with 3 other mcpServers entries | Only zai-mcp-server removed; other 3 preserved verbatim |
| GWT-8 | `enable vision` from clean state | `~/.claude.json.bak-<ts>` created with pre-modification contents |
| GWT-11 | `enable` with existing entry but mismatched token | Non-zero exit; `--force` guidance; `~/.claude.json` unchanged |
| GWT-12 | `enable vision` with no `GLM_AUTH_TOKEN` | Non-zero exit; token registration guide; `~/.claude.json` unchanged |
| GWT-13 | `enable vision --scope project` | `.mcp.json` (project root) written; `~/.claude.json` untouched |
| GWT-14 | `enable all` with `node` absent on PATH | Non-zero exit; install guidance; `~/.claude.json` unchanged |
| GWT-16 | `enable vision` with user-defined `my-custom-server` entry | `my-custom-server` preserved verbatim; new `zai-mcp-server` added |

## Quality Gates (Compact)

- All 10 REQs covered by ≥1 GWT scenario each
- `go test ./...` passes (no regression)
- `go test -race ./internal/cli/...` passes
- `go vet ./...` zero warnings
- New package coverage ≥85%
- Both `settings-management.md` files (template + root mirror) updated and identical
- `make build` regenerates `internal/template/embedded.go`
- 16-language neutrality preserved (CLAUDE.local.md §15)

## Exclusions

1. **EX-1**: Mainland China (`Z_AI_MODE=ZHIPU`) deferred to v0.2
2. **EX-2**: GLM plan tier (Lite vs Pro) pre-validation skipped — runtime 401 surfaces from Z.AI directly
3. **EX-3**: Other mcpServers entries (`context7`, `sequential-thinking`, `moai-lsp`) policies unchanged
4. **EX-4**: Mode switch (`cc` ↔ `glm` ↔ `cg`) does NOT auto enable/disable zai-mcp-server (explicit user command only)
5. **EX-5**: `~/.claude.json` schema strict validation deferred (Claude Code runtime validates)
6. **EX-6**: Pre-flight sample test (Vision OCR ping) deferred to optional `--test` flag in v0.2

## Risk-to-REQ Mapping (Compact)

| Risk | Mitigation REQ |
|------|----------------|
| R1: Existing entry overwrite | REQ-GMC-006 (`--force` guard) |
| R2: Lite plan user, Vision 401 | EX-2 (clear messaging, no pre-check) |
| R3: Node.js < 22 | REQ-GMC-009 (graceful fail) |
| R4: User vs project scope | REQ-GMC-008 (`--scope project` opt-in) |
| R5: Mode switch corruption | EX-4 (no auto-toggle) |
| R6: Other mcpServers damage | REQ-GMC-010 (scope limited to one key) |
| R7: Atomic write failure | REQ-GMC-005 (backup before write) |

## Files Affected (Anticipated)

- `internal/cli/glm_tools.go` (new) or `internal/cli/glm.go` (extension)
- `internal/cli/glm_tools_test.go` (new)
- `internal/template/templates/.claude/rules/moai/core/settings-management.md` (template)
- `.claude/rules/moai/core/settings-management.md` (mirror)
- `CHANGELOG.md`

## Plan Audit Readiness

- ✓ 9-field frontmatter canonical schema (`id`, `version`, `status`, `created_at`, `updated_at`, `author`, `priority`, `labels`, `issue_number`)
- ✓ EARS distribution table accurate (Section 3.0): 2+4+1+1+2 = 10
- ✓ Exclusions section ≥1 entry (6 entries)
- ✓ acceptance.md self-contained (no external defer, D5 anti-pattern avoided)
- ✓ research.md present (Phase 0.5 deep research)
- ✓ Risks ≥1 with mitigation strategies (R1~R7)
