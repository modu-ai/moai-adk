---
id: SPEC-20261123-lane-rules-review
title: "Lane rules from the 2026-11-23 review: clear after every card, one card per lane, no Python test by hand, a real run of the dangerous path, a whole-suite run every five landings"
status: draft
revision: 2
created: 2026-11-23
---

# Lane rules from the 2026-11-23 review: clear after every card, one card per lane, no Python test by hand, a real run of the dangerous path, a whole-suite run every five landings

## Intent

The lane review of 2026-11-23 measured 19 FAIL in 65 verdicts with every lane check green, 187 block notes in 48 hours, seven of ten lanes waiting on one receipt check of about 440 test binaries, one lane idle 19 hours after its card, two lanes at 94 to 95 percent context, and Python tests still run by hand after the Rust-only rule. The user's instructions of that day become five rules in `CLAUDE.md`, each with its provenance row, so that every session loads them. The skill workflow text follows in a later card. Revision 2 adds a sixth rule of the same day: every instruction to a lane is written in the user's conversation language, and only subagent prompts are English, because the leader had sent lane dispatches in English for a day.

## Requirements

- REQ-1: `## Leader supervision` in `CLAUDE.md` shall state that after every finished card of a lane the leader checks the ledger and reads the lane screen and sends `/clear` through `orca terminal send` and then the next line and that a lane ends a finished card with `LANE-DONE` as its last line.
- REQ-2: `## Agent loop` in `CLAUDE.md` shall state that a lane works one card at a time and runs one receipt check at a time and starts no follow-up slice for as long as another card of the lane is open.
- REQ-3: `CLAUDE.md` shall say that tests are Rust only and that nobody runs a Python test module by hand: the check set and the scan item carry no such command and the item names the Rust-only tests card by its title as the place the remaining Python step is removed.
- REQ-4: `## Agent loop` in `CLAUDE.md` shall state that a change that deletes or resets or starts a process or builds a permission rule carries in its READY a run of the dangerous path in a throwaway.
- REQ-5: `## Leader supervision` in `CLAUDE.md` shall state that every five landings the leader runs every Rust test on the integration branch tip once and that a red test stops landings until it is fixed.
- REQ-6: `## Rule provenance` in `CLAUDE.md` shall hold one row per new rule with the user instruction date 2026-11-23 and its cause.
- REQ-7: `CLAUDE.md` shall stay under its 44 KiB budget of 45,056 bytes.
- REQ-8: The new items and the new rows in `CLAUDE.md` shall name no card number.
- REQ-9: The Writing language item in `CLAUDE.md` shall state that every instruction to a lane is in the user's conversation language and that only subagent prompts are English.

## Acceptance

- AC-1 (REQ-1): `CLAUDE.md` has an item that starts `- Clear after every card` in `## Leader supervision`, and the hand-off item no longer says `start the next round at once`.
  verify: `grep -n "Clear after every card" CLAUDE.md` and `grep -c "start the next round at once" CLAUDE.md`
- AC-2 (REQ-2): `CLAUDE.md` has an item that starts `- One card at a time` in `## Agent loop`.
  verify: `grep -n "One card at a time" CLAUDE.md`
- AC-3 (REQ-3): `CLAUDE.md` holds neither `python3 -m unittest discover -s tests` nor `tests.repo_scans_test tests.source_manifest_test`, and its `- Tests are Rust only` item names the Rust-only tests card by its title.
  verify: `grep -c -e "python3 -m unittest discover -s tests" -e "tests.repo_scans_test tests.source_manifest_test" CLAUDE.md` and `grep -n "Tests are Rust only" CLAUDE.md`
- AC-4 (REQ-4): `CLAUDE.md` has an item that starts `- Dangerous path` in `## Agent loop`.
  verify: `grep -n "Dangerous path" CLAUDE.md`
- AC-5 (REQ-5): `CLAUDE.md` has an item that starts `- Whole suite every five landings` in `## Leader supervision`.
  verify: `grep -n "Whole suite every five landings" CLAUDE.md`
- AC-6 (REQ-6): `## Rule provenance` holds six new rows, each with `2026-11-23` and `cause:` and two cells.
  verify: `grep -c "^| .*2026-11-23.*cause:" CLAUDE.md`
- AC-7 (REQ-7): `CLAUDE.md` is under 45,056 bytes, and the two tests that read it pass.
  verify: `wc -c CLAUDE.md` and `cargo nextest run --locked --offline --test repo_claude_md --test recurring_problems_rule`
- AC-8 (REQ-8): the ten new lines of `CLAUDE.md` name no card number.
  verify: `grep -n -e "Clear after every card" -e "clear after every card" -e "One card at a time" -e "one card and one receipt" -e "Tests are Rust only" -e "tests are Rust only" -e "Dangerous path" -e "dangerous path" -e "Whole suite every five landings" -e "every five landings" CLAUDE.md | grep -c -E "(^|[^A-Za-z0-9])C-[0-9]"`
- AC-9 (REQ-9): the Writing language item holds the sentence on lane instructions, and no line of `CLAUDE.md` lists cross-session messages as English.
  verify: `grep -c "Every instruction to a lane" CLAUDE.md` and `grep -c "cross-session messages, subagent prompts" CLAUDE.md`

## Scope

- write: `CLAUDE.md`
- write: `.moai/specs/SPEC-20261123-lane-rules-review/spec.md`

## Plan

1. Add two items to `## Leader supervision` and two items to `## Agent loop`, replace the scan-tests-first item with the Rust-only item, and drop the unittest line from the check set.
2. Let the hand-off item end in `LANE-DONE` as the last line, in place of "start the next round at once", which the new rule contradicts.
3. Add five rows to `## Rule provenance`, each with the instruction date and its cause.
4. Run `moai spec check`, commit once with the SPEC trailers, run `moai trailer-check`, then one `moai check affected` on the committed head.

## Result

Measured with the `verify` commands above: the text checks on the working tree before the first commit, and the tests of AC-7 on the tree of that commit.

- AC-1 met: `CLAUDE.md` line 141 starts the item and names the ledger check, the screen read, `/clear` through `orca terminal send`, the next line and `LANE-DONE` as the last line; the count of `start the next round at once` is 0.
- AC-2 met: `CLAUDE.md` line 106.
- AC-3 met: the count of the two Python strings is 0, and line 108 names the Rust-only tests card by its title. `unittest` stays in two older lines, the audit caps item and its provenance row, both as a cap that says no.
- AC-4 met: `CLAUDE.md` line 107.
- AC-5 met: `CLAUDE.md` line 142.
- AC-6 met: six rows with `2026-11-23` and `cause:`, each with two cells (five in revision 1, the language row in revision 2).
- AC-7 met: `wc -c CLAUDE.md` prints 43851, which is 1205 under 45056 (the file had 41913). `moai check affected` picks neither of the two targets for a change of `CLAUDE.md` (its 43 targets include `tests/next_flow_skill_scan.rs`), so the verify line ran once by hand in a heavy-run slot: 10 tests run, 10 passed, 0 skipped.
- AC-8 met: the count of card numbers in the ten new lines is 0.
- AC-9 met (revision 2): `Every instruction to a lane` counts 1 and `cross-session messages, subagent prompts` counts 0; the six-row count of AC-6 is 6, and `wc -c CLAUDE.md` prints 44900, 156 under 45056.
