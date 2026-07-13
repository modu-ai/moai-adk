# SPEC-DOCTOR-PROMOTION-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifact set authored 2026-07-13 by manager-spec: spec.md (GEARS REQ-DP-001..004, inline AC-DP-001..004, exclusions) + plan.md (M1-M5, decision-reversibility ordered) + this skeleton. Tier S, status `draft`.

Plan-audit: PASS 0.80 (Tier S threshold 0.75). Major-defect fixes applied 2026-07-13 (spec v0.1.1): D1 — AC-DP-001 gate hardened against vacuous `go test -run` pass; D3 — plan.md M2 corrected to Detail-text-only suggestion (fix-aggregation collects CheckFail rows only; check is Warn-only per REQ-DP-004).

## §E.2 Run-phase Evidence

Run-phase executed 2026-07-13 by manager-develop (cycle_type=tdd, RED→GREEN→REFACTOR). RED evidence: `go test -run TestCheckPluginDeployment ./internal/cli` → exit 1, `undefined: checkPluginDeployment` (log: `.moai/state/verify/doctor-promotion/red.log`). All commands run from repo root at commit `31c5b60bd`.

| Row | Gate (verbatim command) | Actual Output | Status |
|-----|------------------------|---------------|--------|
| AC-DP-001 | `grep -c 'func TestCheckPluginDeployment' internal/cli/doctor_promotion_test.go` AND `go test -run TestCheckPluginDeployment ./internal/cli` | grep=1 (≥1); test exit=0, 6/6 subtests PASS (`.moai/state/verify/doctor-promotion/e1-acdp001.log`) | PASS |
| AC-DP-002 | `grep -c 'checkPluginDeployment' internal/cli/doctor.go` | 3 (≥2: 1 registration in moaiChecks + 1 func def + 1 doc comment) | PASS |
| AC-DP-003 | `go test ./internal/cli` | exit=0, `ok github.com/modu-ai/moai-adk/internal/cli 20.220s` (`.moai/state/verify/doctor-promotion/e1-full.log`) | PASS |
| AC-DP-004 | `grep -c 'moai init' internal/cli/doctor_promotion_test.go` + Detail assertion in Warn rows | grep=3; Warn rows assert Detail contains `moai init` + deployed + binary versions; Detail-producing site doctor.go `checkPluginDeployment` fmt.Sprintf | PASS |
| INV never-Fail (REQ-DP-004) | table rows: malformed yaml + unreadable(dir) fixtures; blanket `Status == CheckFail → t.Fatalf` on every row | 6/6 subtests PASS incl. both graceful rows | PASS |
| INV no-writes (REQ-DP-002) | code inspection: check path uses `os.ReadFile` only; `git show 31c5b60bd --numstat -- internal/cli/doctor.go` → 56 added, 0 deleted | no write API in check path; suggestion rides Detail text only | PASS |
| INV fix-aggregation untouched | `git show 31c5b60bd -- internal/cli/doctor.go` deletions = 0 | doctor.go:120-128 CheckFail-only aggregation block unmodified | PASS |
| INV golden reconciliation | `UPDATE_GOLDEN=1 go test ./internal/cli -run 'TestDoctor_Current|TestDoctor_NoColor'` then `git diff --stat internal/cli/testdata/` | 3 goldens shifted by exactly +1 OK row (`Plugin Deployment  no plugin marker (binary-managed)`) + `Pass 11→12` pill; regenerated via established UPDATE_GOLDEN convention (plan.md §B mitigation) | PASS |
| INV cross-platform (spec §5) | `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` | both exit=0; no syscall, stdlib os/filepath/regexp only | PASS |
| INV LOC budget (spec §5) | `git show --numstat HEAD -- internal/cli/doctor.go internal/cli/doctor_promotion_test.go` | doctor.go +56/-0, doctor_promotion_test.go +88/-0 → 144 < 150 | PASS |
| INV lint/vet | `go vet ./internal/cli`; `golangci-lint run --timeout=2m --new-from-rev=a6b72043a ./internal/cli/...` | vet exit=0; lint `0 issues.` (`.moai/state/verify/doctor-promotion/e5-lint.log`) | PASS |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-13
run_commit_sha: 31c5b60bd  # M1-M4 code commit; this M5 evidence commit follows it
run_status: all-green
ac_pass_count: 4
ac_fail_count: 0
preserve_list_post_run_count: 0  # TDD greenfield check — no PRESERVE list; zero pre-existing tests modified or deleted
l44_pre_commit_fetch: "git fetch origin main + rev-list --left-right → 0 25 (origin ahead 0, local ahead 25) — clean, no divergence"
l44_post_push_fetch: "N/A — push deferred to user per run directive (parallel-workstream commits share this branch)"
new_warnings_or_lints_introduced: 0  # golangci-lint --new-from-rev=a6b72043a → 0 issues; go vet clean
cross_platform_build:
  darwin_native: exit 0
  windows_amd64: exit 0  # GOOS=windows GOARCH=amd64 go build ./...
total_run_phase_files: 7  # doctor.go, doctor_promotion_test.go, 3 testdata goldens, spec.md frontmatter, progress.md
m1_to_mN_commit_strategy: "2 commits direct to main (Hybrid Trunk Tier S): 31c5b60bd = M1-M4 RED→GREEN code + golden regen + spec.md draft→in-progress; M5 = this evidence commit. Golden testdata regen is an in-scope cascade follow-up (plan.md §B golden mitigation; established UPDATE_GOLDEN convention)."
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-13
sync_commit_sha: PENDING-BACKFILL  # backfilled in a follow-up commit per established placeholder-backfill pattern
sync_status: all-green
ac_pass_count: 4  # matches spec.md §3 SSOT AC-DP-001..004 (grep -c '^### AC-DP-' spec.md → 4)
changelog_entry_position: 1  # first entry under [Unreleased] > ### Added (verified: grep -c 'SPEC-DOCTOR-PROMOTION-001' CHANGELOG.md → 0 pre-emission, 1 post-emission)
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed (merged 3-phase close on this sync commit)"
b12_self_test_a: "grep -c 'SPEC-DOCTOR-PROMOTION-001' CHANGELOG.md (pre-emission) -> 0, halt-not-triggered"
b12_self_test_b: "grep -cE '^### AC-DP-' spec.md -> 4; CHANGELOG entry references 4/4 AC PASS -> match"
b12_self_test_c: "ls internal/cli/doctor.go internal/cli/doctor_promotion_test.go -> both exist, verified before CHANGELOG claim"
```

## §F Phase 0.95 Mode Selection

- Input: tier=S, scope=2 files (doctor.go + doctor_promotion_test.go), domains=1 (Go CLI), concurrency benefit=LOW (coding-heavy)
- Mode evaluation: trivial=no (semantic TDD) | background=no (writes) | agent-team=RETIRED | parallel=no (single-domain coding) | workflow=no (<30 files) | **sub-agent=SELECTED**
- Decision: sub-agent
- Justification: Tier S single-check coding task; sequential manager-develop per Anthropic coding-task parallelism caveat. Implementation Kickoff approved by user (autonomous progression) prior to this log.
