# Jev Negative Results

Two records the product carries so a future reader finds them before
re-litigating them: the measured rejection of the premise-death use of the
judgment capability, and the disposition of the `scripts/jev/` working copy.
This file is tracked but deliberately not template content — it sits outside
the template tree, so it may carry the measured figures the template
neutrality rules forbid elsewhere.

## 1. Premise-death judgment — measured and rejected

Using the judgment capability to decide whether a stale backlog card's
premise is dead (so the card could be skipped before dispatch) was measured
and rejected. A wrong "dead" marking silently skips a card and leaves no
signal, which makes this the expensive direction — and the measurement put
the model barely above coin-flip on exactly that direction:

| Figure | Measured value | Reference |
|---|---|---|
| `premise_dead` call precision | 14/48 = **29.2%** | 25% base rate — barely above always answering "dead" at random |
| 2-class accuracy (dead / not-dead) | **58.9%** | 75.0% for the constant "always alive" answer — the constant wins |
| English control (40 cards, 2-class) | **67.5%** | still below the 75.0% constant |
| Confidence-threshold sweep | flat across 0.30–0.80 | raising the gate lowered adoption, not errors |

The interpretation is the load-bearing part, not the arithmetic: the
capability's answers were more accurate than chance and still less useful
than a one-line constant, because the decision's cost is asymmetric. A card
wrongly left alive costs one dispatch; a card wrongly marked dead costs the
work nobody did and nobody noticed. An instrument in the 29% precision range
on the expensive direction cannot be promoted into that decision, and the
sweep shows no confidence threshold rescues it.

`moai todo triage` stays model-free: no requirement in this chain routes a
model answer into it, and this record exists in part so that stays a
settled decision rather than an unmade one.

**Provenance — cited, not re-measured.** The figures above come from the
2026-09-22 measurement round. The full measurement record (per-card tables,
gate-by-gate precision, sample construction, and the translation control's
sampling limits) lives in gitignored local evidence:
`.moai/reports/t943/verdict.md` in the development checkout — a file any
given clone may not carry — and in the maintainer's local development guide
(§30). This file cites those origins and re-measures nothing. Re-running the
experiment needs a NEW reason (a changed decision shape, a changed cost
model, a changed instrument), not a new attempt: the gate sweep was already
flat and the English arm was already measured.

## 2. `scripts/jev/` disposition

The `scripts/jev/` directory (`triage.sh`, `route.sh`, `ask.sh`) is an
**uncommitted working copy** in the development checkout:

- **Not committed** — it appears in no branch of this repository's history.
- **Not distributed** — it has no template mirror, so `moai init` and
  `moai update` never ship it to a user project; it reached no user at any
  point.
- **Not maintained** — it is superseded by the Go package (`internal/jev`),
  which is the canonical implementation of the capability and the one thing
  the tests cover.

The scripts' behaviour (a triage pre-check before dispatching a stale card,
routing a blocked question to a decision grade) is available through the
binary's own surfaces and the gated MCP tool. Nothing in this disposition
deletes anything from anyone's working tree: the working copy remains where
it is until its owner removes it.
