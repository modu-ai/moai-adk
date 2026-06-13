---
id: SPEC-GITSTRATEGY-SAVE-ISOLATION-001
title: "Restore git_strategy section isolation in ConfigManager.Save() (regression from SAVE-WIRING M1)"
version: "0.1.0"
status: implemented
created: 2026-06-13
updated: 2026-06-13
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/config"
lifecycle: spec-anchored
tags: "config, save, git-strategy, isolation, regression, web-console, dirty-tracking"
era: V3R6
tier: S
---

# SPEC-GITSTRATEGY-SAVE-ISOLATION-001 — Restore git_strategy section isolation in Save()

## HISTORY

- 2026-06-13 (draft): Plan-phase artifacts authored. Regression fix for GitHub Issue #1064 — a scoped project-config write (`internal/web` writeProjectConfig: quality + git_convention only) mutates the unrelated `git_strategy` section because `ConfigManager.Save()` unconditionally re-serializes the full default git_strategy tree, obliterating any on-disk content the loader did not model (e.g. a sentinel). Regression introduced by `33215af27` (SPEC-PREPUSH-SAVE-WIRING-001 M1, "wire git_strategy into Save() WRITE path"). The guard test `TestWriteProjectConfigSectionIsolation` (added in `e83864047`, SPEC-WEB-CONSOLE-003 M5) is currently RED on `main`.

## §A. Context

### A.1 The defect (Issue #1064)

`go test ./internal/web/ -run TestWriteProjectConfigSectionIsolation` FAILS on `main`. The test (`internal/web/projectconfig_scope_test.go`) seeds `git-strategy.yaml` with a sentinel — `git_strategy:\n  sentinel: DO_NOT_TOUCH\n` — and asserts that a project-config write (`writeProjectConfig`, scoped to the quality + git_convention sections) leaves `git-strategy.yaml` byte-identical. Observed behavior: the sentinel is expanded into the entire compiled-default git_strategy tree (mode / provider / manual / personal / team / hooks / ...).

### A.2 Root cause (confirmed via git archaeology + live tree)

The failure is the interaction of three independently-correct pieces:

1. **Loader partial-override contract** (`internal/config/loader.go` `loadGitStrategySection`): the wrapper is seeded with compiled defaults, then the on-disk YAML is unmarshalled over it so user-omitted keys retain defaults. The sentinel key (`sentinel:`) is not in the struct, so it is silently dropped — `cfg.GitStrategy` ends up equal to the full compiled defaults.
2. **Unconditional Save() write set** (`internal/config/manager.go` `Save()`, the `saveSection("git-strategy.yaml", ...)` block at ~line 191, added by `33215af27`): `Save()` re-serializes ALL six owned sections (user / language / quality / git-convention / git-strategy / llm) on every call, regardless of which were modified.
3. **Web writer scope intent** (`internal/web/projectconfig.go` `writeProjectConfig`): intends to touch ONLY quality + git_convention via `SetSection` + `Save`, per REQ-WC3-007. It has no way to tell `Save()` to skip the sections it did not modify.

Combined: `writeProjectConfig` calls `Save()`, which writes `git-strategy.yaml` from `cfg.GitStrategy` (= full defaults), overwriting the sentinel.

### A.3 Why the other out-of-scope sections survive

`workflow.yaml` and `harness.yaml` survive because `Save()` does **not** own them at all (no `saveSection` call for them). `git_convention.yaml` survives a round-trip because its on-disk content matches its struct exactly, so re-serialization is content-stable. `git_strategy` is uniquely fragile because (a) it IS in the `Save()` write set and (b) its on-disk content (a sentinel) is not modeled by its struct, so re-serialization is lossy.

### A.4 Scope discipline (regression fix, not redesign)

This SPEC restores section isolation ONLY. It does NOT re-architect git_strategy Save, does NOT change the loader partial-override contract, and does NOT alter the git_strategy schema, validators, defaults, or templates. The SAVE-WIRING feature (git_strategy set→save→reload round-trip) MUST remain functional and its tests MUST stay green.

## §B. Requirements (GEARS)

### B.1 Primary isolation requirement

- **REQ-GSI-001** (Event-driven): **When** a caller persists configuration via `ConfigManager.Save()` after modifying only a subset of owned sections, the config manager **shall** leave every unmodified owned section's on-disk file content unchanged.

- **REQ-GSI-002** (Ubiquitous): The `ConfigManager` **shall not** expand the on-disk `git-strategy.yaml` content into the compiled-default git_strategy tree during a `Save()` that did not modify the git_strategy section.

