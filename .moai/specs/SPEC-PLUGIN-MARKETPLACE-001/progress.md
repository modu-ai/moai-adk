# Progress — SPEC-PLUGIN-MARKETPLACE-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-03 (revision for plan-audit iteration 2; the revised artifacts are not yet audited)
tier: L
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md (progress.md not counted; decision-index.md is authored because `interview.decision_gate` is `on`)
budget: 24 requirements, 24 acceptance criteria (Tier L ceilings 25 and 25)
plan_base_sha: 7109e0900 (branch WT-marketplace-core-plugin, base develop); the revision was made on tree 3766cef05, which differs from the base in the SPEC artifacts only
run_start_sha: pending — set by the orchestrator to the commit that carries the final plan-phase revision of the artifacts
open_decisions: OD-1 to OD-13 in spec.md §5, mirrored as Q1 to Q13 in decision-index.md; verdict lines empty; every default-bound clause carries a `default pending OD-n` marker and spec.md §5 holds the marker table
inputs: .moai/reports/t1435/inputs/t1434-verdict.md (local-only, `.gitignore:235`); the committed carrier of the same verdict is .moai/specs/SPEC-PLUGIN-LOAD-SCOPE-001/progress.md section E.2; .moai/reports/t1435/plan-audit.md (iteration 1 verdict, FAIL 0.72)
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

### Audit-iteration-1 disposition

Every defect of `.moai/reports/t1435/plan-audit.md` is fixed; none is rejected, and none is skipped. "Evidence" names the re-run that
the fix rests on, in `research.md` (R-nn) or `acceptance.md` (L-nn).

| Defect | Disposition | Fixed where | Evidence |
|--------|-------------|-------------|----------|
| D1 REQ-010 contradicts AC-010, edge cases, plan M3 | fixed | REQ-013 (non-zero exit, timeout, invalid pin: one guidance block) and REQ-014 (absent tool: exactly one skip line, no block) replace the old REQ-010; AC-013 and AC-014 assert the same split; the edge-case list, plan M3 notes and design §3.2 say the same | the four locations re-read after the edit |
| D2 four decisions outside the OD table | fixed | OD-9 harness gating, OD-10 target profile, OD-11 payload language, OD-12 scope, with options, a default and evidence (spec §5, decision-index Q9 to Q12); REQ-010, REQ-012, REQ-005, REQ-006, REQ-011 follow them; the "`--llm` does not gate the step" sentence is gone from the plan. One more decision, OD-13 (the doctor's Codex read path), was added because the D10 measurement found a side effect | P-28 (R-09: `init.go:133`, `:192-195`, `:997-1003` read), P-29 (R-10), P-32 (R-08: 18 locale conditionals, all on `description`/`argument-hint` lines), P-33 (R-07: `--scope` flags), P-34 (R-05) |
| D3 defaults hard-coded without markers | fixed | 14 requirements carry a `default pending OD-n` marker (REQ-001, 003, 004, 005, 006, 007, 008, 010, 011, 012, 015, 018, 021, 022); the §5 marker table names each clause and what the other verdict changes; every affected criterion carries an `Alternate` line; the Definition of Done now requires a verdict or the line `default adopted` per row and a delta re-audit | `grep -c 'default pending OD-' spec.md` = 17 lines: the 14 requirement lines, the HISTORY row, the §2 preface and one §3 constraint |
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

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
