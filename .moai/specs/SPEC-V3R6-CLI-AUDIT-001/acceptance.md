---
id: SPEC-V3R6-CLI-AUDIT-001
title: "Acceptance Criteria — moai CLI inventory + dead command + integration analysis (research baseline)"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
tier: M
tags: "cli, audit, research, sprint-2, sprint-7-baseline"
---

# Acceptance Criteria — SPEC-V3R6-CLI-AUDIT-001

All criteria below are **binary-verifiable** with concrete commands. Each AC links back to one or more REQ-CLA-XXX from spec.md §2.

## AC-CLA-001 — Subcommand inventory ≥40 entries with structured fields

**Linked REQ**: REQ-CLA-001
**Linked Milestone**: M1

**Given**: The audit research is invoked via `/moai run SPEC-V3R6-CLI-AUDIT-001`
**When**: Milestone M1 (Subcommand Inventory) completes
**Then**: The audit report at `.moai/reports/cli-audit/audit-{ISO-DATE}.md` contains a `## §1 Subcommand Inventory` section with ≥40 table rows, each row representing one subcommand with 5 columns: full command path, parent, file:line registration, flags, brief use description.

**Verification**:
```bash
# Audit report exists
ls .moai/reports/cli-audit/audit-*.md | wc -l
# Expected: ≥1

# Subcommand row count (table rows starting with `| `moai ` indicating inventory entries)
grep -cE '^\| `moai ' .moai/reports/cli-audit/audit-*.md
# Expected: ≥40

# Section header present
grep "^## §1 Subcommand Inventory" .moai/reports/cli-audit/audit-*.md | wc -l
# Expected: 1

# Total subcommand count footer
grep "Total subcommands:" .moai/reports/cli-audit/audit-*.md | wc -l
# Expected: 1

# Total flag count footer
grep "Total flag definitions:" .moai/reports/cli-audit/audit-*.md | wc -l
# Expected: 1
```

---

## AC-CLA-002 — Every inventoried subcommand classified with grep evidence

**Linked REQ**: REQ-CLA-002
**Linked Milestone**: M2

**Given**: Milestone M1 inventory is complete
**When**: Milestone M2 (Dead-Command Classification) completes
**Then**: Every subcommand from M1 appears in `## §2 Dead-Command Classification` table classified as active/internal-only/dead-suspect, with evidence count per source class (A/B/C/D/E/F → e.g., `2/1/0/0/0/3`). A `### §2.1 Dead-Suspect Candidates` subsection enumerates the dead-suspect list with reproducible grep commands. The `harness-observe*` family preliminary classification is confirmed or refuted with evidence.

**Verification**:
```bash
# §2 section header
grep "^## §2 Dead-Command Classification" .moai/reports/cli-audit/audit-*.md | wc -l
# Expected: 1

# §2.1 subsection
grep "^### §2.1 Dead-Suspect Candidates" .moai/reports/cli-audit/audit-*.md | wc -l
# Expected: 1

# Classification table contains 3 classification labels
grep -cE '\| (active|internal-only|dead-suspect) \|' .moai/reports/cli-audit/audit-*.md
# Expected: ≥40 (one classification per M1 subcommand, parity)

# At least one dead-suspect classified
grep -c '| dead-suspect |' .moai/reports/cli-audit/audit-*.md
# Expected: ≥1

# Evidence count format present (NxNxNxNxNxN pattern)
grep -cE '\| [0-9]+/[0-9]+/[0-9]+/[0-9]+/[0-9]+/[0-9]+ \|' .moai/reports/cli-audit/audit-*.md
# Expected: ≥40 (parity with subcommand count)

# harness-observe family explicit mention
grep -c 'harness-observe' .moai/reports/cli-audit/audit-*.md
# Expected: ≥4 (4 observe-family commands referenced)
```

---

## AC-CLA-003 — Integration map covers init/update/profile triad with 4 sub-sections + mermaid

**Linked REQ**: REQ-CLA-003
**Linked Milestone**: M3

**Given**: Milestones M1 and M2 complete
**When**: Milestone M3 (Integration Map) completes
**Then**: The audit report `## §3 Integration Map` contains 4 sub-sections (§3.1 moai init flow, §3.2 moai update 10-flag matrix, §3.3 moai cc -p profile system, §3.4 Cross-cutting concerns) and at least one mermaid TD diagram showing the 3-way relationship between init/update/profile commands. The §3.2 flag matrix is a 10×10 table.

