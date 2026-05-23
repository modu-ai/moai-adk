---
id: SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001
title: "Template mirror drift cleanup: 4-file mechanical mirror parity — Lifecycle Progress"
version: "0.1.0"
status: draft
created: 2026-05-24
updated: 2026-05-24
author: GOOS행님
priority: P3
phase: "v3.0.0"
module: "internal/template/templates/.claude"
lifecycle: spec-anchored
tags: "template-mirror, drift-fix, sprint-7-entry, tier-s, mechanical-cleanup"
---

# SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001 — Progress (4-phase Lifecycle)

## §E. Lifecycle Status

| Phase | Status | Commit SHA | Timestamp | Notes |
|-------|--------|------------|-----------|-------|
| Plan | audit-ready (TBD post-plan-auditor) | TBD (self-commit) | 2026-05-24T<TBD>Z | 4 SPEC artifacts written, plan-auditor pending |
| Run | not-started | (M1 pending) | — | Tier S single milestone M1 |
| Sync | not-started | — | — | B12 9th self-test pending |
| Mx | EVALUATE-pending | — | — | A4 `.go` registry change → Mx Step C EVALUATE per `mx-tag-protocol.md` §a (NOT SKIP) |

Status transitions:
- `draft` (current) → `audit-ready` (after plan-auditor PASS) → `implemented` (after M1 run-commit) → `completed` (after sync + mx complete)

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_complete_at: 2026-05-24T<TBD>Z              # set on self-commit
plan_commit_sha: <TBD — self-commit SHA>         # populated post-push
plan_status: draft                                # → audit-ready after plan-auditor iter-1 PASS ≥ 0.75
plan_auditor_iter1_score: TBD                    # populated by plan-auditor execution
plan_auditor_iter1_verdict: TBD                  # PASS / FAIL / WARN
artifact_count: 4                                # spec.md + plan.md + acceptance.md + progress.md
artifact_line_count_total: TBD                   # sum of wc -l on 4 artifacts
preserve_list_count_at_plan_commit: 11           # verified 2026-05-24 pre-plan: 8 M + 3 ??
preserve_list_pre_plan_snapshot:
  - .claude/output-styles/moai/einstein.md      # M
  - .claude/output-styles/moai/moai.md           # M
  - .moai/config/sections/git-convention.yaml    # M
  - .moai/config/sections/language.yaml          # M
  - .moai/config/sections/quality.yaml           # M
  - .moai/harness/usage-log.jsonl                # M
  - internal/template/templates/.claude/output-styles/moai/einstein.md  # M
  - internal/template/templates/.claude/output-styles/moai/moai.md      # M
  - .moai/harness/observations.yaml              # ??
  - .moai/research/v3.0-redesign-2026-05-23.md   # ??
  - i18n-validator                               # ??
l44_pre_plan_fetch: "0 0"                        # verified 2026-05-24
l51_self_check_status: PASSED                    # SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001 regex validated
l51_decomposition:
  - SPEC: prefix
  - -V3R6: V[A-Z] + 3R6[A-Z0-9]*
  - -TEMPLATE: [A-Z][A-Z0-9]*
  - -MIRROR: [A-Z][A-Z0-9]*
  - -DRIFT: [A-Z][A-Z0-9]*
  - -001: \d{3}$ tail
```

## §E.2 Run-phase Evidence (TBD — populated by manager-develop M1 self-verification)

| AC ID | Verification Command | Expected Output | Actual Output | Status |
|-------|----------------------|-----------------|---------------|--------|
| AC-TMD-001 | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift/spec-workflow.md' -v` | `--- PASS: TestRuleTemplateMirrorDrift/spec-workflow.md` | TBD | TBD |
| AC-TMD-002 | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift/agent-common-protocol.md' -v` | `--- PASS: TestRuleTemplateMirrorDrift/agent-common-protocol.md` | TBD | TBD |
| AC-TMD-003 | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift/plan-auditor.md' -v` | `--- PASS: TestRuleTemplateMirrorDrift/plan-auditor.md` | TBD | TBD |
| AC-TMD-004 | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift/hooks-system.md' -v` | `--- PASS: TestRuleTemplateMirrorDrift/hooks-system.md` (NEW subtest) | TBD | TBD |
| AC-TMD-005 | `go vet ./...; echo "vet_exit=$?"; golangci-lint run --timeout=2m \| tail -1` | `vet_exit=0` AND `0 issues.` | TBD | TBD |

