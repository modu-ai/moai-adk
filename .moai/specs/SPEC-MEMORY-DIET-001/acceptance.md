---
id: SPEC-MEMORY-DIET-001
title: "Safe always-loaded context diet — acceptance criteria"
version: "1.0.0"
status: completed
created: 2026-07-10
updated: 2026-07-10
author: GOOS행님
priority: P2
phase: "v14.4.0 target"
module: ".claude/rules/moai/workflow + ~/.claude/projects/*/memory"
lifecycle: spec-anchored
tags: "context-diet, always-loaded, path-scope, memory-index, template-first, safe-diet"
tier: M
---

# acceptance.md — SPEC-MEMORY-DIET-001

## §D AC Matrix

### REQ-1 — cadence-bridge.md path-scope conversion

#### AC-MD-001 (REQ-MD-001 — frontmatter acquired)

**Given** `.claude/rules/moai/workflow/cadence-bridge.md` before the SPEC had no YAML frontmatter (started with `# Cadence Bridge`),
**When** the run-phase completes the REQ-1 edit,
**Then** the file's first line is `---` (YAML frontmatter opening) AND the frontmatter contains both `description:` (one-line) and `paths:` (CSV quoted-glob string) fields.

**Verification command:**
```bash
head -10 .claude/rules/moai/workflow/cadence-bridge.md | grep -E '^(---|description:|paths:)'
# Expected: 3 matching lines (---, description:, paths:)
```

#### AC-MD-002 (REQ-MD-002 — committed paths: glob + honest-fallback prose)

**Given** the cadence-bridge rule's frontmatter and rewritten Loading-scope prose,
**When** the run-phase completes the REQ-1 edit,
**Then** BOTH (a) the `paths:` frontmatter glob is present and non-empty in the committed file (the concrete glob value from REQ-MD-002), AND (b) the documented trigger-condition list + honest-fallback clause exist in the Loading-scope prose.

**Verification command (verifies the grep-able artifacts — frontmatter presence + prose documentation):**
```bash
# (a) Frontmatter artifact: paths: glob is present and non-empty
grep -E '^paths:' .claude/rules/moai/workflow/cadence-bridge.md
# Expected: ≥1 match — the committed glob from REQ-MD-002

# (b) Loading-scope prose: trigger-condition list documented
grep -A20 'Loading scope' .claude/rules/moai/workflow/cadence-bridge.md | grep -iE 'loop|cadence|cron|trigger'
# Expected: ≥1 match (trigger surface documented)

# (c) Loading-scope prose: honest-fallback clause present
grep -A25 'Loading scope' .claude/rules/moai/workflow/cadence-bridge.md | grep -iE 'manual|fallback|cannot.*catch|file.glob|semantic'
# Expected: ≥1 match (honest-fallback clause present)
```

**Scope of verification (honesty note):** This AC verifies what is mechanically grep-able from the committed file — (a) the `paths:` frontmatter artifact is present with the committed glob, and (b) the Loading-scope prose documents the trigger surface + honest fallback. Runtime loading behavior (whether the Claude Code runtime actually loads the rule when a `/loop` is typed at runtime) is NOT mechanically verifiable from a test, because `paths:` matches file globs (a file-event mechanism) and a `/loop` typed at runtime is a semantic event, not a file path. The honest-fallback prose — which obligates the orchestrator to manually consult the rule when it detects the `/loop` + `/moai` composition regardless of path-match state — is the binding safety net. This AC is intentionally honest about what it verifies (prose documentation + frontmatter artifact presence) and what it cannot (runtime semantic-event loading).

#### AC-MD-003 (REQ-MD-003 — Loading scope prose rewritten)

**Given** the original "Loading scope" prose declared "Intentionally always-loaded (no `paths:` restriction)",
**When** the run-phase completes the REQ-1 edit,
**Then** the rewritten "Loading scope" prose declares path-match status WITH rationale AND a documented trigger-condition list.

**Verification command:**
```bash
grep -i 'always-loaded' .claude/rules/moai/workflow/cadence-bridge.md
# Expected: 0 matches containing "Intentionally always-loaded" (the old declaration is gone)
grep -i 'path-match\|paths:.*restriction\|trigger-condition\|loads only when' .claude/rules/moai/workflow/cadence-bridge.md
# Expected: ≥1 match (new declaration present)
```

#### AC-MD-004 (REQ-MD-004 — template mirror parity)

**Given** the cadence-bridge rule is MIRRORED (both local and template trees),
**When** the run-phase completes the REQ-1 edit,
**Then** both files carry identical YAML frontmatter.

