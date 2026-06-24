# plan.md — SPEC-V3R6-SESSION-HANDOFF-SSOT-ALIGN-001

> Implementation plan for the SSOT↔render drift unification. Tier S, doc-convention, new Go code = 0.

## §A. Context

See spec.md §A/§B. The two surfaces (session-handoff.md SSOT + moai.md §8 render surface) drifted in 3 real places (D3/D5a/D6/D9) because duplicated content is maintained manually. Track 1 closed tree-pair + file-internal defects (D1/D2/D7/D8) and the SSOT↔render axis recurred because there is no mechanical prevention. This plan designs the recurrence-mitigation mechanism and decomposes the work into milestones.

> **iter-2 correction (per plan-auditor falsification verdict)**: the original iter-1 thesis over-promised by calling the SSOT-pointer + sentinel mechanism a "recurrence-prevention" mechanism. The plan-auditor correctly falsified this: a future editor who changes one surface without reading the other surface's sentinel produces silent drift — no test fails, no CI breaks. The mechanism is a **mitigation + visibility** convention (surfaces drift to a reading editor), NOT a **prevention** guarantee. This iter-2 revision downgrades the language throughout. The honest residual-risk note in §F.5 is preserved (plan-auditor confirmed it honest per verification-claim-integrity §3.5). The only mechanical catch is the deferred Go lint rule (§F.6).

## §B. Known Issues (from baseline, iter-2)

| ID | Surface | Site | Defect |
|----|---------|------|--------|
| D3 | session-handoff.md | L93, L111, L198 | retired `wave` filename token + memory-filename examples (moai.md §8 already clean) |
| D5a | session-handoff.md L254 vs moai.md §8 L312 vs L681 | 3 sites | "Pre-emit self-check" label collision — 3 different concerns under one name |
| ~~D5b~~ | ~~moai.md §8 L681~~ | ~~label "all 9" vs 8 enumerated items~~ | **WITHDRAWN (iter-2)** — the label "all 9" and the enumerated count ALREADY AGREE (9=9); the iter-1 "8 items" claim was a windowed-grep undercount that excluded L682. See research.md §A.5 for the honest mea culpa. |
| D6 | moai.md §8 + session-handoff.md Localization Table | Localization Table section | no explicit fallback rule for non-en/ko/ja/zh locales |
| D9 | session-handoff.md | 3 separated catalogues | 3 separated anti-pattern catalogues, no single index |
| ~~D-conflation~~ | ~~session-handoff.md Localization Table~~ | ~~hypothetical 16-prog-lang conflation~~ | **DROPPED (iter-2)** — the doctrine already cleanly separates the `conversation_language` (ISO-639) axis from the 16-programming-language axis; zero baseline conflation to guard against. See research.md §A.6a. |

## §C. Pre-flight (verifications the implementer runs before editing)

1. Confirm byte-parity baseline holds: `cmp .claude/rules/moai/workflow/session-handoff.md internal/template/templates/.claude/rules/moai/workflow/session-handoff.md && cmp .claude/output-styles/moai/moai.md internal/template/templates/.claude/output-styles/moai/moai.md` — both must exit 0.
2. Re-run the baseline content greps (research.md §A) and confirm the defect sites are present. **Anchor by content pattern, NOT by line number** (line-number drift asymmetry — LOCAL may be N lines longer than TEMPLATE; the same content sits at different line numbers in each tree; see feedback memory `feedback_line_number_drift_asymmetry` and the auto-memory `feedback_line_number_drift_asymmetry` chain). Line numbers in research.md are secondary attribution only.
3. Confirm no sibling SPEC AC cites the bare "Pre-emit self-check" label as a literal (grep across `.moai/specs/`).
4. ***(iter-2 NEW)* For every MUST AC, run the AC grep against the pre-fix baseline and record the baseline output (must be 0 matches or otherwise FAIL). This is the vacuous-pass guard — if any MUST AC PASSES at baseline, the AC is defective and must be re-anchored before implementation proceeds.** The iter-1 SPEC had 3 vacuous-pass ACs (D3/D4/D5 in the audit report); the iter-2 acceptance.md re-anchors all of them with baseline-failing greps, and this pre-flight step verifies the re-anchoring is correct.

## §D. Constraints

