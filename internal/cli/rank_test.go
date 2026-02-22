package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/rank"
)

func TestRankCmd_Exists(t *testing.T) {
	if rankCmd == nil {
		t.Fatal("rankCmd should not be nil")
	}
}

func TestRankCmd_Use(t *testing.T) {
	if rankCmd.Use != "rank" {
		t.Errorf("rankCmd.Use = %q, want %q", rankCmd.Use, "rank")
	}
}

func TestRankCmd_Short(t *testing.T) {
	if rankCmd.Short == "" {
		t.Error("rankCmd.Short should not be empty")
	}
}

func TestRankCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "rank" {
			found = true
			break
		}
	}
	if !found {
		t.Error("rank should be registered as a subcommand of root")
	}
}

func TestRankCmd_HasSubcommands(t *testing.T) {
	expected := []string{"login", "status", "logout", "sync", "exclude", "include", "register"}
	for _, name := range expected {
		found := false
		for _, cmd := range rankCmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("rank should have %q subcommand", name)
		}
	}
}

func TestRankCmd_SubcommandCount(t *testing.T) {
	count := len(rankCmd.Commands())
	if count != 7 {
		t.Errorf("rank should have 7 subcommands, got %d", count)
	}
}

func TestRankCmd_SubcommandShortDescriptions(t *testing.T) {
	for _, cmd := range rankCmd.Commands() {
		if cmd.Short == "" {
			t.Errorf("rank subcommand %q should have a short description", cmd.Name())
		}
	}
}

func TestRankCmd_HelpOutput(t *testing.T) {
	// Verify help shows all subcommands
	var names []string
	for _, cmd := range rankCmd.Commands() {
		names = append(names, cmd.Name())
	}

	joined := strings.Join(names, ",")
	if !strings.Contains(joined, "login") || !strings.Contains(joined, "logout") {
		t.Errorf("rank subcommands should include login and logout, got: %s", joined)
	}
}

