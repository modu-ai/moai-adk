# SPEC-NAVIGATOR-SYNC-003 — Research

> Tier L research artifact. Records: (§1) the design-report asset inventory verified against current `main`; (§2) external references; (§3) the non-overlap verification result (the load-bearing cross-reference to the five predecessor SPECs); (§4) the SPEC-ID re-numbering note.

## §1. Design-report asset inventory — verified against current `main`

The design report (`.moai/reports/navigator-redesign-bas-20260805.md` §자산 재사용 line 31-39) lists assets to reuse. Each is verified below against the current `main` (read-only Bash, 2026-08-06):

| Asset (design report claim) | Verified location on `main` | M4 reuse | Status |
|-----------------------------|----------------------------|----------|--------|
| `internal/navigator/astx/` — Tier 3 deterministic engine | `internal/navigator/astx/{astx.go,enrich.go,measure_cgo.go,measure_nocgo.go}`; `type Symbol struct` at `astx.go:22`; `EnrichRows` at `enrich.go:212`; `Provenance` at `enrich.go:21` | Imported as-is; output extended additively (per-symbol signature + declaration + references) | VERIFIED — reuse honored (REQ-NS3-011/013) |
| `/moai codemaps` Phase 2 dep-graph | `.moai/project/codemaps/{overview.md, modules.md, data-flow.md, dependencies.md, docs-truth.md, entry-points.md}` (auto-generated, present) | `dependencies.md` is the Tier 1 `module_tree.json` scaffold seed (read-only) | VERIFIED — reuse honored (REQ-NS3-004) |
| `@MX:SPEC` token — code→SPEC back-pointer (absorb) | `internal/mx/scanner.go:139-155` + `internal/mx/spec_association.go:17-20` (3-source: path + body + sub-line `@MX:SPEC`) | ALREADY bridged into the graph by M0 (REQ-NS-005, bridge-not-absorb). M4 does NOT re-handle `@MX:SPEC`. | VERIFIED — M0 owns this; M4 is consumer-only on M0 |
| M0's graph-join + binding-token trio | `internal/navigator/sync/{schema,scan,join,write,mx_bridge,provenance,util}.go` (9 files); `.moai/project/navigator/nav-graph.json`; `.claude/rules/moai/workflow/nav-tokens.md` | `nav-graph.json` consumed read-only; `tiers.json` emitted as OVERLAY (REQ-NS3-018) | VERIFIED — reuse honored |
| Existing SPEC 3-kind non-overlap contracts (REQ-PN-016, REQ-NA-011, REQ-NT-018/019) | `.moai/specs/SPEC-PROJECT-NAVIGATOR-{001,002,003}/spec.md` (all `status: completed`) | Carried forward as REQ-NS3-017 (see §3) | VERIFIED — see §3 |

**Additional baselines observed** (not in the design report's reuse list, but load-bearing for M4's design):

| Baseline | Location | M4 use |
|----------|----------|--------|
| Pre-existing ADR home | `.moai/decisions/{lsp-client-choice.md, skill-rename-map.yaml}` (ADR-shaped, pre-date the formal scheme) | Tier 2 ADR home; grandfathered immutable (design.md §1.D2) |
| M1 Detect implementation | `internal/navigator/detect/` + `internal/hook/navigator_detect{,_test,_nonoverlap_test,_hardening_test,_coverage_test}.go` + a branch in `internal/hook/post_tool.go`; writes `.moai/state/navigator-detect/<session-id>.jsonl` | Consumer-only (REQ-NS3-016); M4's blueprint↔code drift surfaces as a debt signal in M1's detect channel |

**Conclusion**: every asset the design report names for reuse is present on `main` at the claimed location. No asset needs to be rebuilt. M4 is purely additive ABOVE these assets.

## §2. External references (verified via the design report's §레퍼런스 list)

The design report (line 73-74) lists verified external references. M4's tier design maps to them as:

