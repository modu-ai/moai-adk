package status

import (
	"testing"
	"time"
)

// --- FormatTimeAgo ---

func TestFormatTimeAgo(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "just now (0 seconds)",
			duration: 0,
			want:     "just now",
		},
		{
			name:     "just now (30 seconds)",
			duration: 30 * time.Second,
			want:     "just now",
		},
		{
			name:     "1 minute",
			duration: 1 * time.Minute,
			want:     "1m ago",
		},
		{
			name:     "5 minutes",
			duration: 5 * time.Minute,
			want:     "5m ago",
		},
		{
			name:     "59 minutes",
			duration: 59 * time.Minute,
			want:     "59m ago",
		},
		{
			name:     "1 hour",
			duration: 1 * time.Hour,
			want:     "1h ago",
		},
		{
			name:     "23 hours",
			duration: 23 * time.Hour,
			want:     "23h ago",
		},
		{
			name:     "1 day",
			duration: 24 * time.Hour,
			want:     "1d ago",
		},
		{
			name:     "7 days",
			duration: 7 * 24 * time.Hour,
			want:     "7d ago",
		},
		{
			name:     "30 days",
			duration: 30 * 24 * time.Hour,
			want:     "30d ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a time in the past
			pastTime := time.Now().Add(-tt.duration)
			got := FormatTimeAgo(pastTime)
			if got != tt.want {
				t.Errorf("FormatTimeAgo() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- GitInfo struct ---

func TestGitInfoStruct(t *testing.T) {
	info := &GitInfo{
		Branch:     "main",
		IsClean:    true,
		HasChanges: false,
	}

	if info.Branch != "main" {
		t.Errorf("Branch = %q, want %q", info.Branch, "main")
	}
	if !info.IsClean {
		t.Error("IsClean should be true")
	}
	if info.HasChanges {
		t.Error("HasChanges should be false")
	}
}

func TestGitInfoStruct_DirtyState(t *testing.T) {
	info := &GitInfo{
		Branch:     "feature/test",
		IsClean:    false,
		HasChanges: true,
	}

	if info.Branch != "feature/test" {
		t.Errorf("Branch = %q, want %q", info.Branch, "feature/test")
	}
	if info.IsClean {
		t.Error("IsClean should be false")
	}
	if !info.HasChanges {
		t.Error("HasChanges should be true")
	}
}

// --- CommitInfo struct ---

func TestCommitInfoStruct(t *testing.T) {
	commit := CommitInfo{
		Hash:    "abc1234",
		Message: "feat: add new feature",
		TimeAgo: "2h ago",
	}

	if commit.Hash != "abc1234" {
		t.Errorf("Hash = %q, want %q", commit.Hash, "abc1234")
	}
	if commit.Message != "feat: add new feature" {
		t.Errorf("Message = %q", commit.Message)
	}
	if commit.TimeAgo != "2h ago" {
		t.Errorf("TimeAgo = %q, want %q", commit.TimeAgo, "2h ago")
	}
}
