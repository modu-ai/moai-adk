# Acceptance Criteria — SPEC-PLUGIN-MARKETPLACE-001

This file is the verification layer. Requirements are the `REQ-XXX` entries in `spec.md` §2 (GEARS); each
criterion here is a binary-testable Given/When/Then with the same number.

## Conventions

- **Tree pin.** Document-level pin: base tree `7109e0900` (branch `WT-marketplace-core-plugin`, base `develop`).
  It binds every RED-now cell below that carries no pin of its own. RED cells were measured at that tree and
  are not re-quoted at any later tree without re-measuring.
- **Scratch homes.** `<empty-claude-home>` and `<empty-codex-home>` name empty directories under the session
  scratchpad, created by `mkdir -p` and shown empty by `ls -A` before the first command. Every command that
  writes a registry runs with `CLAUDE_CONFIG_DIR` or `CODEX_HOME` set to one of them. No command here targets
  a real profile and none uses the network (SPEC §3).
- **Two cells per criterion** (`verification-completeness.md` §2): a RED-now cell (the command observed red on
  the base tree, with the reason it is red) and a green-path cell (the milestone that flips it and what the
  output becomes). Where a RED-now command would be vacuous, a positive control that fires on a known-good
  input is named in the ledger.
- **Command form.** Single invocations measured to run in this worktree session (SPEC P-25): anchored
  `go test … -run '^Name$'` (the `VacuousTestAssertion` lint requires the anchors); a Go-test criterion
  passes only when the named `--- PASS: Name (…)` line is present, because `go test` exits 0 with
  `no tests to run` when a selector matches nothing. The worktree guard refuses the same pattern when the
  command is bundled with a redirect and `; echo EXIT=$?`, so exit codes of these commands are read from the
  tool result, not captured in the same invocation.
- **Paths marked (OD-4)** are the default of that decision; the criterion follows the decision.

| AC | Requirement | Milestone | Flips from |
|----|-------------|-----------|------------|
| AC-001 | REQ-001 Claude marketplace manifest | M1 | E-1 |
| AC-002 | REQ-002 Codex marketplace manifest | M1 | E-2 |
| AC-003 | REQ-003 dual plugin manifests, equal version | M1 | E-3 |
| AC-004 | REQ-004 derivation, rename independence | M2 | E-4 |
| AC-005 | REQ-005 copy and render fidelity | M2 | E-5 |
| AC-006 | REQ-006 MCP entry derived | M2 | E-6 |
| AC-007 | REQ-007 scaffold-only exclusions | M2 | E-7 |
| AC-008 | REQ-008 drift gate | M2 | E-8 |
| AC-009 | REQ-009 install sequence and environment | M3 | E-9 |
| AC-010 | REQ-010 fail-open and idempotent | M3 | E-9, E-10 |
| AC-011 | REQ-011 opt-out, self-invocation, test inertness | M3 | E-9, E-11 |
| AC-012 | REQ-012 install scripts | M3 | E-12 |
| AC-013 | REQ-013 doctor Plugin Version check | M4 | E-13 |
| AC-014 | REQ-014 doctor registration and bounded output | M4 | E-13 |
| AC-015 | REQ-015 version derived from the SSOT | M1 | E-14 |
| AC-016 | REQ-016 release tag check | M4 | E-15 |

## Acceptance Criteria

### AC-001 — Claude marketplace manifest (REQ-001)

- **Given** the run's tree after M1, **When** the manifest is validated strictly under an empty scratch home
  and read with jq, **Then** validation passes, the marketplace is named `moai-adk`, and exactly one plugin
  entry named `moai` exists.
- **Verify (a):** `CLAUDE_CONFIG_DIR=<empty-claude-home> claude plugin validate .claude-plugin/marketplace.json --strict`
  — expected exit 0, output contains `✔ Validation passed`.
- **Verify (b):** `jq -c '[.name,[.plugins[].name]]' .claude-plugin/marketplace.json`
  — expected `["moai-adk",["moai"]]`.
- **RED-now:** E-1 — exit 1, `File not found` for the manifest. The reason is absence: the control in E-1
  (same command on a present valid scratch marketplace) exits 0.
