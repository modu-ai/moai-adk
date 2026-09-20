package cli

// pty capture cases for package cli (acceptance.md §B, design.md §11). The
// harness lives in internal/cli/ptycaptest; this file holds the one product-
// path child dispatcher of the package and the capture tests over it.
//
// Tests named TestPtyCapture_* need tmux and run only under MOAI_PTY_CAPTURE=1.
// Later cases (the AC-ITI-003 init case) add a case to TestPtyCaptureChild's
// switch rather than a second child function: ptycaptest.ChildTestName names
// exactly one function per package.

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode"

	"github.com/modu-ai/moai-adk/internal/cli/ptycaptest"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/pkg/version"
	"github.com/spf13/cobra"
)

const (
	// ptycapCaseDowngradeConfirm runs `moai update --version <older tag>` up to
	// the downgrade confirmation (runVersionBranch) on the TTY.
	ptycapCaseDowngradeConfirm = "downgrade-confirm"

	// The running version the child pretends to be and the older tag it asks
	// for, so isVersionDowngrade holds and the confirmation renders.
	ptycapDowngradeCurrent = "v9.9.9"
	ptycapDowngradeTarget  = "v1.0.0"

	// ptycapCaseInitFirstScreen runs the real `moai init` command path
	// (validateInitFlags + runInit on the package-level initCmd, no flag set)
	// in the case working directory on the TTY.
	ptycapCaseInitFirstScreen = "init-first-screen"

	// ptycapInitLangAnchor is the title of the first init wizard question
	// (wizard.InitQuestions). Reaching it proves the run passed the profile
	// block of runInit and entered the wizard block.
	ptycapInitLangAnchor = "Select conversation language"
	// ptycapInitCancelAnchor is what runInit prints on wizard.ErrCancelled.
	ptycapInitCancelAnchor = "Initialization cancelled."
)

// ptycapNoNetwork refuses every request: the child never reaches a server.
type ptycapNoNetwork struct{}

func (ptycapNoNetwork) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("ptycap child: network blocked")
}

// ptycapBlockNetwork closes every network seam the child's product paths can
// reach: the deferred update notice (init) and the --version install client.
func ptycapBlockNetwork(t *testing.T) {
	deferredUpdateEnabled = func(*cobra.Command) bool { return false }
	deferredUpdateCheck = func(*cobra.Command, *Dependencies) *deferredUpdateResult { return nil }
	versionInstallHTTPClient = &http.Client{Transport: ptycapNoNetwork{}}
	t.Setenv(config.EnvUpdateURL, "")
}

// TestPtyCaptureChild is the program a pty session runs. It records its
// effective environment first, blocks the network seams, then runs the named
// case on the real TTY.
func TestPtyCaptureChild(t *testing.T) {
	name := os.Getenv(ptycaptest.ChildEnv)
	if name == "" {
		t.Skip("runs only inside a pty capture session")
	}
	if err := ptycaptest.RecordEnv(); err != nil {
		t.Fatalf("record child env: %v", err)
	}
	ptycapBlockNetwork(t)
	switch name {
	case ptycapCaseDowngradeConfirm:
		origVersion := version.Version
		version.Version = ptycapDowngradeCurrent
		t.Cleanup(func() { version.Version = origVersion })
		// A binary path under the case temp dir: nothing is installed, the
		// blocked client fails the tag resolution first, but the path must never
		// be the test binary itself.
		versionInstallBinaryPath = filepath.Join(t.TempDir(), "moai")
		cmd := &cobra.Command{Use: "update"}
		cmd.Flags().Bool("yes", false, "")
		cmd.Flags().Bool("binary", true, "")
		cmd.SetOut(os.Stdout)
		err := runVersionBranch(cmd, ptycapDowngradeTarget)
		t.Logf("runVersionBranch returned: %v", err)
	case ptycapCaseInitFirstScreen:
		// The real command path, in cobra's own order: initCmd is the
		// package-level command with every init flag registered and none set,
		// so this is plain `moai init` in the case working directory. No
		// further network seam is reached before the wizard — the deferred
		// update notice is already stubbed above, and everything runInit does
		// in front of the wizard (git lookup, remote detection, profile read)
		// is local.
		if err := validateInitFlags(initCmd, nil); err != nil {
			t.Fatalf("validateInitFlags: %v", err)
		}
		err := runInit(initCmd, nil)
		t.Logf("runInit returned: %v", err)
	default:
		t.Fatalf("unknown child case %q", name)
	}
}

