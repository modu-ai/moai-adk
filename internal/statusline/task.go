package statusline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TaskData holds active task information from the session state.
type TaskData struct {
	Command string // e.g., "plan", "run", "sync"
	SpecID  string // e.g., "SPEC-001"
	Stage   string // e.g., "analyze", "preserve", "improve"
	Active  bool
}

// sessionState represents the last session state file format.
type sessionState struct {
	LastUpdated        string      `json:"last_updated"`
	CurrentBranch      string      `json:"current_branch"`
	UncommittedChanges string      `json:"uncommitted_changes"`
	UncommittedFiles   int         `json:"uncommitted_files"`
	SpecsInProgress    []string    `json:"specs_in_progress"`
	ActiveTask         *activeTask `json:"active_task,omitempty"`
}

type activeTask struct {
	Command string `json:"command"`
	SpecID  string `json:"spec_id"`
	Stage   string `json:"stage"`
}

// taskCollector caches and retrieves active task information.
type taskCollector struct {
	mu        sync.RWMutex
	cache     *TaskData
	cacheTime time.Time
	ttl       time.Duration
	statePath string
}

// newTaskCollector creates a task collector with the given TTL.
func newTaskCollector(ttl time.Duration) *taskCollector {
	homeDir, err := os.UserHomeDir()
	// If home directory cannot be determined, use empty path
	// (the collector will simply return empty task data)
	if err != nil {
		return &taskCollector{
			ttl:       ttl,
			statePath: "",
		}
	}
	return &taskCollector{
		ttl:       ttl,
		statePath: filepath.Join(homeDir, ".moai", "memory", "last-session-state.json"),
	}
}

// get retrieves the current active task data.
func (t *taskCollector) get() *TaskData {
	// Always use write lock for simplicity - this is a fast path anyway
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cache != nil && time.Since(t.cacheTime) < t.ttl {
		return t.cache
	}

	data := t.read()
	t.cache = data
	t.cacheTime = time.Now()

	return data
}

func (t *taskCollector) read() *TaskData {
	data := &TaskData{}

	bytes, err := os.ReadFile(t.statePath)
	if err != nil {
		return data
	}

	var state sessionState
	if err := json.Unmarshal(bytes, &state); err != nil {
		return data
	}

	if state.ActiveTask != nil {
		data.Command = state.ActiveTask.Command
		data.SpecID = state.ActiveTask.SpecID
		data.Stage = state.ActiveTask.Stage
		data.Active = true
	}

	return data
}

// Format returns the formatted task string.
func (t *TaskData) Format() string {
	if !t.Active || t.Command == "" {
		return ""
	}

	result := "[" + t.Command
	if t.SpecID != "" {
		result += " " + t.SpecID
	}
	if t.Stage != "" {
		result += "-" + t.Stage
	}
	result += "]"

	return result
}

// Global task collector instance.
var globalTaskCollector = newTaskCollector(time.Second)

// CollectTask retrieves active task information from the session state.
func CollectTask() *TaskData {
	return globalTaskCollector.get()
}
