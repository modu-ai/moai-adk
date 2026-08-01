package cli

import "testing"

// TestPRWatchCmd_FlagSetUnchanged pins the flag surface of `moai pr watch`.
//
// The CI-loop distribution-isolation work edits only user-visible strings in
// pr_watch_cmd.go (the Long help text and two stderr notices). Two of those
// three strings live inside RunE, so they sit within the blast radius of an
// editing accident that could drop or rename a flag. The CLI-level invariance
// check for that runs by hand at judgment time; this test is the durable
// in-process counterpart, asserting the flag set and its defaults survive.
//
// Deliberately NOT asserted here: exit codes and stderr wording. Exit codes
// depend on filesystem state (--abort's behavior turns on whether the CI-watch
// state file exists), which belongs to the CLI-level check rather than a unit
// test; stderr wording is what the edit intentionally changes.
func TestPRWatchCmd_FlagSetUnchanged(t *testing.T) {
	cmd := newPRWatchCmd()

	boolFlags := map[string]bool{
		"abort":  false,
		"report": false,
	}
	for name, wantDefault := range boolFlags {
		f := cmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("flag --%s missing from `moai pr watch`", name)
			continue
		}
		got, err := cmd.Flags().GetBool(name)
		if err != nil {
			t.Errorf("flag --%s is not a bool flag: %v", name, err)
			continue
		}
		if got != wantDefault {
			t.Errorf("flag --%s default = %v, want %v", name, got, wantDefault)
		}
	}

	branch, err := cmd.Flags().GetString("branch")
	if err != nil {
		t.Fatalf("flag --branch missing or not a string flag: %v", err)
	}
	if branch != "main" {
		t.Errorf("flag --branch default = %q, want %q", branch, "main")
	}

	// Args contract: --abort needs no positional PR number; without it, one is
	// required. Both branches are user-visible behavior the string edit must
	// not disturb.
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("Args accepted zero positional args without --abort, want error")
	}
	if err := cmd.Flags().Set("abort", "true"); err != nil {
		t.Fatalf("set --abort: %v", err)
	}
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("Args rejected zero positional args with --abort set: %v", err)
	}
}
