# SPEC-FACTORY-BOOT-MINIMAL-001 — decision index

Authored at plan phase (`interview.decision_gate: on`, `.moai/config/sections/interview.yaml:6`). Stateless on the status axis; the SPEC's lifecycle lives in spec.md alone. Rows route by the four labels DECIDED / POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER. The published Default rule is stated in the manager-spec agent definition, which the authority register does not admit, so no row below claims POLICY-COVERED on it.

### Q1: Should the bootstrap notice drop the two-flag auto-mode explanation (and the leader's routing / stagger / gate-summary / status blocks), superseding the SESSION-START-GUIDE-I18N-001 contract that requires that explanation in the notice?

Label: FOUNDER
Class: product-level
Authority anchor: none in the committed register (the instruction lives in the backlog card, which is not a committed artifact)
Why unresolved: the change removes user-visible guidance from a shipped session-start surface and inverts tests a completed SPEC pinned; that is a product-level call, and the committed register holds no record deciding it.
Operator verdict: CARD-DIRECTED 2026-10-22T21:05:59Z — the operator's own instruction in card t1923 (10-21), read verbatim from `moai todo show t1923` by manager-spec (t1923 plan phase) and pinned below. The pinned clause orders exactly this removal; the SPEC implements it as REQ-FBM-001/002/003. A reader who does not accept a card instruction as an operator verdict should treat this row as open and hold the Kickoff.

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
Default: keep the sentence verbatim (rule: the option that preserves current behavior — six test markers in `lane_spawn_authority_test.go` and the t224 incident fix depend on it; it is already one sentence)
Alternate: rewrite it shorter and amend the marker tests
Why unresolved: wording judgment; no register entry decides it.
Operator verdict: DEFAULT-APPLIED 2026-10-22T21:05:59Z manager-spec (t1923, plan phase)

### Q4: Do the 0.6.0 budget numbers (composite 9,146, core 4,092, producer 4,797) change?

Label: FOUNDER
Class: implementation-level
Default: keep all three numbers; split the producer 4,797 into a measured non-factory allowance 1,950 plus a factory fixed-part ceiling 2,847 (rule: the option that preserves current behavior — no contract number moves)
Alternate: lower the composite budget (e.g. 9,000) to buy more headroom under the 10,000 cap
Why unresolved: the card permits a documented change but does not require one; measurements (plan.md §B) show the split is enough to catch the English-lane risk.
Operator verdict: DEFAULT-APPLIED 2026-10-22T21:05:59Z manager-spec (t1923, plan phase)

### Q5: Where does the relocated guidance land?

Label: FOUNDER
Class: implementation-level
Default: the detail companions (`factory-dispatch-detail.md` / `factory-dispatch-mechanics.md`) or the todo workflow skill — never the two role-gated rule files (rule: the option with the smaller user-visible surface; it also keeps the role core byte-identical and out of t1637's lane)
Alternate: add the text to `factory-dispatch.md` outside the role-core markers
Why unresolved: placement judgment; no register entry decides it.
Operator verdict: DEFAULT-APPLIED 2026-10-22T21:05:59Z manager-spec (t1923, plan phase)

### Q6: What is the "channel pointer" in the card's element list?

Label: FOUNDER
Class: implementation-level
Default: the existing leader socket line (the channel address the launcher captured) plus one line naming the on-disk location of the relocated guidance (rule: the option that preserves current behavior — the socket line already ships)
Alternate: a pointer to the cross-session messaging rule only
Why unresolved: the card term is undefined; no register entry decides it.
Operator verdict: DEFAULT-APPLIED 2026-10-22T21:05:59Z manager-spec (t1923, plan phase)

### Q7: How is the operator's AGENTS.md byte-ceiling patch applied, given it no longer applies to the current tree?

Label: FOUNDER
Class: implementation-level
Default: carry the patch's two replacement paragraphs onto the current tree, restoring the word "only" in "WT worktrees live only under" (rule: the option that preserves current behavior — the exclusivity is an obligation the patch weakened); target ≤ 24,276 bytes (300 bytes headroom, a judgment value)
Alternate: skip the patch, since the ceiling tests are already green at 24,501 bytes (75 bytes headroom)
Why unresolved: the card asks to review and apply; the tree moved under the patch (#1869), and the headroom target is a judgment value.
Operator verdict: DEFAULT-APPLIED 2026-10-22T21:05:59Z manager-spec (t1923, plan phase)
