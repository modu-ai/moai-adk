# SPEC Review Report: SPEC-DESIGN-DOCS-V31-001
Iteration: 1/3
Verdict: FAIL
Overall Score: 0.76  (Tier L PASS threshold = 0.85)

> Adversarial plan-phase audit. M1 Context Isolation active — author reasoning ignored;
> only spec.md / plan.md / acceptance.md / research.md / progress.md evaluated.
> Handoff staging at `/tmp/moai-design-handoff/moai-adk/project/` and live
> `docs-site/static/moai-brand.css` were read directly to verify SPEC claims.

---

## Must-Pass Results

- **[FAIL] MP-1 REQ number consistency.** 29 REQs are uniquely numbered across 6
  groups (REQ-IA-001..004, REQ-NB-001..004, REQ-DS-001..007, REQ-KO-001..006,
  REQ-I18N-001..005, REQ-BL-001..003). No gaps, no duplicates *within* the
  REQ-<GROUP>-<NN> scheme. Sequential consistency holds. **BUT** the count itself
  is over-budget — see MP-2/Defect D2. Evidence: `grep -c '^\*\*REQ-' spec.md` → 29.

- **[PASS] MP-2 EARS/GEARS format compliance** (requirement layer). All 29 REQ
  entries match one of the five GEARS patterns. Spot check: REQ-IA-003 `When …
  SHALL` (Event-driven); REQ-IA-004 `While … SHALL` (State-driven); REQ-NB-004
  `Where … MAY` (capability gate); REQ-DS-003 `SHALL NOT` (Unwanted canonical
  negative); the majority Ubiquitous `The <subject> SHALL`. The two entries
  labeled `(Event-detected)` (REQ-KO-003, REQ-I18N-002) use the canonical
  `When … SHALL` structure but carry a non-standard label — see D7 (MINOR).
  Judgment binds the `REQ-XXX` requirement layer only; ACs are Given-When-Then
  by design (verification layer) and are not modality-checked here.

- **[PASS] MP-3 YAML frontmatter validity.** All 12 canonical fields present
  with correct types. `id: SPEC-DESIGN-DOCS-V31-001` (string, regex match);
  `version: "0.1.0"` (quoted semver); `status: draft` (enum-valid);
  `created: 2026-08-11`, `updated: 2026-08-11` (ISO dates); `priority: High`
  (enum-allowed alias of P1); `phase: "v3.1-rc.1 target"` (release target —
  NOT a prohibited lifecycle token); `module: docs-site`; `lifecycle:
  spec-anchored`; `tags:` comma-separated string; `tier: L` (optional field).
  No rejected snake_case aliases. Verified field-by-field against
  `.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12.

- **[N/A → PASS] MP-4 Section 22 language neutrality.** SPEC is scoped to a
  single-language docs-site (Hugo + Korean-content-first). No
  multi-language-tool enumeration obligation. REQ-BL-002 carries the
  template-neutrality constraint for any mirrored artifact. Auto-passes.

- **[PASS] MP-5 D7 cross-SPEC reconciliation.** D7 grep extracted 13 SPEC-ID
  references (SPEC-DESIGN-DOCSV2-001, SPEC-FACTORY-MODE-001,
  SPEC-INFINITE-GOAL-001, SPEC-PROJECT-NAVIGATOR-001/002/003,
  SPEC-HIERARCHICAL-TEAM-001, SPEC-AUDIT-MULTI-MODEL-001,
  SPEC-AUTONOMY-TIERS-001, SPEC-MODEL-PROFILE-MATRIX-001,
  SPEC-AGENT-MODEL-ENFORCE-001, SPEC-STOPCHAIN-TRIM-001,
  SPEC-AGENT-PARALLEL-OPT-001, SPEC-CC2219-UPSTREAM-ALIGN-001,
  SPEC-DOCSITE-ADVANCED-001). Catalog in §F.1 records each one's `status:` —
  11 `completed` + 1 `in-progress` (SPEC-AGENT-MODEL-ENFORCE-001, explicitly
  flagged as "partial"). No referenced SPEC is `retired`/`superseded`/
  `archived`. SPEC-DESIGN-DOCSV2-001 is the predecessor (AUTHORIZED unfreeze
  documented in §D FROZEN design baseline clause + §F.2). No BLOCKING finding.

- **[PASS] MP-6 D8 cross-platform discipline.** SPEC body contains zero
  occurrences of the literal `syscall`. Auto-PASS per D8-4.

- **[FAIL] MP-7 clarification gate.** `grep -rn '\[NEEDS CLARIFICATION'`
  against research.md returns **3 matches**:
  - research.md:199 `[NEEDS CLARIFICATION: 12-vs-13 section reconciliation]`
  - research.md:200 `[NEEDS CLARIFICATION: book CTA target]`
  - research.md:201 `[NEEDS CLARIFICATION: mascot asset licensing]`

  Per the MP-7 contract, any match is a must-pass failure that forces
  `Verdict: FAIL` regardless of aggregate score. progress.md §E.1 flags all
  three as "non-blocking design decisions, suitable for Implementation
  Kickoff Approval resolution" — that framing is **incorrect**. The MP-7
  contract states: "a high aggregate score never auto-resolves an open
  clarification marker." The orchestrator MUST resolve each via
  `AskUserQuestion` (preload `ToolSearch(query: "select:AskUserQuestion")`)
  BEFORE Implementation Kickoff Approval, then the auditor re-verifies.
  The 12-vs-13 marker specifically blocks M0 (IA freeze) exit criteria
  (AC-M0-001, AC-M0-003) — M0 cannot produce a "frozen" IA while the
  section count is undecided.

---

## Category Scores (0.0–1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.70 | 0.50 band | REQ-KO-006 "friend explainability" is a self-audit with no objective standard; the `wc -w` rubric description (§A.2 "excluding code blocks, tables, frontmatter, shortcode arguments, HTML comments") diverges from the actual §C.7 awk which strips ONLY fenced code blocks; REQ-DS-004 references `moai-logo-4-W` but the staged file is `moai-logo-4-WH.png`. Multiple requirements require interpretation. |
| Completeness | 0.90 | 0.75 band | All required sections present (HISTORY §A, WHY/Context, WHAT/Goals, REQUIREMENTS §C, ACCEPTANCE in acceptance.md, Risks §E, Dependencies §F, Out of Scope §B.2). Six `### Out of Scope — <topic>` H3 sub-headings with `-` bullets (dark theme / handoff runtime JS / moai web / content tooling / URL restructure / analytics-search-comments). Frontmatter complete. One non-critical gap: no dedicated design.md (Tier L optional in practice — research.md absorbs the design analysis). |
| Testability | 0.70 | 0.50 band | The "book-level prose" AC is the load-bearing contribution and is **measurable but flawed** — see D3. AC-M4-X-003 ("friend-explainability trail non-vacuous") cannot be evaluated without subjective judgment ("vacuous" has no mechanical definition). AC-M0-003's "structural diff tolerance" is MAY-level. Several ACs require judgment calls. |
| Traceability | 0.88 | 0.75 band | Every REQ has ≥1 AC; every AC references a REQ that exists. REQ→AC mapping is dense (29 REQ, 32 AC). One indirect mapping: REQ-KO-006 → AC-M4-X-003 (the friend-explainability trail) — the mapping exists but the AC verifies a sidecar artifact, not the rubric outcome. No orphaned ACs. |

