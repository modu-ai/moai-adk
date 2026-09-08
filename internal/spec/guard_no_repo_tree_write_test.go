package spec

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// TestNoTestWritesRepoTree — SPEC-HARNESS-EVIDENCE-WRITE-001 REQ-006 guard.
//
// Invariant: by default, no test in this package writes into the repository
// working tree. Run output belongs in a per-run t.TempDir() directory; reading
// tracked evidence under .moai/reports/** is legitimate (evidence-as-input)
// and is NOT flagged — the discriminant is the write primitive, not the path
// string alone.
//
// Predicate (mechanical, per file, single pass, no test execution):
//
//  1. A repo-anchored literal is a quoted string whose content contains
//     ".moai/" — i.e. the string itself spells a repo-relative path. A bare
//     ".moai" SEGMENT (as in filepath.Join(tempDir, ".moai", "logs"), a
//     temp-fixture layout) does not match, so temp-scoped fixtures stay green.
//  2. An identifier is flagged when it is initialized from an expression
//     containing a repo-anchored literal or a previously flagged identifier
//     (one pass, same file). This chains const -> filepath.Join alias ->
//     write call, the shape of the original defect (130846ab2).
//  3. A finding is a line invoking a file-write primitive (os.WriteFile,
//     os.MkdirAll, os.Create, os.OpenFile, ioutil.WriteFile) that mentions a
//     flagged identifier or carries a repo-anchored literal directly.
//
// Boundary — what this guard does NOT catch (documented per SPEC §C):
//   - dynamically built paths: bare ".moai" segments joined under a repo root
//     (filepath.Join(findRepoRoot(t), ".moai", "reports")), fmt.Sprintf, or
//     string concatenation of fragments;
//   - a repo-anchored path carried through a helper's PARAMETERS (the scan is
//     scope-blind: a flagged value passed as an argument is not re-flagged
//     inside the callee's body);
//   - multi-line write statements whose argument lands on a continuation line;
//   - this file's own source (its pattern literals would self-match; it
//     declares no write primitives);
//   - non-test sources (scope is *_test.go in this package only).
func TestNoTestWritesRepoTree(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed; cannot locate package directory")
	}
	pkgDir := filepath.Dir(thisFile)
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		t.Fatalf("read %s: %v", pkgDir, err)
	}
	thisBase := filepath.Base(thisFile)
	var findings []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, "_test.go") || name == thisBase {
			continue
		}
		findings = append(findings, scanTestSourceForRepoWrites(filepath.Join(pkgDir, name))...)
	}
	for _, f := range findings {
		t.Errorf("%s", f)
	}
}

// guardWritePrimitives are the textual markers of file-writing primitives.
var guardWritePrimitives = []string{
	"os.WriteFile(",
	"os.MkdirAll(",
	"os.Create(",
	"os.OpenFile(",
	"ioutil.WriteFile(",
}

// guardAnchoredLiteral matches a quoted string that spells a repo-relative
// path through .moai (the slash inside the literal is the discriminator
// against a bare ".moai" path segment under a temp fixture root).
var guardAnchoredLiteral = regexp.MustCompile(`"[^"]*\.moai/[^"]*"`)

// guardAssign captures `name := rhs`, `name = rhs`, and `const|var name = rhs`
// single-name assignments at statement start. Composite forms (multi-name
// `a, b := ...`, struct-literal keys, map entries) are deliberately not
// tracked — see the boundary list above.
var guardAssign = regexp.MustCompile(`^(?:(?:const|var)\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*(?::=|=)\s*(.+)$`)

// stripLineComment removes a trailing // comment, honoring string literals.
func stripLineComment(line string) string {
	inStr := false
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case '"':
			inStr = !inStr
		case '/':
			if !inStr && i+1 < len(line) && line[i+1] == '/' {
				return line[:i]
			}
		}
	}
	return line
}

// scanTestSourceForRepoWrites applies the REQ-006 predicate to one test file
// and returns human-readable findings (empty when the file is clean).
func scanTestSourceForRepoWrites(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return []string{fmt.Sprintf("%s: open: %v", filepath.Base(path), err)}
	}
	defer func() { _ = f.Close() }()

	base := filepath.Base(path)
	var findings []string
	flagged := map[string]bool{} // identifiers bound to repo-anchored paths
	sc := bufio.NewScanner(f)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		code := stripLineComment(sc.Text())

		// Finding: write primitive reaching a repo-anchored target.
		if prim := guardWritePrimitiveIn(code); prim != "" {
			if guardAnchoredLiteral.MatchString(code) {
				findings = append(findings, fmt.Sprintf("%s:%d: %s writes a repo-anchored path (direct .moai/ literal)", base, lineNo, prim))
			} else if ident, ok := guardFlaggedIdentIn(code, flagged); ok {
				findings = append(findings, fmt.Sprintf("%s:%d: %s writes via %q (repo-anchored)", base, lineNo, prim, ident))
			}
		}

		// Alias tracking: flag identifiers initialized from anchored paths.
		if groups := guardAssign.FindStringSubmatch(strings.TrimSpace(code)); groups != nil {
			name, rhs := groups[1], groups[2]
			if guardAnchoredLiteral.MatchString(rhs) || guardHasFlaggedIdent(rhs, flagged) {
				flagged[name] = true
			}
		}
	}
	return findings
}

// guardWritePrimitiveIn returns the first file-write primitive marker on the
// line, or "".
func guardWritePrimitiveIn(code string) string {
	for _, p := range guardWritePrimitives {
		if strings.Contains(code, p) {
			return strings.TrimSuffix(p, "(")
		}
	}
	return ""
}

// guardHasFlaggedIdent reports whether code references any flagged identifier
// as a whole word.
func guardHasFlaggedIdent(code string, flagged map[string]bool) bool {
	_, ok := guardFlaggedIdentIn(code, flagged)
	return ok
}

// guardFlaggedIdentIn returns one flagged identifier referenced by code.
func guardFlaggedIdentIn(code string, flagged map[string]bool) (string, bool) {
	for ident := range flagged {
		re, err := regexp.Compile(`\b` + regexp.QuoteMeta(ident) + `\b`)
		if err != nil {
			continue
		}
		if re.MatchString(code) {
			return ident, true
		}
	}
	return "", false
}