- New Go code = 0 (spec.md §E constraint 1).
- Byte-parity for both files (spec.md §E constraint 2).
- GEARS format (spec.md §E constraint 3).
- Mechanically verifiable ACs (spec.md §E constraint 4).
- `era: V3R6` explicit override (spec.md §E constraint 5).

## §E. Self-Verification (manager-develop §E analog)

This is a doc-convention SPEC; the §E self-verification is grep/diff-based, not go-test-based. The implementer MUST produce verbatim command + output for:

- E1 — every modified defect site (D3/D5a/D6/D9) shows the fix in place (grep). *(iter-2: D5b removed — no fix site for a non-defect.)*
- E2 — byte-parity holds for both modified files (`cmp` exits 0).
- E3 — self-check sentinel present on BOTH surfaces (grep).
- E4 — SSOT-pointer present in moai.md §8 Session Handoff block (grep).
- E5 — D5a label disambiguation: grep A (concern-name presence) returns ≥3 matches post-fix (0 at baseline); grep B (count-only qualifier absence) returns 0 matches post-fix (1 at baseline). Both greps fail at baseline — no vacuous-pass.
- E6 — D9 consolidated index present at one canonical location with forward links to all AP codes.
- E7 — ***(iter-2 NEW, vacuous-pass guard)*** for every MUST AC (AC-SHA-001/003/004/005/006/007), the baseline output AND the post-fix output are both recorded. The baseline output MUST demonstrate the AC FAILS at baseline (0 matches where ≥N is required, or ≥N where 0 is required). The post-fix output MUST demonstrate the AC PASSES. This is the direct application of the D3/D4/D5 lesson from plan-auditor iter-1.

## §F. Recurrence-Mitigation Mechanism Design (core design challenge)

This section is the heart of the SPEC. Track 1 closed 4 defects and the SSOT↔render drift recurred because there is no mechanical prevention. The constraint is doc-only (new Go code = 0), so a Go lint rule is OUT OF SCOPE (that would be a separate code SPEC). The mechanism must live inside the doctrine files themselves.

> **iter-2 scope-honesty note**: per the plan-auditor iter-1 falsification verdict, this mechanism is a **mitigation + visibility** convention — it surfaces drift to a reading editor but does not mechanically prevent it. The §F.5 residual-risk note is preserved as honest. The §F.6 deferred lint rule is the only mechanical catch.

### F.1 Candidate approaches (evaluated, not all selected)

| # | Approach | Trade-off |
|---|----------|-----------|
| 1 | **SSOT-pointer pattern** — reduce moai.md §8 Session Handoff block to a cross-reference to session-handoff.md | (+) eliminates duplicated content that drifts; (−) moai.md §8 is the RENDER surface the orchestrator reads at every output; a pure pointer degrades render-time usability (extra hop to the SSOT for the 6-block skeleton). |
| 2 | **Self-check sentinel** — add a numbered self-check item in BOTH files verifying the other surface's item count / table row count / filename-token vocabulary matches this surface's | (+) manual-but-attributable parity check, low-cost; (−) still manual — a future editor can ignore the sentinel. |
| 3 | **Single canonical vocabulary table** — one SSOT table that moai.md §8 references rather than duplicates | (+) any label/locale/filename token defined once; (−) alone it does not cover the 6-block skeleton or the self-check labels, only the vocabulary tokens. |
| 4 | **Hybrid** — pointer for stable content + thin render-only additions in moai.md §8 explicitly marked "render-only, not canonical" | (+) preserves render usability while eliminating canonical duplication; (−) requires a clear marker convention ("render-only") that itself must be respected. |

### F.2 Falsification question for each approach

The chosen mechanism must be defensible against: **"if a future editor changes one surface, will this mechanism catch the drift?"**

- **Approach 1 (pure pointer)**: catches drift by *eliminating* the duplicated content — there is nothing to drift because the canonical lives in one place. Falsification: a future editor copies content back into moai.md §8 "for convenience", recreating the duplication. The pure pointer has no marker saying "do not duplicate".
- **Approach 2 (sentinel alone)**: catches drift by *flagging* it — the sentinel on the other surface fails. Falsification: the editor ignores the sentinel or deletes it. No mechanical enforcement.
- **Approach 3 (vocab table)**: catches drift only for vocabulary tokens, not for skeletons or self-checks. Insufficient alone.
- **Approach 4 (hybrid)**: catches drift by *eliminating canonical duplication* (pointer) + *flagging residual duplication* (sentinel verifying the "render-only" additions do not re-introduce canonical content). The "render-only, not canonical" marker is itself a sentinel — a future editor copying canonical content into a render-only block violates the marker.

