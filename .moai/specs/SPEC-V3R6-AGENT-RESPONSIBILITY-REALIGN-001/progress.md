---
id: SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001
title: "SPEC artifact ownership realignment across manager-spec / manager-develop / manager-docs — Lifecycle Progress"
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

# SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001 — Progress (4-phase Lifecycle)

## §E. Lifecycle Status

| Phase | Status | Commit SHA | Timestamp | Notes |
|-------|--------|------------|-----------|-------|
| Plan | audit-ready (TBD post-plan-auditor) | TBD (orchestrator commit) | 2026-05-24T<TBD>Z | 4 SPEC artifacts written, plan-auditor pending |
| Run | not-started | (M1-M3 pending) | — | Tier S 3 milestones (M1 manager-spec / M2 manager-develop / M3 manager-docs + schema doc), bundled OR separate commits per manager-develop discretion |
| Sync | not-started | — | — | B12 self-test pending. CRITICAL: this SPEC's sync will be the FIRST sync to apply the new ownership policy on itself — manager-docs MUST touch only CHANGELOG.md + 4 frontmatter status + progress.md §Sync-phase Audit-Ready Signal (NOT spec.md / plan.md / acceptance.md bodies) |
| Mx | SKIP-eligible (TBD) | — | — | Pure .md edits, 0 .go files in scope → Mx Step C SKIP per `mx-tag-protocol.md` §a IF @MX tag count delta = 0 across 7 EXTEND files |

