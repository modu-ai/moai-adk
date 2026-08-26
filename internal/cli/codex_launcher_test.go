package cli

// codex_launcher_test.go — SPEC-CODEX-LAUNCHER-001 M3 command-surface tests.
//
// AC coverage in this file:
//   - AC-CL-001 — command registration in the LAUNCH group
//   - AC-CL-002 — verb routing closed sets, unknown-token rejection, cwd
//     axis, argv shapes, exit-code propagation, stdio identity, passthrough
//   - AC-CL-003 — --spawn parity, shared tmux-absent diagnostic bytes,
//     readout --spawn rejection
//   - AC-CL-016 — app delegation argv, failure without follow-up launches,
//     the closed set of launched programs, output passthrough (both paths)

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/spf13/cobra"
)

// ─── the capture harness (acceptance "포착 seam") ───────────────────────────
//
// ONE recorder wraps BOTH process-starting sites: the direct path
// (codexDirectLaunchFn receives the assembled *exec.Cmd) and the spawn path
// (codexSpawnLaunchFn receives the new-window target). Seven fields are
// recorded per launch; the stdio trio is read from the *exec.Cmd fields AS
// ASSIGNED, never recomputed.

type codexLaunchRecord struct {
	Program string
	Argv    []string
	Dir     string
	Via     string // "direct" | "spawn"
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

type codexLaunchCapture struct {
	mu      sync.Mutex
	records []codexLaunchRecord

	// failDirectWith, when non-nil, is the error the direct stub returns —
	// the exit-code propagation cells drive it per-cell.
	failDirectWith error
}

func (c *codexLaunchCapture) add(r codexLaunchRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.records = append(c.records, r)
}

// count reports total launches across BOTH sites (the acceptance's
// "두 자리의 합").
func (c *codexLaunchCapture) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.records)
}

func (c *codexLaunchCapture) countVia(via string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := 0
	for _, r := range c.records {
		if r.Via == via {
			n++
		}
	}
	return n
}

// programBasenames is the closed-set axis of AC-CL-016: the set of executable
// basenames every captured launch started.
func (c *codexLaunchCapture) programBasenames() map[string]bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	set := map[string]bool{}
	for _, r := range c.records {
		set[filepath.Base(r.Program)] = true
	}
	return set
}

// withCodexLaunchCapture stubs both launch seams (recording seven fields per
// call) and the binary lookup (sentinel path). Direct runs return
// cap.failDirectWith; spawn runs succeed.
func withCodexLaunchCapture(t *testing.T) *codexLaunchCapture {
	t.Helper()
	cap := &codexLaunchCapture{}
	prevDirect, prevSpawn, prevLook := codexDirectLaunchFn, codexSpawnLaunchFn, codexLookPath
	codexDirectLaunchFn = func(cmd *exec.Cmd) error {
		cap.add(codexLaunchRecord{
			Program: cmd.Path, Argv: cmd.Args, Dir: cmd.Dir, Via: "direct",
			Stdin: cmd.Stdin, Stdout: cmd.Stdout, Stderr: cmd.Stderr,
		})
		return cap.failDirectWith
	}
	codexSpawnLaunchFn = func(dir, program string, args []string) error {
		cap.add(codexLaunchRecord{
			Program: program, Argv: append([]string{program}, args...), Dir: dir, Via: "spawn",
		})
		return nil
	}
	codexLookPath = func(string) (string, error) { return sentinelCodexBinaryPath, nil }
	t.Cleanup(func() {
		codexDirectLaunchFn, codexSpawnLaunchFn, codexLookPath = prevDirect, prevSpawn, prevLook
	})
	return cap
}

// withCodexProjectRoot pins the project-root seam to root (AC-CL-002's cwd
// axis runs through it, never through the raw process cwd).
func withCodexProjectRoot(t *testing.T, root string) *int {
	t.Helper()
	calls := 0
	prev := findProjectRootFn
	findProjectRootFn = func() (string, error) { calls++; return root, nil }
	t.Cleanup(func() { findProjectRootFn = prev })
	return &calls
}

// runCodexCmd invokes runCodex on a FRESH command (no global codexCmd state
// to race on) with buffer-backed streams.
func runCodexCmd(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	var out, errB bytes.Buffer
	c := &cobra.Command{Use: "codex"}
	c.SetOut(&out)
	c.SetErr(&errB)
	err := runCodex(c, args)
	return out.String(), errB.String(), err
}

