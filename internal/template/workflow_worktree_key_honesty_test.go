package template

// workflow_worktree_key_honesty_test.go — SPEC-CONFIG-KEY-HONESTY-001, for the
// worktree toggles as the SHIPPED file states them (#1705).
//
// The template's comment told the user that `auto_cleanup` was declared but
// not read, grouped into one sentence with `auto_merge`, which really has no
// reader. `auto_cleanup` gates both auto-cleanup paths, and each of them runs
// a `git worktree remove` immediately after reading it. A user who believed
// the file would either flip the key while reasoning about something else, or
// rule it out after an unexplained deletion; both readings lose work.
//
// The wording is not what is pinned here — the claim is. For each toggle, what
// the shipped comment says about being read has to match whether the Go tree
// reads the field, in both directions.

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const workflowTemplatePath = ".moai/config/sections/workflow.yaml"

// worktreeToggles maps each `workflow.worktree.*` key to its struct field on
// WorkflowWorktreeConfig (internal/config/types.go).
var worktreeToggles = map[string]string{
	"auto_create":          "AutoCreate",
	"auto_merge":           "AutoMerge",
	"auto_cleanup":         "AutoCleanup",
	"session_name_pattern": "SessionNamePattern",
}

// unreadClaims are the phrasings the template uses to call a key inert.
var unreadClaims = []string{"not read", "reserved"}

// commentLines returns the `#` comment lines of a YAML document.
func commentLines(content string) []string {
	var out []string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			out = append(out, line)
		}
	}
	return out
}

// fieldReaders returns the `internal/` source lines that READ
// WorkflowWorktreeConfig.<field>. A line that assigns to the field is not a
// reader, and neither is a comment — the point of the audit is code that acts
// on the value.
func fieldReaders(t *testing.T, field string) []string {
	t.Helper()

	needle := ".Worktree." + field
	var readers []string
	// The test's own package directory is internal/template, so `..` is
	// internal/ and the walk covers every production package under it.
	root := filepath.Join("..")
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for i, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if !strings.Contains(trimmed, needle) || strings.HasPrefix(trimmed, "//") {
				continue
			}
			// `x.Worktree.Field = v` writes the value; `== ` compares it.
			if idx := strings.Index(trimmed, needle); idx >= 0 {
				rest := strings.TrimSpace(trimmed[idx+len(needle):])
				if strings.HasPrefix(rest, "=") && !strings.HasPrefix(rest, "==") {
					continue
				}
			}
			readers = append(readers, filepath.ToSlash(path)+":"+strconv.Itoa(i+1))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	return readers
}

func TestWorkflowTemplate_WorktreeToggleReaderClaimsAreTrue(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	data, err := fs.ReadFile(fsys, workflowTemplatePath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error: %v", workflowTemplatePath, err)
	}
	comments := commentLines(string(data))

	for key, field := range worktreeToggles {
		readers := fieldReaders(t, field)

		var claimed []string
		var mentioned bool
		for _, line := range comments {
			if !strings.Contains(line, key) {
				continue
			}
			mentioned = true
			lower := strings.ToLower(line)
			for _, claim := range unreadClaims {
				if strings.Contains(lower, claim) {
					claimed = append(claimed, strings.TrimSpace(line))
				}
			}
		}

		switch {
		case len(readers) > 0 && len(claimed) > 0:
			t.Errorf(
				"%s: the shipped %s calls %q unread (%q), but %d production line(s) read it: %v. "+
					"auto_cleanup gates a `git worktree remove`, so a wrong claim here loses work",
				key, workflowTemplatePath, key, claimed, len(readers), readers,
			)
		case len(readers) > 0 && !mentioned:
			t.Errorf(
				"%s: %d production line(s) read it (%v) and the shipped %s never mentions the key. "+
					"A toggle that acts must say so where the user sets it",
				key, len(readers), readers, workflowTemplatePath,
			)
		case len(readers) == 0 && len(claimed) == 0:
			t.Errorf(
				"%s: no production line reads it, and the shipped %s does not say so. "+
					"An inert key that reads as live invites the user to set it and wait",
				key, workflowTemplatePath,
			)
		}
	}
}
