---
id: SPEC-TEMPLATE-I18N-CANONICAL-001
title: "Multilingual Canonical-Form Sweep for Always-Loaded Template Rule Files"
version: "0.2.0"
status: completed
created: 2026-07-10
updated: 2026-07-10
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/template/templates/.claude/rules/moai"
lifecycle: spec-anchored
tags: "i18n, multilingual, canonical-form, template, neutrality, english-canonical, localization, always-loaded"
era: V3R6
tier: L
related_specs: [SPEC-TEMPLATE-RULES-CLEANUP-001, SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001]
---

# SPEC-TEMPLATE-I18N-CANONICAL-001 — Multilingual Canonical-Form Sweep for Always-Loaded Template Rule Files

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-10 | manager-spec | Initial plan-phase authoring (Tier L). Synthesized from the 4-lens independent audit (5 agents, 815K tokens, 26 tool calls) of Korean-canonical content across the 5 always-loaded template rule files. 13 REQ (incl. REQ-I18N-013 unwanted-behavior) / 23 AC IDs (27 sub-checks). |
| 0.2.0 | 2026-07-10 | manager-spec | Plan-audit fix (PASS-WITH-DEBT 0.89, 6 defects): D1 retract BSD-grep false-negative claim (does NOT reproduce on this Darwin 25.5.0 env — `grep` is ugrep 7.5.0; both `rg` and `grep -oE` return 1176; the earlier 0-count was a broken `ggrep\|wc -l` fallback masking a missing binary, NOT a Unicode-range defect) → downgrade to portability + windowed-grep rationale; D2 re-baseline askuser template↔local §25-intentional drift (local retains `SPEC-V3R6-ASKUSER-DECISION-MEMORY-001` provenance, template neutral) + narrow AC-I18N-019 to sweep-induced-only; D3 fix REQ/AC headline counts (13 REQ / 23 AC IDs / 27 sub-checks: 20 mech / 4 hybrid / 3 manual); D4 fix §15 doctrine path (`CLAUDE.local.md` line 611, not the doctrine file which has §25 only); D5 add design.md (Tier L 5-artifact contract per spec-workflow.md:143); D6 name concrete race source in R6 + add askuser-drift R8. |

---

## §A Purpose & Context

### §A.1 The problem (audit-verified, not inferred)

A 4-lens independent audit (non-Korean end-user / doctrinal compliance / localization-completeness / cross-file consistency + synthesis) examined the five always-loaded rule files under `internal/template/templates/.claude/rules/moai/` that ship to every user via `moai init` and are auto-loaded into every Claude session. The audit found:

- **0 critical** — runtime paste-ready resume output is locale-correct for non-Korean users. The Localization Table (en/ko inline + ja/zh extension in `session-handoff-examples.md`) plus the render-time substitution instruction (`session-handoff.md` line ~80, "substitute the localized text … while keeping `✂` and `─` characters verbatim, per `conversation_language`") genuinely govern the skeleton's cut-line markers and Block 1/3/5/6 headers. 25 of 27 Korean instances in `session-handoff.md` route through localization. No non-Korean user receives mechanically wrong user-visible output today.
- **3 major / 5 minor** — biased Korean-canonical framing + 2 literal Korean leaks into user-private state + one unreadable Korean policy section + systemic authoring inconsistency + a CI enforcement gap. These are genuine quality debt for a tool shipped to 16-language users, not a broken-output crisis.

The headline finding: the file the user asked about (`session-handoff.md`, 186 Hangul syllables) is **not** the worst offender. `askuser-protocol.md` carries **1,176 Hangul syllables** including a ~50-line normative policy section ("Recommendation Placement Principles" — Fisher-information emission timing, 5 HARD principles) authored entirely in Korean prose that a non-Korean orchestrator session cannot parse from the file.

### §A.2 Baseline measurements (verified by ripgrep at plan baseline `39c74d777`)

