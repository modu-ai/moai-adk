# Design — SPEC-HUMANIZE-001

Design decisions for the copy-genre layer. This is a documentation-authoring design (skill markdown structure), not a software architecture.

## §A Module structure — copy layer within one module file (NOT separate files)

**Decision**: Add the copy layer as an appended section inside each existing language module file (`modules/korean.md`, `modules/english.md`, `modules/japanese.md`, `modules/chinese.md`) — NOT as new separate files (e.g. `korean-copy.md`).

**Rationale**:
- The SKILL.md **language routing table** routes to exactly ONE module per language (`Korean → modules/korean.md`, etc.). A separate copy file would require a second routing dimension (language × genre), complicating the routing structure that REQ-HUM (scope item 5) says to keep.
- The port source keeps prose + copy (A–J + L + M) in one `ai-tell-taxonomy.md` file — the same single-file-per-language shape.
- One file per language keeps the byte-identity mirror check (REQ-HUM-008) simple: 5 files in, 5 files out.
- The copy layer is genre-scoped internally via a "Copy Layer" heading + a copy-mode note, so a detector/agent can still restrict copy rules to copy contexts without a separate file.

**Structure per module** (appended after the existing prose "Rewrite Examples" / grading, before "Source & License"):
```
## Copy Layer (<language>)
<genre scope note: headlines / CTA / landing / brand story / slides>
### Detection Categories (copy)   — table: ID | tell | why AI | severity
### Severity Rationale (copy)
### Before/After Rewrite Examples (copy)   — examples in target language only
### <boundary analysis, where applicable — JA 体言止め, ZH 对偶/排比>
### High-False-Positive Signals (copy)
```

The trailing per-module "Source & License" block stays LAST (REQ-HUM-011).

## §B ID prefix scheme — per-module native continuation, with English the deliberate exception

Each module's copy layer continues that module's **existing native ID scheme** — except English, which gets a deliberately distinct `ENC-` prefix.

| Module | Prose scheme | Copy scheme | Rationale |
|--------|--------------|-------------|-----------|
| korean.md | A–J letters + numbered subcats (A-1…A-19) | **A-20…A-25** (extends A), **L-1…L-8** (new cat L), **M-1…M-3** (new cat M) | Faithful port — IDs match the source taxonomy exactly. A-20…A-25 are copy-translationese subcategories of the existing translationese category A; L and M are genuinely new categories in the source. |
| english.md | EN-A…EN-J | **ENC-1…ENC-9** (new `ENC-` prefix) | Deliberate exception. `ENC-` (English Copy) is chosen over continuing `EN-K…` because (1) it mirrors the Korean genre separation, (2) several ENC tells are *copy instances of a prose category* (ENC-2 = EN-B on a headline; ENC-1 = EN-A vocabulary at a headline) — a distinct prefix makes the "same rhetorical move, different genre surface" relationship explicit instead of pretending it is a brand-new lexical class, (3) it lets a detector scope copy rules to copy contexts and avoid firing them on body prose. Parent prose categories are named per ENC row. |
| japanese.md | JA-01…JA-09 (numeric) | **JA-10…JA-14** (continues numeric) | The Japanese prose module has no letter categories, only a numeric sequence; continuing it (JA-10…JA-14) is the natural extension. |
| chinese.md | CN-A…CN-K (letters) | **CN-L…CN-Q** (continues letters) | The Chinese prose module uses a letter sequence; continuing it (CN-L…CN-Q) is the natural extension. |

**Why the asymmetry is correct, not sloppy**: the copy layer continues each module's *own* native scheme; the schemes differ because the four prose modules were authored with different ID conventions historically. English is the only module where a distinct copy prefix was deliberately introduced, because the research argued the parent-category relationship is strong enough to warrant making it explicit. Documenting the asymmetry here prevents a future reviewer from "normalizing" it into a single scheme (which would erase the intentional English genre-separation signal).

## §C Copy-mode guard promotion into shared SKILL.md

**Decision**: Promote the copy-mode over-editing guard and the dual grading tables from a per-module concern to the **shared, all-language SKILL.md section**, because the guard and grading logic are language-agnostic (they operate on fact anchors and change-rate, not on language-specific tells).

