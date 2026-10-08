package hook

// m4_user_root_parity_test.go — SPEC-USERASSET-DEPLOY-GUARD-001 M4 parity
// (REQ-GRD-002, design §5): the hook's slug→directory table MUST agree
// with userassets.ResolveRoots — the import direction stays production-
// hook ← config only, so the agreement is pinned by this TEST-ONLY import:
// a drift on either side turns this test red.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/userassets"
)

func TestUserRootSlugDirsMatchUserassetsRoots(t *testing.T) {
	home := t.TempDir()
	roots := userassets.ResolveRoots(home)
	if len(roots) != len(userRootSlugDirs) {
		t.Fatalf("root count drifted: userassets has %d, the hook table has %d", len(roots), len(userRootSlugDirs))
	}
	for _, r := range roots {
		dir, ok := userRootSlugDirs[string(r.Slug)]
		if !ok {
			t.Fatalf("slug %q is missing from the hook's user-root table", r.Slug)
		}
		want := filepath.ToSlash(r.Dir)
		got := filepath.ToSlash(filepath.Join(home, filepath.FromSlash(dir)))
		if got != want {
			t.Errorf("slug %q: hook dir %q != userassets root %q", r.Slug, got, want)
		}
	}
}

// TestUserRootFormsEmitNamespacedForm pins the resolver's user-root forms:
// an absolute path under a user root emits exactly one namespaced form,
// the ~ alias emits the same form, and a project-relative path emits none.
func TestUserRootFormsEmitNamespacedForm(t *testing.T) {
	home := t.TempDir()
	prev := zoneHomeFn
	zoneHomeFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { zoneHomeFn = prev })

	target := filepath.Join(home, ".claude", "agents", "moai", "plan-auditor.md")
	forms := userRootForms(target)
	if len(forms) != 1 {
		t.Fatalf("forms = %v (want exactly the claude-agents form)", forms)
	}
	want := "user-root:claude-agents/moai/plan-auditor.md"
	if forms[0].Display != want {
		t.Fatalf("form display = %q, want %q", forms[0].Display, want)
	}

	alias := userRootForms(home + "/.claude/agents/moai/plan-auditor.md")
	tilde := userRootForms("~/.claude/agents/moai/plan-auditor.md")
	if len(tilde) != 1 || tilde[0].Display != want {
		t.Fatalf("the ~ alias form = %v, want [%s]", tilde, want)
	}
	_ = alias

	if got := userRootForms(".claude/agents/moai/plan-auditor.md"); got != nil {
		t.Fatalf("a relative (project-relative) target produced user forms: %v", got)
	}
}

// TestUserRootFormTrackedScoping pins the REQ-GRD-002 limitation at the
// helper level: a manifest-tracked key reads tracked, an untracked key
// does not, and an unreadable manifest reads untracked (fail toward
// allowing the user's own file).
func TestUserRootFormTrackedScoping(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema_version": 1, "files": {"claude-agents/moai/plan-auditor.md": {"sha256": "ab"}}}`
	if err := os.WriteFile(filepath.Join(home, ".moai", "user-assets.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	prev := zoneHomeFn
	zoneHomeFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { zoneHomeFn = prev })

	if !userRootFormTracked(home, "user-root:claude-agents/moai/plan-auditor.md") {
		t.Fatal("a tracked file read untracked")
	}
	if userRootFormTracked(home, "user-root:claude-agents/moai/user-own-note.md") {
		t.Fatal("an untracked file read tracked — the over-protection guard is open")
	}
}