// codexWantLaunches asserts the two-site launch sum.
func codexWantLaunches(t *testing.T, cap *codexLaunchCapture, total, direct, spawn int) {
	t.Helper()
	if got := cap.count(); got != total {
		t.Errorf("launches (both sites) = %d, want %d", got, total)
	}
	if got := cap.countVia("direct"); got != direct {
		t.Errorf("direct launches = %d, want %d", got, direct)
	}
	if got := cap.countVia("spawn"); got != spawn {
		t.Errorf("spawn launches = %d, want %d", got, spawn)
	}
}

// requireTmuxSpawnEnv skips when the tmux/moai binaries are absent — the
// same discipline spawn_test.go uses (the spawn precondition check runs the
// REAL LookPath calls; stubbing them would stop the test from covering it).
func requireTmuxSpawnEnv(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skipf("tmux binary absent: %v", err)
	}
	if _, err := exec.LookPath("moai"); err != nil {
		t.Skipf("moai binary absent: %v", err)
	}
}

// ─── AC-CL-001 — command registration ─────────────────────────────────────

// TestCodexCommand_RegisteredInLaunchGroup pins the group-block shape on the
// surface users actually see: this repo's root help is the custom tui.Section
// renderer (help.go rootHelpGroups), whose launcher section is titled
// "Launchers" — the cobra group Title itself never reaches the default usage
// template here. The heading appears exactly once and its block contains all
// four launcher names (cc, glm, cg, codex) in ONE block, plus the symbolic
// cobra-group comparison.
func TestCodexCommand_RegisteredInLaunchGroup(t *testing.T) {
	var help bytes.Buffer
	rootCmd.SetOut(&help)
	rootCmd.SetErr(&help)
	t.Cleanup(func() { rootCmd.SetOut(nil); rootCmd.SetErr(nil) })
	if err := rootCmd.Help(); err != nil {
		t.Fatalf("root help: %v", err)
	}

	const heading = "Launchers"
	text := help.String()
	if n := strings.Count(text, heading); n != 1 {
		t.Fatalf("launcher section heading %q appears %d times in root help, want exactly 1\n--- help text ---\n%s", heading, n, text)
	}

	// The section block runs from the heading line to the first blank line;
	// each row is "moai <name> <description>", so the COMMAND NAME is the
	// second token. All four launchers must sit in this ONE block.
	lines := strings.Split(text, "\n")
	block := []string{}
	inBlock := false
	for _, ln := range lines {
		trimmed := strings.TrimSpace(ln)
		if inBlock && strings.HasPrefix(trimmed, heading) {
			t.Fatalf("launcher heading appears twice (second block at %q)", trimmed)
		}
		if strings.HasPrefix(trimmed, heading) {
			inBlock = true
			continue
		}
		if inBlock {
			if trimmed == "" {
				break
			}
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 && fields[0] == "moai" {
				block = append(block, fields[1])
			}
		}
	}
	got := map[string]bool{}
	for _, tok := range block {
		got[tok] = true
	}
	for _, want := range []string{"cc", "glm", "cg", "codex"} {
		if !got[want] {
			t.Errorf("launcher %q missing from the launchers section block (block commands: %v)", want, block)
		}
	}

	// Symbolic comparison — no group-ID string restated in the test.
	if codexCmd.GroupID != ccCmd.GroupID {
		t.Errorf("codexCmd.GroupID (%q) != ccCmd.GroupID (%q)", codexCmd.GroupID, ccCmd.GroupID)
	}
	if codexCmd.Parent() != rootCmd {
		t.Errorf("codexCmd.Parent() != rootCmd")
	}
}

// TestCodexCommand_HelpExitsZero — `moai codex --help` prints help with rc 0.
// DisableFlagParsing means runCodex itself owns the interception. The fresh
// command carries a Short and RunE so the default help template renders the
// Usage block exactly as the real (runnable) codexCmd would.
func TestCodexCommand_HelpExitsZero(t *testing.T) {
	newHelpCmd := func() (*cobra.Command, *bytes.Buffer) {
		var out bytes.Buffer
		c := &cobra.Command{
			Use:   "codex",
			Short: "Codex launcher",
			RunE:  func(*cobra.Command, []string) error { return nil },
		}
		c.SetOut(&out)
		return c, &out
	}
	c, out := newHelpCmd()
	if err := runCodex(c, []string{"--help"}); err != nil {
		t.Fatalf("--help error = %v, want nil (rc 0)", err)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Errorf("--help output missing Usage: block, got %q", out.String())
	}
	c2, out2 := newHelpCmd()
	if err := runCodex(c2, []string{"-h"}); err != nil || !strings.Contains(out2.String(), "Usage:") {
		t.Errorf("-h: err=%v out=%q, want help with rc 0", err, out2.String())
	}
}

// ─── AC-CL-002 — verb routing ──────────────────────────────────────────────

