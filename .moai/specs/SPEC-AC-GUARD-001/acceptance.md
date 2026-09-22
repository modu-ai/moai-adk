---
id: SPEC-AC-GUARD-001
title: "AC authoring convention and corpus disposition for worktree-guard-refused acceptance criteria"
version: "0.1.0"
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
priority: P2
phase: "v3.1.0"
module: ".moai/specs,.claude/rules/moai/workflow"
lifecycle: spec-anchored
tier: M
tags: "worktree-guard,acceptance-criteria,authoring-convention,census,measurement-first"
---

# Acceptance — SPEC-AC-GUARD-001

## §D AC Matrix

| AC | Requirement | Description | Severity | Verification |
|----|-------------|-------------|----------|--------------|
| AC-001 | REQ-003 | Verified per-file census over the FULL corpus exists with four-field evidence per row | Must | Direct (E1) |
| AC-002 | REQ-004 | Census governs; proxy retired from normative use | Must | Direct (E1) |
| AC-003 | REQ-006 | Convention boundary map + authoring rule at t287 section, C1/C2 parity modulo internal-trace classes, build green | Must | Direct (E3) |
| AC-004 | REQ-005 | Every rewritten block executes without refusal, semantics preserved | Must | Direct (E2) |
| AC-005 | REQ-004 | "Leave"-dispositioned files carry zero content diff | Must | Direct (E5) |
| AC-006 | REQ-002 | No AC verification relocated into script files | Must | Direct (E4) |
| AC-007 | REQ-001 | Census closure: every predicate-flagged file has a final disposition; zero unresolved rows (corpus-level plain-command property) | Must | Direct (E1) |

Traceability note (plan-audit-iter1 D1): REQ-001 (corpus-wide plain-command property) is covered DIRECTLY by AC-007 (census closure), not indirectly — the closure check is grep-verifiable with guard-safe patterns.

## §D.1 AC-001 — Verified full-corpus census with four-field evidence

**Given** the develop tree at the census-start SHA (pinned in the disposition table)
**When** the verified per-file census runs over the FULL acceptance.md corpus scanned at BLOCK level using the family predicates from spec.md §2.3 plus the named sweeps (corpus-wide `$(git`-text sweep in guard-safe `[g]it` spelling; multi-line open-substitution sweep per spec §2.5.1), with fresh guard samples per family (>= 2 per family)
**Then** a disposition table exists at `.moai/specs/SPEC-AC-GUARD-001/` evidence path (or `.moai/reports/t1067/`) carrying, per file: the evidence command, its verbatim stdout, its exit code as its own field, and the tree SHA; every row's disposition (rewrite / leave / unprobeable) is traceable to a recorded probe; and the table RECEIVES every population the sweeps surface — including files outside the conjunct-selected 55 (iter1-measured: 41 `$(git`-text files outside it, 23 multi-line-substitution files), for which the 55 is a prior snapshot, never a scope filter.

## §D.2 AC-002 — Proxy retired from normative use

**Given** the proxy verdict (spec.md §2.4: 55 is neither an upper nor a lower bound)
**When** any SPEC artifact, evidence record, or rule text produced by this work states a family size
**Then** it cites the verified census count with its pinned tree SHA — the bare proxy figure (54/55) never appears as a census claim or a bound; where a historical figure is quoted for provenance it is labeled "proxy count at <sha>".

## §D.3 AC-003 — Convention home carries the measured boundary map

**Given** the t287 section `## Refused Commands in a Worktree-Isolated Session` at `.claude/rules/moai/workflow/worktree-integration.md`
**When** M2 completes
**Then** (a) the section documents the refused families A and B with probe IDs, the executable shapes P4/P6/P8/P9, the P8 boundary-unknown label, and the "names-git-without-executing" corroboration; (b) the normative authoring rule (plain verbs, separate `echo "exit=$?"`, no git in `$( )`, no write/rm tail, no script-file relocation) is present; (c) the template mirror `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` carries the section identical MODULO INTERNAL-TRACE CLASSES — a GENERIC working example, with zero SPEC-ID literals, audit citations, internal dates, or commit SHAs in the mirror file (template-internal-isolation-doctrine §25.1 class C1 / anti-pattern AP-25.2; the SPEC-ID pointer to AC-AEC-013 lives in the C1 local copy only); (d) `make build` exits 0.

## §D.4 AC-004 — Rewritten blocks are guard-executable and semantics-preserving

**Given** each AC block the M1 census verified to refuse
**When** the M3 rewrite lands
**Then** every rewritten command, executed as a single Bash invocation in a worktree-isolated session, completes WITHOUT the guard refusal; its rewritten form still yields the observable stdout, the exit code as its own field, and the pinned tree SHA the original verification intended; and no rewritten block contains a new `rm -rf` or tree-write tail inside the audited tree.

## §D.5 AC-005 — Leave-disposition is a true no-op

**Given** files the M1 census dispositioned "leave, no rewrite" (expected: the exit-capture-only population, per the P4/P6 shape measurement)
**When** M3 completes
**Then** `git diff <census-base-SHA> -- <those files>` is empty; any change to a leave-file is a defect against REQ-004's census-governs rule.

## §D.6 AC-006 — No script-file relocation

**Given** the full diff of this SPEC's work
**When** reviewed
**Then** zero AC verifications that previously ran inline have been moved into script files or gained a new script-invocation indirection (`.sh` references introduced where an inline command previously ran); where a verification could not be expressed as one compound invocation, it was REDUCED instead.

## §D.7 AC-007 — Census closure (REQ-001, corpus-level)

**Given** the census predicate sweeps executed over the full corpus in guard-safe single-invocation form (e.g. `grep -rlE '\$\( *[g]it' --include=acceptance.md .moai/specs` and the multi-line open-substitution sweep)
**When** the M1 census closes and the M3 rewrites land
**Then** every predicate-flagged file appears in the disposition table with a final disposition (leave-verified / rewrite-executed / rewritten-verified), zero rows remain "unprobeable" or unresolved, and every final disposition is consistent with REQ-001's plain-command convention — the corpus-wide property holds by census closure, mechanically checkable against the sweep file-list (sweep hits ⊆ disposition-table file set).

## §D.8 Edge cases

- A file matching no family predicate yet refusing on probe → disposition "unprobeable"/rewrite, recorded as a NEW family observation with its probe; the boundary map gains the row (the map is measured, not closed).
- The census tooling itself refused mid-run (spec.md §2.6 hazard) → the census must be re-expressed as plain single invocations; the refusal is recorded, never bypassed.
- develop moving under the census → re-run the affected probes and re-pin the tree SHA (verification-completeness.md §4); never re-cite a stale count.
- A leave-file acquires a family-A block from a concurrent merge → out of census scope; noted as residual risk, disposition deferred to the next occurrence.

## §D.9 Quality gates

- Definition of Done: AC-001..AC-007 all PASS with evidence paths still resolving at sync-audit time; E1-E5 self-verification recorded in progress.md §E; no [NEEDS CLARIFICATION] markers remain unresolved at Implementation Kickoff Approval.