// ptycapDowngradeFrame seeds a project whose language.yaml says ko, runs the
// downgrade-confirm child, waits for the target tag, verifies the child's
// effective environment, exports the capture as <export>.txt, and returns it.
// The real HOME watch comparison runs at the end of the calling test.
func ptycapDowngradeFrame(t *testing.T, export string) string {
	t.Helper()
	bin := ptycaptest.BuildChild(t, ".")
	c := ptycaptest.NewCase(t, "")
	checkHome := ptycaptest.WatchRealHome(t, c.Dir)
	t.Cleanup(checkHome)

	sections := filepath.Join(c.Dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "language.yaml"),
		[]byte("language:\n    conversation_language: ko\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	sentinel := ptycaptest.OpenSentinel(t)
	before := ptycaptest.ListSessions(t)
	if !slices.Contains(before, sentinel) {
		t.Fatalf("sentinel %s not listed before the run: %v", sentinel, before)
	}
	s := ptycaptest.Start(t, c, bin, ptycapCaseDowngradeConfirm)
	frame := s.WaitFor(ptycapDowngradeTarget, ptycaptest.AnchorTimeout)
	ptycaptest.VerifyChildEnv(t, c)
	ptycaptest.Export(t, export, frame)
	s.SendKeys("C-c")
	s.Close()
	after := ptycaptest.ListSessions(t)
	if strings.Join(before, ",") != strings.Join(after, ",") {
		t.Errorf("moai-ptycap- session set changed: before=%v after=%v", before, after)
	}
	return frame
}

// confirmButtonPairs are the Yes/No label pairs of the four locales
// (wizard UIStrings ConfirmYes/ConfirmNo), used only to find the button line.
var confirmButtonPairs = [][2]string{{"Yes", "No"}, {"예", "아니오"}, {"はい", "いいえ"}, {"是", "否"}}

// downgradeConfirmLines locates the title line (carries the target tag), the
// first content line after it (description), and the button line. Each index
// is -1 when absent.
func downgradeConfirmLines(frame string) (lines []string, title, desc, buttons int, pair [2]string) {
	lines = strings.Split(ptycaptest.StripANSI(frame), "\n")
	title, desc, buttons = -1, -1, -1
	for i, l := range lines {
		if title < 0 && strings.Contains(l, ptycapDowngradeTarget) {
			title = i
			continue
		}
		if title < 0 {
			continue
		}
		for _, p := range confirmButtonPairs {
			if strings.Contains(l, p[0]) && strings.Contains(l, p[1]) {
				buttons, pair = i, p
				break
			}
		}
		if buttons >= 0 {
			break
		}
		if desc < 0 && contentStart(l) >= 0 {
			desc = i
		}
	}
	return lines, title, desc, buttons, pair
}

// contentStart returns the display column of the first character of line
// that is neither a space nor the form border bar, or -1 for a line with no
// such character.
func contentStart(line string) int {
	for i, r := range line {
		if unicode.IsSpace(r) || r == '┃' || r == '│' {
			continue
		}
		return ptycaptest.DisplayColumn(line, line[i:])
	}
	return -1
}

// requireDowngradeConfirmLines is the positive-existence gate: title,
// description, and button lines each present, in that order.
func requireDowngradeConfirmLines(t *testing.T, frame string) (lines []string, title, desc, buttons int, pair [2]string) {
	t.Helper()
	lines, title, desc, buttons, pair = downgradeConfirmLines(frame)
	if title < 0 || desc < 0 || buttons < 0 {
		t.Fatalf("downgrade confirm lines: title=%d description=%d buttons=%d (want all >= 0); frame:\n%s", title, desc, buttons, frame)
	}
	t.Logf("title line %q · description line %q · button line %q", lines[title], lines[desc], lines[buttons])
	return lines, title, desc, buttons, pair
}

// TestPtyCapture_DowngradeConfirmLocalized — REQ-ITI-011 on the real TTY: in
// a project whose language.yaml says ko, the downgrade confirmation renders
// its title, description, buttons, and help action labels in Korean
// (AC-ITI-015 (a) reachability: the resolved locale's title is on screen).
func TestPtyCapture_DowngradeConfirmLocalized(t *testing.T) {
	ptycaptest.Gate(t)
	frame := ptycapDowngradeFrame(t, "downgrade-confirm-localized")
	lines, title, desc, buttons, _ := requireDowngradeConfirmLines(t, frame)

	wantTitle := "다운그레이드할까요? " + ptycapDowngradeCurrent + " → " + ptycapDowngradeTarget
	if !strings.Contains(lines[title], wantTitle) {
		t.Errorf("title line %q lacks the ko title %q", lines[title], wantTitle)
	}
	if !strings.Contains(lines[desc], "요청한 태그가 지금 실행 중인 버전보다 오래되었습니다.") {
		t.Errorf("description line %q is not the ko description", lines[desc])
	}
	if !strings.Contains(lines[buttons], "예") || !strings.Contains(lines[buttons], "아니오") {
		t.Errorf("button line %q lacks the ko labels 예 / 아니오", lines[buttons])
	}
	help := strings.Join(lines[buttons+1:], "\n")
	for _, want := range []string{"전환", "제출"} {
		if !strings.Contains(help, want) {
			t.Errorf("help line lacks the ko action label %q; below the buttons:\n%s", want, help)
		}
	}
	for _, en := range []string{"toggle", "submit", "Downgrade", "older than the running version"} {
		if strings.Contains(frame, en) {
			t.Errorf("ko frame still carries the English string %q", en)
		}
	}
}

// TestPtyCapture_DowngradeConfirmButtonAlignment — AC-ITI-015 (a) on the real
// TTY: the first button label starts at the description's first column.
// Pre-fix values: v1 22-column indent, v2 7 (acceptance.md AC-ITI-015). The
// fix is REQ-ITI-014 (plan.md M7), so this stays red until then.
func TestPtyCapture_DowngradeConfirmButtonAlignment(t *testing.T) {
	ptycaptest.Gate(t)
	frame := ptycapDowngradeFrame(t, "downgrade-confirm-alignment")
	lines, _, desc, buttons, pair := requireDowngradeConfirmLines(t, frame)

	descCol := contentStart(lines[desc])
	first := pair[0]
	if c0, c1 := ptycaptest.DisplayColumn(lines[buttons], pair[0]), ptycaptest.DisplayColumn(lines[buttons], pair[1]); c1 >= 0 && c1 < c0 {
		first = pair[1]
	}
	buttonCol := ptycaptest.DisplayColumn(lines[buttons], first)
	t.Logf("description starts at column %d; first button label %q starts at column %d", descCol, first, buttonCol)
	if buttonCol != descCol {
		t.Errorf("first button label %q at column %d, description at column %d; want equal", first, buttonCol, descCol)
	}
}

// TestPtyCapture_InitFirstScreen — AC-ITI-003 on the real TTY: `moai init` in
// a profile-less temp HOME and an empty working directory reaches the
// conversation-language question with no profile confirmation in front of it,
// and Ctrl+C ends the run with the cancel line.
//
// The absence assertions are not read alone: the anchor and the four option
// lines are required first, so "No profile found" being absent means the run
// passed the profile block (init.go runInit) and rendered the wizard, not that
// neither ever ran.
//
// Red until REQ-ITI-001 removes the profile confirmation from runInit
// (plan.md M4). Until then the confirm holds the screen and the first anchor
// never appears within the deadline.
func TestPtyCapture_InitFirstScreen(t *testing.T) {
	ptycaptest.Gate(t)
	bin := ptycaptest.BuildChild(t, ".")
	c := ptycaptest.NewCase(t, "")
	checkHome := ptycaptest.WatchRealHome(t, c.Dir)
	t.Cleanup(checkHome)

	sentinel := ptycaptest.OpenSentinel(t)
	before := ptycaptest.ListSessions(t)
	if !slices.Contains(before, sentinel) {
		t.Fatalf("sentinel %s not listed before the run: %v", sentinel, before)
	}

	s := ptycaptest.Start(t, c, bin, ptycapCaseInitFirstScreen)
	frame := s.WaitFor(ptycapInitLangAnchor, ptycaptest.AnchorTimeout)
	ptycaptest.VerifyChildEnv(t, c)
	ptycaptest.Export(t, "init-first-screen", frame)

	// Positive existence first (acceptance.md §B P6): the anchor line and the
	// four option lines of the conversation-language question.
	ptycaptest.RequireLines(t, frame, ptycapInitLangAnchor,
		"English", "Korean (한국어)", "Japanese (日本語)", "Chinese (中文)")

	// Only then the absences.
	if strings.Contains(frame, "No profile found") {
		t.Errorf("the first screen still carries the profile confirmation; frame:\n%s", frame)
	}
	for i, l := range strings.Split(ptycaptest.StripANSI(frame), "\n") {
		for _, p := range confirmButtonPairs {
			if strings.Contains(l, p[0]) && strings.Contains(l, p[1]) {
				t.Errorf("line %d %q is a %s / %s confirm button line", i, l, p[0], p[1])
			}
		}
	}

	s.SendKeys("C-c")
	cancelled := s.WaitFor(ptycapInitCancelAnchor, ptycaptest.AnchorTimeout)
	ptycaptest.Export(t, "init-first-screen-cancelled", cancelled)
	s.Close()

	after := ptycaptest.ListSessions(t)
	if strings.Join(before, ",") != strings.Join(after, ",") {
		t.Errorf("moai-ptycap- session set changed: before=%v after=%v", before, after)
	}
}

// TestPtyCapture_SkipWithoutGate — AC-ITI-019 (a) over this package's capture
// tests (the child plus the TestPtyCapture_* set).
func TestPtyCapture_SkipWithoutGate(t *testing.T) {
	ptycaptest.Gate(t)
	ptycaptest.AssertSkipWithoutGate(t, ptycaptest.BuildChild(t, "."), "^TestPtyCapture", 6)
}

// TestPtyCapture_FailWithoutTmux — AC-ITI-019 (b) over this package's
// TestPtyCapture_* tests.
func TestPtyCapture_FailWithoutTmux(t *testing.T) {
	ptycaptest.Gate(t)
	ptycaptest.AssertFailWithoutTmux(t, ptycaptest.BuildChild(t, "."), "^TestPtyCapture_", 5)
}
