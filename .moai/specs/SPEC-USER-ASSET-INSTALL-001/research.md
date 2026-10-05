---
id: SPEC-USER-ASSET-INSTALL-001
title: "research.md — codebase research and source-verification table"
version: "0.5.0"
created: 2026-10-05
updated: 2026-10-05
author: manager-spec
---

# Research — Deploy Surface, Profiles, Update/Doctor Contracts, Dispositions

Tree baseline: this card worktree `WT-user-asset-copy` @ `6643c7bba` (develop tip).
All `file:line` citations below were read in THIS run against THIS tree
(`git rev-parse --short HEAD` → `6643c7bba`, verified at session start).

Baseline re-pin (iter4, 2026-10-05): the card tree absorbed develop
`a158b4b5f` (merge `51976e651`), so the working baseline moves
`6643c7bba` → `51976e651`. The load-bearing pins were re-derived on the
merged tree and ALL HOLD: `templates/.claude/skills/` = 38 dirs (37
`moai-*` + plain `moai`), `templates/.agents/skills/` = 17 (no plain), the
single `moai` dirs under `.claude/agents/` and `.codex/agents/`, the
`codex_readiness.go` anchors (`:131` consumer, `:215-217` definition), and
`codexStaleSkillFinding` (`:857-870`). ONE anchor drift observed: the
`inspectSkillMirror` comment block now starts at `:425` (was `:427-428`);
the `func inspectSkillMirror` line is `:429` UNCHANGED (re-measured on
`51976e651` this run), and every live citation in these artifacts names the
function line `:429` — so no citation changes. The table rows below keep
their original `6643c7bba` attributions (they were true where and when
measured); run-phase milestones re-measure against the landing tip per plan
§C.2.

Reference inputs (read-only scratchpad, treated as hypotheses, never cited as
ground truth): the t1509 research note set at
`/private/tmp/claude-501/.../scratchpad/agentsmd/` — `agentsmd-research.md`,
`agentsmd-mapping.md`, `agentsmd-open-questions.md`, `parity-matrix.md`,
`AGENTS.new.md`, `CLAUDE.new.md`. Every concrete claim below was re-verified in
this tree; where the note and the tree disagree, the tree wins and the row says
so.

## 1. Source-Verification Table

Obligation numbers follow the dispatch (source-verification obligations 1-6).