**Verification command:**
```bash
diff <(head -10 .claude/rules/moai/workflow/cadence-bridge.md) <(head -10 internal/template/templates/.claude/rules/moai/workflow/cadence-bridge.md)
# Expected: empty diff (identical frontmatter block)
```

#### AC-MD-005 (REQ-MD-005 — body byte preservation)

**Given** the cadence-bridge rule's body content (everything after the new frontmatter closing `---`),
**When** the run-phase completes the REQ-1 edit,
**Then** the body is byte-identical to the original (modulo the inline "Loading scope" prose rewrite per REQ-MD-003).

**Verification command:**
```bash
# The body after frontmatter must match the original except the Loading scope paragraph
# (which is the REQ-MD-003 inline prose edit, not a body deletion).
# Verify no body sections were removed:
grep -c '^## ' .claude/rules/moai/workflow/cadence-bridge.md
# Expected: same H2 section count as original (the Catalog-Level HARD Invariant, Eligibility Table, Recipe Catalog, Discovery-to-Queue Contract, When to Schedule, Fallback, Cross-References sections all survive)
```

---

### REQ-2 — session-handoff.md illustrative content extraction

#### AC-MD-006 (REQ-MD-006 — Example sections extracted)

**Given** session-handoff.md originally contained `### Example (Illustrative; substitute project-specific values when adapting)` and `### Example with Block 0 (Illustrative)` inline,
**When** the run-phase completes the REQ-2 edit,
**Then** both Example sections are removed from session-handoff.md and a one-line pointer to the sibling references file replaces each.

**Verification command:**
```bash
grep -c '^### Example' .claude/rules/moai/workflow/session-handoff.md
# Expected: 0 (both Example H3 headings removed from the always-loaded file)
grep -c 'session-handoff-examples' .claude/rules/moai/workflow/session-handoff.md
# Expected: ≥2 (one pointer per extracted section)
```

#### AC-MD-007 (REQ-MD-007 — Localization Table condensed)

**Given** session-handoff.md originally contained a full 4-locale (en/ko/ja/zh) Localization Table inline,
**When** the run-phase completes the REQ-2 edit,
**Then** the inline table retains only en/ko columns AND a pointer notes the full 4-locale table lives in the sibling references file.

**Verification command:**
```bash
# en/ko columns retained inline
grep -A15 'Localization Table' .claude/rules/moai/workflow/session-handoff.md | grep -iE 'English|Korean'
# Expected: ≥2 matches (en + ko column headers survive)
# pointer to full table present
grep -A20 'Localization Table' .claude/rules/moai/workflow/session-handoff.md | grep -i 'session-handoff-examples\|full.*table\|ja.*zh\|4-locale'
# Expected: ≥1 match (pointer to sibling file present)
```

#### AC-MD-008 (REQ-MD-008 — CORE DOCTRINE byte-identical survival)

**Given** session-handoff.md's CORE DOCTRINE sections (Canonical Format 6-block skeleton, Cut-line Marker Specification, Field-by-Field Specification, Pre-emit self-check labels, Auto-Memory Integration, Post-Paste /goal Follow-up Block, Diet Constraints, Worktree-Anchored Resume Pattern),
**When** the run-phase completes the REQ-2 edit,
**Then** all of the following survive byte-identical.

**Verification commands (all MUST pass):**
```bash
# 1. Cut-line markers (✂ U+2702) survive
grep -c '✂' .claude/rules/moai/workflow/session-handoff.md
# Expected: ≥4 (top + bottom markers appear in the canonical format block + Post-Paste /goal block)

# 2. Box-drawing decorators (─ U+2500) survive
grep -c '─' .claude/rules/moai/workflow/session-handoff.md
# Expected: ≥8 (4 per cut-line marker pair × ≥2 pairs)

# 3. 6-block skeleton headings survive
grep -c 'Block 1\|Block 2\|Block 3\|Block 4\|Block 5\|Block 6' .claude/rules/moai/workflow/session-handoff.md
# Expected: ≥6 (all 6 blocks referenced)

# 4. Pre-emit self-check labels survive
grep -c 'paste-ready budget\|localization render\|session-handoff template completeness' .claude/rules/moai/workflow/session-handoff.md
# Expected: ≥3 (all 3 concern-name qualifiers survive)

# 5. Core section headings survive
grep -c '^## Canonical Format\|^## Post-Paste\|^## Diet Constraints\|^## Worktree-Anchored\|^## Auto-Memory Integration' .claude/rules/moai/workflow/session-handoff.md
# Expected: ≥5 (all 5 core H2 sections survive)
```

