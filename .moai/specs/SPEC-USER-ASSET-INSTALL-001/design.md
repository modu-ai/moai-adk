---
id: SPEC-USER-ASSET-INSTALL-001
title: "design.md — user-folder asset install architecture"
version: "0.1.0"
created: 2026-10-05
updated: 2026-10-05
author: manager-spec
---

# Design — User-Folder Common-Asset Install

## 1. Target Model

One install axis changes: WHERE common skills and agents live.

```
BEFORE (today, two modes)                    AFTER (this SPEC)
├─ project .claude/skills + commands         ├─ USER ~/.claude/skills, ~/.claude/agents   (Claude)
│   (local mode) OR plugin carrier           ├─ USER $HOME/.agents/skills                 (Codex skills)
│   (plugin mode)                            ├─ USER ~/.codex/agents                      (Codex agents)
├─ project .claude/agents, .codex/agents     ├─ project: settings, AGENTS.md, lock file,
├─ project .agents/skills mirror                 project-only harness (hooks, .mcp.json,
└─ plugin marketplace + install step             output-styles, rules, commands wrappers)
                                             └─ NO plugin anything
```

Per-profile settings folders (`~/.moai/claude-profiles/<name>`, the
`CLAUDE_CONFIG_DIR` roots) remain per-profile for SETTINGS and are never an
install target for the shared asset tree (decision-index D-Q3 governs how a
profile session sees the user-level assets).

## 2. Components

### 2.1 User-asset installer (new, `internal/template` sibling or subpackage)

- Reads the same embedded tree + catalog (`//go:embed all:templates`,
  `catalog.yaml`) the project deployer reads — zero new asset sources.
- Writes to four user roots, resolved at run time:
  - Claude: `$HOME/.claude/skills/`, `$HOME/.claude/agents/`
  - Codex: `$HOME/.agents/skills/`, `$HOME/.codex/agents/`
- Per-file rule set (inherited semantics from V17 evidence):
  - target exists AND per-user manifest tracks it → refreshable (hash compare)
  - target exists AND manifest does NOT track it → SKIP + report (collision;
    the file may be the user's — mirrors `rehomeOneSkill`'s skip-and-report and
    `UserCreated` provenance semantics)
  - target absent → install
- Per-file failure (permissions, EISDIR, …) → continue + surface in summary
  (fail-open per file, loud at the end; never a silent partial install).
- Project-root confinement (`validateDeployPath`) gets a user-scope counterpart:
  every destination MUST be under one of the four declared roots — the user
  installer refuses any path outside them (path-trust boundary).

### 2.2 Per-user manifest (new)

- One JSON document, location per decision gate D-Q2 (candidates:
  `~/.moai/user-assets.json` — the leading candidate because `~/.moai/` is
  already moai's user-level state home per `internal/paths/paths.go`;
  `~/.claude/moai-manifest.json`; per-root split files).
- Schema: `schema_version`, `installed_at`, `moai_version`, `files: {path →
  {sha256, bundle, installed_at}}`, `collisions: [{path, first_seen_at}]`.
- Reuses the sha256 hex convention of `catalog.yaml` entries and the triple-hash
  spirit of the project manifest (`internal/manifest/types.go`) minus the parts
  the user scope does not need (no 3-way merge at user scope: collision = skip).
- Schema-version gate: unknown `schema_version` → refuse manifest-driven
  removal (REQ-021); install/refresh may still proceed in append-only fashion.

### 2.3 Bundle taxonomy (catalog extension)

- `catalog.yaml` gains a user-install view: L0 core (REQ-003's content, exact
  agent list per D-Q1) + opt-in bundles (REQ-004). The existing
  `optional_packs` structure is the natural carrier for bundles; the remaining
  current-`core` entries that are NOT L0 reclassify into bundles. The catalog
  is the single membership SSOT (no second name list in code — the
  `publishedSkillNames` hardening pattern shows why drift guards exist).
- Factory (multi-lane operation) is L0 per the operator decision D3:
  `moai-factory-foreman` + `moai-lane-watchdog` ride the core bundle; the
  factory doctor check (`checkFactoryRun`) keeps reading project state as
  today.

