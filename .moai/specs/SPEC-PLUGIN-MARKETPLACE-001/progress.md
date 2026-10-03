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

### Plan-audit delta (iteration 4, one-time leader extension) and Kickoff wait (recorded by the lane, not the author)

- Delta audit (audited_sha 58ee0bdb2, plan-artifact hash 9052d51b01fc9b409adf25c720bc897f81254cf1b7dab30e676e7e754ac5e20b): PASS-WITH-DEBT, score 0.87, `.moai/reports/t1435/plan-audit-iter4-delta.md`. ND-1, ND-2, ND-3 FIXED; new blocking-class or regression defects: none; counts 25 REQ / 25 AC / 14 OD; lint clean. Scores 0.72 → 0.84 → 0.86 → 0.87.
- Residuals NR-1..NR-5 are run-phase debt, carried as binding run instructions BI-1..BI-3 from the verdict file (they live in the verdict, not in the SPEC, so the plan-artifact hash stays 9052d51b…): BI-1 run-start env pre-flight before the first write, re-measured at dispatch and not carried from the audit; BI-2 the omit-flag negative control (AC-025 (f)) and AC-018 (a) assertion (3) implement NR-1's preconditions, i.e. the decoy directories are asserted to exist BEFORE the control runs and the control aborts without running the installer if one is missing; BI-3 the discoverable script implements NR-3's pid-named poison, recording wrapper and local-path grep, with (e) labelled static.
- The auditor states that if the leader's safety order is read as covering any control whose safety rests on an unasserted precondition, NR-1 is the one clause that would turn the verdict into FAIL; it is a single-clause fix inside AC-025 (f) and AC-018 (a), and BI-2 is the run-phase form of that fix.
- Heading repair found by the lane after the delta audit: the `## §E.2 Run-phase Evidence` heading was missing from `58ee0bdb2` (present at `cc46749d9`, `a0c8ad7cb` and `3766cef05`), leaving its `_<pending run-phase>_` body orphaned under the previous section. The `§E.2` heading is parser-load-bearing (`internal/spec/era.go`, spec-frontmatter-schema § progress.md Section Map), so it is restored in the commit that carries this note. progress.md is not in the plan-artifact hash subject set, so the audited hash is unchanged; the delta audit's Gaps did not name it.
- Kickoff status: NOT entered. auto-semantics §9.1 names verdict PASS as the autonomous entry condition; this verdict is PASS-WITH-DEBT. WAIT: reason = whether PASS-WITH-DEBT with BI-1..BI-3 carried satisfies the autonomous Kickoff predicate; whom = leader; recheck = leader reply or a re-read of this section. Operator batch acceptance of OD-1..OD-14 is recorded above (via leader relay).

## §E.2 Run-phase Evidence

### R0 (BI-2: decoy preconditions of the installer cases and the omit-flag negative control)

Run-phase worker: `Agent(general-purpose)` carrying the manager-develop role text, `cycle_type=tdd`, in the card worktree.

#### Pre-flight (recorded before any edit, 2026-10-03)

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435
$ git rev-parse HEAD
45edfc6fd2fda3914e085fdd1b0c0e9c5ddf4e6b
$ git branch --show-current
WT-marketplace-core-plugin
$ git status --short
(empty)
$ env | cut -d= -f1 | grep -E '^(CLAUDE_CODE_PLUGIN_|BASH_ENV$|ENV$|BASH_FUNC_|GOBIN$|GOPATH$|GOFLAGS$|GOENV$|CODEX_SQLITE_HOME$|XDG_)'
(empty; grep exit 1)                                         # BI-1, start of the first measurement
$ grep -c moai-factory-foreman internal/template/catalog.yaml
0                                                            # base does not contain t1399
```

#### Commits (tdd order: tooling, observed RED, GREEN)

| Step | SHA | Content |
|------|-----|---------|
| infrastructure | `65c175af8dfbd8c598911a34184de16c08e912d7` | `scripts/protected-set-hash.sh` (AC-025 (d) bracket); spec.md `status: draft` to `in-progress` (first run-phase commit; `updated:` already read 2026-10-03, so unchanged); this section's pre-flight |
| RED | `cdd2e5ac30a13854fda5e287a10d015f2f647380` | `scripts/test-plugin-install-step.sh` without the decoy precondition |
| GREEN | `81916fb751cb2f89dc0d8b94c99187595b314a36` | the precondition (`preflight_decoys`, `abort`) and the status-2 demand of the decoy-missing control |
| REFACTOR | none | no separate REFACTOR step was run; the GREEN commit already folds in the one tightening (status-2 demand) found by the mutants |

#### Claim

1. BI-2 is implemented in the harness: before every installer run the harness asserts that both decoy install roots exist (`$S/decoy-gobin` and `$S/decoy-gopath/bin`, the directory `install.sh:254-257` really installs into, per NR-1's note on assertion (3)), lie under the resolved scratch root, that `go` resolves to the stub and that the stub answers `go env GOBIN` / `go env GOPATH` with exactly those directories; otherwise it aborts with status 2 and the installer is not run.
2. The four `installer-*` cases pass `--install-dir <scratch>/inst-<case>/bin` and assert (1) `go` is the stub, (2) the resolved installed path lies under the case's own directory, (3) neither decoy holds a `moai`; the omit-flag control turns all four red by the decoy assertion with the stub `go` shown consulted; normal mode: three PASS and one PENDING, `LEAK=0`.
3. The decoy-missing control proves non-execution by a recording wrapper (`installer-invocations.log`) and by abort's own exit status, and went red against two guard mutants.
4. `scripts/protected-set-hash.sh` sees empty-directory, new-file and content changes in a canary home.

#### Evidence (verbatim, this run, this tree)

Protected-set bracket (AC-025 (d)), real roots, the caller's own `CLAUDE_CONFIG_DIR`/`CODEX_HOME`/`HOME`:

```
$ sh scripts/protected-set-hash.sh        # before, after commit 65c175af8, before the first harness run
PROTECTED-SET 753334575dbf1fe141254a71de36ff7eda25227130c6454bd24e333b3491f2aa entries=190
$ sh scripts/protected-set-hash.sh        # after the last harness run, tree HEAD 81916fb75
PROTECTED-SET 753334575dbf1fe141254a71de36ff7eda25227130c6454bd24e333b3491f2aa entries=190
$ diff real-before.txt real-after.txt     # the two --dump entry lists
(empty; exit 0)                                              # LEAK=0
```

BI-1 (`env | cut -d= -f1 | grep -E '^(CLAUDE_CODE_PLUGIN_|BASH_ENV$|ENV$|BASH_FUNC_|GOBIN$|GOPATH$|GOFLAGS$|GOENV$|CODEX_SQLITE_HOME$|XDG_)'`) was run at the start of every measurement batch (hash self-test, before-line, the first harness run, the GREEN runs, the final runs): each printed nothing, grep exit 1.

Hash script self-test on a canary home (`CLAUDE_CONFIG_DIR` and `CODEX_HOME` under the scratchpad, both shown empty by `ls -A` first):

```
empty homes                                   PROTECTED-SET 978abca5584009c769332f27cb1dda3c398f889b1c934888d6a2ed30998d2508 entries=35
+ plugins/data/x (empty directories only)     PROTECTED-SET b8da8d8c76d3fcc7e8736500dc4d1f668edae4be42954bc6ccaab640f700b4de entries=37
  control: find <claude home> -type f         (no output)    # a files-only listing sees nothing
+ settings.json {"a": 1}                      PROTECTED-SET c36a2ee008c7664d9199087a8972125ba984ea1932c191b031fafcc1330ce28e entries=37
  settings.json content {"a": 2}              PROTECTED-SET a42722ccd3ef4468011b27eff76bddde2f4662815a07667c89ed73713327e32b entries=37
--roots-file on the saved roots, no env       PROTECTED-SET b8da8d8c... entries=37   # equals the live reading of the same state
```

RED, observed at an assertion (`cdd2e5ac3`, harness without the precondition), run before the commit:

```
$ sh scripts/test-plugin-install-step.sh --negative-control-decoy-missing bin/moai
scrub: enumerated and unset 34 names
FAIL decoy-missing-aborts-before-installer (removed decoy-gobin: rc=0, aborted=no, installer runs=1)
FAIL decoy-missing-aborts-before-installer (removed decoy-gopath/bin: rc=0, aborted=no, installer runs=1)
RESULT negative-control-decoy-missing: the installer RAN with a decoy missing
LEAK=0 (real roots unchanged: PROTECTED-SET 753334575dbf1fe141254a71de36ff7eda25227130c6454bd24e333b3491f2aa entries=190)
exit=1
```

The RED run was safe: one decoy removed per iteration leaves the other as `install.sh`'s root (`install.sh:254-261`), so `$HOME/.local/bin` is reached only when both are absent, which no run arranged. Normal mode and the omit-flag control passed at that commit (their output equals the GREEN output below); only the new control was red.

GREEN (`81916fb75`), final runs at the same tree:

```
$ sh scripts/test-plugin-install-step.sh --negative-control-decoy-missing bin/moai          # exit 0
scrub: enumerated and unset 34 names
PASS decoy-missing-aborts-before-installer (removed decoy-gobin: rc=2, installer runs=0)
PASS decoy-missing-aborts-before-installer (removed decoy-gopath/bin: rc=2, installer runs=0)
RESULT negative-control-decoy-missing: the installer never ran with a decoy missing
LEAK=0 (real roots unchanged: PROTECTED-SET 753334575dbf1fe141254a71de36ff7eda25227130c6454bd24e333b3491f2aa entries=190)

$ sh scripts/test-plugin-install-step.sh bin/moai                                            # normal, exit 0
scrub: enumerated and unset 34 names
PENDING installer-calls-verb-by-installed-path: install-directory assertions pass; the verb-call assertion lands with install.sh in M3
PASS installer-optout
PASS installer-set-e-guard
PASS installer-old-binary-unknown-verb
RESULT pass=3 fail=0 pending=1
LEAK=0 (real roots unchanged: PROTECTED-SET 753334575dbf1fe141254a71de36ff7eda25227130c6454bd24e333b3491f2aa entries=190)

$ sh scripts/test-plugin-install-step.sh --negative-control-install-dir bin/moai              # omit-flag, exit 0
scrub: enumerated and unset 34 names
RED installer-calls-verb-by-installed-path: installed-path-not-under-case-dir(no moai at <scratch>/inst-installer-calls-verb-by-installed-path/bin); decoy-holds-moai(<scratch>/decoy-gobin/moai)
RED installer-optout: installed-path-not-under-case-dir(no moai at <scratch>/inst-installer-optout/bin); decoy-holds-moai(<scratch>/decoy-gobin/moai)
RED installer-set-e-guard: installed-path-not-under-case-dir(no moai at <scratch>/inst-installer-set-e-guard/bin); decoy-holds-moai(<scratch>/decoy-gobin/moai)
RED installer-old-binary-unknown-verb: installed-path-not-under-case-dir(no moai at <scratch>/inst-installer-old-binary-unknown-verb/bin); decoy-holds-moai(<scratch>/decoy-gobin/moai)
RESULT negative-control-install-dir: red set and green set are exactly the expected ones
LEAK=0 (real roots unchanged: PROTECTED-SET 753334575dbf1fe141254a71de36ff7eda25227130c6454bd24e333b3491f2aa entries=190)
```

(`<scratch>` abbreviates the `mktemp -d` path, `/private/var/folders/.../T/tmp.XXXXXXXXXX`; the full path is in the raw output of the run.)

Guard mutants (copies of the harness under the scratchpad, each with its own copy of `install.sh` and the hash script; the decoy-missing control run against each):

```
mutant 1: preflight loop checks only the GOBIN decoy
PASS decoy-missing-aborts-before-installer (removed decoy-gobin: rc=2, installer runs=0)
FAIL decoy-missing-aborts-before-installer (removed decoy-gopath/bin: rc=0, aborted=no, installer runs=1)   exit 1
mutant 2: abort() prints but does not exit (every "exit 2; }" replaced by "true; }")
FAIL decoy-missing-aborts-before-installer (removed decoy-gobin: rc=1, aborted=yes, installer runs=0)
FAIL decoy-missing-aborts-before-installer (removed decoy-gopath/bin: rc=1, aborted=yes, installer runs=0)   exit 1
```

Mutant 2 first survived (rc=1 from a `set -e` accident, installer runs=0), which is why the control was tightened to demand abort's own status 2 before GREEN was committed; the table above is the run after that change.

Static (AC-025 (e), the two scripts that exist): `grep -nE '(^|[^A-Za-z_])HOME=|(^|[^A-Za-z_])export HOME' scripts/test-plugin-install-step.sh scripts/protected-set-hash.sh` printed nothing, exit 1; control `grep -c 'CODEX_HOME' scripts/test-plugin-install-step.sh` printed `2`. `sh -n` passed on both scripts and `bash -n install.sh` passed.

#### Baseline-attribution

Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435`, branch `WT-marketplace-core-plugin`, base HEAD `45edfc6fd`, measured at HEAD `65c175af8` (hash self-test, before-line), `cdd2e5ac3` (RED run), `81916fb75` (GREEN and final runs, after-line), all in this run. The judging tools are the shell scripts of this tree, invoked by path (`sh scripts/...`); no installed build of anything judged them. `bin/moai` (gitignored) was built by `go build -o bin/moai ./cmd/moai` from this tree at HEAD `65c175af8`, exit 0, and is only the archive payload of the `real` kind: no case executes it (it is not on the harness PATH and `install.sh` does not call the verb yet). The installer under test is this tree's unchanged `install.sh`.

#### Gaps

