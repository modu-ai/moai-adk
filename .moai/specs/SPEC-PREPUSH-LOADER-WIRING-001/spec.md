---
id: SPEC-PREPUSH-LOADER-WIRING-001
title: "Wire git_strategy config section into the loader chain (READ path)"
version: "0.2.0"
status: completed
created: 2026-06-08
updated: 2026-06-10
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/config"
lifecycle: spec-anchored
tags: "config, loader, git-strategy, dead-config, prepush"
era: V3R6
tier: S
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-08 | manager-spec | Initial draft — READ-only loader wiring for the `git_strategy` config section, closing the 3rd dead-config gap in the PREPUSH chain. |

---

## A. Context & Problem Statement

`internal/config/loader.go` `Load()` (`loader.go:31-92`) invokes 14 per-section loaders (15 after this SPEC adds `loadGitStrategySection`)
(`user`, `language`, `quality`, `git_convention`, `llm`, `ralph`, `state`, `workflow`,
`statusline`, `research`, `constitution`, `context`, `interview`, `design`) — but there is
**no `loadGitStrategySection`**. As a direct consequence, `cfg.GitStrategy` is never
populated from `git-strategy.yaml`; after `Load()` it retains only the compiled defaults
produced by `NewDefaultConfig()` / `NewDefaultGitStrategyConfig()` (`defaults.go:232`).

The user's `.moai/config/sections/git-strategy.yaml` (top-level key `git_strategy:`, with
e.g. `mode: team` + `team.hooks.pre_push: enforce`) is therefore **silently ignored at
runtime**. This is the root cause that left SPEC-PREPUSH-MODE-WIRING-001's
`resolvePrePushAction()` reading `ActiveModeProfile().Hooks.PrePush` and getting the
compiled default `warn` regardless of what the YAML file declares — the consumer is correct
per its own AC contract, but the chain is not end-to-end functional because the READ half
of the config pipeline is missing.

### Dead-config chain (this is the 3rd follow-up)

1. **SPEC-PREPUSH-WIRING-001** — wired `enforce_on_push` into the shell pre-push hook.
2. **SPEC-PREPUSH-MODE-WIRING-001** — wired `git_strategy.<mode>.hooks.pre_push` into
   `runPrePush` via `resolvePrePushAction()` (the CONSUMER).
3. **SPEC-PREPUSH-LOADER-WIRING-001 (this SPEC)** — wires the `git_strategy` SECTION into
   the config loader so the prior SPEC's reader sees the user's actual `git-strategy.yaml`
   values, not the compiled default (the PRODUCER / READ path).

### Ground-truth source facts (verified against live source this plan-phase)

- `cfg.GitStrategy` is typed `GitStrategyConfig` (local `config` package, NOT `models.`),
  declared at `types.go:20` (`GitStrategy GitStrategyConfig \`yaml:"git_strategy"\``).
- `loadGitConventionSection` (`loader.go:149-161`) is the exact sibling pattern to mirror:
  seed wrapper with `cfg.GitConvention` → `loadYAMLFile(dir, "git-convention.yaml", wrapper)`
  → on `loaded` set `cfg.GitConvention = wrapper.GitConvention` + `l.loadedSections["git_convention"] = true`;
  on error `slog.Warn` + keep defaults.
- `gitConventionFileWrapper` exists at `types.go:1071-1073`; **`gitStrategyFileWrapper` does
  NOT exist** and must be created (mirroring it, but wrapping the local-package
  `GitStrategyConfig` rather than `models.GitConventionConfig`).
- `loadYAMLFile` (`loader.go:398-413`) uses plain `yaml.Unmarshal` — **NOT** strict / no
  `KnownFields(true)` — so unknown keys in `git-strategy.yaml` are silently ignored; loading
  cannot fail on extra keys (no strict-mode regression risk).
- The `loadXxxSection` contract is partial-override: the wrapper is seeded with current
  defaults, then `yaml.Unmarshal` overlays file values, so keys absent from the file keep
  their default.
