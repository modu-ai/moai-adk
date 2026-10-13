package hook

// t1619_zone_coverage_test.go — card t1619 bundle A: changed-line closure for
// the user-root and user-manifest protection helpers of the hook package. Each
// test drives a real input through the production helper and asserts the
// observed outcome. The package-level home and working-directory lookups are
// replaced only inside the test, and the originals are restored on cleanup.
// No production code is changed.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// t1619SetHome points zoneHomeFn at home for the duration of the test.
func t1619SetHome(t *testing.T, home string) {
	t.Helper()
	orig := zoneHomeFn
	zoneHomeFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { zoneHomeFn = orig })
}

// t1619FailHome makes the home lookup fail for the duration of the test.
func t1619FailHome(t *testing.T) {
	t.Helper()
	orig := zoneHomeFn
	zoneHomeFn = func() (string, error) { return "", errors.New("no home directory") }
	t.Cleanup(func() { zoneHomeFn = orig })
}

// t1619FailGetwd makes the working-directory lookup fail for the duration of
// the test.
func t1619FailGetwd(t *testing.T) {
	t.Helper()
	orig := zoneGetwd
	zoneGetwd = func() (string, error) { return "", errors.New("no working directory") }
	t.Cleanup(func() { zoneGetwd = orig })
}

// t1619WriteFile writes body to path, creating the parent directories.
func t1619WriteFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// t1619HasDisplay reports whether forms carries a form with the given display.
func t1619HasDisplay(forms []zoneForm, display string) bool {
	for _, f := range forms {
		if f.Display == display {
			return true
		}
	}
	return false
}

// TestUserRootZoneFormsNilWithoutHomeOrWorkingDirectory covers the two
// fail-closed exits of the user-root and user-manifest form builders: no home
// directory, and a relative candidate with no working directory to join onto.
func TestUserRootZoneFormsNilWithoutHomeOrWorkingDirectory(t *testing.T) {
	t.Run("home lookup fails", func(t *testing.T) {
		t1619FailHome(t)
		if forms := userRootZoneForms("/anywhere/file"); forms != nil {
			t.Fatalf("userRootZoneForms with no home = %+v, want nil", forms)
		}
		if forms := userManifestProtectForms("~/.moai/user-assets.json"); forms != nil {
			t.Fatalf("userManifestProtectForms with no home = %+v, want nil", forms)
		}
	})
	t.Run("relative candidate without a working directory", func(t *testing.T) {
		t1619SetHome(t, t.TempDir())
		t1619FailGetwd(t)
		if forms := userRootZoneForms("relative/file"); forms != nil {
			t.Fatalf("userRootZoneForms of a relative path = %+v, want nil", forms)
		}
		if forms := userManifestProtectForms("relative/.moai/user-assets.json"); forms != nil {
			t.Fatalf("userManifestProtectForms of a relative path = %+v, want nil", forms)
		}
	})
}

// TestUserRootKeyParsesOnlyUserRootForms covers the namespace parser: only a
// user-root form with a non-empty rest yields a manifest key.
func TestUserRootKeyParsesOnlyUserRootForms(t *testing.T) {
	if key, ok := userRootKey("user-root:claude-skills/moai-x/skill.md"); !ok || key != "claude-skills/moai-x/skill.md" {
		t.Fatalf("userRootKey of a user-root form = (%q, %v), want the manifest key", key, ok)
	}
	for _, display := range []string{"claude-skills/moai-x/skill.md", "user-root:"} {
		if key, ok := userRootKey(display); ok {
			t.Fatalf("userRootKey(%q) = (%q, true), want no key", display, key)
		}
	}
}

// TestUserRootManagedKeysFailsClosedOnUnparseableManifest covers the managed-key
// reader: a manifest whose JSON does not parse yields no keys.
func TestUserRootManagedKeysFailsClosedOnUnparseableManifest(t *testing.T) {
	home := t.TempDir()
	t1619WriteFile(t, filepath.Join(home, ".moai", "user-assets.json"), "{ not json")
	if keys := userRootManagedKeys(home); keys != nil {
		t.Fatalf("userRootManagedKeys of an unparseable manifest = %v, want nil", keys)
	}
}

