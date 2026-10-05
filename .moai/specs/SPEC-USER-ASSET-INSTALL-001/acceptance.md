---
id: SPEC-USER-ASSET-INSTALL-001
title: "acceptance.md — acceptance criteria matrix"
version: "0.4.0"
created: 2026-10-05
updated: 2026-10-05
author: manager-spec
---

# Acceptance — SPEC-USER-ASSET-INSTALL-001

Verification layer: every criterion is binary-testable; the Blocker criteria
carry explicit Given-When-Then renderings in §D.1. Requirement layer (GEARS
obligations) lives in spec.md §2; the REQ↔AC map is spec.md §3. All
user-folder tests run against isolated temp HOMEs (plan §D).

Two-cell discipline (`.claude/rules/moai/development/verification-completeness.md`
§2): every release-blocking criterion (7 Blocker-class + 14 Major-class, plus
new Blocker AC-025) carries a RED-now cell — observed on the
pre-implementation tree and carried as an evidence-ledger entry (§D.2b) with
command, verbatim stdout, exit code, and tree SHA — and a green-path cell
naming the milestone that flips it. All RED cells were executed by
manager-spec: first on the baseline tree
`b965a3912c0e97ef81aeeea773019e633591e1cd` (branch `WT-user-asset-copy`,
status clean) in the iter1-repair audit run, then RE-EXECUTED IN FULL — all
22 cells plus both positive controls, outputs byte-identical — on tree
`cfb9033582eff27f9031e1a6438d8558aaa48115` in the iter2-repair audit run
(the source bytes are identical between the two trees: both intervening
commits touch only SPEC artifacts); no cell carries an invented
value. Minor criteria (AC-019/023/024) are not release-blocking and carry
green paths only.

## D. AC Matrix

