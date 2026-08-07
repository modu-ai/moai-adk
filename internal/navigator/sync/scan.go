package sync

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"log/slog"
)

// logPathName is the relative path (under the project root) of the fail-open
// advisory log shared by the sync-package scanners and the capability gate.
const logPathName = ".moai/logs/navigator-sync.log"

// decIDRegex is the `<id>` grammar for `@NAV:DEC-<id>` tokens (REQ-NS-003):
// uppercase ASCII + digits + internal hyphens, anchored at the head by the
// `@NAV:DEC-` prefix and at the tail by the grammar-class terminator (any
// non-class rune, or end-of-input).
var decIDRegex = regexp.MustCompile(`@NAV:DEC-([A-Z][A-Z0-9-]*)`)

// symIDRegex is the `<symbol>` grammar for `@NAV:SYM:<symbol>` tokens
// (REQ-NS-004): identifier-shaped, language-neutral.
var symIDRegex = regexp.MustCompile(`@NAV:SYM:([A-Za-z_][A-Za-z0-9_.]*)`)

// decRawRegex matches the literal `@NAV:DEC-` prefix (any trailing bytes) so
// the scanner can detect malformed occurrences (`@NAV:DEC-` with empty id)
// that the strict grammar decIDRegex cannot match.
var decRawRegex = regexp.MustCompile(`@NAV:DEC-`)

// symRawRegex matches the literal `@NAV:SYM:` prefix (any trailing bytes) so
// the scanner can detect malformed occurrences (`@NAV:SYM:` with empty
// symbol).
var symRawRegex = regexp.MustCompile(`@NAV:SYM:`)

// designDocFiles is the fixed design-doc root (D3): three project-scope docs.
var designDocFiles = []string{
	".moai/project/product.md",
	".moai/project/structure.md",
	".moai/project/tech.md",
}

// designDocGlob is the recursive glob for `.moai/docs/**/*.md` (D3).
const designDocGlob = ".moai/docs"

// codeScanGlobs are the source-tree scan roots for `@NAV:SYM` code-side
// scanning (REQ-NS-004): walks Go source under the project root, excluding
// *_test.go and vendored paths.
var codeScanGlobs = []string{"."}

// ScanDec scans the design-doc root for `@NAV:DEC-<id>` occurrences (REQ-NS-003)
// and emits one BindingRecord per occurrence. The commitSHA stamps every
// record (the git baseline provenance per REQ-NS-002). Malformed occurrences
// (empty id) are skipped with a diagnostic written to logPath (REQ-NS-017);
// the scan never aborts on a malformed token.
//
// Fail-open: a missing design-doc file or an unreadable path is logged at
// debug and skipped — the scan returns whatever records it collected.
func ScanDec(projectRoot, commitSHA, logPath string) ([]BindingRecord, []string, error) {
	paths := designDocPaths(projectRoot)
	return scanToken(projectRoot, commitSHA, logPath, paths, decIDRegex, decRawRegex, FamilyNavDec, "DEC")
}

// ScanSym scans Go source (`*.go` excluding `*_test.go` + vendor) and design
// docs (`.moai/project/**/*.md` + `.moai/docs/**/*.md`) for
// `@NAV:SYM:<symbol>` occurrences (REQ-NS-004). Emits one BindingRecord per
// occurrence.
func ScanSym(projectRoot, commitSHA, logPath string) ([]BindingRecord, []string, error) {
	paths := symCodePaths(projectRoot)
	recs1, diags1, err := scanToken(projectRoot, commitSHA, logPath, paths, symIDRegex, symRawRegex, FamilyNavSym, "SYM")
	if err != nil {
		return recs1, diags1, err
	}
	mdPaths := designDocPaths(projectRoot)
	recs2, diags2, err := scanToken(projectRoot, commitSHA, logPath, mdPaths, symIDRegex, symRawRegex, FamilyNavSym, "SYM")
	if err != nil {
		return recs1, append(diags1, diags2...), nil
	}
	all := append(recs1, recs2...)
	sortRecords(all)
	return all, append(diags1, diags2...), nil
}