Status transitions per the NEW Status Transition Ownership Matrix this SPEC defines (forward-looking — the matrix takes effect on this SPEC's own sync-phase):

- `draft` (current — set by manager-spec at plan-phase) → `in-progress` (set by manager-develop on M1 commit start) → `implemented` (set by manager-docs on sync commit) → `completed` (set by manager-docs OR orchestrator on Mx chore commit)

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_complete_at: 2026-05-24T18:50:00Z           # set on orchestrator commit
plan_commit_sha: TBD                              # populated post-push (orchestrator backfill if needed; the plan-PR push commit SHA references this entry)
plan_status: audit-ready                          # PASS at iter-1 with margin +0.125
plan_auditor_iter1_score: 0.875                  # threshold 0.75, margin +0.125
plan_auditor_iter1_verdict: PASS                 # MP-1/2/3 PASS, MP-4 N/A (governance domain). 0 MAJOR defects. 2 SHOULD-FIX (D1 §B.1 file count 5↔7 drift / D2 REQ-006↔AC-006 proxy gap with §D.7 forward-looking compensation) + 1 MINOR (D3 awk regex fragility). Recommendation: commit-as-is. Dimensions: Clarity 0.88 / Completeness 0.95 / Testability 0.85 / Traceability 0.82.
artifact_count: 4                                # spec.md + plan.md + acceptance.md + progress.md
artifact_line_count_total: TBD                   # sum of wc -l on 4 artifacts (target: ~1100-1500 lines aggregate, with spec.md being the largest)
preserve_list_count_at_plan_commit: 7            # verified 2026-05-24 pre-plan: 4 M (config + harness) + 3 ?? (research artifacts + i18n-validator) — exact count may be 7 or 8 depending on i18n-validator presence
preserve_list_pre_plan_snapshot:
  - .moai/config/sections/git-convention.yaml    # M
  - .moai/config/sections/language.yaml          # M
  - .moai/config/sections/quality.yaml           # M
  - .moai/harness/usage-log.jsonl                # M
  - .moai/harness/observations.yaml              # ??
  - .moai/research/anthropic-best-practices-2026-05-24.md  # ??
  - .moai/research/v3.0-redesign-2026-05-23.md   # ??
  - i18n-validator                               # ?? (optional — verify presence at commit time)
l44_pre_plan_fetch: "0 0"                        # verified 2026-05-24 via `git fetch origin main && git rev-list --count --left-right origin/main...HEAD`
l51_self_check_status: PASSED                    # SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001 regex validated per manager-spec L51 protocol
l51_decomposition:
  - SPEC: prefix
  - -V3R6: [A-Z][A-Z0-9]* (V uppercase + 3R6 digit/uppercase mix)
  - -AGENT: [A-Z][A-Z0-9]*
  - -RESPONSIBILITY: [A-Z][A-Z0-9]*
  - -REALIGN: [A-Z][A-Z0-9]*
  - -001: \d{3}$ digit-only tail
  - result: → PASS
audit_origin:
  source: .moai/research/anthropic-best-practices-2026-05-24.md
  findings: F1 (P1 Critical SPEC artifact ownership ambiguity) + F12 (P3 Improvement manager-docs haiku capability mismatch — auto-resolved by F1)
  tier: Tier 2 (audit plan §4)
  precedent_chain: TMD-001 sync commit 009e68c5d (archetype problem) + SIV-001 D-NEW-1 inline-fix pattern (preserved pattern)
sprint_position: Sprint 8 P2 (entry SPEC SIV-001 run-complete pending sync, this SPEC plans the policy that SIV-001 sync will apply)
```

## §E.2 Run-phase Evidence (populated by manager-develop M1-M3 self-verification)

| AC ID | Verification Command | Expected Output | Actual Output | Status |
|-------|----------------------|-----------------|---------------|--------|
| AC-ARR-001 | `grep -c '^## SPEC Artifact Ownership$' .claude/agents/core/manager-spec.md` | `1` | TBD | TBD |
| AC-ARR-002 | `grep -c '^## SPEC Artifact Ownership$' .claude/agents/core/manager-develop.md` | `1` | TBD | TBD |
| AC-ARR-003 | `grep -c '^## SPEC Artifact Ownership$' .claude/agents/core/manager-docs.md` | `1` | TBD | TBD |
| AC-ARR-004 | `grep -c '^## Status Transition Ownership Matrix$' .claude/rules/moai/development/spec-frontmatter-schema.md` | `1` | TBD | TBD |
| AC-ARR-005 | `for agent in manager-spec manager-develop manager-docs; do diff -q .claude/agents/core/$agent.md internal/template/templates/.claude/agents/core/$agent.md \|\| { echo "DRIFT: $agent"; exit 1; }; done; echo "ALL_PAIRS_BYTE_IDENTICAL"` | `ALL_PAIRS_BYTE_IDENTICAL` | TBD | TBD |
| AC-ARR-006 | `go vet ./... 2>&1; echo "vet_exit=$?"; golangci-lint run --timeout=2m 2>&1 \| tail -1` | `vet_exit=0` AND `0 issues.` | TBD | TBD |
| AC-ARR-007 | `awk '/^## Status Transition Ownership Matrix$/,/^## [^S]/' .claude/rules/moai/development/spec-frontmatter-schema.md \| grep -cE '(manager-spec\|manager-develop\|manager-docs)'` | `≥ 6` | TBD | TBD |

Invariant verifications (no AC mapping — per acceptance.md §D.4):

| Invariant | Verification | Expected | Actual | Status |
|-----------|-------------|----------|--------|--------|
| PRESERVE 7-8 unchanged | `git status --porcelain \| wc -l` post-M1-M3 commits | 7-8 (identical to pre-plan snapshot in §E.1) | TBD | TBD |
| No source modification beyond declared 7 files | `git diff --name-only HEAD..HEAD~N` (N = number of M1-M3 commits) | exactly 7-8 paths from §A.2 EXTEND list | TBD | TBD |
| description: field updates per REQ-ARR-007 | `for agent in manager-spec manager-develop manager-docs; do head -10 .claude/agents/core/$agent.md \| grep -E 'description:.*Artifact Ownership' \|\| echo "MISSING: $agent"; done` | no `MISSING:` output | TBD | TBD |
| spec.md ✓ No findings | `go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/spec.md` | `✓ No findings` | TBD | TBD |
| Path-specific staging | `git diff --cached --name-only` post-add | exactly 7-8 paths | TBD | TBD |
| L44 HARD pre-spawn fetch (pre-M1) | `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` | `0 0` | TBD | TBD |
| L44 HARD post-push fetch (post-M3) | same command | `0 0` | TBD | TBD |
| Mirror invariant test (if registered) | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift\|TestLateBranchTemplateMirror' -v` | subtests for `manager-{spec,develop,docs}.md` PASS (if in allowlist) | TBD | TBD |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: TBD                              # set on M-final commit by manager-develop (post-push timestamp)
run_commit_sha: TBD                               # M-final commit SHA (or M1+M2+M3 chain SHAs)
run_status: TBD                                   # implemented when all 7 [HARD] ACs PASS + all invariants PASS
ac_pass_count: TBD                                # target: 7 (all [HARD] AC-ARR-001..007 PASS)
ac_fail_count: TBD                                # target: 0 (no [HARD] AC FAIL)
preserve_list_post_run_count: TBD                 # target: 7-8 (verbatim from §E.1 pre-plan snapshot)
l44_pre_commit_fetch: TBD                         # target: "0 0"
l44_post_push_fetch: TBD                          # target: "0 0"
new_warnings_or_lints_introduced: TBD             # target: 0 (go vet 0 + golangci-lint `0 issues.`)
cross_platform_build:
  linux_amd64: TBD                                # target: pass (go build ./... exit 0)
  darwin_arm64: TBD                               # target: pass (host)
  windows_amd64: TBD                              # target: pass (GOOS=windows GOARCH=amd64 go build ./... exit 0)
total_run_phase_files: TBD                        # target: 7 (3 agent operational sources + 3 template mirrors + 1 schema doc; progress.md update is the 8th but accounted separately as Lifecycle metadata)
m1_to_m3_commit_strategy: TBD                     # bundled (single commit) OR separate (3 commits) per manager-develop discretion within Tier S envelope
```

## §E.4 Sync-phase Audit-Ready Signal (THIS SPEC IS THE FIRST CANARY)

This SPEC is **forward-looking** in that its own sync-phase is the FIRST application of the new ownership policy. manager-docs running `/moai sync SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001` MUST adhere to the new boundaries declared in REQ-ARR-003 + REQ-ARR-006:

- **Allowed**: CHANGELOG.md `[Unreleased]` entry + 4 SPEC artifact frontmatter `in-progress → implemented` transition + this progress.md `§E.4` block update
- **Forbidden**: any modification to spec.md / plan.md / acceptance.md body content (§A through §H sections)

If sync-phase reveals a need to modify SPEC body content (e.g., last-minute REQ rewording, AC clarification), manager-docs MUST return a structured blocker report; orchestrator re-delegates to manager-spec for the body edit; THEN re-invokes manager-docs to complete sync.

```yaml
sync_complete_at: TBD
sync_commit_sha: TBD
sync_status: TBD                                  # completed when CHANGELOG entry + 4 frontmatter status + this §E.4 update done
b12_self_test_a_pre_emission_grep: TBD            # target: 0 (`grep -c 'SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001' CHANGELOG.md` pre-emission)
b12_self_test_a_post_emission_grep: TBD           # target: 1 (post-emission)
b12_self_test_b_ac_count_match: TBD               # target: 7 (acceptance.md SSOT AC count: `grep -cE '^\| \*\*AC-ARR-[0-9]+\*\*' .moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/acceptance.md`)
b12_self_test_c_file_paths_verified: TBD          # target: PASS (manager-docs Read plan.md §A.2 EXTEND entries: 7 files listed, all present in run commit + 1 progress.md = 8 total post-sync)
changelog_entry_position: TBD                     # target: [Unreleased] ### Changed section (this is a governance/policy SPEC, not a bug fix)
frontmatter_status_transitions:
  spec.md: TBD                                    # target: draft → implemented
  plan.md: TBD                                    # target: draft → implemented
  acceptance.md: TBD                              # target: draft → implemented
  progress.md: TBD                                # target: draft → implemented (or directly draft → completed depending on Mx)
canary_compliance_check:
  spec_md_body_modified: TBD                      # MUST be false (frontmatter status field only)
  plan_md_body_modified: TBD                      # MUST be false (frontmatter status field only)
  acceptance_md_body_modified: TBD                # MUST be false (frontmatter status field only)
  rationale: |
    This SPEC defines the policy that manager-docs MUST NOT modify spec/plan/acceptance body content.
    Its own sync is the FIRST test of compliance. If manager-docs modifies any SPEC body section
    here (other than frontmatter status field), the canary FAILS and the new policy is not yet binding
    on this SPEC's lifecycle — record as REGRESSION + blocker report for retrospective audit.
```

## §E.5 Mx-phase Audit-Ready Signal

```yaml
mx_complete_at: TBD
mx_disposition: TBD                              # SKIP-eligible per spec.md §C.3 IF @MX tag count delta = 0 across 7 EXTEND files; else EVALUATE
mx_disposition_rationale: |
  Per .claude/rules/moai/workflow/mx-tag-protocol.md §a, Mx Step C SKIP condition requires:
    (1) .go file count = 0 in commit AND
    (2) @MX tag count delta = 0
  This SPEC modifies ONLY .md files (3 agent operational sources + 3 template mirrors + 1 schema doc = 7 .md files; 0 .go files).
  Condition (1) PASS automatically.
  Condition (2) requires verification: did the new `## SPEC Artifact Ownership` body sections introduce any @MX tags (TODO, NOTE, WARN, ANCHOR, REASON, LEGACY)? Expected: NO — ownership declarations are pure policy text, no code annotations.
  If @MX tag count delta = 0: Mx Step C SKIP per §a.
  If @MX tag count delta > 0: Mx Step C EVALUATE — orchestrator scans the new @MX tags + validates pairing rules (@MX:WARN MUST have paired @MX:REASON, etc.) per `mx-tag-protocol.md`.
mx_tag_count_delta:
  source_agent_files: TBD                         # `grep -c '@MX' .claude/agents/core/manager-{spec,develop,docs}.md` pre vs post — target: 0 delta
  mirror_agent_files: TBD                         # same metric on template mirror — target: 0 delta (mirror parity)
  schema_doc: TBD                                 # `grep -c '@MX' .claude/rules/moai/development/spec-frontmatter-schema.md` pre vs post — target: 0 delta
mx_step_c_verdict: TBD                            # SKIP if delta = 0; EVALUATE-PASS if delta > 0 and all new @MX tags are properly paired
```

## §E.6 4-Phase Lifecycle Close Signal

```yaml
lifecycle_close_at: TBD                           # final closure timestamp (Mx chore commit OR sync commit if Mx SKIP)
final_status: TBD                                 # completed when 4-phase lifecycle closed: plan + run(M1-M3) + sync + Mx
total_commits: TBD                                # target: 3-5 (plan + run [1-3 bundled or separate] + sync + optional Mx-chore if NOT SKIP)
total_push_count: TBD                             # one per commit (Hybrid Trunk 1-person OSS per CLAUDE.local.md §23.7)
sprint_position: Sprint 8 P2 — Audit Tier 2 COMPLETE — Tier S minimal 1-pass cohort 6/6 (IF SUCCESS)
next_action_paste_ready: |
  TBD — populated by manager-docs during sync-phase. Expected pattern (following session-handoff.md 6-block structure):

  ultrathink. Sprint 8 P3 — SIV-001 sync-phase enter (NEW ownership policy in effect).
  applied lessons: project_sprint8_arr001_complete, project_sprint8_siv001_run_complete, anthropic_audit Tier 2 F1+F12 resolution.

  전제 검증:
  1) git log --oneline -1 → <ARR-001-sync-SHA> (Sprint 8 P2 SPEC 4-phase complete)
  2) git fetch origin main && git rev-list --count --left-right origin/main...HEAD → 0 0 (L44 HARD pre-SIV-sync)
  3) grep -c '^## Status Transition Ownership Matrix$' .claude/rules/moai/development/spec-frontmatter-schema.md → 1 (new policy schema doc binding)
  4) grep -c '^## SPEC Artifact Ownership$' .claude/agents/core/manager-docs.md → 1 (new manager-docs boundary binding)

  실행: /moai sync SPEC-V3R6-SPEC-ID-VALIDATION-001 — manager-docs sync-phase MUST comply with new boundary (CHANGELOG + 4 frontmatter + progress.md §Sync-phase Audit-Ready Signal ONLY, NO spec.md / plan.md / acceptance.md body modifications)

  머지 후: Audit Tier 3 (F3 + F9 + F13) plan-phase 결정 OR Sprint 8 P4 entry SPEC 결정 via AskUserQuestion