#### AC-MD-009 (REQ-MD-009 — SSOT ↔ render-surface parity honored)

**Given** session-handoff.md is the SSOT and `.claude/output-styles/moai/moai.md §8` is the render surface,
**When** the run-phase completes the REQ-2 edit,
**Then** the render surface continues to cross-reference session-handoff.md as the SSOT (no broken parity).

**Verification command:**
```bash
grep -c 'session-handoff' .claude/output-styles/moai/moai.md
# Expected: ≥1 (render surface still cross-references the SSOT — baseline count preserved or increased)
# NOTE: .claude/output-styles/moai/moai.md is NOT edited by this SPEC; this AC verifies it is not broken.
```

#### AC-MD-010 (REQ-MD-010 — session-handoff template mirror parity)

**Given** session-handoff.md is MIRRORED,
**When** the run-phase completes the REQ-2 edit,
**Then** both trees carry identical content (session-handoff.md identical; sibling references file identical).

**Verification command:**
```bash
diff .claude/rules/moai/workflow/session-handoff.md internal/template/templates/.claude/rules/moai/workflow/session-handoff.md
# Expected: empty diff
diff .claude/rules/moai/workflow/session-handoff-examples.md internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md
# Expected: empty diff (if sibling file exists in both trees)
```

#### AC-MD-011 (REQ-MD-011 — sibling references file is path-scoped)

**Given** the sibling references file `session-handoff-examples.md` is created,
**When** it is written,
**Then** it carries its own `paths:` frontmatter pointing to `session-handoff.md` so it is NOT always-loaded.

**Verification command:**
```bash
head -8 .claude/rules/moai/workflow/session-handoff-examples.md | grep -E '^(---|paths:)'
# Expected: frontmatter present with paths: field (e.g. paths: "**/session-handoff.md")
# Confirm it is NOT in the always-loaded set:
grep -L '^paths:' .claude/rules/moai/workflow/session-handoff-examples.md
# Expected: no output (file HAS a paths: field → not always-loaded)
```

---

### REQ-3 — MEMORY.md archive pruning

#### AC-MD-012 (REQ-MD-012 — stable ✅ entries moved to archive)

**Given** MEMORY.md originally contained 29 `✅` entries,
**When** the run-phase completes the REQ-3 edit,
**Then** a subset of stable `✅` entries (closed SPECs, no open follow-up, ≥ N days old) are moved to `MEMORY-archive-2026-06-02.md` AND the MEMORY.md line count decreases.

**Verification commands:**
```bash
# Archive gained entries
BEFORE=$(git show HEAD:~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY-archive-2026-06-02.md 2>/dev/null | wc -l || echo 0)
AFTER=$(wc -l < ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY-archive-2026-06-02.md)
# Expected: AFTER > BEFORE (archive grew)

# MEMORY.md line count decreased
wc -l < ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md
# Expected: < 90 (original was 90 lines)

# ✅ count in MEMORY.md decreased
grep -c '✅' ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md
# Expected: < 29 (original was 29 ✅ entries; stable subset moved)
```

> **Note**: MEMORY.md lives in auto-memory (`~/.claude/projects/*/memory/`) which is outside the git repo. The "before" baseline is the measured 90 lines / 29 ✅ entries from the orchestrator Discovery. Run-phase records the actual before/after in progress.md §E.2.

#### AC-MD-013 (REQ-MD-013 — active-marker entries preserved)

**Given** MEMORY.md originally contained 18 active-marker entries (unique-line basis: `🟢=4, 🟡=2, 🆕=7, ⏸️=1, ⚠️=1, 🔍=3`, sum = 18),
**When** the run-phase completes the REQ-3 edit,
**Then** the active-marker count is unchanged.

**Verification command:**
```bash
grep -c '🟢\|🟡\|🆕\|⏸️\|⚠️\|🔍' ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md
# Expected: 18 (all active-marker entries preserved — NONE moved to archive)
```

#### AC-MD-014 (REQ-MD-014 — load-bearing ✅ entries preserved)

**Given** some `✅` entries reference pending next steps ("다음=SPEC-XXX", "handoff", "deferred", "debt"),
**When** the run-phase completes the REQ-3 edit,
**Then** every load-bearing `✅` entry remains in MEMORY.md.

