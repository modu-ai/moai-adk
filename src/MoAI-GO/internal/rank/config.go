package rank

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// APIBaseURL is the MoAI Rank API base URL
	APIBaseURL = "https://rank.mo.ai.kr/api/v1"
)

// Credentials stores MoAI Rank authentication info
type Credentials struct {
	APIKey    string `json:"api_key"`
	Username  string `json:"username"`
	UserID    string `json:"user_id"`
	CreatedAt string `json:"created_at"`
}

// RankConfig manages rank configuration and exclusions
type RankConfig struct {
	ExcludeProjects []string `yaml:"exclude_projects" json:"exclude_projects"`
}

// credentialsPath returns ~/.moai/rank/credentials.json
func credentialsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".moai", "rank", "credentials.json"), nil
}

// configPath returns ~/.moai/rank/config.yaml
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".moai", "rank", "config.yaml"), nil
}

// LoadCredentials reads credentials from disk
func LoadCredentials() (*Credentials, error) {
	path, err := credentialsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("not logged in. Run 'moai-go rank login' first")
	}
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("invalid credentials file: %w", err)
	}
	return &creds, nil
}

// SaveCredentials writes credentials to disk with restricted permissions
func SaveCredentials(creds *Credentials) error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0600)
}

// DeleteCredentials removes the credentials file
func DeleteCredentials() error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// IsLoggedIn checks if valid credentials exist
func IsLoggedIn() bool {
	creds, err := LoadCredentials()
	return err == nil && creds.APIKey != ""
}

// LoadRankConfig reads the rank config from disk
func LoadRankConfig() (*RankConfig, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &RankConfig{}, nil
		}
		return nil, err
	}
	var cfg RankConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveRankConfig writes the rank config to disk
func SaveRankConfig(cfg *RankConfig) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// AddExclusion adds a project path to the exclusion list
func AddExclusion(projectPath string) error {
	cfg, err := LoadRankConfig()
	if err != nil {
		return err
	}
	normalized := filepath.Clean(projectPath)
	for _, p := range cfg.ExcludeProjects {
		if p == normalized {
			return nil // Already excluded
		}
	}
	cfg.ExcludeProjects = append(cfg.ExcludeProjects, normalized)
	return SaveRankConfig(cfg)
}

// RemoveExclusion removes a project path from the exclusion list
func RemoveExclusion(projectPath string) error {
	cfg, err := LoadRankConfig()
	if err != nil {
		return err
	}
	normalized := filepath.Clean(projectPath)
	var filtered []string
	for _, p := range cfg.ExcludeProjects {
		if p != normalized {
			filtered = append(filtered, p)
		}
	}
	cfg.ExcludeProjects = filtered
	return SaveRankConfig(cfg)
}
