package web

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// TestValidatePrefs_ModelField verifies AC-WC2-002a: validatePrefs accepts the
// 6 canonical model aliases plus empty, and rejects an out-of-list value
// (REQ-WC2-002). The canonical set mirrors the wizard SSOT.
func TestValidatePrefs_ModelField(t *testing.T) {
	valid := []string{"", "opus", "opus[1m]", "sonnet", "sonnet[1m]", "haiku", "opusplan"}
	for _, v := range valid {
		errs := validatePrefs(profile.ProfilePreferences{PermissionMode: "acceptEdits", Model: v})
		if _, ok := errs["model"]; ok {
			t.Errorf("model=%q should be valid, got error: %v", v, errs["model"])
		}
	}
	bogus := []string{"gpt-4", "ultra", "claude", "opus2"}
	for _, v := range bogus {
		errs := validatePrefs(profile.ProfilePreferences{PermissionMode: "acceptEdits", Model: v})
		if _, ok := errs["model"]; !ok {
			t.Errorf("model=%q should be rejected, got no error", v)
		}
	}
}

// TestValidatePrefs_EffortLevelField verifies AC-WC2-003: validatePrefs accepts
// the 5 canonical effort levels plus empty, and rejects an out-of-list value
// (REQ-WC2-003).
func TestValidatePrefs_EffortLevelField(t *testing.T) {
	valid := []string{"", "low", "medium", "high", "xhigh", "max"}
	for _, v := range valid {
		errs := validatePrefs(profile.ProfilePreferences{PermissionMode: "acceptEdits", EffortLevel: v})
		if _, ok := errs["effort_level"]; ok {
			t.Errorf("effort_level=%q should be valid, got error: %v", v, errs["effort_level"])
		}
	}
	bogus := []string{"ultra", "extreme", "none", "xxhigh"}
	for _, v := range bogus {
		errs := validatePrefs(profile.ProfilePreferences{PermissionMode: "acceptEdits", EffortLevel: v})
		if _, ok := errs["effort_level"]; !ok {
			t.Errorf("effort_level=%q should be rejected, got no error", v)
		}
	}
}

// TestValidatePrefs_ModelPolicyField verifies AC-WC2-004: validatePrefs delegates
// to template.IsValidModelPolicy — accepting high/medium/low plus empty, rejecting
// an out-of-list value (REQ-WC2-004).
func TestValidatePrefs_ModelPolicyField(t *testing.T) {
	valid := []string{"", "high", "medium", "low"}
	for _, v := range valid {
		errs := validatePrefs(profile.ProfilePreferences{PermissionMode: "acceptEdits", ModelPolicy: v})
		if _, ok := errs["model_policy"]; ok {
			t.Errorf("model_policy=%q should be valid, got error: %v", v, errs["model_policy"])
		}
	}
	bogus := []string{"ultra", "extreme", "opus", "default"}
	for _, v := range bogus {
		errs := validatePrefs(profile.ProfilePreferences{PermissionMode: "acceptEdits", ModelPolicy: v})
		if _, ok := errs["model_policy"]; !ok {
			t.Errorf("model_policy=%q should be rejected, got no error", v)
		}
	}
}

// TestSaveInvalidModelRejected verifies AC-WC2-002a: a POST /save with an
// out-of-list model is rejected (400), no persistence occurs, and the form
// re-renders with a per-field model error.
func TestSaveInvalidModelRejected(t *testing.T) {
	a := newTestApp(t)
	var wrote, synced bool
	a.writePreferences = func(string, profile.ProfilePreferences) error { wrote = true; return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { synced = true; return nil }
	h := a.routes()

	form := url.Values{
		"__profile":       {"default"},
		"permission_mode": {"acceptEdits"},
		"model":           {"gpt-4"},
	}
	rec := servePost(t, h, "/save", form)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("invalid model status = %d, want 400", rec.Code)
	}
	if wrote || synced {
		t.Error("persistence functions called despite invalid model (state must be unchanged)")
	}
	if body := rec.Body.String(); !strings.Contains(body, "unrecognized model") {
		t.Errorf("per-field model error not rendered:\n%s", body)
	}
}

