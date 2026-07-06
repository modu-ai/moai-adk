# Research — SPEC-CLI-SUBPKG-SPLIT-001

Deep codebase analysis of the flat `internal/cli` package, its cluster structure,
and the coupling that governs the risk of a subpackage split. All quantities below
were measured (not assumed) against the working tree at plan-authoring time via
`find` / `wc` / `grep` / `go build`.

## §A. Baseline Measurements (observed)

| Metric | Value | Command |
|--------|-------|---------|
| Root non-test `.go` files (flat, non-recursive) | **93** | `find internal/cli -maxdepth 1 -name '*.go' ! -name '*_test.go' \| wc -l` |
| Root non-test LOC | **25,838** | `find internal/cli -maxdepth 1 … -exec cat {} + \| wc -l` |
| Recursive non-test files (incl. existing subpackages) | 139 | `find internal/cli -name '*.go' ! -name '*_test.go' \| wc -l` |
| Recursive non-test LOC | 33,321 | recursive `cat \| wc -l` |
| Root **test** files (`_test.go`) | **147** | `find internal/cli -maxdepth 1 -name '*_test.go' \| wc -l` |
| Root **test** LOC | **54,756** | recursive `cat \| wc -l` |
| Root files importing `spf13/cobra` (command-defining) | 53 | `grep -l 'spf13/cobra' … \| wc -l` |
| Root files NOT importing cobra (pure helpers) | 40 | `grep -L 'spf13/cobra' … \| wc -l` |
| Root files coupling the global `deps` variable | **8** | `grep -l '\bdeps\b' … ` minus `deps.go` |
| Cross-platform build-tag files in root | 15 | `grep -l '//go:build' …` |
| Build baseline (green?) | **exit 0** | `go build ./internal/cli/...` |

**Largest single file**: `update.go` = **3,170 LOC** (12% of the root package by itself).

> **Drift re-verification (2026-07-07, pre-audit)**: re-measured against the current working
> tree (5 days after the 2026-07-02 authoring snapshot). Current: 98 root non-test files /
> 26,440 LOC (Δ +5 files / +602 LOC), 153 test files / 56,021 test LOC (Δ +6 files /
> +1,265 LOC), 17 build-tag files (Δ +2), 8 deps-coupled files (unchanged), `update.go`
> 3,173 LOC (Δ +3). All milestone cluster files (M1-M7) remain cohesive in `package cli`;
> 6 existing subpackages intact; kernel helpers still in `package cli` (M5 prerequisite
> unchanged); build baseline still `exit 0`. Growth is contained within existing cluster
> boundaries — does NOT invalidate the cluster map or the milestone ladder. Per-cluster
> counts re-derived at run-phase (not hardcoded from §A).

**Existing subpackages** (the pattern this SPEC follows): `worktree` (22 files / 2,489 LOC),
`preference` (10 / 2,142), `wizard` (7 / 1,485), `harness` (5 / 1,209), `specid` (1 / 60),
`pr` (test-only, `package pr_test`). So the extraction pattern is already established and
proven — this SPEC extends it, it does not invent it.

## §B. Cluster Map (93 root files → cohesive groups)

Files grouped by domain/prefix. `deps✓` = the cluster contains ≥1 file coupling the global
`deps` variable (higher migration risk — needs provider injection). `kernel?` = whether the
cluster likely consumes shared unexported TUI/settings helpers (`renderCard`, `renderKeyValue`,
`renderStatusLine`, `mutateSettingsLocal`, `PrintBanner`) that live in the shared kernel (see §C).

