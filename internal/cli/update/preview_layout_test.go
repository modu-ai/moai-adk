package update

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// t694 regression: the preview TUI rendered at a hardcoded 80×24 (the
// constructor fallback, because the call site passes no Width/Height) and the
// model ignored tea.WindowSizeMsg, so the classification table stayed capped
// at 24-8=16 lines (~14 file rows) on every terminal — the observed "table
// breaks off at 14 rows / conflict 2 not shown". These tests pin the two
// repairs: live resize on WindowSizeMsg, and conflicts sorting to the top of
// the table so the class the user most needs to see never sits below the fold.

// TestPreviewModelResizesOnWindowSizeMsg asserts the model adopts the runtime
// terminal size delivered by bubbletea at program start and on every resize.
func TestPreviewModelResizesOnWindowSizeMsg(t *testing.T) {
	model := newPreviewModel(allFourClassesInputs(), allFourClassesPredicate(), PreviewOptions{Interactive: true})

	// Height() reports the data-row area (the constructor's WithHeight
	// subtracts the table's own header), so assert the RESIZE DELTA rather
	// than an absolute accounting: growing the terminal by 16 rows must grow
	// the visible table by the same 16.
	before := model.table.Height()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m := updated.(*previewModel)
	if got := m.table.Height(); got != before+16 {
		t.Errorf("table height after resize = %d, want before(%d)+16", got, before)
	}
	if got := m.table.Width(); got != 120 {
		t.Errorf("table width after resize = %d, want 120", got)
	}
	if got := m.viewport.Height(); got != 40-previewViewportInset {
		t.Errorf("viewport height after resize = %d, want %d", got, 40-previewViewportInset)
	}
}

// TestPreviewModelConflictsSortFirst asserts conflict rows occupy the top of
// the table: on a height-capped table the class demanding a decision must be
// on the first screen, not buried under add/update rows.
func TestPreviewModelConflictsSortFirst(t *testing.T) {
	model := newPreviewModel(allFourClassesInputs(), allFourClassesPredicate(), PreviewOptions{Interactive: true, Width: 80, Height: 24})

	if len(model.classes) == 0 {
		t.Fatal("fixture produced no classes")
	}
	if model.classes[0].Class != ClassConflict {
		t.Errorf("first table row class = %q, want %q (conflicts sort first)", model.classes[0].Class, ClassConflict)
	}
	wantPath := ".moai/config/config.yaml" // the conflict fixture in allFourClassesInputs
	if model.paths[0] != wantPath {
		t.Errorf("first table row path = %q, want %q", model.paths[0], wantPath)
	}
}
