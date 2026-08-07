# SPEC-NAVIGATOR-SYNC-003 — Design

> Tier L design artifact. Records the 5 resolved pre-flight decisions, the `tiers.json` overlay schema, the architecture, and the load-bearing stance reasoning (blueprint-first vs spec-as-source; Tier 3 2-tier split).
> Cross-references spec.md §C (REQs), plan.md §C (pre-flight decisions), acceptance.md §B (ACs).

## §1. Resolved Pre-flight Decisions (was potential [NEEDS CLARIFICATION] in plan.md §B)

### D1 — SCIP integration scope → astx-only at M4, SCIP forward-compat

**Decision**: M4 ships the astx/Go deterministic structure path. SCIP is forward-compatible, per-language gradual rollout; it is NOT a blocking dependency for M4 acceptance.

**Rationale**: the design report (§리스크 #4) explicitly names "SCIP 인덱서 미성숙 → 완화: astx+빌드 그래프 먼저, SCIP은 언어별 점진." Vendoring a SCIP indexer now would couple M4 to an immature tool and bloat the Go binary with language-specific indexers the project does not yet need. The astx engine (`internal/navigator/astx/`, `astx.go:22` `type Symbol struct`, `enrich.go` `EnrichRows`) is the existing deterministic Go symbol source; M4 extends its output additively (per-symbol signature + declaration + references) rather than replacing it.

**Implementation**: `internal/navigator/tiers/symbol_struct.go` carries a `languageIndexers` dispatch table: Go → astx-extended; other languages → `errIndexerNotConfigured` (degrade to zero per-symbol records, debug log). A future SPEC plugs SCIP indexers into the same dispatch table per language without touching M4's Tier 3 emission path.

### D2 — ADR home + grandfathering → `.moai/decisions/`, existing files immutable

**Decision**: `.moai/decisions/` (which already exists with two ADR-shaped files: `lsp-client-choice.md`, `skill-rename-map.yaml`) is the ADR home. The two pre-existing files are GRANDFATHERED — M4 does NOT rewrite them into the formal four-field template; they remain as-is. New ADRs follow the formal template.

**Rationale**: immutability (REQ-NS3-009) is load-bearing for the ADR pattern — an ADR that can be silently rewritten loses its audit-trail value. Rewriting the two grandfathered files to fit the formal template would violate immutability for zero structural gain (they already carry Decision Date / Status / Context / Decision / Consequences in substance). M4 treats them as first-class ADRs whose shape predates the formal template; the `@NAV:DEC-<id>` resolution matches them by filename.

**Implementation**: `internal/navigator/tiers/adr.go` resolves `@NAV:DEC-<id>` → `.moai/decisions/<id>.md` by filename match (case-insensitive). The four-field parse is BEST-EFFORT: a file missing a field degrades to "field empty" rather than failing. The grandfathered files parse successfully because they carry the substance; their exact heading casing does not need to match the template verbatim.

### D3 — Blueprint authoring loop → scaffold-then-refine, no auto-replace

**Decision**: `module_tree.json` is scaffolded ONCE from `/moai codemaps` `dependencies.md` (read-only), then refined by human or agent. A plain `/moai project` run does NOT overwrite a human-edited `module_tree.json`; re-scaffolding requires an explicit `--rescaffold` flag.

**Rationale**: "authored not generated" (design report Q1) is the blueprint-first stance. Auto-generate-and-replace is the spec-as-source hazard's adjacent failure mode: it makes the blueprint a generated artifact that can never diverge from the code, which collapses Tier 1 back into a regenerated snapshot (the exact failure mode the design report's §현 상태 진诊断 names: "regenerate-and-replace, 역방향 0"). The scaffold-then-refine loop keeps the blueprint AUTHORED (human/agent judgment shapes it) while bootstrapping it from the deterministic dep-graph so the human is not staring at a blank file.

**Implementation**: `internal/navigator/tiers/blueprint.go` writes `module_tree.json` only if it does NOT exist (or `--rescaffold` is passed). When it exists, the run reads it as-is and emits blueprint nodes from the authored content. The deterministic layer (the scaffold) is separable from the authored layer (the refinement), satisfying REQ-NS3-015.

### D4 — Contract surfaces for moai-adk-go → 3 enumerated + registry

**Decision**: moai-adk-go's three Tier 0 contract surfaces are (a) the `nav-graph.json` schema (already build-enforced via M0's byte-stable characterization test), (b) the hook input JSON schemas consumed by `internal/hook/`, and (c) the cobra CLI command + flag schemas at `internal/cli/`. Template-distributed users declare their own surfaces via `.moai/project/blueprint/contracts.yaml`.

