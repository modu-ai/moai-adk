---
id: SPEC-V3-HOOKS-006
title: "Hook Scoping Hierarchy — Project vs User vs Plugin (3-tier)"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P1 High
phase: "v3.0.0 — Phase 2 Hook Protocol v2 Core"
module: "internal/hook/sources.go, internal/hook/registry.go"
dependencies:
  - SPEC-V3-HOOKS-001
  - SPEC-V3-HOOKS-004
  - SPEC-V3-SCH-002
related_gap:
  - gm#15
related_theme: "Theme 1: Hook Protocol v2 — Source Precedence"
breaking: true
bc_id: BC-003
lifecycle: spec-anchored
tags: "hook, v3, scoping, precedence, breaking"
---

# SPEC-V3-HOOKS-006: Hook Scoping Hierarchy

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-22 | GOOS | Initial v3 draft from Wave 4 bundle (Hooks/Commands) |

---

## 1. Goal (목적)

moai-adk today loads hook declarations from a single tier (`.claude/settings.json` project-level). Claude Code merges 8 tiers with strict precedence (policySettings → userSettings → projectSettings → localSettings → pluginHook → skillHook → sessionHook → builtinHook). Ad-hoc users need at minimum user-scoped and local-scoped hooks — shipping a 3-tier hierarchy (user / project / local) in v3.0 aligns moai with CC's foundational layering while deferring the full 8-tier policy+plugin+skill+session+builtin expansion to v3.2 (per master-v3 §9 open question #4).

With source precedence comes dedup across tiers: the same hook declaration appearing in multiple tiers is retained only once using SPEC-V3-HOOKS-004's dedup key. Outermost tier wins.

## 2. Scope (범위)

In-scope:
- Three hook source tiers in precedence order (outermost first): `user` (`~/.moai/hooks.yaml` OR `~/.claude/settings.json`), `project` (`.claude/settings.json`), `local` (`.claude/settings.local.json`).
- `internal/hook/sources.go` implements the merge pipeline.
- Dedup across tiers using `{shell}\0{command}\0{if}` key from SPEC-V3-HOOKS-004.
- Dedup policy: outermost tier (user) wins; inner-tier duplicates emit warn-level log indicating "shadowed by tier X".
- Array merge: per-event array of hook declarations concatenated across tiers after dedup.
- Map merge: for `matcher` grouping, maps merged key-by-key with per-key dedup.
- `moai doctor hook --sources` CLI subcommand surfacing loaded hooks with their source tiers.

Out-of-scope:
- Additional tiers (policy, plugin, skill, session, builtin) — deferred to v3.2.
- Per-hook once-bookkeeping source-tier awareness beyond the dedup key format (already covered in SPEC-V3-HOOKS-003).
- Settings-file layering for non-hook config sections (covered in SPEC-V3-SCH-002).
- Workspace trust gating.

## 3. Environment (환경)

Current moai-adk state:
- Single-tier loading: `.claude/settings.json` in project root is the only source (findings-wave1-moai-current.md §5.1 implicit via `internal/hook/registry.go`).
- No user-scoped hook file. No local-scoped override file.
- No dedup logic across sources (impossible with one source).

Claude Code reference:
- `utils/hooks/hooksSettings.ts:15-21` — HookSource priority enum (findings-wave1-hooks-commands.md §3.2).
- `utils/hooks/hooksSettings.ts:230-270` — merge pipeline (findings-wave1-hooks-commands.md §3.2).
- `utils/hooks.ts:1492-1566` — `getHooksConfig` composition order (findings-wave1-hooks-commands.md §3.3).
- `utils/hooks.ts:1712-1801` — cross-source dedup key computation (findings-wave1-hooks-commands.md §3.9).

Affected modules:
- `internal/hook/sources.go` — new file: tier loaders + merge pipeline.
- `internal/hook/registry.go` — consumes merged declarations instead of single-source list.
- `internal/config/settings/` — new loader for user-tier (`~/.moai/hooks.yaml` OR XDG-spec equivalent) and local-tier (`.claude/settings.local.json`).
- `internal/cli/doctor_hook.go` — new subcommand flag `--sources`.

## 4. Assumptions (가정)

