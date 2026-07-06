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

## §F. Milestones (risk-ascending; one cluster each)

> Milestones are gated by a re-evaluation **checkpoint** after M4. The SPEC does NOT commit to
> reaching M7 — see REQ-CSS-010 and the recommendation in §F.9.

### M1 — Extract `migrate` cluster → `internal/cli/migrate` (LOW risk, proves the recipe)
- Move 9 files (`migrate_agency*.go`, `migration.go`, `migrate_restore_skill.go` + platform siblings)
  + 4 test files (~1,115 test LOC). No deps coupling, minimal kernel dependency.
- Export `MigrationCmd`/`NewMigrateCmd`; rewire `root.go` (`rootCmd.AddCommand(migrationCmd)` →
  `migrate.MigrationCmd`). Recipe steps 1-7 (design.md §B).
- Gate: build matrix + `go test ./...` + `moai --help` diff empty.

### M2 — Extract `profile` cluster → `internal/cli/profile` (LOW risk)
- Move 3 files (`profile_setup.go`, `profile_setup_translations.go`, `profile.go`) + 8 test files
  (~1,112 test LOC). No deps coupling. Verify kernel-helper usage; if any, defer to post-M5.
- Gate as M1.

### M3 — Extract `agentlint` cluster → `internal/cli/agentlint` (LOW-MED risk)
- Move `agent_lint.go` (1,154) + `workflow_lint.go` (235) + the single 2,108-LOC white-box test.
  The large single test file is the friction — verify every unexported reference resolves in the
  new package.
- Gate as M1.

### M4 — Extract `constitution` cluster → `internal/cli/constitution` (LOW risk, IF kernel-free)
- Move `constitution.go` (543) + 3 test files. If it consumes kernel helpers, DEFER behind M5.
- Gate as M1.

### CHECKPOINT (after M4) — re-evaluate marginal value (REQ-CSS-010)
- Recipe proven on ~4,700 LOC across 4 clusters with zero import-cycle work. Assess: was the
  navigability gain worth the churn? Is the remaining test-migration risk (update = 9,283 test LOC)
  justified? Decide: STOP (ship M1-M4) OR continue to M5+.
- Log the decision to `progress.md § Mode Selection` / checkpoint note.

### M5 — Extract `uikit` kernel → `internal/cli/uikit` (MED-HIGH risk; conditional)
- Prerequisite for kernel-dependent clusters. Move + export shared helpers from `render.go`,
  `banner.go`, `settings.go`, `schema_bridge.go`; rewrite all `package cli` callers to `uikit.*`.
  Widely-used → higher blast radius. Only undertaken if M6/M7 are pursued.
- Gate: build matrix + `go test ./...` (helpers exercised through their existing tests).

### M6 — Extract `doctor` cluster → `internal/cli/doctor` (MED risk; conditional, needs M5)
- Move 9 files (2,357 LOC) + 11 test files (~2,706 test LOC). No deps but heavy kernel use →
  imports `uikit` (M5). Gate as M1.

### M7 — Extract `update` cluster → `internal/cli/update` (HIGH risk, HIGHEST value; conditional)
- Move 9 files (5,181 LOC) + 21 test files (**9,283 test LOC**) + deps provider injection
  (`update.UpdateDeps` set from `deps.UpdateChecker`/`deps.UpdateOrch` via `PersistentPreRunE`,
  design.md §C). The dominant test-migration effort. Behind the post-M4 checkpoint.
- Gate as M1, with extra attention to `EnsureUpdate` provider wiring + security constants
  (`allowedUpdateScheme`/`allowedUpdateHost` in `deps.go` are referenced by `update.go` — verify
  they move or are exported).

### §F.9 Recommendation (honest value/risk call)

**Recommend PHASED, lowest-risk-first — NOT a full split, NOT big-bang.** Concretely:

1. **Definitely do M1-M3** (migrate, profile, agentlint): ~4,050 LOC of cohesive clusters,
   kernel-free, deps-free, proves the recipe at low risk with real navigability gain on the
   larger clusters.
2. **Evaluate M4 + checkpoint**: cheap if constitution is kernel-free; otherwise fold into the
   checkpoint decision.
3. **Treat M5-M7 as conditional, gated by the checkpoint.** `update` (M7) is the **highest-value**
   milestone (removes 5,181 LOC from the flat root) but also the **highest-risk** (9,283 test LOC
   + deps injection). Do NOT front-load it — prove the recipe first, then decide. My recommendation
   is to pursue M5+M6 (doctor) only if the M1-M4 experience shows the recipe is smooth, and to
   pursue M7 (update) only with explicit user sign-off given its test surface.
4. **Do NOT extract** the 14 tiny single-file clusters (churn) or the deps/platform-tangled
   `launch/glm`, `hook`, `speccmd` (risk exceeds value) — spec.md §E.

Rationale: the task is a refactor of working code whose gains are invisible to users; per
"reject over-engineering", the plan buys down risk incrementally and reserves the right to stop.
A full 93-file reorganization would be over-engineering; a bounded high-value extraction is not.

## §G. Anti-Patterns

See design.md §G (AP-1 big-bang, AP-2 export-and-import-back cycle, AP-3 orphaned tests,
AP-4 tiny-cluster churn, AP-5 refactor-time behavior tests, AP-6 split platform siblings).

## §H. Cross-References

- research.md — cluster map, measurements, the two dominant risks.
- design.md — target layout, 7-step recipe, provider injection, uikit resolution.
- `internal/cli/CLAUDE.md` — cobra registration, cross-platform, subagent boundary conventions.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Tier L Section A-E template.
