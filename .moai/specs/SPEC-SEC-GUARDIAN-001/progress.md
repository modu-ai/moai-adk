# SPEC-SEC-GUARDIAN-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-07-24
tier: L
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (5 — Tier L complete) + progress.md skeleton
epic: SECURITY-ABSORB (SPEC-2 of the cohort; SPEC-1 = SPEC-SEC-DEEPSCAN-001)
absorption_source: Anthropic official `security-guidance` plugin — in-session always-on 3-layer guardian (Candidate B), reimplemented natively (NOT plugin-installed)
layering: SPEC-2 = light + always-on + inline (L1 PostToolUse pattern warnings / L2 Stop turn-diff review / L3 opt-in commit cross-file review); SPEC-1 = heavy + on-demand + explicit (/moai review --deep). No overlap (spec.md §C).
spec_id_precheck: PASS (regex ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ — decomposition SPEC | SEC | GUARDIAN | 001; Bash one-liner output "regex: PASS"; no dir collision)
go_in_scope: YES (internal/hook/security/ pattern engine + 3 layer handlers + internal/cli/hook.go subcommands + tests — compiled, NOT template content) + template-first shell wrappers + settings.json wiring
default_posture: L1+L2 ON (regex-only, advisory); L3 OFF (dormant, opt-in); blocking everywhere opt-in only
open_clarifications: 0 (both settled by orchestrator — Layer-3 surface = L3-A extend-sync-gate + orchestrator escalation; escalation delivery = orchestrator-mediated Agent())
ac_count: 25 (18 MUST + 7 SHOULD) — AC-SG-025 added for REQ-SG-001 flagship coverage
audit_fix: plan-audit PASS-WITH-DEBT 0.84 (6 additive non-blocking defects) → D1-D6 applied (D1 AC-SG-004 self-collision → Go test + scoped comment-excluded grep; D2 REQ-SG-001 → new AC-SG-025; D3 AC-SG-009 → REQ-SG-022,023; D4 AC-SG-010 → REQ-SG-030,031; D5 plan.md L21/L28 stale-fork corrected to L3-A sibling-hook non-rewrite; D6 AC-SG-023 concrete .js=0 check + AC-SG-003 bounded range 20≤N≤30)
next: plan-audit re-run (plan-auditor) → Implementation Kickoff Approval → run-phase (manager-develop, cycle_type=tdd)

## §E.2 Run-phase Evidence

Run-phase implemented via TDD (RED-GREEN-REFACTOR) by manager-develop. The Go pattern engine
+ 3 layer handlers live in `internal/hook/security/` (compiled, NOT template); the 3 shell
wrappers + settings.json wiring are Template-First (byte-lockstep local sync). The pre-existing
`package security` (ast-grep scanner) was EXTENDED additively — new files `patterns.go` / `scan.go`
/ `diff.go` / `guardian.go` use non-colliding symbols (`VulnClass` / `VulnSeverity` / `GuardianFinding`
distinct from the existing `Severity` / `Finding` / `ScanResult`).

