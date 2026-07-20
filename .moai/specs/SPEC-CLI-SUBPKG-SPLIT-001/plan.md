# Plan — SPEC-CLI-SUBPKG-SPLIT-001

Implementation plan for the phased `internal/cli` subpackage split. Milestones are
risk-ascending; each is one cluster extraction with its own behavior-preservation gate.

## §A. Context

- **Location**: `/Users/goos/MoAI/moai-adk-go/internal/cli` (project root).
- **Baseline (observed)**: `go build ./internal/cli/...` exit 0; 93 root non-test files /
  25,838 LOC; 147 test files / 54,756 test LOC; 8 files couple global `deps`; 15 build-tag files.
- **Artifacts**: `.moai/specs/SPEC-CLI-SUBPKG-SPLIT-001/{spec,plan,acceptance,design,research,progress}.md`
  (Tier L 5-file set + progress).
- **Existing pattern to follow**: `internal/cli/{worktree,harness,preference,wizard,specid,pr}`
  (subpackage exports a cobra command; deps via injected provider — `worktree.WorktreeProvider`).
- **cycle_type**: `ddd` (existing working code, characterization-preservation refactor — behavior
  must be preserved, no new behavior; this maps to ANALYZE-PRESERVE-IMPROVE, not RED-GREEN-REFACTOR).

## §B. Known Issues (auto-injected, filtered to relevant)

- **B1 Cross-platform build tags** [RELEVANT]: 15 build-tag files in root. Platform siblings
  (`*_unix.go`/`*_windows.go`/`*_posix.go`) MUST move together. `GOOS=windows GOARCH=amd64 go build ./...`
  MUST pass after each milestone.
- **B3 Subagent boundary (C-HRA-008)** [RELEVANT]: `internal/cli` code must not call
  `AskUserQuestion`/`mcp__askuser`. The move must not introduce any; grep gate after each milestone.
- **B4 Frontmatter canonical schema** [APPLIED]: 12 canonical fields + `tier: L` + `era: V3R6`.
- **B5 CI 3-tier** [RELEVANT]: spec-lint + golangci-lint + Test (per OS) each fail independently.
- **B8/B10 Working-tree hygiene / PRESERVE** [RELEVANT]: do not touch runtime-managed files
  (`.moai/state/`, `.moai/cache/`, `.moai/logs/`) or unrelated SPEC dirs; `git add` specific paths.
- **B9 Commit + push (Hybrid Trunk)** [RELEVANT]: Tier L → Route B (PR route per `--pr`/Tier L) is
  the default for Tier L, but a maintainability refactor with per-milestone commits MAY use main-direct
  per user choice at Implementation Kickoff Approval. Conventional Commits; `--no-verify` prohibited.
- **B11 AskUserQuestion prohibition** [RELEVANT]: subagent returns blocker reports, never prompts.
- **Import-cycle hazard (SPEC-specific)** [CRITICAL]: subpackages cannot import `package cli`
  (root imports them for `AddCommand`). Kernel-dependent clusters blocked on `uikit` extraction.

## §C. Pre-flight (before M1)

```bash
git branch --show-current && git rev-parse HEAD
go build ./...                              # expect exit 0 (baseline confirmed)
GOOS=windows GOARCH=amd64 go build ./...    # expect exit 0
go test ./internal/cli/... > /tmp/cli-before.txt 2>&1; tail -3 /tmp/cli-before.txt
golangci-lint run --timeout=2m 2>&1 | tail -5   # capture lint baseline (NEW vs pre-existing)
go run ./cmd/moai --help > /tmp/help-before.txt  # behavior snapshot: subcommand list
```

## §D. Constraints (DO NOT VIOLATE)

- PRESERVE: existing subpackages (`worktree/harness/preference/wizard/specid/pr`) unchanged;
  `cli.Execute()` + `deps.go` Composition Root stay in `package cli`.
- No functional change (REQ-CSS-011); no test deletion/skip (REQ-CSS-012); no new behavior tests.
- One cluster per milestone (REQ-CSS-009); atomic behavior-preserving commit per milestone.
- Platform siblings move together (B1); grep-verify no AskUserQuestion (B3).
- Do not extract the 14 tiny single-file clusters or the deps/platform-tangled clusters (spec §E).

