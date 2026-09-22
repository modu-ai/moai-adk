# t263 — file_changed.go die-at-exit: Reproduced, Remedy Not On This Base (2026-09-02, lane-7)

## Claim

The card's defect is REAL and reproduced: the FileChanged hook's incremental MX scan +
sidecar update is a fire-and-forget goroutine in a process that exits immediately, so the
side effect loses the race against process exit and essentially never lands — the exact
die-at-exit shape t216 M4 deleted from the session-start cold-start scan. BUT the card's
prescribed remedy ("M4가 세운 자가빌드 경로를 재사용") is not executable on this base:
the M4 commit lives only on the unmerged branch `WT-hook-wiring-drift` and is absent from
develop. Building any alternative route here would be the second implementation the card
itself forbids. Disposition: verified + sequenced behind t216's integration — the lead
owns that ordering call.

## Evidence

All measurements 2026-09-02, worktree `.claude/worktrees/t263`, branch `WT-incremental-rebuild`
(develop `e45054c56` + this verdict commit), tree-built binary (`go build -o bin/moai ./cmd/moai`).

**Reproduction — binary-level, 5 runs, 0 landings:**

Scratch project (`/tmp/t263-repro.E56j0R`, one `sample.go` carrying an `@MX:ANCHOR`);
per run: remove `.moai/state/mx-index.json`, fire the REAL hook path
(`CLAUDE_PROJECT_DIR=<scratch> ./bin/moai hook file-changed < input.json`, stdin carrying
`file_path`/`change_type`/`cwd`), then stat the sidecar.

| Run | rc | stdout | sidecar after exit |
|---|---|---|---|
| 1 | 0 | `{}` | ABSENT (`.moai/state/` not even created) |
| 2–5 | 0 | — | ABSENT ×4 |

`LANDED=0/5`. Run 1's stderr is EMPTY — no `slog.Warn` from any of runMXScan's fail-closed
early returns, and no state directory was ever created, so the goroutine never reached
`UpdateFile`'s `MkdirAll` (sidecar.go:103): it was killed before completing. The missing
`state/` directory is the discriminator that rules out the other silent-skip paths
(containment, EvalSymlinks, scan failure would each have left a Warn line or a directory).

**The handler is functional; only its lifetime is the defect:**

- `TestFileChanged_SideEffectsCompleted` (`internal/hook/file_changed_test.go:169`) proves
  the sidecar IS written when `WaitForAsync` drains the handler's WaitGroup.
- `deps.go:276` registers the handler; `EventFileChanged` is wired in `settings.json`
  (matcher `FileChanged` → `handle-file-changed.sh`, with the missing-hook fallback) —
  this is a live production path, not dead code.
- `runHookEvent` (`internal/cli/hook.go:267-374`) has NO wait for handler side-effect
  goroutines: its only flush barrier (`registryShutdowner.Shutdown()`, :328) drains the
  async TRACE writer, not `fileChangedHandler.wg` — whose own doc says production callers
  MUST NOT block on it. `Handle` returns `&HookOutput{}` immediately (:121) and the process
  exits, killing the goroutine.

**The prescribed remedy is absent from this base:**

| Check | Command | Result |
|---|---|---|
| M4 commit on develop | `git merge-base --is-ancestor 8aa96bfb1 develop` | rc=1 (not an ancestor) |
| M4 commit on this HEAD | `git merge-base --is-ancestor 8aa96bfb1 HEAD` | rc=1 |
| M4's self-build module | `ls internal/cli/mx_index.go` | No such file |
| M4 commit containment | `git branch -a --contains 8aa96bfb1` | only `WT-hook-wiring-drift` (+ its remote ref) |

`8aa96bfb1 feat(SPEC-HOOK-WIRING-DRIFT-001): M4 MX index build moves to mx query (t216)`
(2026-08-25) changed `internal/cli/mx_index.go` (+90, NEW), `mx_query.go` (±18), and
`session_start.go` (−169, the cold-start scan deletion) — all on `WT-hook-wiring-drift`
only. On develop today: `mx query` still fails with `SidecarUnavailable` on an absent index
(`mx_query.go:100-104`) and only refreshes a STALE index (`:106-122`, REQ-GF-007) — the
M4 behavior the card says to reuse does not exist here. The t216 queue card is itself
still `picked`.

**Sibling sweep (same shape, not fixed by this card):** `config_change.go:73`,
`notification.go:78`, and `task_created.go` spawn the same fire-and-forget goroutines
(bounded by the same `asyncDeadline`) inside the same short-lived hook processes. Same
disease; sequencing them behind t216's landing is the lead's call.

**The card's 부수 item (stale-index path), answered by code read + the same race:**
`Manager.UpdateFile` (`internal/mx/sidecar.go:155-184`) has no staleness logic at all —
it loads (absent/corrupt absorbed into an empty sidecar, :78-88), swaps one file's tags,
writes. It can never return `SidecarUnavailable` (that is the read-side resolver's error,
`resolver_query.go:146`) and never rebuilds; staleness handling lives in the query path
(`MXIndexNeedsRefresh` → `RefreshIndex`). With the incremental writer dying at exit, the
index in production is maintained by `moai mx scan` (manual) and the query-side refresh
alone.

## Baseline-attribution

Measured in this run, this tree, tree-built binary. The 0/5 landing rate is this session's
own measurement; the 2/153 sibling figure is quoted from M4's own commit message
(`8aa96bfb1`) as recorded history, not re-measured. Scratch residue:
`/tmp/t263-repro.E56j0R` (OS-cleared; the sandbox refused `rm -rf` on absolute tmp paths —
left in place deliberately).

## Gaps

- The landing rate was measured at 5 runs (0/5), not M4's 153 — sufficient to establish
  "essentially never lands" alongside M4's own sibling measurement, not a precise rate.
- `WT-hook-wiring-drift`'s completeness (M1–M6 done? sync-audit?) was not audited from
  here — only that the branch exists, contains M4, and re-measured against a develop merge
  (`9a1434912`, `e8050c135`). The lead's t216 sequencing owns that read.

## Residual-risk

- If the lead instead orders the file_changed fix WITHOUT t216, the honest options are
  (a) synchronous in-Handle execution — a reversal of SPEC-V3R6-HOOK-ASYNC-EXPAND-001's
  async design, or (c) bare deletion on a base where the query self-build does not exist —
  leaving the index with no automatic builder at all. Both are worse than sequencing; this
  lane deliberately chose neither.
