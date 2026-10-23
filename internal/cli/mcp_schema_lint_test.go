// mcp_schema_lint_test.go — SPEC-MCP-NEUTRALITY-001 M2 (card t1632,
// REQ-MN-007): one lint sweep over every SERVED tool's input schema. A
// violation is a lint finding, and the sweep fails while any stands:
//
//	R1 additionalProperties carries an explicit policy (never unset);
//	R2 every declared property carries a description;
//	R3 every required name is a declared property;
//	R4 a read-only tool never carries an explicit destructive hint.
//
// The handler-read-argument half of REQ-MN-007 (an argument the handler reads
// must be declared) is not statically derivable from the schemas alone; the
// audit's named instance — codex_task reading an undeclared model — is pinned
// by TestMCPSchemaLint_HandlerReadArgumentsDeclared.
package cli

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// lintServedTools builds the served surface and returns the tools.
func lintServedTools(t *testing.T) []mcp.Tool {
	t.Helper()
	srv := newMoaiMCPServer()
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
		t.Fatalf("ListTools: %v", err)
	}
	return res.Tools
}

func TestMCPSchemaLint_ServedTools(t *testing.T) {
	tools := lintServedTools(t)
	if len(tools) == 0 {
		t.Fatal("the sweep served zero tools — a green lint over nothing asserts nothing")
	}
	var findings []string
	for _, tool := range tools {
		s := tool.InputSchema
		if s.AdditionalProperties == nil {
			findings = append(findings, fmt.Sprintf("R1 %s: additionalProperties not declared", tool.Name))
		}
		for pname, p := range s.Properties {
			pm, ok := p.(map[string]any)
			if !ok {
				continue
			}
			if _, has := pm["description"]; !has {
				findings = append(findings, fmt.Sprintf("R2 %s.%s: no description", tool.Name, pname))
			}
		}
		for _, r := range s.Required {
			if _, ok := s.Properties[r]; !ok {
				findings = append(findings, fmt.Sprintf("R3 %s: required %q is not a declared property", tool.Name, r))
			}
		}
		readOnly := tool.Annotations.ReadOnlyHint != nil && *tool.Annotations.ReadOnlyHint
		destructive := tool.Annotations.DestructiveHint != nil && *tool.Annotations.DestructiveHint
		if readOnly && destructive {
			findings = append(findings, fmt.Sprintf("R4 %s: read-only tool carries a destructive hint", tool.Name))
		}
	}
	if len(findings) > 0 {
		t.Fatalf("schema lint findings (%d):\n%s", len(findings), strings.Join(findings, "\n"))
	}
}

// TestMCPSchemaLint_HandlerReadArgumentsDeclared pins the audit's named
// handler-read drift: goal_status declared a required field its enforcement
// does not require (codex_task, the audit's other named instance, was removed
// with card t1909). The lint pins the declared set by name.
func TestMCPSchemaLint_HandlerReadArgumentsDeclared(t *testing.T) {
	tools := lintServedTools(t)
	declared := map[string]map[string]bool{}
	for _, tool := range tools {
		set := map[string]bool{}
		for pname := range tool.InputSchema.Properties {
			set[pname] = true
		}
		declared[tool.Name] = set
	}
	for _, tc := range []struct {
		tool, arg string
	}{
		// codex_task left with the general codex execution surface (card t1909);
		// the pin keeps the remaining named instance.
		{"goal_status", "session_id"},
	} {
		if !declared[tc.tool][tc.arg] {
			t.Errorf("%s does not declare %q, but the handler reads it", tc.tool, tc.arg)
		}
	}
}