## §E. Self-Verification (per milestone) — see acceptance.md for full matrix

Each milestone reports: AC PASS/FAIL matrix, `go build` matrix result, `go test ./...` result,
`moai --help` diff (expect empty), lint status (NEW vs baseline), commit SHA.

## §F. Milestones (recipe-proven-on-clean before any coupling-resolution work)

> Milestones are gated by a re-evaluation **checkpoint after M1 (agentlint)** — the only
> orchestrator-verified clean cluster. The SPEC does NOT commit to reaching M7 — see REQ-CSS-010
> and §F.9. M2 (profile) and M3 (migrate) are **conditional** (coupling-resolution prerequisites
> documented per-axis below); M5-M7 remain conditional (kernel/deps work).
>
> **Provenance**: the original M1 (migrate) was blocked by manager-develop with a structured
> blocker report — orchestrator grep verified a tri-axis coupling the original plan did not
> characterize (see M3 for verbatim file:line evidence). M1-M7 below reflect the corrected
> ordering: agentlint-as-M1 is the only cluster where the extraction recipe can be proven
> without first undertaking coupling-resolution work.

### M1 — Extract `agentlint` cluster → `internal/cli/agentlint` (LOW risk, proves the recipe)

- Move `agent_lint.go` (1,157 LOC) + `workflow_lint.go` (235 LOC) + the 2,108-LOC white-box
  `agent_lint_test.go` + `workflow_lint_test.go` (177 LOC). The single 2,108-LOC white-box test
  is the friction — verify every unexported reference resolves in the new package.
- **Verified-clean status** (orchestrator-confirmed, NOT assumed — see research.md §C.1 for the
  14-file kernel survey evidence):
  - **Kernel-render-free**: neither `agent_lint.go` nor `workflow_lint.go` appears in the 14-file
    kernel-using set (`banner.go, hook.go, init.go, init_layout.go, inventory.go, launcher.go,
    migrate_agency.go, profile_setup_translations.go, render.go, research.go, root.go,
    schema_bridge.go, settings.go, update.go`).
  - **Zero reverse-dependency**: no `package cli` non-test file references agentlint-exported
    symbols (`grep -rE 'AgentLint|WorkflowLint' internal/cli/*.go` excluding `_test.go` and the
    cluster files themselves returns no matches). `root.go` registers the cobra commands via
    `AddCommand` but does not otherwise touch cluster internals.
  - **Single forward-symbol-dep — co-locatable**: the cluster uses 3 sentinel string constants
    (`SentinelWorktreeMissing` / `SentinelWorktreeOnReadonly` / `SentinelWorktreeRequired`,
    defined `sentinels.go:13/17/21`) at `agent_lint.go:689,736` and `workflow_lint.go:84,89,100,
    105,115,120`. A grep across all non-test `package cli` files confirms these constants are
    consumed ONLY by agentlint cluster files → they co-locate with the cluster (move the `const`
    block into `internal/cli/agentlint`, or inline the 3 string literals). **No cycle.**
- Gate: build matrix + `go test ./...` + `moai --help` diff empty.

### CHECKPOINT (after M1) — re-evaluate before any coupling-resolution milestone (REQ-CSS-010)

- Recipe proven on a genuinely-clean cluster (~1,392 LOC + ~2,285 test LOC) with zero
  coupling-resolution work. This is the ONLY orchestrator-verified clean cluster; every
  subsequent milestone requires design-time coupling resolution before extraction can begin.
- Assess: was the navigability gain on agentlint worth the churn? Are the remaining conditional
  milestones (M2 profile name-collision, M3 migrate tri-axis, M5 uikit kernel, M6 doctor, M7
  update) worth their coupling-resolution cost? Decide: STOP (ship M1 only) OR continue into
  conditional milestones.
- Log the decision to `progress.md § Mode Selection` / checkpoint note.

### M2 — Extract `profile` cluster → `internal/cli/profile` (LOW-MED risk; CONDITIONAL)