| Reference | Tier | Role in M4 |
|-----------|------|------------|
| **CodeWiki** (FSoft-AI4Code/ACL2026) | Tier 3 | The `--update`/`--compare-to` incremental narrative technique + `metadata.json` per-symbol tracking (REQ-NS3-012); hierarchical decomposition grounds the 2-tier split (design.md §5) |
| **Martin Fowler SDD 3-level** | Tier 1 | The spec-anchored reversible-optimum stance (REQ-NS3-006, design.md §4) — the load-bearing rejection of both spec-first (deletion) and spec-as-source (MDD) |
| **Kiro Design** (7-section) | Tier 1 | The per-module `overview.md` template structure (REQ-NS3-005) |
| **SCIP** (Sourcegraph) | Tier 3 | Forward-compat per-language deterministic structure source (REQ-NS3-013); NOT vendored at M4 (리스크 #4) |
| **log4brains / ADR** (Architecture Decision Records) | Tier 2 | The immutable + supersede ADR pattern (REQ-NS3-008/009); the four canonical fields (Decision Date / Status / Context / Decision / Consequences) |
| **OpenAPI / Protobuf contract-first** | Tier 0 | The build-enforced drift-immune contract principle (REQ-NS3-001/002); translated to Go-native surfaces (design.md §1.D4) |
| **falconer living-documentation** | cross-cutting | The 3-element loop (detect→route→fix); M4 produces the map the loop operates on (M1 Detect shipped; M2 Route + M3 Fix are sibling SPECs) |
| **Augment SDD** ("precise small context > million tokens") | §A rationale | Validates the 오라버니 insight (design report §핵심 테제); the headline success metric (REQ-NS3-022) measures it directly |
| **tree-sitter / LSIF / Doxygen** | Tier 3 alt | Cited as the deterministic-structure design space SCIP occupies; not vendored |

These references are the design report's verified set (the report's §레퍼런스 explicitly marks them "검증됨"). M4 does NOT introduce new external dependencies beyond what the design report already vetted.

## §3. Non-overlap verification result (load-bearing)

M4 integrates ABOVE five predecessor SPECs. The non-overlap contracts each predecessor codified are carried forward unchanged; M4 does NOT relax any. The verification maps each predecessor REQ to its M4 carry-forward:

| Predecessor REQ | Predecessor SPEC | Contract (one-line) | M4 carry-forward | Verification |
|-----------------|------------------|---------------------|------------------|--------------|
| REQ-PN-016 | SPEC-PROJECT-NAVIGATOR-001 (regen, completed) | Navigator ↔ LSEL non-overlap | REQ-NS3-017 | M4 writes no LSEL surface; grep-pair AC-NS3-017 |
| REQ-NA-011 | SPEC-PROJECT-NAVIGATOR-002 (audit, completed) | Audit ↔ LSEL + Audit ↔ 003 non-overlap | REQ-NS3-017 | M4 writes no 002 audit surface; AC-NS3-017/018 |
| REQ-NT-018 | SPEC-PROJECT-NAVIGATOR-003 (enrich, completed) | Enricher ↔ LSEL non-overlap | REQ-NS3-017 | M4 writes no LSEL surface; AC-NS3-017 |
| REQ-NT-019 | SPEC-PROJECT-NAVIGATOR-003 (enrich, completed) | Enricher ↔ 002 audit non-overlap | REQ-NS3-017/018 | M4 writes no `audit-report`; AC-NS3-018 |
| REQ-NS-005 | SPEC-NAVIGATOR-SYNC-001 (M0, completed) | `@MX:SPEC` bridged not absorbed; `internal/mx/` untouched | REQ-NS3-016 | M4 is consumer-only on M0; AC-NS3-016 |
| REQ-NS-013 | SPEC-NAVIGATOR-SYNC-001 (M0, completed) | Integration layer ↔ LSEL non-overlap | REQ-NS3-017 | M4 writes no LSEL surface; AC-NS3-017 |
| REQ-NS-015 | SPEC-NAVIGATOR-SYNC-001 (M0, completed) | Integration layer ↔ 002 audit write non-overlap | REQ-NS3-017/018 | M4 writes no `audit-report`; AC-NS3-018 |
| REQ-NS-016 | SPEC-NAVIGATOR-SYNC-001 (M0, completed) | Integration layer ↔ 003 enrich write non-overlap | REQ-NS3-017/018 | M4 writes no `capability-symbols`; AC-NS3-018 |
| REQ-NS2-005 | SPEC-NAVIGATOR-SYNC-002 (M1, completed) | Detect consumer-only on M0 + mx | REQ-NS3-016 | M4 consumer-only on M1; AC-NS3-016 |
| REQ-NS2-008 | SPEC-NAVIGATOR-SYNC-002 (M1, completed) | Detect ↔ 3 predecessors non-overlap | REQ-NS3-017/018 | M4 writes no predecessor surface; AC-NS3-017/018 |

