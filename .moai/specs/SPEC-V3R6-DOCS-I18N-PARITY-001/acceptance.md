# Acceptance Criteria — SPEC-V3R6-DOCS-I18N-PARITY-001

All criteria are **binary** and **command-verifiable**. The authoritative tool is
`scripts/docs-i18n-check.sh`. Run all commands from the repo root
(`/Users/goos/MoAI/moai-adk-go`). Each AC is PASS iff the stated count/exit is met.

> Baseline captured at plan time (HEAD `038f6e793`): **53 errors total** =
> 4 frontmatter + 26 H1 + 23 glossary.

---

## D. AC Matrix

| AC ID | Category | Verifies REQ | Target |
|-------|----------|--------------|--------|
| AC-DIP-001 | C1 frontmatter | REQ-DIP-001 | 0 frontmatter errors |
| AC-DIP-002 | C2 H1 (Set A) | REQ-DIP-002 | 0 H1 errors in the 5 all-locale files |
| AC-DIP-003 | C2 H1 (Set B stubs) | REQ-DIP-002, 006 | 0 H1 errors in multi-llm stubs |
| AC-DIP-004 | C3 glossary | REQ-DIP-003, 006 | 0 glossary errors |
| AC-DIP-005 | C3 SPEC-First (zh) | REQ-DIP-003 | SPEC-First present in 2 zh files |
| AC-DIP-006 | full warn-mode gate | REQ-DIP-004 | `Errors:   0` |
| AC-DIP-007 | full strict-mode gate | REQ-DIP-005 | exit 0 |
| AC-DIP-008 | file-path parity intact | REQ-DIP-007 | 78 .md per locale, no add/remove |
| AC-DIP-009 | authoring-contract regression guard | REQ-DIP-008 | 0 forbidden URLs; canonical URL intact; Mermaid non-TD count not increased vs baseline (≤ 4) |

### Traceability — REQ ↔ AC (no orphan REQ, no dangling AC)

| REQ | Mapped AC(s) |
|-----|--------------|
| REQ-DIP-001 | AC-DIP-001 |
| REQ-DIP-002 | AC-DIP-002, AC-DIP-003 |
| REQ-DIP-003 | AC-DIP-004, AC-DIP-005 |
| REQ-DIP-004 | AC-DIP-006 |
| REQ-DIP-005 | AC-DIP-007 |
| REQ-DIP-006 | AC-DIP-003, AC-DIP-004 |
| REQ-DIP-007 | AC-DIP-008 |
| REQ-DIP-008 | **AC-DIP-009** |

Every REQ-DIP-001..008 maps to at least one AC; every AC-DIP-001..009 maps to at
least one REQ. No orphaned requirement, no dangling criterion.

---

## D.1 Category-level criteria

### AC-DIP-001 — C1 frontmatter (4 → 0)

`prompt-caching.md` in all 4 locales has a frontmatter block with a non-empty `title`.

```bash
DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh 2>&1 | grep -c "no frontmatter block"
# PASS iff output == 0
```

Supplementary (each file starts with `---`):
```bash
for L in ko en ja zh; do head -1 "docs-site/content/$L/cost-optimization/prompt-caching.md"; done
# PASS iff every line == ---
```

### AC-DIP-002 — C2 H1 Set A (5 all-locale files → 0)

The 5 files that were missing H1 in all 4 locales now each have an `#` H1.

```bash
DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh 2>&1 \
  | grep "no H1 heading" \
  | grep -E "advanced/harness-profiles|core-concepts/harness-engineering|db/migration-patterns|getting-started/profile|workflow-commands/moai-design" \
  | wc -l
# PASS iff output == 0
```

### AC-DIP-003 — C2 H1 Set B multi-llm stubs (en/ja/zh → 0)

The `draft: true` stub pages `multi-llm/cg-mode.md` and `multi-llm/model-policy.md`
(en/ja/zh) now contain an `#` H1.

```bash
DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh 2>&1 \
  | grep "no H1 heading" | grep "multi-llm" | wc -l
# PASS iff output == 0
```

Combined C2 check (Set A + Set B):
```bash
DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh 2>&1 | grep -c "no H1 heading"
# PASS iff output == 0
```

### AC-DIP-004 — C3 glossary parity (23 → 0)

