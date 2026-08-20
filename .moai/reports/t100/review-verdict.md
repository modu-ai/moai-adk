# t100 review verdict — edges.jsonl BFS reader (find_callers / blast_radius)

- Reviewer: lead session (review hub)
- Card: t100 · Branch: `WT-t100` @ `02d9cac05` (base `0ede5db6a`) · Worktree: `.claude/worktrees/agent-a8b0420c635295c5a`
- Delta reviewed: `0ede5db6a..02d9cac05` (6 files, +594/−1: reader/query/CLI + 3 test files)
- Lens: default 4-perspective + new-logic precision review
- Evidence read: `.moai/reports/t100/report.md` (81 lines, 5-section, real-data query outputs preserved alongside)
- Note: verdict file in the lead's release-v311 tree (worktree isolation — same as t103/t110).

## Verdict: PASS

## 1. Dispatch focus items (all five verified)

| # | Focus | Check performed | Result |
|---|-------|-----------------|--------|
| 1 | mx-spec bidirectional rule soundness / no direction contamination | `BlastRadius` read in full: rev-map gets `rev[target]+=source` for ALL kinds (import/spec-depends reverse-only — importer/dependent is affected, never the converse) plus `rev[source]+=target` for `KindMXSpec` only. Each kind contributes exactly its declared directions; the BFS is over that kind-scoped map, so no query leaks across directions. Q2a evidence (file blast reaches its SPEC) demonstrates the intended bidirectional reach; Q1 (--callers SPEC → 7 files, 12 edges deduped) confirms edge orientation | PASS |
| 2 | --specs-no-code universe completeness (623) | Reviewer-measured on the primary: 626 SPEC-prefixed dir entries, 625 with spec.md, 623 LoadSpecDependencies keys. The loader's skip semantics (ReadFile error / parseFrontmatter error / empty ID → skip) are DECLARED contract ("mirrors LoadSpecModules: same glob, same parseFrontmatter, same skip-on-error") — so 623 is faithful filtering, not omission. The 3-entry delta (1 dir sans spec.md, ~2 unparseable/empty-ID) is a data-hygiene matter, not a t100 defect | PASS (hygiene note) |
| 3 | errors.Is replacement caught by tests | `graph.go:95` uses `errors.Is(err, os.ErrNotExist)` with the %w-chain rationale comment. Two-level test corroboration: reader-level (`reader_test.go:134` absent file → err non-nil) and CLI-level (`TestGraphQueryCmd_MissingArtifact:212` asserts the friendly error points at 'moai graph build' — the errors.Is branch's message) | PASS |
| 4 | [HARD] caveat on every --specs-no-code output | Single text output path (no --json flag; flags are callers/blast/fanin/specs-no-code/limit/edges/root); `unreferencedSpecCaveat` Fprintln'd unconditionally in the --specs-no-code branch; `TestGraphQueryCmd_SpecsNoCodeIncludesCaveat:176` pins both the NOTE text and the `미연결 ≠ 미구현` string | PASS |
| 5 | Interface stability for t107 + 257-vs-258 | `FindCallers`/`BlastRadius` exported pure functions (`([]Edge, string) []string`) — names fixed per the t107 consumption promise. 257 edges (mx-spec 89 vs t99's 90) = writer regenerated against the tree AFTER later merges (t106 et al.); the reader consumes format only — unaffected. Accepted lane-attributed with sound reasoning | PASS |

## 2. Code quality observations

- `LoadJSONL`: blank-line skip, 1-based line-number errors, scanner error propagation — clean.
- `ImportFanIn` counts DISTINCT importers per target (map-of-maps), deterministic sort (count desc, path asc) — correct ranking semantics.
- `UnreferencedSpecs` is a pure function with the universe passed in — testable, no I/O coupling.
- CLI selector discipline: exactly one of four selectors, enforced + tested (`RequiresExactlyOneSelector`).

## 3. Baseline attribution

Code reads @ `02d9cac05`. Test outputs (graph ok 0.681s, cli TestGraph ok 1.502s, vet/lint, GOOS windows/linux vet, real-data queries) lane-attributed on the card worktree. Reviewer independently measured the SPEC universe counts on the primary.

## 4. Gaps

- Full suite not run locally (lane discipline — CI owns it); cli package only TestGraph-filtered.
- tag-kind absence (no @MX:DEBT edges) accepted from t99's prior investigation, not re-measured — mitigated by the fanin stand-in design.
- The 3-entry SPEC data delta (626/625/623) not attributed to named SPECs — hygiene follow-up material.

## 5. Residual risks

- mx-spec bidirectionality over-approximates "implementation clusters": a file implementing two SPECs joins both clusters (declared design; query consumers should read results as closure over co-implementation).
- ImportFanIn coverage is bounded by codemaps dependencies.md artifacts.
- Data hygiene: 1 SPEC dir without spec.md + ~2 spec.md with unparseable/empty frontmatter IDs are invisible to --specs-no-code (consistent with the loader's declared contract; a hygiene card could name and fix them).
