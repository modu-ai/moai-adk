// release_fixture_test.go: shared scratch-repository fixtures for the release
// mechanics guards (SPEC-GITHUB-FLOW-DEFAULT-001 M3, AC-GFD-007..010).
//
// The release scripts under test read a git repository and, for the matrix
// gate, the GitHub CLI. Neither may touch this repository or the network, so
// every test builds a throwaway bare "origin" plus a working clone under
// t.TempDir() and puts a stub `gh` first on PATH. Nothing here creates a tag or
// pushes anywhere except to those scratch origins.
//
// All helpers carry the `rls` prefix: internal/template/ holds many sibling
// test files in one package and the short generic names are taken.
package template_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	rlsReleaseScriptRel = "scripts/release.sh"
	rlsProvenanceRel    = "scripts/verify-release-provenance.sh"
	rlsDeadline         = 90 * time.Second
)

var (
	rlsSHARe  = regexp.MustCompile(`[0-9a-f]{40}`)
	rlsANSIRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

// rlsNorm strips colour codes, replaces every full commit SHA with a stable
// token and trims the result, so a verdict can be compared across fixtures.
func rlsNorm(s string) string {
	s = rlsANSIRe.ReplaceAllString(s, "")
	s = rlsSHARe.ReplaceAllString(s, "<sha>")
	return strings.TrimSpace(s)
}

// rlsRequireTools skips (visibly) where the fixture cannot run. The scripts
// are bash and the fixtures are git; a missing tool is a Gap, never a pass.
func rlsRequireTools(t *testing.T, tools ...string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("release scripts are bash and the fixtures use POSIX paths; not exercised on windows")
	}
	for _, tool := range append([]string{"bash", "git"}, tools...) {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not available; the release script fixtures cannot run", tool)
		}
	}
}

// rlsEnv builds a hermetic environment: no inherited GIT_* or MOAI_* variables
// (a lane's kanban and release-harness variables must not steer the script),
// no global or system git config, fixed author and dates.
func rlsEnv(extra ...string) []string {
	var env []string
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "GIT_") || strings.HasPrefix(kv, "MOAI_") {
			continue
		}
		env = append(env, kv)
	}
	env = append(env,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_TERMINAL_PROMPT=0",
		"GIT_AUTHOR_NAME=Fixture",
		"GIT_AUTHOR_EMAIL=fixture@example.invalid",
		"GIT_COMMITTER_NAME=Fixture",
		"GIT_COMMITTER_EMAIL=fixture@example.invalid",
		"GIT_AUTHOR_DATE=2026-01-01T00:00:00Z",
		"GIT_COMMITTER_DATE=2026-01-01T00:00:00Z",
	)
	return append(env, extra...)
}

type rlsResult struct {
	exit int
	out  string // stdout and stderr interleaved, exactly as the process wrote them
}

// rlsRun runs a command and returns its exit code and combined output. A
// failure to start the process is fatal; a non-zero exit is a result.
func rlsRun(t *testing.T, dir string, env []string, name string, args ...string) rlsResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), rlsDeadline)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("%s %v exceeded %s and was killed (a hang, not a verdict)\npartial output: %s", name, args, rlsDeadline, out)
	}
	if err == nil {
		return rlsResult{exit: 0, out: string(out)}
	}
	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		return rlsResult{exit: exitErr.ExitCode(), out: string(out)}
	}
	t.Fatalf("start %s: %v", name, err)
	return rlsResult{}
}

// rlsGit runs git in dir and fails the test on a non-zero exit.
func rlsGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	res := rlsRun(t, dir, rlsEnv(), "git", args...)
	if res.exit != 0 {
		t.Fatalf("git %v in %s: exit %d\n%s", args, dir, res.exit, res.out)
	}
	return strings.TrimSpace(res.out)
}

// rlsRepo is a bare scratch origin and a working clone of it.
type rlsRepo struct {
	t      *testing.T
	origin string
	work   string
}

