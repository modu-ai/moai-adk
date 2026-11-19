# Internal: build-folder

When to read: before the first build or test of a task, as `moai-impl`, `moai-audit` or a lane, and again when the task ends. It applies to cargo builds; another build tool follows the same idea with its own output folder.

Build output lives outside the worktree, in one place the user excluded from Spotlight, because build folders inside worktrees kept the file watcher (`fseventsd`) busy and held 285.8 GiB. It replaces the per-lane `/tmp/moai-rc-target-<lane>` folder for agent and audit builds.

1. Folder: set `CARGO_TARGET_DIR` to `~/.moai/build/<project>-<worktree name>/`: one folder per worktree, never shared between worktrees, never inside a worktree. `<project>` is the primary checkout's folder name; the primary checkout, and a worktree whose name already starts with `<project>-`, use the name alone.
2. Always set `CARGO_INCREMENTAL=0`: incremental output was 2.9 of a 4.2 GiB test build.
3. Owner record: the folder holds `.moai-build.json` with `schema_version` 1, `git_common_dir` and `worktree` (both absolute). Without it `moai optimize-plan` never offers the folder for deletion.
4. Setting it: set both variables and write the record yourself; item 9 names the command that does it for a lane.
5. Refusal: when HOME is not absolute or not writable, or a link sits on the folder's path, do not follow it and do not build there. Report it; if you build in the worktree's own `target/` instead, say so under `Not run` or in the report and delete it as the `Cleanup` rule of your agent says.
6. Evidence: logs, probes and scratch files stay in your worktree (an ignored folder such as `target/<name>/`), never in the build folder; deleting the build output never deletes evidence.
7. End of task: delete your folder's cargo output and report the freed size under `Cleanup`. Change into `~/.moai/build/<folder>` and delete `debug`, `release`, `tmp` and the other cargo folders by relative path; a recursive delete of an absolute path is refused. This is the one place in the real HOME an agent writes or deletes.
8. Leftovers: `moai optimize-plan` (OPTIMIZE route) lists the folders under `~/.moai/build/` with their sizes; a folder is a candidate only when its record proves the worktree is gone or its branch merged. One goes only after the user picks its row and a fresh preview still marks it `trim`.
9. `moai lane-env --lane <lane>` sets this same folder and record, plus the lane's `CARGO_HOME`. `moai heavy-slot-acquire` warns only when `CARGO_TARGET_DIR` is shared with another worktree (the former per-lane `/tmp` folder, or another worktree's folder or record) or lies inside another worktree.
