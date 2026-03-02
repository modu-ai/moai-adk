package wizard

import (
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// moaiSectionsDir은 프로젝트 루트 기준 sections 디렉토리 경로를 반환한다.
func moaiSectionsDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "config", "sections")
}

// ReadLocaleFromProject는 language.yaml에서 conversation_language를 읽는다.
// 파일이 없거나 파싱에 실패하면 빈 문자열을 반환한다.
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

// ReadGitHubUsernameFromConfig는 user.yaml에서 github_username을 읽는다.
// 파일이 없거나 값이 없으면 빈 문자열을 반환한다.
func ReadGitHubUsernameFromConfig(projectRoot string) string {
	return readUserField(projectRoot, "github_username")
}

// ReadGitLabUsernameFromConfig는 user.yaml에서 gitlab_username을 읽는다.
// 파일이 없거나 값이 없으면 빈 문자열을 반환한다.
func ReadGitLabUsernameFromConfig(projectRoot string) string {
	return readUserField(projectRoot, "gitlab_username")
}

// readUserField는 user.yaml의 user 섹션에서 특정 필드를 읽는다.
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

// IsGhAuthenticated는 gh CLI가 인증되어 있는지 확인한다.
// gh가 없거나 인증되지 않은 경우 false를 반환한다.
func IsGhAuthenticated() bool {
	// gh CLI 존재 여부 확인
	if _, err := exec.LookPath("gh"); err != nil {
		return false
	}

	// gh auth status 실행 — 인증된 경우 exit 0
	cmd := exec.Command("gh", "auth", "status")
	err := cmd.Run()
	return err == nil
}
