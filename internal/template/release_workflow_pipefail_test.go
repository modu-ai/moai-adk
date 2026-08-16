// release_workflow_pipefail_test.go: CI guard against the `grep -q` SIGPIPE
// hazard in .github/workflows/release.yml.
//
// The release provenance gate runs its seven checks in one `set -euo pipefail`
// step. Under pipefail a pipeline reports the rightmost non-zero exit, and
// `grep -q` exits the moment it matches — closing the pipe while the upstream
// producer is still writing. The producer then dies of SIGPIPE (141), so the
// pipeline reports 141 even though the match SUCCEEDED, and the surrounding
// `if ! <pipeline>; then fail ...` fires on a passing check.
//
// What decides whether it fires is NOT the total input size — it is how much
// the producer still has left to write at the moment grep exits. If that
// remainder fits the ~64KB pipe buffer the producer completes and exits 0; if
// it does not, the producer blocks on a pipe that is about to close and dies.
//
// Measured on the v3.1.0 release, which is where this was found:
//
//	check 5  CHANGELOG.md, 416,407 B, match at line 8      -> ~416KB left  -> 141
//	check 2  tag annotation, 271,034 B, match at line 288  -> 82 B left    -> 0
//	check 6  system.yaml, 1,894 B                          -> < 1 KB left  -> 0
//
// So check 2 survives a 271KB input purely because its provenance trailer sits
// three lines from the end. That is a property of the data, not of the code:
// move the trailer earlier — or let anything grow below the match — and the
// same line fails the same way. Checks 2 and 6 are fixed here for that reason,
// not because they were observed failing.
//
// Sentinel on failure: PIPEFAIL_SIGPIPE_HAZARD
//
// Design notes:
//   - TestReleaseWorkflowNoQuietGrepUnderPipefail is the enforcing guard: it
//     reads the workflow and rejects any `| grep -q`. It fails on the
//     pre-fix file and passes on the fixed one.
//   - TestPipefailSigpipeMechanism is the evidence behind that guard. It
//     demonstrates the failure and its remedy on a synthetic pipeline, so the
//     rule above is not taken on trust. It passes both before and after the
//     workflow fix because it tests the shell mechanism, not the file.
//   - A `grep -q` that is NOT fed by a pipe (a file argument, a here-string)
//     has no upstream producer to kill and is therefore safe. The guard
//     anchors on the pipe for that reason.
//   - This lives under internal/template/ (NOT internal/template/templates/)
//     alongside branch_protection_parity_test.go, the existing precedent for
//     asserting on .github/ content, and reuses its
//     findProjectRootForMirrorTest helper. _test.go files are not
//     user-distributed template content, so the template-neutrality CI guard
//     does not apply.
package template_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// releaseWorkflowRelPath is the repo-relative path of the release workflow
// whose provenance gate carries the seven checks.
const releaseWorkflowRelPath = ".github/workflows/release.yml"

// pipeIntoGrep finds a pipe feeding a grep and captures everything after it.
// The pipe is the load-bearing part: without an upstream producer there is
// nothing for grep's early exit to kill, so `grep -q file` is safe.
var pipeIntoGrep = regexp.MustCompile(`\|\s*grep\s+(\S.*)$`)

// shortOptsTakingValue are grep's single-letter options whose value follows,
// either glued to the option or as the next token. Anything after one of them
// in the same cluster is that value, not another option.
const shortOptsTakingValue = "efmABCD"

// longOptsTakingValue are the --name forms whose value follows as a separate
// token when not given as --name=value.
var longOptsTakingValue = map[string]bool{
	"regexp": true, "file": true, "max-count": true,
	"after-context": true, "before-context": true, "context": true,
}

// isQuietGrep reports whether a grep invocation suppresses output, which is
// what makes it exit early and kill its producer. It walks the option tokens
// rather than pattern-matching the line, because quiet has three spellings
// (`-q`, a cluster containing q, `--quiet`/`--silent`) and because an option's
// own value must not be read as more options — `grep -e -q` searches for the
// literal string "-q" and is NOT quiet.
func isQuietGrep(args []string) bool {
	for i := 0; i < len(args); i++ {
		tok := args[i]

		switch {
		case tok == "--":
			return false // everything after this is the pattern
		case strings.HasPrefix(tok, "--"):
			name := tok[2:]
			if eq := strings.IndexByte(name, '='); eq >= 0 {
				name = name[:eq]
			} else if longOptsTakingValue[name] {
				i++ // its value is the next token
			}
			if name == "quiet" || name == "silent" {
				return true
			}
		case strings.HasPrefix(tok, "-") && len(tok) > 1:
			for j := 1; j < len(tok); j++ {
				if tok[j] == 'q' {
					return true
				}
				if strings.IndexByte(shortOptsTakingValue, tok[j]) >= 0 {
					if j == len(tok)-1 {
						i++ // the value is the next token
					}
					break // the rest of this cluster is the value
				}
			}
		default:
			return false // the pattern; options are done
		}
	}
	return false
}

