package wizard

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// Run executes the wizard and returns the result.
// Each question is run as an independent huh.Form to avoid the YOffset scroll bug in huh v0.8.x.
func Run(questions []Question, styles *Styles) (*WizardResult, error) {
	return RunWithLocale(questions, styles, "")
}

// RunWithDefaults runs the wizard with default questions for the given project root.
// If locale is not empty, the wizard UI is displayed in that language.
func RunWithDefaults(projectRoot, locale string) (*WizardResult, error) {
	questions := DefaultQuestions(projectRoot)
	return RunWithLocale(questions, nil, locale)
}

// wizardTotalSteps is the canonical step count for the init wizard (AC-CLI-TUI-007).
// It reflects the 6-step init flow from screens.jsx:ScreenInit.
const wizardTotalSteps = 6

// RunWithLocale initializes the locale and runs the wizard.
//
// @MX:NOTE: [AUTO] Each visible question is rendered with tui.Stepper(current, wizardTotalSteps)
// printed to stdout before the huh form. Stepper uses nil Theme (auto light/dark via lipgloss).
func RunWithLocale(questions []Question, styles *Styles, locale string) (*WizardResult, error) {
	if len(questions) == 0 {
		return nil, ErrNoQuestions
	}

	result := &WizardResult{}
	currentLocale := locale
	theme := newMoAIWizardTheme()
	visibleIdx := 0

	for i := range questions {
		q := &questions[i]

		// Skip questions whose condition is not met.
		if q.Condition != nil && !q.Condition(result) {
			continue
		}

		visibleIdx++
		// Display step indicator above the huh form (AC-CLI-TUI-007).
		// Stepper uses nil Theme so it auto-selects light/dark via lipgloss adaptive color.
		fmt.Println(tui.Stepper(visibleIdx, wizardTotalSteps, nil))

		g := buildQuestionGroup(q, result, &currentLocale)
		form := huh.NewForm(g).
			WithTheme(theme).
			WithAccessible(false)

		if err := form.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil, ErrCancelled
			}
			return nil, fmt.Errorf("wizard error: %w", err)
		}
	}

	return result, nil
}

// buildQuestionGroup creates a huh.Group for a single question.
// Conditional questions use WithHideFunc to check visibility at runtime.
func buildQuestionGroup(q *Question, result *WizardResult, locale *string) *huh.Group {
	var field huh.Field

	switch q.Type {
	case QuestionTypeSelect:
		field = buildSelectField(q, result, locale)
	case QuestionTypeInput:
		field = buildInputField(q, result, locale)
	}

	g := huh.NewGroup(field)

	// Apply conditional visibility.
	if q.Condition != nil {
		cond := q.Condition
		g = g.WithHideFunc(func() bool {
			return !cond(result)
		})
	}

	return g
}

// buildSelectField creates a huh.Select field for a select-type question.
func buildSelectField(q *Question, result *WizardResult, locale *string) *huh.Select[string] {
	var selected string

	// Set default value as initial selection.
	if q.Default != "" {
		selected = q.Default
	}

	// Build options eagerly at form-construction time using the current locale.
	// Each question runs as its own sequential Form, so locale is already set
	// by the time subsequent questions are built.
	//
	// We deliberately avoid OptionsFunc here: huh v0.8.x OptionsFunc forces
	// s.height = defaultHeight (10) when no explicit height is set. Once
	// s.height > 0, updateViewportHeight() resets viewport.YOffset = s.selected
	// on *every* Update() call, causing the viewport to always scroll so the
	// selected item is at the top — hiding options above the cursor.
	//
	// Using Options() (static) with no Height() call keeps s.height == 0, so
	// updateViewportHeight() takes the auto-size branch, sizes the viewport to
	// exactly the number of options, and never resets YOffset. Navigation keys
	// move only the cursor highlight; the visible option list stays fixed.
	lq := GetLocalizedQuestion(q, *locale)
	opts := make([]huh.Option[string], len(lq.Options))
	for i, opt := range lq.Options {
		key := opt.Label
		if opt.Desc != "" {
			key = opt.Label + " - " + opt.Desc
		}
		opts[i] = huh.NewOption(key, opt.Value)
	}

	sel := huh.NewSelect[string]().
		TitleFunc(func() string {
			lq := GetLocalizedQuestion(q, *locale)
			return lq.Title
		}, locale).
		DescriptionFunc(func() string {
			lq := GetLocalizedQuestion(q, *locale)
			return lq.Description
		}, locale).
		Options(opts...).
		Value(&selected)

	// Wire up value storage after each change.
	sel.Validate(func(val string) error {
		saveAnswer(q.ID, val, result, locale)
		return nil
	})

	return sel
}

