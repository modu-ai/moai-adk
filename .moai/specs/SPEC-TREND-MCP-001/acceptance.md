# acceptance.md — SPEC-TREND-MCP-001

> Given-When-Then acceptance criteria. Tier M. 16 ACs across 16 REQs. Each AC is binary-testable and traces to exactly one REQ-TMC-NNN. Verification-claim-integrity §3 applies: every PASS row cites a verbatim command + observed output (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk).
>
> v0.1.1 — AC-TMC-009 (seam-reuse) folded into REQ-TMC-008 inline traceability (no standalone AC; the requirement is verified by M2 code inspection); AC-TMC-012 (concurrent-author) folded into AC-TMC-005 as a second scenario. Result: 16 ACs (was 18), within the Tier M ceiling.

## §D. AC Matrix

### M1 — Template-managed .mcp.json + bundled neutral entries

**AC-TMC-001** (Given-When-Then; traces to REQ-TMC-001, REQ-TMC-004)
- **Given** the template source at `internal/template/templates/.mcp.json` is authored and `make build` is run,
- **When** an auditor reads the regenerated catalog and the resulting `.mcp.json`,
- **Then** exactly three entries are active in the distributed default `mcpServers` map (`context7`, `chrome-devtools`, `playwright`); `moai` is opt-in (absent from the distributed default per REQ-TMC-003) and `ast-grep` is documented in the recipe catalogue but absent from the active map (default-disabled per REQ-TMC-004, activated via `moai mcp add ast-grep ...`). No JSONC/`$comment` form appears in the template.
- Verification: `jq '.mcpServers | keys' internal/template/templates/.mcp.json` → `["chrome-devtools", "context7", "playwright"]` (exactly 3); `grep -c "ast-grep" .moai/docs/mcp-recipes.md` → `≥1`; `grep -c '\\$comment' internal/template/templates/.mcp.json` → `0`.

**AC-TMC-002** (Given-When-Then; traces to REQ-TMC-002)
- **Given** the template `.mcp.json` is committed,
- **When** `go test -run TestTemplateNeutrality ./internal/template/...` runs,
- **Then** the test PASSES — the file contains no SPEC IDs, no commit SHAs, no macOS-bias paths, no `CLAUDE.local.md` references, no `PR #N` references, and no resolved secrets (every env reference is a `${VAR}` literal).
- Verification: `go test -run TestTemplateNeutrality -v ./internal/template/... 2>&1 | grep -E "PASS|FAIL"` → `PASS`.

**AC-TMC-003** (Given-When-Then; traces to REQ-TMC-002)
- **Given** the template `.mcp.json` is committed,
- **When** `go test -run TestInternalContentLeak ./internal/template/...` runs,
- **Then** the test PASSES — the SPEC-ID / commit-SHA / date-literal leak detector finds zero violations in the new `.mcp.json`.
- Verification: `go test -run TestInternalContentLeak -v ./internal/template/... 2>&1 | grep -E "PASS|FAIL"` → `PASS`.

**AC-TMC-004** (Given-When-Then; traces to REQ-TMC-003)
- **Given** the distributed template `.mcp.json`,
- **When** an auditor reads the `moai` entry,
- **Then** the entry is opt-in — it ships in a form that does NOT auto-start the local stdio server unless the user explicitly enables it (omitted-from-active-map, documented in the recipe catalogue), preserving MOAI-MCP-SERVER-001 REQ-MCP-002's single-server opt-in property.
- Verification: the active `mcpServers` map at distributed default does NOT include a runnable `moai` entry whose `command: "moai"` would be executed by the Claude Code runtime at session start without user action. Per plan.md §B.B4, the omitted-from-active-map rendering choice applies to both `ast-grep` and `moai` — both are documented-only at the distributed default, and the user runs `moai mcp add ...` to activate. The active distributed map is exactly `context7 + chrome-devtools + playwright` (3 entries).

### M2 — Generic atomic-RMW entry-management CLI

**AC-TMC-005** (Given-When-Then; traces to REQ-TMC-005; includes the concurrent-writer scenario formerly AC-TMC-012)
- **Given** a clean `.mcp.json` with the distributed 3-entry default,
- **When** the user runs `moai mcp add my-tool --command npx --args "-y" --args "my-tool-mcp" --scope project`,
- **Then** the entry is registered via `mutateClaudeJSONAtomic` (flock acquired, backup created, atomic publish), the resulting `.mcp.json` contains the entry, and every unrelated entry is preserved unchanged.
- Verification: `grep -c "my-tool-mcp" .mcp.json` → `1`; `grep -c "chrome-devtools\|context7\|playwright" .mcp.json` → `3` (unchanged); a `.mcp.json.bak-*` backup file exists.
- **Concurrent-writer sub-scenario** (merged from former AC-TMC-012): **Given** the `claudeJSONGuardPreLockHook` test injection is armed (simulating a concurrent Claude Code writer), **When** `moai mcp add my-tool` runs under that injection, **Then** the compare-retry path recovers within `claudeJSONGuardMaxRetries` (3) attempts and the entry lands WITHOUT losing the concurrent writer's changes. Verification: `go test -run TestMCP_Add_ConcurrentWriter -v ./internal/cli/... 2>&1 | grep -E "PASS|FAIL"` → `PASS`.

