# Design — SPEC-PLUGIN-MARKETPLACE-001

Architecture and control flow for the marketplace, the derived `moai` plugin, the install step, the doctor check and the
release coupling. The requirements are in `spec.md` §2, the build order in `plan.md` §3, the observations behind every
premise in `research.md`. Names below are proposals the run phase may change; the observable behavior they describe is
fixed by the requirements. A clause marked `default pending OD-n` follows the Open Decision of that number.

## 1. Architecture overview

```
internal/template/templates/          (SSOT, embedded in the binary)
  .claude/skills/**   .claude/commands/moai/*   .mcp.json        pkg/version (fallback Version)
        │                    │                     │                     │
        └───────── pluginemit (generator; reads the embedded FS through catalog + SlimFS) ──────┐
                                                                                                 ▼
 committed, drift-checked output:  .claude-plugin/marketplace.json      .agents/plugins/marketplace.json
                                   plugins/moai/.claude-plugin/plugin.json   plugins/moai/.codex-plugin/plugin.json
                                   plugins/moai/skills/**   plugins/moai/commands/<name>.md   plugins/moai/.mcp.json
 make plugin-emit  (writes)        make plugin-emit-check  (read-only; ahead of `build`)

 moai init ──► install step ──► claude / codex CLIs ──► <config home>/plugins/…        (REQ-010 to REQ-017)
 moai plugin install ─┘             ▲
 install.sh / .ps1 / .bat ──────────┘  (call the installed binary's verb by its installed path)   (REQ-018, REQ-019)

 moai doctor ──► "Plugin Version" ──► <Claude home>/plugins/installed_plugins.json   (file read)
                                  └─► codex plugin list --json  (bounded)           (REQ-020 to REQ-023)

 release.yml verify-provenance, check 8 ──► scripts/check-plugin-version.sh <tag>     (REQ-024)
```

Two properties organise everything else. **The payload has one source** (the template tree) so a rename there changes the
payload and nothing else needs an edit; the drift gate turns a forgotten regeneration into a build failure. **The install
step touches a person's real profile**, so every path that can reach it is either opt-out-able, harness-gated, or inert
(tests), and every failure is non-fatal.

## 2. Payload derivation

### 2.1 Inputs and the tier filter

- The embedded template file system (`//go:embed all:templates`), never the working tree's own `.claude/` copy: a
  maintainer-local edit cannot reach the payload.
- The catalog (`internal/template/catalog.yaml`) through `LoadEmbeddedCatalog`, and the init-time view `SlimFS(rawFS, cat)`
  (`internal/template/slim_fs.go:214`), which hides every non-core catalog entry. At the OD-8 default the payload is exactly
  what a default `moai init` deploys: 24 core skills (the catalog also holds 13 optional-pack skills and 12 agents, 11 of them
  core, none shipped at the OD-3 default). Commands are not catalog entries, so the filter passes them through (P-16).
- The version SSOT: `pkg/version.Version`, the no-ldflags fallback that each bump commit rewrites (`pkg/version/version.go:11`).
- The MCP source: `.mcp.json` of the template tree.

### 2.2 Per-kind rules

| Kind | Source (template tree) | Destination | Rule |
|------|------------------------|-------------|------|
| Skill | `.claude/skills/<dir>/**`, core tier | `plugins/moai/skills/<dir>/**` | every file copied byte for byte; nothing rendered, including the 16 files that contain brace text (REQ-005) |
| Command | `.claude/commands/moai/<stem>.md` or `<stem>.md.tmpl` | `plugins/moai/commands/<stem>.md` | non-`.tmpl` copied; `.tmpl` rendered with the default context and the suffix dropped; **flat**, no subdirectory (REQ-006) |
| MCP entry | `.mcp.json` `mcpServers.moai` | `plugins/moai/.mcp.json` | the one entry copied with its `command` and `args`; the template's second server (`context7`) and the `staggeredStartup` block are not carried (REQ-007, default pending OD-2) |
| Agents | `.claude/agents/moai/*.md` | `plugins/moai/agents/*.md` | only where OD-3 admits them; none at the default |
| Everything else | rules, `CLAUDE.md`, `AGENTS.md`, settings, `.moai`, output styles, workflows, mods, hooks | none | not emitted; an allow-list, so a new scaffold-only directory cannot leak (REQ-008) |

