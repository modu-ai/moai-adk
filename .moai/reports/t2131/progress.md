# t2131 progress (#1969 ② 2/3: internal/cli, 11 Windows failures)

## Claim
Operator policy 2026-11-06: on Windows only the portable properties are checked; POSIX keeps the exact bit checks. Applied in `internal/cli`:
1. Production: `validateMissionTestsReceipt` (`goal.go`) required `Perm()&0o077 == 0`, which is false for every file on Windows (0666) -> `tests_receipt_unsafe`. It now goes through `hostNoGroupOrOtherAccess` (new `internal/cli/filemode_portable.go`): true on Windows, exact on POSIX; regular-file and symlink checks unchanged.
2. Tests: six exact-mode assertions use `permMatchesOnHost` (Windows accepts any reported mode; POSIX equality kept): `clean_home_test.go` (dir 0700), `jev_key_test.go` (GLM and Jev credential 0600), `feedback_participation_test.go` (log 0600), `goal_blocked_question_regression_test.go` (mission file 0600), `memory_fold_test.go` (MEMORY.md 0600).

## Evidence
- Failure list (CI table `categorized.tsv`, POSIX_MODE_PERM rows of package internal/cli): `clean_home_test.go:272`, `jev_key_test.go:196/229`, `feedback_participation_test.go:111`, `goal_blocked_question_regression_test.go:79`, `memory_fold_test.go:750`, and the goal mission tests (`goal_mission_test.go:123/192/250/317/489`: `receipt_unsafe`, `tests_receipt_unsafe`, `git owner: lease missing or unsafe`).
- `go vet ./internal/cli` and `GOOS=windows go vet ./internal/cli` printed nothing; `gofmt -l` on the touched files printed nothing.
- `go test ./internal/cli -run 'TestNoGroupOrOtherAccessOn|TestCleanHome_RetentionFromHomeTier|TestGlmKeyFlagSavesKeySameStorage|TestJevKeyFlagSavesCredential|TestOutboxLogAppendOnly0600|TestMissionBlockedQuestionPersistedShapeUnchanged|TestMemoryFold_VerbatimFiling|TestAutoMissionTestsReceiptTrustBoundary' -count=1` -> `ok  github.com/modu-ai/moai-adk/internal/cli  2.602s`.
- Decision function pinned by `filemode_portable_t2131_test.go` (`TestNoGroupOrOtherAccessOn`).

## Baseline-attribution
Tree t2131 (branch WT-internal-cli-tests), commands ran in this run against this tree.

## Gaps
- The goal mission failures `receipt_unsafe` (governance/completion receipts) and `git owner: lease missing or unsafe` come from the production checks in `internal/mission`, fixed by sibling card t2130 (merge-ready, not in this tree's base). Those cli tests are expected to pass on Windows only after t2130 lands; they are not claimed fixed here.
- No Windows host: the Windows behavior is the CI race matrix's verdict. No real RED on this platform.
- If any of these cli tests also has a weak-mode negative step (chmod 0644 then expect refusal) it was not found by grep (`goal_mission_test.go` only chmods to a caller-provided mode at line 287) — checked, none needs gating.
- Focused runs only, no whole-package suite.
- Codex not approved for this card: Claude audit only.

## Residual-risk
Windows no longer enforces owner-only access for the tests receipt through mode bits (none exist there); same policy as t2130.
