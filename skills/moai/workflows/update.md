# UPDATE

`/moai update` (`$moai update` in Codex CLI) updates the current project: its MoAI-managed files, its settings, and state left by the earlier Go generation of MoAI (user decision of 2026-11-13 on #2634, Issue #2696). Every write goes through a preview the user sees and an apply bound to that preview's digest, through the same commands PROJECT uses; this route adds no writer of its own.

The user layer — the installed skill, the binary and the global instructions under HOME — is the shell command `moai-cli update`, and this route never runs it. When the user asks for that, tell them to run it in a terminal: it previews like `moai-cli update-plan`, asks once, and moves an earlier, Go-era `moai` skill folder aside only after they confirm its backup preview (`moai-cli update-apply --confirm-backup-replace`), with `moai recovery-plan` as the way back for a host left pending.

## Input

The current project, its `project.yaml` and the baselines under `manifests/` in its state folder (`.moai/` once `state_root.generation` is `rust`, otherwise `.moai-cli/`), and the user's answers.

## Output

For each step, the preview the user saw and either its applied receipt or the reason it was skipped; at the end, whether every plan now reads `unchanged`.

## Procedure

1. **Detect.** Run `moai project-detect <project>`. It writes nothing. A project of the earlier Go generation usually has no profile of this generation yet. When `result.profile` is `null` but `result.state_root.generation` is `go` or `interrupted`, do step 2 first, then send the user to SETUP (or `/moai project`) for the profile and stop; steps 3 and 4 need a profile. When `result.profile` is `null` and the generation is anything else, there is no profile to update — only when there is no profile and no Go-generation state is there nothing to update — so run SETUP (or `/moai project`) instead and stop here.
2. **Go-generation state.** When `result.state_root.generation` is `go`, follow PROJECT step 2 exactly: ask first; after a yes show `moai go-state-plan <project>` with its `items`, `go_backup`, `state_moves`, `restores`, `root_files` and `warnings`; apply with `moai go-state-apply <project> --plan-digest <digest> --permission-reference <ref>` only after a second yes (the Go `.moai/` is kept whole as `.moai-backup`, or `.moai-backup-<UTC>` when that name is taken, and a new `.moai/` is made); then report the receipt and the way back, `moai go-state-revert-plan <project> --backup <UTC>` and `moai go-state-revert <project> --backup <UTC> --plan-digest <digest> --permission-reference <ref>`. A no skips the step and nothing moves. When the generation is `rust` or `none`, the project is already on this generation; say so and go on. When it is `interrupted`, a move stopped midway: show `moai go-state-revert-plan <project> --backup <UTC>` and move nothing else until the user decides. Any other generation (`mixed`, `unknown`, `empty`, or `refuse`, which covers a `.moai` that is a link or not a folder) is reported with its `reason` and nothing is moved (a link is never followed): say what was found and let the user decide outside this route.
3. **Settings.** Pass the current profile, with the facts `project-detect` found added and every value the user set kept, on stdin to `moai project-plan <project>`. It merges three ways against the recorded baseline: a key the user changed by hand is held as `conflict`, never overwritten. Show the difference, and only after the user approves run `moai project-apply <project> --plan-digest <digest> --permission-reference <ref>` with the plan's `result` on stdin. A profile without a baseline follows PROJECT step 4's two-step rule.
4. **Managed files.** Refresh, one plan and one approval at a time:
   - the AGENTS.md managed section and the REVIEW.md section: `moai agents-plan <project>`, then `moai agents-apply <project> --plan-digest <digest> --permission-reference <ref>`;
   - the generated agent and skill files: `moai generation-plan <project>`, then `moai generation-apply <project> --plan-digest <digest> --permission-reference <ref>`;
   - only when the project opted in to GitHub templates: `moai github-plan <project>`, then `moai github-apply <project> --plan-digest <digest> --permission-reference <ref>`.
   - the project's Claude Code settings: `moai claude-settings-plan <project>`, then `moai claude-settings-apply <project> --plan-digest <digest> --permission-reference <ref>`. A key MoAI manages that the user set themselves (for example their own `statusLine`) is a conflict and the plan writes nothing. When the plan offers `confirm_replace`, show the user the key with its current and new value, and run `claude-settings-apply … --confirm-replace <token>` only after their explicit confirmation: the existing file is first kept byte for byte under `settings-backups/` in the state folder (the plan's `backup_root`), only that key changes, and the answer names the `cp` command that puts the original back (Issue #2383).
   A file or managed part the user edited is planned as `conflict` and kept; report it and let the user decide. Text outside the managed markers is never touched.
5. **Re-run.** Run the plans of steps 2–4 once more. When every one reads `unchanged` (and `go-state-plan` offers no move), report the project as up to date: a re-run plans `unchanged`. Anything else is reported as what is left, never as done.

## Exceptions and Permissions

A running session that reads `.claude/` or writes `.moai/` cannot be detected: ask the user to close such sessions before an apply. Never write a target whose current state could not be read, nor one under a recorded HOLD. A changed project refuses an old digest; preview again instead of retrying. Never edit `project.yaml`, a managed section or a Go-era file by hand to get past a conflict, and never delete what a step moves: it is kept where the step's receipt says (`.moai-backup`, `.moai/backups/` or `.moai-cli/backups/`).

In a project, run the read-only `moai go-state-plan <project>` and show the user every warning it gives about `.worktreeinclude`: a line there that selects secret files such as `.env` makes Claude Code copy them into every new worktree, including the worktrees of the `moai-impl` and `moai-audit` agents (Issue #2670). Name the line and let the user decide; never edit that file for them.

## Completion Check

Report each step as applied (with its receipt), skipped (with the reason, including the user's no) or conflicted (with the files). The project is up to date only when the re-run plans read `unchanged`. Host use of `/moai update` for this route is verified only after the installed skill carries this version (not_run until then).