// TestSaveValidModelPersisted verifies AC-WC2-002b: a POST /save with a canonical
// model persists (200) and the written prefs carry the value.
func TestSaveValidModelPersisted(t *testing.T) {
	a := newTestApp(t)
	var wrotePrefs profile.ProfilePreferences
	a.writePreferences = func(_ string, prefs profile.ProfilePreferences) error {
		wrotePrefs = prefs
		return nil
	}
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	h := a.routes()

	form := url.Values{
		"__profile":       {"default"},
		"permission_mode": {"acceptEdits"},
		"model":           {"sonnet[1m]"},
	}
	rec := servePost(t, h, "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid model save status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}
	if wrotePrefs.Model != "sonnet[1m]" {
		t.Errorf("persisted model = %q, want sonnet[1m]", wrotePrefs.Model)
	}
}

// TestSaveInvalidEffortLevelRejected verifies AC-WC2-003: an out-of-list
// effort_level is rejected (400), state unchanged.
func TestSaveInvalidEffortLevelRejected(t *testing.T) {
	a := newTestApp(t)
	var wrote bool
	a.writePreferences = func(string, profile.ProfilePreferences) error { wrote = true; return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	h := a.routes()

	form := url.Values{
		"__profile":       {"default"},
		"permission_mode": {"acceptEdits"},
		"effort_level":    {"ultra"},
	}
	rec := servePost(t, h, "/save", form)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("invalid effort_level status = %d, want 400", rec.Code)
	}
	if wrote {
		t.Error("WritePreferences called despite invalid effort_level")
	}
	if body := rec.Body.String(); !strings.Contains(body, "unrecognized effort level") {
		t.Errorf("per-field effort_level error not rendered:\n%s", body)
	}
}

// TestSaveValidEffortLevelPersisted verifies AC-WC2-003: a canonical effort_level
// persists (200).
func TestSaveValidEffortLevelPersisted(t *testing.T) {
	a := newTestApp(t)
	var wrotePrefs profile.ProfilePreferences
	a.writePreferences = func(_ string, prefs profile.ProfilePreferences) error {
		wrotePrefs = prefs
		return nil
	}
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	h := a.routes()

	form := url.Values{
		"__profile":       {"default"},
		"permission_mode": {"acceptEdits"},
		"effort_level":    {"xhigh"},
	}
	rec := servePost(t, h, "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid effort_level save status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}
	if wrotePrefs.EffortLevel != "xhigh" {
		t.Errorf("persisted effort_level = %q, want xhigh", wrotePrefs.EffortLevel)
	}
}

// TestSaveInvalidModelPolicyRejected verifies AC-WC2-004: an out-of-list
// model_policy is rejected (400) via template.IsValidModelPolicy, state unchanged.
func TestSaveInvalidModelPolicyRejected(t *testing.T) {
	a := newTestApp(t)
	var wrote bool
	a.writePreferences = func(string, profile.ProfilePreferences) error { wrote = true; return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	h := a.routes()

	form := url.Values{
		"__profile":       {"default"},
		"permission_mode": {"acceptEdits"},
		"model_policy":    {"ultra"},
	}
	rec := servePost(t, h, "/save", form)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("invalid model_policy status = %d, want 400", rec.Code)
	}
	if wrote {
		t.Error("WritePreferences called despite invalid model_policy")
	}
	if body := rec.Body.String(); !strings.Contains(body, "unrecognized model policy") {
		t.Errorf("per-field model_policy error not rendered:\n%s", body)
	}
}

// TestRenderModelEffortPolicyAreSelects verifies AC-WC2-005 / REQ-WC2-005: the
// model, effort_level, and model_policy fields render as <select> dropdowns with
// their canonical option sets plus the empty-default option, and NO <input
// type="text"> remains for those three field names.
func TestRenderModelEffortPolicyAreSelects(t *testing.T) {
	a := newTestApp(t)
	a.readPreferences = func(string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{
			Model:       "sonnet[1m]",
			EffortLevel: "xhigh",
			ModelPolicy: "medium",
		}, nil
	}
	a.listProfiles = func() []profile.ProfileEntry {
		return []profile.ProfileEntry{{Name: "default", Current: true}}
	}
	body := serveGet(t, a.routes(), "/").Body.String()

	// Negative: no text inputs for the three constrained fields.
	for _, name := range []string{"model", "effort_level", "model_policy"} {
		needle := `<input type="text" id="` + name + `"`
		if strings.Contains(body, needle) {
			t.Errorf("field %q is still a text input; want <select>", name)
		}
	}

	// Positive: each field is a <select> carrying its name attribute.
	for _, name := range []string{"model", "effort_level", "model_policy"} {
		needle := `name="` + name + `"`
		idx := strings.Index(body, needle)
		if idx < 0 {
			t.Errorf("field %q not found in rendered page", name)
			continue
		}
		// The name attribute must belong to a <select> opening tag. Find the
		// nearest preceding '<' and confirm it opens a <select.
		open := strings.LastIndex(body[:idx], "<")
		if open < 0 || !strings.HasPrefix(body[open:], "<select") {
			t.Errorf("field %q name attribute is not on a <select> element", name)
		}
	}

	// Each canonical option must be present as an <option value="...">.
	for _, opt := range []string{"opus", "opus[1m]", "sonnet", "sonnet[1m]", "haiku", "opusplan"} {
		if !strings.Contains(body, `<option value="`+opt+`"`) {
			t.Errorf("model option %q missing from rendered selects", opt)
		}
	}
	for _, opt := range []string{"low", "medium", "high", "xhigh", "max"} {
		if !strings.Contains(body, `<option value="`+opt+`"`) {
			t.Errorf("effort/policy option %q missing from rendered selects", opt)
		}
	}

	// The currently-selected values must be marked selected.
	for _, want := range []string{
		`<option value="sonnet[1m]" selected`,
		`<option value="xhigh" selected`,
		`<option value="medium" selected`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("expected current value marked selected: %q\nbody:\n%s", want, body)
		}
	}
}

// TestSaveValidModelPolicyPersisted verifies AC-WC2-004 / AC-WC2-007: a canonical
// model_policy persists (200) to the profile store. The synced prefs also carry
// the value (SyncToProjectConfig owns the profile-only decision — REQ-WC2-007 adds
// no new config section).
func TestSaveValidModelPolicyPersisted(t *testing.T) {
	a := newTestApp(t)
	var wrotePrefs profile.ProfilePreferences
	a.writePreferences = func(_ string, prefs profile.ProfilePreferences) error {
		wrotePrefs = prefs
		return nil
	}
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	h := a.routes()

	form := url.Values{
		"__profile":       {"default"},
		"permission_mode": {"acceptEdits"},
		"model_policy":    {"medium"},
	}
	rec := servePost(t, h, "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid model_policy save status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}
	if wrotePrefs.ModelPolicy != "medium" {
		t.Errorf("persisted model_policy = %q, want medium", wrotePrefs.ModelPolicy)
	}
}
