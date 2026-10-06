# Research — SPEC-INIT-SHRINK-001

## 1. Stamps

- Tree: branch `WT-moai-init-slim`, HEAD `3f3ebb763` (= the local `develop` tip this card's
  worktree was created from). Every observation below was taken in this worktree session at this
  tree, 2026-10-03.
- Tool versions on this machine (context only, not a claim about the target runtimes):
  claude 2.1.287 and codex 0.160.0 are the versions the t1434 verdict is stamped with; this card
  ran no real-runtime command yet (the REQ-008 measurement is run-phase work).
- Inputs read: the t1434 verdict
  (`/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1434/.moai/reports/t1434/verdict.md`, read-only
  reference), SPEC-PLUGIN-MARKETPLACE-001 (`/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/…
  SPEC-PLUGIN-MARKETPLACE-001/`, read-only reference), and this tree's sources.

## 2. Discipline of this revision

- Every premise in `spec.md` §1.3 was observed with a command or a file read in this session; the
  commands and deciding outputs are recorded here as R-nn and cited from the P-table.
- No real profile, real home, network, or plugin command was touched. All commands were read-only
  (grep/find/ls/cat/wc) or Go test executions in this tree; the two `go test` runs compiled and ran
  test binaries in scratch (`t.TempDir`) isolation as the repo's tests do.
- The worktree guard refused one compound command (a `sed` with a runtime-computed argument inside
  a `$(...)` pipeline); it was re-run as separate plain commands — recorded here because a refusal
  is a fact about the session, not noise (`verification-claim-integrity.md` §3.1).

## 3. Observations

- **R-01** (P-01) Tree pin.
  - `git rev-parse --short HEAD` → `3f3ebb763`
  - `git branch --show-current` → `WT-moai-init-slim`
- **R-02** (P-02) Template skill set and catalog tiers.
  - `ls internal/template/templates/.claude/skills/ | grep -c .` → `41`
  - `grep -c 'tier: core' internal/template/catalog.yaml` → `36`
  - `grep -c 'tier: optional-pack' internal/template/catalog.yaml` → `13`
  - `grep -c 'tier: harness-generated' internal/template/catalog.yaml` → `1`
  - Note: t1435's P-16 recorded 35 core entries at its older base; this tree carries one more
    (the t1399 rename wave landed here). This card cites its own tree.
- **R-03** (P-03) Command template set: `ls internal/template/templates/.claude/commands/moai/` →
  17 entries — `clean.md.tmpl, codemaps.md.tmpl, e2e.md.tmpl, feedback.md.tmpl, fix.md.tmpl,
  gate.md.tmpl, goal.md.tmpl, gtd.md, harness.md.tmpl, loop.md.tmpl, mx.md.tmpl, plan.md.tmpl,
  project.md.tmpl, review.md.tmpl, run.md.tmpl, sync.md.tmpl, todo.md` (15 `.tmpl` + 2 plain).
- **R-04** (P-04) Codex command-skill mirror template:
  `find internal/template/templates/.agents -type f | wc -l` → `17` (`moai-clean` … `moai-todo`,
  one `SKILL.md` per `/moai` command).
- **R-05** (P-05) Deployer selection: read of `internal/cli/init.go:738-767` — the switch over
  `agentWiringSelection` with `shouldDistributeAll(cmd)` choosing the slim constructor and
  `emitSlimModeNotice` otherwise; quoted comment: "the harness contract outranks the distribute-all
  mode".
- **R-06** (P-06) Update step table: read of `internal/cli/update_template_sync.go:330-432` —
  steps `Backup`, `Validate Templates`, `cleanManagedPathsStage` (which runs `archiveLegacySkills`
  → `presentLegacySkillIDs` snapshot → `InventoryManagedPaths` → `guardFirstDestructiveStep` →
  `deploy.CleanMoaiManagedPaths(projectRoot, out, tmplFS)`), `Deploy Templates`, `Restore Settings`.
- **R-07** (P-07) Managed clean targets: read of `internal/cli/update/deploy/deploy.go:40-87` —
  `ManagedCleanTargets` entries include `.claude/settings.json`, `.claude/commands/moai`,
  `.claude/agents/moai`, `.claude/skills/moai*` (IsGlob), `.claude/rules/moai`,
  `.claude/output-styles/moai`, `.claude/hooks/moai`.
- **R-08** (P-08) Backup exemption: read of `internal/cli/update/deploy/deploy.go:95-106` —
  "BEFORE each root is removed, every regular file tmplFS does not carry at the same relative path
  is copied into the run's pre-clean backup … Template-managed files are NOT backed up: deployment
  rewrites them moments later … (measured 2026-08-15: 12 files vanished this way, among them
  .moai/config/astgrep-rules and dev-only rules under .claude/rules/moai)". The `tmplFS` the caller
  passes is `template.EmbeddedTemplates()` (R-06), the **raw** tree.
- **R-09** (P-09) Archive contract: read of `internal/cli/update_archive.go` — `archiveSkill`
  (source-absent → nil; archive-match → nil; drift → `ARCHIVE_DRIFT`), `--force` drift backup to
  `.moai/archive/skills/v2.16-drift-<UTC-ISO8601>/<id>/`, symlink entries skipped
  (`backup.IsSymlinkEntry`, REQ-SEC-003), per-entry failures accumulated, `total:` summary always
  emitted.
- **R-10** (P-10) List-freshness guard: `internal/cli/update_archive.go:41-44` —
  "TestLegacySkillIDsNotEmbedded (update_archive_guard_test.go) asserts this list stays disjoint
  from the embedded template skill set".
