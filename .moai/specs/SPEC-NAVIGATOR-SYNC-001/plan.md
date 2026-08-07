# SPEC-NAVIGATOR-SYNC-001 — Implementation Plan

> Tier L · status: draft · v0.1.0 · plan-phase only (Implementation Kickoff Approval held by orchestrator)

## §A. Context

BAS Epic M0 — the critical-path entry. All later milestones (M1–M5, planned as `SPEC-NAVIGATOR-SYNC-002..006`) depend on the graph-join substrate that M0 produces. The design report (`.moai/reports/navigator-redesign-bas-20260805.{md,html}`, session 3db3f943) is the SSOT for rationale; this plan is the HOW.

Three concrete assets are reused, not reinvented (design report §4):

1. `internal/navigator/astx/enrich.go:21` — `Provenance` model (git baseline, no wall-clock). M0's REQ-NS-009 carries it forward byte-for-byte into the join layer.
2. `internal/navigator/astx/enrich.go:122` — header-driven `parseCapabilityMap` (robust to column reorder). M0 reuses this to parse `capability-map.md` for spec-node discovery.
3. `internal/cli/navigator_enrich.go:128` — `atomicWrite` + `NAVIGATOR_PRE_RENAME_BARRIER` test hook. M0's REQ-NS-010 mirrors this pattern verbatim.

Design-report precision correction (recorded in spec.md HISTORY): the lean twin calls `@MX:SPEC` "absorbed", but `internal/mx/scanner.go:139-155` + `spec_association.go:17-20` ALREADY consume it as a third SPEC-association source. M0 therefore BRIDGES (not absorbs) the mx-scanner output into the Navigator graph and MUST NOT modify `internal/mx/`.

## §B. Known Issues

1. **mx-scanner output contract** — the `SpecAssociator.AssociateWithDiagnostics(tag)` API returns per-tag associations; M0 needs a project-wide `specID → []tag` inverse. This is a read-only consumer reshape; the location was a pre-flight decision at M0.1 — resolved in §C / design.md §1.D1. Default proposal: consumer-side in `internal/navigator/sync/`, to honor REQ-NS-005 ("SHALL NOT modify `internal/mx/`") strictly.
2. **`@NAV:DEC` id grammar vs SPEC-ID domain tokens** — the grammar `[A-Z][A-Z0-9-]*` overlaps with SPEC-ID domain tokens (`AUTH`, `V3R6`, etc.). A design-decision id like `@NAV:DEC-AUTH` would collide visually with `SPEC-AUTH-NNN`. Mitigation: the `@NAV:DEC-` prefix is the unambiguous discriminator; the id alone never appears without the prefix. Documented in template rule.
3. **Symbol-node identity** — 003's `capability-symbols.json` uses `PrimarySymbols[].Name` as the symbol identifier; `@NAV:SYM:<symbol>` tokens authored in design docs or code comments may use unqualified or package-qualified forms. M0 MUST adopt a canonical form (package-qualified, language-aware) — resolved in §C / design.md §1.D2.
4. **SubagentStart hook stale spec reference** — the launch context carried `spec:SPEC-PROJECT-NAVIGATOR-003` (a closed SPEC). Confirmed not a collision with this SPEC's ID; ignored. Not a blocker.

## §C. Pre-flight Decisions (each gated, default proposed)

**D1 — SpecAssociator consumer reshape location** — RESOLVED — see design.md §1.D1 (consumer-side `internal/navigator/sync/mx_bridge.go`; no `internal/mx/` modification).

