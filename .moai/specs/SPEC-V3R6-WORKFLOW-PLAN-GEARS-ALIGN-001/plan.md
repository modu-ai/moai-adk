---
id: SPEC-V3R6-WORKFLOW-PLAN-GEARS-ALIGN-001
artifact: plan
version: "0.1.3"
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
plan_commit_sha: "27afbca1e"
sync_commit_sha: "bd52b70e5"
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-25 | manager-spec | Initial plan-phase draft — Sprint 10 GEARS sweep cohort entry SPEC #4 of 8. Tier M, 1-pass run-phase target, 6-milestone decomposition (M1-M6), 4 local + 4 mirror = 8 .md files in scope, 13 edit zones (4 in plan.md + 3 in clarity-interview.md + 6 in spec-assembly.md + 0 in context-discovery.md). |
| 0.1.1 | 2026-05-25 | manager-spec | iter-2 focused fix per plan-auditor iter-1. M4 staging count 11→10 (Option A — aligned with acceptance.md AC-WPG-010). spec-assembly.md edit zones 6→7 (cross-link addition to spec-frontmatter-schema.md SSOT counts as edit zone). Total edit zones 13→14. AC count 10→11. Predicted iter-2 plan-auditor: 0.90+ skip-eligible. |
| 0.1.2 | 2026-05-25 | manager-spec | iter-3 mechanical fix per plan-auditor iter-2 PASS-WITH-DEBT 0.873. D_new1 RESOLVED: §B.1 "13 edit zones" → "14 edit zones" + §B.3 MP-1 "13 edit zones counted" → "14 edit zones counted" (residual stale counts from iter-1 incomplete patch). D_new2 RESOLVED: §B.3 MP-3 "13 REQs × 10 ACs" → "13 REQs × 11 ACs". D_new3 RESOLVED: HISTORY table added (Option A — consistency with spec.md and acceptance.md). Predicted iter-3 plan-auditor: 0.92+ skip-eligible (Consistency 0.74→0.92 + Completeness 0.92→0.94). |
| 0.1.3 | 2026-05-25 | manager-docs | Sync-phase completion: plan_commit_sha backfilled to `27afbca1e` (iter-3 final plan), sync_commit_sha field added (pending backfill post-commit). HISTORY entry added. |

## §A — Lifecycle Table

| Phase | Milestone | Owner | Audit-ready signal | Commit SHA |
|-------|-----------|-------|--------------------|------------|
| Plan | M0 (artifact creation) | manager-spec | spec.md + plan.md + acceptance.md + progress.md committed | `<pending>` |
| Plan | Phase 0.5 (plan-auditor verification) | orchestrator | plan-auditor independent audit ≥ 0.85 PASS (target 0.90 skip-eligible) | (verification only) |
| Run | M1 (local file notation edits, 4 files) | manager-develop | 4 local files edited with GEARS-first + EARS legacy footnote pattern | `<pending>` |
| Run | M2 (template mirror parity sync, 4 files) | manager-develop | mirror files byte-for-byte identical to local files via `diff -q` | `<pending>` |
| Run | M3 (sentinel verification + lint regression check) | manager-develop | `grep -E 'IF .* THEN'` 0 NEW occurrences + `moai spec lint` 0 regressions on 88 pre-v3 SPECs | `<pending>` |
| Run | M4 (status transition + AC verification) | manager-develop | spec.md frontmatter `status: draft → in-progress` + 11/11 AC PASS | `<pending>` |
| Sync | M5 (manager-docs sync-phase emission) | manager-docs | sync_commit_sha 4-artifact backfill + CHANGELOG entry + status `in-progress → implemented` | `<pending>` |
| Mx | M6 (orchestrator Mx Step C judge) | orchestrator | Mx Step C SKIP-eligible per mx-tag-protocol.md §a (markdown-only edits, 0 .go files, 0 @MX delta) + 4-phase close marker | `<pending>` |

## §B — Run-phase Strategy

Tier M minimal 1-pass strategy targeting plan-auditor 0.87+ PASS, ideal skip-eligible 0.90+. Single manager-develop spawn delegation covering Section A-E (M1-M4). Phase 0.5 plan-auditor independent verification gate per CONST-V3R5-026 self-audit. Sync-phase manager-docs spawn (M5). Mx-phase orchestrator judge (M6).