| # | Claim under test (note/card) | Verdict | Measured evidence (this tree @ 6643c7bba) |
|---|---|---|---|
| V1 | Deploy walk lives in `internal/template` and writes under the project root | CONFIRMED | `internal/template/deployer.go:146` (`Deploy`) and `:157` (`DeployWithResult`) walk the embedded FS; `:451` `validateDeployPath` confines destinations to the project root. Deploy targets today are project-relative only — there is no user-folder write path in the deployer. |
| V2 | Deploy mode split exists: `local` (full payload) vs `plugin` (thin payload, skills/commands carried by the plugin) | CONFIRMED | `internal/template/deployer_mode.go:39-46` (`DeployModeLocal`, `DeployModePlugin`); `:52-62` (`PluginMirrorPolicy` none/rehome); `:81-89` `isPluginExcludedPath` excludes `.claude/skills/` + `.claude/commands/` (and `.agents/skills/` under `MirrorPolicyNone`) from the plugin payload; `:103-131` `stripMoaiFromMcpJSON` removes the `moai` MCP entry from the project render on the plugin path. |
| V3 | Skill mirror: canonical `.claude/skills`, Codex mirror `.agents/skills` (project-relative) | CONFIRMED | `internal/template/skill_mirror.go:51` (`CanonicalSkillsRelDir = ".claude/skills"`), `:65` (`MirrorSkillsRelDir = filepath.Join(".agents", "skills")`), `:79-90` (`DeployResult.SkillMirrors`, `ProtectedSkips`). All PROJECT-relative today. |
| V4 | Codex command skills are committed template files, not generated at deploy time | CONFIRMED | `internal/template/published_skills.go:29-36` pins the exact set of 17 published command skills (`moai-clean` … `moai-todo`); the committed tree `internal/template/templates/.agents/skills/` carries exactly those 17 directories (ls, this run). |
| V5 | Skill/agent catalog with tiers already exists | CONFIRMED | `internal/template/catalog.yaml` — sections `core` (25 skills + 11 agents), `optional_packs` (backend, deployment, design, devops, frontend, testing — 13 skills, 0 agents), `harness_generated` (builder-harness). Typed loader `internal/template/catalog_loader.go:21-34` (`TierCore`, `TierHarnessGenerated`, `TierOptionalPackPrefix`); entries carry name/tier/path/hash/version. Catalog hashes are sha256 hex (`:52-53` field docs). |
| V6 | Plugin generator + committed plugin tree exist; retire targets confirmed | CONFIRMED | Generator `internal/template/pluginemit/pluginemit.go:46-49` (`ClaudeMarketplacePath = ".claude-plugin/marketplace.json"`, `CodexMarketplacePath = ".agents/plugins/marketplace.json"`, `ClaudePluginPath`/`CodexPluginPath` under `plugins/moai`), `:86` `Emit`. Committed outputs present: `.claude-plugin/marketplace.json`, `.agents/plugins/marketplace.json`, `plugins/moai/{.claude-plugin, .codex-plugin, .mcp.json, commands/ (17), skills/ (25)}` (ls, this run). Build wiring: `Makefile:34` (`build:` depends on `plugin-emit-check`), `Makefile:62-72` (`plugin-emit`, `plugin-emit-check`). Golden/drift tests: `internal/template/pluginemit/*_test.go` (8 test files; 13 .go files total in the package). Roster-guard exemption note: `internal/harness/rosterguard/check.go:338-341`. |
| V7 | Init carries the plugin install step (both harnesses) | CONFIRMED | `internal/cli/init.go:431` plugin-path selection (`--no-plugin` flag, `MOAI_SKIP_PLUGIN_INSTALL` env, `--all`), `:445-456` the "install could not be demonstrated" block naming `claude plugin marketplace add … ; claude plugin install …` and the codex equivalent; `internal/cli/plugin_install.go:8` documents the same command shape; `:184` `runPluginInstallStep`, `:200` `installPluginFor`. |
| V8 | Project lock file = `.moai/manifest.json` with triple-hash provenance | CONFIRMED | `internal/manifest/manifest.go:18` ("Load reads the manifest from {projectRoot}/.moai/manifest.json"); `internal/manifest/types.go:14-36` provenance classes (`template_managed`, `user_modified`, `user_created`, `deprecated`, `generated_managed`); `:92-107` `Manifest{Version, DeployedAt, Files}` + `FileEntry{Provenance, TemplateHash, DeployedHash, CurrentHash, Parts}`. The card's "프로젝트 잠금 파일" maps to this existing artifact — no new project-side lock file is required. |
| V9 | User-folder writes today: settings + hooks cleanup ONLY; no user-level skills/agents | CONFIRMED | `internal/cli/update.go:919-920` `globalMoaiHooksDir(homeDir)` = `~/.claude/hooks/moai`; `:927+` `ensureGlobalSettingsEnv` REMOVES that dir and manages `~/.claude/settings.json` env keys (`PATH`, `CLAUDE_DISABLE_PATH_WARNING`, `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`). No skills/agents deploy to `$HOME` exists: the `os.UserHomeDir` grep across `internal/cli` returns only statusline, doctor, launcher, tokens, mcp_codex, flag_slot, doctor_mcp_provider — none deploy assets. |
| V10 | Profiles: `~/.moai/claude-profiles/<name>` are isolated `CLAUDE_CONFIG_DIR`s; doctor follows CLAUDE_CONFIG_DIR | CONFIRMED | `internal/paths/paths.go:91` (`ProfilesDir` → `~/.moai/claude-profiles`); `internal/cli/profile.go:20-22` ("Each profile is an isolated Claude configuration directory (CLAUDE_CONFIG_DIR)"); `internal/cli/cc.go:36` (`moai cc -p` sets `CLAUDE_CONFIG_DIR`); `internal/cli/doctor_plugin_version.go:44` ("The Claude home follows CLAUDE_CONFIG_DIR, else …"); `internal/config/envkeys.go:620` (`EnvClaudeConfigDir`). What moai WRITES into a profile dir today: nothing asset-shaped found — profile dirs are created for launcher isolation (memory-store migration code exists: `internal/cli/migrate_profiles.go`), not for skill/agent installs. Visibility of user-level skills inside a profile session is NOT settled by the card — routed to decision-index D-Q3 (closed at plan phase in the iter1 repair: declared limitation, premise P6). |
| V11 | Codex agents deploy as emitted TOML per project today | CONFIRMED | Committed tree `internal/template/templates/.codex/agents/moai/*.toml` (12 files: builder-harness, e2e-tester, manager-design, manager-develop, manager-docs, manager-git, manager-lead, manager-spec, manager-todo, plan-auditor, super-advisor, sync-auditor — ls, this run); emission source `internal/template/agentemit/`. The card's Codex target `~/.codex/agents` receives NO files today. |
| V12 | `moai update` re-syncs templates with a namespace contract | CONFIRMED | `internal/cli/update.go:134` `runUpdate` main flow; update-side global-settings work at `:919-1050`; namespace protection implemented in `internal/cli/update.go` + `internal/cli/update_archive.go` (protected dirs `.claude/` local config, `.moai/project/`, `.moai/specs/` — per `internal/template/CLAUDE.md` "Protected directories during sync"). NOTE: `internal/template/CLAUDE.md` §24.4 text names `.claude/agents/{core,expert,meta}/` as template-managed, but the actual template tree carries `templates/.claude/agents/moai/` — the convention text is stale relative to the tree (observation recorded; the run-phase namespace work must read the code, not that doc paragraph). |
| V13 | `moai doctor` check inventory (what the SPEC extends) | CONFIRMED | `internal/cli/doctor.go:181-358` grouped checks: system / moai / workspace. Relevant rows: `{"Plugin Deployment", checkPluginDeployment}` and `{pluginVersionCheckName, checkPluginVersion}` (SPEC-PLUGIN-MARKETPLACE-001 REQ-020..023, card t1435); `{"Skills Allowlist", checkSkillsAllowlist}`; `{"Hook Delivery", checkHookDelivery}`; `{"Hook Wiring Drift", checkHookWiringDrift}`. Codex-side: `internal/cli/doctor_codex.go:429` `inspectSkillMirror`, `:870` `codexStaleSkillFinding`. `internal/cli/doctor_plugin_version.go:7` reads `<CLAUDE_CONFIG_DIR \| ~/.claude>/plugins/installed_plugins.json`. Codex readiness (added iter2 D19): `internal/cli/codex_readiness.go:121` `probeCodexReadiness` consumes `countCodexAgentTOMLs(root)` at `:131`, defined `:215-217` — the agent-TOML count reads `.codex/agents/moai/*.toml` under the PROJECT root. All are PROJECT-scope or retired-carrier-scope checks; no user-asset comparison exists. |
| V14 | "핵심 에이전트 5종" — which 5 agents are L0 | UNRESOLVED (routed to decision gate) | Catalog core tier carries 11 agents: sync-auditor, manager-develop, manager-docs, manager-git, manager-lead, manager-spec, plan-auditor, super-advisor, manager-todo, manager-design, e2e-tester (`internal/template/catalog.yaml` core.agents, this run); builder-harness is `harness_generated`. The card names "핵심 에이전트 5종" without enumerating them. Candidate reading A (leading): manager-spec, manager-develop, manager-docs, plan-auditor, sync-auditor — the plan→run→sync chain including its two auditors. Candidate readings B/C exist (including manager-git; including manager-lead). Per the dispatch this is NOT decided here — decision-index D-Q1, and REQ-003 binds to the gate's resolution rather than to a name list. |
| V15 | "plan·run·sync" in the L0 definition | AMBIGUOUS (routed) | Two readings: (a) the published command skills `moai-plan`, `moai-run`, `moai-sync` (V4 set); (b) the workflow skills `moai-workflow-spec` / `moai-workflow-tdd` / `moai-workflow-ddd` (+docs). Reading (a) is more literal (the card lists them beside agents/hooks/factory, all install units). decision-index D-Q4; REQ-003 names both readings as gate input. |
| V16 | Note claims (agentsmd-research.md) about the deploy surface used in the draft mapping | PARTIALLY CONFIRMED, two corrections | The note's guard inventory rows (byte ceilings, disclosure tests, link contract) were read in this tree and hold at their cited files EXCEPT: the note reads tree `2f492df19`; this tree is `6643c7bba` (newer). Correction 1: the note's mapping §3 says "None of the 14 [skills] has a `.agents/skills` template entry of its own; they reach Codex through the deploy-time mirror" — in THIS tree `templates/.agents/skills/` holds 17 COMMITTED published command skills and `harness_fs.go:14-19` documents the deploy-time re-home of the catalog alongside them (`templates/.claude/skills/` names 38 directories on this tree: 37 `moai-*` + the plain `moai` pack dir — `/bin/ls` measured, iter2 D21 correcting the stale "34-directory" figure); both mechanisms coexist (published = committed files, catalog = re-home), which the note conflates. Correction 2: the note says `deployer_mode` MirrorPolicyNone "deploys no mirror at all" — confirmed, but the mirror it refers to is PROJECT-relative; the card's Codex target `$HOME/.agents/skills` is a NEW surface no current code writes. |
| V17 | Existing per-file collision rule to inherit (never overwrite user files) | CONFIRMED | `internal/template/deployer_mode.go:160-181` `rehomeOneSkill`: a non-link entry at the mirror path is skipped and reported ("a non-symlink entry already exists … left untouched"); `published_skills.go:56-63` `recordProtectedSkip` appends skipped published-skill paths to `DeployResult.ProtectedSkips`; `internal/manifest/types.go:23-25` `UserCreated` provenance ("Never modify"). The user-folder installer reuses this semantic, not this code (project-scope today). |
| V18 | Embedded assets make an offline user install possible | CONFIRMED | `internal/template/embed.go` `//go:embed all:templates` + `//go:embed catalog.yaml` (per `internal/template/CLAUDE.md`; embed directives read this run); the binary carries every asset it would install — no network fetch is part of any deploy path today. |
| V19 | The L0 factory skills already exist as catalog core entries (premise P2 satisfiable without new assets) | CONFIRMED | `internal/template/catalog.yaml:41` (`- name: moai-factory-foreman`) and `:46` (`- name: moai-lane-watchdog`), both under catalog core with `path: templates/.claude/skills/...` entries (grep, this run; iter1 D11 row). The operator premise P2 (factory in L0) therefore re-classifies existing core entries — no new skill assets are required. |

