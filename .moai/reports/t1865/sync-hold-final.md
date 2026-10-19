# t1865 — final sync hold (after the shrink audit)

Card t1865, SPEC-UPDATE-OWNERSHIP-REPAIR-001, branch WT-update-ownership-repair. The single final sync audit allowed by the leader's disposition B (board 86c2) ran on the tree after the shrink and returned FAIL. By that disposition there is no further correction and no further audit by the lane. The tree is preserved and pushed.

## State

- Audited and current HEAD: `4a625be15622eeb610df317807831d51e4257e36`, tree clean. Branch preserved on origin at the same SHA (preservation push under board d-20261019T002431Z-18ba, no PR).
- Verdict file: `.moai/reports/t1865/sync-audit.md` (`verdict: FAIL`, `required_backend: codex fail`, `convergence_overall: fail`), receipt `rcpt-e012674ec01973db90b4bb52`. The round-4 audit is kept as `sync-audit-r4.md`; rounds 1 to 3 as `sync-audit-r1.md` to `-r3.md`.
- Rejection records: 0.
- Tests the auditor ran on the final tree, all exit 0: the nine chunk-13 selectors (PASS 101), the acceptance-criteria selectors (PASS 100), restore-bookkeeping and deployed-version selectors (PASS 14); gofmt clean, go vet clean. Lane run before the audit: the eight core tests ok. Not run by anyone: the full `internal/cli` suite, coverage, `-race`, Windows test execution, golangci-lint on the final tree by the auditor.

## Findings of the final audit

| Id | Severity | Where | What |
|---|---|---|---|
| F1 | BLOCKING (codex P2) | `internal/cli/migrate_ownership.go` around line 414 | When a path component is a symlink, `os.Stat(abs)` only picks the reason (`read-failed` or `symlink-escape`), but it follows the link, so a project link that points into a profile root gets its target stat'ed, which REQ-OWR-005 forbids. The lane read the lines and confirms the call; the reproduction was not run by the lane or the auditor. |
| F2 | SHOULD-FIX | same file, around line 427 | A regular file swapped for a FIFO between `root.Lstat` and `root.Open` makes `moai update` hang. Already disclosed as a limit in progress.md E.4 caveat 6. |
| F3 | MINOR | `escapesRoot` | The stdlib message is matched against the full error text including the file name, so a missing file named `path escapes from parent.md` gets the wrong reason label (it stays unread). |
| F4 | MINOR | `pruneManagedEmptyDirs` | Not gated by `priorManifestFailed`; removes empty directories only, REQ-OWR-016 (no file) holds. |

The N-A, N-B and N-C findings of round 4 are closed. Removal stays disabled (`removalEnabled = false`).

## For the leader / operator (the lane does not choose)

- F1 and F2 are small and local: F1 drops the `os.Stat` probe (a link component then always keeps the file unread as `symlink-escape`, dangling links included) and needs a differential row with a link into a fake profile root; F2 opens without blocking and re-checks the kind on the opened handle. That is another correction round plus another audit, which the lane may not start: the ceiling and disposition B are used up.
- The lane records, without deciding, that the shrink path has now been audited once and the required codex gate found two new edges in the newest helper, as in each earlier round.

## Preserved

Tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865`, branch `WT-update-ownership-repair`, origin SHA above. Nothing deleted.
