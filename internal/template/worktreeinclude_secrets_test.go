package template

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
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
// and credential documents. The first eighteen are the Rust product's `SECRET_PROBES`
// (`src/worktree_include.rs` on v2026.11.0, audit 6857107471 N1); the last is the
// per-machine Claude Code settings file, which holds ANTHROPIC_AUTH_TOKEN in GLM mode
// (user decision on #2670, audit S1).
var worktreeIncludeSecretProbes = []string{
	".env",
	".env.local",
	".env.production",
	".env.development",
	"apps/web/.env",
	"apps/web/.env.local",
	"id_rsa",
	"certs/server.key",
	"certs/server.pem",
	"certs/client.p12",
	"config/master.key",
	"credentials.json",
	"service-account.json",
	"config/secrets.yml",
	"secrets/token",
	".aws/credentials",
	".npmrc",
	".netrc",
	".claude/settings.local.json",
}

// worktreeIncludeSecretLines returns the pattern lines of a .worktreeinclude body that
// select a secret probe. It reads .gitignore syntax closely enough for the template:
// comments and blank lines are skipped, a pattern with no inner slash (after a leading
// `**/`) matches any path component, a pattern with one is anchored at the root, and a
// trailing slash names a folder. Negations (`!`) are skipped: they can only un-select.
// Names compare without case, as the Rust check does (git on the default macOS file
// system).
func worktreeIncludeSecretLines(body string) []string {
	var hits []string
	for _, raw := range strings.Split(body, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		pattern := strings.ToLower(strings.TrimSuffix(line, "/"))
		// A leading `**/` matches in every folder: the same as an unanchored name.
		pattern = strings.TrimPrefix(pattern, "**/")
		anchored := strings.Contains(strings.TrimPrefix(pattern, "/"), "/")
		pattern = strings.TrimPrefix(pattern, "/")
		for _, probe := range worktreeIncludeSecretProbes {
			if worktreeIncludeSelects(pattern, anchored, strings.ToLower(probe)) {
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
}

// The repository's own copy at the root (user decision on #2670, audit 6857107471 N2):
// the worktrees of this repository must not get secrets either, and the copy stays the
// same as the template so the two cannot drift apart again.
func TestRepositoryWorktreeIncludeCopiesNoSecrets(t *testing.T) {
	root, err := os.ReadFile(filepath.Join("..", "..", ".worktreeinclude"))
	if err != nil {
		t.Fatalf("the repository .worktreeinclude is unreadable: %v", err)
	}
	if hits := worktreeIncludeSecretLines(string(root)); len(hits) > 0 {
		t.Fatalf("the repository .worktreeinclude selects secret files %q (Issue #2670)", hits)
	}
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	tmpl, err := fs.ReadFile(fsys, ".worktreeinclude")
	if err != nil {
		t.Fatalf("the template .worktreeinclude is not embedded: %v", err)
	}
	if string(root) != string(tmpl) {
		t.Fatalf("the repository .worktreeinclude differs from the template; keep them identical")
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
	// The removed per-machine settings line (audit S1, user decision on #2670).
	if got := worktreeIncludeSecretLines("# Per-machine settings\n.claude/settings.local.json\n"); len(got) != 1 {
		t.Fatalf("checker missed the removed settings line: %q", got)
	}
	for _, line := range []string{
		"*.pem", "certs/", "**/.env", "/.env", "credentials.json",
		// Audit 6857107471 N1: lines the eight-probe checker let through.
		"apps/web/.env.local", ".aws/credentials", "config/secrets.yml", "*.p12",
		// The rest of the Rust probe list, and case.
		".env.development", "config/master.key", "service-account.json", "secrets/",
		".npmrc", ".netrc", ".claude/", ".ENV",
	} {
		if len(worktreeIncludeSecretLines(line+"\n")) != 1 {
			t.Errorf("checker missed %q", line)
		}
	}
	for _, line := range []string{"# .env", "!.env", ".envrc", ".env.example", ".claude/settings.json", "node_modules/"} {
		if hits := worktreeIncludeSecretLines(line + "\n"); len(hits) != 0 {
			t.Errorf("checker flagged a safe line %q", line)
		}
	}
}

// worktreeIncludeRuleExample returns the first ```text block under the `.worktreeinclude`
// heading of the worktree rules document: the example an agent copies into a project.
func worktreeIncludeRuleExample(t *testing.T, doc string) string {
	t.Helper()
	_, section, ok := strings.Cut(doc, "### `.worktreeinclude`")
	if !ok {
		t.Fatalf("the rules document has no `.worktreeinclude` section")
	}
	_, block, ok := strings.Cut(section, "```text\n")
	if !ok {
		t.Fatalf("the `.worktreeinclude` section has no example block")
	}
	block, _, ok = strings.Cut(block, "```")
	if !ok {
		t.Fatalf("the `.worktreeinclude` example block is not closed")
	}
	return block
}

// Audit 6858803521 B1: the deployed worktree rules recommended `.env` and `.env.local` as
// .worktreeinclude examples, so an agent following them would put the removed lines back.
// Both the shipped copy and this repository's own copy are checked.
func TestWorktreeIncludeRuleExampleCopiesNoSecrets(t *testing.T) {
	const rule = ".claude/rules/moai/workflow/worktree-integration-ops.md"
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	shipped, err := fs.ReadFile(fsys, rule)
	if err != nil {
		t.Fatalf("the shipped rules document is not embedded: %v", err)
	}
	own, err := os.ReadFile("../../" + rule)
	if err != nil {
		t.Fatalf("read this repository's rules document: %v", err)
	}
	for name, doc := range map[string]string{"shipped": string(shipped), "repository": string(own)} {
		if hits := worktreeIncludeSecretLines(worktreeIncludeRuleExample(t, doc)); len(hits) > 0 {
			t.Errorf("the %s rules document recommends secret .worktreeinclude lines %q "+
				"(Issue #2670)", name, hits)
		}
	}
}
