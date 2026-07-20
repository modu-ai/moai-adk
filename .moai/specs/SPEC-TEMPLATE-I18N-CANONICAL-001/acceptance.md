# Acceptance Criteria — SPEC-TEMPLATE-I18N-CANONICAL-001

Each AC declares its verification **Method**: `mechanical` (deterministic command → PASS/FAIL), `hybrid` (mechanical proxy + one-line manual confirmation), or `manual` (scenario / evidence cross-check). Commands assume repo root `/Users/goos/MoAI/moai-adk-go` and **ripgrep (`rg`)** for Hangul ranges, for portability across locale-handling differences and to avoid the windowed-grep undercount risk (`sed -n <window> | grep`; `feedback_windowed_grep_undercount_authoring`). (An earlier draft claimed "BSD grep returns 0 on macOS" — retracted at v0.2.0 after verification: on this Darwin 25.5.0 env `grep` is ugrep 7.5.0 and returns the correct count alongside `rg`; the AC commands are unchanged.)

**Baseline reference** (plan-phase, `39c74d777`, ripgrep-verified): askuser-protocol.md 1,176 Hangul; session-handoff.md 186; context-window-management.md 18; agent-common-protocol.md 0; goal-directive.md 0. Localization Table row count in session-handoff.md = 9.

---

## §D AC Matrix

### §D.1 Design Invariant — localization mechanism preservation (CRITICAL)

| AC ID | Method | Requirement | Verification |
|-------|--------|-------------|--------------|
| AC-I18N-001a | mechanical | REQ-I18N-001 | Localization Table row count in `session-handoff.md` post-sweep ≥ 9 (plan-baseline floor). Command: `grep -cE '^\|.*(English\|Korean\|Japanese\|Chinese\|Copy from here\|여기부터\|Block\|Cut-line)' internal/template/templates/.claude/rules/moai/workflow/session-handoff.md` → count ≥ 9 (REQ-I18N-006 adds the memory-heading row, so expected ≥ 10). |
| AC-I18N-001b | mechanical | REQ-I18N-001 / 013 | The 4 locale columns are present post-sweep: `English` and `Korean` in `session-handoff.md` inline table; `Japanese` and `Chinese` in `session-handoff-examples.md`. Commands: `grep -c '^| English |' session-handoff.md` ≥ 1; `grep -c '^| Korean |' session-handoff.md` ≥ 1; `grep -Ei 'japanese\|chinese' session-handoff-examples.md` returns the extension-table header. |
| AC-I18N-001c | hybrid | REQ-I18N-001 | **ko-locale render path preserved**: the Localization Table Korean column is populated for every Block 1/3/5/6 header + cut-line marker element, AND the render-time substitution instruction still references `conversation_language`. Mechanical: `grep -c 'conversation_language' session-handoff.md` ≥ 1 (the substitution instruction survives); `grep -E '여기부터\|전제 검증\|실행\|머지 후' session-handoff.md` returns ≥ 1 hit INSIDE the Localization Table Korean column (the Korean data survives, just not as the default skeleton). Manual: confirm the Korean column cells are intact. |
| AC-I18N-001d | hybrid | REQ-I18N-001 | **en-locale render path is now the default**: the canonical 6-block skeleton uses English-first headers. Mechanical: the skeleton block (§ Canonical Format) contains `Preconditions:` / `Run:` / `After merge:` (or locale-neutral placeholders) and does NOT contain `전제 검증:` / `실행:` / `머지 후:` as default-skeleton headers. Manual: confirm the English column is populated. |
| AC-I18N-001e | mechanical | REQ-I18N-001 | **No new Korean-canonical prose leaks**: post-sweep Hangul count in the 3 swept files drops from baseline (askuser 1,176 → toward 0; session-handoff 186 → residual only in loc-table column + cut-line marker examples; context-window 18 → toward 0). Command: `for f in <3 swept files>; do rg -oN '[가-힣]' "$f" \| wc -l; done` — counts ≤ pre-sweep baselines (no regression). |

### §D.2 askuser-protocol.md (M1)

