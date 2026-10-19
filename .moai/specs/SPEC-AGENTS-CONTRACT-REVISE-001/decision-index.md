# SPEC-AGENTS-CONTRACT-REVISE-001 — Decision Index

Authored at plan phase (decision gate on); revised at the 2026-10-19 correction round (items: auto-merge conditions, size-diet target replacement, CLI/slash forms, review triggers + Failed state). Stateless on the status axis — the SPEC's lifecycle lives in `spec.md` alone. Every row carries Detect → Explain → Ask and no embedded recommendation beyond a rule-ranked Default.

## Always-loaded measurement table (HEAD `9b1dcde024`, 2026-10-19 — leader correction record "리더 정정")

Detection rule: repo-tracked session surfaces; a `.claude/rules/**/*.md` file counts as always-loaded when its frontmatter carries no `paths:` key (conditional rules load only for matching paths). Measured with `wc -m` in this tree, this run.

| Surface | Chars (`wc -m`) |
|---|---|
| `AGENTS.md` | 23,680 |
| `AGENTS.local.md` | 50,579 |
| Output style `.claude/output-styles/moai/moai.md` | 60,888 |
| Always-loaded rules (13 files, no `paths:`) — largest: agent-common-protocol 18,298 · askuser-protocol 15,163 · moai-constitution 15,076 · session-handoff 14,332 · verification-claim-integrity 14,235 · factory-dispatch-core 9,999 · cross-session-messaging-core 8,072 · main-checkout-branch-guard 5,257 · context-window-management 5,014 · cache-aware-execution 5,009 · goal-directive 3,618 · moai-mcp-tools 1,718 · native-idiom-and-register 1,712 | 117,503 |
| **Measured total (this detection rule)** | **252,650** |

**Derived-share projection (leader correction — leader-derived figures, explicitly not operator-set).** Governing target: always-loaded TOTAL ≤ 140,000 chars (hard limit 150,000). The baseline is PINNED to the pre-card measurement at commit `9b1dcde024` and is never re-measured into the formula at close time, so this card's own reduction cannot relax its own gate. Required combined reduction: 252,650 − 140,000 = **112,650**. This card's confirmed scope — moving the four named chunks (§4.0 L182-238 = 7,024 · §4.1 L239-292 = 11,283 · §2.0 L91-120 = 3,967 · References L482-500 = 2,761; sum ≈ 25,035 — awk byte-lengths, C locale, not `wc -m` counts; re-measured at this HEAD) into paths:-scoped companions — proceeds regardless and saves ≈ 25,035 (less pointer-line overhead; the unit skew vs the `wc -m` total is ~1% here and not decision-bearing). The residual savings must come from card t1887's core-rules diet: **R_savings ≥ 112,650 − 25,035 ≈ 87,615 chars of the 117,503-char rules block (≈ 75%)**. Sanity check: with a full rules diet and AGENTS.local.md at ≈ 25,544 post-move, the total lands at ≈ 110,112 ≤ 140,000 (headroom ≈ 29,888); with a partial t1887 diet, the untouched output-style (60,888) and AGENTS.md (23,680) surfaces are the named next diet candidates.

Gaps of this table: session-side stores (auto-memory `MEMORY.md`), static-preloaded skill bodies, and conditional (`paths:`-scoped) rules are outside the repo-tracked measurement; t1887's actual delta is unknown here and enters the projection as a required-savings bound, not a measurement.

### Q1: Which copy of AGENTS.local.md is the diet baseline, and what target governs?

Label: FOUNDER
Authority anchor: none admissible — the candidate anchor (`AGENTS.local.md` §0.1/§0.2, the canonical-copy discriminator) is the very file under revision and sits outside the authority register; the corrected target arrived through the leader chat channel (2026-10-19), which is not a registerable anchor.
Why unresolved: the original ≤12,000-chars figure came from a bad baseline — 31,971 measured on a primary working copy 1,587 commits behind main — and is **withdrawn as a governing target** (provenance: bad baseline → leader correction). The governing target is now the always-loaded TOTAL ≤ 140,000 chars (hard 150,000) with AGENTS.local.md's share derived (table above).
Class: implementation-level
Default: committed copy in the card's own tree at close time, measured against the always-loaded-total model above (rule: preserves current behavior — the committed branch-point copy is what §0.1 governs today).
Alternate: any fixed per-file char target (including the withdrawn 12,000).
Operator verdict: DEFAULT-APPLIED 2026-10-19T05:23:44Z manager-spec (lane-44, card t1900); superseded target recorded per leader correction of the same date.

### Q2: What companion-doc structure receives the moved AGENTS.local.md content?

Label: FOUNDER
Authority anchor: none admissible — the operator dispatch names only an example path (`.claude/rules/local/agents-local-detail.md`), and no config-section or constitution clause covers companion-doc structure.
Why unresolved: the operator's brief gives an example, not a decision; one companion versus per-topic split changes the paths: glob surface that AC-ACR-009 must verify.
Class: implementation-level
Default: a single paths:-scoped companion at the operator-suggested path `.claude/rules/local/agents-local-detail.md` (rule: smaller user-visible surface — one glob, one index of moved clauses).
Alternate: per-topic companions (flow / integration-chain / references) with separate globs.
Operator verdict: DEFAULT-APPLIED 2026-10-19T05:23:44Z manager-spec (lane-44, card t1900)

### Q3: How is the size model met when the four named chunks total only ≈25,035 of the 112,650-char required reduction?

Label: FOUNDER
Authority anchor: none admissible — the corrected target is leader-derived (chat channel, 2026-10-19), and the arithmetic gap is measured only in this tree (table above; spec.md §C MF-6).
Why unresolved: the four-chunk move is confirmed scope that proceeds regardless, but it covers ≈25,035 of the 112,650-char required reduction; compression beyond the four chunks is real and its depth depends on t1887's diet, which this card cannot measure.
Class: implementation-level
Default: execute the four-chunk move unconditionally; record any further AGENTS.local.md compression in this index as "리더 정정" with the measurement table — a leader-derived figure, explicitly not operator-set; the residual (≈87,615 chars) is assigned to card t1887's core-rules diet share in the projection (rule: the only option consistent with the leader-corrected ≤140,000 model).
Alternate: escalate for additional diet cards (output-style 60,888 / AGENTS.md 23,680 surfaces) before compressing AGENTS.local.md further.
Operator verdict: DEFAULT-APPLIED 2026-10-19T05:23:44Z manager-spec (lane-44, card t1900); leader correction recorded same date.

### Q4: How are the automated card-dispatch command forms written in the revised docs?

Label: FOUNDER
Authority anchor: none admissible — the CLI/slash distinction arrived through the leader correction channel (2026-10-19); no config section or constitution clause pins the surface forms.
Why unresolved: the docs historically mixed the CLI shell form with the slash-skill forms; the correction fixes the distinction (shell = `moai todo --auto` only; slash-skill = `/moai todo --auto-lane`, `/moai todo --auto-leader`) but the sweep boundary (docs only vs code strings) needed separation.
Class: implementation-level
Default: this card fixes the forms in the revised documents only; code-side strings (SessionStart hook, codex launcher) are card t1695's delegated scope (rule: preserves current behavior — smallest surface for a docs card).
Alternate: include the code strings in this card (rejected — out of a docs card's scope and owned by t1695).
Operator verdict: DEFAULT-APPLIED 2026-10-19T05:23:44Z manager-spec (lane-44, card t1900); leader correction recorded same date.
