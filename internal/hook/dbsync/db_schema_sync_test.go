package dbsync_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook/dbsync"
)

// TestMatchesMigrationPattern verifies that only migration-pattern paths match.
func TestMatchesMigrationPattern(t *testing.T) {
	t.Parallel()

	patterns := []string{
		"prisma/schema.prisma",
		"alembic/versions/**/*.py",
		"db/migrate/**/*.rb",
		"migrations/**/*.sql",
		"supabase/migrations/**/*.sql",
		"sql/migrations/**/*.sql",
	}

	tests := []struct {
		name     string
		filePath string
		want     bool
	}{
		{"prisma schema", "prisma/schema.prisma", true},
		{"alembic migration", "alembic/versions/001_initial.py", true},
		{"rails migration", "db/migrate/20240101_create_users.rb", true},
		{"sql migration", "migrations/001_init.sql", true},
		{"supabase migration", "supabase/migrations/20240101_init.sql", true},
		{"sql migrations subdir", "sql/migrations/v1/001.sql", true},
		{"random go file", "internal/server/server.go", false},
		{"schema.md (excluded doc)", ".moai/project/db/schema.md", false},
		{"package.json", "package.json", false},
		{"README.md", "README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := dbsync.MatchesMigrationPattern(tt.filePath, patterns)
			if got != tt.want {
				t.Errorf("MatchesMigrationPattern(%q) = %v, want %v", tt.filePath, got, tt.want)
			}
		})
	}
}

// TestIsExcluded verifies that excluded patterns (recursion guard) are detected.
func TestIsExcluded(t *testing.T) {
	t.Parallel()

	excluded := []string{
		".moai/project/db/**",
		".moai/cache/**",
		".moai/logs/**",
	}

	tests := []struct {
		name     string
		filePath string
		want     bool
	}{
		{"schema.md excluded", ".moai/project/db/schema.md", true},
		{"erd.mmd excluded", ".moai/project/db/erd.mmd", true},
		{"cache file excluded", ".moai/cache/db-sync/proposal.json", true},
		{"log file excluded", ".moai/logs/db-sync-errors.log", true},
		{"migration sql NOT excluded", "migrations/001.sql", false},
		{"prisma NOT excluded", "prisma/schema.prisma", false},
		{"source file NOT excluded", "internal/server/server.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := dbsync.IsExcluded(tt.filePath, excluded)
			if got != tt.want {
				t.Errorf("IsExcluded(%q) = %v, want %v", tt.filePath, got, tt.want)
			}
		})
	}
}

