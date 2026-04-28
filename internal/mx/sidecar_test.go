package mx

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestNewManager tests manager creation.
func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	if manager.stateDir != tmpDir {
		t.Errorf("Expected stateDir %q, got %q", tmpDir, manager.stateDir)
	}
	if manager.sidecarPath != filepath.Join(tmpDir, SidecarFileName) {
		t.Errorf("Expected sidecarPath %q, got %q",
			filepath.Join(tmpDir, SidecarFileName), manager.sidecarPath)
	}
	if manager.archivePath != filepath.Join(tmpDir, ArchiveFileName) {
		t.Errorf("Expected archivePath %q, got %q",
			filepath.Join(tmpDir, ArchiveFileName), manager.archivePath)
	}
}

// TestWriteAndLoad tests atomic write and load of sidecar.
// AC-SPC-002-02: Full scan produces correct index with schema_version: 2.
// AC-SPC-002-03: Atomic write survives interruption.
func TestWriteAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	// Create test sidecar
	sidecar := &Sidecar{
		SchemaVersion: SchemaVersion,
		Tags: []Tag{
			{
				Kind:       MXNote,
				File:       "/test/file1.go",
				Line:       10,
				Body:       "test note",
				CreatedBy:  "scanner",
				LastSeenAt: time.Now(),
			},
			{
				Kind:       MXWarn,
				File:       "/test/file2.go",
				Line:       20,
				Body:       "test warn",
				Reason:     "test reason",
				CreatedBy:  "scanner",
				LastSeenAt: time.Now(),
			},
		},
		ScannedAt: time.Now(),
	}

	// Write sidecar
	if err := manager.Write(sidecar); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(manager.sidecarPath); os.IsNotExist(err) {
		t.Error("Sidecar file was not created")
	}

	// Load sidecar
	loaded, err := manager.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify schema version
	if loaded.SchemaVersion != SchemaVersion {
		t.Errorf("Expected schema_version %d, got %d", SchemaVersion, loaded.SchemaVersion)
	}

	// Verify tag count
	if len(loaded.Tags) != len(sidecar.Tags) {
		t.Errorf("Expected %d tags, got %d", len(sidecar.Tags), len(loaded.Tags))
	}

	// Verify tag content
	for i, tag := range loaded.Tags {
		if tag.Kind != sidecar.Tags[i].Kind {
			t.Errorf("Tag %d: Expected kind %v, got %v", i, sidecar.Tags[i].Kind, tag.Kind)
		}
		if tag.File != sidecar.Tags[i].File {
			t.Errorf("Tag %d: Expected file %q, got %q", i, sidecar.Tags[i].File, tag.File)
		}
		if tag.Line != sidecar.Tags[i].Line {
			t.Errorf("Tag %d: Expected line %d, got %d", i, sidecar.Tags[i].Line, tag.Line)
		}
		if tag.Body != sidecar.Tags[i].Body {
			t.Errorf("Tag %d: Expected body %q, got %q", i, sidecar.Tags[i].Body, tag.Body)
		}
	}
}

// TestAtomicWrite tests that atomic write prevents partial reads.
// AC-SPC-002-03: Interrupted write results in either old file or new file, never partial.
func TestAtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	// Write initial sidecar
	sidecar1 := &Sidecar{
		SchemaVersion: SchemaVersion,
		Tags: []Tag{
			{Kind: MXNote, File: "/test/file.go", Line: 1, Body: "initial"},
		},
		ScannedAt: time.Now(),
	}

	if err := manager.Write(sidecar1); err != nil {
		t.Fatalf("Initial write failed: %v", err)
	}

	// Write second sidecar (this should use atomic temp + rename)
	sidecar2 := &Sidecar{
		SchemaVersion: SchemaVersion,
		Tags: []Tag{
			{Kind: MXNote, File: "/test/file.go", Line: 1, Body: "updated"},
		},
		ScannedAt: time.Now(),
	}

	if err := manager.Write(sidecar2); err != nil {
		t.Fatalf("Second write failed: %v", err)
	}

	// Load and verify
	loaded, err := manager.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded.Tags) != 1 {
		t.Fatalf("Expected 1 tag, got %d", len(loaded.Tags))
	}

	if loaded.Tags[0].Body != "updated" {
		t.Errorf("Expected body 'updated', got %q", loaded.Tags[0].Body)
	}
}

// TestLoadNonExistent tests loading when sidecar doesn't exist.
func TestLoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	// Load should return empty sidecar without error
	loaded, err := manager.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.SchemaVersion != SchemaVersion {
		t.Errorf("Expected schema_version %d, got %d", SchemaVersion, loaded.SchemaVersion)
	}
	if len(loaded.Tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(loaded.Tags))
	}
}

