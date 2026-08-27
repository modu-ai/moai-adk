# Progress — SPEC-CLI-TEST-CWD-ISOLATION-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-08-28
- Artifacts: spec.md / plan.md / acceptance.md (**Tier M, 3 artifacts** — rationale in
  spec.md §C: measured scope is S-shaped; M stands on the canonical 3-file artifact set,
  the iter-1-FAIL retry-slot reality, and the guard+per-test verification ceremony;
  downgrade path documented for independent audit judgment) + this progress skeleton.
- Budget: REQ 5 / AC 5 — Tier M ceilings (16 / 16) respected.
- Evidence base: **measured RED reproducer** (lead session, this worktree, base `d34a789a4`,
  2026-08-28; probe record `.moai/reports/t334/red-probe.md`): env-scrubbed
  `go test ./internal/cli -run 'Kanban|Factory' -count=1` → rc=0, 0.869 s, recreates
  `state/todo/leads.json` + `state/factory/workers.json`. Negative probes: internal/config,
  internal/kanban, `Config|Cache|Settings` — all clean (config-cache reclassified
  historical; internal/hook not probed, no claim). Base-tree fact: 4 committed `.moai` dirs
  on `d34a789a4` → tree judgments are baseline deltas, never emptiness. Plus t317 §G-1
  (`0ad4b52ba`) and primary-checkout residue re-verified 2026-08-28.
- **plan-audit iter1 (2026-08-28): FAIL 0.775 vs Tier M 0.80** — report
  `.moai/reports/t334/plan-audit-iter1.md`. v0.2.1 revision = D1 (baseline-delta scan) +
  D2 (tier M restored, acceptance.md retained) + D3 (attribution aligned, red-probe.md
  cited) + D4 (internal/hook struck) + D5/D7 (wording). iter2 delta re-audit pending.

## §E.2 Run-phase Evidence

### M1 — RED re-established (frozen reproducer, attribution triple)

- **Tree**: HEAD `9114ddbc3` (branch `WT-cli-test-cwd`, base `origin/develop` `d34a789a4`), worktree
  `.claude/worktrees/t334`, run-phase session, 2026-08-28.
- **Pre-run state**: `internal/cli/.moai` absent at pre-flight (plan-phase probe residue already
  gone); baseline scan `/tmp/t334-moai-pre.txt` = exactly the 4 committed dirs
  (`internal/harness/router/testdata/{keyword-force,normal,spec-overrides}/.moai` +
  `internal/template/templates/.moai`).
- **Command** (frozen reproducer, verbatim incl. env scrub):

  ```
  unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_PROJECT_DIR && go test ./internal/cli -run 'Kanban|Factory' -count=1
  ```

- **Verbatim output**: `ok  \tgithub.com/modu-ai/moai-adk/internal/cli\t0.913s` (rc=0)
- **RED judgment (post-run)**:
  - `find internal/cli/.moai -type f` → `internal/cli/.moai/state/todo/leads.json` +
    `internal/cli/.moai/state/factory/workers.json` (≥1 file; name-agnostic AC-001 satisfied)
  - baseline delta: `diff /tmp/t334-moai-pre.txt /tmp/t334-moai-post.txt` → `0a1 >
    ./internal/cli/.moai` (exit 1 = delta dirty)
- **RED for the right reason**: residue created by THIS run from the recorded clean baseline
  (4-line scan re-verified before the run); post-run `rm -rf internal/cli/.moai` restored the
  scan to baseline equality (diff exit 0).
