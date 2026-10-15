package testguard

import (
	"os"
	"path/filepath"
	"testing"
)

// Fake path strings only: none of these exist, so the judgment runs on the
// string rules and works on every platform without a build tag.
const (
	fakeHome = "/fake/users/me"
	fakeTmp  = "/fake/users/me/AppData/Local/Temp" // temp dir under the home (Windows layout)
)

func TestTempExempt(t *testing.T) {
	cases := []struct {
		name     string
		home     string
		tmp      string
		path     string
		wantExem bool
	}{
		{"temp under home, home not under temp: exempt", fakeHome, fakeTmp, fakeTmp + "/x/gitconfig", true},
		{"temp equal to path root: exempt", fakeHome, fakeTmp, fakeTmp, true},
		{"home under temp: not exempt", fakeTmp + "/h", fakeTmp, fakeTmp + "/h/x", false},
		{"home equal to temp: not exempt", fakeTmp, fakeTmp, fakeTmp + "/x", false},
		{"sibling of temp under home: not exempt", fakeHome, fakeTmp, fakeHome + "/AppData/Local/Temp2/x", false},
		{"home dir itself: not exempt", fakeHome, fakeTmp, fakeHome, false},
		{"outside both: not exempt", fakeHome, fakeTmp, "/elsewhere/x", false},
		{"relative path: not exempt", fakeHome, fakeTmp, "AppData/Local/Temp/x", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tempExempt(tc.home, tc.tmp, tc.path); got != tc.wantExem {
				t.Fatalf("tempExempt(%q, %q, %q) = %v, want %v", tc.home, tc.tmp, tc.path, got, tc.wantExem)
			}
		})
	}
}

func TestCheckGlobalConfigPin(t *testing.T) {
	cases := []struct {
		name    string
		home    string
		tmp     string
		file    string
		wantErr bool
	}{
		{"temp file under home via temp dir: accepted", fakeHome, fakeTmp, fakeTmp + "/moai-cli-gitconfig-1", false},
		{"temp file outside home: accepted", fakeHome, "/fake/tmp", "/fake/tmp/moai-cli-gitconfig-1", false},
		{"real global config: refused", fakeHome, "/fake/tmp", fakeHome + "/.gitconfig", true},
		{"real global config even when temp is under home: refused", fakeHome, fakeTmp, fakeHome + "/.gitconfig", true},
		{"file under home outside the temp dir: refused", fakeHome, fakeTmp, fakeHome + "/docs/gitconfig", true},
		{"sibling temp dir under home: refused", fakeHome, fakeTmp, fakeHome + "/AppData/Local/Temp2/gitconfig", true},
		{"home under temp, file under home: refused", fakeTmp + "/h", fakeTmp, fakeTmp + "/h/gitconfig", true},
		{"relative file: refused", fakeHome, fakeTmp, "gitconfig", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkGlobalConfigPin(tc.home, tc.tmp, tc.file)
			if (err != nil) != tc.wantErr {
				t.Fatalf("checkGlobalConfigPin(%q, %q, %q) error = %v, wantErr %v", tc.home, tc.tmp, tc.file, err, tc.wantErr)
			}
		})
	}
}

// The exported wrappers read os.TempDir: a file there is accepted when the home
// is unrelated to it, and the home's own global config is never accepted.
func TestExportedWrappersUseOSTempDir(t *testing.T) {
	root := filepath.Join(os.TempDir(), "moai-wrapper-probe")
	home := filepath.Join(t.TempDir(), "outside-home")
	if err := CheckGlobalConfigPin(home, filepath.Join(root, "gitconfig")); err != nil {
		t.Fatalf("a file under os.TempDir with an unrelated home was refused: %v", err)
	}
	if TempDirExemption(home, root) {
		t.Fatal("exemption granted although the home lies under os.TempDir")
	}
	if err := CheckGlobalConfigPin(home, filepath.Join(home, ".gitconfig")); err == nil {
		t.Fatal("the real global config path was accepted")
	}
}
