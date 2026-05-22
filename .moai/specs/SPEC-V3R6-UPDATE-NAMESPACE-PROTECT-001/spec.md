---
id: SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001
title: "moai update User-Owned Namespace Protection + Backup Standardization"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli/update.go, internal/cli/update_archive.go, internal/defs/dirs.go"
lifecycle: spec-anchored
tags: "update, namespace, backup, harness, protection, contract, tier-m"
tier: M
---

# SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 — `moai update` User-Owned Namespace Protection + Backup Standardization

## §1. HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-23 | manager-spec | Initial draft. Tier M, 3-artifact SPEC. Codifies CLAUDE.local.md §24.4 "moai update Contract" into Go implementation. |

## §2. Motivation

CLAUDE.local.md §24.4 (commit `03a586b79`, 2026-05-23) introduced a written contract for `moai update` namespace handling that distinguishes **moai-managed** (overwritable, no backup) from **user-owned** (must preserve, must back up). The contract enumerates six namespace categories with binary "delete" / "preserve" verdicts and binary "backup" / "no backup" verdicts.

The current Go implementation in `internal/cli/update.go` partially implements this contract:

| Contract row | Current implementation | Gap |
|--------------|------------------------|-----|
| `moai-*` skills overwrite | `isMoaiManaged` at line 1140 covers `.claude/skills/moai-*` and `.claude/rules/moai/` correctly | None |
| `my-harness-*` skills preserve | `isUserAreaPath` at line 1113 covers `.claude/skills/my-harness-*` correctly | No user-namespace backup created |
| `.claude/agents/{core,expert,meta}/` overwrite | `isMoaiManaged` at line 1178 covers these correctly | None |
| `.claude/agents/harness/` preserve | **CONTRADICTION**: `isMoaiManaged` at line 1178 classifies `harness` as MoAI-managed, but §24.4 mandates preserve | High-risk policy contradiction |
| `.moai/harness/` preserve | **NO PROTECTION**: no function references this path | Total absence of protection |
| User direct-added assets preserve | Partial via `isMoaiManaged` returning false for non-prefixed names | No explicit backup |

The contract also calls for backup-and-preserve for the three user-owned categories, but the current backup mechanism (`backupMoaiConfig` in `update.go:1335` writes to `.moai-backups/YYYYMMDD_HHMMSS/`) only covers `.moai/config/`. There is no general user-namespace backup mechanism.

This SPEC closes those gaps and standardizes the backup directory layout.

## §3. Goals

1. Resolve the `.claude/agents/harness/` classification contradiction in favor of CLAUDE.local.md §24.4 (preserve, user-owned).
2. Extend protection to `.moai/harness/` directory (currently unprotected).
3. Introduce a general user-owned namespace backup mechanism distinct from the existing `.moai/config/` backup and from the existing archive-drift backup.
4. Standardize the new backup directory layout at `.moai/backups/update-{YYYY-MM-DDTHH-MM-SSZ}/` (ISO-8601 UTC).
5. Introduce an abort sentinel `UPDATE_USER_NAMESPACE_VIOLATION` that fires when a destructive operation would touch a user-owned artifact, before any modification.

## §4. Non-Goals

See `### §4 Out of Scope` below for the full enumeration.

### §4 Out of Scope

- **Backup retention/cleanup policy** for `.moai/backups/update-*/` directories (7-day pruning, max-N rolling window, etc.). The existing `cleanup_old_backups` in `update.go:1678` covers only the `.moai-backups/` location with a hardcoded `keepCount=5`; mirroring or generalizing that policy to the new `.moai/backups/update-*/` directory is deferred to a follow-up SPEC.
- **New CLI flags** including `--no-backup`, `--namespace-strict`, `--backup-dir <path>`. The current SPEC defines `REQ-UNP-008` as a future-proof EARS requirement, but the flag itself is not introduced.
- **Sync mechanism overhaul**: other improvements to the `moai update` deploy/restore/merge pipeline beyond namespace protection and backup standardization.
- **Sequential Thinking deprecation**: handled by separate `SPEC-V3R6-SEQ-THINKING-RETIRE-001`.
- **Migration of existing `.moai-backups/` backups** to the new `.moai/backups/` layout. The two backup roots coexist after this SPEC; consolidation is deferred to a follow-up SPEC.
- **Archive-drift backup mechanism** (`.moai/archive/skills/v2.16-drift-*/` in `update_archive.go`): this is a distinct concern handled by `SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001`. The two backup mechanisms remain in separate directory trees with separate ownership.

