package wizard

import (
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// moaiSectionsDir returns the path to the sections directory relative to the project root.
func moaiSectionsDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "config", "sections")
}

// ReadLocaleFromProject reads the conversation_language from language.yaml.
// Returns an empty string if the file is missing or parsing fails.
func ReadLocaleFromProject(projectRoot string) string {
	langPath := filepath.Join(moaiSectionsDir(projectRoot), "language.yaml")
	data, err := os.ReadFile(langPath)
	if err != nil {
		return ""
	}

	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return ""
	}

	langSection, ok := parsed["language"].(map[string]any)
	if !ok {
		return ""
	}

	locale, _ := langSection["conversation_language"].(string)
	return locale
}

// ReadGitHubUsernameFromConfig reads the github_username from user.yaml.
// Returns an empty string if the file is missing or the field is absent.
func ReadGitHubUsernameFromConfig(projectRoot string) string {
	return readUserField(projectRoot, "github_username")
}

// ReadGitLabUsernameFromConfig reads the gitlab_username from user.yaml.
// Returns an empty string if the file is missing or the field is absent.
func ReadGitLabUsernameFromConfig(projectRoot string) string {
	return readUserField(projectRoot, "gitlab_username")
}

// readUserField reads a specific field from the user section of user.yaml.
func readUserField(projectRoot, field string) string {
	userPath := filepath.Join(moaiSectionsDir(projectRoot), "user.yaml")
	data, err := os.ReadFile(userPath)
	if err != nil {
		return ""
	}

	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return ""
	}

	userSection, ok := parsed["user"].(map[string]any)
	if !ok {
		return ""
	}

	value, _ := userSection[field].(string)
	return value
}

// IsGhAuthenticated checks whether the gh CLI is authenticated.
// Returns false if gh is not installed or not authenticated.
func IsGhAuthenticated() bool {
	// Check whether the gh CLI is available
	if _, err := exec.LookPath("gh"); err != nil {
		return false
	}

	// Run gh auth status — exits 0 when authenticated
	cmd := exec.Command("gh", "auth", "status")
	err := cmd.Run()
	return err == nil
}
