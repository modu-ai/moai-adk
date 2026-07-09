# Research — SPEC-TEMPLATE-I18N-CANONICAL-001

> **Provenance**: this research baseline is synthesized from the 4-lens independent audit workflow conducted 2026-07-10 (5 agents, 815,165 tokens, 26 tool calls, GLM-5.2 1M backend). The full JSON result is preserved at `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/046ce5f6-3c3b-4640-924e-e25f07e442e3/tasks/wfd51kjub.output`. Workflow transcript: `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/072fa7bb-4d0f-427b-8980-f3a32cb94d86/subagents/workflows/wf_ef98c41e-4a8/`. Memory topic: `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/project_i18n_canonical_sweep_audit_2026_07.md`.

---

## §A Audit methodology

**Trigger**: GOOS행님 question — "session-handoff.md에 한글이 있으면 다국어 사용자 문제 아닌가?" ("If session-handoff.md contains Korean, isn't that a multilingual-user problem?").

**Design**: 4 independent parallel lenses + 1 synthesis agent, each scoped to a single analytical dimension to avoid confirmation bias:

| # | Lens | Scope | Tokens |
|---|------|-------|--------|
| 1 | Non-Korean end-user | en/ja/zh + the other 13 supported languages reading this doctrine inside their own session | 170,307 |
| 2 | Doctrinal compliance | CLAUDE.local.md §15 (16-language neutrality) + §25 (Template Internal-Content Isolation) conformance | 201,124 |
| 3 | Localization-completeness | For every Korean instance, is there a working localization path? | 167,247 |
| 4 | Cross-file consistency | Is the Korean-canonical pattern isolated to session-handoff.md or systemic? | 161,301 |
| 5 | Synthesis | Reconcile the 4 lenses; resolve severity disagreements; produce the verdict | 115,186 |

**Total**: 815,165 tokens, 26 tool calls, all 5 agents completed (`state: done`). The synthesis reconciled severity inflation: lens 4 called the askuser policy section "critical"; lens 1 called the memory heading and goal-first example "major" — both reconciled downward in synthesis (askuser → major; memory heading/goal-first → minor) because nothing produces broken runtime output.

---

## §B Findings synthesis (0 critical / 3 major / 5 minor, deduplicated)

### §B.1 Major 1 — askuser-protocol.md policy section (localization-doctrine-bias)

**Surface**: `core/askuser-protocol.md` § "Recommendation Placement Principles" (lines ~127-175, ~47 lines, ~1,176 Hangul syllables total in the file).

An entire normative policy-doctrine section — 5 numbered HARD principles governing AskUserQuestion recommendation placement (emission timing via Fisher information `I=p(1−p)`, question ordering by information gain, statistical-majority default, precondition statement, adaptive strength) — is authored and shipped in Korean prose. This is NOT a localization table (no English column) and NOT an illustrative example. It is binding `[ZONE:Evolvable]` policy with mathematical rationale.

The template stripped the dev-version SPEC-ID provenance (`SPEC-V3R6-ASKUSER-DECISION-MEMORY-001`, neutrality-good per §25) but left the Korean prose intact. A non-Korean orchestrator session bound by these HARD rules cannot parse them from the file directly — it must infer behavior from the English cross-references.

**Fix**: translate the entire section to English-canonical prose matching the two clean sibling files (`agent-common-protocol.md`, `goal-directive.md`). Preserve `[ZONE:Evolvable]` tags, Fisher-information math, 5-principle structure, and cross-references verbatim in English. Also translate the 3 Korean worked examples (anti-pattern ~line 32, Epic-8 AskUserQuestion ~line 246, correct-pattern ~line 436). **Highest-leverage single fix in the audit.**

### §B.2 Major 2 — session-handoff.md Korean-canonical framing (localization-framing)

**Surface**: `workflow/session-handoff.md` canonical 6-block skeleton (lines ~38-53), Localization Table header (line ~69: `| Element | English | Korean (canonical) |`), Field-by-Field Spec (lines ~109-115).

The canonical 6-block skeleton — the single most-imitated artifact in the file — is rendered Korean-first (`전제 검증:` / `실행:` / `머지 후:` / `여기부터 복사`). The Localization Table explicitly labels its Korean column `"(canonical)"`, a deliberate elevation marker.

