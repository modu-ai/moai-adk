---
id: SPEC-PLUGIN-MARKETPLACE-001
title: "Make modu-ai/moai-adk a marketplace that carries a derived moai core plugin, installed at init and install time and version-checked by doctor"
version: "0.1.0"
status: draft
created: 2026-10-03
updated: 2026-10-03
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/template/pluginemit, internal/cli, plugins/moai, install scripts"
lifecycle: spec-anchored
tags: "plugin, marketplace, codex, install, doctor, version-alignment, derivation, design-card-2"
tier: M
depends_on: [SPEC-PLUGIN-LOAD-SCOPE-001]
---

# SPEC-PLUGIN-MARKETPLACE-001 — moai marketplace, derived core plugin, install step, version alignment

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-10-03 | manager-spec | Initial plan-phase draft (card t1435, Class C, Mods design card 2 of 6, operator instruction 2026-10-02). Built on the t1434 verdict (SPEC-PLUGIN-LOAD-SCOPE-001, completed) and on observations made in this run under scratch config homes at base tree 7109e0900. The design source named by the card (a claude.ai artifact) was not readable by the lane, so every design detail that the card text and the inputs do not fix is an Open Decision (§5), never an assumption. |

## 1. Background and Premise

### 1.1 What this SPEC does and does not do

`moai init` copies the full project scaffold into every project. The Mods design series moves part of
that payload into a plugin distributed through a marketplace. Card t1434 measured what a plugin can
carry (its verdict is the committed table in `.moai/specs/SPEC-PLUGIN-LOAD-SCOPE-001/progress.md §E.2`).
This card (t1435, 2 of 6) only **adds**: a marketplace served by the `modu-ai/moai-adk` repository
itself, a `moai` core plugin whose payload is derived from the template tree, an install step run by
`moai init` and by the three install scripts, and a `moai doctor` check that the installed plugin
version matches the binary. It **removes nothing** from the init payload (init shrink is card t1438).

### 1.2 Inputs and their limits

- Card text (operator instruction 2026-10-02): the repository becomes a marketplace
  (`.claude-plugin/marketplace.json`; for Codex a dual manifest, precedent the moai-cowork
  marketplace); `moai init` and the install script run `claude plugin marketplace add modu-ai/moai-adk`
  and `claude plugin install moai@moai-adk` when `claude` exists, fail-open with guidance on failure,
  Codex likewise; a doctor check compares the binary version with the plugin version.
- The t1434 verdict (claude 2.1.287, codex 0.160.0, 2026-10-02; every claim holds for those two
  versions only). Cells cited below as `R0x` are rows of that table.
- Observations made by this run at base tree 7109e0900 (§1.3). Commands that write a registry ran
  only under empty scratch config homes; nothing was run against the real Claude or Codex profile.
- Not available: the claude.ai design artifact, any authenticated session, any network call to
  `github.com/modu-ai/moai-adk` (the repository carries no marketplace manifest yet).

### 1.3 Observed premises

Rows were observed by the author in this run (tree 7109e0900) unless the source column says otherwise.
Raw outputs of the scratch-home commands are kept in the session scratchpad, not in the repository.

