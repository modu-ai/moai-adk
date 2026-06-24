# research.md — SPEC-V3R6-SESSION-HANDOFF-SSOT-ALIGN-001

> Baseline-attributed evidence. Every drift claim cites a re-run grep with observed output (verification-claim-integrity §2). Numbers are not hand-copied from the orchestrator's prompt — they are re-measured here.

## §A. Baseline Measurements (re-run, attributed)

### §A.1 D3 — retired `wave` filename token (session-handoff.md, LOCAL tree)

**Command**:
```bash
grep -nE 'wave|Wave|WAVE' .claude/rules/moai/workflow/session-handoff.md
```

**Observed output** (re-run 2026-06-17):
```
9:Long workflows (multi-SPEC waves, multi-milestone implementation) accumulate context...
20:| 4 | PR creation success when more SPECs remain in the current wave | After `gh pr create` success...
80:- **Block 6**: separator + ... — RECOMMENDED for multi-SPEC waves or follow-up...
93:applied lessons: project_wave6_myproj001_plan_ready, lessons #9 wave-split.
111:1. Save the message to a memory project entry. Filename pattern: `project_<wave>_<spec>_<status>.md` (e.g., `project_wave6_wf002_complete.md`).
197:ultrathink. SPEC-MYPROJ-001 Wave N 진입.
198:applied lessons: project_myproj_prev_wave_complete, lessons #12 #13 #14.
319:- `feedback_large_spec_wave_split.md` (auto-memory) — wave-split rationale
```

**Classification**:
- **Filename-token defect (in scope for REQ-SHA-001)**: L93 (`project_wave6_myproj001_plan_ready`), L111 (`project_<wave>_<spec>_<status>.md` + `project_wave6_wf002_complete.md`), L198 (`project_myproj_prev_wave_complete`).
- **Example-block noun (judgment call)**: L197 (`Wave N 진입` — part of the Block 1 example; migrate to `Sprint N 진입` or `Round N 진입` per the new vocabulary).
- **Generic-English-noun prose (preserved per REQ-SHA-002)**: L9 ("multi-SPEC waves"), L20 ("current wave"), L80 ("multi-SPEC waves").
- **Auto-memory filename reference (out of scope)**: L319 (`feedback_large_spec_wave_split.md` — this is an actual memory filename that exists on disk; renaming it is out of scope for this doc-convention SPEC, which modifies doctrine prose only).

### §A.2 D3 — `wave` token in moai.md §8 (render surface, LOCAL tree)

**Command**:
```bash
grep -nE 'wave|Wave|WAVE' .claude/output-styles/moai/moai.md
```

**Observed output**:
```
(empty — 0 matches)
```

**Conclusion**: moai.md §8 is already clean of the `wave` token. D3 is a session-handoff.md-only defect on this axis. The fix is applied to session-handoff.md (both trees); moai.md needs no D3 edit.

### §A.3 D3 — `wave` token in session-handoff.md (TEMPLATE tree)

**Command**:
```bash
grep -nE 'wave|Wave|WAVE' internal/template/templates/.claude/rules/moai/workflow/session-handoff.md
```

**Observed output**: identical to §A.1 (same 7 lines at same line numbers). Byte-parity holds at baseline.

### §A.4 D5 — "Pre-emit self-check" label collision

**Command (session-handoff.md LOCAL)**:
```bash
grep -nE 'Pre-emit self-check|self-check|Self-check' .claude/rules/moai/workflow/session-handoff.md
```

**Observed output**:
```
254:### Pre-emit self-check (8 items)
316:- `.claude/output-styles/moai/moai.md` §8 ... the canonical render surface for the 6-block template + pre-emit self-check; ...
```

**Command (moai.md LOCAL)**:
```bash
grep -nE 'Pre-emit self-check|self-check|Self-check' .claude/output-styles/moai/moai.md
```

**Observed output**:
```
190:Format and self-check rules: see §8 Session Handoff Template below.
312:**Pre-emit self-check (verify before printing any §8-derived block):**
681:Pre-emit self-check (MUST verify all 9 before printing):
722:- [HARD] Pre-emit self-check: every banner/template-derived block MUST pass §8 Localization Contract self-check before printing.
```

**Conclusion**: the bare label "Pre-emit self-check" appears at 3 distinct sites with distinct concerns (this is the **D5a label-collision defect** — a real defect, in scope for REQ-SHA-003):
- session-handoff.md L254 — **paste-ready budget self-check** (8 items).
- moai.md L312 — **localization render self-check** (9 items; see §A.5a below).
- moai.md L681 — **session-handoff template completeness self-check** (9 items; label claims "all 9" and the enumerated count IS 9 — see §A.5 for the iter-1 mea culpa).

### §A.5a D5 (L312 label) — count verification (no drift here either)

