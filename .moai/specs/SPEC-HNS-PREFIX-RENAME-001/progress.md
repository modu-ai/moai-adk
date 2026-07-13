---
id: SPEC-HNS-PREFIX-RENAME-001
updated: 2026-07-13
document: progress
plan_status: audit-ready
---

# Progress — SPEC-HNS-PREFIX-RENAME-001

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored 2026-07-13 by manager-spec: spec.md (26 REQs, REQ-HPR-001..026), plan.md (M1–M4 milestones, zero NEEDS CLARIFICATION markers as of iteration 2), acceptance.md (16 ACs + 3 GWT scenarios + 6 edge cases).
- SPEC ID self-check: `decomposition: SPEC ✓ | HNS ✓ | PREFIX ✓ | RENAME ✓ | 001 ✓ → PASS` (executed regex evidence: `PASS`, re-run at iteration 2).
- Scope baseline measured 2026-07-13 (see plan.md §A.1); run-phase must re-verify anchors.
- **Iteration 2 revision (2026-07-13; plan-auditor iter-1 FAIL 0.86 → defects D1–D5 addressed; artifacts at v0.1.1)**:
  - **D1 (BLOCKING)**: artifact prefix fixed to **lowercase `hns-`** per final user decision (lowercase-kebab matches the Claude Code skill/agent naming convention; uppercase runtime-acceptance risk eliminated). NEEDS CLARIFICATION marker, pre-M3 probe gate, and fallback branch all removed — zero markers remain. SPEC ID and REQ-HPR IDs unchanged.
  - **D2**: plan.md §B.2 collision analysis re-grounded with live evidence — legacy `REQ-HNS-*`/`AC-HNS-*` comment tokens found in 5 of 8 production Go files (16 occurrences: update.go 10, doctor_harness.go 2, prefix_conflict.go 2, doctor_skills.go 1, frozen_guard.go 1); lowercase `hns-` = **0 case-sensitive matches** in the production/artifact tree (live grep 2026-07-13); [HARD] all run-phase sweeps case-sensitive (no `grep -i`).
  - **D3**: REQ-HPR-012 corrected — doctor Runner resolution is manifest `runner_workflow`-driven (prefix-agnostic path join, ground-truth read of `internal/cli/harness/doctor.go`); only `runnerSpecialistRE` needs the dual pattern `(harness|hns)-[a-z0-9-]+-specialist`.
  - **D4**: AC-HPR-012 non-target baseline-delta file set pinned to 10 named files (4 handle-harness-observe hook templates, settings.json.tmpl, 4 config section files, catalog_loader.go).
  - **D5**: spec.md §E group-B row gains AC-HPR-007 as REQ-HPR-009's binding AC.
  - Consistency note: lowercase `hns-` also satisfies the doctor NAME-prefix resolution (doctor_harness.go `skills:` reference resolution, plan-time ~L276) without a third pattern (plan.md §B.1). M2-before-M3 ordering invariant unchanged. plan_status remains audit-ready.

## §E.2 Run-phase Evidence

Run-phase executed 2026-07-13 by manager-develop (cycle_type=tdd) in L1 worktree `worktree-agent-a45698e9be95b3450` (base fast-forwarded to 86cdbd97d). Commits: M1 `16fc75bc6`, M2 `a9d2879b0`, M3 `f55aefef3`, M4 (this commit). Verbatim evidence files: `.moai/state/verify/hns-rename/` (gitignored runtime state).

### AC PASS/FAIL matrix (16/16 PASS)