- **Green path:** M1 emits the file; (a) prints `✔ Validation passed`.
- **Mutant probe:** deleting `metadata.description` turns (a) red (P-02), and an entry version that differs
  from `plugins/moai/.claude-plugin/plugin.json` turns it red (P-03); the tool, not this text, kills both.

### AC-002 — Codex marketplace manifest (REQ-002)

- **Given** the tree after M1, **When** the repository root is added as a Codex marketplace under an empty
  scratch `CODEX_HOME` and the manifest is read, **Then** the root is accepted and the manifest has the
  moai-cowork shape.
- **Verify (a):** `CODEX_HOME=<empty-codex-home> timeout 60 codex plugin marketplace add .`
  — expected exit 0, first line begins ``Added marketplace `moai-adk` from``.
- **Verify (b):** `jq -c '[.name,[.plugins[].name],.plugins[0].source,.plugins[0].policy]' .agents/plugins/marketplace.json`
  — expected `["moai-adk",["moai"],{"source":"local","path":"./plugins/moai"},{"installation":"AVAILABLE","authentication":"ON_INSTALL"}]` (path per OD-4).
- **Honesty note:** Codex accepts any manifest (R14-codex), so (a) shows the root is recognized and nothing
  about which fields Codex reads; (b) is the shape pin. Field effect stays UNOBSERVED (SPEC RK-5).
- **RED-now:** E-2 — exit 1, `marketplace root does not contain a supported manifest`; the control in E-2
  (a root holding `.agents/plugins/marketplace.json`) exits 0.
- **Green path:** M1 emits `.agents/plugins/marketplace.json`.
- **Mutant probe:** a manifest with `source` as a bare string, or a missing `policy`, passes (a) and fails (b).

### AC-003 — Dual plugin manifests with one version (REQ-003)

- **Given** the tree after M1, **When** the plugin root is validated strictly and both manifests are read,
  **Then** validation passes and both manifests carry the same `name` and `version`.
- **Verify (a):** `CLAUDE_CONFIG_DIR=<empty-claude-home> claude plugin validate plugins/moai --strict`
  — expected exit 0, `✔ Validation passed`.
- **Verify (b):** `jq -c '[.name,.version]' plugins/moai/.claude-plugin/plugin.json plugins/moai/.codex-plugin/plugin.json`
  — expected two identical lines, `["moai","<version>"]`.
- **Verify (c):** `jq -c '[.skills,.mcpServers.moai]' plugins/moai/.codex-plugin/plugin.json`
  — expected `["./skills/",{"command":"moai","args":["mcp-server"]}]`; and
  `jq -c .mcpServers.moai plugins/moai/.codex-plugin/plugin.json internal/template/templates/.mcp.json`
  — expected two identical lines. Codex accepts any manifest (R14-codex), so this pins the shape and says
  nothing about which fields Codex reads (G-4).
- **Third site:** equality with the marketplace entry is checked by AC-001 (a), whose strict validation fails
  on a differing entry version (P-03).
- **RED-now:** E-3 — exit 1, `File not found` for `plugins/moai`; control: a scratch plugin directory with a
  manifest passes strict validate (P-04).
- **Green path:** M1 emits both manifests.
- **Mutant probe:** a Codex manifest whose version differs from the Claude one makes (b) print two different
  lines.

### AC-004 — Derivation from the template tree, independent of names (REQ-004)

- **Given** a synthetic template tree and catalog whose skill and command names differ from the real ones and
  that holds one non-core entry, **When** the generator runs over it, **Then** the emitted skill and command
  set equals exactly the synthetic core set, and the generator source contains no real component name.
- **Verify (a):** `go test ./internal/template/pluginemit -run '^TestEmitDerivesFromTree$' -count=1 -v`
  — expected `--- PASS: TestEmitDerivesFromTree (…)` and ``ok  	github.com/modu-ai/moai-adk/internal/template/pluginemit``.
- **Verify (b):** `grep -rln moai-kanban-foreman internal/template/pluginemit`
  — expected no output, exit 1. Control (must print at least 1): `grep -c moai-kanban-foreman internal/template/catalog.yaml`
  (prints 2 at the base tree; after t1399 lands, the control uses `moai-factory-foreman`).
- **RED-now:** E-4 — `directory not found`, `[setup failed]`, exit 1. This is a tool-failure-class red
  (the package is absent); the run phase's RED record must show the compiled test failing at its assertion
  (`tdd-result-contract.md`).
