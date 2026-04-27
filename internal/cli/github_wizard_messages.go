// Package cli provides Wizard multilingual messages.
package cli

// Messages contains multilingual messages for the GitHub Init Wizard.
type Messages struct {
	SelectLLM    string
	SuccessTitle string
	SuccessBody  string

	// Model selection prompts
	SelectClaudeModel  string
	SelectCodexModel   string
	SelectGeminiModel  string
	SelectZAIModel     string

	// Yes/No labels
	YesLabel string
	NoLabel  string

	// Model descriptions (localized)
	ModelNames map[string]string
}

// GetMessages returns messages for the specified language.
func GetMessages(lang string) *Messages {
	messages := map[string]*Messages{
		"ko": {
			SelectLLM:         "? 코드 리뷰에 사용할 LLM을 선택하세요 (여러 개 선택 가능):",
			SuccessTitle:      "✅ GitHub Actions 초기화 완료!",
			SuccessBody:       "\n다음 단계:\n  1. LLM 인증: moai github auth claude\n  2. PR 생성 후 자동 리뷰 확인\n",
			SelectClaudeModel: "? Claude 모델을 선택하세요:",
			SelectCodexModel:  "? OpenAI 모델을 선택하세요:",
			SelectGeminiModel: "? Gemini 모델을 선택하세요:",
			SelectZAIModel:    "? GLM (Z.AI) 모델을 선택하세요:",
			YesLabel:          "예",
			NoLabel:           "아니오",
			ModelNames: map[string]string{
				"claude-opus-4-7":  "Claude Opus 4.7 (최고 성능)",
				"claude-sonnet-4-6": "Claude Sonnet 4.6 (균형)",
				"gpt-5.5":         "GPT-5.5 (최신)",
				"gpt-5.3-codex":   "GPT-5.3-Codex (코드 최적화)",
				"gemini-pro-latest": "Gemini Pro (최신)",
				"glm-5.1":         "GLM-5.1 (최신)",
				"glm-4.7":         "GLM-4.7 (추론 특화)",
			},
		},
		"en": {
			SelectLLM:         "? Select LLMs for code review (multiple selection available):",
			SuccessTitle:      "✅ GitHub Actions initialization complete!",
			SuccessBody:       "\nNext steps:\n  1. LLM authentication: moai github auth claude\n  2. Create PR and verify auto-review\n",
			SelectClaudeModel: "? Select model for Claude:",
			SelectCodexModel:  "? Select model for OpenAI:",
			SelectGeminiModel: "? Select model for Gemini:",
			SelectZAIModel:    "? Select model for GLM (Z.AI):",
			YesLabel:          "Yes",
			NoLabel:           "No",
			ModelNames: map[string]string{
				"claude-opus-4-7":  "Claude Opus 4.7 (Highest performance)",
				"claude-sonnet-4-6": "Claude Sonnet 4.6 (Balanced)",
				"gpt-5.5":         "GPT-5.5 (Latest)",
				"gpt-5.3-codex":   "GPT-5.3-Codex (Code-optimized)",
				"gemini-pro-latest": "Gemini Pro (Latest)",
				"glm-5.1":         "GLM-5.1 (Latest)",
				"glm-4.7":         "GLM-4.7 (Reasoning)",
			},
		},
		"ja": {
			SelectLLM:         "? コードレビューに使用するLLMを選択してください（複数選択可能）:",
			SuccessTitle:      "✅ GitHub Actions初期化完了!",
			SuccessBody:       "\n次のステップ:\n  1. LLM認証: moai github auth claude\n  2. PR作成後、自動レビュー確認\n",
			SelectClaudeModel: "? Claudeモデルを選択してください:",
			SelectCodexModel:  "? OpenAIモデルを選択してください:",
			SelectGeminiModel: "? Geminiモデルを選択してください:",
			SelectZAIModel:    "? GLM (Z.AI)モデルを選択してください:",
			YesLabel:          "はい",
			NoLabel:           "いいえ",
			ModelNames: map[string]string{
				"claude-opus-4-7":  "Claude Opus 4.7 (最高性能)",
				"claude-sonnet-4-6": "Claude Sonnet 4.6 (バランス)",
				"gpt-5.5":         "GPT-5.5 (最新)",
				"gpt-5.3-codex":   "GPT-5.3-Codex (コード最適化)",
				"gemini-pro-latest": "Gemini Pro (最新)",
				"glm-5.1":         "GLM-5.1 (最新)",
				"glm-4.7":         "GLM-4.7 (推論特化)",
			},
		},
		"zh": {
			SelectLLM:         "? 选择用于代码审查的 LLM（可多选）:",
			SuccessTitle:      "✅ GitHub Actions 初始化完成!",
			SuccessBody:       "\n下一步:\n  1. LLM 认证: moai github auth claude\n  2. 创建 PR 后验证自动审查\n",
			SelectClaudeModel: "? 选择 Claude 模型:",
			SelectCodexModel:  "? 选择 OpenAI 模型:",
			SelectGeminiModel: "? 选择 Gemini 模型:",
			SelectZAIModel:    "? 选择 GLM (Z.AI) 模型:",
			YesLabel:          "是",
			NoLabel:           "否",
			ModelNames: map[string]string{
				"claude-opus-4-7":  "Claude Opus 4.7 (最高性能)",
				"claude-sonnet-4-6": "Claude Sonnet 4.6 (均衡)",
				"gpt-5.5":         "GPT-5.5 (最新)",
				"gpt-5.3-codex":   "GPT-5.3-Codex (代码优化)",
				"gemini-pro-latest": "Gemini Pro (最新)",
				"glm-5.1":         "GLM-5.1 (最新)",
				"glm-4.7":         "GLM-4.7 (推理优化)",
			},
		},
	}

	if msg, ok := messages[lang]; ok {
		return msg
	}
	return messages["en"]
}
