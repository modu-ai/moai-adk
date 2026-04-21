package dbsync_test

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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

// TestHandleDBSchemaSync_PathTraversalRejected verifies that crafted paths escaping the
// project root are rejected before matchGlob or parseMigrationStub can observe them.
// Regression guard: without the filepath.Clean + "../" check, "migrations/../../../etc/passwd.sql"
// passes the "migrations/**/*.sql" glob and is fed into os.ReadFile.
func TestHandleDBSchemaSync_PathTraversalRejected(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Only "../"-escape paths are explicitly rejected by this guard. Absolute paths
	// outside the project are handled by matchGlob's prefix check downstream (covered
	// by TestHandleDBSchemaSync_NonMigrationFile and TestMatchesMigrationPattern).
	malicious := []string{
		"migrations/../../../etc/passwd.sql",
		"migrations/../../secrets.sql",
		"../escape.sql",
		"../../..",
	}

	for _, path := range malicious {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			cfg := dbsync.Config{
				FilePath:          path,
				MigrationPatterns: []string{"migrations/**/*.sql"},
				ExcludedPatterns:  dbsync.DefaultExcludedPatterns,
				StateFile:         filepath.Join(tmpDir, "last-seen.json"),
				ProposalFile:      filepath.Join(tmpDir, "proposal.json"),
				ErrorLogFile:      filepath.Join(tmpDir, "db-sync-errors.log"),
				DebounceWindow:    10 * time.Second,
			}

			result := dbsync.HandleDBSchemaSync(cfg)
			if result.ExitCode != 0 {
				t.Errorf("traversal path %q should exit 0, got %d", path, result.ExitCode)
			}
			if result.Decision != dbsync.DecisionSkip {
				t.Errorf("traversal path %q should return skip, got %q", path, result.Decision)
			}
			// Proposal file must NOT be created — would indicate the path was processed.
			if _, err := os.Stat(cfg.ProposalFile); err == nil {
				t.Errorf("traversal path %q produced proposal.json (info disclosure)", path)
			}
		})
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

// ---------------------------------------------------------------------------
// SPEC-DB-SYNC-HARDEN-001: H1/H2/H4 hardening tests
// ---------------------------------------------------------------------------

// TestHandleDBSchemaSync_OversizedFile — AC-1 (handler-level view of the size guard).
// Creates a 2MB synthetic migration file and verifies (a) exit 0, (b) proposal.json is
// written with parsed_content = "", (c) ErrorLogFile contains the expected prefix.
// The strict AC-1 shape assertion (Truncated field) lives in the internal test file.
func TestHandleDBSchemaSync_OversizedFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	migrationDir := filepath.Join(tmpDir, "migrations")
	if err := os.MkdirAll(migrationDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	bigPath := filepath.Join(migrationDir, "huge.sql")
	// Write 2 MiB of 'X' bytes — well above maxMigrationFileSize (1 MiB).
	payload := bytes.Repeat([]byte("X"), 2<<20)
	if err := os.WriteFile(bigPath, payload, 0o644); err != nil {
		t.Fatalf("write oversized: %v", err)
	}

	logFile := filepath.Join(tmpDir, "errors.log")
	proposalFile := filepath.Join(tmpDir, "proposal.json")
	cfg := dbsync.Config{
		FilePath:          bigPath,
		MigrationPatterns: []string{"**/*.sql"},
		ExcludedPatterns:  dbsync.DefaultExcludedPatterns,
		StateFile:         filepath.Join(tmpDir, "last-seen.json"),
		ProposalFile:      proposalFile,
		ErrorLogFile:      logFile,
		DebounceWindow:    10 * time.Second,
	}

	result := dbsync.HandleDBSchemaSync(cfg)
	if result.ExitCode != 0 {
		t.Errorf("oversized should exit 0, got %d", result.ExitCode)
	}

	// proposal.json must exist and have parsed_content == "" per REQ-H1-002.
	data, err := os.ReadFile(proposalFile)
	if err != nil {
		t.Fatalf("read proposal: %v", err)
	}
	var p dbsync.Proposal
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatalf("proposal json: %v", err)
	}
	if p.ParsedContent != "" {
		t.Errorf("ParsedContent = %q, want empty string", p.ParsedContent)
	}

	// ErrorLogFile must contain exactly one size-guard log line.
	logData, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	const want = "parseMigrationStub: file exceeds maxMigrationFileSize="
	got := strings.Count(string(logData), want)
	if got != 1 {
		t.Errorf("log occurrences of %q = %d, want 1; log content:\n%s", want, got, logData)
	}
}

