package initializer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anthropics/moai-adk-go/internal/output"
	"gopkg.in/yaml.v3"
)

// LanguageConfig represents the language.yaml structure
type LanguageConfig struct {
	ConversationLanguage     string `yaml:"conversation_language"`
	ConversationLanguageName string `yaml:"conversation_language_name"`
	AgentPromptLanguage      string `yaml:"agent_prompt_language"`
	GitCommitMessages        string `yaml:"git_commit_messages"`
	CodeComments             string `yaml:"code_comments"`
	Documentation            string `yaml:"documentation"`
	ErrorMessages            string `yaml:"error_messages"`
}

// UserConfig represents the user.yaml structure
type UserConfig struct {
	User struct {
		Name string `yaml:"name"`
	} `yaml:"user"`
}

// ConfigWriter handles writing configuration files
type ConfigWriter struct {
	targetDir string
}

// NewConfigWriter creates a new config writer
func NewConfigWriter(targetDir string) *ConfigWriter {
	return &ConfigWriter{
		targetDir: targetDir,
	}
}

// WriteLanguageConfig writes the language.yaml file
func (cw *ConfigWriter) WriteLanguageConfig(code LanguageCode) error {
	// Create language config based on selection
	config := LanguageConfig{
		ConversationLanguage:     string(code),
		ConversationLanguageName: getLanguageName(code),
		AgentPromptLanguage:      "en",
		GitCommitMessages:        "en",
		CodeComments:             "en",
		Documentation:            "en",
		ErrorMessages:            "en",
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling language config: %w", err)
	}

	// Create config directory
	configDir := filepath.Join(cw.targetDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Write file
	langFile := filepath.Join(configDir, "language.yaml")
	if err := os.WriteFile(langFile, data, 0644); err != nil {
		return fmt.Errorf("error writing language.yaml: %w", err)
	}

	fmt.Println(output.SuccessStyle.Render(fmt.Sprintf("✓ Created: %s", langFile)))
	return nil
}

// WriteUserConfig writes the user.yaml file
func (cw *ConfigWriter) WriteUserConfig(userName string) error {
	// Create user config
	config := UserConfig{}
	config.User.Name = userName

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling user config: %w", err)
	}

	// Create config directory
	configDir := filepath.Join(cw.targetDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Write file
	userFile := filepath.Join(configDir, "user.yaml")
	if err := os.WriteFile(userFile, data, 0644); err != nil {
		return fmt.Errorf("error writing user.yaml: %w", err)
	}

	fmt.Println(output.SuccessStyle.Render(fmt.Sprintf("✓ Created: %s", userFile)))
	return nil
}

// getLanguageName returns the display name for a language code
func getLanguageName(code LanguageCode) string {
	switch code {
	case LanguageEnglish:
		return "English"
	case LanguageKorean:
		return "Korean (한국어)"
	case LanguageJapanese:
		return "Japanese (日本語)"
	case LanguageChinese:
		return "Chinese (中文)"
	default:
		return "English"
	}
}
