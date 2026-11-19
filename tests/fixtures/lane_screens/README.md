# Lane screens (SPEC-20261119-lane-screen-guard, card C-120)

Every file named `synthetic-*.txt` is a screen written by hand in the shape of a Claude Code lane's rendered screen. No file holds text from a real session: nothing was recorded (decision D-5, leader note 3114: no confirmation prompt or usage-limit screen was on display to record).

Two facts that were seen for real are in the tests as given:

1. A glm lane's usage-limit screen reads `Your limit will reset at 2026-11-19 17:30:55` with **no time zone**, and the zone was in fact UTC+8. A time without a zone is therefore "unknown" together with its text, never read as local time or as UTC (`synthetic-usage-limit-glm.txt`).
2. A narrow window can leave a tail of three rows that shows only `esc to interrupt`; with too few rows the state is "unknown" (`synthetic-narrow-tail.txt`).

A confirmation prompt is told apart by a numbered `Yes` row together with a numbered `No` row; the `Do you want` question is not required. Three synthetic screens guard against such a prompt being read as `working` or `idle` (audit note 3212, B1, found each of them read that way): `synthetic-ascii-cursor-dialog.txt` (the cursor is `>`), `synthetic-bordered-dialog.txt` (the dialog is inside a vertical frame) and `synthetic-no-cursor-dialog.txt` (no cursor mark and no `Do you want`, above an idle input line).

A numbered `Yes` row together with a numbered `No` row is a confirmation whatever stands before the numbers (audit note 3362, B1; the reader skips every mark before the first letter or digit): `synthetic-foreign-cursor-no-question-working.txt` (`◆ 1. Yes` and `   2. No` under a working footer, no question row at all), `synthetic-foreign-cursor-far-question-working.txt` (the same under a question sentence none of the known question words is in) and `synthetic-unreadable-choice-working.txt` (`» 1. Yes` and `   2. No` under a working footer). Each of them once read as `working`, so a text and Enter were sent.

A question such as `Do you want …` in the last 8 non-blank rows with a `yes` or `no` word on a row below it, whose choices cannot be read as numbered rows, makes the screen `unknown` (leader R-N1, note 3256, narrowed by S-N2, note 3362): `synthetic-unreadable-choice-idle.txt` (`[1] Yes` and `[2] No` above an idle input line; the number is closed by `]`). A lane that ended its turn with a question and waits at its input line is not: `synthetic-question-then-input-idle.txt` is `idle`.

Replace a synthetic file with a recorded one when a real screen is at hand.
