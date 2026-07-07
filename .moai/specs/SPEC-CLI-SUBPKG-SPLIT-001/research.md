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
| **profile** | 3 | 1,511 | 8 | 1,112 | — | some | `profile_setup` wizard + translations. **NOT clean** (see §C.4): (a) name collision — `internal/profile` already imported as `profile.` in 7 non-test files (`init.go:22`, `init_layout.go:13`, `launcher.go:14`, `profile.go:10`, `profile_setup.go:14`, `update.go:30`, `web.go:11`); (b) reverse-dep — `schema_bridge.go:24` references the `profileSetupText` type (`profile_setup_translations.go:10`, `getProfileText :604`). |
| **lint (agent/workflow)** | 2 | 1,392 | 2 | 2,285 | — | **none** | **VERIFIED CLEAN** (see §C.4): kernel-render-free per the 14-file survey; zero reverse-dep from `package cli`; single forward dep on 3 `SentinelWorktree*` constants (`sentinels.go:13/17/21`) that are used ONLY by this cluster → co-locatable. `agent_lint.go` (1,157) + `workflow_lint.go` (235) + 2,108-LOC white-box `agent_lint_test.go` + `workflow_lint_test.go` (177). |
| **harness (root-level)** | 5 | 1,306 | 5 | 1,036 | — | some | `harness_route/mute/validate/clusters`. **Collision risk**: `internal/cli/harness` subpackage already exists. |
| **migrate** | 9 | 1,256 | 4 | 1,115 | — | **yes** | **TRI-AXIS COUPLED** (see §C.4 — blocked the original M1): (i) kernel — `migrate_agency.go:634`→`RenderError` (`render.go:105`); (ii) shared — `migrate_restore_skill.go:31/44/80`→`validateSkillID`/`archiveVersion`/`copyDirAll` (`update_archive.go:99/27/193`, shared with update cluster); (iii) reverse — `update.go:1922-1928` constructs `&migrateAgencyRunner{}` writing unexported fields; `update.go:1933-1935` type-asserts `*MigrateError` + `ErrMigrateNoSource`; `update_archive.go` constructs `&MigrateError{}` at 5 sites (`102/109/116/138/151`). Gates on M5 (uikit) + shared-helper co-extraction + `update.go` encapsulation refactor. |
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

