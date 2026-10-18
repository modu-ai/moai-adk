# decision-index.md — SPEC-AGENT-CORE4-001 (piece 1)

Per `interview.decision_gate: on`. No product-level row carries a default or DEFAULT-APPLIED in
this piece. Authority register: committed artifacts only; the card text, the leader's rulings and
the design documents are not in it.

### Q1: One SPEC or several?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none)
Why unresolved: settled by the leader's ruling after plan-audit iter1 (split along the auditor's
seams); recorded here because the ruling is not a committed artifact.
Operator verdict: LEADER-RULED 2026-10-18 card t1649 (piece 1 = M0-M2)

### Q4: Does C1 become byte-identical to C2 for the carried four, reversing `AGENTS.local.md` §2.0?

Label: FOUNDER
Class: product-level
Authority anchor: (none — `AGENTS.local.md` §2.0 states C1 and C2 are intentionally not byte-identical, and is outside the register)
Why unresolved: REQ-AC4-006 and AC-AC4-006 require C1 = C2 for four agents, which contradicts that
[HARD] sentence; the sentence is user-owned and is not edited by this SPEC. This is the one open
row piece 1 needs answered before M2 is implementable. M0 and M1 do not depend on it.
Operator verdict:

### Q16: Development methodology

Label: POLICY-COVERED
Authority anchor: `.moai/config/sections/quality.yaml` `constitution.development_mode` (`tdd`)
Why unresolved: nothing; the plan's RED-first ordering follows the setting.
Operator verdict:

## Open rows inherited unchanged by pieces 2 and 3 (not needed for piece 1)

OPEN, no default: Q2 (card "TOML 12/12" vs 9 files; piece 1 measures N/N fresh only for files it
generates), Q3 (retire absorbed names by refusal vs alias), Q6 (auditor guard default), Q7 (memory
cap enforcement), Q8 (hide-mode ladder), Q9 (evidence-needed: preload probe without a live
runtime), Q10 (docs-site and README staleness), Q11 (unprefixed generic names). Definitions are in
commit `b4ea2d6c93` `decision-index.md`. The former DEFAULT-APPLIED rows Q5 (KB unit), Q12 (Codex
hiding), Q13 (refusal mechanism, tied to Q3), Q14 (auditor phase argument) and Q15 (`auditor`
substring coexistence) are withdrawn here and re-open in the pieces that need them.
