---
id: SPEC-NAVIGATOR-SYNC-003
title: "Navigator Sync (BAS M4) — 4-tier addressable map (Contract / Blueprint / ADR / Symbol extension over the M0 graph)"
version: "0.1.0"
status: completed
created: 2026-08-06
updated: 2026-08-07
author: manager-spec
priority: P1
phase: "v3.3 target"
module: navigator-sync
lifecycle: spec-anchored
tier: L
era: V3R6
tags: "navigator, sync, 4-tier, contract, blueprint, adr, symbol, scip, bas-epic"
related_specs: [SPEC-NAVIGATOR-SYNC-001, SPEC-NAVIGATOR-SYNC-002, SPEC-PROJECT-NAVIGATOR-001, SPEC-PROJECT-NAVIGATOR-002, SPEC-PROJECT-NAVIGATOR-003]
depends_on: [SPEC-NAVIGATOR-SYNC-001]
---

# SPEC-NAVIGATOR-SYNC-003 — Navigator Sync (BAS Epic M4) — 4-tier Addressable Map

## HISTORY

- 2026-08-06 (initial draft) — BAS (Blueprint-Anchored Synchronization) Epic M4, plan-phase. M0 (`SPEC-NAVIGATOR-SYNC-001`, PR #1375 squash `f4bd58acbd`, MERGED 2026-08-05, `status: completed`) delivered the SSOT binding-token trio (`@NAV:DEC-<id>` / `@NAV:SYM:<symbol>` / `@MX:SPEC:<id>` bridged not absorbed per REQ-NS-005) and the graph-join schema layer at `internal/navigator/sync/` producing `.moai/project/navigator/nav-graph.json` (nodes: decision/spec/symbol; edges: dec-edge/spec-edge/sym-edge). M1 (`SPEC-NAVIGATOR-SYNC-002`, PR #1379 squash `304907b6d`, MERGED 2026-08-06, `status: completed`) delivered the Falconer Detect element (PostToolUse changed-path → affected-graph-rows mapping at `internal/navigator/detect/` + `internal/hook/navigator_detect*.go`, writing `.moai/state/navigator-detect/<session-id>.jsonl`). M4 is part (B) of the BAS redesign — the 4-tier addressable map — authored from the design report `.moai/reports/navigator-redesign-bas-20260805.{md,html}` §제안 BAS (B) line 28, §세 물음 답 Q1/Q2 line 42-43, §마일스톤 M4 line 51, §리스크 line 54+ (session 3db3f943). M4 EXTENDS the M0 graph with four tier-specific enrichment overlays — Tier 0 Contract (build-enforced, drift-immune), Tier 1 Blueprint (authored module_tree + overview, LLM entry map), Tier 2 Decision (ADR, immutable + supersede), Tier 3 Symbol (astx→SCIP, 2-tier deterministic + LLM narrative) — WITHOUT modifying M0/M1 producers. M4 depends on M0 only (`depends_on: [SPEC-NAVIGATOR-SYNC-001]`, `status: completed`); it is parallel to M1/M2/M3 (no cross-dependency per the design report §마일스톤).
- **SPEC-ID re-numbering note (transparency)**: M0's §E projected the Epic's SPEC IDs as 002=M1, 003=M2 (Route), 004=M3 (Fix), 005=M4, 006=M5. In practice the IDs are assigned in AUTHORING order, not milestone order: 002=M1 (authored 2nd), and this SPEC is 003=M4 (authored 3rd), because M4 was prioritized ahead of M2 (Route) and M3 (Fix). M2 and M3 will receive 004/005 (or later) when authored. This re-numbering is recorded here so a reader cross-referencing M0's projection is not confused; it changes no technical content. Confirmed `SPEC-NAVIGATOR-SYNC-003` was free (no prior directory).

## §A. User Story

**As a** Claude Code session (and the human orchestrator) orienting in a large, evolving repo,
**I want** the Navigator graph to carry four tiers of depth — a build-enforced contract layer, an authored blueprint entry map, an immutable decision (ADR) layer, and a per-symbol structure + narrative layer —
**so that** I can orient precisely (the 오라버니 insight: "the LLM reads partially to save tokens and misses the big picture → it needs a blueprint") WITHOUT the Navigator regenerating the whole map on every change, and WITHOUT a single documentation pattern failing across contract-heavy, design-heavy, decision-heavy, and symbol-heavy surfaces alike.

The design report's core thesis (§핵심 테제): "the map's value comes not from adding tokens but from reducing tokens while preserving the big picture." Today the M0 graph is addressable but SHALLOW — a decision node is an identifier with no ADR body; a symbol node is a name with no signature/docstring/callers; there is no contract layer (build-enforced schemas) and no blueprint layer (authored module entry map). M4 deepens each node type into a tier. The four tiers are a PORTFOLIO (design report §세 물음 답 Q3): no single pattern suffices — Tier 0 is contract-first (build-enforced), Tiers 1 and 3 are living-doc (detect→route→fix), Tier 2 is ADR (replace-not-edit).

**Fowler spec-anchored stance (load-bearing, design report Q1)**: Tier 1 Blueprint is "authored," NOT "generated." It is spec-anchored — a reversible optimum between spec-first (deletion, where the spec is thrown away after coding) and spec-as-source MDD (the trap where the spec generates code and drift is a build failure). The blueprint documents the INTENDED architecture for LLM orientation; when code and blueprint diverge, the code is the truth and the divergence is a documentation-debt signal (surfaced by M1 Detect), NOT a build failure.

## §B. Scope

### §B.1 In Scope — M4 only (4-tier enrichment overlays)

1. **Tier 0 — Contract**: a fourth graph entity type `contract` (additive), build-enforced drift check, enumerated contract surfaces for moai-adk-go (nav-graph schema, hook JSON schemas, CLI flag schemas) + a template-distributed contract registry for downstream users.
2. **Tier 1 — Blueprint**: an AUTHORED `module_tree.json` + per-module `overview.md` (Kiro Design 7-section template) under `.moai/project/blueprint/`, drafted from `/moai codemaps` dep-graph output but refined by human/agent (blueprint-first, NOT auto-generate-and-replace).
3. **Tier 2 — Decision (ADR)**: formalize `.moai/decisions/` as the ADR home; each `@NAV:DEC-<id>` token (M0) resolves to an immutable ADR; supersede chain via `supersedes:` field + new ADR.
4. **Tier 3 — Symbol (astx→SCIP, 2-tier)**: extend `internal/navigator/astx/` (imported as-is) to emit per-symbol signature + declaration + references; add an LLM narrative layer (docstring + call-context) via CodeWiki-style `--update`/`--compare-to` incremental technique; SCIP forward-compatible per-language (M4 ships the astx/Go path, defers SCIP per design report 리스크 #4).
5. **Tier-overlay artifact**: a single `tiers.json` overlay at `.moai/project/navigator/tiers.json` that consumers JOIN with M0's `nav-graph.json` — M4 does NOT overwrite `nav-graph.json` (M0 owns it).
6. **2-tier architecture principle**: Tiers 1 and 3 are each a 2-tier split (deterministic structure layer + LLM narrative layer), separable so the deterministic layer is never blocked by LLM availability.

### §B.2 Carry-forward non-overlap invariants (load-bearing)

The five predecessor SPECs codified non-overlap contracts. M4 integrates ABOVE all of them; it does NOT relax any:

- **REQ-PN-016** (001) — Navigator ↔ LSEL non-overlap.
- **REQ-NA-011** (002) — Audit ↔ LSEL + Audit ↔ 003 non-overlap.
- **REQ-NT-018 / REQ-NT-019** (003) — Enricher ↔ LSEL + Enricher ↔ 002 audit non-overlap.
- **REQ-NS-013 / REQ-NS-015 / REQ-NS-016** (M0) — integration layer ↔ LSEL + 002 audit + 003 enrich non-overlap, extended to the M0 integration-layer write surface.
- **REQ-NS2-005 / REQ-NS2-008** (M1) — Detect consumer-only on M0 + mx; Detect non-overlap with the 3 predecessors.

M4 carries these forward as REQ-NS3-016 / REQ-NS3-017 / REQ-NS3-018, extended to the 4-tier-layer write surface. The verification is recorded in research.md §3.

## §C. Requirements (GEARS notation)

### §C.1 Tier 0 — Contract (build-enforced, drift-immune)

#### REQ-NS3-001 (Ubiquitous — contract node type, additive schema)

The 4-tier layer SHALL extend the M0 nav-graph node set with a fourth entity type `contract`. The extension SHALL be additive-only: the existing three entity types (`decision`, `spec`, `symbol` per REQ-NS-007) and three edge types (`dec-edge`, `spec-edge`, `sym-edge` per REQ-NS-008) SHALL keep their names and shapes. A `contract` node SHALL carry `contract_kind` ∈ `{schema, allowlist, openapi}`, `contract_path` (absolute path to the declaration), and `drift_status` ∈ `{unknown, aligned, drifted}`.

#### REQ-NS3-002 (Capability gate — build-enforced drift check)

**Where** a contract node is registered, the 4-tier layer SHALL emit a build-time drift check that validates the declared contract against the implementation reality. **When** the declared contract and the implementation diverge, the check SHALL fail with a non-zero exit code in CI contexts and SHALL emit an advisory log line (not a failure) in local interactive contexts. The drift check SHALL be fail-open with respect to the Navigator graph itself: a contract-drift finding SHALL NOT block `nav-graph.json` or `tiers.json` emission.

#### REQ-NS3-003 (Ubiquitous — contract surfaces enumerated)

The 4-tier layer SHALL recognize three contract surfaces for moai-adk-go itself: (a) the `nav-graph.json` schema (already build-enforced via M0's byte-stable characterization test), (b) the hook input JSON schemas consumed by `internal/hook/`, and (c) the cobra CLI command + flag schemas at `internal/cli/`. Template-distributed users SHALL declare their own contract surfaces via a contract registry file (`.moai/project/blueprint/contracts.yaml`) listing `{contract_kind, contract_path, validator_command}` triples; an empty registry degrades gracefully to "no contract nodes emitted."

### §C.2 Tier 1 — Blueprint (authored, LLM entry map)

#### REQ-NS3-004 (Ubiquitous — module_tree.json schema, authored)

The 4-tier layer SHALL introduce a `module_tree.json` artifact at `.moai/project/blueprint/module_tree.json` describing the module hierarchy: each entry carries `package_path`, `display_name`, `layer` ∈ `{presentation, domain, infrastructure, measurement}` (the four moai-adk-go layers, language-neutralized for template distribution), `responsibility` (one-line), and `depends_on` (list of package paths). The artifact SHALL be produced as an AUTHORED draft scaffolded from `/moai codemaps` dep-graph output (`dependencies.md`), then refined by human or agent; it SHALL NOT be auto-generated-and-replaced on every run.

#### REQ-NS3-005 (Ubiquitous — overview.md per-module, Kiro 7-section)

The 4-tier layer SHALL introduce a per-module `overview.md` template at `.moai/project/blueprint/<module>/overview.md` structured on the Kiro Design seven sections: Component Architecture, Data Flow, Data Model, Error Handling, Test Strategy, Implementation Approach, Migration. The template SHALL be authored blueprint-first and SHALL carry a provenance block recording the last-authoring commit.

#### REQ-NS3-006 (Ubiquitous — blueprint-first stance, NOT spec-as-source)

The blueprint layer SHALL be spec-anchored per Fowler's reversible-optimum definition (design report Q1): it documents the INTENDED architecture for LLM orientation and is REVERSIBLE — when code and blueprint diverge, the code is the truth, and the divergence is a documentation-debt signal surfaced by M1 Detect (REQ-NS2-003), NOT a build failure. The blueprint layer SHALL NOT be spec-as-source MDD: it SHALL NOT generate code, and a blueprint↔code drift SHALL NOT fail the build, fail a test, or block a tool call.

#### REQ-NS3-007 (Ubiquitous — blueprint node + module-edges in graph)

The 4-tier layer SHALL wire blueprint nodes into the graph additively: entity_type `blueprint` (one per `module_tree.json` entry) and two new edge types — `module-edge` (blueprint→blueprint, sourced from `depends_on`) and `owns-edge` (blueprint→symbol, joining the authored module to the symbols 003 extracted within it). The new edge types SHALL NOT rename or redefine M0's existing `dec-edge`/`spec-edge`/`sym-edge`.

### §C.3 Tier 2 — Decision (ADR, immutable + supersede)

#### REQ-NS3-008 (Ubiquitous — ADR home + four canonical fields)

The 4-tier layer SHALL adopt `.moai/decisions/` (which already exists with ADR-shaped files) as the ADR home. Each ADR SHALL be an immutable markdown file named `<dec-id>.md` corresponding to a `@NAV:DEC-<id>` token (M0 REQ-NS-003), carrying the four canonical ADR fields: Decision Date, Status, Context, Decision, Consequences. A `@NAV:DEC-<id>` token whose `<id>` has no corresponding ADR file SHALL degrade gracefully (the decision node remains in the graph with `adr_path` empty).

#### REQ-NS3-009 (Event-detected — supersede chain, immutable prior)

**When** a decision is revised, the 4-tier layer SHALL create a NEW ADR file carrying a `supersedes:` field linking to the prior ADR's `<dec-id>` and SHALL set the prior ADR's `Status:` field to `Superseded`. The prior ADR body (Context / Decision / Consequences) SHALL NOT be edited — only its `Status:` line flips. The graph's decision node for the prior decision SHALL gain a `superseded_by` directed edge to the new decision node.

#### REQ-NS3-010 (Ubiquitous — decision node enrichment, additive)

The 4-tier layer SHALL enrich M0 decision nodes additively with `adr_path` (absolute path to the ADR file, empty when no ADR exists) and `superseded_by` (the superseding decision id, empty when current). The enrichment SHALL ride in the `tiers.json` overlay (REQ-NS3-018), NOT in `nav-graph.json` (M0 owns that artifact); consumers JOIN the two.

### §C.4 Tier 3 — Symbol (astx→SCIP, 2-tier)

#### REQ-NS3-011 (Ubiquitous — deterministic structure layer, astx extended)

The 4-tier layer SHALL extend the deterministic structure layer by importing `internal/navigator/astx/` as-is and emitting, per symbol, a structured record carrying `signature` (the Go signature or language-equivalent), `declaration_path` + `declaration_line` (where the symbol is defined), and `references` (the list of caller locations). This upgrades 003's flat `PrimarySymbols[].Name` list (a frequency-sorted name set) to a per-symbol structured record.

#### REQ-NS3-012 (Ubiquitous — LLM narrative layer scaffolding)

The 4-tier layer SHALL introduce a per-symbol narrative slot at `.moai/project/navigator/symbols/<symbol>.md` populated via a CodeWiki-style `--update` / `--compare-to` incremental technique: a `metadata.json` sidecar tracks `last_updated_commit` per symbol, and the narrative is re-drafted ONLY when the symbol's deterministic record changed since that commit. The narrative SHALL be LLM-drafted and human/agent-approved; it SHALL NOT be auto-committed without approval.

#### REQ-NS3-013 (Capability gate — SCIP forward-compat, per-language gradual)

**Where** a SCIP indexer exists and is configured for a language, the 4-tier layer MAY consume SCIP output as the deterministic structure source for that language. **Where** no SCIP indexer exists — the current moai-adk-go state, where Go uses astx — the layer SHALL use astx and SHALL defer SCIP integration to a per-language gradual rollout (design report 리스크 #4 mitigation). M4 ships the astx/Go path; SCIP integration is forward-compatible (additive) and is NOT a blocking dependency for M4 acceptance.

#### REQ-NS3-014 (Ubiquitous — symbol node enrichment, additive)

The 4-tier layer SHALL enrich M0 symbol nodes additively with `signature`, `declaration_path` + `declaration_line`, `references` (caller list, capped at a configurable N to bound output size), and `narrative_path` (absolute path to the narrative file, empty when no narrative exists). A symbol without a narrative SHALL degrade to structure-only (deterministic record present, narrative_path empty).

### §C.5 2-tier architecture principle (cross-cutting)

#### REQ-NS3-015 (Ubiquitous — 2-tier split: deterministic + LLM narrative)

Tiers 1 and 3 SHALL each be implemented as a 2-tier split: a deterministic structure layer (machine-produced, reproducible, no LLM required — `module_tree.json` scaffolded from `/moai codemaps` dep-graph for Tier 1; astx/SCIP per-symbol records for Tier 3) and an LLM narrative layer (human/agent-authored, reversible — per-module `overview.md` prose for Tier 1; per-symbol docstring + call-context for Tier 3). The two layers SHALL be separable so that the deterministic layer is produced and consumed WITHOUT depending on LLM availability, and so that LLM narrative unavailability degrades gracefully (structure-only) rather than blocking.

### §C.6 Non-overlap + consumer-only (cross-cutting)

#### REQ-NS3-016 (Ubiquitous — consumer-only on M0/M1 producers; bridge-not-absorb)

The 4-tier layer SHALL be strictly a CONSUMER of M0's `internal/navigator/sync/` outputs (`nav-graph.json`) and M1's `internal/navigator/detect/` + `internal/hook/navigator_detect*.go` outputs (the detect impact record). The 4-tier layer SHALL NOT modify `internal/navigator/sync/`, `internal/navigator/detect/`, `internal/hook/navigator_detect*.go`, or `internal/hook/post_tool.go`. This carries forward the M0 bridge-not-absorb principle (REQ-NS-005, REQ-NS-012) and the M1 consumer-only principle (REQ-NS2-005) to the M4 surface. Graph-schema extension is additive-only (forward-compatible per `.claude/rules/moai/workflow/nav-tokens.md`).

#### REQ-NS3-017 (Ubiquitous — non-overlap with 3 predecessors + LSEL)

The 4-tier layer SHALL NOT write to the three predecessor Navigator chains' surfaces — `.moai/project/navigator/capability-map.md` (001), `.moai/project/navigator/audit-report.{md,json}` (002), `.moai/project/navigator/capability-symbols.{md,json}` (003) — and SHALL NOT read or write any LSEL surface (`.moai/lessons-inbox.jsonl`, `.moai/state/lsel/`, `memory/feedback_*.md`, `hns-lsel-*` skills, `.moai/state/lsel/proposals/`). This carries forward REQ-PN-016 (001), REQ-NA-011 (002), REQ-NT-018/019 (003), and M0's REQ-NS-013/015/016 to the M4 boundary.

#### REQ-NS3-018 (Ubiquitous — write-surface isolation, overlay-not-overwrite)

The 4-tier layer's write surfaces SHALL be limited to exactly: `.moai/project/blueprint/module_tree.json`, `.moai/project/blueprint/<module>/overview.md`, `.moai/project/blueprint/contracts.yaml`, `.moai/decisions/<dec-id>.md` (NEW ADR files only — the ADR PROSE body is immutable per REQ-NS3-009; the one editable field is the `Status:` metadata line, which the supersede operation flips from `Accepted` to `Superseded`, and that metadata flip is NOT a body edit), `.moai/project/navigator/tiers.json` (the tier-enrichment overlay), and `.moai/project/navigator/symbols/<symbol>.md` (per-symbol narrative). The layer SHALL NOT overwrite `.moai/project/navigator/nav-graph.json` (M0 owns it); M4 emits `tiers.json` as an OVERLAY that consumers JOIN with `nav-graph.json`.

### §C.7 Verification + distribution (cross-cutting)

#### REQ-NS3-019 (Ubiquitous — provenance, no wall-clock)

Every deterministic-layer tier artifact (`module_tree.json`, `tiers.json`, the astx per-symbol structure records folded into `tiers.json`) SHALL carry a provenance block traceable to a git baseline: `extract_commit_sha` (`git rev-parse HEAD`) and `captured_at` (committer date of that SHA, `git log -1 --format=%cI`), using NO wall-clock timestamp. This carries forward M0's REQ-NS-009 and 003's `enrich.go:21` `Provenance` model, so two runs on the same HEAD produce byte-identical deterministic-layer output.

#### REQ-NS3-020 (Event-detected — fail-open on every error mode)

**When** the 4-tier layer encounters any failure — `nav-graph.json` absent (M0 not yet run), `tiers.json` unparseable, astx extraction error, narrative file absent, or per-tier timeout exceeded — the layer SHALL fail-open: exit code 0, emit no tier enrichment for the affected node, surface NO user-facing error, and append a single diagnostic line to `.moai/logs/navigator-sync.log`. The layer SHALL NEVER abort the calling `/moai project` or `/moai codemaps` step. This carries forward M0's REQ-NS-011 and 003's fail-open contract.

#### REQ-NS3-021 (Capability gate — template-first distribution)

**Where** the 4-tier layer ships author-facing surfaces — the Kiro 7-section `overview.md` template, the `module_tree.json` schema doc, the ADR template (the four canonical fields), the `contracts.yaml` registry format, the per-symbol narrative template — the documentation SHALL be authored under `internal/template/templates/` first, regenerated via `make build`, and verified by the `internal/template/internal_content_leak_test.go` neutrality guard (CLAUDE.local.md §2 [HARD] Template-First Rule + §25 Template Internal-Content Isolation). The run-phase touch list (template paths + mirror paths) is named in plan.md §D so compliance is planned, not discovered late.

#### REQ-NS3-022 (Capability gate — LLM context-precision success metric, mechanically measured)

**Where** the 4-tier layer's headline success metric is evaluated — the design report §성공 지표 "blueprint 선행 독서 후 추가 파일 reads 수, 현 상태 대비 ≥ 40% 감소" (the 오라버니 insight's direct measurement, named the most important metric) — the percentage SHALL be produced by the mechanical measurement procedure named in acceptance.md §D.AC-NS3-022 (a fixture corpus of orientation tasks + the count of additional file reads with vs without blueprint pre-read). The percentage SHALL NOT be a narrative claim, a sampled estimate, or a carried-over number — every assertion MUST be attributable to a run command + observed output (per `verification-claim-integrity.md` §1.1 surface 3 + §2 attribution).

## §D. Constraints (Non-Functional)

- **Language**: Go (the 4-tier layer is a Go binary/library under `internal/navigator/tiers/`, importing `internal/navigator/astx/` as-is).
- **Determinism**: deterministic-layer artifacts (REQ-NS3-004 `module_tree.json`, REQ-NS3-019 `tiers.json`) MUST be byte-identical on re-run against the same HEAD. LLM-narrative artifacts (overview.md prose, per-symbol docstrings) are NOT deterministic and carry their own `last_updated_commit` provenance.
- **Idempotence**: re-running the tier-enrichment on the same inputs produces the same `tiers.json` overlay.
- **Additive-only schema**: graph extension is forward-compatible (per `.claude/rules/moai/workflow/nav-tokens.md`); existing node/edge types keep their names and shapes.
- **Overlay-not-overwrite**: M4 NEVER overwrites `nav-graph.json` (REQ-NS3-018); consumers JOIN `tiers.json` with `nav-graph.json`.
- **Fail-open**: the layer never aborts any caller; all errors log and return exit 0 (REQ-NS3-020).
- **Performance budget**: the tier-enrichment runs inside the existing `/moai project` timeout envelope. The astx per-symbol extraction (Tier 3 deterministic) reuses 003's bounded scan; the LLM narrative layer is opt-in (not run on every `/moai project` invocation).
- **Provenance alignment** (`verification-claim-integrity.md`): every tier enrichment is attributable to a git baseline (REQ-NS3-019). `tiers.json` + the per-symbol narrative `metadata.json` are the Evidence; the enrichment claim is the Claim.
- **Non-overlap**: REQ-NS3-016 / REQ-NS3-017 / REQ-NS3-018 are inviolable.
- **Template-First**: REQ-NS3-021 binds any documentation edit to the template-first cycle.
- **Blueprint-first, NOT spec-as-source**: REQ-NS3-006 is the load-bearing stance decision — the blueprint does not generate code and drift is not a build failure.

## §E. Out of Scope

### Out of Scope — BAS Epic M2 (Route) and M3 (Fix)

- Promoting an M1 Detect advisory into a tracked work item (M2 Route) and auto-drafting an incremental regeneration of the map (M3 Fix). M4 produces the 4-tier map that M2/M3 will eventually route-to and fix; M4 does NOT implement the route or fix loop. Owned by separate SPECs (M2 = `SPEC-NAVIGATOR-SYNC-004` or later; M3 = `SPEC-NAVIGATOR-SYNC-005` or later, per the re-numbering note in HISTORY).

### Out of Scope — BAS Epic M5 (Brownfield reverse-extraction)

- Reverse-extracting docs from existing code via the tessl `document --code` pattern (M5). M4's Tier 1 Blueprint is authored/scaffolded, not reverse-extracted wholesale; M5 is the brownfield path that may seed blueprints from code, and it depends on M4 (design report §마일스톤: "M5 — M4 완료 후").

### Out of Scope — SCIP indexer implementation for non-Go languages

- Building or vendoring a SCIP indexer for any language. M4 ships the astx/Go deterministic path (REQ-NS3-013) and leaves the SCIP integration as forward-compatible, per-language gradual rollout (design report 리스크 #4). Non-Go symbol extraction continues to use 003's existing AST-extraction surface unchanged.

### Out of Scope — modification of M0/M1 producers or the 3 predecessor chains

- M4 does NOT modify `internal/navigator/sync/` (M0), `internal/navigator/detect/` or `internal/hook/navigator_detect*.go` or `internal/hook/post_tool.go` (M1), `internal/navigator/astx/` (001/003, imported as-is), `internal/mx/`, `scripts/navigator-audit.sh`, `scripts/navigator-regen.sh`. These are CONSUMED (REQ-NS3-016).

### Out of Scope — auto-committing LLM narrative without approval

- The Tier 1 overview.md prose and the Tier 3 per-symbol narrative are LLM-drafted and SHALL be human/agent-approved before commit (REQ-NS3-012). Auto-committing narrative drafts into the working tree or a SPEC branch without an approval gate is out of scope and is the spec-as-source hazard's adjacent failure mode.

### Out of Scope — LSEL surfaces

- M4 strictly avoids LSEL surfaces (REQ-NS3-017). The harness self-evolution concern is owned by SPEC-LSEL-LOCAL-EVOLUTION-001.

### Out of Scope — a new top-level `moai` subcommand

- M4's tier-enrichment is invoked from the existing `/moai project` surface as a sibling step (the same way M0's `navigator-sync` and 003's `navigator-enrich` are wired). No new top-level `moai` subcommand; the tier step is a Hidden cobra command mirroring M0's `navigator-sync` Hidden subcommand.

### Out of Scope — making blueprint↔code drift a build failure

- Per REQ-NS3-006 (the blueprint-first stance), blueprint↔code drift is a documentation-debt signal surfaced by M1 Detect, NOT a build failure, NOT a test failure, and NOT a tool-call block. Building a CI gate that fails on blueprint drift is explicitly out of scope; only Tier 0 Contract drift is build-enforced (REQ-NS3-002), and the two MUST NOT be conflated.

## §F. AC Matrix (summary — full Given-When-Then detail in acceptance.md)

| AC-ID     | REQ-ID         | Summary                                                                                    |
|-----------|----------------|--------------------------------------------------------------------------------------------|
| AC-NS3-001 | REQ-NS3-001    | `tiers.json` carries `contract` entity type additively; existing 3 types unchanged          |
| AC-NS3-002 | REQ-NS3-002    | Contract drift → non-zero exit in CI, advisory log locally; never blocks graph emission     |
| AC-NS3-003 | REQ-NS3-003    | 3 moai-adk-go contract surfaces recognized + empty registry degrades gracefully             |
| AC-NS3-004 | REQ-NS3-004    | `module_tree.json` authored draft scaffolded from codemaps; NOT auto-replace                |
| AC-NS3-005 | REQ-NS3-005    | Per-module `overview.md` carries Kiro 7 sections + provenance                               |
| AC-NS3-006 | REQ-NS3-006    | Blueprint drift is debt signal NOT build failure (grep: no drift-fail test exists)          |
| AC-NS3-007 | REQ-NS3-007    | `blueprint` node + `module-edge`/`owns-edge` additive; existing edges unchanged             |
| AC-NS3-008 | REQ-NS3-008    | `.moai/decisions/<dec-id>.md` ADR carries 4 canonical fields; missing-ADR degrades          |
| AC-NS3-009 | REQ-NS3-009    | Supersede creates new ADR + flips prior Status; prior body unedited; `superseded_by` edge  |
| AC-NS3-010 | REQ-NS3-010    | Decision node enriched with `adr_path` + `superseded_by` in `tiers.json` overlay            |
| AC-NS3-011 | REQ-NS3-011    | Per-symbol record carries signature + declaration + references (astx extended)              |
| AC-NS3-012 | REQ-NS3-012    | Narrative `metadata.json` tracks `last_updated_commit`; re-draft only on deterministic Δ   |
| AC-NS3-013 | REQ-NS3-013    | Go astx path ships; SCIP absent → graceful (no SCIP dependency at acceptance)               |
| AC-NS3-014 | REQ-NS3-014    | Symbol node enriched with signature/declaration/references/narrative_path                   |
| AC-NS3-015 | REQ-NS3-015    | 2-tier split: deterministic layer produced without LLM; narrative degrades gracefully       |
| AC-NS3-016 | REQ-NS3-016    | Consumer-only: M0/M1 producer paths untouched (grep + git diff empty)                       |
| AC-NS3-017 | REQ-NS3-017    | Non-overlap grep touches no predecessor surface + no LSEL surface                           |
| AC-NS3-018 | REQ-NS3-018    | Write-surface isolation: only the 6 named paths; `nav-graph.json` never overwritten         |
| AC-NS3-019 | REQ-NS3-019    | Provenance no wall-clock; byte-identical deterministic re-run                                |
| AC-NS3-020 | REQ-NS3-020    | Fail-open on each error mode (graph absent / unparseable / astx error / narrative absent)   |
| AC-NS3-021 | REQ-NS3-021    | Template-first: author-facing surfaces under `internal/template/templates/` + neutrality green |
| AC-NS3-022 | REQ-NS3-022    | ≥40% reads-reduction metric produced by named fixture procedure, not narrative              |

## §G. Cross-References

- **Design report (SSOT for this SPEC's rationale)** — `.moai/reports/navigator-redesign-bas-20260805.{md,html}` §제안 BAS (B) line 28 (4-tier detail), §자산 재사용 line 31-39, §세 물음 답 Q1/Q2/Q3 line 42-44, §마일스톤 M4 line 51, §리스크 line 54-60, §성공 지표 line 62-67 (session 3db3f943).
- **SPEC-NAVIGATOR-SYNC-001** (M0, `status: completed`, PR #1375 `f4bd58acbd`) — produces `nav-graph.json`, the graph M4 extends via the `tiers.json` overlay. Non-overlap REQ-NS-013/015/016 carried forward as REQ-NS3-017. `depends_on` target.
- **SPEC-NAVIGATOR-SYNC-002** (M1, `status: completed`, PR #1379 `304907b6d`) — produces the Detect impact record at `.moai/state/navigator-detect/`; M4's blueprint↔code drift is surfaced as a documentation-debt signal by M1 Detect (REQ-NS3-006). Non-overlap REQ-NS2-005/008 carried forward.
- **SPEC-PROJECT-NAVIGATOR-001** (regen, completed) — produces `capability-map.md`. Non-overlap REQ-PN-016 carried forward as REQ-NS3-017.
- **SPEC-PROJECT-NAVIGATOR-002** (audit, completed) — produces `audit-report.{md,json}`. Non-overlap REQ-NA-011 carried forward.
- **SPEC-PROJECT-NAVIGATOR-003** (enrich, completed) — produces `capability-symbols.json`; M4 Tier 3 reuses its `PrimarySymbols[]` and `internal/navigator/astx/` engine. Non-overlap REQ-NT-018/019 carried forward.
- **`.claude/rules/moai/workflow/nav-tokens.md`** — the M0 binding-token trio author surface (the `@NAV:DEC-<id>` Tier 2 resolves; the `@NAV:SYM:<symbol>` Tier 3 enriches).
- **`internal/navigator/astx/`** — the Tier 3 deterministic engine (imported as-is, extended additively).
- **`.moai/decisions/`** — the existing ADR home (Tier 2 formalizes it).
- **`.moai/project/codemaps/`** — the `/moai codemaps` auto-generated surface (Tier 1 `module_tree.json` is scaffolded from its `dependencies.md`).
- **`.claude/rules/moai/core/verification-claim-integrity.md`** §1.1 surface 3 + §2 — the success-metric attribution obligation (REQ-NS3-022).
- **CLAUDE.local.md §2 [HARD] Template-First Rule + §25 Template Internal-Content Isolation** — bind REQ-NS3-021.
