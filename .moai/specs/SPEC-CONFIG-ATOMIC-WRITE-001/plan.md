# SPEC-CONFIG-ATOMIC-WRITE-001 — Implementation Plan

> Slice (b) of `SPEC-CONFIG-TIER-PERSIST-001`. Owns atomic, mode-preserving writes for
> `.moai/config` + CLI persistence. Plan-phase only — this document authorizes no code changes.

## §A Context

- **Worktree**: `ctp-atomic-write` (CWD = worktree root). HEAD = `ed70e4354`, clean, `0 0` vs
  `origin/main`.
- **Parent SPEC**: `SPEC-CONFIG-TIER-PERSIST-001` (status: draft, Tier M-exceeding, 35 REQ).
  This slice extracts the parent's §D.4 (REQ-CTP-021..027).
- **Parent split memo**: `.moai/specs/SPEC-CONFIG-TIER-PERSIST-001/spec.md` `## §L Split Branches`.
- **Code baseline**: `ed70e4354`. All file:line evidence in `spec.md` §A is attributable to this
  tree (`git diff --stat ed70e4354 -- internal/` is empty at plan authoring time).
- **Prior art in-repo** (build on these, do not re-invent):
  - `internal/settings/yamlpatch/yamlpatch.go:202` — `os.Chmod(tmpName, info.Mode().Perm())`
    before rename.
  - `internal/cli/update_deny_migration.go:95-100` — stat-then-preserve-mode.
  - `internal/cli/update_disk_backup.go:39` — `var diskBackupWriteFile = os.WriteFile` injectable
    seam (the testability pattern).
- **Named constant**: `internal/defs/perms.go:11` `defs.FilePerm` (`0o644`).

## §B Known Issues (auto-injected, filtered to relevant categories)

- **B1 Cross-platform build tags** — not in scope (no `syscall` use); but
  `GOOS=windows GOARCH=amd64 go build ./...` MUST still pass after the fix.
- **B4 Frontmatter canonical schema** — `created:`/`updated:`/`tags:` only; no snake_case aliases.
- **B5 CI 3-tier awareness** — spec-lint, golangci-lint, Test (per OS) each fail independently.
  The guard test (REQ-CAW-007) is a NEW test; baseline it before claiming NEW=0.
- **B6 spec-lint heading convention** — `## Out of Scope` (h2) alone triggers `MissingExclusions`;
  `### Out of Scope — <topic>` (h3) sub-sections are required (satisfied in `spec.md` §C).
- **B8 Working tree hygiene** — do NOT modify runtime-managed files (`.moai/state/*`,
  `.moai/harness/*`, `.moai/cache/*`); scope edits to `internal/config/` + `internal/cli/` only.
- **B10 Untouched paths PRESERVE** — do NOT touch `internal/cli/update/backup/`,
  `internal/cli/update/merge/`, or `internal/core/project/` (those are sibling slice (c) territory).
- **B11 AskUserQuestion prohibited** — this is `internal/config` + `internal/cli` (subagent
  boundary per `internal/cli/CLAUDE.md`); no `AskUserQuestion` / `mcp__askuser` calls introduced.

## §C Pre-flight (run before M1)

```bash
git branch --show-current
git rev-parse HEAD                                     # expect ed70e4354 (or descendant)
go build ./...
GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m 2>&1 | tail -5          # baseline NEW vs pre-existing
grep -rn 'os\.WriteFile' internal/config/ internal/cli/ | grep -v '_test.go'   # enumerate call sites
ls internal/defs/perms.go                              # confirm named-constant SSOT exists
```

## §D Constraints (DO NOT VIOLATE)

- **PRESERVE** (do not modify): `internal/cli/update/backup/restore.go`,
  `internal/cli/update/merge/merge.go`, `internal/core/project/initializer.go` — these are
  sibling-slice-(c) territory. The helper is extracted with a public-enough API that they can
  adopt it later; this SPEC does not migrate them.
- **No hardcoded mode literals** in new/changed code: route through `defs.FilePerm` or a named
  `defs` constant. Existing call sites that already use `defs.FilePerm` (update_wizard.go,
  update.go:1004) are fine on the mode axis; they only need the atomicity fix.
- **No `--no-verify`**, no `--amend`, no force-push.
- **Conventional Commits** with `🗿 MoAI` trailer; per-SPEC scope (`fix(SPEC-CONFIG-ATOMIC-WRITE-001): M{N} ...`).
- **Test isolation**: `t.TempDir()` for every AC test; never write to project root or `~/.claude`.
- **Error wrapping**: `fmt.Errorf("...: %w", err)`, never string concat.
- **English** code/comments/godoc (`code_comments: en`).

