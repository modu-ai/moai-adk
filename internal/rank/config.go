package rank

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// Credentials holds the user's authentication credentials.
type Credentials struct {
	APIKey    string `json:"api_key"`
	Username  string `json:"username"`
	UserID    string `json:"user_id"`
	CreatedAt string `json:"created_at"`
	DeviceID  string `json:"device_id,omitempty"`
}

// Config holds Rank service configuration.
type Config struct {
	BaseURL string `json:"base_url"`
}

// NewConfig creates a Config with defaults and environment variable overrides.
func NewConfig() *Config {
	cfg := &Config{
		BaseURL: DefaultBaseURL,
	}
	if envURL := os.Getenv("MOAI_RANK_API_URL"); envURL != "" {
		cfg.BaseURL = envURL
	}
	return cfg
}

// CredentialStore defines the interface for credential persistence.
type CredentialStore interface {
	Save(creds *Credentials) error
	Load() (*Credentials, error)
	Delete() error
	HasCredentials() bool
	GetAPIKey() (string, error)
}

// FileCredentialStore implements CredentialStore using the filesystem.
// Credentials are stored at ~/.moai/rank/credentials.json.
type FileCredentialStore struct {
	dir      string
	credPath string
}

// Compile-time interface check.
var _ CredentialStore = (*FileCredentialStore)(nil)

// NewFileCredentialStore creates a FileCredentialStore with the given directory.
// If dir is empty, it defaults to ~/.moai/rank/.
func NewFileCredentialStore(dir string) *FileCredentialStore {
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		dir = filepath.Join(home, defs.MoAIDir, defs.RankSubdir)
	}
	return &FileCredentialStore{
		dir:      dir,
		credPath: filepath.Join(dir, defs.CredentialsJSON),
	}
}

// Save persists credentials to disk atomically with secure file permissions.
// Directory permissions: 0700, file permissions: 0600.
func (s *FileCredentialStore) Save(creds *Credentials) error {
	if err := os.MkdirAll(s.dir, defs.CredDirPerm); err != nil {
		return fmt.Errorf("create credential directory: %w", err)
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	// Atomic write: write to temp file, then rename.
	tmpFile := s.credPath + ".tmp"
	if err := os.WriteFile(tmpFile, data, defs.CredFilePerm); err != nil {
		return fmt.Errorf("write temp credential file: %w", err)
	}

	if err := os.Rename(tmpFile, s.credPath); err != nil {
		// Clean up temp file on rename failure.
		_ = os.Remove(tmpFile)
		return fmt.Errorf("rename credential file: %w", err)
	}

	return nil
}

// Load reads credentials from disk.
// Returns nil without error if the file does not exist or contains invalid JSON.
func (s *FileCredentialStore) Load() (*Credentials, error) {
	data, err := os.ReadFile(s.credPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read credential file: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		// Graceful handling of corrupted JSON.
		return nil, nil
	}

	return &creds, nil
}

// Delete removes the credentials file.
// Returns nil if the file does not exist.
func (s *FileCredentialStore) Delete() error {
	err := os.Remove(s.credPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove credential file: %w", err)
	}
	return nil
}

// HasCredentials checks whether a credentials file exists.
func (s *FileCredentialStore) HasCredentials() bool {
	_, err := os.Stat(s.credPath)
	return err == nil
}

// GetAPIKey loads credentials and returns only the API key.
// Returns an empty string without error if no credentials exist.
func (s *FileCredentialStore) GetAPIKey() (string, error) {
	creds, err := s.Load()
	if err != nil {
		return "", err
	}
	if creds == nil {
		return "", nil
	}
	return creds.APIKey, nil
}
