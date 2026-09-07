# SPEC-CODEX-COVER-RESIDUAL-001 — Mutant ledger (card t519)

Every mutant below was applied to already-correct production source, observed, and reverted. No
mutant is present in any commit (REQ-CCR-007, REQ-CCR-008).

**Tree**: worktree `.claude/worktrees/t519`, branch `WT-codex-cover-residual`.
**Production-file integrity**: `internal/cli/codex_review_gate.go` SHA256
`9356669bdeb39f433897b7fc82c7e4cf036ca8ab24197301558b8c7bcedc3500` and `internal/cli/mcp_codex.go`
SHA256 `2a2ad2a9fd3a84566a372b01b3367223cd37fa0e71136f9543e2827e528ad8da` — captured before the first
mutation and re-observed identical after every revert.

**Method**: each mutant was applied with `perl -i`, the scoped test run under the env-scrubbed
compound invocation `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR
MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -run '<selector>' -v ./internal/cli/`, then the
file restored from a byte-exact backup under `.moai/state/verify/t519/` and the same command re-run.

## Summary

| Mutant | Target | Verdict |
|---|---|---|
| M1 | codex_review_gate.go:186 | RED (AC-CCR-001 only) |
| M1b | codex_review_gate.go:188 | RED (AC-CCR-001 + AC-CCR-002) |
| M2 | codex_review_gate.go:191 | RED (AC-CCR-003) |
| M2-cf | AC-CCR-003 fixture counterfactual | **did NOT fire** — boundary evidence, see below |
| M3a | codex_review_gate.go:195 | RED (AC-CCR-004) |
| M3b | codex_review_gate.go:196 | RED (AC-CCR-004) |
| M3-vac | codex_review_gate.go:196 | **did NOT fire** — vacuity confirmed by measurement |
| M4 | codex_review_gate.go:201 | RED (AC-CCR-005) |
| M5a | mcp_codex.go:702-704 | RED (AC-CCR-006, panic) |
| M5b | mcp_codex.go:703 | RED (AC-CCR-006) |
| M6 | codex_review_gate.go (trailing newline) | FIRED (AC-CCR-007 absence guard) |

Fired RED: 8. Did not fire: 2 (both deliberate boundary probes, neither counted as adoption
evidence). M6 is an absence guard and is counted as FIRED rather than RED — it has no test verdict,
only a `git status --short` observation.

---

## M1 — delete the malformed-stdin stderr diagnostic

- **Edit**: delete `codex_review_gate.go:186` (`_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "codex-review-gate: invalid stdin JSON (%v); emitting default output\n", err)`)
- **Command**: `-run 'TestRunCodexReviewGate_InvalidStdinFailsOpen|TestRunCodexReviewGate_EmptyStdinFailsOpen'`
- **RED (verbatim)**:

```
=== RUN   TestRunCodexReviewGate_InvalidStdinFailsOpen
    codex_review_gate_wiring_test.go:74: stderr must carry a codex-review-gate diagnostic, got ""
--- FAIL: TestRunCodexReviewGate_InvalidStdinFailsOpen (0.00s)
=== RUN   TestRunCodexReviewGate_EmptyStdinFailsOpen
--- PASS: TestRunCodexReviewGate_EmptyStdinFailsOpen (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.883s
```

AC-CCR-002 passing here is the predicted behaviour, not a miss: AC-CCR-002 asserts only a nil
`Execute()` error and an ALLOW on stdout, so the stderr deletion cannot reach it. This is precisely
why M1b exists as AC-CCR-002's separate adoption basis.

- **GREEN after revert (verbatim)**:

```
=== RUN   TestRunCodexReviewGate_InvalidStdinFailsOpen
--- PASS: TestRunCodexReviewGate_InvalidStdinFailsOpen (0.00s)
=== RUN   TestRunCodexReviewGate_EmptyStdinFailsOpen
--- PASS: TestRunCodexReviewGate_EmptyStdinFailsOpen (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.701s
```

