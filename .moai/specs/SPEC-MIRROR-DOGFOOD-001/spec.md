---
id: SPEC-MIRROR-DOGFOOD-001
title: "TestRuleTemplateMirrorDrift/worktree-integration.md RED repair — parity restore + dogfood record relocation"
version: "0.1.0"
status: completed
created: 2026-09-23
updated: 2026-09-23
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/template"
lifecycle: spec-anchored
tags: "template-mirror, neutrality, dogfood-record, rule-drift, tier-m"
tier: M
related_specs: [SPEC-AC-GUARD-001, SPEC-AUDIT-EXPORT-CLAUSE-001, SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001]
---

# SPEC-MIRROR-DOGFOOD-001

## §A Overview

The `TestRuleTemplateMirrorDrift/worktree-integration.md` subtest is RED: the local
managed copy of `worktree-integration.md` diverges from its distributed template mirror
by 2 hunks (22 changed lines) because a card-t1067 dogfood record (SPEC IDs, measurement
dates, a `.moai/reports/` census path) was written directly into the local managed copy
instead of a durable dev-only home. This SPEC encodes the already-settled purpose
determination into a minimal two-file repair: restore byte parity by adopting the
template's neutral form, and relocate the displaced dogfood record into a tracked,
`paths:`-scoped file under `.claude/rules/local/`.

## §B Background and Doctrinal Basis (Purpose Determination)

The determination of which copy may carry what content was settled by measurement
(research.md §6, all observations pinned to HEAD `d323f68fd`, branch `WT-mirror-drift`):

- **Template copy** (`internal/template/templates/.claude/rules/moai/**`): NEUTRAL
  distributed artifact. MUST NOT carry moai-adk internal card records — neutrality
  classes C1-C8 (no SPEC IDs, no REQ tokens, no internal dates, no commit SHAs, no
  `.moai/reports/` paths). Enforced by `TestTemplateNoInternalContentLeak` and the CI
  template-neutrality guard. Current state is correct; ZERO changes.
- **Local managed copy** (`.claude/rules/moai/**`): deployed artifact under a
  byte-parity obligation (`TestRuleTemplateMirrorDrift`). NOT a durable home for
  dogfood records — the managed root is wiped-and-redeployed by every `moai update`
  (`CleanMoaiManagedPaths`, CLAUDE.local.md §2.3).
- **Dogfood records**: durable home is `.claude/rules/local/` (git-tracked, unmanaged,
  no template mirror — established by the 5 existing dev-only protocol files).

Attribution: the divergence-causing string entered ONLY the local copy, via commit
`dbe1a6941` (SPEC-AC-GUARD-001 M2, card t1067); the template side has never carried it.
t1072 (`427ec4455`) and t1073 (`7795226c2`) edited both copies properly (dual-copy) and
are not part of the divergence.

**Purpose-determination record (plan-phase deliverable, not a run-phase requirement):**
the requirement to record this determination and the two rejected directions is
DISCHARGED by this §B together with research.md §6-§7; it was verified by the iter-1
plan audit's Direction Fidelity check and therefore carries no run-phase AC.

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-09-23 | 0.1.0 | Initial plan-phase artifacts authored by manager-spec (card t1086) |
| 2026-09-23 | 0.1.0 | Plan-audit iter-1 FAIL 0.72 annotation cycle — REQ reclassification/renumber (REQ-MD-001..005), tier S→M, staged/commit-scoped diff check, RED four-element cell, record-file `paths:` scope, census disposition, one-clause REQ reflow |

## §C Requirements (GEARS)

- **REQ-MD-001** (Ubiquitous): The local managed copy of `worktree-integration.md` shall be byte-identical to its template mirror, adopting the template's existing neutral form verbatim.
  - Local path: `.claude/rules/moai/workflow/worktree-integration.md`; mirror path:
    `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`.
    No re-authoring of either side's content.

- **REQ-MD-002** (Ubiquitous): The template copy `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` shall carry ZERO changes in this SPEC's changeset.

- **REQ-MD-003** (Ubiquitous): The mirror test file `internal/template/rule_template_mirror_test.go` shall carry ZERO changes in this SPEC's changeset.

- **REQ-MD-004** (Ubiquitous): The t1067 dogfood record shall be relocated into the pinned NEW git-tracked file `.claude/rules/local/wt-ac-restatement-record.md`.
  - The record carries, in substance: the AC-AEC-013 restatement pointer
    (`SPEC-AUDIT-EXPORT-CLAUSE-001/acceptance.md:622-680`), the corpus-census path
    (`.moai/reports/t1067/census-20260922.md`) WITH its non-live disposition (research.md
    §2 — the census artifact was never committed; the path is historical), the
    card-t1067 attribution, and the 2026-09-22 measurement dates.
  - The record file carries a `paths:` frontmatter scope keyed to the rule file it
    annotates (`paths: "**/.claude/rules/moai/workflow/worktree-integration.md"` — the
    rule-authoring.md glob convention) so it loads
    conditionally and never joins the always-loaded surface (rule-authoring.md slot 1).

