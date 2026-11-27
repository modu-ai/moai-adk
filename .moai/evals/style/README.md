# Output style eval: translationese in a language other than English

This folder grades the wording rule of the output style (`skills/moai/references/home/output-style.md`):
in a language other than English the writer uses the standard wording working developers of that
language use, and rereads the reply to polish what reads as translated. The style text holds the
principle and the polish step only; it holds no word list (user direction, ledger C-592 note 10532).
The pairs that show what the polish should change live here, outside the skill package and outside
its manifest, so a word list never reaches the model through the style.

## Files

- `cases.json` — pairs of a sentence in translationese (`translationese`) and the polished
  sentence (`polished`), with the language (`language`) and the `source` that confirms the
  polished wording. Today the file holds Korean pairs only: each one cites the user's decision in
  the card ledger. A pair for another language is added only after a native speaker or an official
  source has confirmed its wording, and its `source` says which.
- `controls.json` — sentences that must come back unchanged (an identifier, a loanword developers
  already say, an English term that is the standard wording, a protocol line). A rewritten control
  is over-polishing.

## Running and scoring

No `moai` command runs a model against these files yet (AC-14 of
`SPEC-20261127-output-style-moai-style` is `unknown` until one exists). Until then, give a
reviewer a request in the language of a case that carries the `translationese` sentence, without
the `polished` sentence, and compare the reply with it: it passes when the reply reads like the
`polished` sentence in wording, and a control passes when it comes back as written. The count is a
ruler for comparing runs, not a truth.

`tests/home_style_ownership_cli.rs` keeps the style text free of every `translationese` sentence
here and checks the shape of both files.

## Adding a case

Write the pair yourself from a real correction and name its source. Never add a pair from memory of
how a language "usually" sounds; an unconfirmed pair is worse than none.
