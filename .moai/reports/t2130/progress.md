# t2130 progress (#1969 ② 1/3: internal/mission, 21 Windows failures)

## Claim
Operator policy 2026-11-06: on Windows only the portable properties are checked; POSIX keeps the exact bit checks. Applied in `internal/mission`:
1. Production: the receipt readers required `Perm() == 0o600` (`readGovernanceReceipt`, `LoadCompletionReceipt`) and `git_owner` required no group/other bits. On Windows `os.Stat` reports 0666, so every receipt read as `receipt_unsafe` / "lease missing or unsafe". The three checks now go through `hostPrivateModeOK` / `hostNoGroupOrOtherAccess` (new `internal/mission/filemode.go`), which return true on Windows and keep the exact comparison elsewhere. Regular-file and no-symlink checks are unchanged on every platform.
2. Tests: the three exact-mode assertions (`completion_receipt_test.go`, `governance_receipt_jev_test.go`, `coverage_test.go`) use the same helper.

## Evidence
- Source: `.moai/reports/pr-sweep-20261106/pr1969-ci-analysis.md` (t2077 tree): `mission governance: receipt_unsafe`, `receipt mode = -rw-rw-rw- err=<nil>, want 0600` (mission 21).
- Decision functions pinned by `filemode_t2130_test.go` (`TestPrivateModeOK`, `TestNoGroupOrOtherAccess`: windows 0666/0444 pass; linux/darwin 0640/0644/0666 fail; 0600/0400 pass).
- `go test ./internal/mission -count=1` -> `ok  github.com/modu-ai/moai-adk/internal/mission  9.331s` (whole package, existing POSIX assertions unchanged in effect).
- `GOOS=windows go vet ./internal/mission` printed nothing; `gofmt -l internal/mission` printed nothing.

## Baseline-attribution
Tree t2130 (branch WT-internal-mission-tests); commands ran in this run against this tree.

## Gaps
- No RED against the real Windows behavior (no Windows host); the decision functions are pinned with explicit `goos` values and the production call sites were verified on this platform. CI Windows race is the verdict.
- Other mission tests might still fail on Windows for other reasons (path spelling, symlink creation); only the permission-bit cause named in the analysis was addressed.
- Codex review was not requested (leader: Codex not approved for this card); Claude audit only. No `card-review.md` from Codex; see `card-review.md` for the self-review note.
- Focused package run only.

## Audit round 1 (FAIL 0.62 @19aa331da4) and its repair
- F1 (blocking): simulating Windows with the host wrappers forced to "windows" showed three negative tests would newly fail there (`completion_receipt_test.go` "weak receipt accepted", `git_owner_test.go` "weak lease mode accepted", `governance_receipt_test.go` "weak mode err=<nil>"): each chmods a file to 0644 and expects a refusal, which no longer exists on Windows by policy. Repaired: those weak-mode expectations run only where the mode is detectable (`weakModeDetectable()`, false on Windows); every other step of the three tests (missing, outside path, symlink, malformed JSON) still runs on Windows.
- Audit gaps carried: Codex gate unmet by instruction (not approved for this card); O2 Windows ACL not enforced (disclosed residual risk, accepted by the auditor).
- After the repair: `go test ./internal/mission -count=1` -> `ok  github.com/modu-ai/moai-adk/internal/mission  11.241s`; `GOOS=windows go vet ./internal/mission` and `gofmt -l internal/mission` printed nothing.

## Residual-risk
On Windows the receipts' "private to the owner" property is no longer enforced by mode bits (there are none); access control there rests on the user profile directory ACLs. This is the operator-chosen policy.
