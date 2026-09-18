package ptycaptest

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
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

const (
	// GateEnv enables pty capture tests ("1").
	GateEnv = "MOAI_PTY_CAPTURE"
	// ChildEnv names the case a TestPtyCaptureChild process runs.
	ChildEnv = "MOAI_PTY_CAPTURE_CHILD"
	// SelfTestEnv selects a forced-failure self-test child (fail|timeout).
	SelfTestEnv = "MOAI_PTY_CAPTURE_SELFTEST"
	// CaseRootEnv lets a parent choose a self-test child's case root, so the
	// parent knows the case working directory for the HOME watch list.
	CaseRootEnv = "MOAI_PTY_CAPTURE_CASEROOT"
	// OutDirEnv is an optional absolute directory captures are exported to;
	// without it they land in the test's temp dir.
	OutDirEnv = "MOAI_PTY_CAPTURE_OUT"
	// CanaryEnv carries a per-case random value, passed only with -e.
	CanaryEnv = "MOAI_PTY_ENV_CANARY"
	// EnvOutEnv is where the child records its effective environment.
	EnvOutEnv = "MOAI_PTY_ENV_OUT"

	// ChildTestName is the test a capture session runs in the child binary.
	ChildTestName = "TestPtyCaptureChild"

	// SessionPrefix starts every tmux session name the harness produces.
	SessionPrefix = "moai-ptycap-"
	// Width and Height are the pty size of every capture session.
	Width  = 80
	Height = 30
	// Term is the TERM the child runs under.
	Term = "xterm-256color"

	// AnchorTimeout is the default wait for an anchor to appear.
	AnchorTimeout = 10 * time.Second

	pollInterval    = 100 * time.Millisecond
	tmuxTimeout     = 10 * time.Second
	buildTimeout    = 5 * time.Minute
	childRunTimeout = 2 * time.Minute
	childTestLimit  = "60s"
)

// kanbanVars is every MOAI_KANBAN* variable in internal/config/envkeys.go;
// each is passed to the child as an empty value so a lane's identity never
// reaches product code. TestPtycapScrubList_KanbanVarsMatchEnvKeys keeps this
// list in step with envkeys.go.
var kanbanVars = []string{
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

// Gate skips without MOAI_PTY_CAPTURE=1 and fails when the gate is on but
// tmux is not on PATH. It must be the first call of every capture test.
func Gate(tb testing.TB) {
	tb.Helper()
	if os.Getenv(GateEnv) != "1" {
		tb.Skipf("pty capture runs only under %s=1", GateEnv)
	}
	if runtime.GOOS == "windows" {
		tb.Skip("pty capture is not available on Windows")
	}
	if _, err := exec.LookPath("tmux"); err != nil {
		tb.Fatalf("%s=1 but tmux is not on PATH: %v", GateEnv, err)
	}
}

var sessionLabelRe = regexp.MustCompile(`[^A-Za-z0-9-]+`)

// SessionName is the ONLY producer of tmux session names in the harness:
// moai-ptycap-<label>-<8 hex>.
// @MX:ANCHOR: [AUTO] sole producer of tmux session names for every capture test
// @MX:REASON: session ownership (acceptance P7) relies on every opened session carrying SessionPrefix; cleanup and the before/after set comparison read only that prefix
func SessionName(label string) string {
	label = strings.Trim(sessionLabelRe.ReplaceAllString(label, "-"), "-")
	if len(label) > 40 {
		label = strings.TrimRight(label[:40], "-")
	}
	if label == "" {
		label = "case"
	}
	return SessionPrefix + label + "-" + randHex(4)
}

func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand: %v", err))
	}
	return hex.EncodeToString(b)
}

// runTmux runs one tmux command with a deadline.
func runTmux(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), tmuxTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "tmux", args...).CombinedOutput()
	return string(out), err
}

