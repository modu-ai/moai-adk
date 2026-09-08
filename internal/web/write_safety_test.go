package web

// SPEC-WEB-WRITE-SAFETY-001 — web-layer RED-first tests carrying AC-WWS-003,
// AC-WWS-004 (value-invariant save via the full handleSave path) and
// AC-WWS-006 (duplicate form values are not silently resolved to the first
// value). Reduced from the M1(d) live reproduction: a value-invariant POST
// /save rewrote feedback.yaml (blank-line loss), git-strategy.yaml (dirty-gate
// bypass re-marshal), and corrupted unedited workflow scalars via first-value
// adoption — all observed against the pre-fix tree.

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// seedWriteSafetyFixture writes minimal section fixtures into a test project
// root, mirroring the two files the live reproduction saw rewritten. The
// feedback fixture carries a blank line and comments (presentation elements
// REQ-WWS-005 protects); git-strategy carries mode: team.
func seedWriteSafetyFixture(t *testing.T, root string) (feedbackBefore, gitStrategyBefore string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	feedbackBefore = "feedback:\n" +
		"    # repository comment line one\n" +
		"    repository: modu-ai/moai-adk\n" +
		"\n" +
		"    # auto_submit comment\n" +
		"    auto_submit: false\n"
	if err := os.WriteFile(filepath.Join(dir, "feedback.yaml"), []byte(feedbackBefore), 0o644); err != nil {
		t.Fatal(err)
	}
	gitStrategyBefore = "git_strategy:\n    mode: team\n"
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(gitStrategyBefore), 0o644); err != nil {
		t.Fatal(err)
	}
	return feedbackBefore, gitStrategyBefore
}

// TestHandleSaveValueInvariantLeavesSectionsByteIdentical carries AC-WWS-003
// and AC-WWS-004 through the full write path: a POST /save that submits values
// equal to the persisted ones must leave BOTH section files byte-identical.
// RED on the defective tree — applySchemaEdits rewrites feedback.yaml (blank
// line dropped by re-encode) and git-strategy.yaml (unconditional
// SetSection("git_strategy") raises the dirty flag, so Save() re-marshals).
func TestHandleSaveValueInvariantLeavesSectionsByteIdentical(t *testing.T) {
	root := t.TempDir()
	feedbackBefore, gitStrategyBefore := seedWriteSafetyFixture(t, root)

	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }

	form := url.Values{
		"__profile":                     {"default"},
		"git_strategy.mode":             {"team"},             // fixture's persisted value — unchanged
		"feedback.repository":           {"modu-ai/moai-adk"}, // fixture's persisted value — unchanged
		"feedback.auto_submit__present": {"1"},                // submitted unchecked → false = persisted value
	}
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("value-invariant save status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}

	feedbackAfter, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "feedback.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(feedbackAfter) != feedbackBefore {
		t.Errorf("value-invariant save rewrote feedback.yaml\n--- before ---\n%s\n--- after ---\n%s", feedbackBefore, feedbackAfter)
	}
	gitStrategyAfter, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "git-strategy.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(gitStrategyAfter) != gitStrategyBefore {
		t.Errorf("value-invariant save rewrote git-strategy.yaml (dirty-gate bypassed)\n--- before ---\n%s\n--- after ---\n%s", gitStrategyBefore, gitStrategyAfter)
	}
}

// TestHandleSaveGitStrategyRealChangeStillRewrites is the AC-WWS-004 positive
// control through the full write path: a genuinely changed git_strategy value
// must still rewrite git-strategy.yaml. Without it, a green on the
// value-invariant test above cannot be distinguished from a dead save path.
func TestHandleSaveGitStrategyRealChangeStillRewrites(t *testing.T) {
	root := t.TempDir()
	seedWriteSafetyFixture(t, root)

	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }

	form := url.Values{
		"__profile":         {"default"},
		"git_strategy.mode": {"personal"}, // real change from persisted "team"
	}
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("real-change save status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}
	after, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "git-strategy.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), "mode: personal") {
		t.Errorf("positive control failed: changed git_strategy value not persisted:\n%s", after)
	}
}