Aggregate (simple mean): **0.795**. Harmonic mean (penalizes the low
dimensions, per skeptical-stance doctrine): **0.772**. Both below the
Tier L PASS threshold of 0.85. Combined with MP-7 FAIL, verdict is FAIL.

---

## Defects Found (structured defect-list)

### D1 — MP-7 clarification gate violation
- **Artifact**: research.md:199-201
- **Description**: Three unresolved `[NEEDS CLARIFICATION]` markers exist at
  audit time. progress.md §E.1 characterizes them as "non-blocking design
  decisions, suitable for Implementation Kickoff Approval resolution," but
  the MP-7 contract is score-independent and treats any open marker as a
  must-pass failure. The 12-vs-13 marker specifically blocks M0 (the first
  milestone) — IA freeze cannot exit without it.
- **Severity**: critical
- **Class**: blocking
- **Required fix**: Before re-audit, the orchestrator MUST run three
  `AskUserQuestion` rounds resolving (a) 12-vs-13 section reconciliation
  (fold cost-optimization into multi-llm/advanced vs retain as 13th card),
  (b) book CTA target (verify book.mo.ai.kr exists or omit the card), (c)
  mascot asset licensing (confirm Apache-2.0 / project license or document
  attribution). Each resolution is then written into spec.md §G (design
  decisions) or §B.2 (out-of-scope with rationale), and the `[NEEDS
  CLARIFICATION]` marker in research.md is replaced with `[RESOLVED:
  <decision>]`. Re-audit verifies zero marker matches.

### D2 — REQ/AC count exceeds Tier L ceiling (29 > 25 REQ, 32 > 25 AC)
- **Artifact**: spec.md §C (29 REQ entries), acceptance.md §B (32 AC entries)
- **Description**: `spec-workflow.md` § SPEC Complexity Tier caps Tier L at
  "25 requirements" and "25 acceptance criteria" (ceilings apply
  independently). This SPEC carries 29 REQs and 32 ACs — both over-budget.
  The taxonomy's own guidance: "an over-budget SPEC is the same
  over-formalization failure the tier taxonomy exists to prevent."
  The plan.md §G.1 Epic-split escape valve addresses M4 (content) scope,
  but does NOT address the REQ/AC budget overrun.
- **Severity**: major
- **Class**: blocking
- **Required fix**: Pick one — (a) trim REQ count to ≤25 by consolidating
  the NEW-badge group (REQ-NB-001/002/003/004 could collapse to 2 REQs:
  one for the flag mechanism, one for the rendering surfaces), and
  consolidate REQ-KO-006 into REQ-KO-001 (the rubric already enforces
  friend-explainability via AC-M4-X-003, so the standalone REQ is
  redundant); OR (b) formally invoke the Epic-split NOW (at plan-phase)
  rather than deferring it as a "documented option" — split M4 + the KO
  REQs into `SPEC-DESIGN-DOCS-V31-CONTENT-KO-001` so each SPEC lands
  within budget. Option (a) is the smaller revision.

### D3 — "Book-level prose" rubric: measurement command diverges from rubric description
- **Artifact**: acceptance.md §C.7 (the awk pipeline) vs §A.2 (rubric
  description: "Word count is measured on Korean prose (body text
  excluding code blocks, tables, frontmatter, shortcode arguments, HTML
  comments)")