The baseline below was measured with `rg -oN '[가-힣]'` (ripgrep) — preferred for Hangul-range verification for portability across locale-handling differences and to avoid the windowed-grep undercount risk (`sed -n <window> | grep` pipelines undercount multibyte ranges; memory: `feedback_windowed_grep_undercount_authoring`). The counts match the audit's figures exactly:

| File | Hangul syllables | Canonical form | Role |
|------|-----------------|----------------|------|
| `core/askuser-protocol.md` | 1,176 | Korean-canonical (policy section + examples) | **Worst offender** — sweep target M1 |
| `workflow/session-handoff.md` | 186 | Korean-canonical (skeleton + loc-table label) | Sweep target M2 |
| `workflow/context-window-management.md` | 18 | Korean-canonical (resume format example) | Sweep target M3 |
| `core/agent-common-protocol.md` | 0 | English-canonical | **Clean baseline** — preserve (M0) |
| `workflow/goal-directive.md` | 0 | English-canonical | **Clean baseline** — preserve (M0) |

The two clean files prove the project already has an English-canonical authoring convention for always-loaded rule files. The three leaking files violate that convention.

### §A.3 The design invariant (CRITICAL — load-bearing)

The localization **mechanism** is SOUND and MUST be preserved end-to-end:

1. The 4-locale Localization Table (`session-handoff.md` § Localization Table: en/ko inline columns; ja/zh extension columns live in `session-handoff-examples.md` per the inline table's own pointer).
2. The render-time substitution instruction: the orchestrator reads `conversation_language` from `.moai/config/sections/language.yaml` at render time and substitutes the localized text between the `✂────` decorators (cut-line markers preserved verbatim across all locales) and the locale rendering for each Block 1/3/5/6 placeholder.
3. The English-skeleton fallback rule: for ISO-639 codes not in the 4-locale table (fr, de, es, pt, vi, …), the structural skeleton falls back to the **English** column.

This SPEC is NOT "remove Korean." Korean remains an equal-tier locale column in every Localization Table. What changes is the **canonical framing**: the declared "canonical" default and the Korean-first default skeleton move to English (matching the clean baseline files + the English-fallback behavior), and the `"(canonical)"` privilege annotation is removed from the Korean column header.

### §A.4 Doctrinal basis

- `CLAUDE.local.md` §15 — Template language neutrality (16-language equal treatment).
- `.moai/docs/template-internal-isolation-doctrine.md` §25 — Template Internal-Content Isolation (Allowed vs Forbidden content classes C1–C8).

**Doctrinal blind spot (documented in `research.md` §C):** §25's forbidden classes (SPEC IDs, REQ/AC tokens, audit citations, dates, commit SHAs, macOS paths, etc.) cover *internal-development traces* but are silent on *natural-language canonical-form choice*. §15 governs 16 **programming** languages, not natural languages. So the Korean-canonical framing is letter-compliant (it cannot violate a doctrine that does not address the dimension) but is a spirit-level concern under the template-genericity ethos. This SPEC's M4 milestone proposes a doctrine amendment closing that gap, and a CI lint class enforcing it.

### §A.5 Scope summary

Three always-loaded template rule files are swept (priority order: askuser → session-handoff → context-window). The two clean baseline files are preserved untouched. An optional CI lint + doctrine amendment (M4) prevents regression. All edits follow the Template-First rule (source-of-truth is `internal/template/templates/`; local `.claude/` syncs byte-identical at run-phase).

---

## §B Requirements (GEARS)

### §B.1 Design invariant — localization mechanism preservation

- **REQ-I18N-001** (Ubiquitous, design invariant): The sweep shall preserve the localization mechanism end-to-end — the 4-locale Localization Table (en/ko inline + ja/zh extension in `session-handoff-examples.md`), the render-time substitution instruction keyed on `conversation_language`, and the English-skeleton fallback rule for non-table ISO-639 codes — such that a ko-locale user still receives Korean cut-line markers and Block headers, an en-locale user receives English, and the Localization Table row count does not decrease from the plan-baseline count.
- **REQ-I18N-013** (Unwanted behavior): The sweep shall not remove the Korean column from any Localization Table, shall not delete any existing Localization Table row, and shall not reduce the 4-locale coverage (en/ko/ja/zh) of any element present at the plan baseline — Korean remains an equal-tier locale column; only the "canonical" declaration and the Korean-first default skeleton change.

### §B.2 askuser-protocol.md — policy section + worked examples (M1, largest surface)

- **REQ-I18N-002** (Ubiquitous): The `askuser-protocol.md` § "Recommendation Placement Principles" section (~50 lines, 5 numbered HARD principles governing AskUserQuestion recommendation placement — emission timing via Fisher information `I=p(1−p)`, question ordering by information gain, statistical-majority default, precondition statement, adaptive strength) shall be translated to English-canonical prose, preserving verbatim in English: the `[ZONE:Evolvable]` tags, the Fisher-information mathematics and notation, the 5-principle numbered structure, and the cross-references (`verification-claim-integrity.md §1.1 surface 3`, `design.md §A.4` / `§B.2`).
- **REQ-I18N-003** (Ubiquitous): The three Korean worked examples in `askuser-protocol.md` — the anti-pattern block (~line 32), the Epic-8 `AskUserQuestion` worked example (~line 246), and the correct-pattern example (~line 436) — shall be translated to English-canonical payloads. The `(권장)` / `(Recommended)` token discussion stays intact (it is locale-aware by design and the Localization Table carries its own row).

### §B.3 session-handoff.md — skeleton, loc-table labels, 2 leaks (M2)

- **REQ-I18N-004** (Ubiquitous): The canonical 6-block skeleton in `session-handoff.md` § "Canonical Format (Verbatim Spec)" shall be rendered English-first (`Preconditions:` / `Run:` / `After merge:` / `Follow-up:`) OR in locale-neutral placeholders (`<Block 3 header per Localization Table>`) — NOT Korean-first. The Korean renderings (`전제 검증:` / `실행:` / `머지 후:` / `후속:` / `여기부터 복사` / `여기까지 복사`) shall appear ONLY inside the Localization Table's Korean column, never as the default skeleton.
- **REQ-I18N-005** (Ubiquitous): The `"(canonical)"` annotation shall be removed from the Localization Table Korean column header (column becomes `Korean`, equal-tier with `English`); the phrase `"primary locales"` (line ~67) shall be replaced by `"inline locales"` or equivalent neutral tiering language; and the internal contradiction between the tiering prose and the English-fallback rule (line ~82) shall be resolved so the declared structural default matches the fallback behavior.
- **REQ-I18N-006** (Ubiquitous, localization-completeness gap): The memory-section heading `## 다음 세션 시작점 (paste-ready resume message)` (lines ~225/229) shall gain a row in the Localization Table (en `## Next Session Entry Point` / ko `## 다음 세션 시작점`, with ja/zh in `session-handoff-examples.md`), and the render-time substitution instruction shall be extended to name the memory heading as a substitutable element; the close-time-pruning cross-reference (line ~229) shall match by content ("the next-session-start-point section") rather than by the Korean literal.
- **REQ-I18N-007** (Ubiquitous, localization-completeness gap): The goal-first bootstrap illustrative condition (line ~163, currently `/goal SPEC-X run 재개: memory의 …`) shall be rewritten in English-canonical prose matching the file's documented English-skeleton fallback, with a one-line note that the condition text follows the user's `conversation_language`.

### §B.4 context-window-management.md — resume format example (M3)

- **REQ-I18N-008** (Ubiquitous): The canonical resume-message format example in `context-window-management.md` § "Orchestrator Responsibilities" (lines ~66-71, currently `ultrathink. Epic <N> 이어서 진행 … 다음 단계: … 완료 후: …`) shall be rendered English-canonical, and the file shall cross-reference `session-handoff.md § Localization Table` for locale renderings rather than redefining a parallel Korean-canonical format.

### §B.5 CI lint + doctrine amendment (M4 — regression prevention)

- **REQ-I18N-009** (Capability gate): **Where** the template-neutrality CI guard exists (`internal/template/template_neutrality_audit_test.go` `neutralityClasses` slice, classes C1/C2/C4/C5/C6/C8), the sweep shall add a new lint class (e.g., `C9-natural-language-canonical-form`) that detects natural-language canonical-form prose bias outside explicit localization tables — mirroring the existing neutrality-class detector structure — so a future edit entrenching a locale's canonical-form privilege cannot pass CI green silently.
- **REQ-I18N-010** (Capability gate): **Where** the template-internal-isolation doctrine SSOTs govern template content — `.moai/docs/template-internal-isolation-doctrine.md` §25 (Template Internal-Content Isolation) and `CLAUDE.local.md` §15 "템플릿 언어 중립성" (line 611) — the sweep shall propose an amendment documenting natural-language canonical-form neutrality as a governed dimension (closing the doctrinal blind spot identified in `research.md` §C), with the amendment scope limited to a new sub-section that does not alter existing C1–C8 class definitions.

### §B.6 Baseline preservation + Template-First sync

- **REQ-I18N-011** (Unwanted behavior): The two clean English-canonical baseline files — `core/agent-common-protocol.md` and `workflow/goal-directive.md` — shall remain byte-identical pre/post sweep (0 Hangul outside any localization table, preserved verbatim); the sweep shall not touch these files.
- **REQ-I18N-012** (Ubiquitous): All edits shall be authored in the template tree first (`internal/template/templates/.claude/rules/moai/...`) per the Template-First rule (CLAUDE.local.md §2), and the local tree (`.claude/rules/moai/...`) shall be synced byte-identical for every modified file post-run-phase (`diff -rq` clean).

---

## §C Non-functional constraints

- **Zero Go source changes except the optional M4 lint.** M1–M3 are markdown-only. M4 adds exactly one new test class to `internal/template/template_neutrality_audit_test.go` (and optionally the doctrine markdown amendment). No production Go code changes.
- **Write scope whitelist** (run-phase): the 3 always-loaded rule files (× 2 trees: template + local), `internal/template/template_neutrality_audit_test.go` (M4 only), `.moai/docs/template-internal-isolation-doctrine.md` (M4 only), and `.moai/specs/SPEC-TEMPLATE-I18N-CANONICAL-001/*`. Everything else — including the 2 clean baseline files, all Go production code, all skills, all other rule files — is PRESERVE.
- **Localization mechanism is sacrosanct.** REQ-I18N-001 + REQ-I18N-013 bind absolutely: the sweep changes canonical *framing*, never localization *coverage*.
- **Verification claims** follow `.claude/rules/moai/core/verification-claim-integrity.md` — every AC PASS cites an executed command + verbatim output. Hangul-count verification uses `rg -oN '[가-힣]'` (ripgrep) for portability across locale-handling differences; windowed-grep pipelines (`sed -n <window> | grep`) are a known undercount risk (`feedback_windowed_grep_undercount_authoring`) and are prohibited for this purpose.

---

## §D Acceptance

The full AC matrix (27 sub-checks across 23 AC IDs — AC-I18N-001 carries 5 sub-branches a–e: 20 mechanical / 4 hybrid / 3 manual), verification commands (ripgrep-based), Given-When-Then scenarios, and REQ↔AC traceability live in `acceptance.md` (Tier L third artifact). The design-invariant ACs (AC-I18N-001a–e) are the highest-priority gates: they verify the localization mechanism still produces correct ko-locale and en-locale render output post-sweep.

---

## §E Known limitations & recorded decisions

1. **"Canonical framing" ≠ "remove Korean".** The single most important constraint (REQ-I18N-001/013): Korean stays as an equal-tier locale column in every Localization Table. The sweep moves the declared canonical default to English (matching the clean baseline + the English-fallback behavior) and removes the `"(canonical)"` privilege annotation. A reviewer who reads this SPEC as "de-Koreanize the templates" has misread it.
2. **Runtime output was never broken.** The audit verified 25 of 27 session-handoff.md Korean instances route through localization. This SPEC addresses documentation-framing bias + 2 literal leaks + authoring inconsistency, not a correctness defect. The 2 leaked strings (memory heading, goal-first example) land in user-private state / a documented-alternative example — real but low-severity.
3. **M4 (CI lint + doctrine) is optional/conditional.** The doctrinal blind spot means the CI gap is *consistent with* (not a defect of) the current doctrine. Adding the lint before the doctrine speaks would put enforcement ahead of policy. M4 is in scope but if run-phase reveals the doctrine amendment is contentious, M4 may split into a follow-up SPEC (blocker-report path).
4. **askuser-protocol.md is the highest-leverage fix, not session-handoff.md.** The user asked about session-handoff.md; the audit found askuser-protocol.md carries ~6× the Korean-canonical surface (1,176 vs 186 syllables) including the only pure-prose-doctrine instance. Priority order (M1 askuser → M2 session-handoff → M3 context-window) reflects leverage, not the order of the user's question.
5. **Hangul-count tooling portability.** All Hangul-count ACs use `rg -oN '[가-힣]'` (ripgrep) for consistency across locale-handling differences; windowed-grep pipelines (`sed -n <window> | grep`) undercount multibyte ranges and are prohibited for this purpose (`feedback_windowed_grep_undercount_authoring`). (An earlier plan-draft claimed "BSD grep returns 0" — retracted at v0.2.0 after verification: on this Darwin 25.5.0 env `grep` is ugrep 7.5.0 and `grep -oE '[가-힣]'` returns the correct 1176 count alongside `rg`; the original 0-count was a broken `ggrep|wc -l` fallback masking a missing binary, not a grep Unicode-range defect.)

---

## Out of Scope

The following are explicitly out of scope for SPEC-TEMPLATE-I18N-CANONICAL-001:

### Out of Scope — Non-always-loaded rule files

- Rule files NOT in the always-loaded set (files gated by `paths:` frontmatter restrictions, skill-local rules, language-specific rules under `.claude/rules/moai/languages/`) are not swept in this SPEC. The systemic finding (research.md §B.4) recommends a follow-up sweep if those files are later found to carry Korean-canonical content.

### Out of Scope — Removing Korean from localization tables

- Korean remains an equal-tier locale column in every Localization Table (REQ-I18N-013). Removing Korean content from loc-tables, examples meant to demonstrate Korean rendering, or the ko column itself is explicitly prohibited — this SPEC changes canonical *framing*, never locale *coverage*.

### Out of Scope — Production Go code changes

- M1–M3 are markdown-only. M4 adds one test class + optionally a doctrine markdown sub-section. No changes to production Go source under `internal/`, `pkg/`, `cmd/` (other than the test file) are in scope.

### Out of Scope — The two clean baseline files

- `core/agent-common-protocol.md` and `workflow/goal-directive.md` are 0-Hangul English-canonical at baseline and are PRESERVE (REQ-I18N-011). Any needed change there is a separate SPEC.

### Out of Scope — docs-site / README / non-rule-file localization

- The docs-site (`adk.mo.ai.kr`) and project READMEs have their own i18n workflow (SPEC-V3R6-DOCS-I18N-* lineage). This SPEC governs only the template-distributed always-loaded *rule* files.

### Out of Scope — Localization mechanism redesign

- The 4-locale table architecture, the render-time substitution instruction, and the English-fallback rule are PRESERVED as-is (REQ-I18N-001). Redesigning the localization mechanism (e.g., moving to gettext/ICU, adding new locales) is a separate architectural SPEC.