// TestCodexVerbRouting_ClosedSets derives the two token sets from the routing
// table SYMBOLICALLY and asserts the closed-set equations: launch == {cli,
// app}, readout == {"", status}. An implementation routing unknown tokens to
// a launch dies on set inequality.
func TestCodexVerbRouting_ClosedSets(t *testing.T) {
	launch, readout := map[string]bool{}, map[string]bool{}
	for tok, kind := range codexVerbRouting {
		if kind.launches() {
			launch[tok] = true
		} else {
			readout[tok] = true
		}
	}
	if !reflect.DeepEqual(launch, map[string]bool{"cli": true, "app": true}) {
		t.Errorf("launch verb set = %v, want exactly {cli, app}", launch)
	}
	if !reflect.DeepEqual(readout, map[string]bool{"": true, "status": true}) {
		t.Errorf("readout token set = %v, want exactly {\"\", status}", readout)
	}
}

// TestCodexVerbRouting_LaunchCountsPerVerb — bare/status never launch; cli
// and app each launch exactly once (two-site sum).
func TestCodexVerbRouting_LaunchCountsPerVerb(t *testing.T) {
	withCodexSetupProbe(t, minimalProbeStub())
	for _, tc := range []struct {
		name       string
		args       []string
		wantTotal  int
		wantDirect int
		wantSpawn  int
	}{
		{"bare", nil, 0, 0, 0},
		{"status", []string{"status"}, 0, 0, 0},
		{"cli", []string{"cli"}, 1, 1, 0},
		{"app", []string{"app"}, 1, 1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, t.TempDir())
			_, _, err := runCodexCmd(t, tc.args...)
			if err != nil {
				t.Fatalf("runCodex(%v) error = %v", tc.args, err)
			}
			codexWantLaunches(t, cap, tc.wantTotal, tc.wantDirect, tc.wantSpawn)
		})
	}
}

// TestCodexVerbRouting_UnknownTokenRejected — six probe tokens (bogus, cl,
// CLI, Cli, --model, -x) all reject: launch count 0, rc non-zero, diagnostic
// EXACTLY equal to the usage constant.
func TestCodexVerbRouting_UnknownTokenRejected(t *testing.T) {
	for _, tok := range []string{"bogus", "cl", "CLI", "Cli", "--model", "-x"} {
		t.Run(tok, func(t *testing.T) {
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, t.TempDir())
			_, stderr, err := runCodexCmd(t, tok)
			if err == nil {
				t.Fatalf("token %q accepted (err nil), want rejection", tok)
			}
			if code, ok := ResolveExitCode(err); !ok || code == 0 {
				t.Errorf("token %q exit code = (%d, %v), want non-zero deliberate code", tok, code, ok)
			}
			if want := codexUsageDiag + "\n"; stderr != want {
				t.Errorf("token %q diagnostic = %q, want exact usage constant %q", tok, stderr, want)
			}
			codexWantLaunches(t, cap, 0, 0, 0)
		})
	}
}

// TestCodexVerbRouting_CwdCrossChecked — the launch cwd is the PROJECT ROOT
// resolution, not the process cwd: three root scenarios (project root, a
// subdir call still resolving to the root, a worktree root) all propagate
// the resolved root into the captured Dir.
func TestCodexVerbRouting_CwdCrossChecked(t *testing.T) {
	root := "/repo/project"
	worktreeRoot := "/repo/project/.claude/worktrees/t9"
	for _, tc := range []struct {
		name string
		root string
	}{
		{"called at project root", root},
		{"called from a subdirectory still resolves to the root", root},
		{"called inside a worktree resolves to the worktree root", worktreeRoot},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cap := withCodexLaunchCapture(t)
			calls := withCodexProjectRoot(t, tc.root)
			if _, _, err := runCodexCmd(t, "cli"); err != nil {
				t.Fatalf("cli: %v", err)
			}
			if cap.count() != 1 {
				t.Fatalf("launches = %d, want 1", cap.count())
			}
			if got := cap.records[0].Dir; got != tc.root {
				t.Errorf("captured Dir = %q, want resolved project root %q", got, tc.root)
			}
			if *calls != 1 {
				t.Errorf("project-root seam calls = %d, want 1 (cwd flows through the seam, not os.Getwd)", *calls)
			}
		})
	}
}

