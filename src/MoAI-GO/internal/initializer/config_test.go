package initializer

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// --- NewConfigWriter ---

func TestNewConfigWriter(t *testing.T) {
	cw := NewConfigWriter("/tmp/project")
	if cw == nil {
		t.Fatal("NewConfigWriter returned nil")
	}
	if cw.targetDir != "/tmp/project" {
		t.Errorf("targetDir = %q, want %q", cw.targetDir, "/tmp/project")
	}
}

// --- WriteLanguageConfig ---

func TestWriteLanguageConfig(t *testing.T) {
	tests := []struct {
		name     string
		code     LanguageCode
		wantLang string
		wantName string
	}{
		{
			name:     "English",
			code:     LanguageEnglish,
			wantLang: "en",
			wantName: "English",
		},
		{
			name:     "Korean",
			code:     LanguageKorean,
			wantLang: "ko",
			wantName: "Korean (한국어)",
		},
		{
			name:     "Japanese",
			code:     LanguageJapanese,
			wantLang: "ja",
			wantName: "Japanese (日本語)",
		},
		{
			name:     "Chinese",
			code:     LanguageChinese,
			wantLang: "zh",
			wantName: "Chinese (中文)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cw := NewConfigWriter(tmpDir)

			if err := cw.WriteLanguageConfig(tt.code); err != nil {
				t.Fatalf("WriteLanguageConfig() error = %v", err)
			}

			// Verify file was created
			langFile := filepath.Join(tmpDir, ".moai", "config", "sections", "language.yaml")
			data, err := os.ReadFile(langFile)
			if err != nil {
				t.Fatalf("failed to read language.yaml: %v", err)
			}

			// Unmarshal and check
			var config LanguageConfig
			if err := yaml.Unmarshal(data, &config); err != nil {
				t.Fatalf("failed to unmarshal language.yaml: %v", err)
			}

			if config.ConversationLanguage != tt.wantLang {
				t.Errorf("ConversationLanguage = %q, want %q", config.ConversationLanguage, tt.wantLang)
			}
			if config.ConversationLanguageName != tt.wantName {
				t.Errorf("ConversationLanguageName = %q, want %q", config.ConversationLanguageName, tt.wantName)
			}
			if config.AgentPromptLanguage != "en" {
				t.Errorf("AgentPromptLanguage = %q, want %q", config.AgentPromptLanguage, "en")
			}
			if config.GitCommitMessages != "en" {
				t.Errorf("GitCommitMessages = %q, want %q", config.GitCommitMessages, "en")
			}
			if config.CodeComments != "en" {
				t.Errorf("CodeComments = %q, want %q", config.CodeComments, "en")
			}
			if config.Documentation != "en" {
				t.Errorf("Documentation = %q, want %q", config.Documentation, "en")
			}
			if config.ErrorMessages != "en" {
				t.Errorf("ErrorMessages = %q, want %q", config.ErrorMessages, "en")
			}
		})
	}
}

// --- WriteUserConfig ---

func TestWriteUserConfig(t *testing.T) {
	tests := []struct {
		name     string
		userName string
	}{
		{name: "standard name", userName: "John Doe"},
		{name: "empty name", userName: ""},
		{name: "unicode name", userName: "Tester"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cw := NewConfigWriter(tmpDir)

			if err := cw.WriteUserConfig(tt.userName); err != nil {
				t.Fatalf("WriteUserConfig() error = %v", err)
			}

			// Verify file was created
			userFile := filepath.Join(tmpDir, ".moai", "config", "sections", "user.yaml")
			data, err := os.ReadFile(userFile)
			if err != nil {
				t.Fatalf("failed to read user.yaml: %v", err)
			}

			// Unmarshal and check
			var config UserConfig
			if err := yaml.Unmarshal(data, &config); err != nil {
				t.Fatalf("failed to unmarshal user.yaml: %v", err)
			}

			if config.User.Name != tt.userName {
				t.Errorf("User.Name = %q, want %q", config.User.Name, tt.userName)
			}
		})
	}
}