- Move 3 files (`profile_setup.go`, `profile_setup_translations.go`, `profile.go`) + 8 test files
  (~1,112 test LOC). **Two blocking issues** (characterized at plan-phase, NOT assumed):
  1. **Name collision with `internal/profile`**: the existing package
     `github.com/modu-ai/moai-adk/internal/profile` is imported as `profile.` in 7 non-test
     files — `init.go:22`, `init_layout.go:13`, `launcher.go:14`, `profile.go:10`,
     `profile_setup.go:14`, `update.go:30`, `web.go:11`. Creating `internal/cli/profile` forces
     import-alias churn across all 7 call sites (alias one of the two imports) OR renaming the
     new subpackage (e.g. `internal/cli/profilesetup`).
  2. **`schema_bridge.go` reverse-dep**: `schema_bridge.go:24` declares
     `var schemaFieldBridge = map[string]func(t profileSetupText) tuiLabel{...}`, referencing the
     `profileSetupText` type (defined `profile_setup_translations.go:10`, returned by
     `getProfileText` at `:604`). Moving `profile_setup_translations.go` into the subpackage
     forces `schema_bridge.go` (a kernel file that stays in `package cli`) to import the new
     subpackage — a kernel→subpackage import that risks a cycle.
- **Resolution prerequisite**: name-collision analysis (alias vs rename) + `schema_bridge.go`
  coupling resolution (either lift `profileSetupText` + `getProfileText` into a leaf package both
  import, OR split the bridge table into the new subpackage). Marked CONDITIONAL — skip if either
  blocker is uneconomic.
- Gate as M1, after the blocking issues are resolved at design time.

### M3 — Extract `migrate` cluster → `internal/cli/migrate` (MED-HIGH risk; CONDITIONAL, post-uikit)

- Move 9 files (`migrate_agency*.go`, `migration.go`, `migrate_restore_skill.go` + platform
  siblings) + 4 test files (~1,115 test LOC). **Tri-axis coupling** (orchestrator-verified at
  the original M1 spawn via grep, NOT assumed — this is the blocker that demoted M1→M3):
  - **(i) Kernel axis** — `migrate_agency.go:634` calls `RenderError(err)` (defined `render.go:105`,
    a uikit/M5-kernel candidate). Migrate cannot move until `RenderError` lives in a neutral leaf
    package both `package cli` and `internal/cli/migrate` import. **Gates on M5 (uikit).**
  - **(ii) Shared-helper axis** — `migrate_restore_skill.go:31` calls `validateSkillID(skillID)`
    (defined `update_archive.go:99`); `:44` references the `archiveVersion` const (defined
    `update_archive.go:27`); `:80` calls `copyDirAll(archiveDir, targetDir)` (defined
    `update_archive.go:193`). These 3 symbols are co-located in `update_archive.go` and shared
    with the `update` cluster (M7). Resolution: co-extract to a neutral helper package (e.g.
    `internal/cli/archiveutil`) imported by both `migrate` and `update`, OR move with one cluster
    and export-import from the other.
  - **(iii) Reverse-dependency axis** — `update.go:1922-1928` constructs `&migrateAgencyRunner{...}`
    literal writing its unexported fields (`projectRoot`, `homeDir`, `dryRun`, `force`);
    `update.go:1933-1935` type-asserts `runErr.(*MigrateError)` and reads
    `me.Code == ErrMigrateNoSource`; `update_archive.go` constructs `&MigrateError{}` at 5 sites
    (lines `102, 109, 116, 138, 151`). A subpackage cannot expose unexported struct fields to
    `package cli`. Resolution: introduce `NewMigrateAgencyRunner(opts MigrateAgencyRunnerOpts)
    *migrateAgencyRunner` (or an interface) so `update.go:1922` no longer writes unexported
    fields; export `MigrateError`, its `Code` field, and the `ErrMigrateNoSource` const.
- **Resolution prerequisites**: M5 (uikit) for axis (i) + shared-helper co-extraction for axis
  (ii) + `update.go` constructor/field-encapsulation refactor for axis (iii). The `update.go`
  refactor lands in M7 (update cluster extraction) OR a prior refactor commit.
- Gate as M1, after ALL three axes are resolved at design time.

### M4 — Extract `constitution` cluster → `internal/cli/constitution` (LOW risk, IF kernel-free)

- Move `constitution.go` (543) + 3 test files. If it consumes kernel helpers, DEFER behind M5.
- Gate as M1.