// TestCodexVerbRouting_AppArgvExact — app's captured argv is exactly
// [codex, app] (argv[0] compared by basename so the sentinel path itself
// stays the stub's business).
func TestCodexVerbRouting_AppArgvExact(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())
	if _, _, err := runCodexCmd(t, "app"); err != nil {
		t.Fatalf("app: %v", err)
	}
	if cap.count() != 1 {
		t.Fatalf("launches = %d, want 1", cap.count())
	}
	rec := cap.records[0]
	if filepath.Base(rec.Argv[0]) != "codex" {
		t.Errorf("argv[0] basename = %q, want codex", filepath.Base(rec.Argv[0]))
	}
	if want := []string{"app"}; !reflect.DeepEqual(rec.Argv[1:], want) {
		t.Errorf("argv tail = %v, want exactly %v", rec.Argv[1:], want)
	}
}

// TestCodexVerbRouting_ExitCodePropagation — five cells (0, 1, 2, 126, 127):
// the direct seam's returned code is moai's exit code verbatim. A
// success-only or collapse-to-one implementation dies on the four non-zero
// cells differing from each other.
func TestCodexVerbRouting_ExitCodePropagation(t *testing.T) {
	for _, code := range []int{0, 1, 2, 126, 127} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, t.TempDir())
			if code != 0 {
				cap.failDirectWith = &exitCodeError{code: code}
			}
			_, _, err := runCodexCmd(t, "cli")
			if code == 0 {
				if err != nil {
					t.Fatalf("code 0 cell: err = %v, want nil", err)
				}
				return
			}
			got, ok := ResolveExitCode(err)
			if !ok {
				t.Fatalf("code %d cell: no deliberate exit code in %v", code, err)
			}
			if got != code {
				t.Errorf("exit code = %d, want %d", got, code)
			}
		})
	}
}

// TestCodexVerbRouting_StdioIdentity — the direct launch hands the child the
// parent's OWN os.Stdin/os.Stdout/os.Stderr values (interface identity, not
// kind-of-same): the precondition interactive tty inheritance rides on.
func TestCodexVerbRouting_StdioIdentity(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())
	if _, _, err := runCodexCmd(t, "cli"); err != nil {
		t.Fatalf("cli: %v", err)
	}
	rec := cap.records[0]
	if rec.Stdin != os.Stdin {
		t.Errorf("captured Stdin (%T %p) is not the test process os.Stdin (%T %p)", rec.Stdin, rec.Stdin, os.Stdin, os.Stdin)
	}
	if rec.Stdout != os.Stdout {
		t.Errorf("captured Stdout (%T %p) is not the test process os.Stdout (%T %p)", rec.Stdout, rec.Stdout, os.Stdout, os.Stdout)
	}
	if rec.Stderr != os.Stderr {
		t.Errorf("captured Stderr (%T %p) is not the test process os.Stderr (%T %p)", rec.Stderr, rec.Stderr, os.Stderr, os.Stderr)
	}
}

// TestCodexVerbRouting_PassthroughTailExact — tokens after -- (spaces,
// quotes, $, =) reach codex verbatim: count, order, and bytes.
func TestCodexVerbRouting_PassthroughTailExact(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())
	tail := []string{"--model", "o3", "a b", "$x", "--flag=v"}
	args := append([]string{"cli", "--"}, tail...)
	if _, _, err := runCodexCmd(t, args...); err != nil {
		t.Fatalf("cli with tail: %v", err)
	}
	if cap.count() != 1 {
		t.Fatalf("launches = %d, want 1", cap.count())
	}
	rec := cap.records[0]
	want := append([]string{"cli"}, tail...)
	if !reflect.DeepEqual(rec.Argv[1:], want) {
		t.Errorf("argv tail = %#v, want exactly %#v", rec.Argv[1:], want)
	}
}

// ─── AC-CL-003 — --spawn parity ────────────────────────────────────────────