**Rationale**: the design report (line 28) names "Tauri allowlist / OpenAPI / schema" as Tier 0 examples — none is native to a Go CLI. Translating the contract-first principle to moai-adk-go means naming the surfaces where a declared shape is validated against reality at build time. The three enumerated surfaces each have an existing validator (M0's byte-stable test; the hook JSON unmarshaling path; cobra's flag registration). A template-distributed user's contract surfaces depend on their stack (a Tauri app really does have an allowlist; an API service really does have an OpenAPI doc), so the registry is user-declared.

**Implementation**: `internal/navigator/tiers/contract.go` ships the three moai-adk-go surfaces as built-in `ContractSurface` entries + loads `contracts.yaml` for user-declared surfaces. Each entry names a `validator_command` (the build-time drift check, REQ-NS3-002). An empty registry yields zero contract nodes (graceful).

### D5 — Overlay vs overwrite → `tiers.json` overlay, `nav-graph.json` untouched

**Decision**: M4 emits `.moai/project/navigator/tiers.json` as an OVERLAY. Consumers JOIN `tiers.json` with M0's `nav-graph.json` on `(entity_type, identifier)`. M4 NEVER overwrites `nav-graph.json`.

**Rationale**: `nav-graph.json` is M0's artifact (REQ-NS-006); M4 modifying it would violate consumer-only (REQ-NS3-016) and would couple M4's emit cycle to M0's emit cycle (a regeneration race). The overlay pattern keeps the two artifacts independently produced and independently byte-stable; a consumer that does not understand the tier fields ignores them (forward-compat per `nav-tokens.md`). The JOIN key `(entity_type, identifier)` is M0's composite primary key (design.md §2 of M0), so the join is deterministic.

**Implementation**: `internal/navigator/tiers/overlay.go` emits `tiers.json` with its own provenance block (REQ-NS3-019). The `tiers.json` schema is self-contained — it carries enriched node fields keyed by `(entity_type, identifier)` + the new edge types. A consumer JOIN is a simple merge; M4 does not implement the consumer (M1 Detect and future M2 Route are the consumers).

## §2. `tiers.json` Overlay Schema (normative)

```json
{
  "provenance": {
    "extract_commit_sha": "<git rev-parse HEAD>",
    "captured_at": "<git log -1 --format=%cI of HEAD>"
  },
  "tier0_contracts": [
    {
      "identifier": "nav-graph-schema",
      "contract_kind": "schema",
      "contract_path": ".moai/project/navigator/nav-graph.json",
      "validator_command": "go test -run TestRun_AC009_ByteIdenticalReRun ./internal/navigator/sync/...",
      "drift_status": "aligned"
    }
  ],
  "tier1_blueprints": [
    {
      "identifier": "internal/navigator/sync",
      "display_name": "Navigator Sync (M0 graph-join)",
      "layer": "infrastructure",
      "responsibility": "SSOT binding-token trio + graph-join",
      "depends_on": ["internal/navigator/astx", "internal/mx"],
      "overview_path": ".moai/project/blueprint/internal/navigator/sync/overview.md"
    }
  ],
  "tier2_decisions": [
    {
      "identifier": "AUTH-STRATEGY",
      "adr_path": ".moai/decisions/AUTH-STRATEGY.md",
      "superseded_by": ""
    },
    {
      "identifier": "AUTH-STRATEGY-V2",
      "adr_path": ".moai/decisions/AUTH-STRATEGY-V2.md",
      "superseded_by": "",
      "supersedes": "AUTH-STRATEGY"
    }
  ],
  "tier3_symbols": [
    {
      "identifier": "internal/navigator/sync.Join",
      "signature": "func Join(...) (Graph, error)",
      "declaration_path": "internal/navigator/sync/join.go",
      "declaration_line": 212,
      "references": [
        { "path": "internal/cli/navigator_sync.go", "line": 88 }
      ],
      "narrative_path": ".moai/project/navigator/symbols/internal_navigator_sync_Join.md"
    }
  ],
  "tier_edges": [
    { "edge_type": "module-edge", "source_node": "blueprint:internal/navigator/sync", "target_node": "blueprint:internal/navigator/astx" },
    { "edge_type": "owns-edge", "source_node": "blueprint:internal/navigator/sync", "target_node": "symbol:internal/navigator/sync.Join" },
    { "edge_type": "superseded_by", "source_node": "decision:AUTH-STRATEGY", "target_node": "decision:AUTH-STRATEGY-V2" }
  ]
}
```

