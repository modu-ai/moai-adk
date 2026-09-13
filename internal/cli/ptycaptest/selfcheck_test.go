package ptycaptest

// Self-checks for the harness (harness.go), the real HOME watch list
// (homewatch.go), and the render helpers (render.go). The harness has to be
// trustworthy before any rendering verdict is read from it: it must fail when
// it cannot run, fail when an anchor never appears, never leave a session
// behind, never kill a session it did not open, and observe — not assume — the
// child environment.
//
// The pty self-checks run a trivial fixture child (TestPtyCaptureChild below)
// so they exercise the harness itself, independent of any product screen. The
// product-screen normal run lives with the product package (internal/cli/wizard).
//
// Tests named TestPtyCapture_* need tmux and run only under MOAI_PTY_CAPTURE=1.
// Everything else here is pure and runs in every `go test`.

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

const (
	// fixtureCase is the fixture child case: it prints fixtureAnchor and the
	// fixture lines, then blocks on stdin until the session is interrupted.
	fixtureCase = "fixture-anchor"
	// fixtureAnchor is what the fixture child paints first.
	fixtureAnchor = "PTYCAP FIXTURE READY"
	// neverAnchor is a string no case ever renders.
	neverAnchor = "ZZ-PTYCAP-ANCHOR-NEVER-RENDERED"
	// selfTestTimeout bounds the forced-timeout self-test's wait.
	selfTestTimeout = 2 * time.Second
)

// fixtureLines are painted after the anchor; the normal run requires them.
var fixtureLines = []string{"English", "Korean (한국어)", "Japanese (日本語)", "Chinese (中文)"}

// ---------------------------------------------------------------------------
// Render helpers
// ---------------------------------------------------------------------------

