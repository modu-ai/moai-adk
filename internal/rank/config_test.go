package rank

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNewConfig_Defaults(t *testing.T) {
	// Ensure env var is not set for this test.
	t.Setenv("MOAI_RANK_API_URL", "")

	cfg := NewConfig()
	if cfg.BaseURL != DefaultBaseURL {
		t.Errorf("expected base URL %q, got %q", DefaultBaseURL, cfg.BaseURL)
	}
}

func TestNewConfig_EnvOverride(t *testing.T) {
	t.Setenv("MOAI_RANK_API_URL", "https://custom.example.com")

	cfg := NewConfig()
	if cfg.BaseURL != "https://custom.example.com" {
		t.Errorf("expected env override URL, got %q", cfg.BaseURL)
	}
}

func TestFileCredentialStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	store := NewFileCredentialStore(dir)

	creds := &Credentials{
		APIKey:    "test-api-key-123",
		Username:  "testuser",
		UserID:    "uid-456",
		CreatedAt: "2026-01-15T10:00:00Z",
	}

	// Save credentials.
	if err := store.Save(creds); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load credentials.
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("Load returned nil")
	}

	if loaded.APIKey != creds.APIKey {
		t.Errorf("APIKey: expected %q, got %q", creds.APIKey, loaded.APIKey)
	}
	if loaded.Username != creds.Username {
		t.Errorf("Username: expected %q, got %q", creds.Username, loaded.Username)
	}
	if loaded.UserID != creds.UserID {
		t.Errorf("UserID: expected %q, got %q", creds.UserID, loaded.UserID)
	}
	if loaded.CreatedAt != creds.CreatedAt {
		t.Errorf("CreatedAt: expected %q, got %q", creds.CreatedAt, loaded.CreatedAt)
	}
}

func TestFileCredentialStore_Load_NoFile(t *testing.T) {
	dir := t.TempDir()
	store := NewFileCredentialStore(dir)

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load should not return error for missing file: %v", err)
	}
	if loaded != nil {
		t.Error("Load should return nil for missing file")
	}
}

func TestFileCredentialStore_Load_CorruptedJSON(t *testing.T) {
	dir := t.TempDir()
	store := NewFileCredentialStore(dir)

	// Write corrupted JSON.
	credPath := filepath.Join(dir, "credentials.json")
	if err := os.WriteFile(credPath, []byte("{invalid json}}}"), 0o600); err != nil {
		t.Fatal(err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load should not return error for corrupted JSON: %v", err)
	}
	if loaded != nil {
		t.Error("Load should return nil for corrupted JSON")
	}
}

func TestFileCredentialStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store := NewFileCredentialStore(dir)

	creds := &Credentials{APIKey: "key-to-delete"}
	if err := store.Save(creds); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists.
	if !store.HasCredentials() {
		t.Fatal("expected credentials to exist after save")
	}

	// Delete.
	if err := store.Delete(); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify file no longer exists.
	if store.HasCredentials() {
		t.Error("expected credentials to not exist after delete")
	}
}

func TestFileCredentialStore_Delete_NoFile(t *testing.T) {
	dir := t.TempDir()
	store := NewFileCredentialStore(dir)

	// Delete a non-existent file should not error.
	if err := store.Delete(); err != nil {
		t.Fatalf("Delete should not fail for non-existent file: %v", err)
	}
}

func TestFileCredentialStore_HasCredentials(t *testing.T) {
	dir := t.TempDir()
	store := NewFileCredentialStore(dir)

	if store.HasCredentials() {
		t.Error("HasCredentials should return false when no file exists")
	}

	creds := &Credentials{APIKey: "test-key"}
	if err := store.Save(creds); err != nil {
		t.Fatal(err)
	}

	if !store.HasCredentials() {
		t.Error("HasCredentials should return true after save")
	}
}

func TestFileCredentialStore_GetAPIKey(t *testing.T) {
	dir := t.TempDir()
	store := NewFileCredentialStore(dir)

	// No credentials: empty string, no error.
	key, err := store.GetAPIKey()
	if err != nil {
		t.Fatalf("GetAPIKey should not error when no creds: %v", err)
	}
	if key != "" {
		t.Errorf("expected empty key, got %q", key)
	}

	// Save credentials and retrieve key.
	creds := &Credentials{APIKey: "my-secret-key"}
	if err := store.Save(creds); err != nil {
		t.Fatal(err)
	}

	key, err = store.GetAPIKey()
	if err != nil {
		t.Fatalf("GetAPIKey failed: %v", err)
	}
	if key != "my-secret-key" {
		t.Errorf("expected %q, got %q", "my-secret-key", key)
	}
}

func TestFileCredentialStore_FilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix file permissions not supported on Windows")
	}

	dir := t.TempDir()
	store := NewFileCredentialStore(dir)

	creds := &Credentials{APIKey: "perm-test-key"}
	if err := store.Save(creds); err != nil {
		t.Fatal(err)
	}

	// Check file permissions (0600).
	// Note: Windows doesn't support Unix file permissions, so skip this check
	if runtime.GOOS != "windows" {
		credPath := filepath.Join(dir, "credentials.json")
		info, err := os.Stat(credPath)
		if err != nil {
			t.Fatal(err)
		}

		perm := info.Mode().Perm()
		if perm != 0o600 {
			t.Errorf("expected file permissions 0600, got %04o", perm)
		}
	}
}

