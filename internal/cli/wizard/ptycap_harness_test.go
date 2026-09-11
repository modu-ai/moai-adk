package wizard

// pty capture harness (acceptance.md §B, design.md §11).
//
// A capture case runs this package's test binary — built with `go test -c`
// into a temp dir — inside a detached tmux session sized 80x30, and reads the
// painted screen with `tmux capture-pane -p`. That is the only place a render
// defect is observed the way a terminal shows it; View() goldens are the
// regression guard, the pty capture is the repair verdict.
//
// Safety contract:
//   - Runs only under MOAI_PTY_CAPTURE=1; tmux missing under the gate is a
//     FAIL, never a SKIP.
//   - Every session name comes from ptycapSessionName (prefix moai-ptycap-),
//     and every session is killed by its exact name from a t.Cleanup
//     registered right after creation. No kill-server, no prefix sweep.
//   - The child environment is passed variable by variable with tmux -e and
//     then observed: the child records what it actually received, and the
//     parent checks that record instead of trusting the -e flags.
//   - Every subprocess is bounded by a context deadline, and the child test
//     binary by -test.timeout.

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

const (
	// ptycapGateEnv enables pty capture tests ("1").
	ptycapGateEnv = "MOAI_PTY_CAPTURE"
	// ptycapChildEnv names the case a TestPtyCaptureChild process runs.
	ptycapChildEnv = "MOAI_PTY_CAPTURE_CHILD"
	// ptycapSelfTestEnv selects a forced-failure self-test child (fail|timeout).
	ptycapSelfTestEnv = "MOAI_PTY_CAPTURE_SELFTEST"
	// ptycapCaseRootEnv lets a parent choose a self-test child's case root, so
	// the parent knows the case working directory for the HOME watch list.
	ptycapCaseRootEnv = "MOAI_PTY_CAPTURE_CASEROOT"
	// ptycapOutDirEnv is an optional absolute directory captures are exported
	// to; without it they land in the test's temp dir.
	ptycapOutDirEnv = "MOAI_PTY_CAPTURE_OUT"
	// ptycapCanaryEnv carries a per-case random value, passed only with -e.
	ptycapCanaryEnv = "MOAI_PTY_ENV_CANARY"
	// ptycapEnvOutEnv is where the child records its effective environment.
	ptycapEnvOutEnv = "MOAI_PTY_ENV_OUT"

	ptycapSessionPrefix = "moai-ptycap-"
	ptycapWidth         = 80
	ptycapHeight        = 30
	ptycapTerm          = "xterm-256color"

	ptycapPollInterval    = 100 * time.Millisecond
	ptycapAnchorTimeout   = 10 * time.Second
	ptycapSelfTestTimeout = 2 * time.Second
	ptycapTmuxTimeout     = 10 * time.Second
	ptycapBuildTimeout    = 5 * time.Minute
	ptycapChildRunTimeout = 2 * time.Minute
	ptycapChildTestLimit  = "60s"

	// ptycapCaseInitFirstPage runs the init wizard form (en) on the TTY.
	ptycapCaseInitFirstPage = "init-first-page"
	// ptycapNeverAnchor is a string no case ever renders.
	ptycapNeverAnchor = "ZZ-PTYCAP-ANCHOR-NEVER-RENDERED"
)

// ptycapKanbanVars is every MOAI_KANBAN* variable in internal/config/envkeys.go;
// each is passed to the child as an empty value so a lane's identity never
// reaches product code. TestPtycapScrubList_KanbanVarsMatchEnvKeys keeps this
// list in step with envkeys.go.
var ptycapKanbanVars = []string{
	config.EnvMoaiKanban,
	config.EnvMoaiKanbanSpec,
	config.EnvMoaiKanbanID,
	config.EnvMoaiKanbanLabel,
	config.EnvMoaiKanbanSettingsInjected,
	config.EnvMoaiKanbanLeadAddr,
	config.EnvMoaiKanbanBackend,
	config.EnvMoaiKanbanCard,
	config.EnvMoaiKanbanLeadName,
}

