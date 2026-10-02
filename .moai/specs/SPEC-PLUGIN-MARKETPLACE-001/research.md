# Research — SPEC-PLUGIN-MARKETPLACE-001

The observations behind each premise row of `spec.md` §1.3, with the command, the deciding output, the source and the version
stamp. Rows P-01 to P-27 were observed at iteration 0 (tree `7109e0900`) and re-checked by plan-audit iteration 1
(`.moai/reports/t1435/plan-audit-iter1.md`, Evidence §B and §C); this file does not repeat their raw output. Rows P-28 to P-41 and the
corrections to P-19 and P-20 were observed in the revision at tree `3766cef05` and are recorded below as R-nn, in the order the
`spec.md` last column cites them. Rows P-42 to P-46 (R-23 to R-29) were observed in the revision after plan-audit iteration 2, at tree
`b6a0522a0`, whose diff to the audited tree `d6987e59c` is the SPEC's own `progress.md` only. A statement here is an observation unless
it says `read` (source or document read, not executed) or `inferred`.

## 1. Stamps

| Item | Value |
|------|-------|
| Claude Code | `2.1.287 (Claude Code)` |
| Codex CLI | `codex-cli 0.160.0` |
| Go | `go1.26.8 darwin/arm64`; `go.mod` declares `go 1.26.8` |
| Platform | Darwin 27.0.0, arm64 |
| Tree measured | `3766cef05` (R-01 to R-22), `b6a0522a0` (R-23 to R-29), branch `WT-marketplace-core-plugin`, clean at start; HEAD and branch re-read at the end of each revision |
| Judging build for repository tooling | `go build -o <scratch>/moai-t1435 ./cmd/moai` at `3766cef05`; `moai-t1435 version` printed `moai-adk v3.1.3` / `v3.1.3   none   built unknown` (no ldflags: the commit stamp is by build procedure, not self-attested). The installed `moai` is v3.2.0-rc.26 |
| Date | 2026-10-03 |

`<scratch>` is `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/2f10c8c5-67ea-41c2-9b61-6242acc465c3/scratchpad/s2`.

## 2. Discipline of this revision

- Every `claude plugin …` and `codex plugin …` command that writes a registry ran with `CLAUDE_CONFIG_DIR` or `CODEX_HOME` set to
  an empty directory under `<scratch>` (`ls -A` showed `total 0` before first use); the local-path source was a fixture under
  `<scratch>`. No command touched the real Claude profile or `~/.codex`, and `stat` of neither was taken before or after because no
  command aimed at them; the two reads of real-home files were the moai-cowork manifests (R-06), as at iteration 0.
- Four read-only network reads were made, none of them a plugin or marketplace command: one documentation fetch (R-18) and three
  GitHub metadata reads (R-17). No acceptance command in `acceptance.md` makes a network call.
- Fixtures were built by a script in the scratchpad (`build_fx.py`) from the template tree: the 24 core skills named by
  `internal/template/catalog.yaml`, the 17 command sources with each `{{if eq .ConversationLanguage …}}…{{else}}X{{end}}` chain
  replaced by its `else` text (the English render; the real generator will use the template renderer), the template `.mcp.json` or a
  one-server version, and hand-written manifests carrying version `3.1.3`. They approximate the plan's payload shape and are not the
  payload.

## 3. Observations

### R-01 — A nested `commands/<dir>/` layout is not counted (P-30; D6)

Fixtures `fx-core-flat` (commands at `plugins/moai/commands/<name>.md`) and `fx-core-nested` (the same 17 files at
`plugins/moai/commands/moai/<name>.md`), each with the 24 core skills; homes `h-core-flat` and `h-core-nested`, both empty.

```
CLAUDE_CONFIG_DIR=<scratch>/h-core-flat claude plugin validate <scratch>/fx-core-flat/.claude-plugin/marketplace.json --strict
  -> ✔ Validation passed
CLAUDE_CONFIG_DIR=<scratch>/h-core-flat claude plugin validate <scratch>/fx-core-flat/plugins/moai --strict
  -> ✔ Validation passed
CLAUDE_CONFIG_DIR=<scratch>/h-core-flat claude plugin marketplace add <scratch>/fx-core-flat --json
  -> {"command":"marketplace-add","outcome":"ok","marketplace":"moai-adk","message":"Successfully added marketplace: moai-adk (declared in user settings)"}
CLAUDE_CONFIG_DIR=<scratch>/h-core-flat claude plugin install moai@moai-adk --json
  -> {"command":"install","outcome":"ok","plugin":"moai@moai-adk","pluginId":"moai@moai-adk","scope":"user","message":"Successfully installed plugin: moai@moai-adk (scope: user)"}
CLAUDE_CONFIG_DIR=<scratch>/h-core-flat claude plugin details moai@moai-adk
  -> Skills (41)  clean, codemaps, e2e, feedback, fix, gate, goal, gtd, harness, loop, moai, moai-domain-design-dna,
     moai-domain-html-report, moai-domain-svg-infographic, moai-foundation-cc, moai-foundation-core, moai-foundation-quality,
     moai-foundation-thinking, moai-harness-learner, moai-jev-skill-suggestion, moai-kanban-foreman, moai-lane-watchdog,
     moai-meta-harness, moai-ref-git-workflow, moai-ref-jev-question-design, moai-ref-testing-pyramid, moai-workflow-ddd,
     moai-workflow-docs-claim-check, moai-workflow-loop, moai-workflow-project, moai-workflow-spec, moai-workflow-tdd,
     moai-workflow-testing, moai-workflow-worktree, mx, plan, project, review, run, sync, todo
```

The same four steps for `fx-core-nested`, then:

```
CLAUDE_CONFIG_DIR=<scratch>/h-core-nested claude plugin details moai@moai-adk
  -> Skills (24)  moai, moai-domain-design-dna, moai-domain-html-report, moai-domain-svg-infographic, moai-foundation-cc, …
     (the 24 skill names; none of clean … todo)
```

Reading. Flat commands are counted (41 = 24 skills + 17 commands); nested commands are installed and not counted (24). Whether the
runtime *registers* a nested file for invocation is unobserved: only the `details` inventory was observed. The auditor's iteration-1
fixture reproduced the same split with two commands (`Skills (39)` against `Skills (37)`). `claude plugin details` has no JSON form:
`claude plugin details … --json` printed `error: unknown option '--json'`.

