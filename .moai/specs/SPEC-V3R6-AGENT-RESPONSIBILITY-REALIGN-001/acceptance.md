---
id: SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001
title: "SPEC artifact ownership realignment across manager-spec / manager-develop / manager-docs — Acceptance Criteria"
version: "0.1.0"
status: draft
created: 2026-05-24
updated: 2026-05-24
author: GOOS행님
priority: P1
phase: "v3.0.0"
module: ".claude/agents/core"
lifecycle: spec-anchored
tags: "agent-ownership, soc, manager-spec, manager-develop, manager-docs, status-transition, schema, audit-tier-2, anthropic-best-practice"
---

# SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001 — Acceptance Criteria

## §D. AC Matrix (canonical PASS/FAIL gates)

This file is the **canonical AC SSOT**. Every AC has a single verifiable command and a single expected output. progress.md §Run-phase Evidence references this matrix verbatim.

| AC ID | Severity | REQs Covered | Verification Command | Expected Output | Status |
|-------|----------|--------------|----------------------|-----------------|--------|
| **AC-ARR-001** | [HARD] | REQ-ARR-001 | `grep -c '^## SPEC Artifact Ownership$' .claude/agents/core/manager-spec.md` | `1` | TBD |
| **AC-ARR-002** | [HARD] | REQ-ARR-002 | `grep -c '^## SPEC Artifact Ownership$' .claude/agents/core/manager-develop.md` | `1` | TBD |
| **AC-ARR-003** | [HARD] | REQ-ARR-003 | `grep -c '^## SPEC Artifact Ownership$' .claude/agents/core/manager-docs.md` | `1` | TBD |
| **AC-ARR-004** | [HARD] | REQ-ARR-004 | `grep -c '^## Status Transition Ownership Matrix$' .claude/rules/moai/development/spec-frontmatter-schema.md` | `1` | TBD |
| **AC-ARR-005** | [HARD] | REQ-ARR-005 | `for agent in manager-spec manager-develop manager-docs; do diff -q .claude/agents/core/$agent.md internal/template/templates/.claude/agents/core/$agent.md \|\| { echo "DRIFT: $agent"; exit 1; }; done; echo "ALL_PAIRS_BYTE_IDENTICAL"` | `ALL_PAIRS_BYTE_IDENTICAL` (3 source-mirror pairs all byte-identical) | TBD |
| **AC-ARR-006** | [HARD] | REQ-ARR-006 (negative; vet/lint gate) | `go vet ./... 2>&1; echo "vet_exit=$?"; golangci-lint run --timeout=2m 2>&1 \| tail -1` | `vet_exit=0` AND `0 issues.` | TBD |
| **AC-ARR-007** | [HARD] | REQ-ARR-004 (matrix completeness) | `awk '/^## Status Transition Ownership Matrix$/,/^## [^S]/' .claude/rules/moai/development/spec-frontmatter-schema.md \| grep -cE '(manager-spec\|manager-develop\|manager-docs)'` | `≥ 6` (matrix has at least 6 rows mentioning the 3 manager agents — covering draft / draft→in-progress / in-progress→implemented / implemented→completed / →superseded / →archived / →rejected transitions) | TBD |

## §D.1 Severity convention

- **[HARD]**: Must-pass. FAIL on any [HARD] AC blocks SPEC closure; manager-develop returns blocker report.
- **[SHOULD]**: Should-pass. FAIL with documented PASS-WITH-DEBT acceptable in rare cases (none in this SPEC).
- **[MAY]**: Optional verification, no closure blocking.

All 7 ACs in this SPEC are [HARD]. PASS-WITH-DEBT is NOT acceptable.

## §D.2 Status enum

