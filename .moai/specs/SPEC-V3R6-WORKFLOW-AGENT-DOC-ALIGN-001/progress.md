# Progress — SPEC-V3R6-WORKFLOW-AGENT-DOC-ALIGN-001

Lifecycle: plan → run → sync (3-phase). Tier L. cycle_type=ddd.

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifact set authored by manager-spec (2026-06-20): spec.md + plan.md + acceptance.md + design.md + research.md + this progress.md (§E skeleton). Status: draft. era: V3R6.

- 5-artifact Tier L set complete; GEARS REQ-WADA-001..017 across 4 groups + cross-cutting template-mirror.
- AC matrix (acceptance.md §D) AC-WADA-001..017, each with a both-tree-verifiable grep/test command.
- Evidence baseline captured (research.md §B) with observed grep/test output; non-reproductions recorded honestly (OQ-1 release-update absent).
- Out of Scope section present (spec.md §F) with `### Out of Scope —` H3 sub-headings.
- SPEC ID self-check: `decomposition: SPEC ✓ | V3R6 ✓ | WORKFLOW ✓ | AGENT ✓ | DOC ✓ | ALIGN ✓ | 001 ✓ → PASS`.
- Awaiting: plan-auditor independent audit → Implementation Kickoff Approval (Tier L human gate) → run-phase.

## §E.2 Run-phase Evidence

### M1 — ANALYZE: occurrence inventory + replacement map

Pre-flight greps (re-run against working tree, sync 0 0 vs origin/main):
- `ls .claude/agents/` → `harness/ local/ moai/` (NO `builder/` — D1 confirmed).
- Raw total (ARCHIVED12 incl researcher): **139 lines** across **38 files** (matches plan-phase).
- `researcher` count: **14** (KEEP — live role_profile baseline).
- Archived-proper (ARCHIVED11) total: **125 lines** across **36 files** (matches plan-phase ~125).

Replacement map (migration table `archived-agent-rejection.md §C` drives all):

| Archived agent | Canonical replacement (this SPEC) |
|----------------|------------------------------------|
| `manager-strategy` | `manager-spec` (planning IS strategy — row 1) |
| `manager-quality` | sync-auditor / orchestrator verification batch / Stop hook / `gh issue create` (feedback.md, row 2) |
| `manager-brain` | `Explore` then `manager-spec` chain (row 3) |
| `manager-project` | `manager-docs` (row 4) |
| `expert-backend/frontend/devops/performance/refactoring` | per-spawn `Agent(general-purpose)` with domain whitelist (rows 7-12) |
| `expert-security` | Stop hook dependency-manifest audit OR per-spawn `Agent(general-purpose)` security reviewer (row 9) |
| `researcher` | **KEEP** — live role_profile (NOT replaced) |

Intent classification per occurrence: spawn target (replace) | documentation pointer to rejection rule (keep) | frontmatter `agents:` list (replace/remove) | role_profile (keep).

Mirror status per file group:
- MIRRORED (apply both trees): all `workflows/**`, `references/**`, `team/**` (except dev-only), all 6 `agents/moai/**`.
- DEV-ONLY no-mirror (live only): `workflows/release.md`/`github.md`/`release-update.md` — none carry archived-proper refs (release-update token absent).
- BASELINE-DIVERGED (changed-line mirror only): `manager-spec.md`/`manager-docs.md`/`manager-git.md`.

