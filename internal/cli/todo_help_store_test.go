// todo_help_store_test.go — card t437. The `moai todo` help text is a store
// claim compiled into the binary, and the t395 completeness sweep
// (internal/template/backlog_json_disclosure_mirror_test.go) walks
// `internal/template/templates/` only, so it cannot see a Go string literal
// by construction. That sweep passing is therefore no evidence about this
// layer — the two stale claims it could not reach sat in this package's own
// help text, one matching each of its two patterns.
//
// This file closes that axis with the same two patterns, run over the Go and
// templ sources instead of the template tree.
package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// todoStoreClaimPatterns are the t395 patterns verbatim. Pattern 1 cannot see
// the home-fallback site and pattern 2 cannot see the primary one, which is
// why both are enumerated rather than merged into one alternation.
var todoStoreClaimPatterns = map[string]*regexp.Regexp{
	"primary (state/todo/backlog.json)":               regexp.MustCompile(`state/todo/backlog\.json`),
	"home-fallback (todo/<project-key>/backlog.json)": regexp.MustCompile(`moai/todo/<project-key>/backlog\.json`),
}

// todoStoreClaimRoots are the source trees compiled into the binary.
var todoStoreClaimRoots = []string{"internal", "cmd", "pkg"}

// TestTodoHelp_NamesTheDatabaseStore asserts the claim at its point of use:
// the string a reader actually sees from `moai todo --help`.
func TestTodoHelp_NamesTheDatabaseStore(t *testing.T) {
	long := newTodoCmd().Long
	if !strings.Contains(long, ".moai/state/todo/backlog.db") {
		t.Errorf("todo Long help does not name the canonical store .moai/state/todo/backlog.db")
	}
	for name, re := range todoStoreClaimPatterns {
		if re.MatchString(long) {
			t.Errorf("todo Long help still carries the stale %s claim — backlog.json is an export or legacy leftover, not the queue", name)
		}
	}
}

// TestTodoStoreClaims_NoStaleGoSourceSite is the Go-layer completeness sweep.
// Test files are excluded: they carry the patterns as checkers and fixtures,
// which is the correct place for them to appear.
func TestTodoStoreClaims_NoStaleGoSourceSite(t *testing.T) {
	repoRoot := filepath.Join("..", "..")
	hits := map[string][]string{}

	for _, root := range todoStoreClaimRoots {
		walkRoot := filepath.Join(repoRoot, root)
		if _, err := os.Stat(walkRoot); err != nil {
			t.Fatalf("stat %s: %v", walkRoot, err)
		}
		err := filepath.Walk(walkRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			base := filepath.Base(path)
			if strings.HasSuffix(base, "_test.go") {
				return nil
			}
			if ext := filepath.Ext(base); ext != ".go" && ext != ".templ" {
				return nil
			}
			// The template tree is the t395 sweep's subject, not this one's.
			if strings.Contains(filepath.ToSlash(path), "/internal/template/templates/") {
				return nil
			}
			raw, rerr := os.ReadFile(path)
			if rerr != nil {
				return rerr
			}
			for i, line := range strings.Split(string(raw), "\n") {
				for name, re := range todoStoreClaimPatterns {
					if re.MatchString(line) {
						loc := filepath.ToSlash(path) + ":" + strconv.Itoa(i+1)
						hits[name] = append(hits[name], loc+" — "+strings.TrimSpace(line))
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", walkRoot, err)
		}
	}

	for name, found := range hits {
		t.Errorf("pattern %s expected zero non-test source matches, got %d:\n%s",
			name, len(found), strings.Join(found, "\n"))
	}
}