- **Green path:** M2.
- **Mutant probe:** a generator holding a hand-copied list of the 35 current core names passes against the
  real tree and fails (a); a generator that ignores the tier fails (a) on the synthetic non-core entry.

### AC-005 — Copy and render fidelity (REQ-005)

- **Given** the committed payload after M2, **When** each payload file is compared with its source,
  **Then** every non-`.tmpl` file is byte-identical and the payload commands carry no template syntax.
- **Verify (a):** `go test ./internal/template/pluginemit -run '^TestEmitFidelity$' -count=1 -v`
  — expected `--- PASS: TestEmitFidelity (…)` (every non-`.tmpl` payload file equals its source bytes; every
  `.tmpl` file equals the English-default render with the suffix dropped).
- **Verify (b):** `diff -r internal/template/templates/.claude/skills/moai plugins/moai/skills/moai`
  — expected no output, exit 0.
- **Verify (c):** `grep -rlE '\{\{' plugins/moai/commands` — expected no output, exit 1. Skills are not
  searched: 16 skill files legitimately contain `{{` and are copied verbatim (plan §3 M2).
- **RED-now:** E-5 — `plugins/moai/skills/moai: No such file or directory`, exit 2. Control in E-5: the same
  `grep -rlE` form over the template command sources lists 15 files, so (c) is meaningful.
- **Green path:** M2.
- **Mutant probe:** a generator that renders every file would alter the 16 brace-bearing skill files and fail
  (a); one that copies `.tmpl` files unrendered fails (c).

### AC-006 — MCP entry copied from the template (REQ-006)

- **Given** the payload after M2, **When** the plugin's `.mcp.json` is compared with the template's,
  **Then** the plugin declares only `moai` and its `command` and `args` equal the template entry.
- **Verify (a):** `jq -c .mcpServers.moai plugins/moai/.mcp.json internal/template/templates/.mcp.json`
  — expected two identical lines, `{"command":"moai","args":["mcp-server"]}`.
- **Verify (b):** `jq -c '.mcpServers|keys' plugins/moai/.mcp.json` — expected `["moai"]`.
- **Verify (c):** `go test ./internal/template/pluginemit -run '^TestMCPEntryDerivedFromTemplate$' -count=1 -v`
  — expected `--- PASS: TestMCPEntryDerivedFromTemplate (…)` (changes the entry inside a synthetic template tree
  and asserts the emitted entry follows).
- **RED-now:** E-6 — `Could not open file plugins/moai/.mcp.json`, exit 2; control: the template entry prints
  `{"command":"moai","args":["mcp-server"]}`, exit 0.
- **Green path:** M2.
- **Mutant probe:** a retyped literal passes (a) today and fails (c) as soon as the template entry changes;
  an entry for `context7` fails (b).
- **Stated limit:** Claude's dedupe key was observed only with a fixture that matched on command (SPEC RK-2);
  matching on both command and args is the strict superset.

### AC-007 — Scaffold-only components stay out (REQ-007)

- **Given** the payload after M2, **When** the top level of the plugin root is listed and the allow-list test
  runs, **Then** only the allowed entries exist.
- **Verify (a):** `find plugins/moai -maxdepth 1 -mindepth 1 -print`
  — expected exactly five lines, order unspecified: `plugins/moai/.claude-plugin`, `plugins/moai/.codex-plugin`,
  `plugins/moai/.mcp.json`, `plugins/moai/commands`, `plugins/moai/skills` (default of OD-2 and OD-3; with
  OD-3 option (a) a sixth, `plugins/moai/agents`, is the only addition).
- **Verify (b):** `go test ./internal/template/pluginemit -run '^TestPayloadAllowList$' -count=1 -v`
  — expected `--- PASS: TestPayloadAllowList (…)` (a synthetic template tree gains `rules`, `hooks`,
  `output-styles`, `workflows`, `CLAUDE.md`; none is emitted).
- **RED-now:** E-7 — `bfs: error: plugins/moai: No such file or directory`, exit 1. An absent directory is
  also what makes this read red, so the green path is the five-line listing, and the control in E-7 shows the
  same `find` form lists the template `.claude` entries (it would also list `rules` and `hooks` there).
