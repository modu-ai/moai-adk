package tui

// SPEC-TUI-CHECKLIST-001 (card t1969) M2: renderer tests for the group
// checklist. Internal package so a test can inject a probe catalogue language
// and force the colour profile; neither is reachable from tui_test.

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"

	"github.com/charmbracelet/colorprofile"
)

// sgrRe matches an SGR colour/attribute sequence (not the erase-line CSI).
var sgrRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

// checklistGolden compares got to testdata/<name>.golden, regenerating the file
// when UPDATE_GOLDEN=1.
func checklistGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden %s: %v", path, err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (run with UPDATE_GOLDEN=1 to generate)", path, err)
	}
	if got != string(want) {
		t.Errorf("checklist output mismatch for %s\ngot:\n%s\nwant:\n%s", name, got, want)
	}
}

// fixedClock returns a clock whose first reading is the base instant and every
// later reading is base+elapsed, so the elapsed field is deterministic.
func fixedClock(elapsed time.Duration) func() time.Time {
	base := time.Unix(1_700_000_000, 0)
	n := 0
	return func() time.Time {
		n++
		if n == 1 {
			return base
		}
		return base.Add(elapsed)
	}
}

// finalScreen reduces raw renderer output to what a terminal shows once every
// in-place redraw has landed: on each line, everything up to the last erase
// prefix is discarded.
func finalScreen(raw string) string {
	lines := strings.Split(raw, "\n")
	for i, l := range lines {
		if idx := strings.LastIndex(l, progressClearPrefix); idx >= 0 {
			lines[i] = l[idx+len(progressClearPrefix):]
		}
	}
	return strings.Join(lines, "\n")
}

func intp(n int) *int { return &n }

// newTestChecklist builds a renderer over a buffer with a deterministic clock.
func newTestChecklist(lang string, th *Theme, elapsed time.Duration) (*Checklist, *bytes.Buffer) {
	var buf bytes.Buffer
	c := NewChecklist(ChecklistOptions{Out: &buf, Theme: th, Lang: lang, Clock: fixedClock(elapsed)})
	return c, &buf
}

// runCleanInit drives the init-shaped scenario: five groups, twelve items, one
// spinner tick each, then the init-shaped summary.
func runCleanInit(c *Checklist) {
	c.Header("checklist.header.init", EnvSummary{Platform: "darwin/arm64", Version: "v3.3.0", Dir: "~/work/demo"})
	plan := [][2]string{
		{"templates", "detect"}, {"templates", "validate"}, {"templates", "deploy"},
		{"config", "autonomy"}, {"config", "profile"}, {"config", "optins"}, {"config", "harness"}, {"config", "globals"},
		{"hooks", "prepush"}, {"hooks", "precommit"},
		{"mcp", "entry"},
		{"verify", "manifest"},
	}
	for _, p := range plan {
		it := c.Item(p[0], p[1])
		it.Update()
		it.Detail("collapsed detail for " + p[1])
		it.Done()
	}
	c.Summary(ChecklistSummary{
		CreatedFiles: intp(2), CreatedDirs: intp(13),
		DeployMode: "plugin", NextSteps: "/moai plan",
	})
}

// withProbeLang registers a test-only catalogue language so a scenario can use
// group and item keys no production catalogue defines.
func withProbeLang(t *testing.T) string {
	t.Helper()
	const lang = "probe"
	catalogMu.Lock()
	catalogCache[lang] = catalog{
		"checklist.group.alpha":      "Alpha",
		"checklist.group.beta":       "Beta",
		"checklist.item.alpha.one":   "Alpha one",
		"checklist.item.alpha.two":   "Alpha two",
		"checklist.item.beta.one":    "Beta one",
		"checklist.item.beta.two":    "Beta two",
		"checklist.header.probe":     "Probe run",
		"checklist.status.probe.why": "because",
	}
	catalogMu.Unlock()
	t.Cleanup(func() {
		catalogMu.Lock()
		delete(catalogCache, lang)
		catalogMu.Unlock()
	})
	return lang
}

// forceProfile pins the package colour profile for one test, so "no SGR" is a
// meaningful assertion: under go test the detected profile is NoTTY and strips
// every sequence regardless of what the renderer emitted.
func forceProfile(t *testing.T, p colorprofile.Profile) {
	t.Helper()
	_ = outputProfile() // run the one-time detection so the restore value is real
	orig := outputProfileVal
	outputProfileVal = p
	t.Cleanup(func() { outputProfileVal = orig })
}

var forceProfileMu sync.Mutex // documents that profile-forcing tests never run in parallel

