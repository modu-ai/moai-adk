# SPEC-SYNC-STRATEGY-KEY-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-08-27
plan_audit_iter: 2 (iter-1 verdict FAIL 0.86 blocking D2 — `.moai/reports/t303/plan-audit-SPEC-SYNC-STRATEGY-KEY-001-v0.1.0.md`; D2/D3/D1/D4/D5 applied in v0.2.0, D6/D7 routed to orchestrator)
artifacts: spec.md (0.2.0, draft), plan.md, acceptance.md, progress.md — Tier M (4 files incl. progress skeleton)
tree: d29b8942e (worktree t303, branch WT-sync-strategy-key)

## §E.2 Run-phase Evidence

Tree: `63b4628a6` (worktree `t303`, branch `WT-sync-strategy-key`). Full command + verbatim output battery: `.moai/state/verify/t303/ac-battery.txt`; verdict with the 5-section evidence format: `.moai/reports/t303/run-verdict.md`. `T` = `internal/template/templates`, `D` = `T/.claude/skills/moai/workflows/sync/delivery.md`.

| AC | Severity | Command | Actual Output | Status |
|----|----------|---------|---------------|--------|
| AC-SYK-001 | MUST | `grep -c 'git_strategy.{mode}.workflow' D` / `grep -c 'Read \`github.spec_git_workflow\`' D` | `6` / `0` | PASS |
| AC-SYK-002 | MUST | `grep -rn 'spec_git_workflow' T/ \| wc -l`; refined `\| grep -v 'Legacy key fallback' \| wc -l` | `1` (the D1 fallback sentinel only, delivery.md:33); refined `0` — baseline was 10 | PASS (documented refinement) |
| AC-SYK-003 | SHOULD | `grep -c 'v3.3.0' D`; `grep -n 'main_direct\|feature_branch' D` | `1`; hits only at D:39 and D:40, both inside the fallback mapping table | PASS |
| AC-SYK-004 | MUST | `grep -n '{github-flow, git-flow}' D` | D:29 — unmatched-value stop naming the offending value and the canonical domain, "no PR, no push", missing subkey is not a default | PASS |
| AC-SYK-005 | MUST | `grep -c 'WT-' D`; `awk` WT-block `\| grep -c 'gh pr'` | `6`; `0` gh-pr tokens inside the WT-* block | PASS |
| AC-SYK-006 | MUST | `awk '/^##### Strategy:/{s=$0} /matches no defined route/{c[s]++} END{...}' D`; `grep -c 'Default strategy\|Otherwise,' D` | `##### Strategy: github-flow 1` + `##### Strategy: git-flow 1` (exactly 2 lines); negative probe `0` | PASS |
| AC-SYK-007 | SHOULD | `grep -c 'delivery.md'` / `grep -c 'merge --no-ff'` on `.claude/rules/local/gitflow-lane-protocol.md`; same on `D` | `2` / `0`; `D` = `1` — procedure owned by the shipped skill, cited by the dev rule | PASS |
| AC-SYK-008 | MUST | `grep -rn 'gitflow-lane-protocol' T/ \| wc -l`; `grep -n 'workflow:' T/.moai/config/sections/git-strategy.yaml.tmpl`; `git diff 0931789b6 -- T/ \| grep '^+' \| grep -oE 'SPEC-[A-Z0-9-]+-[0-9]{3}' \| wc -l` | `0`; three `workflow: github-flow` lines (13/45/81), no private value leaked; `0` SPEC tokens on added template lines | PASS |
| AC-SYK-009 | MUST | `go test ./internal/config/ -run TestShippedConfigKeysHaveReaders -count=1`; `grep -c 'github.spec_git_workflow' internal/config/testdata/shipped_key_inventory.yaml` | `ok github.com/modu-ai/moai-adk/internal/config 0.870s` (exit 0); `0` | PASS |
| AC-SYK-010 | MUST | `grep -c 'github.spec_git_workflow'` / `grep -c 'automation.auto_branch'` on `T/.claude/skills/moai-workflow-project/schemas/tab_schema.json` | `0` / `3`; JSON re-parsed valid after the edit | PASS |
| AC-SYK-011 | SHOULD | `sed -n '25p'` on template and local `doc-execution.md` | both: `- Read project configuration: git_strategy.mode, git_strategy.{mode}.workflow, conversation_language` | PASS |
| AC-SYK-012 | MUST | (1) local dead-key grep refined; (2) `diff` tab_schema template↔local; (3) `grep -c 'SPEC-SYNC-PARALLEL-DOCS-001'` local doc-execution.md; (4) `grep -n 'workflow:' .moai/config/sections/git-strategy.yaml` | (1) `0`; (2) empty; (3) `3` — A5 block intact; (4) `workflow: git-flow` at L8 | PASS |

