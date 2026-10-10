# SPEC-USER-SETTINGS-PROTECT-001 — Acceptance Criteria

This file is the verification layer. Each criterion is written as Given-When-Then, names the command that verifies it, and carries two cells where the criterion is release-blocking: a RED-now cell (the criterion observed on the pre-implementation tree, with the command and its verbatim output) and a green-path cell (what flips it and what the passing output becomes). A criterion whose RED-now cell is pending is marked as pending and listed in plan.md G-14. The requirement layer (GEARS) lives in spec.md §2 and is not restated here.

Pre-implementation tree for every RED-now cell: `2aab5f797` on `WT-3-2-0`. Evidence identifiers (E-n, G-n, Q-n) refer to plan.md §A.2, plan.md §A.4, and decision-index.md.

Class key: release-blocking (the run cannot close without it), regression-guard (green today; must stay green), gate (a precondition the run checks before it changes anything).

### AC-001 — The default init path keeps the user's allow, ask, deny, and unmodelled keys (REQ-001)

- Given: a USER-scope settings file holding `permissions.defaultMode` "plan", `allow`, `ask`, `deny`, `additionalDirectories`, and a sibling `env` key.
- When: `moai init` runs with the default (empty) autonomy tier, which resolves to semi-auto (E-6).
- Then: `allow`, `ask`, `deny`, `additionalDirectories`, and `env` are unchanged. Only `defaultMode` may change, and its disposition follows the Q1 verdict (AC-003).
- Contract anchor: SPEC-INIT-WIZARD-REPAIR-001 §4 (decision-index.md Q8, DECIDED). The M1 preservation test that §4 requires is absent from the tree (plan.md E-23), so the run adds it under this criterion.
- Verifying command: `go -C <worktree> test -count=1 ./internal/core/project/ ./internal/config/toolpolicy/` (assertions authored in the run; selectors recorded in progress.md at M2).
- Class: release-blocking.
- RED-now (observed): probe 2 on this tree (plan.md E-4). The AFTER body is `{"permissions": {"defaultMode": "acceptEdits"}, "env": {"A": "1"}}`. The `allow` and `deny` lists are gone. The probe's own exit is 0 because it only logs. The red is the missing lists, which the assertion above would detect.
- Green path: after M2 the writer splices only `defaultMode` and keeps the lists and unmodelled keys. Probe 2 AFTER keeps `allow` and `deny`.

### AC-002 — The permissions region is byte-identical when the defaultMode already matches (REQ-002; card judgement 1)

- Given: a USER-scope settings file whose `defaultMode` already equals the resolved tier default, with the lists written in a non-canonical layout.
- When: `moai init` runs.
- Then: the bytes of the `permissions` region are identical before and after the run. Any diff of the region is empty.
- Verifying command: `go -C <worktree> test -count=1 ./internal/core/project/` (byte-equality assertion authored in the run). Operator replay on fixture files: `diff <(sed -n '/"permissions"/,/^  }/p' before.json) <(sed -n '/"permissions"/,/^  }/p' after.json)` prints nothing.
- Class: release-blocking.
- RED-now: pending (plan.md G-14). Code reading predicts red: `renderPermissionsObject` re-serializes the region in canonical layout (plan.md E-7), so a non-canonical layout changes bytes even when no value changes. Not observed in plan phase.
- Green path: after M2 the writer returns without rewriting the region when the resolved values match the file.

### AC-003 — An existing defaultMode that differs from the tier default follows the Q1 verdict (REQ-003) — BLOCKED

- Given: a USER-scope settings file with `defaultMode` "plan" while the resolved tier default is "acceptEdits".
- When: `moai init` runs.
- Then: the defaultMode is handled as decision Q1 records: either overwritten with the tier default or kept.
- Verifying command: `go -C <worktree> test -count=1 ./internal/core/project/` (the Q1 case, authored after the verdict).
- Class: release-blocking after the verdict. Status: blocked until Q1 carries an operator verdict (REQ-003 says the run does not start before then).
- RED-now (observed on this tree): plan.md E-3 and E-4 show the current disposition is overwrite (`plan` becomes `acceptEdits`). If Q1 resolves to keep, this criterion flips.
- Green path: the disposition the Q1 verdict records, with the test naming that verdict.