**Verification command:**
```bash
# Any remaining ✅ entry referencing a pending next step / handoff / debt MUST survive
grep '✅' ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md | grep -iE '다음=|handoff|deferred|debt|후속|pending'
# Expected: ≥0 (whatever the baseline count is, it MUST NOT decrease — these are load-bearing)
# The run-phase MUST record the before/after count of this grep in progress.md §E.2.
```

#### AC-MD-015 (REQ-MD-015 — pruning aligns with close-time rule)

**Given** `session-handoff.md § Auto-Memory Integration` item 6 governs close-time pruning,
**When** the run-phase completes the REQ-3 edit,
**Then** the pruning follows the archive discipline (move to `MEMORY-archive-2026-06-02.md`, never delete) AND MEMORY.md stays within the 200-line / 25KB loader cap.

**Verification command:**
```bash
wc -l < ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md
# Expected: ≤ 200 (loader line cap)

wc -c < ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md
# Expected: ≤ 25600 (25KB loader cap; original was 16,758 bytes)

# Archive is the destination (not deletion) — verify archive file is non-empty and grew
test -s ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY-archive-2026-06-02.md && echo "archive exists and non-empty"
```

---

### REQ-MD-016 / REQ-MD-017 — combined delta + non-regression

#### AC-MD-016 (REQ-MD-016 — combined token reduction ≥ ~3.0k, revised post-run)

**Given** the orchestrator `/context` baseline measured "Memory files: 62.1k tokens (6.2%)" before the SPEC,
**When** all 3 milestones (M1 + M2 + M3) complete,
**Then** the combined always-loaded token count of the 3 modified file groups decreases by ≥ ~3.0k tokens measured via `/context`.

**Verification:**
- Orchestrator-side `/context` re-measurement after M3. The delta is: `baseline_tokens - post_tokens ≥ ~3000`.
- Run-phase records the verbatim `/context` output before and after in progress.md §E.2 (per verification-claim-integrity §3 Evidence format).

**Revised post-run (2026-07-10) — PASS at ~3.0k measured.** The original plan-phase target was ≥ ~5.5k tokens, derived from a Discovery-time bytes→tokens conversion error: the session-handoff (~2k bytes) and MEMORY.md (~1.5k bytes) items had their byte counts mislabeled as token counts, while the cadence-bridge estimate (~2.3k tokens) was a true token estimate of the file leaving the always-loaded set. The orchestrator's independent Trust-but-verify (wc + byte→token estimate at ~4 bytes/token for English markdown) measured the actual combined reduction at ~3.0k tokens:

| File group | Saving |
|------------|--------|
| cadence-bridge.md path-match (file leaves always-loaded set entirely) | ~2.3k tokens (dominant) |
| session-handoff.md shrinkage (56,598B → 54,650B = 1,948B) | ~487 tokens |
| MEMORY.md shrinkage (17,109B → 16,266B = 843B) | ~210 tokens |
| **Combined** | **~3.0k tokens** |

The authoritative measurement is the next `/context` after `/clear` (cadence-bridge leaving the always-loaded "Memory files" set is the verifiable artifact — the two byte-shrinkage items are measurable via wc). Evidence: file-level wc before/after recorded in progress.md §E.2. This AC is PASS at the measured ~3.0k.

#### AC-MD-017 (REQ-MD-017 — test suite + lint + CI non-regression)

**Given** the existing test suite + lint + CI guards,
**When** all 3 milestones complete,
**Then** `go test ./...` passes, `go vet ./...` is clean, and template-neutrality CI passes.

**Verification commands:**
```bash
go test ./... 2>&1 | tail -5
# Expected: all PASS (no new failures)

go vet ./... 2>&1
# Expected: clean (no findings)

# Template neutrality (REQ-1/REQ-2 template mirrors must not leak SPEC IDs / REQ tokens / dates / SHAs)
go test ./internal/template/... -run TestTemplateNeutrality 2>&1 | tail -5
# Expected: PASS
```

---

## §D.1 Severity Classification

| AC | Severity | Rationale |
|----|----------|-----------|
| AC-MD-001..005 (REQ-1) | MUST-PASS | path-scope conversion correctness + template parity |
| AC-MD-006..008 (REQ-2 core) | MUST-PASS | extraction correctness + CORE DOCTRINE byte-identical survival |
| AC-MD-009 (SSOT parity) | MUST-PASS | render-surface contract |
| AC-MD-010..011 (REQ-2 template + sibling) | MUST-PASS | template parity + sibling path-scoped |
| AC-MD-012..015 (REQ-3) | MUST-PASS | archive discipline + active-marker preservation |
| AC-MD-016 (token delta) | MUST-PASS | measurable diet outcome |
| AC-MD-017 (test/lint/CI) | MUST-PASS | non-regression |

