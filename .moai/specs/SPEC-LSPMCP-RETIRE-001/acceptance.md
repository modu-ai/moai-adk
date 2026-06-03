# Acceptance Criteria — SPEC-LSPMCP-RETIRE-001

## §A — Given-When-Then Scenarios

### Scenario 1 — Dead branch removed, build green (happy path)

- **Given** the `internal/mcp` package and `internal/cli/mcp.go` exist and `internal/cli/mcp.go`
  is the only non-test importer of `internal/mcp`,
- **When** the run-phase deletes the `internal/mcp` package directory and `internal/cli/mcp.go`
  and runs `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...`,
- **Then** both builds exit 0, no source file imports `github.com/modu-ai/moai-adk/internal/mcp`,
  and `moai --help` no longer lists the `mcp` command.

### Scenario 2 — Template edited, rendered JSON still valid, tests green

- **Given** `.mcp.json.tmpl` contains a `moai-lsp` server entry alongside `context7` and a
  `staggeredStartup` block, and two `settings_test.go` functions assert the `moai-lsp` entry,
- **When** the run-phase removes the `moai-lsp` entry, runs `make build`, removes the two
  moai-lsp-dependent test functions, and runs `go test ./internal/template/...`,
- **Then** the rendered `.mcp.json.tmpl` is valid JSON on both the windows and non-windows
  template branches, the `context7` entry and `staggeredStartup` block are unchanged, and the
  template test suite passes with zero failures.

### Scenario 3 — internal/lsp untouched (preserve boundary)

- **Given** the `internal/lsp` package (the active LSP client with 10 consumers) exists,
- **When** the entire retirement is applied,
- **Then** `git diff --quiet -- internal/lsp/` reports no changes (exit 0), and
  `internal/cli/deps.go` `gopls.NewBridge` wiring is unchanged.

### Scenario 4 — Predecessor superseded (state transition)

- **Given** `SPEC-LSPMCP-001/spec.md` has frontmatter `status: archived`,
- **When** the run-phase applies the supersession,
- **Then** the predecessor frontmatter shows `status: superseded` and a
  `superseded_by: SPEC-LSPMCP-RETIRE-001` field, and the predecessor body content is otherwise
  unchanged.

### Scenario 5 — Hidden importer (edge case / blocker)

- **Given** the pre-flight isolation check `grep "moai-adk/internal/mcp" --include="*.go"`
  (excluding `_test.go` and `internal/mcp/`),
- **When** the result returns more than one line (a hidden importer beyond `internal/cli/mcp.go`),
- **Then** the run-phase HALTS, does NOT delete the package, and returns a structured blocker
  report naming the hidden importer so the orchestrator can widen scope.

## §B — Edge Cases

- **Trailing-comma after `context7` block**: removing the `moai-lsp` entry (the last entry in
  `mcpServers`) must not leave a dangling comma after the `context7` object — the rendered
  output must parse as valid JSON.
- **Windows template branch**: the `{{- if eq .Platform "windows"}}` / `{{- else}}` branches
  inside `context7` must both still render valid JSON after the moai-lsp removal.
- **Mirror byte-parity**: the two settings-management.md copies must end byte-identical; an
  asymmetric edit must be caught by `embedded_mirror_test.go`.
- **glm_tools_test.go synthetic fixture**: this test builds its own `moai-lsp` map entry and
  does NOT read the real template — it must continue to pass unchanged (it is verifying GLM
  tooling preserves arbitrary MCP entries, not the moai-lsp template entry specifically).

## §C — Definition of Done

- [ ] `internal/mcp/` directory removed (all 10 files).
- [ ] `internal/cli/mcp.go` removed; `moai mcp` / `moai mcp lsp` absent from `moai --help`.
- [ ] `.mcp.json.tmpl` `moai-lsp` entry removed; `context7` + `staggeredStartup` intact.
- [ ] `make build` run; `embedded.go` regenerated and consistent.
- [ ] `settings_test.go` moai-lsp test dependencies removed; template tests pass.
- [ ] Both settings-management.md copies edited byte-identically; mirror-parity test passes.
- [ ] `SPEC-LSPMCP-001` transitioned `archived → superseded` with `superseded_by` field.
- [ ] `internal/lsp` byte-unchanged (preserve guard passes).
- [ ] Host + windows cross-compile both exit 0.
- [ ] Full test suite green (`go test ./...`).
- [ ] No internal-content-isolation violation introduced into `templates/**`.

## §D — AC Matrix (binary PASS/FAIL, with verification commands)

