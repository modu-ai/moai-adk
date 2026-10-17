package cli

// factory_anchor_override_placement_test.go — SPEC-LAUNCHER-ARGV-001 REQ-LA-005
// (card-review finding on card t1748): the override flag counts only where it
// stands, never as another option's VALUE — and that holds when the option that
// takes the value is the last letter of a bundled short cluster (`-cn <name>`:
// `c` is a boolean, `n` takes the next token). The launcher segment used to ask
// "does this whole token take the next token as its value?" against a table
// keyed by whole tokens, so a cluster was never a key and the value was read as
// the override.

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestLauncherAnchorOverrideFlagPlacement(t *testing.T) {
	pinSnapshotModel(t)

	// The consumption function itself, over the placements of the flag.
	t.Run("consumption_by_placement", func(t *testing.T) {
		cases := []struct {
			name        string
			args        []string
			wantRest    []string
			wantAllowed bool
		}{
			{"alone", []string{"-c", "--allow-card-anchor", "hello"}, []string{"-c", "hello"}, true},
			{"value of a bundled cluster ending in a value option", []string{"-cn", "--allow-card-anchor", "hello"}, []string{"-cn", "--allow-card-anchor", "hello"}, false},
			{"value of a lone short value option", []string{"-n", "--allow-card-anchor", "hello"}, []string{"-n", "--allow-card-anchor", "hello"}, false},
			{"value of a long value option", []string{"--name", "--allow-card-anchor", "hello"}, []string{"--name", "--allow-card-anchor", "hello"}, false},
			{"attached cluster value takes no next token", []string{"-cnNAME", "--allow-card-anchor"}, []string{"-cnNAME"}, true},
			{"cluster ending in an optional-value option leaves a flag-shaped token alone", []string{"-cw", "--allow-card-anchor"}, []string{"-cw"}, true},
			{"the separator is never a value", []string{"-cn", "--", "--allow-card-anchor"}, []string{"-cn", "--", "--allow-card-anchor"}, false},
			// One argv, one interpretation: consumeLaunchArgs eats the token after
			// --model / --permission-mode / -m even when it is `--`; any other
			// option's following `--` is the launcher's own separator.
			{"the separator is the value of a launcher-consumed option", []string{"--model", "--", "--allow-card-anchor"}, []string{"--model", "--"}, true},
			{"the separator is the value of -m", []string{"-m", "--", "--allow-card-anchor"}, []string{"-m", "--"}, true},
			{"the separator after another value option is the launcher separator", []string{"--name", "--", "--allow-card-anchor"}, []string{"--name", "--", "--allow-card-anchor"}, false},
			{"a cluster of booleans only", []string{"-cc", "--allow-card-anchor"}, []string{"-cc"}, true},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				rest, allowed := consumeAllowCardAnchor(tc.args)
				if !reflect.DeepEqual(rest, tc.wantRest) || allowed != tc.wantAllowed {
					t.Errorf("consumeAllowCardAnchor(%q) = (%q, %v), want (%q, %v)", tc.args, rest, allowed, tc.wantRest, tc.wantAllowed)
				}
			})
		}
	})

	// Through the real leader entry, both arms.
	for _, arm := range leaderArms {
		// The flag stands as -n's value: the anchored launch is still refused and
		// the token is the backend's, not the launcher's.
		t.Run(arm+"/value_of_a_cluster_option_is_not_an_override", func(t *testing.T) {
			fx := newLeaderFixture(t)
			obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "-cn", allowCardAnchorFlag, "hello", "-w", "WT-card"})
			assertRefusedBeforeMutation(t, arm, obs)
			refusalNames(t, arm, obs, fx.card, "WT-card")
			if _, err := os.Stat(filepath.Join(fx.root, config.DefaultLeaderAnchorAuditRelPath)); err == nil {
				t.Errorf("%s: an audit row exists, so the value was recorded as an override", arm)
			}

			fx = newLeaderFixture(t)
			plain := driveLeaderLaunch(t, fx, arm, []string{"-f", "-cn", allowCardAnchorFlag, "hello", "-w", "plain"})
			if plain.err != nil || !plain.launched {
				t.Fatalf("%s: plain-tree launch: err=%v launched=%v", arm, plain.err, plain.launched)
			}
			want := []string{"-cn", allowCardAnchorFlag, "hello"}
			found := false
			for i := 0; i+len(want) <= len(plain.execArgs); i++ {
				found = found || reflect.DeepEqual(plain.execArgs[i:i+len(want)], want)
			}
			if !found {
				t.Errorf("%s: backend argv %q lacks the untouched run %q", arm, plain.execArgs, want)
			}
		})
		// Control: with the value attached there is no next-token value, so the
		// flag stands alone and the override is honored (and recorded).
		t.Run(arm+"/attached_cluster_value_leaves_the_flag_standing", func(t *testing.T) {
			fx := newLeaderFixture(t)
			obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "-cnNAME", allowCardAnchorFlag, "-w", "WT-card"})
			if obs.err != nil || !obs.launched {
				t.Fatalf("%s: override launch: err=%v launched=%v", arm, obs.err, obs.launched)
			}
			if _, err := os.Stat(filepath.Join(fx.root, config.DefaultLeaderAnchorAuditRelPath)); err != nil {
				t.Errorf("%s: no audit row for an honored override: %v", arm, err)
			}
		})
	}
}