- **`git status --short` after revert**: `?? internal/cli/codex_review_gate_wiring_test.go` — the
  new test file only; no production source listed.

## M1b — return the parse error instead of failing open

- **Edit**: `codex_review_gate.go:188` — `return emitHookOutput(cmd.OutOrStdout(), &hook.HookOutput{})` → `return err`
- **Command**: `-run 'TestRunCodexReviewGate_InvalidStdinFailsOpen|TestRunCodexReviewGate_EmptyStdinFailsOpen'`
- **RED (verbatim)**:

```
=== RUN   TestRunCodexReviewGate_InvalidStdinFailsOpen
    codex_review_gate_wiring_test.go:70: invalid stdin must not error (fail-open); got parse stdin: invalid character 'n' looking for beginning of object key string
--- FAIL: TestRunCodexReviewGate_InvalidStdinFailsOpen (0.00s)
=== RUN   TestRunCodexReviewGate_EmptyStdinFailsOpen
    codex_review_gate_wiring_test.go:85: empty stdin must not error (fail-open); got parse stdin: unexpected end of JSON input
--- FAIL: TestRunCodexReviewGate_EmptyStdinFailsOpen (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.866s
```

Both arms fire, confirming the plan-audit F3 prediction: empty stdin does produce a non-nil `err`
(`unexpected end of JSON input`), so M1b reaches AC-CCR-002 where M1 could not.

- **GREEN after revert (verbatim)**:

```
=== RUN   TestRunCodexReviewGate_InvalidStdinFailsOpen
--- PASS: TestRunCodexReviewGate_InvalidStdinFailsOpen (0.00s)
=== RUN   TestRunCodexReviewGate_EmptyStdinFailsOpen
--- PASS: TestRunCodexReviewGate_EmptyStdinFailsOpen (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.701s
```

- **`git status --short` after revert**: `?? internal/cli/codex_review_gate_wiring_test.go` only.

## M2 — force the gate enabled

- **Edit**: `codex_review_gate.go:191` — `enabled := readCodexReviewGateEnabled(projectDir)` → `enabled := true`
- **Command**: `-run 'TestRunCodexReviewGate_HappyPathAllow'`
- **RED (verbatim)**:

```
=== RUN   TestRunCodexReviewGate_HappyPathAllow
    codex_review_gate_wiring_test.go:109: codex must not be consulted when the gate is disabled
--- FAIL: TestRunCodexReviewGate_HappyPathAllow (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.856s
```

The RED is the `t.Fatal` guard inside `withCodexLookPath`, as AC-CCR-003 requires — not an
incidental failure elsewhere.

- **GREEN after revert (verbatim)**:

```
=== RUN   TestRunCodexReviewGate_HappyPathAllow
--- PASS: TestRunCodexReviewGate_HappyPathAllow (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.695s
```

## M2-cf — counterfactual: M2 with `withChangeDetector` removed (did NOT fire)

Not an adoption mutant. It measures the claim acceptance.md AC-CCR-003 makes in prose — that
`withChangeDetector(t, true)` is what makes M2 detectable — rather than taking that claim on trust.

- **Edit**: M2 still applied, PLUS the line `withChangeDetector(t, true)` removed from
  `TestRunCodexReviewGate_HappyPathAllow` (test-file edit, reverted from a byte-exact backup).
- **Observed (verbatim)**:

```
=== RUN   TestRunCodexReviewGate_HappyPathAllow
--- PASS: TestRunCodexReviewGate_HappyPathAllow (0.02s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.868s
```

**Finding**: the test PASSES under M2 once the detector swap is dropped — M2 becomes vacuous
exactly as predicted. The production detector runs `git status` against a non-git `t.TempDir()`,
returns false, and the mutated handler ALLOWs at step 3 before ever reaching `codexLookPath`. The
fixture line is therefore load-bearing and must not be removed as redundant. Both files were
restored before proceeding; `internal/cli/codex_review_gate.go` re-hashed to
`9356669bdeb39f433897b7fc82c7e4cf036ca8ab24197301558b8c7bcedc3500`.