**Join key** — each enrichment record's `identifier` joins to the corresponding `nav-graph.json` node's `identifier` within the same `entity_type`. A record whose `identifier` has no `nav-graph.json` counterpart is still emitted (M4 may surface a symbol the M0 graph missed); a `nav-graph.json` node with no `tiers.json` enrichment is consumed as-is.

**Forward-compatibility** — fields are additive only. Later milestones may add tier sections or fields; existing fields keep their names and shapes. A consumer that does not understand a tier section ignores it.

## §3. Architecture (data flow)

```
                                  ┌──────────────────────────────────────┐
                                  │   internal/navigator/tiers/          │
                                  │   (NEW — this SPEC)                  │
                                  │                                      │
  M0 nav-graph.json ─────────────►│  READ-ONLY consumer (REQ-NS3-016)    │
  (decision/spec/symbol nodes)    │  → join key (entity_type, identifier) │
                                  │                                      │
  /moai codemaps dependencies.md ►│  Tier 1 scaffold seed (read-only)     │
                                  │  → module_tree.json draft             │
                                  │                                      │
  .moai/project/blueprint/ ──────►│  Tier 1 AUTHORED module_tree.json    │
  module_tree.json (authored)     │  → blueprint nodes + module-edges     │
                                  │                                      │
  .moai/project/blueprint/ ──────►│  Tier 0 contract registry            │
  contracts.yaml (registry)       │  → contract nodes + drift check       │
                                  │                                      │
  .moai/decisions/<id>.md ───────►│  Tier 2 ADR resolution               │
  (ADR files, immutable)          │  → adr_path + superseded_by           │
                                  │                                      │
  internal/navigator/astx/ ──────►│  Tier 3 deterministic structure       │
  (imported as-is, extended)      │  → signature + declaration + refs     │
                                  │                                      │
  .moai/project/navigator/ ──────►│  Tier 3 LLM narrative (opt-in)        │
  symbols/<symbol>.md + metadata  │  → narrative_path (2-tier separable)  │
                                  │                                      │
                                  │  Provenance (git baseline, no clock)  │
                                  │  Atomic write (renames .tmp)          │
                                  │  Fail-open capability gate            │
                                  │                                      │
                                  └─────────────────┬────────────────────┘
                                                    │
                                                    ▼
                                  .moai/project/navigator/tiers.json
                                  (OVERLAY — consumers JOIN with nav-graph.json;
                                   nav-graph.json itself is NEVER overwritten)
```

## §4. Blueprint-first vs spec-as-source (the load-bearing stance, design report Q1)

The design report (§세 물음 답 Q1) positions Tier 1 Blueprint as "저술" (authored), explicitly NOT "생성" (generated). It cites Fowler's spec-anchored definition as the reversible optimum between two failure modes:

| Stance | What it is | Failure mode | M4 position |
|--------|-----------|--------------|-------------|
| **spec-first (deletion)** | Write the spec, code it, throw the spec away | The spec was a one-time scaffold; orientation knowledge is lost after coding | REJECTED — M4's blueprint persists |
| **spec-as-source (MDD)** | The spec GENERATES code; spec↔code drift is a build failure | The spec becomes a maintenance burden; any code change forces a spec change; the spec can never be wrong, only incomplete | REJECTED (REQ-NS3-006) — M4's blueprint does NOT generate code; drift is debt, not failure |
| **spec-anchored (Fowler)** | The spec documents intended architecture for orientation; reversible; code is the truth on divergence | (the reversible optimum — no catastrophic failure mode) | ADOPTED |

**Why this matters operationally**: a spec-as-source MDD stance would force a CI gate that fails whenever the blueprint and code diverge. On a rapidly-evolving codebase, that gate would fire constantly (every refactor, every rename) and the team would either (a) abandon the blueprint or (b) spend the refactor budget keeping the blueprint in lockstep — both defeat the purpose. The spec-anchored stance lets the blueprint drift and surfaces the drift as a documentation-debt signal (via M1 Detect, REQ-NS2-003) that a human triages, NOT a build failure that a CI enforces.

**Kiro Design 7-section template** (design report Q1): the per-module `overview.md` carries seven sections — Component Architecture, Data Flow, Data Model, Error Handling, Test Strategy, Implementation Approach, Migration. These are the seven surfaces an orienting LLM (or human) needs to understand a module's intent without reading its full source. The template is the authored layer's structure; the prose within each section is the human/agent-authored content.

