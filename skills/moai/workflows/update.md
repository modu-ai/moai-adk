# UPDATE

## Input

The installation owner, the version and digest on every installed host, and the user's current changes.

## Output

An update plan covering all hosts, a per-file three-way change preview, and per-host pending→committed apply receipts with reload state (`update-plan` / `update-apply` in 2026.11.0-rc.1), plus a read-only recovery plan and a recovery apply for hosts left pending (`recovery-plan` / `recovery-apply`).

## Procedure

1. Identify the MoAI skill, settings, and owner on every installed host. Selecting individual hosts is not the default.
2. For the package that will be copied independently to each host, compare its signature, digest, protocol, schema, host matrix, and product generation. The new RC uses moai-cli with its own update feed and state, and never calls the existing Go moai through PATH in its place. Distinguish the first product switch on a host from later UPDATEs.
3. Present the baseline/current/new 3-way diff for every target together with permission changes and user edits.
4. Pin in-progress runs to the existing version, preserve user-modified files, and apply the allowed changes alongside them. Do not share runtimes through symlinks.
5. Confirm reload, explicit invocation, a protocol smoke test, and the version on each host, then aggregate the overall result.
6. An existing `moai` skill folder this tool did not install (for example an earlier-generation, Go-era MoAI skill) blocks the host. `update-plan` then previews a backup-then-replace for that host: the files with their hashes and modes, the backup location under `~/.moai-cli/backups/`, and a `confirm_token` bound to the folder's contents. Show the preview to the user and run `update-apply … --confirm-backup-replace <token>` only after their explicit confirmation for that host; replacing the folder changes the `/moai` skill of every running session, so the user also chooses when. The original folder is moved whole into the backup beside a manifest, never deleted, and `recovery-plan` / `recovery-apply` can move it back byte for byte.
7. On partial failure, keep the hosts that succeeded. On a failed host, restore only the managed files this request changed whose current hash and ownership confirm the restore is safe, and preserve any further user changes. Never assume an update across several hosts is atomic. A failed host stays `pending` and takes no new update until it is resolved: run `recovery-plan` (read-only; each file is `matches_after`, `matches_before`, `conflict` or `missing` by its current bytes), show the user the plan, and only with their approval run `recovery-apply`. It removes files this request created and restores replaced files from their known-good copy, never writes a host that committed, keeps every conflict, and does not retry after a failure. A host with conflicts left stays `pending`; report those conflicts instead of calling the update recovered.

## Exceptions and Permissions

Never force-overwrite user changes, untracked files with the same name, or unknown keys. Never treat a new scope, permission, or channel as implicitly approved. Never hide a partial failure or a recovery conflict.

In a project, show the user every `go-state-plan` warning about `.worktreeinclude`: a line there that selects secret files such as `.env` makes Claude Code copy them into every new worktree, including the worktrees of the `moai-impl` and `moai-audit` agents (Issue #2670). Name the line and let the user decide; never edit that file for them.

## Completion Check

State the result for every installed host and every remaining conflict. Never report a verification or external effect that was not performed as complete.
