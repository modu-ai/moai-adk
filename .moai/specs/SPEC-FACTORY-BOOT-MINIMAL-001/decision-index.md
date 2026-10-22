# SPEC-FACTORY-BOOT-MINIMAL-001 — decision index

Authored at plan phase (`interview.decision_gate: on`, `.moai/config/sections/interview.yaml:6`). Stateless on the status axis; the SPEC's lifecycle lives in spec.md alone. Rows route by the four labels DECIDED / POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER. The published Default rule is stated in the manager-spec agent definition, which the authority register does not admit, so no row below claims POLICY-COVERED on it.

Revision 0.2.0 (plan-audit round 1 fix): Q1 is narrowed to the auto-mode explanation alone and its verdict source is stated plainly; Q5 names the concrete destinations; Q8, Q9, and Q10 are new.

### Q1: Should the bootstrap notices (leader and lane) drop the two-flag auto-mode explanation, inverting the SESSION-START-GUIDE-I18N-001 tests that pin its presence?

Label: FOUNDER
Class: product-level
Authority anchor: none in the committed register (the instruction lives in the backlog card, which is not a committed artifact)
Why unresolved: removing the explanation takes user-visible guidance out of a shipped session-start surface and inverts seven tests a completed SPEC landed. The completed SPEC's requirement text is conditional — its REQ-004 binds "where the guide text explains the two-mode split" and its REQ-005 binds "when the text names a flag" — so no requirement sentence is superseded; the tests, however, pin the explanation's presence. The committed register holds no record deciding the removal.
Operator verdict: CARD-DIRECTED 2026-10-22T21:05:59Z — this verdict is the operator's written instruction in backlog card t1923, not an interview answer; manager-spec transcribed it from `moai todo show t1923` and pins the clause below. The card names exactly this element ("자동 모드 2플래그 해설 … 디스크/온라인 문서로 이동"), and the operator's 10-22 correction on the same card keeps the start-message part in scope while allowing it to be split off. A reader who does not accept a card instruction as an operator verdict should treat this row as open and hold the Kickoff.

```text
범위: (1) 부팅 공지 최소화 — 리더·레인 모두 run id·슬롯 라벨·스폰 권한 한 줄·채널 포인터·docs URL만 남기고 자동 모드 2플래그 해설 등은 디스크/온라인 문서로 이동 (2)
```

### Q2: Does the leader notice keep the lane-start sentence (the `moai cc -l` / `moai glm -l` / `moai codex -l` entries), which the card's element list does not name?

