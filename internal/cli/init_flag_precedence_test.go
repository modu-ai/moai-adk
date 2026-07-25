package cli

// SPEC-CLI-WIZARD-RESTRUCTURE-001 M4 — flag-vs-wizard precedence (C41,
// AC-WIZ-016, REQ-WIZ-020).
//
// Once C20 removes the `if result.StandardMode` application gate, the Page-3
// wizard answers reach `opts` on EVERY interactive init. Without a precedence
// rule they would unconditionally overwrite the flag-seeded values, inverting
// today's behaviour and contradicting the documented `--profile` rule
// (init.go: "the wizard fills opts.Profile only when the flag is absent, so the
// flag takes precedence over the wizard answer").
//
// The two bool helpers cannot express this rule by VALUE: `getBoolFlag(cmd,
// "enable-lsp")` returns false both when the flag is absent and when the user
// typed `--enable-lsp=false`, and `getBoolFlagWithDefault(cmd,
// "enforce-quality", true)` returns true both when absent and when the user
// typed `--enforce-quality=true`. The `--enforce-quality=false` and
// `--enable-lsp=false` cases below are therefore the discriminating rows: they
// FAIL against any value-only implementation and pass only when the
// implementation probes `cmd.Flags().Changed(<name>)`.

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/core/project"
)

// seedOptsFromFlags mirrors the four Page-3 `opts` seeds in runInit, using the
// SAME flag getters, so the table below resolves against a faithful baseline.
func seedOptsFromFlags(cmd *cobra.Command) project.InitOptions {
	return project.InitOptions{
		ProjectMode:    getStringFlag(cmd, "project-mode"),
		LSPEnabled:     getBoolFlag(cmd, "enable-lsp"),
		EnforceQuality: getBoolFlagWithDefault(cmd, "enforce-quality", true),
		DesignEnabled:  getBoolFlagWithDefault(cmd, "enable-design", true),
	}
}

// wizardAnswers builds a Page-3 wizard result. Every field is stated
// explicitly at each call site (DAMP over DRY): a partially-specified literal
// would silently carry Go's zero value where the real wizard carries its own
// seeded default (RunWithDefaultsModes seeds EnforceQuality/DesignEnabled true),
// which is exactly the fixture error this signature prevents.
func wizardAnswers(projectMode string, lsp, quality, design bool) *wizard.WizardResult {
	return &wizard.WizardResult{
		ProjectMode:    projectMode,
		LSPEnabled:     lsp,
		EnforceQuality: quality,
		DesignEnabled:  design,
	}
}

