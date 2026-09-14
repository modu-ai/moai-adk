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

	"charm.land/huh/v2"

	"github.com/modu-ai/moai-adk/internal/cli/ptycaptest"
)

// The pty capture cases (SPEC-CLI-TUX-RENDER-I18N-001 acceptance.md §B):
// S1 init first page, S2 a confirm-question fixture through the real
// buildUnifiedForm, S3 the downgrade confirm, S4 the profile wizard groups —
// each at en and ko, 80 columns.
const (
	ptycapCaseInitFirstPage = "init-first-page"

	ptycapCaseInitFirstPageKo = "init-first-page-ko"

	ptycapCaseConfirmFixtureEn = "confirm-fixture-en"
	ptycapCaseConfirmFixtureKo = "confirm-fixture-ko"

	ptycapCaseDowngradeConfirmEn = "downgrade-confirm-en"
	ptycapCaseDowngradeConfirmKo = "downgrade-confirm-ko"

	ptycapCaseProfileGroupsEn = "profile-groups-en"
	ptycapCaseProfileGroupsKo = "profile-groups-ko"
)

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
		runChildForm(t, buildUnifiedForm(InitQuestions(mustCwd()), &WizardResult{}, "en"))
	case ptycapCaseInitFirstPageKo:
		runChildForm(t, buildUnifiedForm(InitQuestions(mustCwd()), &WizardResult{}, "ko"))
	case ptycapCaseConfirmFixtureEn:
		runChildForm(t, buildUnifiedForm(confirmFixtureQuestions(), &WizardResult{}, "en"))
	case ptycapCaseConfirmFixtureKo:
		runChildForm(t, buildUnifiedForm(confirmFixtureQuestions(), &WizardResult{}, "ko"))
	case ptycapCaseDowngradeConfirmEn:
		runChildForm(t, NewDowngradeConfirmForm("en", "v9.9.9", "v1.0.0", new(bool)))
	case ptycapCaseDowngradeConfirmKo:
		runChildForm(t, NewDowngradeConfirmForm("ko", "v9.9.9", "v1.0.0", new(bool)))
	case ptycapCaseProfileGroupsEn:
		runChildForm(t, NewProfileForm(profileStepperOptions(), profileInitial("en"), "en"))
	case ptycapCaseProfileGroupsKo:
		runChildForm(t, NewProfileForm(profileStepperOptions(), profileInitial("ko"), "ko"))
	default:
		t.Fatalf("unknown child case %q", name)
	}
}

// mustCwd returns the working directory, failing the child when it cannot be
// read (the init questions embed it in a prompt).
func mustCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "/tmp"
	}
	return cwd
}

// runChildForm runs one form on the child TTY and logs the outcome.
func runChildForm(t *testing.T, form *huh.Form) {
	t.Helper()
	err := form.Run()
	t.Logf("form returned: %v", err)
}

// confirmFixtureQuestions is the S2 surface: one confirm question rendered
// through the real buildUnifiedForm path (the current init/profile sets carry
// no confirm question, so the fixture is the canonical confirm surface).
func confirmFixtureQuestions() []Question {
	return []Question{{
		ID: "fixture_confirm", Group: "Basic", Type: QuestionTypeConfirm,
		Title: "Fixture confirm title", Description: "Fixture confirm description text", Default: "false",
	}}
}

