package mx

import (
	"testing"
	"time"
)

func TestTagKind(t *testing.T) {
	tests := []struct {
		name string
		kind TagKind
	}{
		{"NOTE kind", MXNote},
		{"WARN kind", MXWarn},
		{"ANCHOR kind", MXAnchor},
		{"TODO kind", MXTodo},
		{"LEGACY kind", MXLegacy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.kind == "" {
				t.Error("TagKind should not be empty")
			}
		})
	}
}

func TestTagKey(t *testing.T) {
	tag := Tag{
		Kind: MXNote,
		File: "/path/to/file.go",
		Line: 42,
	}

	expected := "/path/to/file.go:NOTE:42"
	if got := tag.Key(); got != expected {
		t.Errorf("Tag.Key() = %v, want %v", got, expected)
	}
}

func TestTagIsStale(t *testing.T) {
	tests := []struct {
		name     string
		lastSeen time.Time
		stale    bool
	}{
		{
			name:     "fresh tag",
			lastSeen: time.Now().Add(-1 * time.Hour),
			stale:    false,
		},
		{
			name:     "6 days old",
			lastSeen: time.Now().Add(-6 * 24 * time.Hour),
			stale:    false,
		},
		{
			name:     "7 days old",
			lastSeen: time.Now().Add(-7 * 24 * time.Hour),
			stale:    false, // Exactly 7 days is not stale
		},
		{
			name:     "8 days old",
			lastSeen: time.Now().Add(-8 * 24 * time.Hour),
			stale:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := Tag{LastSeenAt: tt.lastSeen}
			if got := tag.IsStale(); got != tt.stale {
				t.Errorf("Tag.IsStale() = %v, want %v", got, tt.stale)
			}
		})
	}
}

func TestTagFields(t *testing.T) {
	tag := Tag{
		Kind:       MXWarn,
		File:       "/test/file.go",
		Line:       10,
		Body:       "dangerous goroutine without Done()",
		Reason:     "Must call Done() to avoid goroutine leak",
		AnchorID:   "test-anchor",
		CreatedBy:  "scanner",
		LastSeenAt: time.Now(),
	}

	if tag.Kind != MXWarn {
		t.Errorf("Expected Kind MXWarn, got %v", tag.Kind)
	}
	if tag.File != "/test/file.go" {
		t.Errorf("Expected File /test/file.go, got %v", tag.File)
	}
	if tag.Line != 10 {
		t.Errorf("Expected Line 10, got %v", tag.Line)
	}
	if tag.Reason == "" {
		t.Error("WARN tag should have Reason")
	}
}

func TestTagKindString(t *testing.T) {
	tests := []struct {
		kind TagKind
		want string
	}{
		{MXNote, "NOTE"},
		{MXWarn, "WARN"},
		{MXAnchor, "ANCHOR"},
		{MXTodo, "TODO"},
		{MXLegacy, "LEGACY"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := string(tt.kind); got != tt.want {
				t.Errorf("TagKind string = %v, want %v", got, tt.want)
			}
		})
	}
}
