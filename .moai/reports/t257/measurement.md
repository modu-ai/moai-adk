# t257 Measurement — `InstallPrePushHook` silent overwrite of user-modified hook

- Card: t257 (Class B — defect, cause unknown)
- Date: 2026-08-25
- Predecessor: t230 / SPEC-PRECOMMIT-PRESERVE-001 (merged, PR #1647) — behavioral standard
- Verdict: **REPRODUCED**

## Claim

`InstallPrePushHook` (`internal/cli/hook_install.go:125-155`) silently destroys a
marker-bearing, user-modified `.git/hooks/pre-push` on reinstall. It gates only on the
MoAI marker: marker present → unconditional `os.WriteFile` overwrite (lines 138-152);
marker absent → `ErrUserHookExists`. There is no content comparison, no provenance
record, no backup, and no disclosure — the exact defect shape t230 eliminated for the
pre-commit installer. The wrapper (`installPrePushHookOptional`, lines 165-178) has a
single output writer and no branch that could disclose a replacement.

Discrete claims, each measured below:

1. Reinstalling over a marker-bearing, user-modified pre-push hook loses the user's
   edit with no backup (REPRODUCED — `TestInstallPrePushHook_UserModified_ReinstallBacksUp` RED).
2. The wrapper reports the reinstall as an ordinary install — no disclosure
   (REPRODUCED — `TestInstallPrePushHook_UserModified_ReplacementDisclosed` RED).
3. The no-marker foreign-hook scenario already meets the t230 standard
   (existing `TestInstallPrePushHook_PreservesUserHook` GREEN).
4. The existing suite codifies the defect as expected behavior
   (existing `TestInstallPrePushHook_OverwritesMoaiHook` GREEN on a silent overwrite
   that t230 semantics classify as user-modified).

## Evidence (verbatim test output)

Command (scoped, this worktree):

```
go test ./internal/cli/ -run 'TestInstallPrePushHook_UserModified' -count=1 -v
```

Output (verbatim, exit 1):

```
=== RUN   TestInstallPrePushHook_UserModified_ReinstallBacksUp
    hook_install_prepush_preserve_test.go:98: SILENT LOSS: reinstall of a marker-bearing user-modified pre-push hook took no backup — the user patch is unrecoverable
--- FAIL: TestInstallPrePushHook_UserModified_ReinstallBacksUp (0.21s)
=== RUN   TestInstallPrePushHook_UserModified_ReplacementDisclosed
    hook_install_prepush_preserve_test.go:127: SILENT REPLACEMENT: wrapper output "  Pre-push hook installed (.git/hooks/pre-push)\n" does not disclose the replacement of the user-modified hook (t230 standard: notice naming the backup)
--- FAIL: TestInstallPrePushHook_UserModified_ReplacementDisclosed (0.26s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.491s
FAIL
```

Note on claim 1's failure shape: the test also asserts the hook WAS replaced with the
canonical content after reinstall; that assertion produced no error — so the overwrite
happened and returned nil while the user patch was destroyed. The failure is "silent
loss", not "install failed".

Command (existing pre-push suite coexistence):

```
go test ./internal/cli/ -run 'TestInstallPrePushHook' -count=1 -v
```

Output (tail, verbatim — the 4 pre-existing tests all PASS; only the two measurement
tests FAIL):

```
=== RUN   TestInstallPrePushHook_FreshRepo
--- PASS: TestInstallPrePushHook_FreshRepo (0.21s)
=== RUN   TestInstallPrePushHook_SkipFlag
--- PASS: TestInstallPrePushHook_SkipFlag (0.22s)
=== RUN   TestInstallPrePushHook_PreservesUserHook
--- PASS: TestInstallPrePushHook_PreservesUserHook (0.19s)
=== RUN   TestInstallPrePushHook_OverwritesMoaiHook
--- PASS: TestInstallPrePushHook_OverwritesMoaiHook (0.31s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.603s
FAIL
```

Command (vet):

```
go vet ./internal/cli/
```

Output: none — exit 0.

Reproduction procedure (mirrors the t230 pre-commit defect): isolated repo under
`t.TempDir()` → `git init` → baseline `InstallPrePushHook(false)` → user patch appended
to the installed body (marker lines in the first 3 lines untouched) → reinstall
`InstallPrePushHook(false)` → assert the t230 standard (backup holding pre-run bytes +
disclosed replacement). Both assertions fail on current code exactly as the hypothesis
predicts; the RED run itself proves the tests can fail, so no mutation step was needed.

Code-path anchors (read, this tree):

- `internal/cli/hook_install.go:138-148` — existing-hook branch checks the marker only;
  no read/classify/backup (contrast `internal/cli/hook_install_precommit.go:246-281`).
- `internal/cli/hook_install.go:150-152` — unconditional `os.WriteFile` of
  `prePushHookContent`.
- `internal/cli/hook_install.go:165-178` — wrapper with a single `out io.Writer`; no
  warning writer, no backup-notice branch (contrast
  `internal/cli/hook_install_precommit.go:395-435`, the REQ-PCP-004 warn-writer wiring).

## Baseline-attribution

- Tree: `.claude/worktrees/t257`, branch `WT-prepush-overwrite`,
  HEAD `539349c5b093d88fb6bf75a9c8a76e4fefbd138b` (re-read via
  `git rev-parse HEAD` immediately before writing this report).
- Working tree at measurement time: exactly one untracked file —
  `internal/cli/hook_install_prepush_preserve_test.go` (the measurement test itself).
  No production code modified, nothing committed, nothing pushed.
- All commands above were run in this run, against this tree; outputs are the outputs
  observed in this run, not carried over.

## Code-path comparison — pre-commit (t230, fixed) vs pre-push (current)

| Scenario | pre-commit (t230 fixed) | pre-push (current) | t230 standard met? |
|---|---|---|---|
| Fresh install | write hook + SHA-256 provenance sidecar; nil | write hook; nil — no record written | Partial: no record, so future attribution is impossible (latent, not lossy by itself) |
| Marker-bearing, unchanged vs record | classify `unmodified` (record basis) → quiet overwrite, no backup | no comparison at all; blind overwrite | Observably equivalent in this narrow case (no harm done) |
| Marker-bearing + user patch | classify `user-modified` (record mismatch, or no record + content diff → noisy `undecidable-legacy`) → backup pre-run bytes to `pre-commit.bak.<stamp>` BEFORE replace → warn-writer notice naming the backup | blind `os.WriteFile`; user patch destroyed; returns nil; no backup; no notice | **NO — the defect (REPRODUCED)** |
| No marker (foreign hook) | `ErrUserHookExists`; bytes preserved; no backup | `ErrUserHookExists`; bytes preserved | YES (parity — existing test GREEN) |

Existing-suite hazard for the fix: `TestInstallPrePushHook_OverwritesMoaiHook`
(`internal/cli/hook_install_test.go:138-171`) plants a marker-bearing legacy hook with
no provenance record — under t230 semantics the `undecidable-legacy` basis, i.e.
user-modified — and asserts the silent overwrite as CORRECT behavior. A t230-style fix
will turn that test red; it must be rewritten with the fix, not preserved.

## Gaps

- Only the marker+user-patch and no-marker scenarios were measured by new tests; fresh
  install / unchanged reinstall / skip are covered by the existing suite, not
  re-measured here.
- The disclosure test asserts on the wrapper's single writer because no warning writer
  exists on the pre-push surface. A t230-style fix will add a warn writer and change
  the wrapper signature; this measurement test is then superseded by the fix SPEC's
  own tests (it will not compile unchanged — expected for a measurement artifact).
- Caller wiring (`init.go` / `update_template_sync.go` → `installPrePushHookOptional`)
  was not audited in this run; the wrapper-level measurement stands, the end-to-end
  `moai init` / `moai update` command output was not exercised.
- Windows not exercised (macOS host only); the mechanism is pure file I/O, so
  platform risk is low but unmeasured.

## Residual-risk

- The reproduction patches by appending to the canonical body. An in-place edit that
  rewrites the first 3 marker lines would dodge the marker gate (→ `ErrUserHookExists`,
  preservation) — not measured; the defect as measured is broader than the specific
  patch shape tested (any body edit keeping the marker is silently lost).
- A fix mirroring t230 requires a pre-push provenance sidecar that does not exist yet.
  Until the first post-fix run stamps it, every legacy marker-bearing install is
  `undecidable-legacy` and will be treated as user-modified (noisy first run) — the
  deliberate t230 trade-off (REQ-PCP-005), to be carried into the fix SPEC rather
  than re-litigated.

## Artifacts

- Measurement test (uncommitted): `internal/cli/hook_install_prepush_preserve_test.go`
  - `TestInstallPrePushHook_UserModified_ReinstallBacksUp` (RED)
  - `TestInstallPrePushHook_UserModified_ReplacementDisclosed` (RED)
- This report: `.moai/reports/t257/measurement.md`

---

# Fix (t257 scope (2) — reuse the t230 discriminator)

Executed by the orchestrator (lane-12) directly: the delegated specialist hit a
provider usage limit (429, resets 08:36) after the measurement phase, and the
operator directed continuation in-session.

## Claim

1. The tier-generic core of the t230 machinery — three-way classifier, provenance
   sidecar, pre-replacement backup — was extracted to
   `internal/cli/hook_install_shared.go` (single implementation, `hookTier`
   parameterization); the pre-commit installer now delegates to it with
   behaviour unchanged.
2. `InstallPrePushHook` classifies a marker-bearing existing hook before
   overwriting: user-modified (record mismatch, or undecidable-legacy differing)
   → back up pre-run bytes to `pre-push.bak.<stamp>` BEFORE replacement, then
   disclose; unmodified → quiet overwrite. `prePushHookContent` untouched
   (byte-identical; the t255 same-release trap does not fire).
3. `installPrePushHookOptional` gained a distinct warning writer (REQ-PCP-004
   wiring); callers (`init.go`, `update_template_sync.go`) bind it to stderr.
4. `TestInstallPrePushHook_OverwritesMoaiHook` — which codified the silent
   overwrite as correct — is rewritten as
   `TestInstallPrePushHook_OverwritesMoaiHookBacksUp` (backup + replace).
5. New mirror test `TestInstallPrePushHook_UnchangedReinstall_Quiet` pins the
   quiet direction (no backup, no warning on an unmodified reinstall).

## Evidence (verbatim, this tree)

```
$ go test ./internal/cli/ -run 'PrePushHook' -count=1 -v
=== RUN   TestInstallPrePushHook_UserModified_ReinstallBacksUp
--- PASS: TestInstallPrePushHook_UserModified_ReinstallBacksUp (0.10s)
=== RUN   TestInstallPrePushHook_UserModified_ReplacementDisclosed
--- PASS: TestInstallPrePushHook_UserModified_ReplacementDisclosed (0.11s)
=== RUN   TestInstallPrePushHook_UnchangedReinstall_Quiet
--- PASS: TestInstallPrePushHook_UnchangedReinstall_Quiet (0.11s)
=== RUN   TestInstallPrePushHook_FreshRepo
--- PASS: TestInstallPrePushHook_FreshRepo (0.10s)
=== RUN   TestPrePushHookConventionBlockPlacement
--- PASS: TestPrePushHookConventionBlockPlacement (0.00s)
=== RUN   TestInstallPrePushHook_SkipFlag
--- PASS: TestInstallPrePushHook_SkipFlag (0.10s)
=== RUN   TestInstallPrePushHook_PreservesUserHook
--- PASS: TestInstallPrePushHook_PreservesUserHook (0.10s)
=== RUN   TestInstallPrePushHook_OverwritesMoaiHookBacksUp
--- PASS: TestInstallPrePushHook_OverwritesMoaiHookBacksUp (0.10s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.585s

$ go test ./internal/cli/ -run 'PreCommit' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	12.587s

$ go vet ./internal/cli/ && GOOS=windows GOARCH=amd64 go vet ./internal/cli/
(exit 0, no output)

$ golangci-lint run internal/cli/
0 issues.
```

The two formerly-RED measurement tests PASS unchanged in assertion intent (the
disclosure test moved its assertion to the new warning writer, per the fix
design — the pre-fix single-writer form no longer exists).

## Baseline-attribution

- Tree: `.claude/worktrees/t257`, branch `WT-prepush-overwrite`, base
  `539349c5b` (origin/main at fix start); all commands run in this run against
  this tree.

## Gaps

- The end-to-end `moai init` / `moai update` command output (caller-level) was
  not exercised; wrapper-level tests cover the disclosure, and both call sites
  were compile-verified and bind warn to stderr mirroring the pre-commit line.
- Full `internal/cli` package suite not run locally (repo discipline: scoped
  runs locally, full suite in CI).

## Residual-risk

- First post-fix run over legacy marker-bearing pre-push installs (no sidecar)
  with content differing from canonical will back up + warn — the deliberate
  REQ-PCP-005 noisy-legacy trade-off, carried from t230 unchanged.
- Backup accumulation over repeated user patches (one `pre-push.bak.<stamp>`
  per loss-bearing reinstall) — same noisy-but-safe posture t230 shipped;
  accumulation management remains out of scope (t230 residual-risk, unchanged).