Invariant verifications (no AC mapping — per acceptance.md §D.4):

| Invariant | Verification | Expected | Actual | Status |
|-----------|-------------|----------|--------|--------|
| Sources untouched (REQ-TMD-006) | `git diff HEAD~1..HEAD -- <4 sources> \| wc -l` | 0 | TBD | TBD |
| PRESERVE 11 unchanged (REQ-TMD-007) | `git status --porcelain \| wc -l` post-M1 commit | 11 | TBD | TBD |
| Baseline failures persist (REQ-TMD-010) | `go test ./internal/template/ -v 2>&1 \| grep -c '^--- FAIL:'` net delta | -3 (4 newly PASS + 1 NEW subtest activation = net -3) | TBD | TBD |
| Path-specific staging (REQ-TMD-011, MAY) | `git diff --cached --name-only` post-add | 5 paths exact | TBD | TBD |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: TBD                              # set on M1 self-commit by manager-develop
run_commit_sha: TBD                               # M1 commit SHA (path-specific git add of 5 files)
run_status: draft                                 # → implemented after all 5 ACs PASS + invariants verified
ac_pass_count: TBD                                # expected: 5 (all [HARD])
ac_fail_count: TBD                                # expected: 0
preserve_list_post_run_count: TBD                 # expected: 11 (verbatim from pre-plan)
l44_pre_commit_fetch: TBD                         # expected: "0 0"
l44_post_push_fetch: TBD                          # expected: "0 0"
new_warnings_or_lints_introduced: TBD             # expected: 0
cross_platform_build:
  linux_amd64: TBD                                # expected: exit 0
  darwin_arm64: TBD                               # expected: exit 0
  windows_amd64: TBD                              # expected: exit 0
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: TBD                             # set on manager-docs CHANGELOG entry commit
sync_commit_sha: TBD                              # manager-docs commit SHA
sync_status: pending                              # → completed after CHANGELOG entry + 4 frontmatter draft→implemented
b12_self_test_a_pre_emission_grep: TBD            # expected: 0 (grep -c 'SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001' CHANGELOG.md pre-emission)
b12_self_test_a_post_emission_grep: TBD           # expected: 1
b12_self_test_b_ac_count_match: TBD               # expected: 5 (matches acceptance.md SSOT AC-TMD-001..005)
b12_self_test_c_file_paths_verified: TBD          # expected: PASS (manager-docs Read every plan.md §A.2 EXTEND entry)
changelog_entry_position: TBD                     # expected: under [Unreleased] ### Fixed
frontmatter_status_transitions:
  spec.md: TBD                                    # expected: draft → implemented
  plan.md: TBD                                    # expected: draft → implemented
  acceptance.md: TBD                              # expected: draft → implemented
  progress.md: TBD                                # expected: draft → implemented
```

## §E.5 Mx-phase Audit-Ready Signal

```yaml
mx_complete_at: TBD
mx_disposition: EVALUATE                          # per spec.md §C.3: A4 `.go` registry change present, SKIP NOT eligible
mx_disposition_rationale: |
  Per .claude/rules/moai/workflow/mx-tag-protocol.md §a, Mx Step C SKIP condition requires:
    (1) .go file count = 0 in commit AND
    (2) @MX tag count delta = 0
  This SPEC modifies internal/template/rule_template_mirror_test.go (1 .go file in scope).
  Therefore SKIP NOT eligible.
  Mx Step C EVALUATE: verify @MX tag count source vs mirror delta = 0 across 4 .md mirror pairs
  AND verify the 1 .go registry change does NOT introduce new high fan_in @MX:ANCHOR candidates
  (rule_template_mirror_test.go already has explicit @MX:ANCHOR and @MX:NOTE at lines 24-26 and
  88-89; adding 1 slice entry does not change function fan_in or surface area).