## 2. Current Deploy Model (as measured)

```
moai init (default, plugin mode)                 moai init --no-plugin / --all (local mode)
├─ project: thin payload                         ├─ project: FULL payload
│   ├─ NO .claude/skills/**, NO .claude/commands/**   ├─ .claude/skills/** (catalog entries per slim/all)
│   ├─ .claude/agents/moai/*.md (12)              │   ├─ .claude/agents/moai/*.md (12)
│   ├─ .codex/agents/moai/*.toml (12)             │   ├─ .codex/agents/moai/*.toml (12)
│   ├─ .agents/skills/** (17 published + rehomed catalog per policy)   ├─ .agents/skills/**
│   ├─ settings.json, hooks/, .mcp.json (moai entry stripped)          ├─ settings.json, hooks/, .mcp.json (moai entry present)
│   ├─ AGENTS.md, CLAUDE.md, rules/               │   ├─ AGENTS.md, CLAUDE.md, rules/
│   └─ .moai/manifest.json (lock)                 │   └─ .moai/manifest.json (lock)
└─ harness: claude/codex plugin install (V7)     └─ no plugin step
```

Plugin artifacts (V6) are committed generator outputs at the repo root — retired
by this SPEC. The card's user-folder model replaces BOTH modes' skill/agent
placement: user folders for common assets, project for the slim harness only.