// AC-TCL-001: a clean init renders the golden checklist within the screen
// budget and shows no raw catalogue key.
func TestChecklistGoldenClean(t *testing.T) {
	forceProfileMu.Lock()
	defer forceProfileMu.Unlock()
	c, buf := newTestChecklist("en", nil, 1400*time.Millisecond)
	runCleanInit(c)
	screen := finalScreen(buf.String())
	checklistGolden(t, "checklist-init-clean", screen)

	if strings.Contains(screen, "checklist.") {
		t.Errorf("raw catalogue key rendered:\n%s", screen)
	}
	lines := strings.Split(strings.TrimRight(screen, "\n"), "\n")
	if len(lines) > 24 {
		t.Errorf("clean init drew %d lines, budget is 24", len(lines))
	}
	for _, want := range []string{"\u25c6 Templates", "  \u2713 Detect project structure", "  \u2713 Record manifest", "2 files created", "13 directories", "0 warnings", "1.4s", "Deploy mode: plugin"} {
		if !strings.Contains(screen, want) {
			t.Errorf("screen missing %q:\n%s", want, screen)
		}
	}
	if strings.Contains(screen, "collapsed detail") {
		t.Errorf("a successful item expanded its detail:\n%s", screen)
	}
}

// AC-TCL-002: only the running line is redrawn; a finished line is never
// touched again.
func TestChecklistRedrawOnlyRunningLine(t *testing.T) {
	c, buf := newTestChecklist("en", nil, time.Second)
	c.Header("checklist.header.init", EnvSummary{})
	it := c.Item("templates", "detect")
	it.Update()
	it.Update()
	it.Update()
	it.Done()
	raw := buf.String()

	if got := strings.Count(raw, progressClearPrefix); got != 4 {
		t.Errorf("clear prefixes = %d, want 4 (3 Update + 1 Done)\nraw=%q", got, raw)
	}
	for i := 0; ; {
		j := strings.Index(raw[i:], progressClearPrefix)
		if j < 0 {
			break
		}
		at := i + j
		if at > 0 && raw[at-1] == '\n' {
			t.Errorf("erase prefix at %d follows a newline, so it would erase a finished line\nraw=%q", at, raw)
		}
		i = at + len(progressClearPrefix)
	}
	if strings.Count(raw, "Detect project structure\n") != 1 {
		t.Errorf("finished line must be written exactly once with a newline\nraw=%q", raw)
	}

	// A second item must not touch the first one's finished line.
	before := buf.Len()
	next := c.Item("templates", "validate")
	next.Update()
	next.Done()
	added := buf.String()[before:]
	if strings.Contains(added, "Detect project structure") {
		t.Errorf("a later item redrew an earlier line: %q", added)
	}
}

// AC-TCL-003: a failing item expands every detail line recorded under it and
// the failure message; a successful sibling stays collapsed.
func TestChecklistFailExpandsDetail(t *testing.T) {
	lang := withProbeLang(t)
	c, buf := newTestChecklist(lang, nil, 2*time.Second)
	c.Header("checklist.header.probe", EnvSummary{})
	ok := c.Item("alpha", "one")
	ok.Detail("sibling detail stays collapsed")
	ok.Done()
	bad := c.Item("alpha", "two")
	bad.Detail("first detail")
	bad.Detail("second detail")
	bad.Fail("boom happened")
	screen := finalScreen(buf.String())
	checklistGolden(t, "checklist-probe-fail", screen)

	want := "  \u2717 Alpha two\n    first detail\n    second detail\n    boom happened\n"
	if !strings.Contains(screen, want) {
		t.Errorf("failed item block missing:\nwant substring %q\ngot:\n%s", want, screen)
	}
	if strings.Contains(screen, "sibling detail") {
		t.Errorf("successful sibling expanded:\n%s", screen)
	}
	if strings.Contains(screen, "Beta") {
		t.Errorf("a group never opened was drawn:\n%s", screen)
	}
}

// AC-TCL-004: a warned item ends with the warning glyph and expands only its
// own warnings; keep lines stay expanded without changing the glyph.
func TestChecklistWarnExpandsOnlyWarned(t *testing.T) {
	lang := withProbeLang(t)
	c, buf := newTestChecklist(lang, nil, time.Second)
	c.Header("checklist.header.probe", EnvSummary{})
	a := c.Item("alpha", "one")
	a.Detail("detail one")
	a.Done()
	b := c.Item("alpha", "two")
	b.Detail("detail two")
	b.Warn("careful about two")
	b.Done()
	k := c.Item("beta", "one")
	k.Keep("preserved line stays")
	k.Done()
	screen := finalScreen(buf.String())
	checklistGolden(t, "checklist-probe-warn", screen)

	for _, want := range []string{"  ! Alpha two\n    careful about two\n", "  \u2713 Alpha one\n", "  \u2713 Beta one\n    preserved line stays\n"} {
		if !strings.Contains(screen, want) {
			t.Errorf("missing %q in:\n%s", want, screen)
		}
	}
	if strings.Contains(screen, "detail one") || strings.Contains(screen, "detail two") {
		t.Errorf("collapsed detail leaked:\n%s", screen)
	}
}

