package route

import (
	"strings"
	"testing"

	navsync "github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// TestOwnerResolution exercises the three owner-resolution paths × confidence
// levels (AC-NS4-004, table-driven 004a-004e). Every owner_path MUST be an
// absolute path — never a person, team, or email (falconer binding).
func TestOwnerResolution(t *testing.T) {
	t.Parallel()

	const root = "/test-project"

	// Graph fixture for 004c (missing-via-symbol): the design doc tech.md
	// references @NAV:SYM:auth.ParseBearer, and the symbol is declared in
	// internal/auth/login.go.
	symbolGraph := &navsync.Graph{
		Nodes: []navsync.Node{
			{EntityType: navsync.EntitySymbol, Identifier: "auth.ParseBearer", DisplayName: "ParseBearer"},
		},
		Edges: []navsync.Edge{
			// Design-doc → symbol (the @NAV:SYM reference in tech.md).
			{
				EdgeType:   navsync.EdgeSym,
				SourceNode: "decision:AUTH",
				TargetNode: "symbol:auth.ParseBearer",
				SourcePath: "/test-project/.moai/project/tech.md",
				LineNumber: 42,
			},
			// Code → symbol (the @NAV:SYM declaration in login.go).
			{
				EdgeType:   navsync.EdgeSym,
				SourceNode: "symbol:auth.ParseBearer",
				TargetNode: "symbol:auth.ParseBearer",
				SourcePath: "/test-project/internal/auth/login.go",
				LineNumber: 15,
			},
		},
	}

	// Graph fixture for 004d (missing-doc-fallback): no sym-edges at all.
	emptyGraph := &navsync.Graph{Nodes: nil, Edges: nil}

	tests := []struct {
		name        string
		sourceKind  SourceKind
		ownerPath   string
		confidence  Confidence
		personGuard bool // if true, owner_path MUST NOT contain person patterns
	}{
		// 004a — orphan-direct: non-empty implementation_path → high.
		{
			name:       "004a orphan-direct",
			sourceKind: SourceAuditOrphan,
			ownerPath:  "/test-project/internal/foo.go",
			confidence: ConfidenceHigh,
		},
		// 004b — orphan-empty-impl: empty implementation_path → SPEC-dir, low.
		{
			name:       "004b orphan-empty-impl",
			sourceKind: SourceAuditOrphan,
			ownerPath:  "/test-project/.moai/specs/SPEC-FOO-002/",
			confidence: ConfidenceLow,
		},
		// 004c — missing-via-symbol: @NAV:SYM resolves via graph → medium.
		{
			name:       "004c missing-via-symbol",
			sourceKind: SourceAuditMissing,
			ownerPath:  "/test-project/internal/auth/login.go",
			confidence: ConfidenceMedium,
		},
		// 004d — missing-doc-fallback: no symbol → source.file, low.
		{
			name:       "004d missing-doc-fallback",
			sourceKind: SourceAuditMissing,
			ownerPath:  "/test-project/.moai/project/tech.md",
			confidence: ConfidenceLow,
		},
		// 004e — detect-direct: changed_path → high.
		{
			name:       "004e detect-direct",
			sourceKind: SourceDetect,
			ownerPath:  "/test-project/internal/auth/login.go",
			confidence: ConfidenceHigh,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var gotOwner string
			var gotConf Confidence

			switch tc.sourceKind {
			case SourceAuditOrphan:
				if tc.name == "004a orphan-direct" {
					gotOwner, gotConf = resolveOrphanOwner(OrphanEntry{
						SpecID:             "SPEC-FOO-001",
						Title:              "Foo feature",
						ImplementationPath: "internal/foo.go",
					}, root)
				} else {
					gotOwner, gotConf = resolveOrphanOwner(OrphanEntry{
						SpecID:             "SPEC-FOO-002",
						Title:              "Bar feature",
						ImplementationPath: "",
					}, root)
				}
			case SourceAuditMissing:
				if tc.name == "004c missing-via-symbol" {
					gotOwner, gotConf = resolveMissingOwner(MissingEntry{
						DesignName: "OAuth2",
						Source:     AuditSource{File: ".moai/project/tech.md", HeadingPath: "## Auth"},
					}, symbolGraph, root)
				} else {
					gotOwner, gotConf = resolveMissingOwner(MissingEntry{
						DesignName: "Unknown feature",
						Source:     AuditSource{File: ".moai/project/tech.md", HeadingPath: "## Unknown"},
					}, emptyGraph, root)
				}
			case SourceDetect:
				gotOwner, gotConf = resolveDetectOwner(DetectRecord{
					ChangedPath: "/test-project/internal/auth/login.go",
					ChangedAt:   "2026-01-01T00:00:00Z",
				}, root)
			}

			if gotOwner != tc.ownerPath {
				t.Errorf("owner_path: got %q, want %q", gotOwner, tc.ownerPath)
			}
			if gotConf != tc.confidence {
				t.Errorf("confidence: got %q, want %q", gotConf, tc.confidence)
			}
			if !strings.HasPrefix(gotOwner, "/") {
				t.Errorf("owner_path %q is not absolute", gotOwner)
			}
			// Owner-is-path invariant: no person patterns (AC-NS4-004).
			if strings.ContainsAny(gotOwner, "@") ||
				strings.Contains(gotOwner, "mailto:") ||
				strings.Contains(gotOwner, "users/") {
				t.Errorf("owner_path %q contains a person reference (forbidden)", gotOwner)
			}
		})
	}
}

// TestOwnerNilGraph exercises the nil-graph fail-open path for a missing
// entry: when the M0 graph is absent, owner resolution degrades to the
// design-doc source.file with low confidence (REQ-NS4-009 row 009c, but
// tested here at the owner-resolution level).
func TestOwnerNilGraph(t *testing.T) {
	t.Parallel()

	owner, conf := resolveMissingOwner(MissingEntry{
		DesignName: "Some feature",
		Source:     AuditSource{File: ".moai/project/tech.md"},
	}, nil, "/test-project")

	if conf != ConfidenceLow {
		t.Errorf("nil graph: confidence = %q, want low", conf)
	}
	if owner != "/test-project/.moai/project/tech.md" {
		t.Errorf("nil graph: owner = %q, want design-doc path", owner)
	}
}