- **Green path:** M2.
- **Mutant probe:** an allow-list implemented as a deny-list of `rules` and `hooks` passes today's tree and
  fails (b) on the synthetic `output-styles` entry.

### AC-008 — Read-only drift gate (REQ-008)

- **Given** the tree after M2, **When** the check runs on a clean tree and on mutated artifact sets,
  **Then** it passes on the clean tree, fails on each mutation, and is a prerequisite of `build`.
- **Verify (a):** `make plugin-emit-check` — expected exit 0, `ok` for the pluginemit package.
- **Verify (b):** `go test ./internal/template/pluginemit -run '^TestDriftDetectsMutatedArtifact$' -count=1 -v`
  — expected `--- PASS: TestDriftDetectsMutatedArtifact (…)` with three passing subtests (flipped byte, deleted
  file, extra file).
- **Verify (c):** `grep -n ^build: Makefile` — expected one line that contains `plugin-emit-check`.
- **RED-now:** E-8 — ``make: *** No rule to make target `plugin-emit-check'.  Stop.``, exit 2.
- **Green path:** M1 adds the verbs, M2 extends the check to the payload and adds it to `build:`.
- **Mutant probe:** a check that regenerates before comparing passes (a) and fails nothing in (b); the
  subtest asserting the committed set is left byte-unchanged after a failing check catches it.

### AC-009 — Install sequence and environment (REQ-009)

- **Given** an injected runner and a PATH stub, **When** the step runs after a deployment, **Then** for each
  tool on PATH it runs add then install in that order, never runs install after a failed add, passes the
  environment through unchanged, and prints the config home.
- **Verify:** `go test ./internal/cli -run '^TestPluginInstallStep_Sequence$' -count=1 -v`
  — expected `--- PASS: TestPluginInstallStep_Sequence (…)` and passing subtests `both-present`, `claude-only`,
  `codex-only`, `neither`, `install-skipped-after-add-fails`, `env-unchanged`, `pinned-binary-honored`.
  The recorded argument vectors are exactly `claude plugin marketplace add modu-ai/moai-adk`,
  `claude plugin install moai@moai-adk`, `codex plugin marketplace add modu-ai/moai-adk`,
  `codex plugin add moai@moai-adk`; the runner's `CLAUDE_CONFIG_DIR` and `CODEX_HOME` equal the parent's.
- **RED-now:** E-9 — the `-list` form prints no `TestPluginInstallStep` name and exits 0, which is an empty
  sweep and why the criterion requires the named PASS lines; the control in E-9 lists four `TestBinaryLag_`
  names with the same form.
- **Green path:** M3.
- **Mutant probe:** swapping the two commands, running install after a failed add, or exporting a changed
  `CLAUDE_CONFIG_DIR` each fails one named subtest.
- **Stated limit:** the GitHub source `modu-ai/moai-adk` is never contacted (G-1).

### AC-010 — Fail-open and idempotent (REQ-010)

- **Given** a runner that fails, times out, or reports already-present, and the real tools under an empty
  scratch home, **When** the step runs twice, **Then** no failure changes init's exit status, guidance names
  both manual commands, and a repeated run succeeds.
- **Verify (a):** `go test ./internal/cli -run '^TestPluginInstallStep_FailOpen$' -count=1 -v` — expected
  `--- PASS: TestPluginInstallStep_FailOpen (…)` with subtests `exit-nonzero`, `timeout`, `tool-absent`,
  `invalid-pin`, `already-present`; the first four return no error and print guidance (tool-absent prints one
  skip line), `already-present` prints none.
- **Verify (b):** run twice each, under `<empty-claude-home>` with the repository root as the source (a local
  path stands in for the GitHub source, which AC forbids contacting):
  `CLAUDE_CONFIG_DIR=<empty-claude-home> claude plugin marketplace add ./ --json` then
  `CLAUDE_CONFIG_DIR=<empty-claude-home> claude plugin install moai@moai-adk --json`
  — expected exit 0 on all four runs; first add `"outcome":"ok"` with `Successfully added marketplace: moai-adk`;
  second add `already on disk`; first install `Successfully installed plugin: moai@moai-adk`; second install
  contains `is already installed`. (b) also proves the committed payload installs.
- **RED-now:** E-9 for (a); for (b) E-10 — `"failureCode":"manifest_missing"`, exit 1 at the base tree; the
  control (E-10) shows `./` adds a present marketplace with exit 0.
- **Green path:** M3 for (a); (b) needs M1 and M2 as well.
- **Mutant probe:** a step that treats `already installed` as failure prints guidance in `already-present`
  and fails it; a step that returns the tool's exit code fails the first four subtests.

### AC-011 — Opt-out, self-invocation, test inertness (REQ-011)

- **Given** the opt-out set, moai's own init re-entry, and the test binary, **When** the step is reached,
  **Then** no external command runs.
- **Verify (a):** `go test ./internal/cli -run '^TestPluginInstallStep_OptOut$' -count=1 -v` — expected
  `--- PASS: TestPluginInstallStep_OptOut (…)` (flag `--no-plugin` and env `MOAI_SKIP_PLUGIN_INSTALL` set to `1`
  or `true` each give zero runner calls; an empty or `0` value does not opt out).
- **Verify (b):** `go test ./internal/cli -run '^TestDoctorAgentEmitEmbed_SetsPluginOptOut$' -count=1 -v` —
  expected `--- PASS`; the captured environment of the `init --non-interactive --llm both` child
  (`internal/cli/doctor_agentemit_embed.go:352`) contains the opt-out.
- **Verify (c):** `go test ./internal/cli -run '^TestPluginInstallStep_NoRealRunnerUnderTest$' -count=1 -v` —
  expected `--- PASS`; the default runner called from a test binary performs no exec and returns a refusal,
  and every test that calls `runInit` (five sites, SPEC P-20) installs a fake runner or the opt-out.
- **RED-now:** E-9 for (a) to (c) (no such tests listed); E-11 — `grep -c MOAI_SKIP_PLUGIN_INSTALL internal/config/envkeys.go`
  prints `0`, exit 1, so the constant does not exist yet.
- **Green path:** M3.
- **Mutant probe:** a doctor child that does not set the opt-out fails (b); a default runner that execs
  fails (c).

### AC-012 — Install scripts (REQ-012)

- **Given** the three scripts and the two docs-site copies after M3, **When** they are checked statically,
  **Then** each carries the opt-out and the step call, shell syntax is valid, and the copies are identical.
- **Verify:**
  `bash -n install.sh` (exit 0);
  `grep -c MOAI_SKIP_PLUGIN_INSTALL install.sh` and the same on `install.ps1`, `install.bat`,
  `docs-site/static/install.sh`, `docs-site/static/install.ps1` (each at least 2: help text and the guard);
  `cmp install.sh docs-site/static/install.sh` (exit 0, no output);
  `cmp install.ps1 docs-site/static/install.ps1` (exit 0, no output).
- **RED-now:** E-12 — `grep -c MOAI_SKIP_PLUGIN_INSTALL install.sh` prints `0`, exit 1; control
  `grep -c print_info install.sh` prints `14`, so the form counts tokens in this file.
- **Green path:** M3.
- **Stated limit (G-3):** an end-to-end script run needs a download and is not run offline; behavior lives in
  the Go step (AC-009 to AC-011) and the Windows job of `test-install.yml` exercises `install.bat`.
- **Mutant probe:** editing only the root `install.sh` fails the `cmp` line.

### AC-013 — Doctor Plugin Version check (REQ-013)

- **Given** synthetic config homes and an injected binary version, **When** the check runs, **Then** equality
  is OK, inequality takes the OD-7 severity (default warn) with both versions and the remedy, and every
  indeterminate case is OK or info.
- **Verify (a):** `go test ./internal/cli -run '^TestCheckPluginVersion$' -count=1 -v` — expected `--- PASS` for
  `TestCheckPluginVersion` subtests `equal`, `v-prefix-normalized`, `prerelease-equal`, `mismatch`,
  `not-installed`, `home-absent`, `registry-malformed`, `registry-unknown-shape`, `dev-build`,
  `codex-cache-version`, `claude-config-dir-honored` (a registry under `$CLAUDE_CONFIG_DIR`, never `~/.claude`).
- **Verify (b):** `CLAUDE_CONFIG_DIR=<empty-claude-home> bin/moai doctor --check "Plugin Version"` —
  expected exit 0, output contains `Plugin Version` and `not installed`. `bin/moai` is built from the tree and
  invoked by path (VCI §2.2).
- **RED-now:** E-13 — the `-list` form prints no `TestCheckPluginVersion` name, exit 0 (empty sweep, hence the
  named PASS lines); the control is the `TestBinaryLag_` listing in E-9.
- **Green path:** M4.
- **Mutant probe:** a check that spawns `claude` on the default run is caught by the `home-absent` subtest
  running with a PATH that holds no `claude`; a check that returns fail on a malformed registry fails
  `registry-malformed`.
- **UNOBSERVED:** registry and cache formats at versions other than claude 2.1.287 and codex 0.160.0.

### AC-014 — Doctor registration and bounded output (REQ-014)

- **Given** the new check name, **When** the guarded doctor tests run, **Then** they pass with the check
  registered, and the default run adds one line.
- **Verify:** `go test ./internal/cli -run '^TestBinaryLag_.*$' -count=1 -v` (all `TestBinaryLag_` tests pass,
  including `TestBinaryLag_DoctorCheckNameSetIsUnchanged` with the new name in `namesAddedAfterBaseline`);
  `go test ./internal/cli -run '^TestRunDiagnosticChecks.*$' -count=1` (`ok`);
  `go test ./internal/cli -run '^TestDoctorGolden_.*$' -count=1` (`ok`, goldens regenerated);
  `go test ./internal/cli -run '^TestCheckPluginVersion_OutputBounded$' -count=1 -v` (`--- PASS`: default run
  emits one summary line and an empty detail; `--verbose` adds detail).
- **RED-now:** E-13 (no such tests); the three guard families exist and pass today, shown by the control in E-9.
- **Green path:** M4.
- **Mutant probe:** registering the check without the allowlist entry fails
  `TestBinaryLag_DoctorCheckNameSetIsUnchanged` (lesson of cards t1251 and t1282).

### AC-015 — Versions derived from the SSOT (REQ-015)

- **Given** a synthetic version and the committed manifests, **When** the generator runs, **Then** every
  version field it writes equals the SSOT value without a leading `v`.
- **Verify (a):** `go test ./internal/template/pluginemit -run '^TestVersionStampedFromSSOT$' -count=1 -v`
  — expected `--- PASS: TestVersionStampedFromSSOT (…)` (sets `version.Version` to `v9.8.7-rc.1`; all four
  manifests carry `9.8.7-rc.1`, none carries a leading `v`).
- **Verify (b):** `go test ./internal/template/pluginemit -run '^TestCommittedVersionMatchesSSOT$' -count=1 -v`
  — expected `--- PASS` (committed manifests equal the fallback minus `v`).
- **RED-now:** E-14 — setup failed, exit 1; control `grep -n 'Version = "v' pkg/version/version.go` prints
  `11:	Version = "v3.1.3"`, the SSOT value the committed manifests will carry as `3.1.3`.
- **Green path:** M1.
- **Mutant probe:** a hard-coded version literal passes (b) until a bump and fails (a) immediately.

### AC-016 — Release tag check (REQ-016)

- **Given** the script and the committed plugin version, **When** it is run with a matching and a
  non-matching tag, **Then** it exits 0 for the match and 1, naming both values, for the mismatch; and the
  release workflow runs it.
- **Verify (a):** `sh scripts/check-plugin-version.sh v3.1.3` — expected exit 0 (pinned to the tree's
  fallback `v3.1.3`; re-pin when a bump lands).
- **Verify (b):** `sh scripts/check-plugin-version.sh v9.9.9` — expected exit 1, output names `9.9.9` and
  the committed plugin version. The mismatching tag is the known failing input; no file is mutated.
- **Verify (c):** `grep -n check-plugin-version .github/workflows/release.yml` — expected at least one line
  inside the `verify-provenance` job.
- **RED-now:** E-15 — `sh: scripts/check-plugin-version.sh: No such file or directory`, exit 127.
- **Green path:** M4.
- **Mutant probe:** a script that prints the mismatch and exits 0 passes (a) and fails (b); a workflow that
  does not call it fails (c).

## Edge Cases

- Neither tool on PATH: one skip line, `moai init` exit 0, no guidance block.
- `llm.claude_bin` pin invalid or `MOAI_CLAUDE_BIN` pointing nowhere: the resolver errors; the step prints
  guidance and returns success (AC-010 `invalid-pin`).
- Config home missing or unwritable: the tool's own failure lands in the guidance path; the doctor check
  reports OK with a not-installed message.
- Binary built with `VERSION=v3.2.0-rc.N` against a plugin stamped from the fallback: a mismatch is reported
  at the OD-7 severity; a development build (`IsDevBuild`) is never compared.
- A registry whose `moai@moai-adk` array holds several scopes: the check reads the user-scope entry first and
  states which scope it read.
- `--non-interactive` init (no project `.mcp.json` entry): the step still runs unless opted out (SPEC §1.5-2,
  OD-1, OD-5).

## Quality Gates

- Tested: `go test -cover` at least 85% for `internal/template/pluginemit`, and for the new files of
  `internal/cli` (`plugin_install.go`, `doctor_plugin_version.go`).
- Readable and Unified: `go vet` and the CI-pinned golangci-lint over the changed packages; `gofmt` clean.
- Secured: child processes are started with an argument vector, never through a shell string; no value of the
  environment is printed except the two config-home paths; registry JSON is parsed defensively.
- Trackable: the card id `t1435` is in every commit message on the branch.
- Cross-platform: `GOOS=windows GOARCH=amd64 go build ./...` exits 0.
- Preservation: existing init and update deployment tests pass unmodified (the deployed file set is unchanged).
- Lifecycle lint: `moai spec lint` reports no error on this SPEC directory.

## Definition of Done

All sixteen criteria pass with the commands above, each reported in the five-section format with the judging
build's commit next to the tree HEAD; the Gaps section names G-1 to G-4 (plan §4) and any refusal of an
acceptance command by the worktree guard; the Kickoff gate has a verdict for every row of `decision-index.md`
or records the default each row fell back to; `plan-auditor` verdict is PASS at the Tier M threshold.

## Evidence Ledger (RED-now cells)

Every entry: tree `7109e0900`. Commands ran from the worktree root. `<empty-claude-home>` and
`<empty-codex-home>` were empty directories under the session scratchpad.

```
E-1  AC-001
cmd:    CLAUDE_CONFIG_DIR=<empty-claude-home> claude plugin validate .claude-plugin/marketplace.json --strict
exit:   1
stdout: Validating marketplace manifest: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/.claude-plugin/marketplace.json
        
        ✘ Found 1 error:
        
          ❯ file: File not found: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/.claude-plugin/marketplace.json
        
        ✘ Validation failed
