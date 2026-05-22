---
id: SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001
title: "moai update User-Owned Namespace Protection + Backup Standardization — Plan"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli/update.go, internal/cli/update_archive.go, internal/defs/dirs.go"
lifecycle: spec-anchored
tags: "update, namespace, backup, harness, protection, contract, tier-m, plan"
tier: M
---

# Implementation Plan — SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001

## §1. Baseline (Source-Verified)

This baseline is established by direct read of the current `main` HEAD (`bac893173`) of `internal/cli/update.go`, `internal/cli/update_archive.go`, and `internal/defs/dirs.go`. Numbers below are line counts and line-anchor citations measured at plan-phase entry; manager-develop **shall re-measure** before run-phase M2 modification.

| Path | LOC | Relevant anchors |
|------|----:|------------------|
| `internal/cli/update.go` | 2723 | `isUserAreaPath` at line 1113; `isMoaiManaged` at line 1140; `backupMoaiConfig` at line 1335; `cleanup_old_backups` at line 1678 |
| `internal/cli/update_archive.go` | 354 | `archiveSkill` at line 55; `archiveLegacySkills` at line 245; `archiveVersion = "v2.16"` at line 27; drift backup directory at line 272 |
| `internal/cli/update_test.go` | 2757 | Existing test surface for update flow |
| `internal/cli/update_archive_test.go` | 281 | Existing archive backup tests |
| `internal/defs/dirs.go` | 108 | `BackupsDir = ".moai-backups"` at line 12; `MoAIDir = ".moai"` at line 6 |

### Existing protection inventory

1. **`isUserAreaPath`** (`update.go:1113-1128`): protects `.claude/skills/my-harness-*` and `.claude/agents/my-harness/`. **Does not cover** `.claude/agents/harness/` or `.moai/harness/`.
2. **`isMoaiManaged`** (`update.go:1140-1187`): classifies as MoAI-managed any path under `.moai/config/`, `.moai/evolution/`, `.claude/skills/moai-*`, `.claude/rules/moai/*`, `.claude/agents/{core,expert,meta,harness}/`, plus prefixed names. **CONTRADICTS** CLAUDE.local.md §24.4 on `.claude/agents/harness/`.
3. **`backupMoaiConfig`** (`update.go:1335-1452`): writes a timestamped backup to `.moai-backups/YYYYMMDD_HHMMSS/` covering only `.moai/config/`. Format is `YYYYMMDD_HHMMSS` (no `T`, no `Z`), distinct from the format proposed by REQ-UNP-010.
4. **`cleanup_old_backups`** (`update.go:1678-1726`): retains only the 5 most recent backups in `.moai-backups/`. Hardcoded `keepCount=5`. *Note: snake_case identifier is pre-existing legacy debt; all new functions introduced by this SPEC (`isUserOwnedNamespace`, `backupUserOwnedNamespace`, `assertNoUserOwnedNamespaceTouch`, `newNamespaceBackupStamp`) adhere to Go idiomatic camelCase.*
5. **Archive-drift backup** (`update_archive.go:272-289`): writes to `.moai/archive/skills/v2.16-drift-<UTC-ISO8601>/<id>/` when `--force` is passed and drift is detected. The existing `driftStamp` at `update_archive.go:251` uses compact format `20060102T150405Z` (no hyphens). The new namespace-protection backup (REQ-UNP-010) uses a distinct hyphenated ISO-8601 format `2026-05-23T14-30-00Z` for readability — see spec.md REQ-UNP-010 rationale.

### Key gap: contradictory `harness` classification

`update.go:1178` returns `true` (MoAI-managed) for `.claude/agents/harness/`, which would cause the directory to be deleted-and-reinstalled. CLAUDE.local.md §24.4 demands the opposite. **M2 must resolve this in favor of CLAUDE.local.md §24.4.**

## §2. Approach

The implementation introduces a new function `isUserOwnedNamespace(rel string) bool` distinct from the existing `isUserAreaPath`, then routes every destructive operation through both `isMoaiManaged` AND `isUserOwnedNamespace`. The new function returns `true` for the strict superset enumerated in REQ-UNP-001..003 plus REQ-UNP-009.

