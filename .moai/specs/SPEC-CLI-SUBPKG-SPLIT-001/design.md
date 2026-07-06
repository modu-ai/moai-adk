# Design — SPEC-CLI-SUBPKG-SPLIT-001

Target subpackage layout and the migration strategy for splitting the flat
`internal/cli` package. This document is the HOW; spec.md is the WHAT/WHY.

## §A. Target Layout

The target follows the **existing** `internal/cli/<domain>/` subpackage convention
(`worktree`, `harness`, `preference`, `wizard`, `specid`, `pr`). New subpackages are
domain-named leaf packages that each export a cobra command (or factory) plus receive
dependencies via injected providers.

```
internal/cli/
├── root.go              # STAYS package cli — cobra root + Execute() + all AddCommand wiring
├── deps.go              # STAYS package cli — Composition Root (Dependencies, InitDependencies, providers)
├── version.go clean.go  # STAYS package cli — small root-level commands with no cohesive group
├── uikit/               # NEW leaf package (M-kernel) — shared TUI/settings helpers
│   ├── render.go        #   renderCard, renderKeyValue, renderStatusLine, RenderError (exported)
│   ├── banner.go        #   PrintBanner, PrintWelcomeMessage (exported)
│   ├── settings.go      #   MutateSettingsLocal, WriteFileAtomic (exported)
│   └── schema_bridge.go #   SchemaKeyToTUIField (exported)
├── migrate/             # NEW (M1) — agency migration cluster
├── profile/             # NEW (M2) — profile setup wizard cluster
├── agentlint/           # NEW (M3) — agent_lint + workflow_lint cluster
├── doctor/              # NEW (M-kernel-dependent) — diagnostics cluster
├── update/              # NEW (highest value, conditional) — self-update cluster
├── worktree/ harness/ preference/ wizard/ specid/ pr/   # EXISTING (unchanged)
└── (deferred: launch/, hook/, speccmd/, and 14 tiny single-file clusters — NOT extracted)
```

Naming note: the root-level `harness_*.go` cluster CANNOT take the name `internal/cli/harness`
(already occupied). If ever extracted it must merge INTO the existing `harness` subpackage or
take a distinct name — deferred; see spec.md §E.

## §B. The Extraction Recipe (per kernel-free, deps-free cluster)

Each cluster milestone follows this deterministic 7-step recipe. It is behavior-preserving by
construction (files move; symbols are re-scoped; no logic changes).

1. **Create the leaf package directory** `internal/cli/<group>/` with `package <group>`.
2. **Move the cluster's non-test `.go` files** into it (git mv preserves history). Move all
   platform-tagged siblings together (`*_unix.go`, `*_windows.go`, `*_posix.go`).
3. **Re-scope symbols**: the command variable/factory the cluster registers becomes **exported**
   (`migrateCmd` → `MigrateCmd`, or `newMigrateCmd()` → `NewMigrateCmd()`); any symbol still
   referenced from `package cli` is exported; any symbol used only within the cluster stays
   unexported. Symbols the cluster referenced from `package cli` must resolve to an imported
   package (uikit for helpers) — else the move is blocked (see §D).
4. **Update `root.go` wiring**: replace the local `rootCmd.AddCommand(migrateCmd)` /
   `newXxxCmd()` call with `rootCmd.AddCommand(migrate.MigrateCmd)` + the import. Preserve
   exactly one registration per command (no double-register → cobra panic).
5. **Move the cluster's `_test.go` files** into the new package. White-box tests declared
   `package cli` become `package <group>`; references to symbols still in `package cli` become
   either imported (exported) references or are impossible → the symbol must move too (this is
   the test-migration friction §D quantifies).
6. **Verify**: `go build ./...` exit 0 → `GOOS=windows GOARCH=amd64 go build ./...` exit 0 →
   `go test ./internal/cli/... ./...` exit 0 → `go vet ./...` → `golangci-lint run`.
7. **Commit** the cluster as one atomic behavior-preserving commit (Conventional Commits:
   `refactor(SPEC-CLI-SUBPKG-SPLIT-001): M<N> extract <group> cluster to internal/cli/<group>`).

## §C. deps Provider-Injection Pattern (for deps-coupled clusters)

The 8 deps-coupled files cannot access the `package cli` global `deps` from a subpackage (import
cycle). The proven resolution is the existing worktree pattern — a package-level provider var set
by `root.go`:

```go
// in internal/cli/update/  (the moved cluster)
var UpdateDeps struct {
    Checker      update.Checker
    Orchestrator update.Orchestrator
    EnsureUpdate func() error
}

// in internal/cli/root.go init() or the command's PersistentPreRunE
update.UpdateCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
    if deps == nil { return fmt.Errorf("dependencies not initialized") }
    if err := deps.EnsureUpdate(); err != nil { return err }
    update.UpdateDeps.Checker = deps.UpdateChecker
    update.UpdateDeps.Orchestrator = deps.UpdateOrch
    return nil
}
```

This keeps the Composition Root (`InitDependencies`) in `package cli` — the single place concrete
types are wired — while the subpackage depends only on interfaces + injected values. It matches
`worktree.WorktreeProvider = deps.GitWorktree` verbatim in shape. deps-coupled clusters are
therefore MEDIUM risk (recipe + provider wiring), not just LOW (recipe only).

