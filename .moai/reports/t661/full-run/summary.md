# t661 full internal/cli package run (lead-approved slot, 2026-09-11)

Command: `go test ./internal/cli/ -count=1 -timeout 1200s -v > cli-full.txt 2>&1` (tree HEAD 8861553bd, clean except this dir; see tree-head.txt)
Window: 2026-09-10T18:43:41Z .. 19:03:45Z (UTC). Load before `{ 11.78 14.89 19.07 }`, after `{ 18.30 18.41 20.11 }` on 16 cores.
Pre-check: foreign go test on internal/hook and internal/spec only; go list -deps shows internal/cli 0 for both (precheck-pgrep.txt).

## Result
- exit 1 (cli-full-exit.txt); `panic: test timed out after 20m0s` at cli-full.txt:11673, running test at the moment: TestTodoPR_RendersOutcomeAndConfidence (0s).
- `--- FAIL` top-level (4), names identical to the lead-listed known reds:
  - TestHomeStateChangedSurfaceCoverageConsumesFreshProfile (7.51s)
  - TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite (196.54s)
  - TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition (3.49s)
  - TestAuditLagUsesBinlagSeam (1.16s)
- No other `--- FAIL` line among the tests that reported a result.
- Coverage is INCOMPLETE: defined top-level tests 3614 (grep of `func TestX(t *testing.T)`), started 2992, finished 2739; 622 never started, 253 started (parallel, paused) without a result. Sum of reported top-level durations 1198.12s, i.e. the run consumed the whole timeout.
- SKIP 15, every reason is an env opt-in or environment condition (list in cli-full.txt); TestAuditPinLive_GLMDifferential skipped on `MOAI_AUDIT_PIN_LIVE != 1`, not on the sandbox. None names the home seam.
- C redirect log line count: 1 (cli-full.txt:149).

## Home fingerprint (same command set before/after, stderr separate)
cmp exit 0 for all 8 pairs: sha.out/.err, stat.out/.err, hooks.out/.err, hooksparent.out/.err.

## Residue
The timeout panic skipped TestMain cleanup, so the sandbox dir /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/moai-cli-home-1490004665 remains (empty, 0 entries). Not deleted.

## Verdict status
HOLD: the run timed out with 622 tests never started under load ~18 on 16 cores; the 4-name match covers only the executed part.
