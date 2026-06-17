# progress.md — SPEC-V3R6-WORKFLOW-EFFORT-MAP-001

> Lifecycle progress tracker. §E.1 is populated at plan-phase; §E.2-§E.5 are populated by manager-develop (run), manager-docs (sync), and orchestrator/manager-docs (Mx) respectively per the Status Transition Ownership Matrix.

## §E.1 Plan-phase Audit-Ready Signal

- **SPEC ID:** SPEC-V3R6-WORKFLOW-EFFORT-MAP-001
- **Tier:** M
- **Era:** V3R6 (target; will be auto-classified H-4 once §E.2 + §E.5 carry real commit SHAs)
- **Plan-phase artifacts:** spec.md (created 2026-06-17), plan.md (created 2026-06-17), acceptance.md (created 2026-06-17), progress.md (this file)
- **Plan-phase commit:** _<pending — to be set when the plan-phase commit lands on main>_
- **Pre-verified ground truth:**
  - dynamic-workflows.md / session-handoff.md / output-styles/moai/moai.md mirror parity = 0/0/0 (baseline 2026-06-17)
  - dynamic-workflows.md L74 already states ultracode-not-restored doctrine
  - codemaps-extract.js L53 omits `effort` (inherits session)
  - workflow.yaml `role_profiles` (7 roles) carries no `effort` field
  - `internal/template/model_policy.go` `agentModelMap` (5 agents) + `agentEffortMap` (5 agents) — OUT OF SCOPE
  - `internal/harness/router/effort.go` minimal/standard/thorough → medium/high/xhigh — OUT OF SCOPE
- **Audit-ready self-check:** SPEC ID decomposition PASS (`SPEC ✓ | V3R6 ✓ | WORKFLOW ✓ | EFFORT ✓ | MAP ✓ | 001 ✓ → PASS`); frontmatter 12-canonical-field schema PASS; EARS/GEARS compliance PASS (13 requirements, all GEARS-pattern); exclusions section present (§H, 7 entries); no implementation details in spec.md (WHAT/WHY only).

## §E.2 Run-phase Evidence

**Run-phase exec:** manager-develop, cycle_type=ddd (brownfield doctrine editing), L1 worktree `agent-a82c6d9e6f3a73c3b` on branch `worktree-agent-a82c6d9e6f3a73c3b`. Implementation Kickoff Approval granted (human gate passed); plan-auditor verdict PASS-WITH-DEBT 0.84 (Tier M threshold 0.80); 3 NEW MINOR defects D1-new/D2-new/D3-new ACCEPTED as residual debt (not fixed, per delegation contract).

**Milestone completion:**
- **M1** — SSOT taxonomy section "Purpose-driven model+effort selection" added to `dynamic-workflows.md` + template mirror (same commit). 7-row taxonomy table + agent() opt schema + REQ-WEM-002 explicit-purpose rule + codemaps-extract.js worked example. Byte-parity 0-diff.
- **M2** — `codemaps-extract.js` line 54 `agent()` call now carries `effort: 'low'` co-located with `agentType: 'Explore'` on a single line, plus a 1-line comment documenting the read-only-extract → low mapping.
- **M3** — `effort` field added to all 7 `role_profiles` in both `workflow.yaml` (local + template mirror). Values: researcher=low, analyst=medium, architect=xhigh, designer=medium, implementer=xhigh, tester=medium, reviewer=high. Template comment omits SPEC ID (neutrality); local comment carries SPEC-ID + REQ-WEM-006. 7 effort values byte-aligned across both files.
- **M4** — `session-handoff.md` Block 1 updated in 3 places (Canonical Format skeleton conditional ultracode comment + Field-by-Field Spec sub-bullet + Anti-Patterns general-hygiene entry) + template mirror. Byte-parity 0-diff.
- **M5** — `output-styles/moai/moai.md` §8 render surface updated to match session-handoff Block 1 edit (Canonical skeleton + Pre-emit self-check item + Anti-patterns entry) + template mirror. All 3 mirror pairs verified 0-diff post-edit.

**AC PASS/FAIL matrix (13/13 MUST PASS):**