### B.1 Section A-E delegation envelope

manager-develop receives a single spawn prompt enumerating:
- Section A: file inventory (8 exact paths: 4 local + 4 mirror)
- Section B: 14 edit zones (4 in plan.md, 3 in clarity-interview.md, 7 in spec-assembly.md, 0 in context-discovery.md)
- Section C: 13 REQs (REQ-WPG-001..013) with GEARS-first phrasing patterns
- Section D: 11 mandatory ACs (AC-WPG-001..011) with traceability matrix (iter-2 added AC-WPG-011 closing REQ-WPG-009 trace orphan)
- Section E: verification batch (mirror parity `diff -q`, sentinel `grep -E 'IF .* THEN'`, lint regression, frontmatter status transition)

### B.2 Cohort precedent fidelity

Predecessor pattern observed from FOUNDATION-CORE-GEARS-ALIGN-001 `0156c7003` (Tier M, 1-pass run-phase, 6 commits, scope expansion 10→20 via D1 discovery): run-phase may discover scope variance during M1; if discovered, manager-develop emits a structured Audit-Ready Signal update + AskUserQuestion blocker report rather than silent scope creep. This SPEC anticipates 4 local + 4 mirror = 8 files as fixed; no scope expansion expected (the 4 .md files are exhaustively enumerated in §C.1).

### B.3 Plan-auditor skip-eligible projection

| Dimension | Target | Rationale |
|-----------|--------|-----------|
| MP-1 Scope clarity | ≥ 0.90 | 8 files explicitly enumerated; 14 edit zones counted; no ambiguity |
| MP-2 GEARS/EARS notation rigor | ≥ 0.90 | 100% self-dogfood (13/13 REQs in GEARS form); compound + capability-gate present |
| MP-3 Traceability matrix | ≥ 0.90 | 13 REQs × 11 ACs explicit mapping in acceptance.md |
| MP-4 Risk-mitigation pairing | ≥ 0.85 | 5 risks identified, each with explicit mitigation referencing milestone/AC |
| Weighted overall | ≥ 0.87 | Aim ≥ 0.90 for skip-eligible |

## §C — Scope (Honest Discovered Inventory)

### §C.1 Discovered file inventory (4 local + 4 mirror = 8 .md files)

**Local files** (full paths):
```
/Users/goos/MoAI/moai-adk-go/.claude/skills/moai/workflows/plan.md
/Users/goos/MoAI/moai-adk-go/.claude/skills/moai/workflows/plan/clarity-interview.md
/Users/goos/MoAI/moai-adk-go/.claude/skills/moai/workflows/plan/context-discovery.md
/Users/goos/MoAI/moai-adk-go/.claude/skills/moai/workflows/plan/spec-assembly.md
```

**Template mirror files** (full paths):
```
/Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/skills/moai/workflows/plan.md
/Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/skills/moai/workflows/plan/clarity-interview.md
/Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/skills/moai/workflows/plan/context-discovery.md
/Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md
```

**Variance from estimate** (transparently documented per L46 attribution discipline):

| Source | Count | Notes |
|--------|-------|-------|
| Paste-ready estimate | 9 files | Likely included `.gitkeep` placeholder in template mirror |
| Actual discovery | 8 .md files + 1 .gitkeep | `.gitkeep` is 0-byte placeholder in template mirror only |
| Effective edit scope | 8 .md files | `.gitkeep` preserved untouched, no notation content |
| Variance | -1 file (-11%) | Below estimate; honest documentation per discovery |

`diff -q` baseline (pre-edit state — 2026-05-25):
- `plan.md` local == template mirror (identical 7,257 bytes)
- `plan/` directory differs only by `.gitkeep` (template-only, presentational)
- All 3 plan/ sub-files (`clarity-interview.md`, `context-discovery.md`, `spec-assembly.md`) byte-identical local vs mirror

### §C.2 Notation reference distribution (13 EARS refs / 14 edit zones)

