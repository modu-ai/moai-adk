package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAC003_PerEditSpawnsEliminated (A8 / AC-STOPCHAIN-TRIM-003):
// the per-edit PreToolUse `develop-pre-implementation` and PostToolUse
// `develop-post-implementation` hooks MUST be removed from manager-develop.md
// frontmatter (the per-edit double-spawn eliminated). The bulk verification
// work moves to the per-cycle `develop-completion` Stop hook (already
// registered), and the narrow destructive-pattern + scope-discipline check
// remains covered by the GLOBAL handle-pre-tool.sh PreToolUse guard
// (settings.json, fires on Write|Edit|Bash) — NOT by an agent-local per-edit
// hook. So the agent-local per-edit spawn count across an N-edit TDD cycle is
// 0, satisfying AC-003.
//
// This test parses manager-develop.md frontmatter directly: it asserts that
// neither `develop-pre-implementation` nor `develop-post-implementation`
// appears as an agent-hook action in the frontmatter, while `develop-completion`
// (the per-cycle Stop hook) remains registered.
func TestAC003_PerEditSpawnsEliminated(t *testing.T) {
	agentFile := filepath.Join(mustGetwd(t), "..", "..", ".claude", "agents", "moai", "manager-develop.md")
	body, err := os.ReadFile(agentFile)
	if err != nil {
		t.Fatalf("read manager-develop.md: %v", err)
	}
	fm := extractFrontmatter(t, string(body))

	// Per-edit develop hooks MUST be absent (count 0 → AC-003).
	if n := strings.Count(fm, "develop-pre-implementation"); n != 0 {
		t.Fatalf("AC-003 REGRESSION: develop-pre-implementation appears %d time(s) in manager-develop.md frontmatter; per-edit PreToolUse spawn must be eliminated (count 0)", n)
	}
	if n := strings.Count(fm, "develop-post-implementation"); n != 0 {
		t.Fatalf("AC-003 REGRESSION: develop-post-implementation appears %d time(s) in manager-develop.md frontmatter; per-edit PostToolUse spawn must be eliminated (count 0)", n)
	}

	// The per-cycle develop-completion Stop hook MUST remain registered
	// (the bulk verification workload moved HERE per A8 / REQ-002).
	if !strings.Contains(fm, "develop-completion") {
		t.Fatalf("AC-003: develop-completion per-cycle Stop hook missing from manager-develop.md frontmatter (A8 moves the bulk work here, not removes it)")
	}
}

// extractFrontmatter returns the substring between the first pair of '---'
// delimiters (the YAML frontmatter block). Returns the whole body if no
// delimiters are found (defensive — the test then fails on the content checks).
func extractFrontmatter(t *testing.T, body string) string {
	t.Helper()
	first := strings.Index(body, "\n---\n")
	if first < 0 {
		// Maybe the file starts with ---
		if strings.HasPrefix(body, "---\n") {
			first = -1
		} else {
			return body
		}
	}
	start := first + 4
	if strings.HasPrefix(body, "---\n") {
		start = 4
	}
	rest := body[start:]
	second := strings.Index(rest, "\n---\n")
	if second < 0 {
		return body
	}
	return body[:start+second]
}