**AC-TMC-006** (Given-When-Then; traces to REQ-TMC-005)
- **Given** the same `moai mcp add my-tool` is run twice,
- **When** the second invocation completes,
- **Then** the second invocation is an idempotent skip (no backup, no write, exit 0 with a "no change" message) because `mcpEntryEqual` returns true.
- Verification: `moai mcp add my-tool ... 2>&1 | grep -i "no change\|idempotent\|already"` → `≥1` match; only ONE `.mcp.json.bak-*` file exists after both runs.

**AC-TMC-007** (Given-When-Then; traces to REQ-TMC-006)
- **Given** a `.mcp.json` with `my-tool` + `context7` + `chrome-devtools` + `playwright` + `zai-mcp-server` (the GLM entry),
- **When** the user runs `moai mcp remove my-tool --scope project`,
- **Then** ONLY the `my-tool` entry is removed; `context7`, `chrome-devtools`, `playwright`, and `zai-mcp-server` are all preserved unchanged.
- Verification: `grep -c "my-tool" .mcp.json` → `0`; `grep -c "chrome-devtools\|context7\|playwright\|zai-mcp-server" .mcp.json` → `4`.

**AC-TMC-008** (Given-When-Then; traces to REQ-TMC-007)
- **Given** a `.mcp.json` with 4 entries (3 stdio + 1 HTTP with `${VAR}` literal),
- **When** the user runs `moai mcp list --scope project --json`,
- **Then** the output is valid JSON, lists all 4 entries, distinguishes `type: "stdio"` from `type: "http"`, and flags the entry whose env contains a `${VAR}` literal.
- Verification: `moai mcp list --scope project --json | jq '.entries | length'` → `4`; `moai mcp list --scope project --json | jq '.entries[] | select(.type=="http") | .env_ref'` → flags the `${VAR}`-literal entry.

> **REQ-TMC-008 inline verification (no standalone AC — seam-reuse-unchanged is verified by M2 code inspection).** REQ-TMC-008 (seam reuse-unchanged) is verified at M2 by: (a) `grep -n "func mutateClaudeJSONAtomic\|func writeClaudeJSONBytes\|func backupClaudeJSON\|func readClaudeJSONRaw\|func withClaudeJSONLock" internal/cli/glm_tools.go` showing every signature unchanged (same file, same line range ±5, same parameter shape); (b) `grep -c "func mutateClaudeJSONAtomic" internal/cli/*.go` → `1` (only one definition, in `glm_tools.go`); (c) the new `internal/cli/mcp.go` calls them, never redefines them. The former standalone AC-TMC-009 was folded into this inline note because the requirement is self-verifying via the PRESERVE list + grep — a separate AC was redundant.

**AC-TMC-009** (Given-When-Then; traces to REQ-TMC-009)
- **Given** the user attempts `moai mcp add bad --command npx --env API_KEY=sk-secret-123 --scope project`,
- **When** the CLI parses the `--env` flag,
- **Then** the CLI rejects the value with a structured error pointing to the `${VAR}`-literal form, exits non-zero, and writes NO entry to `.mcp.json` (no partial write).
- Verification: `moai mcp add bad ... --env API_KEY=sk-secret-123; echo "exit=$?"` → exit non-zero; `grep -c "bad" .mcp.json` → `0`.

**AC-TMC-010** (Given-When-Then; traces to REQ-TMC-010)
- **Given** the new `internal/cli/mcp.go` is committed,
- **When** `go test -run TestMCP_NoAskUserQuestion -v ./internal/cli/...` runs,
- **Then** the test PASSES — zero `AskUserQuestion` or `mcp__askuser` references in `mcp.go` (excluding comments).
- Verification: `go test -run TestMCP_NoAskUserQuestion -v ./internal/cli/... 2>&1 | grep -E "PASS|FAIL"` → `PASS`.

### M3 — Recipe catalogue + skip rationale + doctrine reconciliation

