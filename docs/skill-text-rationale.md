# Why the shipped skill text lost its citations

SPEC-20261120-skill-text-citations. The files that ship to users (`skills/`, `agents/`, `templates/`, `install/`) carry no card number, no Issue or pull-request number, no citation of a user or leader decision and no name of a moai-adk development script: a user of MoAI cannot follow them. `tests/skill_text_citations.rs` keeps them out and lists, in `PENDING`, the shipped files that still carry some. This document keeps what was removed from each cleaned file, so a developer can still follow a rule back to its source. Issue and pull-request numbers belong to `modu-ai/moai-adk`; card numbers to its card ledger.

Rule of the cleaning: every rule a file states stays; only the citation, the sentence that tells where a rule came from, or the development script name goes. A citation that a test of another change still pins stays in its file for now, and the file stays on `PENDING`.

## In the moai-adk repository

The shipped text no longer names these commands, because a user's repository does not have them. A session working in the moai-adk repository uses them as before:

- `tools/build_folder.py env` (run with `python3 -I`): `eval` its two `export` lines only when it exits 0; it also writes the build-folder record and makes `~/.moai` and `~/.moai/build` (mode 0700) when they are missing. `moai lane-env --lane <lane>` sets the same folder and record, plus the lane's `CARGO_HOME`.
- `tools/build_folder.py clean` (run with `python3 -I`) deletes the folder's cargo output and keeps any other file.
- `tools/build_folder.py remove` deletes the picked rows only when a fresh preview still marks each one `trim`, and refuses all of them otherwise.
- `tools/source_manifest.py write`, then `tools/source_manifest.py check` (both run with `python3 -I`), after a change under `skills/moai/`; a new file needs its `origin` by hand.

- `tools/smoke/run.sh <area>` (`<area>` one of `install`, `update`, `project`, `launcher` or `all`; `--bin PATH` names another binary): boots the built binary in an empty temporary HOME and a temporary git project and never reads the real HOME. Its last line, `smoke: <area> pass|fail steps=<n>`, is the evidence of the check: a `fail` is a failed check, a missing binary is a usage error (exit 2) and not a pass, and a smoke run that was not made is `not_run`. Run it before the first commit of a change that touches install, update, `/moai project` or the launcher.
- `tools/ci_scope.py`: the repository's own CI script that classifies a pull request as docs-only; it runs before the Rust toolchain is installed and therefore stays a script. Its verified-safe list is `docs/` with `.md`, `.txt`, `.png` and `.svg` suffixes (minus files tests read) plus `.moai-cli/reports/`. Its own tests carry the old and the new verdict over recent pull request bodies when it changes.
- `tools/heavy_check_slot.py`: uses the same slot folder and lock as `moai heavy-slot-acquire`, so both take turns in one queue. `tools/self_check.sh --full-python` and `tools/local_check.sh` wrap the python slot when `moai heavy-slot-acquire` is installed and run the tests directly otherwise.
- `tools/source_manifest.py write` also resolves a conflict in `source-manifest.json`: it joins both sides three-way against the common version, recomputes every sha256 and stops (exit 2, nothing written) when a person must choose.

## `agents/audit.md`

- Removed `(SPEC-20261119-check-guard)` after "a pull request that weakens either file is still judged by the base text"; the base-text rule itself stays.

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

## `skills/moai/workflows/audit.md`

- Removed the sources of the CI rule (`decision D4, Issue #1975 comment 6840449133`), `(Issue #2680)` after the gate-unavailable rule and `(leader decision 6849147698)` after the substitute-evidence rule.
- Removed the sources of the verdict-line rules: `(Issue #2401 comment 6864204207)`, `(Issue #2833 D4)` and `(audit 6869947940 B1, #2821)`.
- Removed `card C-84` after the gate's rule on the head's CI, and `(user decisions 6877084436 and 6878414880, Issue #1975)` after the rule that an author edits the tier line.
- Removed `(Issue #2401 comment 6824103310)` after the rule on citing a run only when it names the tree under audit, `(Issue #3054)` before the cleanup rule of the agent brief and `(Issue #3052)` after the rule that only a repository member's comment counts.
- The sample verdict line now reads `AUDIT #12 (repo-scans, part 4 of 4) …` with the lane line `감사: agent-27 (예약 12)`; before, it carried a real pull request number, the number of a design Issue and a real comment id. The format, the markers `## 독립 review`, `수리 확인 범위` and `새 지적 분류`, and `moai-audit-claim` are unchanged.
- The follow-up classification sample `후속 Issue #123: concrete reason` now reads `후속 Issue #<number>: concrete reason`; the sentence after it already says to replace the number.
- The sentence "The #2388 marker repair preserves the reported audit inputs unchanged:" now reads "The marker check keeps rejecting these inputs, which it once let through:"; the table of the four inputs is unchanged.

## `skills/moai/workflows/mode-agent.md`

