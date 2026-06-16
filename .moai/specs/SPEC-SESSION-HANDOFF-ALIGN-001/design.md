# Design — SPEC-SESSION-HANDOFF-ALIGN-001

> Design decisions for the mixed-content split, token stripping, i18n placeholder strategy,
> and anti-pattern catalogue consolidation. Input: research.md §B (content classification).

## §A. Mixed-content split strategy

The core design problem: the Diet and V0 sections are MIXED content — genuinely generic user-relevant doctrine interleaved with moai-adk-internal dev-incident provenance. Porting them naively to the template leaks internal tokens (§25 violation); porting only the generic parts requires a principled split.

### §A.1 Split principle

**Ship generic doctrine to template; retain dev-incident provenance in lesson memory.**

- "Generic doctrine" = any sentence that is true and useful for ANY MoAI user project, independent of moai-adk's internal SPEC iteration history.
- "Dev-incident provenance" = any sentence whose truth content depends on knowing a specific internal SPEC line (LIFECYCLE-SYNC-GATE-001, HARNESS-NAMESPACE, etc.) or a specific internal incident (Hugo docs server false-abort, claude-md-guide.md filename match).

The generic doctrine earns its place in the always-loaded rule file because it prevents real user-facing regressions (paste-ready bloat, false aborts). The dev-incident provenance earns its place in lesson memory because it is the empirical evidence chain that justifies the doctrine — important for audit but not for daily rule application.

### §A.2 Split decision table

| Section | Generic doctrine (→ template) | Dev-incident provenance (→ lesson memory / strip) |
|---------|-------------------------------|---------------------------------------------------|
| Diet opening L129 | "paste-ready = minimum executable context, not audit trail" + "차수 누적 retry → append-only 누적은 empirical 입증된 anti-pattern" | "(LIFECYCLE-SYNC-GATE-001 line C 1~14차 + HARNESS-NAMESPACE Phase 1B line B 1~5차에서 동일 비대화 패턴 관측)" |
| Diet scope L183 | "Cross-line 일관 적용 (모든 SPEC line)" | "(LIFECYCLE-SYNC-GATE / HARNESS-NAMESPACE / SESSION-AUTO-RESUME 등 모든 SPEC line)" |
| V0 opening L187 | "lsof + cwd 교차 검증 사용; ps aux raw count는 baseline noise" | "(cross-line empirical 입증)" → reword to "empirical 입증" |
| V0-b inline comment L195-199 | "grep -iE 'claude' 단독은 파일명 매칭 false-positive; COMMAND 컬럼 필터 필수" | "(Hugo docs 서버 PID 1개 → 8 entry 오탐, cross-line 입증)" |
| V0 Cross-pollination L213-219 | (none — entire block is iteration history) | ENTIRE BLOCK → lesson memory; rule body gets 1-line pointer |
| V0 AP-V-004 L226 | "lsof +D | grep -iE 'claude'는 filename false-positive; lsof -a -c claude 필수" | "(claude-md-guide.md·claude-design-handoff.md 등)... LIFECYCLE-SYNC-GATE-001 M4 1·2차" |

### §A.3 Layer assignment

- **Template tree** (`internal/template/templates/.../session-handoff.md`): generic doctrine only. Enforced neutral by CI.
- **Local canonical** (`.claude/rules/moai/workflow/session-handoff.md`): byte-identical to template post-M4 (no local-only content remains).
- **Lesson memory** (`~/.claude/projects/{hash}/memory/lessons.md` or a dedicated `L_v0_abort_gate_doctrine_history.md`): the Cross-pollination iteration history. The rule body references this via a 1-line pointer per REQ-SHA-014.

This is a 3-layer split: rule body (generic) ↔ lesson memory (provenance) ↔ CI guard (neutrality invariant).

## §B. Token-stripping enumeration (operational)

The exact strings to strip during M2 (Diet) and M3 (V0) porting. See research.md §B.3 for the authoritative list.

### §B.1 Strip rules

1. **SPEC-IDs**: drop the entire parenthetical or trailing clause that names an internal SPEC. Reword to generic phrasing.
2. **Iteration counts** (`N차`, `line C 9차`): drop entirely. The generic doctrine does not need to cite which iteration surfaced it.
3. **Internal filenames** (`claude-md-guide.md`, `claude-design-handoff.md`): drop. Replace with the generic shape ("파일명에 'claude' 포함된 콘텐츠").
4. **Internal service names** (`Hugo docs 서버`): drop. Replace with the generic shape.
5. **"cross-line 입증" / "cross-line empirical 입증"**: reword to "empirical 입증된" — drop the cross-line chain but preserve the empirical claim (the claim is generic; the chain is internal).

