# Acceptance Criteria — SPEC-INIT-SHRINK-001

This file is the verification layer. Requirements are the `REQ-XXX` entries in `spec.md` §2 (GEARS);
each criterion here is a binary-testable Given/When/Then with the same number.

## Conventions

- **Tree pin.** Every ledger entry below was measured in this worktree session at tree `3f3ebb763`
  (branch `WT-moai-init-slim`). A pin is never re-quoted at a later tree without re-measuring
  (`verification-completeness.md` §4).
- **Two cells per criterion** (`verification-completeness.md` §2): a RED-now cell — the criterion's
  own command observed red on this tree, with the reason it is red, in the Evidence Ledger — and a
  green-path cell naming the milestone that flips it and the passing output. Where the RED-now
  command would be vacuous, a positive control that fires on a known-good input is named in the
  ledger (L-02 is the shared positive control: the guard-test pattern exists and passes on this
  tree).
- **Command form.** Single invocations measured in this worktree session: anchored
  `go test … -run '^Name$' -count=1 -v`, plain `sh scripts/<name>.sh <arg>`, `grep`/`find`/`jq`
  reads. A Go-test criterion passes only when the named `--- PASS: <Name>` line is present —
  `go test` exits 0 with `no tests to run` when a selector matches nothing, so every Go-test
  RED-now cell is red by the PASS-line rule, not by the exit code. `grep -c 0`-shaped outputs are
  recorded with their non-echoed exit status where the tool does not surface it.
- **Isolation.** Scratch `CLAUDE_CONFIG_DIR`, `CODEX_HOME`, `MOAI_HOME`, and scratch project roots
  (`t.TempDir` in Go tests; session-scratchpad directories for scripts), one fresh set per
  criterion. Neither a command nor a script sets `HOME` (the worktree guard refuses it; moving it
  into a script file is not a way round that). No acceptance command writes outside its scratch
  set and the tree under verification. The one real-runtime surface is the REQ-008 measurement
  script (local fixture, `--plugin-dir`, no marketplace network), which carries the scrub and the
  protected-set hash of REQ-020.
- **Defaults.** A criterion bound to an Open Decision carries `default pending OD-n` and an
  `Alternate` line stating how it changes if the verdict differs (`spec.md` §5 marker table).
- **Static checks are labelled static.** A text check proves a token or shape is present, not that
  behavior holds; where a criterion carries one it says so, and a behavior check sits beside it.

| AC | Requirement | Milestone | RED-now entry |
|----|-------------|-----------|---------------|
| AC-001 | REQ-001 thin default deploy set | M2 | L-01 |
| AC-002 | REQ-002 template sources retained | M2 | L-03 |
| AC-003 | REQ-003 `--no-plugin` full local payload | M2 | L-04 |
| AC-004 | REQ-004 skipped/failed install guidance | M2 | L-05 |
| AC-005 | REQ-005 `.mcp.json` policy (default pending OD-1) | M2 | L-06 |
| AC-006 | REQ-006 Codex mirror policy (default pending OD-6) | M2 | L-07 |
| AC-007 | REQ-007 `--all` semantics (default pending OD-7) | M2 | L-08 |
| AC-008 | REQ-008 resolution gate (default pending OD-2) | M1 | L-09 |
| AC-009 | REQ-009 deploy-mode record (default pending OD-5) | M1 | L-10 |
| AC-010 | REQ-010 migration classification | M1 | L-11 |
| AC-011 | REQ-011 identical removal + count (default pending OD-3) | M3 | L-12 |
| AC-012 | REQ-012 archive-before-removal, abort-on-backup-failure | M3 | L-13 |
| AC-013 | REQ-013 foreign untouched, symlink refusal | M1 | L-14 |
| AC-014 | REQ-014 idempotence | M3 | L-15 |
| AC-015 | REQ-015 migration trigger (default pending OD-4) | M3 | L-16 |
| AC-016 | REQ-016 update mode-scoped deployer + honest accounting | M3 | L-17 |
| AC-017 | REQ-017 local-mode update keeps full scope | M3 | L-18 |
| AC-018 | REQ-018 mode switch only via init | M3 | L-19 |
| AC-019 | REQ-019 no resurrection | M3 | L-20 |
| AC-020 | REQ-020 hermetic verification | M1/M3 | L-21, L-22 |
| AC-021 | REQ-021 docs + guarded surfaces | M4 | L-23 |

## Acceptance Criteria

