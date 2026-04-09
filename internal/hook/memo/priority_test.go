package memo

import (
	"sort"
	"testing"
)

func TestPriorityOrdering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sections []Section
		wantOrder []Priority
	}{
		{
			name: "sections sorted by priority ascending",
			sections: []Section{
				{Priority: P4Low, Title: "low"},
				{Priority: P2High, Title: "high"},
				{Priority: P1Required, Title: "required"},
				{Priority: P3Medium, Title: "medium"},
			},
			wantOrder: []Priority{P1Required, P2High, P3Medium, P4Low},
		},
		{
			name: "same priority sorted by title",
			sections: []Section{
				{Priority: P2High, Title: "zebra"},
				{Priority: P2High, Title: "apple"},
			},
			wantOrder: []Priority{P2High, P2High},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sorted := make([]Section, len(tt.sections))
			copy(sorted, tt.sections)
			sort.Slice(sorted, func(i, j int) bool {
				if sorted[i].Priority != sorted[j].Priority {
					return sorted[i].Priority < sorted[j].Priority
				}
				return sorted[i].Title < sorted[j].Title
			})

			for i, want := range tt.wantOrder {
				if sorted[i].Priority != want {
					t.Errorf("index %d: got priority %d, want %d", i, sorted[i].Priority, want)
				}
			}
		})
	}
}

func TestPriorityLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		priority Priority
		want     string
	}{
		{P1Required, "P1"},
		{P2High, "P2"},
		{P3Medium, "P3"},
		{P4Low, "P4"},
		{Priority(99), "P?"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := priorityLabel(tt.priority)
			if got != tt.want {
				t.Errorf("priorityLabel(%d) = %q, want %q", tt.priority, got, tt.want)
			}
		})
	}
}

func TestPriorityConstants(t *testing.T) {
	t.Parallel()

	if P1Required >= P2High {
		t.Errorf("P1Required (%d) must be less than P2High (%d)", P1Required, P2High)
	}
	if P2High >= P3Medium {
		t.Errorf("P2High (%d) must be less than P3Medium (%d)", P2High, P3Medium)
	}
	if P3Medium >= P4Low {
		t.Errorf("P3Medium (%d) must be less than P4Low (%d)", P3Medium, P4Low)
	}
}
