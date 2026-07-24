# SPEC-SEC-DEEPSCAN-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-07-24
tier: L
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (5 — Tier L complete)
epic: SECURITY-ABSORB (SPEC-1 of the cohort)
surface_decision: /moai review --deep (new /moai security REJECTED — retired by SPEC-SUBCOMMAND-RETIRE-001)
spec_id_precheck: PASS (regex ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ — decomposition SPEC | SEC | DEEPSCAN | 001)
open_clarifications: 0 (both resolved via AskUserQuestion — single-commit=`--commit <SHA>`; retention=no-auto-prune/defer-SPEC-2)
audit_fix: iter-1 plan-audit FAIL 0.84 (MP-7 firewall) → D1-D5 applied (markers deleted, decisions settled, AC count 27, retire status completed, REQ-SDS-052 rung-scoped)
next: plan-audit re-run (plan-auditor) → Implementation Kickoff Approval → run-phase (manager-develop)

## §E.2 Run-phase Evidence

Deliverable: `--deep` mode playbook added to the `review` workflow skill (both trees, byte-identical) + command `argument-hint` gains `[--deep] [--patch] [--commit <SHA>]` (`.tmpl` + rendered local). No Go code, no shipped `.js`, no new agent. Verification is the acceptance.md grep/diff/file-existence matrix (the SPEC-ID / date tokens are intentionally ABSENT from template content per §15/§25 neutrality — the playbook cites requirements by behavior, not by REQ/AC ID).

Files changed (run-phase):
- `internal/template/templates/.claude/skills/moai/workflows/review.md` (Supported Flags + new `## --deep Mode` section)
- `.claude/skills/moai/workflows/review.md` (byte-identical local sibling)
- `internal/template/templates/.claude/commands/moai/review.md.tmpl` (argument-hint)
- `.claude/commands/moai/review.md` (rendered argument-hint)

### AC PASS/FAIL matrix (27 ACs)

| AC | Sev | Status | Verification command (evidence) | Actual output |
|----|-----|--------|---------------------------------|---------------|
| AC-SDS-001 | MUST | PASS | `grep -- '--deep' review.md.tmpl && grep -iE '^#+ .*--deep' review.md` | both match |
| AC-SDS-002 | MUST | PASS | `grep -rIn 'moai security\|security subcommand' .claude/skills/moai/ \| grep -v workflows/review.md` | empty (retirement prose in review.md only; no revival) |
| AC-SDS-003 | SHOULD | PASS | `grep -- '--repo' '--commit' '--patch' '--staged\|--branch'` (job menu) | all match |
| AC-SDS-004 | SHOULD | PASS | `grep -i 'off by default'` | match |
| AC-SDS-010 | MUST | PASS | six-phase ordered list line numbers 319<320<321<322<323<324 | ascending |
| AC-SDS-011 | SHOULD | PASS | `grep 'Skill("moai-ref-{owasp-checklist,llm-security,secops,supply-chain}")'` | all 4 match |
| AC-SDS-012 | SHOULD | PASS | `grep -i 'read-only'` | match (phases 1-4 marked READ-ONLY) |
| AC-SDS-020 | MUST | PASS | `grep -iE 'REACHABILITY\|IMPACT\|DEFENSES' && grep -iE '2[- ]of[- ]3\|quorum'` | match |
| AC-SDS-021 | MUST | PASS | `grep -iE 'non-unanimous.*medium\|cap.*confidence.*medium'` | match |
| AC-SDS-022 | SHOULD | PASS | `grep -iE 'voter independence\|refute-skewed'` | match |
| AC-SDS-023 | SHOULD | PASS | `grep -iE 'unconfirmed candidates\|excluded from the confirmed'` | match |
| AC-SDS-030 | MUST | PASS | `grep -iE 'isolation.*worktree\|scratch clone'` | match |
| AC-SDS-031 | MUST | PASS | `grep -iE 'independent reviewer\|vouch'` | match |
| AC-SDS-032 | SHOULD | PASS | `grep -i 'instead of a patch'` | match |
| AC-SDS-033 | MUST | PASS | `grep -iE 'never.*(apply\|applied)\|git apply' && grep 'one finding = one patch = one PR'` | match |
| AC-SDS-040 | MUST | PASS | `grep -E '\.moai/reports/security-deepscan'` | match (not under .moai/specs) |
| AC-SDS-041 | SHOULD | PASS | `grep F<i>/F1 + severity/confidence/impact/exploit/recommendation` | all match |
| AC-SDS-042 | SHOULD | PASS | `grep -iE 'jsonl\|one finding per line'` | match |
| AC-SDS-043 | SHOULD | PASS | `grep -iE 'revision stamp\|scanned commit\|working tree'` | match |
| AC-SDS-044 | MUST | PASS | `grep -E '\.gitignore'` | match |
| AC-SDS-050 | SHOULD | PASS | `grep -iE 'v2\.1\.154\|dynamic workflows.*(require\|unavailable)'` | match |
| AC-SDS-051 | SHOULD | PASS | `grep -iE 'degrad\|fallback' && grep PRIMARY && grep DEGRADED` | 3-rung ladder |
| AC-SDS-052 | SHOULD | PASS | `grep 'security-review' && grep -iE 'quorum preserved\|preserve.*quorum'` | match |
| AC-SDS-060 | MUST | PASS | `diff -q review.md(local) review.md(template)` | identical (byte-parity) |
| AC-SDS-061 | MUST | PASS | `grep -E 'SPEC-SEC-DEEPSCAN\|2026-07-24' template-files` + `go test -count=1 ./internal/template/` (leak) | 0 matches; leak test PASS |
| AC-SDS-062 | MUST | PASS | `go build ./...` exit 0 + `go test -count=1 ./internal/template/...` PASS + `find templates -name '*.js'` empty + catalog.yaml unchanged | no new agent / no Go change / no .js |
| AC-SDS-063 | SHOULD | PASS | positive `agent...prompt user` grep minus negations | empty (agents return blocker reports; "never prompt the user") |