control: same command (file-path form) on a present, internally consistent scratch marketplace -> "✔ Validation passed", exit 0
```

```
E-2  AC-002
cmd:    CODEX_HOME=<empty-codex-home> timeout 60 codex plugin marketplace add .
exit:   1
stdout: Error: invalid marketplace file `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435`: marketplace root does not contain a supported manifest
control: a root holding .agents/plugins/marketplace.json -> "Added marketplace `mini-mkt` from <root>.", exit 0
```

```
E-3  AC-003
cmd:    CLAUDE_CONFIG_DIR=<empty-claude-home> claude plugin validate plugins/moai --strict
exit:   1
stdout: Validating plugin manifest: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/plugins/moai
        
        ✘ Found 1 error:
        
          ❯ file: File not found: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/plugins/moai
        
        ✘ Validation failed
control: strict validate of a scratch plugin directory with a manifest (version 3.2.0-rc.23) -> "✔ Validation passed", exit 0
```

```
E-4  AC-004 (and the AC-015 package)
cmd:    go test ./internal/template/pluginemit -run '^TestEmitDerivesFromTree$' -count=1 -v
exit:   1
stdout: # ./internal/template/pluginemit
        stat /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/internal/template/pluginemit: directory not found
        FAIL	./internal/template/pluginemit [setup failed]
        FAIL
