# Implementation Plan — SPEC-STEERING-ALIGN-OUTPUT-STYLE-SLIM-001

> Tier M. Epic Steering-Align SPEC 4 of 5 (P5). Lifecycle: plan → run → sync (3-phase).
> The CORE deliverable is §C.3 — the per-passage KEEP / CUT / POINTER classification table.
> Run-phase execution SHALL be mechanical from §C.3; no run-phase judgment call about render-SSOT-vs-duplicated-prose.

---

## A. Context

This SPEC diets the always-loaded output-style `moai.md` (782 lines / 55306 bytes, both trees byte-identical). §8 Response Templates is 66% of the file (L211-731). The diet is MODERATE (user-confirmed): §8 Session Handoff pointer-ization + §8 duplicate-prose deletion, gated by the render-SSOT preservation invariant (spec.md §A.3). It mirrors the COMPLETED sibling P2 CLAUDEMD-DIET-001 technique (M-DELETE + M-POINTER) and inherits its AC-009 over-cut gate verbatim (here AC-OSS-006).

### A.5 PRESERVE list (do NOT touch — B8/B10 working-tree hygiene)

The run-phase touches EXACTLY these paths:

- `internal/template/templates/.claude/output-styles/moai/moai.md` (edited FIRST per Template-First)
- `.claude/output-styles/moai/moai.md` (LIVE — must end byte-identical to the template tree)
- `.moai/specs/SPEC-STEERING-ALIGN-OUTPUT-STYLE-SLIM-001/spec.md` (frontmatter status transition only)
- `.moai/specs/SPEC-STEERING-ALIGN-OUTPUT-STYLE-SLIM-001/progress.md` (§E.2/§E.3/§E.4 run/sync evidence)

Everything else is PRESERVE (do NOT touch):