// TestHandleSaveGreenfieldSectionCreation carries SPEC-SEAM-GREENFIELD-001
// AC-004: the first save on an ABSENT section file (template-drift state —
// SPEC-MCP-CONSOLE-001/PRECOMMIT-GATE-SCOPE-001 grew sectionRootKeys without
// backfilling older projects) must return 200 and create the file carrying the
// submitted edit and mode 0644. The submission is a REAL change (REQ-6, C6):
// mcp.tools.spec_progress.enabled is fail-open (an absent key means ON,
// AbsentDefault "true"), so an explicit OFF differs from the absent-default
// interpretation and passes the C6 value-invariant gate — a no-op submission
// would be silently skipped and the guard would be vacuously green with no fix.
func TestHandleSaveGreenfieldSectionCreation(t *testing.T) {
	root := t.TempDir()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	// mcp.yaml is deliberately absent — greenfield is the state under test.

	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }

	form := url.Values{
		"__profile":                                {"default"},
		"mcp.tools.spec_progress.enabled":          {""},  // unchecked → false
		"mcp.tools.spec_progress.enabled__present": {"1"}, // submitted (not unsubmitted)
	}
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("greenfield first-save status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}

	raw, err := os.ReadFile(filepath.Join(sections, "mcp.yaml"))
	if err != nil {
		t.Fatalf("section file not created by first save: %v", err)
	}
	if !strings.Contains(string(raw), "spec_progress") || !strings.Contains(string(raw), "enabled: false") {
		t.Errorf("submitted edit not persisted in created mcp.yaml:\n%s", raw)
	}
	info, err := os.Stat(filepath.Join(sections, "mcp.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o644 {
		t.Errorf("created file mode = %o, want 644 (template convention)", got)
	}
}

// TestHandleSaveUntouchedRenderedBodyLeavesTrackedConfigByteIdentical carries
// sync-audit F1-D through the RENDER→PARSE→GATE loop: it asks the real
// renderer what it shows for an ABSENT default-ON bool, submits exactly that
// state back (the untouched-save scenario — the user changed nothing and
// clicked Save), and requires the tracked section file to stay
// byte-identical. Unlike the ApplySchemaEdits-direct polarity tests, this
// walks the full loop the defect lived in: the pre-repair renderer drew the
// absent key as OFF, the untouched save submitted the empty value, the gate
// recorded it as a real change, and `todo: enabled: false` appeared —
// silently disabling the runtime guidance.
func TestHandleSaveUntouchedRenderedBodyLeavesTrackedConfigByteIdentical(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// The audit-experiment fixture: a workflow.yaml with NO todo block (the
	// near-universal state — the distributed template ships none).
	before := "workflow:\n    project:\n        continuation: card\n"
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(before), 0o644); err != nil {
		t.Fatal(err)
	}

	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }
	h := a.routes()

	// 1. Render, and read back what the renderer shows for the absent key.
	getRec := httptest.NewRequest(http.MethodGet, "/settings", nil)
	getRec.Host = "127.0.0.1:8080"
	getRes := httptest.NewRecorder()
	h.ServeHTTP(getRes, getRec)
	if getRes.Code != http.StatusOK {
		t.Fatalf("GET /settings status = %d, want 200", getRes.Code)
	}
	checkedValue := ""
	for _, m := range regexp.MustCompile(`name="workflow.todo.enabled" value="([^"]*)" checked`).FindAllStringSubmatch(getRes.Body.String(), -1) {
		checkedValue = m[1] // the checked radio's submitted value (last wins, browser-identical)
	}
	if checkedValue == "" {
		t.Fatal("renderer shows no checked radio for the absent default-ON key — expected ON (value=1) after the polarity repair")
	}

	// 2. Submit exactly the rendered state (companion + checked value).
	form := url.Values{
		"__profile":                      {"default"},
		"workflow.todo.enabled__present": {"1"},
		"workflow.todo.enabled":          {checkedValue},
	}
	rec := servePost(t, h, "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("untouched save status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}

	// 3. The untouched save must not touch the tracked file.
	after, err := os.ReadFile(filepath.Join(dir, "workflow.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != before {
		t.Errorf("untouched rendered submission rewrote the tracked section file (no-op expected)\n--- before ---\n%s\n--- after ---\n%s", before, after)
	}
}

