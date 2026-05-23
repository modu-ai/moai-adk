---
id: SPEC-V3R6-LEGACY-CLEANUP-001
title: "Implementation Plan — v2.x agency keyword residual cleanup"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: Medium
tags: "cleanup, legacy, v3-roadmap, sprint-2, docs, brand-design"
phase: "v3.0.0"
module: "docs"
lifecycle: spec-anchored
tier: L
related_specs: [SPEC-AGENCY-ABSORB-001, SPEC-V3R6-CHANGELOG-CLEANUP-001]
---

# Implementation Plan — SPEC-V3R6-LEGACY-CLEANUP-001

## Overview

This **Tier L** SPEC (reclassified iter-2 from Tier M per spec-workflow.md SSoT: 31 files > 15 = Tier L) performs a coordinated cleanup of residual `agency` keyword references across 31 user-facing files. The work is structured into 4 milestones (M1–M4) with files-touched-per-milestone capped at ≤10 to retain Tier M files-per-M voluntary discipline (Tier L does not mandate the ≤10 cap, but smaller per-commit blast radius improves CI gating).

**Execution mode**: Hybrid Trunk Tier L with 1-person OSS **explicit per-SPEC override** of CLAUDE.local.md §23 [HARD] feat-branch+PR rule (per user decision 2026-05-23 AskUserQuestion Q2). Direct main push (no PR required) for this SPEC only; default flow for future Tier L SPECs remains §23 [HARD] feat-branch + auto PR. pre-push hook warn-only, 4 CI status checks gating. See spec.md §A.5 for override rationale + lesson L32 candidate.

## Milestone Breakdown

### M1 — Backup + Skills + Rule (Files touched: 5 + backup manifest)

**Scope**: Create backup artifact and edit the 4 skill files + 1 rule file. This is the lowest-risk milestone (no docs-site Hugo build dependency, no 4-locale parity concern).

**Files touched**:
1. `.moai/backups/legacy-cleanup-{ISO-DATE}/` (new directory + manifest.json)
2. `.claude/skills/moai-domain-brand-design/SKILL.md`
3. `.claude/skills/moai-domain-copywriting/SKILL.md`
4. `.claude/skills/moai-workflow-gan-loop/SKILL.md`
5. `.claude/skills/moai/workflows/design.md`
6. `.claude/rules/moai/design/constitution.md`

**Tasks**:
- T1.1: Create backup directory `.moai/backups/legacy-cleanup-$(date -u +%Y-%m-%dT%H%M%SZ)/` and copy all 31 in-scope files (preserving paths) (REQ-LCL-001)
- T1.2: Generate `manifest.json` listing path + SHA256 + bytes for each backed-up file (REQ-LCL-002)
- T1.3: Per-file inspection of 5 skill/rule files — classify each `agency` occurrence into 4 categories (REQ-LCL-006)
- T1.4: Apply surgical edits per category strategy (REQ-LCL-005)
- T1.5: Spot-verify SHA256 of PRESERVE paths sample (5 files) remained unchanged
- T1.6: Commit `plan(SPEC-V3R6-LEGACY-CLEANUP-001): M1 — backup + skills/rule cleanup`

**Verification (binary)**:
- `ls .moai/backups/legacy-cleanup-*/manifest.json` returns 1 result
- `jq 'length' .moai/backups/legacy-cleanup-*/manifest.json` returns 31
- `grep -c 'agency' .claude/skills/moai-domain-brand-design/SKILL.md` returns ≤1 (after edit)
- `grep -c 'agency' .claude/rules/moai/design/constitution.md` returns ≤1 (after edit)
- `go test ./internal/template/...` PASS (template lint guards)

**Risk**: LOW. Skills/rules are self-contained markdown; no Hugo or test dependency.

---

### M2 — docs-site ko + en (Files touched: 10)

**Scope**: Edit ko (5) + en (5) docs-site files. Hugo build validated after edit. Parity tracker initialized.

**Files touched**:
1. `docs-site/content/ko/design/_index.md`
2. `docs-site/content/ko/design/gan-loop.md`
3. `docs-site/content/ko/design/getting-started.md`
4. `docs-site/content/ko/design/migration-guide.md`
5. `docs-site/content/ko/workflow-commands/moai-design.md`
6. `docs-site/content/en/design/_index.md`
7. `docs-site/content/en/design/gan-loop.md`
8. `docs-site/content/en/design/getting-started.md`
9. `docs-site/content/en/design/migration-guide.md`
10. `docs-site/content/en/workflow-commands/moai-design.md`

**Tasks**:
- T2.1: Per-file inspection of 10 docs-site files (5 ko + 5 en) — classify each `agency` occurrence
- T2.2: Apply surgical edits to ko files first (5 files)
- T2.3: Mirror the same semantic edit to en files (5 files), preserving the per-file 1:1 parity (REQ-LCL-010)
- T2.4: Create parity tracker note documenting "edit applied to <ko-file>, mirrored to <en-file>" for each of 5 pairs
- T2.5: Run `hugo --source docs-site --quiet` and confirm exit 0 (REQ-LCL-011)
- T2.6: Commit `plan(SPEC-V3R6-LEGACY-CLEANUP-001): M2 — docs-site ko + en cleanup`

