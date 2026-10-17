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