### B.2 No-regression of the SAVE-WIRING feature

- **REQ-GSI-003** (Event-driven): **When** a caller invokes `SetSection("git_strategy", ...)` and then `Save()`, the config manager **shall** persist the supplied git_strategy values such that a fresh `Load()` recovers them (set→save→reload fidelity preserved).

- **REQ-GSI-004** (State-driven): **While** no `git-strategy.yaml` file exists on disk, **when** a caller invokes `Save()`, the config manager **shall** create `git-strategy.yaml` carrying the top-level `git_strategy:` key (the SAVE-WIRING file-creation contract is preserved for the greenfield case).

### B.3 Cross-caller safety

- **REQ-GSI-005** (Ubiquitous): The `ConfigManager` **shall** preserve the observable persistence behavior of every existing production caller of `Save()` (`internal/web` ×2, `internal/cli/profile_setup.go`, `internal/profile/sync.go`) — no caller signature change and no behavior change beyond the isolation restoration.

### B.4 Verification surface

- **REQ-GSI-006** (Event-detected): **When** the test suite runs, the failing guard test `TestWriteProjectConfigSectionIsolation` **shall** pass, AND the existing SAVE-WIRING tests (`TestConfigManagerSaveGitStrategyRoundTrip`, `TestConfigManagerSaveCreatesGitStrategyFile`) **shall** continue to pass without being weakened.

## §C. Success Criteria

1. `go test ./internal/web/ -run TestWriteProjectConfigSectionIsolation` is GREEN.
2. `go test ./internal/config/ -run 'TestConfigManagerSaveGitStrategyRoundTrip|TestConfigManagerSaveCreatesGitStrategyFile'` is GREEN.
3. `go test ./internal/config/... ./internal/web/... ./internal/cli/... ./internal/profile/...` is GREEN (no caller regression).
4. The guard test is not weakened (its sentinel-equality assertion is intact; no `t.Skip`, no assertion loosening).
5. No change to git_strategy schema (`types.go`), validators (`validation.go`), defaults factory, loader partial-override contract, or templates.

## §D. Exclusions (What NOT to Build)

### Out of Scope

The following are explicitly **OUT OF SCOPE** for this SPEC and are intentionally deferred or left untouched:

- **EX-1**: Do NOT re-architect the git_strategy Save path or change the `gitStrategyFileWrapper` type. Only restore section isolation.
- **EX-2**: Do NOT modify the loader partial-override contract (`loadGitStrategySection` seeding behavior). The sentinel being dropped on load is intentional loader behavior; the fix is on the write side.
- **EX-3**: Do NOT change the git_strategy schema, defaults factory, validators, or the shipped `git-strategy.yaml` template.
- **EX-4**: Do NOT touch any SPEC-SEC-HARDEN-001 file scope (security: permission / tmux / lsp / resilience). This SPEC's file scope is `internal/config` (git_strategy Save path) + the existing reproduction test.
- **EX-5**: Do NOT add a generic "write only dirty sections" feature beyond the minimum needed to fix the isolation defect. If a dirty-tracking mechanism is introduced, it is scoped to restoring the isolation invariant, not a broader Save() redesign.
- **EX-6**: Do NOT alter the four existing production callers of `Save()` (web ×2, profile_setup, profile/sync) — their call sites stay byte-unchanged.
- **EX-7**: Do NOT add a new SPEC ID with an alphabetic suffix or any frontmatter snake_case alias.

## §E. Affected Files (anticipated; plan.md finalizes)

| File | Change kind | Note |
|------|-------------|------|
| `internal/config/manager.go` | edit | the fix mechanism (Save isolation) — see plan.md §F |
| `internal/web/projectconfig_scope_test.go` | unchanged (reproduction) | the failing guard test; must go GREEN without edit |
| `internal/config/manager_save_git_strategy_test.go` | unchanged (no-regression) | SAVE-WIRING round-trip; must stay GREEN |

## §F. Cross-References

- Regression-introducing commit: `33215af27` (SPEC-PREPUSH-SAVE-WIRING-001 M1)
- Guard test origin: `e83864047` (SPEC-WEB-CONSOLE-003 M5)
- Predecessor SPEC (the feature this isolates): `.moai/specs/SPEC-PREPUSH-SAVE-WIRING-001/`
- GitHub Issue: #1064
- Module conventions: `internal/config/CLAUDE.md`
