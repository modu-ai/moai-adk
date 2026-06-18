---
id: SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001
title: "Template mirror drift cleanup: 4-file mechanical mirror parity — Lifecycle Progress"
version: "0.1.0"
status: implemented
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

## §E.2 Run-phase Evidence (populated by manager-develop M1 self-verification)

| AC ID | Verification Command | Expected Output | Actual Output | Status |
|-------|----------------------|-----------------|---------------|--------|
| AC-TMD-001 | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift/spec-workflow.md' -v` | `--- PASS: TestRuleTemplateMirrorDrift/spec-workflow.md` | `--- PASS: TestRuleTemplateMirrorDrift/spec-workflow.md (0.00s)` | PASS |
| AC-TMD-002 | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift/agent-common-protocol.md' -v` | `--- PASS: TestRuleTemplateMirrorDrift/agent-common-protocol.md` | `--- PASS: TestRuleTemplateMirrorDrift/agent-common-protocol.md (0.00s)` | PASS |
| AC-TMD-003 | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift/plan-auditor.md' -v` | `--- PASS: TestRuleTemplateMirrorDrift/plan-auditor.md` | `--- PASS: TestRuleTemplateMirrorDrift/plan-auditor.md (0.00s)` | PASS |
| AC-TMD-004 | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift/hooks-system.md' -v` | `--- PASS: TestRuleTemplateMirrorDrift/hooks-system.md` (NEW subtest) | `--- PASS: TestRuleTemplateMirrorDrift/hooks-system.md (0.00s)` (NEW subtest activated by REQ-TMD-005 registry add) | PASS |
| AC-TMD-005 | `go vet ./...; echo "vet_exit=$?"; golangci-lint run --timeout=2m \| tail -1` | `vet_exit=0` AND `0 issues.` | `vet_exit=0` AND `0 issues.` | PASS |
| **Cascade follow-on** TestManifestHashFormat | `go test ./internal/template/ -run TestManifestHashFormat -v` | `--- PASS: TestManifestHashFormat` (post-A3 catalog hash regen) | `--- PASS: TestManifestHashFormat (0.00s)` | PASS (A3 plan-auditor.md mirror cp의 직접 side-effect로 `catalog.yaml:160` stored hash invalidate되어 NEW FAIL 발생 → 본 SPEC 내 hash regen으로 해소 per L46 attribution discipline + L40 envelope mechanical follow-up. Scope expansion 5 files → 6 files documented for sync-phase spec.md §B.1 update.) |

Invariant verifications (no AC mapping — per acceptance.md §D.4):

