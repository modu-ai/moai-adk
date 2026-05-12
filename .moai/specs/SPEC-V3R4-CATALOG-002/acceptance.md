# Acceptance Criteria — SPEC-V3R4-CATALOG-002

Given-When-Then scenarios for the slim distribution layer (SlimFS + CLI integration). Scenarios derived from EARS Requirements in `spec.md`. Edge cases derived from `plan.md` §"Risks & Mitigations".

> **Catalog Entry Definition (consistent with CATALOG-001 §Overview)**: "Catalog entry" 는 `internal/template/templates/.claude/skills/` 바로 아래 top-level 디렉토리 1개 (37개) 또는 `internal/template/templates/.claude/agents/moai/` 바로 아래 `.md` 파일 1개 (28개) 를 의미한다. 총 65 entries — 20 core skills + 20 core agents + 9 optional packs (~17 skills + 8 agents) + 1 harness-generated agent.

## Scenario 1: Default slim init deploys core-tier entries plus all non-catalog files

**Given**

- The binary embeds `catalog.yaml` with 65 entries (20 core skills + 20 core agents + 9 optional packs containing ~17 skills + 8 agents + 1 harness-generated agent).
- `moai init` is invoked in an empty target directory.
- Neither `MOAI_DISTRIBUTE_ALL` environment variable nor the `--all` flag is set.

**When**

The `runInit` function executes through the path `template.EmbeddedTemplates()` → `template.LoadEmbeddedCatalog()` → `template.SlimFS(rawFS, cat)` → `template.NewDeployerWithRenderer(slim, renderer)` → `deployer.Deploy(...)`.

**Then**

- The target project's `.claude/skills/` directory MUST contain exactly the 20 directories listed under `catalog.core.skills` in `catalog.yaml` (e.g., `moai/`, `moai-foundation-cc/`, `moai-workflow-spec/`, ...).
- The target project's `.claude/agents/moai/` directory MUST contain exactly the 20 `.md` files listed under `catalog.core.agents` (e.g., `manager-spec.md`, `expert-backend.md` if core; based on actual manifest contents).
- The target project MUST NOT contain any of the 9 optional packs' skills or agents (e.g., `moai-domain-backend/`, `expert-mobile.md`).
- The target project MUST NOT contain the harness-generated `builder-harness.md` agent.
- All non-catalog template files MUST deploy verbatim:
  - `.claude/rules/**` (rule files)
  - `.claude/output-styles/**`
  - `.claude/commands/**` (if present)
  - `.moai/config/sections/*.yaml`
  - `CLAUDE.md`, `.gitignore`
- Stdout MUST include an informational message containing the substring `"slim mode"` and the substring `"--all"` (REQ-CATALOG-002-020 indirectly; explicit text per plan.md M2-T2.4).

> Maps: REQ-CATALOG-002-001, REQ-CATALOG-002-002, REQ-CATALOG-002-004, REQ-CATALOG-002-007, REQ-CATALOG-002-010, REQ-CATALOG-002-011.

## Scenario 2: Env var opt-out restores full deploy

**Given**

- `moai init` is invoked in an empty target directory.
- The environment variable `MOAI_DISTRIBUTE_ALL=1` is set (or `MOAI_DISTRIBUTE_ALL=true` / `MOAI_DISTRIBUTE_ALL=TRUE` / `MOAI_DISTRIBUTE_ALL=True`).
- The `--all` flag is NOT set.

**When**

The `runInit` function evaluates `shouldDistributeAll(cmd)` which returns `true`, causing the SlimFS wrap step to be skipped.

**Then**

- The target project's `.claude/skills/` directory MUST contain ALL 37 top-level skill directories (including optional packs' skills and harness-generated harness if any).
- The target project's `.claude/agents/moai/` directory MUST contain ALL 28 agent `.md` files (including optional pack agents and `builder-harness.md`).
- The deployment surface MUST be bit-identical to pre-CATALOG-002 behavior (verified by counting files vs main HEAD `0d4bf14ef` baseline OR by running `moai init` from the previous binary and comparing).
- No stdout informational message about slim mode MUST appear.

> Maps: REQ-CATALOG-002-012.

## Scenario 3: --all flag opt-out restores full deploy

**Given**

- `moai init` is invoked in an empty target directory.
- The `--all` flag is set: `moai init --all` or `moai init my-app --all`.
- The `MOAI_DISTRIBUTE_ALL` environment variable is NOT set.

**When**

