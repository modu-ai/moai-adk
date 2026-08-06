# SPEC-NAVIGATOR-SYNC-003 — Implementation Plan

> Tier L · status: draft · v0.1.0 · plan-phase only (Implementation Kickoff Approval held by orchestrator; plan-audit runs on these working-tree files next)

## §A. Context

BAS Epic M4 — the 4-tier addressable map. The design report (`.moai/reports/navigator-redesign-bas-20260805.{md,html}`, session 3db3f943) is the SSOT for rationale; this plan is the HOW. M4 EXTENDS the M0 graph (`SPEC-NAVIGATOR-SYNC-001`, `status: completed`) with four tier-specific enrichment overlays. M4 is the third-authored SPEC in the Epic (HISTORY re-numbering note: 003=M4, not M2 as M0 projected).

### §A.1 What M4 builds ON (consumer-only, do not modify)

| Producer | Artifact M4 consumes | Owner SPEC |
|----------|---------------------|------------|
| `internal/navigator/sync/` | `nav-graph.json` (nodes: decision/spec/symbol; edges: dec-edge/spec-edge/sym-edge) | M0 (REQ-NS-006) |
| `internal/navigator/detect/` + `internal/hook/navigator_detect*.go` | `.moai/state/navigator-detect/<session-id>.jsonl` (M1 surfaces blueprint↔code drift) | M1 (REQ-NS2-003) |
| `internal/navigator/astx/` | `Symbol` struct (`astx.go:22`), `EnrichRows`, `PrimarySymbols[]` (Tier 3 deterministic engine, imported as-is) | 001/003 |
| `internal/mx/` | `SpecAssociator` output (already bridged into the graph by M0 REQ-NS-005) | MX-ASSOCIATION |
| `.moai/project/codemaps/dependencies.md` | dep-graph (Tier 1 `module_tree.json` scaffold seed) | `/moai codemaps` |
| `.moai/decisions/` | existing ADR-shaped markdown (Tier 2 formalizes this home) | (pre-existing) |

### §A.2 What M4 produces (the 4-tier overlays)

1. `.moai/project/navigator/tiers.json` — the tier-enrichment OVERLAY (REQ-NS3-018). Consumers JOIN with `nav-graph.json`. Carries the enriched node fields (contract/blueprint nodes; decision `adr_path`+`superseded_by`; symbol signature/declaration/references/narrative_path) + the new edge types (`module-edge`, `owns-edge`, `superseded_by`).
2. `.moai/project/blueprint/module_tree.json` — Tier 1 authored module tree (REQ-NS3-004).
3. `.moai/project/blueprint/<module>/overview.md` — Tier 1 Kiro 7-section per-module overview (REQ-NS3-005).
4. `.moai/project/blueprint/contracts.yaml` — Tier 0 contract registry (REQ-NS3-003).
5. `.moai/decisions/<dec-id>.md` — Tier 2 ADRs (NEW only; existing bodies immutable per REQ-NS3-009).
6. `.moai/project/navigator/symbols/<symbol>.md` + `metadata.json` — Tier 3 per-symbol narrative (REQ-NS3-012).

### §A.3 New Go package

`internal/navigator/tiers/` — the 4-tier enrichment engine. Imports `internal/navigator/astx/` as-is, reads `nav-graph.json` (M0) read-only, emits `tiers.json` + the author-facing artifact drafts. Wired as a Hidden cobra subcommand mirroring M0's `navigator-sync` (`internal/cli/navigator_sync.go`).

## §B. Known Issues