- **Description**: The actual measurement command is:
  ```bash
  awk '/^```/{c++; next} c%2==2' "$f" | wc -w
  ```
  This pipeline strips ONLY fenced code blocks (lines starting with
  ``` at column 0). It does NOT strip: (1) YAML frontmatter between
  `---` fences — every page's frontmatter contributes ~20-50 spurious
  "words" (`id:`, `title:`, `weight:`, etc.); (2) table rows — a 9-row
  reference table inflates the count by ~50+ pipe-separated tokens; (3)
  shortcode arguments `{{< mascot pointing >}}` — counted as 3 words; (4)
  HTML comments. The measurement therefore systematically over-counts
  versus the rubric's stated semantics. A page that is genuinely under
  the floor can pass by padding with tables and shortcode calls — the
  rubric is gaming-able. The awk's fence-state logic is otherwise correct
  (traced: c=0 → outside, c=1 → inside, c=2 → outside), but its scope is
  too narrow.
- **Severity**: critical (this is the load-bearing AC)
- **Class**: blocking
- **Required fix**: Replace the §C.7 pipeline with one that matches the
  rubric. Concretely:
  ```bash
  awk '
    /^---[[:space:]]*$/ && NR<=3 {fm=!fm; next}    # strip frontmatter
    fm {next}
    /^```/{c++; next}                              # strip code fences
    c%2==1 {next}
    /^\|/{next}                                    # strip table rows
    /^\s*<!--/{in_html=1} /^\s*-->/{in_html=0; next}   # strip HTML comments
    in_html {next}
    {gsub(/\{\{[^}]*\}\}/, ""); gsub(/\{\{<[^>]*>\}\}/, ""); print}  # strip shortcodes
  ' "$f" | wc -w
  ```
  OR re-state §A.2 to honestly declare what the awk measures ("body text
  excluding fenced code blocks only; frontmatter, tables, and shortcodes
  are included in the count") and re-calibrate the per-class floors
  against that honest metric. The first option is preferred — the floors
  were calibrated against "real prose," so the measurement should match.

### D4 — REQ-KO-006 / AC-M4-X-003 friend-explainability is structurally unverifiable
- **Artifact**: spec.md REQ-KO-006, acceptance.md §A.4 + AC-M4-X-003
- **Description**: REQ-KO-006 requires "the answer MUST be demonstrably
  yes" to the friend-explainability question, but §A.4 admits "the rubric
  is enforced by a self-audit step recorded in the page's authoring trail,
  not by a runtime check." AC-M4-X-003 then verifies "every page has a
  non-vacuous two-sentence summary recorded." The term "non-vacuous" has
  no mechanical definition — an author who writes "this page explains X
  and shows Y" passes the AC, regardless of whether a real reader could
  explain the concept. The AC verifies the *presence of a sidecar text*,
  not the *outcome* the rubric names. This is the weakest AC in the SPEC.
- **Severity**: major
- **Class**: blocking
- **Required fix**: Either (a) demote REQ-KO-006 + AC-M4-X-003 to a
  SHOULD-level soft rubric (remove from MUST ACs in §D), acknowledging it
  as an authoring best-practice rather than a binary gate; OR (b) tighten
  the AC to require the two-sentence summary to contain at least one
  causal connector ("왜냐하면", "때문에", "결과적으로") and at least one
  concrete noun from the page's H1 — still imperfect but mechanically
  testable.

### D5 — REQ-DS-004 mascot/logo filename mismatch
- **Artifact**: spec.md REQ-DS-004 (line ~105)
- **Description**: REQ-DS-004 names the logo variants as `moai-logo-1`,
  `moai-logo-4`, `moai-logo-4-W`. The actual staged file at
  `/tmp/moai-design-handoff/moai-adk/project/assets/` is `moai-logo-4-WH.png`
  (verified by `ls`). The `-W` vs `-WH` mismatch will cause M1's
  asset-onboarding step to either rename the file (introducing a name not
  present in the handoff) or fail to find it.
- **Severity**: minor
- **Class**: blocking (cheap fix, blocks M1)
- **Required fix**: Update REQ-DS-004 to read `moai-logo-4-WH` (or rename
  the staged file to `moai-logo-4-W.png` before M1 — but the SPEC is the
  place to record the canonical name).

### D6 — Handoff README provenance scope wider than docs-site
- **Artifact**: research.md §A.1 (tone/copy/color/type sourced from handoff README)
- **Description**: The handoff README opens with "모두의AI (mo.ai.kr) 는
  한국의 AI 사용자가 모이는 커뮤니티 플랫폼" and lists product surfaces
  (`/projects`, `/beta`, `/news`, `/academy`) and 6 upcoming consumer
  services (모두의 사주, 바닐라 바게트, etc.) — this is the *community
  platform* context, NOT the moai-adk-go docs-site context. The 6 `.dc.html`
  screen prototypes ARE docs-site screens ("01 Docs Home", "02 Getting
  Started", etc.), so the design vocabulary (tokens, type, voice) is
  brand-wide and applicable to docs-site. But REQ-KO-005 imports the
  community-platform voice rules ("모두의" appears at least once on the
  home page) verbatim into docs-site — this may or may not be the user's
  intent.
- **Severity**: minor
- **Class**: optional
- **Required fix**: Add a one-paragraph note in research.md §A.1
  acknowledging that the README's product context (community platform)
  differs from the docs-site surface, and that the SPEC deliberately
  extracts the brand-wide voice/tone/color rules while ignoring the
  community-platform-specific product copy. No REQ change needed.

### D7 — Non-standard GEARS label "(Event-detected)"
- **Artifact**: spec.md REQ-KO-003, REQ-I18N-002
- **Description**: Both entries are labeled `(Event-detected)` but the
  canonical GEARS pattern label is "Event-driven". The structure
  (`When … SHALL`) is correct, so MP-2 passes — but the label is
  non-standard and could confuse a future reader.
- **Severity**: minor
- **Class**: optional
- **Required fix**: Rename the label from `(Event-detected)` to
  `(Event-driven)` in both REQ entries.

### D8 — "Epic-split escape valve" trigger is vague
- **Artifact**: plan.md §F M4 "Epic-split escape valve" paragraph;
  spec.md §G.1
- **Description**: The escape valve fires "if at any point the orchestrator
  (or the user at a mid-run checkpoint) judges M4's scope exceeds a single
  run-phase's manageable span." No concrete trigger — page-count threshold,
  turn-count ceiling, sub-milestone failure count, or wall-time signal.
  The mission asks: is this "defined precisely enough to trigger mid-run,
  or is it vague?" It is vague.
- **Severity**: major
- **Class**: blocking
- **Required fix**: Add a concrete trigger to plan.md §F M4. Suggested:
  "The Epic-split is auto-recommended when EITHER (a) the cumulative
  turn count across M4.1-M4.N exceeds 1.5× the per-milestone ceiling
  tracked in progress.md, OR (b) any single sub-milestone fails its
  §B.5 AC block twice with the same root cause. The orchestrator
  surfaces the split recommendation via `AskUserQuestion` with three
  options: split-now / continue-with-debt / abort."

### D9 — M5 "Korean verification gate" structure risks theater
- **Artifact**: plan.md §F M5; acceptance.md §B.6
- **Description**: M5 is labeled "a GATE, not a work item — its output is a
  pass/fail decision." Its three ACs (AC-M5-001..003) overlap with the M4
  sub-milestone gates (AC-M4-X-001..004 already verify per-section) and
  with M7 (AC-M7-001..006 verifies the same things 4-locale). The M5 gate
  adds value only if it is a *strict* gate — i.e. M6 (translation) cannot
  begin until M5 passes. plan.md does say "if any AC fails, M4 is
  revisited before M6 begins" — but M5 is a SHOULD-priority milestone per
  acceptance.md §D ("SHOULD — AC-M6-X-001..003" are SHOULD, and M5 is the
  gate for them). The gate's MUST/SHOULD priority is unclear.
- **Severity**: minor
- **Class**: optional
- **Required fix**: Promote AC-M5-001, AC-M5-002, AC-M5-003 to MUST in
  §D severity classification, and add a sentence to plan.md §F M5 making
  explicit: "M6 (en derivation) MUST NOT begin until every MUST AC in
  §B.6 passes; this is the ko→en gate."

### D10 — ko-first i18n wall-clock cost acknowledged but not quantified
- **Artifact**: plan.md §B.5 "Sequential i18n wall-time"
- **Description**: plan.md acknowledges "ko → en → ja → zh with each locale
  fully verified before the next means the i18n phase (M5–M7) is the
  longest serial segment. Parallelization is explicitly forbidden by the
  user decision. The plan does NOT attempt to parallelize; instead it
  minimizes per-locale wall-time by pre-computing the translation manifest
  in M3." Honest — but never quantifies the multiplier. The user decision
  (Q4) is ko-first sequential; the cost is roughly 3× the per-locale
  derivation time (en + ja + zh each fully verified). The plan should
  state this explicitly so the user at Implementation Kickoff Approval
  can make an informed decision about the Epic-split.
- **Severity**: minor
- **Class**: optional
- **Required fix**: Add to plan.md §B.5 a one-line quantification:
  "Sequential i18n is approximately 3× the wall-clock of a single
  locale derivation; the manifest pre-computation in M3 reduces the
  per-locale constant but does not collapse the 3× multiplier. The
  user's ko-first decision (Q4) is the binding constraint."

---

## Regression Check
(Iteration 1 — no prior iteration to regress against.)

---

## Recommendation

**Implementation Kickoff Approval: BLOCKED** — return to manager-spec for
revision. The SPEC cannot enter run-phase until the following are resolved:

### Must fix before re-audit (CRITICAL/MAJOR blocking)

1. **Resolve the 3 [NEEDS CLARIFICATION] markers** (D1). The orchestrator
   runs three `AskUserQuestion` rounds, the answers are written into
   spec.md §G / §B.2, and the research.md markers become `[RESOLVED: …]`.
   The 12-vs-13 reconciliation specifically blocks M0 — it cannot be
   deferred past the gate.

2. **Bring REQ/AC count within Tier L ceiling (≤25 each)** (D2). Either
   consolidate REQ-NB and REQ-KO-006, OR formally invoke the Epic-split
   NOW at plan-phase (split M4 + KO REQs into a child SPEC).

3. **Fix the `wc -w` measurement command** (D3) so it actually strips
   frontmatter, tables, shortcode arguments, and HTML comments as §A.2
   describes — OR re-state §A.2 to honestly declare what the awk measures
   and re-calibrate the floors.

4. **Tighten or demote REQ-KO-006 / AC-M4-X-003** (D4) — the
   friend-explainability AC is currently unverifiable.

5. **Fix REQ-DS-004 mascot filename** (`moai-logo-4-W` → `moai-logo-4-WH`)
   (D5).

6. **Add a concrete trigger to the Epic-split escape valve** (D8) —
   page-count threshold or turn-count ceiling, not "if the orchestrator
   judges."

### Optional polish (MINOR, non-blocking)

7. Acknowledge handoff README's community-platform provenance in
   research.md §A.1 (D6).
8. Rename GEARS label `(Event-detected)` → `(Event-driven)` in REQ-KO-003
   and REQ-I18N-002 (D7).
9. Promote M5 gate ACs to MUST and add the explicit "M6 MUST NOT begin
   until M5 passes" sentence (D9).
10. Quantify the ko-first 3× wall-clock multiplier in plan.md §B.5 (D10).

### Rationale for the blocking recommendation

The Tier L PASS threshold is 0.85; the aggregate score is 0.76-0.80
depending on averaging method. More importantly, three CRITICAL blocking
findings stand:

- MP-7 is a hard must-pass failure (3 open clarification markers).
- The "book-level prose" rubric — explicitly identified in the mission as
  the load-bearing AC — has a measurement pipeline that does not match
  its own description (D3). A rubric that claims objectivity but is
  gaming-able is worse than an honest subjective rubric, because it
  transfers blame to the auditor who flags the gap.
- The REQ/AC count exceeds the Tier L ceiling (D2) — the SPEC's own
  tier taxonomy says over-budget is the over-formalization failure the
  taxonomy exists to prevent.

The M4 scale concern (D8) is real but manageable IF the Epic-split
trigger is concretely defined. As written, "if the orchestrator judges"
is too vague to act on mid-run.

The design-handoff verification was clean: token-diff claims in research.md
§D match the actual `colors_and_type.css` and `moai-brand.css` byte-for-byte
on the PRIMARY/INK/BG trio and the neutral-ramp deltas. Mascot/logo assets
are present (6 + 3). The NEW-badge mechanism (REQ-NB-001..004) is sound
in design — dual-source (frontmatter flag + `_meta.yaml` list + shortcode)
with union semantics is a clean pattern, and the `added_in: "v3.1"`
preferred form correctly enables a future mechanical sunset sweep. The
SPEC's structural engineering is good; the defects are in measurement
fidelity, scope budget, and unresolved clarifications.

Resolving D1-D5 + D8 should bring the SPEC to ~0.87 and clear the MP-7
gate. Re-audit at iteration 2 is scoped to the enumerated defect delta.

---

# SPEC Review Report: SPEC-DESIGN-DOCS-V31-001
Iteration: 2/3
Verdict: PASS
Overall Score: 0.87  (Tier L PASS threshold = 0.85)

> Iteration-2 re-audit. Scope: enumerated defect delta from iteration 1
> (D1–D10) plus a fresh adversarial pass for new defects the consolidation
> may have introduced. M1 Context Isolation active — author reasoning in
> the v0.2.0 HISTORY entry and progress.md is treated as claim, not evidence;
> every fix is verified against the actual file content.

---

## Must-Pass Results (iteration 2)

- **[PASS] MP-1 REQ number consistency.** 25 REQs verified by
  `grep -oE '^\*\*REQ-[A-Z]+-[0-9]+\*\*' spec.md | sort -u | wc -l` → 25.
  Groups: IA-001..004 (4), NB-001..002 (2), DS-001..007 (7), KO-001..004 (4),
  I18N-001..005 (5), BL-001..003 (3). No gaps, no duplicates. Within the
  Tier L ceiling of 25 (spec-workflow.md § SPEC Complexity Tier). D2 RESOLVED.

- **[PASS] MP-2 EARS/GEARS format compliance** (requirement layer). All 25
  REQs match one of the five GEARS patterns. Spot check: REQ-IA-003 / REQ-KO-003
  / REQ-I18N-002 `When … SHALL` (Event-driven — D7 label fix verified);
  REQ-IA-004 / REQ-DS-002 / REQ-KO-002 / REQ-I18N-003 `While … SHALL/MUST`
  (State-driven); REQ-NB-002 compound `[While …][Where …] SHALL` (PASS-equivalent
  per M3 compound-clause rule); REQ-BL-002 `Where … SHALL NOT` (Capability gate);
  REQ-DS-003 `SHALL NOT` (Unwanted canonical negative); the majority Ubiquitous
  `The <subject> SHALL`. Judgment binds the `REQ-XXX` requirement layer only;
  the 25 ACs are Given-When-Then by design (verification layer) and are not
  modality-checked here. No new GEARS malformation introduced by the
  iteration-2 consolidations.

- **[PASS] MP-3 YAML frontmatter validity.** All 12 canonical fields present
  with correct types, verified field-by-field against
  `.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12.
  `id: SPEC-DESIGN-DOCS-V31-001` (string, regex match); `version: "0.2.0"`
  (quoted semver); `status: draft` (enum-valid); `created: 2026-08-11`,
  `updated: 2026-08-11` (ISO dates); `priority: High` (enum-allowed alias of
  P1); `phase: "v3.1-rc.1 target"` (release target — NOT a prohibited
  lifecycle token per the whole-value comparison rule); `module: docs-site`;
  `lifecycle: spec-anchored`; `tags:` comma-separated string. `tier: L` and
  `related_specs:` are optional fields, correctly used. No rejected snake_case
  aliases (`created_at` / `updated_at` / `labels` / `spec_id` all absent).

