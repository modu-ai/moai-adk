package llmswitch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/anthropics/moai-adk-go/internal/output"
)

// GLM environment variable keys managed during backend switching
var GLMEnvKeys = []string{
	"ANTHROPIC_AUTH_TOKEN",
	"ANTHROPIC_BASE_URL",
	"ANTHROPIC_DEFAULT_HAIKU_MODEL",
	"ANTHROPIC_DEFAULT_SONNET_MODEL",
	"ANTHROPIC_DEFAULT_OPUS_MODEL",
}

// envVarPattern matches ${VAR_NAME} patterns for substitution
var envVarPattern = regexp.MustCompile(`\$\{(\w+)\}`)

// SwitchToClaude removes GLM environment variables from settings.local.json
func SwitchToClaude(projectDir string) error {
	settingsLocal := filepath.Join(projectDir, ".claude", "settings.local.json")

	if !HasGLMEnv(settingsLocal) {
		fmt.Println(output.WarningStyle.Render("Already using Claude backend."))
		return nil
	}

	data, err := loadJSON(settingsLocal)
	if err != nil {
		fmt.Println(output.WarningStyle.Render("Already using Claude backend."))
		return nil
	}

	env, ok := data["env"].(map[string]any)
	if !ok {
		fmt.Println(output.WarningStyle.Render("Already using Claude backend."))
		return nil
	}

	for _, key := range GLMEnvKeys {
		delete(env, key)
	}

	if len(env) == 0 {
		delete(data, "env")
	} else {
		data["env"] = env
	}

	if err := saveJSON(settingsLocal, data); err != nil {
		return fmt.Errorf("failed to write settings.local.json: %w", err)
	}

	fmt.Println()
	fmt.Println(output.SuccessStyle.Render("Switched to Claude backend"))
	fmt.Println(output.MutedStyle.Render("  Removed GLM env from: .claude/settings.local.json"))
	fmt.Println()
	fmt.Println(output.WarningStyle.Render("  Restart Claude Code to apply changes."))
	return nil
}

// SwitchToGLM loads GLM config template, substitutes env vars, and merges into settings.local.json
func SwitchToGLM(projectDir string) error {
	settingsLocal := filepath.Join(projectDir, ".claude", "settings.local.json")
	glmConfigPath := filepath.Join(projectDir, ".moai", "llm-configs", "glm.json")

	if HasGLMEnv(settingsLocal) {
		fmt.Println(output.WarningStyle.Render("Already using GLM backend."))
		return nil
	}

	if !GLMEnvExists() {
		fmt.Println(output.ErrorStyle.Render("Error: GLM API key not found."))
		fmt.Println(output.MutedStyle.Render("  Please run: moai-go glm <your-api-key> to set your key first."))
		return fmt.Errorf("GLM API key not found")
	}

	// Show masked key (error ignored: GLMEnvExists() already verified)
	existingKey, err := LoadGLMKeyFromEnv()
	if err == nil && existingKey != "" {
		masked := existingKey
		if len(masked) > 8 {
			masked = masked[:8] + "..."
		}
		fmt.Println(output.SuccessStyle.Render(fmt.Sprintf("✓ Using GLM API key: %s", masked)))
	}

	// Check GLM config exists
	if _, err := os.Stat(glmConfigPath); os.IsNotExist(err) {
		fmt.Println(output.ErrorStyle.Render("Error: GLM config not found at .moai/llm-configs/glm.json"))
		fmt.Println(output.MutedStyle.Render("  Run 'moai-go init' or create the config manually."))
		return fmt.Errorf("GLM config not found")
	}

	// Load GLM config template
	glmData, err := loadJSON(glmConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load GLM config: %w", err)
	}

	glmEnv, ok := glmData["env"].(map[string]any)
	if !ok {
		return fmt.Errorf("GLM config missing 'env' section")
	}

	// Substitute environment variables
	substitutedEnv := make(map[string]any)
	var allMissing []string
	for key, value := range glmEnv {
		strVal, ok := value.(string)
		if !ok {
			substitutedEnv[key] = value
			continue
		}
		result, missing := substituteEnvVars(strVal)
		substitutedEnv[key] = result
		allMissing = append(allMissing, missing...)
	}

	if len(allMissing) > 0 {
		unique := uniqueStrings(allMissing)
		fmt.Println(output.ErrorStyle.Render(fmt.Sprintf("Error: Missing credential(s): %s", strings.Join(unique, ", "))))
		fmt.Println(output.MutedStyle.Render("  Please run 'moai-go glm <key>' to set your API key."))
		return fmt.Errorf("missing credentials: %s", strings.Join(unique, ", "))
	}

	// Load or create settings.local.json
	localData, err := loadJSON(settingsLocal)
	if err != nil {
		localData = make(map[string]any)
	}

	// Merge GLM env
	existingEnv, ok := localData["env"].(map[string]any)
	if !ok {
		existingEnv = make(map[string]any)
	}
	for key, val := range substitutedEnv {
		existingEnv[key] = val
	}
	localData["env"] = existingEnv

	// Ensure .claude directory exists
	claudeDir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	if err := saveJSON(settingsLocal, localData); err != nil {
		return fmt.Errorf("failed to write settings.local.json: %w", err)
	}

	fmt.Println()
	fmt.Println(output.SuccessStyle.Render("Switched to GLM backend"))
	fmt.Println(output.MutedStyle.Render("  Added GLM env to: .claude/settings.local.json"))
	fmt.Println(output.MutedStyle.Render("  Environment variables have been substituted."))
	fmt.Println()
	fmt.Println(output.WarningStyle.Render("  Restart Claude Code to apply changes."))
	return nil
}

