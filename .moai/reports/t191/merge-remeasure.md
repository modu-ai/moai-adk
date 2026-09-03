# t191 — merge-tree re-measurement (lane-4, integration window)

## Claim

The t191 branch, after absorbing local `develop`, builds clean and its touched
packages pass, with two inherited reds that are byte-identically develop's.

## Evidence

- absorb merge commit: `d77adf6b2` (`Merge branch 'develop' into WT-project-continuation`)
  - base: `WT-project-continuation` @ `55885aae3` (16 commits ahead of develop)
  - absorbed: local `develop` @ `a239cf050`
  - conflicts: 1 file — `CHANGELOG.md`, `## [Unreleased] / ### Added` head insertion
    on both sides. Resolved by keeping BOTH sides, t191's entry on top (last landing).
    Auto-merged without conflict: `internal/config/defaults.go`, `internal/config/types.go`.

- `go build ./...` → rc=0 (`remeasure-build.log`, empty)
- `go vet ./internal/cli/wizard/... ./internal/config/... ./internal/core/project/...
   ./internal/web/... ./internal/settings/... ./internal/cli/` → rc=0 (`remeasure-vet.log`, empty)
- `go test ./internal/cli/wizard/... ./internal/config/... ./internal/core/project/...
   ./internal/web/... ./internal/settings/...` → rc=1 (`remeasure-scoped.log`)
  - ok: `internal/cli/wizard`, `internal/config/atomicfile`, `internal/config/toolpolicy`,
    `internal/core/project`, `internal/web`, `internal/settings`,
    `internal/settings/agentfm`, `internal/settings/yamlpatch`
  - FAIL: `internal/config` — `TestAlwaysLoadedTokenBudget` only
- `go test ./internal/cli/` → rc=1, 602.377s (`remeasure-cli.log`)
  - FAIL: `TestRunDoctor_WithExport`, `TestRunDoctor_WithFix`, `TestRunDoctor_Verbose`,
    `TestRunDoctor_AllFlags`, `TestRunDoctor_VerboseAndDetail`, `TestRunDoctor_ExportMode`,
    `TestDoctorCmd_Execution` — all one cause: `doctor: 1 check(s) failed`,
    `✗ Agent Emit Embed — moai embeds stale agent-emit artifacts (11/11 compared):
    manager-docs.toml, manager-lead.toml`

## Baseline-attribution

Both reds are develop's, established by measurement rather than inference:

- `TestAlwaysLoadedTokenBudget` (76,939 tokens vs budget 76,400, overflow 539).
  The measured surface is a pure function of `.claude/rules/moai/**` (no-`paths:`
  files), `CLAUDE.md`, `AGENTS.md`, and `.claude/output-styles/moai/moai.md`
  (`internal/config/token_budget_guard.go:220` `alwaysLoadedSurface`).
  `git diff --name-only develop HEAD -- .claude/rules/moai CLAUDE.md AGENTS.md
  .claude/output-styles/moai` → EMPTY: the merge tree's surface is byte-identical to
  develop's, so develop alone measures the same 76,939 and fails identically.
  Pending fix is the budget raise to 77,200 carried by card t453 (window-pending).

- `Agent Emit Embed`. `git diff --name-only develop HEAD -- .claude/agents
  internal/template/templates/.claude/agents internal/template/templates/.codex`
  → EMPTY. t191's own 39-file change set (`git diff --name-only develop...55885aae3`)
  contains 0 agent files. Same class as the drift the dispatch pre-declared
  (`4244c4a06` 소관); the observed stale files here are `manager-docs.toml` and
  `manager-lead.toml`, not `sync-auditor.toml`.

## Gaps

- Full suite NOT run locally (repo discipline: CI owns the full-suite verdict).
  Packages measured: the 6 the change set touches.
- `make build` / `make agents-emit` NOT run — running the emitter would repair
  develop's inherited drift inside this card, which is out of scope.
- No cross-platform (darwin/windows) measurement; CI owns that axis.
- CodeRabbit / PR-level review not applicable — this card lands via local develop merge.

## Residual-risk

- The two inherited reds travel into develop unchanged by this merge. They were
  already there; this card neither repairs nor worsens them.
- `internal/cli` took 602s under a loaded machine; a re-run under different load
  could surface timing-sensitive tests not seen here.
