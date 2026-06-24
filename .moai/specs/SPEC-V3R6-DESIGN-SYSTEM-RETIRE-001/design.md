# design.md — Retirement Architecture Decisions

> Records the per-file retirement-treatment decision (delete vs migrate vs
> preserve-as-historical) and the invariants that bind the retirement.

## §D1. Decision: deletion, not migration

**Decision**: The `moai-design-system` skill is **deleted**, not migrated
to a surviving skill.

**Alternatives considered**:
1. **Delete** (chosen) — the skill's functional content is largely covered
   by `moai-domain-brand-design` + `moai-domain-frontend` + Context7 MCP
   lookups. The 3 residual gaps (Style Dictionary, icon-set selection, UX
   writing — research.md §R2.2) are narrow tool-selection topics that no
   agent was reaching anyway (phantom status, research.md §R1).
2. **Migrate content into `moai-domain-brand-design`** — rejected. The
   brand-design skill already over-covers the design-system scope (per
   `project_skill_audit_followup_wip` memory line 16). Migration would
   duplicate content already present.
3. **Migrate content into a new `moai-domain-ux-writing` skill** —
   rejected. The 3 residual gaps do not justify a new skill; UX writing is
   a narrow topic better served by Context7 lookups against e.g.
   Material Design / Apple HIG docs.
4. **Reassign to a different agent's preload list** — rejected. The
   phantom-status root cause is that no agent's workflow needs the skill.
   Reassigning preloads would paper over the symptom without addressing
   the overlap.

**Rationale**: the R4 audit's design-cluster merge (P-S13) produced a
skill that was immediately redundant with the brand-design + frontend
skills it was supposed to complement. Retirement is the honest correction.

## §D2. Per-file treatment decision table

| File category | Treatment | Rationale |
|---------------|-----------|-----------|
| `.claude/skills/moai-design-system/SKILL.md` (local) | DELETE | The skill body; deletion is the retirement. Already staged via stash-pop. |
| `internal/template/templates/.claude/skills/moai-design-system/SKILL.md` (template) | DELETE | Template-neutrality §15/§25 requires symmetric deletion. Already staged. |
| Skill directories (both trees) | RMDIR (empty) | Empty skill directories are orphans; CLAUDE.local.md §2 protected-directory hygiene. |
| `internal/cli/doctor_skills.go` allowlist literal | REMOVE LITERAL + REMOVE `// design (1)` comment | Static core-skill list must not reference a deleted skill. The category-comment becomes stale when the only entry is removed. |
| `internal/cli/doctor_skills_test.go` test case | REMOVE 4-line struct block | Test referencing a non-existent static core skill is a dead test. |
| `internal/template/catalog.yaml` entry block | REMOVE 5-line block | Catalog must not list a deleted skill; downstream `moai init` / `moai update` would otherwise emit a broken skill directory. |
| `internal/design/dtcg/frozen_guard_test.go` allowedPaths (×2) | REMOVE path literal from BOTH slices | The frozen-path guard whitelists paths that are allowed to be written despite DTCG FROZEN enforcement. A non-existent path in the whitelist is dead fixture; removing it keeps the test honest. |
| `moai-workflow-design/SKILL.md` description | REMOVE `(see moai-design-system)` parenthetical ONLY | The pointer dangles after retirement. The rest of the description is preserved verbatim — surgical edit, no rewrite. |
| `docs-site/content/{en,ko,ja,zh}/advanced/skill-guide.md` | REMOVE table row + UPDATE section header consistently across 4 locales | 4-locale parity invariant. The section header `### Design (Design System) - 1 skill` becomes either removed or `- 0 skills` with a note; the SAME treatment applies to all 4 locales. |

## §D3. Invariants binding the retirement

### §D3.1 Template-neutrality symmetry (CLAUDE.local.md §15, §25)

Every file category that has BOTH a local and a template instance MUST be
edited symmetrically. The stash-pop seed already satisfies this for the 2
SKILL.md files. The remaining categories are single-instance (Go files,
catalog.yaml, docs-site are each in one tree only), so the symmetry
invariant reduces to: the SKILL.md deletion must remain symmetric at
commit time.

