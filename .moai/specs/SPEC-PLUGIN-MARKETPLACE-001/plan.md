# Implementation Plan — SPEC-PLUGIN-MARKETPLACE-001

## 1. Overview

The deliverable is code and generated artifacts, not a measurement. Development mode is `tdd`
(`.moai/config/sections/quality.yaml`, `constitution.development_mode`), `cycle_type: tdd`, delegated to
manager-develop per milestone. Tier M (plan-audit ceiling 2, `harness.yaml` `plan_audit_tier_ceilings`).

What gets built, in one paragraph. A Go generator (`pluginemit`, modeled on the existing `commandemit`
and `agentemit` emit-and-golden-pin packages) reads the embedded template tree through the same
core-tier filter `moai init` uses (`LoadEmbeddedCatalog` plus `SlimFS`, `internal/template/slim_fs.go:214`),
and writes (a) the `moai` plugin payload under `plugins/moai/`, (b) the two plugin manifests, and (c) the
two marketplace manifests. Every version it writes comes from the one version SSOT. A read-only drift
check and an explicit regenerate verb mirror `agents-emit` / `commands-emit` in the Makefile. A small
install step, written once in Go with an injected command runner, is called by `moai init` and by a
`moai` verb the install scripts call. A new doctor check reads the installed plugin version from the
Claude and Codex homes. A shell script compares the plugin version with a release tag.

Why derivation and not a hand list. Card t1399 (lane 3) renames a skill and rule files while this card is
in flight (SPEC §6 RK-4, P-23). A generator that holds no component name survives either land order.

## 2. Decision-reversibility ordering

Milestones below follow build dependency (manifests, then payload, then install, then doctor and release),
because each needs the previous one to exist. Review attention should follow reversibility instead: settle
these first, most likely to change at the top, because each moves files in more than one milestone.

1. OD-1 (install default while the scaffold remains), OD-2 (hooks and MCP), OD-3 (agents): user-facing
   behavior and a safety posture. They change M2 (payload set), M3 (when the step runs) and the risk list.
2. OD-4 (payload location and pin): changes every manifest path, the Makefile verbs and the release check.
3. OD-5, OD-6 (opt-out surface, script mechanism): change M3 only.
4. OD-7, OD-8 (doctor severity, tier scope): change one function each.

Mechanical steps (Makefile wiring, golden regeneration, allowlist entries) are deliberately last in each
milestone.

## 3. Milestones (priority-ordered, no time estimates)

### M1 — Emitter skeleton, manifests, validation (Priority: High) — REQ-001, 002, 003, 015

New files (names are proposals; the package path follows `commandemit` / `agentemit`):
- `internal/template/pluginemit/pluginemit.go` — options, entry point, the golden-update switch
  (`PLUGIN_EMIT_UPDATE`, same convention as `COMMAND_EMIT_UPDATE` and `AGENTEMIT_UPDATE`).
- `internal/template/pluginemit/manifest.go` — the four manifests, the version derivation.
- `internal/template/pluginemit/manifest_test.go`, `golden_test.go`.

Generated and committed by `make plugin-emit`:
- `.claude-plugin/marketplace.json`, `.agents/plugins/marketplace.json`
- `plugins/moai/.claude-plugin/plugin.json`, `plugins/moai/.codex-plugin/plugin.json`

Modified: `Makefile` (`plugin-emit`, `plugin-emit-check`, `.PHONY`; the check joins `build`'s prerequisites
in M2, once it also covers the payload).

Content rules, copied from precedents so no field is invented:
- Claude marketplace: `name` `moai-adk`; `owner` `{"name": "modu-ai"}` (the moai-cowork marketplace owner);
  `metadata.description` (strict validate requires it, P-02) and `metadata.version`; one entry with `name`,
  `source`, `version`, `description`, `category` (P-11). Entry `source` follows OD-4.
- Codex marketplace: `name` `moai-adk`; `interface.displayName`; one entry with `source`
  `{"source": "local", "path": …}`, `policy` `{"installation": "AVAILABLE", "authentication": "ON_INSTALL"}`,
  `category` (P-11). Codex accepts any manifest (R14-codex), so the shape is held by a golden test, not by
  the tool.
- Plugin manifests: `name` `moai`, `version`, `description`, `author` `{"name": "MoAI-ADK"}` (the author
  `mods/moai-board/.claude-plugin/plugin.json` uses), `homepage` and `repository`
  `https://github.com/modu-ai/moai-adk`, `license` `Apache-2.0` (the repository `LICENSE`). The Codex
  manifest additionally names `skills: "./skills/"` and the MCP entry (P-11); the Claude manifest relies on
  default directory discovery as the precedent does.
