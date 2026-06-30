# Progress — SPEC-GLM-MODEL-ALLOWLIST-001

## §E.1 Plan-phase Audit-Ready Signal

- **Tier**: M (standard) — 확정 (plan.md §B)
- **SPEC ID self-check**: `decomposition: SPEC ✓ | GLM ✓ | MODEL ✓ | ALLOWLIST ✓ | 001 ✓ → PASS`
  (canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`)
- **Artifacts**: spec.md + plan.md + acceptance.md + progress.md (4-file plan-phase set)
- **Frontmatter**: 12 canonical fields validated, `created`/`updated` (NOT `created_at`/`updated_at`),
  `tags` comma-separated string (NOT `labels` array), status: draft.
- **Out of Scope**: 5개 `### Out of Scope —` H3 하위 섹션 + `-` bullets (spec.md §F).
- **Requirements**: REQ-GMA-001~008 (GEARS — Event-driven/Ubiquitous/State-driven/Where 혼합).
- **Root cause**: 검증된 인과 사슬 6단계 (spec.md §A.2, 이번 세션 Read/grep 전수 확인).
- **Approach**: A 권장 (settings.local.json `enforceAvailableModels:false`, inject↔clear 패리티),
  B는 우선순위 실측(M1/AC-GMA-006) 실패 시 fallback.
- **Selection origin**: SPEC-CC2178-MODEL-POLICY-REPAIR-001 GLM 사각지대 후속(supersession 아님).
- 다음: plan-auditor 독립 감사 → 구현 착수 승인 → `/moai run`.

## §E.2 Run-phase Evidence

**Approach C (static template allowlist expansion) — implemented. cycle_type=tdd.**

### AC PASS/FAIL Matrix (acceptance.md AC-GMA-001~008)

| AC | REQ | Status | Verification command | Actual output |
|----|-----|--------|----------------------|---------------|
| AC-GMA-001 | REQ-GMA-001 | DEFERRED-MANUAL | live `moai glm` session (not scriptable in subagent; GLM in dev project forbidden per CLAUDE.local.md §13) | premise near-definitional (allowlisting canonical alias `opus[1m]` removes the CC 2.1.176 redirect block) — see Manual M1 Procedure below |
| AC-GMA-002 | REQ-GMA-002 | PASS | `git diff internal/template/templates/.claude/settings.json.tmpl` + grep | only line 378 `availableModels` changed; line 377 `"model": "sonnet"` + line 379 `"enforceAvailableModels": true` byte-unchanged (diff shows single `-/+` pair on the availableModels line) |
| AC-GMA-003 | REQ-GMA-003 | PASS | `go test -run TestApplyGLMMode_NoSettingsLocalPollution ./internal/cli/` + `grep -rn 'enforceAvailableModels\|availableModels' internal/cli/ \| grep -v _test.go` | `ok internal/cli 0.622s`; grep = NO output (no Go runtime write of model-allowlist keys) |
| AC-GMA-004 | REQ-GMA-004 | PASS | `grep -n 'availableModels' internal/template/templates/.claude/settings.json.tmpl` | line 378: `["sonnet", "opus", "haiku", "opus[1m]", "sonnet[1m]"]` |
| AC-GMA-005 | REQ-GMA-005 | PASS | `grep -n 'modelCanonical' internal/web/validate.go` | `modelCanonical = []string{"opus", "opus[1m]", "sonnet", "sonnet[1m]", "haiku", "opusplan"}` — added `opus[1m]`/`sonnet[1m]` are both subset members |
| AC-GMA-006 | REQ-GMA-006 | DEFERRED-MANUAL | M1 /tmp-project live unblock measurement (un-runnable in subagent) | implemented M2/M3 (low-risk/reversible: 2 canonical aliases added to allowlist); manual procedure documented below for orchestrator/user post-deploy |
| AC-GMA-007 | REQ-GMA-007 | PASS | `go test -run TestSettingsTemplateDefaultModelLever ./internal/template/` | `ok internal/template 5.143s` — test updated to expect expanded 5-entry allowlist; `model:"sonnet"` + `enforceAvailableModels:true` assertions kept (invariant) |
| AC-GMA-008 | REQ-GMA-008 | PASS | `go test -run TestRuleTemplateMirrorDrift ./internal/template/` + 2 greps | `ok`; `grep -c 'GLM-mode reconciliation'` = 1 in BOTH trees; `model-policy.md` registered in `workflowOptMirroredPaths` (rule_template_mirror_test.go) |

### Manual M1 Procedure (AC-GMA-001/006 — DEFERRED-MANUAL, orchestrator/user post-deploy)

The live `moai glm` unblock confirmation is NOT scriptable in a subagent AND GLM
integration in the dev project is forbidden (CLAUDE.local.md §13). The Approach C
premise is near-definitional: allowlisting the canonical alias `opus[1m]` removes
the CC 2.1.176 `enforceAvailableModels`-redirect block. To confirm post-deploy:

1. In an affected project (e.g. `~/moai/claude.mo.ai.kr`), run `moai update` to
   re-render `settings.json` with the expanded `availableModels`.
2. Verify the rendered allowlist: `grep availableModels .claude/settings.json` →
   shows `opus[1m]` and `sonnet[1m]`.
3. Launch `moai glm` and check `/model` (or the active-model statusline):
   the active model is the GLM high model (glm-5.2), NOT claude-sonnet.
4. Confirm absence of the warning "Model opus[1m] is restricted by your
   organization's settings. Using claude-sonnet instead."

If the Sonnet fallback persists despite `opus[1m]` being in the allowlist, the
Approach C premise is falsified -> fall back to Approach A (settings.local.json
runtime override, team/CG-only) per spec.md §F, via a blocker report to the
orchestrator.

### §25 Neutralization Decision (plan-auditor D1-minor)

The pre-existing template mirror of `model-policy.md` carried internal-work dates
(research-milestone dates plus GitHub issue-close dates) in the
`[1m] Constraint Re-Verification` section. To enroll `model-policy.md` into
byte-parity (`workflowOptMirroredPaths`) AND keep the template mirror §25-clean,
the internal dates were **STRIPPED from BOTH trees** (live + template) — the
load-bearing content (issue numbers, open/closed state, labels, verdict) is
preserved; only the bare ISO dates are removed. Verified clean: `model-policy.md`
has **0 matches** even under `MOAI_TEMPLATE_LEAK_STRICT=1` (the future strict
date-detection tier). Decision: STRIP (chosen over document-residual-debt; low-risk
since the dates were internal-trail noise for a distributed template).

### Files changed (Section D allow-list, exactly 5 source files)

- `internal/template/templates/.claude/settings.json.tmpl` (line 378 availableModels expansion)
- `internal/template/settings_test.go` (TestSettingsTemplateDefaultModelLever — 5-entry expectation, invariants kept)
- `.claude/rules/moai/development/model-policy.md` (GLM-mode reconciliation subsection + date strip)
- `internal/template/templates/.claude/rules/moai/development/model-policy.md` (byte-identical mirror)
- `internal/template/rule_template_mirror_test.go` (model-policy.md enrolled in workflowOptMirroredPaths)

No Go runtime code changed (glm.go / launcher.go / settings.go untouched). No
`embedded.go` golden file exists (directory `//go:embed all:templates`); `make build`
regenerated the binary + catalog.yaml (catalog unchanged — settings/rules edits do
not affect skill/agent catalog hashes).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-06-23"
run_commit_sha: "e84c05e6c8b71c701bdea259566fbbb35cd6fb15"
run_status: implemented
ac_pass_count: 6              # AC-GMA-002/003/004/005/007/008 mechanical PASS
ac_deferred_manual_count: 2   # AC-GMA-001/006 live-GLM (orchestrator/user post-deploy)
ac_fail_count: 0
preserve_list_post_run_count: 0   # glm.go / launcher.go / settings.go / settings.local.json all unchanged
l44_pre_commit_fetch: "0 0 (origin/main == HEAD at run start; no parallel-session race)"
l44_post_push_fetch: "0 0 (HEAD == origin/main == e84c05e6c post-push; fully synced)"
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues (== baseline 0)
cross_platform_build:
  native: pass
  windows_amd64: pass
total_run_phase_files: 5    # 5 source files (Section D allow-list)
m1_to_mN_commit_strategy: "single M1 commit (Approach C = one logical change); draft -> in-progress on this commit per ownership matrix"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: "2026-07-01"
sync_commit_sha: "<backfill-after-commit>"
sync_status: completed
ac_matrix_summary: "6 AC PASS (AC-GMA-002/003/004/005/007/008) + 2 DEFERRED-MANUAL (AC-GMA-001/006) — all in-scope mechanical verification complete; deferred manual procedures documented in §E.2 Manual M1 Procedure for post-deploy live-GLM confirmation"
frontmatter_status_transitions:
  spec_md_status: "in-progress → completed"
  spec_md_updated: "2026-06-23 → 2026-07-01"
changelog_entry_position: "[Unreleased] § Fixed (SPEC-GLM-MODEL-ALLOWLIST-001 inline Korean entry with CHANGELOG content-anchor)"
readme_update_required: false
  rationale: "technical change with no user-facing doc delta (template-only settings.json.tmpl change)"
b12_self_test_results:
  changelog_duplicate_grep: "0 (no prior SPEC-GLM-MODEL-ALLOWLIST-001 entry in CHANGELOG.md)"
  ac_count_match: "8 AC rows in acceptance.md (6 PASS + 2 DEFERRED-MANUAL), 8 rows in sync-phase coverage"
  file_path_verification: "all 5 Section D allow-list files verified exist via ls: PASS"
```
