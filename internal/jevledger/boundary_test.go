package jevledger

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// AC-JEVL-014 (SPEC-JEV-LEDGER-001): the recorder and the consumer-config
// resolver stay outside the request path. The import check is written
// STRICTER than the AC wording (binding run condition N1): instead of a short
// denylist, it is an allowlist plus a network/exec denylist, so net/rpc,
// net/smtp, os/exec and a third-party HTTP client are all refused, and
// syscall is refused outside the open-flags files (open_unix.go /
// open_other.go), the only two that need it (debt D4).
//
// The read-back check is lexical; a read through another spelling is a
// residual risk owned by review.

const yamlLibrary = "gopkg.in/yaml.v3"

// stdlib reports whether an import path is the standard library: the first
// path element of a non-standard module path contains a dot.
func stdlib(path string) bool {
	first, _, _ := strings.Cut(path, "/")
	return !strings.Contains(first, ".")
}

// networkOrExec reports the standard-library packages that can open a socket
// or start a process, whatever their sub-path.
func networkOrExec(path string) bool {
	return path == "net" || strings.HasPrefix(path, "net/") || path == "os/exec"
}

// classify returns "" when the import is allowed and the reason otherwise.
// allowYAML widens the allowlist by the yaml library for the one file that
// holds the two UnmarshalYAML(*yaml.Node) methods. allowSyscall widens it by
// syscall for the open-flags files only (open_unix.go / open_other.go): the
// recorder core must not grow syscall use unnoticed (sync audit debt D4 of
// card t1762, repaired by card t1979).
func classify(path string, allowYAML, allowSyscall bool) string {
	switch {
	case networkOrExec(path):
		return "can reach the network or start a process"
	case path == "syscall" && !allowSyscall:
		return "imports syscall, allowed only in the open-flags files"
	case allowYAML && path == yamlLibrary:
		return ""
	case !stdlib(path):
		return "is not the standard library"
	}
	return ""
}

// importFinding judges one import-path literal exactly as it is spelled in the
// source: strconv.Unquote resolves both the double-quoted and the backtick
// form, and a literal it cannot resolve is itself a finding, never an allowed
// import. It returns "" when the import is allowed.
func importFinding(file, literal string, allowYAML, allowSyscall bool) string {
	path, err := strconv.Unquote(literal)
	if err != nil {
		return file + " has an unparsable import path " + literal + ": " + err.Error()
	}
	if why := classify(path, allowYAML, allowSyscall); why != "" {
		return file + " imports " + path + ": " + why
	}
	return ""
}

// scanImports parses the import clauses of files and returns one finding per
// refused import, plus the number of files that were actually parsed.
func scanImports(t *testing.T, files []string, allowYAML, allowSyscall bool) (findings []string, scanned int) {
	t.Helper()
	fset := token.NewFileSet()
	for _, file := range files {
		f, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		scanned++
		for _, imp := range f.Imports {
			if finding := importFinding(filepath.Base(file), imp.Path.Value, allowYAML, allowSyscall); finding != "" {
				findings = append(findings, finding)
			}
		}
	}
	return findings, scanned
}

// nonTestGoFiles lists the non-test Go files of dir.
func nonTestGoFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir %s: %v", dir, err)
	}
	var out []string
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			out = append(out, filepath.Join(dir, name))
		}
	}
	return out
}

// allowSyscallInFile reports whether file is one of the two files that hold
// the open flags (open_unix.go / open_other.go), the only places syscall is
// allowed in this package. The allowlist is EXACT — a basename of open_*.go
// is not enough (leader audit F-B, repair cycle R2 of card t1979): the prefix
// form admitted syscall in any file merely named open_<anything>.go.
func allowSyscallInFile(file string) bool {
	switch filepath.Base(file) {
	case "open_unix.go", "open_other.go":
		return true
	}
	return false
}

