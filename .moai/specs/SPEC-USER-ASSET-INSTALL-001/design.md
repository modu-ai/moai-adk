---
id: SPEC-USER-ASSET-INSTALL-001
title: "design.md — user-folder asset install architecture"
version: "0.5.0"
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
install target for the shared asset tree (decision-index D-Q3: closed at plan
phase — profile sessions do not see the shared user assets in v1; premise P6).

## 2. Components

### 2.1 User-asset installer (new, `internal/template` sibling or subpackage)

- Reads the same embedded tree + catalog (`//go:embed all:templates`,
  `catalog.yaml`) the project deployer reads — zero new asset sources.
- Writes to four user roots, resolved at run time:
  - Claude: `$HOME/.claude/skills/`, `$HOME/.claude/agents/`
  - Codex: `$HOME/.agents/skills/`, `$HOME/.codex/agents/`
- Per-file rule set — the four-state hash truth table (iter2 D16; state ×
  action for a manifest-tracked file):
  - **manifest-match** (current == manifest hash, ≠ shipped): refresh →
    rewrite to shipped bytes + re-record hash/version (REQ-008); removal →
    remove (REQ-009).
  - **up-to-date** (current == manifest hash == shipped): refresh → no-op,
    not counted; removal → remove.
  - **manifest-stale** (current ≠ manifest hash, == shipped bytes — e.g. a
    prior refresh crashed before the manifest write): refresh → repair the
    manifest entry to the shipped state WITHOUT a file rewrite, counted under
    "refreshed" (REQ-011); removal → remove — the bytes are moai's own, which
    is exactly REQ-009's shipped-bytes alternative (iter4 D25: the one
    removal rule is REQ-009 as extended — current == manifest hash OR current
    == shipped bytes where a shipped source exists; this table arm and the
    REQ now state the same rule).
  - **divergence** (current equals NEITHER manifest hash NOR shipped bytes):
    preserve (REQ-023) — at refresh, back up the shipped replacement to the
    backup home `~/.moai/backups/<root-slug>/<relpath>` (iter4 D33: the
    layout is ROOT-SLUG-PREFIXED — each root gets a fixed slug, e.g.
    `claude-skills`, `claude-agents`, `agents-skills`, `codex-agents`, so
    `~/.claude/skills/x` and `$HOME/.agents/skills/x` cannot collide on
    `skills/x`; the `<relpath>` is Cleaned and `..`-rejected before joining,
    and the backup home is judged on RESOLVED paths exactly like the four
    roots — resolved once per run, destination resolved immediately before
    the write, an escape from the resolved backup root refused; C2's sole
    out-of-root write carve-out for user-folder asset writes — NOT
    "alongside the user folder", which would sit outside the four roots or
    pollute them as an untracked file); at removal of a file dropped from
    every bundle, NO shipped-bytes backup exists, so the file is preserved
    in place + reported; the tracked path is unmodified by refresh AND by
    removal; always reported.
  - **missing** (tracked in manifest, absent on disk): refresh → reinstall +
    record; removal → drop the manifest entry, count as removed.
  - target exists AND manifest does NOT track it → SKIP + report (collision;
    the file may be the user's — mirrors `rehomeOneSkill`'s skip-and-report and
    `UserCreated` provenance semantics)
  - target absent AND untracked → install
- Init trigger semantics (round-5 F3): init's install is PER-ASSET-STATE
  based, not manifest-absence based — for each L0 and opted-in-bundle asset,
  an absent target, or one whose bytes differ from its manifest record, is
  (re)installed; a present, manifest-matching target is a no-op. A manifest
  left by a PARTIAL install therefore does not suppress the run: init
  completes the shortfall idempotently, sharing the shortfall-append
  semantics with update's upgrade branch (§2.4 state (a)). The former
  "no per-user install" wording (v0.4.0 and earlier) left the
  partial-install retry skipping the missing assets with no project
  fallback after M4.
- Per-file failure (permissions, EISDIR, …) → continue + surface in summary
  (fail-open per file, loud at the end; never a silent partial install).