- `loadedSections` map (`loader.go:20,35`) tracks which sections loaded; `"git_strategy"`
  must be added on success (mirror siblings).
- The `git-strategy.yaml` top-level key is `git_strategy:` (verified in both the template
  `git-strategy.yaml.tmpl:4` and local `.moai/config/sections/git-strategy.yaml`).
- The in-memory `SetSection`/`GetSection` `case "git_strategy"` already exists
  (`manager.go:261-262` get, `:307-312` set) and is correct — **not to be touched**.
- `Save()` (`manager.go:155-196`) persists only `user`/`language`/`quality`/`git-convention`/`llm`
  — it does **NOT** persist `git-strategy.yaml`. The WRITE half is also dead, but it is
  **OUT OF SCOPE** here (see §C Exclusions).
- The compiled default `PrePush` is `"warn"` for all three mode profiles (manual / personal /
  team — `defaults.go:249,261,276`). The live `git-strategy.yaml` declares `mode: team` +
  `team.hooks.pre_push: enforce`. This is precisely the divergence the dead-config masks.

---

## B. Requirements (GEARS)

### Functional Requirements

**REQ-PLW-001 (Event-driven — file present).**
**When** `Load()` is invoked against a `.moai/config/sections/` directory that contains a
`git-strategy.yaml` file, the config loader **shall** populate `cfg.GitStrategy` from the
file contents (overlaying file values onto the seeded compiled defaults).

**REQ-PLW-002 (Event-driven — loaded-section bookkeeping).**
**When** `git-strategy.yaml` is present and parses successfully, the config loader
**shall** set `loadedSections["git_strategy"] = true` such that
`LoadedSections()["git_strategy"]` returns `true`.

**REQ-PLW-003 (Event-driven — file absent).**
**When** no `git-strategy.yaml` file exists in the sections directory, the config loader
**shall** leave `cfg.GitStrategy` at its compiled defaults and **shall not** set
`loadedSections["git_strategy"]` (mirroring the absent-file behavior of every sibling
loader).

**REQ-PLW-004 (State-driven — partial override).**
**While** a `git-strategy.yaml` file specifies only a subset of keys, the config loader
**shall** override exactly those specified keys and **shall** preserve the compiled
defaults for every unspecified key.

**REQ-PLW-005 (Event-driven — malformed / unknown keys).**
**When** `git-strategy.yaml` contains unknown or extraneous keys, the config loader
**shall not** fail `Load()` (the loader uses non-strict `yaml.Unmarshal`); it **shall**
load the keys it recognizes and ignore the rest.

**REQ-PLW-006 (Ubiquitous — wrapper struct existence).**
The `internal/config` package **shall** provide a `gitStrategyFileWrapper` struct whose
top-level YAML key is `git_strategy:` and whose single field wraps the local-package
`GitStrategyConfig` (mirroring `gitConventionFileWrapper`).

**REQ-PLW-007 (Ubiquitous — loader function + wiring).**
The `internal/config` package **shall** provide a `loadGitStrategySection(dir string, cfg *Config)`
method that is an exact structural mirror of `loadGitConventionSection`, AND `Load()`
**shall** invoke it (the wired call placed adjacent to `loadGitConventionSection` for
natural git-section grouping).

**REQ-PLW-008 (Ubiquitous — end-to-end chain completion).**
The config loader **shall** make the user's `git-strategy.yaml` mode-profile hook values
observable to downstream readers such that, after `Load()`, when the file declares
`mode: team` + `team.hooks.pre_push: enforce`, `cfg.GitStrategy.ActiveModeProfile()`
returns a profile whose `Hooks.PrePush == "enforce"` (proving SPEC-PREPUSH-MODE-WIRING-001's
`resolvePrePushAction()` now reads the YAML value end-to-end, instead of the compiled
default `warn`).

### Non-Functional Requirements