| AC ID | Method | Requirement | Verification |
|-------|--------|-------------|--------------|
| AC-I18N-002 | mechanical | REQ-I18N-002 | The § "Recommendation Placement Principles" section is English-canonical: `rg -oN '[가-힣]'` scoped to the section (awk-windowed between `## ` headings) returns 0 (the `(권장)` token in cross-reference prose is the only permitted residual, and only if it appears in a locale-aware discussion). |
| AC-I18N-003 | mechanical | REQ-I18N-003 | The 3 worked examples (anti-pattern ~line 32, Epic-8 ~line 246, correct-pattern ~line 436) are English-canonical: Hangul count in those example blocks is 0 (excluding the `(권장)` token discussion, which is locale-aware by design). |
| AC-I18N-004 | mechanical | REQ-I18N-002 | The `[ZONE:Evolvable]` tags, Fisher-information formula `I=p(1−p)`, the `p ≈ 0.5` cold-start heuristic, and the 5-principle numbering are preserved verbatim in English. Commands: `grep -c '\[ZONE:Evolvable\]' askuser-protocol.md` ≥ baseline; `grep 'I=p(1' askuser-protocol.md` returns the formula; `grep -E '^[0-9]\. ' section` returns the 5 principles. |
| AC-I18N-005 | mechanical | REQ-I18N-002 | Cross-references preserved verbatim: `grep 'verification-claim-integrity.md §1.1 surface 3' askuser-protocol.md` ≥ 1; `grep 'design.md §A.4' askuser-protocol.md` ≥ 1; `grep 'design.md §B.2' askuser-protocol.md` ≥ 1. |

### §D.3 session-handoff.md (M2)

| AC ID | Method | Requirement | Verification |
|-------|--------|-------------|--------------|
| AC-I18N-006 | mechanical | REQ-I18N-004 | The canonical 6-block skeleton uses English-first or locale-neutral headers. Command: scoped to the § Canonical Format fenced block, `grep -E '전제 검증:|실행:|머지 후:|후속:' ` returns 0 hits IN THE SKELETON BLOCK (Korean renderings appear only inside the Localization Table Korean column). |
| AC-I18N-007 | mechanical | REQ-I18N-005 | The string `Korean (canonical)` does not appear anywhere in `session-handoff.md`. Command: `grep -c 'Korean (canonical)' session-handoff.md` → 0. The column header is `Korean` (equal-tier). |
| AC-I18N-008 | mechanical | REQ-I18N-005 | The phrase `primary locales` does not appear; replaced by `inline locales` or equivalent neutral tiering. Command: `grep -c 'primary locales' session-handoff.md` → 0; `grep -E 'inline locales\|inline columns' session-handoff.md` ≥ 1. |
| AC-I18N-009 | hybrid | REQ-I18N-005 | The line-67/82 tiering contradiction is resolved. Manual: the tiering prose is consistent with the English-fallback behavior (the declared structural default matches the fallback). Mechanical: the fallback rule (line ~82) still names English as the skeleton fallback for non-table ISO-639 codes. |
| AC-I18N-010 | mechanical | REQ-I18N-006 | The Localization Table has a row for the memory heading: `grep -E 'Next Session Entry Point\|다음 세션 시작점' session-handoff.md` returns the inline en/ko row; `grep -E '次セッション開始点\|下一会话起点' session-handoff-examples.md` returns the ja/zh extension row. |
| AC-I18N-011 | mechanical | REQ-I18N-006 | The close-time-pruning cross-reference (line ~229) matches by content, not by Korean literal. Command: the cross-reference reads "the next-session-start-point section" (or equivalent content-keyed phrasing), not `## 다음 세션 시작점` as a literal anchor. |
| AC-I18N-012 | mechanical | REQ-I18N-007 | The goal-first bootstrap example (line ~163) is English-canonical with a locale note. Command: scoped to the goal-first bootstrap fenced block, `rg -oN '[가-힣]'` count is 0 (no Korean-only condition); a locale note ("condition text follows the user's `conversation_language`" or equivalent) is present. |

### §D.4 context-window-management.md (M3)

| AC ID | Method | Requirement | Verification |
|-------|--------|-------------|--------------|
| AC-I18N-013 | mechanical | REQ-I18N-008 | The resume-message format example (§ Orchestrator Responsibilities, lines ~66-71) is English-canonical. Command: scoped to the fenced format block, `rg -oN '[가-힣]'` count is 0 (or the block carries an explicit locale note + cross-reference). |
| AC-I18N-014 | mechanical | REQ-I18N-008 | The file cross-references `session-handoff.md` Localization Table. Command: `grep -E 'session-handoff.md.*Localization Table\|session-handoff.md § Localization' context-window-management.md` ≥ 1. |