// ptycapGate skips without MOAI_PTY_CAPTURE=1 and fails when the gate is on
// but tmux is not on PATH. It must be the first call of every capture test.
func ptycapGate(t *testing.T) {
	t.Helper()
	if os.Getenv(ptycapGateEnv) != "1" {
		t.Skipf("pty capture runs only under %s=1", ptycapGateEnv)
	}
	if runtime.GOOS == "windows" {
		t.Skip("pty capture is not available on Windows")
	}
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Fatalf("%s=1 but tmux is not on PATH: %v", ptycapGateEnv, err)
	}
}

var sessionLabelRe = regexp.MustCompile(`[^A-Za-z0-9-]+`)

// ptycapSessionName is the ONLY producer of tmux session names in this
// package: moai-ptycap-<label>-<8 hex>.
func ptycapSessionName(label string) string {
	label = strings.Trim(sessionLabelRe.ReplaceAllString(label, "-"), "-")
	if len(label) > 40 {
		label = strings.TrimRight(label[:40], "-")
	}
	if label == "" {
		label = "case"
	}
	return ptycapSessionPrefix + label + "-" + randHex(4)
}

func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand: %v", err))
	}
	return hex.EncodeToString(b)
}

// ptycapTmux runs one tmux command with a deadline.
func ptycapTmux(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ptycapTmuxTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "tmux", args...).CombinedOutput()
	return string(out), err
}

// ptycapKill kills exactly one session by exact name; absence is not an error.
func ptycapKill(name string) {
	_, _ = ptycapTmux("kill-session", "-t", "="+name)
}

// ptycapListSessions returns the sorted moai-ptycap- session names. Sessions
// with any other name are never read, compared, or touched.
func ptycapListSessions(t *testing.T) []string {
	t.Helper()
	out, err := ptycapTmux("list-sessions", "-F", "#{session_name}")
	if err != nil {
		if strings.Contains(out, "no server running") || strings.Contains(out, "error connecting") {
			return nil
		}
		t.Fatalf("tmux list-sessions: %v: %s", err, out)
	}
	var names []string
	for _, l := range strings.Split(out, "\n") {
		if strings.HasPrefix(l, ptycapSessionPrefix) {
			names = append(names, l)
		}
	}
	sort.Strings(names)
	return names
}

// ptycapOpenSentinel opens an idle moai-ptycap-sentinel-<rand> session that
// must survive every harness run; it is killed by exact name at cleanup.
func ptycapOpenSentinel(t *testing.T) string {
	t.Helper()
	name := ptycapSessionName("sentinel")
	out, err := ptycapTmux("new-session", "-d", "-s", name, "-x", "20", "-y", "5", "sleep 600")
	t.Cleanup(func() { ptycapKill(name) })
	if err != nil {
		t.Fatalf("open sentinel %s: %v: %s", name, err, out)
	}
	return name
}

// ptycapCase is one capture case: its working directory and the temp HOME /
// MOAI_HOME the child gets, the canary, and the env record path.
type ptycapCase struct {
	Dir      string
	Home     string
	MoaiHome string
	Canary   string
	EnvOut   string
}

// ptycapNewCase makes a case under root (a fresh t.TempDir() when empty).
func ptycapNewCase(t *testing.T, root string) *ptycapCase {
	t.Helper()
	dir := t.TempDir()
	if root != "" {
		dir = filepath.Join(root, "case")
	}
	c := &ptycapCase{
		Dir:      dir,
		Home:     filepath.Join(dir, "home"),
		MoaiHome: filepath.Join(dir, "moai-home"),
		Canary:   randHex(8),
		EnvOut:   filepath.Join(dir, "child-env.txt"),
	}
	for _, d := range []string{c.Home, c.MoaiHome} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("case dir %s: %v", d, err)
		}
	}
	return c
}

