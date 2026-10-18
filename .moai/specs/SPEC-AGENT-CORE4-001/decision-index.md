# decision-index.md — SPEC-AGENT-CORE4-001 (piece 1)

Per `interview.decision_gate: on`. No product-level row carries a default or DEFAULT-APPLIED.
Operator verdicts are recorded in `progress.md` `## §F Operator verdicts`, not here, because this
file is inside the plan-artifact hash (`internal/runtime/audit_cache.go` plan artifact set); an
edit here after audit invalidates the audit.

### Q1: One SPEC or several?
Label: FOUNDER
Class: implementation-level
Authority anchor: (none)
Why unresolved: settled by the leader's ruling after plan-audit iter1; recorded because the ruling is not a committed artifact.
Operator verdict: LEADER-RULED 2026-10-18 card t1649 (piece 1 = former M0-M2)

### Q4: Does C1 become byte-identical to C2 for the carried four, reversing `AGENTS.local.md` §2.0?
Label: FOUNDER
Class: product-level
Authority anchor: (none — `AGENTS.local.md` §2.0 is outside the register)
Why unresolved: REQ-AC4-007 requires C1 = C2 for four agents, contradicting the [HARD] sentence "C1 <-> C2 is not byte-identical (an intentional divergence)". Gates M2 only.
Operator verdict:

### Q17: Does C2 stop being the hand-edited source, with C2 and C3 generated from agent sources from M1 on?
Label: FOUNDER
Class: product-level
Authority anchor: (none — `AGENTS.local.md` §2.0 rows C2 and C3 and its edit workflow are outside the register)
Why unresolved: §2.0 says C2 is "hand-edited — the neutral source layer" and C3 is emitted from C2. From M1, a hand edit of any C2 file turns `make build` red through `agents-emit-check`, and every concurrent lane that edits these twelve files must move the edit into `agentemit/source/`. Gates M1 onward (landing); M0 is implementable and landable without it; M1 can be built and tested on the branch but must not land first. Amendment to name: §2.0 rows C2 and C3, the "run `make agents-emit`" workflow, and the "neutral" wording.
Operator verdict:

### Q18: Tier of the narrowed SPEC
Label: FOUNDER
Class: implementation-level
Authority anchor: (none — the Tier table is a rule file outside the register)
Why unresolved: lane judgment on audit iter2 D7: about 24 authored files (39 with generated copies) and 2,969 lines of new tracked agent text; by files this is Tier L, by REQ/AC counts Tier M would fit. Chosen: Tier L, 12 REQ / 12 AC under ceilings 25 / 25, artifact set restored to five. The leader may rule otherwise.
Operator verdict: LANE-JUDGED 2026-10-18 card t1649 (Tier L)

### Q19: Which catalog citation survives on manager-design line 206?
Label: FOUNDER
Class: implementation-level
Authority anchor: (none)
Why unresolved: C1 cites `.moai/config/sections/delegation.yaml`, C2 cites `CLAUDE.md` § 4. The plan text uses the C1 citation because `delegation.yaml` exists in this repository and `CLAUDE.md` does not. A different choice changes two post pins (`acceptance.md` §3).
Operator verdict:

### Q16: Development methodology
Label: POLICY-COVERED
Authority anchor: `.moai/config/sections/quality.yaml` `constitution.development_mode` (`tdd`)
Why unresolved: nothing; the RED-first ordering follows the setting.
Operator verdict:

## Open rows inherited unchanged by pieces 2 and 3 (not needed for piece 1)

OPEN, no default: Q2 (card "TOML 12/12" vs 9 files), Q3 (absorbed names: refusal vs alias), Q6
(auditor guard default), Q7 (memory cap enforcement), Q8 (hide-mode ladder), Q9 (evidence-needed:
preload probe without a live runtime), Q10 (docs-site and README staleness), Q11 (unprefixed
generic names). Definitions: commit `b4ea2d6c93` `decision-index.md`. The former DEFAULT-APPLIED
rows Q5 (KB unit), Q12 (Codex hiding), Q13 (refusal mechanism, tied to Q3), Q14 (auditor phase
argument) and Q15 (`auditor` substring coexistence) are withdrawn and re-open in the pieces that
need them.
