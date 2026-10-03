# t1487 verdict — develop red CI repair (run 37117806044, job 111187960009)

Branch `WT-develop-red-vocab-repair`, base local develop `924414a11`, fix commit `9c9a55d67`.

## Claim

The four CI failures are test-side staleness/isolation defects introduced by card t1399 (launcher rename); all four pass after test-only fixes. No production code, doc, or template changed; no guard weakened.

## Evidence

RED (local, env scrubbed, `go test -count=1 -run '<4 names>' ./internal/cli/`):

```
--- FAIL: TestAutoPickDocDoctrine (0.00s)
    todo_auto_pick_doc_test.go:184: read .claude/rules/moai/workflow/kanban-dispatch.md: open .../kanban-dispatch.md: no such file or directory
--- FAIL: TestAutoPickMirrorParity (0.00s)
    todo_auto_pick_doc_test.go:282: read .claude/rules/moai/workflow/kanban-dispatch.md: ... no such file or directory
--- FAIL: TestProductionStringLiteralsUseLeaderLaneVocabulary (0.12s)
    vocabulary_guard_test.go:240: stale allowlist entry: ../web/viewmodel_ops.go no longer contains "lead" — remove the entry
```

`TestBareMoaiPrintsBannerAndHelp` passed alone; reproduced by order (`-run 'TestNoPhantomBrain|TestBareMoaiPrintsBannerAndHelp'`):

```
--- FAIL: TestBareMoaiPrintsBannerAndHelp (0.00s)
    launcher_characterization_m1_test.go:138: the banner is missing from stdout: ""
```

Root causes:

1. AutoPick x2 — test added by t1448 `3af5605ab` hard-codes `kanban-dispatch.md`, `kanban-dispatch-detail.md`, `moai-kanban-foreman/SKILL.md`; t1399 renamed them to `factory-dispatch*.md` / `moai-factory-foreman`. Fix: path rename in the test only (live/mirror one-line drift still present and still checked).
2. Vocabulary guard — allowlist entry for `../web/viewmodel_ops.go` (`08864a090`, t1256); t1399 M9 `529a294d5` removed `legacyLeaderRole`. Fix: drop the entry, as the guard itself demands.
3. Banner — test added by t1399 M1 `1379abc9a`; cobra keeps the `help` flag value across `Execute` on the shared `rootCmd`, so an earlier `--help` test (e.g. help_linter_stale_test.go) leaves it true and the bare run prints help without `Run`. Fix: reset the help flag before the bare run.

GREEN:

```
go test -count=1 -run 'TestNoPhantomBrain|TestAutoPickDocDoctrine|TestAutoPickMirrorParity|TestBareMoaiPrintsBannerAndHelp|TestProductionStringLiteralsUseLeaderLaneVocabulary' ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	1.209s
go test -count=3 -run 'Help|Root|Phantom|Banner|AutoPick|LeaderLaneVocabulary|Launcher' ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	189.451s
go vet ./internal/cli/           -> clean
golangci-lint (v2.1.6) run ./internal/cli/ -> 0 issues.
```

## Baseline-attribution

Measured in this run on worktree tree at `924414a11` (RED) and `9c9a55d67` (GREEN), darwin, lane env unset in the same invocation.

## Gaps

- Whole-package `go test -timeout 30m ./internal/cli/` was attempted (pre-banner-fix build) and hit `panic: test timed out after 30m0s` at load avg ~20-28; not a verdict. In that partial run the other three tests did not fail; only the banner test failed (expected, fix not yet compiled in). Full-package verdict is left to CI.
- No goldens/testdata reference the changed literals (only test-file paths changed).

## Residual-risk

Other tests in `internal/cli` may share the same sticky-flag order dependency on `rootCmd`; only the failing one was fixed.
