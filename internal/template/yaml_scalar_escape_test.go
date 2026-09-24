package template

// Card t1162: user-supplied strings rendered into double-quoted YAML scalars of
// the config section templates must parse back to exactly the value supplied.
// Before the fix the templates interpolated the raw value between the quotes,
// so `Kim "Goos"` made the file unparseable and `a\b` silently became a
// backspace (0x08).

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// hostileScalarValues are values a user can type (wizard answer, flag, profile
// preference, or a hand edit the update carries forward) that are ordinary
// strings but carry YAML-significant characters.
var hostileScalarValues = []struct{ label, value string }{
	{"embedded double quotes", `Kim "Goos"`},
	{"backslash escape", `a\b`},
	{"double backslash", `a\\b`},
	{"colon space and hash", `a: b #c`},
	{"leading single quote", `'quoted`},
	{"leading ampersand", `&anchor`},
	{"leading asterisk", `*alias`},
	{"hangul and emoji", "홍길동 🚀"},
	{"next line U+0085", "a\u0085b"},
	{"line separator U+2028", "a\u2028b"},
	{"tab", "a\tb"},
	{"newline", "a\nb"},
	{"bell control", "a\x07b"},
	{"delete control", "a\x7fb"},
	{"byte order mark", "a\ufeffb"},
	{"noncharacter U+FFFE", "a\ufffeb"},
	{"carriage return", "a\rb"},
	{"C1 control U+009B", "a\u009bb"},
}

// userInputScalars lists every template field that carries a user-supplied
// string into a YAML scalar, with the section template and the key path it
// renders under.
var userInputScalars = []struct {
	field, tmpl string
	set         func(*TemplateContext, string)
	path        []string
}{
	{"UserName", ".moai/config/sections/user.yaml.tmpl", func(c *TemplateContext, v string) { c.UserName = v }, []string{"user", "name"}},
	{"ProjectName", ".moai/config/sections/project.yaml.tmpl", func(c *TemplateContext, v string) { c.ProjectName = v }, []string{"project", "name"}},
	{"GitHubUsername", ".moai/config/sections/git-strategy.yaml.tmpl", func(c *TemplateContext, v string) { c.GitHubUsername = v }, []string{"git_strategy", "github_username"}},
	{"GitLabInstanceURL", ".moai/config/sections/git-strategy.yaml.tmpl", func(c *TemplateContext, v string) { c.GitLabInstanceURL = v }, []string{"git_strategy", "gitlab", "instance_url"}},
	{"ConversationLanguage", ".moai/config/sections/language.yaml.tmpl", func(c *TemplateContext, v string) { c.ConversationLanguage = v }, []string{"language", "conversation_language"}},
	{"GitCommitMessages", ".moai/config/sections/language.yaml.tmpl", func(c *TemplateContext, v string) { c.GitCommitMessages = v }, []string{"language", "git_commit_messages"}},
	{"CodeComments", ".moai/config/sections/language.yaml.tmpl", func(c *TemplateContext, v string) { c.CodeComments = v }, []string{"language", "code_comments"}},
	{"Documentation", ".moai/config/sections/language.yaml.tmpl", func(c *TemplateContext, v string) { c.Documentation = v }, []string{"language", "documentation"}},
}

// renderScalar renders tmpl through the embedded renderer init and update use
// and returns the string found at path, or a failure description.
func renderScalar(t *testing.T, r Renderer, tmpl string, ctx *TemplateContext, path []string) (got string, problem string) {
	t.Helper()
	out, err := r.Render(tmpl, ctx)
	if err != nil {
		return "", "render error: " + err.Error()
	}
	var node any
	if err := yaml.Unmarshal(out, &node); err != nil {
		return "", "rendered YAML does not parse: " + err.Error()
	}
	for _, key := range path {
		m, ok := node.(map[string]any)
		if !ok {
			return "", "path " + strings.Join(path, ".") + " reached a non-mapping"
		}
		node = m[key]
	}
	s, ok := node.(string)
	if !ok {
		return "", "value at " + strings.Join(path, ".") + " is not a string"
	}
	return s, ""
}

func TestSectionTemplates_UserInputScalarsRoundTrip(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	r := NewRenderer(fsys)
	for _, f := range userInputScalars {
		for _, v := range hostileScalarValues {
			t.Run(f.field+"/"+v.label, func(t *testing.T) {
				ctx := NewTemplateContext()
				f.set(ctx, v.value)
				got, problem := renderScalar(t, r, f.tmpl, ctx, f.path)
				if problem != "" {
					t.Fatalf("%s = %q: %s", f.field, v.value, problem)
				}
				if got != v.value {
					t.Fatalf("%s round-trip: rendered %q parses back as %q", f.field, v.value, got)
				}
			})
		}
	}
}
