# acceptance.md — SPEC-V3R6-SESSION-HANDOFF-SSOT-ALIGN-001

> Mechanically verifiable ACs. Every AC names the exact command + expected output. No prose-only ACs.
>
> **iter-2 revision**: iter-1 had 9 ACs (AC-SHA-001..009); iter-2 DELETES AC-SHA-004 (D5b non-defect — REQ-SHA-004 deleted per D1) and AC-SHA-006 (D4 hypothetical non-conflation — REQ-SHA-006 dropped per D4, doctrine already clean). Surviving ACs renumbered AC-SHA-001..007 with no gaps. Additionally: AC-SHA-003 second regex fixed (D5 — was logically inverted); the former AC-SHA-005 (now AC-SHA-004) re-anchored to the Localization Table section with a baseline-absent phrase (D3 — was vacuously satisfied by unrelated "fallback" mentions); all line-number-window greps converted to content-anchored greps (D7).

## §D. AC Matrix

| AC ID | REQ | Severity | Verification Class |
|-------|-----|----------|--------------------|
| AC-SHA-001 | REQ-SHA-001 | MUST | grep (content) |
| AC-SHA-002 | REQ-SHA-002 | SHOULD | grep (content) |
| AC-SHA-003 | REQ-SHA-003 | MUST | grep (qualified concern-name labels; inverted-regex fixed) |
| AC-SHA-004 | REQ-SHA-004 | MUST | grep (Localization-Table-scoped, baseline-absent phrase) |
| AC-SHA-005 | REQ-SHA-005 | MUST | grep (index present + forward links) |
| AC-SHA-006 | REQ-SHA-006 | MUST | grep (SSOT-pointer + sentinel) |
| AC-SHA-007 | REQ-SHA-007 | MUST | byte-parity (cmp both files both trees) |

## §D.1 AC Detail

### AC-SHA-001 (REQ-SHA-001, MUST) — D3 `wave` filename token migrated

**Given** session-handoff.md (both trees) contains the filename pattern line,
**When** the verifier runs:
```bash
grep -nE 'project_<wave>|project_wave6|project_myproj_prev_wave' \
  .claude/rules/moai/workflow/session-handoff.md \
  internal/template/templates/.claude/rules/moai/workflow/session-handoff.md
```
**Then** the command returns **0 matches** (the retired tokens are gone), AND
```bash
grep -nE 'project_<sprint>|project_<round>|project_sprint6|project_round' \
  .claude/rules/moai/workflow/session-handoff.md
```
returns **≥1 match** (the new vocabulary is present).

### AC-SHA-002 (REQ-SHA-002, SHOULD) — D3 generic-English-noun `wave` preserved in prose

**Given** session-handoff.md prose describes multi-SPEC scheduling,
**When** the verifier runs:
```bash
grep -nE 'multi-SPEC waves|current wave' .claude/rules/moai/workflow/session-handoff.md
```
**Then** the command returns **≥1 match** (the generic prose noun is preserved, not mechanically rewritten). *(Line numbers are intentionally NOT cited in this AC — the content pattern is the gate, per the line-number-drift-asymmetry lesson.)*

### AC-SHA-003 (REQ-SHA-003, MUST) — D5a "Pre-emit self-check" labels disambiguated to concern-names

> *iter-2 D5 + D6 correction.* The iter-1 AC had two defects: (a) its second regex was logically inverted (it matched the qualified forms it was supposed to exclude — D5), and (b) it did not catch the D6 ambiguity where a count-only qualifier `(8 items)` is not a concern-name. The iter-2 AC uses a two-grep structure, both of which FAIL at baseline and PASS only after the fix.

**Given** the bare label "Pre-emit self-check" collides across 3 sites (session-handoff.md L254 header; moai.md §8 localization-render label; moai.md §8 template-completeness label) with 3 distinct concerns,

**When** the verifier runs **grep A** — the **concern-name presence** grep (verify the 3 canonical concern-names are present):
```bash
grep -nE 'Pre-emit self-check \((paste-ready budget|localization render|session-handoff template completeness)\)' \
  .claude/rules/moai/workflow/session-handoff.md \
  .claude/output-styles/moai/moai.md
```
**Then** the command returns **≥3 matches** (one per collision site, each carrying a distinct canonical concern-name), AND

**When** the verifier runs **grep B** — the **count-only qualifier absence** grep (verify no count-only qualifier remains):
```bash
grep -nE 'Pre-emit self-check \([0-9]+ items\)' \
  .claude/rules/moai/workflow/session-handoff.md \
  .claude/output-styles/moai/moai.md
```
**Then** the command returns **0 matches** (no count-only qualifier survives; the D6 ambiguity is closed).