## §5. Functional Requirements (EARS)

### REQ-UNP-001 — `.claude/skills/my-harness-*` preservation (Ubiquitous)

The system **shall** preserve all directories matching the pattern `.claude/skills/my-harness-*/` and their contents across every `moai update` invocation. Preservation means the directory and every file within it must remain unchanged in content, mode, and existence after the update completes.

### REQ-UNP-002 — `.claude/agents/harness/` preservation (Ubiquitous)

The system **shall** preserve the directory `.claude/agents/harness/` and all files contained within it across every `moai update` invocation. This requirement resolves the existing classification contradiction with `isMoaiManaged` at `update.go:1178` in favor of CLAUDE.local.md §24.4.

### REQ-UNP-003 — `.moai/harness/` preservation (Ubiquitous)

The system **shall** preserve the directory `.moai/harness/` (including `main.md`, `interview-results.md`, extension subdirectories, and any user-authored files) across every `moai update` invocation.

### REQ-UNP-004 — User-owned namespace backup creation (Ubiquitous)

Before any destructive operation runs during `moai update`, the system **shall** create a backup directory at `.moai/backups/update-{YYYY-MM-DDTHH-MM-SSZ}/` containing a copy of every existing user-owned artifact enumerated in REQ-UNP-001, REQ-UNP-002, and REQ-UNP-003. If none of those paths exist, no backup directory is created.

### REQ-UNP-005 — Sync-target exclusion (Event-Driven)

When the `moai update` sync-target enumeration encounters a path that satisfies the user-owned namespace check (the union of REQ-UNP-001, REQ-UNP-002, REQ-UNP-003, and REQ-UNP-009), the system **shall** add the path to a protected list and exclude it from every subsequent delete, overwrite, and merge operation in the same `moai update` invocation.

### REQ-UNP-006 — Pre-modification abort sentinel (Unwanted)

If a destructive operation (delete, overwrite, merge) is about to be applied to a path that satisfies the user-owned namespace check, **then** the system **shall** abort with stderr output containing the literal sentinel string `UPDATE_USER_NAMESPACE_VIOLATION` and exit with a non-zero status code, before any file modification occurs.

### REQ-UNP-007 — Backup atomicity (State-Driven)

While backup creation is in progress, the system **shall** block all destructive operations until the backup completes and a backup-success marker (file `.moai/backups/update-{ISO-DATE}/.complete`) is written.

### REQ-UNP-008 — `--no-backup` flag warning (Optional, future-proof) — *deferred, no AC in this SPEC*

> **Deferred Requirement**: The `--no-backup` CLI flag is **not implemented in this SPEC** (see §4 Out of Scope). REQ-UNP-008 is documented as a future-proof EARS requirement to anchor the policy invariant for the follow-up SPEC. The traceability matrix in `acceptance.md` marks this REQ as N/A intentionally.

Where a future `--no-backup` flag is added to `moai update`, the system **shall** emit a stderr warning at update start (e.g., `WARN: --no-backup specified; user-owned namespaces are still protected but no rollback artifact will be created`) and continue to enforce REQ-UNP-001, REQ-UNP-002, REQ-UNP-003, and REQ-UNP-006. Backup skip **shall not** imply protection skip.

### REQ-UNP-009 — User direct-added asset preservation (Ubiquitous)

The system **shall** preserve user direct-added assets across every `moai update` invocation. User direct-added assets are defined as:

