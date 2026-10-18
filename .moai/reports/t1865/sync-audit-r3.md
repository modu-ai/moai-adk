verdict: FAIL
audited_sha: 1d9b06a737685c32f47485f76578e5b7ae892ea7
required_backend: codex fail
convergence_overall: fail

auditor-model: claude-sonnet-5-5

# Sync audit round 3 (delta) - SPEC-UPDATE-OWNERSHIP-REPAIR-001, card t1865

Rounds 1 and 2 are preserved in sync-audit-r1.md and sync-audit-r2.md. This file replaces the round-2 file as the current verdict.

## Claim

The round-3 delta (N1 dcbafe568c, N2 90296319b4, boundary fixes 1003a9edd4, docs) is correctly implemented in code, but the required codex backend FAILED with three new findings. A required-backend fail forces FAIL.

## Evidence

- `git rev-parse HEAD` -> `1d9b06a737685c32f47485f76578e5b7ae892ea7`; `git rev-parse --show-toplevel` -> `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865`; `git status --short` clean.
- Audit plan (`moai verify audit-plan`): `config_status: ok`, `cross_model_required: true`, `enforced_required: ["codex"]`, codex gate `required`, claude and glm `off`.
- `mcp__moai__audit_multi` (project_root = worktree, target baseBranch, no explicit gates), verbatim members: `per_backend_verdicts`: backend `codex`, gate `required`, verdict `fail`; `overall_verdict: "fail"`; `plan_source: "config"`; `residual_risk_note: "required-backend FAIL: codex"`; `audit_receipt: "rcpt-665b06f8d214d9a97633b90c"`; `build_commit: ee981a932`. The codex leg did NOT fail for the sandbox reason "no go files to analyze"; it returned three substantive findings (below). This verdict is not final until the orchestrator's convergence_check passes (it cannot: gate unmet).
- Invariants: `git diff --quiet cdd72a814 HEAD -- internal/config/protected_zone.go` / `internal/config/userroots.go` / `internal/profile` printed "unchanged" for all three. `removalEnabled` appears only as `var removalEnabled = false` in non-test code (migrate_project_assets.go:40). `t.Skip` grep over the card's test files: only a comment and a `first.Skipped` log, no call. `git ls-files --others --exclude-standard` empty (no stray files). No code in the diff opens a profile root.
- Code read at HEAD (`git diff 38c8ca5376 HEAD` on the four Go sources):
  - N1: `IsPriorManifestBookkeeping` returns false when `filepath.ToSlash(relPath)` contains "/", so only root-level entries match. Correct.
  - N2: `claimPriorManifestDir` cleans runDir first, calls seam `priorManifestMarkerCopy`, on error does `os.Remove(sibling)` (best effort) and returns an error; `copyRunMarker` treats absent or non-regular marker as nothing to copy, and CRLF is handled by `strings.TrimSpace` on the first line. 99-sibling cap fails closed with an error.
  - Boundary: `withinRootBoundary` now uses `rel != ".."` and `".."+Separator` prefix (element-wise); `pathLeavesRoot` makes a relative root absolute and EvalSymlinks-resolved, empty root fails closed, EvalSymlinks error on the target returns false and the following read fails on the same condition (ELOOP, ENOTDIR, EACCES all yield read-failed).
- Codex findings, each cross-read against the code:
  - F-A: migrate_ownership.go classifyUnregistered/pathLeavesRoot then `os.ReadFile(abs)` (line ~384) is check-then-use on a path; codex reports a swap race reading an outside file at the 17th iteration. Plausible from the code: no handle pinning.
  - F-B: `backup.WriteRunMarker` is `os.WriteFile`, which leaves an empty marker when the write fails after creation; `os.Remove(sibling)` then fails with ENOTEMPTY (comment in code admits best effort), so the sibling stays. Confirmed by reading; the injected-seam test cannot reach this because the seam fails before any file is created.
  - F-C: on a case-insensitive file system (darwin default), a link whose target spells an in-project directory in a different case compares as outside (string compare of Rel), so the file is kept as symlink-escape. Fail-closed false positive.

