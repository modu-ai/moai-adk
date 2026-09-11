package wizard

// ProfileOptions carries the option lists for the profile wizard select fields.
//
// The cli package builds these lists — values from the settings schema, labels
// through its label bridge — and passes them in, so this package never resolves
// schema labels itself and the cli → wizard import direction holds
// (design.md §3). Language feeds conversation_language and the three
// artifact-language questions.
type ProfileOptions struct {
	Language        []Option
	Model           []Option
	ModelPolicy     []Option
	EffortLevel     []Option
	PermissionMode  []Option
	DevelopmentMode []Option
}

// ProfileResult is the profile wizard's own answer set: the nine profile
// preference fields plus the project development_mode. It is deliberately a
// separate type from WizardResult so profile and init answers never share one
// struct (design.md §4). As the argument to ProfileQuestions it supplies each
// question's initial (pre-selected) value.
type ProfileResult struct {
	ConversationLang string
	UserName         string
	GitCommitLang    string
	CodeCommentLang  string
	DocLang          string
	Model            string
	ModelPolicy      string
	EffortLevel      string
	PermissionMode   string
	DevelopmentMode  string
}

// Profile question group labels. Labels are partition keys only — the form
// never renders them — so they name the page, not a translatable heading.
// conversation_language sits alone on the first page so its answer is saved
// (field blur) before the next page renders in the chosen language.
const (
	profileGroupLanguage  = "profile-language"
	profileGroupIdentity  = "profile-identity"
	profileGroupLanguages = "profile-languages"
	profileGroupModel     = "profile-model"
	profileGroupProject   = "profile-project"
)

// profileQuestionSpec is one row of the profile question table.
type profileQuestionSpec struct {
	id    string
	typ   QuestionType
	group string
	opts  func(ProfileOptions) []Option
	value func(ProfileResult) string
}

// profileQuestionTable is the single ordered definition of the profile question
// set (design.md §2.2). ProfileQuestionIDs and ProfileQuestions both read it.
var profileQuestionTable = []profileQuestionSpec{
	{"conversation_language", QuestionTypeSelect, profileGroupLanguage,
		func(o ProfileOptions) []Option { return o.Language }, func(r ProfileResult) string { return r.ConversationLang }},
	{"user_name", QuestionTypeInput, profileGroupIdentity,
		nil, func(r ProfileResult) string { return r.UserName }},
	{"git_commit_lang", QuestionTypeSelect, profileGroupLanguages,
		func(o ProfileOptions) []Option { return o.Language }, func(r ProfileResult) string { return r.GitCommitLang }},
	{"code_comment_lang", QuestionTypeSelect, profileGroupLanguages,
		func(o ProfileOptions) []Option { return o.Language }, func(r ProfileResult) string { return r.CodeCommentLang }},
	{"doc_lang", QuestionTypeSelect, profileGroupLanguages,
		func(o ProfileOptions) []Option { return o.Language }, func(r ProfileResult) string { return r.DocLang }},
	{"model", QuestionTypeSelect, profileGroupModel,
		func(o ProfileOptions) []Option { return o.Model }, func(r ProfileResult) string { return r.Model }},
	{"model_policy", QuestionTypeSelect, profileGroupModel,
		func(o ProfileOptions) []Option { return o.ModelPolicy }, func(r ProfileResult) string { return r.ModelPolicy }},
	{"effort_level", QuestionTypeSelect, profileGroupModel,
		func(o ProfileOptions) []Option { return o.EffortLevel }, func(r ProfileResult) string { return r.EffortLevel }},
	{"permission_mode", QuestionTypeSelect, profileGroupModel,
		func(o ProfileOptions) []Option { return o.PermissionMode }, func(r ProfileResult) string { return r.PermissionMode }},
	{"development_mode", QuestionTypeSelect, profileGroupProject,
		func(o ProfileOptions) []Option { return o.DevelopmentMode }, func(r ProfileResult) string { return r.DevelopmentMode }},
}

// ProfileQuestionIDs returns the profile question ids in display order.
func ProfileQuestionIDs() []string {
	ids := make([]string, len(profileQuestionTable))
	for i, spec := range profileQuestionTable {
		ids[i] = spec.id
	}
	return ids
}

// ProfileQuestions returns the profile wizard question set: ten unconditional
// questions whose English title/description come from the profile translation
// table, whose option lists are copies of the caller's lists, and whose
// defaults are the caller's initial values.
//
// @MX:NOTE: [AUTO] Not wired yet — the cli profile setup still runs its own
// form; the v2 absorption (design.md §2.2) routes it through this set.
func ProfileQuestions(opts ProfileOptions, initial ProfileResult) []Question {
	en := profileQuestionTexts["en"]
	qs := make([]Question, len(profileQuestionTable))
	for i, spec := range profileQuestionTable {
		q := Question{
			ID:          spec.id,
			Type:        spec.typ,
			Title:       en[spec.id].Title,
			Description: en[spec.id].Description,
			Default:     spec.value(initial),
			Group:       spec.group,
		}
		if spec.opts != nil {
			q.Options = append([]Option(nil), spec.opts(opts)...)
		}
		qs[i] = q
	}
	return qs
}
