---
id: SPEC-V3R6-GRAPH-FRESHNESS-002
title: "Graph freshness remediation: t250 CR round-2 adopted findings, test/code policy cleanup, and predecessor SPEC correction-and-close"
version: "1.0.0"
status: draft
created: 2026-08-26
updated: 2026-08-26
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/graph"
lifecycle: spec-anchored
tags: "graph, freshness, remediation, cr-followup, test-policy, spec-close"
era: V3R6
tier: M
related_specs: [SPEC-V3R6-GRAPH-FRESHNESS-001]
---

## §A. Problem Statement

Card t250 (#1648, squash `6786c3fa4`, SPEC-V3R6-GRAPH-FRESHNESS-001) landed green, and its CodeRabbit round-2 review produced **69 findings**. A full-sweep triage against the current tree (`c9eed8ac6`, branch `WT-t250-followup` — `.moai/reports/t279/triage-table.md`, backed by `verify-graph.md` / `verify-cli.md` / `verify-docs.md`) reconciled all 69:

| Verdict | Count | Disposition |
|---|---|---|
| Adopted (this SPEC executes) | 29 | Sections A-1 (11) · A-2 (10 + absorbed Minor-2b) · A-3 (3) · A-4 (5) |
| Follow-up candidates (separate cards) | 5 | F1-F5 — out of scope, recorded in §F |
| Rejected (invalid premise) | 2 | R1/R2 — refuted with evidence, no action |
| Deferred-by-design (recorded deviation) | 1 | D1 — stands as documented |
| Already fixed (no action) | 33 | verify reports cite the fixing file:line |

The 29 adopted findings are real, current-tree-verified defects and debt: untested branches with observed zero-coverage baselines, test fixtures that break on Windows paths, a test that can panic instead of fail, contract doc-comments that contradict behavior, a stale CHANGELOG count, and three SPEC-body statements in the predecessor that the shipped implementation has already outgrown.

The predecessor SPEC itself is also unfinished lifecycle-wise: `status: implemented` and `§E.4 sync_commit_sha: pending-backfill` (its §E.1–§E.4 layout is the modern schema — no §E.5 section is wanted) — it has never been closed. Closing it correctly (backfilling the sync SHA ahead of the close, then attempting `moai spec close`) is part of this remediation, not an afterthought.

**Why a new SPEC, not an increment**: the predecessor is implemented with a frozen audit trail (22 AC audit evidence in §E.2/§E.3 must not be disturbed post-hoc). This remediation — including the predecessor's own correction (3 body edits, version 1.1.0 → 1.2.0) and close — is a separate delivery whose M4 owns those predecessor-file amendments.

During authoring, the card also gained and delivered an urgent first item (commit `52f7ba135`, recorded as M0): the #1648 squash merge had orphaned t250's codemaps stamp commit `0d15864ae90b` — the commit existed only on the branch — leaving main's graph-freshness check not-comparable and red for every inheriting PR. M0 restamped against main-reachable `c9eed8ac6`. The structural defect (squash merge × stamp interaction) and its countermeasure proposals are fenced out as follow-up F5 (§F; research.md §7); the recurrence obligation stays in scope as REQ-GFR-014: every later stamp names a main-reachable commit, and a branch-HEAD restamp is forbidden.

## §B. Scope

### §B.1 In Scope

One already-executed precondition (M0) plus four milestones (execution order M1 → M4):

- **M0 — codemaps provenance restamp (already executed, precondition)**: commit `52f7ba135`, delivered before SPEC issuance — restamp against main-reachable `c9eed8ac6`, repairing the `0d15864ae90b` orphan left by the #1648 squash. Evidence chain (mx scan → build → stamp → rebuild → check, all fresh exit 0): triage-table.md §F5. Not a delegation unit; the run/sync phases inherit its state, and REQ-GFR-014 carries its forward obligation.
- **M1 — test-policy remediation** (triage A-1 findings #1–#10): internal/graph, internal/cli, internal/hook test files — new branch tests, CGO skip guards, vacuous-pass guards, fixture serialization, a result-shape helper, deterministic timing.
- **M2 — code-policy remediation** (triage A-2 findings #12–#21 + Minor-2b): contract honesty in internal/graph, error-surfacing consistency across graph/cli/mx, `--edges`-aware freshness, shipped-key inventory reclassification, comment/literal cleanup.
- **M3 — astx + docs/wording** (triage A-1 finding #11 + A-4 findings #26–#30): go.scm raw-string import capture, astx CGO build tags with `!cgo` fallback, dependencies.md bullet, m5-baseline.md redaction, CHANGELOG count fix, docs-site ko phrase.
- **M4 — predecessor SPEC correction + close** (triage A-3 findings #23–#25 + the card's close task): manager-spec re-delegation applies the three predecessor body corrections and the version bump; `§E.4 sync_commit_sha` is manually backfilled to `2fc4b40a6` ahead of the close (a populated SHA also satisfies the close's 3-phase precondition — no §E.5 section is authored; the modern schema retired it); `moai spec close` is attempted with fallback decided on observed CLI output; predecessor status reaches `completed`.

### §B.2 Boundary decisions

- Finding #11 (astx CGO test tags) is an A-1 test-policy item but **executes in M3** with the astx group (go.scm), per the triage's own grouping note — the astx package is touched once, in one milestone. M1 therefore carries A-1 findings #1–#10; the adopted total remains 29.
- The predecessor's 2 pass-with-debt ACs (AC-GF-012, AC-GF-022) are **recorded, not cleared**: their debt (codemaps md writer adoption; per-task baseline unavailability) is unaffected by this remediation and stays out of scope.
- CR thread resolution on PR #1648 is a card-level sync-phase activity (triage §"PR #1648 스레드 정리 매핑"), not a run-phase milestone of this SPEC.

## §C. History

| Date | Event |
|---|---|
| 2026-08-25 | t250 lands via #1648 (squash `6786c3fa4`); predecessor SPEC status `implemented`, sync SHA pending-backfill |
| 2026-08-26 | CR round-2 69 findings triaged against tree `c9eed8ac6` (3 read-only verify reports); card t279 picked; this SPEC authored |
| 2026-08-26 | M0 executed pre-issuance: codemaps restamp `52f7ba135` against main-reachable `c9eed8ac6`, repairing the `0d15864ae90b` orphan (triage-table.md §F5) |

## §D. Requirements (GEARS notation)

### M1 — test-policy remediation

#### REQ-GFR-001 — CGO skip guards in graph query tests (Ubiquitous)

The graph package's code-query tests that depend on tree-sitter extraction (`TestFileAPI_SignaturesOnly`, `TestFindCodeAndTraceCalls`) shall skip with an explanatory reason when the extraction layer reports `Supported: false`, never fail, so a `!cgo` toolchain yields SKIP verdicts instead of failures (finding 3855002004).

#### REQ-GFR-002 — Named coverage for untested behaviors (Ubiquitous)

The test suites named in triage A-1 shall carry a named test for each currently-untested behavior, each stating its failing input: the citation hash-inconsistency branch (3855001995), Symbol/Via match-content assertions (3855001933), required-parameter rejection plus `..`-path traversal cases for the three MCP handlers (3855001948), and the all-layers-fresh gate notice (3855002099).

#### REQ-GFR-003 — Fixture and assertion hygiene (Ubiquitous)

Provenance fixtures and result-shape checks in the affected test files shall be machine-serialized or helper-guarded: tree-A positive assertions alongside the contamination check (3855002013), `strconv`-built fixture names replacing the hand-rolled itoa/padding helpers (3855149281), `json.Marshal`-built provenance JSON replacing string interpolation (3855001906), and a shared result-shape helper replacing every unguarded `res.Content[0]` indexing site including `mcp_code_tools_test.go:80` (3855001928 + Minor-2a).

#### REQ-GFR-004 — Deterministic budget-overrun observation (When)

**When** the budget-overrun test executes, it shall observe a deterministic, injected duration signal from the refresh path (a new seam in `graph_refresh_cli.go`) rather than assuming any real refresh exceeds the configured budget (3855149237); the seam shall default to wall-clock measurement in production so CLI behavior is unchanged.

### M2 — code-policy remediation

#### REQ-GFR-005 — internal/graph contract honesty (Ubiquitous)

The check/meta layer's documented contracts shall match its behavior: the error path either returns layer reports or the doc comment stops claiming a complete report on failure paths (3855149289); `sidecarAbsentReason` is an immutable constant with an accurate comment (3855149309); the MX-index staleness threshold is injected by the caller rather than hardcoded inside `MXIndexNeedsRefresh` (3855149315); and the source-fingerprint comparison rule lives in one shared helper used by both `EdgesSourcesMoved` and `checkEdges` (3855149325).

#### REQ-GFR-006 — Consistent error surfacing (Ubiquitous)

Error surfacing across the touched production files shall be uniform: graph-build extraction errors wrapped with operation context via `%w` at `symbol.go`'s two bare-return sites (3855149332); the MCP code tools shall use the package's `toolJSON`/`toolErr` instead of the private `jsonToolResult`/`NewToolResultError` (3855001978); `graph stamp` filesystem errors at the CLI boundary shall not leak absolute local paths to the user (3855149248); the `gitOut` comment shall stop claiming empty output with nil error never happens (3855149357); and the swapped CR-ID comments, the literal scan-window bound at `codequery.go:323`, and the "shared by the MCP tool description" overclaim shall be corrected (Minor-2b).

#### REQ-GFR-007 — Selected-artifact freshness evaluation (When)

**When** the graph refresh/build path runs with `--edges` selecting a non-default artifact, the refresh-needed decision shall evaluate the selected artifact's provenance and fingerprints, not the default artifact's (3855149254).

#### REQ-GFR-008 — Shipped key inventory accuracy (Ubiquitous)

The five `graph_freshness` keys in `internal/config/testdata/shipped_key_inventory.yaml` shall be reclassified from class `R` to class `W` with evidence naming the three production readers (`gate.go:167`, `pre_tool.go:840`, `graph_refresh_cli.go:53`) (3855001991).

### M3 — astx + docs/wording

#### REQ-GFR-009 — Raw-string import capture (When)

**When** astx Go extraction meets an import path written as a raw string literal, the `go.scm` query shall capture it as `@code.import` exactly as interpreted string literals are captured today (3855002146).

#### REQ-GFR-010 — astx CGO test gating (Ubiquitous)

The astx package's CGO-dependent positive tests shall carry `//go:build cgo` build tags with a `!cgo` fallback test, so a `!cgo` toolchain exercises the fallback rather than failing on `Supported=false` (3855002141).

#### REQ-GFR-011 — Documentation accuracy and redaction (Ubiquitous)

The four adopted documentation corrections shall land: the `dependencies.md` hook summary bullet names `graph`, matching the Mermaid edge (3855001858); `m5-baseline.md` replaces the developer-local transcript path with a repository-relative label (3855001863); `CHANGELOG.md`'s cumulative MCP count reads "25 to 28" (3855001901); and docs-site ko `graph.md` reads "오래되었으면" (3855149226).

### M4 — predecessor SPEC correction + close

#### REQ-GFR-012 — Predecessor SPEC body correction (Ubiquitous)

The predecessor SPEC-V3R6-GRAPH-FRESHNESS-001 artifacts, amended via manager-spec re-delegation per the ownership matrix, shall read: `REQ-GF-004` carries a third When-clause defining exit 2 for not-comparable system errors (no verdict — neither fresh, stale, nor absent — and the failing operation named), matching the implemented and documented 0/1/2 contract (3855001890); `acceptance.md` §D.1 lists AC-GF-008 in the MUST row and no longer in the SHOULD row, matching §D.4 closure gate 2 whose mutant kill is already observed (3855001874); AC-GF-020's Then-clause carries the non-Go declaration-set qualifier — for non-Go languages the response returns the extracted declaration set without non-exported filtering (3855001867); and `spec.md` frontmatter version reads `1.2.0` with the `updated` field advanced.

#### REQ-GFR-013 — Predecessor close with observed evidence (When)

**When** M4 reaches the close step, the predecessor's `progress.md` §E.4 `sync_commit_sha` shall first be manually backfilled with `2fc4b40a6` (the D3 placeholder-backfill exemption surface) — ahead of the close, because `needsSHABackfill` recognizes only four backfillable forms (empty, `(this commit)`, `(pending)`, `<pending>`; closer.go:397-405) and the predecessor's prose placeholder (`pending-backfill — …`, progress.md:261) matches none of them: without the manual backfill, the close would succeed and freeze that placeholder as the permanent §E.4 value — the ordering prevents placeholder-freezing, not wrong-SHA resolution (the `resolveRecentSpecCommitSHA` auto-resolution path exists in code but is unreachable for this placeholder form) — and no `§E.5` section shall be authored: the predecessor follows the modern 4-section schema, and the close's precondition accepts the 3-phase predicate (§E.4 marker + non-empty `sync_commit_sha`). Then `moai spec close SPEC-V3R6-GRAPH-FRESHNESS-001` shall be attempted with its CLI output recorded verbatim; **Where** the observed output shows a precondition failure, the `--backfill-only` path shall run with its output likewise recorded — the successful path is decided by observed CLI output, never pre-decided. The end state, on whichever path succeeded: predecessor `spec.md` frontmatter `status: completed` and `§E.4` `sync_commit_sha: 2fc4b40a6`.

### Cross-cutting — delivering-PR invariants

#### REQ-GFR-014 — Final-stamp main-reachability (When)

**When** this SPEC's delivering PR reaches its final state, the codemaps provenance stamp shall name a commit reachable from main — the M0 restamp's `c9eed8ac6` or a later main-reachable commit — and the stamp shall not be refreshed against a branch-local HEAD: a squash merge re-orphans branch-HEAD stamps and reddens main's graph-freshness check for every inheriting PR (the `0d15864ae90b` orphaning that made M0 urgent; triage-table.md §F5). A stamp refresh becomes necessary only where described-source churn would exceed the codemaps threshold; the planned M1-M3 churn (~15-20 files) stays below threshold 40 by design, so the M0 stamp is expected to carry to merge unchanged.

## §E. Constraints

- **No new dependencies**; no dependency-version changes.
- **Scope discipline**: the 33 already-fixed findings' sites are cited, not churned; the rejected (R1/R2), deferred (D1), and follow-up (F1-F5) items are untouched (§F).
- **Behavioral conservatism**: M1 is test-only; M2/M3 production edits are the minimal changes the findings name. Where a finding admits two remedies (return reports vs. correct the doc comment), the cheaper honest one may be taken, but the contradiction must be gone.
- **Verification discipline**: targeted `go test ./internal/<pkg>/...` for affected packages only — NEVER `go test ./...` locally (CLAUDE.local.md §4/§6); CI on the delivering PR is the full-suite judge. `go vet` on touched packages. Docs-site changes verified with `hugo -s docs-site --minify --gc`.
- **SPEC ownership**: predecessor `spec.md`/`acceptance.md`/`plan.md` body edits and the §E.4 SHA backfill happen ONLY via manager-spec re-delegation (D-NEW-1 inline-fix pattern; D3 placeholder-backfill exemption); the close CLI performs its own atomic commit.
- **Test isolation**: new tests use `t.TempDir()`; no OTEL env vars in parallel tests; no real-network or LLM paths.
- **No `depends_on`**: the predecessor is `implemented`, not `completed` — a strict depends_on pre-flight would block circularly (completing it is this SPEC's M4). The relation is carried as `related_specs`.
- **Stamp reachability (HARD)**: at no phase of this SPEC is the codemaps stamp refreshed against a branch-local HEAD; the delivering PR's final stamp names a main-reachable commit (REQ-GFR-014).

## §F. Exclusions (What NOT to Build)

### Out of Scope — workflow SHA pinning (follow-up F1)

- Pinning `graph-freshness.yml` actions to commit SHAs and adding `persist-credentials: false` (3855001854). A repo-wide policy decision — every workflow uses the same `@v7` tag convention; pinning one file diverges from it. Separate card.

### Out of Scope — ctx propagation into graph APIs (follow-up F2)

- Threading `context.Context` through `graph.FileAPI`/`FindCode`/`TraceCalls` (3855001962). A functional API change, not a policy fix. Separate card.

### Out of Scope — fingerprint anchoring without git (follow-up F3)

- Anchoring the stamp on the content fingerprint when git is unavailable (3855149371). Conflicts with REQ-GF-003's commit-anchor design; honest-absent is the intended behavior. Separate card.

### Out of Scope — scanned-package evidence separation (follow-up F4)

- Returning scanned-package evidence separately from import edges in `symbol.Extract` (3855002093). Return-signature change with call-site blast radius. Separate card.

### Out of Scope — squash-merge × stamp reachability countermeasures (follow-up F5)

- The CI pre-merge stamp-reachability guard: a graph-freshness-job step verifying that `provenance.json`'s `commit_sha` is an ancestor of the PR base (origin/main) — the only point that catches an orphan stamp before merge (primary proposal).
- The `moai graph stamp codemaps --commit <main-ancestor>` explicit-commit mode (ergonomic support), and the recorded-but-not-proposed heavier alternatives (post-merge main restamp routine; merge-commit strategy switch).
- Full incident record and analysis: `.moai/reports/t279/triage-table.md` §F5. Delivered inside THIS SPEC: only the recurrence constraint (REQ-GFR-014) and the executed M0 restamp — the countermeasure itself is follow-up card candidate F5.

### Out of Scope — rejected findings R1/R2

- R1 (3855149188): the moai-mcp-tools.md 28/24 totals match `catalog.go:41-70` exactly; the CR miscounted. No edit.
- R2 (3855149192): the committed absolute `tree_root` in provenance.json is REQ-GF-003-mandated design for a tracked artifact; `check.go:153-161` documents it citing this very comment-id. No edit.

### Out of Scope — deferred-by-design D1

- D1 (3855149345): the unconfigured (nil) skip notice stays silent — recorded deviation preserving the pre-existing silent-pass contract (progress.md §E.2 AC-GF-006, gate.go:1185-1189). No edit.

### Out of Scope — predecessor debt clearance and CR thread resolution

- Clearing AC-GF-012/AC-GF-022 pass-with-debt (codemaps md writer adoption rides its next regeneration; the per-task baseline is unobtainable) — M4 records them, it does not resolve them.
- Resolving the 42 unresolved PR #1648 threads: card-level sync-phase work per the triage's mapping section, after this SPEC's PR merges.

## §G. Success Criteria

| Milestone | Success verdict |
|---|---|
| M1 | Each of the 10 findings has its named test/change present and green; `CGO_ENABLED=0` legs pass with skips, not failures |
| M2 | Each of the 10+Minor-2b findings' observable state flipped (const, helper, `%w`, class W, corrected comments); targeted package tests green |
| M3 | go.scm captures raw-string imports (named test); astx CGO gating in place; all four doc targets read corrected; hugo build green |
| M4 | Predecessor carries the exit-2 clause, the MUST-row promotion, the language qualifier, version 1.2.0; §E.4 SHA backfilled `2fc4b40a6` ahead of the close (no §E.5 authored); close output recorded verbatim (whichever path); `status: completed` |
| Cross-cutting | Delivering PR's final stamp names a main-reachable commit (`c9eed8ac6` unless a recorded main-reachable refresh); `moai graph check` green on the PR head |

## §H. Cross-References

- Evidence base: `.moai/reports/t279/triage-table.md` (scope definition; §F5 = M0 restamp record + squash-merge × stamp structural defect) + `verify-graph.md` / `verify-cli.md` / `verify-docs.md` (per-finding current-tree citations, baseline `c9eed8ac6`)
- Predecessor: `.moai/specs/SPEC-V3R6-GRAPH-FRESHNESS-001/` (spec.md, acceptance.md, progress.md — M4 amendment targets)
- Lineage: card t250 → PR #1648 (squash `6786c3fa4`) → CR round-2 (`.moai/reports/t250/cr-round2-comments.md`) → card t279
- Close CLI contract: `moai spec close --help` + code-verified at `internal/spec/closer.go` (`validatePreconditions` :694 — precondition 2 accepts legacy §E.5 OR §E.4 marker + non-empty `sync_commit_sha`; `hasGenuinePassWithDebtVerdict` :667 scans acceptance.md only; backfill resolution :324; explicit-path staging :340; generated subject :354)
- Verification conventions: CLAUDE.local.md §4/§6 (targeted tests, never local full suite)
