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
- CORRECTION (audit round 1, F2): the first version of this file claimed no weak-mode negative step needed gating; that was wrong. `TestAutoMissionTestsReceiptTrustBoundary` writes a 0644 receipt (via the helper at `goal_mission_test.go:287`) and expects a refusal, which no longer exists on Windows. It is now gated by `weakModeRefusalExpected()` (false on Windows); the test's other steps still run there.
- Audit F3 (inference, not run): `TestAutoMissionProductionGitCommitOwner` (line ~192) passes the tests-receipt check but is expected to stop at the governance receipt check in `internal/mission` (`Perm() != 0o600`), fixed by t2130; that row also depends on t2130. Audit F4 (outside the diff): `migrate_cg.go:260` has the same `&0077` check; it is not among the 11 failures and was not touched.
- Focused runs only, no whole-package suite.
- Codex not approved for this card: Claude audit only.

## Residual-risk
Windows no longer enforces owner-only access for the tests receipt through mode bits (none exist there); same policy as t2130.
