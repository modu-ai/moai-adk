---
id: SPEC-CONFIG-DEAD-SWEEP-001
title: "Remove dead non-ralph config (cache.yaml, research.yaml, state.state_dir) with zero Go runtime consumers"
version: 0.1.0
status: draft
created: 2026-08-04
updated: 2026-08-04
author: manager-spec
priority: P1
phase: "v3.x target"
module: config-dead-sweep
lifecycle: spec-first
tags: "dead-config, cache-yaml, research-yaml, state-dir, template-neutrality, config-honesty"
tier: S
related_specs: [SPEC-RALPH-CONFIG-REDESIGN-001, SPEC-CONFIG-KEY-HONESTY-001]
---

# SPEC-CONFIG-DEAD-SWEEP-001 — Remove dead non-ralph config (cache/research/state_dir)

## HISTORY

- 2026-08-04 — Initial draft. Codifies the non-ralph half of the dead-config sweep identified during the ralph.yaml redesign investigation. ralph.yaml keys (including `stale_seconds`) are owned by `SPEC-RALPH-CONFIG-REDESIGN-001` — the two SPECs partition the dead-config space so run-phases do not collide on `ralph.yaml`. Cross-reference added in §H.

## §A. User Story

**As a** MoAI user editing `.moai/config/sections/*.yaml` files expecting my edits to take effect,
**I want** the dead config files and keys that have ZERO Go runtime consumers removed from both the local project and the distributed template,
**so that** the configuration surface is honest — every YAML key a user can edit is actually read by running code, and the template no longer ships misleading knobs.

**Outcome hypotheses:**
- The verification-claim-integrity hazard ("users edit a file expecting effect") is eliminated for the three named dead surfaces.
- Template shrinks by 2 dead section files (`cache.yaml`, `research.yaml`) and 1 dead key (`state.state_dir`).
- `LoadCacheConfig` orphan code and the `cfg.Research`/`cfg.State.StateDir` dead fields are removed, reducing reader cognitive load and lint noise.

## §B. Context and Background

Investigation across `internal/`, `cmd/`, `pkg/` (grep, non-test) established three dead non-ralph config surfaces. Each is documented with its evidence below.

### §B.1 `cache.yaml` — orphaned loader

`internal/config/cache_config.go:15-29` carries an explicit `@MX:DEBT` marker:

> `LoadCacheConfig now has no caller. Only ValidSessionTTLs survives, consumed by the moai web settings seam for the session_ttl select options; cache.yaml itself round-trips through that seam without any code acting on its values.`

The cache_control injector that once consumed this file was removed (moai does not own the API call path; see `cache-aware-execution.md`). The file exists in BOTH:
- Local: `.moai/config/sections/cache.yaml`
- Template: `internal/template/templates/.moai/config/sections/cache.yaml`

**Live sub-surface:** `ValidSessionTTLs()` (cache_config.go:76) IS consumed by `internal/settings/schema_sections.go:328` for the web settings dropdown. This accessor MUST be preserved (or extracted to a standalone constant) — removing it would break the web settings seam. Everything else in cache_config.go (`LoadCacheConfig`, the `CacheConfig` struct, the file loader registration) is dead.

### §B.2 `research.yaml` — zero consumers

`internal/config/loader.go:277` loads the file into `cfg.Research` via a `researchFileWrapper`, but the `Research` field has ZERO runtime readers across Go and agent code (`budget_cap_tokens` / `target_score` are unused). Local + template files both exist.

### §B.3 `state.state_dir` — dead field, hardcoded constant is the SSOT

`internal/config/types.go:562` carries `StateDir string yaml:"state_dir"`. A grep across `internal/`, `cmd/`, `pkg/` for any reader of the field value (`cfg.State.StateDir` / `.State.StateDir`) returned ZERO matches. The actual path resolution is:
- `internal/cli/state.go:210-211` — `findStateDir()` walks up the tree looking for the hardcoded literal `.moai/state/`.
- `internal/worktree/state_guard.go:25` — hardcodes `StateDirRel=".moai/state"`.
- `internal/config/defaults.go:150` — `DefaultStateDir = ".moai/state"` constant (used only to populate the dead field at `defaults.go:620`).

**Design decision (see plan.md §F M2):** Option (a) — remove the `state_dir` key from `state.yaml` + the `cfg.State.StateDir` field + its default-population line. Keep the hardcoded literal in `findStateDir()`/`state_guard.go` as the SSOT. Tradeoff documented in plan.md.

### §B.4 Stale comment fixes (fold-in)

Two comments actively mislead readers and contradict the verified liveness of their targets:
- `internal/config/loader.go:345` marks `learning` as "Legacy sub-system (out-of-scope)" but it is LIVE — consumed at `internal/cli/hook.go:551-1106`.
- `internal/config/audit_registry.go:75` says observability="no Go loader yet" but `internal/config/observability_master.go:85` reads the `enabled` key live — should read "partial direct-read".

## §C. Requirements (GEARS)

### REQ-CDS-001 (Ubiquitous)
The MoAI config loader shall expose only configuration keys that have at least one live Go runtime consumer (read by non-test code in `internal/`, `cmd/`, or `pkg/`).

### REQ-CDS-002 (Ubiquitous)
The MoAI distributed template (`internal/template/templates/.moai/config/sections/`) shall not ship YAML section files whose entire content has zero runtime effect.

