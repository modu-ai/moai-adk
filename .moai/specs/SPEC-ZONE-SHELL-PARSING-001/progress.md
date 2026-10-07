# SPEC-ZONE-SHELL-PARSING-001 — progress

Card t1574 · Tier M · branch WT-shell-parsing-guard · status: in-progress (run-phase closed 2026-10-08; sync-phase pending)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-07T23:00:22+09:00
note: Tier M artifact set (spec.md, plan.md, acceptance.md, progress.md) authored 2026-10-07 at tree b9ef003808da2ac1dfe42e5374d5bde4302f46b3; RED evidence ledger at acceptance.md §B. Plan-phase audit closed at iter4+gate (PASS 0.93, hash 5be3fdaab68d9637e55cff3a0c760202d9d328c9329da010dab1e70a94282a4a).

## §E.2 Run-phase Evidence

Run-phase: 2026-10-07 → 2026-10-08, manager-develop, cycle_type=tdd, worktree `.moai/worktrees/t1574`, branch `WT-shell-parsing-guard`. Evidence tree: M1 `b4d95502d` + M2 `b6ff5867a` + the M3 lint fix (this commit) — every command below ran in this run against this tree's content state.

### AC binary matrix

| AC | Status | Verification command | Actual output |
|----|--------|---------------------|---------------|
| AC-ZSP-001 | PASS | `go test ./internal/hook/ -run 'TestProtectedZoneShellParsingBypasses/(dash_dash_operand\|conditional_function_declaration\|command_env_prefix)' -count=1 -v -timeout 2m` | `--- PASS: …/dash_dash_operand`, `…/conditional_function_declaration`, `…/command_env_prefix` + `ok github.com/modu-ai/moai-adk/internal/hook 0.668s` |
| AC-ZSP-002 | PASS | `go test ./internal/hook/ -run 'TestProtectedZoneShellParsingBypasses/recursion_state_reset' -count=1 -v -timeout 3m` | `--- PASS: …/recursion_state_reset (0.01s)` + `ok … 0.745s`; the run carries the `category=loop-unbounded` deny WARN — termination AND fail-closed verdict |
| AC-ZSP-003 | PASS | RED-first at base (E8 below), then `go test ./internal/hook/ -run 'TestProtectedZoneShellParsingBypasses/executable_path_function_absorption' -count=1 -v -timeout 2m` | RED `decision="allow" reason="", want deny` (pre-fix) → GREEN `--- PASS: …/executable_path_function_absorption (0.00s)` |
| AC-ZSP-004 | PASS | `go test ./internal/hook/ -run TestProtectedZoneShellParsingMatrix -count=1 -v -timeout 10m` | `parsing-matrix sweep: 48 cells` + `ok … 1.137s`, exit 0; 48 `--- PASS` subtest lines counted; no `[no tests to run]` |
| AC-ZSP-005 | PASS | `go test ./internal/hook/ -run 'TestProtectedZone$' -count=1 -timeout 10m` | `ok … 4.999s` (post-fix; baseline at b9ef00380 was `ok … 4.515s`/4.622s — same family, unchanged verdicts); the four frozen subtests' assertion lines byte-identical (below) |
| AC-ZSP-006 | PASS | `go vet ./internal/hook/` · `golangci-lint run ./internal/hook/...` · `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -cover ./internal/hook/ -timeout 30m` (slot lease `hook-test-suite`) | vet exit 0; lint `0 issues.` exit 0; suite `ok … 486.577s coverage: 87.5% of statements` exit 0, zero `stack overflow`/`fatal error` tokens |

### E8 — K4 RED-first evidence (Invariant i; verbatim, captured BEFORE the K4 fix landed)

Command: `go test ./internal/hook/ -run 'TestProtectedZoneShellParsingBypasses/executable_path_function_absorption' -count=1 -v -timeout 2m` (tree b9ef00380 + the test-only addition).

```
=== RUN   TestProtectedZoneShellParsingBypasses
=== RUN   TestProtectedZoneShellParsingBypasses/executable_path_function_absorption
    protected_zone_shell_parsing_test.go:108: rm() { :; }; /bin/rm zone_dir/secret.md: decision="allow" reason="", want deny
--- FAIL: TestProtectedZoneShellParsingBypasses (0.01s)
    --- FAIL: TestProtectedZoneShellParsingBypasses/executable_path_function_absorption (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.805s
```