### §D.5 CI lint + doctrine amendment (M4)

| AC ID | Method | Requirement | Verification |
|-------|--------|-------------|--------------|
| AC-I18N-015 | mechanical | REQ-I18N-009 | The template-neutrality CI test has a new class (e.g., `C9-natural-language-canonical-form`) that detects locale canonical-form prose bias outside localization tables. Commands: `grep -E 'C9\|natural-language-canonical-form\|canonical-form' internal/template/template_neutrality_audit_test.go` ≥ 1; `go test ./internal/template/ -run TestTemplateNeutrality` exit 0. |
| AC-I18N-016 | hybrid | REQ-I18N-010 | The doctrine file carries a new sub-section documenting natural-language canonical form as a governed dimension. Mechanical: `grep -E 'natural-language\|canonical-form\|natural language neutrality' .moai/docs/template-internal-isolation-doctrine.md` ≥ 1. Manual: confirm the amendment does NOT alter existing C1–C8 class definitions (scope-limit check). |

### §D.6 Baseline preservation

| AC ID | Method | Requirement | Verification |
|-------|--------|-------------|--------------|
| AC-I18N-017 | mechanical | REQ-I18N-011 | The 2 clean baseline files are byte-identical pre/post sweep. Command: `git diff 39c74d777..HEAD -- internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md internal/template/templates/.claude/rules/moai/workflow/goal-directive.md .claude/rules/moai/core/agent-common-protocol.md .claude/rules/moai/workflow/goal-directive.md` → empty (4 paths untouched). |
| AC-I18N-018 | mechanical | REQ-I18N-011 | The 2 clean baseline files remain 0-Hangul. Command: `rg -oN '[가-힣]' agent-common-protocol.md goal-directive.md` → 0 for both (pre-sweep baseline preserved). |

### §D.7 Template-First + sync

| AC ID | Method | Requirement | Verification |
|-------|--------|-------------|--------------|
| AC-I18N-019 | mechanical | REQ-I18N-012 | **Sweep-induced changes** are synced template ↔ local for the modified files; **pre-existing §25-intentional dev-trace drift is preserved** (NOT silently "synced away"). Command: for session-handoff.md, context-window-management.md, session-handoff-examples.md → `diff -rq <template> <local>` identical (sweep edits synced). For askuser-protocol.md → the diff is EXACTLY the pre-existing §25-intentional drift (template neutral, local retains `SPEC-V3R6-ASKUSER-DECISION-MEMORY-001` provenance) PLUS the sweep's English-canonical translation applied to BOTH trees — i.e. the sweep-induced delta is byte-identical across trees, the pre-existing drift is untouched. (D2 narrowed: the "byte-identical" target applied only at the `39c74d777` baseline; at current HEAD askuser is intentionally diverged.) |
| AC-I18N-020 | mechanical | REQ-I18N-012 | `go build ./...` exit 0 (no production Go touched except the M4 test file); `go test ./internal/template/...` exit 0 (no regression). |

### §D.8 Manual scenarios (Given-When-Then)

| AC ID | Method | Requirement | Scenario |
|-------|--------|-------------|----------|
| AC-I18N-021 | manual | REQ-I18N-002 | **Given** a non-Korean orchestrator session reading `askuser-protocol.md`, **when** it reaches the § Recommendation Placement Principles section, **then** it can parse all 5 HARD principles directly from the file in English (the Fisher-information emission-timing rule, the question-ordering rule, the statistical-majority default, the precondition-statement rule, the adaptive-strength rule) without needing to infer from cross-references. |
| AC-I18N-022 | manual | REQ-I18N-001 | **Given** a ko-locale user (`.moai/config/sections/language.yaml` `conversation_language: ko`), **when** the orchestrator renders a paste-ready resume, **then** the cut-line markers and Block 1/3/5/6 headers are Korean (Localization Table Korean column + render-time substitution both intact). |
| AC-I18N-023 | manual | REQ-I18N-001 | **Given** an en-locale user (`conversation_language: en`), **when** the orchestrator renders a paste-ready resume, **then** the cut-line markers and Block 1/3/5/6 headers are English (the English column is now the default skeleton, consistent with the English-fallback rule). |

---

## §D.9 Traceability (REQ → AC)