Every canonical glossary term present in a ko source appears verbatim
(case-sensitive) in its en/ja/zh counterparts.

```bash
DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh 2>&1 | grep -c "glossary term"
# PASS iff output == 0
```

Per-term spot checks (verbatim presence):
```bash
# Claude Code (9): en/ja/zh × {multi-llm/_index, multi-llm/cg-mode, multi-llm/model-policy}
for L in en ja zh; do for F in multi-llm/_index multi-llm/cg-mode multi-llm/model-policy; do
  grep -Fq -- "Claude Code" "docs-site/content/$L/$F.md" && echo "OK $L/$F" || echo "FAIL $L/$F"; done; done

# MoAI-ADK (9): en/ja/zh × {contributing/_index, multi-llm/_index, multi-llm/model-policy}
for L in en ja zh; do for F in contributing/_index multi-llm/_index multi-llm/model-policy; do
  grep -Fq -- "MoAI-ADK" "docs-site/content/$L/$F.md" && echo "OK $L/$F" || echo "FAIL $L/$F"; done; done

# moai-adk (3): en/ja/zh × {contributing/_index}
for L in en ja zh; do
  grep -Fq -- "moai-adk" "docs-site/content/$L/contributing/_index.md" && echo "OK $L" || echo "FAIL $L"; done
# PASS iff every line starts with OK
```

### AC-DIP-005 — C3 SPEC-First in zh getting-started (2 → 0)

The literal string `SPEC-First` appears in the two flagged zh files.

```bash
grep -Fq -- "SPEC-First" docs-site/content/zh/getting-started/introduction.md && echo OK1 || echo FAIL1
grep -Fq -- "SPEC-First" docs-site/content/zh/getting-started/update.md && echo OK2 || echo FAIL2
# PASS iff OK1 and OK2
```

### AC-DIP-009 — authoring-contract regression guard (canonical-URL / Mermaid TD-only / forbidden-URL)