The default `TemplateContext` carries `ConversationLanguage: "en"` (`internal/template/context.go:88`), so every
`{{if eq .ConversationLanguage "<xx>"}}…{{else}}…{{end}}` chain resolves to its `else` branch, the English text. The only
actions in the 15 `.tmpl` command sources are those chains, on `description:` and `argument-hint:` lines (P-32); the
`commandemit` package already publishes the same English variant. Under OD-11 option (b) the generator renders once per
locale; under option (c) it renders no commands.

Why flat. The template sources sit one level down (`.claude/commands/moai/<name>.md`), which is what makes the project
copies list as `moai:<name>`. Copying that shape into a plugin installs commands the runtime inventory does not count: the
same 17 files nested at `commands/moai/` printed `Skills (24)` and listed none; flat they printed `Skills (41)` (P-30). The
plugin's own name supplies the namespace, so no directory is needed for it.

### 2.3 Manifests

| File | Fields (all copied from a precedent or required by strict validation) |
|------|------------------------------------------------------------------------|
| `.claude-plugin/marketplace.json` | `name` `moai-adk`; `owner.name`; `metadata.description`, `metadata.version`; `plugins[0]`: `name` `moai`, `source`, `version`, `description`, `category` |
| `.agents/plugins/marketplace.json` | `name` `moai-adk`; `interface.displayName`; `plugins[0]`: `name`, `source` `{source: local, path}`, `policy` `{installation: AVAILABLE, authentication: ON_INSTALL}`, `category`; **no `version` key** (P-35) |
| `plugins/moai/.claude-plugin/plugin.json` | `name` `moai`, `version`, `description`, `author`, `homepage`, `repository`, `license` |
| `plugins/moai/.codex-plugin/plugin.json` | the same, plus `skills: "./skills/"`, `mcpServers.moai` (default pending OD-2) and an `interface` block |

Version carriers: four fields hold the version — `metadata.version` and `plugins[0].version` of the Claude marketplace, and
`version` of each plugin manifest. The Codex marketplace holds none, so the Codex plugin version is the Codex plugin
manifest's, which is also what `codex plugin list --json` reports and what names the cache directory (P-34, P-35). Strict
validation compares only the entry version with `plugin.json` (P-03); `metadata.version` equality is pinned by AC-001 (c).

### 2.4 Determinism and the drift check

The generator emits sorted paths, LF newlines, JSON with two-space indentation and a trailing newline, mode 0644, and no
timestamp, so two runs over one tree are byte-identical. `make plugin-emit` writes; `make plugin-emit-check` runs the
golden comparison read-only (byte difference, missing file, extra file) with `PLUGIN_EMIT_UPDATE` scrubbed, and is a
prerequisite of `build` like `agents-emit-check`. The check never regenerates (AC-009 `committed-set-unchanged`). The
version enters the same comparison, so a bump without `make plugin-emit` fails the build.

## 3. Install step

One Go function, called by `moai init` and by the `moai plugin install` verb, with an injected command runner.

### 3.1 Tool selection

| Caller | Tools acted on |
|--------|----------------|
| `moai init`, harness `claude` (or no `--llm`) | Claude (default pending OD-9) |
| `moai init`, harness `gpt` | Codex (default pending OD-9) |
| `moai init`, harness `both` | Claude, then Codex |
| `moai plugin install`, the install scripts | every tool found on PATH (no harness exists) |

