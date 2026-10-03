# Plan — SPEC-INIT-SHRINK-001 (card t1438)

## 1. Overview

This plan shrinks the default `moai init` deploy to a thin project scaffold, routes update and the
Codex mirror by a persisted deploy-mode record, and migrates existing projects (duplicate removal
with backup). The plugin carrier itself is SPEC-PLUGIN-MARKETPLACE-001 (t1435, in flight — run-phase
sequencing after its landing is leader-designated); this card consumes its install step, opt-out
surface, and verdict pins by reference (P-18: none of that exists at base tree `3f3ebb763`).

The change is one deployer file-set change plus three supporting units:

1. **Deploy-mode record** (`internal/config` + `internal/cli`) — the persisted `plugin|local` key
   every consumer reads instead of inferring (REQ-009).
2. **Thin deploy set** (`internal/template` deployer family + `internal/cli/init.go`) — the default
   deploy carries no skills/commands; `--no-plugin` and `--all` carry the full local payload
   (REQ-001..007).
3. **Migration** (`internal/cli/update*` + a new migrate unit) — classify, archive-then-remove
   modified, remove identical, leave foreign, idempotent (REQ-010..015).
4. **Reference-resolution gate** (`scripts/check-bare-name-resolution.sh` + a Go wrapper test) —
   the measurement that must run and be recorded before the flip merges (REQ-008, OD-2).

## 2. Decision-reversibility ordering

Milestones are ordered by how likely a decision inside them is to change (data-model and user-facing
first, mechanical last):

1. **M1 — data model, most likely to change.** The mode-record key location and value domain
   (OD-5), the migration classification classes and their archive layout (OD-3), and the
   resolution-gate harness shape (OD-2's input) are the decisions every later milestone consumes;
   they land first and are cheapest to revisit.
2. **M2 — user-facing behavior (the flip).** The default-path file set, the MCP-entry policy
   (OD-1), mirror policy (OD-6), `--all` semantics (OD-7), and the guidance card are what a user
   sees; verdicts here rewrite M2 and ripple into M3's classification scope, not into M1's types.
3. **M3 — update scope and migration execution.** Machinery driven by M1's types and M2's file
   set; its own decisions (trigger surface OD-4, accounting) are narrower.
4. **M4 — mechanical.** Docs, flag help, guarded test surfaces, e2e assertions — last, lowest
   change likelihood.

The REQ-008 measurement runs inside M1 because its verdict routes M2's scope (OD-2 fallbacks
change what the flip deploys and rewrite); the flip does not merge before the verdict is recorded.

## 3. Milestones (priority-ordered, no time estimates)

### M1 — Mode record, classification types, resolution gate (Priority High)

| Deliverable | Files | Requirement |
|---|---|---|
| Deploy-mode key: `deployment_mode: plugin\|local` in `.moai/config/sections/llm.yaml` (OD-5 settled (a) 2026-10-03), reader/writer beside `ReadHarness`/`ApplyHarness` | `internal/config/` (reader), `internal/template/` (ApplyDeployMode beside ApplyHarness), tests | REQ-009 |
| Migration classification types: `identical` / `modified` / `foreign` over the dropped roots — the class gate is template carriage (a file the template render does not carry is foreign: `moai-custom` and every user-created skill or command, preserved byte-for-byte), the manifest the fast path between the two removable classes with conservative absence/stale routing to `modified` within template-carried files (RK-9); archive layout constants (`.moai/archive/skills/<tag>/…`, standalone-file variant) | `internal/cli/update/migrate_classify.go` (new), `internal/cli/update/migrate_classify_test.go`, `internal/cli/update/defs` constants | REQ-010 |
| Resolution-gate harness: `scripts/check-bare-name-resolution.sh <fixture-dir>` — drives the real tool runtimes under scratch config homes (t1434 `--plugin-dir` route; no `HOME=`; live-enumerated env scrub; before/after protected-set hash) answering the three questions of REQ-008; plus `internal/cli` Go wrapper test `TestResolutionGateHarness` asserting the script's PASS/FAIL/RESULT line shape and its negative control | `scripts/check-bare-name-resolution.sh`, `scripts/test-bare-name-resolution.sh` (self-test + negative control), `internal/cli/update/migrate_classify_test.go` or a dedicated `resolution_gate_test.go` | REQ-008, REQ-020 |
| Guard: classification disjointness — foreign files never classified identical/modified, including a glob-hit foreign name (`moai-custom`) and user-created skills (the P-10 guard precedent) | same test file | REQ-013 |