### F.3 Recommendation: Hybrid (Approach 4 = Approach 1 + Approach 2)

**Primary mechanism**: SSOT-pointer + self-check sentinel (hybrid).

- The moai.md §8 Session Handoff block reduces its duplicated canonical content (6-block skeleton, cut-line marker spec, Localization Table, anti-pattern catalogue) to cross-references to the SSOT (session-handoff.md), **plus** a thin render-specific addition (the orchestrator-facing self-check items needed at print time) explicitly marked `<!-- render-only, not canonical — canonical lives in session-handoff.md -->`.
- Each surface carries a numbered self-check sentinel verifying the other surface's structural invariant (Localization Table row count, self-check item count, filename-token vocabulary).

**Why hybrid over the alternatives**:
- Pure pointer (Approach 1) degrades render usability — the orchestrator reads moai.md §8 at every output and a pure pointer forces a hop to the SSOT for the skeleton.
- Sentinel alone (Approach 2) is insufficient — it flags drift but does not eliminate the duplicated content that drifts.
- Vocab table (Approach 3) is too narrow — only covers tokens, not skeletons or self-checks.
- Hybrid (Approach 4) eliminates canonical duplication (pointer) AND flags any re-introduction (sentinel + "render-only" marker), preserving render usability.

### F.4 Implementation pattern for the hybrid

For each duplicated canonical block in moai.md §8:

```
<!-- render-only, not canonical — canonical lives in .claude/rules/moai/workflow/session-handoff.md §<section-name> -->
> **Session Handoff 6-block skeleton**: see `.claude/rules/moai/workflow/session-handoff.md` § Canonical Format.
> Render-time additions (orchestrator-facing self-check items):
- [ ] <render-specific-item-1>
- [ ] <render-specific-item-2>
```

The SSOT (session-handoff.md) carries the canonical skeleton + its own self-check. The render surface carries the pointer + render-only additions + its own self-check sentinel. The sentinel in each file verifies the *other* file's invariant.

### F.5 Residual risk (honest, 5-section evidence format per verification-claim-integrity §3)

- **Claim**: the hybrid mechanism **mitigates** recurrence of SSOT↔render drift and **surfaces** drift to a reading editor. *(iter-2 CORRECTION: the prior "prevents recurrence" wording was downgraded per plan-auditor iter-1 falsification — the mechanism is mitigation + visibility, not mechanical prevention.)*
- **Evidence**: the mechanism eliminates duplicated canonical content (pointer) and flags re-introduction (sentinel + "render-only" marker). The self-check sentinel in each file references the other file's structural invariant by name.
- **Baseline-attribution**: the 3 real defect sites (D3/D5a/D6/D9) are the measured baseline (research.md §A). Each is a site of duplicated content that drifted; the hybrid eliminates the duplication. *(iter-2: D5b removed — not a real defect.)*
- **Gaps**: the mechanism is doc-only; a future editor can ignore or delete the sentinel. The "render-only, not canonical" marker is a convention, not a mechanical enforcement. A code-level lint rule that greps for canonical content outside the SSOT would close this gap mechanically — but that is out of scope (new Go code = 0) and is flagged as a deferred follow-up SPEC (§F.6).
- **Residual-risk**: the hybrid raises the visibility bar (visible drift signal + marker convention) but does **not** *guarantee* prevention. If a future editor copies canonical content into a render-only block AND deletes the sentinel, drift recurs silently. The deferred lint-rule follow-up is the only mechanical guarantee. This is honestly admitted here per verification-claim-integrity §3.5.

### F.6 Deferred follow-up (NOT in this SPEC)

A code-level lint rule under `internal/spec/` that mechanically detects canonical Session Handoff content outside the SSOT (session-handoff.md) would close the residual-risk gap. This is explicitly out of scope for this SPEC (new Go code = 0) and is recommended as a follow-up code SPEC after this doc-convention SPEC lands.