Consequence: clusters split into tiers based on verified coupling (see §C.4 for the per-cluster
evidence derived from the orchestrator's 14-file kernel-helper survey):
- **Verified-clean cluster** (agent_lint only) — kernel-render-free AND zero reverse-dep AND
  forward-deps co-locatable. Movable with NO kernel work and NO coupling resolution.
- **Coupled clusters requiring resolution before extraction** (migrate, profile, plus the
  kernel-dependent set). Migrate is tri-axis coupled (kernel + shared-helper + reverse-dep);
  profile has a name collision + `schema_bridge.go` reverse-dep. Both gated on design-time
  resolution work, NOT movable as-is.
- **Kernel-dependent clusters** (doctor, status/statusline, inventory, session, constitution,
  design_folder) — blocked on a prior uikit-extraction milestone (M5).
- **Presumed-kernel-free single-file clusters** (github, astgrep, v2_detection,
  branch_protection, loop, tool_policy, web, telemetry, research) — NOT in the 14-file kernel
  survey, but NOT promoted per spec.md §E (tiny-cluster churn).

This is verifiable per-cluster at extraction time via `go build ./...` failing on an
undefined-symbol reference; it must NOT be assumed clean. The M1-blocker incident (original M1
migrate) proved that front-matter "kernel-free" claims without verbatim file:line verification
are unsound — see §C.4 for the corrected per-cluster coupling map.

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

## §C.4 Per-Cluster Coupling Verification (M1-blocker corrective evidence)

> **Provenance**: the original plan.md listed migrate as "kernel-free, deps-free, minimal kernel
> dependency — Lowest-risk candidate." At run-phase M1 spawn, manager-develop blocked with a
> structured blocker report; orchestrator grep then verified a tri-axis coupling the plan did
> not characterize. This section attaches the verbatim evidence and corrects the cluster map.

### §C.4.1 Kernel-helper survey — 14 files using ≥1 kernel symbol

The orchestrator surveyed all `internal/cli/*.go` non-test files for usage of at least one of
the 9 kernel-helper symbols (`renderCard`, `renderKeyValue`, `renderStatusLine`, `RenderError`,
`PrintBanner`, `printWelcomeMessage`, `mutateSettingsLocal`, `writeFileAtomic`,
`schemaKeyToTUIField`). The 14 files in the kernel-using set:

```
banner.go, hook.go, init.go, init_layout.go, inventory.go, launcher.go,
migrate_agency.go, profile_setup_translations.go, render.go, research.go,
root.go, schema_bridge.go, settings.go, update.go
```

**Evidence**:
```
$ grep -lnE 'renderCard|renderKeyValue|renderStatusLine|RenderError|PrintBanner|printWelcomeMessage|mutateSettingsLocal|writeFileAtomic|schemaKeyToTUIField' internal/cli/*.go | grep -v _test.go | sort
```

Files NOT in the set (kernel-render-free): `agent_lint.go`, `workflow_lint.go`, `design_folder.go`,
`astgrep.go`, `loop.go`, `branch_protection.go`, `telemetry.go`, `tool_policy.go`,
`constitution.go`, `migrate_restore_skill.go`, `migration.go`, plus migrate cluster's other
non-`migrate_agency.go` files. Kernel-render-free is necessary but NOT sufficient for
movability — see §C.4.2..§C.4.4 for the per-cluster coupling that actually governs risk.

### §C.4.2 `agentlint` cluster — VERIFIED CLEAN (the only clean candidate)

- Files: `agent_lint.go` (1,157 LOC) + `workflow_lint.go` (235 LOC) + `agent_lint_test.go`
  (2,108 LOC) + `workflow_lint_test.go` (177 LOC).
- **Kernel-render-free**: neither file appears in the 14-file set above (verified by grep).
- **Zero reverse-dep**: `grep -rE 'AgentLint|WorkflowLint' internal/cli/*.go` excluding `_test.go`
  and the cluster files returns no matches — no `package cli` non-test file references any
  agentlint-exported symbol.
- **Single forward-dep — co-locatable**: the cluster uses 3 sentinel constants
  (`SentinelWorktreeMissing` `sentinels.go:13`, `SentinelWorktreeOnReadonly` `sentinels.go:17`,
  `SentinelWorktreeRequired` `sentinels.go:21`) at `agent_lint.go:689,736` and
  `workflow_lint.go:84,89,100,105,115,120`. A cross-file grep confirms these constants are
  consumed ONLY by agentlint cluster files → move the `const` block with the cluster.
- **Friction**: the 2,108-LOC white-box test. Every unexported reference in
  `agent_lint_test.go` must resolve in the new `internal/cli/agentlint` package.

### §C.4.3 `migrate` cluster — TRI-AXIS COUPLED (blocked the original M1)

- **(i) Kernel axis** — `migrate_agency.go:634` calls `RenderError(err)` (defined `render.go:105`).
  Evidence:
  ```
  internal/cli/migrate_agency.go:634:		fmt.Fprintln(os.Stderr, RenderError(err))
  ```
  Resolution: gate on M5 (uikit extraction).

- **(ii) Shared-helper axis** — `migrate_restore_skill.go` calls 3 symbols co-located in
  `update_archive.go` (shared with the update cluster):
  ```
  internal/cli/migrate_restore_skill.go:31:	if err := validateSkillID(skillID); err != nil {  → update_archive.go:99
  internal/cli/migrate_restore_skill.go:44:	... "skills", archiveVersion, skillID ...         → update_archive.go:27 (const)
  internal/cli/migrate_restore_skill.go:80:	if err := copyDirAll(archiveDir, targetDir); ...  → update_archive.go:193
  ```
  Resolution: co-extract to a neutral helper package (e.g. `internal/cli/archiveutil`) imported
  by both `migrate` and `update`, OR move with one cluster and export-import from the other.

- **(iii) Reverse-dependency axis** — `package cli` constructs migrate-cluster types and reads
  unexported fields, which a subpackage cannot expose:
  ```
  internal/cli/update.go:1922:	r := &migrateAgencyRunner{           ← writes unexported fields
  internal/cli/update.go:1923:		projectRoot: projectRoot,
  internal/cli/update.go:1924:		homeDir:     homeDir,
  internal/cli/update.go:1925:		dryRun:      dryRun,
  internal/cli/update.go:1927:		force: false,
  internal/cli/update.go:1933:	if me, ok := runErr.(*MigrateError); ok && me.Code == ErrMigrateNoSource {
  internal/cli/update_archive.go:102,109,116,138,151:	return &MigrateError{...}   ← 5 construction sites
  ```
  Resolution: introduce `NewMigrateAgencyRunner(opts MigrateAgencyRunnerOpts) *migrateAgencyRunner`
  (or an interface) so `update.go:1922` no longer writes unexported fields; export `MigrateError`,
  its `Code` field, and the `ErrMigrateNoSource` const. This refactor lands in M7 (update cluster
  extraction) OR a prior refactor commit.

### §C.4.4 `profile` cluster — NOT CLEAN (two blocking issues)

- **(a) Name collision with `internal/profile`** — the existing package
  `github.com/modu-ai/moai-adk/internal/profile` is imported as `profile.` in 7 non-test files:
  ```
  internal/cli/init.go:22, init_layout.go:13, launcher.go:14, profile.go:10,
  profile_setup.go:14, update.go:30, web.go:11
  ```
  Evidence:
  ```
  $ grep -ln '"github.com/modu-ai/moai-adk/internal/profile"' internal/cli/*.go | grep -v _test.go
  ```
  Creating `internal/cli/profile` forces import-alias churn across all 7 call sites OR renaming
  the new subpackage (e.g. `internal/cli/profilesetup`).

- **(b) `schema_bridge.go` reverse-dep** — `schema_bridge.go:24` references the `profileSetupText`
  type (a kernel file that stays in `package cli`):
  ```
  internal/cli/schema_bridge.go:24:	var schemaFieldBridge = map[string]func(t profileSetupText) tuiLabel{
  internal/cli/profile_setup_translations.go:10:	type profileSetupText struct {
  internal/cli/profile_setup_translations.go:604:	func getProfileText(lang string) profileSetupText {
  ```
  Moving `profile_setup_translations.go` forces `schema_bridge.go` to import the new subpackage —
  a kernel→subpackage import that risks a cycle. Resolution: lift `profileSetupText` +
  `getProfileText` into a leaf package both import, OR split the bridge table into the new
  subpackage.

### §C.4.5 Why this changes the milestone ordering

The original plan.md listed migrate / profile / agentlint as M1 / M2 / M3 — all "LOW risk,
kernel-free." The §C.4 evidence proves:
- agentlint is the ONLY genuinely-clean cluster;
- migrate is tri-axis coupled and CANNOT ship without M5 (uikit) + shared-helper co-extraction +
  `update.go` encapsulation refactor;
- profile has a name collision and a `schema_bridge.go` reverse-dep that must be resolved first.

Therefore plan.md §F now sequences M1=agentlint (prove the recipe), then a checkpoint, then
conditional milestones (M2 profile, M3 migrate) each behind documented coupling resolution.

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

The measured data supports a **phased, recipe-proven-on-clean-first** extraction that is
explicitly bounded — NOT a full 93-file reorganization:

1. **Prove the recipe on the one verified-clean cluster first (agentlint)** — the M1-blocker
   incident proved that cluster characterization without verbatim file:line verification is
   unsound. Agentlint is the ONLY cluster where the recipe can be proven with zero
   coupling-resolution work (see §C.4.2).
2. **Checkpoint decision after agentlint.** If the recipe felt smooth AND the user accepts the
   coupling-resolution cost, continue; otherwise STOP at M1.
3. **Profile (M2) and migrate (M3) are conditional**, each behind documented coupling resolution
   (§C.4.3, §C.4.4). Migrate additionally gates on M5 (uikit) for its kernel axis + M7 (update)
   for its reverse-dep refactor. **The original M1 blocker proved migrate is NOT low-risk.**
4. **Extract the shared kernel to a leaf package** only when a kernel-dependent cluster is worth
   the cost (a distinct, higher-risk milestone — M5 uikit).
5. **Tackle the highest-value clusters (update, doctor) after the recipe is proven AND the
   post-M1 checkpoint passes**, each behind its conditional gate.
6. **Defer the 14 tiny single-file clusters indefinitely** — moving a 69-370 LOC file is churn.
7. **Defer the deps-tangled + platform-tangled clusters (glm/launch, hook)** unless a specific
   pain point justifies them.

The honest conclusion (spec.md §A): the maintainability gain is real but incremental and
non-user-observable; the test-migration + import-cycle risk is substantial. A big-bang split is
NOT justified. A bounded extraction starting with the one verified-clean cluster (M1 agentlint),
gated by a post-M1 checkpoint, with conditional follow-ups each behind documented coupling
resolution, IS justified.