// TestDebounce_FirstCallNotDebounced verifies the first invocation always proceeds.
func TestDebounce_FirstCallNotDebounced(t *testing.T) {
	t.Parallel()

	stateDir := t.TempDir()
	stateFile := filepath.Join(stateDir, "last-seen.json")

	debounced, err := dbsync.CheckDebounce(stateFile, "migrations/001.sql", 10*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if debounced {
		t.Error("first call should NOT be debounced")
	}
}

// TestDebounce_SameFileWithinWindow triggers debounce.
func TestDebounce_SameFileWithinWindow(t *testing.T) {
	t.Parallel()

	stateDir := t.TempDir()
	stateFile := filepath.Join(stateDir, "last-seen.json")

	// First call — writes state
	_, err := dbsync.CheckDebounce(stateFile, "migrations/001.sql", 10*time.Second)
	if err != nil {
		t.Fatalf("setup first call: %v", err)
	}

	// Second call immediately — same file within window
	debounced, err := dbsync.CheckDebounce(stateFile, "migrations/001.sql", 10*time.Second)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if !debounced {
		t.Error("second call for same file within 10s window should be debounced")
	}
}

// TestDebounce_DifferentFileNotDebounced verifies different files bypass debounce.
func TestDebounce_DifferentFileNotDebounced(t *testing.T) {
	t.Parallel()

	stateDir := t.TempDir()
	stateFile := filepath.Join(stateDir, "last-seen.json")

	// First call with file A
	_, err := dbsync.CheckDebounce(stateFile, "migrations/001.sql", 10*time.Second)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Second call with file B — different file, should not be debounced
	debounced, err := dbsync.CheckDebounce(stateFile, "migrations/002.sql", 10*time.Second)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if debounced {
		t.Error("different file should NOT be debounced")
	}
}

// TestDebounce_ExpiredWindowNotDebounced verifies expired window allows re-processing.
func TestDebounce_ExpiredWindowNotDebounced(t *testing.T) {
	t.Parallel()

	stateDir := t.TempDir()
	stateFile := filepath.Join(stateDir, "last-seen.json")

	// Write state with a timestamp 15 seconds in the past
	state := dbsync.DebounceState{
		FilePath:  "migrations/001.sql",
		Timestamp: time.Now().Add(-15 * time.Second),
	}
	data, _ := json.Marshal(state)
	if err := os.WriteFile(stateFile, data, 0o644); err != nil {
		t.Fatalf("write state: %v", err)
	}

	// Should NOT be debounced because window expired
	debounced, err := dbsync.CheckDebounce(stateFile, "migrations/001.sql", 10*time.Second)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if debounced {
		t.Error("call after debounce window expired should NOT be debounced")
	}
}

// TestBuildProposal_ValidJSON verifies the proposal JSON structure.
func TestBuildProposal_ValidJSON(t *testing.T) {
	t.Parallel()

	proposal := dbsync.BuildProposal("prisma/schema.prisma", "stub-parsed-content")

	if proposal.FilePath != "prisma/schema.prisma" {
		t.Errorf("FilePath = %q, want %q", proposal.FilePath, "prisma/schema.prisma")
	}
	if proposal.ParsedContent != "stub-parsed-content" {
		t.Errorf("ParsedContent = %q, want %q", proposal.ParsedContent, "stub-parsed-content")
	}
	if proposal.Decision != "ask-user" {
		t.Errorf("Decision = %q, want %q", proposal.Decision, "ask-user")
	}

	// Must be serializable to JSON (REQ-009)
	_, err := json.Marshal(proposal)
	if err != nil {
		t.Errorf("proposal must be JSON serializable: %v", err)
	}
}

// TestHandleDBSchemaSync_EmptyPath exits 0 silently.
func TestHandleDBSchemaSync_EmptyPath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cfg := dbsync.Config{
		FilePath:          "",
		MigrationPatterns: []string{"migrations/**/*.sql"},
		ExcludedPatterns:  dbsync.DefaultExcludedPatterns,
		StateFile:         filepath.Join(tmpDir, "last-seen.json"),
		ProposalFile:      filepath.Join(tmpDir, "proposal.json"),
		ErrorLogFile:      filepath.Join(tmpDir, "db-sync-errors.log"),
		DebounceWindow:    10 * time.Second,
	}

	result := dbsync.HandleDBSchemaSync(cfg)
	if result.ExitCode != 0 {
		t.Errorf("empty path should exit 0, got %d", result.ExitCode)
	}
	if result.Decision != dbsync.DecisionSkip {
		t.Errorf("empty path should return skip decision, got %q", result.Decision)
	}
}

// TestHandleDBSchemaSync_ExcludedPattern exits 0 (recursion guard, AC-3).
func TestHandleDBSchemaSync_ExcludedPattern(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cfg := dbsync.Config{
		FilePath:          ".moai/project/db/schema.md",
		MigrationPatterns: []string{"migrations/**/*.sql"},
		ExcludedPatterns:  dbsync.DefaultExcludedPatterns,
		StateFile:         filepath.Join(tmpDir, "last-seen.json"),
		ProposalFile:      filepath.Join(tmpDir, "proposal.json"),
		ErrorLogFile:      filepath.Join(tmpDir, "db-sync-errors.log"),
		DebounceWindow:    10 * time.Second,
	}

	result := dbsync.HandleDBSchemaSync(cfg)
	if result.ExitCode != 0 {
		t.Errorf("excluded path should exit 0, got %d", result.ExitCode)
	}
}

// TestHandleDBSchemaSync_NonMigrationFile exits 0 (no match, REQ-003).
func TestHandleDBSchemaSync_NonMigrationFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cfg := dbsync.Config{
		FilePath:          "internal/server/server.go",
		MigrationPatterns: []string{"migrations/**/*.sql"},
		ExcludedPatterns:  dbsync.DefaultExcludedPatterns,
		StateFile:         filepath.Join(tmpDir, "last-seen.json"),
		ProposalFile:      filepath.Join(tmpDir, "proposal.json"),
		ErrorLogFile:      filepath.Join(tmpDir, "db-sync-errors.log"),
		DebounceWindow:    10 * time.Second,
	}

	result := dbsync.HandleDBSchemaSync(cfg)
	if result.ExitCode != 0 {
		t.Errorf("non-migration file should exit 0, got %d", result.ExitCode)
	}
}

