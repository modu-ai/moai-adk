package template

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// SeqThinkingSentinel is the failure marker emitted when sequential-thinking
// literals reappear in the MoAI-ADK project namespace.
//
// Sentinel: SEQ_THINKING_REINTRODUCED
// Origin: SPEC-V3R6-SEQ-THINKING-RETIRE-001 (REQ-STR-004 / REQ-STR-005)
// Allow-list: internal/template/seq_thinking_retire_audit_allowlist.txt
//
// This test guards against accidental reintroduction of the
// `sequential-thinking` MCP server reference into agent frontmatter,
// skill bodies, rule documents, root configuration files, and root
// documentation. The only permitted occurrence is the Backward
// Compatibility marker line in moai-foundation-thinking/SKILL.md
// (registered in the allow-list file).
const SeqThinkingSentinel = "SEQ_THINKING_REINTRODUCED"

// seqThinkingPattern matches both the hyphenated and concatenated forms
// of sequential-thinking. The pattern is case-sensitive on the hyphenated
// form to avoid false-positives from the title-cased BC marker phrase
// "Sequential Thinking MCP support retired" — the BC marker uses
// title-case + no hyphen and therefore does not match this regex.
//
// Tokens detected:
//   - sequential-thinking         (.mcp.json server key, settings allow patterns)
//   - sequentialthinking          (frontmatter tool token concatenation)
//   - mcp__sequential-thinking__* (permission glob)
//
// Tokens NOT detected (intentional — these are BC marker phrasing):
//   - "Sequential Thinking" (title case, no hyphen)
//   - "Sequential-Thinking" (title case, hyphen)
var seqThinkingPattern = regexp.MustCompile(`sequential-thinking|sequentialthinking`)

// scanRoots lists the relative roots to walk for the sequential-thinking
// literal search. These mirror REQ-STR-001 scope and exclude
// `.claude/worktrees/` (ephemeral agent sandboxes containing snapshot copies).
var scanRoots = []string{
	".claude/agents",
	".claude/skills",
	".claude/rules",
	"internal/template/templates/.claude/agents",
	"internal/template/templates/.claude/skills",
	"internal/template/templates/.claude/rules",
}

// scanFiles lists individual non-source files to scan: configuration files
// and root documentation.
var scanFiles = []string{
	".mcp.json",
	".claude/settings.json",
	"internal/template/templates/.mcp.json.tmpl",
	"internal/template/templates/.claude/settings.json.tmpl",
	"CLAUDE.md",
}

// projectRoot resolves to the repository root regardless of where the
// test binary is invoked from. Tests under internal/template/ run with
// cwd set to that package directory; we walk two levels up.
func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error: %v", err)
	}
	// internal/template → repo root
	root := filepath.Clean(filepath.Join(wd, "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("project root resolution failed: %s has no go.mod (cwd=%s)", root, wd)
	}
	return root
}

// loadAllowlist reads the audit allow-list file and returns the set of
// permitted substrings. Lines beginning with `#` are treated as comments;
// empty lines are skipped.
func loadAllowlist(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("open allow-list %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()

	var entries []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entries = append(entries, line)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("read allow-list %s: %v", path, err)
	}
	return entries
}

// isAllowed returns true when the line text matches any allow-list entry
// via substring containment. Substring matching keeps the allow-list
// simple — each entry is a verbatim fragment of a permitted line.
func isAllowed(line string, allowlist []string) bool {
	for _, entry := range allowlist {
		if strings.Contains(line, entry) {
			return true
		}
	}
	return false
}

// scanFileForSeqThinking walks lines of a file and reports any line
// containing a sequential-thinking literal that is not whitelisted.
func scanFileForSeqThinking(t *testing.T, path string, allowlist []string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()

	var violations []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if !seqThinkingPattern.MatchString(line) {
			continue
		}
		if isAllowed(line, allowlist) {
			continue
		}
		violations = append(violations, formatViolation(path, lineNum, line))
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return violations
}

func formatViolation(path string, line int, text string) string {
	return path + ":" + seqThinkingItoa(line) + ": " + strings.TrimSpace(text)
}

func seqThinkingItoa(n int) string {
	// Minimal allocation-free integer-to-string for line numbers (1..N).
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// TestSeqThinkingRetired enforces REQ-STR-001 / REQ-STR-005 namespace
// cleanliness for sequential-thinking literals across the MoAI-ADK
// project surface. The only permitted occurrence is the BC marker line
// in moai-foundation-thinking/SKILL.md (registered in the allow-list).
//
// On failure, the test emits the SEQ_THINKING_REINTRODUCED sentinel
// (grep-discoverable per AC-STR-006) and lists every violating
// file:line so the contributor can remove the reintroduction.
func TestSeqThinkingRetired(t *testing.T) {
	t.Parallel()

	root := projectRoot(t)
	allowlistPath := filepath.Join(root, "internal", "template", "seq_thinking_retire_audit_allowlist.txt")
	allowlist := loadAllowlist(t, allowlistPath)

	var violations []string

	// 1) Walk source-tree roots.
	for _, rel := range scanRoots {
		base := filepath.Join(root, rel)
		walkErr := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}
			if d.IsDir() {
				return nil
			}
			// Markdown, YAML/YML, JSON, and JSON template files are the
			// only carriers for these literals in scope; skip binaries.
			ext := strings.ToLower(filepath.Ext(d.Name()))
			switch ext {
			case ".md", ".yaml", ".yml", ".json", ".tmpl", ".go":
				// Test source files in internal/template/ legitimately
				// reference the literals as test fixtures; the audit test
				// file itself contains the pattern in regex form.
				if strings.HasSuffix(path, "seq_thinking_retire_audit_test.go") {
					return nil
				}
				violations = append(violations, scanFileForSeqThinking(t, path, allowlist)...)
			}
			return nil
		})
		if walkErr != nil {
			t.Fatalf("walk %s: %v", base, walkErr)
		}
	}

	// 2) Scan individual configuration and documentation files.
	for _, rel := range scanFiles {
		violations = append(violations, scanFileForSeqThinking(t, filepath.Join(root, rel), allowlist)...)
	}

	if len(violations) > 0 {
		t.Errorf(
			"%s: %d sequential-thinking literal(s) reintroduced into MoAI-ADK namespace.\n"+
				"Retirement policy: SPEC-V3R6-SEQ-THINKING-RETIRE-001.\n"+
				"Allow-list (one entry per permitted line): %s\n"+
				"Violations:\n  - %s\n"+
				"Remediation: remove the literal, or — if the line is the\n"+
				"moai-foundation-thinking BC marker — extend the allow-list\n"+
				"file with the exact substring.",
			SeqThinkingSentinel,
			len(violations),
			allowlistPath,
			strings.Join(violations, "\n  - "),
		)
	}
}
