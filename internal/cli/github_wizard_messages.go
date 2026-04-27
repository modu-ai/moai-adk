// Package cli provides Wizard multilingual messages.
package cli

// Messages contains multilingual messages for the GitHub Init Wizard.
type Messages struct {
	SelectLLM    string
	SuccessTitle string
	SuccessBody  string
}

// GetMessages returns messages for the specified language.
func GetMessages(lang string) *Messages {
	messages := map[string]*Messages{
		"ko": {
			SelectLLM:    "? 코드 리뷰에 사용할 LLM을 선택하세요 (여러 개 선택 가능):",
			SuccessTitle: "✅ GitHub Actions 초기화 완료!",
			SuccessBody:  "\n다음 단계:\n  1. LLM 인증: moai github auth claude\n  2. PR 생성 후 자동 리뷰 확인\n",
		},
		"en": {
			SelectLLM:    "? Select LLMs for code review (multiple selection available):",
			SuccessTitle: "✅ GitHub Actions initialization complete!",
			SuccessBody:  "\nNext steps:\n  1. LLM authentication: moai github auth claude\n  2. Create PR and verify auto-review\n",
		},
		"ja": {
			SelectLLM:    "? コードレビューに使用するLLMを選択してください（複数選択可能）:",
			SuccessTitle: "✅ GitHub Actions初期化完了!",
			SuccessBody:  "\n次のステップ:\n  1. LLM認証: moai github auth claude\n  2. PR作成後、自動レビュー確認\n",
		},
		"zh": {
			SelectLLM:    "? 选择用于代码审查的 LLM（可多选）:",
			SuccessTitle: "✅ GitHub Actions 初始化完成!",
			SuccessBody:  "\n下一步:\n  1. LLM 认证: moai github auth claude\n  2. 创建 PR 后验证自动审查\n",
		},
	}

	if msg, ok := messages[lang]; ok {
		return msg
	}
	return messages["en"] // Default to English
}