### AC-001 — Thin default deploy set (REQ-001)

- **Given** the run's tree after M2, **When** the default-path deployer (no opt-out, slim mode) is
  constructed and its deploy walk is executed against a scratch project root, **Then** the written
  file set contains no path under `.claude/skills/` or `.claude/commands/`, and contains the
  instruction files, rules, agents, settings template render, and `.moai/config` render the default
  deploys today.
- **Verify (a):** `go test ./internal/cli -run '^TestDefaultDeploySetExcludesSkillsAndCommands$' -count=1 -v`
  — expected `--- PASS: TestDefaultDeploySetExcludesSkillsAndCommands ` present, exit 0.
- **Verify (b):** `go test ./internal/template -run '^TestDeployerModeSplitsFileSet$' -count=1 -v`
  — the same expectation at the deployer level (the option excludes skills/commands from
  `ListTemplates`).
- **Alternate (OD-2 (b)):** the command set keeps the router skill and commands in the scaffold;
  the exclusion narrows to the non-command skills and both tests rename their fixture set.

### AC-002 — Template sources retained (REQ-002)

- **Given** the same tree, **When** the embedded template tree is enumerated and the emit-check
  machinery runs, **Then** every skill and command source present at the base tree is still
  present, `catalog.yaml` parses with the same tier counts (36 core / 13 optional-pack /
  1 harness-generated at the pin), and `commands-emit-check` / `agents-emit-check` pass.
- **Verify (a):** `go test ./internal/template -run '^TestEmbeddedSkillAndCommandSourcesRetained$' -count=1 -v`
  — expected `--- PASS:` present (GREEN path; M2). RED-now: the test does not exist (L-03), and the
  base-tree counts it will assert are pinned by L-24.
- **Verify (b):** `make commands-emit-check agents-emit-check` — expected exit 0 before and after
  the flip (the machinery is untouched; a red here is a derivation regression, not a shrink).

### AC-003 — `--no-plugin` full local payload (REQ-003)

- **Given** a scratch project root, **When** init runs with the opt-out set (the t1435 `--no-plugin`
  flag, or `MOAI_SKIP_PLUGIN_INSTALL=1` in the test harness), **Then** the deployed file set equals
  today's default payload — skills, commands, project `.mcp.json` moai entry, and Codex mirror
  included — and the deployer selected is the local one.
- **Verify:** `go test ./internal/cli -run '^TestNoPluginPathDeploysFullLocalPayload$' -count=1 -v`
  — expected `--- PASS: TestNoPluginPathDeploysFullLocalPayload ` present, exit 0.
- **Alternate (OD-6 (b)/(c)):** the mirror arm of the fixture expectation changes per the mirror
  verdict; the skills/commands/MCP arms do not.

### AC-004 — Skipped/failed install guidance (REQ-004)

- **Given** the default path with the install step stubbed to fail (or the opt-out set so it is
  skipped), **When** init completes deployment, **Then** exactly one guidance block names both
  recourses (`--no-plugin` re-run; the manual install commands) and the init exit status is 0.
- **Verify:** `go test ./internal/cli -run '^TestShrinkInitGuidanceOnMissingPlugin$' -count=1 -v`
  (GREEN path, M2). RED-now: L-05.

### AC-005 — `.mcp.json` policy on the default path (REQ-005, default pending OD-1)

- **Given** the default path, **When** init provisions the project `.mcp.json`, **Then** the file
  carries `context7` and `staggeredStartup` and (default) carries no `moai` entry; on the
  `--no-plugin` path it carries the `moai` entry with `command: moai, args: [mcp-server]`;
  an explicit `--llm gpt` run still declines the project entry.
- **Verify:** `go test ./internal/cli -run '^TestDefaultPathMcpEntryPolicy$' -count=1 -v` —
  expected `--- PASS:` present, exit 0.
- **Alternate (OD-1 (b)):** the default-path arm flips to expect the `moai` entry present;
  (OD-1 (c)) a third arm expects it exactly when the install step failed.

### AC-006 — Codex mirror policy (REQ-006, default pending OD-6)

- **Given** the deploy-mode split of M2, **When** a plugin-mode deploy runs, **Then** no
  `.agents/skills` entry is created and none is left behind; when a local-mode deploy runs (and on
  a codex-only project), the mirror deploys exactly as today (symlink-or-copy, P-11/P-12).