// TestLauncherAnchorSeparatorAfterNonValueOption — SPEC-LAUNCHER-ARGV-001
// REQ-LA-005 / REQ-LA-003 (sync audit iteration 2, R1). A following `--` is the
// VALUE of `--model`, `-m` and `--permission-mode` only in their exact
// separate-token forms; after every other launcher-consumed form (`--model=v`,
// `--permission-mode=v`, `-b`, `--bypass`) it is the launcher's own separator, so
// a `--allow-card-anchor` after it belongs to the backend and is NOT an override.
func TestLauncherAnchorSeparatorAfterNonValueOption(t *testing.T) {
	pinSnapshotModel(t)

	nonValueForms := [][]string{
		{"--model=opus"},
		{"--permission-mode=plan"},
		{"-b"},
		{"--bypass"},
	}
	tail := []string{allowCardAnchorFlag, "-w", "WT-card"}

	// The consumption function and the view, in process.
	t.Run("flag_after_the_launcher_separator_is_the_backends", func(t *testing.T) {
		for _, form := range nonValueForms {
			args := append(append(append([]string{}, form...), argSeparator), tail...)
			rest, allowed := consumeAllowCardAnchor(args)
			if allowed || !reflect.DeepEqual(rest, args) {
				t.Errorf("consumeAllowCardAnchor(%q) = (%q, %v), want the argv untouched and no override", args, rest, allowed)
			}
			view := resolveLaunchView(rest)
			if !reflect.DeepEqual(view.FinalArgv, tail) {
				t.Errorf("%q: FinalArgv = %q, want the tail %q unconsumed", args, view.FinalArgv, tail)
			}
			if !view.Selection.Present || view.Selection.Value != "WT-card" {
				t.Errorf("%q: selection = %+v, want -w WT-card selected", args, view.Selection)
			}
		}
	})

	// Through the real leader entry on both arms: the guard refuses, nothing is
	// mutated, and no override audit row is written.
	for _, arm := range leaderArms {
		for _, form := range nonValueForms {
			name := arm + "/" + form[0]
			t.Run(name+"/refused", func(t *testing.T) {
				fx := newLeaderFixture(t)
				args := append(append(append([]string{"-f"}, form...), argSeparator), tail...)
				obs := driveLeaderLaunch(t, fx, arm, args)
				assertRefusedBeforeMutation(t, name, obs)
				refusalNames(t, name, obs, fx.card, "WT-card")
				if _, err := os.Stat(filepath.Join(fx.root, config.DefaultLeaderAnchorAuditRelPath)); err == nil {
					t.Errorf("%s: an audit row exists, so the backend-side flag was recorded as an override", name)
				}
			})
			// Control: the same flag standing in the LAUNCHER segment (before the
			// separator) is the override, and is honored and recorded.
			t.Run(name+"/control_flag_in_launcher_segment", func(t *testing.T) {
				fx := newLeaderFixture(t)
				args := append(append([]string{"-f"}, form...), allowCardAnchorFlag, "-w", "WT-card")
				obs := driveLeaderLaunch(t, fx, arm, args)
				if obs.err != nil || !obs.launched {
					t.Fatalf("%s: override in the launcher segment: err=%v launched=%v, want a launch", name, obs.err, obs.launched)
				}
				if _, err := os.Stat(filepath.Join(fx.root, config.DefaultLeaderAnchorAuditRelPath)); err != nil {
					t.Errorf("%s: no audit row for an honored override: %v", name, err)
				}
			})
		}

		// Control: the exact separate-token `--model` takes the `--` as its value, so
		// the flag after it stands in the launcher segment and IS the override.
		t.Run(arm+"/control_separator_is_the_value_of_model", func(t *testing.T) {
			fx := newLeaderFixture(t)
			obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "--model", argSeparator, allowCardAnchorFlag, "-w", "WT-card"})
			if obs.err != nil || !obs.launched {
				t.Fatalf("%s: err=%v launched=%v, want the override honored", arm, obs.err, obs.launched)
			}
			if _, err := os.Stat(filepath.Join(fx.root, config.DefaultLeaderAnchorAuditRelPath)); err != nil {
				t.Errorf("%s: no audit row for an honored override: %v", arm, err)
			}
		})
	}
}