// killSession kills exactly one session by exact name; absence is not an error.
func killSession(name string) {
	_, _ = runTmux("kill-session", "-t", "="+name)
}

// ListSessions returns the sorted moai-ptycap- session names. Sessions with
// any other name are never read, compared, or touched.
func ListSessions(tb testing.TB) []string {
	tb.Helper()
	out, err := runTmux("list-sessions", "-F", "#{session_name}")
	if err != nil {
		if strings.Contains(out, "no server running") || strings.Contains(out, "error connecting") {
			return nil
		}
		tb.Fatalf("tmux list-sessions: %v: %s", err, out)
	}
	var names []string
	for _, l := range strings.Split(out, "\n") {
		if strings.HasPrefix(l, SessionPrefix) {
			names = append(names, l)
		}
	}
	sort.Strings(names)
	return names
}

// OpenSentinel opens an idle moai-ptycap-sentinel-<rand> session that must
// survive every harness run; it is killed by exact name at cleanup.
func OpenSentinel(tb testing.TB) string {
	tb.Helper()
	name := SessionName("sentinel")
	out, err := runTmux("new-session", "-d", "-s", name, "-x", "20", "-y", "5", "sleep 600")
	tb.Cleanup(func() { killSession(name) })
	if err != nil {
		tb.Fatalf("open sentinel %s: %v: %s", name, err, out)
	}
	return name
}

// Case is one capture case: its working directory and the temp HOME /
// MOAI_HOME the child gets, the canary, and the env record path.
type Case struct {
	Dir      string
	Home     string
	MoaiHome string
	Canary   string
	EnvOut   string
}

// NewCase makes a case under root (a fresh temp dir when root is empty).
func NewCase(tb testing.TB, root string) *Case {
	tb.Helper()
	dir := tb.TempDir()
	if root != "" {
		dir = filepath.Join(root, "case")
	}
	c := &Case{
		Dir:      dir,
		Home:     filepath.Join(dir, "home"),
		MoaiHome: filepath.Join(dir, "moai-home"),
		Canary:   randHex(8),
		EnvOut:   filepath.Join(dir, "child-env.txt"),
	}
	for _, d := range []string{c.Home, c.MoaiHome} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			tb.Fatalf("case dir %s: %v", d, err)
		}
	}
	return c
}

// ChildEnv is the child environment scrub list (acceptance.md §B P4), in
// order, as VAR=value entries passed one per tmux -e.
func (c *Case) ChildEnv() []string {
	env := []string{
		"HOME=" + c.Home,
		config.EnvHome + "=" + c.MoaiHome,
		config.EnvClaudeConfigDir + "=",
	}
	for _, k := range kanbanVars {
		env = append(env, k+"=")
	}
	return append(env,
		"TERM="+Term,
		CanaryEnv+"="+c.Canary,
		EnvOutEnv+"="+c.EnvOut,
	)
}

// recordedVars is the scrub list the child records (MOAI_PTY_ENV_OUT itself
// excluded), in list order.
func recordedVars() []string {
	vars := []string{"HOME", config.EnvHome, config.EnvClaudeConfigDir}
	vars = append(vars, kanbanVars...)
	return append(vars, "TERM", CanaryEnv)
}