The rule is the one `wireCodexUnlessClaude` (`init.go:192-195`) and the `mcpDeclined` switch (`:997-1003`) already apply to
the other tool-specific surfaces: the `--llm` flag's own help says `gpt` deploys "AGENTS.md + Codex surfaces only — no
.claude/ tree" and a default init deploys the Claude tree (`init.go:133`). Without a gate, a default init would write the Codex
profile whenever `codex` is installed, and `--llm gpt` would write the Claude profile (OD-9 option (b)).

### 3.2 Sequence and outcomes

For each selected tool, in order Claude then Codex:

1. Opt-out set → return; nothing runs (REQ-015).
2. The command runner is the default one and the process is a test binary → refuse; nothing runs (REQ-017, §3.4).
3. Resolve the binary. Claude goes through the pinned-binary resolution (`MOAI_CLAUDE_BIN`, then `llm.claude_bin`, then PATH);
   Codex through `exec.LookPath("codex")` (`codexWiringLookPath`, `doctor_codex.go:39`).
   - Not found → one skip line, next tool (REQ-014).
   - Pin invalid → one guidance block, next tool (REQ-013).
4. Print the config home the tool will resolve (REQ-012).
5. Run `marketplace add modu-ai/moai-adk` under a deadline of `config.DefaultPluginInstallCommandTimeout` (60 seconds).
   Exit 0 → continue. Non-zero or deadline → one guidance block, no install, next tool (REQ-013).
6. Run `plugin install moai@moai-adk` (Codex `plugin add moai@moai-adk`) under the same deadline. Exit 0 → done. Non-zero or
   deadline → one guidance block (REQ-013).
7. The function returns nil in every case.

Exit 0 is success whatever the message: both tools exit 0 on a repeat run (P-05, P-06, P-10), so idempotence needs no
message parsing, and the `already-present-is-success` subtest pins it. No `--scope` and no `--yes` are passed (REQ-011): the
tool default scope applies (P-33), and `--yes` is for command-source installs, which a local or GitHub relative-path
marketplace entry is not.

The resolver change. `resolveLaunchClaudeBinary()` (`claude_binary.go:33`) returns plain errors and reads the pin through
`findProjectRoot()`, the working directory's project. The step needs the two failure classes apart (REQ-013 and REQ-014 print
different things) and should read the pin of the project being initialised, not of whatever project the working directory
sits in. The run phase therefore adds a typed error for the not-found class and a variant that takes the project root; the
launch path keeps calling the unchanged entry point. This is read from the source, not executed.

Worst case. Four commands each held to 60 seconds is 240 seconds (RK-15). It is reached only when every command hangs, and the
opt-out removes it. A step-level deadline would shorten it but adds a second constant; the per-command bound is what REQ-013
states.

### 3.3 Environment, profile and scope

The child inherits the process environment unchanged (REQ-012, default pending OD-10), so the profile it acts on is the one
`CLAUDE_CONFIG_DIR` names, else `~/.claude`, and for Codex `CODEX_HOME`, else `~/.codex`. The step prints that home before its
first command so the person sees which profile moves. The launcher is different: `moai cc` resolves a project-recorded profile
from a launch ledger (`launcher.go:159`, `profile.go:461-495`), which init does not; a project has an entry only after a prior
launch, so on a first init the recorded-profile option (OD-10 (b)) degenerates to the inherited environment. RK-17 records the
mismatch that remains.

Scope is the tool default, user scope (REQ-011, default pending OD-12). The scripts and the verb have no project, so user scope
is the only form they can use; project and local scope would also write into the project's `.claude/settings*.json`, behavior
no run has observed (G-5).

### 3.4 Test-binary inertness

REQ-017 asks that no `go test` run reach a real profile whichever test calls `runInit` or `initCmd.RunE`. P-20 counts 32
`runInit(` call sites in 18 test files and 9 `initCmd.RunE(` sites in 5; editing each is neither planned nor needed. The default
runner is one package-level seam that refuses while the process is a test binary, so a test that wants the step to run injects a
runner and every other test is inert by construction.