// childEnv is the child environment scrub list (acceptance.md §B P4), in
// order, as VAR=value entries passed one per tmux -e.
func (c *ptycapCase) childEnv() []string {
	env := []string{
		"HOME=" + c.Home,
		config.EnvHome + "=" + c.MoaiHome,
		config.EnvClaudeConfigDir + "=",
	}
	for _, k := range ptycapKanbanVars {
		env = append(env, k+"=")
	}
	return append(env,
		"TERM="+ptycapTerm,
		ptycapCanaryEnv+"="+c.Canary,
		ptycapEnvOutEnv+"="+c.EnvOut,
	)
}

// ptycapRecordedVars is the scrub list the child records (MOAI_PTY_ENV_OUT
// itself excluded), in list order.
func ptycapRecordedVars() []string {
	vars := []string{"HOME", config.EnvHome, config.EnvClaudeConfigDir}
	vars = append(vars, ptycapKanbanVars...)
	return append(vars, "TERM", ptycapCanaryEnv)
}

// ptycapRecordEnv is the child side of the effective-environment observation:
// before any product code runs it writes VAR=value for every recorded var to
// MOAI_PTY_ENV_OUT, and its working directory to MOAI_PTY_ENV_OUT.cwd.
func ptycapRecordEnv() error {
	out := os.Getenv(ptycapEnvOutEnv)
	if out == "" {
		return fmt.Errorf("%s is empty", ptycapEnvOutEnv)
	}
	var b strings.Builder
	for _, k := range ptycapRecordedVars() {
		b.WriteString(k + "=" + os.Getenv(k) + "\n")
	}
	if err := os.WriteFile(out, []byte(b.String()), 0o644); err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	return os.WriteFile(out+".cwd", []byte(cwd+"\n"), 0o644)
}

// checkChildEnvRecord applies the four effective-environment assertions to a
// child's record and returns every problem found.
func checkChildEnvRecord(c *ptycapCase, record, realHome, realMoai string) []string {
	var problems []string
	got := map[string]string{}
	lines := 0
	for _, l := range strings.Split(strings.TrimRight(record, "\n"), "\n") {
		if l == "" {
			continue
		}
		lines++
		k, v, _ := strings.Cut(l, "=")
		got[k] = v
	}
	want := ptycapRecordedVars()
	// (1) one line per recorded variable
	if lines != len(want) {
		problems = append(problems, fmt.Sprintf("record has %d lines, want %d", lines, len(want)))
	}
	for _, k := range want {
		if _, ok := got[k]; !ok {
			problems = append(problems, "record lacks "+k)
		}
	}
	// (2) canary proves the record belongs to this case's child
	if got[ptycapCanaryEnv] != c.Canary {
		problems = append(problems, fmt.Sprintf("canary %q, want %q", got[ptycapCanaryEnv], c.Canary))
	}
	// (3) HOME and MOAI_HOME are the case paths, not the real ones
	if got["HOME"] != c.Home || got["HOME"] == realHome {
		problems = append(problems, fmt.Sprintf("HOME %q, want the case path %q (real HOME %q)", got["HOME"], c.Home, realHome))
	}
	if got[config.EnvHome] != c.MoaiHome || got[config.EnvHome] == realMoai {
		problems = append(problems, fmt.Sprintf("%s %q, want the case path %q (real %q)", config.EnvHome, got[config.EnvHome], c.MoaiHome, realMoai))
	}
	// (4) profile selector and lane identity are empty
	if v := got[config.EnvClaudeConfigDir]; v != "" {
		problems = append(problems, fmt.Sprintf("%s %q, want empty", config.EnvClaudeConfigDir, v))
	}
	for _, k := range ptycapKanbanVars {
		if v := got[k]; v != "" {
			problems = append(problems, fmt.Sprintf("%s %q, want empty", k, v))
		}
	}
	if v := got["TERM"]; v != ptycapTerm {
		problems = append(problems, fmt.Sprintf("TERM %q, want %q", v, ptycapTerm))
	}
	return problems
}

