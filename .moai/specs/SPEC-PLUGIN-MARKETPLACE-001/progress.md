# Progress — SPEC-PLUGIN-MARKETPLACE-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-03 (final delta for ND-1 to ND-3 after plan-audit iteration 3, FAIL 0.86 at a0c8ad7cb; the final delta is not yet audited)
tier: L
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md (progress.md not counted; decision-index.md is authored because `interview.decision_gate` is `on`)
budget: 25 requirements, 25 acceptance criteria (Tier L ceilings 25 and 25: both at the ceiling, so any further requirement means splitting the SPEC)
plan_base_sha: 7109e0900 (branch WT-marketplace-core-plugin, base develop); the iteration-3 delta was made on tree b6a0522a0, which differs from the tree plan-audit iteration 2 audited (d6987e59c) in this file only; the final delta (ND-1 to ND-3) was made on tree cc46749d9, which differs from the tree plan-audit iteration 3 audited (a0c8ad7cb) in this file only (`git diff --name-only a0c8ad7cb cc46749d9`)
run_start_sha: pending — set by the orchestrator to the commit that carries the final plan-phase revision of the artifacts
open_decisions: OD-1 to OD-14 in spec.md §5, mirrored as Q1 to Q14 in decision-index.md; verdict lines empty; every default-bound clause carries a `default pending OD-n` marker and spec.md §5 holds the marker table (checked both ways, see "Verification of the iteration-3 revision")
inputs: .moai/reports/t1435/inputs/t1434-verdict.md (local-only, `.gitignore:235`); the committed carrier of the same verdict is .moai/specs/SPEC-PLUGIN-LOAD-SCOPE-001/progress.md section E.2; .moai/reports/t1435/plan-audit-iter1.md (iteration 1 verdict, FAIL 0.72); .moai/reports/t1435/plan-audit-iter2.md (iteration 2 verdict, FAIL 0.84; local-only like the first)
scratch_evidence: the revision's scratch-home probes live in the session scratchpad and are not part of the repository; their commands and deciding output are in research.md and acceptance.md (evidence ledger)

### Tier decision

The SPEC is Tier L because plan-audit iteration 1 ruled it so (defect D7): the hand-authored change set is about 42 files against the
Tier L guidance of more than 15 (`plan.md` §3 lists them), and the Tier M choice had been justified in the draft only by "to match the
requested artifact set", which is not a tier criterion. The tier was not chosen to gain audit iterations. Its consequences are the ones
the ruling names — `design.md` and `research.md` are required, the threshold rises from 0.80 to 0.85, and the requirement and criterion
ceilings are 25 each — and one that is incidental: `harness.yaml` `plan_audit_tier_ceilings.L` is 3 spawns against 2 for Tier M. This
revision carries 24 requirements and 24 criteria, so it is under the ceilings by one each, with the packed requirements of iteration 0
un-packed.

### Dispatch instructions (recorded for the run-phase delegation)

- **Land order with t1399 (plan-audit D21).** The card whose merge lands second turns `make plugin-emit-check` red (the golden test fails
  in the full suite) until `make plugin-emit` is run. t1399's lane scopes its verification to its own selectors, so the red would otherwise
  surface on `origin/develop` CI after the leader's batch push and t1399 is told nothing. Required leader line at dispatch, to the lane that
  lands second: "run `make plugin-emit` inside your merge and commit the delta (`plugins/moai/**` and the four manifests)". Owner: the lane
  landing second. This line is a recorded requirement; it has not been delivered to any lane by this revision.
- **Installer cases and install directory (plan-audit iteration 3, ND-2; binding for the run-phase delegation).** Every `installer-*` case
  of `scripts/test-plugin-install-step.sh` passes `--install-dir <scratch>/inst-<case>/bin` and asserts the installed path under the
  scratch root (`go` resolves to the harness stub, `realpath <install dir>/moai` lies under the case's own directory, no decoy default
  root holds a `moai`); the harness's stub `go` returns two decoy directories inside the scratch for `go env GOBIN` and
  `go env GOPATH`. Without the flag `install.sh` installs into `$GOBIN`, `$GOPATH/bin` or the real `$HOME/.local/bin`, none of which the
  protected-set hash covers (REQ-025, AC-018 (a), AC-025 (f), design §6, plan M3 and §6). A case written without the flag must be a red.
- **Discoverable script (ND-3).** `scripts/check-plugin-discoverable.sh` is the one script that runs the real `claude`; it starts exactly
  `plugin marketplace add <local repo path>`, `plugin install moai@moai-adk` and `plugin details moai@moai-adk`, behind the same scrub
  (marker comments `# scrub:begin` / `# scrub:end`), from a scratch directory, with a scratch home it refuses unless empty (REQ-025,
  AC-006 (b), (d), (e)).

### Verification of this revision

- Lint: `moai-t1435 spec lint SPEC-PLUGIN-MARKETPLACE-001 --strict` printed `✓ No findings — all SPEC documents are valid`. The first run, before
  one ledger line was reworded, printed one `VacuousTestAssertion` warning at the old line 778 and `0 error(s), 1 warning(s)`.
- Judging build: `go build -o <scratchpad>/s2/moai-t1435 ./cmd/moai` at HEAD `3766cef05`. It was built without ldflags and self-reports `moai-adk v3.1.3`,
  `v3.1.3   none   built unknown`, so its commit is attributed by build procedure and not self-attested. The installed `moai` is v3.2.0-rc.26 and was
  not used for the lint. The `moai` MCP server that answered `spec_audit` reported build v3.2.0-rc.23 (commit d194083fb), older than the tree; it
  was used for the era classification only (`EraAutoDetected`, V3R6, `modern_era_clean: 1`, one INFO finding).
- Counts: `grep -c '^- REQ-0' spec.md` = 24; `grep -c '^### AC-0' acceptance.md` = 24; `grep -n 'NEEDS CLARIFICATION'` over the directory selected no line.
- Tree: `git rev-parse --short HEAD` = `3766cef05` and `git branch --show-current` = `WT-marketplace-core-plugin` at the end of the revision, as at the start;
  `git status --short` listed only files of this SPEC directory. Nothing was committed.

### Verification of the iteration-3 revision

`<s>` is the session scratchpad `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/2f10c8c5-67ea-41c2-9b61-6242acc465c3/scratchpad`.