**AC-TMC-011** (Given-When-Then; traces to REQ-TMC-011)
- **Given** the recipe catalogue at `.moai/docs/mcp-recipes.md` is authored,
- **When** `grep -c "^## " .moai/docs/mcp-recipes.md` runs,
- **Then** the catalogue carries at least 10 `## ` headings — one per tool classified in §3.7 (Playwright, ast-grep, Semgrep, GitHub MCP, Postgres/neon, Sentry, Codecov, Sequential Thinking skip, Filesystem skip, Git skip, Memory/KG skip, Brave/Exa skip).
- Verification: `grep -c "^## " .moai/docs/mcp-recipes.md` → `≥10`.

**AC-TMC-012** (Given-When-Then; traces to REQ-TMC-012)
- **Given** the Semgrep opt-in recipe in the catalogue,
- **When** the user applies the copy-pasteable `.mcp.json` snippet OR runs the one-line `moai mcp add semgrep ...` equivalent,
- **Then** the resulting `.mcp.json` entries are byte-equivalent after normalization (same `command`, same `args`, same `env` keys, same `type`).
- Verification: run both paths against a clean temp dir; `jq -S '.mcpServers.semgrep'` produces identical output.

**AC-TMC-013** (Given-When-Then; traces to REQ-TMC-013)
- **Given** the doctrine update to `.claude/rules/moai/core/settings-management.md` is committed (and mirrored to template),
- **When** `grep -n "no third-party entries" internal/template/templates/.claude/rules/moai/core/settings-management.md` runs,
- **Then** the literal "no third-party entries" is either removed or qualified with "THAT CARRY SECRETS, require credentials, or fail §25 neutrality" — the narrower clause is the load-bearing one, and the multi-entry provisioning contract (Context7, chrome-devtools, Playwright default-on; ast-grep default-disabled; `moai` local stdio opt-in) is described explicitly.
- Verification: `grep -A 2 "no third-party" internal/template/templates/.claude/rules/moai/core/settings-management.md | grep -c "SECRETS\|credentials\|§25\|neutrality"` → `≥1`.

**AC-TMC-014** (Given-When-Then; traces to REQ-TMC-014)
- **Given** every opt-in recipe in the catalogue,
- **When** `grep -c "supply, do not redefine\|supply evidence\|gate replacement\|SSOT" .moai/docs/mcp-recipes.md` runs,
- **Then** at least one match per opt-in recipe is present — no recipe is presented as a gate replacement.
- Verification: count matches ≥ number of opt-in recipes (Semgrep + GitHub MCP + Postgres/neon + Sentry + Codecov = 5 minimum).

### Cross-cutting

**AC-TMC-015** (Given-When-Then; traces to REQ-TMC-015)
- **Given** the new `internal/cli/mcp.go` is committed,
- **When** `grep -E "os\.Getenv\(\"[A-Z" internal/cli/mcp.go` runs (search for inline env-var names),
- **Then** zero matches — every env-var name is referenced via a constant in `internal/config/envkeys.go`.
- Verification: `grep -c 'os.Getenv("' internal/cli/mcp.go` → `0` (any env access uses constants from `envkeys.go`).

**AC-TMC-016** (Given-When-Then; traces to REQ-TMC-016)
- **Given** the template `.mcp.json` + the doctrine mirror are committed,
- **When** `make build` runs and produces the regenerated catalog,
- **Then** the catalog includes the new `.mcp.json` and the regenerated `.claude/rules/moai/core/settings-management.md` mirror, AND the repo-root copies (`.mcp.json`, `.claude/rules/moai/core/settings-management.md`) are byte-aligned with the regenerated template output.
- Verification: `diff <(cat internal/template/templates/.mcp.json) <(cat .mcp.json)` → only allowed deviations (e.g. the local repo `.mcp.json` carries the same entries; deviations are documented in plan.md if any).

## §D.1 Severity Classification

| AC | Severity | Rationale |
|----|----------|-----------|
| AC-TMC-001, 002, 003, 004 | MUST | template `.mcp.json` is the load-bearing new surface; §25 neutrality is a binary CI gate; opt-in preservation reconciles a completed-SPEC invariant |
| AC-TMC-005, 006, 007, 009 | MUST | atomic-RMW + idempotency + partial-delete safety + secret-rejection + concurrent-writer recovery (folded into AC-TMC-005) — any failure corrupts the user's `.mcp.json` or leaks a secret |
| AC-TMC-008, 010 | SHOULD | `list` output shape + subagent-boundary static guard — important but not corruption-class |
| REQ-TMC-008 (inline verification, no standalone AC) | MUST | seam-reuse-unchanged is the hard scope-discipline invariant — forking the atomic-RMW guard is the named hazard; verified by M2 code inspection (grep signatures unchanged) |
| AC-TMC-011, 012, 013, 014 | SHOULD | catalogue + doctrine reconciliation — load-bearing for the user-facing contract but not corruption-class |
| AC-TMC-015, 016 | MUST | hardcoding prevention + Template-First mirror parity — CI-gate-class |

