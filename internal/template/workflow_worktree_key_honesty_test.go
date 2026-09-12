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
	"regexp"
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

// keySection matches the start of a per-key comment clause, `# auto_cleanup:`.
var keySection = regexp.MustCompile(`^#\s*([a-z_]+):`)

// goFile matches a source path named inside a comment.
var goFile = regexp.MustCompile(`[\w./-]+\.go`)

// commentSections groups the `#` comment lines of a YAML document into the
// clause each one belongs to: a clause starts at a `# <key>:` line and runs
// until the next such line or the end of the comment block. The claim about a
// key spans more than the line that names it — `auto_cleanup`'s names its two
// reader files on the lines after — so a per-line reading would miss half of
// what the file asserts.
func commentSections(content string) map[string][]string {
	sections := map[string][]string{}
	current := ""
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") {
			current = ""
			continue
		}
		if match := keySection.FindStringSubmatch(trimmed); match != nil {
			current = match[1]
		}
		if current != "" {
			sections[current] = append(sections[current], trimmed)
		}
	}
	return sections
}

// namedGoFiles returns the base names of the source files a clause names.
func namedGoFiles(lines []string) map[string]bool {
	named := map[string]bool{}
	for _, line := range lines {
		for _, match := range goFile.FindAllString(line, -1) {
			named[filepath.Base(match)] = true
		}
	}
	return named
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
	sections := commentSections(string(data))

	for key, field := range worktreeToggles {
		readers := fieldReaders(t, field)
		clause, mentioned := sections[key]

		claimedUnread := false
		for _, line := range clause {
			lower := strings.ToLower(line)
			for _, claim := range unreadClaims {
				if strings.Contains(lower, claim) {
					claimedUnread = true
				}
			}
		}

		if len(readers) == 0 {
			if !claimedUnread {
				t.Errorf(
					"%s: no production line reads it, and the shipped %s does not say so. "+
						"An inert key that reads as live invites the user to set it and wait",
					key, workflowTemplatePath,
				)
			}
			continue
		}

		if !mentioned {
			t.Errorf(
				"%s: %d production line(s) read it (%v) and the shipped %s has no clause for the key. "+
					"A toggle that acts must say so where the user sets it",
				key, len(readers), readers, workflowTemplatePath,
			)
			continue
		}
		if claimedUnread {
			t.Errorf(
				"%s: the shipped %s calls it unread (%q), but %d production line(s) read it: %v. "+
					"auto_cleanup gates a `git worktree remove`, so a wrong claim here loses work",
				key, workflowTemplatePath, clause, len(readers), readers,
			)
			continue
		}

		// Every reading file has to be named, and every named file has to
		// read: the clause for auto_cleanup names two paths, and losing
		// either one would leave the file describing a reader that is gone
		// while the remaining one keeps the claim technically alive.
		readingFiles := map[string]bool{}
		for _, reader := range readers {
			readingFiles[filepath.Base(strings.SplitN(reader, ":", 2)[0])] = true
		}
		named := namedGoFiles(clause)
		for file := range readingFiles {
			if !named[file] {
				t.Errorf(
					"%s: %s reads the field but the shipped %s clause does not name it (%q). "+
						"The reader list in the comment is what the user checks the key against",
					key, file, workflowTemplatePath, clause,
				)
			}
		}
		for file := range named {
			if !readingFiles[file] {
				t.Errorf(
					"%s: the shipped %s names %s as a reader, but nothing there reads the field. "+
						"Readers found: %v",
					key, workflowTemplatePath, file, readers,
				)
			}
		}
	}
}