// TestSubmitSyncBatches_BatchFailure verifies that when SubmitSessionsBatch returns
// a result with Failed > 0, those sessions are NOT marked as synced, while sessions
// from successful batches ARE marked as synced.
func TestSubmitSyncBatches_BatchFailure(t *testing.T) {
	tmpDir := t.TempDir()

	// Create fake transcript files so SyncState.MarkSynced can stat them.
	var paths []string
	for i := range 5 {
		p := filepath.Join(tmpDir, "transcript"+string(rune('A'+i))+".jsonl")
		if err := os.WriteFile(p, []byte(`{}`), 0o644); err != nil {
			t.Fatalf("create transcript file: %v", err)
		}
		paths = append(paths, p)
	}

	// Build matching session submissions
	sessions := make([]*rank.SessionSubmission, len(paths))
	for i := range sessions {
		sessions[i] = &rank.SessionSubmission{
			SessionHash:  "hash" + string(rune('A'+i)),
			InputTokens:  100,
			OutputTokens: 50,
		}
	}

	var buf bytes.Buffer
	ctx := context.Background()

	t.Run("partial_failure_skips_sync_marking", func(t *testing.T) {
		syncState2, _ := rank.NewSyncState(filepath.Join(tmpDir, "ss2.json"))
		buf.Reset()
		failClient := &mockRankClient{
			submitBatchFunc: func(_ context.Context, batch []*rank.SessionSubmission) (*rank.BatchResult, error) {
				return &rank.BatchResult{Success: true, Processed: len(batch), Succeeded: len(batch) - 3, Failed: 3}, nil
			},
		}

		res := submitSyncBatches(ctx, failClient, sessions, paths, syncState2, &buf)

		// All 5 sessions were submitted (no HTTP error)
		if res.Submitted != 5 {
			t.Errorf("Submitted = %d, want 5", res.Submitted)
		}
		if res.FailedTotal != 3 {
			t.Errorf("FailedTotal = %d, want 3", res.FailedTotal)
		}
		if res.ErroredTotal != 0 {
			t.Errorf("ErroredTotal = %d, want 0", res.ErroredTotal)
		}

		// Verify NO sessions are marked as synced (because batch had failures)
		for _, p := range paths {
			if syncState2.IsSynced(p) {
				t.Errorf("session %q should NOT be marked as synced after partial failure", filepath.Base(p))
			}
		}

		// Verify output mentions the failure count
		output := buf.String()
		if !strings.Contains(output, "Failed: 3") {
			t.Errorf("output should mention 3 failed sessions, got %q", output)
		}
	})

	t.Run("success_marks_synced", func(t *testing.T) {
		syncState3, _ := rank.NewSyncState(filepath.Join(tmpDir, "ss3.json"))
		buf.Reset()
		okClient := &mockRankClient{
			submitBatchFunc: func(_ context.Context, batch []*rank.SessionSubmission) (*rank.BatchResult, error) {
				return &rank.BatchResult{Success: true, Processed: len(batch), Succeeded: len(batch), Failed: 0}, nil
			},
		}

		res := submitSyncBatches(ctx, okClient, sessions, paths, syncState3, &buf)

		if res.Submitted != 5 {
			t.Errorf("Submitted = %d, want 5", res.Submitted)
		}
		if res.FailedTotal != 0 {
			t.Errorf("FailedTotal = %d, want 0", res.FailedTotal)
		}

		// Verify ALL sessions are marked as synced
		for _, p := range paths {
			if !syncState3.IsSynced(p) {
				t.Errorf("session %q should be marked as synced after success", filepath.Base(p))
			}
		}
	})

	t.Run("http_error_tracks_errored", func(t *testing.T) {
		syncState4, _ := rank.NewSyncState(filepath.Join(tmpDir, "ss4.json"))
		buf.Reset()
		errClient := &mockRankClient{
			submitBatchFunc: func(_ context.Context, batch []*rank.SessionSubmission) (*rank.BatchResult, error) {
				return nil, errors.New("connection refused")
			},
		}

		res := submitSyncBatches(ctx, errClient, sessions, paths, syncState4, &buf)

		// HTTP error means nothing was submitted
		if res.Submitted != 0 {
			t.Errorf("Submitted = %d, want 0", res.Submitted)
		}
		if res.ErroredTotal != 5 {
			t.Errorf("ErroredTotal = %d, want 5", res.ErroredTotal)
		}

		// Verify NO sessions are marked as synced
		for _, p := range paths {
			if syncState4.IsSynced(p) {
				t.Errorf("session %q should NOT be marked as synced after HTTP error", filepath.Base(p))
			}
		}

		// Verify output mentions the batch failure
		output := buf.String()
		if !strings.Contains(output, "connection refused") {
			t.Errorf("output should mention error, got %q", output)
		}
	})

	t.Run("summary_message_format", func(t *testing.T) {
		// Simulate a scenario where submitSyncBatches returns mixed results
		// and verify the final summary message format.
		buf.Reset()
		mixedClient := &mockRankClient{
			submitBatchFunc: func(_ context.Context, batch []*rank.SessionSubmission) (*rank.BatchResult, error) {
				return &rank.BatchResult{Success: true, Processed: len(batch), Succeeded: 3, Failed: 2}, nil
			},
		}

		res := submitSyncBatches(ctx, mixedClient, sessions, paths, nil, &buf)

		// Verify the counts that would feed into the summary
		succeededTotal := res.Submitted - res.FailedTotal
		if succeededTotal != 3 {
			t.Errorf("succeededTotal = %d, want 3", succeededTotal)
		}
		if res.FailedTotal != 2 {
			t.Errorf("FailedTotal = %d, want 2", res.FailedTotal)
		}
		if res.ErroredTotal != 0 {
			t.Errorf("ErroredTotal = %d, want 0", res.ErroredTotal)
		}
	})
}
