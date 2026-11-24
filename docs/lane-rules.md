# Lane rules

The detail behind the "Lane rules" section of `CLAUDE.md`. A lane is one session that works one card in a worktree of its own. The flow itself (take, write guard, READY, landing) is `skills/moai/references/flows/next-flow.md` and `skills/moai/workflows/work.md`; this file holds what those leave to the repository: where lanes work, how they hand a card back and what they never do. Where a rule moved here from the earlier `CLAUDE.md`, `docs/claude-md-rule-provenance.md` names the line it came from.

## One card at a time

- A lane works one card at a time and runs one receipt check (`moai check affected`) at a time; a follow-up slice does not start while another card of the lane is open.
- A non-blocking finding goes in the note with its file and line; a lane adds no card for it, and the leader adds it as a criterion of the SPEC's open follow-up card (`docs/leader-rules.md`, Refill).
- A READY note or a `LANE-DONE` line hands a card back; it is never the start signal for the next card. The next card is taken once, when the ledger shows the previous card done or released, and only when no leader is registered; with a registered leader the lane waits for that leader's dispatch.
- New lane sessions are opened only by the operator; no session spawns another.
- A session-end notice means: finish the card in hand, take no new card, then report `SESSION-END-READY`. It never means stop the work in hand. A card in `review` cannot be released (`card release` refuses it; `card reopen` first), so its session stays until the card is landed.

## Worktree entry and exit

- Start a card in a worktree made with an absolute path under `<project>/.claude/worktrees/` and entered by that same path with `EnterWorktree`; the steps and the reason (the host asks the person about a path that does not exist) are in `skills/moai/workflows/work.md`. Leave with `ExitWorktree(keep)` after the hand-off. A lane never deletes a worktree.
- Never move into or out of a worktree with a shell `cd`, nor with a `cd <worktree> &&` command prefix. A worktree first entered by shell `cd` is finished as it is (`ExitWorktree` is a no-op there), and the next card starts with `EnterWorktree`.
- Worktrees made earlier under `~/.moai/worktrees/`, `<primary>/.moai-cli/worktrees/` or `.claude/worktrees/` with another name are finished where they are, are not reused for the next card and are not moved or deleted. `~/.moai/worktrees/projects/` belongs to the Go generation and is never touched by a lane.
- After the card is landed, the leader trims its cargo output at once and removes the worktree after reporting the list to the user. The plan is `moai worktree-gc-plan . --integration-ref <local integration ref>` (read-only): `remove` only when the head is merged, no tracked change, no untracked file, no ignored file outside cargo output, no process inside and no change in 120 minutes; it is run again right before acting, and a changed `plan_digest` means check again. `git worktree remove` without `--force` still deletes the ignored files of a clean worktree, so a worktree that holds lane notes under `target/<name>/` is never a `remove` candidate. The plan does not read card leases or files held open by a process working elsewhere; the leader checks both.
- Cleanup removes cargo output only: profile folders (`debug/`, `release/`, with `.cargo-lock`), target-triple folders (with their own `CACHEDIR.TAG`), `CACHEDIR.TAG`, `.rustc_info.json` and `tmp/`, never the whole `target/`. `moai worktree-gc-plan` lists exactly those paths as `trim_paths` and everything else as `preserved_in_target`.
- Keep notes, bodies and audit logs under an ignored `target/<name>/` only while the card is open; copy anything that must survive into the card or the SPEC.
- Make a throwaway copy of the repository (a mutant copy), a temporary HOME and a `mktemp` folder only under `<worktree>/target/<name>/`, and point `TMPDIR` there for a run that makes its own temporary folders (`TMPDIR=<worktree>/target/<name> mktemp -d`). Remove each by a relative path before READY: a path inside your own worktree is the one place a lane deletes, the rule against absolute and home paths is unchanged, and the cleanup above does not remove `target/<name>/` for you. Never make them in the shared checkout, the system temporary folder or the session scratchpad. Whatever you could not remove goes in the READY note's `leftovers:` line.

