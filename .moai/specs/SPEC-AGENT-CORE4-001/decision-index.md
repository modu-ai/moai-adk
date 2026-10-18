# decision-index.md — SPEC-AGENT-CORE4-001 (piece 1)

Per `interview.decision_gate: on`. This file is inside the plan-artifact hash
(`internal/runtime/audit_cache.go`), so an edit here after audit invalidates the audit; this
revision changes it deliberately. Verdicts on rows that are still open are recorded in `progress.md`
`## §G Operator verdicts`. No product-level row carries a default or DEFAULT-APPLIED.

### Q1: One SPEC or several?
Label: FOUNDER
Class: implementation-level
Authority anchor: (none)
Why unresolved: settled by the leader's ruling after plan-audit iter1; not a committed artifact.
Operator verdict: LEADER-RULED 2026-10-18 card t1649 (piece 1 = former M0-M2)

### Q4: Does C1 become byte-identical to C2, reversing `AGENTS.local.md` §2.0?
Label: FOUNDER
Class: product-level
Authority anchor: (none — the decision board is not in the committed authority register)
Why unresolved: closed.
Operator verdict: DECIDED by the operator — decision board `d-20261018T062320Z-5952` (decided_by=operator, 2026-10-18, scope card:t1649): agent copies become generated outputs; the agent source is the only edit surface; C1 is byte-identical to C2; the contrary §2.0 sentences are superseded and rewritten by this work (REQ-AC4-013).

### Q17: Does C2 stop being the hand-edited source, generated from agent sources from M1 on?
Label: FOUNDER
Class: product-level
Authority anchor: (none — as Q4)
Why unresolved: closed.
Operator verdict: DECIDED by the operator — the same board decision `d-20261018T062320Z-5952`: C2 is rendered from the agent source; hand-editing C2 fails `make build`; other in-flight cards editing agent files directly move their edits to the source layer when they land after piece 1.

### Q18: Tier of the narrowed SPEC
Label: FOUNDER
Class: implementation-level
Authority anchor: (none — the Tier table is a rule file outside the register)
Why unresolved: lane judgment: about 25 authored files (39+ with generated copies) and 2,969 lines of new tracked agent text; by files this is Tier L. 13 REQ / 13 AC under ceilings 25 / 25; artifact set five plus progress and decision-index. The leader may rule otherwise.
Operator verdict: LANE-JUDGED 2026-10-18 card t1649 (Tier L)

### Q19: Which catalog citation survives on manager-design line 206?
Label: FOUNDER
Class: implementation-level
Authority anchor: (none)
Why unresolved: corrected facts (grep): C2 cites `CLAUDE.md` § 4 (the file does not exist in this repository); C1 cites `.moai/config/sections/delegation.yaml` (the file exists, holds "13 retained agents" at line 14 and manager-design at line 97, but has no "Selection Decision Tree"). The tree is in `.claude/skills/moai-foundation-core/modules/agents-reference.md:46` and its row 6 is manager-git, so the sentence "manager-design is row 6" is stale under either citation. The plan keeps the C1 citation because three roster-guard registry rows match `(?:§ 4|delegation\.yaml`) (N retained agents` and a different file would turn them red. Changing it changes two derived results and the registry patterns.
Operator verdict:

### Q20: Is C1 generated for the eight in piece 1?
Label: FOUNDER
Class: implementation-level
Authority anchor: (none)
Why unresolved: the operator decision says C1 equals C2; for the eight, C1 differs from C2 in three files (manager-develop, manager-spec, plan-auditor), and C2 uses user-scope `~/.claude/` paths that C1 does not. Generating C1 for the eight would change shipped project bytes that piece 2 retires anyway.
Default: C1 of the eight stays hand-managed until piece 2 absorbs them; the rewritten §2.0 says C1 equals C2 for generated agents (rule: preserves current behavior)
Alternate: generate C1 for all twelve in M2
Operator verdict: DEFAULT-APPLIED 2026-10-18 card t1649 plan-author (manager-spec)

### Q16: Development methodology
Label: POLICY-COVERED
Authority anchor: `.moai/config/sections/quality.yaml` `constitution.development_mode` (`tdd`)
Why unresolved: nothing; the RED-first ordering follows the setting.
Operator verdict:

## Open rows inherited unchanged by pieces 2 and 3 (not needed for piece 1)

OPEN, no default: Q2, Q3, Q6, Q7, Q8, Q9 (evidence-needed), Q10, Q11. Definitions: commit
`b4ea2d6c93` `decision-index.md`. The former DEFAULT-APPLIED rows Q5, Q12, Q13, Q14, Q15 are
withdrawn and re-open in the pieces that need them.
