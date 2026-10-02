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
| Skill | `.claude/skills/<dir>/**`, core tier | `plugins/moai/skills/<dir>/**` | every file copied byte for byte; nothing rendered, including the 16 files that contain brace text (REQ-005); `.sh` files written `0755`, every other file `0644` (§2.4) |
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

The generator emits sorted paths, LF newlines, JSON with two-space indentation and a trailing newline, and no timestamp, so two
runs over one tree are byte-identical. **Modes** follow the deployer's rule (`internal/template/deployer.go:275-276`): `0755` for a
`.sh` file, `0644` for every other file. The embedded template tree cannot supply a mode to copy — `embed.FS` reports 0444 for an
executable and a plain file alike (R-26) — so the suffix is the only rule derivable from it, and it is the rule `moai init` applies to
the scaffold copy the plugin copy will replace at t1438. The live case is `moai-workflow-project/scripts/`: `navigator-audit.sh` and
`navigator-enrich.sh` are `100755` in git, `navigator-regen.sh` is `100644`, and the deployer writes all three `0755`. A marketplace
plugin is delivered from git, where the committed mode is the delivered mode, so writing every file `0644` (this section's text
before plan-audit iteration 2) would ship scripts a skill might execute directly without the bit; whether any skill does was not
observed (`navigator.md:88` runs the third through `bash`), so the harm is unconfirmed and the rule is the cheap way to keep the two
copies equal. `make plugin-emit` writes; `make plugin-emit-check` runs the golden comparison read-only (byte difference, **mode
difference**, missing file, extra file) with `PLUGIN_EMIT_UPDATE` scrubbed, and is a prerequisite of `build` like
`agents-emit-check`. The check never regenerates (AC-009 `committed-set-unchanged`). The version enters the same comparison, so a bump
without `make plugin-emit` fails the build.

**Name independence.** The generator selects components from data only. The one literal it may hold that equals a template
component name is the plugin identifier `moai`, which it must write (plugin name, plugin path, MCP key) and which is also the
catalog's first skill. `TestGeneratorHoldsNoComponentNames` scans the string literals of its non-test sources, split on `/` and `\`,
whole token and case-sensitive, against the catalog names and command stems less `moai`; ordinary words that are command stems
(`run`, `sync`, `plan`) are safe because identifiers, comments and sentence-shaped literals are not tokens. The scan carries its own
positive control and fails on an empty sweep (AC-004 (b), R-27).

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

REQ-017 asks that no `go test` run reach a real profile whichever test calls `runInit`, `initCmd.RunE` or the doctor registry. P-20
counts 32 `runInit(` call sites in 18 test files and 9 `initCmd.RunE(` sites in 5, and P-43 counts seven unfiltered
`runDiagnosticChecks(false, "")` sites in six; editing each is neither planned nor needed. The default runner is one package-level
seam that refuses while the process is a test binary, so a test that wants the step or the probe to run injects a runner and every
other test is inert by construction. **One seam serves both callers**: the install step and the doctor's Codex probe start their
commands through the same variable, so there is one refusal to keep correct and one place a test double replaces. It guards the
process start; the read of the Codex home is guarded separately, by `TestMain` redirecting the `codexUserHomeDir` seam (§4).

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
- **Home and runner pins.** The check resolves the Codex home through `resolveCodexHomeDir()` (the `codexUserHomeDir` seam,
  `mcp_codex.go:2138-2152`), hands it to the child as `CODEX_HOME` in an otherwise unchanged copy of the parent's environment, and
  starts the child through the refusing runner of §3.4. Two independent guards keep a test off the real Codex: the runner refusal
  (no process starts) and `TestMain` wrapping `codexUserHomeDir` with the existing `homeRedirectingFn` and clearing `CODEX_HOME`
  (no read of the real home; `main_test.go:269-299` redirects `userHomeDirFn` only today, P-43). A canary directory standing in for
  a real home, hashed before and after a registry-wide run, and recording `codex` and `claude` shims first on PATH pin both (with
  `CLAUDE_CODE_VERSION` pinned, because the existing `Claude Code` check execs `claude --version` itself when it is unset, P-48); a check that
  resolved the home with `os.UserHomeDir()` itself, or a runner that started a process, turns `TestCheckPluginVersion_HomeIsolation`
  red (AC-021 (c)).
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
  therefore sets `CODEX_HOME` to empty. (That pin covers the golden test only; the other five files with an unfiltered registry run
  are covered by the two guards above.)

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

