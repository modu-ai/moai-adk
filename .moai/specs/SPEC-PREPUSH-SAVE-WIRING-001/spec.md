---
id: SPEC-PREPUSH-SAVE-WIRING-001
title: "Wire git_strategy config section into the Save() WRITE path (READ/WRITE symmetry)"
version: "0.1.0"
status: draft
created: 2026-06-11
updated: 2026-06-11
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/config"
lifecycle: spec-anchored
tags: "config, save, git-strategy, dead-config, prepush, write-path"
era: V3R6
tier: S
---

# SPEC-PREPUSH-SAVE-WIRING-001 — Wire git_strategy into the Save() WRITE path

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-11 | manager-spec | Initial draft — 4th and final SPEC in the PREPUSH dead-config chain; completes the READ/WRITE symmetry left open by SPEC-PREPUSH-LOADER-WIRING-001 (AC-PLW-008 explicitly deferred the WRITE leg). |

---

## A. Context (Problem & Ground Truth)

### The READ/WRITE asymmetry this SPEC closes

The PREPUSH dead-config chain wired the pre-push hook engine end-to-end on the **READ** side.
The final structural gap is on the **WRITE** side: `ConfigManager.Save()` persists exactly 5
sections (`user` / `language` / `quality` / `git-convention` / `llm`), but `git_strategy` is the
one section that is **populated, loaded, in-memory-mutable (via `SetSection`), and validated — yet
never written back to disk**. This SPEC adds the `git-strategy.yaml` WRITE leg as a 1:1 mirror of
the existing `git-convention.yaml` WRITE leg, completing the symmetry.

This is **latent infrastructure**, not live dead-config. There is currently no production caller
that mutates `cfg.GitStrategy` and then calls `Save()`. The honest framing: this SPEC (1) closes a
documented READ/WRITE asymmetry — the loader reads `git_strategy` but `Save()` cannot write it
back, so a config round-trip silently drops user edits to that section — and (2) pre-stages the
**web-console git_strategy editor export seam** (a separate future web-console-cohort SPEC) that
WILL call `SetSection("git_strategy", ...)` → `Save()`. Without this WRITE leg that future caller
would persist nothing.

### The PREPUSH dead-config chain (do not re-derive)

1. **SPEC-PREPUSH-WIRING-001** (completed) — wired the dormant pre-push hook engine.
2. **SPEC-PREPUSH-MODE-WIRING-001** (completed) — wired `git_strategy.<mode>.hooks.pre_push`
   into the runtime READER (`resolvePrePushAction`). Discovered the 3rd dead-config: the whole
   `git_strategy` section was unwired in the loader.
3. **SPEC-PREPUSH-LOADER-WIRING-001** (completed) — wired `git_strategy` into the config loader
   READ path (`loadGitStrategySection` + `gitStrategyFileWrapper`). Its AC-PLW-008 explicitly
   DEFERRED the WRITE/Save leg ("Save() unchanged — WRITE path not added").
4. **SPEC-PREPUSH-SAVE-WIRING-001 (this SPEC)** — adds the `git-strategy.yaml` WRITE leg to
   `Save()`, the inverse of #3, completing the READ/WRITE round-trip.

### Ground-truth source facts (verified against live source this plan-phase)

- `ConfigManager.Save()` is at `internal/config/manager.go:155` (`func (m *ConfigManager) Save() error`),
  carrying `@MX:ANCHOR fan_in=12 across 4 files`. It writes each section atomically via
  `saveSection(sectionsDir, "<file>.yaml", <wrapper>)` (temp-file + `os.Rename`).
- `Save()` currently calls `saveSection` for **exactly 5 sections** (`manager.go:170-193`):
  `user.yaml`, `language.yaml`, `quality.yaml`, `git-convention.yaml`
  (line 185-188: `gitConventionFileWrapper{GitConvention: m.config.GitConvention}`), `llm.yaml`.
  **`git-strategy.yaml` is absent** — this is the dead-config to fix.
- The exact sibling to mirror is the git-convention WRITE leg:
  `saveSection(sectionsDir, "git-convention.yaml", gitConventionFileWrapper{GitConvention: m.config.GitConvention})`.
  The git-strategy WRITE leg is:
  `saveSection(sectionsDir, "git-strategy.yaml", gitStrategyFileWrapper{GitStrategy: m.config.GitStrategy})`.