### §B.2 Verification

Post-port, run the AC-SHA-002 + AC-SHA-004 grep (research.md §B.3). Expected: zero matches on the ported blocks. If any match remains, the strip missed a token — halt, fix, re-run.

### §B.3 What NOT to strip

- The `ps aux` / `lsof` command literals — these are canonical technical content, not internal tokens.
- The `STRICT 0` / `STRICT ≤2` threshold annotations — generic.
- The AP-D-001..005 and AP-V-001..004 identifiers — these are catalogue IDs, not SPEC-IDs. They are generic and ship verbatim. (Distinction: `AP-D-001` is an anti-pattern label internal to this rule file; `SPEC-V3R6-*-001` is an internal SPEC identifier that leaks moai-adk dev state.)
- The `[ZONE:Evolvable] [HARD]` zone markers — generic.

## §C. i18n placeholder strategy

### §C.1 Skeleton verb reconciliation

**Current state (contradiction — content present in BOTH trees)**:
- session-handoff.md LOCAL L32/L85/L183: `ultrathink. <SPEC-ID> <phase> 진입.` (Korean verb baked into skeleton)
- session-handoff.md TEMPLATE L17/L32/L85/L183: **identical** `ultrathink. <SPEC-ID> <phase> 진입.` — the Korean skeleton verb is present in BOTH trees (verified by iter-2 Probe E: `grep -n '진입'` returns the same line set on both files). This is NOT a local-only drift; it is a content-internal-i18n-debt present identically in both copies.
- moai.md §8 L647: `ultrathink. <SPEC-ID or Sprint N> <phase> entering.` (English verb)

Both files' default `conversation_language` is `en`. The Korean-locked skeleton verb contradicts the English-default config and the moai.md §8 English rendering — and the contradiction is present in BOTH trees, so the fix is a content change applied to both trees per the Template-First Rule (template first, mirror to local in the same commit), NOT a local→template realignment. A run-phase agent must NOT scope this edit to local-only; doing so would re-introduce the exact LOCAL↔TEMPLATE drift this SPEC exists to close.

**Resolution (REQ-SHA-011)**: replace the literal verb with a placeholder `<entering verb>` in the canonical 6-block skeleton, in BOTH trees in the same M5 commit. The verb translates per the header table (Block 1 entering-verb row):

| Locale | Verb form |
|--------|-----------|
| en | `entering` |
| ko | `진입` |
| ja | `開始` (or 開始.) |
| zh | `进入` |

**Example blocks** (L82-98, L277-296): retain a concrete rendering (Korean `진입` is acceptable in the illustrative Example IF clearly marked as a ko-rendering, matching moai.md §8 discipline which uses placeholders in the skeleton but concrete locale renderings in examples). The cleaner option per Finding #14 of the analysis report: convert example headers to locale-neutral placeholders too, OR add English-rendered sibling examples. Decision for M5: use `<entering verb>` in the canonical skeleton; retain concrete `진입` in the Korean-default Example but add a sibling note "en: entering, ja: 開始, zh: 进入" so non-ko emitters have a rendering to imitate.

### §C.2 Header translation table replication

The 4-locale header table currently lives only in `moai.md §8`. REQ-SHA-010 replicates the Block 1 / Block 3 / Block 5 / Block 6 header rows into session-handoff.md's Localization Table section.

**Why replicate rather than cross-reference**: session-handoff.md is the DECLARED SSOT (moai.md §8 explicitly defers to it). An SSOT that delegates its own load-bearing locale data downstream is incoherent. Replicating the 4-locale table into the SSOT is the correct layering; moai.md §8 becomes the render surface that consumes the SSOT.

**Consistency requirement**: the replicated table MUST be consistent with moai.md §8's renderings (AC-SHA-010 cross-verifies). Byte-identity is NOT required (moai.md §8 may carry additional render-surface-specific columns); content consistency IS required.

### §C.3 Trigger #1 model-label drift elimination

**Current state (drift vector — content present in BOTH trees)**:
- session-handoff.md LOCAL L17: "1M context model (Opus 4.7): 50%"
- session-handoff.md TEMPLATE L17: **identical** "1M context model (Opus 4.7): 50%" — the "Opus 4.7" label is present in BOTH trees (verified by iter-2 Probe E: `grep -n 'Opus 4\.7'` returns L17 on both files). This is NOT a local-only drift; the label is a content-internal-i18n-debt present identically in both copies.
- context-window-management.md L15 (SSOT): "Opus 4.8 (1M) | 1,000,000 tokens | 50%"

