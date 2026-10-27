package wizard

// Card t1966 — wizard regression tests (items 4, 10).

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/modu-ai/moai-adk/internal/cli/ptycaptest"
)

// Item 4: pins the counts the questions.go / wizard.go doc comments state
// (four shared questions, five init questions on four pages).
func TestQuestionSetCounts_MatchDocComments(t *testing.T) {
	root := t.TempDir()
	if n := len(DefaultQuestions(root)); n != 4 {
		t.Errorf("DefaultQuestions = %d questions, doc comment says four", n)
	}
	init := InitQuestions(root)
	if len(init) != 5 {
		t.Errorf("InitQuestions = %d questions, doc comment says five", len(init))
	}
	pages := map[string]bool{}
	for i := range init {
		pages[init[i].Group] = true
	}
	if len(pages) != 4 {
		t.Errorf("InitQuestions spans %d pages, doc comment says four: %v", len(pages), pages)
	}
}

// Item 10, per-locale contract (audit D1): the select height is the total of
// the title, the description and the option rows of the locale it is asked for
// — it follows that locale's description rather than the longest one across
// locales. (Before D1 this test pinned the old contract, "height >= the longest
// locale's description": a height fixed at build time had to cover the worst
// locale and left blank card rows in the others.)
func TestSelectHeight_FollowsLocaleDescription(t *testing.T) {
	const id = "t1966_height_probe"
	long := strings.Repeat("긴 설명 문장이 여러 줄로 감싸집니다 ", 10)
	translations["ko"][id] = QuestionTranslation{Title: "t", Description: long}
	t.Cleanup(func() { delete(translations["ko"], id) })

	q := &Question{
		ID:          id,
		Type:        QuestionTypeSelect,
		Title:       "t",
		Description: "short",
		Options:     []Option{{Label: "a", Value: "a"}, {Label: "b", Value: "b"}},
	}

	rows := func(desc string) int {
		return lipgloss.Height(lipgloss.NewStyle().Width(selectFieldUsableWidth).Render(desc))
	}
	if rows(long) <= rows("short") {
		t.Fatalf("precondition: localized description must wrap to more rows than English (%d vs %d)", rows(long), rows("short"))
	}
	if got, want := selectHeight(q, "ko"), 1+rows(long)+len(q.Options); got != want {
		t.Errorf("selectHeight(ko) = %d, want %d (the ko description's rows)", got, want)
	}
	if got, want := selectHeight(q, "en"), 1+rows("short")+len(q.Options); got != want {
		t.Errorf("selectHeight(en) = %d, want %d (the en description's rows, not the longest locale's)", got, want)
	}
}

// optionRowsAt is the number of rows the rendered option rows of opts occupy
// at the select field's usable width (a row that fits still counts 1).
func optionRowsAt(opts []Option) int {
	total := 0
	for _, o := range alignOptionLabels(opts) {
		total += lipgloss.Height(lipgloss.NewStyle().Width(selectFieldUsableWidth).Render(o.Key))
	}
	return total
}

// Item 10 (audit F6): selectHeight must count every option row at the width
// the renderer wraps it to, not one row per option. Per-locale contract (D1):
// the count is exact (a height above the rows renders a blank card row), where
// the old max-over-locales contract only required ">=".
func TestSelectHeight_CountsWrappedOptionRows(t *testing.T) {
	long := strings.Repeat("a deliberately long option description that wraps ", 4)
	q := &Question{
		ID:          "t1966_wrap_probe",
		Type:        QuestionTypeSelect,
		Title:       "t",
		Description: "short",
		Options: []Option{
			{Label: "first", Value: "first", Desc: long},
			{Label: "second", Value: "second", Desc: "fits"},
			{Label: "third", Value: "third", Desc: long},
		},
	}
	if optionRowsAt(q.Options) <= len(q.Options) {
		t.Fatalf("precondition: option rows must wrap (%d rows for %d options)", optionRowsAt(q.Options), len(q.Options))
	}
	want := 1 + wrappedDescriptionRows(q.Description) + optionRowsAt(q.Options)
	if got := selectHeight(q, "en"); got != want {
		t.Errorf("selectHeight = %d, want %d (wrapped option rows counted)", got, want)
	}
}

