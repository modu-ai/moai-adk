package config

// loader_identity.go — single-key readers for project.name and user.name,
// modelled on LoadGitMode. Their consumer is the update render path: the update
// builds its template context outside the Loader's lifecycle and, on the
// template-sync path, must read the values before the managed cleanup removes
// .moai/config. Rendering project.yaml / user.yaml with the name the project
// already carries keeps an unchanged name byte-identical across an update
// instead of rendering it empty.

import (
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
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
// or the empty string when the key is absent, the file is missing, the file
// cannot be parsed, or the value cannot be rendered verbatim (see
// renderableIdentity). The empty string is also the template-context default,
// so a failed read renders exactly what the update rendered before this reader.
func LoadProjectName(projectRoot string) string {
	w := &projectNameFileWrapper{}
	loaded, err := loadYAMLFile(filepath.Join(projectRoot, ".moai", "config", "sections"), "project.yaml", w)
	if err != nil || !loaded {
		return ""
	}
	return renderableIdentity(w.Project.Name)
}

// LoadUserName returns user.name for the project rooted at projectRoot, with
// the same empty-string fallbacks as LoadProjectName.
func LoadUserName(projectRoot string) string {
	w := &userNameFileWrapper{}
	loaded, err := loadYAMLFile(filepath.Join(projectRoot, ".moai", "config", "sections"), "user.yaml", w)
	if err != nil || !loaded {
		return ""
	}
	return renderableIdentity(w.User.Name)
}

// renderableIdentity returns name when it survives the update render verbatim,
// and "" otherwise.
//
// The templates place the name inside a double-quoted YAML scalar
// (`name: "{{.ProjectName}}"`), and the renderer then rejects any output that
// still matches its unexpanded-token pattern (internal/template/renderer.go:
// `${VAR}`, `{{VAR}}`, `$VAR`). A name that trips the token check halts the
// whole update at "Validate Templates"; a name that breaks the scalar makes the
// section's 3-way merge fail on every update. The renderer's predicate cannot
// be reused here — internal/template imports this package — so this is a
// strict superset of both hazards, pinned against the real renderer by
// TestIdentityLoader_AcceptedNamesRenderVerbatim in internal/cli:
//
//   - `$`, `{{`, `}}`: any substring the token pattern can match begins with
//     `$` or `{{` inside the value (the surrounding `"` cannot start one);
//   - `"`, `\`: terminate or escape inside a double-quoted scalar;
//   - any non-graphic rune (control characters, line/paragraph separators,
//     format characters such as a byte-order mark): YAML rejects or rewrites
//     them inside a scalar.
//
// Returning "" is safe for the merge: the render then carries an empty name,
// the user's value differs from it, and the 3-way merge keeps the user's value
// as a customization — the pre-t1139 behavior for exactly these names.
func renderableIdentity(name string) string {
	if !utf8.ValidString(name) ||
		strings.ContainsAny(name, `$"\`) ||
		strings.Contains(name, "{{") ||
		strings.Contains(name, "}}") {
		return ""
	}
	for _, r := range name {
		if !unicode.IsGraphic(r) {
			return ""
		}
	}
	return name
}