mx_tag_count_delta:
  source_total: TBD                               # expected: 0 (4 .md sources unchanged per REQ-TMD-006)
  mirror_total: TBD                               # expected: matches source (4 .md mirrors now byte-identical)
  go_files: TBD                                   # expected: 0 new tags (registry add is mechanical slice insertion)
mx_step_c_verdict: TBD                            # expected: EVALUATE-PASS (no @MX tag changes required)
```

## §E.6 4-Phase Lifecycle Close Signal

```yaml
lifecycle_close_at: TBD
final_status: draft                               # → completed after all 4 phases close + status transition
total_commits: TBD                                # expected: 3 (plan self-commit + M1 run + manager-docs sync)
                                                  # OR 4 if Mx orchestrator-direct chore needed (likely yes for EVALUATE-PASS frontmatter finalization)
total_push_count: TBD                             # expected: 3-4 pushes (one per commit, hybrid trunk)
sprint_position: Sprint 7 entry                   # follows Sprint 2 P4 trio close 38a638d3c
next_action_paste_ready: TBD                      # populated by manager-docs sync (paste-ready resume message for next session)
```

## §E.7 L46 Attribution Discipline (post-merge audit)

Per spec.md §A.4 + REQ-TMD-010:

| Test | Pre-SPEC status | Post-SPEC expected | Attribution if persists |
|------|-----------------|-------------------|------------------------|
| TestRuleTemplateMirrorDrift/spec-workflow.md | FAIL | PASS (cleared by AC-TMD-001) | — |
| TestRuleTemplateMirrorDrift/agent-common-protocol.md | FAIL | PASS (cleared by AC-TMD-002) | — |
| TestRuleTemplateMirrorDrift/plan-auditor.md | FAIL | PASS (cleared by AC-TMD-003) | — |
| TestRuleTemplateMirrorDrift/hooks-system.md | (not registered) | PASS (NEW subtest, cleared by AC-TMD-004) | — |
| TestRuleTemplateMirrorDrift/* (6 other subtests) | PASS | PASS (unchanged) | — |
| TestLateBranchTemplateMirror/* | PASS (post-TMC-001 close) | PASS (unchanged) | — |
| TestBackwardCompatibility | FAIL | FAIL (DEFERRED) | Category B2 — Sprint 8 SPEC pending |
| TestAgentFrontmatterAudit | FAIL | FAIL (DEFERRED) | Category B3 — Sprint 8 SPEC pending |
| TestAllAgentsInCatalog | FAIL | FAIL (DEFERRED) | Category B2 — Sprint 8 SPEC pending |
| TestEmbeddedTemplates_AgentDefinitions | FAIL | FAIL (DEFERRED) | Category B4 — Sprint 8 SPEC pending |
| TestLoadCatalog | FAIL | FAIL (DEFERRED) | Category B2 — Sprint 8 SPEC pending |
| TestLoadEmbeddedCatalog_Success | FAIL | FAIL (DEFERRED) | Category B4 — Sprint 8 SPEC pending |
| TestRetirementCompletenessAssertion × 2 | FAIL | FAIL (DEFERRED) | Category B1 — Sprint 8 SPEC pending |

Net post-SPEC baseline failure count delta: **-3** (4 PASS gains + 1 NEW subtest = net 3 subtests removed from FAIL count).

## §E.8 Cross-References

- spec.md §F — REQ-TMD-001..011 canonical SSOT
- plan.md §F — M1 implementation procedure
- acceptance.md §D — AC-TMD-001..005 binary PASS/FAIL matrix + §D.4 invariants
- TMC-001 precedent `.moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001/progress.md` — Tier S 4-phase lifecycle pattern
- `.claude/rules/moai/workflow/mx-tag-protocol.md` §a — Mx Step C SKIP/EVALUATE rules
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § B12 — manager-docs CHANGELOG discipline self-test
- MEMORY.md L33 (Tier S minimal sustained pattern), L40 (per-SPEC envelope), L44 (HARD pre-spawn fetch), L45 (PRESERVE discipline), L46 (attribution), L48 (spec.md SSOT canonical), L49 (trust-but-verify independent batches), L51 (proposed SPEC ID regex pre-write self-check)