| File | Line ranges | EARS refs | Edit zones |
|------|-------------|-----------|------------|
| `plan.md` | Lines 4, 37, 54, 62 | 4 | 4 (description block + intro + 2 phase routing table cells) |
| `clarity-interview.md` | Lines 169, 175, 184 | 3 | 3 (Phase 1B agent task + output description + transition header) |
| `spec-assembly.md` | Lines 73, 138, 260, 418, 506, 530 | 6 | 7 (Phase 2 frontmatter intro + frontmatter checklist + traceability summary + quality gate + completion check + Phase 1B JWT example + NEW cross-link to spec-frontmatter-schema.md SSOT per REQ-WPG-009) |
| `context-discovery.md` | (none) | 0 | 0 (mirror parity only) |
| **Total** | | **13** | **14** |

### §C.3 Out-of-scope (preserved verbatim — L48 SSOT discipline)

- All 88 pre-v3 SPEC bodies in `.moai/specs/`
- All predecessor SPEC bodies (GEARS-MIGRATION-001, SKILL-GEARS-ALIGN-001, PLAN-AUDITOR-GEARS-ALIGN-001, FOUNDATION-CORE-GEARS-ALIGN-001)
- `.claude/skills/moai-foundation-core/*` (closed `0156c7003`)
- `.claude/skills/moai-workflow-spec/*` (closed `ebe492670`)
- `.claude/agents/meta/plan-auditor.md` (closed `ebe492670`)
- `internal/spec/lint.go` and all `lint_*_test.go` files
- `.claude/skills/moai/team/plan.md` (team-mode plan, separate scope)
- All other `.claude/skills/moai/workflows/*` files (run.md, sync.md, mx.md, loop.md, fix.md, clean.md, feedback.md, project.md, design.md)

## §D — Milestone Decomposition

### M1 — Local file GEARS-first edits (4 files)

Path-specific scope:
- `.claude/skills/moai/workflows/plan.md` — 4 edit zones (description frontmatter, intro para, 2 phase routing table cells)
- `.claude/skills/moai/workflows/plan/clarity-interview.md` — 3 edit zones (Phase 1B agent task, Phase 1B output description, transition header to spec-assembly)
- `.claude/skills/moai/workflows/plan/spec-assembly.md` — 7 edit zones (Phase 2 intro, frontmatter checklist intro, REQ format description, AC traceability summary, quality gate criteria, JWT example + NEW cross-link to spec-frontmatter-schema.md SSOT per REQ-WPG-009)
- `.claude/skills/moai/workflows/plan/context-discovery.md` — no edits (mirror parity only)

Edit pattern (GEARS-first with EARS legacy footnote):
- Replace verbatim "EARS format" → "GEARS notation (EARS retained as legacy reference, 6-month backward-compat window)"
- Replace verbatim "EARS structure with all 5 requirement types" → "GEARS structure with the 5 GEARS patterns (Ubiquitous, Event-driven `When`, State-driven `While`, Capability-gate `Where`, Event-detected unwanted — see `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format)"
- Replace verbatim "EARS ↔ AC coverage" → "GEARS ↔ AC coverage"
- Replace verbatim "EARS-format requirements" → "GEARS-notation requirements (EARS legacy form accepted for pre-v3 SPECs until 2026-11-22)"
- Add cross-link footnote to `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format at first GEARS reference in each file
- In `spec-assembly.md` Phase 2 SPEC document creation section, add explicit cross-link to `.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12 Required Fields (REQ-WPG-009: cross-link instead of inline 12-field restatement to reduce drift risk per §22 schema SSOT)

Verifies REQ-WPG-001, REQ-WPG-002, REQ-WPG-003, REQ-WPG-004, REQ-WPG-005, REQ-WPG-006, REQ-WPG-008, REQ-WPG-009, REQ-WPG-013.

### M2 — Template mirror parity sync (4 files)

For each of the 4 local files modified in M1, propagate the exact byte changes to the template mirror counterpart:

```
.claude/skills/moai/workflows/plan.md
  → internal/template/templates/.claude/skills/moai/workflows/plan.md

.claude/skills/moai/workflows/plan/clarity-interview.md
  → internal/template/templates/.claude/skills/moai/workflows/plan/clarity-interview.md

.claude/skills/moai/workflows/plan/context-discovery.md
  → internal/template/templates/.claude/skills/moai/workflows/plan/context-discovery.md

.claude/skills/moai/workflows/plan/spec-assembly.md
  → internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md
```