- Version: `pkg/version.Version` with a leading `v` stripped, in all four manifests (P-22, P-04: a prerelease
  value is accepted). Strict validate fails when an entry version differs from plugin.json (P-03), so that
  disagreement is a tool-detected failure.

RED first (tdd): `TestManifestsGolden` and `TestVersionStampedFromSSOT` fail on an empty package; the
version test sets `version.Version` to a synthetic value, so it does not depend on the current stamp.

Exit: AC-001, AC-002, AC-003, AC-015.

### M2 — Payload derivation, exclusions, drift gate (Priority: High) — REQ-004 to 008

New files:
- `internal/template/pluginemit/payload.go` — fs walk over the core-tier view; skills and commands (agents
  only where OD-3 admits them).
- `internal/template/pluginemit/mcp.go` — the MCP entry copied from the template `.mcp.json`.
- `internal/template/pluginemit/drift.go` — compare committed against emitted: byte difference, missing
  file, extra file.
- Tests: `payload_test.go` (synthetic-tree derivation, fidelity, allow-list), `drift_test.go` (mutants).

Generated and committed: `plugins/moai/skills/**`, `plugins/moai/commands/**`, `plugins/moai/.mcp.json`.

Modified: `Makefile` — `plugin-emit-check` added to the `build:` prerequisite line (`Makefile:34`),
regeneration only behind `plugin-emit`, `PLUGIN_EMIT_UPDATE` scrubbed in the check exactly as
`agents-emit-check` scrubs its switch.

Approach points that carry decisions:
- The payload set is an allow-list: only `skills`, `commands`, `.mcp.json` and the two manifest directories
  (plus `agents` if OD-3 admits). A deny-list of rules, hooks and settings would let a new scaffold-only
  directory leak into the payload unnoticed (REQ-007, AC-007).
- `.tmpl` files are rendered with the default template context. Command sources use only
  `.ConversationLanguage` (observed: the only action in 15 files), whose default is English; the
  `commandemit` package likewise publishes the English variant. Non-`.tmpl` files, including the 16 skill
  files that merely contain `{{` (P-15 listing: html-report mustache templates, doc templates, a JSON
  schema), are copied byte for byte and never rendered.
- No component name literal appears in the generator. The core-tier filter is data (the catalog).
- The generator never reads the working tree's own `.claude/` copy; it reads the embedded template tree, so
  maintainer-local drift cannot reach the payload.

RED first: `TestEmitDerivesFromTree` builds a synthetic template tree and synthetic catalog (names unlike
the real ones, one non-core entry) and asserts the emitted set equals the synthetic core set; a hand-copied
name list or a tier-blind generator fails it.

Exit: AC-004, AC-005, AC-006, AC-007, AC-008.

### M3 — Install step (Priority: High) — REQ-009 to 012

New files:
- `internal/cli/plugin_install.go` — the step: PATH resolution (`resolveLaunchClaudeBinary` for Claude,
  `codexWiringLookPath` for Codex), the two-command sequence, bounded time per command, guidance text, the
  injected runner. A package-level runner variable is the seam; its default must not be reachable from a
  test binary without an injection (RK-8).
- `internal/cli/plugin_install_test.go`, plus the `moai` verb that OD-6 option (a) needs (a leaf under an
  existing noun group or a new `plugin` group; the name is settled in M3 and listed in `moai --help`).
- `internal/config/envkeys.go` — the opt-out constant (`MOAI_SKIP_PLUGIN_INSTALL`; env names live there,
  `internal/cli/CLAUDE.md` Env var access).

Modified:
- `internal/cli/init.go` — the call, placed adjacent to `wireCodexUnlessClaude` in the `runInit` tail
  (`init.go:192` is the neighbouring call); the `--no-plugin` flag next to the other `initCmd.Flags()` lines;
  guidance goes through the printer's stderr, human status being stderr (`internal/cli/CLAUDE.md` Output
  streams).
- `internal/cli/doctor_agentemit_embed.go:352` — the self-invocation environment gains the opt-out.
- `install.sh`, `install.ps1`, `install.bat`, and byte-identical copies `docs-site/static/install.sh`,
  `docs-site/static/install.ps1` (the parity job in `test-install.yml` diffs them).

Behavior notes the run phase must hold:
- Two commands per tool, second only after the first exits 0 (REQ-009). No `--scope` is passed, so Claude's
  own default (user scope, P-06) applies. No `--yes` is passed: it is needed only for command-source
  installs, and a prompt with no TTY fails into the guidance path.
