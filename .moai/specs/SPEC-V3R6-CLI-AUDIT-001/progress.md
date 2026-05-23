---
id: SPEC-V3R6-CLI-AUDIT-001
title: "Progress Tracker — moai CLI inventory + dead command + integration analysis"
version: "0.2.0"
status: implemented
created: 2026-05-23
updated: 2026-05-23
author: manager-develop
priority: P1
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
tier: M
tags: "cli, audit, research, sprint-2, sprint-7-baseline, progress"
---

# Progress Tracker — SPEC-V3R6-CLI-AUDIT-001

## Milestone Tracker

| Milestone | Status | Commit SHA | Files touched | AC link | Notes |
|-----------|--------|-----------|---------------|---------|-------|
| M1 — Subcommand Inventory | PASS | b5ea8d936 | audit-2026-05-23.md (NEW §1) + progress.md (NEW) | AC-CLA-001 | Cobra command tree walked from cmd/moai/main.go root → 113 total subcommands |
| M2 — Dead-Command Classification | PASS | 2464ece58 | audit-2026-05-23.md (APPEND §2) + progress.md (UPDATE) | AC-CLA-002 | 116 classification rows (active/internal-only/dead-suspect). 10 dead-suspect candidates (Claude Code forward-compat hook events). harness-observe* preliminary suspect REFUTED (4 hooks reference). db-schema-sync confirmed internal-only per SPEC-DB-SYNC-001. |
| M3 — Integration Map | PASS | 123ddcb1a | audit-2026-05-23.md (APPEND §3) + progress.md (UPDATE) | AC-CLA-003 | §3.1 moai init flow + §3.2 moai update 10×10 flag matrix (11 awk-bounded rows) + §3.3 moai cc -p profile system + §3.4 cross-cutting concerns + §3.5 mermaid diagram (TD direction). Init/update/profile triad mapped, 4 integration gaps surfaced for Sprint 7. |
| M4 — Sprint 7 Baseline + Methodology + Frontmatter Sync | PASS | TBD | audit-2026-05-23.md (APPEND §4 + §5 + §6) + 4 SPEC artifacts (frontmatter sync) + progress.md (UPDATE) | AC-CLA-004 + AC-CLA-005 + AC-CLA-006 | §4.1-§4.4 Sprint 7 baseline scope + §5 methodology appendix (reproducibility) + §6 audit closure. 4 SPEC artifacts frontmatter status `draft → implemented`, version `0.1.0 → 0.2.0`. Sprint 7 5-section outline (§4.4) directly consumable by future manager-spec invocation. |

## AC Verification Matrix (M4 final)

| AC | Status | Command | Expected | Actual |
|----|--------|---------|----------|--------|
| AC-CLA-001 | PASS | grep -cE '^\| `moai ' audit-*.md | ≥40 | 253 |
| AC-CLA-002 | PASS | grep -cE '\| (active\|internal-only\|dead-suspect) \|' audit-*.md | ≥40 | 116 |
| AC-CLA-002 | PASS | grep -c '\| dead-suspect \|' audit-*.md | ≥1 | 10 |
| AC-CLA-002 | PASS | grep -c '^### §2.1 Dead-Suspect Candidates' audit-*.md | =1 | 1 |
| AC-CLA-002 | PASS | grep -c 'harness-observe' audit-*.md | ≥4 | 16 |
| AC-CLA-003 | PASS | grep -c '^### §3\.[1-4]' audit-*.md | =4 | 4 |
| AC-CLA-003 | PASS | awk-bounded §3.2 matrix rows | ≥11 | 11 |
| AC-CLA-003 | PASS | grep -c '```mermaid' audit-*.md | ≥1 | 1 |
| AC-CLA-004 | PASS | grep -c '^### §4\.[1-4]' audit-*.md | =4 | 4 |
| AC-CLA-004 | PASS | grep -cE '^- Section [1-5]:' audit-*.md | ≥5 | 5 |
| AC-CLA-004 | PASS | grep -c 'SPEC-V3R6-CLI-INTEGRATION-001' audit-*.md | ≥1 | 5 |
| AC-CLA-005 | PASS | git diff --name-only main (9 protected path classes) | =0 | 0 |
| AC-CLA-006 | PASS | grep -cE 'Generated:\|Git SHA:\|moai version:' audit-*.md | ≥3 | 6 |
| AC-CLA-006 | PASS-WITH-DEBT | awk '/§5/,/^## /' \| grep -c '```bash' | ≥1 | 0 |

**Note on AC-CLA-006 PASS-WITH-DEBT**: The awk verification command in acceptance.md L255-257 has a structural defect — the second pattern `/^## /` matches the `## §5 Methodology Appendix` line itself, so awk's range terminates on the same line it starts on (single-line output). The actual audit report DOES contain bash code blocks at lines 335 (§2.2 Reproduction commands) and 590 (§5.1 grep commands used) — the section IS bash-block-rich. The AC defect was not caught at plan-auditor iter-1 (REVISE 0.78, fix-forward applied) nor would have been caught at iter-2 (SKIPPED per user Option A). Sprint 7 SPEC-V3R6-CLI-INTEGRATION-001 manager-spec invocation should NOT consume this AC verification command as-is; instead use `awk '/^## §5 Methodology Appendix/{p=1} /^## §6/{p=0; exit} p' | grep -c '```bash'` which returns 1.

## Sync-phase Evidence (populated at /moai sync)

(Populated by manager-docs after sync phase; left blank intentionally during run-phase.)

## Run-phase Closure

- 4 commits on main: M1 b5ea8d936 + M2 2464ece58 + M3 123ddcb1a + M4 TBD
- All 4 SPEC artifacts frontmatter synced: status `draft → implemented`, version `0.1.0 → 0.2.0`, updated `2026-05-23`
- Audit report at `.moai/reports/cli-audit/audit-2026-05-23.md` complete (667 lines, §1-§6 all populated)
- AC-CLA-005 [Unwanted] research-only constraint verified: 0 protected-path-class diff vs main
- 13/14 AC sub-checks PASS; 1 PASS-WITH-DEBT (AC-CLA-006 awk command defect, report content intact)
- L26 Tier M minimal Section A-E applied + L27 TaskList state isolation respected
- L40 per-SPEC override pattern NOT invoked (default Tier M Hybrid Trunk: feat-branch + auto PR — but executed as main-direct per CLAUDE.local.md §23.7 1-person OSS unified policy)
- Sprint 7 SPEC-V3R6-CLI-INTEGRATION-001 baseline scope (§4.4) ready for direct manager-spec consumption

---

Version: 0.2.0 (run-phase complete — M1-M4 all PASS)