| # | Observation | Command or read | Deciding output (verbatim) | Source |
|---|-------------|-----------------|----------------------------|--------|
| P-01 | Tool versions | `claude --version`; `codex --version` | `2.1.287 (Claude Code)`; `codex-cli 0.160.0` | author |
| P-02 | Strict validate demands a marketplace description | `claude plugin validate <marketplace.json> --strict` (scratch home) | `❯ description: No marketplace description provided. Adding a description helps users understand what this marketplace offers`, exit 1 | author |
| P-03 | Strict validate compares entry version with plugin.json | same, entry `0.0.1` vs plugin.json `3.2.0-rc.23` | `❯ plugins[0].version: Entry declares version "0.0.1" but plugins/mini/.claude-plugin/plugin.json says "3.2.0-rc.23". At install time, plugin.json wins (calculatePluginVersion precedence) — the entry version is silently ignored.`, exit 1 | author |
| P-04 | A prerelease plugin version passes strict validate | `claude plugin validate <plugin dir> --strict`, version `3.2.0-rc.23` | `✔ Validation passed`, exit 0 | author |
| P-05 | Claude marketplace add is idempotent | `claude plugin marketplace add <local path> --json` twice (scratch home) | exit 0 both; second: `Marketplace 'mini-mkt' already on disk — declared in user settings` | author |
| P-06 | Claude install is idempotent and reports versions | `claude plugin install mini@mini-mkt --json` twice | exit 0 both; second: `Plugin "mini@mini-mkt" is already installed (scope: user) …` with `"installedVersion":"0.0.1","availableVersion":"0.0.1"` | author |
| P-07 | Install writes only under the config home | `ls -A <scratch home>/plugins` after P-06 | `cache`, `installed_plugins.json`, `known_marketplaces.json`, `marketplaces`; `claude plugin list --json` names `mini@mini-mkt`, `version` `0.0.1`, `scope` `user` | author |
| P-08 | Claude registry file shape | read of the real profile's `plugins/installed_plugins.json` (a file read, no command) | `{"version": 2, "plugins": {"<name>@<marketplace>": [{"scope", "installPath", "version", "installedAt", "lastUpdated"}]}}`; shape seen at 2.1.287 only | author |
| P-09 | Codex needs its own manifest path | `codex plugin marketplace add .` at base tree (scratch `CODEX_HOME`) | `Error: invalid marketplace file `<root>`: marketplace root does not contain a supported manifest`, exit 1 | author |
| P-10 | Codex marketplace and plugin add are idempotent | marketplace add and `codex plugin add mini@mini-mkt` twice, with `.agents/plugins/marketplace.json` present | exit 0 each; `Marketplace `mini-mkt` is already added from …`; `codex plugin list --json` has `"pluginId": "mini@mini-mkt"`, `"version": "0.0.1"`; cache root `<CODEX_HOME>/plugins/cache/mini-mkt/mini/0.0.1` | author |
| P-11 | moai-cowork dual-manifest shape | read of `~/.codex/.tmp/marketplaces/moai-cowork/.claude-plugin/marketplace.json`, `…/.agents/plugins/marketplace.json`, `~/.codex/plugins/cache/moai-cowork/moai-accountant/1.3.8/{.claude-plugin,.codex-plugin}/plugin.json` | Claude marketplace entries: `name`, `source: "./plugins/<name>"`, `version`, `description`, `category`, top-level `owner`, `metadata`. Codex marketplace entries: `name`, `source: {"source": "local", "path": "./plugins/<name>"}`, `policy: {"installation": "AVAILABLE", "authentication": "ON_INSTALL"}`, `category`, top-level `interface.displayName`. The Codex plugin manifest names `skills: "./skills/"`; the Claude one does not | author |
| P-12 | Marketplace source forms in the official catalog | read of `…/plugins/marketplaces/claude-plugins-official/.claude-plugin/marketplace.json` | `"source": "git-subdir"` with `url`,`path`,`ref`,`sha`; `"source": "url"` with `url`,`sha` | author |
| P-13 | Plugin subagents ignore three frontmatter fields | documentation, `https://code.claude.com/docs/en/sub-agents` § "Choose the subagent scope" (WebFetch this run) | "plugin subagents don't support the `hooks`, `mcpServers`, or `permissionMode` frontmatter fields. These fields are ignored when loading agents from a plugin." | doc |
| P-14 | Template agents depend on those fields | `grep` over `internal/template/templates/.claude/agents/moai/*.md` | 12 of 12 carry `permissionMode` (`plan`: super-advisor, sync-auditor; `default`: e2e-tester, plan-auditor; `acceptEdits` 2; `bypassPermissions` 6); 4 carry `hooks:` or `mcpServers:` (manager-docs, manager-develop, manager-spec, sync-auditor); 10 carry `skills:` | author |
| P-15 | Template component shape | `find`/`grep` over `internal/template/templates/.claude` | skills 312 files, 0 `.tmpl`; agents 12 files, 0 `.tmpl`; commands 17 files, 15 `.tmpl` and all 15 contain `{{`; hooks 48; rules 92; output-styles 3 | author |
| P-16 | Catalog tiers decide what init deploys | `grep -E '^\s+tier:' internal/template/catalog.yaml`; `internal/template/slim_fs.go:214` | 35 `core`, 13 `optional-pack:*`, 1 `harness-generated` entries; `SlimFS(rawFS, cat)` hides every non-core entry; init defaults to slim mode (`--all` bypasses, `internal/cli/init.go:85`) | author |
| P-17 | Emit-and-check precedent | `Makefile:34`, `Makefile:38-62` | `build: agents-emit-check commands-emit-check tool-policy-drift-check templ-generate`; regeneration only behind `agents-emit` / `commands-emit`; golden tests switched by `AGENTEMIT_UPDATE` / `COMMAND_EMIT_UPDATE` | author |
| P-18 | Install scripts carry no plugin step | `grep -n -i plugin install.sh install.ps1 install.bat` | no line in any of the three; `cmp` shows `docs-site/static/install.sh` and `install.ps1` identical to the root originals; `.github/workflows/test-install.yml` job `install-script-parity` diffs both | author |
| P-19 | init is the place the step attaches, and its MCP asymmetry | `internal/cli/init.go:192` (`wireCodexUnlessClaude`), `:659-669` | the non-interactive path leaves `MCPProvision` false (no project `.mcp.json` moai entry) while the interactive path provisions it; harness values are `claude`, `gpt`, `both` | author |
| P-20 | moai calls its own init and tests call it directly | `internal/cli/doctor_agentemit_embed.go:352`; `internal/cli/coverage_improvement_test.go:2427,2475,3831,3892,4349` | `exec.Command(execPath, "init", target, "--non-interactive", "--llm", "both")` with `os.Environ()`; five test call sites of `runInit` | author |
| P-21 | New doctor checks must register on guarded surfaces | `internal/cli/binary_lag_test.go:198`; `internal/cli/testdata/doctor-{light,dark,nocolor}.golden` | `namesAddedAfterBaseline` allowlist and `TestBinaryLag_DoctorCheckNameSetIsUnchanged`; three golden outputs | author |
| P-22 | Version SSOT and stamps | `pkg/version/version.go:11`; `.moai/config/sections/system.yaml`; `.goreleaser.yml:22`; `Makefile:20`; `.github/workflows/release.yml` check 6; `.moai/docs/version-management.md` § Files Requiring Version Sync | `Version = "v3.1.3"`; `version: v3.1.3`; the installed binary prints `moai-adk v3.2.0-rc.26`; bump commits rewrite seven stamp files by hand; the registry test sweeps tracked files for the literal `v3.1.3` token | author |
| P-23 | The t1399 rename has not landed in this base | `grep -rn moai-kanban-foreman` over `internal`, `.claude`, `.agents`, `.moai/config`; `grep -c moai-kanban-foreman internal/template/catalog.yaml` | the old name is referenced from `internal/template/catalog.yaml` (2), the template skill and its local copy, `.claude/loop.md` and its template copy, and eight Go files (two non-test, six test); `moai-factory-foreman` occurs nowhere in those trees | author |
| P-24 | main only advances through release PRs | `.claude/rules/local/repo-local-pr-policy.md` | "`main` advances ONLY through release pull requests (`release/vX.Y.Z` → `main`, merge-commit strategy …)" | author |
| P-25 | Command-form boundary in this worktree session | plain `go test <pkg> -run '^TestX$' -count=1 -v`, `go test -list '^TestX$'`, `sh scripts/<file> <arg>`, `jq`, `find`, `diff -r`, `make <target>` accepted; the same `-run '^TestX$'` bundled with `> file 2>&1; echo EXIT=$?` refused, as was `sh scripts/<file>` bundled with `> file; cat; grep` | refusal text: "this command runs go with the text ^TestEmitDerivesFromTree$ inside a construct too complex to verify" | author |
| P-26 | Paths the new manifests will use are not git-ignored | `git check-ignore -v .claude-plugin/marketplace.json plugins/moai/.claude-plugin/plugin.json plugins/moai/.codex-plugin/plugin.json .agents/plugins/marketplace.json` | no output, exit 1; control `git check-ignore -v .moai/reports/t1435/inputs/t1434-verdict.md` prints `.gitignore:235:.moai/reports/*` | author |
| P-27 | House convention for the add command | read of the moai-cowork marketplace `metadata.description` | "Claude Code에서 `claude plugin marketplace add modu-ai/moai-cowork`로 한 번에 등록하세요" (GitHub shorthand `owner/repo`) | author |

