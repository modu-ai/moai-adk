# Acceptance Criteria — SPEC-STEERING-ALIGN-OUTPUT-STYLE-SLIM-001

> GEARS-format acceptance criteria with re-runnable verification commands.
> Every AC names: the claim (GEARS), the command, and the PASS predicate.
> SSOT for AC enumeration is this file (NOT progress.md).
> All paths relative to project root. The output-style file path is long; the shell vars `LIVE` / `TMPL` below abbreviate the two trees.

```bash
LIVE=.claude/output-styles/moai/moai.md
TMPL=internal/template/templates/.claude/output-styles/moai/moai.md
```

---

## D. Acceptance Criteria Matrix

### AC-OSS-001 — Per-tree line-count drops into the derived range (SOFT band)

**GEARS (State-driven)**: **While** the diet is in effect, the always-loaded `moai.md` line count SHALL drop from 782 into the derived target range [530, 630] (plan.md §C.4) in BOTH trees, by the SAME amount — OR land above 630 ONLY when ALL preservation invariants (AC-OSS-003 banner skeletons + AC-OSS-004 translation tables + AC-OSS-008 directives) hold AND the run-phase records, in progress.md §E.2, an explicit "preservation-forced fewer cuts" rationale naming the specific render-SSOT passages that were KEPT rather than cut (REQ-OSS-008). The escape is NOT a blanket "any reduction below 782 passes" — it is gated on the preservation invariants being the documented cause of landing above 630.

**Given** the baseline 782/782 lines, **When** the diet is applied to both trees, **Then** each tree's `wc -l` lands in [530, 630] AND the two counts are equal — OR (behavioral-PASS) lands above 630 with both trees equal AND below 782 AND AC-OSS-003/004/008 all PASS AND a preservation-forced rationale recorded in progress.md §E.2.

```bash
L=$(wc -l < "$LIVE"); T=$(wc -l < "$TMPL")
echo "LIVE=$L TEMPLATE=$T (baseline 782 each; soft band [530,630])"
# Primary PASS: 530 <= L <= 630  AND  L == T
if [ "$L" -ge 530 ] && [ "$L" -le 630 ] && [ "$L" -eq "$T" ]; then
  echo "AC-OSS-001 PASS"
# Behavioral-PASS escape (REQ-OSS-008): L in (630, 782) AND L == T AND the preservation
# invariants are the documented cause. The line-count test alone CANNOT grant behavioral-PASS;
# it is conditional on AC-OSS-003 (14 banners) + AC-OSS-004 (8 locale tables) + AC-OSS-008
# (§9/§10 directives) ALL passing AND a progress.md §E.2 rationale naming the KEPT render-SSOT
# passages. A reduction that lands >630 WITHOUT those invariants passing is FAIL, not BEHAVIORAL-PASS.
elif [ "$L" -gt 630 ] && [ "$L" -lt 782 ] && [ "$L" -eq "$T" ]; then
  echo "AC-OSS-001 BEHAVIORAL-PASS CANDIDATE — REQUIRES: AC-OSS-003 PASS + AC-OSS-004 PASS + AC-OSS-008 PASS + progress.md §E.2 preservation-forced rationale naming KEPT render-SSOT passages. Verify those before accepting."
else
  echo "AC-OSS-001 FAIL"
fi
```

PASS predicate: both in [530,630] AND equal — OR (behavioral-PASS) both equal AND in (630,782) AND **AC-OSS-003 + AC-OSS-004 + AC-OSS-008 all PASS** AND a progress.md §E.2 rationale names the specific render-SSOT passages preservation forced to KEEP. The behavioral-PASS is gated on the preservation invariants being the documented cause — NOT a mere `L < 782`. Over-cutting render-SSOT to force into the band is forbidden (AP-OSS-003); under-cutting to dodge a hard cut is caught by the requirement that the rationale name actual KEPT render-SSOT passages.

---

### AC-OSS-002 — Template ↔ live byte-parity preserved

**GEARS (Ubiquitous)**: The two `moai.md` trees SHALL remain byte-identical after the diet.

**Given** byte-identical baseline (diff exit 0), **When** the diet is applied to both, **Then** `diff` still exits 0.

```bash
diff "$LIVE" "$TMPL"; echo "DIFF_EXIT=$?"
# PASS: DIFF_EXIT=0
```

PASS predicate: `DIFF_EXIT=0` (no output).

---

### AC-OSS-003 — All 14 banner skeletons survive (render-SSOT)

**GEARS (Unwanted behavior)**: The diet SHALL NOT drop any of the 14 banner skeletons. Each banner's `### ` sub-heading MUST remain present and renderable.

