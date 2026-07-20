package cli

// Help-group usage-frequency ordering (SPEC-CLI-TUX-V3-004 REQ-TUX4-007,
// M4d — keep-fang verdict). fang v2 renders help rows in cobra's Commands()
// order (fang help.go iterates c.Commands() verbatim), and cobra's default
// alphabetical sort is gated by cobra.EnableCommandSorting — so the reorder
// point is the cobra command registration order, with sorting disabled.
//
// Group IDs (launch/project/tools) and command membership are UNCHANGED
// (plan §D group-stability constraint): only the row order within each group
// moves. Ungrouped commands (the fang COMMANDS section) keep their prior
// alphabetical order.

import (
	"sort"

	"github.com/spf13/cobra"
)

// helpGroupFrequency is the usage-frequency leading order per help group.
// Commands not listed keep their prior (alphabetical) relative order AFTER
// the listed ones — the map does not need to enumerate every member.
//
// Frequency rationale (recorded in progress.md §E.2): launch — cc is the
// default launcher, glm the cost-optimized backend, cg the hybrid mode;
// project — init is the canonical entry, status/doctor are the daily
// health surfaces, update periodic, migrate/pr occasional; tools — hook is
// machine-invoked by every Claude Code session, spec/session/mx/loop are the
// core SPEC-lifecycle verbs, the remainder are occasional utilities.
var helpGroupFrequency = map[string][]string{
	"launch":  {"cc", "glm", "cg"},
	"project": {"init", "status", "doctor", "update", "migrate", "pr"},
	"tools":   {"hook", "spec", "session", "mx", "loop", "handoff", "model", "constitution", "state"},
}

// helpFrequencyRank returns the ordering rank of cmd within its group's
// frequency list; unlisted commands rank after all listed ones (stable sort
// preserves their prior relative order).
func helpFrequencyRank(list []string, cmd *cobra.Command) int {
	for i, name := range list {
		if cmd.Name() == name {
			return i
		}
	}
	return len(list)
}

// reorderRootHelpCommands re-registers root's subcommands so each help group
// lists its commands by usage frequency (REQ-TUX4-007). The reorder is
// position-preserving across groups: each group's members occupy the same
// slice positions as before, so group blocks and the ungrouped section are
// unaffected. Idempotent — safe to call on every Execute().
func reorderRootHelpCommands(root *cobra.Command) {
	cobra.EnableCommandSorting = false

	cmds := root.Commands()
	ordered := make([]*cobra.Command, len(cmds))
	copy(ordered, cmds)

	for gid, list := range helpGroupFrequency {
		var idxs []int
		var members []*cobra.Command
		for i, c := range ordered {
			if c.GroupID == gid {
				idxs = append(idxs, i)
				members = append(members, c)
			}
		}
		sort.SliceStable(members, func(a, b int) bool {
			return helpFrequencyRank(list, members[a]) < helpFrequencyRank(list, members[b])
		})
		for k, i := range idxs {
			ordered[i] = members[k]
		}
	}

	root.RemoveCommand(cmds...)
	root.AddCommand(ordered...)
}