The `runInit` function evaluates `shouldDistributeAll(cmd)` which returns `true` because `cmd.Flags().GetBool("all") == true`.

**Then**

- The deployment surface MUST be identical to Scenario 2 (full deploy).
- No stdout informational message about slim mode MUST appear.
- The `init --help` text MUST list the `--all` flag with description containing the substring `"slim"` or `"core templates"`.

> Maps: REQ-CATALOG-002-013, REQ-CATALOG-002-019.

## Scenario 4: SlimFS hides every non-core entry from fs.FS access

**Given**

- The embedded `catalog.yaml` is loaded into `*Catalog`.
- `SlimFS(rawFS, cat)` returns a wrapped fs.FS.

**When**

The audit test `TestSlimFS_HidesNonCoreEntries` iterates `cat.AllEntries()` and for each entry where `entry.Tier != template.TierCore`:
- Strips the `"templates/"` prefix from `entry.Path` to obtain `relPath`.
- Calls `_, err := fs.Stat(slim, relPath)`.

The audit test `TestSlimFS_WalkDirNoLeak` invokes `fs.WalkDir(slim, ".", walkFn)` and collects every visited path.

**Then**

- `fs.Stat` MUST return an error satisfying `errors.Is(err, fs.ErrNotExist)` for every non-core entry path; otherwise, the test fails with `t.Errorf("CATALOG_SLIM_LEAK: <path> tier=<tier>")`.
- `fs.WalkDir` MUST NOT visit any path that matches a hidden non-core entry's prefix; otherwise, the test fails with `t.Errorf("CATALOG_SLIM_WALK_LEAK: <path>")`.
- All sentinel emissions in M3 tests MUST use `t.Errorf` (NOT `t.Logf` — CATALOG-001 evaluator-active iter 1 EC3 lesson).

> Maps: REQ-CATALOG-002-010, REQ-CATALOG-002-011, REQ-CATALOG-002-014, REQ-CATALOG-002-017.

## Scenario 5: SlimFS preserves all non-catalog template paths

**Given**

- The embedded `catalog.yaml` enumerates 37 skills + 28 agents, leaving many other template paths unmanaged (e.g., `.claude/rules/`, `.claude/output-styles/`, `.moai/config/`, `CLAUDE.md`, `.gitignore`, `.claude/settings.example.json`).
- `SlimFS(rawFS, cat)` returns a wrapped fs.FS.

**When**

The audit test `TestSlimFS_PreservesNonCatalogFiles` calls `fs.Stat(slim, path)` for each member of a predefined non-catalog path list:
- `".claude/rules/moai/core/zone-registry.md"` (or another stable rule file)
- `".moai/config/sections/quality.yaml"`
- `"CLAUDE.md"`
- `".gitignore"`

**Then**

- Every listed non-catalog path MUST exist in the slim FS; if any path returns an error, the test fails with `t.Errorf("CATALOG_SLIM_OVER_FILTER: <path>")`.
- The audit test `TestSlimFS_PreservesCoreEntries` MUST also pass: every entry with `tier == TierCore` is reachable, with sentinel `CATALOG_SLIM_CORE_MISSING: <path>` on failure.

> Maps: REQ-CATALOG-002-015, REQ-CATALOG-002-016.

## Scenario 6: D7 lock preserved — deployer.go and update.go are unchanged

**Given**

- Pre-merge HEAD: `0d4bf14ef` (CATALOG-001 eval-1 fix, main).
- Post-merge HEAD: the SPEC-V3R4-CATALOG-002 merge commit.

**When**

A reviewer runs:
- `git diff 0d4bf14ef..HEAD -- internal/template/deployer.go`
- `git diff 0d4bf14ef..HEAD -- internal/cli/update.go`
- `go test ./internal/template/ -run Deploy`
- `go test ./internal/cli/ -run Update`

**Then**

- `git diff` for `deployer.go` MUST be empty (0 lines added, 0 removed).
- `git diff` for `update.go` MUST be empty.
- All existing `deployer_test.go` test cases MUST remain GREEN without modification.
- All existing `update_test.go` test cases MUST remain GREEN without modification.

> Maps: REQ-CATALOG-002-005, REQ-CATALOG-002-006, REQ-CATALOG-002-018.

## Scenario 7: moai update continues to deploy full FS

**Given**

- A project exists that was initialized in slim mode (only core entries on disk).
- `moai update` is invoked in that project.

**When**

The `runUpdate` flow in `internal/cli/update.go:445-479` executes, which calls `template.EmbeddedTemplates()` (unfiltered) and constructs `template.NewDeployerWithRendererAndForceUpdate(embedded, renderer, true)`.