// UpdateGLMKey saves a GLM API key to ~/.moai/.env.glm
func UpdateGLMKey(apiKey string) error {
	envGLMPath, err := getEnvGLMPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(envGLMPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	content := fmt.Sprintf("# GLM API Key for MoAI-ADK\n# Generated by moai-go\nGLM_API_KEY=\"%s\"\n", apiKey)
	if err := os.WriteFile(envGLMPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write .env.glm: %w", err)
	}

	fmt.Println(output.SuccessStyle.Render(fmt.Sprintf("✓ GLM API key updated: %s", envGLMPath)))
	fmt.Println(output.MutedStyle.Render("  Use 'moai-go glm' to switch to GLM backend."))
	return nil
}

// HasGLMEnv checks if settings.local.json has GLM environment variables
func HasGLMEnv(settingsLocalPath string) bool {
	data, err := loadJSON(settingsLocalPath)
	if err != nil {
		return false
	}
	env, ok := data["env"].(map[string]any)
	if !ok {
		return false
	}
	_, exists := env["ANTHROPIC_BASE_URL"]
	return exists
}

// GLMEnvExists checks if ~/.moai/.env.glm exists
func GLMEnvExists() bool {
	path, err := getEnvGLMPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// LoadGLMKeyFromEnv reads the GLM API key from ~/.moai/.env.glm
func LoadGLMKeyFromEnv() (string, error) {
	path, err := getEnvGLMPath()
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read .env.glm: %w", err)
	}
	return parseEnvFile(string(content), "GLM_API_KEY"), nil
}

// substituteEnvVars replaces ${VAR_NAME} patterns with credential values
func substituteEnvVars(value string) (string, []string) {
	var missing []string
	result := envVarPattern.ReplaceAllStringFunc(value, func(match string) string {
		varName := envVarPattern.FindStringSubmatch(match)[1]
		credValue := getCredentialValue(varName)
		if credValue == "" {
			missing = append(missing, varName)
			return match
		}
		return credValue
	})
	return result, missing
}

// getCredentialValue resolves a variable name to its value
func getCredentialValue(varName string) string {
	// For GLM keys, check .env.glm first
	if varName == "GLM_API_KEY" || varName == "GLM_API_TOKEN" {
		key, err := LoadGLMKeyFromEnv()
		if err == nil && key != "" {
			return key
		}
	}
	// Fall back to environment variable
	return os.Getenv(varName)
}

// getEnvGLMPath returns the path to ~/.moai/.env.glm
func getEnvGLMPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".moai", ".env.glm"), nil
}

// parseEnvFile extracts a value for a key from dotenv-format content
func parseEnvFile(content, key string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		if k == key {
			// Remove surrounding quotes
			v = strings.Trim(v, `"'`)
			return v
		}
	}
	return ""
}

// loadJSON reads and parses a JSON file
func loadJSON(path string) (map[string]any, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data map[string]any
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// saveJSON writes a map as formatted JSON to a file
func saveJSON(path string, data map[string]any) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(content, '\n'), 0644)
}

// uniqueStrings returns a deduplicated slice
func uniqueStrings(input []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range input {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