// TestUserRootTracksAnyFailsClosedWithoutHomeOrReadableManifest covers the
// tracked-ness judgment's fail-closed exits: a display that is not a user-root
// form, no home, no manifest, an unparseable manifest, and a root slug the
// install roots do not name.
func TestUserRootTracksAnyFailsClosedWithoutHomeOrReadableManifest(t *testing.T) {
	t.Run("display is not a user-root form", func(t *testing.T) {
		home := t.TempDir()
		t1619WriteFile(t, filepath.Join(home, ".moai", "user-assets.json"), `{"files": {"claude-skills/a.md": {}}}`)
		if userRootTracksAny(home, "claude-skills/a.md") {
			t.Fatal("a bare key without the user-root prefix read as tracked")
		}
	})
	t.Run("no home directory", func(t *testing.T) {
		t1619FailHome(t)
		if userRootTracksAny("", "user-root:claude-skills/a.md") {
			t.Fatal("a form with no resolvable home read as tracked")
		}
	})
	t.Run("manifest absent", func(t *testing.T) {
		if userRootTracksAny(t.TempDir(), "user-root:claude-skills/a.md") {
			t.Fatal("a form with no manifest read as tracked")
		}
	})
	t.Run("manifest unparseable", func(t *testing.T) {
		home := t.TempDir()
		t1619WriteFile(t, filepath.Join(home, ".moai", "user-assets.json"), "{ not json")
		if userRootTracksAny(home, "user-root:claude-skills/a.md") {
			t.Fatal("a form under an unparseable manifest read as tracked")
		}
	})
	t.Run("unknown root slug", func(t *testing.T) {
		home := t.TempDir()
		t1619WriteFile(t, filepath.Join(home, ".moai", "user-assets.json"), `{"files": {"claude-skills/a.md": {}}}`)
		if userRootTracksAny(home, "user-root:bogus-slug/x.md") {
			t.Fatal("a form naming an unknown root slug read as tracked")
		}
	})
}

// TestUserRootZoneFormsSkipsMalformedManagedKeys covers the managed-key loop of
// the form builder: keys with an unknown slug, no rest, or a parent-directory
// segment are skipped, and the legitimate tracked file still yields its form.
func TestUserRootZoneFormsSkipsMalformedManagedKeys(t *testing.T) {
	home := t.TempDir()
	t1619SetHome(t, home)
	skill := filepath.Join(home, ".claude", "skills", "moai-x", "skill.md")
	t1619WriteFile(t, skill, "x\n")
	t1619WriteFile(t, filepath.Join(home, ".moai", "user-assets.json"),
		`{"files": {"bogus-slug/x.md": {}, "claude-skills": {}, "claude-skills/../escape.md": {}, "claude-skills/moai-x/skill.md": {}}}`)

	forms := userRootZoneForms(skill)
	if !t1619HasDisplay(forms, "user-root:claude-skills/moai-x/skill.md") {
		t.Fatalf("the legitimate tracked file lost its form: %+v", forms)
	}
	for _, f := range forms {
		if strings.Contains(f.Display, "bogus-slug") || strings.Contains(f.Display, "escape") {
			t.Fatalf("a malformed managed key produced a form: %q", f.Display)
		}
	}
}

// TestZoneUserRootCoveredWithoutUserRootEntry covers the compiled-in manifest
// protection. With a readable zone and no user-root entry declared, the user
// manifest is still covered as the user_manifest category. A zone that did not
// load is not judged this way.
func TestZoneUserRootCoveredWithoutUserRootEntry(t *testing.T) {
	t1619SetHome(t, t.TempDir())
	category, display, covered := zoneUserRootCovered("~/.moai/user-assets.json",
		config.ProtectedZoneLoad{State: config.ZoneStateOK})
	if !covered || category != "user_manifest" || display != ".moai/user-assets.json" {
		t.Fatalf("zoneUserRootCovered(manifest, ok zone) = (%q, %q, %v), want (user_manifest, .moai/user-assets.json, true)", category, display, covered)
	}

	category, display, covered = zoneUserRootCovered("~/.moai/user-assets.json",
		config.ProtectedZoneLoad{State: config.ZoneStateAbsent})
	if covered || category != "" || display != "" {
		t.Fatalf("zoneUserRootCovered with an absent zone = (%q, %q, %v), want no coverage", category, display, covered)
	}
}