- The child inherits the invoking environment unchanged; the step never sets or clears `CLAUDE_CONFIG_DIR`
  or `CODEX_HOME` and never enumerates other profile directories. The line it prints names the config home
  the tool resolves (`CLAUDE_CONFIG_DIR`, else `~/.claude`; `CODEX_HOME`, else `~/.codex`) so the person
  sees which profile moved.
- Absence of a tool is a one-line skip, not a warning block. A non-zero exit, a timeout, an invalid
  `llm.claude_bin` pin: one guidance block naming both manual commands, step returns success.
- `--llm` does not gate the step: the card says "when `claude` exists … codex likewise", and the scripts
  have no harness selection. OD-1 and OD-5 are the controls.
- Windows: `install.ps1` and `install.bat` call the same verb after the binary lands; `install.bat` is
  exercised end to end by the Windows job in `test-install.yml` (RK-12).

RED first: `TestPluginInstallStep_Sequence`, `_FailOpen`, `_OptOut`, `_NoRealRunnerUnderTest` fail on an
absent step; the last one places a PATH shim `claude` and `codex` that record invocations, calls the
existing `runInit` paths, and asserts zero recorded calls.

Exit: AC-009, AC-010, AC-011, AC-012.

### M4 — Doctor check and release coupling (Priority: Medium) — REQ-013, 014, 016

New files:
- `internal/cli/doctor_plugin_version.go` and `doctor_plugin_version_test.go` — modeled on
  `doctor_mcp_version.go`: a constant check name, one function with the home and the binary version
  injected so no test mutates package state.
- `scripts/check-plugin-version.sh` — reads `plugins/moai/.claude-plugin/plugin.json` and compares with the
  tag argument minus a leading `v`; exit 0 equal, exit 1 with both values on mismatch.

Modified:
- `internal/cli/doctor.go` — one registry row.
- `internal/cli/binary_lag_test.go` — the allowlist entry (a bare identifier for a constant-registered name,
  `namesAddedAfterBaseline`), and the status-set test if the check can emit info.
- `internal/cli/testdata/doctor-{light,dark,nocolor}.golden` — regenerated with `UPDATE_GOLDEN=1`.
- `.github/workflows/release.yml` — a check 8 in `verify-provenance` that runs the script against the tag.

Doctor read path. The Claude version is read from `<config home>/plugins/installed_plugins.json`, key
`moai@moai-adk` (P-08), with no subprocess: the cost is one file read regardless of project size, which the
Advisory-Check Discipline in `coding-standards.md` requires of anything on a latency-sensitive path. The
Codex version is the directory name under `<CODEX_HOME>/plugins/cache/moai-adk/moai/` (P-10). What is
UNOBSERVED: formats at any version other than claude 2.1.287 and codex 0.160.0; any state when `claude` is
absent but a registry file remains (the file is read regardless; absence of the tool is not consulted).
`IsDevBuild` (`pkg/version/version.go`) decides the dev-build case. Severity follows OD-7; the default
never gates a doctor exit on absence.

Release coupling. The generator owns every version field (REQ-015), so a bump is: rewrite the seven
existing stamp files (`.moai/docs/version-management.md`), then run `make plugin-emit`; skipping the second
step turns `plugin-emit-check` red at `make build` and the golden test red in `go test`. The plugin files
carry the version without a leading `v`, so the literal-token registry sweep (`version_stamp_registry_test.go`,
token `v3.1.3`) does not see them and needs no new entry; the drift gate is their guard instead. Release
check 8 closes the tag side: the tagged tree's plugin version must equal the tag.

Exit: AC-013, AC-014, AC-016.

## 4. Verification plan (scratch-home commands only)

Rules for every acceptance command:
- `CLAUDE_CONFIG_DIR` and `CODEX_HOME` point at empty directories created under the session scratchpad
  (`mkdir -p`, then `ls -A` shows empty) before the first command. Placeholders in `acceptance.md`:
  `<empty-claude-home>`, `<empty-codex-home>`.
- Forbidden against any real home, as in t1434: install, uninstall, marketplace add, remove, update, enable,
  disable, configure, `plugin update`, `codex exec`. `claude plugin list`, `validate`, `details` and the
  read-only help verbs are the only commands the run may aim at a real home, and only if a scratch home
  cannot answer; none of the 16 criteria needs it.
- No network. `marketplace add` takes a local path (the repository root in the run worktree). The product
  command `marketplace add modu-ai/moai-adk` is verified only through the injected runner (AC-009) and stays
  UNOBSERVED against GitHub until the manifests are on `main`.