**Internal contradiction**: line ~82's own fallback rule sends the 12 non-en/ko/ja/zh locales to the **English** column for the structural skeleton, not Korean. So English is the de facto structural default while Korean is labeled canonical — the label and the fallback behavior are inconsistent.

**Mitigation (why this is framing-bias, not broken-output)**: line ~80's render-time substitution + the loc-table rescue the runtime paste-ready output. 25 of 27 Korean instances route through localization. The defect is documentation-reading / mental-model bias, not emitted-handoff correctness.

**Fix**: flip the canonical skeleton to English-first OR locale-neutral placeholders; drop `"(canonical)"` from the Korean column header; resolve the line ~67/82 tiering contradiction.

### §B.3 Major 3 — systemic authoring-discipline gap (systemic-authoring-discipline)

**Surface**: cross-file, the 5 always-loaded rule set.

| File | Hangul syllables | Canonical form |
|------|-----------------|----------------|
| `core/agent-common-protocol.md` | 0 | English-canonical ✓ |
| `workflow/goal-directive.md` | 0 | English-canonical ✓ |
| `core/askuser-protocol.md` | 1,176 | Korean-canonical ✗ |
| `workflow/session-handoff.md` | 186 | Korean-canonical ✗ |
| `workflow/context-window-management.md` | 18 | Korean-canonical ✗ |

2 of 5 files are cleanly English-canonical; 3 of 5 leak Korean-canonical content. The two clean files prove the project has an English-canonical baseline and authoring convention; the three leaking files violate it. **Fixing session-handoff.md alone leaves the strictly larger askuser surface untouched and creates a false sense of resolution.** This is why the sweep is Tier L (3+ files, systemic), not a per-file patch.

### §B.4 Minor findings (5)

1. **Memory heading leak** — `## 다음 세션 시작점 (paste-ready resume message)` (session-handoff.md lines ~225/229) has no Localization Table row and is not reached by the render-time substitution instruction. A non-Korean orchestrator following the rule literally writes a Korean heading into a non-Korean user's private memory topic file. Lenses disagreed (end-user lens: major; localization lens: minor); resolved to minor because memory files are internal state read primarily by Claude and the parenthetical English gloss softens the heading.

2. **Goal-first bootstrap example leak** — line ~163 `/goal SPEC-X run 재개: memory의 …` is entirely Korean with no Localization Table row and no "adapt to your locale" note. A non-Korean user copying it verbatim pastes a Korean `/goal` condition evaluated against their non-Korean transcript. Lowered to minor because it is explicitly "NOT the default" and labeled "Illustrative".

3. **context-window-management.md resume format** — lines ~66-71 render the canonical resume-message format entirely in Korean with no in-file localization table and no pointer to session-handoff.md's table. Smaller volume (18 syllables) but a cleaner instance of the Korean-canonical-default defect because zero in-file mitigation is present.

4. **Korean-first ordering** — Trigger #3 (line ~19) lists Korean example phrases first; Field-by-Field Spec (lines ~109-115) renders Korean as default with English parenthetical; line ~67 labels en/ko "primary locales". Functionally covered by the loc-table + line ~80 substitution; impact is cosmetic anchoring of a non-Korean orchestrator's mental model.

5. **CI enforcement gap** — `internal/template/template_neutrality_audit_test.go` classes C1/C2/C4/C5/C6/C8 + `internal_content_leak_test.go` classes C1–C8/S1–S3 have zero coverage for natural-language canonical-form bias. A future edit entrenching Korean (or swapping to privilege another locale) would pass CI green. This is an observation about enforcement coverage conditional on the doctrine being extended — not a defect in session-handoff.md itself.

---

## §C Doctrinal blind-spot analysis

### §C.1 The gap (letter-compliant, spirit-level concern)

`CLAUDE.local.md` §15 ("Template language neutrality") and `.moai/docs/template-internal-isolation-doctrine.md` §25 ("Template Internal-Content Isolation") together define the template-content governance:

- **§25.1** defines Allowed vs Forbidden content classes. The forbidden classes (C1–C8) cover *internal-development traces*: SPEC IDs, REQ/AC tokens, audit citations, internal dates, commit SHAs, macOS-bias paths, CLAUDE.local references, PR numbers, Go-impl paths, Go cross-compile env vars. None of these address *natural-language canonical-form choice*.
- **§15** governs 16 **programming** languages (go, python, typescript, …, flutter, swift) with explicit equal-treatment rules. It does not address natural languages.

So the Korean-canonical framing in session-handoff.md / askuser-protocol.md is **letter-compliant** — a file cannot violate a doctrine that does not address the relevant dimension. But it is a **spirit-level concern**: the template-genericity ethos that §25 opens with ("범용 자산 / generic assets deployed to external users") extends by analogy to natural languages. A tool shipped to 16-language users that labels one natural language "(canonical)" crosses from "happens to be Korean" into "Korean is declared canonical," which is harder to defend under the genericity spirit.

### §C.2 Two resolution paths

The doctrinal lens offered two paths; this SPEC takes a hybrid:

- **(a) Locale-neutral resolution** (taken by M1–M3): rewrite the canonical-framing surface to English-canonical or locale-neutral placeholders, with all Localization Table columns equal-tier. Aligns with the template-genericity spirit without requiring a doctrine change.
- **(b) Doctrinal-legitimation resolution** (proposed by M4): extend the doctrine file's §25 and/or `CLAUDE.local.md` §15 (line 611, "템플릿 언어 중립성") to explicitly govern natural-language neutrality. This legitimates the enforcement layer (the CI lint) and closes the blind spot so future regressions are detectable.

