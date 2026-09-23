package cli

// Card t1139 sync re-audit N1: a name init stored verbatim must survive forced
// updates (a name init stores altered — see the known-limitation test at the
// end of this file — is outside that guarantee). The update render context carries the existing project.name /
// user.name; a name that context drops renders as "" while the snapshot BASE
// still holds the init render of that name, so BASE == OLD and the 3-way merge
// takes the empty NEW value — the name is erased and the update reports
// success. The names below are ones init accepts and renders verbatim, so the
// update render must carry them too.
//
// prepareSafeInitHome sets environment variables, the update helper chdirs,
// and captureProcessStderr swaps os.Stderr, so no test in this file may run in
// parallel.

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// initOriginNames are names given directly to init (--name for the project,
// the wizard answer for the user) that init renders verbatim.
var initOriginNames = []struct{ label, name string }{
	{"lone dollar digit", "cost$5"},
	{"dollar lowercase", "my$app"},
	{"lone template open", "a{{b"},
	{"zwj emoji", "👩\u200d💻 x"},
	{"zwj family emoji", "👨\u200d👩\u200d👧 family"},
	{"tab", "a\tb"},
	{"hangul", "홍길동"},
	{"padded", " padded "},
	{"apostrophe", "it's"},
	{"yaml punctuation", "a#b: c"},
}

// assertInitWroteName is the precondition: init must have written the name
// verbatim, or the update assertions below say nothing about init-origin names.
func assertInitWroteName(t *testing.T, root, name string) {
	t.Helper()
	if got := sectionValue(t, sectionsFile(root, "project.yaml"), "project", "name"); got != name {
		t.Fatalf("after init: project.name = %q, want %q (init did not write the name verbatim)", got, name)
	}
	if got := sectionValue(t, sectionsFile(root, "user.yaml"), "user", "name"); got != name {
		t.Fatalf("after init: user.name = %q, want %q (init did not write the name verbatim)", got, name)
	}
}

// TestUpdateForce_InitOriginNamesSurvive: two forced updates after an init
// that wrote the name itself must succeed, keep both names verbatim, and print
// no section-merge failure.
func TestUpdateForce_InitOriginNamesSurvive(t *testing.T) {
	for _, tc := range initOriginNames {
		t.Run(tc.label, func(t *testing.T) {
			root := initIdentityProjectNamed(t, "proj", tc.name, tc.name)
			assertInitWroteName(t, root, tc.name)

			for i := 1; i <= 2; i++ {
				output, syncErr := runForcedTemplateSyncResult(t, root)
				if syncErr != nil {
					t.Fatalf("update #%d halted: %v", i, syncErr)
				}
				for _, line := range strings.Split(output, "\n") {
					if strings.Contains(line, "merge failed") &&
						(strings.Contains(line, "user.yaml") || strings.Contains(line, "project.yaml")) {
						t.Errorf("update #%d: section merge failed: %q", i, line)
					}
				}
				if got := sectionValue(t, sectionsFile(root, "project.yaml"), "project", "name"); got != tc.name {
					t.Errorf("update #%d: project.name = %q, want %q", i, got, tc.name)
				}
				if got := sectionValue(t, sectionsFile(root, "user.yaml"), "user", "name"); got != tc.name {
					t.Errorf("update #%d: user.name = %q, want %q", i, got, tc.name)
				}
			}
		})
	}
}

