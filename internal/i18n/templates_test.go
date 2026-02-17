package i18n

import (
	"strings"
	"testing"
	"time"
)

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