### AC-004 — The settings template ships no permissions.defaultMode unless Q2 says so (REQ-004)

- Given: the settings template at `internal/template/templates/.claude/settings.json.tmpl`.
- When: the number of `"defaultMode"` keys in the template is counted.
- Then: the count is 0 while decision Q2 is open.
- Verifying command: `grep -c '"defaultMode"' internal/template/templates/.claude/settings.json.tmpl` prints `0`.
- Class: regression-guard.
- RED-now (observed, green today): the command printed `0` with exit status 1 (plan.md E-12). The template carries `permissions` at lines 436 and 557 and no defaultMode.
- Green path: unchanged unless Q2 resolves to ship a defaultMode. If it does, this criterion is replaced by the Q2 verdict's test.

### AC-005 — The PROJECT-scope policy path keeps user-added project allow entries unless Q5 says otherwise (REQ-005) — BLOCKED

- Given: a tool-policy document is present and the project settings file holds a user-added `permissions.allow` entry that the document does not list.
- When: `project.ApplyAutonomyTierBundle` runs the full-bundle path with the document.
- Then: the user-added entry is kept (merge disposition), or regenerated away when Q5 records the regenerate disposition.
- Verifying command: `go -C <worktree> test -count=1 ./internal/core/project/` (the policy-path case, authored after Q5).
- Class: release-blocking after the verdict. Status: blocked on Q5.
- RED-now: pending (plan.md G-12). Code reading predicts red under the current code: `internal/config/toolpolicy/tier_render.go:76-81` sets the PROJECT block's `Allow` from the document. Not observed in plan phase.
- Green path: the disposition Q5 records, with the test naming that verdict.

### AC-006 — Every package that reaches a home-resolving function sandboxes MOAI_HOME in its TestMain (REQ-006)

- Given: the six packages of plan.md §A.4 Basis items 8 to 13: homestate, escalation, factory, factorymsg, web, and contract/receipt.
- When: each package's TestMain is checked for the sandbox install.
- Then: each `main_test.go` contains an `EnvHome` or `MOAI_HOME` sandbox install, and `internal/contract/receipt/main_test.go` exists with a TestMain.
- Verifying command: `grep -L -E 'EnvHome|MOAI_HOME' internal/homestate/main_test.go internal/escalation/main_test.go internal/factory/main_test.go internal/factorymsg/main_test.go internal/web/main_test.go` must print nothing after M6. Plus `ls internal/contract/receipt/main_test.go` must succeed.
- Class: release-blocking.
- RED-now (observed on this tree): the command printed five names, exit status 0: `internal/escalation/main_test.go`, `internal/factory/main_test.go`, `internal/homestate/main_test.go`, `internal/factorymsg/main_test.go`, `internal/web/main_test.go`. `ls internal/contract/receipt/main_test.go` printed `No such file or directory` (plan.md E-9, and the run of this session).
- Green path: after M6 the grep prints nothing and the receipt TestMain file exists.

### AC-007 — No test run writes under the operator's real home; the ~/.moai/run entry count does not increase (REQ-007; card judgement 2)

- Given: a listing of the operator's home taken by the operator before one test run (P-4), and a marker file created before the run.
- When: `go test -count=1 ./...` runs on the post-fix tree, under a throwaway account or a copy of the home (AP-2).
- Then: no path under `~/.moai`, `~/.claude`, `~/.codex`, or `~/.agents` is created or modified after the marker, and the entry count under `~/.moai/run` is unchanged.
- Verifying command: `find "$HOME/.moai" "$HOME/.claude" "$HOME/.codex" "$HOME/.agents" -newer <marker> 2>/dev/null | wc -l` must print `0`. And `find "$HOME/.moai/run" -mindepth 1 -maxdepth 1 | wc -l` before and after must print the same number.
- Class: release-blocking.
- RED-now: pending (plan.md G-8 and G-10). Not observable in plan phase, because the baseline is outside the worktree and an unsandboxed run would write the operator's home.
- Green path: after M6, the `-newer` count is `0` and the run-directory count is unchanged.

