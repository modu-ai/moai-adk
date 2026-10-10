---
id: SPEC-ZONE-SHELL-CD-TRACKING-001
title: "Acceptance criteria — cd destination tracking (post-`--` operand incl. hyphen-leading, in-project absolute), cd-class regression cells"
version: "0.2.0"
created: 2026-10-09
updated: 2026-10-09
author: manager-spec
tier: M
---

# SPEC-ZONE-SHELL-CD-TRACKING-001 — acceptance

## §A Scenario Conventions

Every criterion below is a Given-When-Then scenario, binary-testable: the Given names the tree state, the When names one runnable command, the Then names one observable output. Verification commands use the plain single-invocation form. The guard under test is `checkProtectedZoneShell` via the landed test harness (`newZoneRoot`/`zoneShippedDoc`/`zoneTestHandler` fixture family); a `deny` is `decision="deny"`, an `allow` is `decision="allow"` with an empty category.

## §B Evidence Ledger

Four-element rule (verification-completeness §2.1): a release-blocking RED cell carries the command, its verbatim stdout, its exit code, and the tree SHA — and RED must be red for the stated reason. Where a cited RED cannot be re-executed on the current tree, the criterion loses release-blocking eligibility, is classified regression-guard, and is NOT recorded as a pass (the undecidable disposition).

**EV-ZSCD-001 (inherited, DEMOTED — regression-guard-pending).** The predecessor's EV-6 probe: both cd shapes observed `decision="allow"` at tree `b9ef003808da2ac1dfe42e5374d5bde4302f46b3`, 2026-10-07 (verbatim record: `.moai/reports/t1584/red-reproduction.md`, inherited from `.moai/reports/t1574/`). The probe file (`internal/hook/zz_probe_cd_test.go`) was deleted after observation and never re-executed — this evidence CANNOT be re-executed on the current tree, so it is the demoted inheritance, never a pass. Its verbatim record:

```
=== RUN   TestZZProbeCdTracking
    zz_probe_cd_test.go:21: cmd="cd -- zone_dir && rm a.log" decision="allow" reason=""
    zz_probe_cd_test.go:21: cmd="cd /private/var/folders/.../TestZZProbeCdTracking1443033861/001/zone_dir && rm a.log" decision="allow" reason=""
--- PASS: TestZZProbeCdTracking (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.604s
```

_The inline quote elides the runner's temp path with `...`; the full verbatim output is the cited `.moai/reports/t1584/red-reproduction.md` (inherited from `.moai/reports/t1574/`)._

**EV-ZSCD-003 (RED-now — the `cd_track_` discriminator; four-element cell captured at plan phase; the carrier and the four elements added per plan-audit round 4 D1; the `cd_` inventory corrected in v0.2.0 per audit round 1 D1).** The landed sweep already carries five `cd_` cells — `bare_cd_tracking_control` (:127, deny), `wrapped_cd_env_not_tracked` (:125, allow), `wrapped_cd_nohup_not_tracked` (:126, allow), `d6_cd_chain_budget` (:166, deny-unbounded), `d6_cd_chain_then_covered_rm` (:167, deny) — the landed external-execution rule and t1574 set-budget cells, NOT cd-tracking coverage; a bare `cd_` selector counts them (two ALLOW included), so it is the wrong instrument. The NEW cd-tracking group carries the distinct `cd_track_` cell-name prefix.

- **Verbatim stdout** (raw, unmodified; 18,794 bytes, 170 lines; stderr 0 bytes). The four header lines inside the fence (lines 39-42) are run metadata; the verbatim comparison begins at line 43:

