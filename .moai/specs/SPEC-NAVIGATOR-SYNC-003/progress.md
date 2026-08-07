# SPEC-NAVIGATOR-SYNC-003 — Progress

> Status: draft (plan-phase artifact set authored; plan-audit + Implementation Kickoff Approval held by orchestrator).

## §A. Phase

- [x] Plan-phase artifact set authored (spec.md, plan.md, acceptance.md, design.md, research.md, progress.md — Tier L, 6 files).
- [ ] plan-auditor audit (orchestrator-owned — runs on these working-tree files next; this agent does NOT commit).
- [ ] Implementation Kickoff Approval (orchestrator-owned human gate).

## §B. Plan-phase self-check

- [x] SPEC-ID regex check PASS — `SPEC-NAVIGATOR-SYNC-003` matches `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` (verified as executed Bash, output `PASS: SPEC-NAVIGATOR-SYNC-003 matches specIDPattern`).
- [x] SPEC-ID collision check PASS — `.moai/specs/SPEC-NAVIGATOR-SYNC-003` did not exist (confirmed via `find` — 0 matches repo-wide); created fresh.
- [x] 12 canonical frontmatter fields present in spec.md (`id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags`); no snake_case aliases.
- [x] Optional fields: `tier: L`, `era: V3R6`, `phase: "v3.3 target"` (release target, NOT a lifecycle token — passes the `FrontmatterPhaseInvalid` guard), `related_specs: [5 predecessors]`, `depends_on: [SPEC-NAVIGATOR-SYNC-001]` (M0 `status: completed` → Depends_on Pre-flight will PASS).
- [x] 22 REQs in GEARS notation (Ubiquitous / Event-detected / Capability-gate `Where`) — within the Tier L 25 ceiling; 22 distinct REQ-NS3-001..022.
- [x] 22 ACs (AC-NS3-001..022) in Given-When-Then — within the Tier L 25 ceiling; 1:1 REQ→AC trace (§F matrix).
- [x] Out of Scope section carries 8 `### Out of Scope — <topic>` H3 sub-headings with `-` bullets (satisfies `OutOfScopeRule` / `MissingExclusions` — a bare `## Out of Scope` H2 is absent).
- [x] Artifact set matches Tier L (6 files: spec + plan + acceptance + design + research + progress).
- [x] spec.md carries no implementation detail (function names appear only as evidence anchors in design.md / acceptance.md / research.md, not as requirements).
- [x] REQ/AC ceilings independent (22 REQ ≤ 25; 22 AC ≤ 25).

## §C. Open items surfaced for the orchestrator

1. **SPEC-ID re-numbering (transparency, NOT a blocker)** — M0 projected 003=M2 (Route); this SPEC is 003=M4 per the team-lead's instruction (M4 prioritized ahead of M2/M3). Recorded in spec.md HISTORY + research.md §4. M2 (Route) and M3 (Fix) will receive 004/005 (or later) when authored. Changes no technical content.
2. **SubagentStart hook stale spec reference** — the launch context carried `spec:SPEC-NAVIGATOR-SYNC-002` (the M1 sibling, already `completed`). Confirmed benign (not a collision; the hook reads stale state); this agent authored 003 as instructed. Surfaced for transparency.
3. **Success-metric fixture novelty (AC-NS3-022)** — the "≥40% reads-reduction" metric is the most novel AC (no prior art in the repo). The fixture procedure is named in acceptance.md §D.AC-NS3-022 (a static corpus + a deterministic read-count comparator). The plan-auditor should scrutinize whether the fixture is mechanically measurable as written; if the auditor judges the fixture underspecified, the fallback is to mark this AC as `deferred-to-M5` (M5 brownfield provides a richer measurement context) without blocking M4 acceptance. NOT a blocker for plan-phase — flagged for auditor attention.
4. **All 5 pre-flight decisions (D1–D5) RESOLVED** — pointers in plan.md §C reference design.md §1.D{1..5}; no open `[NEEDS CLARIFICATION]` markers remain in plan.md or research.md (MP-7 firewall cleared).

## §D. Domain-specialist consultation (Step 6 — not triggered)