| REQ | AC(s) |
|-----|-------|
| REQ-I18N-001 (design invariant) | AC-I18N-001a / 001b / 001c / 001d / 001e, AC-I18N-022, AC-I18N-023 |
| REQ-I18N-002 (askuser policy) | AC-I18N-002, AC-I18N-004, AC-I18N-005, AC-I18N-021 |
| REQ-I18N-003 (askuser examples) | AC-I18N-003 |
| REQ-I18N-004 (session-handoff skeleton) | AC-I18N-006 |
| REQ-I18N-005 (loc-table labels + contradiction) | AC-I18N-007, AC-I18N-008, AC-I18N-009 |
| REQ-I18N-006 (memory heading) | AC-I18N-010, AC-I18N-011 |
| REQ-I18N-007 (goal-first example) | AC-I18N-012 |
| REQ-I18N-008 (context-window resume format) | AC-I18N-013, AC-I18N-014 |
| REQ-I18N-009 (CI lint) | AC-I18N-015 |
| REQ-I18N-010 (doctrine amendment) | AC-I18N-016 |
| REQ-I18N-011 (baseline preservation) | AC-I18N-017, AC-I18N-018 |
| REQ-I18N-012 (Template-First + sync) | AC-I18N-019, AC-I18N-020 |
| REQ-I18N-013 (unwanted — no Korean removal) | AC-I18N-001a / 001b / 001c |

Every REQ maps to ≥ 1 AC. Coverage complete.

---

## §D.10 Severity and gates

- **MUST-PASS (blocking)**: AC-I18N-001a / 001b / 001c / 001d (design invariant — localization mechanism preservation), AC-I18N-017 / 018 (baseline preservation), AC-I18N-007 (the "(canonical)" label removal — the headline framing fix), AC-I18N-019 (Template-First parity).
- **SHOULD-PASS**: AC-I18N-002 / 003 / 004 / 005 (askuser translation fidelity), AC-I18N-006 / 010 / 011 / 012 (session-handoff skeleton + 2 leak fixes), AC-I18N-013 / 014 (context-window), AC-I18N-020 (build/test green).
- **NICE-TO-HAVE (M4 conditional)**: AC-I18N-015 / 016 (CI lint + doctrine amendment). If M4 is split to a follow-up SPEC per plan.md §F M4 conditional-halt, these ACs defer.

**Definition of Done**: all MUST-PASS + SHOULD-PASS ACs green; design invariant verified end-to-end (ko-locale and en-locale render paths produce correct localized output); the 2 clean baseline files untouched; Template-First parity holds; M4 either green or formally deferred via blocker-report + follow-up SPEC.

---

## §D.11 Edge cases

1. **The `(권장)` token** appears in askuser-protocol.md worked examples and is locale-aware by design (the Localization Table carries its own row: ko `(권장)` / en `(Recommended)`). Translation of the worked examples MUST preserve the `(권장)` token discussion — it is not a "Korean leak" but a documented locale-aware marker. AC-I18N-003 explicitly excludes it from the Hangul-count zero-target.
2. **The `✂────` cut-line markers and `─` decorators** are preserved verbatim across all locales (they are symbols, not text). The sweep touches only the text between them. REQ-I18N-001 implicitly covers this (the cut-line markers are part of the localization mechanism).
3. **`session-handoff-examples.md` split-site editing**: REQ-I18N-006's memory-heading row lands in BOTH the inline table (en/ko) AND the examples file (ja/zh). Forgetting the examples file would drop the ja/zh columns and violate REQ-I18N-013. AC-I18N-010 checks both sites.
4. **Windowed-grep undercount (the real hazard, v0.2.0-corrected)**: if run-phase verification uses a windowed-grep pipeline (`sed -n <window> | grep '[가-힣]'`), it can undercount multibyte ranges — a false PASS. All Hangul-count ACs specify ripgrep (`rg -oN '[가-힣]'`); plan.md §B records the hazard. NOTE: the earlier "BSD grep returns 0 on macOS" framing was retracted at v0.2.0 — on this env `grep -oE` (ugrep 7.5.0) returns the correct 1176 count; whole-file `grep -oE` is NOT the hazard, windowed-grep pipelines are.
5. **M4 C9 false-positives**: the new lint class must scope to OUTSIDE explicit localization tables. Legitimate Korean content (loc-table cells, examples meant to demonstrate Korean rendering, the `(권장)` token) must be allow-listed. Threshold tuning is a run-phase concern.
