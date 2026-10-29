# t1995 handoff evidence

Base: 4c4217a19d98237cd2964f34a8fa942726737415. The Factory source patch was transferred with SHA-256 702acd0fcd58b7c20e765129f14f94fff3a579c77ef140f0803583a391c4af44. The lint hook compatibility patch is the supplied 0edba9e31e3dea9191bbb422ff262a6bbfbe6686 diff, SHA-256 59d404c6c16f306df8eefc9f57097a8611da89ad17f622b35d8240f7f348c6c2.

The original reproduction found same-owner expired lease refused owned, expiry-collected merge-ready holding the implementation slot, and a refused abandon changing an expired row. The isolated regression fixtures pass after the change. A PR-open card releasing the lane was already a passing baseline.

Verification on this tree: targeted race tests exit 0; full internal/homestate test exit 0; CLI package test under a non-root PID namespace with an init that reaps children exit 0, with 81 skip events; full golangci-lint run exit 0; `sh scripts/ci/test-lint-fast.sh` exit 0; `make lint-fast` exit 0; `make build` exit 0. Earlier full CLI attempts failed under umask 0077 and a missing ast-grep/tag, then under a root PID namespace; their failures were not counted as passes. The target process cancellation test passed under the reaping namespace.

Gaps: independent audit and native PR review not run; live Factory cases skipped; process-incarnation fencing beyond registration and heartbeat ordering remains to review; full pre-push preflight is not passed; production card-cycle and landing unverified. The abandon reproduction is not proof of the field incident. This is evidence for review, not a PASS verdict.

Recovery: pre-push 2026-10-27 09:52:18–10:39:30 UTC failed (exit 1), with inherited GIT_DIR contaminating the original worktree via i18n-validator fixtures. The old worktree and test-created commits were preserved. No previous full preflight PASS is claimed. A protected linked-worktree reproduction demonstrated HEAD, index, config and files remain unchanged with the recovery fix; the four impacted packages passed focused race tests. The actual Lefthook entrypoint was checked with a stub make; the full preflight and independent audit remain pending at this commit.
