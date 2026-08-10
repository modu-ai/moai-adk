package delegationmap

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestSubagentBoundary_NoAskUserQuestion mirrors the routing package's guard.
// No source file under internal/harness/delegationmap/ may contain an
// executable call-site invocation of AskUserQuestion or mcp__askuser: this
// analyzer runs in subagent/hook context, and the orchestrator owns every
// user interaction — including the Tier-4 approval gate this package's
// proposals feed.
//
// Detection targets call-site signatures on non-comment lines so documentation
// prose naming the boundary does not false-positive.
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
					violations = append(violations, path+":"+strconv.Itoa(lineNum+1)+" "+needle)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if len(violations) > 0 {
		t.Errorf("subagent boundary VIOLATED in delegationmap package:\n%s", strings.Join(violations, "\n"))
	}
}

// TestNoMapWritePath is the map-write half of AC-HLA-016. The delegation map's
// path literal must appear in exactly one file, and no write primitive may
// target the map at all.
//
// This is a source-level check rather than a behavioral one deliberately: a
// behavioral test can only prove the map was not written on the paths it
// happens to exercise, while the absence of a write primitive proves it for
// every path.
func TestNoMapWritePath(t *testing.T) {
	t.Parallel()

	writePrimitives := []string{"os.WriteFile(", "os.Create(", "os.OpenFile(", "yaml.Marshal("}
	var pathLiteralFiles []string

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
		body := string(data)
		if strings.Contains(body, "delegation.yaml") {
			pathLiteralFiles = append(pathLiteralFiles, path)
		}
		for _, prim := range writePrimitives {
			if strings.Contains(body, prim) {
				t.Errorf("%s uses the write primitive %s; this package writes no files", path, prim)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	if len(pathLiteralFiles) != 1 || !strings.HasSuffix(pathLiteralFiles[0], "mapreader.go") {
		t.Errorf("the delegation.yaml literal must appear in mapreader.go alone; found in %v", pathLiteralFiles)
	}
}

// TestNoIndependentLedgerPathLiteral is the reader half of AC-HLA-002: the
// ledger file name comes from the producer package, never from a literal
// declared here that would be free to drift from it.
func TestNoIndependentLedgerPathLiteral(t *testing.T) {
	t.Parallel()

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
		if strings.Contains(string(data), "routing-ledger.jsonl") {
			t.Errorf("%s declares its own ledger path literal instead of using routing.LedgerFileName", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}
