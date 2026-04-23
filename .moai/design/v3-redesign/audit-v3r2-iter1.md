# v3R2 Plan Audit — Iteration 1

> Auditor: plan-auditor (adversarial, independent)
> Date: 2026-04-24
> Scope: `docs/design/major-v3-master.md` (1,212 lines) + 35 SPEC-V3R2-*/spec.md
> Inputs: `.moai/design/v3-redesign/research/r1..r6.md` + `synthesis/{design-principles,problem-catalog,pattern-library}.md`
> Reasoning context ignored per M1 Context Isolation. Ran against spec.md files + master + research only.

## Executive Summary

- **Verdict: NOT-READY**
- **Critical: 7** / **High: 9** / **Medium: 7** / **Low: 4** / **Total: 27**
- Top blockers: (1) The dependency graph contains **4 true cycles** among ORC-001↔{002,003,004} and WF-001↔MIG-001, making the DAG assumption false and topological ordering impossible. (2) Two **phase-order violations** (HRN-001 P5 → MIG-003 P8; WF-001 P4 → MIG-001 P8) contradict the release sequencing in master §9 — a Phase-4 SPEC cannot depend on a Phase-8 SPEC. (3) **BC ownership is inconsistent**: BC-V3R2-002 and BC-V3R2-006 are each claimed by two SPECs with semantically contradictory scopes; BC-V3R2-EXT004 violates the canonical `BC-V3R2-NNN` schema; 4 master-declared BCs (-004, -007, -012, -013) have **no owning SPEC**. (4) **SPC-001 amends a FROZEN zone** (the SPEC system EARS format listed as FROZEN in master §1.3) but depends only on CON-001, bypassing the CON-002 amendment / graduation protocol that the FROZEN zone model requires. (5) Four SPECs have frontmatter `phase:` values that disagree with master §9 Release Plan assignments (RT-005, ORC-005, WF-005, WF-006), meaning the release schedule in master and the per-SPEC phase metadata are not reconcilable.

---

## Defect Catalog

### Critical (7)

**D-CRIT-001: Dependency graph contains 4 cycles (violates D4 DAG requirement)**
- Cycle A: `SPEC-V3R2-ORC-001 ↔ SPEC-V3R2-ORC-002` — ORC-001/spec.md:frontmatter deps `[ORC-002, ORC-003, ORC-004]`; ORC-002/spec.md:frontmatter deps `[ORC-001]`. Evidence: `SPEC-V3R2-ORC-001/spec.md:L14-18` and `SPEC-V3R2-ORC-002/spec.md:L14-17`.
- Cycle B: `SPEC-V3R2-ORC-001 ↔ SPEC-V3R2-ORC-003`. Evidence: `SPEC-V3R2-ORC-003/spec.md:L14-18`.
- Cycle C: `SPEC-V3R2-ORC-001 ↔ SPEC-V3R2-ORC-004`. Evidence: `SPEC-V3R2-ORC-004/spec.md:L14-18`.
- Cycle D: `SPEC-V3R2-WF-001 ↔ SPEC-V3R2-MIG-001`. `SPEC-V3R2-WF-001/spec.md:L13` has `dependencies: [MIG-001]`; `SPEC-V3R2-MIG-001/spec.md:L15` has `dependencies: [EXT-004, WF-001, MIG-002, MIG-003]`.
- Severity: **critical**. `dependencies:` frontmatter defines a topological implementation order; a cycle means neither SPEC can be landed first.

**D-CRIT-002: Phase-order violations — Phase-N SPEC depends on Phase-M>N SPEC**
- `SPEC-V3R2-HRN-001/spec.md:L10` phase = "Phase 5"; its `dependencies:` includes `SPEC-V3R2-MIG-003` which is Phase 8 (`SPEC-V3R2-MIG-003/spec.md:L10`). A Phase-5 release cannot require a Phase-8 SPEC.
- `SPEC-V3R2-WF-001/spec.md:L10` phase = "Phase 4"; `dependencies: [SPEC-V3R2-MIG-001]` which is Phase 8.
- Severity: **critical**. Implies the Release Plan in master §9 (L983-996) cannot be executed alpha-by-alpha.