- `gitStrategyFileWrapper` **already exists** at `types.go:1076-1081` (added by LOADER-WIRING-001
  for the READ path). It wraps the local-package `GitStrategyConfig` under YAML key `git_strategy:`.
  **REUSE it; do NOT create a new wrapper.** It is the symmetric mirror of `gitConventionFileWrapper`
  (`types.go:1071-1073`).
- The READ-path filename / key contract from LOADER-WIRING-001 (`loader.go:166-180`) is:
  file `git-strategy.yaml`, top-level key `git_strategy`, wrapper field `GitStrategy`. The WRITE
  leg MUST use the **same filename + same wrapper** so the round-trip is byte-consistent.
- `SetSection("git_strategy", value)` / `GetSection("git_strategy")` accessors **already exist and
  are correct** (`manager.go:261-262` get-case, `:307-312` set-case, expecting `GitStrategyConfig`).
  The export seam is therefore **already complete** — this SPEC does NOT need to add or touch any
  get/set case. (Correction to an earlier assumption: the in-memory section accessors are already
  wired; only the Save() leg is missing.)
- No `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl`-to-local seeding is
  needed for Save(): `Save()` CREATEs `git-strategy.yaml` from the in-memory `cfg.GitStrategy`
  (which carries `NewDefaultGitStrategyConfig()` defaults when no file was loaded) on first persist.
  The sections directory is `MkdirAll`-ensured at `manager.go:166` before any `saveSection`.
- `saveSection` (`manager.go:373`) marshals via `yaml.Marshal` and `atomicWrite`s — no per-section
  special-casing; adding one more `saveSection` call is the entire change surface.

---

## B. Requirements (GEARS)

### Functional Requirements

**REQ-PSW-001 (Event-driven — Save persists git-strategy).**
**When** `Save()` is invoked on an initialized `ConfigManager`, the config manager **shall**
persist the in-memory `cfg.GitStrategy` to `<sectionsDir>/git-strategy.yaml` via
`saveSection(sectionsDir, "git-strategy.yaml", gitStrategyFileWrapper{GitStrategy: m.config.GitStrategy})`.

**REQ-PSW-002 (Event-driven — file created on first persist).**
**When** `Save()` runs against a sections directory that does not yet contain `git-strategy.yaml`,
the config manager **shall** create the file from the in-memory `cfg.GitStrategy` (carrying the
compiled defaults when no prior file was loaded), consistent with how every sibling `saveSection`
call creates its file.

**REQ-PSW-003 (Ubiquitous — wrapper reuse).**
The git-strategy WRITE leg **shall** reuse the existing `gitStrategyFileWrapper` struct
(`types.go:1076-1081`, top-level key `git_strategy:`); it **shall not** introduce a new wrapper
type. The same wrapper is used by both the READ leg (`loadGitStrategySection`) and this WRITE leg,
guaranteeing a symmetric round-trip.

**REQ-PSW-004 (Ubiquitous — round-trip fidelity).**
The config manager **shall** persist `git_strategy` such that a non-default value set in memory via
`SetSection("git_strategy", ...)` then `Save()`d is recoverable by a fresh `Load()` /
`Reload()` — i.e. `set → save → reload` preserves the value (e.g. `GitStrategy.Mode` or
`GitStrategy.Team.Hooks.PrePush`).

**REQ-PSW-005 (State-driven — existing sections unaffected).**
**While** the git-strategy WRITE leg is added, the config manager **shall** continue to persist the
5 pre-existing sections (`user` / `language` / `quality` / `git-convention` / `llm`) exactly as
before, with no change to their filenames, wrappers, ordering semantics, or atomic-write behavior.

### Non-Functional Requirements

**REQ-PSW-006 (Ubiquitous — minimal Tier S footprint, no scope creep).**
The change **shall** be confined to adding exactly one `saveSection(... "git-strategy.yaml" ...)`
call inside `Save()` (`manager.go`) plus accompanying unit tests; it **shall not** add any new
validator, alter `defaults.go`, add a template file, refactor `Save()` into a loop / registry, or
touch `SetSection` / `GetSection` / `loader.go` / `validation.go`.

**REQ-PSW-007 (Ubiquitous — pattern parity).**
The new WRITE leg **shall** be a structural mirror of the git-convention WRITE leg
(`manager.go:185-188`): same `saveSection` signature shape, same `fmt.Errorf("save ...: %w", err)`
error-wrap idiom, placed adjacent to the git-convention leg for natural git-section grouping. No
new behavioral idiom is introduced into `Save()`.

---