// TestFlagBeatsWizard_Page3Settings pins AC-WIZ-016: for each of the four
// overlapping settings, an EXPLICITLY-supplied flag survives the wizard result,
// and an absent flag yields to it.
func TestFlagBeatsWizard_Page3Settings(t *testing.T) {
	// flags maps flag name -> value string to Set (Set marks the flag Changed).
	cases := []struct {
		name   string
		flags  map[string]string
		result *wizard.WizardResult
		want   project.InitOptions
	}{
		// --- project-mode (string) ---------------------------------------
		{
			name:   "project-mode: flag supplied, wizard disagrees -> flag wins",
			flags:  map[string]string{"project-mode": "team"},
			result: wizardAnswers("personal", true, true, true),
			want:   project.InitOptions{ProjectMode: "team", LSPEnabled: true, EnforceQuality: true, DesignEnabled: true},
		},
		{
			name:   "project-mode: flag supplied, wizard agrees -> flag value retained",
			flags:  map[string]string{"project-mode": "team"},
			result: wizardAnswers("team", true, true, true),
			want:   project.InitOptions{ProjectMode: "team", LSPEnabled: true, EnforceQuality: true, DesignEnabled: true},
		},
		{
			name:   "project-mode: flag absent -> wizard answer applies",
			flags:  nil,
			result: wizardAnswers("team", true, true, true),
			want:   project.InitOptions{ProjectMode: "team", LSPEnabled: true, EnforceQuality: true, DesignEnabled: true},
		},
		{
			name:   "project-mode: flag absent, wizard empty -> stays empty",
			flags:  nil,
			result: wizardAnswers("", true, true, true),
			want:   project.InitOptions{ProjectMode: "", LSPEnabled: true, EnforceQuality: true, DesignEnabled: true},
		},

		// --- enable-lsp (bool, flag default false) ------------------------
		{
			// DISCRIMINATING ROW: a value-only implementation sees false here
			// and cannot tell it from "absent", so it lets the wizard's true
			// through and FAILS.
			name:   "enable-lsp: explicit --enable-lsp=false, wizard true -> flag wins (false)",
			flags:  map[string]string{"enable-lsp": "false"},
			result: wizardAnswers("", true, true, true),
			want:   project.InitOptions{LSPEnabled: false, EnforceQuality: true, DesignEnabled: true},
		},
		{
			name:   "enable-lsp: explicit --enable-lsp=true, wizard false -> flag wins (true)",
			flags:  map[string]string{"enable-lsp": "true"},
			result: wizardAnswers("", false, true, true),
			want:   project.InitOptions{LSPEnabled: true, EnforceQuality: true, DesignEnabled: true},
		},
		{
			name:   "enable-lsp: flag absent, wizard true -> wizard applies",
			flags:  nil,
			result: wizardAnswers("", true, true, true),
			want:   project.InitOptions{LSPEnabled: true, EnforceQuality: true, DesignEnabled: true},
		},
		{
			name:   "enable-lsp: flag absent, wizard false -> wizard applies",
			flags:  nil,
			result: wizardAnswers("", false, true, true),
			want:   project.InitOptions{LSPEnabled: false, EnforceQuality: true, DesignEnabled: true},
		},

		// --- enforce-quality (bool, flag default true) --------------------
		{
			// DISCRIMINATING ROW: a value-only implementation cannot tell an
			// explicit false from the flag's own true default.
			name:   "enforce-quality: explicit --enforce-quality=false, wizard true -> flag wins (false)",
			flags:  map[string]string{"enforce-quality": "false"},
			result: wizardAnswers("", true, true, true),
			want:   project.InitOptions{LSPEnabled: true, EnforceQuality: false, DesignEnabled: true},
		},
		{
			name:   "enforce-quality: explicit --enforce-quality=true, wizard false -> flag wins (true)",
			flags:  map[string]string{"enforce-quality": "true"},
			result: wizardAnswers("", true, false, true),
			want:   project.InitOptions{LSPEnabled: true, EnforceQuality: true, DesignEnabled: true},
		},
		{
			name:   "enforce-quality: flag absent, wizard false -> wizard applies",
			flags:  nil,
			result: wizardAnswers("", true, false, true),
			want:   project.InitOptions{LSPEnabled: true, EnforceQuality: false, DesignEnabled: true},
		},
		{
			name:   "enforce-quality: flag absent, wizard true -> wizard applies",
			flags:  nil,
			result: wizardAnswers("", true, true, true),
			want:   project.InitOptions{LSPEnabled: true, EnforceQuality: true, DesignEnabled: true},
		},

		// --- enable-design (bool, flag default true) ----------------------
		{
			// DISCRIMINATING ROW: same shape as enforce-quality — the flag's
			// default is true, so only Changed() distinguishes an explicit false.
			name:   "enable-design: explicit --enable-design=false, wizard true -> flag wins (false)",
			flags:  map[string]string{"enable-design": "false"},
			result: wizardAnswers("", true, true, true),
			want:   project.InitOptions{LSPEnabled: true, EnforceQuality: true, DesignEnabled: false},
		},
		{
			name:   "enable-design: flag absent, wizard false -> wizard applies",
			flags:  nil,
			result: wizardAnswers("", true, true, false),
			want:   project.InitOptions{LSPEnabled: true, EnforceQuality: true, DesignEnabled: false},
		},
		{
			name:   "enable-design: flag absent, wizard true -> wizard applies",
			flags:  nil,
			result: wizardAnswers("", true, true, true),
			want:   project.InitOptions{LSPEnabled: true, EnforceQuality: true, DesignEnabled: true},
		},

		// --- all four flags supplied at once ------------------------------
		{
			name: "all four flags supplied, wizard disagrees on every one -> flags win",
			flags: map[string]string{
				"project-mode":    "team",
				"enable-lsp":      "false",
				"enforce-quality": "false",
				"enable-design":   "false",
			},
			result: wizardAnswers("personal", true, true, true),
			want: project.InitOptions{
				ProjectMode:    "team",
				LSPEnabled:     false,
				EnforceQuality: false,
				DesignEnabled:  false,
			},
		},
		{
			name:   "no flags supplied, wizard answers every one -> wizard wins throughout",
			flags:  nil,
			result: wizardAnswers("team", true, false, false),
			want: project.InitOptions{
				ProjectMode:    "team",
				LSPEnabled:     true,
				EnforceQuality: false,
				DesignEnabled:  false,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newInitTestCmd()
			for name, val := range tc.flags {
				if err := cmd.Flags().Set(name, val); err != nil {
					t.Fatalf("set --%s=%s: %v", name, val, err)
				}
			}

			opts := seedOptsFromFlags(cmd)
			applyWizardPage3ToOpts(cmd, tc.result, &opts)

			if opts.ProjectMode != tc.want.ProjectMode {
				t.Errorf("ProjectMode = %q, want %q", opts.ProjectMode, tc.want.ProjectMode)
			}
			if opts.LSPEnabled != tc.want.LSPEnabled {
				t.Errorf("LSPEnabled = %v, want %v", opts.LSPEnabled, tc.want.LSPEnabled)
			}
			if opts.EnforceQuality != tc.want.EnforceQuality {
				t.Errorf("EnforceQuality = %v, want %v", opts.EnforceQuality, tc.want.EnforceQuality)
			}
			if opts.DesignEnabled != tc.want.DesignEnabled {
				t.Errorf("DesignEnabled = %v, want %v", opts.DesignEnabled, tc.want.DesignEnabled)
			}
		})
	}
}

