package cli

// M4d contract tests for the help-group frequency reorder
// (SPEC-CLI-TUX-V3-004 REQ-TUX4-007, AC-TUX4-008). Test names match the AC
// run patterns 'HelpGroupOrder' / 'HelpGolden'.

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// groupNames returns the names of rootCmd's commands in Commands() order,
// filtered to one group ID.
func groupNames(gid string) []string {
	var names []string
	for _, c := range rootCmd.Commands() {
		if c.GroupID == gid {
			names = append(names, c.Name())
		}
	}
	return names
}

// TestHelpGroupOrder_FrequencyWithinGroups verifies the frequency-listed
// commands lead each group in declared order after the reorder.
func TestHelpGroupOrder_FrequencyWithinGroups(t *testing.T) {
	reorderRootHelpCommands(rootCmd)

	for gid, want := range helpGroupFrequency {
		got := groupNames(gid)
		// The listed commands that exist must appear as a prefix of the group,
		// in the declared order (unlisted members follow).
		gi := 0
		for _, name := range want {
			found := false
			for _, g := range got {
				if g == name {
					found = true
					break
				}
			}
			if !found {
				continue // frequency entry not registered in this build — skip
			}
			if gi >= len(got) || got[gi] != name {
				t.Errorf("group %q order mismatch: want %v as leading sequence, got %v", gid, want, got)
				break
			}
			gi++
		}
	}
}

// TestHelpGroupOrder_MembershipUnchanged verifies the reorder moves rows only:
// group IDs and command membership stay identical before/after (plan §D
// group-stability constraint).
func TestHelpGroupOrder_MembershipUnchanged(t *testing.T) {
	before := map[string]string{}
	for _, c := range rootCmd.Commands() {
		before[c.Name()] = c.GroupID
	}

	reorderRootHelpCommands(rootCmd)

	after := map[string]string{}
	for _, c := range rootCmd.Commands() {
		after[c.Name()] = c.GroupID
	}
	if len(before) != len(after) {
		t.Fatalf("command count changed: before %d, after %d", len(before), len(after))
	}
	for name, gid := range before {
		if after[name] != gid {
			t.Errorf("command %q group changed: %q -> %q", name, gid, after[name])
		}
	}
	// The three canonical groups still exist.
	gids := map[string]bool{}
	for _, g := range rootCmd.Groups() {
		gids[g.ID] = true
	}
	for _, want := range []string{"launch", "project", "tools"} {
		if !gids[want] {
			t.Errorf("group ID %q missing after reorder", want)
		}
	}
}

// TestHelpGroupOrder_Idempotent verifies repeated reorders are stable
// (Execute() may run multiple times in-process).
func TestHelpGroupOrder_Idempotent(t *testing.T) {
	reorderRootHelpCommands(rootCmd)
	first := append([]string{}, groupNames("project")...)
	reorderRootHelpCommands(rootCmd)
	second := groupNames("project")
	if strings.Join(first, ",") != strings.Join(second, ",") {
		t.Errorf("reorder not idempotent: %v vs %v", first, second)
	}
}

// TestHelpGolden_FangGroupHeaders verifies the adopted (keep-fang) help
// surface end-to-end: the three group-header literals render, and the
// frequency order survives into the rendered rows (AC-TUX4-008(2)(3)).
// Asserts headers + ordering rules rather than a byte snapshot — help goldens
// are kept assertion-based per acceptance.md §C (command additions must not
// break this test).
func TestHelpGolden_FangGroupHeaders(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	reorderRootHelpCommands(rootCmd)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	if err := runFang(context.Background(), rootCmd); err != nil {
		t.Fatalf("runFang --help: %v", err)
	}
	out := buf.String()

	// Adopted-surface header literals (keep-fang verdict, 2026-07-20 재실측).
	for _, header := range []string{"LAUNCH COMMANDS:", "PROJECT COMMANDS:", "TOOLS:"} {
		if !strings.Contains(out, header) {
			t.Errorf("fang help should carry group header %q, got:\n%s", header, out)
		}
	}
	// Zero ANSI under NO_COLOR (REQ-TUX4-008 matrix).
	if strings.Contains(out, "\x1b") {
		t.Error("NO_COLOR fang help must carry zero ANSI escape sequences")
	}
	// Frequency order renders: init precedes doctor in PROJECT, cc precedes cg
	// in LAUNCH, hook precedes mx in TOOLS (spot-check of the leading order).
	idx := func(s string) int { return strings.Index(out, s) }
	if a, b := idx("\n    init "), idx("\n    doctor "); a < 0 || b < 0 || a > b {
		t.Errorf("PROJECT order: init(%d) must precede doctor(%d)\n%s", a, b, out)
	}
	if a, b := idx("\n    glm "), idx("\n    cg "); a < 0 || b < 0 || a > b {
		t.Errorf("LAUNCH order: glm(%d) must precede cg(%d)\n%s", a, b, out)
	}
	if a, b := idx("\n    hook "), idx("\n    mx "); a < 0 || b < 0 || a > b {
		t.Errorf("TOOLS order: hook(%d) must precede mx(%d)\n%s", a, b, out)
	}
}