- **[N/A → PASS] MP-4 Section 22 language neutrality.** SPEC is scoped to a
  single-language docs-site (Hugo + Korean-content-first). No multi-language-
  tool enumeration obligation. REQ-BL-002 carries the template-neutrality
  constraint for any mirrored artifact. Auto-passes per MP-4 N/A precedent.

- **[PASS] MP-5 D7 cross-SPEC reconciliation.** No referenced SPEC is in a
  `retired` / `superseded` / `archived` state. Iteration-1 §F.1 catalog
  verified each referenced SPEC's `status:` — 11 `completed` + 1 `in-progress`
  (SPEC-AGENT-MODEL-ENFORCE-001, explicitly flagged "partial"). The
  `related_specs:` frontmatter list is a subset of the §F.1 catalog; no new
  cross-SPEC reference introduced at v0.2.0. No BLOCKING finding.

- **[PASS] MP-6 D8 cross-platform discipline.** SPEC body contains zero
  occurrences of the literal `syscall` (`grep -c 'syscall' spec.md` → 0).
  Auto-PASS per D8-4.

- **[PASS] MP-7 clarification gate.**
  `grep -rn '\[NEEDS CLARIFICATION:' spec.md plan.md acceptance.md research.md progress.md`
  → 0 matches. The three v0.1.0 markers in research.md §E (lines 201–203)
  are now `[RESOLVED: …]` with the user decision recorded verbatim. The only
  remaining textual matches for the string `NEEDS CLARIFICATION` (uncolonned)
  live in this audit.md file (historical iteration-1 record) and in two
  prose sentences inside research.md:205 and progress.md:15 that *state* "zero
  markers remain" — neither is an open marker. D1 RESOLVED.