The scripts exist because a Go test cannot answer the question. **No script sets `HOME`.** The iteration-1 text did, to get past
the worktree guard's refusal of a `HOME=` prefix; repository doctrine says moving a command into a script file is not a way round
the guard (`kanban-dispatch-mechanics.md:105`: "reduce the verification rather than route it around the guard"), and plan §6 forbids
it. Isolation is the scratch `CLAUDE_CONFIG_DIR`, `CODEX_HOME` and `MOAI_HOME` (the product's own override of the `~/.moai` root,
`paths.go:75-83`), a working directory with no project in it, stubs, the scrub below, and the protected-set hash. The one case these
cannot cover is the real binary's `init`, which writes `$HOME/.claude/settings.json` and for which `HOME` is the only seam (P-42):
that is OD-14, and at its default (a) no harness case runs it (G-8).

Two scripts, two contracts (plan-audit iteration 3, ND-3). `scripts/test-plugin-install-step.sh` is the stub harness: `claude`, `codex`
and `go` resolve to stubs it installed, so it cannot start a real tool. `scripts/check-plugin-discoverable.sh` is **not** under the stub
rule, on purpose: its subject is the real Claude runtime's inventory of an installed plugin (P-30 measured the nested layout invisible
only there), and a stub that replayed a recorded `details` output would test the parser and not the layout. It keeps every other
control of the harness and names the three verbs it may start (last paragraph of this section); the carve-out is measured, not
assumed (P-49: the same three verbs under an empty scratch home left the real roots equal before and after). Both scripts sit inside
the same protected-set bracket from M3 on, when the hash script exists; before that the scratch home, the scrub and the canary are
the controls.

The installer is the other place a harness could write outside its scratch (plan-audit iteration 3, ND-2). Without `--install-dir`,
`install.sh` installs into `go env GOBIN`, else `$GOPATH/bin`, else `$HOME/.local/bin` (P-47; on this machine the first target is
`/Users/goos/go/bin`, which holds a real `moai`), and none of the three is in the protected set. The rule is therefore **prevention,
asserted per case**: every `installer-*` case passes `--install-dir <scratch>/inst-<case>/bin`, a stub `go` returns two decoy
directories inside the scratch for `go env GOBIN` and `go env GOPATH`, and each case fails unless the installed binary's resolved path
lies under its own directory and neither decoy holds a `moai`. `$HOME/.local/bin` is reached only when both decoys are absent, which
the harness never arranges. The hash route was considered and not taken: the protected set is **not** extended to the three install
roots, because an entry hash cannot see the overwrite of a `moai` file that a real `$GOBIN` already holds (control in R-31: equal
listing hash across a content change) and a content hash of that file would count the repository's own rc-install procedure
(`gitflow-lane-protocol.md` §9, `cp bin/moai ~/go/bin/moai`) as a leak. `install.ps1` and `install.bat` take the same option (P-47)
and run in no harness case (G-3).

Isolation mechanism, in order (REQ-025, AC-025):

1. **Capture the protected roots** before anything is changed: `sh scripts/protected-set-hash.sh --save-roots <file>` writes the
   roots derived from the caller's environment (`CLAUDE_CONFIG_DIR`, else `~/.claude`; `CODEX_HOME`, else `~/.codex`), then the
   before hash is taken from that file.