// TestCodexSpawn_Parity — cli/app --spawn route through the spawn site
// exactly once with ZERO direct launches, and the spawn capture's
// (program, argv) is token-identical to the SAME tail's direct capture.
func TestCodexSpawn_Parity(t *testing.T) {
	requireTmuxSpawnEnv(t)
	tail := []string{"--model", "o3", "a b", "$x", "--flag=v"}

	for _, verb := range []string{"cli", "app"} {
		t.Run(verb+" spawn once, direct zero", func(t *testing.T) {
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, t.TempDir())
			prev := inTmuxFn
			inTmuxFn = func() bool { return true }
			t.Cleanup(func() { inTmuxFn = prev })
			if _, _, err := runCodexCmd(t, verb, "--spawn"); err != nil {
				t.Fatalf("%s --spawn: %v", verb, err)
			}
			codexWantLaunches(t, cap, 1, 0, 1)
		})
	}

	t.Run("spawn capture token-identical to direct capture", func(t *testing.T) {
		requireTmuxSpawnEnv(t)
		cap := withCodexLaunchCapture(t)
		withCodexProjectRoot(t, t.TempDir())
		prev := inTmuxFn
		inTmuxFn = func() bool { return true }
		t.Cleanup(func() { inTmuxFn = prev })

		if _, _, err := runCodexCmd(t, append([]string{"cli", "--"}, tail...)...); err != nil {
			t.Fatalf("direct: %v", err)
		}
		if _, _, err := runCodexCmd(t, append([]string{"cli", "--spawn", "--"}, tail...)...); err != nil {
			t.Fatalf("spawn: %v", err)
		}
		if cap.count() != 2 {
			t.Fatalf("launches = %d, want 2 (one per site)", cap.count())
		}
		direct, spawn := cap.records[0], cap.records[1]
		if direct.Via != "direct" || spawn.Via != "spawn" {
			t.Fatalf("record order: %q then %q, want direct then spawn", direct.Via, spawn.Via)
		}
		if direct.Program != spawn.Program {
			t.Errorf("program mismatch: direct %q vs spawn %q", direct.Program, spawn.Program)
		}
		if !reflect.DeepEqual(direct.Argv, spawn.Argv) {
			t.Errorf("argv mismatch:\n direct = %#v\n spawn  = %#v", direct.Argv, spawn.Argv)
		}
		// The spawn site opens a NEW WINDOW: no stdio reassignment exists
		// there, and the parity assertion above never needed one.
		if spawn.Stdin != nil || spawn.Stdout != nil || spawn.Stderr != nil {
			t.Errorf("spawn capture carries stdio %v/%v/%v; the new-window target has none", spawn.Stdin, spawn.Stdout, spawn.Stderr)
		}
	})
}

// TestCodexSpawn_TmuxAbsentDiagnosticBytes — with no $TMUX, the codex spawn
// failure and the `moai cc --spawn` failure carry BYTE-IDENTICAL error text
// (same shared check), launch count stays zero.
func TestCodexSpawn_TmuxAbsentDiagnosticBytes(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())
	prev := inTmuxFn
	inTmuxFn = func() bool { return false }
	t.Cleanup(func() { inTmuxFn = prev })

	_, _, codexErr := runCodexCmd(t, "cli", "--spawn")
	if codexErr == nil {
		t.Fatal("codex cli --spawn outside tmux: err nil, want the tmux diagnostic")
	}
	var ccCmdBuf bytes.Buffer
	ccCmd.SetOut(&ccCmdBuf)
	ccErr := ccCmd.RunE(ccCmd, []string{"--spawn"})
	if ccErr == nil {
		t.Fatal("cc --spawn outside tmux: err nil (fixture broken)")
	}
	if ccErr.Error() != codexErr.Error() {
		t.Errorf("diagnostic bytes differ:\n cc    = %q\n codex = %q", ccErr.Error(), codexErr.Error())
	}
	if !strings.Contains(codexErr.Error(), "tmux session required") {
		t.Errorf("codex diagnostic missing the shared tmux phrase: %q", codexErr.Error())
	}
	codexWantLaunches(t, cap, 0, 0, 0)
}

// TestCodexSpawn_RejectedOnReadoutForms — bare/status --spawn refuse with
// rc non-zero and zero launches: a readout is not a new-window target.
func TestCodexSpawn_RejectedOnReadoutForms(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"bare", []string{"--spawn"}},
		{"status", []string{"status", "--spawn"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, t.TempDir())
			_, stderr, err := runCodexCmd(t, tc.args...)
			if err == nil {
				t.Fatalf("%v accepted, want rejection", tc.args)
			}
			if code, ok := ResolveExitCode(err); !ok || code == 0 {
				t.Errorf("exit code = (%d, %v), want non-zero", code, ok)
			}
			if want := codexSpawnReadoutDiag + "\n"; stderr != want {
				t.Errorf("stderr = %q, want %q", stderr, want)
			}
			codexWantLaunches(t, cap, 0, 0, 0)
		})
	}
}

// ─── AC-CL-016 — app delegation ────────────────────────────────────────────

// TestCodexApp_FailureHasNoFollowup — when the launch seam returns a failure
// (codex cannot find its app), NO further process starts and moai's rc
// equals the seam's rc.
func TestCodexApp_FailureHasNoFollowup(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())
	cap.failDirectWith = &exitCodeError{code: 126}
	_, _, err := runCodexCmd(t, "app")
	if err == nil {
		t.Fatal("app with failing seam: err nil")
	}
	codexWantLaunches(t, cap, 1, 1, 0) // exactly the one failed attempt
	got, ok := ResolveExitCode(err)
	if !ok || got != 126 {
		t.Errorf("exit code = (%d, %v), want 126", got, ok)
	}
}