ACs: AC-008, AC-009, AC-010, AC-013, AC-020 (harness halves).

### M2 — The flip: thin default deploy + full local counterpart (Priority High)

| Deliverable | Files | Requirement |
|---|---|---|
| Deployer file-set split: a deploy-mode option on the Claude/Dual/Codex deployer family that excludes `.claude/skills/**` and `.claude/commands/**` from `ListTemplates`/deploy walk on the plugin path; local path unchanged byte-for-byte | `internal/template/deployer.go`, `internal/template/harness_fs.go` (option plumb-through), `internal/template/deployer_mode_test.go` (new) | REQ-001, REQ-002, REQ-003 |
| `init` wiring: default path selects the thin deployer (respecting slim/full catalog routing, P-05); `--no-plugin` and `--all` select the local deployer; mode record written on every run; success-card and slim-mode notice text updated | `internal/cli/init.go`, `internal/cli/init_workflow_flags.go` (if flag registration lives there), `internal/cli/init_mode_test.go` (new) | REQ-001, REQ-003, REQ-007, REQ-009 |
| MCP-entry policy on the default path (OD-1 settled (c) 2026-10-03: the project `moai` entry is written only when the install-outcome probe (design §2.4) reads `not-demonstrated` or `opted-out` — the init surface of the arm mapping: the file set and the mode record follow the deploy path (REQ-001/REQ-009), the probe governs the entry and the guidance only — a probe-`confirmed` install writes no entry, the plugin is the sole carrier; context7 + staggeredStartup preserved; `--no-plugin` writes as today; decline semantics restated) | `internal/cli/init.go` (`provisionMCPEntryUnlessDeclined` call site, sequenced after the install step so it can read the probe outcome), the probe beside the t1435 step, `internal/cli/init_mcp_policy_test.go` (new) | REQ-005 |
| Codex mirror policy (OD-6 settled (a) + condition 2026-10-03: plugin mode deploys no mirror and creates no `.agents/skills` entries — gated on the Codex actual-execution verification; where that verification cannot be produced the mirror stays with its entries re-homed to real directory copies rendered from the embedded template tree, never dangling symlinks into the undeployed `.claude/skills/**`; local mode deploys as today) | `internal/template/skill_mirror.go` / `deployer.go` option, `internal/template/skill_mirror_mode_test.go` (new) | REQ-006 |
| Guidance block: skipped/failed install on the default path names both recourses, fail-open | `internal/cli/init.go`, covered in `init_mode_test.go` | REQ-004 |

ACs: AC-001..AC-007 (minus AC-008), AC-009 re-verified end to end.

### M3 — Update scope + migration execution (Priority Medium)

| Deliverable | Files | Requirement |
|---|---|---|
| Update deployer selection by mode record (`newTemplateSyncDeployer` reads the OD-5 key); absent record → migration path (OD-4 settled (a, amended) 2026-10-03): install step fail-open under the opt-out, its outcome read through the post-install list-surface probe (design §2.4 — the migration surface of the arm mapping; the probe snapshots each acted tool's installed-plugin list surface before the step and diffs it against the post-execution state); probe `confirmed` (this-install success demonstrated by the diff: ref present post-run AND absent from the pre-snapshot for every acted tool) → classify, then remove/archive per REQ-011/012 (removal only after every archive in the batch succeeded), then write the record `plugin`; probe `not-demonstrated` (every other diff outcome, decided only by the observable diff) — no dedupe, no removal, record `local`, and no path records `plugin`; probe `opted-out` — full local payload through the normal local path, record `local` | `internal/cli/update_template_sync.go`, `internal/cli/update_migrate.go` (new), the probe beside the t1435 step, `internal/cli/update_migrate_test.go` (new) | REQ-015, REQ-016 |
| Archive-then-remove: modified classified files archived through `archiveSkill` semantics **before** removal; the removal executor's scope is the classified list (design §3 step 4) — the migration passes Clean the classified removal list (or removes through its own guarded path with Clean's dropped roots skipped for that run); the global `ManagedCleanTargets` walk never runs over the dropped roots in a migration run; archive-write failure aborts before any removal | `internal/cli/update/deploy/deploy.go` (target-list parameterization or the guarded bypass), `internal/cli/update_migrate.go`, tests | REQ-012 |
| Removal + reporting: identical files removed with counts; foreign untouched; idempotence (second run: zero removals, zero archives) | same files | REQ-011, REQ-013, REQ-014 |
| No resurrection: plugin-mode update never re-creates dropped components; mode switch guidance printed (update never flips the record) | same files | REQ-018, REQ-019 |
| Accounting honesty: `managedRedeployCount` / `restoredSet` / `preCleanFiles` report what this run deployed and removed | `internal/cli/update_template_sync.go`, `internal/cli/update/deploy/deploy.go`, tests | REQ-016 |