### 1.4 Placement of each component kind in the v1 plugin

Derived from the t1434 cells. `PLUGIN-OK` means a live Claude session listed the component; it does not
say the component behaves identically to its project copy (P-13, P-14).

| Component | Template source | In the v1 payload | Basis |
|-----------|-----------------|-------------------|-------|
| Skills | `.claude/skills/**`, catalog core tier | Yes | R01 PLUGIN-OK (namespaced `<plugin>:<skill>`) |
| Commands | `.claude/commands/moai/*` | Yes, rendered with the English default context | R03 PLUGIN-OK; Codex turns a plugin `commands/` into generated skills (R03-codex PARTIAL) |
| MCP server | `.mcp.json` `moai` entry | Yes, identical command and args | R06 PLUGIN-OK; duplicate suppressed when the project has the same command |
| Agents | `.claude/agents/moai/*.md` | Open (OD-3); not shipped under the default | R02 PLUGIN-OK for listing only; P-13/P-14: `permissionMode`, `hooks`, `mcpServers` ignored |
| Hook registrations | `settings.json.tmpl` | Open (OD-2); not shipped under the default | R07/R08 measured one registration; double registration unmeasured |
| Rules, instruction files | `.claude/rules/**`, `CLAUDE.md`, `AGENTS.md` | No | R09, R10, R11 PROJECT-ONLY; other carriers UNMEASURED |
| Output styles, plugin settings keys | `.claude/output-styles/**` | No | R05, R12 activation unobserved |
| Workflow scripts, `loop.md`, `.moai/**` config | `.claude/workflows`, `.moai` | No | not a measured component kind; stays in the scaffold |
| Mods | `mods/moai-board` | No | R04 time-bound; any later mod entry needs a project fallback |