- Lint: `CLAUDE_CONFIG_DIR=<s>/i3/lint-claude CODEX_HOME=<s>/i3/lint-codex MOAI_HOME=<s>/i3/lint-moai <s>/i3/moai spec lint
  SPEC-PLUGIN-MARKETPLACE-001 --strict` printed `✓ No findings — all SPEC documents are valid`, on the edited files, and again after
  the last edit of the delta. Judging build: `go build -o <s>/i3/moai ./cmd/moai` at HEAD `b6a0522a0`, built without ldflags (the
  earlier builds from this tree self-reported `moai-adk v3.1.3` and `none`, so its commit is attributed by build procedure and not
  self-attested; this build's own `version` output was not read). The installed `moai` (v3.2.0-rc.26) was not used.
- Counts: `grep -c '^- REQ-0' spec.md` printed `25`; `grep -c '^### AC-0' acceptance.md` printed `25`; `grep -c 'default pending OD-' spec.md`
  printed `24` lines (a line count, which does not test completeness).
- Marker completeness, both sides: `python3 <s>/i3/markcheck.py .moai/specs/SPEC-PLUGIN-MARKETPLACE-001` printed `checked=30
  requirement/criterion pairs, rows=14, REQ lines=25, AC sections=25, failures=0`; each of the 30 pair lines reads
  `req-marker=True AC-nnn-marker=True AC-nnn-alternate=True -> ok`. Controls: a copy with one marker and one `Alternate` removed printed
  `failures=2`; the files at `d6987e59c` printed `failures=11`.
- Ordering (the plan-auditor's CN-4 form, `<s>/cn4.sh`): `COLLECTED: 4 milestones in plan order (M1 M2 M3 M4), 14 exit bindings, 58 ordering
  candidates`, no `CONFLICT` line. (The audit's run read 13 bindings and 41 candidates before AC-025 and the added text.)
- `grep -rn 'plan-audit\.md' .moai/specs/SPEC-PLUGIN-MARKETPLACE-001` printed no line.
- Tree: `git rev-parse --short HEAD` printed `b6a0522a0` and `git branch --show-current` printed `WT-marketplace-core-plugin` at the start and at
  the end of the delta. `git status --short` listed seven modified files, all under `.moai/specs/SPEC-PLUGIN-MARKETPLACE-001/`
  (`spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `decision-index.md`, `progress.md`); nothing was staged or committed,
  and no agent-memory file was written.
- Scratch work, none of it in the repository: `i3/scan` (the name scan), `i3/good` and `i3/bad1..3`, `i3/emb` (the embed mode program),
  `i3/pl-good` and `i3/pl-bad`, `i3/h` (the stand-in harness, `protected-set-hash.sh` and a stub product), `i3/ctl`, `i3/mc` and `i3/old` (the
  marker-check controls), `i3/markcheck.py`, `i3/moai`. Stand-ins use stub tools only; no real `claude`, `codex` or `moai init` was run; the
  only product-binary run was the lint, with scratch `CLAUDE_CONFIG_DIR`, `CODEX_HOME` and `MOAI_HOME` set.

### Audit-iteration-1 disposition

The author claimed that every defect of `.moai/reports/t1435/plan-audit-iter1.md` was fixed; plan-audit iteration 2 found D3 only partly
fixed (see the iteration-2 disposition below, which also corrects the D3 row of this table). "Evidence" names the re-run that the fix
rests on, in `research.md` (R-nn) or `acceptance.md` (L-nn).

| Defect | Disposition | Fixed where | Evidence |
|--------|-------------|-------------|----------|
| D1 REQ-010 contradicts AC-010, edge cases, plan M3 | fixed | REQ-013 (non-zero exit, timeout, invalid pin: one guidance block) and REQ-014 (absent tool: exactly one skip line, no block) replace the old REQ-010; AC-013 and AC-014 assert the same split; the edge-case list, plan M3 notes and design §3.2 say the same | the four locations re-read after the edit |
| D2 four decisions outside the OD table | fixed | OD-9 harness gating, OD-10 target profile, OD-11 payload language, OD-12 scope, with options, a default and evidence (spec §5, decision-index Q9 to Q12); REQ-010, REQ-012, REQ-005, REQ-006, REQ-011 follow them; the "`--llm` does not gate the step" sentence is gone from the plan. One more decision, OD-13 (the doctor's Codex read path), was added because the D10 measurement found a side effect | P-28 (R-09: `init.go:133`, `:192-195`, `:997-1003` read), P-29 (R-10), P-32 (R-08: 18 locale conditionals, all on `description`/`argument-hint` lines), P-33 (R-07: `--scope` flags), P-34 (R-05) |
| D3 defaults hard-coded without markers | **partial, corrected by plan-audit iteration 2 and completed in iteration 3** | 14 requirements carried a `default pending OD-n` marker (REQ-001, 003, 004, 005, 006, 007, 008, 010, 011, 012, 015, 018, 021, 022); the §5 marker table named each clause and what the other verdict changes; the Definition of Done requires a verdict or the line `default adopted` per row and a delta re-audit. The sentence that stood here, "every affected criterion carries an `Alternate` line", was false: nine default-bound criteria carried no marker and three no `Alternate` line | `grep -c 'default pending OD-' spec.md` = 17 lines at the time (the 14 requirement lines, the HISTORY row, the §2 preface and one §3 constraint), a count that does not test completeness; the iteration-2 disposition below carries the check that does |
| D4 AC-013(b) reaches the real Codex home | fixed | AC-022 (b) sets `CLAUDE_CONFIG_DIR` and `CODEX_HOME` to scratch homes; AC-023 (d) `TestDoctorGolden_IgnoresCallerCodexHome` and its static companion; plan M4 lists `internal/cli/doctor_golden_test.go` as modified (`captureDoctorCmd` gains the `CODEX_HOME` scrub). Note: the harness already pins `HOME` (`:111`), so only a `CODEX_HOME` set in the caller's environment escapes it | P-21, R-20 |
| D5 P-20 false, AC-011(c) unsatisfiable | fixed | P-20 recounted: 33 `runInit(` lines in 18 test files, one a string literal, so 32 call sites; 9 `initCmd.RunE(` sites in 5 files; one `exec.Command` re-entry; two e2e invocations. REQ-016 (automated callers set the opt-out, e2e script named), REQ-017 (one mechanism, no per-site edit), AC-016 (enumerating guard test), AC-017 (pinned binary and PATH shims untouched); RK-8 and plan M3 name the e2e script | R-11, R-12 |
| D6 undiscoverable commands, layout unspecified | fixed | REQ-006 (flat `commands/<name>.md`, no subdirectory); AC-006 (layout test, scratch-home inventory script, refusal of a non-empty argument); names a user sees stated in spec §1.4 with the unobserved caveat. The auditor's observation reproduced on the core-only payload shape: flat `Skills (41)`, nested `Skills (24)` | P-30, R-01 |
| D7 Tier L required | fixed | tier L; `design.md` and `research.md` added; REQ-010/011/013 un-packed (24 requirements, ceiling 25); the tier decision above | the ruling, `spec-workflow.md` § SPEC Complexity Tier |
| D8 AC-012 verifies tokens; verb unspecified | fixed | AC-018: `scripts/test-plugin-install-step.sh` drives `install.sh` offline with a stub `curl`, a pinned local archive and stub tools, and requires four named cases (installed-path call, opt-out, `set -e` guard, old binary without the verb); `install.ps1` and `install.bat` are checked statically and labelled static; REQ-019 and AC-019 give the verb a name (`moai plugin install`, proposed), exit codes and a help surface. `pwsh` exists but the worktree guard refuses it, recorded as Gap G-3, not as an absence | P-36, P-37, P-39, R-13, R-14 |
| D9 AC-014 passes without registration | fixed | AC-023 (a): `grep -c 'Plugin Version'` on each of the three goldens must print 1 (RED prints 0; the control prints 1 for an existing row); AC-022 (b) requires the name in the output because an unregistered `doctor --check` prints `Pass 0 Warn 0 Fail 0`; `TestBinaryLag_AllowlistKeysAreLiveNames` is cited as the binding for the allowlist key | P-38, R-15 |
| D10 Codex read path from the cache directory | fixed | REQ-021 and AC-021: state from `codex plugin list --json`, bounded, never from the cache directory; subtests `cache-without-registration` and `multiple-cache-versions`. Re-run: after the stanza was deleted the list printed `"installed": []` while `plugins/cache/moai-adk/moai/3.1.3` remained. The same run found that starting the CLI creates `tmp/arg0` in an empty home, which became OD-13 (option (b) reads `config.toml` and starts no process) | P-34, R-05 |
| D11 Codex marketplace has no version field | fixed | REQ-002 and AC-002 (c): no `version` key in the Codex marketplace; REQ-003: the four version-carrying fields (Claude marketplace `metadata.version` and entry, the two plugin manifests); AC-003 (d); the Codex plugin version lives in `.codex-plugin/plugin.json` | P-35, R-06 (Codex marketplace 0 `version` keys; Claude marketplace 19; 18 of 18 Codex plugin manifests carry one) |
| D12 "a bounded time" has no value | fixed | `config.DefaultPluginInstallCommandTimeout` = 60 seconds named in REQ-013 and asserted by AC-013 (a) `timeout` and (b); the doctor probe is `config.DefaultPluginVersionProbeTimeout` = 3 seconds in REQ-021 and AC-021; the worst case (4 x 60 s) is RK-15 | `internal/config/defaults.go` holds the existing timeout constants (read) |
| D13 Q1 says listing cost cannot be measured offline | fixed | decision-index Q1 and OD-1, §1.5-1 and P-31 carry the figures: core-only about 3,653 always-on tokens (41 entries), skills alone about 3,286, all tiers about 6,483, each the tool's estimate; Q1 narrowed to shadowing, collision and the scaffold's own listing cost | R-02, R-03 |
| D14 REQ-005 overclaims | fixed | REQ-005 now says no rendered payload file carries a template action or locale-conditional text and that non-`.tmpl` brace-bearing files (16 skill files) are copied unchanged; AC-005 (c) searches only the rendered commands | P-15, P-32 |
| D15 AC-004(b) and AC-016(a) rot | fixed | AC-004 (b) is `TestGeneratorHoldsNoComponentNames`, derived from the catalog and the command stems with a positive control; the release-script criterion (now AC-024 (a)) reads `version.Version` in a test instead of pinning `v3.1.3` | acceptance.md AC-004, AC-024 |
| D16 shared scratch home | fixed | one fresh empty home per criterion, `<claude-home:AC-nnn>` and `<codex-home:AC-nnn>`, never reused across criteria | acceptance.md Conventions |
| D17 subtest naming gaps | fixed | AC-009 (b) names four subtests including `committed-set-unchanged`; AC-012 names `prints-config-home-claude` and `prints-config-home-codex`; every Go criterion lists its subtests | acceptance.md |
| D18 PATH shim misses a pinned binary | fixed | AC-017 `pin-and-path-shims-untouched` pins `MOAI_CLAUDE_BIN` to a recording script and also shims `claude` and `codex` on PATH, with scratch homes | acceptance.md AC-017 |
| D19 cross-reference slips | fixed | RK-4 cites `plan.md` §5; plan M3 cites `init.go:1011` for the call and `:192` for the definition; P-19 and §1.5-2 add that `--llm both` forces provisioning (`:997-1003`) and `--llm gpt` declines it | R-09 |
| D20 `-list` surrogate in the RED cells | fixed | every Go-test RED cell quotes the criterion's own command and its `testing: warning: no tests to run` / `[no tests to run]` output, red by the PASS-line rule, with a control (`TestBinaryLag_AllowlistKeysAreLiveNames` prints `--- PASS:`); the package-absent reds are disclosed as tool-failure class | L-13 and the ledger gaps paragraph |
| D21 owner of the red when t1399 lands second | fixed | plan §5 names the lane landing second as owner (runs `make plugin-emit` in its merge) and the leader line is recorded under "Dispatch instructions" above; RK-4 points at both | plan §5 |
| D22 P-11 omits `mcpServers` and `interface` | fixed | P-11 states them with the re-measured counts (12 of 18 Codex manifests carry `mcpServers`, 18 of 18 carry `interface`); plan M1 gives the Codex plugin manifest an `interface` block with the precedent's keys and golden-pinned values | R-06 |
| D23 release runbook not in the change set | fixed | plan M4 modifies `.moai/docs/version-management.md` (a "Generated version carriers" group: the bump step `make plugin-emit`, the pre-tag step `check-plugin-discoverable.sh`); AC-024 (d) pins both lines. The release harness's own runbook was not located in the tree (G-6) | P-40, R-19 |
| D24 small items | fixed | `timeout` is gone from AC-002 (a) (stock macOS has none); OD-7 and the edge cases state that a maintainer's `make build` against the fallback `v3.1.3` warns at the default; the legacy `Plugin Deployment` check beside `Plugin Version` is noted in design §4 and AC-023 (a) states that the names do not collide | R-15 (`doctor-nocolor.golden:31`) |

Totals: 24 defects, 24 fixed, 0 rejected, 0 skipped. Where a fix departs from the auditor's suggested remedy: D10 asked for the installed
state to come from the `config.toml` registration; this revision followed the orchestrator's instruction to read it through
`codex plugin list --json` (which answers from that registration) and kept the `config.toml` read as OD-13 option (b) because of the
`tmp/arg0` side effect.

### Audit-iteration-2 disposition

Source: `.moai/reports/t1435/plan-audit-iter2.md` (FAIL 0.84 at audited_sha d6987e59c; Tier L threshold 0.85). Delta scope approved by the
leader: the defects below only, no renumbering of requirements or criteria, no new feature. D1 to D24 other than D3 were confirmed FIXED by
that audit and are untouched. Evidence ids: R-nn in `research.md`, L-nn in `acceptance.md`; tree `b6a0522a0` unless stated; every scratch
fixture ran in the session scratchpad and no command wrote to a real profile or home.

| Id | Disposition | Fixed where | Evidence |
|----|-------------|-------------|----------|
| D3 residue (nine criteria without the marker; three without an `Alternate`; four requirements unmarked; progress.md claim) | fixed | markers on AC-004, 005, 006, 007, 009, 018, 019, 020, 024; `Alternate` lines added to AC-009, AC-020, AC-024 (and to AC-002, found by the check below); REQ-002, 009, 020, 024 marked (REQ-004 for OD-3 and REQ-019 for OD-9 too); marker table: OD-4 gains REQ-002, REQ-009, REQ-024, OD-12 gains REQ-020, OD-9 gains REQ-019; the D3 row of the iteration-1 table above corrected | `python3 i3/markcheck.py <spec dir>`: `checked=30 requirement/criterion pairs, rows=14, REQ lines=25, AC sections=25, failures=0` (it reads each table row, then the REQ line and the `### AC-nnn` section, whitespace flattened). Its first run on the edited files found two gaps I had left (AC-002 `Alternate`, AC-003 marker for OD-4), fixed before the run quoted. Control: a copy with one marker and one `Alternate` removed printed `failures=2`. Run on the audited tree (`git show d6987e59c:…` copies) it printed `failures=11` — more than the audit's nine criteria because it also counts a capital-D `Default pending` and the REQ side |
| N1 REQ-004 and AC-004 (b) unsatisfiable by correct code | fixed | REQ-004 (exempt identifier `moai`, whole literal or `/`-segment match); AC-004 (b) with the scan defined (what is scanned, tokenisation, the set N and the exemption, in-test positive control, empty-sweep failure); design §2.4; plan §3 M2 | R-27 and L-38: a correct stand-in generator printed `PASS`; three hard-coded mutants (`"moai-foundation-core"` as a map key, as a path, a lone `"gtd"`) printed `FAIL`, exit 1; the real `commandemit` (3 files, 81 literals) and `agentemit` (6 files, 267 literals) printed `PASS`; the real `internal/template` printed `FAIL 62 component-name literal(s)` |
| N2 AC-023 (b) would start a real `codex`; home unpinned | fixed | REQ-017 (one runner for the step and the doctor probe), REQ-021 (home from the existing seam, handed to the child as `CODEX_HOME`); AC-021 (c) `TestCheckPluginVersion_HomeIsolation` with a canary directory, a recording-shim registry run, and a `TestMain` guard; AC-023 (b) note; plan §3 M4 names `internal/cli/main_test.go:269-299` and `doctor_plugin_version.go`; design §3.4 and §4 | R-24: `mcp_codex.go:2138` `var codexUserHomeDir = os.UserHomeDir`; `grep -n 'CODEX_HOME\|codexHomeEnvVar\|codexUserHomeDir' internal/cli/main_test.go` printed nothing; seven unfiltered `runDiagnosticChecks(false, "")` sites in six files (the audit named three tests); seven files assign `codexUserHomeDir` and restore the captured value. L-30 also-line: `[no tests to run]` for the new test |
| N3 harness leaves the Claude pin and project unscrubbed | fixed | REQ-025 and AC-025: poison planted first, live-enumerated scrub of `MOAI_*`, `CLAUDE_*`, `CODEX_*`, pid-named variables so no typed list can pass, scratch working directory with no `.moai` above it, isolation cases, three negative controls; design §6; plan §3 M3 and §4 | R-25 (resolver order read at `claude_binary.go:33-48`), R-28 and L-39: normal run `RESULT pass=5 fail=0` after `scrub: enumerated and unset 37 names`; scrub disabled: four `RED`, one `green`, recorder executed twice, shim zero; pin from the project alone: two `RED`, three `green`; typed-list scrub: `RED isolation-env-scrubbed` alone |
| N4 harness sets `HOME` in a script file | fixed (design) and one decision opened | `HOME` removed from design §6, acceptance Conventions, plan §4 and §6, spec §3; new OD-14 with Q14 in `decision-index.md` and a marker row; the `init-*` harness lines of AC-010, 012, 013, 014, 015 moved to Alternate lines; Gap G-8; REQ-025 forbids a `HOME` assignment and AC-025 (e) greps for one | R-23: `init.go:881-892` and `autonomy_bundle.go:71-83` write `$HOME/.claude/settings.json` whatever `CLAUDE_CONFIG_DIR` says, so the nine `init-*` cases have no isolated form without `HOME`; `paths.go:53-58` is HOME-first and `MOAI_HOME` moves `~/.moai` only. The grep of AC-025 (e) printed no line on the stand-in scripts, and on a scratch file holding `CODEX_HOME=`, `HOME=`, `export HOME` and `env HOME=` it matched exactly the last three |
| N5 AC-003 (c) bound to M1, MCP unit in M2 | fixed | plan §3: `mcp.go` and `mcp_test.go` (`TestMCPEntryDerivedFromTemplate`) moved into M1, removed from M2; file count recounted (7 + 5 + 20 + 13 = 45); AC-003 (c) and AC-007 green-path and table text agree; exit lines unchanged | plan M1 file list and M2 file list re-read after the edit; the Codex manifest of REQ-003 needs the entry, so the unit cannot be later than the manifest |
| N6 OD-3 (a) says "all 12" | fixed | OD-3 (a): every agent file the OD-8 tier admits, 11 at the default `core`, 12 only if OD-8 admits every tier; decision-index Q3 title | `sed -n '126,181p' internal/template/catalog.yaml | grep -c -- '- name:'` printed `11` (the `core` agents section); the `harness_generated` section lists one agent, `builder-harness`; the template agent directory holds 12 files; R-22 |
| N7 OD-11 (c) treated three ways | fixed | marker table OD-11 row, AC-006 Alternate, Definition of Done and AC-005 (c) now say one thing: REQ-006 void, AC-006 (a) and the command half of (b) do not apply, its skills half and (c) stay, AC-005 (c) has no directory, AC-008 loses `commands`; OD-11 row (c) of §5 | the four places re-read after the edit |
| N8 payload drops the executable bit | fixed | REQ-005 (mode rule), REQ-009 (mode difference is drift); AC-005 (a) and (d); AC-009 (b) gains `mode-flipped`; design §2.2 and §2.4 (the 0644 sentence is gone); plan M2; P-45 | R-26 and L-37: `deployer.go:275-276` gives `.sh` 0755; git modes `100755`, `100755`, `100644`; a scratch `//go:embed` program printed `-r--r--r--` for a mode-755 and a mode-644 file, so no source mode exists to copy and the suffix rule is the only derivable one; the check form `find <tree> -name '*.sh' ! -perm 755 -print` printed the one 0644 source, three paths on a 0644 stand-in payload, nothing on a 0755 one, and the control printed three |
| N9 `timeout` subtest reads as a 60-second wait | fixed (trivial) | AC-013 (a): the step takes its bound as an injected parameter, `timeout` runs with a few milliseconds, new subtest `bound-equals-constant` pins that production passes `config.DefaultPluginInstallCommandTimeout` | text re-read; the constant's value stays pinned by (b) |
| N10 two RED-now cells not the criterion's own command | fixed (trivial) | L-10 re-measured with AC-007 (a) and (b); L-24 labelled a surrogate and dropped from AC-018's RED-now list | `jq -c .mcpServers.moai plugins/moai/.mcp.json internal/template/templates/.mcp.json` exit 2, `Could not open file plugins/moai/.mcp.json`, and the template entry on stdout; `jq -c '.mcpServers|keys' plugins/moai/.mcp.json` exit 2, same message |
| N11 iteration-1 report cited by a path that no longer exists | fixed | the old bare report name is replaced by `plan-audit-iter1.md` in spec.md, progress.md, research.md and acceptance.md; research sources also cite `plan-audit-iter2.md` | `grep -rn 'plan-audit\.md' .moai/specs/SPEC-PLUGIN-MARKETPLACE-001` matched six places before the edits (the audit's list: spec.md:L25, progress.md:L13 and L48, research.md:L5 and L341, acceptance.md:L10) and, after them, the command is re-run in the verification block below |
| N12 REQ-018, REQ-022, REQ-023 still pack obligations | **skipped, with reason** | none | `grep -c '^- REQ-0' spec.md` printed `25` and `grep -c '^### AC-0' acceptance.md` printed `25`: both at the Tier L ceiling (25 and 25). Splitting a requirement adds a requirement past the ceiling and, by the one-to-one numbering, a criterion, so it is not a pure text split; the auditor classed it optional and "split only if the criteria become hard to bind", and each of the three criteria binds its clauses to named subtests |
| N13 M4 remeasure list omits `version_sync_list_test.go` | fixed | plan M4 modified-list (the new group opens under `**Generated version carriers:**`, never as a bullet of the Version Stamps list) and §5 remeasure list; AC-024 (e) preservation guard | `go test ./internal/cli -run '^TestVersionSyncListNamesOnlyExistingPaths$' -count=1 -v` printed `--- PASS: TestVersionSyncListNamesOnlyExistingPaths (0.00s)` and `ok … 1.365s` (L-40); `version_sync_list_test.go:30-60` read: section ends at the next bold label or `###` heading, exactly seven entries |

Totals: 14 items (D3 residue and N1 to N13): 13 fixed (N4 fixed in design with OD-14 opened for the one case that cannot be isolated),
1 skipped (N12) with the reason and the counts above. New requirement: REQ-025; new criterion: AC-025; new Open Decision: OD-14 (Q14);
new premise rows P-42 to P-46; new research rows R-23 to R-29; new ledger cells L-37 to L-40; new gaps G-8 and G-9. Final counts: 25
requirements, 25 acceptance criteria, 14 Open Decisions.

Where a fix departs from the auditor's suggested remedy. N1: the audit offered scoping the scan to selection data or a synthetic-tree-only
test; this revision kept both and defined the scan precisely, because the synthetic-tree test alone cannot see one hard-coded name that
the synthetic names do not trigger (mutant probe of AC-004). N2: the audit offered a refusing seam for the probe or a stubbed PATH seam
with a pin; this revision did both (one shared refusing runner, and a `TestMain`-level redirect rather than seven per-test pins). N4: the
audit offered a reduced harness or an Open Decision; this revision did the first for everything the explicit seams can isolate and the
second for the nine `init-*` cases. N5: the audit offered either direction; `mcp.go` moved into M1, with AC-007 left in M2 as the audit
suggested.

### Gaps of this revision

- No criterion's own command exists yet; every RED-now cell is the observed red of a command whose target is absent. L-26 is a surrogate
  (the base `install.sh` driven offline), not a cell of a command a criterion cites.
- The three scripts the criteria cite (`test-plugin-install-step.sh`, `check-plugin-discoverable.sh`, `check-plugin-version.sh`) are specified
  and not written; their behavior is shown only by the surrogate observations R-01, R-13 and L-26.
- `pwsh` was refused by the worktree guard, as was a command prefixed `HOME=`; `install.ps1` and `install.bat` behavior is not executed (G-3).
- GitHub: `claude plugin marketplace add modu-ai/moai-adk` and the Codex form were not run; `main` returns 404 for `.claude-plugin` and `plugins/`.
- Unobserved: shadowing, collision and double firing between the scaffold copy and the plugin; the `moai:<name>` invocation form for this plugin;
  whether the runtime registers a nested `commands/<dir>/` file for invocation; `--scope project|local` behavior; the scaffold's own listing cost;
  `codex plugin marketplace upgrade`; any tool version other than claude 2.1.287 and codex 0.160.0; Windows.
- The status of `grep` and `find` commands that select nothing is not echoed by the tool; the ledger records the documented status and says so.
- The `moai` MCP server used for `spec_audit` is an older build than the tree; the lint ran on the tree-built binary.

### Gaps added by the iteration-3 delta

- The package, the three scripts and every Go test the criteria name still do not exist. The new cells L-38 and L-39 are stand-ins, written
  to the definitions in AC-004 (b) and AC-025 and run in the scratchpad; L-37 is the criterion's own check form on trees that exist (the
  template source tree and two stand-in payloads); L-40 is a control. The first run-phase RED record must show the real tests and the real
  harness failing at an assertion, not at `[no tests to run]` or `No such file or directory`.
- G-8 (new, plan §4): the real binary's `moai init` is not run by any harness case at the OD-14 default, because it writes
  `$HOME/.claude/settings.json` and `HOME` is the only seam (R-23). Not observed: init's own process start of the step.
- G-9 (new, plan §4): the protected-set hash cannot see `tmp/arg0` or `.tmp/git-*` under the real Codex home (other Codex processes write
  there, R-29), nor entries deeper than its declared depth inside an existing marketplace clone or plugin cache.
- The real `moai` binary was not run under the stand-in harness; the stand-in models the resolver order and one out-of-scratch write only. Whether
  the real binary writes anything outside the scratch homes with the scrub applied is for the first run of the real harness (the hash is
  the instrument).
- The two worktree-guard refusals met during this delta, both on compound commands (a heredoc-and-`cd` chain, and `find` with a
  `$CLAUDE_CONFIG_DIR` operand), were answered by splitting into plain commands and by writing files with the file tool; no measurement
  was substituted. `find` of the real profile by variable path was refused, so the real-root listing was taken by the hash script instead.
- The real-root hash was read-only and its two lists were dumped to the scratchpad; the only real-profile reads of this delta are those.
  No command of this delta wrote to a real profile or home (the scratch `CLAUDE_CONFIG_DIR`, `CODEX_HOME` and `MOAI_HOME` were set on the
  one product-binary run, the lint).

### Plan-audit iteration 2 and lane wait (recorded by the lane, not the author)

- Iteration 1 (audited_sha 3766cef05): FAIL 0.72, `.moai/reports/t1435/plan-audit-iter1.md`.
- Iteration 2 (audited_sha d6987e59c): FAIL 0.84, below the Tier L threshold 0.85, `.moai/reports/t1435/plan-audit-iter2.md`. D1..D24: 23 fixed, D3 partial. New blocking-class defects N1..N5 (the rest N6..N13 are minor or major-optional).
- Tier note: iteration 1 ran under the Tier M ceiling (`plan_audit_tier_ceilings.M: 2`, so iteration 2 was the last Tier M spawn). The re-tier to Tier L came from the auditor's D7 ruling (file count) and carries ceiling 3 (`plan_audit_tier_ceilings.L: 3`).
- WAIT, resolved: the lane waited on a ceiling decision (whom = leader; options offered: (a) iteration 3 under Tier L after fixing D3 + N1..N5, (b) PASS-with-debt with N1..N5 carried into the run phase as explicit debt, (c) scope cut or split at the M2/M3 boundary). **The leader approved (a): iteration 3 runs under the Tier L ceiling of 3**, limited to the delta of the iteration-2 verdict (the disposition table above); no requirement or criterion was renumbered and no feature was added. This is the last plan-audit spawn the ceiling admits: a FAIL at iteration 3 goes to PASS-with-debt, scope reduction or an explicit operator override (`spec-workflow.md` § SPEC Complexity Tier), not to a fourth round.
- The leader's safety order for this delta (binding): N2, N3 and N4 are one family with the t1434 leak, so every acceptance command and every harness script is isolated by scratch config homes, and a break of the isolation must turn a criterion red. Carried out as REQ-025 and AC-025 (leak canary, poisoned pin, live-enumerated scrub, directory-entry hash with `LEAK=0` required) and, for the doctor, AC-021 (c).

### Plan-audit iteration 3 and lane wait (recorded by the lane, not the author)

- Iteration 3 (audited_sha a0c8ad7cb, Tier L ceiling 3, approved by the leader): FAIL, score 0.86 (meets the 0.85 threshold; the verdict rests on three open blocking-class defects), `.moai/reports/t1435/plan-audit-iter3.md`. Scores 0.72 → 0.84 → 0.86.
- Iteration-2 open items: 12 fixed (D3 residue, N1, N3..N11, N13), N2 partial (ND-1), N12 skipped by the author (optional).
- Open blocking-class defects: ND-1 (AC-021 (c) subtest `registry-wide-starts-nothing` fails on correct work — existing `checkClaudeCode` runs `claude --version` when `CLAUDE_CODE_VERSION` is unset, reproduced with a recording shim), ND-2 (only one of four installer harness cases names `--install-dir <scratch>`; without it install.sh installs into `$GOBIN`/`$GOPATH/bin`/real `$HOME/.local/bin`, which the protected-set hash does not cover), ND-3 (REQ-025 requires stubs for `claude`/`codex` in every harness script; AC-006 (b) requires `check-plugin-discoverable.sh` to run the real `claude`). ND-4..ND-11 are minor and none can produce a false green.
- The lane has NOT started an iteration 4 (ceiling reached). WAIT: reason = ceiling reached with open blocking-class defects; whom = leader; recheck = leader reply or a re-read of this section. Options sent to the leader: (1) PASS-with-debt, carrying ND-1..ND-3 into the run phase as explicit debt with the auditor's conditions (ND-2 only if the run-phase delegation requires every `installer-*` case to pass `--install-dir <scratch>` and a case asserts the installed path lies under the scratch root), (2) one more delta round by explicit leader extension of the ceiling (the three fixes are one-clause text changes), (3) split at the M2/M3 boundary (removes ND-3 only; ND-1 and ND-2 sit in the M3/M4 half).
- WAIT, resolved (2026-10-03): the leader chose option (2), one more delta round, by an explicit ceiling extension (next section). The
  lane then applied the final delta below; it did not start a fourth audit.

### 운영자 일괄 수용 10-03

On 2026-10-03 the operator accepted, in one batch, the "default if unanswered" value of every open decision OD-1 to OD-14. **The source
is the leader's relay, by a cross-session message; it is not an answer typed in this session**, and this lane received the leader's
statement, not the operator's own words. `decision-index.md` records it in its own section of the same heading, and each of Q1 to Q14
carries `accepted default (operator batch 10-03, via leader relay)` followed by the accepted value, copied from that row's default in the
`spec.md` §5 table (checked: `grep -c '^- Operator verdict: accepted default (operator batch 10-03, via leader relay)' decision-index.md`
printed `14`, and `grep -c '^- Operator verdict:$' decision-index.md` printed `0`). No default changed, so no clause of the marker table
moved; no recommendation was added and no option text was edited. The relay is not a committed artifact and is not used as an authority
anchor: no row's label changed.

### Ceiling extension for ND-1 to ND-3 (recorded by the lane)

The leader granted one explicit ceiling extension, once, on 2026-10-03, for ND-1, ND-2 and ND-3 only, after the Tier L plan-audit
ceiling of 3 was reached at iteration 3. Scope held: the three defects plus the bookkeeping above; no requirement, criterion or Open
Decision added or renumbered (25 / 25 / 14); ND-4 to ND-11 and N12 not touched; no unaffected section rewritten; no feature added. The
follow-up audit looks at the three defects and a regression check only.

### Audit-iteration-3 disposition

Source: `.moai/reports/t1435/plan-audit-iter3.md` (FAIL 0.86 at audited_sha a0c8ad7cb; the score meets the 0.85 threshold, the FAIL rests on
ND-1 to ND-3). Evidence ids: P-nn in `spec.md` §1.3, R-nn in `research.md`, L-nn in `acceptance.md`; tree `cc46749d9` (it differs from the
audited tree in this file only); every real-`claude` and real-`go` command ran under empty scratch homes or read-only, and the real roots
were hashed before and after.

| Id | Disposition | Fixed where | Evidence |
|----|-------------|-------------|----------|
| ND-1 `registry-wide-starts-nothing` fails on correct work | **fixed**, option (ii): the subtest pins `CLAUDE_CODE_VERSION` (as `doctor_golden_test.go:177` does) and keeps "record empty" over both shims. Option (i), scoping the record to `codex` and `plugin`-first calls, was not taken because it would have to allow `claude --version`, which would also admit a new check's own `claude --version` | acceptance.md AC-021 (c) subtest text, RED-now line, ledger L-42; spec.md REQ-017 (scope clause) and P-48; design §4; plan M4 | R-30. Unmodified `TestRunDiagnosticChecks_All` under recording `claude` and `codex` shims, empty scratch homes: unpinned, `--- PASS … (9.69s)` and the record is `claude --version` (the auditor's line reproduced); pinned: `--- PASS … (6.61s)` and no record file. Stand-in of the restated subtest (Go test overlaid with `go test -overlay`, no repository write): PASS on the unmodified registry with 41 checks swept and an empty record; FAIL on a mutant that runs `codex plugin list --json` (record `codex plugin list --json`) and on one that runs `claude plugin list --json`. The `os.UserHomeDir()` mutant of the same bullet is unchanged and still attributed to `canary-home-not-touched` and `child-env-carries-resolved-home` (ND-8 untouched) |
| ND-2 installer cases not required to pass `--install-dir`, no check would see a write | **fixed**, prevention plus a per-case assertion, not the hash route. Every `installer-*` case passes `--install-dir <scratch>/inst-<case>/bin` and fails unless `go` resolves to the stub, the resolved installed path lies under the case's own directory, and neither decoy default root holds a `moai`; a stub `go` makes `install.sh`'s default roots decoys in the scratch; new negative control `--negative-control-install-dir`. The protected set is **not** extended to `$GOBIN`, `$GOPATH/bin`, `$HOME/.local/bin`, and design §6, plan M3 and AC-025 (d) say why | spec.md REQ-025 (clause and negative control), P-47; acceptance.md Conventions, AC-018 (a) with RED-now and mutant probe, AC-025 Given/Then, (a), (d), new (f), RED-now and mutant probe, ledger L-41; design §6 (new paragraph, step 4, case list, negative controls); plan M3 harness bullet, §4, §6 "Required for the harness"; this file's dispatch instructions | R-31. `install.sh:244-264` read: `--install-dir`, else `go env GOBIN`, else `$GOPATH/bin`, else `$HOME/.local/bin`; on this machine the bare target is `/Users/goos/go/bin`, which holds a real `moai`. The real `install.sh` driven offline by four stand-in cases: with the flag `PASS` x4, `LEAK=0`; without it `FAIL` x4 (`installed-path-not-under-install-dir`, `a-default-install-root-holds-moai`), real roots unchanged (`/Users/goos/go/bin entries=185 … moai=08fb8046e077`, `/Users/goos/.local/bin entries=44 … moai=none`). Control: an entry listing is equal across an overwrite of an existing `moai` (listing `7da953f4599d` both times, content `01d09d19c213` then `8a7bfaefd046`), which is why a listing hash cannot guard a real `$GOBIN`. `install.ps1` and `install.bat` both take `--install-dir` (P-47) and run in no harness case: static only (G-3) |
| ND-3 REQ-025 stubs vs AC-006 (b) real `claude` | **fixed**, carve-out, not a recorded stub: REQ-025 scopes the stub rule to `scripts/test-plugin-install-step.sh` and names `scripts/check-plugin-discoverable.sh` as the one script that runs the real `claude`, for exactly three verbs, behind the same scrub, from a scratch directory, with a scratch home it refuses unless empty, starting no `codex`. A recorded-output stub was rejected: P-30 measured the nested layout invisible only in the real runtime, so a replay would test the parser | spec.md REQ-025, P-49; acceptance.md Conventions, AC-006 (b) rewritten and new (d), (e), RED-now, green path and mutant probe, ledger L-43; design §6 (intro paragraph and the discoverable paragraph); plan M2 bullet, §4, §6 | R-32. Real `marketplace add <fixture> --json`, `install moai@moai-adk --json`, `details` under an empty scratch home: `"outcome":"ok"` twice and `Skills (41)`; the real roots read `PROTECTED-SET a245f41ac9cfed9029f2c1d27b75acc08d26492774227775f2407fbef777cf5f entries=193` before and after, dumps identical. Stand-in script with a poisoned caller environment (recorder pin, `CLAUDE_CODE_PLUGIN_CACHE_DIR` naming a canary): scrub on, canary `entries=1` before and after, recorder never ran, `ok: 41 names listed, 0 missing`; scrub off, canary `entries=1` then `entries=391` and no `plugins/` in the scratch home while the inventory line still read `ok`; refusal control exit 2 |
| ND-4 to ND-11 | untouched by design | none | the auditor's own statement (`plan-audit-iter3.md`, "Disposition if the operator elects PASS-with-debt"): "The minor items ND-4 to ND-11 are optional and safe to carry: each is a wording, count or cross-reference error, or a non-default option's consequence, and none can produce a false green." Known open at this revision, unchanged: the "six test files" count (ND-4), the harness poison set stated three ways (ND-5), the bracket starting at M3 (ND-6), OD-14 (b) and (c) consequences (ND-7), the AC-021 (c) mutant attribution (ND-8), the three safety claims (ND-9), the stale Tier paragraph above (ND-10), the exact-mode and scratch-cwd fragilities (ND-11) |
| N12 | untouched (skipped since iteration 2, optional) | none | both counts are at the Tier L ceiling 25 / 25; the iteration-3 audit classed it "NOT FIXED (skipped, disclosed)" and "Not counted against the verdict" |

Totals: ND-1, ND-2, ND-3 fixed; ND-4 to ND-11 and N12 untouched by design. Counts after the delta: 25 requirements, 25 acceptance criteria,
14 Open Decisions. New premise rows P-47 to P-49; new research rows R-30 to R-32; new ledger cells L-41 (stand-in), L-42 (scoping
control, not a RED) and L-43 (stand-in); no new gap row in plan §4, and the one observation below.

### Final delta: verification, gaps, and one observation beyond the three defects

`<s>` is the session scratchpad `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/2f10c8c5-67ea-41c2-9b61-6242acc465c3/scratchpad`
and `<n4>` is `<s>/n4`.

- Lint: `CLAUDE_CONFIG_DIR=<n4>/h/c-lint CODEX_HOME=<n4>/h/x-lint MOAI_HOME=<n4>/h/m-lint <n4>/moai spec lint SPEC-PLUGIN-MARKETPLACE-001 --strict`
  printed `✓ No findings — all SPEC documents are valid`. Judging build: `go build -o <n4>/moai ./cmd/moai` at HEAD
  `cc46749d99512f43751f2e4de7f72cee431b863b` (the same value before and after the build), built without ldflags, so it self-reports
  `moai-adk v3.1.3` and `v3.1.3   none   built unknown`: its commit is attributed by build procedure, not self-attested. The installed
  `moai` was not used. The same command was run again after the last edit of the delta and printed the same line.
- Counts: `grep -c '^- REQ-0' spec.md` printed `25`; `grep -c '^### AC-0' acceptance.md` printed `25`; `grep -o 'OD-[0-9][0-9]*' spec.md | sort -u |
  wc -l` printed `14`, the section-5 question table holds 14 rows, and `grep -c '^## Q[0-9]*:' decision-index.md` printed `14`.
- Real roots: `sh <s>/i3/h/protected-set-hash.sh --dump <file> --roots-file <n4>/real-roots.txt` (roots saved from the caller's environment before any
  command) printed `PROTECTED-SET a245f41ac9cfed9029f2c1d27b75acc08d26492774227775f2407fbef777cf5f entries=193` before the first real-tool
  command, after the ND-3 verbs, after the stand-in script runs, and after the last experiment; `cmp` of the first and last dumps:
  identical.
- Tree: `git rev-parse --short HEAD` printed `cc46749d9` and the branch is `WT-marketplace-core-plugin` at the start and at the end; `git status
  --short` lists only files of this SPEC directory; nothing was staged or committed; no agent-memory file was written. Scratch files (shims,
  stand-in scripts, overlays, canaries, fixtures) are under `<n4>` and not in the repository. No command of this delta wrote to a real
  profile or home; no marketplace add, install or uninstall ran against a real profile; no acceptance command used the network (the
  marketplace source was a local fixture path).
- Gaps: the real harness, the real `check-plugin-discoverable.sh` and every new Go test still do not exist, so the three fixes are shown
  through stand-ins (the real `install.sh`, the real `claude` and the real registry-wide test were driven; the harness cases, the script and
  the new subtests were not), and the first run-phase RED record must show the real ones failing at an assertion. The stand-in installer
  cases model the install-directory property only: the base `install.sh` makes no verb call, so the verb assertions of the four cases
  are not exercised. `install.ps1` and `install.bat` were read, not run. The registry-wide runs read real files through pre-existing checks
  (status quo, read-only, as in iteration 3). In this shell `ls -A` of an empty directory prints `total 0`, `.` and `..` (three lines), which is
  how emptiness was shown. The worktree guard refused no command of this delta.
- Observation beyond the three defects, reported and not fixed (outside the delta's scope): P-49 shows `CLAUDE_CODE_PLUGIN_CACHE_DIR`
  moves the plugin tree out of a scratch `CLAUDE_CONFIG_DIR`. Acceptance commands that run a registry-writing tool inline set only
  `CLAUDE_CONFIG_DIR` or `CODEX_HOME` (the `claude` ones are read-only `validate`; the Codex write verbs of AC-002 and AC-021 (b) are
  inline, and which `CODEX_*` variables move Codex's writes was not measured). This session did not carry the variable. It is the class
  of ND-9 (i), the protected roots following the caller's environment; the discoverable script's scrub closes it for that script only.

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