// RecordEnv is the child side of the effective-environment observation:
// before any product code runs it writes VAR=value for every recorded var to
// MOAI_PTY_ENV_OUT, and its working directory to MOAI_PTY_ENV_OUT.cwd.
func RecordEnv() error {
	out := os.Getenv(EnvOutEnv)
	if out == "" {
		return fmt.Errorf("%s is empty", EnvOutEnv)
	}
	var b strings.Builder
	for _, k := range recordedVars() {
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
func checkChildEnvRecord(c *Case, record, realHome, realMoai string) []string {
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
	want := recordedVars()
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
	if got[CanaryEnv] != c.Canary {
		problems = append(problems, fmt.Sprintf("canary %q, want %q", got[CanaryEnv], c.Canary))
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
	for _, k := range kanbanVars {
		if v := got[k]; v != "" {
			problems = append(problems, fmt.Sprintf("%s %q, want empty", k, v))
		}
	}
	if v := got["TERM"]; v != Term {
		problems = append(problems, fmt.Sprintf("TERM %q, want %q", v, Term))
	}
	return problems
}

// VerifyChildEnv reads the child's record and fails the case on any problem.
// It also prints the child's working directory next to the repository path
// and requires them to differ.
func VerifyChildEnv(tb testing.TB, c *Case) {
	tb.Helper()
	record, err := os.ReadFile(c.EnvOut)
	if err != nil {
		tb.Fatalf("child env record %s: %v", c.EnvOut, err)
	}
	realHome, err := os.UserHomeDir()
	if err != nil {
		tb.Fatalf("real HOME: %v", err)
	}
	if problems := checkChildEnvRecord(c, string(record), realHome, filepath.Join(realHome, ".moai")); len(problems) > 0 {
		tb.Fatalf("child effective environment: %v\nrecord:\n%s", problems, record)
	}
	cwdRaw, err := os.ReadFile(c.EnvOut + ".cwd")
	if err != nil {
		tb.Fatalf("child cwd record: %v", err)
	}
	childCwd := strings.TrimSpace(string(cwdRaw))
	repo := repoRoot(tb)
	tb.Logf("child effective environment verified (%d vars); child cwd %s; repository %s", len(recordedVars()), childCwd, repo)
	if resolved, err := filepath.EvalSymlinks(childCwd); err == nil {
		childCwd = resolved
	}
	if childCwd == repo || strings.HasPrefix(childCwd, repo+string(filepath.Separator)) {
		tb.Fatalf("child ran inside the repository (%s)", childCwd)
	}
}

// repoRoot walks up from the working directory to the directory holding go.mod.
func repoRoot(tb testing.TB) string {
	tb.Helper()
	dir, err := os.Getwd()
	if err != nil {
		tb.Fatal(err)
	}
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		dir = resolved
	}
	for d := dir; ; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			return d
		}
		if filepath.Dir(d) == d {
			tb.Fatalf("no go.mod above %s", dir)
		}
	}
}

// BuildChild returns the child test binary: the package pkg built with
// `go test -c` into a temp dir. pkg is resolved against the working directory,
// which `go test` sets to the calling package's directory, so callers pass ".".
// A self-test child is itself such a build, so it reuses its own executable
// instead of building again.
func BuildChild(tb testing.TB, pkg string) string {
	tb.Helper()
	if os.Getenv(SelfTestEnv) != "" {
		exe, err := os.Executable()
		if err != nil {
			tb.Fatalf("self-test child executable: %v", err)
		}
		return exe
	}
	bin := filepath.Join(tb.TempDir(), "child.test")
	ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "go", "test", "-c", "-o", bin, pkg).CombinedOutput()
	if err != nil {
		tb.Fatalf("go test -c %s: %v\n%s", pkg, err, out)
	}
	tb.Logf("built child test binary %s", bin)
	return bin
}