| AC | Status | Verification Command | Actual Output (tail) |
|----|--------|---------------------|----------------------|
| AC-HPR-001 | PASS | `grep -c 'harness-<name>'` + `grep -c 'hns-<name>'` on 4 template contract docs | `harness-<name>`=0 in all 4; `hns-<name>`: SKILL.md 1 / harness-builder.md 12 (all 6 artifact-type contracts) / harness-build-entry.md 2 / moai-meta-harness 1 |
| AC-HPR-002 | PASS | `diff -q` 4 template↔live pairs (+4 secondary pairs) | 8/8 `PARITY OK` |
| AC-HPR-003 | PASS | `go test ./internal/cli/ -run 'TestUserOwnedNamespace_HNS\|TestUserAreaPath_HNS'` | `ok internal/cli` — hns- user-owned; `hnsx-foo`/`hnsfoo`/`HNS-x` NOT matched (exact byte-exact HasPrefix); moai-harness-learner NOT user-owned |
| AC-HPR-004 | PASS | `go test ./internal/cli/ -run TestUpdateNamespaceHNS_TriGenerationPreservation` | E2E t.TempDir sandbox: 7 artifacts across 3 generations survive `cleanMoaiManagedPaths` byte-identical (SHA-256) |
| AC-HPR-005 | PASS | full suite incl. pre-existing template-overwrite tests, assertions unmodified | `go test -count=1 ./...` exit 0 (evidence: m4-full-suite.log) |
| AC-HPR-006 | PASS | `go test ./internal/cli/harness/ -run 'TestListHarnesses\|TestEditHarness\|TestRemoveHarness\|TestHarnessArtifactBelongsTo' -v` | 14 PASS incl. TestListHarnesses_HNSManifest, TestRemoveHarness_MixedGeneration (hns- Runner + legacy specialist removed atomically, bystanders untouched), hns-release-update shadowing case |
| AC-HPR-007 | PASS | `go test ./internal/cli/harness/ -run TestDoctor -v` + `go test ./internal/harness/ -run 'PrefixConflict\|Frozen' -v` | TestRunnerSpecialistRE_HNSDualPattern PASS (both generations matched); TestDoctor_HNSHarness_Passes PASS **pre-change** (proves manifest-driven prefix-agnostic Runner resolution — REQ-HPR-012 ground truth); TestDoctor_HNSDanglingSpecialist_Detected PASS; TestFrozenGuard_HNSAllowed + TestDetectPrefixConflicts_HNSPrefix PASS |
| AC-HPR-008 | PASS | `./bin/moai harness doctor; echo exit=$?` + `go run ./cmd/moai doctor` | doctor: `Scanned 3 harness(es): 0 ERROR, 2 INFO` exit=0 (hns- Runner + specialist resolve); `moai doctor` emits no hns- complaint (classifySkill hns-→INFO pinned by TestClassifySkill_HNS); Harness 5-Layer L1-L6 FAIL verified byte-identical to pre-M3 baseline a9d2879b0 via git-archive checkout — pre-existing dev-repo state, no delta |
| AC-HPR-009 | PASS | `go test ./internal/template/ -run TestSplitHarnessNamespaceNoLeak -v` | PASS; `splitHarnessAgentPrefixes` carries BOTH `harness-{release-update,github,release}` AND `hns-{release-update,github,release}` |
| AC-HPR-010 | PASS | red-team: plant `templates/.claude/skills/hns-probe/SKILL.md` → test → revert → test | Planted: `FAIL ... Leaked entries: [.claude/skills/hns-probe]`; reverted: `ok` (evidence: m4-redteam-planted.txt / m4-redteam-reverted.txt) |
| AC-HPR-011 | PASS | scoped stale-ref sweep (acceptance formula, case-sensitive) | **0 matches** after M4 (m4-stale-ref-final.txt); post-M3 interim 10 matches were all `.moai/docs/dev-only-commands-isolation.md` (resolved in M4) |
| AC-HPR-012 | PASS | `grep -rc` over the 10 pinned non-target files, before M1 vs after M4 | sorted per-file counts identical (ac012-baseline-before-m1.txt vs ac012-after-m4.txt); 29-file classification table below |
| AC-HPR-013 | PASS | Read doctrine docs | harness-namespace-doctrine.md: §24.1 hns- canonical row + legacy row, §24.4 hns- update-contract rows, new §24.6 rename record (Builder emits hns- only, tri-generation recognition); dev-only-commands-isolation.md: artifact tables + checklists → hns- names + dual-name CI guard note |
| AC-HPR-014 | PASS | `make build && go test -count=1 ./... ; echo exit=$?` + neutrality | make build exit=0; full suite exit=0 (99 pkgs ok); TestLanguageNeutrality + TestInternalContentLeak PASS; `grep -r 'SPEC-HNS-PREFIX-RENAME' internal/template/templates/` = 0 |
| AC-HPR-015 | PASS | `git log --stat 86cdbd97d..HEAD -- CLAUDE.local.md` | CLAUDE.local.md absent from every SPEC commit; §21/§24 pointer flag list delivered in the completion report |
| AC-HPR-016 | PASS | `go test ./internal/cli/... ./internal/harness/... ./internal/template/... -count=1` + `git log --diff-filter=D --stat -- '*_test.go'` | all legacy-prefix tests PASS with assertions unmodified; 0 test files deleted; Runner JS L56/L91 show `hns-release-update-specialist` |

### REQ-HPR-004 — 29-file template classification table

Rename-target, edited (9): `skills/moai/SKILL.md` (contract line + namespace dual-pattern statement), `skills/moai/workflows/harness-builder.md` (GENERATE contract, 12 tokens), `skills/moai/workflows/harness-build-entry.md` (2 tokens + agent path), `skills/moai-meta-harness/SKILL.md` (Runner contract), `skills/moai/workflows/harness.md` (Tier-4 apply-target prose → dual-pattern), `skills/moai-meta-harness/references/seven-phase-workflow.md` (generation contract → hns-), `skills/moai/workflows/project/meta-harness.md` (L1 trigger table + smoke-gate dangling-ref prose), `skills/moai-harness-learner/SKILL.md` (auto-update targets → tri-generation), `rules/moai/development/skill-authoring.md` (Skills Namespace Policy + Deprecated Skill Slots concrete names).