// TestFlagBeatsWizard_MatchesProfilePrecedence pins the AC-WIZ-016 consistency
// clause: the four Page-3 settings resolve in the SAME direction as the
// documented `--profile` rule — an explicitly-supplied flag is never
// overwritten by the wizard answer. Asserted as one invariant across all four
// so a future per-field regression cannot pass by covering only the bools.
func TestFlagBeatsWizard_MatchesProfilePrecedence(t *testing.T) {
	cmd := newInitTestCmd()
	for name, val := range map[string]string{
		"project-mode":    "team",
		"enable-lsp":      "false",
		"enforce-quality": "false",
		"enable-design":   "false",
	} {
		if err := cmd.Flags().Set(name, val); err != nil {
			t.Fatalf("set --%s=%s: %v", name, val, err)
		}
	}

	opts := seedOptsFromFlags(cmd)
	// Stand-in for the --profile seed at the same point in runInit; the helper
	// must leave it alone (it is resolved by its own documented rule).
	opts.Profile = "low"
	before := opts

	// A wizard result that disagrees with every supplied flag.
	applyWizardPage3ToOpts(cmd, &wizard.WizardResult{
		ProjectMode:    "personal",
		LSPEnabled:     true,
		EnforceQuality: true,
		DesignEnabled:  true,
	}, &opts)

	if opts.ProjectMode != before.ProjectMode ||
		opts.LSPEnabled != before.LSPEnabled ||
		opts.EnforceQuality != before.EnforceQuality ||
		opts.DesignEnabled != before.DesignEnabled {
		t.Errorf("explicitly-supplied flags must survive the wizard result (the --profile rule):\n before: %+v\n after:  %+v", before, opts)
	}
	// The --profile seed is untouched by the Page-3 application — the helper
	// must not reach outside its four fields.
	if opts.Profile != "low" {
		t.Errorf("Profile = %q, want %q (applyWizardPage3ToOpts must not touch it)", opts.Profile, "low")
	}
}
