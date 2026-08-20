package statusline

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The fetch tests. forge_test.go covers everything up to the exec boundary and
// stops there deliberately; these tests cross it with stub forge CLIs, so the
// values are proven to flow from the forge binary through forgeCount into the
// cache — without contacting github.com or gitlab.com at all. A test that
// shells out to the real `gh` would spend API quota per run and flake on rate
// limits (exactly the failure the feature must tolerate), so the only forge
// binaries on PATH are fakes that answer instantly.

// writeForgeStub writes a stub forge CLI named bin into a fresh temp dir and
// returns the dir. Each invocation appends its arguments to calls.log inside
// the dir, then answers in the shape that call expects:
//
//   - `repo view` — the totals call GitHub now uses: two lines, issues first;
//   - `api rate_limit` — a budget far above forgeRateFloor, so the guard lets
//     the fetch through unless a test overrides it (see writeRateLimitedStub);
//   - `pr`/`mr` and anything else — the one-integer enumerating answers the
//     GitLab path still uses.
//
// fail makes every call exit 1 — the shape a rate-limited or otherwise failing
// forge CLI takes.
func writeForgeStub(t *testing.T, bin string, issueCount, prCount int, fail bool) string {
	t.Helper()
	return writeForgeStubWithBudget(t, bin, issueCount, prCount, fail, 5000)
}

// writeForgeStubWithBudget is writeForgeStub with the remaining API budget the
// stub reports, so a test can drive the rate guard.
func writeForgeStubWithBudget(t *testing.T, bin string, issueCount, prCount int, fail bool, budget int) string {
	t.Helper()

	dir := t.TempDir()
	calls := filepath.Join(dir, "calls.log")
	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$*\" >> \"" + calls + "\"\n"
	if fail {
		script += "exit 1\n"
	}
	script += "case \"$1 $2\" in\n" +
		"  'api rate_limit') echo " + strconv.Itoa(budget) + " ; exit 0 ;;\n" +
		"  'repo view') echo " + strconv.Itoa(issueCount) + " ; echo " + strconv.Itoa(prCount) + " ; exit 0 ;;\n" +
		"esac\n"
	script += "case \"$1\" in pr|mr) echo " + strconv.Itoa(prCount) +
		" ;; *) echo " + strconv.Itoa(issueCount) + " ;; esac\n"

	path := filepath.Join(dir, bin)
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// stubCalls returns what the stub in dir logged, one line per invocation.
func stubCalls(t *testing.T, dir string) []string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(dir, "calls.log"))
	if err != nil {
		return nil
	}
	out := strings.FieldsFunc(string(data), func(r rune) bool { return r == '\n' })
	return out
}

