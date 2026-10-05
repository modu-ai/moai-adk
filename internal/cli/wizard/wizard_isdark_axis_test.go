package wizard

// AC-ITI-011 (3): the downgrade confirm and the profile wizard's first group
// render DIFFERENT ANSI-inclusive strings under the two wizardIsDark axes,
// and each axis's render matches its saved golden. The axis is forced through
// the package variable only — the environment is never touched.

import (
	"flag"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/ptycaptest"
)

// updateAxisGolden rewrites this file's axis goldens (the wizard package has
// its own test binary, so -update-golden is registered here separately).
var updateAxisGolden = flag.Bool("update-golden", false, "rewrite the wizardIsDark axis goldens under testdata/axis")

// forceAxis pins the light/dark axis for the duration of one draw and returns
// the restore func.
func forceAxis(dark bool) func() {
	orig := wizardIsDark
	wizardIsDark = func() bool { return dark }
	return func() { wizardIsDark = orig }
}

// drawConfirmAxis draws the downgrade confirm under the forced axis.
func drawConfirmAxis(t *testing.T, dark bool) string {
	t.Helper()
	restore := forceAxis(dark)
	defer restore()
	var v bool
	form := NewDowngradeConfirmForm("en", "v9.9.9", "v1.0.0", &v)
	d := ptycaptest.NewFormDriver(t, form)
	return d.View()
}

// drawProfileFirstGroupAxis draws the profile wizard's first page under the
// forced axis.
func drawProfileFirstGroupAxis(t *testing.T, dark bool) string {
	t.Helper()
	restore := forceAxis(dark)
	defer restore()
	initial := ProfileResult{ConversationLang: "en", GitCommitLang: "en", CodeCommentLang: "en", DocLang: "en"}
	form := NewProfileForm(profileStepperOptions(), initial, "en")
	d := ptycaptest.NewFormDriver(t, form)
	return d.View()
}

func TestWizardIsDark_AxisGoldens(t *testing.T) {
	cases := []struct {
		name string
		draw func(t *testing.T, dark bool) string
	}{
		{"downgrade-confirm", drawConfirmAxis},
		{"profile-first-group", drawProfileFirstGroupAxis},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dark := tc.draw(t, true)
			light := tc.draw(t, false)
			if dark == light {
				t.Fatal("dark and light renders are identical — the forced axis must change the ANSI stream")
			}
			if err := ptycaptest.CompareGolden("testdata/axis", tc.name+"-dark", dark, *updateAxisGolden); err != nil {
				t.Fatal(err)
			}
			if err := ptycaptest.CompareGolden("testdata/axis", tc.name+"-light", light, *updateAxisGolden); err != nil {
				t.Fatal(err)
			}
		})
	}
}