### 2b. Round-5 verification (2026-10-05, tree `064ff9960` — the codex gate's four findings verified at source before pinning)

| # | Claim under test (round-5 gate) | Verdict | Measured evidence (this tree @ 064ff9960) |
|---|---|---|---|
| W1 | The L0 agents' static preload skills are {moai-foundation-core, moai-workflow-spec, moai-foundation-quality} | CONFIRMED (with one precision) | Agent-body frontmatter: manager-spec `skills: moai-foundation-core, moai-workflow-spec` (templates/.claude/agents/moai/manager-spec.md:13-15); manager-develop `moai-foundation-core` (:14-15); manager-docs `moai-foundation-core` (:13-14); sync-auditor `moai-foundation-quality` (:13-14). PRECISION: plan-auditor carries NO static `skills:` preload — its body states "This agent carries no static `skills:` preload" verbatim (plan-auditor.md:431), with on-demand read-only Skill() invocations only (moai-foundation-quality for TRUST 5 scoring; moai-ref-cross-model-audit for cross-model) — it contributes nothing to the static union; moai-foundation-quality enters via sync-auditor, not plan-auditor. |
| W2 | The three published command skills invoke the `moai` dispatcher | CONFIRMED | templates/.agents/skills/moai-plan/SKILL.md:9, moai-run:9, moai-sync:9 — each: "Harness without a skill loader (Codex CLI): read `.agents/skills/moai/SKILL.md` (the mirrored dispatcher body)". The reference is PROJECT-relative — gone after M4 (finding 2's premise). |
| W3 | The default flows' dispatcher routing injects further skills | CONFIRMED | templates/.claude/skills/moai/SKILL.md routing table: plan row :124 "Skills: moai-workflow-spec, moai-foundation-thinking (per delegation.yaml)"; run row :132 "moai-workflow-tdd, moai-workflow-ddd"; sync row :140 "moai-workflow-project"; run-ddd row :188 "moai-workflow-ddd". Tier-2 closure = {moai-foundation-thinking, moai-workflow-tdd, moai-workflow-ddd, moai-workflow-project}. |
| W4 | The published command skills are GENERATED artifacts with a source-level regeneration pipeline | CONFIRMED | Sources: `.claude/commands/moai/` consumed READ-ONLY by `internal/template/commandemit` (`CommandsRoot: ".claude/commands/moai"`, commandemit.go:51; body verbatim from source, commandemit.go:8). Regeneration: `make commands-emit` = `COMMAND_EMIT_UPDATE=1 go test ./internal/template/commandemit/... -run TestGoldenCommittedArtifactsMatchEmission` (Makefile:51-52). Drift guard: `commands-emit-check` read-only, rides the `build:` prerequisite chain (Makefile:34, :55-60) — a hand-edited copy fails the next build. |
| W5 | The AGENTS.md skill-path sentences needing rebind | CONFIRMED (with a numbering correction) | AGENTS.md.tmpl:40-41: "the deployed skill is in `.agents/skills/<name>/SKILL.md` for Codex and `.claude/skills/<name>/SKILL.md` for Claude. The `both` profile installs both paths." — project-relative, needs the user-folder rebind post-M4. CORRECTION: the gate called these "§3" sentences; in AGENTS.md.tmpl §3 is "Worktrees" and the skill-path sentences sit in the unnumbered preamble paragraph — the rebind scope is unchanged, the label was off. |
| W6 | The dispatcher's INTERNAL workflow references are project-relative and break post-M4 (fold A2) | CONFIRMED, count corrected upward | `Read .claude/skills/moai/workflows/<name>.md` appears EIGHTEEN times in templates/.claude/skills/moai/SKILL.md (:126 plan, :134 run, :142 sync — the fold's three — plus gate :150, e2e :158, goal :167, gtd :182, fix :190, and the remaining rows; `grep -c` = 18). All one class: every internal workflow reference rebinds, not only the L0 three. PRECISION on the dispatch wording: the dispatcher is NOT a commandemit output — templates/.claude/skills/moai/SKILL.md IS its source layer (deployed verbatim); the fix is a direct template-source edit (template-first rule), no emitter regeneration involved. |
| W7 | On-demand `Skill()` invoke sites in the L0 agent bodies (fold B1) | CONFIRMED, swept | Full sweep of the five bodies: manager-spec :40 (workflow-spec, already preload), :244 foundation-thinking, :245 foundation-quality, :246 workflow-ddd, :247 workflow-tdd, :248 workflow-testing, :249 workflow-project, :250 workflow-worktree; manager-develop :235 tdd, :236 ddd, :237 testing, :238 foundation-quality, :239 spec, :240 thinking, :241 project, :242 worktree; manager-docs :37/:220 project, :221 spec, :222 foundation-quality, :223 thinking, :224 domain-html-report; plan-auditor :237/:762/:764 foundation-quality (read-only), ref-cross-model-audit; sync-auditor :191/:226 ref-cross-model-audit, :223 ref-owasp-checklist, :224 ref-testing-pyramid, :225 foundation-core. NEW union members beyond the eight: `moai-workflow-testing`, `moai-workflow-worktree` (both measured present under templates/.claude/skills/). Classified OUT as per-mission domain injections: `moai-ref-cross-model-audit`, `moai-ref-owasp-checklist`, `moai-ref-testing-pyramid`, `moai-domain-html-report`. Union total: TEN. |
| W8 | `checkSkillsAllowlist` miswarns post-M4 (fold A3) | CONFIRMED | `internal/cli/doctor.go:957-958` (comment :957, `func checkSkillsAllowlist` :958): reads `filepath.Join(projectRoot, ".claude", "skills")` and sets `CheckWarn` ".claude/skills/ not found" on `os.IsNotExist` — a healthy post-M4 install (no project skills) draws the spurious warn exactly as the gate observed. Disposition pinned: REPOINT to the user-install path in M4 (not removed — the allowlist integrity check survives scoped to the user folders). |

