# M1 RED evidence — SPEC-CODEMAPS-ACCURACY-001 (t304)

Command (from worktree root `.claude/worktrees/t304`, this run):

```
go test ./internal/graph/ -run TestCheckCitations -count=1
```

Exit code: 1

Verbatim output (pre-GREEN, tests reference the not-yet-implemented citations axis):

```
# github.com/modu-ai/moai-adk/internal/graph [github.com/modu-ai/moai-adk/internal/graph.test]
internal/graph/check_citations_test.go:53:9: undefined: checkCitations
internal/graph/check_citations_test.go:54:18: undefined: LayerCitations
internal/graph/check_citations_test.go:55:46: undefined: LayerCitations
internal/graph/check_citations_test.go:57:19: undefined: MetricPositiveCitedPathAbsence
internal/graph/check_citations_test.go:58:48: undefined: MetricPositiveCitedPathAbsence
internal/graph/check_citations_test.go:73:74: undefined: LayerCitations
internal/graph/check_citations_test.go:82:10: undefined: checkCitations
internal/graph/check_citations_test.go:106:9: undefined: checkCitations
internal/graph/check_citations_test.go:130:13: undefined: normalizeCitedPath
internal/graph/check_citations_test.go:140:9: undefined: checkCitations
internal/graph/check_citations_test.go:140:9: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/graph [build failed]
FAIL
```

Baseline-attribution: tree `.claude/worktrees/t304` @ `061985ec8` (WT-codemaps-accuracy), 2026-09-02, this run. RED-before-GREEN is falsifiable via this item.