// TestLauncherAnchorOptionValueSpanConsistency makes "one argv, one
// interpretation" a tested property: for every token form the launcher consumes
// (and a few it passes through), optionValueSpan's answer to "does the following
// `--` get taken as this option's value" equals what consumeLaunchArgs actually
// did with that `--`, and the guard's override placement follows the same fact.
func TestLauncherAnchorOptionValueSpanConsistency(t *testing.T) {
	pinSnapshotModel(t)

	forms := []string{
		// every case and default-branch form consumeLaunchArgs handles
		"-b", "--bypass", "--permission-mode", "--permission-mode=plan",
		"-c", "--continue", "--model", "--model=opus", "-m",
		// forms it passes through to the backend
		"--name", "-n", "-w", "--worktree", "-p",
	}
	for _, tok := range forms {
		t.Run(tok, func(t *testing.T) {
			c := consumeLaunchArgs([]string{tok, argSeparator, "TAIL"})
			tookSeparator := (c.ModelSet && c.Model == argSeparator) || (c.PermissionModeSet && c.PermissionMode == argSeparator)
			want := 0
			if tookSeparator {
				want = 1
			}
			if got := optionValueSpan([]string{tok, argSeparator, "TAIL"}, 0); got != want {
				t.Errorf("optionValueSpan(%q, --, TAIL) = %d, but consumeLaunchArgs took the separator as a value = %v (want span %d)", tok, got, tookSeparator, want)
			}
			// The launcher's own view of the separator: eaten as a value (the tail is
			// the only pass-through) or dropped as the separator (the tail follows
			// whatever the token itself passed through).
			if n := len(c.PassThrough); n == 0 || c.PassThrough[n-1] != "TAIL" {
				t.Errorf("consumeLaunchArgs(%q, --, TAIL).PassThrough = %q, want it to end in TAIL", tok, c.PassThrough)
			}
			for _, p := range c.PassThrough {
				if p == argSeparator {
					t.Errorf("consumeLaunchArgs(%q, --, TAIL).PassThrough = %q still carries the separator", tok, c.PassThrough)
				}
			}
			// The override placement agrees: a flag right after the `--` counts only
			// when the launcher took that `--` as a value (so the flag stands in the
			// launcher segment).
			if _, allowed := consumeAllowCardAnchor([]string{tok, argSeparator, allowCardAnchorFlag}); allowed != tookSeparator {
				t.Errorf("consumeAllowCardAnchor(%q, --, flag) allowed = %v, want %v (launcher took the separator as a value = %v)", tok, allowed, tookSeparator, tookSeparator)
			}
		})
	}
}