### M5 — Extract `uikit` kernel → `internal/cli/uikit` (MED-HIGH risk; conditional)

- Prerequisite for kernel-dependent clusters (M3 migrate axis-i, M6 doctor, M7 update). Move +
  export shared helpers from `render.go`, `banner.go`, `settings.go`, `schema_bridge.go`;
  rewrite all `package cli` callers to `uikit.*`. Widely-used → higher blast radius. Only
  undertaken if M3/M6/M7 are pursued.
- Gate: build matrix + `go test ./...` (helpers exercised through their existing tests).

### M6 — Extract `doctor` cluster → `internal/cli/doctor` (MED risk; conditional, needs M5)

- Move 9 files (2,357 LOC) + 11 test files (~2,706 test LOC). No deps but heavy kernel use →
  imports `uikit` (M5). Gate as M1.

### M7 — Extract `update` cluster → `internal/cli/update` (HIGH risk, HIGHEST value; conditional)

- Move 9 files (5,181 LOC) + 21 test files (**9,283 test LOC**) + deps provider injection
  (`update.UpdateDeps` set from `deps.UpdateChecker`/`deps.UpdateOrch` via `PersistentPreRunE`,
  design.md §C). The dominant test-migration effort. Behind the post-M1 checkpoint.
- **M3 (migrate) gate**: the `update.go:1922` constructor refactor + `MigrateError`/
  `ErrMigrateNoSource` export MUST land here (or in a prior refactor commit) before M3 can ship.
- Gate as M1, with extra attention to `EnsureUpdate` provider wiring + security constants
  (`allowedUpdateScheme`/`allowedUpdateHost` in `deps.go` are referenced by `update.go` — verify
  they move or are exported).

### §F.9 Recommendation (honest value/risk call — revised post-M1-blocker)

**Recommend PHASED, recipe-proven-on-clean-first — NOT a full split, NOT big-bang.** Concretely:

1. **Definitely do M1 (agentlint)** — the only orchestrator-verified clean cluster. Proves the
   7-step recipe (design.md §B) with zero coupling-resolution work.
2. **Checkpoint decision after M1.** If the recipe felt smooth AND the user accepts the
   coupling-resolution cost, continue; otherwise STOP at M1.
3. **M2 (profile) and M3 (migrate) are conditional**, each behind a documented coupling
   resolution. M2 (profile) needs name-collision + `schema_bridge.go` resolution. M3 (migrate)
   needs M5 (uikit) + shared-helper co-extraction + `update.go` encapsulation refactor — it
   cannot ship until M5 + M7 refactor land. **The original M1 blocker proved migrate is NOT
   low-risk; do NOT treat M3 as a quick win.**
4. **M4 (constitution)**: cheap if kernel-free; otherwise fold into the M5-conditional set.
5. **Treat M5-M7 as conditional, gated by the post-M1 checkpoint.** `update` (M7) is the
   **highest-value** milestone (removes 5,181 LOC from the flat root) but also the **highest-risk**
   (9,283 test LOC + deps injection + M3 migrate axis-iii refactor). Do NOT front-load it.
6. **Do NOT extract** the 14 tiny single-file clusters (churn) or the deps/platform-tangled
   `launch/glm`, `hook`, `speccmd` (risk exceeds value) — spec.md §E.

Rationale: the M1-blocker incident proved that cluster characterization without verbatim
file:line verification is unsound. The revised ordering earns recipe confidence on a
genuinely-clean cluster first (M1 agentlint), then gates every coupling-resolution milestone
behind a user-facing checkpoint. Per "reject over-engineering", the plan reserves the right to
stop at M1. A full 93-file reorganization would be over-engineering; a bounded clean extraction
(M1) + opt-in conditional follow-ups is not.

## §G. Anti-Patterns

See design.md §G (AP-1 big-bang, AP-2 export-and-import-back cycle, AP-3 orphaned tests,
AP-4 tiny-cluster churn, AP-5 refactor-time behavior tests, AP-6 split platform siblings).

## §H. Cross-References

- research.md — cluster map, measurements, the two dominant risks.
- design.md — target layout, 7-step recipe, provider injection, uikit resolution.
- `internal/cli/CLAUDE.md` — cobra registration, cross-platform, subagent boundary conventions.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Tier L Section A-E template.