**D2 — Symbol-node canonical form** — RESOLVED — see design.md §1.D2 (language-aware package-qualified; dedup against 003's `PrimarySymbols[].Name` by exact-then-suffix match).

**D3 — Design-doc scan root** — RESOLVED — see design.md §1.D3 (`.moai/project/{product,structure,tech}.md` + `.moai/docs/**/*.md`; excludes `.moai/specs/`, `.moai/reports/`, `.moai/state/`).

## §D. Constraints (restated from spec.md §D, operationalized)

- **Determinism** — sort everything before serialization: node list by `(entity_type, identifier)`, edge list by `(edge_type, source_node, target_node, source_path, line_number)`. `map[string]V` iteration goes through sorted keys. This is how REQ-NS-009's byte-identical re-run guarantee is met.
- **Fail-open** — all scanner/parse errors log to `.moai/logs/navigator-sync.log` and are swallowed; the cobra `RunE` returns nil unconditionally (mirrors `navigator_enrich.go`).
- **No wall-clock** — `captured_at` is `git log -1 --format=%cI` of HEAD. No `time.Now()`.
- **Hidden CLI** — the join step is a Hidden cobra command, wired into the existing `/moai project` flow as a sibling step (NOT a new top-level `moai` subcommand).

## §E. Self-Verification (plan-phase, the §E.1 audit-ready signal)

This plan-phase artifact set is audit-ready when:

- [ ] All 18 REQs (REQ-NS-001..018) trace to at least one AC in acceptance.md.
- [x] All 3 pre-flight decisions (D1/D2/D3) RESOLVED — pointers in §C reference design.md §1.D{1,2,3}; no open clarification markers remain in plan.md or research.md (MP-7 firewall cleared).
- [ ] The 3 non-overlap invariants (REQ-NS-013/015/016) are machine-testable via grep-pair protocol (acceptance.md §A.1).
- [ ] The design-report precision correction (mx-scanner is a consumer, not a modification target) is reflected in plan.md §A + REQ-NS-005.
- [ ] Every verification claim in acceptance.md names the command + the observable output form, per `verification-claim-integrity.md` §2 attribution.

## §F. Milestones (ordered by decision-reversibility — highest-change-likelihood first)

### M0.1 — Schema + pre-flight decisions (most reversible, lead)

**Decision**: the `nav-graph.json` schema, the symbol-node canonical form, the design-doc scan root, the mx-bridge location. These are the decisions most likely to change; they lead so human review sees them first.

- Deliverable: `internal/navigator/sync/schema.go` — the Go types for `provenance`, `node`, `edge`, `binding_record`. JSON tags stable. Forward-compatible (additive only).
- Deliverable: `internal/navigator/sync/design.md` (this SPEC's design.md) recording the 3 pre-flight decisions RESOLVED with their default values.
- Exit gate: schema types compile, design.md carries the 3 resolved decisions, no open clarification markers remain in plan.md (the §C pre-flight decisions are RESOLVED pointers, not open questions).

### M0.2 — Non-overlap invariant tests (load-bearing guard, early)

**Why early**: the non-overlap contracts (REQ-NS-013/015/016) are load-bearing; getting them wrong late is catastrophic. Establish the invariant tests BEFORE writing the join engine, so every later step runs against them.

- Deliverable: `internal/navigator/sync/nonoverlap_test.go` — grep-pair tests (per the repo's `feedback_deadfield_grep_pair_protocol.md` lesson): (a) write-surface isolation, (b) LSEL non-overlap, (c) 002 audit write non-overlap, (d) 003 enrich write non-overlap.
- Deliverable: `.moai/logs/navigator-sync.log` path constant (REQ-NS-017 sink).
- Exit gate: tests RED initially (no implementation yet), then GREEN once M0.3/M0.4 land.

### M0.3 — Token scanners (mechanical, parallelizable)

- Deliverable: `internal/navigator/sync/scan_dec.go` — `@NAV:DEC-<id>` scanner over the design-doc scan root (REQ-NS-003).
- Deliverable: `internal/navigator/sync/scan_sym.go` — `@NAV:SYM:<symbol>` scanner over code + design (REQ-NS-004).
- Deliverable: `internal/navigator/sync/malformed_test.go` — malformed-token diagnostic fixture (REQ-NS-017).
- Reuse: the regex-shape mirrors `internal/mx/scanner.go:396` `extractSpecRef` (`@MX:SPEC:\s*(SPEC-[A-Z0-9][A-Z0-9-]*)`).

### M0.4 — Mx-bridge + join engine (core logic, depends on M0.1 + M0.3)

- Deliverable: `internal/navigator/sync/mx_bridge.go` — SpecAssociator consumer reshape (REQ-NS-005).
- Deliverable: `internal/navigator/sync/join.go` — the join engine. Inputs: `capability-map.md` (header-driven via 003's `parseCapabilityMap`), `capability-symbols.json` (003 output), `audit-report.json` (002 output, advisory), dec-bindings, sym-bindings, mx-associations. Output: in-memory `graph` struct.
- Deliverable: `internal/navigator/sync/provenance.go` — Provenance block (REQ-NS-009), reusing 003's `enrich.go:21` model.
- Deliverable: `internal/navigator/sync/write.go` — atomic-write (REQ-NS-010) + barrier test hook, mirroring `navigator_enrich.go:128`.
- Deliverable: `internal/navigator/sync/capability_gate.go` — fail-open when capability-map.md absent (REQ-NS-011).

### M0.5 — Hidden CLI subcommand + `/moai project` wiring

- Deliverable: `internal/cli/navigator_sync.go` — Hidden cobra subcommand `navigator-sync` (mirrors `navigator_enrich.go:54`). Flags: `--project-root`, `--capability-map`, `--capability-symbols`, `--audit-report`, `--out` (defaults: the standard navigator paths).
- Wiring: invoked as a sibling step inside `/moai project` AFTER 001's regen + 003's enrich. The caller already sequences 001→003; M0 adds the join step as a tail.
- Exit gate: `moai navigator-sync --project-root <tmp>` produces `nav-graph.json` in a temp dir (test fixture).

### M0.6 — Template-first documentation + verification close-out (final, mechanical)

- Deliverable: `internal/template/templates/.claude/rules/moai/workflow/nav-tokens.md` — documents `@NAV:DEC-<id>` and `@NAV:SYM:<symbol>` for distributed template users. Authored template-first per CLAUDE.local.md §2.
- `make build` regeneration + `internal/template/internal_content_leak_test.go` neutrality guard green (REQ-NS-018).
- Deliverable: full `go test ./internal/navigator/sync/...` green; `go test ./internal/cli/...` green for the new navigator-sync subcommand test.
- Deliverable: characterization test that re-runs the join twice on the same HEAD and asserts byte-identical output (REQ-NS-009).

## §G. Anti-Patterns (specific to this SPEC)

- **AP-NS-001 — Re-scan code for `@MX:SPEC`**: duplicating `internal/mx/scanner.go`'s job. Forbidden by REQ-NS-005. Bridge, don't re-scan.
- **AP-NS-002 — Modify the 3 chains or `internal/mx/`**: forbidden by REQ-NS-012 + REQ-NS-005. Integration ABOVE, not INTO.
- **AP-NS-003 — Wall-clock timestamp anywhere**: forbidden by REQ-NS-009. The day a `time.Now()` lands in `join.go`, byte-identical re-run breaks.
- **AP-NS-004 — Touch LSEL surfaces**: forbidden by REQ-NS-013. The `feedback_deadfield_grep_pair_protocol.md` lesson applies — use grep-pair tests, not bare grep.
- **AP-NS-005 — Template doc authored in local `.claude/rules/` first**: forbidden by CLAUDE.local.md §2. Template-first means `internal/template/templates/` first, then `make build`, then verify local received it.
- **AP-NS-006 — Unsorted map iteration in JSON serialization**: breaks byte-identical re-run (REQ-NS-009). Sort keys explicitly.

## §H. Cross-References

- spec.md §B.2 — non-overlap carry-forward table.
- acceptance.md §A.1 — grep-pair test methodology for non-overlap ACs.
- design.md — the 3 pre-flight decisions RESOLVED with defaults + the schema JSON example.
- research.md — design-report §4 asset inventory verified against current main, external references (CodeWiki, falconer, Fowler SDD, SCIP).
- `.claude/rules/moai/core/verification-claim-integrity.md` §2 — baseline-attribution obligation every AC must meet.
- `feedback_deadfield_grep_pair_protocol.md` (memory) — grep-pair methodology for REQ-NS-013/015/016 ACs.
- `feedback_template_first_mirror_runphase.md` (memory) — template-first is a run-phase obligation, but documenting the template edit in plan-phase M0.6 is the plan-phase seed.
