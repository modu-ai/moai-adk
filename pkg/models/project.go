// @MX:NOTE: [AUTO] 프로젝트 유형(web-app, api, cli, library)과 설정을 정의합니다. YAML에서 로드되는 데이터 모델입니다.
package models

// ProjectType represents the type of project.
type ProjectType string

const (
	ProjectTypeWebApp  ProjectType = "web-app"
	ProjectTypeAPI     ProjectType = "api"
	ProjectTypeCLI     ProjectType = "cli"
	ProjectTypeLibrary ProjectType = "library"
)

// ProjectConfig represents the project configuration.
type ProjectConfig struct {
	Name            string      `yaml:"name" json:"name"`
	Type            ProjectType `yaml:"type" json:"type"`
	Language        string      `yaml:"language" json:"language"`
	Framework       string      `yaml:"framework" json:"framework"`
	Description     string      `yaml:"description" json:"description"`
	TemplateVersion string      `yaml:"template_version" json:"template_version"`
}