### AC-008 — Under the sandbox, RunProjectDir and the store resolvers resolve under the temporary MOAI_HOME root (REQ-008)

- Given: a test binary whose TestMain installs the MOAI_HOME sandbox.
- When: the existing store-resolver tests run, and a RunProjectDir assertion runs in homestate.
- Then: every returned path begins with the sandbox root, and none begins with the operator's real `~/.moai`.
- Verifying command: `go -C <worktree> test -count=1 -run '^(TestStoreDirUsesQueueProjectKey|TestStoreDirMatchesEscalation)$' ./internal/escalation/ ./internal/contract/receipt/`, plus the RunProjectDir assertion authored in homestate (name recorded in progress.md at M6).
- Class: release-blocking.
- RED-now: the two named tests exist (plan.md E-21). The homestate half is pending: 47 of 57 `internal/homestate/*_test.go` files contain neither `EnvHome` nor `MOAI_HOME` (plan.md E-9, `grep -L -E 'EnvHome|MOAI_HOME' internal/homestate/*_test.go | wc -l` printed `47`). Whether a homestate test resolves outside the sandbox is observed by the RunProjectDir assertion, which is pending.
- Green path: after M6 the RunProjectDir assertion passes and the two named tests keep passing.

### AC-009 — The review-gate live Codex test sets CODEX_HOME to a temporary root (REQ-009; card item b)

- Given: the live review-gate test run with `codex` on PATH and `MOAI_SKIP_LIVE_CODEX` unset, under a sandboxed home.
- When: `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey` runs (plan.md E-10, line 35).
- Then: CODEX_HOME points at a temporary root, and no file under the operator's `~/.codex` changes (AC-007 manifest).
- Verifying command: `grep -n 'CODEX_HOME' internal/cli/codex_review_gate_live_test.go` must print at least one `t.Setenv` line after M5.
- Class: release-blocking.
- RED-now (observed on this tree): `grep -n 'CODEX_HOME' internal/cli/codex_review_gate_live_test.go` printed nothing and the echo printed `codex-home-matches-exit=1` (plan.md E-10). The file has no CODEX_HOME reference.
- Green path: after M5 the grep prints a `t.Setenv` line for CODEX_HOME, and the audit-fixture pattern (plan.md E-10, lines 72 and 78) is followed.

### AC-010 — A missing sandbox fails the run with a named guard finding (REQ-010)

- Given: a test package whose TestMain omits the MOAI_HOME sandbox install (a mutant built with `go test -overlay`, as in the probes, so the repository is not edited).
- When: the guard test runs.
- Then: the run fails with a named finding that names the package, and the same guard turns green when the mutant is removed.
- Verifying command: `go -C <worktree> test -count=1 ./internal/testhome/` on the tree, and the same command with `-overlay` applied to a mutant package that omits the sandbox install (test name recorded in progress.md at M6).
- Class: release-blocking.
- RED-now: pending (plan.md G-14). The only TestMain-level guard name observed on this tree is `TestMainSandboxesProfileLeaseEnv` (plan.md E-21), which covers the cli package only. A repository-wide guard was not observed.
- Green path: the mutant run fails with the named finding, and the tree run passes.

### AC-011 — clean --home scans the run and db categories under the Q3 rule (REQ-011; card item e)

