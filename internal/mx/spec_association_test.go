package mx

import (
	"testing"
)

// TestExtractSpecIDs_Patterns tests extraction across various SPEC ID patterns.
func TestExtractSpecIDs_Patterns(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected []string
	}{
		{
			name:     "단순 SPEC ID",
			body:     "ANCHOR for SPEC-AUTH-001",
			expected: []string{"SPEC-AUTH-001"},
		},
		{
			name:     "V3R2 형식 SPEC ID",
			body:     "SPEC-V3R2-SPC-004 구현",
			expected: []string{"SPEC-V3R2-SPC-004"},
		},
		{
			name:     "여러 SPEC ID",
			body:     "SPEC-AUTH-001과 SPEC-DB-002",
			expected: []string{"SPEC-AUTH-001", "SPEC-DB-002"},
		},
		{
			name:     "SPEC ID 없음",
			body:     "일반 설명",
			expected: []string{},
		},
		{
			name:     "소문자 SPEC-ID 매칭 안됨",
			body:     "spec-auth-001 참조",
			expected: []string{},
		},
		{
			name:     "중복 제거",
			body:     "SPEC-AUTH-001 두 번: SPEC-AUTH-001",
			expected: []string{"SPEC-AUTH-001"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractSpecIDs(tt.body)

			if len(got) != len(tt.expected) {
				t.Errorf("SPEC ID 수: 기대 %d, 실제 %d (got=%v, expected=%v)",
					len(tt.expected), len(got), got, tt.expected)
				return
			}

			seen := make(map[string]bool)
			for _, id := range got {
				seen[id] = true
			}

			for _, id := range tt.expected {
				if !seen[id] {
					t.Errorf("누락된 SPEC ID: %s (got=%v)", id, got)
				}
			}
		})
	}
}

// TestSpecAssociator_Associate_ByBody tests body-based SPEC association.
func TestSpecAssociator_Associate_ByBody(t *testing.T) {
	associator := NewSpecAssociator(map[string][]string{})

	tag := Tag{
		Kind:     MXAnchor,
		File:     "internal/auth/handler.go",
		Line:     10,
		Body:     "ANCHOR for SPEC-AUTH-001 handler",
		AnchorID: "anchor-auth",
	}

	specs := associator.Associate(tag)

	found := false
	for _, s := range specs {
		if s == "SPEC-AUTH-001" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("본문 기반 연결 실패: SPEC-AUTH-001 없음 (got=%v)", specs)
	}
}

// TestSpecAssociator_Associate_ByModulePath tests file-path-based SPEC association.
func TestSpecAssociator_Associate_ByModulePath(t *testing.T) {
	specModules := map[string][]string{
		"SPEC-AUTH-001": {"internal/auth/"},
		"SPEC-DB-002":   {"internal/db/", "internal/cache/"},
	}

	associator := NewSpecAssociator(specModules)

	tag := Tag{
		Kind: MXNote,
		File: "internal/auth/handler.go",
		Body: "일반 설명 (SPEC ID 없음)",
		Line: 5,
	}

	specs := associator.Associate(tag)

	found := false
	for _, s := range specs {
		if s == "SPEC-AUTH-001" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("경로 기반 연결 실패: SPEC-AUTH-001 없음 (got=%v, file=%s)",
			specs, tag.File)
	}
}

// TestSpecAssociator_Associate_NoDuplicate verifies that there are no duplicate SPEC IDs.
func TestSpecAssociator_Associate_NoDuplicate(t *testing.T) {
	specModules := map[string][]string{
		"SPEC-AUTH-001": {"internal/auth/"},
	}

	associator := NewSpecAssociator(specModules)

	// Tag associated with SPEC-AUTH-001 by both path and body.
	tag := Tag{
		Kind: MXAnchor,
		File: "internal/auth/handler.go",
		Body: "ANCHOR for SPEC-AUTH-001",
		Line: 10,
	}

	specs := associator.Associate(tag)

	// Must contain no duplicates.
	seen := make(map[string]int)
	for _, s := range specs {
		seen[s]++
	}

	for specID, count := range seen {
		if count > 1 {
			t.Errorf("중복 SPEC ID: %s (%d번 등장)", specID, count)
		}
	}
}

// TestIsFileUnderModules verifies whether a file path lives under a module path.
func TestIsFileUnderModules(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		modules  []string
		expected bool
	}{
		{
			name:     "파일이 모듈 하위",
			file:     "internal/auth/handler.go",
			modules:  []string{"internal/auth/"},
			expected: true,
		},
		{
			name:     "파일이 모듈 하위 아님",
			file:     "internal/cache/store.go",
			modules:  []string{"internal/auth/"},
			expected: false,
		},
		{
			name:     "여러 모듈 중 하나",
			file:     "internal/db/query.go",
			modules:  []string{"internal/auth/", "internal/db/"},
			expected: true,
		},
		{
			name:     "모듈 없음",
			file:     "internal/auth/handler.go",
			modules:  []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isFileUnderModules(tt.file, tt.modules)
			if got != tt.expected {
				t.Errorf("isFileUnderModules(%q, %v): 기대 %v, 실제 %v",
					tt.file, tt.modules, tt.expected, got)
			}
		})
	}
}