- **Verify:** `go test ./internal/template -run '^TestCodexMirrorFollowsDeployMode$' -count=1 -v`
  — expected `--- PASS:` present, exit 0.
- **Alternate (OD-6 (b)/(c)):** the plugin-mode arm flips to always-deploy, or the local arm is
  void with migration removal on local projects too.

### AC-007 — `--all` semantics (REQ-007, default pending OD-7)

- **Given** `--all` (default pending OD-7: a local full deploy), **When** init runs with `--all`,
  **Then** the deployed set is the `--no-plugin` payload plus the optional-pack catalog entries,
  and the mode record reads `local`.
- **Verify:** `go test ./internal/cli -run '^TestAllFlagDeploysAllTiersLocally$' -count=1 -v`.
- **Alternate (OD-7 (b)/(c)):** the expectation narrows to the optional-pack delta, or asserts the
  deprecation notice.

### AC-008 — Resolution gate (REQ-008, default pending OD-2)

- **Given** the M1 harness, **When** `scripts/check-bare-name-resolution.sh <fixture>` runs against
  a local fixture plugin under scratch config homes, **Then** it prints one `PASS`/`FAIL <case>`
  line per question (bare-skill resolution, plugin-command-body `Skill("moai")` resolution, Codex
  generated-skill naming), a final `RESULT pass=<n> fail=<m>`, and its verdict is recorded in
  `progress.md` §E.2 **before** the M2 flip merges; the scaffold instruction files reference only
  names resolvable in each deployment mode.
- **Verify (a):** `sh scripts/check-bare-name-resolution.sh <fixture-dir>` — expected `RESULT
  pass=3 fail=0` (or the recorded verdict routing OD-2).
- **Verify (b):** `go test ./internal/cli -run '^TestResolutionGateHarness$' -count=1 -v` — the
  harness line-shape and negative-control test.
- **Alternate (OD-2):** the recorded verdict routes to (a) mode-aware rewrite (the instruction-file
  sweep joins M2), (b) partial shrink, or (c) hold — each changes which M2 deliverables merge, not
  this criterion's harness shape.

### AC-009 — Deploy-mode record (REQ-009, default pending OD-5)

- **Given** init run on a scratch project (default and `--no-plugin`/`--all` variants), **When** it
  completes, **Then** the OD-5 key holds `plugin` or `local` matching the run, on re-init too, and
  the update path reads it through the same seam.
- **Verify:** `go test ./internal/cli -run '^TestDeployModeRecordRoundTrip$' -count=1 -v`.
- **Alternate (OD-5 (b)/(c)):** the key location moves, or the record is an inference — every
  criterion naming the key retargets.

### AC-010 — Migration classification (REQ-010)

- **Given** a fixture old-project tree (deployed skill copy identical to the template render; one
  modified; one foreign user skill; one with an absent manifest record), **When** the classifier
  runs, **Then** the four files classify identical / modified / foreign / modified respectively,
  and the three printed counts match.
- **Verify:** `go test ./internal/cli -run '^TestMigrationClassification$' -count=1 -v`.

### AC-011 — Identical removal + count (REQ-011, default pending OD-3)

- **Given** a classified set holding template-identical dropped components, **When** the migration
  runs, **Then** they are removed from the project tree, the removed count is printed, and no
  archive copy of an identical component is written.
- **Verify:** `go test ./internal/cli -run '^TestMigrationRemovesIdenticalDroppedComponents$' -count=1 -v`.
- **Alternate (OD-3 (b)/(c)):** the no-archive expectation flips, or the criterion asserts
  report-only output.

### AC-012 — Archive-before-removal (REQ-012)

- **Given** a classified set holding one modified skill directory and one modified command file,
  **When** the migration runs, **Then** both are archived (skill through the `archiveSkill`
  layout; the command file into the standalone-file archive) before removal, the archived copies
  byte-match the pre-run content, and an injected archive-write failure aborts with nothing
  removed. The negative control: on the pre-fix tree, the same modified skill is deleted with no
  archive copy (the P-08 exemption) — the control must be shown failing before the fix and passing
  after.
- **Verify:** `go test ./internal/cli -run '^TestMigrationArchivesModifiedBeforeRemoval$' -count=1 -v`
  (carries the negative-control subtest).

### AC-013 — Foreign untouched, symlink refusal (REQ-013)

- **Given** a fixture tree holding a foreign user skill and a symlinked entry, **When** the
  migration runs, **Then** the foreign file is byte-unchanged and still present, and the symlink
  is neither dereferenced, archived, nor followed.
