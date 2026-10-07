---
description: The measured values of the card-issuance rule — local-only, never distributed. Each value cites the baseline figure row it was re-derived from; re-measure per project before trusting any of them elsewhere.
paths: "**/.claude/skills/moai/workflows/gtd.md,**/.claude/agents/moai/manager-todo.md"
---

# Card Issuance Thresholds (local-only)

> Values only, by design — the mechanism lives in `card-issuance.md`. Every
> value below was re-derived from THIS repository's measured queue
> distribution; the citation names the baseline figure that grounds it. A
> different project measures its own distribution and sets its own numbers.

| # | Item | Value | Baseline grounding |
|---|------|-------|--------------------|
| 1 | Similar-card ceiling (presented matches per add) | 3 | card-text-fixed value |
| 2 | Display floor (token-set Jaccard) | 0.30 — alerts 8.1% of all pairs, 4.1% of the recent window; 5% recall over recorded relations | SB03, SB05 |
| 3 | Component key depth | 2 path segments — depth-3 shared pairs measured 0 | SB07 |
| 4 | Card size lower-bound trigger | < 50 estimated product lines AND ≤ 3 product files (29% of the measured queue) | GB09 |
| 5 | Card size upper bound | > 1799 product lines OR > 22 product files (the new P90) | GB03, GB04 |
| 6 | Derivation depth ceiling | 2 — depth ≥ 3 measured 74 cards (6.0%) | QB06 |
| 7 | Concurrent-progress ceiling | 16 in flight — the 3-hour-window P90 (median 5, max 41) | QB10 |
| 8 | Hub threshold | max 72h-window overlap ≥ 10 — the five hub files measured 20/15/10/10/10 | GB15 |
| 9 | Git probe time budget | carried by the probe's own code constant (the M1 measurement; branches had grown to 422 at re-derivation) | SB08 |
| 10 | Web graph node bound | 300 — the union of 214 relation-named and 39 open cards, with headroom | QB04, QB02 |

Rows 1–9 are the card-issuance values this rule exists to hold. Row 10 is the
web view's bound — held here because it is the same measurement-shaped table
and the baseline sizes it; the graph view's code constant carries the value.
