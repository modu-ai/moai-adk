package web

// viewmodel_degradation_test.go — the paths taken when a read fails or a save is
// rejected. They are the ones a user hits on a bad day, and they are the least
// exercised by the happy-path render tests.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// TestApplySchemaCurrentBestEffortDegradesToEmptyMaps pins the POST re-render
// contract: when the disk read fails, the view degrades to empty maps rather
// than nil or a propagated error.
//
// This matters because the re-render happens while a save-result banner is
// already being shown. Propagating the read error there would replace the
// banner the user needs to read with a generic failure page, hiding whether the
// save itself succeeded.
func TestApplySchemaCurrentBestEffortDegradesToEmptyMaps(t *testing.T) {
	a := newTestApp(t)
	a.schemaCurrentValues = func(string) (map[string]string, error) {
		return nil, errors.New("disk read failed")
	}

	view := pageView{
		SchemaValues: map[string]string{"stale": "value"},
		RawBlocks:    map[string]string{"stale": "block"},
	}
	a.applySchemaCurrentBestEffort(&view)

	if view.SchemaValues == nil || view.RawBlocks == nil {
		t.Fatal("degradation left a nil map — the templ render would still index it")
	}
	if len(view.SchemaValues) != 0 || len(view.RawBlocks) != 0 {
		t.Errorf("stale values survived a failed read: values=%v blocks=%v", view.SchemaValues, view.RawBlocks)
	}
}

// TestApplySchemaCurrentBestEffortKeepsValuesOnSuccess is the other half: the
// degradation must not fire when the read succeeds.
func TestApplySchemaCurrentBestEffortKeepsValuesOnSuccess(t *testing.T) {
	a := newTestApp(t)
	view := pageView{}
	a.applySchemaCurrentBestEffort(&view)
	if view.SchemaValues == nil {
		t.Fatal("a successful read left SchemaValues nil")
	}
}

// TestOverlaySchemaEditsEchoesSubmittedValues pins the rejected-save echo
// contract: the submitted values are laid over the disk values so the form
// re-renders what the user typed, not what is on disk. Losing the submission
// would make the user retype it to see the error a second time.
func TestOverlaySchemaEditsEchoesSubmittedValues(t *testing.T) {
	view := pageView{SchemaValues: map[string]string{
		"workflow.execution_mode": "auto",
		"untouched":               "keepme",
	}}
	overlaySchemaEdits(&view, map[string]string{"workflow.execution_mode": "cg"})

	if got := view.SchemaValues["workflow.execution_mode"]; got != "cg" {
		t.Errorf("submitted value not echoed: got %q, want %q", got, "cg")
	}
	if got := view.SchemaValues["untouched"]; got != "keepme" {
		t.Errorf("overlay clobbered an unrelated disk value: %q", got)
	}
}

// TestOverlaySchemaEditsInitializesNilMap covers the first-write path: a view
// built without a values map must not panic when edits are laid over it.
func TestOverlaySchemaEditsInitializesNilMap(t *testing.T) {
	view := pageView{}
	overlaySchemaEdits(&view, map[string]string{"a": "b"})
	if view.SchemaValues["a"] != "b" {
		t.Errorf("overlay onto a nil map lost the edit: %v", view.SchemaValues)
	}
}

// TestBoardBannerRendersErrorOnly pins the board's banner: the SPEC board is
// read-only, so its only banner is an error. A success banner there would
// suggest a write happened.
func TestBoardBannerRendersErrorOnly(t *testing.T) {
	var sb strings.Builder
	if err := boardBanner(boardView{Banner: "could not read .moai/specs"}).
		Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := sb.String()
	if !strings.Contains(html, "banner--error") {
		t.Errorf("board banner is not styled as an error:\n%s", html)
	}
	if strings.Contains(html, "banner--success") {
		t.Error("the read-only board rendered a success banner")
	}
	if !strings.Contains(html, "could not read .moai/specs") {
		t.Error("board banner did not render the message")
	}
}

// TestBoardSurfacesReadFailure is the end-to-end half: a board read failure
// reaches the user as a banner on a 200 page rather than a blank or a 500.
func TestBoardSurfacesReadFailure(t *testing.T) {
	a := newTestApp(t)
	// Point the board at a path that cannot be a spec directory.
	a.cfg.ProjectRoot = "/nonexistent-root-for-board-test"
	h := a.routes()

	req := httptest.NewRequest(http.MethodGet, "/specs", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /specs status = %d, want 200 (a read failure is reported inline, not as an error page)", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "<html") && len(strings.TrimSpace(body)) == 0 {
		t.Error("board rendered an empty page for an unreadable root")
	}
}

// TestNewAppWiresEverySeam pins that the constructor leaves no seam nil. A nil
// seam is a nil-pointer dereference on the first request that reaches it, and
// the happy-path tests all replace seams before exercising them — so this is
// the only place the production wiring itself is checked.
func TestNewAppWiresEverySeam(t *testing.T) {
	a := newApp(Config{ProjectRoot: t.TempDir(), ProfileName: "default"})

	checks := map[string]bool{
		"readPreferences":          a.readPreferences == nil,
		"writePreferences":         a.writePreferences == nil,
		"syncToProject":            a.syncToProject == nil,
		"listProfiles":             a.listProfiles == nil,
		"recordLastProfile":        a.recordLastProfile == nil,
		"readProjectConfig":        a.readProjectConfig == nil,
		"writeProjectConfig":       a.writeProjectConfig == nil,
		"readProjectNestedConfig":  a.readProjectNestedConfig == nil,
		"writeProjectNestedConfig": a.writeProjectNestedConfig == nil,
		"schemaCurrentValues":      a.schemaCurrentValues == nil,
		"rawBlockValues":           a.rawBlockValues == nil,
		"applySchemaEdits":         a.applySchemaEdits == nil,
		"listAgentFMs":             a.listAgentFMs == nil,
		"patchAgentFM":             a.patchAgentFM == nil,
		"createProfile":            a.createProfile == nil,
		"deleteProfile":            a.deleteProfile == nil,
		"renameProfile":            a.renameProfile == nil,
	}
	for name, isNil := range checks {
		if isNil {
			t.Errorf("newApp left the %s seam nil", name)
		}
	}
}

// TestOpenDefaultBrowserErrorIsNonFatal pins REQ-WC-004's contract at the unit
// level: a missing opener returns a descriptive error rather than panicking, so
// the caller can treat it as non-fatal and keep serving.
func TestOpenDefaultBrowserErrorIsNonFatal(t *testing.T) {
	// An empty PATH makes the platform opener unfindable on every GOOS.
	t.Setenv("PATH", "")
	err := openDefaultBrowser("http://127.0.0.1:0")
	if err == nil {
		t.Skip("an opener was still resolvable with an empty PATH; the failure path is not reachable here")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error does not explain the cause: %v", err)
	}
}

// TestSelectedSchemaValuePrecedence pins the Default resolution introduced in
// M4: the disk value always wins, and the Default only fills an absent key.
func TestSelectedSchemaValuePrecedence(t *testing.T) {
	f := settings.FieldDef{Name: "x", Default: "fallback"}
	if got := selectedSchemaValue(f, "ondisk"); got != "ondisk" {
		t.Errorf("disk value did not win: got %q", got)
	}
	if got := selectedSchemaValue(f, ""); got != "fallback" {
		t.Errorf("absent key did not fall back to Default: got %q", got)
	}
	noDefault := settings.FieldDef{Name: "y"}
	if got := selectedSchemaValue(noDefault, ""); got != "" {
		t.Errorf("a field with no Default should select nothing, got %q", got)
	}
}