- User-tier hook file location: `~/.moai/hooks.yaml` is canonical; `~/.claude/settings.json` is tolerated for CC compat but only its `hooks` section is consumed. XDG-Home-var override is honored via `os.UserHomeDir()`.
- Local-tier hook file location: `.claude/settings.local.json` is gitignored by default (matches CC convention).
- User-tier dominance: outermost tier wins on dedup (opposite of simple "last-wins" deep-merge used in SPEC-V3-SCH-002 for scalar fields, because hook declarations are behaviorally additive with explicit dedup).
- Null-byte separator in dedup key is safe because none of `{shell, command, if}` can contain raw `\0` on any supported platform.

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

- REQ-HOOKS-006-001: The system SHALL load hook declarations from exactly three tiers in v3.0: user, project, local.
- REQ-HOOKS-006-002: The precedence order (outermost first) SHALL be: user, project, local.
- REQ-HOOKS-006-003: The user-tier SHALL be loaded from `~/.moai/hooks.yaml` if present; else the `hooks` section of `~/.claude/settings.json` if present; else the user tier is empty.
- REQ-HOOKS-006-004: The project-tier SHALL be loaded from `.claude/settings.json` in the project root.
- REQ-HOOKS-006-005: The local-tier SHALL be loaded from `.claude/settings.local.json` in the project root.
- REQ-HOOKS-006-006: The system SHALL compute a dedup key per declaration using `{shell}\0{command}\0{if}` as defined in SPEC-V3-HOOKS-004.
- REQ-HOOKS-006-007: When the same dedup key appears in multiple tiers, the outermost tier's declaration SHALL be retained; inner-tier duplicates SHALL be dropped.
- REQ-HOOKS-006-008: The system SHALL annotate each loaded hook declaration with a `sourceTier` field ∈ {user, project, local} for observability and dedup reporting.

### 5.2 Event-driven Requirements

- REQ-HOOKS-006-010: WHEN a hook is shadowed by an outer tier, the loader SHALL emit a warn-level log `hook shadowed: key=... outer=user inner=project`.
- REQ-HOOKS-006-011: WHEN `moai doctor hook --sources` is invoked, the CLI SHALL print a table of every loaded hook with columns `event | matcher | if | sourceTier | dedupShadowed`.
- REQ-HOOKS-006-012: WHEN any of the three source files fail to parse (invalid YAML/JSON), the loader SHALL treat the failing tier as empty, emit an error log naming the file, and continue loading other tiers.

### 5.3 State-driven Requirements

- REQ-HOOKS-006-020: WHILE a hook's source tier is `local`, the loader SHALL mark the declaration as `volatile: true` so that `moai doctor` can distinguish per-machine hooks from shared project hooks.
- REQ-HOOKS-006-021: WHILE `.moai/config/sections/system.yaml` sets `hook.user_tier.enabled: false`, the loader SHALL skip the user-tier load entirely.

### 5.4 Optional Features

- REQ-HOOKS-006-030: WHERE the environment variable `MOAI_HOOKS_DEBUG=1` is set, the loader SHALL emit debug-level logs for each tier load including raw file path, number of declarations parsed, and number shadowed.
- REQ-HOOKS-006-031: WHERE a user-tier hook references `$HOME` or `~/`, the path SHALL be expanded using `os.UserHomeDir()` before subprocess exec.

### 5.5 Complex Requirements

- REQ-HOOKS-006-040: IF a project-tier hook uses `command: "$CLAUDE_PROJECT_DIR/script.sh"` AND the same dedup key exists in the user-tier with a different absolute path, THEN the user-tier path is used (outermost wins) and a warn log names both paths; ELSE the project-tier path is used.
- REQ-HOOKS-006-041: IF `hook.user_tier.enabled: false` AND a user-tier file exists, THEN the loader SHALL skip it silently in normal mode but log an informational message when `MOAI_HOOKS_DEBUG=1` is set; ELSE the user tier loads normally.

## 6. Acceptance Criteria (수용 기준 요약)