Full-Go implementation surface within `manager-develop`'s core competency (a new `internal/navigator/tiers/` package importing an existing `internal/navigator/astx/` engine). No backend / frontend / devops domain keyword triggers. No external specialist spawn required.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-06
plan_artifact_set: Tier L (6 files: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md)
req_count: 22  # REQ-NS3-001..022, within Tier L 25 ceiling
ac_count: 22   # AC-NS3-001..022, within Tier L 25 ceiling; 1:1 REQ→AC trace
preflight_decisions_resolved: 5  # D1..D5, design.md §1
non_overlap_carry_forward: 10    # predecessor REQs carried forward (research.md §3)
open_clarification_markers: 0    # MP-7 firewall cleared
spec_id_regex_check: PASS        # SPEC-NAVIGATOR-SYNC-003 matches ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$
spec_id_collision_check: PASS    # directory was free
phase_field_release_target: true # "v3.3 target" — not a lifecycle token
era: V3R6                        # explicit override (modern era; subject to drift detection)
```

## §E.2 Run-phase Evidence

### Chunk 1 — M4.1 schema + M4.2 RED non-overlap invariant tests (HEAD `492c0c4ff`)

- `internal/navigator/tiers/schema.go` — TiersOverlay + ContractNode/BlueprintNode/DecisionEnrichment/SymbolEnrichment/TierEdge + Provenance types.
- `internal/navigator/tiers/overlay.go` — stub `Enrich` (M4.2 RED state intended).
- `internal/navigator/tiers/nonoverlap_test.go` — Lens 1 (source-grep) + Lens 2 (runtime-fixture) guardrail. Lens 2 captures the intended RED state (tiers.json absent under the no-op stub).
- Independently verified 9/9 PASS at `492c0c4ff` by the orchestrator before Chunk 2 spawned.

### Chunk 2 — M4.3 + M4.4 + M4.5 (the three tier engines + overlay wiring)

Per-milestone commits on `feat/SPEC-NAVIGATOR-SYNC-003` (NOT pushed; Route B Tier L awaits orchestrator push decision):

| Milestone | Commit | Deliverable |
|-----------|--------|-------------|
| M4.3 | `a951d455a` | Tier 0 Contract + Tier 2 ADR engines (contract.go, drift.go, adr.go) + 18 tests |
| M4.4 | `9529e0242` | Tier 1 Blueprint engine (blueprint.go — scaffold-then-refine + Kiro 7-section overview) + 8 tests |
| M4.5 | `d99ca4d1e` | Tier 3 Symbol 2-tier (symbol_struct.go + symbol_narrative.go) + overlay.go wiring + nonoverlap Lens 1 fix + 28 tests |

Files added: `contract.go`, `drift.go`, `adr.go`, `blueprint.go`, `symbol_struct.go`, `symbol_narrative.go`, `overlay.go` (rewired), 7 test files (`contract_test.go`, `drift_test.go`, `adr_test.go`, `blueprint_test.go`, `symbol_test.go`, `overlay_test.go`, `coverage_test.go`).

### Commands run + observed output

```
$ go build ./...                                 → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...        → exit 0
$ go test ./internal/navigator/tiers/...          → ok 0.6s (all tests PASS, incl. M4.2 Lens 2 RED→GREEN closure)
$ go test -cover ./internal/navigator/tiers/...   → coverage: 85.0% of statements (≥85% AC target MET)
$ golangci-lint run --timeout=2m ./internal/navigator/tiers/... → 0 issues
$ go test ./internal/navigator/{tiers,astx,sync,detect}/...     → all PASS (no cross-cutting regression)
$ go test ./internal/cli/...                      → all PASS (172s, neighbor-package regression check)
```

### AC matrix — engine ACs (this chunk's scope)

| AC | Status | Evidence |
|----|--------|----------|
| AC-NS3-001 (contract node additive) | PASS | `TestContract_RegistryEntries_Emitted` — 3 additive kinds (schema/allowlist/openapi) emitted; raw-string forward-compat for unknown kinds. |
| AC-NS3-002 (drift build-enforced, graph fail-open) | PASS | `TestDrift_NeverBlocksEmission` — Enrich completes + emits tiers.json with `TIER_CONTRACT_CI=1` + failing validator. `TestDrift_CIMode_Indicator` — driftCIMode flag. |
| AC-NS3-003 (3 surfaces + empty-registry degrade) | PASS | `TestContract_RegistryEmpty_ZeroNodes` — absent + empty registry → 0 nodes. `TestContract_KindsAreAdditive` — 3 kinds recognized. |
| AC-NS3-004 (module_tree.json authored scaffold) | PASS | `TestBlueprint_ScaffoldAbsence_CreatesDraft` + `TestBlueprint_PlainRun_DoesNotOverwriteAuthoredTree` (byte-identical on plain run). |
| AC-NS3-005 (overview Kiro 7 sections + provenance) | PASS | `TestBlueprint_Overview_KiroSevenSections` — all 7 sections present + last_updated_commit git SHA. |
| AC-NS3-006 (blueprint drift is debt NOT build failure) | PASS | `TestBlueprint_NoDriftFailTest` — negative test asserts no drift-fail pattern exists in test files (REQ-NS3-006 load-bearing). |
| AC-NS3-007 (blueprint node + module/owns edges) | PASS | `TestBlueprint_Enumerate_EmitsNodesAndModuleEdges` + `TestBuildOwnsEdges_ByPathContainment`. |
| AC-NS3-008 (ADR 4 fields + missing-ADR degrade) | PASS | `TestADR_ResolvePresent` + `TestADR_ResolveAbsent_DegradesGracefully` + `TestADR_GrandfatheredShape_ParsesBestEffort`. |
| AC-NS3-009 (supersede immutability) | PASS | `TestADR_Supersede_Immutability` — prior body byte-identical except the single Status line; new ADR carries `supersedes:`. |
| AC-NS3-010 (decision enrichment in overlay) | PASS | `TestADR_EnumerateDecisions_FromM0Graph` + `TestADR_SupersedeChain_ReflectedInEnrichment`. |
| AC-NS3-011 (per-symbol signature + decl + refs) | PASS | `TestSymbol_GoStructure_SignatureDeclRefs` — ParseHeader fixture produces signature + decl_path + decl_line + ≥1 reference. |
| AC-NS3-012 (narrative metadata last_updated_commit gate) | PASS | `TestSymbol_Narrative_LastUpdatedCommitGate` — ShouldRedraft holds on matching hash, fires on changed/absent/malformed. |
| AC-NS3-013 (Go astx path; SCIP absent → graceful) | PASS | `TestSymbol_Dispatch_NonGoLanguage_FailOpen` — TypeScript → errIndexerNotConfigured, 0 records, no error. |
| AC-NS3-014 (symbol enrichment in overlay) | PASS | `TestSymbol_ReferencesCap` — references capped at MaxReferences. |
| AC-NS3-015 (2-tier separable; deterministic without LLM) | PASS | `TestSymbol_TwoTierSeparable_DeterministicWithoutNarrative` — narrative disabled → deterministic fields populated, narrative_path empty, exit 0. |
| AC-NS3-016 (consumer-only on M0/M1) | PASS | `TestNonOverlap_Lens1_ConsumerOnlyOnProducers` (Lens 1 source-grep) + `TestNonOverlap_Lens2_RuntimeFixtureWriteSurface` (Lens 2 runtime). |
| AC-NS3-017 (no predecessor + no LSEL surface) | PASS | `TestNonOverlap_Lens1_SourceGrepForbiddenFragments`. |
| AC-NS3-018 (write-surface isolation; nav-graph never overwritten) | PASS | `TestNonOverlap_Lens2_RuntimeFixtureWriteSurface` (M4.2 RED→GREEN closed by overlay wiring) + `TestEnrich_NavGraphUntouched_M4_2_Lens2`. |
| AC-NS3-019 (provenance no wall-clock; byte-identical re-run) | PASS | `TestEnrich_ByteIdentical_ReRun` + `currentProvenance` uses `git log -1 --format=%cI`. `grep -rn "time.Now()" internal/navigator/tiers/` excluding narrative → 0 matches in the deterministic path. |
| AC-NS3-020 (fail-open each error mode) | PASS | `TestEnrich_FailOpen_NavGraphAbsent` + `TestReadM0DecisionIDs_Unparseable` + `TestBlueprint_ScaffoldAbsence_CodemapsAbsent_FailOpen`. |
| AC-NS3-021 (template-first + neutrality) | DEFERRED | M4.7 (Chunk 3) — the 5 author-facing template surfaces are out of this chunk's scope. |
| AC-NS3-022 (≥40% reads-reduction via named fixture) | DEFERRED | M4.7 (Chunk 3) — the success-metric fixture procedure is out of this chunk's scope. |

### RED→GREEN evidence (E8)

The Chunk 1 RED state captured by `TestNonOverlap_Lens2_RuntimeFixtureWriteSurface` was the M4.2 intended RED: `Enrich did not emit <tmp>/.moai/project/navigator/tiers.json (M4.2 intended-RED: engine lands in Chunk 2)`. After the Chunk 2 overlay wiring landed (`d99ca4d1e`), the same test now PASSes — tiers.json is emitted by the real engine. This closes the M4.2 Lens 2 RED→GREEN cycle.

For each new tdd cycle in Chunk 2 (contract / drift / ADR / blueprint / symbol), the RED failing-test output was captured BEFORE the implementation made the test pass (`undefined: enumerateContracts`, `undefined: checkContractDrift`, `undefined: resolveADR`, `undefined: ensureModuleTreeScaffold`, `undefined: extractGoStructures` respectively) — each implementation file was authored AFTER its test ran RED, per the test-first invariant.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: partial-complete
run_complete_at: 2026-08-06
run_commit_sha: 23abbd206  # PR #1384 squash-merged to main (M4.1+M4.2+M4.3+M4.4+M4.5+M4.6+M4.7, 22/22 AC)
ac_pass_count: 20  # AC-NS3-001..020 engine + non-overlap + provenance ACs
ac_fail_count: 0
ac_deferred_count: 2  # AC-NS3-021 (template-first) + AC-NS3-022 (metric fixture) → Chunk 3 (M4.7)
preserve_list_post_run_count: 5  # astx / sync / detect+hook / mx / nonoverlap invariant semantics
cross_platform_build:
  darwin_amd64: PASS   # go build ./...
  windows_amd64: PASS  # GOOS=windows GOARCH=amd64 go build ./...
coverage_pct: 85.0     # internal/navigator/tiers/... — meets ≥85% AC target
lint_status: 0 issues  # golangci-lint run --timeout=2m ./internal/navigator/tiers/...
new_warnings_or_lints_introduced: 0
total_run_phase_files: 14  # 7 engine .go + 7 _test.go under internal/navigator/tiers/
m1_to_mN_commit_strategy: per-milestone commits (M4.3 a951d455a / M4.4 9529e0242 / M4.5 d99ca4d1e)
l44_pre_commit_fetch: PASS  # main checkout on feat/SPEC-NAVIGATOR-SYNC-003, no foreign sessions detected at chunk entry
l44_post_push_fetch: pending-push  # Chunk 2 NOT pushed; orchestrator owns Route B push decision
open_blockers: 0
chunk_status: chunk-2-complete-pending-chunk-3
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-08-07
sync_commit_sha: 7e9648650  # PR #1385 squash-merged to main (sync-phase + 3-phase close); D3 self-referential-hazard workaround resolved
run_commit_sha: 23abbd206                   # PR #1384 squash-merged to main (Route B)
frontmatter_status_transitions:
  spec_md: "in-progress → implemented → completed"  # merged into the single sync commit (3-phase close)
  updated_field_refreshed: "2026-08-07"
changelog_entry_position: "[Unreleased] / Added"   # SPEC-NAVIGATOR-SYNC-003 entry appended
readme_decision: skip                              # internal Navigator subsystem; navigator-tiers is Hidden + no user-facing command surface (consistent with NS1/NS2/NS3 predecessor precedent)
docs_site_decision: skip                           # internal subsystem; no docs-site 4-locale page (consistent with predecessor precedent)
mx_tag_validation: sub-step-complete               # MX tag validation is a sync sub-step (no separate Mx-phase commit)
ac_pass_count_final: 22                            # AC-NS3-001..022 — all PASS post Chunk 3 (M4.6+M4.7 closed the 2 prior DEFERRED AC)
ac_fail_count_final: 0
coverage_pct_final: 85.0                           # internal/navigator/tiers/... — ≥85% AC target MET
open_blockers: 0
```