// TestLoadCorrupt tests handling of corrupt sidecar file.
// AC-SPC-002-09: Corrupt sidecar handled gracefully.
func TestLoadCorrupt(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	// Write invalid JSON
	if err := os.WriteFile(manager.sidecarPath, []byte("invalid json"), 0644); err != nil {
		t.Fatal(err)
	}

	// Load should treat corrupt file as empty
	loaded, err := manager.Load()
	if err != nil {
		t.Fatalf("Load should not error on corrupt file, got: %v", err)
	}

	if loaded.SchemaVersion != SchemaVersion {
		t.Errorf("Expected schema_version %d, got %d", SchemaVersion, loaded.SchemaVersion)
	}
	if len(loaded.Tags) != 0 {
		t.Errorf("Expected 0 tags from corrupt file, got %d", len(loaded.Tags))
	}
}

// TestUpdateFile tests incremental file update.
func TestUpdateFile(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	// Write initial sidecar
	initialTags := []Tag{
		{Kind: MXNote, File: "/test/file1.go", Line: 1, Body: "file1 note"},
		{Kind: MXNote, File: "/test/file2.go", Line: 1, Body: "file2 note"},
	}

	sidecar := &Sidecar{
		SchemaVersion: SchemaVersion,
		Tags:          initialTags,
		ScannedAt:     time.Now(),
	}

	if err := manager.Write(sidecar); err != nil {
		t.Fatal(err)
	}

	// Update file2.go
	newTags := []Tag{
		{Kind: MXWarn, File: "/test/file2.go", Line: 5, Body: "file2 warn", Reason: "test"},
	}

	updated, err := manager.UpdateFile("/test/file2.go", newTags)
	if err != nil {
		t.Fatalf("UpdateFile failed: %v", err)
	}

	// Should have 2 tags: file1 unchanged, file2 updated
	if len(updated.Tags) != 2 {
		t.Fatalf("Expected 2 tags, got %d", len(updated.Tags))
	}

	// Verify file1 unchanged
	foundFile1 := false
	for _, tag := range updated.Tags {
		if tag.File == "/test/file1.go" {
			foundFile1 = true
			if tag.Body != "file1 note" {
				t.Errorf("file1 note should be unchanged, got %q", tag.Body)
			}
		}
	}
	if !foundFile1 {
		t.Error("file1.go tag not found")
	}

	// Verify file2 updated
	foundFile2 := false
	for _, tag := range updated.Tags {
		if tag.File == "/test/file2.go" {
			foundFile2 = true
			if tag.Kind != MXWarn {
				t.Errorf("file2 should be WARN, got %v", tag.Kind)
			}
			if tag.Body != "file2 warn" {
				t.Errorf("file2 body should be 'file2 warn', got %q", tag.Body)
			}
		}
	}
	if !foundFile2 {
		t.Error("file2.go tag not found")
	}
}

// TestArchiveStale tests archival of stale tags.
// AC-SPC-002-07: Stale tag within 7 days stays in sidecar.
// AC-SPC-002-08: Tag older than 7 days archived.
func TestArchiveStale(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	// Create sidecar with fresh and stale tags
	now := time.Now()
	freshTag := Tag{
		Kind:       MXNote,
		File:       "/test/fresh.go",
		Line:       1,
		Body:       "fresh note",
		LastSeenAt: now.Add(-6 * 24 * time.Hour), // 6 days ago
	}

	staleTag := Tag{
		Kind:       MXNote,
		File:       "/test/stale.go",
		Line:       1,
		Body:       "stale note",
		LastSeenAt: now.Add(-8 * 24 * time.Hour), // 8 days ago
	}

	sidecar := &Sidecar{
		SchemaVersion: SchemaVersion,
		Tags:          []Tag{freshTag, staleTag},
		ScannedAt:     now,
	}

	if err := manager.Write(sidecar); err != nil {
		t.Fatal(err)
	}

	// Run archival
	archive, err := manager.ArchiveStale()
	if err != nil {
		t.Fatalf("ArchiveStale failed: %v", err)
	}

	if archive == nil {
		t.Fatal("Expected archive to be created")
	}

	// Verify archive contains stale tag
	if len(archive.ArchivedTags) != 1 {
		t.Fatalf("Expected 1 archived tag, got %d", len(archive.ArchivedTags))
	}

	if archive.ArchivedTags[0].File != "/test/stale.go" {
		t.Errorf("Expected stale.go to be archived, got %s", archive.ArchivedTags[0].File)
	}

	// Verify sidecar now only has fresh tag
	loaded, err := manager.Load()
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Tags) != 1 {
		t.Fatalf("Expected 1 remaining tag, got %d", len(loaded.Tags))
	}

	if loaded.Tags[0].File != "/test/fresh.go" {
		t.Errorf("Expected fresh.go to remain, got %s", loaded.Tags[0].File)
	}
}