ACs: AC-011, AC-012, AC-014..AC-019.

### M4 — Docs, guarded surfaces, e2e (Priority Medium)

| Deliverable | Files | Requirement |
|---|---|---|
| Flag help, success card, slim-mode notice text describing the thin deploy and both paths | `internal/cli/init.go` (help strings), covered in M2 tests | REQ-021 |
| README + docs-site init/migration pages describing thin deploy, `--no-plugin`, migration, and the locale note (RK-8) | `README*.md`, `docs-site/content/**` init pages | REQ-021 |
| Guarded test surfaces updated in the same change set: settings snapshot expectations, template-count tests, update dry-run preview goldens, e2e journey assertions (`e2e/cli/tux3_journeys.sh` deployed-file assertions) | the owning test files of each surface | REQ-016, REQ-021 |
| Static doc-claim test: init docs mention `--no-plugin` and the plugin carrier | a grep-guard test beside the init docs tests | REQ-021 |

ACs: AC-021 closes; AC-016/AC-017 re-verified on the updated surfaces.

## 4. Verification plan (scratch-home commands only)

- Every criterion's command form is the plain single invocation measured in this worktree session:
  anchored `go test … -run '^Name$' -count=1`, `sh scripts/<name>.sh <arg>`, `jq`/`grep`/`find`
  reads. No `HOME=` prefix, no `pwsh`, no piped-redirect bundles.
- Init-flow behavior is asserted by Go tests with injected deployers and scratch project roots
  (`t.TempDir`); no harness case runs the real binary's `init` (the t1435 OD-14 verdict,
  inherited — the `$HOME/.claude/settings.json` write of t1435 P-42 stays out of reach).
- The REQ-008 measurement is the one real-runtime surface: `scripts/check-bare-name-resolution.sh`
  runs `claude -p --plugin-dir <local fixture>` under a scratch `CLAUDE_CONFIG_DIR` (and the Codex
  render scan under a scratch `CODEX_HOME`), with the live-enumerated scrub and the protected-set
  hash of REQ-020; the fixture is a local plugin directory, so no marketplace network call happens.
- Migration and update-scope behavior is asserted against fixture project trees built in `t.TempDir`
  (old-project shape: deployed skills/commands + `.moai/manifest.json` records; the manifest-absent
  case is its own test per RK-9).
- Guarded-surface criteria (goldens, snapshots, e2e) cite their owning tests by name; a Go-test
  criterion passes only when its `--- PASS:` line is present (the `[no tests to run]` rule).

## 5. Pre-flight and land order (for the run-phase delegation)

- Pre-flight: `git rev-parse --short HEAD` + `git branch --show-current` re-read; `go build ./...`;
  the anchor probes of `acceptance.md`'s ledger re-run for the affected packages.
- Land-order dependency: t1435 lands on `develop` first (leader-designated). It is `completed` in
  this tree since the 2026-10-03 absorb of develop 6770c714f (at the authoring pin 3f3ebb763 the
  consumed install step, `--no-plugin` flag, and runner seam did not exist — P-18); run-phase
  pre-flight verifies each consumed surface by its landed name before M1 starts. The consumed
  surfaces include the probe's list commands — `claude plugin list` and `codex plugin list --json`,
  read for both the pre-execution snapshot and the post-execution state — and the REQ-017 runner
  seam they start through; a missing or renamed surface there is a
  blocker report, not a local improvisation.
