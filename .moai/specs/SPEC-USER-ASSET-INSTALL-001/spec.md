---
id: SPEC-USER-ASSET-INSTALL-001
title: "Install common skills and agents into per-user folders (no plugin carrier), slim the project payload to settings + AGENTS.md + lock file + project-only harness, and retire the pluginemit and deployer_mode surfaces"
version: "0.5.0"
status: draft
created: 2026-10-05
updated: 2026-10-05
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/template, internal/cli (init, update, doctor), internal/manifest"
lifecycle: spec-anchored
tags: "user-folder-install, assets, manifest, update, doctor, plugin-retirement, deployer-mode, bundles, factory, D3-D4"
tier: L
related_specs: [SPEC-PLUGIN-MARKETPLACE-001, SPEC-INIT-SHRINK-001, SPEC-CODEX-COMMAND-SKILLS-001]
---

# SPEC-USER-ASSET-INSTALL-001 — Common Assets to User Folders (D3/D4)

## HISTORY

- 2026-10-05: v0.1.0 created by manager-spec (card t1509, plan phase, Tier L).
  Operator decisions D3/D4 (v3.2 operator decisions, delivered with card
  t1509) are settled premises: (D3) L0 core includes factory; (D4) no plugin
  carrier — moai copies common skills and agents into per-user folders. Open
  sub-decisions are recorded in decision-index.md and referenced from §5.
- 2026-10-05: v0.2.0 plan-audit iter1 repair (card t1509, verdict FAIL 0.64).
  Defects D1-D13 closed: D1 decision register renumbered to D-Q1..D-Q6;
  D2 RED-now + green-path cells authored for every release-blocking AC
  (acceptance.md evidence ledger, baseline tree b965a3912); D3 tracked-file
  divergence preserve (REQ-023, REQ-008/009 gated on current-hash ==
  manifest-hash, REQ-011 fifth count, AC-006/008 rewritten); D4 release-chain
  gates dispositioned in M6 (release.yml Check 8, both check-plugin scripts);
  D5 hook payload defined project-deployed-but-counted-in-L0 (REQ-003);
  D6 first-install trigger bound to init (REQ-024); D7 dead gate options
  dropped, D-Q3/D-Q6 closed by constraint (premises P5/P6); D8 four-root
  confinement AC with parent-symlink sentinel (AC-025) + symlink-resolved
  confinement specified; D9 acceptance-layer claims corrected (GWT for
  Blockers, primary REQ marks, §3 map sync, corrupt-manifest AC clause);
  D10-D13 counts/V-row/Codex-doctor-repoint/per-file-manifest-version.
- 2026-10-05: v0.3.0 plan-audit iter2 repair (card t1509, verdict FAIL 0.75,
  12/13 iter1 defects verified resolved). Closed: iter1-D12 residue/D19 (M4
  repoint list corrected — the agent-count consumer is
  `probeCodexReadiness`/`countCodexAgentTOMLs`, codex_readiness.go:131/:217;
  the `codexStaleSkillFinding` attribution withdrawn); D14 upgrade-path
  stranding (REQ-024 upgrade arm + REQ-020 same-run gating, AC-020 extended —
  the 25-AC ceiling is why the Blocker upgrade criterion is an extension of
  AC-020, not a 26th AC); D15 AC-011 rebuilt on catalog-derived placements
  (the plain `moai` dirs); D16 REQ-023 truth table + `~/.moai/` backup home
  with the C2 carve-out; D17 confinement edges (resolved-root containment,
  leaf-symlink policy, write posture; AC-025 arms); D18 bundle selection
  surface (REQ-004 + design §2.3 + AC-018); D20 claims corrections; D21-D23
  optional nits taken. Evidence ledger re-executed in full (22 cells + 2
  positive controls, verbatim, tree cfb903358) and re-pinned.
