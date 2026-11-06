auditor-model: claude-opus-5-5

---
verdict: FAIL
audited_sha: 76b6dad2666ca62989beec395b3b6ca3f6e72a9b
---

# t2137 sync-audit verdict (iteration 1) — summary

Full report: `.moai/reports/t2137/sync-audit-iter1.md`

- Score (substance): 0.93 — Functionality 95 PASS / Security 95 PASS / Craft 90 PASS / Consistency 92 PASS
- Result: FAIL on a process gate only. The code fix has no defect.

## Why FAIL

F1 [High, blocking, process] `.moai/config/sections/workflow.yaml:38` `codex: required`. `moai verify audit-plan` reports `"cross_model_required": true` and `"enforced_required": [ "codex" ]`. The dispatch forbids codex for this card, so no codex receipt exists and the receipt guard refuses a PASS without one. Required fix: no code change. The operator either approves a codex review (then re-audit and cite the receipt) or accepts the card on this Claude-only result.

## Deciding lines (verbatim, this run, HEAD 76b6dad266)

```
=== RUN   TestHomeGuard_RepositoryHasNoNewHomeAccess
    repo_guard_test.go:96: scanned 3927 test files: 160 violations, seed 184 entries, 0 new, 24 stale seed entries (stale entries do not fail; regenerate with TESTGUARD_WRITE_SEED=1)
--- PASS: TestHomeGuard_RepositoryHasNoNewHomeAccess (11.06s)
exit=0
ok  	github.com/modu-ai/moai-adk/internal/testguard	6.889s
--- PASS: TestSanitizeLeavesOrdinaryTextAlone (0.00s)
--- PASS: TestFcFixtureKeepsMoaiHomeInsideTheTestSandbox (1.00s)
golangci-lint run --new-from-rev=04d8c27a7e ./internal/agenticfactory/core/ ./internal/cli/ -> 0 issues. (exit 0)
gofmt -l <both files> -> (no output), exit 0
git diff --quiet 04d8c27a7e HEAD -- internal/testguard/ -> exit 0 (seed and guard untouched)
```

Marker check: `// home-guard:allow deliberate input: the sanitizer must leave a tilde path alone` at `internal/agenticfactory/core/sanitize_test.go:95`, directly above the literal at :96. It matches `allowComment` (`//\s*home-guard:allow\s+\S`, scan.go:30) and the placement rule (`markerLines[line-1]`, scan.go:96). The tilde input is a deliberate case in the sanitizer's ordinary-text table. The reworded `t.Errorf` message at `internal/cli/factory_fixture_home_test.go:36` changes only text.

Optional, non-blocking: F2 (cosmetic alignment split at sanitize_test.go:95-96), F3 (comments at factory_fixture_home_test.go:14,17 keep `~/.moai`; comments are not scanned by design).

Gaps: codex gate not exercised (F1); pre-fix failure not re-run here; lint scoped to new lines only; full package suites for core/cli not run.