- **Verify:** `go test ./internal/cli -run '^TestMigrationLeavesForeignFilesUntouched$' -count=1 -v`.

### AC-014 — Idempotence (REQ-014)

- **Given** a project the migration already ran on, **When** update runs the migration again,
  **Then** it removes nothing, archives nothing, and reports zero counts.
- **Verify:** `go test ./internal/cli -run '^TestMigrationIdempotent$' -count=1 -v`.

### AC-015 — Migration trigger (REQ-015, default pending OD-4)

- **Given** a fixture project with no mode record, **When** update runs, **Then** (default) the
  install step runs fail-open under the opt-out before classification, the dedupe follows it, and
  the record is written `plugin`; with the opt-out set, the full local payload deploys, nothing is
  removed, and the record is written `local`.
- **Verify:** `go test ./internal/cli -run '^TestUpdateMigratesLegacyProject$' -count=1 -v`.
- **Alternate (OD-4 (b)/(c)):** the plugin arm becomes guidance-plus-local-record, or names
  `moai migrate`.

### AC-016 — Update mode-scoped deployer + honest accounting (REQ-016)

- **Given** a `plugin`-mode project, **When** update's template sync runs, **Then** the selected
  deployer is the thin one, no dropped component is re-deployed, and the outcome summary counts
  what this run actually deployed and removed (a zero-redeploy run reports zero).
- **Verify:** `go test ./internal/cli -run '^TestUpdatePluginModeSkipsDroppedRedeploy$' -count=1 -v`.

### AC-017 — Local-mode update keeps full scope (REQ-017)

- **Given** a `local`-mode project, **When** update runs, **Then** the deploy, 3-way merge, and
  snapshot semantics match today's behavior (settings/MCP/template snapshots staged before the
  restore step, P-20).
- **Verify:** `go test ./internal/cli -run '^TestUpdateLocalModeKeepsFullScope$' -count=1 -v`.

### AC-018 — Mode switch only via init (REQ-018)

- **Given** any project, **When** update runs, **Then** the mode record is unchanged by update and
  the switch guidance names the init re-entry.
- **Verify:** `go test ./internal/cli -run '^TestUpdateNeverFlipsModeRecord$' -count=1 -v`.

### AC-019 — No resurrection (REQ-019)

- **Given** a migrated `plugin`-mode project, **When** update runs again with `--force`, **Then**
  `.claude/skills/**`, `.claude/commands/**`, and mirror entries stay absent.
- **Verify:** `go test ./internal/cli -run '^TestUpdateForceDoesNotResurrectDropped$' -count=1 -v`.

### AC-020 — Hermetic verification (REQ-020)

- **Given** the whole acceptance suite, **When** it runs, **Then** no case reaches a real profile
  or real home: the default external-command runner refuses to exec under a test binary, no
  command or script sets `HOME`, the measurement script's scrub is live-enumerated, and its run is
  bracketed by an equal before/after protected-set hash (directory entries included); a negative
  control that disables the scrub fails the harness.
- **Verify (a):** `go test ./internal/cli -run '^TestShrinkVerificationNeverReachesRealHome$' -count=1 -v`
  — the injected-runner and no-HOME guards.
- **Verify (b):** the measurement script's `isolation-*` cases (`scripts/test-bare-name-resolution.sh`)
  print `PASS` per case and `RESULT pass=<n> fail=0`.

### AC-021 — Docs + guarded surfaces (REQ-021)

- **Given** the M4 change set, **When** the guarded surfaces run, **Then** the init flag help and
  success card describe the thin deploy and both paths (static: a grep-guard test), the README and
  docs-site init pages name `--no-plugin` and the plugin carrier, and every file-set-asserting
  surface (settings snapshot, template-count, dry-run preview, e2e journey) passes against the
  post-shrink tree in the same change set.
- **Verify (a):** `go test ./internal/cli -run '^TestInitDocsDescribeThinDeploy$' -count=1 -v`
  (static grep guard).
