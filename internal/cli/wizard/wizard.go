package wizard

import (
	"errors"
	"fmt"
	"image/color"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// Run executes the wizard and returns the result.
//
// SPEC-CLI-TUX-V3-002 M2c (REQ-TUX2-006): the wizard presents a SINGLE
// multi-group huh v2 form. The former one-question-one-form workaround for
// the huh v0.8.x YOffset scroll defect is retired — the M2a spike verified
// the defect is resolved in huh v2 (ensureVisible minimum-scroll clamping
// replaced the unconditional viewport.YOffset=selected reset; evidence in
// progress.md §E.2).
func Run(questions []Question, styles *Styles) (*WizardResult, error) {
	return RunWithLocale(questions, styles, "")
}

// RunWithDefaults runs the wizard with default questions for the given project root.
// If locale is not empty, the wizard UI is displayed in that language.
func RunWithDefaults(projectRoot, locale string) (*WizardResult, error) {
	questions := DefaultQuestions(projectRoot)
	return RunWithLocale(questions, nil, locale)
}

// RunWithDefaultsModes runs the wizard with mode flags controlling Phase 1 question visibility.
// standardMode=true presents Phase 1 questions; advancedMode=true implies standardMode.
func RunWithDefaultsModes(projectRoot, locale string, standardMode, advancedMode bool) (*WizardResult, error) {
	// Merge default + Phase 1 questions
	questions := DefaultQuestions(projectRoot)
	questions = append(questions, Phase1Questions(projectRoot)...)

	// When advanced mode requested, check Phase 2 prerequisites and append stubs.
	if advancedMode {
		gate := IsAdvancedWizardReady()
		questions = append(questions, Phase2Questions(gate)...)
	}

	// Pre-populate mode flags so Condition funcs see them from the start
	result := &WizardResult{
		StandardMode: standardMode || advancedMode,
		AdvancedMode: advancedMode,
		// Phase 1 boolean defaults (applied before wizard so non-interactive path works)
		EnforceQuality:            true,
		CoverageExemptionsEnabled: false,
		DesignEnabled:             true,
		ClaudeDesignEnabled:       true,
	}

	// Run wizard with pre-populated result
	if err := runWithResult(questions, nil, locale, result); err != nil {
		return nil, err
	}
	return result, nil
}

// RunWithLocale initializes the locale and runs the wizard.
//
// @MX:NOTE: [AUTO] The wizard is one huh v2 form: consecutive unconditional
// questions share a field group; each conditional question is its own group
// with a lazily-evaluated hide func. A skip-focus Note at the top of every
// group renders the tui.Stepper with a dynamic denominator (REQ-TUX2-008).
func RunWithLocale(questions []Question, styles *Styles, locale string) (*WizardResult, error) {
	result := &WizardResult{}
	if err := runWithResult(questions, styles, locale, result); err != nil {
		return nil, err
	}
	return result, nil
}

// runWithResult builds the unified multi-group form and runs it into an
// existing WizardResult (instead of creating a new one).
func runWithResult(questions []Question, _ *Styles, locale string, result *WizardResult) error {
	if len(questions) == 0 {
		return ErrNoQuestions
	}

	// Behavior preservation: when every question's condition is false for the
	// pre-populated result, the former per-question loop skipped all forms and
	// returned the result untouched — do the same without running the form.
	if TotalVisibleQuestions(questions, result) == 0 {
		return nil
	}

	form := buildUnifiedForm(questions, result, locale)
	if err := form.Run(); err != nil {
		return mapFormErr(err)
	}
	return nil
}

// mapFormErr converts huh errors into wizard sentinel errors.
func mapFormErr(err error) error {
	if errors.Is(err, huh.ErrUserAborted) {
		return ErrCancelled
	}
	return fmt.Errorf("wizard error: %w", err)
}

// buildUnifiedForm assembles the single multi-group form (REQ-TUX2-006).
func buildUnifiedForm(questions []Question, result *WizardResult, locale string) *huh.Form {
	currentLocale := locale
	groups := buildFormGroups(questions, result, &currentLocale)
	return huh.NewForm(groups...).
		WithTheme(newMoAIWizardTheme()).
		WithAccessible(false)
}

// buildFormGroups partitions questions into huh groups: consecutive
// unconditional questions sharing the same Group label merge into one
// multi-field group; each conditional question becomes its own single-field
// group whose WithHideFunc wraps the question Condition. huh evaluates hide
// funcs lazily at navigation time, so conditions on answers collected earlier
// in the same form work (field Blur runs Validate → saveAnswer before the
// next group's hide func is consulted).
func buildFormGroups(questions []Question, result *WizardResult, locale *string) []*huh.Group {
	var groups []*huh.Group
	var pending []*Question
	pendingLabel := ""

	flush := func() {
		if len(pending) == 0 {
			return
		}
		fields := make([]huh.Field, 0, len(pending)+1)
		fields = append(fields, stepperNote(questions, pending[0], result))
		for _, q := range pending {
			fields = append(fields, buildField(q, result, locale))
		}
		groups = append(groups, huh.NewGroup(fields...))
		pending = nil
	}

	for i := range questions {
		q := &questions[i]
		if q.Condition != nil {
			flush()
			groups = append(groups, buildConditionalGroup(questions, q, result, locale))
			continue
		}
		if len(pending) > 0 && q.Group != pendingLabel {
			flush()
		}
		pendingLabel = q.Group
		pending = append(pending, q)
	}
	flush()
	return groups
}

// buildConditionalGroup wraps a conditional question in its own group with a
// stepper note and a hide func derived from the question Condition.
func buildConditionalGroup(questions []Question, q *Question, result *WizardResult, locale *string) *huh.Group {
	g := huh.NewGroup(
		stepperNote(questions, q, result),
		buildField(q, result, locale),
	)
	cond := q.Condition
	return g.WithHideFunc(func() bool { return !cond(result) })
}

// buildQuestionGroup creates a huh.Group for a single question.
// Conditional questions use WithHideFunc to check visibility at runtime.
func buildQuestionGroup(q *Question, result *WizardResult, locale *string) *huh.Group {
	g := huh.NewGroup(buildField(q, result, locale))
	if q.Condition != nil {
		cond := q.Condition
		g = g.WithHideFunc(func() bool {
			return !cond(result)
		})
	}
	return g
}

// buildField dispatches a question to its typed huh v2 field builder.
func buildField(q *Question, result *WizardResult, locale *string) huh.Field {
	switch q.Type {
	case QuestionTypeSelect:
		return buildSelectField(q, result, locale)
	case QuestionTypeInput:
		return buildInputField(q, result, locale)
	case QuestionTypeConfirm:
		return buildConfirmField(q, result, locale)
	default:
		// Defensive: QuestionType is a closed 3-value enum; treat unknown
		// values as free-text input rather than crashing the form.
		return buildInputField(q, result, locale)
	}
}

// stepperDenominator is the single dynamic source for the wizard stepper
// denominator (REQ-TUX2-008): the count of currently-visible questions.
// It replaces the removed hardcoded 6-step constant (AC-CLI-TUI-007 successor).
func stepperDenominator(questions []Question, result *WizardResult) int {
	return TotalVisibleQuestions(questions, result)
}

// visibleQuestionIndex returns the 1-based position of the question with the
// given ID among the currently-visible questions (stepper numerator).
func visibleQuestionIndex(questions []Question, result *WizardResult, id string) int {
	idx := 0
	for i := range questions {
		q := &questions[i]
		if q.Condition != nil && !q.Condition(result) {
			continue
		}
		idx++
		if q.ID == id {
			return idx
		}
	}
	return idx
}

// stepperNote renders the live step indicator above each group (AC-CLI-TUI-007
// succession). The huh v2 eval mechanism re-computes TitleFunc whenever the
// bound result struct changes (hashstructure deep-hash), so numerator and
// denominator both track conditional-question visibility as answers land.
func stepperNote(questions []Question, first *Question, result *WizardResult) *huh.Note {
	id := first.ID
	return huh.NewNote().TitleFunc(func() string {
		return tui.Stepper(
			visibleQuestionIndex(questions, result, id),
			stepperDenominator(questions, result),
			nil,
		)
	}, result)
}

// buildSelectField creates a huh.Select field for a select-type question.
func buildSelectField(q *Question, result *WizardResult, locale *string) *huh.Select[string] {
	var selected string

	// Set default value as initial selection.
	if q.Default != "" {
		selected = q.Default
	}

	// Build options eagerly at form-construction time using the current
	// locale. Static Options() keep the select on the defect-free layout
	// path (the M2a spike verified huh v2 clamps scrolling minimally via
	// ensureVisible; OptionsFunc is still unnecessary here because the
	// option sets are fixed).
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

	// Wire up value storage (huh runs Validate on field completion/blur).
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
	case "plan_type":
		result.PlanType = value
	case "development_mode":
		result.DevelopmentMode = value
	case "report_format":
		result.ReportFormat = value
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
	// Phase 1 fields (REQ-IWE-001..005)
	case "project_mode":
		result.ProjectMode = value
	case "harness_profile":
		result.HarnessProfile = value
	}
	_ = locale // locale is kept for GetLocalizedQuestion compatibility
}