`isUserAreaPath` remains in place (NFR-UNP-005 additivity) and continues to be called from its existing call sites. `isUserOwnedNamespace` becomes the new authoritative check; `isMoaiManaged` is amended to remove `harness` from its `.claude/agents/{core,expert,meta,harness}` switch (line 1178).

The backup mechanism is implemented in a new file `internal/cli/update_namespace_protect.go` to keep the existing `update.go` boundary stable. The function `backupUserOwnedNamespace(projectRoot, isoStamp string) (string, error)` writes to `.moai/backups/update-<isoStamp>/`. The `cmdUpdate` flow is amended to call this function immediately after `backupMoaiConfig` (or to call the two in parallel where safe — single-thread for simplicity in M2).

Sentinel `UPDATE_USER_NAMESPACE_VIOLATION` is emitted by a new defensive check `assertNoUserOwnedNamespaceTouch(planedOps []deployOp) error` that runs before the existing template overlay write loop.

## §3. Milestones

Milestones are priority-ordered (no time estimates per `agent-common-protocol.md` §Time Estimation).

### M1 — Baseline measurement and decision lock-in (Priority: High)

**Owner**: manager-develop (delegated by orchestrator) or orchestrator-direct.

**Activities**:
- Re-read the five baseline files cited in §1 above and confirm line anchors are unchanged.
- Confirm `internal/cli/update_test.go` baseline test count via `go test -count=1 ./internal/cli/... -run TestUpdate -v 2>&1 | grep -c "^=== RUN"`.
- Confirm `internal/cli/update_archive_test.go` baseline via `go test -count=1 ./internal/cli/... -run TestArchive -v 2>&1 | grep -c "^=== RUN"`.
- Decide: keep `isUserAreaPath` or fold into `isUserOwnedNamespace` (NFR-UNP-005 says additivity, so keep both).
- Decide: parallel call to `backupUserOwnedNamespace` and `backupMoaiConfig`, or sequential. Sequential preferred for predictable stderr in M2.
- **Deliverable**: a `baseline.txt` artifact (or inline annotation in run-phase progress.md) documenting baseline LOC, test count, and the two decisions.

**Exit criteria**: AC-UNP-013 baseline portion passes (`make build` clean, `go test -race ./internal/cli/...` passes pre-modification).

### M2 — Namespace protection logic (Priority: High)

**Owner**: manager-develop (cycle_type=ddd, Section A-E REQUIRED).

**Activities**:
- Add `isUserOwnedNamespace(rel string) bool` to `internal/cli/update.go` immediately below `isUserAreaPath` (preserve adjacency for code review).
- Function body covers: `.claude/skills/my-harness-*`, `.claude/agents/harness/` (and its files), `.moai/harness/` (and contents), plus the user direct-added asset rules from REQ-UNP-009.
- Amend `isMoaiManaged` at line 1178 to remove `harness` from the `core, expert, meta, harness` switch.
- Update the function's godoc to cite CLAUDE.local.md §24.4 and this SPEC ID.
- Route every existing call site that currently uses `isUserAreaPath` to also check `isUserOwnedNamespace` (or replace; pick the simpler path during run-phase).
- Run `go test -race ./internal/cli/...` and `golangci-lint run ./internal/cli/...`.

**Exit criteria**: AC-UNP-001, AC-UNP-002, AC-UNP-003, AC-UNP-007, AC-UNP-008 pass.

### M3 — Backup mechanism standardization (Priority: High)

**Owner**: manager-develop (cycle_type=ddd, Section A-E REQUIRED).

