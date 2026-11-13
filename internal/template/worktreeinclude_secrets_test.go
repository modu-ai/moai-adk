package template

import (
	"io/fs"
	"path"
	"strings"
	"testing"
)

// Issue #2670: Claude Code copies every gitignored file a project's .worktreeinclude
// selects into each new worktree, including the worktree of every subagent spawned with
// `isolation: worktree`. The shipped template listed `.env` and `.env.*`, so each worktree
// got a copy of the project's secrets. These tests keep secret-selecting lines out of the
// template that `moai init` / `moai update` deploy.

// worktreeIncludeSecretProbes stand for files a project keeps out of git because they hold
// credentials: environment files at the root and nested, private keys and certificates,
// and credential documents.
var worktreeIncludeSecretProbes = []string{
	".env",
	".env.local",
	".env.production",
	"apps/web/.env",
	"id_rsa",
	"certs/server.key",
	"certs/server.pem",
	"credentials.json",
}

// worktreeIncludeSecretLines returns the pattern lines of a .worktreeinclude body that
// select a secret probe. It reads .gitignore syntax closely enough for the template:
// comments and blank lines are skipped, a pattern with no inner slash (after a leading
// `**/`) matches any path component, a pattern with one is anchored at the root, and a
// trailing slash names a folder. Negations (`!`) are skipped: they can only un-select.
func worktreeIncludeSecretLines(body string) []string {
	var hits []string
	for _, raw := range strings.Split(body, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		pattern := strings.TrimSuffix(line, "/")
		// A leading `**/` matches in every folder: the same as an unanchored name.
		pattern = strings.TrimPrefix(pattern, "**/")
		anchored := strings.Contains(strings.TrimPrefix(pattern, "/"), "/")
		pattern = strings.TrimPrefix(pattern, "/")
		for _, probe := range worktreeIncludeSecretProbes {
			if worktreeIncludeSelects(pattern, anchored, probe) {
				hits = append(hits, line)
				break
			}
		}
	}
	return hits
}

func worktreeIncludeSelects(pattern string, anchored bool, probe string) bool {
	parts := strings.Split(probe, "/")
	for i := range parts {
		// A pattern that names a folder above the file selects the file too.
		prefix := strings.Join(parts[:i+1], "/")
		subject := parts[i]
		if anchored {
			subject = prefix
		}
		if ok, _ := path.Match(pattern, subject); ok {
			return true
		}
	}
	return false
}

func TestWorktreeIncludeTemplateCopiesNoSecrets(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	data, err := fs.ReadFile(fsys, ".worktreeinclude")
	if err != nil {
		t.Fatalf("the template .worktreeinclude is not embedded: %v", err)
	}
	if hits := worktreeIncludeSecretLines(string(data)); len(hits) > 0 {
		t.Fatalf("the shipped .worktreeinclude selects secret files %q; every new worktree "+
			"would get a copy of them (Issue #2670)", hits)
	}
	if !strings.Contains(string(data), ".claude/settings.local.json") {
		t.Fatalf("the per-machine settings line is gone from the template; only the secret " +
			"lines were meant to be removed")
	}
}

// The checker must flag the lines the template used to ship, or the test above proves
// nothing.
func TestWorktreeIncludeSecretCheckerFlagsTheRemovedLines(t *testing.T) {
	old := "# Environment files\n.env\n.env.*\n"
	got := worktreeIncludeSecretLines(old)
	if len(got) != 2 || got[0] != ".env" || got[1] != ".env.*" {
		t.Fatalf("checker missed the removed template lines: %q", got)
	}
	for _, line := range []string{"*.pem", "certs/", "**/.env", "/.env", "credentials.json"} {
		if len(worktreeIncludeSecretLines(line+"\n")) != 1 {
			t.Errorf("checker missed %q", line)
		}
	}
	for _, line := range []string{"# .env", "!.env", ".envrc", ".claude/settings.local.json", "node_modules/"} {
		if hits := worktreeIncludeSecretLines(line + "\n"); len(hits) != 0 {
			t.Errorf("checker flagged a safe line %q", line)
		}
	}
}
