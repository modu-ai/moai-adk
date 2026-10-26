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

## Round 2 (re-review) — pending

Repair commit: 05934379a4 (8/8 fixed-with-RED-evidence per manager-develop report; lane re-observation: build OK, fetchweb/jev fresh-run ok, off-gate + E2E + R6/R7 repair tests PASS at 05934379a4).