- **Verify (b):** `go test ./internal/cli -run '^TestUpdateDryRunPreviewGoldens$' -count=1 -v` (or
  the owning golden test's current name — M4 pins it).

## Edge Cases

- **Manifest absent or stale on an old project** (RK-9): every dropped-root file under a missing
  record classifies modified — archived, never silently removed.
- **Foreign skill whose name starts with `moai-`**: classification uses the manifest + template
  render, not the name prefix alone — a user skill named `moai-custom` that the manifest does not
  track and the template does not carry is foreign (P-19's rule stays the name gate for update's
  existing protection; the migration's class gate is the render compare).
- **A project already migrated, then `--no-plugin` re-init**: the init re-entry redeploys the full
  local payload and flips the record to `local` — the documented switch surface (REQ-018), not a
  migration defect.
- **Version-compare skip** (RK-7): the migration is keyed to the post-shrink binary's version bump;
  a re-run on an up-to-date project is the idempotence case (AC-014), not a second migration.
- **Codex-only project migrating**: the re-homed `.agents/skills` real directories (P-12) are in
  the dropped-root classification scope; under OD-6 (a) a plugin-mode codex project loses nothing
  the plugin does not carry, and the local-mode codex project keeps the re-homing byte-for-byte.
- **Concurrent cards renaming components** (t1399 class): classification compares against the
  render of the tree it runs on; no hand list can go stale (P-10's lesson generalized).

## Quality Gates

- TRUST 5: every new test surface carries table-driven cases; the migration unit keeps ≥85%
  coverage on its package (`go test -cover ./internal/cli/...`); `gofmt`/`golangci-lint` clean per
  the repo toolchain; no new hard-coded env names (the `envkeys.go` rule).
- RED-first: each M1-M3 deliverable's criterion is observed red (this ledger) before its
  implementation flips it; the negative control of AC-012 is shown failing pre-fix.
- The plan-auditor PASS threshold for Tier L (0.85) governs this plan phase.

## Definition of Done

- All 21 criteria carry a green verification (command + observed output) in `progress.md` §E.2/§E.3.
- Every Open Decision row in `decision-index.md` carries an operator verdict line or the line
  `default adopted`, and the marker table's alternates were applied where a verdict differs.
- The REQ-008 verdict is recorded in §E.2 before the M2 flip merges (AC-008 ordering clause).
- No criterion rests on a measurement taken at another tree or another run (baseline-integrity,
  `verification-claim-integrity.md` §2).

## Evidence Ledger (RED-now cells)

Each entry: command (single invocation), verbatim stdout, exit code, tree SHA `3f3ebb763`.
Exit codes were read from the tool result (the worktree guard refuses redirect-and-echo bundles).

- **L-01** (AC-001, AC-016 GREEN-path anchors)
  - Command: `go test ./internal/cli -run '^TestDefaultDeploySetExcludesSkillsAndCommands$' -count=1 -v`
  - Stdout (verbatim, tail):
    ```
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	1.542s [no tests to run]
    ```
  - Exit code: 0. Red reason: the criterion's named test does not exist yet — red by the
    PASS-line rule. Flipped by M2 (AC-001).
- **L-02** (shared positive control)
  - Command: `go test ./internal/cli -run '^TestLegacySkillIDsNotEmbedded$' -count=1 -v`
  - Stdout (verbatim, tail):
    ```
        --- SKIP: TestLegacySkillIDsNotEmbedded/manifest_error (0.00s)
        --- SKIP: TestLegacySkillIDsNotEmbedded/manifest_empty (0.00s)
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	1.618s
    ```
  - Exit code: 0. Proves the guard-test pattern and its anchor form work on this tree (the
    existing P-10 guard runs and passes).
- **L-03** (AC-002)
  - Command: `go test ./internal/template -run '^TestEmbeddedSkillAndCommandSourcesRetained$' -count=1 -v`
  - Stdout (verbatim, tail): `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/template	0.458s [no tests to run]`
    (same shape as L-01; measured for the template package at this tree).
  - Exit code: 0; red by the PASS-line rule. Flipped by M2.
- **L-04** (AC-003)
  - Command: `go test ./internal/cli -run '^TestNoPluginPathDeploysFullLocalPayload$' -count=1 -v`
  - Stdout: `PASS` / `ok  	...internal/cli	1.312s [no tests to run]`. Exit 0; red by the
    PASS-line rule. Flipped by M2.
- **L-05** (AC-004) — the AC-004 test does not exist; same no-tests-to-run shape as L-04 (the
  AC-004 anchor will be captured at M2 authoring; the shape cell L-04 stands for the form).
- **L-06** (AC-005)
  - Command: `go test ./internal/cli -run '^TestDefaultPathMcpEntryPolicy$' -count=1 -v`
  - Stdout: `PASS` / `ok  	...internal/cli	1.338s [no tests to run]`. Exit 0; red by the
    PASS-line rule. Flipped by M2.
- **L-07** (AC-006)
  - Command: `go test ./internal/template -run '^TestCodexMirrorFollowsDeployMode$' -count=1 -v`
  - Stdout: `PASS` / `ok  	...internal/template	0.291s [no tests to run]`. Exit 0; red by the
    PASS-line rule. Flipped by M2.
- **L-08** (AC-007) — the AC-007 test does not exist; the L-01 shape cell stands for the form
  (internal/cli package, anchored `-run`).
- **L-09** (AC-008) — the harness script and its Go wrapper test do not exist; the L-01 shape cell
  stands for the form; the script's own `isolation-*` negative controls are M1 deliverables.
- **L-10** (AC-009)
  - Command: `go test ./internal/cli -run '^TestDeployModeRecordRoundTrip$' -count=1 -v`
  - Stdout: `PASS` / `ok  	...internal/cli	1.354s [no tests to run]`. Exit 0; red by the
    PASS-line rule. Flipped by M1.
- **L-11** (AC-010) — the AC-010 test does not exist; L-01 shape cell stands for the form.
- **L-12** (AC-011) — the AC-011 test does not exist; L-01 shape cell stands for the form.
- **L-13** (AC-012)
  - Command: `go test ./internal/cli -run '^TestMigrationArchivesModifiedBeforeRemoval$' -count=1 -v`
  - Stdout: `PASS` / `ok  	...internal/cli	1.362s [no tests to run]`. Exit 0; red by the
    PASS-line rule. Flipped by M3. The negative control (pre-fix silent deletion) is a subtest
    shown failing before the fix lands.
- **L-14** (AC-013) — the AC-013 test does not exist; L-01 shape cell stands for the form.
- **L-15** (AC-014) — the AC-014 test does not exist; L-01 shape cell stands for the form.
- **L-16** (AC-015) — the AC-015 test does not exist; L-01 shape cell stands for the form.
- **L-17** (AC-016)
  - Command: `go test ./internal/cli -run '^TestUpdatePluginModeSkipsDroppedRedeploy$' -count=1 -v`
  - Stdout: `PASS` / `ok  	...internal/cli	1.542s [no tests to run]`. Exit 0; red by the
    PASS-line rule. Flipped by M3.
- **L-18** (AC-017) — the AC-017 test does not exist; L-01 shape cell stands for the form.
- **L-19** (AC-018) — the AC-018 test does not exist; L-01 shape cell stands for the form.
- **L-20** (AC-019) — the AC-019 test does not exist; L-01 shape cell stands for the form.
- **L-21** (AC-020)
  - Command: `go test ./internal/cli -run '^TestShrinkVerificationNeverReachesRealHome$' -count=1 -v`
  - Stdout: `PASS` / `ok  	...internal/cli	1.555s [no tests to run]`. Exit 0; red by the
    PASS-line rule. Flipped by M1/M3.
- **L-22** (AC-020) — the measurement script's isolation cases do not exist yet (script is an M1
  deliverable); L-09's disposition covers the shape.
- **L-23** (AC-021) — the AC-021 guard tests do not exist; L-01 shape cell stands for the form.
- **L-24** (AC-002 premise pins, positive evidence at the pin tree)
  - Commands and deciding outputs (each a single invocation):
    - `ls internal/template/templates/.claude/skills/ | grep -c .` → `41`
    - `grep -c 'tier: core' internal/template/catalog.yaml` → `36`
    - `grep -c 'tier: optional-pack' internal/template/catalog.yaml` → `13`
    - `grep -c 'tier: harness-generated' internal/template/catalog.yaml` → `1`
    - `ls internal/template/templates/.claude/commands/moai/ | wc -l` (equivalent listing measured:
      19 entries — 17 `*.md.tmpl`, `gtd.md`, `todo.md`)
    - `find internal/template/templates/.agents -type f | wc -l` → `17`
    - `grep -rn 'Skill("' internal/template/templates/CLAUDE.md internal/template/templates/AGENTS.md.tmpl | wc -l` → `5`
    - `grep -rln 'Skill("moai' internal/template/templates/.claude/rules/ | wc -l` → `7`
    - `grep -c 'plugin install' internal/cli/init.go` → `0` (exit 1, no line echoed — recorded, not
      echoed by the tool)
  - These pin the AC-002 expectations and the P-17/P-18 premise figures at tree `3f3ebb763`.