- Four-root confinement is judged on RESOLVED paths, covering all three
  second-order edges (iter2 D17):
  1. **Symlinked root** — each root is resolved once at install start
     (`filepath.EvalSymlinks`); a root that is itself a symlink (dotfile-manager
     `~/.claude` → elsewhere) is a legal boundary at its RESOLVED location —
     containment is judged resolved-root vs resolved-destination, so such a
     setup neither collapses into wholesale refusal nor escapes the boundary.
  2. **Leaf symlink inside a root** — a managed leaf that is itself a symlink
     is resolved; if it resolves outside its own root's resolved tree it is
     NEVER written through: refused at install, classified as divergence
     (REQ-023 preserve + report) at refresh/removal.
  3. **TOCTOU (check-then-open)** — the resolve-then-write window is NARROWED
     by posture, not closed by it (iter4 D26 downgrading the earlier
     "closed" claim, which was false as stated): the payload is written to a
     temp file created inside the validated RESOLVED directory and moved
     onto the final path with an atomic rename — `rename(2)` does not follow
     a symlink on its destination's final component, so a leaf swapped in
     after validation is replaced, not followed — with the resolved parent
     re-validated immediately before the rename (the
     O_NOFOLLOW-equivalent semantics for the update path). DECLARED
     LIMITATION: `rename(2)` still traverses INTERMEDIATE directory
     components, so a parent directory swapped to an outside-pointing
     symlink after that re-validation but before the rename can redirect the
     rename outside the root (codex reproduced the sequence on a /tmp model,
     iter3 D26). The remaining window is the instant between the
     re-validation and the rename; full closure needs descriptor-relative
     operations (held dirfd + the openat/renameat family, or
     O_NOFOLLOW|O_DIRECTORY on the parent), which are POSIX-only and are
     declared out of scope for v1 (the SPEC pins no OS-specific
     system-call dependency — D8; C1's cross-platform determinism). The
     limitation is
     doc-visible (acceptance §D.7) and AC-025 asserts the posture itself.
     Manifest writes keep the existing `atomicWriteFile` pattern.

### 2.2 Per-user manifest (new)

- One JSON document at `~/.moai/user-assets.json` (D-Q2, adjudicated
  2026-10-05 — leader default, operator-contestable; `~/.moai/` is already
  moai's user-level state home per `internal/paths/paths.go`). The manifest
  is the SPEC's own state file — a separately named out-of-root write
  destination alongside the REQ-023 backup home, distinct from the C2
  user-folder asset-write surface (iter4 D34).
- Schema: `schema_version`, `installed_at`, `bundles: [opted-in bundle
  names]`, `files: {path → {sha256, bundle,
  installed_at, moai_version}}`, `collisions: [{path, first_seen_at}]`. The
  installing moai version is PER FILE (REQ-006): the field records the moai
  build that successfully wrote that file; a failed write leaves the prior
  entry and its version untouched, so REQ-013's partial-failure continuation
  yields accurate mixed-version history. There is deliberately no top-level
  `moai_version` — a single top-level value cannot record that state. The
  `bundles:` list is the recorded opt-in selection (REQ-004, iter2 D18) —
  the state `moai bundle add|remove` mutates and `moai update` reads.
- Unknown-field preservation (iter2 D23; iter4 D27 extending its scope):
  EVERY manifest write carries through fields the writing binary does not
  understand (top-level and per-file), whatever schema_version it reads —
  including the append-only install/refresh writes REQ-021 permits against
  an UNKNOWN schema_version, which are exactly the older-binary-rewrites-
  newer-manifest case this rule exists to make safe — the
  consumers-ignore-unknown-fields premise of acceptance §D.7. An
  implementation whose decode would drop unknown fields refuses the write
  instead of silently shrinking the document.
- Reuses the sha256 hex convention of `catalog.yaml` entries and the triple-hash
  spirit of the project manifest (`internal/manifest/types.go`) minus the parts
  the user scope does not need (no 3-way merge at user scope: collision = skip).
- Read-modify-write serialization (round-5 F4): manifest mutation is
  serialized per user — a user-level lock (or equivalent single-writer
  discipline) spans manifest READ → asset changes → manifest SAVE. Without
  it, concurrent init/update/`moai bundle` operations from different
  projects of the same user interleave read-modify-write cycles and the
  last writer silently erases the earlier run's selections (bundle list
  entries, file records). The lock is a user-level concern (one lock file
  beside the manifest under `~/.moai/`), not a project-level one — the
  concurrent writers are different projects sharing one manifest.
- Schema-version gate: unknown `schema_version` → refuse manifest-driven
  removal (REQ-021); install/refresh may still proceed in append-only fashion.

### 2.3 Bundle taxonomy (catalog extension)

- `catalog.yaml` gains a user-install view: L0 core (REQ-003's content — the
  five agents per the resolved D-Q1: manager-spec, manager-develop,
  manager-docs, plan-auditor, sync-auditor; the moai-plan/moai-run/moai-sync
  published command skills per D-Q4; the hook payload deploys project-side;
  factory) + opt-in bundles (REQ-004). The existing `optional_packs`
  structure is the natural carrier for bundles: the six existing optional
  packs stand as-is, and the current-`core` entries that are NOT L0
  re-bundle by theme (resolved D-Q5, 2026-10-05 — leader default,
  operator-contestable). The published Codex command-skill set folds into
  the same catalog view (iter4 D29): the plan/run/sync three ride L0 (D-Q4)
  and the remaining fourteen take their bundles from the reclassification —
  `publishedSkillNames` does not survive as a second name list (the catalog
  is the single membership SSOT — the `publishedSkillNames` hardening
  pattern shows why drift guards exist).
- Factory (multi-lane operation) is L0 per the operator decision D3:
  `moai-factory-foreman` + `moai-lane-watchdog` ride the core bundle; the
  factory doctor check (`checkFactoryRun`) keeps reading project state as
  today.
- Selection surface (iter2 D18): bundle MEMBERSHIP is declared in the
  catalog (above); the user's opt-in SELECTION is the `bundles:` list in the
  per-user manifest (§2.2). `moai init --bundles <name,...>` sets the
  initial selection (default: L0 only); `moai bundle add <name>` / `moai
  bundle remove <name>` adjust the list and apply the install/removal of
  exactly that bundle's catalog entries (respecting REQ-010 collision and
  REQ-023 divergence semantics); `moai update` honors the recorded selection —
  installs/refreshes L0 + opted-in bundles, prunes per REQ-009 (no longer in
  L0 nor any opted-in bundle). Milestones: the `--bundles` init flag in M2;
  the `moai bundle` command and update honoring in M3.
