---
id: SPEC-WORKTREE-DOCS-001
title: "Worktree Workflow Documentation Harmonization (L1/L2/L3 Opt-In Policy)"
version: "0.1.0"
status: draft
created: 2026-05-17
updated: 2026-05-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "worktree, documentation, harmonization, override, opt-in, terminology"
---

# Acceptance Criteria — SPEC-WORKTREE-DOCS-001

All criteria are binary PASS / FAIL. Run-PR cannot merge until every AC reports PASS.

## AC-WTD-001 — Force-Worktree [HARD] Mandate Removal

**Given** the run-PR has been applied to main, with edits to `spec-workflow.md` and
`CLAUDE.md` §14.

**When** the audit command runs:
```
grep -rn '\[HARD\].*\(MUST\|must\).*worktree' .claude/rules/moai/workflow/spec-workflow.md CLAUDE.md
```

**Then** the command MUST return zero lines that mandate creating, reusing, or using
worktrees as a [HARD] requirement.

**Verification command**:
```
test "$(grep -rEn '\[HARD\][^.]*MUST[^.]*worktree' \
  .claude/rules/moai/workflow/spec-workflow.md \
  CLAUDE.md \
  | wc -l)" -eq 0
```

**Binary**: PASS if `$?` is 0; FAIL otherwise.

**Implements**: REQ-WTD-003.

## AC-WTD-002 — Terminology Glossary Present

**Given** the run-PR includes T1 (terminology glossary insertion).

**When** the verification command runs:
```
grep -n "Terminology Glossary" .claude/rules/moai/workflow/worktree-integration.md
```

**Then** the command MUST return at least 1 match, AND the glossary must contain
distinct rows for L1, L2, L3, and "git worktree" (4 row anchors).

**Verification commands**:
```
grep -c "Terminology Glossary" .claude/rules/moai/workflow/worktree-integration.md
grep -cE "^\| (L1|L2|L3|git worktree)" .claude/rules/moai/workflow/worktree-integration.md
```

**Binary**: PASS if first command returns ≥1 AND second command returns exactly 4;
FAIL otherwise.

**Implements**: REQ-WTD-002.

## AC-WTD-003 — Terminology Prefix Application

**Given** the run-PR includes T2 (terminology prefix application).

**When** the audit script runs:
```
scripts/audit-workflow-terminology.sh \
  .claude/rules/moai/workflow/spec-workflow.md \
  .claude/rules/moai/workflow/worktree-integration.md \
  .claude/rules/moai/workflow/worktree-state-guard.md \
  .claude/rules/moai/workflow/session-handoff.md \
  CLAUDE.md
```

**Then** the script MUST exit 0 — every "worktree" prose mention in those 5 files
is either prefixed (L1/L2/L3/git), parenthetically clarified, inside a code/CLI
literal, or within the terminology glossary block.

**Fallback verification** (if script is not available — T0 is SHOULD-have):
```
# Manual grep: count bare "worktree" mentions NOT preceded by L1/L2/L3/git/SPEC/Plan/Claude Code
grep -nE '(^|[^a-zA-Z0-9\-_/.])worktree' <files> \
  | grep -vE '(L1 worktree|L2 worktree|L3 worktree|git worktree|SPEC worktree|Plan worktree|Claude Code Native Worktree)'
```

Manual review of remaining matches confirms each is in a glossary/code/heading context.

**Binary**: PASS if script exits 0 (or manual review yields 0 ambiguous matches);
FAIL otherwise.

**Implements**: REQ-WTD-002.

## AC-WTD-004 — User Policy 2026-05-17 Documentation

**Given** the run-PR has been applied.

**When** the cross-file grep runs:
```
grep -rn "2026-05-17\|feedback_worktree_autonomous" \
  .claude/rules/moai/workflow/spec-workflow.md \
  .claude/rules/moai/workflow/worktree-state-guard.md \
  .claude/rules/moai/workflow/session-handoff.md \
  CLAUDE.md
```

**Then** the command MUST return at least 4 matches (at least 1 per file). Each
match must reference the user policy re-definition OR the new memory file by name.

**Binary**: PASS if total matches ≥ 4 across the 4 listed files; FAIL otherwise.

**Implements**: REQ-WTD-001.

## AC-WTD-005 — Backward Compat: Wave 5 Primitive Preserved

**Given** the run-PR has been applied.

**When** the verification runs:
```
test -f internal/cli/worktree/guard.go && \
  grep -c "func.*[Ss]napshot\|func.*[Vv]erify\|func.*[Rr]estore" internal/cli/worktree/guard.go
```

**Then** the file `internal/cli/worktree/guard.go` MUST still exist AND the three
primitive function declarations MUST be intact (count ≥ 3).

**Also**: `worktree-state-guard.md` MUST contain the substring "dormant" or
"Operational status (2026-05-17)" — proving the dormancy banner was added per T5.

**Verification**:
```
grep -c "Operational status (2026-05-17)\|dormant" \
  .claude/rules/moai/workflow/worktree-state-guard.md
```

**Binary**: PASS if guard.go exists AND function count ≥ 3 AND dormancy grep ≥ 1;
FAIL otherwise.

**Implements**: REQ-WTD-005.

## AC-WTD-006 — Memory Hierarchy Integrity

**Given** the run-PR includes T7 + T8 (memory file creation + index patching).

**When** the verification commands run:
```
F1=~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_worktree_autonomous.md
F2=~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_worktree_never_use.md
F3=~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md

test -f "$F1"
grep -c "^name: feedback-worktree-autonomous" "$F1"
grep -c "Why:\|How to apply:\|Effective from\|Supersedes" "$F1"
head -15 "$F2" | grep -c "SUPERSEDED"
grep -c "feedback_worktree_autonomous" "$F3"
```

