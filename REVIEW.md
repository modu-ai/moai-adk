# Review instructions

Read with AGENTS.md. `moai review-package` hands this file to every reviewer from the
base commit, and `moai review-check` refuses a finding without a severity, a file and
line in HEAD, evidence quoted from that line, and a confirmed verification.

## Important (blocks the merge)

- A behaviour change without a test that failed before the change and passes after it.
- An enforced limit without tests on both sides of it: N is accepted, N+1 is refused.
- A test that cannot fail when the code it guards is removed: every fixed option, guard or
  branch the change relies on needs an input where it changes the result.
- A promise of identical output whose test varies only HOME: the repository's own
  `.git/config` and the worktree `.gitattributes` must be varied too.
- A test that writes outside its temporary directory, reads or writes the real HOME, or
  reaches the network.
- A commit whose `Moai-Claim:` trailer is not in the last paragraph after `🗿 MoAI`, or a
  trailer paragraph with any line that is not a trailer.
- Deleting, overwriting or moving files without a preview, a receipt and a way back;
  following a symbolic link, including a linked parent folder, where the code promises not to.
- A path or pathspec handed to git and then trusted as a literal name without checking the
  entry git returned (for example `ls-tree -- :` lists every root entry).
- Text from input (titles, paths) placed into Markdown, HTML or a machine-read line without
  escaping, so it can split a table or plant a second machine line.
- A new command that changes another command's `--help` text.

## Nit (never blocks; report at most five)

- Wording that disagrees with the code it describes (module comments, rows in docs).
- An input the code drops without a `skipped` entry or a reason.

Formatting and lints belong to `cargo fmt --check` and `cargo clippy -D warnings`; do not
report them.

## Pre-existing

Report a defect that is already on the base branch as `pre_existing`, with a follow-up Issue
suggestion. It never blocks this pull request.

## Path-specific items

Items that depend on which paths the diff touches (tests, `skills/moai/**`, golden files,
`Cargo.lock`) arrive in the package's `checklist` key, from `tools/review_checklist.json`
at the base commit; they are not repeated here.
