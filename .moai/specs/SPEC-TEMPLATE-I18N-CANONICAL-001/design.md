# Design — SPEC-TEMPLATE-I18N-CANONICAL-001

> Tier L 5-artifact contract member (`spec.md` + `plan.md` + `acceptance.md` + `design.md` + `research.md`, per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier). The `progress.md` sibling is an additional operational artifact (§E skeleton) required by the manager-spec agent protocol; it does NOT replace this file.

---

## §A Architecture intent — the canonical-framing move

### §A.1 What moves, what stays

The sweep is a **framing relocation**, not a content deletion. Concretely:

| Surface | Before (Korean-canonical) | After (English-canonical) | Localization mechanism |
|---------|--------------------------|--------------------------|----------------------|
| `session-handoff.md` canonical 6-block skeleton | Korean-first headers (`전제 검증:` / `실행:` / `머지 후:`) | English-first (`Preconditions:` / `Run:` / `After merge:`) OR locale-neutral placeholders | **PRESERVED** — Korean renderings move INTO the Localization Table Korean column; the 4-locale table + render-time substitution + English-fallback survive unchanged |
| `session-handoff.md` Localization Table header | `Korean (canonical)` | `Korean` (equal-tier) | **PRESERVED** — all 4 locale columns (en/ko inline + ja/zh extension) survive; row count ≥ 9 (floor) + 1 new memory-heading row |
| `askuser-protocol.md` Recommendation Placement Principles | ~50 lines Korean prose (binding `[ZONE:Evolvable]` policy) | English-canonical prose | n/a (this is prose doctrine, not a localization-table surface) |
| `context-window-management.md` resume format example | Korean-only | English-canonical + cross-ref to session-handoff.md loc-table | **PRESERVED** — the format now defers to the shared loc-table |

### §A.2 The invariant boundary (design-contract)

The invariant is drawn tightly so run-phase has a crisp go/no-go boundary:

- **The localization *mechanism* is sacrosanct**: the 4-locale table data, the render-time substitution instruction keyed on `conversation_language`, and the English-skeleton fallback rule for non-table ISO-639 codes ALL survive the sweep with semantics unchanged.
- **The canonical *framing* moves**: the default-skeleton rendering, the `"(canonical)"` privilege annotation, and the Korean-first tiering prose are the only things that change.

This separation is what makes the design-invariant ACs (AC-I18N-001a–e) mechanically checkable: row counts, column presence, and the substitution instruction are grep-able; the framing move is a diff on specific blocks.

### §A.3 Why English-canonical (not locale-neutral placeholders)

Two resolution paths were available (research.md §C.2): (a) locale-neutral placeholders, or (b) English-canonical framing matching the two clean baseline files. This SPEC takes path (b) because:

1. **Consistency with the clean baseline.** `agent-common-protocol.md` and `goal-directive.md` (both 0-Hangul, always-loaded) establish the project's English-canonical authoring convention. The three leaking files should match them.
2. **Consistency with the English-fallback rule.** `session-handoff.md` line ~82 already sends the 12 non-en/ko/ja/zh locales to the English column for the structural skeleton. Making the canonical skeleton English aligns the default with the fallback behavior (resolves the D-tiering contradiction).
3. **Concrete renderability.** Locale-neutral placeholders (`<Block 3 header per Localization Table>`) are harder for an orchestrator to reason about than a concrete English default with a documented substitution path. English-first + loc-table is the lower-friction mental model.

Korean remains an equal-tier locale column — the `Korean` column in the Localization Table carries the concrete Korean renderings, and the render-time substitution instruction produces Korean output for ko-locale sessions.

---

## §B CI lint class C9 — `natural-language-canonical-form`

### §B.1 Detector pattern (mirrors C1–C8 structure)

The new class mirrors the existing `neutralityClasses` slice entries in `internal/template/template_neutrality_audit_test.go`. Each existing class is a `neutralityClass` struct with a `name`, a regexp `pattern`, and an allow-list mechanism. C9 follows the same shape:

- **name**: `C9-natural-language-canonical-form`
- **detection surface**: natural-language canonical-form prose bias OUTSIDE explicit localization tables. Concretely:
  - a `"(canonical)"` column-label detector — flags any Localization Table column header carrying a `(canonical)` parenthetical (privilege-marker).
  - a Hangul-concentration heuristic for canonical-skeleton / prose blocks — flags blocks of Korean prose outside loc-table regions that exceed a threshold (tuning at run-phase; the askuser ~1176-syllable policy section is the reference target; the `(권장)` token in a locale-aware discussion is allow-listed).
- **allow-list**: explicit Localization Table regions (the en/ko inline table in `session-handoff.md`, the ja/zh extension in `session-handoff-examples.md`, and any `| Element | <lang> |` markdown table row) are excluded — Korean content inside a loc-table cell is legitimate, not a defect.

### §B.2 Scope boundary (what C9 does NOT check)

- C9 does NOT flag Korean content inside localization tables (that is the localization mechanism doing its job).
- C9 does NOT flag the `(권장)` / `(Recommended)` token (locale-aware by design; the loc-table carries its own row).
- C9 does NOT flag the `✂────` cut-line markers or `─` decorators (symbols, not text).
- C9 does NOT enforce a zero-Korean policy — it flags canonical-form *privilege* (the "(canonical)" label) and *prose-doctrine concentration* (Korean binding-policy sections outside loc-tables), not incidental Korean in examples.

### §B.3 Severity

Binary FAIL on the `"(canonical)"` column-label detector (a deliberate privilege marker); advisory WARN on the Hangul-concentration heuristic (threshold-dependent). This mirrors the C1 (binary FAIL) vs C2 (advisory WARN) split in the existing class set.

---

## §C Doctrine amendment structure

### §C.1 Target SSOTs (D4-corrected)

The doctrine has TWO SSOT locations (verified at plan baseline):
- `.moai/docs/template-internal-isolation-doctrine.md` §25 — Template Internal-Content Isolation (owns the C1–C8 forbidden-content classes; internal-trace dimension).
- `CLAUDE.local.md` §15 — "템플릿 언어 중립성" (line 611; owns 16-programming-language neutrality).

§15 in CLAUDE.local.md is the natural-language dimension's home (it already governs language neutrality for programming languages; the amendment extends its spirit to natural languages). §25 in the doctrine file is the enforcement-classes home (where C9 would be catalogued alongside C1–C8).

### §C.2 Amendment shape

A new sub-section — proposed placement: `CLAUDE.local.md` §15 (or a cross-referenced §25.6 in the doctrine file) — with this scope:

- **Statement**: natural-language canonical-form neutrality is a governed dimension. Template-distributed content declares English as the canonical framing default (matching the English-fallback rule); all other supported natural languages are equal-tier locale columns in localization tables.
- **Forbidden**: declaring any natural-language column "(canonical)" in a localization table; authoring binding prose doctrine (HARD/SHOULD rules) in a non-English language outside a localization table.
- **Allowed**: concrete locale renderings inside localization tables; the `(권장)`/`(Recommended)` locale-aware token; symbols (`✂`, `─`).
- **Scope limit**: the amendment does NOT alter existing C1–C8 class definitions; it adds a new dimension alongside them.

### §C.3 Conditional-halt (M4 split path)

If the doctrine amendment is contentious at run-phase (e.g., a reviewer argues natural-language neutrality is out-of-scope for §25's internal-trace focus, or that CLAUDE.local.md §15's programming-language scope should not extend to natural languages), M4 splits into a follow-up SPEC. The CI lint C9 is conditional on the doctrine speaking first (research.md §C.2 path (b)); landing enforcement ahead of policy is the named anti-pattern. The blocker-report path governs this split.

---

## §D Cross-references

- `spec.md` §A.3 (design invariant) + §B.1 (REQ-I18N-001/013) — the invariant boundary this design operationalizes.
- `spec.md` §B.5 (REQ-I18N-009/010) — the C9 lint + doctrine amendment requirements.
- `research.md` §C (doctrinal blind-spot analysis) — the gap this design closes.
- `plan.md` §F M4 (CI lint + doctrine amendment milestone) — the run-phase execution plan.
- `acceptance.md` AC-I18N-015/016 (M4 gates) + AC-I18N-001a–e (design-invariant gates).
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — the Tier L 5-artifact contract this file satisfies.