// AC-TCL-005: each summary field appears exactly once, in both shapes.
func TestChecklistSummaryFields(t *testing.T) {
	t.Run("update_shape", func(t *testing.T) {
		c, buf := newTestChecklist("en", nil, 1500*time.Millisecond)
		c.Summary(ChecklistSummary{
			Changed: intp(7), Preserved: intp(3), Warnings: 1,
			BackupPath: ".moai/backups/20261027T000000Z", RecoverCmd: "moai restore --from latest",
		})
		out := finalScreen(buf.String())
		for _, want := range []string{"7 changed", "3 preserved", "1 warnings", "1.5s", "Backup: .moai/backups/20261027T000000Z", "Recover: moai restore --from latest"} {
			if n := strings.Count(out, want); n != 1 {
				t.Errorf("%q appears %d times, want 1:\n%s", want, n, out)
			}
		}
		if strings.Contains(out, "files created") || strings.Contains(out, "directories") {
			t.Errorf("init-shaped counts leaked into an update summary:\n%s", out)
		}
	})
	t.Run("init_shape", func(t *testing.T) {
		c, buf := newTestChecklist("en", nil, 1500*time.Millisecond)
		c.Summary(ChecklistSummary{
			CreatedFiles: intp(2), CreatedDirs: intp(13), DeployMode: "plugin", NextSteps: "/moai plan",
			Warnings: 1, OutsideWarnings: []string{"outside warning text"},
		})
		out := finalScreen(buf.String())
		for _, want := range []string{"2 files created", "13 directories", "Deploy mode: plugin", "Next steps: /moai plan", "outside warning text", "Warnings:", "1.5s"} {
			if n := strings.Count(out, want); n != 1 {
				t.Errorf("%q appears %d times, want 1:\n%s", want, n, out)
			}
		}
	})
	t.Run("failed_headline", func(t *testing.T) {
		c, buf := newTestChecklist("en", nil, time.Second)
		c.Summary(ChecklistSummary{Failed: true})
		if out := finalScreen(buf.String()); !strings.Contains(out, "  Failed \u00b7 0 warnings \u00b7 1.0s") {
			t.Errorf("failed headline wrong:\n%s", out)
		}
	})
}

// AC-TCL-006: the monochrome theme and the NO_COLOR-style profiles produce no
// SGR colour sequence; the colour theme under full colour does (positive
// control, so the assertion cannot pass vacuously).
func TestChecklistNoColorHasNoSGR(t *testing.T) {
	forceProfileMu.Lock()
	defer forceProfileMu.Unlock()

	strip := func(s string) string { return strings.ReplaceAll(s, progressClearPrefix, "") }

	t.Run("positive_control", func(t *testing.T) {
		forceProfile(t, colorprofile.TrueColor)
		th := LightTheme()
		c, buf := newTestChecklist("en", &th, time.Second)
		runCleanInit(c)
		if !sgrRe.MatchString(strip(buf.String())) {
			t.Fatalf("control run produced no SGR; the no-colour assertion below would be vacuous\nraw=%q", buf.String())
		}
	})
	t.Run("monochrome_theme", func(t *testing.T) {
		forceProfile(t, colorprofile.TrueColor)
		th := MonochromeTheme()
		c, buf := newTestChecklist("en", &th, 1400*time.Millisecond)
		runCleanInit(c)
		if m := sgrRe.FindString(strip(buf.String())); m != "" {
			t.Errorf("monochrome theme emitted SGR %q on a clean run\nraw=%q", m, buf.String())
		}
		checklistGolden(t, "checklist-init-nocolor", finalScreen(buf.String()))

		// The failure and warning paths must be as plain as the clean one.
		cf, bf := newTestChecklist("en", &th, time.Second)
		cf.Header("checklist.header.init", EnvSummary{})
		w := cf.Item("config", "autonomy")
		w.Warn("a warning")
		w.Done()
		cf.Item("templates", "deploy").Fail("a failure")
		cf.Summary(ChecklistSummary{Failed: true, Warnings: 1, OutsideWarnings: []string{"outside"}})
		if m := sgrRe.FindString(strip(bf.String())); m != "" {
			t.Errorf("monochrome theme emitted SGR %q on the fail/warn path\nraw=%q", m, bf.String())
		}
	})
	t.Run("ascii_profile_has_no_colour_params", func(t *testing.T) {
		forceProfile(t, colorprofile.Ascii)
		th := LightTheme()
		c, buf := newTestChecklist("en", &th, time.Second)
		runCleanInit(c)
		// The ASCII profile leaves bare resets (ESC [ m); those carry no colour.
		plain := strings.ReplaceAll(strip(buf.String()), "\x1b[m", "")
		if m := sgrRe.FindString(plain); m != "" {
			t.Errorf("ascii profile emitted colour SGR %q\nraw=%q", m, buf.String())
		}
	})
}