func TestReleaseWorkflowNoQuietGrepUnderPipefail(t *testing.T) {
	t.Parallel()

	root := findProjectRootForMirrorTest(t)
	path := filepath.Join(root, releaseWorkflowRelPath)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", releaseWorkflowRelPath, err)
	}

	var offenders []string
	for i, line := range strings.Split(string(data), "\n") {
		m := pipeIntoGrep.FindStringSubmatch(line)
		if m != nil && isQuietGrep(strings.Fields(m[1])) {
			offenders = append(offenders, strings.TrimSpace(line)+"  ("+releaseWorkflowRelPath+":"+strconv.Itoa(i+1)+")")
		}
	}

	if len(offenders) > 0 {
		t.Errorf("PIPEFAIL_SIGPIPE_HAZARD: %s pipes into `grep -q` inside a `set -o pipefail` step.\n"+
			"`grep -q` exits on first match and closes the pipe; the upstream producer dies of SIGPIPE (141),\n"+
			"so pipefail reports 141 for a pipeline whose match SUCCEEDED and the check fails on a passing release.\n"+
			"Fix: drop -q and redirect instead (`| grep -E 'pat' >/dev/null`) so grep consumes all input.\n"+
			"Offending lines:\n  %s", releaseWorkflowRelPath, strings.Join(offenders, "\n  "))
	}
}

// TestIsQuietGrep pins the detector itself. A guard that silently misses a
// spelling is worse than no guard, because the green run reads as coverage.
func TestIsQuietGrep(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args string
		want bool
		why  string
	}{
		{"-q 'pat'", true, "the plain form"},
		{"-qE 'pat'", true, "q in a cluster"},
		{"-qxF 'pat'", true, "q first in a longer cluster"},
		{"-Eq 'pat'", true, "q last in a cluster"},
		{"-E -q 'pat'", true, "q as its own token after another option"},
		{"--quiet 'pat'", true, "the long form"},
		{"--silent 'pat'", true, "the long form's synonym"},
		{"-i --quiet 'pat'", true, "long form after a short one"},

		{"-E 'pat'", false, "no quiet anywhere"},
		{"-c 'pat'", false, "-c counts, it does not suppress early"},
		{"-e -q", false, "-q is -e's pattern, not an option"},
		{"-e 'pat' -q", true, "-e consumed its value, so this -q is an option"},
		{"-f patterns.txt -q", true, "-f consumed its value"},
		{"-m 1 -q", true, "-m consumed its numeric value"},
		{"-- -q", false, "-- ends options; -q is the pattern"},
		{"'q' file", false, "a pattern that happens to be q"},
		{"--regexp -q", false, "--regexp consumed -q as its value"},
		{"--regexp=-q --quiet", true, "--regexp took its value inline, so --quiet still counts"},
	}

	for _, tc := range tests {
		t.Run(tc.args, func(t *testing.T) {
			t.Parallel()
			if got := isQuietGrep(strings.Fields(tc.args)); got != tc.want {
				t.Errorf("isQuietGrep(%q) = %v, want %v — %s", tc.args, got, tc.want, tc.why)
			}
		})
	}
}

func TestPipefailSigpipeMechanism(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("mechanism demo needs a POSIX shell with seq; the enforcing guard above is platform-neutral")
	}
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available; the enforcing guard above is platform-neutral")
	}
	if _, err := exec.LookPath("seq"); err != nil {
		t.Skip("seq not available; the enforcing guard above is platform-neutral")
	}

	// 500k lines is far past the ~64KB pipe buffer, so the producer cannot
	// finish before grep's early exit closes the pipe. The match is on the
	// first line, which makes that exit immediate.
	const producer = "seq 1 500000"

	tests := []struct {
		name     string
		script   string
		wantExit int
		why      string
	}{
		{
			name:     "quiet grep reports SIGPIPE despite matching",
			script:   "set -o pipefail; " + producer + " | grep -q '^1$'",
			wantExit: 141,
			why:      "this is the defect: the pattern matched, yet the pipeline reports failure",
		},
		{
			name:     "consuming grep reports the match",
			script:   "set -o pipefail; " + producer + " | grep '^1$' >/dev/null",
			wantExit: 0,
			why:      "this is the fix: without -q grep reads all input, so the producer never sees SIGPIPE",
		},
		{
			name:     "consuming grep still reports a genuine miss",
			script:   "set -o pipefail; " + producer + " | grep '^no-such-line$' >/dev/null",
			wantExit: 1,
			why:      "the fix must not turn every check into a pass",
		},
		{
			name:     "quiet grep survives when the match is near the end",
			script:   "set -o pipefail; " + producer + " | grep -q '^500000$'",
			wantExit: 0,
			why: "this is why check 2 passed on a 271KB tag annotation: its trailer sits 82 bytes " +
				"from the end, so the producer had nothing left to block on. Total input size is " +
				"not the variable — what is left to write after the match is",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := runBashExitCode(t, tc.script); got != tc.wantExit {
				t.Errorf("exit code = %d, want %d\nscript: %s\nwhy this matters: %s", got, tc.wantExit, tc.script, tc.why)
			}
		})
	}
}

// bashDeadline bounds each shell so a wedged pipe cannot hold the job open
// until the 30-minute CI timeout. The scripts here finish in well under a
// second, so this is a backstop rather than a budget: it is set generously
// because these tests run alongside the rest of the suite, and a deadline
// tight enough to trip under load would report contention as a defect.
const bashDeadline = 60 * time.Second

// runBashExitCode runs script under bash and returns its exit code. A failure
// to start bash at all is fatal; a non-zero exit is a result, not an error.
func runBashExitCode(t *testing.T, script string) int {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), bashDeadline)
	defer cancel()

	err := exec.CommandContext(ctx, "bash", "-c", script).Run()
	if ctx.Err() != nil {
		t.Fatalf("script exceeded %s and was killed — treat as a hang, not a verdict: %s", bashDeadline, script)
	}
	if err == nil {
		return 0
	}

	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		return exitErr.ExitCode()
	}
	t.Fatalf("run bash: %v", err)
	return -1
}
