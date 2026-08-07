# progress.md — SPEC-TREND-MCP-001

> Tier M. Lifecycle: plan → run → sync. progress.md Section Map: `.claude/rules/moai/development/spec-frontmatter-schema.md` § progress.md Section Map. The `§E.*` namespace is parser-load-bearing for era classification — do NOT rename `§E.N` headings or the `sync_commit_sha` field (era.go matches them literally).

## §E.1 Plan-phase Audit-Ready Signal

_(populated at plan-phase close)_

- `plan_status:` audit-ready
- `plan_complete_at:` _(pending plan-phase commit)_
- `plan_artifact_set:` spec.md + plan.md + acceptance.md + progress.md (Tier M, 4 artifacts incl. progress.md skeleton)
- `tier:` M
- `req_count:` 16
- `ac_count:` 18
- `needs_clarification:` 1 (§B.B1 Tier/scope decision — full Tier M vs collapsed Tier S; orchestrator's user-question channel resolves before Implementation Kickoff Approval)

## §E.2 Run-phase Evidence

M1 — Template `.mcp.json` + §25 neutrality (commit `c436bbd0d`)
- AC-TMC-001 PASS: `jq '.mcpServers | keys' internal/template/templates/.mcp.json` → `["chrome-devtools","context7","playwright"]` (exactly 3). ast-grep + moai omitted from active map; documented in `.moai/docs/mcp-recipes.md` (REQ-TMC-003/004). No `$comment` form (`grep -c '\$comment' internal/template/templates/.mcp.json` → 0).
- AC-TMC-002 PASS: `go test -run TestTemplateNeutrality ./internal/template/...` → ok (PASS, no SPEC ID / SHA / macOS path / `CLAUDE.local.md` ref / `PR #N` ref in the new `.mcp.json`).
- AC-TMC-003 PASS: `go test -run TestTemplateNoInternalContentLeak ./internal/template/...` → ok (PASS, no SPEC-ID/SHA/date leak).
- AC-TMC-004 PASS: distributed default carries no runnable `moai` entry (omitted from active map per REQ-TMC-003; `moai mcp-server` stays opt-in via `moai init`/`moai web`).
- AC-TMC-016 PASS: `diff internal/template/templates/.mcp.json .mcp.json` → empty (byte-identical repo-root mirror).
- Regression guard: `internal/template/mcp_template_neutrality_test.go::TestMCPNeutralityTemplateShape` asserts shape + secret hygiene + leak scan.

M2 — Generic `moai mcp add|remove|list` CLI (commit `79782e023`)
- AC-TMC-005 PASS: `go test -run TestMCP_Add_RegistersEntry_PreservesUnrelated ./internal/cli` → ok (atomic-RMW add; unrelated context7/chrome-devtools/playwright preserved; `.claude.json.bak-*` created).
- AC-TMC-005 PASS (concurrent-writer sub-scenario): `go test -run TestMCP_Add_ConcurrentWriter ./internal/cli` → ok (claudeJSONGuardPreLockHook injection; both external write + new entry survive).
- AC-TMC-006 PASS: `go test -run TestMCP_Add_IdempotentSkip ./internal/cli` → ok (exactly ONE `.claude.json.bak-*` after two identical adds).
- AC-TMC-007 PASS: `go test -run TestMCP_Remove_PartialDelete ./internal/cli` → ok (only `my-tool` removed; zai-mcp-server + 3 baseline preserved).
- AC-TMC-008 PASS: `go test -run TestMCP_List_JSON ./internal/cli` → ok (valid JSON; 4 entries; distinguishes stdio/http; flags `${SEMGREP_API_TOKEN}` env ref).
- AC-TMC-009 PASS: `go test -run TestMCP_Add_SecretRejection ./internal/cli` → ok (positional secret value → structured error pointing to `${VAR}` form; NO entry written).
- AC-TMC-010 PASS: `go test -run TestMCP_NoAskUserQuestion ./internal/cli` → ok (zero `AskUserQuestion`/`mcp__askuser` in mcp.go).
- AC-TMC-015 PASS: `grep -c 'os.Getenv("' internal/cli/mcp.go` → 0 (no inline env-var reads; AC-TMC-015).
- REQ-TMC-008 inline-verified: `grep -c '^func mutateClaudeJSONAtomic' internal/cli/*.go | grep -v ':0'` → `internal/cli/glm_tools.go:1` (single definition; signatures unchanged at glm_tools.go:467/476/541/634/655; mcp.go CALLS the helpers, never redefines them).

M3 — Doctrine reconciliation + recipe catalogue (commit `64773cfef`)
- AC-TMC-011 PASS: `grep -c '^## ' .moai/docs/mcp-recipes.md` → 12 (≥10).
- AC-TMC-012 PASS: byte-equivalence verified via Go seam (`moaiMcpAdd semgrep --type http --url … --headers 'Authorization=Bearer ${SEMGREP_API_TOKEN}'` produces `{"headers":{"Authorization":"Bearer ${SEMGREP_API_TOKEN}"},"type":"http","url":"https://semgrep.example.com/mcp"}` — byte-identical to the catalogue snippet after normalization).
- AC-TMC-013 PASS: `grep -A2 'no third-party' internal/template/templates/.claude/rules/moai/core/settings-management.md | grep -c 'SECRETS\|credentials\|§25\|neutrality'` → 1 (≥1).
- AC-TMC-014 PASS: `grep -c 'supply, do not redefine\|Supply, do not redefine' .moai/docs/mcp-recipes.md` → 7 (≥5).

Cross-cutting (all milestones)
- `go test -count=1 ./internal/cli/...` → ok (full cli package green; no cascading failures).
- `go test -race -run TestMCP_ ./internal/cli` → ok (flock + compare-retry race-clean).
- `GOOS=windows GOARCH=amd64 go build ./...` → green (no platform-specific code; no syscall package).
- `golangci-lint run --timeout=3m ./internal/cli/` → 0 issues.
- `make build` → green (catalog.yaml regenerated; embedded FS recompiled).

## §E.3 Run-phase Audit-Ready Signal

- `run_status:` audit-ready
- `run_complete_at:` 2026-08-07
- `run_commit_sha:` 64773cfef (M3 head; M1 c436bbd0d, M2 79782e023, M3 64773cfef)
- `ac_pass_count:` 16
- `ac_fail_count:` 0
- `preserve_list_post_run_count:` 5 (glm_tools.go body, mcp_server.go, mcp_audit.go, mcp_codex.go, mcp_glm.go + their tests — all verified untouched per the PRESERVE list §A.4)
- `m1_to_mN_commit_strategy:` per-milestone Conventional Commit (M1 feat template/.mcp.json first introduction; M2 feat generic CLI; M3 feat doctrine + recipes)
- `l44_pre_commit_fetch:` not applicable (Route B / feat branch; pre-spawn sync check not required for this delegation)
- `l44_post_push_fetch:` pending push
- `new_warnings_or_lints_introduced:` 0
- `cross_platform_build.linux:` green
- `cross_platform_build.darwin:` green
- `cross_platform_build.windows:` green (GOOS=windows GOARCH=amd64)
- `total_run_phase_files:` 8 (internal/template/templates/.mcp.json NEW; internal/template/mcp_template_neutrality_test.go NEW; .mcp.json MODIFIED; internal/cli/mcp.go NEW; internal/cli/mcp_test.go NEW; internal/cli/mcp_boundary_test.go NEW; internal/cli/root.go MODIFIED; .moai/docs/mcp-recipes.md NEW; .claude/rules/moai/core/settings-management.md + template mirror MODIFIED)

## §E.4 Sync-phase Audit-Ready Signal

_(pending sync-phase)_

`sync_commit_sha:` _(pending sync-phase backfill)_
