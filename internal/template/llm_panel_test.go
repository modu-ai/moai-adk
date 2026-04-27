package template

import (
	"testing"
	"testing/fstest"
)

// TestLLMPanelTemplate_Rendering verifies that llm-panel.yml.tmpl can be parsed
// and rendered without errors. This test reproduces issue #733 where unescaped
// GitHub Actions expressions ${{ }} cause Go text/template parse errors.
func TestLLMPanelTemplate_Rendering(t *testing.T) {
	t.Run("reproduce_issue_733", func(t *testing.T) {
		// This is the BROKEN content from llm-panel.yml.tmpl that causes the parse error
		// The issue is on lines 28, 33, 38 where ${{ }} expressions are not properly escaped
		brokenTemplate := `name: LLM Review Panel
on:
  pull_request:
    types: [opened]

permissions:
  contents: read
  pull-requests: write

jobs:
  panel:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Post LLM Panel Comment
        uses: actions/github-script@v7
        with:
          github-token: ${{ "{{" }} secrets.GITHUB_TOKEN }}}

          script: |
            gh pr comment ${{ "{{" }} github.event.pull_request.number }} --body "## LLM Code Review Status


| LLM | Status |
|-----|--------|
| Claude | Pending (add /claude comment) |
| Codex | ${{ "{{" } } } {
  if secrets.CODEX_AUTH_JSON != '' then 'Ready'
  else 'Token missing'
  end
{{ "}" } } |
`

		fs := fstest.MapFS{
			"llm-panel.yml.tmpl": &fstest.MapFile{
				Data: []byte(brokenTemplate),
			},
		}

		r := NewRenderer(fs)

		// This should fail with a parse error because of unescaped ${{ }} on line 28
		_, err := r.Render("llm-panel.yml.tmpl", nil)
		if err == nil {
			t.Error("expected parse error for unescaped GitHub Actions expressions, got nil")
		} else {
			t.Logf("Got expected error (reproducing issue #733): %v", err)
		}
	})

	t.Run("fixed_template_renders", func(t *testing.T) {
		// This is the FIXED version with proper escaping
		fixedTemplate := `name: LLM Review Panel
on:
  pull_request:
    types: [opened]

permissions:
  contents: read
  pull-requests: write

jobs:
  panel:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Post LLM Panel Comment
        uses: actions/github-script@v7
        with:
          github-token: ${{ "{{" }} secrets.GITHUB_TOKEN }}}

          script: |
            gh pr comment ${{ "{{" }} github.event.pull_request.number }} --body "## LLM Code Review Status


| LLM | Status |
|-----|--------|
| Claude | Pending (add /claude comment) |
| Codex | ${{ "{{" }} secrets.CODEX_AUTH_JSON != '' && '✓ Ready' || '⚠️ Token missing' {{ "}}" }} |
| Gemini | ${{ "{{" }} secrets.GEMINI_API_KEY != '' && '✓ Ready' || '⚠️ Token missing' {{ "}}" }} |
| GLM | ${{ "{{" }} secrets.GLM_API_KEY != '' && '✓ Ready' || '⚠️ Token missing' {{ "}}" }} |

Trigger individual reviews:
- Add /claude comment to trigger Claude
- Add /codex comment to trigger Codex
- Add /gemini comment to trigger Gemini
- Add /glm comment to trigger GLM
"
`

		fs := fstest.MapFS{
			"llm-panel.yml.tmpl": &fstest.MapFile{
				Data: []byte(fixedTemplate),
			},
		}

		r := NewRenderer(fs)

		// This should succeed with the fixed template
		result, err := r.Render("llm-panel.yml.tmpl", nil)
		if err != nil {
			t.Fatalf("expected no error with fixed template, got: %v", err)
		}

		// Verify the output contains properly escaped GitHub Actions expressions
		output := string(result)
		if !contains(output, "${{ secrets.GITHUB_TOKEN }}") {
			t.Error("output should contain escaped GitHub Actions token reference")
		}
		if !contains(output, "${{ secrets.CODEX_AUTH_JSON != '' && '✓ Ready' || '⚠️ Token missing' }}") {
			t.Error("output should contain properly escaped Codex conditional")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