**Verification (binary)**:
- `find docs-site/content/ko/design docs-site/content/en/design docs-site/content/ko/workflow-commands docs-site/content/en/workflow-commands -name "*.md" | xargs grep -l agency | wc -l` returns ≤4 (allowing legitimate retired-reference mentions, REQ-LCL-013)
- `hugo --source docs-site --quiet; echo $?` returns `0`
- Parity tracker note present in M2 commit message body

**Risk**: MEDIUM. Hugo build dependency; markdown front-matter parsing strict.

---

### M3 — docs-site ja + zh (Files touched: 10)

**Scope**: Mirror M2's edits to ja (5) + zh (5) docs-site files. Final Hugo build + 4-locale parity verification.

**Files touched**:
1. `docs-site/content/ja/design/_index.md`
2. `docs-site/content/ja/design/gan-loop.md`
3. `docs-site/content/ja/design/getting-started.md`
4. `docs-site/content/ja/design/migration-guide.md`
5. `docs-site/content/ja/workflow-commands/moai-design.md`
6. `docs-site/content/zh/design/_index.md`
7. `docs-site/content/zh/design/gan-loop.md`
8. `docs-site/content/zh/design/getting-started.md`
9. `docs-site/content/zh/design/migration-guide.md`
10. `docs-site/content/zh/workflow-commands/moai-design.md`

**Tasks**:
- T3.1: Apply M2-mirrored edits to ja files (5 files), using DeepL / professional translation reference for terminology (PRESERVE i18n quality)
- T3.2: Apply M2-mirrored edits to zh files (5 files), same approach
- T3.3: Cross-locale parity verification — for each of 5 documents, confirm ko/en/ja/zh have semantically equivalent edits (REQ-LCL-010)
- T3.4: Run `hugo --source docs-site --quiet` final build (REQ-LCL-011)
- T3.5: Run global grep `grep -rln -E '\bagency\b|\.agency/|/agency/' docs-site/content/` and verify count is symmetric across 4 locales
- T3.6: Commit `plan(SPEC-V3R6-LEGACY-CLEANUP-001): M3 — docs-site ja + zh cleanup + 4-locale parity verified`

**Verification (binary)**:
- `for loc in ko en ja zh; do echo "$loc: $(grep -rln agency docs-site/content/$loc/design docs-site/content/$loc/workflow-commands 2>/dev/null | wc -l)"; done` shows equal counts across all 4 locales
- `hugo --source docs-site --quiet; echo $?` returns `0`
- Lighthouse score for docs-site index ≥80 (per CLAUDE.local.md §17 docs-site policy — already enforced by CI)

**Risk**: MEDIUM. i18n translation quality; risk of asymmetric edits drifting locales.

---

### M4 — Root markdown + CHANGELOG + verification (Files touched: 6)

**Scope**: Edit the 6 root markdown files (CHANGELOG, CLAUDE.md, READMEs × 4 locales). Final regression verification.

**Files touched**:
1. `CHANGELOG.md` (per REQ-LCL-007: pre-v3.0 entries preserved; REQ-LCL-008: `[Unreleased]` and v3.0+ may be refactored)
2. `CLAUDE.md`
3. `README.md`
4. `README.ko.md`
5. `README.ja.md`
6. `README.zh.md`

**Tasks**:
- T4.1: Inspect `CHANGELOG.md` and identify which `agency` occurrences are pre-v3.0 (preserve) vs v3.0+/`[Unreleased]` (refactor per REQ-LCL-008)
- T4.2: Apply surgical edits to CHANGELOG.md per category (REQ-LCL-007 + REQ-LCL-008)
- T4.3: Edit `CLAUDE.md` — per-file inspection, classify each `agency` occurrence
- T4.4: Edit the 4 README locale variants in parity (REQ-LCL-010 semantic equivalence)
- T4.5: Final regression verification — run full verification batch (see Verification batch below)
- T4.6: Commit `plan(SPEC-V3R6-LEGACY-CLEANUP-001): M4 — root markdown + final verification`

**Verification (binary)** — parallel 5-cmd batch per `.claude/rules/moai/workflow/verification-batch-pattern.md`:
```bash
# 1. In-scope keyword count
grep -rln -E '\bagency\b|\.agency/|/agency/' CHANGELOG.md CLAUDE.md README.md README.ko.md README.ja.md README.zh.md .claude/skills/moai-domain-brand-design/ .claude/skills/moai-domain-copywriting/ .claude/skills/moai-workflow-gan-loop/ .claude/skills/moai/workflows/design.md .claude/rules/moai/design/constitution.md docs-site/content/ | wc -l
# Expected: ≤5 (REQ-LCL-013)

# 2. PRESERVE set unchanged — SHA256 sample
shasum -a 256 .moai/design/v3-legacy/*.md | head -3
# Expected: matches manifest.json pre-edit hashes

# 3. Hugo build green
hugo --source docs-site --quiet; echo $?
# Expected: 0

# 4. Go test regression (no .go files touched)
go test -count=1 ./... | tail -3
# Expected: PASS, no new failures

# 5. golangci-lint baseline
golangci-lint run --timeout=2m | tail -3
# Expected: 0 issues (or matching pre-edit baseline)
```