func rlsNewRepo(t *testing.T) *rlsRepo {
	t.Helper()
	base := t.TempDir()
	r := &rlsRepo{t: t, origin: filepath.Join(base, "origin.git"), work: filepath.Join(base, "work")}
	rlsGit(t, base, "init", "-q", "--bare", "-b", "main", r.origin)
	rlsGit(t, base, "init", "-q", "-b", "main", r.work)
	rlsGit(t, r.work, "remote", "add", "origin", r.origin)
	return r
}

// commit writes files (creating directories), commits them on the current
// branch and returns the new commit SHA.
func (r *rlsRepo) commit(files map[string]string, msg string) string {
	r.t.Helper()
	for rel, content := range files {
		abs := filepath.Join(r.work, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			r.t.Fatalf("mkdir for %s: %v", rel, err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			r.t.Fatalf("write %s: %v", rel, err)
		}
	}
	rlsGit(r.t, r.work, "add", "-A")
	rlsGit(r.t, r.work, "commit", "-q", "-m", msg)
	return rlsGit(r.t, r.work, "rev-parse", "HEAD")
}

func (r *rlsRepo) git(args ...string) string {
	r.t.Helper()
	return rlsGit(r.t, r.work, args...)
}

// installRepoFile copies a file of the real repository into the scratch clone
// as an UNTRACKED file (the provenance script reads tags through `git show`,
// so the working-tree copy never enters a verdict). A file that does not exist
// yet is skipped: before the script is written, its absence is the observation.
func (r *rlsRepo) installRepoFile(rel string) bool {
	r.t.Helper()
	src := filepath.Join(findProjectRootForMirrorTest(r.t), filepath.FromSlash(rel))
	data, err := os.ReadFile(src)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false
		}
		r.t.Fatalf("read %s: %v", rel, err)
	}
	dst := filepath.Join(r.work, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		r.t.Fatalf("mkdir for %s: %v", rel, err)
	}
	if err := os.WriteFile(dst, data, 0o755); err != nil {
		r.t.Fatalf("write %s: %v", rel, err)
	}
	return true
}

// rlsWorkflowStep is the part of a workflow step the guards care about.
type rlsWorkflowStep struct {
	Name string            `yaml:"name"`
	Env  map[string]string `yaml:"env"`
	Run  string            `yaml:"run"`
}

type rlsWorkflowDoc struct {
	Jobs map[string]struct {
		Steps []rlsWorkflowStep `yaml:"steps"`
	} `yaml:"jobs"`
}

// rlsProvenanceStep returns the verify-provenance job's step that carries the
// TAG environment variable (the step whose body is the provenance gate).
func rlsProvenanceStep(t *testing.T) rlsWorkflowStep {
	t.Helper()
	path := filepath.Join(findProjectRootForMirrorTest(t), filepath.FromSlash(releaseWorkflowRelPath))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", releaseWorkflowRelPath, err)
	}
	var doc rlsWorkflowDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse %s: %v", releaseWorkflowRelPath, err)
	}
	job, ok := doc.Jobs["verify-provenance"]
	if !ok {
		t.Fatalf("%s has no verify-provenance job", releaseWorkflowRelPath)
	}
	var found []rlsWorkflowStep
	for _, s := range job.Steps {
		if s.Env["TAG"] != "" {
			found = append(found, s)
		}
	}
	if len(found) != 1 {
		t.Fatalf("verify-provenance must have exactly one step with a TAG env, found %d", len(found))
	}
	return found[0]
}

// rlsWriteScript writes body to a file in dir and returns its path.
func rlsWriteScript(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o755); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

// rlsMustContain fails the test unless every wanted substring is in out.
func rlsMustContain(t *testing.T, out string, wants ...string) {
	t.Helper()
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("output lacks %q\n--- output ---\n%s", w, out)
		}
	}
}