// TestCodexApp_LaunchedProgramsClosedSet — across the launch-bearing cells,
// the captured program basenames are exactly {codex} (axis 1 of the
// closed-set equation; the source-scanning axis lives in the guards file).
func TestCodexApp_LaunchedProgramsClosedSet(t *testing.T) {
	requireTmuxSpawnEnv(t)
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())
	prev := inTmuxFn
	inTmuxFn = func() bool { return true }
	t.Cleanup(func() { inTmuxFn = prev })

	for _, args := range [][]string{
		{"cli"},
		{"app"},
		{"cli", "--", "--flag=v"},
		{"cli", "--spawn"},
		{"app", "--spawn"},
	} {
		if _, _, err := runCodexCmd(t, args...); err != nil {
			t.Fatalf("runCodex(%v): %v", args, err)
		}
	}
	got := cap.programBasenames()
	want := map[string]bool{"codex": true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("launched program basenames = %v, want exactly %v", got, want)
	}
}

// codexAppMessageFixture returns the ABSOLUTE path of the platform leg of
// the AC-CL-016 output fixture (no GOOS skip: every platform runs its own
// leg). Absolute because the launched child runs with Dir set to the project
// root — a relative fixture path would stop resolving there.
func codexAppMessageFixture(t *testing.T) string {
	t.Helper()
	name := "codex-app-message.sh"
	if runtime.GOOS == "windows" {
		name = "codex-app-message.bat"
	}
	path := filepath.Join("testdata", name)
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs %s: %v", path, err)
	}
	if _, err := os.Stat(abs); err != nil {
		t.Fatalf("fixture %s: %v", abs, err)
	}
	return abs
}

// codexAppMessageWant runs the fixture DIRECTLY and returns its exact stdout
// bytes — the byte-for-byte reference the launcher must reproduce.
func codexAppMessageWant(t *testing.T) string {
	t.Helper()
	out, err := exec.Command(codexAppMessageFixture(t)).Output()
	if err != nil {
		t.Fatalf("fixture direct run: %v", err)
	}
	return string(out)
}

// withStdoutCapture temporarily swaps the test process's os.Stdout for a
// pipe, runs fn, and returns everything written. Interface identity in the
// stdio-identity cell is preserved: the launcher assigns the CURRENT value
// of os.Stdout, which is now the pipe — both sides read the same variable.
func withStdoutCapture(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	runErr := fn()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("drain pipe: %v", err)
	}
	return buf.String(), runErr
}

// TestCodexApp_OutputPassthroughDirect — a fixture writing "install the
// desktop app from ..." on stdout survives the launcher byte-for-byte (no
// filtering, rewriting, or reinterpretation) on the DIRECT path.
func TestCodexApp_OutputPassthroughDirect(t *testing.T) {
	fixture := codexAppMessageFixture(t)
	want := codexAppMessageWant(t)

	prevLook, prevDirect := codexLookPath, codexDirectLaunchFn
	codexLookPath = func(string) (string, error) { return fixture, nil }
	codexDirectLaunchFn = func(cmd *exec.Cmd) error { return cmd.Run() }
	t.Cleanup(func() { codexLookPath, codexDirectLaunchFn = prevLook, prevDirect })
	withCodexProjectRoot(t, t.TempDir())

	got, err := withStdoutCapture(t, func() error {
		c := &cobra.Command{Use: "codex"}
		return runCodex(c, []string{"app"})
	})
	if err != nil {
		t.Fatalf("app run: %v", err)
	}
	if got != want {
		t.Errorf("direct passthrough mismatch:\n got  = %q\n want = %q", got, want)
	}
}

// TestCodexApp_OutputPassthroughSpawn — the same fixture through the SPAWN
// site: the stub executes the new-window target in-process (the only way a
// detached window's bytes are observable) and the output is again identical.
func TestCodexApp_OutputPassthroughSpawn(t *testing.T) {
	requireTmuxSpawnEnv(t)
	fixture := codexAppMessageFixture(t)
	want := codexAppMessageWant(t)

	prevLook, prevSpawn := codexLookPath, codexSpawnLaunchFn
	codexLookPath = func(string) (string, error) { return fixture, nil }
	codexSpawnLaunchFn = func(dir, program string, args []string) error {
		c := exec.Command(program, args...)
		c.Dir = dir
		// A new-window child inherits the window's streams — modeled here by
		// the current (pipe-swapped) os.Stdout/os.Stderr, so the fixture's
		// bytes are observable exactly as they would be in the window.
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}
	t.Cleanup(func() { codexLookPath, codexSpawnLaunchFn = prevLook, prevSpawn })
	withCodexProjectRoot(t, t.TempDir())
	prev := inTmuxFn
	inTmuxFn = func() bool { return true }
	t.Cleanup(func() { inTmuxFn = prev })

	got, err := withStdoutCapture(t, func() error {
		c := &cobra.Command{Use: "codex"}
		return runCodex(c, []string{"app", "--spawn"})
	})
	if err != nil {
		t.Fatalf("app --spawn run: %v", err)
	}
	if got != want {
		t.Errorf("spawn passthrough mismatch:\n got  = %q\n want = %q", got, want)
	}
}

