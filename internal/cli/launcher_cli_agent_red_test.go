//go:build afred

package cli

import "testing"

// RED on purpose (Issue #1948 S46): `-a` / `--agent` is not implemented, and
// today it is silently passed through (observed: err=nil, no factory entry,
// the token stays in Rest; beside -f or -l it stays in Rest too). The expected
// behavior is "consumed as the agent entry, or an explicit error", never a
// silent pass-through. Run with: go test -tags afred ./internal/cli/ -run RED_
// These stay out of the default build so they cannot break CI before the
// implementation lands.

func consumedOrRefused(t *testing.T, args []string, flag string) bool {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	e, err := parseLauncherEntry(args)
	if err != nil {
		return true
	}
	return !containsFlag(e.Rest, flag)
}

func TestRED_AgentFlag_NotSilentlyAbsorbed(t *testing.T) {
	for _, tc := range []struct {
		args []string
		flag string
	}{
		{[]string{"-a"}, "-a"},
		{[]string{"-a", "x"}, "-a"},
		{[]string{"--agent"}, "--agent"},
		{[]string{"-f", "-a"}, "-a"},
		{[]string{"-l", "-a"}, "-a"},
	} {
		if !consumedOrRefused(t, tc.args, tc.flag) {
			t.Errorf("%v: %s silently passed through to the backend", tc.args, tc.flag)
		}
	}
}