**What moves into the shared section**:
1. **Prose-mode vs copy-mode distinction** (new subsection under the existing "Over-Editing Guardrails"). Prose mode = the existing 30%/50% change-rate guard. Copy mode = the fact-anchor preservation guard.
2. **Copy-mode fact-anchor guard** (REQ-HUM-005): meaning invariance in copy mode = numbers, dates, prices, proper nouns, legal notation preserved 100% + the core promise/benefit preserved; expression & sentence structure MAY be rewritten. The change-rate guard is REPLACED (not merely relaxed) in copy mode, because legitimate copy humanization routinely rewrites >30% of a headline while preserving every fact anchor.
3. **Dual grading tables** (REQ-HUM-006):
   - Prose-mode grade: residual S1/S2 + change-rate band + improvement % (the existing table).
   - Copy-mode grade: residual S1 (incl. the S1 copy tells) + fact-anchor loss count + self-verification, with NO change-rate band.
4. **Mode selection note**: default from genre (column/report/blog/formal → prose; copy/headline/CTA/landing/slides → copy), overridable by explicit mode.

**What stays per-module**: the language-specific tell categories, severities, boundary analyses, and before/after examples. The shared section only gains the mode/guard/grading machinery.

## §D Retired-Python consequence (design impact + mitigation)

**Decision**: Do NOT port the source skill's quantitative Python layer (`metrics.py`, `metrics_v2.py`, `humanize_html.py`, `baseline*.json`, `tests/`). The skill remains LLM-heuristic-only.

**Consequence chain**:
1. The source computed character-level edit distance (Levenshtein ÷ source length) and per-metric before/after deltas in Python — reproducible, deterministic numbers.
2. Without it, the 30%/50% change-rate guard and the "improvement %" grade input become **LLM estimates**, not measurements. Two runs on the same text may estimate slightly different change rates.
3. This is acceptable because the skill's value is the *catalogue + rewrite guidance*, not a reproducible score; the Python layer was an optional analysis leg even in the source.

**Mitigation** (REQ-HUM-014): the shared SKILL.md instructs **conservative judgment near the thresholds** — when the estimated change rate is borderline (near 30% or near 50%), treat it as OVER the threshold (WARN at ~30%, HALT at ~50%). This biases toward caution, so the loss of a precise metric never lets an over-edit slip through. The loss of reproducible before/after percentages is recorded as a known limitation in `spec.md` §E.

## §E Non-transfer design encoding (how the three findings become module structure)

The three non-transfer findings (`research.md` §7) are encoded structurally, not merely noted:

1. **English (M-2 non-transfer)**: the English "Copy Layer" has NO ENC row for predicate-less fragments. The slide section (§2.2 in research) explicitly lists the fragment-headline as a HIGH-false-positive signal, not a removable category. The English "High-False-Positive Signals (copy)" subsection names "fragment/verbless headlines" as natural English register.
2. **Japanese (体言止め non-transfer)**: JA-10 is authored as a frequency-gated S2 with a dedicated "体言止め Boundary Analysis" subsection stating the presence-vs-over-reliance distinction and the ≥3-consecutive threshold. The Korean particle-ending tell is deliberately absent. The "High-False-Positive Signals" subsection names "a single strategic 体言止め" as legitimate craft.
3. **Chinese (对偶/排比 non-transfer)**: CN-F-adjacent copy handling is authored via a "对偶/排比 Boundary Analysis" subsection stating content-first-vs-template-first, with count demoted to a weak signal. CN-M is scoped to the binary-contrast headline shape; a "Dash-Contrast Applicability" note lists the 破折号 —— 解释说明 use as a false-positive guard.

The false-positive guard scenarios in `acceptance.md` §D.3 verify each of these encodings behaves correctly (a human slide fragment, one strategic 体言止め, one crafted 排比 are each NOT flagged).

## §F Neutrality-stripping design (§25 compliance)

The Korean port carries source-taxonomy prose dense with internal artifacts that MUST be stripped before shipping (REQ-HUM-009):
- Internal version annotations: `v2.2 신규`, `v2.6 확장`, `v1.1 신규` → removed (the shipped module carries no internal versioning; the SKILL.md `metadata.version` is the only version marker).
- `_source_anchor:` lines citing internal specs (`ai-tell-ko-copy-spec.md`), `.moai/reports/research-*` paths → removed; the academic provenance is summarized in this SPEC's `research.md` §6.4 (which lives in `.moai/specs/`, NOT distributed), not in the shipped module.
- Sample-count provenance ("14개 샘플 중 6회 재현") → removed from shipped content (kept in `research.md` §1).
- The shipped Korean module's instruction prose is English (REQ-HUM-010); only the before/after examples stay Korean.