---

## Per-Finding Resolution Table (iteration-1 D1–D10)

| ID | Severity (iter-1) | Status | Evidence (iteration-2) |
|----|------|--------|------------------------|
| **D1** | critical (MP-7) | **RESOLVED** | research.md:201–203 — all three markers rewritten `[RESOLVED: …]` with the user decision (12-section IA / book.mo.ai.kr HTTP 200 / Apache-2.0). 12-section decision cascades to spec.md §B.1 goal #1 (L41), REQ-IA-001 (L80), plan.md M0 scope (L60) + exit criteria (L62), acceptance.md AC-M0-001 (L78). Book CTA URL verified-live text present in acceptance.md AC-M2-002 (L96). Mascot Apache-2.0 text present in research.md §E third bullet. `grep -rn '\[NEEDS CLARIFICATION:' {spec,plan,acceptance,research,progress}.md` → 0 matches. |
| **D2** | major | **RESOLVED** | REQ count 29 → 25 (`grep -oE '^\*\*REQ-…' spec.md \| sort -u \| wc -l` → 25). AC count 32 → 25 (`grep -c '^\*\*AC-' acceptance.md` → 25). Both within Tier L ceiling. Every v0.1.0 REQ has a v0.2.0 successor with a documented merge rationale — no silent drop. Merge map: NB-001+002→NB-001 (mechanism+shortcode, same indicator surface); NB-003+004→NB-002 (sidebar+header+section, same flag→render surface); KO-004→KO-001 pillar 3 (infographic floor belongs to rubric); KO-005→KO-004 (renumber); KO-006→KO-001 pillar 5 (friend-explainability belongs to rubric). |
| **D3** | critical (load-bearing AC) | **RESOLVED** | acceptance.md §C item 7 (L158–175): the awk pipeline now strips (a) YAML frontmatter between first two `---` fences, (b) fenced code blocks, (c) pipe-table rows, (d) Hugo shortcodes (`{{< … >}}` and `{{ … }}`), (e) HTML comments. State-machine traced on a 23-line synthetic fixture (frontmatter + table + shortcode + code block + 4-line `<!-- … -->` block + prose): output 7 words (5 + 2), correct. The mission-critical edge case "HTML comments spanning multiple lines" — traced: opener line `/^[[:space:]]*<!--/` sets `in_html=1`; intermediate lines skipped by `in_html {next}`; closer line `/-->[[:space:]]*$/` resets `in_html=0`. Verified on real ko samples per the orchestrator's independent run (what-is-moai-adk.md 2133 words, trust-5.md 599 words — both reported by manager-spec; not re-measured by this auditor but the pipeline logic is correct). Two narrow residual edge cases noted (see N1 below) — not blocking. |
| **D4** | major | **RESOLVED** | acceptance.md §A.4 (L45–64): friend-explainability predicate is now mechanical. (5a) Korean causal connector grep: `왜냐하면\|때문에\|따라서\|그래서\|덕분에` (5-entry allowlist). (5b) Concrete capability noun grep: `SPEC\|TRUST 5\|harness\|goal\|factory\|에이전트\|관리자 에이전트\|Geekdoc\|Pretendard` (9-entry allowlist). (5c) ≥60 bytes sidecar size. Three grep commands provided; no NLP judgment required. Promoted to MUST via AC-M5-004. Allowlist coverage check: `에이전트` + `SPEC` + `harness` cover the dominant M4 vocabulary (agent / SPEC-workflow / harness are universal concepts in MoAI docs); the allowlist is not so narrow as to be vacuous. |
| **D5** | minor (blocking) | **RESOLVED** | spec.md REQ-DS-004 (L102): `moai-logo-4-WH` (corrected). The historical typo `moai-logo-4-W` appears only inside the D5-fix parenthetical explaining the correction — not as a canonical asset name. |
| **D6** | minor (optional) | **RESOLVED** | research.md §A.1 (L13): "Provenance note (v0.2.0 D6 fix)" paragraph explicitly scopes the handoff README as community-platform voice and documents that REQ-KO-004 imports only the subset relevant to docs-site. |
| **D7** | minor (optional) | **RESOLVED** | spec.md REQ-KO-003 (L116) and REQ-I18N-002 (L124): both labeled `(Event-driven)`. The historical `(Event-detected)` label survives only inside the D7-fix parentheticals. |
| **D8** | major | **RESOLVED** | plan.md §F M4 (L97–103): Epic-split escape valve now has three concrete triggers — T1 (per-sub-milestone turn overshoot N=40), T2 (cumulative run-phase turn overshoot T=200), T3 (any single sub-milestone fails its §B.5 AC block twice with the same root cause). The split-surfaces-via-AskUserQuestion with three options (split-now / continue-with-debt / abort). spec.md §G.1 (L203–205) references the quantified triggers. |
| **D9** | minor (optional) | **RESOLVED** | plan.md §F M5 (L105–109): "M6 (en derivation) MUST NOT begin until every MUST AC in §B.6 passes; this is the ko→en gate." acceptance.md §B.6 header (L120) repeats the M6-gating sentence normatively. acceptance.md §D severity classification (L183): AC-M5-001..004 listed under MUST. |
| **D10** | minor (optional) | **RESOLVED** | plan.md §B.5 known-issue #5 (L21): "the i18n phase is approximately **3× the wall-clock of a single locale derivation**" with full quantification explaining M3's manifest pre-computation reduces the per-locale constant but does not collapse the 3× multiplier, and that the user's ko-first decision is the binding constraint. |