- L0 transitive runtime skill closure (round-5 F1): catalog L0 entries
  carry PER-ENTRY skill dependencies, and the L0 view explicitly enumerates
  the resolved transitive closure — the installer copies the enumerated
  set; it never discovers dependencies by walking skill bodies at runtime.
  The closure, verified against the template sources on tree `064ff9960`
  (research §2b), is EIGHT skills in two tiers:
  - Tier 1 — invocation/preload (required for the flow to start): `moai`
    (the dispatcher — invoked by all three published command skills),
    `moai-foundation-core` (static preload of manager-spec, manager-develop,
    manager-docs), `moai-workflow-spec` (static preload of manager-spec),
    `moai-foundation-quality` (static preload of sync-auditor).
    plan-auditor declares NO static `skills:` preload (its body says so
    verbatim) — it contributes nothing to the static union.
  - Tier 2 — delegation-injected by the default flows' dispatcher routing
    rows: `moai-foundation-thinking` (plan row), `moai-workflow-tdd` and
    `moai-workflow-ddd` (run rows), `moai-workflow-project` (sync row).
  A catalog drift guard pins this enumeration to its sources — the agent
  frontmatter `skills:` unions, the dispatcher's routing-table Skills
  lines, and the command skills' dispatcher references — so the closure
  cannot rot silently (the same drift-guard pattern as the L0 list guard).
- User-side dispatcher mirror (round-5 F2): the Codex skill root
  `$HOME/.agents/skills/` carries `moai/SKILL.md` — the dispatcher mirror
  the published command skills reference — the user-folder successor of the
  retired project-side mirror; without it a post-M4 Codex CLI harness
  cannot resolve the `read .agents/skills/moai/SKILL.md` instruction, which
  is why the reference itself is also rebound at source (§2.5).