**Activities**:
- Create `internal/cli/update_namespace_protect.go`.
- Implement `backupUserOwnedNamespace(projectRoot, isoStamp string) (backupDir string, err error)`. Returns empty string if no user-owned content exists.
- The function enumerates all paths matching `isUserOwnedNamespace`, copies them to `.moai/backups/update-<isoStamp>/<rel-path>`, preserving directory hierarchy. After the copy, writes a `.complete` marker file with the timestamp.
- Helper `newNamespaceBackupStamp() string`: returns `time.Now().UTC().Format("2006-01-02T15-04-05Z")`. Note the `Z` suffix and the hyphens (colon substitution).
- Add a new constant `NamespaceBackupsSubdir = "backups"` to `internal/defs/dirs.go` (so the path becomes `.moai/backups/update-<stamp>/`).
- Wire `cmdUpdate` flow: between the existing `backupMoaiConfig` call (already at `update.go:691`) and the deploy step, insert `backupUserOwnedNamespace`. Stderr emits a `tui.CheckLine` matching existing pattern.

**Exit criteria**: AC-UNP-004, AC-UNP-006, AC-UNP-010 pass.

### M4 — Sentinel + tests (Priority: High)

**Owner**: manager-develop (cycle_type=ddd, Section A-E REQUIRED).

**Activities**:
- Implement `assertNoUserOwnedNamespaceTouch(planedOps []deployOp) error` in `internal/cli/update_namespace_protect.go`. The function iterates the planned operation list, applies `isUserOwnedNamespace`, and returns `fmt.Errorf("UPDATE_USER_NAMESPACE_VIOLATION: would touch user-owned path: %s", rel)` on the first hit.
- Wire the assertion into `cmdUpdate` before any template overlay write begins.
- Create `internal/cli/update_namespace_protect_test.go` with table-driven scenarios (5+ cases):
  - `my-harness-*` skill directory preserved across update.
  - `.claude/agents/harness/` directory preserved across update.
  - `.moai/harness/` directory preserved across update.
  - User direct-added `.claude/agents/custom-agent.md` preserved (REQ-UNP-009).
  - User direct-added `.claude/skills/custom-skill/` preserved (REQ-UNP-009).
  - Violation simulation: planted target in the deploy plan triggers sentinel.
- Update existing tests in `internal/cli/update_test.go` and `internal/cli/update_archive_test.go` if their expectations conflict with the new behavior (in particular: any test that currently expects `.claude/agents/harness/` to be deleted MUST be updated).

**Exit criteria**: AC-UNP-005, AC-UNP-009, AC-UNP-011 pass.

### M5 — Cross-platform validation + run-phase chore (Priority: Medium)

**Owner**: manager-develop or orchestrator-direct.

**Activities**:
- Verify `go test -race ./internal/cli/...` PASS on darwin (host).
- Cross-build verification: `GOOS=linux GOARCH=amd64 go build ./cmd/moai`, `GOOS=windows GOARCH=amd64 go build ./cmd/moai`.
- If CI infrastructure available, ensure `Test ubuntu`, `Test macos`, `Test windows` workflows pick up the new test file.
- Update SPEC status from `draft` to `implemented`, version `0.1.0` to `0.2.0`, increment `updated:` date.
- Write `progress.md` summarizing M1-M5 evidence with AC matrix.

**Exit criteria**: AC-UNP-012, AC-UNP-013 pass.

## §4. Technical Approach Detail

### §4.1 New `isUserOwnedNamespace` (proposed body sketch)