```text
# command: go test ./internal/hook/ -run '^TestProtectedZoneShellParsingMatrix$' -count=1 -v -timeout 5m
# tree-sha: 7d34e50924c93e0d2c099ef353f8bd9754233cc5 (observation HEAD; internal/, cmd/, go.mod and go.sum are identical to a604f89a3, see the attribution note after the fence)
# toolchain: go version go1.26.8 darwin/arm64
# exit: 0
=== RUN   TestProtectedZoneShellParsingMatrix
    protected_zone_shell_matrix_test.go:210: parsing-matrix sweep: 60 cells
=== RUN   TestProtectedZoneShellParsingMatrix/conditional_decl_paren_form
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
=== RUN   TestProtectedZoneShellParsingMatrix/conditional_decl_kw_form
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
=== RUN   TestProtectedZoneShellParsingMatrix/conditional_decl_if_branch
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
=== RUN   TestProtectedZoneShellParsingMatrix/if_else_all_branches_declare
=== RUN   TestProtectedZoneShellParsingMatrix/elif_all_branches_declare
=== RUN   TestProtectedZoneShellParsingMatrix/elif_partial_declare_stays_conditional
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret.md"
=== RUN   TestProtectedZoneShellParsingMatrix/if_one_branch_declares_control
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret.md"
=== RUN   TestProtectedZoneShellParsingMatrix/conditional_decl_for_body
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
=== RUN   TestProtectedZoneShellParsingMatrix/conditional_decl_while_body
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
=== RUN   TestProtectedZoneShellParsingMatrix/while_cond_decl_certain
=== RUN   TestProtectedZoneShellParsingMatrix/until_cond_decl_certain
=== RUN   TestProtectedZoneShellParsingMatrix/while_body_decl_conditional_same_loop
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
=== RUN   TestProtectedZoneShellParsingMatrix/conditional_decl_case_arm
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
=== RUN   TestProtectedZoneShellParsingMatrix/certain_shadow_control_paren
=== RUN   TestProtectedZoneShellParsingMatrix/certain_shadow_control_kw
=== RUN   TestProtectedZoneShellParsingMatrix/transitive_conditional_decl
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
=== RUN   TestProtectedZoneShellParsingMatrix/certain_chain_shadow_control
=== RUN   TestProtectedZoneShellParsingMatrix/wrapper_command
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
=== RUN   TestProtectedZoneShellParsingMatrix/wrapper_command_p
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
=== RUN   TestProtectedZoneShellParsingMatrix/wrapper_env
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
=== RUN   TestProtectedZoneShellParsingMatrix/wrapper_env_assignment
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
=== RUN   TestProtectedZoneShellParsingMatrix/wrapper_nohup
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
=== RUN   TestProtectedZoneShellParsingMatrix/wrapper_builtin
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
=== RUN   TestProtectedZoneShellParsingMatrix/wrapper_nested_env_command
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
=== RUN   TestProtectedZoneShellParsingMatrix/wrapper_command_cp
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
=== RUN   TestProtectedZoneShellParsingMatrix/wrapper_command_mv
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
=== RUN   TestProtectedZoneShellParsingMatrix/wrapper_nonzone_target
=== RUN   TestProtectedZoneShellParsingMatrix/wrapper_command_ls_control
=== RUN   TestProtectedZoneShellParsingMatrix/wrapper_bare_env_control
=== RUN   TestProtectedZoneShellParsingMatrix/wrapper_nohup_cat_control
=== RUN   TestProtectedZoneShellParsingMatrix/wrapped_cd_env_not_tracked
=== RUN   TestProtectedZoneShellParsingMatrix/wrapped_cd_nohup_not_tracked
=== RUN   TestProtectedZoneShellParsingMatrix/bare_cd_tracking_control
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/harmless"
=== RUN   TestProtectedZoneShellParsingMatrix/d1_shadow_command
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
=== RUN   TestProtectedZoneShellParsingMatrix/d1_shadow_command_p
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
=== RUN   TestProtectedZoneShellParsingMatrix/d1_shadow_env
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
=== RUN   TestProtectedZoneShellParsingMatrix/d1_shadow_nohup
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
=== RUN   TestProtectedZoneShellParsingMatrix/d1_shadow_builtin
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
=== RUN   TestProtectedZoneShellParsingMatrix/recursion_sibling_seed_d5
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
=== RUN   TestProtectedZoneShellParsingMatrix/recursion_mutual
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
=== RUN   TestProtectedZoneShellParsingMatrix/recursion_loop_nested
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
=== RUN   TestProtectedZoneShellParsingMatrix/recursion_width_stress_d2
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=loop-unbounded route=human next=return-blocker-report path=loop"
=== RUN   TestProtectedZoneShellParsingMatrix/recursion_unbounded_loop
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=loop-unbounded route=human next=return-blocker-report path=loop"
=== RUN   TestProtectedZoneShellParsingMatrix/dash_dash_first_arg
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=-zone/secret.md"
=== RUN   TestProtectedZoneShellParsingMatrix/dash_dash_after_option
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=-zone/secret.md"
=== RUN   TestProtectedZoneShellParsingMatrix/dash_dash_cp
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
=== RUN   TestProtectedZoneShellParsingMatrix/dash_dash_mv
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
=== RUN   TestProtectedZoneShellParsingMatrix/dash_dash_outside_control
=== RUN   TestProtectedZoneShellParsingMatrix/d6_cd_chain_budget
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=loop-unbounded route=human next=return-blocker-report path=loop"
=== RUN   TestProtectedZoneShellParsingMatrix/d6_cd_chain_then_covered_rm
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
=== RUN   TestProtectedZoneShellParsingMatrix/d9_absent_manifest_budget_deny
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=loop-unbounded route=human next=return-blocker-report path=loop"
=== RUN   TestProtectedZoneShellParsingMatrix/exec_path_no_shadow
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret.md"
=== RUN   TestProtectedZoneShellParsingMatrix/exec_path_with_shadow
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret.md"
=== RUN   TestProtectedZoneShellParsingMatrix/exec_path_relative
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
=== RUN   TestProtectedZoneShellParsingMatrix/exec_path_nonverb_control
=== RUN   TestProtectedZoneShellParsingMatrix/quoting_single_dashdash
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=-zone/x"
=== RUN   TestProtectedZoneShellParsingMatrix/quoting_double_wrapper
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
=== RUN   TestProtectedZoneShellParsingMatrix/quoting_escaped
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
=== RUN   TestProtectedZoneShellParsingMatrix/quoting_double_conditional
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
=== RUN   TestProtectedZoneShellParsingMatrix/quoting_single_exec_path
2026/10/10 02:19:37 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
--- PASS: TestProtectedZoneShellParsingMatrix (0.33s)
    --- PASS: TestProtectedZoneShellParsingMatrix/conditional_decl_paren_form (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/conditional_decl_kw_form (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/conditional_decl_if_branch (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/if_else_all_branches_declare (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/elif_all_branches_declare (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/elif_partial_declare_stays_conditional (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/if_one_branch_declares_control (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/conditional_decl_for_body (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/conditional_decl_while_body (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/while_cond_decl_certain (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/until_cond_decl_certain (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/while_body_decl_conditional_same_loop (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/conditional_decl_case_arm (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/certain_shadow_control_paren (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/certain_shadow_control_kw (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/transitive_conditional_decl (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/certain_chain_shadow_control (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/wrapper_command (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/wrapper_command_p (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/wrapper_env (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/wrapper_env_assignment (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/wrapper_nohup (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/wrapper_builtin (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/wrapper_nested_env_command (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/wrapper_command_cp (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/wrapper_command_mv (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/wrapper_nonzone_target (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/wrapper_command_ls_control (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/wrapper_bare_env_control (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/wrapper_nohup_cat_control (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/wrapped_cd_env_not_tracked (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/wrapped_cd_nohup_not_tracked (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/bare_cd_tracking_control (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/d1_shadow_command (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/d1_shadow_command_p (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/d1_shadow_env (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/d1_shadow_nohup (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/d1_shadow_builtin (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/recursion_sibling_seed_d5 (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/recursion_mutual (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/recursion_loop_nested (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/recursion_width_stress_d2 (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/recursion_unbounded_loop (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/dash_dash_first_arg (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/dash_dash_after_option (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/dash_dash_cp (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/dash_dash_mv (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/dash_dash_outside_control (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/d6_cd_chain_budget (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/d6_cd_chain_then_covered_rm (0.03s)
    --- PASS: TestProtectedZoneShellParsingMatrix/d9_absent_manifest_budget_deny (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/exec_path_no_shadow (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/exec_path_with_shadow (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/exec_path_relative (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/exec_path_nonverb_control (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/quoting_single_dashdash (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/quoting_double_wrapper (0.01s)
    --- PASS: TestProtectedZoneShellParsingMatrix/quoting_escaped (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/quoting_double_conditional (0.00s)
    --- PASS: TestProtectedZoneShellParsingMatrix/quoting_single_exec_path (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	1.337s
```

