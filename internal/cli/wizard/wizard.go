package wizard

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// statuslineSegmentPrefix is the prefix used for statusline segment question IDs.
const statuslineSegmentPrefix = "statusline_seg_"

// Run executes the wizard and returns the result.
// Each question runs as its own independent huh.Form to avoid the huh v0.8.x
// YOffset scroll bug that occurs when multiple groups share a single viewport.
func Run(questions []Question, styles *Styles) (*WizardResult, error) {
	if len(questions) == 0 {
		return nil, ErrNoQuestions
	}

	result := &WizardResult{}
	locale := ""
	theme := newMoAIWizardTheme()

	for i := range questions {
		q := &questions[i]

		// Pre-check condition: skip questions whose condition is not met.
		if q.Condition != nil && !q.Condition(result) {
			continue
		}

		g := buildQuestionGroup(q, result, &locale)
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

// RunWithDefaults runs the wizard with default questions for the given project root.
func RunWithDefaults(projectRoot string) (*WizardResult, error) {
	questions := DefaultQuestions(projectRoot)
	return Run(questions, nil)
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
	case "locale":
		result.Locale = value
		*locale = value
	case "user_name":
		result.UserName = value
	case "project_name":
		result.ProjectName = value
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
	case "git_commit_lang":
		result.GitCommitLang = value
	case "code_comment_lang":
		result.CodeCommentLang = value
	case "doc_lang":
		result.DocLang = value
	case "model_policy":
		result.ModelPolicy = value
	case "agent_teams_mode":
		result.AgentTeamsMode = value
	case "max_teammates":
		result.MaxTeammates = value
	case "default_model":
		result.DefaultModel = value
	case "teammate_display":
		result.TeammateDisplay = value
	case "statusline_preset":
		result.StatuslinePreset = value
	default:
		if after, ok := strings.CutPrefix(id, statuslineSegmentPrefix); ok {
			segName := after
			if result.StatuslineSegments == nil {
				result.StatuslineSegments = make(map[string]bool)
			}
			result.StatuslineSegments[segName] = (value == "true")
		}
	}
}

// newMoAIWizardTheme creates a huh.Theme with MoAI wizard branding.
func newMoAIWizardTheme() *huh.Theme {
	t := huh.ThemeBase()

	// Map wizard brand colors to huh theme.
	primary := lipgloss.AdaptiveColor{Light: "#C45A3C", Dark: ColorPrimary}
	secondary := lipgloss.AdaptiveColor{Light: "#5B21B6", Dark: ColorSecondary}
	green := lipgloss.AdaptiveColor{Light: "#059669", Dark: ColorSuccess}
	red := lipgloss.AdaptiveColor{Light: "#DC2626", Dark: ColorError}
	text := lipgloss.AdaptiveColor{Light: "#111827", Dark: ColorText}
	muted := lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: ColorMuted}
	border := lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: ColorBorder}

	t.Focused.Base = t.Focused.Base.BorderForeground(border)
	t.Focused.Card = t.Focused.Base
	t.Focused.Title = t.Focused.Title.Foreground(primary).Bold(true)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(primary).Bold(true).MarginBottom(1)
	t.Focused.Description = t.Focused.Description.Foreground(muted)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(red)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(red)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(primary).SetString("▸ ")
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(primary)
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(primary)
	t.Focused.Option = t.Focused.Option.Foreground(text)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(primary)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(green)
	t.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(green).SetString("◆ ")
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(text)
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(muted).SetString("◇ ")
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(primary)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(muted)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(secondary)
	t.Focused.FocusedButton = t.Focused.FocusedButton.
		Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
		Background(primary)
	t.Focused.BlurredButton = t.Focused.BlurredButton.
		Foreground(text).
		Background(lipgloss.AdaptiveColor{Light: "#E5E7EB", Dark: "#374151"})
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