**Resolution tally: 10/10 RESOLVED. 0 PARTIAL. 0 NOT-RESOLVED.**

---

## Category Scores (0.0–1.0, rubric-anchored — iteration 2)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.90 | 0.75 band (high side) | All 25 REQs have single unambiguous interpretations. Friend-explainability is now a mechanical predicate (D4). The wc measurement pipeline matches its rubric description (D3). Mascot filename corrected (D5). Epic-split trigger quantified (D8). Residual: the §C.7 awk has two narrow edge cases (see N1) — a tooling imperfection, not a requirement-text ambiguity. |
| Completeness | 0.85 | 0.75 band | All required sections present (HISTORY §A, WHY §A, WHAT §B, REQUIREMENTS §C, ACCEPTANCE in acceptance.md, Risks §E, Dependencies §F, Out of Scope §B.2 with six H3 sub-headings). Frontmatter complete (12 canonical fields). One structural gap: no dedicated `design.md` (Tier L 5-artifact set per spec-frontmatter-schema.md calls for spec+plan+acceptance+design+research; this SPEC ships spec+plan+acceptance+research+progress instead). The design-analysis content (token DIFF §D, screen-by-screen §A.2–§A.4) is absorbed into research.md — substantive coverage exists, structural form does not. Same non-critical gap noted at iteration 1; unchanged at v0.2.0. |
| Testability | 0.90 | 0.75 band (high side) | All 25 ACs are binary-testable. The load-bearing prose-floor AC now has a corrected pipeline (D3). The friend-explainability AC has a mechanical predicate (D4). M5 promoted to MUST with normative M6-gating (D9). Residual: REQ-DS-002 (prefers-reduced-motion → 1ms degradation) has no direct behavioral AC (see N2); the awk inline-comment edge case (N1) could produce measurement drift on pages that use inline `<!-- … -->` comments. |
| Traceability | 0.85 | 0.75 band | Every REQ has ≥1 AC. Two indirect mappings: REQ-DS-002 (reduced-motion) — no direct AC verifies the 1ms-degradation behavior; REQ-BL-002 (template-neutrality) — enforced by the existing CI guard `internal/template/internal_content_leak_test.go` (per CLAUDE.local.md §25), not by a SPEC AC. Both are invariants enforced by other mechanisms; the indirect mapping is documented, not absent. No orphaned ACs. Every AC references a milestone whose concern maps to one or more REQs. |