**Attribution note (round-7 repair of plan-audit round 6 D1).** The run was observed at 7d34e5092, the commit whose tree it read. The next commit, c92d8d81b (parent 7d34e5092), is the round-4 repair and changes SPEC documents only (`git diff --stat 7d34e5092 c92d8d81b`: five files under `.moai/specs/`, no code). The Go code the run exercised is identical at the observation HEAD and at the audited tree: `git diff --stat 7d34e5092 a604f89a3 -- internal cmd go.mod go.sum` prints nothing. Toolchain at the repair: `go version` = go1.26.8 darwin/arm64, as on line 41. Gap: the worktree status before the run is not re-observable now, so this note makes no clean-status claim.

- **Counts over the block above** (measured on the captured file): `=== RUN` lines 61 (the parent plus 60 cells; the parent's `t.Logf` line reads `parsing-matrix sweep: 60 cells`); `--- PASS` lines 61; `--- FAIL` lines 0; lines containing `cd_track_` **0**; lines containing `cd_` 10 (the RUN and PASS lines of the five landed cells).
- **Environment:** this session has `MOAI_KANBAN_ID`, `MOAI_KANBAN_BACKEND` and `MOAI_KANBAN_SETTINGS_INJECTED` set, and the run above used them. Cross-check (not the cell command): `unset MOAI_KANBAN_ID MOAI_KANBAN_BACKEND MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook/ -run '^TestProtectedZoneShellParsingMatrix$' -count=1 -v -timeout 5m` exits 0, and its stdout equals the block above after masking timestamps and durations.
- **Reading (red for the stated reason):** the cd-tracking group is absent. The parent ran its full catalogue (60 cells, no FAIL) and no executed cell carries the `cd_track_` prefix, so the criterion's non-empty-group claim fails because the group does not exist, not because the selector missed it.
- **Residual risk:** the discriminator shows the group is absent in this tree; it does not show that a future group is correct. M1 entry re-runs the command and re-pins the SHA; the counts, not the timing fields, are the comparison key.

**EV-ZSCD-002 (RED-now — CAPTURED at M1 step 1: RED record .moai/reports/t1584/red-reproduction-m1.md, GREEN record .moai/reports/t1584/green-m1.md).** The three-shape reproduction set run against the PRE-FIX guard at the M1-entry tree (pin recorded at capture time), each shape expected in the reset state (`decision="allow"`): `cd -- zone_dir && rm a.log`, `cd <fixture-root>/zone_dir && rm a.log`, `cd -- -zone && rm a.log` (against a `-zone/` fixture). Until this capture exists with its four elements, AC-ZSCD-001..003 stay regression-guard-pending and no green claim is recorded against them. The capture is the ONLY re-establishment path (spec D5; plan M1 step 1 owns it — the predecessor's D8 lesson).

**EV-ZSCD-004 (AC-ZSCD-005 — plan-phase green baseline; NOT a RED cell, see the classification note under §D).** Tree SHA `7d34e5092` (as EV-ZSCD-003). Single invocation, with `-v` so the executed count is on the record.

- **Command:** `go test ./internal/hook/ -run '^TestProtectedZone$' -count=1 -v -timeout 5m`
- **Exit code (separate field):** `0`
- **Verbatim stdout** (raw, unmodified; 29,715 bytes, 199 lines; stderr 0 bytes):

```text
=== RUN   TestProtectedZone
=== RUN   TestProtectedZone/Normalization
    protected_zone_test.go:106: swept=33
=== RUN   TestProtectedZone/FileTools
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Edit
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Edit
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Edit
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Edit
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Edit
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Edit
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Edit
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Edit
2026/10/10 02:21:02 WARN file access security check tool_name=Write decision=deny reason="Path traversal detected: file is outside project directory"
=== RUN   TestProtectedZone/FileTools/symlinks
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:02 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=secret.md"
=== NAME  TestProtectedZone/FileTools
    protected_zone_guard_test.go:214: swept=26
=== RUN   TestProtectedZone/ShellMutation
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/patch.diff"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_moai route=human next=return-blocker-report path=.moai"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_glob route=human next=return-blocker-report path=tests"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_logs route=human next=return-blocker-report path=.moai/logs/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_logs route=human next=return-blocker-report path=.moai/logs/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_logs route=human next=return-blocker-report path=.moai/logs/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_logs route=human next=return-blocker-report path=.moai/logs/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_logs route=human next=return-blocker-report path=.moai/logs/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_logs route=human next=return-blocker-report path=.moai/logs/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_logs route=human next=return-blocker-report path=.moai/logs/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_logs route=human next=return-blocker-report path=.moai/logs/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_logs route=human next=return-blocker-report path=.moai/logs/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_logs route=human next=return-blocker-report path=.moai/logs/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_logs route=human next=return-blocker-report path=.moai/logs/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_logs route=human next=return-blocker-report path=.moai/logs/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_logs route=human next=return-blocker-report path=.moai/logs/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_sp route=human next=return-blocker-report path=zone dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_deep route=human next=return-blocker-report path=a/a/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_deep route=human next=return-blocker-report path=a/a/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=loop-unbounded route=human next=return-blocker-report path=loop"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.log"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner manifest=invalid route=human next=return-blocker-report path=protected-zone.yaml"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:02 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_deep route=human next=return-blocker-report path=a/a/secret"
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret"
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/sentinel.txt"
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/sentinel.txt"
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir"
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_glob route=human next=return-blocker-report path=Tests"
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/guard.go"
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/guard.go"
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/a.md"
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/x"
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret.md"
    protected_zone_guard_test.go:1143: swept=101
=== RUN   TestProtectedZone/ShellQuoting
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/q1.log"
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/q2.log"
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/q3.log"
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=lnk/dir/q4.log"
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=lnk/dir/q5.log"
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/q6.log"
2026/10/10 02:21:03 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/q7.log"
    protected_zone_guard_test.go:1299: swept=9
=== RUN   TestProtectedZone/ManifestStates
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_HOOK_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
    protected_zone_guard_test.go:336: swept=15
=== RUN   TestProtectedZone/NonRegression
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_HOOK_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_RULE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_INSTRUCTION_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_AGENT_VIOLATION agent_id="" tool_name=Edit
    protected_zone_guard_test.go:394: swept=19
=== RUN   TestProtectedZone/DenyReason
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
    protected_zone_guard_test.go:449: swept=9
=== RUN   TestProtectedZone/NoManifestReadForOthers
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
    protected_zone_guard_test.go:487: swept=18
=== RUN   TestProtectedZone/AuditRow
=== RUN   TestProtectedZone/AuditRow/Deny
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
=== RUN   TestProtectedZone/AuditRow/Absent
=== RUN   TestProtectedZone/AuditRow/Invalid
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
=== RUN   TestProtectedZone/AuditRow/AppendFailure
2026/10/10 02:21:03 WARN protected zone audit append failed error="mkdir /private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestProtectedZoneAuditRowAppendFailure516343670/001/.moai/logs: not a directory"
2026/10/10 02:21:03 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
=== RUN   TestProtectedZone/BaselineCovered
    protected_zone_guard_test.go:707: swept=21
=== RUN   TestProtectedZone/Liveness
=== RUN   TestProtectedZone/Liveness/MatcherGroup
=== RUN   TestProtectedZone/Liveness/ManifestParse
=== RUN   TestProtectedZone/Liveness/DeadEntries
    protected_zone_guard_test.go:1222: resolved=36 skipped=11
=== RUN   TestProtectedZone/Liveness/HandlerDenies
2026/10/10 02:21:04 WARN harness frozen zone violation sentinel=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION agent_id="" tool_name=Write
--- PASS: TestProtectedZone (1.73s)
    --- PASS: TestProtectedZone/Normalization (0.00s)
    --- PASS: TestProtectedZone/FileTools (0.47s)
        --- PASS: TestProtectedZone/FileTools/symlinks (0.01s)
    --- PASS: TestProtectedZone/ShellMutation (0.14s)
    --- PASS: TestProtectedZone/ShellQuoting (0.01s)
    --- PASS: TestProtectedZone/ManifestStates (0.17s)
    --- PASS: TestProtectedZone/NonRegression (0.28s)
    --- PASS: TestProtectedZone/DenyReason (0.00s)
    --- PASS: TestProtectedZone/NoManifestReadForOthers (0.26s)
    --- PASS: TestProtectedZone/AuditRow (0.10s)
        --- PASS: TestProtectedZone/AuditRow/Deny (0.03s)
        --- PASS: TestProtectedZone/AuditRow/Absent (0.03s)
        --- PASS: TestProtectedZone/AuditRow/Invalid (0.00s)
        --- PASS: TestProtectedZone/AuditRow/AppendFailure (0.03s)
    --- PASS: TestProtectedZone/BaselineCovered (0.00s)
    --- PASS: TestProtectedZone/Liveness (0.29s)
        --- PASS: TestProtectedZone/Liveness/MatcherGroup (0.00s)
        --- PASS: TestProtectedZone/Liveness/ManifestParse (0.00s)
        --- PASS: TestProtectedZone/Liveness/DeadEntries (0.29s)
        --- PASS: TestProtectedZone/Liveness/HandlerDenies (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	2.290s
```

- **Counts:** `=== RUN` 21; `--- PASS` 21; `--- FAIL` 0.
- **Environment cross-check (not the cell command):** `unset MOAI_KANBAN_ID MOAI_KANBAN_BACKEND MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook/ -run '^TestProtectedZone$' -count=1 -v -timeout 5m` exits 0; its stdout equals the block above after masking timestamps and durations, except one WARN log line whose random `t.TempDir` suffix differs.
- **Second command of AC-ZSCD-005:** `go test ./internal/hook/ -run '^TestProtectedZoneShellParsingMatrix$' -count=1 -v -timeout 5m` is the same invocation as EV-ZSCD-003, whose run gives exit 0, 61 PASS, 0 FAIL.
- **Reading:** both families are green at the pre-implementation tree with a non-zero executed count; a preservation claim has no red state here, so this is a baseline, not a RED cell.

**EV-ZSCD-005 (selector behaviour — measured to correct the empty-sweep claim in §D.1 AC-ZSCD-004; not a RED cell).** Tree SHA `7d34e5092`. The anchored selector is used because the spec linter's `VacuousTestAssertion` rule flags an unanchored `-run` prefix, which also selects longer test names; the anchored form names no test either.

- **Command:** `go test ./internal/hook/ -run '^cd_track_$' -count=1 -v -timeout 5m`
- **Exit code:** `0`
- **Verbatim stdout** (111 bytes, 3 lines; stderr 0 bytes):

```text
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	1.227s [no tests to run]
```

- **Reading:** the selector matches no test, so no cell executes (zero `=== RUN` lines) and the exit code is 0 anyway; a `cd_track_` result proves a sweep only with the executed-count obligation in §D.1 AC-ZSCD-004.

**EV-ZSCD-006 (AC-ZSCD-006 — plan-phase component baselines; NOT a RED cell, see the classification note under §D).** Tree SHA `7d34e5092`. Scoped components only; the hook-suite component is a Gap.

- **Command A:** `go vet ./internal/hook/` — exit `0`; stdout 0 bytes; stderr 0 bytes.
- **Command B:** `golangci-lint run ./internal/hook/...` — tool `golangci-lint v2.1.6` (`golangci-lint --version`); exit `0`; stderr 0 bytes; verbatim stdout (10 bytes):

```text
0 issues.
```

- **Gap (not measured at plan phase):** `go test ./internal/hook/ -timeout 30m`, the hook-suite component. It runs for minutes and the lane-local rule requires a slot lease, which plan §C item 4 places at M1 entry and M3; its plan-phase state is unknown, and the M1-entry observation is its first baseline.
- **Environment:** the three `MOAI_KANBAN_*` names of EV-ZSCD-003 are set in this session; neither command executes tests, so no scrubbed run of these two commands was measured.

## §D AC Matrix (two-cell discipline: RED-now + green path)

| AC | Claim | RED-now | Green path | Classification |
|----|-------|---------|------------|----------------|
| AC-ZSCD-001 | The three-shape cd reproduction set is captured verbatim against the pre-fix guard, then flips to deny | EV-ZSCD-002 (to be captured at M1 step 1; the EV-ZSCD-001 inheritance is §2.1-demoted) | M1 fixes flip all three cells to deny; verbatim post-fix output recorded | release-blocking once captured; regression-guard-pending until then |
| AC-ZSCD-002 | `cd -- -zone && rm a.log` asserts DENY against a `-zone/` fixture | captured at M1 step 1 (pre-fix: allow — the reset) | M1 K3 fix flips it; permanent cell in the M2 `cd_track_` group | release-blocking once captured; regression-guard-pending until then |
| AC-ZSCD-003 | The in-project absolute-destination cell asserts the destination is tracked (protected deletion denied) | captured at M1 step 1 (pre-fix: allow — the reset) | M1 K2 fix flips it; permanent cell in the M2 `cd_track_` group | release-blocking once captured; regression-guard-pending until then |
| AC-ZSCD-004 | The cd-TRACKING matrix group exists in `TestProtectedZoneShellParsingMatrix`, non-empty, `-v` per-cell output, empty-list-fails intact | EV-ZSCD-003 (the `cd_track_` discriminator, four-element cell captured at plan phase): the pre-work run prints zero `cd_track_` cells among 60 executed cells — the five landed `cd_` cells acknowledged, the three NEW shapes unswept (SHA re-pinned at M1 entry) | LANDED at M2, commit ec1eec308: six `cd_track_` cells, captured in .moai/reports/t1584/green-m2-matrix-raw.txt (RUN 67, PASS 67, FAIL 0). The parent test's empty-cell-list `t.Fatal` keeps guarding the catalogue whenever that parent runs | release-blocking |
| AC-ZSCD-005 | The landed `TestProtectedZone` family and the landed parsing matrix pass unchanged | no RED state exists (preserved behavior; the Given presupposes the landed fixes); plan-phase green baselines EV-ZSCD-004 (family) and EV-ZSCD-003 (matrix) | green at each milestone exit; `git diff` over the landed test files shows no assertion-line edit | regression-guard (preserved behavior; reclassified from release-blocking, see the note below) |
| AC-ZSCD-006 | Scoped verification batch green: hook package suite, `go vet ./internal/hook/`, `golangci-lint run ./internal/hook/...` | no RED state exists (quality gate; the Given presupposes completed M1–M2); plan-phase component baselines EV-ZSCD-006 (vet, lint); hook-suite component a Gap until M1 entry | RUN at M3, HEAD ec1eec308: `go test ./internal/hook/` ok 298.839s; `go vet` exit 0; `golangci-lint` 0 issues at HEAD and at baseline bf9532147; captured in .moai/reports/t1584/green-m3-test-raw.txt, green-m3-lint-head.txt, green-m3-lint-baseline.txt, green-m3-matrix-raw.txt, green-m3.md | regression-guard (quality gate; reclassified from release-blocking, see the note below) |

**Classification note (plan-audit round 4, D1b).** AC-ZSCD-005 and AC-ZSCD-006 are regression guards, not release gates. A release-blocking criterion needs a RED-now cell, a red state on the pre-implementation tree, and neither can have one: AC-ZSCD-005's Given is "tree state with this SPEC's fixes landed" and AC-ZSCD-006's Given is "the completed M1–M2 work", so both are post-implementation checks. Their plan-phase observations (EV-ZSCD-004, EV-ZSCD-003 for the matrix, EV-ZSCD-006) are green baselines, not RED cells; both stay in the definition of done as verbatim M3 evidence. Restoring release-blocking status requires a RED-now cell that the Given can produce on the pre-implementation tree.

## §D.1 Given-When-Then Scenarios

**AC-ZSCD-001** — RED re-establishment and flip.
- **Given** the pre-fix guard at the M1-entry tree (SHA pinned at capture), when the three-shape reproduction set runs with `-v` and per-shape verdict output, then each shape prints `decision="allow"` (the reset state — red for the STATED reason: the `:447` two-operand reset and the `:451` hyphen/absolute resets, never a selector miss), the commands' verbatim stdout and exit codes are recorded in acceptance.md §B, and after the M1 fixes the same three shapes print `decision="deny"`.

**AC-ZSCD-002** — the D7 hyphen-leading operand cell denies.
- **Given** a fixture root whose zone manifest covers `-zone/`, when `cd -- -zone && rm a.log` is judged, then the verdict is `deny` (category `probe_zone`) — post-fix; pre-fix the same command answers `allow` (captured at M1 step 1). A mutant that restores the hyphen-prefix reset for post-`--` operands fails this cell.

**AC-ZSCD-003** — the in-project absolute destination is tracked.
- **Given** a fixture root whose zone manifest covers `zone_dir/`, when `cd <fixture-root>/zone_dir && rm a.log` is judged (the absolute destination built from the fixture root), then the verdict is `deny` — the destination is tracked under its root-relative form; pre-fix the same command answers `allow`. A mutant that keeps the `zoneIsAbs` reset for in-project absolutes fails this cell; an outside-root absolute destination (`cd /tmp/outside && rm a.log`) stays `allow` and is NOT a matrix cell (REQ-ZSCD-003).

**AC-ZSCD-004** — the cd-TRACKING regression group is swept, never silent.
- **Given** the landed sweep runner `TestProtectedZoneShellParsingMatrix` (protected_zone_shell_matrix_test.go:205) — which already carries five `cd_` cells (`bare_cd_tracking_control`, the two wrapped-cd ALLOW cells, the two d6 cd-chain cells), when the suite runs with `-v -count=1` at the pre-work tree, then zero cells match the `cd_track_` prefix (EV-ZSCD-003 — the three NEW shapes are unswept); and after M2 lands the `cd_track_` group (post-`--` operand across `rm`/`cp`/`mv`, the `cd -- -zone` deny cell, the absolute-destination cell), every zone-covered `cd_track_` cell prints its own `--- PASS` line with a deny verdict, and the empty-cell-list `t.Fatal` (`protected_zone_shell_matrix_test.go:208`, inside the parent test `TestProtectedZoneShellParsingMatrix` at :205) still fails a swept-empty catalogue whenever that parent runs. A selector that names no test never enters the parent: `-run '^cd_track_$'` exits 0 with `testing: warning: no tests to run` (EV-ZSCD-005), so a `cd_track_` result counts only with the executed-count obligation — at least one `cd_track_` cell executed and its `--- PASS` lines counted.

**AC-ZSCD-005** — preserved behavior.
- **Given** tree state with this SPEC's fixes landed, when `go test ./internal/hook/ -run '^TestProtectedZone$' -count=1 -v -timeout 5m` and `go test ./internal/hook/ -run '^TestProtectedZoneShellParsingMatrix$' -count=1 -v -timeout 5m` run (`-v`, so the executed count is on the record: a non-verbose `ok` can be a zero-cell pass), then both pass with a non-zero executed count, and `git diff` over the landed test files shows no assertion-line edit on pre-existing cells — no already-covered form changed verdict (REQ-ZSCD-004); the landed cd-chain budget cell and the bare-cd-only rule are untouched.

**AC-ZSCD-006** — scoped verification batch.
- **Given** the completed M1–M2 work, when the env-scrubbed batch runs (slot lease → `go test ./internal/hook/ -timeout 30m` → `go vet ./internal/hook/` → `golangci-lint run ./internal/hook/...`), then all three exit 0 with no NEW findings vs the M1-entry baseline, and the outputs are recorded verbatim in progress.md §E.2.

## §D.2 Edge Cases

- `cd -- dir1 dir2` (two operands after the separator): bash rejects the cd, the shell stays put — the reset is CORRECT and stays; the matrix may pin it as a preserve observation, never as an allow control for a hyphen-leading shape.
- `cd -` (OLDPWD): stays reset (the guard cannot know OLDPWD); not a matrix cell.
- `cd -- -L dir`-shaped option-before-separator forms: the hyphen reset stays in force BEFORE the `--`; only post-`--` operands become ordinary operands.
- `cd <root>/../<root>/zone_dir`-shaped absolutes: cleaned inside the root → tracked (REQ-ZSCD-002 lexical rule); cleaned outside → reset.
- `cd <outside-root>/deep && rm ../../<project>/zone_dir/a.log` (the reaching-back spelling, named in v0.2.0 per audit D2): answers allow today — the accepted under-match of REQ-ZSCD-003; never a matrix cell; hardening it is the decision-index Q2 operator pin's alternative, not this card's scope.
- Dynamic destination (`cd $d && rm a.log`): keeps the `?dynamic` under-match (parent §C.6) — unchanged by this SPEC.

## §D.3 Quality Gates

TRUST 5: Tested (AC-ZSCD-001..004, hook suite green); Readable/Unified (English round-style comments matching file density, gofmt/golangci-lint clean — AC-ZSCD-006); Secured (fail-closed direction, the D7 disposition — AC-ZSCD-002); Trackable (Conventional Commits referencing this SPEC ID, card t1584 in commit messages).

## §E Traceability

| AC | REQ | Evidence |
|----|-----|----------|
| AC-ZSCD-001 | REQ-ZSCD-001, REQ-ZSCD-002, REQ-ZSCD-003 | EV-ZSCD-001 (demoted), EV-ZSCD-002 (M1 step 1) |
| AC-ZSCD-002 | REQ-ZSCD-001, REQ-ZSCD-005 | EV-ZSCD-002 shape 3 + the M2 cell |
| AC-ZSCD-003 | REQ-ZSCD-002, REQ-ZSCD-005 | EV-ZSCD-002 shape 2 + the M2 cell |
| AC-ZSCD-004 | REQ-ZSCD-005 | EV-ZSCD-003 + the M2 runner run |
| AC-ZSCD-005 | REQ-ZSCD-004 | EV-ZSCD-004 (plan-phase green baseline) + EV-ZSCD-003 (matrix) + the M3 re-run |
| AC-ZSCD-006 | REQ-ZSCD-001, REQ-ZSCD-002, REQ-ZSCD-003, REQ-ZSCD-004, REQ-ZSCD-005 | EV-ZSCD-006 (plan-phase component baselines) + the M3 batch |

## §F Indirect Verification

- REQ-ZSCD-003 (outside-root reset kept) is verified indirectly: AC-ZSCD-003's scenario asserts the outside-root control stays allow OUTSIDE the matrix, and AC-ZSCD-005's family re-run pins that no landed allow outside-root behavior flipped.
- REQ-ZSCD-004 (preserved behavior) is verified indirectly by AC-ZSCD-005's no-assertion-edit diff plus the family/matrix green.

## §G Closure Gates

- Definition of Done: AC-ZSCD-001..006 all GREEN with verbatim evidence; the three RED captures exist with four elements each (command, stdout, exit, tree SHA); no FOUNDER-blocked decision remains open in decision-index.md; progress.md §E.2/§E.3 populated; the D8b sweep conclusion (plan §B — fix-here 0) re-stated in the completion report.
- Forward-looking check: the NEW cd-tracking cells carry the `cd_track_` name prefix, so a future selector audit can count the cd-TRACKING class without parsing the catalogue — a bare `cd_` selector is the wrong instrument on this tree (it counts the five landed `cd_` cells, two of them ALLOW).
