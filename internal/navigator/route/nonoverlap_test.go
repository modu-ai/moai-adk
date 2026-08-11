package route

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	navsync "github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// TestNonOverlap_ForbiddenWriteTargets exercises AC-NS4-005b/006: the
// Route production source files MUST NOT write to forbidden surfaces
// (internal/navigator/sync/, detect/, tiers/, hook/, mx/, or the three
// predecessor chains' outputs). This test scans the route package source
// for forbidden path fragments appearing in a write context.
func TestNonOverlap_ForbiddenWriteTargets(t *testing.T) {
	t.Parallel()

	// Forbidden path fragments that MUST NOT appear as write targets in
	// the Route source (carries forward M0's nonoverlap_test.go pattern
	// from internal/navigator/sync/nonoverlap_test.go).
	forbiddenWriteTargets := []string{
		"capability-map.md",
		"capability-symbols",
		"nav-graph.json",
		"tiers.json",
		"audit-report.md",
		"audit-report.json",
		"blueprint/",
		"decisions/",
		"navigator/symbols/",
	}

	matches, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range matches {
		// Skip test files — they name these fragments as assertion targets.
		if strings.HasSuffix(m, "_test.go") {
			continue
		}
		b, err := os.ReadFile(m)
		if err != nil {
			t.Fatal(err)
		}
		src := string(b)
		for _, f := range forbiddenWriteTargets {
			// Allow READ references (os.ReadFile, json.Unmarshal on a
			// variable). Flag only if the fragment appears alongside a
			// write verb (os.WriteFile, os.Rename, os.OpenFile with
			// O_WRONLY/O_CREATE).
			if strings.Contains(src, f) {
				// Check if it's in a write context.
				lines := strings.Split(src, "\n")
				for lineNo, line := range lines {
					if !strings.Contains(line, f) {
						continue
					}
					trimmed := strings.TrimSpace(line)
					// Skip comments and string literals that document the path.
					if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
						continue
					}
					// Flag if the line contains a write verb.
					if strings.Contains(line, "WriteFile") || strings.Contains(line, "Rename") ||
						strings.Contains(line, "OpenFile") {
						t.Errorf("%s:%d: forbidden write target %q appears alongside a write verb",
							filepath.Base(m), lineNo+1, f)
					}
				}
			}
		}
	}
}

// TestNonOverlap_NoPredecessorMutation asserts the Route source does NOT
// import or call any mutating function from the predecessor packages
// (sync, detect, tiers, hook, mx). The Route layer is consumer-only
// (REQ-NS4-005).
func TestNonOverlap_NoPredecessorMutation(t *testing.T) {
	t.Parallel()

	// The Route source MAY import navsync for READ-ONLY types
	// (Graph, Edge, Node). It MUST NOT call any Write/Mutate/Run function
	// from those packages.
	forbiddenCalls := []string{
		"sync.WriteGraph",
		"sync.Run(",
		"sync.atomicWrite",
		"detect.Traverse",
		"tiers.Enrich",
	}

	matches, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range matches {
		if strings.HasSuffix(m, "_test.go") {
			continue
		}
		b, err := os.ReadFile(m)
		if err != nil {
			t.Fatal(err)
		}
		src := string(b)
		for _, call := range forbiddenCalls {
			if strings.Contains(src, call) {
				t.Errorf("%s: forbidden predecessor call %q (Route is consumer-only)",
					filepath.Base(m), call)
			}
		}
	}
}

// TestNonOverlap_OwnerIsPathInvariant exercises AC-NS4-004: every owner_path
// across the corpus MUST be a code/doc path — NEVER a person, team, email,
// or CODEOWNERS reference. The grep-guard checks for '@', 'mailto:',
// and 'users/' patterns across ALL promoted work items.
func TestNonOverlap_OwnerIsPathInvariant(t *testing.T) {
	t.Parallel()

	const root = "/fixture-project"
	graph := &navsync.Graph{
		Nodes: []navsync.Node{
			{EntityType: navsync.EntitySymbol, Identifier: "auth.ParseBearer"},
		},
		Edges: []navsync.Edge{
			symEdge(root+"/.moai/project/tech.md", "symbol:auth.ParseBearer", 10),
			symEdge(root+"/internal/auth/login.go", "symbol:auth.ParseBearer", 5),
		},
	}
	audit := &AuditReport{
		Missing: []MissingEntry{
			{DesignName: "Auth", Source: AuditSource{File: ".moai/project/tech.md"}},
		},
		Orphan: []OrphanEntry{
			{SpecID: "SPEC-X-001", Title: "X", ImplementationPath: "internal/x.go"},
			{SpecID: "SPEC-X-002", Title: "Y", ImplementationPath: ""},
		},
	}
	detectRows := []DetectRecord{
		{ChangedPath: root + "/internal/auth/login.go"},
	}

	items := Promote(audit, detectRows, graph, root)

	for i, item := range items {
		if strings.Contains(item.OwnerPath, "@") {
			t.Errorf("item[%d]: owner_path %q contains '@' (person reference forbidden)", i, item.OwnerPath)
		}
		if strings.Contains(item.OwnerPath, "mailto:") {
			t.Errorf("item[%d]: owner_path %q contains 'mailto:' (email forbidden)", i, item.OwnerPath)
		}
		if strings.Contains(item.OwnerPath, "users/") {
			t.Errorf("item[%d]: owner_path %q contains 'users/' (CODEOWNERS-style forbidden)", i, item.OwnerPath)
		}
		if strings.Contains(item.OwnerPath, "git blame") {
			t.Errorf("item[%d]: owner_path %q contains 'git blame' (person binding forbidden)", i, item.OwnerPath)
		}
	}
}
