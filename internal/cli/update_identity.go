package cli

// update_identity.go — the user-owned section values the update render context
// carries: project.name / user.name (card t1139) and the language, development
// mode, and git provider values (card t1147). The update renders the section
// files with the values the project already has, so an unchanged value stays
// byte-identical across the update instead of reverting to the template default.

import (
	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
)

const (
	projectSectionTemplate     = ".moai/config/sections/project.yaml.tmpl"
	userSectionTemplate        = ".moai/config/sections/user.yaml.tmpl"
	languageSectionTemplate    = ".moai/config/sections/language.yaml.tmpl"
	qualitySectionTemplate     = ".moai/config/sections/quality.yaml.tmpl"
	gitStrategySectionTemplate = ".moai/config/sections/git-strategy.yaml.tmpl"
)

// @MX:NOTE: [AUTO] single source of user-owned values for all three update render contexts (template-sync validate + deploy, clean-reinstall deploy); rendersVerbatim decides whether a value survives the update
// loadUpdateUserValues returns one context option carrying every user-owned
// section value the update renders with: the names from loadUpdateIdentity,
// plus conversation_language, the output languages, development_mode, and the
// git provider, github_username, and gitlab.instance_url (card t1147). Each of
// the latter is carried under the same exact round-trip rule as the names
// (see loadUpdateIdentity): its option is applied only if rendering its
// section template with it parses back to exactly the stored value. An empty,
// unknown, or non-round-tripping value leaves the template default, and the
// 3-way merge then keeps a hand edit as a customization.
func loadUpdateUserValues(projectRoot string) template.ContextOption {
	projectName, userName := loadUpdateIdentity(projectRoot)
	opts := []template.ContextOption{
		template.WithProject(projectName, projectRoot),
		template.WithUser(userName),
	}
	if fsys, err := template.EmbeddedTemplates(); err == nil {
		renderer := template.NewRenderer(fsys)
		v := config.LoadUpdateRenderValues(projectRoot)
		for _, k := range []struct {
			tmpl, value string
			opt         template.ContextOption
			path        []string
		}{
			{languageSectionTemplate, v.ConversationLanguage, template.WithLanguage(v.ConversationLanguage), []string{"language", "conversation_language"}},
			{languageSectionTemplate, v.GitCommitMessages, template.WithOutputLanguages(v.GitCommitMessages, "", ""), []string{"language", "git_commit_messages"}},
			{languageSectionTemplate, v.CodeComments, template.WithOutputLanguages("", v.CodeComments, ""), []string{"language", "code_comments"}},
			{languageSectionTemplate, v.Documentation, template.WithOutputLanguages("", "", v.Documentation), []string{"language", "documentation"}},
			{qualitySectionTemplate, v.DevelopmentMode, template.WithDevelopmentMode(v.DevelopmentMode), []string{"constitution", "development_mode"}},
			{gitStrategySectionTemplate, v.GitProvider, template.WithGitProvider(v.GitProvider), []string{"git_strategy", "provider"}},
			{gitStrategySectionTemplate, v.GitHubUsername, template.WithGitHubUsername(v.GitHubUsername), []string{"git_strategy", "github_username"}},
			{gitStrategySectionTemplate, v.GitLabInstanceURL, template.WithGitLabInstanceURL(v.GitLabInstanceURL), []string{"git_strategy", "gitlab", "instance_url"}},
		} {
			// Each value is checked alone, so one unrenderable value cannot
			// break the parse of its neighbours in the same file.
			if k.value != "" && rendersVerbatim(renderer, k.tmpl, template.NewTemplateContext(k.opt), k.value, k.path...) {
				opts = append(opts, k.opt)
			}
		}
	}
	return func(c *template.TemplateContext) {
		for _, opt := range opts {
			opt(c)
		}
	}
}

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
//     hand edit no render could have written ("$TEAM", "{{.Version}}" — the
//     renderer's unexpanded-token guard rejects both), its BASE differs from
//     the user's value and the 3-way merge keeps the user's value as a
//     customization.
//
// The section templates escape the name inside its double-quoted scalar
// (yamlEscape, card t1162), so a name carrying `"` or `\` renders and parses
// back verbatim; init stores it unaltered and the update keeps it. The
// remaining gap is a stored value that does not round-trip AND equals BASE:
// the merge takes the empty render. An init render no longer produces such a
// value: a name the renderer rejects makes the deploy return the render error
// instead of writing the file.
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
	if !rendersVerbatim(renderer, projectSectionTemplate, ctx, projectName, "project", "name") {
		projectName = ""
	}
	if !rendersVerbatim(renderer, userSectionTemplate, ctx, userName, "user", "name") {
		userName = ""
	}
	return projectName, userName
}

// rendersVerbatim reports whether rendering tmpl with ctx succeeds and the
// output's value at path parses back as exactly value. The empty value is
// always accepted.
func rendersVerbatim(renderer template.Renderer, tmpl string, ctx *template.TemplateContext, value string, path ...string) bool {
	if value == "" {
		return true
	}
	out, err := renderer.Render(tmpl, ctx)
	if err != nil {
		return false
	}
	var node any
	if err := yaml.Unmarshal(out, &node); err != nil {
		return false
	}
	for _, key := range path {
		m, ok := node.(map[string]any)
		if !ok {
			return false
		}
		node = m[key]
	}
	got, ok := node.(string)
	return ok && got == value
}