Simple mean: **0.875**. Harmonic mean (per skeptical-stance doctrine): **0.874**.
Both above the Tier L PASS threshold of 0.85. Combined with all 7 must-pass
criteria PASS, verdict is PASS.

---

## Defects Found (iteration 2 — new findings, severity-ranked)

> All findings below are **optional** (Class: optional). None are blocking.
> The v0.2.0 SPEC clears every must-pass criterion and every iteration-1
> blocking defect. These are residuals a future sync-phase amendment MAY
> address; they do not gate Implementation Kickoff Approval.

### N1 — §C.7 awk: two narrow HTML-comment edge cases leak
- **Artifact**: acceptance.md §C item 7 (L161–174 awk pipeline)
- **Description**: The HTML-comment state machine has two narrow edge cases
  verified by direct trace:
  (a) **Inline opener not at column 0.** A line like `text <!-- comment`
  does NOT match `/^[[:space:]]*<!--/`, so `in_html` is never set and the
  comment text leaks into the word count. Subsequent lines until `-->` are
  also not stripped.
  (b) **Trailing text after closer on opener line.** A line like
  `<!-- comment --> trailing` matches the opener rule (sets `in_html=1`),
  but does NOT match the closer rule `/-->[[:space:]]*$/` (line ends with
  "trailing", not `-->`). The line is skipped by `in_html {next}`, AND
  `in_html` stays `1` — subsequent prose lines are stripped until the next
  line ending in `-->`.
- **Severity**: minor
- **Class**: optional
- **Why non-blocking**: Hugo markdown authors typically place HTML comments
  on their own lines (block-level). The mission's test corpus
  (what-is-moai-adk.md, trust-5.md) was measured and produced sane numbers,
  suggesting the inline patterns are absent from the actual content. The
  rubric description in §A.2 ("HTML comments") does not specify inline vs
  block, so the tool's behavior is a defensible interpretation.