Run-phase verification: `diff <(git show :.claude/skills/moai-design-system/SKILL.md 2>/dev/null) <(git show :internal/template/templates/.claude/skills/moai-design-system/SKILL.md 2>/dev/null)` — both must be absent (deleted) symmetrically.

### §D3.2 Harness namespace alignment (CLAUDE.local.md §24)

`moai-design-system` is a `moai-*` skill = template-managed. Retirement
(deletion) is the correct treatment per §24.4's delete-vs-preserve matrix
— the `moai-*` namespace is sync-overwrite, so deleting from the template
source propagates correctly on `moai update`. There is no
`harness-*`/user-owned dimension to this skill.

### §D3.3 spec-lint clean

`spec.md` passes `moai spec lint` with 0 findings. The §C Out-of-Scope
section uses the `### Out of Scope — <topic>` H3 convention per
`SPEC-V3R6-OUTOFSCOPE-GUIDANCE-ALIGN-001` to satisfy `OutOfScopeRule`.

### §D3.4 Historical-reference immutability

Files listed in §C "Out of Scope — Historical references" (CHANGELOG,
.moai/specs/SPEC-V3R{2,3,4,5}-*/, .moai/research/**, .moai/archive/**,
.moai/release/**, .moai/decisions/**, .moai/brain/**, .moai/state/**,
.moai/backups/**, .moai/design/v3-redesign/**, .claude/worktrees/**) are
records-of-the-past and MUST NOT be modified. The
`moai-design-craft/RETIRED.md` Substitute pointer
(`Substitute: moai-design-system`) is explicitly preserved — RETIRED.md
files are historical "where did this skill go" records and remain valid
even after the substitute itself retires.

## §D4. Anti-pattern catalogue

- **AP-DSR-001 — Migrating content instead of deleting**: rejected at §D1.
  Migration duplicates already-present coverage.
- **AP-DSR-002 — Editing only the local OR only the template SKILL.md**:
  violates template-neutrality §15/§25. The stash-pop seed did both
  correctly; run-phase must preserve both deletions at commit.
- **AP-DSR-003 — Rewriting the `moai-workflow-design` description prose
  beyond the parenthetical**: scope creep. Only the `(see moai-design-system)`
  pointer is removed.
- **AP-DSR-004 — Modifying CHANGELOG or RETIRED.md to scrub
  `moai-design-system` mentions**: violates historical-reference
  immutability (§D3.4).
- **AP-DSR-005 — Inconsistent 4-locale docs-site treatment** (e.g.
  removing the row in EN but only rewriting the header in KO): violates
  the 4-locale parity invariant (REQ-DSR-007).
- **AP-DSR-006 — Leaving the empty `// design (1)` comment** in
  doctor_skills.go after removing the literal: stale comment referencing a
  now-empty category. Remove the comment line alongside the literal.

## §D5. Tier classification rationale

**Tier L** (confirmed). The orchestrator's pre-classification was Tier L;
this author confirms. Justification against the Tier table in
`.claude/rules/moai/workflow/spec-workflow.md`:

- **8 file categories** (skill files ×2 trees, skill dirs ×2, Go allowlist,
  Go test, catalog, frozen-guard test ×2 sites, cross-skill ref, docs-site
  ×4 locales) — exceeds the Tier M file-count threshold.
- **Go code changes** in `internal/cli/` and `internal/design/dtcg/` —
  Tier L territory (not just doc/skill edits).
- **Frozen-guard test fixture** — touching the DTCG FROZEN enforcement
  surface elevates risk; the test must continue to pass with the fixture
  updated.
- **4-locale docs-site parity** — cross-locale consistency is a Tier L
  concern (per the docs-site i18n rules doctrine, CLAUDE.local.md §17).
- **Template-neutrality symmetry** — local + template tree coordination.

Tier L triggers the 5-artifact structure (this design.md + research.md +
the canonical 3) and the thorough harness level.

## §D6. Forward-links (out of scope, recorded for traceability)

- A follow-up SPEC (not authored here) MAY add a retired-skill redirect
  mechanism so manual `Skill("moai-design-system")` invocations produce a
  helpful deprecation message instead of a bare "not found". Accepted as
  residual risk in research.md §R3.2.
- A follow-up catalog-reshuffling SPEC MAY reorganize the `design` pack
  after this retirement leaves it sparse. Out of scope per §C.