## §D. Import-Cycle Resolution (the kernel milestone)

Kernel-dependent clusters (doctor, status, statusline, inventory, ...) reference unexported
helpers in `package cli` (`renderCard`, `mutateSettingsLocal`, ...). Because `package cli` imports
the subpackages for registration, those helpers cannot be imported back. Resolution:

1. **Extract the shared kernel to `internal/cli/uikit`** (a leaf package that imports neither
   `package cli` nor any command subpackage). Export the helpers (`RenderCard`, `RenderKeyValue`,
   `RenderStatusLine`, `RenderError`, `PrintBanner`, `MutateSettingsLocal`, `WriteFileAtomic`,
   `SchemaKeyToTUIField`).
2. **Update every current caller in `package cli`** to `uikit.RenderCard(...)` etc. (root-package
   commands that stay keep working through the import).
3. **Only after uikit exists** may kernel-dependent clusters be extracted (they import uikit).

The kernel milestone is a distinct, higher-risk step because the helpers are widely used across
the root package (verified: `render.go` exposes 9 helpers, `banner.go` 8, `settings.go` 3). It is
deliberately sequenced AFTER the kernel-free clusters prove the recipe, and is only undertaken if
a kernel-dependent cluster is judged worth the cost.

## §E. Behavior-Preservation Verification Model

Because this is a refactor of working code (§ research.md D), each milestone's correctness is
established by CHARACTERIZATION, not new tests:

- **Snapshot before**: `go test ./... > before.txt`, `go build` matrix green, `moai --help`
  capture. (The build baseline is already green — exit 0.)
- **Snapshot after**: identical test pass set, identical build matrix, identical `moai --help`
  command list (same subcommands registered, same order groups).
- **No new behavior tests are written** — the existing 54,756 LOC of tests ARE the behavior
  contract. The migration moves them; it does not weaken them. A milestone that cannot keep the
  moved tests green is reverted (recipe step 6 gate).

## §F. Migration Ordering (risk-ascending) — feeds plan.md milestones

| Order | Cluster | Tier of risk | Rationale |
|-------|---------|--------------|-----------|
| M1 | `migrate` | LOW | 1,256 LOC, 9 files, no deps, minimal kernel. Proves the recipe. |
| M2 | `profile` | LOW | 1,511 LOC, self-contained wizard, no deps. |
| M3 | `agentlint` | LOW-MED | 1,389 LOC, no deps, but a single 2,108-LOC white-box test file to migrate. |
| M4 | `constitution` (if kernel-free) OR checkpoint | LOW | 543 LOC. If it needs kernel helpers, defer behind M5. |
| **Checkpoint** | — | — | Re-evaluate marginal value. Recipe proven on ~4,700 LOC. Decide whether to continue. |
| M5 | `uikit` kernel extraction | MED-HIGH | Prerequisite for kernel-dependent clusters. Widely-used helpers. |
| M6 | `doctor` | MED | 2,357 LOC, no deps but heavy kernel use → needs M5 first. |
| M7 | `update` | HIGH | 5,181 LOC + 9,283 test LOC + deps injection. Highest value, highest risk. Behind checkpoint. |
| Deferred | `launch/glm`, `hook`, `speccmd` | HIGH | deps + platform tangle; defer unless specific pain. |
| Never (churn) | 14 tiny single-file clusters | n/a | Splitting ≤370 LOC files is pure churn. |

Each milestone is independently shippable, independently test-verified, and independently
committed (recipe step 7). The SPEC does NOT commit to reaching M7 — the checkpoint after M4
governs whether the higher-risk milestones proceed.

## §G. Anti-Patterns (design-level)

- **AP-1 — Big-bang split**: moving all 26 clusters in one change. Rejected: a single test-compile
  break blocks the entire package's tests; unreviewable diff; violates "surgical changes".
- **AP-2 — Exporting kernel helpers from `package cli` and importing back**: creates an import
  cycle (root imports subpackages). Use the `uikit` leaf package instead (§D).
- **AP-3 — Moving a command without its `_test.go` files**: leaves white-box tests referencing
  moved-away unexported symbols → compile break. Tests move WITH the cluster (recipe step 5).
- **AP-4 — Splitting tiny single-file clusters**: `telemetry.go` (69 LOC) → own package is churn
  with negative value (more files, more import lines, zero navigability gain).
- **AP-5 — Adding behavior tests during the move**: this is a pure refactor; new tests conflate
  refactor risk with feature risk. Characterization only (§E).
- **AP-6 — Splitting platform-tagged siblings across milestones**: `*_unix.go` and `*_windows.go`
  of the same command MUST move together, else the cross-platform build breaks.

## §H. Cross-References

- `internal/cli/CLAUDE.md` — cobra registration convention, cross-platform build-tag rule,
  subagent boundary (C-HRA-008), `syscall.Exec` / tmux-pane launch patterns, settings.go mutation
  helper rule.
- `internal/cli/root.go` — the `AddCommand` wiring surface every milestone touches.
- `internal/cli/deps.go` — the Composition Root; the `EnsureGit` / `EnsureUpdate` lazy-init +
  provider-injection precedent (`worktree.WorktreeProvider`).
- research.md §C — the two dominant risks (import cycle, test surface) this design mitigates.
