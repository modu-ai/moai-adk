package template

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestRootLevelCommandsThinPattern verifies that root-level slash command files
// in the project's .claude/commands/ directory (not in the embedded template tree)
// follow the thin command pattern: YAML frontmatter with required fields,
// a body of less than 20 non-empty lines, a Skill() invocation, and for any
// Skill("<name>") reference, the corresponding .claude/skills/<name>/ directory
// must exist (partial migration gate).
//
// Source: SPEC-V3R2-WF-002 REQ-WF002-001, REQ-WF002-002, REQ-WF002-015
func TestRootLevelCommandsThinPattern(t *testing.T) {
	t.Parallel()

	// Ascend two levels from this test file location (internal/template/) to the project root
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	projectRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")
	commandsDir := filepath.Join(projectRoot, ".claude", "commands")
	skillsDir := filepath.Join(projectRoot, ".claude", "skills")

	// Verify path existence
	if _, err := os.Stat(commandsDir); os.IsNotExist(err) {
		t.Fatalf("commands directory does not exist: %s", commandsDir)
	}

	fsys := os.DirFS(commandsDir)

	// Collect only first-level (.md, .md.tmpl) files — exclude subdirectories such as agency/
	var cmdFiles []string
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		// Do not descend into directories other than the root
		if d.IsDir() {
			if path != "." {
				return fs.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".md.tmpl") {
			cmdFiles = append(cmdFiles, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir error: %v", err)
	}

	if len(cmdFiles) == 0 {
		t.Fatal("no root-level command files found under .claude/commands/")
	}

	for _, path := range cmdFiles {
		path := path // capture loop variable
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) error: %v", path, readErr)
			}

			content := string(data)
			fm, body, parseErr := parseFrontmatterAndBody(content)
			if parseErr != "" {
				t.Fatalf("parse error: %s", parseErr)
			}

			// R1: verify required frontmatter fields are present
			if _, ok := fm["description"]; !ok {
				t.Error("missing required frontmatter field: description")
			}
			if _, ok := fm["allowed-tools"]; !ok {
				t.Error("missing required frontmatter field: allowed-tools")
			}

			// R2: allowed-tools must be a CSV string (YAML array not allowed)
			if at, ok := fm["allowed-tools"]; ok {
				if strings.HasPrefix(strings.TrimSpace(at), "-") {
					t.Error("allowed-tools must be CSV string, not YAML array")
				}
			}

			// R3: body non-empty line count < 20 (REQ-WF002-001, REQ-WF002-002)
			bodyLines := countNonEmptyLines(body)
			if bodyLines >= 20 {
				t.Errorf("body has %d non-empty lines (max 19 for thin commands)", bodyLines)
			}

			// R4: body must contain a Skill() invocation
			if !strings.Contains(body, "Skill(") {
				t.Errorf("body does not contain Skill() invocation")
			}

			// R5: partial migration gate (REQ-WF002-015)
			// When body references Skill("<name>"), the .claude/skills/<name>/ directory must exist
			checkSkillDirExists(t, body, skillsDir, path)
		})
	}

	t.Logf("audited %d root-level command files", len(cmdFiles))
}

// checkSkillDirExists extracts Skill("...") references from body and verifies
// that the corresponding .claude/skills/<name>/ directory exists.
// REQ-WF002-015 partial migration gate.
func checkSkillDirExists(t *testing.T, body, skillsDir, cmdPath string) {
	t.Helper()

	// Extract Skill("name") pattern occurrences from body
	remaining := body
	for {
		idx := strings.Index(remaining, `Skill("`)
		if idx < 0 {
			break
		}
		after := remaining[idx+len(`Skill("`):]
		end := strings.Index(after, `"`)
		if end < 0 {
			break
		}
		skillName := after[:end]
		if skillName != "" {
			skillPath := filepath.Join(skillsDir, skillName)
			skillPathMD := skillPath + ".md"
			_, errDir := os.Stat(skillPath)
			_, errMD := os.Stat(skillPathMD)
			if os.IsNotExist(errDir) && os.IsNotExist(errMD) {
				t.Errorf(
					"THIN_WRAPPER_PARTIAL_MIGRATION: %s references Skill(%q) but .claude/skills/%s/ does not exist (REQ-WF002-015)",
					cmdPath, skillName, skillName,
				)
			}
		}
		remaining = after[end+1:]
	}
}