// profileInitial seeds the profile form with the locale pre-selection.
func profileInitial(locale string) ProfileResult {
	return ProfileResult{ConversationLang: locale, GitCommitLang: "en", CodeCommentLang: "en", DocLang: "en"}
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

// confirmProbe names the header line and the affirmative button label a
// confirm surface's frame must show; the blank rows between them are the
// confirm-internal gap AC-TRI-003 measures.
type confirmProbe struct {
	header      string
	affirmative string
}

// baselineSurfaces is the REQ-TRI-001 capture set: S1-S4 at en and ko, each
// with its reachability anchor(s) and, where the surface carries a confirm
// field, the blank-row probe.
var baselineSurfaces = []struct {
	name      string
	childCase string
	anchor    string
	extra     []string
	confirm   *confirmProbe
}{
	{
		name: "init-first-page-en", childCase: ptycapCaseInitFirstPage,
		anchor: "Select conversation language",
		extra:  []string{"English", "Korean (한국어)", "Japanese (日本語)", "Chinese (中文)"},
	},
	{
		name: "init-first-page-ko", childCase: ptycapCaseInitFirstPageKo,
		anchor: "대화 언어 선택",
		extra:  []string{"English", "Korean (한국어)"},
	},
	{
		name: "confirm-fixture-en", childCase: ptycapCaseConfirmFixtureEn,
		anchor:  "Fixture confirm description text",
		extra:   []string{"Fixture confirm title"},
		confirm: &confirmProbe{header: "Fixture confirm description text", affirmative: "Yes"},
	},
	{
		name: "confirm-fixture-ko", childCase: ptycapCaseConfirmFixtureKo,
		anchor:  "Fixture confirm description text",
		extra:   []string{"Fixture confirm title", "예"},
		confirm: &confirmProbe{header: "Fixture confirm description text", affirmative: "예"},
	},
	{
		name: "downgrade-confirm-en", childCase: ptycapCaseDowngradeConfirmEn,
		anchor:  "Downgrade v9.9.9",
		extra:   []string{"The requested tag is older"},
		confirm: &confirmProbe{header: "The requested tag is older", affirmative: "Yes"},
	},
	{
		name: "downgrade-confirm-ko", childCase: ptycapCaseDowngradeConfirmKo,
		anchor:  "다운그레이드할까요?",
		extra:   []string{"요청한 태그가", "예"},
		confirm: &confirmProbe{header: "요청한 태그가", affirmative: "예"},
	},
	{
		name: "profile-groups-en", childCase: ptycapCaseProfileGroupsEn,
		anchor: "Select your language",
		extra:  []string{"Chinese (中文)"},
	},
	{
		name: "profile-groups-ko", childCase: ptycapCaseProfileGroupsKo,
		anchor: "언어를 선택하세요",
	},
}

// TestPtyCapture_BaselineSurfaces is REQ-TRI-001: captures every interactive
// surface at en/ko x 80 columns before any repair edit, asserts each frame's
// anchor reachability (the AC-TRI-001 premise), MEASURES the confirm-internal
// blank rows (the AC-TRI-003 repair-before observation), and exports the
// frames. Frames land in MOAI_PTY_CAPTURE_OUT when set.
func TestPtyCapture_BaselineSurfaces(t *testing.T) {
	ptycaptest.Gate(t)
	bin := ptycaptest.BuildChild(t, ".")

	for _, tc := range baselineSurfaces {
		t.Run(tc.name, func(t *testing.T) {
			c := ptycaptest.NewCase(t, "")
			s := ptycaptest.Start(t, c, bin, tc.childCase)
			frame := s.WaitFor(tc.anchor, ptycaptest.AnchorTimeout)

			wants := append([]string{tc.anchor}, tc.extra...)
			for _, w := range wants {
				if !strings.Contains(frame, w) {
					t.Errorf("frame lacks anchor %q (reachability premise, AC-TRI-001)", w)
				}
			}

			if tc.confirm != nil {
				n := confirmBlankRows(t, frame, tc.confirm.header, tc.confirm.affirmative)
				t.Logf("CONFIRM-INTERNAL-BLANK-ROWS %s = %d", tc.name, n)
			}

			path := ptycaptest.Export(t, "baseline-"+tc.name, frame)
			if b, err := os.ReadFile(path); err != nil || !strings.Contains(string(b), tc.anchor) {
				t.Errorf("capture file %s does not carry the anchor %q (err %v)", path, tc.anchor, err)
			}

			s.SendKeys("C-c")
			s.Close()
		})
	}
}

// confirmBlankRows counts the blank rows (border glyph + spaces, per the
// AC-ITI-016 empty-card-row shape) between the confirm header's last line and
// the button row in a captured frame. ANSI is stripped first; display
// judgement runs on plain text.
func confirmBlankRows(t *testing.T, frame, header, affirmative string) int {
	t.Helper()
	lines := strings.Split(ptycaptest.StripANSI(frame), "\n")
	headerIdx := -1
	for i, line := range lines {
		if strings.Contains(line, header) {
			headerIdx = i
		}
	}
	if headerIdx < 0 {
		t.Fatalf("frame lacks confirm header %q\n%s", header, frame)
	}
	buttonsIdx := -1
	for i := headerIdx + 1; i < len(lines); i++ {
		if strings.Contains(lines[i], affirmative) {
			buttonsIdx = i
			break
		}
	}
	if buttonsIdx < 0 {
		t.Fatalf("frame lacks button row %q after header %q\n%s", affirmative, header, frame)
	}
	n := 0
	for i := headerIdx + 1; i < buttonsIdx; i++ {
		if emptyCardRow(lines[i]) || strings.TrimSpace(lines[i]) == "" {
			n++
		}
	}
	return n
}

// TestPtyCapture_SkipWithoutGate — AC-ITI-019 (a) over this package's capture
// tests.
func TestPtyCapture_SkipWithoutGate(t *testing.T) {
	ptycaptest.Gate(t)
	ptycaptest.AssertSkipWithoutGate(t, ptycaptest.BuildChild(t, "."), "^TestPtyCapture", 6)
}

// TestPtyCapture_FailWithoutTmux — AC-ITI-019 (b) over this package's
// TestPtyCapture_ tests (the Child test is excluded by the underscore).
func TestPtyCapture_FailWithoutTmux(t *testing.T) {
	ptycaptest.Gate(t)
	ptycaptest.AssertFailWithoutTmux(t, ptycaptest.BuildChild(t, "."), "^TestPtyCapture_", 5)
}