**Then**

- The deploy step MUST iterate the full embedded FS (all 65 catalog entries + non-catalog files).
- Non-core entries that were absent in the slim-initialized project MAY be re-added by update (since `--force-update == true`). This is intentional pre-CATALOG-004 behavior; the safe drift sync is deferred to SPEC-V3R4-CATALOG-004.
- No CATALOG-002 code path MUST trigger from `runUpdate`. Specifically, `SlimFS` MUST NOT be called from `update.go`.

> Maps: REQ-CATALOG-002-009.
>
> **Note**: This scenario codifies the "init-only filter" boundary. Update users who want to preserve their slim project should defer running `moai update` until SPEC-V3R4-CATALOG-004 introduces drift-aware sync.

## Edge Cases

### EC1: catalog.yaml absent or corrupt at init time

- **Condition**: A custom-built binary is missing `catalog.yaml` from its embed, OR the file is present but malformed YAML.
- **Behavior**: `LoadEmbeddedCatalog()` returns an error. `runInit` wraps that error with the prefix `"CATALOG_LOAD_FAILED: "` and returns immediately, BEFORE invoking the Deployer or writing any files. The target project directory MUST remain empty (or unchanged if pre-existing).
- **Recovery**: Reinstall a properly-built `moai` binary. Verify with `moai version` and `go doctor` or equivalent.
- **Note**: The fast-fail behavior is critical because a partial deployment could leave the project in an inconsistent state.
- **Maps**: REQ-CATALOG-002-008.

### EC2: Both MOAI_DISTRIBUTE_ALL and --all set simultaneously

- **Condition**: User runs `MOAI_DISTRIBUTE_ALL=1 moai init --all`.
- **Behavior**: `shouldDistributeAll(cmd)` returns `true` regardless of which condition triggered. Slim filter is bypassed; full FS is deployed exactly once. No error or warning.
- **Recovery**: N/A (intended idempotent combinability).
- **Maps**: REQ-CATALOG-002-019.

### EC3: MOAI_DISTRIBUTE_ALL set to a non-recognized value

- **Condition**: User runs `MOAI_DISTRIBUTE_ALL=0 moai init` or `MOAI_DISTRIBUTE_ALL=yes moai init` or `MOAI_DISTRIBUTE_ALL= moai init` (empty).
- **Behavior**: `shouldDistributeAll(cmd)` returns `false` because the env value is not `"1"` and is not case-insensitive `"true"`. Slim mode is active. The stdout informational message is printed.
- **Recovery**: Use the canonical values `"1"` or `"true"`.
- **Note**: This narrow matching rule (OQ1 권장안) prevents accidental opt-out from typos like `MOAI_DISTRIBUTE_ALL=ye`.
- **Maps**: REQ-CATALOG-002-012.

### EC4: Nested path under a core skill (e.g., moai/workflows/plan.md)

- **Condition**: The `moai` skill is tier=`core` with `path == "templates/.claude/skills/moai/"`. Inside this directory are sub-files like `workflows/plan.md`, `team/leader.md`, `references/*`.
- **Behavior**: SlimFS MUST allow all sub-paths under the core skill's directory to pass through. The filter operates at the catalog-entry granularity (top-level directory or single `.md` file), not at the sub-file level. `fs.Stat(slim, ".claude/skills/moai/workflows/plan.md")` MUST succeed.
- **Audit verification**: `TestSlimFS_PreservesCoreEntries` includes a sub-assertion that verifies at least one nested path (e.g., `.claude/skills/moai/workflows/plan.md`) is reachable.
- **Note**: This invariant is critical because the `moai` skill is a multi-module container. CATALOG-001's `catalog_doc.md` defines this semantics: sub-files are modules of the entry, not separate entries.
- **Maps**: REQ-CATALOG-002-010 (allow-by-default for core), REQ-CATALOG-002-015.

## Quality Gates