Same threshold, different model label. The label is non-load-bearing (the threshold is the operational quantity), but the label drift creates confusion about which file is authoritative. The drift is present in BOTH session-handoff.md copies (LOCAL and TEMPLATE) relative to the context-window-management.md SSOT — so the fix is a content change applied to both session-handoff.md trees per the Template-First Rule, NOT a local→template realignment. A run-phase agent must NOT scope this edit to local-only.

**Resolution (REQ-SHA-012)**: replace the inline model-class numbers in L17 with a pointer to `context-window-management.md § Context Window Targets`, in BOTH session-handoff.md trees in the same M5 commit. The Trigger #1 row becomes: "Context usage crosses model-specific threshold — see `.claude/rules/moai/workflow/context-window-management.md § Context Window Targets` for the per-model-class threshold table." This kills the drift vector at the source — the label now lives in exactly one place.

## §D. Anti-pattern catalogue consolidation approach

### §D.1 Current state (3 disjoint catalogues)

1. **General prose** (L115-125): 9 bullets, NO IDs. Resume-hygiene patterns (free-form prose, no preconditions, no ultrathink, no auto-memory, duplicate entries, no source_session_id, trivial-task forcing, no cut-line markers, translated ✂/─).
2. **AP-D catalogue** (L160-166, inside Diet): 5 entries, AP-D-001..005. Paste-ready budget violations (lessons bloat, precondition prose, sub-step nesting, directive escalation, ceremonial reminder).
3. **AP-V catalogue** (L221-226, inside V0): 4 entries, AP-V-001..004. V0 abort-gate violations (ps aux raw-count, guilt-trip tracking, override option, filename grep false-positive).

**Problem**: the three catalogues overlap (AP-D-002 precondition-prose ≈ general L118 "no preconditions"; AP-D-001 lessons-bloat ≈ general L117 implicit), use different ID schemes (none / AP-D / AP-V), and do not reference each other. A reader encountering one cannot discover the others.

### §D.2 Consolidation decision (REQ-SHA-013)