**Then**:
- `$F1` MUST exist.
- `$F1` MUST have frontmatter `name: feedback-worktree-autonomous` (exactly 1 match).
- `$F1` body MUST contain all 4 substrings (count ≥ 4).
- `$F2` first 15 lines MUST contain "SUPERSEDED" marker (count ≥ 1).
- `$F3` MUST reference the new memory file by name (count ≥ 1).

**Binary**: PASS if all 5 checks pass; FAIL otherwise.

**Implements**: REQ-WTD-004.

## AC-WTD-007 — No Stale Cross-References Outside Edit Scope (Audit, not blocking)

**Given** the run-PR has been applied.

**When** the cross-file audit runs:
```
grep -rEn 'MUST (use|create|reuse)[^.]*(isolation: "worktree"|worktree)' \
  .claude/skills/ \
  | grep -v 'spec-workflow.md\|worktree-integration.md\|worktree-state-guard.md\|session-handoff.md'
```

**Then** the command MAY return matches in files OUTSIDE the 6 affected files.
Each match becomes a candidate for a follow-up SPEC (NOT in scope for this SPEC).
If matches are found, they are recorded as deferred items in lessons.md.

**Binary**: This AC is NON-BLOCKING — it surfaces follow-up scope, not a gate.
PASS unconditionally; output is informational.

**Implements**: R1 mitigation per plan.md § Risk Analysis.

## AC-WTD-008 — Spec Lint Clean

**Given** all 5 plan-phase artifacts in `.moai/specs/SPEC-WORKTREE-DOCS-001/` are
present (spec.md, plan.md, design.md, acceptance.md, scenarios.md).

**When** the spec lint runs:
```
moai spec lint --strict
```

**Then** the command MUST output `✓ No findings` (or its equivalent zero-finding
status line) for the SPEC-WORKTREE-DOCS-001 directory.

**Binary**: PASS if output contains "No findings" AND exit code is 0; FAIL otherwise.

**Implements**: SPEC frontmatter schema compliance (REQ-SPC-003-006).

## AC-WTD-009 — Go Test Suite Still Passes (No Accidental Code-Touch)

**Given** the run-PR has been applied (documentation-only edits).

**When** the test suite runs:
```
go test ./...
```

**Then** the command MUST exit 0 with no NEW test failures attributable to this
SPEC. Pre-existing flaky tests (e.g., `TestSupervisor_NonZeroExit` ETXTBSY race per
CLAUDE.local.md §18.11) are excluded — if they fail, re-run once.

**Binary**: PASS if `go test ./...` exits 0 (after at most 1 retry of pre-existing
flaky tests); FAIL otherwise.

**Implements**: CC-3 (no code path changes) + R4 mitigation per plan.md § Risk Analysis.

## Performance Gates

### PG-1 — Token-Load Impact (Advisory, not blocking)

**Threshold**: Total post-edit token-cost increase for `paths`-matched rule loads is
expected to be ≤ 800 tokens per active phase (across 5 affected files).

**Measurement**: Compare `wc -c` before/after the run-PR for each affected file;
approximate token count as bytes / 4.

**Binary**: Advisory only. If increase exceeds 1500 tokens, surface as a comment on
the run-PR for review but do not block merge.

## Edge Cases

### EC-1 — User opens session and `feedback_worktree_never_use.md` is auto-loaded BEFORE `feedback_worktree_autonomous.md`

**Behavior**: MEMORY.md is read at session start. If user's `MEMORY.md` lists
`feedback_worktree_never_use` higher (or only-only), Claude may apply the old
policy until it reads the new file.

**Mitigation**: T8 ensures the OLD index entry (if present) carries `[SUPERSEDED by
feedback_worktree_autonomous]` marker. Future agents see the supersede chain
deterministically.

### EC-2 — Wave 5 primitive is invoked manually by an agent prompt during run-phase

**Behavior**: This is supported (manual invocation works). Acceptance gate AC-WTD-005
verifies the binary still exists. Agents can call `moai worktree snapshot/verify/restore`
even though orchestrator skills do not.

### EC-3 — User passes `--worktree` flag despite the new default opt-in policy

**Behavior**: This is supported — the policy is opt-in, not opt-out. plan.md's
`--worktree` flag remains accepted; session-handoff.md Block 0 still applies for
that case.

### EC-4 — Cross-SPEC merge conflict in `spec-workflow.md` with SPEC-V3R4-WORKFLOW-SPLIT-001

**Behavior**: design.md § CC-4 documents this risk. Run-PR rebases against latest
main HEAD before run-phase begins. If the Bundle F SPEC has already split
spec-workflow.md into sub-skills, this SPEC's edits target the appropriate
sub-file post-split.

## Definition of Done

The SPEC is **DONE** when all of the following are true:

1. Plan-PR is merged into main with status `MERGED` per `gh pr view`.
2. Run-PR is merged into main with status `MERGED`.
3. Sync-PR is merged into main with status `MERGED`.
4. All 9 acceptance criteria (AC-WTD-001 through AC-WTD-009) report PASS.
5. The frontmatter `status` field in spec.md is `completed` (post-sync transition).
6. `MEMORY.md` index includes the new `feedback_worktree_autonomous.md` entry.
7. `feedback_worktree_never_use.md` body and (if present) MEMORY.md index entry
   are marked `[SUPERSEDED by feedback_worktree_autonomous]`.
8. CHANGELOG.md (top of file) carries an entry describing the doctrine harmonization.
9. lessons.md captures the policy-shift lesson with category `workflow`.