The neutrality grep (AC-HUM-008, rewritten v0.2.0) enforces this with **six** zero-match categories (internal SPEC IDs, REQ/AC tokens, commit SHAs, version annotations, `.moai/reports/research` paths, and internal work-dates). The date class is **body-scoped**: SKILL.md is the only file with YAML frontmatter (its `created:`/`updated:` are legitimate skill metadata), so its frontmatter is stripped before the date grep; the four modules carry no frontmatter, so any ISO date there is a body-date leak. This keeps the check reachable to 0 on a normal-frontmatter tree, and aligns it with the real CI guard's narrow tier (which does NOT flag frontmatter dates — `S1-internal-date` is strict-tier only). The v0.3.0 re-authored `korean.md` is scanned like any other module (the maintainer's-taxonomy source is dense with internal annotations that must still be stripped).

## §G Attribution & Licensing Design — RESOLVED v0.3.0 (user decision at Implementation Kickoff gate)

**Decision (user, at the Kickoff gate)**: the Korean module is **re-authored as an original work** from the maintainer's own taxonomy, referencing im-not-ai only as inspiration. The MIT dependency is dissolved entirely: NO NOTICE.md, `license: Apache-2.0` unchanged (single license, matching the repo LICENSE and all template skills), and the attribution blocks become a license-free **courtesy credit** — "category-catalogue structure inspired by the im-not-ai (Humanize KR) project."

**Why the approach is sound**:
1. **The source taxonomy is the maintainer's own IP.** The reference source (`general-humanize-korean` at claude.mo.ai.kr) contains **zero** references to im-not-ai or epoko — verified by `grep -rn 'im-not-ai|epoko'` over the entire source skill directory (no matches). Re-authoring from it attaches no third-party license obligation.
2. **The MIT claim was an inaccurate self-description, not a content lineage.** The v1.0.0 `modules/korean.md` *claims* to be "a faithful port of im-not-ai" (line 133), but the copy-layer content this SPEC works from has no im-not-ai lineage. The claim is the problem; removing the claim (and re-authoring the module so the prose layer, too, derives from the maintainer's taxonomy) removes the encumbrance at its root.
3. **Inspiration and independently-authored structure do not attach MIT obligations.** A courtesy credit naming im-not-ai as structural inspiration carries no license claim, no copyright line, and no notice obligation.
4. **Single-license consistency.** `license: Apache-2.0` matches the repository LICENSE and all 18 template skills; no compound SPDX expression, no per-material split, no dangling pointers.

**Design consequences**:
- REQ-HUM-001 (v0.3.0): the run phase REWRITES `modules/korean.md` in full (prose A–J + copy layer) rather than appending; pattern IDs/severities may align with the maintainer's taxonomy freely.
- REQ-HUM-015 (v0.3.0): the 5 dangling `See NOTICE.md` pointers are removed; each attribution block becomes the courtesy credit. AC-HUM-016 greps the skill dir for `NOTICE.md` → 0 matches.
- REQ-HUM-016 (v0.3.0): `license:` field untouched; AC-HUM-017 greps for `MIT License` → 0 matches and asserts `license: Apache-2.0` verbatim.
- The v0.2.0 "Unresolved question" (aggregate-license intent; whether structural attribution rises to an MIT "substantial portion") is **RESOLVED by this decision** — it is moot once no ported content remains and the credit carries no license claim.

**Superseded v0.2.0 design (record only)**: the prior §G design (mixed-license aggregate: NOTICE.md with the full MIT permission-notice body + `license: "Apache-2.0 AND MIT"` compound + per-material split) was authored under the then-believed premise that the Korean module was genuinely MIT-ported. The premise was falsified by the source-lineage grep; that design is superseded, not merely deferred.

## §H Cross-References
- `spec.md` §C (non-transfer constraints), §D (REQ set incl. D.5 licensing REQ-HUM-015/016).
- `research.md` §1–§7 (evidence + the three findings).
- `acceptance.md` §D.2/§D.3 (8-sample + false-positive scenarios); §D.1 AC-HUM-016/017 (attribution cleanup + license unchanged).
- Reference source (maintainer's own taxonomy, zero im-not-ai lineage): `general-humanize-korean` `ai-tell-taxonomy.md` + `SKILL.md`. im-not-ai (https://github.com/epoko77-ai/im-not-ai) is credited as inspiration only.
- `CLAUDE.local.md` §25 + `internal/template/CLAUDE.md` (mirror parity, byte-embed).
