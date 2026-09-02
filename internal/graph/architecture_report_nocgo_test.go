//go:build !cgo

package graph

import (
	"strings"
	"testing"
)

// AC-GR-010 (SPEC-GRAPH-REPORT-001, B5 nocgo semantics): under a CGO-disabled
// build the code layer is absent — zero code-call edges reach the artifact —
// and the report still emits with every section present, the code-dependent
// sections present-but-empty carrying the stated reason, and no error
// surface. Mirrors shortestpath_nocgo_test.go's pattern.
func TestNoCGOArchitectureReportEmptySections(t *testing.T) {
	root := tierFixture(t)
	all, _, _, err := BuildWithCodeLayers(root)
	if err != nil {
		t.Fatalf("BuildWithCodeLayers under !cgo: %v", err)
	}
	for _, e := range all {
		if e.Kind == KindCodeCall {
			t.Fatalf("code-call edge under !cgo: %+v", e)
		}
	}

	body := RenderArchitectureReport(all, 0)
	for _, want := range []string{
		"## God Nodes",
		"## Surprising Connections",
		"## Import Cycles",
		"code layer absent: CGO disabled or no extraction",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("nocgo report missing %q; body:\n%s", want, body)
		}
	}
}