```go
// isUserOwnedNamespace returns true when the relative project path belongs
// to a user-owned namespace per CLAUDE.local.md §24.4 contract and SPEC
// SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 REQ-UNP-001..003 + REQ-UNP-009.
//
// Strict superset of isUserAreaPath. Additive: existing call sites of
// isUserAreaPath are not removed.
//
// Protected patterns:
//   - .claude/skills/my-harness-*    (REQ-UNP-001)
//   - .claude/agents/harness/        (REQ-UNP-002 — overrides isMoaiManaged)
//   - .moai/harness/                  (REQ-UNP-003)
//   - .claude/skills/<custom>/        when prefix != "moai-" and name != "moai" (REQ-UNP-009)
//   - .claude/agents/<custom>.md      when neither under {core,expert,meta,harness} nor prefixed "moai-" (REQ-UNP-009)
func isUserOwnedNamespace(rel string) bool {
    norm := strings.ReplaceAll(rel, "\\", "/")
    // REQ-UNP-001
    if strings.HasPrefix(norm, ".claude/skills/my-harness-") { return true }
    // REQ-UNP-002 — explicit override of isMoaiManaged for this prefix
    if norm == ".claude/agents/harness" || strings.HasPrefix(norm, ".claude/agents/harness/") { return true }
    // REQ-UNP-003
    if norm == ".moai/harness" || strings.HasPrefix(norm, ".moai/harness/") { return true }
    // REQ-UNP-009 user direct-added skill dirs
    if strings.HasPrefix(norm, ".claude/skills/") {
        rest := strings.TrimPrefix(norm, ".claude/skills/")
        seg := strings.SplitN(rest, "/", 2)[0]
        if seg != "" && seg != "moai" && !strings.HasPrefix(seg, "moai-") { return true }
    }
    // REQ-UNP-009 user direct-added agent files
    if strings.HasPrefix(norm, ".claude/agents/") {
        rest := strings.TrimPrefix(norm, ".claude/agents/")
        seg := strings.SplitN(rest, "/", 2)[0]
        switch seg {
        case "core", "expert", "meta":
            return false  // MoAI system agents
        case "harness":
            return true   // REQ-UNP-002 (also covered above)
        }
        if !strings.HasPrefix(seg, "moai-") && !strings.HasPrefix(seg, "manager-") &&
            !strings.HasPrefix(seg, "expert-") && !strings.HasPrefix(seg, "builder-") &&
            !strings.HasPrefix(seg, "evaluator-") {
            return true
        }
    }
    return false
}
```

Note: the prefix list for system-agent filenames (`manager-`, `expert-`, `builder-`, `evaluator-`) must be confirmed during M1 by reading the current contents of `internal/template/templates/.claude/agents/` to ensure the list matches actually-shipped agent names.

### §4.2 Backup directory structure

```
.moai/
  backups/                              # NEW (this SPEC)
    update-2026-05-23T14-30-00Z/        # ISO-8601 UTC, colons → hyphens
      .complete                          # Marker (REQ-UNP-007)
      .claude/
        skills/
          my-harness-test/SKILL.md
        agents/
          harness/test-specialist.md
      .moai/
        harness/main.md
.moai-backups/                          # PRE-EXISTING (.moai/config/ only)
  20260523_143000/...
.moai/archive/skills/                   # PRE-EXISTING (legacy skill archive)
  v2.16/...
  v2.16-drift-2026-05-23T14-30-00Z/...  # PRE-EXISTING (drift backup)
```

Three distinct backup roots, three distinct concerns. **No consolidation in this SPEC.**

### §4.3 Sentinel string discipline

Sentinel `UPDATE_USER_NAMESPACE_VIOLATION` is emitted via `fmt.Errorf` and propagated up the `cmdUpdate` return chain. The exact string is verified by `acceptance.md` AC-UNP-005 via grep. No localization (per `language.yaml` `error_messages: en`).

### §4.4 Risks

| Risk | Mitigation |
|------|------------|
| R1: `isMoaiManaged` removal of `harness` breaks existing tests that asserted `harness/` was managed | M4 mandates updating those tests as part of the same change. |
| R2: New `.moai/backups/` directory not in `.gitignore` of generated projects | M3 should also patch `internal/template/templates/.gitignore` or document an addition. (Out-of-scope candidate — flag during M3.) |
| R3: Concurrent `moai update` runs collide on identical ISO-second timestamp | NFR-UNP-004 mitigates: numeric suffix `-1`, `-2`, or skip-if-byte-identical. |
| R4: User has authored a path under `.claude/agents/` that begins with `moai-` (false positive: classified as MoAI but actually user-owned) | Documented edge case; resolution deferred to follow-up SPEC. M1 confirms via grep that no real-world deployment relies on this anti-pattern. |
| R5: cross-platform path separator normalization mismatch on Windows | Existing NFR-UNP-003 mandate via `strings.ReplaceAll(rel, "\\", "/")`, matching prior precedent at `update.go:1115`. |
| R6: Sibling SPEC `SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001` (currently `draft` on main) modifies the same `cmdUpdate` flow region | **Decision (2026-05-23, audit-fix iter 2)**: This SPEC modifies `isMoaiManaged` at `update.go:1178` (~lines 1178-1180) + adds new functions `isUserOwnedNamespace` / `backupUserOwnedNamespace` (new file or appended to `update.go`). ARCHIVE-CONTRACT-001 modifies archive backup logic in `update_archive.go` + `--force` propagation in `update.go`. The two scopes are **disjoint** (different line ranges, different file regions). No `depends_on` required; **parallel-safe**. Either SPEC may land first; whichever lands second adjusts only its own touched region. |

