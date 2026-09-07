# t457 develop landing — gofmt re-measurement evidence

- **Date**: 2026-09-03 · window holder: lane-t465 session 463a35f3 (card t465 closing t457 first, per lead dispatch)
- **Merge**: `d41ba7479` = `git merge --no-ff WT-gofmt-drift` into develop @`7835148d3` — 155 files, +865/−762, brings card t457 (gofmt entire tree, 154 pre-existing unformatted files + format-gate absence verdict).

## Measurement (this run, this tree)

| Command | Tree | Result |
|---|---|---|
| `gofmt -l . \| wc -l` | merged develop `d41ba7479` (`HEAD^{tree}` = `adf5e255`) | **0** |
| `git ls-files -z '*.go' \| xargs -0 gofmt -l \| wc -l` | same | **0** |
| `gofmt -l . \| wc -l` | card branch after re-absorb `71f2930db` (develop `4e4607abe` absorbed) | 0 |
| `gofmt -l . \| wc -l` | pre-cleanup baseline develop `d592b0551` (t465 tree) | 154 (evidence: `.moai/reports/t465/gofmt-l.txt`) |

The merged tree carries code from three cards landed in this window batch (t461, WT-absorb-verdict, t457); the 0-count therefore covers the combined batch tree on the format axis, not t457's changes alone.

## Contention incident (reported to lead before retry, per t191 directive)

First merge attempt failed `fatal: Unable to write index.` (exit 128, no "another process" message) at develop `7835148d3`. Worktree gitdir had no lock; common gitdir held a 0-byte `index.lock` created 14:48:53 (same minute as the failure) with no open fd and no live git writer — judged a stale crash artifact after 32 min and removed under ordinary stale-lock handling. Retried once after reporting to the lead; merge succeeded. HEAD/MERGE_HEAD/worktree status verified unchanged between attempts (no partial effect).

## Residual

- Push of develop is the lead's batch action; this tree is local-only until then.
- t465's activation gate (`git merge-base --is-ancestor e1fdf00d1 HEAD` + `gofmt -l` = 0) becomes judgeable on the WT-format-gate-zero branch only after it absorbs this develop.