Post-edit verification:
```bash
diff -q .claude/skills/moai/workflows/plan.md internal/template/templates/.claude/skills/moai/workflows/plan.md
diff -r .claude/skills/moai/workflows/plan/ internal/template/templates/.claude/skills/moai/workflows/plan/
```

Expected: only difference is `.gitkeep` (template-mirror-only placeholder).

Verifies REQ-WPG-007, REQ-WPG-012.

### M3 — Sentinel verification + lint regression check

Post-M2 verification batch (orchestrator-parallel multi-Bash per agent-common-protocol.md §Parallel Execution):

```bash
# Sentinel 1: No NEW IF/THEN modality introductions in modified files
grep -rE 'IF .* THEN' .claude/skills/moai/workflows/plan.md .claude/skills/moai/workflows/plan/

# Sentinel 2: No NEW IF/THEN in template mirrors
grep -rE 'IF .* THEN' internal/template/templates/.claude/skills/moai/workflows/plan.md internal/template/templates/.claude/skills/moai/workflows/plan/

# Sentinel 3: GEARS notation present in all 3 content files
grep -c 'GEARS' .claude/skills/moai/workflows/plan.md .claude/skills/moai/workflows/plan/clarity-interview.md .claude/skills/moai/workflows/plan/spec-assembly.md

# Sentinel 4: EARS legacy reference retained (REQ-WPG-008)
grep -c 'EARS' .claude/skills/moai/workflows/plan.md .claude/skills/moai/workflows/plan/clarity-interview.md .claude/skills/moai/workflows/plan/spec-assembly.md

# Sentinel 5: Lint regression check on 88 pre-v3 SPECs
go run ./cmd/moai spec lint --json 2>&1 | jq '[.findings[] | select(.rule == "LegacyEARSKeyword")] | length'
```

Expected: Sentinels 1+2 = 0 (no IF/THEN), Sentinel 3 ≥ 3 (GEARS present in each), Sentinel 4 ≥ 3 (EARS legacy retained per REQ-WPG-008), Sentinel 5 lint count matches pre-edit baseline (no regression).

Verifies REQ-WPG-008, REQ-WPG-010, REQ-WPG-011, REQ-WPG-012.

### M4 — Status transition + AC final verification

Per Status Transition Ownership Matrix (`spec-frontmatter-schema.md` § Status Transition Ownership Matrix), manager-develop owns `draft → in-progress` transition. M4 actions:
1. Update spec.md frontmatter `status: draft → in-progress` (frontmatter-only edit, body content untouched per L48 SSOT)
2. Update spec.md frontmatter `updated: <today-ISO>`
3. Append progress.md §E.2 Run-phase Audit-Ready Signal with M1+M2+M3 verification evidence
4. Verify all 11 mandatory ACs (AC-WPG-001..011) PASS via inline assertion table in progress.md §E.2
5. Pre-commit staging assertion (L59): `git diff --cached --name-only | sort -u | wc -l` must equal exactly 10 (8 .md files: 4 local + 4 mirror + spec.md frontmatter + progress.md). If M1+M2+M3+M4 split across multiple commits, each commit asserts its own subset count.

Verifies REQ-WPG-007, REQ-WPG-009, REQ-WPG-012, REQ-WPG-013.

### M5 — Sync-phase manager-docs emission

Per Status Transition Ownership Matrix, manager-docs owns `in-progress → implemented` transition. M5 actions:
1. Update spec.md + plan.md + acceptance.md + progress.md frontmatter `sync_commit_sha: <commit-sha>` (4-artifact atomic backfill per L60)
2. Update spec.md frontmatter `status: in-progress → implemented`
3. Update spec.md frontmatter `updated: <today-ISO>`
4. CHANGELOG entry under v0.2.0 or appropriate version section
5. progress.md §E.4 Sync-phase Audit-Ready Signal with sync_commit_sha + CHANGELOG line citation

### M6 — Mx-phase orchestrator Step C judge

Per mx-tag-protocol.md §a (Mx Step C decision rubric):
- This SPEC modifies 0 .go files (markdown-only edits)
- @MX:NOTE / @MX:WARN / @MX:ANCHOR / @MX:TODO delta = 0
- 0 goroutines introduced
- 0 fan_in ≥ 3 changes

