package cli

// AC-AFL-001: the ledger is a local store. Its import graph holds no Issue or
// remote client, and its default file is in the MoAI home, not the project.

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

const (
	afModulePath = "github.com/modu-ai/moai-adk" // verified with `go list -m`
	afLedgerPkg  = afModulePath + "/internal/agenticfactory/ledger"
)

var afForbiddenLedgerSuffixes = []string{"/forge", "/mirror", "/gitref", "/effect/ghintake", "/effect/ctxledger", "/effect/labels"}

// ledgerImportViolations judges a `go list -deps` output: the list must hold the
// package itself (an empty or unrelated list asserts nothing) and no forbidden path.
func ledgerImportViolations(deps []string, self string) error {
	var found bool
	var bad []string
	for _, d := range deps {
		d = strings.TrimSpace(d)
		if d == self {
			found = true
		}
		for _, suf := range afForbiddenLedgerSuffixes {
			if strings.HasSuffix(d, suf) {
				bad = append(bad, d)
			}
		}
	}
	if !found {
		return fmt.Errorf("the dependency list (%d entries) does not contain %s: the sweep is empty", len(deps), self)
	}
	if len(bad) > 0 {
		return fmt.Errorf("the ledger imports an Issue or remote client: %v", bad)
	}
	return nil
}

// ledgerPathProblem judges the default ledger path: exactly the keyed state-directory
// path, and not inside the project directory.
func ledgerPathProblem(got, want, project string) error {
	if got != want {
		return fmt.Errorf("default ledger path %s, want %s", got, want)
	}
	if rel, err := filepath.Rel(project, got); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("default ledger path %s is inside the project directory %s", got, project)
	}
	return nil
}

func TestLedgerImportGraph_NoIssueOrRemoteClient(t *testing.T) {
	// The checks must fail on a known-bad input before their verdict on the tree means anything.
	t.Run("fixture_forbidden_import_is_reported", func(t *testing.T) {
		deps := []string{"fmt", afLedgerPkg, afModulePath + "/internal/agenticfactory/effect/ghintake"}
		if err := ledgerImportViolations(deps, afLedgerPkg); err == nil || !strings.Contains(err.Error(), "ghintake") {
			t.Fatalf("a ghintake import must be reported, got %v", err)
		}
	})
	t.Run("fixture_empty_list_is_reported", func(t *testing.T) {
		if err := ledgerImportViolations(nil, afLedgerPkg); err == nil {
			t.Fatal("an empty dependency list must fail")
		}
	})
	t.Run("fixture_clean_list_passes", func(t *testing.T) {
		if err := ledgerImportViolations([]string{"fmt", afLedgerPkg}, afLedgerPkg); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("fixture_path_under_project_is_reported", func(t *testing.T) {
		project := filepath.Join(t.TempDir(), "p")
		in := filepath.Join(project, "af", "ledger.db")
		if err := ledgerPathProblem(in, in, project); err == nil || !strings.Contains(err.Error(), "inside the project") {
			t.Fatalf("a path under the project must be reported, got %v", err)
		}
	})

	// Assertion 1: the real import graph.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	out, err := exec.CommandContext(ctx, "go", "list", "-deps", afLedgerPkg).Output()
	if err != nil {
		t.Fatalf("go list -deps %s: %v", afLedgerPkg, err)
	}
	deps := strings.Fields(string(out))
	if err := ledgerImportViolations(deps, afLedgerPkg); err != nil {
		t.Error(err)
	}
	t.Logf("go list -deps swept %d packages", len(deps))

	// Assertion 2: the default ledger file, with the project directory pinned.
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), ".moai"))
	project := filepath.Join(t.TempDir(), "my-project")
	t.Setenv("CLAUDE_PROJECT_DIR", project)
	dir, err := homestate.ProjectDir(project)
	if err != nil {
		t.Fatal(err)
	}
	got, err := afLedgerPath("")
	if err != nil {
		t.Fatal(err)
	}
	if err := ledgerPathProblem(got, filepath.Join(dir, "af", "ledger.db"), project); err != nil {
		t.Error(err)
	}
}