- Removed the Issue, decision and card labels that followed rules, as the diff shows them: Issues 1975 (four comment references), 2345, 2390, 2401 (four), 2515 (six), 2516, 2554, 2581, 2592 (two), 2612 (four), 2619, 2664, 2668, 2677, 2679, 2836 (two), 2840, 2859, 2879 (two), 2880, 2899 (plain, D-O2, G2, G3 twice, G5), 2903, 2904, 2933 (three), 3005 (four plain references and F3, F4 and F5, F9, N4, N5, N6, N11, N17, P3b), 3052; the numbers 3029 and 3040 (an audit of it); `decision D4`, `D6`, `D7`, `L3` (twice), `L5`, `L7`; `user decision U6`, `D-O1`, `D-O2`, `2026-11-10`, `2026-11-16` and `2026-11-17` (twice); two `leader decision` references with comment ids; and `card C-53` (with `C-53 S3`, twice, and `C-53 S6`), `card C-99`, `C-84`. "after S3" now reads "after the gate reads card branches"; "Until C-53 S3" now reads "Until the gate reads card branches".
- The sample line `… next=card C-62` of the session-id step now reads `… next=card C-<n>`.
- The section "Session control with Orca" lost the same labels as in `mode-leader.md` (the two files carry the section word for word, and a test checks that).
- Removed the repository's own scripts from the `heavy-checks` step: `tools/self_check.sh` ran the machine items "in one go" and, with `--pre-ready`, added the whole-tree list; `tools/self_check.sh --full-python` and `tools/local_check.sh` wrapped the python slot. The text keeps the rules: run the whole-tree list (the SPEC check, the repository scans, the manifest tests, the skill package test) before reporting the head ready, run a whole Python suite only through the python slot (the hand form `moai heavy-slot-acquire …` stays), and run the project's own local check before opening a pull request.
- Removed from the `pr-gate` step the moai-adk-only pre-push examples `python3 -m unittest tests.ci_scope_test` (when a test now reads a file under `docs/`) and `cargo test --offline --locked --test ac_command_lint` (when a test now runs `git`); the text says to run the project's checks that caught pull requests only in CI, and keeps `git interpret-trailers --parse`. A session working in the moai-adk repository still runs those two commands by hand.

- Added, by the user's decisions of 2026-11-20 (the wording was replaced three times that day; the last is a six-item rule): the local finish line (1 the lane commits on its own card branch, pushes and gets the independent audit; 2 after the PASS it makes the merge onto the local tip in its own worktree, runs the check-affected command of the tool built there and posts `MERGE-READY`; 3 one lane at a time under a lock the primary checkout is fast-forwarded with `git merge --ff-only`, by the leader's supervision round where the host isolates lane sessions, or by the one allowed guard command `git -C <primary> merge --ff-only refs/moai/lane-merge`; 4 builds for local testing and installation come from the verified local tip, the leader checks the whole tip from time to time and repairs it with a revert commit, and the status line shows commits ahead and behind; 5 only for a real deployment the leader creates a release branch and one pull request; 6 a card that landed locally is closed with `card done` on its landing evidence). The same paragraph stands once in `pr-gate` of `mode-leader.md`, `merge-train` there points to it, and `verify.md` names the release pull request's CI. The decisions called the flow by moai-adk's internal name, cited themselves with a date and named the repository's branch; the shipped text uses none of the three (the scan forbids them), and the date is kept here. `card done` does not yet accept a card without a merged pull request, which the shipped text says in one clause.

## `skills/moai/workflows/mode-leader.md`

- Removed the source labels of the decision, Issue and card numbers: the ledger switch (`user decision 2026-11-16, card C-53`), `C-53 S3` ("until C-53 S3 lands" now reads "until the gate reads card branches") and `C-53 S6`, `decision L3`, `decision L5`, `decision L6` and `decision L7`, `decision D5`, `decision D7` ("decision D7 … sets it to eight" now reads "the recorded decision sets it to eight") and `decision D4`, `user decision D-2`, `user decision D-O3 of 2026-11-14` with its replaced `D-3`, `user decision U6 (2026-11-12, ruleset 26791615)`, `user decision 2026-11-17` (twice) and the reference `user decision 2026-11-17 "CI 성공 + PASS"` with the two decisions it replaced, `leader decision 2026-11-17`, `decision 2026-11-16`, and the two `user instruction of 2026-11-15` with a comment id.
- Removed the Issue and pull request numbers that followed rules, as the diff shows them: Issues 1975 (four comment references), 2385, 2388, 2515 (three times), 2516, 2612 (twice), 2619, 2664, 2665 (twice), 2668, 2678, 2679, 2680, 2681, 2879, 2899 (G1, G2, G3 twice, G4), 2902, 2905, 2906, 2928, 2933 (twice), 3005 (four plain references and F1, F2, F3, F4, F6, N1, N3, N4, N5, N6, N8, N12, N14, N15, N17, P3a), 3015, 3052; the numbers 2401 (three times, two of them as `ledger row` 23, 31 and 33), 2693, 2708, 2777 and 3029; and `card C-81`, `card C-84` and `card C-99`.
- Removed the stories that came with some of them: the count of 35 dispatches and 47 startable cards (dispatches must name a card); a leader that set its goal `blocked` and stayed stopped for hours (waiting is not blocked); the 14 pull requests merged without `moai gate-check` (a merge that skipped the gate is still named in the round summary); the compacted leader that ran a 9.7-hour-old instruction again (the section still says to check first); seven hours of mail without a Run (a mail to a `term_…` handle is still not kept for a lane that is not checking); the Codex process whose tool shells dropped `ORCA_*` (the sentence keeps the fact).
- Changed wording beyond a removed parenthesis: "…under the integration gate of user decision U6 (…)" became "…under the integration gate"; "Audits (user decision 2026-11-17):" became "Audits:"; "Check execution and CI citation rules follow Issue #2385 and are not restated here." became "…follow the VERIFY workflow and are not restated here."; the measurement note for a missing `ORCA_TERMINAL_HANDLE` lost "by the audit of #3029", Orca's comment id and its Issue.

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

