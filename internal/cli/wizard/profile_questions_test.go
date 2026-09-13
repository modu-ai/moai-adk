package wizard

import (
	"reflect"
	"testing"
)

// profileFieldIDs is the REQ-ITI-004 field set in design.md §2.2 order: the
// nine profile preferences plus the project development_mode.
var profileFieldIDs = []string{
	"conversation_language",
	"user_name",
	"git_commit_lang",
	"code_comment_lang",
	"doc_lang",
	"model",
	"model_policy",
	"effort_level",
	"permission_mode",
	"development_mode",
}

// sampleProfileOptions returns distinct, recognisable option lists so a
// question wired to the wrong list is visible in the assertion output.
func sampleProfileOptions() ProfileOptions {
	return ProfileOptions{
		Language: []Option{
			{Label: "English", Value: "en", Desc: "English"},
			{Label: "Korean (한국어)", Value: "ko", Desc: "한국어"},
		},
		Model:           []Option{{Label: "(runtime default)", Value: ""}, {Label: "opus", Value: "opus"}},
		ModelPolicy:     []Option{{Label: "(none)", Value: ""}, {Label: "High", Value: "high"}},
		EffortLevel:     []Option{{Label: "(runtime default)", Value: ""}, {Label: "max", Value: "max"}},
		PermissionMode:  []Option{{Label: "Auto accept edits", Value: "acceptEdits"}, {Label: "Plan", Value: "plan"}},
		DevelopmentMode: []Option{{Label: "(project default)", Value: ""}, {Label: "TDD", Value: "tdd"}},
	}
}

func questionIDs(qs []Question) []string {
	ids := make([]string, len(qs))
	for i := range qs {
		ids[i] = qs[i].ID
	}
	return ids
}

func TestProfileQuestions_IDSetMatchesProfileFieldSet(t *testing.T) {
	qs := ProfileQuestions(sampleProfileOptions(), ProfileResult{})
	if len(qs) == 0 {
		t.Fatal("ProfileQuestions returned no questions")
	}
	if got := questionIDs(qs); !reflect.DeepEqual(got, profileFieldIDs) {
		t.Fatalf("profile question ids (ordered):\n got  %v\n want %v", got, profileFieldIDs)
	}
	if got := ProfileQuestionIDs(); !reflect.DeepEqual(got, profileFieldIDs) {
		t.Fatalf("ProfileQuestionIDs():\n got  %v\n want %v", got, profileFieldIDs)
	}
}

func TestProfileQuestions_EveryQuestionVisible(t *testing.T) {
	qs := ProfileQuestions(sampleProfileOptions(), ProfileResult{})
	if len(qs) != len(profileFieldIDs) {
		t.Fatalf("want %d questions, got %d", len(profileFieldIDs), len(qs))
	}
	for i := range qs {
		if qs[i].Condition != nil {
			t.Errorf("question %q carries a Condition; the profile set has no conditional question (design.md §2.2, N = 10)", qs[i].ID)
		}
		if qs[i].Title == "" || qs[i].Description == "" {
			t.Errorf("question %q: empty title or description", qs[i].ID)
		}
	}
}

func TestProfileQuestions_GroupPartition(t *testing.T) {
	want := map[string]string{
		"conversation_language": "profile-language",
		"user_name":             "profile-identity",
		"git_commit_lang":       "profile-languages",
		"code_comment_lang":     "profile-languages",
		"doc_lang":              "profile-languages",
		"model":                 "profile-model",
		"model_policy":          "profile-model",
		"effort_level":          "profile-model",
		"permission_mode":       "profile-model",
		"development_mode":      "profile-project",
	}
	qs := ProfileQuestions(sampleProfileOptions(), ProfileResult{})
	if len(qs) != len(want) {
		t.Fatalf("want %d questions, got %d", len(want), len(qs))
	}
	for i := range qs {
		if got := qs[i].Group; got != want[qs[i].ID] {
			t.Errorf("question %q: group %q, want %q", qs[i].ID, got, want[qs[i].ID])
		}
	}
	// conversation_language must be alone on the first page: its answer has to
	// be saved (field blur) before the next group renders (REQ-ITI-007).
	if qs[1].Group == qs[0].Group {
		t.Errorf("conversation_language shares its group with %q; it must be alone on the first page", qs[1].ID)
	}
}

func TestProfileQuestions_TypesAndOptionsPassThrough(t *testing.T) {
	opts := sampleProfileOptions()
	want := map[string][]Option{
		"conversation_language": opts.Language,
		"git_commit_lang":       opts.Language,
		"code_comment_lang":     opts.Language,
		"doc_lang":              opts.Language,
		"model":                 opts.Model,
		"model_policy":          opts.ModelPolicy,
		"effort_level":          opts.EffortLevel,
		"permission_mode":       opts.PermissionMode,
		"development_mode":      opts.DevelopmentMode,
	}
	qs := ProfileQuestions(opts, ProfileResult{})
	if len(qs) != len(profileFieldIDs) {
		t.Fatalf("want %d questions, got %d", len(profileFieldIDs), len(qs))
	}
	for i := range qs {
		q := qs[i]
		if q.ID == "user_name" {
			if q.Type != QuestionTypeInput {
				t.Errorf("user_name: type %v, want input", q.Type)
			}
			if len(q.Options) != 0 {
				t.Errorf("user_name: input question carries %d options", len(q.Options))
			}
			continue
		}
		if q.Type != QuestionTypeSelect {
			t.Errorf("%s: type %v, want select", q.ID, q.Type)
		}
		if !reflect.DeepEqual(q.Options, want[q.ID]) {
			t.Errorf("%s: options\n got  %+v\n want %+v", q.ID, q.Options, want[q.ID])
		}
	}
}

func TestProfileQuestions_OptionSlicesAreCopies(t *testing.T) {
	opts := sampleProfileOptions()
	qs := ProfileQuestions(opts, ProfileResult{})
	if len(qs) < 3 || len(qs[0].Options) == 0 || len(qs[2].Options) == 0 {
		t.Fatalf("need conversation_language and git_commit_lang with options, got %d questions", len(qs))
	}
	qs[0].Options[0].Label = "MUTATED"
	if opts.Language[0].Label == "MUTATED" {
		t.Fatal("mutating a question option wrote through to the caller's option list")
	}
	if qs[2].Options[0].Label == "MUTATED" {
		t.Fatal("two language questions share one option slice")
	}
}

func TestProfileQuestions_InitialValuesBecomeDefaults(t *testing.T) {
	initial := ProfileResult{
		ConversationLang: "ko",
		UserName:         "GOOS",
		GitCommitLang:    "en",
		CodeCommentLang:  "ja",
		DocLang:          "zh",
		Model:            "opus",
		ModelPolicy:      "high",
		EffortLevel:      "max",
		PermissionMode:   "plan",
		DevelopmentMode:  "tdd",
	}
	want := map[string]string{
		"conversation_language": "ko",
		"user_name":             "GOOS",
		"git_commit_lang":       "en",
		"code_comment_lang":     "ja",
		"doc_lang":              "zh",
		"model":                 "opus",
		"model_policy":          "high",
		"effort_level":          "max",
		"permission_mode":       "plan",
		"development_mode":      "tdd",
	}
	qs := ProfileQuestions(sampleProfileOptions(), initial)
	if len(qs) != len(want) {
		t.Fatalf("want %d questions, got %d", len(want), len(qs))
	}
	for i := range qs {
		if qs[i].Default != want[qs[i].ID] {
			t.Errorf("%s: default %q, want %q", qs[i].ID, qs[i].Default, want[qs[i].ID])
		}
	}
}
