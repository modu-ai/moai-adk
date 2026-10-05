package web

// worktree_base_branch_guard_test.go — SPEC-WORKTREE-BASEREF-001 M5 (card t313).
//
// The anti-dead-key regression guard, modelled on dead_config_guard_test.go.
// AC-WBR-010 asserts presence; AC-WBR-012 asserts the guard's EXISTENCE and its
// three-part conjunction — schema, render, AND a live consumer. A grep showing
// the key's name in source text does not satisfy any of them.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/settings"
)

const worktreeBaseBranchKey = "git_strategy.worktree_base_branch"

// TestWorktreeBaseBranchLiveKeyGuard is the three-part conjunction. All three
// properties are asserted here, in one test, so removing the field breaks the
// guard rather than quietly shrinking the console.
func TestWorktreeBaseBranchLiveKeyGuard(t *testing.T) {
	// (1) present in settings.AllFields()
	var found bool
	for _, f := range settings.AllFields() {
		if f.Name == worktreeBaseBranchKey {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("(1) %q is absent from settings.AllFields(); the generic form parser can never collect it as an edit target", worktreeBaseBranchKey)
	}

	// (2) a form control carrying the name reaches the rendered HTML
	html := renderConsolePage(t)
	if !strings.Contains(html, `name="`+worktreeBaseBranchKey+`"`) {
		t.Errorf("(2) the rendered console carries no form control named %q", worktreeBaseBranchKey)
	}

	// (3) the key reaches a consumer: the git write seam receives the value.
	//
	// This is the property a grep cannot establish and the one that makes the
	// key live rather than merely stored. It is asserted end to end — a config
	// file carrying the value, through the real config reader, out to the seam
	// that would run `git remote set-head`.
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "git_strategy:\n    mode: manual\n    worktree_base_branch: develop\n"
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	var setHeadArgs []string
	restore := swapAlignmentSeamsForGuard(t, &setHeadArgs)
	defer restore()

	hook.RunWorktreeBaseAlignment(root)

	if len(setHeadArgs) != 1 || setHeadArgs[0] != "develop" {
		t.Errorf("(3) the git write seam received %v, want exactly one call with \"develop\" — the stored key does not reach a consumer", setHeadArgs)
	}
}

// swapAlignmentSeamsForGuard drives the alignment step deterministically: the
// primary-checkout gate reports true, origin/HEAD names a different branch, and
// the configured value resolves. The write seam records what it was handed.
func swapAlignmentSeamsForGuard(t *testing.T, setHeadArgs *[]string) func() {
	t.Helper()
	origPrimary := hook.WorktreeBaseBranchInPrimaryCheckout
	origResolvable := hook.WorktreeBaseBranchResolvable
	origSetHead := hook.WorktreeBaseBranchSetHead
	origOriginHead := hook.WorktreeBaseBranchOriginHead

	hook.WorktreeBaseBranchInPrimaryCheckout = func() bool { return true }
	hook.WorktreeBaseBranchResolvable = func(string) bool { return true }
	hook.WorktreeBaseBranchSetHead = func(branch string) error {
		*setHeadArgs = append(*setHeadArgs, branch)
		return nil
	}
	hook.WorktreeBaseBranchOriginHead = func() (string, error) { return "main", nil }

	return func() {
		hook.WorktreeBaseBranchInPrimaryCheckout = origPrimary
		hook.WorktreeBaseBranchResolvable = origResolvable
		hook.WorktreeBaseBranchSetHead = origSetHead
		hook.WorktreeBaseBranchOriginHead = origOriginHead
	}
}

// TestWorktreeBaseBranchRendersFreeTextInput is AC-WBR-011's render half.
//
// The attribute order asserted here is the order the renderer ACTUALLY emits,
// measured in this tree. The criterion was written with a two-branch
// `type="text" name=` / `name= type="text"` condition because that order was
// not measured at plan time; the measured form matches NEITHER branch — the
// renderer emits `class="in" type="text" id="<name>" name="<name>"`, with `id`
// between the two attributes. The condition is collapsed to the single measured
// form rather than shipped as an either-or.
func TestWorktreeBaseBranchRendersFreeTextInput(t *testing.T) {
	html := renderConsolePage(t)
	want := `<input class="in" type="text" id="` + worktreeBaseBranchKey + `" name="` + worktreeBaseBranchKey + `"`
	if !strings.Contains(html, want) {
		t.Errorf("the console does not render a text input for %q; expected the emitted form:\n%s", worktreeBaseBranchKey, want)
	}
	// A closed option set would render a <select> or a radio group carrying the
	// same name. Neither is permitted (REQ-WBR-014).
	if strings.Contains(html, `<select id="`+worktreeBaseBranchKey+`"`) ||
		strings.Contains(html, `type="radio" name="`+worktreeBaseBranchKey+`"`) {
		t.Errorf("%q rendered as a closed option set; it must be free text", worktreeBaseBranchKey)
	}
}