```

## §E.7 L46 Attribution Discipline (post-merge audit)

Per spec.md §A.4 + audit report §3 F1/F12:

| Test | Pre-SPEC status | Post-SPEC expected | Attribution if persists |
|------|-----------------|-------------------|------------------------|
| TestRuleTemplateMirrorDrift/manager-spec.md (if registered) | PASS | PASS (preserved by AC-ARR-005 mirror parity) | — |
| TestRuleTemplateMirrorDrift/manager-develop.md (if registered) | PASS | PASS (preserved by AC-ARR-005) | — |
| TestRuleTemplateMirrorDrift/manager-docs.md (if registered) | PASS | PASS (preserved by AC-ARR-005) | — |
| Manager agent registry tests (TestBackwardCompatibility, TestAgentFrontmatterAudit, TestAllAgentsInCatalog, TestLoadCatalog, TestEmbeddedTemplates_AgentDefinitions, TestLoadEmbeddedCatalog_Success) | FAIL (pre-existing per TMD-001 §B.2) | FAIL (DEFERRED) | Category B2-B4 — Sprint 8+ SPEC pending; UNRELATED to this SPEC scope |

Net post-SPEC baseline failure count delta: **0** (this SPEC adds zero new tests and resolves zero pre-existing test failures; the policy enforcement is body-level/schema-level, not test-level).

Forward-looking behavioral test (post-merge): SIV-001 sync-phase MUST comply with new manager-docs boundary (no spec.md / plan.md / acceptance.md body modifications). Verified retrospectively per acceptance.md §D.7.

## §E.8 Cross-References

- spec.md §F — REQ-ARR-001..009 canonical SSOT
- plan.md §F — M1-M3 implementation procedure
- acceptance.md §D — AC-ARR-001..007 binary PASS/FAIL matrix + §D.4 invariants + §D.7 forward-looking behavioral verification
- TMD-001 precedent `.moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001/spec.md` — sync-phase scope expansion archetype that motivated this SPEC
- SIV-001 precedent `.moai/specs/SPEC-V3R6-SPEC-ID-VALIDATION-001/spec.md` — L51 sister SPEC pattern + D-NEW-1 inline-fix pattern preserved
- Audit origin `.moai/research/anthropic-best-practices-2026-05-24.md` §3 F1 + F12 + §4 Tier 2
- `.claude/rules/moai/workflow/mx-tag-protocol.md` §a — Mx Step C SKIP/EVALUATE rules
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § B12 — manager-docs CHANGELOG discipline self-test
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field canonical SSOT (will receive new `## Status Transition Ownership Matrix` section in this SPEC's run-phase)
- MEMORY.md L33 (Tier S minimal sustained pattern), L40 (per-SPEC envelope), L44 (HARD pre-spawn fetch), L45 (PRESERVE discipline), L46 (attribution), L48 (spec.md SSOT canonical), L49 (trust-but-verify independent batches), L51 (SPEC ID regex pre-write self-check protocol, SIV-001 sister SPEC), Behavior #2 (Manage Confusion Actively) + Behavior #5 (Maintain Scope Discipline) from moai-constitution.md