- **R-11** (P-11) Skill mirror: read of `internal/template/skill_mirror.go:1-33` —
  ".agents/skills/<name> as a relative symlink to ../../.claude/skills/<name>, falling back to a
  real directory copy where symlink creation is unavailable"; "Lifecycle management of the mirror
  (removing a mirror whose skill was renamed or retired) is deliberately NOT handled here — it
  belongs to the clean path".
- **R-12** (P-12) Codex-only re-homing: read of `internal/cli/init.go:739-746` — "a codex-only
  selection reroutes the WHOLE deployment through the harnessFS wrapper — claude-only surfaces
  hidden, the skill catalog re-homed to .agents/skills as real directories, skill mirror off";
  `internal/template/harness_fs.go:393` confirms the codex path wraps in harnessFS with the mirror
  disabled.
- **R-13** (P-13) Manifest: `internal/manifest/types.go:92` — "Manifest represents the file
  tracking manifest stored at .moai/manifest.json"; `internal/manifest/manifest.go:14-18` —
  `Load` reads `{projectRoot}/.moai/manifest.json`. `internal/guardstate/manifest.go:28` records
  why the manifest does not live under paths the wholesale clean wipes.
- **R-14** (P-14) Re-init redirect: `internal/cli/init.go:838-843` — "did you mean 'moai update'
  (refresh templates in place)? Re-run with --force only to reinitialize from scratch".
- **R-15** (P-15) MCP decline table: `internal/cli/init.go:997-1004` —
  `mcpDeclined := !opts.MCPProvision`; `case agentWiringGPT: mcpDeclined = true`;
  `case agentWiringBoth: mcpDeclined = false`; comment: "the harness selection is the more specific
  declaration about the MCP surface, and it wins WHEREVER IT CAME FROM".
- **R-16** (P-16) `.mcp.json` template: `cat internal/template/templates/.mcp.json` — `mcpServers`
  with `moai` (`"command": "moai", "args": ["mcp-server"]`) and `context7`
  (`npx -y @upstash/context7-mcp@latest`), plus `staggeredStartup` (`enabled: true, delayMs: 500,
  connectionTimeout: 15000`).
- **R-17** (P-17) Bare-name reference surface:
  - `grep -rn 'Skill("' internal/template/templates/CLAUDE.md internal/template/templates/AGENTS.md.tmpl | wc -l` → `5`
  - `grep -rln 'Skill("moai' internal/template/templates/.claude/rules/ | wc -l` → `7`
- **R-18** (P-18) Install step absent at this tree:
  - `grep -c 'plugin install' internal/cli/init.go` → `0` (exit 1, no output line — the tool does
    not echo grep's status; recorded here).
  - t1435 SPEC frontmatter (`/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/…/spec.md`):
    `status: in-progress`.
- **R-19** (P-19) User-owned namespace: read of `internal/cli/update/plan/plan.go:200-235` —
  user direct-added skills are any name except `moai` / `moai-` prefix under `.claude/skills/`;
  `.claude/agents/harness/` is user-owned (REQ-UNP-002).
- **R-20** (P-20) Snapshot-as-BASE: read of `internal/cli/update_template_sync.go:462-480` —
  `StageDeployedSettingsSnapshot`, `StageDeployedMCPSnapshot`, `writeTemplateSnapshotBestEffort`
  inside Deploy Templates before Restore Settings; card t1139 rationale quoted in source.
- **R-21** (P-21) Managed-path rule: read of `internal/cli/update/plan/plan.go:236-285` —
  `IsMoaiManaged` returns true for `.moai/config/**`, `.moai/evolution/**`, and
  `skills/rules/commands/output-styles/hooks` + `moai-` prefix; agents `core/expert/meta` or
  `moai-` prefix.
- **R-22** (RED-now cells) The Evidence Ledger entries L-01..L-23 of `acceptance.md` were executed
  in this session at this tree; the two decisive shapes:
  - `go test ./internal/template -run '^TestDeployerFileSetExcludesSkillsAndCommands$' -count=1 -v` →
    `testing: warning: no tests to run` / `PASS` / `ok github.com/modu-ai/moai-adk/internal/template 0.458s [no tests to run]`
  - `go test ./internal/cli -run '^TestLegacySkillIDsNotEmbedded$' -count=1 -v` → `PASS` /
    `ok github.com/modu-ai/moai-adk/internal/cli 1.618s` (the positive control).
- **R-23** (guard refusal, recorded) One compound command was refused by the worktree guard:
  a `sed -n "$(grep … )"` pipeline ("this command runs sed with a value computed at runtime …
  inside a construct too complex to verify"). Re-run as separate plain reads; no substitution was
  made for the refused form's output beyond what the plain reads produced.

## 4. Sources

- This tree: `internal/cli/init.go`, `internal/cli/update_template_sync.go`,
  `internal/cli/update_archive.go`, `internal/cli/update/deploy/deploy.go`,
  `internal/cli/update/plan/plan.go`, `internal/template/skill_mirror.go`,
  `internal/template/harness_fs.go`, `internal/template/templates/**`,
  `internal/template/catalog.yaml`, `internal/manifest/types.go`.
- Read-only references (other worktrees, never written):
  - `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1434/.moai/reports/t1434/verdict.md`
  - `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/.moai/specs/SPEC-PLUGIN-MARKETPLACE-001/`
    (`spec.md`, `decision-index.md`, `acceptance.md` conventions)

## 5. What this research did not observe

- Any real Claude or Codex runtime behavior (no `claude -p`, no `codex` command) — the REQ-008
  measurement is run-phase work under the t1434 routes.
- Any network call, marketplace command, plugin install, or write outside this tree.
- The post-shrink behavior itself — every requirement-level claim is a design statement backed by
  the mechanism observations above, not a measurement of the new behavior.
- Windows behavior of the migration and update paths (no Windows host; the guard refuses `pwsh`),
  matching the t1435 gap treatment.
