package routing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSubagentBoundary_NoAskUserQuestion is the C-HRA-008 binary CI guard for
// the routing package (B3/B11). No source file under internal/harness/routing/
// may contain an executable call-site invocation of AskUserQuestion or
// mcp__askuser — the routing capture path runs in subagent/hook context where
// the orchestrator owns all user interaction.
//
// Detection targets actual call-site signatures ("AskUserQuestion(",
// "mcp__askuser(") on non-comment lines, excluding test files, so documentation
// prose mentioning the AskUserQuestion boundary does not false-positive.
func TestSubagentBoundary_NoAskUserQuestion(t *testing.T) {
	t.Parallel()

	forbidden := []string{"AskUserQuestion(", "mcp__askuser("}
	var violations []string

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
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
		for lineNum, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}
			for _, needle := range forbidden {
				if strings.Contains(line, needle) {
					violations = append(violations, path+":"+itoa(lineNum+1)+" "+needle)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if len(violations) > 0 {
		t.Errorf("C-HRA-008 VIOLATED in routing package:\n%s", strings.Join(violations, "\n"))
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
