package web

// profile_crud_branches_test.go — branches of profile_crud.go that no test reached
// (Issue #1948, B15): the directory helpers' name guards, a malformed form body on
// the three CRUD routes, and rename to the same name. Tests only.
//
// Defect answer: these tests catch a reserved or traversal name reaching the
// filesystem helpers, a malformed body that is accepted instead of refused with
// 400, and a rename whose new name equals the old one being attempted.

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProfileDirHelpersRefuseReservedAndTraversalNames(t *testing.T) {
	for _, name := range []string{"", "default", "../escape", "a/b", ".hidden"} {
		err := createProfileDir(name)
		if err == nil || !strings.Contains(err.Error(), "invalid or reserved profile name") {
			t.Errorf("createProfileDir(%q) = %v, want an invalid-or-reserved refusal", name, err)
		}
	}
	base := t.TempDir()
	t.Setenv("MOAI_HOME", base)
	before := dirEntries(t, base)
	for _, tc := range []struct{ oldName, newName, want string }{
		{"default", "fresh", "invalid or reserved profile name"},
		{"", "fresh", "invalid or reserved profile name"},
		{"zz-old", "../escape", "invalid or reserved profile name"},
		{"zz-old", "default", "invalid or reserved profile name"},
	} {
		err := renameProfileDir(tc.oldName, tc.newName)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("renameProfileDir(%q, %q) = %v, want %q", tc.oldName, tc.newName, err, tc.want)
		}
	}
	if after := dirEntries(t, base); after != before {
		t.Errorf("a refused name touched the filesystem: before %q after %q", before, after)
	}
}

func dirEntries(t *testing.T, dir string) string {
	t.Helper()
	var names []string
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err == nil {
			names = append(names, p)
		}
		return nil
	})
	return strings.Join(names, "|")
}

func postRawForm(t *testing.T, h http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestProfileRoutesRefuseAMalformedFormBody(t *testing.T) {
	a := newTestApp(t)
	called := false
	a.createProfile = func(string) error { called = true; return nil }
	a.deleteProfile = func(string) error { called = true; return nil }
	a.renameProfile = func(string, string) error { called = true; return nil }
	for _, path := range []string{"/profile/create", "/profile/rename", "/profile/delete"} {
		rec := postRawForm(t, a.routes(), path, "profile_name=%zz")
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "could not parse form") {
			t.Errorf("POST %s with a malformed body = %d, want 400 could not parse form", path, rec.Code)
		}
	}
	if called {
		t.Errorf("a malformed body reached a profile seam")
	}
}

func TestProfileRenameToTheSameNameIsRefused(t *testing.T) {
	a := newTestApp(t)
	called := false
	a.renameProfile = func(string, string) error { called = true; return nil }
	rec := servePost(t, a.routes(), "/profile/rename", url.Values{"profile_name": {"zz-same"}, "new_name": {"zz-same"}})
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "same as the current one") {
		t.Errorf("rename to the same name = %d, want 400 same as the current one", rec.Code)
	}
	if called {
		t.Errorf("the rename seam ran for an identical name")
	}
}
