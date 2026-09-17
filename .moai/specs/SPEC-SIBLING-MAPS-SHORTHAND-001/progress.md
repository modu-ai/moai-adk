# Progress — SPEC-SIBLING-MAPS-SHORTHAND-001

Card t801. Tier M. Run-phase base: `881aa4bb8` (branch `WT-spec-lint-coverage`).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-18
tier: M
artifacts: spec.md, plan.md, acceptance.md (+ this progress.md)
requirements: 7 (REQ-SMS-001..007) — Tier M ceiling 16
acceptance_criteria: 12 (AC-SMS-001..009, AC-SMS-GATE-001..003) — Tier M ceiling 16
scope_note: narrowed from SPEC-SPEC-LINT-COVERAGE-COLLECT-001 (plan-audit FAIL 0.74) per the lead's split; the heading-collection axis moved to card t894.

## §E.2 Run-phase Evidence

Evidence report, exported to the PRIMARY checkout so the citation still resolves
once this worktree is gone (`.moai/reports/*` is gitignored, so the in-worktree
copy reaches no clone):
`/Users/goos/MoAI/moai-adk-go/.moai/reports/t801/m2-evidence.md`
(sha256 `b318d27ec77bb98d15b5261de980de1894b21eeb2f0201f978d15b440b09de5c`).
It carries every AC with its command, its verbatim output, and the tree the
measurement was taken against. The whole `.moai/reports/t801/` directory —
fixtures, before/after corpus artifacts, the three plan-audit iterations — is
exported alongside it.

Measuring instrument: a build made FROM this tree and invoked BY PATH
(`go build -o <scratch>/moai ./cmd/moai`, exit 0). Measured at HEAD `4a6034cca`,
tree `e99e98ead280775da40de22597ed95c8642e5b80`. Run-phase base `881aa4bb8`.

| AC | Verdict | Deciding measurement |
|---|---|---|
| AC-SMS-001 | PASS | fixture A: 0 `CoverageIncomplete`, 1 `MissingExclusions`, exit 0 |
| AC-SMS-002 | PASS | fixture E: exactly 2 lines, `REQ-FIXE-001` + `REQ-FIXE-002` |
| AC-SMS-003 | PASS | fixture F: exactly 2 lines, `REQ-FIXF-003` + `REQ-FIXF-004` |
| AC-SMS-004 | PASS | fixture G: exactly 1 line, `REQ-FIXG-002` |
| AC-SMS-005 | PASS | both greps print `1` (regression control, green before and after) |
| AC-SMS-006 | PASS | `git diff 881aa4bb8..HEAD -- internal/spec/ears.go` empty, exit 0 |
| AC-SMS-007 | PASS | fixture B before == after (`0 error(s), 1 warning(s)`) |
| AC-SMS-008 | PASS | mutant built exit 0; false `REQ-FIXA-002` line reappears verbatim |
| AC-SMS-009 | PASS | corpus 2018 → 2018; counts file and (file, REQ) id list both diff-empty |
| AC-SMS-010 | PASS | `6` / `3` / `1` — call-site named on both paths, shared-rule test PASS |
| AC-SMS-GATE-001 | PASS | `go test ./internal/spec/...` → `ok … 121.219s` |
| AC-SMS-GATE-002 | PASS | `go vet ./internal/spec/...` silent |
| AC-SMS-GATE-003 | PASS | `golangci-lint run internal/spec/...` → `0 issues.` |

Gaps and residual risk are named in the evidence report's own sections; the
load-bearing ones: the full suite and the platform matrix are CI's verdict, not
this lane's, and a dead duplicate expander coexisting with a genuine call to the
shared helper is unreachable by AC-SMS-005/010 and rests on review of the M1 diff.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-09-18
milestones: M1 (locator + expander + tests + fixtures E/F/G), M2 (evidence)
acceptance_criteria: 13 PASS / 0 FAIL / 0 SKIP
evidence_path: /Users/goos/MoAI/moai-adk-go/.moai/reports/t801/m2-evidence.md (primary checkout)
measuring_build_tree: e99e98ead280775da40de22597ed95c8642e5b80
run_base: 881aa4bb8

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-18
sync_commit_sha: pending-backfill-sync   # a commit cannot cite its own hash; backfilled in a following commit
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-SIBLING-MAPS-SHORTHAND-001' CHANGELOG.md → 0 before append (no duplicate entry)"
b12_self_test_b: "AC ids in acceptance.md: 15 distinct tokens, of which 13 are live criteria (AC-SMS-001..010 + AC-SMS-GATE-001..003); AC-FIXE-001 / AC-FIXF-001 are fixture CONTENT quoted inside AC-SMS-002 / AC-SMS-003, not criteria. CHANGELOG entry cites 13 — matches progress.md §E.3 (13 PASS / 0 FAIL / 0 SKIP)."
b12_self_test_c: "every file path in the CHANGELOG entry verified present via ls internal/spec/lint_coverage_sibling*.go (lint_coverage_sibling_maps.go, lint_coverage_sibling_maps_test.go, lint_coverage_sibling.go, lint_coverage_sibling_table.go) and internal/spec/ears.go"
changelog_entry_position: "CHANGELOG.md [Unreleased] → ### Fixed, first bullet"
frontmatter_status_transitions:
  spec.md: "in-progress → completed (status only; updated already read 2026-09-18, the sync date)"
  plan.md: "no YAML frontmatter — nothing to transition"
  acceptance.md: "no YAML frontmatter — nothing to transition"
  progress.md: "no YAML frontmatter — nothing to transition"
canary_compliance_check: "n/a — this SPEC defines no forward-looking policy that its own sync tests"
docs_surface: "none — internal lint-behaviour repair with zero user-visible surface change (corpus delta 0); README*.md and docs-site/** deliberately untouched"
```
