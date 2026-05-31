# Progress — SPEC-V3R6-DOCS-I18N-PARITY-001

## §E — Run-Phase Progress Tracking

### §E.1 Phase Routing

**Phase**: Run-Phase, docs-site 4-locale i18n parity clearance
**Agent**: manager-docs (direct author of this SPEC)
**Mode Selection**: sub-agent (trivial docs-only work, no code implementation)
**Cycle Type**: N/A (not applicable — documentation work, no DDD/TDD needed)
**Isolation**: Main checkout (no worktree)

### §E.2 Run-Phase Audit-Ready Signal

Baseline: 53 errors (4 frontmatter + 26 H1 + 23 glossary)

**Completion Evidence:**

- **M1 (C1: Frontmatter)** — 4 → 0 errors
  - Added YAML frontmatter (`title`, `weight`, `draft`) to docs-site/content/{ko,en,ja,zh}/cost-optimization/prompt-caching.md
  - Files modified: 4

- **M2 (C2: H1 headings Set A)** — 20 → 0 errors
  - Added H1 heading (`# <title>`) immediately after frontmatter for 5 files (advanced/harness-profiles, core-concepts/harness-engineering, db/migration-patterns, getting-started/profile, workflow-commands/moai-design) across 4 locales
  - Files modified: 20

- **M2+M3 (C2+C3: H1 + glossary for multi-llm stubs)** — 6 H1 + glossary errors → 0
  - Added H1 heading to en/ja/zh multi-llm/cg-mode.md (3 files)
  - Added H1 heading to en/ja/zh multi-llm/model-policy.md (3 files)
  - Added invariant glossary terms (`Claude Code`, `MoAI-ADK`) to stub content
  - Files modified: 6

- **M3 (C3: Glossary parity)** — 23 → 0 errors
  - Added `Claude Code` glossary term to en/ja/zh multi-llm/_index.md (3 files)
  - Added `MoAI-ADK` + `moai-adk` glossary terms to en/ja/zh contributing/_index.md (3 files)
  - Added `SPEC-First` glossary term to zh/getting-started/introduction.md (1 file)
  - Added `SPEC-First` to zh/getting-started/update.md (1 file)
  - Files modified: 8

- **M4 (Green gate)** — Verified both modes pass:
  - `DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh` → `Errors: 0, Warnings: 0, OK: all 4 locales pass`
  - `bash scripts/docs-i18n-check.sh` (strict mode) → exit 0

**Total files modified: 39 across all 4 locales (ko, en, ja, zh)**

### §E.3 Acceptance Criteria Verification

| AC | Status | Evidence |
|-----|--------|----------|
| AC-DIP-001 | **PASS** | C1 frontmatter check: 0 errors |
| AC-DIP-002 | **PASS** | C2 Set A H1 check: 0 errors |
| AC-DIP-003 | **PASS** | C2 Set B H1 check: 0 errors |
| AC-DIP-004 | **PASS** | C3 glossary check: 0 errors |
| AC-DIP-005 | **PASS** | C3 SPEC-First zh check: 2 files present |
| AC-DIP-006 | **PASS** | Warn-mode check: `Errors: 0` |
| AC-DIP-007 | **PASS** | Strict-mode check: exit 0 |
| AC-DIP-008 | **PASS** | File parity: 78 .md per locale, 0 orphan lines |
| AC-DIP-009 | **PASS** | Regression guard: 0 forbidden URLs, canonical URL present, Mermaid non-TD count = 4 (baseline) |

**All 9 ACs: PASS**

### §E.4 Phase Closure Status

- **Status**: `in-progress` (run-phase completed, awaiting sync-phase)
- **Run-phase commit**: Prepared (staged: 39 files, awaiting git commit)
- **Blocker count**: 0
- **Known issues**: None

---

## §D Run-Phase Scope Compliance

✓ No Go code modifications (docs-only)
✓ No template source changes (template mirror rules preserved)
✓ No new .md files created (file-path parity maintained at 78 per locale)
✓ No forbidden URLs introduced (AC-DIP-009 verified)
✓ No Mermaid non-TD diagrams added (count held at baseline 4)
✓ Glossary terms preserved as invariants (case-sensitive, no translation)

---

**Run-phase ready for commit.**