// saveBoolAnswer stores a boolean answer in the result.
func saveBoolAnswer(id string, value bool, result *WizardResult) {
	switch id {
	case "lsp_enabled":
		result.LSPEnabled = value
	case "enforce_quality":
		result.EnforceQuality = value
	case "coverage_exemptions_enabled":
		result.CoverageExemptionsEnabled = value
	case "design_enabled":
		result.DesignEnabled = value
	case "claude_design_enabled":
		result.ClaudeDesignEnabled = value
	}
}

// buildConfirmField creates a huh.Confirm field for a boolean question.
func buildConfirmField(q *Question, result *WizardResult, locale *string) *huh.Confirm {
	// Parse default value
	value := q.Default == "true"

	conf := huh.NewConfirm().
		TitleFunc(func() string {
			lq := GetLocalizedQuestion(q, *locale)
			return lq.Title
		}, locale).
		DescriptionFunc(func() string {
			lq := GetLocalizedQuestion(q, *locale)
			return lq.Description
		}, locale).
		Value(&value)

	qID := q.ID
	conf = conf.Validate(func(v bool) error {
		saveBoolAnswer(qID, v, result)
		return nil
	})

	return conf
}

// newMoAIWizardTheme adapts the MoAI wizard styles to the huh v2 Theme
// interface: huh resolves the terminal background once and passes the isDark
// axis to moaiWizardStyles (this replaces the lipgloss v1 AdaptiveColor
// mechanism used before the v2 migration).
//
// @MX:ANCHOR: [AUTO] Consumed by every buildUnifiedForm invocation; single huh.Theme factory
// @MX:REASON: All wizard groups share one theme; changes here affect the entire init wizard UX
func newMoAIWizardTheme() huh.Theme {
	return huh.ThemeFunc(moaiWizardStyles)
}