// TestCodexDirect_RawExitErrorIsNotSilent pins the execerr discipline on the
// codex launch path: a NON-exit failure (start failure) is described via
// StatusDetail and NEVER resolves to a deliberate exit code — while the
// child-exit path (above) deliberately propagates.
func TestCodexDirect_RawExitErrorIsNotSilent(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())
	// A sentinel error that is neither an ExitCoder nor an exec.ExitError.
	cap.failDirectWith = errors.New("boom: binary vanished")
	_, _, err := runCodexCmd(t, "cli")
	if err == nil {
		t.Fatal("start failure swallowed")
	}
	if _, ok := ResolveExitCode(err); ok {
		t.Errorf("a non-exit start failure resolved as a deliberate exit code: %v", err)
	}
	if !strings.Contains(err.Error(), "boom: binary vanished") {
		t.Errorf("failure text lost the cause: %q", err.Error())
	}
}

// TestCodexApp_RealChildExitCodePropagates — the REAL child-exit axis: the
// fixture child is driven to exit 7 through its env hook and moai's rc must
// be 7 (the deliberate propagation converts the raw exec.ExitError; the
// stubbed seam cells above prove the mapping, this cell proves it against a
// genuine subprocess exit). Runs on every platform (each leg uses its own
// fixture).
func TestCodexApp_RealChildExitCodePropagates(t *testing.T) {
	fixture := codexAppMessageFixture(t)
	t.Setenv("CODEX_FIXTURE_EXIT", "7")

	prevLook, prevDirect := codexLookPath, codexDirectLaunchFn
	codexLookPath = func(string) (string, error) { return fixture, nil }
	codexDirectLaunchFn = func(cmd *exec.Cmd) error { return cmd.Run() }
	t.Cleanup(func() { codexLookPath, codexDirectLaunchFn = prevLook, prevDirect })
	withCodexProjectRoot(t, t.TempDir())

	c := &cobra.Command{Use: "codex"}
	err := runCodex(c, []string{"app"})
	if err == nil {
		t.Fatal("child exited 7, moai reported success")
	}
	got, ok := ResolveExitCode(err)
	if !ok || got != 7 {
		t.Errorf("exit code = (%d, %v), want deliberate 7 (err: %v)", got, ok, err)
	}
}

// TestCodexSpawn_RealAssemblyThroughStubTmux drives the REAL spawn assembly
// (checkSpawnPrereqs -> defaultCodexSpawnLaunch -> buildCodexSpawnCommand)
// with only the tmux primitive stubbed: the command string handed to tmux
// must be the token-by-token shell-quoted codex line, and a tmux failure
// wraps without launching anything else.
func TestCodexSpawn_RealAssemblyThroughStubTmux(t *testing.T) {
	requireTmuxSpawnEnv(t)
	fixture := codexAppMessageFixture(t)

	prevLook, prevSpawn, prevTmux, prevInTmux := codexLookPath, codexSpawnLaunchFn, tmuxSpawnFn, inTmuxFn
	codexLookPath = func(string) (string, error) { return fixture, nil }
	// Keep the REAL spawn implementation; stub only the tmux primitive.
	codexSpawnLaunchFn = defaultCodexSpawnLaunch
	inTmuxFn = func() bool { return true }
	var gotDir, gotCommand string
	tmuxSpawnFn = func(dir, command string) (string, error) {
		gotDir, gotCommand = dir, command
		return "%9", nil
	}
	t.Cleanup(func() {
		codexLookPath, codexSpawnLaunchFn, tmuxSpawnFn, inTmuxFn = prevLook, prevSpawn, prevTmux, prevInTmux
	})
	root := t.TempDir()
	withCodexProjectRoot(t, root)

	var spawnOut string
	var runErr error
	spawnOut, runErr = func() (string, error) {
		c := &cobra.Command{Use: "codex"}
		return withStdoutCapture(t, func() error { return runCodex(c, []string{"cli", "--spawn", "--", "--flag=v", "a b"}) })
	}()
	if runErr != nil {
		t.Fatalf("spawn run: %v", runErr)
	}
	if gotDir != root {
		t.Errorf("tmux cwd = %q, want the project root %q", gotDir, root)
	}
	// The command line is the shell-quoted token sequence: the codex path,
	// the verb, and the verbatim tail — a token containing a space must
	// arrive quoted, not split.
	wantCommand := shellQuote(fixture) + " cli --flag=v " + shellQuote("a b")
	if gotCommand != wantCommand {
		t.Errorf("tmux command = %q, want %q", gotCommand, wantCommand)
	}
	if !strings.Contains(spawnOut, "Spawned pane %9") {
		t.Errorf("spawn report missing pane line: %q", spawnOut)
	}

	// A tmux failure surfaces as a wrapped error and nothing else launches.
	tmuxSpawnFn = func(string, string) (string, error) { return "", errors.New("no server") }
	c2 := &cobra.Command{Use: "codex"}
	err := runCodex(c2, []string{"cli", "--spawn"})
	if err == nil || !strings.Contains(err.Error(), "no server") {
		t.Errorf("tmux failure = %v, want wrapped no-server error", err)
	}
}

