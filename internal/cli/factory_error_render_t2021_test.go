package cli

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

// Card t2021 repairs 2 and 3: the entry-path refusals must reach the operator
// with the exact flag spelling they typed. fang's ErrorText style title-cases the
// first word of every error, so a message that STARTS with a flag prints "-L",
// "--Lane", "-F/--Factory", "-A", "--Leader", "--Factory-Run", "-K/--Kanban" or
// "--Key" — uppercase -F/-L/-A/-K are separate flags, so the text would name a
// flag the operator never typed. These tests go through the real output path
// (runFang -> fang.Execute -> moaiErrorHandler), not err.Error().
//
// Every row is an input the parser refuses before anything launches; a row that
// stopped being refused would try to start a session, so keep rows refusal-only.

// renderEntryError runs the real root command through runFang and returns what
// the error handler printed on stderr.
func renderEntryError(t *testing.T, args ...string) string {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	var errOut bytes.Buffer
	rootCmd.SetErr(&errOut)
	rootCmd.SetOut(io.Discard)
	rootCmd.SetArgs(args)
	t.Cleanup(func() {
		rootCmd.SetErr(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)
	})
	if err := runFang(context.Background(), rootCmd); err == nil {
		t.Fatalf("%v: runFang returned nil, want the entry refusal", args)
	}
	return errOut.String()
}