## M3a — delete the handler-error stderr diagnostic

- **Edit**: delete `codex_review_gate.go:195` (`_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "codex-review-gate: error:", gateErr)`)
- **Command**: `-run 'TestRunCodexReviewGate_HandlerErrorFailsOpen'`
- **RED (verbatim)**:

```
=== RUN   TestRunCodexReviewGate_HandlerErrorFailsOpen
    codex_review_gate_wiring_test.go:152: stderr must carry the handler-error diagnostic, got ""
--- FAIL: TestRunCodexReviewGate_HandlerErrorFailsOpen (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.854s
```

## M3b — propagate the handler error out of Execute

- **Edit**: `codex_review_gate.go:196` — `return emitHookOutput(cmd.OutOrStdout(), &hook.HookOutput{})` → `return gateErr`
- **Command**: `-run 'TestRunCodexReviewGate_HandlerErrorFailsOpen'`
- **RED (verbatim)**:

```
=== RUN   TestRunCodexReviewGate_HandlerErrorFailsOpen
    codex_review_gate_wiring_test.go:148: handler error must not error out of Execute (fail-open); got fake: codex exited non-zero
--- FAIL: TestRunCodexReviewGate_HandlerErrorFailsOpen (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.897s
```

## M3-vac — emit `out` instead of the empty ALLOW (did NOT fire — vacuity confirmed)

Pre-recorded vacuous in spec.md §D.2 and acceptance.md §D. Run here NOT as adoption evidence but to
verify the pre-record by measurement, since an unverified vacuity claim is itself an unobserved
claim.

- **Edit**: `codex_review_gate.go:196` — `&hook.HookOutput{}` → `out`
- **Observed (verbatim)**:

```
=== RUN   TestRunCodexReviewGate_HandlerErrorFailsOpen
--- PASS: TestRunCodexReviewGate_HandlerErrorFailsOpen (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.856s
```

**Finding**: PASS confirms vacuity. `HandleCodexReviewGate` returns its empty `allow` value
alongside the error, so both forms serialize to `{}`. AC-CCR-004's adoption rests on M3a and M3b,
which both fired. This row maps the guard's boundary: the test does not — and cannot — discriminate
which empty-ALLOW value the error arm emits.

## M4 — replace the BLOCK with an ALLOW on the success path

- **Edit**: `codex_review_gate.go:201` — `emitHookOutput(cmd.OutOrStdout(), out)` → `emitHookOutput(cmd.OutOrStdout(), &hook.HookOutput{})`
- **Command**: `-run 'TestRunCodexReviewGate_BlockVerdictPropagates'`
- **RED (verbatim)**:

```
=== RUN   TestRunCodexReviewGate_BlockVerdictPropagates
    codex_review_gate_wiring_test.go:177: expected BLOCK to propagate through the RunE, got "{}\n"
    codex_review_gate_wiring_test.go:180: BLOCK must carry a non-empty reason, got "{}\n"
--- FAIL: TestRunCodexReviewGate_BlockVerdictPropagates (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.853s
```

- **GREEN after revert — all five M1 tests (verbatim)**:

```
=== RUN   TestRunCodexReviewGate_InvalidStdinFailsOpen
--- PASS: TestRunCodexReviewGate_InvalidStdinFailsOpen (0.00s)
=== RUN   TestRunCodexReviewGate_EmptyStdinFailsOpen
--- PASS: TestRunCodexReviewGate_EmptyStdinFailsOpen (0.00s)
=== RUN   TestRunCodexReviewGate_HappyPathAllow
--- PASS: TestRunCodexReviewGate_HappyPathAllow (0.00s)
=== RUN   TestRunCodexReviewGate_HandlerErrorFailsOpen
--- PASS: TestRunCodexReviewGate_HandlerErrorFailsOpen (0.00s)
=== RUN   TestRunCodexReviewGate_BlockVerdictPropagates
--- PASS: TestRunCodexReviewGate_BlockVerdictPropagates (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.719s
```

