---
id: SPEC-MIRROR-DOGFOOD-001
title: "Acceptance — mirror parity restore + dogfood record relocation"
created: 2026-09-23
card: t1086
---

# Acceptance — SPEC-MIRROR-DOGFOOD-001

## §D AC Matrix

Two-cell discipline (verification-completeness.md §2): AC-MD-001's RED-now cell was
observed on this tree at HEAD `d323f68fd` (branch `WT-mirror-drift`, clean) — full raw
output + exit code at `evidence/red-baseline-d323f68fd.txt`, four-element form in
research.md §1. No criterion below is vacuous-by-construction; each names the milestone
that flips it (all flip at M1).

| AC | Requirement | Criterion | RED-now (pre-M1) | Flips at |
|----|-------------|-----------|------------------|----------|
| AC-MD-001 | REQ-MD-001 | Mirror subtest green | OBSERVED RED, exit 1 (research.md §1; `d323f68fd`) | M1 |
| AC-MD-002 | REQ-MD-002, REQ-MD-003 | Template tree + mirror test untouched (staged + commit-scoped) | N/A (absence assertion, pinned to `d323f68fd`) | M1 |
| AC-MD-003 | REQ-MD-001 | Byte identity via `cmp` | OBSERVED FAIL-equivalent (61473 vs 61638 B) | M1 |
| AC-MD-004 | REQ-MD-004 | Record relocated, tracked, complete, `paths:`-scoped | ABSENT today (record sits in managed root) | M1 |
| AC-MD-005 | REQ-MD-005 | Neutrality leak test green | PASS today (must stay PASS — preserve-surface) | M1 |
| AC-MD-006 | spec.md §F constraint 4 | `make build` green; failure-handling duty honored | PASS today (must stay PASS — preserve-surface) | M1 |

REQ-MD-001..005 all carry ACs. REQ coverage note: the purpose-determination duty is a
plan-phase deliverable discharged in spec.md §B (no run-phase AC by design); AC-MD-006
anchors to the spec.md §F failure-handling/build constraint it verifies.

## §D.1 AC-MD-001 — Mirror subtest green

- **Given** the worktree at base `d323f68fd` where
  `TestRuleTemplateMirrorDrift/worktree-integration.md` fails with sentinel
  `RULE_TEMPLATE_MIRROR_DRIFT` at exit code 1 (RED-now cell: command
  `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift' -count=1 -v`, full
  raw output attached at `evidence/red-baseline-d323f68fd.txt`)
- **When** `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift' -count=1`
  runs after M1
- **Then** the command exits 0 and all 9 subtests report PASS, including
  `worktree-integration.md`, `frontend.md`, and `hooks-system.md`.

## §D.2 AC-MD-002 — Template tree untouched (scope check)

- **Given** the base commit `d323f68fd` and both in-scope file changes staged by
  explicit pathspec
- **When** `git diff --cached --name-only` runs pre-commit, and after the M1 commit
  `git show --name-only --format= <M1-commit-SHA>` runs
- **Then** BOTH outputs contain EXACTLY the 2 in-scope paths — no path under
  `internal/template/templates/` and no change to
  `internal/template/rule_template_mirror_test.go`. (The commit-scoped second
  assertion is the one that survives the SPEC artifact directory also landing on the
  branch; the range-diff form against `d323f68fd` is deliberately NOT used — its
  expected total is ordering-dependent and false under every ordering.)

## §D.3 AC-MD-003 — Byte identity

- **Given** the two copies of `worktree-integration.md`
- **When** `cmp .claude/rules/moai/workflow/worktree-integration.md
  internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` runs
- **Then** it exits 0 with no output (byte-identical; parity restored in the
  template -> local direction only).

## §D.4 AC-MD-004 — Relocated record

- **Given** the pinned new file `.claude/rules/local/wt-ac-restatement-record.md`
- **When** `git ls-files .claude/rules/local/` runs and the file content is inspected
- **Then** the file is git-tracked and carries all four record elements: the
  `SPEC-AUDIT-EXPORT-CLAUSE-001/acceptance.md:622-680` pointer, the
  `.moai/reports/t1067/census-20260922.md` census path together with its historical
  disposition line (never committed; not live evidence), the card-t1067 attribution,
  and the 2026-09-22 measurement dates — and its YAML frontmatter carries
  `paths:` with the exact glob value
  `"**/.claude/rules/moai/workflow/worktree-integration.md"`.

## §D.5 AC-MD-005 — Neutrality preserved

- **Given** the template copy carries zero changes (AC-MD-002)
- **When** `go test ./internal/template/ -run TestTemplateNoInternalContentLeak
  -count=1` runs after M1
- **Then** it exits 0 (no internal-content leak in the distributed template).

## §D.6 AC-MD-006 — Build gate + failure handling

- **Given** the changeset touches no template source
- **When** `make build` runs after M1
- **Then** it exits 0. If any acceptance verification fails instead, the run phase
  stops and reports the failing command output verbatim before any retry (spec.md §F
  constraint 4 — the constraint this AC anchors).

## §D.7 Severity and traceability

- AC-MD-001, AC-MD-003, AC-MD-004: release-blocking (they ARE the repair).
- AC-MD-002, AC-MD-005: release-blocking preserve-surface (a regression here means the
  repair went in the rejected direction).
- AC-MD-006: release-blocking gate hygiene.
- Edge cases: partial `cp` (interrupted write) is caught by AC-MD-003's `cmp`;
  a record file missing one element or the `paths:` scope is caught by AC-MD-004's
  inspection; a stray template edit is caught by AC-MD-002's staged + commit-scoped
  assertions.

## Definition of Done

1. All six ACs verified with verbatim command output recorded in progress.md §E.2.
2. Both files committed together in one commit, staged by explicit pathspec.
3. `git show --name-only --format= <M1-commit-SHA>` shows exactly the 2 in-scope paths
   (nothing outside them).
4. MX tag finding recorded: None (documentation-only change).