| AC | Status | Evidence command | Observed output |
|----|--------|------------------|-----------------|
| AC-WEM-001 | PASS | `grep -cn "Purpose-driven model+effort selection..." dynamic-workflows.md` | 2 |
| AC-WEM-002 | PASS | `grep -cn "SHALL set .effort. explicitly..." dynamic-workflows.md` | 2 |
| AC-WEM-003 | PASS | `grep -nE "agentType: 'Explore'.*effort: 'low'\|effort: 'low'.*agentType: 'Explore'" codemaps-extract.js` | line 54 match (single-line co-located) |
| AC-WEM-004 | PASS | local role keys / effort keys | 7 / 7 |
| AC-WEM-005 | PASS | template effort count / value diff | 7 / 0 diff lines |
| AC-WEM-006 | PASS | `git diff --name-only origin/main -- internal/ \| grep '\.go$'` + struct-field-absence | 0 Go files / 0 Effort fields in RoleProfileEntry |
| AC-WEM-007 | PASS | conditional structure / literal /effort ultracode / doctrine ref | 3 / 3 / 2 |
| AC-WEM-008 | PASS | moai.md §8 conditional structure / literal /effort ultracode | 3 / 3 |
| AC-WEM-009a | PASS | `diff session-handoff.md <template>` | 0 |
| AC-WEM-009b | PASS | `diff dynamic-workflows.md <template>` + `diff moai.md <template>` | 0 / 0 |
| AC-WEM-010 | PASS | section heading / official citation / codemaps ref | 1 / 1 / 1 |
| AC-WEM-011 | PASS | SPEC ID / REQ token / NEW feedback_ leakage | 0 / 0 / 0 NEW (2 pre-existing grandfathered refs confirmed not added by this SPEC via `git diff \| grep '^\+.*feedback_'` = 0) |
| AC-WEM-012 | PASS | files modified vs origin/main = 9 run-phase targets exactly (3 doctrine SSOT + 3 template mirrors + 2 config + 1 workflow script) + 4 SPEC artifacts; 0 dirty-tree YAMLs / 0 untracked design-docs-reports / 0 sibling SPEC dirs absorbed | 9 + 4 = 13 paths, matches plan.md §C enumeration |

**Decision #1 invariant (no Go source under internal/ modified):** PRIMARY `git diff --name-only origin/main -- internal/ | grep '\.go$'` returns 0 lines. SUPPORTING `grep -A6 'type RoleProfileEntry struct' internal/config/types.go | grep -i effort` returns 0 matches (struct has no Effort field → declarative YAML cannot couple to Go even if a reader existed).

**Mirror parity (post-edit, all 3 pairs):** dynamic-workflows.md 0-diff, session-handoff.md 0-diff (CI-enforced via `rule_template_mirror_test.go` allowlist), output-styles/moai/moai.md 0-diff.

**Template neutrality:** `internal/template/templates/` content under the 4 edited paths contains 0 SPEC IDs, 0 REQ tokens, 0 NEW feedback_ refs, 0 Audit citations. The 2 pre-existing `feedback_` refs in session-handoff.md template mirror (L171 `feedback_worktree_autonomous`, L352 `feedback_large_spec_wave_split`) are grandfathered per doctrine §25.1 allow-list and confirmed not introduced by this SPEC.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-17
run_commit_sha: "<pending — populated at M1 commit; run-phase commits land on worktree-agent-a82c6d9e6f3a73c3b, NOT pushed (orchestrator coordinates push)"
run_status: audit-ready
ac_pass_count: 13
ac_fail_count: 0
preserve_list_post_run_count: 5  # the 5 dirty-tree config YAMLs remain unabsorbed (AC-WEM-012)
l44_pre_commit_fetch: n/a  # L1 worktree branch; orchestrator handles pre-spawn fetch at integration
l44_post_push_fetch: n/a  # no push performed by manager-develop per delegation contract
new_warnings_or_lints_introduced: 0  # doctrine/YAML/JS/MD only; no Go build/lint surface touched
cross_platform_build:
  go_build_main: n/a  # no Go changes (decision #1)
  go_build_windows: n/a
  spec_lint: pending  # orchestrator Trust-but-verify batch
  template_neutrality_ci: pending  # orchestrator Trust-but-verify batch
  mirror_parity_ci: pending  # rule_template_mirror_test.go — orchestrator Trust-but-verify batch
total_run_phase_files: 13  # 9 run-phase edit targets + 4 SPEC artifacts
m1_to_mN_commit_strategy: "per-milestone Conventional Commits with Authored-By-Agent: manager-develop trailer; M1 carries draft→in-progress frontmatter transition on spec.md; progress.md §E.2/§E.3 owned by manager-develop"
```

**Residual debt (ACCEPTED, not fixed per delegation contract):** D1-new (git-diff as PRIMARY for AC-WEM-006 vs struct-field-absence — already the SPEC's iter-2 resolution, no action), D2-new (dynamic-workflows.md file-size ~28KB post-edit — below readability threshold, no action), D3-new (iter/rev naming in plan.md history — cosmetic, no action).

**Gaps (미검증):** orchestrator-side Trust-but-verify 7-item batch NOT run by manager-develop (orchestrator owns this post-run-phase). spec-lint / template-neutrality CI / `rule_template_mirror_test.go` CI NOT run in this L1 worktree (no Go toolchain execution by manager-develop per decision #1 spirit; orchestrator verifies). Actual cost savings from `effort: 'low'` on codemaps-extract.js NOT measured (downstream token-cost A/B, out of scope per acceptance.md §D.2).

**Residual risk:** the `effort` YAML field is declarative-only — orchestrator runtime consumption is an LLM behavior verified by field-existence + documentation, not by a code path (per acceptance.md §D.2 indirect verification). A future SPEC wiring the field into Go runtime enforcement MUST cite this SPEC as the declarative SSOT origin (acceptance.md §D.4).

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — populated by manager-docs>_

## §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase — populated by orchestrator-direct or manager-docs>_