**D-CRIT-003: BC-V3R2-002 double-claimed with conflicting semantics**
- Master §8 L961: `BC-V3R2-002 | Agent frontmatter requires effort field` → owner should be ORC-003 (effort calibration).
- `SPEC-V3R2-ORC-003/spec.md:L28` `bc_id: [BC-V3R2-002]` — matches master semantics.
- `SPEC-V3R2-RT-002/spec.md:L13` `bc_id: [BC-V3R2-002]` — but RT-002 is "Permission Stack + Bubble Mode" (semantically unrelated). This is a labeling error; RT-002 likely intended a different BC number.
- Severity: **critical**.

**D-CRIT-004: BC-V3R2-006 double-claimed**
- Master §8 L965: `BC-V3R2-006 | 48 skills → 24 via merge clusters` → owner is WF-001.
- `SPEC-V3R2-WF-001/spec.md:L25` `bc_id: [BC-V3R2-006]` — matches master.
- `SPEC-V3R2-ORC-001/spec.md:L33` `bc_id: [BC-V3R2-005, BC-V3R2-006, BC-V3R2-009, BC-V3R2-016]` — ORC-001 wrongly also claims BC-V3R2-006. ORC-001 is 22→17 agent consolidation, not skill consolidation.
- Severity: **critical**.

**D-CRIT-005: BC-V3R2-EXT004 violates canonical BC-V3R2-NNN schema**
- `SPEC-V3R2-EXT-004/spec.md:L19` `bc_id: [BC-V3R2-EXT004]`. This ID is not in master §8 catalog (which uses strict 3-digit numeric IDs `BC-V3R2-001..018`). It cannot be cross-referenced from master.
- Severity: **critical**. Breaks D6 + D11 frontmatter schema uniformity.

**D-CRIT-006: 4 master-declared BCs have no owning SPEC**
- Master §8 declares BC-V3R2-001 through BC-V3R2-018. Assigned across all 35 SPECs: 001, 002, 003, 005, 006, 008, 009, 010, 011, 014, 015, 016, 017, 018.
- Unassigned: `BC-V3R2-004` (Agent tool removal — should be ORC-002), `BC-V3R2-007` (Agentless flip — should be WF-004), `BC-V3R2-012` (/98 /99 extraction — should be WF-002), `BC-V3R2-013` (config loaders — should be MIG-003 / HRN-001).
- Evidence: grep of all 35 `bc_id:` frontmatter returns zero hits for these four.
- Severity: **critical**. Every BC listed in master §8 must be traceable to the SPEC that lands it; without an owner, there is no implementation commitment.

**D-CRIT-007: SPC-001 amends a FROZEN zone but bypasses CON-002 amendment protocol (D16)**
- Master §1.3 L48-56 lists "SPEC system EARS format" among the 7 FROZEN invariants.
- `SPEC-V3R2-SPC-001/spec.md:L2-3` is titled "EARS + hierarchical acceptance criteria" with `breaking: true, bc_id: [BC-V3R2-011]`. It explicitly modifies the SPEC EARS format (hierarchical nesting + flat-to-hierarchical migration per master L970).
- `SPEC-V3R2-SPC-001/spec.md:L13-14` dependencies: `[SPEC-V3R2-CON-001]` only. It does NOT depend on SPEC-V3R2-CON-002 (amendment protocol / 5-layer safety gate).
- Compare with HRN-002 which correctly routes a FROZEN-zone amendment through CON-002 (`SPEC-V3R2-HRN-002/spec.md:L15-18`).
- Severity: **critical**. Violates FROZEN governance contract.

### High (9)

**D-HIGH-001: 4 frontmatter phase mismatches with master §9 Release Plan**
- `SPEC-V3R2-RT-005/spec.md:L10` = "Phase 2 — Runtime Hardening"; master L987 assigns it to Phase 1.
- `SPEC-V3R2-ORC-005/spec.md:L10` = "Phase 6 — Multi-Mode Workflow"; master L989 assigns it to Phase 3.
- `SPEC-V3R2-WF-005/spec.md:L10` = "Phase 6"; master L990 assigns it to Phase 4.
- `SPEC-V3R2-WF-006/spec.md:L10` = "Phase 7"; master L992 assigns it to Phase 6.

**D-HIGH-002: EXT-003 is not listed in master §9 Release Plan**
- Master §9 Release Plan rows Phase 1-8 omit EXT-003. Master L1077 says EXT-003 is "design-only / Deferred".
- But `SPEC-V3R2-EXT-003/spec.md:L10` = "Phase 7 — Extension". The SPEC is physically present (full spec written); if design-only it should carry `lifecycle: spec-first` AND be explicitly marked deferred. It uses `lifecycle: spec-first` (OK) but the missing Release Plan entry is inconsistent.

