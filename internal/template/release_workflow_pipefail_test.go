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
// The hazard is size-dependent: it stays invisible while the producer's output
// fits the ~64KB pipe buffer, and becomes deterministic once it does not.
// CHANGELOG.md crossed that line (416KB at the time of writing), which is why
// check 5 was the one that broke — but checks 2 and 6 had the identical shape
// and would have broken as soon as their inputs grew.
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
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// releaseWorkflowRelPath is the repo-relative path of the release workflow
// whose provenance gate carries the seven checks.
const releaseWorkflowRelPath = ".github/workflows/release.yml"

// quietGrepAfterPipe matches a pipe feeding a grep that carries -q in any flag
// cluster (-q, -qE, -qxF, ...). The pipe is the load-bearing part: without an
// upstream producer there is nothing for grep's early exit to kill.
var quietGrepAfterPipe = regexp.MustCompile(`\|\s*grep\s+-[A-Za-z]*q`)

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
		if quietGrepAfterPipe.MatchString(line) {
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

// runBashExitCode runs script under bash and returns its exit code. A failure
// to start bash at all is fatal; a non-zero exit is a result, not an error.
func runBashExitCode(t *testing.T, script string) int {
	t.Helper()

	err := exec.Command("bash", "-c", script).Run()
	if err == nil {
		return 0
	}

	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		return exitErr.ExitCode()
	}
	t.Fatalf("run bash: %v", err)
	return -1
}
