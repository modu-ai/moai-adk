---
id: SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001
title: "Template mirror cascade: implementation plan"
version: "0.1.0"
status: implemented
created: 2026-05-24
updated: 2026-05-24
author: GOOS행님
priority: P3
phase: "v3.0.0"
module: "internal/template/templates/.claude/skills/moai/workflows/plan"
lifecycle: spec-anchored
tags: "template-mirror, cascade, drift-fix, tier-s, sprint-2-p4-3"
---

# Implementation Plan — SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001

## 1. Approach Summary

Single-file mechanical content overwrite. The mirror at `internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md` is brought back to byte-for-byte parity with the operational source at `.claude/skills/moai/workflows/plan/spec-assembly.md`. The source predates the WORKFLOW-LEAN-001 Phase 1.6 addition; the mirror lags by 32 lines (Phase 1.6 Tier Judgment Socratic Question block + tier-conditional artifact set rules). Zero production code change. Zero source change. Zero behavior change in any runtime path. The test `TestLateBranchTemplateMirror/spec-assembly.md` flips from FAIL to PASS.

## 2. Milestones

Tier S minimal — **single milestone** sufficient.

### M1 — Mirror file overwrite (mechanical)

**Owner**: manager-develop (Tier S minimal cycle per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability).

**Edit map**:

| Location | Operation | Source | Method | Expected delta |
|----------|-----------|--------|--------|----------------|
| `internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md` | complete content overwrite | `.claude/skills/moai/workflows/plan/spec-assembly.md` (548 lines, 28,423 bytes) | `cp <source> <mirror>` (preserves byte-for-byte fidelity) OR `Read` source full + `Write` mirror with identical content | +32 lines, +2,484 bytes in mirror; 0 changes in source |

**Estimated LOC**: 32 (net additions to mirror only — though they appear as 32 new lines added to the mirror, the source is the canonical reference; no source LOC changed).
**Estimated files**: 1 modified (mirror), 0 created, 0 deleted, 0 sourced edited.
**Tier confirmation**: S (well within ≤300 LOC / ≤5 files / ≤1 milestone envelope; no per-SPEC override needed; L40 discipline preserved).

**Steps** (manager-develop performs in sequence):

1. Pre-flight read: `Read .claude/skills/moai/workflows/plan/spec-assembly.md` to confirm source is at 548 lines / 28,423 bytes (snapshot expected content).
2. Pre-flight verify: `wc -c -l .claude/skills/moai/workflows/plan/spec-assembly.md internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md` → confirm 548/28,423 and 516/25,939 respectively (matches §A.2 spec.md drift evidence).
3. Pre-flight diff: `diff .claude/skills/moai/workflows/plan/spec-assembly.md internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md | wc -l` → confirm 34 lines (32 missing + 2 diff header markers).
4. Apply edit: `cp .claude/skills/moai/workflows/plan/spec-assembly.md internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md` (or equivalent `Read` source + `Write` mirror with identical content).
5. Post-edit verify byte parity: `wc -c internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md | awk '{print $1}'` → exactly `28423`.
6. Post-edit verify diff cleared: `diff .claude/skills/moai/workflows/plan/spec-assembly.md internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md | wc -l` → exactly `0`.
7. Post-edit verify source untouched: `git diff -- .claude/skills/moai/workflows/plan/spec-assembly.md | wc -l` → exactly `0`.
8. Targeted test PASS: `go test ./internal/template/ -run 'TestLateBranchTemplateMirror/spec-assembly' -v` → `--- PASS: TestLateBranchTemplateMirror/spec-assembly.md`.
9. Quality baseline preserved: `go vet ./... 2>&1 | wc -l` → `0`; `golangci-lint run --timeout=2m 2>&1 | tail -1` → `0 issues.`.
10. Stage mirror only: `git add internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md` (REQ-TMC-007 path-specific staging).
11. Commit on main per Hybrid Trunk Tier S/M discipline (CLAUDE.local.md §23.7): Conventional Commits `fix(SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001): M1 — mirror parity for spec-assembly.md (+32 lines)` + body referencing source SPEC-V3R5-WORKFLOW-LEAN-001 + `🗿 MoAI <email>` trailer.
12. Push to `origin/main`.