| AC | Verifies (primary first) | Binary test | RED now | Green path |
|---|---|---|---|---|
| AC-001 | REQ-001, REQ-024 | After `moai init` on a fresh temp HOME: L0 skill dirs exist under `$HOME/.claude/skills/` with bytes matching the per-user manifest hashes | EV-001 | M2 |
| AC-002 | REQ-001, REQ-022, REQ-024 | After the same init: `$HOME/.agents/skills/<skill>/SKILL.md` + `~/.codex/agents/<name>.toml` exist for the L0 set | EV-002 | M2 |
| AC-003 | REQ-006 | Per-user manifest written; every installed path carries sha256 + bundle + per-file moai version | EV-003 | M1+M2 |
| AC-004 | REQ-012 | Second install run: zero file writes (mtime/hash proof), zero-delta report | EV-004 | M2 |
| AC-005 | REQ-008 | Shipped-byte change where current hash == manifest hash → update rewrites the tracked file; hash + version refreshed | EV-005 | M3 |
| AC-006 | REQ-009, REQ-023, REQ-021 | Dropped-from-bundles tracked file (current == manifest hash) → removed; dropped file with diverged hash → preserved in place + reported with NO shipped-bytes backup (none exists); manifest-stale file (current == shipped, ≠ manifest) at removal → removed under REQ-009's shipped-bytes alternative (iter4 D25); tracked file missing on disk at removal → manifest entry dropped + counted removed; foreign schema_version → removal refused while install/refresh proceed | EV-006 | M3 |
| AC-007 | REQ-010 | Untracked file at an install target → left byte-identical, collision reported | EV-007 | M2 |
| AC-008 | REQ-023, REQ-010 | Tracked file whose current hash matches neither manifest nor shipped bytes → preserved, shipped replacement backed up under `~/.moai/`, path unmodified by refresh, divergence reported; manifest-stale file (current == shipped ≠ manifest) → manifest repaired, no file rewrite, counted under refreshed; tracked file missing on disk → reinstalled at refresh | EV-008 | M3 |
| AC-009 | REQ-014 | Doctor user-install check: detects manifest-tracked file deleted / hash-modified / untracked entry in the four roots; the repointed project-scope Codex diagnostics (`inspectSkillMirror`, `probeCodexReadiness`/`countCodexAgentTOMLs`) report a correct user install as clean (no false drift after M4) | EV-009 | M4+M5 |
| AC-010 | REQ-015 | Doctor project-vs-lock check: detects project file absent from lock AND lock entry absent from project | EV-010 | M5 |
| AC-011 | REQ-005 | Post-init project tree contains NO catalog-derived common-asset placement: no `.claude/skills/moai*` directory (incl. the plain `moai` pack dir), no `.claude/agents/moai/`, no `.codex/agents/moai/`, no `.agents/skills/moai*` — the ban holds because the 17 published Codex command skills move user-side (D-Q4/D-Q5, design §2.5; the project-scope "command wrappers" list names non-skill files only — iter4 D29) | EV-011 | M4 |
| AC-012 | REQ-005, REQ-001 | Post-init project tree contains settings, AGENTS.md, lock file, hooks, `.mcp.json` (with moai MCP entry) | EV-012 | M4 |
| AC-013 | REQ-016 | Repo tree carries no `plugins/moai/`, no marketplace manifests; `make build` green without plugin-emit targets; boundary grep over `.github/workflows/` + `scripts/` + Makefile: zero references to the deleted check scripts (`check-plugin-version.sh`, `check-plugin-discoverable.sh`) | EV-013 | M6 |
| AC-014 | REQ-017 | Boundary grep: init/update paths hold zero plugin marketplace/install invocations | EV-014 | M6 |
| AC-015 | REQ-018 | Boundary grep + build: zero `DeployModePlugin`/`PluginMirrorPolicy` references; single deploy payload shape | EV-015 | M7 |
| AC-016 | REQ-019 | Doctor output carries no "Plugin Deployment"/"Plugin Version" carrier rows; each removed or repointed per REQ-019 with the owning requirement cited in the commit (hard delete, P5); the migration advisory row (manual `claude plugin uninstall` step for prior plugin installs, design §4) is present as a doctor informational row | EV-016 | M5+M6 |
| AC-017 | REQ-003 | Installed L0 user-folder set equals the resolved gate answer (5 agents by name, plan/run/sync surface, factory); the hook payload is project-deployed (AC-012), not user-folder content | EV-017 | M0+M2 |
| AC-018 | REQ-004 | Via `moai init --bundles` and the `moai bundle` add/remove commands: bundle install adds exactly the bundle's catalog entries, bundle removal takes exactly them, the manifest's bundle list reflects each change, and `moai update` honors the recorded selection — a file of a shipped-but-DESELECTED bundle (the artifact of `moai bundle remove`) is pruned under the selection-based criterion (REQ-009/design §2.3), not kept (iter4 D28 flip-criterion arm) | EV-018 | M0+M2+M3 |
| AC-019 | REQ-002 | Profile dirs (`~/.moai/claude-profiles/<name>`) byte-unchanged by install/update (settings isolation) | n/a (Minor) | M2/M3 |
| AC-020 | REQ-020, REQ-024 | Migration: template-managed project common skills AND agents removed (incl. the plain `moai` dirs); user-modified preserved + reported; user-created untouched; each removal is gated per-asset on its user counterpart being manifest-tracked with a matching hash (REQ-024 upgrade arm) — after the run the user holds the user-folder placement (no neither-state). Three machine-state arms (iter4 D24): (a) manifest ALREADY exists (another project's update / partial install) → missing counterparts installed append-only before their project-side removal, no stall, no neither-state; (b) optional-pack (non-L0) template-managed project asset with no manifest and no `--bundles` → stays project-side, reported, not installed into an unopted bundle, not removed; (c) a user-side write FAILS mid-upgrade (read-only dir fixture) → the failed file's project counterpart is NOT removed (stays + reported), remaining files complete, summary lists the failure | EV-020 | M4 |
| AC-021 | REQ-021 | Manifest with foreign schema_version → install/refresh proceed, removal refuses with named error; corrupt manifest JSON → removal refuses, corruption reported, rebuild-from-scan offered (doctor informational), never auto-delete; a manifest write carrying unknown fields preserves them (or refuses the write) — under a KNOWN schema AND under a FOREIGN schema_version on an append-only install/refresh write (the older-binary-rewrites-newer-manifest round-trip; iter4 D27 arm) | EV-021 | M3 |
| AC-022 | REQ-011 | Mixed update run summary: installed/refreshed/removed/collision-skipped/divergence-preserved counts each match a seeded fixture | EV-022 | M3 |
| AC-023 | REQ-013 | Read-only target dir → remaining files still processed; summary lists the failure with path + reason | n/a (Minor) | M2 |
| AC-024 | REQ-007 | Install executes with network disabled (no dial in trace); same output as online run | n/a (Minor) | M2 |
| AC-025 | REQ-001 (C2 confinement) | Installer refuses a destination outside the four roots (refusal names the path, nothing written); a parent-symlink sentinel (destination parent chain symlinked outside the roots) is refused after symlink resolution — sentinel target unmodified; a symlinked ROOT installs into its resolved location and an escape from the RESOLVED root is refused; a managed leaf replaced by an outside-pointing symlink is never written through (refused at install, divergence-classified at refresh/removal); a REQ-023 backup write whose resolved destination escapes the resolved backup home `~/.moai/backups/<root-slug>/` is refused (iter4 D33 arm); the write posture itself is asserted — every user-folder write goes through a temp file inside the validated resolved directory plus an atomic rename with the resolved parent re-validated immediately before it (iter4 D26 posture arm — the POST-validation parent swap is the declared race limitation, design §2.1: a true mid-flight swap fixture would be an inherently racy, nondeterministic test at this layer, so the deterministic assertion is the posture, not the race) | EV-023 | M2 |

## D.1 Severity and Blocker Given-When-Then

Severity classes:
- Blocker: AC-001, AC-002, AC-003, AC-007, AC-011, AC-013, AC-020, AC-025
- Major: AC-004, AC-005, AC-006, AC-008, AC-009, AC-010, AC-012, AC-014,
  AC-015, AC-016, AC-017, AC-018, AC-021, AC-022
- Minor: AC-019, AC-023, AC-024

Explicit Given-When-Then renderings for every Blocker criterion:

- **AC-001** — Given a fresh temp HOME with no prior per-user install and a
  completed `moai init` (REQ-024); When the L0 skill set is compared against
  `$HOME/.claude/skills/`; Then every L0 skill directory exists and its file
  bytes hash (sha256) to the per-user manifest's recorded value.
- **AC-002** — Given the same fresh-HOME init; When the Codex roots are
  inspected; Then `$HOME/.agents/skills/<skill>/SKILL.md` exists for each L0
  skill and `~/.codex/agents/<name>.toml` exists for each L0 agent.
- **AC-003** — Given the install of AC-001; When the per-user manifest is
  read; Then every installed path carries sha256, owning bundle, and the moai
  version that installed that file.
- **AC-007** — Given a temp HOME holding a user-created file at an install
  target path the manifest does not track; When the installer runs; Then the
  file is byte-identical afterward and the run reports the collision naming
  the path, the reason, and the suggested action.
- **AC-011** — Given a project initialized after M4; When the project tree is
  scanned for skill/agent assets; Then no catalog-derived common-asset
  placement exists anywhere under the project: no `.claude/skills/moai*`
  directory (the plain `moai` pack dir included — measured baseline: 38 skill
  dirs, 37 `moai-*` + 1 plain `moai`), no `.claude/agents/moai/`, no
  `.codex/agents/moai/`, and no `.agents/skills/moai*` directory.
- **AC-013** — Given the M6 retirement; When the repo tree and the build are
  inspected; Then no `plugins/moai/` tree and no marketplace manifest exists,
  and `make build` is green without the plugin-emit targets.
- **AC-020** — Given an existing project holding template-managed,
  user-modified, and user-created `moai-*` skills (and template-managed
  agents under `.claude/agents/moai/`), and NO per-user install on the
  machine; When the REQ-020 migration runs inside `moai update`; Then the
  user-side first install (REQ-024 upgrade arm) completes BEFORE any removal,
  template-managed files are removed, user-modified files are
  preserved with a report, user-created files are untouched, every
  disposition is reported, and after the run the user holds the user-folder
  placement (no neither-state). Three machine-state arms (iter4 D24):
  (a) Given the same project on a machine whose per-user manifest ALREADY
  exists (written by another project's update or a partial install) while
  this project's counterparts are missing; When the migration runs; Then the
  missing counterparts are installed append-only BEFORE their project-side
  removal (no stall, no neither-state); (b) Given an upgrading project
  holding a non-L0 (optional-pack) template-managed skill, no manifest, and
  no `--bundles`; When the migration runs; Then the non-L0 skill remains
  project-side with a report — not installed into an unopted bundle, not
  removed — while L0 counterparts are installed before their removal;
  (c) Given the upgrade install where one user-side write fails (read-only
  directory fixture); When the migration runs; Then the failed file's
  project counterpart is NOT removed (stays project-side + reported),
  remaining files complete, and the summary lists the failure.
- **AC-025** — Given a destination path outside the four roots, a
  sentinel target whose parent directory is a symlink pointing outside the
  roots, a root that is itself a symlink (dotfile-manager `~/.claude`), a
  managed leaf replaced by a symlink pointing outside the root, and a
  REQ-023 backup destination whose resolved path escapes the resolved backup
  home; When the installer validates and writes; Then the outside-root
  path is refused by name with nothing written, the parent-symlink
  escape is refused with the sentinel target unmodified, the symlinked
  root's install lands inside the resolved root while an escape from the
  RESOLVED root is refused, the outside-pointing leaf symlink is never
  followed (no write through it; refused at install, divergence-classified at
  refresh/removal), the escaping backup write is refused like any four-root
  escape (iter4 D33), and every user-folder write is observable as the
  declared posture — temp file inside the validated resolved directory +
  atomic rename with the resolved parent re-validated immediately before it
  (iter4 D26: the post-validation parent swap remains the declared race
  limitation; the mid-flight swap itself is not deterministically testable
  at this layer and is asserted via the posture instead).

## D.2 Traceability

Every REQ-001..REQ-024 maps to ≥1 AC (spec.md §3); every AC maps to at least
one REQ, with the primary REQ marked first in the Verifies column — AC-001,
AC-002, AC-006, AC-008, AC-012, and AC-020 carry secondary REQs after their
primary; AC-025 carries a single REQ (its `(C2 confinement)` note is a
constraint reference, not a REQ).
No orphan AC, no uncovered REQ. AC-009's repoint-clean clause and AC-016's
advisory-row clause verify REQ-014/REQ-019 text directly (iter4 D32 fold —
the design-mandated behavior they assert now rides its requirement).

## D.2b Evidence Ledger — RED-now baseline (iter1 repair, D2)

Carrier for the matrix's RED-now cells (verification-completeness.md §2.1):
each entry carries the four elements — command (single read-only invocation),
verbatim stdout, exit code, and the tree SHA. All entries were executed in one
audit run on the baseline tree below; exit codes were metered in the same run.
Empty stdout with a non-zero exit is a complete observation.

Proxy-cell note (iter2 D20e): the token-absence cells are BASELINE absence
probes establishing the RED state. For cells whose green path lands in a
package the command does not name (EV-001, EV-002, EV-004, EV-007, EV-021 —
the new installer / user-manifest packages per design §2.1-2.2), the flip
evidence is the green milestone's own test named in the green-path line and
the cell is NOT re-run as flip evidence; for cells whose green path genuinely
touches the grepped file (EV-005/006/008 update.go, EV-009/010/016 doctor.go,
EV-017 catalog.yaml, EV-018 update.go — the M3/M5/M0 wiring), the same
command is the flip instrument and its post-green output is named in the
green-path line.

Baseline tree SHA (binds every entry):
`cfb9033582eff27f9031e1a6438d8558aaa48115` (iter2-repair full re-execution,
all 22 cells + 2 positive controls verbatim; identical outputs previously
measured on the iter1 baseline `b965a3912c0e97ef81aeeea773019e633591e1cd`).

### EV-001 — AC-001
- Command: `grep -c UserAsset internal/cli/update.go`
- Stdout: `0`
- Exit code: 1
- Why red (right reason): update carries no user-asset phase — no code writes
  L0 skills into the user root; the installer (M2) and the init trigger
  (REQ-024) do not exist yet.
- Green path: M2 — the temp-HOME installer test asserts L0 dirs + byte match
  and passes, flipping the cell. Proxy note (iter2 D20e): the AC's flip
  evidence is that installer test itself — the installer code lands in the
  new user-asset package and the init trigger in `init.go`, not in
  `update.go`, so this grep is a baseline absence probe of the wrong file for
  the flip and is not expected to change (the M3 update-phase wiring is
  measured by EV-005/006/008's cells instead).

### EV-002 — AC-002
- Command: `grep -c UserHomeDir internal/template/deployer.go`
- Stdout: `0`
- Exit code: 1
- Why red: the deploy path never resolves the user home (research V1) — the
  four-root installer writing `$HOME/.agents/skills` and `~/.codex/agents`
  does not exist.
- Green path: M2 — the Codex-roots installer test on a temp HOME flips it.
  Proxy note (iter2 D20e): the flip evidence is that installer test — the
  user-home resolution lands in the new user-asset package, not in
  `deployer.go` (whose project deploy walk is intentionally untouched), so
  this grep is a baseline probe, not the flip instrument.

### EV-003 — AC-003
- Command: `grep -rn user-assets internal/cli internal/template internal/manifest`
- Stdout: (none)
- Exit code: 1
- Why red: no per-user manifest surface exists anywhere in the tree (the
  token `user-assets` is absent from all three candidate packages); D-Q2's
  location/name gate is itself unresolved.
- Green path: M1 (manifest subsystem) + M2 (manifest writes) — the manifest
  test asserts sha256 + bundle + per-file moai version per path.

### EV-004 — AC-004
- Command: `grep -c "user-asset" internal/cli/update.go`
- Stdout: `0`
- Exit code: 1
- Why red: there is no repeatable user-asset run to replay — the phase the
  idempotency property belongs to does not exist. (update.go's one
  "idempotent" mention, line 394, is a comment on runCleanReinstall — a
  different surface.)
- Green path: M2 — the double-install test proves zero writes + zero-delta
  report. Proxy note (iter2 D20e): the flip evidence is that installer test —
  idempotency lives in the new user-asset package at M2 (update.go's
  user-asset phase wiring arrives with M3), so this grep is a baseline probe.

### EV-005 — AC-005
- Command: `grep -c installed_at internal/cli/update.go`
- Stdout: `0`
- Exit code: 1
- Why red: update records nothing per user-folder file today — no manifest
  entry exists to refresh, so the hash-refresh behavior is absent.
- Green path: M3 — the update refresh test on a temp HOME (mutate shipped
  bytes → update → file rewritten, manifest hash + version updated).
- Flip expectation: `grep -c installed_at internal/cli/update.go` ≥ 1, exit
  0 (iter4 D30c).

### EV-006 — AC-006
- Command: `grep -c schema_version internal/cli/update.go`
- Stdout: `0`
- Exit code: 1
- Why red: update has no manifest-driven removal and no schema gate — neither
  the removal path (REQ-009/023) nor the refusal (REQ-021) exists; the
  extended arms (no-backup dropped divergence, missing-at-removal manifest
  drop, iter2 D16) are equally absent.
- Green path: M3 — the removal/refusal test set on seeded fixtures flips it,
  now covering the dropped-divergence, manifest-stale-at-removal, and
  missing-at-removal arms.
- Flip expectation: `grep -c schema_version internal/cli/update.go` ≥ 1,
  exit 0 (iter4 D30c).

### EV-007 — AC-007
- Command: `grep -c ProtectedSkips internal/cli/update.go`
- Stdout: `0`
- Exit code: 1
- Why red: the skip-and-report semantic exists only in the project deployer
  (`deployer_mode.go` rehome path, research V17); update's user-asset phase
  has no collision surface.
- Green path: M2 — the untracked-collision installer test flips it. Proxy
  note (iter2 D20e): the flip evidence is that installer test — the collision
  surface lives in the new user-asset package, not in `update.go`, so this
  grep is a baseline probe (the M3 summary-count wiring is EV-022's cell).

### EV-008 — AC-008
- Command: `grep -ci divergence internal/cli/update.go`
- Stdout: `0`
- Exit code: 1
- Why red: the tracked-file divergence preserve semantics are new in this
  repair (REQ-023); no code branches on a manifest-vs-disk hash mismatch, and
  the extended truth-table arms (backup under `~/.moai/`, manifest-stale
  repair, missing-file reinstall — iter2 D16) have no implementation.
- Green path: M3 — the divergence test (tracked file edited by hand → update
  preserves + backs up + reports) and the truth-table tests (manifest-stale
  repair without rewrite; missing-file reinstall) flip it.
- Flip expectation: `grep -ci divergence internal/cli/update.go` ≥ 1, exit
  0 (iter4 D30c).

### EV-009 — AC-009
- Command: `grep -ci "user install" internal/cli/doctor.go`
- Stdout: `0`
- Exit code: 1
- Why red: doctor carries no user-install integrity check (research V13: all
  checks are project-scope or retired-carrier-scope).
- Green path: M4+M5 — M4's repointed-diagnostics clean-install regression
  test (the repoint-clean arm, REQ-014) and M5's doctor check tests
  (missing/modified/untracked) flip it (iter4 D30a: matrix row is M4+M5).
- Flip expectation: `grep -ci "user install" internal/cli/doctor.go` ≥ 1,
  exit 0 (iter4 D30c).

### EV-010 — AC-010
- Command: `grep -c "manifest.json" internal/cli/doctor.go`
- Stdout: `0`
- Exit code: 1
- Why red: doctor never reads the project lock file — the three "manifest"
  mentions in doctor.go are skills-allowlist comments about the embedded
  template manifest, not a project-vs-lock comparison (REQ-015 is new).
- Green path: M5 — the both-directions drift test flips it.
- Flip expectation: `grep -c "manifest.json" internal/cli/doctor.go` ≥ 1,
  exit 0 (iter4 D30c).

### EV-011 — AC-011
- Command: `/bin/ls internal/template/templates/.claude/skills/moai-workflow-spec`
- Stdout:
  `modules`
  `references`
  `SKILL.md`
- Exit code: 0
- Tree pin (criterion-level, wins over the document pin):
  `51976e65165e7de0cfa3e0e3bc766ba7389bc46f` — re-executed with the pinned
  form in the iter4 delta round. Carried-nit resolution (iter4): the cell's
  earlier unpinned form (`ls …`) reproduced its long-format stdout only
  through this environment's shell-profile alias — observed three times
  across the iter2/iter3 audits — so the command is pinned to `/bin/ls` and
  the cell re-executed; the pinned form is the portable observation (plain
  name listing, no long format).
- Why red: the embedded tree still carries the skills init deploys into
  projects today — `/bin/ls internal/template/templates/.claude/skills/`
  names 38 directories (37 `moai-*` plus the plain `moai` pack dir the
  former `moai-*` glob missed, iter2 D15; re-measured 38 on tree
  `51976e651` in the iter4 round), and the agent placements
  `.claude/agents/moai/` and `.codex/agents/moai/` are likewise present
  (measured this tree); the AC's post-M4 emptiness does not hold. Positive
  control inherent — the deployed-to-be-removed asset is observed present.
- Green path: M4 — the post-init payload assertion over the AC-011
  placement set (incl. the plain `moai` dirs) flips it.

### EV-012 — AC-012
- Command: `grep -c stripMoaiFromMcpJSON internal/template/deployer_mode.go`
- Stdout: `2`
- Exit code: 0
- Why red: on the default (plugin) deploy path the moai MCP entry is stripped
  from `.mcp.json` today (research V2), so AC-012's `.mcp.json`-with-moai-entry
  requirement is unmet.
- Green path: M4 — the project-payload assertion (moai entry always present)
  flips it.

### EV-013 — AC-013
- Command: `test ! -d plugins/moai`
- Stdout: (none)
- Exit code: 1
- Why red: the plugin payload tree is present — the retirement (M6) has not
  run. Positive control observed in the same audit batch: `ls plugins/moai`
  lists `.claude-plugin`, `.codex-plugin`, `.mcp.json`, `commands`, `skills`;
  `.claude-plugin/marketplace.json` exists (586 bytes).
- Green path: M6 — the same `test ! -d` command exits 0 post-retirement and
  `make build` is green without the plugin-emit targets; the extended
  release-chain arm (iter2 D22) flips via the workflows/scripts/Makefile
  boundary grep over the deleted check scripts running zero-hit (exit 1).

### EV-014 — AC-014
- Command: `grep -c installPluginFor internal/cli/plugin_install.go`
- Stdout: `3`
- Exit code: 0
- Why red: init's plugin install step exists (research V7: `init.go:431`,
  `:445-456`); the boundary the AC demands is crossed today.
- Green path: M6 — the file is deleted (its absence verified by a `test ! -f`
  form, since a grep AT the deleted path exits 2 on a missing file, not
  zero-hit — iter2 D20d), and AC-014's init/update boundary grep over the
  surviving init/update paths runs zero-hit (exit 1).

### EV-015 — AC-015
- Command: `grep -c DeployModePlugin internal/template/deployer_mode.go`
- Stdout: `3`
- Exit code: 0
- Why red: the deploy-mode split is live (`deployer_mode.go:43-62` family) —
  the positive control and the red are the same observation.
- Green path: M7 — the same grep exits 1 (zero hits) after the surface is
  retired.

### EV-016 — AC-016
- Command: `grep -c checkPluginDeployment internal/cli/doctor.go`
- Stdout: `3`
- Exit code: 0
- Why red: the doctor row `{"Plugin Deployment", checkPluginDeployment}` is
  registered today (research V13); the carrier rows the AC forbids exist.
- Green path: M5+M6 — the golden doctor output test shows no carrier rows and
  the repointed/replacement rows per REQ-019 (the REQ-019 advisory row
  included, iter4 D32).
- Flip expectation: `grep -c checkPluginDeployment internal/cli/doctor.go` →
  0, exit 1 (the registration is removed; doctor.go itself survives — iter4
  D30c).

### EV-017 — AC-017
- Command: `grep -ci l0 internal/template/catalog.yaml`
- Stdout: `0`
- Exit code: 1
- Why red: the catalog has no L0 view (sections are `core`,
  `optional_packs`, `harness_generated` under `catalog:`) — the L0 set the
  AC compares against is undefined until M0 resolves D-Q1/D-Q4.
- Green path: M0+M2 — the catalog drift guard pins the resolved L0 list and
  the installer test asserts the installed set equals it.
- Flip expectation: `grep -ci l0 internal/template/catalog.yaml` ≥ 1, exit 0
  (iter4 D30c).

### EV-018 — AC-018
- Command: `grep -ci bundle internal/cli/update.go`
- Stdout: `0`
- Exit code: 1
- Why red: no bundle-unit install/remove exists in the update flow; bundles
  (REQ-004) are introduced by M0's catalog view and consumed by M2/M3.
- Green path: M0+M2+M3 — the M0 catalog view, M2's `--bundles` installer
  test, and M3's `moai bundle` add/remove + update-honors-selection tests
  flip it (iter4 D30b: matrix row is M0+M2+M3, matching design §2.3's M3
  assignment of the command and the update honoring).
- Flip expectation: `grep -ci bundle internal/cli/update.go` ≥ 1, exit 0
  (iter4 D30c).

### EV-020 — AC-020
- Command: `grep -n "only reinstalls" internal/cli/doctor.go`
- Stdout: `981:		// 'moai update' (which only reinstalls the manifest's own skills) would be`
- Exit code: 0
- Why red: the tree's own comment documents the contrary behavior — update
  REINSTALLS project skills today and never offers their removal; the
  provenance-classified migration (M4) and the REQ-024 upgrade arm (the same
  run's completed user install BEFORE the removal, iter2 D14) do not exist.
  (update.go's single
  `user_modified` mention, line 590, is a historical-incident comment, not
  migration code.)
- Green path: M4 — the three-provenance-class migration report test plus the
  upgrade-order test (install-before-removal, no neither-state) flip it.

### EV-021 — AC-021
- Command: `grep -c schema_version internal/manifest/types.go`
- Stdout: `0`
- Exit code: 1
- Why red: neither the project manifest schema nor any user-manifest code
  carries a schema-version gate; the refusal (REQ-021), the corrupt-JSON
  recovery path, and the unknown-field preservation arm (iter2 D23) have no
  implementation.
- Green path: M1+M3 — the foreign-schema, corrupt-manifest, and
  unknown-field-preservation tests flip it. Proxy note (iter2 D20e): the
  grep target is the PROJECT manifest types file; the user-manifest
  subsystem (M1) lives in its own package per design §2.2, so the flip
  evidence is the M1/M3 test set itself, not this grep's count.

### EV-022 — AC-022
- Command: `grep -c collision internal/cli/update.go`
- Stdout: `0`
- Exit code: 1
- Why red: update reports no collision (or divergence) counts — the four+1
  count categories of REQ-011 arrive with M3.
- Green path: M3 — the seeded-fixture summary test asserts each count.

### EV-023 — AC-025
- Command: `grep -c EvalSymlinks internal/template/deployer.go`
- Stdout: `0`
- Exit code: 1
- Why red: the only path confinement in the tree is `validateDeployPath`
  (deployer.go:451-477), which is lexical-only (Clean + `..` rejection +
  string-prefix containment — research V1, iter1 D8); no symlink resolution
  exists anywhere in the deploy path, and no user installer exists to enforce
  the four-root boundary — the extended arms (symlinked-root containment,
  leaf-symlink policy, temp+rename write posture; iter2 D17) are equally
  absent.
- Green path: M2 — the confinement test set (outside-root refusal +
  parent-symlink + symlinked-root + leaf-symlink sentinels) flips it.

## D.3 Indirect Verification

- AC-013/AC-015 verify ABSENCE: their RED-now cells ARE the positive controls
  (EV-013, EV-015 — the retired artifacts observed present at baseline). The
  post-M6/M7 re-run must show the same commands flip (exit 0 / exit 1
  respectively); a zero-hit grep without the observed baseline control would
  prove nothing.
- AC-024's primary evidence is the no-dial trace from the binary's own
  offline operation; a network-less CI lane run is supplementary evidence,
  never a replacement for the trace.

## D.4 Edge Cases (covered inside the ACs above)

- Colliding DIRECTORY (not file) at a skill path → AC-007 semantics apply.
- Manifest present but corrupt JSON → AC-021's second clause: refuse removal,
  report corruption, offer rebuild-from-scan (doctor informational), never
  auto-delete.
- Tracked file hand-edited after install (hash diverges from manifest AND
  shipped bytes) → AC-008 at refresh, AC-006's second clause at removal
  (REQ-023 preserve + report).
- Partial prior install (manifest newer/older than tree) → AC-009 reports the
  divergence class.
- Simultaneous project lock drift and user-install drift → doctor reports
  both independently (AC-009 + AC-010 compose).

## D.5 Quality Gate Criteria

- TRUST 5: 85%+ coverage on new packages; `golangci-lint` clean; gofmt clean;
  conventional commits referencing card t1509.
- Full `make build` (with the post-M6 prerequisite list) green at M6/M7 exit.
- No test writes outside temp dirs (verified by the suite's existing
  HOME-seam tests staying green).

## D.6 Definition of Done

- All Blocker + Major ACs PASS with observed evidence; Minor ACs PASS or carry
  an explicit accepted-debt note in progress.md §E.2.
- decision-index D-Q1..D-Q6 all carry resolutions at close — operator verdicts
  on the four open gates (D-Q1/D-Q2/D-Q4/D-Q5, resolved at kickoff) and the
  plan-phase constraint closures recorded for D-Q3/D-Q6; none open at close.
- REQ-016..019 boundary greps recorded with their baseline positive controls
  (EV-013..EV-016) re-observed flipped in §E.2.
- CHANGELOG entry; retiring SPEC notes SPEC-PLUGIN-MARKETPLACE-001's carrier
  as retired (relationship recorded, history preserved).

## D.7 Forward-Looking Checks

- The manifest schema carries room for future per-file metadata (no breaking
  change when a field is added — consumers ignore unknown fields).
- The bundle taxonomy tolerates a future pack→bundle rename without a
  manifest migration (bundle is a string label, not an enum).
- D-Q3's declared limitation (P6) is doc-visible so a later SPEC can lift it
  without re-deriving the analysis.
