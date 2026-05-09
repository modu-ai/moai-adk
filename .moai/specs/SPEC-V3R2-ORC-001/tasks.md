---
spec_id: SPEC-V3R2-ORC-001
phase: "1B — Tasks"
created_at: 2026-05-09
author: manager-spec
total_tasks: 24
---

# Tasks: Agent roster consolidation (22 → 17)

Discrete task breakdown for SPEC-V3R2-ORC-001 plan execution. Each task has
an ID `T-ORC001-NN`, milestone owner, dependency declarations, and a clear
acceptance handle pointing to acceptance.md.

Task naming: `T-ORC001-NN` (zero-padded 2-digit, no dashes within NN).

---

## M1 — Verification + Delivery Contract Definition (P0)

### T-ORC001-01 — Re-verify manager-cycle carry-over

- **Owner**: manager-spec
- **Depends on**: (none — entry task)
- **Action**: Read `manager-cycle.md` (template) and confirm the four
  citation rows from research.md §3.1 (L17, L54, L60-66, L8-15).
- **Output**: notation in plan progress (M1.1 ✓ checkmark).
- **AC handle**: AC-ORC-001-02.

### T-ORC001-02 — Re-verify manager-ddd / manager-tdd stubs

- **Owner**: manager-spec
- **Depends on**: T-ORC001-01
- **Action**: Read both stub files; confirm `retired: true`,
  `retired_replacement: manager-cycle`, `tools: []`, `skills: []`.
- **Output**: notation in progress.md.
- **AC handle**: AC-ORC-001-05 (subset).

### T-ORC001-03 — Capture diff baseline

- **Owner**: expert-backend
- **Depends on**: T-ORC001-02
- **Action**: Run `diff -r internal/template/templates/.claude/agents/moai/
  .claude/agents/moai/ > .moai/specs/SPEC-V3R2-ORC-001/diff-baseline.txt`.
- **Output**: artefact file capturing the 5 known divergences (see
  research.md §1).
- **AC handle**: AC-ORC-001-06 (informational baseline).

### T-ORC001-04 — Capture trigger snapshot for source bodies

- **Owner**: expert-backend
- **Depends on**: T-ORC001-03
- **Action**: Run extraction script over the 5 retiring source bodies
  (builder-agent, builder-skill, builder-plugin, expert-debug,
  expert-testing) and save to
  `.moai/specs/SPEC-V3R2-ORC-001/trigger-source-snapshot.txt`.
- **Output**: snapshot file referenced by AC-12 verification (snapshot
  needed BEFORE bodies are replaced with stubs in M2).
- **AC handle**: AC-ORC-001-12 (prerequisite).

### T-ORC001-05 — Resolve OQ-1..5 in HISTORY

- **Owner**: manager-spec
- **Depends on**: T-ORC001-04
- **Action**: Add HISTORY entry to spec.md (version 0.1.1) noting the OQ
  resolutions.
- **Output**: spec.md HISTORY updated.
- **AC handle**: (none direct; documentation hygiene).

---

## M2 — Retire 5 Agents + Create builder-platform (P0)

### T-ORC001-06 — Author builder-platform.md

- **Owner**: expert-backend
- **Depends on**: T-ORC001-04 (snapshot needed for trigger union)
- **Action**: Create
  `internal/template/templates/.claude/agents/moai/builder-platform.md` per
  plan §2.2 M2.1. Include `artifact_type` enum, 5-phase workflow, trigger
  union (deduped) from snapshot.
- **Output**: NEW file ~140-180 lines.
- **AC handle**: AC-ORC-001-03, AC-ORC-001-12.

### T-ORC001-07 — Retire builder-agent body to stub

- **Owner**: expert-backend
- **Depends on**: T-ORC001-06
- **Action**: Replace
  `internal/template/templates/.claude/agents/moai/builder-agent.md` body
  with retired-stub schema (`retired: true`, `retired_replacement:
  builder-platform`, `retired_param_hint: "artifact_type=agent"`, `tools:
  []`, `skills: []`); body = 1-line redirect + Migration Guide table.
- **Output**: file size shrinks from 116 → ~40 lines.
- **AC handle**: AC-ORC-001-05, AC-ORC-001-13.

### T-ORC001-08 — Retire builder-skill body to stub

- **Owner**: expert-backend
- **Depends on**: T-ORC001-07
- **Action**: Same as T-ORC001-07 with `retired_param_hint:
  "artifact_type=skill"`.
- **Output**: file size 105 → ~40 lines.
- **AC handle**: AC-ORC-001-05, AC-ORC-001-13.

### T-ORC001-09 — Retire builder-plugin body to stub

- **Owner**: expert-backend
- **Depends on**: T-ORC001-08
- **Action**: Same with `retired_param_hint: "artifact_type=plugin"`.
- **Output**: file size 149 → ~40 lines.
- **AC handle**: AC-ORC-001-05, AC-ORC-001-13.