## §E Self-Verification Deliverables (run-phase, per VCI 5-section format)

Each E-item reports (a) command, (b) verbatim observed output, (c) baseline-attribution
`(this run, this tree, HEAD <sha>)`.

- **E1 — AC Binary PASS/FAIL Matrix** (AC-CAW-001..007 vs `go test -run ./internal/config/... ./internal/cli/...`).
- **E2 — Cross-platform build**: `go build ./...` AND `GOOS=windows GOARCH=amd64 go build ./...`.
- **E3 — Coverage**: `go test -cover ./internal/config/...` ≥ 85%; `go test -cover ./internal/cli/...` ≥ 85%.
- **E4 — Subagent boundary grep**: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/config/ internal/cli/ | grep -v '_test.go' | grep -v '//'` → 0 matches.
- **E5 — Lint**: `golangci-lint run --timeout=2m`; NEW issues reported explicitly, baseline separated.
- **E6 — Branch HEAD + push state**.
- **E7 — Blocker report** (if any).
- **E8 — RED failure output** (TDD): the verbatim failing-test output captured BEFORE GREEN for
  AC-CAW-002 (mode-preservation round-trip) and AC-CAW-006 (temp-cleanup-on-error).

**Full-suite verification (feedback-full-test-suite-verification)**: E1 MUST be runnable as
`go test ./...`, not only package-local. Any AC that could only be verified package-locally is
flagged as a Gap.

## §F Milestones (ordered by decision-reversibility — highest-change-likelihood first)

### M1 — Shared atomic-write helper: API design + implementation (HIGH reversal risk)

The helper's signature is the decision most likely to change and the one every later milestone
depends on. Lock it before touching any call site.

- **Locked decision (D3)**: the helper is `atomicfile.Write(path string, data []byte, defaultMode os.FileMode) error`
  in a new sub-package `internal/config/atomicfile/`. Justification: sub-package isolation makes
  the helper the single choke point the guard (REQ-CAW-007) whitelists without entangling it with
  `ConfigManager` state in the parent `internal/config` package; importable by both
  `internal/config` (parent imports child — valid in Go) and `internal/cli` without import cycles.
- Implement: temp file in `filepath.Dir(path)` → write data → `os.Stat(path)`: on success
  `os.Chmod(tmp, existingMode.Perm())`, on `os.IsNotExist` `os.Chmod(tmp, defaultMode)` → `os.Rename`.
- Temp cleanup: `defer os.Remove(tmpName)` on the error path (already present in the current
  `atomicWrite`; preserve it).
- Unit-test the helper directly (modes 0600, 0644, 0750; new-file case; error-path cleanup).
- **API sign-off required before M2.**

### M2 — Rewrite `internal/config/manager.go` `atomicWrite` to delegate to the helper

- Replace the body of `atomicWrite` (manager.go:420-438) with a call to the M1 helper.
- `saveSection` (manager.go:408-417) passes `defs.FilePerm` as the default mode.
- Preserve the existing `defer os.Remove(tmpName)` semantics via the helper's own cleanup.
- Add the mode-preservation round-trip test in `manager_test.go` using `t.TempDir()`.

### M3 — CLI call-site remediation (mechanical, once M1 API is locked)

Route each listed call site through the M1 helper. This is mechanical adoption, not design.

- `internal/cli/harness.go:390` — replace bare `os.WriteFile(configPath, newData, 0o644)` with the
  helper call (`defs.FilePerm` default). **Fixes both non-atomicity AND the hardcoded literal.**
- `internal/cli/update_cleanup.go:243,258,446` — route through helper.
- `internal/cli/update_preserve_inventory.go:350` — route through helper.
- `internal/cli/update_wizard.go:182,222,277,309,346` — route through helper (already uses
  `defs.FilePerm`; only the atomicity fix is needed).
- `internal/cli/update.go:1004` — route through helper (already uses `defs.FilePerm`).
- `internal/config/toolpolicy/tier_render.go:146` — route through helper. Writes settings.json
  persistence path via `renderIntoFile(path)`; hardcoded `0o644` violates REQ-CAW-005.
- `internal/config/toolpolicy/codegen.go:244` — route through helper. Writes
  `.claude/settings.json` via `BuildInto(path, ...)`; hardcoded `0o644` violates REQ-CAW-005.
- **Out of scope** (do NOT touch): `internal/cli/update/backup/restore.go`,
  `internal/cli/update/merge/merge.go`, `internal/core/project/initializer.go` — sibling slice (c).
- **Exempted** (not config writes): `update_cleanup.go:69` (lock file, `O_EXCL`),
  `update_cleanup.go:390` (probe file at `0o600` — deliberate; the probe is not a config file).

#### Classified out-of-scope sites (D2 — backup-dir artifacts)

These `os.WriteFile` call sites were classified and EXEMPTED — they do not target config or
settings persistence paths. The guard (REQ-CAW-007) exempts them by predicate, not by human
judgment per call. Each row records the site, its resolved target path, and the classification
predicate.

| Site | Resolved target path | Classification predicate |
|------|----------------------|--------------------------|
| `internal/cli/update_recovery_manifest.go:78` | `backupDir/<recoveryManifestFileName>` | target path is under a backup directory, not `.moai/config/**` or a settings path |
| `internal/cli/update_namespace_protect.go:242` | `backupDir/.complete` | target path is a backup-directory atomicity marker, not a config or settings path |

The guard's target-path resolution algorithm (REQ-CAW-007) reaches the same classification
mechanically; the table above is the expected output, not a human override.

### M4 — Regression guard test (REQ-CAW-007)

- Add a guard test (e.g. `internal/config/atomicwrite_guard_test.go` or
  `internal/cli/no_bare_writefile_test.go`) that scans non-test `.go` files under
  `internal/config/` and `internal/cli/` for `os.WriteFile` calls whose target path resolves to
  `.moai/config/**` or settings/config persistence paths, and fails on any match outside an
  enumerated exemption list.
- **Exemptions to enumerate**: the `diskBackupWriteFile` seam declaration
  (`update_disk_backup.go:39`), probe/lock/log writes that do not target config paths, and the
  helper's own internal temp-file write (the helper writes a temp, not a config path, via
  `os.CreateTemp`).
- **Mutation-verify**: temporarily reintroduce a bare `os.WriteFile` to a config path in a
  throwaway edit, confirm the guard fails, then revert. Capture the RED output for E8.

### M5 — Full-suite verification gate

- `go test ./...` green (NOT package-local only — per feedback-full-test-suite-verification).
- `GOOS=windows GOARCH=amd64 go build ./...` succeeds.
- `golangci-lint run --timeout=2m` NEW=0 (baseline separated).
- Coverage on `internal/config` and touched `internal/cli` packages ≥ 85%.
- All 7 AC PASS.

## §G Anti-Patterns

- **AP-1 — Inlining a mode literal** in the helper or any call site. Route through `defs.FilePerm`
  or a named `defs` constant.
- **AP-2 — Migrating the deferred (c)-slice writers** (`backup/restore.go`, `merge.go`,
  `core/project/initializer.go`) inside this SPEC. They are explicitly out of scope; touching them
  crosses the split boundary and re-bloats this slice.
- **AP-3 — Package-local-only test verification**. The AC MUST pass under `go test ./...`; a test
  that only passes under `go test ./internal/config/` is a Gap, not a PASS.
- **AP-4 — Widening the guard to every `os.WriteFile`** under `internal/cli/`. The guard targets
  config/settings *persistence paths*; legitimate non-config writes (probe, lock, log) are
  exempted by path, not silently.
- **AP-5 — Re-inventing the pattern**. `yamlpatch.go:202` already has the canonical
  Chmod-before-rename; follow it.
- **AP-6 — Dropping the `defer os.Remove(tmpName)`** when rewriting `atomicWrite`. The error-path
  cleanup is load-bearing (REQ-CAW-006, AC-CAW-006).

## §H Cross-References

- `spec.md` (this SPEC) — §A defects, §C scope boundary, §D requirements.
- `acceptance.md` (this SPEC) — AC-CAW-001..007 Given-When-Then.
- `.moai/specs/SPEC-CONFIG-TIER-PERSIST-001/spec.md` `## §L Split Branches` — lineage record.
- `internal/settings/yamlpatch/yamlpatch.go:202` — canonical pattern reference.
- `internal/cli/update_deny_migration.go:95-100` — stat-then-preserve pattern.
- `internal/cli/update_disk_backup.go:39` — injectable seam (testability pattern + guard exemption).
- CLAUDE.local.md §6 (Test Isolation), §14 (Hardcoding prevention).