For completeness, the L312 localization-render self-check item count was verified (re-run 2026-06-17):
```bash
sed -n '314,322p' .claude/output-styles/moai/moai.md | grep -c '^- \['
```
Observed output: `9`. No label-vs-count drift at L312 either. The only D5-axis defect is the **label collision** (D5a), not any count drift.

### §A.5 D5b — WITHDRAWN (windowed-grep undercount; label and enumerated count already agree at 9)

**Iter-1 mea culpa.** The original §A.5 here claimed a label-vs-count drift at moai.md §8 L681 ("MUST verify all 9 before printing" vs allegedly 8 enumerated items). That claim was the product of a **windowed-grep undercount**. The prior command `sed -n '683,696p' ... | grep -c '^- \['` started its window at L683, which **excluded the first checkbox item at L682**. The correct count is **9 items at L682–L690**, matching the label "all 9" exactly.

This is a verification-claim-integrity §2 baseline-attribution defect — the SPEC's own measurement was the error source, not the orchestrator's. The orchestrator's delegation prompt flagged the count as 9; the plan-auditor's iter-1 independent re-measurement (3 separate methods) confirmed 9. The SPEC's prior "8" was wrong.

**Corrected measurement (re-run 2026-06-17, 3 methods converging on 9)**:

Method 1 — content-anchored grep, no line-number window (authoritative):
```bash
grep -nE '^- \[ \]' .claude/output-styles/moai/moai.md | awk -F: '$1>=675 && $1<=705'
```
Observed output (9 lines, L682–L690):
```
682:- [ ] Block 1 starts with `ultrathink.` ...
683:- [ ] Block 2 lists ≥1 memory file ...
684:- [ ] Block 2 includes source_session_id ...
685:- [ ] Block 4 has ≤4 preconditions ...
686:- [ ] Block 5 is a single primary action ...
687:- [ ] L3 worktree Block 0 ...
688:- [ ] Cut-line markers present ...
689:- [ ] Block 6 workflow-context header ...
690:- [ ] Block 2 source_session_id environment fallback ...
```

Method 2 — awk NR window:
```bash
awk 'NR>681 && NR<693 && /^- \[/{c++} END{print c}' .claude/output-styles/moai/moai.md
```
Observed output: `9`.

Method 3 — wider sed window starting AT the label line (not after it):
```bash
sed -n '681,700p' .claude/output-styles/moai/moai.md | grep -cE '^- \['
```
Observed output: `9`.

**Conclusion**: the label "MUST verify all 9 before printing" at L681 and the enumerated 9 items at L682–L690 **already agree (9 = 9)**. There is **no D5b defect**. REQ-SHA-004 and AC-SHA-004 (which codified this non-defect) are **DELETED** in this iter-2 revision.