note:   measured as a plain single invocation; the same pattern bundled with a redirect and `; echo EXIT=$?` was refused by the worktree guard
```

```
E-5  AC-005
cmd:    diff -r internal/template/templates/.claude/skills/moai plugins/moai/skills/moai
exit:   2
stdout: diff: plugins/moai/skills/moai: No such file or directory
control: grep -rlE '\{\{' internal/template/templates/.claude/commands -> 15 lines, each internal/template/templates/.claude/commands/moai/<name>.md.tmpl, exit 0
```

```
E-6  AC-006
cmd:    jq -c .mcpServers plugins/moai/.mcp.json
exit:   2
stdout: jq: error: Could not open file plugins/moai/.mcp.json: No such file or directory
control: jq -c .mcpServers.moai internal/template/templates/.mcp.json -> {"command":"moai","args":["mcp-server"]}, exit 0
```

```
E-7  AC-007
cmd:    find plugins/moai -maxdepth 1 -mindepth 1 -print
exit:   1
stdout: bfs: error: plugins/moai: No such file or directory.
control: find internal/template/templates/.claude -maxdepth 1 -mindepth 1 -print -> 9 entries (loop.md, output-styles, workflows, agents, settings.json.tmpl, hooks, rules, commands, skills), exit 0
```

```
E-8  AC-008
cmd:    make plugin-emit-check
exit:   2
stdout: make: *** No rule to make target `plugin-emit-check'.  Stop.
```