### R-02 — Offline always-on estimates (P-31; D13)

Same command, the tool's own projected cost lines:

| Fixture | Entries | `Always-on` line |
|---------|---------|------------------|
| core-only, commands flat (24 skills + 17 commands) | 41 | `~3,653 tok   added to every session` |
| core-only, commands nested (24 skills counted) | 24 | `~3,286 tok   added to every session` |
| all tiers, commands flat (37 skills + 17 commands), `h-all-flat` | 54 | `~6,483 tok   added to every session` |

The tool prints "Token counts are estimates and may differ from actual usage." The figures describe the plugin side only; the
scaffold's own listing cost is not reported by any offline command (`details` covers plugins), so "two copies cost about twice"
is inferred. The auditor's iteration-1 all-tier fixture with two commands printed `~6,161 tok` (39 entries), consistent with
54 entries at `~6,483`. Component rows with the largest always-on share in the all-tier run: `moai-ref-secops ~370`,
`moai-ref-seo ~360`, `moai-ref-supply-chain ~320`, `moai-ref-llm-security ~290`, `moai-domain-design-dna ~260`.

### R-03 — MCP server count with one entry (P-31, P-07)

After rewriting the fixture's `plugins/moai/.mcp.json` to carry only `mcpServers.moai` (`{"command":"moai","args":["mcp-server"]}`),
`claude plugin details moai@moai-adk` under `h-core-flat` printed `MCP servers (1)  moai  (tool schemas resolved at runtime; not
counted)` and an unchanged `Always-on: ~3,653 tok`. With the template's own `.mcp.json` (two servers) it had printed `MCP servers (2)
moai, context7`. The template file also carries `staggeredStartup`. The edit to the fixture was reflected in the installed plugin
without reinstalling.

### R-04 — Strict validation of the core-only fixture (P-02 to P-04)

Both `claude plugin validate … --strict` runs of R-01 passed for a marketplace with `owner`, `metadata.description`,
`metadata.version`, an entry carrying `version`, and a plugin manifest carrying `author`, `homepage`, `repository` and `license`.
This covers manifests only; it does not validate skill or command bodies.

### R-05 — Codex installed state is `config.toml` registration, not the cache (P-34; D10)

Fixture `fx-core-flat` gained `.agents/plugins/marketplace.json` (the moai-cowork shape, no `version`) and
`plugins/moai/.codex-plugin/plugin.json` (`version` `3.1.3`, `skills`, `mcpServers.moai`). Scratch `CODEX_HOME=<scratch>/cx-d10`, empty.

```
CODEX_HOME=<scratch>/cx-d10 codex plugin marketplace add <scratch>/fx-core-flat
  -> Added marketplace `moai-adk` from <scratch>/fx-core-flat.
     Installed marketplace root: <scratch>/fx-core-flat
CODEX_HOME=<scratch>/cx-d10 codex plugin add moai@moai-adk
  -> Added plugin `moai` from marketplace `moai-adk`.
     Installed plugin root: <scratch>/cx-d10/plugins/cache/moai-adk/moai/3.1.3
CODEX_HOME=<scratch>/cx-d10 codex plugin list --json
  -> {"installed":[{"pluginId":"moai@moai-adk","name":"moai","marketplaceName":"moai-adk","version":"3.1.3","installed":true,
      "enabled":true,"source":{"source":"local","path":"<scratch>/fx-core-flat/plugins/moai"},"marketplaceSource":{"sourceType":
      "local","source":"<scratch>/fx-core-flat"},"installPolicy":"AVAILABLE","authPolicy":"ON_INSTALL"}],"available":[]}
config.toml -> [marketplaces.moai-adk] source_type = "local" source = "<scratch>/fx-core-flat" ; [plugins."moai@moai-adk"] enabled = true
```

After deleting the `[plugins."moai@moai-adk"]` stanza from the scratch `config.toml` (an edit of a scratch file):

```
CODEX_HOME=<scratch>/cx-d10 codex plugin list --json  ->  { "installed": [], "available": [] }
ls <scratch>/cx-d10/plugins/cache/moai-adk/moai      ->  3.1.3
```

So a doctor that read the cache directory would report a plugin the tool does not consider installed. The supported
`codex plugin remove` clears the cache (auditor, iteration 1), so the divergent state needs a hand edit or another tool; the
hazard is real and narrower than "any removal".

Latency and a side effect. `/usr/bin/time -p codex plugin list --json` printed `real 0.02` in `cx-d10` and `real 0.01` in an empty
home; `/usr/bin/time -p claude plugin list --json` printed `real 0.26`. Run against the empty home `<scratch>/cx-empty`
(`ls -A` showed `total 0`), the CLI left a `tmp/` directory in it (`find` showed `tmp/arg0`): starting the CLI writes into its
home even for a read-only verb. `claude plugin list --json` under the installed fixture printed
`"id": "moai@moai-adk", "version": "3.1.3", "scope": "user", "enabled": true` and an `installPath` under
`<home>/plugins/cache/moai-adk/moai/3.1.3`.

### R-06 — Where each tool keeps the plugin version (P-35, P-11; D11, D22)

```
ls -d ~/.codex/plugins/cache/moai-cowork/*/*/.codex-plugin/plugin.json | wc -l          -> 18
grep -l '"mcpServers"' <those>  | wc -l                                                  -> 12
grep -l '"interface"'  <those>  | wc -l                                                  -> 18
grep -L '"version"'    <those>  | wc -l                                                  -> 0
grep -c '"version"' ~/.codex/.tmp/marketplaces/moai-cowork/.agents/plugins/marketplace.json   -> 0
grep -c '"version"' ~/.codex/.tmp/marketplaces/moai-cowork/.claude-plugin/marketplace.json    -> 19
```

The read of `~/.codex/plugins/cache/moai-cowork/moai-accountant/1.3.8/.codex-plugin/plugin.json` showed `name`, `version`
(`1.3.8`, equal to the cache directory name), `description`, `author`, `homepage`, `repository`, `license`, `keywords`, `skills`,
an inline `mcpServers` object and an `interface` block with `displayName`, `shortDescription`, `longDescription`, `developerName`,
`category`, `capabilities`, `websiteURL`, `defaultPrompt`, `brandColor`. The Codex marketplace read earlier in R-06's source
(`…/.agents/plugins/marketplace.json`) has `$schema`, `$comment`, `name`, `interface` (`displayName`, `websiteURL`) and `plugins[]`
entries with `name`, `source`, `policy`, `category` and no `version`. The cache of all marketplaces held 54 Codex plugin manifests
when counted over every marketplace, 19 with `mcpServers` and 52 with `interface`; the 18, 12 and 18 above are the moai-cowork subset,
the precedent. `~/.codex` was read, never written.

### R-07 — Scope and prompt flags (P-33; OD-12)

```
claude plugin install --help        -> -s, --scope <scope>  Installation scope: user, project, or local (default: "user")
                                       -y, --yes … (required when stdin or stdout is not a TTY)
