package statusline

import (
	"strings"
	"testing"
)

// backlogData is a session line whose backlog segment has something to say —
// the suppression assertions below must not be able to pass because the data
// was empty.
func backlogData() *StatusData {
	return &StatusData{
		SessionName: "lane-1",
		Backlog:     BacklogCounts{Picked: 2, Queued: 7, Available: true},
	}
}

func boolPtr(v bool) *bool { return &v }

// TestRendererBacklogSegmentGating is AC-T-003 (REQ-2 surface 2).
//
// Two suppression paths reach the same segment and they must not overwrite
// each other: the pre-existing statusline.yaml `backlog: false` and the new
// workflow.todo.enabled flag. Either one off means the segment is off; both
// absent means it is on. The third case is what pins that the new flag did not
// take ownership of a decision the old one already owned.
func TestRendererBacklogSegmentGating(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		segments    map[string]bool
		todoEnabled *bool
		wantSegment bool
	}{
		{
			name:        "todo disabled, backlog segment key absent",
			segments:    nil,
			todoEnabled: boolPtr(false),
			wantSegment: false,
		},
		{
			name:        "control: todo key absent, backlog segment key absent",
			segments:    nil,
			todoEnabled: nil,
			wantSegment: true,
		},
		{
			name:        "control: todo explicitly true, backlog segment key absent",
			segments:    nil,
			todoEnabled: boolPtr(true),
			wantSegment: true,
		},
		{
			name:        "the pre-existing statusline.yaml path still suppresses",
			segments:    map[string]bool{SegmentBacklog: false},
			todoEnabled: boolPtr(true),
			wantSegment: false,
		},
		{
			name:        "both off",
			segments:    map[string]bool{SegmentBacklog: false},
			todoEnabled: boolPtr(false),
			wantSegment: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRenderer("default", true, tc.segments)
			r.SetTodoEnabled(tc.todoEnabled)

			got := r.renderSessionLine(backlogData())
			has := strings.Contains(got, "TODO:")
			if has != tc.wantSegment {
				t.Fatalf("TODO segment present = %v, want %v (line: %q)", has, tc.wantSegment, got)
			}
			// Identity is never gated by the todo flag — only the backlog
			// segment is. A guard that swallowed the whole line would pass the
			// assertion above for the wrong reason.
			if !strings.Contains(got, "lane-1") {
				t.Fatalf("session identity lost from the line: %q", got)
			}
		})
	}
}