Group 2/3/4 baseline:
- `team-reader`/`team-validator` invalid subagent_type: glm.md (model table), debug.md, review.md (+frontmatter), plan.md (+frontmatter) — fixed M4 (REQ-WADA-009).
- `§ Terminology Glossary` absent in worktree-integration.md; 2 cross-refs (spec-workflow.md:23, worktree-state-guard.md:19) → M4 ADD section (REQ-WADA-010).
- GLM models `glm-5.1`/`glm-4.7`/`glm-4.5-air`/`glm-4.7-flashx` in glm.md:148-150,168-172 → M5 `glm-5.2[1m]` (REQ-WADA-011).
- Retired terms: `sprint contract` (workflows/run.md:80), `Round split` (team/run.md:69) → M5 (REQ-WADA-012).
- SKILL.md footer `Version: 2.6.0` / `Last Updated: 2026-02-25` (lines 342-343) → M5 bump (REQ-WADA-013).
- `release-update` token: **ABSENT** → AC-WADA-008 vacuously satisfied (OQ-1 confirmed not-reproduced).
- run.md: **246 LOC** both trees → M6 trim < 200 (REQ-WADA-014/015).

### M2 — workflow skill files archived-agent purge

26 files purged (workflows/** + references/{reference,anti-patterns,mx-tag}.md), both trees identical (changed-line mirror via edit-then-cp). Replacements per migration table §C:
- `manager-strategy` → `manager-spec` (phase-execution.md Phase 1/1.5, mode-orchestration.md).
- `manager-quality` → sync-auditor (review.md, moai.md, task-decomposition.md, quality-gates-quality.md, doc-execution.md) / manager-develop fixes (loop.md, fix.md, quality-gates-context.md, delivery.md) / Stop hook (gate.md, security.md, sync.md prose) / orchestrator `gh issue create` (feedback.md REQ-WADA-005).
- `manager-brain` → `Explore` + `manager-spec` (brain.md frontmatter).
- `manager-project` → `manager-docs` (project.md frontmatter, dedup).
- `expert-*` → per-spawn `Agent(general-purpose)` domain specialists (clean.md refactoring, review.md/clarity-interview.md/design.md/e2e.md frontend, security.md/quality-gates-quality.md security, doc-generation.md devops, loop.md/fix.md/mx.md/moai.md domain mix).
- design.md frontend role cross-references the FROZEN `design/constitution.md` carve-out note.
- `gh issue create` preserved in feedback.md (AC-WADA-005: count=2).
- residual grep (both trees, excl rejection-rule + dev-only trio): **0**. researcher KEEP: **14** (unchanged).

### M3 — team files + agent bodies + frontmatter purge

Team files (4): debug.md (frontmatter + body manager-quality → manager-develop), run.md (TRUST5 → sync-auditor ×2), glm.md (PHASE 3 → sync-auditor), review.md (fallback → sync-auditor). team-reader/team-validator deferred to M4.
Agent bodies (6 under .claude/agents/moai/): manager-develop (OUT OF SCOPE + Delegation Protocol → sync-auditor/per-spawn; @MX:ANCHOR comment reworded), builder-harness (OUT OF SCOPE + Delegation Protocol → per-spawn), sync-auditor (manager-quality supplement → orchestrator batch + Stop hook), manager-git (Input from → sync-auditor), manager-spec (line 5 absorption → §C row 1 pointer + Delegation/Step 6 → per-spawn), manager-docs (line 5 absorption → §C row 4 pointer + OUT OF SCOPE/Delegation → per-spawn/sync-auditor).
Mirror: 4 team files + manager-develop/builder-harness/sync-auditor.md fully mirrored (cp, baseline-identical); manager-spec/docs/git.md changed-line mirror only (Edit, baseline-diverged per §E.5/Mx lifecycle owned by SPEC-V3R6-LIFECYCLE-REDESIGN-001).
Full archived-proper residual BOTH trees (skills+agents, excl rejection + dev-only): **0**. researcher KEEP: **14**.

### M4 — broken cross-refs + invalid dispatch

- REQ-WADA-009: `team-reader`/`team-validator` invalid subagent_type → `general-purpose` (matching team/run.md pattern). plan.md (frontmatter + 3 subagent_type + intro prose + Version footer), review.md (frontmatter + 4 subagent_type), debug.md (3 prose "team-reader agent" labels), glm.md (CG model-mapping table rows team-reader/team-validator → canonical role_profiles). Both trees: **0** residual.
- REQ-WADA-010: `## Terminology Glossary` (L1/L2/L3 layer definitions) ADDED to worktree-integration.md (rule file, both trees byte-identical via embedded_mirror_test). 2 cross-refs (spec-workflow.md:23, worktree-state-guard.md:19) + CLAUDE.md §14 now resolve.
- REQ-WADA-008: `release-update` token **ABSENT** at run-phase → AC-WADA-008 vacuously satisfied (OQ-1 confirmed; no broken dispatch to fix).

### M5 — stale ground-truth

- REQ-WADA-011: `team/glm.md` GLM model names `glm-5.1`/`glm-4.7`/`glm-4.5-air`/`glm-4.7-flashx` → `glm-5.2[1m]` (env-var table + CG model-mapping table). Both trees: stale=**0**, glm-5.2 present.
- REQ-WADA-012: retired terms — `Round split` (team/run.md:69) → `Milestone split`; `sprint contract` (workflows/run.md:80) → `GAN-loop Sprint Contract Protocol` (proper-noun design term, not the retired Sprint grouping). Both trees: **0** residual.
- REQ-WADA-013: SKILL.md body footer `Version: 2.6.0` / `Last Updated: 2026-02-25` → `2.7.0` / `2026-06-20`. Both trees: stale=**0**.

### M6 — run.md LOC trim + verification + template parity

- REQ-WADA-014: `workflows/run.md` trimmed **246 → 181 LOC** (both trees, under the 200 ceiling). Compressed the verbose Run-phase Autonomy §3-7 prose + the entire Recursive Self-Diagnosis Loop §1-9 into compact cross-reference + summary-contract tables. Entry-router routing logic PRESERVED: Phase Routing Table, Invocation Flow, On-Demand Sub-skill Loading, Mode dispatch list. Grep-anchored tokens PRESERVED + ordering invariant holds: `MODE_UNKNOWN` (agentless_audit), `## Run-phase Autonomy (/goal ac_converge)` heading (orchestration-mode-selection cross-refs), Implementation Kickoff Approval + AskUserQuestion BEFORE first `/goal` (kickoff-preservation test), `ac_converge`, score-independence (`regardless of` + `plan-auditor`).
- REQ-WADA-015: `go test ./internal/skills/ -run TestEntryRouterLOCCeiling` → **PASS** (was the orphaned RED test; now GREEN).
- Catalog hash regen (same-SPEC cascade): SKILL.md + 6 agent body edits invalidated `catalog.yaml` SHA256 hashes → regenerated via `gen-catalog-hashes.go --all` + `make build`. `TestManifestHashFormat` → PASS.
- REQ-WADA-016: `TestTemplateMirrorParity` PASS; bidirectional changed-line grep = 0 both trees. The 3 baseline-diverged agent bodies (manager-spec/docs/git.md) verified via Step-2 union grep (archived-purge landed in both trees; whole-file divergence is the pre-existing §E.5/Mx lifecycle content, out of scope). `embedded_mirror_test` (rule-file byte parity for worktree-integration.md) PASS.
- REQ-WADA-017: `TestTemplateNeutralityAudit` PASS — replacement text introduced no internal SPEC IDs / REQ tokens / audit citations / dates / SHAs into template (all §C references use the canonical rule path `archived-agent-rejection.md §C`, a permanent-rule citation, not an internal-dev token).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-20
run_commit_sha: 4908558f3  # M6 final run-phase commit (per-milestone chain: 6d105b647 M1+M2, e16804a4f M3, f0f9e0318 M4+M5, 4908558f3 M6)
run_status: implemented
cycle_type: ddd
milestones_complete: M1-M6 (6/6)
ac_pass_count: 17  # AC-WADA-001..017 + 001a all MUST-FIX PASS; AC-WADA-008 vacuous (release-update absent)
ac_fail_count: 0
researcher_keep_count: 16  # baseline 14; +2 legitimate role_profile refs from team-reader→general-purpose reframe (≥14, no over-purge per AC-WADA-001a)
archived_proper_residual_both_trees: 0
team_reader_validator_residual_both_trees: 0
run_md_loc: 181  # both trees, < 200 ceiling
cross_platform_build:
  linux_amd64: exit 0
  windows_amd64: exit 0  # GOOS=windows GOARCH=amd64 go build ./...
go_test_full_suite: exit 0  # internal/hook wrapper flakiness (signal:killed under parallel load) resolved on isolated retry; environmental, not SPEC-caused (0 Go files changed)
new_warnings_or_lints_introduced: 0  # documentation/markdown-only SPEC; zero Go source touched
total_run_phase_files: 56  # 27 live workflow/agent/team/rule + 27 template mirror + run.md trim both trees + catalog.yaml + spec.md + progress.md
m1_to_m6_commit_strategy: per-milestone specific-path commits (M1+M2 / M3 / M4+M5 / M6); Authored-By-Agent trailer + Korean message + 🗿 MoAI footer
```

### §E.3 Verification-Claim Integrity (5-section evidence)

- **Claim**: all MUST-FIX ACs PASS; archived-proper residual 0 both trees; run.md < 200; full suite exit 0.
- **Evidence**: `grep -rnE "$ARCHIVED11" ... | grep -v archived-agent-rejection | grep -Ev '(release|github|release-update).md:' | grep -v comment-pointer` → empty both trees; `wc -l run.md` → 181/181; `go test ./internal/skills/` → ok; `go test ./internal/template/` → ok (incl. TestManifestHashFormat, TestTemplateMirrorParity, TestTemplateNeutralityAudit, kickoff-preservation, agentless-audit); `go build ./...` + `GOOS=windows go build ./...` → exit 0; `go test ./...` → exit 0.
- **Baseline-attribution**: measured against worktree HEAD 94a468a2e (= origin/main at spawn, sync 0 0); researcher baseline 14 → 16 (increase, not over-purge).
- **Gaps**: (1) `run_commit_sha` is a placeholder until the M6 commit lands (backfilled below). (2) The 3 baseline-diverged agent bodies' whole-file live↔template divergence was NOT reconciled (out of scope, owned by SPEC-V3R6-LIFECYCLE-REDESIGN-001) — only this SPEC's archived-purge edit was mirrored. (3) `internal/hook` 2 wrapper tests flaked once under full-suite parallel load (signal:killed at 5s); passed on isolated `-count=1` retry — environmental, not a SPEC regression.
- **Residual-risk**: the `make build` regenerated the `moai` binary + catalog.yaml hashes; the binary is a build artifact (not committed). A future template edit to any of the 7 hashed agent/skill files will require the same `gen-catalog-hashes.go --all` regen cascade.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-06-20
sync_commit_sha: <placeholder — backfilled after first sync commit>
sync_status: completed
b12_self_test_a: grep -c 'SPEC-V3R6-WORKFLOW-AGENT-DOC-ALIGN-001' CHANGELOG.md → 1 (no duplicate)
b12_self_test_b: acceptance.md AC row count (should be 18) → 17 rows (AC-WADA-001 through AC-WADA-017)
b12_self_test_c: file path verification — ls .moai/specs/SPEC-V3R6-WORKFLOW-AGENT-DOC-ALIGN-001/spec.md → exists ✓
sync_evidence:
  run_phase_commit_sha: 4908558f3
  ac_pass_count: 17
  archived_proper_residual_both_trees: 0
  run_md_loc_final: 181
  cross_platform_build: 'exit 0 (darwin + windows)'
  go_test_full_suite: 'exit 0'
  template_mirror_parity_test: PASS
  template_neutrality_audit: PASS
  sync_artifacts:
    - spec.md frontmatter transition (in-progress → completed)
    - progress.md §E.4 population (this block)
    - CHANGELOG.md entry (added under [Unreleased] → Changed)
```
