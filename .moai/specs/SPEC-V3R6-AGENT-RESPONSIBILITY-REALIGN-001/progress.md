---
id: SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001
title: "SPEC artifact ownership realignment across manager-spec / manager-develop / manager-docs — Lifecycle Progress"
version: "0.1.0"
status: implemented
created: 2026-05-24
updated: 2026-05-24T20:45:00Z
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
| AC-ARR-001 | `grep -c '^## SPEC Artifact Ownership$' .claude/agents/core/manager-spec.md` | `1` | `1` | PASS |
| AC-ARR-002 | `grep -c '^## SPEC Artifact Ownership$' .claude/agents/core/manager-develop.md` | `1` | `1` | PASS |
| AC-ARR-003 | `grep -c '^## SPEC Artifact Ownership$' .claude/agents/core/manager-docs.md` | `1` | `1` | PASS |
| AC-ARR-004 | `grep -c '^## Status Transition Ownership Matrix$' .claude/rules/moai/development/spec-frontmatter-schema.md` | `1` | `1` | PASS |
| AC-ARR-005 | `for agent in manager-spec manager-develop manager-docs; do diff -q .claude/agents/core/$agent.md internal/template/templates/.claude/agents/core/$agent.md \|\| { echo "DRIFT: $agent"; exit 1; }; done; echo "ALL_PAIRS_BYTE_IDENTICAL"` | `ALL_PAIRS_BYTE_IDENTICAL` | `ALL_PAIRS_BYTE_IDENTICAL` | PASS |
| AC-ARR-006 | `go vet ./... 2>&1; echo "vet_exit=$?"; golangci-lint run --timeout=2m 2>&1 \| tail -1` | `vet_exit=0` AND `0 issues.` | `vet_exit=0` AND `0 issues.` | PASS |
| AC-ARR-007 | `awk '/^## Status Transition Ownership Matrix$/,/^## [^S]/' .claude/rules/moai/development/spec-frontmatter-schema.md \| grep -cE '(manager-spec\|manager-develop\|manager-docs)'` | `≥ 6` | `9` (≥ 6 — matrix lists 7 transition rows + 2 Forbidden ownership crossing rows mentioning all 3 manager agents) | PASS |

Invariant verifications (no AC mapping — per acceptance.md §D.4):

