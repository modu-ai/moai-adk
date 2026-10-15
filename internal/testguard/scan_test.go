package testguard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// tilde is the home-expansion prefix. It is assembled from two literals on
// purpose: this file is itself scanned by the repository-wide guard below, and
// a single string literal holding the prefix would be flagged as the very
// pattern under test. Every "~/" in this file goes through this variable.
var tilde = "~" + "/"

// incidentSource reproduces the t1778 incident: the real home is read through
// os.UserHomeDir and the result is written to.
const incidentSource = `package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDoctorCodexGlobalMemory(t *testing.T) {
	home, _ := os.UserHomeDir()
	_ = os.WriteFile(filepath.Join(home, ".codex", "memories", "MEMORY.md"), []byte("x"), 0o644)
}
`

// fixtureBody wraps body lines in a function of a package that imports os.
func fixtureBody(body ...string) string {
	return "package cli\n\nimport \"os\"\n\nfunc f() {\n" + strings.Join(body, "\n") + "\n}\n"
}

// lineOf returns the 1-based line of the first line of src that contains needle.
func lineOf(t *testing.T, src, needle string) int {
	t.Helper()
	for i, line := range strings.Split(src, "\n") {
		if strings.Contains(line, needle) {
			return i + 1
		}
	}
	t.Fatalf("fixture has no line containing %q", needle)
	return 0
}

// writeFile creates path (and its parents) with content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// findModuleRoot returns the nearest ancestor of start (inclusive) that holds a go.mod.
func findModuleRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// TestScanSource_FlagsIncidentPattern is the RED case from the card: an
// os.UserHomeDir read followed by a write to the result must be reported.
func TestScanSource_FlagsIncidentPattern(t *testing.T) {
	vs, err := ScanSource("codex_memory_test.go", []byte(incidentSource))
	if err != nil {
		t.Fatalf("ScanSource: %v", err)
	}
	want := lineOf(t, incidentSource, "os.UserHomeDir()")
	if len(vs) != 1 || vs[0].Kind != KindUserHomeDir || vs[0].Line != want {
		t.Fatalf("ScanSource(incident fixture) = %+v; want exactly one %q violation at line %d",
			vs, KindUserHomeDir, want)
	}
}

// TestScanSource_FlagsTildeLiteral reports a "~/" string literal.
func TestScanSource_FlagsTildeLiteral(t *testing.T) {
	src := "package cli\n\nconst codexMemory = \"" + tilde + ".codex/memories/MEMORY.md\"\n"
	vs, err := ScanSource("tilde_test.go", []byte(src))
	if err != nil {
		t.Fatalf("ScanSource: %v", err)
	}
	want := lineOf(t, src, "codexMemory")
	if len(vs) != 1 || vs[0].Kind != KindTildeLiteral || vs[0].Line != want {
		t.Fatalf("ScanSource(tilde fixture) = %+v; want exactly one %q violation at line %d",
			vs, KindTildeLiteral, want)
	}
}

// TestScanSource_TildeInsideRawStringReportsItsOwnLine checks that a "~/" deep
// in a multi-line raw string is reported on the line where it actually sits.
func TestScanSource_TildeInsideRawStringReportsItsOwnLine(t *testing.T) {
	src := "package cli\n\nconst fixture = `first\nsecond " + tilde + ".codex\nthird`\n"
	vs, err := ScanSource("raw_test.go", []byte(src))
	if err != nil {
		t.Fatalf("ScanSource: %v", err)
	}
	want := lineOf(t, src, "second")
	if len(vs) != 1 || vs[0].Kind != KindTildeLiteral || vs[0].Line != want {
		t.Fatalf("ScanSource(raw fixture) = %+v; want one %q violation at line %d",
			vs, KindTildeLiteral, want)
	}
}

// TestScanSource_IgnoresCommentsAndOtherSelectors: a comment mentioning the
// call, the environment read of HOME, and a bare "~" element are not home reads.
func TestScanSource_IgnoresCommentsAndOtherSelectors(t *testing.T) {
	src := "package cli\n\nimport (\n\t\"os\"\n\t\"path/filepath\"\n)\n\n" +
		"// os.UserHomeDir() is described here, and \"" + tilde + ".codex\" is named in prose.\n" +
		"func home() string {\n" +
		"\t_ = os.Getenv(\"HOME\")\n" +
		"\treturn filepath.Join(\"~\", \".codex\")\n" +
		"}\n"
	vs, err := ScanSource("clean_test.go", []byte(src))
	if err != nil {
		t.Fatalf("ScanSource: %v", err)
	}
	if len(vs) != 0 {
		t.Fatalf("ScanSource(clean fixture) = %+v; want no violations", vs)
	}
}

