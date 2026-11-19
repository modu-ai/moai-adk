---
id: SPEC-20261119-bundle-audit-tier
title: "moai bundle check reads the tier a PASS verdict states and reports a PASS below the audit floor"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-bundle-tool
---

# moai bundle check reads the tier a PASS verdict states and reports a PASS below the audit floor

## Intent

Card C-119. `moai bundle check` knows that a head has an audit floor (`gate::tier_of(Some(0), <paths of tip...head>)`: 0, 1 or 2) and that a `PASS` note names exactly the head, but it cannot tell at which tier the `PASS` was given: a floor-2 head is satisfied today by a tier-1 `PASS` (audit N-d of slice 1, the Risks of `SPEC-20261119-bundle-tool`, D-7 of `SPEC-20261119-bundle-ready-check`).

The leader decided on 2026-11-19 (card C-119 note 3113, D-7, reversible, category none): a verdict line states the tier it was audited at in the text before ` head `, as `AUDIT C-<n> (<SPEC>, tier <N>) head <sha> base <sha> — verdict: PASS`. The change is additive: a line without a tier and the pull request comment format keep reading, and the gate's verdict rules (`src/gate.rs`, `src/audit_gate.rs`) are not changed in this card; if they must change, that is a tier 2 slice of its own.

What this slice builds, in `bundle` only:

- `check` reads the tier a `PASS` note states.
- `check` reports `audit_tier_low` when every `PASS` that names the head states a tier and the highest is below the floor.
- `body` shows the stated tier in the heads table and refuses a head with `audit_tier_low` (it already refuses any finding of `check`).

## Non-goals

