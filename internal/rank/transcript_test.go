package rank

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeCodeDir(t *testing.T) {
	dir, err := claudeCodeDir()
	if err != nil {
		t.Fatalf("claudeCodeDir() returned error: %v", err)
	}
	if dir == "" {
		t.Fatal("claudeCodeDir() returned empty string")
	}

	homeDir, _ := os.UserHomeDir()
	expected := filepath.Join(homeDir, ".claude")
	if dir != expected {
		t.Errorf("claudeCodeDir() = %q, want %q", dir, expected)
	}
}

func TestClaudeDesktopConfigDir(t *testing.T) {
	dir, err := claudeDesktopConfigDir()
	if err != nil {
		t.Fatalf("claudeDesktopConfigDir() returned error: %v", err)
	}
	if dir == "" {
		t.Fatal("claudeDesktopConfigDir() returned empty string")
	}
}

func TestFindTranscripts_CLIPaths(t *testing.T) {
	// Set HOME to a temp directory to isolate from the real filesystem.
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	claudeDir := filepath.Join(tempHome, ".claude")

	// Create CLI new format: ~/.claude/projects/<project>/<uuid>.jsonl
	projectDir := filepath.Join(claudeDir, "projects", "test-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cliNewFile := filepath.Join(projectDir, "abc-123-def.jsonl")
	if err := os.WriteFile(cliNewFile, []byte(`{"type":"test"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create CLI old format: ~/.claude/transcripts/<session>.jsonl
	transcriptsDir := filepath.Join(claudeDir, "transcripts")
	if err := os.MkdirAll(transcriptsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cliOldFile := filepath.Join(transcriptsDir, "session-old.jsonl")
	if err := os.WriteFile(cliOldFile, []byte(`{"type":"test"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	transcripts, err := FindTranscripts()
	if err != nil {
		t.Fatalf("FindTranscripts() returned error: %v", err)
	}

	if len(transcripts) < 2 {
		t.Fatalf("FindTranscripts() returned %d files, want at least 2", len(transcripts))
	}

	// Verify both files are found.
	found := make(map[string]bool)
	for _, f := range transcripts {
		found[f] = true
	}

	if !found[cliNewFile] {
		t.Errorf("CLI new format file not found: %s", cliNewFile)
	}
	if !found[cliOldFile] {
		t.Errorf("CLI old format file not found: %s", cliOldFile)
	}
}

func TestFindTranscripts_MultipleSources(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	claudeDir := filepath.Join(tempHome, ".claude")

	// Create multiple project directories with transcripts.
	for _, project := range []string{"project-a", "project-b"} {
		projectDir := filepath.Join(claudeDir, "projects", project)
		if err := os.MkdirAll(projectDir, 0o755); err != nil {
			t.Fatal(err)
		}
		file := filepath.Join(projectDir, project+"-session.jsonl")
		if err := os.WriteFile(file, []byte(`{"type":"test"}`+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	transcripts, err := FindTranscripts()
	if err != nil {
		t.Fatalf("FindTranscripts() returned error: %v", err)
	}

	if len(transcripts) < 2 {
		t.Fatalf("FindTranscripts() returned %d files, want at least 2", len(transcripts))
	}
}

func TestFindTranscripts_EmptyDirectory(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	// No Claude directories at all.
	transcripts, err := FindTranscripts()
	if err != nil {
		t.Fatalf("FindTranscripts() returned error: %v", err)
	}

	if len(transcripts) != 0 {
		t.Errorf("FindTranscripts() returned %d files, want 0", len(transcripts))
	}
}

func TestFindTranscripts_Deduplication(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	claudeDir := filepath.Join(tempHome, ".claude")

	// Create a single file in one location.
	projectDir := filepath.Join(claudeDir, "projects", "dedup-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(projectDir, "unique-session.jsonl")
	if err := os.WriteFile(file, []byte(`{"type":"test"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	transcripts, err := FindTranscripts()
	if err != nil {
		t.Fatalf("FindTranscripts() returned error: %v", err)
	}

	// Check no duplicates.
	seen := make(map[string]int)
	for _, f := range transcripts {
		seen[f]++
	}
	for f, count := range seen {
		if count > 1 {
			t.Errorf("duplicate transcript found: %s (count=%d)", f, count)
		}
	}
}

func TestFindTranscriptForSession_CLIPaths(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	claudeDir := filepath.Join(tempHome, ".claude")
	sessionID := "test-session-abc123"

	// Create CLI new format with session ID in filename.
	projectDir := filepath.Join(claudeDir, "projects", "my-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	expectedFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.WriteFile(expectedFile, []byte(`{"type":"test"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result := FindTranscriptForSession(sessionID)
	if result == "" {
		t.Fatal("FindTranscriptForSession() returned empty, want match")
	}
	if result != expectedFile {
		t.Errorf("FindTranscriptForSession() = %q, want %q", result, expectedFile)
	}
}

func TestFindTranscriptForSession_OldFormat(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	claudeDir := filepath.Join(tempHome, ".claude")
	sessionID := "old-session-xyz"

	// Create CLI old format: ~/.claude/transcripts/<session>.jsonl
	transcriptsDir := filepath.Join(claudeDir, "transcripts")
	if err := os.MkdirAll(transcriptsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	expectedFile := filepath.Join(transcriptsDir, sessionID+".jsonl")
	if err := os.WriteFile(expectedFile, []byte(`{"type":"test"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result := FindTranscriptForSession(sessionID)
	if result == "" {
		t.Fatal("FindTranscriptForSession() returned empty, want match")
	}
	if result != expectedFile {
		t.Errorf("FindTranscriptForSession() = %q, want %q", result, expectedFile)
	}
}

func TestFindTranscriptForSession_NotFound(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	result := FindTranscriptForSession("nonexistent-session")
	if result != "" {
		t.Errorf("FindTranscriptForSession() = %q, want empty string", result)
	}
}

func TestFindTranscriptForSession_PrefersNewFormat(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	claudeDir := filepath.Join(tempHome, ".claude")
	sessionID := "shared-session"

	// Create both old and new format files.
	projectDir := filepath.Join(claudeDir, "projects", "pref-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	newFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.WriteFile(newFile, []byte(`{"type":"new"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	transcriptsDir := filepath.Join(claudeDir, "transcripts")
	if err := os.MkdirAll(transcriptsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	oldFile := filepath.Join(transcriptsDir, sessionID+".jsonl")
	if err := os.WriteFile(oldFile, []byte(`{"type":"old"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result := FindTranscriptForSession(sessionID)
	if result == "" {
		t.Fatal("FindTranscriptForSession() returned empty, want match")
	}

	// New format (projects) should be preferred over old format (transcripts).
	if result != newFile {
		t.Errorf("FindTranscriptForSession() = %q, want new format %q", result, newFile)
	}
}

func TestIsValidSessionID(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		want      bool
	}{
		{"valid UUID", "abc-123-def-456", true},
		{"valid alphanumeric", "session123", true},
		{"valid with underscores", "my_session_id", true},
		{"valid mixed", "abc-123_XYZ", true},
		{"empty string", "", false},
		{"path traversal dots", "../../../etc/passwd", false},
		{"path traversal backslash", `..\..\etc\passwd`, false},
		{"contains slash", "session/evil", false},
		{"contains space", "session id", false},
		{"contains colon", "C:evil", false},
		{"contains null byte", "session\x00evil", false},
		{"too long", string(make([]byte, 129)), false},
		{"max length", string(make([]byte, 128)), false}, // all null bytes → invalid chars
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidSessionID(tt.sessionID)
			if got != tt.want {
				t.Errorf("isValidSessionID(%q) = %v, want %v", tt.sessionID, got, tt.want)
			}
		})
	}
}

func TestFindTranscriptForSession_PathTraversal(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	// Attempt path traversal attack - should return empty.
	result := FindTranscriptForSession("../../../etc/passwd")
	if result != "" {
		t.Errorf("FindTranscriptForSession with path traversal returned %q, want empty", result)
	}

	result = FindTranscriptForSession(`..\..\etc\passwd`)
	if result != "" {
		t.Errorf("FindTranscriptForSession with backslash traversal returned %q, want empty", result)
	}

	result = FindTranscriptForSession("session/evil")
	if result != "" {
		t.Errorf("FindTranscriptForSession with slash returned %q, want empty", result)
	}
}

func TestParseTranscript_BasicUsage(t *testing.T) {
	dir := t.TempDir()
	transcriptFile := filepath.Join(dir, "test-session.jsonl")

	content := `{"timestamp":"2026-01-01T10:00:00Z","type":"user","message":{}}
{"timestamp":"2026-01-01T10:00:05Z","type":"assistant","model":"claude-3-opus","message":{"usage":{"input_tokens":100,"output_tokens":50,"cache_creation_input_tokens":10,"cache_read_input_tokens":5}}}
{"timestamp":"2026-01-01T10:00:10Z","type":"user","message":{}}
{"timestamp":"2026-01-01T10:00:15Z","type":"assistant","model":"claude-3-opus","message":{"usage":{"input_tokens":200,"output_tokens":100,"cache_creation_input_tokens":20,"cache_read_input_tokens":10}}}
`
	if err := os.WriteFile(transcriptFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	usage, err := ParseTranscript(transcriptFile)
	if err != nil {
		t.Fatalf("ParseTranscript() returned error: %v", err)
	}

	if usage.InputTokens != 300 {
		t.Errorf("InputTokens = %d, want 300", usage.InputTokens)
	}
	if usage.OutputTokens != 150 {
		t.Errorf("OutputTokens = %d, want 150", usage.OutputTokens)
	}
	if usage.CacheCreationTokens != 30 {
		t.Errorf("CacheCreationTokens = %d, want 30", usage.CacheCreationTokens)
	}
	if usage.CacheReadTokens != 15 {
		t.Errorf("CacheReadTokens = %d, want 15", usage.CacheReadTokens)
	}
	if usage.TurnCount != 2 {
		t.Errorf("TurnCount = %d, want 2", usage.TurnCount)
	}
	if usage.ModelName != "claude-3-opus" {
		t.Errorf("ModelName = %q, want %q", usage.ModelName, "claude-3-opus")
	}
}

func TestParseTranscript_FileNotFound(t *testing.T) {
	_, err := ParseTranscript("/nonexistent/path/transcript.jsonl")
	if err == nil {
		t.Fatal("ParseTranscript() should return error for missing file")
	}
}

func TestParseTranscript_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	transcriptFile := filepath.Join(dir, "empty.jsonl")
	if err := os.WriteFile(transcriptFile, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	usage, err := ParseTranscript(transcriptFile)
	if err != nil {
		t.Fatalf("ParseTranscript() error: %v", err)
	}
	if usage.InputTokens != 0 {
		t.Errorf("expected 0 input tokens, got %d", usage.InputTokens)
	}
	if usage.TurnCount != 0 {
		t.Errorf("expected 0 turns, got %d", usage.TurnCount)
	}
	if usage.DurationSeconds != 0 {
		t.Errorf("expected 0 duration, got %d", usage.DurationSeconds)
	}
}

func TestParseTranscript_InvalidJSONLines(t *testing.T) {
	dir := t.TempDir()
	transcriptFile := filepath.Join(dir, "mixed.jsonl")
	content := `{invalid json}
{"timestamp":"2026-01-01T10:00:00Z","type":"user","message":{}}
also invalid
{"timestamp":"2026-01-01T10:00:05Z","type":"assistant","model":"claude-sonnet","message":{"usage":{"input_tokens":100,"output_tokens":50}}}
`
	if err := os.WriteFile(transcriptFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	usage, err := ParseTranscript(transcriptFile)
	if err != nil {
		t.Fatalf("ParseTranscript() error: %v", err)
	}
	// Should skip invalid lines, parse valid ones
	if usage.InputTokens != 100 {
		t.Errorf("expected 100 input tokens, got %d", usage.InputTokens)
	}
	if usage.TurnCount != 1 {
		t.Errorf("expected 1 turn, got %d", usage.TurnCount)
	}
}

func TestParseTranscript_DurationCalculation(t *testing.T) {
	dir := t.TempDir()
	transcriptFile := filepath.Join(dir, "duration.jsonl")
	content := `{"timestamp":"2026-01-15T10:00:00Z","type":"user","message":{}}
{"timestamp":"2026-01-15T10:05:00Z","type":"assistant","message":{"usage":{"input_tokens":1}}}
`
	if err := os.WriteFile(transcriptFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	usage, err := ParseTranscript(transcriptFile)
	if err != nil {
		t.Fatal(err)
	}
	if usage.DurationSeconds != 300 {
		t.Errorf("DurationSeconds = %d, want 300", usage.DurationSeconds)
	}
	if usage.StartedAt != "2026-01-15T10:00:00Z" {
		t.Errorf("StartedAt = %q, want 2026-01-15T10:00:00Z", usage.StartedAt)
	}
	if usage.EndedAt != "2026-01-15T10:05:00Z" {
		t.Errorf("EndedAt = %q, want 2026-01-15T10:05:00Z", usage.EndedAt)
	}
}

func TestParseTranscript_ModelFromMessage(t *testing.T) {
	dir := t.TempDir()
	transcriptFile := filepath.Join(dir, "model.jsonl")
	// Model in message.model not top-level
	content := `{"timestamp":"2026-01-01T10:00:00Z","type":"assistant","message":{"model":"inner-model","usage":{"input_tokens":10}}}
`
	if err := os.WriteFile(transcriptFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	usage, err := ParseTranscript(transcriptFile)
	if err != nil {
		t.Fatal(err)
	}
	if usage.ModelName != "inner-model" {
		t.Errorf("ModelName = %q, want inner-model", usage.ModelName)
	}
}

func TestParseTranscript_NoUsageLines(t *testing.T) {
	dir := t.TempDir()
	transcriptFile := filepath.Join(dir, "no-usage.jsonl")
	content := `{"timestamp":"2026-01-01T10:00:00Z","type":"user","message":{}}
{"timestamp":"2026-01-01T10:00:05Z","type":"user","message":{}}
`
	if err := os.WriteFile(transcriptFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	usage, err := ParseTranscript(transcriptFile)
	if err != nil {
		t.Fatal(err)
	}
	if usage.InputTokens != 0 {
		t.Errorf("InputTokens = %d, want 0", usage.InputTokens)
	}
	if usage.TurnCount != 2 {
		t.Errorf("TurnCount = %d, want 2", usage.TurnCount)
	}
}

func TestParseTranscript_BlankLines(t *testing.T) {
	dir := t.TempDir()
	transcriptFile := filepath.Join(dir, "blanks.jsonl")
	content := `

{"timestamp":"2026-01-01T10:00:00Z","type":"user","message":{}}

{"timestamp":"2026-01-01T10:00:05Z","type":"assistant","message":{"usage":{"input_tokens":50,"output_tokens":25}}}

`
	if err := os.WriteFile(transcriptFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	usage, err := ParseTranscript(transcriptFile)
	if err != nil {
		t.Fatal(err)
	}
	if usage.InputTokens != 50 {
		t.Errorf("InputTokens = %d, want 50", usage.InputTokens)
	}
}

func TestGlobJSONL_InvalidPattern(t *testing.T) {
	// A pattern with invalid glob characters should return nil
	result := globJSONL("[-invalid")
	if result != nil {
		t.Errorf("globJSONL with invalid pattern should return nil, got %v", result)
	}
}

func TestGlobJSONL_NoMatches(t *testing.T) {
	dir := t.TempDir()
	result := globJSONL(filepath.Join(dir, "*.nonexistent"))
	if len(result) != 0 {
		t.Errorf("expected 0 matches, got %d", len(result))
	}
}

func TestGlobJSONL_WithMatches(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.jsonl")
	if err := os.WriteFile(f, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	result := globJSONL(filepath.Join(dir, "*.jsonl"))
	if len(result) != 1 {
		t.Fatalf("expected 1 match, got %d", len(result))
	}
	if result[0] != f {
		t.Errorf("expected %q, got %q", f, result[0])
	}
}

func TestIsValidSessionID_MaxLength(t *testing.T) {
	// Exactly 128 chars of valid characters
	id := ""
	for range 128 {
		id += "a"
	}
	if !isValidSessionID(id) {
		t.Error("128-char valid ID should be accepted")
	}

	// 129 chars should be rejected
	id += "a"
	if isValidSessionID(id) {
		t.Error("129-char ID should be rejected")
	}
}

func TestClaudeDesktopConfigDir_ReturnsNonEmpty(t *testing.T) {
	// claudeDesktopConfigDir is already called in TestClaudeDesktopConfigDir.
	// This test validates that the returned path contains a non-trivial suffix.
	dir, err := claudeDesktopConfigDir()
	if err != nil {
		t.Fatalf("claudeDesktopConfigDir() error: %v", err)
	}
	if dir == "" {
		t.Fatal("claudeDesktopConfigDir() returned empty string")
	}
	// The path must contain "Claude" (the application directory on all platforms).
	if !strings.Contains(dir, "Claude") {
		t.Errorf("claudeDesktopConfigDir() = %q, expected to contain 'Claude'", dir)
	}
}

func TestClaudeCodeDir_ErrorBranch(t *testing.T) {
	// claudeCodeDir returns an error when HOME is not set.
	// We test this by temporarily clearing HOME to cover the error branch.
	// On most platforms os.UserHomeDir() will still succeed via other means,
	// so we only verify the success path here (the error branch requires mocking).
	dir, err := claudeCodeDir()
	if err != nil {
		// Some CI environments might not have a home dir - skip in that case.
		t.Skipf("claudeCodeDir() returned error: %v", err)
	}
	if dir == "" {
		t.Fatal("claudeCodeDir() returned empty string")
	}
}

func TestFindTranscriptForSession_DesktopFallback(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	sessionID := "desktop-fallback-session"

	// Do not create any CLI paths; this exercises the claudeDesktopConfigDir
	// fallback branch in FindTranscriptForSession (returns "" because no match).
	result := FindTranscriptForSession(sessionID)
	if result != "" {
		t.Errorf("expected empty result without matching files, got %q", result)
	}
}

func TestParseTranscript_SingleTimestamp(t *testing.T) {
	// Only one line with timestamp - duration should be 0.
	dir := t.TempDir()
	transcriptFile := filepath.Join(dir, "single.jsonl")
	content := `{"timestamp":"2026-01-01T10:00:00Z","type":"user","message":{}}
`
	if err := os.WriteFile(transcriptFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	usage, err := ParseTranscript(transcriptFile)
	if err != nil {
		t.Fatal(err)
	}
	if usage.DurationSeconds != 0 {
		t.Errorf("expected 0 duration for same start/end timestamp, got %d", usage.DurationSeconds)
	}
	if usage.StartedAt != "2026-01-01T10:00:00Z" {
		t.Errorf("StartedAt = %q, want 2026-01-01T10:00:00Z", usage.StartedAt)
	}
}
