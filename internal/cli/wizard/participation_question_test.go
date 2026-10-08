package wizard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setParticipationHome points MOAI_HOME at a temporary directory, optionally
// seeding a user-scoped participation file. The real home is never touched.
func setParticipationHome(t *testing.T, body string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	if body != "" {
		dir := filepath.Join(home, "config")
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("mkdir config: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "participation.yaml"), []byte(body), 0o600); err != nil {
			t.Fatalf("write participation.yaml: %v", err)
		}
	}
}

func TestParticipationQuestion(t *testing.T) {
	setParticipationHome(t, "")
	root := t.TempDir()

	initQs := InitQuestions(root)
	count := 0
	idx := -1
	jevIdx := -1
	groups := map[string]int{}
	for i, q := range initQs {
		if q.ID == ParticipationQuestionID {
			count++
			idx = i
			groups[q.Group]++
		}
		if q.ID == JevQuestionID {
			jevIdx = i
		}
	}
	if count != 1 {
		t.Fatalf("InitQuestions carries %d participation questions, want exactly 1", count)
	}
	if idx < jevIdx {
		t.Fatalf("participation question at %d is not after the Jev question at %d", idx, jevIdx)
	}
	if groups[initQs[idx].Group] != 1 {
		t.Fatalf("participation group %q is shared: %d questions carry it", initQs[idx].Group, groups[initQs[idx].Group])
	}

	pq := QuestionByID(initQs, ParticipationQuestionID)
	if pq.Type != QuestionTypeConfirm {
		t.Fatalf("participation question type = %v, want confirm", pq.Type)
	}
	if pq.Default != "false" {
		t.Fatalf("participation default = %q, want false with no stored value", pq.Default)
	}
	if pq.Required {
		t.Fatal("participation question is Required; it must default to a decline")
	}

	// The stored user value pre-selects the default: re-running init over an
	// existing consent never silently withdraws it.
	setParticipationHome(t, "participation:\n  enabled: true\n  asked: true\n")
	pq2 := QuestionByID(Page3Questions(root), ParticipationQuestionID)
	if pq2 == nil {
		t.Fatal("Page3Questions lost the participation question")
	}
	if pq2.Default != "true" {
		t.Fatalf("participation default = %q with a stored true, want true", pq2.Default)
	}

	// Absent from the shared and reconfigure sets.
	for _, q := range DefaultQuestions(root) {
		if q.ID == ParticipationQuestionID {
			t.Fatal("participation question leaked into DefaultQuestions")
		}
	}
	for _, q := range ReconfigureQuestions(root) {
		if q.ID == ParticipationQuestionID {
			t.Fatal("participation question leaked into ReconfigureQuestions")
		}
	}
	// The removed-question set still removes feedback_auto_submit and never
	// removes the new question.
	if !sliceContains(removedQuietInit, "feedback_auto_submit") {
		t.Fatal("feedback_auto_submit is no longer in the removed set")
	}
	if sliceContains(removedQuietInit, ParticipationQuestionID) {
		t.Fatal("feedback_participation must not be in the removed set")
	}
}

func sliceContains(ss []string, s string) bool {
	for _, got := range ss {
		if got == s {
			return true
		}
	}
	return false
}

// TestParticipationQuestion_FourLocalesStateFacts asserts that each of the
// four wizard locales states every fact REQ-ANON-005 requires, through
// locale-specific fact phrases (one per required statement).
func TestParticipationQuestion_FourLocalesStateFacts(t *testing.T) {
	setParticipationHome(t, "")
	root := t.TempDir()

	enQ := QuestionByID(InitQuestions(root), ParticipationQuestionID)
	if enQ == nil {
		t.Fatal("no participation question in the init set")
	}

	locales := []struct {
		name      string
		title     string
		desc      string
		factParts []string
	}{
		{
			name:  "en",
			title: enQ.Title,
			desc:  enQ.Description,
			factParts: []string{
				"public GitHub issue",                 // fact 1: public, tied to the account
				"your own GitHub account",             // fact 1b
				"creation time",                       // fact 2: creation time is public
				"no per-report confirmation",          // fact 3: automatic
				"later comments",                      // fact 4: subscription
				"publicly associated",                 // fact 5: public association with use
				"fixed fields",                        // fact 6: what is sent
				"moai feedback participation preview", // fact 7a: preview command
				"moai web",                            // fact 7b: console surface
				"cannot be recalled",                  // fact 8: irrevocable
				"model-subscription tokens",           // fact 9a: token spend
				"template text",                       // fact 9b: fallback
			},
		},
		{
			name:  "ko",
			title: translations["ko"][ParticipationQuestionID].Title,
			desc:  translations["ko"][ParticipationQuestionID].Description,
			factParts: []string{
				"공개 이슈",
				"GitHub 계정",
				"생성 시각",
				"확인을 받지 않고",
				"구독",
				"공개적으로 연결",
				"고정 필드",
				"moai feedback participation preview",
				"moai web",
				"취소할 수 없습니다",
				"모델 구독 토큰",
				"템플릿 본문",
			},
		},
		{
			name:  "ja",
			title: translations["ja"][ParticipationQuestionID].Title,
			desc:  translations["ja"][ParticipationQuestionID].Description,
			factParts: []string{
				"公開イシュー",
				"GitHub アカウント",
				"作成日時",
				"確認はありません",
				"購読",
				"公開的に結び付",
				"固定フィールド",
				"moai feedback participation preview",
				"moai web",
				"取り消すことはできません",
				"モデル購読トークン",
				"テンプレート本文",
			},
		},
		{
			name:  "zh",
			title: translations["zh"][ParticipationQuestionID].Title,
			desc:  translations["zh"][ParticipationQuestionID].Description,
			factParts: []string{
				"公开 issue",
				"GitHub 账号",
				"创建时间",
				"确认",
				"订阅",
				"公开关联",
				"固定字段",
				"moai feedback participation preview",
				"moai web",
				"无法",
				"模型订阅",
				"模板文本",
			},
		},
	}

	for _, loc := range locales {
		if strings.TrimSpace(loc.title) == "" {
			t.Errorf("locale %s: empty participation title", loc.name)
		}
		combined := loc.title + "\n" + loc.desc
		for _, fact := range loc.factParts {
			if !strings.Contains(combined, fact) {
				t.Errorf("locale %s: participation text misses required fact phrase %q", loc.name, fact)
			}
		}
	}
}
