package worktree

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppendDivergenceLog_CreatesMarkdownAndJSON(t *testing.T) {
	root := t.TempDir()
	entry := DivergenceLogEntry{
		Timestamp:  time.Date(2026, 5, 8, 12, 30, 0, 0, time.UTC),
		SnapshotID: "snap-test-001",
		AgentName:  "test-agent",
		Divergence: Divergence{
			HeadChanged:    true,
			PreHeadSHA:     "aaaaaaaaaa",
			PostHeadSHA:    "bbbbbbbbbb",
			UntrackedAdded: []string{".moai/specs/SPEC-X/note.md"},
		},
	}
	mdPath, jsonPath, err := AppendDivergenceLog(root, entry)
	if err != nil {
		t.Fatalf("AppendDivergenceLog: %v", err)
	}
	if !strings.Contains(mdPath, "2026-05-08.md") {
		t.Errorf("expected mdPath to contain date, got %s", mdPath)
	}
	mdData, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read md: %v", err)
	}
	mdStr := string(mdData)
	if !strings.Contains(mdStr, "Snapshot ID: snap-test-001") {
		t.Errorf("md missing snapshot id: %s", mdStr)
	}
	if !strings.Contains(mdStr, "test-agent") {
		t.Errorf("md missing agent name: %s", mdStr)
	}
	if !strings.Contains(mdStr, "HeadChanged: true") {
		t.Errorf("md missing HeadChanged: %s", mdStr)
	}
	if !strings.Contains(mdStr, ".moai/specs/SPEC-X/note.md") {
		t.Errorf("md missing UntrackedAdded entry: %s", mdStr)
	}

	// JSON sidecar must round-trip.
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json sidecar: %v", err)
	}
	var loaded DivergenceLogEntry
	if err := json.Unmarshal(jsonData, &loaded); err != nil {
		t.Fatalf("unmarshal json sidecar: %v", err)
	}
	if loaded.SnapshotID != entry.SnapshotID {
		t.Errorf("snapshot id mismatch: got %s want %s", loaded.SnapshotID, entry.SnapshotID)
	}
}

func TestAppendDivergenceLog_AppendsSameDay(t *testing.T) {
	root := t.TempDir()
	t1 := DivergenceLogEntry{
		Timestamp:  time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC),
		SnapshotID: "snap-001",
		Divergence: Divergence{HeadChanged: true},
	}
	t2 := DivergenceLogEntry{
		Timestamp:  time.Date(2026, 5, 8, 14, 0, 0, 0, time.UTC),
		SnapshotID: "snap-002",
		Divergence: Divergence{BranchChanged: true},
	}
	if _, _, err := AppendDivergenceLog(root, t1); err != nil {
		t.Fatalf("first append: %v", err)
	}
	mdPath, _, err := AppendDivergenceLog(root, t2)
	if err != nil {
		t.Fatalf("second append: %v", err)
	}
	data, _ := os.ReadFile(mdPath)
	s := string(data)
	if !strings.Contains(s, "snap-001") || !strings.Contains(s, "snap-002") {
		t.Errorf("expected both snapshot IDs in same-day file, got: %s", s)
	}
}

func TestWriteSuspectFlag(t *testing.T) {
	root := t.TempDir()
	flag := SuspectFlag{
		SnapshotID:  "snap-suspect",
		AgentName:   "agent-x",
		Reason:      SuspectReasonEmptyWorktreePath,
		PushBlocked: true,
	}
	flagPath, err := WriteSuspectFlag(root, flag)
	if err != nil {
		t.Fatalf("WriteSuspectFlag: %v", err)
	}
	if !strings.Contains(flagPath, "worktree-suspect-snap-suspect.flag") {
		t.Errorf("unexpected flag path: %s", flagPath)
	}
	data, err := os.ReadFile(flagPath)
	if err != nil {
		t.Fatalf("read flag: %v", err)
	}
	var loaded SuspectFlag
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshal flag: %v", err)
	}
	if loaded.SnapshotID != "snap-suspect" {
		t.Errorf("snapshot id mismatch: %s", loaded.SnapshotID)
	}
	if loaded.DetectedAt.IsZero() {
		t.Errorf("expected DetectedAt to be auto-populated")
	}
	if loaded.Reason != SuspectReasonEmptyWorktreePath {
		t.Errorf("reason mismatch: %s", loaded.Reason)
	}
}

func TestFilterExcludedPorcelain(t *testing.T) {
	cases := []struct {
		name       string
		lines      []string
		exclusions []string
		want       []string
	}{
		{
			"untracked under .moai/state/ filtered",
			[]string{"?? .moai/state/snapshot.json", " M src/auth.go"},
			[]string{".moai/state/"},
			[]string{" M src/auth.go"},
		},
		{
			"empty exclusions keeps all",
			[]string{"?? a.go", " M b.go"},
			[]string{},
			[]string{"?? a.go", " M b.go"},
		},
		{
			"rename with new path under exclusion",
			[]string{"R  old.go -> .moai/state/new.json"},
			[]string{".moai/state/"},
			[]string{},
		},
		{
			"short line preserved",
			[]string{"?? "},
			[]string{".moai/"},
			[]string{"?? "},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := filterExcludedPorcelain(tc.lines, tc.exclusions)
			if !equalSlices(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestSnapshot_DegradesOnEmptyRepo verifies graceful degradation when no commits exist.
func TestSnapshot_DegradesOnEmptyRepo(t *testing.T) {
	dir := t.TempDir()
	runOrFail(t, dir, "git", "init", "-q", "-b", "main")
	runOrFail(t, dir, "git", "config", "user.email", "test@example.com")
	runOrFail(t, dir, "git", "config", "user.name", "Test")
	// No commit — git rev-parse HEAD will fail.
	snap, err := Capture(t.Context(), CaptureOptions{RepoDir: dir})
	if err != nil {
		t.Fatalf("Capture should not error on empty repo: %v", err)
	}
	if snap.HeadSHA != "" {
		t.Errorf("expected HeadSHA empty on unborn HEAD, got %s", snap.HeadSHA)
	}
}

// dirSep helps avoid the unused-import lint for filepath when other tests may not use it.
var _ = filepath.Separator