func TestFileCredentialStore_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	store := NewFileCredentialStore(dir)

	// Write initial credentials.
	creds1 := &Credentials{APIKey: "key-v1", Username: "user1"}
	if err := store.Save(creds1); err != nil {
		t.Fatal(err)
	}

	// Overwrite with new credentials.
	creds2 := &Credentials{APIKey: "key-v2", Username: "user2"}
	if err := store.Save(creds2); err != nil {
		t.Fatal(err)
	}

	// Verify new credentials are loaded.
	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.APIKey != "key-v2" {
		t.Errorf("expected key-v2, got %q", loaded.APIKey)
	}
	if loaded.Username != "user2" {
		t.Errorf("expected user2, got %q", loaded.Username)
	}

	// Verify no temp file remains.
	tmpPath := filepath.Join(dir, "credentials.json.tmp")
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("temp file should not exist after successful save")
	}
}

func TestFileCredentialStore_JSONFormat(t *testing.T) {
	dir := t.TempDir()
	store := NewFileCredentialStore(dir)

	creds := &Credentials{
		APIKey:    "json-key",
		Username:  "jsonuser",
		UserID:    "json-uid",
		CreatedAt: "2026-02-01T12:00:00Z",
	}
	if err := store.Save(creds); err != nil {
		t.Fatal(err)
	}

	// Read raw file and verify it is valid indented JSON.
	credPath := filepath.Join(dir, "credentials.json")
	data, err := os.ReadFile(credPath)
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]string
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("saved file is not valid JSON: %v", err)
	}

	if parsed["api_key"] != "json-key" {
		t.Errorf("expected api_key json-key, got %q", parsed["api_key"])
	}
}

func TestNewFileCredentialStore_DefaultDir(t *testing.T) {
	store := NewFileCredentialStore("")
	// Should not panic and should have a non-empty path.
	if store.credPath == "" {
		t.Error("credPath should not be empty with default dir")
	}
	if store.dir == "" {
		t.Error("dir should not be empty with default dir")
	}
}

func TestNewFileCredentialStore_CustomDir(t *testing.T) {
	store := NewFileCredentialStore("/tmp/custom-rank")
	if store.dir != "/tmp/custom-rank" {
		t.Errorf("expected dir /tmp/custom-rank, got %q", store.dir)
	}
	expected := filepath.Join("/tmp/custom-rank", "credentials.json")
	if store.credPath != expected {
		t.Errorf("expected credPath %q, got %q", expected, store.credPath)
	}
}

func TestFileCredentialStore_Save_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Use a deep nested path that doesn't exist yet
	nestedDir := filepath.Join(tmpDir, "deep", "nested", "rank")
	store := NewFileCredentialStore(nestedDir)

	creds := &Credentials{APIKey: "nested-key"}
	if err := store.Save(creds); err != nil {
		t.Fatalf("Save should create directories: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load after Save: %v", err)
	}
	if loaded == nil || loaded.APIKey != "nested-key" {
		t.Errorf("expected nested-key, got %v", loaded)
	}
}

func TestFileCredentialStore_Save_ReadOnlyDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("read-only directory test not reliable on Windows")
	}

	tmpDir := t.TempDir()
	store := NewFileCredentialStore(tmpDir)

	// Save initial credentials
	creds := &Credentials{APIKey: "init-key"}
	if err := store.Save(creds); err != nil {
		t.Fatal(err)
	}

	// Make directory read-only
	if err := os.Chmod(tmpDir, 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(tmpDir, 0o755) })

	err := store.Save(&Credentials{APIKey: "new-key"})
	if err == nil {
		t.Skip("expected error writing to read-only dir, but succeeded (possibly running as root)")
	}
}

func TestFileCredentialStore_DeviceID_Serialization(t *testing.T) {
	dir := t.TempDir()
	store := NewFileCredentialStore(dir)

	creds := &Credentials{
		APIKey:   "key",
		DeviceID: "abc123def456",
	}
	if err := store.Save(creds); err != nil {
		t.Fatal(err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.DeviceID != "abc123def456" {
		t.Errorf("DeviceID = %q, want abc123def456", loaded.DeviceID)
	}
}

func TestCredentials_JSONSerialization(t *testing.T) {
	creds := Credentials{
		APIKey:    "test-key",
		Username:  "testuser",
		UserID:    "uid-123",
		CreatedAt: "2026-01-01T00:00:00Z",
	}

	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatal(err)
	}

	var parsed Credentials
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}

	if parsed.APIKey != creds.APIKey {
		t.Errorf("APIKey mismatch: %q vs %q", parsed.APIKey, creds.APIKey)
	}
	if parsed.Username != creds.Username {
		t.Errorf("Username mismatch: %q vs %q", parsed.Username, creds.Username)
	}
}
