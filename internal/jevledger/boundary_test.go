package jevledger

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AC-JEVL-014 (SPEC-JEV-LEDGER-001): the recorder and the consumer-config
// resolver stay outside the request path. The import check is written
// STRICTER than the AC wording (binding run condition N1): instead of a short
// denylist, it is an allowlist plus a network/exec denylist, so net/rpc,
// net/smtp, os/exec and a third-party HTTP client are all refused.
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
// holds the two UnmarshalYAML(*yaml.Node) methods.
func classify(path string, allowYAML bool) string {
	switch {
	case networkOrExec(path):
		return "can reach the network or start a process"
	case allowYAML && path == yamlLibrary:
		return ""
	case !stdlib(path):
		return "is not the standard library"
	}
	return ""
}

// scanImports parses the import clauses of files and returns one finding per
// refused import, plus the number of files that were actually parsed.
func scanImports(t *testing.T, files []string, allowYAML bool) (findings []string, scanned int) {
	t.Helper()
	fset := token.NewFileSet()
	for _, file := range files {
		f, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		scanned++
		for _, imp := range f.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if why := classify(path, allowYAML); why != "" {
				findings = append(findings, filepath.Base(file)+" imports "+path+": "+why)
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

func TestLedgerPackage_ImportBoundary(t *testing.T) {
	t.Run("recorder-package", func(t *testing.T) {
		findings, scanned := scanImports(t, nonTestGoFiles(t, "."), false)
		if scanned == 0 {
			t.Fatal("the scan counted no file in internal/jevledger, so it established nothing")
		}
		for _, f := range findings {
			t.Errorf("(a) %s", f)
		}
	})

	t.Run("config-resolver-file", func(t *testing.T) {
		resolver := filepath.Join("..", "config", "jev_consumers.go")
		if _, err := os.Stat(resolver); err != nil {
			t.Fatalf("jev_consumers.go was not found, so the scan below would be vacuous: %v", err)
		}
		findings, scanned := scanImports(t, []string{resolver}, true)
		if scanned != 1 {
			t.Fatalf("scanned %d files, want exactly jev_consumers.go", scanned)
		}
		for _, f := range findings {
			t.Errorf("(b) %s", f)
		}
	})

	// Positive controls: the classifier refuses what it must, and allows the
	// yaml library only where the AC allows it.
	t.Run("classifier-controls", func(t *testing.T) {
		for _, path := range []string{
			"net", "net/http", "net/rpc", "net/smtp", "net/http/httputil", "os/exec",
			"github.com/modu-ai/moai-adk/internal/jev", "github.com/modu-ai/moai-adk/internal/jevcred",
			"github.com/go-resty/resty/v2", yamlLibrary,
		} {
			if classify(path, false) == "" {
				t.Errorf("(a) classifier allows %q", path)
			}
		}
		for _, path := range []string{
			"net", "net/http", "net/rpc", "net/smtp", "os/exec",
			"github.com/modu-ai/moai-adk/internal/jev", "github.com/modu-ai/moai-adk/internal/jevcred",
			"github.com/go-resty/resty/v2",
		} {
			if classify(path, true) == "" {
				t.Errorf("(b) classifier allows %q", path)
			}
		}
		if classify(yamlLibrary, true) != "" {
			t.Error("(b) classifier refuses the yaml library in the resolver file")
		}
		for _, path := range []string{"encoding/json", "fmt", "os", "path/filepath", "syscall", "time"} {
			if classify(path, false) != "" || classify(path, true) != "" {
				t.Errorf("classifier refuses the standard-library package %q", path)
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
		}{
			{"net-http", "net/http", false},
			{"net-rpc", "net/rpc", false},
			{"net-smtp", "net/smtp", false},
			{"os-exec", "os/exec", false},
			{"internal-jev", "github.com/modu-ai/moai-adk/internal/jev", false},
			{"third-party-http-client", "github.com/go-resty/resty/v2", false},
			{"yaml-in-recorder", yamlLibrary, false},
			{"resolver-net-http", "net/http", true},
			{"resolver-net-rpc", "net/rpc", true},
			{"resolver-net-smtp", "net/smtp", true},
			{"resolver-os-exec", "os/exec", true},
			{"resolver-internal-jev", "github.com/modu-ai/moai-adk/internal/jev", true},
			{"resolver-third-party-http-client", "github.com/go-resty/resty/v2", true},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				file := filepath.Join(t.TempDir(), "mutant.go")
				src := "package mutant\n\nimport _ \"" + tc.imp + "\"\n"
				if err := os.WriteFile(file, []byte(src), 0o600); err != nil {
					t.Fatal(err)
				}
				findings, scanned := scanImports(t, []string{file}, tc.allowYAML)
				if scanned != 1 || len(findings) != 1 {
					t.Errorf("scanned=%d findings=%v, want the import %q refused", scanned, findings, tc.imp)
				}
			})
		}
		// The yaml library in the resolver file is the one allowed non-stdlib
		// import: it must NOT be a finding, or the green path is unreachable.
		file := filepath.Join(t.TempDir(), "ok.go")
		if err := os.WriteFile(file, []byte("package ok\n\nimport _ \""+yamlLibrary+"\"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if findings, _ := scanImports(t, []string{file}, true); len(findings) != 0 {
			t.Errorf("the yaml library is refused in the resolver file: %v", findings)
		}
	})
}

// readBack lists the spellings of reading a file that the recorder must never
// contain (REQ-JEVL-011: the recorder never opens the ledger for reading).
var readBack = []string{"os.Open(", "os.ReadFile(", "O_RDONLY"}

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
