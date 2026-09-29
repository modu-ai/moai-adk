// todo_verdict_readjudication_test.go — SPEC-TODO-TRANSITION-STAMPS-001
// AC-TST-010: the stored verdict cites the axis-F ref the printed line named,
// and re-running the axis-F predicate against that recorded ref reproduces
// both the stored verdict and the delivering SHA for the seeded commit.
// Consistency of storage with the predicate, never a replacement: the record
// stores the answer's coordinates; the re-run supplies the SHA the record
// deliberately does not store.
package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// gitOutIn runs `git -C dir <args...>` and returns stdout, failing the test
// on a non-zero exit. Test-side only — the production path never shells git
// from the tests.
func gitOutIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return string(out)
}

// TestDoneVerdict_ReAdjudicationReDerivesSHA — AC-TST-010.
func TestDoneVerdict_ReAdjudicationReDerivesSHA(t *testing.T) {
	root, store := todoFixture(t)
	seedLandedRepo(t, root)
	ref := landedRefName(t, root)

	// The seeded control: exactly one commit on the landed ref carries an
	// attributing subject (the axis-F trailing-parenthetical form).
	seededSHA := strings.TrimSpace(gitOutIn(t, root, "log", ref, "--format=%H", "--grep=(t1)", "-n", "1"))
	if seededSHA == "" {
		t.Fatalf("the seeded control is broken: no attributing commit on %s", ref)
	}

	if _, _, err := runTodo(t, "add", "--pick", "readjudication control"); err != nil {
		t.Fatalf("add --pick: %v", err)
	}
	stdout, _, err := runTodo(t, "done", "t1", "--require-landed")
	if err != nil {
		t.Fatalf("done --require-landed: %v", err)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	v := rec.Archived[0].LandingVerdict
	if v == nil {
		t.Fatalf("landing_verdict is NULL, want the persisted record")
	}

	// The stored ref is the ref the printed line named.
	printedRef := ""
	for _, field := range strings.Fields(stdout) {
		if after, ok := strings.CutPrefix(field, "ref="); ok {
			printedRef = after
		}
	}
	if printedRef == "" || v.Ref != printedRef {
		t.Errorf("stored ref = %q, want the printed ref= suffix %q", v.Ref, printedRef)
	}

	// Re-run the predicate against the RECORDED ref: the verdict reproduces.
	q := kanban.GitLandedQuerier{Run: todoRunCommand, Ref: v.Ref}
	answer, err := q.Landed("t1")
	if err != nil {
		t.Fatalf("re-adjudication query failed: %v", err)
	}
	if answer != v.Verdict {
		t.Errorf("re-adjudication against recorded ref %s answered %q, stored verdict is %q", v.Ref, answer, v.Verdict)
	}

	// Re-derive the delivering SHA the record does not store: the attributing
	// commit on the recorded ref's subject stream.
	stream := gitOutIn(t, root, "log", v.Ref, "--format=%H%x09%s")
	derived := ""
	for _, line := range strings.Split(strings.TrimSpace(stream), "\n") {
		sha, subject, _ := strings.Cut(line, "\t")
		if strings.HasSuffix(subject, fmt.Sprintf("(%s)", "t1")) {
			derived = sha
			break
		}
	}
	if derived == "" {
		t.Fatalf("no attributing commit found on %s — the SHA is not re-derivable from the recorded ref", v.Ref)
	}
	if derived != seededSHA {
		t.Errorf("re-derived delivering SHA = %s, want the seeded commit %s", derived, seededSHA)
	}
}