## Baseline-attribution

Judging build for the MCP review: commit `ee981a932`, an ancestor of tree HEAD `1d9b06a73` (build_lag reported by the tool); the audit read the tree source at HEAD through the overlay, not the installed binary's behavior. My own measurements are git/grep/source reads at HEAD in this run.

## Gaps

- Not run by me: any go test, coverage, -race (machine load average 155 at audit time; brief forbade whole-package runs). The RED cells and mutants in E.2 chunk 11, and the 43-row TestPathBoundaryCounterexamples, were therefore NOT re-observed by me; claims about them rest on the lane's record only.
- Not re-checked in this round for lack of budget: the progress.md E.4 caveats 1-8 wording, the CHANGELOG single entry, the sync_commit_sha, the design.md D-1 row by 0a33265120, the SPEC-wording debts and the 21 vs 22 AC counter note, and the 12 binding conditions (N4-remainder, N14-N24). I read only the sync-audit repair paragraph of progress.md (consistent with HEAD, final run code commit 1003a9edd4). These stay UNVERIFIED.
- Windows execution, case-insensitive FS and symlink-race reproductions were not run by me; F-A and F-C are codex-reproduced, F-B is code-confirmed.
- `sync-audit.md` before this write was byte-identical to `sync-audit-r2.md` (same size, 21943); not independently diffed.

## Residual-risk

Each round has found a new edge in the newest helpers; a round 4 may find more. Removal is disabled (`removalEnabled` false) and all three findings fail closed or only leave residue, so no data-loss path was found.

## Score (harmonic, with must-pass firewall)

| Dimension | Score | Verdict |
|---|---|---|
| Functionality (40%) | 80/100 | PASS (N1, N2, boundary fixes verified in code; test execution not re-observed) |
| Security (25%) | 74/100 | UNVERIFIED-for-pass: codex required gate failed on a check-then-use race |
| Craft (20%) | 78/100 | PASS (cleanup gap F-B) |
| Consistency (15%) | 85/100 | PASS |

Overall FAIL, driven by the required codex backend (gate unmet), not by a must-pass score collapse.

## Findings

- F-A [SHOULD-FIX] [blocking for the required gate] internal/cli/migrate_ownership.go:384 - check-then-use race: the boundary is checked with EvalSymlinks and the original path is then re-opened by `os.ReadFile`; a concurrent link swap reads a file outside the project (hashed, not deleted). Requires a local attacker writing inside the project during the run. Introduced by this card only as a residual of its non-atomic guard; at base no guard existed (the read was unconditional). Required fix: open the file once without following a final link (O_NOFOLLOW or open then fstat and verify the opened handle's resolved location), read from that handle; or disclose as accepted debt in E.4 with the attacker model.
- F-B [SHOULD-FIX] internal/cli/update_prior_manifest.go:198 - when the marker write fails after the file was created, an empty `.moai-run` remains, `os.Remove(sibling)` fails ENOTEMPTY, and the sibling leaks (one per failed run, up to the 99 cap, then the claim errors). Introduced by the N2 fix of this card. Required fix: on error remove `sibling/.moai-run` before removing the directory, report a cleanup error, and add a test using a write failure that creates the file (not the injected seam that fails first).
- F-C [MINOR] internal/cli/migrate_project_assets.go:629 - on case-insensitive file systems a link target spelled with different case than the real root compares outside and is kept as symlink-escape. Fail closed; the same string comparison was already in `disposition` at base. Required fix: compare with os.SameFile on the ancestor chain, or disclose as a caveat.
- No other new defect found in IsPriorManifestBookkeeping, copyRunMarker (CRLF, extra lines, non-regular marker), the 99 cap, trailing-separator cleaning, a relative or symlinked project root, or EvalSymlinks error kinds.

## Recommendations

Fix F-B (small, test-first), decide F-A (fix or disclose with attacker model in E.4), disclose F-C; then request a delta re-audit scoped to these three.

AUDIT-VERDICT: FAIL spec=SPEC-UPDATE-OWNERSHIP-REPAIR-001 receipts=rcpt-665b06f8d214d9a97633b90c