The other four classes were RED at base per acceptance.md §B EV-1/EV-2/EV-4 (dash-dash re-observed in this run's pre-flight: `rm -- -zone/secret.md: decision="allow" reason="", want deny`).

### Frozen-subtest preservation (D4)

The test file was untracked at base, so `git diff` has no base side; the evidence is the authoring history: the file entered version control at M1 `b4d95502d` through two purely-additive edits only (the fifth `t.Run` registration line + the appended `testShellBypassExecutablePathAbsorption` function) — no existing line was modified. Post-state assertion lines, byte-identical to the plan-phase EV-1/EV-2 texts:

```
35:	wantZoneDeny(t, "rm -- -zone/secret.md", d, r, harnessLearnerIdentity, "category", "probe_zone")
50:	wantZoneDeny(t, "false && rm() { :; }; rm zone_dir/secret.md", d, r, harnessLearnerIdentity, "category", "probe_zone")
67:		wantZoneDeny(t, cmd, d, r, harnessLearnerIdentity, "category", "probe_zone")
90:	if d == "" && r == "" {
108:	wantZoneDeny(t, "rm() { :; }; /bin/rm zone_dir/secret.md", d, r, harnessLearnerIdentity, "category", "probe_zone")
```

### Matrix unbounded cells — deny verdicts observable (AC-ZSP-004)

`go test ./internal/hook/ -run 'TestProtectedZoneShellParsingMatrix/(d6_cd_chain_budget|recursion_width_stress_d2|recursion_unbounded_loop|d9_absent_manifest_budget_deny)' -count=1 -v -timeout 5m`: four `category=loop-unbounded` deny WARN lines (one per cell) + four `--- PASS`. The d9 cell runs under `newZoneRoot(t, "", "")` — a fixture root with no zone manifest — and still denies (D9a).

### Scope

`git status --short` at M3: `M internal/hook/protected_zone_shell.go` (the M3 lint fix) + untracked `plan.md`/`acceptance.md` (plan-phase artifacts — landing owned by the lane/spec side, not staged by run-phase) + `progress.md` (this file). No other path touched. Commits: M1 `b4d95502d`, M2 `b6ff5867a`, M3 (this commit).

### Run-phase interpretation record (for the lane/leader)

1. **Walk-abort semantics.** REQ-ZSP-006's abort is implemented as: the GLOBAL visit budget (`zoneWalkerEntryBudget=10000`) aborts the walk immediately — every deeper entry returns; the recursion and cwds bounds set the unbounded flag WITHOUT halting the walk (their work is already bounded by the entry counter). Rationale: halting on those two bounds leaves a covered statement following a bound-tripping construct unjudged and flips the recorded category of the family-frozen cell `while false; do cd docs; done; rm zone_dir/secret` (guard_test :953, expects `probe_zone`, REQ-ZSP-008/D4) — the two requirements collide on that shape and the entry-budget-abort reading is the one that keeps the frozen family green while bounding all work.
2. **Plan M2 D6 cell divergence (report, not a silent deviation).** The plan's D6 wording — a cd chain followed by a zone-covered rm asserting the unbounded deny — is not co-satisfiable with the family freeze: that shape answers the covered candidate's more specific deny (candidate precedence, frozen by :953's class). Authored as two cells instead: the bare chain (`d6_cd_chain_budget`) asserting the set-budget unbounded deny — the REQ-ZSP-007 cell — and `d6_cd_chain_then_covered_rm` asserting the candidate deny, documenting the precedence.
3. **Read-only + unbounded widening (intended, D9a).** A command whose walk trips a bound with no mutating form now answers `loop-unbounded` deny (an incomplete walk may not answer allow) where the pre-fix code answered allow via the `!w.mutating` short-circuit. No family cell freezes the old allow (the full family run is green).
4. **K7 order.** Final verdict order: covered-candidate deny (precedence — the more specific verdict the family freezes) → unbounded deny (BEFORE the `!w.mutating` short-circuit, regardless of manifest state) → `!w.mutating` → invalid-manifest deny → absent-manifest COMPLETED-walk degrade.
5. **Lint first reading.** The first lint run printed 107 issues while another session's golangci-lint held the run lock; two clean serial re-runs after the ineffassign fix read `0 issues.` — the 107-issue reading was contention-polluted and is superseded by the serial observations.

### Gate round — turn-end codex gate findings folded (commit `123409baf`)

Both findings RED-first (tests authored and observed failing before the fixes), TDD cycle closed green, family `ok 2.738s`, vet exit 0, lint `0 issues.`, native + windows builds exit 0 at `123409baf`.

| Finding | RED evidence (verbatim, pre-fix, this run) | Fix | Post-fix |
|---------|-------------------------------------------|-----|----------|
| Wrapped-cd over-denial (P2): after the K3 strip, `env cd`/`nohup cd` were tracked as parent-shell directory changes | `env cd zone_dir; rm harmless: decision="deny" … path=zone_dir/harmless", want allow` (same for `nohup cd zone_dir`) | `zoneStripWrapperPrefix` reports whether it stripped; the cd branch fires for a BARE cd head only — a stripped head resolves as an external execution whose cd cannot move the parent shell | matrix cells `wrapped_cd_env_not_tracked` / `wrapped_cd_nohup_not_tracked` PASS (allow); the control `bare_cd_tracking_control` (`cd zone_dir; rm harmless` → deny) PASS — tracking stays live |
| Read-only fast path (P2): K7 moved the unbounded check ahead of the mutating short-circuit, and every command then loaded the manifest — loader-count probe: baseline 0, current 1, even for `echo hello` | `read-only command performed 1 zone loads, want 0` (TestProtectedZoneShellReadOnlyFastPath) | `!w.mutating && !w.unbounded` returns BEFORE `loadZone`; the K7 semantics unchanged — the unbounded denial still precedes every allow answer, regardless of manifest state | fast-path test PASS: `echo hello` → 0 loads; the unbounded read-only-tail cell still loads and denies |

Matrix: 51 cells (`parsing-matrix sweep: 51 cells`), all PASS. Final-tree full-suite re-measurement (slot lease `hook-test-suite`, tree `123409baf` content): `ok github.com/modu-ai/moai-adk/internal/hook 406.964s coverage: 87.5% of statements`, exit 0, zero `stack overflow`/`fatal error` tokens — the E3 coverage figure and the suite verdict now carry the same final tree.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-10-08T00:34:16+09:00
run_commit_sha: 123409baf
gate_round: folded 2 P2 gate findings RED-first (commit 123409baf)
run_status: complete
ac_pass_count: 6
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-performed (lane owns push; no pre-commit fetch owed in the worktree lane protocol)
l44_post_push_fetch: not-performed (push owned by the lane/leader)
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin_arm64: pass (`go build ./...` exit 0)
cross_platform_build.windows_amd64: pass (`GOOS=windows GOARCH=amd64 go build ./...` exit 0)
total_run_phase_files: 3
m1_to_mN_commit_strategy: M1 five fixes + spec.md draft→in-progress transition (b4d95502d) → M2 parsing-matrix sweep (b6ff5867a) → M3 lint fix + this evidence file (this commit)

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase (manager-docs)_

## §F Phase 4 Mode Selection

Input parameters: tier M; scope = 2-3 files (`internal/hook/protected_zone_shell.go` + parsing test file + optional matrix test file); domains = 1 (Go implementation, `internal/hook` only); file language mix = 100% Go; concurrency benefit = LOW (coding-heavy TDD, order-dependent milestones); agent-team prereqs = not requested (no operator `--team`).

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | multi-milestone implementation, not a trivial edit |
| serial | **YES** | coding-heavy Go work — Anthropic coding-task parallelism caveat; one manager-develop per milestone cycle |
| fanout | no | no independent multi-domain research; milestones are order-dependent (K4 RED-before-fix invariant, frozen-assertion protection) |
| sweep | no | semantic new-code work, not mechanical-uniform bulk; ~3 files far below the ~30-file bar |

Decision: serial

Justification: the run is one TDD implementation cycle over a single package whose milestones must execute in order — the K4 subtest must be observed RED before its fix lands (M1 step 2), and the four frozen subtests must stay byte-identical throughout. Coding-heavy work takes serial per the Anthropic finding; no fan-out surface exists in the plan. Kickoff gate: met in the autonomous form (iter4+gate PASS 0.93, hash 5be3fdaa unchanged, decision record at `.moai/reports/t1574/red-reproduction.md` § 게이트 재실행 결과 + Kickoff 결정), pre-authorized by the leader ruling (α) condition 3.