**REQ-PLW-009 (Ubiquitous — minimal Tier S footprint).**
The change **shall** be confined to one new wrapper struct (`types.go`), one new loader
method + one wired call (`loader.go`), and accompanying unit tests; it **shall not**
refactor the loader, introduce a generic section-registry, or alter unrelated loaders.

**REQ-PLW-010 (Ubiquitous — pattern parity).**
The new loader method **shall** follow the exact error/bookkeeping semantics of the sibling
loaders (`slog.Warn` + keep defaults on error; set `loadedSections` flag on success only),
so that no new behavioral idiom is introduced into the loader.

---

## C. Exclusions (What NOT to Build)

### Out of Scope

The following are explicitly **OUT OF SCOPE** for this SPEC and are intentionally deferred or
left untouched:

- **WRITE / Save path** — adding `git-strategy.yaml` persistence to `Save()`
  (`manager.go:155-196`) and a corresponding `saveSection("git-strategy.yaml", ...)` call.
  The WRITE half is also a dead-config gap, but there is no production caller that modifies
  `cfg.GitStrategy` and then saves it, so READ-only is the user-approved scope this session.
  `Save()` MUST remain byte-unchanged with respect to git-strategy.
- **`SetSection` / `GetSection` changes** — the in-memory `case "git_strategy"` accessors
  (`manager.go:261-262`, `:307-312`) already work and MUST NOT be modified.
- **`hook_pre_push.go` / `resolvePrePushAction()`** — the consumer from
  SPEC-PREPUSH-MODE-WIRING-001 is correct and is not touched; this SPEC only makes its input
  reflect real YAML.
- **`validation.go` enum gate** — `validation.go` already validates `git_strategy.*`; once the
  section loads, validation runs on the real values. No new enum gate is added, consistent
  with SPEC-PREPUSH-MODE-WIRING-001's resolver-side normalization decision.
- **`defaults.go` / templates** — `NewDefaultGitStrategyConfig()` and the shipped
  `git-strategy.yaml.tmpl` are correct and unchanged.
- **Environment-variable overrides for `git_strategy`** — no `MOAI_GIT_STRATEGY_*` env
  override layer is added; that is a separate concern beyond this loader-wiring scope.
- **Generic section-registry refactor** — the loader's per-section method pattern is
  preserved as-is; no abstraction is introduced.

---

## D. Acceptance Criteria Summary

Full Given-When-Then scenarios live in `acceptance.md`. The AC families are:

- AC-PLW-001..004 — file present / absent / partial / malformed behavior (REQ-PLW-001..005)
- AC-PLW-005..006 — dead-config elimination (wrapper + loader symbol + wired call grep)
- AC-PLW-007 — end-to-end chain completion via `ActiveModeProfile().Hooks.PrePush == "enforce"`
- AC-PLW-008 — scope boundary: `Save()` unchanged (`grep -c 'git-strategy.yaml' manager.go` stays 0)

---

## E. Traceability

| REQ | AC | Module symbol |
|-----|-----|---------------|
| REQ-PLW-001 | AC-PLW-001 | `loadGitStrategySection` (overlay) |
| REQ-PLW-002 | AC-PLW-001 | `loadedSections["git_strategy"]` |
| REQ-PLW-003 | AC-PLW-002 | absent-file path |
| REQ-PLW-004 | AC-PLW-003 | partial overlay |
| REQ-PLW-005 | AC-PLW-004 | non-strict `yaml.Unmarshal` |
| REQ-PLW-006 | AC-PLW-005 | `gitStrategyFileWrapper` (types.go) |
| REQ-PLW-007 | AC-PLW-005, AC-PLW-006 | `loadGitStrategySection` + `Load()` wired call |
| REQ-PLW-008 | AC-PLW-007 | `ActiveModeProfile().Hooks.PrePush` |
| REQ-PLW-009 | AC-PLW-008 | footprint bound |
| REQ-PLW-010 | AC-PLW-001..004 | pattern parity |