func TestWriteLanguageConfig_CreatesDirectoryStructure(t *testing.T) {
	tmpDir := t.TempDir()
	cw := NewConfigWriter(tmpDir)

	if err := cw.WriteLanguageConfig(LanguageEnglish); err != nil {
		t.Fatalf("WriteLanguageConfig() error = %v", err)
	}

	// Verify directory structure was created
	sectionsDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	info, err := os.Stat(sectionsDir)
	if err != nil {
		t.Fatalf("sections directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("sections path is not a directory")
	}
}

// --- WriteLanguageConfig error paths ---

func TestWriteLanguageConfig_UnknownLanguageCode(t *testing.T) {
	tmpDir := t.TempDir()
	cw := NewConfigWriter(tmpDir)

	// Unknown code should still work (falls back to "English")
	if err := cw.WriteLanguageConfig(LanguageCode("xx")); err != nil {
		t.Fatalf("WriteLanguageConfig() error = %v", err)
	}

	langFile := filepath.Join(tmpDir, ".moai", "config", "sections", "language.yaml")
	data, err := os.ReadFile(langFile)
	if err != nil {
		t.Fatalf("failed to read language.yaml: %v", err)
	}

	var config LanguageConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if config.ConversationLanguage != "xx" {
		t.Errorf("ConversationLanguage = %q, want %q", config.ConversationLanguage, "xx")
	}
	if config.ConversationLanguageName != "English" {
		t.Errorf("ConversationLanguageName = %q, want %q (default fallback)", config.ConversationLanguageName, "English")
	}
}

func TestWriteUserConfig_CreatesDirectoryStructure(t *testing.T) {
	tmpDir := t.TempDir()
	cw := NewConfigWriter(tmpDir)

	if err := cw.WriteUserConfig("TestUser"); err != nil {
		t.Fatalf("WriteUserConfig() error = %v", err)
	}

	sectionsDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	info, err := os.Stat(sectionsDir)
	if err != nil {
		t.Fatalf("sections directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("sections path is not a directory")
	}
}

func TestWriteUserConfig_SpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()
	cw := NewConfigWriter(tmpDir)

	specialName := "User O'Brien & Co."
	if err := cw.WriteUserConfig(specialName); err != nil {
		t.Fatalf("WriteUserConfig() error = %v", err)
	}

	userFile := filepath.Join(tmpDir, ".moai", "config", "sections", "user.yaml")
	data, err := os.ReadFile(userFile)
	if err != nil {
		t.Fatalf("failed to read user.yaml: %v", err)
	}

	var config UserConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if config.User.Name != specialName {
		t.Errorf("User.Name = %q, want %q", config.User.Name, specialName)
	}
}

// --- LanguageConfig struct ---

func TestLanguageConfigStruct(t *testing.T) {
	config := LanguageConfig{
		ConversationLanguage:     "ko",
		ConversationLanguageName: "Korean",
		AgentPromptLanguage:      "en",
		GitCommitMessages:        "en",
		CodeComments:             "en",
		Documentation:            "en",
		ErrorMessages:            "en",
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("yaml.Marshal error: %v", err)
	}

	var parsed LanguageConfig
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("yaml.Unmarshal error: %v", err)
	}

	if parsed.ConversationLanguage != "ko" {
		t.Errorf("ConversationLanguage = %q, want %q", parsed.ConversationLanguage, "ko")
	}
}

// --- UserConfig struct ---

func TestUserConfigStruct(t *testing.T) {
	config := UserConfig{}
	config.User.Name = "TestDeveloper"

	data, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("yaml.Marshal error: %v", err)
	}

	var parsed UserConfig
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("yaml.Unmarshal error: %v", err)
	}

	if parsed.User.Name != "TestDeveloper" {
		t.Errorf("User.Name = %q, want %q", parsed.User.Name, "TestDeveloper")
	}
}

// --- getLanguageName ---

func TestGetLanguageName(t *testing.T) {
	tests := []struct {
		code LanguageCode
		want string
	}{
		{LanguageEnglish, "English"},
		{LanguageKorean, "Korean (한국어)"},
		{LanguageJapanese, "Japanese (日本語)"},
		{LanguageChinese, "Chinese (中文)"},
		{LanguageCode("unknown"), "English"}, // default fallback
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			got := getLanguageName(tt.code)
			if got != tt.want {
				t.Errorf("getLanguageName(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}
