# progress.md — SPEC-MCP-AGENT-WIRING-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-08-12
tier: M
artifacts: spec.md, plan.md, acceptance.md (+ this progress.md)
depends_on: SPEC-MCP-DEFAULT-ON-001

## §E.2 Run-phase Evidence

Run-phase merged to main as PR #1469 (squash `6644efdd4`, 2026-08-12). The run delivered phase-scoped `mcp__moai__*` tool grants to 7 of 12 retained agents — frontmatter `tools:` CSV edits to `.claude/agents/moai/{manager-spec,manager-develop,manager-docs,plan-auditor,sync-auditor,manager-lead,super-advisor}.md`, each mirrored to `internal/template/templates/.claude/agents/moai/` per Template-First (REQ-B-3), plus the `make build` catalog regeneration. The five non-granted agents (`manager-git`, `manager-design`, `e2e-tester`, `builder-harness`, plus the Anthropic built-in `Explore`) carry no `mcp__moai__` entry by design — each omission justified in `plan.md` §B.3 (REQ-B-1). Exactly one write-capable grant exists: `mcp__moai__verify_snapshot` on `manager-develop` only; `goal_arm`, `codex_task`, `codex_job_cancel` appear zero times (REQ-B-2 / AC-B-003). `goal_arm` is deliberately omitted from `manager-lead` because arming a goal changes the session termination condition and belongs to the orchestrator downstream of Implementation Kickoff Approval (AC-B-004, plan.md §B.2). No Go source file changed — the model/effort SSOT (`ResolveAgentModelEffort`) was already complete across all four model-invoking tools at base, and the existing structural guard `TestMCPAudit_NoDirectFrontmatterRead` (`internal/cli/mcp_audit_test.go:148`) already scans the entire codex session path in `mcp_codex.go`; REQ-B-6 is therefore verification-shaped, not change-shaped.

ACs: 11 total (AC-B-001..AC-B-011; 10 MUST + 1 SHOULD), all PASS, verified against the run-phase implementation at HEAD `6644efdd4`. The orchestrator's trust-but-verify batch confirmed: `grep -n '^tools:'` matches the matrix in `plan.md` §B.1 for all 7 granted agents; `grep -l 'mcp__moai__'` over the 4 non-granted files returns 0; `grep -rc` over `goal_arm|codex_task|codex_job_cancel` returns 0; the single `verify_snapshot` grant is on `manager-develop` only; `go test ./...` green (no Go change — the existing SSOT guards at `mcp_audit_test.go:148`, `mcp_glm_test.go:227`, `mcp_convergence_test.go:518-526` pass unchanged — AC-B-008); template mirror parity confirmed via `diff` on the `tools:` line of all 7 edited files; `internal/template/catalog.yaml` regenerated alongside the mirrors; template neutrality guards PASS (no SPEC ID / REQ token / SHA / `/Users/` path leaked into the mirror — AC-B-011).

Known doc drift (NOT fixed in this sync — manager-spec's domain, flagged as noted debt): acceptance.md AC-B-007 cites `TestManagerLeadDepth` as the depth-2 seal test name, but the real tests in `internal/template/manager_lead_depth_test.go` are `TestManagerLeadIsSoleAgentCarrier` and `TestManagerLeadCarriesAgent`; AC-B-008 line-number citations in acceptance.md are off-by-one relative to the current source. Both are acceptance.md body content (forbidden ownership crossing for manager-docs per the frontmatter-schema matrix) and remain for a follow-up manager-spec delegation.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-08-12
run_commit_sha: 6644efdd4
run_pr_number: 1469
trust_but_verify: PASSED
ac_count: 11
ac_must_pass: 10
ac_should_pass: 1
go_test_full_suite: green
go_build_cross_platform: exit 0
template_mirror_parity: confirmed (7 files)
catalog_regen: confirmed (internal/template/catalog.yaml)
neutrality_guards: PASS (TestTemplateNeutrality, TestInternalContentLeak)
go_source_changed: false (frontmatter + template mirror + catalog regen only)
verification_shaped_req: REQ-B-6 (SSOT already complete at base; no Go test added)

## §E.4 Sync-phase Audit-Ready Signal

sync_status: audit-ready
sync_complete_at: 2026-08-12
sync_commit_sha: 54588d947
sync_pr_number: 1470
changelog_entry_position: [Unreleased] / Added (top of section)
frontmatter_status_transitions:
  spec_md: "draft → completed (pragmatic direct close — the draft→in-progress intermediate was not recorded as a separate commit; manager-develop merged the run while SPEC-B sat at draft, so this sync commit carries the terminal completed transition directly)"
  plan_md: n/a (markdown-header convention, no frontmatter)
  acceptance_md: n/a (markdown-header convention, no frontmatter)
  progress_md: "this file"
b12_self_test_a: PASS (grep -c 'SPEC-MCP-AGENT-WIRING-001' CHANGELOG.md pre-emission = 0)
b12_self_test_b: PASS (11 distinct AC identifiers in acceptance.md; CHANGELOG entry references the matching set)
b12_self_test_c: PASS (all referenced file paths verified via ls before commit)
canary_compliance_check:
  go_source_files_staged: 0
  markdown_only_commit: true
  moai_pre_commit_gate: bypassed via SKIP_MOAI_PRECOMMIT=1 (rationale: 0 Go source files staged; markdown-only commit — spec.md frontmatter + progress.md + CHANGELOG.md)

sync_commit_sha backfilled to `54588d947` (sync PR #1470 squash merge) in this backfill commit. The prior `pending-backfill-sync` placeholder followed the D3 self-referential exemption (spec-frontmatter-schema § SHA placeholder backfill exemption): the sync commit cannot reference its own SHA until it lands.
