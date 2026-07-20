package update

import (
	"strings"
	"testing"
)

// These tests define the M3c preview contract (SPEC-CLI-TUX-V3-003):
//   - AC-TUX3-008: TUI table renders per-class counts + rows with keyboard nav
//   - AC-TUX3-009: selecting a row reaches the per-file diff viewport
//   - AC-TUX3-010: --yes / non-TTY falls back to plain text consuming the SAME
//     classification model, with ZERO ANSI escape sequences under NO_COLOR / piped
//   - AC-TUX3-014: BOTH the TUI table AND the text fallback label user-owned
//     entries `preserved (user-owned)`
//
// The preview MUST consume update.Classify (REQ-TUX3-001/002 single source of
// truth) — no parallel heuristic (plan.md §G anti-pattern).

// allFourClassesInputs is a fixture spanning every ChangeClass so count-based
// assertions exercise the full classification surface. The predicate marks
// only the `.claude/skills/hns-my/SKILL.md` path as user-owned (matches the
// isUserOwnedNamespace semantics pinned by TestNamespacePredicateSuperset).
func allFourClassesInputs() []FilePreviewInput {
	return []FilePreviewInput{
		{RelPath: "templates/new_file.yaml", Exists: false, Conflict: false, Diff: "+name: new"},
		{RelPath: ".claude/settings.json", Exists: true, Conflict: false, Diff: "-old\n+new"},
		{RelPath: ".claude/skills/hns-my/SKILL.md", Exists: true, Conflict: false, Diff: ""}, // user-owned
		{RelPath: ".moai/config/config.yaml", Exists: true, Conflict: true, Diff: "<<<<<\n====\n>>>>>"},
	}
}

func allFourClassesPredicate() UserOwnedPredicate {
	return func(relPath string) bool {
		return strings.HasPrefix(relPath, ".claude/skills/hns-")
	}
}

// countOf is a tiny assertion helper that counts non-overlapping occurrences.
func countOf(haystack, needle string) int {
	if needle == "" {
		return 0
	}
	return strings.Count(haystack, needle)
}

// -------------------------------------------------------------------
// AC-TUX3-008 — TUI table renders per-class counts + rows + keyboard nav
// -------------------------------------------------------------------

// TestPreviewTableRendersPerClassCounts asserts the TUI table view surfaces a
// per-class count summary for every ChangeClass (AC-TUX3-008). The classifier
// is the single source of truth, so the counts MUST reflect exactly the
// classification produced by update.Classify over the fixture.
func TestPreviewTableRendersPerClassCounts(t *testing.T) {
	in := allFourClassesInputs()
	model := newPreviewModel(in, allFourClassesPredicate(), PreviewOptions{Interactive: true, Width: 80, Height: 24})

	rendered := model.tableView()

	for _, class := range []ChangeClass{ClassAdd, ClassUpdate, ClassPreserveUserOwned, ClassConflict} {
		label := class.String()
		if !strings.Contains(rendered, label) {
			t.Errorf("tableView missing class label %q; rendered:\n%s", label, rendered)
		}
	}
	// Each class has exactly one fixture row, so the count summary must be
	// present. We assert the preserve class count line explicitly because the
	// user-owned visibility (AC-TUX3-014) depends on it.
	if !strings.Contains(rendered, "preserved (user-owned)") {
		t.Errorf("tableView missing 'preserved (user-owned)' label (AC-TUX3-014); rendered:\n%s", rendered)
	}
}

// TestPreviewTableRendersEveryFileRow asserts every input file appears as a
// row in the table view (AC-TUX3-008 — "per-class counts and rows").
func TestPreviewTableRendersEveryFileRow(t *testing.T) {
	in := allFourClassesInputs()
	model := newPreviewModel(in, allFourClassesPredicate(), PreviewOptions{Interactive: true, Width: 80, Height: 24})

	rendered := model.tableView()
	for _, f := range in {
		if !strings.Contains(rendered, f.RelPath) {
			t.Errorf("tableView missing file row for %q; rendered:\n%s", f.RelPath, rendered)
		}
	}
}