**D-HIGH-003: Master §11 count says "~35 SPECs" (correct at 35) but §8 / §11 do not state REQ/AC totals; actual REQ count ≈ 695, AC ≈ 402 G/W/T triplets + 147 AC-NNN labels.**
- Master §11.8 L1086: "Total: 35 SPECs." No REQ/AC count given.
- Wave-5 input stated "~550 REQs, ~350 ACs"; actual ≈ 695 REQs / ≈ 549 ACs aggregate. Master should publish these figures for auditability.

**D-HIGH-004: 5 REQ ID duplicates within single SPECs (D2)**
- `REQ-EXT001-004` appears on `SPEC-V3R2-EXT-001/spec.md:L110` and again as a second definition within the same requirements block.
- `REQ-EXT001-006` same pattern.
- `REQ-EXT004-004` in `SPEC-V3R2-EXT-004/spec.md`.
- `REQ-WF003-016` in `SPEC-V3R2-WF-003/spec.md`.
- `REQ-WF006-013` in `SPEC-V3R2-WF-006/spec.md`.
- Either a requirement is redefined (critical) or the second occurrence is a cross-reference — the pattern needs review.

**D-HIGH-005: REQ-to-SHALL/SHOULD/MAY normative verb coverage gaps (D1)**
- ORC-003: 14 REQs, only 13 SHALL. One REQ has no EARS normative verb.
- EXT-002: 19 REQs, 16 SHALL, 0 SHOULD, 0 MAY. Three REQs lack normative verb.
- WF-002: 15 REQs, 13 SHALL. Two REQs lack normative verb.
- EXT-003: 17 REQs, 15 SHALL. Two REQs lack normative verb.
- MIG-002: 19 REQs, 17 SHALL. Two REQs lack normative verb.

**D-HIGH-006: ORC-004 over-uses SHOULD (D1 rubric anchor)**
- `SPEC-V3R2-ORC-004/spec.md` SHALL=16, SHOULD=5. ORC-004 is a "MUST Rule" SPEC (title: "Worktree MUST Rule for write-heavy role profiles") yet 5 SHOULD markers weaken its normative stance. Either upgrade to SHALL or restate scope.

**D-HIGH-007: No Wave 1/2 file:line citations in §10 Traceability (D5)**
- All 35 §10 Traceability sections cite `master-v3 §N`, `pattern-library.md §X`, `problem-catalog.md Cluster N`. None cite concrete `file:line` locations as D5 requires (example: `r1-ai-harness-papers.md:L234`).
- Spot-check `SPEC-V3R2-CON-001/spec.md:L174-184`: "R1 §18 Constitutional AI", "R3 §4 Adoption Candidate 7" — section citations, not file:line.

**D-HIGH-008: ORC-005 layer boundary mixing (D15)**
- `SPEC-V3R2-ORC-005/spec.md:L22` theme: "Layer 4 — Orchestration + Layer 6 — Workflow (cross-layer)". Frontmatter `module:` spans `workflow.yaml + team_spawn.go + worktree-integration.md + team-protocol.md`. Master §11.4 places ORC-005 under Layer 4 only. If cross-layer, either split the SPEC or explicitly document layer-crossing rationale with owner.

**D-HIGH-009: EXT-003 "design-only, deferred" status is inconsistent**
- Master §1.3 and §11.7 describe EXT-003 as "no implementation in v3 / deferred".
- `SPEC-V3R2-EXT-003/spec.md` has 17 REQs with full EARS statements. For a design-only deferred SPEC, body should be a placeholder, not a fully-specified REQ set.

### Medium (7)

**D-MED-001: Master top-5 BC summary (L979) cites BC-V3R2-006 which is double-owned (see D-CRIT-004). Summary wording implies single owner.**

**D-MED-002: Mixed frontmatter tag set — some SPECs use `related_gap`, others `related_problem`, others both; the schema is not uniform.**
- `SPEC-V3R2-CON-001` has `related_problem` + `related_pattern` + `related_principle`. `SPEC-V3R2-EXT-001` has `related_gap` only. `SPEC-V3R2-HRN-001` has `related_problem` only. No documented precedence rule.

