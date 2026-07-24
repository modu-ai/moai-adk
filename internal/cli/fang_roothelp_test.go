package cli

// Root-help predicate unit tests (SPEC-CLI-TUX-INIT-UPDATE-001 M3,
// REQ-TUXIU-055, AC-TUXIU-024). The predicate inspects ONLY os.Args[1:] shape
// and decides whether the explicit root-help surface (moai --help / -h / help)
// should print the restored logo BEFORE fang.Execute. It MUST NOT match the
// empty arg vector (no-args is already covered by PrintBanner in rootCmd.Run —
// matching [] would double-print the logo on the most-visible surface) nor any
// subcommand-help shape (moai help <sub>, moai <sub> --help).

import "testing"

func TestIsRootHelpArgs(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want bool
	}{
		// NOT matched: empty arg vector — no-args prints the logo via
		// rootCmd.Run→PrintBanner; matching [] here double-prints (HARD).
		{"empty_no_args", []string{}, false},
		// Matched: the three explicit root-help shapes.
		{"double_dash_help", []string{"--help"}, true},
		{"short_h", []string{"-h"}, true},
		{"help_subcommand", []string{"help"}, true},
		// NOT matched: subcommand-help shapes.
		{"help_with_sub", []string{"help", "init"}, false},
		{"sub_with_help_flag", []string{"init", "--help"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isRootHelpArgs(tc.args); got != tc.want {
				t.Errorf("isRootHelpArgs(%q) = %v, want %v", tc.args, got, tc.want)
			}
		})
	}
}