// TestPreviewTableClassificationMatchesClassify asserts the preview's
// classification is derived from update.Classify (REQ-TUX3-001/002 single
// source of truth — plan.md §G anti-pattern defense: no parallel heuristic).
func TestPreviewTableClassificationMatchesClassify(t *testing.T) {
	in := allFourClassesInputs()
	pred := allFourClassesPredicate()
	classes := classifyAll(in, pred)

	wantByPath := map[string]ChangeClass{
		"templates/new_file.yaml":            ClassAdd,
		".claude/settings.json":              ClassUpdate,
		".claude/skills/hns-my/SKILL.md":     ClassPreserveUserOwned,
		".moai/config/config.yaml":           ClassConflict,
	}
	for _, c := range classes {
		want, ok := wantByPath[c.RelPath]
		if !ok {
			t.Fatalf("unexpected path in classification: %q", c.RelPath)
		}
		if c.Class != want {
			t.Errorf("path %q classified as %q; want %q (Classify single source of truth)", c.RelPath, c.Class, want)
		}
	}
}

// -------------------------------------------------------------------
// AC-TUX3-009 — row select → per-file diff viewport reachable
// -------------------------------------------------------------------

// TestPreviewSelectRowReachesDiffViewport asserts that selecting a row in the
// table transitions the model to a diff-view state whose rendered content
// contains the selected file's diff (AC-TUX3-009).
func TestPreviewSelectRowReachesDiffViewport(t *testing.T) {
	in := allFourClassesInputs()
	model := newPreviewModel(in, allFourClassesPredicate(), PreviewOptions{Interactive: true, Width: 80, Height: 24})

	if model.currentView() != previewTableView {
		t.Fatalf("initial view = %v; want %v", model.currentView(), previewTableView)
	}

	// Select the first row (templates/new_file.yaml with diff "+name: new").
	model = model.selectRow()

	if model.currentView() != previewDiffView {
		t.Fatalf("after selectRow, view = %v; want %v", model.currentView(), previewDiffView)
	}
	diffRendered := model.diffView()
	want := in[0].Diff
	if want != "" && !strings.Contains(diffRendered, want) {
		t.Errorf("diffView missing diff content %q; rendered:\n%s", want, diffRendered)
	}
}

// TestPreviewDiffViewportEscReturnsToTable asserts the viewport → table
// transition is reversible (keyboard navigation contract — AC-TUX3-008/009).
func TestPreviewDiffViewportEscReturnsToTable(t *testing.T) {
	in := allFourClassesInputs()
	model := newPreviewModel(in, allFourClassesPredicate(), PreviewOptions{Interactive: true, Width: 80, Height: 24})

	model = model.selectRow()
	if model.currentView() != previewDiffView {
		t.Fatalf("selectRow did not enter diff view: %v", model.currentView())
	}
	model = model.backToTable()
	if model.currentView() != previewTableView {
		t.Errorf("backToTable did not return to table view: %v", model.currentView())
	}
}

// -------------------------------------------------------------------
// AC-TUX3-010 --yes / non-TTY text fallback + ZERO ANSI under NO_COLOR / piped
// -------------------------------------------------------------------

// TestPreviewFallbackNonInteractiveTextSummary asserts that when Interactive
// is false (--yes or non-TTY), the preview returns a plain-text classification
// summary consuming the SAME classification model (AC-TUX3-010).
func TestPreviewFallbackNonInteractiveTextSummary(t *testing.T) {
	in := allFourClassesInputs()
	out := renderFallback(classifyAll(in, allFourClassesPredicate()), true /* noColor */)

	for _, class := range []ChangeClass{ClassAdd, ClassUpdate, ClassPreserveUserOwned, ClassConflict} {
		if !strings.Contains(out, class.String()) {
			t.Errorf("fallback text missing class label %q; output:\n%s", class.String(), out)
		}
	}
}

// TestPreviewFallbackZeroANSIUnderNoColor asserts the text fallback emits ZERO
// ANSI escape sequences when NoColor is true (AC-TUX3-010 — `\x1b[` absence).
func TestPreviewFallbackZeroANSIUnderNoColor(t *testing.T) {
	in := allFourClassesInputs()
	out := renderFallback(classifyAll(in, allFourClassesPredicate()), true /* noColor */)

	if n := countOf(out, "\x1b["); n != 0 {
		t.Errorf("fallback text contains %d ANSI escape sequences under NO_COLOR (AC-TUX3-010 requires zero); output:\n%s", n, out)
	}
}

