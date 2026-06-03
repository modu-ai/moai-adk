# Implementation Plan — SPEC-LSPMCP-RETIRE-001

> Tier M. cycle_type recommendation: **ddd** (ANALYZE-PRESERVE-IMPROVE) — this is a
> removal against an existing brownfield codebase where the dominant risk is breaking a
> different system (`internal/lsp`) by mistake. ANALYZE (confirm the dead branch) →
> PRESERVE (characterize the build + test green baseline + the PRESERVE boundary) →
> IMPROVE (delete, then re-verify) fits a deletion better than RED-GREEN-REFACTOR.
> The orchestrator MAY override to `tdd` if quality.yaml `development_mode: tdd` is set.

## §A — Context

- **Work location**: `/Users/goos/MoAI/moai-adk-go` (main checkout — this is the
  moai-adk-go dev project itself, NOT a downstream user project).
- **Branch / HEAD**: `main`, synced with origin/main at plan time
  (`git rev-list --count --left-right origin/main...HEAD` = `0 0`, no race).
- **SPEC artifacts**: `.moai/specs/SPEC-LSPMCP-RETIRE-001/{spec.md, plan.md, acceptance.md}`.
- **Predecessor**: `SPEC-LSPMCP-001` (status: archived, phase "v2.x - Legacy") — this SPEC
  supersedes it.
- **Existing infrastructure**:
  - REMOVE target: `internal/mcp` package (10 files, ~1091 LOC incl. tests) +
    `internal/cli/mcp.go` + the `moai-lsp` entry in `.mcp.json.tmpl`.
  - PRESERVE target: `internal/lsp` (entire suite, 10 consumers), `gopls.NewBridge` wiring
    in `deps.go`, `moai lsp doctor` (`lsp_doctor.go`), the `context7` + `staggeredStartup`
    template entries.
- **Tier rationale**: ~8-10 files affected (Tier M 5-15 range); LOC change is
  deletion-dominant (~1091 removed). Multi-domain (Go package + CLI + template +
  test-guards + docs mirror) and a critical preserve boundary warrant the 3-file artifact
  set and the Section A-E delegation template.

## §B — Known Issues (auto-injected, domain-filtered)

**B1. Cross-platform build tags** — `internal/cli/mcp.go` imports `syscall` (uses
`syscall.SIGINT, syscall.SIGTERM`). Removing the file removes that usage; no NEW build-tag
work needed, but the run-phase MUST verify `GOOS=windows GOARCH=amd64 go build ./...`
still passes after removal (the `.mcp.json.tmpl` has `{{- if eq .Platform "windows"}}`
branches — verify JSON validity on both).

**B2. Cross-SPEC policy conflict scan** — `SPEC-CC2122-MCP-001` (implemented) owns the
`context7 alwaysLoad: true` logic; `SPEC-LSP-CORE-002`/`SPEC-LSP-FLAKY-002` own
`internal/lsp`. Neither must be touched. Run `grep -rn "Retired\|superseded" internal/mcp`
returns no conflict markers (the package was never marked retired — this SPEC introduces
the retirement).

**B4. Frontmatter canonical schema** — predecessor `SPEC-LSPMCP-001` frontmatter uses
`status: archived`; the transition to `superseded` MUST use canonical field names
(`status`, `updated`) and add `superseded_by:`. No snake_case aliases.

**B5. CI 3-tier awareness** — spec-lint, golangci-lint, and Test (per OS) are independent.
Removing `internal/mcp` reduces the test surface; expect Test to stay green. spec-lint runs
against the new SPEC artifacts.

**B6. spec-lint heading convention** — the `OutOfScopeRule` (`internal/spec/lint.go`,
`MissingExclusions`) requires the spec.md body to contain the literal token "out of scope"
AND an H3/H4 sub-heading whose text contains the "out of scope" infix with ≥1 `-` list item
beneath it. An H2 `## Exclusions (What NOT to Build)` alone does NOT satisfy the rule. spec.md
therefore carries `### Out of Scope — Retirement boundary` (H3 + `-` bulleted items) under the
Exclusions section, per the heading convention in `internal/spec/CLAUDE.md`. Verify spec-lint
shows ZERO ERROR for this SPEC via `go run ./cmd/moai spec lint`.

**B8. Working-tree hygiene** — do NOT commit unrelated untracked files. The repo has many
untracked docs-site files at plan time (see git status); the run-phase commit MUST `git add`
only the specific removal/edit paths.