> **Baseline behavior (FAIL-at-baseline verification — no vacuous-pass)**:
> - **grep A at baseline**: returns **0 matches** (none of the 3 canonical concern-names appear yet). The AC's ≥3-match threshold FAILS at baseline. ✓
> - **grep B at baseline**: returns **1 match** (`session-handoff.md:254:### Pre-emit self-check (8 items)`). The AC's 0-match threshold FAILS at baseline. ✓
> - Both greps flip to PASS only after the implementer replaces the count-only `(8 items)` and the action-flavored qualifiers with the 3 canonical concern-names at all 3 collision sites. The L722 prose reference (`- [HARD] Pre-emit self-check: every banner/...`) is NOT a collision site (it is inline prose, not a header/list-label) and is intentionally out of scope for this AC; the implementer verifies manually at §E that L722 is not regressed.

### AC-SHA-004 (REQ-SHA-004, MUST) — D6 explicit localization fallback rule (Localization-Table-scoped, baseline-absent phrase)

> *Renumbered from iter-1 AC-SHA-005. iter-2 D3 correction: the iter-1 grep was vacuously satisfied at baseline by 3 unrelated "fallback" mentions (L76 source_session_id, L132 anti-pattern, L166 launcher). The iter-2 AC anchors to the Localization Table section with a phrase that does NOT appear anywhere in the baseline file, so it FAILS at baseline and PASSES only after the fix.*

**Given** session-handoff.md Localization Table covers only en/ko/ja/zh and has NO explicit fallback rule at baseline,
**When** the verifier runs the baseline-absent-phrase grep (scoped to the Localization Table region by anchoring on the section header and the next 40 lines):
```bash
# Anchor by content (the "Localization Table" section header), NOT by line number.
# Require a phrase that does NOT appear anywhere in the baseline file.
awk '/^###? Localization Table/{flag=1} flag&&/^##[^#]/{exit} flag' \
  .claude/rules/moai/workflow/session-handoff.md \
  | grep -E 'English.*canonical fallback skeleton|locales not.*table.*English column|ISO-639.*not in.*table.*fallback'
```
**Then** the command returns **≥1 match** in the Localization Table section.

> **Baseline behavior (FAIL-at-baseline check)**: at baseline, the awk-extracted Localization Table region contains NONE of the required phrases (`English.*canonical fallback skeleton` / `locales not.*table.*English column` / `ISO-639.*not in.*table.*fallback`). The implementer confirms this at the §E self-verification step by running the same command pre-fix and observing 0 matches, then post-fix and observing ≥1 match. This is the D3 re-anchoring: no vacuous-pass at baseline.

### AC-SHA-005 (REQ-SHA-005, MUST) — D9 consolidated anti-pattern index

> *Renumbered from iter-1 AC-SHA-007. No content change; only the AC ID renumbered.*

**Given** session-handoff.md has 3 separated anti-pattern catalogues,
**When** the verifier runs:
```bash
grep -nE 'Anti-Pattern Index|consolidated anti-pattern index' .claude/rules/moai/workflow/session-handoff.md
```
**Then** the command returns **≥1 match** identifying the canonical index location, AND the index contains forward links to every AP code:
```bash
grep -cE 'AP-D-00[1-5]|AP-V-00[1-4]' .claude/rules/moai/workflow/session-handoff.md
```
returns **≥9** (5 AP-D + 4 AP-V codes referenced; the index is the navigational single-source even if the detail sections carry the prose).

### AC-SHA-006 (REQ-SHA-006, MUST) — recurrence-mitigation mechanism (hybrid)

> *Renumbered from iter-1 AC-SHA-008. No content change beyond the AC ID renumbering and the mechanism language downgrade ("prevention" → "mitigation").*