- AC-HOOKS-006-01: Given a hook declaration in both `~/.moai/hooks.yaml` (user) and `.claude/settings.json` (project) with matching `{shell, command, if}`, When loaded, Then only the user-tier instance is active and the project-tier instance is logged as shadowed. (maps REQ-HOOKS-006-007, REQ-HOOKS-006-010)
- AC-HOOKS-006-02: Given `moai doctor hook --sources`, When executed on a project with 3 user + 5 project + 2 local hooks (1 dedup across project/local), When output prints, Then the table shows 10 rows (3+5+2) with 1 marked `dedupShadowed=true`. (maps REQ-HOOKS-006-011)
- AC-HOOKS-006-03: Given a project-tier file with invalid JSON, When loaded, Then the tier is treated as empty, an error log names the file, and user-tier and local-tier still load. (maps REQ-HOOKS-006-012)
- AC-HOOKS-006-04: Given `hook.user_tier.enabled: false`, When loaded, Then no user-tier file is read and tier composition is `[project, local]` only. (maps REQ-HOOKS-006-021, REQ-HOOKS-006-041)
- AC-HOOKS-006-05: Given a hook in `.claude/settings.local.json`, When loaded, Then its `sourceTier == "local"` and `volatile == true`. (maps REQ-HOOKS-006-008, REQ-HOOKS-006-020)
- AC-HOOKS-006-06: Given a user-tier hook `command: "~/scripts/check.sh"`, When resolved, Then the path is expanded via `os.UserHomeDir()` prior to exec. (maps REQ-HOOKS-006-031)
- AC-HOOKS-006-07: Given a user-tier hook and a project-tier hook with identical dedup key but different commands, When loaded, Then the user-tier command is used and a warn log names both paths. (maps REQ-HOOKS-006-040)

## 7. Constraints (제약)

- Technical: Go 1.22+. YAML parse via existing `gopkg.in/yaml.v3`; JSON parse via stdlib `encoding/json`.
- Backward compat: Existing v2 single-source users see identical behavior when user-tier and local-tier files are absent. BC-003 documents this change.
- Platform: User-tier file path uses `os.UserHomeDir()` for cross-platform HOME resolution.
- Performance: Tier loading adds ≤ 20 ms at session start even with 100 hooks per tier.
- Security: Local-tier is per-machine and MUST NOT be checked into version control (gitignored by template default).

## 8. Risks & Mitigations (리스크 및 완화)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Users unaware of tier precedence discover unexpected shadowing | M | M | `moai doctor hook --sources` surfaces the full precedence view; docs-site migration guide explains with examples. |
| User-tier file leaks credentials across projects | L | H | Docs-site emphasizes project-tier for repo-specific hooks; user-tier should not contain secrets; `moai doctor` flags hooks with suspicious env-var references. |
| Local-tier gitignore missing in existing repos | M | L | `moai init` and `moai update` auto-append `.claude/settings.local.json` to `.gitignore` if absent. |
| Parse failure in one tier cascades to skipped dispatch | L | M | REQ-HOOKS-006-012 isolates tier load failures; `moai doctor` reports skipped tiers explicitly. |
| Deep-merge ambiguity between hook array semantics (this SPEC) and scalar fields (SPEC-V3-SCH-002) | L | M | Hooks use explicit dedup across tiers; non-hook scalars use last-wins deep-merge. Policy is documented in both SPECs' constraints. |

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3-HOOKS-001 (declaration shape).
- SPEC-V3-HOOKS-004 (dedup key format).
- SPEC-V3-SCH-002 (3-tier settings layering, broader than just hooks).

### 9.2 Blocks

- SPEC-V3-PLG-001 (plugin hooks become an additional tier in v3.2; this SPEC's 3-tier foundation is the anchor).

### 9.3 Related

- SPEC-V3-HOOKS-003 (once-hook bookkeeping key format must match dedup key here).
- SPEC-V3-HOOKS-005 (missing event handlers register across all three tiers uniformly).

## 10. Traceability (추적성)

- Theme: master-v3 Section 3.1 (Theme 1 — Hook Protocol v2) source-precedence sub-feature; master-v3 §9 open question #4 (3-tier scope reduction rationale).
- Gap rows: gm#15 (High — source precedence pipeline, scope-reduced from 8 tiers to 3 tiers in v3.0).
- BC-ID: BC-003 (hook settings source layering — flat → 3-tier).
- Wave 1 sources: findings-wave1-hooks-commands.md §3.2 (source precedence), §3.3 (loading pipeline), §3.9 (dedup), §12 (source references including `utils/hooks/hooksSettings.ts:15-21, 230-270` and `utils/hooks.ts:1712-1801`).
- Priority: P1 High (foundation for future plugin + policy tiers; ships Phase 2).