// ptycapVerifyChildEnv reads the child's record and fails the case on any
// problem. It also prints the child's working directory next to the repository
// path and requires them to differ.
func ptycapVerifyChildEnv(t *testing.T, c *ptycapCase) {
	t.Helper()
	record, err := os.ReadFile(c.EnvOut)
	if err != nil {
		t.Fatalf("child env record %s: %v", c.EnvOut, err)
	}
	realHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("real HOME: %v", err)
	}
	if problems := checkChildEnvRecord(c, string(record), realHome, filepath.Join(realHome, ".moai")); len(problems) > 0 {
		t.Fatalf("child effective environment: %v\nrecord:\n%s", problems, record)
	}
	cwdRaw, err := os.ReadFile(c.EnvOut + ".cwd")
	if err != nil {
		t.Fatalf("child cwd record: %v", err)
	}
	childCwd := strings.TrimSpace(string(cwdRaw))
	repo := repoRoot(t)
	t.Logf("child effective environment verified (%d vars); child cwd %s; repository %s", len(ptycapRecordedVars()), childCwd, repo)
	if resolved, err := filepath.EvalSymlinks(childCwd); err == nil {
		childCwd = resolved
	}
	if childCwd == repo || strings.HasPrefix(childCwd, repo+string(filepath.Separator)) {
		t.Fatalf("child ran inside the repository (%s)", childCwd)
	}
}

// repoRoot walks up from the package dir to the directory holding go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		dir = resolved
	}
	for d := dir; ; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			return d
		}
		if filepath.Dir(d) == d {
			t.Fatalf("no go.mod above %s", dir)
		}
	}
}

// ptycapBuildChild returns the child test binary: this package built with
// `go test -c` into a temp dir. A self-test child is itself such a build, so
// it reuses its own executable instead of building again.
func ptycapBuildChild(t *testing.T) string {
	t.Helper()
	if os.Getenv(ptycapSelfTestEnv) != "" {
		exe, err := os.Executable()
		if err != nil {
			t.Fatalf("self-test child executable: %v", err)
		}
		return exe
	}
	bin := filepath.Join(t.TempDir(), "wizard.test")
	ctx, cancel := context.WithTimeout(context.Background(), ptycapBuildTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "go", "test", "-c", "-o", bin, ".").CombinedOutput()
	if err != nil {
		t.Fatalf("go test -c: %v\n%s", err, out)
	}
	t.Logf("built child test binary %s", bin)
	return bin
}