- Paths under `.claude/agents/` (whether file or sub-directory) whose first segment immediately after `agents/` is none of `core`, `expert`, `meta`, `harness`, and whose filename does not begin with `moai-`.
- Directories under `.claude/skills/` whose name does not begin with `moai-` and is not equal to `moai`.

Preservation for REQ-UNP-009 means no delete and no overwrite; backup is **optional** for these paths and is omitted when no template-side conflict is detected.

### REQ-UNP-010 — Backup directory naming convention (Ubiquitous)

The system **shall** name every backup directory created by REQ-UNP-004 using the regex `^\.moai/backups/update-\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}Z$`. The timestamp segment uses ISO-8601 UTC with colons replaced by hyphens for Windows-safe filenames. Note: the existing `driftStamp` at `update_archive.go:251` uses a different compact format `20060102T150405Z` (no separators); this SPEC **intentionally adopts the hyphenated ISO-8601 form** for human readability and explicit date-time delimitation in the new `.moai/backups/` root. The two formats are distinct conventions, not a single shared format.

## §6. Non-Functional Requirements

| Code | Requirement |
|------|-------------|
| NFR-UNP-001 | Backup creation latency: **shall** complete in under 5 seconds for a typical user-owned namespace footprint (10 files, 1 MB total). |
| NFR-UNP-002 | Backup directory permissions: directory **shall** be created with mode `0o755`, files with mode `0o644` (matching existing `defs.DirPerm` / `defs.FilePerm`). |
| NFR-UNP-003 | Cross-platform: protection logic **shall** normalize backslashes to forward slashes (matching existing `strings.ReplaceAll(rel, "\\", "/")` at `update.go:1115`) and work on darwin, linux, and windows. |
| NFR-UNP-004 | Idempotency: invoking `moai update` twice in quick succession **shall not** create duplicate backups for the same second; if a backup directory with the same ISO timestamp already exists, the second invocation **shall** append a numeric suffix `-1`, `-2`, ... or skip the backup if the destination is byte-identical. |
| NFR-UNP-005 | The new protection function **shall** be additive — `isUserAreaPath` (existing) keeps working as-is, and the new function (proposed name `isUserOwnedNamespace`) **shall** return `true` for the strict superset of paths previously matched by `isUserAreaPath`. |

## §7. Constitutional Constraints

- **B9 prohibition** (no autonomous `git pull/fetch/rebase` by manager-develop during run-phase): explicitly cited in `plan.md` Section D.
- **Tier M canonical 12-field frontmatter** per `.claude/rules/moai/development/spec-frontmatter-schema.md`.
- **Out of Scope section** required as `### §X Out of Scope` h3 heading (per `internal/spec/lint.go:704` `OutOfScopeRule`).
- **Template-First Rule**: any rule documents added by this SPEC must mirror to `internal/template/templates/...`. This SPEC does not propose new rule documents; it modifies Go code only.
- **manager-develop-prompt-template Section A-E required** for Tier M run-phase delegation.

## §8. Cross-References

- `CLAUDE.local.md` §24 — `moai update` Contract (canonical policy, §24.4 namespace table)
- `.claude/rules/moai/development/skill-authoring.md` § Skills Namespace Policy
- `.claude/rules/moai/development/agent-authoring.md` § Agent Directory Convention
- `.claude/skills/moai-meta-harness/SKILL.md` § Namespace Separation
- chore commit `03a586b79` (CLAUDE.local.md §24.4 namespace policy)
- chore commit `d0782a365` (CI guard pattern reference)
- chore commit `4f1135684` (template pollution cleanup precedent)
- chore commit `d065fba95` (Tier M plan-phase precedent)
- Sibling SPEC `SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001` (archive backup, distinct directory tree)
- Existing SPEC `SPEC-V3R3-HARNESS-001` (REQ-HARNESS-004, originates `isUserAreaPath`)
- Existing SPEC `SPEC-V3R6-AGENT-FOLDER-SPLIT-001` (originates `core/expert/meta/harness` classification at `update.go:1178`)
