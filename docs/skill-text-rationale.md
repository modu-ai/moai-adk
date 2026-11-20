# Why the shipped skill text lost its citations

SPEC-20261120-skill-text-citations. The files that ship to users (`skills/`, `agents/`, `templates/`, `install/`) carry no card number, no Issue or pull-request number, no citation of a user or leader decision and no name of a moai-adk development script: a user of MoAI cannot follow them. `tests/skill_text_citations.rs` keeps them out and lists, in `PENDING`, the shipped files that still carry some. This document keeps what was removed from each cleaned file, so a developer can still follow a rule back to its source. Issue and pull-request numbers belong to `modu-ai/moai-adk`; card numbers to its card ledger.

Rule of the cleaning: every rule a file states stays; only the citation, the sentence that tells where a rule came from, or the development script name goes. A citation that a test of another change still pins stays in its file for now, and the file stays on `PENDING`.

## In the moai-adk repository

The shipped text no longer names these commands, because a user's repository does not have them. A session working in the moai-adk repository uses them as before:

- `tools/build_folder.py env` (run with `python3 -I`): `eval` its two `export` lines only when it exits 0; it also writes the build-folder record and makes `~/.moai` and `~/.moai/build` (mode 0700) when they are missing. `moai lane-env --lane <lane>` sets the same folder and record, plus the lane's `CARGO_HOME`.
- `tools/build_folder.py clean` (run with `python3 -I`) deletes the folder's cargo output and keeps any other file.
- `tools/build_folder.py remove` deletes the picked rows only when a fresh preview still marks each one `trim`, and refuses all of them otherwise.
- `tools/source_manifest.py write`, then `tools/source_manifest.py check` (both run with `python3 -I`), after a change under `skills/moai/`; a new file needs its `origin` by hand.

## `skills/moai/SKILL.md`

- Removed `(user decision 2026-11-10, #2286)` after the two mode flags.
- Removed `(Issue #2515)` after the round-end sentence.
- Removed `Issue #2592:` in the delegation sentence.
- Removed `(Issue #2619)` after the question order.
- Removed `card C-81:` in the tag rule.
- Removed `(Issue #3005 N6: five such questions got no answer and a leader waited seven turns, then marked its goal `blocked`)`, the story behind the rule on non-blocking question tools.
- Removed `(Issue #2678)` after the report rule.

## `skills/moai/references/adapters/aside.md`

- Removed `(user decision, Issue #2634 comment 6845653775)` in the CLI-only rule.

## `skills/moai/references/adapters/orca.md`

- Removed `(user decision, Issue #2634 comment 6845653775)` in the CLI-only rule.
- Removed `(Issue #2899 G2, G4)` after the clear contract.
- Removed `(Issue #3005 AC-7, AC-8)` in the optional-Orca sentence.
- Removed `Issue #3005, or in the Issue its cell names` in the table legend.
- Removed `tracked in Issue #3005` in the unsupported paragraph.
- Removed `Settled in Issue #3005 P10 and P11` in the unsupported paragraph.

## `skills/moai/references/internal/build-folder.md`

- Removed `User decision 2026-11-17 (card C-82, "빌드 폴더를 한 곳으로 모으기"):` at the start of the file.
- Removed item 4's repository command `tools/build_folder.py env` (run with `python3 -I`; `eval` its two `export` lines only when it exits 0; it also writes the record and makes `~/.moai` and `~/.moai/build`, mode 0700, when they are missing); the commands are listed above.
- Removed item 7's repository command `tools/build_folder.py clean` (run with `python3 -I`; it keeps any other file); the commands are listed above.
- Removed item 8's repository command `tools/build_folder.py remove` (it deletes the picked rows only when a fresh preview still marks each one `trim`, and refuses all of them otherwise); the commands are listed above.
- Removed `as tools/build_folder.py env does` in item 9.
- Added, after the audit: item 4 says a session outside a lane sets the two variables itself; item 8 says a picked folder that no longer shows `trim` is left alone (the repository command refused those rows); item 9 gives the form `e=$(moai lane-env --lane <lane>) && eval "$e"`, taken from the lane-env hint in `src/heavy_slot.rs`, which keeps the old condition of item 4 that its output is evaluated only when the command exits 0.

## `skills/moai/references/internal/commit-trailers.md`

- Removed `(user decision U-B, 2026-11-17)` and the history sentence `They replace the planned Moai-Claim trailer`.
- Removed `(leader decision 2026-11-17, SDD-6 re-audit B3)` in item 3.

## `skills/moai/references/internal/git-safety.md`

- Removed `#2811` after `moai worktree-plan`.
- Removed `(user decision, Issue #2673)` in item 7.
- Removed `(#2876, from the Go generation's main-checkout branch guard)` in item 8.
- Removed `(Issue #2673 comment 6849413590)` after the measured version.

## `skills/moai/references/internal/review.md`

- Removed `(#2876, from the Go generation's security decision contract)` in item 6.

## `skills/moai/workflows/optimize.md`

- Removed `(leader decision on Issue #2724)` in the first paragraph.
- Removed `(card C-82)` in step 3.
- Removed `(Issue #2930 item 6)` in step 5.

## `skills/moai/workflows/read.md`

- Removed `#2592` in the Kiro and Cursor CLI line.

## `skills/moai/workflows/run.md`

- Removed `(20 such update-branches on 2026-11-13, Issue #2401 comment 6864202133)` after the shared catch-all test rule.
- Removed `(#2876, from the Go generation's coding standards and CI repair protocol)` after the CI repair rule.

## `skills/moai/workflows/setup.md`

- Removed `(user decision 2026-11-16, cards C-51 and C-52)` after the ledger sentence of step 2.
- Removed `(Issue #2817)` at the end of step 2.
- Removed `(the external skill install operations were removed, Issue #2286, user decision 6774899518 row 35)` in step 3.
- Removed `(user decision 2026-11-16, card C-51)` in step 8.

## `skills/moai/workflows/sync.md`

- Removed `(user decision 2026-11-10, #2384)` after the forge-only ledger rule.
- Removed `(Issue #2388; integration gate of user decision U6, 2026-11-12, Issue #2515)` in the merge-train heading.
- Removed `; decision D7, Issue #2554 comment 6860022304` in the merge-train shared-file exception.