func TestT2021_EntryRefusalsKeepTheTypedFlagCase(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		want    string // exact text that must appear
		notWant string // the re-cased corruption that must not appear
	}{
		// Repair 2 / round 2: the removed agent flags and the entry-token conflict.
		{"cc -a", []string{"cc", "-a"}, "-a flag was removed; use -l", "-A flag"},
		{"cc -a=x", []string{"cc", "-a=x"}, "-a flag was removed; use -l", "-A flag"},
		{"cc --agent", []string{"cc", "--agent"}, "--agent flag was removed; use -l", "--Agent"},
		{"cc --agent=x", []string{"cc", "--agent=x"}, "--agent flag was removed; use -l", "--Agent"},
		{"glm -a", []string{"glm", "-a"}, "-a flag was removed; use -l", "-A flag"},
		{"glm --agent", []string{"glm", "--agent"}, "--agent flag was removed; use -l", "--Agent"},
		{"cc -- --agent x", []string{"cc", "--", "--agent", "x"}, "--agent flag was removed; use --lane", "--Agent"},
		{"cc -- --agent=x", []string{"cc", "--", "--agent=x"}, "--agent flag was removed; use --lane", "--Agent"},
		{"cc -f -l", []string{"cc", "-f", "-l"}, "-f/--factory and -l/--lane are two entry tokens", "-F/--Factory"},
		{"cc -l -f", []string{"cc", "-l", "-f"}, "-f/--factory and -l/--lane are two entry tokens", "-F/--Factory"},

		// Repair 3: -f / --factory carrying a value.
		{"cc -f=x", []string{"cc", "-f=x"}, "-f/--factory flag takes no argument", "-F/--Factory"},
		{"cc --factory=x", []string{"cc", "--factory=x"}, "-f/--factory flag takes no argument", "-F/--Factory"},
		{"cc -f 4", []string{"cc", "-f", "4"}, "-f/--factory flag takes no argument", "-F/--Factory"},
		{"glm -f lane-2", []string{"glm", "-f", "lane-2"}, "-f/--factory flag takes no argument", "-F/--Factory"},

		// Repair 3: -l / --lane carrying a value, or a name.
		{"cc -l=x", []string{"cc", "-l=x"}, "-l/--lane flag takes no argument", "-L/--Lane"},
		{"cc --lane=x", []string{"cc", "--lane=x"}, "-l/--lane flag takes no argument", "-L/--Lane"},
		{"cc -l x", []string{"cc", "-l", "x"}, "-l/--lane flag takes no argument", "-L/--Lane"},
		{"cc -l --name foo", []string{"cc", "-l", "--name", "foo"}, "-l/--lane flag already names the role", "-L/--Lane"},
		{"cc -f --name lane-2", []string{"cc", "-f", "--name", "lane-2"}, "--name lane-<n> value does not select a lane", "--Name"},

		// Repair 3: --leader.
		{"cc --leader", []string{"cc", "--leader"}, "--leader flag requires a leader label", "--Leader"},
		{"cc -l --leader lead-1", []string{"cc", "-l", "--leader", "lead-1"}, "--leader value \"lead-1\" is the legacy leader spelling", "--Leader"},
		{"cc -f --leader leader-2", []string{"cc", "-f", "--leader", "leader-2"}, "--leader flag names the leader session", "--Leader"},
		{"cc --leader x --factory-run y", []string{"cc", "--leader", "x", "--factory-run", "y"}, "--leader and --factory-run flags name two different selectors", "--Leader"},

		// Repair 3: --factory-run.
		{"cc --factory-run", []string{"cc", "--factory-run"}, "--factory-run flag requires a run id", "--Factory-Run"},
		{"cc --factory-run=", []string{"cc", "--factory-run="}, "--factory-run flag requires a run id", "--Factory-Run"},

		// Repair 3: --clear-policy and --no-auto-dispatch.
		{"cc -l --clear-policy", []string{"cc", "-l", "--clear-policy"}, "--clear-policy flag requires a clear policy value", "--Clear-Policy"},
		{"cc -l --clear-policy bogus", []string{"cc", "-l", "--clear-policy", "bogus"}, "--clear-policy flag takes clear-each", "--Clear-Policy"},
		{"cc --clear-policy clear-each", []string{"cc", "--clear-policy", "clear-each"}, "--clear-policy flag selects a factory lane's clear policy", "--Clear-Policy"},
		{"cc -f --clear-policy clear-each", []string{"cc", "-f", "--clear-policy", "clear-each"}, "--clear-policy flag selects a lane's clear policy", "--Clear-Policy"},
		{"cc -l --no-auto-dispatch=x", []string{"cc", "-l", "--no-auto-dispatch=x"}, "--no-auto-dispatch flag takes no value", "--No-Auto-Dispatch"},
		{"cc --no-auto-dispatch", []string{"cc", "--no-auto-dispatch"}, "--no-auto-dispatch flag selects a factory lane's dispatch mode", "--No-Auto-Dispatch"},
		{"cc -f --no-auto-dispatch", []string{"cc", "-f", "--no-auto-dispatch"}, "--no-auto-dispatch flag selects a lane's dispatch mode", "--No-Auto-Dispatch"},

		// Repair 3: the legacy --name spellings.
		{"cc --name lead", []string{"cc", "--name", "lead"}, "--name value \"lead\" is the legacy leader spelling", "--Name"},
		{"cc --name worker-2", []string{"cc", "--name", "worker-2"}, "--name value \"worker-2\" is the legacy lane label", "--Name"},
		{"cc --name worker", []string{"cc", "--name", "worker"}, "--name value \"worker\" is a legacy factory spelling", "--Name"},

		// Repair 3: the retired -k / --kanban entry.
		{"cc -k", []string{"cc", "-k"}, "-k/--kanban flag is retired", "-K/--Kanban"},
		{"cc --kanban", []string{"cc", "--kanban"}, "-k/--kanban flag is retired", "-K/--Kanban"},
		{"glm -k", []string{"glm", "-k"}, "-k/--kanban flag is retired", "-K/--Kanban"},

		// Repair 3: the glm --key save form's refusals (nothing is stored).
		{"glm --key", []string{"glm", "--key"}, "--key flag requires a value", "--Key"},
		{"glm --key abc extra", []string{"glm", "--key", "abc", "extra"}, "--key flag cannot be combined", "--Key"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := renderEntryError(t, tc.args...)
			if !strings.Contains(got, tc.want) {
				t.Errorf("printed error lacks %q:\n%s", tc.want, got)
			}
			if strings.Contains(got, tc.notWant) {
				t.Errorf("printed error carries the re-cased flag %q:\n%s", tc.notWant, got)
			}
		})
	}
}
