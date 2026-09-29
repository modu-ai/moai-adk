package web

// SPEC-WEB-SAVE-LOSSLESS-001 — web-layer lossless contract tests through the
// full POST /save path (handleSave). RED-first against the pre-implementation
// tree: typed scalar writes and the nested seam rode SetSection → Save, a full
// re-marshal that dropped the GitHub #1731 issue-named unmodeled keys and
// rewrote unrelated section files.
//
// Isolation: profile.BaseDirOverride keeps WritePreferences out of $HOME (same
// pattern as integration_test.go setupRealProject); every fixture lives under
// t.TempDir().

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/settings"
)

// losslessFixture is the seed the AC-WSL-009 tests share: the issue-named
// unmodeled key sits inside quality.yaml, and git-convention.yaml carries a
// convention to edit.
const losslessQuality = `constitution:
  development_mode: tdd
  session_effort_default: xhigh  # local note
  # hand-maintained comment
`

const losslessGitConvention = `git_convention:
  # convention comment
  convention: auto
`

const losslessUser = `user:
  name: original
  github_username: example-user
`

const losslessLLM = `llm:
  mode: ""
  # llm hand comment
  glm:
    models:
      high: glm-5.2
`

func seedLosslessWebProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"quality.yaml":        losslessQuality,
		"git-convention.yaml": losslessGitConvention,
		"user.yaml":           losslessUser,
		"llm.yaml":            losslessLLM,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func readWebSection(t *testing.T, root, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", name+".yaml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// changedLines is the web-local copy of the one-line-diff predicate (internal
// settings helpers are not exported to the web package).
func changedLines(t *testing.T, before, after string) []string {
	t.Helper()
	b := strings.Split(before, "\n")
	a := strings.Split(after, "\n")
	if len(a) != len(b) {
		t.Fatalf("line count changed: before=%d after=%d\n--- before ---\n%s\n--- after ---\n%s", len(b), len(a), before, after)
	}
	var out []string
	for i := range b {
		if a[i] != b[i] {
			out = append(out, a[i])
		}
	}
	return out
}

// newLosslessApp builds a Console app over the seeded project with the profile
// store isolated in a temp dir.
func newLosslessApp(t *testing.T, root string) *app {
	t.Helper()
	orig := profile.BaseDirOverride
	profile.BaseDirOverride = t.TempDir()
	t.Cleanup(func() { profile.BaseDirOverride = orig })

	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }
	return a
}

// differingModelOption returns an llm.glm.models.high option value that
// differs from the fixture's persisted value, so the submission is a real
// change (a no-op submission would make the guard vacuous).
func differingModelOption(t *testing.T, persisted string) string {
	t.Helper()
	f, ok := settings.Field("llm.glm.models.high")
	if !ok {
		t.Fatal("llm.glm.models.high not registered")
	}
	for _, o := range f.Options {
		if o.Value != persisted {
			return o.Value
		}
	}
	t.Fatal("no differing model option available")
	return ""
}