// Item 10 (audit F6), per-locale contract (D1): for every real select question
// and every locale, selectHeight is that locale's own title + description +
// rendered option rows. Hiding no option is proven at render level by
// TestSelectCard_* (the card must show every option); this pins the arithmetic.
// (Before D1 this test required the height to cover the longest description of
// ANY locale plus the locale's option rows.)
func TestSelectHeight_MatchesLocaleRows(t *testing.T) {
	checked := 0
	for _, q := range selectQuestions(t) {
		for _, locale := range allLocales {
			lq := GetLocalizedQuestion(&q, locale)
			want := 1 + wrappedDescriptionRows(lq.Description) + optionRowsAt(lq.Options)
			if got := selectHeight(&q, locale); got != want {
				t.Errorf("%s/%s: selectHeight = %d, want %d", q.ID, locale, got, want)
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("control failed: no select question was checked")
	}
}

// Audit iter16 F1/F2: the slot a select occupies in the form is the same for
// every locale and never below any locale's own height — the longest description
// and the most option rows across the base text and every translation. (If it
// followed the live locale, the group would be measured at the start locale's
// height only.)
func TestSelectSlotHeight_CoversEveryLocale(t *testing.T) {
	const id = "t1966_slot_probe"
	translations["ko"][id] = QuestionTranslation{Title: "t", Description: strings.Repeat("긴 설명 문장이 여러 줄로 감싸집니다 ", 10)}
	t.Cleanup(func() { delete(translations["ko"], id) })
	probe := Question{
		ID:          id,
		Type:        QuestionTypeSelect,
		Title:       "t",
		Description: "short",
		Options:     []Option{{Label: "a", Value: "a"}, {Label: "b", Value: "b"}},
	}

	checked := 0
	for _, q := range append(selectQuestions(t), probe) {
		slot := selectSlotHeight(&q)
		tallest := 0
		for _, locale := range allLocales {
			h := selectHeight(&q, locale)
			tallest = max(tallest, h)
			if slot < h {
				t.Errorf("%s/%s: slot %d is below the locale's own height %d", q.ID, locale, slot, h)
			}
			checked++
		}
		if slot < tallest {
			t.Errorf("%s: slot %d is below the tallest locale height %d", q.ID, slot, tallest)
		}
	}
	if checked == 0 {
		t.Fatal("control failed: no select question was checked")
	}
	if got, want := selectSlotHeight(&probe), selectHeight(&probe, "ko"); got != want {
		t.Errorf("probe slot = %d, want the ko height %d (the ko description is the longest)", got, want)
	}
}

// --- Audit D1 (REQ-ITI-015 / AC-ITI-016): a select card fits its live locale ---

// cardRowsOf returns the text of every card row of a rendered frame (the lines
// that open with the card border), border removed.
func cardRowsOf(frame string) []string {
	var rows []string
	for _, l := range strings.Split(ptycaptest.StripANSI(frame), "\n") {
		if strings.HasPrefix(l, "┃") {
			rows = append(rows, strings.TrimPrefix(l, "┃"))
		}
	}
	return rows
}

// squash drops all whitespace so wrapped text compares independent of where
// the renderer broke the lines.
func squash(s string) string { return strings.Join(strings.Fields(s), "") }

// assertSelectCard checks the focused select card of q in the frame: its
// title, description and EVERY option are present (nothing hidden), and the
// last card row is not blank (no empty row below the last option).
func assertSelectCard(t *testing.T, label, frame string, q *Question, locale string) {
	t.Helper()
	rows := cardRowsOf(frame)
	if len(rows) == 0 {
		t.Errorf("%s: no select card rendered:\n%s", label, frame)
		return
	}
	lq := GetLocalizedQuestion(q, locale)
	want := lq.Title + lq.Description
	for _, o := range alignOptionLabels(lq.Options) {
		want += o.Key
	}
	if first := strings.TrimSpace(strings.Split(ptycaptest.StripANSI(frame), "\n")[0]); !strings.Contains(first, "/") {
		t.Errorf("%s: the step indicator is not the first line (page scrolled):\n%s", label, ptycaptest.StripANSI(frame))
	}
	got := squash(strings.ReplaceAll(strings.Join(rows, "\n"), "▸", ""))
	if !strings.Contains(got, squash(want)) {
		t.Errorf("%s: the card hides part of the question (title, description, or an option):\n%s", label, ptycaptest.StripANSI(frame))
	}
	if last := rows[len(rows)-1]; strings.TrimSpace(last) == "" {
		t.Errorf("%s: blank card row below the last option:\n%s", label, ptycaptest.StripANSI(frame))
	}
}

// selectQuestions returns every select question of the wizard sets (copies
// with the visibility condition cleared, so each renders alone in a form).
func selectQuestions(t *testing.T) []Question {
	t.Helper()
	root := t.TempDir()
	var all []Question
	all = append(all, DefaultQuestions(root)...)
	all = append(all, GitQuestions()...)
	all = append(all, Page3Questions(root)...)
	var out []Question
	for _, q := range all {
		if q.Type == QuestionTypeSelect {
			q.Condition = nil
			out = append(out, q)
		}
	}
	return out
}

var allLocales = []string{"en", "ko", "ja", "zh"}

// Audit D1: for every real select question (report_format and autonomy_tier
// among them) and a synthetic long-description one, in every locale, the
// rendered card shows every option and leaves no empty card row below the last.
func TestSelectCard_NoBlankRowPerLocale(t *testing.T) {
	const id = "t1966_card_probe"
	translations["ko"][id] = QuestionTranslation{Title: "제목", Description: strings.Repeat("긴 설명 문장이 여러 줄로 감싸집니다 ", 10)}
	t.Cleanup(func() { delete(translations["ko"], id) })
	synthetic := Question{
		ID:          id,
		Type:        QuestionTypeSelect,
		Title:       "Title",
		Description: "short",
		Options:     []Option{{Label: "a", Value: "a"}, {Label: "b", Value: "b"}, {Label: "c", Value: "c"}},
	}

	questions := append(selectQuestions(t), synthetic)
	seen := map[string]bool{}
	for i := range questions {
		q := &questions[i]
		seen[q.ID] = true
		for _, locale := range allLocales {
			d := ptycaptest.NewFormDriver(t, buildUnifiedForm([]Question{*q}, &WizardResult{}, locale))
			assertSelectCard(t, q.ID+"/"+locale, d.View(), q, locale)
		}
	}
	for _, want := range []string{"report_format", "autonomy_tier", id} {
		if !seen[want] {
			t.Errorf("control failed: select question %q was not checked", want)
		}
	}
}

// languageIndex is a locale's position in the conversation_language select.
func languageIndex(t *testing.T, locale string) int {
	t.Helper()
	for i, l := range allLocales {
		if l == locale {
			return i
		}
	}
	t.Fatalf("unknown locale %q", locale)
	return -1
}

// autonomyQuestion returns the autonomy_tier select (page 2 of the init form).
func autonomyQuestion(t *testing.T) *Question {
	t.Helper()
	qs := InitQuestions(t.TempDir())
	q := QuestionByID(qs, "autonomy_tier")
	if q == nil {
		t.Fatal("init set lacks autonomy_tier")
	}
	return q
}

// Audit D1, locale switch: the card re-fits when conversation_language changes
// the live locale — from every locale to every other, both shorter-to-longer
// and longer-to-shorter, page 2 shows every option and no blank card row.
func TestSelectCard_LocaleSwitch(t *testing.T) {
	autonomy := autonomyQuestion(t)
	for _, from := range allLocales {
		for _, to := range allLocales {
			if from == to {
				continue
			}
			t.Run(from+"-to-"+to, func(t *testing.T) {
				d := ptycaptest.NewFormDriver(t, buildUnifiedForm(InitQuestions(t.TempDir()), &WizardResult{}, from))
				for range languageIndex(t, to) {
					d.Down()
				}
				d.Enter() // conversation_language
				d.Enter() // user_name -> page 2
				assertSelectCard(t, "page 2 after "+from+"->"+to, d.View(), autonomy, to)
			})
		}
	}
}

// Audit D1, round trip: switching back and forth (second visits to a locale
// are served from huh's evaluation cache, not recomputed) keeps the card fitted.
func TestSelectCard_LocaleRoundTrip(t *testing.T) {
	autonomy := autonomyQuestion(t)
	d := ptycaptest.NewFormDriver(t, buildUnifiedForm(InitQuestions(t.TempDir()), &WizardResult{}, "en"))
	shiftTab := tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	cursor := 0
	for i, to := range []string{"ja", "en", "ja", "zh", "en", "ko", "zh"} {
		if i > 0 {
			d.Send(shiftTab) // page 2 -> user_name
			d.Send(shiftTab) // -> conversation_language
		}
		for cursor < languageIndex(t, to) {
			d.Down()
			cursor++
		}
		for cursor > languageIndex(t, to) {
			d.Send(tea.KeyPressMsg{Code: tea.KeyUp})
			cursor--
		}
		d.Enter() // conversation_language
		d.Enter() // user_name -> page 2
		assertSelectCard(t, "round trip step "+to, d.View(), autonomy, to)
	}
}

// Audit D1 + audit iter16 F4, group cap: on a terminal too short for the whole
// card, the form hands the select a height cap and the card must stay inside it
// AND scroll — every option, the wrapped last one included, is reachable with
// the cursor. A card that ignores the cap is clipped by the group viewport with
// its last option unreachable (the cap-removal mutant of the audit survived the
// earlier `0 < rows < full` assertion, which a clipped card also satisfies).
func TestSelectCard_RespectsGroupHeightCap(t *testing.T) {
	autonomy := autonomyQuestion(t)
	d := ptycaptest.NewFormDriver(t, buildUnifiedForm(InitQuestions(t.TempDir()), &WizardResult{}, "en"))
	d.Send(tea.WindowSizeMsg{Width: 80, Height: 9})
	d.Enter() // conversation_language
	d.Enter() // user_name -> page 2
	full := selectHeight(autonomy, "en")
	if got := len(cardRowsOf(d.View())); got == 0 || got >= full {
		t.Errorf("card rows on a 9-line terminal = %d, want a clamped card (0 < rows < %d):\n%s", got, full, ptycaptest.StripANSI(d.View()))
	}
	walkOptions(t, "9-line terminal", d, autonomy, "en", 0)
	if !strings.Contains(squash(strings.Join(cardRowsOf(d.View()), "\n")), squash("(Docker/gVisor/etc.)")) {
		t.Errorf("the last option's wrapped row is hidden:\n%s", ptycaptest.StripANSI(d.View()))
	}
}

// --- Audit iter16 (F1, F2): the group measures the form once, at the start locale ---
//
// huh sizes a group when a WindowSizeMsg arrives and hands a field the group's
// height cap only if the field is taller than the group at that moment. Both
// happen in the START locale, so a locale switch must not change what the group
// measures: F1 was the stepper scrolling off the first page after a switch to a
// taller locale, F2 a field that grew after the switch holding no cap, its last
// option unreachable.

var shiftTab = tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}

// switchedInitDriver builds the init form in locale `from` on an 80-column,
// `height`-row terminal, picks `to` in conversation_language and confirms it,
// so the form stands on user_name (page 1) in the target locale.
func switchedInitDriver(t *testing.T, from, to string, height int) *ptycaptest.FormDriver {
	t.Helper()
	d := ptycaptest.NewFormDriver(t, buildUnifiedForm(InitQuestions(t.TempDir()), &WizardResult{}, from))
	if height != 40 {
		d.Send(tea.WindowSizeMsg{Width: 80, Height: height})
	}
	for range languageIndex(t, to) {
		d.Down()
	}
	d.Enter() // conversation_language -> user_name
	return d
}

// assertStepperFirst fails when the frame's first line is not the step
// indicator (the group scrolled), the AC-ITI-021 property.
func assertStepperFirst(t *testing.T, label, frame string) {
	t.Helper()
	plain := ptycaptest.StripANSI(frame)
	if first := strings.TrimSpace(strings.Split(plain, "\n")[0]); !strings.Contains(first, "/") {
		t.Errorf("%s: the step indicator is not the first line (page scrolled):\n%s", label, plain)
	}
}

// assertFrameShows fails when the whole frame (card border and cursor removed)
// does not contain text, compared independent of where lines wrapped.
func assertFrameShows(t *testing.T, label, frame, text string) {
	t.Helper()
	plain := strings.ReplaceAll(strings.ReplaceAll(ptycaptest.StripANSI(frame), "┃", ""), "▸", "")
	if !strings.Contains(squash(plain), squash(text)) {
		t.Errorf("%s: the frame hides part of the question:\n%s", label, ptycaptest.StripANSI(frame))
	}
}

// questionText is a question's title, description and every rendered option
// row, in the locale.
func questionText(q *Question, locale string) string {
	lq := GetLocalizedQuestion(q, locale)
	text := lq.Title + lq.Description
	for _, o := range alignOptionLabels(lq.Options) {
		text += o.Key
	}
	return text
}

// walkOptions moves the cursor of the focused select q from option `cursor` up
// to the first option, then down through every option, and requires each
// option's complete rendered text (every wrapped row) to be visible, with the
// cursor on screen, while it is selected.
func walkOptions(t *testing.T, label string, d *ptycaptest.FormDriver, q *Question, locale string, cursor int) {
	t.Helper()
	for range cursor {
		d.Send(tea.KeyPressMsg{Code: tea.KeyUp})
	}
	for i, o := range alignOptionLabels(GetLocalizedQuestion(q, locale).Options) {
		if i > 0 {
			d.Down()
		}
		frame := d.View()
		if !strings.Contains(ptycaptest.StripANSI(frame), "▸") {
			t.Errorf("%s: option %d: no cursor visible:\n%s", label, i, ptycaptest.StripANSI(frame))
		}
		got := squash(strings.ReplaceAll(strings.Join(cardRowsOf(frame), "\n"), "▸", ""))
		if !strings.Contains(got, squash(o.Key)) {
			t.Errorf("%s: option %d (%q) is not fully visible while selected:\n%s", label, i, o.Key, ptycaptest.StripANSI(frame))
		}
	}
}

// F1: after any locale switch, on a tall terminal, the first line of every page
// the user can reach is the stepper, and the select cards show every option
// with no empty card row below the last — all 16 start x target pairs.
func TestLocaleSwitch_StepperStaysFirstAndCardsFit(t *testing.T) {
	autonomy := autonomyQuestion(t)
	lang := QuestionByID(InitQuestions(t.TempDir()), "conversation_language")
	for _, from := range allLocales {
		for _, to := range allLocales {
			t.Run(from+"-to-"+to, func(t *testing.T) {
				t.Parallel()
				d := switchedInitDriver(t, from, to, 40)
				assertStepperFirst(t, "user_name focused", d.View())
				assertFrameShows(t, "user_name focused: language card", d.View(), questionText(lang, to))
				d.Enter() // user_name -> page 2
				assertSelectCard(t, "page 2", d.View(), autonomy, to)
				d.Send(shiftTab) // page 2 -> user_name
				assertStepperFirst(t, "page 1 back, user_name focused", d.View())
				d.Send(shiftTab) // -> conversation_language
				assertSelectCard(t, "page 1 back, conversation_language focused", d.View(), lang, to)
			})
		}
	}
}

// F2: on a short terminal every option of both select questions can be brought
// into view with the cursor, including the wrapped second row of the last
// autonomy option, whichever locale the form started in and switched to.
func TestLocaleSwitch_ShortTerminalReachesEveryOption(t *testing.T) {
	autonomy := autonomyQuestion(t)
	lang := QuestionByID(InitQuestions(t.TempDir()), "conversation_language")
	for _, height := range []int{9, 10, 12} {
		for _, from := range allLocales {
			for _, to := range allLocales {
				t.Run(fmt.Sprintf("h%d-%s-to-%s", height, from, to), func(t *testing.T) {
					t.Parallel()
					d := switchedInitDriver(t, from, to, height)
					d.Enter() // user_name -> page 2
					walkOptions(t, "autonomy_tier", d, autonomy, to, 0)
					d.Send(shiftTab) // page 2 -> user_name
					d.Send(shiftTab) // -> conversation_language, cursor on the chosen locale
					walkOptions(t, "conversation_language", d, lang, to, languageIndex(t, to))
				})
			}
		}
	}
}