**Given** moai.md §8 Session Handoff block previously duplicated canonical content,
**When** the verifier runs:
```bash
grep -nE 'render-only, not canonical|canonical lives in.*session-handoff\.md' .claude/output-styles/moai/moai.md
```
**Then** the command returns **≥1 match** identifying the SSOT-pointer marker, AND
```bash
grep -nE 'self-check sentinel|parity check.*session-handoff|verify.*other surface' \
  .claude/rules/moai/workflow/session-handoff.md .claude/output-styles/moai/moai.md
```
returns **≥1 match in EACH file** (both surfaces carry a sentinel referencing the other surface's structural invariant).

### AC-SHA-007 (REQ-SHA-007, MUST) — byte-parity for both files both trees

> *Renumbered from iter-1 AC-SHA-009. No content change beyond the AC ID renumbering.*

**Given** both files were modified in both trees,
**When** the verifier runs:
```bash
cmp .claude/rules/moai/workflow/session-handoff.md \
    internal/template/templates/.claude/rules/moai/workflow/session-handoff.md
echo "session-handoff exit: $?"
cmp .claude/output-styles/moai/moai.md \
    internal/template/templates/.claude/output-styles/moai/moai.md
echo "moai exit: $?"
```
**Then** both `cmp` commands exit **0** (byte-identical). Non-zero exit fails the AC.

## §D.2 Severity Classification

- **MUST** (AC-SHA-001, 003, 004, 005, 006, 007): failure blocks merge. These are the load-bearing drift closures + the recurrence-mitigation mechanism + byte-parity.
- **SHOULD** (AC-SHA-002): failure is a debt flag, non-blocking. The generic-English-noun preservation is a judgment call; if the implementer decides to harmonize the prose noun too, that is acceptable as long as REQ-SHA-001 is met.

## §D.3 Edge Cases

- **Line-number drift asymmetry**: the LOCAL tree may be N lines longer than TEMPLATE (trapped local-only sections). Every AC grep in this iter-2 revision anchors by **content pattern**; no AC uses a hard line-number window (`sed -n 'NNN,MMMp'` or `awk 'NR>NNN && NR<MMM'`) as its primary gate. This is the direct correction of the mechanism by which the iter-1 D5b false-defect entered (a `sed -n '683,696p'` window excluded the disproof line at L682). Line numbers may appear in comments as secondary attribution; the grep content pattern is always the primary gate.
- **Sentinel deletion by future editor**: the self-check sentinel is manual. AC-SHA-006 verifies presence at merge time; it cannot prevent future deletion. This is the honest residual risk (plan.md §F.5).
- **"Render-only" marker convention violation**: a future editor copying canonical content into a render-only block violates the marker convention but does not fail any current AC. The deferred lint-rule follow-up (plan.md §F.6) is the mechanical catch.
- ***(iter-2 NEW)* Vacuous-pass detection**: every MUST AC in this iter-2 revision is verified to FAIL at baseline (pre-fix) and PASS only after the fix. The implementer runs each AC grep against the pre-fix baseline at §E self-verification and records the baseline output (must be 0 matches or otherwise failing); then runs it post-fix (must be the declared PASS output). This guards against the iter-1 D3/D4/D5 vacuous-pass class.

## §D.4 Quality Gate Criteria

- All MUST ACs pass (mechanically verified by grep/cmp).
- SHOULD AC passes OR is explicitly marked as deferred debt with rationale.
- §E self-verification (plan.md §E) produces verbatim command + output for E1–E7, including the baseline-vs-post-fix comparison for each MUST AC.
- Byte-parity holds for both files both trees (AC-SHA-007).

## §D.5 Definition of Done

- [ ] All 3 real drifts (D3/D5a/D6/D9) have ≥1 fix site in both trees. *(D5b was withdrawn — see research.md §A.5; no fix site exists for a non-defect.)*
- [ ] Recurrence-mitigation mechanism (hybrid: SSOT-pointer + sentinel) applied to both surfaces.
- [ ] Byte-parity holds for both files both trees.
- [ ] §E self-verification output reproduced in the run-phase evidence, including baseline-vs-post-fix comparison for every MUST AC (no vacuous-pass).
- [ ] No new Go code added (new-Go-code=0 constraint honored).
- [ ] `era: V3R6` explicit frontmatter override present in spec.md (verified at plan-phase creation).
- [ ] *(iter-2 NEW)* No AC uses a hard line-number window as its primary gate (all content-anchored).

## §D.6 Forward-Looking Checks (post-merge sustainability)

- The self-check sentinel in each file references the *other* file's structural invariant by **content** (concern-name or invariant description), not by line number, so it survives line-number drift.
- The "render-only, not canonical" marker convention is documented in plan.md §F.4 so a future editor has a referenceable spec for the convention.
- The deferred lint-rule follow-up (plan.md §F.6) is recorded as a recommended future code SPEC, so the residual-risk gap has an explicit closure path.

## §D.7 Indirect Verification (cannot be directly grepped)

- **Render-time usability preservation**: the hybrid mechanism (REQ-SHA-006) is designed to preserve the orchestrator's ability to read moai.md §8 at output time without an extra hop for the 6-block skeleton. This is a design property, not a directly-greppable AC. The plan-auditor evaluates it via plan.md §F.1–F.3 (the trade-off analysis).
- **Sentinel effectiveness against future drift**: the sentinel raises the visibility bar but is not a mechanical guarantee (plan.md §F.5 residual risk). The plan-auditor evaluates whether the residual-risk note is honest (per verification-claim-integrity §3.5) and whether the deferred lint-rule follow-up is properly scoped.
