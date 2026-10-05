package cli

// Card t1139 sync re-audit N1: a name init stored verbatim must survive forced
// updates (since card t1162 the section templates escape the name, so init
// stores every name verbatim). The update render context carries the existing project.name /
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
	// Card t1162: the section templates escape the name inside its
	// double-quoted scalar, so YAML escape characters are stored verbatim.
	// "lone backslash" also reproduces, byte for byte, the project.yaml and
	// snapshot a pre-fix init left behind for a typed `a\\b` (it rendered
	// `name: "a\\b"` and stored `a\b`), so it covers projects initialized
	// before the fix as well.
	{"lone backslash", `a\b`},
	{"yaml backslash escape", `a\\b`},
	{"embedded double quotes", `Kim "Goos"`},
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

// TestLoadUpdateIdentity_InitEscapedNamesRoundTrip replaces the former
// known-limitation pin (sync-audit N3, card t1162): with the name escaped in
// the render, a name carrying a YAML escape character is stored verbatim and
// round-trips through loadUpdateIdentity instead of reading back as "".
func TestLoadUpdateIdentity_InitEscapedNamesRoundTrip(t *testing.T) {
	for _, name := range []string{`a\\b`, `Kim "Goos"`} {
		// One subtest per init: the home guard's seam cleanups run at
		// (sub)test end, so two inits in one test body trip it.
		t.Run(name, func(t *testing.T) {
			root := initIdentityProjectNamed(t, "proj", name, name)
			assertInitWroteName(t, root, name)
			gotProject, gotUser := loadUpdateIdentity(root)
			if gotProject != name || gotUser != name {
				t.Errorf("loadUpdateIdentity = (%q, %q), want (%q, %q)", gotProject, gotUser, name, name)
			}
		})
	}
}
