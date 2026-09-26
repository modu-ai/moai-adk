package cli

// codex_kanban.go — `moai codex -k`: Kanban Mode entry for the Codex door, on
// the same terms as `moai cc -k`.
//
// The launcher uses DisableFlagParsing with a closed-set verb table, so the
// kanban tokens are intercepted before the verb lookup exactly as -f and -w
// are, and are never forwarded to the Codex child. The shape decisions are not
// re-implemented here: the tokens are handed to the cc surface's own parsers
// (parseKanbanFlag, parseCompanionLabel, parseLeadLabel), and the name claims
// go through the same registries (resolveCompanionName, appendLeadName), so
// the two doors cannot drift on what a shape means.
//
// Accepted shapes, lead and companion only:
//
//	-k [SPEC-ID]                        → the kanban lead
//	-k [SPEC-ID] --name <non-companion> → the lead, under the operator's name
//	-k --name <plan|run|sync>           → a kanban companion
//
// The factory shapes cc also accepts under -k (a count, a worker name) are
// refused: the Codex factory door is -f.

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// codexKanbanUsageDiag is the one-line diagnostic every unsupported -k shape
// receives.
const codexKanbanUsageDiag = "unsupported kanban shape - usage: moai codex -k [SPEC-ID] [--name <lead-name>] (lead) | moai codex -k --name <plan|run|sync> (companion); the factory entry is moai codex -f"

// codexKanbanEntry is the parsed -k request. nameArgs keeps the operator's
// --name tokens as typed, for the cc name parsers.
type codexKanbanEntry struct {
	enabled   bool
	spec      string
	companion string
	nameArgs  []string
}

// stripCodexKanbanFlag removes the kanban tokens (-k/--kanban with its value,
// and --name/-n with theirs) from the verb-position head. A routed verb
// following -k stays a verb: the verb slot and the SPEC slot are the same
// token position.
func stripCodexKanbanFlag(head []string) ([]string, codexKanbanEntry, error) {
	rest := make([]string, 0, len(head))
	var kanbanToks, nameToks []string
	for i := 0; i < len(head); i++ {
		token := head[i]
		switch {
		case token == kanbanFlagShort || token == kanbanFlagLong:
			kanbanToks = append(kanbanToks, token)
			if i+1 < len(head) && !strings.HasPrefix(head[i+1], "-") && !codexHeadTokenIsVerb(head[i+1]) {
				kanbanToks = append(kanbanToks, head[i+1])
				i++
			}
		case strings.HasPrefix(token, kanbanFlagShort+"="), strings.HasPrefix(token, kanbanFlagLong+"="):
			kanbanToks = append(kanbanToks, token)
		case token == nameFlagLong || token == nameFlagShort:
			if i+1 >= len(head) || strings.HasPrefix(head[i+1], "-") {
				return nil, codexKanbanEntry{}, errors.New(codexKanbanUsageDiag)
			}
			nameToks = append(nameToks, token, head[i+1])
			i++
		case strings.HasPrefix(token, nameFlagLong+"="), strings.HasPrefix(token, nameFlagShort+"="):
			nameToks = append(nameToks, token)
		default:
			rest = append(rest, token)
		}
	}
	if len(kanbanToks) == 0 {
		if len(nameToks) > 0 {
			// Codex has no session-name flag of its own; a name only means
			// something as a kanban role.
			return nil, codexKanbanEntry{}, errors.New(codexKanbanUsageDiag)
		}
		return rest, codexKanbanEntry{}, nil
	}
	p, err := parseKanbanFlag(append(append([]string{}, kanbanToks...), nameToks...))
	if err != nil {
		return nil, codexKanbanEntry{}, fmt.Errorf("%s (%v)", codexKanbanUsageDiag, err)
	}
	if p.FactoryEnabled {
		return nil, codexKanbanEntry{}, errors.New(codexKanbanUsageDiag)
	}
	entry := codexKanbanEntry{enabled: true, spec: p.Spec, nameArgs: nameToks}
	if label, ok := parseCompanionLabel(nameToks); ok {
		entry.companion = label
	}
	return rest, entry, nil
}

// applyCodexKanbanEntry publishes the kanban launch facts for e with backend
// codex, claiming the session name by the cc rules, and returns the restore
// func. The order mirrors the cc lead and companion branches.
func applyCodexKanbanEntry(cmd *cobra.Command, e codexKanbanEntry) func() {
	root := launchProjectRoot()
	if e.companion != "" {
		final := resolveCompanionName(root, e.companion, cmd.ErrOrStderr())
		restoreMode := enterKanbanCompanionMode(final)
		restoreFacts := exportKanbanLaunchFacts(e.spec, codexFactoryBackend)
		return func() {
			restoreFacts()
			restoreMode()
		}
	}
	leadLabel, _ := parseLeadLabel(e.nameArgs)
	restoreMode := enterKanbanMode(e.spec, leadLabel)
	restoreFacts := exportKanbanLaunchFacts(e.spec, codexFactoryBackend)
	// The name reaches the session through the environment: the Codex child
	// has no --name flag to carry it.
	_, leadName := appendLeadName(e.nameArgs, root, cmd.ErrOrStderr())
	restoreName := exportLeadSessionName(leadName)
	return func() {
		restoreName()
		restoreFacts()
		restoreMode()
	}
}