// scanToken is the shared per-token scanner: walks the given files, applies
// re per line, and emits one BindingRecord per match. The rawRe detects
// malformed occurrences (literal prefix present but the strict grammar does
// not match); each malformed occurrence is skipped with a diagnostic
// (REQ-NS-017).
func scanToken(projectRoot, commitSHA, logPath string, files []string, re, rawRe *regexp.Regexp, family TokenFamily, kind string) ([]BindingRecord, []string, error) {
	var records []BindingRecord
	var diagnostics []string
	for _, p := range files {
		// Skip test fixtures and vendored paths even when the caller's root
		// glob picks them up. (The DEC scanner has no excluded paths today;
		// the SYM scanner pre-filters via symCodePaths. This guard defends
		// the doc tree where a stray `vendor` directory might exist.)
		if shouldSkipScanPath(p) {
			continue
		}
		content, err := os.ReadFile(p)
		if err != nil {
			slog.Debug("sync: scan skip", "path", p, "error", err)
			continue
		}
		lineNum := 0
		scanner := bufio.NewScanner(strings.NewReader(string(content)))
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			// Strict-grammar matches produce binding records.
			matches := re.FindAllStringSubmatchIndex(line, -1)
			matchedRanges := make([][2]int, 0, len(matches))
			for _, m := range matches {
				id := line[m[2]:m[3]]
				matchedRanges = append(matchedRanges, [2]int{m[0], m[1]})
				records = append(records, BindingRecord{
					TokenFamily: family,
					Identifier:  id,
					SourcePath:  p,
					LineNumber:  lineNum,
					CommitSHA:   commitSHA,
				})
			}
			// Raw-prefix occurrences that are NOT covered by any strict match
			// are malformed tokens → diagnostic + skip (REQ-NS-017).
			for _, rawIdx := range rawRe.FindAllStringIndex(line, -1) {
				covered := false
				for _, mr := range matchedRanges {
					if rawIdx[0] >= mr[0] && rawIdx[1] <= mr[1] {
						covered = true
						break
					}
				}
				if covered {
					continue
				}
				diagnostics = append(diagnostics,
					fmt.Sprintf("malformed @NAV:%s token at %s:%d (empty or invalid identifier); skipped", kind, p, lineNum))
				appendLog(logPath,
					fmt.Sprintf("navigator-sync: malformed @NAV:%s token at %s:%d (empty or invalid identifier); skipped", kind, relOrRoot(projectRoot, p), lineNum))
			}
		}
		if err := scanner.Err(); err != nil {
			slog.Debug("sync: scan read error", "path", p, "error", err)
		}
	}
	sortRecords(records)
	return records, diagnostics, nil
}

// designDocPaths returns the absolute, sorted list of design-doc paths to
// scan: the three fixed project docs (D3) plus `.moai/docs/**/*.md`.
func designDocPaths(projectRoot string) []string {
	var out []string
	for _, rel := range designDocFiles {
		out = append(out, filepath.Join(projectRoot, filepath.FromSlash(rel)))
	}
	docsRoot := filepath.Join(projectRoot, filepath.FromSlash(designDocGlob))
	if _, err := os.Stat(docsRoot); err == nil {
		_ = filepath.WalkDir(docsRoot, func(path string, d os.DirEntry, werr error) error {
			if werr != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".md") {
				return nil
			}
			out = append(out, path)
			return nil
		})
	}
	sort.Strings(out)
	return out
}

// symCodePaths returns the absolute, sorted list of Go source files under the
// project root, excluding *_test.go and vendored paths (REQ-NS-004).
func symCodePaths(projectRoot string) []string {
	var out []string
	for _, root := range codeScanGlobs {
		base := filepath.Join(projectRoot, root)
		_ = filepath.WalkDir(base, func(path string, d os.DirEntry, werr error) error {
			if werr != nil {
				return nil
			}
			if d.IsDir() {
				name := d.Name()
				if name == "vendor" || name == "node_modules" || name == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			if strings.HasSuffix(path, "_test.go") {
				return nil
			}
			// Pre-filter: only files containing the literal `@NAV:SYM:`
			// substring (acceptance §D edge case for large monorepos).
			if !fileContainsNeedle(path, "@NAV:SYM:") {
				return nil
			}
			out = append(out, path)
			return nil
		})
	}
	sort.Strings(out)
	return out
}

// fileContainsNeedle reports whether path's bytes include needle, fail-open
// false on read error.
func fileContainsNeedle(path, needle string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(b), needle)
}

// shouldSkipScanPath defends against scanning test fixtures or vendored paths
// that the design-doc walk might pick up.
func shouldSkipScanPath(path string) bool {
	clean := filepath.ToSlash(path)
	if strings.Contains(clean, "/vendor/") {
		return true
	}
	if strings.HasSuffix(clean, "_test.go") {
		return true
	}
	return false
}

// sortRecords orders records deterministically by
// (token_family, identifier, source_path, line_number) — required for
// byte-identical re-runs (REQ-NS-009).
func sortRecords(r []BindingRecord) {
	sort.SliceStable(r, func(i, j int) bool {
		if r[i].TokenFamily != r[j].TokenFamily {
			return r[i].TokenFamily < r[j].TokenFamily
		}
		if r[i].Identifier != r[j].Identifier {
			return r[i].Identifier < r[j].Identifier
		}
		if r[i].SourcePath != r[j].SourcePath {
			return r[i].SourcePath < r[j].SourcePath
		}
		return r[i].LineNumber < r[j].LineNumber
	})
}