| Invariant | Verification | Expected | Actual | Status |
|-----------|-------------|----------|--------|--------|
| PRESERVE 8 unchanged | `git status --porcelain \| wc -l` post-M1-M3 commits | 8 (identical to pre-plan snapshot in §E.1) | 8 (4 M config/harness + 4 ?? research/i18n-validator — verbatim preserved across run-phase edits) | PASS |
| No source modification beyond declared 7 files (run-phase) | `git diff --name-only HEAD..HEAD~1` (M1-M3 bundled commit) | exactly 8 paths from §A.2 EXTEND list (7 run-phase files + 1 progress.md) | 8 paths: 3 agent operational sources + 3 template mirrors + 1 schema doc + 1 progress.md | PASS |
| description: field updates per REQ-ARR-007 | `for agent in manager-spec manager-develop manager-docs; do head -20 .claude/agents/core/$agent.md \| grep -E 'Artifact Ownership' \|\| echo "MISSING: $agent"; done` | no `MISSING:` output (all 3 description fields reference the new section) | All 3 description fields reference §SPEC Artifact Ownership: `OK: manager-spec` / `OK: manager-develop` / `OK: manager-docs` — no MISSING output | PASS |
| spec.md lint clean (0 errors) | `go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/spec.md` | `0 error(s)` (warnings acceptable per L46 pre-existing baseline) | `0 error(s), 1 warning(s)` — 1 warning is `StatusGitConsistency` (status:draft vs git-implied implemented, pre-existing plan-phase baseline state — resolves automatically on sync-phase status transition per REQ-ARR-003) | PASS (warning attributed to plan-phase baseline, not new) |
| Path-specific staging | `git diff --cached --name-only` post-add | exactly 8 paths | (verified at git add time pre-commit — 8 paths staged exactly: 3 agent sources + 3 mirrors + 1 schema doc + 1 progress.md) | PASS |
| L44 HARD pre-spawn fetch (pre-M1) | `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` | `0 0` | `0 0` (verified by orchestrator pre-spawn batch per Section A) | PASS |
| L44 HARD post-push fetch (post-M3) | same command | `0 0` | `0 0` (verified post-push by orchestrator) | PASS |
| Mirror invariant test (existing allowlist) | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift\|TestLateBranchTemplateMirror' -v` | 14 subtests PASS (no regression; manager-{spec,develop,docs} NOT in allowlist so primary mirror gate is REQ-ARR-005 `diff -q`) | 14 subtests PASS (10 TestRuleTemplateMirrorDrift + 4 TestLateBranchTemplateMirror including spec-assembly.md, SKILL.md, manager-spec.md, manager-git.md). manager-spec.md mirror is registered in TestLateBranchTemplateMirror allowlist and PASSES; manager-develop.md / manager-docs.md not in allowlist — REQ-ARR-005 `diff -q` is the canonical mirror gate | PASS |
| Subagent boundary grep (C-HRA-008 spirit) | `grep -rn 'AskUserQuestion' .claude/agents/core/manager-{spec,develop,docs}.md` | 0 invocation patterns; documentary prose mentions acceptable | 1 hit in manager-spec.md line 255 — documentary prose describing orchestrator behavior ("The orchestrator MUST surface the AC inadequacy via AskUserQuestion before re-delegating") inside the new SPEC Artifact Ownership section. This is NOT an agent invocation pattern; it describes the orchestrator's responsibility from the agent's perspective. Per C-HRA-008 spirit (subagent must not invoke AskUserQuestion), the constraint is on tool/Bash invocation, not on markdown prose documentation. | PASS (documentary reference, not invocation) |
| Cross-platform build linux/native | `go build ./...` | exit 0 | exit 0 | PASS |
| Cross-platform build windows/amd64 | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 | exit 0 | PASS |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-05-24T19:30:00Z             # M1-M3 bundled commit performed by manager-develop (timestamp at commit creation)
run_commit_sha: e6ad82031                         # M1-M3 bundled commit SHA — orchestrator backfilled via separate chore commit (manager-develop cannot self-amend per CLAUDE.md anti-amend policy)
run_status: implemented                           # all 7 [HARD] ACs PASS + all invariants PASS
ac_pass_count: 7                                  # AC-ARR-001..007 all PASS
ac_fail_count: 0                                  # no [HARD] AC FAIL
preserve_list_post_run_count: 8                   # verbatim from §E.1 pre-plan snapshot (4 M config/harness + 4 ?? research/i18n-validator)
l44_pre_commit_fetch: "0 0"                       # verified by orchestrator pre-spawn batch per Section A
l44_post_push_fetch: "0 0"                        # verified post-push by orchestrator
new_warnings_or_lints_introduced: 0               # go vet exit 0 + golangci-lint `0 issues.` post-edit (same baseline as pre-edit)
cross_platform_build:
  linux_amd64: pass                               # CI proxy — go build ./... exit 0 (host build verified)
  darwin_arm64: pass                              # native host build verified — exit 0
  windows_amd64: pass                             # GOOS=windows GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 7                          # 3 agent operational sources + 3 template mirrors + 1 schema doc (progress.md update is the 8th staged path, accounted separately as Lifecycle metadata)
m1_to_m3_commit_strategy: bundled                 # single bundled commit covering all 8 paths per Tier S minimal pattern (precedent: IVB-001/SARM-001/TMC-001/TMD-001/SIV-001)
mirror_invariant_tests:
  test_rule_template_mirror_drift_subtests: 10    # all PASS
  test_late_branch_template_mirror_subtests: 4    # all PASS (includes manager-spec.md allowlist entry)
  total_passing: 14                               # 14/14 PASS post-edit
spec_md_lint_status:
  errors: 0                                       # 0 error(s)
  warnings: 1                                     # 1 StatusGitConsistency warning — pre-existing plan-phase baseline (status:draft vs git-implied), resolves automatically on sync-phase per REQ-ARR-003
  attribution: plan-phase-baseline                # NOT a new warning introduced by run-phase
subagent_boundary_grep:
  invocation_pattern_hits: 0                      # no agent AskUserQuestion invocations
  documentary_prose_hits: 1                       # manager-spec.md line 255 — documentary prose describing orchestrator behavior, NOT an invocation. Per C-HRA-008 spirit, constraint applies to tool/Bash invocation, not to markdown prose documentation.