// TestHandleDBSchemaSync_TableDrivenEdges covers the boundary scenarios named by
// REQ-H4-002 / AC-8. Case names are string literals so AC-8's grep can verify
// their presence directly.
func TestHandleDBSchemaSync_TableDrivenEdges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func(t *testing.T, dir string) (filePath string)
		wantExit   int
		wantDecision string
	}{
		{
			name: "empty_file",
			setup: func(t *testing.T, dir string) string {
				p := filepath.Join(dir, "migrations", "empty.sql")
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				if err := os.WriteFile(p, []byte{}, 0o644); err != nil {
					t.Fatalf("write: %v", err)
				}
				return p
			},
			wantExit:     0,
			wantDecision: dbsync.DecisionAskUser,
		},
		{
			name: "utf8_bom",
			setup: func(t *testing.T, dir string) string {
				p := filepath.Join(dir, "migrations", "bom.sql")
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				content := append([]byte{0xEF, 0xBB, 0xBF}, []byte("CREATE TABLE t(id INT);\n")...)
				if err := os.WriteFile(p, content, 0o644); err != nil {
					t.Fatalf("write: %v", err)
				}
				return p
			},
			wantExit:     0,
			wantDecision: dbsync.DecisionAskUser,
		},
		{
			name: "oversized",
			setup: func(t *testing.T, dir string) string {
				p := filepath.Join(dir, "migrations", "oversized.sql")
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				if err := os.WriteFile(p, bytes.Repeat([]byte("Y"), 2<<20), 0o644); err != nil {
					t.Fatalf("write: %v", err)
				}
				return p
			},
			wantExit:     0,
			wantDecision: dbsync.DecisionAskUser,
		},
		{
			name: "nonexistent",
			setup: func(t *testing.T, dir string) string {
				// Ensure the relative-looking path still matches the glob.
				return filepath.Join(dir, "migrations", "ghost.sql")
			},
			wantExit:     0,
			wantDecision: dbsync.DecisionSkip,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			filePath := tt.setup(t, tmpDir)
			cfg := dbsync.Config{
				FilePath:          filePath,
				MigrationPatterns: []string{"**/*.sql"},
				ExcludedPatterns:  dbsync.DefaultExcludedPatterns,
				StateFile:         filepath.Join(tmpDir, "last-seen.json"),
				ProposalFile:      filepath.Join(tmpDir, "proposal.json"),
				ErrorLogFile:      filepath.Join(tmpDir, "errors.log"),
				DebounceWindow:    10 * time.Second,
			}
			result := dbsync.HandleDBSchemaSync(cfg)
			if result.ExitCode != tt.wantExit {
				t.Errorf("%s: ExitCode = %d, want %d", tt.name, result.ExitCode, tt.wantExit)
			}
			if result.Decision != tt.wantDecision {
				t.Errorf("%s: Decision = %q, want %q", tt.name, result.Decision, tt.wantDecision)
			}
		})
	}
}