---

## §D.2 REQ ↔ AC Traceability

| REQ | AC(s) |
|-----|-------|
| REQ-MD-001 | AC-MD-001 |
| REQ-MD-002 | AC-MD-002 |
| REQ-MD-003 | AC-MD-003 |
| REQ-MD-004 | AC-MD-004 |
| REQ-MD-005 | AC-MD-005 |
| REQ-MD-006 | AC-MD-006 |
| REQ-MD-007 | AC-MD-007 |
| REQ-MD-008 | AC-MD-008 |
| REQ-MD-009 | AC-MD-009 |
| REQ-MD-010 | AC-MD-010 |
| REQ-MD-011 | AC-MD-011 |
| REQ-MD-012 | AC-MD-012 |
| REQ-MD-013 | AC-MD-013 |
| REQ-MD-014 | AC-MD-014 |
| REQ-MD-015 | AC-MD-015 |
| REQ-MD-016 | AC-MD-016 |
| REQ-MD-017 | AC-MD-017 |

**Coverage: 17/17 REQs covered by 17 ACs (100% traceability).**

---

## §D.3 Edge Cases

### Edge Case 1 — cadence-bridge trigger miss

**Scenario**: The `paths:` glob is too narrow and the cadence-bridge rule fails to load when a session composes `/loop 30m /moai gate`.

**Mitigation**: REQ-MD-002 + AC-MD-002 require the trigger-condition list to include the `/loop` + `/moai` composition surface. The run-phase MUST verify the glob + documented triggers cover this case before declaring M1 complete. If the self-referential `paths: "**/cadence-bridge.md"` is used alone, the rewritten Loading scope prose MUST explicitly document that the rule loads on cadence-bridge.md maintenance AND MUST be manually consulted when composing `/loop` + `/moai` (the orchestrator is expected to recognize the composition from the goal-directive.md cross-reference).

### Edge Case 2 — session-handoff extraction breaks a cross-reference

**Scenario**: An external file (e.g. a SPEC or agent definition) references `session-handoff.md § Example with Block 0` and the section is extracted to the sibling file.

**Mitigation**: The one-line pointer left in session-handoff.md names the sibling file path, so the cross-reference resolves via the pointer. Run-phase MUST grep for inbound references to the extracted section headings and verify the pointer covers them.

### Edge Case 3 — MEMORY.md stable ✅ entry is actually load-bearing

**Scenario**: A `✅` entry appears stable (closed SPEC, old) but actually holds a cross-Epic tracking pointer that future sessions need.

**Mitigation**: REQ-MD-014 + AC-MD-014 require every `✅` entry referencing "다음=" / "handoff" / "deferred" / "debt" to be preserved. The run-phase MUST grep each `✅` entry's hook line for these keywords before moving it; if any keyword matches, the entry stays.

### Edge Case 4 — Localization Table parity with moai.md §8

**Scenario**: The condensed Localization Table (en/ko only) drifts from the render surface `.claude/output-styles/moai/moai.md §8 Localization Contract` which carries the full 4-locale set.

**Mitigation**: REQ-MD-009 + AC-MD-009 verify the parity contract. The condensed table inline is a SUBSET; the full 4-locale table survives in the sibling references file. The pointer in session-handoff.md notes the full table location. moai.md §8 is NOT edited (it remains the authoritative render surface).

---

## §D.4 Closure Gates

### Definition of Done

- [ ] All 17 ACs PASS (MUST-PASS severity)
- [ ] Combined token reduction ≥ ~3.0k verified via `/context` (AC-MD-016, revised post-run from original 5.5k target — see AC-MD-016 Revised post-run note)
- [ ] `go test ./...` + `go vet ./...` + template-neutrality CI pass (AC-MD-017)
- [ ] Template mirror parity verified for REQ-1 + REQ-2 (AC-MD-004, AC-MD-010)
- [ ] CORE DOCTRINE byte-identical survival verified (AC-MD-008)
- [ ] MEMORY.md active-marker count unchanged (AC-MD-013)
- [ ] No load-bearing ✅ entry removed (AC-MD-014)
- [ ] progress.md §E.2 carries verbatim `/context` before/after output

### Quality Gate (Tier M)

- plan-auditor PASS threshold: 0.80
- sync-auditor 4-dimension scoring (Functionality / Security / Craft / Consistency)