- **REQ-MD-005** (Event-driven): When `TestTemplateNoInternalContentLeak` runs after the repair, the template copy shall remain free of internal-content leaks (neutrality classes C1-C8), evidenced by the test exiting 0.

## §D Scope

- Files changed (exactly 2):
  1. `.claude/rules/moai/workflow/worktree-integration.md` — restored to template bytes.
  2. `.claude/rules/local/wt-ac-restatement-record.md` — NEW tracked file (pinned
     filename, `paths:`-scoped frontmatter) carrying the relocated t1067 record.
- Files explicitly untouched: everything under `internal/template/templates/**` and
  `internal/template/rule_template_mirror_test.go`.
- Verification surface: `go test ./internal/template/ -run
  'TestRuleTemplateMirrorDrift|TestTemplateNoInternalContentLeak'`, `cmp`, staged-diff
  and commit-scoped name-only assertions (acceptance.md §D.2), `make build`.

## §E Out of Scope

### Out of Scope — Template copy modifications

- Any edit to `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` (the template's neutral form is the parity target, taken verbatim).
- Any edit to `internal/template/rule_template_mirror_test.go` (no gate code changes).
- Any edit to any other template mirror file (`frontend.md`, `hooks-system.md` etc. all PASS).

### Out of Scope — Gate relaxation

- Excluding or allowlisting `worktree-integration.md` out of `TestRuleTemplateMirrorDrift` byte parity (the gate's non-vacuity is evidenced by the current RED; weakening it would hide future real drift).
- Any `lint.skip` or skip-flag route around the failing subtest.

### Out of Scope — Rejected repair direction

- `cp` local -> template (ships `card t1067`, `2026-09-22`, SPEC IDs, and `.moai/reports/` paths into the distributed template — violates §25 neutrality C1-C8 and the CI neutrality guard).

### Out of Scope — Broader record hygiene

- Migrating other historical dogfood records that may sit inside `.claude/rules/moai/` — only the t1067 record displaced by this divergence is in scope.
- Re-authoring or summarizing the teaching content of either hunk — the template's neutral restatement is adopted as-is.
- Creating any template counterpart for the record file — it is mirror-free by design (research.md §5).

## §F Constraints

- [HARD] Single-commit changeset: both changed files land in the SAME commit, staged by
  explicit pathspec. The record file is mirror-free by design — no template counterpart
  exists or may be created for it.
- [HARD] `make build` must exit 0 after the changes (template-embed build gate; this
  SPEC changes no template source, so the build must remain green — a red build here
  means something outside scope was touched).
- [HARD] No commit is authored during plan phase; the run phase owns the repair commit.
- [HARD] When any acceptance verification of this SPEC fails (mirror subtest RED, `cmp`
  mismatch, neutrality leak, or `make build` failure), the run phase shall stop and
  report the failing command output verbatim before any repair retry — never silently
  re-attempt or relax a gate. (AC-MD-006 anchors to this constraint.)
- Documentation-only change: no Go source behavior changes, no new @MX annotations
  warranted (Phase 14 MX planning finding: **None** — no exported functions, no
  goroutines, no complexity-bearing code touched).

## §G Success Criteria

See `acceptance.md` §D AC matrix (AC-MD-001..AC-MD-006). Summary: mirror subtest green
(8/9 -> 9/9), template tree untouched, byte identity via `cmp`, relocated record
git-tracked with all four record elements plus `paths:` scope, neutrality test green,
`make build` green.

## §H Cross-References

- research.md — first-hand measurements (four-element RED cell, diff hunks, attribution probe, neutrality PASS, census disposition)
- `evidence/red-baseline-d323f68fd.txt` — full raw RED output, exit code 1, pinned to `d323f68fd`
- SPEC-AC-GUARD-001 — the SPEC whose M2 commit (`dbe1a6941`) introduced the record into the local copy
- SPEC-AUDIT-EXPORT-CLAUSE-001 — owner of the referenced AC-AEC-013 restatement (`acceptance.md:622-680`)
- SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001 — the byte-parity gate being honored (not weakened)
- CLAUDE.local.md §2 / §2.3 — template-first rule and the managed-root wipe (`CleanMoaiManagedPaths`)
- `.moai/docs/template-internal-isolation-doctrine.md` §25.1 — C1-C8 content-class catalogue
- `.claude/rules/moai/development/rule-authoring.md` — the always-loaded-slot duty the record file's `paths:` scope satisfies