- **Required fix (optional, defer to run-phase or sync-phase amendment)**:
  extend the closer regex to `/.*-->[[:space:]]*$/` (matches `-->` anywhere
  followed by optional whitespace) and add an inline-comment gsub
  `gsub(/<!--[^>]*-->/, "")` before the print, OR re-state §A.2 to
  honestly declare the tool strips block-level HTML comments only.

### N2 — REQ-DS-002 (prefers-reduced-motion) has no direct AC
- **Artifact**: spec.md REQ-DS-002 (L98); acceptance.md §B (no AC references
  reduced-motion behavior)
- **Description**: REQ-DS-002 mandates `prefers-reduced-motion: reduce`
  degradation to 1ms for every transition. AC-M1-001 verifies the token
  vocabulary byte-for-byte (which includes `--easing-bounce` etc.) but does
  not verify the runtime behavior under reduced-motion. The REQ is covered
  indirectly (the tokens are present; the behavior follows) but a behavioral
  grep like `grep -E 'prefers-reduced-motion' docs-site/static/*.css` would
  close the gap directly.
- **Severity**: minor
- **Class**: optional
- **Required fix (optional)**: add an AC-M1-005 "Given the production CSS,
  when grepped for `prefers-reduced-motion`, then a media-query block is
  present that sets every transition-duration to 1ms" — OR accept the
  indirect coverage via the token-port AC.

### N3 — Changelog vercel.json redirect: navigation change vs URL change
- **Artifact**: plan.md §F M0 scope (L60) + exit criteria (L62)
- **Description**: plan.md M0 says "Emit the `vercel.json` redirect entries
  for the `changelog` slug change (sidebar → footer link)". Moving changelog
  from the sidebar to the footer is a navigation-placement change, not a URL
  change — the `/changelog/` URL likely continues to resolve to the same
  page. A vercel.json redirect is needed only when the URL slug itself
  changes. plan.md M0 exit criterion (2) explicitly allows an empty redirect
  diff ("MAY be empty if no other slugs changed"), so the run-phase can
  resolve this correctly — but the plan wording conflates navigation
  placement with URL change.
- **Severity**: minor
- **Class**: optional
- **Required fix (optional)**: clarify in plan.md M0 that the vercel.json
  redirect is required only IF the changelog URL slug itself changes; if
  only the navigation placement changes (sidebar → footer link, URL
  unchanged), no redirect entry is needed.

---

## Regression Check (iteration 2 vs iteration 1)

All 10 iteration-1 defects (D1–D10) verified RESOLVED above. No regression
on any dimension: Clarity 0.70 → 0.90 (+0.20), Completeness 0.90 → 0.85
(−0.05; the iteration-1 score of 0.90 was generous given the missing
design.md — iteration 2 scores it more honestly against the rubric band),
Testability 0.70 → 0.90 (+0.20), Traceability 0.88 → 0.85 (−0.03; the
iteration-1 traceability score did not flag REQ-DS-002 / REQ-BL-002 as
indirect — iteration 2 does). Aggregate harmonic mean: 0.772 → 0.874
(+0.10). Score regression check: iteration-2 score is HIGHER than
iteration-1 — the LEAN workflow STOP-on-regression trigger does not fire.

The Completeness and Traceability scores were re-anchored to the rubric
bands more strictly at iteration 2; the dimension scores are NOT comparable
across iterations 1→2 because iteration 1 applied the bands loosely. The
iteration-2 scores are the authoritative reading.

---

## Recommendation

**Implementation Kickoff Approval: PROCEED** — the SPEC is ready for
run-phase entry.

### Rationale (per must-pass criterion)

- MP-1 (REQ consistency): 25 REQs, sequential, no gaps/duplicates. PASS.
- MP-2 (GEARS): all 25 REQs match a canonical pattern. PASS.
- MP-3 (frontmatter): all 12 canonical fields present, correct types. PASS.
- MP-4 (language neutrality): N/A (single-language docs-site scope). PASS.
- MP-5 (D7 cross-SPEC): no referenced SPEC retired/superseded/archived.
  PASS.
- MP-6 (D8 syscall): zero `syscall` in body. PASS.
- MP-7 (clarification gate): zero open `[NEEDS CLARIFICATION:` markers. PASS.

### Score rationale

Aggregate harmonic mean 0.874 > Tier L threshold 0.85. All seven must-pass
criteria PASS. All ten iteration-1 defects RESOLVED. The three new
iteration-2 findings (N1/N2/N3) are all Class: optional — none block the
run-phase. The LEAN STOP-on-regression signal does not fire (iter-2 score
> iter-1 score).

### Optional debt to surface at Implementation Kickoff Approval

The orchestrator MAY surface N1/N2/N3 to the user as SHOULD-level debt the
run-phase MAY address (not MUST):
- N1 (awk edge cases): narrow; the run-phase MAY harden the stripper or
  accept the current coverage.
- N2 (REQ-DS-002 reduced-motion AC): the run-phase MAY add a behavioral AC.
- N3 (changelog redirect wording): the run-phase M0 will resolve this when
  it inspects whether the changelog slug actually changes.

None of these prevent the kickoff. TheSPEC's structural engineering is
sound; the v0.2.0 revision substantively addresses every blocking defect
from iteration 1, and the new adversarial pass found only optional
residuals. The Epic-split escape valve (D8) is concretely quantified; the
ko→en gate (D9) is normative; the load-bearing prose-floor AC (D3) now
matches its description. The SPEC is audit-ready.