1. **SCIP indexer absence (design report 리스크 #4)** — no SCIP indexer is vendored; Go uses astx. M4 ships the astx/Go deterministic path and treats SCIP as forward-compatible per-language rollout (REQ-NS3-013). NOT a blocker — resolved in §C / design.md §1.D1.
2. **`.moai/decisions/` pre-existing content shape** — the two existing files (`lsp-client-choice.md`, `skill-rename-map.yaml`) are ADR-shaped but pre-date the formal ADR scheme. M4 MUST NOT edit them in place (immutability, REQ-NS3-009); they are grandfathered as-is. New ADRs follow the formal four-field template. Resolved in §C / design.md §1.D2.
3. **Blueprint authoring loop ambiguity** — "authored not generated" leaves the exact authoring loop unspecified. Resolved via the scaffold-then-refine loop (§C / design.md §1.D3): `/moai codemaps` dep-graph → draft `module_tree.json` → human/agent refines → NOT auto-replace on subsequent runs (deterministic layer only re-scaffolds when explicitly invoked with `--rescaffold`).
4. **Contract drift enforcement surface for moai-adk-go** — the design report names "Tauri allowlist / OpenAPI / schema" as examples, none native to a Go CLI. Resolved by enumerating the three Go-native contract surfaces (nav-graph schema, hook JSON schemas, CLI flag schemas) + a template-distributed `contracts.yaml` registry (§C / design.md §1.D4).
5. **Success-metric measurement (REQ-NS3-022)** — the "≥40% reads-reduction" metric requires a fixture that simulates LLM orientation with vs without blueprint pre-read. This is the most novel AC and the one most at risk of narrative-claim substitution. Resolved by naming the exact fixture procedure in acceptance.md §D.AC-NS3-022 (a static corpus + a deterministic read-count comparator), NOT a narrative.

## §C. Pre-flight Decisions (each resolved, default proposed)

**D1 — SCIP integration scope → astx-only at M4, SCIP forward-compat** — RESOLVED — see design.md §1.D1.

**D2 — ADR home + grandfathering → `.moai/decisions/`, existing files immutable** — RESOLVED — see design.md §1.D2.

**D3 — Blueprint authoring loop → scaffold-then-refine, no auto-replace** — RESOLVED — see design.md §1.D3.

**D4 — Contract surfaces for moai-adk-go → 3 enumerated + registry** — RESOLVED — see design.md §1.D4.

**D5 — Overlay vs overwrite → `tiers.json` overlay, `nav-graph.json` untouched** — RESOLVED — see design.md §1.D5.

## §D. Constraints (restated from spec.md §D, operationalized)

- **Determinism** — deterministic-layer artifacts sort everything before serialization (nodes by `(entity_type, identifier)`, edges by `(edge_type, source_node, target_node, source_path, line_number)`), `map[string]V` iterates sorted keys, NO `time.Now()` in the deterministic path. This is how REQ-NS3-019's byte-identical re-run is met.
- **Fail-open** — all tier errors log to `.moai/logs/navigator-sync.log` and are swallowed; the cobra `RunE` returns nil unconditionally (mirrors M0's `join.go` / 003's `navigator_enrich.go`).
- **No wall-clock (deterministic layer)** — `captured_at` is `git log -1 --format=%cI` of HEAD. The LLM-narrative layer carries its own `last_updated_commit` (also a git SHA, not wall-clock) but is NOT byte-identical across runs (narrative is authored).
- **Additive schema** — `tiers.json` carries new node types (`contract`, `blueprint`) and new edge types (`module-edge`, `owns-edge`, `superseded_by`) WITHOUT redefining M0's existing types. A consumer that does not understand the new types ignores them (forward-compat per `nav-tokens.md`).
- **Hidden CLI** — the tier step is a Hidden cobra subcommand (`navigator-tiers`), wired into `/moai project` AFTER M0's `navigator-sync` join step. NOT a new top-level `moai` subcommand.
- **Template-First touch list (REQ-NS3-021)** — author-facing surfaces authored under `internal/template/templates/` first, then `make build`, then local mirror:
  - `internal/template/templates/.moai/project/blueprint/overview.template.md` (Kiro 7-section)
  - `internal/template/templates/.moai/project/blueprint/module_tree.schema.json`
  - `internal/template/templates/.moai/project/blueprint/contracts.yaml` (registry format doc, neutral)
  - `internal/template/templates/.moai/decisions/adr-template.md` (four canonical fields)
  - `internal/template/templates/.moai/project/navigator/symbols/narrative.template.md`
  - Local mirrors: `.moai/project/blueprint/`, `.moai/decisions/`, `.moai/project/navigator/symbols/`.

## §E. Self-Verification (plan-phase, the §E.1 audit-ready signal)

This plan-phase artifact set is audit-ready when:

- [ ] All 22 REQs (REQ-NS3-001..022) trace to at least one AC in acceptance.md (§F matrix confirms 1:1).
- [x] All 5 pre-flight decisions (D1–D5) RESOLVED — pointers in §C reference design.md §1.D{1..5}; all clarification items resolved, none remain in plan.md or research.md (MP-7 firewall cleared).
- [ ] The non-overlap invariants (REQ-NS3-016/017/018) are machine-testable via the grep-pair protocol (acceptance.md §A.1) — consumer-only on M0/M1, no predecessor surface, no LSEL surface, write-surface isolation.
- [ ] The blueprint-first stance (REQ-NS3-006) is negatively testable: NO drift-fail CI gate exists (a grep asserting the absence of a blueprint-drift build failure).
- [ ] The success metric (REQ-NS3-022) names a mechanical fixture procedure, not a narrative.
- [ ] Every verification claim in acceptance.md names the command + the observable output form, per `verification-claim-integrity.md` §2 attribution.

## §F. Milestones (ordered by decision-reversibility — highest-change-likelihood first)

### M4.1 — Schema + tier-overlay shape + pre-flight decisions (most reversible, lead)

**Decision**: the `tiers.json` overlay schema, the `contract`/`blueprint` node shapes, the new edge types, the 5 pre-flight decisions. These are the decisions most likely to change; they lead so human review sees them first.

- Deliverable: `internal/navigator/tiers/schema.go` — Go types for the overlay (ContractNode, BlueprintNode, enriched DecisionNode/SymbolNode, ModuleEdge/OwnsEdge/SupersededByEdge). JSON tags stable, additive.
- Deliverable: design.md §2 (the `tiers.json` schema JSON example) recording the resolved shapes.
- Exit gate: schema types compile; design.md carries the 5 resolved decisions; no open clarification markers in plan.md.

### M4.2 — Non-overlap invariant tests (load-bearing guard, early)

**Why early**: the non-overlap contracts (REQ-NS3-016/017/018) are load-bearing; getting them wrong late is catastrophic. Establish the invariant tests BEFORE writing the tier engine, so every later step runs against them (mirrors M0's M0.2).

- Deliverable: `internal/navigator/tiers/nonoverlap_test.go` — grep-pair tests (per `feedback_deadfield_grep_pair_protocol.md`): (a) consumer-only on M0/M1 producer paths, (b) no predecessor write surface, (c) no LSEL surface, (d) write-surface isolation (only the 6 named paths), (e) `nav-graph.json` never overwritten (overlay-only).
- Deliverable: source-comment hygiene — production source in `internal/navigator/tiers/` MUST NOT literally name forbidden LSEL/predecessor paths (parallel to M0 design.md §5 source-discipline rule).
- Exit gate: tests RED initially (no implementation yet), then GREEN once M4.3–M4.6 land.

### M4.3 — Tier 0 Contract + Tier 2 ADR (decision-layer, medium reversibility)

**Why ahead of Tier 1/3**: Tier 0 and Tier 2 are the most decision-shaped (contract registry format, ADR supersede semantics) — human review should see them before the heavier Tier 1/3 engines.

- Deliverable: `internal/navigator/tiers/contract.go` — contract-node emission from `contracts.yaml` registry + the 3 moai-adk-go enumerated surfaces (REQ-NS3-001/003); `drift.go` — the build-time drift check (REQ-NS3-002), fail-open to the graph.
- Deliverable: `internal/navigator/tiers/adr.go` — ADR resolution (`@NAV:DEC-<id>` → `.moai/decisions/<dec-id>.md`), four-field parse, supersede chain (REQ-NS3-008/009), decision-node enrichment (REQ-NS3-010).
- Deliverable: `internal/navigator/tiers/adr_test.go` — supersede-immutability test (prior body unedited; only Status flips).

### M4.4 — Tier 1 Blueprint (authored layer, medium reversibility)

- Deliverable: `internal/navigator/tiers/blueprint.go` — `module_tree.json` scaffold from `/moai codemaps` `dependencies.md` (read-only) + authored-merge (does NOT auto-replace; `--rescaffold` opt-in) (REQ-NS3-004); overview.md template instantiation per module (REQ-NS3-005).
- Deliverable: `internal/navigator/tiers/blueprint_test.go` — assert scaffold-then-refine does NOT overwrite a human-edited `module_tree.json` on a plain run (the blueprint-first stance, REQ-NS3-006, negatively tested).
- Deliverable: blueprint node + `module-edge`/`owns-edge` emission into `tiers.json` (REQ-NS3-007).

### M4.5 — Tier 3 Symbol 2-tier (deterministic engine extension, lower reversibility)

- Deliverable: `internal/navigator/tiers/symbol_struct.go` — per-symbol structured record (signature + declaration + references) by importing `internal/navigator/astx/` as-is and extending its output additively (REQ-NS3-011). SCIP forward-compat stub (REQ-NS3-013): a language→indexer dispatch table where Go = astx and other languages = "not configured, degrade".
- Deliverable: `internal/navigator/tiers/symbol_narrative.go` — per-symbol narrative slot + `metadata.json` `last_updated_commit` + the CodeWiki-style `--update`/`--compare-to` gate (REQ-NS3-012). Narrative is drafted ONLY when the deterministic record changed since `last_updated_commit`.
- Deliverable: symbol-node enrichment into `tiers.json` (REQ-NS3-014) + the 2-tier separability test (REQ-NS3-015): deterministic layer produced with the LLM narrative path stubbed out.

### M4.6 — Overlay join + Hidden CLI + `/moai project` wiring + provenance/fail-open

- Deliverable: `internal/navigator/tiers/overlay.go` — the `tiers.json` emission with provenance (REQ-NS3-019) + atomic-write (mirrors M0 `write.go`) + fail-open (REQ-NS3-020). Join semantics: `tiers.json` is self-contained (carries its own provenance + the enriched nodes/edges); consumers JOIN with `nav-graph.json` on `(entity_type, identifier)`.
- Deliverable: `internal/cli/navigator_tiers.go` — Hidden cobra subcommand `navigator-tiers` mirroring `navigator_sync.go`. Registered in `internal/cli/root.go`. Invoked as a sibling step inside `/moai project` AFTER M0's `navigator-sync`.
- Deliverable: characterization test — re-run the tier enrichment twice on the same HEAD, assert byte-identical `tiers.json` (deterministic layer; REQ-NS3-019).

### M4.7 — Template-first documentation + success-metric fixture + verification close-out (final, mechanical)

- Deliverable: the 5 template-first author-facing surfaces listed in §D (overview.template.md, module_tree.schema.json, contracts.yaml, adr-template.md, narrative.template.md) under `internal/template/templates/`, all neutrality-guard-clean (no SPEC IDs, no REQ tokens, no internal dates per CLAUDE.local.md §25).
- `make build` regeneration + `internal/template/internal_content_leak_test.go` neutrality guard green (REQ-NS3-021).
- Deliverable: the REQ-NS3-022 success-metric fixture under `internal/navigator/tiers/metrics_fixture_test.go` — a static orientation-task corpus + a deterministic read-count comparator (with-blueprint vs without-blueprint), emitting the percentage as observed output. Acceptance reads the observed percentage, NOT a hardcoded constant.
- Deliverable: full `go test ./internal/navigator/tiers/...` green; `go test ./internal/cli/ -run NavigatorTiers` green; `go test ./...` green; `golangci-lint` clean; cross-platform `GOOS=windows GOARCH=amd64 go build ./...` exit 0.

## §G. Anti-Patterns (specific to this SPEC)

- **AP-NS3-001 — Auto-generate-and-replace the blueprint**: forbidden by REQ-NS3-004/006. The blueprint is authored; `module_tree.json` is scaffolded ONCE and refined. A plain `/moai project` run SHALL NOT overwrite a human-edited blueprint.
- **AP-NS3-002 — Make blueprint drift a build failure**: forbidden by REQ-NS3-006. Only Tier 0 Contract drift is build-enforced (REQ-NS3-002); conflating the two is the spec-as-source trap.
- **AP-NS3-003 — Overwrite `nav-graph.json`**: forbidden by REQ-NS3-018. M4 emits `tiers.json` as an OVERLAY; consumers JOIN.
- **AP-NS3-004 — Modify M0/M1 producers or `internal/mx/`**: forbidden by REQ-NS3-016. Consumer-only, bridge-not-absorb (carries forward M0 REQ-NS-005).
- **AP-NS3-005 — Edit an existing ADR body**: forbidden by REQ-NS3-009. Supersede creates a NEW ADR; the prior body is immutable (only `Status:` flips).
- **AP-NS3-006 — Wall-clock timestamp in the deterministic layer**: forbidden by REQ-NS3-019. The LLM-narrative layer carries `last_updated_commit` (a SHA, not wall-clock).
- **AP-NS3-007 — Touch LSEL or predecessor surfaces**: forbidden by REQ-NS3-017. Use grep-pair tests, not bare grep (the `feedback_deadfield_grep_pair_protocol.md` lesson; M0 design.md §5 source-discipline rule).
- **AP-NS3-008 — Narrative-claim the success metric**: forbidden by REQ-NS3-022 + `verification-claim-integrity.md` §1.1 surface 3. The ≥40% figure MUST be the observed output of the named fixture, not an assertion.
- **AP-NS3-009 — Template doc authored in local `.moai/` first**: forbidden by CLAUDE.local.md §2. Template-first means `internal/template/templates/` first, then `make build`, then verify local received it.
- **AP-NS3-010 — Block on SCIP absence**: forbidden by REQ-NS3-013. SCIP is forward-compat per-language; M4 ships astx/Go and degrades gracefully elsewhere.

## §H. Cross-References

- spec.md §B.2 — non-overlap carry-forward table.
- spec.md §C — the 22 REQs this plan operationalizes.
- acceptance.md §A.1 — grep-pair methodology for non-overlap ACs.
- acceptance.md §D.AC-NS3-022 — the success-metric fixture procedure.
- design.md §1 — the 5 resolved pre-flight decisions (D1–D5).
- design.md §2 — the `tiers.json` overlay schema JSON example.
- design.md §3 — architecture / data flow.
- design.md §4 — blueprint-first vs spec-as-source reasoning (Fowler reversible optimum).
- design.md §5 — Tier 3 2-tier split reasoning (CodeWiki hierarchical decomposition).
- research.md §1 — design-report asset inventory verified against current main.
- research.md §3 — non-overlap verification result (cite REQ-PN-016, REQ-NA-011, REQ-NT-018/019, REQ-NS-013/015/016, REQ-NS2-005/008).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 + §2 — success-metric attribution.
- `feedback_deadfield_grep_pair_protocol.md` (memory) — grep-pair methodology for non-overlap ACs.
- `feedback_template_first_mirror_runphase.md` (memory) — template-first is a run-phase obligation; this plan seeds it in M4.7.