// TestPreviewFallbackZeroANSIWhenPiped asserts the text fallback emits ZERO
// ANSI escape sequences when the output is piped (non-TTY), independent of
// NO_COLOR (AC-TUX3-010 — piped output path).
func TestPreviewFallbackZeroANSIWhenPiped(t *testing.T) {
	in := allFourClassesInputs()
	// NoColor=false but the piped/non-TTY path still strips ANSI because the
	// fallback never emits color codes regardless of NO_COLOR (the TUI owns
	// color rendering; the fallback is structurally color-free).
	out := renderFallback(classifyAll(in, allFourClassesPredicate()), false /* noColor */)

	if n := countOf(out, "\x1b["); n != 0 {
		t.Errorf("fallback text contains %d ANSI escape sequences on piped output (AC-TUX3-010 requires zero); output:\n%s", n, out)
	}
}

// -------------------------------------------------------------------
// AC-TUX3-014 — BOTH surfaces label `preserved (user-owned)`
// -------------------------------------------------------------------

// TestPreservedLabelInTUITable asserts the TUI table surface contains the
// `preserved (user-owned)` label (AC-TUX3-014).
func TestPreservedLabelInTUITable(t *testing.T) {
	in := allFourClassesInputs()
	model := newPreviewModel(in, allFourClassesPredicate(), PreviewOptions{Interactive: true, Width: 80, Height: 24})

	rendered := model.tableView()
	if !strings.Contains(rendered, "preserved (user-owned)") {
		t.Errorf("TUI table missing 'preserved (user-owned)' label (AC-TUX3-014); rendered:\n%s", rendered)
	}
}

// TestPreservedLabelInTextFallback asserts the text fallback surface contains
// the `preserved (user-owned)` label (AC-TUX3-014).
func TestPreservedLabelInTextFallback(t *testing.T) {
	in := allFourClassesInputs()
	out := renderFallback(classifyAll(in, allFourClassesPredicate()), true /* noColor */)

	if !strings.Contains(out, "preserved (user-owned)") {
		t.Errorf("text fallback missing 'preserved (user-owned)' label (AC-TUX3-014); output:\n%s", out)
	}
}

// TestPreservedLabelDerivedFromSharedPredicate asserts the `preserved
// (user-owned)` classification is derived from the SAME injected predicate at
// BOTH surfaces — the preview MUST NOT implement a parallel heuristic
// (REQ-TUX3-002 / plan.md §G anti-pattern).
func TestPreservedLabelDerivedFromSharedPredicate(t *testing.T) {
	// Predicate that marks NOTHING as user-owned → no file should classify as
	// preserved, on EITHER surface.
	noneOwned := UserOwnedPredicate(func(string) bool { return false })
	in := allFourClassesInputs()

	tuiOut := newPreviewModel(in, noneOwned, PreviewOptions{Interactive: true, Width: 80, Height: 24}).tableView()
	fbOut := renderFallback(classifyAll(in, noneOwned), true)

	if strings.Contains(tuiOut, "preserved (user-owned)") {
		t.Errorf("TUI table shows 'preserved (user-owned)' but predicate marks NOTHING user-owned — parallel heuristic present (REQ-TUX3-002 violation); output:\n%s", tuiOut)
	}
	if strings.Contains(fbOut, "preserved (user-owned)") {
		t.Errorf("text fallback shows 'preserved (user-owned)' but predicate marks NOTHING user-owned — parallel heuristic present (REQ-TUX3-002 violation); output:\n%s", fbOut)
	}
}

// -------------------------------------------------------------------
// AC-TUX3-008 (convergence) — single entry point consumed by both call sites
// -------------------------------------------------------------------

// TestPreviewClassificationEntryPointExists asserts the single preview entry
// point exists with the expected signature so BOTH update.go ConfirmMerge call
// sites (Known Issue #7) can converge on it. This is a compile-time contract
// check: if the entry point is renamed or its signature drifts, this test
// fails to compile.
func TestPreviewClassificationEntryPointExists(t *testing.T) {
	in := allFourClassesInputs()
	// Interactive=false → text fallback path (no live TUI), returns confirmed=true.
	confirmed, err := PreviewClassification(in, allFourClassesPredicate(), PreviewOptions{Interactive: false, NoColor: true})
	if err != nil {
		t.Fatalf("PreviewClassification fallback returned error: %v", err)
	}
	if !confirmed {
		t.Errorf("PreviewClassification fallback returned confirmed=false; want true (non-interactive proceeds)")
	}
}