- t1399 note: the t1399 rename wave has landed in this base; if a sibling card renames a skill or
  command again, the plugin derivation (t1435) turns its own emit-check red, and this card's
  classification compares against the render of whatever tree it runs on — no hand list exists to
  go stale (the P-10 lesson).
- OD-2 is settled (a) with its both-surfaces condition (2026-10-03): the measurement's recorded
  verdict routes M2's rewrite scope per that settlement, shipping only what the test proves
  resolvable in each mode. A measurement that arrives unresolved or non-reproducing is NOT
  improvised around — record it and route back to the leader (blocker report) before M2 merges.

## 6. Constraints for manager-develop

- PRESERVE: `internal/template/templates/**` skill and command sources (they are t1435's derivation
  source — deleting any is a defect); `internal/template/catalog.yaml` as data; the archive
  contract's idempotency semantics (P-09); the user-owned namespace rules (P-19/P-21).
- The Clean-step change must keep the existing safety contract: a failed backup aborts before any
  removal (P-08's REQ-UDS-008 rule generalized), and symlinks are never dereferenced (REQ-013).
- The migration must be idempotent (REQ-014) and its removal scoped to the classified sets
  (REQ-013) — a migration that removes a user skill is the worst defect this card can ship.
- No `--no-verify`, no force-push, conventional commits with the card id; forbidden commands per
  the worktree guard (`HOME=`, `pwsh`, bare `git worktree add`).
- Conventional commit subjects: `feat(SPEC-INIT-SHRINK-001): M{N} …`, card id `t1438` in the body,
  `🗿 MoAI` trailer, `Authored-By-Agent: manager-develop` on run-phase transitions.

## 7. Anti-patterns

- Classifying against the **raw** template render instead of the deploy-mode render — that is
  exactly the P-08 exemption this card must close (RK-1).
- Scoping the migration's removal by handing the Clean walk a scoped template FS — `tmplFS` scopes
  only the backup (P-08); the walk removes every glob match with no classification gate, so the
  removal scope must be the classified list itself (design §3 step 4).
- Writing the mode record by inference at update time (REQ-018) instead of persisting it at init.
- Running the resolution gate against the real profile or with a `HOME` override — the t1434 route
  is scratch-home + `--plugin-dir`, and the guard refuses the override (REQ-020).
- Deleting template sources because the deploy no longer carries them (breaks t1435 derivation and
  REQ-003).
- Treating the plugin as present on an old project's update without the fail-open install step
  (RK-4) — or running a network-touching install step with no opt-out escape.
- Asserting a deployed file set in a golden/snapshot test without updating it in the same change
  set (RK-11) — the guard then fails the wrong tree.
- Classifying a template-absent file into a removal class because its name matches the managed
  glob — the class gate is template carriage; `moai-custom` and every user-created skill or
  command is foreign and preserved byte-for-byte (REQ-010/REQ-013).
- Keying a probe arm on the install step's internal behavior — its commands, exit path, and
  output prose are not observable to the caller (the step is fail-open-silent); arms key only on
  the pre-execution snapshot vs post-execution list-surface diff and the opt-out setting
  (design §2.4).

## 8. Cross-references

- `spec.md` §2 (REQ-001..021), §5 (OD-1..OD-8 + marker table), §6 (RK-1..RK-12).
- `design.md` §2 (deployer mode split), §3 (migration pipeline), §4 (resolution gate), §5
  (per-OD alternatives).
- `research.md` R-01..R-14 (commands and outputs behind P-01..P-21).
- `acceptance.md` AC-001..AC-021 + Evidence Ledger.
- SPEC-PLUGIN-MARKETPLACE-001 (`depends_on`): §5 pins OD-1..OD-14; REQ-010/015/016/017/018/019.
- SPEC-PLUGIN-LOAD-SCOPE-001 (completed): `progress.md §E.2` verdict table (R01-R14).