- **TBD**: Pre-run-phase placeholder (default in plan-phase).
- **PASS**: Verification command produced expected output.
- **FAIL**: Verification command did NOT produce expected output → blocker report obligation.
- **PASS-WITH-DEBT**: Documented baseline failure attributable to sibling SPEC per L46 (NOT applicable to this SPEC's 7 ACs).

## §D.3 Coverage to REQ-ARR-001..009 traceability

| REQ-ARR | Covered by AC | Notes |
|---------|---------------|-------|
| REQ-ARR-001 (manager-spec ownership section) | AC-ARR-001 | Direct grep on section heading |
| REQ-ARR-002 (manager-develop ownership section) | AC-ARR-002 | Direct grep on section heading |
| REQ-ARR-003 (manager-docs ownership section) | AC-ARR-003 | Direct grep on section heading |
| REQ-ARR-004 (schema doc transition matrix) | AC-ARR-004 + AC-ARR-007 | AC-ARR-004 = heading presence; AC-ARR-007 = matrix completeness (≥6 rows) |
| REQ-ARR-005 (mirror parity for 3 agents) | AC-ARR-005 | Direct `diff -q` 3 pairs |
| REQ-ARR-006 (ownership crossing — Unwanted Behavior) | AC-ARR-006 indirect | Cannot directly test "manager-docs did not modify spec.md" at static analysis time; the negative behavior is policy enforced via the new ownership sections themselves (REQ-ARR-001/002/003) + future hook enforcement (REQ-ARR-009 OPTIONAL). AC-ARR-006 vet/lint gate catches structural regressions; behavioral compliance verified through future SPEC sync-phase observation (e.g., the next post-merge SPEC sync-phase MUST follow the new boundary, observed retrospectively). |
| REQ-ARR-007 (frontmatter description: update) | (verified by §D.4 indirect) | No AC mapping — description updates are cosmetic + verified by `grep` on each agent's frontmatter ‘description:’ field referencing the new section name. Per §C.1 decision rule (matches L48 precedent: REQ-SARM-007 / REQ-TMC-007 / REQ-TMD-007 / REQ-SIV-005). |
| REQ-ARR-008 (CLAUDE.md narrative update — OPTIONAL MAY) | (no AC mapping) | OPTIONAL per spec.md §B.2 — no AC required per §C.1 decision rule. |
| REQ-ARR-009 (PostToolUse hook enforcement — OPTIONAL MAY) | (no AC mapping) | OPTIONAL per spec.md §B.2 — deferred to follow-up SPEC if desired. |

## §D.4 Indirect verification (no AC mapping but required for closure)

These checks are NOT in the AC matrix because they are **invariants** rather than test outcomes. manager-develop verifies them as part of M-final self-verification per plan.md §E:

| Invariant | Verification | Expected |
|-----------|-------------|----------|
| PRESERVE list unchanged | `git status --porcelain` pre vs post run-phase commits | 7-8 entries identical status |
| No source modification beyond declared 7 files | `git diff --name-only HEAD..HEAD~1` (or N commits depending on M1-M3 bundling) | exactly 7 files: 3 agent operational sources + 3 template mirrors + 1 schema doc + 1 progress.md = max 8 paths |
| `description:` field updates per REQ-ARR-007 | `for agent in manager-spec manager-develop manager-docs; do head -10 .claude/agents/core/$agent.md \| grep -E 'description:.*Artifact Ownership' \|\| echo "MISSING: $agent"; done` | no `MISSING:` output (all 3 description fields reference the new section) |
| spec.md `✓ No findings` | `go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/spec.md` | `✓ No findings` |
| Path-specific staging | `git diff --cached --name-only` post-add | exactly 7-8 paths from §A.2 EXTEND list (7 run-phase files + 1 progress.md) |
| L44 HARD pre-spawn fetch | `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` | `0 0` pre-spawn AND `0 0` post-push |
| Mirror invariant test (if registered) | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift\|TestLateBranchTemplateMirror' -v` | all subtests for `manager-{spec,develop,docs}.md` PASS (if in allowlist) or absent (if not — REQ-ARR-005 is the primary mirror parity gate via `diff -q`) |

## §D.5 Closure gate

SPEC closure (status transition `in-progress → implemented`) requires:

1. All 7 [HARD] ACs in §D matrix = PASS
2. All 7 invariants in §D.4 verified
3. `🗿 MoAI` trailer present in all run-phase commits (M1-M3 bundled or separate)
4. `git push origin main` succeeded (L44 post-push fetch = `0 0`)
5. manager-develop returned structured self-verification report per plan.md §E

If any of the 5 gates fail, SPEC remains `in-progress` and manager-develop returns blocker report to orchestrator.

## §D.6 Verification command reproducibility

All verification commands in §D are reproducible from project root `/Users/goos/MoAI/moai-adk-go`. No environment variables required. Commands assume `go test`, `go vet`, `golangci-lint`, `git`, `grep`, `awk`, `diff` are on PATH.

## §D.7 Behavior verification (forward-looking, post-merge)

Because REQ-ARR-006 (ownership crossing — Unwanted Behavior) cannot be statically verified at SPEC closure time, behavioral compliance is verified retrospectively on the **next SPEC sync-phase after this SPEC merges**:

| Forward-looking check | Expected outcome | Verification method |
|-----------------------|-----------------|---------------------|
| Next sync-phase by manager-docs touches SPEC body files (spec.md / plan.md / acceptance.md) | `manager-docs` returns blocker report + orchestrator re-delegates to `manager-spec` for SPEC body edits, THEN re-invokes `manager-docs` for CHANGELOG | `git log --author=...` AND `git show --stat <sync-commit>` — sync-commit MUST NOT include changes to spec.md / plan.md / acceptance.md bodies (frontmatter status field update is the only acceptable spec.md change in sync-phase) |
| Next run-phase by manager-develop encounters AC inadequacy | `manager-develop` returns blocker report; orchestrator re-delegates to `manager-spec` for AC tightening; THEN re-delegates to `manager-develop` to continue M-N | observed via orchestrator turn structure (manager-spec invocation between two manager-develop invocations within the same SPEC run-phase) |
| SIV-001 sync-phase (next sync after this SPEC merges) | manager-docs touches only CHANGELOG.md + 4 frontmatter status fields + progress.md §Sync-phase Audit-Ready Signal | retrospective `git show` on the SIV-001 sync commit |

These forward-looking checks are documented for completeness but are NOT closure gates for this SPEC (they apply to FUTURE SPECs that benefit from this SPEC's policy).
