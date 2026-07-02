package github

// i18n_helper_test.go — test-only multilingual comment generation helper.
//
// Relocated from the former internal/i18n package (SPEC-DEADPKG-INVESTIGATE-001):
// i18n had zero production importers and was consumed solely by the github
// integration tests (integration_test.go + issue_close_integration_test.go).
// Folding it into a package-github _test.go file keeps the CommentGenerator
// test-scoped (compiled only under `go test`) instead of shipping it as dead
// production code. Behavior is preserved verbatim, including the original
// unit-test coverage.

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"text/template"
	"time"
)

// Sentinel errors for comment generation. All errors can be checked with errors.Is().
var (
	// ErrTemplateNotFound indicates the requested template does not exist.
	ErrTemplateNotFound = errors.New("i18n: template not found")

	// ErrTemplateExecution indicates the template failed to render.
	ErrTemplateExecution = errors.New("i18n: template execution failed")

	// ErrInvalidData indicates the provided template data is invalid.
	ErrInvalidData = errors.New("i18n: invalid template data")
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

// --- Unit tests (relocated verbatim from internal/i18n/templates_test.go) ---

func fixedTime() time.Time {
	return time.Date(2026, 2, 16, 16, 30, 0, 0, time.UTC)
}

func sampleData() *CommentData {
	return &CommentData{
		Summary:         "Added user authentication feature",
		PRNumber:        456,
		IssueNumber:     123,
		MergedAt:        fixedTime(),
		TimeZone:        "KST",
		CoveragePercent: 92,
	}
}

func TestCommentGenerator_Generate(t *testing.T) {
	gen := NewCommentGenerator()

	tests := []struct {
		name     string
		lang     string
		data     *CommentData
		contains []string
		wantErr  bool
	}{
		{
			name: "english",
			lang: "en",
			data: sampleData(),
			contains: []string{
				"resolved successfully",
				"#456",
				"2026-02-16",
				"Added user authentication feature",
			},
		},
		{
			name: "korean",
			lang: "ko",
			data: sampleData(),
			contains: []string{
				"성공적으로 해결",
				"#456",
				"2026-02-16",
			},
		},
		{
			name: "japanese",
			lang: "ja",
			data: sampleData(),
			contains: []string{
				"解決されました",
				"#456",
				"2026-02-16",
			},
		},
		{
			name: "chinese",
			lang: "zh",
			data: sampleData(),
			contains: []string{
				"已成功解决",
				"#456",
				"2026-02-16",
			},
		},
		{
			name: "fallback for unsupported language",
			lang: "de",
			data: sampleData(),
			contains: []string{
				"resolved successfully",
				"#456",
			},
		},
		{
			name: "fallback for empty language",
			lang: "",
			data: sampleData(),
			contains: []string{
				"resolved successfully",
				"#456",
			},
		},
		{
			name:    "nil data returns error",
			lang:    "en",
			data:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.lang, tt.data)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("result missing %q\ngot:\n%s", want, result)
				}
			}
		})
	}
}

func TestCommentGenerator_Generate_IncludesTimestamp(t *testing.T) {
	gen := NewCommentGenerator()
	data := sampleData()

	result, err := gen.Generate("en", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "16:30") {
		t.Errorf("result missing time '16:30'\ngot:\n%s", result)
	}
	if !strings.Contains(result, "KST") {
		t.Errorf("result missing timezone 'KST'\ngot:\n%s", result)
	}
}

func TestCommentGenerator_Generate_IncludesCoverage(t *testing.T) {
	gen := NewCommentGenerator()
	data := sampleData()

	result, err := gen.Generate("en", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "92%") {
		t.Errorf("result missing coverage '92%%'\ngot:\n%s", result)
	}
}

func TestCommentGenerator_Generate_ZeroCoverageOmitted(t *testing.T) {
	gen := NewCommentGenerator()
	data := sampleData()
	data.CoveragePercent = 0

	result, err := gen.Generate("en", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With zero coverage, the coverage line should not appear
	if strings.Contains(result, "Coverage") || strings.Contains(result, "coverage") {
		// Allow the word in context, but "0%" specifically should not appear as a metric
		if strings.Contains(result, "0%") {
			t.Errorf("result should not show 0%% coverage\ngot:\n%s", result)
		}
	}
}
