---
id: SPEC-CONFIG-DEAD-SWEEP-001
title: "Re-remove dead non-ralph config (research.yaml, state.state_dir) reverted by the build-recovery commit"
version: 0.2.0
status: in-progress
created: 2026-08-04
updated: 2026-08-14
author: manager-spec
priority: P1
phase: "v3.x target"
module: config-dead-sweep
lifecycle: spec-first
tags: "dead-config, research-yaml, state-dir, revert-recovery, template-neutrality, config-honesty"
tier: S
related_specs: [SPEC-RALPH-CONFIG-REDESIGN-001, SPEC-CONFIG-KEY-HONESTY-001, SPEC-WEBCONF-SIMPLIFY-001]
---

# SPEC-CONFIG-DEAD-SWEEP-001 — Re-remove dead non-ralph config (research/state_dir)

## HISTORY

- 2026-08-04 — Initial draft. Codifies the non-ralph half of the dead-config sweep identified during the ralph.yaml redesign investigation. ralph.yaml keys (including `stale_seconds`) are owned by `SPEC-RALPH-CONFIG-REDESIGN-001` — the two SPECs partition the dead-config space so run-phases do not collide on `ralph.yaml`. Cross-reference added in §H.
- 2026-08-14 — **v0.2.0 amendment: scope narrowed from three targets to two, after the first run landed and was partially reverted.**
  - **Land-then-revert discovered.** The v0.1.0 scope was implemented and merged as commit `4c88bbce9` (PR #1325, 26 files, −626 lines), then partially reverted by commit `7171880a9` (PR #1409, "feat: complete accumulated internal code — build recovery"). This SPEC is therefore a **re-run**, not a first run, and the working tree it targets is an asymmetric half-reverted state (§B.0).
  - **`cache.yaml` target DROPPED — REQ-CDS-003 retired.** `SPEC-WEBCONF-SIMPLIFY-001` M3 is a later decision that deliberately retains the baked template `cache.yaml` for runtime consumption (REQ-WC-003). Retirement rationale and the file:line carrying that decision are recorded at §C REQ-CDS-003 (RETIRED). AC-CDS-001, AC-CDS-002, and AC-CDS-006 are retired with it.
  - **Remaining scope narrowed** to `research.yaml` (local + template + its Go plumbing) and the `state_dir` key (template `state.yaml` + `StateConfig.StateDir` + `DefaultStateDir`).
  - **Line-number drift corrected** against the current tree: `StateDir` is at `types.go:682` (was cited as `:562`), its default-population at `defaults.go:742` (was cited as `:620`), the research loader at `loader.go:279` (was cited as `:277`).
  - **Three CI guards named** (§C REQ-CDS-009) with their measured baselines in `acceptance.md`.

## §A. User Story

**As a** MoAI user editing `.moai/config/sections/*.yaml` files expecting my edits to take effect,
**I want** the dead config files and keys that have ZERO Go runtime consumers removed from both the local project and the distributed template,
**so that** the configuration surface is honest — every YAML key a user can edit is actually read by running code, and the template no longer ships misleading knobs.

**Outcome hypotheses:**
- The verification-claim-integrity hazard ("users edit a file expecting effect") is eliminated for the two remaining dead surfaces.
- Template shrinks by 1 dead section file (`research.yaml`) and 1 dead key (`state.state_dir`).
- The `cfg.Research` field with its loader chain and the `cfg.State.StateDir` dead field are removed, reducing reader cognitive load.
- The half-reverted asymmetry of §B.0 is resolved — local and template stop disagreeing about whether these surfaces exist.

## §B. Context and Background

### §B.0 Land-then-revert history (this is a RE-RUN)

The v0.1.0 scope was **implemented and merged**, then **partially reverted**. Any acceptance criterion for this SPEC must therefore be able to distinguish "never done" from "done, then reverted" — a criterion that merely asserts absence would have passed between the two commits and tells us nothing about the current tree.

| Event | Commit | Effect |
|-------|--------|--------|
| First run landed | `4c88bbce9` (PR #1325, 2026-08-07) | Removed all three v0.1.0 targets across 26 files (−626 lines), local and template mirrors both. |
| Partial revert | `7171880a9` (PR #1409, "build recovery") | Restored template `cache.yaml`, template `research.yaml`, template `state.yaml`'s `state_dir` key, local `research.yaml`, and the Go plumbing (`cache_config.go`, `ResearchConfig`, `StateDir`). |

The revert was **asymmetric** — it did not restore everything:

| Surface | State after `7171880a9` |
|---------|-------------------------|
| Local `.moai/config/sections/cache.yaml` | still deleted (correctly) |
| Local `.moai/config/sections/state.yaml` | still cleaned to `state: {}`, carrying a comment attributing the removal to this SPEC |
| Local `.claude/rules/moai/core/settings-management.md` | still cleaned (no `research.yaml` row) |
| Local `.moai/config/sections/research.yaml` | **restored** |
| Template `cache.yaml` / `research.yaml` / `state.yaml:state_dir` | **restored** |
| Go plumbing (`ResearchConfig`, `StateDir`, `cache_config.go`) | **restored** |

The local tree therefore already *documents* a removal that the template contradicts. This asymmetry — not a fresh discovery of dead config — is the condition this amendment targets.

Because `7171880a9` was a broad build-recovery commit, the revert is read as collateral rather than a deliberate reversal of the sweep's rationale. That reading is what makes the re-run legitimate; the one target where a *deliberate* later decision exists is `cache.yaml` (§B.1), and it is dropped for exactly that reason.

`progress.md` §E.2/§E.3 were never populated for the first run — the run's evidence lives only in `4c88bbce9`'s commit body. The re-run's §E.2 must therefore establish its own baseline from the current tree and MUST NOT cite the first run's evidence as its own.

### §B.1 `cache.yaml` — DROPPED from scope (superseded by a later decision)

`internal/settings/sectionroute.go:84-90` records the superseding decision verbatim:

> `SPEC-WEBCONF-SIMPLIFY-001 M3: the 7 remaining former seam sections (harness, ralph, feedback, observability, security, handoff, cache) stay reclassified to RouteExcluded — their tabs are removed and their web write path is gone. They are absent from this map, so RouteForSection returns the zero value (RouteExcluded). Their config keys persist in the baked template YAML for runtime consumption (REQ-WC-003).`

Two points matter and are easy to get backwards:

- `cache` is **NOT** in the `sectionRoutes` map — it is `RouteExcluded` by zero value. It is therefore **not** a live seam write path, and this SPEC must not describe it as one.
- The retention rationale is **REQ-WC-003 baked-template runtime consumption**, not console editability. Deleting the template `cache.yaml` would violate REQ-WC-003, which is a *later* decision than this SPEC's v0.1.0 draft.

`cache.yaml` is consequently out of scope (§E), and REQ-CDS-003 is retired (§C).

### §B.2 `research.yaml` — zero production consumers outside `internal/config`

`internal/config/loader.go:279` (`loadResearchSection`, registered at `internal/config/slice.go:31`, called from `Loader.Load()` at `loader.go:77`) loads the file into `cfg.Research` via `researchFileWrapper` (`types.go:1361-1363`), but no production code outside `internal/config` itself reads the field. The section was removed from the web console surface by `SPEC-WEB-CONSOLE-012` as an unnamed section, there is no `SectionResearch` SectionID in `internal/settings/schema.go`, and `research` is NOT among the seven sections REQ-WC-003 retains — so §B.1's retention argument does not extend to it.

The full dead plumbing:

| Symbol | Location |
|--------|----------|
| `loadResearchSection` | `internal/config/loader.go:279` (+ call site `loader.go:77`) |
| slice registration | `internal/config/slice.go:31` |
| `researchFileWrapper` | `internal/config/types.go:1361-1363` (+ `resolver.go:805`) |
| `ResearchConfig` type | `internal/config/types.go:837-838` |
| `Config.Research` field | `internal/config/types.go:31` |
| `NewDefaultResearchConfig` | `internal/config/defaults.go:362-364` (+ call site `defaults.go:349`) |
| audit registry row | `internal/config/audit_registry.go:39` |
| doc row (template only) | `internal/template/templates/.claude/rules/moai/core/settings-management.md:101` |
| YAML files | `.moai/config/sections/research.yaml`, `internal/template/templates/.moai/config/sections/research.yaml` |

### §B.3 `state.state_dir` — dead field, hardcoded constant is the SSOT

`internal/config/types.go:682` carries `StateDir string yaml:"state_dir"`. A grep across `internal/`, `cmd/`, `pkg/` for any reader of the field value (`.State.StateDir`) returns ZERO non-test matches. The actual path resolution is:
- `internal/cli/state.go:210-211` — `findStateDir()` walks up the tree looking for the hardcoded literal `.moai/state/`.
- `internal/worktree/state_guard.go:25` — hardcodes `StateDirRel = ".moai/state"`.
- `internal/config/defaults.go:150` — `DefaultStateDir = ".moai/state"` constant, whose only reader is the dead-field population at `defaults.go:742`.

**Named false positive.** `goal.StateDir` at `internal/cli/goal.go:338` and `:522` is a **different symbol** — the package constant `internal/goal/state.go:17` (`".moai/state/goal"`). It is unrelated to `StateConfig.StateDir` and MUST NOT be touched. The same applies to `StateDirRel`, `ChainStateDir`, `detectStateDir`, `navigatorDetectStateDir`, `convergenceStateDir`, `routingStateDir`, and `ensureStateDir` — all independent symbols that merely share the substring.

**Asymmetry note.** The local `state.yaml` is already clean (`state: {}`); only the template still carries the key. The removal is therefore template-side plus Go-side.

**Design decision (unchanged from v0.1.0, see plan.md §F M2):** Option (a) — remove the key + field + default. Keep the hardcoded literal as SSOT.

### §B.4 Stale comment fixes (fold-in)

Two comments actively mislead readers and contradict the verified liveness of their targets:
- `internal/config/loader.go:345` marks `learning` as "Legacy sub-system (out-of-scope)" but it is LIVE — consumed at `internal/cli/hook.go:551-1106`.
- `internal/config/audit_registry.go:75` says observability="no Go loader yet" but `internal/config/observability_master.go:85` reads the `enabled` key live — should read "partial direct-read".

## §C. Requirements (GEARS)

### REQ-CDS-001 (Ubiquitous)
The MoAI config loader shall expose only configuration keys that have at least one live Go runtime consumer (read by non-test code in `internal/`, `cmd/`, or `pkg/`).

### REQ-CDS-002 (Ubiquitous)
The MoAI distributed template (`internal/template/templates/.moai/config/sections/`) shall not ship YAML section files whose entire content has zero runtime effect.

### REQ-CDS-003 — RETIRED 2026-08-14 (superseded by SPEC-WEBCONF-SIMPLIFY-001 REQ-WC-003)

> **Retired text (v0.1.0, for the record):** *When `LoadCacheConfig` has no caller and `cache.yaml` has no runtime consumer, the maintainer shall remove the `cache.yaml` file (local + template), the `LoadCacheConfig` function, the `CacheConfig` struct type if orphaned, and any loader registration, while preserving `ValidSessionTTLs()`.*

**Retirement reason.** `SPEC-WEBCONF-SIMPLIFY-001` M3 — a decision later than this SPEC's v0.1.0 draft — deliberately retains the baked template `cache.yaml` for runtime consumption. The retention is stated at **`internal/settings/sectionroute.go:88-89`**: *"Their config keys persist in the baked template YAML for runtime consumption (REQ-WC-003)."* Removing the template `cache.yaml` would violate REQ-WC-003.

**Binding consequence.** This SPEC's run-phase MUST NOT modify `internal/config/cache_config.go`, `internal/template/templates/.moai/config/sections/cache.yaml`, the `cache` entry in `acknowledgedDedicatedLoaders`, or the `cache.yaml` doc row in `settings-management.md`. The already-deleted **local** `.moai/config/sections/cache.yaml` stays deleted — restoring it is NOT in scope either (REQ-WC-003 binds the baked template, not the local project copy).

The retirement withdraws AC-CDS-001, AC-CDS-002, and AC-CDS-006 (`acceptance.md` §D.0).

### REQ-CDS-004 (Event-detected)
**When** `cfg.Research` has zero production readers outside `internal/config` itself, the maintainer shall remove the `research.yaml` file (local + template), the `loadResearchSection` loader (`internal/config/loader.go:279`) with its `Loader.Load()` call site (`loader.go:77`) and its `slice.go:31` registration, the `researchFileWrapper` (`types.go:1361`, `resolver.go:805`), the `ResearchConfig` type (`types.go:837`), the `Config.Research` field (`types.go:31`), the `NewDefaultResearchConfig` constructor (`defaults.go:362`) with its `defaults.go:349` call site, the `audit_registry.go:39` row, and the template `settings-management.md:101` doc row.

### REQ-CDS-005 (Event-detected)
**When** `cfg.State.StateDir` has zero production readers and the path is hardcoded as the literal `.moai/state/` in `findStateDir()` and `state_guard.go`, the maintainer shall remove the `state_dir` key from the **template** `state.yaml` (the local copy is already clean), the `StateConfig.StateDir` field at `internal/config/types.go:682`, its default-population at `internal/config/defaults.go:742`, and the now-orphaned `DefaultStateDir` constant at `internal/config/defaults.go:150`, while leaving the hardcoded literal as the single source of truth.

### REQ-CDS-005b (Unwanted behaviour)
The maintainer shall not modify `goal.StateDir` (`internal/goal/state.go:17`), `StateDirRel` (`internal/worktree/state_guard.go:25`), `ChainStateDir`, `detectStateDir`, `navigatorDetectStateDir`, `convergenceStateDir`, `routingStateDir`, or `ensureStateDir` — these are independent symbols that share the `StateDir` substring and are unrelated to the dead config field.

### REQ-CDS-006 (Ubiquitous)
The maintainer shall correct the two stale comments: `loader.go:345` (`learning` is LIVE, consumed at `cli/hook.go:551-1106`) and `audit_registry.go:75` (observability is "partial direct-read" via `observability_master.go:85`, not "no Go loader yet").

### REQ-CDS-007 (Capability gate)
**Where** the project is rebuilt after removal, `go build ./...` and `go test ./...` shall both exit 0, and `moai update` shall continue to merge the remaining sections correctly (RestoreMoaiConfig unaffected).

### REQ-CDS-008 (Capability gate)
**Where** `template-neutrality-check.yaml` CI guard runs, the removed template files shall not leave behind any forbidden content class (internal SPEC IDs, REQ tokens, commit SHAs) — either cleanly deleted or verified clean before deletion.

### REQ-CDS-009 (Capability gate)
**Where** the three config-audit CI guards run after the removal, each shall stay green **without** a new suppression entry:

| Guard | File | Reaction to this SPEC's changes |
|-------|------|-------------------------------|
| `TestAuditLoaderCompleteness` (`YAML_SECTION_NO_LOADER`) | `internal/config/audit_loader_completeness_test.go:56` | Enumerates the template sections directory. `research` is currently covered by `loaded["research"]`. Deleting the template file **and** the loader together keeps it green; deleting only one of the two breaks it (loader-only removal → `YAML_SECTION_NO_LOADER: research`). |
| `TestAuditRegistry_AllRegisteredStructsExist` | `internal/config/audit_registry_test.go:56` | Its `knownSections` list contains `"research"`; removing the `audit_registry.go:39` row without removing this list entry fails with *registry missing section "research"*. Both edits land together. |
| `TestStructYAMLSymmetry` (`CONFIG_STRUCT_YAML_MISMATCH`) | `internal/config/audit_struct_yaml_symmetry_test.go:32` | **Unaffected.** Neither `StateConfig` nor `ResearchConfig` is among the 7 `symmetryCases` (constitution, context, interview, design, statusline, git-convention, gate). The guard must stay green with its case count unchanged at 7. |

The maintainer shall NOT add `research` to `acknowledgedUnloadedSections` or to `yamlAuditExceptions` — a suppression entry would record the section as intentionally-unloaded-but-present, which is the opposite of the removal this SPEC performs.

### REQ-CDS-010 (Event-detected)
**When** the run-phase records its §E.2 evidence, it shall establish its own baseline from the current tree and shall not cite commit `4c88bbce9`'s evidence as its own, because that run's result was partially reverted by `7171880a9` (§B.0).

## §D. Constraints

- **Template-First:** edits to `internal/template/templates/.moai/config/sections/` happen in the template source first, then `make build` regenerates the embedded FS, then the local copy is synced.
- **No behavior change:** the user-visible behavior of `moai update`, `moai web`, and `findStateDir()` MUST be unchanged.
- **`cache.yaml` is frozen:** REQ-CDS-003 is retired; `cache_config.go` and the template `cache.yaml` are untouchable in this SPEC (REQ-WC-003).
- **`internal/settings` test fixtures stay:** `internal/settings/testdata/sections/research.yaml` and the tests that seed from it (`sectionwrite_test.go:181`, `schema_sections_test.go:445/602/657`) read from `testdata/`, not from the template, and are unaffected by the template deletion. They MUST NOT be deleted as part of this sweep — `TestWriteSectionViaSeamRejectsResearchPreservesFile` asserts the *rejection* behaviour that survives the section's removal.
- **Non-ralph scope only:** this SPEC MUST NOT touch `ralph.yaml` or any ralph.* key — that surface is owned by `SPEC-RALPH-CONFIG-REDESIGN-001`.

## §E. Out of Scope

### Out of Scope — cache.yaml and all cache config plumbing

- `internal/template/templates/.moai/config/sections/cache.yaml` — deliberately retained by `SPEC-WEBCONF-SIMPLIFY-001` REQ-WC-003 (`internal/settings/sectionroute.go:88-89`).
- `internal/config/cache_config.go` — `LoadCacheConfig`, the `CacheConfig` struct, and `ValidSessionTTLs()` are all left as-is.
- The `cache` entry in `acknowledgedDedicatedLoaders` and the `cache.yaml` row in `settings-management.md`.
- Restoring the already-deleted local `.moai/config/sections/cache.yaml` — also out of scope.

### Out of Scope — internal/settings research test fixtures

- `internal/settings/testdata/sections/research.yaml` and the seam-rejection tests that consume it stay unchanged.
- `sectionroute_test.go:38` (`"research": RouteExcluded`) stays — routing classification is independent of whether the section file ships.

### Out of Scope — ralph.yaml and all ralph.* keys

- `ralph.yaml` (local + template) — owned by `SPEC-RALPH-CONFIG-REDESIGN-001`.
- `ralph.stale_seconds`, `Session.StaleSeconds`, the stale_seconds injection pipeline — all ralph.yaml keys are out of scope here.

### Out of Scope — deeper state_dir redesign

- Wiring `findStateDir()` to read from `cfg.State.StateDir` (option (b) in the design decision) is out of scope; the smaller option (a) is chosen.

### Out of Scope — usage-log migration

- Migrating any consumer to read from `cfg.State.StateDir` if one is discovered post-removal is a follow-up concern, not part of this sweep.

### Out of Scope — other dead-config surfaces not listed

- Any dead config surface other than research.yaml and state.state_dir is out of scope (e.g., dead workflow.yaml keys, dead harness.yaml keys). Each deserves its own SPEC.

### Out of Scope — preventing a future revert

- Adding a regression guard that would make a future broad "build recovery" commit unable to silently restore these surfaces is a separate concern. This SPEC re-applies the removal; it does not build the ratchet that would keep it applied.

## §F. Acceptance Criteria Summary

See `acceptance.md` for the full AC matrix (AC-CDS-003 through AC-CDS-014, with AC-CDS-001/002/006 retired). Every live AC maps to a REQ above, is Given-When-Then binary-testable, and carries a measured baseline taken in this worktree.

## §G. Risks

- **Re-revert.** The dominant risk is not that the removal is wrong but that it is undone again by a broad recovery commit, as in §B.0. Mitigation within this SPEC: AC-CDS-011 pins the exact half-reverted starting state, so a run against an already-changed tree fails loudly instead of proceeding on a stale premise. Preventing the re-revert itself is out of scope (§E).
- **Partial application breaks a CI guard.** Removing the loader without the template file (or the registry row without the `knownSections` entry) turns a green guard red. Mitigation: REQ-CDS-009 names the coupled pairs; AC-CDS-009/AC-CDS-010 assert each guard green.
- **Suppression instead of removal.** The cheapest way to make a red guard green is to add `research` to an allowlist — which would preserve exactly the dishonest surface this SPEC removes. Mitigation: AC-CDS-012 asserts no new allowlist entry.
- **Hidden reader of state_dir:** a reader not caught by the grep would regress. Mitigation: AC-CDS-007 asserts `go build ./...` and `go test ./...` green post-removal.
- **Template-neutrality CI regression:** Mitigation: AC-CDS-013; deleted files trivially pass.

## §H. Cross-References

- **`SPEC-WEBCONF-SIMPLIFY-001`** — its M3 / REQ-WC-003 supersedes REQ-CDS-003 (retention stated at `internal/settings/sectionroute.go:88-89`). Any future attempt to re-open the `cache.yaml` target must first reconcile with that SPEC.
- **`SPEC-WEB-CONSOLE-012`** — removed `research` from the console surface as an unnamed section, which is why `research` is absent from the seven sections REQ-WC-003 retains.
- **`SPEC-RALPH-CONFIG-REDESIGN-001`** — owns the ralph half of this dead-config sweep. Run-phases of the two SPECs are partitioned: this SPEC touches only `research.yaml`/`state.yaml`; that SPEC touches only `ralph.yaml`. Run them in separate worktrees or sequence them to avoid PR-scope collision.
- **Commits `4c88bbce9` / `7171880a9`** — the land and the partial revert (§B.0).
- **`SPEC-CONFIG-KEY-HONESTY-001`** — predecessor; established the "every key must have a consumer" invariant this SPEC operationalizes for three named surfaces.
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 — the defect-claim hazard this SPEC addresses (users editing dead config expecting effect).
- `CLAUDE.local.md` §2 [HARD] Template-First Rule — governs the template edit → `make build` → local sync order.
- `CLAUDE.local.md` §25 Template Internal-Content Isolation — governs what the removed template files may carry.