- 2026-10-05: v0.4.0 iter4 delta round (card t1509; the ONE authorized round
  per the operator disposition on the iter3 ceiling hold). Closed D24-D34 +
  the EV-011 /bin/ls nit: D24 removal gate re-keyed per-asset (REQ-020/024,
  design §2.4, AC-020 arms); D25/D28 one removal rule (REQ-009 extended to
  manifest-hash OR shipped-bytes; the selection-based criterion restated as
  the sole general rule — design §2.4 re-keyed to it; AC-006/018 arms); D26
  TOCTOU claim downgraded "closed" → "narrowed" + declared limitation +
  AC-025 posture arm; D27 unknown-field preservation extended to ALL
  manifest writes (REQ-021, AC-021 arm); D29 the 17 published Codex command
  skills dispositioned (D-Q4: moai-plan/moai-run/moai-sync ride L0; the
  other fourteen re-bundle per D-Q5; the project-scope "command wrappers"
  lists scoped to non-skill command files); D30 ledger green paths matched
  to the matrix + per-cell flip expectations; D31 stale progress pin re-bound
  + M0 heading de-overstated; D32 repoint-clean/advisory-row folded into
  REQ-014/REQ-019; D33 backup home pinned
  (`~/.moai/backups/<root-slug>/<relpath>` + resolved-path judgment +
  sanitization); D34 C2 sole-write clause scoped to user-folder asset
  writes, the per-user manifest named as the SPEC's own state file. Founder
  gates adjudicated (operator/leader decision 2026-10-05): D-Q1 reading A
  (the plan→run→sync chain plus its two auditors); D-Q2 `~/.moai/
  user-assets.json` (leader default, operator-contestable); D-Q4 published
  command skills (leader default, operator-contestable); D-Q5 six packs
  stand + theme re-bundle (leader default, operator-contestable). Baseline
  re-pinned post-absorption: 6643c7bba → 51976e651 (develop a158b4b5f
  absorbed; load-bearing pins re-verified holding on this tree).
- 2026-10-05: v0.5.0 round-5 gate-fix round (card t1509; the codex review
  gate's four findings, operator-authorized disposition (i) before the gate
  re-run). F1 — L0's transitive runtime skill dependencies enumerated in the
  catalog L0 view (REQ-003/REQ-004, design §2.3 closure table + drift guard,
  AC-017 executable-flow arm "default-install-runs"); the closure verified
  against the agent bodies and dispatcher routing, not transcribed (manager-
  spec/develop/docs preload moai-foundation-core, manager-spec also
  moai-workflow-spec, sync-auditor moai-foundation-quality, plan-auditor
  declares NO static preload; the dispatcher `moai` is invoked by all three
  command skills; plan/run/sync delegation rows inject moai-foundation-
  thinking, moai-workflow-tdd/dd, moai-workflow-project). F2 — Codex
  dispatcher references rebind user-side at SOURCE level (REQ-001: the
  user-side dispatcher mirror `$HOME/.agents/skills/moai/`; design §2.5:
  the published command skills are GENERATED — sources
  `.claude/commands/moai/`, emitter `internal/template/commandemit`,
  regeneration `make commands-emit`, drift guard `commands-emit-check` in
  the build chain; AGENTS.md.tmpl skill-path sentences rebind; AC-002
  loading arm). F3 — init's trigger re-keyed per-asset-state (REQ-024,
  design §2.1, AC-001 partial-manifest retry arm). F4 — manifest
  read-modify-write serialized per user (REQ-006, design §2.2, AC-018
  concurrent arm). No new REQ (zero-new-REQ holds); all new verification
  absorbed into AC-001/002/017/018 (25/25 ceiling holds).

## 1. Background and Premise

MoAI-ADK today deploys its skill and agent assets per-project (local mode) or
carries them through a moai marketplace plugin (default mode, per
SPEC-PLUGIN-MARKETPLACE-001). The operator has decided (D4) to retire the
plugin carrier: moai itself copies the common skills and agents into the
USER's harness folders — Claude reads `~/.claude/skills` and
`~/.claude/agents`; Codex reads `$HOME/.agents/skills` and `~/.codex/agents`.
Projects keep only what is project-scoped: default settings, AGENTS.md, the
project lock file, and the project-only harness payload. The operator has
further decided (D3) that the L0 core bundle — installed for every user by
`moai init` — contains the plan/run/sync workflow surface, five core agents,
the hook payload, and the factory skill set (multi-lane operation), together
with L0's transitive runtime skill dependencies (the dispatcher and the
preload/delegation skills, enumerated in the catalog view); everything
else ships as opt-in bundles. L0's hook payload constituent deploys with the
project payload (REQ-005), never as a user-folder write (REQ-003).