- G-R0-1 (deviation from the dispatch's "all four GREEN"): the normal run is three `PASS` and one `PENDING`. `installer-calls-verb-by-installed-path` asserts that `install.sh` calls the verb by its installed path, and `install.sh` carries no such call until M3 (out of R0 by instruction). Its install-directory assertions run and pass; its verb assertion is not written, and the line says `PENDING` rather than `PASS` so that nothing reads as a verb check that did not happen. `installer-optout` (zero recorded `plugin` calls) and `installer-set-e-guard` / `installer-old-binary-unknown-verb` (installer exits 0 and prints `Installation complete!`) pass today for the reason that `install.sh` makes no plugin call; they cannot go red on a wrong call until M3 adds one, so their mutant probes (AC-018) are unobserved.
- G-R0-2: the isolation-* and verb-* cases, `--negative-control`, `--negative-control-cwd` and `--typed-list-mutant` do not exist yet (M3 by instruction). The environment scrub is implemented and prints its count, and a post-scrub abort check exists, but `isolation-env-scrubbed` and the typed-list negative control are the verdict of AC-025 and are unobserved.
- G-R0-3: the omit-flag control's `green` side is vacuous in R0: the five isolation cases it expects to stay green do not exist, so only the red set (all four installer cases, each by the decoy assertion with the stub `go` shown consulted) is judged.
- G-R0-4: `scripts/protected-set-hash.sh` was written before its first harness consumer ran and was not itself driven RED-first; its observed failures are the canary self-test above (changes move the hash, a files-only listing is blind to the directory change). The infrastructure commit therefore precedes the observed RED commit by design, not by a RED of its own.
- G-R0-5: refusal (verification-claim-integrity section 3.1): a command of the form `sh scripts/test-plugin-install-step.sh --negative-control-install-dir bin/moai | cut -c1-120; echo "exit=${PIPESTATUS:-n/a}"` was refused by the worktree guard ("construct too complex to verify"); it was re-run as the plain single command and its output is the one quoted above. Nothing was substituted by reading.
- G-R0-6: `$HOME/.local/bin`, `$GOBIN` and `$GOPATH/bin` are not in the protected set by design (`design.md` section 6); that no install reached them rests on the per-case assertions and on the prevention above, and on the real-root hash being equal, not on a hash of those directories. The failing path itself (both decoys absent, `install.sh` writing `$HOME/.local/bin`) was not run, because running it would write a real directory and `HOME` may not be assigned.
- G-R0-7: the real roots were hashed to depth 1 under `~/.moai` and depth 3 under the plugin roots; ambient churn deeper than that, `~/.claude.json`, and `tmp/arg0` under the Codex home are blind spots stated by the SPEC (G-9). `~/.claude/settings.json` was named in the roots but its mtime was not recorded (NR-4 asked for the mtime change to be named if it recurs: its content hash is part of the compared set and was equal).
- Not run, by instruction: `go test` of any package, `install.ps1`, `install.bat`.

#### Residual-risk

- NR-1 is closed for the harness as written; what remains is that the decoy assertions guard only the harness's own runs. A later edit that adds an installer call outside `run_case` (the one function that carries `preflight_decoys` and the recording wrapper) would bypass both; M3 must add cases through `run_case`.
- The decoy precondition proves the decoys exist when the installer starts; a decoy removed by something else during the run is not observed (the post-run decoy-holds-moai assertion still sees a landing in whichever decoy remains).
- `~/.moai` depth-1 entries can change on a live machine (NR-4): a non-zero `LEAK` will need its differing entries attributed (`diff` of the two dumps is printed) rather than assumed.
- The `PENDING` line in normal mode is a standing reminder, not a gate: nothing fails if M3 forgets to replace it; the AC-018 (a) expectation that the line reads `PASS installer-calls-verb-by-installed-path` is what holds M3 to it.

### M1 (emitter skeleton, manifests, MCP derivation, validation)

Run-phase worker: `Agent(general-purpose)` carrying the manager-develop role text, `cycle_type=tdd`, in the card worktree. Scope: M1 only (REQ-001, 002, 003 and the derivation unit of REQ-007; exit AC-001, AC-002, AC-003).

#### Pre-flight (recorded before any edit, 2026-10-03)

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435
$ git rev-parse --short HEAD
f22021fa3
$ git branch --show-current
WT-marketplace-core-plugin
$ git status --short
(empty)
$ env | cut -d= -f1 | grep -E '^(CLAUDE_CODE_PLUGIN_|BASH_ENV$|ENV$|BASH_FUNC_|GOBIN$|GOPATH$|GOFLAGS$|GOENV$|CODEX_SQLITE_HOME$|XDG_)'
(empty; grep exit 1)                                         # BI-1, start of the first measurement batch
$ go build ./...                                             # exit 0
$ GOOS=windows GOARCH=amd64 go build ./...                   # exit 0
$ grep -c moai-factory-foreman internal/template/catalog.yaml
0                                                            # t1399 not landed
$ sh scripts/protected-set-hash.sh                           # before-line, before the first claude/codex-touching command
PROTECTED-SET 753334575dbf1fe141254a71de36ff7eda25227130c6454bd24e333b3491f2aa entries=190
```

B2 cross-SPEC scan (`grep -rn -i 'plugin' internal/template/*.go | grep -i 'retir\|supersed'`): no match. The wider `grep -r "Retired\|superseded" internal/template` matches ten unrelated files (model-policy, tool catalog, retired wrappers and similar); none concerns a plugin or marketplace emitter, so no conflict with a new `internal/template/pluginemit` package.

#### Commits (tdd order: observed RED, GREEN)

| Step | SHA | Content |
|------|-----|---------|
| RED | `7b863c2e5` | `internal/template/pluginemit` as a compiling stub (empty publication, zero MCP entry), the five test functions, hand-authored golden shape pins under `testdata/golden/`, and the pre-flight above |
| GREEN | `6298aa9fe` | `pluginemit.go`, `manifest.go`, `mcp.go`, `Makefile` (`plugin-emit`, `plugin-emit-check`, `.PHONY`), and the four manifests produced by `make plugin-emit` |
| REFACTOR | none | nothing to simplify: the goldens authored before GREEN matched the first emission byte for byte, and the three generator files total about 300 lines against a minimum of about 150 (under the 3x trigger) |

The RED commit precedes the GREEN commit in the commit graph, so the order is witnessed by git and not only asserted here (verification-claim-integrity section 2.3).

#### Claim

1. The generator emits the four manifests (REQ-001, REQ-002, REQ-003): marketplace `moai-adk`, one entry `moai` with `source` `./plugins/moai` (OD-4 default), all four version-carrying fields equal to `pkg/version.Version` minus its leading `v`, the Codex marketplace with no `version` key at either level, the Codex plugin manifest with `skills` and the `moai` MCP entry.
2. The MCP entry is derived from the template `.mcp.json`, not retyped (derivation unit of REQ-007, AC-007 (c)); a missing or malformed source is refused.
3. AC-001, AC-002 and AC-003 pass by their own commands (matrix below). `make plugin-emit-check` is green on the generated tree and red on a hand edit and on absent files, and never writes.
4. No scope beyond M1 was started: no payload (skills, commands, `.mcp.json`), no drift gate beyond bytes, no install step, no doctor check; `plugin-emit-check` is not yet a prerequisite of `build` (M2, plan section 3).

#### Evidence (verbatim, this run, this tree)

RED, observed at assertions (`7b863c2e5` content, run before the commit; none failed at compile or discovery):

```
$ go test ./internal/template/pluginemit/... -count=1 -v
=== RUN   TestManifestsGolden
    golden_test.go:58: .agents/plugins/marketplace.json: not emitted
    golden_test.go:58: .claude-plugin/marketplace.json: not emitted
    golden_test.go:58: plugins/moai/.claude-plugin/plugin.json: not emitted
    golden_test.go:58: plugins/moai/.codex-plugin/plugin.json: not emitted
    golden_test.go:76: emitted 0 files, want exactly the 4 manifests
--- FAIL: TestManifestsGolden (0.00s)
=== RUN   TestGoldenCommittedArtifactsMatchEmission
    golden_test.go:89: emitted set is empty — nothing was compared
--- FAIL: TestGoldenCommittedArtifactsMatchEmission (0.00s)
=== RUN   TestVersionStampedFromSSOT
    manifest_test.go:93: emitted set lacks .claude-plugin/marketplace.json (have 0 files)
--- FAIL: TestVersionStampedFromSSOT (0.00s)
=== RUN   TestCommittedVersionMatchesSSOT
    manifest_test.go:123: committed manifest missing: open ../../../.claude-plugin/marketplace.json: no such file or directory
--- FAIL: TestCommittedVersionMatchesSSOT (0.00s)
=== RUN   TestMCPEntryDerivedFromTemplate
=== RUN   TestMCPEntryDerivedFromTemplate/default_entry
    mcp_test.go:38: emitted set lacks plugins/moai/.codex-plugin/plugin.json (have 0 files)
=== RUN   TestMCPEntryDerivedFromTemplate/changed_command_and_args
    mcp_test.go:38: emitted set lacks plugins/moai/.codex-plugin/plugin.json (have 0 files)
=== RUN   TestMCPEntryDerivedFromTemplate/missing_moai_entry_is_refused
    mcp_test.go:63: Emit succeeded without a moai entry in the template .mcp.json
=== RUN   TestMCPEntryDerivedFromTemplate/malformed_template_is_refused
    mcp_test.go:69: Emit succeeded over a malformed template .mcp.json
--- FAIL: TestMCPEntryDerivedFromTemplate (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/template/pluginemit	0.243s
```

GREEN, tree HEAD `6298aa9fe`:

```
$ go test ./internal/template/pluginemit/... -count=1 -v
--- PASS: TestManifestsGolden (0.00s)
--- PASS: TestGoldenCommittedArtifactsMatchEmission (0.00s)
--- PASS: TestVersionStampedFromSSOT (0.00s)
--- PASS: TestCommittedVersionMatchesSSOT (0.00s)
--- PASS: TestMCPEntryDerivedFromTemplate (0.00s)    (four subtests PASS)
ok  	github.com/modu-ai/moai-adk/internal/template/pluginemit	0.089s
```

Acceptance matrix (every command run at HEAD `6298aa9fe`; scratch homes `<scratchpad>/m1/{claude,codex}-AC-00n`, six directories shown empty by `ls -A` first, one pair per criterion):

| AC | Result | Command | Verbatim output |
|----|--------|---------|-----------------|
| AC-001 (a) | PASS | `CLAUDE_CONFIG_DIR=<claude-home:AC-001> claude plugin validate .claude-plugin/marketplace.json --strict` | `Validating marketplace manifest: …/.claude-plugin/marketplace.json` / `✔ Validation passed` (tool result not flagged as an error; exit code not echoed) |
| AC-001 (b) | PASS | `jq -c '[.name,[.plugins[].name],.plugins[0].source]' .claude-plugin/marketplace.json` | `["moai-adk",["moai"],"./plugins/moai"]` |
| AC-001 (c) | PASS | `jq -e '.metadata.version == .plugins[0].version' .claude-plugin/marketplace.json` | `true` |
| AC-002 (a) | PASS | `CODEX_HOME=<codex-home:AC-002> codex plugin marketplace add .` | ``Added marketplace `moai-adk` from /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435.`` / `Installed marketplace root: …/t1435` |
| AC-002 (b) | PASS | `jq -c '[.name,[.plugins[].name],.plugins[0].source,.plugins[0].policy]' .agents/plugins/marketplace.json` | `["moai-adk",["moai"],{"source":"local","path":"./plugins/moai"},{"installation":"AVAILABLE","authentication":"ON_INSTALL"}]` |
| AC-002 (c) | PASS | `jq -e '[has("version"), (.plugins[0] \| has("version"))] == [false, false]' .agents/plugins/marketplace.json` | `true` |
| AC-003 (a) | PASS | `CLAUDE_CONFIG_DIR=<claude-home:AC-003> claude plugin validate plugins/moai --strict` | `Validating plugin manifest: …/plugins/moai/.claude-plugin/plugin.json` / `✔ Validation passed` |
| AC-003 (b) | PASS | `jq -c '[.name,.version]' plugins/moai/.claude-plugin/plugin.json plugins/moai/.codex-plugin/plugin.json` | `["moai","3.1.3"]` and `["moai","3.1.3"]` |
| AC-003 (c) | PASS | `jq -c '[.skills,.mcpServers.moai]' plugins/moai/.codex-plugin/plugin.json`; `jq -c .mcpServers.moai plugins/moai/.codex-plugin/plugin.json internal/template/templates/.mcp.json` | `["./skills/",{"command":"moai","args":["mcp-server"]}]`; then two identical lines `{"command":"moai","args":["mcp-server"]}` |
| AC-003 (d) | PASS | `go test ./internal/template/pluginemit -run '^TestVersionStampedFromSSOT$' -count=1 -v` | `--- PASS: TestVersionStampedFromSSOT (0.00s)` |
| AC-003 (e) | PASS | `go test ./internal/template/pluginemit -run '^TestCommittedVersionMatchesSSOT$' -count=1 -v` | `--- PASS: TestCommittedVersionMatchesSSOT (0.00s)` |
| AC-007 (c), derivation unit only | PASS | `go test ./internal/template/pluginemit -run '^TestMCPEntryDerivedFromTemplate$' -count=1 -v` | `--- PASS: TestMCPEntryDerivedFromTemplate (0.00s)` and four subtests PASS (AC-007 (a) and (b) are M2) |

Build and lint (this tree, `6298aa9fe`): `go build ./...` completed with no output (exit 0, per the tool result); `GOOS=windows GOARCH=amd64 go build ./...` completed with no output (exit 0); `golangci-lint run --timeout=2m ./internal/template/pluginemit/...` printed `0 issues.` with `golangci-lint has version v2.1.6` (the CI version, `.github/workflows/ci.yml:464`); `go vet ./internal/template/pluginemit/...` no output.

Mutants (each run, then restored; `git status --short` printed nothing afterwards):

```
hand edit: plugins/moai/.claude-plugin/plugin.json license "Apache-2.0" -> "MIT"
$ make plugin-emit-check
--- FAIL: TestGoldenCommittedArtifactsMatchEmission (0.01s)
    golden_test.go:116: plugins/moai/.claude-plugin/plugin.json: committed artifact differs from emission — run `make plugin-emit` or stop hand-editing
plugin-emit drift: committed marketplace and plugin manifests differ from the generator — run `make plugin-emit`
make: *** [plugin-emit-check] Error 1
$ grep -n license plugins/moai/.claude-plugin/plugin.json      # the check wrote nothing
10:  "license": "MIT"
$ make plugin-emit                                              # explicit verb restored the file; the tree was clean again

absent files (before the first make plugin-emit): make plugin-emit-check printed four "committed artifact missing" lines and exited 1 (Error 1).

generator mutant: version hard-coded as "v3.1.3" in Emit
  TestCommittedVersionMatchesSSOT  PASS   (the AC-003 (e) blind spot)
  TestVersionStampedFromSSOT       FAIL   "claude plugin version = 3.1.3, want "9.8.7-rc.1" (SSOT minus the leading v)" (four fields)
  TestManifestsGolden              FAIL   (version 3.1.3 against the pinned 1.2.3)

generator mutant: MCP entry retyped as a literal after DeriveMCPEntry
  TestMCPEntryDerivedFromTemplate/changed_command_and_args FAIL "command = moai, want "alt-launcher" (copied from the template)"
```

Protected-set bracket (AC-025 (d)), real roots, the caller's own environment:

```
$ sh scripts/protected-set-hash.sh        # before, first claude/codex-touching command
PROTECTED-SET 753334575dbf1fe141254a71de36ff7eda25227130c6454bd24e333b3491f2aa entries=190
$ sh scripts/protected-set-hash.sh        # after the last claude/codex command, tree HEAD 6298aa9fe
PROTECTED-SET 753334575dbf1fe141254a71de36ff7eda25227130c6454bd24e333b3491f2aa entries=190
```

The two lines are equal and equal the baseline named in the dispatch: LEAK=0. BI-1 (`env | cut -d= -f1 | grep -E '^(CLAUDE_CODE_PLUGIN_|BASH_ENV$|ENV$|BASH_FUNC_|GOBIN$|GOPATH$|GOFLAGS$|GOENV$|CODEX_SQLITE_HOME$|XDG_)'`) was run at the pre-flight, immediately before the `claude`/`codex` batch, and at the end of the run: each printed nothing, grep exit 1. The only `claude` and `codex` commands run were `claude plugin validate` (twice) and `codex plugin marketplace add .`, each under its own empty scratch home; no install, uninstall or registry-writing command ran against a real home.

#### Baseline-attribution

Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435`, branch `WT-marketplace-core-plugin`, base HEAD `f22021fa3`; RED measured on the `7b863c2e5` content, every GREEN and acceptance measurement at HEAD `6298aa9fe`, all in this run. The judging tools: `go` (go1.26.8), `golangci-lint` v2.1.6, `claude` 2.1.288, `codex-cli` 0.160.0 and `jq`, all installed binaries invoked through PATH; none of them is built from this tree and none is the project's own tooling, so section 2.2 of verification-claim-integrity does not ask for a build-versus-HEAD statement for them. The Go tests judge the tree's own `pluginemit` package, built by `go test` from this tree.

#### Gaps

- G-M1-1: `claude` read 2.1.288 in this run, the t1434 and plan observations are at 2.1.287 (P-01). The strict-validate and `plugin details` behaviours the SPEC cites were not re-measured at 2.1.288 beyond the three validate and add commands above; no difference was observed in those.
- G-M1-2: AC-001 and AC-003 name tool-killed mutants (a deleted `metadata.description` turning (a) red, an entry version differing from `plugin.json`, P-02 and P-03). They were not re-run here; the mutants run in this section are the generator and hand-edit mutants only. The RED-now cells L-01 to L-04 were not re-executed either: the RED evidence above is the Go-test RED of the stub.
- G-M1-3: the exit codes of the three `claude`/`codex` commands and of the `jq -e` forms were read from the tool result (an error result carries the exit code, a clean result carries none), not echoed with `; echo $?`, which the worktree guard refuses (P-25).
- G-M1-4: field effect in Codex stays UNOBSERVED (SPEC G-4, R14-codex): `codex plugin marketplace add .` shows the root is recognized, not which manifest fields Codex reads. The Codex manifest names `skills: "./skills/"`, a directory that does not exist until M2; AC-002 (a) and AC-003 (a) pass without it.
- G-M1-5: the `interface` block values and the two `category` values (`development` for Claude, `Coding` for Codex), the owner `modu-ai`, the display names and the `websiteURL` are generator constants this run chose, from the cowork vocabulary and the repository URL; the SPEC fixes the keys and the shapes, not these values. They are pinned by the golden, so a different choice is a one-line change plus `make plugin-emit`.
- G-M1-6: `plugin-emit-check` is not yet a prerequisite of `build:` (M2, plan section 3), so `make build` does not run it today. Drift detection covers bytes only; mode, missing-file and extra-file differences are M2 (`drift.go`).
- G-M1-7: only `./internal/template/pluginemit/...` was tested and linted (AGENTS.md section 4, plan section 4). The full `go test ./...`, repository-wide lint, the template-neutrality guard, the codemaps drift check and any other test that walks the repository root for new top-level directories (`plugins/`, `.claude-plugin/`) were not run, so an interaction with one of them is unobserved.
- G-M1-8: refusal (verification-claim-integrity section 3.1): one compound command that created the golden directory and wrote four files through shell heredocs was refused by the worktree guard ("too complex to verify"); the four files were written with the file tool instead. Nothing was substituted by reading.
- G-M1-9: the version written to the manifests is the fallback `v3.1.3` minus its `v`, while the installed binary reports `v3.2.0-rc.*` (P-22); a doctor mismatch there is OD-7 and M4, not M1.

#### Residual-risk

- `TestCommittedVersionMatchesSSOT` and the drift guard read `pkg/version.Version`, which equals the fallback only when no `-ldflags -X …Version=` is applied to the test binary; the Makefile applies LDFLAGS to `go build` only, so `make plugin-emit-check` is unaffected, but a future `go test -ldflags` would read the injected value and go red.
- The golden pins and the generator are updated by the same switch (`PLUGIN_EMIT_UPDATE=1`); a careless regeneration refreshes the shape pins together with the output. The hand-authored goldens were committed in RED, before any generator code, so their first form was an independent expectation, but later edits are reviewed only by the diff.
- Strict validation and Codex acceptance are shown at claude 2.1.288 and codex 0.160.0 only (RK-13).
- The payload (skills, commands, `.mcp.json`) does not exist yet, so the manifest descriptions describe components M2 adds; until M2 lands an install of this tree carries a plugin with no components.

### M2 (payload derivation, layout, discoverability, drift gate)

Run-phase worker: `Agent(general-purpose)` carrying the manager-develop role text, `cycle_type=tdd`, in the card worktree. Scope: M2 only (REQ-004 to REQ-009; exit AC-004 to AC-009). M3 and M4 are not started.

#### Pre-flight (recorded before any edit, 2026-10-03)

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435
$ git rev-parse --short HEAD
b49894ef3
$ git branch --show-current
WT-marketplace-core-plugin
$ git status --short
(empty)
$ env | cut -d= -f1 | grep -E '^(CLAUDE_CODE_PLUGIN_|BASH_ENV$|ENV$|BASH_FUNC_|GOBIN$|GOPATH$|GOFLAGS$|GOENV$|CODEX_SQLITE_HOME$|XDG_)'
(empty; grep exit 1)                                         # BI-1, start of the first measurement batch
$ go build ./...                                             # exit 0
$ GOOS=windows GOARCH=amd64 go build ./...                   # exit 0
$ go test ./internal/template/pluginemit/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/template/pluginemit	0.290s     # the M1 baseline, green
$ grep -c moai-factory-foreman internal/template/catalog.yaml
0                                                            # t1399 not landed
$ sh scripts/protected-set-hash.sh                           # before-line, before the first claude command
PROTECTED-SET 753334575dbf1fe141254a71de36ff7eda25227130c6454bd24e333b3491f2aa entries=190
```

B2 cross-SPEC scan (`grep -rn "Retired\|superseded" internal/template --include='*.go' -l`): nineteen unrelated files (model policy, retired wrappers and keys, tool catalog, agent frontmatter audits and similar); none concerns a plugin or marketplace emitter, so no conflict with extending `internal/template/pluginemit`.

#### Design decisions taken in M2 (read from the SPEC text, recorded because the plan leaves the seam open)

- `Emit` now takes the **raw embed layout** (`catalog.yaml` beside `templates/`), loads the catalog with `template.LoadCatalog` and applies the very tier view `moai init` uses, `template.SlimFS`, so the generator holds the filter as data and a tier-blind generator is red on the synthetic non-core entry (AC-004 (a)). The M1 test helpers (`syntheticTemplate`, the golden test's source directory) were adapted to that layout; no M1 assertion changed.
- `Publication` gains `Modes` (path to mode, the deployer rule `.sh` 0755 otherwise 0644); `drift.go` holds `Drift` (read-only: bytes, mode, missing, extra) and `Write` (the one regeneration path: writes, repairs modes with `Chmod`, removes extras inside the three generated roots). The mode comparison is skipped on Windows (no execute bit; `deployer.go` likewise just passes a perm and Windows ignores the bit).

#### Commits (tdd order: observed RED as its own commits, GREEN, REFACTOR)

| Step | SHA | Content |
|------|-----|---------|
| RED 1 | `1139af414` | the M2 tests (`payload_test.go`, `names_test.go`, `drift_test.go`, `script_test.go`), the M1 test helpers adapted to the raw embed layout, and minimal compiling stubs: `Emit` reads the layout and applies the init tier view but emits no payload, `Drift` and `Write` are empty, `scripts/check-plugin-discoverable.sh` is a placeholder that exits 1; plus the pre-flight above |
| RED 2 | `779d9b581` | the shape pin `testdata/golden/payload-mcp.json` and its entry in `goldenFiles` (the manifest golden test asserted exactly four files; the payload makes it five) |
| GREEN | `815f04c83` | `payload.go`, `drift.go` (`Drift`, `Write`), `pluginemit.go` (payload merged into `Emit`, `Modes`), `Makefile` (`plugin-emit-check` before `build`), `scripts/check-plugin-discoverable.sh`, and the 308 files under `plugins/moai/` produced by `make plugin-emit` |
| REFACTOR | `d4a3b1ad5` | the skills walk handles a missing source directory itself (one `WalkDir`, no separate `Stat`); an `@MX:ANCHOR` whose fan_in claim was not true became an `@MX:NOTE`; behavior unchanged |

Both RED commits precede GREEN in the commit graph, so the order is witnessed by git (verification-claim-integrity section 2.3). The RED step is the observed failing output below, captured at `779d9b581` before any GREEN code existed.

#### Claim

1. REQ-004: the payload is derived from the template tree through `template.LoadCatalog` plus `template.SlimFS` (the view `moai init` deploys), and the generator holds no component name; `TestEmitDerivesFromTree` emits exactly the synthetic core set for two disjoint name sets, `TestGeneratorHoldsNoComponentNames` finds no name literal in its non-test sources and fails on a hard-coded one.
2. REQ-005 and REQ-009 modes: skills are copied byte for byte (the brace-bearing files included), `.tmpl` commands are rendered with the English default context, every file is written at the deployer's mode (`.sh` 0755, otherwise 0644); the three `navigator-*.sh` files are 100755 in git although one source is 100644.
3. REQ-006 and REQ-008: commands are flat `plugins/moai/commands/<stem>.md` (17 of them, no subdirectory), the plugin root holds exactly the five allowed entries, the payload `.mcp.json` declares only `moai` with the template command and args.
4. REQ-009: `make plugin-emit-check` is a prerequisite of `build:` and reports byte, mode, missing and extra differences on the committed tree, never writing; `make plugin-emit` is the only regeneration path and also removes extra files and repairs modes.
5. AC-006 (b) to (e): `scripts/check-plugin-discoverable.sh` ran the real Claude CLI under empty scratch homes and printed `ok: 41 names listed, 0 missing` (24 skills and 17 commands); its contract is held by `TestCheckPluginDiscoverable` against a recording stand-in (BI-3).

#### Evidence (verbatim, this run, this tree)

RED, observed at assertions (`779d9b581`; `go test ./internal/template/pluginemit/... -count=1 -v`, selected lines; the full output was kept outside the repository):

```
--- FAIL: TestWriteMaterialisesTree (0.01s)
    drift_test.go:100: a stale file inside the generated root survived a regeneration
--- FAIL: TestDriftDetectsMutatedArtifact (0.10s)
    drift_test.go:168: Drift = [], want exactly {bytes plugins/moai/skills/s/SKILL.md}
    drift_test.go:180: Drift = [], want exactly {mode plugins/moai/skills/s/scripts/t.sh}
    drift_test.go:186: Drift = [], want exactly {mode plugins/moai/.claude-plugin/plugin.json}
    drift_test.go:195: Drift = [], want exactly {missing plugins/moai/commands/c.md}
    drift_test.go:209: Drift = [], want exactly {extra plugins/moai/commands/nested/extra.md}
--- FAIL: TestManifestsGolden (0.01s)
    golden_test.go:62: plugins/moai/.mcp.json: not emitted
--- FAIL: TestEmitDerivesFromTree (0.00s)    (first-names and second-names)
    payload_test.go:107: emitted set differs from the synthetic core set
    payload_test.go:113: rendered command = "", want the English default render
    payload_test.go:117: brace-bearing skill file = "", want it copied unchanged
--- FAIL: TestEmitFidelity (0.01s)
    payload_test.go:170: plugins/moai/commands/mx.md: not emitted (source .claude/commands/moai/mx.md.tmpl)    (one line per payload file)
--- FAIL: TestPayloadCommandsFlat, TestPayloadAllowList, TestPayloadMCPFile     (committed payload absent)
--- FAIL: TestCheckPluginDiscoverable (six subtests)
    script_test.go:193: exit=1 out="check-plugin-discoverable: not implemented\n", want exit 2 and a refused: line
    script_test.go:278: exit=1 out="check-plugin-discoverable: not implemented\n" verbs=0, want exit 1 naming the MCP servers line after all three verbs ran
FAIL	github.com/modu-ai/moai-adk/internal/template/pluginemit
```

`TestGeneratorHoldsNoComponentNames` PASSED at RED and could not be red there: the stub generator holds no name. Its ability to fail is shown by its in-test positive control and by the hard-coded mutant below (G-M2-5).

GREEN (HEAD `d4a3b1ad5` after REFACTOR; `go test ./internal/template/pluginemit/... -count=1 -v`):

```
--- PASS: TestWriteMaterialisesTree (0.01s)
--- PASS: TestDriftDetectsMutatedArtifact (0.04s)    (six subtests PASS)
--- PASS: TestManifestsGolden (0.00s)
--- PASS: TestGoldenCommittedArtifactsMatchEmission (0.24s)
--- PASS: TestVersionStampedFromSSOT (0.00s)
--- PASS: TestCommittedVersionMatchesSSOT (0.00s)
--- PASS: TestMCPEntryDerivedFromTemplate (0.00s)    (four subtests PASS)
--- PASS: TestGeneratorHoldsNoComponentNames (0.00s)
--- PASS: TestEmitDerivesFromTree (0.00s)            (first-names, second-names)
--- PASS: TestEmitFidelity (0.06s)
--- PASS: TestPayloadCommandsFlat (0.05s)
--- PASS: TestPayloadAllowList (0.00s)
--- PASS: TestPayloadMCPFile (0.00s)
--- PASS: TestCheckPluginDiscoverable (1.31s)        (six subtests PASS)
ok  	github.com/modu-ai/moai-adk/internal/template/pluginemit	2.039s
```

Generated payload: `find plugins/moai -type f | wc -l` printed `308` (2 plugin manifests, 1 `.mcp.json`, 17 commands, 288 skill files); `git ls-files plugins | wc -l` printed `308`; 24 skill directories; the generated files were produced by `make plugin-emit`, whose run left the four manifests and the hand-authored shape pins under `testdata/golden/` byte-unchanged (`git status` listed only the new payload).

Acceptance matrix (every command run at HEAD `d4a3b1ad5`; scratch homes `<scratchpad>/m2/{f-b,f-d}` shown empty by `find <home> -mindepth 1 | wc -l` printing `0` first, one per criterion; the exit status of a clean tool result is not echoed, an error result carries it):

| AC | Result | Command | Verbatim output |
|----|--------|---------|-----------------|
| AC-004 (a) | PASS | `go test ./internal/template/pluginemit -run '^TestEmitDerivesFromTree$' -count=1 -v` | `--- PASS: TestEmitDerivesFromTree (0.00s)`, subtests `first-names`, `second-names` PASS, `ok  	github.com/modu-ai/moai-adk/internal/template/pluginemit	0.174s` |
| AC-004 (b) | PASS | `go test ./internal/template/pluginemit -run '^TestGeneratorHoldsNoComponentNames$' -count=1 -v` | `--- PASS: TestGeneratorHoldsNoComponentNames (0.00s)` and `ok` (mutants below) |
| AC-005 (a) | PASS | `go test ./internal/template/pluginemit -run '^TestEmitFidelity$' -count=1 -v` | `--- PASS: TestEmitFidelity (0.06s)` and `ok` |
| AC-005 (b) | PASS | `diff -r internal/template/templates/.claude/skills/moai plugins/moai/skills/moai` | no output; `diff-exit=0` |
| AC-005 (c) | PASS | `grep -rlE '\{\{' plugins/moai/commands` | no output, `grep-exit=1`; control `grep -rlE '\{\{' internal/template/templates/.claude/commands/moai \| wc -l` printed `15` |
| AC-005 (d) | PASS | `find plugins/moai -name '*.sh' ! -perm 755 -print` | no output; control `find plugins/moai -name '*.sh' -print` printed the three `moai-workflow-project/scripts/navigator-{audit,enrich,regen}.sh` paths |
| AC-006 (a) | PASS | `go test ./internal/template/pluginemit -run '^TestPayloadCommandsFlat$' -count=1 -v` | `--- PASS: TestPayloadCommandsFlat (0.04s)` and `ok` |
| AC-006 (b) | PASS | `sh scripts/check-plugin-discoverable.sh <scratchpad>/m2/f-b` (real Claude CLI, claude 2.1.288) | `scrub: enumerated and unset 29 names` / `ok: 41 names listed, 0 missing` |
| AC-006 (c) | PASS | `sh scripts/check-plugin-discoverable.sh internal` | `refused: internal is not an existing empty directory`, `Exit code 2` |
| AC-006 (d) | PASS | `MOAI_CLAUDE_BIN=<recorder> CLAUDE_CODE_PLUGIN_CACHE_DIR=<canary> MOAI_PLANTED_424242=1 CLAUDE_PLANTED_424242=1 CODEX_PLANTED_424242=1 sh scripts/check-plugin-discoverable.sh <scratchpad>/m2/f-d` | `scrub: enumerated and unset 34 names` (29 of the caller plus the 5 planted, so the count exceeds the planted count) / `ok: 41 names listed, 0 missing`; then `find <canary> -mindepth 1 -print \| wc -l` printed `0`, `recorder.log` absent (`find ... -name recorder.log \| wc -l` printed `0`), `<home>/plugins/installed_plugins.json` present |
| AC-006 (e) static | PASS (static) | `grep -nE 'claude +[a-z]' scripts/check-plugin-discoverable.sh`; `grep -c codex scripts/check-plugin-discoverable.sh` | three lines (`89`, `91`, `93`: `plugin marketplace add`, `plugin install`, `plugin details`); `0` |
| AC-007 (a) | PASS | `jq -c .mcpServers.moai plugins/moai/.mcp.json internal/template/templates/.mcp.json` | `{"command":"moai","args":["mcp-server"]}` twice |
| AC-007 (b) | PASS | `jq -c '.mcpServers\|keys' plugins/moai/.mcp.json` | `["moai"]` |
| AC-007 (c) | PASS | `go test ./internal/template/pluginemit -run '^TestMCPEntryDerivedFromTemplate$' -count=1 -v` | `--- PASS: TestMCPEntryDerivedFromTemplate (0.00s)`, four subtests PASS; the payload file is also covered by `TestPayloadMCPFile` PASS |
| AC-008 (a) | PASS | `find plugins/moai -maxdepth 1 -mindepth 1 -print` | `plugins/moai/.mcp.json`, `.claude-plugin`, `commands`, `skills`, `.codex-plugin` (five lines) |
| AC-008 (b) | PASS | `go test ./internal/template/pluginemit -run '^TestPayloadAllowList$' -count=1 -v` | `--- PASS: TestPayloadAllowList (0.00s)` and `ok` |
| AC-009 (a) | PASS | `make plugin-emit-check` | `ok  	github.com/modu-ai/moai-adk/internal/template/pluginemit	0.298s` |
| AC-009 (b) | PASS | `go test ./internal/template/pluginemit -run '^TestDriftDetectsMutatedArtifact$' -count=1 -v` | `--- PASS: TestDriftDetectsMutatedArtifact (0.04s)` with PASS subtests `clean-tree-has-no-drift`, `flipped-byte`, `mode-flipped`, `deleted-file`, `extra-file`, `committed-set-unchanged` |
| AC-009 (c) | PASS (static) | `grep -n ^build: Makefile` | `34:build: agents-emit-check commands-emit-check plugin-emit-check tool-policy-drift-check templ-generate ## Build the binary`; `make -n build` lists the `PLUGIN_EMIT_UPDATE= go test ... plugin-emit drift` recipe as the third prerequisite |

Regression checks of M1 criteria with the payload present (fresh scratch homes): `claude plugin validate plugins/moai --strict` printed `Validating plugin manifest: …/plugins/moai/.claude-plugin/plugin.json` / `✔ Validation passed`; `claude plugin validate .claude-plugin/marketplace.json --strict` printed `✔ Validation passed`.

Build, lint and selectors (HEAD `d4a3b1ad5`): `go build ./...` printed `BUILD-OK` (exit 0); `GOOS=windows GOARCH=amd64 go build ./...` printed `WINBUILD-OK`; `GOOS=windows GOARCH=amd64 go vet ./internal/template/pluginemit/` (at GREEN) printed `WINVET-OK`, so the tests compile for Windows; `golangci-lint run --timeout=2m ./internal/template/pluginemit/...` printed `0 issues.` with `golangci-lint has version v2.1.6` (the CI version); `go vet ./internal/template/pluginemit/` printed `VET-OK`; `go test ./internal/template -run '^(TestTemplateNeutralityAudit|TestTemplateNeutralityAuditC8Preserve|TestLanguageNeutrality|TestMCPNeutralityTemplateShape|TestGTDCanonicalSurfaceGolden)$' -count=1` printed `ok  	github.com/modu-ai/moai-adk/internal/template	4.036s`.

Drift-gate mutants on the committed tree (E6; each run with `make plugin-emit-check`, then restored with the explicit `make plugin-emit`, `git status --short | wc -l` printed `0` after every restore):

```
hand edit   (echo x >> plugins/moai/commands/gtd.md)
  golden_test.go:110: plugins/moai/commands/gtd.md: bytes — run `make plugin-emit` or stop hand-editing
  plugin-emit drift: committed marketplace, plugin manifests or plugin payload differ from the generator — run `make plugin-emit`
  make: *** [plugin-emit-check] Error 1            (git status then listed only " M plugins/moai/commands/gtd.md": the check wrote nothing)
missing file (rm plugins/moai/commands/todo.md)
  golden_test.go:110: plugins/moai/commands/todo.md: missing — run `make plugin-emit` or stop hand-editing        (Error 1)
extra file  (new plugins/moai/skills/moai/extra-file.md)
  golden_test.go:110: plugins/moai/skills/moai/extra-file.md: extra — run `make plugin-emit` or stop hand-editing  (Error 1)
  restore: make plugin-emit removed the extra file (find plugins/moai -name extra-file.md | wc -l printed 0)
wrong mode  (chmod 644 …/navigator-audit.sh)
  golden_test.go:110: plugins/moai/skills/moai-workflow-project/scripts/navigator-audit.sh: mode — run `make plugin-emit` or stop hand-editing   (Error 1)
  restore: make plugin-emit repaired the bit (find plugins/moai -name navigator-audit.sh ! -perm 755 -print | wc -l printed 0)
wrong mode  (chmod 755 plugins/moai/.mcp.json)
  golden_test.go:110: plugins/moai/.mcp.json: mode — run `make plugin-emit` or stop hand-editing                  (Error 1)
```

Generator, drift and script mutants (each applied to a copy-backed source, the named tests run, then restored with `cmp` against the backup, `RESTORED`):

```
generator, hard-coded name (var handCopied = map[string]bool{"moai-foundation-core": true})
  names_test.go:141: the generator holds 1 component-name literal(s):  payload.go: "moai-foundation-core"
  --- FAIL: TestGeneratorHoldsNoComponentNames
generator, tier-blind (view = fs.Sub(raw, "templates") instead of the init tier view)
  --- FAIL: TestEmitDerivesFromTree (both name sets), TestEmitFidelity, TestGoldenCommittedArtifactsMatchEmission
generator, every file 0644
  golden_test.go:110: …/navigator-audit.sh: mode (three scripts); payload_test.go:205: …/navigator-regen.sh: mode -rw-r--r--, want -rwxr-xr-x
  --- FAIL: TestGoldenCommittedArtifactsMatchEmission, TestEmitFidelity
generator, commands nested (commands/mirror/<name>.md)
  --- FAIL: TestEmitDerivesFromTree, TestEmitFidelity, TestPayloadCommandsFlat
generator, .tmpl copied unrendered
  payload_test.go:113: rendered command = "---\ndescription: {{if eq .ConversationLanguage \"ko\"}}korean{{else}}english{{end}}…", want the English default render
  payload_test.go:189: plugins/moai/commands/harness.md: a rendered command carries a template action
generator, scaffold-only rules leak (plugins/moai/rules/…)
  payload_test.go:294: emitted top-level entries = [.claude-plugin .codex-plugin .mcp.json commands rules skills], want exactly [.claude-plugin .codex-plugin .mcp.json commands skills]
  --- FAIL: TestPayloadAllowList, TestEmitDerivesFromTree, TestEmitFidelity
drift, regenerates before comparing (Drift calls Write first)
  clean-tree-has-no-drift PASS; flipped-byte, mode-flipped, deleted-file, extra-file, committed-set-unchanged FAIL
drift, bytes only (mode comparison skipped)
  flipped-byte, deleted-file, extra-file, committed-set-unchanged PASS; mode-flipped FAIL
script, typed-list scrub (unset of two names, count 2)
  script_test.go:244: scrub count = 2, want at least the 5 planted names
  script_test.go:255: a call saw "CODEX_PLANTED_80268=1", want only CLAUDE_CONFIG_DIR=…   (pid-named names survive a typed list)
script, fourth verb through a variable (tool=claude; "$tool" plugin disable …)
  static (e) grep still counts 3 lines (the BI-3 limit of a static check)
  script_test.go:219: the script started 4 tool commands, want exactly 3: [plugin marketplace add … plugin install … plugin disable …]
script, remote marketplace source (modu-ai/moai-adk)
  script_test.go:229: marketplace source "modu-ai/moai-adk" is not an existing local directory
script, scrub removed
  script_test.go:241: no scrub count line in "ok: 41 names listed, 0 missing\n"
```

Real-tool scrub mutant (L-43 repeated on the real script): a copy of the script with the lines between the two markers deleted, run in a symlinked tree with `MOAI_CLAUDE_BIN=<recorder>` and `CLAUDE_CODE_PLUGIN_CACHE_DIR=<empty canary>`:

```
ok: 41 names listed, 0 missing                      # (b)-style output is unchanged
canary entries after: 389 (find <canary> -mindepth 1 | wc -l); the scratch home held settings.json, .claude.json and backups only, no plugins/
```

so the inventory line cannot see the missing scrub and only the canary of (d) does, as the SPEC says.

Protected-set bracket (AC-025 (d)), real roots, the caller's own environment:

```
$ sh scripts/protected-set-hash.sh        # before, before the first claude command (pre-flight)
PROTECTED-SET 753334575dbf1fe141254a71de36ff7eda25227130c6454bd24e333b3491f2aa entries=190
$ sh scripts/protected-set-hash.sh        # after the last claude command, tree HEAD d4a3b1ad5
PROTECTED-SET 753334575dbf1fe141254a71de36ff7eda25227130c6454bd24e333b3491f2aa entries=190
```

The two lines are equal and equal the baseline named in the dispatch: LEAK=0. BI-1 (`env | cut -d= -f1 | grep -E '^(CLAUDE_CODE_PLUGIN_|BASH_ENV$|ENV$|BASH_FUNC_|GOBIN$|GOPATH$|GOFLAGS$|GOENV$|CODEX_SQLITE_HOME$|XDG_)'`) ran at the pre-flight, before the first fixture observation, before the acceptance real-tool batch, before the final real-tool batch and at the end: every run printed nothing (`bi1-grep-exit=1`). The real `claude` verbs that ran were `plugin marketplace add <local path>`, `plugin install moai@moai-adk`, `plugin details moai@moai-adk` and `plugin validate`, each under its own empty scratch `CLAUDE_CONFIG_DIR`; no `codex` command ran in M2; nothing was aimed at a real home.

#### Baseline-attribution

Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435`, branch `WT-marketplace-core-plugin`, base HEAD `b49894ef3`. The RED lines are measured at `779d9b581`; the GREEN test list, the acceptance matrix, the build, lint and selector results, the final protected-set line and the final real-tool runs at HEAD `d4a3b1ad5`; the mutants against the GREEN tree (`815f04c83` content) before the REFACTOR commit; all in this run. The judging tools are installed binaries invoked through PATH: `go` (go1.26.8), `golangci-lint` v2.1.6, `claude` 2.1.288, `jq`, plus the shell scripts of this tree invoked by path; none is built from this tree, so section 2.2 of verification-claim-integrity asks for no build-versus-HEAD statement for them. The Go tests judge the tree's own `pluginemit` package, built by `go test` from this tree. `scripts/check-plugin-discoverable.sh` judged this tree's committed payload and the template tree it names.

#### Gaps

- G-M2-1: `claude` read 2.1.288 here, the SPEC's observations (P-30, P-49) are at 2.1.287. The `plugin details` text shape the script parses was observed at 2.1.288 only; the script fails on a shape it cannot parse (exit 1, `cannot parse the inventory`) rather than passing.
- G-M2-2: the exit codes of the clean real-tool commands and of `diff`/`find` were read from the tool result, not echoed; where an echo was possible (`diff-exit=0`, `grep-exit=1`, `bi1-grep-exit=1`) it is shown. The worktree guard refused every command that contained a shell variable (the scratch path held in `$SP`, a first mutant script written with a heredoc into python over a variable path); each was re-run with literal absolute paths or the file tool. Nothing was substituted by reading the source (verification-claim-integrity section 3.1).
- G-M2-3: REQ-009 says the build fails when the template tree changes without regeneration. The derivation tests change a synthetic tree and the drift mutants change the committed tree; no real template file was edited to watch `plugin-emit-check` turn red (the dispatch limits M2 to reading the template tree). `make build` itself was not run, only `make -n build` (the prerequisite order), because its later steps (`templ-generate`, `gen-catalog-hashes --all`) write files outside M2.
- G-M2-4: the drift mode comparison is "executable or not" (git records only 0644 and 0755, and a checkout applies the user's umask to the rest) and is skipped on Windows; a mode change that keeps the execute bit unchanged (0644 to 0600) is not reported. `TestEmitFidelity` and `TestWriteMaterialisesTree` compare exact modes, but only on emission and on a fresh write.
- G-M2-5: `TestGeneratorHoldsNoComponentNames` was green at RED (the stub holds no name), so it has no RED of its own; its failure is shown by the in-test positive control (path-segment form, lone literal, and a sentence that must not hit) and by the hard-coded mutant above. It scans string literals only, so a name built from pieces at run time (a concatenation) would pass it.
- G-M2-6: `scripts/check-plugin-discoverable.sh` runs on demand and needs a Claude CLI (SPEC G-7). `TestCheckPluginDiscoverable` runs it in every `go test` of the package against a recording stand-in and therefore keeps firing, but only the stand-in's version of the three verbs and of the inventory text; the real-tool behaviour was observed in the runs above, not in the test. The test is skipped on Windows and where `sh` is absent.
- G-M2-7: the AC-006 (d) poison is a `MOAI_CLAUDE_BIN` recorder that the script never reaches (it calls `claude` from PATH, BI-3 iii); only the cache-variable canary is a live control, as the SPEC states. The pid-named names of (d) were planted as `MOAI_PLANTED_424242` and the like, a fixed number and not the shell's own process id, which a typed list also could not contain; the Go test plants names with the test binary's real pid.
- G-M2-8: nothing outside `./internal/template/pluginemit/...`, the five named `internal/template` selectors and `go build ./...` was tested (AGENTS.md section 4). The full suite, other tests that walk the repository root and might see the new top-level `plugins/` directory or its 308 files, repository-wide lint and markdown or i18n checks over `plugins/**` were not run; a grep of `.github/workflows`, `scripts` and the Makefile for `**/*.md` style globs found none.
- G-M2-9: field effect in Codex (SPEC G-4) is unchanged and untested here: no `codex` command ran in M2, and the Codex manifest still names `skills: "./skills/"`, which now exists, but nothing read it. The two-copy behaviour of the plugin and the scaffold (SPEC G-2) and the invocation names `moai:<name>` stay unobserved.
- G-M2-10: the payload carries three `.gitkeep` files that mirror template files (for example `plugins/moai/skills/moai/workflows/plan/.gitkeep`); they are what `moai init` would deploy and are harmless, but they are listed in the plugin's files.

#### Residual-risk

- The committed payload is 308 generated files; a later template edit that skips `make plugin-emit` turns `plugin-emit-check` and the golden test red, which is the intended signal. The card that lands second after t1399 owns that red (plan section 5); the generator needs no edit for the rename because it holds no names.
- `Emit` now takes the raw embed layout and the golden test reads it from disk (`internal/template/`), not from the embedded copy in a built binary; the two are equal only after `make build`. A caller that wants the payload from a binary needs an exported raw layout, which `embed_catalog.go` deliberately withholds.
- The script's expectations come from `plugins/moai/skills` and the template command stems; a skill directory in the payload that the template does not carry would be expected and listed, so the script alone does not catch a stray payload skill (the drift gate does).
- Strict validation, the inventory shape and the 41-name count are shown at claude 2.1.288 only (RK-13).

### M3a (the Go half of M3: install step, runner seam, verb, opt-out, resolver variant)

Run-phase worker: `Agent(general-purpose)` carrying the manager-develop role text, `cycle_type=tdd`, in the card worktree. Scope: M3a only (REQ-010 to REQ-017 and REQ-019 in Go; exit AC-010 to AC-017 and AC-019 (a), (b)). Left for M3b and M4, named here so no later reader assumes them: the install scripts and the docs-site copies, `.github/workflows/test-install.yml`, `TestInstallScriptsPluginStepGuarded` (AC-018 (c); it reads the scripts and cannot pass before they change), the remaining harness cases (`isolation-*`, `verb-*`, `init-*` under OD-14, the three flags) and so AC-018 (a), AC-019 (c) and AC-025, and all of M4 (the doctor check).

#### Pre-flight (recorded before any edit, 2026-10-03)

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435
$ git rev-parse --short HEAD
ce863c085
$ git branch --show-current
WT-marketplace-core-plugin
$ git status --short
(empty)
$ env | cut -d= -f1 | grep -E '^(CLAUDE_CODE_PLUGIN_|BASH_ENV$|ENV$|BASH_FUNC_|GOBIN$|GOPATH$|GOFLAGS$|GOENV$|CODEX_SQLITE_HOME$|XDG_)'
(empty; grep exit 1)                                         # BI-1, start of the first measurement batch
$ go build ./...                                             # exit 0
$ GOOS=windows GOARCH=amd64 go build ./...                   # exit 0
$ go test ./internal/cli -run '^TestBinaryLag_.*$' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	2.965s
$ grep -rn 'runInit(' internal/cli/*_test.go | wc -l
      33                                                     # plan section 5: 33 lines
$ grep -rn 'initCmd.RunE(' internal | wc -l
       9
$ grep -rn '"init"' internal --include='*.go' | grep -v '_test.go' | grep -v 'git' | head
internal/cli/help_order.go:32:	"project": {"init", "status", "doctor", "update", "migrate", "pr"},
internal/cli/doctor_agentemit_embed.go:352:	cmd := exec.Command(execPath, "init", target, "--non-interactive", "--llm", "both")
internal/cli/init.go:449:	wtPath := enterSessionWorktree(swCfg, "init", cmd.ErrOrStderr())
internal/cli/init.go:457:				wtPath, cerr, "init")
internal/hook/types.go:125:	// (official trigger values: "init" | "maintenance"; carried in the shared
                                                             # the one process-start caller is doctor_agentemit_embed.go:352 (1, as planned)
$ grep -n "init " e2e/cli/tux3_journeys.sh        # two runs, lines 104 and 115, as planned
$ sh scripts/protected-set-hash.sh                           # before-line
PROTECTED-SET 753334575dbf1fe141254a71de36ff7eda25227130c6454bd24e333b3491f2aa entries=190
```

No count differs from the plan. B2 cross-SPEC scan (`grep -rn "Retired\|superseded" internal/cli`): hits in `cg.go`, `glm.go`, `harness.go` and a handful of tests (retired CG routing, model-override and harness retirement); none concerns plugins, the install step or the Claude resolver, so no conflict. B3: `grep -rn 'AskUserQuestion' internal/cli/plugin_install*.go` prints nothing (grep exit 1).

#### Design decisions taken in M3a (read from the SPEC text; the plan leaves the seam open)

- The runner seam is an interface (`pluginCommandRunner.Run(ctx, bin, args, env) ([]byte, error)`) with one production type, `execPluginRunner`, held in the package variable `pluginRunner`. It returns the combined output so the guidance reason can carry a bounded tail and so the M4 doctor probe can read `codex plugin list --json` through the same variable. The default refuses under a test binary with `errPluginRunnerRefused`.
- A second, silent pre-check sits in the step: when the active runner is the default type and the process is a test binary the step returns before printing anything. Without it the 33 `runInit` and 9 `initCmd.RunE` test sites would each print a config-home line and a guidance block. The runner's own refusal remains the real guard; the two are independent and each has its own test (see Mutants).
- The test-binary detector is `testing.Testing() || isPluginTestProgram(os.Args[0])`; `isPluginTestProgram` looks at the program name only (`.test` or `.test.exe` on the part after the last `/` or `\`). `isTestEnvironment()` in `glm.go` is not reused.
- `claude_binary.go` gains `claudeNotFoundError` (same message the launcher always printed) and `resolveClaudeBinaryAt(projectRoot)`; `resolveLaunchClaudeBinary()` is now `resolveClaudeBinaryAt` over `findProjectRoot()` and its behavior and error text are unchanged.
- `moai init` passes the project it initialises (`opts.ProjectRoot`) so the `llm.claude_bin` pin read is that project's; `moai plugin install` passes the project the working directory sits in, as the launcher does.
- The config-home line prints the variable's value, else the literal `~/.claude` or `~/.codex`. One success line is printed per tool; the tests count only the skip line and the guidance block.
- `MOAI_SKIP_PLUGIN_INSTALL` opts out on `1` or `true` (case-insensitive); empty and `0` do not.
- `--no-plugin` is registered on `initCmd` only. The mirrored `newInitTestCmd()` in the existing tests has no such flag, and `getBoolFlag` returns false for an absent flag, so no existing test needed an edit.

#### Commits (tdd order: observed RED as its own commit, then GREEN, then REFACTOR)

| Step | SHA | Content |
|------|-----|---------|
| RED | `0080bfb09` | the unit, init, verb, guard and resolver tests, with compiling stubs that perform nothing; the two config constants land here because the tests name them |
| GREEN | `93dee3f2f` | the step, the runner seam, the verb, the `--no-plugin` flag and the init call, the resolver variant, the opt-out on the doctor re-entry and on the two e2e lines |
| REFACTOR | `209de9f55` | a zero-timeout fallback removed (untested), `@MX:WARN` on the real runner, the timeout constant's comment reworded, two tests added that kill mutants the first GREEN left alive |

The RED commit precedes the GREEN commit in the commit graph, which is the only witness of the order (verification-claim-integrity section 2.3). The commit carrying this record follows the REFACTOR commit.

#### RED (observed before any GREEN code, at `0080bfb09`)

Command (one invocation, 15 anchored names, BI-1 printed nothing first): `go test ./internal/cli -count=1 -v -run '^(TestPluginInstallStep_Sequence|TestPluginInstallStep_FailOpen|TestPluginInstallStep_ToolAbsent|TestPluginInstallStep_OptOut|TestPluginInstallStep_Environment|TestPluginInstallStep_NoRealRunnerUnderTest|TestPluginTestProgramDetector|TestInitPluginStep_AfterDeployment|TestInitPluginStep_HarnessGating|TestPluginInstallCmd|TestResolveClaudeBinaryAt_NotFoundIsTyped|TestResolveClaudeBinaryAt_InvalidPinIsNotTheNotFoundClass|TestResolveClaudeBinaryAt_ReadsPinFromGivenRoot|TestPluginOptOutCallersEnumerated|TestDoctorAgentEmitEmbed_SetsPluginOptOut)$'`

Result: exit 1, `FAIL github.com/modu-ai/moai-adk/internal/cli 53.593s`. No panic, no `[build failed]`, no `no tests to run`; the swept set is 15 top-level tests and 41 subtests. Per-selector counts (RED, then the same names at the final HEAD):

| Test | RED top | RED subtests fail/pass | Final top | Final subtests fail/pass |
|------|---------|------------------------|-----------|--------------------------|
| TestPluginInstallStep_Sequence | FAIL | 6 / 1 | PASS | 0 / 7 |
| TestPluginInstallStep_Environment | FAIL | 4 / 0 | PASS | 0 / 4 |
| TestPluginInstallStep_FailOpen | FAIL | 6 / 1 | PASS | 0 / 7 |
| TestPluginInstallStep_ToolAbsent | FAIL | 3 / 0 | PASS | 0 / 3 |
| TestPluginInstallStep_OptOut | FAIL | 2 / 3 | PASS | 0 / 5 |
| TestPluginInstallStep_NoRealRunnerUnderTest | FAIL | 2 / 2 | PASS | 0 / 5 (one added in REFACTOR) |
| TestPluginInstallCmd | FAIL | 5 / 1 | PASS | 0 / 6 |
| TestInitPluginStep_HarnessGating | FAIL | 5 / 0 | PASS | 0 / 5 |
| TestInitPluginStep_AfterDeployment | FAIL | none | PASS | none |
| TestPluginTestProgramDetector | FAIL | none | PASS | none |
| TestResolveClaudeBinaryAt_{NotFoundIsTyped, InvalidPinIsNotTheNotFoundClass, ReadsPinFromGivenRoot} | FAIL x3 | none | PASS x3 | none |
| TestPluginOptOutCallersEnumerated | FAIL | none | PASS | none |
| TestDoctorAgentEmitEmbed_SetsPluginOptOut | FAIL | none | PASS | none |

Verbatim failing assertions (excerpt of `red1.log`, each line the test's own message):

```
claude_binary_root_test.go:22: absent claude returned <nil>, want a *claudeNotFoundError
plugin_install_test.go:135: vectors:
plugin_install_test.go:172: want claude add (failed) + both Codex commands = 3 calls, got []
plugin_install_test.go:328: production wiring bound = 0s, want config.DefaultPluginInstallCommandTimeout (1m0s)
plugin_install_test.go:405: want exactly one skip line, got 0:
plugin_install_test.go:516: default runner returned <nil>, want errPluginRunnerRefused
plugin_install_test.go:530: runInit never reached the plugin step with an injected runner
plugin_install_test.go:593: isPluginTestProgram("/tmp/go-build123/b001/cli.test") = false, want true
plugin_install_test.go:635: init made no plugin call at all
plugin_install_test.go:679: harness map[llm:both]: claude=false codex=false, want claude=true codex=true ([])
plugin_install_guard_test.go:135: ../../internal/cli/doctor_agentemit_embed.go starts `init` but does not reference config.EnvSkipPluginInstall
plugin_install_guard_test.go:154: tux3_journeys.sh:104 runs `'$BIN' init` without MOAI_SKIP_PLUGIN_INSTALL: run_to j1-init 180 "$SANDBOX" "NO_COLOR=1 '$BIN' init proj-j1 --non-interactive --language go --git-mode manual"
plugin_install_guard_test.go:186: the init child did not receive MOAI_SKIP_PLUGIN_INSTALL=1:
```

Eight subtests passed at RED (`OptOut/flag`, `env-1`, `env-true`, `Sequence/no-scope-argument`, `FailOpen/returns-nil`, `PluginInstallCmd/exit-nonzero-on-unknown-flag`, and the two inertness subtests `pin-and-path-shims-untouched`, `rune-callers-covered`): a stub that does nothing satisfies "zero calls", "no scope argument", "returns nil" and "no process started", so these cannot be red against it. The inertness test carries two controls for that: `step-is-reached-positive-control` (RED, an injected runner must record under `runInit`) and, added in REFACTOR, `step-silent-under-default-runner`; the empty-record subtests are shown able to fail by the mutants below, not by a RED.

#### GREEN and the final run (HEAD `209de9f55`)

The final run, same 15 names plus `TestInitPluginStep_OptOut` (added in REFACTOR), BI-1 first (printed nothing, exit 1):

```
$ go test ./internal/cli -count=1 -v -run '^(...16 names...)$'      # output kept in a scratch file
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	24.874s
top PASS 16 FAIL 0 SKIP 0; sub PASS 45 FAIL 0 SKIP 0
```

Guards and neighbours in one run (`-run '^(TestRootCmd_.*|TestHelpGroupOrder_.*|TestHelpGolden_.*|TestNoPhantomBrain|TestHelpRegisteredCommands|TestDoctorCmd_IsSubcommandOfRoot|TestCharacterize_Help_.*|TestInventory_RegisteredOnRoot|TestMCPCmdRegisteredOnRoot|TestBinaryLag_.*|TestInitCmd_.*|TestRunInit_.*|TestResolveLaunchClaudeBinary_.*|TestValidateClaudeBinaryPin_.*)$'`):

```
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	128.971s
top PASS 97 FAIL 0 SKIP 0
```

`go test ./internal/config -count=1`: `ok  github.com/modu-ai/moai-adk/internal/config  12.007s` (measured at the REFACTOR content except for a comment-only edit of `defaults.go`; the build and lint were re-run after it). `go build ./...` exit 0 and `GOOS=windows GOARCH=amd64 go build ./...` exit 0 at `209de9f55`. `golangci-lint run --timeout=5m ./internal/config/... ./internal/cli/` at v2.1.6: `0 issues.` (the first run reported one ST1008 in a new test helper, fixed before the GREEN commit; no baseline issue existed in the two packages). `go vet ./internal/cli ./internal/config` clean. `gofmt -l` lists `internal/cli/mcp_claude.go` and `internal/config/slice.go`, neither touched by this milestone.

#### AC matrix (PASS only where the AC's own command was run and the expected output observed)

| AC | Result | Command | Observed |
|----|--------|---------|----------|
| AC-010 (a) | PASS | `go test ./internal/cli -run '^TestInitPluginStep_AfterDeployment$' -count=1 -v` | `--- PASS: TestInitPluginStep_AfterDeployment (1.43s)` |
| AC-010 (b) | PASS | `... -run '^TestInitPluginStep_HarnessGating$' ...` | `--- PASS` with `claude-default`, `claude-explicit`, `gpt`, `both`, `unrecognized-falls-back-to-claude` |
| AC-011 (a) | PASS | `... -run '^TestPluginInstallStep_Sequence$' ...` | `--- PASS` with `both-present`, `claude-only`, `codex-only`, `install-skipped-after-add-fails`, `already-present-is-success`, `no-scope-argument`, `pinned-binary-honored` |
| AC-011 (b) | PASS (deviation, see G-M3a-2) | `claude plugin marketplace add <repo root> --json` twice, `claude plugin install moai@moai-adk --json` twice, scratch `CLAUDE_CONFIG_DIR` | exit 0 on all four; `Successfully added marketplace: moai-adk`; second add `already on disk`; `Successfully installed plugin: moai@moai-adk (scope: user)`; second install `is already installed (scope: user)` |
| AC-011 (c) | PASS (same deviation) | `codex plugin marketplace add <repo root>` twice, `codex plugin add moai@moai-adk` twice, scratch `CODEX_HOME` | exit 0 on all four; `Added plugin \`moai\` from marketplace \`moai-adk\`.` printed by both `plugin add` runs |
| AC-012 (a) | PASS | `... -run '^TestPluginInstallStep_Environment$' ...` | `--- PASS` with `env-unchanged`, `prints-config-home-claude`, `prints-config-home-codex`, `default-home-when-unset` |
| AC-013 (a) | PASS | `... -run '^TestPluginInstallStep_FailOpen$' ...` | `--- PASS` with `exit-nonzero`, `timeout`, `bound-equals-constant`, `invalid-pin`, `add-fails-no-install`, `guidance-names-both-commands`, `returns-nil` |
| AC-013 (b) | PASS | `grep -n DefaultPluginInstallCommandTimeout internal/config/defaults.go` | one line: `185:	DefaultPluginInstallCommandTimeout = 60 * time.Second` (static; (a) is the behavior check) |
| AC-014 (a) | PASS | `... -run '^TestPluginInstallStep_ToolAbsent$' ...` | `--- PASS` with `claude-absent`, `codex-absent`, `both-absent` |
| AC-015 (a) | PASS | `... -run '^TestPluginInstallStep_OptOut$' ...` | `--- PASS` with `flag`, `env-1`, `env-true`, `env-empty-is-not-optout`, `env-0-is-not-optout` |
| AC-015 (b) | PASS | `grep -c MOAI_SKIP_PLUGIN_INSTALL internal/config/envkeys.go` | `1` (static) |
| AC-016 (a) | PASS | `... -run '^TestDoctorAgentEmitEmbed_SetsPluginOptOut$' ...` | `--- PASS: TestDoctorAgentEmitEmbed_SetsPluginOptOut (0.25s)` |
| AC-016 (b) | PASS | `... -run '^TestPluginOptOutCallersEnumerated$' ...` | `--- PASS: TestPluginOptOutCallersEnumerated (0.82s)`; the tool result of one earlier parallel invocation was exit 144 with no test output (a tool-side kill, not a test result); the single re-run printed the PASS line |
| AC-016 (c) | PASS (static) | `grep -c MOAI_SKIP_PLUGIN_INSTALL e2e/cli/tux3_journeys.sh` | `2` |
| AC-017 | PASS | `... -run '^TestPluginInstallStep_NoRealRunnerUnderTest$' ...` | `--- PASS` with `default-runner-refuses`, `step-silent-under-default-runner`, `step-is-reached-positive-control`, `pin-and-path-shims-untouched`, `rune-callers-covered` (the spec names three subtests; two controls were added) |
| AC-019 (a) | PASS | `... -run '^TestPluginInstallCmd$' ...` | `--- PASS` with `registered-in-root-help`, `help-names-opt-out`, `exit-0-on-fail-open-outcomes`, `exit-nonzero-on-unknown-flag`, `no-harness-filter`, `project-pin-read-from-working-directory` (one subtest added) |
| AC-019 (b) | PASS | `moai-m3a plugin install --help` (binary built from this tree with `go build -o <scratch>/moai-m3a ./cmd/moai`, invoked by path) | exit 0; a `--bogus` flag exits 1 |
| AC-019 (c), AC-018, AC-025 | GAP | the harness cases and the scripts | not built here; M3b |

Not on the matrix and shown only as extra observations: the real `moai plugin install` binary run from a scratch working directory under `env -i PATH=<shim dir>:/usr/bin:/bin CLAUDE_CONFIG_DIR=<scratch> CODEX_HOME=<scratch> MOAI_HOME=<scratch>` with recording shim `claude` and `codex` on PATH printed both config homes and both success lines, exited 0 and recorded, in order, `claude plugin marketplace add modu-ai/moai-adk`, `claude plugin install moai@moai-adk`, `codex plugin marketplace add modu-ai/moai-adk`, `codex plugin add moai@moai-adk`, each child seeing its scratch home; the same run with `MOAI_SKIP_PLUGIN_INSTALL=1` added no record. That is the real default runner on a non-test binary, against shims only.

#### Mutants (verification-completeness section 2; each applied to the GREEN tree, run, then restored with `git checkout -- <file>`)

Killed:
- default runner without its refusal: `default-runner-refuses` red; the other inertness subtests stayed green because the step's pre-check shields them (two sufficient defenses defeat a single removal, which is why `step-silent-under-default-runner` was added).
- the step's pre-check removed alone: `step-silent-under-default-runner` red.
- both removed: `default-runner-refuses`, `step-silent-under-default-runner`, `pin-and-path-shims-untouched` (record `claude-pinned plugin marketplace add modu-ai/moai-adk ...`), `rune-callers-covered` (record `claude-pinned plugin marketplace add ...`) all red, so the empty-record assertions can fail.
- an argument-keyed detector (any argument ending `.test`): the same four plus `TestPluginTestProgramDetector` (`isPluginTestBinary() = false inside go test`).
- install after a failed add: `install-skipped-after-add-fails`, `exit-nonzero`, `timeout`, `add-fails-no-install`, `guidance-names-both-commands`.
- guidance printed for an absent tool: `claude-absent`, `codex-absent`, `both-absent`.
- an invalid pin treated as absent: `invalid-pin`.
- a child environment with a variable added: `env-unchanged`.
- the opt-out flag ignored: `flag`; `0` or any non-empty value counted as opt-out: `env-0-is-not-optout`.
- no per-command deadline: the `timeout` subtest hangs and the run ends `panic: test timed out after 25s`.
- a production bound of 30 s: `bound-equals-constant`.
- the pin ignored for PATH: `both-present`, `claude-only`, `install-skipped-after-add-fails`, `pinned-binary-honored`.
- a verb that acts on Claude only: `no-harness-filter`; a verb that returns an error after the step: `exit-0-on-fail-open-outcomes`, `no-harness-filter`, `project-pin-read-from-working-directory`.
- an init wiring that ignores the harness: `claude-default`, `claude-explicit`, `gpt`, `unrecognized-falls-back-to-claude`; the call moved before deployment: `AfterDeployment` (`2 plugin call(s) ran before the deployed file set was complete`); an init wiring that never reads `--no-plugin`: `TestInitPluginStep_OptOut/flag`.
- the doctor child without the opt-out (`=0`): `TestDoctorAgentEmitEmbed_SetsPluginOptOut` red while the static caller list stayed green; the e2e line without the variable and a new `exec.Command(bin, "init", ...)` in a non-test source: `TestPluginOptOutCallersEnumerated` red, each naming the file or line.

Not killed or not attempted: none of the attempted ones survived. Not attempted: a verb that "ignores --llm" (the verb has no such flag; `no-harness-filter` asserts the flag is absent and that all four calls happen), a runner whose refusal keys on the arguments of one command rather than the process (covered in kind by the argument-keyed detector mutant, not run separately), and any Windows-specific behavior.

#### Protected-set bracket and BI-1

Before: `PROTECTED-SET 753334575dbf1fe141254a71de36ff7eda25227130c6454bd24e333b3491f2aa entries=190` (the dispatch baseline). After my last real-tool command: `PROTECTED-SET e2f43ca5331dcbb7c23bcf84e7575e79390e35ac22376eb751fb88dc2de2f43e entries=190`. **The two lines are not equal, so the bracket does not show LEAK=0.** Attribution, from the dumps (`--dump`), not assumed: the entry listings are identical between the after-dumps taken at 11:27, 11:37 and 11:41, and the only line that moved between any two of them is `CONTENT <hash> /Users/goos/.codex/config.toml`; no path under the protected roots has a birth time later than 10:30 (the baseline was taken before 10:50); `~/.codex/config.toml` was rewritten at 11:12:40, 11:31:32 and 11:37:21 and its newest stanzas are `[projects."/private/var/folders/.../T/TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey764495410/001"]` and `TestCodexLive_ReviewStartBaseBranchIsNotRejected2038186140`, test names of other lanes; no stanza names any test, project or path of this milestone (a grep for `TestPluginInstall`, `TestInitPluginStep`, `TestResolveClaude`, `TestDoctorAgentEmitEmbed`, `TestPluginOptOut`, `plugin-proj`, `optout-proj`, `tier-proj`, `rune-proj`, `t1435` matches one stanza, a `develop` worktree scratchpad of another lane); the Codex desktop app (pid 11250, started 09:01) is live on this machine, and a `codex app-server` (pid 12007) started at 11:37:20, one second before the 11:37:21 rewrite, at a moment when none of my commands ran `codex` (my only earlier `codex` run was `codex --version`, and the AC-011 (c) commands started at 11:37:55); the `codex` commands of AC-011 (c) ran with `CODEX_HOME` pointed at a scratch directory after the 11:37:21 rewrite and the file's mtime did not move afterwards. Two hashes taken twenty seconds apart at 11:28 were equal (`77debbee...`), so the value is stable between ambient writes. What this does not establish: the entry-by-entry baseline dump was not kept (only its hash), so the difference between the baseline and the 11:27 dump cannot be shown line by line (G-M3a-1).

BI-1 (`env | cut -d= -f1 | grep -E '^(CLAUDE_CODE_PLUGIN_|BASH_ENV$|ENV$|BASH_FUNC_|GOBIN$|GOPATH$|GOFLAGS$|GOENV$|CODEX_SQLITE_HOME$|XDG_)'`) ran at the pre-flight, before the RED run, before each later go-test measurement batch and at the end, after the last real-tool command: every run printed nothing (`bi1-exit=1`). The real-tool commands themselves ran under `env -i`, so the ambient environment did not reach them. The real tools that ran were `claude plugin marketplace add`, `claude plugin install` and `codex plugin marketplace add`, `codex plugin add`, each under its own empty scratch home with `env -i PATH=...`; `claude --version` and `codex --version` only otherwise. Nothing was aimed at a real home. The mutant runs that removed the runner's refusal could start only recording scripts and an inert pinned file (the tests pin `MOAI_CLAUDE_BIN`, put recording shims first on PATH and point the config homes at scratch directories).

#### Claim

M3a delivers REQ-010 to REQ-017 and REQ-019 in Go: `moai init` and `moai plugin install` run the two-command sequence per tool through an injected runner whose default refuses under a test binary, fail open, honor the opt-out by flag and environment, print the config home, and the two automated `init` callers set the opt-out. AC-010 to AC-017 and AC-019 (a) and (b) are shown by their own commands; AC-018, AC-019 (c) and AC-025 belong to M3b.

#### Evidence

The commands and verbatim output are in the RED, final-run, AC-matrix, mutant and bracket sections above.

#### Baseline-attribution

Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435`, branch `WT-marketplace-core-plugin`, base HEAD `ce863c085`. RED measured at `0080bfb09`; the first GREEN selector run and the mutants against the tree at `93dee3f2f` (the mutants with the one uncommitted test addition); the final selector run, the guard run, the build, the lint, the AC-matrix commands and the scratch-built `moai-m3a` at `209de9f55`; all in this run. The judging tools are installed binaries invoked through PATH, none built from this tree, so section 2.2 of verification-claim-integrity asks for no build-versus-HEAD statement for them: `go` go1.26.8, `golangci-lint` v2.1.6, `claude` 2.1.288, `codex` 0.160.0. The Go tests judge the tree's own `internal/cli` and `internal/config` packages, built by `go test` from this tree; the verb help and the shim run judged `moai-m3a`, built from this tree at `209de9f55` (its `version` carries no commit stamp because `make build`'s ldflags were not used).

#### Gaps

- G-M3a-1: the protected-set bracket is not equal (see the section above); the cause is attributed, not proven, and the baseline entry dump was not kept. A later protected-set check should save a dump with the before-line.
- G-M3a-2: AC-011 (b) and (c) name `./` and `.` as the source with the repository root as the working directory; both ran with the absolute worktree path from a scratch working directory under `env -i PATH=... CLAUDE_CONFIG_DIR|CODEX_HOME=<scratch>` (the M2 precedent). `claude` read 2.1.288 and `codex` 0.160.0 here, the SPEC's observations are at 2.1.287 and 0.160.0.
- G-M3a-3: the real binary's `init` is not driven (OD-14 default (a), SPEC G-8); init's call into the step is covered by Go tests with an injected runner and by the inertness subtests, not by a real process start. The real default runner was exercised only by the manual verb run against shims.
- G-M3a-4: nothing here contacts the GitHub source `modu-ai/moai-adk` (SPEC G-1); the product commands were observed only through the injected runner and the shims.
- G-M3a-5: only `internal/cli` selectors (the 16 plugin names, the guards and neighbours listed above, 97 top-level tests in the second run) and the full `internal/config` package were run (AGENTS.md section 4); the full `internal/cli` suite and other packages were not.
- G-M3a-6: `TestInitPluginStep_AfterDeployment` judges "deployment complete" by four sentinel files (`CLAUDE.md`, `.claude/settings.json`, `.moai/config/sections/quality.yaml`, `.moai/manifest.json`), not by the whole deployed set; a step moved to a point after those four but before a later deployment write would pass it.
- G-M3a-7: the recording-script tests (`default-runner-refuses`, the two inertness subtests, the doctor child test) skip on Windows; the Windows check is the cross-build and the detector table that includes a `.test.exe` name.
- G-M3a-8: root `moai --help` has two renderers; the fang grouped help (the production `Execute` path) lists `plugin` in TOOLS, the curated TUI table in `help.go` (`rootHelpGroups`) does not and was not changed.
- G-M3a-9: the verb prints one pre-existing `level=WARN msg="config sections directory not found, using defaults"` line when run outside a project; it comes from the root command's config loading and is not part of this milestone.
- G-M3a-10: every command containing a shell variable or a python heredoc that named a git verb was refused by the worktree guard at least once; each was re-run with literal paths or the file tool. Nothing was substituted by reading the source (verification-claim-integrity section 3.1).
- G-M3a-11: the exit codes of the `go test` and `grep` runs were read from the tool results where no echo was possible; echoes exist for `bi1-exit`, `help-exit` and the real-tool commands.

#### Residual-risk

- `moai init` now starts real `claude` and `codex` processes against the invoking person's real profile in production. The opt-out, the harness gate, the 60-second per-command bound and the fail-open outcomes are tested; RK-15's worst case of 240 seconds (four commands that all hang) is unchanged.
- The guidance and skip wording is a proposal (design.md section 3.6); the tests count lines and assert the manual commands and two marker strings (`Install it yourself`, `not found on PATH`), so a reword that keeps them passes.
- `testing` is imported by a non-test source (`testing.Testing()`); the linter accepts it and the binary-size effect was not measured.
- The runner refusal and the step's pre-check are two sufficient defenses; each removal is caught by its own test, but a future change that touches both together relies on the inertness subtests, which were shown red only against the both-removed mutant.
- The runner seam returns output for the M4 doctor probe, which does not exist yet; its use there is unobserved.
- A `moai init` run by someone whose `claude` is a pinned wrapper that prompts for input would hang until the 60-second bound, then print guidance.

### M3b (the scripts, CI and harness half of M3: installers, docs-site copies, test-install.yml, harness cases and flags, static guard)

Run-phase worker: `Agent(general-purpose)` carrying the manager-develop role text, `cycle_type=tdd`, in the card worktree. Scope: M3b only (AC-018, AC-019 (c), AC-025). Out of scope and untouched: all of M4, Go install-step logic, templates. The worker was cut off once by a session rate limit; it resumed from the committed and working-tree state at `f453a6875` with an uncommitted harness edit, re-ran the pre-flight, judged that edit coherent, and committed it as its own RED commit before touching any install script.

#### Pre-flight (recorded before any edit)

```
$ git rev-parse --show-toplevel / --short HEAD / branch --show-current / git status --short
/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435 ; 0c0ebd673 ; WT-marketplace-core-plugin ; (empty)
$ env | cut -d= -f1 | grep -E '^(CLAUDE_CODE_PLUGIN_|BASH_ENV$|ENV$|BASH_FUNC_|GOBIN$|GOPATH$|GOFLAGS$|GOENV$|CODEX_SQLITE_HOME$|XDG_)'
(empty; grep exit 1)                       # BI-1, printed first in every measuring command of this milestone, always empty
$ go build ./...                           # exit 0
$ GOOS=windows GOARCH=amd64 go build ./... # exit 0
$ go test ./internal/cli -count=1 -run '^(TestPluginInstallStep_|TestPluginInstallCmd|TestInitPluginStep_|TestPluginOptOutCallersEnumerated)'
ok  	github.com/modu-ai/moai-adk/internal/cli	21.525s
$ go build -o bin/moai ./cmd/moai          # exit 0 (bin/ is gitignored)
$ sh scripts/protected-set-hash.sh --dump <scratchpad>/m3b/dump-before.txt
PROTECTED-SET e2f43ca5331dcbb7c23bcf84e7575e79390e35ac22376eb751fb88dc2de2f43e entries=190   # the dispatch baseline
$ sh scripts/test-plugin-install-step.sh bin/moai        # the R0 harness before any change
PENDING installer-calls-verb-by-installed-path: install-directory assertions pass; the verb-call assertion lands with install.sh in M3
PASS installer-optout / PASS installer-set-e-guard / PASS installer-old-binary-unknown-verb
RESULT pass=3 fail=0 pending=1
LEAK=0 (real roots unchanged: PROTECTED-SET e2f43ca5...2f43e entries=190)
$ ... --negative-control-install-dir bin/moai : four RED, RESULT negative-control-install-dir: red set and green set are exactly the expected ones, LEAK=0
$ ... --negative-control-decoy-missing bin/moai : two PASS, RESULT negative-control-decoy-missing: the installer never ran with a decoy missing, LEAK=0
```

After the rate-limit resume, at HEAD `f453a6875` (`git status --short`: ` M scripts/test-plugin-install-step.sh` only), a fresh BEFORE dump was taken, because the first one could no longer be trusted as a baseline:

```
$ sh scripts/protected-set-hash.sh --dump <scratchpad>/m3b/dump-before2.txt
PROTECTED-SET 5d8c23fdb87a8e15dcab32dbef5aad51741fa6db4affcb72ddd42625492b061c entries=190
$ diff dump-before.txt dump-before2.txt
190c190
< CONTENT fc0b9b20...baf88 /Users/goos/.codex/config.toml
---
> CONTENT f90b7a97...45809 /Users/goos/.codex/config.toml
```

The entry listing is identical; only the content hash of `~/.codex/config.toml` moved (mtime 12:12), the foreign-writer pattern the dispatch named. The AFTER comparison below is against `dump-before2.txt`.

#### Design decisions taken in M3b (read from the SPEC text; the plan leaves the seam open)

- Opt-out in the scripts is inherited, not re-implemented: the call carries no environment edit, so `MOAI_SKIP_PLUGIN_INSTALL` reaches the verb, which is the single place that decides what counts as an opt-out (`1` or `true`). The message line before the call still prints under the opt-out; the verb then runs nothing.
- `install.sh`: a function `install_plugin` called after `verify_installation`; the call is `if ! "$TARGET_PATH" plugin install; then ... fi` (an `if` condition is exempt from `set -e`); on failure a warning plus the two manual commands.
- `install.ps1`: after `Verify-Installation`, `& $targetPath plugin install` inside `try`, `$LASTEXITCODE` checked for the warning, a `catch` that only warns. No `2>&1` on the call (under `$ErrorActionPreference = "Stop"` Windows PowerShell 5.1 turns redirected native stderr into a terminating error).
- `install.bat`: `"%TARGET_PATH%" plugin install` then `if errorlevel 1 ( echo ... )`. **A deviation from design.md section 3.6, which says the script "ends `exit /b 0` (line 192)":** line 192 is the `:show_help` branch; the main path ended `goto :eof`, which hands back the ERRORLEVEL at that moment, so a verb that exits 1 (an older release, the case the Windows CI job meets) would have failed `call install.bat` although every later line is an `echo`. The main path now ends `exit /b 0`, and `TestInstallScriptsPluginStepGuarded` pins "the first script end after the call is `exit /b 0`".
- Harness: one pass of the same script handles the six modes; the isolation cases judge the environment the product cases run in; `isolation-poisoned-pin-never-executed` runs the real verb and requires the recorder behind the pin to run zero times AND the stub claude to be called at least once (the positive control); `isolation-real-home-unchanged` judges the canary (3 entries, the recorder adds 2 when it can see it) and the protected real roots, after every product case.
- The two installer cases that stub the binary also require that the verb WAS called exactly once, and the old-binary case that the `Unknown command "plugin" for "moai".` text reached the installer's output: without that a script that never calls the verb passes "exit 0 and Installation complete!" vacuously.
- The static guard sweeps the two docs-site copies for byte identity as well as the three scripts, with mutant fixtures for each checker.
- `scripts/test-plugin-install-step.sh` prints `RESULT pass=<n> fail=<m>` (the PENDING status no longer exists).

#### Commits (tdd order: observed RED as two own commits, then GREEN)

| Step | SHA | Content |
|------|-----|---------|
| RED 1 | `f453a6875` | `TestInstallScriptsPluginStepGuarded` (checkers, mutant fixtures, real-script sweep) |
| RED 2 | `4326f5563` | harness: five `isolation-*` cases, three `verb-*` cases, three flags, the real verb-call assertion, the verb-was-called assertions |
| GREEN | `b67f4bee0` | `install.sh`, `install.ps1`, `install.bat`, the two docs-site copies, `test-install.yml` |
| REFACTOR | none | nothing simplifiable found after the mutant runs; no code change, no commit |

The two RED commits precede GREEN in the commit graph (verification-claim-integrity section 2.3).

#### RED (observed at the unchanged install scripts, before any GREEN edit)

Guard (`go test ./internal/cli -run '^TestInstallScriptsPluginStepGuarded$' -count=1 -v`, 1 top-level test, 2 subtests, BI-1 empty):

```
--- FAIL: TestInstallScriptsPluginStepGuarded (0.03s)
    --- PASS: TestInstallScriptsPluginStepGuarded/checkers-detect-mutants (0.00s)
    --- FAIL: TestInstallScriptsPluginStepGuarded/real-scripts (0.03s)
    plugin_install_guard_test.go:424: install.sh: no non-comment line runs "$TARGET_PATH" plugin install
    plugin_install_guard_test.go:424: install.ps1: no non-comment line runs & $targetPath plugin install
    plugin_install_guard_test.go:424: install.bat: no non-comment line runs "%TARGET_PATH%" plugin install
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.780s
```

Harness at the unchanged scripts (`sh scripts/test-plugin-install-step.sh bin/moai`, exit 1, `RESULT pass=9 fail=3`, `LEAK=0`):

```
FAIL installer-calls-verb-by-installed-path: plugin-calls=0(want 4); missing[claude plugin marketplace add modu-ai/moai-adk]; missing[claude plugin install moai@moai-adk]; missing[codex plugin marketplace add modu-ai/moai-adk]; missing[codex plugin add moai@moai-adk]
FAIL installer-set-e-guard: verb-calls=0(want 1)
FAIL installer-old-binary-unknown-verb: verb-calls=0(want 1); unknown-verb-text-not-seen
```

The five isolation cases and the three verb cases test behavior M3a already delivered, so they pass against the real binary at once; their observed failures are (a) the same harness run against a stand-in binary that predates the verb (`<scratchpad>/m3b/old-moai`, exit 1, `RESULT pass=5 fail=7`): `FAIL isolation-poisoned-pin-never-executed: stub-claude-never-called`, `FAIL verb-install-all-tools: exit-1; plugin-calls=0(want 4); ...`, `FAIL verb-no-tools-exit-0: exit-1; skip-lines=0(want 2)`, `FAIL verb-optout-zero-calls: exit-1`; and (b) the three new negative controls, which turn each isolation case red under its own breakage (E4).

#### GREEN and the final run (HEAD `b67f4bee0`, bin/moai judged)

```
$ go test ./internal/cli -run '^TestInstallScriptsPluginStepGuarded$' -count=1 -v
--- PASS: TestInstallScriptsPluginStepGuarded (0.01s)
    --- PASS: .../checkers-detect-mutants   --- PASS: .../real-scripts
$ sh scripts/test-plugin-install-step.sh bin/moai ; exit 0
scrub: enumerated and unset 34 names
PASS isolation-env-scrubbed / isolation-cwd-has-no-project / isolation-resolves-to-stubs / isolation-poisoned-pin-never-executed / isolation-real-home-unchanged
PASS verb-install-all-tools / verb-no-tools-exit-0 / verb-optout-zero-calls
PASS installer-calls-verb-by-installed-path / installer-optout / installer-set-e-guard / installer-old-binary-unknown-verb
RESULT pass=12 fail=0
LEAK=0 (real roots unchanged: PROTECTED-SET 228cd364e524698c045646242724e9a0b2a23797d148b5830a0de195f01b7d71 entries=190)
```

Negative controls at the final tree, each exit 0 with `LEAK=0` and `RESULT <label>: red set and green set are exactly the expected ones`: `--negative-control` (RED env-scrubbed with 33 names left, cwd, poisoned-pin with `recorder-executed(2)`, real-home with `canary-changed(entries 3 -> 5)`; green resolves-to-stubs), `--negative-control-cwd` (RED cwd, poisoned-pin; green the other three, canary 3 -> 3), `--typed-list-mutant` (RED env-scrubbed with 31 names left; green the other four, recorder runs 0), `--negative-control-install-dir` (four RED, each by `decoy-holds-moai` and `installed-path-not-under-case-dir`; five green), `--negative-control-decoy-missing` (two PASS, installer runs 0).

`go build ./...` exit 0 and `GOOS=windows GOARCH=amd64 go build ./...` exit 0 at `b67f4bee0`. M3a regression selectors plus the guard (`TestPluginInstallStep_*`, `TestPluginInstallCmd`, `TestInitPluginStep_*`, `TestPluginOptOutCallersEnumerated`, `TestInstallScriptsPluginStepGuarded`): all `--- PASS`, `ok ... 17.938s`. `golangci-lint run --timeout=5m ./internal/cli/` at v2.1.6: `0 issues.` YAML of `test-install.yml` parsed with `ruby -ryaml` (the `test-sh` step list and the push path list printed as intended).

#### AC matrix (PASS only where the AC's own command was run and the expected output observed)

| AC | Result | Command | Observed |
|----|--------|---------|----------|
| AC-018 (a) | PASS | `sh scripts/test-plugin-install-step.sh bin/moai` | the four `PASS installer-*` lines above, each with its install-directory assertions (go resolves to the stub, installed path under the case directory, no decoy root holds a moai) |
| AC-018 (b) | PASS | `bash -n install.sh` | exit 0 |
| AC-018 (c) | PASS (static) | `go test ./internal/cli -run '^TestInstallScriptsPluginStepGuarded$' -count=1 -v` | `--- PASS`; labelled static, the behavior of `install.ps1` and `install.bat` is not executed locally |
| AC-018 (d) | PASS | `cmp install.sh docs-site/static/install.sh`; `cmp install.ps1 docs-site/static/install.ps1` | exit 0, no output, both |
| AC-019 (a), (b) | PASS (M3a, re-run) | the M3a selector run above; `bin/moai plugin install --help` | `--- PASS: TestPluginInstallCmd`; exit 0 |
| AC-019 (c) | PASS | the normal harness run | `PASS verb-install-all-tools`, `PASS verb-no-tools-exit-0`, `PASS verb-optout-zero-calls` |
| AC-025 (a) | PASS | the normal harness run | the five `PASS isolation-*` lines, `scrub: enumerated and unset 34 names` (N at least 5), `LEAK=0 (real roots unchanged: PROTECTED-SET ... entries=190)` |
| AC-025 (b) | PASS | `... --negative-control bin/moai` | exit 0, the expected RED and green lines, `RESULT negative-control: ...` |
| AC-025 (c) | PASS | `... --negative-control-cwd bin/moai`; `... --typed-list-mutant bin/moai` | exit 0 both, the expected sets |
| AC-025 (d) | GAP (see G-M3b-1) | `sh scripts/protected-set-hash.sh` before and after | the two lines differ, only the `config.toml` content line moved; attributed, not proven |
| AC-025 (e) | PASS | the grep of AC-025 (e) and the control | `grep` exit 1, no line; `grep -c 'CODEX_HOME' scripts/test-plugin-install-step.sh` prints `3` |
| AC-025 (f) | PASS | `... --negative-control-install-dir bin/moai` | exit 0, four RED, five green, `RESULT negative-control-install-dir: ...`, `LEAK=0` |

#### Mutants (each applied to the GREEN tree, run, then restored with `git checkout -- <file>`; harness = `scripts/test-plugin-install-step.sh bin/moai`, guard = the Go static test)

Killed:
- a bare `moai plugin install` in `install.sh`: harness `installer-calls-verb-by-installed-path` (`plugin-calls=0(want 4)`), `installer-set-e-guard` and `installer-old-binary-unknown-verb` red; guard `call without the installed path`.
- the call unguarded (`"$TARGET_PATH" plugin install` as a bare statement): harness `installer-set-e-guard` and `installer-old-binary-unknown-verb` red (`installer-did-not-complete(rc=1)`); guard `unguarded call under set -e`.
- a call that fails the installer on an old binary (`... || exit 1`): harness the same two cases red; the guard stayed green on shape (it reads `||` as guarded), which is the static limit named in AC-018 (c): only the harness kills it.
- the opt-out stripped by the script (`env -u MOAI_SKIP_PLUGIN_INSTALL "$TARGET_PATH" ...`): harness `installer-optout` red (`plugin-calls-recorded(4)`); the guard cannot see it.
- two comment lines that mention the verb and no call: harness three installer cases red; guard `no non-comment line runs ...`.
- `docs-site/static/install.sh` differing by one byte: `cmp` exit 1 (`differ: char 44, line 2`); guard `docs-site/static/install.sh differs from the root install.sh`.
- inside the guard: checker mutants for sh (bare moai, unguarded, comment-only), ps1 (bare moai, no try, catch rethrows, catch exits) and bat (bare moai, `exit /b %errorlevel%`, `goto :eof` as the first end) are all reported as problems, the well-formed shape of each is not (`checkers-detect-mutants`, PASS).
- harness-side: a scrub disabled, a project pin alone, a typed-list scrub and an omitted `--install-dir` each turn exactly the expected cases red (E4 above).

Not killed or not run: a `PATH` that is not reset (it would let the case start the real `claude`, which this milestone must not do), a scrub that runs after the first case, a files-only hash (not rebuilt here: `protected-set-hash.sh` is M-R0 and unchanged), and `install.ps1` / `install.bat` behavior mutants (no pwsh and no Windows host; only the static checkers see them). A harness started below a `.moai` ancestor through `TMPDIR` was tried and is inert on this machine: BSD `mktemp -d` without a template ignores `TMPDIR` (`env TMPDIR=<dir under a .moai> mktemp -d` printed a `/var/folders/...` path), so the run passed all twelve cases; the ancestor check is demonstrated by `--negative-control-cwd`, where the project directory inside the scratch makes `isolation-cwd-has-no-project` red.

#### Protected-set bracket and BI-1

BEFORE (`dump-before2.txt`, taken at `f453a6875` after the resume): `PROTECTED-SET 5d8c23fdb87a8e15dcab32dbef5aad51741fa6db4affcb72ddd42625492b061c entries=190`. AFTER (`dump-after.txt`, after the last harness run and the last mutant): `PROTECTED-SET 228cd364e524698c045646242724e9a0b2a23797d148b5830a0de195f01b7d71 entries=190`. The two lines are not equal. Line-by-line:

```
$ diff dump-before2.txt dump-after.txt
186a187
> CONTENT 0d9321a9...eacff /Users/goos/.codex/config.toml
190d190
< CONTENT f90b7a97...45809 /Users/goos/.codex/config.toml
```

(The line moved position because the dump is sorted by hash; it is one line, the content hash of `~/.codex/config.toml`.) All 189 other entries, directories and content lines, are identical. Classification: foreign writer, not this milestone. `~/.codex/config.toml` is 178,828 bytes with mtime 13:10 (178,211 bytes at 12:12); it holds 953 `[projects.*]` trust stanzas written by other lanes' live-codex tests; the plugin and marketplace stanzas it holds are the pre-existing `moai-cowork`, `openai-*` and `claude-plugins-official` ones, none names `moai-adk`, `moai@moai-adk` or `t1435`; the only stanzas matching `t1435|PluginInstall|pluginemit|moai-adk` are the 12 `hooks.state` entries and the `[projects."/Users/goos/MoAI/moai-adk-go"]` stanza of the primary checkout (pre-existing) and one scratchpad stanza of another lane's `develop` worktree. Within every harness run of this milestone the before and after hashes of the run itself were equal (every run of the harness printed `LEAK=0`, each run lasting seconds), so nothing the harness started wrote under the protected roots; the value moved only between runs. The file content before the run was not kept, so "pure addition of unrelated stanzas" is shown by absence of this lane's names, not by a diff (G-M3b-1).

BI-1 printed nothing (grep exit 1) at the pre-flight, before every harness run, every mutant run, every go-test measurement and at the end.

#### Claim

M3b delivers REQ-018, the harness half of REQ-019 and REQ-025: the three install scripts call `plugin install` of the binary they installed by its installed path and finish successfully when the verb fails, does not exist or is opted out; the docs-site copies are byte-identical; the Unix CI job builds `moai` and runs the stub harness; the harness has the twelve default cases and the five negative-control flags, each control exiting 0 only on exactly its expected red set; `TestInstallScriptsPluginStepGuarded` pins the shapes of the three scripts statically.

#### Evidence

The commands and verbatim outputs are in the RED, final-run, AC-matrix, mutant and bracket sections above; the mutant and control runs were performed against `bin/moai` built at `0c0ebd673` from the Go sources, which M3b does not change.

#### Baseline-attribution

Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1435`, branch `WT-marketplace-core-plugin`, base HEAD `0c0ebd673`. The guard RED was measured at `f453a6875`, the harness RED at the working tree on top of it (committed as `4326f5563` afterwards, unchanged), the GREEN runs, mutants, controls, build, lint and the AC commands at `b67f4bee0`; all in this run. The judging build for the harness is `bin/moai`, built from this tree with `go build -o bin/moai ./cmd/moai` at `0c0ebd673` and invoked by path (verification-claim-integrity section 2.2); no Go non-test source changed after it (the commits since touch only a test file, scripts, workflows and the installers), so the build's behavior is that of the tree's HEAD. The other judging tools are installed binaries resolved through PATH: `go` go1.26.8, `golangci-lint` v2.1.6, `ruby` for the YAML parse.

#### Gaps

- G-M3b-1: the protected-set bracket is not equal (see above); attributed to a foreign writer by content, not proven by a diff of the file's content, which was not kept.
- G-M3b-2: `install.ps1` and `install.bat` were not executed (pwsh exists but is refused by the worktree guard and was not run; there is no Windows host). Their call, guard and exit shape are asserted statically only; `install.bat`'s `exit /b 0` behavior under cmd is unobserved. The Windows CI job (`install.bat` against the latest release, which lacks the verb until it ships) is the first real run.
- G-M3b-3: the edited `test-install.yml` was parsed (ruby) but not run: the new Set up Go, Build moai and harness steps on ubuntu-latest and macos-latest are unobserved on a runner.
- G-M3b-4: two commands were refused by the worktree guard (a `git show` combined with `grep -c` on carriage returns, and a pipe through `cut` followed by a `PIPESTATUS` echo); each was re-run as plain separate commands. Nothing was replaced by reading source.
- G-M3b-5: mutants not run: the `PATH` that is not reset (it would start the real `claude`), the scrub that runs late, the files-only hash; the `TMPDIR` variant of the `.moai` ancestor mutant is inert on BSD `mktemp` (observed) and was replaced by `--negative-control-cwd`.
- G-M3b-6: only `-run` selectors of `internal/cli` and one lint of `./internal/cli/` were run (AGENTS.md section 4); no `go test ./...`.
- G-M3b-7: design.md section 3.6's statement about `install.bat` ending `exit /b 0` at line 192 is wrong for the main path (see Design decisions); the SPEC text was not edited (not this agent's file).
- G-M3b-8: `isolation-real-home-unchanged` includes the real roots, so a foreign writer on a busy machine can turn it red in an otherwise clean run; the case prints `real-roots-changed` and the closing `LEAK=1` block prints the differing entries so it is attributable, but it is not self-healing.

#### Residual-risk

- The guidance wording of the scripts is a proposal; the harness asserts counts, the four vectors and two marker strings (`Installation complete!`, `Unknown command "plugin" for "moai"`), so a reword that drops either marker would turn a case red for a cosmetic reason.
- The informational line before the call is printed even when the opt-out is set; only the verb's tool calls are suppressed.
- On Windows the first `install.bat` run against a release without the verb prints the manual-install warning; that is intended (RK-14) but unobserved.
- The static guard reads shapes (`if`, `||`, `try`/`catch`, `exit /b 0`); a guard that is shaped right and behaves wrong in `install.ps1` or `install.bat` is not detectable locally.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Plan→run Kickoff (decision record, auto-semantics §10), written by the lane at HEAD 1f30f7798 on 2026-10-03:

decision record: decided_by=claude lane-9 orchestrator (card t1435) acting on the leader's disposition evidence_refs=.moai/reports/t1435/plan-audit-iter4-delta.md (PASS-WITH-DEBT 0.87, audited_sha 58ee0bdb2, plan-artifact hash 9052d51b01fc9b409adf25c720bc897f81254cf1b7dab30e676e7e754ac5e20b),leader cross-session message (approval to enter run with PASS-WITH-DEBT, 2026-10-03) ladder_path=plan→run Kickoff row; the §9.1 wording is PASS, so the PASS-WITH-DEBT entry rests on operator approval 10-03 (relayed by the leader)

- Plan-artifact hash unchanged since the verdict: `git diff 58ee0bdb2..HEAD -- spec.md plan.md acceptance.md design.md research.md` printed 0 lines (HEAD 1f30f7798; only progress.md changed). The plan phase records audit-ready: plan-audit-iter4-delta.md PASS-WITH-DEBT; no open blocker. The SPEC's plan artifacts are not touched by the decision record.
- Binding run-phase instructions carried from the verdict: BI-1 (env pre-flight at the start of every measurement, `unset` of any hit in the same compound invocation; at dispatch, 2026-10-03: `env | cut -d= -f1 | grep -E '^(CLAUDE_CODE_PLUGIN_|BASH_ENV$|ENV$|BASH_FUNC_|GOBIN$|GOPATH$|GOFLAGS$|GOENV$|CODEX_SQLITE_HOME$|XDG_)'` printed nothing in this lane); BI-2 (first run task: the decoy-directory existence assertion before the omit-flag negative control, abort without running the installer if a decoy is missing); BI-3 (discoverable script: pid-named poison, recording wrapper, local-path grep, (e) labelled static).

Mode evaluation (inputs: tier L, about 30 hand-authored files, 4 milestones plus the BI-2 first task, coding-heavy, one working tree, Go + shell + manifests):

| Mode | Selected | Reason |
|---|---|---|
| direct | no | not trivial |
| serial | yes | coding-heavy, one writer per tree, milestone-ordered dependencies (M1 → M4) |
| fanout | no | read-only research is done; the writers share one tree |
| sweep | no | not a uniform mechanical transform |

Decision: serial

Justification: one write-capable agent per milestone, sequentially, in this card worktree (one writer per tree). Deviation recorded: the run-phase workers are spawned as `Agent(general-purpose)` carrying the manager-develop role text, not as `manager-develop`, because a typed `manager-develop` spawn auto-isolates into its own L1 worktree and cannot write the card tree (observed on t1318 and recorded in t1434's progress.md §E.2; not re-measured here). Their commits carry `Authored-By-Agent: general-purpose`, so the ownership-transition lint reports the `draft → in-progress` flip as unattributed rather than as owned by manager-develop.
