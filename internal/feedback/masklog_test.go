package feedback

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// maskLogRoot returns a project root whose .moai marker exists, plus the
// Options that point a Scrub call at it.
func maskLogRoot(t *testing.T) (string, Options) {
	t.Helper()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("creating .moai marker: %v", err)
	}
	opt := testOptions()
	opt.ProjectRoot = root
	return root, opt
}

// AC-F-015 — the mask log records kind/where/count and a timestamp, carries no
// raw value, and is written 0o600.
//
// The "no raw value" half is the load-bearing one: the log is the only
// artefact that survives a scrub, so a value written here would turn the
// control into the leak path it exists to close (AP-6).
func TestMaskLogRecordsKindAndCountWithoutRawValue(t *testing.T) {
	t.Parallel()

	root, opt := maskLogRoot(t)
	token := fakeGitHubToken()

	res, err := Scrub(Input{Title: "report", Body: "the token is " + token + " ok"}, opt)
	if err != nil {
		t.Fatalf("Scrub returned error: %v", err)
	}
	if len(res.Findings) == 0 {
		t.Fatalf("expected at least one finding, got none")
	}

	path := MaskLogPathForRoot(root)
	raw, err := os.ReadFile(path) //nolint:gosec // test-owned temp path
	if err != nil {
		t.Fatalf("reading mask log %s: %v", path, err)
	}
	content := string(raw)

	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	last := lines[len(lines)-1]

	// Logged so the AC's evidence carries the actual entry. It is safe to
	// print for the same reason it is safe to write: the line holds no value.
	t.Logf("mask log entry: %s", last)

	stamp, _, ok := strings.Cut(last, " | ")
	if !ok {
		t.Fatalf("last line has no field separator: %q", last)
	}
	if _, err := time.Parse(time.RFC3339, strings.TrimSpace(stamp)); err != nil {
		t.Fatalf("last line does not start with an RFC3339 timestamp: %q (%v)", last, err)
	}
	if !strings.Contains(last, "kind="+KindSecret) {
		t.Errorf("last line records no kind: %q", last)
	}
	if !strings.Contains(last, "count=1") {
		t.Errorf("last line records no count: %q", last)
	}

	if strings.Contains(content, token) {
		t.Fatalf("mask log contains the raw token")
	}
	if tail := token[len(ghTokenPrefix):]; strings.Contains(content, tail[:12]) {
		t.Fatalf("mask log contains a fragment of the raw token")
	}

	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not POSIX on windows")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat mask log: %v", err)
	}
	if got := info.Mode().Perm(); got != maskLogPerm {
		t.Errorf("mask log perm = %O, want %O", got, maskLogPerm)
	}
}

// AC-F-016 — a log that cannot be written never fails the scrub.
//
// The unwritable condition is built by putting a FILE where the log directory
// must be, which fails MkdirAll on every platform.
func TestMaskLogFailureIsFailOpen(t *testing.T) {
	t.Parallel()

	root, opt := maskLogRoot(t)
	logsPath := filepath.Join(root, ".moai", "logs")
	if err := os.WriteFile(logsPath, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("occupying the logs path: %v", err)
	}

	// An env-value fixture rather than a credential one: this test asserts the
	// scrub COMPLETES, so the input must be one the classifier lets through.
	token := envSecretValue
	opt.Environ = func() []string { return []string{"GITHUB_TOKEN=" + token} }
	res, err := Scrub(Input{Body: "the token is " + token}, opt)
	if err != nil {
		t.Fatalf("Scrub returned error on unwritable log: %v", err)
	}
	if res.Verdict != VerdictOK {
		t.Errorf("verdict = %q, want %q", res.Verdict, VerdictOK)
	}
	if strings.Contains(res.Body, token) {
		t.Errorf("body was not masked")
	}
	if len(res.Findings) == 0 {
		t.Errorf("expected findings, got none")
	}
}

// A scrub that masks nothing writes no entry: one line per clean report would
// bury the entries that matter.
func TestMaskLogSkipsCleanScrub(t *testing.T) {
	t.Parallel()

	root, opt := maskLogRoot(t)
	if _, err := Scrub(Input{Body: "the init wizard asks one question too many"}, opt); err != nil {
		t.Fatalf("Scrub returned error: %v", err)
	}
	if _, err := os.Stat(MaskLogPathForRoot(root)); !os.IsNotExist(err) {
		t.Fatalf("expected no mask log, stat err = %v", err)
	}
}

// Without a project root there is nowhere to write, and the scrub still runs.
// This is the property that keeps a Scrub call in a test — or in a checkout
// that happens to sit under a project — off the real .moai tree.
func TestMaskLogRequiresProjectRoot(t *testing.T) {
	t.Parallel()

	res, err := Scrub(Input{Body: "the token is " + fakeGitHubToken()}, testOptions())
	if err != nil {
		t.Fatalf("Scrub returned error without a project root: %v", err)
	}
	if len(res.Findings) == 0 {
		t.Errorf("expected findings, got none")
	}
}