claude plugin marketplace add --help -> --scope <scope>  Where to declare the marketplace: user (default), project, or local
claude plugin update --help         -> Update a plugin to the latest version (restart required to apply);
                                       -s, --scope <scope>  Installation scope: user, project, local, managed (default: auto-detect)
codex plugin add --help             -> options: -c, -m/--marketplace, --enable, --json, --disable, -h; no scope option
codex plugin --help                 -> add, list, marketplace, remove, help
codex plugin marketplace --help     -> add, list, upgrade ("Refresh configured Git marketplace snapshots"), remove, help
```

Help text only: `--scope project` and `--scope local` were not run, so the files they write and the registry shape they produce are
unobserved (G-5).

### R-08 — The locale text in the payload commands (P-32; OD-11, D14)

```
grep -c '{{' internal/template/templates/.claude/commands/moai/*.tmpl   -> 1 or 2 per file over the 15 .tmpl files
grep -h '{{' <the 15 files> | grep -v '^description:\|^argument-hint:' | wc -l   -> 0
grep -rhoE '\{\{[^}]*\}\}' <the 15 files> | sed -E 's/"[a-z]{2}"/"XX"/g' | sort | uniq -c
  -> 36 {{else if eq .ConversationLanguage "XX"}}   18 {{if eq .ConversationLanguage "XX"}}   18 {{end}}   18 {{else}}
sed -n 1,15p …/plan.md.tmpl  ->  description: {{if eq .ConversationLanguage "ko"}}EARS 형식 …{{else}}Create SPEC document with EARS format
                                  requirements and acceptance criteria{{end}}     and the same shape for argument-hint
```

`internal/template/context.go:88` sets the default `ConversationLanguage` to `"en"`. The two non-`.tmpl` commands (`gtd.md`, `todo.md`)
carry no braces. Neither of the 17 command bodies carries a template action.

### R-09 — The harness gates init's tool-specific wiring (P-28, P-19; OD-9, D19)

Read, `internal/cli/init.go`: `:133` the flag help `LLM harness to deploy and wire: claude, gpt, or both (default: claude; gpt deploys AGENTS.md +
Codex surfaces only — no .claude/ tree; both adds Codex wiring to the claude deployment)`; `:162` `case agentWiringGPT, agentWiringBoth`;
`:192-195` `func wireCodexUnlessClaude(…)  { if wiring == agentWiringClaude { return } … }`; `:997-1003` `mcpDeclined := !opts.MCPProvision;
switch agentWiringSelection { case agentWiringGPT: mcpDeclined = true; case agentWiringBoth: mcpDeclined = false }`; `:1011`
`wireCodexUnlessClaude(cmd, agentWiringSelection, opts.ProjectRoot)`. `:659-669` `opts.MCPProvision = true` on the interactive path only.
`.moai/specs/SPEC-INIT-HARNESS-PROMPT-001` is `status: completed`; its REQ-IHP-009 (line 116) states "The harness selection, whatever its
origin (flag or wizard), shall determine whether the MCP-entry provisioning call runs, by one rule: `codex` declines it, `both` forces it,
and `claude` leaves the `mcp_provision` answer intact." That answers the MCP-provisioning question, not whether a plugin install follows the
harness, so `decision-index.md` Q9 cites it as context and not as an authority anchor.

### R-10 — Profile resolution (P-29; OD-10)

Read: `internal/cli/launcher.go:159` `resolved := profile.ResolveLaunchProfileForProject(root, profileName)`;
`internal/profile/profile.go:461-495` — explicit name wins; `MOAI_NO_PROFILE_FALLBACK=1` disables; otherwise a launch ledger in
`GetBaseDir()` maps a project key (or the nearest registered ancestor project) to a profile name, usable only while that profile's
directory exists; no entry returns `""`. `init.go:566` reads `profile.GetCurrentName()`, which follows `CLAUDE_CONFIG_DIR`
(`profile.go:67-92`). Not executed.

### R-11 — Callers of `runInit` and `initCmd.RunE` (P-20; D5)

```
grep -rn 'runInit(' internal/cli/*_test.go | wc -l                       -> 33
grep -rln 'runInit(' internal/cli/*_test.go | wc -l                      -> 18
grep -rn 'runInit(' internal/cli --include='*.go' | grep -v _test.go     -> internal/cli/init.go:426:func runInit(cmd *cobra.Command, args []string) (err error) {
grep -rn 'initCmd.RunE(' internal --include='*.go' | wc -l               -> 9      (5 files: agent_model_flags_retired_test 1, coverage_test 2,
                                                                                 init_channel_test 1, init_deploy_exit_test 1, init_test 4)
```

Per file `runInit(` counts: `coverage_improvement_test.go` 10; `update_settings_snapshot_test.go`, `remaining_coverage_test.go`,
`init_update_profile_path_test.go`, `init_update_notice_test.go`, `init_gitdetect_test.go`, `init_autonomy_wiring_test.go` 2 each; eleven
files 1 each. One of the 33 lines is the string literal `strings.Index(body, "func runInit(")` (`update_settings_snapshot_test.go:475`),
which is not a call, so 32 call sites; the auditor's count of 32 in 18 files is confirmed. Non-test code starts moai's own init once:
`exec.Command(execPath, "init", target, "--non-interactive", "--llm", "both")` at `internal/cli/doctor_agentemit_embed.go:352` with
`cmd.Env = append(os.Environ(), "AGENTEMIT_UPDATE=")`; the other `"init"` literals in non-test Go are `git init` helpers in three fixture packages, a
help-order list (`help_order.go:32`), two session-worktree labels in `init.go` (`:449`, `:457`) and a comment in the hook types. The earlier P-20 figure of "five call sites" listed the five lines of `coverage_improvement_test.go` that
were visible at iteration 0.

### R-12 — The e2e script (P-20)

`e2e/cli/tux3_journeys.sh:104`: `run_to j1-init 180 "$SANDBOX" "NO_COLOR=1 '$BIN' init proj-j1 --non-interactive --language go --git-mode manual"`;
`:115`: `run_to j1b-init-notty 120 "$SANDBOX" "NO_COLOR=1 '$BIN' init proj-j1b"`. `SANDBOX` defaults to `/tmp/moai-e2e`; the script
does not scrub `HOME` or `PATH`. `grep -rln tux3_journeys . --include='*.yml' --include='*.yaml' --include='Makefile' --include='*.md'
--include='*.sh'` lists only `CHANGELOG.md` and SPEC documents; no workflow references the script.

### R-13 — The base `install.sh`, offline (P-37; D8)

`install.sh` has `set -e` (line 5), global `TARGET_PATH` set in `install_binary` (`:274`), and ends `main "$@"` (`:387`). Stub `curl` answered
the archive URL with a pinned local `moai-adk_9.9.9_darwin_arm64.tar.gz` (a stub `moai` script) and failed the checksum URL, which the script
tolerates. Command:
`PATH=<scratch>/inst/shim:/usr/bin:/bin /bin/bash <worktree>/install.sh --version 9.9.9 --install-dir <scratch>/inst/bin` → exit 0,
`[SUCCESS] Installed to: <scratch>/inst/bin/moai`, `[WARNING] Installation completed, but 'moai' command not found in PATH`,
`[SUCCESS] Installation complete!`. Recorded calls: none (`moai-calls.log` and `tool-calls.log` absent); `curl.log` two lines. A first
attempt that prefixed `HOME=<scratch>/inst/home` was refused (R-14); the run above set `PATH` only. `install.ps1` has
`$ErrorActionPreference = "Stop"` at line 6 and a `Verify-Installation` taking `$TargetPath` (`:377`); `install.bat` sets `TARGET_PATH` at
line 145, runs `"%TARGET_PATH%" version` at 168 and ends `exit /b 0` at 192. `.github/workflows/test-install.yml` has the parity job (`:43-54`),
a Unix job (`bash install.sh --help`, `bash -n install.sh`), pwsh jobs (`-help`, a PSParser tokenise, a step named `Test irm pipe iex compatibility`) and a Windows job
that runs the real `install.bat` against the network.

### R-14 — Command forms the worktree guard refused (P-36; `verification-claim-integrity.md` §3.1)

```
pwsh -NoProfile -Command '$PSVersionTable.PSVersion.ToString()'   -> refused: "… this command runs pwsh in a plain command; what it reads
                                                                    or is handed as shell text cannot be shown not to run git. Refusing to run it"
/usr/local/bin/pwsh -NoProfile -Command …                          -> the same refusal
PATH=<…> HOME=<…> /bin/bash install.sh …                           -> refused: "… this command sets HOME, injecting git configuration whose
                                                                    effect on where git writes can't be verified. Refusing to run it"
a heredoc-written script chained with `cd … &&` and three python3 runs -> refused: "this command is too complex to verify that it stays
                                                                    inside the worktree"
```

Accepted forms: `PATH=<dir>:/usr/bin:/bin /bin/bash <script> <args>`, `CLAUDE_CONFIG_DIR=<dir> <binary> …`, `CLAUDE_CONFIG_DIR=<dir>
CODEX_HOME=<dir> <binary> …`, `go test … -run '^X$' -count=1 -v`, `sh scripts/<file> <arg>`. `pwsh` exists (`which pwsh` →
`/usr/local/bin/pwsh`), so the Gap is a refusal, not an absence: PowerShell behavior was not executed.

### R-15 — `doctor --check` with an unregistered name (P-38; D9)

`CLAUDE_CONFIG_DIR=<scratch>/red/c-ac022 CODEX_HOME=<scratch>/red/x-ac022 <scratch>/moai-t1435 doctor --check "Plugin Version"` printed
the System Diagnostics box with `Pass 0    Warn 0    Fail 0` and no row, no error. The same command with the registered name `Plugin Deployment`
printed progress lines (`○ Plugin Deployment`, `✓ Plugin Deployment`), the table row `ok      Plugin Deployment  no plugin marker (binary-managed)`,
`1 ok, 0 warn, 0 fail` and `Pass 1    Warn 0    Fail 0`. `internal/cli/testdata/doctor-{nocolor,light,dark}.golden`
each hold one `MCP Server Version` row (`grep -c` printed 1 in all three) and `doctor-nocolor.golden` has `Plugin Deployment   no plugin marker
(binary-managed)` at line 31, the legacy check (`doctor.go:1120`).

### R-16 — The `plugin` verb is absent (P-39)

`<scratch>/moai-t1435 plugin install --help` → exit 1, `Unknown command "plugin" for "moai".` / `Try --help for usage.` `moai spec --help`
lists `status`, `drift`, `view`, `lint`, `close`, `audit`, `archive`; no `plugin` noun exists in `internal/cli` (`grep -rn 'Use: *"plugin'`
selected no line).

### R-17 — GitHub state (P-41; OD-4)

```
gh repo view modu-ai/moai-adk --json defaultBranchRef,visibility,isPrivate
  -> {"defaultBranchRef":{"name":"main"},"isPrivate":false,"visibility":"PUBLIC"}
gh api repos/modu-ai/moai-adk/contents/.claude-plugin --jq '.[].name'
  -> {"message":"Not Found","documentation_url":"https://docs.github.com/rest/repos/contents#get-repository-content","status":"404"}   gh: Not Found (HTTP 404)
gh api repos/modu-ai/moai-adk/contents/plugins --jq '.[].name'     -> the same 404
```

These are metadata reads of the public repository, made once, 2026-10-03. They close the iteration-1 Gap on the default branch and show
that neither the marketplace directory nor `plugins/` exists on `main`.

### R-18 — Plugin subagent restrictions, re-fetched (P-13)

WebFetch of `https://code.claude.com/docs/en/sub-agents`, 2026-10-03, section "Choose the subagent scope": "For security reasons, plugin
subagents don't support the `hooks`, `mcpServers`, or `permissionMode` frontmatter fields. These fields are ignored when loading agents
from a plugin. If you need them, copy the agent file into `.claude/agents/` or `~/.claude/agents/`. …" and "If you're the plugin's author,
ship the hooks in the plugin's `hooks/hooks.json` and the MCP servers in its `.mcp.json` instead." This closes the iteration-1 Gap that the
quote was unverified.

### R-19 — Release surfaces (P-40; D23)

`.moai/docs/version-management.md:123-144`: "Version Stamps" lists README.md, README.ko.md, README.ja.md, README.zh.md,
`.moai/config/sections/system.yaml`, `docs-site/hugo.toml`, `pkg/version/version.go`; "Release Artifacts" lists `CHANGELOG.md` and the Korean
release notes; line 144 names the guard `internal/cli/version_sync_list_test.go`, which fails when a listed stamp path is missing.
`.github/workflows/release.yml` job `verify-provenance` starts at line 45 and holds checks 1 to 7 (`:71-120`); check 6 compares the SSOT
`system.yaml` version with the tag. `scripts/release.sh` carries no stamp rewrite (`grep -n 'version.go\|system.yaml\|hugo.toml\|README'`
selected no line), so bumps are by hand. The release harness's own runbook was not found under `.claude/` or `.moai/harness`.

### R-20 — The doctor golden harness (P-21; D4, D9)

Read, `internal/cli/doctor_golden_test.go`: `:106` `t.Setenv(config.EnvConfigCacheDisabled, "1")`; `:111` `t.Setenv("HOME", t.TempDir())`; `:112`
`t.Setenv("MOAI_HOME", "")`; `:116` `t.Setenv(config.EnvClaudeConfigDir, "")`; `:123-124` Anthropic base URL and harbor-kite scrub; `:131-138`
`codexWiringLookPath` stubbed. `resolveCodexHomeDir` (`internal/cli/mcp_codex.go:2143`) returns `CODEX_HOME` when it is non-blank, else
`filepath.Join(os.UserHomeDir(), ".codex")`; because `HOME` is pinned, only a `CODEX_HOME` set in the caller's environment escapes the harness.
`internal/cli/binary_lag_test.go:198` `namesAddedAfterBaseline`; `grep -c 'MCP Server Version'` on a golden prints 1, so a registered check leaves
exactly one row.

### R-21 — Test-binary detection precedent (§3.4 of `design.md`; D5, D18)

Read, `internal/cli/glm.go:1058-1071`: `isTestEnvironment()` returns true when `config.EnvTestMode` is `"1"`, or when **any** element of
`os.Args` ends in `.test` or `.test.exe` or contains `go.test`. So `moai init my-app.test` would be classified as a test. Not executed.
`ptycap_child_test.go:105` calls `runInit(initCmd, nil)` inside a child test binary.

### R-22 — Catalog counts (P-16)

From `internal/template/catalog.yaml`: 37 skills (24 `core`, 13 `optional-pack:*`: backend 3, devops 5, frontend 4, design 1) and 12 agents (11
`core`, 1 `harness-generated`); `internal/template/templates/.claude/skills` holds exactly those 37 directories. "35 core" at iteration 0 is
24 skills + 11 agents. Commands are not catalog entries (17 files under `templates/.claude/commands/moai/`: 15 `.tmpl`, `gtd.md`, `todo.md`).

### R-23 — `moai init` writes the user-scope settings file, and which seams move it (P-42; N4, OD-14)

Read, not executed (executing it would write the real home):

```
internal/cli/init.go:881         if homeDir, homeErr := userHomeDirFn(); homeErr == nil {
internal/cli/init.go:890-892     applyAutonomyTierBundleFn( opts.ProjectRoot, filepath.Join(homeDir, ".claude", "settings.json"), projectSettingsPath, opts.AutonomyTier
internal/core/project/autonomy_bundle.go:71-83   semi-auto / unset -> toolpolicy.WriteUserDefaultMode(userSettingsPath, config.TierDefaultMode(effective))
                                 (comment, :46-50: "ONLY the USER-scope defaultMode="acceptEdits" record is written")
internal/cli/glm_tools.go:124    var userHomeDirFn = userHomeDir
internal/paths/paths.go:53-58    func Home(): if h := os.Getenv("HOME"); h != "" { return h, nil } ; return os.UserHomeDir()
internal/paths/paths.go:75-83    MoaiHome(): a non-empty absolute MOAI_HOME is returned verbatim, else Home() joined with ".moai"
```

So a run of the real binary's `init` writes `$HOME/.claude/settings.json` whatever `CLAUDE_CONFIG_DIR` says; only `HOME` moves that
file, and `MOAI_HOME` moves the `~/.moai` root only. `main_test.go:190-215` records that `go test` needs a package-wide
`userHomeDirFn` redirect for this reason. The same repository doctrine that refuses `HOME=` in a command also says a script file is not
a way round the guard (`.claude/rules/moai/workflow/kanban-dispatch-mechanics.md:105`, read), so an `init` case in an offline
harness has no isolated form: OD-14.

### R-24 — The Codex home seam and the registry-wide doctor tests (P-43; N2)

```
internal/cli/mcp_codex.go:2138   var codexUserHomeDir = os.UserHomeDir
internal/cli/mcp_codex.go:2143-2152   resolveCodexHomeDir(): a non-blank CODEX_HOME wins, else codexUserHomeDir() joined with ".codex"
$ grep -n 'CODEX_HOME\|codexHomeEnvVar\|codexUserHomeDir' internal/cli/main_test.go
(no output; the TestMain sandbox at main_test.go:269-299 redirects userHomeDirFn only)
$ grep -rn 'runDiagnosticChecks(false, "")' internal/cli --include='*_test.go'
internal/cli/doctor_hook_wiring_test.go:250:	for _, c := range runDiagnosticChecks(false, "") {
internal/cli/doctor_golden_test.go:263:	checks := runDiagnosticChecks(false, "")
internal/cli/doctor_test.go:280:	checks := runDiagnosticChecks(false, "")
internal/cli/doctor_test.go:422:	checks := runDiagnosticChecks(false, "")
internal/cli/doctor_new_test.go:320:	checks := runDiagnosticChecks(false, "")
internal/cli/doctor_new_test.go:329:	checks := runDiagnosticChecks(false, "")
internal/cli/doctor_flag_slot_test.go:181:	checks := runDiagnosticChecks(false, "")
$ grep -rln 'codexUserHomeDir = ' internal/cli --include='*_test.go'
(7 files: doctor_codex_test.go, codex_skills_path_shape_test.go, codex_skills_prune_test.go, codex_skills_disable_test.go,
 codex_skills_dotdot_test.go, codex_readiness_test.go, doctor_codex_stale_skill_test.go)
```

Seven unfiltered registry runs in six test files, not three tests: the audit named `TestRunDiagnosticChecks_All` and its two siblings
(`doctor_test.go:280`, `doctor_new_test.go:320,329`), and four more sites exist. Only `doctor_golden_test.go` pins the Codex PATH
seam (`:131-138`). The seven files that assign `codexUserHomeDir` per test capture the current value as `orig` and restore it, so a
TestMain-level redirect installed through the same variable is restored by each of them to the sandboxed value, not to the real one.
The existing doctor checks already resolve the real Codex home in those runs (reads); a registered check that starts `codex` is the new
element.

### R-25 — The Claude binary resolver order (P-44; N3)

Read, `internal/cli/claude_binary.go:33-48`:

```
if pin := os.Getenv(config.EnvClaudeBin); pin != "" { return validateClaudeBinaryPin(pin, "env var "+config.EnvClaudeBin) }
if root, err := findProjectRoot(); err == nil { … llm.ClaudeBin != "" → validateClaudeBinaryPin(llm.ClaudeBin, "llm.claude_bin in .moai/config/sections/llm.yaml") }
claudeBin, err := exec.LookPath("claude")
```

A `PATH` shim is the last of three sources, so an inherited `MOAI_CLAUDE_BIN`, or a project pin found from the working directory
upward, outranks it. The stand-in harness of R-28 reproduces both outcomes. This session's own environment carries 29 `MOAI_*`, `CLAUDE_*`
and `CODEX_*` names (`awk 'BEGIN{n=0;for(k in ENVIRON) if (k ~ /^(MOAI|CLAUDE|CODEX)_/) n++; print n}'` printed `29`; the 34 that an
early version of the stand-in printed was these 29 plus the five fixed poison names it plants) and no `MOAI_CLAUDE_BIN`, so the hazard is a supported pin, not an
observed one.

### R-26 — Payload file modes (P-45; N8)

```
internal/template/deployer.go:272-277   perm := fs.FileMode(0o644); if strings.HasSuffix(destRelPath, ".sh") { perm = 0o755 }
$ git ls-files -s internal/template/templates/.claude/skills/moai-workflow-project/scripts
100755 …	…/navigator-audit.sh
100755 …	…/navigator-enrich.sh
100644 …	…/navigator-regen.sh
internal/template/templates/.claude/skills/moai-workflow-project/references/navigator.md:88   bash ".claude/skills/moai-workflow-project/scripts/navigator-regen.sh"
```

A scratch module that embeds one executable and one plain file (`//go:embed all:t`, `x.sh` mode 755, `y.txt` mode 644) and prints each
file's mode from `fs.WalkDir`:

```
t/x.sh -r--r--r--
t/y.txt -r--r--r--
```

`embed.FS` reports 0444 for both, so a generator that reads the embedded template tree has no source mode to copy; the deployer's
suffix rule is the only mode rule derivable from it, and the plugin payload mirrors that rule. Whether a skill executes
`navigator-audit.sh` or `navigator-enrich.sh` directly was not observed (`navigator.md:88` runs the third through `bash`), so the harm of
losing the bit is unconfirmed; a marketplace plugin is taken from git, where the committed mode is the delivered mode (inferred).

The criterion's own check form, `find <tree> -name '*.sh' ! -perm 755 -print`, aimed at three trees:

```
the template source tree (a copier of source modes):  …/navigator-regen.sh                      (one path: 100644 in git)
stand-in payload, every file 0644 (design.md as it was written):  three .sh paths printed
stand-in payload, .sh files 0755:                                   no output
control, `find <good payload> -name '*.sh' -print`:                 three .sh paths printed
```

### R-27 — A name scan that passes correct code and fails hard-coded names (P-16; N1)

Stand-in for `TestGeneratorHoldsNoComponentNames` (an AST scan of the non-test Go sources of a directory; string literals only; each
literal split on `/` and `\`; a token is a hit when it equals a catalog skill or agent name or a command stem of the template tree;
the plugin identifier `moai` is exempt; a positive control that must hit a path-form and a lone-literal form before the sweep is
trusted; an empty sweep fails). Source and fixtures live in the scratchpad `i3/scan`, `i3/good`, `i3/bad1..3`; the tool is not part of
the repository.

```
good  (const pluginName = "moai"; subdirs "skills", "commands"; an error text "cannot sync %q: empty name in the plan"; a comment naming a skill and a command stem)
  CONTROL hits=2/2 (positive control: path-segment form and lone-literal form)
  SWEPT files=1 literals=9 names=65 (catalog=49 commands=17 exempt=1)
  PASS no component-name literal
bad1  (map[string]bool{"moai-foundation-core": true})
  HIT "moai-foundation-core" at …/i3/bad1/gen.go:3:28            SWEPT files=1 literals=1 names=65 …   FAIL 1 component-name literal(s)   [exit 1]
bad2  (const p = "skills/moai-foundation-core/SKILL.md")
  HIT "moai-foundation-core" at …/i3/bad2/gen.go:3:11            FAIL 1 component-name literal(s)   [exit 1]
bad3  (var only = []string{"gtd"})
  HIT "gtd" at …/i3/bad3/gen.go:3:21                              FAIL 1 component-name literal(s)   [exit 1]
real, correct by design:  internal/template/commandemit   SWEPT files=3 literals=81  …   PASS no component-name literal
                          internal/template/agentemit     SWEPT files=6 literals=267 …   PASS no component-name literal
real, hand-listed names:  internal/template               SWEPT files=27 literals=782 …  FAIL 62 component-name literal(s)   [exit 1]
  (first hits: glm_effort_overlay.go:184 "manager-develop"; profile_matrix.go:88 "todo", :98-:109 and :192-:233 the twelve agent names;
   retained_agents.go:11-22 twelve agent names)
```

`names=65` is 49 catalog entries (`- name:` lines of `catalog.yaml`: 37 skills, 12 agents) plus 17 command stems, less the one exempt name
`moai`, which is both the plugin identifier and the catalog's first skill. The real hits show the failure RK-4 exists to prevent in code
that exists today, so the scan is neither vacuous nor satisfiable only by an empty tree.

### R-28 — A stand-in harness whose scrub, pin handling and canary can each be seen to fail (P-42, P-44, P-46; N2 to N4)

Stand-in for the isolation part of `scripts/test-plugin-install-step.sh`: it poisons its own environment first (`MOAI_CLAUDE_BIN` to a
recording script, `MOAI_STUB_LEAK_DIR` to a canary directory that stands in for a real home, one `MOAI_*`, `CLAUDE_*` and `CODEX_*`
variable each, and one more of each family whose name carries the process id, so that no typed list can name it), scrubs by live
enumeration (`awk` over `ENVIRON`, names matching `^(MOAI|CLAUDE|CODEX)_`), sets `PATH` to `<shim>:/usr/bin:/bin`, scratch
`CLAUDE_CONFIG_DIR`, `CODEX_HOME` and working directory, runs a stub product that models the resolver order of R-25 and an out-of-scratch
directory creation, and then judges five cases. It sets no `HOME`, runs no git, and invokes no real `claude`, `codex` or `moai`; the
protected set is read-only hashed (`protected-set-hash.sh`, R-29). Scratchpad `i3/h`. A fourth mode replaces the scrub with a typed list of
the five fixed poison names, as a mutant.

```
$ sh i3/h/harness.sh normal
scrub: enumerated and unset 37 names
PASS isolation-env-scrubbed
PASS isolation-cwd-has-no-project
PASS isolation-resolves-to-stubs
PASS isolation-poisoned-pin-never-executed
PASS isolation-real-home-unchanged
canary before: PROTECTED-SET c00bacbd…f46b68 entries=3
canary after : PROTECTED-SET c00bacbd…f46b68 entries=3
LEAK=0 (real roots unchanged: PROTECTED-SET a245f41a…77cf5f entries=193)
poison.log lines: 0   stub.log lines: 2
RESULT pass=5 fail=0

$ sh i3/h/harness.sh typed-list-mutant      (the scrub is `unset` of the five fixed poison names)
scrub: typed list of 5 names (mutant)
RED   isolation-env-scrubbed
green isolation-cwd-has-no-project
green isolation-resolves-to-stubs
green isolation-poisoned-pin-never-executed
green isolation-real-home-unchanged
LEAK=0 (real roots unchanged: PROTECTED-SET a245f41a…77cf5f entries=193)
poison.log lines: 0   stub.log lines: 2
RESULT typed-list-mutant: red set and green set are exactly the expected ones

$ sh i3/h/harness.sh negative-control        (scrub disabled, working directory inside a project whose llm.yaml pins the recorder)
scrub: DISABLED (negative control)
RED   isolation-env-scrubbed
RED   isolation-cwd-has-no-project
green isolation-resolves-to-stubs
RED   isolation-poisoned-pin-never-executed
RED   isolation-real-home-unchanged
canary before: PROTECTED-SET 20bad580…ac97f99 entries=3
canary after : PROTECTED-SET b261a533…cd2ceb entries=5
LEAK=0 (real roots unchanged: PROTECTED-SET a245f41a…77cf5f entries=193)
poison.log lines: 2   stub.log lines: 0
RESULT negative-control: red set and green set are exactly the expected ones

$ sh i3/h/harness.sh negative-control-cwd    (scrub working, working directory inside the pinning project)
scrub: enumerated and unset 37 names
green isolation-env-scrubbed
RED   isolation-cwd-has-no-project
green isolation-resolves-to-stubs
RED   isolation-poisoned-pin-never-executed
green isolation-real-home-unchanged
canary before: PROTECTED-SET 0a56754d…9908b5 entries=3
canary after : PROTECTED-SET 0a56754d…9908b5 entries=3
LEAK=0 (real roots unchanged: PROTECTED-SET a245f41a…77cf5f entries=193)
poison.log lines: 2   stub.log lines: 0
RESULT negative-control-cwd: red set and green set are exactly the expected ones
```

Reading. A PATH shim alone does not stop a pin: with the scrub disabled the recorder ran twice and the shim not at all (`poison.log
lines: 2   stub.log lines: 0`), and the project pin alone does the same with the scrub working. The case that does not depend on the scrub
(`isolation-resolves-to-stubs`) stays green in every negative mode, so a negative mode does not simply fail everything. The typed-list
mutant passes everything except `isolation-env-scrubbed`, which is why the harness plants names no list can contain: the three
pid-named variables are what no typed list can name: the mutant leaves 32 names behind (the 29 session names and the three pid-named
ones) and that remainder is its `RED`, while the working scrub removes 37 (those 32 plus the five fixed poison names). The canary went
from three entries to five, which are two
empty directories (`home/.claude` and `home/.claude/leaked-dir`).

The directory-entry control, in the scratchpad `i3/ctl`: a canary directory `canary/plugins` hashed before and after creating the empty
directories `plugins/data/p-mcp-inline` (the shape of the t1434 leak):

```
$ sh i3/h/protected-set-hash.sh i3/ctl/canary                 -> PROTECTED-SET 0596b724…324e8 entries=2     (before)
$ find i3/ctl/canary -type f -exec shasum -a 256 {} +         -> (no output)                                  (before)
$ mkdir -p i3/ctl/canary/plugins/data/p-mcp-inline
$ sh i3/h/protected-set-hash.sh i3/ctl/canary                 -> PROTECTED-SET dcf79825…37953 entries=4     (after: different)
$ find i3/ctl/canary -type f -exec shasum -a 256 {} +         -> (no output)                                  (after: unchanged)
```

A files-only manifest cannot tell the two states apart; the entry hash can.

### R-29 — The real roots are noisy on a machine with live sessions (P-46; N4)

`protected-set-hash.sh` with its first root list (the Claude `plugins` tree whole, `<codex>/tmp`, `<codex>/.tmp`) over this machine's real
roots, read only. The before and after hashes of one harness run, a few seconds apart, differed (`entries=151297`, then `entries=151301`);
two later runs agreed with each other (151301), and a third a few minutes later differed again (151310). The diff of the 151301-entry list
and the 151310-entry list:

```
> /Users/goos/.codex/.tmp/git-Ujcu0E
> /Users/goos/.codex/.tmp/git-Ujcu0E/HEAD
> /Users/goos/.codex/.tmp/git-Ujcu0E/objects
> /Users/goos/.codex/.tmp/git-Ujcu0E/refs
> /Users/goos/.codex/tmp/arg0/codex-arg0dKiaGu
> /Users/goos/.codex/tmp/arg0/codex-arg0dKiaGu/.lock
> /Users/goos/.codex/tmp/arg0/codex-arg0dKiaGu/apply_patch
> /Users/goos/.codex/tmp/arg0/codex-arg0dKiaGu/applypatch
> /Users/goos/.codex/tmp/arg0/codex-arg0dKiaGu/codex-execve-wrapper
```

The writer was not identified. No command of this session started `codex` (the stand-in used stubs), and the entry shapes are those the
Codex CLI itself creates (P-34 shows it creating `tmp/arg0`), so another Codex process on the machine is the inferred writer. The entry
list also holds about twenty `…/.tmp/marketplaces/.staging/marketplace-upgrade-*` directories (background upgrade staging). The t1434
verdict classed `plugins/synced` of the Claude profile as AMBIENT for the same reason (`.moai/reports/t1435/inputs/t1434-verdict.md`
line 48).

Second list: the `plugins` trees whole and `<codex>/.tmp/marketplaces` whole, with `plugins/synced` and `.tmp/marketplaces/.staging`
excluded and `<codex>/tmp` and `<codex>/.tmp/git-*` no longer roots. Harness runs at 07:30 to 07:31 read `entries=41337`; a run at 07:42 read
`entries=41338`. The diff of their before-lists is one file, inside an existing official marketplace clone of the real Claude profile:

```
41205a41206
> /Users/goos/.moai/claude-profiles/moai-adk/plugins/marketplaces/claude-plugins-official/plugins/security-guidance/tests/test_review_model.py
```

No command of this session addressed that profile, so the runtime refreshing its marketplace clone is the inferred writer. The same
tree holds `plugins/.trash/<epoch>-<pid>-<id>/<plugin>~g<n>` directories and a sibling `plugins/.last_inuse_sweep`, which reads as the
runtime's own sweep of orphaned plugin versions (inferred, not observed).

Final design, from those observations: depth-limited roots, because what a leak of the t1434 kind adds sits at depth 2 or 3 and what the
runtime churns sits deeper — `<claude>/plugins` to depth 3, `<codex>/plugins` to depth 3, `<codex>/.tmp/marketplaces` to depth 2,
`~/.moai` to depth 1 — plus the file content of `settings.json` (both homes), `installed_plugins.json` and `config.toml`; and three
exclusions by declaration, `<claude>/plugins/synced`, `<claude>/plugins/.trash` and `<codex>/.tmp/marketplaces/.staging`. Over the real
roots that set reads `entries=193`, and nine reads of it (one direct, then two per run of four harness runs) printed the same hash
`a245f41a…77cf5f`.

Limits, stated: a `tmp/arg0` directory created by a started `codex` is invisible to the real-root hash on a machine where Codex runs, so
that leak path is closed by the harness resolving only to stubs and by the test-binary refusal (REQ-017), not by the hash; a new file
deeper than the declared depth inside an existing marketplace clone or plugin cache is invisible; the content of
`known_marketplaces.json` is not hashed (the runtime rewrites it), only its name. A non-zero `LEAK` prints the differing entries, so an
ambient change is attributed to its writer rather than assumed.

## 4. Sources

- `claude` 2.1.287 and `codex` 0.160.0 command output (above); `claude plugin --help` family.
- `https://code.claude.com/docs/en/sub-agents` (R-18).
- The t1434 verdict table: `.moai/specs/SPEC-PLUGIN-LOAD-SCOPE-001/progress.md §E.2` (rows R01 to R14).
- Repository files named in each row.
- `.moai/reports/t1435/plan-audit-iter1.md` (iteration 1 evidence, cited where this revision relied on it rather than re-ran it) and
  `.moai/reports/t1435/plan-audit-iter2.md` (iteration 2: the defects N1 to N13 behind R-23 to R-29).

## 5. What this research did not observe

`claude plugin marketplace add modu-ai/moai-adk` and the Codex equivalent against GitHub; `--scope project|local` behavior;
shadowing, collision or hook double-firing between the scaffold copy and the plugin; whether the runtime registers a nested
`commands/<dir>/` file for invocation; the `moai:<name>` invocation form for this plugin; behavior at any tool version other than the two
above; PowerShell and Windows behavior; the scaffold's own listing cost; whether `codex plugin marketplace upgrade` refreshes a Git
marketplace entry. From the iteration-3 revision: the real `moai` binary was not run under the harness (the stand-in models only the
resolver order and one out-of-scratch write, R-28), so whether the real binary writes anything under the real `~/.moai` or home with the
scrub applied is for the first run of `scripts/test-plugin-install-step.sh`; a real `moai init` was never run (R-23); the real-root hash was
read, never written; no command of this revision wrote to a real profile or home (its only reads of them are the entry hash and the
entry lists dumped from it, R-29).
