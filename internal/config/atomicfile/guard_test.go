package atomicfile

// AC-CAW-007 — Regression guard.
//
// This test mechanically asserts that no non-test source file under
// internal/config/** or internal/cli/** writes to a config/settings-persistence
// path via a bare os.WriteFile. After the M1-M3 work every config/settings
// writer routes through atomicfile.Write (the single choke point); this guard
// FAILS if a future contributor reintroduces a bare os.WriteFile whose target
// resolves to a config-set path (.moai/config/**) or a settings-set path
// (settings.json / settings.json.tmpl / settings.local.json).
//
// Mechanism (Option A per the M4 design — regex + literal/variable resolution):
//
//  1. Walk internal/config/ and internal/cli/ for non-test *.go files.
//  2. Locate every os.WriteFile( call and capture its file:line.
//  3. Extract the first argument (the target path).
//  4. Resolve the target to a string literal:
//     - direct string literal -> use verbatim;
//     - identifier -> scan the same file for `<id> := "..."`, `var <id> = "..."`,
//       or `<id> = "..."` and use the assigned literal;
//     - anything else (filepath.Join(...), function call, unresolvable) -> skip
//       (conservative: do not false-positive; such targets are not literals and
//       the regression shape we guard against is a literal/near-literal config
//       path, which is what a reintroduced bare write looks like).
//  5. If the resolved literal matches a config/settings path pattern -> record
//     a violation naming file:line with the remediation.
//
// Why not a full AST (Option B): the guard's job is regression detection, not
// perfect intent classification. Literal/near-literal detection catches the
// realistic regression (a contributor copies an old config write back in). The
// complexity of full data-flow analysis is not earned here (Enforce Simplicity).
//
// Exempt allowlist: the sites below were reviewed during M3 and are legitimately
// bare os.WriteFile calls whose targets are NOT config/settings paths. They pass
// the predicate naturally (their targets are backup-dir artifacts, probe files,
// lockfiles, or a seam var declaration), so this allowlist is documentation +
// belt-and-suspenders; the predicate is the load-bearing classifier.
//
//	- internal/cli/update_recovery_manifest.go:78   -> backupDir/<manifest>
//	- internal/cli/update_namespace_protect.go:242  -> backupDir/.complete
//	- internal/cli/update_deny_migration.go:100     -> deny-file (mode var)
//	- internal/cli/update_cleanup.go:391            -> case-sensitivity probe
//	- internal/cli/update_cleanup.go:70             -> O_EXCL lockfile (OpenFile)
//	- internal/cli/update_disk_backup.go:39         -> diskBackupWriteFile var

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// TestNoBareOSWriteFileToConfigPaths is the AC-CAW-007 regression guard.
func TestNoBareOSWriteFileToConfigPaths(t *testing.T) {
	t.Parallel()

	repoRoot := findRepoRoot(t)

	// Trees the guard covers: every config/settings writer lives in one of these.
	scanRoots := []string{
		filepath.Join(repoRoot, "internal", "config"),
		filepath.Join(repoRoot, "internal", "cli"),
	}

	var violations []string
	for _, root := range scanRoots {
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			violations = append(violations, scanFileForConfigPathWrites(path, repoRoot)...)
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}

	if len(violations) > 0 {
		t.Fatalf("AC-CAW-007 violation: bare os.WriteFile to a config/settings path detected.\n"+
			"Every config/settings write MUST route through atomicfile.Write (the single\n"+
			"atomic, mode-preserving choke point). Remediate each site below:\n\n%s",
			strings.Join(violations, "\n"))
	}
}

// configPathIndicators are the path substrings/basenames that mark a target as a
// config-set or settings-set persistence path. A bare os.WriteFile whose target
// matches any of these is the regression this guard rejects.
var configPathIndicators = []string{
	".moai/config/",
	"settings.json",
	"settings.json.tmpl",
	"settings.local.json",
}

// writeFileCallRe matches an os.WriteFile( call and captures the raw first
// argument text (up to the first comma at the top nesting level). It is
// intentionally permissive about whitespace.
var writeFileCallRe = regexp.MustCompile(`os\.WriteFile\(\s*([^,]+),`)

// identAssignRe resolves a bare identifier argument to a same-file string-literal
// assignment. It matches `<id> := "<lit>"`, `var <id> = "<lit>"`, `<id> = "<lit>"`.
// The captured group is the literal (without quotes).
func literalForIdent(fileBody, ident string) (string, bool) {
	ident = strings.TrimSpace(ident)
	// Quote-escape the identifier for the regex.
	pat := regexp.MustCompile(`(?:var\s+|)\b` + regexp.QuoteMeta(ident) + `\s*(?::=|=)\s*"((?:[^"\\]|\\.)*)"`)
	m := pat.FindStringSubmatch(fileBody)
	if m != nil {
		return m[1], true
	}
	return "", false
}

