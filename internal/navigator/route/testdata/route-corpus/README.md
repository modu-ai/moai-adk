# Route Accuracy Fixture Corpus (AC-NS4-010)

This directory documents the fixture corpus shape for the ≥70% Route accuracy
measurement (SPEC-NAVIGATOR-SYNC-004 REQ-NS4-010, plan.md §E).

## Corpus Design

| Source | Count | With binding | Confidence | Actionable? |
|--------|-------|-------------|------------|-------------|
| audit-missing (with @NAV:SYM) | 3 | symbol → code via graph | medium | YES |
| audit-missing (no @NAV:SYM) | 3 | design-doc fallback | low | no |
| audit-orphan (with impl_path) | 10 | implementation_path | high | YES |
| audit-orphan (empty impl_path) | 2 | SPEC-dir fallback | low | no |
| detect (unique changed_path) | 12 | changed_path | high | YES |
| **Total** | **30** | | | |

## Accuracy Dual-Arithmetic

- **Happy path** (whole-doc symbol lookup works): 3 medium + 10 high-orphan + 12 high-detect = 25 actionable. 25/30 = **83.3%**.
- **B5 fallback** (all medium collapse to low): 0 + 10 + 12 = 22 actionable. 22/30 = **73.3%**.

Both paths are ≥ 70.0% — the floor survives the fallback with 3.3pp headroom.

The actual fixture data is constructed in-memory by `TestRouteAccuracy` in
`coverage_test.go` (avoids static-file path-resolution issues — the graph
edges use absolute paths that depend on the test root).
