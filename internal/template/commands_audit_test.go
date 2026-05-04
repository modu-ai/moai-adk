package template

import (
	"bufio"
	"io/fs"
	"strings"
	"testing"
)

// TestCommandsThinPattern verifies that all slash command files in the embedded
// templates follow the thin command pattern: YAML frontmatter with required
// fields, and a body of less than 20 lines containing a Skill() invocation.
//
// Source: SPEC-THIN-CMDS-001
func TestCommandsThinPattern(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// Collect all command files (.md and .md.tmpl) under .claude/commands/
	var cmdFiles []string
	walkErr := fs.WalkDir(fsys, ".claude/commands", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".md.tmpl") {
			cmdFiles = append(cmdFiles, path)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("WalkDir error: %v", walkErr)
	}

	if len(cmdFiles) == 0 {
		t.Fatal("no command files found under .claude/commands/")
	}

	// Root command files (e.g., agency.md) are allowed to have longer bodies
	// because they contain router logic. Subcommands must be thin wrappers.
	rootCommands := map[string]bool{
		".claude/commands/agency/agency.md": true,
	}

	for _, path := range cmdFiles {
		t.Run(path, func(t *testing.T) {
			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) error: %v", path, readErr)
			}

			content := string(data)
			fm, body, parseErr := parseFrontmatterAndBody(content)
			if parseErr != "" {
				t.Fatalf("parse error: %s", parseErr)
			}

			// R1: Required frontmatter fields
			if _, ok := fm["description"]; !ok {
				t.Error("missing required frontmatter field: description")
			}
			if _, ok := fm["allowed-tools"]; !ok {
				t.Error("missing required frontmatter field: allowed-tools")
			}

			// R2: allowed-tools must be CSV string (not YAML array)
			if at, ok := fm["allowed-tools"]; ok {
				if strings.HasPrefix(strings.TrimSpace(at), "-") {
					t.Error("allowed-tools must be CSV string, not YAML array")
				}
			}

			// R3 and R4 only apply to subcommands (not root commands)
			if rootCommands[path] {
				return
			}

			// R3: Body LOC must be < 20 (thin wrapper)
			bodyLines := countNonEmptyLines(body)
			if bodyLines >= 20 {
				t.Errorf("body has %d non-empty lines (max 19 for thin commands)", bodyLines)
			}

			// R4: Body should contain Skill() invocation
			if !strings.Contains(body, "Skill(") {
				t.Errorf("body does not contain Skill() invocation")
			}
		})
	}

	t.Logf("audited %d command files", len(cmdFiles))
}

// TestBrainCommandThinPattern verifies that the moai/brain.md command file
// satisfies all Thin Command Pattern requirements for SPEC-V3R3-BRAIN-001.
//
// Source: SPEC-V3R3-BRAIN-001 deliverable #7 (T-A5.1)
func TestBrainCommandThinPattern(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	const brainCmdPath = ".claude/commands/moai/brain.md"

	data, readErr := fs.ReadFile(fsys, brainCmdPath)
	if readErr != nil {
		t.Fatalf("brain.md not found at %q: %v — did you run make build?", brainCmdPath, readErr)
	}

	content := string(data)
	fm, body, parseErr := parseFrontmatterAndBody(content)
	if parseErr != "" {
		t.Fatalf("parse error for %q: %s", brainCmdPath, parseErr)
	}

	// R1: description フィールド必須
	desc, ok := fm["description"]
	if !ok || strings.TrimSpace(desc) == "" {
		t.Error("brain.md: missing or empty 'description' frontmatter field")
	}

	// R2: argument-hint フィールド存在確認
	if _, ok := fm["argument-hint"]; !ok {
		t.Error("brain.md: missing 'argument-hint' frontmatter field")
	}

	// R3: allowed-tools フィールド必須 (CSV string)
	allowedTools, ok := fm["allowed-tools"]
	if !ok {
		t.Error("brain.md: missing 'allowed-tools' frontmatter field")
	} else if strings.HasPrefix(strings.TrimSpace(allowedTools), "-") {
		t.Error("brain.md: allowed-tools must be CSV string, not YAML array")
	}

	// R4: body LOC < 20 (Thin Command Pattern)
	bodyLines := countNonEmptyLines(body)
	if bodyLines >= 20 {
		t.Errorf("brain.md: body has %d non-empty lines (max 19 for thin commands)", bodyLines)
	}

	// R5: Skill() 호출 패턴 존재 확인
	if !strings.Contains(body, "Skill(") {
		t.Error("brain.md: body does not contain Skill() invocation")
	}

	// R6: brain 워크플로우로 라우팅하는지 확인
	if !strings.Contains(body, "brain") {
		t.Error("brain.md: body should reference 'brain' workflow routing")
	}

	t.Logf("brain.md: description=%q, bodyLines=%d, allowedTools=%q", desc, bodyLines, allowedTools)
}

// TestCommandsFrontmatterConsistency checks that all command files have
// consistent frontmatter: description field present, allowed-tools as CSV,
// and no deprecated fields.
func TestCommandsFrontmatterConsistency(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// Deprecated fields that should not appear in command frontmatter
	deprecatedFields := []string{"tools", "disallowed-tools", "model"}

	var cmdFiles []string
	_ = fs.WalkDir(fsys, ".claude/commands", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".md.tmpl") {
			cmdFiles = append(cmdFiles, path)
		}
		return nil
	})

	for _, path := range cmdFiles {
		t.Run(path, func(t *testing.T) {
			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				t.Fatalf("ReadFile error: %v", readErr)
			}

			fm, _, parseErr := parseFrontmatterAndBody(string(data))
			if parseErr != "" {
				t.Fatalf("parse error: %s", parseErr)
			}

			// Check for deprecated fields
			for _, field := range deprecatedFields {
				if _, ok := fm[field]; ok {
					t.Errorf("deprecated frontmatter field found: %s", field)
				}
			}

			// description must be non-empty
			if desc, ok := fm["description"]; ok {
				trimmed := strings.TrimSpace(desc)
				if trimmed == "" || trimmed == `""` {
					t.Error("description field is empty")
				}
			}
		})
	}
}

// parseFrontmatterAndBody splits a command file into frontmatter key-value
// pairs and the body text. Returns an error message string if parsing fails.
// This is a simplified parser that handles the common YAML frontmatter format
// used in Claude Code command files.
func parseFrontmatterAndBody(content string) (map[string]string, string, string) {
	fm := make(map[string]string)

	scanner := bufio.NewScanner(strings.NewReader(content))

	// Expect opening ---
	if !scanner.Scan() {
		return nil, "", "empty file"
	}
	if strings.TrimSpace(scanner.Text()) != "---" {
		return nil, "", "missing opening --- frontmatter delimiter"
	}

	// Read frontmatter lines until closing ---
	var fmLines []string
	foundClose := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			foundClose = true
			break
		}
		fmLines = append(fmLines, line)
	}
	if !foundClose {
		return nil, "", "missing closing --- frontmatter delimiter"
	}

	// Parse simple key: value pairs from frontmatter
	for _, line := range fmLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		idx := strings.Index(trimmed, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(trimmed[:idx])
		val := strings.TrimSpace(trimmed[idx+1:])
		// Strip surrounding quotes from value
		val = strings.Trim(val, `"'`)
		fm[key] = val
	}

	// Remaining content is body
	var bodyLines []string
	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}
	body := strings.Join(bodyLines, "\n")

	return fm, body, ""
}

// countNonEmptyLines counts the number of non-empty, non-whitespace-only lines.
func countNonEmptyLines(s string) int {
	count := 0
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}
	return count
}