```
E-9  AC-009, AC-010 (a), AC-011
cmd:    go test ./internal/cli -list TestPluginInstallStep
exit:   0
stdout: ok  	github.com/modu-ai/moai-adk/internal/cli	1.565s
note:   zero test names listed: an empty sweep, which is why every Go-test criterion requires the named PASS line
control: go test ./internal/cli -list TestBinaryLag_ ->
        TestBinaryLag_OneSeamServesBothSurfaces
        TestBinaryLag_NonGitDirectoryKeepsDoctorExitZero
        TestBinaryLag_AllowlistKeysAreLiveNames
        TestBinaryLag_DoctorCheckNameSetIsUnchanged
        ok  	github.com/modu-ai/moai-adk/internal/cli	1.248s
```

```
E-10  AC-010 (b)
cmd:    CLAUDE_CONFIG_DIR=<empty-claude-home> claude plugin marketplace add ./ --json
exit:   1
stdout: {"command":"marketplace-add","outcome":"failed","message":"Marketplace file not found at /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/.claude-plugin/marketplace.json","failureCode":"manifest_missing"}
note:   the form `add .` is rejected as invalid_source ("Try: owner/repo, https://..., or ./path"), so the criterion uses `./`
control: the same `add ./` run from a directory holding a valid marketplace -> {"command":"marketplace-add","outcome":"ok",...}, exit 0
```