A per-USER manifest file (with hashes) at `~/.moai/user-assets.json` (D-Q2)
makes the user-folder install accountable: `moai update` refreshes changed
files and removes files no longer in L0 or any opted-in bundle — but only
when the file on disk still matches what moai knows it wrote (its manifest
hash, or the shipped bytes where a shipped source still exists — REQ-009;
REQ-023 protects user edits to tracked files) — user-created files with
colliding names are never overwritten (they are reported), and `moai doctor`
compares the installed user tree against the manifest and the project tree
against the project lock file. The upgrade population — projects initialized
under the older per-project model — has no per-user install yet; their first
post-adoption `moai update` installs the missing user counterparts (REQ-024's
upgrade arm) and removes each project-side file only after its own
counterpart is confirmed present user-side (REQ-020's per-asset gate), so no
run leaves the user with neither — and a file whose counterpart could not be
confirmed (a failed write, or a non-L0 asset whose bundle is not opted in)
stays project-side and is reported.

Premises (settled by the operator, or forced by this SPEC's own constraints —
not re-opened during run phase):
- P1 (D4): no plugin, no marketplace — plain file copies from the binary's
  embedded assets.
- P2 (D3): L0 = plan/run/sync + 5 core agents + hook payload (project-deployed)
  + factory; the rest is opt-in bundles. Resolved readings (operator/leader
  decision 2026-10-05): "plan·run·sync" = the published command skills
  `moai-plan`/`moai-run`/`moai-sync` (D-Q4); the five core agents =
  manager-spec, manager-develop, manager-docs, plan-auditor, sync-auditor
  (D-Q1) — factory is already in L0 separately (this premise) and is not
  counted in the five.
- P3: user-created files are inviolable — skip and report.
- P4: per-profile settings folders stay per-profile.
- P5: the retired plugin carrier is hard-deleted atomically (no deprecation
  window). Forced by REQ-016 and C5: any window keeps the carrier generating,
  committing, or shipping, which those rules forbid.
- P6: profile sessions do not see the shared user assets in v1 (declared
  limitation, doc-visible). Forced by REQ-002 and C2: every
  install-into-profile mechanism writes into a `CLAUDE_CONFIG_DIR` profile,
  which is never an install target. A later SPEC may lift the limitation.

## 2. Requirements (GEARS)

### Install model

- REQ-001: The system shall install common skills and agents as plain file
  copies into per-user folders — Claude: `~/.claude/skills` and
  `~/.claude/agents`; Codex: `$HOME/.agents/skills` and `~/.codex/agents` —
  with no plugin or marketplace carrier of any kind. The Codex skill root
  carries the user-side dispatcher mirror (`$HOME/.agents/skills/moai/`)
  that the published command skills reference — the user-folder successor
  of the retired project-side mirror (round-5 F2); the published command
  sources reference user-folder paths, regenerated at source level per
  design §2.5 and never hand-edited.
- REQ-002: The system shall keep per-profile settings folders per-profile: a
  `CLAUDE_CONFIG_DIR` profile directory holds its own settings and is never a
  target of the shared user-asset install.
- REQ-003: The L0 core bundle shall contain the plan/run/sync workflow
  surface — the published command skills `moai-plan`, `moai-run`,
  `moai-sync` (D-Q4) — the five core agents (manager-spec, manager-develop,
  manager-docs, plan-auditor, sync-auditor — D-Q1), the hook payload, and
  the factory skill set; the hook payload deploys with the project payload
  (REQ-005) and is never a user-folder write target — the four roots of C2
  carry no hook destination. L0's content includes its TRANSITIVE runtime
  skill dependencies, explicitly enumerated in the catalog's L0 view — the
  `moai` dispatcher skill the three command skills invoke, the agents'
  preload skills, the default flows' delegation-injected workflow skills,
  and the agents' on-demand `Skill()` invoke sites (the per-mission
  moai-ref-*/moai-domain-* domain injections are explicitly classified out
  — design §2.3 closure table) — so the default plan/run/sync flow loads
  everything it calls from the installed set alone (round-5 F1 + fold B1,
  "default-install-runs").