**D-MED-003: Field ordering inconsistency — some SPECs put `bc_id` before `dependencies`, others after.**
- `SPEC-V3R2-RT-001/spec.md:L13` has `bc_id` near top; `SPEC-V3R2-WF-001/spec.md:L25` has it near bottom. Lints would fail.

**D-MED-004: EXT-004 does NOT depend on CON-001 (Constitution) despite being a Phase-1 Critical SPEC**
- `SPEC-V3R2-EXT-004/spec.md:L16` `dependencies: []`. Phase-1 foundation SPECs should depend on CON-001 as the zone model anchor.

**D-MED-005: WF-006 double-claims "Phase 7 — Extension" in frontmatter but its `related_theme` says "Theme 7 — Extension" — master §9 L992 assigns WF-006 to Phase 6 (Multi-Mode Workflow).** Duplicate with D-HIGH-001 but the cross-reference to `related_theme` compounds the confusion.

**D-MED-006: Inconsistent dependency cardinality**
- `SPEC-V3R2-EXT-004/spec.md:L16` deps = `[]` (Critical SPEC with no deps).
- `SPEC-V3R2-WF-006/spec.md:L16` deps = `[]` (Medium SPEC, OK).
- `SPEC-V3R2-ORC-001/spec.md:L14` deps includes 4 downstream SPECs (ORC-002/3/4 AND CON-001) — which creates the cycles. Dependencies should point only upward (to foundational SPECs).

**D-MED-007: Total REQ count ≈ 695 vs Wave-5 target "~550" — 26% higher than intended. Suggests scope inflation.**

### Low (4)

**D-LOW-001: HISTORY table has only 1 row in most SPECs (v0.1.0 initial). No audit finding — just note that `updated: 2026-04-23` in frontmatter matches HISTORY.**

**D-LOW-002: Some SPECs use Korean headers `## 10. Traceability (추적성)`, others pure English `## 10. Traceability`. Minor i18n style drift.**

**D-LOW-003: HRN-001 mentions `codex` via the word `effortLevel` (false-positive for §13 Non-Goals "Codex integration"). Clear on inspection — no real codex reference. Noted for false-alarm prevention.**

**D-LOW-004: Master §10 R5 says "per BC-V3R2-009, -016" — good cross-referencing. But R14 says "`diff -rq` CI check already present" without a file:line reference, consistent with D-HIGH-007.**

---

## Dimension-by-Dimension Results

| Dim | Band | Verdict | Key Finding |
|-----|------|---------|-------------|
| **D1 EARS compliance** | 0.75 | PASS-with-gaps | 758 SHALL across 35 SPECs; D-HIGH-005 notes 5 SPECs with verb-missing REQs; D-HIGH-006 ORC-004 SHOULD over-use. |
| **D2 REQ ID uniqueness** | 0.75 | PASS-with-gaps | Cross-SPEC IDs globally unique; 5 intra-SPEC duplicates (D-HIGH-004). |
| **D3 AC testability** | 0.75 | PASS | 402 Given/When/Then triplets + 147 AC-NNN labels = 549 ACs total across 35 SPECs. Each SPEC has `## 10` Acceptance section. |
| **D4 Dependency DAG** | 0.00 | **FAIL** | 4 cycles (D-CRIT-001) + 2 phase-order violations (D-CRIT-002). |
| **D5 Traceability (file:line)** | 0.50 | FAIL | §10 sections present 35/35, but cite master/pattern sections not file:line. D-HIGH-007. |
| **D6 BC integrity** | 0.25 | **FAIL** | 2 double-claimed BCs (D-CRIT-003/004); 4 orphan BCs (D-CRIT-006); 1 non-canonical ID (D-CRIT-005). |
| **D7 Phase consistency** | 0.25 | **FAIL** | 4 phase mismatches (D-HIGH-001) + 2 phase-order violations (D-CRIT-002). |
| **D8 Principle/Pattern/Problem coverage** | 1.00 | PASS | Appendix A/B/C present; every Critical problem mapped. |
| **D9 Scope discipline** | 0.75 | PASS | §2 "In Scope" / "Out of Scope" present in all 35 SPECs. |
| **D10 Risk coverage** | 1.00 | PASS | Master §10 15 risks with concrete mitigations. |
| **D11 Frontmatter schema** | 0.50 | FAIL | Priority vocab normalized (4 values); phase vocab 8 values (OK); `bc_id` array form uniform; D-CRIT-005 non-canonical ID; D-MED-002/003 field inconsistency. |
| **D12 Internal consistency** | 1.00 | PASS | No "codex", no time estimates, no emoji, no raw TODOs. |
| **D13 Counts accuracy** | 0.50 | FAIL | Master §11 says 35 SPECs (correct) but REQ/AC totals are absent; actual 695/549 is 26% above Wave-5 target. |
| **D14 Fresh-lens adherence** | 0.85 | PASS | T-1 ACI → SPC-004/RT-006; E-1 Agent-as-Judge fresh-memory → HRN-002; O-6 Agentless → WF-004; P1 Contract → SPC-001. Fresh-lens principles are represented. |
| **D15 Layer boundary** | 0.75 | PASS-with-gaps | ORC-005 explicitly cross-layer (D-HIGH-008); HRN-003 reaches into `.moai/specs/` (Layer 2) which is acceptable for scorer. |
| **D16 FROZEN preservation** | 0.50 | FAIL | HRN-002 correctly uses CON-002 amendment protocol (positive). SPC-001 amends SPEC EARS format (FROZEN) without CON-002 dependency (D-CRIT-007). |