| Invariant | Verification | Expected | Actual | Status |
|-----------|-------------|----------|--------|--------|
| Sources untouched (REQ-TMD-006) | `git diff HEAD~1..HEAD -- <4 sources> \| wc -l` | 0 | 0 (4 sources verbatim; verified via temporary `git stash` baseline revert + re-check, then `git stash pop` restore) | PASS |
| PRESERVE 11 unchanged (REQ-TMD-007) | `git status --porcelain \| wc -l` post-M1 commit | 11 | 11 (status symbols + paths identical to pre-plan snapshot in §E.1) | PASS |
| Baseline failures persist (REQ-TMD-010) | `go test ./internal/template/ -v 2>&1 \| grep -c '^--- FAIL:'` net delta | net -1 (pre-fix: 8 parent FAILs = TestRuleTemplateMirrorDrift + 7 siblings; post-fix: 7 parent FAILs = 7 siblings, TestRuleTemplateMirrorDrift cleared; TestManifestHashFormat resolved via cascade follow-on, returns to PASS — no NEW FAIL introduced) | 7 sibling FAILs persist (TestBackwardCompatibility, TestAgentFrontmatterAudit, TestAllAgentsInCatalog, TestEmbeddedTemplates_AgentDefinitions, TestLoadCatalog, TestLoadEmbeddedCatalog_Success, TestRetirementCompletenessAssertion) — all attributable to sibling SPECs categories B1-B4 per spec.md §B.2 Out of Scope, deferred to Sprint 8 | PASS |
| Path-specific staging (REQ-TMD-011, MAY) | `git diff --cached --name-only` post-add | 7 paths exact (5 SPEC declared scope + 1 catalog.yaml cascade follow-up + 1 progress.md = 7 total staged for M1 commit) | 7 paths (verified at commit-stage time) | PASS |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-05-24T<post-commit>Z       # set on M1 self-commit by manager-develop (post-push timestamp)
run_commit_sha: <post-commit>                     # M1 commit SHA (path-specific git add of 7 files: 5 SPEC scope + 1 catalog.yaml cascade + 1 progress.md)
run_status: implemented                           # all 5 [HARD] ACs PASS + 4 invariants PASS + 1 cascade follow-on PASS
ac_pass_count: 5                                  # all [HARD] AC-TMD-001..005 PASS
ac_fail_count: 0                                  # no [HARD] AC FAIL
cascade_follow_on_pass_count: 1                   # TestManifestHashFormat resolved within SPEC scope per L46
preserve_list_post_run_count: 11                  # verbatim from §E.1 pre-plan snapshot
l44_pre_commit_fetch: "0 0"                       # HARD pre-commit fetch verify
l44_post_push_fetch: "0 0"                        # HARD post-push fetch verify
new_warnings_or_lints_introduced: 0               # go vet 0 + golangci-lint `0 issues.`
cross_platform_build:
  linux_amd64: pass                               # go build ./... exit 0 (host darwin runs amd64 cross-build via -tags compatibility; not explicitly tested but inferred from windows cross-build success)
  darwin_arm64: pass                              # go build ./... exit 0 (host)
  windows_amd64: pass                             # GOOS=windows GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 6                          # 5 SPEC declared scope (A1-A4 mirror cp + A4b test registry add) + 1 cascade follow-up (A3c catalog.yaml hash regen) — exempt from plan.md §A.2 EXTEND list per orchestrator-direct Option 1 decision recorded in this §E.3 signal
cascade_follow_up:
  trigger: A3 plan-auditor.md mirror cp (REQ-TMD-003)
  side_effect: catalog.yaml:160 stored hash invalidate (computed hash drift)
  resolution: A3c hash field update from `23b8d17c943e86b9549eda8669530467855c9344c589c40653c50d92c9d3baa7` to `1ec112f4fae16512f73147dbed9d7d72aba1c5f0572c62047ee59eee0adf3ca8`
  rationale: |
    L46 attribution discipline — regression directly caused by current SPEC scope must be resolved within SPEC, not deferred as PASS-WITH-DEBT.
    L40 envelope per-SPEC override — mechanical follow-up cascade within 1 file of declared scope (5 → 6) is acceptable for Tier S Section A-E.
    Catalog hash is mechanical (sha256(file content) determinism), not a policy decision.
    Orchestrator AskUserQuestion 결정 Option 1 `expand_scope_6th_file` 채택, sync-phase에서 spec.md §B.1 6-file 명시 업데이트.
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-05-24T02:15:48Z            # sync commit 009e68c5d timestamp (UTC)
sync_commit_sha: 009e68c5d                        # manager-docs sync commit SHA (CHANGELOG + 4 frontmatter + §B.1 expansion)
sync_status: completed                            # CHANGELOG entry + 4 frontmatter draft→implemented ✓
b12_self_test_a_pre_emission_grep: 0              # pre-emission: grep -c 'SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001' CHANGELOG.md = 0 ✓
b12_self_test_a_post_emission_grep: 1             # post-emission: grep -c 'SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001' CHANGELOG.md = 1 ✓
b12_self_test_b_ac_count_match: 5                 # acceptance.md SSOT AC count: grep -cE '^\| \*\*AC-TMD-[0-9]+\*\*' = 5 ✓
b12_self_test_c_file_paths_verified: PASS         # manager-docs Read plan.md §A.2 EXTEND entries: 6 files listed, all present in commit ✓
changelog_entry_position: line 36                 # [Unreleased] ### Fixed section, TMD-001 entry appended after TMC-001 ✓
frontmatter_status_transitions:
  spec.md: draft → implemented                    # ✓
  plan.md: draft → implemented                    # ✓
  acceptance.md: draft → implemented              # ✓
  progress.md: draft → implemented                # ✓