- Given: a sandboxed home that holds run/ and db/ entries older than the retention window, with one entry referenced by a live record and one carved-out entry.
- When: `moai clean --home --force` runs.
- Then: entries that the Q3 rule makes deletable are deleted; the live-referenced entry and the carved-out entry are kept.
- Verifying command: `go -C <worktree> test -count=1 -run '^TestCleanHome' ./internal/cli/`, with run and db cases added at M4 (existing names listed in plan.md E-21).
- Class: release-blocking.
- RED-now (observed on this tree): the category scan has projects (lines 236 and 249), debug (286), releases (293 and 458), logs (306), and backups (329). Neither `run` nor `db` appears (plan.md E-11). The run and db assertion is pending until its case is authored.
- Green path: after M4 the run and db categories appear in the scan under the Q3 rule, and the case passes.

### AC-012 — Dry-run stays the default and --force deletes only allowlisted categories (REQ-012)

- Given: a sandboxed home with deletable entries in the allowlisted categories and carved-out segments.
- When: `moai clean --home` runs without `--force`, then with it.
- Then: the dry run mutates nothing. The forced run deletes only allowlisted categories and keeps the carved-out segments.
- Verifying command: `go -C <worktree> test -count=1 -run '^(TestCleanHome_DryRunMutatesNothing|TestCleanHome_ForceDeletesOnlyAllowlistedCategories|TestCleanHomeCarveOut_ForcePreservesCarvedSegments)$' ./internal/cli/`.
- Class: regression-guard.
- RED-now: pending (plan.md G-14). The three test names exist (plan.md E-21). The run records their current result on the pre-implementation tree before any change.
- Green path: all three pass after the change, unmodified.

### AC-013 — The ordering landings are ancestors of the run base before run starts (REQ-013; gate)

- Given: the run base tip and the three landing SHAs the leader names for t1619, t1578, and t1591.
- When: each SHA is tested for ancestry against the run base.
- Then: each test exits 0 before any run-phase change begins.
- Verifying command: `git merge-base --is-ancestor <landing-sha> <run-base-tip>; echo "exit=$?"` for each of the three SHAs.
- Class: gate.
- RED-now (observed on this tree, with representative commits from plan.md E-14): `git merge-base --is-ancestor 8108eb256 HEAD` printed `t1619 ancestor-of-HEAD exit=1`. The same test for `e73a7cbf5` (t1578) and `a372a984c` (t1591) printed exit=1. The gate is red today.
- Green path: after the leader names the landing SHAs and they land, each test exits 0.

## Edge cases

- EC-1 — A settings file that does not exist yet: init creates it with the tier defaultMode and no lists (current behaviour). The run keeps this path and asserts it in the AC-001 test.
- EC-2 — A settings file with `permissions` absent or empty: no lists are invented.
- EC-3 — A settings file that is not valid JSON: the writer fails without rewriting the file (the current error path is kept).
- EC-4 — A list containing a duplicate entry: preserved as written; no deduplication is introduced.
- EC-5 — A HOME path that contains spaces or is a symbolic link: the sandbox root is resolved the same way the real home would be.
- EC-6 — A test package that never reaches a home-resolving function: no sandbox is required beyond the guard's own rule (REQ-010 applies to reaching packages only).

## Quality gate criteria

- QG-1: every release-blocking criterion has a RED-now cell that is observed or explicitly pending with its gap id.
- QG-2: every criterion names its verifying command.
- QG-3: no criterion depends on the operator's home directory except AC-007, which runs on a throwaway account.
- QG-4: no criterion names a test that does not exist, except the run-phase names recorded in progress.md when they are authored.
- QG-5: the count of criteria is 13, within the Tier M ceiling of 16.

## Definition of Done

- All release-blocking criteria (AC-001, AC-002, AC-003 after Q1, AC-005 after Q5, AC-006 to AC-011) are green on the post-fix tree, with verbatim output recorded.
- Regression guards (AC-004, AC-012) are green on the post-fix tree.
- The gate (AC-013) is green before run begins and again before sync.
- The operator-home measurement (AC-007) is recorded on a throwaway account.
- Q1, Q2, Q3, and Q5 carry operator verdicts in decision-index.md. Q4 and Q6 carry run-phase evidence.
- The t1594 scope (S-7) is either closed with its card text or recorded as deferred by the leader.