**Done criteria**: 5 ACs PASS in progress.md §Run-phase Evidence + commit + push completed.

**No further milestones**: Tier S minimal envelope (1 milestone sufficient for 1-file mechanical content overwrite).

## 3. Technical Approach

### 3.1 Why complete overwrite vs surgical line-by-line patch

The drift is 32 lines covering a coherent documentation block (Phase 1.6 Tier Judgment Socratic Question + tier-conditional artifact set rules). Surgical line-by-line patching would require precise line-range knowledge plus careful adjacency handling. `cp` (or `Read` + `Write`) is mechanical, deterministic, and produces guaranteed byte-for-byte parity. The template-mirror invariant test `TestLateBranchTemplateMirror/spec-assembly.md` asserts byte parity (size + diff), so the strictest verification matches the strictest implementation.

### 3.2 Why this fix is safe for `moai init` / `moai update` consumers

The mirror file is consumed only by:

- `moai init` (initial template deployment to a new project) — copies the mirror verbatim into the user's `.claude/skills/moai/workflows/plan/spec-assembly.md`. Bringing the mirror to parity with the operational source means new projects ship with the same Phase 1.6 documentation that the operational `/moai plan` workflow already follows.
- `moai update` (template sync to existing project) — same effect: the user's existing skill body is updated to include the Phase 1.6 block. Per CLAUDE.local.md §24.4, `moai-*` namespace skills (which `moai/workflows/plan/` is) are template-managed and overwritten on update; user-owned `my-harness-*` skills are not touched. This fix is contract-compliant.

### 3.3 Why source is forbidden from modification

The source `.claude/skills/moai/workflows/plan/spec-assembly.md` is the canonical operational skill currently consumed by `/moai plan` workflow at every invocation. Modifying it as a side effect of this SPEC would:

- Violate REQ-TMC-002 (no operational regression invariant).
- Risk inadvertent behavior change in `/moai plan` workflow.
- Inflate scope beyond Tier S minimal envelope.

The minimal-change discipline is: mirror lags source → bring mirror to source, not the other way around.

## 4. Risk Mitigation

| Risk | Severity | Mitigation |
|------|---------:|------------|
| Source content has uncommitted edits at run-time that would propagate into the mirror inadvertently | Low | manager-develop pre-flight step 1 `Read` source + step 7 `git diff -- .claude/skills/moai/workflows/plan/spec-assembly.md` = 0 verifies clean source state pre-copy AND clean source state post-copy (no source modification side effect). |
| Other dirty/untracked files in working tree get accidentally staged | Low | REQ-TMC-007 path-specific `git add internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md` (step 10) — only the mirror is staged. PRESERVE list per REQ-TMC-004 covers the 7+2 known dirty paths. |
| Parallel session race condition during commit (L9 pattern) | Low | Pre-spawn fetch + `git rev-list --count --left-right origin/main...HEAD` = `0 0` check before commit per `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check L1. |
| Sibling pre-existing baseline failures persist (10+ failures from TEMPLATE-MIRROR-DRIFT-001 family + catalog/agent-folder drift) | Expected | REQ-TMC-006 explicitly accepts this state — net delta after this SPEC's fix = -1 baseline failure (the `spec-assembly.md` row only). Other failures attributable to sibling SPECs per L46 attribution discipline. |
| Future LEAN workflow updates to source would re-trigger drift | Medium | Out of scope for this SPEC. Future master `SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001` should consider a CI guard or pre-commit hook that automatically syncs template mirrors on source edits (deferred to Sprint 7). |

## 5. Dependencies

- **Predecessors**: SPEC-V3R6-I18N-VALIDATOR-BUDGET-001 (Sprint 2 P4.1, merged `d3ed4727d`) and SPEC-V3R6-SKILLS-AUDIT-RUN-MD-001 (Sprint 2 P4.2, merged `5e0dc6a9b`) — both clean baselines establishing Tier S minimal Section A-E precedent.
- **No blocking dependencies**: This SPEC is independent of any other in-flight or pending SPEC.
- **No code dependencies**: Edit isolated to single template mirror file.

## 6. Definition of Done (Plan-level)

- M1 complete with all 12 steps performed and verification commands PASS
- 5/5 ACs verified PASS (see acceptance.md)
- spec.md `spec-lint` clean (✓ No findings)
- Frontmatter canonical 12 fields valid across all 4 artifacts
- progress.md updated with `audit-ready` signal (this plan-phase) + run/sync/mx phase rows ready for downstream agents
- Commit on main + push to `origin/main` per Hybrid Trunk 1-person OSS discipline (CLAUDE.local.md §23.7)

## 7. Validation Strategy

### Phase 0.5 Plan Audit Gate

- plan-auditor invocation: iter-1 threshold **0.75** (Tier S) with **MP-2 EARS format** obligation.
- Expected outcome: PASS on iter-1 (all REQ-TMC-001..007 in EARS form; §B.2 explicit Out-of-Scope present; L46 attribution in §A.4 explicit; cascade scope in §B.1 explicit).
- If REVISE: orchestrator-direct fix-forward per L32 precedent (Tier S small scope, ≤ 5 textual edits).

### Local verification command sequence (verbatim, ordered)

```bash
# 1. Pre-flight: confirm drift baseline
wc -c -l .claude/skills/moai/workflows/plan/spec-assembly.md \
         internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md
