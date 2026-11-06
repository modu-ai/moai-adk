package sqliteuri

import (
	"database/sql"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// The Windows race matrix failed 474 tests with "invalid uri authority: C:":
// a drive path reached the driver as file://C:/… so the drive letter parsed as
// the URI authority. These tests pin the DSN text itself, so they hold on any OS.

func TestFile_WindowsDriveIsPathNotAuthority(t *testing.T) {
	cases := []struct {
		name, path, query, want string
	}{
		{"backslash drive path", `C:\Users\dev\x.db`, "", "file:///C:/Users/dev/x.db"},
		{"slash drive path", "C:/x/y.db", "", "file:///C:/x/y.db"},
		{"lowercase drive", `d:\data\z.db`, "", "file:///d:/data/z.db"},
		{"drive path with query", `C:\x\y.db`, "mode=ro", "file:///C:/x/y.db?mode=ro"},
		{"posix absolute", "/tmp/x/y.db", "", "file:///tmp/x/y.db"},
		{"special characters are encoded", `C:\a b\c#d%e.db`, "", "file:///C:/a%20b/c%23d%25e.db"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := url.ParseQuery(tc.query)
			if err != nil {
				t.Fatal(err)
			}
			if got := File(tc.path, q); got != tc.want {
				t.Fatalf("File(%q, %q) = %q, want %q", tc.path, tc.query, got, tc.want)
			}
		})
	}
}

func TestFile_HostIsNeverTheDriveLetter(t *testing.T) {
	u, err := url.Parse(File(`C:\Users\dev\x.db`, url.Values{"mode": {"ro"}}))
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "" || u.Path != "/C:/Users/dev/x.db" {
		t.Fatalf("host=%q path=%q, want empty host and /C:/Users/dev/x.db", u.Host, u.Path)
	}
}

func TestFile_RelativePathDoesNotBecomeAuthority(t *testing.T) {
	got := File("rel.db", nil)
	if !strings.HasPrefix(got, "file:///") {
		t.Fatalf("File(rel.db) = %q, want a file:/// URI", got)
	}
}

func TestFile_OpensARealFileOnThisOS(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dir with space", "a#b.db")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", File(path, url.Values{"_pragma": {"busy_timeout(100)"}}))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec(`CREATE TABLE t(x INTEGER)`); err != nil {
		t.Fatalf("create through the built DSN: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the database file is not at the requested path: %v", err)
	}
}

var handBuiltDSN = regexp.MustCompile(`"file:"\s*\+`)

// TestNoHandBuiltFileDSNInProduction keeps a hand-concatenated "file:"+path from
// returning: that form is exactly what cannot carry a Windows drive path.
func TestNoHandBuiltFileDSNInProduction(t *testing.T) {
	var hits []string
	for _, root := range []string{"../../internal", "../../cmd"} {
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
				return err
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			if handBuiltDSN.Match(b) {
				hits = append(hits, filepath.ToSlash(p))
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(hits) > 0 {
		t.Fatalf("production files build a file: DSN by hand (use sqliteuri.File): %v", hits)
	}
}