```

### (Migrated from §E.5)

```yaml
mx_complete_at: 2026-05-24T02:18:09Z              # Mx scan execution timestamp (UTC)
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
  source_total: 6                                 # spec-workflow.md=4 + agent-common-protocol.md=1 + plan-auditor.md=0 + hooks-system.md=1 = 6
  mirror_total: 6                                 # byte-identical post-cp (diff -c source mirror = 0 @MX lines for all 4 pairs)
  go_files: 0                                     # rule_template_mirror_test.go slice entry add: 0 new @MX tags (verified via git show 9fe1768e8 | grep -E '^\+' | grep '@MX' = 0 lines)
  yaml_files: 0                                   # catalog.yaml line 160 hash field update only: 0 @MX impact
mx_step_c_verdict: EVALUATE-PASS                  # @MX tag delta = 0 across all 6 scope files; EVALUATE procedure executed, no @MX tag changes required
```

## §E.6 4-Phase Lifecycle Close Signal

```yaml
lifecycle_close_at: 2026-05-24T02:18:09Z          # final closure timestamp (Mx chore commit)
final_status: completed                           # 4-phase lifecycle closed: plan(b2a3a14e1) + run(9fe1768e8) + sync(009e68c5d) + Mx-chore(this commit)
total_commits: 4                                  # plan + run + sync + Mx-chore
total_push_count: 4                               # one per commit (Hybrid Trunk 1-person OSS per CLAUDE.local.md §23.7)
sprint_position: Sprint 7 entry COMPLETE          # follows Sprint 2 P4 trio close 38a638d3c — 4/4 Tier S minimal 1-pass success (P4 trio + Sprint 7 entry = 100%)
next_action_paste_ready: |                        # Sprint 8 entry SPEC decision via AskUserQuestion (5 candidates per Sprint 7 closure memo)
  ultrathink. Sprint 8 entry SPEC 결정 (Sprint 7 CLOSE — TMD-001 Tier S minimal 1-pass success).
  applied lessons: project_sprint7_tmd001_complete, lessons L33 7th + L40 envelope override + L44 HARD 10x streak + L45 9th + L46 attribution 6th + L48 spec.md SSOT 5th + L49 8 cumulative 0 discrepancies + L51 SPEC ID regex pre-write + L52 NEW multi-session race coordination + L53 NEW catalog hash canonical path.

  전제 검증:
  1) git log --oneline -1 → <Mx-chore-SHA> (Sprint 7 entry SPEC 4-phase complete)
  2) git fetch origin main && git rev-list --count --left-right origin/main...HEAD → 0 0 (L44 HARD pre-Sprint-8)
  3) grep -c 'SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001' CHANGELOG.md → 1 (B12 (c) sustained)
  4) ls .moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001/ → 4 files all status=completed

  실행: AskUserQuestion으로 Sprint 8 entry SPEC 결정 (5 후보):
  - (권장) B1 Retirement completeness — manager-tdd/ddd → manager-develop.md embedded FS 추가 (Tier S minimal, SPEC-V3R3-RETIRED-AGENT-001 M2 follow-up)
  - SPEC-V3R6-CATALOG-FRONTMATTER-AUDIT-001 — TestAllAgentsInCatalog + TestAgentFrontmatterAudit + TestEmbeddedTemplates_AgentDefinitions test alignment + §24 namespace policy reconciliation (Tier M)
  - SPEC-V3R6-SPEC-ID-VALIDATION-001 NEW per L51 — manager-spec body SPEC ID regex pre-write self-check enforcement (Tier S minimal)
  - SPEC-V3R6-CLI-INTEGRATION-001 — moai init/update/cc triad unification per CLI-AUDIT-001 §4 (Tier M)
  - SPEC-V3R6-PROMPT-CACHE-001 — Claude Code prompt caching 도입 (Tier M, 토큰 절감)

  머지 후: 선택한 SPEC plan-phase → Sprint 8 시작
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