func TestLedgerPackage_ImportBoundary(t *testing.T) {
	t.Run("recorder-package", func(t *testing.T) {
		files := nonTestGoFiles(t, ".")
		if len(files) == 0 {
			t.Fatal("the scan counted no file in internal/jevledger, so it established nothing")
		}
		for _, file := range files {
			// syscall is allowed only where the open flags live (open_unix.go,
			// open_other.go), never in the recorder core (debt D4).
			findings, _ := scanImports(t, []string{file}, false, allowSyscallInFile(file))
			for _, f := range findings {
				t.Errorf("(a) %s", f)
			}
		}
	})

	// Negative control (leader audit F-B, repair cycle R2): the syscall
	// allowlist is exactly the two open-flags files. A file merely NAMED
	// open_*.go — open_zzexample.go here, synthetic, in a t.TempDir() outside
	// the package and outside the repo — must be REFUSED syscall.
	t.Run("syscall-allowlist-is-exact", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "open_zzexample.go")
		if err := os.WriteFile(file, []byte("package mutant\n\nimport _ \"syscall\"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		findings, scanned := scanImports(t, []string{file}, false, allowSyscallInFile(file))
		if scanned != 1 || len(findings) != 1 {
			t.Errorf("scanned=%d findings=%v, want the syscall import in open_zzexample.go refused", scanned, findings)
		}
	})

	t.Run("config-resolver-file", func(t *testing.T) {
		resolver := filepath.Join("..", "config", "jev_consumers.go")
		if _, err := os.Stat(resolver); err != nil {
			t.Fatalf("jev_consumers.go was not found, so the scan below would be vacuous: %v", err)
		}
		findings, scanned := scanImports(t, []string{resolver}, true, false)
		if scanned != 1 {
			t.Fatalf("scanned %d files, want exactly jev_consumers.go", scanned)
		}
		for _, f := range findings {
			t.Errorf("(b) %s", f)
		}
	})

	// Positive controls: the classifier refuses what it must, and allows the
	// yaml library only where the AC allows it, and syscall only in the
	// open-flags files.
	t.Run("classifier-controls", func(t *testing.T) {
		for _, path := range []string{
			"net", "net/http", "net/rpc", "net/smtp", "net/http/httputil", "os/exec", "syscall",
			"github.com/modu-ai/moai-adk/internal/jev", "github.com/modu-ai/moai-adk/internal/jevcred",
			"github.com/go-resty/resty/v2", yamlLibrary,
		} {
			if classify(path, false, false) == "" {
				t.Errorf("(a) classifier allows %q", path)
			}
		}
		for _, path := range []string{
			"net", "net/http", "net/rpc", "net/smtp", "os/exec", "syscall",
			"github.com/modu-ai/moai-adk/internal/jev", "github.com/modu-ai/moai-adk/internal/jevcred",
			"github.com/go-resty/resty/v2",
		} {
			if classify(path, true, false) == "" {
				t.Errorf("(b) classifier allows %q", path)
			}
		}
		if classify(yamlLibrary, true, false) != "" {
			t.Error("(b) classifier refuses the yaml library in the resolver file")
		}
		if classify("syscall", false, true) != "" {
			t.Error("classifier refuses syscall in an open-flags file")
		}
		for _, path := range []string{"encoding/json", "errors", "fmt", "io/fs", "os", "path/filepath", "time"} {
			if classify(path, false, false) != "" || classify(path, true, false) != "" {
				t.Errorf("classifier refuses the standard-library package %q", path)
			}
		}
	})

	// A literal that cannot be resolved to a path is a finding in its own right:
	// the scanner must not let an import it cannot read through as allowed.
	t.Run("unparsable-literal-is-a-finding", func(t *testing.T) {
		for _, literal := range []string{`"net/http`, "`net/http", `net/http`, `'xy'`, ``} {
			if importFinding("x.go", literal, false, false) == "" {
				t.Errorf("an import literal that does not resolve (%q) was allowed", literal)
			}
		}
		for _, literal := range []string{`"fmt"`, "`fmt`"} {
			if got := importFinding("x.go", literal, false, false); got != "" {
				t.Errorf("the standard-library import %s was refused: %s", literal, got)
			}
		}
	})

	// Mutant probe (run condition N1): each refused import, written into a real
	// file and fed through the scanner, must produce a finding. A scanner that
	// silently matched no file, or matched everything, fails here.
	t.Run("mutants", func(t *testing.T) {
		cases := []struct {
			name, imp string
			allowYAML bool
			allowSys  bool // widen the allowlist by syscall, as the open-flags files do
			raw       bool // spell the import path as a backtick string literal
		}{
			{"net-http", "net/http", false, false, false},
			{"net-rpc", "net/rpc", false, false, false},
			{"net-smtp", "net/smtp", false, false, false},
			{"os-exec", "os/exec", false, false, false},
			{"syscall-in-recorder", "syscall", false, false, false},
			{"internal-jev", "github.com/modu-ai/moai-adk/internal/jev", false, false, false},
			{"third-party-http-client", "github.com/go-resty/resty/v2", false, false, false},
			{"yaml-in-recorder", yamlLibrary, false, false, false},
			{"resolver-net-http", "net/http", true, false, false},
			{"resolver-net-rpc", "net/rpc", true, false, false},
			{"resolver-net-smtp", "net/smtp", true, false, false},
			{"resolver-os-exec", "os/exec", true, false, false},
			{"resolver-syscall", "syscall", true, false, false},
			{"resolver-internal-jev", "github.com/modu-ai/moai-adk/internal/jev", true, false, false},
			{"resolver-third-party-http-client", "github.com/go-resty/resty/v2", true, false, false},
			// The same refusals spelled with backticks (sync audit F3): a scanner
			// that only trims double quotes passes every one of these.
			{"raw-net-http", "net/http", false, false, true},
			{"raw-os-exec", "os/exec", false, false, true},
			{"raw-internal-jev", "github.com/modu-ai/moai-adk/internal/jev", false, false, true},
			{"raw-yaml-in-recorder", yamlLibrary, false, false, true},
			{"raw-syscall-in-recorder", "syscall", false, false, true},
			{"raw-resolver-net-http", "net/http", true, false, true},
			{"raw-resolver-os-exec", "os/exec", true, false, true},
			{"raw-resolver-third-party-http-client", "github.com/go-resty/resty/v2", true, false, true},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				file := filepath.Join(t.TempDir(), "mutant.go")
				quote := `"`
				if tc.raw {
					quote = "`"
				}
				src := "package mutant\n\nimport _ " + quote + tc.imp + quote + "\n"
				if err := os.WriteFile(file, []byte(src), 0o600); err != nil {
					t.Fatal(err)
				}
				findings, scanned := scanImports(t, []string{file}, tc.allowYAML, tc.allowSys)
				if scanned != 1 || len(findings) != 1 {
					t.Errorf("scanned=%d findings=%v, want the import %q refused", scanned, findings, tc.imp)
				}
			})
		}
		// The yaml library in the resolver file is the one allowed non-stdlib
		// import: it must NOT be a finding, or the green path is unreachable.
		for _, quote := range []string{`"`, "`"} {
			file := filepath.Join(t.TempDir(), "ok.go")
			if err := os.WriteFile(file, []byte("package ok\n\nimport _ "+quote+yamlLibrary+quote+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			if findings, _ := scanImports(t, []string{file}, true, false); len(findings) != 0 {
				t.Errorf("the yaml library is refused in the resolver file (%s quoting): %v", quote, findings)
			}
		}
		// syscall in an open-flags file is the one allowed syscall import: it
		// must NOT be a finding, or open_unix.go itself could never pass (D4).
		for _, quote := range []string{`"`, "`"} {
			file := filepath.Join(t.TempDir(), "open_ok.go")
			if err := os.WriteFile(file, []byte("package ok\n\nimport _ "+quote+"syscall"+quote+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			if findings, _ := scanImports(t, []string{file}, false, true); len(findings) != 0 {
				t.Errorf("syscall is refused in an open-flags file (%s quoting): %v", quote, findings)
			}
		}
	})
}

// readBack lists the spellings of reading a file that the recorder must never
// contain (REQ-JEVL-011: the recorder never opens the ledger for reading).
// os.O_RDWR is listed as its qualified spelling (sync audit debt D5 of card
// t1762, repaired by card t1979): a bare "O_RDWR" would false-positive on a
// hypothetical constant of the same name, and an open that can read is a read.
var readBack = []string{"os.Open(", "os.ReadFile(", "O_RDONLY", "os.O_RDWR"}

func findReadBack(src string) []string {
	var hits []string
	for _, spelling := range readBack {
		if strings.Contains(src, spelling) {
			hits = append(hits, spelling)
		}
	}
	return hits
}

func TestLedgerPackage_NeverReadsLedger(t *testing.T) {
	files := nonTestGoFiles(t, ".")
	if len(files) == 0 {
		t.Fatal("the scan counted no file in internal/jevledger, so it established nothing")
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if hits := findReadBack(string(data)); len(hits) > 0 {
			t.Errorf("%s contains %v", filepath.Base(file), hits)
		}
	}
	// Positive control: the detector can fail.
	for _, spelling := range readBack {
		if len(findReadBack("x := "+spelling+"path)")) == 0 {
			t.Errorf("the detector does not flag %q", spelling)
		}
	}
	// os.OpenFile is the write-side open and must not be mistaken for os.Open(.
	if len(findReadBack("os.OpenFile(path, flags, 0o600)")) != 0 {
		t.Error("the detector flags os.OpenFile, so the recorder's own write open could never pass")
	}
}
