package statusline

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// SPEC-STATUSLINE-OFFICIAL-USAGE-001 (REQ-SOU-001, REQ-SOU-005): structural
// removal guards. The usage collector's three removed surfaces — the Haiku
// probe and OAuth usage endpoint (api.anthropic.com), the OAuth beta header,
// and the credential reads (macOS keychain, ~/.claude/.credentials.json,
// ~/.claude/credentials.json) — must never reappear in a production file of
// this package, and the retired identifiers must stay deleted. The path
// markers close the renamed-symbol mutant of the plan-audit (acceptance
// §E.1): a credential read reintroduced under a new name still carries the
// path string in source.

// removedPathMarkers are the removed-path strings the guard fails on. The
// credentials entry is a regex covering both credential files
// (~/.claude/.credentials.json and ~/.claude/credentials.json).
var removedPathMarkers = []struct {
	name string
	re   *regexp.Regexp
}{
	{"api.anthropic.com", regexp.MustCompile(`api\.anthropic\.com`)},
	{"oauth-2025-04-20", regexp.MustCompile(`oauth-2025-04-20`)},
	{"find-generic-password", regexp.MustCompile(`find-generic-password`)},
	{"Claude Code-credentials", regexp.MustCompile(`Claude Code-credentials`)},
	{"haiku-4-5", regexp.MustCompile(`haiku-4-5`)},
	{"credentials.json (either credential file)", regexp.MustCompile(`credentials\.json`)},
}

// retiredUsageSymbols are the deleted collector identifiers; a same-name
// reintroduction fails here, a renamed one is caught by the path markers
// above.
var retiredUsageSymbols = []struct {
	name string
	re   *regexp.Regexp
}{
	{"UsageProvider", regexp.MustCompile(`\bUsageProvider\b`)},
	{"NewUsageCollector", regexp.MustCompile(`\bNewUsageCollector\b`)},
	{"usageCollector", regexp.MustCompile(`\busageCollector\b`)},
	{"UsageResult", regexp.MustCompile(`\bUsageResult\b`)},
	{"UsageData", regexp.MustCompile(`\bUsageData\b`)},
	{"usageCacheFile", regexp.MustCompile(`\busageCacheFile\b`)},
	{"readOAuthToken", regexp.MustCompile(`\breadOAuthToken\b`)},
	{"readTokenFromKeychain", regexp.MustCompile(`\breadTokenFromKeychain\b`)},
}

// productionStatuslineFiles returns every non-test .go file under this
// package directory. Test files are excluded: they legitimately name the
// retired identifiers in guard tests and fixtures.
func productionStatuslineFiles(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read statusline package dir: %v", err)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		files = append(files, filepath.Join(".", e.Name()))
	}
	if len(files) == 0 {
		t.Fatal("no production .go files found in statusline package — the guard scanned nothing")
	}
	return files
}

// guardScan scans the production files for one marker list and returns the
// offending "file:line" hits.
func guardScan(t *testing.T, markers []struct {
	name string
	re   *regexp.Regexp
}) []string {
	t.Helper()
	var hits []string
	for _, f := range productionStatuslineFiles(t) {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		for i, line := range strings.Split(string(data), "\n") {
			for _, m := range markers {
				if m.re.MatchString(line) {
					hits = append(hits, f+":"+itoaGuard(i+1)+" matches "+m.name)
				}
			}
		}
	}
	return hits
}

// itoaGuard is a tiny local itoa so the guard test carries no strconv
// dependency beyond the standard library it already needs.
func itoaGuard(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// TestRemovedCollectionSurfacesAbsent fails the suite when any removed-path
// marker reappears in a production file under internal/statusline
// (SPEC-STATUSLINE-OFFICIAL-USAGE-001 AC-SOU-002, release-blocking).
func TestRemovedCollectionSurfacesAbsent(t *testing.T) {
	if hits := guardScan(t, removedPathMarkers); len(hits) > 0 {
		t.Errorf("removed usage-collection surface(s) reappeared in production files:\n%s",
			strings.Join(hits, "\n"))
	}
}

// TestRetiredUsageSymbolsAbsent fails the suite when any retired collector
// identifier reappears in a production file under internal/statusline
// (SPEC-STATUSLINE-OFFICIAL-USAGE-001 AC-SOU-006).
func TestRetiredUsageSymbolsAbsent(t *testing.T) {
	if hits := guardScan(t, retiredUsageSymbols); len(hits) > 0 {
		t.Errorf("retired usage symbol(s) reappeared in production files:\n%s",
			strings.Join(hits, "\n"))
	}
}