- REQ-004: The system shall ship every common asset outside L0 as an opt-in
  bundle; a bundle is installed or removed as a unit, bundle membership is
  declared in the shipped catalog — entries carrying their skill
  dependencies, so the installer resolves L0's transitive closure from the
  catalog's explicit enumeration, never by discovering dependencies at
  runtime (round-5 F1) — the opt-in selection is recorded in the per-user
  manifest — set by `moai init --bundles` and adjusted by `moai bundle
  add|remove` — and `moai update` honors the recorded selection.
- REQ-005: `moai init` shall deploy to the project only the default settings,
  AGENTS.md, the project lock file, and the project-only harness payload
  (hooks and `.mcp.json` included); it shall not copy any common skill or
  agent file into the project.

### Per-user manifest

- REQ-006: The system shall maintain a per-user manifest recording every
  user-folder file it installed, carrying the sha256 hash of the installed
  bytes, the owning bundle, and the moai version that installed it.
  Manifest read-modify-write is serialized per user — a user-level lock (or
  equivalent) spans manifest read → asset changes → manifest save — so
  concurrent init/update/bundle operations from different projects cannot
  lose one another's writes (round-5 F4).
- REQ-007: The user-asset install shall operate offline: the binary carries
  every asset it installs, and no network access is part of the install,
  refresh, or removal path.

### moai update

- REQ-008: When `moai update` runs against a manifest-tracked user-folder file
  whose current hash equals its manifest hash while differing from the shipped
  bytes, the system shall refresh the file to the shipped bytes and record the
  new hash and installing version in the manifest.
- REQ-009: When a file recorded in the per-user manifest is no longer part of
  L0 or of any bundle recorded as opted-in in the manifest (the
  selection-based criterion — REQ-004), and its current hash equals its
  manifest hash, or equals the shipped bytes where a shipped source for the
  file still exists (for a file dropped from every bundle no shipped source
  exists, so only the manifest-hash alternative applies — iter4 D25), `moai
  update` shall remove it from the user folder and from the manifest; a file
  matching neither hash is preserved and reported (REQ-023 divergence).
- REQ-010: When the target path of an install, refresh, or removal holds a
  file the per-user manifest does not track, the system shall leave that file
  untouched and report the collision.
- REQ-011: When a `moai update` run completes its user-asset phase, the system
  shall report the counts of installed, refreshed (including manifest-only
  repairs of the REQ-023 truth table), removed, collision-skipped, and
  divergence-preserved files.
- REQ-012: The user-asset install shall be idempotent: repeating it against an
  already-current tree changes no file and reports zero deltas.
- REQ-013: When a user-folder write fails, the system shall continue
  processing the remaining files and surface every failure in the run summary.

### moai doctor

- REQ-014: `moai doctor` shall compare every file recorded in the per-user
  manifest against the installed user-folder tree and report missing,
  modified, and untracked entries; the project-scope Codex asset diagnostics
  repointed in M4 (`inspectSkillMirror`, the `probeCodexReadiness`/
  `countCodexAgentTOMLs` pair) shall report a correct user install as clean
  — no false drift once project assets stop emitting (iter4 D32 fold).
- REQ-015: `moai doctor` shall compare the project tree against the project
  lock file and report drift in both directions.

### Retired carrier disposition

- REQ-016: The system shall no longer generate, commit, or ship the moai
  plugin artifacts — the marketplace manifests, the plugin manifests, and the
  plugin payload tree.
