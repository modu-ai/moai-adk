# Progress — SPEC-V3R6-RULES-CATALOG-SCRUB-001

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID regex self-check: `decomposition: SPEC ✓ | V3R6 ✓ | RULES ✓ | CATALOG ✓ | SCRUB ✓ | 001 ✓ → PASS`
- Plan-phase artifacts authored: spec.md + plan.md + acceptance.md + progress.md (4-file set)
- Frontmatter: 12 canonical fields present; `era: V3R6` set explicitly (plan-phase pre-populated progress.md avoids H-2 transient misclassification)
- Exclusions section: present (`spec.md` §B Exclusions — PRESERVE 4 reference files + role_profiles)
- Defect inventory: 15 defect groups (D1-D15) across 18 files, each with verified file:line + canonical replacement
- Milestone grouping: M1 (P0 ci-protocols) → M2 (P0 agent-authoring) → M3 (core/) → M4 (workflow/) → M5 (development/) → M6 (languages/) → M7 (design SHOULD) → M-final (build + verify)
- AC count: 21 (AC-RCS-001..021), each grep-verifiable
- Risk decisions: D14 design/constitution scoped as SHOULD with carve-out-note default (no FROZEN-zone mechanical rename); D8 agent-recount → role-based language replacing brittle count
- Template-First: all 18 target files confirmed to have mirrors; M-final runs `make build` + neutrality test
- status: draft (initial)

## §E.2 Run-phase Evidence

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-RCS-001 (D4 ci-autofix manager-quality live-spawn) | PASS | `grep -nE 'manager-quality' .../ci-autofix-protocol.md \| grep -vE 'archived .manager-quality\|replaced by'` | no live-spawn line (1 documentary carve-out at L107 remains) |
| AC-RCS-002 (D5 ci-watch manager-quality) | PASS | `grep -cE 'manager-quality' .../ci-watch-protocol.md` | 0 |
| AC-RCS-003 (D10 agent-authoring 8-agent catalog) | PASS | `grep -cE 'expert-(backend\|frontend\|security\|devops\|performance\|refactoring)\|manager-(strategy\|quality\|brain\|project)' .../agent-authoring.md` | 0 (Explore present: 3) |
| AC-RCS-004 (D1 agent-hooks archived rows) | PASS | `grep -cE 'expert-backend\|expert-frontend\|expert-devops\|manager-quality' .../agent-hooks.md` | 1 (documentary "removed during consolidation" note; 0 live table rows) |
| AC-RCS-005 (D2 agent-common-protocol) | PASS | `grep -c 'expert-\*' + grep -cE backtick-form manager-quality diagnostic` | 0 + 0 (Explore: 2) |
| AC-RCS-006 (D6 agent-teams-pattern dead path) | PASS | `grep -c 'manager-strategy' .../agent-teams-pattern.md` | 0 |
| AC-RCS-007 (D7 spec-workflow phantom team agents) | PASS | `grep -cE 'team-validator\|team-tester\|backend-dev\|frontend-dev' .../spec-workflow.md` | 0 |
| AC-RCS-008 (D8 worktree-integration archived names) | PASS | `grep -cE 'expert-backend\|expert-frontend\|expert-refactoring' .../worktree-integration.md` | 0 (role_profile lines L83/166/210/236 preserved) |
| AC-RCS-009 (D9 worktree-state-guard escalation target) | PASS | `grep -c 'claude-code-guide' + grep -c 'Explore'` | 0 + 1 |
| AC-RCS-010 (D11 orchestrator-templates invalid subagent_type) | PASS | `grep -cE 'subagent_type: *"(analyzer\|designer\|implementer\|reviewer)"'` | 0 |
| AC-RCS-011 (D12 model-policy archived in tier cells) | PASS | `grep -cE '(\| \| , )(manager-)?(strategy\|quality\|researcher)(,\| \|)'` | 0 (role-profile/retained: 3) |
| AC-RCS-012 (D13 language boilerplate 16-file parity) | PASS | `for f in 16 langs: grep -c 'manager-quality'` | all 16 = 0 |
| AC-RCS-013 (D14 design/constitution carve-out) | PASS | `grep -c 'archived-agent-rejection' .../design/constitution.md` | 1 (carve-out note at L20) |
| AC-RCS-014 (D3 zone-registry cross-reference) | PASS | `grep -c 'archived-agent-rejection\|general-purpose' .../zone-registry.md` | 1 (CONST-V3R2-064 clause inline cross-ref) |
| AC-RCS-015 (D15 skill-authoring expert-backend) | PASS | `grep -c 'expert-backend' .../skill-authoring.md` | 0 (replaced with manager-develop) |
| AC-RCS-016 (PRESERVE files intact) | PASS | `grep -cE archived-name + git diff --name-only` | 4 files unchanged (archived enumerations preserved) |
| AC-RCS-017 (role_profiles unchanged) | PASS | `grep -rn 'role_profile' worktree-integration.md zone-registry.md` | researcher/analyst/reviewer/implementer/tester/designer present |
| AC-RCS-018 (template mirror parity) | PASS | Group 1: 17× `diff -q` (all PARITY-OK); Group 2: `go test -run TestRuleTemplateMirrorDrift` ok; Group 3: `go test -run TestTemplateNoInternalContentLeak` ok | all PASS |
| AC-RCS-019 (template neutrality) | PASS | `go test -run TestTemplateNeutralityAudit` | ok |
| AC-RCS-020 (make build + go build) | PASS | `make build && go build ./...` | make build exit 0; go build ./... exit 0 |
| AC-RCS-021 (catalog-wide residual) | PASS | `grep -rlE archived-names \| grep -v PRESERVE-4` | residual lines are documentary/carve-out only (no live spawn/example/hook-action); agent-common-protocol L37 PRESERVE per SPEC §B |
| spec-lint | PASS | `bin/moai spec lint .../spec.md` | ✓ No findings — all SPEC documents are valid |