## `skills/moai/workflows/verify.md`

- Removed the paragraph that told to run `tools/smoke/run.sh <area>` before the first commit of a change that touches install, update, `/moai project` or the launcher; the command is described above ("In the moai-adk repository").

The removed paragraph, word for word (the smoke harness test reads it here; a session working in the moai-adk repository follows it by hand):

In the moai-adk repository, before the first commit of a change that touches install, update, `/moai project` or the launcher, run the smoke harness first: `tools/smoke/run.sh <area>`, with `<area>` one of `install`, `update`, `project`, `launcher` or `all`, on the debug binary just built (`--bin PATH` names another one). It boots the built binary in an empty temporary HOME and a temporary git project and never reads the real HOME. Take its last line, `smoke: <area> pass|fail steps=<n>`, as the evidence of the check; a `fail` is a failed check, a missing binary is a usage error (exit 2) and not a pass, and a smoke run that was not made is `not_run`, never `passed`.
- Removed the sources of the check-run rules: `(#2385)`, `(#2388)`, `#2380`, `(leader decision 2026-11-10 on the #2395 question)`, `(decision D4, Issue #1975 comment 6840449133)`, `(user decision U10, Issue #2554)`, `(Issue #2563)` (the old workflow name `rc-ci` stays, now as "the workflow was once named `rc-ci`") and `(SPEC-20261117-supported-platforms)`.
- Removed the sources of the build-folder and slot rules: `(Issue #2841, card C-82)`, `(Issue #2879 D3 — a warning, not a refusal)` ("a warning, not a refusal" stays), `(leader decision, Issue #2677 comment 6859259006, made per worktree by card C-82)`, the date and comment of the two worktrees that ran the other tree's binary (`2026-11-13`, `Issue #2677 comment 6851241454`; the observation of 7 of 8 tests stays), `the leader's canonical wording (Issue #2385 comment 6777290466; user decision, Issue #2389 comment 6774899518)` ("Build-folder cleanup follows these rules:"), `(Issue #2879, user instruction of 2026-11-14)`, `Issue #2879 D4`, `(the leader's numbers, Issue #2677 comment 6850581277)`, `(card C-88)`, `(Issue #2930 item 6; …)`, `Issue #2930, card C-88`, `(Issue #2677; …, Issue #2930)`, `(Issues #2491, #2677)` and `Issue #2934` ("is the job of `worktree-plan` and `worktree-move-plan`").
- Removed the sentences about the repository's own scripts: `tools/self_check.sh --full-python` and `tools/local_check.sh` wrapping the python slot, `tools/heavy_check_slot.py` sharing the slot folder, and `tools/ci_scope.py` as the CI scope classifier with its verified-safe list (the text now says the project's own CI script decides, and keeps a usual list); "Both are unix only" now reads "The slot commands are unix only". The example hook "a Go-generation `pre-push` that runs `make ci-local`" now reads "a `pre-push` that runs the project's whole CI target".
- Removed the sources of the test rules: `(Issue #2666)`, `(card C-99)` after "Record home", `(Issue #2661)`, `(`ci_install_ripgrep_test` #2469, `mutate_test` #2502)` and `(#2527, design comment 6822993459)`.
- Removed the `WF-OBS #n` labels in front of the change-specific and shell rules (9, 25, 26, 29, 30 and 17, 19, 16) and the Issue numbers `(Issue #2836)` and `(Issue #2837)`; the manifest rule no longer names the repository's `write` command (see above). The section "Lane worktrees and build cache" lost `(Issue #2930, items 3 and 5, R-5; decision D-f)` and `(user decision 2026-11-17, card C-82)`.
- Added the `M-Flow finish line` section before `## Procedure`: it is the section of `sync.md` word for word (the paired change wrote it once for every file that tells how a change ends, and its test reads each of those files for the same ten sentences), without the sentence about steps 5 to 8 of the sync procedure, which does not apply to this file. `M-Flow` is a product flow name here, not a development word.

## `skills/moai/workflows/sync.md`

- Removed `(user decision 2026-11-10, #2384)` after the forge-only ledger rule.
- Removed `(Issue #2388; integration gate of user decision U6, 2026-11-12, Issue #2515)` in the merge-train heading.
- Removed `; decision D7, Issue #2554 comment 6860022304` in the merge-train shared-file exception.