- REQ-017: The init and update flows shall not invoke any plugin marketplace
  or plugin install command for either harness.
- REQ-018: The deploy-mode split — the plugin payload selection and the plugin
  mirror policy — shall be retired from the deployer surface.
- REQ-019: `moai doctor` shall no longer report plugin deployment or plugin
  version as checks of the retired carrier; each is repointed to the
  user-manifest comparison or removed with its owning requirement cited.
  Doctor shall carry the migration advisory row — the manual `claude plugin
  uninstall` step for prior plugin installs (design §4) — as an
  informational row (iter4 D32 fold).

### Migration and compatibility

- REQ-020: When a project carries template-managed common skills or agents
  from an earlier deployment, `moai update` shall offer and apply the
  migration that removes each project-side file only after its user
  counterpart is confirmed present — manifest-tracked with a hash matching
  the installed bytes (the per-asset removal gate; iter4 D24 re-keying
  REQ-024's upgrade arm) — preserving user-modified and user-created files
  and reporting each disposition; a file whose counterpart failed to
  install (REQ-013 per-file failure) or lies outside L0 while its bundle is
  not opted in stays project-side and is reported.
- REQ-021: The per-user manifest shall carry a schema version; the system
  shall refuse manifest-driven removal against an unknown schema version
  while still permitting append-only install and refresh; EVERY manifest
  write — under a known or an unknown schema version — shall preserve
  fields it does not understand (the consumers-ignore-unknown-fields
  premise of acceptance §D.7; this is what makes the append-only permission
  safe against an older binary rewriting a newer manifest — iter4 D27), and
  an implementation that would drop unknown fields refuses the write
  instead.

### Codex agent parity

- REQ-022: Codex agent definitions shall follow the same install, refresh,
  divergence, collision, and removal rules as skills, landing in
  `~/.codex/agents`.

### Tracked-file divergence

- REQ-023: When a file tracked in the per-user manifest has a current hash
  equal to neither its manifest hash nor the shipped bytes, the system shall
  preserve the installed file — backing up the shipped replacement to the
  backup home `~/.moai/backups/<root-slug>/<relpath>` (root-slug-prefixed
  layout per design §2.1 — iter4 D33; C2's sole out-of-root write carve-out
  for user-folder asset writes) when
  shipped bytes exist, and omitting the backup when the file is dropped from
  every bundle, where no shipped bytes exist — leave the tracked path
  unmodified by refresh and by removal, and report the divergence; the
  remaining states follow the design §2.1 truth table: manifest-stale (current
  equals the shipped bytes while differing from the manifest hash) is repaired
  in the manifest without a file rewrite and counted as refreshed, and a
  tracked file missing on disk is reinstalled at refresh or dropped from the
  manifest at removal.

### First-install trigger

- REQ-024: When `moai init` runs, the system shall ensure every L0 and
  opted-in-bundle asset is present user-side before the run reports
  success — the judgment is PER-ASSET-STATE applying the REQ-023 truth
  table exactly as update does: an ABSENT tracked target is installed; a
  present tracked file is refreshed, manifest-repaired, or
  divergence-preserved per the table (a user-edited file's bytes are never
  clobbered by init — round-5 fold A1); an untracked target is a REQ-010
  collision. A manifest left by a PARTIAL install does not suppress the
  run: init completes the shortfall idempotently, sharing the
  shortfall-append semantics with the update upgrade arm below (round-5
  F3). When `moai update` runs
  against a project carrying prior-model common skills or agents (the
  REQ-020 upgrade population — init ran under the pre-SPEC model), the run
  shall confirm every user counterpart REQ-020 will remove is present
  user-side — manifest-tracked with a hash matching the installed bytes —
  installing the missing L0 counterparts in the same run and before
  REQ-020's project-side removal (the upgrade first install), so no run
  leaves the user with neither placement; a manifest that already exists
  (written by another project's update or by a partial install) does not
  suppress the arm — missing counterparts are installed append-only (iter4
  D24); the upgrade first install covers L0 only — project assets outside
  L0 are not installed until the user opts into their bundle, and they are
  not removed while unconfirmed; subsequent `moai update` runs refresh and
  prune that install (REQ-008/009) rather than performing the first
  install.

## 3. Acceptance Criteria (summary)

The authoritative matrix lives in `acceptance.md` (AC-001..AC-025, 25
criteria, each binary-testable; Blocker criteria carry explicit
Given-When-Then renderings there, plus RED-now + green-path cells for every
release-blocking criterion). Coverage map: REQ-001 → AC-001/002/012/025;
REQ-002 → AC-019; REQ-003 → AC-017; REQ-004 → AC-018; REQ-005 → AC-011/012;
REQ-006 → AC-003; REQ-007 → AC-024; REQ-008 → AC-005; REQ-009 → AC-006;
REQ-010 → AC-007/008; REQ-011 → AC-022; REQ-012 → AC-004; REQ-013 → AC-023;
REQ-014 → AC-009; REQ-015 → AC-010; REQ-016 → AC-013; REQ-017 → AC-014;
REQ-018 → AC-015; REQ-019 → AC-016; REQ-020 → AC-020; REQ-021 → AC-006/021;
REQ-022 → AC-002; REQ-023 → AC-006/008; REQ-024 → AC-001/002/020.

## 4. Constraints

- C1: The install, refresh, and removal paths are offline and deterministic —
  the same binary version against the same tree produces the same result.
- C2: Every user-folder destination is confined to the four declared roots
  (`~/.claude/skills`, `~/.claude/agents`, `$HOME/.agents/skills`,
  `~/.codex/agents`); confinement is judged on resolved paths — each root is
  resolved once at install start (a symlinked root, e.g. a dotfile-manager
  `~/.claude`, is a legal boundary at its resolved location), each
  destination is resolved immediately before write, and the destination's
  resolved path must remain inside its own root's resolved tree; a leaf that
  resolves outside its root is refused at install and classified as
  divergence (REQ-023 preserve + report) at refresh/removal, never written
  through; the sole permitted write destination outside the four roots for
  USER-FOLDER ASSET writes is the REQ-023 backup home
  `~/.moai/backups/<root-slug>/<relpath>`, judged on resolved paths and
  sanitized like the four roots (iter4 D33); the per-user manifest
  (`~/.moai/user-assets.json`, D-Q2) is the SPEC's own state file — a
  second, separately named out-of-root write destination that is not a
  user-folder asset write (iter4 D34); the resolve-then-write window is
  NARROWED by the design §2.1 write posture (temp file + atomic rename
  inside the validated resolved directory, the resolved parent re-validated
  immediately before the rename) — a parent directory swapped to an
  outside-pointing symlink after that re-validation remains a declared race
  limitation, not a closed window (iter4 D26).
- C3: The project lock file (`.moai/manifest.json`) keeps its existing role
  and schema; this SPEC extends doctor's READING of it, not its format.
- C4: User-facing collision, divergence, and failure reports are actionable:
  each names the path, the reason, and the suggested action.
- C5: The retired plugin artifacts are removed together with their golden
  tests, Makefile targets, build-chain wiring, and release-workflow checks in
  the same change — no orphaned drift check may remain.
- C6: The migration path never deletes a file classified `user_modified` or
  `user_created` without an explicit operator-facing report; `user_modified`
  files are preserved (backup + report), not silently replaced.
- C7: TRUST 5 gates apply (TDD mode is the project's `development_mode`);
  every REQ above carries at least one automated verification.

## 5. Open Decisions

Recorded in `decision-index.md`. The four founder gates were adjudicated
2026-10-05 (operator/leader decision relayed with the iter4 authorization)
and carry their verdicts in the register:
- D-Q1 (gate for REQ-003; was BLOCKING M0/M1): RESOLVED — the five core
  agents are manager-spec, manager-develop, manager-docs, plan-auditor,
  sync-auditor (reading A: the plan→run→sync chain plus its two auditors;
  factory is separately in L0 per P2 and is not counted in the five).
- D-Q2 (gate for REQ-006; was BLOCKING M0/M1): RESOLVED —
  `~/.moai/user-assets.json` (leader default, operator-contestable).
- D-Q4 (input to REQ-003): RESOLVED — "plan·run·sync" names the published
  command skills `moai-plan`/`moai-run`/`moai-sync`, NOT the
  `moai-workflow-*` skills (leader default, operator-contestable).
- D-Q5: RESOLVED — the six existing optional packs stand as-is;
  current-core remainders re-bundle by theme (leader default,
  operator-contestable).

Resolved at plan phase by constraint (iter1 repair D7; recorded as premises
P5/P6 and closed in the decision register):
- D-Q3 (profile visibility): profiles do not see the shared user assets in v1
  (declared limitation) — every install-into-profile option violated REQ-002
  and C2. Doc-visible per acceptance §D.7.
- D-Q6 (disposition depth): hard delete of the retired plugin surfaces — the
  deprecation-window option violated REQ-016 and C5.

## 6. Non-goals and Out of Scope

### Out of Scope — plugin replacement mechanisms

- No new package-manager, marketplace, or registry mechanism replaces the
  plugin: the card is explicit that the copy IS the distribution.
- No network fetching, version resolution, or dependency resolution at install
  time (REQ-007's offline premise).

### Out of Scope — non-asset harness surfaces

- Output styles, rules, workflows, and command wrappers stay project-scoped in
  this SPEC; only skills and agents move to user folders. "Command wrappers"
  names non-skill command files only — the 17 published Codex command skills
  (research V4) are common skills and move to user folders: the plan/run/sync
  three via L0 (D-Q4), the remaining fourteen re-bundled per D-Q5 (design
  §2.5). The hook payload is part of L0 (P2) but deploys project-side
  (REQ-003/REQ-005). [NEEDS CLARIFICATION is NOT raised: the card names
  skills/agents/hooks/factory only.]
- Shell-hook scripts stay project-deployed; their packaging follows the L0
  hook payload definition but no new hook execution model is introduced.

### Out of Scope — Claude Code / Codex runtime behavior

- Teaching either harness new discovery paths beyond the four documented
  folders is out of scope; the SPEC works within the folders the harnesses
  already scan.
- Codex leader-mode, cross-session messaging, or any parity gap listed in the
  v3.2 parity matrix remains out of scope.

### Out of Scope — cleanup of unrelated surfaces

- The stale `internal/template/CLAUDE.md` namespace-convention paragraph
  (`agents/{core,expert,meta}` vs the actual `agents/moai/` tree) is corrected
  only if the run phase touches that file anyway; it is not this SPEC's
  deliverable.
- The AGENTS.md/CLAUDE.md content redesign (the t1509 note set's D1/D2 work)
  is a separate card; this SPEC only preserves the deployed instruction files
  in the project payload.

## 7. Dependencies and Prior Art

- SPEC-PLUGIN-MARKETPLACE-001 (completed): introduced the plugin carrier this
  SPEC retires; its doctor checks (REQ-020..023 there) are dispositioned in
  REQ-019 here, and its release-workflow gate (REQ-024 there — release.yml
  Check 8, `scripts/check-plugin-version.sh`) plus the discoverability check
  it contracts (`scripts/check-plugin-discoverable.sh`, REQ-025 carve-out
  there) are retired with the carrier in M6 under REQ-016/C5 — no orphaned
  release check remains.
- SPEC-INIT-SHRINK-001 (completed): introduced the deploy-mode split
  (`deployer_mode.go`) and the mirror policy this SPEC retires (REQ-018), and
  the collision/rehome semantics (V17) this SPEC inherits at user scope.
- SPEC-CODEX-COMMAND-SKILLS-001 (completed): the published command skills and
  their protected-skip rule (R-011) — the user install carries the same
  namespace.
- `internal/manifest` (ADR-007): the provenance model (template_managed /
  user_modified / user_created / deprecated) reused conceptually for the
  migration classification (REQ-020).