### 1.5 Interactions detected before any code exists

These bind the requirements and are carried as risks in §6. None is resolved by this card's
scratch-home verification, because each needs an authenticated session.

1. **Two copies until t1438.** The scaffold copy and the plugin copy coexist. R01 shows plugin skills
   listed under a namespaced name; the project commands already list as `moai:<command>` in this
   session's skill listing, and a plugin named `moai` also yields `moai:<command>`. Whether the two
   merge, shadow or collide is UNOBSERVED, as is the listing cost of two copies.
2. **MCP decline.** `moai init --non-interactive` leaves no project `.mcp.json` moai entry (P-19) and
   the code comment says an explicit decline is honored absolutely. A plugin carrying a `moai` MCP
   server would start it anyway, because R06's suppression needs a project server with the same
   command to exist.
3. **Bare-name references.** Command and skill bodies call `Skill("moai")` and similar bare names; a
   plugin skill is namespaced. Whether a bare name resolves to the namespaced skill is UNMEASURED and
   becomes load-bearing at t1438.
4. **The doctor re-enters init.** The Agent Emit Embed check runs `init --non-interactive --llm both`
   (P-20); with an install step that command would act on the real profile.
5. **First-release gap.** `main` carries the new manifests only after the release PR that contains
   them (P-24); until then `claude plugin marketplace add modu-ai/moai-adk` has no manifest to read.

