package wizard

// M7 layout acceptance tests: AC-ITI-015 (confirm button left alignment),
// AC-ITI-016 (no blank lines between fields, no empty card rows), AC-ITI-017
// (option description column aligned by DISPLAY width). All draws are real
// View() renders through the ptycaptest driver (80×40).

import (
	"slices"
	"strings"
	"testing"

	"charm.land/huh/v2"

	"github.com/modu-ai/moai-adk/internal/cli/ptycaptest"
)

// firstNonBlank returns the first line with visible content.
func firstNonBlank(frame string) string {
	for _, line := range strings.Split(frame, "\n") {
		if strings.TrimSpace(line) != "" {
			return line
		}
	}
	return ""
}

// emptyCardRow reports whether a stripped line is an EMPTY CARD ROW: a
// border glyph followed by spaces only (AC-ITI-016's defect shape).
func emptyCardRow(line string) bool {
	trimmed := strings.TrimRight(line, " ")
	if trimmed == "" {
		return false
	}
	return strings.Trim(trimmed, "┃│") == ""
}

// drawInitFirstPage renders the init wizard's first page.
func drawInitFirstPage(t *testing.T) string {
	t.Helper()
	form := buildUnifiedForm(InitQuestions("/tmp/layout-init"), &WizardResult{}, "")
	return ptycaptest.StripANSI(ptycaptest.NewFormDriver(t, form).View())
}

// drawConversationLanguageRows extracts the option rows of the
// conversation_language select from a frame.
func drawConversationLanguageRows(t *testing.T, frame string) []string {
	t.Helper()
	var rows []string
	for _, line := range strings.Split(frame, "\n") {
		for _, label := range []string{"English", "Korean (한국어)", "Japanese (日本語)", "Chinese (中文)"} {
			if strings.Contains(line, label) {
				rows = append(rows, line)
			}
		}
	}
	if len(rows) != 4 {
		t.Fatalf("conversation_language option rows = %d, want 4; frame:\n%s", len(rows), frame)
	}
	return rows
}

// descStartColumns returns the display column at which each option row's
// description begins (the " - " separator before the description).
func descStartColumns(rows []string) []int {
	cols := make([]int, 0, len(rows))
	for _, row := range rows {
		cols = append(cols, ptycaptest.DisplayColumn(row, " - "))
	}
	return cols
}

// TestConfirmButton_LeftAligned is AC-ITI-015's golden half over both
// surfaces: the downgrade confirm and a wizard confirm fixture. The first
// button label's start display column must equal the description's first
// character's display column (pre-fix: the v2 button sat at column 7).
func TestConfirmButton_LeftAligned(t *testing.T) {
	// Surface (a): the downgrade confirm.
	form := NewDowngradeConfirmForm("en", "v9.9.9", "v1.0.0", new(bool))
	frame := ptycaptest.StripANSI(ptycaptest.NewFormDriver(t, form).View())

	descCol := -1
	buttonCol := -1
	for _, line := range strings.Split(frame, "\n") {
		if descCol < 0 && strings.Contains(line, "The requested tag is older") {
			descCol = ptycaptest.DisplayColumn(line, "The requested tag")
		}
		if buttonCol < 0 && strings.Contains(line, "Yes") {
			buttonCol = ptycaptest.DisplayColumn(line, "Yes")
		}
	}
	if descCol < 0 || buttonCol < 0 {
		t.Fatalf("frame lacks the description or button line; descCol=%d buttonCol=%d\n%s", descCol, buttonCol, frame)
	}
	if descCol != buttonCol {
		t.Errorf("button starts at display column %d, description at %d — they must align", buttonCol, descCol)
	}
	if err := ptycaptest.CompareGolden("testdata/axis", "confirm-alignment-downgrade", frame, *updateAxisGolden); err != nil {
		t.Fatal(err)
	}

	// Surface (b): a wizard confirm fixture through the real buildConfirmField.
	q := Question{ID: "fixture_confirm", Group: "Basic", Type: QuestionTypeConfirm,
		Title: "Fixture confirm title", Description: "Fixture confirm description text", Default: "false"}
	conf := buildConfirmField(&q, &WizardResult{}, new(string))
	fixtureForm := huh.NewForm(huh.NewGroup(conf)).WithTheme(newMoAIWizardTheme()).WithAccessible(false)
	fframe := ptycaptest.StripANSI(ptycaptest.NewFormDriver(t, fixtureForm).View())

	fdescCol := -1
	fbuttonCol := -1
	for _, line := range strings.Split(fframe, "\n") {
		if fdescCol < 0 && strings.Contains(line, "Fixture confirm description text") {
			fdescCol = ptycaptest.DisplayColumn(line, "Fixture confirm description")
		}
		if fbuttonCol < 0 && strings.Contains(line, "Yes") {
			fbuttonCol = ptycaptest.DisplayColumn(line, "Yes")
		}
	}
	if fdescCol < 0 || fbuttonCol < 0 {
		t.Fatalf("fixture frame lacks the description or button line; descCol=%d buttonCol=%d\n%s", fdescCol, fbuttonCol, fframe)
	}
	if fdescCol != fbuttonCol {
		t.Errorf("fixture button starts at display column %d, description at %d — they must align", fbuttonCol, fdescCol)
	}
}