---

## Cross-cutting Observations

1. **Systemic pattern: reverse dependencies.** ORC-001 declares deps on ORC-002/003/004 (downstream siblings). The correct pattern is parent SPEC (CON-001) first, then ORC-001 alone; sibling coordination belongs in a program-plan, not the dependency graph. Four cycles stem from this pattern.

2. **BC ownership discipline is broken.** Of 18 master-declared BCs, only 14 are claimed; 2 are double-claimed; 1 non-canonical ID exists. Recommend: build `moai spec bc-check` lint (fits naturally into SPC-003 scope) that enforces bijection between master §8 table and SPEC `bc_id:` fields.

3. **§10 Traceability is qualitative, not quantitative.** The D5 requirement for file:line citations is not met anywhere. Recommend writer pass: add `file:Lstart-Lend` for each master §/research citation.

4. **Scope inflation.** Actual REQ count (695) is 26% above Wave-5 target (~550). Worst contributors: RT-006 (35 REQs), RT-007 (32), RT-002..005 (27 each). Consider collapsing obvious duplicates or factoring out shared REQs into CON-003 consolidation.

5. **FROZEN zone vigilance gap.** Only HRN-002 demonstrates the correct FROZEN-amendment pattern (explicit CON-002 dependency, explicit before/after clause text, explicit graduation-protocol invocation). SPC-001 — which modifies another FROZEN invariant — skips this pattern entirely. This suggests the Wave 4 SPEC writer did not internalize which changes qualify as FROZEN-zone amendments.

---

## Counts & Statistics

| Metric | Expected | Actual | Delta |
|--------|----------|--------|-------|
| SPEC count | 35 | 35 | 0 |
| REQ count | ~550 | 695 | +145 (+26%) |
| AC count | ~350 | 549 (402 GWT + 147 AC-NNN) | +199 (+57%) |
| Breaking SPECs | n/a | 13 | — |
| Unique priorities | 4 | 4 (P0/P1/P2/P3) | 0 |
| Distinct phases | 8 | 8 | 0 |
| Unique BCs claimed | 18 | 14 (+ 1 non-canonical) | −4 orphans |
| Double-claimed BCs | 0 | 2 (002, 006) | +2 |
| DAG cycles | 0 | 4 | +4 |
| Phase-order violations | 0 | 2 | +2 |
| Phase mismatches vs master §9 | 0 | 4 | +4 |

---

## Verdict & Recommendation

**NOT-READY.** 7 Critical defects make the plan not executable as-is. The dependency graph has true cycles, the release plan contradicts per-SPEC phase metadata in two places, and BC ownership is neither injective nor surjective against master §8.

### Ordered fix list (must complete before Iter-2)

1. **Break dependency cycles** (addresses D-CRIT-001).
   - Remove ORC-002, ORC-003, ORC-004 from `SPEC-V3R2-ORC-001/spec.md` `dependencies:` (they are siblings that depend on ORC-001, not the other way around).
   - Remove WF-001 from `SPEC-V3R2-MIG-001/spec.md` `dependencies:` (MIG-001 migrator executes rewrites that reference WF-001's merge map; define the map in a shared artifact or have MIG-001 read the skill rename table from a file produced by WF-001's landing commit — not as a dep).

