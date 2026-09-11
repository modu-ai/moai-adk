package wizard

// Self-checks for the pty capture harness (ptycap_harness_test.go), the real
// HOME watch list (ptycap_homewatch_test.go), and the render helpers
// (ptycap_render_test.go). The harness has to be trustworthy before any
// rendering verdict is read from it: it must fail when it cannot run, fail
// when an anchor never appears, never leave a session behind, never kill a
// session it did not open, and observe — not assume — the child environment.
//
// Tests named TestPtyCapture_* need tmux and run only under MOAI_PTY_CAPTURE=1.
// Everything else here is pure and runs in every `go test`.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

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
			if got := stripANSI(tc.in); got != tc.want {
				t.Errorf("stripANSI(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestPtycapDisplayColumn(t *testing.T) {
	line := "┃ Korean (한국어) - 한국어"
	// "┃ " = 2 columns, "Korean (" = 8, "한국어" = 3 wide runes = 6, ") " = 2.
	if got := displayColumn(line, "- 한국어"); got != 18 {
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
	if got := displayColumn(line, "absent"); got != -1 {
		t.Errorf("display column of an absent substring = %d, want -1", got)
	}
	if got := displayColumn("Japanese (日本語) - 日本語", "Japanese"); got != 0 {
		t.Errorf("display column at line start = %d, want 0", got)
	}
}

func TestPtycapCompareGolden(t *testing.T) {
	dir := t.TempDir()

	if err := compareGolden(dir, "missing", "frame\n", false); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("absent golden without update: want a 'missing' error, got %v", err)
	}
	if err := compareGolden(dir, "fresh", "line one\nline two\n", true); err != nil {
		t.Fatalf("update of an absent golden: %v", err)
	}
	if b, err := os.ReadFile(filepath.Join(dir, "fresh.golden")); err != nil || string(b) != "line one\nline two\n" {
		t.Fatalf("update wrote %q (err %v)", b, err)
	}
	if err := compareGolden(dir, "fresh", "line one\nline two\n", false); err != nil {
		t.Fatalf("identical frame: %v", err)
	}
	err := compareGolden(dir, "fresh", "line one\nline 2\n", false)
	if err == nil {
		t.Fatal("changed frame compared equal")
	}
	if !strings.Contains(err.Error(), "line 2") || !strings.Contains(err.Error(), "line two") {
		t.Errorf("mismatch error does not show both sides of the changed line: %v", err)
	}
	if err := compareGolden(dir, "fresh", "line one\nline 2\n", true); err != nil {
		t.Fatalf("update of a changed golden: %v", err)
	}
	if err := compareGolden(dir, "fresh", "line one\nline 2\n", false); err != nil {
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
			before, err := homeWatchSnapshot(home, []string{key})
			if err != nil {
				t.Fatal(err)
			}
			// Seeded: W1 default preferences, W2 p1, W3, the W4 directory, W5 .zshrc.
			if before.existing() != 5 {
				t.Fatalf("seeded fake HOME shows %d existing watch entries, want 5 (positive existence first)", before.existing())
			}
			tc.mutate(t, home)
			after, err := homeWatchSnapshot(home, []string{key})
			if err != nil {
				t.Fatal(err)
			}
			changed := homeWatchDiff(before, after)
			t.Logf("result %q: entries=%d existing=%d changed=%v", tc.name, len(after), after.existing(), changed)
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
	for _, name := range ptycapKanbanVars {
		inList[name] = true
	}
	for _, m := range declared {
		if !inList[m[1]] {
			t.Errorf("envkeys.go declares %s but the child scrub list does not carry it", m[1])
		}
	}
	if len(ptycapKanbanVars) != len(declared) {
		t.Errorf("scrub list carries %d MOAI_KANBAN vars, envkeys.go declares %d", len(ptycapKanbanVars), len(declared))
	}
}

func TestPtycapChildEnv_Values(t *testing.T) {
	c := ptycapNewCase(t, "")
	env := map[string]string{}
	var order []string
	for _, kv := range c.childEnv() {
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
		ptycapEnvOutEnv:           filepath.Join(c.Dir, "child-env.txt"),
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
	if env[ptycapCanaryEnv] == "" || env[ptycapCanaryEnv] != c.Canary {
		t.Errorf("canary %q does not match the case canary %q", env[ptycapCanaryEnv], c.Canary)
	}
	for _, k := range ptycapKanbanVars {
		v, ok := env[k]
		if !ok {
			t.Errorf("child env does not carry %s", k)
		} else if v != "" {
			t.Errorf("child env %s=%q, want empty", k, v)
		}
	}
	if other := ptycapNewCase(t, ""); other.Canary == c.Canary {
		t.Error("two cases share a canary")
	}
}

func TestPtycapCheckChildEnvRecord(t *testing.T) {
	c := ptycapNewCase(t, "")
	realHome, realMoai := "/real/home", "/real/home/.moai"
	good := func() map[string]string {
		m := map[string]string{
			"HOME":                    filepath.Join(c.Dir, "home"),
			config.EnvHome:            filepath.Join(c.Dir, "moai-home"),
			config.EnvClaudeConfigDir: "",
			"TERM":                    "xterm-256color",
			ptycapCanaryEnv:           c.Canary,
		}
		for _, k := range ptycapKanbanVars {
			m[k] = ""
		}
		return m
	}
	render := func(m map[string]string) string {
		var b strings.Builder
		for _, k := range ptycapRecordedVars() {
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
		"wrong canary":         func(m map[string]string) { m[ptycapCanaryEnv] = "not-the-canary" },
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
	a := ptycapSessionName("TestSomething/sub case")
	b := ptycapSessionName("TestSomething/sub case")
	for _, n := range []string{a, b, ptycapSessionName("sentinel")} {
		if !re.MatchString(n) {
			t.Errorf("session name %q does not match %s", n, re)
		}
	}
	if a == b {
		t.Errorf("two names for the same label collide: %q", a)
	}
	if !strings.HasPrefix(ptycapSessionName("sentinel"), ptycapSessionPrefix+"sentinel-") {
		t.Error("sentinel name must read moai-ptycap-sentinel-<rand>")
	}
}

// ---------------------------------------------------------------------------
// Child helpers (run only as a child of a harness test)
// ---------------------------------------------------------------------------

// TestPtyCaptureChild is the program a pty session runs. It records its
// effective environment first, then runs the named case on the real TTY.
func TestPtyCaptureChild(t *testing.T) {
	name := os.Getenv(ptycapChildEnv)
	if name == "" {
		t.Skip("runs only inside a pty capture session")
	}
	if err := ptycapRecordEnv(); err != nil {
		t.Fatalf("record child env: %v", err)
	}
	switch name {
	case ptycapCaseInitFirstPage:
		cwd, _ := os.Getwd()
		form := buildUnifiedForm(InitQuestions(cwd), &WizardResult{}, "en")
		err := form.Run()
		t.Logf("form returned: %v", err)
	default:
		t.Fatalf("unknown child case %q", name)
	}
}

// TestPtyCaptureSelfTestChild opens a real harness session and then fails on
// purpose (fail) or waits for an anchor that never renders (timeout). A parent
// test runs it as a separate `go test` process and inspects what it left.
func TestPtyCaptureSelfTestChild(t *testing.T) {
	mode := os.Getenv(ptycapSelfTestEnv)
	if mode == "" {
		t.Skip("runs only as a self-test child")
	}
	ptycapGate(t)
	bin := ptycapBuildChild(t)
	c := ptycapNewCase(t, os.Getenv(ptycapCaseRootEnv))
	s := ptycapStart(t, c, bin, ptycapCaseInitFirstPage)
	switch mode {
	case "fail":
		s.WaitFor("Select conversation language", ptycapAnchorTimeout)
		t.Fatal("forced failure after the session opened")
	case "timeout":
		s.WaitFor(ptycapNeverAnchor, ptycapSelfTestTimeout)
		t.Fatal("unreachable: WaitFor must fail on a missing anchor")
	default:
		t.Fatalf("unknown self-test mode %q", mode)
	}
}

// ---------------------------------------------------------------------------
// pty self-checks (need tmux, gated by MOAI_PTY_CAPTURE=1)
// ---------------------------------------------------------------------------

// TestPtyCapture_NormalRun — AC-ITI-020 (1) plus the effective-environment
// observation and the real HOME watch comparison.
func TestPtyCapture_NormalRun(t *testing.T) {
	ptycapGate(t)
	bin := ptycapBuildChild(t)
	c := ptycapNewCase(t, "")
	checkHome := ptycapWatchRealHome(t, c.Dir)

	sentinel := ptycapOpenSentinel(t)
	before := ptycapListSessions(t)
	if !containsString(before, sentinel) {
		t.Fatalf("sentinel %s not listed before the run: %v", sentinel, before)
	}

	s := ptycapStart(t, c, bin, ptycapCaseInitFirstPage)
	frame := s.WaitFor("Select conversation language", ptycapAnchorTimeout)
	ptycapVerifyChildEnv(t, c)
	requireLines(t, frame, "Select conversation language",
		"English", "Korean (한국어)", "Japanese (日本語)", "Chinese (中文)")
	path := ptycapExport(t, "normal-run-init-first-page", frame)
	if b, err := os.ReadFile(path); err != nil || !strings.Contains(string(b), "Select conversation language") {
		t.Fatalf("capture file %s does not carry the anchor (err %v)", path, err)
	}
	s.SendKeys("C-c")
	s.Close()

	after := ptycapListSessions(t)
	t.Logf("moai-ptycap- sessions before=%v after=%v", before, after)
	if strings.Join(before, ",") != strings.Join(after, ",") {
		t.Errorf("moai-ptycap- session set changed: before=%v after=%v", before, after)
	}
	checkHome()
}

func TestPtyCapture_ForcedFailure(t *testing.T) {
	ptycapGate(t)
	ptycapRunSelfTest(t, "fail", "forced failure after the session opened")
}

func TestPtyCapture_ForcedTimeout(t *testing.T) {
	ptycapGate(t)
	ptycapRunSelfTest(t, "timeout", "not visible within")
}

// ptycapRunSelfTest runs TestPtyCaptureSelfTestChild in a separate process and
// checks AC-ITI-020 (2)/(3): the child reports FAIL with the expected message,
// the session it opened is gone, the sentinel survives, and the real HOME
// watch list is unchanged.
func ptycapRunSelfTest(t *testing.T, mode, wantMsg string) {
	t.Helper()
	bin := ptycapBuildChild(t)
	caseRoot := t.TempDir()
	checkHome := ptycapWatchRealHome(t, filepath.Join(caseRoot, "case"))

	sentinel := ptycapOpenSentinel(t)
	before := ptycapListSessions(t)
	if !containsString(before, sentinel) {
		t.Fatalf("sentinel %s not listed before the run: %v", sentinel, before)
	}

	out, err := ptycapRunChild(t, bin, ptycapSubprocessEnv(t, map[string]string{
		ptycapGateEnv:     "1",
		ptycapSelfTestEnv: mode,
		ptycapCaseRootEnv: caseRoot,
	}), "-test.run", "^TestPtyCaptureSelfTestChild$", "-test.v", "-test.timeout", "60s")
	t.Logf("self-test child (%s) exit=%v output:\n%s", mode, err, out)
	ptycapExport(t, "selftest-"+mode+"-child-output", out)

	if err == nil {
		t.Errorf("self-test child exited 0; want a failing exit")
	}
	if !strings.Contains(out, "--- FAIL: TestPtyCaptureSelfTestChild") {
		t.Errorf("self-test child output carries no FAIL line")
	}
	if !strings.Contains(out, wantMsg) {
		t.Errorf("self-test child output lacks %q", wantMsg)
	}
	opened := ptycapSessionsNamedIn(out)
	if len(opened) == 0 {
		t.Fatal("self-test child reported no opened session (positive existence first)")
	}
	after := ptycapListSessions(t)
	t.Logf("opened=%v before=%v after=%v", opened, before, after)
	for _, n := range opened {
		if containsString(after, n) {
			t.Errorf("session %s opened by the failing child survived", n)
		}
	}
	if !containsString(after, sentinel) {
		t.Errorf("sentinel %s did not survive", sentinel)
	}
	if strings.Join(before, ",") != strings.Join(after, ",") {
		t.Errorf("moai-ptycap- session set changed: before=%v after=%v", before, after)
	}
	checkHome()
}

// TestPtyCapture_SkipWithoutGate — AC-ITI-019 (a).
func TestPtyCapture_SkipWithoutGate(t *testing.T) {
	ptycapGate(t)
	bin := ptycapBuildChild(t)
	sentinel := ptycapOpenSentinel(t)
	before := ptycapListSessions(t)
	if !containsString(before, sentinel) {
		t.Fatalf("sentinel %s not listed before the run: %v", sentinel, before)
	}

	out, err := ptycapRunChild(t, bin, ptycapSubprocessEnv(t, map[string]string{ptycapGateEnv: ""}),
		"-test.run", "^TestPtyCapture", "-test.v")
	t.Logf("ungated child exit=%v output:\n%s", err, out)
	ptycapExport(t, "ac019a-ungated-child-output", out)
	if err != nil {
		t.Errorf("ungated child exited non-zero: %v", err)
	}
	results := topLevelResults(out)
	if len(results) < 7 {
		t.Fatalf("ungated child reported %d top-level results, want >= 7 capture tests: %v", len(results), results)
	}
	for name, res := range results {
		if res != "SKIP" {
			t.Errorf("ungated: %s reported %s, want SKIP", name, res)
		}
	}
	after := ptycapListSessions(t)
	if strings.Join(before, ",") != strings.Join(after, ",") {
		t.Errorf("moai-ptycap- session set changed: before=%v after=%v", before, after)
	}
}

// TestPtyCapture_FailWithoutTmux — AC-ITI-019 (b).
func TestPtyCapture_FailWithoutTmux(t *testing.T) {
	ptycapGate(t)
	bin := ptycapBuildChild(t)
	emptyPath := t.TempDir()
	out, err := ptycapRunChild(t, bin, ptycapSubprocessEnv(t, map[string]string{
		ptycapGateEnv: "1",
		"PATH":        emptyPath,
	}), "-test.run", "^TestPtyCapture_", "-test.v")
	t.Logf("tmux-less child exit=%v output:\n%s", err, out)
	ptycapExport(t, "ac019b-no-tmux-child-output", out)
	if err == nil {
		t.Errorf("tmux-less child exited 0; want a failing exit")
	}
	results := topLevelResults(out)
	if len(results) < 5 {
		t.Fatalf("tmux-less child reported %d top-level results, want >= 5: %v", len(results), results)
	}
	for name, res := range results {
		if res != "FAIL" {
			t.Errorf("tmux absent: %s reported %s, want FAIL", name, res)
		}
	}
	if !strings.Contains(out, "tmux") {
		t.Error("failure output does not name tmux")
	}
}

// ---------------------------------------------------------------------------
// small helpers used by the self-checks
// ---------------------------------------------------------------------------

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
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