// contentPrefix returns the frame's lines up to (not including) the first
// blank line. huh's form viewport fills the space below the last field row
// with blanks — those are structural. The contiguous prefix is where fields
// must join with no gap (AC-ITI-016).
func contentPrefix(frame string) []string {
	lines := strings.Split(frame, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			return lines[:i]
		}
	}
	return lines
}

// assertNoEmptyCardRows fails when any stripped line is an empty card row
// (a border glyph followed by spaces only).
func assertNoEmptyCardRows(t *testing.T, name, frame string) {
	t.Helper()
	for i, line := range strings.Split(frame, "\n") {
		if emptyCardRow(line) {
			t.Errorf("%s line %d is an empty card row (┃ + spaces only) (REQ-ITI-015)", name, i+1)
		}
	}
}

// TestLayout_NoBlankBetweenFields is AC-ITI-016: on the init first page and
// every profile group, the contiguous content prefix carries ALL the page's
// field anchors (no separator blank split the fields apart), and the frame
// has no empty card rows. Pre-fix values: 1 blank line between fields, 4
// empty card rows on the init first page.
func TestLayout_NoBlankBetweenFields(t *testing.T) {
	t.Run("init-first-page", func(t *testing.T) {
		frame := drawInitFirstPage(t)
		assertNoEmptyCardRows(t, "init first page", frame)
		prefix := strings.Join(contentPrefix(frame), "\n")
		for _, anchor := range append([]string{"Select conversation language", "Enter your name"},
			[]string{"English", "Korean (한국어)", "Japanese (日本語)", "Chinese (中文)"}...) {
			if !strings.Contains(prefix, anchor) {
				t.Errorf("init first page content prefix lacks %q — a gap split the fields", anchor)
			}
		}
	})

	t.Run("profile-all-groups", func(t *testing.T) {
		initial := ProfileResult{ConversationLang: "en", GitCommitLang: "en", CodeCommentLang: "en", DocLang: "en"}
		form := NewProfileForm(profileStepperOptions(), initial, "en")
		d := ptycaptest.NewFormDriver(t, form)
		for _, g := range []struct {
			entries int
			name    string
			anchors []string
		}{
			{1, "language", []string{"Select your language", "Chinese (中文)"}},
			{1, "identity", []string{"User name"}},
			{3, "languages", []string{"Git commit message language", "Documentation language"}},
			{4, "model", []string{"Default model override", "Permission mode"}},
			{1, "project", []string{"Development mode"}},
		} {
			frame := ptycaptest.StripANSI(d.View())
			assertNoEmptyCardRows(t, g.name+" group", frame)
			prefix := strings.Join(contentPrefix(frame), "\n")
			for _, anchor := range g.anchors {
				if !strings.Contains(prefix, anchor) {
					t.Errorf("%s group content prefix lacks %q — a gap split the fields", g.name, anchor)
				}
			}
			for range g.entries {
				d.Enter()
			}
		}
		if form.State != huh.StateCompleted {
			t.Fatalf("profile form must complete, state=%v", form.State)
		}
	})
}

// TestOptionDescriptionColumn_DisplayWidthAligned is AC-ITI-017: on BOTH
// conversation_language selects (init and profile), the four option rows'
// description start display columns are equal — East Asian wide runes count 2
// (display width, not runes or bytes).
func TestOptionDescriptionColumn_DisplayWidthAligned(t *testing.T) {
	t.Run("init", func(t *testing.T) {
		cols := descStartColumns(drawConversationLanguageRows(t, drawInitFirstPage(t)))
		uniq := slices.Clone(cols)
		slices.Sort(uniq)
		uniq = slices.Compact(uniq)
		if len(uniq) != 1 {
			t.Errorf("init conversation_language description start columns = %v, want one aligned column", cols)
		}
	})

	t.Run("profile", func(t *testing.T) {
		initial := ProfileResult{ConversationLang: "en", GitCommitLang: "en", CodeCommentLang: "en", DocLang: "en"}
		form := NewProfileForm(profileStepperOptions(), initial, "en")
		frame := ptycaptest.StripANSI(ptycaptest.NewFormDriver(t, form).View())
		cols := descStartColumns(drawConversationLanguageRows(t, frame))
		uniq := slices.Clone(cols)
		slices.Sort(uniq)
		uniq = slices.Compact(uniq)
		if len(uniq) != 1 {
			t.Errorf("profile conversation_language description start columns = %v, want one aligned column", cols)
		}
	})
}