M4 is **conditional**: until the doctrine speaks to natural-language neutrality, adding a CI guard would put enforcement ahead of policy (the doctrinal lens's own caveat). M4 proposes the amendment + the lint together so policy and enforcement land simultaneously.

### §C.3 Why the runtime output is NOT broken (the design invariant)

The localization mechanism has three layers that together produce locale-correct runtime output:

1. **Localization Table** — en/ko inline columns in `session-handoff.md`; ja/zh extension columns in `session-handoff-examples.md` (verified present in the template tree).
2. **Render-time substitution** — line ~80: "Read `conversation_language` from `.moai/config/sections/language.yaml` at render time; substitute the localized text between the `✂────` decorators (cut-line markers preserved verbatim across all locales) and the locale rendering for each Block 1/3/5/6 placeholder."
3. **English-skeleton fallback** — line ~82: for ISO-639 codes not in the 4-locale table (fr, de, es, pt, vi, …), the structural skeleton falls back to the English column.

The audit verified (localization-completeness lens) that 25 of 27 Korean instances in session-handoff.md route through this mechanism. The 2 that don't (memory heading, goal-first example) are the minor-completeness gaps addressed by REQ-I18N-006/007. **This is why the design invariant (REQ-I18N-001/013) is load-bearing: the mechanism is sound and MUST be preserved while the framing is swept.**

---

## §D Baseline measurements (ripgrep-verified at plan baseline `39c74d777`)

Measured with `rg -oN '[가-힣]' <file> | wc -l` (ripgrep — preferred for portability across locale-handling differences and to avoid the windowed-grep undercount risk; see §E):

| File | rg Hangul count | Audit count | Match |
|------|----------------|-------------|-------|
| `core/askuser-protocol.md` | 1,176 | ~1,176 | ✓ |
| `workflow/session-handoff.md` | 186 | 186 | ✓ |
| `workflow/context-window-management.md` | 18 | 18 | ✓ |
| `core/agent-common-protocol.md` | 0 | 0 | ✓ |
| `workflow/goal-directive.md` | 0 | 0 | ✓ |

Spot-checks (specific Korean strings present in template files):
- `askuser-protocol.md`: 14 matches for `전제 검증|다음 세션|권장|Recommendation Placement`
- `session-handoff.md`: 10 matches for `여기부터|전제 검증|Korean (canonical)|다음 세션 시작점`
- Template ↔ local askuser-protocol.md line-count parity at the `39c74d777` baseline: 477 = 477 (byte-parity held at that SHA). NOTE: at current HEAD the two trees are INTENTIONALLY DIVERGED (24 diff lines) — the template tree is §25-neutral (0 `SPEC-V3R6-ASKUSER-DECISION-MEMORY-001` provenance refs) while the local tree retains 6 dev-trace provenance pointers (`AC-ADM-005..017`, Epic 7 TMC-001, §24 namespace). This is §25-intentional drift, not a sync defect; AC-I18N-019 is narrowed accordingly (sweep-induced changes sync; pre-existing dev-trace drift preserved).

Localization Table row count baseline (`session-handoff.md`, Mechanical-countable rows touching locale elements): **9** — this is the floor for REQ-I18N-001's "row count does not decrease" invariant. The sweep ADDS rows (memory heading per REQ-I18N-006), so post-sweep count ≥ 9 + 1 = 10.

---

## §E Methodological hazards (recorded for run-phase)

1. **Hangul-count tooling portability**. `rg -oN '[가-힣]'` (ripgrep) is the preferred tool for Hangul-range verification, for portability across locale-handling differences and to avoid the windowed-grep undercount risk (`sed -n <window> | grep` pipelines undercount multibyte ranges; memory `feedback_windowed_grep_undercount_authoring`). **Retraction note (v0.2.0)**: an earlier draft of this research claimed "BSD `grep -oE '[가-힣]'` returns 0 on macOS" — that does NOT reproduce on this Darwin 25.5.0 env, where `grep` is `ugrep 7.5.0` and both `rg` and `grep -oE` return the correct 1176 count (verified independently by orchestrator and plan-auditor). The original plan-draft 0-count was a broken-shell-fallback artifact (`ggrep ... | wc -l` masked a missing-binary failure because `wc -l` exits 0, so the `||` fallback never triggered), NOT a grep Unicode-range defect. The cited memory `feedback_windowed_grep_undercount_authoring` is about WINDOWED grep (`sed -n <window> | grep` undercount), not whole-file char-class grep — this SPEC over-generalized it. The AC commands (`rg`) are unchanged; testability is unimpaired.

2. **Severity reconciliation is non-obvious.** Individual lenses inflated severity (lens 4 called askuser "critical"; lens 1 called memory heading "major"). The synthesis reconciled these downward. Run-phase implementers reading this research MUST cite the synthesis verdicts (§B), not individual lens findings, when prioritizing — the lens-level severities were honestly-disaggregated and resolved.

3. **The audit read the TEMPLATE files, not local copies.** All audit paths are `internal/template/templates/.claude/rules/moai/...` (the Template-First source-of-truth). This SPEC's scope is the template files; local sync is a run-phase consequence, not a separate audit target.

4. **`session-handoff-examples.md` carries the ja/zh extension columns.** The inline table in `session-handoff.md` carries only en/ko; the full 4-locale table (en/ko/ja/zh) lives in `session-handoff-examples.md § Localization Table (Full 4-Locale)`. The design invariant covers BOTH files — the sweep MUST verify ja/zh columns survive in the examples file even when editing the inline table.

---

## §F Cross-references

- **Audit JSON (full)**: `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/046ce5f6-3c3b-4640-924e-e25f07e442e3/tasks/wfd51kjub.output`
- **Audit memory topic**: `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/project_i18n_canonical_sweep_audit_2026_07.md`
- **Doctrine**: `.moai/docs/template-internal-isolation-doctrine.md` §25 (Template Internal-Content Isolation) + `CLAUDE.local.md` §15 line 611 (템플릿 언어 중립성)
- **Clean baseline files** (the English-canonical target style): `core/agent-common-protocol.md`, `workflow/goal-directive.md`
- **Lineage**: `SPEC-TEMPLATE-RULES-CLEANUP-001` (Tier L template-rules precedent), `SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001` (neutrality-audit precedent)
- **Related feedback**: `feedback_windowed_grep_undercount_authoring` (WINDOWED grep undercount — `sed -n <window> | grep`; NOT a whole-file char-class grep defect, see §E.1 retraction), `feedback_hypothesis_as_defect` (audit→SPEC verification discipline), `feedback_defect_claim_verification` (tool-based verification)