2. **Plant poison in the script's own environment**: `MOAI_CLAUDE_BIN` naming a recording script, a `MOAI_*` variable naming a canary
   directory that stands in for a real home, one variable of each of the `MOAI_*`, `CLAUDE_*` and `CODEX_*` families, and one more of
   each whose name carries the process id, so that no typed list of names can contain it. The poison makes the check independent
   of what the caller's shell happens to hold.
3. **Scrub by live enumeration**: `awk` over `ENVIRON` selects every name matching `^(MOAI|CLAUDE|CODEX)_` and each is unset. On the
   measured session this removes 37 names: the session's own 29, the five fixed poison names and the three pid-named ones (R-28). `ANTHROPIC_*` and the rest are left alone: the
   rule is the three families the audit named.
4. **Set what the run needs**: `PATH` of `<shim>:/usr/bin:/bin`; stub `claude` and `codex` that append their argv to a log; a stub
   `go` that answers `go env GOBIN` and `go env GOPATH` with two decoy directories inside the scratch (created before the run); a stub
   `curl` serving a pinned archive and failing for `checksums.txt`; scratch `CLAUDE_CONFIG_DIR`, `CODEX_HOME`, `MOAI_HOME`; a scratch
   working directory. Each case that needs an opt-out sets `MOAI_SKIP_PLUGIN_INSTALL` for itself after the scrub; each installer case
   passes `--install-dir` for itself.
5. **Judge** five isolation cases, then the product cases, then take the after hash from the same roots file.

Why the working directory matters as much as the scrub: the Claude resolver tries `MOAI_CLAUDE_BIN`, then `llm.claude_bin` from the
project found from the working directory upward, then PATH (P-44, R-25). A `PATH` shim is the last of the three, so a pin outranks it,
and a project pin survives any environment scrub. The case `isolation-cwd-has-no-project` walks from the working directory to `/` and
fails on any `.moai`; the negative controls plant the pin both ways.

`scripts/test-plugin-install-step.sh <moai>`: prints `PASS <name>` or `FAIL <name>: <reason>` per case and a last line
`RESULT pass=<n> fail=<m>`; the exit status is non-zero on any `FAIL`; it runs in CI (the Unix job of `test-install.yml`). Default
cases (12): `isolation-env-scrubbed`, `isolation-cwd-has-no-project`, `isolation-resolves-to-stubs`,
`isolation-poisoned-pin-never-executed`, `isolation-real-home-unchanged`; `verb-install-all-tools`, `verb-no-tools-exit-0`,
`verb-optout-zero-calls`; `installer-calls-verb-by-installed-path`, `installer-optout`, `installer-set-e-guard`,
`installer-old-binary-unknown-verb`. The archive holds the `moai` binary under test, or a stub for the failure cases. **Every
`installer-*` case passes `--install-dir <scratch>/inst-<case>/bin` and asserts the installed path under the scratch root**: it fails
unless `go` resolves to the stub, `realpath <install dir>/moai` exists and lies under that case's own directory, and neither decoy
default root holds a `moai` (AC-018 (a)). A case counts
only logged calls whose first argument is `plugin`, so a tool invoked for another reason does not pollute it. Nine more cases exist
only under OD-14 (b) or (c): `init-claude-harness`, `init-gpt-harness`, `init-both-harness`, `init-config-home-printed`,
`init-add-fails-skips-install`, `init-tool-fails-exit-0`, `init-no-tools-one-skip-line`, `init-flag-optout-zero-calls`,
`init-env-optout-zero-calls`.