## §5. Verification Strategy

- **Static**: `golangci-lint run ./internal/cli/...` clean post-M2, post-M3, post-M4.
- **Unit**: `go test -race ./internal/cli/...` clean, new test file `internal/cli/update_namespace_protect_test.go` 5+ table cases PASS.
- **Integration**: smoke test `moai update --dry-run` on a synthetic fixture containing all three user-owned categories — visible "preserve" reporting in stdout.
- **Cross-platform**: `GOOS=linux/windows GOARCH=amd64 go build ./cmd/moai` clean.
- **Sentinel**: grep `UPDATE_USER_NAMESPACE_VIOLATION` in stderr of a planted-violation test invocation; exit code != 0.

## §6. Dependencies

- None blocking — this SPEC may proceed on `main` HEAD as of `bac893173`.
- Sibling SPEC `SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001` (currently `draft`): if landed first, this SPEC rebases. See R6.
- B9 prohibition active throughout run-phase: manager-develop must not run `git pull/fetch/rebase` autonomously.

## §7. Run-Phase Delegation Note

The orchestrator's manager-develop delegation prompt for M2-M4 **must include**:

- Section A: Context (this SPEC ID, baseline LOC, CLAUDE.local.md §24.4 quote).
- Section B: Operating constraints (B9 prohibition, no `git pull/fetch/rebase`, no autonomous commit body editing).
- Section C: Mandatory inputs (the §4 sketches, the AC matrix from `acceptance.md`).
- Section D: Working tree discipline (PRESERVE list, no drive-by refactors per Agent Core Behavior 5).
- Section E: Completion evidence (test output, lint output, cross-platform build output, sentinel grep).

### §7.1 Conventional Commits per Milestone (audit-fix iter 2)

Each milestone deliverable **shall** use the following Conventional Commits prefix in the commit message title:

| Milestone | Commit prefix (title) |
|-----------|------------------------|
| M1 | `chore(SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001): M1 baseline measurement` |
| M2 | `feat(SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001): M2 isUserOwnedNamespace + isMoaiManaged 정정` |
| M3 | `feat(SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001): M3 backupUserOwnedNamespace + atomicity .complete marker` |
| M4 | `test(SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001): M4 namespace violation sentinel + 5+ table cases` |
| M5 | `chore(SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001): M5 mark implemented v0.2.0 + cross-platform validation` |

Commit body **shall** include `🗿 MoAI <email@mo.ai.kr>` footer per CLAUDE.local.md §4 (git_commit_messages: ko convention).

## §8. Out-of-Scope Items (cross-reference with spec.md §4 Out of Scope)

See `spec.md` §4 Out of Scope. The same enumeration applies. No additional out-of-scope items at the plan level.

### §8 Out of Scope

- Backup retention/cleanup policy for `.moai/backups/update-*/` directories (deferred to follow-up SPEC; existing `cleanup_old_backups` covers only `.moai-backups/`)
- New CLI flags including `--no-backup`, `--namespace-strict`, `--backup-dir <path>` (REQ-UNP-008 is future-proof EARS but flag not introduced in this SPEC)
- Sync mechanism overhaul beyond namespace protection and backup standardization
- Sequential Thinking deprecation (handled by separate `SPEC-V3R6-SEQ-THINKING-RETIRE-001`)
- Migration of existing `.moai-backups/` content to the new `.moai/backups/` layout (two roots coexist)
- Archive-drift backup mechanism (`.moai/archive/skills/v2.16-drift-*/`) — distinct concern handled by `SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001`

Canonical list: `spec.md` §4 Out of Scope.