The eight-skill closure union (W1-W3): `moai`, `moai-foundation-core`,
`moai-workflow-spec`, `moai-foundation-quality` (tier 1) +
`moai-foundation-thinking`, `moai-workflow-tdd`, `moai-workflow-ddd`,
`moai-workflow-project` (tier 2) — all eight measured present under
`templates/.claude/skills/`. The catalog L0 view enumerates this union
explicitly (design §2.3); the M0 drift guard pins it to the W1-W2 sources.

## 3. Gaps (explicitly NOT verified)

- Runtime behavior was NOT executed this run (no `moai init`, no `moai update`,
  no `moai doctor` invoked): all V-rows are file/line readings, not executions.
  Where a row's claim matters to a run-phase milestone, the milestone's own
  acceptance criteria re-measure it.
- The `.claude-plugin/marketplace.json` and `plugins/moai/**` contents were
  listed (names, counts) but not byte-audited; disposition (M6) removes them and
  the golden tests already pin their current bytes.
- The stale `.claude/agents/{core,expert,meta}` convention text (V12 note) was
  not traced to its owning doc owner; recorded as an observation, not repaired.
- Profile-visibility mechanics (whether Claude Code reads `$CLAUDE_CONFIG_DIR/skills`
  exclusively, or falls back to `~/.claude/skills`) were NOT measured — no
  Claude Code runtime was invoked, and none is needed for v1: D-Q3 is CLOSED
  at plan phase by constraint (decision-index POLICY-COVERED; premise P6) —
  profiles do not see the shared user assets in v1, a declared doc-visible
  limitation, because every install-into-profile option violates REQ-002 and
  C2. No run-phase measurement milestone exists for a policy choice that is
  not open (iter2 D20c correcting this stale pre-closure text).
- `git log` archaeology on WHY `pluginemit` golden tests pin the tree was not
  done; disposition follows the card, not history.

## 4. Residual Risk

- The plugin carrier was introduced by a completed SPEC (SPEC-PLUGIN-MARKETPLACE-001,
  status `completed`) with its own doctor checks and golden tests; retiring it
  touches ~8 test files (13 .go files total), 2 Makefile targets, the build
  chain and release chain, and the roster guard note — the blast radius is
  wide, and M6/M7 sequence it deliberately.
- If the operator later reverses D3/D4 (the dispatch calls them settled), M6/M7
  are the milestones to drop — they are ordered last for exactly that reason.