- `SPEC-DIVECC-INVENTORY-VIEW-001/*` (parallel session's SPEC — explicit B10 PRESERVE)
- any other `.moai/specs/SPEC-*/` directory
- runtime-managed files (`.moai/harness/*`, `.moai/state/*`, `.moai/cache/*`, `.moai/logs/*`)
- parallel session research/audit artifacts (`.moai/research/*`, `.moai/reports/*`)
- the `.claude/rules/...` SSOT files themselves (M-POINTER points AT them, never edits them — §D)
- CLAUDE.local.md (FUTURE LOCAL-DIET P6 scope)

---

## B. Known Issues (filtered to relevant categories per Tier M)

- **B4. Frontmatter Canonical Schema** — `created:`/`updated:`/`tags:` (snake_case alias forbidden). spec.md uses the canonical 12 + `tier: M` + `era: V3R6`. Verified.
- **B6. spec-lint Heading regulation** — `## Out of Scope` (h2) alone → `MissingExclusions` ERROR; this SPEC's spec.md §D uses `### Out of Scope — <topic>` h3 sub-headings with `-` bullets (4 of them). Verified.
- **B8/B10. Working Tree Hygiene + Untouched Paths PRESERVE** — §A.5 PRESERVE list; `git add` specific paths only; the parallel SPEC-DIVECC-INVENTORY-VIEW-001 directory MUST NOT be touched.
- **B9. Git Commit + Push self (Hybrid Trunk 1-person OSS)** — main-direct per Tier M; Conventional Commits; `--no-verify` forbidden.
- **B11. AskUserQuestion 금지 (Subagent Boundary)** — manager-develop returns a structured blocker report; never prompts the user. The §8 Session Handoff WebFetch-style re-confirmation is NOT needed here (no upstream-doc claim to re-verify — the SSOT is internal `session-handoff.md`).

Not relevant (omitted): B1 cross-platform build tags (markdown-only diet, no syscall), B2 cross-SPEC retired-policy scan (no Go package touched), B3/B7 subagent-boundary/observer.go (no Go code), B5 CI 3-tier (no new Go test), B12 CHANGELOG (sync-phase, manager-docs).

---

## C. Pre-flight + the KEEP / CUT / POINTER classification table (CORE DELIVERABLE)

### C.1 Pre-flight baselines (run BEFORE any edit; record observed)

```bash
# 1. line + byte counts both trees (baseline 782 / 55306 each)
wc -l .claude/output-styles/moai/moai.md internal/template/templates/.claude/output-styles/moai/moai.md
wc -c .claude/output-styles/moai/moai.md internal/template/templates/.claude/output-styles/moai/moai.md

# 2. byte-parity (must be exit 0 at baseline AND after diet)
diff .claude/output-styles/moai/moai.md internal/template/templates/.claude/output-styles/moai/moai.md; echo $?

# 3. render-SSOT inventory (the protected set — MUST survive)
grep -cE '^### ' .claude/output-styles/moai/moai.md            # banner + sub-section count (capture baseline)
grep -c 'Header translation table\|Cut-line Marker translation' .claude/output-styles/moai/moai.md  # 8 banner tables + cut-line (capture baseline)

# 4. neutrality baseline (no NEW internal artifacts to be added by the diet)
go test ./internal/template/ -run TestTemplateNeutralityAudit 2>&1 | tail -3

# 5. output-styles count + parity guards (must pass at baseline AND after diet)
go test ./internal/template/ -run 'TestOutputStylesExactlyTwo|TestOutputStylesTemplateLiveParity' 2>&1 | tail -3
```

### C.2 SSOT verification for POINTER candidates — the prose-duplication bar (verification-claim-integrity.md C-7)

A passage is a POINTER candidate ONLY if its DISTINCTIVE content is tool-verified to live in an external SSOT. The "this is duplicated" claim is a defect claim → must be grep-verified BEFORE proposing the cut. Verified this plan-phase (commands + observed in spec.md §F.1):

| Candidate passage | External SSOT | Distinctive-content grep | Observed | Verdict |
|-------------------|---------------|--------------------------|----------|---------|
| §8 Session Handoff field-by-field / auto-memory procedure / anti-pattern prose (L657-728 narration around the tables) | `session-handoff.md` | `grep -c '6-block\|Block 1\|Block 6\|Field-by-Field'` | 23 | **POINTER** (prose owned externally) |
| §8 Session Handoff source_session_id + env-fallback narration | `session-handoff.md` | `grep -c 'source_session_id\|environment-fallback\|moai session current'` | 6 | **POINTER** (prose owned externally) |
| §8 Session Handoff effort-ultracode re-set narration | `session-handoff.md` + `dynamic-workflows.md` | `grep -c 'effort ultracode\|workflow fan-out'` | 3 | **POINTER** (prose owned externally) |
| §10 Output Rules AskUserQuestion + preview-field prose (L751-752) | `askuser-protocol.md` | `grep -c 'select:AskUserQuestion\|max 4 questions\|Free-form\|Preview Field'` | 18 | **POINTER candidate** (already 1-line + cross-ref; minimal/no gain — likely KEEP, run-phase confirms) |
| Epic Stats / Epic Status taxonomy EXPLANATION (NOT the banner skeleton/table) | `sprint-round-naming.md` | `grep -c 'Epic\|Milestone\|multi-SPEC\|within-SPEC'` | 50 | **POINTER candidate** (taxonomy prose only; banner skeleton + table stay render-SSOT) |

DEMOTE-to-KEEP (render-SSOT — NO external owner; 0 distinctive-content hits expected at the render level):

| Passage | Why KEEP | Render-SSOT class |
|---------|----------|-------------------|
| 14 banner skeletons (the fenced templates themselves) | orchestrator renders these at output time; no other file carries them | render-SSOT |
| 8 per-banner en/ko/ja/zh header-translation tables | the locale rendering SSOT read at output time | render-SSOT |
| §8 Localization Contract ko-canonical mapping tables (L253-274 label→ko; L299-318 banner-body-prose→ko) | the orchestrator reads the ko-canonical mapping at render time | render-SSOT |
| §8 Session Handoff cut-line marker table (L679) + header table (L686) + parity sentinel (L653) | bound by the drift-mitigation parity contract — 4-locale columns MUST stay at render surface (REQ-OSS-006) | render-SSOT + parity-contract |
| §9 Language Rules / §10 Output Rules directives + verbatim-preserve symbol list (L234-238) + `ultrathink.` token | distinct binding directives | KEEP (behavioral) |

### C.3 The per-section KEEP / CUT / POINTER classification table

Mechanism legend: **KEEP** = render-SSOT or behavioral directive (per-line test = YES, removing causes wrong render). **CUT** = M-DELETE (per-line-test FAIL; meaning carried by surviving skeleton + SSOT pointer). **POINTER** = M-POINTER (duplicated explanation collapsed to 1-line cross-ref; render skeleton preserved).

| moai.md section | Lines | Classification | Rationale |
|-----------------|-------|----------------|-----------|
| §1 Core Identity | L16-33 | KEEP | identity prose, short, behavioral |
| §2 Cannot-Do | L37-49 | KEEP | [HARD] hard-limit directives |
| §3 Four-Step State Machine | L52-99 | KEEP | the orchestrator's operating loop (behavioral) |
| §4 Delegation Decision | L101-144 | KEEP | delegation decision + Forced Delegation Table + Token-Cost Axis (behavioral). NOTE: §4 Forced Delegation Table contains archived-agent names (expert-backend etc.) — flagged as a forward cleanup candidate but OUT OF SCOPE here (this is a diet, not a neutrality/archived-agent sweep, §D) |
| §5 Checkpoint Verification Gate | L148-169 | KEEP | gate criteria (behavioral) |
| §6 Persistence & Context Awareness | L172-197 | KEEP (one POINTER sub-line) | persistence doctrine is behavioral; the Session-Boundary-Handoff 5-trigger detail at L189-196 already cross-refs `session-handoff.md` §When-To-Generate — minor; run-phase may tighten but low gain |
| §7 Temp File Hygiene | L200-208 | KEEP | cleanup checklist (behavioral) |
| **§8 Localization Contract [HARD]** | **L213-331** | **KEEP (render-SSOT) — minor CUT of duplicated framing prose only** | the ko-canonical mapping tables (L253-274, L299-318) + the verbatim-preserve list (L234-238) + the pre-emit self-check (L320-330) are render-SSOT. The "Root cause of the defect" / "Fallback rule" explanatory prose (L276-278) MAY be lightly CUT if its meaning is fully carried by the surviving tables — small gain; preservation WINS |
| §8 14 banner skeletons + 8 header tables | L332-647 | KEEP (render-SSOT) | the banner templates + en/ko/ja/zh tables; orchestrator reads at output time; NO external owner (REQ-OSS-003) |
| §8 Epic Stats / Epic Status taxonomy EXPLANATION | within L512-580 | POINTER (taxonomy prose only) | the Epic=multi-SPEC / Milestone=within-SPEC EXPLANATION is owned by `sprint-round-naming.md` → 1-line cross-ref. The banner skeleton + the header table STAY render-SSOT |
| **§8 Session Handoff [HARD]** | **L648-731 (~84L)** | **POINTER (the primary candidate)** | self-declares render-only (L652) + points at `session-handoff.md`. CONDENSE the duplicated field-by-field narration (L657-728) to the render skeleton + a 1-line pointer. PRESERVE: the cut-line marker table (L679), the header table (L686), the parity sentinel (L653), the fenced 6-block render skeleton (L657-675), the verbatim cut-line markers. This is where the bulk of the ~150-250L reduction comes from |
| §9 Language Rules [HARD] | L732-739 | KEEP | binding directives + verbatim-preserve symbol list (REQ-OSS-003) |
| §10 Output Rules [HARD] | L743-752 | KEEP (already-pointer prose) | the AskUserQuestion + preview-field bullets already cross-ref `askuser-protocol.md`; minimal further gain → KEEP |
| §11 Reference Links | L756-767 | KEEP | already a pointer list (no duplication) |
| §12 Service Philosophy | L771-782 | KEEP | short closing prose |

**Where the reduction comes from (honest attribution)**: predominantly the §8 Session Handoff block (L648-731, ~84L) condensed to ~15-20L render skeleton + pointer (≈ −65L), plus light CUT of the §8 Localization Contract framing prose (L276-278 + scattered duplicated narration) and the Epic Stats/Status taxonomy EXPLANATION pointer-ization (≈ −15-30L). Net estimate −150 to −250L. **NO banner skeleton, NO translation table, NO ko-canonical mapping table is cut.**

### C.4 Derived target (REQ-OSS-008)

Baseline 782L. Derived range: **[530, 630]** (≈ −150 to −250L), SOFT band. The floor 530 reflects an aggressive-but-honest §8 Session Handoff condense; the ceiling 630 reflects a conservative condense that preserves more narration. If the honest classification (after the AC-OSS-006 run-phase re-grep) lands ABOVE 630 because preservation forces fewer cuts, run-phase updates the derived range here with justification rather than over-cutting render-SSOT content (REQ-OSS-008 — behavioral-PASS over numeric-proxy). The AC-OSS-001 PASS predicate uses [530, 630] as the soft target with the explicit behavioral-PASS escape clause.

---

## D. Key Decisions

### D.1 Render-SSOT preservation is the load-bearing decision (the over-cut defense)

The §8 file is 66% render templates. The single most dangerous failure mode is cutting a render-SSOT passage that LOOKS duplicated but is read at output time (the P2 D1 over-cut, applied to output-style). DECISION: the §A.3 render-SSOT table is authoritative; the run-phase AC-OSS-006 re-grep gates EVERY POINTER edit (0 distinctive-SSOT-hits → block, force KEEP). Default on ambiguity = KEEP.

### D.2 §8 Session Handoff is the primary M-POINTER target (highest-confidence)

It ALREADY self-declares render-only (L652) and names its SSOT (`session-handoff.md`). The parity sentinel (L653) tells us EXACTLY what must stay (the 4-locale tables + sentinel) vs what can collapse (the field-by-field narration / auto-memory procedure / anti-pattern prose). This is the cleanest, highest-gain, lowest-risk cut. DECISION: condense L657-728 narration to the render skeleton + pointer; preserve L653/L679/L686 verbatim.

### D.3 M-SCOPE NOT used

No new paths-scoped rule is created from the output-style. The diet uses M-DELETE + M-POINTER only (spec.md §A.4). DECISION: M-SCOPE absent; do not claim it.

### D.4 Pointer-ization style (consistency with P2)

Use a single consistent pointer form: `> Canonical: see .claude/rules/moai/workflow/session-handoff.md § <Section> for the full <thing>.` Mirrors the P2 CLAUDEMD-DIET-001 pointer style. The render skeleton stays inline; only the EXPLANATION points out.

### D.5 §4 Forced Delegation Table archived-agent names — OUT OF SCOPE (forward note)

§4 (L113-127) lists archived agents (`expert-backend`, `manager-quality`, etc.) that were archived by the agent-catalog consolidation. This is a real staleness, but it is a NEUTRALITY/ARCHIVED-AGENT cleanup, NOT a diet. DECISION: OUT OF SCOPE (§D exclusion — this SPEC is a diet, not a sweep). Recorded as a forward candidate for a future SPEC; this SPEC does NOT touch §4 content.

### D.6 Neutrality is a guard, not a goal

Any pre-existing SPEC-ID *example* in §8 worked-examples (e.g. the L555/L540 `SPEC-<DOMAIN>-NNN` placeholders) MAY be neutralized as a side benefit IF a passage carrying it is being edited anyway, but neutralization is NOT a goal and MUST NOT expand the edit scope. DECISION: scope discipline — `TestTemplateNeutralityAudit` is the guard (no NEW leak); no proactive neutrality sweep.

---

## E. Milestones (priority-based, no time estimates)

- **M1 — Diet-candidate enumeration + per-passage classification lock**: from §C.3, lock the exact line ranges of each CUT and POINTER passage in the TEMPLATE tree. For every POINTER candidate, re-run the §C.2 distinctive-content grep against its SSOT and record ≥1 hit (AC-OSS-006 precondition). Any 0-hit candidate is reclassified KEEP. Output: the locked edit list (template-tree line ranges + per-edit SSOT-hit evidence).
- **M2 — §8 Session Handoff pointer-ization (template tree)**: condense L657-728 narration to the render skeleton + 1-line pointer. PRESERVE the cut-line marker table (L679), the header table (L686), the parity sentinel (L653), the fenced 6-block render skeleton, the verbatim cut-line markers (REQ-OSS-001 / REQ-OSS-006 / C-8). Apply to `internal/template/templates/.claude/output-styles/moai/moai.md` FIRST.
- **M3 — §8 Localization-Contract + Epic-taxonomy duplicate-prose deletion (template tree)**: light CUT of duplicated framing prose (§8 Localization Contract L276-278 + scattered narration) + POINTER the Epic Stats/Status taxonomy EXPLANATION to `sprint-round-naming.md`. PRESERVE the ko-canonical mapping tables, the banner skeletons, the header tables (REQ-OSS-003).
- **M4 — Dual-tree identical apply + byte-identical verification**: re-embed via `make build`; copy/apply the IDENTICAL diff to the LIVE tree; verify `diff .claude/output-styles/moai/moai.md internal/template/templates/.claude/output-styles/moai/moai.md` → exit 0 (REQ-OSS-005 / AC-OSS-002). Run the output-styles parity + count CI guards (`TestOutputStylesExactlyTwo`, `TestOutputStylesTemplateLiveParity`) + neutrality guard (`TestTemplateNeutralityAudit`) (AC-OSS-007).
- **M5 — Render self-check (all render-SSOT survives)**: verify all 14 banner skeletons present (AC-OSS-003), all 8 header-translation tables + ko-canonical mapping tables intact (AC-OSS-004), §9/§10 directives + verbatim-preserve symbol list + `ultrathink.` token survive (AC-OSS-008), line-count in [530, 630] soft band or behavioral-PASS justification (AC-OSS-001), byte-sum reduced (AC-OSS-005), every deleted-prose line has a verified external-SSOT home (AC-OSS-009). Record evidence in progress.md §E.2/§E.3.

### Milestone risk notes

- M2 is the highest-gain, highest-care milestone: the parity sentinel + 4-locale tables MUST survive the condense (C-8 / REQ-OSS-006). A careless collapse breaks the drift-mitigation parity contract the sentinel enforces.
- M1's AC-OSS-006 re-grep is the over-cut circuit-breaker: a POINTER candidate that drops to 0 SSOT-hits at run-time (SSOT changed between plan and run) MUST be reclassified KEEP and reported as plan-vs-run drift — NOT silently deleted.
- M4 byte-parity is the Template-First invariant; the LIVE tree must end byte-identical to the template tree.

---

## F. Anti-Patterns

- **AP-OSS-001 — Cutting a banner skeleton or translation table**: render-SSOT content (the orchestrator reads it at output time); removing it deletes behavioral render content. Forbidden (REQ-OSS-003 / AC-OSS-003/004).
- **AP-OSS-002 — Collapsing the §8 Session Handoff parity sentinel or 4-locale tables**: breaks the drift-mitigation parity contract the sentinel enforces (C-8 / REQ-OSS-006). Pointer-ization shrinks only the duplicated narration AROUND them.
- **AP-OSS-003 — Over-cutting render-SSOT to hit the ~530-630L band**: numeric-proxy over behavioral-PASS — the exact P2 D1 over-cut failure mode. Preservation WINS (REQ-OSS-008).
- **AP-OSS-004 — Proposing a POINTER without grep-verifying the SSOT owns the prose**: a "this is duplicated" claim is a defect claim (verification-claim-integrity.md §1.1); must be tool-verified (C-7 / AC-OSS-006).
- **AP-OSS-005 — Banner-template restructure**: the rejected "적극" option; OUT OF SCOPE (§D). MODERATE bound is a duplicate-prose diet, not a banner redesign.
- **AP-OSS-006 — Neutrality/archived-agent sweep scope-creep**: §4 Forced Delegation Table archived-agent staleness is a forward candidate, NOT this SPEC's scope (D.5). This SPEC is a diet, not a sweep.
- **AP-OSS-007 — Touching SPEC-DIVECC-INVENTORY-VIEW-001 or any other SPEC dir**: parallel-session PRESERVE (§A.5 / B10).

---

## G. Cross-References

- spec.md §A.3 (render-SSOT preservation table — the load-bearing invariant), §A.4 (M-DELETE / M-POINTER mechanisms), §B (REQ-OSS-001..009), §F.1 (verified ground-truth).
- acceptance.md (AC-OSS-001..009 with re-runnable commands).
- `.claude/rules/moai/workflow/session-handoff.md` — SSOT for the §8 Session Handoff render-only block (M2 pointer target).
- `.claude/rules/moai/core/askuser-protocol.md` — SSOT for AskUserQuestion / preview-field (§10 candidate).
- `.claude/rules/moai/development/sprint-round-naming.md` — SSOT for Epic/Milestone taxonomy (M3 Epic-taxonomy pointer target).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability (Tier M full 5-section delegation template required).
- SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001 plan.md (sibling P2, COMPLETED) — the KEEP/CUT/POINTER classification-table shape + the prose-duplication bar (§C.2) this plan mirrors.
- CLAUDE.local.md §2 (Template-First), §15 (neutrality), §25 (internal-content isolation).