- **Integrity after revert**: `internal/cli/codex_review_gate.go` SHA256
  `9356669bdeb39f433897b7fc82c7e4cf036ca8ab24197301558b8c7bcedc3500` (identical to the pre-mutation
  capture); `git status --short` lists only the new test file.

## M5a — delete the pid nil guard (RED arrives as a panic)

Run in isolation (`-run 'TestCodexSessionHandlePid'`) per acceptance.md AC-CCR-010, because the
panic aborts the remaining tests in the same binary. Go emits the `--- FAIL:` line ahead of the
trace, so the criterion is satisfied as written; the trace is recorded alongside it so the capture
is not mistaken for truncated.

- **Edit**: delete `mcp_codex.go:702-704` (`if h == nil || h.conn == nil { return 0 }`)
- **RED (verbatim, head of output)**:

```
=== RUN   TestCodexSessionHandlePid
--- FAIL: TestCodexSessionHandlePid (0.00s)
panic: runtime error: invalid memory address or nil pointer dereference [recovered, repanicked]
[signal SIGSEGV: segmentation violation code=0x2 addr=0x0 pc=0x1013046f8]

goroutine 66 [running]:
testing.tRunner.func1.2({0x1049575e0, 0x104cd1100})
	/Users/goos/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.4.darwin-arm64/src/testing/testing.go:1974 +0x1a0
testing.tRunner.func1()
	/Users/goos/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.4.darwin-arm64/src/testing/testing.go:1977 +0x318
panic({0x1049575e0?, 0x104cd1100?})
	/Users/goos/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.4.darwin-arm64/src/runtime/panic.go:860 +0x12c
github.com/modu-ai/moai-adk/internal/cli.(*codexSessionHandle).pid(...)
	/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t519/internal/cli/mcp_codex.go:702
github.com/modu-ai/moai-adk/internal/cli.TestCodexSessionHandlePid(0x3968b138c908)
	/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t519/internal/cli/mcp_codex_test.go:677 +0x28
testing.tRunner(0x3968b138c908, 0x104b14438)
	/Users/goos/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.4.darwin-arm64/src/testing/testing.go:2036 +0xc4
created by testing.(*T).Run in goroutine 1
	/Users/goos/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.4.darwin-arm64/src/testing/testing.go:2101 +0x3a8
```

The trace names `mcp_codex.go:702` reached from `mcp_codex_test.go:677` — the typed-nil receiver
arm — which is exactly the short-circuit the disjunction guard provides.

## M5b — return -1 instead of 0 from the guard

- **Edit**: `mcp_codex.go:703` — `return 0` → `return -1`
- **RED (verbatim)**:

```
=== RUN   TestCodexSessionHandlePid
    mcp_codex_test.go:678: pid on a nil handle = -1, want 0
    mcp_codex_test.go:681: pid on a handle with no conn = -1, want 0
--- FAIL: TestCodexSessionHandlePid (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.850s
```

Both zero-arms mismatch, as predicted. The third arm (`fakeCodexConnPID`) is untouched by this
mutant and stays silent, which is correct — it does not pass through the guard.

- **GREEN after revert (verbatim)**:

```
=== RUN   TestCodexSessionHandlePid
--- PASS: TestCodexSessionHandlePid (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.682s
```

- **Integrity after revert**: `internal/cli/mcp_codex.go` SHA256
  `2a2ad2a9fd3a84566a372b01b3367223cd37fa0e71136f9543e2827e528ad8da` (identical to the pre-mutation
  capture); `git status --short` listed only ` M internal/cli/mcp_codex_test.go`.