// scanFileForConfigPathWrites returns one violation string per os.WriteFile call
// in the file whose target resolves to a config/settings path. path is the
// absolute file path; repoRoot is used to render a project-relative site label.
func scanFileForConfigPathWrites(path, repoRoot string) []string {
	body, err := os.ReadFile(path)
	if err != nil {
		// Unreadable file is not a guard concern; skip silently.
		return nil
	}
	src := string(body)

	rel, relErr := filepath.Rel(repoRoot, path)
	if relErr != nil {
		rel = path
	}

	// Build a line-indexed view so each match can be attributed to its line.
	lines := strings.Split(src, "\n")

	var out []string
	for _, m := range writeFileCallRe.FindAllStringSubmatchIndex(src, -1) {
		argRaw := strings.TrimSpace(src[m[2]:m[3]])
		lineNo := lineOf(lines, m[0])

		resolved, ok := resolveArg(src, argRaw)
		if !ok {
			continue // unresolvable (filepath.Join, call, etc.) -> conservative skip
		}
		if !isConfigSettingsPath(resolved) {
			continue
		}
		out = append(out, formatViolation(rel, lineNo, argRaw, resolved))
	}
	return out
}

// resolveArg resolves a raw first-argument token to a string literal value. It
// handles direct literals and same-file identifier assignments; everything else
// returns ok=false (conservative skip).
func resolveArg(fileBody, argRaw string) (string, bool) {
	if len(argRaw) >= 2 && argRaw[0] == '"' {
		// Direct string literal. Strip the quotes and unquote via Go rules.
		return unquoteDoubleQuoted(argRaw)
	}
	// Identifier (possibly a simple selector like pkg.X is NOT handled — those
	// are not local literals). Only bare identifiers resolve here.
	if isIdentifier(argRaw) {
		if lit, ok := literalForIdent(fileBody, argRaw); ok {
			return lit, true
		}
	}
	return "", false
}

var identRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func isIdentifier(s string) bool { return identRe.MatchString(s) }

// unquoteDoubleQuoted strips the surrounding quotes from a Go double-quoted
// string literal and applies the common escape unescapes. It is intentionally
// simple; config paths do not contain exotic escapes.
func unquoteDoubleQuoted(s string) (string, bool) {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return "", false
	}
	inner := s[1 : len(s)-1]
	inner = strings.NewReplacer(
		`\"`, `"`,
		`\\`, `\`,
		`\n`, "\n",
		`\t`, "\t",
	).Replace(inner)
	return inner, true
}

// isConfigSettingsPath reports whether a resolved path targets a config-set or
// settings-set persistence path.
func isConfigSettingsPath(resolved string) bool {
	for _, ind := range configPathIndicators {
		if strings.Contains(resolved, ind) {
			return true
		}
	}
	// A resolved path that is exactly a settings basename also counts.
	base := filepath.Base(resolved)
	switch base {
	case "settings.json", "settings.json.tmpl", "settings.local.json":
		return true
	}
	return false
}

func formatViolation(rel string, lineNo int, argRaw, resolved string) string {
	return "  - " + rel + ":" + itoa(lineNo) + " — os.WriteFile(" + argRaw + ", ...)" +
		" resolves to config/settings path " + quote(resolved) +
		"; route through atomicfile.Write instead"
}

// lineOf returns the 1-based line number of a byte offset in the pre-split lines.
func lineOf(lines []string, byteOffset int) int {
	consumed := 0
	for i, ln := range lines {
		if consumed+len(ln) >= byteOffset {
			return i + 1
		}
		// +1 for the '\n' that Split removed.
		consumed += len(ln) + 1
	}
	return len(lines)
}

// findRepoRoot locates the repository root by walking up from this test file's
// directory until both internal/config and internal/cli exist as siblings.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed — cannot locate test file")
	}
	dir := filepath.Dir(thisFile)
	for i := 0; i < 10; i++ {
		if dirIs(filepath.Join(dir, "internal", "config")) && dirIs(filepath.Join(dir, "internal", "cli")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("could not locate repo root containing internal/config and internal/cli from " + thisFile)
	return ""
}

func dirIs(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

// itoa/quote avoid importing strconv only for two trivial helpers used in
// violation formatting, keeping the guard dependency-free beyond stdlib.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func quote(s string) string { return "\"" + s + "\"" }
