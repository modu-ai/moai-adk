---
id: SPEC-CONFIG-ATOMIC-WRITE-001
title: "atomic, mode-preserving writes for .moai/config and CLI persistence"
version: "0.2.0"
status: draft
created: 2026-08-12
updated: 2026-08-12
author: manager-spec
priority: P0
phase: "v3.0.2 target"
module: "internal/config"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "config, atomic-write, file-mode, data-integrity, persistence, write-helper"
related_specs: [SPEC-CONFIG-TIER-PERSIST-001, SPEC-V3R5-ATOMIC-WRITE-001]
depends_on: []
---

# SPEC-CONFIG-ATOMIC-WRITE-001

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-08-12 | Initial draft. Slice (b) carved out of the over-large parent `SPEC-CONFIG-TIER-PERSIST-001` (35 REQ, Tier-M-exceeding, 3-way split recommended). Owns the parent's §D.4 REQ-CTP-021..027 (atomic-write + mode-preservation) as an independently-shippable unit. Live defects verified against HEAD `ed70e4354`. |
| 0.2.0 | 2026-08-12 | Iteration-2 re-author per plan-auditor FAIL 0.74 → clear Tier M threshold 0.80. D1: reframe §A motivation as forward-looking invariant (not "files already narrowed"); defer REQ-CTP-025/026 (mode-widening migration) to follow-up sibling `SPEC-CONFIG-MODE-MIGRATE-001`; correct every REQ-mapping line. D2: classify 4 omitted `os.WriteFile` sites (2 backup-dir exempt, 2 toolpolicy settings-persistence → M3 remediation). D3: lock helper identifier `atomicfile.Write`. D4: lowercase `SHALL` → `shall` in REQ-CAW-004a. D5: add SPEC-V3R5-ATOMIC-WRITE-001 relationship qualifier. |

## §A Problem / Motivation

The `.moai/config` and CLI persistence layer writes configuration through two distinct code paths,
and both violate one or both of two invariants that a configuration store depends on:

1. **Atomicity** — a write either fully lands or does not land at all. A crash mid-write must leave
   either the complete new content or the complete prior content, never a truncated file.
2. **Mode preservation** — the destination file's permission bits (mode) survive the write
   unchanged. A write must not narrow the mode (a 0600 temp clobbering a 0644 file) nor widen it
   (a 0644 hardcode clobbering a 0600 secret file).

**Defect 1 — `internal/config/manager.go:420-438` narrows the mode on every section save.**

`atomicWrite` creates its temp file with `os.CreateTemp`, which uses mode 0600. `os.Rename` then
makes that 0600 file the live config file. No `os.Chmod` call exists on this path, so a 0644 config
file is permanently narrowed to 0600 after any section save. Atomicity is satisfied by the rename;
**mode preservation is the gap**. `defs.FilePerm` (`internal/defs/perms.go:11`, `0o644`) is never
applied. Verified at HEAD `ed70e4354`:

```
$ sed -n '420,438p' internal/config/manager.go
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".moai-config-*.tmp")  // CreateTemp → mode 0600
	...
	return os.Rename(tmpName, path)  // ← no os.Chmod: result inherits temp's 0600
}
```

**Defect 2 — `internal/cli/harness.go:390` is both non-atomic and hardcoded.**

```
if err := os.WriteFile(configPath, newData, 0o644); err != nil { ... }
```

Bare `os.WriteFile` truncates before writing (partial-write window on crash) AND hardcodes `0o644`
rather than `defs.FilePerm`. Both invariants are violated in one line.

**Defect 3 — `internal/cli/update*.go` production writes are bare `os.WriteFile` with hardcoded
modes.** Verified at HEAD `ed70e4354`:

| File:line | Mode literal | Atomic? |
|-----------|--------------|---------|
| `internal/cli/update_cleanup.go:243` | `0o644` | no |
| `internal/cli/update_cleanup.go:258` | `0o644` | no |
| `internal/cli/update_cleanup.go:446` | `0o644` | no |
| `internal/cli/update_preserve_inventory.go:350` | `0o644` | no |
| `internal/cli/update_wizard.go:182,222,277,309,346` | `defs.FilePerm` | no |
| `internal/cli/update.go:1004` | `defs.FilePerm` | no |

**Positive prior art already in-repo (build on, do not re-invent):**

- `internal/settings/yamlpatch/yamlpatch.go:202` calls `os.Chmod(tmpName, info.Mode().Perm())`
  before renaming — the canonical mode-preservation pattern.
- `internal/cli/update_deny_migration.go:95-100` reads the existing mode via `os.Stat` and writes
  with the preserved mode (not yet atomic, but the stat-then-preserve pattern is the right half).
- `internal/cli/update_disk_backup.go:39` declares `var diskBackupWriteFile = os.WriteFile` — an
  injectable test seam (the testability pattern the guard tests should mirror).

