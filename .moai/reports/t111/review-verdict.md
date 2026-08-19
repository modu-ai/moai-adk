# t111 review verdict — backup-before-delete in CleanMoaiManagedPaths

- Reviewer: lead session (operator-confirmed direct verdict; review hub)
- Card: t111 · Worktree: `.claude/worktrees/t111` · Branch: `WT-t111` @ `aefaddb71` (base `5c3141372`)
- Delta reviewed: `5c3141372..aefaddb71` (16 files, +525/−53; 1 evidence + 15 code/test)
- Lens: `--deep` (backup-delete pipeline + file/path manipulation — destructive adjacency)
- Evidence read: `.claude/worktrees/t111/.moai/reports/t111/run-evidence.md` (61 lines, 5-section)
- Defect class: the 2026-08-15 incident (12 local-only files destroyed by `moai update`; CLAUDE.local.md §2.3)

## Verdict: PASS

## 1. Claims reviewed (dispatch + evidence vs this review's direct reads)

| # | Claim | Check performed | Result |
|---|-------|-----------------|--------|
| 1 | Commit/stat | `git log` + `diff --stat` — exact (16 files, +525/−53) | PASS |
| 2 | Backup-then-delete contract | `backupThenRemove` read in full: backup FIRST, `os.RemoveAll` SECOND, backup failure returns BEFORE removal (both file and directory branches; glob loop and config root included) | PASS |
| 3 | ① Exclusion condition correctness | File target → skipped iff template carries the exact path (`templateCarries`); dir target → `templateManagedPaths` set under the prefix, only absent paths copied; template-absent prefix → empty set → everything backed up. Managed files are redeployed moments later — their only copy is never at stake | PASS |
| 4 | ② Backup-failure abort actually tested | `TestCleanMoaiManagedPaths_BackupFailureAbortsRemoval` exists (`deploy_preclean_backup_test.go:189`) + 4 more: BackupsUnmanagedFiles, ConfigTreeReachesBackup (the astgrep-rules incident class), SettingsFileBackedUp, GlobMatchBackedUp | PASS |
| 5 | ③ Rejected-alternatives rationale | Evidence §1.4: marker (needs foreknowledge, useless post-incident), manifest-only-delete (stale old-template files linger), preserve_paths (ops burden, recurs when unconfigured) — adopted hybrid reuses the in-file REQ-UDS-008 pattern. Rationale sound | PASS |
| 6 | fs.FS injection | Signature `(root, out, tmplFS fs.FS)`; caller loads `template.EmbeddedTemplates()`; deploy stays leaf (no template import); tests use `fstest.MapFS` to control the managed/unmanaged split | PASS |
| 7 | Embed-load failure abort | Caller code read: load error returned before cleanup runs — "abort rather than delete blind" | PASS |
| 8 | Config root covered | `.moai/config` now routes through `backupThenRemove` (previously relied on the wholesale Backup step, which missed unmanaged subtrees — the incident's exact class) | PASS |
| 9 | Backup location safe from self-deletion | `defs.BackupsDir = ".moai-backups"` (internal/defs/dirs.go:12) — repo-root sibling, NOT under any managed clean root | PASS |
| 10 | Destructive-sites registry migration | Row moved `CleanMoaiManagedPaths` → `backupThenRemove`, Protection text updated to the t111 contract; drift guard `TestDestructiveTargetRegistry_CoversAllSites` PASS | PASS |
| 11 | Single timestamp per run | `backupBase` computed once — every root's unmanaged files under one `.moai-backups/<ts>/pre-clean/` tree | PASS |
| 12 | Non-regular files | Symlinks/fifos skipped (no link escape out of the backup tree — same rule as `copyTree`) | PASS |
| 13 | Not-exist tolerance preserved | `backupThenRemove` returns no-op success on IsNotExist — historic behavior kept | PASS |
| 14 | Test outputs | Lane-attributed: deploy suite ok 0.860s (new 5 included), cli `-run "Update|Clean|Deploy"` ok 17.353s, Destructive -v 4/4 PASS, lint 0 | PASS-attributed |

## 2. Evidence (this review's commands)

- `git log / diff --stat / --name-only` @ `aefaddb71`
- Full diff read: `deploy.go` (core), `update_template_sync.go` (caller wiring), `update_destructive_registry.go` (row migration)
- Test-function inventory via `git show …deploy_preclean_backup_test.go | grep 'func Test'`
- `git grep 'BackupsDir ='` — backup root location
- Read: run-evidence.md in full

## 3. Baseline attribution

Code reads @ `aefaddb71`; test outputs lane-attributed (same discipline as prior verdicts — CI owns the full matrix). Note the lane's own honest baseline note: initial investigation read the primary's `deploy.go` (main checkout) and mis-edited once before re-reading the release revision — corrected before commit; release-base diff reviewed here is the corrected one.

## 4. Gaps

- `moai update` not executed end-to-end (dev-project discipline, CLAUDE.local.md §13 spirit) — contract covered by the 5 dedicated unit tests instead.
- Full cli suite not run locally (lane-local discipline).
- Operator requirement 1 (fallback queue adoption) is deliberately split: t106 owns adoption, t111 owns update-path losslessness — two-card split recorded; t106 already merged review-PASS.

## 5. Residual risks

- Backup is a copy, not an auto-restore — the operator restores manually from `.moai-backups/<ts>/pre-clean/`; the progress line ("backed up N unmanaged file(s)") is the only announcement.
- `CleanupOldBackups` pruning applies to pre-clean backups too — restore points age out by the existing retention policy (restore promptly).
- Non-regular files (symlinks) are excluded — a link itself can still be lost.
- Mergeable-set in-memory preservation (REQ-UDS-001/002) now coexists with the disk backup for the same files — harmless duplication, kept deliberately.
- Integration note: WT-t111 branched from `5c3141372`; the release tip has since advanced (t85+t94, t95, t97-alignment, t99, t106 merges — currently `0ede5db6a`). t111's surfaces (update/deploy + cli update tests) do not overlap the landed cards' files materially; merge against the current tip after fetch.
