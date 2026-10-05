package config

// workflow_key_honesty_test.go — card t682 (GH #1705): the shipped template
// must never describe a workflow.worktree.* key as unread/reserved while a
// production reader consumes it. GH #1705 was exactly that drift: the file
// users receive called auto_cleanup a "reserved" key while two auto-cleanup
// paths (session-exit disposal and the PR-merge sweep) gated worktree
// removal on it. The prose was repaired on develop; this guard makes its
// return RED.

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// workflowKeyClaims maps each workflow.worktree key to the comment phrases
// that mark it as unread/reserved in the shipped file. A key carrying at
// least one production reader must not have any of these phrases on a line
// that names it.
var workflowKeyClaims = map[string][]string{
	"auto_cleanup":         {"declared but not read", "no code path consumes", "(reserved)"},
	"auto_merge":           {"declared but not read", "no code path consumes", "(reserved)"},
	"session_name_pattern": {"declared but not read", "no code path consumes", "(reserved)"},
}

// workflowKeyAccessors maps each key to the Go field-access shape a
// production reader exhibits. The accessor form (not the bare field name)
// is what separates a real consumer from a comment or a struct declaration
// that merely mentions the field.
var workflowKeyAccessors = map[string]string{
	"auto_cleanup":         ".Workflow.Worktree.AutoCleanup",
	"auto_merge":           ".Workflow.Worktree.AutoMerge",
	"session_name_pattern": ".Workflow.Worktree.SessionNamePattern",
}

// countKeyReaders walks the internal/ source tree and counts, per key, the
// non-test Go files that access the field through a config receiver. The
// walk root is the package's parent (internal/), so the guard sees every
// production package, including ones added later.
func countKeyReaders(t *testing.T) map[string]int {
	t.Helper()

	readers := map[string]int{}
	root := ".."
	rootClean := filepath.Clean(root)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// The walk root itself may be named ".." (or "."), which the
			// hidden-dir rule below would skip wholesale — guard against
			// skipping the very tree being walked.
			if path != rootClean {
				name := d.Name()
				if name == "testdata" || strings.HasPrefix(name, ".") {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		src := string(data)
		for key, accessor := range workflowKeyAccessors {
			if strings.Contains(src, accessor) {
				readers[key]++
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return readers
}

// staleClaimFindings returns one finding per template line that names a key
// with production readers while calling the key unread/reserved. A key with
// zero readers is exempt: calling it reserved there is honest.
func staleClaimFindings(tmpl string, readers map[string]int) []string {
	var findings []string
	for i, line := range strings.Split(tmpl, "\n") {
		for key, phrases := range workflowKeyClaims {
			if readers[key] == 0 || !strings.Contains(line, key) {
				continue
			}
			for _, phrase := range phrases {
				if strings.Contains(line, phrase) {
					findings = append(findings, fmt.Sprintf(
						"shipped workflow.yaml line %d claims %s is unread/reserved (%q) but %d production reader(s) consume it — GH #1705 drift",
						i+1, key, phrase, readers[key]))
				}
			}
		}
	}
	return findings
}

// shippedWorkflowTemplate reads the template-source workflow.yaml the
// binary embeds — the file users actually receive — not the local dogfood
// copy.
func shippedWorkflowTemplate(t *testing.T) string {
	t.Helper()
	path := filepath.Join("..", "template", "templates", ".moai", "config", "sections", "workflow.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read shipped template: %v", err)
	}
	return string(data)
}

// TestWorkflowWorktreeKeyDocsMatchReaders is the t682 guard. The premise
// assertions make the guard non-vacuous: if the auto_cleanup readers GH
// #1705 names ever disappear, the guard fails loudly here instead of
// silently passing an empty check forever.
func TestWorkflowWorktreeKeyDocsMatchReaders(t *testing.T) {
	readers := countKeyReaders(t)

	// Premise: the two readers the issue names still exist. If this fails,
	// the key's reader status moved and the claims map needs re-judging.
	if readers["auto_cleanup"] < 1 {
		t.Fatalf("premise lost: no production reader of Workflow.Worktree.AutoCleanup — the two auto-cleanup paths (session-exit disposal, PR-merge sweep) are gone or renamed; re-judge workflowKeyClaims")
	}

	findings := staleClaimFindings(shippedWorkflowTemplate(t), readers)
	for _, f := range findings {
		t.Error(f)
	}
}

// TestWorkflowWorktreeKeyDocGuardDetectsHistoricalDrift is the guard's
// negative control: the exact comment shape GH #1705 reported must still be
// flagged. A guard that passes on the historical defect text guards nothing.
func TestWorkflowWorktreeKeyDocGuardDetectsHistoricalDrift(t *testing.T) {
	historical := `    # three auto-* toggles default false — the EnterWorktree-first policy.
    # auto_create: read once (internal/cli/worktree_advisory.go) only to select
    # advisory wording; it does not gate worktree creation.
    # auto_merge / auto_cleanup: declared but not read — no code path consumes
    # them (reserved). Kept aligned with defaults.go so the shipped file does
    # not state the opposite of the recorded policy.
    worktree:
        auto_cleanup: false
`
	findings := staleClaimFindings(historical, map[string]int{
		"auto_cleanup":         2,
		"auto_merge":           0,
		"session_name_pattern": 0,
	})
	if len(findings) == 0 {
		t.Fatal("guard must flag the historical GH #1705 comment, found nothing")
	}
	named := false
	for _, f := range findings {
		if strings.Contains(f, "auto_cleanup") {
			named = true
		}
	}
	if !named {
		t.Fatalf("findings must name auto_cleanup, got: %v", findings)
	}
}