// TestScanSource_AllowlistExemptsSameLineOrLineAbove: the marker exempts a
// violation on its own line or on the line directly above, and nowhere else.
func TestScanSource_AllowlistExemptsSameLineOrLineAbove(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"same_line", fixtureBody("\thome, _ := os.UserHomeDir() // home-guard:allow compares the resolver output only")},
		{"line_above", fixtureBody("\t// home-guard:allow compares the resolver output only", "\thome, _ := os.UserHomeDir()")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vs, err := ScanSource("allow_test.go", []byte(tc.src))
			if err != nil {
				t.Fatalf("ScanSource: %v", err)
			}
			if len(vs) != 0 {
				t.Fatalf("ScanSource = %+v; want the allowlisted call exempt", vs)
			}
		})
	}
}

// TestScanSource_AllowlistNeedsAReason: a bare marker does not exempt.
func TestScanSource_AllowlistNeedsAReason(t *testing.T) {
	src := fixtureBody("\t// home-guard:allow", "\thome, _ := os.UserHomeDir()", "\t_ = home")
	vs, err := ScanSource("bare_marker_test.go", []byte(src))
	if err != nil {
		t.Fatalf("ScanSource: %v", err)
	}
	if len(vs) != 1 {
		t.Fatalf("ScanSource = %+v; want one violation: a marker without a reason exempts nothing", vs)
	}
}

// TestScanSource_AllowlistTwoLinesAboveDoesNotExempt: only the directly
// preceding line counts.
func TestScanSource_AllowlistTwoLinesAboveDoesNotExempt(t *testing.T) {
	src := fixtureBody("\t// home-guard:allow compares the resolver output only", "\t_ = 1", "\thome, _ := os.UserHomeDir()")
	vs, err := ScanSource("far_marker_test.go", []byte(src))
	if err != nil {
		t.Fatalf("ScanSource: %v", err)
	}
	if len(vs) != 1 {
		t.Fatalf("ScanSource = %+v; want one violation: the marker is two lines above", vs)
	}
}

// TestScanSource_TracksAliasesAndShadowing: an aliased import of os is still
// os, and a local variable named os is not the package.
func TestScanSource_TracksAliasesAndShadowing(t *testing.T) {
	alias := "package cli\n\nimport host \"os\"\n\nfunc f() {\n\t_, _ = host.UserHomeDir()\n}\n"
	vs, err := ScanSource("alias_test.go", []byte(alias))
	if err != nil {
		t.Fatalf("ScanSource(alias): %v", err)
	}
	if len(vs) != 1 || vs[0].Kind != KindUserHomeDir {
		t.Fatalf("ScanSource(alias) = %+v; want one %q violation", vs, KindUserHomeDir)
	}

	shadow := "package cli\n\nimport \"os\"\n\ntype fake struct{}\n\n" +
		"func (fake) UserHomeDir() (string, error) { return \"\", nil }\n\n" +
		"func f() {\n\tos := fake{}\n\t_, _ = os.UserHomeDir()\n}\n"
	vs, err = ScanSource("shadow_test.go", []byte(shadow))
	if err != nil {
		t.Fatalf("ScanSource(shadow): %v", err)
	}
	if len(vs) != 0 {
		t.Fatalf("ScanSource(shadow) = %+v; a local os value is not the package", vs)
	}
}

// TestScanTree_ScansTestFilesAndSkipsTestdataAndNonTestFiles pins the walk:
// only *_test.go files under the named dirs, never testdata or vendor.
func TestScanTree_ScansTestFilesAndSkipsTestdataAndNonTestFiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "internal", "a", "a_test.go"), incidentSource)
	writeFile(t, filepath.Join(root, "internal", "a", "a.go"), incidentSource)
	writeFile(t, filepath.Join(root, "internal", "a", "testdata", "b_test.go"), incidentSource)
	writeFile(t, filepath.Join(root, "cmd", "c", "c_test.go"), "package c\n")
	writeFile(t, filepath.Join(root, "other", "d_test.go"), incidentSource)

	res, err := ScanTree(root, []string{"internal", "cmd"})
	if err != nil {
		t.Fatalf("ScanTree: %v", err)
	}
	if res.Files != 2 {
		t.Fatalf("ScanTree parsed %d test files; want 2 (a_test.go and c_test.go)", res.Files)
	}
	want := filepath.Join("internal", "a", "a_test.go")
	if len(res.Violations) != 1 || res.Violations[0].File != want {
		t.Fatalf("ScanTree violations = %+v; want exactly one in %s", res.Violations, want)
	}
}
