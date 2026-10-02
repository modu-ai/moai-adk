# Acceptance Criteria — SPEC-PLUGIN-MARKETPLACE-001

This file is the verification layer. Requirements are the `REQ-XXX` entries in `spec.md` §2 (GEARS); each
criterion here is a binary-testable Given/When/Then with the same number.

## Conventions

- **Tree pin.** Entries L-01, L-03, L-04, L-05, L-07, L-10, L-11, L-12, L-15, L-20 and L-24 are carried from the
  iteration-0 ledger (E-1, E-3, E-14, E-4, E-5, E-6, E-7, E-8, E-10, E-11, E-12), measured at base tree
  `7109e0900`, and were re-executed unchanged at `3766cef05` by plan-audit iteration 1 (`plan-audit.md` §A: every
  row "matches ledger: yes") and again by this revision, whose outputs matched each entry's stdout and exit. The two
  trees differ in the SPEC artifacts only (`git diff --name-only 7109e0900 3766cef05`
  lists exactly the five files of this directory, re-run in this revision). Every other entry was measured in this
  revision at tree `3766cef05` and carries that pin. A pin is never re-quoted at a later tree without re-measuring.
- **Scratch homes, one per criterion.** `<claude-home:AC-nnn>` and `<codex-home:AC-nnn>` name empty directories
  under the session scratchpad, created by `mkdir -p` and shown empty by `ls -A` before the first command, **for
  that criterion alone**: the same placeholder never appears under two criteria, so no criterion reads state that
  another one wrote. Within one criterion a placeholder names one directory, and a sequence of commands that
  share a home does so on purpose. A command that can read a Claude or Codex home sets both. No command sets
  `HOME` (the worktree guard refuses it, SPEC P-36); the harness scripts set `HOME`, `PATH` and both homes inside
  themselves. No command here targets a real profile and none uses the network (SPEC §3).
- **Two cells per criterion** (`verification-completeness.md` §2): a RED-now cell (the criterion's own command observed
  red, with the reason it is red, in the Evidence Ledger) and a green-path cell naming the milestone that flips it
  and what the output becomes. Where a RED-now command would be vacuous, a positive control that fires on a known-good
  input is named in the ledger.
- **Command form.** Single invocations measured to run in this worktree session (SPEC P-25): anchored
  `go test … -run '^Name$'` (the `VacuousTestAssertion` lint requires the anchors). A Go-test criterion passes only
  when the named `--- PASS: Name (…)` line is present, because `go test` exits 0 with `no tests to run` when a selector
  matches nothing; every Go-test RED-now cell below therefore quotes the criterion's own command and its
  `[no tests to run]` line, and is red by the PASS-line rule, not by the exit code. The worktree guard refuses the same
  pattern when the command is bundled with a redirect and `; echo EXIT=$?`, so exit codes are read from the tool result.
  A `grep` that selects no line, and a `find` that cannot open its root, are shown by the tool as completed without
  echoing their status; the ledger records the status each documents for that case (1) and marks it as not echoed.
- **Harness scripts.** Two criteria cannot be answered by a Go test: the default runner of the install step refuses to
  exec under a test binary (REQ-017), so the real exec path and the installers are driven by
  `scripts/test-plugin-install-step.sh <path-to-moai>`, which prints `PASS <name>` or `FAIL <name>: <reason>` per case
  and a final `RESULT pass=<n> fail=<m>` line. A criterion cites the case names it needs. `bin/moai` is built from the
  tree and invoked by path (`verification-claim-integrity.md` §2.2). The script builds its own scrubbed `PATH`, stub
  `claude` and `codex` that record their argv, a stub `curl` serving a pinned local archive, and scratch `HOME`,
  `CLAUDE_CONFIG_DIR` and `CODEX_HOME`; it counts only recorded calls whose first argument is `plugin`. Its case
  `harness-precondition` fails when `claude` or `codex` resolves outside the stubs.
- **Static checks are labelled static.** A check that reads text (a grep, a source scan) proves a token or shape is
  present, not that behavior holds; where a criterion carries one it says so, and a behavior check sits beside it.
- **Defaults.** A criterion bound to an Open Decision carries `default pending OD-n` and an `Alternate` line stating
  how it changes if the verdict differs (`spec.md` §5 marker table).

| AC | Requirement | Milestone | RED-now entry |
|----|-------------|-----------|---------------|
| AC-001 | REQ-001 Claude marketplace manifest | M1 | L-01 |
| AC-002 | REQ-002 Codex marketplace manifest, no version field | M1 | L-02 |
| AC-003 | REQ-003 dual plugin manifests, version from the SSOT | M1 | L-03, L-04 |
| AC-004 | REQ-004 derivation, rename independence | M2 | L-05, L-06 |
| AC-005 | REQ-005 copy and render fidelity | M2 | L-07 |
| AC-006 | REQ-006 flat command layout, discoverable inventory | M2 | L-08, L-09 |
| AC-007 | REQ-007 MCP entry derived | M2 | L-10 |
| AC-008 | REQ-008 scaffold-only exclusions | M2 | L-11 |
| AC-009 | REQ-009 drift gate | M2 | L-12 |
| AC-010 | REQ-010 init call and harness gating | M3 | L-13 |
| AC-011 | REQ-011 sequence, idempotence, scope | M3 | L-14, L-15 |
| AC-012 | REQ-012 environment and config home | M3 | L-16 |
| AC-013 | REQ-013 failure guidance and bound | M3 | L-17 |
| AC-014 | REQ-014 absent tool, one skip line | M3 | L-18 |
| AC-015 | REQ-015 opt-out | M3 | L-19, L-20 |
| AC-016 | REQ-016 automated callers set the opt-out | M3 | L-21 |
| AC-017 | REQ-017 test-binary inertness | M3 | L-22 |
| AC-018 | REQ-018 install scripts | M3 | L-23, L-24, L-25, L-26 |
| AC-019 | REQ-019 the install verb | M3 | L-27, L-28 |
| AC-020 | REQ-020 doctor Claude read | M4 | L-29 |
| AC-021 | REQ-021 doctor Codex read | M4 | L-30 |
| AC-022 | REQ-022 doctor comparison and outcomes | M4 | L-31 |
| AC-023 | REQ-023 doctor registration and bounded output | M4 | L-32, L-33 |
| AC-024 | REQ-024 release tag check and runbook | M4 | L-34, L-35, L-36 |

## Acceptance Criteria

### AC-001 — Claude marketplace manifest (REQ-001)

- **Given** the run's tree after M1, **When** the manifest is validated strictly under an empty scratch home
  and read with jq, **Then** validation passes, the marketplace is named `moai-adk`, exactly one plugin entry
  named `moai` exists with `source` `./plugins/moai`, and `metadata.version` equals the entry version.
- **Verify (a):** `CLAUDE_CONFIG_DIR=<claude-home:AC-001> claude plugin validate .claude-plugin/marketplace.json --strict`
  — expected exit 0, output contains `✔ Validation passed`.
- **Verify (b):** `jq -c '[.name,[.plugins[].name],.plugins[0].source]' .claude-plugin/marketplace.json`
  — expected `["moai-adk",["moai"],"./plugins/moai"]` (path default pending OD-4).
- **Verify (c):** `jq -e '.metadata.version == .plugins[0].version' .claude-plugin/marketplace.json`
  — expected output `true`, exit 0.
- **RED-now:** L-01 — exit 1, `File not found` for the manifest. The reason is absence: the control (same
  command on a present valid scratch marketplace) exits 0.
- **Green path:** M1 emits the file; (a) prints `✔ Validation passed`.
- **Alternate (OD-4 b):** the entry gains a `ref`, and (b) prints it; (c) is unchanged. **(OD-4 c):** the tree is not
  committed, and (a) runs against the release-time build.
- **Mutant probe:** deleting `metadata.description` turns (a) red (P-02); an entry version that differs from
  `plugins/moai/.claude-plugin/plugin.json` turns (a) red (P-03); a `metadata.version` that differs from the entry
  turns (c) red whether or not the tool checks it. The tool and jq, not this text, kill each.

### AC-002 — Codex marketplace manifest, no version field (REQ-002)

- **Given** the tree after M1, **When** the repository root is added as a Codex marketplace under an empty
  scratch `CODEX_HOME` and the manifest is read, **Then** the root is accepted, the manifest has the moai-cowork
  shape, and neither the marketplace nor the entry carries a `version`.