## §5. Tier 3 2-tier split (design report Q2)

The design report (§세 물음 답 Q2) splits Tier 3 into two layers:

| Layer | Produced by | Reproducible? | M4 surface |
|-------|-------------|---------------|------------|
| **Deterministic structure** | astx (Go) / SCIP (other languages, forward-compat) | YES — byte-identical on re-run against the same HEAD | signature, declaration_path + line, references (callers) |
| **LLM narrative** | LLM-drafted, human/agent-approved, CodeWiki-style `--update`/`--compare-to` | NO — narrative is authored; `metadata.json` tracks `last_updated_commit` | docstring + call-context prose |

**Why two tiers**: the design report (Q2) cites "순수 LLM은 시스템 언어에서 약함 (C/C++ −3.15%)" — a pure-LLM symbol layer is unreliable on systems languages because the LLM hallucinates signatures and call graphs it cannot mechanically verify. The deterministic structure layer (astx/SCIP) is the ground truth the LLM narrative layer ANNOTATES, never replaces. The two tiers are separable (REQ-NS3-015) so the deterministic layer is produced and consumed WITHOUT an LLM in the loop — an LLM outage never blocks orientation.

**CodeWiki `--update`/`--compare-to` incremental technique** (design report Q2): the narrative is re-drafted ONLY when the symbol's deterministic record changed since `metadata.json`'s `last_updated_commit`. This bounds the LLM cost (only changed symbols are re-drafted) and avoids churn (unchanged symbols keep their approved narrative).

## §6. Non-overlap invariant test design

Carries forward M0 design.md §5. Each non-overlap AC (AC-NS3-016/017/018) uses two grep lenses:

- **Lens 1 — source**: production source under `internal/navigator/tiers/` + `internal/cli/navigator_tiers.go` MUST NOT literally name forbidden paths (`capability-map.md`, `audit-report`, `capability-symbols`, `lessons-inbox`, `state/lsel`, `memory/feedback_`, `hns-lsel`). Source-comment hygiene is the defensive obligation that keeps Lens 1 mechanically sound.
- **Lens 2 — runtime**: a temp-dir fixture snapshots the tree before/after, diffs, asserts the only NEW paths are within the 6 allowed write surfaces.

A bare `lsel` substring is FORBIDDEN as a grep pattern (it would match the SPEC's own non-overlap test file). The consumer-only AC (AC-NS3-016) additionally uses `git diff <commit-range> -- internal/navigator/sync/ internal/navigator/detect/ internal/hook/...` (Lens: the M0/M1 producer paths are untouched in the run-phase commit range).

## §7. Performance envelope

- **Tier 0 contract drift check**: bounded by the validator commands; the three moai-adk-go surfaces reuse existing tests (fast). User-declared validators run at their own cost.
- **Tier 1 blueprint**: scaffold reads `/moai codemaps` `dependencies.md` (already produced, read-only). Authored `module_tree.json` read is O(modules).
- **Tier 2 ADR**: O(ADR files) — `.moai/decisions/` is small (2 files today; expected to grow slowly).
- **Tier 3 deterministic (astx)**: reuses 003's bounded scan; O(Go source files containing the target symbols). The references list is capped at N (configurable) to bound output size.
- **Tier 3 LLM narrative**: opt-in; runs only on changed symbols (CodeWiki `--compare-to` gate). NOT run on every `/moai project` invocation.
- **Overlay emission**: single JSON file, typically < 200KB for a project this size.

No per-edit performance AC (M4 is not a PostToolUse hook); the tier enrichment runs as a `/moai project` sibling step where the existing timeout envelope applies.

## §8. Open questions for later milestones (forward-looking, NOT M4 blockers)

- **Should `tiers.json` be consumed by M1 Detect at edit time?** M4 produces the overlay; a future enhancement could have M1 Detect consult the blueprint layer to answer "did this edit touch a module whose blueprint claims X?" — M4 does not implement the consumer.
- **Should the Tier 3 narrative be auto-committed after human approval, or only ever hand-committed?** M4 leaves the approval gate in place (REQ-NS3-012); the commit policy is a downstream decision.
- **Should the ADR supersede chain support partial superscession (supersedes one field, not the whole decision)?** M4 ships whole-decision superscession only; partial is a forward milestone.
- **Should the contract drift check distinguish "drifted" from "validator unavailable"?** M4 collapses both into `drift_status: drifted` for simplicity; a finer taxonomy is a forward milestone.