- Go tests use the anchored plain form `-run '^Name$'` (the `VacuousTestAssertion` lint requires it; the
  worktree guard accepts it as a single invocation and refuses it bundled with a redirect and `; echo $?`,
  P-25), and the criterion requires the named `--- PASS: Name (…)` line, because `go test` exits 0 with
  `no tests to run` when a selector matches nothing, which would read as a pass.
- Reporting follows the five-section format of `verification-claim-integrity.md` §3, with the judging
  build's commit next to the tree HEAD (§2.2): `bin/moai` built from the tree, invoked by path.

Gaps this plan names instead of closing (carried into the run's Gaps section):
- G-1: the GitHub path. `claude plugin marketplace add modu-ai/moai-adk` against the public repository is
  not run; `main` has no manifest until the release PR (P-24).
- G-2: runtime behavior of two coexisting copies, and plugin-agent permission behavior. Needs an
  authenticated session; belongs to t1438's measurement.
- G-3: script behavior beyond static checks. `install.sh` ends with `main "$@"` and downloads a binary, so
  it is not run offline; the Go step carries the behavior tests, and the Windows CI job exercises
  `install.bat`.
- G-4: Codex field effect (R14-codex): acceptance by `codex plugin add` is shown, field reading is not.

## 5. Pre-flight and land order (for the run-phase delegation)

Record at run start, before any edit:
- `git rev-parse HEAD` and `git branch --show-current` (a moved HEAD means another writer; stop and report).
- Whether the base already contains t1399: `grep -c moai-factory-foreman internal/template/catalog.yaml`
  (0 at base 7109e0900, so the rename has not landed here). Either order is safe because of REQ-004; if
  t1399 lands after this card, the merge that lands second runs `make plugin-emit` and commits the delta.
- Baseline: `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`, and
  `go test ./internal/cli -run '^TestBinaryLag_.*$'` green before touching doctor.
- Scope of local test runs: only packages the change can affect (`internal/template/pluginemit`,
  `internal/cli` with `-run` selectors, `internal/template` golden and neutrality selectors). No
  `go test ./...` (AGENTS.md §4); CI runs the full suite.
- A doctor-touching card remeasures `-run '^TestRunDiagnosticChecks.*$'`, `-run '^TestDoctorGolden_.*$'` and
  `-run '^TestBinaryLag_.*$'` together (the three guards live apart from the card's own tests).

## 6. Constraints for manager-develop

PRESERVE (do not modify): the deployed file set of `moai init` and `moai update` (this card adds, never
removes; existing init and update deploy tests run unmodified), `.claude/` local copies not mirrored from a
template, `.moai/specs/` other than this SPEC's `progress.md §E.2/§E.3` and the `status:`/`updated:` lines,
`mods/moai-board/**`, the `moai-cowork` files under the user home.
Forbidden commands: `--no-verify`, `--amend`, `git add -A`, `git add .`, force push, any registry-writing
`claude plugin` / `codex plugin` verb against a real home, any network-touching marketplace add in an
acceptance command.
Required: Conventional Commits with the card id in every commit message; stage by explicit pathspec; cite
`worktree-integration-ops.md` form rules when an acceptance command is refused.

## 7. Anti-patterns

- A hand-maintained list of skills, agents or commands anywhere in the generator or the Makefile (the
  failure RK-4 exists to prevent).
- A deny-list payload filter (new scaffold-only directories would leak).
- A doctor check that spawns `claude` or `codex` on the default run (unbounded latency on an advisory path).
- An install step whose default runner is reachable from `go test` (RK-8).
- Copying the plugin version by hand into the manifests (REQ-015).
- Reading the real profile's registry in any test; fixtures are synthetic files under `t.TempDir()`.

## 8. Cross-references

- `.moai/specs/SPEC-PLUGIN-LOAD-SCOPE-001/progress.md §E.2` — the t1434 verdict table (R01 to R14).
- `internal/template/commandemit/commandemit.go`, `internal/template/agentemit/` — emit, golden pin, update switch.
- `Makefile:34-62` — `build` prerequisites, `agents-emit*`, `commands-emit*` precedents.
- `internal/cli/doctor_mcp_version.go` — check shape; `internal/cli/binary_lag_test.go:198` — allowlist.
- `.moai/docs/version-management.md` — version SSOT, stamp list, release process.
- `.claude/rules/moai/development/verification-completeness.md` §2 — two-cell adoption used in `acceptance.md`.