**B10. Untouched-paths PRESERVE** — `internal/lsp/**`, `internal/cli/deps.go`,
`internal/cli/lsp_doctor.go`, the `context7`/`staggeredStartup` template entries,
docs-site/* (parallel-session territory). Do not touch.

**B11. AskUserQuestion prohibition** — manager-develop / manager-spec are subagents; return
a structured blocker report instead of prompting the user.

**Template-specific (CLAUDE.local.md §2 / §15 / §25)**:
- §2 Template-First: edit `.mcp.json.tmpl` source FIRST, then `make build` regenerates
  `embedded.go` (which embeds the whole `templates/` tree via `//go:embed all:templates`).
- §15 16-language neutrality: the settings-management.md mirror edit must not elevate any
  language; removal of `moai-lsp` lines is neutral.
- §25 internal-content isolation: do NOT introduce SPEC IDs / REQ tokens / dates / commit
  SHAs into `internal/template/templates/**`. Remove `moai-lsp` lines cleanly without adding
  internal-tracking tokens.
- Mirror parity: `internal/template/embedded_mirror_test.go` enforces byte-identity between
  `.claude/rules/.../settings-management.md` and its `templates/` mirror — BOTH must be
  edited in the SAME commit, identically.

## §C — Pre-flight Check List (run before any removal)

```bash
# 1. Branch + baseline
git branch --show-current && git rev-parse HEAD

# 2. Confirm the dead-branch isolation invariant (must return exactly 1 line: internal/cli/mcp.go)
grep -rn "moai-adk/internal/mcp" --include="*.go" . | grep -v "_test.go" | grep -v "internal/mcp/"

# 3. Green baseline BEFORE removal (capture for before/after comparison)
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
go test ./internal/mcp/... ./internal/template/... ./internal/cli/... 2>&1 | tail -20

# 4. Enumerate the PRESERVE boundary (must remain after removal)
ls internal/lsp/
grep -rln "moai-adk/internal/lsp" --include="*.go" . | grep -v "_test.go" | grep -v "internal/lsp/"

# 5. Confirm moai-lsp test dependencies (2 functions in settings_test.go)
grep -n "moai-lsp" internal/template/settings_test.go
```

## §D — Constraints (DO NOT VIOLATE)

- PRESERVE (byte-unchanged): `internal/lsp/**`, `internal/cli/deps.go`,
  `internal/cli/lsp_doctor.go`, the `context7` + `staggeredStartup` entries in
  `.mcp.json.tmpl`, all docs-site untracked files, all unrelated SPEC directories.
- Edit `internal/template/templates/.mcp.json.tmpl` source first → `make build` (Template-First).
- Edit BOTH settings-management.md copies (source + template mirror) byte-identically in
  the same commit.
- No `--no-verify`, no `--amend`, no force-push to main.
- Conventional Commits format + `Authored-By-Agent:` trailer + `🗿 MoAI` trailer.
- `git add` only the specific removal/edit paths (B8/B10).
- No AskUserQuestion from subagent — blocker report only.

## §E — Self-Verification Deliverables

manager-develop reports the AC PASS/FAIL matrix (acceptance.md §D), the two build outputs
(host + windows cross-compile), the full-suite test result, the dead-reference grep
(0 matches for `internal/mcp`), the `.mcp.json.tmpl` JSON-validity check on both platform
branches, the mirror-parity test result, and the `internal/lsp` preserve-guard
(git diff empty for `internal/lsp/`).

## Milestones (priority-ordered, no time estimates)

- **M1 — Remove the dead branch (Go).** Delete `internal/mcp/` (all 10 files) and
  `internal/cli/mcp.go`. Run `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...`
  to confirm no dangling import. (Satisfies REQ-001, REQ-002, REQ-008.)
- **M2 — Adjust template tests.** Remove `TestMCPTemplateAlwaysLoadAbsentOnMoaiLSP` and the
  `moai-lsp` block of `TestMCPTemplateExistingFieldsPreserved` in `settings_test.go`. Keep
  the `context7` + `staggeredStartup` assertions intact. (Satisfies REQ-005.)
- **M3 — Edit the template + regenerate.** Remove the `moai-lsp` server entry from
  `.mcp.json.tmpl`, run `make build`, verify rendered JSON valid on both platform branches.
  (Satisfies REQ-003, REQ-004, REQ-009.)
- **M4 — Edit the docs mirror pair.** Remove the two `moai-lsp` lines from BOTH
  settings-management.md copies (source + template mirror), byte-identically. (Satisfies
  REQ-006.)
- **M5 — Supersede the predecessor.** Transition `SPEC-LSPMCP-001` frontmatter
  `status: archived → superseded` + add `superseded_by: SPEC-LSPMCP-RETIRE-001`. (Satisfies
  REQ-007.)
- **M6 — Full verification batch.** Run the AC matrix (acceptance.md §D), the preserve
  guard, and the full test suite. Commit.

## Technical Approach

The whole `internal/mcp` package is removed as a unit because it has a single non-test
consumer (`internal/cli/mcp.go`), which is removed in the same milestone — so there is no
intermediate broken state where a live importer references a deleted package. The template
edit is content-only (a JSON block deletion) and the embed is a directory tree
(`//go:embed all:templates`), so `make build` is sufficient to keep `embedded.go`
consistent. The two settings_test.go test functions are removed (not rewritten) because
they exist solely to assert the now-removed entry's shape; the surviving `context7` /
`staggeredStartup` assertions in the same file are unaffected.

## Risks

- **R1 — Accidental `internal/lsp` damage.** Mitigation: AC-LSPMCP-RETIRE-010 preserve guard
  (`git diff --quiet internal/lsp/`); the package-name similarity (`mcp` vs `lsp`) is the
  primary hazard.
- **R2 — Mirror-parity break.** Editing only one settings-management.md copy fails
  `embedded_mirror_test.go`. Mitigation: edit both in the same commit, byte-identically.
- **R3 — Template JSON malformation.** Removing the `moai-lsp` block could leave a trailing
  comma after the `context7` block. Mitigation: AC-LSPMCP-RETIRE-009 renders + `json.Unmarshal`
  on both platform branches.
- **R4 — Hidden importer surfaces.** If a hidden non-test importer of `internal/mcp` exists
  (pre-flight C2 returns >1 line), STOP and widen scope via blocker report rather than
  proceeding.
