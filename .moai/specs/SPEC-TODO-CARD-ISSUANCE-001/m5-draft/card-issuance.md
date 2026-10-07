---
description: Card issuance mechanics — size bounds, the follow-up defect rule, derivation depth, the concurrency ceiling, the issuance checklist, and the debt ledger. Every value is the project's own measurement, held in the local-only thresholds rule that loads alongside this one.
paths: "**/.claude/skills/moai/workflows/gtd.md,**/.claude/agents/moai/manager-todo.md"
---

# Card Issuance

> Mechanism only, by design. This rule states HOW card issuance is judged;
> the numbers live in the local-only `card-issuance-thresholds.md` rule that
> shares its `paths:` trigger, and each is the value the project measured
> from its own queue distribution — a project re-measures and re-sets its
> own. Nothing here names a card, a path inside the measuring repository, or
> a threshold value.

## 1. Card size

A card whose estimated product footprint sits under the lower bound is a
bundle candidate: before issuing it, read the open cards that touch the same
primary files and issue the small card as a bundle member (or folded work)
rather than as a standalone queue row. A card above the upper bound — lines
or files — is a split candidate: issue its parts, not the whole.

Both bounds, and what counts as "product" work inside the estimate, are
thresholds-rule values.

## 2. Follow-up defect rule

A defect a card's own change introduced is repaired INSIDE that card, before
its local merge — it never becomes a follow-up card. A defect the card merely
exposed, one that predates the change, goes to the per-component debt ledger
with a one-line reproduction. The sync review's blocking/optional
classification and the PASS-WITH-DEBT verdict are the downstream consumers of
the same split: blocking means the card's own defect did not get repaired
in-card.

## 3. Derivation depth

A card spawned from a card spawned from a card is the deepest the chain goes.
Past the depth bound, the descendant is issued as its own top-level card (or
the chain is folded into an epic-level container) — deep spawn chains are how
a queue grows cards nobody can trace to a decision. The bound is a
thresholds-rule value.

## 4. Concurrent-progress ceiling

The number of cards in flight at once is capped. The cap exists because the
measured distribution, not intuition, says where a queue's parallelism stops
paying — above it, review latency eats the throughput the concurrency bought.
The ceiling is a thresholds-rule value.

## 5. Issuance checklist

Before a card enters the queue, the issuing session has read, from the `add`
presentation alone:

1. the exact-text and near-duplicate matches (an exact collision is refused
   without `--force`; a near match is a bundle candidate);
2. the same-component neighbors (same depth-2 path prefix — a sequencing or
   bundle candidate);
3. the in-flight overlap (which open card's expected files this card shares,
   or `unmeasured` with the reason);
4. the completed-SPEC matches (is the work already done?);
5. the size bounds — below the lower bound, bundle; above the upper, split;
6. the parent and origin inputs (`--parent` names an existing card; `--origin`
   draws from the closed set);
7. the bundle candidacy the hub chain will order at lease time.

## 6. Debt ledger

Debt is a queue row like any other, carrying the debt origin marker — distinct
from a follow-up origin: a follow-up is the next step of finished work, a debt
row is a known defect deliberately not repaired now. The ledger is queryable
per component (the depth-2 path prefix is the grouping key), so a component's
open debt is one read, not an archaeology project. A debt row carries its
one-line reproduction in the body; a debt row that cannot state how to
reproduce the defect is not debt, it is a hunch, and it does not enter the
queue as one.