// TestHandleDBSchemaSync_Debounced exits 0 on debounce (AC-10).
func TestHandleDBSchemaSync_Debounced(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "last-seen.json")

	// Pre-write a fresh debounce state for the file
	state := dbsync.DebounceState{
		FilePath:  "migrations/001.sql",
		Timestamp: time.Now(),
	}
	data, _ := json.Marshal(state)
	_ = os.WriteFile(stateFile, data, 0o644)

	cfg := dbsync.Config{
		FilePath:          "migrations/001.sql",
		MigrationPatterns: []string{"migrations/**/*.sql"},
		ExcludedPatterns:  dbsync.DefaultExcludedPatterns,
		StateFile:         stateFile,
		ProposalFile:      filepath.Join(tmpDir, "proposal.json"),
		ErrorLogFile:      filepath.Join(tmpDir, "db-sync-errors.log"),
		DebounceWindow:    10 * time.Second,
	}

	result := dbsync.HandleDBSchemaSync(cfg)
	if result.ExitCode != 0 {
		t.Errorf("debounced call should exit 0, got %d", result.ExitCode)
	}
	if result.Decision != dbsync.DecisionDebounced {
		t.Errorf("debounced call should return debounced decision, got %q", result.Decision)
	}
}

// TestHandleDBSchemaSync_WritesProposal verifies proposal.json is written (REQ-009).
func TestHandleDBSchemaSync_WritesProposal(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Create a real migration file for parsing
	migrationFile := filepath.Join(tmpDir, "migrations", "001_init.sql")
	if err := os.MkdirAll(filepath.Dir(migrationFile), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(migrationFile, []byte("CREATE TABLE users (id INTEGER PRIMARY KEY);"), 0o644); err != nil {
		t.Fatalf("write migration: %v", err)
	}

	proposalFile := filepath.Join(tmpDir, "proposal.json")
	cfg := dbsync.Config{
		FilePath:          migrationFile,
		MigrationPatterns: []string{"migrations/**/*.sql", "**/*.sql"},
		ExcludedPatterns:  dbsync.DefaultExcludedPatterns,
		StateFile:         filepath.Join(tmpDir, "last-seen.json"),
		ProposalFile:      proposalFile,
		ErrorLogFile:      filepath.Join(tmpDir, "db-sync-errors.log"),
		DebounceWindow:    10 * time.Second,
	}

	result := dbsync.HandleDBSchemaSync(cfg)
	if result.ExitCode != 0 {
		t.Errorf("expected exit 0 for migration file, got %d", result.ExitCode)
	}

	// proposal.json must exist
	if _, err := os.Stat(proposalFile); os.IsNotExist(err) {
		t.Error("proposal.json should be written after successful parse")
	}

	// Validate JSON structure
	data, err := os.ReadFile(proposalFile)
	if err != nil {
		t.Fatalf("read proposal: %v", err)
	}
	var proposal dbsync.Proposal
	if err := json.Unmarshal(data, &proposal); err != nil {
		t.Fatalf("proposal.json must be valid JSON: %v", err)
	}
	if proposal.Decision != "ask-user" {
		t.Errorf("proposal.Decision = %q, want %q", proposal.Decision, "ask-user")
	}
}

// TestHandleDBSchemaSync_NonBlockingOnParseError verifies AC-3: exit 0 on error.
func TestHandleDBSchemaSync_NonBlockingOnParseError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Point to a non-existent migration file that will cause parse failure
	cfg := dbsync.Config{
		FilePath:          "migrations/nonexistent.sql",
		MigrationPatterns: []string{"migrations/**/*.sql", "**/*.sql"},
		ExcludedPatterns:  dbsync.DefaultExcludedPatterns,
		StateFile:         filepath.Join(tmpDir, "last-seen.json"),
		ProposalFile:      filepath.Join(tmpDir, "proposal.json"),
		ErrorLogFile:      filepath.Join(tmpDir, "db-sync-errors.log"),
		DebounceWindow:    10 * time.Second,
	}

	result := dbsync.HandleDBSchemaSync(cfg)
	// Must exit 0 (non-blocking) even on error (REQ-011, AC-3)
	if result.ExitCode != 0 {
		t.Errorf("parse error should be non-blocking (exit 0), got %d", result.ExitCode)
	}
}