| AC | Severity | Status | Verification command | Actual Output |
|----|----------|--------|----------------------|---------------|
| AC-SG-001 | MUST | PASS | `go test -run TestLayer1Scan ./internal/hook/security/` | ok — finding on `yaml.load` fixture, silent on clean |
| AC-SG-002 | MUST | PASS | `go test -run TestPatternsLanguageNeutral ./internal/hook/security/` | ok — hardcoded-secret fires across all 16 language fixtures |
| AC-SG-003 | MUST | PASS | `go test -run TestPatternClassCoverage ./internal/hook/security/` | ok — 10 classes present, 28 patterns (20 ≤ 28 ≤ 30) |
| AC-SG-004 | MUST | PASS | `go test -run TestLayer1NoLLMOrSubprocess ./…` + scoped grep on scan.go/diff.go | exit 0; grep → 0 matches |
| AC-SG-005 | MUST | PASS | `go test -run TestLayer1NeverBlocks ./…` + `TestLayer1Surfaces` (no decision field) | ok — advisory only, no decision |
| AC-SG-006 | SHOULD | PASS | `grep -A7 security-scan settings.json.tmpl \| grep async` | `"async": true` present (template + local) |
| AC-SG-007 | MUST | PASS | `go test -run TestLayer2TurnDiff ./…` | ok — engine runs over added diff lines |
| AC-SG-008 | SHOULD | PASS | `go test -run TestLayer2Surfaces ./…` | ok — high-severity finding in systemMessage |
| AC-SG-009 | MUST | PASS | `go test -run TestLayer2AdvisoryDefault ./…` (folded into TestLayer2Surfaces) | ok — no LLM/no block without opt-in |
| AC-SG-010 | MUST | PASS | `go test -run TestLayer3CrossFile ./…` | ok — changed+related read; cross-file-idor targeted |
| AC-SG-011 | MUST | PASS | `go test -run TestLayer3DormantDefault ./…` | ok — silent no-op exit 0 when flag unset |
| AC-SG-012 | SHOULD | PASS | design.md §L3 (L3-A sibling commit-time Stop hook + orchestrator escalation) | documented |
| AC-SG-013 | MUST | PASS | `grep -rn 'AskUserQuestion\|mcp__askuser' .claude/hooks/moai/ internal/hook/security/ …` | 0 matches |
| AC-SG-014 | SHOULD | PASS | design.md §C orchestrator-translation path for a block signal | documented |
| AC-SG-015 | MUST | PASS | `go test -run TestAdvisoryFirst ./…` | ok — advisory default; decision:block only with MOAI_SECURITY_BLOCKING |
| AC-SG-016 | SHOULD | PASS | `go test -run TestSkipHookAudit ./…` | ok — `--skip-hook` appends to hook-skip.log |
| AC-SG-017 | MUST | PASS | `diff` each wrapper local↔template `.sh`/`.sh.tmpl`; settings.json security entries in both | 3 wrappers IDENTICAL; settings entries present both sides |
| AC-SG-018 | MUST | PASS | `grep 'SPEC-SEC-GUARDIAN\|2026-07-24' handle-security-*.sh.tmpl` + `InternalContentLeak\|Neutrality` tests | 0 matches; tests PASS |
| AC-SG-019 | MUST | PASS | `go build ./...` + `GOOS=windows go build ./...` + `go test -cover ./internal/hook/security/...` | exit 0 / exit 0 / coverage 90.0% |
| AC-SG-020 | MUST | PASS | `ls patterns.go` + `grep 'yaml.load\|innerHTML' handle-security-*.sh.tmpl` | patterns.go exists; 0 scattered patterns |
| AC-SG-021 | MUST | PASS | `go test -run TestFailOpen ./…` + invalid-JSON smoke | ok — non-git dir / empty stdin / bad JSON → exit 0 silent |
| AC-SG-022 | SHOULD | PASS | `go test -run 'TestHookOutputSchema' ./…` (folded into TestLayer1Surfaces) | ok — additionalContext / systemMessage / decision schema, no unknown fields |
| AC-SG-023 | MUST | PASS | `go test -run Catalog ./internal/template/...` + `find templates -name '*.js' \| wc -l` | PASS (10 agents / 28 skills unchanged); 0 `.js` |
| AC-SG-024 | SHOULD | PASS | design.md §A + spec.md §A.3 SPEC-1 vs SPEC-2 layering | documented |
| AC-SG-025 | MUST | PASS | `grep 'plugin install\|marketplace\|.claude-plugin' …` | 0 matches (native Go + hooks + settings only) |

**Invariants**: existing test suite unbroken (full `go test -count=1 ./...` exit 0); pre-existing
`package security` ast-grep files untouched; `internal/hook/security/` additive only.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-24
run_commit_sha: 8d12ebab8
run_status: complete
ac_pass_count: 25
ac_fail_count: 0
preserve_list_post_run_count: 0   # PRESERVE list untouched: existing wrappers, sync-gate, catalog counts, existing Go handlers
l44_pre_commit_fetch: not-run (run-phase local commit only; no push per Tier-L Route-B — manager-git owns PR)
l44_post_push_fetch: not-run (no push in run-phase)
new_warnings_or_lints_introduced: 0   # golangci-lint on changed packages → 0 issues
cross_platform_build:
  host: pass   # go build ./... exit 0
  windows: pass   # GOOS=windows GOARCH=amd64 go build ./... exit 0
coverage_internal_hook_security: 90.0%   # >= 85% threshold
total_run_phase_files: 20   # 8 Go (4 impl + 4 test) + 3 config/cli-test edits + 2 cli/config edits + 3 wrapper tmpl + 3 wrapper local + settings.json.tmpl + settings.json + spec.md + progress.md
m1_to_mN_commit_strategy: plan-artifacts commit + Go-core commit (draft→in-progress) + Template-First commit
boundary_grep: clean (0 AskUserQuestion/mcp__askuser in hooks or internal/hook/security)
full_suite: pass (go test -count=1 ./... exit 0)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-24
sync_status: audit-ready
sync_commit_sha: d2690740e169213529727f32add7875f0509959b
changelog_entry_position: "[Unreleased] > ### Added (Epic SECURITY-ABSORB, joint entry with SPEC-SEC-DEEPSCAN-001)"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed"
readme_update: "one-line --deep mention added to the /moai Slash Commands review row (shared row, this SPEC ships the guardian layer under the same Epic)"
docs_site_4locale: "DEFERRED — follow-up required (not performed in this sync per scope note)"
b12_self_test_a: "grep -c 'SPEC-SEC-DEEPSCAN-001\|SPEC-SEC-GUARDIAN-001' CHANGELOG.md -> 0 before emission (no duplicate)"
b12_self_test_b: "acceptance.md AC row count 25 == CHANGELOG claim '25/25 AC'"
b12_self_test_c: "ls .moai/specs/SPEC-SEC-DEEPSCAN-001/spec.md .moai/specs/SPEC-SEC-GUARDIAN-001/spec.md CHANGELOG.md README.md -> all exist"
```