Result: 27/27 PASS (13 MUST-PASS + 14 SHOULD-PASS). Evidence log: `.moai/state/verify/deepscan-001/ac_matrix.log` (gitignored runtime state).

### Indirect / runtime-smoke residual (§D.7)

The §D.7 indirect check (a live `/moai review --deep --repo` smoke invocation producing a results dir) is NOT executed by this run-phase implementer: the deep scan is runtime-constructed (no shipped script) and a live multi-agent `Workflow()` run requires the orchestrator + a Dynamic-Workflows-capable runtime, which is outside a run-phase sub-agent's capability. Per §D.7(a) the primary indirect verification is the playbook-completeness ACs above (all PASS). The live smoke run is deferred to orchestrator-driven / interactive verification (residual risk, non-gating — §D.7 is Indirect/Forward-Looking, not in the 27-AC gate).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-24
run_commit_sha: 93d91e2e7
run_status: audit-ready
ac_pass_count: 27
ac_fail_count: 0
ac_must_pass_count: 13
ac_should_pass_count: 14
preserve_list_post_run_count: 0        # PRESERVE list (plan §A.3) untouched; existing --staged/--branch/--security/--file/--design/--critique/--lean/--repo behavior unchanged
l44_pre_commit_fetch: not-run          # feature branch off release/v3.0.2-prep; no origin/main merge in run-phase (Tier L PR created later by manager-git)
l44_post_push_fetch: not-run           # push deferred to PR step per Tier L Route B
new_warnings_or_lints_introduced: 0    # go build ./... exit 0; go test -count=1 ./internal/template/... PASS; spec lint exit 0 (no SEC-DEEPSCAN findings)
cross_platform_build:
  host_darwin: pass                    # go build ./... exit 0
  windows_amd64: pass                  # GOOS=windows GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 4               # 2 workflow-skill copies (byte-identical) + 2 command wrappers (.tmpl + rendered)
byte_parity_workflow_skill: identical  # diff local<->template review.md clean
catalog_yaml_changed: false            # workflow sub-file edit does not alter the moai-skill SKILL.md hash; no new agent/skill
shipped_js_scripts: 0                  # find templates -name '*.js' empty
m1_to_mN_commit_strategy: single-run-commit  # markdown-first SPEC; M1-M6 delivered in one run commit; plan artifacts in a preceding commit
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-24
sync_status: audit-ready
sync_commit_sha: <pending-backfill-next-commit>
changelog_entry_position: "[Unreleased] > ### Added (Epic SECURITY-ABSORB, joint entry with SPEC-SEC-GUARDIAN-001)"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed"
readme_update: "one-line --deep mention added to the /moai Slash Commands review row"
docs_site_4locale: "DEFERRED — follow-up required (not performed in this sync per scope note)"
b12_self_test_a: "grep -c 'SPEC-SEC-DEEPSCAN-001\|SPEC-SEC-GUARDIAN-001' CHANGELOG.md -> 0 before emission (no duplicate)"
b12_self_test_b: "acceptance.md AC row count 27 == CHANGELOG claim '27/27 AC'"
b12_self_test_c: "ls .moai/specs/SPEC-SEC-DEEPSCAN-001/spec.md .moai/specs/SPEC-SEC-GUARDIAN-001/spec.md CHANGELOG.md README.md -> all exist"
```