// Session is one open capture session.
type Session struct {
	tb   testing.TB
	Name string
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// Start opens a detached 80x30 session running ChildTestName for childCase in
// c.Dir with the scrubbed environment. remain-on-exit is set in the same tmux
// invocation so a child that exits leaves its last screen readable until
// cleanup kills the session.
func Start(tb testing.TB, c *Case, bin, childCase string) *Session {
	tb.Helper()
	name := SessionName(tb.Name())
	args := []string{"new-session", "-d", "-s", name,
		"-x", fmt.Sprint(Width), "-y", fmt.Sprint(Height), "-c", c.Dir}
	for _, kv := range c.ChildEnv() {
		args = append(args, "-e", kv)
	}
	// TERM is the one scrub-list variable -e cannot deliver: tmux sets a new
	// pane's TERM from the session's default-terminal option, overriding the
	// session environment (observed: the child recorded screen-256color). The
	// assignment in front of exec sets it for the child process itself; the
	// effective-environment record is what confirms it arrived.
	args = append(args,
		"-e", ChildEnv+"="+childCase,
		"-e", GateEnv+"=",
		"-e", SelfTestEnv+"=",
		"TERM="+Term+" exec "+shellQuote(bin)+" -test.run '^"+ChildTestName+"$' -test.v -test.timeout "+childTestLimit,
		";", "set-option", "-w", "-t", "="+name+":", "remain-on-exit", "on",
	)
	out, err := runTmux(args...)
	// Registered before the error is read: a partially successful invocation
	// (session created, option failed) still leaves nothing behind.
	tb.Cleanup(func() { killSession(name) })
	if err != nil {
		tb.Fatalf("tmux new-session %s: %v: %s", name, err, out)
	}
	tb.Logf("ptycap session: %s", name)
	return &Session{tb: tb, Name: name}
}

// Capture returns the current painted screen.
func (s *Session) Capture() (string, error) {
	return runTmux("capture-pane", "-p", "-t", "="+s.Name+":")
}

// WaitFor polls the screen until anchor appears and returns that capture. It
// fails the test with the last capture when the anchor is not visible within
// timeout.
func (s *Session) WaitFor(anchor string, timeout time.Duration) string {
	s.tb.Helper()
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
			s.tb.Fatalf("ptycap: anchor %q not visible within %s in session %s (last error %v); last capture:\n%s",
				anchor, timeout, s.Name, lastErr, last)
		}
		time.Sleep(pollInterval)
	}
}

// SendKeys sends tmux key names to the session.
func (s *Session) SendKeys(keys ...string) {
	s.tb.Helper()
	args := append([]string{"send-keys", "-t", "=" + s.Name + ":"}, keys...)
	if out, err := runTmux(args...); err != nil {
		s.tb.Errorf("tmux send-keys %v: %v: %s", keys, err, out)
	}
}

// Close kills the session now; the registered cleanup makes this idempotent.
func (s *Session) Close() { killSession(s.Name) }

var exportNameRe = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// Export writes content to <MOAI_PTY_CAPTURE_OUT or a temp dir>/<name>.txt and
// returns the path. The harness never writes into the repository on its own;
// exporting evidence is an explicit choice of the caller's environment.
func Export(tb testing.TB, name, content string) string {
	tb.Helper()
	dir := os.Getenv(OutDirEnv)
	if dir == "" {
		dir = tb.TempDir()
	} else if err := os.MkdirAll(dir, 0o755); err != nil {
		tb.Fatalf("export dir %s: %v", dir, err)
	}
	path := filepath.Join(dir, exportNameRe.ReplaceAllString(name, "-")+".txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		tb.Fatalf("export %s: %v", path, err)
	}
	tb.Logf("ptycap export: %s", path)
	return path
}

var sessionLogRe = regexp.MustCompile(`ptycap session: (` + SessionPrefix + `[A-Za-z0-9-]+)`)

// SessionsNamedIn lists the session names a child reported opening.
func SessionsNamedIn(out string) []string {
	var names []string
	for _, m := range sessionLogRe.FindAllStringSubmatch(out, -1) {
		names = append(names, m[1])
	}
	return names
}