**Given** the 14 banner sub-headings at baseline, **When** the diet is applied, **Then** all 14 are still present.

```bash
for b in "Task Start" "Delegation Dispatch" "Checkpoint Gate" "Insight" \
         "Verification Matrix" "Plan Audit" "Discovery Report" "Race Absorbed" \
         "Epic Stats" "Epic Status" "Completion Report" "Error Recovery" \
         "Progress Board" "Session Handoff"; do
  grep -qE "^### .*${b}" "$LIVE" && echo "OK  $b" || echo "MISSING  $b"
done
N=$(for b in "Task Start" "Delegation Dispatch" "Checkpoint Gate" "Insight" "Verification Matrix" "Plan Audit" "Discovery Report" "Race Absorbed" "Epic Stats" "Epic Status" "Completion Report" "Error Recovery" "Progress Board" "Session Handoff"; do grep -qE "^### .*${b}" "$LIVE" && echo 1; done | wc -l | tr -d ' ')
echo "banner_skeletons=$N (expect 14)"
# PASS: N == 14
[ "$N" -eq 14 ] && echo "AC-OSS-003 PASS" || echo "AC-OSS-003 FAIL"
```

PASS predicate: all 14 banner sub-headings present (`N == 14`). **MUST / Blocking** — banner skeletons are render-SSOT (the over-cut defense for the banner layer).

---

### AC-OSS-004 — All en/ko/ja/zh translation tables + ko-canonical mapping tables intact (render-SSOT)

**GEARS (Unwanted behavior)**: The diet SHALL NOT drop any of the 8 per-banner en/ko/ja/zh header-translation tables, the §8 Session Handoff Cut-line Marker translation table, or the §8 Localization Contract ko-canonical mapping tables. These ARE the render SSOT (orchestrator reads at output time; no external owner).

**Given** the translation tables at baseline, **When** the diet is applied, **Then** the 4-locale column-header table count is unchanged AND the ko-canonical mapping markers survive AND the parity sentinel survives.

```bash
# PRIMARY GATE — the robust render-SSOT table count: the 4-locale column-header row
# `| English | Korean | Japanese | Chinese |` appears once per actual translation table.
# This counts the TABLES THEMSELVES (8 at baseline: 7 per-banner header tables + 1 cut-line marker
# table), NOT prose mentions of the phrase "translation table". A faithful condense preserves all 8.
COLHDR=$(grep -c '| English | Korean | Japanese | Chinese |' "$LIVE")
echo "locale_column_header_tables=$COLHDR (baseline 8 — the verified render-SSOT table count)"
# ko-canonical mapping markers (label→ko + banner-body-prose→ko catalogues)
grep -q 'ko canonical' "$LIVE" && echo "ko-canonical-mapping OK" || echo "ko-canonical-mapping MISSING"
# Session Handoff parity sentinel survives (REQ-OSS-006)
grep -q 'Drift-mitigation self-check sentinel' "$LIVE" && echo "parity-sentinel OK" || echo "parity-sentinel MISSING"
# PASS: COLHDR == 8 AND ko-canonical present AND parity sentinel present
[ "$COLHDR" -eq 8 ] && grep -q 'ko canonical' "$LIVE" && grep -q 'Drift-mitigation self-check sentinel' "$LIVE" \
  && echo "AC-OSS-004 PASS" || echo "AC-OSS-004 FAIL"

# INFORMATIONAL ONLY (NOT a gate) — do NOT gate on this count.
# `grep -c 'Header translation table\|Cut-line Marker translation table'` counts 12 at baseline,
# but 4 of those are PROSE mentions (L278 Fallback-rule prose, L655 6-block-format prose, L704
# pre-emit-self-check prose, plus the L679 cut-line-table intro line). L278/L655/L704 sit INSIDE
# the planned-cut/condense regions, so a faithful diet DECREMENTS this count — gating on it would
# FALSE-FAIL a correct implementation. The PRIMARY gate above (column-header count == 8) is the
# only blocking predicate; this 12-count is recorded for context, not enforcement.
TT=$(grep -c 'Header translation table\|Cut-line Marker translation table' "$LIVE")
echo "[informational] prose-inclusive 'translation table' mentions=$TT (baseline 12; NOT a gate — may decrement on faithful condense)"
```

PASS predicate: the `| English | Korean | Japanese | Chinese |` 4-locale column-header table count is **exactly 8** (the verified render-SSOT table count — the actual tables, not prose mentions), the `ko canonical` mapping markers present, AND the parity sentinel present. The prose-inclusive phrase-mention count (12 at baseline) is INFORMATIONAL ONLY and is NOT gated — 3 of its 4 extra matches sit inside planned-cut regions, so a faithful condense legitimately decrements it. **MUST / Blocking** — translation tables + ko-canonical mappings + parity sentinel are render-SSOT + parity-contract content (REQ-OSS-003 / REQ-OSS-006).