// TestMatchGlob_TableDrivenEdges covers matchGlob boundary scenarios named by
// REQ-H4-002 / AC-8: `trailing_slash`, `double_star_only`, `unicode_path`.
// Accessed via the exported MatchesMigrationPattern entry point.
func TestMatchGlob_TableDrivenEdges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pattern  string
		filePath string
		want     bool
	}{
		{
			name:     "trailing_slash",
			pattern:  "migrations/",
			filePath: "migrations/001.sql",
			want:     false, // trailing slash with no glob does not match a file path
		},
		{
			name:     "trailing_slash_glob",
			pattern:  "migrations/**/",
			filePath: "migrations/001.sql",
			want:     true, // ** with trailing slash matches any nested file
		},
		{
			name:     "double_star_only",
			pattern:  "**",
			filePath: "prisma/schema.prisma",
			want:     true, // standalone ** matches every path
		},
		{
			name:     "double_star_only_root",
			pattern:  "**",
			filePath: "top.sql",
			want:     true,
		},
		{
			name:     "unicode_path",
			pattern:  "migrations/**/*.sql",
			filePath: "migrations/2026-한글/001_초기.sql",
			want:     true,
		},
		{
			name:     "exact_filename",
			pattern:  "prisma/schema.prisma",
			filePath: "prisma/schema.prisma",
			want:     true,
		},
		{
			name:     "basename_in_same_dir",
			pattern:  "migrations/*.sql",
			filePath: "migrations/001.sql",
			want:     true,
		},
		{
			name:     "extension_fallback",
			pattern:  "db/migrate/**/*.rb",
			filePath: "db/migrate/202601_init.rb",
			want:     true,
		},
		{
			name:     "suffix_not_matching_returns_false",
			pattern:  "migrations/**/*.sql",
			filePath: "migrations/readme.txt",
			want:     false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := dbsync.MatchesMigrationPattern(tt.filePath, []string{tt.pattern})
			if got != tt.want {
				t.Errorf("%s: MatchesMigrationPattern(%q, [%q]) = %v, want %v", tt.name, tt.filePath, tt.pattern, got, tt.want)
			}
		})
	}
}