// moaiWizardStyles builds the MoAI-branded huh v2 style set for the resolved
// background. All colour values are derived from internal/tui LightTheme /
// DarkTheme tokens (AC-CLI-TUI-013: no hex literals outside internal/tui/).
func moaiWizardStyles(isDark bool) *huh.Styles {
	t := huh.ThemeBase(isDark)
	c := wizardTokens(isDark)

	fg := func(token string) color.Color { return lipgloss.Color(token) }

	t.Focused.Base = t.Focused.Base.BorderForeground(fg(c.Border))
	t.Focused.Card = t.Focused.Base
	t.Focused.Title = t.Focused.Title.Foreground(fg(c.Primary)).Bold(true)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(fg(c.Primary)).Bold(true).MarginBottom(1)
	t.Focused.Description = t.Focused.Description.Foreground(fg(c.Body))
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(fg(c.Error))
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(fg(c.Error))
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(fg(c.Primary)).SetString("▸ ")
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(fg(c.Primary))
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(fg(c.Primary))
	t.Focused.Option = t.Focused.Option.Foreground(fg(c.Text))
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(fg(c.Primary))
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(fg(c.Success))
	t.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(fg(c.Success)).SetString("◆ ")
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(fg(c.Text))
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(fg(c.Muted)).SetString("◇ ")
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(fg(c.Primary))
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(fg(c.Muted))
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(fg(c.Secondary))
	t.Focused.FocusedButton = t.Focused.FocusedButton.
		Foreground(fg(c.ButtonFg)).
		Background(fg(c.Primary))
	t.Focused.BlurredButton = t.Focused.BlurredButton.
		Foreground(fg(c.Text)).
		Background(fg(c.ButtonBlurredBg))
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

// wizardTokenSet holds resolved tui colour tokens for one background axis.
type wizardTokenSet struct {
	Primary         string
	Secondary       string
	Success         string
	Warning         string
	Error           string
	Muted           string
	Body            string
	Text            string
	Border          string
	ButtonFg        string
	ButtonBlurredBg string
}

// wizardTokens resolves the internal/tui token set for the given background.
// The token-to-role mapping mirrors wizardColors (styles.go): Primary=Accent
// (Claude coral), Secondary=Info, Error=Danger, Muted=Dim, Border=Rule.
func wizardTokens(isDark bool) wizardTokenSet {
	th := tui.LightTheme()
	buttonFg := th.Bg // white-ish text on filled button (light background)
	if isDark {
		th = tui.DarkTheme()
		buttonFg = th.Fg
	}
	return wizardTokenSet{
		Primary:         th.Accent,
		Secondary:       th.Info,
		Success:         th.Success,
		Warning:         th.Warning,
		Error:           th.Danger,
		Muted:           th.Dim,
		Body:            th.Body,
		Text:            th.Fg,
		Border:          th.Rule,
		ButtonFg:        buttonFg,
		ButtonBlurredBg: th.Panel,
	}
}