## §G. Milestones

> *iter-2 renumbering*: iter-1 had 6 milestones (M1–M6) covering 9 REQs. iter-2 drops the D5b work item (M2's count-fix half) and the D-conflation work item (was inside M3), so the milestones compress to 5. REQ references are renumbered to match spec.md §D (REQ-SHA-001..007).

| ID | Milestone | Scope | Priority |
|----|-----------|-------|----------|
| M1 | D3 `wave` token migration | REQ-SHA-001 + REQ-SHA-002: migrate `<wave>` filename token + memory-filename examples to sprint/round; preserve generic-English-noun `wave` in prose. Both trees. | High |
| M2 | D5a label disambiguation (concern-names) | REQ-SHA-003: qualify "Pre-emit self-check" labels with surface-owned **concern-names** (paste-ready budget / localization render / session-handoff template completeness); replace count-only `(8 items)` qualifier at session-handoff.md. Both trees. *(iter-2: D5b count-fix work-item DROPPED — the count already agrees at 9.)* | High |
| M3 | D6 localization fallback codification | REQ-SHA-004: add explicit fallback rule to session-handoff.md Localization Table section, using a baseline-absent phrase; cross-reference moai.md §8 "any other ISO-639 code" clause to the SSOT fallback rule. Both trees. *(iter-2: the D-conflation work-item is DROPPED — doctrine already clean.)* | High |
| M4 | D9 consolidated anti-pattern index | REQ-SHA-005: add single index table at one canonical location in session-handoff.md with forward links to AP-D-001..005 + AP-V-001..004 + general-hygiene items. Both trees. | Medium |
| M5 | Recurrence-mitigation mechanism (hybrid) + self-verification + byte-parity | REQ-SHA-006 + REQ-SHA-007: apply SSOT-pointer pattern to moai.md §8 duplicated canonical content; add self-check sentinels to both surfaces; mark render-only additions. Then §E self-verification E1–E7 (including the vacuous-pass baseline-vs-post-fix comparison for every MUST AC); byte-parity `cmp` for both files both trees; AC sweep. Both trees. | High |

### Sequencing

M1→M2→M3→M4 are independent content edits (different defect sites) and may be batched. M5 (the recurrence-mitigation mechanism + verification) is applied AFTER M1–M4 so the sentinels verify the *final* structural invariants, not intermediate state. M5 includes the verification gate.

## §H. Anti-Patterns (for this plan)

- AP-PLAN-001 — adding a Go lint rule to "make the sentinel mechanical". This violates new-Go-code=0. The lint rule is a deferred follow-up SPEC, not part of this SPEC.
- AP-PLAN-002 — rewriting the generic-English-noun `wave` in prose (L9/L20/L80). REQ-SHA-002 explicitly preserves it; only the filename token is retired.
- AP-PLAN-003 — eliminating the moai.md §8 render-only additions in the name of "pure pointer". This degrades render usability; the hybrid explicitly preserves render-only additions marked as such.
- AP-PLAN-004 — citing line numbers in ACs without content-pattern fallback. Line-number drift asymmetry (LOCAL vs TEMPLATE) means the same content sits at different line numbers in each tree; ACs must be grep-content-based, with line numbers as secondary attribution only.

## §I. Cross-References

- spec.md: `.moai/specs/SPEC-V3R6-SESSION-HANDOFF-SSOT-ALIGN-001/spec.md`
- acceptance.md: `.moai/specs/SPEC-V3R6-SESSION-HANDOFF-SSOT-ALIGN-001/acceptance.md`
- research.md: `.moai/specs/SPEC-V3R6-SESSION-HANDOFF-SSOT-ALIGN-001/research.md`
- Track 1 predecessor: `.moai/specs/SPEC-SESSION-HANDOFF-ALIGN-001/` (completed, commit 82ea1b09f)
- Governing naming rule: `.claude/rules/moai/development/sprint-round-naming.md` AP-SRN-004 (Wave→Round/Sprint retirement)
- Verification doctrine: `.claude/rules/moai/core/verification-claim-integrity.md` §2 (baseline-attribution)
- Era classification: `.claude/rules/moai/workflow/lifecycle-sync-gate.md` (H-override via `era: V3R6` frontmatter)
