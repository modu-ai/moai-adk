# Personal notes for this checkout (ignored by git)

Machine and personal facts only. Team rules are in `CLAUDE.md`; nothing here is a rule for the repository. Facts were read on 2026-11-23 on this machine; recheck one before relying on it.

## This machine

- macOS 27.2.1 on arm64 (`sw_vers`, `uname -m`); Claude Code 2.1.330 (`claude --version`), above the thresholds the earlier notes named (2.1.277 for `AGENTS.md`, 2.1.257 for Fable).
- Sessions run through Orca. Orca gives each Claude profile its own config folder, and the session's `CLAUDE_CONFIG_DIR` is that folder: `~/Library/Application Support/orca/claude-profiles/<profile id>/home`. The user settings the session reads are the profile's `settings.json` and `CLAUDE.md`, not `~/.claude/settings.json` (a separate copy that no session reads). The profile's `CLAUDE.md` imports `~/.claude/CLAUDE.md` with one `@` line.
- Lane terminals are Orca windows; the leader reads and drives them with `orca terminal` (`orca` is `/usr/local/bin/orca`). A session started inside a lane worktree reads that worktree's `CLAUDE.md` and, because the worktree sits under the checkout's `.claude/worktrees/`, this file from the primary checkout too.

## Build and cache

- `cargo` and `rustc` are `~/.cargo/bin`; `sccache` is `/opt/homebrew/bin/sccache`. `RUSTC_WRAPPER` is not set globally: set `RUSTC_WRAPPER=sccache` on test and check builds, not on release builds.
- Each worktree builds in its own `<worktree>/target/`; no `CARGO_TARGET_DIR`.
- Offline builds need a prepared cargo home. The shared folder `/tmp/moai-product-cargo` is used only once it holds `registry/cache`; today it does not exist, and `/tmp/moai-rc-product-cargo/registry` holds only a `CACHEDIR.TAG`. Use the lane's own cache `/tmp/moai-rc-cargo-<lane>` (`moai lane-env --lane <name>` picks it). Fill a new one without the network by copying `registry/` from a lane cache that builds (`mkdir -p /tmp/moai-rc-cargo-<lane>`, then `cp -R /tmp/moai-rc-cargo-<other>/registry /tmp/moai-rc-cargo-<lane>/registry`, when no build uses the source) and check it with `CARGO_HOME=/tmp/moai-rc-cargo-<lane> cargo fetch --offline --locked` (exit 0). Filling the shared folder again from crates.io needs the network and is the leader's act.
- Existing `/tmp/moai-rc-target-*` and `/tmp/moai-target-*` folders outlived their worktrees; they go only by a listed, confirmed plan.
- Heavy runs go through `moai heavy-slot-acquire` or `moai check affected`; the slot limit follows the leader's latest line.

## Tools

- `gh` (`/opt/homebrew/bin/gh`) is signed in as the operator; none of the tools below grants permission to publish or delete.
- Hide (`~/.local/bin/hide`) for panes and showing files, `aside` and `ego-browser` (`~/.local/bin`) for a logged-in browser and page reads. A page you open in Hide waits for the operator; use your own browser tool for your own checks.
- LSP: Claude Code's official plugins only, with servers installed by `rustup` or the toolchain, never by brew. A missing server: guide the install, install after the pick, continue.

## Memory

- Claude auto-memory for this project: `~/Library/Application Support/orca/claude-profiles/<profile id>/home/projects/-Users-goos-MoAI-moai-adk/memory/MEMORY.md` (the index) with the `feedback_*.md`, `project_*.md` and `reference_*.md` files beside it. Re-read the index at session start.

## Personal style

- Chat in Korean, colloquial; written reports in a clean written register. Lane dispatches and notes to lanes in Korean; subagent prompts in English.