// TestHandleSaveDevModeEditPreservesIssueNamedKey carries AC-WSL-009 through
// the full save path: editing development_mode must change exactly the
// development_mode line of quality.yaml and leave the issue-named unmodeled
// key `session_effort_default` (with its comment) byte-intact. RED on the
// pre-implementation tree (SetSection("quality") → Save full re-marshal).
func TestHandleSaveDevModeEditPreservesIssueNamedKey(t *testing.T) {
	root := seedLosslessWebProject(t)
	a := newLosslessApp(t, root)

	form := url.Values{
		"__profile":        {"default"},
		"development_mode": {"ddd"}, // fixture tdd → real change
	}
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("devMode edit status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}

	after := readWebSection(t, root, "quality")
	changed := changedLines(t, losslessQuality, after)
	if len(changed) != 1 || !strings.Contains(changed[0], "development_mode: ddd") {
		t.Fatalf("devMode edit changed %d lines (%v), want exactly the development_mode line\n--- after ---\n%s", len(changed), changed, after)
	}
	if !strings.Contains(after, "session_effort_default: xhigh  # local note") {
		t.Errorf("issue-named unmodeled key lost by a devMode edit:\n%s", after)
	}
	if !strings.Contains(after, "# hand-maintained comment") {
		t.Errorf("comment lost by a devMode edit:\n%s", after)
	}
}

// TestHandleSaveConventionEditLeavesQualityByteIdentical carries the
// AC-WSL-009 convention variant: a git_convention edit changes one
// git-convention.yaml line and leaves quality.yaml (and the issue-named key
// inside it) byte-identical.
func TestHandleSaveConventionEditLeavesQualityByteIdentical(t *testing.T) {
	root := seedLosslessWebProject(t)
	a := newLosslessApp(t, root)

	form := url.Values{
		"__profile":      {"default"},
		"git_convention": {"conventional-commits"}, // fixture auto → real change
	}
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("convention edit status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}

	afterGC := readWebSection(t, root, "git-convention")
	changed := changedLines(t, losslessGitConvention, afterGC)
	if len(changed) != 1 || !strings.Contains(changed[0], "convention: conventional-commits") {
		t.Fatalf("convention edit changed %d lines (%v), want exactly the convention line\n--- after ---\n%s", len(changed), changed, afterGC)
	}
	if afterQ := readWebSection(t, root, "quality"); afterQ != losslessQuality {
		t.Errorf("convention edit rewrote quality.yaml (must be byte-identical)\n--- before ---\n%s\n--- after ---\n%s", losslessQuality, afterQ)
	}
}

// TestHandleSaveUserNameEditSplicesOneRow carries AC-WSL-005 through the full
// save path: a user_name edit changes exactly the name row of user.yaml;
// github_username (unmodeled) survives.
func TestHandleSaveUserNameEditSplicesOneRow(t *testing.T) {
	root := seedLosslessWebProject(t)
	a := newLosslessApp(t, root)

	form := url.Values{
		"__profile": {"default"},
		"user_name": {"renamed-user"},
	}
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("user_name edit status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}

	after := readWebSection(t, root, "user")
	changed := changedLines(t, losslessUser, after)
	if len(changed) != 1 || !strings.Contains(changed[0], "name: renamed-user") {
		t.Fatalf("user_name edit changed %d lines (%v), want exactly the name row\n--- after ---\n%s", len(changed), changed, after)
	}
	if !strings.Contains(after, "github_username: example-user") {
		t.Errorf("unmodeled github_username lost by a name edit:\n%s", after)
	}
}

// TestHandleSaveLLMEditIsolatesOtherSections carries AC-WSL-002 + AC-WSL-001
// together through the full save path: a single llm field edit changes exactly
// one llm.yaml line, and user.yaml / quality.yaml / git-convention.yaml stay
// byte-identical (the D1 collateral-rewrite defect). RED on the
// pre-implementation tree: the typed Save path re-marshaled the unedited
// sections alongside llm.yaml.
func TestHandleSaveLLMEditIsolatesOtherSections(t *testing.T) {
	root := seedLosslessWebProject(t)
	a := newLosslessApp(t, root)

	newModel := differingModelOption(t, "glm-5.2")
	form := url.Values{
		"__profile":           {"default"},
		"llm.glm.models.high": {newModel},
	}
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("llm edit status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}

	afterLLM := readWebSection(t, root, "llm")
	if !strings.Contains(afterLLM, "high: "+newModel) {
		t.Fatalf("llm edit not persisted:\n%s", afterLLM)
	}
	changed := changedLines(t, losslessLLM, afterLLM)
	if len(changed) != 1 || !strings.Contains(changed[0], "high: "+newModel) {
		t.Errorf("llm edit changed %d lines (%v), want exactly the high line\n--- after ---\n%s", len(changed), changed, afterLLM)
	}
	for _, name := range []string{"user", "quality", "git-convention"} {
		want := map[string]string{"user": losslessUser, "quality": losslessQuality, "git-convention": losslessGitConvention}[name]
		if got := readWebSection(t, root, name); got != want {
			t.Errorf("llm edit rewrote %s.yaml (collateral rewrite, AC-WSL-001)\n--- before ---\n%s\n--- after ---\n%s", name, want, got)
		}
	}
}