# Expected: 548/28423 and 516/25939

diff .claude/skills/moai/workflows/plan/spec-assembly.md \
     internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md | wc -l
# Expected: 34 (32 missing lines + 2 diff header markers)

# 2. Apply fix (M1 step 4)
cp .claude/skills/moai/workflows/plan/spec-assembly.md \
   internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md

# 3. Post-edit byte parity
wc -c internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md | awk '{print $1}'
# Expected: 28423

# 4. Post-edit diff cleared
diff .claude/skills/moai/workflows/plan/spec-assembly.md \
     internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md | wc -l
# Expected: 0

# 5. Source untouched
git diff -- .claude/skills/moai/workflows/plan/spec-assembly.md | wc -l
# Expected: 0

# 6. Test PASS
go test ./internal/template/ -run 'TestLateBranchTemplateMirror/spec-assembly' -v
# Expected: --- PASS: TestLateBranchTemplateMirror/spec-assembly.md

# 7. Quality baseline
go vet ./... 2>&1 | wc -l
# Expected: 0
golangci-lint run --timeout=2m 2>&1 | tail -1
# Expected: 0 issues.

# 8. Stage + commit + push (Hybrid Trunk Tier S/M, manager-develop)
git add internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md
git commit -m "$(cat <<'EOF'
fix(SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001): M1 — mirror parity for spec-assembly.md (+32 lines)

TEMPLATE-MIRROR-DRIFT-001 family fix. Mirror at
internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md
brought to byte-for-byte parity with operational source at
.claude/skills/moai/workflows/plan/spec-assembly.md (548 lines / 28,423 bytes).

Drift originated from SPEC-V3R5-WORKFLOW-LEAN-001 which added Phase 1.6
(Tier Judgment Socratic Question) to source without propagating to mirror.

Test signal: TestLateBranchTemplateMirror/spec-assembly.md flipped from FAIL
(RULE_TEMPLATE_MIRROR_DRIFT at rule_template_mirror_test.go:182) to PASS.

Sprint 2 P4.3 (tail of P4 trio after IVB-001 d3ed4727d + SARM-001 5e0dc6a9b).

🗿 MoAI <namgoos@gmail.com>
EOF
)"
git push origin main
```

### Definition of Done

- 5 ACs (AC-TMC-001 through AC-TMC-005) all PASS in progress.md §Run-phase Evidence
- CHANGELOG `[Unreleased]` `### Fixed` entry added by sync-phase manager-docs (Tier S minimal: single line referencing SPEC ID + brief)
- B12 standing-rule guard sync-phase self-test PASS (9th consecutive)
- Mx Step C SKIP justified (0 Go production .go files modified; only template mirror .md file; no @MX:ANCHOR/WARN/NOTE/TODO trigger per `.claude/rules/moai/workflow/mx-tag-protocol.md` §a)
