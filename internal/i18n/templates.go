package i18n

import (
	"bytes"
	"fmt"
	"text/template"
	"time"
)

// CommentData holds the template variables for comment generation.
type CommentData struct {
	// Summary describes the implementation changes.
	Summary string

	// PRNumber is the pull request number.
	PRNumber int

	// IssueNumber is the original GitHub issue number.
	IssueNumber int

	// MergedAt is the timestamp when the PR was merged.
	MergedAt time.Time

	// TimeZone is the display timezone label (e.g., "KST", "UTC").
	TimeZone string

	// CoveragePercent is the test coverage percentage. Zero means omitted.
	CoveragePercent int
}

// CommentGenerator generates multilingual comments for GitHub issues.
type CommentGenerator interface {
	// Generate produces a comment string in the specified language.
	// Falls back to English if the language code is not supported.
	// Returns ErrInvalidData if data is nil.
	Generate(langCode string, data *CommentData) (string, error)
}

// TemplateCommentGenerator implements CommentGenerator using text/template.
type TemplateCommentGenerator struct {
	templates map[string]*template.Template
}

// Compile-time interface check.
var _ CommentGenerator = (*TemplateCommentGenerator)(nil)

// NewCommentGenerator creates a new TemplateCommentGenerator with all
// supported language templates pre-parsed.
func NewCommentGenerator() *TemplateCommentGenerator {
	g := &TemplateCommentGenerator{
		templates: make(map[string]*template.Template),
	}

	for lang, tmplStr := range commentTemplates {
		g.templates[lang] = template.Must(
			template.New(lang).Parse(tmplStr),
		)
	}

	return g
}

// Generate produces a comment string in the specified language.
// Falls back to English if the language code is not supported.
func (g *TemplateCommentGenerator) Generate(langCode string, data *CommentData) (string, error) {
	if data == nil {
		return "", ErrInvalidData
	}

	tmpl, ok := g.templates[langCode]
	if !ok {
		tmpl = g.templates[fallbackLang]
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("%w: %v", ErrTemplateExecution, err)
	}

	return buf.String(), nil
}

// fallbackLang is the default language when the requested code is unsupported.
const fallbackLang = "en"

// commentTemplates holds Go text/template strings keyed by language code.
var commentTemplates = map[string]string{
	"en": `✅ This issue has been resolved successfully!

**Implementation Summary:**
{{.Summary}}
{{- if gt .CoveragePercent 0}}

**Test Coverage:** {{.CoveragePercent}}%
{{- end}}

**Related PR:** #{{.PRNumber}}
**Merged at:** {{.MergedAt.Format "2006-01-02 15:04"}} {{.TimeZone}}

This issue is being closed automatically. If you encounter further problems, please open a new issue.`,

	"ko": `✅ 이슈가 성공적으로 해결되었습니다!

**구현 내용:**
{{.Summary}}
{{- if gt .CoveragePercent 0}}

**테스트 커버리지:** {{.CoveragePercent}}%
{{- end}}

**관련 PR:** #{{.PRNumber}}
**병합 시간:** {{.MergedAt.Format "2006-01-02 15:04"}} {{.TimeZone}}

이슈를 자동으로 종료합니다. 추가 문제가 있으면 새 이슈를 생성해주세요.`,

	"ja": `✅ このイシューは正常に解決されました!

**実装内容:**
{{.Summary}}
{{- if gt .CoveragePercent 0}}

**テストカバレッジ:** {{.CoveragePercent}}%
{{- end}}

**関連PR:** #{{.PRNumber}}
**マージ日時:** {{.MergedAt.Format "2006-01-02 15:04"}} {{.TimeZone}}

このイシューを自動的にクローズします。問題が発生した場合は、新しいイシューを作成してください。`,

	"zh": `✅ 此问题已成功解决！

**实现内容：**
{{.Summary}}
{{- if gt .CoveragePercent 0}}

**测试覆盖率：** {{.CoveragePercent}}%
{{- end}}

**相关PR：** #{{.PRNumber}}
**合并时间：** {{.MergedAt.Format "2006-01-02 15:04"}} {{.TimeZone}}

此问题将自动关闭。如遇到其他问题，请创建新的issue。`,
}