Expected verdict: **Mx Step C SKIP-eligible** (markdown-only notation alignment, no code surface area change).

Orchestrator emits chore commit:
```
chore(SPEC-V3R6-WORKFLOW-PLAN-GEARS-ALIGN-001): Mx-phase audit-ready signal + 4-phase close
```

With progress.md §E.5 Mx-phase Audit-Ready Signal documenting SKIP-eligible verdict and 4-phase close marker.

## §E — Verification Strategy

### E.1 Self-verification 7-item batch (orchestrator-parallel)

```bash
# V1: Mirror parity (4 file pairs)
diff -q .claude/skills/moai/workflows/plan.md internal/template/templates/.claude/skills/moai/workflows/plan.md
diff -r .claude/skills/moai/workflows/plan/ internal/template/templates/.claude/skills/moai/workflows/plan/ | grep -v "^Only in .*: \.gitkeep$"

# V2: GEARS notation present
grep -c 'GEARS' .claude/skills/moai/workflows/plan.md .claude/skills/moai/workflows/plan/clarity-interview.md .claude/skills/moai/workflows/plan/spec-assembly.md

# V3: EARS legacy retained (REQ-WPG-008)
grep -c 'EARS' .claude/skills/moai/workflows/plan.md .claude/skills/moai/workflows/plan/clarity-interview.md .claude/skills/moai/workflows/plan/spec-assembly.md

# V4: No IF/THEN deprecated modality introduced
grep -rE 'IF .* THEN' .claude/skills/moai/workflows/plan.md .claude/skills/moai/workflows/plan/ internal/template/templates/.claude/skills/moai/workflows/plan.md internal/template/templates/.claude/skills/moai/workflows/plan/

# V5: spec.md frontmatter status transition (post-M4)
grep -E '^status:' .moai/specs/SPEC-V3R6-WORKFLOW-PLAN-GEARS-ALIGN-001/spec.md

# V6: Git commit attribution (M1+M2+M3+M4 single commit OR M1 / M2 / M3+M4 split)
git log --oneline -5 -- .claude/skills/moai/workflows/plan.md .claude/skills/moai/workflows/plan/

# V7: Lint regression (no new LegacyEARSKeyword findings)
go run ./cmd/moai spec lint --json 2>&1 | jq '.findings[] | select(.rule == "LegacyEARSKeyword") | .file' | sort -u | wc -l
```

### E.2 Anti-patterns to avoid (L46/L48/L59 discipline)

- **Anti-pattern (L48 SSOT)**: Modifying `.claude/skills/moai-foundation-core/*` or `.claude/skills/moai-workflow-spec/*` files (predecessor SPEC territory). These are out of scope.
- **Anti-pattern (L46 attribution)**: Using `git add -A` or `git add .` in M1-M4. Path-specific add required (8 exact paths).
- **Anti-pattern (L59 staging scope)**: Pre-commit staging area containing files beyond the 8 explicit paths + spec.md + progress.md. Use `git reset` + path-specific re-add if drift detected.
- **Anti-pattern (REQ-WPG-008)**: Deleting all EARS references from modified files. EARS retention is mandated by the 6-month backward-compat window.
- **Anti-pattern (REQ-WPG-011)**: Re-introducing `IF/THEN` deprecated modality during GEARS rewrite.
- **Anti-pattern (cohort cross-attribution)**: Cross-commit contamination from parallel Sprint 10 cohort sessions. Pre-spawn fetch `git fetch origin && git rev-list --count --left-right origin/main...HEAD` must show `0 0` before M1 spawn.

## §F — Cross-References

- `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format (canonical SSOT)
- `.claude/skills/moai-foundation-core/SKILL.md` § GEARS Format (closed by FOUNDATION-CORE-GEARS-ALIGN-001 `0156c7003`)
- `.claude/agents/meta/plan-auditor.md` § MP-2 (closed by PLAN-AUDITOR-GEARS-ALIGN-001 `ebe492670`)
- `SPEC-V3R6-GEARS-MIGRATION-001` v0.2.0 (PR #1046) — canonical lint + backward-compat window
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix
- `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution (7-item verification batch pattern)
- `.claude/rules/moai/workflow/mx-tag-protocol.md` § Step C judge rubric