## C. Exclusions (What NOT to Build)

### Out of Scope

The following are explicitly **OUT OF SCOPE** for this SPEC and are intentionally deferred or left
untouched:

- **The web-console git_strategy editor (the future caller)** — adding a UI surface or HTTP handler
  that calls `SetSection("git_strategy", ...)` → `Save()` is a SEPARATE web-console-cohort SPEC.
  This SPEC builds only the WRITE infrastructure that future caller depends on. Without a caller,
  the WRITE leg is latent (mirrors the WEB-CONSOLE-007/008/009 export-seam pattern: thin config
  export seam, zero new validators).
- **New validators** — `validation.go` already validates `git_strategy.*` on Load. No new
  validation hook, enum gate, or `Validate()` call is added on the WRITE side. Validation continues
  to run at Load time, consistent with SPEC-PREPUSH-MODE-WIRING-001's resolver-side decision.
- **`SetSection` / `GetSection` changes** — the in-memory `case "git_strategy"` accessors
  (`manager.go:261-262`, `:307-312`) already exist and are correct; they MUST NOT be modified.
- **READ-path changes** — `loadGitStrategySection` / `gitStrategyFileWrapper` / `loader.go` from
  SPEC-PREPUSH-LOADER-WIRING-001 are correct and unchanged; this SPEC adds only the inverse WRITE
  leg.
- **`defaults.go` changes** — `NewDefaultGitStrategyConfig()` is correct and unchanged. Save()
  persists whatever is in memory; defaults are not modified to satisfy symmetry.
- **Template addition** — no new `git-strategy.yaml.tmpl` is required; `Save()` CREATEs the file
  from the in-memory config when absent. Adding a template is unnecessary for the WRITE leg and is
  out of scope.
- **`Save()` refactor into a loop / section-registry** — the 5-section explicit-call structure is
  preserved; this SPEC adds one 6th explicit call, not a generic registry.
- **Environment-variable overrides for `git_strategy`** — no `MOAI_GIT_STRATEGY_*` env override is
  added; `applyEnvOverrides` is unchanged.

---

## D. Acceptance Criteria (inline — Tier S)

Full Given-When-Then enumeration lives in `plan.md` § Acceptance. The mandatory minimum set:

| AC | REQ | Assertion |
|----|-----|-----------|
| AC-PSW-001 | REQ-PSW-001, REQ-PSW-002 | After `Save()`, `<sectionsDir>/git-strategy.yaml` exists on disk (grep count of `"git-strategy.yaml"` in `manager.go` Save() is ≥ 1 — the inverse of LOADER-WIRING-001 AC-PLW-008). |
| AC-PSW-002 | REQ-PSW-004 | Round-trip: `SetSection("git_strategy", <non-default>)` → `Save()` → fresh `Load()` recovers the non-default value (e.g. `Mode == "personal"` when default is `"team"`, or `Team.Hooks.PrePush == "enforce"` when default is `"warn"`). |
| AC-PSW-003 | REQ-PSW-003 | The persisted file round-trips through the SAME `gitStrategyFileWrapper` used by the READ leg — top-level key `git_strategy:` present in the written YAML; no new wrapper type added (`grep -c 'type gitStrategyFileWrapper' types.go == 1`). |
| AC-PSW-004 | REQ-PSW-005 | No regression to the 5 existing saved sections: `TestConfigManagerSaveAndReloadRoundTrip` and `TestConfigManagerSaveCreatesDirectory` still pass; user/language/quality/git-convention/llm round-trips unchanged. |
| AC-PSW-005 | REQ-PSW-006 (MUST — no scope creep) | No new validator, no `defaults.go` change, no template added, no `SetSection`/`GetSection`/`loader.go`/`validation.go` edit. Verified by: `git diff --name-only` touches only `internal/config/manager.go` + a `_test.go`; `grep -c 'func validate' internal/config/validation.go` unchanged. |

---

## E. References

- `internal/config/manager.go:155` — `Save()` (the change site)
- `internal/config/manager.go:185-188` — git-convention WRITE leg (mirror reference)
- `internal/config/types.go:1076-1081` — `gitStrategyFileWrapper` (reused, NOT recreated)
- `internal/config/loader.go:166-180` — `loadGitStrategySection` (READ-path counterpart)
- `.moai/specs/SPEC-PREPUSH-LOADER-WIRING-001/` — predecessor (READ leg); AC-PLW-008 deferred this WRITE leg
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — frontmatter SSOT