**Write-surface isolation summary** (REQ-NS3-018): M4's only write surfaces are the 6 named paths — `.moai/project/blueprint/{module_tree.json, <module>/overview.md, contracts.yaml}`, `.moai/decisions/<dec-id>.md` (NEW only), `.moai/project/navigator/tiers.json`, `.moai/project/navigator/symbols/<symbol>.md`. M4 does NOT write `nav-graph.json` (M0 owns it; overlay pattern, design.md §1.D5), does NOT write any predecessor surface, does NOT write any LSEL surface, does NOT modify any M0/M1 producer Go package.

**Consumer-only summary** (REQ-NS3-016): M4 imports `internal/navigator/astx/` as-is (extended additively, not modified), reads `nav-graph.json` (M0) read-only, reads `/moai codemaps` output read-only, reads `.moai/decisions/` read-only (writes NEW ADRs only). The M0/M1 producer Go packages (`internal/navigator/sync/`, `internal/navigator/detect/`, `internal/hook/navigator_detect*.go`, `internal/hook/post_tool.go`) are untouched in M4's run-phase commit range (AC-NS3-016 Lens 1: `git diff` empty).

**Conclusion**: non-overlap is fully verified. M4 is a NEW integration layer ABOVE the five predecessors; no contract is relaxed; every predecessor REQ has a named M4 carry-forward REQ + a named AC.

## §4. SPEC-ID re-numbering note (transparency)

M0's §E (spec.md) projected the Epic's SPEC IDs as a strict sequence: 002=M1, 003=M2 (Route), 004=M3 (Fix), 005=M4 (4-tier map), 006=M5 (brownfield). In practice, the IDs are assigned in AUTHORING order, and M4 (4-tier map) was prioritized ahead of M2 (Route) and M3 (Fix). The actual assignments:

- 001 = M0 (authored 1st, `status: completed`)
- 002 = M1 (authored 2nd, `status: completed`)
- **003 = M4 (authored 3rd, this SPEC)**
- 004 / 005 / later = M2 (Route) / M3 (Fix) / M5 (brownfield), to be authored in a later session

This re-numbering changes NO technical content. It is recorded here (and in spec.md HISTORY) so a reader cross-referencing M0's projection is not confused. The `depends_on: [SPEC-NAVIGATOR-SYNC-001]` relationship is to M0 only (M0 is the sole dependency; M2/M3/M5 are parallel, no cross-dependency). The `related_specs` field carries all five predecessors for traceability.

## §5. Risk register (from design report §리스크, M4-relevant)

| # | Risk (design report) | M4 mitigation | REQ / design.md |
|---|----------------------|---------------|-----------------|
| 1 | spec-as-source trap | spec-anchored stance fixed (blueprint drift is debt, not failure) | REQ-NS3-006; design.md §4 |
| 2 | heavy up-front spec | write the blueprint cheaply, let it evolve (Fowler Verschlimmbesserung warning) | REQ-NS3-004 (scaffold-then-refine); design.md §1.D3 |
| 3 | brownfield fit | defer to M5 (reverse-extraction) | §E Out of Scope |
| 4 | SCIP indexer immaturity | astx+build graph first; SCIP per-language gradual | REQ-NS3-013; design.md §1.D1 |
| 5 | one-directional habit | M1 Detect shipped first; Fix is later (M3) | (Epic-level; M4 produces the map M1 already detects against) |
| 6 | "generation = done" misunderstanding | Fix element (M3) will force the refresh infrastructure | (Epic-level; out of M4 scope) |
