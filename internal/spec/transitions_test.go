package spec

import (
	"testing"
)

func TestClassifyPRTitle(t *testing.T) {
	tests := []struct {
		name         string
		title        string
		wantCategory string
		wantStatus   string
		wantErr      bool
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
			// SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 M2 (mechanism ②): legacy bare
			// docs(sync)/sync prefix는 4-phase model에서 sync-phase = implemented로
			// 분류된다 (completed가 아님 — completed는 close-infix 전용).
			name:         "sync merge - docs(sync) bare prefix → implemented (4-phase)",
			title:        "docs(sync): SPEC-FOO-001 status update",
			wantCategory: "sync-merge",
			wantStatus:   "implemented",
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
		"draft":       false,
		"planned":     false,
		"in-progress": false,
		"implemented": false,
		"completed":   false,
		"superseded":  false,
		"archived":    false,
		"rejected":    false,
	}

	// Check which values are reachable via ClassifyPRTitle
	// SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 M2: completed는 이제 close-infix로만 도달한다
	// (docs(sync) bare prefix → implemented). completed coverage를 close-infix title로 이동.
	testTitles := []string{
		"status(draft): SPEC-001",                                     // draft
		"plan(spec): SPEC-001 — draft",                                // planned
		"chore(SPEC-001): partial work",                               // in-progress
		"feat(SPEC-001): implement",                                   // implemented
		"chore(SPEC-001): Mx-phase audit-ready signal + 4-phase close", // completed (close-infix)
		"status(superseded): SPEC-001 replaced by SPEC-002",           // superseded
		"status(archived): SPEC-001 obsolete",                         // archived
		"status(rejected): SPEC-001 won't fix",                        // rejected
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

// TestClassifyPRTitle_ChoreSpecUnchanged is an AC-LSCSK-003 regression guard.
// chore(spec): classification must return the skip-meta category + empty status (by design).
// This test fails immediately if the chore(spec) classification rule in transitions.go changes.
//
// Note: chore(specs): (plural) has no dedicated rule in transitions.go, so it
// falls through to the generic chore rule ("run-partial", "in-progress").
// This is the intended ClassifyPRTitle behavior; shouldSkipCommitTitle handles the skip separately.
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
			// chore(specs): (plural) has no dedicated rule in transitions.go and
			// falls through to the generic chore rule -> ("run-partial", "in-progress").
			// The walker handles the skip in shouldSkipCommitTitle, so this behavior is normal.
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

// TestClassifyPRTitle_StaleSyncRuleCorrected는 AC-DLC-002 검증 (REQ-DLC-003, REQ-DLC-004).
// SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 M2 (mechanism ②): legacy bare sync/docs(sync)
// prefix는 4-phase model에서 sync-phase = implemented로 분류되어야 한다 (completed 아님).
// close-infix는 여전히 completed의 유일한 신호이며 ClassifyPRTitle에서 먼저 검사되어
// sync rule보다 우선한다.
func TestClassifyPRTitle_StaleSyncRuleCorrected(t *testing.T) {
	tests := []struct {
		name         string
		title        string
		wantCategory string
		wantStatus   string
	}{
		{
			// 핵심: 정규 sync commit subject (implemented 표기 포함)는 implemented로 분류
			name:         "sync(SPEC-X): lifecycle complete — v0.3.0 implemented → implemented",
			title:        "sync(SPEC-EXAMPLE-001): lifecycle complete — v0.3.0 implemented",
			wantCategory: "sync-merge",
			wantStatus:   "implemented",
		},
		{
			// legacy bare sync prefix → implemented (이전: completed)
			name:         "bare sync prefix → implemented (4-phase 정정)",
			title:        "sync(SPEC-EXAMPLE-001): status transition",
			wantCategory: "sync-merge",
			wantStatus:   "implemented",
		},
		{
			// legacy bare docs(sync) prefix → implemented (이전: completed)
			name:         "bare docs(sync) prefix → implemented (4-phase 정정)",
			title:        "docs(sync): legacy bare prefix",
			wantCategory: "sync-merge",
			wantStatus:   "implemented",
		},
		{
			// close-infix는 sync rule보다 먼저 검사되어 이긴다 (D3 edge case)
			name:         "sync(SPEC-X): ... 4-phase close → completed (close-infix wins)",
			title:        "sync(SPEC-EXAMPLE-001): lifecycle 4-phase close",
			wantCategory: "mx-close",
			wantStatus:   "completed",
		},
		{
			// anti-regression: feat는 여전히 implemented
			name:         "feat → implemented (변화 없음)",
			title:        "feat(SPEC-EXAMPLE-001): M1 implementation",
			wantCategory: "run-complete",
			wantStatus:   "implemented",
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

// TestClassifyPRTitle_CloseInfix는 AC-DCA-002 검증 (REQ-DCA-001, REQ-DCA-003).
// 정규 close convention commit (`chore(SPEC-{ID}): ... 4-phase close` 또는
// `Mx-phase audit-ready` infix)은 generic `chore` prefix 규칙보다 먼저
// `completed`로 분류되어야 한다.
func TestClassifyPRTitle_CloseInfix(t *testing.T) {
	tests := []struct {
		name         string
		title        string
		wantCategory string
		wantStatus   string
	}{
		{
			name:         "정규 4-phase close commit은 completed로 분류 (generic chore 아님)",
			title:        "chore(SPEC-EXAMPLE-001): Mx-phase audit-ready signal + 4-phase close",
			wantCategory: "mx-close",
			wantStatus:   "completed",
		},
		{
			// SPEC-V3R6-LIFECYCLE-REDESIGN-001 REQ-LR-020 (D4, AC-LR-012): 새 canonical
			// "3-phase close" infix도 completed로 분류되어야 한다 (legacy "4-phase close"와 함께).
			name:         "정규 3-phase close commit은 completed로 분류 (REQ-LR-020 새 canonical infix)",
			title:        "chore(SPEC-EXAMPLE-001): sync-phase audit-ready signal + 3-phase close",
			wantCategory: "mx-close",
			wantStatus:   "completed",
		},
		{
			// REQ-LR-020 (D4): "3-phase close" infix 단독도 completed (대소문자 무관).
			name:         "3-phase close infix 단독도 completed (대소문자 무관, REQ-LR-020)",
			title:        "CHORE(SPEC-EXAMPLE-001): 3-Phase Close",
			wantCategory: "mx-close",
			wantStatus:   "completed",
		},
		{
			name:         "Mx-phase audit-ready infix 단독도 completed",
			title:        "chore(SPEC-EXAMPLE-001): Mx-phase audit-ready signal",
			wantCategory: "mx-close",
			wantStatus:   "completed",
		},
		{
			name:         "4-phase close infix 단독도 completed (대소문자 무관)",
			title:        "CHORE(SPEC-EXAMPLE-001): 4-Phase Close",
			wantCategory: "mx-close",
			wantStatus:   "completed",
		},
		{
			// SHA-backfill chore는 close-infix가 없으므로 close 규칙에 걸리지 않고
			// generic chore (run-partial, in-progress)로 분류된다. walker는 별도로
			// shouldSkipCommitTitle에서 이 backfill chore를 skip한다 (AC-DCA-003).
			name:         "SHA-backfill chore는 close-infix가 없어 generic chore로 분류 (walker가 별도 skip)",
			title:        "chore(SPEC-EXAMPLE-001): backfill §E.2/§E.5 commit SHA",
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