// TestParseSchemaFormDuplicateFormValuesNotSilentlyFirst carries AC-WWS-006
// (RED-first): when the same form name is submitted twice with different
// values, the parser must NOT quietly adopt the first value — it must join the
// duplicate into the atomic-reject error set (or apply a documented resolution
// rule). RED on the defective tree: parseSchemaForm reads r.PostFormValue /
// r.PostForm[name][0], silently taking the first value.
func TestParseSchemaFormDuplicateFormValuesNotSilentlyFirst(t *testing.T) {
	// Use a real TypeText schema field so the duplicate rides the same code
	// path a browser duplicate submission would.
	var textField string
	for _, f := range settings.AllFields() {
		if schemaEditableField(f) && f.Type == settings.TypeText {
			textField = f.Name
			break
		}
	}
	if textField == "" {
		t.Skip("no editable TypeText schema field available")
	}

	form := url.Values{}
	form.Add(textField, "first-value")
	form.Add(textField, "second-value")
	req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}

	edits, errs := parseSchemaForm(req, nil)
	if len(errs) == 0 {
		t.Errorf("duplicate form value for %q not detected — parser silently adopted %q", textField, edits[textField])
	}
	if _, duplicated := edits[textField]; duplicated {
		t.Errorf("duplicate form value for %q resolved to a single edit %q instead of being detected", textField, edits[textField])
	}
}

// TestParseSchemaFormDuplicateCompanionRejected covers the hidden-bool
// companion duplicate branch: name+"__present" submitted twice is the same
// defect shape as a duplicated value field.
func TestParseSchemaFormDuplicateCompanionRejected(t *testing.T) {
	var boolField string
	for _, f := range settings.AllFields() {
		if schemaEditableField(f) && f.Type == settings.TypeBool {
			boolField = f.Name
			break
		}
	}
	if boolField == "" {
		t.Skip("no editable TypeBool schema field available")
	}

	form := url.Values{}
	form.Add(boolField+"__present", "1")
	form.Add(boolField+"__present", "1")
	form.Add(boolField, "1")
	req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}

	edits, errs := parseSchemaForm(req, nil)
	if _, ok := errs[boolField]; !ok {
		t.Errorf("duplicate companion for %q not detected; edits=%v errs=%v", boolField, edits, errs)
	}
	if _, duplicated := edits[boolField]; duplicated {
		t.Errorf("duplicated companion still produced an edit for %q", boolField)
	}
}

// TestParseSchemaFormEmptySubmitsOptIn documents the EmptySubmits rule the
// duplicate guard must not break: for those fields a "" submission IS the
// value (unset restore), so a single empty submission must still produce an
// edit.
func TestParseSchemaFormEmptySubmitsOptIn(t *testing.T) {
	var field string
	for _, f := range settings.AllFields() {
		if schemaEditableField(f) && f.EmptySubmits {
			field = f.Name
			break
		}
	}
	if field == "" {
		t.Skip("no EmptySubmits field available")
	}

	form := url.Values{}
	form.Add(field, "")
	req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}

	edits, errs := parseSchemaForm(req, nil)
	if v, ok := edits[field]; !ok || v != "" {
		t.Errorf("EmptySubmits field %q: empty submission must produce edit value \"\"; edits=%v errs=%v", field, edits, errs)
	}
}