**Verification**:
```bash
# 4 sub-section headers
grep -c "^### §3\.[1-4]" .moai/reports/cli-audit/audit-*.md
# Expected: 4 (one per sub-section)

# §3.1 moai init flow header
grep "^### §3.1 moai init flow" .moai/reports/cli-audit/audit-*.md | wc -l
# Expected: 1

# §3.2 moai update 10-flag matrix header
grep "^### §3.2 moai update 10-flag matrix" .moai/reports/cli-audit/audit-*.md | wc -l
# Expected: 1

# §3.3 moai cc -p profile system header
grep "^### §3.3 moai cc -p profile system" .moai/reports/cli-audit/audit-*.md | wc -l
# Expected: 1

# §3.4 Cross-cutting concerns header
grep "^### §3.4 Cross-cutting" .moai/reports/cli-audit/audit-*.md | wc -l
# Expected: 1

# 10×10 flag matrix presence (10 flag names from spec.md §2 REQ-CLA-003 listed)
for flag in check shell-env config force yes templates-only binary dry-run no-hooks verbose; do
  grep -c "$flag" .moai/reports/cli-audit/audit-*.md
done
# Expected: each grep returns ≥1 (all 10 flags documented in §3.2 matrix)

# Mermaid TD diagram present
grep -c '```mermaid' .moai/reports/cli-audit/audit-*.md
# Expected: ≥1

# Mermaid uses TD/TB direction (per CLAUDE.local.md §17 Mermaid TD-only rule for consistency)
grep -cE 'graph (TD|TB)' .moai/reports/cli-audit/audit-*.md
# Expected: ≥1 (at least one TD/TB diagram in §3)
```

---

## AC-CLA-004 — Sprint 7 baseline scope directly consumable

**Linked REQ**: REQ-CLA-004
**Linked Milestone**: M4

**Given**: Milestones M1, M2, M3 complete
**When**: Milestone M4 (Sprint 7 Baseline + Methodology) completes
**Then**: The audit report contains `## §4 Sprint 7 Baseline Scope` with 4 sub-sections (§4.1 Recommended unifications, §4.2 Recommended retirements, §4.3 Integration bridging gaps, §4.4 Sprint 7 SPEC scope draft). The §4.4 sub-section is formatted as a 5-section outline that a future manager-spec invocation for SPEC-V3R6-CLI-INTEGRATION-001 can directly consume as input scope.

**Verification**:
```bash
# §4 main section header
grep "^## §4 Sprint 7 Baseline Scope" .moai/reports/cli-audit/audit-*.md | wc -l
# Expected: 1

# 4 sub-section headers
grep -c "^### §4\.[1-4]" .moai/reports/cli-audit/audit-*.md
# Expected: 4

# §4.1 Recommended unifications
grep "^### §4.1 Recommended unifications" .moai/reports/cli-audit/audit-*.md | wc -l
# Expected: 1

# §4.2 Recommended retirements
grep "^### §4.2 Recommended retirements" .moai/reports/cli-audit/audit-*.md | wc -l
# Expected: 1

# §4.3 Integration bridging gaps
grep "^### §4.3 Integration bridging gaps" .moai/reports/cli-audit/audit-*.md | wc -l
# Expected: 1

# §4.4 Sprint 7 SPEC scope draft (must mention 5 sections)
grep "^### §4.4 Sprint 7 SPEC scope draft" .moai/reports/cli-audit/audit-*.md | wc -l
# Expected: 1

# §4.4 contains 5-section outline (Section 1-5 markers)
grep -cE '^- Section [1-5]:|^#### Section [1-5]:' .moai/reports/cli-audit/audit-*.md
# Expected: ≥5

# Explicit Sprint 7 future-SPEC reference
grep -c 'SPEC-V3R6-CLI-INTEGRATION-001' .moai/reports/cli-audit/audit-*.md
# Expected: ≥1 (future SPEC named explicitly for consumption traceability)
```

---

## AC-CLA-005 — Research-only constraint: no code/template/docs-site modification

**Linked REQ**: REQ-CLA-005
**Linked Milestone**: All M1-M4

**Given**: Research execution (M1-M4) completes
**When**: Running `git diff --name-only main -- '*.go' '*.sh' '*.yaml' 'docs-site/' 'internal/template/templates/' '.claude/hooks/' '.claude/skills/' '.claude/rules/' '.claude/agents/'`
**Then**: The diff output is empty. The only changed files are: (a) audit report at `.moai/reports/cli-audit/audit-*.md` (new), (b) 4 SPEC artifacts at `.moai/specs/SPEC-V3R6-CLI-AUDIT-001/{spec,plan,acceptance,progress}.md` (modified or new for progress.md), and optionally (c) CHANGELOG.md if `/moai sync` is invoked post-run.

