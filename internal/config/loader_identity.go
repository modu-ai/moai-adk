package config

// loader_identity.go — single-key readers for project.name and user.name,
// modelled on LoadGitMode. Their consumer is the update render path: the update
// builds its template context outside the Loader's lifecycle and, on the
// template-sync path, must read the values before the managed cleanup removes
// .moai/config. Rendering project.yaml / user.yaml with the name the project
// already carries keeps an unchanged name byte-identical across an update
// instead of rendering it empty.
//
// The readers return the stored value verbatim. Whether the update render can
// carry that value is decided by the caller against the real renderer
// (internal/cli loadUpdateIdentity): internal/template imports this package, so
// the renderer is not reachable from here.

import (
	"path/filepath"
)

type projectNameFileWrapper struct {
	Project struct {
		Name string `yaml:"name"`
	} `yaml:"project"`
}

type userNameFileWrapper struct {
	User struct {
		Name string `yaml:"name"`
	} `yaml:"user"`
}

// LoadProjectName returns project.name for the project rooted at projectRoot,
// or the empty string when the key is absent, the file is missing, or the file
// cannot be parsed. The empty string is also the template-context default, so
// a failed read renders exactly what the update rendered before this reader.
func LoadProjectName(projectRoot string) string {
	w := &projectNameFileWrapper{}
	loaded, err := loadYAMLFile(filepath.Join(projectRoot, ".moai", "config", "sections"), "project.yaml", w)
	if err != nil || !loaded {
		return ""
	}
	return w.Project.Name
}

// LoadUserName returns user.name for the project rooted at projectRoot, with
// the same empty-string fallback as LoadProjectName.
func LoadUserName(projectRoot string) string {
	w := &userNameFileWrapper{}
	loaded, err := loadYAMLFile(filepath.Join(projectRoot, ".moai", "config", "sections"), "user.yaml", w)
	if err != nil || !loaded {
		return ""
	}
	return w.User.Name
}