## §D.2 Traceability (REQ ↔ AC)

| REQ | AC | Severity |
|-----|----|---------|
| REQ-TMC-001 (template `.mcp.json` 3-active-2-documented structural) | AC-TMC-001 | MUST |
| REQ-TMC-002 (no secrets / §25 neutrality) | AC-TMC-002, AC-TMC-003 | MUST |
| REQ-TMC-003 (`moai` opt-in preserved) | AC-TMC-004 | MUST |
| REQ-TMC-004 (ast-grep default-disabled via omission + catalogue + `moai mcp add`) | AC-TMC-001 (same active-map assertion) | MUST |
| REQ-TMC-005 (generic add via atomic-RMW, incl. concurrent-writer recovery) | AC-TMC-005, AC-TMC-006 | MUST |
| REQ-TMC-006 (generic remove partial-delete) | AC-TMC-007 | MUST |
| REQ-TMC-007 (generic list shape) | AC-TMC-008 | SHOULD |
| REQ-TMC-008 (seam reuse-unchanged) | inline verification (no standalone AC — see §D note above) | MUST |
| REQ-TMC-009 (secret-rejection) | AC-TMC-009 | MUST |
| REQ-TMC-010 (subagent boundary) | AC-TMC-010 | SHOULD |
| REQ-TMC-011 (catalogue coverage) | AC-TMC-011 | SHOULD |
| REQ-TMC-012 (snippet ↔ CLI byte-equivalence) | AC-TMC-012 | SHOULD |
| REQ-TMC-013 (doctrine reconciliation) | AC-TMC-013 | SHOULD |
| REQ-TMC-014 (MCP-not-gate-replacement) | AC-TMC-014 | SHOULD |
| REQ-TMC-015 (hardcoding prevention) | AC-TMC-015 | MUST |
| REQ-TMC-016 (Template-First mirror) | AC-TMC-016 | MUST |

## §D.3 Indirect Verification (gates not directly tested by an AC)

- **Template-First parity**: covered by `make build` + `internal/template/templates_test.go` (pre-existing) — AC-TMC-016 verifies the `.mcp.json` mirror specifically.
- **§25 neutrality**: covered by `template-neutrality-check.yaml` CI workflow + `internal_content_leak_test.go` — AC-TMC-002/003 verify the new file specifically.
- **C-HRA-008 subagent boundary**: covered by the canonical `TestWeb_NoAskUserQuestion` pattern — AC-TMC-010 mirrors it for `mcp.go`.
- **Atomic-RMW correctness under concurrent access**: covered by the existing `glm_tools_test.go` injection-point suite — AC-TMC-005's concurrent-writer sub-scenario extends it for the generic CLI.
- **Seam reuse-unchanged (REQ-TMC-008)**: covered by M2 code inspection (grep signatures unchanged) — no standalone AC; see §D inline note.

## §D.4 Closure Gates (forward-looking checks)

- **G1 — go test ./...**: full test suite green at run-phase close (catches cascading failures from the new subcommand).
- **G2 — golangci-lint run**: zero NEW findings vs pre-flight baseline.
- **G3 — GOOS=windows GOARCH=amd64 go build ./...**: cross-platform build clean.
- **G4 — go test -race ./internal/cli/...**: race-detector clean (flock + compare-retry path).
- **G5 — template-neutrality-check.yaml CI workflow passes** on the PR.
- **G6 — spec-lint passes** on this SPEC (GEARS compliance, Out-of-Scope rule, frontmatter schema).

## §D.5 Definition of Done

- All MUST ACs PASS with verbatim evidence (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk per verification-claim-integrity §3).
- All SHOULD ACs PASS OR carry a documented, low-severity gap with a remediation owner.
- The §B.B1 Tier/scope decision is RESOLVED (Tier M confirmed, 2026-08-07) — no `[NEEDS CLARIFICATION]` marker remains in plan.md; the SPEC proceeds to Implementation Kickoff Approval at Tier M.
- The doctrine reconciliation with MOAI-MCP-SERVER-001 REQ-MCP-002 is documented in the SPEC body AND the settings-management.md doctrine update.
- The new template `.mcp.json` passes §25 (both CI guards) on first commit.
- The new generic CLI reuses `mutateClaudeJSONAtomic` unchanged (no fork; verified by M2 grep).
- The recipe catalogue covers all 10 §3.7 tools with both snippet + CLI-equivalent forms and the "supply, do not redefine" note.
