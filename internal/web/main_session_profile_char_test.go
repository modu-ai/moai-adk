package web

// Characterisation tests for the web console surfaces that the subagent
// model/effort inheritance change must leave untouched: the user-preference
// profile routes (create / rename / delete) and the main-session effort save.
// They pin the behaviour observed before any web edit (status codes and the
// files each request writes) so later milestones can prove it unchanged.
// Not parallel-safe: they swap profile.BaseDirOverride and CLAUDE_CONFIG_DIR.

import (
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// charProfileApp builds an app whose profile store lives in a temp dir, whose
// active profile resolves to "default", and whose last-profile ledger write is
// stubbed so no test reaches the developer's home directory.
func charProfileApp(t *testing.T) (a *app, profileBase, projectRoot string) {
	t.Helper()
	profileBase = t.TempDir()
	orig := profile.BaseDirOverride
	profile.BaseDirOverride = profileBase
	t.Cleanup(func() { profile.BaseDirOverride = orig })
	t.Setenv(envClaudeConfigDir, "")

	projectRoot = t.TempDir()
	a = newApp(Config{ProjectRoot: projectRoot, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }
	return a, profileBase, projectRoot
}

// listFiles returns every regular file under root, relative and sorted.
func listFiles(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			rel, rerr := filepath.Rel(root, p)
			if rerr != nil {
				return rerr
			}
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	sort.Strings(out)
	return out
}

func TestCharacterize_PreferenceProfileCreateRenameDelete(t *testing.T) {
	a, base, _ := charProfileApp(t)
	h := a.routes()

	rec := postProfile(t, h, "/profile/create", "work")
	if rec.Code != http.StatusOK {
		t.Fatalf("create status = %d, want 200", rec.Code)
	}
	if fi, err := os.Stat(filepath.Join(base, "work")); err != nil || !fi.IsDir() {
		t.Fatalf("create: profile dir missing (err=%v)", err)
	}

	form := url.Values{"profile_name": {"work"}, "new_name": {"play"}}
	rec = servePost(t, h, profileRenameAction(), form)
	if rec.Code != http.StatusOK {
		t.Fatalf("rename status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(base, "work")); !os.IsNotExist(err) {
		t.Errorf("rename: old dir still present (err=%v)", err)
	}
	if fi, err := os.Stat(filepath.Join(base, "play")); err != nil || !fi.IsDir() {
		t.Errorf("rename: new dir missing (err=%v)", err)
	}

	rec = postProfile(t, h, "/profile/delete", "play")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want 200", rec.Code)
	}
	if _, err := os.Stat(filepath.Join(base, "play")); !os.IsNotExist(err) {
		t.Errorf("delete: dir still present (err=%v)", err)
	}

	// Guards: default and the active profile are refused with a 4xx.
	for _, name := range []string{"default"} {
		rec = postProfile(t, h, "/profile/delete", name)
		if rec.Code < 400 || rec.Code >= 500 {
			t.Errorf("delete %q status = %d, want 4xx", name, rec.Code)
		}
	}
}

func TestCharacterize_MainSessionEffortSaveWritesPreferencesOnly(t *testing.T) {
	a, base, projectRoot := charProfileApp(t)
	h := a.routes()

	form := url.Values{
		"__profile":       {"default"},
		"permission_mode": {"acceptEdits"},
		"effort_level":    {"xhigh"},
	}
	rec := servePost(t, h, "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}

	prefs, err := profile.ReadPreferences("default")
	if err != nil {
		t.Fatalf("read preferences: %v", err)
	}
	if prefs.EffortLevel != "xhigh" {
		t.Errorf("persisted effort_level = %q, want xhigh", prefs.EffortLevel)
	}
	if prefs.PermissionMode != "acceptEdits" {
		t.Errorf("persisted permission_mode = %q, want acceptEdits", prefs.PermissionMode)
	}

	if got := listFiles(t, base); len(got) != 1 || got[0] != "preferences.yaml" {
		t.Errorf("profile store files = %v, want [preferences.yaml]", got)
	}
	// The main-session save never writes llm.yaml or any agent definition.
	for _, f := range listFiles(t, projectRoot) {
		if strings.HasSuffix(f, "sections/llm.yaml") || strings.HasPrefix(f, ".claude/agents/") {
			t.Errorf("main-session save wrote %s", f)
		}
	}
}