### T-ORC001-10 — Retire expert-debug body to stub

- **Owner**: expert-backend
- **Depends on**: T-ORC001-09
- **Action**: Replace expert-debug.md body with retired-stub schema
  (`retired_replacement: manager-quality`, `retired_param_hint:
  "diagnostic-mode"`); reference manager-quality.md Diagnostic Sub-Mode in
  redirect text.
- **Output**: file size 213 → ~40 lines.
- **AC handle**: AC-ORC-001-05, AC-ORC-001-13.

### T-ORC001-11 — Retire expert-testing body to stub (dual-target)

- **Owner**: expert-backend
- **Depends on**: T-ORC001-10
- **Action**: Replace expert-testing.md body with dual-target stub
  (`retired_replacement: manager-cycle | expert-performance`); body
  describes split (strategy → manager-cycle cycle_type=tdd; load →
  expert-performance --deepthink load-test).
- **Output**: file size 116 → ~50 lines.
- **AC handle**: AC-ORC-001-05, AC-ORC-001-13.

### T-ORC001-12 — Run make build (M2 build)

- **Owner**: expert-backend
- **Depends on**: T-ORC001-11
- **Action**: `cd /Users/goos/.moai/worktrees/MoAI-ADK/orc-001-plan && make build`.
- **Output**: regenerated `internal/template/embedded.go`; local
  `.claude/agents/moai/` mirrors template (verified by `diff -r`).
- **AC handle**: AC-ORC-001-06 (M2 partial check).

### T-ORC001-13 — Verify M2 trigger-union test

- **Owner**: manager-quality
- **Depends on**: T-ORC001-12
- **Action**: Run AC-12 verification script comparing builder-platform
  trigger row to snapshot; expect 0 dropped keywords.
- **Output**: PASS/FAIL marker for AC-12 partial (builder-platform half).
- **AC handle**: AC-ORC-001-12 (builder portion).

---

## M3 — Refactor Passes on Existing Agents (P1)

### T-ORC001-14 — Add Diagnostic Sub-Mode + drop Context7 from manager-quality

- **Owner**: expert-backend
- **Depends on**: T-ORC001-12
- **Action**:
  1. Insert "## Diagnostic Sub-Mode" section after Primary Mission.
  2. Inline expert-debug delegation table (semantic equivalent).
  3. Drop `mcp__context7__*` from tools list (L13 of current file).
- **Output**: manager-quality.md gains ~40-60 lines; tools list shrinks.
- **AC handle**: AC-ORC-001-04, AC-ORC-001-10.

### T-ORC001-15 — Scope-shrink manager-project

- **Owner**: expert-backend
- **Depends on**: T-ORC001-14
- **Action**:
  1. Remove L57-65 routing block (3 over-scoped modes).
  2. Insert "Scope Boundary" section with `.moai/project/{product,structure,tech}.md` allowlist + blocker-report template.
- **Output**: manager-project.md changes by ~30 lines.
- **AC handle**: AC-ORC-001-08, AC-ORC-001-14.

### T-ORC001-16 — Trigger dedup for expert-backend

- **Owner**: expert-backend
- **Depends on**: T-ORC001-15
- **Action**:
  1. Reduce L6 EN row to 12-15 high-precision tokens.
  2. Drop standalone `Oracle` from KO/JA/ZH where localized form exists.
- **Output**: expert-backend.md trigger rows shrink; `grep -c "Oracle"` ≤ 4.
- **AC handle**: AC-ORC-001-09.

### T-ORC001-17 — Trigger trim for manager-git

- **Owner**: expert-backend
- **Depends on**: T-ORC001-16
- **Action**: Reduce trigger rows L6-9 to ~8 high-precision tokens per
  P-A16.
- **Output**: manager-git.md trigger rows shrink.
- **AC handle**: (informational — REQ-010 covers this implicitly).

### T-ORC001-18 — Add Write to expert-performance

- **Owner**: expert-backend
- **Depends on**: T-ORC001-17
- **Action**:
  1. Insert `Write` between `Read` and `Edit` in tools list.
  2. Add "Write Scope" body section limiting writes to `.moai/docs/performance-analysis-*.md`.
- **Output**: expert-performance.md tools list grows by 1 entry.
- **AC handle**: AC-ORC-001-15.

### T-ORC001-19 — Add memory: project to plan-auditor

- **Owner**: expert-backend
- **Depends on**: T-ORC001-18
- **Action**: Insert `memory: project` line after `permissionMode: default`
  in plan-auditor frontmatter.
- **Output**: plan-auditor.md frontmatter gains 1 line.
- **AC handle**: AC-ORC-001-11.

### T-ORC001-20 — Run make build (M3 build)