// TestCleanReinstall_InitOriginNameSurvives measures the clean-reinstall
// render context with an init-origin name the over-rejecting filter dropped:
// the reinstall and the follow-up update must both keep it.
func TestCleanReinstall_InitOriginNameSurvives(t *testing.T) {
	const name = "cost$5"
	root := initIdentityProjectNamed(t, "proj", name, name)
	assertInitWroteName(t, root, name)

	writeTestFile(t, root, ".moai/config/sections/system.yaml", "moai:\n    version: v2.16.1\n")
	writeTestFile(t, root, ".claude/agents/moai/manager-strategy.md", "retired\n")

	var out, errOut bytes.Buffer
	result, err := runCleanReinstall(context.Background(), root, CleanReinstallOptions{
		Out:              &out,
		ErrOut:           &errOut,
		RunMigrateAgency: (&stubMigrateRunner{}).Run,
	})
	if err != nil {
		t.Fatalf("runCleanReinstall: %v\nout: %s\nerr: %s", err, out.String(), errOut.String())
	}
	if !result.Detected.IsV2 {
		t.Fatalf("fixture was not detected as v2; the reinstall body never ran (details: %v)", result.Detected.SignalDetails)
	}
	for _, f := range []struct{ file, key string }{{"project.yaml", "project"}, {"user.yaml", "user"}} {
		if got := sectionValue(t, snapshotFile(root, f.file), f.key, "name"); got != name {
			t.Errorf("after reinstall: snapshot %s.name = %q, want %q (the render carries the name)", f.key, got, name)
		}
		if got := sectionValue(t, sectionsFile(root, f.file), f.key, "name"); got != name {
			t.Errorf("after reinstall: %s.name = %q, want %q", f.key, got, name)
		}
	}

	runForcedTemplateSyncAt(t, root)
	for _, f := range []struct{ file, key string }{{"project.yaml", "project"}, {"user.yaml", "user"}} {
		if got := sectionValue(t, sectionsFile(root, f.file), f.key, "name"); got != name {
			t.Errorf("after reinstall + update: %s.name = %q, want %q", f.key, got, name)
		}
	}
}

// TestUpdateForce_KnownLimitation_InitEscapedProjectName pins a KNOWN
// LIMITATION (sync-audit N3), not a desired behavior. project.yaml.tmpl puts
// the name into a double-quoted YAML scalar without escaping it, so a --name
// carrying a YAML backslash escape is stored by init as a DIFFERENT string
// (`a\\b` typed, `a\b` stored). The stored value does not round-trip through
// the render, loadUpdateIdentity reads it as "", and because the snapshot BASE
// holds that same stored value (BASE == OLD) the merge takes the empty render:
// the forced update erases project.name and reports no error. user.name is not
// affected — init rewrites user.yaml after the render, so its BASE differs
// from the stored value and the merge keeps it as a customization.
//
// The loss is pinned rather than skipped on purpose: a skipped test measures
// nothing, while this one keeps the limitation a measured fact and forces the
// follow-up fix (escape identity values in the init/project template render)
// to flip it deliberately. Once init stores the typed name verbatim, the first
// assertion fails with instructions — replace this test with a survival
// assertion then.
func TestUpdateForce_KnownLimitation_InitEscapedProjectName(t *testing.T) {
	const typed, stored = `a\\b`, `a\b`
	root := initIdentityProjectNamed(t, "proj", typed, typed)

	if got := sectionValue(t, sectionsFile(root, "project.yaml"), "project", "name"); got != stored {
		t.Fatalf("after init: project.name = %q, want %q — init no longer alters escaped names; the limitation is fixed, replace this test with a survival assertion", got, stored)
	}
	if got, _ := loadUpdateIdentity(root); got != "" {
		t.Fatalf("loadUpdateIdentity project = %q, want \"\" (the stored value now round-trips; re-derive this test)", got)
	}

	output, syncErr := runForcedTemplateSyncResult(t, root)
	if syncErr != nil {
		t.Fatalf("update halted: %v", syncErr)
	}
	if strings.Contains(output, "merge failed") {
		t.Errorf("update printed a merge failure; the pinned limitation is silent:\n%s", output)
	}
	if got := sectionValue(t, sectionsFile(root, "project.yaml"), "project", "name"); got != "" {
		t.Errorf("KNOWN LIMITATION changed: project.name after update = %q, want \"\" (erased by the empty render)", got)
	}
	if got := sectionValue(t, sectionsFile(root, "user.yaml"), "user", "name"); got != typed {
		t.Errorf("user.name after update = %q, want %q (the limitation is confined to project.name)", got, typed)
	}
}
