# progress.md — SPEC-MCP-DEFAULT-ON-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-08-12
tier: M
artifacts: spec.md, plan.md, acceptance.md (+ this progress.md)

## §E.2 Run-phase Evidence

### AC PASS/FAIL matrix (14 ACs)

| AC | Milestone | Status | Evidence |
|----|-----------|--------|----------|
| AC-A-001 | M1 | PASS | `grep -c 'AMENDED' .moai/specs/SPEC-MOAI-MCP-SERVER-001/spec.md` → `2`; `grep '^status:'` → `status: completed`. Prose-only verification gap (plan-auditor D1) recorded — no binary command for the "amended in place" structural check beyond the AMENDED marker count. |
| AC-A-002 | M1 | PASS | `grep -c 'AMENDED' .moai/specs/SPEC-TREND-MCP-001/spec.md` → `2`; `git diff 0e24dde06 -- .moai/specs/SPEC-TREND-MCP-001/spec.md` shows REQ-TMC-002 body line byte-unchanged; REQ-TMC-004 `$comment` clause unchanged; amended AC-TMC-001 verification line still contains `$comment`. |
| AC-A-003 | M1 | PASS | Reversal rationale present in prose adjacent to each amended criterion (both SPECs' HISTORY `### Amendments` sub-sections + the acceptance.md amendment-notice block). Prose-only verification (plan-auditor debt). |
| AC-A-004 | M2 | PASS | `jq -c '.mcpServers \| keys' internal/template/templates/.mcp.json` → `["moai"]`; `jq '.mcpServers.moai \| has("env")'` → `false`; `$schema` + `staggeredStartup` present and unchanged. |
| AC-A-005 | M2 | PASS | `jq -c '.mcpServers \| keys \| length' .mcp.json` → `4` (chrome-devtools, context7, moai, playwright); `grep -c 'diverge' plan.md` → `≥1` (§A.2). |
| AC-A-006 | M2 | PASS | `go test -run TestMCPNeutrality ./internal/template/...` → PASS; `mcpAllowedActiveKeys = {moai}`; count assertion = 1; forbidden-token regex block byte-unchanged from base (git diff shows no `regexp.MustCompile` lines changed). |
| AC-A-007 | M2 | PASS | `grep -cE '(sk-\|ghp_\|Bearer )' internal/template/templates/.mcp.json` → `0`; `go test -run TestInternalContentLeak ./internal/template/...` → PASS. |
| AC-A-008 | M3 | PASS | `TestProvisionMCPEntryUnlessDeclined_Default` PASS (declined=false → provisions the moai entry through provisionMoaiMCPServerEntryAt → mutateClaudeJSONAtomic). |
| AC-A-009 | M3 | PASS | `TestProvisionMCPEntryUnlessDeclined_Declined` PASS (declined=true → no .mcp.json, silent stdout+stderr). |
| AC-A-010 | M3 | PASS | `grep -A6 'mcp_provision' internal/cli/wizard/questions.go \| grep 'Default:'` → `Default: "true"`; Title positive (`"Provision the moai MCP server?"`); `grep -rc 'mcp_tools_opt_in' internal/cli/` → 0. |
| AC-A-011 | M3 | PASS (SHOULD) | `grep -rc 'provisionMCPEntryIfOptedIn' internal/cli/` → 0; function renamed to `provisionMCPEntryUnlessDeclined`; doc comment rewritten for default-on semantics. |
| AC-A-012 | M4 | PASS | `TestMergeUserFiles_PreservesUserMCPEntry` PASS (moai + my-tool both survive MergeUserFiles 3-way merge); `grep -c '".mcp.json"' internal/cli/update_template_sync.go` → `≥1`. |
| AC-A-013 | M4 | PASS | `grep -c 'no longer ships an MCP template' internal/cli/update_template_sync.go` → `0`; corrected comment now states .mcp.json IS shipped and IS a merge target. |
| AC-A-014 | M5 | PASS | `grep -rc 'Opt-in default-off' internal/cli/wizard/translations.go` → `0`; `grep -rEc 'SPEC-[A-Z]\|REQ-[A-Z]' internal/cli/wizard/translations.go internal/template/templates/.mcp.json` → `0` for both; `go test -run 'TestMCPNeutrality\|TestInternalContentLeak' ./internal/template/...` → PASS. |

### Invariant rows

| Invariant | Status | Evidence |
|-----------|--------|----------|
| C-A-1 (no tool handler/schema change) | PASS | PRESERVE — internal/cli/mcp_server.go tool registration/handlers/schemas untouched (git log -- internal/cli/mcp_server.go = 0 commits this SPEC). |
| C-A-2 (no resolved secrets) | PASS | AC-A-007. |
| C-A-3 (template §25 neutrality) | PASS | AC-A-014 + TestMCPNeutrality + TestInternalContentLeak PASS. |
| C-A-4 (provisioning best-effort non-fatal) | PASS | `TestProvisionMCPEntryUnlessDeclined_FailureIsNonFatal` PASS. |
| C-A-5 (explicit decline honored) | PASS | AC-A-009. |

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-08-12
run_commit_sha: 0c44ef1b8
run_status: audit-ready
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 8
l44_pre_commit_fetch: n/a (worktree isolated, no push)
l44_post_push_fetch: n/a (no push — repo-local Route B, manager-git owns push+PR)
new_warnings_or_lints_introduced: 0
baseline_test_failures:
  - internal/template TestTemplateNoInternalContentLeak — zone-registry.md SPEC-ID leaks (SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001, SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001); confirmed present at base 0e24dde06; this SPEC touched neither zone-registry.md nor this test.
  - internal/template TestRuleDateProvenance — zone-registry.md date leaks (2026-05-04, 2026-05-09 at lines 629-630); confirmed present at base 0e24dde06.
  - internal/astgrep TestRuleSeed — dogfood-experimental language stubs (csharp/elixir/php empty rule dirs, CLAUDE.local.md §2.2); this SPEC touched no astgrep code or config.
cross_platform_build:
  darwin: PASS (go build ./... exit 0)
  linux: PASS (GOOS=linux GOARCH=amd64 go build ./... exit 0)
  windows: PASS (GOOS=windows GOARCH=amd64 go build ./... exit 0)
total_run_phase_files: 18
m1_to_mN_commit_strategy: per-milestone conventional commits (M2 2b380941e, M3 b77339698, M4 82dcaf688, M5 0c44ef1b8); M1 confirmation-only (no edit, no commit — amendments pre-existing at HEAD 01d9227f2).

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- tier: M
- scope: ~8-10 files — `internal/cli/` (init.go, wizard/questions.go, wizard/translations.go, update_template_sync.go, init_mcp_provision_test.go) + `internal/template/` (.mcp.json, mcp_template_neutrality_test.go, deployer) + docs (settings-management.md + mirror, template CLAUDE.md)
- domain count: multi-surface (cli + template + wizard + docs) but a single logical unit (gate-direction flip + shape change + merge safety + doc sweep)
- file language mix: Go + JSON + markdown
- concurrency benefit: LOW — milestones are sequentially dependent (M1 amendments frame the contract; M2-M5 implement against it; neutrality test must land with the template shape; wizard flip must land with the gate rename)
- Agent Teams prereqs: N/A (static layer retired)

Decision: sub-agent (Mode 5)

Justification: Tier M with a sequential milestone dependency chain and a single logical unit of work (a gate flip, not parallelizable new code). Per Anthropic's coding-task parallelism caveat, coding-heavy sequentially-dependent work is the Mode 5 default. M1 is a confirmation grep (amendments already landed in the worktree per plan-auditor D2); M2-M5 are the real implementation. manager-develop with the Tier M Section A-E delegation template, cycle_type per quality.yaml.

Implementation Kickoff Approval: PASSED (owner approved "A+amendment commit then run" at plan→run gate, this session).
Phase 1 Plan Audit Gate: PASS-WITH-DEBT 0.82 (review-1.md); debts (AC-A-001/003 verification lines, D1 sync-time amendment canonicalization) are non-blocking and tracked for run/sync resolution.