**Risk**: LOW-MEDIUM. CHANGELOG editing requires careful pre-v3.0 vs v3.0+ classification. READMEs are 4-locale parity-sensitive.

---

## Technical Approach

### Edit Categories Strategy (per REQ-LCL-006)

For each `agency` occurrence, implementer classifies into one of:

| Category | Strategy | Example |
|----------|----------|---------|
| 1. v2.x concept reference | Replace inline with `"v2.x agency (retired, see [SPEC-AGENCY-ABSORB-001](.moai/specs/SPEC-AGENCY-ABSORB-001/spec.md))"` OR remove if redundant | "The agency framework provides..." → "The v2.x agency (retired, see SPEC-AGENCY-ABSORB-001) provided..." |
| 2. `.agency/` directory path | Remove path reference or redirect to current location | "Stored in `.agency/skills/`" → (sentence removed) |
| 3. Skill frontmatter `description:` mention | Rewrite description to absorption-aware language | "description: ... agency-driven brand design ..." → "description: ... brand design (post-agency-absorption) ..." |
| 4. Historical CHANGELOG entry | Preserve untouched (append-only, REQ-LCL-007) | "v2.5.0 (2026-04): ... agency module ..." → unchanged |

### Backup Manifest Format

```json
{
  "spec_id": "SPEC-V3R6-LEGACY-CLEANUP-001",
  "created_at": "2026-05-23T180000Z",
  "files": [
    {
      "path": ".claude/skills/moai-domain-brand-design/SKILL.md",
      "sha256": "abc123...",
      "bytes": 12345
    }
  ]
}
```

### Tool Discipline

- `grep` for verification (NOT `find` + `xargs`, per agent-common-protocol)
- `Edit` tool for surgical edits (NOT `Write` or `sed`)
- Read-before-edit for every file (HARD rule)
- 4-locale edits applied as 4 separate `Edit` calls per logical change (avoid `Multi-Edit` cross-locale drift)

## Risks and Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Hugo build breaks on markdown edit | Medium | High | Run `hugo --source docs-site --quiet` after every locale group; revert immediately if exit ≠ 0 |
| 4-locale parity drift | Medium | Medium | Parity tracker note per logical edit (M2 task T2.4); cross-locale grep verification (M3 task T3.5) |
| Pre-v3.0 CHANGELOG entries accidentally edited | Low | Medium | M4 task T4.1 explicit classification before edit; verify CHANGELOG diff scope manually |
| Template-source desync after this SPEC merges | High | Low | Documented in §A.6 / §C exclusion #2; SPEC-V3R6-LEGACY-CLEANUP-002 must follow IMMEDIATELY |
| Parallel session race during long run-phase | Medium (L9 reinforced) | Low | `git fetch origin && git log --oneline -15` at M-boundary; rebase if needed |
| Spawn-prompt §C count mismatch propagating | Resolved (this plan) | N/A | §A.1.7 documents 31-file corrected inventory; downstream agents read this plan, not spawn prompt §C |

## Verification Strategy

- **Per-milestone**: Binary AC verification listed in each M section above
- **Final (M4)**: 5-cmd parallel verification batch (in-scope keyword count + PRESERVE SHA256 + Hugo + go test + lint)
- **Out-of-band**: Implementer SHOULD run `find . -name "*.md" -newer .moai/specs/SPEC-V3R6-LEGACY-CLEANUP-001/spec.md -not -path "./.git/*" | sort` to confirm only in-scope files were modified

## Commit Strategy

Each milestone produces one commit on `main` directly (1-person OSS Hybrid Trunk Tier L **explicit per-SPEC override** of CLAUDE.local.md §23 [HARD], per user decision 2026-05-23):

```
plan(SPEC-V3R6-LEGACY-CLEANUP-001): M1 — backup + skills/rule cleanup
plan(SPEC-V3R6-LEGACY-CLEANUP-001): M2 — docs-site ko + en cleanup
plan(SPEC-V3R6-LEGACY-CLEANUP-001): M3 — docs-site ja + zh cleanup + 4-locale parity verified
plan(SPEC-V3R6-LEGACY-CLEANUP-001): M4 — root markdown + final verification
```

The plan-phase artifact commit (this commit) is separate:
```
plan(SPEC-V3R6-LEGACY-CLEANUP-001): Tier M minimal — agency keyword residual cleanup (scope C, 31 files)
```

All commits signed with `🗿 MoAI <email@mo.ai.kr>` per CLAUDE.local.md §4 and observability requirements.
