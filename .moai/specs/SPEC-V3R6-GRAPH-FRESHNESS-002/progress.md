# Progress — SPEC-V3R6-GRAPH-FRESHNESS-002

## §E.1 Plan-phase Audit-Ready Signal

- Artifact set: Tier M 3-file set (spec.md, plan.md, acceptance.md) + research.md (explicit brief addition — provenance/triage/close-path analysis; deviation recorded in research.md §6.1) + progress.md.
- Requirements: 14 REQ (GEARS, pattern-annotated — incl. cross-cutting REQ-GFR-014 final-stamp main-reachability); 16 AC (Given-When-Then) each carrying a RED-now baseline pinned at `c9eed8ac6`; REQ↔AC traceability 100% both directions (acceptance.md §D.2). AC count (16) sits at the Tier M ceiling exactly; REQ count (14) is below its 16 ceiling.
- Plan-audit iter-1: PASS 0.856 (Tier M threshold 0.80), 0 BLOCKING — report at `.moai/reports/plan-audit/SPEC-V3R6-GRAPH-FRESHNESS-002-review-1.md`. D1-D5 textual defects corrected post-audit (D1 premise code-verified: `needsSHABackfill` closer.go:397-405 matches only empty/`(this commit)`/`(pending)`/`<pending>` — the predecessor's `pending-backfill — …` form is unrecognized, so the manual pre-close backfill prevents placeholder-freezing); D6/D7 recorded-not-fixed per auditor discretion.
- Frontmatter: 12 canonical fields + `era: V3R6` (deliberate H-override — plan-phase skeleton lacks sync_commit_sha, which would heuristically classify V3R5) + `tier: M` + `related_specs: [SPEC-V3R6-GRAPH-FRESHNESS-001]`; SPEC ID regex-validated (Bash verbatim `PASS`); ID uniqueness verified against both this worktree's and the primary checkout's `.moai/specs/` catalogs.
- Scope: 29 adopted CR round-2 findings (triage A-1..A-4) + predecessor correction-and-close; 5 follow-ups (F1-F5) / 2 rejections / 1 deferral / 33 already-fixed all fenced out in spec.md §F with citations.
- M0: codemaps provenance restamp `52f7ba135` against main-reachable `c9eed8ac6` executed pre-issuance (squash-merge orphan repair; triage-table.md §F5) — recorded as an already-executed precondition (plan.md §F M0, spec.md §B.1/§C); forward obligation encoded as REQ-GFR-014/AC-GFR-016 (no branch-HEAD restamp; final stamp main-reachable).
- Split decision: new SPEC, not an increment — predecessor is `implemented` with a frozen audit trail; M4 owns its amendment (1.1.0 → 1.2.0) and close as THIS SPEC's scope. No `depends_on` (predecessor `implemented` ≠ `completed`; circular — research.md §6.3).
- Evidence base: `.moai/reports/t279/triage-table.md` + `verify-{graph,cli,docs}.md` (read, not re-derived); `moai spec close --help` + the close implementation code-verified at `internal/spec/closer.go` this run (research.md §4 — §E.5 authoring dropped, manual SHA backfill ordered before the close, pass-with-debt non-blocking).

## §F Phase 4 Mode Selection

Logged by the orchestrator before the first run-phase spawn (2026-08-26, after Implementation Kickoff Approval: approve + autonomous progression, goal armed via `moai goal arm` under session 13b75f36 — mechanical condition: predecessor completed + sync_commit_sha 2fc4b40a6 + this SPEC implemented + targeted package tests green).

Input parameters: tier M · scope ~22 files (11 test/code + astx + docs + SPEC artifacts) · domains: Go tests, Go production policy, astx query + build tags, docs/CHANGELOG/docs-site, SPEC lifecycle (5) · language mix Go+markdown · concurrency benefit LOW (coding-heavy).

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | multi-file, semantic changes across 5 domains |
| serial | **yes** | coding-heavy Tier M; M1→M2→M3 dependency chain (M2 touches files M1 tests; M4 depends on all); Anthropic coding-task parallelism caveat |
| fanout | no | not research-heavy; write-capable parallelism forbidden |
| sweep | no | < 30 files; semantic, multi-rule; not mechanical-uniform |
| agent-team | no | not operator-requested |

Decision: serial — one manager-develop delegation per milestone (M1, M2, M3), M4 split into a manager-spec re-delegation (predecessor body corrections + manual backfill) followed by the close CLI. Boundary case: none (5 domains would tempt fanout, but every domain is write-capable coding work → serial per the tie-breaker).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