// AC-TCL-007: with reduced motion the spinner is the static glyph; the
// control run without it shows the braille frame.
func TestChecklistReducedMotion(t *testing.T) {
	braille := func(s string) bool {
		for _, r := range s {
			if r >= 0x2800 && r <= 0x28ff {
				return true
			}
		}
		return false
	}
	run := func() string {
		c, buf := newTestChecklist("en", nil, time.Second)
		c.Header("checklist.header.init", EnvSummary{})
		it := c.Item("templates", "detect")
		it.Update()
		it.Done()
		return buf.String()
	}

	t.Setenv("MOAI_REDUCED_MOTION", "1")
	reduced := run()
	if braille(reduced) {
		t.Errorf("reduced motion drew a braille frame: %q", reduced)
	}
	if !strings.Contains(reduced, "\u25cf Detect project structure") {
		t.Errorf("reduced motion drew no static glyph: %q", reduced)
	}
	checklistGolden(t, "checklist-init-reduced", finalScreen(reduced))

	t.Setenv("MOAI_REDUCED_MOTION", "")
	if control := run(); !braille(control) {
		t.Errorf("control run without reduced motion drew no braille frame: %q", control)
	}
}

// AC-TCL-009: Korean renders from the ko catalogue; an unsupported language
// renders the English catalogue byte for byte.
func TestChecklistKoreanAndFallback(t *testing.T) {
	cko, bko := newTestChecklist("ko", nil, 1400*time.Millisecond)
	runCleanInit(cko)
	ko := finalScreen(bko.String())
	checklistGolden(t, "checklist-init-ko", ko)
	for _, want := range []string{"\u25c6 \ud15c\ud50c\ub9bf", "\ud504\ub85c\uc81d\ud2b8 \uad6c\uc870 \uac10\uc9c0", "\ud30c\uc77c 2\uac1c \uc0dd\uc131"} {
		if !strings.Contains(ko, want) {
			t.Errorf("ko screen missing %q:\n%s", want, ko)
		}
	}
	if strings.Contains(ko, "checklist.") {
		t.Errorf("raw catalogue key in ko screen:\n%s", ko)
	}

	cen, ben := newTestChecklist("en", nil, 1400*time.Millisecond)
	runCleanInit(cen)
	cxx, bxx := newTestChecklist("xx", nil, 1400*time.Millisecond)
	runCleanInit(cxx)
	if ben.String() != bxx.String() {
		t.Errorf("unsupported language output differs from en\nen=%q\nxx=%q", ben.String(), bxx.String())
	}
	if ben.String() == bko.String() {
		t.Errorf("ko output is identical to en; the catalogue was not consulted")
	}
}

