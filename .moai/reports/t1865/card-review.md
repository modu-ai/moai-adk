# t1865 — card review (advisory)

Card t1865, SPEC-UPDATE-OWNERSHIP-REPAIR-001, branch WT-update-ownership-repair. Tool: `mcp__moai__codex_review` scope=card, backend codex, merge base `48bfe97afa4b78cb1c1ff2c98f1a5680b1441309`, tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865`. The review is static and advisory: it ran no test, build or reproduction command, files no audit receipt and does not replace the sync audit or the leader's evidence read.

## Rounds

| Round | Reviewed HEAD | Verdict | Findings |
|---|---|---|---|
| 1 | `7b758103d8` | fail | 2 x P2 |
| 2 | `f9c4c76d4f` | fail | 1 x P2 |
| 3 | `1b5332af57` | pass | none |

## Findings and dispositions

Every finding was checked against the code by the lane before any action; none was accepted on the review's word alone.

1. **P2, `internal/cli/update_template_sync.go:825` — `manifest.prior.json` copied into `.moai/config`.** Confirmed: the legacy config restore walk skips `backup_metadata.json`, `.template-defaults`, the run marker and the unmanaged config directories, but not the prior-manifest copy kept in the run backup, so a project whose pre-existing config has no `sections/` got `.moai/config/manifest.prior.json`. A defect introduced by this card. Fixed test-first: tests `73e37d3832`, RED cells `9c00e81a2c`, fix `051ddc18ef` (`backup.PriorManifestName` shared constant, skipped by the restore walks only at the backup root; a same-named file in a subdirectory is still restored as user content).
2. **P2, `internal/cli/migrate_project_assets.go:110` — removal backup loses the execute bit.** Confirmed: `backupRemovedAsset` wrote every copy with `defs.FilePerm`. The AC-OWR-15 mode clauses had not caught it because the fixture held no executable file. Latent (the removal switch is off in production). Fixed in the same commits: the fixture gained a 0755 managed file, the candidate carries the mode, the backup copy keeps it.
3. **P2, `internal/cli/migrate_ownership.go:324` — an unregistered FIFO makes `moai update` hang.** Confirmed: the verdict read every non-directory with `os.ReadFile`. A regression against the base, which kept unregistered paths unread. Fixed test-first: test `2eeb0559bd`, RED cell `60d38d1559`, fix `1d5ecbbc7c` (the verdict decides the file kind with `os.Stat` before any read; non-regular files are kept with the new reason `non-regular-file`). Follow-ups: the FIFO test fails loudly instead of skipping where `mkfifo` is refused (`2b801c1971`); the symlink-escape case moved to a `//go:build !windows` file with a fatal on failure (`4f970def44`); the closed keep-reason list in design.md D-1 gained `non-regular-file` by a manager-spec revision (`0a33265120`).

## Found while fixing, not fixed in this card

- The legacy config restore walk also copies `removed-assets/**` and `in-memory-backups/**` into `.moai/config`. `removed-assets` belongs to card t1668 and is empty while the removal switch is off; `in-memory-backups` predates this card. Reported to the leader as a t1668-side follow-up candidate.
- A regular file swapped for a FIFO between the kind check and the read is not covered (an `O_NONBLOCK` open would close it; beyond a minimal change).

## Gaps of this review

Static review only. Behaviour was verified by the lane's own scoped runs (see progress.md section E.2, chunks 5-8, and the sync audit), not by this review. The review cannot see the whole `internal/cli` package suite, `-race`, or Windows execution; those are carried to CI and the landing remeasure.