Label: FOUNDER
Class: implementation-level
Default: keep the lane-start sentence (rule: the option that preserves current behavior — it is the operator's only in-notice instruction for opening lane terminals)
Alternate: drop it and rely on the docs URL
Why unresolved: the card's element list is silent on it; no register entry decides it.
Operator verdict: DEFAULT-APPLIED 2026-10-22T21:05:59Z manager-spec (t1923, plan phase)

### Q3: Is the standing spawn-authority sentence compressed to fit "스폰 권한 한 줄", or kept verbatim?

Label: FOUNDER
Class: implementation-level
Default: keep the sentence verbatim (rule: the option that preserves current behavior — the test markers in `lane_spawn_authority_test.go` and the t224 incident fix depend on it; it is already one sentence)
Alternate: rewrite it shorter and amend the marker tests
Why unresolved: wording judgment; no register entry decides it.
Operator verdict: DEFAULT-APPLIED 2026-10-22T21:05:59Z manager-spec (t1923, plan phase)

### Q4: Do the 0.6.0 budget numbers (composite 9,146, core 4,092, producer 4,797) change?

Label: FOUNDER
Class: implementation-level
Default: keep all three numbers; split the producer 4,797 into a measured non-factory allowance 1,950, the joiner 2, and a factory fixed-part ceiling 2,845 (rule: the option that preserves current behavior — no contract number moves)
Alternate: lower the composite budget (e.g. 9,000) to buy more headroom under the 10,000 cap
Why unresolved: the card permits a documented change but does not require one; measurements (plan.md §B) show the split is enough to catch the English-lane risk.
Operator verdict: DEFAULT-APPLIED 2026-10-22T21:49:04Z manager-spec (t1923, plan phase, revision 0.2.0 — ceiling re-derived from 2,847 to 2,845 once the joiner was counted)

### Q5: Where does the relocated guidance land?

Label: FOUNDER
Class: implementation-level
Default: in the existing documents that already carry each element, both faces — `.claude/skills/moai/workflows/gtd.md` (R1 auto modes, R5 assigned-card lease), `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` (R2 landing, R3 leader's remaining role), `.claude/rules/moai/workflow/auto-semantics.md` §6 (R4 judgment ladder); never the two role-gated rule files (rule: the option with the smaller user-visible surface — nothing new is written, and the role core stays byte-identical and out of t1637's lane)
Alternate: add the text to `factory-dispatch.md` outside the role-core markers
Why unresolved: placement judgment; no register entry decides it.
Operator verdict: DEFAULT-APPLIED 2026-10-22T21:49:04Z manager-spec (t1923, plan phase, revision 0.2.0 — destinations named)

### Q6: What is the "channel pointer" in the card's element list?

Label: FOUNDER
Class: implementation-level
Default: the existing leader socket line (the channel address the launcher captured) plus one line naming the repository path of the relocated guidance (rule: the option that preserves current behavior — the socket line already ships)
Alternate: a pointer to the cross-session messaging rule only
Why unresolved: the card term is undefined; no register entry decides it.
Operator verdict: DEFAULT-APPLIED 2026-10-22T21:05:59Z manager-spec (t1923, plan phase)

### Q7: How is the operator's AGENTS.md byte-ceiling patch applied, given it no longer applies to the current tree?

Label: FOUNDER
Class: implementation-level
Default: carry the patch's two replacement paragraphs onto the current tree, writing "WT worktrees live only under" in place of the patch's "belong under" (rule: the option that preserves current behavior — the exclusivity is an obligation the patch weakened); target ≤ 24,276 bytes (300 bytes headroom, a judgment value)
Alternate: skip the patch, since the ceiling tests are already green at 24,501 bytes (75 bytes headroom)
Why unresolved: the card asks to review and apply; the tree moved under the patch (#1869), and the headroom target is a judgment value.
Operator verdict: DEFAULT-APPLIED 2026-10-22T21:05:59Z manager-spec (t1923, plan phase)

### Q8: Does the leader notice also drop the elements outside the card's list — entry-point guide, fan-out line, card routing, stagger, batch gate summary, operational status, settings line?

Label: FOUNDER
Class: implementation-level
Default: keep them, verbatim (rule: the option that preserves current behavior — the batch gate summary sentence is required by the completed SPEC-AUTONOMY-BATCH-GATE-001 REQ-BGS-016/017 under that SPEC's operator decision Q5, the routing and stagger text is pinned by the t85 tests `TestFactoryLeadNoticeCarriesDispatchDiscipline` and its Korean variant, the entry guide, fan-out, and socket lines by `TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide`; and the leader cell is not over budget — leader en fixed 2,573 ≤ 2,845, so the budget does not force any removal)
Alternate: drop them, superseding REQ-BGS-016/017 and rewriting the t85 tests — a product-level change that would need its own operator verdict
Why unresolved: the card's element list reads as exclusive for both roles, while a completed SPEC decided by the operator requires one of the elements; the operator's 10-22 correction lets the start-message part be split off, and no register entry reconciles the two.
Operator verdict: DEFAULT-APPLIED 2026-10-22T21:49:04Z manager-spec (t1923, plan phase, revision 0.2.0)

### Q9: The card lists "run id · slot label · spawn-authority line · channel pointer · docs URL" for both roles; does the lane notice gain the run id, and does the leader notice gain a spawn-authority line?

Label: FOUNDER
Class: implementation-level
Default: neither — each role keeps the elements it carries today (rule: the option that preserves current behavior — the lane join line names the lane label and the run it joined, the run id rides the lane's environment, and a leader holds no lane spawn authority to state)
Alternate: add the run id to the lane join line and a spawn line to the leader notice
Why unresolved: the card's list does not say whether it is per role or common to both; no register entry decides it.
Operator verdict: DEFAULT-APPLIED 2026-10-22T21:49:04Z manager-spec (t1923, plan phase, revision 0.2.0)

### Q10: Which lane-rule prose leaves the rule so the English lane cell fits the 2,845 ceiling?

Label: FOUNDER
Class: implementation-level
Default: exactly R2-R5 of spec.md §B.3 — the landing detail, the leader's-remaining-role list, the judgment-ladder detail, and the assigned-card `--card` clause — each already present in an existing document that the rule then names by path (rule: the option with the smaller user-visible surface — no new text is written anywhere and every pinned token of REQ-FBM-006 stays)
Alternate: shorten the stall-recovery paragraph, whose rationale sentence has no existing on-disk home and would need new destination text
Why unresolved: selection judgment among unpinned prose; no register entry decides it. The plan-phase simulation (plan.md §B.7, R1-R5 in illustrative wording) measures the English lane startup cell at 2,727 (margin 118); the R5 clause is 104 UTF-16 units of that in English, so without R5 the margin would be about 14.
Operator verdict: DEFAULT-APPLIED 2026-10-22T21:49:04Z manager-spec (t1923, plan phase, revision 0.2.0)