// SubprocessEnv is the environment for a child test binary run directly (not
// in tmux): the parent environment with the scrub list applied — temp HOME
// and MOAI_HOME, empty profile selector and lane identity — and then the
// overrides. An override with an empty value sets the variable empty.
func SubprocessEnv(tb testing.TB, overrides map[string]string) []string {
	tb.Helper()
	c := NewCase(tb, "")
	env := map[string]string{}
	for _, kv := range os.Environ() {
		if k, v, ok := strings.Cut(kv, "="); ok {
			env[k] = v
		}
	}
	for _, kv := range c.ChildEnv() {
		k, v, _ := strings.Cut(kv, "=")
		env[k] = v
	}
	delete(env, CanaryEnv)
	delete(env, EnvOutEnv)
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

// RunChild runs the child test binary directly with a deadline and returns
// its combined output.
func RunChild(tb testing.TB, bin string, env []string, args ...string) (string, error) {
	tb.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), childRunTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = env
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

var topLevelResultRe = regexp.MustCompile(`(?m)^--- (PASS|FAIL|SKIP): (\S+)`)

// topLevelResults maps each top-level test name in `go test -v` output to its
// result. Subtest lines are indented and therefore excluded.
func topLevelResults(out string) map[string]string {
	res := map[string]string{}
	for _, m := range topLevelResultRe.FindAllStringSubmatch(out, -1) {
		res[m[2]] = m[1]
	}
	return res
}

// AssertSkipWithoutGate is AC-ITI-019 (a) for the capture tests of the
// package bin was built from: run without MOAI_PTY_CAPTURE, every top-level
// test matching runPattern reports SKIP (at least minResults of them), and
// the moai-ptycap- session set is unchanged with a sentinel present first.
// The caller calls Gate before it.
func AssertSkipWithoutGate(tb testing.TB, bin, runPattern string, minResults int) {
	tb.Helper()
	sentinel := OpenSentinel(tb)
	before := ListSessions(tb)
	if !slices.Contains(before, sentinel) {
		tb.Fatalf("sentinel %s not listed before the run: %v", sentinel, before)
	}

	out, err := RunChild(tb, bin, SubprocessEnv(tb, map[string]string{GateEnv: ""}),
		"-test.run", runPattern, "-test.v")
	tb.Logf("ungated child exit=%v output:\n%s", err, out)
	Export(tb, "ac019a-ungated-child-output", out)
	if err != nil {
		tb.Errorf("ungated child exited non-zero: %v", err)
	}
	results := topLevelResults(out)
	if len(results) < minResults {
		tb.Fatalf("ungated child reported %d top-level results, want >= %d capture tests: %v", len(results), minResults, results)
	}
	for name, res := range results {
		if res != "SKIP" {
			tb.Errorf("ungated: %s reported %s, want SKIP", name, res)
		}
	}
	after := ListSessions(tb)
	if strings.Join(before, ",") != strings.Join(after, ",") {
		tb.Errorf("moai-ptycap- session set changed: before=%v after=%v", before, after)
	}
}

// AssertFailWithoutTmux is AC-ITI-019 (b) for the capture tests of the
// package bin was built from: run with MOAI_PTY_CAPTURE=1 and a PATH without
// tmux, every top-level test matching runPattern reports FAIL (at least
// minResults of them) and the output names tmux. The caller calls Gate before
// it.
func AssertFailWithoutTmux(tb testing.TB, bin, runPattern string, minResults int) {
	tb.Helper()
	emptyPath := tb.TempDir()
	out, err := RunChild(tb, bin, SubprocessEnv(tb, map[string]string{
		GateEnv: "1",
		"PATH":  emptyPath,
	}), "-test.run", runPattern, "-test.v")
	tb.Logf("tmux-less child exit=%v output:\n%s", err, out)
	Export(tb, "ac019b-no-tmux-child-output", out)
	if err == nil {
		tb.Errorf("tmux-less child exited 0; want a failing exit")
	}
	results := topLevelResults(out)
	if len(results) < minResults {
		tb.Fatalf("tmux-less child reported %d top-level results, want >= %d: %v", len(results), minResults, results)
	}
	for name, res := range results {
		if res != "FAIL" {
			tb.Errorf("tmux absent: %s reported %s, want FAIL", name, res)
		}
	}
	if !strings.Contains(out, "tmux") {
		tb.Error("failure output does not name tmux")
	}
}
