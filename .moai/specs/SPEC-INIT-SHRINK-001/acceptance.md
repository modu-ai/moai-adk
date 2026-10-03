# Acceptance Criteria — SPEC-INIT-SHRINK-001

This file is the verification layer. Requirements are the `REQ-XXX` entries in `spec.md` §2 (GEARS);
each criterion here is a binary-testable Given/When/Then with the same number.

## Conventions

- **Tree pin.** Every ledger entry below was measured in this worktree session at tree `3f3ebb763`
  (branch `WT-moai-init-slim`). A pin is never re-quoted at a later tree without re-measuring
  (`verification-completeness.md` §4).
- **Amendment re-observation (2026-10-03).** Incorporating the settled OD-1..OD-8 leader verdicts
  changed the asserted behavior of AC-005, AC-006, AC-008, AC-009, AC-012, AC-015, AC-016, and
  AC-021. Their RED-now cells were re-observed on the current tree `3906f985b` (post-absorb:
  develop 6770c714f, incl. t1435 completed in-tree and the t1399 rename) — the entries below carry
  the re-measured output, exit code, and pin. Cells whose assertions did not change keep the
  original `3f3ebb763` pin. A second same-day amendment — the cross-model-audit repair
  (D-7/D-8/D-9) — changed the asserted behavior of AC-006, AC-011, AC-013, and AC-015; those four
  criteria's RED-now cells were re-observed on the repair-session tree `a1f17b038` (L-07, L-12,
  L-14, L-16 below carry the fresh command, output, exit code, and pin; L-12 and L-14 gained their
  own selector commands in the same pass, moving off the L-01 shape). A third same-day amendment —
  the leader-ruling repair (D-11/D-12/D-13, 2026-10-04) — changed the asserted behavior of AC-005
  and AC-015 (the probe's residual arm is diff-keyed and renamed `not-demonstrated`); those two
  criteria's RED-now cells were re-observed on the repair tree `5c380a251` (L-06 and L-16 carry
  the fresh runs). AC-010 and AC-013's fixture and rationale wording was tightened to the
  template-carriage predicate (D-12) without changing their asserted classification behavior, so
  their pins stand per the unchanged-assertion rule. A fourth same-day repair — the
  leader-disposition repair (D-14..D-17, 2026-10-04) — changed the asserted behavior of AC-004
  (its Given re-keyed onto the probe's `not-demonstrated` arm; the opt-out arm removed — the
  opt-out path is REQ-003's full local deploy and never triggers guidance) and AC-012 (the
  archive unit restated as the classified file, never a whole directory); their RED-now cells
  were re-observed on the repair tree `d7bc4e539` (L-05 and L-13 carry the fresh runs). The D-15
  surfaces (AC-011's OD-3 fold sentence and the RK-9 Edge bullet) were qualified with the
  template-carriage predicate without changing their asserted classification behavior, so their
  pins stand per the unchanged-assertion rule.
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
- **Defaults.** At authoring, a criterion bound to an Open Decision carried `default pending OD-n`
  and an `Alternate` line stating how it changes if the verdict differs (`spec.md` §5 marker
  table). Since the 2026-10-03 settlement all eight verdicts are in, so each such criterion now
  carries its settled branch and an `Alternates not taken (recorded)` line instead.
- **Static checks are labelled static.** A text check proves a token or shape is present, not that
  behavior holds; where a criterion carries one it says so, and a behavior check sits beside it.

| AC | Requirement | Milestone | RED-now entry |
|----|-------------|-----------|---------------|
| AC-001 | REQ-001 thin default deploy set | M2 | L-01 |
| AC-002 | REQ-002 template sources retained | M2 | L-03 |
| AC-003 | REQ-003 `--no-plugin` full local payload | M2 | L-04 |
| AC-004 | REQ-004 not-demonstrated install guidance | M2 | L-05 |
| AC-005 | REQ-005 `.mcp.json` policy (OD-1 settled (c)) | M2 | L-06 |
| AC-006 | REQ-006 Codex mirror policy (OD-6 settled (a) + condition) | M2 | L-07 |
| AC-007 | REQ-007 `--all` semantics (OD-7 settled (a)) | M2 | L-08 |
| AC-008 | REQ-008 resolution gate (OD-2 settled (a) + condition) | M1 | L-09 |
| AC-009 | REQ-009 deploy-mode record (OD-5 settled (a) + condition) | M1 | L-10 |
| AC-010 | REQ-010 migration classification | M1 | L-11 |
| AC-011 | REQ-011 identical removal + count (OD-3 settled (a) + conditions) | M3 | L-12 |
| AC-012 | REQ-012 archive-before-removal, abort-on-backup-failure | M3 | L-13 |
| AC-013 | REQ-013 foreign untouched, symlink refusal | M1 | L-14 |
| AC-014 | REQ-014 idempotence | M3 | L-15 |
| AC-015 | REQ-015 migration trigger (OD-4 settled (a, amended)) | M3 | L-16 |
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

### AC-004 — Not-demonstrated install guidance on the default path (REQ-004)

- **Given** the default path with the install step stubbed so the post-install list-surface probe
  reads `not-demonstrated` (the plugin ref never enters the post-execution surface of its pre/post
  diff), **When** init completes deployment, **Then** exactly one guidance block names both
  recourses (`--no-plugin` re-run; the manual install commands) and the init exit status is 0.
- **Verify:** `go test ./internal/cli -run '^TestShrinkInitGuidanceOnMissingPlugin$' -count=1 -v`
  (GREEN path, M2). RED-now: L-05 (re-observed 2026-10-04 — see the ledger).

### AC-005 — `.mcp.json` policy on the default path (REQ-005, OD-1 settled (c) 2026-10-03)

- **Given** the default path, **When** init provisions the project `.mcp.json`, **Then** the file
  carries `context7` and `staggeredStartup`; the `moai` entry is present exactly when the
  install-outcome probe (design §2.4) does not demonstrate this-install success from its pre/post
  diff (`not-demonstrated` — the conditional fallback write, a plugin-less Claude user keeps the
  project carrier) and absent when the probe reads `confirmed` (the plugin is the sole carrier);
  on the `--no-plugin` path it carries the `moai` entry with `command: moai,
  args: [mcp-server]`; an explicit `--llm gpt` run still declines the project entry.
- **Verify:** `go test ./internal/cli -run '^TestDefaultPathMcpEntryPolicy$' -count=1 -v` —
  expected `--- PASS:` present, exit 0. Three arms: probe `not-demonstrated` → entry present;
  probe `confirmed` → entry absent; `--no-plugin` (`opted-out`) → entry present.
- **Alternates not taken (recorded):** (b) the default-path arm would flip to expect the `moai`
  entry present on every run; (c) was taken.

### AC-006 — Codex mirror policy (REQ-006, OD-6 settled (a) + condition 2026-10-03)

- **Given** the deploy-mode split of M2, **When** a plugin-mode deploy runs, **Then** no
  `.agents/skills` entry is created and none is left behind — and that absence holds only after
  Codex is verified to actually execute plugin-borne skills (the REQ-008 Codex question or an
  equivalent actual-execution check); where that verification cannot be produced, the mirror stays
  deployed on every path with its entries re-homed to real directory copies rendered from the
  embedded template tree (P-11's copy fallback) — never left as symlinks into a `.claude/skills/**`
  the plugin path does not carry, so no kept entry dangles. When a local-mode deploy runs (and on a
  codex-only project), the mirror deploys exactly as today (symlink-or-copy, P-11/P-12).
- **Verify:** `go test ./internal/template -run '^TestCodexMirrorFollowsDeployMode$' -count=1 -v`
  — expected `--- PASS:` present, exit 0 (the verification-gated arm and the re-homed-fallback arm
  included).
- **Alternates not taken (recorded):** (b) always-deploy, (c) retire on every path.

### AC-007 — `--all` semantics (REQ-007, OD-7 settled (a) 2026-10-03)

- **Given** `--all` (OD-7 settled (a): a local full deploy), **When** init runs with `--all`,
  **Then** the deployed set is the `--no-plugin` payload plus the optional-pack catalog entries,
  and the mode record reads `local`.
- **Verify:** `go test ./internal/cli -run '^TestAllFlagDeploysAllTiersLocally$' -count=1 -v`.
- **Alternates not taken (recorded):** (b) the expectation would narrow to the optional-pack
  delta, (c) it would assert the deprecation notice.

### AC-008 — Resolution gate (REQ-008, OD-2 settled (a) + condition 2026-10-03)

- **Given** the M1 harness, **When** `scripts/check-bare-name-resolution.sh <fixture>` runs against
  a local fixture plugin under scratch config homes, **Then** it prints one `PASS`/`FAIL <case>`
  line per question (bare-skill resolution, plugin-command-body `Skill("moai")` resolution, Codex
  generated-skill naming), a final `RESULT pass=<n> fail=<m>`, and its verdict is recorded in
  `progress.md` §E.2 **before** the M2 flip merges; and (OD-2 settled condition) BOTH the scaffold
  instruction files AND the plugin command bodies reference only the names that test proves
  resolvable in each deployment mode — the mode-aware rewrite ships only proven-resolvable names,
  per mode.
- **Verify (a):** `sh scripts/check-bare-name-resolution.sh <fixture-dir>` — expected `RESULT
  pass=3 fail=0` (or the recorded verdict routing the rewrite scope).
- **Verify (b):** `go test ./internal/cli -run '^TestResolutionGateHarness$' -count=1 -v` — the
  harness line-shape and negative-control test.
- **Alternates not taken (recorded):** (b) partial shrink and (c) hold would change which M2
  deliverables merge, not this criterion's harness shape; (a) was taken with its both-surfaces
  condition.

### AC-009 — Deploy-mode record (REQ-009, OD-5 settled (a) + condition 2026-10-03)

- **Given** init run on a scratch project (default and `--no-plugin`/`--all` variants), **When** it
  completes, **Then** the OD-5 key (`deployment_mode` in `.moai/config/sections/llm.yaml`) holds
  `plugin` or `local` matching the run, on re-init too, and the update path reads it through the
  same seam; and after a full update cycle (Clean → redeploy → restore, P-06/P-07) a re-read
  returns the same value — the recorded `deployment_mode` survives update's redeploy/restore
  process.
- **Verify:** `go test ./internal/cli -run '^TestDeployModeRecordRoundTrip$' -count=1 -v` (the
  survival arm asserts the post-update re-read).
- **Alternates not taken (recorded):** (b) the key location moves, (c) the record is an
  inference — every criterion naming the key retargets.

### AC-010 — Migration classification (REQ-010)

- **Given** a fixture old-project tree (deployed skill copy identical to the template render; one
  modified; one foreign user skill — `moai-custom`, a name the managed glob matches that the
  template render does not carry; one template-carried file with an absent manifest record),
  **When** the classifier runs, **Then** the four files classify identical / modified / foreign /
  modified respectively, and the three printed counts match.
- **Verify:** `go test ./internal/cli -run '^TestMigrationClassification$' -count=1 -v`.

### AC-011 — Identical removal + count (REQ-011, OD-3 settled (a) + conditions 2026-10-03)

- **Given** a classified set holding template-identical dropped components, **When** the migration
  runs, **Then** they are removed from the project tree, the removed count is printed, and no
  archive copy of an identical component is written — and the removal ran through the
  classified-set-scoped executor (REQ-011; design §3 step 4): the removal list the executor
  processed is the classified set, and the global managed-roots walk did not run over the dropped
  roots in this run.
- **OD-3 conditions folded here:** a template-carried file whose manifest entry is missing or
  stale classifies modified (archive-then-remove via REQ-010/REQ-012) and never enters this
  removal set; a file the template render does not carry is foreign under REQ-010 whatever its
  manifest state and is never classified for removal or archive; removal runs only after every
  archive in the batch has succeeded (REQ-012).
- **Verify:** `go test ./internal/cli -run '^TestMigrationRemovesIdenticalDroppedComponents$' -count=1 -v`.
- **Alternates not taken (recorded):** (b) the no-archive expectation would flip, (c) the
  criterion would assert report-only output.

### AC-012 — Archive-before-removal (REQ-012)

- **Given** a classified set holding one modified skill directory and one modified command file,
  **When** the migration runs, **Then** both are archived (skill through the `archiveSkill`
  layout; the command file into the standalone-file archive) before removal, the archived copies
  byte-match the pre-run content; and removal runs only after EVERY archive in the batch has
  succeeded (OD-3 settled condition, 2026-10-03) — an injected archive-write failure anywhere in
  the batch aborts with nothing removed. The negative control: on the pre-fix tree, the same
  modified skill is deleted with no archive copy (the P-08 exemption) — the control must be shown
  failing before the fix and passing after.
- **Verify:** `go test ./internal/cli -run '^TestMigrationArchivesModifiedBeforeRemoval$' -count=1 -v`
  (carries the negative-control subtest).

### AC-013 — Foreign untouched, symlink refusal (REQ-013)

- **Given** a fixture tree holding a foreign user skill at a name the managed glob matches and
  the template render does not carry (`moai-custom` under `.claude/skills/`) and a symlinked
  entry, **When** the migration runs, **Then** the foreign file is byte-unchanged and still
  present — REQ-010 classifies it foreign because the template render does not carry it, so the
  migration's removal list never contains it, even though the Clean step's global walk (P-06/P-07)
  would have removed it — and the symlink is neither dereferenced, archived, nor followed
  (mirror-entry handling is REQ-006's own clause: removed as link entries or re-homed per its
  fallback form, never dereferenced).
- **Verify:** `go test ./internal/cli -run '^TestMigrationLeavesForeignFilesUntouched$' -count=1 -v`.

### AC-014 — Idempotence (REQ-014)

- **Given** a project the migration already ran on, **When** update runs the migration again,
  **Then** it removes nothing, archives nothing, and reports zero counts.
- **Verify:** `go test ./internal/cli -run '^TestMigrationIdempotent$' -count=1 -v`.

### AC-015 — Migration trigger (REQ-015, OD-4 settled (a, amended) 2026-10-03)

- **Given** a fixture project with no mode record, **When** update runs, **Then** the install step
  runs fail-open under the opt-out and its outcome is read through the post-install list-surface
  probe (design §2.4 — a pre-execution snapshot of each acted tool's installed-plugin list surface
  diffed against the post-execution state): probe `confirmed` — the plugin ref present post-run
  AND absent from the pre-snapshot for every acted tool — the dedupe follows classification and
  the record is written `plugin`; probe `not-demonstrated` (every other diff outcome — a
  pre-existing plugin, the ref absent after the run, an unreadable or ambiguous surface, or a
  probe error; arms name observable diff states only, never a cause inside the step) — nothing is
  deduped or removed, the record is written `local`, and no path records `plugin`; probe
  `opted-out` — the full local payload deploys, nothing is removed, and the record is written
  `local`.
- **Verify:** `go test ./internal/cli -run '^TestUpdateMigratesLegacyProject$' -count=1 -v`
  (three arms mapped to the probe: `confirmed`; `not-demonstrated`; `opted-out`).
- **Alternates not taken (recorded):** (b) guidance-plus-local-record, (c) `moai migrate`.

### AC-016 — Update mode-scoped deployer + honest accounting (REQ-016)

- **Given** a `plugin`-mode project on the verification-produced path (REQ-006's fallback deploys
  the re-homed mirror per its own clause and is AC-006's subject), **When** update's template sync
  runs, **Then** the selected
  deployer is the thin one, no dropped skill/command component is re-deployed, the outcome summary counts
  what this run actually deployed and removed (a zero-redeploy run reports zero), and the recorded
  `deployment_mode` value is byte-identical after the run (REQ-016's survival clause, OD-5
  condition).
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

- **Given** a migrated `plugin`-mode project on the verification-produced path (under REQ-006's
  no-verification fallback the mirror's fate is REQ-006's own clause — re-homed copies held
  stable, not resurrection), **When** update runs again with `--force`, **Then**
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
  post-shrink tree in the same change set; and this card's report deliverable states the t1466
  handover scope (OD-8 settled (a) — REQ-008 resolvability only; no broader sweep joins this SPEC).
- **Verify (a):** `go test ./internal/cli -run '^TestInitDocsDescribeThinDeploy$' -count=1 -v`
  (static grep guard).
- **Verify (b):** `go test ./internal/cli -run '^TestUpdateDryRunPreviewGoldens$' -count=1 -v` (or
  the owning golden test's current name — M4 pins it).

## Edge Cases

- **Manifest absent or stale on an old project** (RK-9): every template-carried dropped-root file
  under a missing record classifies modified — archived, never silently removed; dropped-root
  files the template render does not carry (a user-created `moai-custom` skill included) are
  foreign under REQ-010 whatever the record — preserved, never archived, never removed.
- **Foreign skill whose name starts with `moai-`**: the migration's class gate is template
  carriage — a user skill named `moai-custom` is foreign because the template render does not
  carry it, regardless of manifest tracking or the P-19/P-21 managed-name match (those rules stay
  the name gate for update's existing protection and decide only identical-vs-modified among
  template-carried files; template absence alone is foreign — the migration run itself preserves
  it byte-for-byte, never removing or archiving it (REQ-013); beyond that run the file is subject
  to update's existing local-path behavior (REQ-017 — today's merge scope, and the Clean walk's
  P-08 backup-then-remove), which this card does not change: the preservation guarantee is the
  migration run plus REQ-017's machinery, not a byte-for-byte promise for every future update).
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

Each entry: command (single invocation), verbatim stdout, exit code, tree SHA. Entries measured at
the authoring pin carry SHA `3f3ebb763`; the re-observed entries of the 2026-10-03 amendment carry
`3906f985b`. Exit codes were read from the tool result (the worktree guard refuses
redirect-and-echo bundles).

- **L-01** (AC-001, AC-016 GREEN-path anchors)
  - Command: `go test ./internal/cli -run '^TestDefaultDeploySetExcludesSkillsAndCommands$' -count=1 -v`
  - Stdout (verbatim, tail):
    ```
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	1.542s [no tests to run]
    ```
  - Exit code: 0. Red reason: the criterion's named test does not exist yet — red by the
    PASS-line rule. Flipped by M2 (AC-001).
  - Amendment note (2026-10-03): re-run at `3906f985b` for the OD-verdict incorporation produced
    the identical shape — `testing: warning: no tests to run` / `PASS` /
    `ok github.com/modu-ai/moai-adk/internal/cli 0.777s [no tests to run]`, exit 0 — and serves as
    the fresh form observation cited by L-09 and L-23 (L-16 moved onto its own selector in the
    cross-model repair). This entry's own pin (AC-001's assertion did not change) stays
    `3f3ebb763`.
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
- **L-05** (AC-004; re-observed 2026-10-04 at tree `d7bc4e539` — the D-14 repair re-keyed the
  criterion's Given onto the probe's `not-demonstrated` arm and removed the opt-out arm, so the
  cell was re-measured with its own selector)
  - Command: `go test ./internal/cli -run '^TestShrinkInitGuidanceOnMissingPlugin$' -count=1 -v`
  - Stdout (verbatim, at tree `d7bc4e539`):
    ```
    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.787s [no tests to run]
    ```
  - Exit code: 0; red by the PASS-line rule (the criterion's named test does not exist yet).
    Flipped by M2.
- **L-06** (AC-005; re-observed 2026-10-03 — the OD-1 (c) amendment added the conditional third
  arm to the criterion's assertion, so the cell was re-measured)
  - Command: `go test ./internal/cli -run '^TestDefaultPathMcpEntryPolicy$' -count=1 -v`
  - Stdout (verbatim, at tree `3906f985b`):
    ```
    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	1.040s [no tests to run]
    ```
  - Exit code: 0; red by the PASS-line rule (the criterion's named test does not exist yet).
    Flipped by M2.
  - D-13-repair re-observation (2026-10-04, tree `5c380a251` — the probe-arm rewrite changed this
    criterion's assertion): same command, verbatim stdout: `testing: warning: no tests to run` /
    `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/cli	0.901s [no tests to run]`, exit 0 —
    same no-tests-to-run shape.
- **L-07** (AC-006; re-observed 2026-10-03 twice — first for the OD-6 verification-gated arm at
  `3906f985b`, then for the cross-model repair's re-homed-fallback binding at `a1f17b038`)
  - Command: `go test ./internal/template -run '^TestCodexMirrorFollowsDeployMode$' -count=1 -v`
  - Stdout (verbatim, at tree `3906f985b`):
    ```
    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/template	0.374s [no tests to run]
    ```
  - Stdout (verbatim, at tree `a1f17b038` — the D-8 repair re-observation):
    ```
    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/template	0.434s [no tests to run]
    ```
  - Exit code: 0 (both runs); red by the PASS-line rule. Flipped by M2.
- **L-08** (AC-007) — the AC-007 test does not exist; the L-01 shape cell stands for the form
  (internal/cli package, anchored `-run`).
- **L-09** (AC-008) — the harness script and its Go wrapper test do not exist; the L-01 shape cell
  stands for the form; the script's own `isolation-*` negative controls are M1 deliverables.
  Re-observed 2026-10-03 at `3906f985b` via the L-01 form (the OD-2 both-surfaces amendment changed
  this criterion's assertion): the L-01 command's fresh run printed `testing: warning: no tests to
  run` / `PASS` / `ok github.com/modu-ai/moai-adk/internal/cli 0.777s [no tests to run]`, exit 0 —
  same no-tests-to-run shape; the harness test still does not exist.
- **L-10** (AC-009; re-observed 2026-10-03 — the OD-5 survival arm changed the criterion's
  assertion, so the cell was re-measured)
  - Command: `go test ./internal/cli -run '^TestDeployModeRecordRoundTrip$' -count=1 -v`
  - Stdout (verbatim, at tree `3906f985b`):
    ```
    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	1.092s [no tests to run]
    ```
  - Exit code: 0; red by the PASS-line rule. Flipped by M1.
- **L-11** (AC-010) — the AC-010 test does not exist; L-01 shape cell stands for the form.
- **L-12** (AC-011; own selector, observed 2026-10-03 at tree `a1f17b038` — the cross-model repair
  (D-7) extended the criterion's assertion with the classified-set executor scope, so the cell
  moved off the L-01 shape onto its own command)
  - Command: `go test ./internal/cli -run '^TestMigrationRemovesIdenticalDroppedComponents$' -count=1 -v`
  - Stdout (verbatim, at tree `a1f17b038`):
    ```
    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.792s [no tests to run]
    ```
  - Exit code: 0; red by the PASS-line rule (the criterion's named test does not exist yet).
    Flipped by M3.
- **L-13** (AC-012; re-observed 2026-10-03 — the OD-3 every-archive-in-the-batch condition
  strengthened the criterion's assertion, so the cell was re-measured)
  - Command: `go test ./internal/cli -run '^TestMigrationArchivesModifiedBeforeRemoval$' -count=1 -v`
  - Stdout (verbatim, at tree `3906f985b`):
    ```
    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.788s [no tests to run]
    ```
  - Exit code: 0; red by the PASS-line rule. Flipped by M3. The negative control (pre-fix silent
    deletion) is a subtest shown failing before the fix lands.
  - D-17-repair re-observation (2026-10-04, tree `d7bc4e539` — the repair restated the archive
    unit as the classified file, never a whole directory, sharpening the criterion's assertion):
    same command, verbatim stdout:
    ```
    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.667s [no tests to run]
    ```
    exit 0 — same no-tests-to-run shape.
- **L-14** (AC-013; own selector, observed 2026-10-03 at tree `a1f17b038` — the cross-model repair
  (D-7) strengthened the criterion's Given to a glob-hit foreign name, so the cell moved off the
  L-01 shape onto its own command)
  - Command: `go test ./internal/cli -run '^TestMigrationLeavesForeignFilesUntouched$' -count=1 -v`
  - Stdout (verbatim, at tree `a1f17b038`):
    ```
    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.618s [no tests to run]
    ```
  - Exit code: 0; red by the PASS-line rule. Flipped by M1 (the migration classification unit).
- **L-15** (AC-014) — the AC-014 test does not exist; L-01 shape cell stands for the form.
- **L-16** (AC-015; own selector, observed 2026-10-03 at tree `a1f17b038` — the cross-model repair
  (D-9) re-mapped the criterion's three arms onto the post-install list-surface probe, so the cell
  moved off the L-01 shape onto its own command; the earlier OD-4 re-observation at `3906f985b`
  used the L-01 form and printed the identical no-tests-to-run shape)
  - Command: `go test ./internal/cli -run '^TestUpdateMigratesLegacyProject$' -count=1 -v`
  - Stdout (verbatim, at tree `a1f17b038`):
    ```
    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.653s [no tests to run]
    ```
  - Exit code: 0; red by the PASS-line rule. Flipped by M3.
  - D-13-repair re-observation (2026-10-04, tree `5c380a251` — the probe-arm rewrite changed this
    criterion's assertion): same command, verbatim stdout: `testing: warning: no tests to run` /
    `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/cli	0.724s [no tests to run]`, exit 0 —
    same no-tests-to-run shape.
- **L-17** (AC-016; re-observed 2026-10-03 — the OD-5 survival clause extended the criterion's
  assertion, so the cell was re-measured)
  - Command: `go test ./internal/cli -run '^TestUpdatePluginModeSkipsDroppedRedeploy$' -count=1 -v`
  - Stdout (verbatim, at tree `3906f985b`):
    ```
    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.818s [no tests to run]
    ```
  - Exit code: 0; red by the PASS-line rule. Flipped by M3.
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
  Re-observed 2026-10-03 at `3906f985b` via the L-01 form (the OD-8 handover-scope clause extended
  this criterion's assertion): the L-01 command's fresh run printed the identical no-tests-to-run
  shape (`...internal/cli 0.777s [no tests to run]`), exit 0.
- **L-24** (AC-002 premise pins, positive evidence at the pin tree)
  - Commands and deciding outputs (each a single invocation):
    - `ls internal/template/templates/.claude/skills/ | grep -c .` → `41`
    - `grep -c 'tier: core' internal/template/catalog.yaml` → `36`
    - `grep -c 'tier: optional-pack' internal/template/catalog.yaml` → `13`
    - `grep -c 'tier: harness-generated' internal/template/catalog.yaml` → `1`
    - `find internal/template/templates/.claude/commands/moai -maxdepth 1 -type f | wc -l` → `17`
      (15 `*.md.tmpl` plus `gtd.md` and `todo.md`)
    - `find internal/template/templates/.agents -type f | wc -l` → `17`
    - `grep -rn 'Skill("' internal/template/templates/CLAUDE.md internal/template/templates/AGENTS.md.tmpl | wc -l` → `5`
    - `grep -rln 'Skill("moai' internal/template/templates/.claude/rules/ | wc -l` → `7`
    - `grep -c 'plugin install' internal/cli/init.go` → `0` (exit 1, no line echoed — recorded, not
      echoed by the tool)
  - These pin the AC-002 expectations and the P-17/P-18 premise figures at tree `3f3ebb763`.
