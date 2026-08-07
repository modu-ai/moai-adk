# SPEC-NAVIGATOR-SYNC-001 — Design

> Tier L design artifact. Records the schema decisions and the resolved pre-flight defaults.
> Cross-references spec.md §C (REQs), plan.md §C (pre-flight decisions), acceptance.md §B (ACs).

## §1. Resolved Pre-flight Decisions (was [NEEDS CLARIFICATION] in plan.md §C)

### D1 — SpecAssociator consumer reshape location → consumer-side

**Decision**: `internal/navigator/sync/mx_bridge.go`, NOT `internal/mx/`.

**Rationale**: REQ-NS-005 forbids modifying `internal/mx/`. An additive helper inside `internal/mx/spec_association.go` would blur the consumer-only boundary even if technically additive. The bridge belongs on the consumer side.

**Implementation**: `mx_bridge.go` calls `spec_association.NewSpecAssociator(specModules).AssociateWithDiagnostics(tag)` per tag produced by `mx.Scanner` (driven by `internal/mx/scanner.go`'s existing scan path), then inverts the `tag → []specID` map into `specID → []tagLocation` for graph consumption. The `specModules` input (SPEC ID → module paths) is sourced from the same `.moai/specs/` registry walk that 001 uses.

### D2 — Symbol-node canonical form → language-aware package-qualified

**Decision**: Go symbols canonicalize as `<package-path>/<Name>` (e.g. `internal/navigator/sync.Join`); other languages use the bare identifier initially.

**Resolution rule for `@NAV:SYM:<symbol>` tokens authored in design docs or code comments**:

1. If `<symbol>` exactly equals a `PrimarySymbols[].Name` from 003's `capability-symbols.json`, the token resolves to THAT symbol node (dedup).
2. Else if `<symbol>` matches a `PrimarySymbols[].Name` by suffix (last `.`-segment), the token resolves to that node (lenient match, logged at debug).
3. Else, emit a NEW symbol node with the token-authored form as both `identifier` and `display_name` (forward-compatible — M3 may resolve later).

**Rationale**: 003's `PrimarySymbols[].Name` is the existing deterministic symbol identifier; reusing it avoids two parallel symbol taxonomies. The lenient suffix-match covers the common authoring style (`@NAV:SYM:Join` instead of `@NAV:SYM:internal/navigator/sync.Join`).

### D3 — Design-doc scan root → bounded set

**Decision**: scan `.moai/project/{product,structure,tech}.md` (3 fixed files) + `.moai/docs/**/*.md` (recursive).

**Excluded**:
- `.moai/specs/` — already covered by mx-scanner body-based association (re-scanning would double-count).
- `.moai/reports/` — ephemeral reports; not a design-doc surface.
- `.moai/state/`, `.moai/state/lsel/` — REQ-NS-013 LSEL non-overlap.
- `internal/`, `cmd/`, `pkg/` — these are CODE, scanned separately for `@NAV:SYM`.

**Rationale**: the design surface is precisely the 3 project-scope docs + the docs tree. Bounding the scan keeps M0 fast and avoids accidentally surfacing transient reports as design decisions.

## §2. `nav-graph.json` Schema (normative)

```json
{
  "provenance": {
    "extract_commit_sha": "<git rev-parse HEAD>",
    "captured_at": "<git log -1 --format=%cI of HEAD>"
  },
  "nodes": [
    {
      "entity_type": "decision",
      "identifier": "AUTH-STRATEGY",
      "display_name": "Adopt OAuth2"
    },
    {
      "entity_type": "spec",
      "identifier": "SPEC-PROJECT-NAVIGATOR-001",
      "display_name": "Project Navigator — living docs"
    },
    {
      "entity_type": "symbol",
      "identifier": "internal/navigator/sync.Join",
      "display_name": "Join"
    }
  ],
  "edges": [
    {
      "edge_type": "dec-edge",
      "source_node": "decision:AUTH-STRATEGY",
      "target_node": "spec:SPEC-AUTH-001",
      "source_path": ".moai/project/tech.md",
      "line_number": 42
    },
    {
      "edge_type": "spec-edge",
      "source_node": "symbol:internal/runtime/budget.go",
      "target_node": "spec:SPEC-V3R3-ARCH-007",
      "source_path": "internal/runtime/budget.go",
      "line_number": 22
    },
    {
      "edge_type": "sym-edge",
      "source_node": "symbol:pkg.ParseHeader",
      "target_node": "decision:HEADER-PARSER",
      "source_path": "internal/navigator/sync/scan_sym.go",
      "line_number": 5
    }
  ]
}
```

**Node identifier scheme** — each node's `(entity_type, identifier)` pair is the composite primary key. Edges reference nodes via `<entity_type>:<identifier>` (the `source_node` / `target_node` fields).

**Forward-compatibility** — fields are additive only. Future milestones may add fields (e.g. M1 may add a `last_modified_commit` to edges); existing fields keep their names and shapes.

## §3. Architecture (data flow)

```
                                  ┌──────────────────────────────────┐
                                  │   internal/navigator/sync/       │
                                  │   (NEW — this SPEC)              │
                                  │                                  │
  001 capability-map.md ──────────►│  parseCapabilityMap (reuse 003)  │
  (header-driven)                  │  → spec-node discovery           │
                                  │                                  │
  003 capability-symbols.json ────►│  symbol-node source-of-truth     │
                                  │  → symbol-node set               │
                                  │                                  │
  002 audit-report.json ──────────►│  advisory read (graceful skip)   │
  (may be absent)                  │                                  │
                                  │                                  │
  @NAV:DEC-<id> scanner ──────────►│  dec-binding records             │
  (design docs)                    │  → decision nodes + dec-edges    │
                                  │                                  │
  @NAV:SYM:<symbol> scanner ──────►│  sym-binding records             │
  (code + design)                  │  → sym-edges (and new symbol     │
                                  │    nodes per D2 resolution rule)  │
                                  │                                  │
  internal/mx SpecAssociator ─────►│  spec-associations (per-tag)     │
  (CONSUMED, NOT modified)         │  → spec-edges                    │
                                  │                                  │
                                  │  Provenance (git baseline)        │
                                  │  Atomic write (renames .tmp)     │
                                  │  Fail-open capability gate        │
                                  │                                  │
                                  └──────────────┬───────────────────┘
                                                 │
                                                 ▼
                                  .moai/project/navigator/nav-graph.json
                                  (single graph artifact, byte-stable)
```

## §4. Why `@MX:SPEC` is bridged, not absorbed (precision correction)

The design report's lean twin (§3 "제안 BAS (A)") says: "기존 `@MX:SPEC`은 현재 Navigator가 소비 안 함 → 흡수." On inspection of `main`:

- `internal/mx/scanner.go:139-155` DOES consume `@MX:SPEC` sub-lines (REQ-MX-ASSOC-001, SPEC-MX-ASSOCIATION-001).
- `internal/mx/spec_association.go:17-20` unifies three sources: path-based, body-based, sub-line `@MX:SPEC`.

The lean twin's claim is therefore imprecise: `@MX:SPEC` is NOT dormant — the **mx-scanner** consumes it, the **Navigator** does not. M0's job is to bridge the mx-scanner's association output into the Navigator graph, so a code→SPEC edge in `nav-graph.json` reflects ALL three mx-scanner sources. M0 does NOT modify `internal/mx/`. This is recorded in spec.md HISTORY + REQ-NS-005 to prevent an implementer from "absorbing" (re-scanning) `@MX:SPEC` and creating a parallel taxonomy.

## §5. Non-overlap invariant test design

Per `feedback_deadfield_grep_pair_protocol.md`, each non-overlap AC uses two grep lenses (source + runtime). The runtime lens needs a filesystem-audit helper that snapshots the temp-dir tree before and after the join runs, diffs the snapshots, and asserts the only new path is `nav-graph.json`. This helper lives in `internal/navigator/sync/nonoverlap_test.go` and is reused by AC-013/014/015/016.

The grep lens for LSEL (AC-013) uses the literal path fragments from REQ-NS-013: `lessons-inbox`, `state/lsel`, `memory/feedback_`, `hns-lsel`, `state/lsel/proposals`. A bare `lsel` substring is FORBIDDEN as a grep pattern (it would match the SPEC's own non-overlap tests in `internal/navigator/sync/nonoverlap_test.go`, creating a false positive).

**Source-comment hygiene (D2 advisory, parallel to the bare-`lsel` rule)** — source comments and string literals inside `internal/navigator/sync/` MUST NOT literally name the forbidden LSEL paths (`lessons-inbox`, `state/lsel`, `memory/feedback_`, `hns-lsel`). A source file that mentions a forbidden path in a comment would trip AC-013 Lens 1's source-grep even though the binary never touches that path at runtime — a false positive that undermines the two-lens protocol. The non-overlap intent is expressed at the requirement level (REQ-NS-013) and in test-fixture assertions, never as literal forbidden-path strings in production source comments. The AC-013 Lens 1 patterns stay as-is; this note adds the defensive source-discipline obligation that keeps Lens 1 mechanically sound.

## §6. Performance envelope

- Scan root for `@NAV:DEC`: ≤ 5 files (`.moai/project/{3}.md` + `.moai/docs/**/*.md` typically < 50 files).
- Scan root for `@NAV:SYM`: `.moai/project/**/*.md` + `.go` files containing the literal `@NAV:SYM:` substring (pre-filter via ripgrep-style scan, then parse matches only).
- Mx-bridge: one `AssociateWithDiagnostics` call per tag produced by `mx.Scanner.Scan(root)`. The tag count is bounded by the existing mx-scanner's performance — NOT M0's concern.
- Join: O(N+M) where N = total bindings, M = total spec-associations.
- Output: single JSON file, typically < 100KB for a project this size.

No performance ACs in M0; performance characterization belongs to M1 (Detect hook, where latency matters).

## §7. Open questions for M1 (forward-looking, NOT M0 blockers)

- Should the join step run incrementally (M1 PostToolUse hook) or full-scan each time? M0 is full-scan; M1 may add an incremental path.
- Should `nav-graph.json` be gitignored or committed? M0 emits it under `.moai/project/navigator/` alongside `capability-map.md`. The `.gitignore` policy is set by 001 (whichever it chose for capability-map.md, M0 mirrors — recorded here so M0 does not re-decide).
- Should the graph support a `query` subcommand for ad-hoc lookups? M0 emits the artifact only; query tooling is a forward milestone.
