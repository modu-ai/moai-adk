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
// Gate round 26 #1: a /./ spelling resolves to the same form as the plain
// path — the judgment runs on the resolved target, not the raw input.
func TestUserRootFormsEmitNamespacedForm(t *testing.T) {
	home := t.TempDir()
	prev := zoneHomeFn
	zoneHomeFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { zoneHomeFn = prev })

	target := filepath.Join(home, ".claude", "agents", "moai", "plan-auditor.md")
	forms := userRootZoneForms(target)
	if len(forms) != 1 {
		t.Fatalf("forms = %v (want exactly the claude-agents form)", forms)
	}
	want := "user-root:claude-agents/moai/plan-auditor.md"
	if forms[0].Display != want {
		t.Fatalf("form display = %q, want %q", forms[0].Display, want)
	}

	tilde := userRootZoneForms("~/.claude/agents/moai/plan-auditor.md")
	if len(tilde) != 1 || tilde[0].Display != want {
		t.Fatalf("the ~ alias form = %v, want [%s]", tilde, want)
	}

	// Gate round 26 #1: the dot-path and parent-path spellings resolve to
	// the SAME namespaced form — the match judges the resolved target.
	// (Raw concatenation, NOT filepath.Join: Join would clean the spellings
	// away before the resolver ever sees them.)
	dotted := userRootZoneForms(home + "/./.claude/agents/moai/plan-auditor.md")
	if len(dotted) != 1 || dotted[0].Display != want {
		t.Fatalf("the /./ spelling form = %v, want [%s]", dotted, want)
	}
	parent := userRootZoneForms(home + "/.claude/agents/moai/../moai/plan-auditor.md")
	if len(parent) != 1 || parent[0].Display != want {
		t.Fatalf("the /../ spelling form = %v, want [%s]", parent, want)
	}

	// Gate round 34-1: the ROOT ITSELF and an ANCESTOR of the root emit the
	// bare-root form (rm -rf ~/.claude/skills / rm -rf ~/.claude delete
	// every tracked file inside them).
	rootForms := userRootZoneForms(home + "/.claude/agents")
	rootHit := false
	for _, f := range rootForms {
		if f.Display == "user-root:claude-agents" {
			rootHit = true
		}
	}
	if !rootHit {
		t.Fatalf("the root-itself target emitted %v — the bare-root protected form is missing", rootForms)
	}
	ancestorForms := userRootZoneForms(home + "/.claude")
	agentsHit := false
	for _, f := range ancestorForms {
		if f.Display == "user-root:claude-agents" {
			agentsHit = true
		}
	}
	if !agentsHit {
		t.Fatalf("the ancestor target emitted %v — no bare-root form for the contained root", ancestorForms)
	}

	// Gate round 34-2: a CASE-ALIAS spelling of the root emits the same
	// folded form (the containment must not be case-sensitive ahead of the
	// folded manifest-key comparison).
	caseAlias := userRootZoneForms(home + "/.CLAUDE/agents/moai/plan-auditor.md")
	caseHit := false
	for _, f := range caseAlias {
		if f.Display == want {
			caseHit = true
		}
	}
	if !caseHit {
		t.Fatalf("the case-alias spelling emitted %v — the case-alias bypass is open", caseAlias)
	}

	if got := userRootZoneForms(".claude/agents/moai/plan-auditor.md"); got != nil {
		t.Fatalf("a relative (project-relative) target produced user forms: %v", got)
	}
}

// TestUserRootFormTrackedScoping pins the REQ-GRD-002 limitation at the
// helper level: a manifest-tracked key reads tracked, an untracked key
// does not, and an unreadable manifest reads untracked (fail toward
// allowing the user's own file). Gate round 26 #2: a DIRECTORY key is
// covered when any tracked file lives under it (prefix containment).
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

	if !userRootTracksAny(home, "user-root:claude-agents/moai/plan-auditor.md") {
		t.Fatal("a tracked file read untracked")
	}
	if userRootTracksAny(home, "user-root:claude-agents/moai/user-own-note.md") {
		t.Fatal("an untracked file read tracked — the over-protection guard is open")
	}
	// Gate round 26 #2: the directory containment.
	if !userRootTracksAny(home, "user-root:claude-agents/moai") {
		t.Fatal("a directory containing a tracked file read uncovered — the prefix containment is missing")
	}
	if !userRootTracksAny(home, "user-root:claude-agents") {
		t.Fatal("the root directory containing a tracked file read uncovered")
	}
	if userRootTracksAny(home, "user-root:claude-agents/other") {
		t.Fatal("a directory with no tracked content read covered")
	}
}
