# Progress — SPEC-PROJECT-NAVIGATOR-001

> Plan-phase skeleton. §E.2–§E.4 are placeholder headings only — populated by manager-develop (§E.2/§E.3) and manager-docs (§E.4) per the status-transition ownership matrix.

## §E.1 Plan-phase Audit-Ready Signal

_Pending plan-auditor run on this plan-phase artifact set (spec.md + plan.md + acceptance.md + research.md + progress.md)._

## §E.2 Run-phase Evidence

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-PN-001 | PASS | `go test -run TestACPN001 ./internal/template/` | `--- PASS: TestACPN001_ThreeFilesNoExtras` |
| AC-PN-002 | PASS | `go test -run TestACPN002 ./internal/template/` | `--- PASS: TestACPN002_EveryRowCarriesProvenance` |
| AC-PN-005 | PASS | `go test -run TestACPN005 ./internal/template/` | `--- PASS: TestACPN005_IdempotentRegeneration` |
| AC-PN-006 | PASS | `go test -run TestACPN006 ./internal/template/` | `--- PASS: TestACPN006_EmptyProjectResilience` |
| AC-PN-007 | PASS | `go test -run TestACPN007 ./internal/template/` | `--- PASS: TestACPN007_MalformedFrontmatterTolerance` |
| AC-PN-008 | PASS | `go test -run TestACPN008 ./internal/template/` | `--- PASS: TestACPN008_AtomicRenameDeterministic` |
| AC-PN-013 | PASS | `go test -run TestACPN013 ./internal/template/` | `--- PASS: TestACPN013_NonDuplication` |
| AC-PN-015 | PASS | `go test -run TestACPN015 ./internal/template/` | `--- PASS: TestACPN015_NonGoProject` |
| AC-PN-016 | PASS | `go test -run TestACPN016 ./internal/template/` | `--- PASS: TestACPN016_LSELBoundary` |

M1 deliverable: `navigator-regen.sh` regeneration script + `navigator_regen_test.go` (9 AC tests, all GREEN). RED captured first (script absent → test FAILED at `navigator-regen.sh not found`), then GREEN after script landed.

| AC-PN-003 | PASS | wiring: SKILL.md Navigator phase + references/navigator.md §3 chains regen into /moai sync before sync-commit | skill-body instruction |
| AC-PN-004 | PASS | wiring: SKILL.md Navigator phase invokes script alongside product/structure/tech.md | skill-body instruction |
| AC-PN-009 | PASS | `go test -run TestACPN009 ./internal/template/` | `--- PASS: TestACPN009_AmbientBriefBounded` |
| AC-PN-010 | PASS | `go test -run TestACPN010 ./internal/template/` | `--- PASS: TestACPN010_FailOpen*` (2 cases) |
| AC-PN-011 | PASS-WITH-DEBT | wiring in SKILL.md --brief mode + references/navigator.md §5; full-brief load is a skill-mode (agent-driven), not mechanically testable without a session | skill-body instruction |
| AC-PN-012 | PASS | `go test -run TestACPN012 ./internal/template/` | `--- PASS: TestACPN012_Staleness*` (advisory + override) |
| AC-PN-014 | PASS | `go test -run TestTemplateNoInternalContentLeak ./internal/template/` | `ok` — C1 sentinel extended with SPEC-PROJECT-NAVIGATOR-; REQ-PN/AC-PN tokens stripped from distributed files |
| AC-PN-017 | PASS | plan.md Phase 1 consultation section (5 lines, LOC-neutral) + references/navigator.md §6 | workflow skill wiring |
| AC-PN-018 | PASS-WITH-DEBT | run.md Quick Reference carries a one-line consultation pointer (LOC ceiling 200 constrained the wiring to a single line); full procedure in references/navigator.md §6 | workflow skill wiring |

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-05
sync_commit_sha: pending-backfill-sync
sync_status: completed
changelog_entry_position: CHANGELOG.md [Unreleased] / Added (top of section)
frontmatter_status_transitions:
  spec.md: in-progress -> implemented -> completed
  plan.md: n/a (no frontmatter)
  acceptance.md: n/a (no frontmatter)
  progress.md: n/a (no frontmatter)
canary_compliance_check:
  ci_guard: pending (PR-stage; repo-local PR-mandatory per §23)
b12_self_test:
  a: pre-emission grep -c 'SPEC-PROJECT-NAVIGATOR-001' CHANGELOG.md = 0 (no duplicate)
  b: AC count in acceptance.md = 18; CHANGELOG entry references "18 ACs (AC-PN-001..018)" — match
  c: navigator-regen.sh / SKILL.md / handle-session-start-navigator.sh / references/navigator.md paths verified via ls
```

**Docs-site 4-locale follow-up (deferred — does NOT block this code PR)**: the Navigator is a real user-facing feature and a docs-site guide page on `adk.mo.ai.kr` (ko/en/ja/zh) WOULD be valuable eventually. However authoring it is HEAVY (4-locale × Hugo page × the oss-docs harness structure-curator + content-author + locale-translator + verify recipe) and is out of scope for this sync commit. Filed as a follow-up TODO — a separate SPEC (or a `/moai:harness oss-docs` invocation) should pick this up post-merge. Recorded here (not in the CHANGELOG body) so the next maintainer sees it when reading the SPEC close.

