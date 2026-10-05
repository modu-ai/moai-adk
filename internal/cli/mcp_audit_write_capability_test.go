package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// mcp_audit_write_capability_test.go — card t904.
//
// TestMoaiMCPServer_AnnotationsMatchCatalog compares a tool's DECLARED
// annotation against the catalog's WriteCapable flag. Both sides are hand-
// maintained claims, so that guard holds them consistent with each other and
// says nothing about whether either matches what the handler actually does.
// That is the gap this file closes for the audit family: codex_audit and
// audit_multi read as catalog-READ tools for as long as it took the audit
// receipt store to land (SPEC-CODEX-AUDIT-GATE-AXES-001) without the catalog
// following, and no signal existed because a hint that under-reports a write
// simply suppresses an approval prompt.
//
// The assertion direction matters: this test fails if the writes STOP, which
// is when the catalog entry would need to change back. It is the behavioral
// half of the pair whose declaration half is
// internal/mcp TestMoaiMCPTools_ElevenWriteCapable.

// newAuditToolTestRoot returns a canonical temp directory that passes
// validateProjectRoot (it requires a .moai directory). Canonical because macOS
// hands out a symlinked temp path while the resolver returns the resolved one,
// and the test compares paths.
func newAuditToolTestRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	return root
}

// stubAuditBackendsAbsent makes every audit backend report itself unavailable,
// so the call takes its fail-open path: no codex process, no claude process,
// no network. The writes under test happen on that path too — which is the
// point, since a tool that writes even when the review never ran is a tool
// whose read-only hint cannot be conditional on the review succeeding.
func stubAuditBackendsAbsent(t *testing.T) {
	t.Helper()
	codexOrig, claudeOrig, glmOrig := codexLookPath, claudeLookPath, glmKeyLoader
	codexLookPath = func(string) (string, error) { return "", errors.New("t904 test: codex absent") }
	claudeLookPath = func(string) (string, error) { return "", errors.New("t904 test: claude absent") }
	glmKeyLoader = func() string { return "" }
	t.Cleanup(func() {
		codexLookPath, claudeLookPath, glmKeyLoader = codexOrig, claudeOrig, glmOrig
	})
}

// callAuditTool runs one tool over an in-process MCP client and fails the test
// on a transport error or a tool-level error result (a rejected call writes
// nothing, which would otherwise read as "no write observed").
func callAuditTool(t *testing.T, name string, args map[string]any) {
	t.Helper()
	srv := newMoaiMCPServer()
	if srv == nil {
		t.Fatal("newMoaiMCPServer returned nil")
	}
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)
	ctx := context.Background()
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args
	res, err := c.CallTool(ctx, req)
	if err != nil {
		t.Fatalf("%s: CallTool: %v", name, err)
	}
	if res.IsError {
		t.Fatalf("%s: tool returned an error result; a rejected call writes nothing, so the write assertion below would be vacuous: %+v", name, res.Content)
	}
}

// filesUnder lists every regular file below dir, empty when dir is absent.
func filesUnder(dir string) []string {
	var out []string
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err == nil && d != nil && !d.IsDir() {
			out = append(out, p)
		}
		return nil
	})
	return out
}

// TestMCPAuditTools_DeclaredWriteCapableActuallyWrite is the behavioral pin:
// both audit tools the catalog marks WriteCapable must be observed writing
// into the tree the caller named.
func TestMCPAuditTools_DeclaredWriteCapableActuallyWrite(t *testing.T) {
	cases := []struct {
		tool string
		args func(root string) map[string]any
		// wantRel is a directory under the named root that must hold at least
		// one file after the call.
		wantRel string
	}{
		{
			tool: "codex_audit",
			args: func(root string) map[string]any {
				return map[string]any{"project_root": root, "mode": "native", "target": "uncommittedChanges"}
			},
			wantRel: filepath.Join(".moai", "state", "audit-receipts"),
		},
		{
			tool: "audit_multi",
			args: func(root string) map[string]any {
				return map[string]any{"project_root": root, "session_id": "t904-write-capability"}
			},
			wantRel: filepath.Join(".moai", "state", "audit-multi"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.tool, func(t *testing.T) {
			root := newAuditToolTestRoot(t)
			stubAuditBackendsAbsent(t)

			target := filepath.Join(root, tc.wantRel)
			if before := filesUnder(target); len(before) != 0 {
				t.Fatalf("precondition: %s already holds %v", target, before)
			}

			callAuditTool(t, tc.tool, tc.args(root))

			after := filesUnder(target)
			if len(after) == 0 {
				t.Fatalf("%s wrote nothing under %s — if the write was removed deliberately, flip its catalog WriteCapable back to false and its ReadOnlyHint annotation back to true in the same change", tc.tool, target)
			}
			t.Logf("%s wrote %d file(s) under %s", tc.tool, len(after), tc.wantRel)
		})
	}
}