// TestCodexVerbRouting_HelpAfterDashDashIsNotHelp — a --help token AFTER the
// passthrough marker belongs to codex, not to the launcher: no help
// interception, the readout rejects the tail with the usage constant.
func TestCodexVerbRouting_HelpAfterDashDashIsNotHelp(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())
	stdout, stderr, err := runCodexCmd(t, "--", "--help")
	if err == nil {
		t.Fatal("bare readout with passthrough tail accepted, want usage rejection")
	}
	if want := codexUsageDiag + "\n"; stderr != want {
		t.Errorf("stderr = %q, want usage constant", stderr)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty (no help rendered)", stdout)
	}
	codexWantLaunches(t, cap, 0, 0, 0)
}

// TestCodexVerbRouting_UnresolvableRootFallsBackToCwd — when the project root
// cannot be resolved, the launch degrades to the process cwd rather than
// refusing (the cwd axis is "resolve then launch", never "fail closed").
func TestCodexVerbRouting_UnresolvableRootFallsBackToCwd(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	prev := findProjectRootFn
	findProjectRootFn = func() (string, error) { return "", errors.New("no marker") }
	t.Cleanup(func() { findProjectRootFn = prev })

	if _, _, err := runCodexCmd(t, "cli"); err != nil {
		t.Fatalf("cli with unresolvable root: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if cap.count() != 1 || cap.records[0].Dir != cwd {
		t.Errorf("captured Dir = %q (launches %d), want process cwd %q", cap.records[0].Dir, cap.count(), cwd)
	}
}

// ─── M4 — help and example copy (REQ-CL-014) ────────────────────────────────

// TestCodexCommand_HelpCopyGuidance — M4's deliverable is the guidance copy
// itself. Machine-judgeable properties, all derived (never restated):
//
//  1. the wiring action the help names is THE action the readout prints and
//     the init generator accepts — derived from codexWiringAction so the
//     help cannot drift from either surface
//  2. every non-empty verb of the closed routing set (codexVerbRouting) is
//     documented in Long, plus the --spawn and -- passthrough notes — the
//     help cannot drop a verb that routes or document one that does not
//  3. the Example field exists and is copy-pasteable: every non-comment,
//     non-blank line is a moai codex invocation
//
// The neutrality judgment over the same strings lives in
// codex_launcher_guards_test.go (AC-CL-013) and re-runs over whatever copy
// these subtests accept.
func TestCodexCommand_HelpCopyGuidance(t *testing.T) {
	t.Run("wiring action matches the generator", func(t *testing.T) {
		want := strings.TrimPrefix(codexWiringAction, "run ")
		if !strings.Contains(codexCmd.Long, want) {
			t.Errorf("Long does not name the wiring action %q (derived from codexWiringAction)", want)
		}
	})
	t.Run("every routing verb documented", func(t *testing.T) {
		for token := range codexVerbRouting {
			if token == "" {
				continue
			}
			if !strings.Contains(codexCmd.Long, token) {
				t.Errorf("Long does not document routing verb %q", token)
			}
		}
		for _, note := range []string{"--spawn", "--"} {
			if !strings.Contains(codexCmd.Long, note) {
				t.Errorf("Long does not mention %q", note)
			}
		}
	})
	t.Run("examples are copy-pasteable", func(t *testing.T) {
		if codexCmd.Example == "" {
			t.Fatal("Example is empty — M4 requires example copy")
		}
		for _, line := range strings.Split(codexCmd.Example, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			if !strings.HasPrefix(trimmed, "moai codex") {
				t.Errorf("example line %q is not a moai codex invocation", trimmed)
			}
		}
	})
}
