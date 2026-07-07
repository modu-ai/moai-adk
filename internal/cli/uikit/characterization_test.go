package uikit_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/settings"
)

// Characterization tests for uikit-resident helpers that moved from package cli
// (SPEC-CLI-UIKIT-KERNEL-001). These capture EXISTING behavior — no new
// behavior is defined (REQ-CUK-011, design.md §E AP-5: characterization only).

// TestCharacterize_SchemaBridgeResolver covers the RegisterSchemaBridge callback
// dispatch + SchemaKeyToTUIField + FieldDefTUILabel. The resolver is a package
// global set once; these tests register a deterministic test resolver.
func TestCharacterize_SchemaBridgeResolver(t *testing.T) {
	// Register a test resolver that maps known keys to known labels.
	uikit.RegisterSchemaBridge(func(schemaKey, locale string) (uikit.TuiLabel, bool) {
		if schemaKey == "f.test" {
			return uikit.TuiLabel{Title: "TestTitle", Desc: "TestDesc"}, true
		}
		return uikit.TuiLabel{}, false
	})

	// SchemaKeyToTUIField dispatches through the resolver.
	label, ok := uikit.SchemaKeyToTUIField("f.test", "en")
	if !ok {
		t.Fatal("SchemaKeyToTUIField should resolve f.test via resolver")
	}
	if label.Title != "TestTitle" || label.Desc != "TestDesc" {
		t.Errorf("unexpected label: %+v", label)
	}

	// Unknown key returns false.
	_, ok = uikit.SchemaKeyToTUIField("__nonexistent__", "en")
	if ok {
		t.Error("SchemaKeyToTUIField should return false for unknown key")
	}

	// FieldDefTUILabel dispatches through the same resolver via I18nKey.
	f := settings.FieldDef{I18nKey: "f.test"}
	label2, ok2 := uikit.FieldDefTUILabel(f, "en")
	if !ok2 || label2.Title != "TestTitle" {
		t.Errorf("FieldDefTUILabel should resolve via I18nKey, got: %+v ok=%v", label2, ok2)
	}
}

// TestCharacterize_StatusIcon_AllBranches covers CheckOK/CheckWarn/CheckFail/default.
func TestCharacterize_StatusIcon_AllBranches(t *testing.T) {
	cases := []struct {
		status   uikit.CheckStatus
		nonEmpty bool
	}{
		{uikit.CheckOK, true},
		{uikit.CheckWarn, true},
		{uikit.CheckFail, true},
		{uikit.CheckStatus("bogus"), true}, // default returns "?"
	}
	for _, tc := range cases {
		result := uikit.StatusIcon(tc.status)
		if tc.nonEmpty && result == "" {
			t.Errorf("StatusIcon(%q) returned empty", tc.status)
		}
	}
}

// TestCharacterize_VersionHelpers covers the env-var-driven version helpers.
func TestCharacterize_VersionHelpers(t *testing.T) {
	t.Setenv("MOAI_GIT_VERSION_OVERRIDE", "git 2.50.1")
	if v := uikit.GitVersionOverride(); v != "git 2.50.1" {
		t.Errorf("GitVersionOverride = %q, want %q", v, "git 2.50.1")
	}

	t.Setenv("MOAI_GH_VERSION_OVERRIDE", "gh 2.40.0")
	if v := uikit.GhVersionOverride(); v != "gh 2.40.0" {
		t.Errorf("GhVersionOverride = %q, want %q", v, "gh 2.40.0")
	}

	t.Setenv("MOAI_GOOS_OVERRIDE", "linux")
	t.Setenv("MOAI_GOARCH_OVERRIDE", "arm64")
	if v := uikit.GoosArch(); v != "linux/arm64" {
		t.Errorf("GoosArch = %q, want %q", v, "linux/arm64")
	}
}

// TestCharacterize_SymHelpers covers SymSuccess/SymError/SymWarning (non-empty).
func TestCharacterize_SymHelpers(t *testing.T) {
	if v := uikit.SymSuccess(); v == "" {
		t.Error("SymSuccess should return non-empty")
	}
	if v := uikit.SymError(); v == "" {
		t.Error("SymError should return non-empty")
	}
	if v := uikit.SymWarning(); v == "" {
		t.Error("SymWarning should return non-empty")
	}
}