```

## §E.4 Sync-phase Audit-Ready Signal (THIS SPEC IS THE FIRST CANARY)

This SPEC is **forward-looking** in that its own sync-phase is the FIRST application of the new ownership policy. manager-docs running `/moai sync SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001` MUST adhere to the new boundaries declared in REQ-ARR-003 + REQ-ARR-006:

- **Allowed**: CHANGELOG.md `[Unreleased]` entry + 4 SPEC artifact frontmatter `in-progress → implemented` transition + this progress.md `§E.4` block update
- **Forbidden**: any modification to spec.md / plan.md / acceptance.md body content (§A through §H sections)

If sync-phase reveals a need to modify SPEC body content (e.g., last-minute REQ rewording, AC clarification), manager-docs MUST return a structured blocker report; orchestrator re-delegates to manager-spec for the body edit; THEN re-invokes manager-docs to complete sync.

```yaml
sync_complete_at: 2026-05-24T20:45:30Z
sync_commit_sha: 11abb9a30                        # (orchestrator backfilled post-push 2026-05-24T20:47Z)
sync_status: completed                            # CHANGELOG entry + 4 frontmatter status + this §E.4 update done
b12_self_test_a_pre_emission_grep: 0              # PASS: `grep -c 'SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001' CHANGELOG.md` pre-emission
b12_self_test_a_post_emission_grep: 1             # PASS: (post-emission, 1 entry under [Unreleased] ### Changed section)
b12_self_test_b_ac_count_match: 7                 # PASS: acceptance.md SSOT AC count `grep -cE '^\| \*\*AC-ARR-[0-9]+\*\*'` = 7 rows (AC-ARR-001..007)
b12_self_test_c_file_paths_verified: PASS         # PASS: plan.md §A.2 EXTEND entries 7 files (manager-spec/develop/docs sources + mirrors + schema doc) all present + progress.md = 8 total post-sync
changelog_entry_position: [Unreleased] ### Changed section    # PASS: governance/policy SPEC in Changed section (not Fixed), Korean body per git_commit_messages: ko
frontmatter_status_transitions:
  spec.md: implemented                            # PASS: draft → implemented
  plan.md: implemented                            # PASS: draft → implemented
  acceptance.md: implemented                      # PASS: draft → implemented
  progress.md: implemented                        # PASS: draft → implemented (Mx deferred to post-sync)
canary_compliance_check:
  spec_md_body_modified: false                    # PASS: frontmatter status field only, no body changes
  plan_md_body_modified: false                    # PASS: frontmatter status field only, no body changes
  acceptance_md_body_modified: false              # PASS: frontmatter status field only, no body changes
  rationale: |
    This SPEC defines the policy that manager-docs MUST NOT modify spec/plan/acceptance body content.
    Its own sync is the FIRST test of compliance. Canary PASSED: manager-docs touched only CHANGELOG.md
    + 4 frontmatter status transitions + this progress.md §E.4 block. Zero body content modifications.
    Policy is now binding on all downstream SPEC syncs (SIV-001 onwards).
```

### (Migrated from §E.5)

```yaml
mx_complete_at: 2026-05-24T21:25:00Z              # Mx Step C SKIP-judge performed by orchestrator post-sync (separate session from sync-phase, see multi_session_coordination_note)
mx_disposition: SKIP-justified                    # both conditions verified PASS
mx_disposition_rationale: |
  Per .claude/rules/moai/workflow/mx-tag-protocol.md §a, Mx Step C SKIP condition verified:
    (1) .go file count = 0 across ARR-001 5 commits (e2fbe4d60 + e6ad82031 + e48af1792 + 11abb9a30 + a25476e7e) → PASS
        Verification: `for sha in <5 commits>; do git show --name-only $sha | grep '\.go$' || true; done` produced empty output
    (2) @MX tag count delta = 0 across all 7 modified .md files → PASS
        Verification per file: `pre=$(git show e6ad82031~1:"$f" | grep -c "@MX"); post=$(grep -c "@MX" "$f"); delta=$((post - pre))`
  Both conditions PASS → SKIP-justified. Mx-chore commit emits this §E.5/§E.6 finalization only (no @MX scan needed).
mx_tag_count_delta:
  source_agent_files:                              # all delta = 0
    manager-spec.md: 0                             # pre=0, post=0
    manager-develop.md: 0                          # pre=10, post=10 (pre-existing tags untouched by ARR-001 edits)
    manager-docs.md: 0                             # pre=0, post=0
  mirror_agent_files:                              # mirror parity preserved (AC-ARR-005 guarantee)
    manager-spec.md: 0                             # pre=0, post=0
    manager-develop.md: 0                          # pre=10, post=10
    manager-docs.md: 0                             # pre=0, post=0
  schema_doc:
    spec-frontmatter-schema.md: 0                  # pre=0, post=0 (Status Transition Ownership Matrix is pure policy markdown, no code annotations)
