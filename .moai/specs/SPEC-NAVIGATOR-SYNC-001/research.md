# SPEC-NAVIGATOR-SYNC-001 — Research

> Tier L research artifact. Grounds the SPEC in (a) verified repo assets on `main`, (b) external references from the design report.
> Per `verification-claim-integrity.md` §2: every asset claim below is attributed to a path that resolves on `main` at plan-phase time.

## §1. Repo asset inventory (design report §4 — verified against main)

| Design-report claim | Path on main | Verified? | Notes for M0 |
|---|---|---|---|
| `internal/navigator/astx/` Tier-3 deterministic engine | `internal/navigator/astx/{astx,enrich}.go` + `queries/` | YES (8 .go files + queries/) | Reuse: `enrich.go:21 Provenance`, `enrich.go:122 parseCapabilityMap` |
| `enrich.go:122` header-driven parser | `internal/navigator/astx/enrich.go:122` (parseCapabilityMap) | YES (read lines 110-160) | M0 reuses for capability-map parse |
| `enrich.go:21,83` Provenance model | `internal/navigator/astx/enrich.go:21` (struct), used in `ExtractCommitSHA`/`CapturedAt` | YES (read lines 15-30) | M0's REQ-NS-009 carries byte-for-byte |
| `navigator_enrich.go:128` atomic-write + barrier | `internal/cli/navigator_enrich.go:128` (atomicWrite), `:138 NAVIGATOR_PRE_RENAME_BARRIER` | YES (read lines 124-148) | M0's REQ-NS-010 mirrors verbatim |
| `navigator-audit.sh` token-normalization matching | (not inspected in detail — owned by 002, M0 does not modify) | PARTIAL | M0 does NOT reuse the matching; M1 (Detect) may |
| `@MX:SPEC` token "흡수 대상" (absorb target) | `internal/mx/scanner.go:139-155`, `spec_association.go:17-20` | PRECISION CORRECTION — token is NOT dormant | M0 BRIDGES mx output, does NOT re-scan (REQ-NS-005) |
| `/moai codemaps` Phase-2 dep-graph | `.moai/project/codemaps/dependencies.md` present on disk (6 codemaps files) | YES (ls confirmed) | M0 does not consume dep-graph directly; M4 (4-tier map) will |
| 001/002/003 non-overlap REQs | `REQ-PN-016` (001:139), `REQ-NA-011` (002:118), `REQ-NT-018` (003:158), `REQ-NT-019` (003:161) | YES (grep confirmed all 4) | Carried forward as REQ-NS-013/015/016 |

## §2. External references (from design report §레퍼런스 — design-phase cited, not re-verified)

The design report (session 3db3f943) lists these as the external grounding for the BAS approach. They are NOT re-verified at plan-phase (the design phase already established provenance). Listed here for traceability:

- **CodeWiki** (FSoft-AI4Code/ACL2026) — hierarchical decomposition + `--update`/`--compare-to` incremental pattern. Relevant to M3 (Fix) and M4 (Tier-3 extension).
- **DeepWiki / DeepWiki-Open** — living documentation model.
- **falconer living-documentation** — the 3-element Detect→Route→Fix framing M0 underpins (M0 is the substrate; M1/M2/M3 are the 3 elements).
- **Martin Fowler SDD 3-level** — spec-anchored as the reversible optimum (not spec-first deletion, not spec-as-source MDD trap).
- **Augment SDD** — "precise less context > million tokens" — independently validates the design's core thesis.
- **Kiro Design** — 7-section blueprint template (component architecture / data flow / data model / error handling / test strategy / implementation approach / migration). Relevant to M4.
- **GitHub Spec Kit, Tessel (`document --code`), SCIP (Sourcegraph), tree-sitter, LSIF, Doxygen, OpenAPI/Protobuf contract-first, log4brains (ADR), Cucumber, Bazel/Buck** — the broader reference set. M0's substrate design is agnostic to most of these; they bind at M4/M5.

## §3. Verified: mx-scanner consumes `@MX:SPEC` (not dormant)

Direct inspection of `main`:

```
internal/mx/scanner.go:139    // @MX:SPEC sub-line capture (REQ-MX-ASSOC-001).
internal/mx/scanner.go:146    if strings.Contains(upperLine, "@MX:SPEC") {
internal/mx/scanner.go:396    re := regexp.MustCompile(`@MX:SPEC:\s*(SPEC-[A-Z0-9][A-Z0-9-]*)`)
internal/mx/spec_association.go:17  // @MX:NOTE: [AUTO] SpecAssociator — unifies three SPEC-association sources:
                                    //   path-based, body-based, and sub-line (@MX:SPEC).
internal/mx/tag.go:51               // SpecRef is the @MX:SPEC sub-line content (optional, author-intended SPEC
```

**Implication for M0**: the integration layer's `@MX:SPEC`-edge source is `mx.SpecAssociator`'s output, NOT a re-scan. This is codified in REQ-NS-005 and design.md §4.

## §4. Verified: navigator/ output paths on disk

The `.moai/project/navigator/` directory is EMPTY on this checkout (capability-map.md is generated on demand by `/moai project`). `.moai/project/codemaps/` has 6 files including `dependencies.md` (the Phase-2 dep-graph). The M0 output `nav-graph.json` lands alongside `capability-map.md` under `.moai/project/navigator/` — no directory creation needed beyond what 001 already provisions.

## §5. Risk register (from design report §리스크, M0-relevant subset)

| # | Risk | M0 mitigation |
|---|---|---|
| 1 | spec-as-source trap (MDD hazard) | M0 emits `nav-graph.json` as a READ-ONLY artifact; no reverse writes (those are M3). Graph is a derivative, not a source. |
| 2 | Heavy up-front spec | M0 is the smallest meaningful substrate (3 tokens + 1 join). Fowler Verschlimmbesserung — write cheaply, evolve cheaply. |
| 4 | SCIP indexer maturity | M0 does NOT depend on SCIP. M0 reuses 003's tree-sitter-based astx. SCIP is M4. |
| 6 | "Generation = done" misunderstanding | M0 explicitly does NOT include Detect/Route/Fix. The substrate alone is not the BAS; M1-M3 are required for the full Falconer loop. The Epic status is communicated via the milestone map in spec.md §E. |

Risks 3 (brownfield fit) and 5 (unidirectional habit) bind at M5 / M1 respectively; not M0.

## §6. Boundary vs LSEL (explicit carry-forward)

The 3 predecessor SPECs each established a non-overlap boundary vs SPEC-LSEL-LOCAL-EVOLUTION-001. M0 inherits this boundary. Concretely:

- M0 reads: `.moai/project/{*.md, navigator/capability-map.md, navigator/audit-report.json, codemaps/capability-symbols.json}`, `.moai/docs/**/*.md`, the Go source tree (for `@NAV:SYM` only, NOT LSEL surfaces).
- M0 writes: ONLY `.moai/project/navigator/nav-graph.json` (+ `.tmp` transient, `.moai/logs/navigator-sync.log` advisory).
- M0 does NOT touch: `.moai/lessons-inbox.jsonl`, `.moai/state/lsel/`, `memory/feedback_*.md`, `hns-lsel-*`, `.moai/state/lsel/proposals/`.

This is codified in REQ-NS-013/014 and tested via AC-013/014 (grep-pair protocol).