Negative controls. `--negative-control` disables the scrub and starts in a project whose `llm.yaml` pins the recorder;
`--negative-control-cwd` keeps the scrub and starts in that project; `--typed-list-mutant` replaces the scrub with a typed list of
the fixed poison names; `--negative-control-install-dir` runs the four installer cases without `--install-dir` (the omitted flag
installs into a decoy, so all four go red through the installed-path and decoy-root assertions, and the five isolation cases stay
green). Each runs its cases and judges them against an expected set: the cases the break affects must go
red, the ones it does not affect must stay green (`isolation-resolves-to-stubs` does not depend on the scrub, so it is the control that
shows the negative mode does not simply fail everything). The exit status is 0 only when the red set and the green set are exactly the
expected ones; a harness whose scrub removal turned nothing red would exit 1. The negative runs aim only at the canary directory and
still take the real-root hash, which must read `LEAK=0`: a negative control never reaches a real profile.

`scripts/protected-set-hash.sh`: read-only; one line `PROTECTED-SET <sha256> entries=<n>`. It hashes directory entries —
`find <root> -print` over every root, so directories and empty directories are lines — plus the content (`shasum -a 256`) of
`settings.json`, `installed_plugins.json` and `config.toml` found inside the roots. A root that does not exist contributes
`ABSENT <path>`, so creating it changes the hash. Roots: `<claude>/plugins` and `<codex>/plugins` to depth 3, `<codex>/.tmp/marketplaces`
to depth 2, `~/.moai` to depth 1, and the files `<claude>/settings.json`, `~/.claude/settings.json`, `<codex>/config.toml`. Depth is
declared because what the runtime churns inside existing marketplace clones and caches sits deeper than what a leak of the t1434 kind
adds (`plugins/data/<x>`, a new marketplace or cache directory: depth 2 to 3); three subtrees are excluded by declaration because they
change on their own, `<claude>/plugins/synced`, `<claude>/plugins/.trash` and `<codex>/.tmp/marketplaces/.staging`; `<codex>/tmp` and
`<codex>/.tmp/git-*` are not roots (P-46, R-29). `--dump <file>` writes the sorted entry list so a non-zero difference can print the
differing entries. Blind spots, stated: `tmp/arg0` under the real Codex home, entries deeper than the declared depth inside an existing
clone or cache, and the content of `known_marketplaces.json` (G-9).

`scripts/check-plugin-discoverable.sh <empty-claude-home>`: the one script that runs the real `claude` (REQ-025 carve-out). Contract:
it refuses any argument that is not an existing empty directory (exit 2) before anything starts; it applies the live-enumerated scrub
of every `MOAI_*`, `CLAUDE_*` and `CODEX_*` name between the marker comments `# scrub:begin` and `# scrub:end` (so a mutant copy
without it can be made mechanically), moves to a scratch working directory and sets `CLAUDE_CONFIG_DIR` to its argument; it starts
no `codex`, takes a local path as its only source (no network source), and runs exactly three verbs, `claude plugin marketplace add
<repository root> --json`, `claude plugin install moai@moai-adk --json` and `claude plugin details moai@moai-adk`. The scrub is not
decoration: `CLAUDE_CODE_PLUGIN_CACHE_DIR`, a variable the real `claude` honors, moves the whole plugin tree out of
`CLAUDE_CONFIG_DIR` (P-49: 391 entries landed in the named directory, none under the scratch home, and the inventory line still read
`ok`), so AC-006 (d) plants that variable and a poison pin aimed at a canary and requires the canary unchanged. The script reads the
inventory, and the expected names are the directory names under `plugins/moai/skills` plus the command stems of the
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
| OD-14 | §6: whether the nine `init-*` harness cases exist and how they obtain a scratch `HOME` |

Rejected without needing a decision: copying the `.claude/` tree as the payload (hides which parts are derivable and leaks
scaffold-only directories); a tag-pinned `sha` inside the repository (a commit cannot contain its own hash); a doctor check that
runs `claude` (`claude plugin list --json` took 0.26 s, and the registry file answers the same question with one read).

## 9. Not designed here

Init shrink (t1438), the mods UI (t1437), Aside (t1439), rules and instruction files carried by a plugin (unmeasured), plugin
refresh by `moai update`, and any behavior of two coexisting copies in a live session (G-2).