## Hand-off

Implement, run the checks, commit, then:

1. Add one READY note with `moai card note`: first line `LANE-READY-INTEGRATION <card> head=<40hex>`, the base the checks ran on below it, the `moai check affected` receipt of the same head (summary line, every step's exit code, log folder), every acceptance criterion as met, not met or not_run, remaining responsibilities and HOLDs, and one `leftovers:` line: `leftovers: none`, or `leftovers: <path> | <path>` for what you could not remove (`moai card note` refuses a hand-back note without exactly one such line; its answer lists `shared_checkout_new`, the files that appeared in the shared checkout since your reservation, so that a file you made there is named). A head that differs from the checked head names what the extra commits change.
2. Stop writing and confirm nothing is lost: `git status --short` is empty.
3. `ExitWorktree(keep)`.
4. `LANE-DONE` as the last line of the turn.

A lane never waits for the audit or the landing, and its hand-off is not completion.

## Blocks and limits

- Report a real block (permission or hook denial, missing decision, failing gate) with the exact command and stop; a card that needs a decision ends in `LANE-BLOCKED kind=decision …` as a note on the card. Never route around an unapproved permission, spend, HOLD or security gate. Stop after three failures of the same step.
- No forced deletion or process kill, and no branch switch, reset or stash in the shared checkout.
- A rejection is not non-execution: after a denied or interrupted side-effecting command, re-query the real effects before reporting.
- A lane runs its own self-audit and repairs; it never audits another lane's change and never declares PASS for its own.

## Dangerous path

A change that deletes, resets, starts a process or builds a permission rule carries in its READY a run of the dangerous path itself in a throwaway (a link, a plain folder at a registered path, a hung child, a compound command), not only the green tests.

## Guard mutants

The READY carries one `guard-mutant: <test> | <code removed> | exit <n>` line per new guard test (`none | <reason>` without one); `exit 0` is repaired first.

## Audit clone (release pull request only)

An audit of a release pull request reads the exact head in a clone of its own. `moai audit-clone-plan . --pr <n> --head <40hex> --head-branch <branch> --base-branch <integration branch>` prints the commands: a single-branch, no-checkout clone of the head branch into `/tmp/moai-audit-<pr>-<head8>`, a fetch of the base branch only, a detached checkout of the exact head, `rev-parse HEAD` (must equal the head) and `merge-base` (the verdict's `base`). An audit clone builds in its own `target/` too. Never a linked worktree of the author's clone and never `--reference` or `--shared`. After the verdict is posted and read back, copy the logs to the plan's `evidence_dir` and remove your own clone.

## Tests

- Tests are Rust only; nobody runs a Python test module by hand as evidence. `moai check affected` still runs its own remaining Python steps and their result is reported as it comes.
- A new top-level `tests/*.rs` file registers its first word in `tests/areas.toml` and, when it spells a state folder name on purpose, its row in `tests/fixtures/state_folder_literals/allow.tsv`, in the same change; `moai check affected` stops at `test_tables` before the heavy steps and prints the line to add. A new top-level test file that walks the tree (it names `CARGO_MANIFEST_DIR` and calls `read_dir(`), reads a file outside `tests/` with `include_str!` or `include_bytes!`, or opens a checked-in file by its path registers in `tools/repo_scans.json` (a rule file of an entry, an `always` entry with a `--test <name>` check for a reader) or, when it only touches temporary or named files, in `EXEMPT`, `INCLUDE_STR_EXEMPT` or `FILE_READ_EXEMPT` of `tests/repo_scans_table.rs` with the reason; `test_tables` stops the run on a file that is in none of them (`tools/repo_scans.json` is a shared file: write it with `moai card line-lease`).
- Test in temporary projects or fixtures, never against installed user state. Do not claim Windows or Linux results from Mac fixtures.