func TestPtycapStripANSI(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"plain text unchanged", "Select conversation language", "Select conversation language"},
		{"sgr colour", "\x1b[38;2;217;119;87mTitle\x1b[0m", "Title"},
		{"cursor movement", "\x1b[2K\x1b[1Gline", "line"},
		{"private mode", "\x1b[?25lhidden cursor\x1b[?25h", "hidden cursor"},
		{"osc hyperlink bel", "\x1b]8;;https://example.com\x07link\x1b]8;;\x07", "link"},
		{"osc hyperlink st", "\x1b]8;;https://example.com\x1b\\link\x1b]8;;\x1b\\", "link"},
		{"wide text kept", "\x1b[1m한국어\x1b[0m", "한국어"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := StripANSI(tc.in); got != tc.want {
				t.Errorf("StripANSI(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestPtycapDisplayColumn(t *testing.T) {
	line := "┃ Korean (한국어) - 한국어"
	// "┃ " = 2 columns, "Korean (" = 8, "한국어" = 3 wide runes = 6, ") " = 2.
	if got := DisplayColumn(line, "- 한국어"); got != 18 {
		t.Errorf("display column of the description dash = %d, want 18", got)
	}
	// The same position measured in runes and in bytes differs, which is what
	// makes the display-width measurement the one that matters for alignment.
	idx := strings.Index(line, "- 한국어")
	if runes := len([]rune(line[:idx])); runes == 18 {
		t.Errorf("fixture does not separate display width from rune count (both %d)", runes)
	}
	if idx == 18 {
		t.Errorf("fixture does not separate display width from byte offset (both 18)")
	}
	if got := DisplayColumn(line, "absent"); got != -1 {
		t.Errorf("display column of an absent substring = %d, want -1", got)
	}
	if got := DisplayColumn("Japanese (日本語) - 日本語", "Japanese"); got != 0 {
		t.Errorf("display column at line start = %d, want 0", got)
	}
}

func TestPtycapCompareGolden(t *testing.T) {
	dir := t.TempDir()

	if err := CompareGolden(dir, "missing", "frame\n", false); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("absent golden without update: want a 'missing' error, got %v", err)
	}
	if err := CompareGolden(dir, "fresh", "line one\nline two\n", true); err != nil {
		t.Fatalf("update of an absent golden: %v", err)
	}
	if b, err := os.ReadFile(filepath.Join(dir, "fresh.golden")); err != nil || string(b) != "line one\nline two\n" {
		t.Fatalf("update wrote %q (err %v)", b, err)
	}
	if err := CompareGolden(dir, "fresh", "line one\nline two\n", false); err != nil {
		t.Fatalf("identical frame: %v", err)
	}
	err := CompareGolden(dir, "fresh", "line one\nline 2\n", false)
	if err == nil {
		t.Fatal("changed frame compared equal")
	}
	if !strings.Contains(err.Error(), "line 2") || !strings.Contains(err.Error(), "line two") {
		t.Errorf("mismatch error does not show both sides of the changed line: %v", err)
	}
	if err := CompareGolden(dir, "fresh", "line one\nline 2\n", true); err != nil {
		t.Fatalf("update of a changed golden: %v", err)
	}
	if err := CompareGolden(dir, "fresh", "line one\nline 2\n", false); err != nil {
		t.Fatalf("after update the new frame must match: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Real HOME watch list (acceptance.md §B P8)
// ---------------------------------------------------------------------------

func TestHomeWatch_ListShape(t *testing.T) {
	want := []string{"W1", "W2", "W3", "W4", "W5", "W6"}
	seen := map[string]bool{}
	for _, it := range homeWatchList {
		seen[it.ID] = true
		if it.Producer == "" {
			t.Errorf("%s %s: no producing function recorded", it.ID, it.Rel)
		}
		if it.Base != watchBaseHome && it.Base != watchBaseMoai {
			t.Errorf("%s %s: base %q", it.ID, it.Rel, it.Base)
		}
	}
	for _, id := range want {
		if !seen[id] {
			t.Errorf("watch list has no %s row", id)
		}
	}
	w6 := 0
	for _, it := range homeWatchList {
		if it.ID == "W6" {
			w6++
			if it.Kind != watchAbsent || !strings.Contains(it.Rel, watchKeyToken) {
				t.Errorf("W6 row %q must be an absence item carrying %s", it.Rel, watchKeyToken)
			}
		}
	}
	if w6 != 3 {
		t.Errorf("W6 rows = %d, want 3 (db, run, cache/search)", w6)
	}
}

// seedFakeHome builds the AC-ITI-020 (4) fake HOME: one file each for W1, W3,
// W5, the W2 profile p1, and the W4 directory with one file inside.
func seedFakeHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	write := func(rel, body string) {
		t.Helper()
		p := filepath.Join(home, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(".moai/claude-profiles/preferences.yaml", "user_name: a\n")    // W1
	write(".moai/claude-profiles/p1/preferences.yaml", "user_name: b\n") // W2
	write(".claude/settings.json", "{}\n")                               // W3
	write(".claude/hooks/moai/handle.sh", "#!/bin/sh\n")                 // W4
	write(".zshrc", "# rc\n")                                            // W5
	return home
}

func TestHomeWatch_PositiveControl(t *testing.T) {
	workDir := t.TempDir()
	key := homestate.ProjectKey(workDir)

	cases := []struct {
		name     string
		mutate   func(t *testing.T, home string)
		wantPass bool
		wantPath string
	}{
		{"(i) nothing changed", func(*testing.T, string) {}, true, ""},
		{"(ii) W3 one byte changed", func(t *testing.T, home string) {
			if err := os.WriteFile(filepath.Join(home, ".claude/settings.json"), []byte("{}\r"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, false, filepath.Join(".claude", "settings.json")},
		{"(iii) W6 key directory created", func(t *testing.T, home string) {
			if err := os.MkdirAll(filepath.Join(home, ".moai", "db", key), 0o755); err != nil {
				t.Fatal(err)
			}
		}, false, filepath.Join(".moai", "db", key)},
		{"(iv) new W2 glob match", func(t *testing.T, home string) {
			p := filepath.Join(home, ".moai/claude-profiles/p2/preferences.yaml")
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, false, filepath.Join("claude-profiles", "p2", "preferences.yaml")},
		{"(v) excluded subtree written", func(t *testing.T, home string) {
			p := filepath.Join(home, ".moai/claude-profiles/p1/projects/x.jsonl")
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte("{}\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, true, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := seedFakeHome(t)
			before, err := SnapshotHome(home, []string{key})
			if err != nil {
				t.Fatal(err)
			}
			// Seeded: W1 default preferences, W2 p1, W3, the W4 directory, W5 .zshrc.
			if before.Existing() != 5 {
				t.Fatalf("seeded fake HOME shows %d existing watch entries, want 5 (positive existence first)", before.Existing())
			}
			tc.mutate(t, home)
			after, err := SnapshotHome(home, []string{key})
			if err != nil {
				t.Fatal(err)
			}
			changed := DiffSnapshots(before, after)
			t.Logf("result %q: entries=%d existing=%d changed=%v", tc.name, len(after), after.Existing(), changed)
			if tc.wantPass {
				if len(changed) != 0 {
					t.Errorf("want PASS, comparison reported %v", changed)
				}
				return
			}
			if len(changed) == 0 {
				t.Fatalf("want FAIL, comparison reported no change")
			}
			found := false
			for _, p := range changed {
				if strings.Contains(p, tc.wantPath) {
					found = true
				}
			}
			if !found {
				t.Errorf("FAIL output %v does not name %s", changed, tc.wantPath)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Child environment scrub list (acceptance.md §B P4)
// ---------------------------------------------------------------------------

// TestPtycapScrubList_KanbanVarsMatchEnvKeys keeps the scrub list in step with
// internal/config/envkeys.go: a new MOAI_KANBAN* constant there must be added
// to the list, or a lane identity could leak into the child.
func TestPtycapScrubList_KanbanVarsMatchEnvKeys(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "..", "config", "envkeys.go"))
	if err != nil {
		t.Fatal(err)
	}
	declared := regexp.MustCompile(`=\s*"(MOAI_KANBAN[A-Z_]*)"`).FindAllStringSubmatch(string(src), -1)
	if len(declared) == 0 {
		t.Fatal("no MOAI_KANBAN constant found in envkeys.go (positive existence first)")
	}
	inList := map[string]bool{}
	for _, name := range kanbanVars {
		inList[name] = true
	}
	for _, m := range declared {
		if !inList[m[1]] {
			t.Errorf("envkeys.go declares %s but the child scrub list does not carry it", m[1])
		}
	}
	if len(kanbanVars) != len(declared) {
		t.Errorf("scrub list carries %d MOAI_KANBAN vars, envkeys.go declares %d", len(kanbanVars), len(declared))
	}
}

func TestPtycapChildEnv_Values(t *testing.T) {
	c := NewCase(t, "")
	env := map[string]string{}
	var order []string
	for _, kv := range c.ChildEnv() {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			t.Fatalf("malformed env entry %q", kv)
		}
		env[k] = v
		order = append(order, k)
	}
	if len(order) == 0 {
		t.Fatal("empty child env (positive existence first)")
	}
	checks := map[string]string{
		"HOME":                    filepath.Join(c.Dir, "home"),
		config.EnvHome:            filepath.Join(c.Dir, "moai-home"),
		config.EnvClaudeConfigDir: "",
		"TERM":                    "xterm-256color",
		EnvOutEnv:                 filepath.Join(c.Dir, "child-env.txt"),
	}
	for k, want := range checks {
		got, ok := env[k]
		if !ok {
			t.Errorf("child env does not carry %s", k)
			continue
		}
		if got != want {
			t.Errorf("child env %s=%q, want %q", k, got, want)
		}
	}
	if !filepath.IsAbs(env[config.EnvHome]) {
		t.Errorf("%s must be absolute, got %q", config.EnvHome, env[config.EnvHome])
	}
	if env[CanaryEnv] == "" || env[CanaryEnv] != c.Canary {
		t.Errorf("canary %q does not match the case canary %q", env[CanaryEnv], c.Canary)
	}
	for _, k := range kanbanVars {
		v, ok := env[k]
		if !ok {
			t.Errorf("child env does not carry %s", k)
		} else if v != "" {
			t.Errorf("child env %s=%q, want empty", k, v)
		}
	}
	if other := NewCase(t, ""); other.Canary == c.Canary {
		t.Error("two cases share a canary")
	}
}

func TestPtycapCheckChildEnvRecord(t *testing.T) {
	c := NewCase(t, "")
	realHome, realMoai := "/real/home", "/real/home/.moai"
	good := func() map[string]string {
		m := map[string]string{
			"HOME":                    filepath.Join(c.Dir, "home"),
			config.EnvHome:            filepath.Join(c.Dir, "moai-home"),
			config.EnvClaudeConfigDir: "",
			"TERM":                    "xterm-256color",
			CanaryEnv:                 c.Canary,
		}
		for _, k := range kanbanVars {
			m[k] = ""
		}
		return m
	}
	render := func(m map[string]string) string {
		var b strings.Builder
		for _, k := range recordedVars() {
			if v, ok := m[k]; ok {
				b.WriteString(k + "=" + v + "\n")
			}
		}
		return b.String()
	}

	if problems := checkChildEnvRecord(c, render(good()), realHome, realMoai); len(problems) != 0 {
		t.Fatalf("a correct record reported problems: %v", problems)
	}
	mutants := map[string]func(m map[string]string){
		"wrong canary":         func(m map[string]string) { m[CanaryEnv] = "not-the-canary" },
		"real HOME leaked":     func(m map[string]string) { m["HOME"] = realHome },
		"real MOAI_HOME":       func(m map[string]string) { m[config.EnvHome] = realMoai },
		"kanban identity":      func(m map[string]string) { m[config.EnvMoaiKanbanID] = "t999" },
		"claude config dir":    func(m map[string]string) { m[config.EnvClaudeConfigDir] = "/x" },
		"line missing":         func(m map[string]string) { delete(m, "TERM") },
		"home not case-scoped": func(m map[string]string) { m["HOME"] = "/tmp/elsewhere" },
	}
	for name, mutate := range mutants {
		m := good()
		mutate(m)
		if problems := checkChildEnvRecord(c, render(m), realHome, realMoai); len(problems) == 0 {
			t.Errorf("mutant %q: the record check reported no problem", name)
		}
	}
}

func TestPtycapSessionName(t *testing.T) {
	re := regexp.MustCompile(`^moai-ptycap-[A-Za-z0-9-]+-[0-9a-f]{8}$`)
	a := SessionName("TestSomething/sub case")
	b := SessionName("TestSomething/sub case")
	for _, n := range []string{a, b, SessionName("sentinel")} {
		if !re.MatchString(n) {
			t.Errorf("session name %q does not match %s", n, re)
		}
	}
	if a == b {
		t.Errorf("two names for the same label collide: %q", a)
	}
	if !strings.HasPrefix(SessionName("sentinel"), SessionPrefix+"sentinel-") {
		t.Error("sentinel name must read moai-ptycap-sentinel-<rand>")
	}
}

func TestPtycapTopLevelResults(t *testing.T) {
	out := "=== RUN   TestA\n--- PASS: TestA (0.00s)\n=== RUN   TestB\n    --- FAIL: TestB/sub (0.00s)\n--- FAIL: TestB (0.00s)\n--- SKIP: TestC (0.00s)\n"
	got := topLevelResults(out)
	want := map[string]string{"TestA": "PASS", "TestB": "FAIL", "TestC": "SKIP"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: got %s, want %s", k, got[k], v)
		}
	}
}

// ---------------------------------------------------------------------------
// Child helpers (run only as a child of a harness test)
// ---------------------------------------------------------------------------

// TestPtyCaptureChild is the fixture program a pty session runs. It records
// its effective environment first, paints the fixture anchor and lines, then
// blocks on its TTY until the session interrupts it.
func TestPtyCaptureChild(t *testing.T) {
	name := os.Getenv(ChildEnv)
	if name == "" {
		t.Skip("runs only inside a pty capture session")
	}
	if err := RecordEnv(); err != nil {
		t.Fatalf("record child env: %v", err)
	}
	switch name {
	case fixtureCase:
		fmt.Println(fixtureAnchor)
		for _, l := range fixtureLines {
			fmt.Println(l)
		}
		_, err := bufio.NewReader(os.Stdin).ReadString('\n')
		t.Logf("fixture stdin returned: %v", err)
	default:
		t.Fatalf("unknown child case %q", name)
	}
}

// TestPtyCaptureSelfTestChild opens a real harness session and then fails on
// purpose (fail) or waits for an anchor that never renders (timeout). A parent
// test runs it as a separate `go test` process and inspects what it left.
func TestPtyCaptureSelfTestChild(t *testing.T) {
	mode := os.Getenv(SelfTestEnv)
	if mode == "" {
		t.Skip("runs only as a self-test child")
	}
	Gate(t)
	bin := BuildChild(t, ".")
	c := NewCase(t, os.Getenv(CaseRootEnv))
	s := Start(t, c, bin, fixtureCase)
	switch mode {
	case "fail":
		s.WaitFor(fixtureAnchor, AnchorTimeout)
		t.Fatal("forced failure after the session opened")
	case "timeout":
		s.WaitFor(neverAnchor, selfTestTimeout)
		t.Fatal("unreachable: WaitFor must fail on a missing anchor")
	default:
		t.Fatalf("unknown self-test mode %q", mode)
	}
}

// ---------------------------------------------------------------------------
// pty self-checks (need tmux, gated by MOAI_PTY_CAPTURE=1)
// ---------------------------------------------------------------------------

// TestPtyCapture_NormalRun — AC-ITI-020 (1) on the fixture child, plus the
// effective-environment observation and the real HOME watch comparison.
func TestPtyCapture_NormalRun(t *testing.T) {
	Gate(t)
	bin := BuildChild(t, ".")
	c := NewCase(t, "")
	checkHome := WatchRealHome(t, c.Dir)

	sentinel := OpenSentinel(t)
	before := ListSessions(t)
	if !slices.Contains(before, sentinel) {
		t.Fatalf("sentinel %s not listed before the run: %v", sentinel, before)
	}

	s := Start(t, c, bin, fixtureCase)
	frame := s.WaitFor(fixtureAnchor, AnchorTimeout)
	VerifyChildEnv(t, c)
	RequireLines(t, frame, append([]string{fixtureAnchor}, fixtureLines...)...)
	path := Export(t, "normal-run-fixture", frame)
	if b, err := os.ReadFile(path); err != nil || !strings.Contains(string(b), fixtureAnchor) {
		t.Fatalf("capture file %s does not carry the anchor (err %v)", path, err)
	}
	s.SendKeys("C-c")
	s.Close()

	after := ListSessions(t)
	t.Logf("moai-ptycap- sessions before=%v after=%v", before, after)
	if strings.Join(before, ",") != strings.Join(after, ",") {
		t.Errorf("moai-ptycap- session set changed: before=%v after=%v", before, after)
	}
	checkHome()
}

// TestPtyCapture_ForcedFailure — AC-ITI-020 (2).
func TestPtyCapture_ForcedFailure(t *testing.T) {
	Gate(t)
	runSelfTest(t, "fail", "forced failure after the session opened")
}

// TestPtyCapture_ForcedTimeout — AC-ITI-020 (3).
func TestPtyCapture_ForcedTimeout(t *testing.T) {
	Gate(t)
	runSelfTest(t, "timeout", "not visible within")
}

// runSelfTest runs TestPtyCaptureSelfTestChild in a separate process and
// checks AC-ITI-020 (2)/(3): the child reports FAIL with the expected message,
// the session it opened is gone, the sentinel survives, and the real HOME
// watch list is unchanged.
func runSelfTest(t *testing.T, mode, wantMsg string) {
	t.Helper()
	bin := BuildChild(t, ".")
	caseRoot := t.TempDir()
	checkHome := WatchRealHome(t, filepath.Join(caseRoot, "case"))

	sentinel := OpenSentinel(t)
	before := ListSessions(t)
	if !slices.Contains(before, sentinel) {
		t.Fatalf("sentinel %s not listed before the run: %v", sentinel, before)
	}

	out, err := RunChild(t, bin, SubprocessEnv(t, map[string]string{
		GateEnv:     "1",
		SelfTestEnv: mode,
		CaseRootEnv: caseRoot,
	}), "-test.run", "^TestPtyCaptureSelfTestChild$", "-test.v", "-test.timeout", "60s")
	t.Logf("self-test child (%s) exit=%v output:\n%s", mode, err, out)
	Export(t, "selftest-"+mode+"-child-output", out)

	if err == nil {
		t.Errorf("self-test child exited 0; want a failing exit")
	}
	if !strings.Contains(out, "--- FAIL: TestPtyCaptureSelfTestChild") {
		t.Errorf("self-test child output carries no FAIL line")
	}
	if !strings.Contains(out, wantMsg) {
		t.Errorf("self-test child output lacks %q", wantMsg)
	}
	opened := SessionsNamedIn(out)
	if len(opened) == 0 {
		t.Fatal("self-test child reported no opened session (positive existence first)")
	}
	after := ListSessions(t)
	t.Logf("opened=%v before=%v after=%v", opened, before, after)
	for _, n := range opened {
		if slices.Contains(after, n) {
			t.Errorf("session %s opened by the failing child survived", n)
		}
	}
	if !slices.Contains(after, sentinel) {
		t.Errorf("sentinel %s did not survive", sentinel)
	}
	if strings.Join(before, ",") != strings.Join(after, ",") {
		t.Errorf("moai-ptycap- session set changed: before=%v after=%v", before, after)
	}
	checkHome()
}

// TestPtyCapture_SkipWithoutGate — AC-ITI-019 (a) over this package's seven
// capture tests.
func TestPtyCapture_SkipWithoutGate(t *testing.T) {
	Gate(t)
	AssertSkipWithoutGate(t, BuildChild(t, "."), "^TestPtyCapture", 7)
}

// TestPtyCapture_FailWithoutTmux — AC-ITI-019 (b) over this package's five
// TestPtyCapture_* tests.
func TestPtyCapture_FailWithoutTmux(t *testing.T) {
	Gate(t)
	AssertFailWithoutTmux(t, BuildChild(t, "."), "^TestPtyCapture_", 5)
}