Non-target, preserved verbatim (20): `agents/moai/builder-harness.md` (run-time verification: only `harness-generated` tier label — plan's "likely rename-target" overridden by observation), `hooks/moai/handle-harness-observe{,-stop,-subagent-stop,-user-prompt-submit}.sh.tmpl` ×4 (hook names), `settings.json.tmpl` (hook entries), `.moai/config/sections/{harness,interview,workflow}.yaml` + `system.yaml.tmpl` ×4 (config keys/state paths), `rules/moai/core/agent-common-protocol.md` + `rules/moai/core/hooks-system.md` (hook names), `rules/moai/development/coding-standards.md` + `rules/moai/workflow/runtime-recovery-doctrine.md` (generic prose), `workflows/project/{codebase-analysis,doc-generation,mode-detection}.md` ×3 + `workflows/project.md` + `workflows/run/context-loading.md` (state paths/prose), `skills/moai-meta-harness/references/agent-cross-references.md` (0 artifact tokens — verified).

### Coverage non-regression (touched packages, measured both sides)

| Package | Pre-SPEC baseline | Post-M4 | Delta |
|---------|-------------------|---------|-------|
| internal/cli | 72.8% | 72.8% | ±0 (pre-existing sub-85% baseline; non-regression satisfied per acceptance §F) |
| internal/cli/harness | 74.4% | 74.8% | +0.4 |
| internal/harness | 87.3% | 87.3% | ±0 |
| internal/template | 85.3% | 85.3% | ±0 |

### Cascade follow-ups (L46, within scope envelope)

- `internal/template/catalog.yaml` SHA256 regen via `gen-catalog-hashes.go --all` (A3c pattern — M1 SKILL.md body edits invalidated 3 stored hashes; regen in M2 commit + re-run by `make build` in M4).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-13
run_commit_sha: 7873c8f8f552c0c1ac6932d7640a9dbeba196142  # M4 commit (backfilled per D3 exemption)
run_status: audit-ready
ac_pass_count: 16
ac_fail_count: 0
preserve_list_post_run_count: 0  # zero deletions/modifications of user-owned-prefix artifacts by any test or build step
l44_pre_commit_fetch: n/a  # L1 worktree branch; integration push owned by orchestrator (worktree cannot push branch to main directly)
l44_post_push_fetch: n/a
new_warnings_or_lints_introduced: 0  # golangci-lint 0 issues; go vet clean
cross_platform_build:
  darwin: exit 0 (go build ./...)
  windows: exit 0 (GOOS=windows GOARCH=amd64 go build ./...)
total_run_phase_files: 62  # git diff --stat 86cdbd97d..HEAD (M1 19 + M2 16 + M3 17 + M4 doctrine/progress/catalog)
m1_to_mN_commit_strategy: per-milestone commits M1..M4 on L1 worktree branch; orchestrator integrates to main (cherry-pick/merge per stale-worktree lesson)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-13
sync_commit_sha: pending-backfill-hns-prefix-rename-001  # this sync commit cannot reference its own SHA; backfilled in a follow-up chore commit per the SHA placeholder backfill exemption (D3)
sync_status: audit-ready
changelog_entry_position: top-of-Unreleased  # CHANGELOG.md [Unreleased] > Changed
frontmatter_status_transitions:
  spec_md: "in-progress -> completed"
  plan_md: "no change (frontmatter not tracked separately)"
  acceptance_md: "no change (frontmatter not tracked separately)"
  progress_md: "this file — §E.4 populated"
b12_self_test_a: "grep -c 'SPEC-HNS-PREFIX-RENAME-001' CHANGELOG.md (pre-emission) = 0 -> emission proceeded"
b12_self_test_b: "acceptance.md AC-HPR-* row count = 16; CHANGELOG entry states '16/16 AC PASS' — match"
b12_self_test_c: "file paths cited in CHANGELOG entry verified via ls (skills/moai/workflows/harness-builder.md, harness-build-entry.md, moai/SKILL.md, moai-meta-harness/SKILL.md all exist under internal/template/templates/.claude/)"
```

MX Tag validation (sync sub-step, not a separate phase): no new exported functions requiring `@MX:ANCHOR`/`@MX:WARN` were introduced by this rename-only SPEC (prefix string constants + regex pattern updates only); existing MX annotations in touched files (`update.go`, `doctor_harness.go`, `prefix_conflict.go`, `frozen_guard.go`) preserved verbatim.

## §F Phase 0.95 Mode Selection

- Inputs: tier=M, scope≈45 files (templates 29 + Go 8 + tests + local 7), domains=3 (template docs / Go code / local artifacts), language mix=md+go, concurrency benefit=LOW (coding-heavy, M2-before-M3 hard ordering)
- Mode evaluation: trivial NO (multi-file semantic) / background NO (write) / agent-team RETIRED / parallel NO (coding-heavy) / workflow NO (multi-rule, inter-file dependency, <30 uniform) / sub-agent YES
- Decision: sub-agent
- Justification: M1→M4 milestones are sequentially dependent (Go dual-pattern recognition must land before local rename); coding-heavy work defaults to sequential per Anthropic coding-task parallelism caveat.