// rlsGhStub is a stand-in for the GitHub CLI. It is the only seam through which
// release.sh reaches GitHub in these tests: every call is appended to a log
// file (so "no gh run list call" is an assertion on facts, not on the script's
// intent), and `gh run list` / `gh run view` answer from canned JSON files with
// the same projection and --jq semantics the real CLI applies. --commit and
// --event filter the list the way the server does; --json projects each object
// to the requested fields, so a script that reads a field it did not request
// gets null, as it would from the real CLI.
const rlsGhStub = `#!/usr/bin/env bash
set -u
echo "$*" >> "${GH_STUB_LOG:?}"
sub="${1:-} ${2:-}"
case "$sub" in
  "api "*|"api")
    echo "success"
    exit 0
    ;;
  "pr list")
    exit 0
    ;;
  "run list")
    shift 2
    commit=""; event=""; jqexpr="."; fields=""
    while [ $# -gt 0 ]; do
      case "$1" in
        --commit|-c) commit="$2"; shift 2 ;;
        --event|-e) event="$2"; shift 2 ;;
        --jq|-q) jqexpr="$2"; shift 2 ;;
        --json) fields="$2"; shift 2 ;;
        *) shift ;;
      esac
    done
    jq -c --arg c "$commit" --arg e "$event" --arg f "$fields" '
      [ .[]
        | select(($c == "" or .headSha == $c) and ($e == "" or .event == $e))
        | with_entries(select(.key as $k | ($f | split(",")) | index($k))) ]' "${GH_STUB_RUNS:?}" | jq -r "$jqexpr"
    exit 0
    ;;
  "run view")
    id="$3"
    shift 3
    jqexpr="."; fields=""
    while [ $# -gt 0 ]; do
      case "$1" in
        --jq|-q) jqexpr="$2"; shift 2 ;;
        --json) fields="$2"; shift 2 ;;
        *) shift ;;
      esac
    done
    f="${GH_STUB_JOBS_DIR:?}/${id}.json"
    [ -f "$f" ] || { echo "gh stub: no jobs file for run ${id}" >&2; exit 1; }
    jq -c --arg f "$fields" 'with_entries(select(.key as $k | ($f | split(",")) | index($k)))' "$f" | jq -r "$jqexpr"
    exit 0
    ;;
esac
echo "gh stub: unsupported call: $*" >&2
exit 1
`

// rlsStubEnv installs the gh stub in a fresh directory and returns the
// environment that puts it first on PATH, the call-log path and the directory
// the canned run data lives in.
type rlsStub struct {
	env     []string
	log     string
	runs    string
	jobsDir string
}

func rlsNewStub(t *testing.T) *rlsStub {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatalf("mkdir stub bin: %v", err)
	}
	rlsWriteScript(t, bin, "gh", rlsGhStub)
	s := &rlsStub{
		log:     filepath.Join(dir, "gh-calls.log"),
		runs:    filepath.Join(dir, "runs.json"),
		jobsDir: filepath.Join(dir, "jobs"),
	}
	if err := os.MkdirAll(s.jobsDir, 0o755); err != nil {
		t.Fatalf("mkdir stub jobs: %v", err)
	}
	if err := os.WriteFile(s.log, nil, 0o644); err != nil {
		t.Fatalf("create stub log: %v", err)
	}
	s.setRuns(t, "[]")
	s.env = rlsEnv(
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"GH_STUB_LOG="+s.log,
		"GH_STUB_RUNS="+s.runs,
		"GH_STUB_JOBS_DIR="+s.jobsDir,
	)
	return s
}

func (s *rlsStub) setRuns(t *testing.T, json string) {
	t.Helper()
	if err := os.WriteFile(s.runs, []byte(json), 0o644); err != nil {
		t.Fatalf("write stub runs: %v", err)
	}
}

func (s *rlsStub) setJobs(t *testing.T, runID int, json string) {
	t.Helper()
	p := filepath.Join(s.jobsDir, fmt.Sprintf("%d.json", runID))
	if err := os.WriteFile(p, []byte(json), 0o644); err != nil {
		t.Fatalf("write stub jobs: %v", err)
	}
}

// calls returns the logged gh invocations, one per line.
func (s *rlsStub) calls(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile(s.log)
	if err != nil {
		t.Fatalf("read stub log: %v", err)
	}
	var out []string
	for _, l := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}
