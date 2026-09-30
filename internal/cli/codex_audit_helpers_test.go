// codex_audit_helpers_test.go — platform-neutral test helpers shared by the
// Codex audit and codex_task test files. These declarations were formerly
// defined in `//go:build !windows`-tagged files while untagged test files
// (codex_task_test.go) reference them, which broke the Windows compile:
// tagged files vanish on Windows, so newAuditRepo, auditGit, and
// roleAuditResultText became undefined under `GOOS=windows go vet`.
// The helpers are portable — auditGit shells out to git, newAuditRepo builds
// a repo layout through auditGit and the embedded FS, roleAuditResultText is
// pure Go over mcp.CallToolResult — so they live in this untagged file and
// compile on every platform (card t1353).
package cli

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/template"
)

// auditGit runs git in dir with a neutral identity and fails the test on error.
func auditGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@example.invalid", "-c", "commit.gpgsign=false"}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// auditTempDir returns a symlink-free temp dir (macOS /var → /private/var).
func auditTempDir(t *testing.T) string {
	t.Helper()
	d, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// auditRepo is a primary checkout A with two registered worktrees A1 and A2.
type auditRepo struct {
	base, a, a1, a2 string
}

func newAuditRepo(t *testing.T) auditRepo {
	t.Helper()
	base := auditTempDir(t)
	a := filepath.Join(base, "A")
	if err := os.MkdirAll(a, 0o755); err != nil {
		t.Fatal(err)
	}
	auditGit(t, a, "init", "-q", "-b", "main")
	auditGit(t, a, "commit", "-q", "--allow-empty", "-m", "init")
	a1 := filepath.Join(base, "A1")
	a2 := filepath.Join(base, "A2")
	auditGit(t, a, "worktree", "add", "-q", "-b", "wt-one", a1)
	auditGit(t, a, "worktree", "add", "-q", "-b", "wt-two", a2)
	for _, root := range []string{a, a1, a2} {
		installAuditRoles(t, root)
	}
	return auditRepo{base: base, a: a, a1: a1, a2: a2}
}

// installAuditRoles copies the emitted Codex role files into root.
func installAuditRoles(t *testing.T, root string) {
	t.Helper()
	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, ".codex", "agents", "moai")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	entries, err := fs.ReadDir(fsys, ".codex/agents/moai")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		b, err := fs.ReadFile(fsys, ".codex/agents/moai/"+e.Name())
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dst, e.Name()), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// roleAuditResultText concatenates the text content of a tool result.
func roleAuditResultText(res *mcp.CallToolResult) string {
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}