**Scope of this SPEC — forward-looking invariant, not historical repair.** This SPEC guarantees
that every FUTURE write into `.moai/config/**` and CLI persistence paths is atomic and
mode-preserving. The historical damage — files already narrowed to 0600 by Defect 1 on prior
section saves — is real and acknowledged, but explicitly out of this SPEC's repair scope: a
one-shot widening migration is deferred to a follow-up sibling SPEC
(`SPEC-CONFIG-MODE-MIGRATE-001`, not yet authored; see §C Out of Scope — mode-widening migration).
This slice extracts the parent's REQ-CTP-021/022/023/024/027 (the write-path invariant) so the
helper lands first; the migration builds on top of it once the helper is available.

## §B Goals

- Every write into `.moai/config/**` and every CLI persistence write goes through one shared,
  atomic, mode-preserving helper.
- The helper preserves the destination's existing mode across the rename, and creates new files at
  a sane default mode (never `os.CreateTemp`'s 0600, never an inlined `0o644` literal).
- No writer under `internal/config/` or `internal/cli/` reaches the config/settings persistence
  surface via bare `os.WriteFile` after the fix — a guard makes the invariant sticky.
- The fix is verifiable as part of `go test ./...`, not only package-local.

## §C Scope Boundary — the 3-way split of the parent

`SPEC-CONFIG-TIER-PERSIST-001` was an over-large 35-REQ Tier-M-exceeding SPEC. It is split three
ways per the 3-way split recommendation:

| Slice | Topic | Owning SPEC |
|-------|-------|-------------|
| (a) | config tier resolution, explicit-falsey-wins-tier semantics, local-tier reachability | still resident in `SPEC-CONFIG-TIER-PERSIST-001` |
| **(b)** | **atomic, mode-preserving writes for `.moai/config` + CLI persistence** | **`SPEC-CONFIG-ATOMIC-WRITE-001` (THIS SPEC)** |
| (c) | malformed-section writeback-as-defaults handling, gitignore merge, reinstall-loop, template-base-snapshot | still resident in the parent (and where applicable, sibling SPECs) |

**This SPEC owns slice (b) only.** The lineage is recorded in the parent's `## §L Split Branches`
section (added 2026-08-12).

### Out of Scope — tier resolution

- The 8-tier priority system, `MergeAll` zero-value skipping, `SrcLocal` ordering, and the
  typed-Loader / resolver disconnect are slice (a). They stay in the parent.
- The branch-guard local-tier opt-in (`CLAUDE.local.md` §22.9) reachability fix is slice (a).

### Out of Scope — malformed-section writeback

- The fourteen section loaders swallowing parse failures into compiled defaults, and the
  refuse-to-writeback-a-failed-section contract (parent REQ-CTP-019/020), are slice (c).

### Out of Scope — gitignore / reinstall / template-base-snapshot

- `MergeGitignoreFile` header parsing, dropped-pattern migration, and reinstall-loop detection are
  sibling-SPEC concerns (parent §D.5 and `SPEC-UPDATE-REINSTALL-LOOP-002`). The atomic-write
  helper this SPEC introduces MAY be consumed by those writers later, but remediating their call
  sites is deferred to those SPECs.
- `internal/cli/update/backup/restore.go`, `internal/cli/update/merge/merge.go`, and
  `internal/core/project/initializer.go` writes are out of scope here (they are entangled with
  data-survival / template-base-snapshot concerns). The helper is extracted with a public-enough
  API that those sibling SPECs can adopt it; this SPEC does not migrate them.

### Out of Scope — mode-widening migration

- Parent REQ-CTP-025 / REQ-CTP-026 (one-shot widening of any `.moai/config/sections/*.yaml` file
  whose mode is narrower than `defs.FilePerm` back to `defs.FilePerm`) are deferred to follow-up
  sibling SPEC **`SPEC-CONFIG-MODE-MIGRATE-001`** (not yet authored).
- **Rationale**: this SPEC ships the write-path invariant (the helper that guarantees all future
  writes are atomic and mode-preserving). The migration repairs historical damage (the population
  of already-narrowed files) and is independently shippable on top of the helper once it lands.
  Bundling both into one SPEC would re-bloat the slice the split was designed to shrink.
- **Dependency**: `SPEC-CONFIG-MODE-MIGRATE-001` SHOULD declare `depends_on: [SPEC-CONFIG-ATOMIC-WRITE-001]`
  so the migration's own writes route through the helper this SPEC introduces. The migration does
  not re-derive the helper; it consumes `atomicfile.Write` as a caller.

## §D Requirements (GEARS)

### D.1 Atomicity

- **REQ-CAW-001** — The atomic-write helper shall write file content via a temp file in the
  destination's directory followed by `os.Rename`, so that a crash during the write leaves either
  the complete new content or the complete prior content, never a partial write.

- **REQ-CAW-006** — **When** the write fails before the rename step, the atomic-write helper shall
  remove the temp file so that no orphan temp file survives the failed call.

### D.2 Mode preservation

- **REQ-CAW-002** — **When** the destination file already exists, the atomic-write helper shall
  set the temp file's permission bits to the destination's pre-existing mode (read via `os.Stat`)
  before renaming, so the result file's mode is identical to the pre-write mode.

- **REQ-CAW-003** — **When** the destination file does not yet exist, the atomic-write helper shall
  create it at the default mode supplied by the caller, or at `defs.FilePerm` when the caller
  supplies no default; the helper shall not create a new file at `os.CreateTemp`'s 0600 default.

- **REQ-CAW-005** — The atomic-write helper and its callers shall not inline a numeric mode literal
  (`0o600`, `0o644`, `0o750`, ...); new-file defaults shall reference `defs.FilePerm` (or a named
  `defs` constant) and existing-file modes shall be read via `os.Stat`.

### D.3 Shared helper + call-site remediation

- **REQ-CAW-004** — Every write into `.moai/config/**` and every CLI persistence write into
  settings/config artifacts (`internal/config/manager.go` `atomicWrite`/`saveSection`,
  `internal/cli/harness.go`, `internal/cli/update*.go` production writes) shall route through the
  single shared atomic-write helper.

- **REQ-CAW-004a** — The shared helper's call-site remediation shall cover at minimum:
  `internal/config/manager.go:420-438`, `internal/cli/harness.go:390`,
  `internal/cli/update_cleanup.go:243,258,446`, `internal/cli/update_preserve_inventory.go:350`,
  `internal/cli/update_wizard.go:182,222,277,309,346`, `internal/cli/update.go:1004`,
  `internal/config/toolpolicy/tier_render.go:146` (writes rendered `.claude/settings.json` — a
  settings/config persistence path with hardcoded `0o644`), and
  `internal/config/toolpolicy/codegen.go:244` (writes rendered settings JSON to a settings
  persistence path with hardcoded `0o644`). Writers in `internal/cli/update/backup/restore.go`,
  `internal/cli/update/merge/merge.go`, and `internal/core/project/initializer.go` are explicitly
  deferred to sibling slice (c) SPECs.

  **Classified out-of-scope sites** (verified live at HEAD `ed70e4354`, NOT subject to
  remediation in M3 because their target paths are NOT `.moai/config/**` or settings/config
  persistence paths):

  | File:line | Target path | Classification predicate |
  |-----------|-------------|--------------------------|
  | `internal/cli/update_recovery_manifest.go:78` | `backupDir/<recoveryManifestFileName>` | target path is a backup-directory artifact (`backupDir/...`), not `.moai/config/**` or a settings path |
  | `internal/cli/update_namespace_protect.go:242` | `backupDir/.complete` | target path is a backup-directory atomicity marker (`backupDir/.complete`), not a config or settings path |

### D.4 Regression guard

- **REQ-CAW-007** — A guard test shall assert that no non-test source file under `internal/config/`
  or `internal/cli/` writes to `.moai/config/**` or to settings/config persistence paths via a bare
  `os.WriteFile` call, so that a future writer cannot silently reintroduce the
  truncate-then-write window or the mode-narrowing path. Exemptions (e.g. the injected
  `diskBackupWriteFile` seam variable) shall be enumerated explicitly in the guard.

  **Target-path resolution algorithm** (mechanically implementable — the guard MUST decide
  violation-vs-exemption without human judgment at audit time):

  1. For each `os.WriteFile(...)` call expression found in a non-test `.go` file under
     `internal/config/` or `internal/cli/`, inspect the first argument (the path expression).
  2. **String-literal path**: if the first argument is a string literal (or a `filepath.Join` whose
     all components are string literals), resolve it to a literal string.
  3. **Variable path**: if the first argument is a variable (or parameter), trace the variable to
     its assignment within the same function; if the assignment's RHS is a string literal or a
     `filepath.Join` of literals, resolve it. If the variable is a function parameter, trace to the
     caller's argument at the call site within the same package (intra-package call-chain trace).
  4. Match the resolved literal against two target-sets:
     - **config-set**: any path matching `.moai/config/**`
     - **settings-set**: any path whose final component is `settings.json` or `settings.json.tmpl`
       (the Claude Code settings persistence artifacts written by `toolpolicy.SettingsPath` /
       `toolpolicy.TemplateSettingsPath`)
  5. **If the resolved path matches either set**: the call is a VIOLATION unless the call site is
     the helper itself (`atomicfile.Write`'s internal `os.CreateTemp`) or is enumerated in the
     guard's exemption list.
  6. **If the path cannot be resolved to a literal** (dynamic path constructed at runtime from
     user input, or cross-package parameter with no intra-package trace): the guard flags the call
     for manual review (advisory, not auto-fail) — the guard's exemption list MUST annotate each
     such site with its classification predicate.
  7. **Unresolved paths that are demonstrably non-config** (e.g. `backupDir/...`, `*.lock`,
     probe files) are exempted by predicate: "target path is under a backup directory" / "target
     path is a lock/probe/log artifact, not `.moai/config/**` or a settings path".

## §E Non-Functional Constraints

- **Cross-platform**: the temp-file + rename pattern must behave correctly on darwin/linux and
  must not break the `GOOS=windows GOARCH=amd64 go build ./...` build (rename across volumes is
  not a concern because the temp lives in the destination's own directory).
- **No new thresholds / no hardcoded modes** (CLAUDE.local.md §14): any default mode flows through
  `defs.FilePerm` or a named `defs` constant, never an inline literal.
- **Error wrapping** (CLAUDE.local.md §3): `fmt.Errorf("...: %w", err)`; never string concat.
- **Test isolation** (CLAUDE.local.md §6 [HARD]): every AC test uses `t.TempDir()`; no test writes
  to the project root or to `~/.claude`.
- **English code/comments/godoc** (CLAUDE.local.md §3, `code_comments: en`).
- **Subagent boundary**: this is a persistence-layer fix; no `AskUserQuestion` calls introduced
  into `internal/config/` or `internal/cli/` (C-HRA-008 family).

## §F Success Criteria

- `go test ./...` is green on darwin/linux; `GOOS=windows GOARCH=amd64 go build ./...` succeeds.
- The mode-preservation round-trip AC (AC-CAW-002) passes for modes 0600, 0644, and 0750.
- The temp-cleanup AC (AC-CAW-006) passes on both the success path and the error path.
- The guard test (AC-CAW-007) fails when a bare `os.WriteFile` to a config path is reintroduced
  (mutation-verified).
- Package coverage on `internal/config` and the touched `internal/cli` files is ≥ 85%.

## §G Risk Register

| Risk | Mitigation |
|------|------------|
| The shared helper's API (signature, default-mode parameter shape) is the most reversal-prone decision; getting it wrong forces every call site to change twice. | M1 locks the API before any call-site remediation; the API is reviewed at M1 sign-off before M2 begins. |
| Widening the guard to ALL `os.WriteFile` calls under `internal/cli/` produces false positives on legitimate non-config writes (lock files, probe files, log files). | The guard targets config/settings *persistence paths* (`.moai/config/**`, settings.json), not every `os.WriteFile`; exemptions are enumerated. |
| `internal/cli/update_cleanup.go:390` writes a probe file at `0o600` deliberately (a probe, not a config file). | Out-of-scope writes (probe, lock, log) are exempted by path, not by mode literal. |
| `internal/cli/update_disk_backup.go:39` uses an injectable `diskBackupWriteFile` seam for testing; the guard must not reject the seam. | The guard enumerates the seam declaration as an explicit exemption. |
| Renaming the parent's REQs loses traceability. | This SPEC's `related_specs: [SPEC-CONFIG-TIER-PERSIST-001]` and the parent's `## §L Split Branches` section carry the mapping: REQ-CTP-021/022/023/024/027 decompose into REQ-CAW-001..007; REQ-CTP-025/026 are deferred to `SPEC-CONFIG-MODE-MIGRATE-001`. |

## §H Cross-References

- `.moai/specs/SPEC-CONFIG-TIER-PERSIST-001/` — parent SPEC; §D.4 (REQ-CTP-021/022/023/024/027) is
  the source of this slice (5 parent REQs decompose into the 7 child REQ-CAW-001..007). The
  parent's `## §L Split Branches` section records the extraction, including the deferral of
  REQ-CTP-025/026 to `SPEC-CONFIG-MODE-MIGRATE-001`.
- `.moai/specs/SPEC-V3R5-ATOMIC-WRITE-001/` — **prior art**; `status: implemented`. Fixed a
  *different* `atomicWrite` (in `internal/migrate/hook_cleanup.go`) that was a plain
  `os.WriteFile` wrapper. Same atomic-write concept, different code path; this SPEC does NOT
  supersede it. Establishes the repo-wide expectation that "atomicWrite" must actually be atomic.
- `internal/settings/yamlpatch/yamlpatch.go:202` — the canonical `os.Chmod(tmpName,
  info.Mode().Perm())` before-rename pattern; the reference implementation the helper follows.
- `internal/cli/update_deny_migration.go:95-100` — the stat-then-preserve-mode pattern (half of the
  fix; this SPEC makes it atomic too).
- `internal/defs/perms.go:11` — `defs.FilePerm` (`0o644`), the named default-mode constant.
- `CLAUDE.local.md` §6 (Test Isolation — `t.TempDir()`), §14 (Hardcoding prevention).