---

### AC-OSS-005 — Always-on byte-sum reduced per tree (attributed to real mechanisms)

**GEARS (State-driven)**: **While** the diet is in effect, each tree's byte count SHALL drop below the 55306 baseline, the reduction attributable solely to M-DELETE + M-POINTER (NOT a banner restructure, NOT M-SCOPE).

**Given** 55306 bytes/tree at baseline, **When** the diet is applied, **Then** `wc -c` is < 55306 in both trees AND the two are equal.

```bash
BL=$(wc -c < "$LIVE"); BT=$(wc -c < "$TMPL")
echo "bytes live=$BL template=$BT (baseline 55306 each)"
# PASS: BL < 55306 AND BL == BT
[ "$BL" -lt 55306 ] && [ "$BL" -eq "$BT" ] && echo "AC-OSS-005 PASS" || echo "AC-OSS-005 FAIL"
```

PASS predicate: both < 55306 AND equal. The run-phase report MUST attribute the delta to M-DELETE/M-POINTER (predominantly the §8 Session Handoff condense) — NOT a banner restructure (AP-OSS-005).

---

### AC-OSS-006 — POINTER edits gated on prose-duplication re-verification (over-cut defense — P2 D1 pattern)

**GEARS (Event-driven)**: **When** the run-phase is about to apply a POINTER edit to a `moai.md` passage, the run-phase SHALL FIRST re-grep the SSOT rule file for the passage's DISTINCTIVE content; **When** the grep returns 0 hits of the distinctive content, the run-phase SHALL block the POINTER edit and reclassify the passage as KEEP.

**Given** the surviving POINTER set (§8 Session Handoff narration, Epic-taxonomy explanation per plan.md §C.2/§C.3), **When** run-phase re-verifies each before editing, **Then** each shows ≥1 hit of its distinctive content; any 0-hit passage is KEPT, not pointer-ized.

```bash
echo "--- surviving POINTER candidates (expect >=1 distinctive-content hit each) ---"
grep -c '6-block\|Block 1\|Block 6\|Field-by-Field' .claude/rules/moai/workflow/session-handoff.md            # §8 Session Handoff field-by-field narration
grep -c 'source_session_id\|environment-fallback\|moai session current' .claude/rules/moai/workflow/session-handoff.md  # §8 source_session_id narration
grep -c 'effort ultracode\|workflow fan-out' .claude/rules/moai/workflow/session-handoff.md                   # §8 effort-ultracode narration
grep -c 'Epic\|Milestone\|multi-SPEC\|within-SPEC' .claude/rules/moai/development/sprint-round-naming.md       # Epic Stats/Status taxonomy explanation
echo "--- render-SSOT (must stay KEEP — these have NO external owner) ---"
# the banner skeletons + translation tables are render-SSOT: there is intentionally no external SSOT to grep.
# A 0-hit external grep for a render-SSOT passage CONFIRMS it must be KEPT (not pointer-ized).
# PASS: every surviving-POINTER grep >= 1; render-SSOT passages are KEPT by classification (not re-grepped against an external owner)
```

PASS predicate: every surviving POINTER candidate shows ≥1 distinctive-content hit in its named SSOT before its edit. A surviving-POINTER candidate that drops to 0 at run-time (SSOT changed between plan and run) MUST be reclassified KEEP — the run-phase reports this as a plan-vs-run drift, NOT a silent deletion. **MUST / Blocking** — this is the core defense against the over-cut failure mode (P2 D1 / AC-CMD-009 applied to the output-style file).

---

### AC-OSS-007 — Neutrality + output-styles count/parity CI guards still pass

**GEARS (Unwanted behavior)**: The diet SHALL NOT inject any NEW moai-adk internal artifact into the `moai.md` body, AND SHALL NOT break the output-styles count/parity invariants, as judged by the authoritative CI guards.

**Given** the CI guards passing at baseline, **When** the diet is applied, **Then** `TestTemplateNeutralityAudit`, `TestOutputStylesExactlyTwo`, and `TestOutputStylesTemplateLiveParity` STILL pass.

