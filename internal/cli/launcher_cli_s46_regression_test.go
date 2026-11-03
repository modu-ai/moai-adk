package cli

import (
	"reflect"
	"strings"
	"testing"
)

// Issue #1948 S46: old launch forms keep their meaning or fail with an explicit
// error; no form flips roles and none is silently absorbed. Pure argv parsing,
// no session is launched. Companion of launcher_legacy_forms_char_test.go.

// Role inversion is zero: the leader entries never select the lane branch and
// the lane entries never select the leader branch.
func TestS46_NoRoleInversion(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want factoryBranch
	}{
		{[]string{"-f"}, factoryBranchLeader},
		{[]string{"--factory"}, factoryBranchLeader},
		{[]string{"-a"}, factoryBranchLane},
		{[]string{"-a", "--leader", "leader-2"}, factoryBranchLane},
		{[]string{"-w", "n"}, factoryBranchNone},
		{[]string{"-w"}, factoryBranchNone},
	} {
		if _, got := legacyEntry(t, tc.args...); got != tc.want {
			t.Errorf("%v: branch = %v, want %v", tc.args, got, tc.want)
		}
	}
}

// -w composes with either entry and is never consumed by it.
func TestS46_WorktreeFlagKeptWithEntries(t *testing.T) {
	for _, entry := range []string{"-f", "-a"} {
		e, _ := legacyEntry(t, entry, "-w", "n")
		if !containsFlag(e.Rest, "-w") || !containsFlag(e.Rest, "n") {
			t.Errorf("%s -w n: Rest = %v lost the -w pair", entry, e.Rest)
		}
	}
}

// Forms that are refused today are refused with an explicit, named error.
func TestS46_RefusedFormsAreExplicit(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	for _, tc := range []struct {
		args []string
		want string // substring the error must carry
	}{
		{[]string{"-k"}, "retired"},
		{[]string{"--leader", "x"}, "-a"},
		{[]string{"--leader=x"}, "-a"},
		{[]string{"-f", "--leader", "x"}, "-a"},
		{[]string{"-f", "-a"}, ""},
		{[]string{"-l"}, "-a"},
		{[]string{"--lane"}, "-a"},
	} {
		_, err := parseLauncherEntry(tc.args)
		if err == nil {
			t.Errorf("%v: accepted, want an explicit error", tc.args)
			continue
		}
		if tc.want != "" && !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%v: error %q does not name %q", tc.args, err, tc.want)
		}
	}
}

// `--factory` is the long spelling of `-f`: same parse, no extra tokens.
func TestS46_FactoryLongFormEqualsShort(t *testing.T) {
	short, _ := legacyEntry(t, "-f")
	long, _ := legacyEntry(t, "--factory")
	if !reflect.DeepEqual(short, long) {
		t.Errorf("-f = %+v, --factory = %+v", short, long)
	}
}