// writeForgeOverride pins the forge by config so the tests need no git
// repository: forgeOverride alone decides once statusline.yaml names one.
func writeForgeOverride(t *testing.T, root, forge string) {
	t.Helper()

	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "statusline:\n  forge: " + forge + "\n"
	if err := os.WriteFile(filepath.Join(dir, "statusline.yaml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

// seedStaleCache writes a deliberately outdated cache so the tests can tell a
// freshly fetched value from one merely carried over.
func seedStaleCache(t *testing.T, root string, issues, prs int) {
	t.Helper()

	stale := GitHubCounts{
		OpenIssues: issues,
		OpenPRs:    prs,
		FetchedAt:  time.Now().Add(-time.Hour).Unix(),
	}
	if err := writeGitHubCache(root, stale); err != nil {
		t.Fatal(err)
	}
}

// TestRefreshGitHubCounts_FetchesValuesViaGitHubStub is the happy path the
// segment exists for: a fresh refresh asks the forge CLI for both counts and
// the cache afterwards serves exactly what it answered.
func TestRefreshGitHubCounts_FetchesValuesViaGitHubStub(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub forge CLIs are exercised on unix; windows covered by GOOS=windows build/vet")
	}

	root := t.TempDir()
	writeForgeOverride(t, root, "github")
	seedStaleCache(t, root, 7, 3)

	stubDir := writeForgeStub(t, "gh", 42, 17, false)
	t.Setenv("PATH", stubDir+":"+os.Getenv("PATH"))

	if err := RefreshGitHubCounts(context.Background(), root); err != nil {
		t.Fatalf("RefreshGitHubCounts: %v", err)
	}

	got := resolveGitHubCounts(root)
	if !got.Available {
		t.Fatal("cache unavailable after a successful refresh")
	}
	if got.OpenIssues != 42 || got.OpenPRs != 17 {
		t.Errorf("counts = %d/%d, want 42/17 — the stub's answers did not land", got.OpenIssues, got.OpenPRs)
	}
	if got.FetchedAt < time.Now().Add(-time.Minute).Unix() {
		t.Errorf("FetchedAt = %d, want a fresh timestamp", got.FetchedAt)
	}

	// GitHub answers both counts in ONE totals call, and must not enumerate:
	// the enumeration was measured at ten GraphQL points on a large repository
	// against one for the totals, and it reports its own --limit past the cap.
	calls := stubCalls(t, stubDir)
	var sawTotals, sawListing bool
	for _, c := range calls {
		if strings.HasPrefix(c, "repo view") {
			sawTotals = true
		}
		if strings.HasPrefix(c, "issue list") || strings.HasPrefix(c, "pr list") {
			sawListing = true
		}
	}
	if !sawTotals {
		t.Errorf("the totals call was never made, calls = %v", calls)
	}
	if sawListing {
		t.Errorf("the forge was enumerated despite having a totals call, calls = %v", calls)
	}
}

// TestRefreshGitHubCounts_SkipsFetchNearTheRateFloor pins the budget guard:
// with the forge reporting almost no API budget left, the refresh must not
// spend any, and the previous counts must survive.
func TestRefreshGitHubCounts_SkipsFetchNearTheRateFloor(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub forge CLIs are exercised on unix; windows covered by GOOS=windows build/vet")
	}

	root := t.TempDir()
	writeForgeOverride(t, root, "github")
	seedStaleCache(t, root, 11, 4)

	stubDir := writeForgeStubWithBudget(t, "gh", 42, 17, false, forgeRateFloor-1)
	t.Setenv("PATH", stubDir+":"+os.Getenv("PATH"))

	if err := RefreshGitHubCounts(context.Background(), root); err != nil {
		t.Fatalf("RefreshGitHubCounts: %v", err)
	}

	got := resolveGitHubCounts(root)
	if got.OpenIssues != 11 || got.OpenPRs != 4 {
		t.Errorf("counts = %d/%d, want the stale 11/4 — the guard let a fetch through", got.OpenIssues, got.OpenPRs)
	}
	for _, c := range stubCalls(t, stubDir) {
		if strings.HasPrefix(c, "repo view") {
			t.Errorf("a counts call was made below the rate floor: %q", c)
		}
	}
}

// TestRefreshGitHubCounts_UnreadableBudgetStillFetches is the other half of the
// guard's contract: an unanswerable question about the budget is not evidence
// the budget is gone. A forge whose rate query fails must still be counted,
// otherwise the guard silences the segment it exists to protect.
func TestRefreshGitHubCounts_UnreadableBudgetStillFetches(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub forge CLIs are exercised on unix; windows covered by GOOS=windows build/vet")
	}

	root := t.TempDir()
	writeForgeOverride(t, root, "github")
	seedStaleCache(t, root, 11, 4)

	// The stub answers the rate query with a non-number; everything else
	// behaves normally.
	dir := t.TempDir()
	calls := filepath.Join(dir, "calls.log")
	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$*\" >> \"" + calls + "\"\n" +
		"case \"$1 $2\" in\n" +
		"  'api rate_limit') echo 'not a number' ; exit 0 ;;\n" +
		"  'repo view') echo 42 ; echo 17 ; exit 0 ;;\n" +
		"esac\n" +
		"echo 0\n"
	if err := os.WriteFile(filepath.Join(dir, "gh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))

	if err := RefreshGitHubCounts(context.Background(), root); err != nil {
		t.Fatalf("RefreshGitHubCounts: %v", err)
	}
	if got := resolveGitHubCounts(root); got.OpenIssues != 42 || got.OpenPRs != 17 {
		t.Errorf("counts = %d/%d, want 42/17 — an unreadable budget blocked the fetch", got.OpenIssues, got.OpenPRs)
	}
}

// TestRefreshGitHubCounts_ShortTotalsAnswerKeepsStaleValues covers the parse
// boundary: a totals call that answers with fewer than two numbers is a failed
// fetch, not a zero. Treating it as a zero would overwrite good cached counts
// with a parse accident.
func TestRefreshGitHubCounts_ShortTotalsAnswerKeepsStaleValues(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub forge CLIs are exercised on unix; windows covered by GOOS=windows build/vet")
	}

	root := t.TempDir()
	writeForgeOverride(t, root, "github")
	seedStaleCache(t, root, 11, 4)

	dir := t.TempDir()
	script := "#!/bin/sh\n" +
		"case \"$1 $2\" in\n" +
		"  'api rate_limit') echo 5000 ; exit 0 ;;\n" +
		"  'repo view') echo 42 ; exit 0 ;;\n" + // one number where two are due
		"esac\n" +
		"echo 0\n"
	if err := os.WriteFile(filepath.Join(dir, "gh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))

	if err := RefreshGitHubCounts(context.Background(), root); err != nil {
		t.Fatalf("RefreshGitHubCounts: %v", err)
	}
	if got := resolveGitHubCounts(root); got.OpenIssues != 11 || got.OpenPRs != 4 {
		t.Errorf("counts = %d/%d, want the stale 11/4 preserved", got.OpenIssues, got.OpenPRs)
	}
}

// TestRefreshGitHubCounts_FetchesValuesViaGitLabStub runs the same happy path
// against glab, whose change requests are `mr` — the one noun that differs.
// A failing gh stub also sits on PATH: the gitlab override must mean gh is
// never consulted, and its untouched call log is the proof.
func TestRefreshGitHubCounts_FetchesValuesViaGitLabStub(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub forge CLIs are exercised on unix; windows covered by GOOS=windows build/vet")
	}

	root := t.TempDir()
	writeForgeOverride(t, root, "gitlab")
	seedStaleCache(t, root, 7, 3)

	glabDir := writeForgeStub(t, "glab", 5, 2, false)
	ghDir := writeForgeStub(t, "gh", 99, 99, true) // must never be asked
	t.Setenv("PATH", glabDir+":"+ghDir+":"+os.Getenv("PATH"))

	if err := RefreshGitHubCounts(context.Background(), root); err != nil {
		t.Fatalf("RefreshGitHubCounts: %v", err)
	}

	got := resolveGitHubCounts(root)
	if !got.Available || got.OpenIssues != 5 || got.OpenPRs != 2 {
		t.Errorf("counts = %+v, want 5/2 from the glab stub", got)
	}

	if calls := stubCalls(t, ghDir); len(calls) != 0 {
		t.Errorf("the github CLI was consulted under a gitlab override: %v", calls)
	}
	calls := stubCalls(t, glabDir)
	var sawIssue, sawMR bool
	for _, c := range calls {
		if strings.HasPrefix(c, "issue list") {
			sawIssue = true
		}
		if strings.HasPrefix(c, "mr list") {
			sawMR = true
		}
	}
	if !sawIssue || !sawMR {
		t.Errorf("glab was asked for issue and mr listings separately, calls = %v", calls)
	}
}

// TestRefreshGitHubCounts_FailingForgeKeepsStaleValues pins the rate-limit
// contract: when the forge CLI fails — quota exhausted, network gone, whatever
// shape the error takes — the refresh degrades freshness, not the numbers. The
// previous counts must survive so the segment keeps rendering them.
func TestRefreshGitHubCounts_FailingForgeKeepsStaleValues(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub forge CLIs are exercised on unix; windows covered by GOOS=windows build/vet")
	}

	root := t.TempDir()
	writeForgeOverride(t, root, "github")
	seedStaleCache(t, root, 11, 4)

	stubDir := writeForgeStub(t, "gh", 99, 99, true /* every call fails */)
	t.Setenv("PATH", stubDir+":"+os.Getenv("PATH"))

	// A failing forge is not an error to surface — the caller (the detached
	// child) logs nothing either way; the cache is the observable outcome.
	if err := RefreshGitHubCounts(context.Background(), root); err != nil {
		t.Fatalf("RefreshGitHubCounts: %v", err)
	}

	got := resolveGitHubCounts(root)
	if !got.Available {
		t.Fatal("cache unavailable after a failed refresh; stale values were dropped")
	}
	if got.OpenIssues != 11 || got.OpenPRs != 4 {
		t.Errorf("counts = %d/%d, want the stale 11/4 preserved", got.OpenIssues, got.OpenPRs)
	}
}
