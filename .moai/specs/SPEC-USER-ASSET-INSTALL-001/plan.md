---
id: SPEC-USER-ASSET-INSTALL-001
title: "plan.md — implementation plan"
version: "0.1.0"
created: 2026-10-05
updated: 2026-10-05
author: manager-spec
---

# Plan — SPEC-USER-ASSET-INSTALL-001

## A. Context

Card t1509 (Class C, Tier L) implements operator decisions D3/D4: common
skills and agents move from project/plugin placement to per-user folders; the
plugin carrier is retired. The card worktree is
`WT-user-asset-copy` @ `6643c7bba`; development mode is TDD
(`.moai/config/sections/quality.yaml`: `development_mode: tdd`).

Research baseline: `research.md` (19 verification rows, V1-V19, all measured
in this tree). Design: `design.md` (7 components, sequencing logic, risk
table).

## B. Known Issues (observed during research, not repaired here)

1. `internal/template/CLAUDE.md` §24.4 names `.claude/agents/{core,expert,meta}/`
   as the template-managed namespace; the actual tree is `agents/moai/`
   (research V12). The migration classifier (M4) must read the code paths, not
   that paragraph. Correcting the doc is a sync-phase byproduct, not a
   milestone.
2. Two catalog-adjacent name lists exist in code today
   (`publishedSkillNames` in `published_skills.go:29-36`, pinned to the
   committed tree by drift guards). The bundle taxonomy (M0) must extend the
   catalog as the single SSOT and add a drift guard for the L0 list rather
   than introducing a third list.
3. `ensureGlobalSettingsEnv` (`update.go:927+`) already manages user-level
   `~/.claude/settings.json`; the user-asset phase (M3) joins this
   user-surface ownership and must not regress the existing cleanup behavior
   (it is pinned by tests).

## C. Pre-flight (run-phase entry checks)

1. Decision gates D-Q1, D-Q2, D-Q4, D-Q5 resolved by the operator
   (decision-index.md) — D-Q1 and D-Q2 BLOCK M0/M1; D-Q4/D-Q5 feed M0.
   D-Q3 and D-Q6 are closed at plan phase by constraint (decision-index:
   POLICY-COVERED; premises P5/P6 in spec.md §1) — M2 ships without profile
   provisioning (P6) and M6 hard-deletes the carrier (P5).