```bash
go test ./internal/template/ -run 'TestTemplateNeutralityAudit|TestOutputStylesExactlyTwo|TestOutputStylesTemplateLiveParity' 2>&1 | tail -5
# PASS: ok (all three guards green on the post-diet template tree)

# Optional informational delta (NOT the gate): only NEW forbidden tokens vs the pre-diet baseline matter.
git diff --no-index <(git show HEAD:internal/template/templates/.claude/output-styles/moai/moai.md) "$TMPL" \
  | grep '^+' | grep -nE 'SPEC-[A-Z]+-[0-9]{3}|feedback_|/Users/|Audit [0-9]+ Finding|CLAUDE\.local' \
  || echo "no NEW internal-artifact token added by the diet"
```

PASS predicate: `go test … -run 'TestTemplateNeutralityAudit|TestOutputStylesExactlyTwo|TestOutputStylesTemplateLiveParity'` exits 0 (the authoritative gate). The informational diff-grep is a defense-in-depth check that the diet ADDED no new forbidden token — pre-existing neutral-baseline content is NOT a regression and does NOT fail this AC. **MUST / Blocking** — neutrality + output-styles parity are CI-enforced invariants (REQ-OSS-005 / REQ-OSS-007).

---

### AC-OSS-008 — §9/§10 directives + verbatim-preserve symbol list + ultrathink token survive

**GEARS (Unwanted behavior)**: The diet SHALL NOT drop the §9 Language Rules / §10 Output Rules binding directives, the verbatim-preserve emoji/box-drawing/symbol lists, or the `ultrathink.` keyword token.

**Given** the §9/§10 directives + symbol list + token at baseline, **When** the diet is applied, **Then** all survive.

```bash
grep -qE '^## 9\. Language Rules' "$LIVE" && echo "§9 OK" || echo "§9 MISSING"
grep -qE '^## 10\. Output Rules' "$LIVE" && echo "§10 OK" || echo "§10 MISSING"
grep -q 'ultrathink\.' "$LIVE" && echo "ultrathink-token OK" || echo "ultrathink-token MISSING"
# verbatim-preserve symbol list (the emoji/box-drawing preserve directive)
grep -q 'Preserve verbatim' "$LIVE" && echo "verbatim-preserve-list OK" || echo "verbatim-preserve-list MISSING"
# free-form-prohibition directive (§10 HARD)
grep -q 'free-form interrogative prose' "$LIVE" && echo "free-form-prohibition OK" || echo "free-form-prohibition MISSING"
# PASS: §9 present AND §10 present AND ultrathink token present AND verbatim-preserve list present AND free-form prohibition present
```

PASS predicate: §9 + §10 headings present, `ultrathink.` token present, the verbatim-preserve symbol-list directive present, and the §10 free-form-prohibition directive present. **MUST / Blocking** — these are binding behavioral directives (REQ-OSS-003 / REQ-OSS-008).

---

### AC-OSS-009 — Every deleted-prose line has a verified external-SSOT home

**GEARS (Ubiquitous)**: For every passage the diet DELETES or POINTER-izes, the run-phase report SHALL name the external SSOT that owns the removed content AND show the grep evidence (≥1 distinctive-content hit).

**Given** the set of CUT + POINTER passages, **When** the diet is reported, **Then** the run-phase completion report contains a table mapping each removed passage → its SSOT owner → the grep command + observed hit count.

```bash
# This AC is verified by inspecting the run-phase completion report (progress.md §E.2 mechanism-attribution table).
# The table MUST have one row per CUT/POINTER passage with: passage, SSOT owner, grep command, observed hits >=1.
# Cross-check: every SSOT named in the report must be a real file.
for f in .claude/rules/moai/workflow/session-handoff.md \
         .claude/rules/moai/core/askuser-protocol.md \
         .claude/rules/moai/development/sprint-round-naming.md; do
  [ -f "$f" ] && echo "SSOT exists: $f" || echo "SSOT MISSING: $f"
done
# PASS: run report enumerates each removed passage → SSOT owner → grep evidence; every named SSOT file exists
```

PASS predicate: the run-phase report (progress.md §E.2) enumerates each removed passage with its SSOT owner + grep evidence, AND every named SSOT file exists on disk. **MUST / Blocking** — this is the verification-claim-integrity discipline (a deletion claim that "this is owned elsewhere" must be evidenced, not asserted).

---

## D.1 Severity / blocking classification