// AC-TCL-023: only whitelisted glyphs appear in a rendered screen, and no
// checked-in checklist golden contains a raw catalogue key.
func TestChecklistGlyphWhitelist(t *testing.T) {
	lang := withProbeLang(t)
	allowed := map[rune]bool{'\u2713': true, '\u2717': true, '\u00b7': true, '\u25cf': true, '\u25cb': true, '\u25c6': true, '\u25c7': true}

	var screens []string
	for _, l := range []string{"en", "ko", "xx"} {
		c, buf := newTestChecklist(l, nil, time.Second)
		runCleanInit(c)
		screens = append(screens, finalScreen(buf.String()))
	}
	c, buf := newTestChecklist(lang, nil, time.Second)
	c.Header("checklist.header.probe", EnvSummary{Platform: "p", Version: "v", Dir: "d"})
	a := c.Item("alpha", "one")
	a.Warn("warned")
	a.Done()
	b := c.Item("alpha", "two")
	b.Skip("checklist.status.skipped")
	d := c.Item("beta", "one")
	d.Fail("failed")
	c.Summary(ChecklistSummary{Failed: true, OutsideWarnings: []string{"out"}})
	screens = append(screens, finalScreen(buf.String()))

	for _, s := range screens {
		for _, r := range s {
			if r < 0x80 || unicode.IsLetter(r) || unicode.IsNumber(r) || r == '\n' {
				continue
			}
			if !allowed[r] {
				t.Errorf("non-whitelisted rune %U (%q) in screen:\n%s", r, r, s)
			}
			if r == '\u27a4' || r == '\u2013' {
				t.Errorf("forbidden rune %U in screen", r)
			}
		}
	}

	goldens, err := filepath.Glob(filepath.Join("testdata", "checklist-*.golden"))
	if err != nil || len(goldens) == 0 {
		t.Fatalf("no checklist goldens found (err=%v, n=%d)", err, len(goldens))
	}
	for _, g := range goldens {
		data, err := os.ReadFile(g)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(data, []byte("checklist.")) {
			t.Errorf("golden %s contains a raw catalogue key", g)
		}
	}
}

// Skip and implicit-close behaviour backing REQ-TCL-004/-005.
func TestChecklistSkipAndImplicitClose(t *testing.T) {
	c, buf := newTestChecklist("en", nil, time.Second)
	c.Header("checklist.header.init", EnvSummary{Platform: "darwin/arm64"})
	c.Item("hooks", "prepush").Skip("checklist.status.skipped")
	c.Item("mcp", "entry") // left open; the next Item closes it as done
	c.Item("verify", "manifest")
	c.Summary(ChecklistSummary{}) // closes manifest
	out := finalScreen(buf.String())

	if !strings.Contains(out, "\u25cb pre-push hook") || !strings.Contains(out, "skipped") {
		t.Errorf("skip line missing:\n%s", out)
	}
	for _, want := range []string{"  \u2713 MCP server entry\n", "  \u2713 Record manifest\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("implicit close missing %q:\n%s", want, out)
		}
	}
	if strings.HasPrefix(out, "  ") || !strings.HasPrefix(out, "\u25c6 MoAI-ADK init\n  darwin/arm64\n") {
		t.Errorf("header lines wrong:\n%s", out)
	}
}

// A run that opens no item and records no header draws nothing but the summary.
func TestChecklistNoHeaderWhenNeverRecorded(t *testing.T) {
	c, buf := newTestChecklist("en", nil, time.Second)
	c.Summary(ChecklistSummary{})
	if strings.Contains(buf.String(), "\u25c6") {
		t.Errorf("header drawn without Header(): %q", buf.String())
	}
}

// A second terminal transition is a programmer error and panics, matching
// ProgressLineHandle.
func TestChecklistDoubleTerminalPanics(t *testing.T) {
	c, _ := newTestChecklist("en", nil, time.Second)
	it := c.Item("templates", "detect")
	it.Done()
	defer func() {
		if recover() == nil {
			t.Error("second Done did not panic")
		}
	}()
	it.Fail("again")
}

// Calls on a finished item after the terminal transition record nothing.
func TestChecklistRecordAfterDoneIsIgnored(t *testing.T) {
	c, buf := newTestChecklist("en", nil, time.Second)
	it := c.Item("templates", "detect")
	it.Done()
	before := buf.Len()
	it.Update()
	it.Detail("late")
	it.Warn("late")
	it.Keep("late")
	if buf.Len() != before {
		t.Errorf("finished item wrote %q", buf.String()[before:])
	}
}

// Open reports the currently open item so a command-side session can route
// captured bytes to it; it is nil before the first Item and after every close.
func TestChecklistOpenAccessor(t *testing.T) {
	c, _ := newTestChecklist("en", nil, time.Second)
	if c.Open() != nil {
		t.Fatal("Open() non-nil before any Item")
	}
	first := c.Item("templates", "detect")
	if c.Open() != first {
		t.Fatal("Open() is not the item just opened")
	}
	second := c.Item("templates", "validate") // implicit Done of first
	if c.Open() != second {
		t.Fatal("Open() did not follow the implicit close")
	}
	second.Fail("boom")
	if c.Open() != nil {
		t.Fatal("Open() non-nil after Fail")
	}
	c.Item("config", "autonomy").Skip("checklist.status.skipped")
	if c.Open() != nil {
		t.Fatal("Open() non-nil after Skip")
	}
	last := c.Item("config", "profile")
	c.Summary(ChecklistSummary{})
	if c.Open() != nil || !last.done {
		t.Fatal("Summary did not close the open item")
	}
}
