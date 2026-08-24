package cli

import (
	"context"
	"os"
	"sort"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	mcpcat "github.com/modu-ai/moai-adk/internal/mcp"
)

// effectiveReadOnlyHint returns a tool's EFFECTIVE read-only value: the
// declared annotation when present, else the MCP default false. This is the
// REQ-CW-011 baseline — the guard compares EFFECTIVE values, not declaration
// presence, so catalog-write tools left undeclared (goal_arm,
// verify_snapshot) pass while any catalog-read tool whose effective value is
// false fails.
func effectiveReadOnlyHint(t *mcp.Tool) bool {
	if t == nil || t.Annotations.ReadOnlyHint == nil {
		return false // undeclared → MCP default (not read-only)
	}
	return *t.Annotations.ReadOnlyHint
}

// listToolsWithAnnotations constructs an in-process MCP server, runs
// initialize + tools/list, and returns the registered tools (name + effective
// read-only hint).
func listToolsWithAnnotations(t *testing.T) map[string]bool {
	t.Helper()
	srv := newMoaiMCPServer()
	if srv == nil {
		t.Fatal("newMoaiMCPServer returned nil")
	}
	ctx := context.Background()
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	res, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("tools/list: %v", err)
	}
	out := make(map[string]bool, len(res.Tools))
	for i := range res.Tools {
		out[res.Tools[i].Name] = effectiveReadOnlyHint(&res.Tools[i])
	}
	return out
}

// TestMoaiMCPServer_AnnotationsMatchCatalog is the REQ-CW-011 / AC-CW-010
// guard: for EVERY tool in the catalog, the EFFECTIVE read-only annotation
// (declared value, else default false) must equal !WriteCapable.
//
// Why this baseline: Codex's `default_tools_approval_mode = "writes"` prompts
// for tools NOT marked read-only — capability-based, never tool-name
// enumeration (spec §A.4). The approval set is therefore only as correct as
// the per-tool effective annotations: a catalog-READ tool with effective
// false inflates the approval set (base-tree defect: audit_cache, codex_audit,
// glm_audit, audit_multi), and a catalog-WRITE tool marked read-only would
// silently escape approval entirely.
func TestMoaiMCPServer_AnnotationsMatchCatalog(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	got := listToolsWithAnnotations(t)

	var mismatched []string
	for _, def := range mcpcat.MoaiMCPTools() {
		effective, registered := got[def.Name]
		if !registered {
			t.Errorf("catalog tool %q missing from tools/list (registration drift)", def.Name)
			continue
		}
		if effective == def.WriteCapable {
			// effective read-only must be the NEGATION of write-capable
			mismatched = append(mismatched, def.Name)
			t.Errorf("tool %q: effective read-only=%t but catalog WriteCapable=%t (REQ-CW-011 equivalence: effective read-only ⟺ WriteCapable=false)",
				def.Name, effective, def.WriteCapable)
		}
	}
	sort.Strings(mismatched)
	if len(mismatched) > 0 {
		t.Fatalf("%d tool(s) violate the effective-annotation ⟺ catalog equivalence: %v", len(mismatched), mismatched)
	}
}

// TestMoaiMCPServer_AnnotationGuardTeeth documents the guard's two failure
// shapes with synthetic values (AC-CW-010 negative assertions): a missing
// annotation on a catalog-read tool and a read-only declaration on a
// catalog-write tool are BOTH mismatches under the effective-value baseline,
// while an undeclared catalog-write tool is NOT.
func TestMoaiMCPServer_AnnotationGuardTeeth(t *testing.T) {
	type synthetic struct {
		name          string
		declared      *bool // nil = undeclared
		writeCapable  bool
		expectMatchOK bool
	}
	no, yes := false, true
	cases := []synthetic{
		{"read tool, declared true", &yes, false, true},
		{"read tool, undeclared (effective false)", nil, false, false}, // the base-tree 4-tool defect shape
		{"write tool, declared false", &no, true, true},                // codex_task & family
		{"write tool, undeclared (effective false)", nil, true, true},  // goal_arm / verify_snapshot — must PASS
		{"write tool, declared read-only", &yes, true, false},          // the silent-approval-escape shape
	}
	for _, tc := range cases {
		tool := &mcp.Tool{Annotations: mcp.ToolAnnotation{ReadOnlyHint: tc.declared}}
		effective := effectiveReadOnlyHint(tool)
		match := effective != tc.writeCapable // equivalence: effective ⟺ !writeCapable
		if match != tc.expectMatchOK {
			t.Errorf("%s: guard verdict = %t, want %t (effective=%t writeCapable=%t)",
				tc.name, match, tc.expectMatchOK, effective, tc.writeCapable)
		}
	}
}

// TestMoaiMCPServer_AnnotationGuardNoEnvLeak pins the guard's hygiene: it
// runs against a clean temp project root, never the developer's real config.
func TestMoaiMCPServer_AnnotationGuardNoEnvLeak(t *testing.T) {
	if _, err := os.Stat("CLAUDE_PROJECT_DIR"); err == nil {
		t.Skip("unexpected file named CLAUDE_PROJECT_DIR in cwd")
	}
}