2. `git rev-parse --short HEAD` == `6643c7bba` (or the merged successor of
   this SPEC's landing branch).
3. Baseline measurement: `go test ./internal/template/... ./internal/cli/...`
   green before the first change (this worktree's develop tip is the
   baseline).
4. The plugin golden tests pass at baseline (`make plugin-emit-check`) —
   proving the retirement milestones (M6/M7) start from a coherent carrier.

## D. Constraints

- TDD mode: each milestone lands RED → GREEN → REFACTOR evidence.
- C1-C7 of spec.md bind every milestone (offline determinism, four-root
  confinement, lock-file schema freeze, actionable reports, atomic retirement,
  non-destructive migration, TRUST 5).
- Milestone ordering is decision-reversibility: data model first (M0-M1),
  behavior next (M2-M5), demolition last (M6-M7), verification close (M8).
- No milestone edits deployed user folders on the DEVELOPER's real `$HOME`
  during tests: every user-folder test runs against a temp HOME (the
  env-isolated verification form; `userHomeDirFn`-style seam already exists at
  `update.go`).

## E. Self-Verification

Each milestone's exit evidence:
- RED: the failing test name + output line that names the requirement.
- GREEN: the passing run (package-scoped: `go test ./internal/...` for the
  packages the milestone touched, with `-race` for anything concurrent).
- Boundary greps where a REQ claims absence (e.g. REQ-017: no
  `plugin marketplace` invocation survives in init/update paths).
- Re-measure the OWNING packages, not a sample: M6/M7 re-run the full build
  chain (`make build`) because the Makefile prerequisite list changes.

## F. Milestones

### M0 — Bundle taxonomy and L0 resolution (BLOCKING gate: D-Q1, D-Q4, D-Q5)
Extend `internal/template/catalog.yaml` + loader with the user-install view:
L0 core (plan/run/sync surface, five core agents per D-Q1, hook payload,
factory) and opt-in bundles. Add a catalog drift guard pinning the L0 list to
the resolved gate answer. Reclassify current-`core` entries that are not L0
into bundles. Priority: High. Evidence: catalog loader tests + drift guard.

### M1 — Per-user manifest subsystem (BLOCKING gate: D-Q2)
Implement the per-user manifest: schema (schema_version,
files{path → sha256, bundle, installed_at, moai_version}, collisions — the
installing version is PER FILE per REQ-006; there is no top-level moai_version
field, because REQ-013's partial-failure continuation makes mixed-version
states real), atomic read/write, schema-version refusal for removal (REQ-021)
plus the corrupt-JSON recovery path (AC-021 second clause), four-root path
validation (C2). Location per D-Q2. Priority: High. Evidence: unit tests for
load/save/refuse/corrupt-recovery paths.

### M2 — User-folder installer and init trigger
Implement the user-asset installer over the embedded tree: install, collision
skip-and-report (REQ-010), per-file failure isolation (REQ-013), idempotency
(REQ-012), summary counts (REQ-011 groundwork), symlink-resolved four-root
confinement (C2 — resolved-path judgment, not the project-side lexical check;
AC-025). Wire `moai init` as the first-install trigger (REQ-024: install L0 +
opted-in bundles when no per-user manifest exists). Claude roots and Codex
roots, agents included (REQ-022). No profile provisioning ships (D-Q3 closed:
P6 declared limitation). Priority: High. Evidence: table-driven installer
tests on temp HOMEs (collision, failure, idempotency, both harnesses,
confinement refusal incl. the parent-symlink sentinel).

### M3 — `moai update` user-asset phase
Wire the update flow: refresh (REQ-008 — only when the file's current hash
equals its manifest hash), manifest-driven removal (REQ-009 — same
precondition), tracked-file divergence preserve + backup + report (REQ-023),
collision report (REQ-010), summary counts incl. divergence-preserved
(REQ-011), no-manifest advisory (init is the install trigger per REQ-024 —
update performs no first install), ordering before the project phase, no
regression of the existing global-settings cleanup. Priority: High.
Evidence: update-flow tests with temp HOME + project fixture; existing update
tests stay green.

### M4 — Project slimming and migration
Project deploy stops emitting common skills/agents (REQ-005); the payload
keeps settings, AGENTS.md/CLAUDE.md, lock file, hooks, `.mcp.json` (always
with the moai entry again), output-styles, rules, command wrappers. Migration
for existing projects per REQ-020 (provenance-classified removal/preservation
with reports). Repoint the EXISTING project-scope Codex asset diagnostics that
read project skill/agent paths — `inspectSkillMirror` and
`codexStaleSkillFinding`'s mirror/agent-count inputs
(`internal/cli/doctor_codex.go:429` / `:870` family, research V13) — to the
user-install path in the same change, with a regression test asserting a
correct user-install reports clean (after this milestone they misreport a
correct install as drift). Priority: High. Evidence: init/update payload
assertions (project tree carries no `moai-*` skill dirs after migration);
migration report tests for all three provenance classes; the repointed
Codex-diagnostics clean-install regression test.

### M5 — `moai doctor` integration
New doctor checks: user-install integrity (REQ-014) and project-vs-lock
comparison (REQ-015), read-only, report-only, in the doctor house style.
Priority: Medium. Evidence: doctor check tests (missing/modified/untracked;
both drift directions), golden output rows.

### M6 — Plugin carrier disposition (hard delete per P5)
Remove per REQ-016/REQ-017: `internal/template/pluginemit/` (package + 8 test
files; 13 .go files total), committed `plugins/moai/` tree,
`.claude-plugin/marketplace.json`, `.agents/plugins/marketplace.json`,
`internal/cli/plugin_install.go` + its init/update call sites, Makefile
`plugin-emit`/`plugin-emit-check` targets and their `build:` prerequisites,
the roster-guard sweep-skip entry (`rosterguard/check.go:338-341`), plugin
guidance strings in init. Release-chain gates, dispositioned explicitly (iter1
D4): `.github/workflows/release.yml` provenance Check 8 (lines 123-128 —
invokes `scripts/check-plugin-version.sh`, citing SPEC-PLUGIN-MARKETPLACE-001
REQ-024) is removed with the carrier; `scripts/check-plugin-version.sh`
(reads `plugins/moai/.claude-plugin/plugin.json`) and
`scripts/check-plugin-discoverable.sh` (reads `.claude-plugin/marketplace.json`
+ `plugins/moai/.mcp.json`; REQ-025 carve-out there) are deleted — their read
targets are retired artifacts — together with any workflow/Makefile wiring
that references them (sweep references before deletion). SPEC-PLUGIN-
MARKETPLACE-001 REQ-024 is thereby retired with its carrier (recorded in
spec.md §7). Doctor plugin checks repointed/removed (REQ-019). Priority:
High. Evidence: `make build` green without the targets; boundary grep: zero
`pluginemit` / `marketplace` / `plugin install` references in non-test deploy
paths; zero references to the deleted check scripts; roster guard tests
green.

### M7 — deployer_mode retirement
Remove per REQ-018: `DeployModePlugin`, `PluginMirrorPolicy`, the exclusion
walk branch, `stripMoaiFromMcpJSON`, the plugin re-home path,
`RehomeExistingMirrorEntries` (its migration duty is absorbed by M4's
migration). The deployer returns to a single project payload shape. Priority:
Medium. Evidence: `deployer_mode_test.go` deleted with the surface;
`go build ./...` + template package tests green; boundary grep: zero
`DeployModePlugin` / `PluginMirrorPolicy` references.

### M8 — Cross-harness verification and docs sync
End-to-end on both harnesses: init → user folders populated → update refresh
and removal → collision and divergence paths → doctor rows; the D-Q3 declared
limitation (P6) documented in user-facing docs; template/CLAUDE.md
namespace-paragraph correction if touched; CHANGELOG + docs updates.
Priority: Medium. Evidence: the M8 verification matrix in progress.md §E.2.

## G. Anti-Patterns (refuse during run phase)

- Editing `plugins/moai/**` or the marketplace manifests by hand instead of
  deleting them with their generator (they are generator outputs).
- A second name list in code for L0 (the catalog is the SSOT; add a drift
  guard instead).
- Overwriting an untracked user file "because it looks like ours".
- Manifest-driven removal against an unknown schema version.
- Real-`$HOME` writes from tests.
- Deleting the plugin golden tests in a later commit than the generator
  (C5: same change).
- Silently absorbing `user_modified` project files during migration.

## H. Cross-References

- research.md V1-V19 (source verification), §3 Gaps, §4 Residual Risk.
- design.md §2 components, §3 sequencing, §4 risks.
- acceptance.md AC-001..AC-025 (REQ coverage map in spec.md §3; RED-now
  evidence ledger §D.2b per verification-completeness.md §2).
- decision-index.md D-Q1..D-Q6 + P1-P6 premises.
- SPEC-PLUGIN-MARKETPLACE-001, SPEC-INIT-SHRINK-001, SPEC-CODEX-COMMAND-SKILLS-001.
