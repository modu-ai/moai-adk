package spec

import (
	"testing"
)

func TestClassifyPRTitle(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		wantCategory  string
		wantStatus    string
		wantErr       bool
	}{
		{
			name:         "plan merge - standard format",
			title:        "plan(spec): SPEC-FOO-001 — initial draft",
			wantCategory: "plan-merge",
			wantStatus:   "planned",
			wantErr:      false,
		},
		{
			name:         "run complete - feat prefix",
			title:        "feat(SPEC-FOO-001): implement REQ-1",
			wantCategory: "run-complete",
			wantStatus:   "implemented",
			wantErr:      false,
		},
		{
			name:         "sync merge - docs prefix",
			title:        "docs(sync): SPEC-FOO-001 status=completed",
			wantCategory: "sync-merge",
			wantStatus:   "completed",
			wantErr:      false,
		},
		{
			name:         "skip meta - auto-sync",
			title:        "chore(spec): auto-sync status for #999",
			wantCategory: "skip-meta",
			wantStatus:   "",
			wantErr:      false,
		},
		{
			name:         "no-op - revert",
			title:        "revert: feat(SPEC-FOO-001): something",
			wantCategory: "no-op",
			wantStatus:   "",
			wantErr:      false,
		},
		{
			name:         "empty title",
			title:        "",
			wantCategory: "",
			wantStatus:   "",
			wantErr:      true,
		},
		{
			name:         "unknown prefix",
			title:        "unknown: some message",
			wantCategory: "unknown",
			wantStatus:   "",
			wantErr:      false,
		},
		{
			name:         "mixed case prefix",
			title:        "FEAT(SPEC-FOO-001): implement REQ-1",
			wantCategory: "run-complete",
			wantStatus:   "implemented",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category, status, err := ClassifyPRTitle(tt.title)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClassifyPRTitle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if category != tt.wantCategory {
				t.Errorf("ClassifyPRTitle() category = %v, want %v", category, tt.wantCategory)
			}
			if status != tt.wantStatus {
				t.Errorf("ClassifyPRTitle() status = %v, want %v", status, tt.wantStatus)
			}
		})
	}
}

func TestPrefixToStatusCompleteness(t *testing.T) {
	// Verify all canonical enum values are covered
	canonicalValues := map[string]bool{
		"draft":        false,
		"planned":      false,
		"in-progress":  false,
		"implemented":  false,
		"completed":    false,
		"superseded":   false,
		"archived":     false,
		"rejected":     false,
	}

	// Check which values are reachable via ClassifyPRTitle
	testTitles := []string{
		"status(draft): SPEC-001",                        // draft
		"plan(spec): SPEC-001 — draft",                   // planned
		"chore(SPEC-001): partial work",                  // in-progress
		"feat(SPEC-001): implement",                      // implemented
		"docs(sync): SPEC-001 status=completed",          // completed
		"status(superseded): SPEC-001 replaced by SPEC-002", // superseded
		"status(archived): SPEC-001 obsolete",             // archived
		"status(rejected): SPEC-001 won't fix",           // rejected
	}

	for _, title := range testTitles {
		_, status, err := ClassifyPRTitle(title)
		if err == nil && status != "" {
			if _, exists := canonicalValues[status]; exists {
				canonicalValues[status] = true
			}
		}
	}

	allCovered := true
	for value, covered := range canonicalValues {
		if !covered {
			t.Errorf("Canonical status value %q is not covered by ClassifyPRTitle", value)
			allCovered = false
		}
	}

	if !allCovered {
		t.Error("Not all canonical enum values are reachable through transitions")
	}
}

// TestClassifyPRTitle_ChoreSpecUnchanged는 AC-LSCSK-003 regression guard다.
// chore(spec): 분류는 skip-meta 카테고리 + 빈 status를 반환해야 한다 (의도된 설계).
// transitions.go 의 chore(spec) 분류 규칙이 변경되면 이 테스트가 즉시 실패한다.
//
// 주의: chore(specs): (plural) 은 transitions.go에 별도 규칙이 없으므로
// generic chore 규칙 ("run-partial", "in-progress") 으로 분류된다.
// 이는 ClassifyPRTitle의 의도된 동작이며, shouldSkipCommitTitle에서 별도로 처리한다.
func TestClassifyPRTitle_ChoreSpecUnchanged(t *testing.T) {
	tests := []struct {
		name         string
		title        string
		wantCategory string
		wantStatus   string
	}{
		{
			name:         "chore(spec) sweep commit은 skip-meta + 빈 status를 반환해야 함",
			title:        "chore(spec): status drift sweep",
			wantCategory: "skip-meta",
			wantStatus:   "",
		},
		{
			name:         "chore(spec) lint-skip 등록 commit도 동일",
			title:        "chore(spec): status drift 11건 sweep + lint-skip 등록 (lint clean)",
			wantCategory: "skip-meta",
			wantStatus:   "",
		},
		{
			// chore(specs): plural 은 transitions.go에 별도 규칙이 없으므로
			// generic chore 규칙으로 fallthrough → ("run-partial", "in-progress")
			// walker는 shouldSkipCommitTitle에서 별도 skip 처리하므로 이 동작은 정상
			name:         "chore(specs) plural은 generic chore 규칙으로 분류됨 (walker에서 shouldSkipCommitTitle이 처리)",
			title:        "chore(specs): bulk metadata update",
			wantCategory: "run-partial",
			wantStatus:   "in-progress",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category, status, err := ClassifyPRTitle(tt.title)
			if err != nil {
				t.Fatalf("예상치 못한 오류: %v", err)
			}
			if category != tt.wantCategory {
				t.Errorf("category = %q, want %q", category, tt.wantCategory)
			}
			if status != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
		})
	}
}