**Option A — Merge into one catalogue with sub-groups**: single `## Anti-Pattern Catalogue` section with three sub-groups (### General / ### Diet (paste-ready budgets) / ### V0 (abort gate)). Pro: single discovery surface. Con: loses the contextual placement (Diet patterns are most useful next to the Diet budgets; V0 patterns are most useful next to the V0 commands).

**Option B — Cross-link pointers, retain contextual placement**: keep the three catalogues in their current contextual positions (general at L115, AP-D inside Diet, AP-V inside V0) but add a 1-line pointer at the top of each referencing the other two. Pro: preserves contextual utility. Con: still 3 surfaces.

**Decision for M5**: **Option B** (cross-link pointers). Rationale: the Diet and V0 catalogues earn their contextual placement — a reader applying Diet budgets benefits from seeing AP-D-001..005 immediately adjacent, not 100 lines away in a merged catalogue. The cross-link pointers solve discoverability without sacrificing contextual utility. This matches the analysis report's Finding #7 recommendation ("At minimum the three lists must reference each other for discoverability").

**Pointer template** (added to each catalogue):
```
> See also: § Anti-Patterns (general resume hygiene), § Diet Constraints / Anti-pattern catalogue (paste-ready budgets), § V0 Abort Gate Doctrine / Anti-pattern (abort-gate violations).
```

### §D.3 moai.md §8 forward-link (REQ-SHA-015)

Add to the Cross-references section of both copies:
```
- `.claude/output-styles/moai/moai.md §8 (Response Templates → Session Handoff)` — rendered 6-block template + pre-emit self-check; this file is the SSOT, moai.md §8 is the canonical render surface.
```

This closes the currently one-sided bidirectional link (moai.md §8 declares session-handoff.md the SSOT; session-handoff.md now acknowledges moai.md §8 as the render surface).

## §E. Cut-line marker de-duplication approach (REQ-SHA-009)

### §E.1 Current state (4 re-spellings)

The cut-line marker literal (`✂──── 여기부터 복사 ────✂` / `✂──── 여기까지 복사 ────✂`) and the "translate text, preserve ✂/─" rule appear in 4 locations:
1. L27 — § Canonical Format intro (re-spells both markers + the rule)
2. L47-54 — § Cut-line Marker Specification (THE SSOT — full spec lives here)
3. L113 — § Output Surface (re-spells both markers + the rule)
4. L124-125 — § Anti-Patterns (re-spells the rule negatively twice)

### §E.2 De-duplication decision

Keep the full literal spec ONLY in § Cut-line Marker Specification (L47-54). In L27, L113, L124-125, replace the re-spelled literals with pointers:

- **L27** (Canonical Format intro): "Resume message MUST follow this exact 6-block structure, bounded by cut-line markers (see § Cut-line Marker Specification below for the literal marker format and translation rules)."
- **L113** (Output Surface): "...in a fenced ```text``` block bounded by cut-line markers (per § Cut-line Marker Specification)..."
- **L124-125** (Anti-Patterns): "Cut-line markers absent — user cannot identify exact copy boundary" and "Cut-line markers translated contrary to § Cut-line Marker Specification — only the marker text translates; the ✂/─ symbols are preserved verbatim."

### §E.3 Preservation guarantees

- The fenced Example blocks (L29-45, L82-98, L277-296) RETAIN the marker literals verbatim — they are illustrative renderings, not re-spellings of the spec. AC-SHA-009 explicitly scopes the de-duplication to "outside fenced example blocks".
- The `✂` (U+2702) and `─` (U+2500) Unicode characters are load-bearing. Any byte drift in these characters fails the mirror test post-enrollment. The de-duplication touches prose pointers only, never the characters themselves.

## §F. Reader-flow reorganization approach (REQ-SHA-016)

### §F.1 Current state (local)

Heading order in the local file:
1. Loading scope note (L5)
2. Why This Matters (L7-9)
3. When To Generate (L11-23)
4. Canonical Format (L25-98)
5. Auto-Memory Integration (L100-109)
6. Output Surface (L111-113)
7. Anti-Patterns (L115-125)
8. **Diet Constraints (L127-183) — mid-file doctrine insertion**
9. **V0 Abort Gate Doctrine (L185-226) — mid-file doctrine insertion**
10. Worktree-Anchored Resume Pattern (L228-296)
11. Cross-references (L298-306)

The Diet + V0 sections sit BETWEEN Anti-Patterns and Worktree-Anchored, breaking the canonical reader path (WHEN → FORMAT → memory → output → anti-patterns → worktree → cross-refs).

### §F.2 Target state (local, post-M4)

1. Loading scope note
2. Why This Matters
3. When To Generate
4. Canonical Format
5. Auto-Memory Integration
6. Output Surface
7. Anti-Patterns
8. Worktree-Anchored Resume Pattern
9. **Diet Constraints (moved here)**
10. **V0 Abort Gate Doctrine (moved here)**
11. Cross-references

This matches the template's natural order (the template, lacking Diet/V0 pre-port, already has Anti-Patterns → Worktree-Anchored → Cross-references contiguous). After M2/M3 port + M4 restructure, both trees converge on this order.

### §F.3 Move mechanics

The move is byte-preserving on section content — only heading order changes. Both trees move in lockstep (M4 lands the restructure in both copies in the same commit). The mirror test remains green throughout because the section bodies are unchanged; only their relative position in the file differs, and both trees differ identically.

## §G. Alternative approaches considered and rejected

- **Alt-1 (rejected)**: Mark Diet+V0 "local-only, not in template" and keep the file out of the mirror set (Finding #1 option (b) from the analysis report). Rejected because it abandons the generic doctrine to local-only status — users never receive the Diet budgets or V0 commands. The port-and-neutralize approach (this SPEC) is strictly better for user outcomes.
- **Alt-2 (rejected)**: Extract Diet+V0 to dedicated rule files (`paste-ready-diet.md`, `v0-abort-gate.md`) and cross-reference from session-handoff.md. Rejected for Diet (it IS paste-ready doctrine, co-location is correct) and considered for V0 (Finding #4.1 notes V0 is thematically orphaned). Decision: retain V0 co-located because Block 4 emits the V0 precondition; extraction is a possible follow-up (FL-3) if maintenance burden grows.
- **Alt-3 (rejected)**: Merge the 3 anti-pattern catalogues into one (Option A in §D.2). Rejected because contextual placement of AP-D and AP-V next to their respective budgets/commands is more useful than single-surface discovery. Cross-link pointers (Option B) solve discoverability without sacrificing context.
- **Alt-4 (rejected)**: Add session-handoff.md to the §25 carve-out (L50-63 of rule_template_mirror_test.go) instead of achieving true byte-parity. Rejected (EXCL-004) — the carve-out is for files that CANNOT be neutralized; session-handoff.md CAN be neutralized and MUST achieve true byte-parity.

## §H. Risks carried forward

See spec.md §G. The design-level risks:
- The §B token-strip enumeration may miss a token the CI leak test catches — mitigation: the post-port grep (research.md §B.3) is the operational verification.
- The Option-B cross-link approach (§D.2) may be judged insufficient by a future auditor who prefers Option-A merge — mitigation: the decision is reversible (cross-links → merge is a smaller edit than merge → cross-links); defer the re-decision to a follow-up SPEC if the maintainer judges Option-B insufficient in practice.
