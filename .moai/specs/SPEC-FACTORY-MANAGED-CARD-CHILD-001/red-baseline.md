# red-baseline.md — SPEC-FACTORY-MANAGED-CARD-CHILD-001 (card t1440)

Run-phase M1 record. Everything below was observed in this run, on the card worktree `.claude/worktrees/t1440`, branch `WT-codex-card-managed-path`, base commit `e1f790d7e` plus the uncommitted M1 files (inert seam declarations + the two new test files). Each `go test` ran under the lane environment scrub in ONE compound invocation:

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test ./internal/cli -run '<selector>' -count=1 -v
```

Output lines of the long form were cut to 220 columns; `WARN config sections directory not found` noise lines were dropped. Nothing else was edited.

## D3 — inert declarations first, compile proven before any "green on baseline" claim

M1 introduced the new seams as inert declarations before the tests: `managedCodexCardLaunchFunc` (declared in `internal/cli/codex_launcher.go`), and in the new `internal/cli/managed_operator_input.go` the source variable `managedLaneOperatorSource`, the creation counter `managedOperatorPumpsCreated`, a stub `defaultManagedCodexCardLaunch` (returns nil), and a stub pump (`newManagedOperatorPump`, `attach` that reports end of input at once). The loop does not use any of them yet.

```
$ go build ./internal/cli
build_exit=0
$ go vet ./internal/cli          (after the two new test files were added)
vet_exit=0
```

Only after those two exit 0 results did the AC-CC-002 test run on the baseline behavior (below).

## AC-CC-002 is GREEN on baseline behavior (invariant guard), arm 1 of 2

Controlled environment (the test clears the whole process environment and sets a fixed list), literal expected `Path`, `Args` (with an `AGENTS.local.md` subcase so the `-c developer_instructions` pair is present), `Dir`, ordered `Env` slice, and stdio pointer identity, plus zero reads of the operator-input source and zero pumps created.

```
$ ... go test ./internal/cli -run '^TestManagedCardChildSwitchOffKeepsDirectDoor$' -count=3 -v
--- PASS: TestManagedCardChildSwitchOffKeepsDirectDoor (24.30s)
    --- PASS: .../unset (4.11s)   --- PASS: .../empty (7.70s)   --- PASS: .../zero (5.23s)   --- PASS: .../yes (7.26s)
--- PASS: TestManagedCardChildSwitchOffKeepsDirectDoor (25.33s)   (4 subtests PASS)
--- PASS: TestManagedCardChildSwitchOffKeepsDirectDoor (22.47s)   (4 subtests PASS)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	72.946s
```

A first calibration run showed one literal of mine was wrong (`MOAI_AUTONOMY_TIER` value is `fully-autonomous`, hyphen, not underscore); the literal was corrected from the config constant's value, not copied from the failing cmd's whole env.

## D2 — RED of AC-CC-002 by VALUE-ONLY mutations (arm 2 of 2), each reverted

The mutations change one value and keep the key set and shape, so a key-set oracle or an oracle derived from the changed helper would stay green:

| Mutation (single edit in `internal/cli/codex_launcher.go`) | Observed |
|---|---|
| stamp value in `codexCardLaunchEnv`: `MOAI_KANBAN_BACKEND=gpt` becomes `...=gpt-MUTANT` | `--- FAIL: TestManagedCardChildSwitchOffKeepsDirectDoor/unset`, failure at `managed_card_child_test.go:283` (the Env literal comparison, both cards) |
| inherited value inside `codexChildEnv`: each inherited entry gets a `-MUTANT` suffix on its value | `--- FAIL: .../unset`, failure at the same Env comparison line |
| arg value: the `-C` operand `wt` becomes `wt + "-MUTANT"` in `launchCodexCardSession` | `--- FAIL: .../unset`, failure at `managed_card_child_test.go:262` (the Args literal comparison, both cards) |

After each, the edit was reverted by hand; `git diff --stat` then showed `internal/cli/codex_launcher.go | 7 +++++++` (the seam declaration only).

## RED arm of the new behavior tests (before M2/M3), `-run '^TestManagedCardChild'`

```
=== RUN   TestManagedCardChildLaneLoopUsesManagedOwner/switch_1
    managed_card_child_test.go:195: lane seam calls=0 plain seam calls=0 direct door calls=2, want 2/0/0