| Cluster | Files | LOC | Test files | Test LOC | `deps✓` | kernel? | Notes |
|---------|------:|----:|-----------:|---------:|:-------:|:-------:|-------|
| **update** | 9 | 5,181 | 21 | 9,283 | ✓ (`update.go`) | some | Self-update engine. Biggest value, biggest test surface. |
| **launch/glm** | 11 | 3,501 | 7 | 4,740 | ✓ (`glm.go`) | some | `cc`/`cg`/`glm` launchers + `team_spawn` + platform exec split (`syscall.Exec`). Platform-tangled. |
| **doctor** | 9 | 2,357 | 11 | 2,706 | — | **yes** | Diagnostics (`doctor.go` + `doctor_*` + `lsp_doctor`). Heavy `renderCard`/`renderStatusLine` use. |
| **hook** | 3 | 1,598 | 13 | 3,290 | ✓ (`hook.go`, `hook_pre_push.go`) | some | Hook command surface. deps-coupled. |
| **profile** | 3 | 1,511 | 8 | 1,112 | — | some | `profile_setup` wizard + translations. Self-contained. |
| **lint (agent/workflow)** | 2 | 1,389 | 1 | 2,108 | — | some | `agent_lint.go` (1,154) + `workflow_lint.go`. One 2,108-LOC white-box test file. |
| **harness (root-level)** | 5 | 1,306 | 5 | 1,036 | — | some | `harness_route/mute/validate/clusters`. **Collision risk**: `internal/cli/harness` subpackage already exists. |
| **migrate** | 9 | 1,256 | 4 | 1,115 | — | minimal | agency migration (`migrate_agency*` + `migration.go`). No deps, no kernel. **Lowest-risk candidate.** |
| **spec (cmd)** | 7 | 1,192 | 6 | 1,545 | ✓ (`spec_lint.go`) | some | `spec_status/close/audit/view/lint/drift`. deps-coupled via lint. |
| **session/state** | 2 | 649 | 3 | 517 | — | some | `session.go` + `state.go`. |
| **init** | 2 | 559 | — | — | — | some | `init.go` + `init_layout.go`. |
| **constitution** | 1 | 543 | 3 | 460 | — | some | `constitution.go`. |
| **design_folder** | 1 | 370 | — | — | — | some | `design_folder.go`. |
| **inventory** | 1 | 332 | — | — | ✓ | yes | `inventory.go` composes sessions/worktrees/harnesses view. deps-coupled. |
| **tool_policy** | 1 | 310 | — | — | — | some | `tool_policy.go`. |
| **astgrep** | 1 | 284 | — | — | — | — | `astgrep.go`. |
| **loop** | 1 | 257 | — | — | ✓ | — | `loop.go`. deps-coupled. |
| **github/pr** | 2 | 358 | — | — | — | — | `github.go` + `pr_watch_cmd.go`. |
| **v2_detection** | 1 | 209 | — | — | — | — | `v2_detection.go`. |
| **branch_protection** | 1 | 194 | — | — | — | — | `branch_protection.go`. |
| **statusline** | 1 | 159 | — | — | — | some | `statusline.go`. |
| **research (cmd)** | 1 | 150 | — | — | — | — | `research.go`. |
| **web** | 1 | 74 | — | — | — | — | `web.go`. |
| **telemetry** | 1 | 69 | — | — | — | — | `telemetry.go`. |
| **SHARED KERNEL (stays in `package cli`)** | 14 | ~1,692 | — | — | ✓ (`deps.go`, `root.go`) | n/a | `root.go`, `deps.go`, `banner.go`, `render.go`, `help.go`, `version.go`, `status.go`, `clean.go`, `settings.go`, `schema_bridge.go`, `sentinels.go`, `homedir.go`, `console_{windows,others}.go` |

> `worktree_validation.go` (108 LOC) belongs to the already-extracted `worktree` subpackage
> domain and is accounted separately from the root-cluster total. Cluster LOC sum reconciles
> to ~25,730; the ~108 residual is this file. Exact per-cluster counts are re-derived at
> run-phase, not hardcoded from this table.

### Cluster count summary

- **26 candidate clusters** + 1 shared kernel.
- **14 clusters are single-file** (≤ ~370 LOC each): design_folder, inventory, tool_policy,
  astgrep, loop, github/pr(2), v2_detection, branch_protection, statusline, research, web,
  telemetry, constitution. Splitting a 69-LOC file (`telemetry.go`) into its own package is
  **pure churn with near-zero maintainability value** — these are explicitly NOT recommended
  for extraction (see spec.md §E Out of Scope).
- **5 clusters carry the real value** (each > 1,300 LOC AND multi-file): update, launch/glm,
  doctor, hook, profile, lint, harness. These are where a split materially improves navigability.

## §C. The Two Dominant Risks (the crux of the value/risk assessment)

### C.1 Import-cycle hazard (architectural, hard blocker)

`root.go` (`package cli`) imports the subpackages to register their commands:

```go
import "github.com/modu-ai/moai-adk/internal/cli/worktree"
rootCmd.AddCommand(worktree.WorktreeCmd)      // root → subpackage
```

Therefore a subpackage **cannot import `package cli` back** — that is a compile-time import
cycle. But the shared unexported helpers (`renderCard`, `renderKeyValue`, `renderStatusLine`,
`mutateSettingsLocal`, `PrintBanner`, `schemaKeyToTUIField`) live in `package cli`
(`render.go`, `banner.go`, `settings.go`, `schema_bridge.go`). Any cluster that consumes them
(doctor, status, statusline, inventory, and likely others) **cannot** be moved to a subpackage
until those helpers are relocated to a **neutral leaf package** (e.g. `internal/cli/uikit`)
that both `package cli` and the new subpackages import.

Consequence: clusters split cleanly into two tiers:
- **Kernel-free clusters** (migrate, agent_lint, profile, github, astgrep, v2_detection,
  branch_protection, loop, tool_policy, web, telemetry, research) — movable with NO kernel work.
- **Kernel-dependent clusters** (doctor, status/statusline, inventory, session, constitution,
  design_folder) — blocked on a prior kernel-extraction milestone.

This is verifiable per-cluster at extraction time via `go build ./...` failing on an
undefined-symbol reference; it must NOT be assumed clean.

