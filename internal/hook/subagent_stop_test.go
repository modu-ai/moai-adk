package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSubagentStopHandler_EventType(t *testing.T) {
	h := NewSubagentStopHandler()
	if h.EventType() != EventSubagentStop {
		t.Errorf("EventType() = %v, want %v", h.EventType(), EventSubagentStop)
	}
}

func TestSubagentStopHandler_Handle_Basic(t *testing.T) {
	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "full fields",
			input: &HookInput{
				SessionID:           "sess-001",
				AgentID:             "agent-abc",
				AgentName:           "expert-backend",
				AgentTranscriptPath: "/tmp/transcript.json",
				TeamName:            "test-team",
				TeammateName:        "teammate-1",
			},
		},
		{
			name:  "empty input",
			input: &HookInput{},
		},
		{
			name: "session id only",
			input: &HookInput{
				SessionID: "sess-002",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewSubagentStopHandler()
			out, err := h.Handle(context.Background(), tt.input)
			if err != nil {
				t.Errorf("Handle() error = %v, want nil", err)
			}
			if out == nil {
				t.Error("Handle() returned nil output")
			}
		})
	}
}

func TestSubagentStopHandler_Handle_TeamConfigCleanup(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("tmux not available on Windows")
	}

	// Create temporary team config
	tempDir := t.TempDir()
	teamConfigPath := filepath.Join(tempDir, "config.json")

	// Write initial team config
	config := teamConfigFile{
		Name: "test-team",
		Members: []teamMemberDb{
			{
				Name:      "teammate-1",
				AgentID:   "agent-1",
				AgentType: "backend",
				TmuxPaneID: "test-session:0.0",
			},
			{
				Name:      "teammate-2",
				AgentID:   "agent-2",
				AgentType: "frontend",
				TmuxPaneID: "test-session:0.1",
			},
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(teamConfigPath, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// Mock home directory to point to temp dir
	// We can't actually change os.UserHomeDir(), so we'll test the helper directly
	h := &subagentStopHandler{}

	// Test findTeammatePaneID
	paneID, err := h.findTeammatePaneID(teamConfigPath, "teammate-1")
	if err != nil {
		t.Errorf("findTeammatePaneID() error = %v", err)
	}
	if paneID != "test-session:0.0" {
		t.Errorf("findTeammatePaneID() = %v, want test-session:0.0", paneID)
	}

	// Test findTeammatePaneID with non-existent teammate
	_, err = h.findTeammatePaneID(teamConfigPath, "teammate-999")
	if err == nil {
		t.Error("findTeammatePaneID() expected error for non-existent teammate")
	}

	// Test removeTeammateFromConfig
	err = h.removeTeammateFromConfig(teamConfigPath, "teammate-1")
	if err != nil {
		t.Errorf("removeTeammateFromConfig() error = %v", err)
	}

	// Verify teammate was removed
	data, err = os.ReadFile(teamConfigPath)
	if err != nil {
		t.Fatalf("read config after removal: %v", err)
	}

	var updatedConfig teamConfigFile
	if err := json.Unmarshal(data, &updatedConfig); err != nil {
		t.Fatalf("parse config after removal: %v", err)
	}

	if len(updatedConfig.Members) != 1 {
		t.Errorf("expected 1 member after removal, got %d", len(updatedConfig.Members))
	}
	if len(updatedConfig.Members) > 0 && updatedConfig.Members[0].Name != "teammate-2" {
		t.Errorf("expected teammate-2, got %s", updatedConfig.Members[0].Name)
	}
}

func TestSubagentStopHandler_Handle_MissingTeammate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("tmux not available on Windows")
	}

	// Create temporary team config without the target teammate
	tempDir := t.TempDir()
	teamConfigPath := filepath.Join(tempDir, "config.json")

	config := teamConfigFile{
		Name: "test-team",
		Members: []teamMemberDb{
			{
				Name:      "other-teammate",
				AgentID:   "agent-2",
				TmuxPaneID: "test-session:0.1",
			},
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(teamConfigPath, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	h := &subagentStopHandler{}

	// Test findTeammatePaneID with non-existent teammate
	_, err = h.findTeammatePaneID(teamConfigPath, "teammate-1")
	if err == nil {
		t.Error("findTeammatePaneID() expected error for non-existent teammate")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got %v", err)
	}
}

func TestSubagentStopHandler_Handle_NoTmuxPaneID(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("tmux not available on Windows")
	}

	// Create temporary team config with teammate lacking tmuxPaneId
	tempDir := t.TempDir()
	teamConfigPath := filepath.Join(tempDir, "config.json")

	config := teamConfigFile{
		Name: "test-team",
		Members: []teamMemberDb{
			{
				Name:      "teammate-1",
				AgentID:   "agent-1",
				// No TmuxPaneID
			},
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(teamConfigPath, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	h := &subagentStopHandler{}

	// Test findTeammatePaneID with teammate lacking tmuxPaneId
	_, err = h.findTeammatePaneID(teamConfigPath, "teammate-1")
	if err == nil {
		t.Error("findTeammatePaneID() expected error for teammate without tmuxPaneId")
	}
	if !strings.Contains(err.Error(), "no tmuxPaneId") {
		t.Errorf("expected 'no tmuxPaneId' error, got %v", err)
	}
}