- **Owner**: expert-backend
- **Depends on**: T-ORC001-19
- **Action**: `make build`.
- **Output**: regenerated embedded.go; verified `diff -r` empty.
- **AC handle**: AC-ORC-001-06 (M3 partial check).

---

## M4 — Downstream Reference Sync (P1)

### T-ORC001-21 — Sweep skill files for retired references

- **Owner**: manager-docs
- **Depends on**: T-ORC001-20
- **Action**: Run sweep `grep -rln "manager-ddd|manager-tdd|builder-agent|builder-skill|builder-plugin|expert-debug|expert-testing"
  internal/template/templates/.claude/skills/ .claude/skills/ | grep -v examples.md`. Replace each match per REQ-007 migration table.
  Excludes `examples.md` per OQ-3.
- **Output**: ~57 skill file edits (template + local mirror after build).
- **AC handle**: M4 verification gate (no AC ID; checked at M5).

### T-ORC001-22 — Sweep rule files for retired references

- **Owner**: manager-docs
- **Depends on**: T-ORC001-21
- **Action**: Same sweep over
  `internal/template/templates/.claude/rules/` and `.claude/rules/`. ~28
  hits.
- **Output**: rule file edits.
- **AC handle**: M4 verification gate.

### T-ORC001-23 — Update CLAUDE.md Agent Catalog

- **Owner**: manager-docs
- **Depends on**: T-ORC001-22
- **Action**:
  1. Edit root `CLAUDE.md` L106-126: Manager 8→7, Expert 8→6, Builder 3→1.
  2. Update L62, L147, L378 invocation strings.
  3. Edit template `internal/template/templates/CLAUDE.md` L136 and L378.
- **Output**: CLAUDE.md (root + template) reflects new roster counts.
- **AC handle**: AC-ORC-001-01 (informational; documentation parallel).

---

## M5 — REFACTOR + MX Tags + Completion Gate (P0)

### T-ORC001-24 — Apply MX tags + final verification + push

- **Owner**: manager-quality (verification), manager-spec (sign-off)
- **Depends on**: T-ORC001-23
- **Action**:
  1. Apply `@MX:NOTE`, `@MX:WARN`, `@MX:ANCHOR` tags planned in M2-M4 to
     actual code lines.
  2. Resolve all `@MX:TODO` from M4.
  3. Update spec.md §10.1 destiny table for post-R5 additions (per OQ-1,
     OQ-5).
  4. Final `make build && make test`; verify exit 0.
  5. Final `golangci-lint run`; verify clean.
  6. Final `diff -r` template↔local; verify empty.
  7. Run all 17 AC verifications from acceptance.md.
  8. Stage commit `feat(agents): SPEC-V3R2-ORC-001 — roster consolidation
     (22 → 17)`.
  9. Push branch.
  10. Open PR with plan-auditor request for review.
- **Output**: PR open; all ACs PASS.
- **AC handle**: ALL (final integration gate).

---

## Dependency Graph (textual)

```
T-01 → T-02 → T-03 → T-04 → T-05    [M1 chain]
                              ↓
T-06 → T-07 → T-08 → T-09 → T-10 → T-11 → T-12 → T-13    [M2 chain]
                                                    ↓
T-14 → T-15 → T-16 → T-17 → T-18 → T-19 → T-20    [M3 chain]
                                              ↓
T-21 → T-22 → T-23    [M4 chain]
                  ↓
T-24    [M5 final gate]
```

All milestone chains are sequential (no intra-milestone parallelism in the
plan, since file edits are coupled). Inter-milestone dependencies are
explicit (T-12 must complete before T-14; T-20 before T-21; T-23 before
T-24).

---

## Owner Role Summary

| Owner | Tasks | Count |
|-------|-------|------:|
| manager-spec | T-01, T-02, T-05, T-24 (review) | 4 |
| expert-backend | T-03, T-04, T-06, T-07, T-08, T-09, T-10, T-11, T-12, T-14, T-15, T-16, T-17, T-18, T-19, T-20 | 16 |
| manager-quality | T-13, T-24 (verification) | 2 |
| manager-docs | T-21, T-22, T-23 | 3 |

**Total: 24 tasks, 4 owner roles**.

---

## Estimated Diff Size

| Milestone | Files touched (template tree) | Approx lines added | Approx lines removed |
|-----------|-------------------------------:|-------------------:|---------------------:|
| M1 | 0 | 0 (snapshot files only) | 0 |
| M2 | 6 | ~180 (builder-platform new) | ~600 (5 stubs replace bodies) |
| M3 | 6 | ~80 | ~30 |
| M4 | ~85 (skills + rules + CLAUDE.md) | ~120 | ~120 |
| M5 | spec.md + progress.md only | ~20 | 0 |

Total: ~6 + 6 + 85 + 2 = ~99 files changed; net ~−400 lines (mostly stub
shrinkage); ~+250 lines added in builder-platform body and M3 sections.

---

End of tasks.md.