### 2.4 `moai update` integration

- New user-asset phase in the update flow: refresh (only when the file's
  current hash equals its manifest hash and differs from the shipped bytes) →
  removal (manifest-tracked files no longer in L0 nor any bundle recorded as
  opted-in in the manifest — the SELECTION-based criterion, verbatim
  REQ-009; the sole general removal rule — iter4 D28 re-keying the former
  catalog-based wording; precondition: current hash == manifest hash, or ==
  shipped bytes where a shipped source still exists, iter4 D25) →
  tracked-file divergence preserve + backup + report (REQ-023) → summary
  counts (installed/refreshed/removed/collision-skipped/divergence-preserved
  — REQ-011).
- Removal is manifest-driven ONLY: a file in a user folder that the manifest
  does not track is never a removal candidate (REQ-010 collision rule covers
  it; deletion of untracked files is out of scope).
- Upgrade branch (iter2 D14; iter4 D24 re-keying the removal gate per-asset
  — the former labels "init never ran on the machine" and "the same run has
  completed the first install" were both wrong keys: the gate is per-asset
  presence, not a run-level flag): the phase branches on project provenance
  AND on per-asset user-side presence. A project file is removed only when
  its user counterpart is CONFIRMED PRESENT — manifest-tracked with a
  current hash matching the installed bytes (REQ-020). The machine states:
  - (a) manifest ALREADY EXISTS (written by another project's update, or by
    a partial install) on a machine whose project still carries prior-model
    assets: the arm does not stall — missing counterparts are installed
    append-only (REQ-021-safe), present-and-matching counterparts are reused
    as-is, and removal proceeds per-asset.
  - (b) optional-pack assets in the upgrading project: the upgrade first
    install covers L0 ONLY (no manifest, no `--bundles` → the selection is
    L0 per §2.3's default). A non-L0 project asset has no confirmed user
    counterpart, so it STAYS project-side (reported) until the user opts
    into its bundle — never silently installed into a bundle the user did
    not pick, never removed while unconfirmed.
  - (c) partial first-install failure (REQ-013 continues past per-file
    failures): a failed user-side write leaves that file's project
    counterpart UN-REMOVED (its presence is unconfirmed), and the summary
    reports the skipped removal; on a retry the unconfirmed counterparts are
    re-attempted and the removal completes only for files whose counterparts
    then confirm.
  - No manifest AND no prior-model project assets → the advisory; there is
    nothing to install from and nothing to strand (the first install belongs
    to init, REQ-024).
- Ordering: the user-asset phase runs BEFORE the project phase so a
  user-side failure changes no project behavior beyond the per-asset gate
  itself; within the upgrade case the ordering PLUS the per-asset gate is
  the stranding guard — a project file's removal runs only after its own
  counterpart's install has landed and confirmed, so "a mid-update failure
  leaves the project phase untouched" reads per-asset: the affected file's
  counterpart stays project-side, the rest of the phase proceeds.

### 2.5 Project slimming

- Project deploy stops emitting common skills/agents entirely (both former
  modes): the project keeps settings, AGENTS.md/CLAUDE.md, the lock file
  (`.moai/manifest.json` unchanged in role), hooks, `.mcp.json`, output-styles,
  rules, command wrappers. "Command wrappers" names NON-SKILL command files
  only (iter4 D29): the 17 published Codex command skills (research V4) are
  common skills and move to user folders — the moai-plan/moai-run/moai-sync
  three via L0 (D-Q4), the remaining fourteen via the D-Q5 re-bundling (§2.3)
  — so AC-011's `no .agents/skills/moai*` placement ban holds unchanged over
  the post-M4 project tree.
- Dispatcher reference rebind at SOURCE level (round-5 F2): the published
  command skills are GENERATED artifacts — the command sources under
  `.claude/commands/moai/` are consumed READ-ONLY by the
  `internal/template/commandemit` emitter (`CommandsRoot:
  ".claude/commands/moai"`, commandemit.go:51), whose golden-pinned output
  is the committed `templates/.agents/skills/moai-<command>/SKILL.md` set;
  the project-relative `read .agents/skills/moai/SKILL.md` instruction
  (moai-plan/moai-run/moai-sync SKILL.md:9) is therefore fixed in the
  COMMAND SOURCES (to the user-folder path `~/.agents/skills/moai/
  SKILL.md`) and the copies are regenerated with `make commands-emit` —
  never hand-edited; the read-only drift check `commands-emit-check` rides
  the `build:` prerequisite chain (Makefile:34, :51-60) and would reject a
  hand-edited copy at the next build. The AGENTS.md skill-path sentences
  (AGENTS.md.tmpl:40-41 — "the deployed skill is in `.agents/skills/<name>/
  SKILL.md` for Codex and `.claude/skills/<name>/SKILL.md` for Claude")
  rebind the same way at source: post-M4 the Codex sentence names the
  user-folder dispatcher path.
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
  cited in the commit (REQ-019). The existing project-scope Codex asset
  diagnostics are repointed to the user-install path in M4 — the same
  milestone that stops emitting project assets — with a clean-on-correct-
  install regression test extending to the readiness output (iter1 D12
  residue, corrected by iter2 D19): the PROJECT-root readers are
  `inspectSkillMirror` (`internal/cli/doctor_codex.go:429`) and the readiness
  pair `probeCodexReadiness`/`countCodexAgentTOMLs`
  (`internal/cli/codex_readiness.go:131` consumer, `:215-217` definition —
  the agent-TOML count reads `.codex/agents/moai/*.toml` under the project
  root). `codexStaleSkillFinding` (`doctor_codex.go:857-870`) is NOT one of
  them: it reads user-layer `[[skills.config]]` entries and has no
  agent-count input — iter1's attribution to it was wrong; M5 judges whether
  that user-layer check needs its own repoint. Left alone the project-root
  readers would misreport every correct install as
  drift.

### 2.7 Plugin carrier disposition

- `internal/template/pluginemit/` (generator + 8 test files; 13 .go files
  total), the committed `plugins/moai/` tree, `.claude-plugin/marketplace.json`,
  `.agents/plugins/marketplace.json`, Makefile `plugin-emit`/`plugin-emit-check`
  targets (and their `build:` prerequisite), the roster-guard sweep-skip note
  (`internal/harness/rosterguard/check.go:338-341`), and the plugin install
  step (`internal/cli/plugin_install.go` + init/update call sites) go away.
- The release-chain gates go with them (iter1 D4): release.yml provenance
  Check 8 (`release.yml:123-128`, invoking `scripts/check-plugin-version.sh`,
  citing SPEC-PLUGIN-MARKETPLACE-001 REQ-024), `scripts/check-plugin-version.sh`
  (reads `plugins/moai/.claude-plugin/plugin.json`), and
  `scripts/check-plugin-discoverable.sh` (reads `.claude-plugin/marketplace.json`
  + `plugins/moai/.mcp.json`) — all deleted with the carrier (hard delete,
  P5); no orphaned release check remains (C5). SPEC-PLUGIN-MARKETPLACE-001
  REQ-024 is recorded as retired in spec.md §7.
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
| Wide retire blast radius (8 test files / 13 .go files, build chain, roster guard, release chain) | M6/M7 isolated; each deletion cites the owning REQ; golden tests deleted WITH the generator in the same commit; release-chain gates dispositioned in M6 (iter1 D4) |
| Users left with plugin installs from the old model | Migration report names the manual `claude plugin uninstall` step (doctor informational row); no silent divergence |
| Profile sessions not seeing user assets (D-Q3 closed: P6) | Declared limitation documented in M8; doc-visible so a later SPEC can lift it |
| User edits to tracked files silently overwritten or deleted (iter1 D3) | Refresh gated on current-hash == manifest-hash; removal on the one REQ-009 rule (manifest-hash OR shipped-bytes where a shipped source exists — iter4 D25); divergence → preserve (backup) + report + a REQ-011 count category; doctor's "modified" row lists it persistently |
| Manifest corruption / partial write | Atomic write (same pattern as `atomicWriteFile`); schema-version refusal for removal; corrupt-JSON recovery path (AC-021) |