### REQ-CDS-003 (Event-detected)
**When** `LoadCacheConfig` has no caller and `cache.yaml` has no runtime consumer, the maintainer shall remove the `cache.yaml` file (local + template), the `LoadCacheConfig` function, the `CacheConfig` struct type if orphaned, and any loader registration, while preserving `ValidSessionTTLs()` (extracting it to a standalone constant if needed) because it is consumed by the web settings seam at `internal/settings/schema_sections.go:328`.

### REQ-CDS-004 (Event-detected)
**When** `cfg.Research` has zero runtime readers across Go and agent code, the maintainer shall remove the `research.yaml` file (local + template), the `Research` loader code at `internal/config/loader.go:277-284`, and the `Research` field/type if orphaned.

### REQ-CDS-005 (Event-detected)
**When** `cfg.State.StateDir` has zero runtime readers and the path is hardcoded as the literal `.moai/state/` in `findStateDir()` and `state_guard.go`, the maintainer shall remove the `state_dir` key from `state.yaml`, the `State.StateDir` field at `internal/config/types.go:562`, and its default-population at `internal/config/defaults.go:620`, while leaving the hardcoded literal as the single source of truth.

### REQ-CDS-006 (Ubiquitous)
The maintainer shall correct the two stale comments: `loader.go:345` (`learning` is LIVE, consumed at `cli/hook.go:551-1106`) and `audit_registry.go:75` (observability is "partial direct-read" via `observability_master.go:85`, not "no Go loader yet").

### REQ-CDS-007 (Capability gate)
**Where** the project is rebuilt after removal, `go build ./...` and `go test ./...` shall both exit 0, and `moai update` shall continue to merge the remaining sections correctly (RestoreMoaiConfig unaffected).

### REQ-CDS-008 (Capability gate)
**Where** `template-neutrality-check.yaml` CI guard runs, the removed template files shall not leave behind any forbidden content class (internal SPEC IDs, REQ tokens, commit SHAs) — either cleanly deleted or verified clean before deletion.

## §D. Constraints

- **Template-First:** edits to `internal/template/templates/.moai/config/sections/` happen in the template source first, then `make build` regenerates the embedded FS, then the local copy is synced.
- **No behavior change:** the user-visible behavior of `moai update`, `moai web`, the statusline cache-hit signal, and `findStateDir()` MUST be unchanged.
- **`ValidSessionTTLs` is load-bearing:** removing it would break the web settings seam — extraction (not deletion) is mandatory.
- **Non-ralph scope only:** this SPEC MUST NOT touch `ralph.yaml` or any ralph.* key — that surface is owned by `SPEC-RALPH-CONFIG-REDESIGN-001`.

## §E. Out of Scope

### Out of Scope — ralph.yaml and all ralph.* keys

- `ralph.yaml` (local + template) — owned by `SPEC-RALPH-CONFIG-REDESIGN-001`.
- `ralph.stale_seconds`, `Session.StaleSeconds`, the stale_seconds injection pipeline — all ralph.yaml keys are out of scope here.

### Out of Scope — deeper state_dir redesign

- Wiring `findStateDir()` to read from `cfg.State.StateDir` (option (b) in the design decision) is out of scope; the smaller option (a) is chosen.

### Out of Scope — usage-log migration

- Migrating any consumer to read from `cfg.State.StateDir` if one is discovered post-removal is a follow-up concern, not part of this sweep.

### Out of Scope — other dead-config surfaces not listed

- Any dead config surface other than cache.yaml, research.yaml, and state.state_dir is out of scope (e.g., dead workflow.yaml keys, dead harness.yaml keys). Each deserves its own SPEC.

## §F. Acceptance Criteria Summary

See `acceptance.md` for the full AC matrix (AC-CDS-001 through AC-CDS-010). Every AC maps to a REQ above and is Given-When-Then binary-testable.

## §G. Risks

- **ValidSessionTTLs breakage:** if the extraction is botched, the web settings dropdown empties. Mitigation: AC-CDS-006 asserts the seam still resolves the TTL list after extraction.
- **Hidden reader of state_dir:** a reader not caught by the grep (e.g., reflectively, or in a vendor path) would regress. Mitigation: AC-CDS-007 asserts `go build ./...` and `go test ./...` green post-removal.
- **Template-neutrality CI regression:** the removed template files might carry SPEC IDs in comments. Mitigation: AC-CDS-009 asserts the CI guard is green; deleted files trivially pass.

## §H. Cross-References

- **`SPEC-RALPH-CONFIG-REDESIGN-001`** — owns the ralph half of this dead-config sweep. Run-phases of the two SPECs are partitioned: this SPEC touches only `cache.yaml`/`research.yaml`/`state.yaml`; that SPEC touches only `ralph.yaml`. Run them in separate worktrees or sequence them to avoid PR-scope collision.
- **`SPEC-CONFIG-KEY-HONESTY-001`** — predecessor; established the "every key must have a consumer" invariant this SPEC operationalizes for three named surfaces.
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 — the defect-claim hazard this SPEC addresses (users editing dead config expecting effect).
- `CLAUDE.local.md` §2 [HARD] Template-First Rule — governs the template edit → `make build` → local sync order.
- `CLAUDE.local.md` §25 Template Internal-Content Isolation — governs what the removed template files may carry.