| AC | REQ | Description | Verification Command | Expected |
|----|-----|-------------|----------------------|----------|
| AC-LSPMCP-RETIRE-001 | REQ-001 | `internal/mcp` package removed | `test ! -d internal/mcp && echo REMOVED` | `REMOVED` |
| AC-LSPMCP-RETIRE-002 | REQ-002 | `moai mcp` CLI surface removed | `test ! -f internal/cli/mcp.go && echo REMOVED` ; `go run ./cmd/moai --help \| grep -c "^  mcp "` | `REMOVED` ; `0` |
| AC-LSPMCP-RETIRE-003 | REQ-003 | `moai-lsp` template entry removed; context7 + staggeredStartup intact | `grep -c "moai-lsp" internal/template/templates/.mcp.json.tmpl` ; `grep -c "context7\|staggeredStartup" internal/template/templates/.mcp.json.tmpl` | `0` ; `≥2` |
| AC-LSPMCP-RETIRE-004 | REQ-004 | embedded.go regenerated, consistent | `make build && git diff --name-only \| grep -q internal/template/embedded.go && echo REGEN \|\| echo NOCHANGE` | `REGEN` or `NOCHANGE` (consistent either way) |
| AC-LSPMCP-RETIRE-005 | REQ-005 | moai-lsp template tests removed; suite green | `grep -c "moai-lsp" internal/template/settings_test.go` ; `go test ./internal/template/...` | `0` ; `ok` (PASS) |
| AC-LSPMCP-RETIRE-006 | REQ-006 | settings-management.md mirror pair edited identically | `grep -c "moai-lsp" .claude/rules/moai/core/settings-management.md internal/template/templates/.claude/rules/moai/core/settings-management.md` ; `go test ./internal/template/ -run TestRuleTemplateMirrorDrift` | `0` (both) ; PASS |
| AC-LSPMCP-RETIRE-007 | REQ-007 | predecessor superseded | `grep -c "status: superseded" .moai/specs/SPEC-LSPMCP-001/spec.md` ; `grep -c "superseded_by: SPEC-LSPMCP-RETIRE-001" .moai/specs/SPEC-LSPMCP-001/spec.md` | `1` ; `1` (per-token, no line-coincidence) |
| AC-LSPMCP-RETIRE-008 | REQ-008 | no dangling internal/mcp reference; both builds green | `grep -rc "moai-adk/internal/mcp" --include="*.go" . \| grep -v ":0" ; go build ./... ; GOOS=windows GOARCH=amd64 go build ./...` | no matches ; both exit 0 |
| AC-LSPMCP-RETIRE-009 | REQ-009 | rendered template valid JSON (both branches) | `go test ./internal/template/ -run TestMCPTemplate` (surviving subset renders + `json.Unmarshal`) | PASS |
| AC-LSPMCP-RETIRE-010 | REQ-010 | `internal/lsp` byte-unchanged (PRESERVE) | `git diff --quiet -- internal/lsp/ && echo PRESERVED \|\| echo CHANGED` | `PRESERVED` |

## §D.1 — Severity classification

| Severity | ACs | Gate behavior |
|----------|-----|---------------|
| MUST-PASS (blocking) | AC-001, AC-002, AC-003, AC-005, AC-008, AC-010 | Any FAIL blocks closure |
| MUST-PASS (state/docs) | AC-006, AC-007, AC-009 | FAIL blocks closure |
| SHOULD-PASS (consistency) | AC-004 | FAIL warns (embedded.go consistency — `make build` idempotent) |

AC-LSPMCP-RETIRE-010 (the `internal/lsp` preserve guard) is the single most important AC:
a CHANGED result means a different, live system was damaged and the change MUST be reverted.

## §D.2 — Quality gate criteria

- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` both exit 0.
- `go test ./...` green (zero failures); the removal shrinks the test surface.
- `golangci-lint run` introduces no NEW findings vs the pre-flight baseline.
- `moai spec lint` passes on the three new SPEC artifacts.
- Mirror-parity tests (`embedded_mirror_test.go` / `rule_template_mirror_test.go`) pass.
- No internal-content-isolation regression in `internal/template/templates/**`
  (no SPEC ID / REQ token / date / commit SHA introduced by the docs-mirror edit).

## §D.3 — Traceability

| REQ | AC | Milestone |
|-----|-----|-----------|
| REQ-LSPMCP-RETIRE-001 | AC-001 | M1 |
| REQ-LSPMCP-RETIRE-002 | AC-002 | M1 |
| REQ-LSPMCP-RETIRE-003 | AC-003 | M3 |
| REQ-LSPMCP-RETIRE-004 | AC-004 | M3 |
| REQ-LSPMCP-RETIRE-005 | AC-005 | M2 |
| REQ-LSPMCP-RETIRE-006 | AC-006 | M4 |
| REQ-LSPMCP-RETIRE-007 | AC-007 | M5 |
| REQ-LSPMCP-RETIRE-008 | AC-008 | M1 / M6 |
| REQ-LSPMCP-RETIRE-009 | AC-009 | M3 |
| REQ-LSPMCP-RETIRE-010 | AC-010 | M6 (continuous PRESERVE) |
