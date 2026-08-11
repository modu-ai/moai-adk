package kanban

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

const fixtureHeadSHA = "0123456789abcdef0123456789abcdef01234567"

// writeRevision materializes a revision.json inside dir. A nil body writes no
// file at all, which is the absence case the predicate must reject.
func writeRevision(t *testing.T, dir string, body map[string]any) {
	t.Helper()
	encoded, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		t.Fatalf("encoding fixture revision: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "revision.json"), encoded, 0o600); err != nil {
		t.Fatalf("writing fixture revision: %v", err)
	}
}

// writeFindings materializes a findings.jsonl with the supplied raw content. A
// zero-length content string is a genuinely clean scan, not an aborted one.
func writeFindings(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "findings.jsonl"), []byte(content), 0o600); err != nil {
		t.Fatalf("writing fixture findings: %v", err)
	}
}

// matchingRevision is the single accepting shape: same commit, repo scope,
// working tree included.
func matchingRevision() map[string]any {
	return map[string]any{
		"scanned_commit":        fixtureHeadSHA,
		"effort_tier":           "high",
		"working_tree_included": true,
		"scope":                 "repo",
		"generated_at":          "2026-01-02T15:04:05Z",
	}
}

// completeFixture builds a results directory that satisfies every conjunct.
func completeFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeRevision(t, dir, matchingRevision())
	writeFindings(t, dir, "{\"severity\":\"low\"}\n")
	return dir
}

func TestRevisionMatchHappyPath(t *testing.T) {
	dir := completeFixture(t)

	if got := RevisionMatch(dir, fixtureHeadSHA, false); !got {
		t.Errorf("RevisionMatch on a complete matching fixture = false, want true")
	}
}

func TestRevisionMatchAbsentDirectoryIsFalse(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "never-created")

	if got := RevisionMatch(missing, fixtureHeadSHA, false); got {
		t.Errorf("RevisionMatch on an absent results directory = true, want false")
	}
}

func TestRevisionMatchAbsentRevisionJSONIsFalse(t *testing.T) {
	dir := t.TempDir()
	writeFindings(t, dir, "")

	if _, err := LoadRevision(filepath.Join(dir, "revision.json")); err == nil {
		t.Errorf("LoadRevision on an absent revision.json returned nil error, want an error")
	}
	if got := RevisionMatch(dir, fixtureHeadSHA, false); got {
		t.Errorf("RevisionMatch with no revision.json = true, want false")
	}
}

func TestLoadRevisionUnreadableFileIsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits do not gate reads on windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root: permission bits do not deny reads")
	}

	dir := t.TempDir()
	writeRevision(t, dir, matchingRevision())
	writeFindings(t, dir, "")

	path := filepath.Join(dir, "revision.json")
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatalf("chmod fixture revision: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })

	if _, err := LoadRevision(path); err == nil {
		t.Errorf("LoadRevision on an unreadable file returned nil error, want an error")
	}
	if got := RevisionMatch(dir, fixtureHeadSHA, false); got {
		t.Errorf("RevisionMatch with an unreadable revision.json = true, want false")
	}
}

func TestLoadRevisionMalformedJSONIsError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "revision.json"), []byte("{not json"), 0o600); err != nil {
		t.Fatalf("writing malformed fixture: %v", err)
	}
	writeFindings(t, dir, "")

	if _, err := LoadRevision(filepath.Join(dir, "revision.json")); err == nil {
		t.Errorf("LoadRevision on malformed JSON returned nil error, want an error")
	}
	if got := RevisionMatch(dir, fixtureHeadSHA, false); got {
		t.Errorf("RevisionMatch with malformed revision.json = true, want false")
	}
}

func TestRevisionMatchCommitMismatchIsFalse(t *testing.T) {
	dir := t.TempDir()
	body := matchingRevision()
	body["scanned_commit"] = "ffffffffffffffffffffffffffffffffffffffff"
	writeRevision(t, dir, body)
	writeFindings(t, dir, "")

	if got := RevisionMatch(dir, fixtureHeadSHA, false); got {
		t.Errorf("RevisionMatch with a scanned_commit mismatch = true, want false")
	}
}