**Verification**:
```bash
# REQ-CLA-005 [Unwanted] constraint: zero diff in protected paths
git diff --name-only main -- \
  '*.go' \
  '*.sh' \
  '*.yaml' \
  'docs-site/' \
  'internal/template/templates/' \
  '.claude/hooks/' \
  '.claude/skills/' \
  '.claude/rules/' \
  '.claude/agents/' \
  | wc -l
# Expected: 0 (no changes in any protected path class)

# Allowed changes only: .moai/reports/cli-audit/ + .moai/specs/SPEC-V3R6-CLI-AUDIT-001/
git diff --name-only main \
  | grep -vE '^\.moai/reports/cli-audit/|^\.moai/specs/SPEC-V3R6-CLI-AUDIT-001/|^CHANGELOG\.md$' \
  | wc -l
# Expected: 0 (no out-of-scope file changes)

# Audit report file exists
ls .moai/reports/cli-audit/audit-*.md | wc -l
# Expected: ≥1

# 4 SPEC artifacts exist
ls .moai/specs/SPEC-V3R6-CLI-AUDIT-001/{spec,plan,acceptance,progress}.md | wc -l
# Expected: 4

# SPEC frontmatter status synced to implemented at M4
grep "^status: implemented" .moai/specs/SPEC-V3R6-CLI-AUDIT-001/spec.md | wc -l
# Expected: 1 (after M4 commit)

# SPEC version bumped at M4
grep '^version: "0.2.0"' .moai/specs/SPEC-V3R6-CLI-AUDIT-001/spec.md | wc -l
# Expected: 1 (after M4 commit)
```

---

## AC-CLA-006 — Methodology appendix documents reproducibility

**Linked REQ**: REQ-CLA-001 + REQ-CLA-002 + REQ-CLA-003 (cross-cutting reproducibility)
**Linked Milestone**: M4

**Given**: Milestones M1-M4 complete
**When**: Inspecting the audit report `## §5 Methodology Appendix`
**Then**: The appendix lists all grep commands used in M1+M2+M3, files scanned per class, exclusions applied, audit generation date + git SHA + moai version.

**Verification**:
```bash
# §5 Methodology header
grep "^## §5 Methodology Appendix" .moai/reports/cli-audit/audit-*.md | wc -l
# Expected: 1

# Appendix contains grep command examples (code block)
awk '/^## §5 Methodology Appendix/,/^## /' .moai/reports/cli-audit/audit-*.md \
  | grep -c '```bash'
# Expected: ≥1 (at least one bash code block documenting grep commands)

# Audit generation metadata present
grep -cE 'Generated:|Git SHA:|moai version:' .moai/reports/cli-audit/audit-*.md
# Expected: ≥3 (date + git SHA + moai version)
```

---

## Cross-AC Summary Table

| AC | Milestone | Key verification | Severity if fail |
|----|-----------|------------------|------------------|
| AC-CLA-001 | M1 | ≥40 subcommand rows | HIGH (incomplete inventory blocks Sprint 7) |
| AC-CLA-002 | M2 | Every subcommand classified + harness-observe family | HIGH (dead-suspect false positive risk) |
| AC-CLA-003 | M3 | 4 sub-sections + 10×10 flag matrix + mermaid | MEDIUM (integration map incompleteness slows Sprint 7) |
| AC-CLA-004 | M4 | §4 directly consumable by Sprint 7 manager-spec | HIGH (this IS the deliverable purpose) |
| AC-CLA-005 | All M1-M4 | Zero protected path diff | CRITICAL (REQ-CLA-005 [Unwanted] violation) |
| AC-CLA-006 | M4 | Methodology reproducibility | MEDIUM (audit refresh depends on this) |

---

## Definition of Done

[HARD] All 6 ACs MUST PASS for SPEC completion. Any AC failure transitions status from `in-progress` → blocked; resolution requires either (a) deliverable correction + re-verification, or (b) plan-phase revision (rare for research-only SPECs).

Run-phase completion criteria:
- 6/6 ACs PASS (verified via parallel Bash batch)
- 4 commits on feat-branch (M1, M2, M3, M4) OR main-direct (if per-SPEC override granted at run-phase)
- SPEC frontmatter `status: implemented`, `version: 0.2.0`, `updated: <ISO-DATE>`
- Audit report at `.moai/reports/cli-audit/audit-*.md` is complete with §1-§5
- Sprint 7 SPEC-V3R6-CLI-INTEGRATION-001 explicitly named in §4.4 as consumer

---

Version: 0.1.0 (plan-phase draft)