// TestHandleDBSchemaSync_ReadonlyStateDir exercises the "mkdir state dir" error
// branch in HandleDBSchemaSync. When the state directory cannot be created, the
// handler must exit 0 with DecisionSkip (REQ-011 non-blocking contract).
func TestHandleDBSchemaSync_ReadonlyStateDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows: chmod does not restrict writes per POSIX semantics")
	}
	if os.Getuid() == 0 {
		t.Skip("running as root: chmod does not restrict writes")
	}
	t.Parallel()

	tmpDir := t.TempDir()
	readonly := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(readonly, 0o555); err != nil {
		t.Fatalf("mkdir readonly: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(readonly, 0o755) })

	migrationDir := filepath.Join(tmpDir, "migrations")
	if err := os.MkdirAll(migrationDir, 0o755); err != nil {
		t.Fatalf("mkdir migration: %v", err)
	}
	target := filepath.Join(migrationDir, "001.sql")
	if err := os.WriteFile(target, []byte("--"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	// State dir lives inside the read-only parent → os.MkdirAll fails.
	cfg := dbsync.Config{
		FilePath:          target,
		MigrationPatterns: []string{"**/*.sql"},
		ExcludedPatterns:  dbsync.DefaultExcludedPatterns,
		StateFile:         filepath.Join(readonly, "sub", "last-seen.json"),
		ProposalFile:      filepath.Join(tmpDir, "proposal.json"),
		ErrorLogFile:      filepath.Join(tmpDir, "errors.log"),
		DebounceWindow:    10 * time.Second,
	}

	result := dbsync.HandleDBSchemaSync(cfg)
	if result.ExitCode != 0 {
		t.Errorf("state-dir create failure must exit 0, got %d", result.ExitCode)
	}
	if result.Decision != dbsync.DecisionSkip {
		t.Errorf("state-dir create failure must return skip, got %q", result.Decision)
	}
}

// TestHandleDBSchemaSync_ReadonlyProposalDir exercises the proposal-path write
// failure branch. With an already-debounced stale window, HandleDBSchemaSync
// reaches the proposal write step; pointing ProposalFile at a read-only parent
// forces os.WriteFile to fail and the handler to return skip without crashing.
func TestHandleDBSchemaSync_ReadonlyProposalDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows: chmod does not restrict writes per POSIX semantics")
	}
	if os.Getuid() == 0 {
		t.Skip("running as root: chmod does not restrict writes")
	}
	t.Parallel()

	tmpDir := t.TempDir()
	readonly := filepath.Join(tmpDir, "readonly-proposal")
	if err := os.Mkdir(readonly, 0o555); err != nil {
		t.Fatalf("mkdir readonly: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(readonly, 0o755) })

	migrationDir := filepath.Join(tmpDir, "migrations")
	if err := os.MkdirAll(migrationDir, 0o755); err != nil {
		t.Fatalf("mkdir migration: %v", err)
	}
	target := filepath.Join(migrationDir, "001.sql")
	if err := os.WriteFile(target, []byte("--"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	cfg := dbsync.Config{
		FilePath:          target,
		MigrationPatterns: []string{"**/*.sql"},
		ExcludedPatterns:  dbsync.DefaultExcludedPatterns,
		StateFile:         filepath.Join(tmpDir, "last-seen.json"),
		ProposalFile:      filepath.Join(readonly, "proposal", "proposal.json"),
		ErrorLogFile:      filepath.Join(tmpDir, "errors.log"),
		DebounceWindow:    10 * time.Second,
	}

	result := dbsync.HandleDBSchemaSync(cfg)
	if result.ExitCode != 0 {
		t.Errorf("proposal write failure must exit 0, got %d", result.ExitCode)
	}
	if result.Decision != dbsync.DecisionSkip {
		t.Errorf("proposal write failure must return skip, got %q", result.Decision)
	}
}

// TestHandleDBSchemaSync_WinnerLoserOrdering — AC-3 follow-up. Pre-populates a
// fresh state so the fast path is exercised from inside HandleDBSchemaSync,
// covering the debounced branch end-to-end rather than the standalone helper.
func TestHandleDBSchemaSync_WinnerLoserOrdering(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "last-seen.json")
	migrationDir := filepath.Join(tmpDir, "migrations")
	if err := os.MkdirAll(migrationDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	target := filepath.Join(migrationDir, "001.sql")
	if err := os.WriteFile(target, []byte("--noop"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	cfg := dbsync.Config{
		FilePath:          target,
		MigrationPatterns: []string{"**/*.sql"},
		ExcludedPatterns:  dbsync.DefaultExcludedPatterns,
		StateFile:         stateFile,
		ProposalFile:      filepath.Join(tmpDir, "proposal.json"),
		ErrorLogFile:      filepath.Join(tmpDir, "errors.log"),
		DebounceWindow:    50 * time.Millisecond,
	}

	// First invocation — winner.
	r1 := dbsync.HandleDBSchemaSync(cfg)
	if r1.Decision != dbsync.DecisionAskUser {
		t.Errorf("first call: Decision = %q, want %q", r1.Decision, dbsync.DecisionAskUser)
	}

	// Immediate second invocation — must be debounced.
	r2 := dbsync.HandleDBSchemaSync(cfg)
	if r2.Decision != dbsync.DecisionDebounced {
		t.Errorf("second call: Decision = %q, want %q", r2.Decision, dbsync.DecisionDebounced)
	}
}

// TestCheckDebounce_CorruptStateRecovery — AC-8 `corrupt_state_recovery` and
// REQ-H2-003 alignment. When the state file contains invalid JSON, the debounce
// check must treat it as absent and proceed (not crash, not leave torn state).
func TestCheckDebounce_CorruptStateRecovery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload []byte
	}{
		{name: "corrupt_state_recovery", payload: []byte("{truncated-json")},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			stateFile := filepath.Join(tmpDir, "last-seen.json")
			if err := os.WriteFile(stateFile, tt.payload, 0o644); err != nil {
				t.Fatalf("seed corrupt state: %v", err)
			}

			debounced, err := dbsync.CheckDebounce(stateFile, "migrations/001.sql", 10*time.Second)
			if err != nil {
				t.Fatalf("expected nil error on corrupt state, got %v", err)
			}
			if debounced {
				t.Error("corrupt state should be treated as absent (debounced=false)")
			}

			// After recovery, the state file must be valid JSON.
			after, err := os.ReadFile(stateFile)
			if err != nil {
				t.Fatalf("read recovered state: %v", err)
			}
			var state dbsync.DebounceState
			if err := json.Unmarshal(after, &state); err != nil {
				t.Errorf("recovered state file is not valid JSON: %v (content=%q)", err, after)
			}
		})
	}
}

// TestCheckDebounceConcurrency — AC-3. Two goroutines hit CheckDebounce on the
// same (stateFile, filePath) within a short window. Exactly one must return
// debounced=false and the other debounced=true; the state file must end valid JSON.
// Run with `-count=10 -race` for the full AC-3 assertion.
func TestCheckDebounceConcurrency(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "last-seen.json")
	const filePath = "prisma/schema.prisma"
	window := 50 * time.Millisecond

	var (
		wg      sync.WaitGroup
		results = make([]bool, 2)
		errs    = make([]error, 2)
	)
	wg.Add(2)
	for i := 0; i < 2; i++ {
		i := i
		go func() {
			defer wg.Done()
			debounced, err := dbsync.CheckDebounce(stateFile, filePath, window)
			results[i] = debounced
			errs[i] = err
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: unexpected error %v", i, err)
		}
	}

	// Multiset must be {false, true}.
	if results[0] == results[1] {
		t.Errorf("results = %v, want multiset {false, true}", results)
	}

	// State file must be valid JSON after the race resolves.
	data, err := os.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	var state dbsync.DebounceState
	if err := json.Unmarshal(data, &state); err != nil {
		t.Errorf("state file not valid JSON: %v (content=%q)", err, data)
	}
	if state.FilePath != filePath {
		t.Errorf("state.FilePath = %q, want %q", state.FilePath, filePath)
	}
}

// TestCheckDebounce_NoDirectWriteFile — AC-4. Parses db_schema_sync.go and asserts
// that the CheckDebounce call graph (including its unexported helper) contains
// zero `os.WriteFile(...)` calls and at least one `os.Rename(...)` call. This is
// an AST-level guard that survives variable renames.
func TestCheckDebounce_NoDirectWriteFile(t *testing.T) {
	t.Parallel()

	const srcPath = "db_schema_sync.go"
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, srcPath, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse %s: %v", srcPath, err)
	}

	// Walk any function whose name is CheckDebounce or checkDebounceWithLog.
	target := map[string]bool{"CheckDebounce": true, "checkDebounceWithLog": true}
	var (
		writeFileCalls int
		renameCalls    int
		foundTarget    bool
	)
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		if !target[fn.Name.Name] {
			continue
		}
		foundTarget = true
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			if ident.Name != "os" {
				return true
			}
			switch sel.Sel.Name {
			case "WriteFile":
				writeFileCalls++
			case "Rename":
				renameCalls++
			}
			return true
		})
	}
	if !foundTarget {
		t.Fatalf("neither CheckDebounce nor checkDebounceWithLog found in %s", srcPath)
	}
	if writeFileCalls != 0 {
		t.Errorf("CheckDebounce family contains %d os.WriteFile calls; want 0 (use os.CreateTemp + os.Rename)", writeFileCalls)
	}
	if renameCalls < 1 {
		t.Errorf("CheckDebounce family contains %d os.Rename calls; want >= 1", renameCalls)
	}
}