- **Pre-flight companions**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...`
  exit 0; `golangci-lint run --timeout=2m` → `0 issues.` (lint baseline).

### M1 — per-test identification (REQ-4; selector bisect, each run env-scrubbed, `-count=1`)

`go test ./internal/cli -list 'Kanban|Factory'` → 55 tests. Every test/group was run with
`internal/cli/.moai` removed immediately before; producing status judged by
`find internal/cli/.moai -type f` after each run. Result — **5 producing tests**, all others
clean (union = exactly the reproducer's file set):

| # | Test | File:line | Residue produced | Root resolution hit |
|---|------|-----------|------------------|---------------------|
| 1 | `TestCC_FactoryEntryThroughRunCC` | `internal/cli/factory_test.go:747` | `state/todo/leads.json` + `state/factory/workers.json` | `launchProjectRoot()` → `resolveProjectDir()` → cwd (P2+P3) |
| 2 | `TestCC_KanbanEnvMutationIsRestored` | `internal/cli/cc_test.go:513` | `state/todo/leads.json` | same (P3) |
| 3 | `TestCC_KanbanFlagStrippedBeforeLaunch` | `internal/cli/cc_test.go:369` | `state/todo/leads.json` | same (P3) |
| 4 | `TestGLM_FactoryWorkerEntry` | `internal/cli/factory_test.go:829` | `state/factory/workers.json` | same (P2) |
| 5 | `TestGLM_KanbanFlagParity` | `internal/cli/glm_test.go:543` | `state/todo/leads.json` | same (P3) |

- **Mechanism note (why the existing seam does not catch them)**: all 5 already stub
  `findProjectRootFn` → `t.TempDir()` (`installFactoryLaunchSeam` / inline), but the
  name-claim call sites (`appendLeadName`, `resolveFactoryWorkerName`,
  `resolveCompanionName`) resolve their root via `launchProjectRoot()` → `resolveProjectDir()`
  (`internal/cli/session.go:264`) = `$CLAUDE_PROJECT_DIR` or `os.Getwd()` — a DIFFERENT
  resolver that bypasses the stubbed seam. Under the scrubbed env, root = the package cwd
  (`internal/cli`), so the registries land in-tree.
- **Verified clean** (no residue individually or in group): `TestCC_KanbanWritesNoStateRecord`,
  `TestLauncherWritesNoKanbanRecord`, `TestGLMTaskFactoryModeIgnoresModelOverride`, all 11
  `TestEnter*`, the 11-test cobra-family group (`TestCG_*`, `TestReject*`,
  `TestACFM022a/023c_*`, `TestNewMCPCmd_FactoryChildren`, `TestFactoryGenealogyInHelp`,
  `TestMoAIHuhTheme_*`, `TestRunLinkSpec_FactoryError`), and the 24-test pure-logic group
  (`TestParse*`, `TestResolve*`, `TestPrepareKanbanSettings*`, tier/blockcap/constant tests).
- **Frozen per-test selector list (D7 freeze)** — REQ-4 is judged by running each of these
  individually post-fix:
  `TestCC_FactoryEntryThroughRunCC`, `TestCC_KanbanEnvMutationIsRestored`,
  `TestCC_KanbanFlagStrippedBeforeLaunch`, `TestGLM_FactoryWorkerEntry`,
  `TestGLM_KanbanFlagParity`.

### M2 — residue guard implemented + guard RED observed (AC-002)

- **Change**: `internal/cli/main_test.go` TestMain extended with the residue guard
  (pre-`m.Run()` existence snapshot of the entry-cwd `.moai` path, post-`m.Run()` delta check,
  stderr failure naming the detected absolute path, exit-code override to 1). Guard judgment is
  directory-existence (never a file list — the `kanban/`→`todo/` drift); locus pinned from the
  entry cwd so a mid-run chdir cannot move it; delta semantics (pre-existing residue is not this
  run's delta). `go vet ./internal/cli/` exit 0; `gofmt -l internal/cli/main_test.go` empty.
- **Command** (guard RED observation; unfixed tree, same env scrub; the TestMain-anchored guard
  rides every selector, so the frozen selector IS the `<guard>`-augmented form):

  ```
  unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_PROJECT_DIR && go test ./internal/cli -run 'Kanban|Factory' -count=1
  ```

- **Verbatim output** (exit code 1):

  ```
  PASS
  RESIDUE GUARD FAIL: this test run created /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t334/internal/cli/.moai — internal/cli tests must not write .moai state into the package working directory (SPEC-CLI-TEST-CWD-ISOLATION-001 REQ-1/REQ-2/REQ-3). Isolate the producing test's project root (see the SPEC's mechanism ladder), then remove the directory so the guard re-arms for the next run.
  FAIL	github.com/modu-ai/moai-adk/internal/cli	0.895s
  FAIL
  ```

- **Mutant-probe counter-evidence (§D.2)**: the inner tests print `PASS` — the FAIL is the
  guard's alone, and the message names the actually-detected residue path (not a path that can
  never exist), so the guard is not vacuously passing.

### Run-phase context note (user-directed, outside SPEC AC scope)

2026-08-28, mid-M1: the operator instructed immediate removal of ALL residue `.moai`
directories ("더이상 묻지말고 모두 .moai만 제거해라"). Removed: worktree `internal/cli/.moai`
(bisect residue) AND the primary checkout's stale `internal/cli/.moai` (the §D.6
forward-looking item — `state/config-cache.json`, `state/kanban/leads.json`,
`state/factory/workers.json`, Aug-13/20 file set), plus 40+ additional untracked stray
`.moai` directories discovered by a primary-wide scan (`internal/{permission,statusline,web}`,
`docs-site/**`, `.github*`, `.moai/specs/*` nested strays, `.omp`, etc. — all verified
untracked: `git ls-files | grep '/\.moai/'` returns content only under the 4 committed
roots). Post-cleanup scan: exactly the 4 tracked `.moai` roots remain outside backup
archives (`.moai-backups/`, `.moai/backups/`, `docs-site.hextra.backup/` deliberately
preserved — archived snapshots, not live residue). No tracked file touched; no source
change; does not affect the worktree baseline. Code-level isolation for the non-cli
packages (`internal/{permission,statusline,web}` residue regeneration) remains out of
scope per spec.md §B.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs populates; carries sync_commit_sha on close>_