```
E-11  AC-011
cmd:    grep -c MOAI_SKIP_PLUGIN_INSTALL internal/config/envkeys.go
exit:   1
stdout: 0
```

```
E-12  AC-012
cmd:    grep -c MOAI_SKIP_PLUGIN_INSTALL install.sh
exit:   1
stdout: 0
control: grep -c print_info install.sh -> 14, exit 0
```

```
E-13  AC-013, AC-014
cmd:    go test ./internal/cli -list TestCheckPluginVersion
exit:   0
stdout: ok  	github.com/modu-ai/moai-adk/internal/cli	1.364s
note:   zero names listed (empty sweep); control is the TestBinaryLag_ listing in E-9
```

```
E-14  AC-015
cmd:    go test ./internal/template/pluginemit -run '^TestVersionStampedFromSSOT$' -count=1 -v
exit:   1
stdout: # ./internal/template/pluginemit
        stat /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/internal/template/pluginemit: directory not found
        FAIL	./internal/template/pluginemit [setup failed]
        FAIL
control: grep -n 'Version = "v' pkg/version/version.go -> 11:	Version = "v3.1.3", exit 0
```

```
E-15  AC-016
cmd:    sh scripts/check-plugin-version.sh v3.1.3
exit:   127
stdout: sh: scripts/check-plugin-version.sh: No such file or directory
```

Gaps in this ledger: E-4 and E-14 are package-absent reds (tool-failure class); the first run-phase RED record
must show an assertion failure. The `-list` reds (E-9, E-13) read zero names and exit 0, so they establish
absence of the tests and nothing about whether a runner would later select them; the named PASS line is what
closes that. The scratch marketplace used for controls was built by the author in the scratchpad and is not
part of the repository.