func TestRevisionMatchNonRepoScopeIsFalse(t *testing.T) {
	dir := t.TempDir()
	body := matchingRevision()
	body["scope"] = "branch"
	writeRevision(t, dir, body)
	writeFindings(t, dir, "")

	if got := RevisionMatch(dir, fixtureHeadSHA, false); got {
		t.Errorf("RevisionMatch with scope=branch = true, want false")
	}
}

// AC-FM-019a: both halves in one test so the dirty-tree clause is provably the
// discriminating factor rather than an incidental pass.
func TestRevisionMatchDirtyTreeWithWorkingTreeExcluded(t *testing.T) {
	dir := t.TempDir()
	body := matchingRevision()
	body["working_tree_included"] = false
	writeRevision(t, dir, body)
	writeFindings(t, dir, "")

	if got := RevisionMatch(dir, fixtureHeadSHA, true); got {
		t.Errorf("RevisionMatch with a dirty tree and working_tree_included=false = true, want false")
	}
	if got := RevisionMatch(dir, fixtureHeadSHA, false); !got {
		t.Errorf("RevisionMatch with a clean tree and working_tree_included=false = false, want true")
	}
}

// AC-FM-019b: absence of findings.jsonl rejects; a zero-line findings.jsonl
// accepts, because a clean scan writes the file empty while an aborted scan
// characteristically never writes it at all.
func TestRevisionMatchFindingsCompleteness(t *testing.T) {
	dir := t.TempDir()
	writeRevision(t, dir, matchingRevision())

	if got := RevisionMatch(dir, fixtureHeadSHA, false); got {
		t.Errorf("RevisionMatch with no findings.jsonl = true, want false")
	}

	writeFindings(t, dir, "")
	if got := RevisionMatch(dir, fixtureHeadSHA, false); !got {
		t.Errorf("RevisionMatch with a zero-line findings.jsonl = false, want true")
	}
}

func TestRevisionMatchUnparseableFindingsLineIsFalse(t *testing.T) {
	dir := t.TempDir()
	writeRevision(t, dir, matchingRevision())
	writeFindings(t, dir, "{\"severity\":\"low\"}\nnot-json\n")

	if got := RevisionMatch(dir, fixtureHeadSHA, false); got {
		t.Errorf("RevisionMatch with an unparseable findings.jsonl line = true, want false")
	}
}

// Matches is the pure predicate the directory-level RevisionMatch composes. A
// nil revision is the absence case and must reject.
func TestMatchesNilRevisionIsFalse(t *testing.T) {
	if got := Matches(nil, fixtureHeadSHA, false); got {
		t.Errorf("Matches(nil, ...) = true, want false")
	}
}

// AC-FM-020c: the composed suppression decision is an allow-list over a
// *recorded* PRIMARY or FALLBACK. Row 1 is the positive control without which
// rows 2-5 would pass against a function that always returns false.
func TestSuppressStep0551RungAllowList(t *testing.T) {
	primary := RungPrimary
	fallback := RungFallback
	degraded := RungDegraded
	empty := Rung("")
	unknown := Rung("THOROUGH")

	cases := []struct {
		name           string
		rung           *Rung
		revisionMatch  bool
		wantSuppressed bool
	}{
		{"recorded PRIMARY suppresses", &primary, true, true},
		{"recorded FALLBACK suppresses", &fallback, true, true},
		{"recorded DEGRADED does not suppress", &degraded, true, false},
		{"never-recorded rung does not suppress", nil, true, false},
		{"recorded-empty rung does not suppress", &empty, true, false},
		{"unrecognized rung does not suppress", &unknown, true, false},
		{"failing predicate does not suppress even on PRIMARY", &primary, false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := SuppressStep0551(tc.rung, tc.revisionMatch); got != tc.wantSuppressed {
				t.Errorf("SuppressStep0551 = %v, want %v", got, tc.wantSuppressed)
			}
		})
	}
}