Detector. `isTestEnvironment()` (`internal/cli/glm.go:1058`) already exists, but it returns true when **any** argument ends in
`.test` or contains `go.test`, so `moai init my-app.test` would read as a test environment in production and silently skip the
step (read, not executed). The detector must key on the program name (`os.Args[0]`'s base ending `.test` or `.test.exe`) or on
`testing.Testing()` (Go 1.21 and later; `go.mod` declares go 1.26.8). AC-017's second subtest pins that a pinned
`MOAI_CLAUDE_BIN` and PATH shims both stay untouched; a PATH shim alone would miss the pin.

### 3.5 Opt-out and automated callers

The opt-out (default pending OD-5) is the `--no-plugin` flag of `moai init` and the environment variable
`MOAI_SKIP_PLUGIN_INSTALL` set to `1` or `true`; an empty value or `0` does not opt out. The variable name is a constant in
`internal/config/envkeys.go`. Automated callers set it (REQ-016): the Agent Emit Embed doctor check, which runs
`init --non-interactive --llm both` as a child (`doctor_agentemit_embed.go:352`) and would otherwise act on both real profiles,
and the two `moai init` lines of `e2e/cli/tux3_journeys.sh` (lines 104 and 115), which run from `/tmp/moai-e2e` with the
caller's environment and which no workflow references. `TestPluginOptOutCallersEnumerated` keeps the list honest: a new
`exec.Command(<moai>, "init", …)` fails it until the caller is listed and sets the variable.

### 3.6 The verb and the installers

`moai plugin install` (OD-6 option (a); the name is a proposal) runs the step for every tool found, prints on stderr, honors
the opt-out, and exits 0 on every outcome REQ-013 to REQ-015 name; only a usage error exits non-zero (REQ-019).

Each installer calls the installed binary's verb after the binary lands:

- `install.sh`: `TARGET_PATH` is set in `install_binary` (`install.sh:274`) as a global, so it is visible in `main`; the call is
  guarded so that `set -e` (line 5) cannot abort the script after the binary was installed.
- `install.ps1`: the install path is a variable of `Verify-Installation`'s caller; the call sits in a `try` with a non-rethrowing
  `catch`, because `$ErrorActionPreference = "Stop"` (line 6) turns a thrown error into an abort.
- `install.bat`: `"%TARGET_PATH%" plugin install` (`install.bat:145`), with no `errorlevel` test after it; the script ends
  `exit /b 0` (line 192).

The install directory need not be on PATH: the base `install.sh` driven offline printed `Installation completed, but 'moai'
command not found in PATH` (P-37), so a bare `moai` would fail exactly where it is most needed. A script served from `main` may
meet a release that lacks the verb; the call fails with `Unknown command "plugin" for "moai".` (P-39) and the installer must still
finish (RK-14). `docs-site/static/install.sh` and `install.ps1` stay byte copies of the roots; the `install-script-parity` job
already diffs them.

Output wording (a proposal; the tests assert counts and the manual commands, not the prose):

```
moai plugin: Claude Code config home: <path>
note: could not install the moai plugin for Claude Code (<reason>). Install it yourself:
        claude plugin marketplace add modu-ai/moai-adk
        claude plugin install moai@moai-adk
note: claude not found on PATH; skipping the moai plugin install
```

## 4. Doctor check "Plugin Version"

```
doctor ──► checkPluginVersion(homes, probe, binaryVersion)
            │  Claude:  <CLAUDE_CONFIG_DIR | ~/.claude>/plugins/installed_plugins.json
            │             → plugins["moai@moai-adk"] → entries → user-scope first → version     (no subprocess)
            │  Codex:   codex on PATH?  no → skip
            │             yes → `codex plugin list --json` under the resolved CODEX_HOME, deadline 3 s
            │                   → installed[] where pluginId == "moai@moai-adk" → version
            ▼
          strip a leading "v" from both sides → equal: OK | unequal: OD-7 severity (default warn), both versions + remedy
          indeterminate (not installed, home absent, malformed or unknown shape, probe timeout, dev build): OK or info
```

- The Claude read is one file read whatever the project size, which is what `coding-standards.md` § Advisory-Check Discipline
  asks of anything near a latency-sensitive path. `moai doctor` is an on-demand command, not a session-start or Stop path.
- The Codex read asks the CLI because installed state lives in `config.toml` registration, not in the cache: after the
  registration was removed the list printed `"installed": []` while `plugins/cache/moai-adk/moai/3.1.3` remained (P-34). Reading
  the cache directory would report a plugin that is not installed. The cost is a process start: 0.02 s measured, bounded by
  `config.DefaultPluginVersionProbeTimeout` (3 seconds), and a `tmp/arg0` directory the CLI created in an empty Codex home
  (P-34, RK-16). OD-13 option (b) reads the `[plugins."moai@moai-adk"]` stanza in `config.toml` and the cache directory for the
  version instead, starting no process.
- The `v` strip matters: the binary prints `v3.2.0-rc.26` and the manifests carry `3.2.0-rc.26`. A development build
  (`IsDevBuild`) is never compared.
- Remedy text per tool: Claude `claude plugin update moai@moai-adk`; Codex `codex plugin marketplace upgrade moai-adk` then
  `codex plugin add moai@moai-adk`. Both were read from `--help` and not run (G-1).
- The name `Plugin Version` sits beside the legacy `Plugin Deployment` check (`doctor.go:1120`, a `system.yaml` marker). The
  `--verbose` detail line says which one this is.
- Registration surfaces: the check-name allowlist `namesAddedAfterBaseline` (a bare identifier, since the name is registered
  through a constant), the status-set test if the check can emit info, and the three goldens, each of which then holds one
  `Plugin Version` row. `doctor --check <unknown name>` runs nothing and prints `Pass 0 Warn 0 Fail 0` (P-38), so only the
  name in the output, or the golden row, shows registration.
- `captureDoctorCmd` pins `HOME`, `MOAI_HOME`, `CLAUDE_CONFIG_DIR` and the Codex PATH seam. `HOME` already moves the default
  `~/.codex`, so what remains exposed is a `CODEX_HOME` set in the caller's environment, as in a Codex lane; the golden harness
  therefore sets `CODEX_HOME` to empty.

## 5. Release coupling

- Bump. Rewrite the seven Version Stamps (`version-management.md`), then `make plugin-emit`. The plugin files carry the version
  without a `v`, so the literal-token registry sweep does not see them; the drift gate is their guard, and a new "Generated
  version carriers" group in `version-management.md` puts the step where the person bumping looks. Skipping it fails
  `plugin-emit-check` at `make build` and the golden test in `go test`.
- Tag. `release.yml` job `verify-provenance` gains check 8: `scripts/check-plugin-version.sh <tag>` compares the tagged tree's
  `plugins/moai/.claude-plugin/plugin.json` version with the tag minus its `v`. The script runs offline, so the criterion does
  too.
- Pre-tag. `scripts/check-plugin-discoverable.sh <empty-home>` (AC-006) is a runbook step. It needs a Claude CLI, so it cannot
  run in CI; its continued-firing answer is the runbook and the plan-auditor's re-run at each release PR (G-7).
- First release. `main` carries no manifest until the release PR that contains them (P-41), so `claude plugin marketplace add
  modu-ai/moai-adk` fails for any build that carries the step until then. The step's guidance path makes that failure
  harmless, and the manifests and the step ship in one tag (RK-9).

## 6. Harness scripts

Both scripts exist because a Go test cannot answer the question. They set `HOME`, `PATH` and both config homes themselves,
because the worktree guard refuses a command that sets `HOME` and any `pwsh` (P-36), and a script file is a plain command.

`scripts/test-plugin-install-step.sh <moai>`: builds a scratch root; a scrubbed `PATH` (`<shim>:/usr/bin:/bin`) and the case
`harness-precondition` that fails if `claude` or `codex` resolves beyond the stubs; stub `claude` and `codex` that append their
argv to a log; a stub `curl` that serves a pinned archive for the archive URL and fails for `checksums.txt`; the archive holds the
`moai` binary under test, or a stub for the failure cases. Seventeen cases print `PASS <name>` or `FAIL <name>: <reason>`:
`harness-precondition`; `init-claude-harness`, `init-gpt-harness`, `init-both-harness`, `init-config-home-printed`,
`init-add-fails-skips-install`, `init-tool-fails-exit-0`, `init-no-tools-one-skip-line`, `init-flag-optout-zero-calls`,
`init-env-optout-zero-calls`; `verb-install-all-tools`, `verb-no-tools-exit-0`, `verb-optout-zero-calls`;
`installer-calls-verb-by-installed-path`, `installer-optout`, `installer-set-e-guard`, `installer-old-binary-unknown-verb`. A
case counts only logged calls whose first argument is `plugin`, so a tool invoked for another reason does not pollute it. The last
line is `RESULT pass=<n> fail=<m>`; the exit status is non-zero on any `FAIL`. It runs in CI (the Unix job of `test-install.yml`).

`scripts/check-plugin-discoverable.sh <empty-claude-home>`: refuses any argument that is not an existing empty directory (exit 2);
with `CLAUDE_CONFIG_DIR` set to it, adds the repository root as a marketplace, installs `moai@moai-adk` and reads
`claude plugin details`; the expected names are the directory names under `plugins/moai/skills` plus the command stems of the
template tree, so the payload cannot define its own expectation, and the `MCP servers` line must match the keys of
`plugins/moai/.mcp.json`. `claude plugin details` has no JSON form, so the script parses the `Skills (N)  a, b, …` and
`MCP servers (K)  …` lines and fails on a shape it cannot parse (RK-18).

## 7. Failure modes and observability

| Failure | What the person sees | Where it is recorded |
|---------|----------------------|----------------------|
| `marketplace add` fails (no network, manifest not yet on `main`) | one guidance block naming the manual commands | stderr of init |
| tool absent | one skip line | stderr of init |
| a command hangs | the block after 60 seconds | stderr of init |
| pin invalid | the block, naming the pin source in the resolver's message | stderr of init |
| registry or probe unreadable | doctor reports OK or info | doctor output |
| payload drifts from the template | `make build` fails before compiling | `plugin-emit-check` output |
| plugin version differs from the binary | doctor warns (default) with both versions and the remedy | doctor output |
| plugin version differs from the tag | the release workflow fails | `verify-provenance` job |

## 8. Alternatives and where each Open Decision lands

| OD | Where it changes this design |
|----|------------------------------|
| OD-1 | §3.1 and §3.2: the call to the step moves behind a flag |
| OD-2, OD-3, OD-8, OD-11 | §2.2: the payload set |
| OD-4 | §2.3 and §5: paths and the release check |
| OD-5 | §3.5 |
| OD-6 | §3.6: inline commands instead of the verb, no skew, three implementations |
| OD-7 | §4: the severity word |
| OD-9 | §3.1 |
| OD-10 | §3.3 |
| OD-12 | §3.2 steps 5 and 6, §4 Claude read |
| OD-13 | §4 Codex read |

Rejected without needing a decision: copying the `.claude/` tree as the payload (hides which parts are derivable and leaks
scaffold-only directories); a tag-pinned `sha` inside the repository (a commit cannot contain its own hash); a doctor check that
runs `claude` (`claude plugin list --json` took 0.26 s, and the registry file answers the same question with one read).

## 9. Not designed here

Init shrink (t1438), the mods UI (t1437), Aside (t1439), rules and instruction files carried by a plugin (unmeasured), plugin
refresh by `moai update`, and any behavior of two coexisting copies in a live session (G-2).