// ptycapSession is one open capture session.
type ptycapSession struct {
	t    *testing.T
	Name string
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// ptycapStart opens a detached 80x30 session running TestPtyCaptureChild for
// childCase in c.Dir with the scrubbed environment. remain-on-exit is set in
// the same tmux invocation so a child that exits leaves its last screen
// readable until cleanup kills the session.
func ptycapStart(t *testing.T, c *ptycapCase, bin, childCase string) *ptycapSession {
	t.Helper()
	name := ptycapSessionName(t.Name())
	args := []string{"new-session", "-d", "-s", name,
		"-x", fmt.Sprint(ptycapWidth), "-y", fmt.Sprint(ptycapHeight), "-c", c.Dir}
	for _, kv := range c.childEnv() {
		args = append(args, "-e", kv)
	}
	// TERM is the one scrub-list variable -e cannot deliver: tmux sets a new
	// pane's TERM from the session's default-terminal option, overriding the
	// session environment (observed: the child recorded screen-256color). The
	// assignment in front of exec sets it for the child process itself; the
	// effective-environment record is what confirms it arrived.
	args = append(args,
		"-e", ptycapChildEnv+"="+childCase,
		"-e", ptycapGateEnv+"=",
		"-e", ptycapSelfTestEnv+"=",
		"TERM="+ptycapTerm+" exec "+shellQuote(bin)+" -test.run '^TestPtyCaptureChild$' -test.v -test.timeout "+ptycapChildTestLimit,
		";", "set-option", "-w", "-t", "="+name+":", "remain-on-exit", "on",
	)
	out, err := ptycapTmux(args...)
	// Registered before the error is read: a partially successful invocation
	// (session created, option failed) still leaves nothing behind.
	t.Cleanup(func() { ptycapKill(name) })
	if err != nil {
		t.Fatalf("tmux new-session %s: %v: %s", name, err, out)
	}
	t.Logf("ptycap session: %s", name)
	return &ptycapSession{t: t, Name: name}
}

// Capture returns the current painted screen.
func (s *ptycapSession) Capture() (string, error) {
	return ptycapTmux("capture-pane", "-p", "-t", "="+s.Name+":")
}

// WaitFor polls the screen until anchor appears and returns that capture. It
// fails the test with the last capture when the anchor is not visible within
// timeout.
func (s *ptycapSession) WaitFor(anchor string, timeout time.Duration) string {
	s.t.Helper()
	deadline := time.Now().Add(timeout)
	var last string
	var lastErr error
	for {
		out, err := s.Capture()
		if err == nil {
			last = out
			if strings.Contains(out, anchor) {
				return out
			}
		}
		lastErr = err
		if time.Now().After(deadline) {
			s.t.Fatalf("ptycap: anchor %q not visible within %s in session %s (last error %v); last capture:\n%s",
				anchor, timeout, s.Name, lastErr, last)
		}
		time.Sleep(ptycapPollInterval)
	}
}

// SendKeys sends tmux key names to the session.
func (s *ptycapSession) SendKeys(keys ...string) {
	s.t.Helper()
	args := append([]string{"send-keys", "-t", "=" + s.Name + ":"}, keys...)
	if out, err := ptycapTmux(args...); err != nil {
		s.t.Errorf("tmux send-keys %v: %v: %s", keys, err, out)
	}
}

// Close kills the session now; the registered cleanup makes this idempotent.
func (s *ptycapSession) Close() { ptycapKill(s.Name) }

var exportNameRe = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// ptycapExport writes content to <MOAI_PTY_CAPTURE_OUT or t.TempDir()>/<name>.txt
// and returns the path. The harness never writes into the repository on its
// own; exporting evidence is an explicit choice of the caller's environment.
func ptycapExport(t *testing.T, name, content string) string {
	t.Helper()
	dir := os.Getenv(ptycapOutDirEnv)
	if dir == "" {
		dir = t.TempDir()
	} else if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("export dir %s: %v", dir, err)
	}
	path := filepath.Join(dir, exportNameRe.ReplaceAllString(name, "-")+".txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("export %s: %v", path, err)
	}
	t.Logf("ptycap export: %s", path)
	return path
}

var sessionLogRe = regexp.MustCompile(`ptycap session: (` + ptycapSessionPrefix + `[A-Za-z0-9-]+)`)

// ptycapSessionsNamedIn lists the session names a child reported opening.
func ptycapSessionsNamedIn(out string) []string {
	var names []string
	for _, m := range sessionLogRe.FindAllStringSubmatch(out, -1) {
		names = append(names, m[1])
	}
	return names
}

// ptycapSubprocessEnv is the environment for a child test binary run directly
// (not in tmux): the parent environment with the scrub list applied — temp
// HOME and MOAI_HOME, empty profile selector and lane identity — and then the
// overrides. An override with an empty value sets the variable empty.
func ptycapSubprocessEnv(t *testing.T, overrides map[string]string) []string {
	t.Helper()
	c := ptycapNewCase(t, "")
	env := map[string]string{}
	for _, kv := range os.Environ() {
		if k, v, ok := strings.Cut(kv, "="); ok {
			env[k] = v
		}
	}
	for _, kv := range c.childEnv() {
		k, v, _ := strings.Cut(kv, "=")
		env[k] = v
	}
	delete(env, ptycapCanaryEnv)
	delete(env, ptycapEnvOutEnv)
	for k, v := range overrides {
		env[k] = v
	}
	list := make([]string, 0, len(env))
	for k, v := range env {
		list = append(list, k+"="+v)
	}
	sort.Strings(list)
	return list
}

// ptycapRunChild runs the child test binary directly with a deadline and
// returns its combined output.
func ptycapRunChild(t *testing.T, bin string, env []string, args ...string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), ptycapChildRunTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = env
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}
