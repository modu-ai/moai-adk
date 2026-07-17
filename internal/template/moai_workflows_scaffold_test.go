package template

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// TestMoaiWorkflowsScaffoldPresence verifies that the embedded template tree
// ships the MoAI-Workflow scaffold under .moai/workflows/: a README.md that
// explains the workflow file format plus exactly one neutral example workflow
// file. This is the template-deployment presence guard for the user-facing
// simple markdown-workflow layer (workflows saved as .moai/workflows/<name>.md
// and registered on the Claude Code native scheduler).
//
// AC-MWS-018: the scaffold contains a README.md explaining the format and
// exactly one neutral example workflow file.
func TestMoaiWorkflowsScaffoldPresence(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	entries, err := fs.ReadDir(fsys, ".moai/workflows")
	if err != nil {
		t.Fatalf("scaffold directory .moai/workflows missing: %v", err)
	}

	var hasReadme bool
	var exampleFiles []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		switch {
		case name == "README.md":
			hasReadme = true
		case strings.HasSuffix(name, ".md"):
			exampleFiles = append(exampleFiles, name)
		}
	}

	if !hasReadme {
		t.Error("scaffold missing README.md explaining the workflow file format")
	}
	if len(exampleFiles) != 1 {
		t.Errorf("scaffold must contain exactly one neutral example workflow file, found %d: %v", len(exampleFiles), exampleFiles)
	}

	// README must actually explain the format (mention the frontmatter fields).
	if hasReadme {
		data, readErr := fs.ReadFile(fsys, ".moai/workflows/README.md")
		if readErr != nil {
			t.Fatalf("ReadFile README.md error: %v", readErr)
		}
		content := string(data)
		for _, token := range []string{"schedule", "mechanism", "safety"} {
			if !strings.Contains(content, token) {
				t.Errorf("README.md does not explain the %q frontmatter field", token)
			}
		}
	}

	// The example must be a valid workflow shape: frontmatter declaring the
	// required fields plus a non-empty body.
	if len(exampleFiles) == 1 {
		data, readErr := fs.ReadFile(fsys, ".moai/workflows/"+exampleFiles[0])
		if readErr != nil {
			t.Fatalf("ReadFile example error: %v", readErr)
		}
		fm, body, parseErr := parseFrontmatterAndBody(string(data))
		if parseErr != "" {
			t.Errorf("example %s frontmatter parse error: %s", exampleFiles[0], parseErr)
		}
		for _, field := range []string{"name", "description", "schedule", "safety"} {
			if _, ok := fm[field]; !ok {
				t.Errorf("example %s missing required frontmatter field %q", exampleFiles[0], field)
			}
		}
		if strings.TrimSpace(body) == "" {
			t.Errorf("example %s has an empty body (a workflow must carry step instructions)", exampleFiles[0])
		}
	}
}

// TestMoaiWorkflowsScaffoldNeutrality is a focused belt-and-suspenders check
// that the scaffold carries no internal SPEC IDs, internal dates, or commit
// SHAs, so the 16-language template distribution stays neutral. The whole-tree
// internal_content_leak_test is the primary guard; this test scopes the same
// intent to the scaffold directory.
//
// AC-MWS-019: the scaffold contains no internal SPEC IDs, no internal dates,
// and no commit SHAs.
func TestMoaiWorkflowsScaffoldNeutrality(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// Internal SPEC-ID series (this project), ISO date, and 7+ hex commit SHA.
	specID := regexp.MustCompile(`\bSPEC-[A-Z][A-Z0-9]+-[A-Z0-9-]*[0-9]{3}\b`)
	isoDate := regexp.MustCompile(`\b20[0-9]{2}-[01][0-9]-[0-3][0-9]\b`)
	commitSHA := regexp.MustCompile(`\b[0-9a-f]{7,40}\b`)

	walkErr := fs.WalkDir(fsys, ".moai/workflows", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		data, readErr := fs.ReadFile(fsys, path)
		if readErr != nil {
			t.Errorf("ReadFile(%q) error: %v", path, readErr)
			return nil
		}
		content := string(data)
		if m := specID.FindString(content); m != "" {
			t.Errorf("%s leaks an internal SPEC ID: %q", path, m)
		}
		if m := isoDate.FindString(content); m != "" {
			t.Errorf("%s leaks an internal date: %q", path, m)
		}
		if m := commitSHA.FindString(content); m != "" {
			t.Errorf("%s leaks a commit SHA: %q", path, m)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("WalkDir error: %v", walkErr)
	}
}
