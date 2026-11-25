# Internal: build-slots

When to read: in RUN step 1, before the first build, test or lint run in a worktree, and from the failure-table row for a slot that exits 75.

## Slots

- Wrap every heavy build, test or `cargo clippy` command, run by hand or by a hook, in the slot command: `moai heavy-slot-acquire --name <lane> -- cargo test --offline --locked`. It waits for a free slot (printing its place in the queue, the load averages and the free disk space when its place changes), runs the command and returns the command's exit code (128 + the signal number on a signal). When the command fails it prints load and free disk again, so a failure under load or with a full disk reads as such.
- At most as many heavy runs execute at a time as `moai heavy-slot-acquire` has slots. Never count processes with `ps` and never start a run beside the slot command.
- `moai heavy-slot-status` only reads. It lists the held slots (the wrapper, its command, since when) with the machine's load and free disk. Each entry has `alive` (by the recorded pid; null when the record names none); for a dead pid it lists `holders`, the processes with the slot file open. `unavailable` with `holders` null means they could not be looked up, not that nobody holds the slot.
- Each slot is a `flock` lock under `${XDG_CACHE_HOME:-~/.cache}/moai/heavy-slots/`, held by the wrapper, the command and any descendant the command leaves running. The operating system frees it when all are gone, also after SIGKILL, so no slot file is left to clean. A slot folder from before the rename (`${XDG_CACHE_HOME:-~/.cache}/moai-cli/heavy-slots/`) is read while the new folder has no `limits.json`; its held slots show under `earlier_folder`.
- A waiter whose starting process has ended leaves the queue and exits 125 without running the command.
- A slot that stays held after its command ended is read with `moai heavy-slot-status`. Ending the holder is for its owner; a slot file is never edited or deleted by hand, and the command never ends or reorders another lane's work.
- The slot count comes from `limits.json` (default 3). A lane never raises it; choosing it is not a lane's decision. Unix only.

## Limits

The machine's limits live in `limits.json` in the slot folder: `max`, `min_free_gb`, `max_swap_pct`, `min_mem_free_pct`; defaults 3 slots, 50 GB, 80 %, 10 %. Only the person who owns the machine edits it (for example for a CI runner with a small disk).

- A caller's `--max`, `--min-free-gb`, `--max-swap-pct` or `--min-mem-free-pct` takes effect only when stricter than the file, so no caller turns a limit off.
- A limits file that is a symlink, a folder, over 4096 bytes, not a JSON object of those keys, or sets `max` 0 makes the command exit 125 without running anything.
- The command starts nothing (exit 75, with the measured values) while the working folder's file system has less than 50 GB free or, on Linux, swap use is above 80 %. On macOS, which keeps pages swapped long after memory pressure drops, it judges memory pressure: it starts nothing at `kern.memorystatus_vm_pressure_level` 4 (critical) or below 10 % memory free (`kern.memorystatus_level`, the percentage `memory_pressure` prints), and shows swap for information only.
- Exit 75 means the run did not happen: it is `not_run`, never a pass. Report it to the leader; never lower the limit yourself.
- A run that does start with less than 100 GB free prints one warning line on stderr first, naming the floor; stdout and the exit code do not change.

## Worktrees and build folders

- Every lane or audit worktree lives under the project's `.claude/worktrees/<name>`, and builds into the build tool's default folder inside it (for cargo, its own `target/`). Leave `CARGO_TARGET_DIR` unset or point it at that folder. Use one build folder per worktree and never share it with another worktree; use none in the user's home. When the slot command wraps `cargo` with a `CARGO_TARGET_DIR` shared with another worktree, inside another worktree, or naming a retired build folder in the user's home, it prints one warning line saying to unset the variable; it is a warning, not a refusal.
- Build with `CARGO_INCREMENTAL=0`; the lane environment sets it, so write no environment prefix on cargo commands. Reuse the one folder for every check and audit step in that worktree.
- Where the project keeps fixed slots (`.claude/worktrees/lane-01` to `lane-10`, `audit-1` to `audit-3` and `seed`), warm a new or reset slot with an APFS clone of the seed's build folder instead of a cold build: `cp -c -R -p seed/target <slot>/target`.
- Cargo may run a binary built from another tree in the same folder without rebuilding (two worktrees sharing one folder ran 7 of 8 tests against the other tree's binary). After switching to another worktree or audit copy, touch the files that differ, trust only a run whose `Compiling` line names the worktree you are checking, and record the free disk space before the run.
- A lane holds at most 2 worktrees at a time, and none is created under `target/`.

## Cleanup

1. When the audits and tests of its own worktree are done, a lane removes only the cargo output it made itself: the `debug`, `release` and `<target triple>` folders in `target/` or in a cargo folder directly under it. Delete from inside the worktree by relative path (`rm -rf target/debug`), never through an absolute or home path.
2. It never removes `target/` itself or the notes, evidence and scripts under `target/<name>/` by hand, and never touches another lane's worktree.
3. After the card lands, a lane removes its own merged card worktree, and an audit worktree goes when its verdict is posted. Preview first: `moai worktree-gc-plan <project> --integration-ref <ref>` is read-only and prints, per worktree, `remove`, `trim` (cargo output only) or `skip` with reasons, and a digest; `--worktree <path>` (repeatable) plans only that worktree. Then `moai worktree-gc-apply <project> --integration-ref <ref> --plan-digest <digest> --permission-reference <ref>` recomputes the plan with the same inputs, refuses before any write unless its digest equals `--plan-digest`, then trims and removes what the plan lists (no `--force`; branches are kept). The reference is recorded, not verified. The plan removes only a merged worktree with no process in it, not linked, not recently changed (`--recent-minutes`, default 120) and outside the preserved places, whose ignored files are regenerable caches, a lane's temporary files (`target/agent-N/`, `*.log`) or byte-identical copies of the primary checkout's same path. The lane notes in `target/<lane>/` go with the worktree; anything else keeps it and is reported.
4. Cleaning other lanes' worktrees, or many at once, stays the leader's periodic step: `moai worktree-gc-plan --max-checked N --cursor <next_cursor> --integration-ref <ref> <project>`.

Transcript and scratch-file retention is the user's `guarded` setting; no lane or MoAI command edits `~/.claude/settings.json`. For information, Claude Code's `cleanupPeriodDays`: "The default is 30 days and the minimum is 1"; "Scratchpad files last as long as the session's transcript" (code.claude.com/docs/en/claude-directory.md, read 2026-11-14).

## Build cache

Installing sccache, writing `~/.cargo/config.toml`, or setting a global `RUSTC_WRAPPER` or `SCCACHE_*` is the user's decision (`guarded`); nothing in MoAI does it. What the official documents say, read 2026-11-14 (sccache `README.md`, `docs/Configuration.md` and `docs/Local.md`; the Cargo reference):

- Enable it with `build.rustc-wrapper` or `RUSTC_WRAPPER=sccache`.
- `SCCACHE_CACHE_SIZE` is the local cache's maximum size; the default is 10G.
- "Incrementally compiled crates cannot be cached." sccache exits with an error if `CARGO_INCREMENTAL=1`; `SCCACHE_ALLOW_INCREMENTAL=1` lets the build proceed with those crates uncached.
- "By default, absolute paths to files must match to get a cache hit." `SCCACHE_BASEDIRS` normalizes paths; only absolute paths are supported.
- "The local storage only supports a single sccache server at a time." "Multiple concurrent servers will race and cause spurious build failures."
- Cargo: `CARGO_INCREMENTAL` "when set to 0 it will force disabling it". The `dev` profile has `incremental` on and `release` off; the incremental cache is stored in the `target` directory.
