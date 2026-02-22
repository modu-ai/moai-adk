package statusline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestNewTaskCollector_DefaultPath(t *testing.T) {
	tc := newTaskCollector(time.Second)
	if tc == nil {
		t.Fatal("newTaskCollector returned nil")
	}
	if tc.ttl != time.Second {
		t.Errorf("ttl = %v, want %v", tc.ttl, time.Second)
	}
	// statePath should contain ".moai/memory/last-session-state.json"
	if tc.statePath != "" {
		if filepath.Base(tc.statePath) != "last-session-state.json" {
			t.Errorf("statePath base = %q, want last-session-state.json", filepath.Base(tc.statePath))
		}
	}
}

func TestTaskCollector_Read_EmptyPath(t *testing.T) {
	tc := &taskCollector{
		ttl:       time.Second,
		statePath: "",
	}
	data := tc.read()
	if data == nil {
		t.Fatal("read() returned nil, want empty TaskData")
	}
	if data.Active {
		t.Error("expected Active = false for empty statePath")
	}
}

func TestTaskCollector_Read_NonExistentFile(t *testing.T) {
	tc := &taskCollector{
		ttl:       time.Second,
		statePath: "/nonexistent/path/that/does/not/exist.json",
	}
	data := tc.read()
	if data == nil {
		t.Fatal("read() returned nil, want empty TaskData")
	}
	if data.Active {
		t.Error("expected Active = false for non-existent file")
	}
}

func TestTaskCollector_Read_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "last-session-state.json")

	state := sessionState{
		LastUpdated: "2026-01-01",
		ActiveTask: &activeTask{
			Command: "run",
			SpecID:  "SPEC-001",
			Stage:   "analyze",
		},
	}
	b, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if err := os.WriteFile(statePath, b, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tc := &taskCollector{
		ttl:       time.Second,
		statePath: statePath,
	}
	data := tc.read()
	if data == nil {
		t.Fatal("read() returned nil")
	}
	if !data.Active {
		t.Error("expected Active = true")
	}
	if data.Command != "run" {
		t.Errorf("Command = %q, want %q", data.Command, "run")
	}
	if data.SpecID != "SPEC-001" {
		t.Errorf("SpecID = %q, want %q", data.SpecID, "SPEC-001")
	}
	if data.Stage != "analyze" {
		t.Errorf("Stage = %q, want %q", data.Stage, "analyze")
	}
}

func TestTaskCollector_Read_NoActiveTask(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "last-session-state.json")

	state := sessionState{
		LastUpdated: "2026-01-01",
		// No ActiveTask
	}
	b, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if err := os.WriteFile(statePath, b, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tc := &taskCollector{
		ttl:       time.Second,
		statePath: statePath,
	}
	data := tc.read()
	if data == nil {
		t.Fatal("read() returned nil")
	}
	if data.Active {
		t.Error("expected Active = false when no active task")
	}
}

func TestTaskCollector_Read_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "last-session-state.json")

	if err := os.WriteFile(statePath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tc := &taskCollector{
		ttl:       time.Second,
		statePath: statePath,
	}
	data := tc.read()
	if data == nil {
		t.Fatal("read() returned nil")
	}
	if data.Active {
		t.Error("expected Active = false for invalid JSON")
	}
}

func TestTaskCollector_Get_CachesResult(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "last-session-state.json")

	state := sessionState{
		ActiveTask: &activeTask{
			Command: "plan",
			SpecID:  "SPEC-002",
			Stage:   "design",
		},
	}
	b, _ := json.Marshal(state)
	_ = os.WriteFile(statePath, b, 0644)

	tc := &taskCollector{
		ttl:       10 * time.Second,
		statePath: statePath,
	}

	// First call reads from file
	data1 := tc.get()
	if data1 == nil || !data1.Active {
		t.Fatal("expected active task data on first get()")
	}

	// Modify file - cached result should be returned
	emptyState := sessionState{}
	b, _ = json.Marshal(emptyState)
	_ = os.WriteFile(statePath, b, 0644)

	data2 := tc.get()
	if data2 == nil {
		t.Fatal("get() returned nil")
	}
	// Cache should return previous result (still "plan")
	if data2.Command != "plan" {
		t.Errorf("expected cached command 'plan', got %q", data2.Command)
	}
}

func TestTaskCollector_Get_RefreshesAfterTTL(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "last-session-state.json")

	state := sessionState{
		ActiveTask: &activeTask{
			Command: "sync",
			SpecID:  "SPEC-003",
		},
	}
	b, _ := json.Marshal(state)
	_ = os.WriteFile(statePath, b, 0644)

	tc := &taskCollector{
		ttl:       1 * time.Millisecond, // very short TTL
		statePath: statePath,
	}

	// First call
	data1 := tc.get()
	if data1 == nil || !data1.Active {
		t.Fatal("expected active task data on first get()")
	}

	// Wait for TTL to expire
	time.Sleep(10 * time.Millisecond)

	// Modify file
	emptyState := sessionState{}
	b, _ = json.Marshal(emptyState)
	_ = os.WriteFile(statePath, b, 0644)

	// After TTL, should reload from file
	data2 := tc.get()
	if data2 == nil {
		t.Fatal("get() returned nil after TTL")
	}
	if data2.Active {
		t.Error("expected inactive task after TTL expiry and file update")
	}
}

func TestTaskCollector_Get_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "last-session-state.json")

	state := sessionState{
		ActiveTask: &activeTask{
			Command: "run",
			SpecID:  "SPEC-004",
		},
	}
	b, _ := json.Marshal(state)
	_ = os.WriteFile(statePath, b, 0644)

	tc := &taskCollector{
		ttl:       time.Second,
		statePath: statePath,
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data := tc.get()
			if data == nil {
				t.Error("concurrent get() returned nil")
			}
		}()
	}
	wg.Wait()
}

func TestTaskData_Format_Active(t *testing.T) {
	tests := []struct {
		name string
		task TaskData
		want string
	}{
		{
			name: "command only",
			task: TaskData{Active: true, Command: "plan"},
			want: "[plan]",
		},
		{
			name: "command and spec",
			task: TaskData{Active: true, Command: "run", SpecID: "SPEC-001"},
			want: "[run SPEC-001]",
		},
		{
			name: "command spec and stage",
			task: TaskData{Active: true, Command: "run", SpecID: "SPEC-001", Stage: "analyze"},
			want: "[run SPEC-001-analyze]",
		},
		{
			name: "command and stage no spec",
			task: TaskData{Active: true, Command: "sync", Stage: "improve"},
			want: "[sync-improve]",
		},
		{
			name: "not active returns empty",
			task: TaskData{Active: false, Command: "run", SpecID: "SPEC-001"},
			want: "",
		},
		{
			name: "active but empty command returns empty",
			task: TaskData{Active: true, Command: ""},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.task.Format()
			if got != tt.want {
				t.Errorf("Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCollectTask_ReturnsData(t *testing.T) {
	// CollectTask uses the global collector; just verify it doesn't panic and returns non-nil
	data := CollectTask()
	if data == nil {
		t.Error("CollectTask() returned nil")
	}
}
