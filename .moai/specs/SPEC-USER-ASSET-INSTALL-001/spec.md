---
id: SPEC-USER-ASSET-INSTALL-001
title: "Install common skills and agents into per-user folders (no plugin carrier), slim the project payload to settings + AGENTS.md + lock file + project-only harness, and retire the pluginemit and deployer_mode surfaces"
version: "0.3.0"
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
the hook payload, and the factory skill set (multi-lane operation); everything
else ships as opt-in bundles. L0's hook payload constituent deploys with the
project payload (REQ-005), never as a user-folder write (REQ-003).

A per-USER manifest file (with hashes) makes the user-folder install
accountable: `moai update` refreshes changed files and removes files no longer
in L0 or any opted-in bundle — but only when the file on disk still matches
what moai last wrote (REQ-023 protects user edits to tracked files) —
user-created files with colliding names are never overwritten (they are
reported), and `moai doctor` compares the installed user tree against the
manifest and the project tree against the project lock file. The upgrade
population — projects initialized under the older per-project model — has no
per-user install yet; their first post-adoption `moai update` performs the
user install (REQ-024's upgrade arm) BEFORE the project slimming removes the
old placement, so no run leaves the user with neither.

Premises (settled by the operator, or forced by this SPEC's own constraints —
not re-opened during run phase):
- P1 (D4): no plugin, no marketplace — plain file copies from the binary's
  embedded assets.
- P2 (D3): L0 = plan/run/sync + 5 core agents + hook payload (project-deployed)
  + factory; the rest is opt-in bundles.
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
  with no plugin or marketplace carrier of any kind.
- REQ-002: The system shall keep per-profile settings folders per-profile: a
  `CLAUDE_CONFIG_DIR` profile directory holds its own settings and is never a
  target of the shared user-asset install.
- REQ-003: The L0 core bundle shall contain the plan/run/sync workflow
  surface, the five core agents as resolved by decision gate D-Q1, the hook
  payload, and the factory skill set; the hook payload deploys with the
  project payload (REQ-005) and is never a user-folder write target — the
  four roots of C2 carry no hook destination.
- REQ-004: The system shall ship every common asset outside L0 as an opt-in
  bundle; a bundle is installed or removed as a unit, bundle membership is
  declared in the shipped catalog, the opt-in selection is recorded in the
  per-user manifest — set by `moai init --bundles` and adjusted by `moai
  bundle add|remove` — and `moai update` honors the recorded selection.
- REQ-005: `moai init` shall deploy to the project only the default settings,
  AGENTS.md, the project lock file, and the project-only harness payload
  (hooks and `.mcp.json` included); it shall not copy any common skill or
  agent file into the project.

### Per-user manifest

- REQ-006: The system shall maintain a per-user manifest recording every
  user-folder file it installed, carrying the sha256 hash of the installed
  bytes, the owning bundle, and the moai version that installed it.
- REQ-007: The user-asset install shall operate offline: the binary carries
  every asset it installs, and no network access is part of the install,
  refresh, or removal path.

### moai update

- REQ-008: When `moai update` runs against a manifest-tracked user-folder file
  whose current hash equals its manifest hash while differing from the shipped
  bytes, the system shall refresh the file to the shipped bytes and record the
  new hash and installing version in the manifest.
- REQ-009: When a file recorded in the per-user manifest is no longer part of
  L0 or of any bundle recorded as opted-in in the manifest, and its current
  hash equals its manifest hash, `moai update` shall remove it from the user
  folder and from the manifest.
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
  modified, and untracked entries.
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

### Migration and compatibility

- REQ-020: When a project carries template-managed common skills or agents
  from an earlier deployment, `moai update` shall offer and apply the
  migration that removes them from the project only after the same run has
  completed the user-side first install (REQ-024 upgrade arm), preserving
  user-modified and user-created files and reporting each disposition.
- REQ-021: The per-user manifest shall carry a schema version; the system
  shall refuse manifest-driven removal against an unknown schema version while
  still permitting append-only install and refresh; a manifest write performed
  against a known schema version shall preserve fields it does not understand
  (the consumers-ignore-unknown-fields premise of acceptance §D.7) — an
  implementation that would drop unknown fields refuses the write instead.

### Codex agent parity

- REQ-022: Codex agent definitions shall follow the same install, refresh,
  divergence, collision, and removal rules as skills, landing in
  `~/.codex/agents`.

### Tracked-file divergence

- REQ-023: When a file tracked in the per-user manifest has a current hash
  equal to neither its manifest hash nor the shipped bytes, the system shall
  preserve the installed file — backing up the shipped replacement to the
  backup home under `~/.moai/` (C2's sole out-of-root write carve-out) when
  shipped bytes exist, and omitting the backup when the file is dropped from
  every bundle, where no shipped bytes exist — leave the tracked path
  unmodified by refresh and by removal, and report the divergence; the
  remaining states follow the design §2.1 truth table: manifest-stale (current
  equals the shipped bytes while differing from the manifest hash) is repaired
  in the manifest without a file rewrite and counted as refreshed, and a
  tracked file missing on disk is reinstalled at refresh or dropped from the
  manifest at removal.

### First-install trigger

- REQ-024: When `moai init` runs on a machine that has no per-user install,
  the system shall install the L0 core bundle and every opted-in bundle into
  the user folders before the run reports success; when `moai update` runs on
  a machine with no per-user install whose project carries prior-model common
  skills or agents (the REQ-020 upgrade population — init ran under the
  pre-SPEC model), it shall perform that same first install in the same run
  and before REQ-020's project-side removal, so no run leaves the user with
  neither placement; subsequent `moai update` runs refresh and prune that
  install (REQ-008/009) rather than performing the first install.

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
  through; the sole permitted write destination outside the four roots is the
  REQ-023 backup home under `~/.moai/`; the resolve-then-write window is
  closed by the design §2.1 write posture (temp file + atomic rename inside
  the validated resolved directory).
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

Recorded in `decision-index.md`; the open gates may not be silently decided
during run phase:
- D-Q1 (gate for REQ-003; BLOCKS M0/M1): the exact five L0 core agents.
- D-Q2 (gate for REQ-006; BLOCKS M0/M1): the per-user manifest location and
  file name.
- D-Q4 (input to REQ-003): whether "plan·run·sync" names the published command
  skills, the workflow skills, or both.
- D-Q5: bundle granularity — whether the six existing optional packs stand as
  the bundles or current-core remainders re-bundle differently.

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
  this SPEC; only skills and agents move to user folders. The hook payload is
  part of L0 (P2) but deploys project-side (REQ-003/REQ-005). [NEEDS
  CLARIFICATION is NOT raised: the card names skills/agents/hooks/factory
  only.]
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