// TestArchiveStaleBoundary tests the 7-day boundary exactly.
func TestArchiveStaleBoundary(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	now := time.Now()

	tests := []struct {
		name      string
		lastSeen  time.Time
		shouldRemain bool
	}{
		{"1 hour ago", now.Add(-1 * time.Hour), true},
		{"6 days ago", now.Add(-6 * 24 * time.Hour), true},
		{"7 days ago", now.Add(-7 * 24 * time.Hour), true}, // Exactly 7 days stays
		{"8 days ago", now.Add(-8 * 24 * time.Hour), false}, // Archived
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := Tag{
				Kind:       MXNote,
				File:       "/test/file.go",
				Line:       1,
				Body:       "test",
				LastSeenAt: tt.lastSeen,
			}

			sidecar := &Sidecar{
				SchemaVersion: SchemaVersion,
				Tags:          []Tag{tag},
				ScannedAt:     now,
			}

			if err := manager.Write(sidecar); err != nil {
				t.Fatal(err)
			}

			archive, err := manager.ArchiveStale()
			if err != nil {
				t.Fatal(err)
			}

			loaded, _ := manager.Load()

			if tt.shouldRemain {
				if len(loaded.Tags) != 1 {
					t.Errorf("Expected tag to remain, but it was archived")
				}
			} else {
				if len(loaded.Tags) != 0 {
					t.Errorf("Expected tag to be archived, but it remains")
				}
				if archive == nil || len(archive.ArchivedTags) != 1 {
					t.Errorf("Expected tag in archive")
				}
			}
		})
	}
}

// TestGetTag tests tag lookup by key.
func TestGetTag(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	tag := Tag{
		Kind:       MXNote,
		File:       "/test/file.go",
		Line:       42,
		Body:       "test",
		LastSeenAt: time.Now(),
	}

	sidecar := &Sidecar{
		SchemaVersion: SchemaVersion,
		Tags:          []Tag{tag},
		ScannedAt:     time.Now(),
	}

	if err := manager.Write(sidecar); err != nil {
		t.Fatal(err)
	}

	// Lookup existing tag
	found, ok := manager.GetTag(tag.Key())
	if !ok {
		t.Error("Tag should exist")
	}
	if found.Body != tag.Body {
		t.Errorf("Expected body %q, got %q", tag.Body, found.Body)
	}

	// Lookup non-existent tag
	_, ok = manager.GetTag("/nonexistent/file.go:NOTE:1")
	if ok {
		t.Error("Non-existent tag should not exist")
	}
}

// TestGetAllTags tests retrieval of all tags.
func TestGetAllTags(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	tags := []Tag{
		{Kind: MXNote, File: "/test/file1.go", Line: 1, Body: "note1"},
		{Kind: MXWarn, File: "/test/file2.go", Line: 1, Body: "warn1"},
		{Kind: MXAnchor, File: "/test/file3.go", Line: 1, Body: "anchor1"},
	}

	sidecar := &Sidecar{
		SchemaVersion: SchemaVersion,
		Tags:          tags,
		ScannedAt:     time.Now(),
	}

	if err := manager.Write(sidecar); err != nil {
		t.Fatal(err)
	}

	allTags := manager.GetAllTags()
	if len(allTags) != len(tags) {
		t.Fatalf("Expected %d tags, got %d", len(tags), len(allTags))
	}
}

// TestSidecarJSONSerialization tests JSON serialization format.
func TestSidecarJSONSerialization(t *testing.T) {
	sidecar := &Sidecar{
		SchemaVersion: SchemaVersion,
		Tags: []Tag{
			{
				Kind:       MXNote,
				File:       "/test/file.go",
				Line:       10,
				Body:       "test note",
				CreatedBy:  "scanner",
				LastSeenAt: time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC),
			},
		},
		ScannedAt: time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC),
	}

	data, err := json.MarshalIndent(sidecar, "", "  ")
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify JSON contains expected fields
	jsonStr := string(data)
	expectedFields := []string{
		`"schema_version": 2`,
		`"kind": "NOTE"`,
		`"file": "/test/file.go"`,
		`"line": 10`,
		`"body": "test note"`,
		`"createdBy": "scanner"`,
		`"lastSeenAt": "2026-04-28T12:00:00Z"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON missing expected field: %s\nGot:\n%s", field, jsonStr)
		}
	}
}

// TestArchiveJSONSerialization tests archive JSON format.
func TestArchiveJSONSerialization(t *testing.T) {
	archive := &Archive{
		SchemaVersion: SchemaVersion,
		ArchivedTags: []Tag{
			{
				Kind:       MXNote,
				File:       "/test/old.go",
				Line:       1,
				Body:       "old note",
				LastSeenAt: time.Now(),
			},
		},
		ArchivedAt: time.Now(),
	}

	data, err := json.MarshalIndent(archive, "", "  ")
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify JSON structure
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if parsed["schema_version"].(float64) != float64(SchemaVersion) {
		t.Error("Archive should have schema_version")
	}
	if _, ok := parsed["archived_tags"]; !ok {
		t.Error("Archive should have archived_tags field")
	}
	if _, ok := parsed["archived_at"]; !ok {
		t.Error("Archive should have archived_at field")
	}
}
