package wizard

import (
	"strings"

	"charm.land/huh/v2"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// RunProfile runs the absorbed profile wizard (design.md §2.2): the ten
// profile questions as ONE multi-group huh v2 form, each group led by the
// shared step indicator, the help line from the localized key map, and every
// title/description re-localized when the conversation-language answer lands
// (REQ-ITI-007). The initial values arrive as ProfileResult — they become each
// select's pre-selection — and the answers come back the same way.
//
// On any form error — cancellation included — the answers collected so far are
// returned alongside the error, so the caller can name the locale the user had
// reached when the run ended (design.md §5).
func RunProfile(opts ProfileOptions, initial ProfileResult, locale string) (*ProfileResult, error) {
	result := initial
	form := NewProfileForm(opts, initial, locale)
	if err := form.Run(); err != nil {
		return &result, mapFormErr(err)
	}
	return &result, nil
}

// NewProfileForm builds the absorbed profile wizard's form without running
// it: the ten profile questions partitioned by group, each group led by the
// shared step indicator, the help line from the localized key map. RunProfile
// is the runner; this constructor exists so the rendering can be judged from
// the cli package, where the option-label bridge lives (AC-ITI-008 goldens).
func NewProfileForm(opts ProfileOptions, initial ProfileResult, locale string) *huh.Form {
	questions := ProfileQuestions(opts, initial)
	result := initial
	cur := locale
	return buildProfileForm(questions, &result, &cur)
}

// buildProfileForm assembles the profile form: the question set partitioned
// into huh groups by the question Group labels (design.md §2.2 —
// conversation_language alone on the first page so its answer is saved before
// the next group renders), each group headed by the shared step indicator.
func buildProfileForm(questions []Question, result *ProfileResult, locale *string) *huh.Form {
	vis := profileVisibility(questions)
	var groups []*huh.Group
	var pending []*Question
	pendingLabel := ""

	flush := func() {
		if len(pending) == 0 {
			return
		}
		fields := make([]huh.Field, 0, len(pending))
		for _, q := range pending {
			fields = append(fields, buildProfileField(q, result, locale))
		}
		// The step indicator rides the GROUP TITLE, not an in-content note:
		// huh v2 scrolls each group's viewport so the focused field sits at
		// the top, which cuts in-content note rows above the focused field
		// (measured with the formDriver — the same phenomenon class as
		// plan.md §G D4), while the group header renders outside the
		// viewport and cannot be scrolled away. The string is the same
		// tui.Stepper(k, N, nil) the init wizard renders (design.md §4,
		// REQ-ITI-008).
		indicator := tui.Stepper(vis.index(pending[0].ID), vis.count(), nil)
		groups = append(groups, huh.NewGroup(fields...).Title(indicator))
		pending = nil
	}

	for i := range questions {
		q := &questions[i]
		if len(pending) > 0 && q.Group != pendingLabel {
			flush()
		}
		pendingLabel = q.Group
		pending = append(pending, q)
	}
	flush()

	return huh.NewForm(groups...).
		WithTheme(newMoAIWizardTheme()).
		WithKeyMap(localizedKeyMap(*locale)).
		WithAccessible(false)
}

// profileVisibility adapts the profile question set to the generalized
// stepper: the profile set has no conditional questions (design.md §2.2,
// N = 10), so every question is visible and the index is the 1-based display
// position.
func profileVisibility(questions []Question) questionVisibility {
	position := make(map[string]int, len(questions))
	for i := range questions {
		position[questions[i].ID] = i + 1
	}
	return questionVisibility{
		count: func() int { return len(questions) },
		index: func(id string) int { return position[id] },
	}
}

// buildProfileField dispatches a profile question to its typed field builder.
func buildProfileField(q *Question, result *ProfileResult, locale *string) huh.Field {
	if q.Type == QuestionTypeInput {
		return buildProfileInputField(q, result, locale)
	}
	return buildProfileSelectField(q, result, locale)
}

// profileHuhOptions converts a version-neutral option list to huh options,
// composing the same "Label - Desc" row text the init wizard's selects render.
func profileHuhOptions(opts []Option) []huh.Option[string] {
	out := make([]huh.Option[string], len(opts))
	for i, o := range opts {
		key := o.Label
		if o.Desc != "" {
			key = o.Label + " - " + o.Desc
		}
		out[i] = huh.NewOption(key, o.Value)
	}
	return out
}

// buildProfileSelectField creates the huh v2 select for one profile question.
// Title and description re-localize through the locale pointer; the option
// list is the caller's list already labelled in the run's locale (design.md
// §3 — the wizard never resolves schema labels itself).
func buildProfileSelectField(q *Question, result *ProfileResult, locale *string) *huh.Select[string] {
	selected := q.Default

	sel := huh.NewSelect[string]().
		TitleFunc(func() string {
			return LocalizeProfileQuestion(q, *locale).Title
		}, locale).
		DescriptionFunc(func() string {
			return LocalizeProfileQuestion(q, *locale).Description
		}, locale).
		Options(profileHuhOptions(q.Options)...).
		Value(&selected)

	// Wire up value storage (huh runs Validate on field completion/blur).
	sel.Validate(func(val string) error {
		saveProfileAnswer(q.ID, val, result, locale)
		return nil
	})

	return sel
}

// buildProfileInputField creates the huh v2 input for one profile question.
// The profile set's only input is user_name: optional, pre-filled with the
// stored value, and trimmed on save like the init wizard's inputs.
func buildProfileInputField(q *Question, result *ProfileResult, locale *string) *huh.Input {
	value := q.Default

	inp := huh.NewInput().
		TitleFunc(func() string {
			return LocalizeProfileQuestion(q, *locale).Title
		}, locale).
		DescriptionFunc(func() string {
			return LocalizeProfileQuestion(q, *locale).Description
		}, locale).
		Value(&value)

	if q.Default != "" {
		inp = inp.Placeholder(q.Default)
	}

	inp.Validate(func(val string) error {
		saveProfileAnswer(q.ID, strings.TrimSpace(val), result, locale)
		return nil
	})

	return inp
}

// saveProfileAnswer stores one answered profile question in the profile
// result. The conversation-language answer also moves the live locale pointer
// so every subsequent group re-renders in the chosen language (REQ-ITI-007).
func saveProfileAnswer(id, value string, result *ProfileResult, locale *string) {
	switch id {
	case "conversation_language":
		result.ConversationLang = value
		if locale != nil {
			*locale = value
		}
		return
	case "user_name":
		result.UserName = value
	case "git_commit_lang":
		result.GitCommitLang = value
	case "code_comment_lang":
		result.CodeCommentLang = value
	case "doc_lang":
		result.DocLang = value
	case "model":
		result.Model = value
	case "model_policy":
		result.ModelPolicy = value
	case "effort_level":
		result.EffortLevel = value
	case "permission_mode":
		result.PermissionMode = value
	case "development_mode":
		result.DevelopmentMode = value
	}
}
