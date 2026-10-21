// mcp_root_contract_test.go — SPEC-MCP-NEUTRALITY-001 M2 (card t1632): the
// common project_root contract (REQ-MN-004/005). The resolver confines a
// caller-supplied root to a registered worktree of the server's own
// repository, rejects the home directory explicitly, rejects a non-string
// value as an invalid argument instead of coercing it, and exposes the
// provenance block a root-dependent result carries.
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
)

// sdToolReq builds a CallToolRequest with the given arguments.
func sdToolReq(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}}
}

// sdServerFixture is a fixture whose root IS the server root: CLAUDE_PROJECT_DIR
// names it, so confinement compares the caller's root against this repository.
func sdServerFixture(t *testing.T) string {
	t.Helper()
	root, _ := sdMoaiFixture(t)
	t.Setenv(config.EnvClaudeProjectDir, root)
	return root
}

// TestMCPRootConfinement pins REQ-MN-005: the home directory and a sibling
// repository are rejected, a registered worktree of the server's own
// repository is accepted, and a non-string project_root is an invalid
// argument — never coerced to the default.
func TestMCPRootConfinement(t *testing.T) {
	serverRoot := sdServerFixture(t)

	t.Run("the home directory is rejected explicitly", func(t *testing.T) {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skip("no home directory to test against")
		}
		_, err = resolveToolProjectRoot(sdToolReq(map[string]any{"project_root": home}))
		if err == nil {
			t.Fatal("the home directory was accepted as a project root")
		}
		if !strings.Contains(err.Error(), "home") {
			t.Errorf("home rejection = %v, want it to name the home directory", err)
		}
	})

	t.Run("a sibling repository is rejected", func(t *testing.T) {
		sibling, _ := sdMoaiFixture(t) // a different temp-dir repository
		if sibling == serverRoot {
			t.Fatal("fixture collision")
		}
		// Creating the fixture re-stamped CLAUDE_PROJECT_DIR to itself
		// (todoFixture); the server root for this arm is the parent's.
		t.Setenv(config.EnvClaudeProjectDir, serverRoot)
		_, err := resolveToolProjectRoot(sdToolReq(map[string]any{"project_root": sibling}))
		if err == nil {
			t.Fatalf("sibling repository %s was accepted", sibling)
		}
	})

	t.Run("a registered worktree of the server repository is accepted", func(t *testing.T) {
		wt := filepath.Join(filepath.Dir(serverRoot), filepath.Base(serverRoot)+"-wt")
		fcGit(t, serverRoot, "worktree", "add", "-q", wt, "-b", "WT-confinement")
		got, err := resolveToolProjectRoot(sdToolReq(map[string]any{"project_root": wt}))
		if err != nil {
			t.Fatalf("registered worktree %s rejected: %v", wt, err)
		}
		// The accepted path is canonicalized — symlinks resolved — so the
		// comparison runs on the resolved spelling.
		if want, linkErr := filepath.EvalSymlinks(wt); linkErr != nil || got != want {
			t.Errorf("root = %q, want the canonical %q", got, want)
		}
	})

	t.Run("a non-string project_root is an invalid argument, never coerced", func(t *testing.T) {
		_, source, err := resolveToolProjectRootWithSource(sdToolReq(map[string]any{"project_root": 42}))
		if err == nil {
			t.Fatal("project_root 42 was accepted")
		}
		if source != "" {
			t.Errorf("source = %q, want empty (no resolution happened)", source)
		}
		if !strings.Contains(err.Error(), "invalid") {
			t.Errorf("non-string rejection = %v, want an invalid-argument error", err)
		}
	})

	t.Run("a caller-supplied root carries the param source without warning", func(t *testing.T) {
		want, _ := filepath.EvalSymlinks(serverRoot)
		root, source, err := resolveToolProjectRootWithSource(sdToolReq(map[string]any{"project_root": serverRoot}))
		if err != nil || root != want {
			t.Fatalf("server root rejected: root=%q err=%v", root, err)
		}
		if source != "param" {
			t.Errorf("source = %q, want param", source)
		}
		prov := rootProvenanceMap(root, source)
		if _, has := prov["warning"]; has {
			t.Errorf("param resolution carried a warning: %v", prov)
		}
	})
}

// TestMCPRootRequiredForMutating pins AC-MN-006's wire shape: a mutating
// family tool called WITHOUT project_root refuses at the tool layer with
// isError true and the pinned text — argument validation, deliberately
// outside the five policy codes.
func TestMCPRootRequiredForMutating(t *testing.T) {
	res := mcpCallRaw(t, handleFactoryStage, map[string]any{
		"card": "t1", "state": "planning",
	})
	if !res.IsError {
		t.Fatalf("mutating family tool without project_root was accepted: %s", mcpTextOf(res))
	}
	want := "factory_stage: project_root is required — pass the tree to act on (your git rev-parse --show-toplevel)"
	if got := mcpTextOf(res); got != want {
		t.Errorf("wire text = %q, want %q", got, want)
	}
}
