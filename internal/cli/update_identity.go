package cli

// update_identity.go — the project.name / user.name the update render context
// carries (card t1139). The update renders project.yaml / user.yaml with the
// names the project already has, so an unchanged name stays byte-identical
// across the update instead of rendering empty.

import (
	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
)

const (
	projectSectionTemplate = ".moai/config/sections/project.yaml.tmpl"
	userSectionTemplate    = ".moai/config/sections/user.yaml.tmpl"
)

// @MX:NOTE: [AUTO] single identity source for all three update render contexts (template-sync validate + deploy, clean-reinstall deploy); the accept rule decides whether a name survives the update
// loadUpdateIdentity returns the project and user names for the update render
// context, read from the project's existing config. A name is kept iff the
// embedded section template, rendered through the same renderer the update
// deploys with, renders without error and parses back to exactly that name;
// otherwise it is "".
//
// The rule is exact rather than conservative on purpose. The next merge's BASE
// is the snapshot of the previous render, so:
//
//   - A name init stored verbatim — the stored value equals what the render
//     produces for it — round-trips through the same render, so it is kept and
//     the merge sees NEW == BASE == OLD.
//   - A name whose stored form does not round-trip renders "". When it is a
//     hand edit no render could have written ("$TEAM", "{{.Version}}", `"`,
//     `\`), its BASE differs from the user's value and the 3-way merge keeps
//     the user's value as a customization.
//
// Known limitation: a stored value that does not round-trip AND equals BASE
// is erased — the merge takes the empty render. project.yaml.tmpl does not
// escape the name inside its double-quoted scalar, so a --name carrying a YAML
// backslash escape (`a\\b`, `a\nb`, `Kim \"Goos\"`) is stored by init as a
// different string that is also the snapshot BASE; the next update empties
// project.name. Pinned by TestUpdateForce_KnownLimitation_InitEscapedProjectName;
// the fix (escape identity values in the init/project template render) is a
// follow-up.
func loadUpdateIdentity(projectRoot string) (projectName, userName string) {
	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		return "", ""
	}
	renderer := template.NewRenderer(fsys)
	// The two section templates reference only the name, the project's
	// creation timestamp, and the version; the context carries the same values
	// the update's own render contexts carry for them.
	projectName = config.LoadProjectName(projectRoot)
	userName = config.LoadUserName(projectRoot)
	ctx := template.NewTemplateContext(
		template.WithVersion(version.GetVersion()),
		template.WithProject(projectName, projectRoot),
		template.WithUser(userName),
	)
	if !rendersVerbatim(renderer, projectSectionTemplate, "project", projectName, ctx) {
		projectName = ""
	}
	if !rendersVerbatim(renderer, userSectionTemplate, "user", userName, ctx) {
		userName = ""
	}
	return projectName, userName
}

// rendersVerbatim reports whether rendering tmpl with ctx succeeds and the
// output's <key>.name parses back as exactly name.
func rendersVerbatim(renderer template.Renderer, tmpl, key, name string, ctx *template.TemplateContext) bool {
	if name == "" {
		return true
	}
	out, err := renderer.Render(tmpl, ctx)
	if err != nil {
		return false
	}
	var doc map[string]map[string]any
	if err := yaml.Unmarshal(out, &doc); err != nil {
		return false
	}
	got, ok := doc[key]["name"].(string)
	return ok && got == name
}