Notes:
- Worktree isolation: Claude Code runtime auto-materialized worktree `agent-ab3121171d8e4c2d4` (locked). Per user directive (worktree-avoidance), all edits applied in-worktree; shared-main propagation deferred to orchestrator (`git diff → git apply`).
- 3 files diverged from shared-main at worktree creation (ci-watch-protocol, worktree-integration, languages/cpp — all touched by the earlier SPEC-V3R6-RULES-HOTFIX-001 M1 commit). These were re-synced from shared main before editing (file-content copy, NOT `git reset --hard` which is prohibited) so the scrub applies against the current shared-main baseline.
- Neutrality iteration: initial draft of agent-hooks.md L54 + agent-authoring.md §Agent Categories carried a `SPEC-V3R6-AGENT-TEAM-REBUILD-001` reference that `TestTemplateNoInternalContentLeak` (class=C1-spec-id-prefix) rejected; both were de-SPEC-IDed to generic "catalog consolidation" prose and the leak test re-passed.
- settings-management.md L32 `expert-frontend`/`team-designer` was a SHOULD-grade residue (SPEC §B Out of Scope); scrubbed to `Agent(general-purpose)` frontend specialist + designer role_profile to clear the AC-RCS-021 catalog-wide scan cleanly.
- worktree-state-guard.md L28 `expert-refactoring` (snapshot example list, not the D9 L62 target) was also scrubbed to `Agent(general-purpose)` refactoring specialist for the same AC-RCS-021 cleanliness.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-19
run_commit_sha: "<pending — commit to be created; orchestrator applies diff to shared main>"
run_status: complete
ac_pass_count: 21
ac_fail_count: 0
preserve_list_post_run_count: 4   # archived-agent-rejection.md, NOTICE.md, agent-patterns.md, spec-frontmatter-schema.md — all unchanged
l44_pre_commit_fetch: n/a         # worktree isolation; orchestrator owns shared-main propagation
l44_post_push_fetch: n/a          # push deferred to orchestrator
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: pass              # `go build ./...` exit 0
  windows_amd64: not-run          # no Go code changed in this SPEC (rules-content scrub only); cross-platform build tag risk does not apply
total_run_phase_files: 46         # 23 deployed + 23 template-mirror edits (22 distinct files; agent-common-protocol/zone-registry edited on both sides individually due to §25-divergence)
m1_to_mN_commit_strategy: single-run-commit   # M1-M7 applied as one logical scrub; will land as one commit (fix(SPEC-V3R6-RULES-CATALOG-SCRUB-001): M1 ...)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_

## §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase — owned by manager-docs / orchestrator>_