## 2. Requirements (GEARS)

Each requirement has exactly one acceptance criterion with the same number in `acceptance.md`.
Paths marked "(OD-4)" are the default of the decision in §5; the requirement text follows the decision.

- REQ-001: (Ubiquitous) The `modu-ai/moai-adk` repository shall carry at its root a Claude marketplace manifest, `.claude-plugin/marketplace.json`, that names the marketplace `moai-adk`, declares an owner and a `metadata.description`, and lists exactly one plugin entry, `moai`.
- REQ-002: (Ubiquitous) The repository shall carry at its root a Codex marketplace manifest, `.agents/plugins/marketplace.json`, that names the marketplace `moai-adk` and lists exactly one plugin entry, `moai`, in the shape the moai-cowork precedent uses (P-11): a local-path `source` object, a `policy` with `installation` and `authentication`, and a `category`.
- REQ-003: (Ubiquitous) The `moai` plugin root (`plugins/moai/`, OD-4) shall carry a Claude manifest, `.claude-plugin/plugin.json`, and a Codex manifest, `.codex-plugin/plugin.json`, that declare the same `name` (`moai`) and the same `version` as each other and as the entry of the Claude marketplace manifest, the Codex manifest additionally naming `skills: "./skills/"` and the `moai` MCP entry of REQ-006, as the moai-cowork Codex manifests do (P-11).
- REQ-004: (Ubiquitous) The plugin generator shall derive the payload's skills and commands, and its agents where OD-3 admits them, from the embedded template tree restricted to the catalog tier OD-8 names, and shall contain no skill, agent, command or rule name literal, so that renaming a template component changes the payload with no generator edit.
- REQ-005: (Ubiquitous) The generator shall copy every non-`.tmpl` payload file byte for byte and shall render every `.tmpl` payload file with the template default context (English variant), dropping the `.tmpl` suffix, so that the payload carries no template syntax and no locale-conditional text.
- REQ-006: (Ubiquitous) The plugin shall declare exactly one MCP server, `moai`, whose `command` and `args` are copied from the `moai` entry of the template `.mcp.json` rather than retyped, so that Claude's duplicate-command suppression (R06) applies wherever the project carries that entry.
- REQ-007: (Unwanted) The plugin payload shall not carry rules, instruction files (`CLAUDE.md`, `AGENTS.md`), settings, `.moai` configuration, output styles, workflow scripts, mods, or hook registrations, hook registrations being subject to OD-2.
- REQ-008: (Event-driven) When the template tree or the version SSOT changes while the committed payload is not regenerated, the build shall fail before compiling through a read-only drift check, `make plugin-emit-check`, wired ahead of `build`, and regeneration shall stay behind the explicit verb `make plugin-emit` that the check never invokes.
- REQ-009: (Event-driven) When `moai init` completes template deployment and the opt-out (REQ-011) is not set, the init flow shall, for each of `claude` and `codex` that resolves on PATH (Claude through the repository's pinned-binary resolution, so `MOAI_CLAUDE_BIN` and `llm.claude_bin` are honored), run `<tool> plugin marketplace add modu-ai/moai-adk` and then, only when that exited 0, `<tool> plugin install moai@moai-adk` (Codex: `codex plugin add moai@moai-adk`), each as a child of the invoking process whose environment is inherited unchanged, and shall print the config home it acted on.
- REQ-010: (Event-driven) When a plugin command exits non-zero, exceeds a bounded time, or its tool is absent from PATH, the install step shall print guidance that names the manual commands, shall leave the exit status of `moai init` unchanged, and shall treat an outcome the tool itself reports as already present (P-05, P-06, P-10) as success.
- REQ-011: (State-driven) While the opt-out (OD-5) is set, the install step shall run no `claude` or `codex` command; moai's own internal `moai init` invocations (the Agent Emit Embed doctor check) shall set it; and under `go test` the step shall run no external command unless a test injects a runner.
- REQ-012: (Event-driven) When `install.sh`, `install.ps1` or `install.bat` finishes installing the binary, it shall run the same install step under the same opt-out and fail-open contract (mechanism: OD-6), and `docs-site/static/install.sh` and `install.ps1` shall remain byte-identical to the root originals.
- REQ-013: (Event-driven) When `moai doctor` runs, the check "Plugin Version" shall read the installed `moai@moai-adk` version from the resolved Claude config home (`plugins/installed_plugins.json`) and from the resolved Codex home plugin cache, strip a leading `v` from it and from the binary version, report OK on equality, report the OD-7 severity on inequality with both versions and the remedy command, and report OK or info, never warn or fail, when no plugin is installed, the config home is absent, the registry is unreadable or of an unknown shape, or the binary is a development build.
- REQ-014: (Ubiquitous) The "Plugin Version" check shall be registered on every guarded doctor surface (the binary-lag check-name allowlist, the diagnostic status set, the golden outputs) and shall emit one summary line by default and detail lines only under `--verbose`.
- REQ-015: (Ubiquitous) The generator shall derive every version it writes into the four manifests from the version SSOT (the `pkg/version` fallback value that each bump commit rewrites), with a leading `v` stripped, so that the version-carrying fields of the plugin and marketplace manifests cannot drift from one another or from the binary release.
- REQ-016: (Event-driven) When a release tag is verified, the release workflow shall fail when the plugin version in the tagged tree is not equal to the tag with its leading `v` stripped, through a script that also runs offline: `scripts/check-plugin-version.sh <tag>`.

## 3. Constraints

- Tier M: 16 requirements and 16 acceptance criteria (the Tier M ceilings, `spec-workflow.md` § SPEC Complexity Tier). The hand-authored change set is about 30 files, above the 5-15 file guidance, because manifests, generator, init step, doctor check, three install scripts, two docs-site copies, Makefile, release script and workflow all move. Generated files (the committed payload, several hundred) are not counted. Tier M was kept to match the requested artifact set; a plan-auditor tier-up suggestion would add `design.md` and `research.md` and raise the ceilings to 25.
- Template-First: the payload is derived from `internal/template/templates/`; no file is added to the local `.claude/`, `.moai/` or `.agents/` without a template source, and nothing under `plugins/moai/` is edited by hand (it is generated and drift-checked).
- Template neutrality: the payload is derived from files the neutrality CI guard already covers (`.github/workflows/template-neutrality-check.yaml` paths are `internal/template/templates/**`); the committed copy sits outside those paths and is therefore protected by derivation plus REQ-008, not by the neutrality workflow itself.
- Verification uses only scratch-home commands. Every `claude plugin install|marketplace add` and every Codex `plugin add|marketplace add` in an acceptance command runs with `CLAUDE_CONFIG_DIR` or `CODEX_HOME` pointing at an empty directory under the session scratchpad. No acceptance command makes a network call (marketplace add uses a local path).
- The run phase inherits the t1434 forbidden-verb discipline: no install, uninstall, marketplace add, remove, update, enable, disable or configure against any real Claude or Codex home.
- No time estimates; ordering is by priority and milestone sequence.
- Acceptance commands use the plain single-invocation forms measured in this worktree session (P-25): anchored `-run '^Name$'`, no git `$()` nesting, no pipes, redirects or `; echo $?` bundled into one invocation.

## 4. Non-goals and Out of Scope

### Out of Scope — Init shrink

- Removing anything from the `moai init` payload is card t1438 (5 of 6). This card adds the plugin and the install step and changes no deployed file set.

### Out of Scope — Mods status UI and mod entries

- The mods status UI is card t1437. No mod (`mods/moai-board` included) is listed in the marketplace or shipped in the payload; a later mod entry needs a project-side fallback because R04 is time-bound.

### Out of Scope — Aside

- Card t1439 (Aside) is not touched.

### Out of Scope — Rules and instruction files by plugin

- Always-loaded rules, path-scoped rules, `CLAUDE.md` and `AGENTS.md` stay in the project scaffold (R09, R10, R11). Other carriers (a skill, a SessionStart hook returning context) are UNMEASURED and are not used here.

### Out of Scope — Other plugin entries and locales

- Optional-pack and harness-generated catalog entries as separate plugins, and per-locale plugin variants, are not built.

### Out of Scope — Plugin refresh and publishing

- `moai update` does not refresh the plugin (the doctor check only reports and names the remedy); no tag, push, release, pull request, GitHub issue or real-profile write is performed by this card.

## 5. Open Decisions

Ordered by how likely each is to change the plan. The default column is the behavior the run phase takes
if no verdict is recorded, so the run is not blocked; it is a fallback, not a recommendation. Rows are
mirrored in `decision-index.md` (Q-n = OD-n), whose verdict lines the Kickoff gate reads.

| OD | Question | Options | Default if unanswered | Who decides |
|----|----------|---------|-----------------------|-------------|
| OD-1 | While init still deploys the whole scaffold (until t1438), does init install the plugin by default? Effect of both copies coexisting is UNOBSERVED (§1.5-1). | (a) on by default now, as the card text says; (b) opt-in flag until t1438 lands, then default on; (c) on by default but the payload limited to components with no scaffold twin | (a) | operator |
| OD-2 | Which components that have a scaffold twin ship in v1: hook registrations and the MCP entry? Double firing of a hook registered in both places is UNMEASURED (R07/R08 measured one registration); the MCP entry interacts with decline (§1.5-2). | (a) MCP entry only, no hooks; (b) neither; (c) MCP entry plus hooks whose command needs no `$CLAUDE_PROJECT_DIR` | (a) | operator, with measurement in an authenticated session |
| OD-3 | Do agents ship in v1? Plugin agents ignore `permissionMode`, `hooks`, `mcpServers` (P-13) and all 12 template agents carry `permissionMode`; two are read-only by `permissionMode: plan` (P-14). | (a) ship all 12 verbatim; (b) ship only agents that carry none of the three fields (none qualify today); (c) ship no agents in v1 | (c) | operator |
| OD-4 | Where does the payload live and how does the marketplace entry pin it? A sha pin inside the repository cannot be written (a commit cannot contain its own hash); how a GitHub-hosted marketplace refreshes a relative-path entry is UNOBSERVED. | (a) committed generated tree `plugins/moai/`, entry `source: "./plugins/moai"`, no ref pin (tracks the default branch, which advances only by release PRs); (b) option (a) plus a tag `ref` stamped before tagging; (c) tree not committed, built at release into a release-only ref | (a) | operator and release maintainer |
| OD-5 | What is the opt-out surface? Install scripts run before any project exists, so a project config key cannot cover them. | (a) `--no-plugin` on `moai init` plus env `MOAI_SKIP_PLUGIN_INSTALL=1` honored by init and all three scripts; (b) per-tool switches in addition; (c) a user-scope config key | (a) | operator |
| OD-6 | How do the install scripts run the step? | (a) call a `moai` verb after the binary lands (one Go implementation, injected-runner tests); (b) each script inlines the two tool commands, the card's literal text (three implementations plus two docs-site copies, no unit seam) | (a) | operator |
| OD-7 | What severity does a version mismatch carry in `moai doctor`? | (a) warn; (b) fail; (c) info | (a) | operator |
| OD-8 | Which catalog tiers does the `moai` plugin carry? The card says "moai core plugin"; init's default slim mode deploys the core tier (P-16). | (a) core tier only; (b) every tier; (c) core now, optional packs as later plugin entries | (a) | operator |

## 6. Risks

| # | Risk | Observed or inferred | Handling |
|---|------|----------------------|----------|
| RK-1 | Two copies of skills, commands (and agents or MCP under other ODs) until t1438: shadowing, collision or doubled listing cost | inferred from R01/R03/R06; effect UNOBSERVED | OD-1, OD-2, OD-3; the first runtime observation needs an authenticated session and belongs to the run-phase gap list and to t1438 |
| RK-2 | R06 duplicate suppression: the dedupe key (command only, or command and args) is observed only with a fixture that matched on command; a project `.mcp.json` that is absent (non-interactive init, or an explicit decline) lets the plugin start the server | R06 plus P-19 | REQ-006 copies command and args; OD-2; OD-5 opt-out |
| RK-3 | R04 mods are time-bound (rollout switch observed off by other observers) | t1434 verdict condition 1 | no mod in the payload (REQ-007); a later mod entry needs a project fallback |
| RK-4 | t1399 renames `moai-kanban-foreman` to `moai-factory-foreman` and renames rule files; hand lists would break | P-23 | REQ-004 derivation; mechanical consequence: whichever card lands second turns `plugin-emit-check` red until `make plugin-emit` is run in its merge, which is the intended signal; land-order precondition in `plan.md` §6 |
| RK-5 | Codex acceptance proves nothing: `codex plugin add` accepted every manifest in t1434, and agents and hooks are installed-not-observed-active | R14-codex, R02-codex, R07-codex | AC-002 pins shape parity with the precedent and local install success only; field effect is UNOBSERVED and stated as such |
| RK-6 | Plugin agents ignore `permissionMode`, `hooks`, `mcpServers`; read-only agents would run with the session's permission mode | P-13, P-14 | OD-3 |
| RK-7 | Hooks registered in both the project settings and the plugin may fire twice | inferred; UNMEASURED | OD-2; hooks excluded by default |
| RK-8 | The install step reaches the real profile through moai's own `init` re-entry or through tests | P-20 | REQ-011; guard test that no test binary reaches the real runner |
| RK-9 | Until the release PR lands the manifests on `main`, `marketplace add modu-ai/moai-adk` fails for every build that carries the step | P-24 | REQ-010 prints guidance and exits 0; the manifests and the step ship in one tag; the first-release gap is recorded, not hidden |
| RK-10 | Payload commands carry English text only; Korean, Japanese and Chinese descriptions of the scaffold copies are not reproduced | derived from REQ-005 | stated; revisit at t1438 |
| RK-11 | Bare `Skill("moai-…")` and agent names versus namespaced plugin names | UNMEASURED | does not bite while the scaffold stays; recorded for t1438 |
| RK-12 | Windows: `claude` and `codex` resolution and `install.bat` behavior | no Windows host in this run | UNVERIFIED locally; CI job `test-install.yml` runs `install.bat` end to end on Windows and must stay green with the step present |
| RK-13 | The registry file shape (`"version": 2`) and the cache layout are observed at claude 2.1.287 and codex 0.160.0 only | P-08, P-10 | REQ-013 fails open on an unknown shape |
| RK-14 | Tier M sits at the top of its file guidance | §3 | stated; tier-up is the plan-auditor's call |

## 7. Dependencies and Prior Art

- `depends_on`: SPEC-PLUGIN-LOAD-SCOPE-001 (status `completed`; its `progress.md §E.2` is the durable verdict table).
- Land-order dependency: card t1399 (lane 3) renames a skill and rule files. This card's run phase starts either from a base that already contains t1399 or relies on the derivation (REQ-004) and regenerates after the other card lands (§6 RK-4).
- Prior art, read only: the moai-cowork marketplace and its cached plugins (P-11); the local plugin `mods/moai-board/.claude-plugin/plugin.json` (a single-manifest plugin with no marketplace entry); `commandemit` and `agentemit` (emit, golden-pin, read-only check).