### 2.4 `moai update` integration

- New user-asset phase in the update flow: refresh (hash-diff) → removal
  (manifest-tracked files no longer in ANY shipped bundle) → summary counts
  (installed/refreshed/removed/collision-skipped — REQ-011).
- Removal is manifest-driven ONLY: a file in a user folder that the manifest
  does not track is never a removal candidate (REQ-010 collision rule covers
  it; deletion of untracked files is out of scope).
- Ordering: user-asset phase runs BEFORE the project phase so a mid-update
  failure leaves the project phase untouched (project behavior unchanged by a
  user-side failure).

### 2.5 Project slimming

- Project deploy stops emitting common skills/agents entirely (both former
  modes): the project keeps settings, AGENTS.md/CLAUDE.md, the lock file
  (`.moai/manifest.json` unchanged in role), hooks, `.mcp.json`, output-styles,
  rules, command wrappers.
- Migration of EXISTING projects (REQ-020): update classifies current
  project-local `moai-*` skills/agents via the provenance classes —
  `template_managed` → removable (offer + apply), `user_modified` → preserved
  with a report (3-way decision is NOT silent), `user_created` → untouched.
  The existing namespace contract (V12) is the classifier.

### 2.6 `moai doctor` integration

- New checks (workspace group): (1) user-install integrity — every
  manifest-tracked path exists with matching hash; untracked files in the four
  roots listed as informational; (2) project-vs-lock comparison — drift in
  both directions (project file absent from lock; lock entry absent from
  project). Both read-only, report-only (matching the doctor house style).
- Retired-carrier checks (`checkPluginDeployment`, `checkPluginVersion`) are
  repointed to the user-manifest comparison or removed with their SPEC's REQs
  cited in the commit (REQ-019).

### 2.7 Plugin carrier disposition

- `internal/template/pluginemit/` (generator + 13 test files), the committed
  `plugins/moai/` tree, `.claude-plugin/marketplace.json`,
  `.agents/plugins/marketplace.json`, Makefile `plugin-emit`/`plugin-emit-check`
  targets (and their `build:` prerequisite), the roster-guard sweep-skip note
  (`internal/harness/rosterguard/check.go:338-341`), and the plugin install
  step (`internal/cli/plugin_install.go` + init/update call sites) go away.
- `deployer_mode.go`'s split (DeployModePlugin, PluginMirrorPolicy, the
  exclusion walk, the MCP strip, the plugin re-home) is retired; the deployer
  returns to a single project payload shape (the slim one).
- `.mcp.json` rendering returns to always carrying the `moai` entry (the
  plugin was its alternative carrier).

## 3. Sequencing Logic (why the milestones order this way)

Decision-reversibility: M0/M1 pin the data model (bundle taxonomy + manifest
schema + the two decision gates D-Q1/D-Q2) — the highest-change-likelihood
decisions, reviewed first. M2-M5 build the new behavior on the frozen schema.
M6/M7 are the demolitions, last, because they are the milestones the operator
would drop if the direction reverses, and because deleting the carrier before
the replacement installs would strand users between milestones. M8 closes with
cross-harness evidence.

## 4. Risks and Mitigations

| Risk | Mitigation |
|---|---|
| Wide retire blast radius (13+ test files, build chain, roster guard) | M6/M7 isolated; each deletion cites the owning REQ; golden tests deleted WITH the generator in the same commit |
| Users left with plugin installs from the old model | Migration report names the manual `claude plugin uninstall` step (doctor informational row); no silent divergence |
| Profile sessions not seeing user assets (D-Q3 unresolved) | Gate blocks M2 completion on the operator answer; M8 measures the chosen policy |
| Collision rule too strict (user never gets updates after a manual edit) | Summary distinguishes tracked-but-modified (refreshable) from untracked (collision); doctor lists collisions persistently in the manifest |
| Manifest corruption / partial write | Atomic write (same pattern as `atomicWriteFile`); schema-version refusal for removal |