### C.2 Test-migration surface (dominant regression risk)

The root package carries **147 test files / 54,756 test LOC** — more than 2× the non-test LOC.
Go same-package (white-box) tests declared `package cli` freely access unexported symbols. When
a command file moves to `package cli/<group>`:

1. Its `_test.go` files must move with it, AND
2. Any test that referenced a now-unexported helper still in `package cli` (or vice versa)
   breaks compilation, AND
3. `internal/` test helpers shared across files may need to be exported or duplicated.

The **update cluster alone** carries **21 test files / 9,283 test LOC**. Moving `update.go`
means relocating ~9,283 LOC of white-box tests and re-resolving every unexported reference. This
is the single largest reason a big-bang split is high-risk: a compile break in the moved test
set is not a localized failure — it blocks the whole package's test compilation.

### C.3 Secondary risks

- **Global `deps` coupling (8 files)**: `update.go`, `glm.go`, `hook.go`, `hook_pre_push.go`,
  `inventory.go`, `loop.go`, `spec_lint.go`, `root.go`. A moved deps-coupled command loses
  access to the package-level `deps`. `GetDeps()` exists but is in `package cli` (cycle again).
  The proven resolution is the **provider-injection pattern** already used for worktree
  (`worktree.WorktreeProvider = deps.GitWorktree`, set in `root.go` `init()` /
  `PersistentPreRunE`). Each deps-coupled cluster needs an equivalent injected provider.
- **Cross-platform build tags (15 files)**: platform pairs (`*_unix.go` / `*_windows.go` /
  `*_posix.go`) MUST move together, and `GOOS=windows GOARCH=amd64 go build ./...` MUST pass
  after each move (`internal/cli/CLAUDE.md` cross-platform convention).
- **Cobra double-registration panic**: `internal/cli/CLAUDE.md` warns that two subcommands with
  the same `Use:` prefix panic at runtime. Moving a command must preserve exactly one
  `rootCmd.AddCommand` registration.
- **Root-level `harness` cluster vs existing `internal/cli/harness` subpackage**: naming
  collision — the 5 root `harness_*.go` files cannot move to `internal/cli/harness` without
  reconciling with the 5 files already there. Higher-friction than a greenfield subpackage.
- **`team_spawn_lock_test_unix.go`** is a non-test build-tagged helper (name contains "test"
  but does not end `_test.go`) — an edge case that must move with the launch cluster.

## §D. Behavior-Preservation Framing (this is a refactor of WORKING code)

`go build ./internal/cli/...` exits 0 today; the CLI works. This SPEC is a **pure structural
refactor** — the observable behavior of every `moai` subcommand MUST be identical before and
after. There is NO functional change, NO new feature, NO bug fix. The correct verification
model is therefore **characterization**: the full test suite (`go test ./...`) + cross-platform
build + `moai --help` output are the behavior snapshot that each milestone must preserve.

Because the code already works, the split earns its risk ONLY through maintainability gains
(navigability, compile-time boundaries, per-cluster test isolation, fewer merge conflicts) —
none of which are user-observable. This is the tension the plan-auditor and user must weigh
(see spec.md §A VALUE justification and §E Out of Scope).

## §E. Replacement / Continuity Notes

- No public API of `internal/cli` changes for external callers — `cmd/moai/main.go` calls
  `cli.Execute()`, which stays in `package cli`. Verified: `Execute` is defined in `root.go`
  and is the only entry point (`fan_in=3`: main.go, root_test.go, integration_test.go per its
  `@MX:ANCHOR`).
- The composition root (`InitDependencies` in `deps.go`) stays in `package cli` — it is the one
  place concrete types are wired. Subpackages receive dependencies via injected providers, never
  by importing the composition root.

## §F. Recommendation Basis (feeds spec.md / plan.md)

The measured data supports a **phased, lowest-risk-first** extraction that is explicitly bounded
— NOT a full 93-file reorganization:

1. **Prove the recipe on kernel-free, deps-free clusters first** (migrate → profile → lint →
   constitution-if-kernel-free). Zero import-cycle work, moderate LOC, self-contained tests.
2. **Extract the shared kernel to a leaf package** only when a kernel-dependent cluster is worth
   the cost (a distinct, higher-risk milestone).
3. **Tackle the highest-value clusters (update, doctor) after the recipe is proven**, each behind
   a re-evaluation checkpoint.
4. **Defer the 14 tiny single-file clusters indefinitely** — moving a 69-370 LOC file is churn.
5. **Defer the deps-tangled + platform-tangled clusters (glm/launch, hook)** unless a specific
   pain point justifies them.

The honest conclusion (spec.md §A): the maintainability gain is real but incremental and
non-user-observable; the test-migration + import-cycle risk is substantial. A big-bang split is
NOT justified. A bounded phased extraction of the 3-4 highest-value low-risk clusters IS
justified, with an explicit stop-when-marginal-value-drops checkpoint.