Invariants held (all measured on `63b4628a6`):

| Invariant | Command | Actual Output | Status |
|---|---|---|---|
| Step 3.1 Tier Route A/B gate preserved | `git diff 0931789b6 -- D` inspected for the Route gate hunk | no hunk touches the Route A/B gate block | PASS |
| Template neutrality guard | `go test ./internal/template/ -run TestTemplateNeutralityAudit -count=1` | `ok github.com/modu-ai/moai-adk/internal/template 0.524s` (exit 0) | PASS |
| Adjacent template guards | `go test ./internal/template/ -run 'TestInternalContentLeak\|TestRuleTemplateMirror\|TestCommandsAudit' -count=1` | `ok github.com/modu-ai/moai-adk/internal/template 0.249s` (exit 0) | PASS |
| Embedded templates regenerated | `make build` then `git status --short` | exit 0; empty status — catalog.yaml unchanged because `gen-catalog-hashes.go:22` hashes only each skill's root SKILL.md, not sub-files | PASS |
| Local-only content preserved | `diff` template↔local for delivery.md / doc-execution.md | delivery.md: footer 2 lines only; doc-execution.md: the A5 attribution block only | PASS |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-27
run_commit_sha: 63b4628a6      # M4; the M5 progress/status commit follows and is the branch tip
run_status: complete
ac_pass_count: 12
ac_fail_count: 0
must_pass_ac_pass_count: 9     # AC-001/002/004/005/006/008/009/010/012
preserve_list_post_run_count: 3   # Step 3.1 Route A/B gate; git-strategy.yaml.tmpl github-flow defaults; local-only mirror content (delivery.md footer, doc-execution.md A5 block)
l44_pre_commit_fetch: not-run     # no push in run phase; git-flow lane integration is a separate gated step
l44_post_push_fetch: not-run      # nothing pushed
new_warnings_or_lints_introduced: 0
cross_platform_build:
  status: not-applicable
  reason: documentation/template surface only; no new Go production code. The one Go-adjacent touch is testdata (shipped_key_inventory.yaml), covered by the targeted config test.
total_run_phase_files: 8
m1_to_mN_commit_strategy: one commit per milestone group — M1 c374d9605 (template delivery.md), M2 eb31593d1 (doc-execution + tab_schema), M3 1c9295a07 (tmpl key + inventory row, atomic), M4 63b4628a6 (make build + region-scoped mirror sync + local config + dev rule), M5 (this progress/status commit)
carried_debt:
  - id: B5
    text: "CLAUDE.local.md §2.3 (L180-181, L186) still instructs the operator to re-apply `workflow: gitflow` after `moai update`. That value is now out of the canonical domain and would trip the new unmatched-value stop. CLAUDE.local.md is user-owned — surfaced, not edited."
  - id: OBS-1
    text: "tab_schema.json L517 / L727 bind `git_strategy.personal.auto_branch` / `git_strategy.team.auto_branch`, both missing the `.automation` level the Go struct and git-strategy.yaml.tmpl actually use. Pre-existing, outside this SPEC's scope envelope."
  - id: OBS-2
    text: "The rebound SPEC-branching question now overlaps the per-mode auto_branch questions in the personal/team batches. REQ-SYK-010 prescribed the rebind explicitly; the overlap is reported rather than resolved unilaterally."
  - id: OBS-3
    text: "The template-neutrality CI guard triggers on `pull_request: branches: [main]`. Under the repo's git-flow model card branches carry no PR, so this guard may not fire for this card on CI — it was run locally instead (exit 0)."
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
