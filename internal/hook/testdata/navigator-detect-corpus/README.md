# navigator-detect-corpus — Fixture Corpus (AC-NS2-007, REQ-NS2-007)

Deterministic fixture corpus for `TestNavigatorDetectCoverage`
(`internal/hook/navigator_detect_coverage_test.go`). Measures the M1 Detect
layer's **mapping coverage**: of a corpus of changed paths that fall inside the
M0 graph's scan roots, what fraction maps to a non-empty affected-row set when
a matching graph row exists.

This is NOT the unit-test coverage % (the 88.6% figure from M1.1 is
`go test -cover` over the `detect` package). The two numbers measure different
things: the unit-test % measures line coverage of the Traverse code; the
mapping coverage % measures the layer's correctness over a realistic corpus of
changed-path inputs.

## Files

- `nav-graph.json` — the pre-built graph fixture (19 edges spanning dec-edge /
  spec-edge / sym-edge, absolute `source_path` values under `/abs/project/`).
  This is the graph the M0 producer WOULD emit if it scanned the in-scope
  corpus. The `/abs/project/` placeholder is documented per-case in
  `corpus_cases.json`; `filepath.Abs` is idempotent on already-absolute paths,
  so no TempDir is needed for the path-matching engine.
- `corpus_cases.json` — the manifest of 26 cases (21 in-scope + 5
  out-of-scope). Each case carries `changed_path` + `class` + a human note.
- `README.md` (this file) — corpus documentation.

## Case taxonomy (plan.md §E partition)

| class | numerator? | denominator? | Traverse result |
|---|---|---|---|
| `in-scope-mapped` | yes | yes | ≥ 1 row (edge matched via exact equality or directory-prefix fallback REQ-NS2-010) |
| `in-scope-unmapped` | no | yes | 0 rows (path is in-scope by location but no edge indexes it — a real graph never achieves 100% because not every file carries a binding token) |
| `out-of-scope` | no | no | 0 rows (path is outside the M0 scan roots — `.moai/specs/`, `.claude/`, root dotfiles, etc.) |

## Expected ratio

```
coverage = (in-scope-mapped) / (in-scope-mapped + in-scope-unmapped)
         = 18 / (18 + 3)
         = 0.857  (≥ 0.80 threshold → PASS)
```

The 3 in-scope-unmapped cases exercise the denominator deliberately: they are
sibling files in directories the graph indexes (`internal/auth/session.go`,
`internal/middleware/logger.go`, `internal/config/env.go`) but carry no binding
token, so the graph does not index them. A realistic graph never achieves 100%
mapping coverage; the 80% threshold (REQ-NS2-007) leaves headroom for this.

## Why fixture-driven, not repo-driven (plan.md §E)

A repo-driven measurement would re-scan the live repo and produce a number that
drifts with every commit (every edit changes the denominator). A fixture corpus
is deterministic and isolates the Detect layer's correctness from the repo's
current shape — it measures the LAYER, not the repository. Two runs of
`go test ./internal/hook/ -run TestNavigatorDetectCoverage -v` on the same
HEAD produce the same percentage.

## Verification command (acceptance.md §D AC-NS2-007)

```
go test ./internal/hook/ -run TestNavigatorDetectCoverage -v
```

The test emits the observed coverage percentage on the PASS line (or a
failure message with the observed value if `< 80.0`). The percentage printed
is the Evidence; the ≥80% assertion is the Claim (verification-claim-integrity
§2 attribution).
