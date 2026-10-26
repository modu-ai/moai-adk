# card-review — t1968 (SPEC-FETCH-WEB-001)

- backend: codex (mcp codex_review, scope=card) · advisory
- base: 2a1898ea22198e1cac229975e924943969e4b432 (merge base) · tree: lane worktree t1968
- head reviewed: 77027d5e79 (D3 backfill commit)
- verdict: **fail** (advisory) — 8 findings, all P2
- review scope note: static code review vs the pinned base; no tests/builds run by the reviewer

## Findings and dispositions

| # | Severity | Location | Finding | Disposition |
|---|----------|----------|---------|-------------|
| R1 | P2 | internal/fetchweb/scoring.go:215 | stage-2 sends all candidates in ONE request; >64k-token totals rejected by checkBounds after stage-1 → fetch_error | **repair** (round 1) — batch stage-2 requests within the Jev size limit |
| R2 | P2 | internal/fetchweb/blocks.go:257 | `<br>` joins adjacent text nodes without separator → "firstsecond"; word boundary lost | **repair** — violates verbatim guarantee (REQ-FW-006) |
| R3 | P2 | internal/fetchweb/blocks.go:156 | `<li><p>text</p></li>` emits the li block AND revisits the child p → duplicate blocks | **repair** — duplicate output/tokens; violates verbatim/single-emission |
| R4 | P2 | internal/fetchweb/blocks.go:291 | strings.Trim strips the FIRST code line's indentation, changing meaning | **repair** — violates verbatim code preservation |
| R5 | P2 | internal/fetchweb/scoring.go:337 | neighbor context cut by BYTES not runes — CJK yields <400 chars, may split mid-rune | **repair** — rune-based 400-char context |
| R6 | P2 | internal/jev/jev.go:729 | semaphore slot released after headers; body still being read → in-flight can exceed MaxConcurrency | **repair** — hold slot until body read+close |
| R7 | P2 | internal/fetchweb/scoring.go:193 | StaleCacheWarning dropped — stale-model scoring invisible to CLI/MCP users | **repair** — propagate warnings to final output (AC-FW-011 contract) |
| R8 | P2 | internal/cli/jevmeasure_cmd.go:152 | non-regression comparison FAILs don't set exit code (AnyRegression only) | **repair** — any Pass==false → exit 1 per the stated compare contract |

Disposition summary: 8/8 accepted for repair (round 1) — R2/R3/R4 directly violate the SPEC's verbatim-emission guarantee; R6/R7 violate the card's own client contracts; R5 is a CJK correctness defect; R1/R8 are contract-correctness defects. No finding not-adopted. Re-review ceiling: 2 (this was review 1).

## Round 2 (re-review) — verdict: fail, 4 new P2 findings (R1-R8 fixes held)

Review head: d7ca574310. The 8 round-1 fixes were not contested; the reviewer found 4 new issues in the changed code:

| # | Severity | Location | Finding | Disposition |
|---|----------|----------|---------|-------------|
| N1 | P2 | internal/fetchweb/scoring.go:213 | stage-1 windows count body only; question+framing overhead can exceed MaxRequestTokens on many-short-block/long-question pages → fetch_error on normal pages | **repair** (round 2) — request-size-aware stage-1 splitting |
| N2 | P2 | internal/jev/jev.go:639 | cache stored under OVERRIDE model id when response model differs → wrong-model answers cached under wrong key, later clean-hit reads | **repair** — key/store by response model id or refuse on mismatch |
| N3 | P2 | internal/fetchweb/blocks.go:282 | script/style inside p/li/cells bypass chrome exclusion in recursive text extraction → scored + emitted as content | **repair** — chromeTags subtrees skipped in collapseText |
| N4 | P2 | internal/fetchweb/blocks.go:148 | aside headings mutate shared s.path; body blocks after aside inherit the aside's heading path (provenance + same-heading rescue corrupted) | **repair** — save/restore path around aside |

Ceiling accounting: round 1 + this re-review = 2 runs; ONE re-review remains (round 3) before dispositions are raised to the leader.

## Round 3 (final re-review) — verdict: fail, 3 P2 findings — CEILING EXHAUSTED

Review head: 61436e6e68. Findings (static review, no tests run by reviewer):

| # | Severity | Location | Finding | Disposition |
|---|----------|----------|---------|-------------|
| C1 | P2 | internal/fetchweb/blocks.go:155 | N4's fix incomplete: `saved := s.path` shares the backing array — heading truncate+append can overwrite the saved path; aside-following body can still inherit the wrong heading path | **fixed post-ceiling** (slice copy) — round-2 disposition "N4 repaired" was optimistic; corrected here |
| C2 | P2 | internal/fetchweb/blocks.go:168 | `<pre>`/`<table>` inside `<li>`: collapseText includes them in list text AND walkChildren re-emits them → duplicate blocks + flattened code copy | **fixed post-ceiling** (exclude separately-emitted children from list text) |
| C3 | P2 | internal/jevmeasure/bench.go:111 | missing case ID in run results maps to zero value (Contained=false, 100% savings); baseline also missing → PASS/exit 0 without any run result | **fixed post-ceiling** (existence checks; missing results fail explicitly) |

**Ceiling disposition (mechanics § The card-review stage):** 3 review rounds run (initial + 2 re-reviews = the ceiling). The 3 findings are fixed WITHOUT a fourth review — the lane does not widen the ceiling. Residual verification authority passes to the independent sync-audit (sync-auditor), which judges the whole card including these fixes. This ceiling disposition is named in the lane's completion report to the leader.

## Post-ceiling repair record (C1-C3)

Commit 16dc4ae0b4 (+162/−7, 5 files): C1 true-copy aside save (RED: `[Note]` → want `[Main]` observed pre-fix), C2 listItemSkip excludes separately-emitted kinds from li text (RED: duplicate/flattened blocks observed pre-fix), C3 DiffAgainstBaseline existence check (RED: missing-run PASS/exit 0 observed pre-fix). Lane re-observation at 16dc4ae0b4: BUILD_OK, fetchweb/jevmeasure fresh ok, off-gate/block/bench targeted tests ok. Gaps (honest): full internal/cli suite not observed green locally — default 10m timeout under machine load (avg 147-290, foreign concurrent suites), no test FAIL; repository-wide verdict PENDING on integration-branch CI. Residual-risk (recorded by the repairer): skipped-subtree text runs inside li still merge adjacent words (`a<pre/>b` → "ab") — pre-existing ul/ol skip semantics, out of C2's scope, unchanged.