=== RUN   TestManagedCardChildLaneLoopUsesManagedOwner/switch_TRUE_padded
    managed_card_child_test.go:195: lane seam calls=0 plain seam calls=0 direct door calls=2, want 2/0/0
--- FAIL: TestManagedCardChildLaneLoopUsesManagedOwner (7.96s)
--- PASS: TestManagedCardChildSwitchOffKeepsDirectDoor (21.34s)   (4 subtests PASS)
=== RUN   TestManagedCardChildLaunchShape/no_dash_C_and_pair_and_dir
    managed_card_child_test.go:356: lane seam calls=0, want 1
    --- FAIL: .../no_dash_C_and_pair_and_dir
    --- PASS: .../oversize_instruction_skips_owner (1.70s)
    --- PASS: .../owner_refuses_dash_C (0.00s)
--- FAIL: TestManagedCardChildEnvCarriesIdentity: managed_card_child_test.go:411: lane seam calls=0, want 1
--- FAIL: TestManagedCardChildKeepsLauncherClaim: managed_card_child_test.go:451: claim observed at 0 owner launches, want 2
--- FAIL: TestManagedCardChildSecondCardRebinds
    managed_card_child_test.go:520: roster lanes = [], want exactly one lane endpoint          (sequential_cards_rebind)
    managed_card_child_test.go:536: owner launches = 0, want 2 (both cards attempted)          (start_failure_leaves_no_pending_row)
    --- PASS: .../positive_control_pending_row_visible (0.48s)
--- FAIL: TestManagedCardChildDeliversInboxThroughLoop: managed_card_child_test.go:585: the lane endpoint never bound within the watchdog
--- FAIL: TestManagedCardChildSessionEndContinuesLoop
    managed_card_child_test.go:654: owner launches = 0, want 2 (the loop must not stop at the first error)   (owner_error_logged_and_loop_continues)
    --- PASS: .../same_line_form_as_direct_door (4.54s)
--- FAIL: TestManagedCardChildOperatorInputReachesNextSession (exit_exit, quit_quit, exit_quit)
    managed_card_child_test.go:696: owner launches = 0, want 2
FAIL	github.com/modu-ai/moai-adk/internal/cli	106.305s
```

Right-reason reading: every FAIL is the loop still going to the direct door (`direct door calls=2`, owner stub / default owner never reached), which is exactly the missing M2 wiring. The PASSes are by design: `SwitchOffKeepsDirectDoor` (invariant guard), `oversize_instruction_skips_owner` (the size check sits before the branch point, so it holds on both sides — a regression guard, mutation-checked in M2), `owner_refuses_dash_C` (premise pin: the owner rejects `-C`), `positive_control_pending_row_visible` (positive control), `same_line_form_as_direct_door` (the direct door's own stderr form).

## RED arm of the pump unit test (before M3), `-run '^TestManagedOperatorInputPump'`

```
    managed_operator_input_test.go:71: Read = "", EOF; want "a\n", nil          (single_line_per_read)
    managed_operator_input_test.go:85: Read = "", EOF; want "a\n", nil          (multi_line_chunk_after_exit)
    managed_operator_input_test.go:98: Read = "", EOF; want "/quit\n", nil      (quit_token)
    managed_operator_input_test.go:136: session A received 0 lines ending in "", want 11 ending in /exit   (buffer_saturation)
    --- PASS: .../close_unblocks_read (0.05s)
    managed_operator_input_test.go:172: Read = "", EOF; want "x\n", nil          (source_eof)
    managed_operator_input_test.go:185: Read = "", EOF; want "got\n", nil        (fatal_end_residual)
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.184s
```

`close_unblocks_read` passes vacuously against the inert stub (its Read already returns end of input); its mutation check (Close that does not wake a blocked Read) is recorded in progress.md §E.2.

## Gaps

- The RED run was against the tree with inert declarations (not against a tree without them); AC-CC-002's "baseline" arm is therefore the unchanged direct-door code plus inert declarations the loop never calls.
- Windows: the fake App Server tests skip on Windows by design; the cross build is the evidence there.