- Changing `gate::parse_verdict`, `src/gate.rs`, `src/audit_gate.rs` or the pull request comment grammar.
- Telling who wrote a verdict or whether the stated tier is true: the tier is what the line says (the independence of the auditor is the leader's judgement).
- Making a tier mandatory: a verdict without a tier counts as before.
- Writing the tier into a card note: the lane's and the leader's habit, not a tool.

## Requirements

- REQ-1: When `moai bundle check` reads a `PASS` verdict note that names a head, `moai bundle check` shall read the tier the note's first line states before its last ` head ` as the word `tier`, a space and one digit from 0 to 2 (`tier 1`), and show the highest stated tier of the head's `PASS` notes as `pass_tier`.
- REQ-2 (REQ-1): IF every `PASS` note that names the head states a tier and the highest is below the head's audit floor, THEN `moai bundle check` shall report the finding `audit_tier_low` with the floor, the stated tier and the note number.
- REQ-3: `moai bundle check` shall count a `PASS` note that states no tier, two different tiers or a digit above 2 as it does today, so that it satisfies the floor and gives no `pass_tier`.
- REQ-4: `moai bundle check` shall read a verdict line through `gate::parse_verdict` unchanged, so that a line with `tier <N>` reads there exactly as the same line without it.
- REQ-5: `moai bundle body` shall show the stated tier of a head's `PASS` note in the heads table and refuse a head that has the finding `audit_tier_low`.

## Acceptance

- AC-1 (REQ-1, REQ-3): the stated tier is read from before the last ` head `: `(SPEC-x, tier 1)` gives 1, `tier 2` at the end of the prefix gives 2, no word gives none, `tier 1 … tier 2` and `tier 3` give none, and a `tier 1` after the head's own SHA is not read.
  verify: `cargo test --locked --offline --lib bundle::tests::the_stated_tier_is_read_from_before_the_last_head`
- AC-2 (REQ-1, REQ-2): a floor-2 head (a change under `.claude/`, one of the gate's `TIER2_PREFIXES`) with a `PASS` that states `tier 1` has the finding `audit_tier_low` with floor 2, tier 1 and the note number; a floor-1 head with `tier 1` has none; a floor-1 head with `tier 0` has the finding; `pass_tier` is shown.
  verify: `cargo test --locked --offline --test bundle_check_cli a_pass_that_states_a_lower_tier_than_the_floor_is_a_finding`
- AC-3 (REQ-3): a `PASS` with no tier still satisfies a floor-2 head, a second `PASS` without a tier next to a `tier 1` one satisfies it, and a `tier 2` one next to a `tier 1` one satisfies it; `pass_tier` is the highest stated or null.
  verify: `cargo test --locked --offline --test bundle_check_cli a_pass_with_no_stated_tier_still_counts`
- AC-4 (REQ-4): `gate::parse_verdict` reads `AUDIT C-1 (SPEC-x, tier 1) head <sha> base <sha> — verdict: PASS` as the same triple as the line without `(…, tier 1)`, and `src/gate.rs` and `src/audit_gate.rs` are unchanged by this slice.
  verify: `cargo test --locked --offline --test bundle_check_cli the_gate_reads_a_verdict_line_with_a_tier_like_one_without`; `git diff --stat origin/v2026.11.0 -- src/gate.rs src/audit_gate.rs` is empty (read in the Result).
- AC-5 (REQ-5): the body's heads table shows `tier 1` in the audit cell of a head whose `PASS` states it, and `body` refuses a head with `audit_tier_low`, naming it.
  verify: `cargo test --locked --offline --test bundle_body_cli the_body_shows_the_stated_tier_and_refuses_a_low_one`

## Scope

- write: `src/bundle.rs`, `tests/bundle_check_cli.rs`, `tests/bundle_body_cli.rs`, `docs/design-review/appendix/operations/bundle.md`, `.moai/specs/SPEC-20261119-bundle-audit-tier/**`, `.moai/specs/SPEC-20261119-bundle-ready-check/**` (D-7 there becomes accepted), `.moai/specs/SPEC-20261119-bundle-tool/**`.
- read-only: `src/gate.rs`, `src/audit_gate.rs`, `src/ops/bundle.rs`.
- forbidden: `src/gate.rs`, `src/audit_gate.rs`, `src/lib.rs`, `src/main.rs`, any other change under `src/`, the real HOME and ledger as a write target, `skills/`, `.github/`, any push of a ref other than the branch of this slice's commits, any pull request.

## Plan

**Reading the tier (`stated_tier`).** For the first line of a `PASS` note that `gate::parse_verdict` reads, take the text before its last ` head ` (`parse_verdict` takes the verdict's head from the last ` head `, so the prefix is the same text). Find the word `tier ` that starts the prefix or follows a space, `(` or `,`, followed by one digit and then the end of the prefix or a space, `,`, `)` or `;`. One distinct digit from 0 to 2 across all such words is the stated tier; none, two different or a digit above 2 give none (REQ-3). A `tier N` after the verdict's head is not in the prefix and is not read.

**In `judge`.** Among the `PASS` notes that name exactly the head: `pass_tier` is the highest stated tier; the finding `audit_tier_low` {floor, pass_tier, note_id of the highest} is added when the floor is 1 or more, every such note states a tier and the highest is below the floor. A note without a stated tier ends the question: the head passes as before. The newest `PASS` stays `pass_note` as before. `audit_note` says that a stated tier is read and that the author is not verified.

**In `body`.** The audit cell reads `` `PASS` card note <id> (tier <N>): <line> `` when a tier is stated; `audit_tier_low` is a finding of `check`, so `body` already refuses it.

**How lanes and auditors write it.** `AUDIT C-<n> (<SPEC>, tier <N>) head <sha> base <sha> — verdict: PASS`; the documentation page says so. Nothing requires it.

## Risks

- A tier written by hand can be wrong or false: the tool shows it, it does not verify it. A floor-2 head with `tier 2` written by the lane that implemented it still needs the leader's judgement of who audited.
- A card note whose first line has `tier` for another reason (`(SPEC, tier-1 audit)`) is not read: the grammar needs `tier`, a space and a digit.
- Heads audited before this convention have no tier; they count as before, so the new finding only appears for lines that state a low tier.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The tier is read in `bundle` only, from the text before the last ` head `; `gate::parse_verdict` and the audit gate are unchanged | the leader's condition on D-7 (card note 3113): additive, no change to `src/gate.rs` or `src/audit_gate.rs` | a gate change as a tier 2 slice |
| D-2 | A note without a stated tier satisfies the floor, and one such note among several `PASS` notes ends the question | the 63 readable verdict notes of the ledger have no tier; the finding must not appear for them | require a tier |
| D-3 | The highest stated tier counts, not the newest | a lane may repair after a tier 2 audit with a tier 1 delta; the tier-2 evidence on the exact head stays | newest |
| D-4 | Two different tiers in one line, a digit above 2 and a `tier` word after the head's SHA are not read | an ambiguous line must not produce a finding or a pass | refuse the line |

## Result

AC-1 … AC-5: pending — the SPEC is written first; each AC gets its outcome line when the code, the tests and the checks of this slice have run.