// buildInputField creates a huh.Input field for an input-type question.
func buildInputField(q *Question, result *WizardResult, locale *string) *huh.Input {
	var value string
	if q.Default != "" {
		value = q.Default
	}

	inp := huh.NewInput().
		TitleFunc(func() string {
			lq := GetLocalizedQuestion(q, *locale)
			return lq.Title
		}, locale).
		DescriptionFunc(func() string {
			lq := GetLocalizedQuestion(q, *locale)
			return lq.Description
		}, locale).
		Value(&value)

	if q.Default != "" {
		inp = inp.Placeholder(q.Default)
	}

	// Validation and value storage.
	qID := q.ID
	required := q.Required
	defVal := q.Default
	inp = inp.Validate(func(val string) error {
		v := strings.TrimSpace(val)
		if v == "" && defVal != "" {
			v = defVal
		}
		if required && v == "" {
			uiStr := GetUIStrings(*locale)
			return errors.New(uiStr.ErrorRequired)
		}
		saveAnswer(qID, v, result, locale)
		return nil
	})

	return inp
}

// saveAnswer stores an answer in the result.
func saveAnswer(id, value string, result *WizardResult, locale *string) {
	switch id {
	case "project_name":
		result.ProjectName = value
	case "model_policy":
		result.ModelPolicy = value
	case "development_mode":
		result.DevelopmentMode = value
	case "git_mode":
		result.GitMode = value
	case "git_provider":
		result.GitProvider = value
	case "github_username":
		result.GitHubUsername = value
	case "github_token":
		result.GitHubToken = value
	case "gitlab_instance_url":
		result.GitLabInstanceURL = value
	case "gitlab_username":
		result.GitLabUsername = value
	case "gitlab_token":
		result.GitLabToken = value
	}
	_ = locale // locale is kept for GetLocalizedQuestion compatibility
}

// newMoAIWizardTheme creates a huh.Theme with MoAI wizard branding.
//
// All colour values are derived from internal/tui.LightTheme / DarkTheme tokens
// (AC-CLI-TUI-013: no hex literals outside internal/tui/).
//
// @MX:ANCHOR: [AUTO] Called by every RunWithLocale invocation; single huh.Theme factory
// @MX:REASON: All wizard forms share one theme; changes here affect the entire init wizard UX
func newMoAIWizardTheme() *huh.Theme {
	t := huh.ThemeBase()

	// Resolve tui design tokens into adaptive colour values.
	c := wizardColors()

	t.Focused.Base = t.Focused.Base.BorderForeground(c.Border)
	t.Focused.Card = t.Focused.Base
	t.Focused.Title = t.Focused.Title.Foreground(c.Primary).Bold(true)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(c.Primary).Bold(true).MarginBottom(1)
	t.Focused.Description = t.Focused.Description.Foreground(c.Muted)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(c.Error)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(c.Error)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(c.Primary).SetString("▸ ")
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(c.Primary)
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(c.Primary)
	t.Focused.Option = t.Focused.Option.Foreground(c.Text)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(c.Primary)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(c.Success)
	t.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(c.Success).SetString("◆ ")
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(c.Text)
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(c.Muted).SetString("◇ ")
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(c.Primary)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(c.Muted)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(c.Secondary)
	t.Focused.FocusedButton = t.Focused.FocusedButton.
		Foreground(c.ButtonFg).
		Background(c.Primary)
	t.Focused.BlurredButton = t.Focused.BlurredButton.
		Foreground(c.Text).
		Background(c.ButtonBlurredBg)
	t.Focused.Next = t.Focused.FocusedButton

	t.Blurred = t.Focused
	t.Blurred.Base = t.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.Card = t.Blurred.Base
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()

	t.Group.Title = t.Focused.Title
	t.Group.Description = t.Focused.Description

	return t
}