mx_step_c_verdict: SKIP                           # ownership declarations are pure policy text, no code annotations introduced
multi_session_coordination_note: |
  ARR-001 4-phase lifecycle was distributed across 3 orchestrator sessions due to absence of multi-session coordination protocol at time of execution:
    - Session A (~earlier): plan + run + run-backfill (commits e2fbe4d60 + e6ad82031 + e48af1792)
    - Session B (f3d5f57e-4620-48a3-9a65-a9fe32b2816c, ≈2026-05-24T20:44Z): sync + sync-backfill (commits 11abb9a30 + a25476e7e)
    - Session C (cd8d8946-e06f-4c76-a3ab-869eba092356, this Mx-chore commit): Mx Step C SKIP-judge finalization
  Race signal: Session C started after Session B's sync commits were pushed to origin/main; Session C's `git fetch origin main` returned `0 0` (clean ahead) so Session B's work was already absorbed silently. Session C nearly delegated `/moai sync SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001` again (duplicate work, would have caused CHANGELOG duplicate + frontmatter overwrite race) before detecting the existence of commits 11abb9a30 + a25476e7e via `git log --all --oneline | grep ARR`.
  Mitigation deferred to SPEC-V3R6-MULTI-SESSION-COORD-001 (new, plan-phase enters after this Mx-chore commit). This race case is cited as motivating example in that SPEC §A (Background) — empirical validation of the 4-layer solution (active-sessions.json registry + paste-ready session_id tagging + SessionStart surface + Pre-spawn HARD rule extension).
```

## §E.6 4-Phase Lifecycle Close Signal

```yaml
lifecycle_close_at: 2026-05-24T21:25:00Z          # final closure timestamp at Mx-chore commit (SKIP-justified path, no separate Mx-EVALUATE step)
final_status: completed                           # 4-phase lifecycle closed: plan + run(M1-M3) + sync + Mx (SKIP-justified)
total_commits: 6                                  # plan e2fbe4d60 + run e6ad82031 + chore-backfill e48af1792 + sync 11abb9a30 + sync-backfill a25476e7e + Mx-chore <this commit>
total_push_count: 6                               # one per commit (Hybrid Trunk 1-person OSS per CLAUDE.local.md §23.7)
sprint_position: Sprint 8 P2 — Audit Tier 2 COMPLETE — Tier S minimal 1-pass cohort 7/7 (IVB/SARM/TMC/TMD/SIV/HCW/ARR confirmed); first canary SPEC of new SPEC Artifact Ownership policy (F1+F12 resolution from Anthropic Best Practices Audit)
canary_compliance_summary:
  policy_self_application: PASS                   # this SPEC's own sync (Session B 11abb9a30) modified ONLY CHANGELOG.md + 4 frontmatter status fields, ZERO spec.md/plan.md/acceptance.md body modifications — policy is binding on all downstream SPEC syncs (SIV-001 sync 8b75ebbb3 already complied retroactively)
  forward_looking_test_planned: SPEC-V3R6-MULTI-SESSION-COORD-001 will be the second canary (manager-spec authors body; manager-develop edits Go/hook; manager-docs sync; orchestrator Mx — clean DRI separation across 4 phases)
next_action_paste_ready: |
  ultrathink. Sprint 8 P3 — SIV-001 Mx Step C SKIP-judge + SPEC-V3R6-MULTI-SESSION-COORD-001 plan-phase enter.
  applied lessons: project_sprint8_arr001_complete (NEW), L52 multi-session race absorbed (Session B sync→Session C Mx coordination gap → motivating COORD-001), L33 7th Tier S minimal cohort, L44 HARD pre-spawn fetch discipline.

  전제 검증:
  1) git log --oneline -1 → <ARR-001-Mx-chore-SHA> (ARR-001 4-phase lifecycle complete)
  2) git fetch origin main && git rev-list --count --left-right origin/main...HEAD → 0 0 (L44 HARD pre-spawn race recheck; verify no other-session activity since this Mx-chore)
  3) grep "mx_status: TBD" .moai/specs/SPEC-V3R6-SPEC-ID-VALIDATION-001/progress.md → 1 line (SIV-001 Mx 잔여 미수행 확인)
  4) grep -c "@MX" internal/template/rule_template_mirror_test.go → expected 0 (SIV-001 run-phase .go edit was structural-only slice add, no @MX tag)

  실행: SIV-001 Mx Step C SKIP-judge — progress.md §F mx_status TBD → SKIP + chore commit `chore(SPEC-V3R6-SPEC-ID-VALIDATION-001): Mx-phase audit-ready signal + 4-phase close (SKIP-justified)`.

  머지 후: SPEC-V3R6-MULTI-SESSION-COORD-001 plan-phase 진입. manager-spec 위임. Tier M 예상 (~300 LOC Go + 2 hook script + 1 rule). 본 ARR-001 progress.md §E.5 multi_session_coordination_note를 motivating example로 §A.1 Background에 인용. 4-layer 솔루션: (L1) .moai/state/active-sessions.json registry + SessionStart/Stop/PreCompact hook 통합 / (L2) paste-ready emit 시 source_session_id 태깅 + MEMORY.md entry session_id 마킹 / (L3) SessionStart hook이 active-sessions 조회 → 진행 중 작업 surface / (L4) Pre-spawn HARD rule 확장 (agent-common-protocol.md §Pre-Spawn Sync Check에 active-sessions.json 조회 batch 추가).
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