Maps **REQ-DIP-008** (State-driven: *"While editing any docs-site content file, the
editor shall preserve the canonical-URL and Mermaid TD-only rules … and shall not
introduce any forbidden URL"*). The operative word is **introduce** — this AC is a
regression guard against the parity edits adding a NEW contract violation, not a
retroactive sweep of pre-existing docs-site state. Ground-truth contract values are
from `.moai/docs/docs-site-i18n-rules.md` §17.1 (forbidden URLs + canonical URL) and
§17.2 (Mermaid TD-only). Run all commands from repo root.

**Baseline captured at this AC's authoring (HEAD `66da0bdec`)** — these are the
fixed reference counts the guard compares against:

| Dimension | Whole-docs-site baseline | Edited-file-set baseline |
|-----------|--------------------------|--------------------------|
| Forbidden URLs (`docs.moai-ai.dev`, `adk.moai.com`, `adk.moai.kr`) | 0 | 0 |
| Mermaid non-TD direction (`graph LR` etc.) | 4 (all in `core-concepts/harness-engineering.md`, 1 per locale — pre-existing, out of scope per spec.md §C) | 4 (same files) |

**(a) Forbidden-URL absence — target 0 (no NEW forbidden URL introduced).**

The three forbidden URLs (`docs.moai-ai.dev`, `adk.moai.com`, `adk.moai.kr`) must
not appear anywhere in docs-site content. The patterns are anchored with literal
dots so the canonical `adk.mo.ai.kr` (which contains `mo.ai`, not `moai`) is NOT
matched — verified distinct strings.

```bash
grep -rIE 'docs\.moai-ai\.dev|adk\.moai\.com|adk\.moai\.kr' docs-site/content/ | wc -l
# PASS iff output == 0   (baseline 0 — guard against introducing any forbidden URL)
```

**(b) Canonical-URL preservation — the official URL remains the only docs home.**

The single canonical docs URL `https://adk.mo.ai.kr` (§17.1) must still be present
in the site (sanity check that no edit replaced it with a forbidden variant).

```bash
grep -rIFq -- "adk.mo.ai.kr" docs-site/content/ && echo "OK canonical-url-present" || echo "FAIL canonical-url-missing"
# PASS iff output == "OK canonical-url-present"
```

**(c) Mermaid TD-only — no NEW non-TD diagram introduced (non-increase vs baseline).**

§17.2 mandates `flowchart TD` / `graph TB` only; `LR` / `RL` / `BT` directions are
forbidden. The current tree has a baseline of **4** pre-existing `graph LR`
occurrences (in `core-concepts/harness-engineering.md`, one per locale) that this
SPEC does NOT fix — fixing them is a separate, broader docs effort excluded by
spec.md §C. This guard therefore asserts the count does NOT INCREASE (no new non-TD
diagram added by the parity edits), matching REQ-DIP-008's "shall not introduce"
semantics.

```bash
DIRCNT=$(grep -rIE '(flowchart|graph)[[:space:]]+(LR|RL|BT)' docs-site/content/ | wc -l | tr -d '[:space:]')
[ "$DIRCNT" -le 4 ] && echo "OK mermaid-td (count=$DIRCNT ≤ baseline 4)" || echo "FAIL mermaid-td (count=$DIRCNT > baseline 4 — new non-TD diagram introduced)"
# PASS iff output starts with "OK mermaid-td"
```

PASS for AC-DIP-009 iff all three sub-checks (a), (b), (c) pass.

---

## D.2 Closure gates (must-pass)

### AC-DIP-006 — full checker reports 0 errors (warn mode)

```bash
DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh 2>&1 | grep "Errors:"
# PASS iff line == "Errors:   0"
```

### AC-DIP-007 — strict mode exits 0

```bash
bash scripts/docs-i18n-check.sh; echo "exit=$?"
# PASS iff the final stdout line is the OK message AND exit=0
# Expected tail: "OK: all 4 locales pass parity, frontmatter, H1, and glossary checks." + exit=0
```

### AC-DIP-008 — file-path parity unchanged (no regression)

```bash
for L in ko en ja zh; do
  printf "%s: " "$L"; find "docs-site/content/$L" -type f -name '*.md' | wc -l; done
# PASS iff every locale reports 78 (unchanged from baseline) and no
# "missing translation" / "orphan translation" lines appear in Check 1.
DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh 2>&1 | grep -E "missing translation|orphan translation" | wc -l
# PASS iff output == 0
```

---

## D.3 Edge cases

- **EC-1 — H2 mistaken for H1**: a file may already have `## 개요`; adding an H1
  must be an `#` (single hash), not adjusting the H2. Verify with the C2 checker
  count, not by eye.
- **EC-2 — case-sensitive glossary**: `moai-adk` and `MoAI-ADK` are DISTINCT
  required strings; a file needing both must contain both literally.
- **EC-3 — `_index.md` H1 exemption**: `multi-llm/_index.md` needs glossary terms
  (C3) but NOT an H1 (Check 3 skips `_index.md`). Do not add a spurious H1 expecting
  it to matter for C2.
- **EC-4 — stub `draft: true` preserved**: stubs stay `draft: true`; adding H1 +
  glossary sentence does not require flipping to `draft: false`.
- **EC-5 — no new files**: inserting content must stay within existing files;
  creating a new `.md` breaks AC-DIP-008 parity.

---

## D.4 Definition of Done

- [ ] AC-DIP-001 PASS (C1 frontmatter = 0)
- [ ] AC-DIP-002 PASS (C2 Set A = 0)
- [ ] AC-DIP-003 PASS (C2 Set B / multi-llm = 0; combined C2 = 0)
- [ ] AC-DIP-004 PASS (C3 glossary = 0)
- [ ] AC-DIP-005 PASS (SPEC-First in 2 zh files)
- [ ] AC-DIP-006 PASS (`Errors:   0` in warn mode)
- [ ] AC-DIP-007 PASS (strict mode exit 0)
- [ ] AC-DIP-008 PASS (78 .md per locale, 0 parity/orphan lines)
- [ ] AC-DIP-009 PASS (0 forbidden URLs; canonical URL present; Mermaid non-TD count ≤ baseline 4)
- [ ] No edits outside `docs-site/content/{ko,en,ja,zh}/`
- [ ] No glossary term translated; no checker rule relaxed
- [ ] 2 local-ahead parallel-session commits untouched
- [ ] Commit on main-direct (Hybrid Trunk Tier M), authored by manager-docs