## §F Phase 4 Mode Selection

**Input parameters**
- tier: L
- scope (file count est.): ~14 implementation files (`internal/navigator/tiers/` ~9 Go files + `internal/cli/navigator_tiers.go` + 5 template surfaces)
- domain count: 4 (navigator/tiers Go package, CLI wiring, template-first surfaces, SPEC docs)
- file language mix: Go source (primary) + markdown templates + JSON/YAML schemas
- concurrency benefit: LOW (coding-heavy, sequential dependency chain: schema → tier engines → overlay join → close)
- Agent Teams prereqs: N/A (Mode 3 retired)

**Mode evaluation**
| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | substantial new code, 7 milestones |
| 2 background | no | write-capable, result needed before continuing |
| 3 agent-team | no | RETIRED |
| 4 parallel | no | coding-heavy with sequential dependency chain; Anthropic coding-task parallelism caveat |
| 5 sub-agent | YES | coding-heavy sequential, default for new-code work |
| 6 workflow | no | new-code + multi-rule, not mechanical-uniform transform; <30 files |

**Decision**: sub-agent (Mode 5), cycle_type=tdd, milestone-coherent chunks with orchestrator trust-but-verify gates between.

**Justification**: This is coding-heavy new-code work (a new `internal/navigator/tiers/` package with a sequential dependency chain: schema → three tier engines → overlay integration → template-first close). Per Anthropic's coding-task parallelism caveat, sequential sub-agent delegation is the correct default for coding work. Semi-autonomous progression: the orchestrator delegates in milestone-coherent chunks (M4.1+M4.2 foundation → M4.3+M4.4+M4.5 engines → M4.6+M4.7 integration+close), running a trust-but-verify batch at each chunk boundary; `/moai goal` is NOT armed (per-session milestone gates instead). Route B (Tier L, feature branch `feat/SPEC-NAVIGATOR-SYNC-003` + PR per repo-local `enforce_admins` policy). Implementation Kickoff Approval PASSED (user-approved). Phase 1 plan-audit re-execution skip-eligible (PASS-WITH-DEBT 0.85 ≥ Tier L 0.85 threshold, artifacts unchanged since merged plan PR #1383).