| AC | Severity | Blocking? |
|----|----------|-----------|
| AC-OSS-001 | MUST | Blocking (the diet's primary measurable outcome; SOFT band with behavioral-PASS escape) |
| AC-OSS-002 | MUST | Blocking (byte-parity is the Template-First invariant) |
| AC-OSS-003 | MUST | Blocking (all 14 banner skeletons are render-SSOT — banner-layer over-cut defense) |
| AC-OSS-004 | MUST | Blocking (translation tables + ko-canonical mappings + parity sentinel are render-SSOT + parity-contract) |
| AC-OSS-005 | MUST | Blocking (real byte reduction, real-mechanism attribution) |
| AC-OSS-006 | MUST | Blocking (POINTER edits gated on prose-duplication re-verification — core over-cut defense, P2 D1 pattern) |
| AC-OSS-007 | MUST | Blocking (neutrality + output-styles count/parity CI guards) |
| AC-OSS-008 | MUST | Blocking (§9/§10 directives + verbatim-preserve list + ultrathink token survive) |
| AC-OSS-009 | MUST | Blocking (every deleted-prose line has a verified external-SSOT home — verification-claim-integrity) |

All 9 ACs are MUST-blocking. There is no SHOULD AC in this SPEC (unlike P2's AC-CMD-008 §4-nesting SHOULD, which had a residual-risk WebFetch dependency — this SPEC's pointer targets are all internal SSOTs with no upstream-doc uncertainty).

## D.2 Traceability (REQ → AC)

| REQ | AC | Theme |
|-----|----|-------|
| REQ-OSS-001 | AC-OSS-001, AC-OSS-005, AC-OSS-006 | §8 Session Handoff pointer-ization → drop + over-cut gate |
| REQ-OSS-002 | AC-OSS-005, AC-OSS-006, AC-OSS-009 | rule-SSOT pointer-ization → drop + gate + SSOT-home evidence |
| REQ-OSS-003 | AC-OSS-003, AC-OSS-004, AC-OSS-008 | render-SSOT preservation (banners + tables + directives) |
| REQ-OSS-005 | AC-OSS-002, AC-OSS-007 | template-mirror parity + output-styles count guard |
| REQ-OSS-006 | AC-OSS-004 | parity-sentinel + 4-locale tables survive |
| REQ-OSS-007 | AC-OSS-007 | neutrality / isolation |
| REQ-OSS-008 | AC-OSS-001 | derived target range (not arbitrary number; behavioral-PASS escape) |
| REQ-OSS-009 | AC-OSS-006 | POINTER edits gated on prose-duplication re-verification |

## D.3 Definition of Done

- [ ] All 9 MUST-blocking ACs PASS (AC-OSS-001..009).
- [ ] AC-OSS-006: every surviving-POINTER passage re-verified to carry distinctive SSOT prose before its edit; any 0-hit passage reclassified KEEP (reported as plan-vs-run drift, not silent deletion).
- [ ] AC-OSS-003: all 14 banner skeletons present post-diet.
- [ ] AC-OSS-004: all 8 per-banner translation tables + Cut-line Marker table + ko-canonical mapping tables + parity sentinel survive; 4-locale (en/ko/ja/zh) columns intact.
- [ ] `go test ./internal/template/...` green (`TestTemplateNeutralityAudit` + `TestOutputStylesExactlyTwo` + `TestOutputStylesTemplateLiveParity`).
- [ ] `diff $LIVE $TMPL` → exit 0 (byte-parity).
- [ ] spec-lint clean on this SPEC dir (`MissingExclusions` 0 — §D `### Out of Scope —` h3 sub-headings present, 4 of them).
- [ ] Both trees committed + pushed (Conventional Commits, no `--no-verify`).
- [ ] §A.5 PRESERVE-list paths untouched (`git status` shows only the 2 moai.md trees + this SPEC's spec.md(status) + progress.md). SPEC-DIVECC-INVENTORY-VIEW-001 NOT touched.
- [ ] Run-phase completion report (progress.md §E.2) attributes the byte reduction to M-DELETE/M-POINTER (predominantly §8 Session Handoff condense; excludes banner restructure/M-SCOPE) per AC-OSS-005, AND enumerates each removed passage → SSOT owner → grep evidence per AC-OSS-009.

## D.4 Indirect verification note

The "token reduction" is verified INDIRECTLY via line/byte-sum (AC-OSS-001/005), not a real tokenizer — acceptable per the per-line-test framing (spec.md §F.2). If an exact token count is later required, run-phase MAY add a tokenizer measurement, but it is not a blocking AC here.

## D.5 Forward-looking checks (NOT this SPEC's ACs — recorded for the future)

- The §4 Forced Delegation Table archived-agent names (`expert-backend`, `manager-quality`, etc., moai.md L113-127) are stale per the agent-catalog consolidation. This is a NEUTRALITY/ARCHIVED-AGENT cleanup, NOT a diet, and is OUT OF SCOPE here (spec.md §D / plan.md §D.5). Recorded as a forward candidate for a future SPEC. This SPEC's ACs do NOT assert anything about §4 content.
