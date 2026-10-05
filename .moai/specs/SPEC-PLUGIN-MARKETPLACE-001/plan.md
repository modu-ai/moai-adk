# Implementation Plan — SPEC-PLUGIN-MARKETPLACE-001

## 1. Overview

The deliverable is code and generated artifacts, not a measurement. Development mode is `tdd`
(`.moai/config/sections/quality.yaml`, `constitution.development_mode`), `cycle_type: tdd`, delegated to
manager-develop per milestone. Tier L (plan-audit ceiling 3 and threshold 0.85, `harness.yaml`
`plan_audit_tier_ceilings`; the tier follows the file count, `progress.md` §E.1). The design is in `design.md`;
the observations behind the premises are in `research.md`.

What gets built, in one paragraph. A Go generator (`pluginemit`, modeled on the existing `commandemit`
and `agentemit` emit-and-golden-pin packages) reads the embedded template tree through the same
core-tier filter `moai init` uses (`LoadEmbeddedCatalog` plus `SlimFS`, `internal/template/slim_fs.go:214`),
and writes (a) the `moai` plugin payload under `plugins/moai/`, with skills copied and commands rendered flat,
(b) the two plugin manifests, and (c) the two marketplace manifests. Every version it writes comes from the one
version SSOT. A read-only drift check and an explicit regenerate verb mirror `agents-emit` / `commands-emit` in the
Makefile. A small install step, written once in Go with an injected command runner, is called by `moai init` (gated
by the `--llm` harness) and by a `moai plugin install` verb that the install scripts call by their installed path. A new
doctor check reads the installed plugin version from the Claude registry file and from a bounded
`codex plugin list --json`, started through the same refusing runner. A shell script compares the plugin version with a release
tag, and a harness, a discoverability check and a protected-set hash script drive the real exec path and the installers offline under
a scrub that is itself shown able to fail (REQ-025).

Why derivation and not a hand list. Card t1399 (lane 3) renames a skill and rule files while this card is
in flight (SPEC §6 RK-4, P-23). A generator that holds no component name survives either land order.

## 2. Decision-reversibility ordering

Milestones below follow build dependency (manifests, then payload, then install, then doctor and release),
because each needs the previous one to exist. Review attention should follow reversibility instead: settle
these first, most likely to change at the top, because each moves files in more than one milestone. The marker table in
`spec.md` §5 lists, per decision, the clauses that change.

