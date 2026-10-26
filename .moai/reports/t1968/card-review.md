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