- [ ] `go test -race -count=1 ./internal/template/...` PASS (all audit tests + unit tests + existing deployer tests).
- [ ] `go test -race -count=1 ./internal/cli/...` PASS (existing init / update tests + new slim init sub-test).
- [ ] `go vet ./internal/template/... ./internal/cli/...` clean.
- [ ] `golangci-lint run --timeout=5m ./internal/template/... ./internal/cli/...` PASS (0 issues).
- [ ] Test coverage for `internal/template/slim_fs.go` ≥ 90% (critical-package threshold per CLAUDE.local.md §6).
- [ ] All 7 Given-When-Then scenarios pass when run individually AND in `t.Parallel()` mode.
- [ ] All 4 edge cases (EC1-EC4) have corresponding negative test cases in `slim_fs_test.go` or `catalog_slim_audit_test.go`.
- [ ] No HARD rule violations:
  - **D7 lock preservation**: `git diff` of `internal/template/deployer.go` empty (Scenario 6 verification).
  - **Template-First (CLAUDE.local.md §2)**: SlimFS lives in `internal/template/`, not in `internal/template/templates/`.
  - **16-language neutrality (CLAUDE.local.md §15)**: SlimFS treats all paths language-agnostically.
  - **No hardcoded values (CLAUDE.local.md §14)**: Tier comparison uses `template.TierCore` constant (defined in CATALOG-001 catalog_loader.go).
- [ ] CI 6 jobs all green: Lint, Test (ubuntu-latest), Test (macos-latest), Test (windows-latest), Build (linux/amd64), CodeQL.
- [ ] **Cross-platform deny set normalization**: `TestSlimFS_HidesNonCoreEntries` GREEN on Linux + macOS + Windows. Path normalization (forward slash invariant for `fs.FS`) verified pre-emptively.
- [ ] No backward incompatibility:
  - `internal/template/deployer_test.go` existing test cases all GREEN (M4-T4.2).
  - `internal/cli/update_test.go` existing test cases all GREEN (M4-T4.3).
  - `internal/cli/init_test.go` either preserved via M4-T4.1 strategy B (existing tests wrap with `t.Setenv("MOAI_DISTRIBUTE_ALL", "1")`) or updated to slim mode expectations.
- [ ] `CHANGELOG.md` has Unreleased section entry with `BREAKING CHANGE` annotation listing both opt-out mechanisms.

## Review Process Checklist (non-runtime, PR review only)

The following items are not runtime audit checks; they are independent process invariants surfaced to PR reviewers:

- Frontmatter compliance for SPEC documents (spec.md, plan.md, acceptance.md): 9 required fields in spec.md, no legacy aliases (`created`, `updated`, `spec_id`). Verified during plan-auditor independent review.
- D7 lock (deployer.go no-modify): verified via `git diff` empty for that file. plan-auditor flags violations at PR review time as REVISE.
- Boundary clarity with CATALOG-003 / CATALOG-004 / CATALOG-005: plan-auditor confirms Exclusions section explicitly defers these scopes.
- BREAKING CHANGE annotation in CHANGELOG: plan-auditor confirms presence and accuracy.

## Performance Criteria

- **SlimFS construction time**: < 1ms for the 65-entry catalog (deny set is a simple `map[string]struct{}` with O(n) build, O(1) lookup).
- **SlimFS Open/Stat overhead per call**: < 100µs amortized. Single hash-map lookup + delegated underlying call. Negligible compared to file I/O.
- **moai init wall-clock time with slim mode**: NO MORE THAN 5% slower than full deploy at the runInit layer. In practice, slim mode should be FASTER because fewer files are written to disk. Target: slim init ≥ 30% fewer files written → measurable speedup.
- **Binary size impact**: 0 net change. SlimFS code is small (~150-200 LOC); no new embedded assets.
- **Build time impact**: < 100ms increase to `make build` (Go compilation of additional files).
- **Memory footprint at runtime**: Deny set holds ~25 string keys × ~80 chars each ≈ 2KB heap. Negligible.

## Verification Procedure (post-implementation)

For maintainers verifying the merge:

1. **Build the binary**: `make build` succeeds without warnings.
2. **Run audit suite**: `go test ./internal/template/ -run TestSlimFS -v` shows 4 sub-tests PASS.
3. **Run regression suite**: `go test -race -count=1 ./...` shows full GREEN.
4. **Smoke-test slim init**: `./bin/moai init /tmp/smoke-slim` produces a project with ~20 skills (not 37). Stdout shows slim mode message.
5. **Smoke-test full init**: `MOAI_DISTRIBUTE_ALL=1 ./bin/moai init /tmp/smoke-full` produces a project with 37 skills.
6. **Smoke-test --all flag**: `./bin/moai init /tmp/smoke-all --all` produces a project with 37 skills.
7. **Verify D7 lock**: `git diff main..HEAD -- internal/template/deployer.go internal/cli/update.go` is empty.
8. **Verify CHANGELOG**: `grep -A 5 'BREAKING CHANGE' CHANGELOG.md` shows the slim init entry.