1. OD-1 (install default while the scaffold remains), OD-9 (harness gating), OD-10 (target profile), OD-12 (scope), OD-14 (which
   harness cases may run the real binary's `init`): user-facing behavior against a person's real profile, and what the verification
   may touch. They change M3 (when and where the step runs, which harness cases exist) and the risk list.
2. OD-2 (hooks and MCP), OD-3 (agents), OD-8 (tiers), OD-11 (payload language, commands): the payload set and its listing
   cost. They change M2 and the payload size.
3. OD-4 (payload location and pin): changes every manifest path, the Makefile verbs and the release check.
4. OD-5, OD-6 (opt-out surface, script mechanism): change M3 only.
5. OD-7, OD-13 (doctor severity, Codex read path): change one function each.

Mechanical steps (Makefile wiring, golden regeneration, allowlist entries) are deliberately last in each
milestone.

## 3. Milestones (priority-ordered, no time estimates)

The hand-authored change set is the union of the lists below: about 45 files (7 + 5 + 20 + 13), against the Tier L
guidance of more than 15. Generated files (the committed payload) are not counted and are never edited by hand.

### M1 — Emitter skeleton, manifests, MCP derivation, validation (Priority: High) — REQ-001, 002, 003, and the derivation unit of REQ-007

New files (names are proposals; the package path follows `commandemit` / `agentemit`):
- `internal/template/pluginemit/pluginemit.go` — options, entry point, the golden-update switch
  (`PLUGIN_EMIT_UPDATE`, same convention as `COMMAND_EMIT_UPDATE` and `AGENTEMIT_UPDATE`).
- `internal/template/pluginemit/manifest.go` — the four manifests, the version derivation.
- `internal/template/pluginemit/mcp.go` — the MCP entry copied from the template `.mcp.json` (REQ-007). It sits in M1, not M2,
  because the Codex plugin manifest of REQ-003 carries that entry and AC-003 (c) is bound to the exit of M1 (plan-audit iteration
  2, N5); M2 reuses the unit to emit the payload file `plugins/moai/.mcp.json`.
- `internal/template/pluginemit/manifest_test.go`, `golden_test.go`, `mcp_test.go` (`TestMCPEntryDerivedFromTemplate`, AC-007 (c)).

Generated and committed by `make plugin-emit`:
- `.claude-plugin/marketplace.json`, `.agents/plugins/marketplace.json`
- `plugins/moai/.claude-plugin/plugin.json`, `plugins/moai/.codex-plugin/plugin.json`

Modified: `Makefile` (`plugin-emit`, `plugin-emit-check`, `.PHONY`; the check joins `build`'s prerequisites in M2, once it
also covers the payload).

Content rules, copied from precedents so no field is invented:
- Claude marketplace: `name` `moai-adk`; `owner` `{"name": "modu-ai"}` (the moai-cowork marketplace owner);
  `metadata.description` (strict validate requires it, P-02) and `metadata.version`; one entry with `name`,
  `source`, `version`, `description`, `category` (P-11). Entry `source` follows OD-4. `metadata.version` and the entry
  `version` carry one value.
- Codex marketplace: `name` `moai-adk`; `interface.displayName`; one entry with `source`
  `{"source": "local", "path": …}`, `policy` `{"installation": "AVAILABLE", "authentication": "ON_INSTALL"}`,
  `category` (P-11). **No `version` key at either level** (P-35): the precedent has none, and AC-002 (c) pins its
  absence. Codex accepts any manifest (R14-codex), so the shape is held by a golden test, not by the tool.
- Plugin manifests: `name` `moai`, `version`, `description`, `author` `{"name": "MoAI-ADK"}` (strict validate demands an
  author, P-03; `mods/moai-board/.claude-plugin/plugin.json` uses this one), `homepage` and `repository`
  `https://github.com/modu-ai/moai-adk`, `license` `Apache-2.0` (the repository `LICENSE`). The Codex manifest additionally
  names `skills: "./skills/"`, the MCP entry (P-11; under OD-2) and an `interface` block with keys taken from those the
  moai-cowork Codex manifests carry (`displayName`, `shortDescription`, `longDescription`, `developerName`, `category`,
  `websiteURL`; all 18 of those manifests have an `interface` block, P-11) whose values are generator constants pinned by
  the golden; the Claude manifest relies
  on default directory discovery as the precedent does.
- Version: `pkg/version.Version` with a leading `v` stripped, in all four version-carrying fields (P-22, P-04: a
  prerelease value is accepted). Strict validate fails when an entry version differs from plugin.json (P-03), so that
  disagreement is a tool-detected failure; the `metadata.version` equality is held by AC-001 (c).

RED first (tdd): `TestManifestsGolden`, `TestVersionStampedFromSSOT` and `TestMCPEntryDerivedFromTemplate` fail on an empty
package; the version test sets `version.Version` to a synthetic value, so it does not depend on the current stamp, and the MCP test
changes the entry inside a synthetic template tree and asserts the Codex manifest entry follows.

Exit: AC-001, AC-002, AC-003.

### M2 — Payload derivation, layout, discoverability, drift gate (Priority: High) — REQ-004 to 009

New files:
- `internal/template/pluginemit/payload.go` — fs walk over the core-tier view; skills copied, commands rendered and placed
  flat (agents only where OD-3 admits them); every file written with the deployer's mode rule (`.sh` 0755, otherwise 0644), and
  the payload `.mcp.json` written through the M1 unit `mcp.go`.
- `internal/template/pluginemit/drift.go` — compare committed against emitted: byte difference, mode difference, missing
  file, extra file.
- Tests: `payload_test.go` (synthetic-tree derivation, fidelity including modes, layout, the name scan, allow-list), `drift_test.go` (mutants).
- `scripts/check-plugin-discoverable.sh` — installs the emitted marketplace from its local path under a caller-given empty
  scratch home and checks the `claude plugin details` inventory against names derived from the template tree (AC-006);
  refuses a non-empty argument. It is the one script outside REQ-025's stub rule: it runs the real `claude`, for exactly three verbs
  (`plugin marketplace add <repo root> --json`, `plugin install moai@moai-adk --json`, `plugin details moai@moai-adk`), starts no
  `codex`, and applies the live-enumerated scrub between the marker comments `# scrub:begin` and `# scrub:end`, from a scratch working
  directory (and, from M3 on, inside the protected-set bracket); AC-006 (d) shows the scrub able to fail (poison pin and
  `CLAUDE_CODE_PLUGIN_CACHE_DIR` aimed at a canary, P-49) and (e) checks the verb set statically.

Generated and committed: `plugins/moai/skills/**`, `plugins/moai/commands/*.md`, `plugins/moai/.mcp.json`.

Modified: `Makefile` — `plugin-emit-check` added to the `build:` prerequisite line (`Makefile:34`),
regeneration only behind `plugin-emit`, `PLUGIN_EMIT_UPDATE` scrubbed in the check exactly as
`agents-emit-check` scrubs its switch.

Approach points that carry decisions:
- The payload set is an allow-list: only `skills`, `commands`, `.mcp.json` and the two manifest directories
  (plus `agents` if OD-3 admits). A deny-list of rules, hooks and settings would let a new scaffold-only
  directory leak into the payload unnoticed (REQ-008, AC-008).
- **Commands are placed flat, `commands/<name>.md`** (REQ-006). The template sources sit one level down
  (`.claude/commands/moai/<name>.md[.tmpl]`) and mirroring that tree installs a plugin whose inventory counts none of them
  (P-30). Expected invocation names are `moai:<name>` for commands and `moai:<skill>` for skills, read from other plugins'
  listings and not observed for this plugin (SPEC §1.4).
- `.tmpl` files are rendered with the default template context. Command sources use only `.ConversationLanguage`, and
  only on `description:` and `argument-hint:` lines (P-32); its default is English, and the `commandemit` package
  likewise publishes the English variant. Non-`.tmpl` files, including the 16 skill files that merely contain `{{` (P-15),
  are copied byte for byte and never rendered.
- No component name literal appears in the generator, apart from the plugin identifier `moai`, which the generator must write and which
  is also the catalog's first skill name. The core-tier filter is data (the catalog). `TestGeneratorHoldsNoComponentNames` holds this
  against the catalog's own names instead of one literal that t1399 renames, by an AST scan of the string literals of the non-test
  sources, split on `/` and `\`, whole-token and case-sensitive, with an in-test positive control and an empty-sweep failure
  (`acceptance.md` AC-004 (b); the scan passed a stand-in correct generator and the real `commandemit` and `agentemit` packages and
  failed three hard-coded mutants, `research.md` R-27). The iteration-1 wording, "no name as a whole token in any source", could not be
  satisfied by correct code: the plugin id is a skill name and several command stems are ordinary words (plan-audit iteration 2, N1).
- **File modes.** The embedded template tree carries no mode (`embed.FS` reports 0444 for every file, R-26), so the generator cannot
  copy one; it applies the deployer's rule (`internal/template/deployer.go:275-276`): `0755` for a `.sh` file, `0644` for every other
  file, and the drift check compares modes (`mode-flipped`). The three `moai-workflow-project/scripts/*.sh` files are the live case.
  `design.md` §2.4 said 0644 for every file before iteration 2; that was the defect (N8).
- The generator never reads the working tree's own `.claude/` copy; it reads the embedded template tree, so
  maintainer-local drift cannot reach the payload.

RED first: `TestEmitDerivesFromTree` builds a synthetic template tree and synthetic catalog (names unlike
the real ones, one non-core entry) and asserts the emitted set equals the synthetic core set; a hand-copied
name list or a tier-blind generator fails it.

Exit: AC-004, AC-005, AC-006, AC-007, AC-008, AC-009.

### M3 — Install step, verb, scripts, harness (Priority: High) — REQ-010 to 019, 025

New files:
- `internal/cli/plugin_install.go` — the step: tool selection by harness, PATH resolution (`resolveLaunchClaudeBinary`
  for Claude, `codexWiringLookPath` for Codex), the two-command sequence, the per-command bound, the config-home line,
  guidance and skip text, the injected runner. A package-level runner variable is the seam; its default refuses under a test
  binary (REQ-017), and the doctor probe of M4 starts its command through the same variable, so one refusal covers both.
- `internal/cli/plugin_install_test.go` — the unit criteria (AC-010 to AC-015, AC-017).
- `internal/cli/plugin_install_cmd.go` and `plugin_install_cmd_test.go` — the `moai plugin install` verb that OD-6 option (a)
  needs (a new `plugin` noun group with one leaf, in the `tools` help group; AC-019).
- `internal/cli/plugin_install_guard_test.go` — `TestPluginOptOutCallersEnumerated` (AC-016 b) and
  `TestInstallScriptsPluginStepGuarded` (AC-018 c, static).
- `scripts/test-plugin-install-step.sh` — the harness of the conventions block in `acceptance.md`. It plants a poison in its own
  environment (a `MOAI_CLAUDE_BIN` naming a recording script, a variable of each of the `MOAI_*`, `CLAUDE_*` and `CODEX_*` families
  with a pid-carrying name, a `MOAI_*` variable naming a canary directory), scrubs every `MOAI_*`, `CLAUDE_*` and `CODEX_*` name by live
  enumeration, sets its own `PATH`, stub tools (`claude`, `codex` and a `go` whose `go env GOBIN` and `go env GOPATH` answers are two
  decoy directories inside the scratch), a stub `curl`, a pinned local archive, scratch `CLAUDE_CONFIG_DIR`, `CODEX_HOME`,
  `MOAI_HOME` and working directory, and sets no `HOME`. Default cases (12): the five `isolation-*` cases of AC-025, the three
  `verb-*` cases, the four `installer-*` cases. **Every `installer-*` case passes `--install-dir <scratch>/inst-<case>/bin` and
  asserts the installed path under the scratch root** (`go` resolves to the stub, `realpath <install dir>/moai` lies under the
  case's own directory, no decoy default root holds a `moai`; `install.sh` would otherwise install into `$GOBIN`, `$GOPATH/bin` or
  `$HOME/.local/bin`, P-47). The protected set is **not** extended to those three roots (an entry hash cannot see the overwrite of the
  `moai` a real `$GOBIN` holds, and a content hash would read the rc-install procedure as a leak, `design.md` §6): they are closed
  by prevention and by this per-case assertion. The nine `init-*` cases exist only under OD-14 (b) or (c). Flags
  `--negative-control`, `--negative-control-cwd`, `--typed-list-mutant` and `--negative-control-install-dir` (the four installer
  cases without `--install-dir`, all four expected red) run their cases under a deliberate break and exit 0
  only when exactly the expected cases go red (AC-025 (b), (c), (f)).
- `scripts/protected-set-hash.sh` — read-only hash of the protected real set (entries and registry or settings file content,
  depth-limited roots, three declared exclusions, AC-025 (d)); `--save-roots` and `--roots-file` let the harness capture the roots
  before the scrub and hash exactly them after; a non-zero difference prints the differing entries.

Modified:
- `internal/config/envkeys.go` — the opt-out constant (`MOAI_SKIP_PLUGIN_INSTALL`; env names live there,
  `internal/cli/CLAUDE.md` Env var access).
- `internal/config/defaults.go` — `DefaultPluginInstallCommandTimeout` (60 seconds, one command) and
  `DefaultPluginVersionProbeTimeout` (3 seconds, M4), per the repository's hardcoding rule (thresholds live in `defaults.go`).
- `internal/cli/init.go` — the call, placed adjacent to the `wireCodexUnlessClaude` call (`init.go:1011`; the function is
  defined at `:192`) in the `runInit` tail; the `--no-plugin` flag next to the other `initCmd.Flags()` lines; guidance
  goes through the printer's stderr, human status being stderr (`internal/cli/CLAUDE.md` Output streams).
- `internal/cli/claude_binary.go` — a minimal change the step needs: the resolver returns plain errors, so "not found on PATH"
  (REQ-014) and "pin invalid" (REQ-013) cannot be told apart, and it reads the pin from the working directory's project
  (`findProjectRoot()`), not from the init target. The change adds a typed error for the not-found class and a variant that
  takes the project root (`design.md` §3.2). The existing launch path keeps calling the unchanged entry point.
- `internal/cli/doctor_agentemit_embed.go:352` and its test — the self-invocation environment gains the opt-out.
- `e2e/cli/tux3_journeys.sh` — the opt-out on the two `moai init` lines (104 and 115 at this tree).
- `install.sh`, `install.ps1`, `install.bat`, and byte-identical copies `docs-site/static/install.sh`,
  `docs-site/static/install.ps1` (the parity job in `test-install.yml` diffs them).
- `.github/workflows/test-install.yml` — the Unix job builds `moai` and runs the harness script, and the path filters gain
  the harness script and `internal/cli/plugin_install*.go` (this is the continued-firing answer of the harness: it runs
  in CI, not on request).

Behavior notes the run phase must hold:
- Two commands per tool, second only after the first exits 0 (REQ-011). No `--scope` is passed at the OD-12 default, so
  Claude's own default (user scope, P-06, P-33) applies. No `--yes` is passed: it is needed only for command-source
  installs, and a prompt with no TTY fails into the guidance path.
- The child inherits the invoking environment unchanged at the OD-10 default; the step never sets or clears
  `CLAUDE_CONFIG_DIR` or `CODEX_HOME` and never enumerates other profile directories. The line it prints names the config
  home the tool resolves (`CLAUDE_CONFIG_DIR`, else `~/.claude`; `CODEX_HOME`, else `~/.codex`) so the person sees which
  profile moved.
- Absence of a tool is one skip line, not a guidance block (REQ-014). A non-zero exit, a timeout, an invalid
  `llm.claude_bin` pin: one guidance block naming that tool's two manual commands, no further command for that tool,
  step returns success (REQ-013).
- The harness gates the step at the OD-9 default: `claude` selects Claude only, `gpt` Codex only, `both` both, the same
  rule as `wireCodexUnlessClaude` and the `mcpDeclined` switch (P-28). The verb and the scripts have no harness and act on
  every tool found.
- The test-binary detector keys on the program name (`os.Args[0]`) or `testing.Testing()`, not on any argument: the
  existing `isTestEnvironment()` (`internal/cli/glm.go:1058`) scans every argument for a `.test` suffix and would read a
  project named `app.test` as a test (`design.md` §3.4).
- The installers call the verb by the installed path (`$TARGET_PATH`, `%TARGET_PATH%`, the PowerShell install path), never
  by a bare `moai`: the install directory need not be on PATH (P-37). The call is guarded against `set -e` and
  `$ErrorActionPreference = "Stop"`, and a binary that predates the verb must not fail the installer (RK-14).
- Windows: `install.ps1` and `install.bat` call the same verb after the binary lands; `install.bat` is exercised end to end
  by the Windows job in `test-install.yml`, which downloads the latest release, so the call there hits a binary that lacks the
  verb until the release ships — the guard is what keeps that job green (RK-12, RK-14).

RED first: `TestPluginInstallStep_Sequence`, `_FailOpen`, `_ToolAbsent`, `_OptOut`, `_Environment`, `_NoRealRunnerUnderTest`
fail on an absent step. The inertness test pins `MOAI_CLAUDE_BIN` to a recording script and places recording `claude` and
`codex` shims on PATH with scratch `CLAUDE_CONFIG_DIR` and `CODEX_HOME`, calls the existing `runInit` path, and asserts an
empty record — a PATH shim alone would miss a pinned binary.

Exit: AC-010 to AC-019, AC-025.

### M4 — Doctor check, release coupling, runbook (Priority: Medium) — REQ-020 to 024

New files:
- `internal/cli/doctor_plugin_version.go` and `doctor_plugin_version_test.go` — modeled on `doctor_mcp_version.go`: a constant
  check name, one function with the homes, the probe runner and the binary version injected so no test mutates package
  state. Production wiring: the Claude home from `CLAUDE_CONFIG_DIR`, else `~/.claude`; the Codex home from the existing
  `resolveCodexHomeDir()` (`mcp_codex.go:2143-2152`, which rides the `codexUserHomeDir` seam) and never from a path the check
  computes itself; the probe started through the runner variable of `plugin_install.go` (default refuses under a test binary,
  REQ-017) with the parent's environment plus `CODEX_HOME=<that home>`. `TestCheckPluginVersion_HomeIsolation` (AC-021 (c)) lives
  here; its registry-wide subtest pins `CLAUDE_CODE_VERSION` (the existing `Claude Code` check execs `claude --version` itself when it
  is unset, P-48) and then requires an empty record of recording `claude` and `codex` shims.
- `scripts/check-plugin-version.sh` — reads `plugins/moai/.claude-plugin/plugin.json` and compares with the
  tag argument minus a leading `v`; exit 0 equal, exit 1 with both values on mismatch, exit 2 without an argument.
- `internal/template/pluginemit/releasecheck_test.go` — `TestPluginVersionScript` and
  `TestReleaseWorkflowCallsPluginVersionCheck` (the release coupling lives with the generator that writes the version).

Modified:
- `internal/cli/doctor.go` — one registry row.
- `internal/cli/binary_lag_test.go` — the allowlist entry (a bare identifier for a constant-registered name,
  `namesAddedAfterBaseline`), and the status-set test if the check can emit info.
- `internal/cli/doctor_golden_test.go` — `captureDoctorCmd` gains the `CODEX_HOME` scrub (it pins `HOME`, `MOAI_HOME`,
  `CLAUDE_CONFIG_DIR` and the Codex PATH seam but not `CODEX_HOME`, P-21), and `TestDoctorGolden_IgnoresCallerCodexHome`.
- `internal/cli/main_test.go` — `sandboxUserHomeDir` (`:269-299`, installed by `TestMain` at `:368`) also wraps `codexUserHomeDir`
  with `homeRedirectingFn` (so a result equal to the captured real home becomes the sandbox directory) and clears `CODEX_HOME` for the
  package run, and its restore function puts both back. Why here and not a pin in each test: seven unfiltered
  `runDiagnosticChecks(false, "")` call sites in six test files exist (`doctor_hook_wiring_test.go:250`, `doctor_golden_test.go:263`,
  `doctor_test.go:280,422`, `doctor_new_test.go:320,329`, `doctor_flag_slot_test.go:181`, P-43) and a pin per test would leave the
  next one unguarded; the seven test files that assign `codexUserHomeDir` capture the current value as `orig` and restore it, so they
  restore the sandboxed value. The two layers are independent: the refusing runner (REQ-017) stops a process start, this stops a
  read of the real Codex home; `TestCheckPluginVersion_HomeIsolation` pins both (AC-021 (c) `testmain-sandbox-redirects-codex-home`,
  `registry-wide-starts-nothing`).
- `internal/cli/testdata/doctor-{light,dark,nocolor}.golden` — regenerated with `UPDATE_GOLDEN=1`; each then holds one
  `Plugin Version` row (AC-023 a).
- `.github/workflows/release.yml` — a check 8 in `verify-provenance` that runs the script against the tag.
- `.moai/docs/version-management.md` — a third group beside Version Stamps and Release Artifacts, "Generated version
  carriers": the bump step `make plugin-emit`, and the pre-tag step `sh scripts/check-plugin-discoverable.sh <empty-home>`.
  The group opens under its own bold label, `**Generated version carriers:**`, and is never written as a bullet inside the
  Version Stamps list: `TestVersionSyncListNamesOnlyExistingPaths` (`internal/cli/version_sync_list_test.go`) reads the entries
  after `**Version Stamps:**` up to the next bold label or `###` heading and requires exactly seven, so an eighth plain bullet
  there turns it red (plan-audit iteration 2, N13). The release harness's own runbook was not located in this tree (Gap G-6); this
  document is the carrier the card can name.

Doctor read path. The Claude version is read from `<config home>/plugins/installed_plugins.json`, key
`moai@moai-adk` (P-08), with no subprocess: the cost is one file read regardless of project size, which the
Advisory-Check Discipline in `coding-standards.md` requires of anything on a latency-sensitive path. The Codex state and
version come from `codex plugin list --json`, started through the refusing runner with `CODEX_HOME` set to the resolved Codex home,
bounded by `DefaultPluginVersionProbeTimeout` and started only when `codex` resolves on PATH (REQ-021, default pending OD-13): the list
answers from the `config.toml` registration, which the cache directory does not (P-34: after the registration is removed
the list prints `"installed": []` while `plugins/cache/moai-adk/moai/3.1.3` remains). `moai doctor` is an on-demand command,
not a session-start or Stop path, so a bounded probe is within the discipline; the write of `tmp/arg0` into an empty Codex
home is the cost OD-13 names. What is UNOBSERVED: formats at any version other than claude 2.1.287 and codex 0.160.0; any
state when `claude` is absent but a registry file remains (the file is read regardless; absence of the tool is not
consulted). `IsDevBuild` (`pkg/version/version.go`) decides the dev-build case. Severity follows OD-7; the default never
gates a doctor exit on absence.

Release coupling. The generator owns every version field (REQ-003), so a bump is: rewrite the seven existing stamp files
(`.moai/docs/version-management.md`), then run `make plugin-emit`; skipping the second step turns `plugin-emit-check` red at
`make build` and the golden test red in `go test`. The plugin files carry the version without a leading `v`, so the
literal-token registry sweep (`version_stamp_registry_test.go`, token `v3.1.3`) does not see them and needs no new entry; the
drift gate is their guard instead, and the runbook edit above puts the step where the person bumping looks. Release check 8
closes the tag side: the tagged tree's plugin version must equal the tag.

Exit: AC-020 to AC-024.

## 4. Verification plan (scratch-home commands only)

Rules for every acceptance command and every harness script (REQ-025):
- `CLAUDE_CONFIG_DIR` and `CODEX_HOME` point at empty directories created under the session scratchpad
  (`mkdir -p`, then `ls -A` shows empty) before the first command, **one pair per criterion**. Placeholders in
  `acceptance.md`: `<claude-home:AC-nnn>`, `<codex-home:AC-nnn>`. Neither a command nor a script sets `HOME` (the worktree guard
  refuses it, P-36, and a script file is not a way round that, `kanban-dispatch-mechanics.md:105`); `MOAI_HOME` and the two homes
  are the explicit seams, and the one case they cannot cover, the real binary's `init` (P-42), is OD-14 and G-8.
- The install-step harness (`scripts/test-plugin-install-step.sh`) scrubs by live enumeration: every `MOAI_*`, `CLAUDE_*` and
  `CODEX_*` name in its environment is unset, found at run
  time and never from a typed list, after the script has planted its own poison (including names that carry the process id); it then
  sets its own `PATH`, stubs (`claude`, `codex`, `go`, `curl`) and scratch directories, and runs from a scratch directory with no
  `.moai` in it or above it. A pinned
  real `claude` can come from `MOAI_CLAUDE_BIN` or from the working directory's `llm.claude_bin` (P-44), and a `PATH` shim outranks
  neither, so the scrub and the working directory are both required. Every installer case passes `--install-dir <scratch>` and
  asserts the installed path under it (REQ-025, AC-018 (a), AC-025 (f)).
- `scripts/check-plugin-discoverable.sh` is the one script that runs the real `claude` instead of a stub (REQ-025 carve-out, AC-006): it
  may start exactly `claude plugin marketplace add <local repo path>`, `claude plugin install moai@moai-adk` and
  `claude plugin details moai@moai-adk`, under a scratch `CLAUDE_CONFIG_DIR` that it refuses unless it is an existing empty
  directory, behind the same live-enumerated scrub, from a scratch working directory, starting no `codex`, with no network source and
  no `HOME` assignment, inside the same protected-set bracket (from M3 on, when the hash script exists). The scrub matters beyond
  the config home: the real `claude` honors
  `CLAUDE_CODE_PLUGIN_CACHE_DIR`, which moves the plugin tree out of the config home (P-49).
- The whole run is bracketed by `sh scripts/protected-set-hash.sh` before and after, the two lines equal (AC-025 (d)). The
  hash covers directory entries because the t1434 leak was four empty directories.
- Forbidden against any real home, as in t1434: install, uninstall, marketplace add, remove, update, enable,
  disable, configure, `plugin update`, `codex exec`. `claude plugin list`, `validate`, `details` and the
  read-only help verbs are the only commands the run may aim at a real home, and only if a scratch home
  cannot answer; none of the 25 criteria needs it.
- No network. `marketplace add` takes a local path (the repository root in the run worktree). The product
  command `marketplace add modu-ai/moai-adk` is verified only through the injected runner (AC-011) and stays
  UNOBSERVED against GitHub until the manifests are on `main`.
- Go tests use the anchored plain form `-run '^Name$'` (the `VacuousTestAssertion` lint requires it; the
  worktree guard accepts it as a single invocation and refuses it bundled with a redirect and `; echo $?`,
  P-25), and the criterion requires the named `--- PASS: Name (…)` line, because `go test` exits 0 with
  `no tests to run` when a selector matches nothing, which would read as a pass.
- Reporting follows the five-section format of `verification-claim-integrity.md` §3, with the judging
  build's commit next to the tree HEAD (§2.2): `bin/moai` built from the tree, invoked by path.
- A command the worktree guard refuses is named in the report's Gaps with what was done instead
  (`verification-claim-integrity.md` §3.1). The two refusals already measured are `pwsh` and a `HOME=` prefix.

Gaps this plan names instead of closing (carried into the run's Gaps section):
- G-1: the GitHub path. `claude plugin marketplace add modu-ai/moai-adk` against the public repository is
  not run; `main` has no manifest until the release PR (P-24, P-41: `.claude-plugin` and `plugins` return 404 on `main`).
- G-2: runtime behavior of two coexisting copies (shadowing, collision of `moai:<command>`, hook double firing), and
  plugin-agent permission behavior. Needs an authenticated session; belongs to t1438's measurement.
- G-3: script behavior beyond the offline `install.sh` harness. `install.ps1` and `install.bat` are not executed locally:
  `pwsh` is installed at `/usr/local/bin/pwsh` but the worktree guard refuses it (P-36), and there is no Windows host. AC-018 (c)
  is static for them; CI tokenises `install.ps1` and runs `install.bat` end to end, neither of which exercises the step.
- G-4: Codex field effect (R14-codex): acceptance by `codex plugin add` is shown, field reading is not.
- G-5: project and local installation scope (OD-12 b and c): `claude plugin install --scope project|local` behavior, the
  registry shape it writes and the doctor read of it are unobserved; only the flag text was read (P-33).
- G-6: the release harness's own runbook was not located in this tree; the card edits `version-management.md` and the release
  workflow, the two release surfaces it could verify.
- G-7: continued firing of `scripts/check-plugin-discoverable.sh`. It needs a Claude CLI and runs on demand, so nothing tells a
  reader it stopped. Its answer is the runbook step AC-024 (d) pins plus the plan-auditor's re-run at each release PR; the
  harness script, by contrast, runs in CI (M3), and the Go criteria run in the full suite.
- G-8: the real exec path of the binary's `init`. `moai init` writes `$HOME/.claude/settings.json` whatever `CLAUDE_CONFIG_DIR` says
  (P-42), `HOME` is the only seam, and neither a command nor a script may set it (§3 of the SPEC), so at the OD-14 default no
  harness case runs the real binary's `init`. The init flow (harness gating, opt-out, fail-open, skip line, config-home print) is
  asserted by Go tests with an injected runner under the `TestMain` home sandbox; the real default runner is exercised by the verb and
  installer cases. What stays unobserved is init's own call into the step through a real process start.
- G-9: the protected-set hash is blind to `tmp/arg0` and `.tmp/git-*` under the real Codex home (P-46: other Codex processes write
  there), and to entries deeper than its declared depth inside an existing marketplace clone or plugin cache. The probe's own
  `tmp/arg0` write is therefore guarded by the refusing runner (REQ-017), the stubs and `isolation-resolves-to-stubs`, not by the
  hash; a non-zero `LEAK` prints the differing entries so an ambient change is attributed, not assumed.

## 5. Pre-flight and land order (for the run-phase delegation)

Record at run start, before any edit:
- `git rev-parse HEAD` and `git branch --show-current` (a moved HEAD means another writer; stop and report).
- Whether the base already contains t1399: `grep -c moai-factory-foreman internal/template/catalog.yaml`
  (0 at the tree measured here, so the rename has not landed). Either order is safe because of REQ-004.
- **Owner of the red when t1399 lands second.** The card that lands second turns `plugin-emit-check` red (the golden test
  fails in the full suite) until `make plugin-emit` is run. t1399's lane scopes verification to its own selectors
  (`AGENTS.md` §4), so the red would otherwise surface on `origin/develop` CI after the leader's batch push, and t1399 is
  told nothing. The owner is therefore the lane landing second: it runs `make plugin-emit` inside its merge and commits the
  delta (`plugins/moai/**` and the four manifests). The leader's dispatch carries one line to say so, and the line is recorded
  in `progress.md` §E.1 under "dispatch instructions" so the run-phase delegation can cite it.
- Baseline: `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`, and
  `go test ./internal/cli -run '^TestBinaryLag_.*$'` green before touching doctor.
- Re-count the automated callers of `moai init` at run start: `grep -rn 'runInit(' internal/cli/*_test.go` (33 lines, one a
  string literal, at the tree measured here), `grep -rn 'initCmd.RunE(' internal` (9), `grep -rn '"init"' internal --include='*.go'`
  without the `git init` helpers (1: `doctor_agentemit_embed.go`), and `grep -n "init " e2e/cli/tux3_journeys.sh` (two runs).
- The tests that enumerate root commands (`grep -rln 'rootCmd.Commands()' internal/cli/*_test.go`, at least ten files at this tree
  (the listing was cut at ten), among them `help_order_test.go`, `help_linter_stale_test.go` and `doctor_test.go`) are run by name when the verb lands.
- Scope of local test runs: only packages the change can affect (`internal/template/pluginemit`,
  `internal/cli` with `-run` selectors, `internal/config`, `internal/template` golden and neutrality selectors). No
  `go test ./...` (AGENTS.md §4); CI runs the full suite. The harness script is run once, against `bin/moai`.
- A doctor-touching card remeasures `-run '^TestRunDiagnosticChecks.*$'`, `-run '^TestDoctorGolden_.*$'` and
  `-run '^TestBinaryLag_.*$'` together (the three guards live apart from the card's own tests). M4 also edits two things those
  guards do not cover, so its remeasure list adds `-run '^TestVersionSyncListNamesOnlyExistingPaths$'` (it parses the runbook M4
  edits and requires exactly seven Version Stamps; green at `b6a0522a0`, L-40) and, because `main_test.go` is edited, the selectors
  of the files that assign `codexUserHomeDir` (`doctor_codex_test.go`, `codex_skills_*_test.go`, `codex_readiness_test.go`,
  `doctor_codex_stale_skill_test.go`: `-run '^TestCodex.*$'` and `-run '^TestDoctorCodex.*$'`, run before and after the `TestMain`
  change and compared).
- Record the protected-set line (`sh scripts/protected-set-hash.sh`) before the first command of the run and after the last (AC-025 (d)).

## 6. Constraints for manager-develop

PRESERVE (do not modify): the deployed file set of `moai init` and `moai update` (this card adds, never
removes; existing init and update deploy tests run unmodified), the 32 `runInit(` and 9 `initCmd.RunE(` test call sites
(the default-refusing runner covers them, so none is edited), `.claude/` local copies not mirrored from a
template, `.moai/specs/` other than this SPEC's `progress.md §E.2/§E.3` and the `status:`/`updated:` lines,
`mods/moai-board/**`, the `moai-cowork` files under the user home, and everything in `e2e/cli/tux3_journeys.sh` except the
two opt-out prefixes.
Forbidden commands: `--no-verify`, `--amend`, `git add -A`, `git add .`, force push, any registry-writing
`claude plugin` / `codex plugin` verb against a real home, any network-touching marketplace add in an
acceptance command, any command that sets `HOME` or runs `pwsh` (the guard refuses both), and any script that sets `HOME` to get
past that refusal: moving a command into a script file is not a workaround (`kanban-dispatch-mechanics.md:105`), the case that
needs it is OD-14 and G-8.
Required: Conventional Commits with the card id in every commit message; stage by explicit pathspec; cite
`worktree-integration-ops.md` form rules when an acceptance command is refused.
Required for the harness (binding run-phase instruction, plan-audit iteration 3, ND-2): **every `installer-*` case passes
`--install-dir <scratch>/inst-<case>/bin` and asserts the installed path under the scratch root**, with a stub `go` that makes
`install.sh`'s default install roots decoys inside the scratch; a case written without the flag is a red (AC-025 (f)), not a
write into `$GOBIN`, `$GOPATH/bin` or the real `$HOME/.local/bin`. Required for the discoverable script: it keeps the real `claude`, the
three named verbs and the scrub of the carve-out in REQ-025, and nothing else starts.

## 7. Anti-patterns

- A hand-maintained list of skills, agents or commands anywhere in the generator or the Makefile (the
  failure RK-4 exists to prevent).
- A deny-list payload filter (new scaffold-only directories would leak).
- Mirroring the template's `commands/moai/` directory into the plugin (the inventory counts none of it, P-30).
- A doctor check that spawns `claude`, or spawns `codex` without a deadline or when it is absent from PATH.
- An install step whose default runner is reachable from `go test` (RK-8), or a test-binary detector that reads arguments.
- An installer that calls the verb through a bare `moai`, or without a guard against `set -e` and `Stop`.
- Copying the plugin version by hand into the manifests (REQ-003).
- An acceptance criterion or a script that sets `HOME`, shares a scratch home with another criterion, or passes on a mutant that never
  registers the check (AC-023 (a) exists for that).
- Reading the real profile's registry in any test; fixtures are synthetic files under `t.TempDir()`.
- An isolation check with no negative control: a scrub whose removal turns nothing red, a pin check that a `PATH` shim alone would
  satisfy, a manifest that hashes files and not directory entries, or an environment scrub written as a typed list of names.
- A name scan over every token of every source (identifiers, comments, the plugin id), which no correct generator can pass.
- Writing every payload file at one mode, or copying a mode the embedded tree does not carry.

## 8. Cross-references

- `.moai/specs/SPEC-PLUGIN-LOAD-SCOPE-001/progress.md §E.2` — the t1434 verdict table (R01 to R14).
- `design.md` — architecture, payload derivation, install-step control flow, doctor data flow, release coupling.
- `research.md` — the observations behind each premise, with sources and version stamps.
- `internal/template/commandemit/commandemit.go`, `internal/template/agentemit/` — emit, golden pin, update switch.
- `Makefile:34-62` — `build` prerequisites, `agents-emit*`, `commands-emit*` precedents.
- `internal/cli/doctor_mcp_version.go` — check shape; `internal/cli/binary_lag_test.go:198` — allowlist.
- `internal/cli/init.go:133,192,997-1011` — flag help, `wireCodexUnlessClaude`, the `mcpDeclined` switch and the call site.
- `.moai/docs/version-management.md` — version SSOT, stamp list, release process.
- `.claude/rules/moai/development/verification-completeness.md` §2 — two-cell adoption used in `acceptance.md`.