- **Verify (a):** `CODEX_HOME=<codex-home:AC-002> codex plugin marketplace add .`
  — expected exit 0, first line begins ``Added marketplace `moai-adk` from``. No `timeout` wrapper is used (stock
  macOS has none; the tool call's own bound applies).
- **Verify (b):** `jq -c '[.name,[.plugins[].name],.plugins[0].source,.plugins[0].policy]' .agents/plugins/marketplace.json`
  — expected `["moai-adk",["moai"],{"source":"local","path":"./plugins/moai"},{"installation":"AVAILABLE","authentication":"ON_INSTALL"}]` (path default pending OD-4).
- **Verify (c):** `jq -e '[has("version"), (.plugins[0] | has("version"))] == [false, false]' .agents/plugins/marketplace.json`
  — expected output `true`, exit 0. The Codex marketplace shape has no version field (P-35); the Codex plugin
  version is read from `.codex-plugin/plugin.json` (AC-003).
- **Honesty note:** Codex accepts any manifest (R14-codex), so (a) shows the root is recognized and nothing about
  which fields Codex reads; (b) and (c) are the shape pin. Field effect stays UNOBSERVED (SPEC RK-5, G-4).
- **RED-now:** L-02 — exit 1, `marketplace root does not contain a supported manifest`; the control (a root
  holding `.agents/plugins/marketplace.json`) exits 0.
- **Green path:** M1 emits `.agents/plugins/marketplace.json`.
- **Mutant probe:** a manifest with `source` as a bare string, or a missing `policy`, passes (a) and fails (b); an
  invented `version` key passes (a) and (b) and fails (c).

### AC-003 — Dual plugin manifests with one version from the SSOT (REQ-003)

- **Given** the tree after M1, **When** the plugin root is validated strictly, both manifests are read, and the
  generator runs with a synthetic version, **Then** validation passes, both manifests carry the same `name` and
  `version`, and all four version-carrying fields follow the SSOT without a leading `v`.
- **Verify (a):** `CLAUDE_CONFIG_DIR=<claude-home:AC-003> claude plugin validate plugins/moai --strict`
  — expected exit 0, `✔ Validation passed`.
- **Verify (b):** `jq -c '[.name,.version]' plugins/moai/.claude-plugin/plugin.json plugins/moai/.codex-plugin/plugin.json`
  — expected two identical lines, `["moai","<version>"]`.
- **Verify (c):** `jq -c '[.skills,.mcpServers.moai]' plugins/moai/.codex-plugin/plugin.json`
  — expected `["./skills/",{"command":"moai","args":["mcp-server"]}]` (entry default pending OD-2); and
  `jq -c .mcpServers.moai plugins/moai/.codex-plugin/plugin.json internal/template/templates/.mcp.json`
  — expected two identical lines. This pins the shape and says nothing about which fields Codex reads (G-4).
- **Verify (d):** `go test ./internal/template/pluginemit -run '^TestVersionStampedFromSSOT$' -count=1 -v`
  — expected `--- PASS: TestVersionStampedFromSSOT (…)`: with `version.Version` set to `v9.8.7-rc.1`, the Claude
  marketplace `metadata.version`, the entry `version`, the Claude plugin `version` and the Codex plugin `version`
  all carry `9.8.7-rc.1`, none carries a leading `v`, and the Codex marketplace carries no version field.
- **Verify (e):** `go test ./internal/template/pluginemit -run '^TestCommittedVersionMatchesSSOT$' -count=1 -v`
  — expected `--- PASS` (the committed manifests equal the fallback value minus its `v`).
- **Third site:** equality with the marketplace entry is also checked by AC-001 (a), whose strict validation fails on
  a differing entry version (P-03).
- **RED-now:** L-03 for (a) — exit 1, `File not found` for `plugins/moai`; L-04 for (d) — setup failed, exit 1.
- **Green path:** M1 emits both manifests and the generator derives the versions.
- **Alternate (OD-2 b):** (c) drops the `mcpServers` element. **(OD-4):** paths follow the decision.
- **Mutant probe:** a Codex manifest whose version differs from the Claude one makes (b) print two different
  lines; a hard-coded version literal passes (e) until a bump and fails (d) immediately.

### AC-004 — Derivation from the template tree, independent of names (REQ-004)

- **Given** a synthetic template tree and catalog whose skill and command names differ from the real ones and
  that holds one non-core entry, **When** the generator runs over it, **Then** the emitted skill and command
  set equals exactly the synthetic core set, and the generator source contains no component name taken from the
  real catalog or the real command set.
- **Verify (a):** `go test ./internal/template/pluginemit -run '^TestEmitDerivesFromTree$' -count=1 -v`
  — expected `--- PASS: TestEmitDerivesFromTree (…)` and ``ok  	github.com/modu-ai/moai-adk/internal/template/pluginemit``.
- **Verify (b):** `go test ./internal/template/pluginemit -run '^TestGeneratorHoldsNoComponentNames$' -count=1 -v`
  — expected `--- PASS`. The test reads every `name:` of `internal/template/catalog.yaml` and every command stem under
  `internal/template/templates/.claude/commands/moai/`, scans the non-test sources of the generator package for each
  as a whole token, and fails on any hit; it first scans a fixture string holding one of the names and requires the
  hit (a positive control inside the test), so an empty name set or a broken scan fails rather than passes. It
  replaces the iteration-0 grep for one literal name, which went vacuous when t1399 renames that name.
- **RED-now:** L-05 — `directory not found`, `[setup failed]`, exit 1; L-06 the same for (b). Both are
  tool-failure-class reds (the package is absent); the run phase's RED record must show the compiled test failing
  at its assertion (`tdd-result-contract.md`).
- **Green path:** M2.
- **Alternate (OD-3 a/b):** the synthetic tree gains an agents directory and (a) asserts the agent set; **(OD-8 b/c):**
  (a) asserts the tier the decision names.
- **Mutant probe:** a generator holding a hand-copied list of the 24 current core skill names passes against the
  real tree and fails (a) and (b); a generator that ignores the tier fails (a) on the synthetic non-core entry.

### AC-005 — Copy and render fidelity (REQ-005)

- **Given** the committed payload after M2, **When** each payload file is compared with its source,
  **Then** every non-`.tmpl` file is byte-identical and no rendered payload file carries a template action.
- **Verify (a):** `go test ./internal/template/pluginemit -run '^TestEmitFidelity$' -count=1 -v`
  — expected `--- PASS: TestEmitFidelity (…)` (every non-`.tmpl` payload file equals its source bytes, the 16
  brace-bearing skill files included; every `.tmpl` file equals the English-default render with the suffix dropped).
- **Verify (b):** `diff -r internal/template/templates/.claude/skills/moai plugins/moai/skills/moai`
  — expected no output, exit 0.
- **Verify (c):** `grep -rlE '\{\{' plugins/moai/commands` — expected no output, exit 1. Only the rendered files are
  searched; the skills are not, because 16 skill files legitimately contain `{{` and are copied verbatim (plan §3 M2).
- **RED-now:** L-07 — `plugins/moai/skills/moai: No such file or directory`, exit 2. Control in L-07: the same
  `grep -rlE` form over the template command sources lists 15 files, so (c) is meaningful.
- **Green path:** M2.
- **Alternate (OD-11 b):** (a) compares each locale's render; **(OD-11 c):** (c) has no directory to search.
- **Mutant probe:** a generator that renders every file would alter the 16 brace-bearing skill files and fail
  (a); one that copies `.tmpl` files unrendered fails (c).

### AC-006 — Flat command layout, discoverable inventory (REQ-006)

- **Given** the payload after M2 and a Claude CLI, **When** the layout test runs and the emitted marketplace is
  installed from its local path under an empty scratch home, **Then** every template command exists as a flat file,
  no subdirectory exists under `commands/`, and `claude plugin details` lists every expected skill and command name.
- **Verify (a):** `go test ./internal/template/pluginemit -run '^TestPayloadCommandsFlat$' -count=1 -v`
  — expected `--- PASS: TestPayloadCommandsFlat (…)`: one `commands/<stem>.md` per command stem of the template tree
  (the expected set is read from `internal/template/templates/.claude/commands/moai/`, `.tmpl` and `.md` dropped), and
  no directory under `commands/`.
- **Verify (b):** `sh scripts/check-plugin-discoverable.sh <claude-home:AC-006>` — expected exit 0 and a last line
  `ok: <N> names listed, 0 missing`. The script, with `CLAUDE_CONFIG_DIR` set to its argument, runs
  `claude plugin marketplace add <repo-root> --json`, `claude plugin install moai@moai-adk --json` and
  `claude plugin details moai@moai-adk`, and requires the `Skills (N)` line to carry `N` equal to the expected count
  and every expected name, and the `MCP servers (K)` line to match the keys of `plugins/moai/.mcp.json`. The expected
  names are the directory names under `plugins/moai/skills` plus the command stems of the **template** tree, never a
  literal and never read from `plugins/moai/commands`, so a payload that nests or omits its commands cannot define its
  own expectation. On a missing name it prints `missing: <name>` per name and exits 1.
- **Verify (c):** `sh scripts/check-plugin-discoverable.sh internal` — expected exit 2, a line beginning
  `refused:` naming the argument, and no write: the script refuses an argument that is not an existing empty
  directory, which keeps it away from a real profile.
- **Observed basis (P-30):** the same 17 command files flat produced `Skills (41)`; nested at `commands/moai/` the
  inventory printed `Skills (24)` and listed none of them. The nested form passes every other criterion of this
  file, so (a) and (b) are the only guards against it.
- **RED-now:** L-08 — the script is absent, exit 127; L-09 — the layout test's package is absent, exit 1. The base
  tree has no payload to inspect, and the observation behind the criterion is P-30, not a cell of this ledger.
- **Green path:** M2 emits flat commands and the script; (b) needs a Claude CLI on PATH and runs on demand
  (Gap G-7 states its continued-firing answer).
- **Alternate (OD-11 c):** no commands ship; (a) and the command half of (b) do not apply, the skills half stays.
- **Mutant probe:** a generator that mirrors the template tree (`commands/moai/<name>.md`) fails (a) and, on a machine
  with a Claude CLI, (b); a generator that drops one command fails (a) and (b).

### AC-007 — MCP entry copied from the template (REQ-007)

- **Given** the payload after M2, **When** the plugin's `.mcp.json` is compared with the template's,
  **Then** the plugin declares only `moai` and its `command` and `args` equal the template entry.
- **Verify (a):** `jq -c .mcpServers.moai plugins/moai/.mcp.json internal/template/templates/.mcp.json`
  — expected two identical lines, `{"command":"moai","args":["mcp-server"]}`.
- **Verify (b):** `jq -c '.mcpServers|keys' plugins/moai/.mcp.json` — expected `["moai"]`.
- **Verify (c):** `go test ./internal/template/pluginemit -run '^TestMCPEntryDerivedFromTemplate$' -count=1 -v`
  — expected `--- PASS: TestMCPEntryDerivedFromTemplate (…)` (changes the entry inside a synthetic template tree
  and asserts the emitted entry follows).
- **RED-now:** L-10 — `Could not open file plugins/moai/.mcp.json`, exit 2; control: the template entry prints
  `{"command":"moai","args":["mcp-server"]}`, exit 0.
- **Green path:** M2.
- **Alternate (OD-2 b):** the file is absent and (a) to (c) are void.
- **Mutant probe:** a retyped literal passes (a) today and fails (c) as soon as the template entry changes;
  an entry for `context7` (the template's second server) fails (b).
- **Stated limit:** Claude's dedupe key was observed only with a fixture that matched on command (SPEC RK-2);
  matching on both command and args is the strict superset. A plugin `.mcp.json` carrying only `moai` made the
  inventory print `MCP servers (1)  moai` (P-31).

### AC-008 — Scaffold-only components stay out (REQ-008)

- **Given** the payload after M2, **When** the top level of the plugin root is listed and the allow-list test
  runs, **Then** only the allowed entries exist.
- **Verify (a):** `find plugins/moai -maxdepth 1 -mindepth 1 -print`
  — expected exactly five lines, order unspecified: `plugins/moai/.claude-plugin`, `plugins/moai/.codex-plugin`,
  `plugins/moai/.mcp.json`, `plugins/moai/commands`, `plugins/moai/skills` (default pending OD-2, OD-3 and OD-11).
- **Verify (b):** `go test ./internal/template/pluginemit -run '^TestPayloadAllowList$' -count=1 -v`
  — expected `--- PASS: TestPayloadAllowList (…)` (a synthetic template tree gains `rules`, `hooks`,
  `output-styles`, `workflows`, `CLAUDE.md`; none is emitted).
- **RED-now:** L-11 — `bfs: error: plugins/moai: No such file or directory`, exit 1. An absent directory is
  also what makes this read red, so the green path is the five-line listing, and the control in L-11 shows the
  same `find` form lists the template `.claude` entries (it would also list `rules` and `hooks` there).
- **Green path:** M2.
- **Alternate (OD-3 a/b):** a sixth line, `plugins/moai/agents`, is the only addition; **(OD-2 c):** (b) admits the
  hook registrations that need no `$CLAUDE_PROJECT_DIR`; **(OD-11 c):** `plugins/moai/commands` is absent.
- **Mutant probe:** an allow-list implemented as a deny-list of `rules` and `hooks` passes today's tree and
  fails (b) on the synthetic `output-styles` entry.

### AC-009 — Read-only drift gate (REQ-009)

- **Given** the tree after M2, **When** the check runs on a clean tree and on mutated artifact sets,
  **Then** it passes on the clean tree, fails on each mutation, and is a prerequisite of `build`.
- **Verify (a):** `make plugin-emit-check` — expected exit 0, `ok` for the pluginemit package.
- **Verify (b):** `go test ./internal/template/pluginemit -run '^TestDriftDetectsMutatedArtifact$' -count=1 -v`
  — expected `--- PASS: TestDriftDetectsMutatedArtifact (…)` with four passing subtests `flipped-byte`,
  `deleted-file`, `extra-file` and `committed-set-unchanged` (the committed set is left byte-unchanged after a
  failing check).
- **Verify (c):** `grep -n ^build: Makefile` — expected one line that contains `plugin-emit-check`.
- **RED-now:** L-12 — ``make: *** No rule to make target `plugin-emit-check'.  Stop.``, exit 2.
- **Green path:** M1 adds the verbs, M2 extends the check to the payload and adds it to `build:`.
- **Mutant probe:** a check that regenerates before comparing passes (a) and fails (b) at
  `committed-set-unchanged`.

### AC-010 — Init calls the step, gated by the harness (REQ-010)

- **Given** an injected runner, and then the built binary with stub tools, **When** `moai init` completes template
  deployment under each harness, **Then** the step is called after deployment and acts on exactly the tools the
  harness selects.
- **Verify (a):** `go test ./internal/cli -run '^TestInitPluginStep_AfterDeployment$' -count=1 -v` — expected
  `--- PASS: TestInitPluginStep_AfterDeployment (…)`: the runner records no call before the deployed file set is
  complete and at least one call after it (default pending OD-1).
- **Verify (b):** `go test ./internal/cli -run '^TestInitPluginStep_HarnessGating$' -count=1 -v` — expected
  `--- PASS` with subtests `claude-default` (no `--llm`), `claude-explicit`, `gpt`, `both` and
  `unrecognized-falls-back-to-claude`: `claude` records only Claude vectors, `gpt` only Codex vectors, `both` both,
  Claude first (default pending OD-9).
- **Verify (c):** `sh scripts/test-plugin-install-step.sh bin/moai` — expected exit 0 and the lines
  `PASS init-claude-harness`, `PASS init-gpt-harness`, `PASS init-both-harness`: through the real default runner,
  a default `moai init` records two `claude` calls and no `codex` call, `--llm gpt` two `codex` calls and no `claude`
  call, `--llm both` four calls with the Claude pair first.
- **RED-now:** L-13 — the criterion's commands print `[no tests to run]` for (a) and (b), and the script is absent
  (exit 127); the base `moai init` makes no `plugin` call at all.
- **Green path:** M3.
- **Alternate (OD-1 b):** (a) asserts zero calls without the opt-in flag and a call with it; **(OD-9 b):** (b) records
  both tools under every harness and the three script lines change accordingly; **(OD-9 c):** unchanged.
- **Mutant probe:** a step that ignores the harness fails `claude-default` (a Codex vector appears) and
  `PASS init-claude-harness`; a step called before deployment ends fails (a).

### AC-011 — Add then install, idempotent, no scope argument (REQ-011)

- **Given** an injected runner and a PATH stub, and then the real tools under an empty scratch home, **When** the step
  acts on a tool, **Then** it runs add then install in that order, never runs install after a failed add, treats
  exit 0 as success, and a repeated run succeeds.
- **Verify (a):** `go test ./internal/cli -run '^TestPluginInstallStep_Sequence$' -count=1 -v` — expected
  `--- PASS: TestPluginInstallStep_Sequence (…)` with subtests `both-present`, `claude-only`, `codex-only`,
  `install-skipped-after-add-fails`, `already-present-is-success`, `no-scope-argument` and `pinned-binary-honored`.
  The recorded argument vectors are exactly `claude plugin marketplace add modu-ai/moai-adk`,
  `claude plugin install moai@moai-adk`, `codex plugin marketplace add modu-ai/moai-adk`,
  `codex plugin add moai@moai-adk` (no `--scope`, default pending OD-12), and with `MOAI_CLAUDE_BIN` set to a recording
  script the Claude vectors are recorded by that script, not by `claude` on PATH.
- **Verify (b):** run twice each, under `<claude-home:AC-011>`, with the repository root as the source (a local path
  stands in for the GitHub source, which no criterion contacts):
  `CLAUDE_CONFIG_DIR=<claude-home:AC-011> claude plugin marketplace add ./ --json` then
  `CLAUDE_CONFIG_DIR=<claude-home:AC-011> claude plugin install moai@moai-adk --json`
  — expected exit 0 on all four runs; first add `"outcome":"ok"` with `Successfully added marketplace: moai-adk`;
  second add `already on disk`; first install `Successfully installed plugin: moai@moai-adk`; second install
  contains `is already installed`.
- **Verify (c):** `CODEX_HOME=<codex-home:AC-011> codex plugin marketplace add .` then
  `CODEX_HOME=<codex-home:AC-011> codex plugin add moai@moai-adk`, each run twice — expected exit 0 on all four;
  the first `plugin add` prints ``Added plugin `moai` from marketplace `moai-adk`.`` (the second prints the same
  line, P-10).
- **Stated limit:** a local-path marketplace loads in place, so (b) and (c) prove the committed manifests are
  accepted and the commands are idempotent, not that the payload is copied into the cache; the payload's
  discoverability is AC-006. The GitHub source `modu-ai/moai-adk` is never contacted (G-1).
- **RED-now:** L-14 for (a); L-15 for (b) — `"failureCode":"manifest_missing"`, exit 1 at the base tree; the control
  shows `./` adds a present marketplace with exit 0.
- **Green path:** M3 for (a); (b) and (c) need M1 and M2 as well.
- **Alternate (OD-12 b/c):** the Claude vectors gain `--scope project` or `--scope local` and (b) gains the same flag.
- **Mutant probe:** swapping the two commands, running install after a failed add, or treating the tool's own
  `already installed` message as failure each fails one named subtest.

### AC-012 — Environment passed unchanged, config home printed (REQ-012)

- **Given** an injected runner and set or unset config-home variables, **When** the step starts a tool command,
  **Then** the child environment equals the parent's and the config home is printed.
- **Verify (a):** `go test ./internal/cli -run '^TestPluginInstallStep_Environment$' -count=1 -v` — expected
  `--- PASS` with subtests `env-unchanged` (the runner's `CLAUDE_CONFIG_DIR` and `CODEX_HOME` equal the parent's, and
  no variable is added or removed), `prints-config-home-claude`, `prints-config-home-codex` and
  `default-home-when-unset` (the printed home is `~/.claude` or `~/.codex` when the variable is unset). Default
  pending OD-10.
- **Verify (b):** `sh scripts/test-plugin-install-step.sh bin/moai` — expected the line `PASS init-config-home-printed`
  (the scratch `CLAUDE_CONFIG_DIR` path appears on the step's stderr).
- **RED-now:** L-16 — `[no tests to run]`; the base init prints no config home.
- **Green path:** M3.
- **Alternate (OD-10 b):** `env-unchanged` becomes `recorded-profile-used-when-env-unset`; **(OD-10 c):** a
  `skips-on-profile-mismatch` subtest is added.
- **Mutant probe:** a step that sets `CLAUDE_CONFIG_DIR` for the child fails `env-unchanged`; a step that prints
  nothing fails both lines.

### AC-013 — Failure prints guidance and leaves the exit status alone (REQ-013)

- **Given** a runner that fails, times out, or an invalid Claude pin, **When** the step runs, **Then** one guidance
  block naming both manual commands is printed, no further command runs for that tool, and the caller's exit status
  is unchanged.
- **Verify (a):** `go test ./internal/cli -run '^TestPluginInstallStep_FailOpen$' -count=1 -v` — expected
  `--- PASS` with subtests `exit-nonzero`, `timeout` (a runner that blocks until its context ends is cancelled at the
  deadline carried by `config.DefaultPluginInstallCommandTimeout`, and the block is printed), `invalid-pin`,
  `add-fails-no-install`, `guidance-names-both-commands` and `returns-nil`.
- **Verify (b):** `grep -n DefaultPluginInstallCommandTimeout internal/config/defaults.go` — expected one
  declaration line whose value is `60 * time.Second` (static; the `timeout` subtest is the behavior check). The
  value is named so that RK-15's 240-second worst case is a visible number, not a literal in a call.
- **Verify (c):** `sh scripts/test-plugin-install-step.sh bin/moai` — expected the lines
  `PASS init-tool-fails-exit-0` and `PASS init-add-fails-skips-install`: with a stub `claude` that exits 1,
  `moai init` exits 0 and its stderr names both manual commands; when only `marketplace add` fails no
  `claude plugin install` call is recorded.
- **RED-now:** L-17 — `[no tests to run]` for (a), and the constant is absent for (b).
- **Green path:** M3.
- **Mutant probe:** a step that returns the tool's exit code fails `returns-nil` and `PASS init-tool-fails-exit-0`; a
  step with no deadline hangs `timeout`; a step that installs after a failed add fails
  `PASS init-add-fails-skips-install`.

### AC-014 — Absent tool prints one skip line (REQ-014)

- **Given** a PATH without `claude`, without `codex`, or without both, **When** the step is to act on the missing
  tool, **Then** exactly one skip line per missing tool is printed, no guidance block is printed, and the caller's
  exit status is unchanged.
- **Verify (a):** `go test ./internal/cli -run '^TestPluginInstallStep_ToolAbsent$' -count=1 -v` — expected
  `--- PASS` with subtests `claude-absent`, `codex-absent` and `both-absent`: one line each, the guidance block's
  marker text absent, a nil result.
- **Verify (b):** `sh scripts/test-plugin-install-step.sh bin/moai` — expected the line `PASS init-no-tools-one-skip-line`
  (through the real binary, with no tool on PATH, `moai init` exits 0 and prints one skip line and no guidance block).
- **RED-now:** L-18 — `[no tests to run]`.
- **Green path:** M3.
- **Mutant probe:** a step that prints the guidance block for an absent tool fails (a) and (b); a step that prints
  nothing fails (a).

### AC-015 — Opt-out runs no tool command (REQ-015)

- **Given** the opt-out set by flag or by environment, **When** the step is reached, **Then** no external command
  runs.
- **Verify (a):** `go test ./internal/cli -run '^TestPluginInstallStep_OptOut$' -count=1 -v` — expected `--- PASS`
  with subtests `flag`, `env-1`, `env-true`, `env-empty-is-not-optout` and `env-0-is-not-optout`: the first three give
  zero runner calls, the last two do not opt out (default pending OD-5).
- **Verify (b):** `sh scripts/test-plugin-install-step.sh bin/moai` — expected the lines
  `PASS init-flag-optout-zero-calls` and `PASS init-env-optout-zero-calls` (real binary, stub tools, no recorded call).
- **Verify (c):** `grep -c MOAI_SKIP_PLUGIN_INSTALL internal/config/envkeys.go` — expected at least `1` (static: the
  constant exists; the name lives in `envkeys.go`, `internal/cli/CLAUDE.md` Env var access).
- **RED-now:** L-19 for (a); L-20 for (c) — prints `0`.
- **Green path:** M3.
- **Alternate (OD-5 b):** per-tool subtests are added; **(OD-5 c):** a config-key subtest is added.
- **Mutant probe:** an opt-out checked after the commands fails (a) and (b); a flag that exists but is not read fails
  `flag`.

### AC-016 — Automated callers set the opt-out (REQ-016)

- **Given** moai's internal `init` re-entry and the repository's e2e script, **When** each invokes `moai init`,
  **Then** each sets the opt-out, and no other automated caller exists unnoticed.
- **Verify (a):** `go test ./internal/cli -run '^TestDoctorAgentEmitEmbed_SetsPluginOptOut$' -count=1 -v` — expected
  `--- PASS`; the captured environment of the `init --non-interactive --llm both` child
  (`internal/cli/doctor_agentemit_embed.go:352`) contains the opt-out.
- **Verify (b):** `go test ./internal/cli -run '^TestPluginOptOutCallersEnumerated$' -count=1 -v` — expected `--- PASS`.
  The test scans the non-test Go sources under `internal/` for `exec.Command` and `exec.CommandContext` calls whose
  argument list holds the literal `"init"` and whose command is not the literal `"git"` (the `git init` calls of the
  fixture helpers are not moai callers), and fails on any caller not in its list (today the doctor re-entry only); it also reads `e2e/cli/tux3_journeys.sh` and requires the opt-out on every non-comment line
  that runs `'$BIN' init`. A new automated caller therefore turns it red until it is listed and sets the opt-out.
- **Verify (c) — static:** `grep -c MOAI_SKIP_PLUGIN_INSTALL e2e/cli/tux3_journeys.sh` — expected at least `2`
  (lines 104 and 115 at the base tree). A comment would satisfy this count; (b) is the check that does not.
- **RED-now:** L-21 — `[no tests to run]` for (a) and (b), `0` for (c); both e2e invocations carry no opt-out.
- **Green path:** M3.
- **Mutant probe:** a doctor child that does not set the opt-out fails (a); an e2e line without the variable on its
  own line fails (b); a new `exec.Command(bin, "init", …)` fails (b).

### AC-017 — The default runner refuses under a test binary (REQ-017)

- **Given** the test binary, **When** the step runs through its default runner with a pinned Claude binary and tool
  stubs on PATH that record every start, **Then** the default runner refuses and nothing is recorded.
- **Verify:** `go test ./internal/cli -run '^TestPluginInstallStep_NoRealRunnerUnderTest$' -count=1 -v` — expected
  `--- PASS` with subtests `default-runner-refuses` (the default runner called directly returns a refusal and
  starts nothing), `pin-and-path-shims-untouched` (`MOAI_CLAUDE_BIN` points at a recording script, recording `claude`
  and `codex` stubs sit first on PATH, `CLAUDE_CONFIG_DIR` and `CODEX_HOME` are scratch directories, `runInit` runs on a
  temporary project through its default path, and the record stays empty) and `rune-callers-covered` (the same through
  `initCmd.RunE`). The PATH shim alone would miss a pinned binary; the pin is why the second subtest records both.
- **No per-site edit:** SPEC P-20 counts 32 `runInit(` call sites in 18 test files and 9 `initCmd.RunE(` sites in 5; the
  criterion does not require any of them to install a fake. The one mechanism covers all of them, and the callers
  outside the test binary are AC-016's.
- **RED-now:** L-22 — `[no tests to run]`.
- **Green path:** M3.
- **Mutant probe:** a default runner that execs fails `pin-and-path-shims-untouched` (the record is non-empty); a
  detector that keys on any argument ending `.test` instead of the program name would also refuse a project named
  `app.test` in production — `design.md` §3.4 records why that detector is not reused.

### AC-018 — Install scripts call the verb and fail open (REQ-018)

- **Given** the three scripts and the two docs-site copies after M3, **When** the script is run offline with a pinned
  local archive and stub tools, **Then** the installed binary's verb is called by its installed path, an opt-out and a
  failing or missing verb leave the installer successful, and the copies are identical.
- **Verify (a):** `sh scripts/test-plugin-install-step.sh bin/moai` — expected the lines
  `PASS installer-calls-verb-by-installed-path` (`install.sh --install-dir <dir not on PATH>` records the four tool
  vectors through the real binary's verb, and the call's argv0 is the installed path),
  `PASS installer-optout` (zero recorded calls, installer exit 0), `PASS installer-set-e-guard` (a pinned `moai` whose
  verb exits 1: `install.sh` exits 0 and prints `Installation complete!` although it runs under `set -e`) and
  `PASS installer-old-binary-unknown-verb` (a pinned `moai` that prints `Unknown command "plugin" for "moai".` and exits
  1: the same outcome).
- **Verify (b):** `bash -n install.sh` — expected exit 0.
- **Verify (c) — static:** `go test ./internal/cli -run '^TestInstallScriptsPluginStepGuarded$' -count=1 -v` —
  expected `--- PASS`. Labelled static: it reads `install.ps1` and `install.bat` text. For `install.ps1` the verb call
  must sit on a non-comment line inside a `try` block with a `catch` that does not rethrow and must use the install
  path variable, not a bare `moai`, because `$ErrorActionPreference = "Stop"` (line 6) would otherwise abort the
  installer after the binary landed; for `install.bat` the call must use `"%TARGET_PATH%"` and be followed by no
  `exit /b` that propagates its level. The behavior of both is not executed locally (G-3).
- **Verify (d):** `cmp install.sh docs-site/static/install.sh` (exit 0, no output) and
  `cmp install.ps1 docs-site/static/install.ps1` (exit 0, no output).
- **RED-now:** L-23 — the harness script is absent, exit 127; L-24 — `grep -c MOAI_SKIP_PLUGIN_INSTALL install.sh`
  prints `0`; L-25 — `[no tests to run]`; L-26 — the base `install.sh` driven offline makes no plugin call, the red
  the `installer-calls-verb-by-installed-path` case would show.
- **Green path:** M3.
- **Alternate (OD-6 b):** the scripts inline the two tool commands; (a)'s first, third and fourth lines are replaced by
  cases on the inline form and (c) greps the inline commands.
- **Mutant probe:** an unguarded call fails `installer-set-e-guard`; a call through a bare `moai` fails
  `installer-calls-verb-by-installed-path` (the install directory is not on PATH); editing only the root `install.sh`
  fails the first `cmp`; two comment lines that mention the verb fail (a).
- **Stated limit (G-3):** `pwsh` exists at `/usr/local/bin/pwsh` but the worktree guard refuses it (P-36), so no
  PowerShell run backs this criterion; CI runs `install.bat` end to end on Windows and tokenises `install.ps1` under
  pwsh (`test-install.yml`), which neither executes the step.

### AC-019 — The install verb (REQ-019, applies under OD-6 option (a))

- **Given** the verb after M3, **When** it runs with stub tools, with none, and with the opt-out, **Then** it acts on
  every tool found, prints on stderr, and exits 0 on every fail-open outcome.
- **Verify (a):** `go test ./internal/cli -run '^TestPluginInstallCmd$' -count=1 -v` — expected `--- PASS` with
  subtests `registered-in-root-help`, `help-names-opt-out`, `exit-0-on-fail-open-outcomes`,
  `exit-nonzero-on-unknown-flag` and `no-harness-filter`.
- **Verify (b):** `bin/moai plugin install --help` — expected exit 0.
- **Verify (c):** `sh scripts/test-plugin-install-step.sh bin/moai` — expected the lines `PASS verb-install-all-tools`
  (four recorded vectors, exit 0), `PASS verb-no-tools-exit-0` and `PASS verb-optout-zero-calls`.
- **RED-now:** L-27 — `Unknown command "plugin" for "moai".`, exit 1; L-28 — `[no tests to run]`.
- **Green path:** M3.
- **Alternate (OD-6 b):** the criterion does not apply; **(OD-9 c):** `no-harness-filter` becomes `tool-selector`.
- **Mutant probe:** a verb that returns the tool's exit code fails `exit-0-on-fail-open-outcomes`; a verb that applies
  a harness filter fails `no-harness-filter`.

### AC-020 — Doctor reads the Claude registry without a subprocess (REQ-020)

- **Given** synthetic config homes and an injected binary version, **When** the check runs, **Then** the Claude
  installed version comes from the registry file under the resolved home, and no process is started.
- **Verify:** `go test ./internal/cli -run '^TestCheckPluginVersion_ClaudeRead$' -count=1 -v` — expected `--- PASS`
  with subtests `equal`, `v-prefix-normalized`, `prerelease-equal`, `claude-config-dir-honored` (a registry under
  `$CLAUDE_CONFIG_DIR`, never `~/.claude`), `user-scope-first` (a registry whose `moai@moai-adk` array holds several
  scopes: the user-scope entry is read first and the message names the scope) and `no-subprocess` (PATH holds a
  recording `claude` stub and the record stays empty).
- **RED-now:** L-29 — `[no tests to run]`.
- **Green path:** M4.
- **UNOBSERVED:** the registry format at versions other than claude 2.1.287 (RK-13).
- **Mutant probe:** a check that runs `claude plugin list` fails `no-subprocess`; a check that reads `~/.claude`
  regardless fails `claude-config-dir-honored`.

### AC-021 — Doctor reads the Codex state through the CLI, bounded (REQ-021)

- **Given** canned `codex plugin list --json` outputs shaped as the real one (P-34) and an injected runner, **When** the
  check runs, **Then** installed state and version come from that output, the plugin cache is not consulted for
  installed state, and a slow or absent probe degrades to OK or info.
- **Verify (a):** `go test ./internal/cli -run '^TestCheckPluginVersion_CodexRead$' -count=1 -v` — expected `--- PASS`
  with subtests `registered-version`, `list-empty-is-not-installed`, `cache-without-registration` (the list prints
  `"installed": []` while a cache directory holds a version: not installed), `multiple-cache-versions` (the list reports
  one version while the cache holds two: the listed one), `codex-absent-no-spawn`, `probe-timeout-is-info`,
  `probe-malformed-json-is-info` and `probe-bound-equals-constant` (the context deadline equals
  `config.DefaultPluginVersionProbeTimeout`, 3 seconds). Default pending OD-13.
- **Verify (b) — real shape:** under `<codex-home:AC-021>`: `CODEX_HOME=<codex-home:AC-021> codex plugin marketplace add .`,
  `CODEX_HOME=<codex-home:AC-021> codex plugin add moai@moai-adk`, then
  `CODEX_HOME=<codex-home:AC-021> codex plugin list --json` — expected the last output to hold `"pluginId": "moai@moai-adk"`
  and a `"version"` equal to the plugin manifest's, so the canned shapes of (a) are a measured shape, not an invention.
- **RED-now:** L-30 — `[no tests to run]`.
- **Green path:** M4; (b) needs M1 and M2.
- **Alternate (OD-13 b):** subtests read a `config.toml` stanza and the cache directory, and `cache-without-registration`
  is decided by the stanza; **(OD-13 c):** the criterion does not apply.
- **UNOBSERVED:** the list shape at codex versions other than 0.160.0.
- **Mutant probe:** a check that decides installed state from the cache directory fails
  `cache-without-registration`; a probe with no deadline fails `probe-bound-equals-constant`; a check that spawns
  `codex` when it is absent fails `codex-absent-no-spawn`.

### AC-022 — Doctor comparison and outcomes (REQ-022)

- **Given** synthetic homes, a canned probe and an injected binary version, **When** the check runs, **Then** equality
  is OK, inequality takes the OD-7 severity with both versions and the remedy, and every indeterminate case is OK or
  info.
- **Verify (a):** `go test ./internal/cli -run '^TestCheckPluginVersion_Outcomes$' -count=1 -v` — expected `--- PASS`
  with subtests `mismatch-warn` (default pending OD-7), `mismatch-names-both-and-remedy`, `not-installed`,
  `home-absent`, `registry-malformed`, `registry-unknown-shape`, `dev-build` and `codex-timeout`; none of the last six
  returns warn or fail.
- **Verify (b):** `CLAUDE_CONFIG_DIR=<claude-home:AC-022> CODEX_HOME=<codex-home:AC-022> bin/moai doctor --check "Plugin Version"`
  — expected exit 0 and output that contains the row name `Plugin Version` and `Fail 0`. The unregistered form prints
  `Pass 0    Warn 0    Fail 0` and no row (P-38), so the name in the output is what separates the two (a counter
  value would not: an info outcome need not count as a pass). Both homes are scratch, so the check cannot read a real
  profile.
- **RED-now:** L-31 — `[no tests to run]` for (a); for (b) the base binary prints the System Diagnostics box with
  `Pass 0    Warn 0    Fail 0` and no `Plugin Version` row.
- **Green path:** M4.
- **Alternate (OD-7 b/c):** `mismatch-warn` becomes `mismatch-fail` or `mismatch-info`.
- **Mutant probe:** a check that returns fail on a malformed registry fails `registry-malformed`; a check that warns on
  a development build fails `dev-build`; a check that compares `v3.1.3` with `3.1.3` as unequal fails
  `v-prefix-normalized` (AC-020).

### AC-023 — Doctor registration and bounded output (REQ-023)

- **Given** the new check name, **When** the guarded doctor tests run, **Then** the check appears on each golden
  surface, the guarded tests pass with it registered, and the default run adds one line.
- **Verify (a):** `grep -c 'Plugin Version' internal/cli/testdata/doctor-nocolor.golden`, the same on
  `doctor-light.golden` and on `doctor-dark.golden` — expected `1` each. A check that is never registered leaves no
  row and prints `0`, so this fails where the guarded tests below stay green (they pass today and would pass
  without the check). The name `Plugin Deployment` (the legacy `system.yaml` marker check) does not match.
- **Verify (b):** `go test ./internal/cli -run '^TestBinaryLag_.*$' -count=1 -v` (all `TestBinaryLag_` tests pass,
  including `TestBinaryLag_DoctorCheckNameSetIsUnchanged` with the new name in `namesAddedAfterBaseline` and
  `TestBinaryLag_AllowlistKeysAreLiveNames`, which binds the key to a name `doctor.go` really registers);
  `go test ./internal/cli -run '^TestRunDiagnosticChecks.*$' -count=1` (`ok`);
  `go test ./internal/cli -run '^TestDoctorGolden_.*$' -count=1` (`ok`).
- **Verify (c):** `go test ./internal/cli -run '^TestCheckPluginVersion_OutputBounded$' -count=1 -v` — expected
  `--- PASS`: the default run emits one summary line and an empty detail, `--verbose` adds detail.
- **Verify (d):** `go test ./internal/cli -run '^TestDoctorGolden_IgnoresCallerCodexHome$' -count=1 -v` — expected
  `--- PASS`: with `CODEX_HOME` pointing at a directory that holds a registered `moai` plugin, the golden output is
  unchanged, so the harness scrubs `CODEX_HOME` (the `captureDoctorCmd` harness pins `HOME`, `MOAI_HOME` and
  `CLAUDE_CONFIG_DIR` but not `CODEX_HOME`, P-21). Static companion: `grep -n 'CODEX_HOME\|codexHomeEnvVar' internal/cli/doctor_golden_test.go`
  — expected at least one line.
- **RED-now:** L-32 — the three greps print `0`, none of them an artifact of the grep (the control prints `1` for the
  existing `MCP Server Version` row in each golden); L-33 — `[no tests to run]` for (c) and (d), no line for the static
  companion.
- **Green path:** M4.
- **Mutant probe:** registering the check without the allowlist entry fails
  `TestBinaryLag_DoctorCheckNameSetIsUnchanged` (lesson of cards t1251 and t1282); never registering it fails (a);
  a golden harness that leaves `CODEX_HOME` unscrubbed fails (d) on a machine that has the variable set.

### AC-024 — Release tag check and runbook (REQ-024)

- **Given** the script, the release workflow and the runbook after M4, **When** the script is run with a matching and a
  non-matching tag, **Then** it exits 0 for the match and 1, naming both values, for the mismatch; the release workflow
  runs it inside `verify-provenance`; and the runbook names the regeneration and inventory steps.
- **Verify (a):** `go test ./internal/template/pluginemit -run '^TestPluginVersionScript$' -count=1 -v` — expected
  `--- PASS` with subtests `ssot-tag-accepted` (the argument is `version.Version`, read by the test, so the criterion has
  no version literal and does not rot at the first bump), `other-tag-rejected-names-both` and `missing-argument-exit-2`.
- **Verify (b):** `sh scripts/check-plugin-version.sh v9.9.9` — expected exit 1, output names `9.9.9` and the committed
  plugin version. The mismatching tag is the known failing input; no file is mutated.
- **Verify (c):** `go test ./internal/template/pluginemit -run '^TestReleaseWorkflowCallsPluginVersionCheck$' -count=1 -v`
  — expected `--- PASS`: `.github/workflows/release.yml`, parsed as YAML, has a step in job `verify-provenance` whose
  command runs `scripts/check-plugin-version.sh` with the tag. Static companion:
  `grep -n check-plugin-version .github/workflows/release.yml` — expected at least one line.
- **Verify (d) — static:** `grep -n 'plugin-emit' .moai/docs/version-management.md` and
  `grep -n 'check-plugin-discoverable' .moai/docs/version-management.md` — expected at least one line each: the runbook
  lists regeneration as a step of every bump and the inventory check as a pre-tag step (the carrier of G-7).
- **RED-now:** L-34 — the package is absent, exit 1; L-35 — `sh: scripts/check-plugin-version.sh: No such file or
  directory`, exit 127; L-36 — the three greps select no line.
- **Green path:** M4.
- **Mutant probe:** a script that prints the mismatch and exits 0 passes nothing in (a) or (b); a workflow that does not
  call it fails (c); a runbook that omits the step fails (d).

## Edge Cases

- Neither tool on PATH: one skip line per absent tool, `moai init` exit 0, no guidance block (AC-014).
- `llm.claude_bin` pin invalid or `MOAI_CLAUDE_BIN` pointing nowhere: the resolver errors; the step prints the
  guidance block for Claude and returns success (AC-013 `invalid-pin`).
- Config home missing or unwritable: the tool's own failure lands in the guidance path; the doctor check reports OK with a
  not-installed message.
- Binary built with `make build` (`VERSION=v3.2.0-rc.N`) against a plugin stamped from the fallback `v3.1.3`: a mismatch is
  reported at the OD-7 severity, so a maintainer machine warns at the default; a development build (`IsDevBuild`) is never
  compared.
- A registry whose `moai@moai-adk` array holds several scopes: the check reads the user-scope entry first and states which
  scope it read (AC-020).
- `--non-interactive` init (no project `.mcp.json` entry): the step still runs unless opted out (SPEC §1.5-2, OD-1, OD-5).
- `moai init` under `--llm gpt`: only the Codex tool is acted on at the OD-9 default; no Claude profile is written.
- `moai init my-app.test`: a project name ending `.test` must not be read as a test binary (AC-017 mutant note).
- A script served from `main` against a release that lacks the verb: the installer prints the verb's error, finishes, and
  exits as the binary installation did (AC-018 `installer-old-binary-unknown-verb`).
- The Codex cache holds a version directory while `codex plugin list --json` prints `"installed": []` (a hand-edited or
  tool-removed registration): not installed (AC-021 `cache-without-registration`).

## Quality Gates

- Tested: `go test -cover` at least 85% for `internal/template/pluginemit`, and for the new files of
  `internal/cli` (`plugin_install.go`, the verb, `doctor_plugin_version.go`).
- Readable and Unified: `go vet` and the CI-pinned golangci-lint over the changed packages; `gofmt` clean; `sh -n` on
  each new script.
- Secured: child processes are started with an argument vector, never through a shell string; no value of the
  environment is printed except the two config-home paths; registry and probe JSON are parsed defensively; the discoverable
  script refuses any argument that is not an empty directory.
- Trackable: the card id `t1435` is in every commit message on the branch.
- Cross-platform: `GOOS=windows GOARCH=amd64 go build ./...` exits 0.
- Preservation: existing init and update deployment tests pass unmodified (the deployed file set is unchanged); the 32 `runInit(`
  and 9 `initCmd.RunE(` call sites are not edited.
- Lifecycle lint: `moai spec lint` reports no error on this SPEC directory.

## Definition of Done

All twenty-four criteria pass with the commands above (AC-019 only under OD-6 option (a), AC-021 and AC-006's command half
only under their defaults), each reported in the five-section format with the judging build's commit next to the tree HEAD;
the Gaps section names G-1 to G-7 (plan §4) and any refusal of an acceptance command by the worktree guard; the Kickoff
record carries, for every row of `decision-index.md`, either an operator verdict or the line `default adopted` — never
silence — and a verdict that differs from a default changes only the clauses the `spec.md` §5 marker table lists, with the
plan-artifact hash re-computed and `plan-auditor` run on the delta; `plan-auditor` verdict is PASS at the Tier L threshold
(0.85).

## Evidence Ledger (RED-now cells)

Every entry names its tree. Commands ran from the worktree root unless stated. `<scratch>` is the session scratchpad
`/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/2f10c8c5-67ea-41c2-9b61-6242acc465c3/scratchpad/s2`; homes named
`red/…` were empty directories under it (`ls -A` showed `total 0` before first use). `<tree-built-moai>` is
`<scratch>/moai-t1435`, built by `go build -o <scratch>/moai-t1435 ./cmd/moai` at `3766cef05` and self-reporting
`moai-adk v3.1.3 none built unknown` (no ldflags, so its commit stamp is by build procedure, not self-attested).

```
L-01  AC-001   [carried from E-1; tree 7109e0900; re-executed at 3766cef05 by the audit]
cmd:    CLAUDE_CONFIG_DIR=<claude-home:AC-001> claude plugin validate .claude-plugin/marketplace.json --strict
exit:   1
stdout: Validating marketplace manifest: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/.claude-plugin/marketplace.json

        ✘ Found 1 error:

          ❯ file: File not found: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/.claude-plugin/marketplace.json

        ✘ Validation failed
control: same command (file-path form) on a present, internally consistent scratch marketplace -> "✔ Validation passed", exit 0
```

```
L-02  AC-002   [tree 3766cef05; supersedes E-2, whose command carried a timeout wrapper]
cmd:    CODEX_HOME=<scratch>/red/x-ac002 codex plugin marketplace add .
exit:   1
stdout: Error: invalid marketplace file `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435`: marketplace root does not contain a supported manifest
control: a root holding .agents/plugins/marketplace.json -> "Added marketplace `moai-adk` from <root>." (this revision's fixture
         <scratch>/fx-core-flat), exit 0
```

```
L-03  AC-003 (a)   [carried from E-3; tree 7109e0900]
cmd:    CLAUDE_CONFIG_DIR=<claude-home:AC-003> claude plugin validate plugins/moai --strict
exit:   1
stdout: Validating plugin manifest: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/plugins/moai

        ✘ Found 1 error:

          ❯ file: File not found: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/plugins/moai

        ✘ Validation failed
control: strict validate of this revision's fixture plugin directory <scratch>/fx-core-flat/plugins/moai -> "✔ Validation passed"
```

```
L-04  AC-003 (d), and the package of AC-015's iteration-0 form   [carried from E-14; tree 7109e0900]
cmd:    go test ./internal/template/pluginemit -run '^TestVersionStampedFromSSOT$' -count=1 -v
exit:   1
stdout: # ./internal/template/pluginemit
        stat /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/internal/template/pluginemit: directory not found
        FAIL	./internal/template/pluginemit [setup failed]
        FAIL
control: grep -n 'Version = "v' pkg/version/version.go -> 11:	Version = "v3.1.3", exit 0
```

```
L-05  AC-004 (a)   [carried from E-4; tree 7109e0900]
cmd:    go test ./internal/template/pluginemit -run '^TestEmitDerivesFromTree$' -count=1 -v
exit:   1
stdout: # ./internal/template/pluginemit
        stat /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/internal/template/pluginemit: directory not found
        FAIL	./internal/template/pluginemit [setup failed]
        FAIL
```

```
L-06  AC-004 (b)   [tree 3766cef05]
cmd:    go test ./internal/template/pluginemit -run '^TestGeneratorHoldsNoComponentNames$' -count=1 -v
exit:   1
stdout: # ./internal/template/pluginemit
        stat /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/internal/template/pluginemit: directory not found
        FAIL	./internal/template/pluginemit [setup failed]
        FAIL
```

```
L-07  AC-005 (b)   [carried from E-5; tree 7109e0900]
cmd:    diff -r internal/template/templates/.claude/skills/moai plugins/moai/skills/moai
exit:   2
stdout: diff: plugins/moai/skills/moai: No such file or directory
control: grep -rlE '\{\{' internal/template/templates/.claude/commands -> 15 lines, each internal/template/templates/.claude/commands/moai/<name>.md.tmpl, exit 0
also (3766cef05): go test ./internal/template/pluginemit -run '^TestEmitFidelity$' -count=1 -v -> the same [setup failed] output as L-05, exit 1
```

```
L-08  AC-006 (b)   [tree 3766cef05]
cmd:    sh scripts/check-plugin-discoverable.sh <scratch>/red/c-ac006
exit:   127
stdout: sh: scripts/check-plugin-discoverable.sh: No such file or directory
basis:  the criterion exists because of P-30 (R-01): the 17 template commands flat -> `Skills (41)`; nested at commands/moai/ -> `Skills (24)`
```

```
L-09  AC-006 (a)   [tree 3766cef05]
cmd:    go test ./internal/template/pluginemit -run '^TestPayloadCommandsFlat$' -count=1 -v
exit:   1
stdout: # ./internal/template/pluginemit
        stat /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/internal/template/pluginemit: directory not found
        FAIL	./internal/template/pluginemit [setup failed]
        FAIL
```

```
L-10  AC-007   [carried from E-6; tree 7109e0900]
cmd:    jq -c .mcpServers plugins/moai/.mcp.json
exit:   2
stdout: jq: error: Could not open file plugins/moai/.mcp.json: No such file or directory
control: jq -c .mcpServers.moai internal/template/templates/.mcp.json -> {"command":"moai","args":["mcp-server"]}, exit 0
also (3766cef05): go test ./internal/template/pluginemit -run '^TestMCPEntryDerivedFromTemplate$' -count=1 -v -> [setup failed], exit 1
```

```
L-11  AC-008   [carried from E-7; tree 7109e0900]
cmd:    find plugins/moai -maxdepth 1 -mindepth 1 -print
exit:   1
stdout: bfs: error: plugins/moai: No such file or directory.
control: find internal/template/templates/.claude -maxdepth 1 -mindepth 1 -print -> 9 entries (loop.md, output-styles, workflows, agents, settings.json.tmpl, hooks, rules, commands, skills), exit 0
also (3766cef05): go test ./internal/template/pluginemit -run '^TestPayloadAllowList$' -count=1 -v -> [setup failed], exit 1
```

```
L-12  AC-009   [carried from E-8; tree 7109e0900]
cmd:    make plugin-emit-check
exit:   2
stdout: make: *** No rule to make target `plugin-emit-check'.  Stop.
also (3766cef05): go test ./internal/template/pluginemit -run '^TestDriftDetectsMutatedArtifact$' -count=1 -v -> [setup failed], exit 1
```

```
L-13  AC-010   [tree 3766cef05]
cmd:    go test ./internal/cli -run '^TestInitPluginStep_HarnessGating$' -count=1 -v
exit:   0
stdout: testing: warning: no tests to run
        PASS
        ok  	github.com/modu-ai/moai-adk/internal/cli	0.959s [no tests to run]
reason: red by the PASS-line rule: no `--- PASS: TestInitPluginStep_HarnessGating (…)` line exists
control: go test ./internal/cli -run '^TestBinaryLag_AllowlistKeysAreLiveNames$' -count=1 -v ->
         === RUN   TestBinaryLag_AllowlistKeysAreLiveNames
         --- PASS: TestBinaryLag_AllowlistKeysAreLiveNames (0.00s)
         PASS
         ok  	github.com/modu-ai/moai-adk/internal/cli	0.722s
also:   go test ./internal/cli -run '^TestInitPluginStep_AfterDeployment$' -count=1 -v -> the same three lines, ok ... 1.196s [no tests to run]
also:   sh scripts/test-plugin-install-step.sh bin/moai -> exit 127, "sh: scripts/test-plugin-install-step.sh: No such file or directory"
```

```
L-14  AC-011 (a)   [tree 3766cef05]
cmd:    go test ./internal/cli -run '^TestPluginInstallStep_Sequence$' -count=1 -v
exit:   0
stdout: testing: warning: no tests to run
        PASS
        ok  	github.com/modu-ai/moai-adk/internal/cli	0.970s [no tests to run]
reason: red by the PASS-line rule; control as L-13
```

```
L-15  AC-011 (b)   [carried from E-10; tree 7109e0900]
cmd:    CLAUDE_CONFIG_DIR=<claude-home:AC-011> claude plugin marketplace add ./ --json
exit:   1
stdout: {"command":"marketplace-add","outcome":"failed","message":"Marketplace file not found at /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/.claude-plugin/marketplace.json","failureCode":"manifest_missing"}
note:   the form `add .` is rejected as invalid_source ("Try: owner/repo, https://..., or ./path"), so the criterion uses `./`
control: the same `add ./` run from a directory holding a valid marketplace -> {"command":"marketplace-add","outcome":"ok",...}, exit 0
```

```
L-16  AC-012   [tree 3766cef05]
cmd:    go test ./internal/cli -run '^TestPluginInstallStep_Environment$' -count=1 -v
exit:   0
stdout: testing: warning: no tests to run
        PASS
        ok  	github.com/modu-ai/moai-adk/internal/cli	0.777s [no tests to run]
reason: red by the PASS-line rule; control as L-13
```

```
L-17  AC-013   [tree 3766cef05]
cmd:    go test ./internal/cli -run '^TestPluginInstallStep_FailOpen$' -count=1 -v
exit:   0
stdout: testing: warning: no tests to run
        PASS
        ok  	github.com/modu-ai/moai-adk/internal/cli	0.907s [no tests to run]
reason: red by the PASS-line rule; control as L-13
also:   grep -n DefaultPluginInstallCommandTimeout internal/config/defaults.go -> no line selected (status 1, not echoed by the tool)
```

```
L-18  AC-014   [tree 3766cef05]
cmd:    go test ./internal/cli -run '^TestPluginInstallStep_ToolAbsent$' -count=1 -v
exit:   0
stdout: testing: warning: no tests to run
        PASS
        ok  	github.com/modu-ai/moai-adk/internal/cli	0.985s [no tests to run]
reason: red by the PASS-line rule; control as L-13
```

```
L-19  AC-015 (a)   [tree 3766cef05]
cmd:    go test ./internal/cli -run '^TestPluginInstallStep_OptOut$' -count=1 -v
exit:   0
stdout: testing: warning: no tests to run
        PASS
        ok  	github.com/modu-ai/moai-adk/internal/cli	1.011s [no tests to run]
reason: red by the PASS-line rule; control as L-13
```

```
L-20  AC-015 (c)   [carried from E-11; tree 7109e0900]
cmd:    grep -c MOAI_SKIP_PLUGIN_INSTALL internal/config/envkeys.go
exit:   1 (grep's status for a zero count; the tool shows it as completed)
stdout: 0
```

```
L-21  AC-016   [tree 3766cef05]
cmd:    go test ./internal/cli -run '^TestDoctorAgentEmitEmbed_SetsPluginOptOut$' -count=1 -v
exit:   0
stdout: testing: warning: no tests to run
        PASS
        ok  	github.com/modu-ai/moai-adk/internal/cli	0.882s [no tests to run]
reason: red by the PASS-line rule; control as L-13
also:   go test ./internal/cli -run '^TestPluginOptOutCallersEnumerated$' -count=1 -v -> the same three lines, ok ... 0.994s [no tests to run]
also:   grep -c MOAI_SKIP_PLUGIN_INSTALL e2e/cli/tux3_journeys.sh -> stdout 0 (status 1, not echoed by the tool)
```

```
L-22  AC-017   [tree 3766cef05]
cmd:    go test ./internal/cli -run '^TestPluginInstallStep_NoRealRunnerUnderTest$' -count=1 -v
exit:   0
stdout: testing: warning: no tests to run
        PASS
        ok  	github.com/modu-ai/moai-adk/internal/cli	0.863s [no tests to run]
reason: red by the PASS-line rule; control as L-13
```

```
L-23  AC-018 (a)   [tree 3766cef05]
cmd:    sh scripts/test-plugin-install-step.sh bin/moai
exit:   127
stdout: sh: scripts/test-plugin-install-step.sh: No such file or directory
```

```
L-24  AC-018   [carried from E-12; tree 7109e0900]
cmd:    grep -c MOAI_SKIP_PLUGIN_INSTALL install.sh
exit:   1 (grep's status for a zero count)
stdout: 0
control: grep -c print_info install.sh -> 14, exit 0
```

```
L-25  AC-018 (c), static   [tree 3766cef05]
cmd:    go test ./internal/cli -run '^TestInstallScriptsPluginStepGuarded$' -count=1 -v
exit:   0
stdout: testing: warning: no tests to run
        PASS
        ok  	github.com/modu-ai/moai-adk/internal/cli	0.711s [no tests to run]
reason: red by the PASS-line rule; control as L-13
```

```
L-26  AC-018 (a), surrogate for the missing case   [tree 3766cef05]
cmd:    PATH=<scratch>/inst/shim:/usr/bin:/bin /bin/bash install.sh --version 9.9.9 --install-dir <scratch>/inst/bin
        (stub curl serving <scratch>/inst/pin/moai-adk_9.9.9_darwin_arm64.tar.gz; stub claude, codex and moai recording their argv)
exit:   0 (the script reached "[SUCCESS] Installation complete!")
stdout: (excerpt, colour codes omitted) [SUCCESS] Detected platform: darwin_arm64 ... [WARNING] Failed to download checksums
        (verification skipped) ...
        [SUCCESS] Installed to: <scratch>/inst/bin/moai
        [WARNING] Installation completed, but 'moai' command not found in PATH
        [SUCCESS] Installation complete!
observed: no recorded call: <scratch>/inst/moai-calls.log and <scratch>/inst/tool-calls.log do not exist; curl.log holds two
          lines, both answered by the stub. This is the red the case `installer-calls-verb-by-installed-path` would report
          ("recorded plugin calls = 0"); it also shows the install directory is not on PATH, so the call must use the path.
note:   a command that prefixed HOME=<dir> was refused by the worktree guard (SPEC P-36); this run set only PATH.
```

```
L-27  AC-019 (b)   [tree 3766cef05]
cmd:    <tree-built-moai> plugin install --help
exit:   1
stdout: (excerpt; the tool printed blank padded lines around these) ERROR
        Unknown command "plugin" for "moai".
        Try --help for usage.
```

```
L-28  AC-019 (a)   [tree 3766cef05]
cmd:    go test ./internal/cli -run '^TestPluginInstallCmd$' -count=1 -v
exit:   0
stdout: testing: warning: no tests to run
        PASS
        ok  	github.com/modu-ai/moai-adk/internal/cli	0.820s [no tests to run]
reason: red by the PASS-line rule; control as L-13
```

```
L-29  AC-020   [tree 3766cef05]
cmd:    go test ./internal/cli -run '^TestCheckPluginVersion_ClaudeRead$' -count=1 -v
exit:   0
stdout: testing: warning: no tests to run
        PASS
        ok  	github.com/modu-ai/moai-adk/internal/cli	1.042s [no tests to run]
reason: red by the PASS-line rule; control as L-13
```

```
L-30  AC-021   [tree 3766cef05]
cmd:    go test ./internal/cli -run '^TestCheckPluginVersion_CodexRead$' -count=1 -v
exit:   0
stdout: testing: warning: no tests to run
        PASS
        ok  	github.com/modu-ai/moai-adk/internal/cli	0.941s [no tests to run]
reason: red by the PASS-line rule; control as L-13
```

```
L-31  AC-022   [tree 3766cef05]
cmd:    go test ./internal/cli -run '^TestCheckPluginVersion_Outcomes$' -count=1 -v
exit:   0
stdout: testing: warning: no tests to run
        PASS
        ok  	github.com/modu-ai/moai-adk/internal/cli	0.840s [no tests to run]
reason: red by the PASS-line rule; control as L-13
also:   CLAUDE_CONFIG_DIR=<scratch>/red/c-ac022 CODEX_HOME=<scratch>/red/x-ac022 <tree-built-moai> doctor --check "Plugin Version" ->
        ╭────────────────────────────────╮
        │                                │
        │  System Diagnostics            │
        │   Pass 0    Warn 0    Fail 0   │
        │                                │
        ╰────────────────────────────────╯
        (no row, no error; the tool result carried no error marker, status not echoed). So (b) is red by the missing name,
        and an unregistered check would satisfy any assertion made on the exit status alone
control: the same command with the registered name "Plugin Deployment" under the same two scratch homes printed the progress lines
        `○ Plugin Deployment` and `✓ Plugin Deployment`, the table row `ok      Plugin Deployment  no plugin marker (binary-managed)`,
        `1 ok, 0 warn, 0 fail` and `Pass 1    Warn 0    Fail 0`: a registered check leaves its name in the output
```

```
L-32  AC-023 (a)   [tree 3766cef05]
cmd:    grep -c 'Plugin Version' internal/cli/testdata/doctor-nocolor.golden
exit:   1 (grep's status for a zero count)
stdout: 0
also:   the same on doctor-light.golden -> 0; on doctor-dark.golden -> 0
control: grep -c 'MCP Server Version' internal/cli/testdata/doctor-nocolor.golden -> 1, and 1 in each of the other two goldens
```

```
L-33  AC-023 (c) and (d)   [tree 3766cef05]
cmd:    go test ./internal/cli -run '^TestCheckPluginVersion_OutputBounded$' -count=1 -v
exit:   0
stdout: testing: warning: no tests to run
        PASS
        ok  	github.com/modu-ai/moai-adk/internal/cli	0.731s [no tests to run]
also:   go test ./internal/cli -run '^TestDoctorGolden_IgnoresCallerCodexHome$' -count=1 -v -> the same three lines, ok ... 1.014s [no tests to run]
also:   grep -n 'CODEX_HOME\|codexHomeEnvVar' internal/cli/doctor_golden_test.go -> no line selected (status 1, not echoed by the tool)
```

```
L-34  AC-024 (a) and (c)   [tree 3766cef05]
cmd:    go test ./internal/template/pluginemit -run '^TestPluginVersionScript$' -count=1 -v
exit:   1
stdout: # ./internal/template/pluginemit
        stat /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435/internal/template/pluginemit: directory not found
        FAIL	./internal/template/pluginemit [setup failed]
        FAIL
also:   go test ./internal/template/pluginemit -run '^TestReleaseWorkflowCallsPluginVersionCheck$' -count=1 -v -> the same output, exit 1
```

```
L-35  AC-024 (b)   [tree 3766cef05]
cmd:    sh scripts/check-plugin-version.sh v9.9.9
exit:   127
stdout: sh: scripts/check-plugin-version.sh: No such file or directory
```

```
L-36  AC-024 (c) static and (d)   [tree 3766cef05]
cmd:    grep -n check-plugin-version .github/workflows/release.yml
exit:   1 (no line selected; status not echoed by the tool)
stdout: (none)
also:   grep -n 'plugin-emit' .moai/docs/version-management.md -> no line selected
also:   grep -n 'check-plugin-discoverable' .moai/docs/version-management.md -> no line selected
```

Gaps in this ledger: L-05, L-06, L-09 and the `pluginemit` entries marked "also" are package-absent reds (tool-failure class);
the first run-phase RED record must show an assertion failure. The Go-test reds of `internal/cli` read `[no tests to run]`
with exit 0, which establishes absence of the tests and nothing about whether a runner would later select them; the named
PASS line is what closes that, and the control in L-13 shows the same command form prints the PASS line for an existing test.
The `grep` entries, and the `find` of L-11, record status 1 from the documented behavior for a zero count or an unreadable
root: the tool result shows such a command as completed and does not echo the status. L-26 is a surrogate observation of the base `install.sh`, not the cell of a command
this criterion cites; the harness script that will cite it does not exist yet. Scratch fixtures (marketplaces, homes, stub
tools, the pinned archive) were built by the author in the scratchpad and are not part of the repository.