**Lesson (recurrence-prevention for the SPEC's own authoring process)**: hard line-number windows in AC greps are the mechanism by which this false-defect entered. The prior `sed -n '683,696p'` window excluded the very item (L682) that would have disproven the claim. All surviving ACs in this iter-2 revision anchor by **content pattern**, not line number — see D7 correction. This is the same line-number-drift-asymmetry lesson recorded in the auto-memory `feedback_line_number_drift_asymmetry`, now applied to the SPEC's own measurement process.

### §A.6 D6 — localization fallback rule gap

**Command (session-handoff.md LOCAL Localization Table region)**:
```bash
grep -nE 'Localization|ISO-639|fallback|Fallback|conversation_language' .claude/rules/moai/workflow/session-handoff.md
```

**Observed output** (relevant rows):
```
54:- Marker text translates per `conversation_language` (see Localization table below)
56:### Localization Table
58:The cut-line marker text AND the 6-block skeleton verbs/headers translate per `conversation_language`. This table is the SSOT for the locale renderings ...
70:Read `conversation_language` from `.moai/config/sections/language.yaml` at render time; substitute the localized text ...
```

**Observation**: the Localization Table (L56–L70) covers en/ko/ja/zh only (per the orchestrator's baseline). There is NO explicit fallback rule for locales outside that set. The word "fallback" appears at L76 but refers to the *source_session_id environment fallback* (a different concern), NOT locale fallback.

**Command (moai.md §8 L240 — the only coverage for non-en/ko/ja/zh)**:
```bash
sed -n '240p' .claude/output-styles/moai/moai.md
```

**Observed output**:
```
- If `ko` / `ja` / `zh` / any other ISO-639 code: translate every label listed above into that language naturally — use idiomatic phrasing that a native reader would expect, not literal word-by-word translation
```

**Conclusion**: the "any other ISO-639 code" clause at L240 is the ONLY coverage for non-en/ko/ja/zh locales, and it does NOT state (a) the canonical fallback skeleton (English) or (b) what "naturally" means concretely. D6 = codify the explicit fallback rule in the SSOT (REQ-SHA-004, *renumbered from iter-1 REQ-SHA-005*) and cross-reference it from moai.md §8.

### §A.6a D-conflation hypothesis — DROPPED (iter-2; doctrine already clean)

> *iter-2 addition.* The iter-1 SPEC included a REQ-SHA-006 demanding the SSOT "shall not conflate the `conversation_language` axis with the 16-programming-language axis". The plan-auditor iter-1 report (D4) flagged this as a vacuous-absence grep: at baseline there is NO conflation to guard against, so the AC PASSES trivially at merge time without any fix being applied.

**Command (verify zero baseline conflation in session-handoff.md)**:
```bash
grep -nE 'programming.language|16.*language|16-programming' .claude/rules/moai/workflow/session-handoff.md
```

**Observed output**: *(empty — 0 matches)*.

**Command (verify moai.md §8 cleanly uses ISO-639 framing)**:
```bash
grep -nE 'ISO-639|any other ISO' .claude/output-styles/moai/moai.md | head -5
```

**Observed output** (relevant row):
```
- If `ko` / `ja` / `zh` / any other ISO-639 code: translate every label listed above into that language naturally ...
```

**Conclusion**: session-handoff.md contains **zero mentions** of the 16-programming-language axis in any region (and zero in the Localization Table region specifically). moai.md §8 cleanly frames the locale axis as ISO-639 throughout. The two axes (`conversation_language` ISO-639 vs the 16 programming languages) are **already cleanly separated** in the doctrine. There is no conflation to prevent. REQ-SHA-006 / AC-SHA-006 (iter-1) are **DROPPED** in this iter-2 revision as addressing a hypothetical non-defect. The corresponding exclusion is recorded in spec.md §H.

### §A.7 D9 — 3 separated anti-pattern catalogues in session-handoff.md

**Command**:
```bash
grep -nE 'Anti-Pattern|Anti-pattern|AP-D-|AP-V-' .claude/rules/moai/workflow/session-handoff.md
```

**Observed output** (structural rows):
```
123:## Anti-Patterns
125:> See also: § Diet Constraints / Anti-pattern catalogue (AP-D-001..005) and § V0 Abort Gate Doctrine / Anti-pattern (AP-V-001..004). ...
244:### Anti-pattern catalogue
246:> See also: § Anti-Patterns ... and § V0 Abort Gate Doctrine / Anti-pattern (AP-V-001..004). This catalogue covers paste-ready budget violations (AP-D-001..005).
248:- **AP-D-001**: ...
249:- **AP-D-002**: ...
250:- **AP-D-003**: ...
251:- **AP-D-004**: ...
252:- **AP-D-005**: ...
303:### Anti-pattern
305:> See also: § Anti-Patterns ... and § Diet Constraints / Anti-pattern catalogue (AP-D-001..005). This catalogue covers abort-gate violations (AP-V-001..004).
307:- **AP-V-001**: ...
308:- **AP-V-002**: ...
309:- **AP-V-003**: ...
310:- **AP-V-004**: ...
```

**Conclusion**: 3 separated catalogues (L123 general hygiene, L244 AP-D, L303 AP-V) with mutual "See also" cross-references but no single index. REQ-SHA-005 *(renumbered from iter-1 REQ-SHA-007)* adds a single consolidated index table at one canonical location; detail sections remain in place.

### §A.8 Byte-parity baseline (both files both trees)

**Command**:
```bash
wc -c .claude/rules/moai/workflow/session-handoff.md \
     internal/template/templates/.claude/rules/moai/workflow/session-handoff.md \
     .claude/output-styles/moai/moai.md \
     internal/template/templates/.claude/output-styles/moai/moai.md
cmp .claude/rules/moai/workflow/session-handoff.md \
    internal/template/templates/.claude/rules/moai/workflow/session-handoff.md
echo "session-handoff cmp exit: $?"
cmp .claude/output-styles/moai/moai.md \
    internal/template/templates/.claude/output-styles/moai/moai.md
echo "moai cmp exit: $?"
```

**Observed output**:
```
   23648 .claude/rules/moai/workflow/session-handoff.md
   23648 internal/template/templates/.claude/rules/moai/workflow/session-handoff.md
   51018 .claude/output-styles/moai/moai.md
   51018 internal/template/templates/.claude/output-styles/moai/moai.md
  149332 total
session-handoff cmp exit: 0
moai cmp exit: 0
```

**Conclusion**: byte-parity holds at baseline for both files. AC-SHA-007 *(renumbered from iter-1 AC-SHA-009)* verifies it holds post-change.

## §B. Scope-Boundary Verification (3 existing SPECs)

### §B.1 SPEC-SESSION-HANDOFF-ALIGN-001 (completed, commit 82ea1b09f)

**Command**:
```bash
head -20 .moai/specs/SPEC-SESSION-HANDOFF-ALIGN-001/spec.md
```

**Observed key fields**:
```
id: SPEC-SESSION-HANDOFF-ALIGN-001
title: "session-handoff.md template↔local drift closure + mirror-coverage gap + i18n/dedup debt"
status: completed
module: ".claude/rules/moai/workflow/session-handoff.md + internal/template mirror + rule_template_mirror_test.go"
```

**Conclusion**: that SPEC's axis is *template↔local tree drift* (drift between the two TREES) + file-internal D1/D2/D7/D8. My SPEC's axis is *SSOT↔render drift* (drift between two FILES within one tree). Orthogonal. No overlap. spec.md §B.2 cites this boundary explicitly.

### §B.2 SPEC-V3R6-SESSION-HANDOFF-AUTO-001 (completed)

**Observed key fields**:
```
id: SPEC-V3R6-SESSION-HANDOFF-AUTO-001
title: "Session handoff auto-persistence (paste-ready resume → memory + MEMORY.md)"
module: "internal/hook, internal/hook/handoff"
```

**Conclusion**: code SPEC modifying Go hook code. No overlap with my doc-convention SPEC (new Go code = 0).

### §B.3 SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 (completed)

**Observed key fields**:
```
id: SPEC-V3R6-SESSION-LEGACY-COVERAGE-001
title: "internal/session 패키지 test coverage 보강 (77.7% → ≥85%, test-only)"
module: "internal/session"
```

**Conclusion**: test-only SPEC under `internal/session`. No overlap.

## §C. Track 3 Boundary (D4 — NOT absorbed)

D4 = source_session_id 91% fallback dead-feature diagnosis. This is a **code-level** defect rooted in `internal/hook`, `internal/cli/session`, and the multi-session coordination layer. It is scoped to a separate Tier M SPEC (SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001). This SPEC mentions `source_session_id` only as a *content* item that the SSOT↔render surfaces must agree on (the session-handoff template completeness self-check includes a "Block 2 source_session_id environment fallback" item); it does NOT diagnose or repair the dead-feature. The two tracks are independent and may run in parallel.

## §D. Recurrence-Mitigation Mechanism Analysis

> *iter-2 correction*: the section heading is downgraded from "Recurrence-Prevention" to "Recurrence-Mitigation" per the plan-auditor iter-1 falsification verdict (the mechanism surfaces drift to a reading editor; it does not mechanically prevent it).

See plan.md §F for the full falsification analysis of the 4 candidate approaches. Summary:

| Approach | Catches drift by… | Falsification residual |
|----------|-------------------|-----------------------|
| 1. SSOT-pointer | eliminating duplicated content | future editor copies content back "for convenience" |
| 2. Self-check sentinel | flagging drift on the other surface | editor ignores or deletes the sentinel |
| 3. Single vocab table | defining tokens once | too narrow — only tokens, not skeletons/self-checks |
| 4. Hybrid (1+2) | eliminating canonical duplication + flagging re-introduction | editor ignores sentinel AND violates "render-only" marker |

**Recommended primary**: Hybrid (Approach 4). Justification in plan.md §F.3. Residual risk honestly recorded in plan.md §F.5 — the mechanism raises the visibility bar but is **not** a mechanical guarantee *(iter-2 correction — prior wording over-promised)*; the deferred lint-rule follow-up (plan.md §F.6) is the only mechanical catch.

## §E. Gaps (what this research did NOT verify)

- **Render-time usability of the hybrid mechanism**: the claim that the hybrid preserves the orchestrator's ability to read moai.md §8 without an extra hop is a design property (plan.md §F.3), not a directly-greppable measurement. The plan-auditor evaluates it via the trade-off analysis.
- **Sentinel effectiveness against future drift**: the sentinel is manual; AC-SHA-006 *(renumbered from iter-1 AC-SHA-008)* verifies presence at merge time but cannot prevent future deletion. This is the honest residual risk.
- **Sibling SPEC ACs citing the bare "Pre-emit self-check" label**: a grep across `.moai/specs/` for the bare label was not run in this research. The implementer runs it in M2 pre-flight (plan.md §C step 3); if any sibling SPEC AC cites the bare label as a literal, the qualified-name refactor in REQ-SHA-003 must update that citation too.

## §F. Residual Risk

- The hybrid mechanism is doc-only; a future editor can ignore the sentinel or violate the "render-only" marker. The deferred lint-rule follow-up (plan.md §F.6) is the recommended mechanical catch, explicitly out of scope for this SPEC.
- D5a label refactor (qualified concern-names) is additive but introduces a new vocabulary; if a future editor re-introduces a bare "Pre-emit self-check" label, the collision recurs. The sentinel in REQ-SHA-006 *(renumbered from iter-1 REQ-SHA-008)* partially mitigates this by verifying the qualified-name invariant on the other surface.