2. **Resolve phase-order violations** (D-CRIT-002).
   - HRN-001 should not depend on MIG-003. Invert: either MIG-003 depends on HRN-001 (loader for harness.yaml is added as part of Phase 5) OR promote MIG-003 sub-scope (the harness.yaml loader) into HRN-001.
   - WF-001 should not depend on MIG-001. Invert.

3. **Fix BC claims** (D-CRIT-003/004/005/006).
   - Remove `BC-V3R2-002` from `SPEC-V3R2-RT-002/spec.md` (RT-002 does not map to BC-V3R2-002 per master §8). Assign RT-002 a correct BC or leave `bc_id: []`.
   - Remove `BC-V3R2-006` from `SPEC-V3R2-ORC-001/spec.md` `bc_id:` array. Keep only in WF-001.
   - Rename `BC-V3R2-EXT004` → use a canonical `BC-V3R2-NNN` ID from the reserved pool (next available is 019 if extending master, or reuse 007 if EXT-004 semantically covers "Agentless flip"; coordinate with master §8 update).
   - Claim the 4 orphan BCs:
     - BC-V3R2-004 (Agent tool removal) → assign to ORC-002.
     - BC-V3R2-007 (Agentless flip) → assign to WF-004.
     - BC-V3R2-012 (/98 /99 extraction) → assign to WF-002.
     - BC-V3R2-013 (config loaders) → assign to MIG-003 (and/or HRN-001 if the harness loader is separated).

4. **Route SPC-001 through CON-002 graduation** (D-CRIT-007).
   - Add `SPEC-V3R2-CON-002` to `SPEC-V3R2-SPC-001/spec.md` `dependencies:`.
   - Insert explicit "FROZEN-zone amendment" section describing before/after SPEC EARS schema text, analogous to HRN-002.

5. **Reconcile per-SPEC phase with master §9** (D-HIGH-001).
   - Either update `SPEC-V3R2-RT-005/spec.md` phase to Phase 1, or update master §9 row Phase 1 to drop RT-005 and insert it into Phase 2.
   - Same reconciliation for ORC-005, WF-005, WF-006.

6. **Add file:line citations in §10** (D-HIGH-007). One pass across all 35 SPECs replacing `master-v3 §N` with `docs/design/major-v3-master.md:Lstart-Lend` (and similarly for research files).

7. **Resolve intra-SPEC REQ ID duplicates** (D-HIGH-004) — EXT001-004, EXT001-006, EXT004-004, WF003-016, WF006-013.

8. **Normalize missing normative verbs** (D-HIGH-005) — one additional SHALL per flagged REQ.

9. **Add REQ/AC count totals to master §11** (D-HIGH-003).

10. **Decide EXT-003 disposition** (D-HIGH-002/D-HIGH-009) — either mark as `lifecycle: design-only` with placeholder REQ set, or promote to proper Phase-9 execution (outside v3.0.0-GA scope).

Iter-2 should focus exclusively on these 10 items. All Low/Medium defects may be deferred to Iter-3 unless Iter-2 fixes incidentally address them.

---

## Chain-of-Verification Pass

Second-look re-checks performed:

- Re-read all 35 `bc_id:` lines. Confirmed 14 assigned + 1 non-canonical + 4 orphans (D-CRIT-005/006).
- Re-read all 35 `dependencies:` blocks. Confirmed 4 cycles via DFS (D-CRIT-001).
- Re-read master §9 L987-994 line-by-line against each SPEC's `phase:`. Confirmed 4 mismatches (D-HIGH-001).
- Re-checked SPC-001 dependency list (`SPEC-V3R2-SPC-001/spec.md:L13-14`) — CON-001 only, no CON-002. FROZEN-zone amendment without graduation protocol confirmed (D-CRIT-007).
- Re-checked HRN-002 as positive control — `SPEC-V3R2-HRN-002/spec.md:L15-18` explicitly depends on CON-001 AND CON-002. The pattern is available; SPC-001 just did not follow it.
- Spot-re-checked §10 Traceability for 3 SPECs (CON-001 L174-184, RT-001 L204+, WF-003 L230+) — no file:line citations anywhere. D-HIGH-007 confirmed systemic.

No new defects surfaced in second pass.

---

## Regression Check

Not applicable — Iteration 1.

---

End of report.
