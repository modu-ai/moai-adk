package wizard

// pty capture cases for the wizard package (acceptance.md §B, design.md §11).
// The harness, the real HOME watch list, the render helpers, and the harness
// self-checks live in internal/cli/ptycaptest; this file holds only the
// product-path child and the wizard-level runs over it.
//
// Tests named TestPtyCapture_* need tmux and run only under MOAI_PTY_CAPTURE=1.

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/ptycaptest"
)

// ptycapCaseInitFirstPage runs the init wizard form (en) on the TTY.
const ptycapCaseInitFirstPage = "init-first-page"

// TestPtyCaptureChild is the program a pty session runs. It records its
// effective environment first, then runs the named case on the real TTY.
func TestPtyCaptureChild(t *testing.T) {
	name := os.Getenv(ptycaptest.ChildEnv)
	if name == "" {
		t.Skip("runs only inside a pty capture session")
	}
	if err := ptycaptest.RecordEnv(); err != nil {
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

// TestPtyCapture_NormalRun — AC-ITI-020 (1) on the product screen, plus the
// effective-environment observation and the real HOME watch comparison.
func TestPtyCapture_NormalRun(t *testing.T) {
	ptycaptest.Gate(t)
	bin := ptycaptest.BuildChild(t, ".")
	c := ptycaptest.NewCase(t, "")
	checkHome := ptycaptest.WatchRealHome(t, c.Dir)

	sentinel := ptycaptest.OpenSentinel(t)
	before := ptycaptest.ListSessions(t)
	if !slices.Contains(before, sentinel) {
		t.Fatalf("sentinel %s not listed before the run: %v", sentinel, before)
	}

	s := ptycaptest.Start(t, c, bin, ptycapCaseInitFirstPage)
	frame := s.WaitFor("Select conversation language", ptycaptest.AnchorTimeout)
	ptycaptest.VerifyChildEnv(t, c)
	ptycaptest.RequireLines(t, frame, "Select conversation language",
		"English", "Korean (한국어)", "Japanese (日本語)", "Chinese (中文)")
	path := ptycaptest.Export(t, "normal-run-init-first-page", frame)
	if b, err := os.ReadFile(path); err != nil || !strings.Contains(string(b), "Select conversation language") {
		t.Fatalf("capture file %s does not carry the anchor (err %v)", path, err)
	}
	s.SendKeys("C-c")
	s.Close()

	after := ptycaptest.ListSessions(t)
	t.Logf("moai-ptycap- sessions before=%v after=%v", before, after)
	if strings.Join(before, ",") != strings.Join(after, ",") {
		t.Errorf("moai-ptycap- session set changed: before=%v after=%v", before, after)
	}
	checkHome()
}

// TestPtyCapture_SkipWithoutGate — AC-ITI-019 (a) over this package's four
// capture tests.
func TestPtyCapture_SkipWithoutGate(t *testing.T) {
	ptycaptest.Gate(t)
	ptycaptest.AssertSkipWithoutGate(t, ptycaptest.BuildChild(t, "."), "^TestPtyCapture", 4)
}

// TestPtyCapture_FailWithoutTmux — AC-ITI-019 (b) over this package's three
// TestPtyCapture_* tests.
func TestPtyCapture_FailWithoutTmux(t *testing.T) {
	ptycaptest.Gate(t)
	ptycaptest.AssertFailWithoutTmux(t, ptycaptest.BuildChild(t, "."), "^TestPtyCapture_", 3)
}
