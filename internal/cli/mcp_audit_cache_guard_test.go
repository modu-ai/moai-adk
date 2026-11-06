package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestAuditCacheStoreRequiresProjectRoot(t *testing.T) {
	root := sdServerFixture(t)
	report := filepath.Join(root, "verdict.md")
	if err := os.WriteFile(report, []byte("# verdict\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := handleAuditCache(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "audit_cache",
			Arguments: map[string]any{
				"op": "store", "spec_id": "SPEC-ROOT-REQUIRED-001", "hash": "root-required",
				"auditor_version": "plan-auditor", "report_path": report,
			},
		},
	})
	if err != nil {
		t.Fatalf("audit_cache store: %v", err)
	}
	if res == nil || !res.IsError {
		t.Fatalf("audit_cache store without project_root was accepted: %s", mcpTextOf(res))
	}
	if got, want := mcpTextOf(res), "audit_cache: project_root is required for mutating calls"; got != want {
		t.Fatalf("wire text = %q, want %q", got, want)
	}
}

func TestStaleAuditCacheStoreRefusesWithMetadata(t *testing.T) {
	root := sdServerFixture(t)
	forceStaleMCPServer(t)
	report := filepath.Join(root, "verdict.md")
	if err := os.WriteFile(report, []byte("# verdict\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	guarded := mcpStalenessGuard("audit_cache", handleAuditCache)
	res, err := guarded(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "audit_cache",
			Arguments: map[string]any{
				"op": "store", "project_root": root, "spec_id": "SPEC-STALE-STORE-001",
				"hash": "stale-store", "auditor_version": "plan-auditor", "report_path": report,
			},
		},
	})
	if err != nil {
		t.Fatalf("stale audit_cache store: %v", err)
	}
	if res == nil || !res.IsError {
		t.Fatalf("stale audit_cache store was accepted: %s", mcpTextOf(res))
	}
	assertStaleMetadataInBothChannels(t, res)
}

func TestStaleReadResultMetadataIsMirroredToTextJSON(t *testing.T) {
	forceStaleMCPServer(t)
	guarded := mcpStalenessGuard("session_list", func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		res := toolJSON("session_list", map[string]any{"sessions": []string{"s1"}})
		res.Content = append(res.Content, mcp.TextContent{Type: "text", Text: "additional result block"})
		return res, nil
	})
	res, err := guarded(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Name: "session_list", Arguments: map[string]any{}},
	})
	if err != nil {
		t.Fatalf("stale session_list: %v", err)
	}
	if res == nil || res.IsError {
		t.Fatalf("stale read result unexpectedly errored: %s", mcpTextOf(res))
	}
	assertStaleMetadataInBothChannels(t, res)
	if !mcpHasText(res, "additional result block") {
		t.Fatal("stale metadata update discarded an additional result block")
	}
}

func mcpHasText(res *mcp.CallToolResult, want string) bool {
	for _, item := range res.Content {
		if block, ok := item.(mcp.TextContent); ok && block.Text == want {
			return true
		}
	}
	return false
}

func forceStaleMCPServer(t *testing.T) {
	t.Helper()
	installed := filepath.Join(t.TempDir(), "moai")
	startedAt := time.Now().Add(-time.Hour)
	writeFileOfSize(t, installed, 4096, startedAt.Add(time.Minute))
	withStalenessSeams(t, binaryIdentity{
		Path: installed, Size: 2048, ModTime: startedAt, Captured: true,
	}, installed)
}

func assertStaleMetadataInBothChannels(t *testing.T, res *mcp.CallToolResult) {
	t.Helper()
	structured, ok := res.StructuredContent.(map[string]any)
	if !ok || structured["stale_server"] != true {
		t.Fatalf("structured result missing stale_server: %#v", res.StructuredContent)
	}
	if _, ok := structured["_server"].(map[string]any); !ok {
		t.Fatalf("structured result missing _server identity: %#v", structured["_server"])
	}

	var text string
	for _, item := range res.Content {
		if block, ok := item.(mcp.TextContent); ok {
			text = block.Text
			break
		}
	}
	if text == "" {
		t.Fatal("stale result has no text payload")
	}
	var textPayload map[string]any
	if err := json.Unmarshal([]byte(text), &textPayload); err != nil {
		t.Fatalf("stale text payload is not JSON: %q: %v", text, err)
	}
	if textPayload["stale_server"] != true {
		t.Fatalf("text JSON missing stale_server: %v", textPayload)
	}
	if _, ok := textPayload["_server"].(map[string]any); !ok {
		t.Fatalf("text JSON missing _server identity: %v", textPayload)
	}
}
