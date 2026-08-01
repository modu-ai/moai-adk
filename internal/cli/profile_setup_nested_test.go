package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// The TUI nested-config round-trip / empty-preserve / config-manager tests were
// removed with the widgets they covered: the wizard no longer collects the 3
// nested quality fields or the 4 nested git auto-detection fields, so
// nestedTUIInputs, readCurrentNestedConfig and persistProjectNestedConfig no
// longer exist in this package. The shared seam they delegated to
// (settings.ReadProjectNestedConfig / settings.WriteProjectNestedConfig) is
// unchanged and keeps its own round-trip + empty=preserve coverage in
// internal/settings/nested_test.go, which the web console also depends on.

// TestTUINestedConfigNoParallelWriter pins the AP-2 invariant that survived the
// widget removal: profile_setup.go must never grow a parallel yaml.Marshal /
// os.WriteFile config writer. It must reach project config only through the
// config-manager seam (persistProjectConfig). Comment lines are excluded so
// doctrine prose ("no direct yaml.Marshal/os.WriteFile") does not false-positive.
func TestTUINestedConfigNoParallelWriter(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("profile_setup.go")
	if err != nil {
		t.Fatalf("read profile_setup.go: %v", err)
	}
	codeLines := nonCommentLines(string(data))
	if strings.Contains(codeLines, "yaml.Marshal") {
		t.Error("profile_setup.go must NOT call yaml.Marshal directly (AP-2 — use the shared config-manager seam)")
	}
	if strings.Contains(codeLines, "os.WriteFile") {
		t.Error("profile_setup.go must NOT call os.WriteFile directly for config persistence (AP-2)")
	}
	// The surviving project-config write must still go through the seam.
	if !strings.Contains(codeLines, "persistProjectConfig") {
		t.Error("profile_setup.go must drive the config-manager write seam (persistProjectConfig)")
	}
}

// nonCommentLines returns src with whole-line `//` comments stripped, so a grep
// guard tests actual code rather than doctrine prose in comments.
func nonCommentLines(src string) string {
	var b strings.Builder
	for _, line := range strings.Split(src, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

// TestPermissionModeNormalizeAcceptEdits covers AC-WC10-015: acceptEdits normalizes
// to empty string so no redundant override is persisted. The TUI applies this in
// runProfileSetup; the schema's permission_mode Persist.Normalize encodes the same
// semantic for both surfaces.
func TestPermissionModeNormalizeAcceptEdits(t *testing.T) {
	// Schema-level normalization (shared by both surfaces).
	f, ok := settings.Field("permission_mode")
	if !ok {
		t.Fatal("permission_mode field not found in schema")
	}
	if f.Persist.Normalize == nil {
		t.Fatal("permission_mode Persist.Normalize is nil")
	}
	if got := f.Persist.Normalize("acceptEdits"); got != "" {
		t.Errorf("schema normalize(acceptEdits) = %q, want empty", got)
	}
	if got := f.Persist.Normalize("plan"); got != "plan" {
		t.Errorf("schema normalize(plan) = %q, want plan", got)
	}

	// TUI source-level guard: profile_setup.go must still apply the acceptEdits→""
	// normalization in its save path.
	data, err := os.ReadFile("profile_setup.go")
	if err != nil {
		t.Fatalf("read profile_setup.go: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "permissionMode == defaultPermissionMode") {
		t.Error("profile_setup.go must normalize acceptEdits permission mode to empty (REQ-WC10-014)")
	}
}

// TestTUIEmptyLabelsSchemaSourced covers AC-WC10-014 (TUI side): the wizard sources
// its empty-option labels from the schema (settings.EmptyLabelFor), not inline
// literals — so both surfaces render the IDENTICAL canonical label per field.
//
// git_convention was dropped from the asserted set when its Select was removed from
// the wizard; the remaining three are the schema-backed selects the wizard still
// renders with an empty option. model_policy keeps the source grep: it has no schema
// field (the web console dropped it), so its select is still assembled inline.
func TestTUIEmptyLabelsSchemaSourced(t *testing.T) {
	t.Parallel()
	txt := getProfileText("en")
	for _, field := range []string{"model", "effort_level", "development_mode"} {
		want := settings.EmptyLabelFor(field)
		if want == "" {
			t.Errorf("schema declares no empty label for %q", field)
			continue
		}
		opts := schemaSelectOptions(txt, field, true)
		if len(opts) == 0 || opts[0].Value != "" {
			t.Errorf("field %q: wizard offers no empty option", field)
			continue
		}
		if opts[0].Key != want {
			t.Errorf("field %q: empty option label = %q, want the schema label %q", field, opts[0].Key, want)
		}
	}

	data, err := os.ReadFile("profile_setup.go")
	if err != nil {
		t.Fatalf("read profile_setup.go: %v", err)
	}
	// model_policy is CLI-only (no schema field). Its empty option still reads the
	// schema accessor — which currently returns "" because the field was removed
	// from the schema, so the option renders with a blank label. That blank label is
	// a pre-existing defect of the CLI-only field, out of scope here; the guard below
	// only pins that the wizard has not swapped in an inline literal instead.
	if marker := `settings.EmptyLabelFor("model_policy")`; !strings.Contains(string(data), marker) {
		t.Errorf("profile_setup.go must source empty label from schema for model_policy (expected %s)", marker)
	}
}
